import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../client';
import { logger } from '@/lib/logger';
import { formatCount, formatMoney, formatTime } from '@/lib/utils';
import { StatsChannel, type StatsMetricsFormatted } from './stats';
import type { BillingMode, ProbePolicy } from './model';
/**
 * 渠道类型枚举
 */
export enum ChannelType {
    OpenAIChat = 0,
    OpenAIResponse = 1,
    Anthropic = 2,
    Gemini = 3,
    Volcengine = 4,
    OpenAIEmbedding = 5,
    GithubCopilot = 6,
    Antigravity = 7,
    Zen = 8,
}

/**
 * 自动分组类型枚举
 */
export enum AutoGroupType {
    None = 0,   // 不自动分组
    Fuzzy = 1,  // 模糊匹配
    Exact = 2,  // 准确匹配
    Regex = 3,  // 正则匹配
}

export type BaseUrl = {
    url: string;
    delay: number;
};

export type CustomHeader = {
    header_key: string;
    header_value: string;
};

export type ChannelKey = {
    id: number;
    channel_id: number;
    enabled: boolean;
    channel_key: string;
    source_type?: ChannelKeySourceType | string;
    status_code: number;
    last_use_time_stamp: number;
    total_cost: number;
    remark: string;
    allowed_models?: string;
    request_capabilities?: string;
    upstream_site_id?: number;
    upstream_key_name?: string;
};

export type KeyManagementMode = 'classified' | 'pooled';
export type KeyRoutingPolicy = 'round_robin' | 'fill_priority' | 'priority_order';
export type ChannelKeySourceType = 'unknown' | 'public/free' | 'paid/metered' | 'private/internal';

export type RouteTargetOverride = {
    id: number;
    channel_id: number;
    channel_key_id: number;
    model_name: string;
    billing_mode: BillingMode;
    probe_policy: ProbePolicy;
    probe_interval_seconds: number;
    probe_concurrency_limit: number;
};

export type RouteTargetOverrideUpsertRequest = {
    channel_id: number;
    channel_key_id: number;
    model_name: string;
    billing_mode: BillingMode;
    probe_policy: ProbePolicy;
    probe_interval_seconds: number;
    probe_concurrency_limit: number;
};

export type RouteTargetOverrideDeleteRequest = {
    channel_id: number;
    channel_key_id: number;
    model_name: string;
};

export const CHANNEL_KEY_SOURCE_TYPES: ChannelKeySourceType[] = [
    'unknown',
    'public/free',
    'paid/metered',
    'private/internal',
];

export function normalizeKeyManagementMode(value?: string | null): KeyManagementMode {
    const normalized = (value ?? '').trim().toLowerCase();
    return normalized === 'classified' ? 'classified' : 'pooled';
}

export function normalizeKeyRoutingPolicy(value?: string | null): KeyRoutingPolicy {
    const normalized = (value ?? '').trim().toLowerCase();
    switch (normalized) {
        case 'fill_priority':
            return 'fill_priority';
        case 'priority_order':
            return 'priority_order';
        default:
            return 'round_robin';
    }
}

/**
 * 渠道完整数据（与后端 model.Channel 对齐；数组字段在前端保证为 []）
 */
export type Channel = {
    id: number;
    name: string;
    type: ChannelType;
    enabled: boolean;
    key_management_mode?: KeyManagementMode;
    key_routing_policy?: KeyRoutingPolicy;
    base_urls: BaseUrl[];
    keys: ChannelKey[];
    model: string;
    custom_model: string;
    proxy: boolean;
    auto_sync: boolean;
    auto_group: AutoGroupType;
    custom_header: CustomHeader[];
    param_override?: string | null;
    channel_proxy?: string | null;
    match_regex?: string | null;
    upstream_site_id?: number;
    upstream_source?: string;
    stats: StatsChannel;
};

