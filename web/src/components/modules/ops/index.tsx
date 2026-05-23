'use client';

import { useEffect, useMemo, useState } from 'react';
import {
    Activity,
    BarChart3,
    Database,
    Gauge,
    KeyRound,
    Layers3,
    LineChart as LineChartIcon,
    Network,
    ScatterChart as ScatterChartIcon,
    Server,
    ShieldCheck,
    Waypoints,
} from 'lucide-react';
import {
    Bar,
    BarChart,
    CartesianGrid,
    Cell,
    Line,
    LineChart,
    Scatter,
    ScatterChart,
    XAxis,
    YAxis,
} from 'recharts';

import { type OpsEntitySummary, type OpsRecentDetail, type OpsScope, type OpsSeriesPoint, useOpsEntityList, useOpsEntitySeries, useOpsOverview, useOpsRecentDetails } from '@/api/endpoints/ops';
import { AnimatedNumber } from '@/components/common/AnimatedNumber';
import { PageWrapper } from '@/components/common/PageWrapper';
import { Tabs, TabsContent, TabsContents, TabsList, TabsTrigger } from '@/components/animate-ui/components/animate/tabs';
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart';
import { cn } from '@/lib/utils';
import { formatCount } from '@/lib/utils';

type PrimaryTab = 'overview' | 'model' | 'channel' | 'key' | 'cache' | 'ip';
type ChartType = 'line' | 'bar' | 'scatter';
type ChartMetric = 'success_rate' | 'request_count' | 'avg_latency_ms' | 'cache_hit_rate' | 'cache_rate' | 'cache_read_token' | 'cache_write_token';

const PRIMARY_TABS: Array<{ key: PrimaryTab; label: string; icon: typeof Activity }> = [
    { key: 'overview', label: '总览', icon: Gauge },
    { key: 'model', label: '模型', icon: Layers3 },
    { key: 'channel', label: '渠道', icon: Server },
    { key: 'key', label: 'Key', icon: KeyRound },
    { key: 'cache', label: '缓存', icon: Database },
    { key: 'ip', label: 'IP', icon: Network },
];

const CHART_TYPES: Array<{ key: ChartType; label: string; icon: typeof Activity }> = [
    { key: 'line', label: '折线', icon: LineChartIcon },
    { key: 'bar', label: '柱状', icon: BarChart3 },
    { key: 'scatter', label: '点状', icon: ScatterChartIcon },
];

const DEFAULT_ENTITY_KEYS: Record<OpsScope, string> = {
    overall: 'all',
    model: '',
    channel: '',
    channel_key: '',
    api_key: '',
    ip: '',
};

function percent(value: number) {
    return `${(value * 100).toFixed(1)}%`;
}

function duration(value: number) {
    if (value >= 1000) {
        return `${(value / 1000).toFixed(2)}s`;
    }
    return `${Math.round(value)}ms`;
}

function shortNumber(value: number) {
    const formatted = formatCount(value).formatted;
    return `${formatted.value}${formatted.unit}`;
}

function metricLabel(metric: ChartMetric) {
    switch (metric) {
        case 'request_count':
            return '请求量';
        case 'avg_latency_ms':
            return '平均延迟';
        case 'cache_hit_rate':
            return '缓存命中率';
        case 'cache_rate':
            return '缓存率';
        case 'cache_read_token':
            return '缓存读取';
        case 'cache_write_token':
            return '缓存写入';
        default:
            return '成功率';
    }
}

function metricValue(summary: OpsEntitySummary, metric: ChartMetric) {
    switch (metric) {
        case 'request_count':
            return summary.success_count + summary.failure_count;
        case 'avg_latency_ms':
            return summary.avg_latency_ms;
        case 'cache_hit_rate':
            return summary.cache_hit_rate;
        case 'cache_rate':
            return summary.cache_rate;
        case 'cache_read_token':
            return summary.cache_read_token;
        case 'cache_write_token':
            return summary.cache_write_token;
        default:
            return summary.success_rate;
    }
}

