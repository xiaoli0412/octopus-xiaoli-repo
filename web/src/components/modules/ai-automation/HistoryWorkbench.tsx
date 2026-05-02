'use client';

import type { AITask } from '@/api/endpoints/ai-automation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { formatDateTimeByLocale } from '@/lib/locale';
import { type Locale as AppLocale } from '@/stores/setting';

import { EmptyLine, StatusBadge, SubtleAction, TaskType, TASK_TYPES, taskStatusLabel, taskTypeLabel } from './workbench-shared';

type TranslationFn = (key: string, values?: Record<string, any>) => string;

type HistoryWorkbenchProps = {
  t: TranslationFn;
  locale: AppLocale;
  taskHistoryStatus: string;
  setTaskHistoryStatus: (value: string) => void;
  taskHistoryType: string;
  setTaskHistoryType: (value: string) => void;
  taskHistoryKeyword: string;
  setTaskHistoryKeyword: (value: string) => void;
  taskHistoryItems: AITask[];
  taskHistoryPage: number;
  taskHistoryPageCount: number;
  setTaskHistoryPage: (updater: (current: number) => number) => void;
  loading: boolean;
  selectedHistoryTaskID?: number;
  openTask: (id: number) => void;
};

export function HistoryWorkbench(props: HistoryWorkbenchProps) {
  const {
    t,
    locale,
    taskHistoryStatus,
    setTaskHistoryStatus,
    taskHistoryType,
    setTaskHistoryType,
    taskHistoryKeyword,
    setTaskHistoryKeyword,
    taskHistoryItems,
    taskHistoryPage,
    taskHistoryPageCount,
    setTaskHistoryPage,
    loading,
    selectedHistoryTaskID,
    openTask,
  } = props;

  return (
    <Card className="rounded-[28px] border border-border/70 bg-background/75">
      <CardHeader>
        <CardTitle className="text-base">{t('taskHistory.historyTitle')}</CardTitle>
        <CardDescription>{t('taskHistory.historyDesc')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-3 md:grid-cols-4">
          <Select value={taskHistoryStatus} onValueChange={setTaskHistoryStatus}>
            <SelectTrigger aria-label={t('taskHistory.historyStatus')} className="rounded-2xl"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">{t('taskHistory.historyAllStatuses')}</SelectItem>
              {['pending', 'running', 'recoverable', 'succeeded', 'failed', 'failed_unrecoverable', 'canceled'].map((value) => <SelectItem key={value} value={value}>{taskStatusLabel(locale, value)}</SelectItem>)}
            </SelectContent>
          </Select>
          <Select value={taskHistoryType} onValueChange={setTaskHistoryType}>
            <SelectTrigger aria-label={t('taskHistory.historyType')} className="rounded-2xl"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">{t('taskHistory.historyAllTypes')}</SelectItem>
              {TASK_TYPES.map((item) => <SelectItem key={item.key} value={item.key}>{t(item.label)}</SelectItem>)}
            </SelectContent>
          </Select>
          <Input value={taskHistoryKeyword} onChange={(event) => setTaskHistoryKeyword(event.target.value)} placeholder={t('taskHistory.historyKeywordPlaceholder')} className="rounded-2xl md:col-span-2" />
        </div>

        <div className="max-h-[24rem] space-y-3 overflow-y-auto pr-1">
          {loading ? <EmptyLine text={t('taskHistory.historyLoading')} /> : null}
          {!loading && taskHistoryItems.length === 0 ? <EmptyLine text={t('taskHistory.historyEmpty')} /> : null}
          {taskHistoryItems.map((item) => (
            <button key={item.id} type="button" className={`w-full rounded-2xl border p-3 text-left transition ${selectedHistoryTaskID === item.id ? 'border-primary/60 bg-primary/5' : 'border-border/70 bg-background/80 hover:bg-muted/40'}`} onClick={() => openTask(item.id)}>
              <div className="flex items-center justify-between gap-2">
                <div className="text-sm font-medium text-foreground">#{item.id} · {taskTypeLabel(t, item.type as TaskType)}</div>
                <StatusBadge locale={locale} status={item.status} />
              </div>
              <div className="mt-1 text-xs text-muted-foreground">{item.result_summary || item.error_message || t('taskHistory.historyResumeState')}</div>
              <div className="mt-2 text-[11px] text-muted-foreground">{t('taskHistory.historyUpdated')} · {formatDateTimeByLocale(item.updated_at, locale)}</div>
            </button>
          ))}
        </div>

        <div className="flex items-center justify-between gap-3">
          <div className="text-xs text-muted-foreground">{t('taskHistory.historyPage', { page: taskHistoryPage, pages: taskHistoryPageCount })}</div>
          <div className="flex gap-2">
            <SubtleAction type="button" onClick={() => setTaskHistoryPage((current) => Math.max(1, current - 1))} disabled={taskHistoryPage <= 1}>{t('taskHistory.historyPrev')}</SubtleAction>
            <SubtleAction type="button" onClick={() => setTaskHistoryPage((current) => Math.min(taskHistoryPageCount, current + 1))} disabled={taskHistoryPage >= taskHistoryPageCount}>{t('taskHistory.historyNext')}</SubtleAction>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
