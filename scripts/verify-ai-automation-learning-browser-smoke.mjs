import assert from 'node:assert/strict';
import { existsSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { spawn } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, '..');
const argv = new Set(process.argv.slice(2));

function hasArg(flag) {
  return argv.has(flag);
}

function normalizeBaseUrl(value) {
  return value.replace(/\/$/, '');
}

function resolveNodeExe() {
  const explicit = process.env.OCTOPUS_UI_SMOKE_NODE?.trim();
  if (explicit) {
    return explicit;
  }
  return process.execPath;
}

function resolveNpxCliScript(nodePath) {
  const explicit = process.env.OCTOPUS_UI_SMOKE_NPX_SCRIPT?.trim();
  if (explicit) {
    return explicit;
  }

  return path.join(path.dirname(nodePath), 'node_modules', 'npm', 'bin', 'npx-cli.js');
}

function resolveMode() {
  const explicitMode = process.env.OCTOPUS_UI_SMOKE_MODE?.trim();
  if (explicitMode) {
    return explicitMode;
  }
  if (hasArg('--check-only')) {
    return 'check-only';
  }
  if (hasArg('--external')) {
    return 'external';
  }
  if (hasArg('--self-start')) {
    return 'self-start';
  }
  return 'self-start';
}

const mode = resolveMode();
if (!['check-only', 'external', 'self-start'].includes(mode)) {
  throw new Error(`Unsupported OCTOPUS_UI_SMOKE_MODE: ${mode}`);
}

const frontPort = Number(process.env.OCTOPUS_UI_SMOKE_FRONTEND_PORT || 3101);
const backPort = Number(process.env.OCTOPUS_UI_SMOKE_BACKEND_PORT || 18081);
const frontendBaseUrl = normalizeBaseUrl(process.env.OCTOPUS_UI_SMOKE_FRONTEND_URL || `http://127.0.0.1:${frontPort}`);
const backendBaseUrl = normalizeBaseUrl(process.env.OCTOPUS_UI_SMOKE_BACKEND_URL || `http://127.0.0.1:${backPort}`);
const npmCache = process.env.npm_config_cache || path.join(repoRoot, '.codex-tmp', 'npm-cache');
const npmRegistry = process.env.npm_config_registry || 'https://registry.npmmirror.com';
const nodeExe = resolveNodeExe();
const npxCliScript = resolveNpxCliScript(nodeExe);
const goExe = process.env.OCTOPUS_UI_SMOKE_GO || 'go';
const backendBin = process.env.OCTOPUS_UI_SMOKE_BACKEND_BIN?.trim() || '';
const browserName = process.env.OCTOPUS_UI_SMOKE_BROWSER || 'msedge';
const adminUsername = process.env.OCTOPUS_UI_SMOKE_ADMIN_USERNAME || 'admin';
const adminPassword = process.env.OCTOPUS_UI_SMOKE_ADMIN_PASSWORD || 'admin';
const adminToken = process.env.OCTOPUS_UI_SMOKE_ADMIN_TOKEN?.trim() || '';
const sessionName = process.env.OCTOPUS_UI_SMOKE_SESSION || `octopus-ai-learning-${process.pid}`;
const playwrightPackage = process.env.OCTOPUS_UI_SMOKE_PLAYWRIGHT_PACKAGE || '@playwright/cli';
const playwrightCliTimeoutMs = Number(process.env.OCTOPUS_UI_SMOKE_CLI_TIMEOUT_MS || 30000);

function smokeLog(message, details = {}) {
  const suffix = Object.keys(details).length > 0 ? ` ${JSON.stringify(details)}` : '';
  console.log(`[ai-learning-smoke] ${message}${suffix}`);
}

function describeCliArgs(args) {
  return args.map((item) => {
    const text = String(item);
    return text.length > 120 ? `${text.slice(0, 117)}...` : text;
  });
}

function printUsage() {
  console.log([
    'Usage:',
    '  node scripts/verify-ai-automation-learning-browser-smoke.mjs --check-only',
    '  node scripts/verify-ai-automation-learning-browser-smoke.mjs --external',
    '  node scripts/verify-ai-automation-learning-browser-smoke.mjs --self-start',
    '',
    'Environment overrides:',
    '  OCTOPUS_UI_SMOKE_MODE=check-only|external|self-start',
    '  OCTOPUS_UI_SMOKE_FRONTEND_URL=http://127.0.0.1:3101',
    '  OCTOPUS_UI_SMOKE_BACKEND_URL=http://127.0.0.1:18081',
    '  OCTOPUS_UI_SMOKE_BACKEND_BIN=build/octopus-smoke.exe',
    '  OCTOPUS_UI_SMOKE_NPX_SCRIPT=D:/gol1/node_modules/npm/bin/npx-cli.js',
    '  OCTOPUS_UI_SMOKE_ADMIN_TOKEN=<jwt>',
  ].join('\n'));
}

