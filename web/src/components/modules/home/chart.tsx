'use client';

import { useMemo, useState } from 'react';
import { useTranslations } from 'next-intl';
import dayjs from 'dayjs';
import {
	Area,
	AreaChart,
	Bar,
	BarChart,
	CartesianGrid,
	Cell,
	Line,
	LineChart,
	XAxis,
	YAxis,
} from 'recharts';
import { Activity, BarChart3, ChartColumn, Flame } from 'lucide-react';

import { useStatsDaily, useStatsHourly, useStatsTokenBreakdown } from '@/api/endpoints/stats';
import { AnimatedNumber } from '@/components/common/AnimatedNumber';
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart';
import { cn } from '@/lib/utils';
import { formatCount, formatMoney } from '@/lib/utils';

const PERIODS = ['1', '7', '30'] as const;

type TrendMetric = 'cost' | 'count' | 'token';
type TrendView = 'area' | 'line' | 'bar' | 'heatmap';

type TrendPoint = {
	date: string;
	cost: number;
	count: number;
	token: number;
};

type HeatPoint = TrendPoint & {
	heat: number;
	tone: string;
};

function getMetricLabel(metric: TrendMetric, t: ReturnType<typeof useTranslations<'home.chart'>>) {
	switch (metric) {
		case 'count':
			return t('totalRequests');
		case 'token':
			return t('tokenUsage');
		default:
			return t('totalCost');
	}
}

function formatMetricValue(metric: TrendMetric, value: number) {
	if (metric === 'cost') return formatMoney(value).formatted;
	return formatCount(value).formatted;
}

function metricColor(metric: TrendMetric) {
	switch (metric) {
		case 'count':
			return 'var(--chart-2)';
		case 'token':
			return 'var(--chart-3)';
		default:
			return 'var(--chart-1)';
	}
}

function heatTone(value: number, max: number) {
	if (!max || value <= 0) return 'color-mix(in oklab, var(--muted) 75%, transparent)';
	const ratio = value / max;
	if (ratio > 0.8) return 'color-mix(in oklab, var(--chart-1) 82%, transparent)';
	if (ratio > 0.6) return 'color-mix(in oklab, var(--chart-2) 72%, transparent)';
	if (ratio > 0.4) return 'color-mix(in oklab, var(--chart-3) 68%, transparent)';
	if (ratio > 0.2) return 'color-mix(in oklab, var(--chart-4) 58%, transparent)';
	return 'color-mix(in oklab, var(--chart-5) 46%, transparent)';
}

function HeatmapPanel({ data, metric, t }: { data: HeatPoint[]; metric: TrendMetric; t: ReturnType<typeof useTranslations<'home.chart'>> }) {
	const maxValue = data.reduce((value, point) => Math.max(value, point[metric]), 0);
	const formattedPeak = formatMetricValue(metric, maxValue);

	return (
		<div className="space-y-3">
			<div className="flex flex-col gap-2.5 sm:flex-row sm:items-center sm:justify-between">
				<div className="space-y-1">
					<div className="text-[11px] text-muted-foreground">{t('view.heatmap')}</div>
					<div className="text-sm font-semibold text-foreground">{getMetricLabel(metric, t)}</div>
				</div>
				<div className="inline-flex items-center gap-2 self-start rounded-full border border-border/60 bg-background/80 px-2.5 py-1.5 text-[10px] text-muted-foreground">
					<div className="flex items-center gap-1">
						<span className="h-2.5 w-2.5 rounded-full border border-border/50 bg-muted/70" />
						<span className="h-2.5 w-2.5 rounded-full border border-border/40" style={{ background: heatTone(maxValue * 0.35, maxValue) }} />
						<span className="h-2.5 w-2.5 rounded-full border border-border/40" style={{ background: heatTone(maxValue * 0.7, maxValue) }} />
						<span className="h-2.5 w-2.5 rounded-full border border-border/40" style={{ background: heatTone(maxValue, maxValue) }} />
					</div>
					<span className="font-medium text-foreground">{formattedPeak.value}</span>
					{formattedPeak.unit ? <span>{formattedPeak.unit}</span> : null}
				</div>
			</div>
			<div className="grid gap-2 sm:grid-cols-4 md:grid-cols-6 xl:grid-cols-8 2xl:grid-cols-10">
				{data.map((point) => {
					const formatted = formatMetricValue(metric, point[metric]);
					const isPeak = maxValue > 0 && point[metric] === maxValue;

					return (
						<div
							key={point.date}
							className={cn(
								'relative overflow-hidden rounded-xl border px-2.5 py-2.5 transition-colors',
								isPeak
									? 'border-primary/30 bg-background/88 shadow-[0_10px_24px_-18px_color-mix(in_oklab,var(--primary)_75%,transparent)]'
									: 'border-border/55 bg-background/78',
							)}
							style={{
								backgroundImage:
									point.heat > 0
										? `linear-gradient(180deg, ${point.tone} 0%, color-mix(in oklab, var(--background) 95%, transparent) 74%)`
										: undefined,
							}}
						>
							<div className="absolute inset-x-0 top-0 h-1.5 opacity-90" style={{ background: point.tone }} />
							<div className="flex items-start justify-between gap-2">
								<div className="text-[10px] font-medium text-muted-foreground">{point.date}</div>
								<span className="mt-0.5 h-2.5 w-2.5 shrink-0 rounded-full border border-white/35" style={{ background: point.tone }} />
							</div>
							<div className="mt-2 flex items-end gap-1 text-foreground">
								<span className="text-sm font-semibold leading-none tabular-nums">{formatted.value}</span>
								{formatted.unit ? <span className="text-[10px] text-muted-foreground">{formatted.unit}</span> : null}
							</div>
						</div>
					);
				})}
			</div>
		</div>
	);
}

