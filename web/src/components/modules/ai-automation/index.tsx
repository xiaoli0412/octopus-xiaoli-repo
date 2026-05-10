'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import { useTranslations } from 'next-intl';
import { BrainCircuit, ClipboardList, History, Layers3, Play, RotateCcw, Settings2, ShieldCheck, Sparkles, Undo2, Workflow } from 'lucide-react';

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
	type GovernanceDomainPlanView,
	type GovernanceMutation,
	type GovernanceRollbackPointView,
	type GovernanceRuntimePolicyView,
	type GovernanceSessionSummary,
} from '@/api/endpoints/ai-automation';
import { useAPIKeyList } from '@/api/endpoints/apikey';
import { PageWrapper } from '@/components/common/PageWrapper';
import { toast } from '@/components/common/Toast';
import { AnimatedNumber } from '@/components/common/AnimatedNumber';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { formatDateTimeByLocale } from '@/lib/locale';
import { cn } from '@/lib/utils';
import { useSettingStore } from '@/stores/setting';
import { SettingKey, useSetSetting, useSettingList } from '@/api/endpoints/setting';

import { consumeAIAutomationFocusTarget } from './focus-target';

type WorkspaceTab = 'preview' | 'profiles' | 'rollback' | 'history' | 'expert';

const IS_TEST_ENV = process.env.NODE_ENV === 'test';

const AI_AUTOMATION_TEXT: Record<string, string> = {
	'hero.badge': 'AI 自动化总控台',
	'hero.title': '一句目标，统一生成可应用的 AI 自动化方案',
	'hero.desc': '默认同时评估分组路由、价格覆盖、动态路由与 AI 运行时设置。主链尽量简单，复杂控制收进设置与专业模式。',
	'settings.open': '自动化设置',
	'settings.title': '自动化设置',
	'settings.desc': '这里集中保存 AI 处理时的调度策略、并发上限和降级路径，作为治理配置快照，不占主工作区。',
	'settings.useLocalDefault': '使用本机默认地址',
	'settings.useLocalDefaultDesc': '开启后自动填入当前服务的本机地址；关闭后可手动修改请求地址。',
	'settings.strategy': '治理策略',
	'settings.strategyOptions.highest_success_rate': '成功率优先',
	'settings.strategyOptions.balanced_latency': '时延与成功率平衡',
	'settings.strategyOptions.cost_first': '成本优先',
	'settings.dispatchMode': '分发模式',
	'settings.dispatchOptions.single_best': '单一最优',
	'settings.dispatchOptions.bounded_parallel': '有限并发',
	'settings.dispatchOptions.round_robin_review': '轮发复核',
	'settings.maxParallelRuns': '最大并发数',
	'settings.doubleReview': '双阶段复核',
	'settings.doubleReviewDesc': '重要治理任务可启用第二次审阅，降低误判风险。',
	'settings.fallback': '规则降级',
	'settings.fallbackDesc': '当 AI 不可用时，自动回退到确定性规则规划，避免流程中断。',
	'settings.requestBaseUrl': '请求地址',
	'settings.requestBaseUrlDesc': '默认使用当前服务本机地址，也支持改成自定义 AI 请求地址。',
	'settings.requestApiKey': '请求密钥',
	'settings.requestApiKeyDesc': '默认使用第一个可用分发密钥；如需指定其它密钥，可在这里切换。',
	'settings.savedApiKey': '已保存密钥',
	'settings.noApiKey': '暂无可用密钥',
	'settings.requestModel': '请求模型',
	'settings.requestModelDesc': '默认优先选动态学习里成功率最高的模型，也支持手动改成其它模型。',
	'settings.localDefaultKey': '本机默认',
	'settings.cancel': '取消',
	'settings.saving': '保存中...',
	'settings.save': '保存设置',
	'quickGoals.global': '全局总调配',
	'quickGoals.routing': '先整路由分组',
	'quickGoals.pricing': '先补价格缺口',
	'main.goalTitle': '输入目标',
	'main.goalDesc': '直接说你要达成的自动化目标，系统会生成分域方案与变更预览。',
	'main.goalPlaceholder': '例如：把当前可用模型统一整理成运维治理组，并补齐价格缺口与治理策略。',
	'main.aiSummaryTitle': '当前总控状态',
	'main.aiSummaryDesc': '把当前会话、来源、范围与风险压缩到一组状态卡中，减少来回切换。',
	'main.scopeLabel': '当前覆盖范围',
	'main.riskLabel': '风险提示',
	'actions.creating': '生成中...',
	'actions.create': '生成方案',
	'actions.replan': '重新规划',
	'actions.applying': '应用中...',
	'actions.apply': '一键应用',
	'actions.rollback': '回滚到此点',
	'summary.currentGoal': '当前目标',
	'summary.goalEmpty': '还没有输入目标',
	'summary.activePlan': '当前方案摘要',
	'summary.lastApply': '最近一次应用',
	'summary.notApplied': '还没有应用记录',
	'sidebar.title': '状态带',
	'sidebar.runtimePolicy': '治理策略',
	'sidebar.learningOn': '已启用',
	'sidebar.learningOff': '已关闭',
	'workspace.plan': '主方案',
	'workspace.preview': '方案预览',
	'workspace.profiles': '已保存方案',
	'workspace.rollback': '回滚点',
	'workspace.history': '运行记录',
	'workspace.expert': '专业模式',
	'workspace.previewTitle': '方案预览',
	'workspace.previewDesc': '先确认影响范围、风险提示和结构化变更，再决定是否应用。',
	'workspace.profilesTitle': '已保存方案',
	'workspace.profilesDesc': '把当前会话沉淀成可切换的方案资产，后续可以重复启用。',
	'workspace.rollbackTitle': '回滚点',
	'workspace.rollbackDesc': '每次应用前都会保存快照，需要时可以从这里恢复。',
	'workspace.historyTitle': '运行记录',
	'workspace.historyDesc': '查看历史会话、学习摘要和最近治理结果。',
	'workspace.expertTitle': '专业模式',
	'workspace.expertDesc': '为高阶用户保留更深入的审阅深度、清理策略与绑定同步配置。',
	'workspace.detailsTitle': '当前会话',
	'workspace.detailsDesc': '展示当前选中会话的快照校验、应用记录和关键元信息。',
	'preview.groups': '分组变更',
	'preview.items': '分组项变更',
	'preview.overrides': '路由与策略变更',
	'preview.status': '应用状态',
	'preview.ready': '可应用',
	'preview.blocked': '存在阻断',
	'profiles.namePlaceholder': '给这份方案起个名字，方便后续切换。',
	'profiles.create': '保存当前方案',
	'profiles.noSummary': '这份方案还没有补充摘要。',
	'profiles.active': '当前生效',
	'profiles.activate': '启用方案',
	'profiles.empty': '还没有保存方案。先生成并确认一份总控方案。',
	'history.learningTitle': '动态学习摘要',
	'history.learningDesc': '学习数据只作为运行时推荐参考，不直接覆盖渠道、分组或密钥配置。',
	'history.learningEnabled': '学习开关',
	'history.learningSamples': '样本数量',
	'history.learningTopTarget': '当前最高分对象',
	'history.learningUpdated': '最近采样时间',
	'expert.reviewDepth': '审阅深度',
	'expert.cleanup': '清理陈旧项',
	'expert.syncBindings': '同步绑定关系',
	'expert.enabled': '开启',
	'expert.disabled': '关闭',
	'details.currentSession': '会话编号',
	'details.snapshotChecksum': '快照校验值',
	'details.applyHistory': '应用记录',
	'states.idleSummary': '当前还没有可展示的会话摘要。',
	'states.noSnapshot': '还没有快照摘要。',
	'states.noRisk': '当前没有额外风险提示。',
	'states.noPreview': '还没有生成预览内容。',
	'states.noHistory': '还没有历史会话。',
	'states.noRollbackPoints': '当前会话还没有回滚点。',
	'states.noApplyHistory': '还没有应用记录。',
	'toast.goalRequired': '请先输入治理目标。',
	'toast.createFailed': '生成治理方案失败',
	'toast.replanSuccess': '已重新生成方案',
	'toast.replanFailed': '重新生成方案失败',
	'toast.applySuccess': '方案已开始应用',
	'toast.applyFailed': '应用方案失败',
	'toast.rollbackSuccess': '回滚已执行',
	'toast.rollbackFailed': '回滚失败',
	'toast.profileNeedsSession': '请先选中一个会话，再保存方案。',
	'toast.profileNameRequired': '请先填写方案名称。',
	'toast.profileCreated': '方案已保存',
	'toast.profileCreateFailed': '保存方案失败',
	'toast.profileActivated': '方案已启用',
	'toast.profileActivateFailed': '启用方案失败',
	'toast.runtimePolicySaved': '治理策略已保存',
	'toast.runtimePolicySaveFailed': '保存治理策略失败',
};

