import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../client';
import { logger } from '@/lib/logger';

export interface GroupItem {
    id?: number;
    group_id?: number;
    channel_id: number;
    model_name: string;
    priority: number;
    weight: number;
}

export enum GroupMode {
    RoundRobin = 1,
    Random = 2,
    Failover = 3,
    Weighted = 4,
    AIDynamic = 5,
}

export interface Group {
    id?: number;
    name: string;
    mode: GroupMode;
    match_regex: string;
    first_token_time_out?: number;
    session_keep_time?: number;
    retry_rounds?: number;
    retry_delay_ms?: number;
    failover_window_sec?: number;
    race_after_fails?: number;
    race_concurrency?: number;
    items?: GroupItem[];
}

export interface GroupItemAddRequest {
    channel_id: number;
    model_name: string;
    priority: number;
    weight: number;
}

export interface GroupItemUpdateRequest {
    id: number;
    priority: number;
    weight: number;
}

export interface GroupUpdateRequest {
    id: number;
    name?: string;
    mode?: GroupMode;
    match_regex?: string;
    first_token_time_out?: number;
    session_keep_time?: number;
    retry_rounds?: number;
    retry_delay_ms?: number;
    failover_window_sec?: number;
    race_after_fails?: number;
    race_concurrency?: number;
    items_to_add?: GroupItemAddRequest[];
    items_to_update?: GroupItemUpdateRequest[];
    items_to_delete?: number[];
}

export function useGroupList() {
    return useQuery({
        queryKey: ['groups', 'list'],
        queryFn: async () => {
            return apiClient.get<Group[]>('/api/v1/group/list');
        },
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

export function useCreateGroup() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: Group) => {
            return apiClient.post<Group>('/api/v1/group/create', data);
        },
        onSuccess: (data) => {
            logger.log('分组创建成功:', data);
            queryClient.invalidateQueries({ queryKey: ['groups', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
        },
        onError: (error) => {
            logger.error('分组创建失败:', error);
        },
    });
}

export function useUpdateGroup() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: GroupUpdateRequest) => {
            return apiClient.post<Group>('/api/v1/group/update', data);
        },
        onSuccess: (data) => {
            logger.log('分组更新成功:', data);
            queryClient.invalidateQueries({ queryKey: ['groups', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
        },
        onError: (error) => {
            logger.error('分组更新失败:', error);
        },
    });
}

export function useDeleteGroup() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (id: number) => {
            return apiClient.delete<null>(`/api/v1/group/delete/${id}`);
        },
        onSuccess: () => {
            logger.log('分组删除成功');
            queryClient.invalidateQueries({ queryKey: ['groups', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['models', 'capability-inventory'] });
        },
        onError: (error) => {
            logger.error('分组删除失败:', error);
        },
    });
}
