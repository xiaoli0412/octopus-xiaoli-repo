'use client';

import type { ReactNode } from 'react';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { type Locale as AppLocale } from '@/stores/setting';

export type TaskType = 'natural_language' | 'group_suggestion' | 'channel_recognition' | 'price_recognition' | 'model_classification' | 'config_health_check' | 'dynamic_routing_digest';
export type SplitMode = 'single' | 'planner_executor' | 'planner_executor_reviewer' | 'router_mesh';
export type DispatchMode = 'sequential' | 'parallel';
export type ContextScope = 'workspace_full' | 'channels_groups' | 'pricing_models' | 'routing_learning';
export type OperatorMode = 'ops_captain' | 'cartographer' | 'forensics';
export type ViewMode = 'summary' | 'diff' | 'raw';
export type LearningPreset = 'safe' | 'balanced' | 'aggressive';
export type LaneRole = 'primary' | 'planner' | 'executor' | 'reviewer' | 'guard';
export type WorkspaceTab = 'result' | 'profiles' | 'dispatch' | 'assets' | 'history';
export type ToolKey = 'channel_inventory' | 'group_topology' | 'price_catalog' | 'model_catalog' | 'route_overrides' | 'learning_state' | 'profile_write' | 'profile_activate' | 'snapshot_guard';

export type LaneOverride = {
  inheritMain: boolean;
  baseURL: string;
  apiKey: string;
  channelType: string;
  useLocalDefault: boolean;
  model: string;
};

