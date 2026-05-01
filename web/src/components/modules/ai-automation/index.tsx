'use client';

import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react';
import { useTranslations } from 'next-intl';
import { BadgeCheck, Bot, Clock3, Loader2, Play, RefreshCw, Save, Split, SquarePen } from 'lucide-react';

import {
  useAIAutomationConfig,
  useAIProfile,
  useAIProfiles,
  useAIPromptTemplates,
  useActivateAIProfile,
  useAITask,
  useAITaskArtifacts,
  useAITasks,
  useCancelAITask,
  useCreateAIPromptTemplate,
  useCreateAITask,
  useDynamicRouteLearning,
  useFetchAIModels,
  useResetDynamicRouteLearning,
  useRetryAITask,
  useUpdateAIAutomationConfig,
  type AIProfile,
  type AITask,
  type AIAutomationTaskConfigSnapshot,
  type DynamicRouteLearningState,
} from '@/api/endpoints/ai-automation';
import { useSetSetting } from '@/api/endpoints/setting';
import { PageWrapper } from '@/components/common/PageWrapper';
import { toast } from '@/components/common/Toast';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Progress } from '@/components/ui/progress';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { formatDateTimeByLocale } from '@/lib/locale';
import { cn } from '@/lib/utils';
import { type Locale as AppLocale, useSettingStore } from '@/stores/setting';

import { formatConfigProfileLabel, resolveConfigSourceRuntime } from './config-source-logic';
import { consumeAIAutomationFocusTarget } from './focus-target';
import { LearningSummaryActionBar, LearningSummaryPanel } from './LearningSummaryPanel';
import {
  DEFAULT_LEARNING_SUMMARY_SECTION_ENTRIES,
  buildLearningSummarySections,
  buildLearningSummaryViewModel,
  formatLearningSampleTime,
} from './learning-summary';

const DYNAMIC_ROUTING_LEARNING_KEY = 'dynamic_routing_learning_enabled';
const SNAPSHOT_STORAGE_KEY = 'octopus-ai-automation-snapshots';
const MAX_SNAPSHOTS = 8;
const AI_AUTOMATION_REDACTED_SECRET = '[redacted]';

type TaskType = 'natural_language' | 'group_suggestion' | 'channel_recognition' | 'price_recognition' | 'model_classification' | 'config_health_check' | 'dynamic_routing_digest';
type SplitMode = 'single' | 'planner_executor' | 'planner_executor_reviewer' | 'router_mesh';
type DispatchMode = 'sequential' | 'parallel';
type ContextScope = 'workspace_full' | 'channels_groups' | 'pricing_models' | 'routing_learning';
type OperatorMode = 'ops_captain' | 'cartographer' | 'forensics';
type ViewMode = 'summary' | 'diff' | 'raw';
type LearningPreset = 'safe' | 'balanced' | 'aggressive';
type LaneRole = 'primary' | 'planner' | 'executor' | 'reviewer' | 'guard';
type WorkbenchTab = 'run' | 'templates' | 'profiles' | 'history' | 'learning';
type ToolKey = 'channel_inventory' | 'group_topology' | 'price_catalog' | 'model_catalog' | 'route_overrides' | 'learning_state' | 'profile_write' | 'profile_activate' | 'snapshot_guard';

type LaneOverride = {
  inheritMain: boolean;
  baseURL: string;
  apiKey: string;
  channelType: string;
  useLocalDefault: boolean;
  model: string;
};

type Snapshot = {
  id: string;
  label: string;
  created_at: string;
  payload: {
    taskType: TaskType;
    input: string;
    customPrompt: string;
    selectedPromptTemplateIDs: number[];
    contextScope: ContextScope;
    operatorMode: OperatorMode;
    splitMode: SplitMode;
    dispatchMode: DispatchMode;
    parallelism: number;
    selectedToolKeys: ToolKey[];
    model: string;
  };
};

type WorkbenchSection = {
  key: WorkbenchTab;
  titleKey: string;
};

const TASK_TYPES: Array<{ key: TaskType; label: string }> = [
  { key: 'natural_language', label: 'task.types.natural_language' },
  { key: 'group_suggestion', label: 'task.types.group_suggestion' },
  { key: 'channel_recognition', label: 'task.types.channel_recognition' },
  { key: 'price_recognition', label: 'task.types.price_recognition' },
  { key: 'model_classification', label: 'task.types.model_classification' },
  { key: 'config_health_check', label: 'task.types.config_health_check' },
  { key: 'dynamic_routing_digest', label: 'task.types.dynamic_routing_digest' },
];

const QUICK_INTENTS: Array<{ key: string; taskType: TaskType; label: string; text: string }> = [
  { key: 'group_review', taskType: 'group_suggestion', label: 'quickIntents.groupReview', text: 'quickIntents.groupReviewText' },
  { key: 'channel_scan', taskType: 'channel_recognition', label: 'quickIntents.channelScan', text: 'quickIntents.channelScanText' },
  { key: 'price_audit', taskType: 'price_recognition', label: 'quickIntents.priceAudit', text: 'quickIntents.priceAuditText' },
  { key: 'routing_digest', taskType: 'dynamic_routing_digest', label: 'quickIntents.routingDigest', text: 'quickIntents.routingDigestText' },
];

const OUTPUT_STYLES = [
  { key: 'concise', label: 'outputStyles.concise' },
  { key: 'balanced', label: 'outputStyles.balanced' },
  { key: 'operational', label: 'outputStyles.operational' },
] as const;

const RISK_PREFS = [
  { key: 'safe', label: 'riskPrefs.safe' },
  { key: 'normal', label: 'riskPrefs.normal' },
  { key: 'aggressive', label: 'riskPrefs.aggressive' },
] as const;

const VIEW_MODES = [
  { key: 'summary', label: 'viewModes.summary' },
  { key: 'diff', label: 'viewModes.diff' },
  { key: 'raw', label: 'viewModes.raw' },
] as const;

const CONTEXT_SCOPES = [
  { key: 'workspace_full', label: 'task.contextScopes.workspace_full' },
  { key: 'channels_groups', label: 'task.contextScopes.channels_groups' },
  { key: 'pricing_models', label: 'task.contextScopes.pricing_models' },
  { key: 'routing_learning', label: 'task.contextScopes.routing_learning' },
] as const;

const OPERATOR_MODES = [
  { key: 'ops_captain', label: 'task.operatorModes.ops_captain' },
  { key: 'cartographer', label: 'task.operatorModes.cartographer' },
  { key: 'forensics', label: 'task.operatorModes.forensics' },
] as const;

const SPLIT_MODES = [
  { key: 'single', label: 'task.splitModes.single' },
  { key: 'planner_executor', label: 'task.splitModes.planner_executor' },
  { key: 'planner_executor_reviewer', label: 'task.splitModes.planner_executor_reviewer' },
  { key: 'router_mesh', label: 'task.splitModes.router_mesh' },
] as const;

const TOOL_OPTIONS: Array<{ key: ToolKey; label: string }> = [
  { key: 'channel_inventory', label: 'task.tools.channel_inventory.label' },
  { key: 'group_topology', label: 'task.tools.group_topology.label' },
  { key: 'price_catalog', label: 'task.tools.price_catalog.label' },
  { key: 'model_catalog', label: 'task.tools.model_catalog.label' },
  { key: 'route_overrides', label: 'task.tools.route_overrides.label' },
  { key: 'learning_state', label: 'task.tools.learning_state.label' },
  { key: 'profile_write', label: 'task.tools.profile_write.label' },
  { key: 'profile_activate', label: 'task.tools.profile_activate.label' },
  { key: 'snapshot_guard', label: 'task.tools.snapshot_guard.label' },
];

const DEFAULT_TOOL_KEYS: ToolKey[] = ['channel_inventory', 'group_topology', 'model_catalog', 'learning_state', 'profile_write', 'snapshot_guard'];

const workbenchSections: WorkbenchSection[] = [
  { key: 'run', titleKey: 'task.title' },
  { key: 'templates', titleKey: 'task.promptTemplatesTitle' },
  { key: 'profiles', titleKey: 'profiles.title' },
  { key: 'history', titleKey: 'taskHistory.historyTitle' },
  { key: 'learning', titleKey: 'learning.title' },
];

