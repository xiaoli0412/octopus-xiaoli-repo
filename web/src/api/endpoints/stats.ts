import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../client';
import { formatCount, formatMoney, formatTime } from '@/lib/utils';

/**
 * 统计数据
 */
interface StatsMetrics {
    input_token: number;
    output_token: number;
    input_cost: number;
    output_cost: number;
    wait_time: number;
    request_success: number;
    request_failed: number;
}

export interface StatsMetricsFormatted {
    input_token: ReturnType<typeof formatCount>;
    output_token: ReturnType<typeof formatCount>;
    input_cost: ReturnType<typeof formatMoney>;
    output_cost: ReturnType<typeof formatMoney>;
    wait_time: ReturnType<typeof formatTime>;
    request_success: ReturnType<typeof formatCount>;
    request_failed: ReturnType<typeof formatCount>;

    request_count: ReturnType<typeof formatCount>;
    total_token: ReturnType<typeof formatCount>;
    total_cost: ReturnType<typeof formatMoney>;
}

export interface StatsChannel extends StatsMetrics {
    channel_id: number;
}

export interface StatsDaily extends StatsMetrics {
    date: string;
}
export interface StatsDailyFormatted extends StatsMetricsFormatted {
    date: string;
}

export interface StatsTotal extends StatsMetrics {
    id: number;
}
export type StatsTotalFormatted = StatsMetricsFormatted;

export interface StatsHourly extends StatsMetrics {
    hour: number;
    date: string;
}
export interface StatsHourlyFormatted extends StatsMetricsFormatted {
    hour: number;
    date: string;
}
/**
 * API Key 统计数据
 */
export interface StatsAPIKey extends StatsMetrics {
    api_key_id: number;
}

export interface StatsAPIKeyFormatted extends StatsMetricsFormatted {
    api_key_id: number;
}

export interface StatsTokenBreakdownItem {
    key: string;
    label: string;
    input_token: number;
    output_token: number;
    total_token: number;
}

export interface StatsTokenBreakdownFormattedItem {
    key: string;
    label: string;
    input_token: ReturnType<typeof formatCount>;
    output_token: ReturnType<typeof formatCount>;
    total_token: ReturnType<typeof formatCount>;
}

export interface StatsTokenBreakdown {
    total_input_token: number;
    total_output_token: number;
    total_token: number;
    estimated_official_input_cost: number;
    estimated_official_output_cost: number;
    estimated_official_total_cost: number;
    estimated_gateway_input_cost: number;
    estimated_gateway_output_cost: number;
    estimated_gateway_total_cost: number;
    estimated_price_basis?: string;
    estimated_probe_input_cost: number;
    estimated_probe_output_cost: number;
    estimated_probe_total_cost: number;
    recent_probe_count: number;
    recent_probe_success_count: number;
    recent_probe_failed_count: number;
    recent_probe_last_at: number;
    recent_probe_last_status?: string;
    recent_probe_last_channel?: string;
    recent_probe_last_model?: string;
    recent_probe_last_message?: string;
    probe_summary_basis?: string;
    circuit_tracked_count: number;
    circuit_open_count: number;
    circuit_half_open_count: number;
    circuit_closed_count: number;
    circuit_max_remaining_cooldown_sec: number;
    circuit_summary_basis?: string;
    by_channel: StatsTokenBreakdownItem[];
    by_model: StatsTokenBreakdownItem[];
    by_api_key?: StatsTokenBreakdownItem[];
    by_channel_key?: StatsTokenBreakdownItem[];
}