// Internal type: backend may return null for slice fields; normalize to [] in select()
type ChannelServer = Omit<Channel, 'base_urls' | 'custom_header' | 'keys'> & {
    base_urls: BaseUrl[] | null;
    custom_header: CustomHeader[] | null;
    keys: ChannelKey[] | null;
};

/**
 * 创建渠道请求：必填字段 + 可选字段
 */
export type CreateChannelRequest = {
    name: string;
    type: ChannelType;
    enabled?: boolean;
    key_management_mode?: KeyManagementMode;
    key_routing_policy?: KeyRoutingPolicy;
    base_urls: BaseUrl[];
    keys: Array<Pick<ChannelKey, 'enabled' | 'channel_key' | 'source_type' | 'remark' | 'allowed_models' | 'request_capabilities'>>;
    model: string;
    custom_model?: string;
    proxy?: boolean;
    auto_sync?: boolean;
    auto_group?: AutoGroupType;
    custom_header?: CustomHeader[];
    channel_proxy?: string | null;
    param_override?: string | null;
    match_regex?: string | null;
    upstream_site_id?: number;
    upstream_source?: string;
};

/**
 * 更新渠道请求：id + 可选字段 + keys diff
 */
export type UpdateChannelRequest = {
    id: number;
    name?: string;
    type?: ChannelType;
    enabled?: boolean;
    key_management_mode?: KeyManagementMode;
    key_routing_policy?: KeyRoutingPolicy;
    base_urls?: BaseUrl[];
    model?: string;
    custom_model?: string;
    proxy?: boolean;
    auto_sync?: boolean;
    auto_group?: AutoGroupType;
    custom_header?: CustomHeader[];
    channel_proxy?: string | null;
    param_override?: string | null;
    match_regex?: string | null;
    upstream_site_id?: number;
    upstream_source?: string;
    // keys diff
    keys_to_add?: Array<Pick<ChannelKey, 'enabled' | 'channel_key' | 'source_type' | 'remark' | 'allowed_models' | 'request_capabilities' | 'upstream_site_id' | 'upstream_key_name'>>;
    keys_to_update?: Array<{ id: number; enabled?: boolean; channel_key?: string; source_type?: string; remark?: string; allowed_models?: string; request_capabilities?: string; upstream_site_id?: number; upstream_key_name?: string }>;
    keys_to_delete?: number[];
};

export type FetchModelRequest = {
    type: ChannelType;
    base_url: string;
    key: string;
    proxy?: boolean;
    channel_proxy?: string | null;
    match_regex?: string | null;
    custom_header?: CustomHeader[];
};

export type NewAPITokenUsage = {
    available: boolean;
    used_quota?: number;
    remain_quota?: number;
    unlimited?: boolean;
    raw_status_text?: string;
};

export type NewAPIInspectResult = {
    base_url: string;
    api_base_url: string;
    model_count: number;
    models: string[];
    request_capabilities: string[];
    token_usage: NewAPITokenUsage;
    warnings?: string[];
};

export type UpstreamProviderType = 'newapi' | 'sub2api' | 'openai_compatible';
export type UpstreamAuthMode = 'token' | 'access_key' | 'account_password';

export type UpstreamGatewayKey = {
    name?: string;
    key?: string;
    masked_key?: string;
    allowed_models?: string[];
    request_capabilities?: string[];
    groups?: string[];
    status?: string;
    quota?: number;
    quota_used?: number;
    expires_at?: string;
    importable: boolean;
    source_type?: string;
};

export type UpstreamSubscription = {
    name?: string;
    plan?: string;
    status?: string;
    balance?: number;
    expires_at?: string;
    source?: string;
};

export type UpstreamPriceCandidate = {
    name: string;
    canonical_name?: string;
    cache_supported?: boolean;
    cache_policy?: 'supported' | 'unsupported' | 'unknown';
    cache_reason?: string;
    price_source?: string;
    price_matched_key?: string;
    sources?: string[];
    input?: number;
    output?: number;
    cache_read?: number;
    cache_write?: number;
    official_input?: number;
    official_output?: number;
    official_cache_read?: number;
    official_cache_write?: number;
};

