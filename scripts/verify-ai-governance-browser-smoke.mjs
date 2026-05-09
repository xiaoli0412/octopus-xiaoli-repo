import assert from 'node:assert/strict';
import { existsSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { spawn } from 'node:child_process';
import { createServer as createNetServer } from 'node:net';
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

const explicitFrontendUrl = process.env.OCTOPUS_UI_SMOKE_FRONTEND_URL?.trim() || '';
const explicitBackendUrl = process.env.OCTOPUS_UI_SMOKE_BACKEND_URL?.trim() || '';
let frontPort = Number(process.env.OCTOPUS_UI_SMOKE_FRONTEND_PORT || 3101);
let backPort = Number(process.env.OCTOPUS_UI_SMOKE_BACKEND_PORT || 18081);
let frontendBaseUrl = normalizeBaseUrl(explicitFrontendUrl || `http://127.0.0.1:${frontPort}`);
let backendBaseUrl = normalizeBaseUrl(explicitBackendUrl || `http://127.0.0.1:${backPort}`);
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
const sessionName = process.env.OCTOPUS_UI_SMOKE_SESSION || `octopus-ai-governance-${process.pid}`;
const playwrightPackage = process.env.OCTOPUS_UI_SMOKE_PLAYWRIGHT_PACKAGE || '@playwright/cli';
const playwrightCliTimeoutMs = Number(process.env.OCTOPUS_UI_SMOKE_CLI_TIMEOUT_MS || 30000);
let governanceMockPort = Number(process.env.OCTOPUS_UI_SMOKE_AI_GOVERNANCE_MOCK_PORT || 19191);
const governanceManagedGroupName = process.env.OCTOPUS_UI_SMOKE_AI_GOVERNANCE_GROUP || 'AI Governance Managed';
const governanceChannelName = process.env.OCTOPUS_UI_SMOKE_AI_GOVERNANCE_CHANNEL || 'octopus-browser-ai-governance-channel';
const governanceAPIKeyName = process.env.OCTOPUS_UI_SMOKE_AI_GOVERNANCE_APIKEY || 'octopus-browser-ai-governance-key';

function smokeLog(message, details = {}) {
	const suffix = Object.keys(details).length > 0 ? ` ${JSON.stringify(details)}` : '';
	console.log(`[ai-governance-smoke] ${message}${suffix}`);
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
		'  node scripts/verify-ai-governance-browser-smoke.mjs --check-only',
		'  node scripts/verify-ai-governance-browser-smoke.mjs --external',
		'  node scripts/verify-ai-governance-browser-smoke.mjs --self-start',
		'',
		'Environment overrides:',
		'  OCTOPUS_UI_SMOKE_MODE=check-only|external|self-start',
		'  OCTOPUS_UI_SMOKE_FRONTEND_URL=http://127.0.0.1:3101',
		'  OCTOPUS_UI_SMOKE_BACKEND_URL=http://127.0.0.1:18081',
		'  OCTOPUS_UI_SMOKE_AI_GOVERNANCE_MOCK_PORT=19191',
		'  OCTOPUS_UI_SMOKE_BACKEND_BIN=build/octopus-smoke.exe',
		'  OCTOPUS_UI_SMOKE_NPX_SCRIPT=D:/gol1/node_modules/npm/bin/npx-cli.js',
		'  OCTOPUS_UI_SMOKE_ADMIN_TOKEN=<jwt>',
	].join('\n'));
}

