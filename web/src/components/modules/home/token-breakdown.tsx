'use client';

import { useMemo, useState } from 'react';
import { useLocale, useTranslations } from 'next-intl';
import { ArrowDownToLine, ArrowUpFromLine, ChevronDown, ChevronUp, Layers3, Network, ReceiptText, Server, Sparkles, type LucideIcon } from 'lucide-react';
import { AnimatePresence, motion } from 'motion/react';

import { ChannelType, useChannelList } from '@/api/endpoints/channel';
import { useStatsTokenBreakdown, type StatsTokenWindow } from '@/api/endpoints/stats';
import { AnimatedNumber } from '@/components/common/AnimatedNumber';
import { HelpHint } from '@/components/common/HelpHint';
import { cn } from '@/lib/utils';
import { formatCount } from '@/lib/utils';

type BreakdownItem = {
	key: string;
	label: string;
	total_token: { raw: number; formatted: { value: string; unit: string } };
};

type BreakdownType = 'channel' | 'provider' | 'model' | 'apikey';

function windowLabel(window: StatsTokenWindow, t: ReturnType<typeof useTranslations<'home.tokenBreakdown'>>) {
	return t(`window.${window}` as never);
}

function formatProbeLastAt(timestamp: number, locale: string) {
	if (!timestamp) return '--';
	return new Intl.DateTimeFormat(locale, {
		month: '2-digit',
		day: '2-digit',
		hour: '2-digit',
		minute: '2-digit',
	}).format(new Date(timestamp * 1000));
}

function parseChannelId(key: string) {
	const matched = key.match(/^channel:(\d+)$/);
	if (!matched) return null;
	const id = Number.parseInt(matched[1], 10);
	return Number.isFinite(id) ? id : null;
}

function SummaryMetric({ title, value, unit, Icon }: { title: string; value: string | number | undefined; unit?: string; Icon: LucideIcon }) {
	return (
		<div className="rounded-2xl border border-border/60 bg-background/55 px-3.5 py-3">
			<div className="mb-1.5 flex items-center gap-2 text-xs font-medium text-muted-foreground">
				<span className="flex h-7 w-7 items-center justify-center rounded-lg bg-primary/10 text-primary">
					<Icon className="h-3.5 w-3.5" />
				</span>
				<span>{title}</span>
			</div>
			<div className="flex min-w-0 items-baseline gap-1">
				<span className="truncate text-[1.05rem] font-semibold text-foreground">
					<AnimatedNumber value={value ?? 0} />
				</span>
				{unit ? <span className="shrink-0 text-xs text-muted-foreground">{unit}</span> : null}
			</div>
		</div>
	);
}

function RuntimeMetric({ label, value, unit }: { label: string; value: string | number | undefined; unit?: string }) {
	return (
		<div className="rounded-xl bg-background/70 px-3 py-2.5">
			<div className="text-[11px] text-muted-foreground">{label}</div>
			<div className="mt-1 flex min-w-0 items-baseline gap-1 text-sm font-semibold text-foreground">
				<span className="truncate">
					<AnimatedNumber value={value ?? 0} />
				</span>
				{unit ? <span className="shrink-0 text-xs text-muted-foreground">{unit}</span> : null}
			</div>
		</div>
	);
}

