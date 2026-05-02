'use client';

import { useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import { useTranslations } from 'next-intl';

import {
  useAIAutomationConfig,
  useAIProfile,
  useAIProfiles,
  useAIPromptTemplates,
  useActivateAIProfile,
  useAITask,
  useAITaskArtifacts,
  useAITasks,
  useCreateAIPromptTemplate,
  useCreateAITask,
  useDynamicRouteLearning,
  useFetchAIModels,
  useResetDynamicRouteLearning,
  useUpdateAIAutomationConfig,
  type AIAutomationTaskConfigSnapshot,
  type AIProfile,
} from '@/api/endpoints/ai-automation';
import { useSetSetting } from '@/api/endpoints/setting';
import { PageWrapper } from '@/components/common/PageWrapper';
import { toast } from '@/components/common/Toast';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { formatDateTimeByLocale } from '@/lib/locale';
import { type Locale as AppLocale, useSettingStore } from '@/stores/setting';

import { AssetsWorkbench } from './AssetsWorkbench';
import { formatConfigProfileLabel, resolveConfigSourceRuntime } from './config-source-logic';
import { DispatchWorkbench } from './DispatchWorkbench';
import { consumeAIAutomationFocusTarget } from './focus-target';
import { HistoryWorkbench } from './HistoryWorkbench';
import { LearningWorkbench } from './LearningWorkbench';
import {
  DEFAULT_LEARNING_SUMMARY_SECTION_ENTRIES,
  buildLearningSummarySections,
  buildLearningSummaryViewModel,
} from './learning-summary';
import { MainChainCard } from './MainChainCard';
import { ProfilesWorkbench } from './ProfilesWorkbench';
import { buildResultDiffSummary, consumeResultView, stringifyPanelJSON } from './result-logic';
import { ResultPanel } from './ResultPanel';
import { SideStatusCard } from './SideStatusCard';
import {
  apiKeyForMutation,
  apiKeyForRequest,
  ContextScope,
  createLaneOverrides,
  DYNAMIC_ROUTING_LEARNING_KEY,
  DEFAULT_TOOL_KEYS,
  DispatchMode,
  EmptyLine,
  laneRolesFromSplitMode,
  LaneOverride,
  LaneRole,
  LearningPreset,
  loadSnapshots,
  makeSnapshot,
  MAX_SNAPSHOTS,
  OperatorMode,
  saveSnapshots,
  Snapshot,
  SplitMode,
  TaskType,
  ToolKey,
  WORKSPACE_TABS,
  WorkspaceTab,
  taskTypeLabel,
  ViewMode,
} from './workbench-shared';

function buildProfilePreview(profile: AIProfile | undefined, locale: AppLocale, t: (key: string, values?: Record<string, any>) => string, onActivate?: () => void, activatePending?: boolean) {
  if (!profile) {
    return <EmptyLine text={t('profiles.empty')} />;
  }

  return (
    <div className="space-y-3">
      <div className="rounded-2xl border border-border/70 bg-background/80 p-4">
        <div className="flex flex-wrap items-center gap-2">
          <div className="text-base font-semibold text-foreground">{profile.name}</div>
          <div className="rounded-full border border-border/70 px-2 py-0.5 text-[11px] text-muted-foreground">#{profile.id}</div>
          <div className="rounded-full border border-border/70 px-2 py-0.5 text-[11px] text-muted-foreground">{profile.status}</div>
        </div>
        <div className="mt-3 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <div className="rounded-2xl border border-border/70 bg-background/75 p-3">
            <div className="text-[11px] uppercase tracking-[0.12em] text-muted-foreground">{t('profiles.fields.summary')}</div>
            <div className="mt-1 text-sm font-medium text-foreground">{profile.explanation || t('profiles.noActiveLabel')}</div>
          </div>
          <div className="rounded-2xl border border-border/70 bg-background/75 p-3">
            <div className="text-[11px] uppercase tracking-[0.12em] text-muted-foreground">{t('profiles.confidenceTitle')}</div>
            <div className="mt-1 text-sm font-medium text-foreground">{`${Math.round((profile.confidence ?? 0) * 100)}%`}</div>
          </div>
          <div className="rounded-2xl border border-border/70 bg-background/75 p-3">
            <div className="text-[11px] uppercase tracking-[0.12em] text-muted-foreground">{t('profiles.migrationStatus')}</div>
            <div className="mt-1 text-sm font-medium text-foreground">{profile.migration_status || '-'}</div>
          </div>
          <div className="rounded-2xl border border-border/70 bg-background/75 p-3">
            <div className="text-[11px] uppercase tracking-[0.12em] text-muted-foreground">{t('profiles.updatedAt')}</div>
            <div className="mt-1 text-sm font-medium text-foreground">{formatDateTimeByLocale(profile.updated_at, locale)}</div>
          </div>
        </div>
        <div className="mt-3 flex flex-wrap gap-2">
          <button type="button" className="inline-flex h-9 items-center justify-center rounded-2xl bg-primary px-4 text-sm font-medium text-primary-foreground transition hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50" onClick={onActivate} disabled={activatePending}>
            {t('profiles.activate')}
          </button>
        </div>
      </div>
      <div data-testid="ai-profile-structured-preview" className="max-h-[20rem] overflow-y-auto rounded-2xl border border-border/70 bg-muted/20 p-4 text-xs leading-6 text-foreground">
        <pre className="whitespace-pre-wrap break-words">{stringifyPanelJSON(profile.domain_payload ?? profile.versions ?? {}, '{}')}</pre>
      </div>
    </div>
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
  const profilesQuery = useAIProfiles();
  const activateProfile = useActivateAIProfile();
  const learning = useDynamicRouteLearning();
  const resetLearning = useResetDynamicRouteLearning();
  const setSetting = useSetSetting();

  const [workspaceTab, setWorkspaceTab] = useState<WorkspaceTab>('result');
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
  const [outputStyle, setOutputStyle] = useState<'concise' | 'balanced' | 'operational'>('balanced');
  const [riskPreference, setRiskPreference] = useState<'safe' | 'normal' | 'aggressive'>('normal');
  const [viewMode, setViewMode] = useState<ViewMode>('summary');
  const [selectedQuickIntent, setSelectedQuickIntent] = useState('group_review');
  const [selectedToolKeys, setSelectedToolKeys] = useState<ToolKey[]>(DEFAULT_TOOL_KEYS);
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
  const learningSectionRef = useRef<HTMLDivElement | null>(null);
  const pendingFocusTargetRef = useRef<'learning' | null>(null);

  const initializedRef = useRef(false);

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

  const {
    configSourceMode,
    requestedActiveProfile,
    activeProfile,
    sourceFallbackReason,
    runtimeFallbackActive,
    manualDraftBaseURL,
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

  const effectiveTaskModel = recommendedModel?.name || effectiveModel;

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

  const learningSummaryView = buildLearningSummaryViewModel({
    states: learningStates,
    enabled: learningEnabled,
    locale,
    emptyLabel: t('learning.summary.notAvailable'),
    topTargetFormatter: (state) => t('learning.topTargetValue', { model: state.model_name, channel: state.channel_id, key: state.channel_key_id }),
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

  const resultView = consumeResultView(artifacts.data, activeTask?.result_summary);
  const diffSummary = buildResultDiffSummary(selectedProfile, resultView.domainPayload);

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
    pendingFocusTargetRef.current = targetKey;
    setWorkspaceTab('history');
  }, []);

  useLayoutEffect(() => {
    if (workspaceTab !== 'history') return;
    if (pendingFocusTargetRef.current !== 'learning') return;
    const target = learningSectionRef.current;
    if (!target) return;
    target.scrollIntoView({ behavior: 'smooth', block: 'start' });
    pendingFocusTargetRef.current = null;
  }, [workspaceTab]);

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
    const intent = [
      { key: 'group_review', taskType: 'group_suggestion' as TaskType, text: 'quickIntents.groupReviewText' },
      { key: 'channel_scan', taskType: 'channel_recognition' as TaskType, text: 'quickIntents.channelScanText' },
      { key: 'price_audit', taskType: 'price_recognition' as TaskType, text: 'quickIntents.priceAuditText' },
      { key: 'routing_digest', taskType: 'dynamic_routing_digest' as TaskType, text: 'quickIntents.routingDigestText' },
    ].find((item) => item.key === key);
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

    pushSnapshot(`${t('task.snapshotBeforeLaunch')} · ${taskTypeLabel(t, taskType)}`);

    try {
      return await createTask.mutateAsync({
        type: taskType,
        input_text: naturalInput,
        context_scope: contextScope,
        prompt_template_ids: selectedTemplateIDs,
        custom_prompt: customPrompt.trim() || undefined,
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
      setWorkspaceTab('result');
      toast.success(t('toast.taskCreated'));
    }
  };

  const runMultiLane = async () => {
    const launches = [] as Array<{ lane: LaneRole; taskId: number; model: string }>;
    const activeLanes = laneRolesFromSplitMode(splitMode);
    const dispatchedLanes = dispatchMode === 'parallel' ? activeLanes.slice(0, parallelism) : activeLanes;

    for (const lane of dispatchedLanes) {
      const task = await createTaskForLane(lane);
      if (task) {
        launches.push({ lane, taskId: task.id, model: task.selected_model || currentLaneModel(lane) || '-' });
      }
    }

    if (launches.length > 0) {
      setLaneLaunches((current) => [...launches, ...current].slice(0, 8));
      setSelectedTaskID(launches[0].taskId);
      setSelectedHistoryTaskID(undefined);
      setWorkspaceTab('dispatch');
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
      toast.success(t('profiles.activate'));
    } catch (error) {
      toast.error(tSetting('aiAutomationSource.activateFailed'), { description: error instanceof Error ? error.message : undefined });
    }
  };

  const activateProfileFromResult = async () => {
    if (!resultView.resultProfileID) return;
    const resultProfile = profiles.find((item) => item.id === resultView.resultProfileID);
    if (!resultProfile) return;
    await applyProfile(resultProfile);
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

  const selectedConfigSourceText = configSourceMode === 'ai_profile' ? tSetting('aiAutomationSource.aiProfile') : tSetting('aiAutomationSource.manual');
  const currentProfileLabel = formatConfigProfileLabel(activeProfile || requestedActiveProfile) || t('profiles.noActiveLabel');
  const latestSnapshot = snapshots[0];

  const runtimeFallbackNotice = runtimeFallbackActive ? (
    <div className="rounded-2xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-900 dark:text-amber-100">
      <div className="font-medium">{tSetting('aiAutomationSource.runtimeFallbackTitle')}</div>
      <div className="mt-1 text-xs leading-5">{sourceFallbackReason ? tSetting(`aiAutomationSource.fallbackReasons.${sourceFallbackReason}`) : tSetting('aiAutomationSource.runtimeFallbackNotice')}</div>
    </div>
  ) : null;

  const selectedProfilePanel = buildProfilePreview(selectedProfile, locale, t, selectedProfile ? () => applyProfile(selectedProfile) : undefined, activateProfile.isPending);

  return (
    <PageWrapper className="space-y-5" data-testid="ai-automation-page">
      <div className="grid min-h-0 gap-5 xl:grid-cols-[minmax(0,1.65fr)_22rem]">
        <div className="space-y-5">
          <MainChainCard
            t={t}
            locale={locale}
            enabled={!!config?.enabled}
            taskType={taskType}
            setTaskType={setTaskType}
            contextScope={contextScope}
            setContextScope={setContextScope}
            operatorMode={operatorMode}
            setOperatorMode={setOperatorMode}
            naturalInput={naturalInput}
            setNaturalInput={setNaturalInput}
            customPrompt={customPrompt}
            setCustomPrompt={setCustomPrompt}
            selectedQuickIntent={selectedQuickIntent}
            applyQuickIntent={applyQuickIntent}
            effectiveTaskModel={effectiveTaskModel}
            selectedConfigSourceText={selectedConfigSourceText}
            currentProfileLabel={currentProfileLabel}
            runtimeFallbackNotice={runtimeFallbackNotice}
            createTaskPending={createTask.isPending}
            fetchModelsPending={fetchModels.isPending}
            runSingle={runSingle}
            fetchModelList={fetchModelList}
            openProfiles={() => setWorkspaceTab('profiles')}
            activeTask={activeTask}
            currentStatus={activeTask?.status || 'idle'}
          />

          <ResultPanel
            t={t}
            locale={locale}
            activeTask={activeTask}
            viewMode={viewMode}
            setViewMode={setViewMode}
            outputStyle={outputStyle}
            riskPreference={riskPreference}
            resultView={resultView}
            diffSummary={diffSummary}
            onActivateProfile={activateProfileFromResult}
            activateDisabled={activateProfile.isPending || !resultView.resultProfileID}
          />

          <section className="rounded-[28px] border border-border/70 bg-card/95 p-5">
            <div className="flex flex-wrap items-center gap-2 pb-4">
              {WORKSPACE_TABS.map((section) => (
                <button key={section.key} type="button" className={`rounded-2xl border px-3 py-2 text-left text-sm transition ${workspaceTab === section.key ? 'border-primary bg-primary text-primary-foreground' : 'border-border/70 bg-background/70 hover:bg-muted/60'}`} onClick={() => setWorkspaceTab(section.key)}>
                  {t(section.label)}
                </button>
              ))}
            </div>

            {workspaceTab === 'profiles' ? <ProfilesWorkbench t={t} profiles={profiles} selectedProfileID={selectedProfileID} selectProfile={setSelectedProfileID} selectedProfilePanel={selectedProfilePanel} /> : null}
            {workspaceTab === 'dispatch' ? <DispatchWorkbench t={t} splitMode={splitMode} setSplitMode={setSplitMode} dispatchMode={dispatchMode} setDispatchMode={setDispatchMode} parallelism={parallelism} setParallelism={setParallelism} laneOverrides={laneOverrides} updateLaneOverride={updateLaneOverride} laneLaunches={laneLaunches} modelCandidates={modelCandidates} currentLaneModel={currentLaneModel} baseURL={effectiveBaseURL} effectiveTaskModel={effectiveTaskModel} useLocalDefault={effectiveUseLocalDefault} runMultiLane={runMultiLane} disabled={!config?.enabled || createTask.isPending} /> : null}
            {workspaceTab === 'assets' ? <AssetsWorkbench t={t} templates={promptTemplates} selectedTemplateIDs={selectedTemplateIDs} toggleTemplate={(id) => setSelectedPromptTemplateIDs((current) => current.includes(id) ? current.filter((value) => value !== id) : [...current, id])} newTemplateName={newTemplateName} setNewTemplateName={setNewTemplateName} newTemplatePrompt={newTemplatePrompt} setNewTemplatePrompt={setNewTemplatePrompt} newTemplateRequirement={newTemplateRequirement} setNewTemplateRequirement={setNewTemplateRequirement} createTemplate={createTemplate} toolKeys={selectedToolKeys} toggleTool={toggleTool} snapshots={snapshots} snapshotLabel={snapshotLabel} setSnapshotLabel={setSnapshotLabel} createManualSnapshot={createManualSnapshot} restoreSnapshot={restoreSnapshot} clearSnapshots={clearSnapshots} locale={locale} /> : null}
            {workspaceTab === 'history' ? <div className="space-y-5"><HistoryWorkbench t={t} locale={locale} taskHistoryStatus={taskHistoryStatus} setTaskHistoryStatus={(value) => setTaskHistoryStatus(value as typeof taskHistoryStatus)} taskHistoryType={taskHistoryType} setTaskHistoryType={(value) => setTaskHistoryType(value as typeof taskHistoryType)} taskHistoryKeyword={taskHistoryKeyword} setTaskHistoryKeyword={setTaskHistoryKeyword} taskHistoryItems={taskHistoryItems} taskHistoryPage={taskHistoryPage} taskHistoryPageCount={taskHistoryPageCount} setTaskHistoryPage={setTaskHistoryPage} loading={taskHistory.isLoading} selectedHistoryTaskID={selectedHistoryTaskID} openTask={(id) => { setSelectedHistoryTaskID(id); setSelectedTaskID(undefined); }} /><div ref={learningSectionRef} data-ai-focus-target="learning"><LearningWorkbench t={t} locale={locale} learningEnabled={learningEnabled} learningPreset={learningPreset} setLearningPreset={setLearningPreset} toggleLearning={toggleLearning} learningSummaryView={learningSummaryView} learningPrimaryItems={learningPrimaryItems} learningSecondaryItems={learningSecondaryItems} learningStates={learningStates} handleResetLearning={handleResetLearning} /></div></div> : null}
            {workspaceTab === 'result' ? <Card className="rounded-[28px] border border-border/70 bg-background/75"><CardContent className="space-y-4 p-6"><Input aria-label={t('config.baseUrl')} value={baseURL} onChange={(event) => setBaseURL(event.target.value)} placeholder={t('config.baseUrl')} className="rounded-2xl" /><Input aria-label={t('config.apiKey')} value={apiKey} onChange={(event) => setApiKey(event.target.value)} placeholder={t('config.apiKeyPlaceholder')} className="rounded-2xl" /><Input aria-label={t('config.model')} value={modelName} onChange={(event) => setModelName(event.target.value)} placeholder={t('config.modelPlaceholder')} className="rounded-2xl" /><div className="flex items-center justify-between rounded-2xl border border-border/70 bg-muted/25 px-3 py-2"><div><div className="text-sm font-medium text-foreground">{t('config.localDefaultTitle')}</div><div className="text-xs text-muted-foreground">{t('config.localDefaultDesc')}</div></div><input aria-label={t('config.localDefaultTitle')} type="checkbox" checked={useLocalDefault} onChange={(event) => setUseLocalDefault(event.target.checked)} /></div><div className="flex flex-wrap gap-2"><button type="button" className="inline-flex h-9 items-center justify-center rounded-2xl bg-primary px-4 text-sm font-medium text-primary-foreground transition hover:bg-primary/90" onClick={saveConfig}>{t('config.save')}</button></div></CardContent></Card> : null}
          </section>
        </div>

        <aside className="space-y-5">
          <SideStatusCard t={t} effectiveTaskModel={effectiveTaskModel} selectedConfigSourceText={selectedConfigSourceText} currentProfileLabel={currentProfileLabel} learningEnabled={learningEnabled} learningSampleCount={learningSummaryView.display.sampleCount} latestSnapshotLabel={latestSnapshot?.label} latestTaskLabel={activeTask?.result_summary} />
        </aside>
      </div>
    </PageWrapper>
  );
}