export type UpstreamInspectRequest = {
    provider_type: UpstreamProviderType;
    base_url: string;
    auth_mode: UpstreamAuthMode;
    token?: string;
    access_key?: string;
    user_id?: string;
    username?: string;
    password?: string;
};

export type UpstreamGroup = {
    id?: string;
    name: string;
    description?: string;
    platform?: string;
    status?: string;
    rate_multiplier?: number;
    models?: string[];
    request_capabilities?: string[];
    source?: string;
};

export type UpstreamInspectResult = {
    provider_type: UpstreamProviderType;
    auth_mode: UpstreamAuthMode;
    base_url: string;
    api_base_url: string;
    official_source: boolean;
    source_label?: string;
    model_count: number;
    models: string[];
    request_capabilities: string[];
    token_usage: NewAPITokenUsage;
    keys?: UpstreamGatewayKey[];
    groups?: UpstreamGroup[];
    subscriptions?: UpstreamSubscription[];
    price_candidates?: UpstreamPriceCandidate[];
    warnings?: string[];
};

export type UpstreamApplyRequest = {
    inspect: UpstreamInspectRequest;
    target_channel_id?: number;
    upstream_site_id?: number;
    channel_name?: string;
    append_keys?: boolean;
    overwrite_models?: boolean;
    enable_channel?: boolean;
};

export type UpstreamApplyResult = {
    channel: {
        id: number;
        name: string;
        type: ChannelType;
        enabled: boolean;
        base_urls?: BaseUrl[];
        custom_model?: string;
        key_count: number;
    };
    inspect: UpstreamInspectResult;
    created: boolean;
};

/**
 * 获取渠道列表 Hook
 * 
 * @example
 * const { data: channels, isLoading, error } = useChannelList();
 * 
 * if (isLoading) return <Loading />;
 * if (error) return <Error message={error.message} />;
 * 
 * channels?.forEach(channel => console.log(channel.raw.name));
 */
