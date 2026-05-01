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
const sessionName = process.env.OCTOPUS_UI_SMOKE_SESSION || `octopus-ccswitch-${process.pid}`;
const playwrightPackage = process.env.OCTOPUS_UI_SMOKE_PLAYWRIGHT_PACKAGE || '@playwright/cli';
const helpButtonSelector = process.env.OCTOPUS_UI_SMOKE_HELP_SELECTOR || 'button[data-help-hint-trigger="true"]';
const playwrightCliTimeoutMs = Number(process.env.OCTOPUS_UI_SMOKE_CLI_TIMEOUT_MS || 30000);
const ccswitchSmokeAPIKeyName = 'octopus-browser-ccswitch-key';
const ccswitchSmokeGroupName = 'octopus-browser-ccswitch-group';

function smokeLog(message, details = {}) {
  const suffix = Object.keys(details).length > 0 ? ` ${JSON.stringify(details)}` : '';
  console.log(`[ccswitch-smoke] ${message}${suffix}`);
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
    '  node scripts/verify-ccswitch-browser-smoke.mjs --check-only',
    '  node scripts/verify-ccswitch-browser-smoke.mjs --external',
    '  node scripts/verify-ccswitch-browser-smoke.mjs --self-start',
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

async function apiPost(token, pathName, payload) {
  const response = await fetch(`${backendBaseUrl}${pathName}`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(payload),
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(`Request failed for ${pathName}: ${response.status} ${response.statusText}\n${text}`);
  }
  return response.json();
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

async function ensureCCSwitchSeedData(token) {
  await apiPost(token, '/api/v1/group/create', {
    name: ccswitchSmokeGroupName,
    mode: 1,
    items: [],
  });

  await apiPost(token, '/api/v1/apikey/create', {
    name: ccswitchSmokeAPIKeyName,
    enabled: true,
    supported_models: ccswitchSmokeGroupName,
  });
}

async function seedAuthenticatedCCSwitchPage() {
  smokeLog('resolving admin token');
  const token = await resolveAdminToken();
  await ensureCCSwitchSeedData(token);
  const authStorage = JSON.stringify({
    state: {
      token,
      expireAt: new Date(Date.now() + 3600_000).toISOString(),
      isAPIKeyAuth: false,
    },
    version: 0,
  });
  const navStorage = JSON.stringify({ state: { activeItem: 'home', prevItem: 'setting', direction: -1 }, version: 0 });
  const localeStorage = JSON.stringify({ state: { locale: 'zh-Hans' }, version: 0 });

  await runPlaywrightCli(['localstorage-set', 'auth-storage', authStorage]);
  await runPlaywrightCli(['localstorage-set', 'nav-storage', navStorage]);
  await runPlaywrightCli(['localstorage-set', 'octopus-settings', localeStorage]);
  await runPlaywrightCli(['reload']);
}

async function waitForDocTrigger(timeoutMs = 45000) {
  const startedAt = Date.now();
  let lastBodyText = '';
  let lastErrorMessage = '';
  const snapshotEval = `() => ({ docTrigger: !!document.querySelector('[data-testid="navbar-doc-trigger"]'), bodyText: document.body.innerText })`;

  while (Date.now() - startedAt < timeoutMs) {
    try {
      const result = await runPlaywrightCli(['eval', snapshotEval], { raw: true });
      const snapshot = parsePlaywrightJson(result.stdout);
      lastBodyText = snapshot.bodyText || '';
      lastErrorMessage = '';
      if (snapshot.docTrigger) {
        return snapshot;
      }
    } catch (error) {
      lastErrorMessage = error instanceof Error ? error.message : String(error);
    }
    await sleep(1000);
  }

  const errorSuffix = lastErrorMessage ? ` Last eval error: ${lastErrorMessage}` : '';
  throw new Error(`Timed out waiting for doc trigger on ${frontendBaseUrl}. Last body text snippet: ${lastBodyText.slice(0, 200)}${errorSuffix}`);
}

async function waitForCCSwitchModal(timeoutMs = 20000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    const result = await runPlaywrightCli(['eval', `() => ({ modal: !!document.querySelector('[data-testid="doc-modal"]'), ccswitchPanel: !!document.querySelector('[data-testid="ccswitch-panel"]'), progressCard: !!document.querySelector('[data-testid="ccswitch-progress-card"]') })`], { raw: true });
    const snapshot = parsePlaywrightJson(result.stdout);
    if (snapshot.modal && snapshot.ccswitchPanel && snapshot.progressCard) {
      return snapshot;
    }
    await sleep(300);
  }
  throw new Error('Timed out waiting for CC Switch DocModal to open');
}

async function waitForSelector(selector, timeoutMs = 6000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    const result = await runPlaywrightCli(['eval', `() => !!document.querySelector(${JSON.stringify(selector)})`], { raw: true });
    if (parsePlaywrightJson(result.stdout) === true) {
      return true;
    }
    await sleep(120);
  }
  return false;
}

