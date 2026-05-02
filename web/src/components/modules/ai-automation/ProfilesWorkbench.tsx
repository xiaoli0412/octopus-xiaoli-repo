'use client';

import type { ReactNode } from 'react';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

import { EmptyLine } from './workbench-shared';

type TranslationFn = (key: string, values?: Record<string, any>) => string;

type ProfileListItem = {
  id: number;
  name: string;
  version: number;
  status: string;
  explanation: string;
};

type ProfilesWorkbenchProps = {
  t: TranslationFn;
  profiles: ProfileListItem[];
  selectedProfileID?: number;
  selectProfile: (id: number) => void;
  selectedProfilePanel?: ReactNode;
};

export function ProfilesWorkbench(props: ProfilesWorkbenchProps) {
  const { t, profiles, selectedProfileID, selectProfile, selectedProfilePanel } = props;

  return (
    <Card className="rounded-[28px] border border-border/70 bg-background/75">
      <CardHeader>
        <CardTitle className="text-base">{t('profiles.title')}</CardTitle>
        <CardDescription>{t('profiles.desc')}</CardDescription>
      </CardHeader>
      <CardContent className="grid gap-4 xl:grid-cols-[18rem_minmax(0,1fr)]">
        <div className="max-h-[30rem] space-y-3 overflow-y-auto pr-1">
          {profiles.length === 0 ? <EmptyLine text={t('profiles.empty')} /> : profiles.map((profile) => (
            <button key={profile.id} type="button" className={`w-full rounded-2xl border p-3 text-left transition ${selectedProfileID === profile.id ? 'border-primary/60 bg-primary/5' : 'border-border/70 bg-background/80 hover:bg-muted/40'}`} onClick={() => selectProfile(profile.id)}>
              <div className="flex items-center justify-between gap-2">
                <div className="min-w-0 text-sm font-medium text-foreground">{profile.name}</div>
                <div className="rounded-full border border-border/70 px-2 py-0.5 text-[11px] text-muted-foreground">v{profile.version}</div>
              </div>
              <div className="mt-2 text-xs leading-5 text-muted-foreground">{profile.explanation || t('profiles.noActiveHint')}</div>
            </button>
          ))}
        </div>
        <div>{selectedProfilePanel}</div>
      </CardContent>
    </Card>
  );
}