export function useChannelList() {
    return useQuery({
        queryKey: ['channels', 'list'],
        queryFn: async () => {
            return apiClient.get<ChannelServer[]>('/api/v1/channel/list');
        },
        select: (data) => data.map((item) => ({
            raw: ({
                ...item,
                key_management_mode: normalizeKeyManagementMode(item.key_management_mode),
                key_routing_policy: normalizeKeyRoutingPolicy(item.key_routing_policy),
                base_urls: item.base_urls ?? [],
                custom_header: item.custom_header ?? [],
                keys: item.keys ?? [],
            }) satisfies Channel,
            formatted: {
                input_token: formatCount(item.stats.input_token),
                output_token: formatCount(item.stats.output_token),
                total_token: formatCount(item.stats.input_token + item.stats.output_token),
                input_cost: formatMoney(item.stats.input_cost),
                output_cost: formatMoney(item.stats.output_cost),
                total_cost: formatMoney(item.stats.input_cost + item.stats.output_cost),
                request_success: formatCount(item.stats.request_success),
                request_failed: formatCount(item.stats.request_failed),
                request_count: formatCount(item.stats.request_success + item.stats.request_failed),
                wait_time: formatTime(item.stats.wait_time),
            }
        })) as Array<{ raw: Channel; formatted: StatsMetricsFormatted }>,
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

/**
 * 创建渠道 Hook
 * 
 * @example
 * const createChannel = useCreateChannel();
 * 
 * createChannel.mutate({
 *   name: 'OpenAI',
 *   type: ChannelType.OpenAIChat,
 *   base_urls: [{ url: 'https://api.openai.com', delay: 0 }],
 *   keys: [{ enabled: true, channel_key: 'sk-xxx' }],
 *   model: 'gpt-4',
 * });
 */
export function useCreateChannel() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: CreateChannelRequest) => {
            return apiClient.post<ChannelServer>('/api/v1/channel/create', data);
        },
        onSuccess: (data) => {
            logger.log('渠道创建成功:', data);
            queryClient.invalidateQueries({ queryKey: ['channels', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'channel'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
        },
        onError: (error) => {
            logger.error('渠道创建失败:', error);
        },
    });
}

/**
 * 更新渠道 Hook
 * 
 * @example
 * const updateChannel = useUpdateChannel();
 * 
 * updateChannel.mutate({
 *   id: 1,
 *   name: 'OpenAI Updated',
 *   type: ChannelType.OpenAIChat,
 *   enabled: true,
 *   base_urls: [{ url: 'https://api.openai.com', delay: 0 }],
 *   keys_to_add: [{ enabled: true, channel_key: 'sk-xxx' }],
 *   model: 'gpt-4-turbo',
 *   proxy: false,
 * });
 */
export function useUpdateChannel() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: UpdateChannelRequest) => {
            return apiClient.post<ChannelServer>('/api/v1/channel/update', data);
        },
        onSuccess: (data) => {
            logger.log('渠道更新成功:', data);
            queryClient.invalidateQueries({ queryKey: ['channels', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'channel'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
        },
        onError: (error) => {
            logger.error('渠道更新失败:', error);
        },
    });
}

/**
 * 删除渠道 Hook
 * 
 * @example
 * const deleteChannel = useDeleteChannel();
 * 
 * deleteChannel.mutate(1); // 删除 ID 为 1 的渠道
 */
export function useDeleteChannel() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (id: number) => {
            return apiClient.delete<null>(`/api/v1/channel/delete/${id}`);
        },
        onSuccess: () => {
            logger.log('渠道删除成功');
            queryClient.invalidateQueries({ queryKey: ['channels', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'channel'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
        },
        onError: (error) => {
            logger.error('渠道删除失败:', error);
        },
    });
}

/**
 * 启用/禁用渠道 Hook
 * 
 * @example
 * const enableChannel = useEnableChannel();
 * 
 * enableChannel.mutate({ id: 1, enabled: true }); // 启用 ID 为 1 的渠道
 * enableChannel.mutate({ id: 1, enabled: false }); // 禁用 ID 为 1 的渠道
 */
export function useEnableChannel() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: { id: number; enabled: boolean }) => {
            return apiClient.post<null>('/api/v1/channel/enable', data);
        },
        onSuccess: () => {
            logger.log('渠道状态更新成功');
            queryClient.invalidateQueries({ queryKey: ['channels', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'channel'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
        },
        onError: (error) => {
            logger.error('渠道状态更新失败:', error);
        },
    });
}

export type BatchOperationResult = {
	success_count: number;
	failed_count: number;
	errors: string[];
};

export function useBatchEnableChannel() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: async (ids: number[]) => {
			return apiClient.post<BatchOperationResult>('/api/v1/channel/batch-enable', { ids });
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['channels', 'list'] });
		},
	});
}

export function useBatchDisableChannel() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: async (ids: number[]) => {
			return apiClient.post<BatchOperationResult>('/api/v1/channel/batch-disable', { ids });
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['channels', 'list'] });
		},
	});
}

export function useBatchDeleteChannel() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: async (ids: number[]) => {
			return apiClient.post<BatchOperationResult>('/api/v1/channel/batch-delete', { ids });
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['channels', 'list'] });
			queryClient.invalidateQueries({ queryKey: ['models', 'channel'] });
			queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
		},
	});
}

