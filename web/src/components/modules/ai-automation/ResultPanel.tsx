'use client';

import type { AITask } from '@/api/endpoints/ai-automation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { formatDateTimeByLocale } from '@/lib/locale';
import { type Locale as AppLocale } from '@/stores/setting';

import { LearningSummaryActionBar } from './LearningSummaryPanel';
import { compactObjectKeys, type ResultConsumptionView, type ResultDiffSummary, summarizeProtectedAction } from './result-logic';
import { Chip, MiniStat, OUTPUT_STYLES, RISK_PREFS, SubtleAction, ViewMode, VIEW_MODES } from './workbench-shared';

type TranslationFn = (key: string, values?: Record<string, any>) => string;

type ResultPanelProps = {
  t: TranslationFn;
  locale: AppLocale;
  activeTask?: AITask | null;
  viewMode: ViewMode;
  setViewMode: (value: ViewMode) => void;
  outputStyle: string;
  riskPreference: string;
  resultView: ResultConsumptionView;
  diffSummary: ResultDiffSummary;
  onActivateProfile?: () => void;
  activateDisabled?: boolean;
};

export function ResultPanel(props: ResultPanelProps) {
  const { t, locale, activeTask, viewMode, setViewMode, outputStyle, riskPreference, resultView, diffSummary, onActivateProfile, activateDisabled } = props;

  const viewText = viewMode === 'summary'
    ? (resultView.summary || activeTask?.error_message || t('result.empty'))
    : viewMode === 'diff'
      ? `${t('task.diffSummary', { added: diffSummary.added.length, removed: diffSummary.removed.length, shared: diffSummary.shared.length })}\n\n${compactObjectKeys(resultView.domainPayload)}`
      : (resultView.rawOutput || t('result.empty'));

  return (
    <Card className="rounded-[28px] border border-border/70 bg-background/75">
      <CardHeader className="pb-3">
        <CardTitle className="text-base">{t('result.title')}</CardTitle>
        <CardDescription>{activeTask?.result_summary || activeTask?.error_message || t('result.idleSummary')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-3 sm:grid-cols-4">
          <MiniStat label={t('task.progressTitle')} value={`${activeTask?.progress ?? 0}%`} />
          <MiniStat label={t('task.executionToneLabel')} value={t(OUTPUT_STYLES.find((item) => item.key === outputStyle)?.label ?? 'outputStyles.balanced')} />
          <MiniStat label={t('task.executionRiskLabel')} value={t(RISK_PREFS.find((item) => item.key === riskPreference)?.label ?? 'riskPrefs.normal')} />
          <MiniStat label={t('result.updatedAt')} value={activeTask?.updated_at ? formatDateTimeByLocale(activeTask.updated_at, locale) : '-'} />
        </div>

        <Progress value={activeTask?.progress ?? 0} />

        <div className="flex flex-wrap gap-2">
          {VIEW_MODES.map((item) => <Chip key={item.key} active={viewMode === item.key} label={t(item.label)} onClick={() => setViewMode(item.key)} />)}
        </div>

        <div data-testid={`result-view-${viewMode}`} className="min-h-40 rounded-2xl border border-border/70 bg-muted/20 p-4 text-sm leading-6 text-foreground whitespace-pre-wrap">
          {viewText}
        </div>

        <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(0,0.9fr)]">
          <div className="rounded-2xl border border-border/70 bg-background/80 p-4">
            <div className="text-sm font-medium text-foreground">{t('result.profileOutputTitle')}</div>
            <div className="mt-2 text-xs leading-5 text-muted-foreground">{resultView.resultProfileID ? t('result.profileReady') : t('result.profileNotReady')}</div>
            <div className="mt-3 flex flex-wrap gap-2">
              <SubtleAction type="button" onClick={onActivateProfile} disabled={activateDisabled || !resultView.resultProfileID}>{t('result.activateProfile')}</SubtleAction>
            </div>
          </div>

          <div className="rounded-2xl border border-border/70 bg-background/80 p-4">
            <div className="text-sm font-medium text-foreground">{t('result.protectedActionsTitle')}</div>
            <div className="mt-3 max-h-32 space-y-2 overflow-y-auto pr-1">
              {resultView.protectedActions.length === 0 ? <div className="rounded-xl border border-dashed border-border/70 bg-muted/20 px-3 py-3 text-xs text-muted-foreground">{t('result.noProtectedActions')}</div> : resultView.protectedActions.map((action, index) => {
                const item = summarizeProtectedAction(action);
                return (
                  <div key={`${item.key}-${index}`} className="rounded-xl border border-border/60 bg-muted/20 px-3 py-2 text-xs">
                    <div className="font-medium text-foreground">{item.key}</div>
                    <div className="mt-1 text-muted-foreground">{item.executed ? t('result.executed') : t('result.notExecuted')}</div>
                    {item.reason ? <div className="mt-1 text-muted-foreground">{item.reason}</div> : null}
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        <LearningSummaryActionBar
          actions={resultView.resultProfileID ? [{ key: 'activate', label: t('result.activateProfile'), onClick: onActivateProfile ?? (() => undefined), disabled: activateDisabled }] : []}
          hint={resultView.toolExecutionSummary ? t('result.toolExecutionHint') : undefined}
        />
      </CardContent>
    </Card>
  );
}
