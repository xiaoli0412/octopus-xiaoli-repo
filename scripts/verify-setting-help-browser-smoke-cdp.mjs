import assert from 'node:assert/strict';
import fs from 'node:fs';

const argv = new Set(process.argv.slice(2));

function hasArg(flag) {
  return argv.has(flag);
}

function normalizeBaseUrl(value) {
  return value.replace(/\/$/, '');
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

async function fetchJson(url) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Request failed for ${url}: ${response.status} ${response.statusText}`);
  }
  return response.json();
}

const mode = hasArg('--check-only') ? 'check-only' : 'external';
const frontendBaseUrl = normalizeBaseUrl(process.env.OCTOPUS_UI_SMOKE_FRONTEND_URL || 'http://127.0.0.1:3101');
const backendBaseUrl = normalizeBaseUrl(process.env.OCTOPUS_UI_SMOKE_BACKEND_URL || 'http://127.0.0.1:18081');
const cdpBaseUrl = normalizeBaseUrl(process.env.OCTOPUS_UI_SMOKE_CDP_URL || 'http://127.0.0.1:9222');
const adminUsername = process.env.OCTOPUS_UI_SMOKE_ADMIN_USERNAME || 'admin';
const adminPassword = process.env.OCTOPUS_UI_SMOKE_ADMIN_PASSWORD || 'admin';
const commandTimeoutMs = Number(process.env.OCTOPUS_UI_SMOKE_CDP_COMMAND_TIMEOUT_MS || 10000);
const adminToken = process.env.OCTOPUS_UI_SMOKE_ADMIN_TOKEN?.trim() || '';
const cdpTraceFile = process.env.OCTOPUS_UI_SMOKE_CDP_TRACE_FILE?.trim() || '';
const cdpDiagnosticFile = process.env.OCTOPUS_UI_SMOKE_CDP_DIAGNOSTIC_FILE?.trim() || '';
const helpButtonSelector = process.env.OCTOPUS_UI_SMOKE_HELP_SELECTOR || 'button[data-help-hint-trigger="true"]';
const edgeLaunchPreset = process.env.OCTOPUS_UI_SMOKE_EDGE_LAUNCH_PRESET?.trim() || 'unknown';
const edgeProfileStrategy = process.env.OCTOPUS_UI_SMOKE_EDGE_PROFILE_STRATEGY?.trim() || 'unknown';
const modelProbeDefaultVisibleCount = 12;
const modelProbeSeedCount = 14;
const modelProbeSmokePrefix = `octopus-browser-model-probe-${process.pid}`;
function normalizePageBootstrapStrategy(value) {
  const strategy = (value || 'auto').trim() || 'auto';
  const allowedStrategies = new Set(['auto', 'json-new', 'attached-session']);
  if (allowedStrategies.has(strategy)) {
    return strategy;
  }

  return 'auto';
}

function normalizeBootstrapCommandOrder(value) {
  const order = (value || 'page-lifecycle-runtime').trim() || 'page-lifecycle-runtime';
  const allowedOrders = new Set(['page-lifecycle-runtime', 'runtime-page-lifecycle']);
  if (allowedOrders.has(order)) {
    return order;
  }

  return 'page-lifecycle-runtime';
}

const cdpPageBootstrapStrategy = normalizePageBootstrapStrategy(
  process.env.OCTOPUS_UI_SMOKE_CDP_PAGE_BOOTSTRAP_STRATEGY,
);
const cdpBootstrapCommandOrder = normalizeBootstrapCommandOrder(
  process.env.OCTOPUS_UI_SMOKE_CDP_BOOTSTRAP_COMMAND_ORDER,
);

function trace(message, details = undefined) {
  if (!cdpTraceFile) {
    return;
  }

  const timestamp = new Date().toISOString();
  const suffix = details === undefined ? '' : ` ${JSON.stringify(details)}`;
  fs.appendFileSync(cdpTraceFile, `[${timestamp}] ${message}${suffix}\n`, 'utf8');
}

function toSerializable(value) {
  if (value instanceof Error) {
    return {
      name: value.name,
      message: value.message,
      stack: value.stack || '',
      cause: value.cause === undefined ? undefined : toSerializable(value.cause),
    };
  }

  if (Array.isArray(value)) {
    return value.map((entry) => toSerializable(entry));
  }

  if (value && typeof value === 'object') {
    return Object.fromEntries(
      Object.entries(value).map(([key, entry]) => [key, toSerializable(entry)]),
    );
  }

  if (typeof value === 'bigint') {
    return String(value);
  }

  return value;
}

function classifyDiagnostic(error) {
  switch (error?.name) {
    case 'CdpPageBootstrapUnavailableError':
      return {
        classification: 'page_bootstrap_unavailable',
        hint: 'CDP attached-session bootstrap 仍然超时。下一轮应优先拿真实外部 Edge remote debugging 会话做对照，再决定是否继续调整页面断言。',
      };
    case 'CdpPageRuntimeUnavailableError':
      return {
        classification: 'page_runtime_unavailable',
        hint: '页面导航已响应，但 Runtime 或 Emulation 命令仍超时。当前应先按宿主级 CDP runtime 行为处理，直到外部会话证明并非同类问题。',
      };
    case 'CdpPageOpenUnavailableError':
      return {
        classification: 'page_open_unavailable',
        hint: '当前指定的页面打开策略无法建立可用的 CDP 页面会话。下一轮应切换 page bootstrap strategy 做对照，确认是 json-new 入口问题还是宿主级 bootstrap 问题。',
      };
    default:
      return {
        classification: 'generic_error',
        hint: '结合 Node stdout/stderr 与 CDP trace tail 继续缩小失败范围。',
      };
  }
}

function writeDiagnostic(error) {
  if (!cdpDiagnosticFile) {
    return;
  }

  try {
    const { classification, hint } = classifyDiagnostic(error);
    const cause = toSerializable(error?.cause);
    const diagnostic = {
      schemaVersion: 1,
      generatedAt: new Date().toISOString(),
      driver: 'edge-cdp',
      classification,
      errorName: error?.name || 'Error',
      message: error?.message || String(error),
      hint,
      frontendBaseUrl,
      backendBaseUrl,
      cdpBaseUrl,
      edgeLaunchPreset,
      edgeProfileStrategy,
      pageStrategy: cdpPageBootstrapStrategy,
      bootstrapCommandOrder: cdpBootstrapCommandOrder,
      commandTimeoutMs,
      pageMode: cause?.pageMode ?? null,
      fallbackFrom: cause?.fallbackFrom ?? null,
      targetId: cause?.targetId ?? null,
      sessionId: cause?.sessionId ?? null,
      summary: cause?.summary ?? null,
      cause,
      stack: error?.stack || '',
    };

    fs.writeFileSync(cdpDiagnosticFile, `${JSON.stringify(diagnostic, null, 2)}\n`, 'utf8');
    trace('smoke:diagnostic-written', {
      path: cdpDiagnosticFile,
      classification,
      errorName: diagnostic.errorName,
    });
  } catch (diagnosticError) {
    trace('smoke:diagnostic-write-failed', {
      path: cdpDiagnosticFile,
      message: diagnosticError instanceof Error ? diagnosticError.message : String(diagnosticError),
    });
  }
}

async function decodeWebSocketMessage(data) {
  if (typeof data === 'string') {
    return data;
  }

  if (data instanceof ArrayBuffer) {
    return Buffer.from(data).toString('utf8');
  }

  if (ArrayBuffer.isView(data)) {
    return Buffer.from(data.buffer, data.byteOffset, data.byteLength).toString('utf8');
  }

  if (typeof Blob !== 'undefined' && data instanceof Blob) {
    return data.text();
  }

  return String(data);
}

function printUsage() {
  console.log([
    'Usage:',
    '  node scripts/verify-setting-help-browser-smoke-cdp.mjs --check-only',
    '  node scripts/verify-setting-help-browser-smoke-cdp.mjs',
    '',
    'External prerequisites:',
    '  1. Backend is reachable',
    '  2. Frontend is reachable',
    '  3. Edge is running with --remote-debugging-port=9222',
    '',
    'Environment overrides:',
    '  OCTOPUS_UI_SMOKE_FRONTEND_URL=http://127.0.0.1:3101',
    '  OCTOPUS_UI_SMOKE_BACKEND_URL=http://127.0.0.1:18081',
    '  OCTOPUS_UI_SMOKE_CDP_URL=http://127.0.0.1:9222',
    '  OCTOPUS_UI_SMOKE_CDP_COMMAND_TIMEOUT_MS=15000',
    '  OCTOPUS_UI_SMOKE_CDP_PAGE_BOOTSTRAP_STRATEGY=auto|json-new|attached-session',
    '  OCTOPUS_UI_SMOKE_CDP_BOOTSTRAP_COMMAND_ORDER=page-lifecycle-runtime|runtime-page-lifecycle',
    '  OCTOPUS_UI_SMOKE_ADMIN_TOKEN=<jwt>',
  ].join('\n'));
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
    const text = await response.text();
    throw new Error(`Request failed for ${path}: ${response.status} ${response.statusText}\n${text}`);
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
    const text = await response.text();
    throw new Error(`Request failed for ${path}: ${response.status} ${response.statusText}\n${text}`);
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
  for (const item of buildModelProbeSeedModels()) {
    await apiPost(token, '/api/v1/model/create', item);
  }

  const allModels = (await listModels(token))
    .filter((model) => typeof model?.name === 'string' && model.name.trim().length > 0)
    .sort((a, b) => a.name.localeCompare(b.name));
  const scopedModels = allModels.filter((model) => model.name.startsWith(modelProbeSmokePrefix));

  if (scopedModels.length === 0) {
    return {
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
  return {
    interactive: true,
    totalAvailable: scopedModels.length,
    expectedInitialVisibleCount: Math.min(modelProbeDefaultVisibleCount, scopedModels.length),
    expectedAfterShowMoreCount: Math.min(scopedModels.length, modelProbeDefaultVisibleCount * 2),
    expectShowMore: scopedModels.length > modelProbeDefaultVisibleCount,
    expectShowMoreAfterOneClick: scopedModels.length > modelProbeDefaultVisibleCount * 2,
    expectedSearchTerm: searchTarget.canonical_name || searchTarget.name,
    expectedSearchMatchName: searchTarget.name,
  };
}

class CdpConnection {
  constructor(wsUrl) {
    this.wsUrl = wsUrl;
    this.ws = null;
    this.nextId = 0;
    this.pending = new Map();
    this.sessionMethodFallbacks = new Set([
      'Target.createTarget',
      'Target.attachToTarget',
      'Target.detachFromTarget',
      'Target.closeTarget',
    ]);
  }

  async connect() {
    trace('cdp-connect:start', { wsUrl: this.wsUrl });
    this.ws = new WebSocket(this.wsUrl);

    await new Promise((resolve, reject) => {
      const handleOpen = () => {
        cleanup();
        trace('cdp-connect:open', { wsUrl: this.wsUrl });
        resolve();
      };
      const handleError = (event) => {
        cleanup();
        trace('cdp-connect:error', { wsUrl: this.wsUrl, message: event.error?.message || 'CDP websocket failed to open' });
        reject(event.error ?? new Error('CDP websocket failed to open'));
      };
      const cleanup = () => {
        this.ws.removeEventListener('open', handleOpen);
        this.ws.removeEventListener('error', handleError);
      };

      this.ws.addEventListener('open', handleOpen, { once: true });
      this.ws.addEventListener('error', handleError, { once: true });
    });

    this.ws.addEventListener('message', async (event) => {
      try {
        const raw = await decodeWebSocketMessage(event.data);
        const payload = JSON.parse(raw);
        if (!payload.id) {
          trace('cdp-event', { method: payload.method || null, sessionId: payload.sessionId ?? null });
          return;
        }

        const pending = this.pending.get(payload.id);
        if (!pending) {
          trace('cdp-command:orphan-result', { id: payload.id, sessionId: payload.sessionId ?? null });
          return;
        }
        this.pending.delete(payload.id);

        if (payload.error) {
          trace('cdp-command:error', { id: payload.id, method: pending.method, sessionId: payload.sessionId ?? null, message: payload.error.message || JSON.stringify(payload.error) });
          pending.reject(new Error(payload.error.message || JSON.stringify(payload.error)));
          return;
        }

        trace('cdp-command:result', { id: payload.id, method: pending.method, sessionId: payload.sessionId ?? null });
        pending.resolve(payload.result);
      } catch (error) {
        trace('cdp-message:decode-failed', { message: error.message || String(error) });
      }
    });

    this.ws.addEventListener('close', () => {
      trace('cdp-connect:close', { wsUrl: this.wsUrl, pending: this.pending.size });
      for (const pending of this.pending.values()) {
        pending.reject(new Error('CDP websocket closed'));
      }
      this.pending.clear();
    });
  }

  async send(method, params = {}, sessionId) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error('CDP websocket is not connected');
    }

    const id = ++this.nextId;
    const payload = { id, method, params };
    if (sessionId && !this.sessionMethodFallbacks.has(method)) {
      payload.sessionId = sessionId;
    }

    trace('cdp-command:send', { id, method, sessionId: sessionId ?? null });

    const responsePromise = new Promise((resolve, reject) => {
      const timer = setTimeout(() => {
        this.pending.delete(id);
        trace('cdp-command:timeout', { id, method, sessionId: sessionId ?? null, timeoutMs: commandTimeoutMs });
        reject(new Error(`CDP command timed out: ${method}`));
      }, commandTimeoutMs);

      this.pending.set(id, {
        method,
        resolve: (value) => {
          clearTimeout(timer);
          resolve(value);
        },
        reject: (error) => {
          clearTimeout(timer);
          reject(error);
        },
      });
    });

    this.ws.send(JSON.stringify(payload));
    return responsePromise;
  }

  async close() {
    if (!this.ws || this.ws.readyState === WebSocket.CLOSED) {
      return;
    }

    trace('cdp-connect:closing', { wsUrl: this.wsUrl });
    await new Promise((resolve) => {
      const timer = setTimeout(resolve, 2000);
      const handleClose = () => {
        clearTimeout(timer);
        resolve();
      };

      this.ws.addEventListener('close', handleClose, { once: true });
      try {
        this.ws.close();
      } catch {
        clearTimeout(timer);
        resolve();
      }
    });
  }
}

async function connectCdp(wsUrl) {
  const connection = new CdpConnection(wsUrl);
  await connection.connect();
  return connection;
}

async function resolveBrowserWebSocketUrl() {
  const version = await fetchJson(`${cdpBaseUrl}/json/version`);
  if (!version.webSocketDebuggerUrl) {
    throw new Error(`CDP endpoint did not expose webSocketDebuggerUrl: ${JSON.stringify(version)}`);
  }

  trace('cdp-browser-version', {
    browser: version.Browser || '',
    protocolVersion: version['Protocol-Version'] || '',
    wsUrl: version.webSocketDebuggerUrl,
  });

  return version.webSocketDebuggerUrl;
}

async function evaluateCdp(connection, sessionId, expression, { awaitPromise = false } = {}) {
  const response = await connection.send('Runtime.evaluate', {
    expression,
    awaitPromise,
    returnByValue: true,
    userGesture: true,
  }, sessionId);

  if (response.exceptionDetails) {
    throw new Error(`CDP evaluation failed: ${response.exceptionDetails.text || expression}`);
  }

  return response.result?.value;
}

async function waitForDocumentReady(connection, sessionId, timeoutMs = 45000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    try {
      const readyState = await evaluateCdp(connection, sessionId, 'document.readyState');
      if (readyState === 'complete') {
        return;
      }
    } catch {
    }
    await sleep(500);
  }

  throw new Error('Timed out waiting for CDP page readiness');
}

async function setCdpViewport(connection, sessionId, { width, height, mobile }) {
  await connection.send('Emulation.setDeviceMetricsOverride', {
    width,
    height,
    deviceScaleFactor: 1,
    mobile,
    screenWidth: width,
    screenHeight: height,
  }, sessionId);
}

async function bootstrapCdpPageDomain(connection, sessionId) {
  const bootstrapCommandMap = {
    page: {
      name: 'Page.enable',
      run: () => connection.send('Page.enable', {}, sessionId),
    },
    lifecycle: {
      name: 'Page.setLifecycleEventsEnabled(true)',
      run: () => connection.send('Page.setLifecycleEventsEnabled', { enabled: true }, sessionId),
    },
    runtime: {
      name: 'Runtime.enable',
      run: () => connection.send('Runtime.enable', {}, sessionId),
    },
  };
  const bootstrapSequence = cdpBootstrapCommandOrder === 'runtime-page-lifecycle'
    ? ['runtime', 'page', 'lifecycle']
    : ['page', 'lifecycle', 'runtime'];
  const bootstrapCommands = bootstrapSequence.map((key) => bootstrapCommandMap[key]);
  const results = [];

  for (const command of bootstrapCommands) {
    const startedAt = Date.now();
    try {
      await command.run();
      results.push({ name: command.name, ok: true, elapsedMs: Date.now() - startedAt });
    } catch (error) {
      results.push({
        name: command.name,
        ok: false,
        elapsedMs: Date.now() - startedAt,
        message: error instanceof Error ? error.message : String(error),
      });
    }
  }

  trace('cdp-page-bootstrap:summary', { sessionId, bootstrapCommandOrder: cdpBootstrapCommandOrder, results });
  return results;
}

async function probeEarlyPageCommands(connection, sessionId) {
  const probes = [
    {
      name: 'Runtime.evaluate(1+1)',
      run: () => evaluateCdp(connection, sessionId, '1 + 1'),
    },
    {
      name: 'Page.navigate(about:blank#probe)',
      run: () => connection.send('Page.navigate', { url: 'about:blank#probe' }, sessionId),
    },
    {
      name: 'Emulation.setDeviceMetricsOverride(1x1)',
      run: () => connection.send('Emulation.setDeviceMetricsOverride', {
        width: 1,
        height: 1,
        deviceScaleFactor: 1,
        mobile: false,
        screenWidth: 1,
        screenHeight: 1,
      }, sessionId),
    },
  ];
  const results = [];

  for (const probe of probes) {
    const startedAt = Date.now();
    try {
      await probe.run();
      results.push({ name: probe.name, ok: true, elapsedMs: Date.now() - startedAt });
    } catch (error) {
      results.push({
        name: probe.name,
        ok: false,
        elapsedMs: Date.now() - startedAt,
        message: error instanceof Error ? error.message : String(error),
      });
    }
  }

  trace('cdp-page-probe:summary', { sessionId, results });
  return results;
}

function summarizePageCommandProbe(pageBootstrap, pageCommandProbe) {
  const bootstrapFailures = pageBootstrap.filter((entry) => !entry.ok);
  const probeFailures = pageCommandProbe.filter((entry) => !entry.ok);
  const navigateProbe = pageCommandProbe.find((entry) => entry.name === 'Page.navigate(about:blank#probe)');
  const runtimeProbe = pageCommandProbe.find((entry) => entry.name === 'Runtime.evaluate(1+1)');
  const emulationProbe = pageCommandProbe.find((entry) => entry.name === 'Emulation.setDeviceMetricsOverride(1x1)');

  return {
    bootstrapTimedOut: bootstrapFailures.length > 0 && bootstrapFailures.every((entry) => entry.message?.includes('timed out')),
    runtimeTimedOut: Boolean(runtimeProbe && !runtimeProbe.ok && runtimeProbe.message?.includes('timed out')),
    emulationTimedOut: Boolean(emulationProbe && !emulationProbe.ok && emulationProbe.message?.includes('timed out')),
    navigationSucceeded: Boolean(navigateProbe?.ok),
    failureCount: bootstrapFailures.length + probeFailures.length,
  };
}

function createProbeFailure(page, pageBootstrap, pageCommandProbe) {
  const summary = summarizePageCommandProbe(pageBootstrap, pageCommandProbe);
  const error = new Error(`CDP page runtime unavailable on this host for preset ${edgeLaunchPreset} (${page.pageMode})`);
    error.name = 'CdpPageRuntimeUnavailableError';
    error.cause = {
      edgeLaunchPreset,
      edgeProfileStrategy,
      pageStrategy: cdpPageBootstrapStrategy,
      bootstrapCommandOrder: cdpBootstrapCommandOrder,
      pageMode: page.pageMode,
      targetId: page.targetId,
    sessionId: page.sessionId ?? null,
    summary,
    pageBootstrap,
    pageCommandProbe,
  };
  return error;
}

function shouldAbortAfterProbe(pageBootstrap, pageCommandProbe) {
  const summary = summarizePageCommandProbe(pageBootstrap, pageCommandProbe);
  return summary.navigationSucceeded && summary.runtimeTimedOut && summary.emulationTimedOut;
}

function shouldAbortAfterBootstrap(pageBootstrap) {
  return pageBootstrap.length > 0 && pageBootstrap.every((entry) => !entry.ok && typeof entry.message === 'string' && entry.message.includes('timed out'));
}

function createPageOpenUnavailableError(message, cause = {}) {
  const error = new Error(message);
  error.name = 'CdpPageOpenUnavailableError';
  error.cause = {
    edgeLaunchPreset,
    edgeProfileStrategy,
    pageStrategy: cdpPageBootstrapStrategy,
    bootstrapCommandOrder: cdpBootstrapCommandOrder,
    ...cause,
  };
  return error;
}

async function navigateCdp(connection, sessionId, url) {
  await connection.send('Page.navigate', { url }, sessionId);
  await waitForDocumentReady(connection, sessionId);
}

async function reloadCdp(connection, sessionId) {
  await connection.send('Page.reload', {}, sessionId);
  await waitForDocumentReady(connection, sessionId);
}

async function tryOpenJsonNewPage(browserConnection) {
  try {
    const response = await fetch(`${cdpBaseUrl}/json/new?about:blank`, {
      method: 'PUT',
    });
    if (response.ok) {
      const pageInfo = await response.json();
      const pageWebSocketUrl = pageInfo.webSocketDebuggerUrl || '';
      const targetId = pageInfo.id || pageInfo.targetId || '';

      if (pageWebSocketUrl) {
        const pageConnection = await connectCdp(pageWebSocketUrl);
        trace('cdp-target:json-new', { targetId, pageWebSocketUrl: pageWebSocketUrl.replace(cdpWsBaseHostRegex(), '') });
        return {
          browserConnection,
          pageConnection,
          sessionId: undefined,
          targetId,
          pageMode: 'json-new',
          pageWebSocketUrl,
        };
      }

      trace('cdp-target:json-new-missing-websocket', { targetId, payload: pageInfo });
      return null;
    }

    trace('cdp-target:json-new-http-error', { status: response.status, statusText: response.statusText });
    return null;
  } catch (error) {
    trace('cdp-target:json-new-failed', { message: error instanceof Error ? error.message : String(error) });
    return null;
  }
}

async function openCdpPage() {
  const browserWsUrl = await resolveBrowserWebSocketUrl();
  const browserConnection = await connectCdp(browserWsUrl);
  trace('cdp-target:open-strategy', { strategy: cdpPageBootstrapStrategy });

  try {
    if (cdpPageBootstrapStrategy === 'attached-session') {
      return await createAttachedSessionPage(browserConnection);
    }

    const jsonNewPage = await tryOpenJsonNewPage(browserConnection);
    if (jsonNewPage) {
      return jsonNewPage;
    }

    if (cdpPageBootstrapStrategy === 'json-new') {
      throw createPageOpenUnavailableError(
        `CDP page open unavailable with explicit strategy ${cdpPageBootstrapStrategy}`,
        { pageMode: 'json-new' },
      );
    }

    return await createAttachedSessionPage(browserConnection);
  } catch (error) {
    await browserConnection.close();
    throw error;
  }
}

async function createAttachedSessionPage(browserConnection) {
  assert.ok(browserConnection, 'browserConnection is required for attached-session fallback');

  const { targetId } = await browserConnection.send('Target.createTarget', { url: 'about:blank' });
  trace('cdp-target:created', { targetId });
  const { sessionId } = await browserConnection.send('Target.attachToTarget', {
    targetId,
    flatten: true,
  });
  trace('cdp-target:attached', { targetId, sessionId });

  return {
    browserConnection,
    pageConnection: browserConnection,
    sessionId,
    targetId,
    pageMode: 'attached-session',
    pageWebSocketUrl: '',
  };
}

function shouldFallbackToAttachedSession(page, pageBootstrap) {
  return cdpPageBootstrapStrategy === 'auto'
    && page.pageMode === 'json-new'
    && pageBootstrap.length > 0
    && pageBootstrap.every((entry) => !entry.ok && typeof entry.message === 'string' && entry.message.includes('timed out'));
}

async function closeCdpPage(page, { closeBrowserConnection = true } = {}) {
  if (!page) {
    return;
  }

  if (page.sessionId) {
    try {
      await page.browserConnection.send('Target.detachFromTarget', { sessionId: page.sessionId });
      trace('cdp-target:detached', { targetId: page.targetId, sessionId: page.sessionId });
    } catch {
    }
  }

  if (page.targetId) {
    try {
      await page.browserConnection.send('Target.closeTarget', { targetId: page.targetId });
      trace('cdp-target:closed', { targetId: page.targetId });
    } catch {
    }
  }

  if (page.pageConnection && page.pageConnection !== page.browserConnection) {
    try {
      await page.pageConnection.close();
    } catch {
    }
  }

  if (closeBrowserConnection && page.browserConnection) {
    try {
      await page.browserConnection.close();
    } catch {
    }
  }
}

async function ensureUsableCdpPage(page, pageBootstrap) {
  if (!shouldFallbackToAttachedSession(page, pageBootstrap)) {
    return { page, pageBootstrap, fellBack: false };
  }

  trace('cdp-target:fallback-attached-session:start', {
    fromTargetId: page.targetId,
    fromPageMode: page.pageMode,
    bootstrap: pageBootstrap,
  });

  const browserConnection = page.browserConnection;
  await closeCdpPage(page, { closeBrowserConnection: false });

  const fallbackPage = await createAttachedSessionPage(browserConnection);
  trace('smoke-page:opened', {
    targetId: fallbackPage.targetId,
    sessionId: fallbackPage.sessionId,
    pageMode: fallbackPage.pageMode,
    fallbackFrom: page.pageMode,
  });

  const fallbackBootstrap = await bootstrapCdpPageDomain(fallbackPage.pageConnection, fallbackPage.sessionId);
  trace('smoke-page:bootstrap-complete', {
    targetId: fallbackPage.targetId,
    sessionId: fallbackPage.sessionId,
    pageMode: fallbackPage.pageMode,
    fallbackFrom: page.pageMode,
    results: fallbackBootstrap,
  });

  if (shouldAbortAfterBootstrap(fallbackBootstrap)) {
    const bootstrapFailure = new Error(`CDP page bootstrap unavailable on fallback attached session for preset ${edgeLaunchPreset}`);
    bootstrapFailure.name = 'CdpPageBootstrapUnavailableError';
      bootstrapFailure.cause = {
        edgeLaunchPreset,
        edgeProfileStrategy,
        pageStrategy: cdpPageBootstrapStrategy,
        bootstrapCommandOrder: cdpBootstrapCommandOrder,
        pageMode: fallbackPage.pageMode,
        fallbackFrom: page.pageMode,
      targetId: fallbackPage.targetId,
      sessionId: fallbackPage.sessionId ?? null,
      pageBootstrap: fallbackBootstrap,
    };
    trace('smoke-page:bootstrap-blocker', bootstrapFailure.cause);
    throw bootstrapFailure;
  }

  trace('cdp-target:fallback-attached-session:complete', {
    fromTargetId: page.targetId,
    toTargetId: fallbackPage.targetId,
    toSessionId: fallbackPage.sessionId,
  });

  return {
    page: fallbackPage,
    pageBootstrap: fallbackBootstrap,
    fellBack: true,
  };
}

function cdpWsBaseHostRegex() {
  return /^ws:\/\/[^/]+/;
}

async function setBrowserLocalStorageViaCdp(connection, sessionId, storageEntries) {
  const statements = Object.entries(storageEntries)
    .map(([key, value]) => `localStorage.setItem(${JSON.stringify(key)}, ${JSON.stringify(value)});`)
    .join('\n');

  await evaluateCdp(connection, sessionId, `(() => {\n${statements}\nreturn true;\n})()`);
}

async function waitForSettingCardsViaCdp(connection, sessionId, timeoutMs = 45000) {
  const startedAt = Date.now();
  let lastBodyText = '';

  while (Date.now() - startedAt < timeoutMs) {
    try {
      const snapshot = await evaluateCdp(connection, sessionId, `(() => ({
        llmPriceCard: !!document.querySelector('[data-testid="setting-llm-price-card"]'),
        dynamicCard: !!document.querySelector('[data-testid="setting-dynamic-routing-card"]'),
        circuitCard: !!document.querySelector('[data-testid="setting-circuit-breaker-card"]'),
        modelProbeCard: !!document.querySelector('[data-testid="setting-model-probe-card"]'),
        helpButtons: Array.from(document.querySelectorAll(${JSON.stringify(helpButtonSelector)})).length,
        bodyText: document.body?.innerText || '',
      }))()`);
      lastBodyText = snapshot.bodyText || '';

      if (snapshot.llmPriceCard && snapshot.dynamicCard && snapshot.circuitCard && snapshot.modelProbeCard) {
        return snapshot;
      }
    } catch {
    }
    await sleep(1000);
  }

  throw new Error(`Timed out waiting for settings cards on ${frontendBaseUrl}. Last body text snippet: ${lastBodyText.slice(0, 200)}`);
}

async function checkHelpHintsViaCdp(connection, sessionId, selectors) {
  return evaluateCdp(connection, sessionId, `(() => new Promise((resolve) => {
    const selectors = ${JSON.stringify(selectors)};
    const results = [];
    let index = 0;

    const inspectNext = () => {
      if (index >= selectors.length) {
        resolve(results);
        return;
      }

      const selector = selectors[index++];
      const button = document.querySelector(selector);
      if (!button) {
        results.push({ selector, found: false, focused: false, tooltipText: '' });
        inspectNext();
        return;
      }

      button.focus();
      setTimeout(() => {
        const tooltipText = Array.from(document.querySelectorAll('[role="tooltip"]'))
          .map((tooltip) => tooltip.textContent?.trim() || '')
          .find((text) => text.length > 0) || '';

        results.push({
          selector,
          found: true,
          focused: document.activeElement === button,
          tooltipText,
        });
        inspectNext();
      }, 180);
    };

    inspectNext();
  }))()`, { awaitPromise: true });
}

async function main() {
  if (hasArg('--help')) {
    printUsage();
    return;
  }

  if (mode === 'check-only') {
    console.log(JSON.stringify({
      mode,
      driver: 'edge-cdp',
      frontendBaseUrl,
      backendBaseUrl,
      cdpBaseUrl,
      edgeLaunchPreset,
      edgeProfileStrategy,
      pageStrategy: cdpPageBootstrapStrategy,
      bootstrapCommandOrder: cdpBootstrapCommandOrder,
      commandTimeoutMs,
      note: 'check-only does not spawn backend, frontend, or browser processes; edge-cdp expects an external browser with remote debugging enabled',
    }, null, 2));
    return;
  }

  await waitForHttp(`${backendBaseUrl}/healthz`, 15000);
  await waitForHttp(frontendBaseUrl, 20000);
  await waitForHttp(`${cdpBaseUrl}/json/version`, 10000);
  trace('smoke-prerequisites:ready', { backendBaseUrl, frontendBaseUrl, cdpBaseUrl, mode });

  const token = await resolveAdminToken();
  const modelProbeVerification = await prepareModelProbeVerification(token);
  trace('smoke-auth:token-resolved');
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

  const focusSelectors = [
    `[data-testid="setting-llm-price-card"] ${helpButtonSelector}`,
    `[data-testid="setting-dynamic-routing-card"] ${helpButtonSelector}`,
    `[data-testid="setting-circuit-breaker-card"] ${helpButtonSelector}`,
    `[data-testid="setting-model-probe-card"] ${helpButtonSelector}`,
  ];

  let page = await openCdpPage();
  try {
    trace('smoke-page:opened', { targetId: page.targetId, sessionId: page.sessionId, pageMode: page.pageMode });
    let pageBootstrap = await bootstrapCdpPageDomain(page.pageConnection, page.sessionId);
    trace('smoke-page:bootstrap-complete', {
      targetId: page.targetId,
      sessionId: page.sessionId,
      pageMode: page.pageMode,
      results: pageBootstrap,
    });
    const ensuredPage = await ensureUsableCdpPage(page, pageBootstrap);
    page = ensuredPage.page;
    pageBootstrap = ensuredPage.pageBootstrap;
    const pageCommandProbe = await probeEarlyPageCommands(page.pageConnection, page.sessionId);
    trace('smoke-page:probe-complete', {
      targetId: page.targetId,
      sessionId: page.sessionId,
      pageMode: page.pageMode,
      results: pageCommandProbe,
    });
    if (shouldAbortAfterProbe(pageBootstrap, pageCommandProbe)) {
      const probeFailure = createProbeFailure(page, pageBootstrap, pageCommandProbe);
      trace('smoke-page:probe-blocker', probeFailure.cause);
      throw probeFailure;
    }
    await setCdpViewport(page.pageConnection, page.sessionId, { width: 1440, height: 1200, mobile: false });
    await navigateCdp(page.pageConnection, page.sessionId, frontendBaseUrl);
    await setBrowserLocalStorageViaCdp(page.pageConnection, page.sessionId, {
      'auth-storage': authStorage,
      'nav-storage': navStorage,
      'octopus-settings': localeStorage,
    });
    await reloadCdp(page.pageConnection, page.sessionId);

    const desktop = await waitForSettingCardsViaCdp(page.pageConnection, page.sessionId);
    assert.equal(desktop.llmPriceCard, true, 'llm price card should render on desktop');
    assert.equal(desktop.dynamicCard, true, 'dynamic routing card should render on desktop');
    assert.equal(desktop.circuitCard, true, 'circuit breaker card should render on desktop');
    assert.equal(desktop.modelProbeCard, true, 'model probe card should render on desktop');
    assert.ok(desktop.helpButtons >= 8, 'settings help buttons should be visible');
    assert.ok(desktop.bodyText.includes('\u6a21\u578b\u4ef7\u683c'), 'llm price title should be visible');
    assert.ok(desktop.bodyText.includes('\u52a8\u6001\u8def\u7531'), 'dynamic routing title should be visible');
    assert.ok(desktop.bodyText.includes('\u7194\u65ad\u5668\u914d\u7f6e'), 'circuit breaker title should be visible');
    assert.ok(desktop.bodyText.includes('\u6a21\u578b\u63a2\u6d4b\u7b56\u7565'), 'model probe title should be visible');

    const focusChecks = await checkHelpHintsViaCdp(page.pageConnection, page.sessionId, focusSelectors);
    assert.ok(focusChecks.every((item) => item.found && item.focused), 'all targeted help buttons should be focusable');
    assert.ok(focusChecks.every((item) => item.tooltipText.length > 0), 'focused help buttons should surface tooltip text');

    const initialModelProbe = await getModelProbeSnapshotViaCdp(page.pageConnection, page.sessionId);
    assert.equal(initialModelProbe.defaultPathVisible, true, 'model probe default path summary should render by default');
    assert.equal(initialModelProbe.collapsedStateVisible, true, 'model probe should start in collapsed summary mode');
    assert.equal(initialModelProbe.modelListVisible, false, 'model probe rows should not render before expansion or search');

    if (modelProbeVerification.interactive) {
      await clickModelProbeButtonViaCdp(page.pageConnection, page.sessionId, 'setting-model-probe-toggle');
      const expandedModelProbe = await waitForModelProbeStateViaCdp(
        page.pageConnection,
        page.sessionId,
        (snapshot) => snapshot.modelListVisible && snapshot.visibleRowCount === modelProbeVerification.expectedInitialVisibleCount,
        15000,
        'model probe expanded rows',
      );
      assert.equal(expandedModelProbe.collapsedStateVisible, false, 'collapsed placeholder should disappear after expansion');
      if (modelProbeVerification.expectShowMore) {
        assert.equal(expandedModelProbe.showMoreVisible, true, 'show more should appear when more than 12 models are available');
        assert.ok(expandedModelProbe.scrollHeight > expandedModelProbe.scrollClientHeight, 'model probe should keep long lists inside the card scroll region');

        await clickModelProbeButtonViaCdp(page.pageConnection, page.sessionId, 'setting-model-probe-show-more');
        const afterShowMore = await waitForModelProbeStateViaCdp(
          page.pageConnection,
          page.sessionId,
          (snapshot) => snapshot.modelListVisible && snapshot.visibleRowCount === modelProbeVerification.expectedAfterShowMoreCount,
          15000,
          'model probe show more expansion',
        );
        assert.equal(afterShowMore.showMoreVisible, modelProbeVerification.expectShowMoreAfterOneClick, 'show more visibility should match the remaining model count after one expansion');
      }

      await setModelProbeSearchValueViaCdp(page.pageConnection, page.sessionId, modelProbeVerification.expectedSearchTerm);
      const searchedModelProbe = await waitForModelProbeStateViaCdp(
        page.pageConnection,
        page.sessionId,
        (snapshot) => snapshot.modelListVisible && snapshot.rowTexts.some((text) => text.includes(modelProbeVerification.expectedSearchMatchName)),
        15000,
        'model probe canonical search results',
      );
      assert.equal(searchedModelProbe.searchValue, modelProbeVerification.expectedSearchTerm, 'model probe search input should keep the typed keyword');
      assert.equal(searchedModelProbe.visibleRowCount, 1, 'canonical search should narrow the list to the matching model');

      await setModelProbeSearchValueViaCdp(page.pageConnection, page.sessionId, '__octopus_model_probe_unmatched__');
      const emptyModelProbe = await waitForModelProbeStateViaCdp(
        page.pageConnection,
        page.sessionId,
        (snapshot) => snapshot.emptyStateVisible,
        15000,
        'model probe empty search state',
      );
      assert.equal(emptyModelProbe.modelListVisible, false, 'empty state should replace the accordion when no model matches the keyword');
    }

    await setCdpViewport(page.pageConnection, page.sessionId, { width: 375, height: 1200, mobile: true });
    await sleep(300);
    const mobile = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
      width: window.innerWidth,
      llmPriceRect: document.querySelector('[data-testid="setting-llm-price-card"]')?.getBoundingClientRect().width ?? 0,
      dynamicRect: document.querySelector('[data-testid="setting-dynamic-routing-card"]')?.getBoundingClientRect().width ?? 0,
      circuitRect: document.querySelector('[data-testid="setting-circuit-breaker-card"]')?.getBoundingClientRect().width ?? 0,
      modelProbeRect: document.querySelector('[data-testid="setting-model-probe-card"]')?.getBoundingClientRect().width ?? 0,
      bodyWidth: document.body.scrollWidth,
      viewport: document.documentElement.clientWidth,
    }))()`);

    assert.equal(mobile.width, 375, 'mobile viewport should be set to 375px');
    assert.ok(mobile.llmPriceRect > 0, 'llm price card should remain visible on mobile');
    assert.ok(mobile.dynamicRect > 0, 'dynamic routing card should remain visible on mobile');
    assert.ok(mobile.circuitRect > 0, 'circuit breaker card should remain visible on mobile');
    assert.ok(mobile.modelProbeRect > 0, 'model probe card should remain visible on mobile');
    assert.ok(mobile.bodyWidth <= mobile.viewport + 24, 'mobile layout should not introduce large horizontal overflow');

    console.log(JSON.stringify({
      mode,
      driver: 'edge-cdp',
      frontend: frontendBaseUrl,
      backend: backendBaseUrl,
      cdp: cdpBaseUrl,
      pageStrategy: cdpPageBootstrapStrategy,
      bootstrapCommandOrder: cdpBootstrapCommandOrder,
      commandTimeoutMs,
      pageMode: page.pageMode,
      pageBootstrap,
      pageCommandProbe,
      desktopHelpButtons: desktop.helpButtons,
      modelProbeVerification,
      mobileWidth: mobile.width,
      result: 'setting-help-browser-smoke-cdp passed',
    }, null, 2));
  } finally {
    await closeCdpPage(page);
  }
}

