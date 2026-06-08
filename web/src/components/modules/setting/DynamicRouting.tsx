'use client';

import { useState, type ComponentType } from 'react';
import { useTranslations } from 'next-intl';
import { Activity, Gauge, Network, Radio, ShieldCheck, Waves } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { HelpHint } from '@/components/common/HelpHint';
import { toast } from '@/components/common/Toast';
import { useDynamicRouteLearning, useResetDynamicRouteLearning } from '@/api/endpoints/ai-automation';
import { useSettingList, useSetSetting, SettingKey } from '@/api/endpoints/setting';
import { useStatsDynamicRoutingSummary } from '@/api/endpoints/stats';
import { useNavStore } from '@/components/modules/navbar';
import { formatDateTimeByLocale } from '@/lib/locale';
import { useSettingStore } from '@/stores/setting';
import { queueAIAutomationFocusTarget } from '@/components/modules/ai-automation/focus-target';
import { LearningSummaryActionBar, LearningSummaryPanel } from '@/components/modules/ai-automation/LearningSummaryPanel';
import {
	DEFAULT_LEARNING_SUMMARY_SECTION_ENTRIES,
	buildLearningSummarySections,
	buildLearningSummaryViewModel,
} from '@/components/modules/ai-automation/learning-summary';

type BudgetField = {
	key: string;
	icon: ComponentType<{ className?: string }>;
	label: string;
	hint: string;
	value: string;
	initialValue: string;
	onChange: (value: string) => void;
};

const DYNAMIC_ROUTING_SUMMARY_MESSAGE_HEALTH_DISABLED_SCAN_SKIPPED = 'health_disabled_scan_skipped';
const DYNAMIC_ROUTING_SUMMARY_MESSAGE_HEALTH_DISABLED_SCAN_SKIPPED_LEGACY = 'dynamic routing health disabled; summary scan skipped';
const DYNAMIC_ROUTING_MODE_VALUES = ['shadow-ai', 'hybrid', 'metrics-only', 'strict-mechanism', 'incident-safe'] as const;

type DynamicRoutingMode = (typeof DYNAMIC_ROUTING_MODE_VALUES)[number];

function getDisplayValue(value: string, fallback: string) {
	if (value !== '') return value;
	if (fallback !== '') return fallback;
	return '-';
}

function normalizeSummaryStatus(status?: string) {
	if (!status) return 'unknown';
	switch (status) {
		case 'ok':
		case 'skipped':
		case 'error':
			return status;
		default:
			return 'unknown';
	}
}

function normalizeSummaryBasis(basis?: string) {
	if (!basis) return 'unknown';
	switch (basis) {
		case 'daily_summary_scan_no_runtime_mutation':
			return basis;
		default:
			return 'unknown';
	}
}

function normalizeSummaryMessage(message?: string) {
	if (!message) return undefined;
	switch (message) {
		case DYNAMIC_ROUTING_SUMMARY_MESSAGE_HEALTH_DISABLED_SCAN_SKIPPED:
		case DYNAMIC_ROUTING_SUMMARY_MESSAGE_HEALTH_DISABLED_SCAN_SKIPPED_LEGACY:
			return DYNAMIC_ROUTING_SUMMARY_MESSAGE_HEALTH_DISABLED_SCAN_SKIPPED;
		default:
			return undefined;
	}
}

function normalizeDynamicRoutingMode(value?: string): DynamicRoutingMode {
	switch (value) {
		case 'shadow-ai':
		case 'hybrid':
		case 'metrics-only':
		case 'strict-mechanism':
		case 'incident-safe':
			return value;
		default:
			return 'hybrid';
	}
}

function normalizeSummaryDecision(value?: string) {
	if (!value) return 'unknown';
	switch (value) {
		case 'recommended':
		case 'shadow':
		case 'metrics':
		case 'deterministic':
			return value;
	}
	return 'unknown';
}

function normalizeSummaryDecisionReason(value?: string) {
	if (!value) return undefined;
	switch (value) {
		case 'summary_snapshot_runtime_modes':
		case 'summary_snapshot_health_disabled':
			return value;
	}
	return 'unknown';
}

