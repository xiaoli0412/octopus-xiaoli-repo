'use client';

import { useMemo, useState } from 'react';
import { Activity, AlertTriangle, CheckCircle2, RefreshCw, TrendingUp, XCircle } from 'lucide-react';
import {
    Bar,
    BarChart,
    CartesianGrid,
    Line,
    LineChart,
    XAxis,
    YAxis,
} from 'recharts';

import {
    type UpstreamHealthItem,
    type UpstreamHealthStatus,
    type UpstreamUsagePoint,
    useRestoreUpstreamPriority,
    useUpstreamSiteHealth,
    useUpstreamSiteList,
    useUpstreamSiteUsage,
} from '@/api/endpoints/upstream';
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart';
import { cn, formatCount } from '@/lib/utils';

type UsageMetric = 'request_count' | 'total_tokens' | 'cost';
type UsageChartType = 'line' | 'bar';

function statusLabel(status: UpstreamHealthStatus) {
    switch (status) {
        case 'healthy':
            return '健康';
        case 'degraded':
            return '亚健康';
        case 'unhealthy':
            return '不健康';
        default:
            return '未知';
    }
}

function statusIcon(status: UpstreamHealthStatus) {
    switch (status) {
        case 'healthy':
            return CheckCircle2;
        case 'degraded':
            return AlertTriangle;
        case 'unhealthy':
            return XCircle;
        default:
            return Activity;
    }
}

function statusTone(status: UpstreamHealthStatus) {
    switch (status) {
        case 'healthy':
            return 'border-emerald-500/25 bg-emerald-500/[0.04]';
        case 'degraded':
            return 'border-amber-500/25 bg-amber-500/[0.04]';
        case 'unhealthy':
            return 'border-destructive/25 bg-destructive/5';
        default:
            return 'border-border/60 bg-background/55';
    }
}

function statusTextTone(status: UpstreamHealthStatus) {
    switch (status) {
        case 'healthy':
            return 'text-emerald-600 dark:text-emerald-300';
        case 'degraded':
            return 'text-amber-600 dark:text-amber-300';
        case 'unhealthy':
            return 'text-destructive';
        default:
            return 'text-muted-foreground';
    }
}

function formatReason(reason: string) {
    const map: Record<string, string> = {
        site_disabled: '站点已禁用',
        never_refreshed: '从未探测',
        refresh_stale: '探测过期',
        last_refresh_failed: '上次探测失败',
        no_models: '无可用模型',
    };
    if (map[reason]) return map[reason];
    if (reason.startsWith('balance_low_')) return '余额不足';
    if (reason.startsWith('high_error_rate_')) return '错误率高';
    if (reason.startsWith('suppressed_')) return '已自动降级';
    return reason;
}

function metricLabel(metric: UsageMetric) {
    switch (metric) {
        case 'request_count':
            return '请求量';
        case 'total_tokens':
            return '总 Token';
        case 'cost':
            return '费用';
        default:
            return '';
    }
}

function metricValue(point: UpstreamUsagePoint, metric: UsageMetric) {
    switch (metric) {
        case 'request_count':
            return point.request_count;
        case 'total_tokens':
            return point.total_tokens;
        case 'cost':
            return point.cost;
        default:
            return 0;
    }
}

function formatMetric(metric: UsageMetric, value: number) {
    if (metric === 'cost') {
        return `$${value.toFixed(4)}`;
    }
    const formatted = formatCount(value).formatted;
    return `${formatted.value}${formatted.unit}`;
}

