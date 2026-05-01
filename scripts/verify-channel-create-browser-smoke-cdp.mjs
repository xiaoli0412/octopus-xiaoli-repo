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
const smokeScenario = (process.env.OCTOPUS_UI_SMOKE_SCENARIO || 'channel-create').trim() || 'channel-create';
const modelLayoutSeedPrefix = 'octopus-browser-model-';
const homeSmokeChannelName = 'octopus-browser-home-channel';
const homeSmokeGroupName = 'octopus-browser-home-group';
const homeSmokeAPIKeyName = 'octopus-browser-home-key';
const homeSmokeMockPort = Number(process.env.OCTOPUS_UI_SMOKE_HOME_MOCK_PORT || 19191);
const aiLearningSmokeChannelName = 'octopus-browser-ai-learning-channel';
const aiLearningSmokeGroupName = 'octopus-browser-ai-learning-group';
const aiLearningSmokeAPIKeyName = 'octopus-browser-ai-learning-key';
const aiLearningSmokeMockPort = Number(process.env.OCTOPUS_UI_SMOKE_AI_LEARNING_MOCK_PORT || 19171);
const ccswitchSmokeAPIKeyName = 'octopus-browser-ccswitch-key';
const ccswitchSmokeGroupName = 'octopus-browser-ccswitch-group';
const channelPageSmokeName = 'octopus-browser-channel-page';
const channelPageSmokeModel = 'octopus-browser-channel-model';
const channelFetchModelMockPort = Number(process.env.OCTOPUS_UI_SMOKE_CHANNEL_FETCH_MODEL_MOCK_PORT || 19181);
const channelFetchModelList = ['octopus-browser-model-a', 'octopus-browser-model-b', 'octopus-browser-model-c'];
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

function isCdpBootstrapUnavailableMessage(message) {
  if (typeof message !== 'string') {
    return false;
  }

  const normalized = message.toLowerCase();
  return normalized.includes('timed out')
    || normalized.includes('websocket closed')
    || normalized.includes('not connected')
    || normalized.includes('session closed')
    || normalized.includes('target closed');
}

function didAllCdpCommandsFailWithBootstrapUnavailable(entries) {
  return Array.isArray(entries)
    && entries.length > 0
    && entries.every((entry) => !entry.ok && isCdpBootstrapUnavailableMessage(entry.message));
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
    '  node scripts/verify-channel-create-browser-smoke-cdp.mjs --check-only',
    '  node scripts/verify-channel-create-browser-smoke-cdp.mjs',
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
  return response.json();
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
  return response.json();
}

async function apiGatewayPost(apiKey, path, payload) {
  const response = await fetch(`${backendBaseUrl}${path}`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
      Authorization: `Bearer ${apiKey}`,
    },
    body: JSON.stringify(payload),
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(`Gateway request failed for ${path}: ${response.status} ${response.statusText}\n${text}`);
  }
  return response.json();
}

async function ensureModelLayoutSeedData(token) {
  const seedModels = [
    {
      name: `${modelLayoutSeedPrefix}alpha`,
      canonical_name: 'octopus-alpha',
      input: 1.23,
      output: 2.34,
      cache_read: 0.45,
      cache_write: 0.67,
      official_input: 1.5,
      official_output: 2.8,
      official_cache_read: 0.6,
      official_cache_write: 0.9,
      billing_mode: 'per_token',
    },
    {
      name: `${modelLayoutSeedPrefix}beta`,
      canonical_name: 'octopus-beta',
      input: 0,
      output: 0,
      cache_read: 0,
      cache_write: 0,
      official_input: 0,
      official_output: 0,
      official_cache_read: 0,
      official_cache_write: 0,
      billing_mode: 'free',
    },
  ];

  for (const item of seedModels) {
    await apiPost(token, '/api/v1/model/create', item);
  }
}

async function createHomeSmokeMockServer(port) {
  const { createServer } = await import('node:http');
  return new Promise((resolve, reject) => {
    const server = createServer(async (req, res) => {
      if (req.method !== 'POST') {
        res.writeHead(404);
        res.end();
        return;
      }

      const chunks = [];
      for await (const chunk of req) {
        chunks.push(chunk);
      }

      let payload = {};
      try {
        payload = JSON.parse(Buffer.concat(chunks).toString('utf8') || '{}');
      } catch {
        payload = {};
      }

      const modelName = typeof payload.model === 'string' && payload.model ? payload.model : homeSmokeGroupName;
      const response = {
        id: 'chatcmpl-home-smoke-1',
        object: 'chat.completion',
        created: 1713436800,
        model: modelName,
        choices: [{ index: 0, message: { role: 'assistant', content: 'home-smoke-ok' }, finish_reason: 'stop' }],
        usage: { prompt_tokens: 12, completion_tokens: 6, total_tokens: 18 },
      };

      const encoded = Buffer.from(JSON.stringify(response), 'utf8');
      res.writeHead(200, {
        'Content-Type': 'application/json',
        'Content-Length': String(encoded.length),
      });
      res.end(encoded);
    });

    server.once('error', reject);
    server.listen(port, '127.0.0.1', () => {
      server.off('error', reject);
      resolve(server);
    });
  });
}

async function createChannelFetchModelMockServer(port) {
  const { createServer } = await import('node:http');
  return new Promise((resolve, reject) => {
    const requests = [];
    const server = createServer(async (req, res) => {
      if (req.method !== 'GET' || req.url !== '/models') {
        res.writeHead(404);
        res.end();
        return;
      }

      requests.push({
        authorization: req.headers.authorization || '',
        time: Date.now(),
      });

      const encoded = Buffer.from(JSON.stringify({
        data: channelFetchModelList.map((id) => ({ id })),
      }), 'utf8');
      res.writeHead(200, {
        'Content-Type': 'application/json',
        'Content-Length': String(encoded.length),
      });
      res.end(encoded);
    });

    server.once('error', reject);
    server.listen(port, '127.0.0.1', () => {
      server.off('error', reject);
      resolve({
        server,
        getRequests: () => requests.slice(),
      });
    });
  });
}

async function ensureHomeLayoutSeedData(token) {
  const mockServer = await createHomeSmokeMockServer(homeSmokeMockPort);

  try {
    const createdChannel = await apiPost(token, '/api/v1/channel/create', {
      name: homeSmokeChannelName,
      type: 0,
      enabled: true,
      base_urls: [{ url: `http://127.0.0.1:${homeSmokeMockPort}`, delay: 0 }],
      keys: [{ enabled: true, channel_key: 'home-smoke-upstream-key', source_type: 'private/internal' }],
      model: homeSmokeGroupName,
    });

    await apiPost(token, '/api/v1/group/create', {
      name: homeSmokeGroupName,
      mode: 1,
      items: [{ channel_id: createdChannel.data.id, model_name: homeSmokeGroupName, priority: 1, weight: 1 }],
    });

    const createdAPIKey = await apiPost(token, '/api/v1/apikey/create', {
      name: homeSmokeAPIKeyName,
      enabled: true,
    });

    await apiGatewayPost(createdAPIKey.data.api_key, '/v1/chat/completions', {
      model: homeSmokeGroupName,
      messages: [{ role: 'user', content: 'browser smoke homepage seed' }],
    });

    const channelList = await apiGet(token, '/api/v1/channel/list');
    const groupList = await apiGet(token, '/api/v1/group/list');

    return {
      channelCount: Array.isArray(channelList.data) ? channelList.data.length : 0,
      groupCount: Array.isArray(groupList.data) ? groupList.data.length : 0,
    };
  } finally {
    await new Promise((resolve) => mockServer.close(() => resolve()));
  }
}

async function ensureAILearningSeedData(token) {
  await apiPost(token, '/api/v1/setting/set', {
    key: 'dynamic_routing_learning_enabled',
    value: 'true',
  });

  const mockServer = await createHomeSmokeMockServer(aiLearningSmokeMockPort);

  try {
    const createdChannel = await apiPost(token, '/api/v1/channel/create', {
      name: aiLearningSmokeChannelName,
      type: 0,
      enabled: true,
      base_urls: [{ url: `http://127.0.0.1:${aiLearningSmokeMockPort}`, delay: 0 }],
      keys: [{ enabled: true, channel_key: 'ai-learning-upstream-key', source_type: 'private/internal' }],
      model: aiLearningSmokeGroupName,
    });

    await apiPost(token, '/api/v1/group/create', {
      name: aiLearningSmokeGroupName,
      mode: 1,
      items: [{ channel_id: createdChannel.data.id, model_name: aiLearningSmokeGroupName, priority: 1, weight: 1 }],
    });

    const createdAPIKey = await apiPost(token, '/api/v1/apikey/create', {
      name: aiLearningSmokeAPIKeyName,
      enabled: true,
    });

    await apiGatewayPost(createdAPIKey.data.api_key, '/v1/chat/completions', {
      model: aiLearningSmokeGroupName,
      messages: [{ role: 'user', content: 'browser smoke ai learning seed' }],
    });

    await sleep(500);
    const learning = await apiGet(token, '/api/v1/dynamic-routing/learning');

    return {
      enabled: Boolean(learning.data?.enabled),
      stateCount: Array.isArray(learning.data?.states) ? learning.data.states.length : 0,
    };
  } finally {
    await new Promise((resolve) => mockServer.close(() => resolve()));
  }
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

  const apiKeyList = await apiGet(token, '/api/v1/apikey/list');
  const groupList = await apiGet(token, '/api/v1/group/list');

  return {
    apiKeyCount: Array.isArray(apiKeyList.data) ? apiKeyList.data.length : 0,
    groupCount: Array.isArray(groupList.data) ? groupList.data.length : 0,
  };
}