const DEFAULT_LOCAL_BASE_URL = 'http://127.0.0.1:1088/v1';
const NO_API_KEY_VALUE = '__none__';

const DOMAIN_STATUS_LABELS: Record<string, string> = {
	ready: '可执行',
	idle: '暂无变更',
	blocked: '存在阻断',
	failed: '异常',
};

const PROFILE_STATUS_LABELS: Record<string, string> = {
	draft: '草稿',
	ready: '可启用',
	active: '已启用',
	archived: '已归档',
	invalid: '无效',
};

const PRESET_NAME_LABELS: Record<string, string> = {
	balanced: '均衡总控',
	conservative: '保守审阅',
	deep_review: '深度审阅',
};

const PRESET_DESC_LABELS: Record<string, string> = {
	balanced: '平衡分组路由、价格覆盖、动态路由与运行时策略，适合作为默认总控方案。',
	conservative: '优先减少写入和清理动作，适合谨慎上线或先做保守收口。',
	deep_review: '输出更完整的巡检结论、漂移说明和清理建议，适合深度复核。',
};

const DOMAIN_TITLE_LABELS: Record<string, string> = {
	routing_grouping: '分组路由',
	pricing: '价格覆盖',
	dynamic_routing: '动态路由',
	runtime_policy: 'AI 运行时',
};

const DOMAIN_SUMMARY_LABELS: Record<string, string> = {
	routing_grouping: '治理组、分组项与路由目标。',
	pricing: '模型缺价与计费缺口。',
	dynamic_routing: '模式、学习开关与预算。',
	runtime_policy: '分发模式、并发与降级。',
};

const KNOWN_TEXT_REPLACEMENTS: Array<[RegExp, string]> = [
	[/AI Governance Managed/gi, 'AI 自动化治理组'],
	[/Manual AI endpoint/gi, '手动 AI 来源'],
	[/AI profile runtime source/gi, 'AI 方案运行时来源'],
	[/Applied governance changes/gi, '已应用治理变更'],
	[/Applied governance plan/gi, '已应用治理方案'],
	[/Governance apply failed/gi, '治理应用失败'],
	[/Balanced governance/gi, '均衡总控'],
	[/Conservative review/gi, '保守审阅'],
	[/Deep review/gi, '深度审阅'],
	[/Explicit apply is required\.?/gi, '需要显式点击应用。'],
	[/Explicit apply required\.?/gi, '需要显式点击应用。'],
	[/route target overrides/gi, '路由目标覆盖'],
	[/route target/gi, '路由目标'],
	[/(\d+) findings \/ (\d+) decisions/gi, '$1 个发现 / $2 个决策'],
	[/(\d+) channels \/ (\d+) enabled/gi, '$1 个渠道 / $2 个启用'],
	[/(\d+) groups \/ (\d+) group items/gi, '$1 个分组 / $2 条分组项'],
	[/(\d+) route target overrides/gi, '$1 条路由目标覆盖'],
	[/(\d+) models missing price/gi, '$1 个模型缺少价格'],
	[/learning samples (\d+)/gi, '学习样本 $1'],
	[/double-review/gi, '双阶段复核'],
	[/typed mutation/gi, '结构化变更'],
];

function resolveFallbackText(translate: (key: string) => string, namespace: string, key: string, fallbackMap: Record<string, string>) {
	const fallback = fallbackMap[key];
	try {
		const resolved = translate(key);
		if (IS_TEST_ENV) return resolved;
		if (resolved === `${namespace}.${key}` || resolved === key) {
			return fallback ?? resolved;
		}
		return resolved;
	} catch {
		return fallback ?? key;
	}
}