async function readBoundTooltip(scopeSelector) {
  const helpEval = `() => { var root = document.querySelector(${JSON.stringify(scopeSelector)}); var button = root?.querySelector(${JSON.stringify(helpButtonSelector)}); var hintId = button?.getAttribute('data-help-hint-id') || ''; var tooltips = Array.from(document.querySelectorAll('[data-slot="help-hint-content"], [role="tooltip"]')).map(function(node) { return { id: node.id || '', hintId: node.getAttribute('data-help-hint-id') || '', text: (node.textContent || '').trim() }; }); var current = tooltips.find(function(item) { return hintId && (item.id === hintId || item.hintId === hintId); }); return { hintId: hintId, tooltipText: current?.text || '', tooltipCount: tooltips.filter(function(item) { return item.text; }).length }; }`;
  const result = await runPlaywrightCli(['eval', helpEval], { raw: true });
  return parsePlaywrightJson(result.stdout);
}

async function waitForBoundHelpHint(targetSelector, action) {
  const waitEval = `() => new Promise(function(resolve) { var targetSelector = ${JSON.stringify(targetSelector)}; var action = ${JSON.stringify(action)}; var button = document.querySelector(targetSelector); if (!button) { resolve({ found: false, focused: false, hintId: '', tooltipText: '', tooltipState: '', candidates: [] }); return; } var hintId = button.getAttribute('data-help-hint-id') || ''; var startedAt = Date.now(); var timeoutMs = 2000; function getCandidates() { return Array.from(document.querySelectorAll('[data-slot="help-hint-content"], [data-slot="tooltip-content"], [role="tooltip"]')).map(function(tooltip) { return { id: tooltip.id || '', hintId: tooltip.getAttribute('data-help-hint-id') || '', state: tooltip.getAttribute('data-state') || '', text: (tooltip.textContent || '').trim() }; }); } function getBoundTooltip() { var candidates = getCandidates(); for (var i = 0; i < candidates.length; i += 1) { var tooltip = candidates[i]; if (hintId && (tooltip.id === hintId || tooltip.hintId === hintId)) { return tooltip; } } return null; } function finish(tooltip) { resolve({ found: true, focused: document.activeElement === button, hintId: hintId, tooltipText: tooltip ? tooltip.text : '', tooltipState: tooltip ? tooltip.state : '', candidates: getCandidates() }); } function poll() { var tooltip = getBoundTooltip(); if (tooltip && tooltip.text) { finish(tooltip); return; } if (Date.now() - startedAt >= timeoutMs) { finish(tooltip); return; } window.setTimeout(poll, 50); } if (action === 'focus') { button.focus(); } poll(); })()`;
  const result = await runPlaywrightCli(['eval', waitEval], { raw: true });
  return parsePlaywrightJson(result.stdout);
}

async function focusHelpHintWithKeyboard(targetSelector) {
  const prepareFocusEval = `() => { var targetSelector = ${JSON.stringify(targetSelector)}; Array.from(document.querySelectorAll('[data-help-hint-smoke-original-tabindex]')).forEach(function(item) { var original = item.getAttribute('data-help-hint-smoke-original-tabindex'); if (original === '__missing__') { item.removeAttribute('tabindex'); } else if (original !== null) { item.setAttribute('tabindex', original); } item.removeAttribute('data-help-hint-smoke-original-tabindex'); }); var target = document.querySelector(targetSelector); if (!target) { return false; } Array.from(document.querySelectorAll('a[href], button, input, select, textarea, [tabindex]')).forEach(function(item) { item.setAttribute('data-help-hint-smoke-original-tabindex', item.hasAttribute('tabindex') ? item.getAttribute('tabindex') : '__missing__'); item.tabIndex = item === target ? 0 : -1; }); target.scrollIntoView({ block: 'center', inline: 'center' }); if (!document.body.hasAttribute('tabindex')) { document.body.setAttribute('data-help-hint-smoke-body-tabindex-added', 'true'); document.body.tabIndex = -1; } var active = document.activeElement; if (active instanceof HTMLElement) active.blur(); document.body.focus(); return true; }`;
  const result = await runPlaywrightCli(['eval', prepareFocusEval], { raw: true });
  assert.equal(parsePlaywrightJson(result.stdout), true, `keyboard focus target should exist: ${targetSelector}`);
  await runPlaywrightCli(['press', 'Tab']);
}