export function useRouteTargetOverrideList(channelId?: number) {
    return useQuery({
        queryKey: ['route-target', 'list', channelId ?? 'all'],
        queryFn: async () => {
            const suffix = typeof channelId === 'number' && channelId > 0 ? `?channel_id=${channelId}` : '';
            return apiClient.get<RouteTargetOverride[]>(`/api/v1/route-target/list${suffix}`);
        },
        enabled: channelId === undefined || channelId > 0,
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

export function useUpsertRouteTargetOverride() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: RouteTargetOverrideUpsertRequest) => {
            return apiClient.post<RouteTargetOverride>('/api/v1/route-target/upsert', data);
        },
        onSuccess: (data) => {
            logger.log('route target override upsert success:', data);
            queryClient.invalidateQueries({ queryKey: ['route-target', 'list'] });
        },
        onError: (error) => {
            logger.error('route target override upsert failed:', error);
        },
    });
}

export function useDeleteRouteTargetOverride() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: RouteTargetOverrideDeleteRequest) => {
            return apiClient.post<null>('/api/v1/route-target/delete', data);
        },
        onSuccess: () => {
            logger.log('route target override delete success');
            queryClient.invalidateQueries({ queryKey: ['route-target', 'list'] });
        },
        onError: (error) => {
            logger.error('route target override delete failed:', error);
        },
    });
}

/**
 * 获取渠道模型列表 Hook
 * 
 * @example
 * const fetchModel = useFetchModel();
 * 
 * fetchModel.mutate({
 *   type: ChannelType.OpenAIChat,
 *   base_url: 'https://api.openai.com',
 *   key: 'sk-xxx',
 *   proxy: false,
 * });
 * 
 * // 在 onSuccess 中获取模型列表
 * fetchModel.data // ['gpt-4', 'gpt-3.5-turbo', ...]
 */
export function useFetchModel() {
    return useMutation({
        mutationFn: async (data: FetchModelRequest) => {
            return apiClient.post<string[]>('/api/v1/channel/fetch-model', data);
        },
        onSuccess: (data) => {
            logger.log('模型列表获取成功:', data);
        },
        onError: (error) => {
            logger.error('模型列表获取失败:', error);
        },
    });
}

export function useInspectNewAPI() {
    return useMutation({
        mutationFn: async (data: { base_url: string; token: string }) => {
            return apiClient.post<NewAPIInspectResult>('/api/v1/channel/newapi/inspect', data);
        },
        onSuccess: (data) => {
            logger.log('New API 检查完成:', data);
        },
        onError: (error) => {
            logger.error('New API 检查失败:', error);
        },
    });
}

export function useInspectUpstreamGateway() {
    return useMutation({
        mutationFn: async (data: UpstreamInspectRequest) => {
            return apiClient.post<UpstreamInspectResult>('/api/v1/channel/upstream/inspect', data);
        },
        onSuccess: (data) => {
            logger.log('上游站点检查完成:', data);
        },
        onError: (error) => {
            logger.error('上游站点检查失败:', error);
        },
    });
}

export function useApplyUpstreamGateway() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: async (data: UpstreamApplyRequest) => {
            return apiClient.post<UpstreamApplyResult>('/api/v1/channel/upstream/apply', data);
        },
        onSuccess: (data) => {
            logger.log('上游站点已应用到渠道:', data);
            queryClient.invalidateQueries({ queryKey: ['channels', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'channel'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
        },
        onError: (error) => {
            logger.error('上游站点应用失败:', error);
        },
    });
}

/**
 * 获取渠道最后同步时间 Hook
 * 
 * @example
 * const lastSyncTime = useLastSyncTime();
 * 
 * if (lastSyncTime) {
 *   console.log('最后同步时间:', new Date(lastSyncTime).toLocaleString());
 * }
 */
export function useLastSyncTime() {
    return useQuery({
        queryKey: ['channels', 'last-sync-time'],
        queryFn: async () => {
            return apiClient.get<string>('/api/v1/channel/last-sync-time');
        },
        refetchInterval: 30000,
    });
}
/**
 * 同步渠道 Hook
 *
 * @example
 * const syncChannel = useSyncChannel();
 *
 * syncChannel.mutate();
 */
