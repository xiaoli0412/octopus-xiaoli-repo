'use client';

import type { AIModelCandidate } from '@/api/endpoints/ai-automation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';

import { DispatchMode, EmptyLine, LaneOverride, LaneRole, SPLIT_MODES, SplitMode } from './workbench-shared';

type TranslationFn = (key: string, values?: Record<string, any>) => string;

type DispatchWorkbenchProps = {
  t: TranslationFn;
  splitMode: SplitMode;
  setSplitMode: (value: SplitMode) => void;
  dispatchMode: DispatchMode;
  setDispatchMode: (value: DispatchMode) => void;
  parallelism: number;
  setParallelism: (value: number) => void;
  laneOverrides: Record<Exclude<LaneRole, 'primary'>, LaneOverride>;
  updateLaneOverride: <K extends keyof LaneOverride>(lane: Exclude<LaneRole, 'primary'>, key: K, value: LaneOverride[K]) => void;
  laneLaunches: Array<{ lane: LaneRole; taskId: number; model: string }>;
  modelCandidates: AIModelCandidate[];
  currentLaneModel: (lane: LaneRole) => string;
  baseURL: string;
  effectiveTaskModel: string;
  useLocalDefault: boolean;
  runMultiLane: () => void;
  disabled: boolean;
};

export function DispatchWorkbench(props: DispatchWorkbenchProps) {
  const {
    t,
    splitMode,
    setSplitMode,
    dispatchMode,
    setDispatchMode,
    parallelism,
    setParallelism,
    laneOverrides,
    updateLaneOverride,
    laneLaunches,
    currentLaneModel,
    baseURL,
    effectiveTaskModel,
    runMultiLane,
    disabled,
  } = props;

  const lanes: Array<Exclude<LaneRole, 'primary'>> = ['planner', 'executor', 'reviewer', 'guard'];

  return (
    <div className="space-y-5">
      <Card className="rounded-[28px] border border-border/70 bg-background/75">
        <CardHeader>
          <CardTitle className="text-base">{t('dispatch.title')}</CardTitle>
          <CardDescription>{t('dispatch.desc')}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 xl:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)]">
          <div className="space-y-3">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-muted-foreground">{t('task.aiSplitMode')}</label>
              <Select value={splitMode} onValueChange={(value) => setSplitMode(value as SplitMode)}>
                <SelectTrigger aria-label={t('task.aiSplitMode')} className="rounded-2xl bg-background/80"><SelectValue /></SelectTrigger>
                <SelectContent>
                  {SPLIT_MODES.map((item) => <SelectItem key={item.key} value={item.key}>{t(item.label)}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-muted-foreground">{t('task.dispatchModeTitle')}</label>
              <Select value={dispatchMode} onValueChange={(value) => setDispatchMode(value as DispatchMode)}>
                <SelectTrigger aria-label={t('task.dispatchModeTitle')} className="rounded-2xl bg-background/80"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="parallel">{t('task.dispatchModes.parallel.label')}</SelectItem>
                  <SelectItem value="sequential">{t('task.dispatchModes.sequential.label')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-muted-foreground">{t('task.parallelismTitle')}</label>
              <Select value={String(parallelism)} onValueChange={(value) => setParallelism(Number(value))}>
                <SelectTrigger aria-label={t('task.parallelismTitle')} className="rounded-2xl bg-background/80"><SelectValue /></SelectTrigger>
                <SelectContent>
                  {[1, 2, 3, 4].map((count) => <SelectItem key={count} value={String(count)}>{t('task.parallelismOption', { count })}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <Button type="button" className="mt-2 rounded-2xl" onClick={runMultiLane} disabled={disabled}>{t('task.runMultiLane')}</Button>
          </div>

          <div className="grid gap-3 md:grid-cols-2">
            {lanes.map((lane) => {
              const laneOverride = laneOverrides[lane];
              return (
                <div key={lane} className="rounded-2xl border border-border/70 bg-background/80 p-4">
                  <div className="flex items-center justify-between gap-3">
                    <div>
                      <div className="text-sm font-medium text-foreground">{t(`task.laneLabels.${lane}`)}</div>
                      <div className="mt-1 text-xs leading-5 text-muted-foreground">{t(`task.lanePrompts.${lane}`)}</div>
                    </div>
                    <Switch aria-label={t('task.laneEndpointInherit')} checked={laneOverride.inheritMain} onCheckedChange={(value) => updateLaneOverride(lane, 'inheritMain', value)} />
                  </div>
                  <div className="mt-3 space-y-3">
                    <Input value={laneOverride.model} onChange={(event) => updateLaneOverride(lane, 'model', event.target.value)} placeholder={currentLaneModel(lane) || effectiveTaskModel} className="rounded-2xl" />
                    <Input value={laneOverride.baseURL} onChange={(event) => updateLaneOverride(lane, 'baseURL', event.target.value)} placeholder={baseURL} className="rounded-2xl" disabled={laneOverride.inheritMain} />
                    <div className="rounded-2xl border border-border/70 bg-muted/20 px-3 py-2 text-[11px] leading-5 text-muted-foreground">
                      {laneOverride.inheritMain ? t('task.laneEndpointInheritHint') : t('task.laneEndpointCustomHint')}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </CardContent>
      </Card>

      <Card className="rounded-[28px] border border-border/70 bg-background/75">
        <CardHeader>
          <CardTitle className="text-base">{t('task.laneLaunchTitle')}</CardTitle>
          <CardDescription>{t('task.laneLaunchDesc')}</CardDescription>
        </CardHeader>
        <CardContent className="max-h-[16rem] space-y-3 overflow-y-auto pr-1">
          {laneLaunches.length === 0 ? <EmptyLine text={t('task.laneLaunchEmpty')} /> : laneLaunches.map((item) => (
            <div key={`${item.lane}-${item.taskId}`} className="rounded-2xl border border-border/70 bg-background/80 p-3">
              <div className="flex items-center justify-between gap-3">
                <div className="text-sm font-medium text-foreground">{t(`task.laneLabels.${item.lane}`)} · #{item.taskId}</div>
                <div className="text-xs text-muted-foreground">{item.model}</div>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