function localizeKnownText(value?: string) {
	if (!value) return '-';
	if (IS_TEST_ENV) return value;
	let next = value;
	for (const [pattern, replacement] of KNOWN_TEXT_REPLACEMENTS) {
		next = next.replace(pattern, replacement);
	}
	return next;
}

function localizeDomainStatus(status?: string) {
	if (!status) return '-';
	return DOMAIN_STATUS_LABELS[status] ?? localizeSessionStatus(status);
}

function localizeProfileStatus(status?: string) {
	if (!status) return '-';
	return PROFILE_STATUS_LABELS[status] ?? localizeSessionStatus(status);
}

function localizePresetName(id: string, fallback?: string) {
	return PRESET_NAME_LABELS[id] ?? fallback ?? id;
}

function localizePresetDescription(id: string, fallback?: string) {
	return PRESET_DESC_LABELS[id] ?? localizeKnownText(fallback);
}

function localizeDomainTitle(key?: string, fallback?: string) {
	if (key && DOMAIN_TITLE_LABELS[key]) return DOMAIN_TITLE_LABELS[key];
	return localizeKnownText(fallback);
}

function localizeDomainSummary(key?: string, fallback?: string) {
	if (key && DOMAIN_SUMMARY_LABELS[key]) return DOMAIN_SUMMARY_LABELS[key];
	return localizeKnownText(fallback);
}

