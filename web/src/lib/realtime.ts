/**
 * 实时更新 (SSE) 工具与 React Hooks
 *
 * 设计目标：
 * - 用 SSE 替代高频 TanStack Query 轮询（stats total/hourly、logs）
 * - 自动重连（指数退避，最大 30s）
 * - 连续失败 3 次后降级到 TanStack Query 轮询
 * - 保留 TanStack Query 作为初始加载与降级方案（不删除现有 hooks）
 * - 全局连接状态由 RealtimeProvider 管理
 */

import { useEffect, useRef, useState, useCallback, useMemo } from 'react';
import { useQueryClient, useInfiniteQuery, type InfiniteData } from '@tanstack/react-query';
import { getResolvedAuthToken, buildApiUrl, apiClient } from '@/api/client';
import { logger } from '@/lib/logger';
import type { StatsTotal, StatsDaily, StatsHourly } from '@/api/endpoints/stats';
import type { RelayLog } from '@/api/endpoints/log';

/** SSE 连接状态 */
export type ConnectionState = 'connecting' | 'connected' | 'reconnecting' | 'fallback';

/** 自动重连最大退避间隔（毫秒） */
const MAX_RECONNECT_DELAY = 30_000;
/** 触发降级轮询的连续失败次数 */
const FALLBACK_FAILURE_THRESHOLD = 3;
/** 降级轮询间隔（毫秒） */
const FALLBACK_POLL_INTERVAL = 10_000;

interface SSEClient {
    close: () => void;
}

/**
 * 创建一个带自动重连与降级能力的 SSE 客户端。
 *
 * 浏览器 EventSource 不支持自定义 header，因此 JWT token 通过 query param 传递。
 *
 * @param path       SSE 端点路径（如 /api/v1/stream/stats）
 * @param onMessage  收到消息时回调（data 字段已解析为对象）
 * @param onState    连接状态变更回调
 * @returns          含 close() 的客户端句柄
 */
export function createSSEClient(
    path: string,
    onMessage: (data: unknown) => void,
    onState?: (state: ConnectionState) => void,
): SSEClient {
    let closed = false;
    let eventSource: EventSource | null = null;
    let failureCount = 0;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
    let fallbackPollTimer: ReturnType<typeof setInterval> | null = null;

    const setState = (state: ConnectionState) => {
        if (!closed) onState?.(state);
    };

    const connect = () => {
        if (closed) return;
        setState(failureCount === 0 ? 'connecting' : 'reconnecting');

        // EventSource 不支持自定义 header，JWT 通过 query param 传递
        const token = getResolvedAuthToken();
        const fullUrl = token ? buildApiUrl(path, { token }) : buildApiUrl(path);

        try {
            eventSource = new EventSource(fullUrl);
        } catch (e) {
            logger.error('创建 EventSource 失败:', e);
            handleFailure();
            return;
        }

        eventSource.onopen = () => {
            failureCount = 0;
            setState('connected');
        };

        eventSource.onmessage = (event) => {
            try {
                const data: unknown = JSON.parse(event.data);
                onMessage(data);
            } catch (e) {
                logger.error('解析 SSE 数据失败:', e);
            }
        };

        eventSource.onerror = () => {
            eventSource?.close();
            eventSource = null;
            handleFailure();
        };
    };

    const handleFailure = () => {
        if (closed) return;
        failureCount++;

        if (failureCount >= FALLBACK_FAILURE_THRESHOLD) {
            logger.warn(`SSE 连续失败 ${failureCount} 次，降级到轮询模式`);
            setState('fallback');
            startFallbackPoll();
            return;
        }

        // 指数退避：1s, 2s, 4s, 8s ... 上限 30s
        const delay = Math.min(1000 * Math.pow(2, failureCount - 1), MAX_RECONNECT_DELAY);
        setState('reconnecting');
        reconnectTimer = setTimeout(() => {
            reconnectTimer = null;
            connect();
        }, delay);
    };

    const startFallbackPoll = () => {
        if (fallbackPollTimer || closed) return;
        // 降级时通过 onMessage 触发一次“空事件”，让调用方自行 refetch；
        // 这里用一个轻量心跳：传递 __fallback 标记，hook 内部会触发 refetch。
        fallbackPollTimer = setInterval(() => {
            if (closed) return;
            onMessage({ __fallback: true });
        }, FALLBACK_POLL_INTERVAL);
    };

    const close = () => {
        closed = true;
        if (reconnectTimer) {
            clearTimeout(reconnectTimer);
            reconnectTimer = null;
        }
        if (fallbackPollTimer) {
            clearInterval(fallbackPollTimer);
            fallbackPollTimer = null;
        }
        if (eventSource) {
            eventSource.close();
            eventSource = null;
        }
    };

    connect();

    return { close };
}

