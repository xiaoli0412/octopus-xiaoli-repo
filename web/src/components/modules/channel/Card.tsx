import {
    MorphingDialog,
    MorphingDialogTrigger,
    MorphingDialogContainer,
    MorphingDialogContent,
} from '@/components/ui/morphing-dialog';
import { Badge } from '@/components/ui/badge';
import { DollarSign, MessageSquare } from 'lucide-react';
import { cn } from '@/lib/utils';
import { type StatsMetricsFormatted } from '@/api/endpoints/stats';
import { normalizeKeyManagementMode, normalizeKeyRoutingPolicy, type Channel, useEnableChannel } from '@/api/endpoints/channel';
import { CardContent } from './CardContent';
import { useTranslations } from 'next-intl';
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/animate-ui/components/animate/tooltip';
import { Switch } from '@/components/ui/switch';
import { toast } from '@/components/common/Toast';
import { ChannelType } from '@/api/endpoints/channel';

function getChannelTypeLabel(type: ChannelType, t: ReturnType<typeof useTranslations>) {
    switch (type) {
        case ChannelType.OpenAIChat:
            return t('typeNames.openaiChat');
        case ChannelType.OpenAIResponse:
            return t('typeNames.openaiResponse');
        case ChannelType.Anthropic:
            return t('typeNames.anthropic');
        case ChannelType.Gemini:
            return t('typeNames.gemini');
        case ChannelType.Volcengine:
            return t('typeNames.volcengine');
        case ChannelType.OpenAIEmbedding:
            return t('typeNames.openaiEmbedding');
        case ChannelType.GithubCopilot:
            return t('typeNames.githubCopilot');
        case ChannelType.Antigravity:
            return t('typeNames.antigravity');
        case ChannelType.Zen:
            return t('typeNames.zen');
        default:
            return t('typeNames.unknown');
    }
}

export function Card({ channel, stats }: { channel: Channel; stats: StatsMetricsFormatted; layout?: 'grid' | 'list' }) {
    const t = useTranslations('channel.card');
    const enableChannel = useEnableChannel();
    const keyCount = channel.keys?.length ?? 0;
    const channelTypeLabel = getChannelTypeLabel(channel.type, t);
    const keyManagementMode = normalizeKeyManagementMode(channel.key_management_mode);
    const keyRoutingPolicy = normalizeKeyRoutingPolicy(channel.key_routing_policy);
    const modeBadgeClass = keyManagementMode === 'classified'
        ? 'bg-sky-500/10 text-sky-700 dark:text-sky-300 border-sky-500/30'
        : 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border-emerald-500/30';
    const policyBadgeClass = keyRoutingPolicy === 'fill_priority'
        ? 'bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/30'
        : keyRoutingPolicy === 'priority_order'
            ? 'bg-violet-500/10 text-violet-700 dark:text-violet-300 border-violet-500/30'
            : 'bg-muted text-muted-foreground border-muted-foreground/20';

    const handleEnableChange = (checked: boolean) => {
        enableChannel.mutate(
            { id: channel.id, enabled: checked },
            {
                onSuccess: () => {
                    toast.success(checked ? t('toast.enabled') : t('toast.disabled'));
                },
                onError: (error) => {
                    toast.error(error.message);
                },
            }
        );
    };

    return (
        <MorphingDialog>
            <MorphingDialogTrigger data-testid={`channel-card-trigger-${channel.id}`} className="w-full">
                <article data-testid={`channel-card-${channel.id}`} data-channel-name={channel.name} className="relative flex min-h-[15rem] flex-col gap-3 overflow-hidden rounded-3xl border border-border bg-card p-4 text-card-foreground transition-all duration-300">
                    <header className="relative flex items-start justify-between gap-3">
                        <div className="min-w-0 space-y-1">
                            <Tooltip side="top" sideOffset={10} align="center">
                                <TooltipTrigger asChild>
                                    <h3 className="truncate text-base font-bold min-w-0 leading-tight">{channel.name}</h3>
                                </TooltipTrigger>
                                <TooltipContent key={channel.name}>{channel.name}</TooltipContent>
                            </Tooltip>
                            <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground">
                                <span>{channelTypeLabel}</span>
                                <span className="text-border">•</span>
                                {t('keyCount', { count: keyCount })}
                            </div>
                        </div>
                        <Switch
                            checked={channel.enabled}
                            onCheckedChange={handleEnableChange}
                            disabled={enableChannel.isPending}
                            onClick={(e) => e.stopPropagation()}
                        />
                    </header>

                    <div data-testid={`channel-card-badges-${channel.id}`} className="flex flex-wrap items-center gap-2">
                        <Badge variant="outline" className={cn('h-7 rounded-full px-3 text-[11px] font-medium', modeBadgeClass)}>
                            {t(`mode.${keyManagementMode}`)}
                        </Badge>
                        <Badge variant="outline" className={cn('h-7 rounded-full px-3 text-[11px] font-medium', policyBadgeClass)}>
                            {t(`policy.${keyRoutingPolicy}`)}
                        </Badge>
                        <Badge variant="outline" className="h-7 rounded-full px-3 text-[11px] font-medium">
                            {t('keyCountBadge', { count: keyCount })}
                        </Badge>
                    </div>

                    <dl data-testid={`channel-card-metrics-${channel.id}`} className="relative mt-auto grid grid-cols-1 gap-2">
                        <div className="grid min-h-[4.25rem] grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-3 rounded-2xl border border-border/70 bg-background/80 px-3 py-2.5">
                            <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
                                    <MessageSquare className="h-5 w-5" />
                            </span>
                            <dt className="min-w-0 text-sm leading-tight text-muted-foreground">{t('requestCount')}</dt>
                            <dd className="shrink-0 text-sm font-semibold tabular-nums">
                                {stats.request_count.formatted.value}
                                <span className="ml-1 text-xs text-muted-foreground">{stats.request_count.formatted.unit}</span>
                            </dd>
                        </div>

                        <div className="grid min-h-[4.25rem] grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-3 rounded-2xl border border-border/70 bg-background/80 px-3 py-2.5">
                            <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
                                    <DollarSign className="h-5 w-5" />
                            </span>
                            <dt className="min-w-0 text-sm leading-tight text-muted-foreground">{t('totalCost')}</dt>
                            <dd className="shrink-0 text-sm font-semibold tabular-nums">
                                {stats.total_cost.formatted.value}
                                <span className="ml-1 text-xs text-muted-foreground">{stats.total_cost.formatted.unit}</span>
                            </dd>
                        </div>
                    </dl>
                </article>
            </MorphingDialogTrigger>

            <MorphingDialogContainer>
                <MorphingDialogContent data-testid={`channel-detail-dialog-${channel.id}`} className="w-full md:max-w-xl bg-card text-card-foreground px-4 py-2 rounded-3xl max-h-[90vh] overflow-y-auto">
                    <CardContent channel={channel} stats={stats} />
                </MorphingDialogContent>
            </MorphingDialogContainer>
        </MorphingDialog>
    );
}