function buildSnapshotScopeSummary(snapshot?: {
	channels: number;
	enabled_channels: number;
	groups: number;
	group_items: number;
	route_target_overrides: number;
	missing_prices?: number;
}) {
	if (!snapshot) return AI_AUTOMATION_TEXT['states.noSnapshot'];
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

const WORKSPACE_TABS: Array<{ key: WorkspaceTab; icon: typeof ClipboardList; label: string }> = [
	{ key: 'preview', icon: ClipboardList, label: 'workspace.preview' },
	{ key: 'profiles', icon: Layers3, label: 'workspace.profiles' },
	{ key: 'rollback', icon: Undo2, label: 'workspace.rollback' },
	{ key: 'history', icon: History, label: 'workspace.history' },
	{ key: 'expert', icon: BrainCircuit, label: 'workspace.expert' },
];

const QUICK_GOALS = [
	{ key: 'global', label: 'quickGoals.global', value: '请一次性整理分组路由、价格缺口、动态路由策略和 AI 运行时策略。' },
	{ key: 'routing', label: 'quickGoals.routing', value: '请先整理当前分组与路由，把可用模型统一收口到治理组。' },
	{ key: 'pricing', label: 'quickGoals.pricing', value: '请检查当前已配置模型的价格完整性，并给出可补全的缺口方案。' },
];

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

const EXPERT_DEPTH_LABELS: Record<string, string> = {
	standard: '标准审阅',
	deep: '深度审阅',
	light: '轻量审阅',
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
	if (label?.trim()) return localizeKnownText(label);
	return '-';
}

function localizePresetDepth(value?: string) {
	if (!value) return '-';
	return EXPERT_DEPTH_LABELS[value] ?? value;
}

function compactSessionLabel(session?: GovernanceSessionSummary | null) {
	if (!session) return '-';
	return `#${session.id} · ${localizeSessionStatus(session.status)}`;
}

function panelText(key: string) {
	return AI_AUTOMATION_TEXT[key] ?? key;
}

function statusTone(status?: string) {
	switch (status) {
		case 'ready':
			return 'border-primary/30 bg-primary/10 text-primary';
		case 'applied':
			return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300';
		case 'stale':
			return 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300';
		case 'failed':
			return 'border-destructive/30 bg-destructive/10 text-destructive';
		case 'applying':
			return 'border-accent/30 bg-accent/10 text-accent-foreground';
		default:
			return 'border-card-border bg-muted/40 text-muted-foreground';
	}
}

function WorkspacePanel({ title, description, children, testId }: { title: string; description?: string; children: React.ReactNode; testId?: string }) {
	return (
		<section data-testid={testId} className="octo-panel min-h-0 overflow-hidden">
			<div className="border-b border-card-border px-4 py-3.5 sm:px-5">
				<div className="text-sm font-semibold text-card-foreground">{title}</div>
				{description ? <div className="mt-1 text-xs leading-5 text-muted-foreground">{description}</div> : null}
			</div>
			<div className="min-h-0 overflow-y-auto px-4 py-4 sm:px-5">{children}</div>
		</section>
	);
}

function MutationRow({ mutation }: { mutation: GovernanceMutation }) {
	const mutationTypeLabels: Record<string, string> = {
		group_upsert: '分组更新',
		group_item_attach: '挂入分组',
		group_item_detach: '移出分组',
		group_item_reorder: '调整顺序',
		route_target_override_upsert: '写入路由目标',
		route_target_override_delete: '删除路由目标',
		llm_price_upsert: '补全价格',
		dynamic_routing_setting_set: '更新动态路由',
		runtime_policy_set: '更新治理策略',
		strategy_profile_activate: '切换策略',
	};

	return (
		<div className="rounded-2xl border border-card-border/70 bg-muted/20 px-3.5 py-3">
			<div className="flex flex-wrap items-center gap-2">
				<div className="rounded-full border border-card-border bg-background px-2 py-0.5 text-[11px] text-muted-foreground">{mutationTypeLabels[mutation.type] ?? mutation.type}</div>
				<div className="break-all text-sm font-medium text-card-foreground">{localizeKnownText(mutation.summary)}</div>
			</div>
		</div>
	);
}

function DomainCard({ domain }: { domain: GovernanceDomainPlanView }) {
	return (
		<div className="rounded-3xl border border-card-border bg-card p-4 shadow-sm">
			<div className="flex flex-wrap items-start justify-between gap-3">
				<div className="min-w-0 flex-1">
					<div className="truncate text-sm font-semibold text-card-foreground">{localizeDomainTitle(domain.key, domain.title)}</div>
					<div className="mt-1 line-clamp-2 text-xs leading-5 text-muted-foreground">{localizeDomainSummary(domain.key, domain.summary)}</div>
				</div>
				<div className={cn('shrink-0 rounded-full border px-2.5 py-1 text-[11px]', statusTone(domain.status))}>{localizeDomainStatus(domain.status)}</div>
			</div>
			<div className="mt-4 grid grid-cols-2 gap-2.5">
				<div className="rounded-xl border border-card-border/60 bg-muted/20 px-3 py-2">
					<div className="text-[11px] text-muted-foreground">发现问题</div>
					<div className="mt-1 text-[1.1rem] font-semibold text-foreground"><AnimatedNumber value={domain.finding_count} /></div>
				</div>
				<div className="rounded-xl border border-card-border/60 bg-muted/20 px-3 py-2">
					<div className="text-[11px] text-muted-foreground">待应用变更</div>
					<div className="mt-1 text-[1.1rem] font-semibold text-foreground"><AnimatedNumber value={domain.mutation_count} /></div>
				</div>
			</div>
		</div>
	);
}

function StatMiniCard({ title, value, tone = 'default' }: { title: string; value: React.ReactNode; tone?: 'default' | 'emphasis' }) {
	return (
		<div className={cn('rounded-2xl border p-3', tone === 'emphasis' ? 'border-primary/20 bg-primary/8' : 'border-card-border/70 bg-muted/20')}>
			<div className="text-[11px] text-muted-foreground">{title}</div>
			<div className="mt-1 break-all text-sm font-semibold text-card-foreground">{value}</div>
		</div>
	);
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
	const tx = (key: string) => resolveFallbackText(t, 'aiAutomationV2', key, AI_AUTOMATION_TEXT);
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
						<DialogTitle>{tx('settings.title')}</DialogTitle>
						<DialogDescription>{tx('settings.desc')}</DialogDescription>
					</DialogHeader>
					<div className="grid gap-4 px-6 py-5 xl:grid-cols-2">
						<div className="rounded-2xl border border-card-border/60 bg-muted/20 px-4 py-3 xl:col-span-2">
							<div className="flex items-center justify-between gap-3">
								<div>
									<div className="text-sm font-medium text-card-foreground">{tx('settings.useLocalDefault')}</div>
									<div className="mt-1 text-xs text-muted-foreground">{tx('settings.useLocalDefaultDesc')}</div>
								</div>
								<Switch checked={draftUseLocalDefault} onCheckedChange={setDraftUseLocalDefault} />
							</div>
						</div>
						<div className="space-y-2 xl:col-span-2">
							<div className="text-sm font-medium text-card-foreground">{tx('settings.requestBaseUrl')}</div>
							<div className="text-xs text-muted-foreground">{tx('settings.requestBaseUrlDesc')}</div>
							<Input className="h-11 rounded-xl" value={draftBaseURL} disabled={draftUseLocalDefault} onChange={(event) => setDraftBaseURL(event.target.value)} />
						</div>
						<div className="space-y-2 min-w-0">
							<div className="text-sm font-medium text-card-foreground">{tx('settings.requestApiKey')}</div>
							<div className="text-xs text-muted-foreground">{tx('settings.requestApiKeyDesc')}</div>
							<Select value={draftAPIKeySelection} onValueChange={setDraftAPIKeySelection}>
								<SelectTrigger className="h-11 rounded-xl"><SelectValue /></SelectTrigger>
								<SelectContent>
									{apiKeyOptions.map((option) => (
										<SelectItem key={option.value} value={option.value}>{option.hint ? `${option.label} · ${option.hint}` : option.label}</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>
						<div className="space-y-2 min-w-0">
							<div className="text-sm font-medium text-card-foreground">{tx('settings.requestModel')}</div>
							<div className="text-xs text-muted-foreground">{tx('settings.requestModelDesc')}</div>
							<Select value={draftModel} onValueChange={setDraftModel}>
								<SelectTrigger className="h-11 rounded-xl"><SelectValue /></SelectTrigger>
								<SelectContent>
									{modelOptions.map((option) => (
										<SelectItem key={option} value={option}>{option}</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>
						<div className="space-y-2 min-w-0">
							<div className="text-sm font-medium text-card-foreground">{tx('settings.strategy')}</div>
							<Select value={current.strategy} onValueChange={(value) => setDraft({ ...current, strategy: value })}>
								<SelectTrigger className="h-11 rounded-xl"><SelectValue /></SelectTrigger>
								<SelectContent>
									<SelectItem value="highest_success_rate">{tx('settings.strategyOptions.highest_success_rate')}</SelectItem>
									<SelectItem value="balanced_latency">{tx('settings.strategyOptions.balanced_latency')}</SelectItem>
									<SelectItem value="cost_first">{tx('settings.strategyOptions.cost_first')}</SelectItem>
								</SelectContent>
							</Select>
						</div>
						<div className="space-y-2 min-w-0">
							<div className="text-sm font-medium text-card-foreground">{tx('settings.dispatchMode')}</div>
							<Select value={current.dispatch_mode} onValueChange={(value) => setDraft({ ...current, dispatch_mode: value })}>
								<SelectTrigger className="h-11 rounded-xl"><SelectValue /></SelectTrigger>
								<SelectContent>
									<SelectItem value="single_best">{tx('settings.dispatchOptions.single_best')}</SelectItem>
									<SelectItem value="bounded_parallel">{tx('settings.dispatchOptions.bounded_parallel')}</SelectItem>
									<SelectItem value="round_robin_review">{tx('settings.dispatchOptions.round_robin_review')}</SelectItem>
								</SelectContent>
							</Select>
						</div>
						<div className="space-y-2 xl:col-span-2">
							<div className="text-sm font-medium text-card-foreground">{tx('settings.maxParallelRuns')}</div>
							<Input type="number" min={1} className="h-11 rounded-xl" value={current.max_parallel_runs} onChange={(event) => setDraft({ ...current, max_parallel_runs: Math.max(1, Number(event.target.value) || 1) })} />
						</div>
						<div className="rounded-2xl border border-card-border/60 bg-muted/20 px-4 py-3 md:col-span-2">
							<div className="flex items-center justify-between gap-3">
								<div>
									<div className="text-sm font-medium text-card-foreground">{tx('settings.doubleReview')}</div>
									<div className="mt-1 text-xs text-muted-foreground">{tx('settings.doubleReviewDesc')}</div>
								</div>
								<Switch checked={current.double_review_enabled} onCheckedChange={(checked) => setDraft({ ...current, double_review_enabled: checked })} />
							</div>
						</div>
						<div className="rounded-2xl border border-card-border/60 bg-muted/20 px-4 py-3 md:col-span-2">
							<div className="flex items-center justify-between gap-3">
								<div>
									<div className="text-sm font-medium text-card-foreground">{tx('settings.fallback')}</div>
									<div className="mt-1 text-xs text-muted-foreground">{tx('settings.fallbackDesc')}</div>
								</div>
								<Switch checked={current.fallback_to_deterministic} onCheckedChange={(checked) => setDraft({ ...current, fallback_to_deterministic: checked })} />
							</div>
						</div>
					</div>
					<div className="flex justify-end gap-3 border-t border-card-border px-6 py-4">
						<Button variant="outline" className="rounded-xl" onClick={() => onOpenChange(false)}>{tx('settings.cancel')}</Button>
						<Button className="rounded-xl" disabled={isSaving} onClick={() => void onSave({ policy: current, baseURL: draftBaseURL, useLocalDefault: draftUseLocalDefault, apiKeySelection: draftAPIKeySelection, model: draftModel })}>{isSaving ? tx('settings.saving') : tx('settings.save')}</Button>
					</div>
				</div>
			</DialogContent>
		</Dialog>
	);
}

export function AIAutomation() {
	const t = useTranslations('aiAutomationV2');
	const tx = (key: string) => resolveFallbackText(t, 'aiAutomationV2', key, AI_AUTOMATION_TEXT);
	const locale = useSettingStore((state) => state.locale);
	const overview = useAIGovernanceOverview();
	const sessionsQuery = useGovernanceSessions();
	const presetsQuery = useExpertPresets();
	const profilesQuery = useStrategyProfiles();
	const apiKeysQuery = useAPIKeyList();
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
	const [selectedPresetID, setSelectedPresetID] = useState('balanced');
	const [newProfileName, setNewProfileName] = useState('');
	const [settingsOpen, setSettingsOpen] = useState(false);
	const learningSectionRef = useRef<HTMLDivElement | null>(null);
	const pendingFocusTargetRef = useRef<'learning' | null>(null);

	const sessions = sessionsQuery.data ?? [];
	const selectedSessionFromList = useMemo(() => sessions.find((item) => item.id === selectedSessionID) ?? sessions[0], [selectedSessionID, sessions]);
	const selectedSessionQuery = useGovernanceSession(selectedSessionFromList?.id);
	const rollbackPointsQuery = useGovernanceRollbackPoints(selectedSessionFromList?.id);
	const selectedSession = selectedSessionQuery.data;
	const rollbackPoints = rollbackPointsQuery.data ?? selectedSession?.rollback_points ?? [];
	const presets = presetsQuery.data ?? [];
	const strategyProfiles = profilesQuery.data ?? [];
	const apiKeys = apiKeysQuery.data ?? [];
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
		target.scrollIntoView({ behavior: 'smooth', block: 'start' });
		pendingFocusTargetRef.current = null;
	}, [workspaceTab]);

	const selectedPreset = presets.find((item) => item.id === selectedPresetID) ?? presets[0];
	const currentGoal = goal.trim();
	const currentSummary = localizeKnownText(selectedSession?.operator_summary || overview.data?.recent_session?.operator_summary || tx('states.idleSummary'));
	const currentDomains = selectedSession?.plan.domains ?? [];
	const currentStatusText = localizeSessionStatus(selectedSession?.status || overview.data?.recent_session?.status);
	const currentRuntimePolicyText = localizeRuntimePolicy(runtimePolicy);
	const currentSourceLabel = localizeExecutionSourceLabel(overview.data?.execution_source.mode, overview.data?.execution_source.label);
	const currentManagedGroup = overview.data?.managed_group_name || '-';
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
		if (values.size === 0) values.add('gpt-4o');
		return Array.from(values);
	}, [overview.data?.execution_source.model, rankedModel, selectedSession?.plan.mutations]);
	const apiKeyOptions = useMemo(() => {
		const items = apiKeys
			.filter((row) => row.enabled)
			.map((item, index) => ({
				value: item.api_key,
				label: item.name,
				hint: index === 0 ? '第 1 个可用密钥' : undefined,
			}));
		if (currentSavedAPIKey && !items.some((item) => item.value === currentSavedAPIKey)) {
			items.unshift({ value: currentSavedAPIKey, label: tx('settings.savedApiKey'), hint: '当前已保存' });
		}
		if (items.length === 0) {
			items.push({ value: NO_API_KEY_VALUE, label: tx('settings.noApiKey'), hint: undefined });
		}
		return items;
	}, [apiKeys, currentSavedAPIKey, tx]);
	const defaultAPIKeySelection = useMemo(() => {
		if (currentSavedAPIKey) return currentSavedAPIKey;
		return apiKeyOptions.find((item) => item.value !== NO_API_KEY_VALUE)?.value ?? NO_API_KEY_VALUE;
	}, [apiKeyOptions, apiKeys, currentSavedAPIKey]);
	const selectedModelValue = currentSavedModel || rankedModel;
	const settingsBaseUrlValue = currentBaseURL;

	const handleCreateSession = async () => {
		if (!currentGoal) {
			toast.error(tx('toast.goalRequired'));
			return;
		}
		try {
			const result = await createSession.mutateAsync({ goal: currentGoal, expert_preset_id: selectedPresetID });
			setSelectedSessionID(result.id);
			setWorkspaceTab('preview');
			if (!newProfileName.trim()) {
				setNewProfileName(`方案 ${result.id}`);
			}
		} catch (error) {
			toast.error(tx('toast.createFailed'), { description: error instanceof Error ? error.message : undefined });
		}
	};

	const handleReplan = async () => {
		if (!selectedSessionFromList?.id) return;
		try {
			await replanSession.mutateAsync(selectedSessionFromList.id);
			toast.success(tx('toast.replanSuccess'));
		} catch (error) {
			toast.error(tx('toast.replanFailed'), { description: error instanceof Error ? error.message : undefined });
		}
	};

	const handleApply = async () => {
		if (!selectedSessionFromList?.id) return;
		try {
			await applySession.mutateAsync(selectedSessionFromList.id);
			toast.success(tx('toast.applySuccess'));
		} catch (error) {
			toast.error(tx('toast.applyFailed'), { description: error instanceof Error ? error.message : undefined });
		}
	};

	const handleRollback = async (point: GovernanceRollbackPointView) => {
		if (!selectedSessionFromList?.id) return;
		try {
			await rollbackSession.mutateAsync({ id: selectedSessionFromList.id, rollback_point_id: point.id });
			toast.success(tx('toast.rollbackSuccess'));
		} catch (error) {
			toast.error(tx('toast.rollbackFailed'), { description: error instanceof Error ? error.message : undefined });
		}
	};

	const handleCreateProfile = async () => {
		if (!selectedSessionFromList?.id) {
			toast.error(tx('toast.profileNeedsSession'));
			return;
		}
		if (!newProfileName.trim()) {
			toast.error(tx('toast.profileNameRequired'));
			return;
		}
		try {
			await createProfile.mutateAsync({ session_id: selectedSessionFromList.id, name: newProfileName.trim() });
			toast.success(tx('toast.profileCreated'));
		} catch (error) {
			toast.error(tx('toast.profileCreateFailed'), { description: error instanceof Error ? error.message : undefined });
		}
	};

	const handleActivateProfile = async (id: number) => {
		try {
			await activateProfile.mutateAsync(id);
			toast.success(tx('toast.profileActivated'));
		} catch (error) {
			toast.error(tx('toast.profileActivateFailed'), { description: error instanceof Error ? error.message : undefined });
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
			toast.success(tx('toast.runtimePolicySaved'));
			setSettingsOpen(false);
		} catch (error) {
			toast.error(tx('toast.runtimePolicySaveFailed'), { description: error instanceof Error ? error.message : undefined });
		}
	};

	return (
		<div className="h-full min-h-0 overflow-y-auto overscroll-contain rounded-t-3xl">
			<PageWrapper className="space-y-5 pb-24 md:pb-4" data-testid="ai-automation-page">
				<section className="rounded-3xl border border-card-border bg-card p-5 sm:p-6">
					<div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
						<div className="max-w-4xl">
							<div className="inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-[11px] text-primary">
								<Workflow className="size-3.5" />
								{tx('hero.badge')}
							</div>
							<div className="mt-3 text-2xl font-semibold tracking-tight text-card-foreground sm:text-[2rem]">{tx('hero.title')}</div>
							<div className="mt-2 text-sm leading-6 text-muted-foreground">{tx('hero.desc')}</div>
						</div>
						<Button type="button" variant="outline" className="rounded-xl" onClick={() => setSettingsOpen(true)}>
							<Settings2 className="size-4" />
							{tx('settings.open')}
						</Button>
					</div>

					<div className="mt-5 space-y-4">
						<div className="space-y-4">
							<div className="rounded-3xl border border-card-border bg-muted/20 p-4">
								<div className="text-sm font-semibold text-card-foreground">{tx('main.goalTitle')}</div>
								<div className="mt-1 text-xs text-muted-foreground">{tx('main.goalDesc')}</div>
								<Input data-testid="ai-governance-goal-input" className="mt-4 h-12 rounded-xl" placeholder={tx('main.goalPlaceholder')} value={goal} onChange={(event) => setGoal(event.target.value)} />
								<div className="mt-3 flex flex-wrap gap-2">
									{QUICK_GOALS.map((item) => (
										<button key={item.key} type="button" className="rounded-full border border-card-border bg-background px-3 py-2 text-xs text-muted-foreground transition hover:border-primary/30 hover:text-card-foreground" onClick={() => setGoal(item.value)}>
											{tx(item.label)}
										</button>
									))}
								</div>
								<div className="mt-4 flex flex-wrap gap-3">
									<Button data-testid="ai-governance-create-button" className="rounded-xl" onClick={handleCreateSession} disabled={createSession.isPending}><Play className="size-4" />{createSession.isPending ? tx('actions.creating') : tx('actions.create')}</Button>
									<Button data-testid="ai-governance-replan-button" variant="outline" className="rounded-xl" onClick={handleReplan} disabled={!selectedSessionFromList?.id || replanSession.isPending}><RotateCcw className="size-4" />{tx('actions.replan')}</Button>
									<Button data-testid="ai-governance-apply-button" variant="outline" className="rounded-xl border-emerald-500/30 bg-emerald-500/10 text-emerald-700 hover:bg-emerald-500/15 dark:text-emerald-300" onClick={handleApply} disabled={!selectedSession?.preview.can_apply || applySession.isPending}><ShieldCheck className="size-4" />{applySession.isPending ? tx('actions.applying') : tx('actions.apply')}</Button>
								</div>
							</div>

							<div className="grid gap-3 min-[375px]:grid-cols-2 xl:grid-cols-4">
								<StatMiniCard title={tx('summary.currentGoal')} value={currentGoal || tx('summary.goalEmpty')} tone="emphasis" />
								<StatMiniCard title={tx('summary.activePlan')} value={currentSummary} />
								<StatMiniCard title={tx('summary.lastApply')} value={selectedSession?.applied_at ? formatDateTimeByLocale(selectedSession.applied_at, locale) : tx('summary.notApplied')} />
								<StatMiniCard title={tx('sidebar.runtimePolicy')} value={currentRuntimePolicyText} />
							</div>

							<section className="grid gap-4 md:grid-cols-2 2xl:grid-cols-4">
								{currentDomains.length ? currentDomains.map((domain) => <DomainCard key={domain.key} domain={domain} />) : [
									{ key: 'routing_grouping', title: '分组与路由', summary: '整理托管治理组、分组项和路由目标覆盖。', status: 'idle', finding_count: 0, mutation_count: 0 },
									{ key: 'pricing', title: '价格覆盖', summary: '检查已配置模型是否存在缺价与计费缺口。', status: 'idle', finding_count: 0, mutation_count: 0 },
									{ key: 'dynamic_routing', title: '动态路由', summary: '统一查看模式、学习开关和竞速预算。', status: 'idle', finding_count: 0, mutation_count: 0 },
									{ key: 'runtime_policy', title: 'AI 运行时', summary: '管理分发模式、并发上限和规则降级策略。', status: 'idle', finding_count: 0, mutation_count: 0 },
								].map((domain) => <DomainCard key={domain.key} domain={domain as GovernanceDomainPlanView} />)}
							</section>
						</div>
						<div className="rounded-3xl border border-card-border bg-muted/15 p-4">
							<div className="grid gap-3 md:grid-cols-2 2xl:grid-cols-5">
								<StatMiniCard title="当前状态" value={currentStatusText} />
								<StatMiniCard title="当前来源" value={currentSourceLabel} />
								<StatMiniCard title={tx('main.scopeLabel')} value={currentScopeSummary} />
								<StatMiniCard title={tx('main.riskLabel')} value={localizeKnownText(selectedSession?.risk_summary || tx('states.noRisk'))} />
								<StatMiniCard title="托管治理组" value={localizeKnownText(currentManagedGroup)} />
							</div>
						</div>
					</div>
				</section>

				<section className="space-y-4">
					<div className="flex flex-wrap gap-2">{WORKSPACE_TABS.map((tab) => { const Icon = tab.icon; return <button key={tab.key} data-testid={`ai-governance-tab-${tab.key}`} type="button" className={cn('inline-flex h-10 items-center gap-2 rounded-full border px-4 text-sm transition', workspaceTab === tab.key ? 'border-primary/30 bg-primary/10 text-primary' : 'border-card-border bg-card text-muted-foreground hover:border-primary/20 hover:text-card-foreground')} onClick={() => setWorkspaceTab(tab.key)}><Icon className="size-4" />{tx(tab.label)}</button>; })}</div>
					<div className="grid gap-5 2xl:grid-cols-[minmax(0,1.35fr)_minmax(320px,0.65fr)]">
						{workspaceTab === 'preview' ? <WorkspacePanel title={tx('workspace.previewTitle')} description={tx('workspace.previewDesc')} testId="ai-governance-workspace-preview"><div className="space-y-4 xl:col-span-2"><div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4"><StatMiniCard title={tx('preview.groups')} value={<AnimatedNumber value={selectedSession?.preview.impact_counts.groups ?? 0} />} /><StatMiniCard title={tx('preview.items')} value={<AnimatedNumber value={selectedSession?.preview.impact_counts.items ?? 0} />} /><StatMiniCard title={tx('preview.overrides')} value={<AnimatedNumber value={selectedSession?.preview.impact_counts.overrides ?? 0} />} /><StatMiniCard title={tx('preview.status')} value={selectedSession?.preview.can_apply ? tx('preview.ready') : tx('preview.blocked')} /></div><div className="grid gap-3 md:grid-cols-2">{selectedSession?.preview.summary_lines?.length ? selectedSession.preview.summary_lines.map((line, index) => <div key={index} className="rounded-2xl border border-card-border/70 bg-background/70 px-3 py-2 text-sm text-card-foreground">{localizeKnownText(line)}</div>) : null}{selectedSession?.preview.risk_notes?.map((line, index) => <div key={`risk-${index}`} className="rounded-2xl border border-amber-500/20 bg-amber-500/8 px-3 py-2 text-sm text-card-foreground">{localizeKnownText(line)}</div>)}{selectedSession?.preview.apply_blockers?.map((line, index) => <div key={`blocker-${index}`} className="rounded-2xl border border-destructive/20 bg-destructive/8 px-3 py-2 text-sm text-card-foreground">{localizeKnownText(line)}</div>)}</div><div className="space-y-3">{selectedSession?.preview.mutations?.length ? selectedSession.preview.mutations.map((mutation, index) => <MutationRow key={`${mutation.type}-${index}`} mutation={mutation} />) : <div className="rounded-xl border border-dashed border-card-border bg-muted/20 px-4 py-6 text-sm text-muted-foreground">{tx('states.noPreview')}</div>}</div></div></WorkspacePanel> : null}
						{workspaceTab === 'profiles' ? <WorkspacePanel title={tx('workspace.profilesTitle')} description={tx('workspace.profilesDesc')} testId="ai-governance-workspace-profiles"><div className="space-y-4 xl:col-span-2"><div className="flex flex-col gap-3 md:flex-row"><Input className="h-11 rounded-xl" placeholder={tx('profiles.namePlaceholder')} value={newProfileName} onChange={(event) => setNewProfileName(event.target.value)} /><Button variant="outline" className="rounded-xl" onClick={handleCreateProfile}>{tx('profiles.create')}</Button></div><div className="space-y-3">{strategyProfiles.length ? strategyProfiles.map((profile) => <div key={profile.id} className="rounded-2xl border border-card-border/70 bg-muted/20 p-4"><div className="flex flex-wrap items-center justify-between gap-3"><div className="min-w-0 flex-1"><div className="break-all text-sm font-medium text-card-foreground">{localizeKnownText(profile.name)}</div><div className="mt-1 break-all text-xs text-muted-foreground">{localizeKnownText(profile.summary || tx('profiles.noSummary'))}</div></div><div className="flex items-center gap-2"><div className={cn('rounded-full border px-3 py-1 text-[11px]', profile.is_active ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300' : 'border-card-border bg-background text-muted-foreground')}>{profile.is_active ? tx('profiles.active') : localizeProfileStatus(profile.status)}</div><Button variant="outline" className="rounded-xl" onClick={() => handleActivateProfile(profile.id)} disabled={profile.is_active}>{tx('profiles.activate')}</Button></div></div></div>) : <div className="rounded-xl border border-dashed border-card-border bg-muted/20 px-4 py-6 text-sm text-muted-foreground">{tx('profiles.empty')}</div>}</div></div></WorkspacePanel> : null}
						{workspaceTab === 'rollback' ? <WorkspacePanel title={tx('workspace.rollbackTitle')} description={tx('workspace.rollbackDesc')} testId="ai-governance-workspace-rollback"><div className="space-y-3">{rollbackPoints.length ? rollbackPoints.map((point) => <div key={point.id} className="rounded-2xl border border-card-border/70 bg-muted/20 p-4"><div className="flex flex-wrap items-center justify-between gap-3"><div><div className="text-sm font-medium text-card-foreground">#{point.id}</div><div className="mt-1 break-all text-xs text-muted-foreground">{localizeKnownText(point.summary)}</div><div className="mt-1 text-[11px] text-muted-foreground">{formatDateTimeByLocale(point.created_at, locale)}</div></div><Button variant="outline" className="rounded-xl" onClick={() => handleRollback(point)}>{tx('actions.rollback')}</Button></div></div>) : <div className="rounded-xl border border-dashed border-card-border bg-muted/20 px-4 py-6 text-sm text-muted-foreground">{tx('states.noRollbackPoints')}</div>}</div></WorkspacePanel> : null}
						{workspaceTab === 'history' ? <WorkspacePanel title={tx('workspace.historyTitle')} description={tx('workspace.historyDesc')} testId="ai-governance-workspace-history"><div className="space-y-5"><div className="space-y-3">{sessions.length ? sessions.map((session) => <button key={session.id} type="button" className={cn('w-full rounded-2xl border p-4 text-left transition', selectedSessionFromList?.id === session.id ? 'border-primary/30 bg-primary/10 text-primary' : 'border-card-border bg-muted/20 text-card-foreground hover:border-primary/20')} onClick={() => setSelectedSessionID(session.id)}><div className="flex flex-wrap items-center justify-between gap-2"><div className="break-all text-sm font-medium">#{session.id} · {localizeKnownText(session.goal)}</div><div className={cn('rounded-full border px-2.5 py-1 text-[11px]', statusTone(session.status))}>{localizeSessionStatus(session.status)}</div></div><div className="mt-2 break-all text-xs leading-5 text-muted-foreground">{localizeKnownText(session.operator_summary)}</div></button>) : <div className="rounded-xl border border-dashed border-card-border bg-muted/20 px-4 py-6 text-sm text-muted-foreground">{tx('states.noHistory')}</div>}</div><div ref={learningSectionRef} data-ai-focus-target="learning" className="rounded-2xl border border-card-border/70 bg-muted/20 p-4"><div className="flex items-center gap-2 text-sm font-semibold text-card-foreground"><Sparkles className="size-4 text-primary" />{tx('history.learningTitle')}</div><div className="mt-2 text-xs leading-5 text-muted-foreground">{tx('history.learningDesc')}</div><div className="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-4"><StatMiniCard title={tx('history.learningEnabled')} value={learningSummary?.enabled ? tx('sidebar.learningOn') : tx('sidebar.learningOff')} /><StatMiniCard title={tx('history.learningSamples')} value={<><AnimatedNumber value={learningSummary?.sample_count ?? 0} /> 条</>} /><StatMiniCard title={tx('history.learningTopTarget')} value={localizeKnownText(learningSummary?.top_target || '-')} /><StatMiniCard title={tx('history.learningUpdated')} value={learningSummary?.last_sample_at ? formatDateTimeByLocale(new Date(learningSummary.last_sample_at * 1000).toISOString(), locale) : '-'} /></div></div></div></WorkspacePanel> : null}
						{workspaceTab === 'expert' ? <WorkspacePanel title={tx('workspace.expertTitle')} description={tx('workspace.expertDesc')} testId="ai-governance-workspace-expert"><div className="space-y-4 xl:col-span-2"><div className="grid gap-3 sm:grid-cols-3">{presets.map((preset) => <button key={preset.id} type="button" className={cn('rounded-2xl border p-4 text-left transition', selectedPresetID === preset.id ? 'border-primary/30 bg-primary/10 text-primary' : 'border-card-border bg-muted/20 text-card-foreground hover:border-primary/20')} onClick={() => setSelectedPresetID(preset.id)}><div className="text-sm font-medium">{localizePresetName(preset.id, preset.name)}</div><div className="mt-1 text-xs leading-5 text-muted-foreground">{localizePresetDescription(preset.id, preset.description)}</div></button>)}</div>{selectedPreset ? <div className="rounded-2xl border border-card-border/70 bg-muted/20 p-4"><div className="text-sm font-semibold text-card-foreground">{localizePresetName(selectedPreset.id, selectedPreset.name)}</div><div className="mt-2 break-all text-sm leading-6 text-muted-foreground">{localizePresetDescription(selectedPreset.id, selectedPreset.description)}</div><div className="mt-3 grid gap-3 sm:grid-cols-3"><StatMiniCard title={tx('expert.reviewDepth')} value={localizePresetDepth(selectedPreset.review_depth)} /><StatMiniCard title={tx('expert.cleanup')} value={selectedPreset.cleanup_stale ? tx('expert.enabled') : tx('expert.disabled')} /><StatMiniCard title={tx('expert.syncBindings')} value={selectedPreset.sync_bindings ? tx('expert.enabled') : tx('expert.disabled')} /></div></div> : null}</div></WorkspacePanel> : null}
						<WorkspacePanel title={tx('workspace.detailsTitle')} description={tx('workspace.detailsDesc')} testId="ai-governance-workspace-details"><div className="space-y-4 xl:col-span-2"><div className="rounded-2xl border border-card-border/70 bg-muted/20 p-4"><div className="text-[11px] text-muted-foreground">{tx('details.currentSession')}</div><div className="mt-2 text-sm font-medium text-card-foreground">{selectedSession ? `#${selectedSession.id}` : '-'}</div><div className="mt-1 break-all text-xs leading-5 text-muted-foreground">{localizeKnownText(selectedSession?.operator_summary || tx('states.idleSummary'))}</div></div><div className="rounded-2xl border border-card-border/70 bg-muted/20 p-4"><div className="text-[11px] text-muted-foreground">{tx('details.snapshotChecksum')}</div><div className="mt-2 break-all font-mono text-xs leading-6 text-card-foreground">{selectedSession?.snapshot_checksum || '-'}</div></div><div className="rounded-2xl border border-card-border/70 bg-muted/20 p-4"><div className="text-[11px] text-muted-foreground">{tx('details.applyHistory')}</div><div className="mt-3 space-y-2">{selectedSession?.apply_runs?.length ? selectedSession.apply_runs.map((run) => <div key={run.id} className="rounded-xl border border-card-border/60 bg-background/75 p-3 text-xs leading-5 text-card-foreground">#{run.id} · {localizeSessionStatus(run.status)}<div className="mt-1 break-all text-muted-foreground">{localizeKnownText(run.result_summary)}</div></div>) : <div className="rounded-xl border border-dashed border-card-border bg-muted/20 px-3 py-3 text-xs text-muted-foreground">{tx('states.noApplyHistory')}</div>}</div></div></div></WorkspacePanel>
					</div>
				</section>
				<RuntimePolicyDialog open={settingsOpen} onOpenChange={setSettingsOpen} policy={runtimePolicy} baseURL={settingsBaseUrlValue} useLocalDefault={currentUseLocalDefault} apiKeySelection={defaultAPIKeySelection} apiKeyOptions={apiKeyOptions} model={selectedModelValue} modelOptions={modelOptions} onSave={handleSaveRuntimePolicy} isSaving={updateRuntimePolicy.isPending || setSetting.isPending} />
			</PageWrapper>
		</div>
	);
}
