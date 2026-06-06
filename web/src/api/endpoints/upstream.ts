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
    sync_to_channel?: boolean;
    channel_name?: string;
    target_channel_id?: number;
};

export type UpstreamSiteUpdateRequest = {
    id: number;
    name?: string;
    enabled?: boolean;
    auto_refresh?: boolean;
    refresh_interval_secs?: number;
    sync_to_channel?: boolean;
    linked_channel_id?: number;
};

export type UpstreamRefreshRequest = {
    id: number;
    apply_channel?: boolean;
};

export type UpstreamApplyRequest = {
    id: number;
    target_channel_id?: number;
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