function ChartHeaderCard({ title, value, unit }: { title: string; value: string | number; unit?: string }) {
	return (
		<div className="rounded-2xl border border-border/60 bg-background/55 px-3.5 py-2.5">
			<div className="text-[11px] text-muted-foreground">{title}</div>
			<div className="mt-1 text-[1.05rem] font-semibold text-foreground tabular-nums">
				<AnimatedNumber value={value} />
				{unit ? <span className="ml-1 text-xs text-muted-foreground">{unit}</span> : null}
			</div>
		</div>
	);
}

export function StatsChart() {
	const { data: statsDaily } = useStatsDaily();
	const { data: statsHourly } = useStatsHourly();
	const { data: tokenBreakdown } = useStatsTokenBreakdown();
	const t = useTranslations('home.chart');
	const [period, setPeriod] = useState<typeof PERIODS[number]>('1');
	const [metric, setMetric] = useState<TrendMetric>('cost');
	const [view, setView] = useState<TrendView>('area');
	const isHeatmapView = view === 'heatmap';

	const sortedDaily = useMemo(() => {
		if (!statsDaily) return [];
		return [...statsDaily].sort((a, b) => a.date.localeCompare(b.date));
	}, [statsDaily]);

	const chartData = useMemo<TrendPoint[]>(() => {
		if (period === '1') {
			if (!statsHourly) return [];
			return statsHourly.map((stat) => ({
				date: `${stat.hour}:00`,
				cost: stat.total_cost.raw,
				count: stat.request_count.raw,
				token: stat.total_token.raw,
			}));
		}

		const days = Number.parseInt(period, 10);
		return sortedDaily.slice(-days).map((stat) => ({
			date: dayjs(stat.date).format('MM-DD'),
			cost: stat.total_cost.raw,
			count: stat.request_count.raw,
			token: stat.total_token.raw,
		}));
	}, [period, sortedDaily, statsHourly]);

	const totals = useMemo(() => {
		return chartData.reduce(
			(acc, point) => ({
				cost: acc.cost + point.cost,
				count: acc.count + point.count,
				token: acc.token + point.token,
			}),
			{ cost: 0, count: 0, token: 0 },
		);
	}, [chartData]);

	const peakValue = useMemo(() => chartData.reduce((max, point) => Math.max(max, point[metric]), 0), [chartData, metric]);
	const avgValue = chartData.length > 0 ? totals[metric] / chartData.length : 0;
	const latestValue = chartData.length > 0 ? chartData[chartData.length - 1][metric] : 0;
	const formattedTotal = formatMetricValue(metric, totals[metric]);
	const formattedPeak = formatMetricValue(metric, peakValue);
	const formattedAvg = formatMetricValue(metric, avgValue);
	const formattedLatest = formatMetricValue(metric, latestValue);

	const heatmapData = useMemo<HeatPoint[]>(() => {
		const max = chartData.reduce((value, point) => Math.max(value, point[metric]), 0);
		return chartData.map((point) => ({
			...point,
			heat: point[metric],
			tone: heatTone(point[metric], max),
		}));
	}, [chartData, metric]);

	const chartConfig = {
		metric: { label: getMetricLabel(metric, t), color: metricColor(metric) },
	};

	const chartNode = useMemo(() => {
		if (view === 'area') {
			return (
				<AreaChart accessibilityLayer data={chartData} margin={{ left: 8, right: 8, top: 8, bottom: 0 }}>
					<defs>
						<linearGradient id="home-trend-area" x1="0" y1="0" x2="0" y2="1">
							<stop offset="5%" stopColor={metricColor(metric)} stopOpacity={0.88} />
							<stop offset="95%" stopColor={metricColor(metric)} stopOpacity={0.08} />
						</linearGradient>
					</defs>
					<CartesianGrid strokeDasharray="3 3" vertical={false} />
					<XAxis dataKey="date" tickLine={false} axisLine={false} />
					<YAxis tickLine={false} axisLine={false} tickFormatter={(value) => `${formatMetricValue(metric, value).value}${formatMetricValue(metric, value).unit}`} />
					<ChartTooltip cursor={false} content={<ChartTooltipContent indicator="line" />} />
					<Area type="monotone" dataKey={metric} stroke={metricColor(metric)} strokeWidth={2} fill="url(#home-trend-area)" />
				</AreaChart>
			);
		}

		if (view === 'line') {
			return (
				<LineChart accessibilityLayer data={chartData} margin={{ left: 8, right: 8, top: 8, bottom: 0 }}>
					<CartesianGrid strokeDasharray="3 3" vertical={false} />
					<XAxis dataKey="date" tickLine={false} axisLine={false} />
					<YAxis tickLine={false} axisLine={false} tickFormatter={(value) => `${formatMetricValue(metric, value).value}${formatMetricValue(metric, value).unit}`} />
					<ChartTooltip cursor={false} content={<ChartTooltipContent indicator="line" />} />
					<Line type="monotone" dataKey={metric} stroke={metricColor(metric)} strokeWidth={2.2} dot={{ r: 3.5, fill: metricColor(metric) }} activeDot={{ r: 5 }} />
				</LineChart>
			);
		}

		return (
			<BarChart accessibilityLayer data={chartData} margin={{ left: 8, right: 8, top: 8, bottom: 0 }}>
				<CartesianGrid strokeDasharray="3 3" vertical={false} />
				<XAxis dataKey="date" tickLine={false} axisLine={false} />
				<YAxis tickLine={false} axisLine={false} tickFormatter={(value) => `${formatMetricValue(metric, value).value}${formatMetricValue(metric, value).unit}`} />
				<ChartTooltip cursor={false} content={<ChartTooltipContent indicator="line" />} />
				<Bar dataKey={metric} radius={[8, 8, 0, 0]}>
					{chartData.map((entry) => (
						<Cell key={entry.date} fill={entry[metric] === peakValue && peakValue > 0 ? 'var(--chart-2)' : metricColor(metric)} />
					))}
				</Bar>
			</BarChart>
		);
	}, [chartData, metric, peakValue, view]);

	const chartHeight = 'h-[18rem]';
	const channelUsage = tokenBreakdown?.by_channel.length ?? 0;

	const periodLabel = period === '1' ? t('period.today') : period === '7' ? t('period.last7Days') : t('period.last30Days');

	return (
		<section data-testid="home-stats-chart-section" className="rounded-3xl border border-card-border bg-card p-4 text-card-foreground custom-shadow sm:p-5">
			<div className={cn('flex flex-col', isHeatmapView ? 'gap-3' : 'gap-4 xl:flex-row xl:items-start xl:justify-between')}>
				<div className="space-y-3">
					<div>
						<div className="text-lg font-semibold text-foreground">{t('title')}</div>
						<div className="mt-1 text-xs leading-5 text-muted-foreground">{t('subtitle')}</div>
					</div>
					<div className="grid gap-2.5 sm:grid-cols-2 xl:grid-cols-4">
						<ChartHeaderCard title={getMetricLabel(metric, t)} value={formattedTotal.value} unit={formattedTotal.unit} />
						<ChartHeaderCard title={t('peakValue')} value={formattedPeak.value} unit={formattedPeak.unit} />
						<ChartHeaderCard title={t('avgValue')} value={formattedAvg.value} unit={formattedAvg.unit} />
						<ChartHeaderCard title={t('latestValue')} value={formattedLatest.value} unit={formattedLatest.unit} />
					</div>
				</div>

				<div className={cn('flex flex-col gap-2.5', isHeatmapView ? '' : 'xl:items-end')}>
					<div className="flex flex-wrap gap-2 xl:justify-end">
						{([
							{ key: 'cost', label: t('metric.cost') },
							{ key: 'count', label: t('metric.count') },
							{ key: 'token', label: t('metric.token') },
						] as const).map((item) => (
							<button
								key={item.key}
								type="button"
								className={cn('rounded-xl border px-3 py-1.5 text-sm transition', metric === item.key ? 'border-primary/30 bg-primary/10 text-primary' : 'border-border/60 bg-background/70 text-muted-foreground hover:text-foreground')}
								onClick={() => setMetric(item.key)}
							>
								{item.label}
							</button>
						))}
					</div>
					<div className="flex flex-wrap gap-2 xl:justify-end">
						{([
							{ key: 'area', label: t('view.area'), icon: Activity },
							{ key: 'line', label: t('view.line'), icon: Flame },
							{ key: 'bar', label: t('view.bar'), icon: BarChart3 },
							{ key: 'heatmap', label: t('view.heatmap'), icon: ChartColumn },
						] as const).map((item) => {
							const Icon = item.icon;
							return (
								<button
									key={item.key}
									type="button"
									className={cn('inline-flex items-center gap-1.5 rounded-xl border px-3 py-1.5 text-sm transition', view === item.key ? 'border-primary/30 bg-primary/10 text-primary' : 'border-border/60 bg-background/70 text-muted-foreground hover:text-foreground')}
									onClick={() => setView(item.key)}
								>
									<Icon className="size-4" />
									{item.label}
								</button>
							);
						})}
					</div>
					<div className="grid gap-2.5 sm:grid-cols-3 xl:min-w-[18rem]">
						<ChartHeaderCard title={t('timePeriod')} value={periodLabel} />
						<ChartHeaderCard title={t('channelUsage')} value={channelUsage} />
						<ChartHeaderCard title={t('tokenUsage')} value={formatMetricValue('token', totals.token).value} unit={formatMetricValue('token', totals.token).unit} />
					</div>
					<div className="flex flex-wrap gap-2 xl:justify-end">
						{PERIODS.map((item) => (
							<button
								key={item}
								type="button"
								className={cn('rounded-xl border px-3 py-1.5 text-xs transition', period === item ? 'border-primary/30 bg-primary/10 text-primary' : 'border-border/60 bg-background/70 text-muted-foreground hover:text-foreground')}
								onClick={() => setPeriod(item)}
							>
								{item === '1' ? t('period.today') : item === '7' ? t('period.last7Days') : t('period.last30Days')}
							</button>
						))}
					</div>
				</div>
			</div>

			<div
				data-testid="home-stats-chart"
				className={cn(
					isHeatmapView
						? 'mt-3 rounded-[1.8rem] border border-border/60 bg-[linear-gradient(180deg,color-mix(in_oklab,var(--background)_94%,transparent),color-mix(in_oklab,var(--background)_84%,var(--card)))] p-3 sm:p-3.5'
						: 'mt-4 rounded-3xl border border-border/60 bg-background/45 p-3 sm:p-4',
				)}
			>
				{isHeatmapView ? (
					<HeatmapPanel data={heatmapData} metric={metric} t={t} />
				) : (
					<ChartContainer config={chartConfig} className={cn(chartHeight, 'w-full')}>
						{chartNode}
					</ChartContainer>
				)}
			</div>
		</section>
	);
}
