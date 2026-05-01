'use client';

import { useEffect, useMemo, useState } from 'react';
import { useTranslations } from 'next-intl';
import { Bot, ExternalLink, ShieldAlert, ShieldCheck } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useSettingList, useSetSetting, SettingKey } from '@/api/endpoints/setting';
import { useAIAutomationConfig, useActivateAIProfile, useAIProfiles } from '@/api/endpoints/ai-automation';
import { useNavStore } from '@/components/modules/navbar';
import { toast } from '@/components/common/Toast';
import {
    formatConfigProfileLabel,
    resolveProfileSummaryState,
    resolveSelectedProfile,
} from '@/components/modules/ai-automation/config-source-logic';

const CONFIG_SOURCE_MANUAL = 'manual';
const CONFIG_SOURCE_AI_PROFILE = 'ai_profile';
const LOW_CONFIDENCE_THRESHOLD = 0.7;

function formatProfileUpdatedAt(value?: string) {
    if (!value) return '-';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleString();
}

function formatProfileConfidence(value?: number) {
    return `${Math.round((value ?? 0) * 100)}%`;
}

function isProfileActivationBlocked(status?: string) {
    return status === 'invalid' || status === 'archived';
}

function getProfileStatusBadgeClassName(status?: string) {
    switch (status) {
        case 'active':
            return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300';
        case 'ready':
            return 'border-sky-500/30 bg-sky-500/10 text-sky-700 dark:text-sky-300';
        case 'draft':
            return 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300';
        case 'archived':
            return 'border-slate-500/30 bg-slate-500/10 text-slate-700 dark:text-slate-300';
        case 'invalid':
            return 'border-destructive/30 bg-destructive/10 text-destructive';
        default:
            return 'border-border/60 bg-background/75 text-muted-foreground';
    }
}

