import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../client';

/**
 * 审计操作类型（与后端 observability.AuditAction 对齐）
 */
export const AUDIT_ACTIONS = [
    'create',
    'update',
    'delete',
    'enable',
    'disable',
    'login',
    'backup',
    'restore',
] as const;

export type AuditAction = (typeof AUDIT_ACTIONS)[number];

/**
 * 审计资源类型（与后端 observability.ResourceType 对齐）
 */
export const AUDIT_RESOURCE_TYPES = [
    'channel',
    'group',
    'apikey',
    'setting',
    'user',
] as const;

export type AuditResourceType = (typeof AUDIT_RESOURCE_TYPES)[number];

/**
 * 审计日志记录（与后端 model.AuditLog 对齐）
 * before_json / after_json 已在后端脱敏，前端直接展示即可。
 */
export interface AuditLog {
    id: number;
    user_id: number;
    username: string;
    action: string;
    resource_type: string;
    resource_id: string;
    resource_name: string;
    before_json: string;
    after_json: string;
    ip: string;
    user_agent: string;
    created_at: string;
}

/**
 * 审计日志列表查询参数
 */
export interface AuditListParams {
    page?: number;
    page_size?: number;
    start_time?: number;
    end_time?: number;
    user_id?: number;
    action?: string;
    resource_type?: string;
    resource_id?: string;
}

export interface AuditListResult {
    list: AuditLog[];
    total: number;
    page: number;
    page_size: number;
}

function buildParams(params: AuditListParams): Record<string, string | number | boolean> {
    const query: Record<string, string | number | boolean> = {};
    if (typeof params.page === 'number') query.page = params.page;
    if (typeof params.page_size === 'number') query.page_size = params.page_size;
    if (typeof params.start_time === 'number') query.start_time = Math.floor(params.start_time / 1000);
    if (typeof params.end_time === 'number') query.end_time = Math.floor(params.end_time / 1000);
    if (typeof params.user_id === 'number' && params.user_id > 0) query.user_id = params.user_id;
    if (params.action) query.action = params.action;
    if (params.resource_type) query.resource_type = params.resource_type;
    if (params.resource_id) query.resource_id = params.resource_id;
    return query;
}

/**
 * 审计日志列表 Hook
 *
 * @example
 * const { data, isLoading } = useAuditList({ page: 1, page_size: 20, action: 'create' });
 * data?.list.forEach(log => console.log(log.username, log.action));
 */
export function useAuditList(params: AuditListParams) {
    const query = buildParams(params);
    const queryKey = ['audit', 'list', query] as const;

    return useQuery({
        queryKey,
        queryFn: async () => {
            const result = await apiClient.get<AuditListResult>('/api/v1/audit/list', query);
            return result;
        },
        placeholderData: (prev) => prev,
    });
}

/**
 * 审计日志详情 Hook
 *
 * @example
 * const { data } = useAuditDetail(42);
 * console.log(data?.before_json, data?.after_json);
 */
export function useAuditDetail(id: number | null) {
    return useQuery({
        queryKey: ['audit', 'detail', id] as const,
        queryFn: async () => {
            return apiClient.get<AuditLog>(`/api/v1/audit/${id}`);
        },
        enabled: id !== null && id > 0,
    });
}
