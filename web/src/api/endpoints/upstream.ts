import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../client';
import type {
    NewAPITokenUsage,
    UpstreamAuthMode,
    UpstreamGatewayKey,
    UpstreamGroup,
    UpstreamInspectRequest,
    UpstreamInspectResult,
    UpstreamProviderType,
    UpstreamSubscription,
} from './channel';
import type { UpstreamModelPrice } from './model';

export type UpstreamCheckinLogEntry = {
    success: boolean;
    amount?: number;
    message?: string;
    at: string;
};

export type UpstreamSite = {
    id: number;
    name: string;
    provider_type: UpstreamProviderType;
    base_url: string;
    api_base_url: string;
    auth_mode: UpstreamAuthMode;
    enabled: boolean;
    auto_refresh: boolean;
    refresh_interval_secs: number;
    auto_checkin: boolean;
    checkin_interval_secs: number;
    last_checkin_at?: string;
    checkin_log?: string;
    sync_to_channel: boolean;
    linked_channel_id?: number;
    last_refresh_at?: string;
    last_refresh_status?: string;
    last_refresh_message?: string;
    model_count: number;
    key_count: number;
    group_count: number;
    price_count: number;
    subscription_count: number;
    balance_available: boolean;
    balance_used?: number;
    balance_remain?: number;
    balance_unlimited?: boolean;
    balance_alert_threshold?: number;
    last_balance_check_at?: string;
    last_balance_value?: number;
    auto_create_key?: boolean;
    key_quota_limit?: number;
    key_expire_days?: number;
    auto_sync_group?: boolean;
    auto_sync_price?: boolean;
    source_label?: string;
};

export type UpstreamCredential = {
    id: number;
    upstream_site_id: number;
    credential_type: string;
    auth_mode: UpstreamAuthMode;
    display_name: string;
    masked_value: string;
    user_id?: string;
    importable: boolean;
    last_validated_at?: string;
};

export type UpstreamKeySnapshot = {
    id: number;
    upstream_site_id: number;
    name?: string;
    masked_key?: string;
    allowed_models?: string;
    request_capabilities?: string;
    groups?: string;
    status?: string;
    quota?: number;
    quota_used?: number;
    expires_at?: string;
    importable: boolean;
    source_type?: string;
    channel_key_id?: number;
};

export type UpstreamGroupSnapshot = {
    id: number;
    upstream_site_id: number;
    external_id?: string;
    name: string;
    description?: string;
    platform?: string;
    status?: string;
    rate_multiplier?: number;
    models?: string;
    request_capabilities?: string;
    source?: string;
};

export type UpstreamSiteDetail = {
    site: UpstreamSite;
    credentials: UpstreamCredential[];
    keys: UpstreamKeySnapshot[];
    groups: UpstreamGroupSnapshot[];
    prices: UpstreamModelPrice[];
    subscriptions?: UpstreamSubscription[];
    linked_channel?: {
        id: number;
        name: string;
        type: number;
        enabled: boolean;
        key_count: number;
    };
};

export type UpstreamSiteCreateRequest = {
    name?: string;
    provider_type: UpstreamProviderType;
    base_url: string;
    auth_mode: UpstreamAuthMode;
    token?: string;
    access_key?: string;
    user_id?: string;
    username?: string;
    password?: string;
    auto_refresh?: boolean;
    refresh_interval_secs?: number;
    auto_checkin?: boolean;
    checkin_interval_secs?: number;
    sync_to_channel?: boolean;
    auto_sync_group?: boolean;
    auto_sync_price?: boolean;
    auto_create_key?: boolean;
    key_quota_limit?: number;
    key_expire_days?: number;
    channel_name?: string;
    target_channel_id?: number;
};

export type UpstreamSiteUpdateRequest = {
    id: number;
    name?: string;
    enabled?: boolean;
    auto_refresh?: boolean;
    refresh_interval_secs?: number;
    auto_checkin?: boolean;
    checkin_interval_secs?: number;
    sync_to_channel?: boolean;
    auto_sync_group?: boolean;
    auto_sync_price?: boolean;
    linked_channel_id?: number;
    balance_alert_threshold?: number;
    auto_create_key?: boolean;
    key_quota_limit?: number;
    key_expire_days?: number;
};

export type UpstreamRefreshRequest = {
    id: number;
    apply_channel?: boolean;
};

export type UpstreamApplyRequest = {
    id: number;
    target_channel_id?: number;
};

export type UpstreamCreateKeyRequest = {
    site_id: number;
    name: string;
    quota?: number;
    expires_at?: string;
    models?: string[];
    groups?: string[];
};

export type UpstreamCheckinResult = {
    success: boolean;
    amount?: number;
    message?: string;
    at: string;
};

