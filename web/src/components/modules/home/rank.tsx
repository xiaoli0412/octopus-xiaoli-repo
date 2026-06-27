'use client';

import { useMemo } from 'react';
import { Award, Medal, TrendingUp } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { useChannelList } from '@/api/endpoints/channel';
import { useStatsTokenBreakdown } from '@/api/endpoints/stats';
import { AnimatedNumber } from '@/components/common/AnimatedNumber';
import { Tabs, TabsContent, TabsContents, TabsList, TabsTrigger } from '@/components/animate-ui/components/animate/tabs';
import { cn } from '@/lib/utils';

type RankedChannel = {
	id: number;
	name: string;
	summary: string;
	value: string;
	unit?: string;
};

function rankIcon(rank: number) {
	if (rank === 1) return <Award className="size-4 text-[color:var(--chart-2)]" />;
	if (rank === 2) return <Medal className="size-4 text-[color:var(--chart-4)]" />;
	if (rank === 3) return <Medal className="size-4 text-[color:var(--chart-5)]" />;
	return <span className="text-base font-semibold tabular-nums text-muted-foreground">{rank}</span>;
}

export function Rank() {
	const { data: channelData } = useChannelList();
	const { data: tokenBreakdown } = useStatsTokenBreakdown();
	const t = useTranslations('home.rank');

	const tokenByChannel = useMemo(() => {
		const map = new Map<number, { value: string; unit: string; raw: number }>();
		for (const item of tokenBreakdown?.by_channel ?? []) {
			const matched = item.key.match(/^channel:(\d+)$/);
			if (!matched) continue;
			const id = Number.parseInt(matched[1], 10);
			if (!Number.isFinite(id)) continue;
			map.set(id, {
				value: item.total_token.formatted.value,
				unit: item.total_token.formatted.unit,
				raw: item.total_token.raw,
			});
		}
		return map;
	}, [tokenBreakdown?.by_channel]);

	const costItems = useMemo<RankedChannel[]>(() => {
		if (!channelData) return [];
		return [...channelData]
			.sort((a, b) => b.formatted.total_cost.raw - a.formatted.total_cost.raw)
			.slice(0, 8)
			.map((channel) => ({
				id: channel.raw.id,
				name: channel.raw.name,
				summary: `${t('requestCount')} ${channel.formatted.request_count.formatted.value}${channel.formatted.request_count.formatted.unit}`,
				value: channel.formatted.total_cost.formatted.value,
				unit: channel.formatted.total_cost.formatted.unit,
			}));
	}, [channelData, t]);

	const countItems = useMemo<RankedChannel[]>(() => {
		if (!channelData) return [];
		return [...channelData]
			.sort((a, b) => b.formatted.request_count.raw - a.formatted.request_count.raw)
			.slice(0, 8)
			.map((channel) => {
				const successCount = channel.formatted.request_success.raw;
				const failedCount = channel.formatted.request_failed.raw;
				const totalCount = successCount + failedCount;
				const successRate = totalCount > 0 ? (successCount / totalCount) * 100 : 0;
				return {
					id: channel.raw.id,
					name: channel.raw.name,
					summary: `${t('successRate')} ${successRate.toFixed(1)}%`,
					value: channel.formatted.request_count.formatted.value,
					unit: channel.formatted.request_count.formatted.unit,
				};
			});
	}, [channelData, t]);

	const tokenItems = useMemo<RankedChannel[]>(() => {
		if (!channelData) return [];
		return [...channelData]
			.sort((a, b) => (tokenByChannel.get(b.raw.id)?.raw ?? 0) - (tokenByChannel.get(a.raw.id)?.raw ?? 0))
			.slice(0, 8)
			.map((channel) => {
				const token = tokenByChannel.get(channel.raw.id);
				return {
					id: channel.raw.id,
					name: channel.raw.name,
					summary: `${t('requestCount')} ${channel.formatted.request_count.formatted.value}${channel.formatted.request_count.formatted.unit}`,
					value: token?.value ?? '0.00',
					unit: token?.unit,
				};
			});
	}, [channelData, t, tokenByChannel]);

	function renderList(items: RankedChannel[], metricLabel: string) {
		if (items.length === 0) {
			return (
				<div className="flex min-h-[17rem] flex-col items-center justify-center rounded-3xl border border-dashed border-border/60 bg-background/40 text-muted-foreground">
					<TrendingUp className="mb-3 size-10 opacity-30" />
					<div className="text-sm">{t('noData')}</div>
				</div>
			);
		}

		return (
			<div className="max-h-[23rem] overflow-y-auto pr-1">
				<div data-testid="home-rank-list" className="grid gap-2.5 sm:grid-cols-2 xl:grid-cols-3">
					{items.map((item, index) => {
						const rank = index + 1;
						return (
							<article
								key={item.id}
								data-testid={`home-rank-card-${rank}`}
								className="h-full rounded-2xl border border-border/60 bg-background/45 px-3.5 py-3 transition hover:border-primary/20 hover:bg-background/72"
							>
								<div className="flex h-full flex-col gap-3">
									<div className="flex items-start justify-between gap-3">
										<div className="flex min-w-0 items-start gap-3">
											<div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl border border-border/50 bg-card/70">
												{rankIcon(rank)}
											</div>
											<div className="min-w-0">
												<div className="break-all text-sm font-medium leading-5 text-foreground">{item.name}</div>
												<div className="mt-1 line-clamp-2 text-[11px] leading-4 text-muted-foreground">{item.summary}</div>
											</div>
										</div>
										<div className="shrink-0 rounded-full border border-border/50 bg-card/75 px-2 py-0.5 text-[10px] font-medium text-muted-foreground">
											TOP {rank}
										</div>
									</div>
									<div className="flex items-end justify-between gap-3 border-t border-border/50 pt-2.5">
										<div className="text-[11px] text-muted-foreground">{metricLabel}</div>
										<div className="text-right">
											<div className={cn('text-lg font-semibold tabular-nums leading-none text-foreground', rank <= 3 ? 'text-primary' : undefined)}>
												<AnimatedNumber value={item.value} />
												{item.unit ? <span className="ml-1 text-xs text-muted-foreground">{item.unit}</span> : null}
											</div>
										</div>
									</div>
								</div>
							</article>
						);
					})}
				</div>
			</div>
		);
	}

	return (
		<section data-testid="home-rank-section" className="rounded-3xl border border-card-border bg-card p-4 text-card-foreground custom-shadow sm:p-5">
			<Tabs defaultValue="cost">
				<div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
					<div>
						<div className="text-lg font-semibold text-foreground">{t('title')}</div>
						<div className="mt-1 text-xs leading-5 text-muted-foreground">{t('subtitle')}</div>
					</div>
					<TabsList className="w-full justify-start bg-background/60 sm:w-fit">
						<TabsTrigger value="cost">{t('sortByCost')}</TabsTrigger>
						<TabsTrigger value="count">{t('sortByCount')}</TabsTrigger>
						<TabsTrigger value="token">{t('sortByToken')}</TabsTrigger>
					</TabsList>
				</div>

				<TabsContents className="mt-4">
					<TabsContent value="cost">{renderList(costItems, t('sortByCost'))}</TabsContent>
					<TabsContent value="count">{renderList(countItems, t('sortByCount'))}</TabsContent>
					<TabsContent value="token">{renderList(tokenItems, t('sortByToken'))}</TabsContent>
				</TabsContents>
			</Tabs>
		</section>
	);
}