/** 后端推送的 stats 快照（与 op.StatsSnapshot 对齐） */
interface StatsSnapshot {
    total: StatsTotal;
    today: StatsDaily;
    hourly: StatsHourly;
}

/**
 * 订阅实时统计 SSE，并把收到的快照写回 TanStack Query 缓存。
 *
 * 现有的 useStatsTotal / useStatsToday / useStatsHourly 会自动感知缓存更新，
 * 因此无需修改这些 hook 即可拿到实时数据。TanStack Query 仍作为初始加载
 * 与降级方案保留。
 *
 * @returns { stats, isConnected, connectionState }
 */
export function useRealtimeStats() {
    const queryClient = useQueryClient();
    const [connectionState, setConnectionState] = useState<ConnectionState>('connecting');
    const [lastSnapshot, setLastSnapshot] = useState<StatsSnapshot | null>(null);
    const clientRef = useRef<SSEClient | null>(null);

    useEffect(() => {
        // 服务端渲染或未认证时不订阅
        if (typeof window === 'undefined') return;
        if (!getResolvedAuthToken()) return;

        const client = createSSEClient(
            '/api/v1/stream/stats',
            (data) => {
                const snapshot = data as StatsSnapshot & { __fallback?: boolean };
                if (!snapshot || snapshot.__fallback) {
                    // 降级心跳：让 TanStack Query 走原有 refetchInterval
                    return;
                }
                setLastSnapshot(snapshot);
                if (snapshot.total) {
                    queryClient.setQueryData(['stats', 'total'], snapshot.total);
                }
                if (snapshot.today) {
                    queryClient.setQueryData(['stats', 'today'], snapshot.today);
                }
                if (snapshot.hourly) {
                    queryClient.setQueryData<StatsHourly[]>(['stats', 'hourly'], (old) => {
                        if (!old) return [snapshot.hourly];
                        const idx = old.findIndex((h) => h.hour === snapshot.hourly.hour && h.date === snapshot.hourly.date);
                        if (idx >= 0) {
                            const next = old.slice();
                            next[idx] = snapshot.hourly;
                            return next;
                        }
                        return [...old, snapshot.hourly];
                    });
                }
            },
            setConnectionState,
        );
        clientRef.current = client;

        return () => {
            client.close();
            clientRef.current = null;
        };
    }, [queryClient]);

    return {
        stats: lastSnapshot,
        isConnected: connectionState === 'connected',
        connectionState,
    };
}

interface UseRealtimeLogsOptions {
    pageSize?: number;
}

interface UseRealtimeLogsResult {
    logs: RelayLog[];
    isConnected: boolean;
    connectionState: ConnectionState;
    paused: boolean;
    pause: () => void;
    resume: () => void;
    /** 最近一次 SSE 推送的新日志 id（用于高亮闪现动画） */
    latestLogId: number | null;
    hasMore: boolean;
    isLoading: boolean;
    isLoadingMore: boolean;
    loadMore: () => Promise<void>;
    error: Error | null;
}

const realtimeLogsQueryKey = (pageSize: number) => ['logs', 'realtime', pageSize] as const;

/**
 * 订阅实时日志 SSE，整合历史分页加载与实时推送。
 *
 * - 历史日志通过 TanStack Query 无限分页加载
 * - 新日志通过 SSE 推送，暂停期间缓存到 pending 队列，恢复后批量合入
 * - 连续失败 3 次自动降级到轮询
 */
