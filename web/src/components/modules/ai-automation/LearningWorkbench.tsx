'use client';

import type { DynamicRouteLearningState } from '@/api/endpoints/ai-automation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Switch } from '@/components/ui/switch';
import { type Locale as AppLocale } from '@/stores/setting';

import { LearningSummaryActionBar, LearningSummaryPanel } from './LearningSummaryPanel';
import type { LearningSummaryCardItem, LearningSummaryViewModel } from './learning-summary';
import { EmptyLine, MiniStat } from './workbench-shared';

type TranslationFn = (key: string, values?: Record<string, any>) => string;

type LearningWorkbenchProps = {
  t: TranslationFn;
  locale: AppLocale;
  learningEnabled: boolean;
  learningPreset: string;
  setLearningPreset: (value: 'safe' | 'balanced' | 'aggressive') => void;
  toggleLearning: (enabled: boolean) => void;
  learningSummaryView: LearningSummaryViewModel;
  learningPrimaryItems: LearningSummaryCardItem[];
  learningSecondaryItems: LearningSummaryCardItem[];
  learningStates: DynamicRouteLearningState[];
  handleResetLearning: () => void;
};

export function LearningWorkbench(props: LearningWorkbenchProps) {
  const {
    t,
    learningEnabled,
    learningPreset,
    setLearningPreset,
    toggleLearning,
    learningSummaryView,
    learningPrimaryItems,
    learningSecondaryItems,
    learningStates,
    handleResetLearning,
  } = props;

  const learningDisplay = learningSummaryView.display;

  return (
    <div data-testid="ai-automation-learning-section" className="space-y-4">
      <Card data-testid="ai-automation-learning-stage-card" className="rounded-[28px] border border-border/70 bg-background/75">
        <CardHeader>
          <CardTitle className="text-base">{t('learning.stageTitle')}</CardTitle>
          <CardDescription>{t('learning.stageDesc')}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            <MiniStat label={t('learning.summary.statusLabel')} value={learningEnabled ? t('learning.enabled') : t('learning.disabled')} />
            <MiniStat label={t('learning.summary.samplesLabel')} value={learningDisplay.sampleCount} />
            <MiniStat label={t('learning.summary.latestSampleLabel')} value={learningSummaryView.latestSampleLabel} />
            <MiniStat label={t('learning.summary.topTargetLabel')} value={learningSummaryView.topTargetLabel} />
          </div>
        </CardContent>
      </Card>

      <div data-testid="ai-automation-learning-controls" className="grid gap-4 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
        <Card data-testid="ai-automation-learning-preset-card" className="rounded-[28px] border border-border/70 bg-background/75">
          <CardHeader>
            <CardTitle className="text-base">{t('learning.presetTitle')}</CardTitle>
            <CardDescription>{t('learning.presetDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <div data-testid="ai-automation-learning-presets" className="flex flex-wrap gap-2">
              {(['safe', 'balanced', 'aggressive'] as Array<'safe' | 'balanced' | 'aggressive'>).map((preset) => (
                <button
                  key={preset}
                  type="button"
                  data-testid={`ai-automation-learning-preset-${preset}`}
                  aria-pressed={learningPreset === preset}
                  onClick={() => {
                    setLearningPreset(preset);
                    toggleLearning(preset !== 'safe');
                  }}
                  className={`rounded-2xl border px-3 py-2 text-left text-sm transition ${learningPreset === preset ? 'border-primary bg-primary text-primary-foreground' : 'border-border/70 bg-background/70 hover:bg-muted/60'}`}
                >
                  {t(`learning.presets.${preset}`)}
                </button>
              ))}
            </div>
            <div className="text-xs text-muted-foreground">{t('learning.presetHint')}</div>
          </CardContent>
        </Card>

        <Card data-testid="ai-automation-learning-switch-card" className="rounded-[28px] border border-border/70 bg-background/75">
          <CardHeader>
            <CardTitle className="text-base">{t('learning.switchTitle')}</CardTitle>
            <CardDescription>{t('learning.switchDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="flex items-center justify-between gap-4">
            <div data-testid="ai-automation-learning-switch-state" className="text-sm text-muted-foreground">{learningEnabled ? t('learning.enabled') : t('learning.disabled')}</div>
            <Switch data-testid="ai-automation-learning-switch" aria-label={t('learning.switchTitle')} checked={learningEnabled} onCheckedChange={toggleLearning} />
          </CardContent>
        </Card>
      </div>

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
        footer={<LearningSummaryActionBar actions={[{ key: 'reset', label: t('learning.reset'), onClick: handleResetLearning, testId: 'ai-automation-learning-reset', disabled: !learningDisplay.canReset }]} hint={!learningDisplay.canReset ? t('learning.resetDisabledHint') : null} />}
      />

      <div>
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
                <div className="mt-3 text-xs text-muted-foreground">{t('learning.lastSample', { time: learningSummaryView.latestSampleLabel })}</div>
              </div>
            ))}
          </div>
        ) : <EmptyLine testId="ai-automation-learning-empty" text={t('learning.empty')} />}
      </div>
    </div>
  );
}