export interface StatsTokenBreakdownFormatted {
    total_input_token: ReturnType<typeof formatCount>;
    total_output_token: ReturnType<typeof formatCount>;
    total_token: ReturnType<typeof formatCount>;
    estimated_official_input_cost: ReturnType<typeof formatMoney>;
    estimated_official_output_cost: ReturnType<typeof formatMoney>;
    estimated_official_total_cost: ReturnType<typeof formatMoney>;
    estimated_gateway_input_cost: ReturnType<typeof formatMoney>;
    estimated_gateway_output_cost: ReturnType<typeof formatMoney>;
    estimated_gateway_total_cost: ReturnType<typeof formatMoney>;
    estimated_price_basis?: string;
    estimated_probe_input_cost: ReturnType<typeof formatMoney>;
    estimated_probe_output_cost: ReturnType<typeof formatMoney>;
    estimated_probe_total_cost: ReturnType<typeof formatMoney>;
    recent_probe_count: ReturnType<typeof formatCount>;
    recent_probe_success_count: ReturnType<typeof formatCount>;
    recent_probe_failed_count: ReturnType<typeof formatCount>;
    recent_probe_last_at: number;
    recent_probe_last_status?: string;
    recent_probe_last_channel?: string;
    recent_probe_last_model?: string;
    recent_probe_last_message?: string;
    probe_summary_basis?: string;
    circuit_tracked_count: number;
    circuit_open_count: number;
    circuit_half_open_count: number;
    circuit_closed_count: number;
    circuit_max_remaining_cooldown_sec: number;
    circuit_summary_basis?: string;
    by_channel: StatsTokenBreakdownFormattedItem[];
    by_model: StatsTokenBreakdownFormattedItem[];
    by_api_key?: StatsTokenBreakdownFormattedItem[];
    by_channel_key?: StatsTokenBreakdownFormattedItem[];
}

export interface StatsDynamicRoutingSummary {
    last_run_at: string;
    last_success_at: string;
    last_status?: string;
    last_message?: string;
    current_mode?: string;
    effective_mode?: string;
    decision?: string;
    decision_reason?: string;
    health_enabled: boolean;
    channel_count: number;
    enabled_channels: number;
    group_count: number;
    failover_groups: number;
    free_public_keys: number;
    paid_metered_keys: number;
    private_internal_keys: number;
    unknown_keys: number;
    basis?: string;
}
/**
 * 获取今日统计数据 Hook
 */