export type Snapshot = {
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

export const SNAPSHOT_STORAGE_KEY = 'octopus-ai-automation-snapshots';
export const MAX_SNAPSHOTS = 8;
export const DYNAMIC_ROUTING_LEARNING_KEY = 'dynamic_routing_learning_enabled';
export const AI_AUTOMATION_REDACTED_SECRET = '[redacted]';

export const TASK_TYPES: Array<{ key: TaskType; label: string }> = [
  { key: 'natural_language', label: 'task.types.natural_language' },
  { key: 'group_suggestion', label: 'task.types.group_suggestion' },
  { key: 'channel_recognition', label: 'task.types.channel_recognition' },
  { key: 'price_recognition', label: 'task.types.price_recognition' },
  { key: 'model_classification', label: 'task.types.model_classification' },
  { key: 'config_health_check', label: 'task.types.config_health_check' },
  { key: 'dynamic_routing_digest', label: 'task.types.dynamic_routing_digest' },
];

export const QUICK_INTENTS: Array<{ key: string; taskType: TaskType; label: string; text: string }> = [
  { key: 'group_review', taskType: 'group_suggestion', label: 'quickIntents.groupReview', text: 'quickIntents.groupReviewText' },
  { key: 'channel_scan', taskType: 'channel_recognition', label: 'quickIntents.channelScan', text: 'quickIntents.channelScanText' },
  { key: 'price_audit', taskType: 'price_recognition', label: 'quickIntents.priceAudit', text: 'quickIntents.priceAuditText' },
  { key: 'routing_digest', taskType: 'dynamic_routing_digest', label: 'quickIntents.routingDigest', text: 'quickIntents.routingDigestText' },
];

export const VIEW_MODES = [
  { key: 'summary', label: 'viewModes.summary' },
  { key: 'diff', label: 'viewModes.diff' },
  { key: 'raw', label: 'viewModes.raw' },
] as const;

export const CONTEXT_SCOPES = [
  { key: 'workspace_full', label: 'task.contextScopes.workspace_full' },
  { key: 'channels_groups', label: 'task.contextScopes.channels_groups' },
  { key: 'pricing_models', label: 'task.contextScopes.pricing_models' },
  { key: 'routing_learning', label: 'task.contextScopes.routing_learning' },
] as const;

export const OPERATOR_MODES = [
  { key: 'ops_captain', label: 'task.operatorModes.ops_captain' },
  { key: 'cartographer', label: 'task.operatorModes.cartographer' },
  { key: 'forensics', label: 'task.operatorModes.forensics' },
] as const;

export const SPLIT_MODES = [
  { key: 'single', label: 'task.splitModes.single' },
  { key: 'planner_executor', label: 'task.splitModes.planner_executor' },
  { key: 'planner_executor_reviewer', label: 'task.splitModes.planner_executor_reviewer' },
  { key: 'router_mesh', label: 'task.splitModes.router_mesh' },
] as const;

export const OUTPUT_STYLES = [
  { key: 'concise', label: 'outputStyles.concise' },
  { key: 'balanced', label: 'outputStyles.balanced' },
  { key: 'operational', label: 'outputStyles.operational' },
] as const;

export const RISK_PREFS = [
  { key: 'safe', label: 'riskPrefs.safe' },
  { key: 'normal', label: 'riskPrefs.normal' },
  { key: 'aggressive', label: 'riskPrefs.aggressive' },
] as const;

export const TOOL_OPTIONS: Array<{ key: ToolKey; label: string }> = [
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

export const DEFAULT_TOOL_KEYS: ToolKey[] = ['channel_inventory', 'group_topology', 'model_catalog', 'learning_state', 'profile_write', 'snapshot_guard'];

export const WORKSPACE_TABS: Array<{ key: WorkspaceTab; label: string }> = [
  { key: 'result', label: 'views.result' },
  { key: 'profiles', label: 'views.profiles' },
  { key: 'dispatch', label: 'views.dispatch' },
  { key: 'assets', label: 'views.assets' },
  { key: 'history', label: 'views.history' },
];

const TASK_STATUS_LABELS: Record<AppLocale, Record<string, string>> = {
  'zh-Hans': { idle: '空闲', pending: '待处理', running: '运行中', recoverable: '可恢复', succeeded: '成功', failed: '失败', failed_unrecoverable: '不可恢复失败', canceled: '已取消' },
  'zh-Hant': { idle: '空閒', pending: '待處理', running: '執行中', recoverable: '可恢復', succeeded: '成功', failed: '失敗', failed_unrecoverable: '不可恢復失敗', canceled: '已取消' },
  en: { idle: 'Idle', pending: 'Pending', running: 'Running', recoverable: 'Recoverable', succeeded: 'Succeeded', failed: 'Failed', failed_unrecoverable: 'Unrecoverable failure', canceled: 'Canceled' },
  ja: { idle: '待機', pending: '保留', running: '実行中', recoverable: '復旧可能', succeeded: '成功', failed: '失敗', failed_unrecoverable: '回復不能失敗', canceled: 'キャンセル' },
};

export function canUseWindow() {
  return typeof window !== 'undefined';
}

export function createLaneOverrides(): Record<Exclude<LaneRole, 'primary'>, LaneOverride> {
  const base: LaneOverride = { inheritMain: true, baseURL: '', apiKey: '', channelType: 'openai-compatible', useLocalDefault: true, model: '' };
  return { planner: { ...base }, executor: { ...base }, reviewer: { ...base }, guard: { ...base } };
}

export function loadSnapshots(): Snapshot[] {
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

export function saveSnapshots(next: Snapshot[]) {
  if (!canUseWindow()) return;
  window.localStorage.setItem(SNAPSHOT_STORAGE_KEY, JSON.stringify(next.slice(0, MAX_SNAPSHOTS)));
}

export function makeSnapshot(label: string, payload: Snapshot['payload']): Snapshot {
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

export function isRedactedAIAutomationSecret(value: string) {
  return value.trim() === AI_AUTOMATION_REDACTED_SECRET;
}

export function apiKeyForMutation(value: string) {
  const trimmed = value.trim();
  if (trimmed === AI_AUTOMATION_REDACTED_SECRET) return undefined;
  return trimmed;
}

export function apiKeyForRequest(value: string) {
  const trimmed = value.trim();
  if (trimmed === '' || trimmed === AI_AUTOMATION_REDACTED_SECRET) return undefined;
  return trimmed;
}

export function laneRolesFromSplitMode(mode: SplitMode): LaneRole[] {
  if (mode === 'single') return ['primary'];
  if (mode === 'planner_executor') return ['planner', 'executor'];
  if (mode === 'planner_executor_reviewer') return ['planner', 'executor', 'reviewer'];
  return ['planner', 'executor', 'reviewer', 'guard'];
}

export function toneClass(status?: string) {
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

export function joinPrompt(lines: string[]) {
  return lines.filter((item) => item.trim().length > 0).join('\n\n');
}

export function taskTypeLabel(t: (key: string) => string, taskType: TaskType) {
  return t(TASK_TYPES.find((item) => item.key === taskType)?.label ?? 'task.types.natural_language');
}

export function taskStatusLabel(locale: AppLocale, status?: string) {
  return TASK_STATUS_LABELS[locale][status ?? 'idle'] ?? status ?? '-';
}

export function laneLabelKey(lane: LaneRole) {
  return `task.laneLabels.${lane}` as const;
}

export function lanePromptKey(lane: LaneRole) {
  return `task.lanePrompts.${lane}` as const;
}

export function EmptyLine({ testId, text }: { testId?: string; text: string }) {
  return <div data-testid={testId} className="rounded-2xl border border-dashed border-border/70 bg-muted/20 px-4 py-5 text-sm text-muted-foreground">{text}</div>;
}

export function MiniStat({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="rounded-2xl border border-border/70 bg-background/75 p-3">
      <div className="text-[11px] uppercase tracking-[0.12em] text-muted-foreground">{label}</div>
      <div className="mt-1 break-all text-sm font-medium text-foreground">{value}</div>
    </div>
  );
}

export function Chip({ active, label, onClick, testId }: { active: boolean; label: string; onClick: () => void; testId?: string }) {
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

export function StatusBadge({ locale, status }: { locale: AppLocale; status?: string }) {
  return <Badge className={cn('rounded-full px-3 py-1 text-xs', toneClass(status))}>{taskStatusLabel(locale, status)}</Badge>;
}

export function SubtleAction({ children, ...props }: React.ComponentProps<typeof Button>) {
  return <Button variant="outline" size="sm" className={cn('rounded-xl', props.className)} {...props}>{children}</Button>;
}
