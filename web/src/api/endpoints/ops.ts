import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../client';

export type OpsScope = 'overall' | 'model' | 'channel' | 'channel_key' | 'api_key' | 'ip';

export interface OpsEntitySummary {
    scope: OpsScope;
    entity_key: string;
    entity_label: string;
    entity_display_label?: string;
    success_count: number;
    failure_count: number;
    skipped_count: number;
    wait_time: number;
    input_token: number;
    output_token: number;
    cache_read_token: number;
    cache_write_token: number;
    cache_hit_count: number;
    cache_write_count: number;
    cache_create_count: number;
    cache_success_count: number;
    cache_eligible_count: number;
    cache_ineligible_count: number;
    cache_supported: boolean;
    success_rate: number;
    cache_hit_rate: number;
    cache_create_rate: number;
    cache_rate: number;
    avg_latency_ms: number;
}

export interface OpsSeriesPoint {
    bucket_start: number;
    label: string;
    success_count: number;
    failure_count: number;
    skipped_count: number;
    wait_time: number;
    input_token: number;
    output_token: number;
    cache_read_token: number;
    cache_write_token: number;
    cache_hit_count: number;
    cache_write_count: number;
    cache_create_count: number;
    cache_success_count: number;
    cache_eligible_count: number;
    cache_ineligible_count: number;
    cache_supported: boolean;
    success_rate: number;
    cache_hit_rate: number;
    cache_create_rate: number;
    cache_rate: number;
    avg_latency_ms: number;
}

export interface OpsOverview {
    window: string;
    total: OpsEntitySummary;
    top_models: OpsEntitySummary[];
    top_channels: OpsEntitySummary[];
    top_channel_keys: OpsEntitySummary[];
    top_api_keys: OpsEntitySummary[];
    top_ips: OpsEntitySummary[];
}

export interface OpsRecentDetail {
    id: number;
    time: number;
    client_ip: string;
    client_ip_label?: string;
    request_model_name: string;
    actual_model_name: string;
    api_key_id: number;
    channel_id: number;
    channel_name: string;
    channel_key_id: number;
    input_tokens: number;
    output_tokens: number;
    cache_read_tokens: number;
    cache_write_tokens: number;
    cache_supported: boolean;
    use_time: number;
    success: boolean;
    status_code: number;
    error: string;
    attempt_count: number;
}

export function useOpsOverview() {
    return useQuery({
        queryKey: ['ops', 'overview'],
        queryFn: async () => apiClient.get<OpsOverview>('/api/v1/ops/overview'),
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

export function useOpsEntityList(scope: OpsScope, limit = 16) {
    return useQuery({
        queryKey: ['ops', 'entities', scope, limit],
        queryFn: async () => apiClient.get<OpsEntitySummary[]>(`/api/v1/ops/entities?scope=${scope}&limit=${limit}`),
        enabled: Boolean(scope),
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

export function useOpsEntitySeries(scope: OpsScope, entityKey: string) {
    return useQuery({
        queryKey: ['ops', 'series', scope, entityKey],
        queryFn: async () => apiClient.get<OpsSeriesPoint[]>(`/api/v1/ops/series?scope=${scope}&entity_key=${encodeURIComponent(entityKey)}`),
        enabled: Boolean(scope && entityKey),
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

export function useOpsRecentDetails(scope: OpsScope, entityKey: string, limit = 12) {
    return useQuery({
        queryKey: ['ops', 'details', scope, entityKey, limit],
        queryFn: async () => apiClient.get<OpsRecentDetail[]>(`/api/v1/ops/details?scope=${scope}&entity_key=${encodeURIComponent(entityKey)}&limit=${limit}`),
        enabled: Boolean(scope),
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}