export function useStatsToday() {
    return useQuery({
        queryKey: ['stats', 'today'],
        queryFn: async () => {
            return apiClient.get<StatsDaily>('/api/v1/stats/today');
        },
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

/**
 * 获取每日统计数据 Hook
 */
export function useStatsDaily() {
    return useQuery({
        queryKey: ['stats', 'daily'],
        queryFn: async () => {
            return apiClient.get<StatsDaily[]>('/api/v1/stats/daily');
        },
        select: (data) => data.map((item): StatsDailyFormatted => ({
            input_token: formatCount(item.input_token),
            output_token: formatCount(item.output_token),
            total_token: formatCount(item.input_token + item.output_token),
            input_cost: formatMoney(item.input_cost),
            output_cost: formatMoney(item.output_cost),
            total_cost: formatMoney(item.input_cost + item.output_cost),
            wait_time: formatTime(item.wait_time),
            request_success: formatCount(item.request_success),
            request_failed: formatCount(item.request_failed),
            request_count: formatCount(item.request_success + item.request_failed),
            date: item.date,
        })),
        refetchInterval: 3600000, // 1 小时
        refetchOnMount: 'always',
    });
}
/**
 * 获取总统计数据 Hook
 */
export function useStatsHourly() {
    return useQuery({
        queryKey: ['stats', 'hourly'],
        queryFn: async () => {
            return apiClient.get<StatsHourly[]>('/api/v1/stats/hourly');
        },
        select: (data) => data.map((item): StatsHourlyFormatted => ({
            hour: item.hour,
            date: item.date,
            input_token: formatCount(item.input_token),
            output_token: formatCount(item.output_token),
            total_token: formatCount(item.input_token + item.output_token),
            input_cost: formatMoney(item.input_cost),
            output_cost: formatMoney(item.output_cost),
            total_cost: formatMoney(item.input_cost + item.output_cost),
            wait_time: formatTime(item.wait_time),
            request_success: formatCount(item.request_success),
            request_failed: formatCount(item.request_failed),
            request_count: formatCount(item.request_success + item.request_failed),
        })),
        refetchInterval: 10000,// 10 秒
        refetchOnMount: 'always',
    });
}

export function useStatsTotal() {
    return useQuery({
        queryKey: ['stats', 'total'],
        queryFn: async () => {
            return apiClient.get<StatsTotal>('/api/v1/stats/total');
        },
        select: (data) => ({
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
        refetchInterval: 10000,// 10 秒
        refetchOnMount: 'always',
    });
}



/**
 * 获取 API Key 统计数据列表 Hook
 */
export function useStatsAPIKey() {
    return useQuery({
        queryKey: ['stats', 'apikey'],
        queryFn: async () => {
            return apiClient.get<StatsAPIKey[]>('/api/v1/stats/apikey');
        },
        select: (data) => data.map((item): StatsAPIKeyFormatted => ({
            api_key_id: item.api_key_id,
            input_token: formatCount(item.input_token),
            output_token: formatCount(item.output_token),
            total_token: formatCount(item.input_token + item.output_token),
            input_cost: formatMoney(item.input_cost),
            output_cost: formatMoney(item.output_cost),
            total_cost: formatMoney(item.input_cost + item.output_cost),
            wait_time: formatTime(item.wait_time),
            request_success: formatCount(item.request_success),
            request_failed: formatCount(item.request_failed),
            request_count: formatCount(item.request_success + item.request_failed),
        })),
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

export type StatsTokenWindow = '12h' | '1d' | '3d' | '7d' | '30d';

export function useStatsTokenBreakdown(window: StatsTokenWindow = '1d') {
    return useQuery({
        queryKey: ['stats', 'token-breakdown', window],
        queryFn: async () => {
            return apiClient.get<StatsTokenBreakdown>(`/api/v1/stats/token-breakdown?window=${window}`);
        },
        select: (data): StatsTokenBreakdownFormatted => ({
            total_input_token: formatCount(data.total_input_token),
            total_output_token: formatCount(data.total_output_token),
            total_token: formatCount(data.total_token),
            estimated_official_input_cost: formatMoney(data.estimated_official_input_cost),
            estimated_official_output_cost: formatMoney(data.estimated_official_output_cost),
            estimated_official_total_cost: formatMoney(data.estimated_official_total_cost),
            estimated_gateway_input_cost: formatMoney(data.estimated_gateway_input_cost),
            estimated_gateway_output_cost: formatMoney(data.estimated_gateway_output_cost),
            estimated_gateway_total_cost: formatMoney(data.estimated_gateway_total_cost),
            estimated_price_basis: data.estimated_price_basis,
            estimated_probe_input_cost: formatMoney(data.estimated_probe_input_cost),
            estimated_probe_output_cost: formatMoney(data.estimated_probe_output_cost),
            estimated_probe_total_cost: formatMoney(data.estimated_probe_total_cost),
            recent_probe_count: formatCount(data.recent_probe_count),
            recent_probe_success_count: formatCount(data.recent_probe_success_count),
            recent_probe_failed_count: formatCount(data.recent_probe_failed_count),
            recent_probe_last_at: data.recent_probe_last_at,
            recent_probe_last_status: data.recent_probe_last_status,
            recent_probe_last_channel: data.recent_probe_last_channel,
            recent_probe_last_model: data.recent_probe_last_model,
            recent_probe_last_message: data.recent_probe_last_message,
            probe_summary_basis: data.probe_summary_basis,
            circuit_tracked_count: data.circuit_tracked_count,
            circuit_open_count: data.circuit_open_count,
            circuit_half_open_count: data.circuit_half_open_count,
            circuit_closed_count: data.circuit_closed_count,
            circuit_max_remaining_cooldown_sec: data.circuit_max_remaining_cooldown_sec,
            circuit_summary_basis: data.circuit_summary_basis,
            by_channel: data.by_channel.map((item) => ({
                key: item.key,
                label: item.label,
                input_token: formatCount(item.input_token),
                output_token: formatCount(item.output_token),
                total_token: formatCount(item.total_token),
            })),
            by_model: data.by_model.map((item) => ({
                key: item.key,
                label: item.label,
                input_token: formatCount(item.input_token),
                output_token: formatCount(item.output_token),
                total_token: formatCount(item.total_token),
            })),
            by_api_key: data.by_api_key?.map((item) => ({
                key: item.key,
                label: item.label,
                input_token: formatCount(item.input_token),
                output_token: formatCount(item.output_token),
                total_token: formatCount(item.total_token),
            })),
            by_channel_key: data.by_channel_key?.map((item) => ({
                key: item.key,
                label: item.label,
                input_token: formatCount(item.input_token),
                output_token: formatCount(item.output_token),
                total_token: formatCount(item.total_token),
            })),
        }),
        refetchInterval: 15000,
        refetchOnMount: 'always',
    });
}

export function useStatsDynamicRoutingSummary() {
    return useQuery({
        queryKey: ['stats', 'dynamic-routing-summary'],
        queryFn: async () => {
            return apiClient.get<StatsDynamicRoutingSummary>('/api/v1/stats/dynamic-routing-summary');
        },
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}
