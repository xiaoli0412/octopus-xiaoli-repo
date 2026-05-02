'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

import { MiniStat } from './workbench-shared';

type TranslationFn = (key: string, values?: Record<string, any>) => string;

type SideStatusCardProps = {
  t: TranslationFn;
  effectiveTaskModel: string;
  selectedConfigSourceText: string;
  currentProfileLabel: string;
  learningEnabled: boolean;
  learningSampleCount: number;
  latestSnapshotLabel?: string;
  latestTaskLabel?: string;
};

export function SideStatusCard(props: SideStatusCardProps) {
  const { t, effectiveTaskModel, selectedConfigSourceText, currentProfileLabel, learningEnabled, learningSampleCount, latestSnapshotLabel, latestTaskLabel } = props;

  return (
    <Card className="rounded-[28px] border-border/70 bg-card/95">
      <CardHeader>
        <CardTitle className="text-base">{t('sidebar.title')}</CardTitle>
        <CardDescription>{t('sidebar.desc')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <MiniStat label={t('sidebar.executionModel')} value={effectiveTaskModel || '-'} />
        <MiniStat label={t('sidebar.sourceMode')} value={selectedConfigSourceText} />
        <MiniStat label={t('sidebar.activeProfile')} value={currentProfileLabel} />
        <MiniStat label={t('sidebar.learningState')} value={learningEnabled ? t('learning.enabled') : t('learning.disabled')} />
        <MiniStat label={t('learning.summary.samplesLabel')} value={learningSampleCount} />
        <MiniStat label={t('sidebar.latestSnapshot')} value={latestSnapshotLabel || t('task.latestSnapshotEmpty')} />
        <MiniStat label={t('sidebar.latestTask')} value={latestTaskLabel || t('result.idleSummary')} />
      </CardContent>
    </Card>
  );
}
