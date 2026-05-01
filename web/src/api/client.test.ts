import { afterEach, describe, expect, it, vi } from 'vitest';

import { API_BASE_URL, apiClient } from './client';

describe('apiClient base url', () => {
	it('defaults to root path instead of a relative path', () => {
		expect(API_BASE_URL).toBe('/');
	});

	it('falls back to persisted auth token when the runtime store getter is unavailable', async () => {
		const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => new Response(JSON.stringify({ code: 200, message: 'success', data: { ok: true } }), {
			status: 200,
			headers: { 'content-type': 'application/json' },
		}));
		vi.stubGlobal('fetch', fetchMock);

		window.localStorage.setItem('auth-storage', JSON.stringify({
			state: {
				token: 'persisted-token',
			},
			version: 0,
		}));

		await apiClient.post('/api/v1/channel/fetch-model', { base_url: 'http://example.com/v1', key: 'sk-test' });

		expect(fetchMock).toHaveBeenCalledTimes(1);
		const init = fetchMock.mock.calls[0]?.[1] as RequestInit | undefined;
		expect(init?.headers).toBeInstanceOf(Headers);
		expect((init?.headers as Headers).get('Authorization')).toBe('Bearer persisted-token');
	});
});

afterEach(() => {
	window.localStorage.clear();
	vi.unstubAllGlobals();
});

