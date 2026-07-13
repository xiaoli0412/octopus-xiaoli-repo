import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient, buildApiUrl } from '../client';
import { logger } from '@/lib/logger';
import { useAuthStore } from './user';
import { StatsAPIKey, StatsAPIKeyFormatted } from './stats';
import { formatCount, formatMoney, formatTime } from '@/lib/utils';

/**
 * API Key 数据
 */
export interface APIKey {
    id: number;
    name: string;
    api_key: string;
    enabled: boolean;
    expire_at?: number; // Unix 时间戳（秒），不传表示永不过期
    max_cost?: number; // 不传表示无限制
    supported_models?: string; // 不传表示支持所有模型
    rate_limit_rpm?: number; // 每分钟请求数上限，0/undefined 表示不限制
    rate_limit_tpm?: number; // 每分钟 token 数上限，0/undefined 表示不限制
    rate_limit_daily?: number; // 每日请求数上限，0/undefined 表示不限制
    allowed_channels?: string; // JSON 数组，允许的渠道 ID 列表，不传表示不限制
    allowed_groups?: string; // JSON 数组，允许的分组 ID 列表，不传表示不限制
    allowed_capabilities?: string; // JSON 数组，允许的能力列表 (chat/embedding/response/message)，不传表示不限制
    allowed_ip_cidrs?: string; // JSON 数组，允许的 IP CIDR 白名单，不传表示不限制
    response_cache_enabled?: boolean; // 是否启用响应缓存（仅非流式请求），不传表示 false
}

/**
 * API Key Stats 响应（包含 stats 和 info）
 */
export interface APIKeyStatsResponse {
    stats: StatsAPIKey;
    info: APIKey;
}

export interface APIKeyStatsResponseFormatted {
    stats: StatsAPIKeyFormatted;
    info: APIKey;
}

export interface APIKeyAuthStatus {
	ok: boolean;
	api_key_id: number;
	name: string;
	enabled: boolean;
	expire_at?: number;
	supported_models?: string;
	auth_mode: 'api_key';
}

