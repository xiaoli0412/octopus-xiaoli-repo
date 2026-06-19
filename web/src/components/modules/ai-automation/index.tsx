'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import { useTranslations } from 'next-intl';
import { Play, RotateCcw, Settings2, Workflow } from 'lucide-react';

import {
	useActivateStrategyProfile,
	useAIGovernanceLearningSummary,
	useAIGovernanceOverview,
	useApplyGovernanceSession,
	useCreateGovernanceSession,
	useCreateStrategyProfile,
	useExpertPresets,
	useGovernanceRollbackPoints,
	useGovernanceRuntimePolicy,
	useGovernanceSession,
	useGovernanceSessions,
	useReplanGovernanceSession,
	useRollbackGovernanceSession,
	useStrategyProfiles,
	useUpdateGovernanceRuntimePolicy,
	type GovernanceRuntimePolicyView,
	type GovernanceRollbackPointView,
} from '@/api/endpoints/ai-automation';
import { useAPIKeyList } from '@/api/endpoints/apikey';
import { useCapabilityInventory } from '@/api/endpoints/model';
import { SettingKey, useSetSetting, useSettingList } from '@/api/endpoints/setting';
import { PageWrapper } from '@/components/common/PageWrapper';
import { AnimatedNumber } from '@/components/common/AnimatedNumber';
import { toast } from '@/components/common/Toast';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { formatDateTimeByLocale } from '@/lib/locale';
import { cn } from '@/lib/utils';
import { useSettingStore } from '@/stores/setting';

import { consumeAIAutomationFocusTarget } from './focus-target';
import { RunReport } from './run-report';
import { SessionDetail, type WorkspaceTab } from './session-detail';

const DEFAULT_LOCAL_BASE_URL = 'http://127.0.0.1:1088/v1';
const NO_API_KEY_VALUE = '__none__';

const RUNTIME_STRATEGY_LABELS: Record<string, string> = {
	highest_success_rate: '成功率优先',
	balanced_latency: '时延与成功率平衡',
	cost_first: '成本优先',
};

const RUNTIME_DISPATCH_LABELS: Record<string, string> = {
	single_best: '单一最优',
	bounded_parallel: '有限并发',
	round_robin_review: '轮发复核',
};

const SESSION_STATUS_LABELS: Record<string, string> = {
	draft: '草稿',
	planning: '规划中',
	ready: '可应用',
	stale: '需重算',
	applying: '应用中',
	applied: '已应用',
	failed: '失败',
	succeeded: '成功',
	pending: '待处理',
	running: '执行中',
	validating: '校验中',
	rolled_back: '已回滚',
	completed: '已完成',
};

function localizeSessionStatus(status?: string) {
	if (!status) return '-';
	return SESSION_STATUS_LABELS[status] ?? status;
}

function localizeRuntimePolicy(policy?: GovernanceRuntimePolicyView) {
	if (!policy) return '-';
	return `${RUNTIME_STRATEGY_LABELS[policy.strategy] ?? policy.strategy} · ${RUNTIME_DISPATCH_LABELS[policy.dispatch_mode] ?? policy.dispatch_mode}`;
}

function localizeExecutionSourceLabel(mode?: string, label?: string) {
	if (mode === 'manual') return '手动配置';
	if (mode === 'ai_profile') return 'AI 策略方案';
	if (label?.trim()) return label;
	return '-';
}

function localizeKnownText(value?: string) {
	if (!value) return '-';
	return value;
}

function buildSnapshotScopeSummary(snapshot?: {
	channels: number;
	enabled_channels: number;
	groups: number;
	group_items: number;
	route_target_overrides: number;
	missing_prices?: number;
}) {
	if (!snapshot) return '-';
	const parts = [
		`${snapshot.channels} 个渠道 / ${snapshot.enabled_channels} 个启用`,
		`${snapshot.groups} 个分组 / ${snapshot.group_items} 条分组项`,
		`${snapshot.route_target_overrides} 条路由目标覆盖`,
	];
	if (snapshot.missing_prices && snapshot.missing_prices > 0) {
		parts.push(`${snapshot.missing_prices} 个模型缺少价格`);
	}
	return parts.join(' · ');
}

function defaultRuntimePolicy(): GovernanceRuntimePolicyView {
	return {
		strategy: 'highest_success_rate',
		dispatch_mode: 'single_best',
		max_parallel_runs: 2,
		double_review_enabled: false,
		fallback_to_deterministic: true,
		degraded_to_deterministic: true,
	};
}

