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
const sessionName = process.env.OCTOPUS_UI_SMOKE_SESSION || `octopus-backup-${process.pid}`;
const playwrightPackage = process.env.OCTOPUS_UI_SMOKE_PLAYWRIGHT_PACKAGE || '@playwright/cli';
const playwrightCliTimeoutMs = Number(process.env.OCTOPUS_UI_SMOKE_CLI_TIMEOUT_MS || 30000);
const forcedAdminPassword = process.env.OCTOPUS_UI_SMOKE_FORCED_ADMIN_PASSWORD || `${adminPassword}-rotated`;
const backendBootstrapAdminPassword = process.env.OCTOPUS_UI_SMOKE_BACKEND_BOOTSTRAP_ADMIN_PASSWORD?.trim() || (adminPassword === 'admin' ? forcedAdminPassword : adminPassword);
const allowForcedPasswordMutation = process.env.OCTOPUS_UI_SMOKE_ALLOW_FORCE_PASSWORD_CHANGE === 'true' || mode === 'self-start';

function smokeLog(message, details = {}) {
  const suffix = Object.keys(details).length > 0 ? ` ${JSON.stringify(details)}` : '';
  console.log(`[backup-smoke] ${message}${suffix}`);
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
    '  node scripts/verify-backup-browser-smoke.mjs --check-only',
    '  node scripts/verify-backup-browser-smoke.mjs --external',
    '  node scripts/verify-backup-browser-smoke.mjs --self-start',
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

function formatCollectedOutput(value, limit = 2000) {
  const text = String(value || '').trim();
  if (!text) {
    return '(empty)';
  }
  return text.length > limit ? text.slice(-limit) : text;
}

function assertProcessStillRunning(child, label) {
  if (!child || child.exitCode === null) {
    return;
  }
  const stdout = formatCollectedOutput(child.collected?.stdout);
  const stderr = formatCollectedOutput(child.collected?.stderr);
  throw new Error(`${label} exited early with code ${child.exitCode}\nSTDOUT:\n${stdout}\nSTDERR:\n${stderr}`);
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

async function readResponsePayload(response, pathName) {
  const text = await response.text();
  let payload;
  try {
    payload = text ? JSON.parse(text) : null;
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`Invalid JSON response for ${pathName}: ${message}\n${text}`);
  }
  if (!response.ok) {
    throw new Error(`Request failed for ${pathName}: ${response.status} ${response.statusText}\n${text}`);
  }
  return payload;
}

function unwrapPayloadData(payload, pathName) {
  if (!payload || typeof payload !== 'object' || !('data' in payload)) {
    throw new Error(`Expected response envelope with data for ${pathName}: ${JSON.stringify(payload)}`);
  }
  return payload.data;
}

async function loginAdmin(username, password) {
  const response = await fetch(`${backendBaseUrl}/api/v1/user/login`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ username, password, expire: 1440 }),
  });
  const payload = await readResponsePayload(response, '/api/v1/user/login');
  return unwrapPayloadData(payload, '/api/v1/user/login');
}

async function maybeResolveForcedPassword(token, mustChangePassword) {
  if (!mustChangePassword) {
    return { token, password: adminPassword };
  }

  if (!allowForcedPasswordMutation) {
    throw new Error('External backup smoke hit the first-login password-change gate. Provide a non-default admin credential/token, or rerun in self-start mode.');
  }

  smokeLog('forcing bootstrap admin password change for smoke session');
  await apiPost(token, '/api/v1/user/force-change-password', {
    new_password: forcedAdminPassword,
  });
  const refreshed = await loginAdmin(adminUsername, forcedAdminPassword);
  return { token: refreshed.token, password: forcedAdminPassword };
}