async function getModelProbeSnapshotViaCdp(connection, sessionId) {
  return evaluateCdp(connection, sessionId, `(() => { const defaultPath = document.querySelector('[data-testid="setting-model-probe-default-path"]'); const collapsedState = document.querySelector('[data-testid="setting-model-probe-collapsed-state"]'); const emptyState = document.querySelector('[data-testid="setting-model-probe-empty-state"]'); const modelList = document.querySelector('[data-testid="setting-model-probe-model-list"]'); const rows = modelList ? Array.from(modelList.querySelectorAll('[data-slot="accordion-item"]')).map((item) => (item.textContent || '').replace(/\s+/g, ' ').trim()).filter(Boolean) : []; const searchInput = document.querySelector('[data-testid="setting-model-probe-search"] input'); const toggleButton = document.querySelector('[data-testid="setting-model-probe-toggle"]'); const showMoreButton = document.querySelector('[data-testid="setting-model-probe-show-more"]'); const scrollRegion = document.querySelector('[data-testid="setting-model-probe-scroll-region"]'); return { defaultPathVisible: !!defaultPath, collapsedStateVisible: !!collapsedState, emptyStateVisible: !!emptyState, modelListVisible: !!modelList, visibleRowCount: rows.length, rowTexts: rows, searchValue: searchInput instanceof HTMLInputElement ? searchInput.value : '', toggleText: toggleButton ? (toggleButton.textContent || '').trim() : '', showMoreVisible: !!showMoreButton, showMoreText: showMoreButton ? (showMoreButton.textContent || '').trim() : '', scrollClientHeight: scrollRegion instanceof HTMLElement ? scrollRegion.clientHeight : 0, scrollHeight: scrollRegion instanceof HTMLElement ? scrollRegion.scrollHeight : 0 }; })()`);
}

