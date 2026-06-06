import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../client';
import { logger } from '@/lib/logger';

/**
 * LLM 价格信息
 */
export interface LLMPrice {
    input: number;
    output: number;
    cache_read: number;
    cache_write: number;
}

export interface OfficialLLMPrice {
    official_input?: number;
    official_output?: number;
    official_cache_read?: number;
    official_cache_write?: number;
}

export type BillingMode = 'unknown' | 'per_request' | 'per_token' | 'per_quota' | 'flat' | 'free';
export type ProbePolicy = 'passive_only' | 'sparse_single' | 'sequential' | 'concurrent';
export type CachePolicy = 'unknown' | 'supported' | 'unsupported';

export const BILLING_MODE_OPTIONS: BillingMode[] = ['unknown', 'per_request', 'per_token', 'per_quota', 'flat', 'free'];
export const PROBE_POLICY_OPTIONS: ProbePolicy[] = ['passive_only', 'sparse_single', 'sequential', 'concurrent'];

/**
 * LLM 模型信息
 */
export interface LLMInfo extends LLMPrice, OfficialLLMPrice {
    name: string;
    canonical_name?: string;
    billing_mode?: BillingMode;
    probe_policy?: ProbePolicy;
    probe_interval_seconds?: number;
    probe_concurrency_limit?: number;
    cache_policy?: CachePolicy;
    cache_reason?: string;
    cache_supported?: boolean;
    upstream_provider_type?: string;
    upstream_source?: string;
}

/**
 * LLM 渠道关联信息
 */
export interface LLMChannel {
    name: string;
    enabled: boolean;
    channel_id: number;
    channel_name: string;
    key_count?: number;
    request_capabilities?: string[];
    inventory_source?: string;
}

export interface ServiceableModelInventoryItem extends LLMChannel {
    key_count: number;
    inventory_source: string;
}

export interface SelectableGroupModelInventoryItem {
    name: string;
    channel_count: number;
    enabled_channel_count: number;
    key_count: number;
    request_capabilities?: string[];
    inventory_source: string;
}

export interface RoutableModelInventoryItem {
    name: string;
    group_id: number;
    group_name: string;
    channel_count: number;
    enabled_channel_count: number;
    key_count: number;
    request_capabilities?: string[];
    inventory_source: string;
}

export interface CapabilityInventory {
    serviceable_models: ServiceableModelInventoryItem[];
    selectable_models: SelectableGroupModelInventoryItem[];
    routable_models?: RoutableModelInventoryItem[];
}

/**
 * 获取 LLM 模型列表 Hook
 * 
 * @example
 * const { data: models, isLoading, error } = useModelList();
 * 
 * if (isLoading) return <Loading />;
 * if (error) return <Error message={error.message} />;
 * 
 * models?.forEach(model => console.log(model.name, model.input));
 */
export function useModelList(options?: { enabled?: boolean }) {
    return useQuery({
        queryKey: ['models', 'list'],
        queryFn: async () => {
            return apiClient.get<LLMInfo[]>('/api/v1/model/list');
        },
        enabled: options?.enabled ?? true,
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

/**
 * 获取 LLM 模型与渠道关联列表 Hook
 * 
 * @example
 * const { data: channelModels, isLoading, error } = useModelChannelList();
 * 
 * if (isLoading) return <Loading />;
 * if (error) return <Error message={error.message} />;
 * 
 * channelModels?.forEach(item => console.log(item.name, item.channel_name));
 */
export function useModelChannelList() {
    return useQuery({
        queryKey: ['models', 'channel'],
        queryFn: async () => {
            return apiClient.get<LLMChannel[]>('/api/v1/model/channel');
        },
        refetchInterval: 30000,
    });
}

export function useCapabilityInventory() {
    return useQuery({
        queryKey: ['models', 'capability-inventory'],
        queryFn: async () => {
            return apiClient.get<CapabilityInventory>('/api/v1/model/capability-inventory');
        },
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

/**
 * 更新 LLM 模型 Hook
 * 
 * @example
 * const updateModel = useUpdateModel();
 * 
 * updateModel.mutate({
 *   name: 'gpt-4',
 *   input: 0.03,
 *   output: 0.06,
 *   cache_read: 0.015,
 *   cache_write: 0.03,
 * });
 */
export function useUpdateModel() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: LLMInfo) => {
            return apiClient.post<LLMInfo>('/api/v1/model/update', data);
        },
        onSuccess: (data) => {
            logger.log('模型更新成功:', data);
            queryClient.invalidateQueries({ queryKey: ['models', 'list'] });
        },
        onError: (error) => {
            logger.error('模型更新失败:', error);
        },
    });
}

/**
 * 创建 LLM 模型 Hook
 * 
 * @example
 * const createModel = useCreateModel();
 * 
 * createModel.mutate({
 *   name: 'gpt-4',
 *   input: 0.03,
 *   output: 0.06,
 *   cache_read: 0.015,
 *   cache_write: 0.03,
 * });
 */
export function useCreateModel() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: LLMInfo) => {
            return apiClient.post<LLMInfo>('/api/v1/model/create', data);
        },
        onSuccess: (data) => {
            logger.log('模型创建成功:', data);
            queryClient.invalidateQueries({ queryKey: ['models', 'list'] });
        },
        onError: (error) => {
            logger.error('模型创建失败:', error);
        },
    });
}

/**
 * 删除 LLM 模型 Hook
 * 
 * @example
 * const deleteModel = useDeleteModel();
 * 
 * deleteModel.mutate('gpt-4'); // 删除名称为 'gpt-4' 的模型
 */
export function useDeleteModel() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (name: string) => {
            return apiClient.post<null>('/api/v1/model/delete', { name });
        },
        onSuccess: () => {
            logger.log('模型删除成功');
            queryClient.invalidateQueries({ queryKey: ['models', 'list'] });
        },
        onError: (error) => {
            logger.error('模型删除失败:', error);
        },
    });
}

/**
 * 更新 LLM 模型价格 Hook
 * 
 * @example
 * const updatePrice = useUpdateModelPrice();
 * 
 * updatePrice.mutate(); // 触发价格更新
 */
export function useUpdateModelPrice() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async () => {
            return apiClient.post<null>('/api/v1/model/update-price', {});
        },
        onSuccess: () => {
            logger.log('模型价格更新成功');
            queryClient.invalidateQueries({ queryKey: ['models', 'last-update-time'] });
        },
        onError: (error) => {
            logger.error('模型价格更新失败:', error);
        },
    });
}

/**
 * 获取 LLM 模型价格最后更新时间 Hook
 * 
 * @example
 * const { data: lastUpdateTime } = useLastUpdateTime();
 * 
 * if (lastUpdateTime) {
 *   console.log('最后更新:', new Date(lastUpdateTime).toLocaleString());
 * }
 */
export function useLastUpdateTime() {
    return useQuery({
        queryKey: ['models', 'last-update-time'],
        queryFn: async () => {
            return apiClient.get<string>('/api/v1/model/last-update-time');
        },
        refetchInterval: 30000,
    });
}
