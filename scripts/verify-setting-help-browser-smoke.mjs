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
const sessionName = process.env.OCTOPUS_UI_SMOKE_SESSION || `octopus-setting-help-${process.pid}`;
const playwrightPackage = process.env.OCTOPUS_UI_SMOKE_PLAYWRIGHT_PACKAGE || '@playwright/cli';
const helpButtonSelector = process.env.OCTOPUS_UI_SMOKE_HELP_SELECTOR || 'button[data-help-hint-trigger="true"]';
const playwrightCliTimeoutMs = Number(process.env.OCTOPUS_UI_SMOKE_CLI_TIMEOUT_MS || 30000);
const modelProbeDefaultVisibleCount = 12;
const modelProbeSeedCount = 14;
const modelProbeSmokePrefix = `octopus-browser-model-probe-${process.pid}`;

function smokeLog(message, details = {}) {
  const suffix = Object.keys(details).length > 0 ? ` ${JSON.stringify(details)}` : '';
  console.log(`[setting-help-smoke] ${message}${suffix}`);
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
    '  node scripts/verify-setting-help-browser-smoke.mjs --check-only',
    '  node scripts/verify-setting-help-browser-smoke.mjs --external',
    '  node scripts/verify-setting-help-browser-smoke.mjs --self-start',
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

function unwrapApiPayload(payload) {
  if (payload && typeof payload === 'object' && 'data' in payload) {
    return payload.data;
  }
  return payload;
}

async function apiGet(token, path) {
  const response = await fetch(`${backendBaseUrl}${path}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    throw new Error(`Request failed for ${path}: ${response.status} ${response.statusText}\n${await response.text()}`);
  }

  return unwrapApiPayload(await response.json());
}

async function apiPost(token, path, payload) {
  const response = await fetch(`${backendBaseUrl}${path}`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    throw new Error(`Request failed for ${path}: ${response.status} ${response.statusText}\n${await response.text()}`);
  }

  return unwrapApiPayload(await response.json());
}

async function listModels(token) {
  const payload = await apiGet(token, '/api/v1/model/list');
  return Array.isArray(payload) ? payload : [];
}

function buildModelProbeSeedModels() {
  const policies = ['passive_only', 'sparse_single', 'sequential', 'concurrent'];
  return Array.from({ length: modelProbeSeedCount }, (_, index) => {
    const suffix = String(index + 1).padStart(2, '0');
    return {
      name: `${modelProbeSmokePrefix}-${suffix}`,
      canonical_name: `octopus-probe-canonical-${suffix}`,
      input: Number((0.1 + index * 0.01).toFixed(4)),
      output: Number((0.2 + index * 0.01).toFixed(4)),
      cache_read: Number((0.05 + index * 0.005).toFixed(4)),
      cache_write: Number((0.08 + index * 0.005).toFixed(4)),
      official_input: Number((0.12 + index * 0.01).toFixed(4)),
      official_output: Number((0.24 + index * 0.01).toFixed(4)),
      official_cache_read: Number((0.06 + index * 0.005).toFixed(4)),
      official_cache_write: Number((0.09 + index * 0.005).toFixed(4)),
      billing_mode: index % 2 === 0 ? 'per_token' : 'free',
      probe_policy: policies[index % policies.length],
      probe_interval_seconds: 300 + index * 15,
      probe_concurrency_limit: 1 + (index % 3),
    };
  });
}

async function prepareModelProbeVerification(token) {
  let seeded = false;
  if (mode === 'self-start') {
    for (const item of buildModelProbeSeedModels()) {
      await apiPost(token, '/api/v1/model/create', item);
    }
    seeded = true;
  }

  const allModels = (await listModels(token))
    .filter((model) => typeof model?.name === 'string' && model.name.trim().length > 0)
    .sort((a, b) => a.name.localeCompare(b.name));

  const scopedModels = seeded
    ? allModels.filter((model) => model.name.startsWith(modelProbeSmokePrefix))
    : allModels;

  if (scopedModels.length === 0) {
    return {
      seeded,
      interactive: false,
      totalAvailable: 0,
      expectedInitialVisibleCount: 0,
      expectedAfterShowMoreCount: 0,
      expectShowMore: false,
      expectShowMoreAfterOneClick: false,
      expectedSearchTerm: '',
      expectedSearchMatchName: '',
    };
  }

  const searchTarget = scopedModels[Math.min(scopedModels.length - 1, modelProbeSeedCount - 1)];
  const expectedInitialVisibleCount = Math.min(modelProbeDefaultVisibleCount, scopedModels.length);
  const expectedAfterShowMoreCount = Math.min(scopedModels.length, modelProbeDefaultVisibleCount * 2);

  return {
    seeded,
    interactive: true,
    totalAvailable: scopedModels.length,
    expectedInitialVisibleCount,
    expectedAfterShowMoreCount,
    expectShowMore: scopedModels.length > modelProbeDefaultVisibleCount,
    expectShowMoreAfterOneClick: scopedModels.length > modelProbeDefaultVisibleCount * 2,
    expectedSearchTerm: searchTarget.canonical_name || searchTarget.name,
    expectedSearchMatchName: searchTarget.name,
  };
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

async function getModelProbeSnapshot() {
  const snapshotEval = `() => { const defaultPath = document.querySelector('[data-testid="setting-model-probe-default-path"]'); const collapsedState = document.querySelector('[data-testid="setting-model-probe-collapsed-state"]'); const emptyState = document.querySelector('[data-testid="setting-model-probe-empty-state"]'); const modelList = document.querySelector('[data-testid="setting-model-probe-model-list"]'); const rows = modelList ? Array.from(modelList.querySelectorAll('[data-slot="accordion-item"]')).map((item) => (item.textContent || '').replace(/\s+/g, ' ').trim()).filter(Boolean) : []; const searchInput = document.querySelector('[data-testid="setting-model-probe-search"] input'); const toggleButton = document.querySelector('[data-testid="setting-model-probe-toggle"]'); const showMoreButton = document.querySelector('[data-testid="setting-model-probe-show-more"]'); const scrollRegion = document.querySelector('[data-testid="setting-model-probe-scroll-region"]'); return { defaultPathVisible: !!defaultPath, collapsedStateVisible: !!collapsedState, emptyStateVisible: !!emptyState, modelListVisible: !!modelList, visibleRowCount: rows.length, rowTexts: rows, searchValue: searchInput instanceof HTMLInputElement ? searchInput.value : '', toggleText: toggleButton ? (toggleButton.textContent || '').trim() : '', showMoreVisible: !!showMoreButton, showMoreText: showMoreButton ? (showMoreButton.textContent || '').trim() : '', scrollClientHeight: scrollRegion instanceof HTMLElement ? scrollRegion.clientHeight : 0, scrollHeight: scrollRegion instanceof HTMLElement ? scrollRegion.scrollHeight : 0 }; }`;
  const result = await runPlaywrightCli(['eval', snapshotEval], { raw: true });
  return parsePlaywrightJson(result.stdout);
}

async function waitForModelProbeState(predicate, timeoutMs, label) {
  const startedAt = Date.now();
  let lastSnapshot = null;
  while (Date.now() - startedAt < timeoutMs) {
    lastSnapshot = await getModelProbeSnapshot();
    if (predicate(lastSnapshot)) {
      return lastSnapshot;
    }
    await sleep(250);
  }

  throw new Error(`Timed out waiting for ${label}: ${JSON.stringify(lastSnapshot, null, 2)}`);
}

async function clickModelProbeButton(testId) {
  const clickEval = `() => { const button = document.querySelector('[data-testid="${testId}"]'); if (!(button instanceof HTMLButtonElement)) return false; button.click(); return true; }`;
  const result = await runPlaywrightCli(['eval', clickEval], { raw: true });
  assert.equal(parsePlaywrightJson(result.stdout), true, `${testId} should exist`);
}

async function setModelProbeSearchValue(value) {
  const searchEval = `() => { const input = document.querySelector('[data-testid="setting-model-probe-search"] input'); if (!(input instanceof HTMLInputElement)) return false; const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')?.set; if (!setter) return false; setter.call(input, ${JSON.stringify(value)}); input.dispatchEvent(new Event('input', { bubbles: true })); input.dispatchEvent(new Event('change', { bubbles: true })); return true; }`;
  const result = await runPlaywrightCli(['eval', searchEval], { raw: true });
  assert.equal(parsePlaywrightJson(result.stdout), true, 'model probe search input should exist');
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

async function waitForSettingCards(timeoutMs = 45000) {
  const startedAt = Date.now();
  let lastBodyText = '';
  let lastErrorMessage = '';
  const helpButtonSelectorLiteral = JSON.stringify(helpButtonSelector);
  const snapshotEval = `() => ({ llmPriceCard: !!document.querySelector('[data-testid="setting-llm-price-card"]'), dynamicCard: !!document.querySelector('[data-testid="setting-dynamic-routing-card"]'), circuitCard: !!document.querySelector('[data-testid="setting-circuit-breaker-card"]'), modelProbeCard: !!document.querySelector('[data-testid="setting-model-probe-card"]'), helpButtons: Array.from(document.querySelectorAll(${helpButtonSelectorLiteral})).length, bodyText: document.body.innerText })`;

  while (Date.now() - startedAt < timeoutMs) {
    try {
      const result = await runPlaywrightCli(['eval', snapshotEval], { raw: true });
      const snapshot = parsePlaywrightJson(result.stdout);
      lastBodyText = snapshot.bodyText || '';
      lastErrorMessage = '';
      if (snapshot.llmPriceCard && snapshot.dynamicCard && snapshot.circuitCard && snapshot.modelProbeCard) {
        return snapshot;
      }
    } catch (error) {
      lastErrorMessage = error instanceof Error ? error.message : String(error);
    }
    await sleep(1000);
  }

  const errorSuffix = lastErrorMessage ? ` Last eval error: ${lastErrorMessage}` : '';
  throw new Error(`Timed out waiting for settings cards on ${frontendBaseUrl}. Last body text snippet: ${lastBodyText.slice(0, 200)}${errorSuffix}`);
}

async function prepareHelpHintInteractionTargets(selectors) {
  const prepareEval = `() => { const selectors = ${JSON.stringify(selectors)}; return selectors.map(function(selector, index) { var button = document.querySelector(selector); var smokeId = 'setting-help-hint-' + index; if (!button) { return { selector: selector, targetSelector: '', found: false, hintId: '' }; } button.setAttribute('data-help-hint-smoke-target', smokeId); return { selector: selector, targetSelector: '[data-help-hint-smoke-target="' + smokeId + '"]', found: true, hintId: button.getAttribute('data-help-hint-id') || '' }; }); }`;
  const result = await runPlaywrightCli(['eval', prepareEval], { raw: true });
  return parsePlaywrightJson(result.stdout);
}

async function waitForBoundHelpHint(targetSelector, action) {
  const waitEval = `() => new Promise(function(resolve) { var targetSelector = ${JSON.stringify(targetSelector)}; var action = ${JSON.stringify(action)}; var button = document.querySelector(targetSelector); if (!button) { resolve({ found: false, focused: false, hintId: '', tooltipText: '', tooltipState: '', candidates: [] }); return; } var hintId = button.getAttribute('data-help-hint-id') || ''; var startedAt = Date.now(); var timeoutMs = 2000; function getCandidates() { return Array.from(document.querySelectorAll('[data-slot="help-hint-content"], [data-slot="tooltip-content"], [role="tooltip"]')).map(function(tooltip) { return { id: tooltip.id || '', hintId: tooltip.getAttribute('data-help-hint-id') || '', state: tooltip.getAttribute('data-state') || '', text: (tooltip.textContent || '').trim() }; }); } function getBoundTooltip() { var candidates = getCandidates(); for (var i = 0; i < candidates.length; i += 1) { var tooltip = candidates[i]; if (hintId && (tooltip.id === hintId || tooltip.hintId === hintId)) { return tooltip; } } return null; } function finish(tooltip) { resolve({ found: true, focused: document.activeElement === button, hintId: hintId, tooltipText: tooltip ? tooltip.text : '', tooltipState: tooltip ? tooltip.state : '', candidates: getCandidates() }); } function poll() { var tooltip = getBoundTooltip(); if (tooltip && tooltip.text) { finish(tooltip); return; } if (Date.now() - startedAt >= timeoutMs) { finish(tooltip); return; } window.setTimeout(poll, 50); } if (action === 'focus') { button.focus(); } poll(); })`;
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

async function checkHelpHintInteractions(selectors) {
  const targets = await prepareHelpHintInteractionTargets(selectors);
  const results = [];

  for (const target of targets) {
    if (!target.found || !target.targetSelector) {
      results.push({ selector: target.selector, found: false, focused: false, focusTooltipText: '', hoverTooltipText: '', hintId: '' });
      continue;
    }

    await resetHelpHintInteraction();
    await focusHelpHintWithKeyboard(target.targetSelector);
    const focusCheck = await waitForBoundHelpHint(target.targetSelector, 'inspect');

    await resetHelpHintInteraction();
    await runPlaywrightCli(['hover', target.targetSelector]);
    const hoverCheck = await waitForBoundHelpHint(target.targetSelector, 'hover');
    await resetHelpHintInteraction();

    results.push({
      selector: target.selector,
      targetSelector: target.targetSelector,
      found: focusCheck.found && hoverCheck.found,
      focused: focusCheck.focused,
      hintId: target.hintId || focusCheck.hintId || hoverCheck.hintId,
      focusTooltipText: focusCheck.tooltipText || '',
      focusTooltipState: focusCheck.tooltipState || '',
      hoverTooltipText: hoverCheck.tooltipText || '',
      hoverTooltipState: hoverCheck.tooltipState || '',
      focusCandidates: focusCheck.candidates || [],
      hoverCandidates: hoverCheck.candidates || [],
    });
  }

  return results;
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

  const tempDir = mkdtempSync(path.join(tmpdir(), 'octopus-setting-help-browser-'));
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
    const modelProbeVerification = await prepareModelProbeVerification(token);
    const authStorage = JSON.stringify({
      state: {
        token,
        expireAt: new Date(Date.now() + 3600_000).toISOString(),
        isAPIKeyAuth: false,
      },
      version: 0,
    });
    const navStorage = JSON.stringify({ state: { activeItem: 'setting', prevItem: 'home', direction: 1 }, version: 0 });
    const localeStorage = JSON.stringify({ state: { locale: 'zh-Hans' }, version: 0 });

    smokeLog('opening browser', { browserName, sessionName });
    await runPlaywrightCli(['open', frontendBaseUrl, '--browser', browserName, '--profile', browserProfile]);
    smokeLog('seeding browser storage');
    await runPlaywrightCli(['localstorage-set', 'auth-storage', authStorage]);
    await runPlaywrightCli(['localstorage-set', 'nav-storage', navStorage]);
    await runPlaywrightCli(['localstorage-set', 'octopus-settings', localeStorage]);
    await runPlaywrightCli(['reload']);

    smokeLog('waiting for settings cards');
    const desktop = await waitForSettingCards();
    assert.equal(desktop.llmPriceCard, true, 'llm price card should render on desktop');
    assert.equal(desktop.dynamicCard, true, 'dynamic routing card should render on desktop');
    assert.equal(desktop.circuitCard, true, 'circuit breaker card should render on desktop');
    assert.equal(desktop.modelProbeCard, true, 'model probe card should render on desktop');
    assert.ok(desktop.helpButtons >= 8, 'settings help buttons should be visible');
    assert.ok(desktop.bodyText.includes('\u6a21\u578b\u4ef7\u683c'), 'llm price title should be visible');
    assert.ok(desktop.bodyText.includes('\u52a8\u6001\u8def\u7531'), 'dynamic routing title should be visible');
    assert.ok(desktop.bodyText.includes('\u7194\u65ad\u5668\u914d\u7f6e'), 'circuit breaker title should be visible');
    assert.ok(desktop.bodyText.includes('\u6a21\u578b\u63a2\u6d4b\u7b56\u7565'), 'model probe title should be visible');

    const focusSelectors = [
      `[data-testid="setting-llm-price-card"] ${helpButtonSelector}`,
      `[data-testid="setting-dynamic-routing-card"] ${helpButtonSelector}`,
      `[data-testid="setting-circuit-breaker-card"] ${helpButtonSelector}`,
      `[data-testid="setting-model-probe-card"] ${helpButtonSelector}`,
    ];
    smokeLog('checking help hint focus and hover interactions');
    const interactionChecks = await checkHelpHintInteractions(focusSelectors);
    const interactionSummary = JSON.stringify(interactionChecks, null, 2);
    assert.ok(interactionChecks.every((item) => item.found && item.focused), `all targeted help buttons should be focusable\n${interactionSummary}`);
    assert.ok(interactionChecks.every((item) => item.focusTooltipText.length > 0), `focused help buttons should surface tooltip text\n${interactionSummary}`);
    assert.ok(interactionChecks.every((item) => item.hoverTooltipText.length > 0), `hovered help buttons should surface tooltip text\n${interactionSummary}`);

    smokeLog('checking model probe interactions', modelProbeVerification);
    const initialModelProbe = await getModelProbeSnapshot();
    assert.equal(initialModelProbe.defaultPathVisible, true, 'model probe default path summary should render by default');
    assert.equal(initialModelProbe.collapsedStateVisible, true, 'model probe should start in collapsed summary mode');
    assert.equal(initialModelProbe.modelListVisible, false, 'model probe rows should not render before expansion or search');

    if (modelProbeVerification.interactive) {
      await clickModelProbeButton('setting-model-probe-toggle');
      const expandedModelProbe = await waitForModelProbeState(
        (snapshot) => snapshot.modelListVisible && snapshot.visibleRowCount === modelProbeVerification.expectedInitialVisibleCount,
        15000,
        'model probe expanded rows',
      );
      assert.equal(expandedModelProbe.collapsedStateVisible, false, 'collapsed placeholder should disappear after expansion');
      if (modelProbeVerification.expectShowMore) {
        assert.equal(expandedModelProbe.showMoreVisible, true, 'show more should appear when more than 12 models are available');
        assert.ok(expandedModelProbe.scrollHeight > expandedModelProbe.scrollClientHeight, 'model probe should keep long lists inside the card scroll region');

        await clickModelProbeButton('setting-model-probe-show-more');
        const afterShowMore = await waitForModelProbeState(
          (snapshot) => snapshot.modelListVisible && snapshot.visibleRowCount === modelProbeVerification.expectedAfterShowMoreCount,
          15000,
          'model probe show more expansion',
        );
        assert.equal(afterShowMore.showMoreVisible, modelProbeVerification.expectShowMoreAfterOneClick, 'show more visibility should match the remaining model count after one expansion');
      }

      await setModelProbeSearchValue(modelProbeVerification.expectedSearchTerm);
      const searchedModelProbe = await waitForModelProbeState(
        (snapshot) => snapshot.modelListVisible && snapshot.rowTexts.some((text) => text.includes(modelProbeVerification.expectedSearchMatchName)),
        15000,
        'model probe canonical search results',
      );
      assert.equal(searchedModelProbe.searchValue, modelProbeVerification.expectedSearchTerm, 'model probe search input should keep the typed keyword');
      assert.equal(searchedModelProbe.visibleRowCount, 1, 'canonical search should narrow the list to the matching model');

      await setModelProbeSearchValue('__octopus_model_probe_unmatched__');
      const emptyModelProbe = await waitForModelProbeState(
        (snapshot) => snapshot.emptyStateVisible,
        15000,
        'model probe empty search state',
      );
      assert.equal(emptyModelProbe.modelListVisible, false, 'empty state should replace the accordion when no model matches the keyword');
    }

    smokeLog('checking mobile viewport');
    await runPlaywrightCli(['resize', '375', '1200']);
    const mobileEval = `() => ({ width: window.innerWidth, llmPriceRect: document.querySelector('[data-testid="setting-llm-price-card"]')?.getBoundingClientRect().width ?? 0, dynamicRect: document.querySelector('[data-testid="setting-dynamic-routing-card"]')?.getBoundingClientRect().width ?? 0, circuitRect: document.querySelector('[data-testid="setting-circuit-breaker-card"]')?.getBoundingClientRect().width ?? 0, modelProbeRect: document.querySelector('[data-testid="setting-model-probe-card"]')?.getBoundingClientRect().width ?? 0, bodyWidth: document.body.scrollWidth, viewport: document.documentElement.clientWidth })`;
    const mobileResult = await runPlaywrightCli(['eval', mobileEval], { raw: true });
    const mobile = parsePlaywrightJson(mobileResult.stdout);
    assert.equal(mobile.width, 375, 'mobile viewport should be set to 375px');
    assert.ok(mobile.llmPriceRect > 0, 'llm price card should remain visible on mobile');
    assert.ok(mobile.dynamicRect > 0, 'dynamic routing card should remain visible on mobile');
    assert.ok(mobile.circuitRect > 0, 'circuit breaker card should remain visible on mobile');
    assert.ok(mobile.modelProbeRect > 0, 'model probe card should remain visible on mobile');
    assert.ok(mobile.bodyWidth <= mobile.viewport + 24, 'mobile layout should not introduce large horizontal overflow');

    console.log(JSON.stringify({
      mode,
      frontend: frontendBaseUrl,
      backend: backendBaseUrl,
      desktopHelpButtons: desktop.helpButtons,
      interactionChecks: interactionChecks.length,
      modelProbeVerification,
      mobileWidth: mobile.width,
      result: 'setting-help-browser-smoke passed',
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