function HealthCard({
    item,
    onRestore,
    restoring,
}: {
    item: UpstreamHealthItem;
    onRestore: (id: number) => void;
    restoring: boolean;
}) {
    const Icon = statusIcon(item.status);
    return (
        <div className={cn('octo-stat-card flex flex-col gap-2', statusTone(item.status))}>
            <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                    <div className="truncate text-sm font-semibold text-card-foreground">{item.name}</div>
                    <div className={cn('mt-0.5 flex items-center gap-1.5 text-xs font-medium', statusTextTone(item.status))}>
                        <Icon className="size-3.5" />
                        {statusLabel(item.status)}
                        {item.suppressed && <span className="rounded-full border border-destructive/30 px-1.5 py-0 text-[10px]">已降级</span>}
                    </div>
                </div>
                {item.suppressed && (
                    <button
                        type="button"
                        disabled={restoring}
                        onClick={() => onRestore(item.id)}
                        className="inline-flex shrink-0 items-center gap-1 rounded-xl border border-border/60 bg-background/70 px-2 py-1 text-[11px] text-muted-foreground transition hover:text-foreground disabled:opacity-50"
                    >
                        <RefreshCw className={cn('size-3', restoring && 'animate-spin')} />
                        恢复
                    </button>
                )}
            </div>

            <div className="grid grid-cols-2 gap-2 text-xs text-muted-foreground">
                <div>
                    <span className="text-[11px]">模型</span>
                    <div className="font-medium text-card-foreground">{item.model_count}</div>
                </div>
                <div>
                    <span className="text-[11px]">错误率</span>
                    <div className="font-medium text-card-foreground">{(item.error_rate * 100).toFixed(1)}%</div>
                </div>
                <div>
                    <span className="text-[11px]">余额</span>
                    <div className="font-medium text-card-foreground">
                        {item.balance_available ? (item.balance_unlimited ? '不限' : formatCount(item.balance_remain).formatted.value) : '无'}
                    </div>
                </div>
                <div>
                    <span className="text-[11px]">探测</span>
                    <div className="font-medium text-card-foreground">
                        {item.last_refresh_at ? new Date(item.last_refresh_at).toLocaleDateString() : '未探测'}
                    </div>
                </div>
            </div>

            {item.reasons && item.reasons.length > 0 && (
                <div className="flex flex-wrap gap-1">
                    {item.reasons.map((reason) => (
                        <span key={reason} className="rounded-full border border-border/50 bg-background/60 px-1.5 py-0.5 text-[10px] text-muted-foreground">
                            {formatReason(reason)}
                        </span>
                    ))}
                </div>
            )}
        </div>
    );
}