export function useSyncChannel() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: async () => {
            return apiClient.post<null>('/api/v1/channel/sync', {});
        },
        onSuccess: () => {
            logger.log('渠道同步成功');
            queryClient.invalidateQueries({ queryKey: ['channels', 'last-sync-time'] });
            queryClient.invalidateQueries({ queryKey: ['channels', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'channel'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
        },
        onError: (error) => {
            logger.error('渠道同步失败:', error);
        },
    });
}

/**
 * 测试渠道模型 Hook
 *
 * @example
 * const testModels = useTestChannelModels();
 *
 * testModels.mutate({
 *   channel_id: 1,
 *   models: ['gpt-4', 'gpt-3.5-turbo'],
 * });
 */
export type TestModelRequest = {
    channel_id: number;
    models: string[];
};

export type TestModelResult = {
    model: string;
    source_type?: string;
    billing_mode?: BillingMode;
    probe_policy?: ProbePolicy;
    policy_basis?: string;
    passed: boolean;
    error?: string;
    delay?: number;
};

export function useTestChannelModels() {
    return useMutation({
        mutationFn: async (data: TestModelRequest) => {
            return apiClient.post<TestModelResult[]>('/api/v1/channel/test-models', data);
        },
        onSuccess: (data) => {
            logger.log('模型测试完成:', data);
        },
        onError: (error) => {
            logger.error('模型测试失败:', error);
        },
    });
}

export type TestModelByConfigRequest = {
    type: ChannelType;
    enabled?: boolean;
    base_urls: BaseUrl[];
    keys: Array<Pick<ChannelKey, 'enabled' | 'channel_key' | 'source_type' | 'allowed_models' | 'request_capabilities'>>;
    proxy?: boolean;
    channel_proxy?: string | null;
    custom_header?: Array<CustomHeader>;
    key_management_mode?: KeyManagementMode;
    key_routing_policy?: KeyRoutingPolicy;
    models: string[];
};

export function useTestChannelModelsByConfig() {
    return useMutation({
        mutationFn: async (data: TestModelByConfigRequest) => {
            return apiClient.post<TestModelResult[]>('/api/v1/channel/test-models-by-config', data);
        },
        onSuccess: (data) => {
            logger.log('模型(配置)测试完成:', data);
        },
        onError: (error) => {
            logger.error('模型(配置)测试失败:', error);
        },
    });
}

// ---- GitHub Copilot Device Flow ----

export type CopilotDeviceCodeResponse = {
    device_code: string;
    user_code: string;
    verification_uri: string;
    expires_in: number;
    interval: number;
};

export type CopilotPollResponse = {
    access_token?: string;
    token_type?: string;
    scope?: string;
    error?: string;
};

export function useCopilotRequestDeviceCode() {
    return useMutation({
        mutationFn: async () => {
            return apiClient.post<CopilotDeviceCodeResponse>('/api/v1/channel/copilot/device-code', {});
        },
    });
}

export function useCopilotPollToken() {
    return useMutation({
        mutationFn: async (deviceCode: string) => {
            return apiClient.post<CopilotPollResponse>('/api/v1/channel/copilot/poll-token', { device_code: deviceCode });
        },
    });
}

export type AntigravityOAuthStartResponse = {
    state: string;
    auth_url: string;
};

export type AntigravityOAuthPollResponse = {
    status: 'pending' | 'authorized' | 'failed';
    access_token?: string;
    token_type?: string;
    scope?: string;
    error?: string;
};

export function useAntigravityOAuthStart() {
    return useMutation({
        mutationFn: async () => {
            return apiClient.post<AntigravityOAuthStartResponse>('/api/v1/channel/antigravity/oauth/start', {});
        },
    });
}

export function useAntigravityOAuthPoll() {
    return useMutation({
        mutationFn: async (state: string) => {
            return apiClient.post<AntigravityOAuthPollResponse>('/api/v1/channel/antigravity/oauth/poll', { state });
        },
    });
}
