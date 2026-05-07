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
const edgeLaunchPreset = process.env.OCTOPUS_UI_SMOKE_EDGE_LAUNCH_PRESET?.trim() || 'unknown';
const edgeProfileStrategy = process.env.OCTOPUS_UI_SMOKE_EDGE_PROFILE_STRATEGY?.trim() || 'unknown';

function normalizePageBootstrapStrategy(value) {
  const strategy = (value || 'attached-session').trim() || 'attached-session';
  const allowedStrategies = new Set(['auto', 'json-new', 'attached-session']);
  if (allowedStrategies.has(strategy)) {
    return strategy;
  }
  return 'attached-session';
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
    '  node scripts/verify-backup-browser-smoke-cdp.mjs --check-only',
    '  node scripts/verify-backup-browser-smoke-cdp.mjs',
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

function createBootstrapFailure(page, pageBootstrap) {
  const error = new Error(`CDP page bootstrap unavailable on this host for preset ${edgeLaunchPreset} (${page.pageMode})`);
  error.name = 'CdpPageBootstrapUnavailableError';
  error.cause = {
    edgeLaunchPreset,
    edgeProfileStrategy,
    pageStrategy: cdpPageBootstrapStrategy,
    bootstrapCommandOrder: cdpBootstrapCommandOrder,
    pageMode: page.pageMode,
    targetId: page.targetId,
    sessionId: page.sessionId ?? null,
    pageBootstrap,
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
    const bootstrapFailure = createBootstrapFailure(fallbackPage, fallbackBootstrap);
    bootstrapFailure.cause.fallbackFrom = page.pageMode;
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

async function waitForSelector(connection, sessionId, selector, timeoutMs = 6000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    const found = await evaluateCdp(connection, sessionId, `!!document.querySelector(${JSON.stringify(selector)})`);
    if (found) {
      return true;
    }
    await sleep(100);
  }
  return false;
}

async function clickSelectorViaCdp(connection, sessionId, selector) {
  return evaluateCdp(connection, sessionId, `(() => {
    const node = document.querySelector(${JSON.stringify(selector)});
    if (!node) return false;
    node.click();
    return true;
  })()`);
}

async function waitForBackupPageReady(connection, sessionId, timeoutMs = 45000) {
  const startedAt = Date.now();
  let lastBodyText = '';

  while (Date.now() - startedAt < timeoutMs) {
    try {
      const snapshot = await evaluateCdp(connection, sessionId, `(() => {
        const root = document.querySelector('[data-testid="backup-page"]');
        const historyTrigger = document.querySelector('[data-testid="backup-history-trigger"]');
        const fileInput = document.querySelector('[data-testid="backup-page"] input[type="file"]');
        return {
          backupPage: !!root,
          historyTrigger: !!historyTrigger,
          fileInput: !!fileInput,
          bodyText: document.body?.innerText || '',
        };
      })()`);
      lastBodyText = snapshot.bodyText || '';
      if (snapshot.backupPage && snapshot.historyTrigger && snapshot.fileInput) {
        return snapshot;
      }
    } catch {
    }
    await sleep(1000);
  }

  throw new Error(`Timed out waiting for backup page on ${frontendBaseUrl}. Last body text snippet: ${lastBodyText.slice(0, 200)}`);
}

async function waitForImportPreview(connection, sessionId, timeoutMs = 45000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    const snapshot = await evaluateCdp(connection, sessionId, `(() => ({
      pendingApply: !!document.querySelector('[data-testid="backup-pending-apply-panel"]'),
      importResult: !!document.querySelector('[data-testid="backup-import-result-panel"]'),
      compatibility: !!document.querySelector('[data-testid="backup-compatibility-panel"]'),
      previewToken: document.querySelector('[data-testid="backup-pending-apply-meta-preview-token"]')?.textContent?.trim() || '',
    }))()`);
    if (snapshot.pendingApply && snapshot.importResult && snapshot.compatibility) {
      return snapshot;
    }
    await sleep(500);
  }
  throw new Error('Timed out waiting for backup import preview panels');
}

async function waitForAppliedImport(connection, sessionId, timeoutMs = 45000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    const snapshot = await evaluateCdp(connection, sessionId, `(() => ({
      postImport: !!document.querySelector('[data-testid="backup-post-import-validation-panel"]'),
      pendingApplyGone: !document.querySelector('[data-testid="backup-pending-apply-panel"]'),
    }))()`);
    if (snapshot.postImport && snapshot.pendingApplyGone) {
      return snapshot;
    }
    await sleep(500);
  }
  throw new Error('Timed out waiting for applied import panels');
}