function formatMetric(metric: ChartMetric, value: number) {
    switch (metric) {
        case 'success_rate':
        case 'cache_hit_rate':
        case 'cache_rate':
            return percent(value);
        case 'avg_latency_ms':
            return duration(value);
        default:
            return shortNumber(value);
    }
}

function pointValue(point: OpsSeriesPoint, metric: ChartMetric) {
    switch (metric) {
        case 'request_count':
            return point.success_count + point.failure_count;
        case 'avg_latency_ms':
            return point.avg_latency_ms;
        case 'cache_hit_rate':
            return point.cache_hit_rate;
        case 'cache_rate':
            return point.cache_rate;
        case 'cache_read_token':
            return point.cache_read_token;
        case 'cache_write_token':
            return point.cache_write_token;
        default:
            return point.success_rate;
    }
}

function chartColor(metric: ChartMetric) {
    switch (metric) {
        case 'request_count':
            return 'var(--chart-2)';
        case 'avg_latency_ms':
            return 'var(--chart-5)';
        case 'cache_hit_rate':
            return 'var(--chart-3)';
        case 'cache_rate':
            return 'var(--chart-4)';
        case 'cache_read_token':
            return 'var(--chart-1)';
        case 'cache_write_token':
            return 'var(--chart-2)';
        default:
            return 'var(--chart-1)';
    }
}

function cardTone(successRate: number) {
    if (successRate >= 0.95) return 'border-emerald-500/30 bg-emerald-500/8';
    if (successRate >= 0.85) return 'border-primary/30 bg-primary/8';
    if (successRate >= 0.7) return 'border-amber-500/30 bg-amber-500/8';
    return 'border-destructive/30 bg-destructive/8';
}

function SummaryCard({ title, value, hint, icon: Icon }: { title: string; value: string; hint: string; icon: typeof Activity }) {
    return (
        <div className="rounded-2xl border border-card-border bg-background/55 px-4 py-3">
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <Icon className="size-3.5" />
                {title}
            </div>
            <div className="mt-2 text-2xl font-semibold text-foreground">{value}</div>
            <div className="mt-1 text-[11px] leading-5 text-muted-foreground">{hint}</div>
        </div>
    );
}

function EntityCard({
    item,
    active,
    metric,
    onClick,
}: {
    item: OpsEntitySummary;
    active: boolean;
    metric: ChartMetric;
    onClick: () => void;
}) {
    const requestCount = item.success_count + item.failure_count;
    return (
        <button
            type="button"
            onClick={onClick}
            className={cn(
                'w-full rounded-2xl border px-3.5 py-3 text-left transition',
                active
                    ? 'border-primary/35 bg-primary/10 text-primary'
                    : `${cardTone(item.success_rate)} text-card-foreground hover:border-primary/20`,
            )}
        >
            <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                    <div className="truncate text-sm font-medium">{item.entity_label || item.entity_key}</div>
                    <div className="mt-1 text-[11px] text-muted-foreground">{requestCount} 请求 · 成功率 {percent(item.success_rate)}</div>
                </div>
                <div className="rounded-full border border-border/60 bg-background/70 px-2 py-0.5 text-[10px] text-muted-foreground">
                    {metricLabel(metric)}
                </div>
            </div>
            <div className="mt-3 flex items-end justify-between gap-3">
                <div className="text-lg font-semibold tabular-nums">{formatMetric(metric, metricValue(item, metric))}</div>
                <div className="text-[11px] text-muted-foreground">缓存率 {percent(item.cache_rate)}</div>
            </div>
        </button>
    );
}