async function resetHelpHintInteraction() {
  const resetEval = `() => { Array.from(document.querySelectorAll(${JSON.stringify(helpButtonSelector)})).forEach(function(button) { button.dispatchEvent(new MouseEvent('mouseout', { bubbles: true, relatedTarget: document.body })); button.dispatchEvent(new MouseEvent('mouseleave', { bubbles: true, relatedTarget: document.body })); }); Array.from(document.querySelectorAll('[data-help-hint-smoke-original-tabindex]')).forEach(function(item) { var original = item.getAttribute('data-help-hint-smoke-original-tabindex'); if (original === '__missing__') { item.removeAttribute('tabindex'); } else if (original !== null) { item.setAttribute('tabindex', original); } item.removeAttribute('data-help-hint-smoke-original-tabindex'); }); if (document.body.getAttribute('data-help-hint-smoke-body-tabindex-added') === 'true') { document.body.removeAttribute('tabindex'); document.body.removeAttribute('data-help-hint-smoke-body-tabindex-added'); } var active = document.activeElement; if (active instanceof HTMLElement) active.blur(); window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true })); return true; }`;
  await runPlaywrightCli(['eval', resetEval], { raw: true });
  await sleep(380);
}

async function setNameInputValue(selector, value) {
  const result = await runPlaywrightCli(['eval', `() => { var input = document.querySelector(${JSON.stringify(selector)}); if (!(input instanceof HTMLInputElement)) { return { found: false, value: '' }; } input.focus(); input.value = ${JSON.stringify(value)}; input.dispatchEvent(new Event('input', { bubbles: true })); input.dispatchEvent(new Event('change', { bubbles: true })); return { found: true, value: input.value }; }`], { raw: true });
  return parsePlaywrightJson(result.stdout);
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

  const tempDir = mkdtempSync(path.join(tmpdir(), 'octopus-ccswitch-browser-'));
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

    smokeLog('opening browser', { browserName, sessionName });
    await runPlaywrightCli(['open', frontendBaseUrl, '--browser', browserName, '--profile', browserProfile]);
    smokeLog('seeding browser storage');
    await seedAuthenticatedCCSwitchPage();

    smokeLog('waiting for doc trigger');
    const pageReady = await waitForDocTrigger();
    assert.ok(pageReady.docTrigger, 'doc trigger should exist before opening DocModal');

    smokeLog('opening doc modal and switching to CC Switch');
    await runPlaywrightCli(['click', '[data-testid="navbar-doc-trigger"]']);
    await runPlaywrightCli(['click', '[data-testid="doc-modal-tab-ccswitch"]']);
    await waitForCCSwitchModal();

    const desktopResult = await runPlaywrightCli(['eval', `() => ({ width: window.innerWidth, modalWidth: document.querySelector('[data-testid="doc-modal"]')?.getBoundingClientRect().width ?? 0, tabsWidth: document.querySelector('[data-testid="doc-modal-tabs"]')?.getBoundingClientRect().width ?? 0, progressSteps: document.querySelectorAll('[data-testid^="ccswitch-progress-step-"]').length, clientButtons: document.querySelectorAll('[data-testid^="ccswitch-client-"]').length, importDisabled: document.querySelector('[data-testid="ccswitch-import-button"]')?.hasAttribute('disabled') ?? false, modelLockedVisible: !!document.querySelector('[data-testid="ccswitch-model-locked-hint"]'), importLockedVisible: !!document.querySelector('[data-testid="ccswitch-import-locked-hint"]') })`], { raw: true });
    const desktop = parsePlaywrightJson(desktopResult.stdout);
    assert.ok(desktop.modalWidth > 0 && desktop.tabsWidth > 0, 'DocModal should stay visible on desktop');
    assert.equal(desktop.progressSteps, 4, 'CC Switch progress should show four steps');
    assert.equal(desktop.clientButtons, 2, 'CC Switch should render both client buttons');
    assert.equal(desktop.importDisabled, true, 'import button should stay disabled before setup completes');
    assert.equal(desktop.modelLockedVisible, true, 'model locked hint should render before API key selection');
    assert.equal(desktop.importLockedVisible, true, 'import locked hint should render before setup completes');

    smokeLog('checking focus and hover on CC Switch help hint');
    const ccswitchHelpSelector = '[data-testid="ccswitch-panel"] button[data-help-hint-trigger="true"]';
    await resetHelpHintInteraction();
    await focusHelpHintWithKeyboard(ccswitchHelpSelector);
    const focusSnapshot = await waitForBoundHelpHint(ccswitchHelpSelector, 'inspect');
    await resetHelpHintInteraction();
    await runPlaywrightCli(['hover', ccswitchHelpSelector]);
    const hoverSnapshot = await waitForBoundHelpHint(ccswitchHelpSelector, 'hover');
    await resetHelpHintInteraction();
    assert.equal(focusSnapshot.found && focusSnapshot.focused, true, 'CC Switch help hint should be keyboard focusable');
    assert.ok(focusSnapshot.tooltipText.length > 0, 'focused CC Switch help hint should show tooltip text');
    assert.ok(hoverSnapshot.tooltipText.length > 0, 'hovered CC Switch help hint should show tooltip text');

    smokeLog('selecting API key');
    await runPlaywrightCli(['click', '[data-testid="ccswitch-key-trigger"]']);
    await runPlaywrightCli(['click', '[data-slot="select-item"]']);
    const modelReady = await waitForSelector('[data-testid="ccswitch-model-trigger"]', 6000);
    assert.equal(modelReady, true, 'main model selector should remain available after API key selection');

    smokeLog('selecting main model');
    await runPlaywrightCli(['click', '[data-testid="ccswitch-model-trigger"]']);
    await runPlaywrightCli(['click', '[data-slot="select-item"]']);
    const nameInputReady = await waitForSelector('[data-testid="ccswitch-name-input"]', 6000);
    assert.equal(nameInputReady, true, 'name input should unlock after model selection');

    smokeLog('editing profile name');
    const nameState = await setNameInputValue('[data-testid="ccswitch-name-input"]', 'octopus_browser_ccswitch');
    assert.equal(nameState.found, true, 'name input should exist');
    assert.equal(nameState.value, 'octopus_browser_ccswitch', 'name input should accept custom value');

    const advancedReady = await waitForSelector('[data-testid="ccswitch-advanced-trigger"]', 6000);
    assert.equal(advancedReady, true, 'advanced mapping trigger should unlock after required steps');
    smokeLog('expanding advanced mapping');
    await runPlaywrightCli(['click', '[data-testid="ccswitch-advanced-trigger"]']);
    const advancedGridReady = await waitForSelector('[data-testid="ccswitch-advanced-grid"]', 6000);
    assert.equal(advancedGridReady, true, 'advanced mapping grid should expand on demand');

    const readyResult = await runPlaywrightCli(['eval', `() => ({ importDisabled: document.querySelector('[data-testid="ccswitch-import-button"]')?.hasAttribute('disabled') ?? true, importLockedVisible: !!document.querySelector('[data-testid="ccswitch-import-locked-hint"]') })`], { raw: true });
    const readyState = parsePlaywrightJson(readyResult.stdout);
    assert.equal(readyState.importDisabled, false, 'import button should enable after required steps');
    assert.equal(readyState.importLockedVisible, false, 'import locked hint should disappear after required steps');

    smokeLog('checking mobile viewport');
    await runPlaywrightCli(['resize', '375', '1200']);
    const mobileResult = await runPlaywrightCli(['eval', `() => ({ width: window.innerWidth, bodyWidth: document.body.scrollWidth, viewport: document.documentElement.clientWidth, modalWidth: document.querySelector('[data-testid="doc-modal"]')?.getBoundingClientRect().width ?? 0, tabsWidth: document.querySelector('[data-testid="doc-modal-tabs"]')?.getBoundingClientRect().width ?? 0, progressWidth: document.querySelector('[data-testid="ccswitch-progress-card"]')?.getBoundingClientRect().width ?? 0, importWidth: document.querySelector('[data-testid="ccswitch-import-button"]')?.getBoundingClientRect().width ?? 0 })`], { raw: true });
    const mobile = parsePlaywrightJson(mobileResult.stdout);
    assert.equal(mobile.width, 375, 'mobile viewport should be set to 375px');
    assert.ok(mobile.modalWidth > 0 && mobile.tabsWidth > 0 && mobile.progressWidth > 0 && mobile.importWidth > 0, 'CC Switch modal sections should remain visible on mobile');
    assert.ok(mobile.bodyWidth <= mobile.viewport + 24, 'mobile CC Switch modal should not introduce large horizontal overflow');

    console.log(JSON.stringify({
      mode,
      frontend: frontendBaseUrl,
      backend: backendBaseUrl,
      progressSteps: desktop.progressSteps,
      tooltipCount: hoverSnapshot.candidates?.filter?.((item) => item.text)?.length ?? 0,
      mobileWidth: mobile.width,
      result: 'ccswitch-browser-smoke passed',
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