function refreshRuntimeBaseUrls() {
	frontendBaseUrl = normalizeBaseUrl(explicitFrontendUrl || `http://127.0.0.1:${frontPort}`);
	backendBaseUrl = normalizeBaseUrl(explicitBackendUrl || `http://127.0.0.1:${backPort}`);
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

async function stopChildProcess(child) {
	if (!child || child.killed || child.exitCode !== null) {
		return;
	}

	if (process.platform === 'win32') {
		await new Promise((resolve) => {
			const killer = spawn('taskkill', ['/PID', String(child.pid), '/T', '/F'], { cwd: repoRoot, windowsHide: true, stdio: 'ignore' });
			const timeout = setTimeout(resolve, 10000);
			killer.on('close', () => {
				clearTimeout(timeout);
				resolve();
			});
			killer.on('error', () => {
				clearTimeout(timeout);
				resolve();
			});
		});
		return;
	}

	try {
		child.kill('SIGTERM');
	} catch {
		return;
	}
	await new Promise((resolve) => {
		const timeout = setTimeout(() => {
			try {
				child.kill('SIGKILL');
			} catch {
			}
			resolve();
		}, 5000);
		child.once('close', () => {
			clearTimeout(timeout);
			resolve();
		});
		child.once('error', () => {
			clearTimeout(timeout);
			resolve();
		});
	});
}

function sleep(ms) {
	return new Promise((resolve) => setTimeout(resolve, ms));
}

async function canListenOnPort(port) {
	return new Promise((resolve) => {
		const server = createNetServer();
		server.once('error', () => {
			resolve(false);
		});
		server.listen(port, '127.0.0.1', () => {
			server.close(() => resolve(true));
		});
	});
}

async function getFreePort(exclude = new Set()) {
	for (let attempt = 0; attempt < 12; attempt += 1) {
		const port = await new Promise((resolve, reject) => {
			const server = createNetServer();
			server.once('error', reject);
			server.listen(0, '127.0.0.1', () => {
				const address = server.address();
				const selectedPort = typeof address === 'object' && address ? address.port : 0;
				server.close((closeError) => {
					if (closeError) {
						reject(closeError);
						return;
					}
					resolve(selectedPort);
				});
			});
		});
		if (typeof port === 'number' && port > 0 && !exclude.has(port)) {
			return port;
		}
	}
	throw new Error('Unable to resolve a free local TCP port for governance smoke');
}

async function resolveSelfStartPorts() {
	const reservedPorts = new Set();

	if (!explicitBackendUrl) {
		const selectedPort = await getFreePort(reservedPorts);
		smokeLog('self-start backend port selected', { requestedPort: backPort, selectedPort });
		backPort = selectedPort;
		reservedPorts.add(backPort);
	}

	if (!explicitFrontendUrl) {
		frontPort = backPort;
		reservedPorts.add(frontPort);
	}

	const selectedMockPort = await getFreePort(reservedPorts);
	if (selectedMockPort !== governanceMockPort) {
		smokeLog('self-start mock upstream port selected', { requestedPort: governanceMockPort, selectedPort: selectedMockPort });
	}
	governanceMockPort = selectedMockPort;
	reservedPorts.add(governanceMockPort);

	refreshRuntimeBaseUrls();

	if (!explicitFrontendUrl) {
		frontendBaseUrl = backendBaseUrl;
	}
}

async function removeDirectoryWithRetry(targetPath, attempts = 8, delayMs = 500) {
	let lastError;
	for (let attempt = 0; attempt < attempts; attempt += 1) {
		try {
			rmSync(targetPath, { recursive: true, force: true });
			return;
		} catch (error) {
			lastError = error;
			await sleep(delayMs * (attempt + 1));
		}
	}
	throw lastError;
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

async function apiPost(token, routePath, payload) {
	const response = await fetch(`${backendBaseUrl}${routePath}`, {
		method: 'POST',
		headers: {
			'content-type': 'application/json',
			Authorization: `Bearer ${token}`,
		},
		body: JSON.stringify(payload),
	});
	if (!response.ok) {
		const text = await response.text();
		throw new Error(`Request failed for ${routePath}: ${response.status} ${response.statusText}\n${text}`);
	}
	return response.json();
}

async function apiGet(token, routePath) {
	const response = await fetch(`${backendBaseUrl}${routePath}`, {
		headers: {
			Authorization: `Bearer ${token}`,
		},
	});
	if (!response.ok) {
		const text = await response.text();
		throw new Error(`Request failed for ${routePath}: ${response.status} ${response.statusText}\n${text}`);
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

async function createGovernanceMockServer(port) {
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

			const modelName = typeof payload.model === 'string' && payload.model ? payload.model : 'gpt-4o';
			const response = {
				id: 'chatcmpl-ai-governance-smoke-1',
				object: 'chat.completion',
				created: 1713436800,
				model: modelName,
				choices: [{ index: 0, message: { role: 'assistant', content: 'ai-governance-smoke-ok' }, finish_reason: 'stop' }],
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

async function ensureGovernanceSeedData(token) {
	await apiPost(token, '/api/v1/setting/set', {
		key: 'ai_automation_enabled',
		value: 'true',
	});
	await apiPost(token, '/api/v1/setting/set', {
		key: 'dynamic_routing_learning_enabled',
		value: 'true',
	});
	await apiPost(token, '/api/v1/setting/set', {
		key: 'ai_governance_managed_group_name',
		value: governanceManagedGroupName,
	});

	const mockServer = await createGovernanceMockServer(governanceMockPort);

	try {
		await apiPost(token, '/api/v1/model/create', {
			name: 'gpt-4o',
			canonical_name: 'gpt-4o',
			input: 1,
			output: 2,
			cache_read: 0,
			cache_write: 0,
			official_input: 1,
			official_output: 2,
			official_cache_read: 0,
			official_cache_write: 0,
			billing_mode: 'per_token',
			probe_policy: 'concurrent',
			probe_interval_seconds: 300,
			probe_concurrency_limit: 2,
		});
		await apiPost(token, '/api/v1/model/create', {
			name: 'gpt-4.1',
			canonical_name: 'gpt-4.1',
			input: 1.2,
			output: 2.4,
			cache_read: 0,
			cache_write: 0,
			official_input: 1.2,
			official_output: 2.4,
			official_cache_read: 0,
			official_cache_write: 0,
			billing_mode: 'per_token',
			probe_policy: 'sequential',
			probe_interval_seconds: 600,
			probe_concurrency_limit: 1,
		});

		const createdChannel = await apiPost(token, '/api/v1/channel/create', {
			name: governanceChannelName,
			type: 0,
			enabled: true,
			base_urls: [{ url: `http://127.0.0.1:${governanceMockPort}`, delay: 0 }],
			keys: [{ enabled: true, channel_key: 'ai-governance-upstream-key', source_type: 'private/internal', allowed_models: 'gpt-4o,gpt-4.1' }],
			model: 'gpt-4o,gpt-4.1',
		});

		await apiPost(token, '/api/v1/group/create', {
			name: 'legacy-governance-group',
			mode: 1,
			items: [{ channel_id: createdChannel.data.id, model_name: 'gpt-4o', priority: 1, weight: 1 }],
		});

		const createdAPIKey = await apiPost(token, '/api/v1/apikey/create', {
			name: governanceAPIKeyName,
			enabled: true,
		});

		const gatewayResponse = await fetch(`${backendBaseUrl}/v1/chat/completions`, {
			method: 'POST',
			headers: {
				'content-type': 'application/json',
				Authorization: `Bearer ${createdAPIKey.data.api_key}`,
			},
			body: JSON.stringify({
				model: 'legacy-governance-group',
				messages: [{ role: 'user', content: 'browser smoke ai governance seed' }],
			}),
		});
		if (!gatewayResponse.ok) {
			const text = await gatewayResponse.text();
			throw new Error(`Gateway seed request failed: ${gatewayResponse.status} ${gatewayResponse.statusText}\n${text}`);
		}

		await sleep(500);
		const overview = await apiGet(token, '/api/v1/ai/overview');
		const learning = await apiGet(token, '/api/v1/dynamic-routing/learning');
		const groups = await apiGet(token, '/api/v1/group/list');

		return {
			overviewEnabled: Boolean(overview.data?.enabled),
			learningEnabled: Boolean(learning.data?.enabled),
			learningStateCount: Array.isArray(learning.data?.states) ? learning.data.states.length : 0,
			groupCount: Array.isArray(groups.data) ? groups.data.length : 0,
		};
	} finally {
		await new Promise((resolve) => mockServer.close(() => resolve()));
	}
}

async function navigateToFrontendShell() {
	await runPlaywrightCli(['eval', `() => {
		window.location.replace(${JSON.stringify(frontendBaseUrl)});
		return true;
	}`], { raw: true });
}

async function waitForLoginPage(timeoutMs = 30000) {
	const startedAt = Date.now();
	const loginEval = `() => JSON.stringify({ username: !!document.querySelector('[data-testid=login-username-input]'), password: !!document.querySelector('[data-testid=login-password-input]'), submit: !!document.querySelector('[data-testid=login-submit-label]'), bodyText: document.body.innerText })`;
	let lastBodyText = '';
	let lastErrorMessage = '';

	while (Date.now() - startedAt < timeoutMs) {
		try {
			const result = await runPlaywrightCli(['eval', loginEval], { raw: true });
			const snapshot = parsePlaywrightJson(result.stdout);
			lastBodyText = snapshot.bodyText || '';
			lastErrorMessage = '';
			if (snapshot.username && snapshot.password && snapshot.submit) {
				return snapshot;
			}
		} catch (error) {
			lastErrorMessage = error instanceof Error ? error.message : String(error);
		}
		await sleep(1000);
	}

	const errorSuffix = lastErrorMessage ? ` Last eval error: ${lastErrorMessage}` : '';
	throw new Error(`Timed out waiting for login page on ${frontendBaseUrl}. Last body text snippet: ${lastBodyText.slice(0, 200)}${errorSuffix}`);
}

async function waitForAuthenticatedShell(timeoutMs = 45000) {
	const startedAt = Date.now();
	const shellEval = `() => JSON.stringify({ navAI: !!document.querySelector('[data-testid=navbar-route-ai]'), bodyText: document.body.innerText })`;
	let lastBodyText = '';
	let lastErrorMessage = '';

	while (Date.now() - startedAt < timeoutMs) {
		try {
			const result = await runPlaywrightCli(['eval', shellEval], { raw: true });
			const snapshot = parsePlaywrightJson(result.stdout);
			lastBodyText = snapshot.bodyText || '';
			lastErrorMessage = '';
			if (snapshot.navAI) {
				return snapshot;
			}
		} catch (error) {
			lastErrorMessage = error instanceof Error ? error.message : String(error);
		}
		await sleep(1000);
	}

	const errorSuffix = lastErrorMessage ? ` Last eval error: ${lastErrorMessage}` : '';
	throw new Error(`Timed out waiting for authenticated shell on ${frontendBaseUrl}. Last body text snippet: ${lastBodyText.slice(0, 200)}${errorSuffix}`);
}

async function performLogin() {
	await runPlaywrightCli(['eval', `() => {
			const username = document.querySelector('[data-testid=login-username-input]');
			if (username instanceof HTMLInputElement) {
				username.value = ${JSON.stringify(adminUsername)};
				username.dispatchEvent(new Event('input', { bubbles: true }));
			}
			const password = document.querySelector('[data-testid=login-password-input]');
			if (password instanceof HTMLInputElement) {
				password.value = ${JSON.stringify(adminPassword)};
				password.dispatchEvent(new Event('input', { bubbles: true }));
			}
		}`], { raw: true });
	await runPlaywrightCli(['eval', `() => {
			const submit = document.querySelector('[data-testid=login-submit-label]');
			const button = submit ? submit.closest('button') : document.querySelector('button[type=submit]');
			if (button instanceof HTMLButtonElement) {
				button.click();
			}
		}`], { raw: true });
}

async function waitForGovernancePageReady(timeoutMs = 45000) {
	const startedAt = Date.now();
	let lastBodyText = '';
	let lastErrorMessage = '';
	let lastStdout = '';
	const snapshotEval = `() => JSON.stringify({ page: !!document.querySelector('[data-testid=ai-automation-page]'), goalInput: !!document.querySelector('[data-testid=ai-governance-goal-input]'), createButton: !!document.querySelector('[data-testid=ai-governance-create-button]'), planTab: !!document.querySelector('[data-testid=ai-governance-tab-plan]'), previewTab: !!document.querySelector('[data-testid=ai-governance-tab-preview]'), pageText: document.body.innerText })`;

	while (Date.now() - startedAt < timeoutMs) {
		try {
			const result = await runPlaywrightCli(['eval', snapshotEval], { raw: true });
			lastStdout = result.stdout;
			const snapshot = parsePlaywrightJson(result.stdout);
			lastBodyText = snapshot.pageText || '';
			lastErrorMessage = '';
			if (snapshot.page && snapshot.goalInput && snapshot.createButton && snapshot.planTab && snapshot.previewTab) {
				return snapshot;
			}
		} catch (error) {
			lastErrorMessage = error instanceof Error ? error.message : String(error);
		}
		await sleep(1000);
	}

	const errorSuffix = lastErrorMessage ? ` Last eval error: ${lastErrorMessage}` : '';
	const stdoutSuffix = lastStdout ? ` Last stdout: ${lastStdout.slice(0, 300)}` : '';
	throw new Error(`Timed out waiting for AI governance page on ${frontendBaseUrl}. Last body text snippet: ${lastBodyText.slice(0, 200)}${errorSuffix}${stdoutSuffix}`);
}

async function waitForGovernanceSessionResult(timeoutMs = 45000) {
	const startedAt = Date.now();
	let lastBodyText = '';
	let lastErrorMessage = '';
	const previewEval = `() => JSON.stringify({ previewVisible: !!document.querySelector('[data-testid=ai-governance-workspace-preview]'), previewText: document.querySelector('[data-testid=ai-governance-workspace-preview]') ? document.querySelector('[data-testid=ai-governance-workspace-preview]').innerText : '', detailsText: document.querySelector('[data-testid=ai-governance-workspace-details]') ? document.querySelector('[data-testid=ai-governance-workspace-details]').innerText : '', bodyText: document.body.innerText })`;

	while (Date.now() - startedAt < timeoutMs) {
		try {
			const result = await runPlaywrightCli(['eval', previewEval], { raw: true });
			const snapshot = parsePlaywrightJson(result.stdout);
			lastBodyText = snapshot.bodyText || '';
			lastErrorMessage = '';
			if (snapshot.previewVisible && /group_upsert|group_item_attach|group_item_reorder|typed/i.test(snapshot.previewText)) {
				return snapshot;
			}
		} catch (error) {
			lastErrorMessage = error instanceof Error ? error.message : String(error);
		}
		await sleep(1000);
	}

	const errorSuffix = lastErrorMessage ? ` Last eval error: ${lastErrorMessage}` : '';
	throw new Error(`Timed out waiting for AI governance preview result. Last body text snippet: ${lastBodyText.slice(0, 300)}${errorSuffix}`);
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
			governanceManagedGroupName,
			backendLaunch: backendBin ? { kind: 'binary', command: backendBin } : { kind: 'go-run', command: goExe },
			frontendLaunch: { command: nodeExe, cwd: 'web', nextApiBaseUrl: backendBaseUrl },
			nodeExe,
			npxCliScript,
			note: 'check-only does not spawn backend, frontend, or browser processes',
		}, null, 2));
		return;
	}

	const tempDir = mkdtempSync(path.join(tmpdir(), 'octopus-ai-governance-browser-'));
	const browserProfile = path.join(tempDir, 'pw-profile');
	const ownedChildren = [];
	let browserOpened = false;
	let primaryError = null;

	try {
		if (mode === 'self-start') {
			await resolveSelfStartPorts();
			smokeLog('starting local services', { frontendBaseUrl, backendBaseUrl });
			const configPath = writeBackendConfig(tempDir);
			ownedChildren.push(startBackend(configPath));
			await waitForHttp(`${backendBaseUrl}/healthz`, 45000);
			await waitForHttp(frontendBaseUrl, 20000);
		} else {
			smokeLog('using external services', { frontendBaseUrl, backendBaseUrl });
			await waitForHttp(`${backendBaseUrl}/healthz`, 15000);
			await waitForHttp(frontendBaseUrl, 20000);
		}

		smokeLog('resolving admin token');
		const token = await resolveAdminToken();
		smokeLog('seeding governance fixture');
		const seedResult = await ensureGovernanceSeedData(token);
		assert.equal(seedResult.overviewEnabled, true, 'ai governance overview should be enabled after seed');
		assert.equal(seedResult.learningEnabled, true, 'dynamic routing learning should be enabled after seed');
		assert.ok(seedResult.learningStateCount > 0, 'seed should create at least one learning state');
		assert.ok(seedResult.groupCount > 0, 'seed should create at least one group');

		smokeLog('opening browser', { browserName, sessionName });
		await runPlaywrightCli(['open', frontendBaseUrl, '--browser', browserName, '--profile', browserProfile]);
		browserOpened = true;
		smokeLog('waiting for login page');
		await waitForLoginPage();
		smokeLog('logging into workbench');
		await performLogin();
		await waitForAuthenticatedShell();
		await runPlaywrightCli(['eval', `() => {
			const aiButton = document.querySelector('[data-testid=navbar-route-ai]');
			if (aiButton instanceof HTMLButtonElement) {
				aiButton.click();
			}
		}`], { raw: true });

		smokeLog('waiting for governance page');
		const ready = await waitForGovernancePageReady();
		assert.equal(ready.page, true, 'ai governance page should render');
		assert.ok(ready.pageText.includes('AI 自动化 V2') || ready.pageText.includes('AI Automation V2'), 'ai governance hero should be visible');

		smokeLog('creating governance session from UI');
		await runPlaywrightCli(['eval', `() => {
				const input = document.querySelector('[data-testid=ai-governance-goal-input]');
				if (input instanceof HTMLInputElement) {
					input.value = '请整理当前路由与分组，把可用模型统一收口到治理组。';
					input.dispatchEvent(new Event('input', { bubbles: true }));
				}
			}`], { raw: true });
		await runPlaywrightCli(['eval', `() => {
				const button = document.querySelector('[data-testid=ai-governance-create-button]');
				if (button instanceof HTMLButtonElement) button.click();
			}`], { raw: true });

		await sleep(1200);
		await runPlaywrightCli(['eval', `() => {
				const button = document.querySelector('[data-testid=ai-governance-tab-preview]');
				if (button instanceof HTMLButtonElement) button.click();
			}`], { raw: true });

		smokeLog('waiting for governance preview');
		const preview = await waitForGovernanceSessionResult();
		assert.equal(preview.previewVisible, true, 'preview panel should render');
		assert.ok(/group_upsert|group_item_attach|group_item_reorder/i.test(preview.previewText), 'preview should include typed governance mutations');
		assert.ok(/checksum/i.test(preview.detailsText) || /current session/i.test(preview.detailsText), 'details rail should include current session data');

		console.log(JSON.stringify({
			mode,
			frontend: frontendBaseUrl,
			backend: backendBaseUrl,
			seed: seedResult,
			result: 'ai-governance-browser-smoke passed',
		}, null, 2));
	} catch (error) {
		primaryError = error;
		throw error;
	} finally {
		if (browserOpened) {
			try {
				await runPlaywrightCli(['close']);
			} catch (closeError) {
				if (!primaryError) {
					primaryError = closeError;
				}
			}
		}

		for (const child of ownedChildren) {
			try {
				await stopChildProcess(child);
			} catch (stopError) {
				if (!primaryError) {
					primaryError = stopError;
				}
			}
		}

		try {
			await removeDirectoryWithRetry(tempDir);
		} catch (cleanupError) {
			if (!primaryError) {
				throw cleanupError;
			}
			smokeLog('cleanup warning', { message: cleanupError instanceof Error ? cleanupError.message : String(cleanupError), tempDir });
		}
	}
}

main().catch((error) => {
	console.error(error.stack || String(error));
	process.exit(1);
});