const TASK_STATUS_LABELS: Record<AppLocale, Record<string, string>> = {
  'zh-Hans': { idle: '空闲', pending: '待处理', running: '运行中', recoverable: '可恢复', succeeded: '成功', failed: '失败', failed_unrecoverable: '不可恢复失败', canceled: '已取消' },
  'zh-Hant': { idle: '空閒', pending: '待處理', running: '執行中', recoverable: '可恢復', succeeded: '成功', failed: '失敗', failed_unrecoverable: '不可恢復失敗', canceled: '已取消' },
  en: { idle: 'Idle', pending: 'Pending', running: 'Running', recoverable: 'Recoverable', succeeded: 'Succeeded', failed: 'Failed', failed_unrecoverable: 'Unrecoverable failure', canceled: 'Canceled' },
  ja: { idle: '待機', pending: '保留', running: '実行中', recoverable: '復旧可能', succeeded: '成功', failed: '失敗', failed_unrecoverable: '回復不能失敗', canceled: 'キャンセル' },
};

function canUseWindow() {
  return typeof window !== 'undefined';
}

function createLaneOverrides(): Record<Exclude<LaneRole, 'primary'>, LaneOverride> {
  const base: LaneOverride = { inheritMain: true, baseURL: '', apiKey: '', channelType: 'openai-compatible', useLocalDefault: true, model: '' };
  return { planner: { ...base }, executor: { ...base }, reviewer: { ...base }, guard: { ...base } };
}