async function waitForRollbackPreview(connection, sessionId, timeoutMs = 30000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    const snapshot = await evaluateCdp(connection, sessionId, `(() => ({
      panel: !!document.querySelector('[data-testid="backup-rollback-preview-panel"]'),
      title: document.querySelector('[data-testid="backup-rollback-preview-title"]')?.textContent?.trim() || '',
      summaryCells: document.querySelectorAll('[data-testid^="backup-rollback-preview-summary-"]').length,
    }))()`);
    if (snapshot.panel && snapshot.summaryCells >= 3) {
      return snapshot;
    }
    await sleep(500);
  }
  throw new Error('Timed out waiting for rollback preview panel');
}

async function waitForRollbackPreviewCleared(connection, sessionId, timeoutMs = 15000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    const snapshot = await evaluateCdp(connection, sessionId, `(() => ({
      panelGone: !document.querySelector('[data-testid="backup-rollback-preview-panel"]'),
    }))()`);
    if (snapshot.panelGone) {
      return;
    }
    await sleep(300);
  }
  throw new Error('Timed out waiting for rollback preview panel to clear after scope change');
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

async function createBackupImportFile(token) {
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

  const fileBuffer = Buffer.from(await response.arrayBuffer());
  return {
    fileName: 'backup-smoke-import.json',
    fileBase64: fileBuffer.toString('base64'),
  };
}

