'use client';

import type { AIPromptTemplate } from '@/api/endpoints/ai-automation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { formatDateTimeByLocale } from '@/lib/locale';
import { type Locale as AppLocale } from '@/stores/setting';

import { EmptyLine, Snapshot, SubtleAction, ToolKey, TOOL_OPTIONS } from './workbench-shared';

type TranslationFn = (key: string, values?: Record<string, any>) => string;

type AssetsWorkbenchProps = {
  t: TranslationFn;
  templates: AIPromptTemplate[];
  selectedTemplateIDs: number[];
  toggleTemplate: (id: number) => void;
  newTemplateName: string;
  setNewTemplateName: (value: string) => void;
  newTemplatePrompt: string;
  setNewTemplatePrompt: (value: string) => void;
  newTemplateRequirement: string;
  setNewTemplateRequirement: (value: string) => void;
  createTemplate: () => void;
  toolKeys: ToolKey[];
  toggleTool: (key: ToolKey) => void;
  snapshots: Snapshot[];
  snapshotLabel: string;
  setSnapshotLabel: (value: string) => void;
  createManualSnapshot: () => void;
  restoreSnapshot: (snapshot?: Snapshot) => void;
  clearSnapshots: () => void;
  locale: AppLocale;
};

export function AssetsWorkbench(props: AssetsWorkbenchProps) {
  const {
    t,
    templates,
    selectedTemplateIDs,
    toggleTemplate,
    newTemplateName,
    setNewTemplateName,
    newTemplatePrompt,
    setNewTemplatePrompt,
    newTemplateRequirement,
    setNewTemplateRequirement,
    createTemplate,
    toolKeys,
    toggleTool,
    snapshots,
    snapshotLabel,
    setSnapshotLabel,
    createManualSnapshot,
    restoreSnapshot,
    clearSnapshots,
    locale,
  } = props;

  return (
    <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
      <Card className="rounded-[28px] border border-border/70 bg-background/75">
        <CardHeader>
          <CardTitle className="text-base">{t('task.promptTemplatesTitle')}</CardTitle>
          <CardDescription>{t('assets.templatesDesc')}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
          <div className="max-h-[24rem] space-y-3 overflow-y-auto pr-1">
            {templates.length === 0 ? <EmptyLine text={t('task.snapshotEmpty')} /> : templates.map((item) => (
              <label key={item.id} className="flex items-start gap-3 rounded-2xl border border-border/70 bg-background/80 p-3">
                <input type="checkbox" checked={selectedTemplateIDs.includes(item.id)} onChange={() => toggleTemplate(item.id)} />
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
            <Button type="button" className="rounded-2xl" onClick={createTemplate}>{t('task.createTemplate')}</Button>
          </div>
        </CardContent>
      </Card>

      <div className="space-y-5">
        <Card className="rounded-[28px] border border-border/70 bg-background/75">
          <CardHeader>
            <CardTitle className="text-base">{t('task.toolCapabilityTitle')}</CardTitle>
            <CardDescription>{t('task.toolCapabilityDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-3 md:grid-cols-2">
            {TOOL_OPTIONS.map((tool) => (
              <button key={tool.key} type="button" className={`rounded-2xl border p-3 text-left transition ${toolKeys.includes(tool.key) ? 'border-primary/60 bg-primary/5' : 'border-border/70 bg-background/80 hover:bg-muted/40'}`} onClick={() => toggleTool(tool.key)}>
                <div className="flex items-center justify-between gap-2">
                  <div className="text-sm font-medium text-foreground">{t(tool.label)}</div>
                  <div className="rounded-full border border-border/70 px-2 py-0.5 text-[11px] text-muted-foreground">{toolKeys.includes(tool.key) ? t('task.toolActive') : t('task.toolInactive')}</div>
                </div>
              </button>
            ))}
          </CardContent>
        </Card>

        <Card className="rounded-[28px] border border-border/70 bg-background/75">
          <CardHeader>
            <CardTitle className="text-base">{t('task.snapshotPanelTitle')}</CardTitle>
            <CardDescription>{t('task.snapshotPanelDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Input value={snapshotLabel} onChange={(event) => setSnapshotLabel(event.target.value)} placeholder={t('task.snapshotLabelPlaceholder')} className="rounded-2xl" />
            <div className="flex flex-wrap gap-2">
              <Button type="button" className="rounded-2xl" onClick={createManualSnapshot}>{t('task.snapshotCreate')}</Button>
              <SubtleAction type="button" onClick={() => restoreSnapshot(snapshots[0])} disabled={snapshots.length === 0}>{t('task.snapshotRestore')}</SubtleAction>
              <SubtleAction type="button" onClick={clearSnapshots} disabled={snapshots.length === 0}>{t('task.snapshotClear')}</SubtleAction>
            </div>
            <div className="max-h-[14rem] space-y-3 overflow-y-auto pr-1">
              {snapshots.length === 0 ? <EmptyLine text={t('task.snapshotEmpty')} /> : snapshots.map((snapshot) => (
                <button key={snapshot.id} type="button" className="w-full rounded-2xl border border-border/70 bg-background/80 p-3 text-left transition hover:bg-muted/40" onClick={() => restoreSnapshot(snapshot)}>
                  <div className="flex items-center justify-between gap-3">
                    <div className="min-w-0 text-sm font-medium text-foreground">{snapshot.label}</div>
                  </div>
                  <div className="mt-1 text-xs text-muted-foreground">{formatDateTimeByLocale(snapshot.created_at, locale)}</div>
                </button>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