function loadSnapshots(): Snapshot[] {
  if (!canUseWindow()) return [];
  try {
    const raw = window.localStorage.getItem(SNAPSHOT_STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw) as Snapshot[];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function saveSnapshots(next: Snapshot[]) {
  if (!canUseWindow()) return;
  window.localStorage.setItem(SNAPSHOT_STORAGE_KEY, JSON.stringify(next.slice(0, MAX_SNAPSHOTS)));
}

function makeSnapshot(label: string, payload: Snapshot['payload']): Snapshot {
  return {
    id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    label,
    created_at: new Date().toISOString(),
    payload: {
      ...payload,
      selectedPromptTemplateIDs: [...payload.selectedPromptTemplateIDs],
      selectedToolKeys: [...payload.selectedToolKeys],
    },
  };
}

function isRedactedAIAutomationSecret(value: string) {
	return value.trim() === AI_AUTOMATION_REDACTED_SECRET;
}

function apiKeyForMutation(value: string) {
	const trimmed = value.trim();
	if (trimmed === AI_AUTOMATION_REDACTED_SECRET) return undefined;
	return trimmed;
}

function apiKeyForRequest(value: string) {
	const trimmed = value.trim();
	if (trimmed === '' || trimmed === AI_AUTOMATION_REDACTED_SECRET) return undefined;
	return trimmed;
}

function laneRolesFromSplitMode(mode: SplitMode): LaneRole[] {
  if (mode === 'single') return ['primary'];
  if (mode === 'planner_executor') return ['planner', 'executor'];
  if (mode === 'planner_executor_reviewer') return ['planner', 'executor', 'reviewer'];
  return ['planner', 'executor', 'reviewer', 'guard'];
}

function toneClass(status?: string) {
  switch (status) {
    case 'succeeded':
      return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300';
    case 'failed':
    case 'failed_unrecoverable':
      return 'border-rose-500/30 bg-rose-500/10 text-rose-700 dark:text-rose-300';
    case 'running':
      return 'border-sky-500/30 bg-sky-500/10 text-sky-700 dark:text-sky-300';
    default:
      return 'border-border/60 bg-muted/30 text-muted-foreground';
  }
}

function compactObjectKeys(value: unknown) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return '-';
  const keys = Object.keys(value as Record<string, unknown>);
  return keys.length > 0 ? keys.slice(0, 5).join(', ') : '-';
}

function joinPrompt(lines: string[]) {
  return lines.filter((item) => item.trim().length > 0).join('\n\n');
}

function taskTypeLabel(t: ReturnType<typeof useTranslations>, taskType: TaskType) {
  return t(TASK_TYPES.find((item) => item.key === taskType)?.label ?? 'task.types.natural_language');
}

function taskStatusLabel(locale: AppLocale, status?: string) {
  return TASK_STATUS_LABELS[locale][status ?? 'idle'] ?? status ?? '-';
}

function laneLabelKey(lane: LaneRole) {
  return `task.laneLabels.${lane}` as const;
}

function lanePromptKey(lane: LaneRole) {
  return `task.lanePrompts.${lane}` as const;
}

function EmptyLine({ testId, text }: { testId?: string; text: string }) {
  return <div data-testid={testId} className="rounded-2xl border border-dashed border-border/70 bg-muted/20 px-4 py-5 text-sm text-muted-foreground">{text}</div>;
}

function MiniStat({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="rounded-2xl border border-border/70 bg-background/75 p-3">
      <div className="text-[11px] text-muted-foreground">{label}</div>
      <div className="mt-1 break-all text-sm font-medium text-foreground">{value}</div>
    </div>
  );
}

function Chip({ active, label, onClick, testId }: { active: boolean; label: string; onClick: () => void; testId?: string }) {
  return (
    <button
      type="button"
      data-testid={testId}
      aria-pressed={active}
      onClick={onClick}
      className={cn('rounded-2xl border px-3 py-2 text-left text-sm transition', active ? 'border-primary bg-primary text-primary-foreground' : 'border-border/70 bg-background/70 hover:bg-muted/60')}
    >
      {label}
    </button>
  );
}

export function AIAutomation() {
  const t = useTranslations('aiAutomation');
  const tSetting = useTranslations('setting');
  const locale = useSettingStore((state) => state.locale);

  const configQuery = useAIAutomationConfig();
  const updateConfig = useUpdateAIAutomationConfig();
  const fetchModels = useFetchAIModels();
  const promptTemplatesQuery = useAIPromptTemplates();
  const createPromptTemplate = useCreateAIPromptTemplate();
  const createTask = useCreateAITask();
  const retryTask = useRetryAITask();
  const cancelTask = useCancelAITask();
  const profilesQuery = useAIProfiles();
  const activateProfile = useActivateAIProfile();
  const learning = useDynamicRouteLearning();
  const resetLearning = useResetDynamicRouteLearning();
  const setSetting = useSetSetting();

  const [tab, setTab] = useState<WorkbenchTab>('run');
  const [baseURL, setBaseURL] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [channelType, setChannelType] = useState('openai-compatible');
  const [useLocalDefault, setUseLocalDefault] = useState(true);
  const [modelName, setModelName] = useState('');
  const [taskType, setTaskType] = useState<TaskType>('natural_language');
  const [naturalInput, setNaturalInput] = useState('');
  const [customPrompt, setCustomPrompt] = useState('');
  const [selectedPromptTemplateIDs, setSelectedPromptTemplateIDs] = useState<number[]>([]);
  const [contextScope, setContextScope] = useState<ContextScope>('workspace_full');
  const [operatorMode, setOperatorMode] = useState<OperatorMode>('ops_captain');
  const [splitMode, setSplitMode] = useState<SplitMode>('planner_executor_reviewer');
  const [dispatchMode, setDispatchMode] = useState<DispatchMode>('parallel');
  const [parallelism, setParallelism] = useState(3);
  const [outputStyle, setOutputStyle] = useState<(typeof OUTPUT_STYLES)[number]['key']>('balanced');
  const [riskPreference, setRiskPreference] = useState<(typeof RISK_PREFS)[number]['key']>('normal');
  const [viewMode, setViewMode] = useState<ViewMode>('summary');
  const [selectedQuickIntent, setSelectedQuickIntent] = useState(QUICK_INTENTS[0].key);
  const [selectedToolKeys, setSelectedToolKeys] = useState<ToolKey[]>(DEFAULT_TOOL_KEYS);
  const [autoPickRecommendedModel, setAutoPickRecommendedModel] = useState(true);
  const [autoSnapshotEnabled, setAutoSnapshotEnabled] = useState(true);
  const [learningPreset, setLearningPreset] = useState<LearningPreset>('balanced');
  const [learningEnabledDraft, setLearningEnabledDraft] = useState<boolean | null>(null);
  const [snapshotLabel, setSnapshotLabel] = useState('');
  const [selectedProfileID, setSelectedProfileID] = useState<number | undefined>();
  const [selectedTaskID, setSelectedTaskID] = useState<number | undefined>();
  const [selectedHistoryTaskID, setSelectedHistoryTaskID] = useState<number | undefined>();
  const [taskHistoryPage, setTaskHistoryPage] = useState(1);
  const [taskHistoryStatus, setTaskHistoryStatus] = useState<'all' | 'pending' | 'running' | 'recoverable' | 'succeeded' | 'failed' | 'failed_unrecoverable' | 'canceled'>('all');
  const [taskHistoryType, setTaskHistoryType] = useState<'all' | TaskType>('all');
  const [taskHistoryKeyword, setTaskHistoryKeyword] = useState('');
  const [newTemplateName, setNewTemplateName] = useState('');
  const [newTemplatePrompt, setNewTemplatePrompt] = useState('');
  const [newTemplateRequirement, setNewTemplateRequirement] = useState('');
  const [snapshots, setSnapshots] = useState<Snapshot[]>(() => loadSnapshots());
  const [laneOverrides, setLaneOverrides] = useState(() => createLaneOverrides());
  const [laneLaunches, setLaneLaunches] = useState<Array<{ lane: LaneRole; taskId: number; model: string }>>([]);

  const initializedRef = useRef(false);
  const learningSectionRef = useRef<HTMLDivElement | null>(null);

  const config = configQuery.data;
  const profiles = profilesQuery.data ?? [];
  const promptTemplates = promptTemplatesQuery.data ?? [];
  const modelCandidates = fetchModels.data?.candidates ?? [];
  const learningStates = learning.data?.states ?? [];
  const learningEnabledRuntime = learning.data?.enabled;
  const learningEnabledSetting = config?.dynamic_routing_learning_enabled;
  const learningEnabled = learningEnabledDraft ?? learningEnabledRuntime ?? learningEnabledSetting ?? false;

  const taskQuery = useAITask(selectedTaskID);
  const historyTaskQuery = useAITask(selectedHistoryTaskID);
  const selectedProfileQuery = useAIProfile(selectedProfileID ?? profiles[0]?.id);
  const activeTask = historyTaskQuery.data ?? taskQuery.data ?? createTask.data ?? null;
  const artifacts = useAITaskArtifacts(activeTask?.id);
  const selectedProfile = selectedProfileQuery.data;

  const selectedTemplateIDs = selectedPromptTemplateIDs.length > 0
    ? selectedPromptTemplateIDs
    : promptTemplates.filter((item) => item.enabled).slice(0, 2).map((item) => item.id);

  const selectedTemplates = useMemo(
    () => promptTemplates.filter((item) => selectedTemplateIDs.includes(item.id)),
    [promptTemplates, selectedTemplateIDs],
  );

  const {
    configSourceMode,
    requestedActiveProfile,
    activeProfile,
    sourceFallbackReason,
    runtimeFallbackActive,
    manualDraftBaseURL,
    manualDraftAPIKey,
    manualDraftChannelType,
    manualDraftUseLocalDefault,
    manualDraftModel,
    effectiveBaseURL,
    effectiveAPIKey,
    effectiveChannelType,
    effectiveUseLocalDefault,
    effectiveModel,
  } = resolveConfigSourceRuntime(config, { baseURL, apiKey, channelType, modelName, useLocalDefault });

  const recommendedModel = modelCandidates.find((item) => item.recommended)
    ?? modelCandidates.find((item) => item.available && item.free_likely)
    ?? modelCandidates[0];

  const effectiveTaskModel = autoPickRecommendedModel ? (recommendedModel?.name || effectiveModel) : effectiveModel;

  const taskHistory = useAITasks({
    page: taskHistoryPage,
    page_size: 8,
    status: taskHistoryStatus === 'all' ? undefined : taskHistoryStatus,
    type: taskHistoryType === 'all' ? undefined : taskHistoryType,
    keyword: taskHistoryKeyword.trim() || undefined,
  });

  const taskHistoryItems = taskHistory.data?.items ?? [];
  const taskHistoryTotal = taskHistory.data?.total ?? 0;
  const taskHistoryPageSize = taskHistory.data?.page_size ?? 8;
  const taskHistoryPageCount = Math.max(1, Math.ceil(taskHistoryTotal / taskHistoryPageSize));

  const parsedResult = useMemo(() => {
    const raw = artifacts.data?.result_json;
    if (!raw) return null;
    try {
      const parsed = JSON.parse(raw);
      return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed as Record<string, unknown> : null;
    } catch {
      return null;
    }
  }, [artifacts.data?.result_json]);

  const learningSummaryView = buildLearningSummaryViewModel({
    states: learningStates,
    enabled: learningEnabled,
    locale,
    emptyLabel: t('learning.summary.notAvailable'),
    topTargetFormatter: (state: DynamicRouteLearningState) => t('learning.topTargetValue', { model: state.model_name, channel: state.channel_id, key: state.channel_key_id }),
    itemLabels: {
      statusLabel: t('learning.summary.statusLabel'),
      enabledLabel: t('learning.enabled'),
      disabledLabel: t('learning.disabled'),
      samplesLabel: t('learning.summary.samplesLabel'),
      runtimeLabel: t('learning.summary.runtimeLabel'),
      runtimeEnabledLabel: t('learning.summary.runtimeEnabled'),
      runtimeDisabledLabel: t('learning.summary.runtimeDisabled'),
      latestSampleLabel: t('learning.summary.latestSampleLabel'),
      topTargetLabel: t('learning.summary.topTargetLabel'),
    },
    noticeValues: {
      enabledWithSamples: t('learning.summary.noticeEnabledReady', { count: learningStates.length }),
      enabledWithoutSamples: t('learning.summary.noticeEnabledEmpty'),
      disabledWithSamples: t('learning.summary.noticeDisabledWithSamples', { count: learningStates.length }),
      disabledWithoutSamples: t('learning.summary.noticeDisabledEmpty'),
    },
  });

  const { primaryItems: learningPrimaryItems, secondaryItems: learningSecondaryItems } = buildLearningSummarySections({
    items: learningSummaryView.items,
    primaryEntries: DEFAULT_LEARNING_SUMMARY_SECTION_ENTRIES.primaryEntries,
    secondaryEntries: DEFAULT_LEARNING_SUMMARY_SECTION_ENTRIES.secondaryEntries,
    errorScope: 'ai automation learning',
  });

  const learningDisplay = learningSummaryView.display;

  const bindWorkbenchSection = (sectionKey: (typeof workbenchSections)[number]['key']) => (node: HTMLDivElement | null) => {
    if (sectionKey === 'learning') {
      learningSectionRef.current = node;
    }
  };

  useEffect(() => {
    if (initializedRef.current || !config) return;
    initializedRef.current = true;
    setBaseURL(config.manual_config?.base_url || config.base_url || '');
    setApiKey(config.manual_config?.api_key || config.api_key || '');
    setChannelType(config.manual_config?.channel_type || config.channel_type || 'openai-compatible');
    setUseLocalDefault(config.manual_config?.use_local_default ?? config.use_local_default ?? true);
    setModelName(config.manual_config?.model || config.model || '');
    setSelectedPromptTemplateIDs(promptTemplates.filter((item) => item.enabled).slice(0, 2).map((item) => item.id));
    setSelectedProfileID(config.active_ai_profile_id || profiles[0]?.id);
  }, [config, profiles, promptTemplates]);

  useEffect(() => {
    if (selectedPromptTemplateIDs.length > 0) return;
    const enabled = promptTemplates.filter((item) => item.enabled).slice(0, 2).map((item) => item.id);
    if (enabled.length > 0) setSelectedPromptTemplateIDs(enabled);
  }, [promptTemplates, selectedPromptTemplateIDs.length]);

  useEffect(() => {
    setTaskHistoryPage(1);
  }, [taskHistoryKeyword, taskHistoryStatus, taskHistoryType]);

  useEffect(() => {
    saveSnapshots(snapshots);
  }, [snapshots]);

  useEffect(() => {
    const targetKey = consumeAIAutomationFocusTarget();
    if (targetKey !== 'learning') return;
    setTab('learning');
    const target = learningSectionRef.current;
    if (!target) return;
    target.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }, []);

  const snapshotCount = snapshots.length;
  const latestSnapshot = snapshots[0];
  const activeLanes = laneRolesFromSplitMode(splitMode);
  const dispatchedLanes = dispatchMode === 'parallel' ? activeLanes.slice(0, parallelism) : activeLanes;
  const currentStatus = activeTask?.status || 'idle';
  const currentResultJSON = artifacts.data?.result_json || '';
  const taskViewText = viewMode === 'summary'
    ? (activeTask?.result_summary || activeTask?.error_message || t('task.snapshotEmpty'))
    : viewMode === 'diff'
      ? compactObjectKeys(parsedResult)
      : currentResultJSON || t('task.snapshotEmpty');

  const currentTaskPreview = activeTask?.status === 'failed' || activeTask?.status === 'failed_unrecoverable'
    ? activeTask?.error_message || t('task.failureTitle')
    : activeTask?.result_summary || t('task.snapshotEmpty');

  const currentLaneModel = (lane: LaneRole) => {
    if (lane === 'primary') return effectiveTaskModel;
    const override = laneOverrides[lane];
    return override.inheritMain ? effectiveTaskModel : (override.model || effectiveTaskModel);
  };

  const currentLaneConfig = (lane: LaneRole): AIAutomationTaskConfigSnapshot => {
    if (lane === 'primary') {
      return {
        base_url: effectiveBaseURL,
		api_key: apiKeyForRequest(effectiveAPIKey),
        channel_type: effectiveChannelType,
        model: effectiveTaskModel || undefined,
        use_local_default: effectiveUseLocalDefault,
        tool_keys: selectedToolKeys,
      };
    }

    const override = laneOverrides[lane];
    if (override.inheritMain) {
      return {
        base_url: effectiveBaseURL,
		api_key: apiKeyForRequest(effectiveAPIKey),
        channel_type: effectiveChannelType,
        model: currentLaneModel(lane) || undefined,
        use_local_default: effectiveUseLocalDefault,
        tool_keys: selectedToolKeys,
      };
    }

    return {
      base_url: override.baseURL || undefined,
		api_key: apiKeyForRequest(override.apiKey),
      channel_type: override.channelType,
      model: override.model || undefined,
      use_local_default: override.useLocalDefault,
      tool_keys: selectedToolKeys,
    };
  };

  const buildPrompt = (lane: LaneRole) => joinPrompt([
    `${t('task.taskType')}: ${taskTypeLabel(t, taskType)}`,
    `${t('task.contextScope')}: ${t(CONTEXT_SCOPES.find((item) => item.key === contextScope)?.label ?? 'task.contextScopes.workspace_full')}`,
    `${t('task.operatorMode')}: ${t(OPERATOR_MODES.find((item) => item.key === operatorMode)?.label ?? 'task.operatorModes.ops_captain')}`,
    `${t('task.aiSplitMode')}: ${t(SPLIT_MODES.find((item) => item.key === splitMode)?.label ?? 'task.splitModes.planner_executor_reviewer')}`,
    `${t('task.outputStyle')}: ${t(OUTPUT_STYLES.find((item) => item.key === outputStyle)?.label ?? 'outputStyles.balanced')}`,
    `${t('task.riskPreference')}: ${t(RISK_PREFS.find((item) => item.key === riskPreference)?.label ?? 'riskPrefs.normal')}`,
    `${t('task.viewMode')}: ${t(VIEW_MODES.find((item) => item.key === viewMode)?.label ?? 'viewModes.summary')}`,
    `${t('task.resultModelLabel')} ${currentLaneModel(lane) || '-'}`,
    `${t('task.toolCapabilityTitle')}: ${selectedToolKeys.map((key) => t(TOOL_OPTIONS.find((item) => item.key === key)?.label ?? 'task.tools.channel_inventory.label')).join('、')}`,
    lane === 'primary' ? '' : `${t(laneLabelKey(lane))}: ${t(lanePromptKey(lane))}`,
    customPrompt.trim(),
  ]);

  const pushSnapshot = (label: string) => {
    const next = [
      makeSnapshot(label, {
        taskType,
        input: naturalInput,
        customPrompt,
        selectedPromptTemplateIDs: selectedTemplateIDs,
        contextScope,
        operatorMode,
        splitMode,
        dispatchMode,
        parallelism,
        selectedToolKeys,
        model: effectiveTaskModel,
      }),
      ...snapshots,
    ].slice(0, MAX_SNAPSHOTS);
    setSnapshots(next);
  };

  const saveConfig = async () => {
    try {
      await updateConfig.mutateAsync({
        base_url: manualDraftBaseURL,
		api_key: apiKeyForMutation(apiKey),
        channel_type: manualDraftChannelType,
        model: manualDraftModel,
        use_local_default: manualDraftUseLocalDefault,
      });
      toast.success(t('toast.configSaved'));
    } catch (error) {
      toast.error(t('toast.configFailed'), { description: error instanceof Error ? error.message : undefined });
    }
  };

  const toggleLearning = (enabled: boolean) => {
    setLearningEnabledDraft(enabled);
    setSetting.mutate(
      { key: DYNAMIC_ROUTING_LEARNING_KEY, value: enabled ? 'true' : 'false' },
      {
        onSuccess: () => {
          setLearningEnabledDraft(null);
          toast.success(t('toast.learningSaved'));
          learning.refetch();
          configQuery.refetch();
        },
        onError: (error) => {
          setLearningEnabledDraft(null);
          toast.error(t('toast.learningSaveFailed'), { description: error instanceof Error ? error.message : undefined });
        },
      },
    );
  };

  const handleResetLearning = () => {
    resetLearning.mutate(undefined, {
      onSuccess: () => toast.success(t('learning.resetSuccess')),
      onError: (error) => toast.error(t('learning.resetFailed'), { description: error instanceof Error ? error.message : undefined }),
    });
  };

  const applyQuickIntent = (key: string) => {
    const intent = QUICK_INTENTS.find((item) => item.key === key);
    if (!intent) return;
    setSelectedQuickIntent(intent.key);
    setTaskType(intent.taskType);
    setNaturalInput((current) => (current.trim() ? `${current.trim()}\n\n${t(intent.text)}` : t(intent.text)));
  };

  const createTaskForLane = async (lane: LaneRole) => {
    if (!naturalInput.trim()) {
      toast.warning(t('toast.inputRequired'));
      return null;
    }

    if (autoSnapshotEnabled) {
      pushSnapshot(`${t('task.snapshotBeforeLaunch')} · ${taskTypeLabel(t, taskType)}`);
    }

    try {
      return await createTask.mutateAsync({
        type: taskType,
        input_text: naturalInput,
        context_scope: contextScope,
        prompt_template_ids: selectedTemplateIDs,
        custom_prompt: buildPrompt(lane),
        config_snapshot: currentLaneConfig(lane),
      });
    } catch (error) {
      toast.error(t('toast.taskFailed'), { description: error instanceof Error ? error.message : undefined });
      return null;
    }
  };

  const runSingle = async () => {
    const task = await createTaskForLane('primary');
    if (task) {
      setSelectedTaskID(task.id);
      setSelectedHistoryTaskID(undefined);
      toast.success(t('toast.taskCreated'));
    }
  };

  const runMultiLane = async () => {
    const launches: Array<AITask> = [];
    for (const lane of dispatchedLanes) {
      const task = await createTaskForLane(lane);
      if (task) launches.push(task);
    }

    if (launches.length > 0) {
      setLaneLaunches((current) => [
        ...launches.map((task, index) => ({ lane: dispatchedLanes[index], taskId: task.id, model: task.selected_model || currentLaneModel(dispatchedLanes[index]) || '-' })),
        ...current,
      ].slice(0, 8));
      setSelectedTaskID(launches[0].id);
      setSelectedHistoryTaskID(undefined);
      toast.success(t('task.multiLaneLaunched', { count: launches.length }));
    }
  };

  const createTemplate = async () => {
    if (!newTemplateName.trim() || !newTemplatePrompt.trim()) {
      toast.warning(t('toast.templateRequired'));
      return;
    }

    try {
      const created = await createPromptTemplate.mutateAsync({
        name: newTemplateName.trim(),
        task_type: taskType,
        prompt: newTemplatePrompt.trim(),
        work_requirement: newTemplateRequirement.trim() || undefined,
      });
      setSelectedPromptTemplateIDs((current) => [...new Set([...current, created.id])]);
      setNewTemplateName('');
      setNewTemplatePrompt('');
      setNewTemplateRequirement('');
      toast.success(t('toast.templateCreated'));
    } catch (error) {
      toast.error(t('toast.templateFailed'), { description: error instanceof Error ? error.message : undefined });
    }
  };

  const fetchModelList = async () => {
    try {
      const result = await fetchModels.mutateAsync({
        base_url: effectiveBaseURL,
		api_key: apiKeyForRequest(effectiveAPIKey),
        channel_type: effectiveChannelType,
        use_local_default: effectiveUseLocalDefault,
      });
      if (result.selected_name && configSourceMode !== 'ai_profile') {
        setModelName(result.selected_name);
      }
      toast.success(t('toast.modelsFetched'));
    } catch (error) {
      toast.error(t('toast.modelsFetchFailed'), { description: error instanceof Error ? error.message : undefined });
    }
  };

  const applyProfile = async (profile: AIProfile) => {
    try {
      await activateProfile.mutateAsync(profile.id);
      setSelectedProfileID(profile.id);
      toast.success(tSetting('saved'));
    } catch (error) {
      toast.error(tSetting('aiAutomationSource.activateFailed'), { description: error instanceof Error ? error.message : undefined });
    }
  };

  const toggleTool = (key: ToolKey) => {
    setSelectedToolKeys((current) => current.includes(key) ? current.filter((item) => item !== key) : [...current, key]);
  };

  const updateLaneOverride = <K extends keyof LaneOverride>(lane: Exclude<LaneRole, 'primary'>, key: K, value: LaneOverride[K]) => {
    setLaneOverrides((current) => ({ ...current, [lane]: { ...current[lane], [key]: value } }));
  };

  const restoreSnapshot = (snapshot: Snapshot | undefined) => {
    if (!snapshot) return;
    const payload = snapshot.payload;
    setTaskType(payload.taskType);
    setNaturalInput(payload.input || '');
    setCustomPrompt(payload.customPrompt || '');
    setSelectedPromptTemplateIDs(payload.selectedPromptTemplateIDs || []);
    setContextScope(payload.contextScope || 'workspace_full');
    setOperatorMode(payload.operatorMode || 'ops_captain');
    setSplitMode(payload.splitMode || 'planner_executor_reviewer');
    setDispatchMode(payload.dispatchMode || 'parallel');
    setParallelism(payload.parallelism || 3);
    setSelectedToolKeys((payload.selectedToolKeys && payload.selectedToolKeys.length > 0 ? payload.selectedToolKeys : DEFAULT_TOOL_KEYS) as ToolKey[]);
    toast.success(t('task.snapshotRestored'));
  };

  const clearSnapshots = () => {
    setSnapshots([]);
    toast.success(t('task.snapshotsCleared'));
  };

  const createManualSnapshot = () => {
    const label = snapshotLabel.trim() || `${t('task.snapshotCreate')} ${new Date().toLocaleTimeString()}`;
    pushSnapshot(label);
    setSnapshotLabel('');
    toast.success(t('task.snapshotCreated'));
  };

  const runDisabled = !config?.enabled || createTask.isPending;
  const selectedConfigSourceText = configSourceMode === 'ai_profile' ? tSetting('aiAutomationSource.aiProfile') : tSetting('aiAutomationSource.manual');
  const currentProfileLabel = formatConfigProfileLabel(activeProfile || requestedActiveProfile) || t('profiles.noActiveLabel');

  return (
    <PageWrapper className="space-y-5" data-testid="ai-automation-page">
      <section className="grid min-h-0 gap-5 xl:grid-cols-[minmax(0,1.65fr)_22rem]">
        <Card className="overflow-hidden rounded-3xl border-border/70 bg-card/95">
          <CardHeader className="space-y-3 border-b border-border/60 pb-4">
            <div className="flex flex-wrap items-center gap-2">
              <Badge variant="outline" className="rounded-full border-border/70 bg-background/70 px-3 py-1 text-[11px] uppercase tracking-[0.18em] text-muted-foreground">
                {t('hero.badge')}
              </Badge>
              <Badge className={cn('rounded-full px-3 py-1 text-xs', toneClass(currentStatus))}>{taskStatusLabel(locale, currentStatus)}</Badge>
            </div>
            <div className="space-y-1.5">
              <CardTitle className="text-2xl font-semibold tracking-tight">{t('hero.title')}</CardTitle>
              <CardDescription className="max-w-3xl text-sm leading-6 text-muted-foreground">{t('task.desc')}</CardDescription>
            </div>
          </CardHeader>

          <CardContent className="space-y-5 p-5">
            <div className="grid gap-3 md:grid-cols-4">
              <MiniStat label={t('config.activeSourceTitle')} value={selectedConfigSourceText} />
              <MiniStat label={t('task.effectiveModelTitle')} value={effectiveTaskModel || '-'} />
              <MiniStat label={t('profiles.activeTitle')} value={currentProfileLabel} />
              <MiniStat label={t('learning.summary.samplesLabel')} value={learningDisplay.sampleCount} />
            </div>

            {runtimeFallbackActive ? (
              <div className="rounded-2xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-900 dark:text-amber-100">
                <div className="font-medium">{tSetting('aiAutomationSource.runtimeFallbackTitle')}</div>
                <div className="mt-1 text-xs leading-5">{sourceFallbackReason ? tSetting(`aiAutomationSource.fallbackReasons.${sourceFallbackReason}`) : tSetting('aiAutomationSource.runtimeFallbackNotice')}</div>
              </div>
            ) : null}

            <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_18rem]">
              <div className="space-y-4">
                <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-muted-foreground">{t('task.taskType')}</label>
                    <Select value={taskType} onValueChange={(value) => setTaskType(value as TaskType)}>
                      <SelectTrigger aria-label={t('task.taskType')} className="rounded-2xl bg-background/80"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        {TASK_TYPES.map((item) => <SelectItem key={item.key} value={item.key}>{t(item.label)}</SelectItem>)}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-muted-foreground">{t('task.contextScope')}</label>
                    <Select value={contextScope} onValueChange={(value) => setContextScope(value as ContextScope)}>
                      <SelectTrigger aria-label={t('task.contextScope')} className="rounded-2xl bg-background/80"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        {CONTEXT_SCOPES.map((item) => <SelectItem key={item.key} value={item.key}>{t(item.label)}</SelectItem>)}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-muted-foreground">{t('task.operatorMode')}</label>
                    <Select value={operatorMode} onValueChange={(value) => setOperatorMode(value as OperatorMode)}>
                      <SelectTrigger aria-label={t('task.operatorMode')} className="rounded-2xl bg-background/80"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        {OPERATOR_MODES.map((item) => <SelectItem key={item.key} value={item.key}>{t(item.label)}</SelectItem>)}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-muted-foreground">{t('task.aiSplitMode')}</label>
                    <Select value={splitMode} onValueChange={(value) => setSplitMode(value as SplitMode)}>
                      <SelectTrigger aria-label={t('task.aiSplitMode')} className="rounded-2xl bg-background/80"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        {SPLIT_MODES.map((item) => <SelectItem key={item.key} value={item.key}>{t(item.label)}</SelectItem>)}
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                <div className="rounded-3xl border border-border/70 bg-background/70 p-4">
                  <div className="space-y-2">
                    <div className="text-sm font-medium text-foreground">{t('task.currentTaskTypeTitle')}</div>
                    <textarea
                      className="min-h-40 w-full resize-none rounded-2xl border border-border/70 bg-background px-4 py-3 text-sm outline-none transition focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30"
                      placeholder={t('task.placeholder')}
                      value={naturalInput}
                      onChange={(event) => setNaturalInput(event.target.value)}
                    />
                    <textarea
                      className="min-h-24 w-full resize-none rounded-2xl border border-border/60 bg-background/80 px-4 py-3 text-sm outline-none transition focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30"
                      placeholder={t('task.customPromptPlaceholder')}
                      value={customPrompt}
                      onChange={(event) => setCustomPrompt(event.target.value)}
                    />
                  </div>
                </div>

                <div className="flex flex-wrap items-center gap-3 rounded-3xl border border-border/70 bg-background/70 p-4">
                  <Button type="button" className="rounded-2xl" onClick={runSingle} disabled={runDisabled}>
                    {createTask.isPending ? <Loader2 className="size-4 animate-spin" /> : <Play className="size-4" />}
                    {t('task.run')}
                  </Button>
                  <Button type="button" variant="outline" className="rounded-2xl" onClick={runMultiLane} disabled={runDisabled}>
                    <Split className="size-4" />
                    {t('task.runMultiLane')}
                  </Button>
                  <Button type="button" variant="outline" className="rounded-2xl" onClick={fetchModelList} disabled={!config?.enabled || fetchModels.isPending}>
                    {fetchModels.isPending ? <Loader2 className="size-4 animate-spin" /> : <RefreshCw className="size-4" />}
                    {t('config.fetchModels')}
                  </Button>
                  <Button type="button" variant="outline" className="rounded-2xl" onClick={saveConfig}>
                    <Save className="size-4" />
                    {t('config.save')}
                  </Button>
                </div>

                <Card className="rounded-3xl border border-border/70 bg-background/75">
                  <CardHeader className="pb-3">
                    <CardTitle className="text-base">{t('task.quickIntentsTitle')}</CardTitle>
                    <CardDescription>{t('task.quickIntentsDesc')}</CardDescription>
                  </CardHeader>
                  <CardContent className="flex flex-wrap gap-2">
                    {QUICK_INTENTS.map((intent) => (
                      <Chip key={intent.key} active={selectedQuickIntent === intent.key} label={t(intent.label)} onClick={() => applyQuickIntent(intent.key)} />
                    ))}
                  </CardContent>
                </Card>
              </div>

              <div className="space-y-4">
                <Card className="rounded-3xl border border-border/70 bg-background/75">
                  <CardHeader className="pb-3">
                    <CardTitle className="text-base">{t('config.title')}</CardTitle>
                    <CardDescription>{t('config.desc')}</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    <Input aria-label={t('config.baseUrl')} value={baseURL} onChange={(event) => setBaseURL(event.target.value)} placeholder={t('config.baseUrl')} className="rounded-2xl" />
					<Input aria-label={t('config.apiKey')} value={apiKey} onChange={(event) => setApiKey(event.target.value)} placeholder={t('config.apiKeyPlaceholder')} className="rounded-2xl" />
                    <Select value={channelType} onValueChange={setChannelType}>
                      <SelectTrigger aria-label={t('config.channelType')} className="rounded-2xl"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="openai-compatible">{t('config.channelTypes.openaiCompatible')}</SelectItem>
                        <SelectItem value="openai">{t('config.channelTypes.openai')}</SelectItem>
                        <SelectItem value="anthropic">{t('config.channelTypes.anthropic')}</SelectItem>
                        <SelectItem value="gemini">{t('config.channelTypes.gemini')}</SelectItem>
                      </SelectContent>
                    </Select>
                    <Input aria-label={t('config.model')} value={modelName} onChange={(event) => setModelName(event.target.value)} placeholder={t('config.modelPlaceholder')} className="rounded-2xl" />
                    <div className="flex items-center justify-between rounded-2xl border border-border/70 bg-muted/25 px-3 py-2">
                      <div>
                        <div className="text-sm font-medium text-foreground">{t('config.localDefaultTitle')}</div>
                        <div className="text-xs text-muted-foreground">{t('config.localDefaultDesc')}</div>
                      </div>
                      <Switch aria-label={t('config.localDefaultTitle')} checked={useLocalDefault} onCheckedChange={setUseLocalDefault} />
                    </div>
                  </CardContent>
                </Card>

                <Card className="rounded-3xl border border-border/70 bg-background/75">
                  <CardHeader>
                    <CardTitle className="text-base">{t('models.title')}</CardTitle>
                    <CardDescription>{t('models.desc')}</CardDescription>
                  </CardHeader>
                  <CardContent className="max-h-[26rem] space-y-3 overflow-y-auto pr-1">
                    {modelCandidates.length === 0 ? <EmptyLine text={t('models.empty')} /> : modelCandidates.map((item) => (
                      <div key={`${item.name}-${item.channel_name || item.source}`} className="rounded-2xl border border-border/70 bg-background/80 p-3">
                        <div className="flex flex-wrap items-center gap-2">
                          <div className="min-w-0 flex-1 text-sm font-medium text-foreground">{item.name}</div>
                          {item.recommended ? <Badge variant="outline">{t('models.badges.recommended')}</Badge> : null}
                          {item.free_likely ? <Badge variant="outline">{t('models.badges.free')}</Badge> : null}
                        </div>
                        <div className="mt-2 grid gap-2 sm:grid-cols-2">
                          <MiniStat label={t('models.success')} value={`${Math.round((item.success_rate || 0) * 100)}%`} />
                          <MiniStat label={t('models.latency')} value={`${item.avg_latency_ms ?? 0} ms`} />
                        </div>
                      </div>
                    ))}
                  </CardContent>
                </Card>
              </div>
            </div>

            <Card className="rounded-3xl border border-border/70 bg-background/75">
              <CardHeader className="pb-3">
                <CardTitle className="text-base">{t('task.resultPreviewTitle')}</CardTitle>
                <CardDescription>{currentTaskPreview}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid gap-3 sm:grid-cols-4">
                  <MiniStat label={t('task.progressTitle')} value={`${activeTask?.progress ?? 0}%`} />
                  <MiniStat label={t('task.executionToneLabel')} value={t(OUTPUT_STYLES.find((item) => item.key === outputStyle)?.label ?? 'outputStyles.balanced')} />
                  <MiniStat label={t('task.executionRiskLabel')} value={t(RISK_PREFS.find((item) => item.key === riskPreference)?.label ?? 'riskPrefs.normal')} />
                  <MiniStat label={t('task.executionViewLabel')} value={t(VIEW_MODES.find((item) => item.key === viewMode)?.label ?? 'viewModes.summary')} />
                </div>
                <Progress value={activeTask?.progress ?? 0} />
                <div className="flex flex-wrap gap-2">
                  {VIEW_MODES.map((item) => <Chip key={item.key} active={viewMode === item.key} label={t(item.label)} onClick={() => setViewMode(item.key)} />)}
                </div>
                <div className="min-h-36 rounded-2xl border border-border/70 bg-muted/20 p-4 text-sm leading-6 text-foreground whitespace-pre-wrap" data-testid={`result-view-${viewMode}`}>
                  {taskViewText}
                </div>
              </CardContent>
            </Card>
          </CardContent>
        </Card>

        <aside className="space-y-5">
          <Card className="rounded-3xl border-border/70 bg-card/95">
            <CardHeader>
              <CardTitle className="text-base">{t('hero.learning')}</CardTitle>
              <CardDescription>{learningDisplay.runtimeKey === 'runtimeEnabled' ? t('learning.summary.runtimeEnabled') : t('learning.summary.runtimeDisabled')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <MiniStat label={t('task.effectiveModelTitle')} value={effectiveTaskModel || '-'} />
              <MiniStat label={t('config.activeSourceTitle')} value={selectedConfigSourceText} />
              <MiniStat label={t('profiles.activeTitle')} value={currentProfileLabel} />
              <MiniStat label={t('task.snapshotActive')} value={snapshotCount} />
              {latestSnapshot ? <MiniStat label={t('task.latestSnapshotTitle')} value={latestSnapshot.label} /> : null}
            </CardContent>
          </Card>
        </aside>
      </section>

      <section className="rounded-3xl border border-border/70 bg-card/95 p-5">
        <div className="flex flex-wrap items-center gap-2 pb-4">
          {workbenchSections.map((section) => <Chip key={section.key} active={tab === section.key} label={t(section.titleKey)} onClick={() => setTab(section.key)} />)}
        </div>

        {tab === 'run' ? (
          <div className="grid gap-5 xl:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)]">
            <Card className="rounded-3xl border border-border/70 bg-background/75">
              <CardHeader>
                <CardTitle className="text-base">{t('task.toolCapabilityTitle')}</CardTitle>
                <CardDescription>{t('task.toolCapabilityDesc')}</CardDescription>
              </CardHeader>
              <CardContent className="grid gap-3 md:grid-cols-2">
                {TOOL_OPTIONS.map((tool) => (
                  <button key={tool.key} type="button" className={cn('rounded-2xl border border-border/70 bg-background/80 p-3 text-left transition hover:bg-muted/40', selectedToolKeys.includes(tool.key) && 'border-primary/60')} onClick={() => toggleTool(tool.key)}>
                    <div className="flex items-center justify-between gap-2">
                      <div className="text-sm font-medium text-foreground">{t(tool.label)}</div>
                      <Badge variant="outline">{selectedToolKeys.includes(tool.key) ? t('task.toolActive') : t('task.toolInactive')}</Badge>
                    </div>
                  </button>
                ))}
              </CardContent>
            </Card>

            <Card className="rounded-3xl border border-border/70 bg-background/75">
              <CardHeader>
                <CardTitle className="text-base">{t('task.snapshotPanelTitle')}</CardTitle>
                <CardDescription>{t('task.snapshotPanelDesc')}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <Input value={snapshotLabel} onChange={(event) => setSnapshotLabel(event.target.value)} placeholder={t('task.snapshotLabelPlaceholder')} className="rounded-2xl" />
                <div className="flex flex-wrap gap-2">
                  <Button type="button" className="rounded-2xl" onClick={createManualSnapshot}>{t('task.snapshotCreate')}</Button>
                  <Button type="button" variant="outline" className="rounded-2xl" onClick={() => restoreSnapshot(latestSnapshot)} disabled={!latestSnapshot}>{t('task.snapshotRestore')}</Button>
                  <Button type="button" variant="outline" className="rounded-2xl" onClick={clearSnapshots} disabled={snapshots.length === 0}>{t('task.snapshotClear')}</Button>
                </div>
                <div className="max-h-[14rem] space-y-3 overflow-y-auto pr-1">
                  {snapshots.length === 0 ? <EmptyLine text={t('task.snapshotEmpty')} /> : snapshots.map((snapshot) => (
                    <button key={snapshot.id} type="button" className="w-full rounded-2xl border border-border/70 bg-background/80 p-3 text-left transition hover:bg-muted/40" onClick={() => restoreSnapshot(snapshot)}>
                      <div className="flex items-center justify-between gap-3">
                        <div className="min-w-0 text-sm font-medium text-foreground">{snapshot.label}</div>
                        <Clock3 className="size-4 text-muted-foreground" />
                      </div>
                      <div className="mt-1 text-xs text-muted-foreground">{formatDateTimeByLocale(snapshot.created_at, locale)}</div>
                    </button>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        ) : null}

        {tab === 'templates' ? (
          <Card className="rounded-3xl border border-border/70 bg-background/75">
            <CardHeader>
              <CardTitle className="text-base">{t('task.promptTemplatesTitle')}</CardTitle>
              <CardDescription>{selectedTemplates.length > 0 ? selectedTemplates.map((item) => item.name).join(' · ') : t('task.snapshotEmpty')}</CardDescription>
            </CardHeader>
            <CardContent className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
              <div className="max-h-[26rem] space-y-3 overflow-y-auto pr-1">
                {promptTemplates.length === 0 ? <EmptyLine text={t('task.snapshotEmpty')} /> : promptTemplates.map((item) => (
                  <label key={item.id} className="flex items-start gap-3 rounded-2xl border border-border/70 bg-background/80 p-3">
                    <input type="checkbox" checked={selectedTemplateIDs.includes(item.id)} onChange={() => setSelectedPromptTemplateIDs((current) => current.includes(item.id) ? current.filter((value) => value !== item.id) : [...current, item.id])} />
                    <div className="min-w-0">
                      <div className="text-sm font-medium text-foreground">{item.name}</div>
                      <div className="text-xs leading-5 text-muted-foreground">{item.prompt}</div>
                    </div>
                  </label>
                ))}
              </div>
              <div className="space-y-3">
                <Input value={newTemplateName} onChange={(event) => setNewTemplateName(event.target.value)} placeholder={t('task.templateNamePlaceholder')} className="rounded-2xl" />
                <textarea className="min-h-28 w-full resize-none rounded-2xl border border-border/70 bg-background px-4 py-3 text-sm outline-none" value={newTemplatePrompt} onChange={(event) => setNewTemplatePrompt(event.target.value)} placeholder={t('task.templatePromptPlaceholder')} />
                <textarea className="min-h-24 w-full resize-none rounded-2xl border border-border/70 bg-background px-4 py-3 text-sm outline-none" value={newTemplateRequirement} onChange={(event) => setNewTemplateRequirement(event.target.value)} placeholder={t('task.templateRequirementPlaceholder')} />
                <Button type="button" className="rounded-2xl" onClick={createTemplate}><SquarePen className="size-4" />{t('task.createTemplate')}</Button>
              </div>
            </CardContent>
          </Card>
        ) : null}

        {tab === 'profiles' ? (
          <Card className="rounded-3xl border border-border/70 bg-background/75">
            <CardHeader>
              <CardTitle className="text-base">{t('profiles.title')}</CardTitle>
              <CardDescription>{t('profiles.desc')}</CardDescription>
            </CardHeader>
            <CardContent className="grid gap-4 xl:grid-cols-[18rem_minmax(0,1fr)]">
              <div className="max-h-[28rem] space-y-3 overflow-y-auto pr-1">
                {profiles.length === 0 ? <EmptyLine text={t('profiles.empty')} /> : profiles.map((profile) => (
                  <button key={profile.id} type="button" className={cn('w-full rounded-2xl border border-border/70 bg-background/80 p-3 text-left transition hover:bg-muted/40', selectedProfileID === profile.id && 'border-primary/60')} onClick={() => setSelectedProfileID(profile.id)}>
                    <div className="flex items-center justify-between gap-2">
                      <div className="min-w-0 text-sm font-medium text-foreground">{profile.name}</div>
                      <Badge variant="outline">v{profile.version}</Badge>
                    </div>
                    <div className="mt-2 text-xs text-muted-foreground">{profile.explanation || t('profiles.noActiveHint')}</div>
                  </button>
                ))}
              </div>
              <div className="space-y-3">
                {selectedProfile ? (
                  <>
                    <div className="rounded-2xl border border-border/70 bg-background/80 p-4">
                      <div className="flex flex-wrap items-center gap-2">
                        <div className="text-base font-semibold text-foreground">{selectedProfile.name}</div>
                        <Badge variant="outline">#{selectedProfile.id}</Badge>
                        <Badge variant="outline">{selectedProfile.status}</Badge>
                      </div>
                      <div className="mt-3 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
                        <MiniStat label={t('profiles.fields.summary')} value={selectedProfile.explanation || t('profiles.noActiveLabel')} />
                        <MiniStat label={tSetting('aiAutomationSource.profileConfidence')} value={`${Math.round((selectedProfile.confidence ?? 0) * 100)}%`} />
                        <MiniStat label={t('profiles.migrationStatus')} value={selectedProfile.migration_status || '-'} />
                        <MiniStat label={tSetting('aiAutomationSource.profileUpdatedAt')} value={formatDateTimeByLocale(selectedProfile.updated_at, locale)} />
                      </div>
                      <div className="mt-3 flex flex-wrap gap-2">
                        <Button type="button" className="rounded-2xl" onClick={() => applyProfile(selectedProfile)} disabled={activateProfile.isPending}><Bot className="size-4" />{t('profiles.activate')}</Button>
                      </div>
                    </div>
                    <div data-testid="ai-profile-structured-preview" className="max-h-[18rem] overflow-y-auto rounded-2xl border border-border/70 bg-muted/20 p-4 text-xs leading-6 text-foreground">
                      <pre className="whitespace-pre-wrap break-words">{JSON.stringify(selectedProfile.domain_payload ?? selectedProfile.versions ?? {}, null, 2)}</pre>
                    </div>
                  </>
                ) : <EmptyLine text={t('profiles.empty')} />}
              </div>
            </CardContent>
          </Card>
        ) : null}

        {tab === 'history' ? (
          <Card className="rounded-3xl border border-border/70 bg-background/75">
            <CardHeader>
              <CardTitle className="text-base">{t('taskHistory.historyTitle')}</CardTitle>
              <CardDescription>{t('taskHistory.historyDesc')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-3 md:grid-cols-4">
                <Select value={taskHistoryStatus} onValueChange={(value) => setTaskHistoryStatus(value as typeof taskHistoryStatus)}>
                  <SelectTrigger aria-label={t('taskHistory.historyStatus')} className="rounded-2xl"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">{t('taskHistory.historyAllStatuses')}</SelectItem>
                    {['pending', 'running', 'recoverable', 'succeeded', 'failed', 'failed_unrecoverable', 'canceled'].map((value) => <SelectItem key={value} value={value}>{taskStatusLabel(locale, value)}</SelectItem>)}
                  </SelectContent>
                </Select>
                <Select value={taskHistoryType} onValueChange={(value) => setTaskHistoryType(value as typeof taskHistoryType)}>
                  <SelectTrigger aria-label={t('taskHistory.historyType')} className="rounded-2xl"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">{t('taskHistory.historyAllTypes')}</SelectItem>
                    {TASK_TYPES.map((item) => <SelectItem key={item.key} value={item.key}>{t(item.label)}</SelectItem>)}
                  </SelectContent>
                </Select>
                <Input value={taskHistoryKeyword} onChange={(event) => setTaskHistoryKeyword(event.target.value)} placeholder={t('taskHistory.historyKeywordPlaceholder')} className="rounded-2xl md:col-span-2" />
              </div>
              <div className="max-h-[24rem] space-y-3 overflow-y-auto pr-1">
                {taskHistory.isLoading ? <EmptyLine text={t('taskHistory.historyLoading')} /> : null}
                {!taskHistory.isLoading && taskHistoryItems.length === 0 ? <EmptyLine text={t('taskHistory.historyEmpty')} /> : null}
                {taskHistoryItems.map((item) => (
                  <button key={item.id} type="button" className={cn('w-full rounded-2xl border border-border/70 bg-background/80 p-3 text-left transition hover:bg-muted/40', selectedHistoryTaskID === item.id && 'border-primary/60')} onClick={() => { setSelectedHistoryTaskID(item.id); setSelectedTaskID(undefined); }}>
                    <div className="flex items-center justify-between gap-2">
                      <div className="text-sm font-medium text-foreground">#{item.id} · {taskTypeLabel(t, item.type as TaskType)}</div>
                      <Badge className={cn('rounded-full px-2 py-0.5 text-xs', toneClass(item.status))}>{taskStatusLabel(locale, item.status)}</Badge>
                    </div>
                    <div className="mt-1 text-xs text-muted-foreground">{item.result_summary || item.error_message || t('taskHistory.historyResumeState')}</div>
                    <div className="mt-2 text-[11px] text-muted-foreground">{t('taskHistory.historyUpdated')} · {formatDateTimeByLocale(item.updated_at, locale)}</div>
                  </button>
                ))}
              </div>
              <div className="flex items-center justify-between gap-3">
                <div className="text-xs text-muted-foreground">{t('taskHistory.historyPage', { page: taskHistoryPage, pages: taskHistoryPageCount })}</div>
                <div className="flex gap-2">
                  <Button type="button" variant="outline" size="sm" className="rounded-xl" onClick={() => setTaskHistoryPage((current) => Math.max(1, current - 1))} disabled={taskHistoryPage <= 1}>{t('taskHistory.historyPrev')}</Button>
                  <Button type="button" variant="outline" size="sm" className="rounded-xl" onClick={() => setTaskHistoryPage((current) => Math.min(taskHistoryPageCount, current + 1))} disabled={taskHistoryPage >= taskHistoryPageCount}>{t('taskHistory.historyNext')}</Button>
                </div>
              </div>
            </CardContent>
          </Card>
        ) : null}

        <div ref={bindWorkbenchSection('learning')} data-ai-focus-target="learning" data-testid="ai-automation-learning-section">
          <Card data-testid="ai-automation-learning-stage-card" className="rounded-3xl border border-border/70 bg-background/75">
            <CardHeader>
              <CardTitle className="text-base">{t('learning.stageTitle')}</CardTitle>
              <CardDescription>{t('learning.stageDesc')}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
                <MiniStat label={t('learning.summary.statusLabel')} value={learningEnabled ? t('learning.enabled') : t('learning.disabled')} />
                <MiniStat label={t('learning.summary.samplesLabel')} value={learningDisplay.sampleCount} />
                <MiniStat label={t('learning.summary.latestSampleLabel')} value={formatLearningSampleTime(learningSummaryView.summary.latestState?.last_sample_at, locale, t('learning.summary.notAvailable'))} />
                <MiniStat label={t('learning.summary.topTargetLabel')} value={learningSummaryView.topTargetLabel} />
              </div>
            </CardContent>
          </Card>

          <div data-testid="ai-automation-learning-controls" className="mt-4 grid gap-4 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
            <Card data-testid="ai-automation-learning-preset-card" className="rounded-3xl border border-border/70 bg-background/75">
              <CardHeader>
                <CardTitle className="text-base">{t('learning.presetTitle')}</CardTitle>
                <CardDescription>{t('learning.presetDesc')}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <div data-testid="ai-automation-learning-presets" className="flex flex-wrap gap-2">
                  {(['safe', 'balanced', 'aggressive'] as LearningPreset[]).map((preset) => (
                    <button
                      key={preset}
                      type="button"
                      data-testid={`ai-automation-learning-preset-${preset}`}
                      aria-pressed={learningPreset === preset}
                      onClick={() => {
                        setLearningPreset(preset);
                        const enabled = preset !== 'safe';
                        toggleLearning(enabled);
                      }}
                      className={cn(
                        'rounded-2xl border px-3 py-2 text-left text-sm transition',
                        learningPreset === preset ? 'border-primary bg-primary text-primary-foreground' : 'border-border/70 bg-background/70 hover:bg-muted/60',
                      )}
                    >
                      {t(`learning.presets.${preset}`)}
                    </button>
                  ))}
                </div>
                <div className="text-xs text-muted-foreground">{t('learning.presetHint')}</div>
              </CardContent>
            </Card>

            <Card data-testid="ai-automation-learning-switch-card" className="rounded-3xl border border-border/70 bg-background/75">
              <CardHeader>
                <CardTitle className="text-base">{t('learning.switchTitle')}</CardTitle>
                <CardDescription>{t('learning.switchDesc')}</CardDescription>
              </CardHeader>
              <CardContent className="flex items-center justify-between gap-4">
                <div data-testid="ai-automation-learning-switch-state" className="text-sm text-muted-foreground">
                  {learningEnabled ? t('learning.enabled') : t('learning.disabled')}
                </div>
                <Switch data-testid="ai-automation-learning-switch" aria-label={t('learning.switchTitle')} checked={learningEnabled} onCheckedChange={(enabled) => { setLearningEnabledDraft(enabled); toggleLearning(enabled); }} />
              </CardContent>
            </Card>
          </div>

          <div className="mt-4">
            <LearningSummaryPanel
              primaryGrid={{
                items: learningPrimaryItems,
                testId: 'ai-automation-learning-state-summary',
                className: 'grid gap-2 sm:grid-cols-3',
                cardClassName: 'rounded-xl border border-border/60 bg-background/80 px-3 py-2',
              }}
              secondaryGrid={learningSecondaryItems.length > 0 ? {
                items: learningSecondaryItems,
                testId: 'ai-automation-learning-secondary-summary',
                className: 'grid gap-2 sm:grid-cols-2',
                cardClassName: 'rounded-xl border border-border/60 bg-background/80 px-3 py-2',
              } : undefined}
              noticeTitle={t('learning.summary.noticeTitle')}
              noticeBody={learningSummaryView.notice}
              noticeClassName="rounded-xl border border-emerald-500/20 bg-emerald-500/5 px-3 py-2 text-xs leading-5 text-muted-foreground"
              footer={
                <LearningSummaryActionBar
                  actions={[
                    {
                      key: 'reset',
                      label: t('learning.reset'),
                      onClick: handleResetLearning,
                      testId: 'ai-automation-learning-reset',
                      disabled: resetLearning.isPending || !learningDisplay.canReset,
                    },
                  ]}
                  hint={!learningDisplay.canReset ? t('learning.resetDisabledHint') : null}
                />
              }
            />
          </div>

          <div className="mt-4">
            {learningStates.length > 0 ? (
              <div data-testid="ai-automation-learning-states" className="grid gap-3 md:grid-cols-2">
                {learningStates.map((state) => (
                  <div key={state.id} data-testid={`ai-automation-learning-state-${state.id}`} className="rounded-2xl border border-border/70 bg-background/80 p-4">
                    <div className="font-medium text-foreground">{state.model_name}</div>
                    <div className="mt-1 text-xs text-muted-foreground">{t('learning.channelKeySummary', { channel: state.channel_id, key: state.channel_key_id })}</div>
                    <div className="mt-3 grid gap-2 sm:grid-cols-2">
                      <MiniStat label={t('learning.success')} value={state.success_count} />
                      <MiniStat label={t('learning.failure')} value={state.failure_count} />
                      <MiniStat label={t('learning.fallback')} value={state.fallback_count} />
                      <MiniStat label={t('learning.raceWinner')} value={state.race_winner_count} />
                    </div>
                    <div className="mt-3 text-xs text-muted-foreground">{t('learning.lastSample', { time: formatLearningSampleTime(state.last_sample_at, locale, t('learning.summary.notAvailable')) })}</div>
                  </div>
                ))}
              </div>
            ) : <EmptyLine testId="ai-automation-learning-empty" text={t('learning.empty')} />}
          </div>
        </div>
      </section>
    </PageWrapper>
  );
}
