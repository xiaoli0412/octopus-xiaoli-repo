'use client';

import { createContext, useContext, useMemo, type ReactNode } from 'react';
import { useRealtimeStats, type ConnectionState } from '@/lib/realtime';

interface RealtimeContextValue {
    /** 统计 SSE 连接状态 */
    statsConnectionState: ConnectionState;
    /** 统计 SSE 是否已连接 */
    statsConnected: boolean;
    /** 最近一次收到的统计快照 */
    statsSnapshot: ReturnType<typeof useRealtimeStats>['stats'];
}

const RealtimeContext = createContext<RealtimeContextValue>({
    statsConnectionState: 'connecting',
    statsConnected: false,
    statsSnapshot: null,
});

/**
 * RealtimeProvider 在应用根注入，管理全局 SSE 连接生命周期。
 *
 * - 统计推送（/api/v1/stream/stats）作为全局单例连接，在已认证后常驻
 * - 日志推送（/api/v1/stream/logs）按需由 Log 模块通过 useRealtimeLogs 订阅
 * - 未认证时不建立连接，避免无效重连
 */
export function RealtimeProvider({ children }: { children: ReactNode }) {
    const { stats, isConnected, connectionState } = useRealtimeStats();

    const value = useMemo<RealtimeContextValue>(
        () => ({
            statsConnectionState: connectionState,
            statsConnected: isConnected,
            statsSnapshot: stats,
        }),
        [stats, isConnected, connectionState],
    );

    return <RealtimeContext.Provider value={value}>{children}</RealtimeContext.Provider>;
}

/** 获取全局实时连接状态 */
export function useRealtimeContext(): RealtimeContextValue {
    return useContext(RealtimeContext);
}