function UsageChart({
    points,
    metric,
    chartType,
}: {
    points: UpstreamUsagePoint[];
    metric: UsageMetric;
    chartType: UsageChartType;
}) {
    const color = 'var(--chart-1)';
    const chartConfig = { metric: { label: metricLabel(metric), color } };

    const chartData = useMemo(
        () =>
            points.map((point) => ({
                ...point,
                value: metricValue(point, metric),
                label: point.date.slice(5),
            })),
        [points, metric],
    );

    const hasData = useMemo(() => chartData.some((p) => p.value > 0), [chartData]);

    if (!hasData) {
        return (
            <div className="flex h-[6rem] items-center justify-center rounded-xl border border-dashed border-card-border bg-background/35 px-4 text-center text-sm leading-6 text-muted-foreground">
                所选时间范围内暂无用量数据。
            </div>
        );
    }

    return (
        <ChartContainer config={chartConfig} className="h-[10rem] w-full">
            {chartType === 'line' ? (
                <LineChart data={chartData} margin={{ left: 8, right: 8, top: 8, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} />
                    <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={18} />
                    <YAxis tickLine={false} axisLine={false} />
                    <ChartTooltip cursor={false} content={<ChartTooltipContent indicator="line" />} />
                    <Line type="monotone" dataKey="value" stroke={color} strokeWidth={2.2} dot={{ r: 3, fill: color }} activeDot={{ r: 5 }} />
                </LineChart>
            ) : (
                <BarChart data={chartData} margin={{ left: 8, right: 8, top: 8, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} />
                    <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={18} />
                    <YAxis tickLine={false} axisLine={false} />
                    <ChartTooltip cursor={false} content={<ChartTooltipContent indicator="line" />} />
                    <Bar dataKey="value" radius={[8, 8, 0, 0]} fill={color} />
                </BarChart>
            )}
        </ChartContainer>
    );
}

export function UpstreamHealthPanel() {
    const { data: sites = [] } = useUpstreamSiteList();
    const { data: healthItems = [] } = useUpstreamSiteHealth();
    const [selectedSiteId, setSelectedSiteId] = useState<number | undefined>(sites[0]?.id);
    const [usageDays, setUsageDays] = useState<7 | 30>(7);
    const [usageMetric, setUsageMetric] = useState<UsageMetric>('request_count');
    const [usageChartType, setUsageChartType] = useState<UsageChartType>('line');
    const { mutate: restore, isPending: restoring } = useRestoreUpstreamPriority();
    const { data: usage } = useUpstreamSiteUsage(selectedSiteId, usageDays);

    const healthById = useMemo(() => {
        const map = new Map<number, UpstreamHealthItem>();
        for (const item of healthItems) {
            map.set(item.id, item);
        }
        return map;
    }, [healthItems]);

    const selectedHealth = selectedSiteId != null ? healthById.get(selectedSiteId) : undefined;

    return (
        <div className="space-y-3">
            <div className="octo-section">
                <div className="octo-toolbar">
                    <div className="flex items-center gap-2 text-base font-semibold text-card-foreground">
                        <Activity className="size-4.5 text-primary" />
                        上游站点健康
                    </div>
                    <div className="text-xs text-muted-foreground">基于探测时间、余额、模型可用性与错误率评估</div>
                </div>
                <div className="mt-2 grid gap-2 md:grid-cols-2 xl:grid-cols-3">
                    {sites.length === 0 ? (
                        <div className="octo-stat-card text-xs text-muted-foreground">暂无上游站点</div>
                    ) : (
                        sites.map((site) => {
                            const item = healthById.get(site.id);
                            if (!item) return null;
                            return (
                                <button
                                    key={site.id}
                                    type="button"
                                    onClick={() => setSelectedSiteId(site.id)}
                                    className={cn('text-left', selectedSiteId === site.id && 'ring-1 ring-primary/30')}
                                >
                                    <HealthCard item={item} onRestore={restore} restoring={restoring} />
                                </button>
                            );
                        })
                    )}
                </div>
            </div>

            <div className="octo-section">
                <div className="octo-toolbar">
                    <div className="flex items-center gap-2 text-base font-semibold text-card-foreground">
                        <TrendingUp className="size-4.5 text-primary" />
                        用量趋势
                    </div>
                    <div className="text-xs text-muted-foreground">{selectedHealth?.name ?? '选择站点'} · 最近 {usageDays} 天</div>
                </div>

                <div className="mt-2 flex flex-wrap gap-2">
                    {([7, 30] as const).map((d) => (
                        <button
                            key={d}
                            type="button"
                            onClick={() => setUsageDays(d)}
                            className={cn(
                                'rounded-xl border px-2.5 py-1.5 text-xs transition',
                                usageDays === d ? 'border-primary/30 bg-primary/10 text-primary' : 'border-border/60 bg-background/70 text-muted-foreground hover:text-foreground',
                            )}
                        >
                            {d} 天
                        </button>
                    ))}
                    {(['request_count', 'total_tokens', 'cost'] as const).map((m) => (
                        <button
                            key={m}
                            type="button"
                            onClick={() => setUsageMetric(m)}
                            className={cn(
                                'rounded-xl border px-2.5 py-1.5 text-xs transition',
                                usageMetric === m ? 'border-primary/30 bg-primary/10 text-primary' : 'border-border/60 bg-background/70 text-muted-foreground hover:text-foreground',
                            )}
                        >
                            {metricLabel(m)}
                        </button>
                    ))}
                    {(['line', 'bar'] as const).map((t) => (
                        <button
                            key={t}
                            type="button"
                            onClick={() => setUsageChartType(t)}
                            className={cn(
                                'rounded-xl border px-2.5 py-1.5 text-xs transition',
                                usageChartType === t ? 'border-primary/30 bg-primary/10 text-primary' : 'border-border/60 bg-background/70 text-muted-foreground hover:text-foreground',
                            )}
                        >
                            {t === 'line' ? '折线' : '柱状'}
                        </button>
                    ))}
                </div>

                <div className="mt-2 rounded-xl border border-card-border bg-background/40 p-2">
                    <UsageChart points={usage?.points ?? []} metric={usageMetric} chartType={usageChartType} />
                </div>

                {usage && usage.points.length > 0 && (
                    <div className="mt-2 grid grid-cols-2 gap-2 text-xs text-muted-foreground md:grid-cols-4">
                        <div className="octo-stat-card">
                            <div className="text-[11px]">总请求</div>
                            <div className="text-sm font-semibold text-card-foreground">
                                {formatCount(usage.points.reduce((sum, p) => sum + p.request_count, 0)).formatted.value}
                            </div>
                        </div>
                        <div className="octo-stat-card">
                            <div className="text-[11px]">总 Token</div>
                            <div className="text-sm font-semibold text-card-foreground">
                                {formatCount(usage.points.reduce((sum, p) => sum + p.total_tokens, 0)).formatted.value}
                            </div>
                        </div>
                        <div className="octo-stat-card">
                            <div className="text-[11px]">总费用</div>
                            <div className="text-sm font-semibold text-card-foreground">
                                ${usage.points.reduce((sum, p) => sum + p.cost, 0).toFixed(4)}
                            </div>
                        </div>
                        <div className="octo-stat-card">
                            <div className="text-[11px]">成功率</div>
                            <div className="text-sm font-semibold text-card-foreground">
                                {(() => {
                                    const total = usage.points.reduce((sum, p) => sum + p.request_count, 0);
                                    const success = usage.points.reduce((sum, p) => sum + p.success_count, 0);
                                    return total > 0 ? `${((success / total) * 100).toFixed(1)}%` : '-';
                                })()}
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
