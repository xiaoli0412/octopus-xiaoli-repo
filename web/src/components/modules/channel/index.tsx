'use client';

import { useMemo, useState } from 'react';
import { ChannelType, useChannelList } from '@/api/endpoints/channel';
import { Card } from './Card';
import { useSearchStore, useToolbarViewOptionsStore } from '@/components/modules/toolbar';
import { VirtualizedGrid } from '@/components/common/VirtualizedGrid';
import { cn } from '@/lib/utils';
import { useNavStore } from '@/components/modules/navbar/nav-store';

function normalizeKeyword(value: string) {
    return value.trim().toLowerCase();
}

function getProviderFilterKey(type: ChannelType) {
    switch (type) {
        case ChannelType.OpenAIChat:
        case ChannelType.OpenAIResponse:
        case ChannelType.OpenAIEmbedding:
            return 'openai';
        case ChannelType.Anthropic:
            return 'anthropic';
        case ChannelType.Gemini:
            return 'gemini';
        case ChannelType.Volcengine:
            return 'volcengine';
        case ChannelType.GithubCopilot:
            return 'github-copilot';
        case ChannelType.Antigravity:
            return 'antigravity';
        case ChannelType.Zen:
            return 'zen';
        default:
            return 'openai';
    }
}

function getChannelTypeSearchTokens(type: ChannelType) {
    switch (type) {
        case ChannelType.OpenAIChat:
            return ['openai', 'chat', 'openai chat', '聊天'];
        case ChannelType.OpenAIResponse:
            return ['openai', 'responses', 'response', '响应'];
        case ChannelType.OpenAIEmbedding:
            return ['openai', 'embedding', 'embeddings', '向量'];
        case ChannelType.Anthropic:
            return ['anthropic', 'claude'];
        case ChannelType.Gemini:
            return ['gemini', 'google'];
        case ChannelType.Volcengine:
            return ['volcengine', '火山', '火山引擎'];
        case ChannelType.GithubCopilot:
            return ['github', 'copilot', 'github copilot'];
        case ChannelType.Antigravity:
            return ['antigravity'];
        case ChannelType.Zen:
            return ['zen'];
        default:
            return [];
    }
}

export function Channel() {
    const { data: channelsData } = useChannelList();
    const setActiveItem = useNavStore((s) => s.setActiveItem);
    const pageKey = 'channel' as const;
    const searchTerm = useSearchStore((s) => s.getSearchTerm(pageKey));
    const channelDensity = useToolbarViewOptionsStore((s) => s.channelDensity);
    const sortOrder = useToolbarViewOptionsStore((s) => s.getSortOrder(pageKey));
    const filter = useToolbarViewOptionsStore((s) => s.channelFilter);
    const providerFilter = useToolbarViewOptionsStore((s) => s.channelProviderFilter);
    const modelKeyword = useToolbarViewOptionsStore((s) => s.channelModelKeyword);
    const keyKeyword = useToolbarViewOptionsStore((s) => s.channelKeyKeyword);

    const sortedChannels = useMemo(() => {
        if (!channelsData) return [];
        return [...channelsData].sort((a, b) => (sortOrder === 'asc' ? a.raw.id - b.raw.id : b.raw.id - a.raw.id));
    }, [channelsData, sortOrder]);

    const visibleChannels = useMemo(() => {
        const term = normalizeKeyword(searchTerm);
        const modelTerm = normalizeKeyword(modelKeyword);
        const keyTerm = normalizeKeyword(keyKeyword);

        const bySearch = !term
            ? sortedChannels
            : sortedChannels.filter((channel) => {
                const parts = [
                    channel.raw.name,
                    channel.raw.model,
                    channel.raw.custom_model,
                    ...getChannelTypeSearchTokens(channel.raw.type),
                    ...channel.raw.keys.flatMap((key) => [
                        key.channel_key,
                        key.remark,
                        key.source_type,
                        key.allowed_models,
                    ]),
                ];

                return parts.some((part) => part?.toLowerCase().includes(term));
            });

        const byStatus = filter === 'enabled'
            ? bySearch.filter((c) => c.raw.enabled)
            : filter === 'disabled'
                ? bySearch.filter((c) => !c.raw.enabled)
                : bySearch;

        const byProvider = providerFilter === 'all'
            ? byStatus
            : byStatus.filter((channel) => getProviderFilterKey(channel.raw.type) === providerFilter);

        const byModel = !modelTerm
            ? byProvider
            : byProvider.filter((channel) => {
                const models = [
                    channel.raw.model,
                    channel.raw.custom_model,
                    ...channel.raw.keys.map((key) => key.allowed_models ?? ''),
                ];
                return models.some((part) => part?.toLowerCase().includes(modelTerm));
            });

        const byKey = !keyTerm
            ? byModel
            : byModel.filter((channel) => {
                const keyParts = channel.raw.keys.flatMap((key) => [
                    key.channel_key,
                    key.remark,
                    key.source_type,
                    key.allowed_models,
                ]);
                return keyParts.some((part) => part?.toLowerCase().includes(keyTerm));
            });

        return byKey;
    }, [sortedChannels, searchTerm, filter, providerFilter, modelKeyword, keyKeyword]);

    return (
        <div data-testid="channel-page" data-layout="grid" data-density={channelDensity} className="flex h-full min-h-0 flex-col gap-3">
            <div className="flex shrink-0 items-center gap-2 rounded-2xl border border-border/60 bg-card/80 p-1.5">
                <div className="flex-1 px-2 text-xs font-medium text-muted-foreground">普通渠道列表</div>
                <button
                    type="button"
                    onClick={() => setActiveItem('upstream')}
                    className={cn(
                        'h-8 rounded-xl border border-border/60 bg-background/60 px-3 text-xs font-medium text-muted-foreground transition hover:bg-muted/50 hover:text-foreground',
                    )}
                >
                    从上游添加
                </button>
            </div>
            <div className="min-h-0 flex-1">
                <VirtualizedGrid
                    items={visibleChannels}
                    layout="grid"
                    columns={{ default: 1, md: 2, lg: 3, xl: 3, '2xl': 3 }}
                    estimateItemHeight={channelDensity === 'compact' ? 214 : 248}
                    gap={channelDensity === 'compact' ? 10 : 12}
                    getItemKey={(item) => `channel-${item.raw.id}`}
                    renderItem={(item) => <Card channel={item.raw} stats={item.formatted} density={channelDensity} />}
                />
            </div>
        </div>
    );
}