async function waitForModelProbeStateViaCdp(connection, sessionId, predicate, timeoutMs, label) {
  const startedAt = Date.now();
  let lastSnapshot = null;

  while (Date.now() - startedAt < timeoutMs) {
    lastSnapshot = await getModelProbeSnapshotViaCdp(connection, sessionId);
    if (predicate(lastSnapshot)) {
      return lastSnapshot;
    }
    await sleep(250);
  }

  throw new Error(`Timed out waiting for ${label}: ${JSON.stringify(lastSnapshot, null, 2)}`);
}

async function clickModelProbeButtonViaCdp(connection, sessionId, testId) {
  const clicked = await evaluateCdp(connection, sessionId, `(() => { const button = document.querySelector('[data-testid="${testId}"]'); if (!(button instanceof HTMLButtonElement)) return false; button.click(); return true; })()`);
  assert.equal(clicked, true, `${testId} should exist`);
}

async function setModelProbeSearchValueViaCdp(connection, sessionId, value) {
  const changed = await evaluateCdp(connection, sessionId, `(() => { const input = document.querySelector('[data-testid="setting-model-probe-search"] input'); if (!(input instanceof HTMLInputElement)) return false; const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')?.set; if (!setter) return false; setter.call(input, ${JSON.stringify(value)}); input.dispatchEvent(new Event('input', { bubbles: true })); input.dispatchEvent(new Event('change', { bubbles: true })); return true; })()`);
  assert.equal(changed, true, 'model probe search input should exist');
}

main().catch((error) => {
  writeDiagnostic(error);
  trace('smoke:error', { message: error.message, stack: error.stack || '' });
  console.error(error.stack || String(error));
  process.exitCode = 1;
});
