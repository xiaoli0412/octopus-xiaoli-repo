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
import type { ChannelCardDensity } from '@/components/modules/toolbar/view-options-store';

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

export function Card({ channel, stats, density = 'normal' }: { channel: Channel; stats: StatsMetricsFormatted; density?: ChannelCardDensity }) {
    const t = useTranslations('channel.card');
    const enableChannel = useEnableChannel();
    const keyCount = channel.keys?.length ?? 0;
    const channelTypeLabel = getChannelTypeLabel(channel.type, t);
    const isCompact = density === 'compact';
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
                <article
                    data-testid={`channel-card-${channel.id}`}
                    data-channel-name={channel.name}
                    data-density={density}
                    className={cn(
                        'relative flex flex-col overflow-hidden border border-border bg-card text-card-foreground transition-all duration-300',
                        isCompact ? 'min-h-[13rem] gap-2.5 rounded-[1.6rem] p-3' : 'min-h-[15rem] gap-3 rounded-3xl p-4'
                    )}
                >
                    <header className={cn('relative flex items-start justify-between', isCompact ? 'gap-2.5' : 'gap-3')}>
                        <div className="min-w-0 space-y-1">
                            <Tooltip side="top" sideOffset={10} align="center">
                                <TooltipTrigger asChild>
                                    <h3 className={cn('min-w-0 truncate font-bold leading-tight', isCompact ? 'text-[15px]' : 'text-base')}>{channel.name}</h3>
                                </TooltipTrigger>
                                <TooltipContent key={channel.name}>{channel.name}</TooltipContent>
                            </Tooltip>
                            <div className={cn('flex flex-wrap items-center gap-y-1 text-muted-foreground', isCompact ? 'gap-x-1.5 text-[11px]' : 'gap-x-2 text-xs')}>
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

                    <div data-testid={`channel-card-badges-${channel.id}`} className={cn('flex flex-wrap items-center', isCompact ? 'gap-1.5' : 'gap-2')}>
                        <Badge variant="outline" className={cn('rounded-full font-medium', isCompact ? 'h-6 px-2.5 text-[10px]' : 'h-7 px-3 text-[11px]', modeBadgeClass)}>
                            {t(`mode.${keyManagementMode}`)}
                        </Badge>
                        <Badge variant="outline" className={cn('rounded-full font-medium', isCompact ? 'h-6 px-2.5 text-[10px]' : 'h-7 px-3 text-[11px]', policyBadgeClass)}>
                            {t(`policy.${keyRoutingPolicy}`)}
                        </Badge>
                        <Badge variant="outline" className={cn('rounded-full font-medium', isCompact ? 'h-6 px-2.5 text-[10px]' : 'h-7 px-3 text-[11px]')}>
                            {t('keyCountBadge', { count: keyCount })}
                        </Badge>
                        {channel.upstream_site_id ? (
                            <Badge variant="outline" className={cn('rounded-full border-cyan-500/25 bg-cyan-500/10 font-medium text-cyan-700 dark:text-cyan-300', isCompact ? 'h-6 px-2.5 text-[10px]' : 'h-7 px-3 text-[11px]')}>
                                上游：{channel.upstream_source || `#${channel.upstream_site_id}`}
                            </Badge>
                        ) : null}
                    </div>

                    <dl data-testid={`channel-card-metrics-${channel.id}`} className={cn('relative mt-auto grid grid-cols-1', isCompact ? 'gap-1.5' : 'gap-2')}>
                        <div className={cn('grid grid-cols-[auto_minmax(0,1fr)_auto] items-center border border-border/70 bg-background/80', isCompact ? 'min-h-[3.5rem] gap-2.5 rounded-xl px-2.5 py-2' : 'min-h-[4.25rem] gap-3 rounded-2xl px-3 py-2.5')}>
                            <span className={cn('flex items-center justify-center bg-primary/10 text-primary', isCompact ? 'h-9 w-9 rounded-lg' : 'h-10 w-10 rounded-xl')}>
                                    <MessageSquare className={cn(isCompact ? 'h-4 w-4' : 'h-5 w-5')} />
                            </span>
                            <dt className={cn('min-w-0 leading-tight text-muted-foreground', isCompact ? 'text-xs' : 'text-sm')}>{t('requestCount')}</dt>
                            <dd className={cn('shrink-0 font-semibold tabular-nums', isCompact ? 'text-[13px]' : 'text-sm')}>
                                {stats.request_count.formatted.value}
                                <span className={cn('ml-1 text-muted-foreground', isCompact ? 'text-[10px]' : 'text-xs')}>{stats.request_count.formatted.unit}</span>
                            </dd>
                        </div>

                        <div className={cn('grid grid-cols-[auto_minmax(0,1fr)_auto] items-center border border-border/70 bg-background/80', isCompact ? 'min-h-[3.5rem] gap-2.5 rounded-xl px-2.5 py-2' : 'min-h-[4.25rem] gap-3 rounded-2xl px-3 py-2.5')}>
                            <span className={cn('flex items-center justify-center bg-primary/10 text-primary', isCompact ? 'h-9 w-9 rounded-lg' : 'h-10 w-10 rounded-xl')}>
                                    <DollarSign className={cn(isCompact ? 'h-4 w-4' : 'h-5 w-5')} />
                            </span>
                            <dt className={cn('min-w-0 leading-tight text-muted-foreground', isCompact ? 'text-xs' : 'text-sm')}>{t('totalCost')}</dt>
                            <dd className={cn('shrink-0 font-semibold tabular-nums', isCompact ? 'text-[13px]' : 'text-sm')}>
                                {stats.total_cost.formatted.value}
                                <span className={cn('ml-1 text-muted-foreground', isCompact ? 'text-[10px]' : 'text-xs')}>{stats.total_cost.formatted.unit}</span>
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