function DetailsTable({ rows }: { rows: OpsRecentDetail[] }) {
    if (!rows.length) {
        return (
            <div className="rounded-2xl border border-dashed border-card-border bg-background/35 px-4 py-6 text-sm text-muted-foreground">
                最近 12 小时没有可展示的调用明细。
            </div>
        );
    }

    return (
        <div className="overflow-hidden rounded-2xl border border-card-border bg-background/40">
            <div className="grid grid-cols-[9rem_1fr_1fr_7rem_7rem] gap-3 border-b border-card-border px-4 py-3 text-[11px] text-muted-foreground">
                <div>时间 / IP</div>
                <div>模型</div>
                <div>渠道 / Key</div>
                <div>状态</div>
                <div>缓存</div>
            </div>
            <div className="max-h-[24rem] overflow-y-auto">
                {rows.map((row) => (
                    <div key={row.id} className="grid grid-cols-[9rem_1fr_1fr_7rem_7rem] gap-3 border-b border-card-border/60 px-4 py-3 text-xs last:border-b-0">
                        <div className="space-y-1">
                            <div className="font-medium text-card-foreground">{new Date(row.time * 1000).toLocaleTimeString()}</div>
                            <div className="break-all text-muted-foreground">{row.client_ip || '-'}</div>
                        </div>
                        <div className="space-y-1">
                            <div className="break-all text-card-foreground">{row.actual_model_name || row.request_model_name}</div>
                            <div className="break-all text-muted-foreground">{row.request_model_name}</div>
                        </div>
                        <div className="space-y-1">
                            <div className="break-all text-card-foreground">{row.channel_name || `Channel ${row.channel_id}`}</div>
                            <div className="text-muted-foreground">Key #{row.channel_key_id || '-'}</div>
                        </div>
                        <div className="space-y-1">
                            <div className={cn('font-medium', row.success ? 'text-emerald-600 dark:text-emerald-300' : 'text-destructive')}>
                                {row.success ? '成功' : '失败'}
                            </div>
                            <div className="text-muted-foreground">{duration(row.use_time)}</div>
                        </div>
                        <div className="space-y-1 text-muted-foreground">
                            <div>读 {shortNumber(row.cache_read_tokens)}</div>
                            <div>写 {shortNumber(row.cache_write_tokens)}</div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}

function SeriesChart({ data, metric, chartType }: { data: OpsSeriesPoint[]; metric: ChartMetric; chartType: ChartType }) {
    const color = chartColor(metric);
    const chartConfig = { metric: { label: metricLabel(metric), color } };

    const chartData = useMemo(() => data.map((point, index) => ({
        ...point,
        chartMetric: pointValue(point, metric),
        index,
    })), [data, metric]);

    const peak = useMemo(() => chartData.reduce((value, point) => Math.max(value, point.chartMetric), 0), [chartData]);

    return (
        <ChartContainer config={chartConfig} className="h-[20rem] w-full">
            {chartType === 'line' ? (
                <LineChart data={chartData} margin={{ left: 8, right: 8, top: 8, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} />
                    <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={18} />
                    <YAxis tickLine={false} axisLine={false} />
                    <ChartTooltip cursor={false} content={<ChartTooltipContent indicator="line" />} />
                    <Line type="monotone" dataKey="chartMetric" stroke={color} strokeWidth={2.2} dot={{ r: 3, fill: color }} activeDot={{ r: 5 }} />
                </LineChart>
            ) : chartType === 'bar' ? (
                <BarChart data={chartData} margin={{ left: 8, right: 8, top: 8, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} />
                    <XAxis dataKey="label" tickLine={false} axisLine={false} minTickGap={18} />
                    <YAxis tickLine={false} axisLine={false} />
                    <ChartTooltip cursor={false} content={<ChartTooltipContent indicator="line" />} />
                    <Bar dataKey="chartMetric" radius={[8, 8, 0, 0]}>
                        {chartData.map((entry) => (
                            <Cell key={`${entry.bucket_start}-${entry.index}`} fill={entry.chartMetric === peak && peak > 0 ? 'var(--chart-2)' : color} />
                        ))}
                    </Bar>
                </BarChart>
            ) : (
                <ScatterChart data={chartData} margin={{ left: 8, right: 8, top: 8, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} />
                    <XAxis dataKey="label" type="category" tickLine={false} axisLine={false} minTickGap={18} />
                    <YAxis dataKey="chartMetric" type="number" tickLine={false} axisLine={false} />
                    <ChartTooltip cursor={false} content={<ChartTooltipContent indicator="dot" />} />
                    <Scatter data={chartData} dataKey="chartMetric" fill={color} />
                </ScatterChart>
            )}
        </ChartContainer>
    );
}

function ScopeWorkspace({
    title,
    desc,
    scope,
    metric,
    onMetricChange,
    chartType,
    onChartTypeChange,
}: {
    title: string;
    desc: string;
    scope: OpsScope;
    metric: ChartMetric;
    onMetricChange: (metric: ChartMetric) => void;
    chartType: ChartType;
    onChartTypeChange: (type: ChartType) => void;
}) {
    const { data: entities = [] } = useOpsEntityList(scope, 18);
    const [selectedEntityKey, setSelectedEntityKey] = useState(DEFAULT_ENTITY_KEYS[scope]);
    const sortedEntities = useMemo(
        () => [...entities].sort((a, b) => metricValue(b, metric) - metricValue(a, metric)),
        [entities, metric],
    );

    useEffect(() => {
        if (scope === 'overall') {
            setSelectedEntityKey('all');
            return;
        }
        if (!sortedEntities.length) {
            setSelectedEntityKey('');
            return;
        }
        if (!sortedEntities.some((item) => item.entity_key === selectedEntityKey)) {
            setSelectedEntityKey(sortedEntities[0].entity_key);
        }
    }, [scope, selectedEntityKey, sortedEntities]);

    const { data: series = [] } = useOpsEntitySeries(scope, selectedEntityKey || (scope === 'overall' ? 'all' : ''));
    const { data: details = [] } = useOpsRecentDetails(scope, selectedEntityKey || (scope === 'overall' ? 'all' : ''), 12);

    const selectedEntity = useMemo(
        () => sortedEntities.find((item) => item.entity_key === selectedEntityKey) ?? null,
        [selectedEntityKey, sortedEntities],
    );

    const metrics: ChartMetric[] = scope === 'ip'
        ? ['success_rate', 'request_count', 'avg_latency_ms']
        : ['success_rate', 'request_count', 'avg_latency_ms', 'cache_hit_rate', 'cache_rate', 'cache_read_token', 'cache_write_token'];

    return (
        <section className="grid gap-5 xl:grid-cols-[minmax(280px,0.8fr)_minmax(0,1.2fr)]">
            <div className="space-y-4">
                <div className="rounded-3xl border border-card-border bg-card p-4">
                    <div className="text-lg font-semibold text-card-foreground">{title}</div>
                    <div className="mt-1 text-xs leading-5 text-muted-foreground">{desc}</div>
                </div>
                <div className="space-y-3">
                    {(scope === 'overall' ? [] : sortedEntities).map((item) => (
                        <EntityCard
                            key={item.entity_key}
                            item={item}
                            metric={metric}
                            active={item.entity_key === selectedEntityKey}
                            onClick={() => setSelectedEntityKey(item.entity_key)}
                        />
                    ))}
                    {scope !== 'overall' && !sortedEntities.length ? (
                        <div className="rounded-2xl border border-dashed border-card-border bg-background/35 px-4 py-6 text-sm text-muted-foreground">
                            当前 12 小时内还没有这个维度的统计数据。
                        </div>
                    ) : null}
                </div>
            </div>
            <div className="space-y-4">
                <div className="rounded-3xl border border-card-border bg-card p-4">
                    <div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
                        <div className="space-y-2">
                            <div className="text-sm text-muted-foreground">当前对象</div>
                            <div className="text-xl font-semibold text-card-foreground">
                                {scope === 'overall' ? '全局成功率' : (selectedEntity?.entity_label || selectedEntityKey || '暂无对象')}
                            </div>
                            {selectedEntity ? (
                                <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
                                    <span>{selectedEntity.success_count + selectedEntity.failure_count} 请求</span>
                                    <span>成功率 {percent(selectedEntity.success_rate)}</span>
                                    <span>缓存率 {percent(selectedEntity.cache_rate)}</span>
                                    <span>平均延迟 {duration(selectedEntity.avg_latency_ms)}</span>
                                </div>
                            ) : null}
                        </div>
                        <div className="flex flex-col gap-2">
                            <div className="flex flex-wrap gap-2">
                                {metrics.map((item) => (
                                    <button
                                        key={item}
                                        type="button"
                                        className={cn(
                                            'rounded-xl border px-3 py-1.5 text-sm transition',
                                            metric === item ? 'border-primary/30 bg-primary/10 text-primary' : 'border-border/60 bg-background/70 text-muted-foreground hover:text-foreground',
                                        )}
                                        onClick={() => onMetricChange(item)}
                                    >
                                        {metricLabel(item)}
                                    </button>
                                ))}
                            </div>
                            <div className="flex flex-wrap gap-2 xl:justify-end">
                                {CHART_TYPES.map((item) => {
                                    const Icon = item.icon;
                                    return (
                                        <button
                                            key={item.key}
                                            type="button"
                                            className={cn(
                                                'inline-flex items-center gap-1.5 rounded-xl border px-3 py-1.5 text-sm transition',
                                                chartType === item.key ? 'border-primary/30 bg-primary/10 text-primary' : 'border-border/60 bg-background/70 text-muted-foreground hover:text-foreground',
                                            )}
                                            onClick={() => onChartTypeChange(item.key)}
                                        >
                                            <Icon className="size-4" />
                                            {item.label}
                                        </button>
                                    );
                                })}
                            </div>
                        </div>
                    </div>
                    <div className="mt-4 rounded-3xl border border-card-border bg-background/40 p-3">
                        <SeriesChart data={series} metric={metric} chartType={chartType} />
                    </div>
                </div>

                <div className="rounded-3xl border border-card-border bg-card p-4">
                    <div className="flex items-center gap-2 text-lg font-semibold text-card-foreground">
                        <Waypoints className="size-4.5 text-primary" />
                        最近调用明细
                    </div>
                    <div className="mt-1 text-xs leading-5 text-muted-foreground">只展示最近 12 小时内与当前维度相关的调用明细，用于二级排查。</div>
                    <div className="mt-4">
                        <DetailsTable rows={details} />
                    </div>
                </div>
            </div>
        </section>
    );
}

export function Ops() {
    const { data: overview } = useOpsOverview();
    const [tab, setTab] = useState<PrimaryTab>('overview');
    const [keyScope, setKeyScope] = useState<OpsScope>('channel_key');
    const [cacheScope, setCacheScope] = useState<OpsScope>('model');
    const [metric, setMetric] = useState<ChartMetric>('success_rate');
    const [chartType, setChartType] = useState<ChartType>('line');

    useEffect(() => {
        if (tab === 'cache') {
            setMetric('cache_hit_rate');
            return;
        }
        setMetric('success_rate');
    }, [tab]);

    const total = overview?.total;
    const totalRequests = total ? total.success_count + total.failure_count : 0;

    return (
        <PageWrapper data-testid="ops-page" className="h-full min-h-0 overflow-y-auto overscroll-contain space-y-5 pb-24 md:pb-4 rounded-t-3xl md:space-y-6">
            <section className="rounded-3xl border border-card-border bg-card p-5 sm:p-6">
                <div className="flex flex-col gap-5 xl:flex-row xl:items-start xl:justify-between">
                    <div className="max-w-4xl">
                        <div className="inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-[11px] text-primary">
                            <ShieldCheck className="size-3.5" />
                            独立运维工作台
                        </div>
                        <div className="mt-3 text-2xl font-semibold tracking-tight text-card-foreground sm:text-[2rem]">成功率、缓存与调用观测</div>
                        <div className="mt-2 text-sm leading-6 text-muted-foreground">统一观察总成功率、模型/渠道/双 Key 成功率、缓存命中与缓存率，以及近 12 小时 IP 调用情况。</div>
                    </div>
                    <div className="rounded-2xl border border-card-border bg-background/55 px-4 py-3 text-sm text-muted-foreground">
                        聚合窗口：最近 12 小时
                        <div className="mt-1 text-xs">5 分钟聚合桶 + 最近明细下钻</div>
                    </div>
                </div>

                <div className="mt-5 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
                    <SummaryCard title="总成功率" value={total ? percent(total.success_rate) : '-'} hint={`${shortNumber(totalRequests)} 请求`} icon={ShieldCheck} />
                    <SummaryCard title="缓存命中率" value={total ? percent(total.cache_hit_rate) : '-'} hint={`读取 ${shortNumber(total?.cache_read_token ?? 0)}`} icon={Database} />
                    <SummaryCard title="缓存率" value={total ? percent(total.cache_rate) : '-'} hint={`写入 ${shortNumber(total?.cache_write_token ?? 0)}`} icon={Layers3} />
                    <SummaryCard title="平均延迟" value={total ? duration(total.avg_latency_ms) : '-'} hint={`成功 ${shortNumber(total?.success_count ?? 0)} / 失败 ${shortNumber(total?.failure_count ?? 0)}`} icon={Activity} />
                </div>
            </section>

            <Tabs value={tab} onValueChange={(value) => setTab(value as PrimaryTab)}>
                <TabsList className="w-full justify-start overflow-x-auto bg-background/55">
                    {PRIMARY_TABS.map((item) => {
                        const Icon = item.icon;
                        return (
                            <TabsTrigger key={item.key} value={item.key}>
                                <Icon className="size-4" />
                                {item.label}
                            </TabsTrigger>
                        );
                    })}
                </TabsList>

                <TabsContents className="mt-4">
                    <TabsContent value="overview">
                        <section className="grid gap-5 xl:grid-cols-2">
                            <div className="rounded-3xl border border-card-border bg-card p-4">
                                <div className="text-lg font-semibold text-card-foreground">热点维度</div>
                                <div className="mt-1 text-xs text-muted-foreground">先看过去 12 小时最活跃的模型、渠道、Key 和 IP，再进入下方二级页签细查。</div>
                                <div className="mt-4 grid gap-3 sm:grid-cols-2">
                                    {[{
                                        title: '模型',
                                        items: overview?.top_models ?? [],
                                    }, {
                                        title: '渠道',
                                        items: overview?.top_channels ?? [],
                                    }, {
                                        title: '渠道 Key',
                                        items: overview?.top_channel_keys ?? [],
                                    }, {
                                        title: '访问 IP',
                                        items: overview?.top_ips ?? [],
                                    }].map((group) => (
                                        <div key={group.title} className="rounded-2xl border border-card-border bg-background/40 p-3">
                                            <div className="mb-3 text-sm font-medium text-card-foreground">{group.title}</div>
                                            <div className="space-y-2">
                                                {group.items.slice(0, 4).map((item) => (
                                                    <div key={item.entity_key} className="flex items-center justify-between gap-3 rounded-xl border border-card-border/60 bg-background/60 px-3 py-2">
                                                        <div className="min-w-0 truncate text-sm text-card-foreground">{item.entity_label || item.entity_key}</div>
                                                        <div className="text-xs text-muted-foreground">{percent(item.success_rate)}</div>
                                                    </div>
                                                ))}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>

                            <ScopeWorkspace
                                title="全局趋势"
                                desc="全局趋势按请求级成功率、缓存率和延迟聚合，可直接切换折线、柱状和点状图。"
                                scope="overall"
                                metric={metric}
                                onMetricChange={setMetric}
                                chartType={chartType}
                                onChartTypeChange={setChartType}
                            />
                        </section>
                    </TabsContent>

                    <TabsContent value="model">
                        <ScopeWorkspace
                            title="模型成功率"
                            desc="按模型查看成功率、缓存命中、缓存率和延迟变化。"
                            scope="model"
                            metric={metric}
                            onMetricChange={setMetric}
                            chartType={chartType}
                            onChartTypeChange={setChartType}
                        />
                    </TabsContent>

                    <TabsContent value="channel">
                        <ScopeWorkspace
                            title="渠道成功率"
                            desc="按渠道查看尝试结果。失败重试会单独计入对应渠道，便于识别不稳定节点。"
                            scope="channel"
                            metric={metric}
                            onMetricChange={setMetric}
                            chartType={chartType}
                            onChartTypeChange={setChartType}
                        />
                    </TabsContent>

                    <TabsContent value="key">
                        <div className="mb-4 flex flex-wrap gap-2">
                            {([
                                { key: 'channel_key', label: '渠道 Key' },
                                { key: 'api_key', label: '分发 Key' },
                            ] as const).map((item) => (
                                <button
                                    key={item.key}
                                    type="button"
                                    className={cn(
                                        'rounded-xl border px-3 py-1.5 text-sm transition',
                                        keyScope === item.key ? 'border-primary/30 bg-primary/10 text-primary' : 'border-border/60 bg-background/70 text-muted-foreground hover:text-foreground',
                                    )}
                                    onClick={() => setKeyScope(item.key)}
                                >
                                    {item.label}
                                </button>
                            ))}
                        </div>
                        <ScopeWorkspace
                            title={keyScope === 'channel_key' ? '渠道 Key 成功率' : '分发 Key 成功率'}
                            desc={keyScope === 'channel_key' ? '按渠道内部 Key 观察成功率与缓存命中。' : '按用户分发 Key 观察调用成功率、缓存率与延迟。'}
                            scope={keyScope}
                            metric={metric}
                            onMetricChange={setMetric}
                            chartType={chartType}
                            onChartTypeChange={setChartType}
                        />
                    </TabsContent>

                    <TabsContent value="cache">
                        <div className="mb-4 flex flex-wrap gap-2">
                            {([
                                { key: 'model', label: '模型缓存' },
                                { key: 'channel', label: '渠道缓存' },
                                { key: 'channel_key', label: '渠道 Key 缓存' },
                                { key: 'api_key', label: '分发 Key 缓存' },
                            ] as const).map((item) => (
                                <button
                                    key={item.key}
                                    type="button"
                                    className={cn(
                                        'rounded-xl border px-3 py-1.5 text-sm transition',
                                        cacheScope === item.key ? 'border-primary/30 bg-primary/10 text-primary' : 'border-border/60 bg-background/70 text-muted-foreground hover:text-foreground',
                                    )}
                                    onClick={() => setCacheScope(item.key)}
                                >
                                    {item.label}
                                </button>
                            ))}
                        </div>
                        <ScopeWorkspace
                            title="缓存命中与缓存率"
                            desc="按维度查看缓存命中率、缓存率、缓存读取和缓存写入变化。"
                            scope={cacheScope}
                            metric={metric}
                            onMetricChange={setMetric}
                            chartType={chartType}
                            onChartTypeChange={setChartType}
                        />
                    </TabsContent>

                    <TabsContent value="ip">
                        <ScopeWorkspace
                            title="IP 访问统计"
                            desc="按 IP 查看调用成功率、请求量和延迟。该统计从请求链路底部采样并进入 12 小时聚合桶。"
                            scope="ip"
                            metric={metric}
                            onMetricChange={setMetric}
                            chartType={chartType}
                            onChartTypeChange={setChartType}
                        />
                    </TabsContent>
                </TabsContents>
            </Tabs>
        </PageWrapper>
    );
}