export function useRealtimeLogs(options: UseRealtimeLogsOptions = {}): UseRealtimeLogsResult {
    const { pageSize = 20 } = options;
    const queryClient = useQueryClient();

    const [connectionState, setConnectionState] = useState<ConnectionState>('connecting');
    const [paused, setPaused] = useState(false);
    const [latestLogId, setLatestLogId] = useState<number | null>(null);
    const [error, setError] = useState<Error | null>(null);
    const clientRef = useRef<SSEClient | null>(null);
    const pausedQueueRef = useRef<RelayLog[]>([]);
    // 用 ref 跟踪 paused 状态，避免 pause/resume 触发 SSE 重连
    const pausedRef = useRef(false);
    pausedRef.current = paused;

    const logsQuery = useInfiniteQuery({
        queryKey: realtimeLogsQueryKey(pageSize),
        initialPageParam: 1,
        queryFn: async ({ pageParam }) => {
            const params = new URLSearchParams();
            params.set('page', String(pageParam));
            params.set('page_size', String(pageSize));
            const result = await apiClient.get<RelayLog[] | null>(`/api/v1/log/list?${params.toString()}`);
            return result ?? [];
        },
        getNextPageParam: (lastPage, allPages) => {
            if (!lastPage || lastPage.length < pageSize) return undefined;
            return allPages.length + 1;
        },
        staleTime: Infinity,
        refetchOnMount: 'always',
    });

    // 合并无限分页日志并按时间倒序去重
    const logs = useMemo(() => {
        const pages = logsQuery.data?.pages ?? [];
        const seen = new Set<number>();
        const merged: RelayLog[] = [];
        for (const page of pages) {
            for (const log of page) {
                if (seen.has(log.id)) continue;
                seen.add(log.id);
                merged.push(log);
            }
        }
        merged.sort((a, b) => b.time - a.time);
        return merged;
    }, [logsQuery.data]);

    const loadMore = useCallback(async () => {
        if (!logsQuery.hasNextPage || logsQuery.isFetchingNextPage) return;
        try {
            await logsQuery.fetchNextPage();
        } catch (e) {
            logger.error('加载更多日志失败:', e);
        }
    }, [logsQuery]);

    const pushLogIntoCache = useCallback((log: RelayLog) => {
        queryClient.setQueryData<InfiniteData<RelayLog[], number>>(
            realtimeLogsQueryKey(pageSize),
            (old) => {
                if (!old) {
                    return { pages: [[log]], pageParams: [1] };
                }
                const exists = old.pages.some((p) => p?.some((x) => x.id === log.id));
                if (exists) return old;
                const firstPage = old.pages[0] ?? [];
                return { ...old, pages: [[log, ...firstPage], ...old.pages.slice(1)] };
            },
        );
        setLatestLogId(log.id);
    }, [pageSize, queryClient]);

    useEffect(() => {
        if (typeof window === 'undefined') return;
        if (!getResolvedAuthToken()) return;

        const client = createSSEClient(
            '/api/v1/stream/logs',
            (data) => {
                const log = data as RelayLog & { __fallback?: boolean };
                if (!log || log.__fallback) {
                    // 降级心跳：触发一次 refetch
                    void logsQuery.refetch();
                    return;
                }
                if (pausedRef.current) {
                    pausedQueueRef.current.push(log);
                    return;
                }
                pushLogIntoCache(log);
            },
            (state) => {
                setConnectionState(state);
                setError(state === 'fallback' ? new Error('SSE 降级到轮询') : null);
            },
        );
        clientRef.current = client;

        return () => {
            client.close();
            clientRef.current = null;
        };
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [pageSize, queryClient, pushLogIntoCache]);

    const pause = useCallback(() => setPaused(true), []);
    const resume = useCallback(() => {
        setPaused(false);
        // 恢复时把暂停期间积压的日志批量合入
        const queued = pausedQueueRef.current;
        if (queued.length > 0) {
            pausedQueueRef.current = [];
            for (const log of queued) {
                pushLogIntoCache(log);
            }
        }
    }, [pushLogIntoCache]);

    return {
        logs,
        isConnected: connectionState === 'connected',
        connectionState,
        paused,
        pause,
        resume,
        latestLogId,
        hasMore: !!logsQuery.hasNextPage,
        isLoading: logsQuery.isLoading,
        isLoadingMore: logsQuery.isFetchingNextPage,
        loadMore,
        error: error ?? (logsQuery.error as Error | null),
    };
}