function BreakdownBoard({
	title,
	items,
	type,
	emptyText,
	badge,
}: {
	title: string;
	items: BreakdownItem[];
	type: BreakdownType;
	emptyText: string;
	badge: string;
}) {
	const Icon = type === 'channel' ? Network : type === 'provider' ? Server : type === 'apikey' ? ReceiptText : Layers3;

	return (
		<div className="rounded-3xl border border-border/60 bg-background/45 p-3.5">
			<div className="mb-3 flex items-center justify-between gap-3">
				<div className="flex min-w-0 items-center gap-2">
					<Icon className="size-4 shrink-0 text-primary" />
					<div className="truncate text-sm font-semibold text-foreground">{title}</div>
				</div>
				<div className="shrink-0 rounded-full border border-border/50 bg-card/70 px-2.5 py-0.5 text-[11px] text-muted-foreground">{badge}</div>
			</div>

			<div className="max-h-[18rem] space-y-2 overflow-y-auto pr-1">
				{items.length === 0 ? (
					<div className="rounded-2xl border border-dashed border-border/60 bg-background/50 px-3 py-4 text-sm text-muted-foreground">{emptyText}</div>
				) : (
					items.map((item) => (
						<div key={item.key} className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3 rounded-2xl border border-border/50 bg-card/60 px-3 py-2.5">
							<div className="min-w-0 break-all text-sm text-foreground">{item.label}</div>
							<div className="shrink-0 text-right text-sm font-semibold tabular-nums text-foreground">
								<AnimatedNumber value={item.total_token.formatted.value} />
								{item.total_token.formatted.unit ? <span className="ml-1 text-xs text-muted-foreground">{item.total_token.formatted.unit}</span> : null}
							</div>
						</div>
					))
				)}
			</div>
		</div>
	);
}