function StatMiniCard({ title, value, tone = 'default' }: { title: string; value: React.ReactNode; tone?: 'default' | 'emphasis' }) {
	return (
		<div className={cn('rounded-xl border px-2.5 py-2', tone === 'emphasis' ? 'border-primary/20 bg-primary/5' : 'border-card-border/70 bg-background/55')}>
			<div className="text-[11px] text-muted-foreground">{title}</div>
			<div className="mt-0.5 line-clamp-2 break-all text-sm font-semibold text-card-foreground">{value}</div>
		</div>
	);
}

function RuntimePolicyDialog({
	open,
	onOpenChange,
	policy,
	baseURL,
	useLocalDefault,
	apiKeySelection,
	apiKeyOptions,
	model,
	modelOptions,
	onSave,
	isSaving,
}: {
	open: boolean;
	onOpenChange: (value: boolean) => void;
	policy?: GovernanceRuntimePolicyView;
	baseURL: string;
	useLocalDefault: boolean;
	apiKeySelection: string;
	apiKeyOptions: Array<{ value: string; label: string; hint?: string }>;
	model: string;
	modelOptions: string[];
	onSave: (value: { policy: GovernanceRuntimePolicyView; baseURL: string; useLocalDefault: boolean; apiKeySelection: string; model: string }) => Promise<void>;
	isSaving: boolean;
}) {
	const t = useTranslations('aiAutomationV2');
	const [draft, setDraft] = useState<GovernanceRuntimePolicyView | null>(null);
	const [draftBaseURL, setDraftBaseURL] = useState('');
	const [draftUseLocalDefault, setDraftUseLocalDefault] = useState(true);
	const [draftAPIKeySelection, setDraftAPIKeySelection] = useState('local-default');
	const [draftModel, setDraftModel] = useState('');

	useEffect(() => {
		if (policy) setDraft(policy);
	}, [policy]);

	useEffect(() => {
		setDraftBaseURL(baseURL);
	}, [baseURL]);

	useEffect(() => {
		setDraftUseLocalDefault(useLocalDefault);
	}, [useLocalDefault]);

	useEffect(() => {
		setDraftAPIKeySelection(apiKeySelection);
	}, [apiKeySelection]);

	useEffect(() => {
		setDraftModel(model);
	}, [model]);

	const current = draft ?? policy ?? defaultRuntimePolicy();

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="max-w-3xl rounded-[1.6rem] border-card-border p-0">
				<div className="rounded-[1.6rem] bg-card">
					<DialogHeader className="border-b border-card-border px-6 py-5">
						<DialogTitle>{t('settings.title')}</DialogTitle>
						<DialogDescription>{t('settings.desc')}</DialogDescription>
					</DialogHeader>
					<div className="grid gap-4 px-6 py-5 xl:grid-cols-2">
						<div className="rounded-2xl border border-card-border/60 bg-muted/20 px-4 py-3 xl:col-span-2">
							<div className="flex items-center justify-between gap-3">
								<div>
									<div className="text-sm font-medium text-card-foreground">{t('settings.useLocalDefault')}</div>
									<div className="mt-1 text-xs text-muted-foreground">{t('settings.useLocalDefaultDesc')}</div>
								</div>
								<Switch checked={draftUseLocalDefault} onCheckedChange={setDraftUseLocalDefault} />
							</div>
						</div>
						<div className="space-y-2 xl:col-span-2">
							<div className="text-sm font-medium text-card-foreground">{t('settings.requestBaseUrl')}</div>
							<div className="text-xs text-muted-foreground">{t('settings.requestBaseUrlDesc')}</div>
							<Input className="h-11 rounded-xl" value={draftBaseURL} disabled={draftUseLocalDefault} onChange={(event) => setDraftBaseURL(event.target.value)} />
						</div>
						<div className="min-w-0 space-y-2">
							<div className="text-sm font-medium text-card-foreground">{t('settings.requestApiKey')}</div>
							<div className="text-xs text-muted-foreground">{t('settings.requestApiKeyDesc')}</div>
							<Select value={draftAPIKeySelection} onValueChange={setDraftAPIKeySelection}>
								<SelectTrigger className="h-11 rounded-xl"><SelectValue /></SelectTrigger>
								<SelectContent>
									{apiKeyOptions.map((option) => (
										<SelectItem key={option.value} value={option.value}>{option.hint ? `${option.label} · ${option.hint}` : option.label}</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>
						<div className="min-w-0 space-y-2">
							<div className="text-sm font-medium text-card-foreground">{t('settings.requestModel')}</div>
							<div className="text-xs text-muted-foreground">{t('settings.requestModelDesc')}</div>
							<Select value={draftModel} onValueChange={setDraftModel}>
								<SelectTrigger className="h-11 rounded-xl"><SelectValue /></SelectTrigger>
								<SelectContent>
									{modelOptions.map((option) => (
										<SelectItem key={option} value={option}>{option}</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>
						<div className="min-w-0 space-y-2">
							<div className="text-sm font-medium text-card-foreground">{t('settings.strategy')}</div>
							<Select value={current.strategy} onValueChange={(value) => setDraft({ ...current, strategy: value })}>
								<SelectTrigger className="h-11 rounded-xl"><SelectValue /></SelectTrigger>
								<SelectContent>
									<SelectItem value="highest_success_rate">{t('settings.strategyOptions.highest_success_rate')}</SelectItem>
									<SelectItem value="balanced_latency">{t('settings.strategyOptions.balanced_latency')}</SelectItem>
									<SelectItem value="cost_first">{t('settings.strategyOptions.cost_first')}</SelectItem>
								</SelectContent>
							</Select>
						</div>
						<div className="min-w-0 space-y-2">
							<div className="text-sm font-medium text-card-foreground">{t('settings.dispatchMode')}</div>
							<Select value={current.dispatch_mode} onValueChange={(value) => setDraft({ ...current, dispatch_mode: value })}>
								<SelectTrigger className="h-11 rounded-xl"><SelectValue /></SelectTrigger>
								<SelectContent>
									<SelectItem value="single_best">{t('settings.dispatchOptions.single_best')}</SelectItem>
									<SelectItem value="bounded_parallel">{t('settings.dispatchOptions.bounded_parallel')}</SelectItem>
									<SelectItem value="round_robin_review">{t('settings.dispatchOptions.round_robin_review')}</SelectItem>
								</SelectContent>
							</Select>
						</div>
						<div className="space-y-2 xl:col-span-2">
							<div className="text-sm font-medium text-card-foreground">{t('settings.maxParallelRuns')}</div>
							<Input type="number" min={1} className="h-11 rounded-xl" value={current.max_parallel_runs} onChange={(event) => setDraft({ ...current, max_parallel_runs: Math.max(1, Number(event.target.value) || 1) })} />
						</div>
						<div className="rounded-2xl border border-card-border/60 bg-muted/20 px-4 py-3 md:col-span-2">
							<div className="flex items-center justify-between gap-3">
								<div>
									<div className="text-sm font-medium text-card-foreground">{t('settings.doubleReview')}</div>
									<div className="mt-1 text-xs text-muted-foreground">{t('settings.doubleReviewDesc')}</div>
								</div>
								<Switch checked={current.double_review_enabled} onCheckedChange={(checked) => setDraft({ ...current, double_review_enabled: checked })} />
							</div>
						</div>
						<div className="rounded-2xl border border-card-border/60 bg-muted/20 px-4 py-3 md:col-span-2">
							<div className="flex items-center justify-between gap-3">
								<div>
									<div className="text-sm font-medium text-card-foreground">{t('settings.fallback')}</div>
									<div className="mt-1 text-xs text-muted-foreground">{t('settings.fallbackDesc')}</div>
								</div>
								<Switch checked={current.fallback_to_deterministic} onCheckedChange={(checked) => setDraft({ ...current, fallback_to_deterministic: checked })} />
							</div>
						</div>
					</div>
					<div className="flex justify-end gap-3 border-t border-card-border px-6 py-4">
						<Button variant="outline" className="rounded-xl" onClick={() => onOpenChange(false)}>{t('settings.cancel')}</Button>
						<Button className="rounded-xl" disabled={isSaving} onClick={() => void onSave({ policy: current, baseURL: draftBaseURL, useLocalDefault: draftUseLocalDefault, apiKeySelection: draftAPIKeySelection, model: draftModel })}>{isSaving ? t('settings.saving') : t('settings.save')}</Button>
					</div>
				</div>
			</DialogContent>
		</Dialog>
	);
}

const QUICK_GOALS = [
	{ key: 'global', label: 'quickGoals.global', value: '统一评估分组路由、价格覆盖、动态路由和运行时设置。' },
	{ key: 'routing', label: 'quickGoals.routing', value: '先整理当前分组与路由。' },
	{ key: 'pricing', label: 'quickGoals.pricing', value: '先检查当前模型的价格覆盖缺口。' },
];

export function AIAutomation() {
	const t = useTranslations('aiAutomationV2');
	const locale = useSettingStore((state) => state.locale);
	const overview = useAIGovernanceOverview();
	const sessionsQuery = useGovernanceSessions();
	const presetsQuery = useExpertPresets();
	const profilesQuery = useStrategyProfiles();
	const apiKeysQuery = useAPIKeyList();
	const capabilityInventoryQuery = useCapabilityInventory();
	const learningSummaryQuery = useAIGovernanceLearningSummary();
	const runtimePolicyQuery = useGovernanceRuntimePolicy();
	const createSession = useCreateGovernanceSession();
	const replanSession = useReplanGovernanceSession();
	const applySession = useApplyGovernanceSession();
	const rollbackSession = useRollbackGovernanceSession();
	const createProfile = useCreateStrategyProfile();
	const activateProfile = useActivateStrategyProfile();
	const updateRuntimePolicy = useUpdateGovernanceRuntimePolicy();
	const setSetting = useSetSetting();

	const [goal, setGoal] = useState('');
	const [workspaceTab, setWorkspaceTab] = useState<WorkspaceTab>('preview');
	const [selectedSessionID, setSelectedSessionID] = useState<number | undefined>();
	const [newProfileName, setNewProfileName] = useState('');
	const [settingsOpen, setSettingsOpen] = useState(false);
	const learningSectionRef = useRef<HTMLDivElement | null>(null);
	const pendingFocusTargetRef = useRef<'learning' | null>(null);

	const sessions = sessionsQuery.data ?? [];
	const selectedSessionFromList = useMemo(() => sessions.find((item) => item.id === selectedSessionID) ?? sessions[0], [selectedSessionID, sessions]);
	const selectedSessionQuery = useGovernanceSession(selectedSessionFromList?.id);
	const rollbackPointsQuery = useGovernanceRollbackPoints(selectedSessionFromList?.id);
	const selectedSession = selectedSessionQuery.data;
	const hasSession = Boolean(selectedSessionFromList?.id);
	const rollbackPoints = rollbackPointsQuery.data ?? selectedSession?.rollback_points ?? [];
	const presets = presetsQuery.data ?? [];
	const strategyProfiles = profilesQuery.data ?? [];
	const apiKeys = apiKeysQuery.data ?? [];
	const capabilityInventory = capabilityInventoryQuery.data;
	const settings = useSettingList().data ?? [];
	const settingMap = useMemo(() => new Map(settings.map((item) => [item.key, item.value])), [settings]);
	const learningSummary = learningSummaryQuery.data ?? overview.data?.learning;
	const runtimePolicy = runtimePolicyQuery.data ?? overview.data?.runtime_policy;

	useEffect(() => {
		if (!selectedSessionID && sessions[0]?.id) {
			setSelectedSessionID(sessions[0].id);
		}
	}, [selectedSessionID, sessions]);

	useEffect(() => {
		const targetKey = consumeAIAutomationFocusTarget();
		if (!targetKey) return;
		pendingFocusTargetRef.current = targetKey;
		setWorkspaceTab('history');
	}, []);

	useEffect(() => {
		if (workspaceTab !== 'history') return;
		if (pendingFocusTargetRef.current !== 'learning') return;
		const target = learningSectionRef.current;
		if (!target) return;
		if (workspaceTab === 'history') {
			target.scrollIntoView({ behavior: 'smooth', block: 'start' });
			pendingFocusTargetRef.current = null;
		}
	}, [workspaceTab]);

	const currentGoal = goal.trim();
	const currentSummary = localizeKnownText(selectedSession?.operator_summary || overview.data?.recent_session?.operator_summary || t('states.idleSummary'));
	const currentRuntimePolicyText = localizeRuntimePolicy(runtimePolicy);
	const currentSourceLabel = localizeExecutionSourceLabel(overview.data?.execution_source.mode, overview.data?.execution_source.label);
	const currentScopeSummary = buildSnapshotScopeSummary(selectedSession?.snapshot_summary);
	const currentUseLocalDefault = (settingMap.get(SettingKey.AIAutomationUseLocalDefault) ?? `${overview.data?.execution_source.use_local_default ?? true}`) === 'true';
	const currentBaseURL = currentUseLocalDefault
		? DEFAULT_LOCAL_BASE_URL
		: (settingMap.get(SettingKey.AIAutomationBaseUrl)?.trim() || overview.data?.execution_source.base_url || DEFAULT_LOCAL_BASE_URL);
	const currentSavedAPIKey = settingMap.get(SettingKey.AIAutomationAPIKey)?.trim() || '';
	const currentSavedModel = settingMap.get(SettingKey.AIAutomationModel)?.trim() || '';
	const rankedModel = useMemo(() => {
		const target = learningSummary?.top_target?.split('/')[0]?.trim();
		if (target) return target;
		return overview.data?.execution_source.model || 'gpt-4o';
	}, [learningSummary?.top_target, overview.data?.execution_source.model]);
	const modelOptions = useMemo(() => {
		const values = new Set<string>();
		if (rankedModel) values.add(rankedModel);
		if (overview.data?.execution_source.model) values.add(overview.data.execution_source.model);
		selectedSession?.plan.mutations.forEach((mutation) => {
			if (mutation.group_item_attach?.model_name) values.add(mutation.group_item_attach.model_name);
			if (mutation.group_item_detach?.model_name) values.add(mutation.group_item_detach.model_name);
			if (mutation.route_target_override_upsert?.model_name) values.add(mutation.route_target_override_upsert.model_name);
			if (mutation.route_target_override_delete?.model_name) values.add(mutation.route_target_override_delete.model_name);
			if (mutation.llm_price_upsert?.name) values.add(mutation.llm_price_upsert.name);
		});
		(capabilityInventory?.selectable_models ?? []).forEach((item) => {
			if (item.name) values.add(item.name);
		});
		if (currentSavedModel) values.add(currentSavedModel);
		if (values.size === 0) values.add('gpt-4o');
		return Array.from(values).sort((a, b) => a.localeCompare(b));
	}, [capabilityInventory?.selectable_models, currentSavedModel, overview.data?.execution_source.model, rankedModel, selectedSession?.plan.mutations]);
	const apiKeyOptions = useMemo(() => {
		const items = apiKeys
			.filter((row) => row.enabled)
			.map((item, index) => ({
				value: item.api_key,
				label: item.name,
				hint: index === 0 ? t('settings.firstAvailableApiKeyHint') : undefined,
			}));
		if (currentSavedAPIKey && !items.some((item) => item.value === currentSavedAPIKey)) {
			items.unshift({ value: currentSavedAPIKey, label: t('settings.savedApiKey'), hint: t('settings.savedApiKeyHint') });
		}
		if (items.length === 0) {
			items.push({ value: NO_API_KEY_VALUE, label: t('settings.noApiKey'), hint: undefined });
		}
		return items;
	}, [apiKeys, currentSavedAPIKey, t]);
	const defaultAPIKeySelection = useMemo(() => {
		if (currentSavedAPIKey) return currentSavedAPIKey;
		return apiKeyOptions.find((item) => item.value !== NO_API_KEY_VALUE)?.value ?? NO_API_KEY_VALUE;
	}, [apiKeyOptions, currentSavedAPIKey]);
	const selectedModelValue = currentSavedModel || rankedModel;

	const handleCreateSession = async () => {
		if (!currentGoal) {
			toast.error(t('toast.goalRequired'));
			return;
		}
		try {
			const result = await createSession.mutateAsync({ goal: currentGoal, expert_preset_id: presets.find((item) => item.id === 'balanced')?.id ?? presets[0]?.id });
			setSelectedSessionID(result.id);
			setWorkspaceTab('preview');
			if (!newProfileName.trim()) {
				setNewProfileName(`方案 ${result.id}`);
			}
		} catch (error) {
			toast.error(t('toast.createFailed'), { description: error instanceof Error ? error.message : undefined });
		}
	};

	const handleReplan = async () => {
		if (!selectedSessionFromList?.id) return;
		try {
			await replanSession.mutateAsync(selectedSessionFromList.id);
		} catch (error) {
			toast.error(t('toast.replanFailed'), { description: error instanceof Error ? error.message : undefined });
		}
	};

	const handleApply = async () => {
		if (!selectedSessionFromList?.id) return;
		try {
			await applySession.mutateAsync(selectedSessionFromList.id);
		} catch (error) {
			toast.error(t('toast.applyFailed'), { description: error instanceof Error ? error.message : undefined });
		}
	};

	const handleRollback = async (point: GovernanceRollbackPointView) => {
		if (!selectedSessionFromList?.id) return;
		try {
			await rollbackSession.mutateAsync({ id: selectedSessionFromList.id, rollback_point_id: point.id });
		} catch (error) {
			toast.error(t('toast.rollbackFailed'), { description: error instanceof Error ? error.message : undefined });
		}
	};

	const handleCreateProfile = async () => {
		if (!selectedSessionFromList?.id) {
			toast.error(t('toast.profileNeedsSession'));
			return;
		}
		if (!newProfileName.trim()) {
			toast.error(t('toast.profileNameRequired'));
			return;
		}
		try {
			await createProfile.mutateAsync({ session_id: selectedSessionFromList.id, name: newProfileName.trim() });
		} catch (error) {
			toast.error(t('toast.profileCreateFailed'), { description: error instanceof Error ? error.message : undefined });
		}
	};

	const handleActivateProfile = async (id: number) => {
		try {
			await activateProfile.mutateAsync(id);
		} catch (error) {
			toast.error(t('toast.profileActivateFailed'), { description: error instanceof Error ? error.message : undefined });
		}
	};

	const handleSaveRuntimePolicy = async ({
		policy,
		baseURL,
		useLocalDefault,
		apiKeySelection,
		model,
	}: {
		policy: GovernanceRuntimePolicyView;
		baseURL: string;
		useLocalDefault: boolean;
		apiKeySelection: string;
		model: string;
	}) => {
		try {
			await updateRuntimePolicy.mutateAsync(policy);
			await setSetting.mutateAsync({ key: SettingKey.AIAutomationUseLocalDefault, value: useLocalDefault ? 'true' : 'false' });
			await setSetting.mutateAsync({ key: SettingKey.AIAutomationBaseUrl, value: useLocalDefault ? DEFAULT_LOCAL_BASE_URL : (baseURL.trim() || currentBaseURL || DEFAULT_LOCAL_BASE_URL) });
			await setSetting.mutateAsync({ key: SettingKey.AIAutomationAPIKey, value: apiKeySelection === NO_API_KEY_VALUE ? '' : apiKeySelection.trim() });
			await setSetting.mutateAsync({ key: SettingKey.AIAutomationModel, value: model.trim() || selectedModelValue });
			setSettingsOpen(false);
		} catch (error) {
			toast.error(t('toast.runtimePolicySaveFailed'), { description: error instanceof Error ? error.message : undefined });
		}
	};

	return (
		<div className="octo-workbench">
			<PageWrapper className="space-y-3" data-testid="ai-automation-page" disableAnimations>
				<section className="octo-section">
					<div className="octo-toolbar">
						<div className="flex flex-wrap items-center gap-3">
							<div className="flex items-center gap-2 text-lg font-semibold tracking-tight text-card-foreground">
								<Workflow className="size-5 text-primary" />
								{t('hero.title')}
							</div>
							<div className="octo-chip">{t('hero.badge')}</div>
						</div>
						<Button type="button" variant="outline" className="h-9 rounded-xl" onClick={() => setSettingsOpen(true)}>
							<Settings2 className="size-4" />
							{t('settings.open')}
						</Button>
					</div>

					<div className="mt-3 space-y-3">
						<div className="rounded-xl border border-card-border bg-background/55 p-2.5">
							<div className="flex flex-wrap items-center justify-between gap-2">
								<div className="text-sm font-semibold text-card-foreground">{t('main.goalTitle')}</div>
								<div className="hidden text-xs text-muted-foreground md:block">{t('main.goalDesc')}</div>
							</div>
							<div className="mt-2 grid gap-2 xl:grid-cols-[minmax(0,1fr)_auto]">
								<Input
									data-testid="ai-governance-goal-input"
									className="h-9 rounded-xl"
									placeholder={t('main.goalPlaceholder')}
									value={goal}
									onChange={(event) => setGoal(event.target.value)}
								/>
								<div className="flex flex-wrap gap-2">
									<Button data-testid="ai-governance-create-button" className="h-9 rounded-xl px-3" onClick={handleCreateSession} disabled={createSession.isPending}>
										<Play className="size-4" />
										{createSession.isPending ? t('actions.creating') : t('actions.create')}
									</Button>
									{hasSession ? (
										<Button data-testid="ai-governance-replan-button" variant="outline" className="h-9 rounded-xl px-3" onClick={handleReplan} disabled={replanSession.isPending}>
											<RotateCcw className="size-4" />
											{t('actions.replan')}
										</Button>
									) : null}
								</div>
							</div>
							<div className="mt-2 flex flex-wrap gap-2">
								{QUICK_GOALS.map((item) => (
									<button
										key={item.key}
										type="button"
										className="rounded-full border border-card-border bg-background px-2.5 py-1 text-xs text-muted-foreground transition hover:border-primary/30 hover:text-card-foreground"
										onClick={() => setGoal(item.value)}
									>
										{t(item.label)}
									</button>
								))}
							</div>
						</div>

						<div className="grid gap-2 min-[375px]:grid-cols-2 xl:grid-cols-4">
							<StatMiniCard title={t('summary.currentGoal')} value={currentGoal || t('summary.goalEmpty')} tone="emphasis" />
							<StatMiniCard title={t('summary.activePlan')} value={currentSummary} />
							<StatMiniCard title={t('summary.lastApply')} value={selectedSession?.applied_at ? formatDateTimeByLocale(selectedSession.applied_at, locale) : t('summary.notApplied')} />
							<StatMiniCard title={t('sidebar.runtimePolicy')} value={currentRuntimePolicyText} />
						</div>

						{selectedSession ? (
							<div className="rounded-xl border border-card-border bg-muted/15 p-2.5">
								<div className="grid gap-2 min-[375px]:grid-cols-2 xl:grid-cols-4">
									<StatMiniCard title={t('runReport.currentStatus')} value={localizeSessionStatus(selectedSession.status)} />
									<StatMiniCard title={t('runReport.sourceLabel')} value={currentSourceLabel} />
									<StatMiniCard title={t('main.scopeLabel')} value={currentScopeSummary} />
									<StatMiniCard title={t('main.riskLabel')} value={localizeKnownText(selectedSession?.risk_summary || t('states.noRisk'))} />
								</div>
							</div>
						) : null}
					</div>
				</section>

				<RunReport session={selectedSession} locale={locale} onApply={handleApply} isApplying={applySession.isPending} />

				<section className="octo-panel p-3">
					<div ref={learningSectionRef} data-ai-focus-target="learning" className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
						<StatMiniCard title={t('history.learningEnabled')} value={learningSummary?.enabled ? t('sidebar.learningOn') : t('sidebar.learningOff')} />
						<StatMiniCard title={t('history.learningSamples')} value={<><AnimatedNumber value={learningSummary?.sample_count ?? 0} /> 条</>} />
						<StatMiniCard title={t('history.learningTopTarget')} value={learningSummary?.top_target ? localizeKnownText(learningSummary.top_target) : '-'} />
						<StatMiniCard title={t('history.learningUpdated')} value={learningSummary?.last_sample_at ? formatDateTimeByLocale(String(learningSummary.last_sample_at), locale) : '-'} />
					</div>
				</section>

				<SessionDetail
					tab={workspaceTab}
					onTabChange={setWorkspaceTab}
					session={selectedSession}
					sessions={sessions}
					rollbackPoints={rollbackPoints}
					strategyProfiles={strategyProfiles}
					presets={presets}
					learningSummary={learningSummary}
					selectedSessionID={selectedSessionFromList?.id}
					onSelectSession={setSelectedSessionID}
					newProfileName={newProfileName}
					onNewProfileNameChange={setNewProfileName}
					onCreateProfile={handleCreateProfile}
					onActivateProfile={handleActivateProfile}
					onRollback={handleRollback}
					isCreatingProfile={createProfile.isPending}
					isApplying={applySession.isPending}
					locale={locale}
				/>

				<RuntimePolicyDialog
					open={settingsOpen}
					onOpenChange={setSettingsOpen}
					policy={runtimePolicy}
					baseURL={currentBaseURL}
					useLocalDefault={currentUseLocalDefault}
					apiKeySelection={defaultAPIKeySelection}
					apiKeyOptions={apiKeyOptions}
					model={selectedModelValue}
					modelOptions={modelOptions}
					onSave={handleSaveRuntimePolicy}
					isSaving={updateRuntimePolicy.isPending || setSetting.isPending}
				/>
			</PageWrapper>
		</div>
	);
}