async function ensureChannelPageSeedData(token) {
  const createdChannel = await apiPost(token, '/api/v1/channel/create', {
    name: channelPageSmokeName,
    type: 0,
    enabled: true,
    key_management_mode: 'classified',
    key_routing_policy: 'fill_priority',
    base_urls: [{ url: 'https://example.com', delay: 23 }],
    keys: [
      {
        enabled: true,
        channel_key: 'sk-octopus-browser-channel-primary',
        source_type: 'paid/metered',
        remark: 'Primary browser key',
        allowed_models: `${channelPageSmokeModel},gpt-4o-mini`,
      },
      {
        enabled: false,
        channel_key: 'sk-octopus-browser-channel-fallback',
        source_type: 'public/free',
        remark: 'Fallback browser key',
        allowed_models: 'fallback-model',
      },
    ],
    model: channelPageSmokeModel,
    custom_model: 'gpt-4o-mini',
  });

  const created = createdChannel.data;
  const firstKey = Array.isArray(created.keys) ? created.keys[0] : null;
  if (!created?.id || !firstKey?.id) {
    throw new Error('channel-page seed channel did not return a persisted key');
  }

  await apiPost(token, '/api/v1/route-target/upsert', {
    channel_id: created.id,
    channel_key_id: firstKey.id,
    model_name: channelPageSmokeModel,
    billing_mode: 'per_request',
    probe_policy: 'concurrent',
    probe_interval_seconds: 120,
    probe_concurrency_limit: 2,
  });

  const channelList = await apiGet(token, '/api/v1/channel/list');
  const routeTargetList = await apiGet(token, `/api/v1/route-target/list?channel_id=${created.id}`);

  return {
    channelId: created.id,
    keyId: firstKey.id,
    channelCount: Array.isArray(channelList.data) ? channelList.data.length : 0,
    routeTargetCount: Array.isArray(routeTargetList.data) ? routeTargetList.data.length : 0,
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
    bootstrapUnavailable: didAllCdpCommandsFailWithBootstrapUnavailable(pageBootstrap),
    bootstrapTimedOut: bootstrapFailures.length > 0 && bootstrapFailures.every((entry) => entry.message?.includes('timed out')),
    runtimeTimedOut: Boolean(runtimeProbe && !runtimeProbe.ok && runtimeProbe.message?.includes('timed out')),
    emulationTimedOut: Boolean(emulationProbe && !emulationProbe.ok && emulationProbe.message?.includes('timed out')),
    navigationSucceeded: Boolean(navigateProbe?.ok),
    navigationUnavailable: Boolean(navigateProbe && !navigateProbe.ok && isCdpBootstrapUnavailableMessage(navigateProbe.message)),
    runtimeUnavailable: Boolean(runtimeProbe && !runtimeProbe.ok && isCdpBootstrapUnavailableMessage(runtimeProbe.message)),
    emulationUnavailable: Boolean(emulationProbe && !emulationProbe.ok && isCdpBootstrapUnavailableMessage(emulationProbe.message)),
    failureCount: bootstrapFailures.length + probeFailures.length,
  };
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
    fallbackFrom: page.fallbackFrom ?? null,
    targetId: page.targetId,
    sessionId: page.sessionId ?? null,
    pageBootstrap,
  };
  return error;
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
    fallbackFrom: page.fallbackFrom ?? null,
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
  return (summary.navigationSucceeded && summary.runtimeTimedOut && summary.emulationTimedOut)
    || (summary.bootstrapUnavailable && summary.navigationUnavailable && summary.runtimeUnavailable && summary.emulationUnavailable);
}