export function TokenBreakdown() {
	const t = useTranslations('home.tokenBreakdown');
	const tChannel = useTranslations('channel.form');
	const locale = useLocale();
	const [window, setWindow] = useState<StatsTokenWindow>('1d');
	const [dimension, setDimension] = useState<'channel' | 'model' | 'api_key' | 'channel_key'>('channel');
	const [showRuntimeDetails, setShowRuntimeDetails] = useState(false);
	const [showMoreLists, setShowMoreLists] = useState(false);
	const { data } = useStatsTokenBreakdown(window);
	const { data: channels = [] } = useChannelList();

	const channelTypeMap = useMemo(() => new Map(channels.map((item) => [item.raw.id, item.raw.type])), [channels]);

	const providerItems = useMemo(() => {
		const providerMap = new Map<string, { key: string; label: string; input_token: number; output_token: number; total_token: number }>();

		const getProviderMeta = (channelType?: ChannelType) => {
			switch (channelType) {
				case ChannelType.OpenAIChat:
				case ChannelType.OpenAIResponse:
				case ChannelType.OpenAIEmbedding:
					return { key: 'openai-compatible', label: t('providerOpenAICompatible') };
				case ChannelType.Anthropic:
					return { key: 'anthropic', label: tChannel('typeAnthropic') };
				case ChannelType.Gemini:
					return { key: 'gemini', label: tChannel('typeGemini') };
				case ChannelType.Volcengine:
					return { key: 'volcengine', label: tChannel('typeVolcengine') };
				case ChannelType.GithubCopilot:
					return { key: 'github-copilot', label: tChannel('typeGithubCopilot') };
				case ChannelType.Antigravity:
					return { key: 'antigravity', label: tChannel('typeAntigravity') };
				case ChannelType.Zen:
					return { key: 'zen', label: tChannel('typeZen') };
				default:
					return { key: 'unknown', label: t('providerUnknown') };
			}
		};

		(data?.by_channel ?? []).forEach((item) => {
			const channelId = parseChannelId(item.key);
			const { key, label } = getProviderMeta(channelId === null ? undefined : channelTypeMap.get(channelId));
			const existing = providerMap.get(key) ?? { key, label, input_token: 0, output_token: 0, total_token: 0 };
			existing.input_token += item.input_token.raw;
			existing.output_token += item.output_token.raw;
			existing.total_token += item.total_token.raw;
			providerMap.set(key, existing);
		});

		return Array.from(providerMap.values())
			.map((item) => ({
				key: item.key,
				label: item.label,
				input_token: formatCount(item.input_token),
				output_token: formatCount(item.output_token),
				total_token: formatCount(item.total_token),
			}))
			.sort((a, b) => b.total_token.raw - a.total_token.raw);
	}, [channelTypeMap, data?.by_channel, t, tChannel]);

	const listLimit = showMoreLists ? 8 : 5;
	const topProviders = useMemo(() => providerItems.slice(0, listLimit), [providerItems, listLimit]);
	const topChannels = useMemo(() => (data?.by_channel ?? []).slice(0, listLimit), [data?.by_channel, listLimit]);
	const topModels = useMemo(() => (data?.by_model ?? []).slice(0, listLimit), [data?.by_model, listLimit]);
	const topAPIKeys = useMemo(() => (data?.by_api_key ?? []).slice(0, listLimit), [data?.by_api_key, listLimit]);
	const topChannelKeys = useMemo(() => (data?.by_channel_key ?? []).slice(0, listLimit), [data?.by_channel_key, listLimit]);

	const canExpandLists = providerItems.length > 5 || (data?.by_channel.length ?? 0) > 5 || (data?.by_model.length ?? 0) > 5 || (data?.by_api_key?.length ?? 0) > 5 || (data?.by_channel_key?.length ?? 0) > 5;
	const estimatedDelta = useMemo(() => (data?.estimated_official_total_cost.raw ?? 0) - (data?.estimated_gateway_total_cost.raw ?? 0), [data?.estimated_gateway_total_cost.raw, data?.estimated_official_total_cost.raw]);
	const recentProbeAtLabel = useMemo(() => formatProbeLastAt(data?.recent_probe_last_at ?? 0, locale), [data?.recent_probe_last_at, locale]);
	const recentProbeStatusLabel = useMemo(() => {
		switch (data?.recent_probe_last_status) {
			case 'success':
				return t('probeStatus.success');
			case 'selected':
				return t('probeStatus.selected');
			case 'failed':
				return t('probeStatus.failed');
			default:
				return data?.recent_probe_last_status || '--';
		}
	}, [data?.recent_probe_last_status, t]);

	return (
		<section data-testid="home-breakdown-section" className="rounded-3xl border border-card-border bg-card p-4 text-card-foreground sm:p-5">
			<div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
				<div>
					<div className="flex items-center gap-2">
						<div className="text-lg font-semibold text-foreground">{t('title')}</div>
						<HelpHint ariaLabel={t('title')}>{t('summaryHint')}</HelpHint>
					</div>
					<div className="mt-1 text-xs leading-5 text-muted-foreground">{t('subtitle')}</div>
				</div>

				<div className="flex flex-wrap gap-2 xl:justify-end">
					<button
						type="button"
						data-testid="home-runtime-toggle"
						onClick={() => setShowRuntimeDetails((value) => !value)}
						className="inline-flex items-center gap-1.5 rounded-xl border border-border/60 bg-background/70 px-3 py-2 text-sm text-muted-foreground transition hover:border-primary/30 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:ring-offset-2 focus-visible:ring-offset-background"
					>
						<Sparkles className="size-4" />
						<span>{showRuntimeDetails ? t('hideRuntime') : t('showRuntime')}</span>
					</button>

					<div className="inline-flex rounded-2xl border border-border/60 bg-background/70 p-1 text-xs">
						{(['12h', '1d', '3d', '7d', '30d'] as StatsTokenWindow[]).map((item) => (
							<button
								key={item}
								type="button"
								className={cn('rounded-xl px-2.5 py-1.5 transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:ring-offset-2 focus-visible:ring-offset-background', window === item ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:text-foreground')}
								onClick={() => setWindow(item)}
							>
								{windowLabel(item, t)}
							</button>
						))}
					</div>

					{canExpandLists ? (
						<button
							type="button"
							data-testid="home-breakdown-list-toggle"
							onClick={() => setShowMoreLists((value) => !value)}
							className="inline-flex items-center gap-1.5 rounded-xl border border-border/60 bg-background/70 px-3 py-2 text-sm text-muted-foreground transition hover:border-primary/30 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:ring-offset-2 focus-visible:ring-offset-background"
						>
							{showMoreLists ? <ChevronUp className="size-4" /> : <ChevronDown className="size-4" />}
							<span>{showMoreLists ? t('hideMoreLists') : t('showMoreLists')}</span>
						</button>
					) : null}
				</div>
			</div>

			<div className="mt-4 grid gap-3 min-[375px]:grid-cols-2 xl:grid-cols-4">
				<SummaryMetric title={t('total')} value={data?.total_token.formatted.value ?? 0} unit={data?.total_token.formatted.unit} Icon={Layers3} />
				<SummaryMetric title={t('input')} value={data?.total_input_token.formatted.value ?? 0} unit={data?.total_input_token.formatted.unit} Icon={ArrowDownToLine} />
				<SummaryMetric title={t('output')} value={data?.total_output_token.formatted.value ?? 0} unit={data?.total_output_token.formatted.unit} Icon={ArrowUpFromLine} />
				<SummaryMetric title={t('estimatedGateway')} value={data?.estimated_gateway_total_cost.formatted.value ?? 0} unit={data?.estimated_gateway_total_cost.formatted.unit} Icon={ReceiptText} />
			</div>

			<div className="mt-4 flex flex-wrap gap-2">
				{[
					{ key: 'channel', label: t('byChannel') },
					{ key: 'model', label: t('byModel') },
					{ key: 'api_key', label: t('byAPIKey') },
					{ key: 'channel_key', label: t('byChannelKey') },
				].map((item) => (
					<button
						key={item.key}
						type="button"
						className={cn('rounded-full border px-3 py-1.5 text-xs transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:ring-offset-2 focus-visible:ring-offset-background', dimension === item.key ? 'border-primary/30 bg-primary/10 text-primary' : 'border-border bg-background text-muted-foreground hover:border-primary/20 hover:text-foreground')}
						onClick={() => setDimension(item.key as typeof dimension)}
					>
						{item.label}
					</button>
				))}
			</div>

			<div className="mt-4 grid gap-3 xl:grid-cols-2">
				<BreakdownBoard title={t('byProvider')} items={topProviders} emptyText={t('noData')} type="provider" badge={t('topItems', { count: topProviders.length })} />
				{dimension === 'channel' ? <BreakdownBoard title={t('byChannel')} items={topChannels} emptyText={t('noData')} type="channel" badge={t('topItems', { count: topChannels.length })} /> : null}
				{dimension === 'model' ? <BreakdownBoard title={t('byModel')} items={topModels} emptyText={t('noData')} type="model" badge={t('topItems', { count: topModels.length })} /> : null}
				{dimension === 'api_key' ? <BreakdownBoard title={t('byAPIKey')} items={topAPIKeys} emptyText={t('noData')} type="apikey" badge={t('topItems', { count: topAPIKeys.length })} /> : null}
				{dimension === 'channel_key' ? <BreakdownBoard title={t('byChannelKey')} items={topChannelKeys} emptyText={t('noData')} type="apikey" badge={t('topItems', { count: topChannelKeys.length })} /> : null}
			</div>

			<AnimatePresence initial={false}>
				{showRuntimeDetails ? (
					<motion.div
						data-testid="home-runtime-panel"
						initial={{ opacity: 0, y: 12, height: 0 }}
						animate={{ opacity: 1, y: 0, height: 'auto' }}
						exit={{ opacity: 0, y: -8, height: 0 }}
						transition={{ duration: 0.24, ease: 'easeOut' }}
						className="overflow-hidden"
					>
						<div className="mt-4 rounded-3xl border border-border/60 bg-background/45 p-4">
							<div className="mb-3 flex items-center gap-2">
								<div className="text-sm font-medium text-foreground">{t('runtimePanelTitle')}</div>
								<HelpHint ariaLabel={t('runtimePanelTitle')}>{t('runtimePanelHint')}</HelpHint>
							</div>
							<div className="grid gap-3 xl:grid-cols-3">
								<div className="rounded-2xl border border-border/60 bg-card/50 p-3.5">
									<div className="mb-2 flex items-center gap-2 text-sm font-medium text-foreground">
										<ReceiptText className="size-4" />
										<span>{t('estimatedPriceTitle')}</span>
										<HelpHint ariaLabel={t('estimatedPriceTitle')}>{t('estimatedPriceHint')}</HelpHint>
									</div>
									<div className="grid gap-2 sm:grid-cols-2">
										<RuntimeMetric label={t('estimatedGateway')} value={data?.estimated_gateway_total_cost.formatted.value ?? 0} unit={data?.estimated_gateway_total_cost.formatted.unit} />
										<RuntimeMetric label={t('estimatedOfficial')} value={data?.estimated_official_total_cost.formatted.value ?? 0} unit={data?.estimated_official_total_cost.formatted.unit} />
										<RuntimeMetric label={t('estimatedDelta')} value={estimatedDelta.toFixed(2)} unit="$" />
										<RuntimeMetric label={t('probeCost')} value={data?.estimated_probe_total_cost.formatted.value ?? 0} unit={data?.estimated_probe_total_cost.formatted.unit} />
									</div>
								</div>

								<div className="rounded-2xl border border-border/60 bg-card/50 p-3.5">
									<div className="mb-2 flex items-center gap-2 text-sm font-medium text-foreground">
										<Server className="size-4" />
										<span>{t('circuitSummaryTitle')}</span>
									</div>
									<div className="grid gap-2 sm:grid-cols-2">
										<RuntimeMetric label={t('circuitTracked')} value={data?.circuit_tracked_count ?? 0} />
										<RuntimeMetric label={t('circuitOpen')} value={data?.circuit_open_count ?? 0} />
										<RuntimeMetric label={t('circuitHalfOpen')} value={data?.circuit_half_open_count ?? 0} />
										<RuntimeMetric label={t('circuitCooldown')} value={data?.circuit_max_remaining_cooldown_sec ?? 0} unit="秒" />
									</div>
								</div>

								<div className="rounded-2xl border border-border/60 bg-card/50 p-3.5">
									<div className="mb-2 flex items-center gap-2 text-sm font-medium text-foreground">
										<Network className="size-4" />
										<span>{t('probeSummaryTitle')}</span>
									</div>
									<div className="grid gap-2 sm:grid-cols-2">
										<RuntimeMetric label={t('probeCount')} value={data?.recent_probe_count.formatted.value ?? 0} unit={data?.recent_probe_count.formatted.unit} />
										<RuntimeMetric label={t('probeSuccess')} value={data?.recent_probe_success_count.formatted.value ?? 0} unit={data?.recent_probe_success_count.formatted.unit} />
										<RuntimeMetric label={t('probeFailed')} value={data?.recent_probe_failed_count.formatted.value ?? 0} unit={data?.recent_probe_failed_count.formatted.unit} />
										<div className="rounded-xl bg-background/70 px-3 py-2.5">
											<div className="text-[11px] text-muted-foreground">{t('probeLastAt')}</div>
											<div className="mt-1 text-sm font-semibold text-foreground">{recentProbeAtLabel}</div>
										</div>
									</div>
									<div className="mt-2 rounded-xl bg-background/70 px-3 py-2 text-xs leading-5 text-muted-foreground break-all">
										{t('probeLastDetail', {
											status: recentProbeStatusLabel,
											channel: data?.recent_probe_last_channel || '--',
											model: data?.recent_probe_last_model || '--',
										})}
										{data?.recent_probe_last_message ? <div className="mt-1 break-all">{data.recent_probe_last_message}</div> : null}
									</div>
								</div>
							</div>
						</div>
					</motion.div>
				) : null}
			</AnimatePresence>
		</section>
	);
}