async function resolveAdminToken() {
  if (adminToken) {
    return adminToken;
  }
  const passwordCandidates = [...new Set([adminPassword, backendBootstrapAdminPassword].map((value) => value.trim()).filter(Boolean))];
  let lastError;
  for (const candidate of passwordCandidates) {
    try {
      const login = await loginAdmin(adminUsername, candidate);
      const session = await maybeResolveForcedPassword(login.token, login.must_change_password);
      return session.token;
    } catch (error) {
      lastError = error;
    }
  }
  throw lastError ?? new Error('Unable to resolve admin token');
}

async function apiPost(token, pathName, payload) {
  const response = await fetch(`${backendBaseUrl}${pathName}`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(payload),
  });
  const body = await readResponsePayload(response, pathName);
  return unwrapPayloadData(body, pathName);
}

async function apiGet(token, pathName) {
  const response = await fetch(`${backendBaseUrl}${pathName}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  const body = await readResponsePayload(response, pathName);
  return unwrapPayloadData(body, pathName);
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
    OCTOPUS_ADMIN_PASSWORD: backendBootstrapAdminPassword,
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

async function createBackupImportFile(token, tempDir) {
  await apiPost(token, '/api/v1/channel/create', {
    name: 'octopus-browser-backup-channel',
    type: 0,
    enabled: true,
    base_url: 'https://backup-smoke.example.com/v1',
    key_management_mode: 'pooled',
    key_routing_policy: 'round_robin',
    channel_key: 'sk-backup-smoke-upstream',
    models: ['gpt-4o'],
  });

  const response = await fetch(`${backendBaseUrl}/api/v1/setting/export?include_logs=false&include_stats=false&include_secrets=true&format=standard`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(`Export request failed: ${response.status} ${response.statusText}\n${text}`);
  }

  const filePath = path.join(tempDir, 'backup-smoke-import.json');
  const fileBuffer = Buffer.from(await response.arrayBuffer());
  writeFileSync(filePath, fileBuffer);
  return {
    filePath,
    fileName: 'backup-smoke-import.json',
    fileBase64: fileBuffer.toString('base64'),
  };
}

async function seedAuthenticatedBackupPage() {
  smokeLog('resolving admin token');
  const token = await resolveAdminToken();
  const authStorage = JSON.stringify({
    state: {
      token,
      expireAt: new Date(Date.now() + 3600_000).toISOString(),
      isAPIKeyAuth: false,
      mustChangePassword: false,
    },
    version: 0,
  });
  const navStorage = JSON.stringify({ state: { activeItem: 'setting', prevItem: 'home', direction: 1 }, version: 0 });
  const localeStorage = JSON.stringify({ state: { locale: 'zh-Hans' }, version: 0 });

  await runPlaywrightCli(['localstorage-set', 'auth-storage', authStorage]);
  await runPlaywrightCli(['localstorage-set', 'nav-storage', navStorage]);
  await runPlaywrightCli(['localstorage-set', 'octopus-settings', localeStorage]);
  await runPlaywrightCli(['reload']);

  return token;
}

async function scrollSettingsContainers() {
  const result = await runPlaywrightCli(['eval', `() => {
    const scrollables = Array.from(document.querySelectorAll('*')).filter((node) => {
      if (!(node instanceof HTMLElement)) return false;
      const style = window.getComputedStyle(node);
      return /(auto|scroll)/.test(style.overflowY) && node.scrollHeight - node.clientHeight > 24;
    });
    for (const node of scrollables) {
      node.scrollTop = node.scrollHeight;
    }
    window.scrollTo({ top: document.body.scrollHeight, behavior: 'instant' });
    return { scrollables: scrollables.length, bodyHeight: document.body.scrollHeight };
  }`], { raw: true });
  return parsePlaywrightJson(result.stdout);
}

async function waitForBackupPage(timeoutMs = 45000) {
  const startedAt = Date.now();
  let lastBodyText = '';
  let lastErrorMessage = '';
  const snapshotEval = `() => {
    const root = document.querySelector('[data-testid="backup-page"]');
    const historyTrigger = document.querySelector('[data-testid="backup-history-trigger"]');
    const fileInput = document.querySelector('[data-testid="backup-page"] input[type="file"]');
    return {
      backupPage: !!root,
      historyTrigger: !!historyTrigger,
      fileInput: !!fileInput,
      bodyText: document.body.innerText,
    };
  }`;

  while (Date.now() - startedAt < timeoutMs) {
    try {
      const result = await runPlaywrightCli(['eval', snapshotEval], { raw: true });
      const snapshot = parsePlaywrightJson(result.stdout);
      lastBodyText = snapshot.bodyText || '';
      lastErrorMessage = '';
      if (snapshot.backupPage && snapshot.historyTrigger && snapshot.fileInput) {
        return snapshot;
      }
      await scrollSettingsContainers();
    } catch (error) {
      lastErrorMessage = error instanceof Error ? error.message : String(error);
    }
    await sleep(1000);
  }

  const errorSuffix = lastErrorMessage ? ` Last eval error: ${lastErrorMessage}` : '';
  throw new Error(`Timed out waiting for backup page on ${frontendBaseUrl}. Last body text snippet: ${lastBodyText.slice(0, 200)}${errorSuffix}`);
}

async function setBackupFileInput(importFile) {
  const setFileEval = `() => {
    const input = document.querySelector('[data-testid="backup-page"] input[type="file"]');
    if (!(input instanceof HTMLInputElement)) {
      return { found: false };
    }
    const binary = window.atob(${JSON.stringify(importFile.fileBase64)});
    const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0));
    const file = new File([bytes], ${JSON.stringify(importFile.fileName)}, { type: 'application/json' });
    const dt = new DataTransfer();
    dt.items.add(file);
    input.files = dt.files;
    input.dispatchEvent(new Event('change', { bubbles: true }));
    return { found: true, fileName: input.files?.[0]?.name || '', size: input.files?.[0]?.size || 0 };
  }`;
  const result = parsePlaywrightJson((await runPlaywrightCli(['eval', setFileEval], { raw: true })).stdout);
  assert.equal(result.found, true, 'backup file input should exist before upload');
  assert.equal(result.fileName, importFile.fileName, 'backup file input should receive the generated import file');
  assert.ok(result.size > 0, 'backup file input should receive a non-empty import file');
}

async function waitForImportPreview(timeoutMs = 45000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    const result = await runPlaywrightCli(['eval', `() => ({ pendingApply: !!document.querySelector('[data-testid="backup-pending-apply-panel"]'), importResult: !!document.querySelector('[data-testid="backup-import-result-panel"]'), compatibility: !!document.querySelector('[data-testid="backup-compatibility-panel"]'), previewToken: document.querySelector('[data-testid="backup-pending-apply-meta-preview-token"]')?.textContent?.trim() || '' })`], { raw: true });
    const snapshot = parsePlaywrightJson(result.stdout);
    if (snapshot.pendingApply && snapshot.importResult && snapshot.compatibility) {
      return snapshot;
    }
    await sleep(500);
  }
  throw new Error('Timed out waiting for backup import preview panels');
}

async function waitForAppliedImport(timeoutMs = 45000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    const result = await runPlaywrightCli(['eval', `() => ({ postImport: !!document.querySelector('[data-testid="backup-post-import-validation-panel"]'), pendingApplyGone: !document.querySelector('[data-testid="backup-pending-apply-panel"]') })`], { raw: true });
    const snapshot = parsePlaywrightJson(result.stdout);
    if (snapshot.postImport && snapshot.pendingApplyGone) {
      return snapshot;
    }
    await sleep(500);
  }
  throw new Error('Timed out waiting for applied import panels');
}

async function waitForRollbackPreview(timeoutMs = 30000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    const result = await runPlaywrightCli(['eval', `() => ({ panel: !!document.querySelector('[data-testid="backup-rollback-preview-panel"]'), title: document.querySelector('[data-testid="backup-rollback-preview-title"]')?.textContent?.trim() || '', summaryCells: document.querySelectorAll('[data-testid^="backup-rollback-preview-summary-"]').length })`], { raw: true });
    const snapshot = parsePlaywrightJson(result.stdout);
    if (snapshot.panel && snapshot.summaryCells >= 3) {
      return snapshot;
    }
    await sleep(500);
  }
  throw new Error('Timed out waiting for rollback preview panel');
}

async function ensureSnapshotHistoryExists(token, timeoutMs = 30000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    const snapshots = await apiGet(token, '/api/v1/setting/import-snapshots');
    if (Array.isArray(snapshots) && snapshots.length > 0) {
      return snapshots;
    }
    await sleep(500);
  }
  throw new Error('Timed out waiting for import snapshot history after apply');
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
      note: 'check-only does not spawn backend, frontend, or browser processes',
    }, null, 2));
    return;
  }

  const tempDir = mkdtempSync(path.join(tmpdir(), 'octopus-backup-browser-'));
  const browserProfile = path.join(tempDir, 'pw-profile');
  const ownedChildren = [];

  try {
    if (mode === 'self-start') {
      smokeLog('starting local services', { frontendBaseUrl, backendBaseUrl });
      const configPath = writeBackendConfig(tempDir);
      const backendChild = startBackend(configPath);
      const frontendChild = startFrontend();
      ownedChildren.push(backendChild);
      ownedChildren.push(frontendChild);
      await sleep(1200);
      assertProcessStillRunning(backendChild, 'backend self-start process');
      assertProcessStillRunning(frontendChild, 'frontend self-start process');
      await waitForHttp(`${backendBaseUrl}/healthz`, 45000);
      assertProcessStillRunning(backendChild, 'backend self-start process');
      await waitForHttp(frontendBaseUrl, 60000);
      assertProcessStillRunning(frontendChild, 'frontend self-start process');
    } else {
      smokeLog('using external services', { frontendBaseUrl, backendBaseUrl });
      await waitForHttp(`${backendBaseUrl}/healthz`, 15000);
      await waitForHttp(frontendBaseUrl, 20000);
    }

    const token = await resolveAdminToken();
    const importFile = await createBackupImportFile(token, tempDir);

    smokeLog('opening browser', { browserName, sessionName });
    await runPlaywrightCli(['open', frontendBaseUrl, '--browser', browserName, '--profile', browserProfile]);
    smokeLog('seeding browser storage');
    await seedAuthenticatedBackupPage();

    smokeLog('waiting for backup page');
    const desktop = await waitForBackupPage();
    assert.equal(desktop.backupPage, true, 'backup page should render on desktop');
    assert.equal(desktop.historyTrigger, true, 'backup history trigger should render on desktop');
    assert.equal(desktop.fileInput, true, 'backup file input should render on desktop');
    assert.ok(desktop.bodyText.includes('导出快照'), 'backup export section should be visible');
    assert.ok(desktop.bodyText.includes('导入与预检'), 'backup import section should be visible');
    assert.ok(desktop.bodyText.includes('导入快照历史'), 'backup history section should be visible');

    smokeLog('setting backup import file');
    await setBackupFileInput(importFile);
    smokeLog('running import dry-run');
    await runPlaywrightCli(['click', '[data-testid="backup-import-button"]']);
    const preview = await waitForImportPreview();
    assert.ok(preview.previewToken.length > 0, 'preview token should be shown after dry-run');

    smokeLog('opening compatibility details and import tooling');
    await runPlaywrightCli(['click', '[data-testid="backup-compatibility-toggle"]']);
    await runPlaywrightCli(['click', '[data-testid="backup-import-remaining-migration-trigger"]']);
    const detailsResult = await runPlaywrightCli(['eval', `() => ({ signalList: !!document.querySelector('[data-testid="backup-compatibility-signal-list"]'), importRemainingPanel: !!document.querySelector('[data-testid="backup-import-remaining-migration-panel"]'), detailsCount: document.querySelectorAll('[data-testid^="backup-compatibility-signal-"]').length })`], { raw: true });
    const details = parsePlaywrightJson(detailsResult.stdout);
    assert.equal(details.signalList, true, 'compatibility signal list should open after toggle');
    assert.equal(details.importRemainingPanel, true, 'import remaining migration panel should open after toggle');
    assert.ok(details.detailsCount >= 1, 'compatibility signal list should expose at least one signal item');

    smokeLog('confirming and applying same import');
    await runPlaywrightCli(['click', '[data-testid="backup-apply-confirm-switch"]']);
    await runPlaywrightCli(['click', '[data-testid="backup-apply-same-import-button"]']);
    await waitForAppliedImport();

    smokeLog('opening history and previewing rollback snapshot');
    await runPlaywrightCli(['click', '[data-testid="backup-history-trigger"]']);
    await ensureSnapshotHistoryExists(token);
    const historyReady = await runPlaywrightCli(['eval', `() => ({ panel: !!document.querySelector('[data-testid="backup-history-panel"]'), list: !!document.querySelector('[data-testid="backup-history-list"]'), items: document.querySelectorAll('[data-testid^="backup-history-item-"]').length })`], { raw: true });
    const history = parsePlaywrightJson(historyReady.stdout);
    assert.equal(history.panel, true, 'history panel should open after trigger');
    assert.equal(history.list, true, 'history list should render after opening history');
    assert.ok(history.items >= 1, 'history list should show at least one imported snapshot after apply');

    await runPlaywrightCli(['click', '[data-testid="backup-history-list"] [data-testid="backup-history-preview-button"]']);
    const rollback = await waitForRollbackPreview();
    assert.equal(rollback.panel, true, 'rollback preview should render after clicking preview');
    assert.equal(rollback.title, '回滚预览', 'rollback preview title should use localized copy');

    smokeLog('opening history remaining migration section');
    await runPlaywrightCli(['click', '[data-testid="backup-remaining-migration-trigger"]']);
    const historyDetailsResult = await runPlaywrightCli(['eval', `() => ({ panel: !!document.querySelector('[data-testid="backup-remaining-migration-panel"]'), triggers: document.querySelectorAll('[data-testid^="backup-remaining-migration-section-trigger-"]').length })`], { raw: true });
    const historyDetails = parsePlaywrightJson(historyDetailsResult.stdout);
    assert.equal(historyDetails.panel, true, 'history remaining migration panel should open after toggle');
    assert.ok(historyDetails.triggers >= 1, 'history remaining migration panel should expose section triggers');

    smokeLog('checking mobile viewport');
    await runPlaywrightCli(['resize', '375', '1400']);
    const mobileResult = await runPlaywrightCli(['eval', `() => ({ width: window.innerWidth, backupRect: document.querySelector('[data-testid="backup-page"]')?.getBoundingClientRect().width ?? 0, historyPanelRect: document.querySelector('[data-testid="backup-history-panel"]')?.getBoundingClientRect().width ?? 0, rollbackRect: document.querySelector('[data-testid="backup-rollback-preview-panel"]')?.getBoundingClientRect().width ?? 0, bodyWidth: document.body.scrollWidth, viewport: document.documentElement.clientWidth })`], { raw: true });
    const mobile = parsePlaywrightJson(mobileResult.stdout);
    assert.equal(mobile.width, 375, 'mobile viewport should be set to 375px');
    assert.ok(mobile.backupRect > 0, 'backup page should remain visible on mobile');
    assert.ok(mobile.historyPanelRect > 0, 'history panel should remain visible on mobile');
    assert.ok(mobile.rollbackRect > 0, 'rollback preview should remain visible on mobile');
    assert.ok(mobile.bodyWidth <= mobile.viewport + 24, 'mobile backup page should not introduce large horizontal overflow');

    console.log(JSON.stringify({
      mode,
      frontend: frontendBaseUrl,
      backend: backendBaseUrl,
      previewTokenVisible: preview.previewToken.length > 0,
      historyItems: history.items,
      rollbackSummaryCells: rollback.summaryCells,
      mobileWidth: mobile.width,
      result: 'backup-browser-smoke passed',
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
