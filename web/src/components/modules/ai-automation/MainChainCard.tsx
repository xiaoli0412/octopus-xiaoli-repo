'use client';

import type { ReactNode } from 'react';
import { Bot, Layers3, Sparkles } from 'lucide-react';

import type { AITask } from '@/api/endpoints/ai-automation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { type Locale as AppLocale } from '@/stores/setting';

import {
  Chip,
  CONTEXT_SCOPES,
  ContextScope,
  MiniStat,
  OPERATOR_MODES,
  OperatorMode,
  QUICK_INTENTS,
  StatusBadge,
  TASK_TYPES,
  TaskType,
} from './workbench-shared';

type TranslationFn = (key: string, values?: Record<string, any>) => string;

type MainChainCardProps = {
  t: TranslationFn;
  locale: AppLocale;
  enabled: boolean;
  taskType: TaskType;
  setTaskType: (value: TaskType) => void;
  contextScope: ContextScope;
  setContextScope: (value: ContextScope) => void;
  operatorMode: OperatorMode;
  setOperatorMode: (value: OperatorMode) => void;
  naturalInput: string;
  setNaturalInput: (value: string) => void;
  customPrompt: string;
  setCustomPrompt: (value: string) => void;
  selectedQuickIntent: string;
  applyQuickIntent: (key: string) => void;
  effectiveTaskModel: string;
  selectedConfigSourceText: string;
  currentProfileLabel: string;
  runtimeFallbackNotice?: ReactNode;
  createTaskPending: boolean;
  fetchModelsPending: boolean;
  runSingle: () => void;
  fetchModelList: () => void;
  openProfiles: () => void;
  activeTask?: AITask | null;
  currentStatus?: string;
};

export function MainChainCard(props: MainChainCardProps) {
  const {
    t,
    locale,
    enabled,
    taskType,
    setTaskType,
    contextScope,
    setContextScope,
    operatorMode,
    setOperatorMode,
    naturalInput,
    setNaturalInput,
    customPrompt,
    setCustomPrompt,
    selectedQuickIntent,
    applyQuickIntent,
    effectiveTaskModel,
    selectedConfigSourceText,
    currentProfileLabel,
    runtimeFallbackNotice,
    createTaskPending,
    fetchModelsPending,
    runSingle,
    fetchModelList,
    openProfiles,
    activeTask,
    currentStatus,
  } = props;

  return (
    <Card className="overflow-hidden rounded-[28px] border-border/70 bg-card/95">
      <CardHeader className="border-b border-border/60 pb-5">
        <div className="flex flex-wrap items-center gap-2">
          <div className="rounded-full border border-border/70 bg-background/70 px-3 py-1 text-[11px] uppercase tracking-[0.18em] text-muted-foreground">
            {t('hero.badge')}
          </div>
          <StatusBadge locale={locale} status={currentStatus} />
        </div>
        <div className="space-y-1.5">
          <CardTitle className="text-[1.75rem] font-semibold tracking-tight">{t('hero.title')}</CardTitle>
          <CardDescription className="max-w-3xl text-sm leading-6 text-muted-foreground">{t('hero.desc')}</CardDescription>
        </div>
      </CardHeader>

      <CardContent className="space-y-5 p-5">
        <div className="grid gap-3 md:grid-cols-4">
          <MiniStat label={t('config.activeSourceTitle')} value={selectedConfigSourceText} />
          <MiniStat label={t('task.effectiveModelTitle')} value={effectiveTaskModel || '-'} />
          <MiniStat label={t('profiles.activeTitle')} value={currentProfileLabel} />
          <MiniStat label={t('result.currentStatus')} value={activeTask?.result_summary || t('result.idleSummary')} />
        </div>

        {runtimeFallbackNotice}

        <div className="grid gap-4 xl:grid-cols-[minmax(0,1.2fr)_minmax(16rem,0.8fr)]">
          <div className="space-y-4">
            <div className="grid gap-3 sm:grid-cols-3">
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
            </div>

            <div className="rounded-[26px] border border-border/70 bg-background/70 p-4">
              <div className="space-y-2">
                <div className="text-sm font-medium text-foreground">{t('task.currentTaskTypeTitle')}</div>
                <textarea
                  className="min-h-44 w-full resize-none rounded-2xl border border-border/70 bg-background px-4 py-3 text-sm outline-none transition focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30"
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

            <div className="flex flex-wrap items-center gap-3 rounded-[26px] border border-border/70 bg-background/70 p-4">
              <Button type="button" className="rounded-2xl" onClick={runSingle} disabled={!enabled || createTaskPending}>
                <Bot className="size-4" />
                {t('task.run')}
              </Button>
              <Button type="button" variant="outline" size="sm" className="rounded-xl" onClick={fetchModelList} disabled={!enabled || fetchModelsPending}>
                <Sparkles className="size-4" />
                {t('config.fetchModels')}
              </Button>
              <Button type="button" variant="outline" size="sm" className="rounded-xl" onClick={openProfiles}>
                <Layers3 className="size-4" />
                {t('profiles.openWorkspace')}
              </Button>
            </div>

            <Card className="rounded-[26px] border border-border/70 bg-background/75">
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
            <Card className="rounded-[26px] border border-border/70 bg-background/75">
              <CardHeader className="pb-3">
                <CardTitle className="text-base">{t('config.shortTitle')}</CardTitle>
                <CardDescription>{t('config.shortDesc')}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <MiniStat label={t('config.activeSourceTitle')} value={selectedConfigSourceText} />
                <MiniStat label={t('task.effectiveModelTitle')} value={effectiveTaskModel || '-'} />
                <MiniStat label={t('profiles.activeTitle')} value={currentProfileLabel} />
                <MiniStat label={t('result.currentStatus')} value={activeTask?.result_summary || t('result.idleSummary')} />
              </CardContent>
            </Card>

            <Card className="rounded-[26px] border border-border/70 bg-background/75">
              <CardHeader className="pb-3">
                <CardTitle className="text-base">{t('task.taskType')}</CardTitle>
                <CardDescription>{t(`task.typeDescriptions.${taskType}`)}</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="rounded-2xl border border-border/70 bg-muted/20 p-4 text-sm leading-6 text-muted-foreground">
                  {t('task.mainChainHint')}
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
