import { createElement, type ReactNode } from 'react';
import { act, renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useAuthStore } from './user';
import { APIKeyAuthStatus, useAPIKeyLogin } from './apikey';
import { API_BASE_URL } from '../client';

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

function createWrapper() {
	const queryClient = new QueryClient({
		defaultOptions: {
			queries: { retry: false },
			mutations: { retry: false },
		},
	});

	return function Wrapper({ children }: { children: ReactNode }) {
		return createElement(QueryClientProvider, { client: queryClient }, children);
	};
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

		const { result: hook } = renderHook(() => useAPIKeyLogin(), { wrapper: createWrapper() });
		let result: APIKeyAuthStatus | undefined;
		await act(async () => {
			result = await hook.current.mutateAsync('  sk-octopus-test-123  ');
		});

		expect(result).toEqual(successStatus);
		expect(fetchMock).toHaveBeenCalledTimes(1);
		expect(fetchMock).toHaveBeenCalledWith(`${API_BASE_URL}api/v1/apikey/login`, expect.objectContaining({
			method: 'GET',
			headers: expect.objectContaining({
				Authorization: 'Bearer sk-octopus-test-123',
			}),
		}));

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

		const { result } = renderHook(() => useAPIKeyLogin(), { wrapper: createWrapper() });
		let mutationError: unknown;
		await act(async () => {
			try {
				await result.current.mutateAsync('sk-octopus-disabled-123');
			} catch (error) {
				mutationError = error;
			}
		});
		expect(mutationError).toBeInstanceOf(Error);
		expect((mutationError as Error).message).toBe('API key is disabled');

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

		await act(async () => {
			await useAuthStore.getState().checkAuth();
		});

		expect(fetchMock).toHaveBeenCalledTimes(1);
		const firstCall = fetchMock.mock.calls.at(0);
		expect(firstCall).toBeTruthy();
		expect(firstCall?.[0]).toBe(`${API_BASE_URL}api/v1/apikey/login`);
		const init = firstCall?.[1] as RequestInit | undefined;
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