function shouldAbortAfterBootstrap(pageBootstrap) {
  return didAllCdpCommandsFailWithBootstrapUnavailable(pageBootstrap);
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
  fallbackPage.fallbackFrom = page.pageMode;
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

async function waitForChannelPageReady(connection, sessionId, timeoutMs = 45000) {
  const startedAt = Date.now();
  let lastBodyText = '';

  while (Date.now() - startedAt < timeoutMs) {
    try {
      const snapshot = await evaluateCdp(connection, sessionId, `(() => ({
        channelTrigger: !!document.querySelector('[data-testid="toolbar-create-trigger-channel"]'),
        bodyText: document.body?.innerText || '',
      }))()`);
      lastBodyText = snapshot.bodyText || '';

      if (snapshot.channelTrigger) {
        return snapshot;
      }
    } catch {
    }
    await sleep(1000);
  }

  throw new Error(`Timed out waiting for channel page on ${frontendBaseUrl}. Last body text snippet: ${lastBodyText.slice(0, 200)}`);
}

async function waitForModelPageReady(connection, sessionId, timeoutMs = 45000) {
  const startedAt = Date.now();
  let lastBodyText = '';

  while (Date.now() - startedAt < timeoutMs) {
    try {
      const snapshot = await evaluateCdp(connection, sessionId, `(() => ({
        modelTrigger: !!document.querySelector('[data-testid="toolbar-view-options-trigger-model"]'),
        modelPage: !!document.querySelector('[data-testid="model-page"]'),
        bodyText: document.body?.innerText || '',
      }))()`);
      lastBodyText = snapshot.bodyText || '';

      if (snapshot.modelTrigger && snapshot.modelPage) {
        return snapshot;
      }
    } catch {
    }
    await sleep(1000);
  }

  throw new Error(`Timed out waiting for model page on ${frontendBaseUrl}. Last body text snippet: ${lastBodyText.slice(0, 200)}`);
}

async function waitForHomePageReady(connection, sessionId, timeoutMs = 45000) {
  const startedAt = Date.now();
  let lastBodyText = '';

  while (Date.now() - startedAt < timeoutMs) {
    try {
      const snapshot = await evaluateCdp(connection, sessionId, `(() => ({
        homePage: !!document.querySelector('[data-testid="home-page"]'),
        totalSection: !!document.querySelector('[data-testid="home-total-section"]'),
        breakdownSection: !!document.querySelector('[data-testid="home-breakdown-section"]'),
        chartSection: !!document.querySelector('[data-testid="home-stats-chart-section"]'),
        rankSection: !!document.querySelector('[data-testid="home-rank-section"]'),
        activitySection: !!document.querySelector('[data-testid="home-activity-section"]'),
        bodyText: document.body?.innerText || '',
      }))()`);
      lastBodyText = snapshot.bodyText || '';

      if (snapshot.homePage && snapshot.totalSection && snapshot.breakdownSection && snapshot.chartSection && snapshot.rankSection && snapshot.activitySection) {
        return snapshot;
      }
    } catch {
    }
    await sleep(1000);
  }

  throw new Error(`Timed out waiting for home page on ${frontendBaseUrl}. Last body text snippet: ${lastBodyText.slice(0, 200)}`);
}

async function waitForAILearningPageReady(connection, sessionId, timeoutMs = 45000) {
  const startedAt = Date.now();
  let lastBodyText = '';

  while (Date.now() - startedAt < timeoutMs) {
    try {
      const snapshot = await evaluateCdp(connection, sessionId, `(() => ({
        page: !!document.querySelector('[data-testid="ai-automation-page"]'),
        section: !!document.querySelector('[data-testid="ai-automation-learning-section"]'),
        stage: !!document.querySelector('[data-testid="ai-automation-learning-stage-card"]'),
        controls: !!document.querySelector('[data-testid="ai-automation-learning-controls"]'),
        presetCard: !!document.querySelector('[data-testid="ai-automation-learning-preset-card"]'),
        presetButtons: document.querySelectorAll('[data-testid^="ai-automation-learning-preset-"]').length,
        switchCard: !!document.querySelector('[data-testid="ai-automation-learning-switch-card"]'),
        switchButton: !!document.querySelector('[data-testid="ai-automation-learning-switch"]'),
        summary: !!document.querySelector('[data-testid="ai-automation-learning-state-summary"]'),
        secondarySummary: !!document.querySelector('[data-testid="ai-automation-learning-secondary-summary"]'),
        states: !!document.querySelector('[data-testid="ai-automation-learning-states"]'),
        bodyText: document.body?.innerText || '',
      }))()`);
      lastBodyText = snapshot.bodyText || '';

      if (
        snapshot.page &&
        snapshot.section &&
        snapshot.stage &&
        snapshot.controls &&
        snapshot.presetCard &&
        snapshot.presetButtons >= 3 &&
        snapshot.switchCard &&
        snapshot.switchButton &&
        snapshot.summary &&
        snapshot.secondarySummary &&
        snapshot.states
      ) {
        return snapshot;
      }
    } catch {
    }
    await sleep(1000);
  }

  throw new Error(`Timed out waiting for AI automation learning page on ${frontendBaseUrl}. Last body text snippet: ${lastBodyText.slice(0, 200)}`);
}

async function waitForDocModalReady(connection, sessionId, timeoutMs = 20000) {
  const startedAt = Date.now();
  let lastBodyText = '';

  while (Date.now() - startedAt < timeoutMs) {
    try {
      const snapshot = await evaluateCdp(connection, sessionId, `(() => ({
        docTrigger: !!document.querySelector('[data-testid="navbar-doc-trigger"]'),
        modal: !!document.querySelector('[data-testid="doc-modal"]'),
        ccswitchPanel: !!document.querySelector('[data-testid="ccswitch-panel"]'),
        bodyText: document.body?.innerText || '',
      }))()`);
      lastBodyText = snapshot.bodyText || '';

      if (snapshot.docTrigger && snapshot.modal && snapshot.ccswitchPanel) {
        return snapshot;
      }
    } catch {
    }
    await sleep(250);
  }

  throw new Error(`Timed out waiting for DocModal on ${frontendBaseUrl}. Last body text snippet: ${lastBodyText.slice(0, 200)}`);
}

async function waitForHomeStatsLoaded(connection, sessionId, timeoutMs = 20000) {
  const startedAt = Date.now();

  while (Date.now() - startedAt < timeoutMs) {
    const snapshot = await evaluateCdp(connection, sessionId, `(() => ({
      totalSummaryCards: document.querySelectorAll('[data-testid^="home-total-summary-card-"]').length,
      breakdownLists: document.querySelectorAll('[data-testid="home-breakdown-lists"] > div').length,
      rankItems: document.querySelectorAll('[data-testid="home-rank-section"] [class*="rounded-2xl"]').length,
      activityCells: document.querySelectorAll('[data-testid="home-activity-grid"] button').length,
      chartSvg: !!document.querySelector('[data-testid="home-stats-chart"] svg'),
    }))()`);

    if (snapshot.totalSummaryCards >= 3 && snapshot.breakdownLists >= 3 && snapshot.chartSvg && snapshot.activityCells > 0) {
      return snapshot;
    }

    await sleep(500);
  }

  throw new Error('Timed out waiting for homepage stats widgets to populate');
}

async function waitForAILearningDataLoaded(connection, sessionId, timeoutMs = 20000) {
  const startedAt = Date.now();

  while (Date.now() - startedAt < timeoutMs) {
    const snapshot = await evaluateCdp(connection, sessionId, `(() => ({
      stateCards: document.querySelectorAll('[data-testid="ai-automation-learning-states"] > [data-testid^="ai-automation-learning-state-"]').length,
      emptyVisible: !!document.querySelector('[data-testid="ai-automation-learning-empty"]'),
      switchState: document.querySelector('[data-testid="ai-automation-learning-switch"]')?.getAttribute('aria-checked') || '',
      resetDisabled: document.querySelector('[data-testid="ai-automation-learning-reset"]')?.hasAttribute('disabled') ?? true,
      firstStateTitle: document.querySelector('[data-testid="ai-automation-learning-states"] > [data-testid^="ai-automation-learning-state-"] .font-medium')?.textContent?.trim() || '',
    }))()`);

    if (snapshot.stateCards >= 1 && snapshot.emptyVisible === false && snapshot.switchState === 'true' && snapshot.resetDisabled === false) {
      return snapshot;
    }

    await sleep(250);
  }

  throw new Error('Timed out waiting for AI automation learning state cards to populate');
}

async function waitForAILearningResetState(connection, sessionId, timeoutMs = 10000) {
  const startedAt = Date.now();

  while (Date.now() - startedAt < timeoutMs) {
    const snapshot = await evaluateCdp(connection, sessionId, `(() => ({
      stateCards: document.querySelectorAll('[data-testid="ai-automation-learning-states"] > [data-testid^="ai-automation-learning-state-"]').length,
      emptyVisible: !!document.querySelector('[data-testid="ai-automation-learning-empty"]'),
      resetDisabled: document.querySelector('[data-testid="ai-automation-learning-reset"]')?.hasAttribute('disabled') ?? false,
    }))()`);

    if (snapshot.stateCards === 0 && snapshot.emptyVisible && snapshot.resetDisabled) {
      return snapshot;
    }

    await sleep(250);
  }

  throw new Error('Timed out waiting for AI automation learning reset state');
}

async function waitForChannelCardsLoaded(connection, sessionId, timeoutMs = 20000) {
  const startedAt = Date.now();

  while (Date.now() - startedAt < timeoutMs) {
    const snapshot = await evaluateCdp(connection, sessionId, `(() => ({
      page: !!document.querySelector('[data-testid="channel-page"]'),
      cards: document.querySelectorAll('[data-testid^="channel-card-"]').length,
      toolbarTrigger: !!document.querySelector('[data-testid="toolbar-view-options-trigger-channel"]'),
      createTrigger: !!document.querySelector('[data-testid="toolbar-create-trigger-channel"]'),
    }))()`);

    if (snapshot.page && snapshot.cards > 0 && snapshot.toolbarTrigger && snapshot.createTrigger) {
      return snapshot;
    }

    await sleep(500);
  }

  throw new Error('Timed out waiting for channel cards to populate');
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

async function waitForEvaluation(connection, sessionId, expression, timeoutMs = 6000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    const value = await evaluateCdp(connection, sessionId, expression);
    if (value) {
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

async function setInputValueViaCdp(connection, sessionId, selector, value) {
  return evaluateCdp(connection, sessionId, `(() => {
    const input = document.querySelector(${JSON.stringify(selector)});
    if (!(input instanceof HTMLInputElement)) {
      return { found: false, value: '' };
    }
    input.focus();
    input.value = '';
    input.dispatchEvent(new Event('input', { bubbles: true }));
    input.value = ${JSON.stringify(value)};
    input.dispatchEvent(new Event('input', { bubbles: true }));
    input.dispatchEvent(new Event('change', { bubbles: true }));
    return { found: true, value: input.value };
  })()`);
}

async function hoverSelectorViaCdp(connection, sessionId, selector) {
  const target = await evaluateCdp(connection, sessionId, `(() => {
    const node = document.querySelector(${JSON.stringify(selector)});
    if (!(node instanceof HTMLElement)) {
      return { found: false, x: 0, y: 0 };
    }
    node.scrollIntoView({ block: 'center', inline: 'center' });
    const rect = node.getBoundingClientRect();
    return {
      found: rect.width > 0 && rect.height > 0,
      x: rect.left + Math.max(Math.min(rect.width / 2, rect.width - 1), 1),
      y: rect.top + Math.max(Math.min(rect.height / 2, rect.height - 1), 1),
    };
  })()`);

  if (!target.found) {
    return false;
  }

  await connection.send('Input.dispatchMouseEvent', {
    type: 'mouseMoved',
    x: target.x,
    y: target.y,
    buttons: 0,
  }, sessionId);
  await sleep(200);
  return true;
}

async function resetHelpHintInteractionViaCdp(connection, sessionId) {
  await evaluateCdp(connection, sessionId, `(() => {
    const active = document.activeElement;
    if (active instanceof HTMLElement) {
      active.blur();
    }
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    return true;
  })()`);
  await connection.send('Input.dispatchMouseEvent', {
    type: 'mouseMoved',
    x: 1,
    y: 1,
    buttons: 0,
  }, sessionId);
  await sleep(200);
}

async function waitForDialogOpen(connection, sessionId, timeoutMs = 20000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    const snapshot = await evaluateCdp(connection, sessionId, `(() => ({
      open: !!document.querySelector('[data-testid="channel-create-dialog"]'),
      flow: !!document.querySelector('[data-testid="new-channel-flow-card"]'),
      form: !!document.querySelector('[data-testid="new-channel-form"]'),
    }))()`);
    if (snapshot.open && snapshot.flow && snapshot.form) {
      return snapshot;
    }
    await sleep(400);
  }
  throw new Error('Timed out waiting for channel create dialog to open');
}

async function waitForKeyItemExpanded(connection, sessionId, targetSelector, timeoutMs = 4000) {
  const startedAt = Date.now();

  while (Date.now() - startedAt < timeoutMs) {
    const snapshot = await evaluateCdp(connection, sessionId, `(() => {
      const selector = ${JSON.stringify(targetSelector)};
      const item = document.querySelector(selector);
      const primary = document.querySelector(selector.replace('key-item', 'key-primary'));
      return {
        found: !!item,
        state: item?.getAttribute('data-state') || '',
        primaryVisible: !!primary,
        primaryText: primary?.textContent?.trim() || '',
        valueInput: !!primary?.querySelector('input[type="text"]'),
        statusBadges: Array.from(primary?.querySelectorAll('[data-slot], .rounded-2xl, .rounded-full') || []).map((node) => (node.textContent || '').trim()).filter(Boolean),
      };
    })()`);
    if (snapshot.state === 'open' && snapshot.primaryVisible) {
      return snapshot;
    }
    await sleep(80);
  }

  return evaluateCdp(connection, sessionId, `(() => {
    const selector = ${JSON.stringify(targetSelector)};
    const item = document.querySelector(selector);
    const primary = document.querySelector(selector.replace('key-item', 'key-primary'));
    return {
      found: !!item,
      state: item?.getAttribute('data-state') || '',
      primaryVisible: !!primary,
      primaryText: primary?.textContent?.trim() || '',
      valueInput: !!primary?.querySelector('input[type="text"]'),
      statusBadges: Array.from(primary?.querySelectorAll('[data-slot], .rounded-2xl, .rounded-full') || []).map((node) => (node.textContent || '').trim()).filter(Boolean),
    };
  })()`);
}

async function focusHelpHintAndReadTooltip(connection, sessionId) {
  const focusResult = await evaluateCdp(connection, sessionId, `(() => {
    const button = document.querySelector('[data-testid="channel-create-dialog"] ${helpButtonSelector}');
    if (!button) {
      return { found: false, hintId: '' };
    }
    button.focus();
    return {
      found: true,
      hintId: button.getAttribute('data-help-hint-id') || '',
    };
  })()`);

  if (!focusResult.found) {
    return { found: false, hintId: '', tooltipText: '', tooltipCount: 0 };
  }

  await sleep(200);

  return evaluateCdp(connection, sessionId, `(() => {
    const button = document.querySelector('[data-testid="channel-create-dialog"] ${helpButtonSelector}');
    const hintId = button?.getAttribute('data-help-hint-id') || '';
    const tooltips = Array.from(document.querySelectorAll('[data-slot="help-hint-content"], [role="tooltip"]')).map((node) => ({
      id: node.id || '',
      hintId: node.getAttribute('data-help-hint-id') || '',
      text: (node.textContent || '').trim(),
    }));
    const current = tooltips.find((item) => hintId && (item.id === hintId || item.hintId === hintId));
    return {
      found: !!button,
      hintId,
      tooltipText: current?.text || '',
      tooltipCount: tooltips.filter((item) => item.text).length,
    };
  })()`);
}

async function focusScopedHelpHintAndReadTooltip(connection, sessionId, scopeSelector) {
  const focusResult = await evaluateCdp(connection, sessionId, `(() => {
    const root = document.querySelector(${JSON.stringify(scopeSelector)});
    const button = root?.querySelector(${JSON.stringify(helpButtonSelector)});
    if (!button) {
      return { found: false, hintId: '' };
    }
    button.focus();
    return {
      found: true,
      hintId: button.getAttribute('data-help-hint-id') || '',
    };
  })()`);

  if (!focusResult.found) {
    return { found: false, hintId: '', tooltipText: '', tooltipCount: 0 };
  }

  await sleep(200);

  return evaluateCdp(connection, sessionId, `(() => {
    const root = document.querySelector(${JSON.stringify(scopeSelector)});
    const button = root?.querySelector(${JSON.stringify(helpButtonSelector)});
    const hintId = button?.getAttribute('data-help-hint-id') || '';
    const tooltips = Array.from(document.querySelectorAll('[data-slot="help-hint-content"], [role="tooltip"]')).map((node) => ({
      id: node.id || '',
      hintId: node.getAttribute('data-help-hint-id') || '',
      text: (node.textContent || '').trim(),
    }));
    const current = tooltips.find((item) => hintId && (item.id === hintId || item.hintId === hintId));
    return {
      found: !!button,
      hintId,
      tooltipText: current?.text || '',
      tooltipCount: tooltips.filter((item) => item.text).length,
    };
  })()`);
}

async function hoverScopedHelpHintAndReadTooltip(connection, sessionId, scopeSelector) {
  const selector = `${scopeSelector} ${helpButtonSelector}`;
  const hovered = await hoverSelectorViaCdp(connection, sessionId, selector);

  if (!hovered) {
    return { found: false, hintId: '', tooltipText: '', tooltipCount: 0 };
  }

  await sleep(200);

  return evaluateCdp(connection, sessionId, `(() => {
    const root = document.querySelector(${JSON.stringify(scopeSelector)});
    const button = root?.querySelector(${JSON.stringify(helpButtonSelector)});
    const hintId = button?.getAttribute('data-help-hint-id') || '';
    const tooltips = Array.from(document.querySelectorAll('[data-slot="help-hint-content"], [role="tooltip"]')).map((node) => ({
      id: node.id || '',
      hintId: node.getAttribute('data-help-hint-id') || '',
      text: (node.textContent || '').trim(),
    }));
    const current = tooltips.find((item) => hintId && (item.id === hintId || item.hintId === hintId));
    return {
      found: !!button,
      hintId,
      tooltipText: current?.text || '',
      tooltipCount: tooltips.filter((item) => item.text).length,
    };
  })()`);
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
  if (smokeScenario === 'model-layout') {
    await ensureModelLayoutSeedData(token);
  } else if (smokeScenario === 'channel-page') {
    await ensureChannelPageSeedData(token);
  } else if (smokeScenario === 'home-layout') {
    await ensureHomeLayoutSeedData(token);
  } else if (smokeScenario === 'ai-learning') {
    await ensureAILearningSeedData(token);
  } else if (smokeScenario === 'ccswitch') {
    await ensureCCSwitchSeedData(token);
  }
  trace('smoke-auth:token-resolved');
  const activeNavItem = smokeScenario === 'model-layout'
    ? 'model'
    : smokeScenario === 'home-layout'
      ? 'home'
      : smokeScenario === 'ai-learning'
        ? 'ai'
        : 'channel';
  const authStorage = JSON.stringify({
    state: {
      token,
      expireAt: new Date(Date.now() + 3600_000).toISOString(),
      isAPIKeyAuth: false,
    },
    version: 0,
  });
  const navStorage = JSON.stringify({
    state: {
      activeItem: activeNavItem,
      prevItem: 'home',
      direction: 1,
    },
    version: 0,
  });
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

    if (smokeScenario === 'model-layout') {
      const modelPage = await waitForModelPageReady(page.pageConnection, page.sessionId);
      assert.equal(modelPage.modelTrigger, true, 'model layout trigger should exist on model page');

      const cardsReady = await waitForSelector(page.pageConnection, page.sessionId, '[data-testid="model-card-0"]', 12000);
      assert.equal(cardsReady, true, 'model page should render at least one model card');

      const desktop = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        pageLayout: document.querySelector('[data-testid="model-page"]')?.getAttribute('data-layout') || '',
        cardCount: document.querySelectorAll('[data-slot="model-card"]').length,
        firstCardLayout: document.querySelector('[data-testid="model-card-0"]')?.getAttribute('data-layout') || '',
        hasGridMeta: !!document.querySelector('[data-testid="model-card-0"] [data-slot="model-card-meta"]'),
        hasCompactMeta: !!document.querySelector('[data-testid="model-card-0"] [data-slot="model-card-meta-compact"]'),
        firstCardName: document.querySelector('[data-testid="model-card-0"]')?.getAttribute('data-model-name') || '',
      }))()`);
      assert.equal(desktop.pageLayout, 'grid', 'model page should default to normal layout');
      assert.equal(desktop.firstCardLayout, 'grid', 'first model card should default to normal layout');
      assert.equal(desktop.hasGridMeta, true, 'normal layout should expose model meta pills');
      assert.equal(desktop.hasCompactMeta, false, 'normal layout should not expose compact meta block');
      assert.ok(desktop.cardCount >= 2, 'seeded models should render at least two cards');
      assert.ok(desktop.firstCardName.startsWith(modelLayoutSeedPrefix), 'first model card should match the seeded model prefix');

      const viewOptionsOpened = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="toolbar-view-options-trigger-model"]');
      assert.equal(viewOptionsOpened, true, 'model toolbar view options trigger should be clickable');
      const viewOptionsVisible = await waitForSelector(page.pageConnection, page.sessionId, '[data-testid="toolbar-view-options-content-model"]');
      assert.equal(viewOptionsVisible, true, 'model toolbar view options popover should open');

      const compactClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="toolbar-layout-list-model"]');
      assert.equal(compactClicked, true, 'compact layout option should be clickable');
      const compactApplied = await waitForSelector(page.pageConnection, page.sessionId, '[data-testid="model-page"][data-layout="list"]', 6000);
      assert.equal(compactApplied, true, 'compact layout should apply to model page');

      const compact = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        pageLayout: document.querySelector('[data-testid="model-page"]')?.getAttribute('data-layout') || '',
        firstCardLayout: document.querySelector('[data-testid="model-card-0"]')?.getAttribute('data-layout') || '',
        hasGridMeta: !!document.querySelector('[data-testid="model-card-0"] [data-slot="model-card-meta"]'),
        hasCompactMeta: !!document.querySelector('[data-testid="model-card-0"] [data-slot="model-card-meta-compact"]'),
      }))()`);
      assert.equal(compact.pageLayout, 'list', 'model page should switch to compact layout');
      assert.equal(compact.firstCardLayout, 'list', 'first model card should switch to compact layout');
      assert.equal(compact.hasGridMeta, false, 'compact layout should hide normal meta block');
      assert.equal(compact.hasCompactMeta, true, 'compact layout should expose compact meta block');

      const normalClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="toolbar-layout-grid-model"]');
      assert.equal(normalClicked, true, 'normal layout option should be clickable');
      const normalApplied = await waitForSelector(page.pageConnection, page.sessionId, '[data-testid="model-page"][data-layout="grid"]', 6000);
      assert.equal(normalApplied, true, 'normal layout should re-apply to model page');

      await setCdpViewport(page.pageConnection, page.sessionId, { width: 375, height: 1200, mobile: true });
      await sleep(300);
      const mobile = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        width: window.innerWidth,
        bodyWidth: document.body.scrollWidth,
        viewport: document.documentElement.clientWidth,
        pageWidth: document.querySelector('[data-testid="model-page"]')?.getBoundingClientRect().width ?? 0,
        firstCardWidth: document.querySelector('[data-testid="model-card-0"]')?.getBoundingClientRect().width ?? 0,
        triggerWidth: document.querySelector('[data-testid="toolbar-view-options-trigger-model"]')?.getBoundingClientRect().width ?? 0,
      }))()`);
      assert.equal(mobile.width, 375, 'mobile viewport should be set to 375px');
      assert.ok(mobile.pageWidth > 0, 'model page should remain visible on mobile');
      assert.ok(mobile.firstCardWidth > 0, 'model card should remain visible on mobile');
      assert.ok(mobile.triggerWidth > 0, 'view options trigger should remain visible on mobile');
      assert.ok(mobile.bodyWidth <= mobile.viewport + 24, 'model page should not introduce large horizontal overflow');

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
        cardCount: desktop.cardCount,
        mobileWidth: mobile.width,
        result: 'model-layout-browser-smoke-cdp passed',
      }, null, 2));
    } else if (smokeScenario === 'channel-page') {
      const channelPage = await waitForChannelPageReady(page.pageConnection, page.sessionId);
      assert.equal(channelPage.channelTrigger, true, 'channel page trigger should exist');

      const loaded = await waitForChannelCardsLoaded(page.pageConnection, page.sessionId);
      assert.ok(loaded.cards >= 1, 'channel page should render at least one channel card');

      const desktop = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        width: window.innerWidth,
        pageLayout: document.querySelector('[data-testid="channel-page"]')?.getAttribute('data-layout') || '',
        cardCount: document.querySelectorAll('[data-testid^="channel-card-"][data-channel-name]').length,
        firstCardName: document.querySelector('[data-testid^="channel-card-"][data-channel-name]')?.getAttribute('data-channel-name') || '',
        firstCardBadges: document.querySelectorAll('[data-testid^="channel-card-badges-"] .inline-flex, [data-testid^="channel-card-badges-"] [class*="badge"]').length,
        firstCardMetrics: document.querySelectorAll('[data-testid^="channel-card-metrics-"] > div').length,
      }))()`);
      assert.equal(desktop.pageLayout, 'grid', 'channel page should default to grid layout');
      assert.ok(desktop.cardCount >= 1, 'channel page should expose channel cards');
      assert.ok(desktop.firstCardName.length > 0, 'first channel card should expose its name');
      assert.ok(desktop.firstCardBadges >= 2, 'channel card should expose summary badges');
      assert.ok(desktop.firstCardMetrics >= 2, 'channel card should expose summary metrics');

      const viewOptionsOpened = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="toolbar-view-options-trigger-channel"]');
      assert.equal(viewOptionsOpened, true, 'channel toolbar view options trigger should be clickable');
      const viewOptionsVisible = await waitForSelector(page.pageConnection, page.sessionId, '[data-testid="toolbar-view-options-content-channel"]', 6000);
      assert.equal(viewOptionsVisible, true, 'channel toolbar view options popover should open');

      const providerOpenAI = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="toolbar-channel-provider-openai"]');
      assert.equal(providerOpenAI, true, 'channel provider filter option should be clickable');

      const modelKeywordTyped = await setInputValueViaCdp(page.pageConnection, page.sessionId, '[data-testid="toolbar-channel-model-keyword"]', channelPageSmokeModel);
      assert.equal(modelKeywordTyped.found, true, 'channel model filter input should exist');
      assert.equal(modelKeywordTyped.value, channelPageSmokeModel, 'channel model filter input should accept keyword');

      const keyKeywordTyped = await setInputValueViaCdp(page.pageConnection, page.sessionId, '[data-testid="toolbar-channel-key-keyword"]', 'Primary browser key');
      assert.equal(keyKeywordTyped.found, true, 'channel key filter input should exist');
      assert.equal(keyKeywordTyped.value, 'Primary browser key', 'channel key filter input should accept keyword');

      const filtered = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        cardCount: document.querySelectorAll('[data-testid^="channel-card-"][data-channel-name]').length,
        names: Array.from(document.querySelectorAll('[data-testid^="channel-card-"][data-channel-name]')).map(node => node.getAttribute('data-channel-name') || ''),
      }))()`);
      assert.ok(filtered.cardCount >= 1, 'channel page should keep at least one matching card after filter');
      assert.ok(filtered.names.some((name) => name === channelPageSmokeName), 'channel page filter should retain the seeded channel');

      const detailOpened = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid^="channel-card-trigger-"]');
      assert.equal(detailOpened, true, 'channel card trigger should open detail dialog');
      const detailVisible = await waitForSelector(page.pageConnection, page.sessionId, '[data-testid^="channel-detail-dialog-"]', 6000);
      assert.equal(detailVisible, true, 'channel detail dialog should open');

      const detail = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        detailView: !!document.querySelector('[data-testid^="channel-detail-view-"]'),
        routingSection: !!document.querySelector('[data-testid^="channel-routing-section-"]'),
        routeTargetSummary: !!document.querySelector('[data-testid^="channel-route-target-summary-"]'),
        keyFilter: !!document.querySelector('[data-testid^="channel-key-filter-"]'),
        keyAccordion: !!document.querySelector('[data-testid^="channel-key-accordion-"]'),
        keyItems: document.querySelectorAll('[data-testid^="channel-key-item-"]').length,
      }))()`);
      assert.equal(detail.detailView, true, 'channel detail view should render inside dialog');
      assert.equal(detail.routingSection, true, 'channel detail should show routing section');
      assert.equal(detail.routeTargetSummary, true, 'channel detail should show route-target summary');
      assert.equal(detail.keyFilter, true, 'channel detail should expose key filter input');
      assert.equal(detail.keyAccordion, true, 'channel detail should expose key accordion');
      assert.ok(detail.keyItems >= 1, 'channel detail should expose at least one key item');

      const detailKeyKeywordTyped = await setInputValueViaCdp(page.pageConnection, page.sessionId, '[data-testid^="channel-key-filter-"]', 'Primary browser key');
      assert.equal(detailKeyKeywordTyped.found, true, 'channel detail key filter input should accept keyword');

      const filteredDetail = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        keyItems: document.querySelectorAll('[data-testid^="channel-key-item-"]').length,
        noMatch: !!document.querySelector('[data-testid^="channel-key-accordion-"] + div'),
      }))()`);
      assert.equal(filteredDetail.keyItems, 1, 'channel detail key filter should reduce the seeded keys to one match');
      assert.equal(filteredDetail.noMatch, false, 'channel detail key filter should keep a matching result visible');

      const firstKeyExpanded = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid^="channel-key-trigger-"]');
      assert.equal(firstKeyExpanded, true, 'channel detail key trigger should expand');
      const keyModelsVisible = await waitForSelector(page.pageConnection, page.sessionId, '[data-testid^="channel-key-models-"]', 6000);
      assert.equal(keyModelsVisible, true, 'channel detail should reveal allowed models block after expanding a key');

      await setCdpViewport(page.pageConnection, page.sessionId, { width: 375, height: 1200, mobile: true });
      await sleep(300);
      const mobile = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        width: window.innerWidth,
        bodyWidth: document.body.scrollWidth,
        viewport: document.documentElement.clientWidth,
        pageWidth: document.querySelector('[data-testid="channel-page"]')?.getBoundingClientRect().width ?? 0,
        dialogWidth: document.querySelector('[data-testid^="channel-detail-dialog-"]')?.getBoundingClientRect().width ?? 0,
        keyTriggerWidth: document.querySelector('[data-testid^="channel-key-trigger-"]')?.getBoundingClientRect().width ?? 0,
      }))()`);
      assert.equal(mobile.width, 375, 'channel page mobile viewport should be set to 375px');
      assert.ok(mobile.pageWidth > 0, 'channel page should remain visible on mobile');
      assert.ok(mobile.dialogWidth > 0, 'channel detail dialog should remain visible on mobile');
      assert.ok(mobile.keyTriggerWidth > 0, 'channel detail key trigger should remain visible on mobile');
      assert.ok(mobile.bodyWidth <= mobile.viewport + 24, 'channel page should not introduce large horizontal overflow');

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
        cardCount: desktop.cardCount,
        filteredCardCount: filtered.cardCount,
        detailKeyItems: detail.keyItems,
        mobileWidth: mobile.width,
        result: 'channel-page-browser-smoke-cdp passed',
      }, null, 2));
    } else if (smokeScenario === 'home-layout') {
      const homePage = await waitForHomePageReady(page.pageConnection, page.sessionId);
      assert.equal(homePage.homePage, true, 'home page root should exist');

      const loaded = await waitForHomeStatsLoaded(page.pageConnection, page.sessionId);
      assert.ok(loaded.totalSummaryCards >= 3, 'homepage should render three total summary cards');
      assert.ok(loaded.breakdownLists >= 3, 'homepage should render three token breakdown lists');
      assert.equal(loaded.chartSvg, true, 'homepage chart should render an svg');
      assert.ok(loaded.activityCells > 0, 'homepage activity grid should render day cells');

      const desktop = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        width: window.innerWidth,
        homePageWidth: document.querySelector('[data-testid="home-page"]')?.getBoundingClientRect().width ?? 0,
        totalWidth: document.querySelector('[data-testid="home-total-section"]')?.getBoundingClientRect().width ?? 0,
        mainGridWidth: document.querySelector('[data-testid="home-main-grid"]')?.getBoundingClientRect().width ?? 0,
        breakdownWidth: document.querySelector('[data-testid="home-breakdown-section"]')?.getBoundingClientRect().width ?? 0,
        chartWidth: document.querySelector('[data-testid="home-stats-chart-section"]')?.getBoundingClientRect().width ?? 0,
        rankWidth: document.querySelector('[data-testid="home-rank-section"]')?.getBoundingClientRect().width ?? 0,
        activityWidth: document.querySelector('[data-testid="home-activity-section"]')?.getBoundingClientRect().width ?? 0,
        runtimeVisible: !!document.querySelector('[data-testid="home-runtime-panel"]'),
        breakdownListCount: document.querySelectorAll('[data-testid="home-breakdown-lists"] > div').length,
      }))()`);
      assert.ok(desktop.homePageWidth > 0, 'homepage should be visible on desktop');
      assert.ok(desktop.totalWidth > 0 && desktop.mainGridWidth > 0, 'homepage main sections should be visible on desktop');
      assert.ok(desktop.breakdownWidth > 0 && desktop.chartWidth > 0 && desktop.rankWidth > 0 && desktop.activityWidth > 0, 'homepage cards should remain visible on desktop');
      assert.equal(desktop.runtimeVisible, false, 'runtime details should stay collapsed by default');
      assert.equal(desktop.breakdownListCount, 3, 'homepage should show provider/channel/model lists');

      const runtimeToggleClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="home-runtime-toggle"]');
      assert.equal(runtimeToggleClicked, true, 'homepage runtime toggle should be clickable');
      const runtimeVisible = await waitForSelector(page.pageConnection, page.sessionId, '[data-testid="home-runtime-panel"]', 6000);
      assert.equal(runtimeVisible, true, 'homepage runtime panel should expand');

      const runtime = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        runtimePanel: !!document.querySelector('[data-testid="home-runtime-panel"]'),
        priceCard: !!document.querySelector('[data-testid="home-runtime-price-card"]'),
        circuitCard: !!document.querySelector('[data-testid="home-runtime-circuit-card"]'),
        probeCard: !!document.querySelector('[data-testid="home-runtime-probe-card"]'),
      }))()`);
      assert.equal(runtime.runtimePanel, true, 'runtime panel should stay visible after expand');
      assert.equal(runtime.priceCard, true, 'runtime price card should render');
      assert.equal(runtime.circuitCard, true, 'runtime circuit card should render');
      assert.equal(runtime.probeCard, true, 'runtime probe card should render');

      await setCdpViewport(page.pageConnection, page.sessionId, { width: 375, height: 1200, mobile: true });
      await sleep(300);
      const mobile = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        width: window.innerWidth,
        bodyWidth: document.body.scrollWidth,
        viewport: document.documentElement.clientWidth,
        totalWidth: document.querySelector('[data-testid="home-total-section"]')?.getBoundingClientRect().width ?? 0,
        breakdownWidth: document.querySelector('[data-testid="home-breakdown-section"]')?.getBoundingClientRect().width ?? 0,
        chartWidth: document.querySelector('[data-testid="home-stats-chart-section"]')?.getBoundingClientRect().width ?? 0,
        activityWidth: document.querySelector('[data-testid="home-activity-section"]')?.getBoundingClientRect().width ?? 0,
      }))()`);
      assert.equal(mobile.width, 375, 'mobile viewport should be set to 375px');
      assert.ok(mobile.totalWidth > 0 && mobile.breakdownWidth > 0 && mobile.chartWidth > 0 && mobile.activityWidth > 0, 'homepage sections should remain visible on mobile');
      assert.ok(mobile.bodyWidth <= mobile.viewport + 24, 'homepage should not introduce large horizontal overflow');

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
        breakdownListCount: desktop.breakdownListCount,
        activityCells: loaded.activityCells,
        mobileWidth: mobile.width,
        result: 'home-layout-browser-smoke-cdp passed',
      }, null, 2));
    } else if (smokeScenario === 'ai-learning') {
      const learningPage = await waitForAILearningPageReady(page.pageConnection, page.sessionId);
      assert.equal(learningPage.page, true, 'AI automation page root should exist');
      assert.equal(learningPage.section, true, 'AI learning section should exist');
      assert.equal(learningPage.presetButtons, 3, 'AI learning section should expose three preset buttons');

      const loaded = await waitForAILearningDataLoaded(page.pageConnection, page.sessionId);
      assert.ok(loaded.stateCards >= 1, 'AI learning should render at least one state card after seeding');
      assert.ok(loaded.firstStateTitle.length > 0, 'AI learning state cards should expose a model title');

      const desktop = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        width: window.innerWidth,
        pageWidth: document.querySelector('[data-testid="ai-automation-page"]')?.getBoundingClientRect().width ?? 0,
        sectionWidth: document.querySelector('[data-testid="ai-automation-learning-section"]')?.getBoundingClientRect().width ?? 0,
        summaryWidth: document.querySelector('[data-testid="ai-automation-learning-state-summary"]')?.getBoundingClientRect().width ?? 0,
        secondarySummaryWidth: document.querySelector('[data-testid="ai-automation-learning-secondary-summary"]')?.getBoundingClientRect().width ?? 0,
        presetButtons: document.querySelectorAll('[data-testid^="ai-automation-learning-preset-"]').length,
        switchState: document.querySelector('[data-testid="ai-automation-learning-switch"]')?.getAttribute('aria-checked') || '',
        stateCards: document.querySelectorAll('[data-testid="ai-automation-learning-states"] > [data-testid^="ai-automation-learning-state-"]').length,
        resetDisabled: document.querySelector('[data-testid="ai-automation-learning-reset"]')?.hasAttribute('disabled') ?? true,
      }))()`);
      assert.ok(desktop.pageWidth > 0 && desktop.sectionWidth > 0, 'AI learning shell should stay visible on desktop');
      assert.ok(desktop.summaryWidth > 0 && desktop.secondarySummaryWidth > 0, 'AI learning summaries should stay visible on desktop');
      assert.equal(desktop.presetButtons, 3, 'AI learning should keep three preset buttons on desktop');
      assert.equal(desktop.switchState, 'true', 'AI learning switch should default to enabled after seeding');
      assert.ok(desktop.stateCards >= 1, 'AI learning should expose state cards on desktop');
      assert.equal(desktop.resetDisabled, false, 'AI learning reset should be enabled when samples exist');

      const aggressiveClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="ai-automation-learning-preset-aggressive"]');
      assert.equal(aggressiveClicked, true, 'AI learning aggressive preset should be clickable');
      const aggressiveSelected = await waitForEvaluation(page.pageConnection, page.sessionId, `document.querySelector('[data-testid="ai-automation-learning-preset-aggressive"]')?.getAttribute('aria-pressed') === 'true'`, 4000);
      assert.equal(aggressiveSelected, true, 'AI learning aggressive preset should become selected after click');

      const resetClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="ai-automation-learning-reset"]');
      assert.equal(resetClicked, true, 'AI learning reset button should be clickable when samples exist');
      const resetState = await waitForAILearningResetState(page.pageConnection, page.sessionId);
      assert.equal(resetState.emptyVisible, true, 'AI learning should show the empty state after reset');

      const switchClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="ai-automation-learning-switch"]');
      assert.equal(switchClicked, true, 'AI learning switch should be clickable');
      const switchDisabled = await waitForEvaluation(page.pageConnection, page.sessionId, `document.querySelector('[data-testid="ai-automation-learning-switch"]')?.getAttribute('aria-checked') === 'false'`, 6000);
      assert.equal(switchDisabled, true, 'AI learning switch should flip to disabled after click');

      await setCdpViewport(page.pageConnection, page.sessionId, { width: 375, height: 1400, mobile: true });
      await sleep(300);
      const mobile = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        width: window.innerWidth,
        bodyWidth: document.body.scrollWidth,
        viewport: document.documentElement.clientWidth,
        pageWidth: document.querySelector('[data-testid="ai-automation-page"]')?.getBoundingClientRect().width ?? 0,
        sectionWidth: document.querySelector('[data-testid="ai-automation-learning-section"]')?.getBoundingClientRect().width ?? 0,
        summaryWidth: document.querySelector('[data-testid="ai-automation-learning-state-summary"]')?.getBoundingClientRect().width ?? 0,
        emptyWidth: document.querySelector('[data-testid="ai-automation-learning-empty"]')?.getBoundingClientRect().width ?? 0,
      }))()`);
      assert.equal(mobile.width, 375, 'AI learning mobile viewport should be set to 375px');
      assert.ok(mobile.pageWidth > 0 && mobile.sectionWidth > 0, 'AI learning shell should remain visible on mobile');
      assert.ok(mobile.summaryWidth > 0 && mobile.emptyWidth > 0, 'AI learning summary and empty state should remain visible on mobile');
      assert.ok(mobile.bodyWidth <= mobile.viewport + 24, 'AI learning page should not introduce large horizontal overflow');

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
        stateCards: loaded.stateCards,
        mobileWidth: mobile.width,
        result: 'ai-automation-learning-browser-smoke-cdp passed',
      }, null, 2));
    } else if (smokeScenario === 'ccswitch') {
      const docTriggerReady = await waitForSelector(page.pageConnection, page.sessionId, '[data-testid="navbar-doc-trigger"]', 12000);
      assert.equal(docTriggerReady, true, 'doc trigger should be visible before opening DocModal');

      const docOpened = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="navbar-doc-trigger"]');
      assert.equal(docOpened, true, 'doc trigger should be clickable');
      const ccswitchTabClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="doc-modal-tab-ccswitch"]');
      assert.equal(ccswitchTabClicked, true, 'CC Switch tab should be clickable');

      const modal = await waitForDocModalReady(page.pageConnection, page.sessionId);
      assert.equal(modal.modal, true, 'DocModal should render after opening');

      const desktop = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        width: window.innerWidth,
        modalWidth: document.querySelector('[data-testid="doc-modal"]')?.getBoundingClientRect().width ?? 0,
        tabsWidth: document.querySelector('[data-testid="doc-modal-tabs"]')?.getBoundingClientRect().width ?? 0,
        progressSteps: document.querySelectorAll('[data-testid^="ccswitch-progress-step-"]').length,
        clientButtons: document.querySelectorAll('[data-testid^="ccswitch-client-"]').length,
        importDisabled: document.querySelector('[data-testid="ccswitch-import-button"]')?.hasAttribute('disabled') ?? false,
        modelLockedVisible: !!document.querySelector('[data-testid="ccswitch-model-locked-hint"]'),
        importLockedVisible: !!document.querySelector('[data-testid="ccswitch-import-locked-hint"]'),
      }))()`);
      assert.ok(desktop.modalWidth > 0 && desktop.tabsWidth > 0, 'DocModal should stay visible on desktop');
      assert.equal(desktop.progressSteps, 4, 'CC Switch progress should show four steps');
      assert.equal(desktop.clientButtons, 2, 'CC Switch should render both client buttons');
      assert.equal(desktop.importDisabled, true, 'CC Switch import should stay disabled before selections');
      assert.equal(desktop.modelLockedVisible, true, 'main model locked hint should render before API key selection');
      assert.equal(desktop.importLockedVisible, true, 'import locked hint should render before setup completes');

      const helpScopeSelector = '[data-testid="ccswitch-panel"]';
      await resetHelpHintInteractionViaCdp(page.pageConnection, page.sessionId);
      const focusHelpSnapshot = await focusScopedHelpHintAndReadTooltip(page.pageConnection, page.sessionId, helpScopeSelector);
      assert.equal(focusHelpSnapshot.found, true, 'CC Switch panel should expose a help hint');
      assert.ok(focusHelpSnapshot.hintId, 'CC Switch help hint should have a stable id');
      assert.ok(focusHelpSnapshot.tooltipText.length > 0, 'focused CC Switch help hint should show tooltip text');

      await resetHelpHintInteractionViaCdp(page.pageConnection, page.sessionId);
      const hoverHelpSnapshot = await hoverScopedHelpHintAndReadTooltip(page.pageConnection, page.sessionId, helpScopeSelector);
      assert.equal(hoverHelpSnapshot.found, true, 'CC Switch panel should expose a hoverable help hint');
      assert.ok(hoverHelpSnapshot.tooltipText.length > 0, 'hovered CC Switch help hint should show tooltip text');

      await resetHelpHintInteractionViaCdp(page.pageConnection, page.sessionId);

      const keyOpened = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="ccswitch-key-trigger"]');
      assert.equal(keyOpened, true, 'API key selector should open');
      const keySelected = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid^="ccswitch-key-item-"]');
      assert.equal(keySelected, true, 'API key option should be selectable');

      const modelUnlocked = await waitForSelector(page.pageConnection, page.sessionId, '[data-testid="ccswitch-name-locked-hint"]', 6000);
      assert.equal(modelUnlocked, true, 'name locked hint should remain visible until model is chosen');

      const modelOpened = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="ccswitch-model-trigger"]');
      assert.equal(modelOpened, true, 'main model selector should open after API key selection');
      const modelSelected = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid^="ccswitch-model-item-"]');
      assert.equal(modelSelected, true, 'main model option should be selectable');

      const nameInputVisible = await waitForSelector(page.pageConnection, page.sessionId, '[data-testid="ccswitch-name-input"]', 6000);
      assert.equal(nameInputVisible, true, 'profile name input should unlock after model selection');

      const nameFilled = await evaluateCdp(page.pageConnection, page.sessionId, `(() => {
        const input = document.querySelector('[data-testid="ccswitch-name-input"]');
        if (!(input instanceof HTMLInputElement)) {
          return { found: false, value: '' };
        }
        input.focus();
        input.value = 'octopus_browser_ccswitch';
        input.dispatchEvent(new Event('input', { bubbles: true }));
        input.dispatchEvent(new Event('change', { bubbles: true }));
        return { found: true, value: input.value };
      })()`);
      assert.equal(nameFilled.found, true, 'profile name input should be writable');
      assert.equal(nameFilled.value, 'octopus_browser_ccswitch', 'profile name input should accept custom value');

      const advancedVisible = await waitForSelector(page.pageConnection, page.sessionId, '[data-testid="ccswitch-advanced-trigger"]', 6000);
      assert.equal(advancedVisible, true, 'advanced mapping should unlock for Claude after name confirmation');
      const advancedOpened = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="ccswitch-advanced-trigger"]');
      assert.equal(advancedOpened, true, 'advanced mapping trigger should be clickable');
      const advancedGridVisible = await waitForSelector(page.pageConnection, page.sessionId, '[data-testid="ccswitch-advanced-grid"]', 6000);
      assert.equal(advancedGridVisible, true, 'advanced mapping grid should expand on demand');

      const importEnabled = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        importDisabled: document.querySelector('[data-testid="ccswitch-import-button"]')?.hasAttribute('disabled') ?? true,
        importLockedVisible: !!document.querySelector('[data-testid="ccswitch-import-locked-hint"]'),
      }))()`);
      assert.equal(importEnabled.importDisabled, false, 'import button should enable after the required steps');
      assert.equal(importEnabled.importLockedVisible, false, 'import locked hint should disappear after the required steps');

      await setCdpViewport(page.pageConnection, page.sessionId, { width: 375, height: 1200, mobile: true });
      await sleep(300);
      const mobile = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        width: window.innerWidth,
        bodyWidth: document.body.scrollWidth,
        viewport: document.documentElement.clientWidth,
        modalWidth: document.querySelector('[data-testid="doc-modal"]')?.getBoundingClientRect().width ?? 0,
        tabsWidth: document.querySelector('[data-testid="doc-modal-tabs"]')?.getBoundingClientRect().width ?? 0,
        progressWidth: document.querySelector('[data-testid="ccswitch-progress-card"]')?.getBoundingClientRect().width ?? 0,
        importWidth: document.querySelector('[data-testid="ccswitch-import-button"]')?.getBoundingClientRect().width ?? 0,
      }))()`);
      assert.equal(mobile.width, 375, 'mobile viewport should be set to 375px');
      assert.ok(mobile.modalWidth > 0 && mobile.progressWidth > 0 && mobile.importWidth > 0, 'CC Switch modal should remain visible on mobile');
      assert.ok(mobile.tabsWidth > 0, 'CC Switch tabs should remain visible on mobile');
      assert.ok(mobile.bodyWidth <= mobile.viewport + 24, 'CC Switch modal should not introduce large horizontal overflow');

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
        progressSteps: desktop.progressSteps,
        tooltipCount: hoverHelpSnapshot.tooltipCount,
        interactionChecks: 2,
        mobileWidth: mobile.width,
        result: 'ccswitch-browser-smoke-cdp passed',
      }, null, 2));
    } else {
      const channelPage = await waitForChannelPageReady(page.pageConnection, page.sessionId);
      assert.equal(channelPage.channelTrigger, true, 'channel create trigger should exist on channel page');

      const opened = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="toolbar-create-trigger-channel"]');
      assert.equal(opened, true, 'channel create trigger should be clickable');
      await waitForDialogOpen(page.pageConnection, page.sessionId);

      const desktop = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        width: window.innerWidth,
        dialog: !!document.querySelector('[data-testid="channel-create-dialog"]'),
        flowTitle: document.querySelector('[data-testid="new-channel-flow-card"]')?.textContent?.trim() || '',
        basicTitle: document.querySelector('[data-testid="new-channel-basic-section"]')?.textContent?.trim() || '',
        keySection: !!document.querySelector('[data-testid="new-channel-key-section"]'),
        keyItems: document.querySelectorAll('[data-testid^="new-channel-key-item-"]').length,
        helpButtons: document.querySelectorAll('[data-testid="channel-create-dialog"] ${helpButtonSelector}').length,
      }))()`);
      assert.equal(desktop.dialog, true, 'channel create dialog should stay open on desktop');
      assert.ok(desktop.flowTitle.length > 0, 'flow summary should be visible in create dialog');
      assert.ok(desktop.basicTitle.length > 0, 'basic section heading should be visible');
      assert.equal(desktop.keySection, true, 'key section should render');
      assert.ok(desktop.keyItems >= 1, 'at least one key item should render');
      assert.ok(desktop.helpButtons >= 6, 'channel create dialog should show multiple help hints');

      const firstKeyClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="new-channel-key-trigger-0"]');
      assert.equal(firstKeyClicked, true, 'first key card trigger should exist');
      const expanded = await waitForKeyItemExpanded(page.pageConnection, page.sessionId, '[data-testid="new-channel-key-item-0"]');
      assert.equal(expanded.state, 'open', 'first key item should expand');
      assert.equal(expanded.primaryVisible, true, 'expanded key primary section should become visible');
      assert.equal(expanded.valueInput, true, 'expanded key primary area should include the real key input');
      assert.ok(expanded.primaryText.length > 0, 'expanded key primary area should contain explanatory copy');
      assert.ok(expanded.statusBadges.length > 0, 'expanded key primary area should surface status text');

      const helpSnapshot = await focusHelpHintAndReadTooltip(page.pageConnection, page.sessionId);
      assert.equal(helpSnapshot.found, true, 'channel create dialog should expose a help hint button');
      assert.ok(helpSnapshot.hintId, 'focused help button should have a stable help-hint id');
      assert.ok(helpSnapshot.tooltipText.length > 0, 'focused help button should show tooltip text');

      const modelMock = await createChannelFetchModelMockServer(channelFetchModelMockPort);
      try {
        const baseUrlFilled = await setInputValueViaCdp(page.pageConnection, page.sessionId, '[id="new-channel-base-0"]', `http://127.0.0.1:${channelFetchModelMockPort}`);
        assert.equal(baseUrlFilled.found, true, 'base url input should be writable');
        assert.equal(baseUrlFilled.value, `http://127.0.0.1:${channelFetchModelMockPort}`, 'base url should accept the mock upstream url');

        const keyFilled = await setInputValueViaCdp(page.pageConnection, page.sessionId, '[id="new-channel-key-value-0"]', 'sk-browser-fetch-model');
        assert.equal(keyFilled.found, true, 'key input should be writable after expanding the first key item');
        assert.equal(keyFilled.value, 'sk-browser-fetch-model', 'key input should accept the smoke credential');

        const fetchClicked = await clickSelectorViaCdp(page.pageConnection, page.sessionId, '[data-testid="new-channel-key-fetch-models-0"]');
        assert.equal(fetchClicked, true, 'per-key fetch-model button should be clickable');

        const dialogReady = await waitForSelector(page.pageConnection, page.sessionId, '[data-testid="new-channel-model-select-dialog"]', 8000);
        assert.equal(dialogReady, true, 'model select dialog should open after fetching models');

        const modelListSnapshot = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
          title: document.querySelector('[data-testid="new-channel-model-select-title"]')?.textContent?.trim() || '',
          optionCount: document.querySelectorAll('[data-testid^="new-channel-model-option-"]').length,
          options: Array.from(document.querySelectorAll('[data-testid^="new-channel-model-option-"]')).map((node) => node.textContent?.trim() || '').filter(Boolean),
        }))()`);
        assert.ok(modelListSnapshot.title.length > 0, 'model select dialog should expose a readable title');
        assert.ok(modelListSnapshot.optionCount >= channelFetchModelList.length, 'model select dialog should render fetched model options');
        assert.ok(modelListSnapshot.options.some((item) => item.includes(channelFetchModelList[0])), 'fetched models should appear in the selection dialog');

        const mockRequests = modelMock.getRequests();
        assert.ok(mockRequests.length >= 1, 'fetch-model smoke mock should receive at least one request');
        assert.equal(mockRequests.at(-1)?.authorization, 'Bearer sk-browser-fetch-model', 'fetch-model request should forward the entered key as bearer token');
      } finally {
        await new Promise((resolve) => modelMock.server.close(() => resolve()));
      }

      await setCdpViewport(page.pageConnection, page.sessionId, { width: 375, height: 1200, mobile: true });
      await sleep(300);
      const mobile = await evaluateCdp(page.pageConnection, page.sessionId, `(() => ({
        width: window.innerWidth,
        dialogWidth: document.querySelector('[data-testid="channel-create-dialog"]')?.getBoundingClientRect().width ?? 0,
        bodyWidth: document.body.scrollWidth,
        viewport: document.documentElement.clientWidth,
        keyTriggerWidth: document.querySelector('[data-testid="new-channel-key-trigger-0"]')?.getBoundingClientRect().width ?? 0,
      }))()`);

      assert.equal(mobile.width, 375, 'mobile viewport should be set to 375px');
      assert.ok(mobile.dialogWidth > 0, 'dialog should remain visible on mobile');
      assert.ok(mobile.keyTriggerWidth > 0, 'key trigger should remain visible on mobile');
      assert.ok(mobile.bodyWidth <= mobile.viewport + 24, 'mobile dialog should not introduce large horizontal overflow');

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
        desktopHelpButtons: desktop.helpButtons,
        keyItems: desktop.keyItems,
        fetchedModels: channelFetchModelList.length,
        mobileWidth: mobile.width,
        result: 'channel-create-browser-smoke-cdp passed',
      }, null, 2));
    }
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
