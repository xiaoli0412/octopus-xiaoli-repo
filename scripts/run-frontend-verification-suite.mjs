#!/usr/bin/env node

import { spawn } from 'node:child_process';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');
const webDir = path.join(repoRoot, 'web');
const nodeExe = process.execPath;

const suites = {
  settings: [
    { label: 'backup-logic', args: ['--experimental-strip-types', path.join(repoRoot, 'scripts', 'verify-backup-logic.mjs')] },
    { label: 'backup-component', args: [path.join(repoRoot, 'scripts', 'verify-backup-component.cjs')] },
    { label: 'help-hint-accessible', args: [path.join(repoRoot, 'scripts', 'verify-help-hint-accessible.mjs')] },
    { label: 'dynamic-routing-help', args: [path.join(repoRoot, 'scripts', 'verify-dynamic-routing-help.mjs')] },
    { label: 'circuit-breaker-help', args: [path.join(repoRoot, 'scripts', 'verify-circuit-breaker-help.mjs')] },
    { label: 'model-probe-help', args: [path.join(repoRoot, 'scripts', 'verify-model-probe-help.mjs')] },
    { label: 'setting-info-logic', args: [path.join(repoRoot, 'scripts', 'verify-setting-info-logic.mjs')] },
    { label: 'ai-config-profile-summary', args: [path.join(repoRoot, 'scripts', 'verify-ai-config-profile-summary.mjs')] },
    { label: 'ai-learning-focus', args: [path.join(repoRoot, 'scripts', 'verify-ai-automation-learning-focus.mjs')] },
  ],
  screenshot: [
    { label: 'locale-consistency', args: [path.join(repoRoot, 'scripts', 'verify-locale-consistency.mjs')] },
    { label: 'home-layout', args: [path.join(repoRoot, 'scripts', 'verify-home-layout.mjs')] },
    { label: 'channel-create-flow', args: [path.join(repoRoot, 'scripts', 'verify-channel-create-flow.mjs')] },
    { label: 'channel-presentation', args: [path.join(repoRoot, 'scripts', 'verify-channel-presentation.mjs')] },
    { label: 'group-create-flow', args: [path.join(repoRoot, 'scripts', 'verify-group-create-flow.mjs')] },
    { label: 'llm-price-boundary', args: [path.join(repoRoot, 'scripts', 'verify-llm-price-boundary.mjs')] },
    { label: 'browser-smoke-wrapper-alignment', args: [path.join(repoRoot, 'scripts', 'verify-browser-smoke-wrapper-alignment.mjs')] },
    { label: 'settings-no-browser', suite: 'settings' },
    { label: 'ccswitch-flow', args: [path.join(repoRoot, 'scripts', 'verify-ccswitch-flow.mjs')] },
    { label: 'route-target-copy', args: [path.join(repoRoot, 'scripts', 'verify-route-target-copy.mjs')] },
  ],
};

function parseSuiteName(argv) {
  const suiteName = argv[2] ?? 'settings';
  if (!(suiteName in suites)) {
    const validSuites = Object.keys(suites).join(', ');
    throw new Error(`Unknown verification suite '${suiteName}'. Expected one of: ${validSuites}`);
  }
  return suiteName;
}

function runStep(label, args) {
  return new Promise((resolve, reject) => {
    const child = spawn(nodeExe, args, {
      cwd: repoRoot,
      env: { ...process.env },
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

      const detail = signal ? `signal ${signal}` : `exit code ${code ?? 1}`;
      reject(new Error(`${label} failed with ${detail}`));
    });
  });
}

async function runSuite(name, seen = new Set()) {
  if (seen.has(name)) {
    throw new Error(`Verification suite cycle detected at '${name}'`);
  }

  seen.add(name);
  for (const step of suites[name]) {
    if (step.suite) {
      await runSuite(step.suite, seen);
      continue;
    }
    await runStep(step.label, step.args);
  }
  seen.delete(name);
}

async function main() {
  process.chdir(webDir);
  const suiteName = parseSuiteName(process.argv);
  await runSuite(suiteName);
  console.log(`${suiteName} verification suite passed`);
}

main().catch((error) => {
  console.error(error.stack || String(error));
  process.exitCode = 1;
});