export function SettingDynamicRouting() {
	const t = useTranslations('setting');
	const locale = useSettingStore((state) => state.locale);
	const { data: settings } = useSettingList();
	const dynamicRoutingSummary = useStatsDynamicRoutingSummary();
	const learning = useDynamicRouteLearning();
	const resetLearning = useResetDynamicRouteLearning();
	const setSetting = useSetSetting();
	const setActiveItem = useNavStore((state) => state.setActiveItem);

	const openAIAutomationLearning = () => {
		queueAIAutomationFocusTarget('learning');
		setActiveItem('ai');
	};

	const healthEnabledSetting = settings?.find((setting) => setting.key === SettingKey.DynamicRoutingHealthEnabled)?.value === 'true';
	const learningEnabledSetting = settings?.find((setting) => setting.key === SettingKey.DynamicRoutingLearningEnabled)?.value === 'true';
	const routingModeSetting = normalizeDynamicRoutingMode(settings?.find((setting) => setting.key === SettingKey.DynamicRoutingMode)?.value);
	const globalBudgetSetting = settings?.find((setting) => setting.key === SettingKey.RaceGlobalBudget)?.value ?? '';
	const groupBudgetSetting = settings?.find((setting) => setting.key === SettingKey.RaceGroupBudget)?.value ?? '';
	const channelBudgetSetting = settings?.find((setting) => setting.key === SettingKey.RaceChannelBudget)?.value ?? '';
	const keyBudgetSetting = settings?.find((setting) => setting.key === SettingKey.RaceKeyBudget)?.value ?? '';
	const probeBudgetSetting = settings?.find((setting) => setting.key === SettingKey.RaceProbeBudget)?.value ?? '';

	const [healthEnabledDraft, setHealthEnabledDraft] = useState<boolean | null>(null);
	const [learningEnabledDraft, setLearningEnabledDraft] = useState<boolean | null>(null);
	const [routingModeDraft, setRoutingModeDraft] = useState<DynamicRoutingMode | null>(null);
	const [globalBudgetDraft, setGlobalBudgetDraft] = useState<string | null>(null);
	const [groupBudgetDraft, setGroupBudgetDraft] = useState<string | null>(null);
	const [channelBudgetDraft, setChannelBudgetDraft] = useState<string | null>(null);
	const [keyBudgetDraft, setKeyBudgetDraft] = useState<string | null>(null);
	const [probeBudgetDraft, setProbeBudgetDraft] = useState<string | null>(null);

	const healthEnabled = healthEnabledDraft ?? healthEnabledSetting;
	const learningEnabledRuntime = learning.data?.enabled;
	const learningEnabled = learningEnabledDraft ?? learningEnabledRuntime ?? learningEnabledSetting;
	const learningStates = learning.data?.states ?? [];
	const learningSummaryView = buildLearningSummaryViewModel({
		states: learningStates,
		enabled: !!learningEnabled,
		locale,
		emptyLabel: t('dynamicRouting.learning.notAvailable'),
		topTargetFormatter: (state) => t('dynamicRouting.learning.topTargetValue', {
			model: state.model_name,
			channel: state.channel_id,
			key: state.channel_key_id,
		}),
		itemLabels: {
			statusLabel: t('dynamicRouting.learning.statusLabel'),
			enabledLabel: t('dynamicRouting.learning.enabledBadge'),
			disabledLabel: t('dynamicRouting.learning.disabledBadge'),
			samplesLabel: t('dynamicRouting.learning.samplesLabel'),
			runtimeLabel: t('dynamicRouting.learning.runtimeLabel'),
			runtimeEnabledLabel: t('dynamicRouting.learning.runtimeEnabled'),
			runtimeDisabledLabel: t('dynamicRouting.learning.runtimeDisabled'),
			latestSampleLabel: t('dynamicRouting.learning.latestSampleLabel'),
			topTargetLabel: t('dynamicRouting.learning.topTargetLabel'),
		},
		noticeValues: {
			enabledWithSamples: t('dynamicRouting.learning.summaryEnabledWithSamples'),
			enabledWithoutSamples: t('dynamicRouting.learning.summaryEnabledNoSamples'),
			disabledWithSamples: t('dynamicRouting.learning.summaryDisabledWithSamples'),
			disabledWithoutSamples: t('dynamicRouting.learning.summaryDisabledNoSamples'),
		},
		runtimeTestId: 'setting-dynamic-routing-learning-runtime',
		topTargetValueClassName: 'break-words',
	});
	const learningDisplay = learningSummaryView.display;
	const { primaryItems: learningPrimaryItems, secondaryItems: learningSecondaryItems } = buildLearningSummarySections({
		items: learningSummaryView.items,
		primaryEntries: [
			...DEFAULT_LEARNING_SUMMARY_SECTION_ENTRIES.primaryEntries.slice(0, 1),
			{
				key: 'scope',
				label: t('dynamicRouting.learning.scopeLabel'),
				value: t('dynamicRouting.learning.scopeValue'),
			},
			...DEFAULT_LEARNING_SUMMARY_SECTION_ENTRIES.primaryEntries.slice(1),
		],
		secondaryEntries: [
			...DEFAULT_LEARNING_SUMMARY_SECTION_ENTRIES.secondaryEntries,
			{
				key: 'details',
				label: t('dynamicRouting.learning.detailsLabel'),
				value: t('dynamicRouting.learning.detailsValue'),
			},
		],
		errorScope: 'dynamic routing learning',
	});
	const handleResetLearning = () => {
		resetLearning.mutate(undefined, {
			onSuccess: () => toast.success(t('dynamicRouting.learning.resetSuccess')),
			onError: (error) => toast.error(t('dynamicRouting.learning.resetFailed'), { description: error instanceof Error ? error.message : undefined }),
		});
	};
	const routingMode = routingModeDraft ?? routingModeSetting;
	const globalBudget = globalBudgetDraft ?? globalBudgetSetting;
	const groupBudget = groupBudgetDraft ?? groupBudgetSetting;
	const channelBudget = channelBudgetDraft ?? channelBudgetSetting;
	const keyBudget = keyBudgetDraft ?? keyBudgetSetting;
	const probeBudget = probeBudgetDraft ?? probeBudgetSetting;

	const handleStringSave = (key: string, value: string, initialValue: string, onSaved?: () => void) => {
		if (value === initialValue) return;

		setSetting.mutate(
			{ key, value },
			{
				onSuccess: () => {
					toast.success(t('saved'));
					onSaved?.();
				},
			}
		);
	};

	const handleBooleanSave = (key: string, value: boolean, initialValue: boolean) => {
		if (value === initialValue) return;

		setSetting.mutate(
			{ key, value: value ? 'true' : 'false' },
			{
				onSuccess: () => {
					toast.success(t('saved'));
					learning.refetch?.();
					if (key === SettingKey.DynamicRoutingHealthEnabled) {
						setHealthEnabledDraft(null);
					}
					if (key === SettingKey.DynamicRoutingLearningEnabled) {
						setLearningEnabledDraft(null);
					}
				},
			}
		);
	};

	const handleNumberSave = (key: string, value: string, initialValue: string) => {
		if (value === initialValue) return;

		setSetting.mutate(
			{ key, value },
			{
				onSuccess: () => {
					toast.success(t('saved'));
					switch (key) {
						case SettingKey.RaceGlobalBudget:
							setGlobalBudgetDraft(null);
							break;
						case SettingKey.RaceGroupBudget:
							setGroupBudgetDraft(null);
							break;
						case SettingKey.RaceChannelBudget:
							setChannelBudgetDraft(null);
							break;
						case SettingKey.RaceKeyBudget:
							setKeyBudgetDraft(null);
							break;
						case SettingKey.RaceProbeBudget:
							setProbeBudgetDraft(null);
							break;
					}
				},
			}
		);
	};

	const budgetFields: BudgetField[] = [
		{
			key: SettingKey.RaceGlobalBudget,
			icon: Gauge,
			label: t('dynamicRouting.globalBudget'),
			hint: t('dynamicRouting.globalBudgetHint'),
			value: globalBudget,
			initialValue: globalBudgetSetting,
			onChange: setGlobalBudgetDraft,
		},
		{
			key: SettingKey.RaceGroupBudget,
			icon: Network,
			label: t('dynamicRouting.groupBudget'),
			hint: t('dynamicRouting.groupBudgetHint'),
			value: groupBudget,
			initialValue: groupBudgetSetting,
			onChange: setGroupBudgetDraft,
		},
		{
			key: SettingKey.RaceChannelBudget,
			icon: Activity,
			label: t('dynamicRouting.channelBudget'),
			hint: t('dynamicRouting.channelBudgetHint'),
			value: channelBudget,
			initialValue: channelBudgetSetting,
			onChange: setChannelBudgetDraft,
		},
		{
			key: SettingKey.RaceKeyBudget,
			icon: Radio,
			label: t('dynamicRouting.keyBudget'),
			hint: t('dynamicRouting.keyBudgetHint'),
			value: keyBudget,
			initialValue: keyBudgetSetting,
			onChange: setKeyBudgetDraft,
		},
		{
			key: SettingKey.RaceProbeBudget,
			icon: Waves,
			label: t('dynamicRouting.probeBudget'),
			hint: t('dynamicRouting.probeBudgetHint'),
			value: probeBudget,
			initialValue: probeBudgetSetting,
			onChange: setProbeBudgetDraft,
		},
	];

	const summary = dynamicRoutingSummary.data;
	const modeCards = [
		{ label: t('dynamicRouting.summary.currentMode'), value: summary ? t(`dynamicRouting.modeOptions.${normalizeDynamicRoutingMode(summary.current_mode)}`) : '-' },
		{ label: t('dynamicRouting.summary.effectiveMode'), value: summary ? t(`dynamicRouting.modeOptions.${normalizeDynamicRoutingMode(summary.effective_mode)}`) : '-' },
		{ label: t('dynamicRouting.summary.decision'), value: summary ? t(`dynamicRouting.summary.decisionValues.${normalizeSummaryDecision(summary.decision)}`) : '-' },
		{ label: t('dynamicRouting.summary.healthSwitch'), value: summary?.health_enabled ? t('dynamicRouting.summary.enabled') : t('dynamicRouting.summary.disabled') },
	];
	const keyMixCards = [
		{ label: t('dynamicRouting.summary.freePublicKeys'), value: summary?.free_public_keys ?? '-' },
		{ label: t('dynamicRouting.summary.paidMeteredKeys'), value: summary?.paid_metered_keys ?? '-' },
		{ label: t('dynamicRouting.summary.privateInternalKeys'), value: summary?.private_internal_keys ?? '-' },
		{ label: t('dynamicRouting.summary.unknownKeys'), value: summary?.unknown_keys ?? '-' },
	];

	const formatDateTime = (value?: string) => {
		if (!value) return t('dynamicRouting.summary.notAvailable');
		const parsed = new Date(value);
		if (Number.isNaN(parsed.getTime())) return value;
		return formatDateTimeByLocale(parsed, locale);
	};

	const summaryStatusLabel = summary
		? t(`dynamicRouting.summary.statusValues.${normalizeSummaryStatus(summary.last_status)}`)
		: t('dynamicRouting.summary.notAvailable');
	const summaryDecisionLabel = summary
		? t(`dynamicRouting.summary.decisionValues.${normalizeSummaryDecision(summary.decision)}`)
		: t('dynamicRouting.summary.notAvailable');
	const summaryBasisLabel = summary?.basis
		? t(`dynamicRouting.summary.basisValues.${normalizeSummaryBasis(summary.basis)}`)
		: t('dynamicRouting.summary.notAvailable');
	const normalizedDecisionReason = normalizeSummaryDecisionReason(summary?.decision_reason);
	const summaryDecisionReasonLabel = normalizedDecisionReason
		? t(`dynamicRouting.summary.decisionReasonValues.${normalizedDecisionReason}`)
		: null;
	const normalizedMessageKey = normalizeSummaryMessage(summary?.last_message);
	const summaryMessageLabel = summary?.last_message
		? normalizedMessageKey
			? t(`dynamicRouting.summary.messageValues.${normalizedMessageKey}`)
			: t('dynamicRouting.summary.messageValues.unmapped')
		: null;
	const summaryMessageDetail = summary?.last_message && !normalizedMessageKey ? summary.last_message : null;
	const summaryMetaItems = [
		{ label: t('dynamicRouting.summary.lastSuccessLabel'), value: formatDateTime(summary?.last_success_at) },
		{ label: t('dynamicRouting.summary.basisLabel'), value: summaryBasisLabel },
		...(summaryDecisionReasonLabel ? [{ label: t('dynamicRouting.summary.decisionReasonLabel'), value: summaryDecisionReasonLabel }] : []),
		...(summaryMessageLabel ? [{ label: t('dynamicRouting.summary.messageLabel'), value: summaryMessageLabel }] : []),
		...(summaryMessageDetail ? [{ label: t('dynamicRouting.summary.messageDetailLabel'), value: summaryMessageDetail }] : []),
	];

	return (
		<div data-testid="setting-dynamic-routing-card" className="octo-setting-card">
			<div className="space-y-2">
				<h2 className="octo-setting-heading">
					<ShieldCheck className="size-4" />
					{t('dynamicRouting.title')}
				</h2>
				<span className="sr-only">{t('dynamicRouting.defaultPathTitle')}</span>
				<span className="sr-only">
					{t('dynamicRouting.defaultPathDesc')}
					{t('dynamicRouting.healthEnabledHint')}
					{t('dynamicRouting.healthEnabledDesc')}
					{t('dynamicRouting.budgetSummaryHint')}
					{t('dynamicRouting.budgetSummaryDesc')}
					{t('dynamicRouting.advancedHint')}
					{t('dynamicRouting.advancedDesc')}
				</span>
			</div>

			<div className="rounded-2xl border border-border/60 bg-muted/20 p-3">
				<div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
					<div>
						<div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
							<span>{t('dynamicRouting.healthEnabled')}</span>
						</div>
					</div>
					<div className="flex items-center gap-3 self-start md:self-center">
						<span className="rounded-full border border-border/60 bg-background/80 px-2.5 py-1 text-xs text-muted-foreground">
							{healthEnabled ? t('dynamicRouting.summary.enabled') : t('dynamicRouting.summary.disabled')}
						</span>
						<Switch
							aria-label={t('dynamicRouting.healthEnabled')}
							checked={healthEnabled}
							onCheckedChange={(checked) => {
								setHealthEnabledDraft(checked);
								handleBooleanSave(SettingKey.DynamicRoutingHealthEnabled, checked, healthEnabledSetting);
							}}
						/>
					</div>
				</div>
			</div>

			<div data-testid="setting-dynamic-routing-learning-card" className="space-y-3 rounded-2xl border border-border/60 bg-muted/20 p-3">
				<div className="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
					<div>
						<div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
							<span>{t('dynamicRouting.learning.title')}</span>
						</div>
					</div>
					<div className="flex items-center gap-3 self-start xl:self-center">
						<span className="rounded-full border border-border/60 bg-background/80 px-2.5 py-1 text-xs text-muted-foreground">
							{learningEnabled ? t('dynamicRouting.learning.enabledBadge') : t('dynamicRouting.learning.disabledBadge')}
						</span>
						<Switch
							aria-label={t('dynamicRouting.learning.switchTitle')}
							checked={learningEnabled}
							onCheckedChange={(checked) => {
								setLearningEnabledDraft(checked);
								handleBooleanSave(SettingKey.DynamicRoutingLearningEnabled, checked, learningEnabledSetting);
							}}
						/>
					</div>
				</div>

				<LearningSummaryPanel
					primaryGrid={{
						items: learningPrimaryItems,
						className: 'grid gap-2 sm:grid-cols-2 xl:grid-cols-4',
						cardClassName: 'rounded-xl border border-border/60 bg-background/80 px-3 py-2',
						valueClassName: 'text-card-foreground',
					}}
					secondaryGrid={{
						items: learningSecondaryItems,
						testId: 'setting-dynamic-routing-learning-summary',
						className: 'grid gap-2 sm:grid-cols-2 xl:grid-cols-3',
						cardClassName: 'rounded-xl border border-border/60 bg-background/80 px-3 py-2',
						valueClassName: 'text-card-foreground',
					}}
					noticeTitle={t('dynamicRouting.learning.summaryTitle')}
					noticeBody={learningSummaryView.notice}
					noticeClassName="rounded-xl border border-emerald-500/20 bg-emerald-500/5 px-3 py-2 text-xs leading-5 text-muted-foreground"
					footer={(
						<LearningSummaryActionBar
							actions={[
								{
									key: 'open-ai',
									label: t('dynamicRouting.learning.open'),
									onClick: openAIAutomationLearning,
									testId: 'setting-dynamic-routing-learning-open-ai',
								},
								{
									key: 'reset',
									label: t('dynamicRouting.learning.reset'),
									onClick: handleResetLearning,
									testId: 'setting-dynamic-routing-learning-reset',
									disabled: resetLearning.isPending || !learningDisplay.canReset,
								},
							]}
							hint={!learningDisplay.canReset ? t('dynamicRouting.learning.resetDisabledHint') : null}
						/>
					)}
				/>
			</div>

			<div className="space-y-3 rounded-2xl border border-border/70 bg-muted/10 p-3">
				<div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
					<div>
						<div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
							<span>{t('dynamicRouting.summary.title')}</span>
						</div>
					</div>
					<div className="text-xs text-muted-foreground">
						{dynamicRoutingSummary.isLoading
							? t('dynamicRouting.summary.loading')
							: `${t('dynamicRouting.summary.lastRunLabel')}: ${formatDateTime(summary?.last_run_at)}`}
					</div>
				</div>

				{dynamicRoutingSummary.isError ? (
					<div className="rounded-xl border border-red-500/20 bg-red-500/5 px-3 py-2 text-xs text-red-600 dark:text-red-400">
						{t('dynamicRouting.summary.loadFailed')}
					</div>
				) : null}

				<div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
					<div className="rounded-xl border border-border/60 bg-background/80 px-3 py-2">
						<div className="text-[11px] text-muted-foreground">{t('dynamicRouting.summary.status')}</div>
						<div className="mt-1 text-sm font-medium text-card-foreground">{summaryStatusLabel}</div>
					</div>
					<div className="rounded-xl border border-border/60 bg-background/80 px-3 py-2">
						<div className="text-[11px] text-muted-foreground">{t('dynamicRouting.summary.enabledChannels')}</div>
						<div className="mt-1 text-sm font-medium text-card-foreground">{summary ? `${summary.enabled_channels}/${summary.channel_count}` : '-'}</div>
					</div>
					<div className="rounded-xl border border-border/60 bg-background/80 px-3 py-2">
						<div className="text-[11px] text-muted-foreground">{t('dynamicRouting.summary.failoverGroups')}</div>
						<div className="mt-1 text-sm font-medium text-card-foreground">{summary ? `${summary.failover_groups}/${summary.group_count}` : '-'}</div>
					</div>
					<div className="rounded-xl border border-border/60 bg-background/80 px-3 py-2">
						<div className="text-[11px] text-muted-foreground">{t('dynamicRouting.summary.lastDecision')}</div>
						<div className="mt-1 text-sm font-medium text-card-foreground">{summaryDecisionLabel}</div>
					</div>
				</div>

				<Accordion type="single" collapsible className="w-full rounded-2xl border border-border/60 bg-background/60 px-4">
					<AccordionItem value="dynamic-routing-summary-details" className="border-none">
						<AccordionTrigger
							data-testid="setting-dynamic-routing-summary-details-trigger"
							className="py-4 text-left hover:no-underline"
							addon={(
								<HelpHint className="mt-1 size-3.5">
									<div className="space-y-1.5">
										<p>{t('dynamicRouting.summary.detailsHint')}</p>
										<p>{t('dynamicRouting.summary.detailsDesc')}</p>
									</div>
								</HelpHint>
							)}
						>
							<div className="space-y-1">
								<div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
									<span>{t('dynamicRouting.summary.detailsTitle')}</span>
								</div>
							</div>
						</AccordionTrigger>
						<AccordionContent className="space-y-4 border-t pb-4 pt-4">
							<div className="space-y-2">
								<div className="text-xs font-medium text-muted-foreground">{t('dynamicRouting.summary.modeSnapshotTitle')}</div>
								<div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
									{modeCards.map((item) => (
										<div key={item.label} className="rounded-xl border border-border/60 bg-background/80 px-3 py-2">
											<div className="text-[11px] text-muted-foreground">{item.label}</div>
											<div className="mt-1 text-sm font-medium text-card-foreground">{item.value}</div>
										</div>
									))}
								</div>
							</div>

							<div className="space-y-2">
								<div className="text-xs font-medium text-muted-foreground">{t('dynamicRouting.summary.keyMixTitle')}</div>
								<div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
									{keyMixCards.map((item) => (
										<div key={item.label} className="rounded-xl border border-border/60 bg-background/80 px-3 py-2">
											<div className="text-[11px] text-muted-foreground">{item.label}</div>
											<div className="mt-1 text-sm font-medium text-card-foreground">{item.value}</div>
										</div>
									))}
								</div>
							</div>

							<div className="space-y-2">
								<div className="text-xs font-medium text-muted-foreground">{t('dynamicRouting.summary.metaTitle')}</div>
								<div className="grid grid-cols-1 gap-2 sm:grid-cols-2 xl:grid-cols-3">
									{summaryMetaItems.map((item) => (
										<div key={item.label} className="rounded-xl border border-border/60 bg-background/80 px-3 py-2">
											<div className="text-[11px] text-muted-foreground">{item.label}</div>
											<div className="mt-1 text-sm text-card-foreground break-words">{item.value}</div>
										</div>
									))}
								</div>
							</div>
						</AccordionContent>
					</AccordionItem>
				</Accordion>
			</div>

			<div className="rounded-2xl border border-border/60 bg-muted/20 p-3">
				<div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
					<div className="space-y-1.5">
						<div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
							<span>{t('dynamicRouting.mode')}</span>
						</div>
					</div>
					<div className="w-full md:w-72">
						<Select
							value={routingMode}
							onValueChange={(value) => {
								const nextMode = normalizeDynamicRoutingMode(value);
								setRoutingModeDraft(nextMode);
								handleStringSave(SettingKey.DynamicRoutingMode, nextMode, routingModeSetting, () => {
									setRoutingModeDraft(null);
								});
							}}
						>
							<SelectTrigger aria-label={t('dynamicRouting.mode')} className="rounded-xl bg-background">
								<SelectValue />
							</SelectTrigger>
							<SelectContent className="rounded-xl">
								{DYNAMIC_ROUTING_MODE_VALUES.map((mode) => (
									<SelectItem key={mode} value={mode} className="rounded-xl">
										{t(`dynamicRouting.modeOptions.${mode}`)}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</div>
				</div>
			</div>

			<div className="space-y-3 rounded-2xl border border-border/60 bg-muted/20 p-3">
				<div>
					<div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
						<Gauge className="h-4 w-4 text-muted-foreground" />
						<span>{t('dynamicRouting.budgetSummaryTitle')}</span>
					</div>
				</div>

				<div className="grid grid-cols-2 gap-2 xl:grid-cols-5">
					{budgetFields.map((field) => (
						<div key={field.key} className="rounded-xl border border-border/60 bg-background/80 px-3 py-2">
							<div className="text-[11px] text-muted-foreground">{field.label}</div>
							<div className="mt-1 text-sm font-medium text-card-foreground">{getDisplayValue(field.value, field.initialValue)}</div>
						</div>
					))}
				</div>
			</div>

			<Accordion type="single" collapsible className="w-full rounded-2xl border border-border/60 bg-background/60 px-3">
				<AccordionItem value="dynamic-routing-advanced" className="border-none">
					<AccordionTrigger
						data-testid="setting-dynamic-routing-advanced-trigger"
						className="py-4 text-left hover:no-underline"
					>
						<div className="space-y-1">
							<div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
								<span>{t('dynamicRouting.advancedTitle')}</span>
							</div>
						</div>
					</AccordionTrigger>
					<AccordionContent className="space-y-4 border-t pb-4 pt-4">
						{budgetFields.map((field) => {
							const Icon = field.icon;

							return (
								<div key={field.key} className="flex flex-col gap-3 rounded-2xl border border-border/50 bg-card px-4 py-4 md:flex-row md:items-center md:justify-between">
									<div className="space-y-1.5">
										<div className="flex items-center gap-3 text-sm font-medium text-card-foreground">
											<Icon className="h-4 w-4 text-muted-foreground" />
											<span>{field.label}</span>
											<HelpHint>{field.hint}</HelpHint>
										</div>
									</div>
									<Input
										aria-label={field.label}
										type="number"
										value={field.value}
										onChange={(event) => field.onChange(event.target.value)}
										onBlur={() => handleNumberSave(field.key, field.value, field.initialValue)}
										placeholder={t('dynamicRouting.budgetPlaceholder')}
										className="w-full rounded-xl md:w-52"
									/>
								</div>
							);
						})}
					</AccordionContent>
				</AccordionItem>
			</Accordion>
		</div>
	);
}
