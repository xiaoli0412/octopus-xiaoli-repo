import type { ApiError } from './types';
import { HttpStatus } from './types';
import { logger } from '@/lib/logger';

export const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || '/';
const AUTH_STORAGE_KEY = 'auth-storage';

function joinApiUrl(baseUrl: string, path: string, searchParams?: string): string {
    const normalizedBase = baseUrl.endsWith('/') ? baseUrl.slice(0, -1) : baseUrl;
    const normalizedPath = path.startsWith('/') ? path : `/${path}`;
    const url = `${normalizedBase}${normalizedPath}`;
    return searchParams ? `${url}?${searchParams}` : url;
}

/**
 * 获取认证 Store（延迟导入以避免循环依赖）
 */
let getAuthStore: (() => { token: string | null; logout: () => void }) | null = null;

export function setAuthStoreGetter(getter: () => { token: string | null; logout: () => void }) {
    getAuthStore = getter;
}

function getPersistedAuthToken(): string | null {
    if (typeof window === 'undefined') return null;

    try {
        const raw = window.localStorage.getItem(AUTH_STORAGE_KEY);
        if (!raw) return null;

        const parsed = JSON.parse(raw) as { state?: { token?: unknown } };
        return typeof parsed?.state?.token === 'string' && parsed.state.token.trim()
            ? parsed.state.token
            : null;
    } catch (error) {
        logger.warn('Failed to read persisted auth token:', error);
        return null;
    }
}

function resolveAuthToken(): string | null {
    if (typeof window === 'undefined') return null;

    if (getAuthStore) {
        const store = getAuthStore();
        if (store.token) {
            return store.token;
        }
    }

    return getPersistedAuthToken();
}

/**
 * 全局错误处理
 */
const handleError = (error: ApiError) => {
    logger.error('API Error:', error);

    // 401 未授权，调用 store 的 logout
    if (error.code === HttpStatus.UNAUTHORIZED) {
        if (getAuthStore) {
            const store = getAuthStore();
            store.logout();
        }
    }
};

/**
 * 处理响应
 */
async function handleResponse<T>(response: Response): Promise<T> {
    const contentType = response.headers.get('content-type');
    const isJson = contentType?.includes('application/json');

    let data: unknown;
    if (isJson) {
        data = await response.json();
    } else {
        data = await response.text();
    }

    if (!response.ok) {
        const error: ApiError = {
            code: response.status,
            message: (data && typeof data === 'object' && 'message' in data && typeof data.message === 'string')
                ? data.message
                : (typeof data === 'string' ? data : response.statusText),
        };

        handleError(error);
        throw error;
    }

    // 如果是标准的 ApiResponse 格式，返回 data 字段
    if (data && typeof data === 'object' && 'data' in data) {
        return data.data as T;
    }

    return data as T;
}

/**
 * 发送请求
 */
async function request<T>(
    method: string,
    path: string,
    body?: BodyInit,
    params?: Record<string, string | number | boolean>
): Promise<T> {
    // 构建 URL
    const searchParams = params ? new URLSearchParams(
        Object.entries(params).map(([k, v]) => [k, String(v)])
    ).toString() : '';
    const url = joinApiUrl(API_BASE_URL, path, searchParams);

    // 构建请求头
    const headers = new Headers();

    // 只在有 body 时设置 Content-Type
    if (body) {
        headers.set('Content-Type', 'application/json');
    }

    // 添加 Authorization - 从 zustand store 获取 token
    const authToken = resolveAuthToken();
    if (authToken) {
        headers.set('Authorization', `Bearer ${authToken}`);
    }

    // 发送请求
    const response = await fetch(url.toString(), {
        method,
        headers,
        body,
    });

    return handleResponse<T>(response);
}

/**
 * API 客户端 - 基础 HTTP 方法
 */
export const apiClient = {
    /**
     * GET 请求
     */
    get: <T>(path: string, params?: Record<string, string | number | boolean>): Promise<T> =>
        request<T>('GET', path, undefined, params),

    /**
     * POST 请求
     */
    post: <T>(path: string, data?: unknown, params?: Record<string, string | number | boolean>): Promise<T> =>
        request<T>('POST', path, data ? JSON.stringify(data) : undefined, params),

    /**
     * PUT 请求
     */
    put: <T>(path: string, data?: unknown, params?: Record<string, string | number | boolean>): Promise<T> =>
        request<T>('PUT', path, data ? JSON.stringify(data) : undefined, params),

    /**
     * DELETE 请求
     */
    delete: <T>(path: string, params?: Record<string, string | number | boolean>): Promise<T> =>
        request<T>('DELETE', path, undefined, params),

    /**
     * PATCH 请求
     */
    patch: <T>(path: string, data?: unknown, params?: Record<string, string | number | boolean>): Promise<T> =>
        request<T>('PATCH', path, data ? JSON.stringify(data) : undefined, params),

    /**
     * Raw fetch for file/blob downloads while still reusing auth logic.
     */
    raw: async (path: string, params?: Record<string, string | number | boolean>): Promise<Response> => {
        const searchParams = params ? new URLSearchParams(
            Object.entries(params).map(([k, v]) => [k, String(v)])
        ).toString() : '';
        const url = joinApiUrl(API_BASE_URL, path, searchParams);

        const headers = new Headers();
        const authToken = resolveAuthToken();
        if (authToken) {
            headers.set('Authorization', `Bearer ${authToken}`);
        }

        const response = await fetch(url.toString(), { method: 'GET', headers });
        if (!response.ok) {
            const text = await response.text();
            const error: ApiError = { code: response.status, message: text || response.statusText };
            handleError(error);
            throw error;
        }
        return response;
    },
};