export type UpstreamCreateKeyResult = {
    name: string;
    key: string;
    masked_key: string;
};

export type UpstreamHealthStatus = 'healthy' | 'degraded' | 'unhealthy';

export type UpstreamHealthItem = {
    id: number;
    name: string;
    status: UpstreamHealthStatus;
    enabled: boolean;
    last_refresh_at?: string;
    last_refresh_status?: string;
    model_count: number;
    balance_available: boolean;
    balance_remain: number;
    balance_unlimited: boolean;
    balance_alert: boolean;
    error_rate: number;
    suppressed: boolean;
    reasons?: string[];
};

export type UpstreamUsagePoint = {
    date: string;
    timestamp: number;
    request_count: number;
    input_tokens: number;
    output_tokens: number;
    total_tokens: number;
    cost: number;
    success_count: number;
    failure_count: number;
};

export type UpstreamUsageResponse = {
    site_id: number;
    days: number;
    points: UpstreamUsagePoint[];
    channel_ids?: number[];
};

export type UpstreamInspectPreview = UpstreamInspectResult & {
    keys?: UpstreamGatewayKey[];
    groups?: UpstreamGroup[];
    token_usage: NewAPITokenUsage;
};

export function useUpstreamSiteList() {
    return useQuery({
        queryKey: ['upstream', 'list'],
        queryFn: async () => apiClient.get<UpstreamSite[]>('/api/v1/upstream/list'),
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

export function useUpstreamSiteDetail(id?: number) {
    return useQuery({
        queryKey: ['upstream', 'detail', id ?? 'none'],
        queryFn: async () => apiClient.get<UpstreamSiteDetail>(`/api/v1/upstream/detail/${id}`),
        enabled: Boolean(id),
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

export function useInspectUpstreamSite() {
    return useMutation({
        mutationFn: async (data: UpstreamInspectRequest) => apiClient.post<UpstreamInspectPreview>('/api/v1/upstream/inspect', data),
    });
}

export function useCreateUpstreamSite() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: async (data: UpstreamSiteCreateRequest) => apiClient.post<UpstreamSiteDetail>('/api/v1/upstream/create', data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['upstream'] });
            queryClient.invalidateQueries({ queryKey: ['channels', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models'] });
        },
    });
}

export function useUpdateUpstreamSite() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: async (data: UpstreamSiteUpdateRequest) => apiClient.post<UpstreamSite>('/api/v1/upstream/update', data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['upstream'] });
        },
    });
}

export function useRefreshUpstreamSite() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: async (data: UpstreamRefreshRequest) => apiClient.post<UpstreamSiteDetail>('/api/v1/upstream/refresh', data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['upstream'] });
            queryClient.invalidateQueries({ queryKey: ['channels', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models'] });
        },
    });
}

export function useCheckinUpstreamSite() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: async (id: number) => apiClient.post<UpstreamCheckinResult>('/api/v1/upstream/checkin', { id }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['upstream'] });
        },
    });
}

export function useApplyUpstreamSite() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: async (data: UpstreamApplyRequest) => apiClient.post('/api/v1/upstream/apply', data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['upstream'] });
            queryClient.invalidateQueries({ queryKey: ['channels', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models'] });
        },
    });
}

export function useDeleteUpstreamSite() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: async (id: number) => apiClient.delete<null>(`/api/v1/upstream/delete/${id}`),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['upstream'] });
            queryClient.invalidateQueries({ queryKey: ['models'] });
        },
    });
}

export function useCreateUpstreamKey() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: async (data: UpstreamCreateKeyRequest) => apiClient.post<UpstreamCreateKeyResult>('/api/v1/upstream/create-key', data),
        onSuccess: (_data, variables) => {
            queryClient.invalidateQueries({ queryKey: ['upstream', 'detail', String(variables.site_id)] });
            queryClient.invalidateQueries({ queryKey: ['upstream', 'list'] });
        },
    });
}

export function useUpstreamSiteHealth() {
    return useQuery({
        queryKey: ['upstream', 'health'],
        queryFn: async () => apiClient.get<UpstreamHealthItem[]>('/api/v1/upstream/health'),
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

export function useUpstreamSiteUsage(id?: number, days = 7) {
    return useQuery({
        queryKey: ['upstream', 'usage', id ?? 'none', days],
        queryFn: async () => apiClient.get<UpstreamUsageResponse>(`/api/v1/upstream/usage/${id}?days=${days}`),
        enabled: Boolean(id),
        refetchInterval: 60000,
        refetchOnMount: 'always',
    });
}

export function useRestoreUpstreamPriority() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: async (id: number) => apiClient.post<null>(`/api/v1/upstream/restore-priority/${id}`),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['upstream'] });
        },
    });
}