function spawnProcess(command, args, options = {}) {
  const child = spawn(command, args, {
    cwd: repoRoot,
    stdio: ['ignore', 'pipe', 'pipe'],
    windowsHide: true,
    ...options,
  });
  let stdout = '';
  let stderr = '';
  child.stdout?.on('data', (chunk) => {
    stdout += chunk.toString();
  });
  child.stderr?.on('data', (chunk) => {
    stderr += chunk.toString();
  });
  child.collected = {
    get stdout() {
      return stdout;
    },
    get stderr() {
      return stderr;
    },
  };
  return child;
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function waitForHttp(url, timeoutMs = 30000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    try {
      const response = await fetch(url);
      if (response.ok) {
        return;
      }
    } catch {
    }
    await sleep(500);
  }
  throw new Error(`Timed out waiting for ${url}`);
}

async function resolveAdminToken() {
  if (adminToken) {
    return adminToken;
  }
  const response = await fetch(`${backendBaseUrl}/api/v1/user/login`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ username: adminUsername, password: adminPassword, expire: 1440 }),
  });
  const payload = await response.json();
  if (!response.ok) {
    throw new Error(`Login failed: ${JSON.stringify(payload)}`);
  }
  return payload.data.token;
}

async function runPlaywrightCli(args, { raw = false } = {}) {
  if (!existsSync(npxCliScript)) {
    throw new Error(`npx-cli.js not found: ${npxCliScript}`);
  }

  const env = { ...process.env, npm_config_cache: npmCache, npm_config_registry: npmRegistry };
  const finalArgs = [];
  if (sessionName) {
    finalArgs.push(`-s=${sessionName}`);
  }
  if (raw) {
    finalArgs.push('--raw');
  }
  finalArgs.push(...args);
  smokeLog('playwright-cli start', { raw, args: describeCliArgs(args) });

  const invokeCli = () => new Promise((resolve, reject) => {
    const startedAt = Date.now();
    const child = spawn(nodeExe, [npxCliScript, '-y', playwrightPackage, ...finalArgs], {
      cwd: repoRoot,
      env,
      windowsHide: true,
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    let stdout = '';
    let stderr = '';
    let settled = false;
    const timeout = setTimeout(() => {
      if (settled) {
        return;
      }
      settled = true;
      try {
        child.kill('SIGTERM');
      } catch {
      }
      smokeLog('playwright-cli timeout', { raw, args: describeCliArgs(args), timeoutMs: playwrightCliTimeoutMs, stdoutBytes: stdout.length, stderrBytes: stderr.length });
      reject(new Error(`playwright-cli ${finalArgs.join(' ')} timed out after ${playwrightCliTimeoutMs}ms\nSTDOUT:\n${stdout}\nSTDERR:\n${stderr}`));
    }, playwrightCliTimeoutMs);
    child.stdout.on('data', (chunk) => {
      stdout += chunk.toString();
    });
    child.stderr.on('data', (chunk) => {
      stderr += chunk.toString();
    });
    child.on('error', (error) => {
      if (settled) {
        return;
      }
      settled = true;
      clearTimeout(timeout);
      reject(error);
    });
    child.on('close', (code) => {
      if (settled) {
        return;
      }
      settled = true;
      clearTimeout(timeout);
      if (code === 0) {
        smokeLog('playwright-cli done', { raw, args: describeCliArgs(args), elapsedMs: Date.now() - startedAt, stdoutBytes: stdout.length, stderrBytes: stderr.length });
        resolve({ stdout, stderr });
        return;
      }
      smokeLog('playwright-cli failed', { raw, args: describeCliArgs(args), elapsedMs: Date.now() - startedAt, code, stdoutBytes: stdout.length, stderrBytes: stderr.length });
      reject(new Error(`playwright-cli ${finalArgs.join(' ')} failed with code ${code}\nSTDOUT:\n${stdout}\nSTDERR:\n${stderr}`));
    });
  });

  for (let attempt = 0; attempt < 6; attempt += 1) {
    try {
      return await invokeCli();
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      if (/is not open, please run open first/i.test(message)) {
        await sleep(1000);
        continue;
      }
      throw error;
    }
  }

  throw new Error(`playwright-cli ${finalArgs.join(' ')} failed after retry`);
}

function parsePlaywrightJson(stdout) {
  try {
    const parsed = JSON.parse(stdout.trim());
    return typeof parsed === 'string' ? JSON.parse(parsed) : parsed;
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`Unable to parse playwright raw JSON output: ${message}\n${stdout}`);
  }
}

function writeBackendConfig(tempDir) {
  const configPath = path.join(tempDir, 'config.json');
  const dbPath = path.join(tempDir, 'octopus.db');
  const config = {
    server: {
      host: '127.0.0.1',
      port: backPort,
      static_dir: path.join(repoRoot, 'static', 'out').replace(/\\/g, '/'),
    },
    database: {
      type: 'sqlite',
      path: dbPath.replace(/\\/g, '/'),
    },
    log: { level: 'info' },
  };
  writeFileSync(configPath, JSON.stringify(config, null, 2));
  return configPath;
}

function startBackend(configPath) {
  const env = {
    ...process.env,
    OCTOPUS_ADMIN_USERNAME: adminUsername,
    OCTOPUS_ADMIN_PASSWORD: adminPassword,
  };
  if (backendBin) {
    const backendCommand = path.isAbsolute(backendBin) ? backendBin : path.join(repoRoot, backendBin);
    return spawnProcess(backendCommand, ['start', '--config', configPath], { env });
  }
  return spawnProcess(goExe, ['run', 'main.go', 'start', '--config', configPath], { env });
}

function startFrontend() {
  return spawnProcess(nodeExe, ['node_modules/next/dist/bin/next', 'dev', '--port', String(frontPort)], {
    cwd: path.join(repoRoot, 'web'),
    env: {
      ...process.env,
      NEXT_PUBLIC_API_BASE_URL: backendBaseUrl,
    },
  });
}

async function waitForLearningSection(timeoutMs = 45000) {
  const startedAt = Date.now();
  let lastBodyText = '';
  let lastErrorMessage = '';
  let historyTabClicked = false;
  const snapshotEval = `() => ({
    page: !!document.querySelector('[data-testid="ai-automation-page"]'),
    historyTab: !!Array.from(document.querySelectorAll('button')).find((button) => button.textContent?.includes('历史与学习') || button.textContent?.includes('History and learning') || button.textContent?.includes('歷史與學習') || button.textContent?.includes('履歴と学習')),
    section: !!document.querySelector('[data-testid="ai-automation-learning-section"]'),
    stage: !!document.querySelector('[data-testid="ai-automation-learning-stage-card"]'),
    controls: !!document.querySelector('[data-testid="ai-automation-learning-controls"]'),
    presetCard: !!document.querySelector('[data-testid="ai-automation-learning-preset-card"]'),
    switchCard: !!document.querySelector('[data-testid="ai-automation-learning-switch-card"]'),
    summary: !!document.querySelector('[data-testid="ai-automation-learning-state-summary"]'),
    secondarySummary: !!document.querySelector('[data-testid="ai-automation-learning-secondary-summary"]'),
    states: !!document.querySelector('[data-testid="ai-automation-learning-states"]'),
    bodyText: document.body.innerText,
  })`;

  while (Date.now() - startedAt < timeoutMs) {
    try {
      const result = await runPlaywrightCli(['eval', snapshotEval], { raw: true });
      const snapshot = parsePlaywrightJson(result.stdout);
      lastBodyText = snapshot.bodyText || '';
      lastErrorMessage = '';
      if (!historyTabClicked && snapshot.page && snapshot.historyTab && !snapshot.section) {
        await runPlaywrightCli(['eval', `() => {
          const button = Array.from(document.querySelectorAll('button')).find((node) => node.textContent?.includes('历史与学习') || node.textContent?.includes('History and learning') || node.textContent?.includes('歷史與學習') || node.textContent?.includes('履歴と学習'));
          if (button instanceof HTMLButtonElement) button.click();
        }`], { raw: true });
        historyTabClicked = true;
        await sleep(300);
        continue;
      }
      if (snapshot.page && snapshot.section && snapshot.stage && snapshot.controls && snapshot.presetCard && snapshot.switchCard && snapshot.summary && snapshot.secondarySummary && snapshot.states) {
        return snapshot;
      }
    } catch (error) {
      lastErrorMessage = error instanceof Error ? error.message : String(error);
    }
    await sleep(1000);
  }

  const errorSuffix = lastErrorMessage ? ` Last eval error: ${lastErrorMessage}` : '';
  throw new Error(`Timed out waiting for AI automation learning section on ${frontendBaseUrl}. Last body text snippet: ${lastBodyText.slice(0, 200)}${errorSuffix}`);
}

async function main() {
  if (hasArg('--help')) {
    printUsage();
    return;
  }

  if (mode === 'check-only') {
    console.log(JSON.stringify({
      mode,
      frontendBaseUrl,
      backendBaseUrl,
      browserName,
      sessionName,
      backendLaunch: backendBin ? { kind: 'binary', command: backendBin } : { kind: 'go-run', command: goExe },
      frontendLaunch: { command: nodeExe, cwd: 'web', nextApiBaseUrl: backendBaseUrl },
      nodeExe,
      npxCliScript,
      note: 'check-only does not spawn backend, frontend, or browser processes',
    }, null, 2));
    return;
  }

  const tempDir = mkdtempSync(path.join(tmpdir(), 'octopus-ai-learning-browser-'));
  const browserProfile = path.join(tempDir, 'pw-profile');
  const ownedChildren = [];

  try {
    if (mode === 'self-start') {
      smokeLog('starting local services', { frontendBaseUrl, backendBaseUrl });
      const configPath = writeBackendConfig(tempDir);
      ownedChildren.push(startBackend(configPath));
      ownedChildren.push(startFrontend());
      await waitForHttp(`${backendBaseUrl}/healthz`, 45000);
      await waitForHttp(frontendBaseUrl, 60000);
    } else {
      smokeLog('using external services', { frontendBaseUrl, backendBaseUrl });
      await waitForHttp(`${backendBaseUrl}/healthz`, 15000);
      await waitForHttp(frontendBaseUrl, 20000);
    }

    smokeLog('resolving admin token');
    const token = await resolveAdminToken();
    const authStorage = JSON.stringify({
      state: {
        token,
        expireAt: new Date(Date.now() + 3600_000).toISOString(),
        isAPIKeyAuth: false,
      },
      version: 0,
    });
    const navStorage = JSON.stringify({ state: { activeItem: 'ai', prevItem: 'home', direction: 1 }, version: 0 });
    const localeStorage = JSON.stringify({ state: { locale: 'zh-Hans' }, version: 0 });

    smokeLog('opening browser', { browserName, sessionName });
    await runPlaywrightCli(['open', frontendBaseUrl, '--browser', browserName, '--profile', browserProfile]);
    smokeLog('seeding browser storage');
    await runPlaywrightCli(['localstorage-set', 'auth-storage', authStorage]);
    await runPlaywrightCli(['localstorage-set', 'nav-storage', navStorage]);
    await runPlaywrightCli(['localstorage-set', 'octopus-settings', localeStorage]);
    await runPlaywrightCli(['reload']);

    smokeLog('waiting for learning section');
    const desktop = await waitForLearningSection();
    assert.equal(desktop.page, true, 'ai automation page should render on desktop');
    assert.equal(desktop.section, true, 'learning section should render on desktop');
    assert.equal(desktop.stage, true, 'learning stage card should render on desktop');
    assert.equal(desktop.controls, true, 'learning controls should render on desktop');
    assert.equal(desktop.presetCard, true, 'learning preset card should render on desktop');
    assert.equal(desktop.switchCard, true, 'learning switch card should render on desktop');
    assert.equal(desktop.summary, true, 'learning primary summary should render on desktop');
    assert.equal(desktop.secondarySummary, true, 'learning secondary summary should render on desktop');
    assert.equal(desktop.states, true, 'learning states grid should render on desktop');
    assert.ok(desktop.bodyText.includes('AI 自动化'), 'AI automation shell should be visible');
    assert.ok(desktop.bodyText.includes('动态路由学习'), 'learning title should be visible');
    assert.ok(desktop.bodyText.includes('学习预设'), 'preset title should be visible');
    assert.ok(desktop.bodyText.includes('本地 AI 学习开关'), 'learning switch title should be visible');

    smokeLog('checking mobile viewport');
    await runPlaywrightCli(['resize', '375', '1400']);
    const mobileEval = `() => ({ width: window.innerWidth, sectionRect: document.querySelector('[data-testid="ai-automation-learning-section"]')?.getBoundingClientRect().width ?? 0, summaryRect: document.querySelector('[data-testid="ai-automation-learning-state-summary"]')?.getBoundingClientRect().width ?? 0, bodyWidth: document.body.scrollWidth, viewport: document.documentElement.clientWidth })`;
    const mobileResult = await runPlaywrightCli(['eval', mobileEval], { raw: true });
    const mobile = parsePlaywrightJson(mobileResult.stdout);
    assert.equal(mobile.width, 375, 'mobile viewport should be set to 375px');
    assert.ok(mobile.sectionRect > 0, 'learning section should remain visible on mobile');
    assert.ok(mobile.summaryRect > 0, 'learning summary should remain visible on mobile');
    assert.ok(mobile.bodyWidth <= mobile.viewport + 24, 'mobile learning layout should not introduce large horizontal overflow');

    console.log(JSON.stringify({
      mode,
      frontend: frontendBaseUrl,
      backend: backendBaseUrl,
      mobileWidth: mobile.width,
      result: 'ai-automation-learning-browser-smoke passed',
    }, null, 2));
  } finally {
    try {
      await runPlaywrightCli(['close']);
    } catch {
    }

    for (const child of ownedChildren) {
      if (child && !child.killed) {
        child.kill('SIGTERM');
      }
    }

    rmSync(tempDir, { recursive: true, force: true });
  }
}

main().catch((error) => {
  console.error(error.stack || String(error));
  process.exit(1);
});
