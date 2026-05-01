'use client';

import { useChannelList } from '@/api/endpoints/channel';
import { useMemo } from 'react';
import { useTranslations } from 'next-intl';
import { TrendingUp } from 'lucide-react';
import { Tabs, TabsList, TabsTrigger, TabsContents, TabsContent } from '@/components/animate-ui/components/animate/tabs';
import { useStatsTokenBreakdown } from '@/api/endpoints/stats';

type SortMode = 'cost' | 'count';
type ChannelData = NonNullable<ReturnType<typeof useChannelList>['data']>[number];

export function Rank() {
    const { data: channelData } = useChannelList();
    const { data: tokenBreakdown } = useStatsTokenBreakdown();
    const t = useTranslations('home.rank');

    const tokenByChannel = useMemo(() => {
        const map = new Map<number, { value: string; unit: string }>();
        for (const item of tokenBreakdown?.by_channel ?? []) {
            const matched = item.key.match(/^channel:(\d+)$/);
            if (!matched) continue;
            const id = Number.parseInt(matched[1], 10);
            if (!Number.isFinite(id)) continue;
            map.set(id, {
                value: item.total_token.formatted.value,
                unit: item.total_token.formatted.unit,
            });
        }
        return map;
    }, [tokenBreakdown?.by_channel]);

    const rankedByCost = useMemo<ChannelData[]>(() => {
        if (!channelData) return [];
        return [...channelData].sort((a, b) => b.formatted.total_cost.raw - a.formatted.total_cost.raw);
    }, [channelData]);

    const rankedByCount = useMemo<ChannelData[]>(() => {
        if (!channelData) return [];
        return [...channelData].sort((a, b) => b.formatted.request_count.raw - a.formatted.request_count.raw);
    }, [channelData]);

    const getMedalEmoji = (rank: number): string => {
        switch (rank) {
            case 1: return '🥇';
            case 2: return '🥈';
            case 3: return '🥉';
            default: return '';
        }
    };

    const renderList = (channels: ChannelData[], mode: SortMode) => {
        if (channels.length === 0) {
            return (
                <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
                    <TrendingUp className="w-12 h-12 mb-3 opacity-30" />
                    <p className="text-sm">{t('noData')}</p>
                </div>
            );
        }
        return (
            <div className="space-y-2.5 max-h-[340px] overflow-y-auto">
                {channels.map((channel, index) => {
                    const rank = index + 1;
                    const medal = getMedalEmoji(rank);
                    const channelToken = tokenByChannel.get(channel.raw.id);

                    return (
                        <div
                            key={channel.raw.id}
                            className="rounded-2xl border border-border/50 bg-muted/20 px-3 py-2.5 transition-colors hover:bg-accent/5"
                        >
                            <div className="grid grid-cols-[auto,minmax(0,1fr),auto] items-center gap-3">
                                <div className="w-8 h-8 rounded-lg flex items-center justify-center font-bold text-lg shrink-0">
                                    {medal || rank}
                                </div>

                                <div className="min-w-0">
                                    <p className="font-medium text-sm truncate">{channel.raw.name}</p>
                                    <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] text-muted-foreground">
                                        {mode === 'count' ? (() => {
                                            const successCount = channel.formatted.request_success.raw;
                                            const failedCount = channel.formatted.request_failed.raw;
                                            const totalCount = successCount + failedCount;
                                            const successRate = totalCount > 0 ? (successCount / totalCount) * 100 : 0;

                                            return <span>{t('successRate')}: {successRate.toFixed(1)}%</span>;
                                        })() : <span>{t('requestCount')}: {channel.formatted.request_count.formatted.value}{channel.formatted.request_count.formatted.unit}</span>}
                                        <span>{t('channelUsage')}: {channelToken ? `${channelToken.value}${channelToken.unit}` : '--'}</span>
                                    </div>
                                </div>

                                <div className="flex items-center gap-1 text-right shrink-0">
                                    {mode === 'count' ? (
                                        <div className="flex items-center gap-1 text-sm font-medium tabular-nums">
                                            <span className="text-accent">
                                                {channel.formatted.request_success.formatted.value}
                                                <span className="text-xs text-muted-foreground">
                                                    {channel.formatted.request_success.formatted.unit}
                                                </span>
                                            </span>
                                            <span className="text-muted-foreground/40 font-light">/</span>
                                            <span className="text-destructive">
                                                {channel.formatted.request_failed.formatted.value}
                                                <span className="text-xs text-muted-foreground">
                                                    {channel.formatted.request_failed.formatted.unit}
                                                </span>
                                            </span>
                                        </div>
                                    ) : (
                                        <span className="font-semibold text-base">
                                            {channel.formatted.total_cost.formatted.value}
                                            <span className="text-xs text-muted-foreground">
                                                {channel.formatted.total_cost.formatted.unit}
                                            </span>
                                        </span>
                                    )}
                                </div>
                            </div>
                        </div>
                    );
                })}
            </div>
        );
    };

    return (
        <div data-testid="home-rank-section" className="rounded-3xl bg-card text-card-foreground border-card-border border p-4">
            <Tabs defaultValue="cost">
                <div className="flex items-center justify-between">
                    <h3 className="font-semibold text-base">{t('title')}</h3>
                    <TabsList>
                        <TabsTrigger value="cost">{t('sortByCost')}</TabsTrigger>
                        <TabsTrigger value="count">{t('sortByCount')}</TabsTrigger>
                    </TabsList>
                </div>
                <TabsContents>
                    <TabsContent value="cost">
                        {renderList(rankedByCost, 'cost')}
                    </TabsContent>
                    <TabsContent value="count">
                        {renderList(rankedByCount, 'count')}
                    </TabsContent>
                </TabsContents>
            </Tabs>
        </div>
    );
}
