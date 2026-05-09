import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { API_BASE_URL } from '../client';
import { loginAPIKeySession, type APIKeyAuthStatus } from './apikey';
import { useAuthStore } from './user';

type FetchMock = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;

const successStatus: APIKeyAuthStatus = {
	ok: true,
	api_key_id: 7,
	name: 'dashboard-key',
	enabled: true,
	expire_at: 1_800_000_000,
	supported_models: 'gpt-4o,claude-3-5-sonnet',
	auth_mode: 'api_key',
};

function resetAuthStore() {
	useAuthStore.setState({
		isAuthenticated: false,
		isLoading: false,
		isAPIKeyAuth: false,
		token: null,
		expireAt: null,
		mustChangePassword: false,
		apiKeyStatus: null,
	});
}

describe('api key auth flow', () => {
	beforeEach(() => {
		resetAuthStore();
		window.localStorage.clear();
	});

	afterEach(() => {
		resetAuthStore();
		window.localStorage.clear();
		vi.restoreAllMocks();
		vi.unstubAllGlobals();
	});

	it('validates with server before persisting api key auth', async () => {
		const fetchMock = vi.fn<FetchMock>(async () => new Response(JSON.stringify({
			code: 200,
			message: 'success',
			data: successStatus,
		}), {
			status: 200,
			headers: { 'content-type': 'application/json' },
		}));
		vi.stubGlobal('fetch', fetchMock);

		const persistAuth = vi.fn((apiKey: string, status: APIKeyAuthStatus) => {
			useAuthStore.getState().setAPIKeyAuth(apiKey, status);
		});

		const result = await loginAPIKeySession('  sk-octopus-test-123  ', persistAuth);

		expect(result).toEqual(successStatus);
		expect(fetchMock).toHaveBeenCalledTimes(1);
		expect(fetchMock).toHaveBeenCalledWith(`${API_BASE_URL}api/v1/apikey/login`, expect.objectContaining({
			method: 'GET',
			headers: expect.objectContaining({
				Authorization: 'Bearer sk-octopus-test-123',
			}),
		}));
		expect(persistAuth).toHaveBeenCalledWith('sk-octopus-test-123', successStatus);

		const state = useAuthStore.getState();
		expect(state.isAuthenticated).toBe(true);
		expect(state.isAPIKeyAuth).toBe(true);
		expect(state.token).toBe('sk-octopus-test-123');
		expect(state.apiKeyStatus).toEqual(successStatus);
	});

	it('does not persist invalid api key auth state on failed validation', async () => {
		const fetchMock = vi.fn<FetchMock>(async () => new Response(JSON.stringify({
			code: 401,
			message: 'API key is disabled',
		}), {
			status: 401,
			headers: { 'content-type': 'application/json' },
		}));
		vi.stubGlobal('fetch', fetchMock);

		const persistAuth = vi.fn((apiKey: string, status: APIKeyAuthStatus) => {
			useAuthStore.getState().setAPIKeyAuth(apiKey, status);
		});

		await expect(loginAPIKeySession('sk-octopus-disabled-123', persistAuth)).rejects.toThrow('API key is disabled');
		expect(persistAuth).not.toHaveBeenCalled();

		const state = useAuthStore.getState();
		expect(state.isAuthenticated).toBe(false);
		expect(state.isAPIKeyAuth).toBe(false);
		expect(state.token).toBeNull();
		expect(state.apiKeyStatus).toBeNull();
	});

	it('restores structured api key status during auth check', async () => {
		useAuthStore.setState({
			isAuthenticated: true,
			isLoading: true,
			isAPIKeyAuth: true,
			token: 'sk-octopus-restore-123',
			expireAt: null,
			mustChangePassword: false,
			apiKeyStatus: null,
		});

		const fetchMock = vi.fn<FetchMock>(async () => new Response(JSON.stringify({
			code: 200,
			message: 'success',
			data: successStatus,
		}), {
			status: 200,
			headers: { 'content-type': 'application/json' },
		}));
		vi.stubGlobal('fetch', fetchMock);

		await useAuthStore.getState().checkAuth();

		expect(fetchMock).toHaveBeenCalledTimes(1);
		const [input, init] = fetchMock.mock.calls[0] as [RequestInfo | URL, RequestInit | undefined];
		expect(input).toBe(`${API_BASE_URL}api/v1/apikey/login`);
		expect(init?.headers).toBeInstanceOf(Headers);
		expect((init?.headers as Headers).get('Authorization')).toBe('Bearer sk-octopus-restore-123');

		const state = useAuthStore.getState();
		expect(state.isAuthenticated).toBe(true);
		expect(state.isAPIKeyAuth).toBe(true);
		expect(state.isLoading).toBe(false);
		expect(state.mustChangePassword).toBe(false);
		expect(state.apiKeyStatus).toEqual(successStatus);
	});
});
