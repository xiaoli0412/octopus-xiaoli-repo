#!/usr/bin/env node

import { spawn } from 'node:child_process';
import { createRequire } from 'node:module';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');
const webDir = path.join(repoRoot, 'web');
const nodeExe = process.execPath;
const webRequire = createRequire(path.join(webDir, 'package.json'));
const clonePatchMarker = Symbol('octopus.clone.patch');

function sanitizeCloneable(value, seen = new WeakMap()) {
  if (typeof value === 'function') {
    return undefined;
  }

  if (value === null || value === undefined) {
    return value;
  }

  if (typeof value !== 'object') {
    return value;
  }

  if (
    value instanceof Date ||
    value instanceof RegExp ||
    value instanceof ArrayBuffer ||
    ArrayBuffer.isView(value) ||
    value instanceof URL
  ) {
    return value;
  }

  if (seen.has(value)) {
    return seen.get(value);
  }

  if (Array.isArray(value)) {
    const clone = [];
    seen.set(value, clone);
    for (const item of value) {
      clone.push(sanitizeCloneable(item, seen));
    }
    return clone;
  }

  if (value instanceof Map) {
    const clone = new Map();
    seen.set(value, clone);
    for (const [key, entry] of value.entries()) {
      clone.set(sanitizeCloneable(key, seen), sanitizeCloneable(entry, seen));
    }
    return clone;
  }

  if (value instanceof Set) {
    const clone = new Set();
    seen.set(value, clone);
    for (const entry of value.values()) {
      clone.add(sanitizeCloneable(entry, seen));
    }
    return clone;
  }

  const clone = Object.create(Object.getPrototypeOf(value) === null ? null : Object.prototype);
  seen.set(value, clone);
  for (const [key, entry] of Object.entries(value)) {
    clone[key] = sanitizeCloneable(entry, seen);
  }
  return clone;
}

function patchThreadWorkerMessages(nextWorkerInstance) {
  const workerPool = nextWorkerInstance?._worker?._workerPool?._workers ?? [];

  for (const worker of workerPool) {
    if (!worker || typeof worker.send !== 'function' || worker[clonePatchMarker]) {
      continue;
    }

    const originalSend = worker.send.bind(worker);
    worker.send = (request, onStart, onEnd, onCustomMessage) => {
      const sanitizedRequest = sanitizeCloneable(request);
      return originalSend(sanitizedRequest, onStart, onEnd, onCustomMessage);
    };
    worker[clonePatchMarker] = true;
  }
}

function runStep(label, file, args, { cwd, env } = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(file, args, {
      cwd,
      env: env ? { ...process.env, ...env } : process.env,
      stdio: 'inherit',
      windowsHide: true,
    });

    child.on('error', (error) => {
      reject(new Error(`${label} failed to start: ${error.message}`));
    });

    child.on('exit', (code, signal) => {
      if (code === 0) {
        resolve();
        return;
      }

      const suffix = signal ? `signal ${signal}` : `exit code ${code ?? 1}`;
      reject(new Error(`${label} failed with ${suffix}`));
    });
  });
}

async function runPatchedNextBuild() {
  process.env.OCTOPUS_NEXT_SKIP_INTERNAL_TS = '1';
  process.env.OCTOPUS_NEXT_ENABLE_WORKER_THREADS = '1';

  const workerModuleId = webRequire.resolve('next/dist/lib/worker');
  const workerModule = webRequire(workerModuleId);
  const OriginalWorker = workerModule.Worker;

  class PatchedWorker extends OriginalWorker {
    constructor(workerPath, options = {}) {
      if (options?.enableWorkerThreads) {
        const sanitizedOptions = { ...options };

        // `jest-worker` thread mode uses structured clone for worker options.
        // Next passes progress callbacks that are fine for child processes but
        // not cloneable for worker threads on Windows.
        delete sanitizedOptions.onActivity;
        delete sanitizedOptions.onActivityAbort;

        super(workerPath, sanitizedOptions);
        patchThreadWorkerMessages(this);
        return;
      }

      super(workerPath, options);
    }
  }

  webRequire.cache[workerModuleId].exports = {
    __esModule: true,
    Worker: PatchedWorker,
    getNextBuildDebuggerPortOffset: workerModule.getNextBuildDebuggerPortOffset,
  };

  const nextBuild = webRequire('next/dist/build').default;
  await nextBuild(webDir);
}

async function main() {
  await runStep(
    'TypeScript preflight',
    nodeExe,
    [path.join(webDir, 'node_modules', 'typescript', 'bin', 'tsc'), '--noEmit', '-p', path.join(webDir, 'tsconfig.json')],
    { cwd: webDir },
  );

  await runPatchedNextBuild();

  await runStep(
    'Static sync',
    nodeExe,
    [path.join(repoRoot, 'scripts', 'sync-web-static.mjs')],
    { cwd: repoRoot },
  );
}

main().catch((error) => {
  console.error(error.stack || String(error));
  process.exitCode = 1;
});