export function SettingAIAutomationSource() {
    const t = useTranslations('setting');
    const { data: settings } = useSettingList();
    const configQuery = useAIAutomationConfig();
    const profilesQuery = useAIProfiles();
    const setSetting = useSetSetting();
    const activateProfile = useActivateAIProfile();
    const setActiveItem = useNavStore((state) => state.setActiveItem);
    const [selectedProfileID, setSelectedProfileID] = useState<number>();

    const persistedSourceMode = settings?.find((setting) => setting.key === SettingKey.ConfigSourceMode)?.value || CONFIG_SOURCE_MANUAL;
    const persistedActiveProfileID = settings?.find((setting) => setting.key === SettingKey.ActiveAIProfileID)?.value || '0';
    const profiles = profilesQuery.data ?? [];
    const numericPersistedActiveProfileID = Number(persistedActiveProfileID);
    const requestedSourceMode = configQuery.data?.requested_config_source_mode ?? persistedSourceMode;
    const resolvedSourceMode = configQuery.data?.config_source_mode ?? persistedSourceMode;
    const requestedActiveProfileID = configQuery.data?.requested_active_ai_profile_id ?? numericPersistedActiveProfileID;
    const resolvedActiveProfileID = configQuery.data?.active_ai_profile_id ?? numericPersistedActiveProfileID;
    const requestedProfileSummary = configQuery.data?.requested_active_ai_profile;
    const activeProfileSummary = configQuery.data?.active_ai_profile;
    const sourceFallbackReason = configQuery.data?.source_fallback_reason ?? '';
    const { activeProfile, requestedProfile, persistedProfile, resolvedProfile } = useMemo(
        () => resolveProfileSummaryState({
            profiles,
            requestedActiveProfileID,
            resolvedActiveProfileID,
            persistedActiveProfileID: numericPersistedActiveProfileID,
            requestedProfileSummary,
            activeProfileSummary,
        }),
        [
            activeProfileSummary,
            numericPersistedActiveProfileID,
            profiles,
            requestedActiveProfileID,
            requestedProfileSummary,
            resolvedActiveProfileID,
        ]
    );

    useEffect(() => {
        if (profiles.length === 0) {
            if (selectedProfileID !== undefined) {
                setSelectedProfileID(undefined);
            }
            return;
        }
        if (selectedProfileID && profiles.some((profile) => profile.id === selectedProfileID)) {
            return;
        }
        setSelectedProfileID((resolvedProfile?.id ?? activeProfile?.id ?? requestedProfile?.id ?? profiles[0].id));
    }, [activeProfile?.id, profiles, requestedProfile?.id, resolvedProfile?.id, selectedProfileID]);

    const selectedProfile = useMemo(
        () => resolveSelectedProfile({
            profiles,
            selectedProfileID,
            requestedProfile,
            resolvedProfile,
            activeProfile,
            persistedProfile,
        }),
        [activeProfile, persistedProfile, profiles, requestedProfile, resolvedProfile, selectedProfileID]
    );
    const hasActiveProfile = !!resolvedProfile;
    const selectedProfileIsActive = !!selectedProfile && selectedProfile.id === resolvedActiveProfileID;
    const selectedProfileStatus = selectedProfile?.status ?? 'unknown';
    const selectedProfileActivationBlocked = isProfileActivationBlocked(selectedProfileStatus);
    const runtimeFallbackNoticeVisible = requestedSourceMode === CONFIG_SOURCE_AI_PROFILE && resolvedSourceMode === CONFIG_SOURCE_MANUAL;

    const profileRiskHints = useMemo(() => {
        if (!selectedProfile) return [];

        const hints: string[] = [];
        if (selectedProfileStatus === 'invalid') {
            hints.push(t('aiAutomationSource.riskInvalid'));
        }
        if (selectedProfileStatus === 'archived') {
            hints.push(t('aiAutomationSource.riskArchived'));
        }
        if (selectedProfileStatus === 'draft') {
            hints.push(t('aiAutomationSource.riskDraft'));
        }
        if ((selectedProfile.confidence ?? 0) < LOW_CONFIDENCE_THRESHOLD) {
            hints.push(t('aiAutomationSource.riskLowConfidence', { confidence: formatProfileConfidence(selectedProfile.confidence) }));
        }
        return hints;
    }, [selectedProfile, selectedProfileStatus, t]);

    const saveSourceMode = (value: string) => {
        setSetting.mutate(
            { key: SettingKey.ConfigSourceMode, value },
            { onSuccess: () => toast.success(t('saved')) }
        );
    };

    const openAICenter = () => setActiveItem('ai');

    const applySelectedProfile = async () => {
        if (!selectedProfile) {
            toast.info(t('aiAutomationSource.selectProfileFirst'));
            openAICenter();
            return;
        }

        if (selectedProfileActivationBlocked) {
            toast.info(t('aiAutomationSource.profileBlocked'));
            openAICenter();
            return;
        }

        if (selectedProfileIsActive) {
            if (resolvedSourceMode !== CONFIG_SOURCE_AI_PROFILE) {
                saveSourceMode(CONFIG_SOURCE_AI_PROFILE);
            }
            return;
        }

        try {
            await activateProfile.mutateAsync(selectedProfile.id);
            toast.success(t('saved'));
        } catch (error) {
            toast.error(t('aiAutomationSource.activateFailed'), {
                description: error instanceof Error ? error.message : undefined,
            });
        }
    };

    const handleSourceChange = async (value: string) => {
        if (value === CONFIG_SOURCE_MANUAL) {
            saveSourceMode(value);
            return;
        }

        if (!selectedProfile && !hasActiveProfile) {
            toast.info(t('aiAutomationSource.selectProfileFirst'));
            openAICenter();
            return;
        }

        await applySelectedProfile();
    };

    return (
        <div className="octo-setting-card">
            <div className="flex items-start gap-3">
                <div className="rounded-2xl bg-primary/10 p-2 text-primary">
                    <Bot className="size-5" />
                </div>
                <div className="min-w-0 flex-1">
                    <h3 className="text-base font-semibold text-card-foreground">{t('aiAutomationSource.title')}</h3>
                    <p className="mt-1 text-sm leading-6 text-muted-foreground">{t('aiAutomationSource.desc')}</p>
                </div>
            </div>

            <div className="octo-setting-row">
                <div className="octo-setting-label">{t('aiAutomationSource.mode')}</div>
                <div className="space-y-2">
                    <Select value={resolvedSourceMode} onValueChange={handleSourceChange}>
                        <SelectTrigger aria-label={t('aiAutomationSource.mode')} className="h-10 w-full rounded-xl bg-background md:max-w-sm">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value={CONFIG_SOURCE_MANUAL}>{t('aiAutomationSource.manual')}</SelectItem>
                            <SelectItem value={CONFIG_SOURCE_AI_PROFILE}>{t('aiAutomationSource.aiProfile')}</SelectItem>
                        </SelectContent>
                    </Select>
                    <p className="text-xs leading-5 text-muted-foreground">
                        {hasActiveProfile
                                ? t('aiAutomationSource.activeProfile', { id: formatConfigProfileLabel(requestedSourceMode === CONFIG_SOURCE_AI_PROFILE ? requestedProfile : resolvedProfile) || String(requestedSourceMode === CONFIG_SOURCE_AI_PROFILE ? requestedActiveProfileID : resolvedActiveProfileID) })
                            : t('aiAutomationSource.noActiveProfile')}
                    </p>
                    {runtimeFallbackNoticeVisible ? (
                        <div className="rounded-2xl border border-amber-500/30 bg-amber-500/10 p-3 text-xs leading-5 text-amber-900 dark:text-amber-100">
                            <div className="flex items-start gap-2">
                                <ShieldAlert className="mt-0.5 size-4 shrink-0" />
                                <div className="space-y-1">
                                    <div className="font-medium">{t('aiAutomationSource.runtimeFallbackTitle')}</div>
                                    <div>{t('aiAutomationSource.runtimeFallbackNotice')}</div>
                                    {sourceFallbackReason ? <div>{t(`aiAutomationSource.fallbackReasons.${sourceFallbackReason}`)}</div> : null}
                                </div>
                            </div>
                        </div>
                    ) : null}
                </div>
            </div>

            <div className="octo-setting-row">
                <div className="octo-setting-label">{t('aiAutomationSource.profileLabel')}</div>
                <div className="space-y-2.5">
                    {profiles.length > 0 ? (
                        <>
                            <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
                                <Select value={selectedProfile ? String(selectedProfile.id) : undefined} onValueChange={(value) => setSelectedProfileID(Number(value))}>
                                    <SelectTrigger aria-label={t('aiAutomationSource.profileLabel')} className="h-10 w-full rounded-xl bg-background md:max-w-sm">
                                        <SelectValue placeholder={t('aiAutomationSource.profilePlaceholder')} />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {profiles.map((profile) => (
                                            <SelectItem key={profile.id} value={String(profile.id)}>{profile.name}</SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                                <Button
                                    type="button"
                                    variant="outline"
                                    className="h-10 rounded-xl sm:min-w-40"
                                    onClick={applySelectedProfile}
                                    disabled={!selectedProfile || selectedProfileActivationBlocked || activateProfile.isPending || (selectedProfileIsActive && resolvedSourceMode === CONFIG_SOURCE_AI_PROFILE)}
                                >
                                    {activateProfile.isPending ? t('aiAutomationSource.usingSelected') : t('aiAutomationSource.useSelected')}
                                </Button>
                            </div>
                            <p className="text-xs leading-5 text-muted-foreground">
                                {selectedProfileIsActive
                                    ? t('aiAutomationSource.activeProfile', { id: formatConfigProfileLabel(requestedSourceMode === CONFIG_SOURCE_AI_PROFILE ? requestedProfile : resolvedProfile) || String(requestedSourceMode === CONFIG_SOURCE_AI_PROFILE ? requestedActiveProfileID : resolvedActiveProfileID) })
                                    : t('aiAutomationSource.selectedProfilePending')}
                            </p>
                        </>
                    ) : (
                        <p className="text-xs leading-5 text-muted-foreground">{t('aiAutomationSource.profileEmpty')}</p>
                    )}
                </div>
            </div>

            {selectedProfile ? (
                <div className="rounded-2xl border border-border/70 bg-muted/20 p-3">
                    <div className="text-sm font-medium text-card-foreground">{t('aiAutomationSource.selectedProfile')}</div>
                    <div className="mt-2 flex flex-wrap items-center gap-2">
                        <div className="text-base font-semibold text-card-foreground">{selectedProfile.name}</div>
                        <Badge variant="outline" className={getProfileStatusBadgeClassName(selectedProfileStatus)}>
                            {t(`aiAutomationSource.statusValues.${selectedProfileStatus}`)}
                        </Badge>
                        {selectedProfileIsActive && resolvedSourceMode === CONFIG_SOURCE_AI_PROFILE ? (
                            <Badge variant="outline" className="border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300">
                                {t('aiAutomationSource.inUseBadge')}
                            </Badge>
                        ) : null}
                    </div>
                    <div className="mt-2 grid gap-2 sm:grid-cols-2 xl:grid-cols-5">
                        <div className="rounded-xl border border-border/60 bg-background/75 px-3 py-2 text-xs">
                            <div className="text-muted-foreground">{t('aiAutomationSource.profileId')}</div>
                            <div className="mt-1 font-medium text-card-foreground">#{selectedProfile.id}</div>
                        </div>
                        <div className="rounded-xl border border-border/60 bg-background/75 px-3 py-2 text-xs">
                            <div className="text-muted-foreground">{t('aiAutomationSource.profileVersion')}</div>
                            <div className="mt-1 font-medium text-card-foreground">v{selectedProfile.version}</div>
                        </div>
                        <div className="rounded-xl border border-border/60 bg-background/75 px-3 py-2 text-xs">
                            <div className="text-muted-foreground">{t('aiAutomationSource.profileConfidence')}</div>
                            <div className="mt-1 font-medium text-card-foreground">{formatProfileConfidence(selectedProfile.confidence)}</div>
                        </div>
                        <div className="rounded-xl border border-border/60 bg-background/75 px-3 py-2 text-xs">
                            <div className="text-muted-foreground">{t('aiAutomationSource.profileUpdatedAt')}</div>
                            <div className="mt-1 font-medium text-card-foreground">{formatProfileUpdatedAt(selectedProfile.updated_at)}</div>
                        </div>
                        <div className="rounded-xl border border-border/60 bg-background/75 px-3 py-2 text-xs">
                            <div className="text-muted-foreground">{t('aiAutomationSource.status')}</div>
                            <div className="mt-1 font-medium text-card-foreground">{t(`aiAutomationSource.statusValues.${selectedProfileStatus}`)}</div>
                        </div>
                    </div>
                    <p className="mt-2 text-xs leading-5 text-muted-foreground">
                        {selectedProfile.explanation || t('aiAutomationSource.profileSummaryFallback')}
                    </p>
                    {profileRiskHints.length > 0 ? (
                        <div className="mt-3 rounded-2xl border border-amber-500/30 bg-amber-500/10 p-3 text-xs leading-5 text-amber-900 dark:text-amber-100">
                            <div className="flex items-start gap-2">
                                <ShieldAlert className="mt-0.5 size-4 shrink-0" />
                                <div className="space-y-1">
                                    <div className="font-medium">{t('aiAutomationSource.riskTitle')}</div>
                                    {profileRiskHints.map((hint) => (
                                        <div key={hint}>{hint}</div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    ) : (
                        <div className="mt-3 rounded-2xl border border-emerald-500/30 bg-emerald-500/10 p-3 text-xs leading-5 text-emerald-900 dark:text-emerald-100">
                            <div className="flex items-start gap-2">
                                <ShieldCheck className="mt-0.5 size-4 shrink-0" />
                                <div>{t('aiAutomationSource.readyHint')}</div>
                            </div>
                        </div>
                    )}
                    <div className="mt-2 grid gap-1 text-xs leading-5 text-muted-foreground">
                        <div>{t('aiAutomationSource.manualSafety')}</div>
                        <div>{t('aiAutomationSource.fallbackHint')}</div>
                    </div>
                </div>
            ) : (
                <div className="rounded-2xl border border-dashed border-border/70 bg-muted/20 p-3 text-xs leading-5 text-muted-foreground">
                    <div>{t('aiAutomationSource.noActiveProfile')}</div>
                    <div className="mt-1">{t('aiAutomationSource.manualSafety')}</div>
                </div>
            )}

            <div className="flex justify-start sm:justify-end">
                <Button type="button" variant="outline" className="h-10 w-full rounded-xl sm:w-auto sm:min-w-40" onClick={openAICenter}>
                    <ExternalLink className="size-4" />
                    {t('aiAutomationSource.openCenter')}
                </Button>
            </div>
        </div>
    );
}