async function setBackupFileInputViaCdp(connection, sessionId, importFile) {
  const result = await evaluateCdp(connection, sessionId, `(() => {
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
  })()`);
  assert.equal(result.found, true, 'backup file input should exist before upload');
  assert.equal(result.fileName, importFile.fileName, 'backup file input should receive the generated import file');
  assert.ok(result.size > 0, 'backup file input should receive a non-empty import file');
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
  const importFile = await createBackupImportFile(token);
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
    if (shouldAbortAfterBootstrap(pageBootstrap)) {
      const bootstrapFailure = createBootstrapFailure(page, pageBootstrap);
      trace('smoke-page:bootstrap-blocker', bootstrapFailure.cause);
      throw bootstrapFailure;
    }
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

    const desktop = await waitForBackupPageReady(page.pageConnection, page.sessionId);
    assert.equal(desktop.backupPage, true, 'backup page should render on desktop');
    assert.equal(desktop.historyTrigger, true, 'backup history trigger should render on desktop');
    assert.equal(desktop.fileInput, true, 'backup file input should render on desktop');
    assert.ok(desktop.bodyText.includes('导出快照'), 'backup export section should be visible');
    assert.ok(desktop.bodyText.includes('导入与预检'), 'backup import section should be visible');
    assert.ok(desktop.bodyText.includes('导入快照历史'), 'backup history section should be visible');

    await setBackupFileInputViaCdp(page.pageConnection, page.sessionId, importFile);
    const importClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="backup-import-button"]');
    assert.equal(importClicked, true, 'backup import button should be clickable');
    const preview = await waitForImportPreview(page.pageConnection, page.sessionId);
    assert.ok(preview.previewToken.length > 0, 'preview token should be shown after dry-run');

    const compatibilityToggled = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="backup-compatibility-toggle"]');
    assert.equal(compatibilityToggled, true, 'compatibility toggle should be clickable');
    const importMigrationToggled = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="backup-import-remaining-migration-trigger"]');
    assert.equal(importMigrationToggled, true, 'import remaining migration trigger should be clickable');
    const detailSnapshot = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
      signalList: !!document.querySelector('[data-testid="backup-compatibility-signal-list"]'),
      importRemainingPanel: !!document.querySelector('[data-testid="backup-import-remaining-migration-panel"]'),
      detailCount: document.querySelectorAll('[data-testid^="backup-compatibility-signal-"]').length,
    }))()`);
    assert.equal(detailSnapshot.signalList, true, 'compatibility signal list should open after toggle');
    assert.equal(detailSnapshot.importRemainingPanel, true, 'import remaining migration panel should open after toggle');
    assert.ok(detailSnapshot.detailCount >= 1, 'compatibility signal list should expose at least one signal item');

    const confirmSwitchClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="backup-apply-confirm-switch"]');
    assert.equal(confirmSwitchClicked, true, 'apply confirm switch should be clickable');
    const applyClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="backup-apply-same-import-button"]');
    assert.equal(applyClicked, true, 'apply same import button should be clickable after confirmation');
    await waitForAppliedImport(page.pageConnection, page.sessionId);

    const historyOpened = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="backup-history-trigger"]');
    assert.equal(historyOpened, true, 'backup history trigger should be clickable');
    await ensureSnapshotHistoryExists(token);
    const history = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
      panel: !!document.querySelector('[data-testid="backup-history-panel"]'),
      list: !!document.querySelector('[data-testid="backup-history-list"]'),
      items: document.querySelectorAll('[data-testid^="backup-history-item-"]').length,
    }))()`);
    assert.equal(history.panel, true, 'history panel should open after trigger');
    assert.equal(history.list, true, 'history list should render after opening history');
    assert.ok(history.items >= 1, 'history list should show at least one imported snapshot after apply');

    const rollbackScopeEditor = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
      editor: !!document.querySelector('[data-testid="backup-rollback-scope-editor"]'),
      summary: document.querySelector('[data-testid="backup-rollback-scope-current-summary"]')?.textContent?.trim() || '',
    }))()`);
    assert.equal(rollbackScopeEditor.editor, true, 'rollback scope editor should render after opening history');
    assert.ok(rollbackScopeEditor.summary.includes('回滚范围：整包快照恢复'), 'rollback scope summary should default to full snapshot restore before selective rollback');

    const selectiveRollbackClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="backup-rollback-selective-switch"]');
    assert.equal(selectiveRollbackClicked, true, 'rollback selective switch should be clickable');
    const rollbackRoutingClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="backup-rollback-scope-routing"]');
    const rollbackStatsClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="backup-rollback-scope-stats"]');
    const rollbackLogsClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="backup-rollback-scope-logs"]');
    assert.equal(rollbackRoutingClicked, true, 'rollback routing scope switch should be clickable');
    assert.equal(rollbackStatsClicked, true, 'rollback stats scope switch should be clickable');
    assert.equal(rollbackLogsClicked, true, 'rollback logs scope switch should be clickable');
    const rollbackScopeSelected = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
      grid: !!document.querySelector('[data-testid="backup-rollback-scope-grid"]'),
      summary: document.querySelector('[data-testid="backup-rollback-scope-current-summary"]')?.textContent?.trim() || '',
    }))()`);
    assert.equal(rollbackScopeSelected.grid, true, 'rollback scope grid should render after enabling selective rollback');
    assert.ok(rollbackScopeSelected.summary.includes('回滚范围：模型数据、API 密钥、系统设置'), 'rollback scope summary should reflect the narrowed rollback domains');

    const previewRollbackClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="backup-history-list"] [data-testid="backup-history-preview-button"]');
    assert.equal(previewRollbackClicked, true, 'history preview button should be clickable');
    const rollback = await waitForRollbackPreview(page.pageConnection, page.sessionId);
    assert.equal(rollback.panel, true, 'rollback preview should render after clicking preview');
    assert.equal(rollback.title, '回滚预览', 'rollback preview title should use localized copy');

    const rollbackDetails = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
      metaScopeText: document.querySelector('[data-testid="backup-rollback-preview-meta-scope"]')?.textContent?.trim() || '',
      metaScopeRaw: document.querySelector('[data-testid="backup-rollback-preview-meta-scope"]')?.getAttribute('data-raw-value') || '',
      routeDiffPanel: !!document.querySelector('[data-testid="backup-rollback-route-diff-panel"]'),
      routeDiffRows: document.querySelectorAll('[data-testid^="backup-rollback-route-diff-row-title-"]').length,
      routeDiffCurrent: document.querySelector('[data-testid="backup-rollback-route-diff-current-0"]')?.textContent?.trim() || '',
      routeDiffSnapshot: document.querySelector('[data-testid="backup-rollback-route-diff-snapshot-0"]')?.textContent?.trim() || '',
    }))()`);
    assert.ok(rollbackDetails.metaScopeText.includes('回滚范围：模型数据、API 密钥、系统设置'), 'rollback preview should echo the narrowed rollback scopes');
    assert.equal(rollbackDetails.metaScopeRaw, 'models,api_keys,settings', 'rollback preview raw scope should match the narrowed rollback scopes');
    assert.equal(rollbackDetails.routeDiffPanel, true, 'rollback route-diff compare panel should render after previewing a snapshot');
    assert.ok(rollbackDetails.routeDiffRows >= 1, 'rollback route-diff compare panel should expose at least one compare row');
    assert.ok(rollbackDetails.routeDiffCurrent.includes('current-primary:gpt-4o'), 'rollback route-diff current-state cell should stay visible after preview');
    assert.ok(rollbackDetails.routeDiffSnapshot.includes('snapshot-primary:gpt-4o'), 'rollback route-diff snapshot-state cell should stay visible after preview');

    const rollbackSettingsClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="backup-rollback-scope-settings"]');
    assert.equal(rollbackSettingsClicked, true, 'rollback settings scope switch should be clickable');
    await waitForRollbackPreviewCleared(page.pageConnection, page.sessionId);
    const rollbackScopeNarrowed = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
      summary: document.querySelector('[data-testid="backup-rollback-scope-current-summary"]')?.textContent?.trim() || '',
      previewGone: !document.querySelector('[data-testid="backup-rollback-preview-panel"]'),
    }))()`);
    assert.equal(rollbackScopeNarrowed.previewGone, true, 'rollback preview should clear after rollback scopes change');
    assert.ok(rollbackScopeNarrowed.summary.includes('回滚范围：模型数据、API 密钥'), 'rollback scope summary should update after narrowing settings away');

    const previewRollbackClickedAgain = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="backup-history-list"] [data-testid="backup-history-preview-button"]');
    assert.equal(previewRollbackClickedAgain, true, 'history preview button should stay clickable after rollback scopes change');
    await waitForRollbackPreview(page.pageConnection, page.sessionId);
    const rollbackDetailsNarrowed = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
      metaScopeText: document.querySelector('[data-testid="backup-rollback-preview-meta-scope"]')?.textContent?.trim() || '',
      metaScopeRaw: document.querySelector('[data-testid="backup-rollback-preview-meta-scope"]')?.getAttribute('data-raw-value') || '',
    }))()`);
    assert.ok(rollbackDetailsNarrowed.metaScopeText.includes('回滚范围：模型数据、API 密钥'), 'rollback preview should refresh with the latest narrowed rollback scopes');
    assert.equal(rollbackDetailsNarrowed.metaScopeRaw, 'models,api_keys', 'rollback preview raw scope should refresh after rollback scopes change');

    const historyMigrationToggled = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="backup-remaining-migration-trigger"]');
    assert.equal(historyMigrationToggled, true, 'history remaining migration trigger should be clickable');
    const historyDetails = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
      panel: !!document.querySelector('[data-testid="backup-remaining-migration-panel"]'),
      triggers: document.querySelectorAll('[data-testid^="backup-remaining-migration-section-trigger-"]').length,
    }))()`);
    assert.equal(historyDetails.panel, true, 'history remaining migration panel should open after toggle');
    assert.ok(historyDetails.triggers >= 1, 'history remaining migration panel should expose section triggers');

    await setCdpViewport(page.pageConnection, page.sessionId, { width: 375, height: 1400, mobile: true });
    await sleep(300);
    const mobile = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
      width: window.innerWidth,
      backupRect: document.querySelector('[data-testid="backup-page"]')?.getBoundingClientRect().width ?? 0,
      historyPanelRect: document.querySelector('[data-testid="backup-history-panel"]')?.getBoundingClientRect().width ?? 0,
      rollbackRect: document.querySelector('[data-testid="backup-rollback-preview-panel"]')?.getBoundingClientRect().width ?? 0,
      bodyWidth: document.body.scrollWidth,
      viewport: document.documentElement.clientWidth,
    }))()`);
    assert.equal(mobile.width, 375, 'mobile viewport should be set to 375px');
    assert.ok(mobile.backupRect > 0, 'backup page should remain visible on mobile');
    assert.ok(mobile.historyPanelRect > 0, 'history panel should remain visible on mobile');
    assert.ok(mobile.rollbackRect > 0, 'rollback preview should remain visible on mobile');
    assert.ok(mobile.bodyWidth <= mobile.viewport + 24, 'mobile backup page should not introduce large horizontal overflow');

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
      previewTokenVisible: preview.previewToken.length > 0,
      historyItems: history.items,
      rollbackSummaryCells: rollback.summaryCells,
      rollbackScopeRaw: rollbackDetails.metaScopeRaw,
      rollbackScopeNarrowedRaw: rollbackDetailsNarrowed.metaScopeRaw,
      mobileWidth: mobile.width,
      result: 'backup-browser-smoke-cdp passed',
    }, null, 2));
  } finally {
    await closeCdpPage(page);
  }
}

main().catch((error) => {
  writeDiagnostic(error);
  trace('smoke:error', { message: error.message, stack: error.stack || '' });
  console.error(error.stack || String(error));
  process.exitCode = 1;
});