export async function loginAPIKeySession(
	apiKey: string,
	persistAuth: (apiKey: string, status: APIKeyAuthStatus) => void,
): Promise<APIKeyAuthStatus> {
	const trimmedKey = apiKey.trim();
	const response = await fetch(buildApiUrl('/api/v1/apikey/login'), {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${trimmedKey}`,
		},
	});
	const contentType = response.headers.get('content-type') || '';
	const isJson = contentType.includes('application/json');
	const payload = isJson ? await response.json() : await response.text();
	if (!response.ok) {
		const message = typeof payload === 'object' && payload && 'message' in payload && typeof payload.message === 'string'
			? payload.message
			: (typeof payload === 'string' ? payload : response.statusText);
		throw new Error(message || 'API key login failed');
	}
	const data = typeof payload === 'object' && payload && 'data' in payload
		? (payload.data as APIKeyAuthStatus)
		: (payload as APIKeyAuthStatus);
	persistAuth(trimmedKey, data);
	return data;
}

/**
 * API Key 登录 Hook（服务端校验成功后再落本地认证状态）
 */
export function useAPIKeyLogin() {
    const { setAPIKeyAuth, logout } = useAuthStore();

    return useMutation({
		mutationFn: async (apiKey: string) => loginAPIKeySession(apiKey, setAPIKeyAuth),
        onError: (error) => {
            logout();
            logger.error('API Key 登录失败:', error);
        },
    });
}

/**
 * 获取当前 API Key 的详细统计数据 Hook（仅 API Key 登录用户使用）
 */
export function useAPIKeyDashboardStats() {
    const { isAPIKeyAuth, isAuthenticated } = useAuthStore();

    return useQuery({
        queryKey: ['apikey', 'dashboard', 'stats'],
        queryFn: () => apiClient.get<APIKeyStatsResponse>('/api/v1/apikey/stats'),
        select: (data): APIKeyStatsResponseFormatted => ({
            stats: {
                api_key_id: data.stats.api_key_id,
                input_token: formatCount(data.stats.input_token),
                output_token: formatCount(data.stats.output_token),
                total_token: formatCount(data.stats.input_token + data.stats.output_token),
                input_cost: formatMoney(data.stats.input_cost),
                output_cost: formatMoney(data.stats.output_cost),
                total_cost: formatMoney(data.stats.input_cost + data.stats.output_cost),
                wait_time: formatTime(data.stats.wait_time),
                request_success: formatCount(data.stats.request_success),
                request_failed: formatCount(data.stats.request_failed),
                request_count: formatCount(data.stats.request_success + data.stats.request_failed),
            },
            info: data.info,
        }),
        enabled: isAPIKeyAuth && isAuthenticated,
        refetchInterval: 30000,
    });
}

/**
 * 创建 API Key 请求
 */
export type CreateAPIKeyRequest = Omit<APIKey, 'id' | 'api_key'> & { enabled?: boolean };

/**
 * 更新 API Key 请求
 */
export type UpdateAPIKeyRequest = Pick<APIKey, 'id'> & CreateAPIKeyRequest;

/**
 * 获取 API Key 列表 Hook
 * 
 * @example
 * const { data: apiKeys, isLoading, error } = useAPIKeyList();
 * 
 * if (isLoading) return <Loading />;
 * if (error) return <Error message={error.message} />;
 * 
 * apiKeys?.forEach(key => console.log(key.name));
 */
export function useAPIKeyList() {
    return useQuery({
        queryKey: ['apikeys', 'list'],
        queryFn: async () => {
            return apiClient.get<APIKey[]>('/api/v1/apikey/list');
        },
        refetchInterval: 30000,
    });
}

/**
 * 创建 API Key Hook
 * 
 * @example
 * const createAPIKey = useCreateAPIKey();
 * 
 * createAPIKey.mutate({
 *   name: 'My API Key',
 * });
 */
export function useCreateAPIKey() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: CreateAPIKeyRequest) => {
            return apiClient.post<APIKey>('/api/v1/apikey/create', data);
        },
        onSuccess: (data) => {
            logger.log('API Key 创建成功:', data);
            queryClient.invalidateQueries({ queryKey: ['apikeys', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
        },
        onError: (error) => {
            logger.error('API Key 创建失败:', error);
        },
    });
}

/**
 * 更新 API Key Hook
 * 
 * @example
 * const updateAPIKey = useUpdateAPIKey();
 * 
 * updateAPIKey.mutate({
 *   id: 1,
 *   name: 'Updated API Key',
 *   enabled: false,
 * });
 */
export function useUpdateAPIKey() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: UpdateAPIKeyRequest) => {
            return apiClient.post<APIKey>('/api/v1/apikey/update', data);
        },
        onSuccess: (data) => {
            logger.log('API Key 更新成功:', data);
            queryClient.invalidateQueries({ queryKey: ['apikeys', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
        },
        onError: (error) => {
            logger.error('API Key 更新失败:', error);
        },
    });
}

/**
 * 删除 API Key Hook
 * 
 * @example
 * const deleteAPIKey = useDeleteAPIKey();
 * 
 * deleteAPIKey.mutate(1); // 删除 ID 为 1 的 API Key
 */
export function useDeleteAPIKey() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (id: number) => {
            return apiClient.delete<null>(`/api/v1/apikey/delete/${id}`);
        },
        onSuccess: () => {
            logger.log('API Key 删除成功');
            queryClient.invalidateQueries({ queryKey: ['apikeys', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
        },
        onError: (error) => {
            logger.error('API Key 删除失败:', error);
        },
    });
}

export type BatchOperationResult = {
	success_count: number;
	failed_count: number;
	errors: string[];
};

export function useBatchEnableAPIKey() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: async (ids: number[]) => {
			return apiClient.post<BatchOperationResult>('/api/v1/apikey/batch-enable', { ids });
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['apikeys', 'list'] });
			queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
		},
	});
}

export function useBatchDisableAPIKey() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: async (ids: number[]) => {
			return apiClient.post<BatchOperationResult>('/api/v1/apikey/batch-disable', { ids });
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['apikeys', 'list'] });
			queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
		},
	});
}

export function useBatchDeleteAPIKey() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: async (ids: number[]) => {
			return apiClient.post<BatchOperationResult>('/api/v1/apikey/batch-delete', { ids });
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['apikeys', 'list'] });
			queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
		},
	});
}

/**
 * 获取当前 API Key 的统计数据 Hook
 * 
 * 此接口使用 API Key 认证，通过 API Key 获取对应的统计数据
 * 
 * @example
 * const { data: stats, isLoading } = useAPIKeyStats();
 */
export function useAPIKeyStats() {
    return useQuery({
        queryKey: ['apikey', 'stats'],
        queryFn: async () => {
            return apiClient.get<StatsAPIKey>('/api/v1/apikey/stats');
        },
        select: (data): StatsAPIKeyFormatted => ({
            api_key_id: data.api_key_id,
            input_token: formatCount(data.input_token),
            output_token: formatCount(data.output_token),
            total_token: formatCount(data.input_token + data.output_token),
            input_cost: formatMoney(data.input_cost),
            output_cost: formatMoney(data.output_cost),
            total_cost: formatMoney(data.input_cost + data.output_cost),
            wait_time: formatTime(data.wait_time),
            request_success: formatCount(data.request_success),
            request_failed: formatCount(data.request_failed),
            request_count: formatCount(data.request_success + data.request_failed),
        }),
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}
