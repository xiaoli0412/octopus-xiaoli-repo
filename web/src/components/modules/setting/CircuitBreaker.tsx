'use client';

import { useMemo, useState, type ComponentType } from 'react';
import { useTranslations } from 'next-intl';
import { AlertTriangle, Hash, ShieldAlert, Timer, TimerOff, Waves } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { useSettingList, useSetSetting, SettingKey } from '@/api/endpoints/setting';
import { useStatsTokenBreakdown } from '@/api/endpoints/stats';
import { toast } from '@/components/common/Toast';
import { HelpHint } from '@/components/common/HelpHint';

type SettingField = {
    key: string;
    icon: ComponentType<{ className?: string }>;
    label: string;
    hint: string;
    placeholder: string;
    value: string;
    initialValue: string;
    onChange: (value: string) => void;
};

type TranslationFn = ReturnType<typeof useTranslations>;

function formatSeconds(seconds: number, t: TranslationFn) {
    if (seconds <= 0) return t('circuitBreaker.status.ready');
    return t('circuitBreaker.status.cooldownValue', { seconds });
}

export function SettingCircuitBreaker() {
    const t = useTranslations('setting');
    const { data: settings } = useSettingList();
    const { data: tokenBreakdown } = useStatsTokenBreakdown();
    const setSetting = useSetSetting();

    const thresholdSetting = settings?.find((s) => s.key === SettingKey.CircuitBreakerThreshold)?.value ?? '';
    const cooldownSetting = settings?.find((s) => s.key === SettingKey.CircuitBreakerCooldown)?.value ?? '';
    const maxCooldownSetting = settings?.find((s) => s.key === SettingKey.CircuitBreakerMaxCooldown)?.value ?? '';

    const [thresholdDraft, setThresholdDraft] = useState<string | null>(null);
    const [cooldownDraft, setCooldownDraft] = useState<string | null>(null);
    const [maxCooldownDraft, setMaxCooldownDraft] = useState<string | null>(null);

    const threshold = thresholdDraft ?? thresholdSetting;
    const cooldown = cooldownDraft ?? cooldownSetting;
    const maxCooldown = maxCooldownDraft ?? maxCooldownSetting;

    const handleSave = (key: string, value: string, initialValue: string) => {
        if (value === initialValue) return;

        setSetting.mutate({ key, value }, {
            onSuccess: () => {
                toast.success(t('saved'));
                if (key === SettingKey.CircuitBreakerThreshold) {
                    setThresholdDraft(null);
                } else if (key === SettingKey.CircuitBreakerCooldown) {
                    setCooldownDraft(null);
                } else if (key === SettingKey.CircuitBreakerMaxCooldown) {
                    setMaxCooldownDraft(null);
                }
            }
        });
    };

    const thresholdValue = Number.parseInt(threshold, 10);
    const cooldownValue = Number.parseInt(cooldown, 10);
    const maxCooldownValue = Number.parseInt(maxCooldown, 10);

    const normalizedThreshold = Number.isFinite(thresholdValue) && thresholdValue > 0 ? thresholdValue : 5;
    const normalizedCooldown = Number.isFinite(cooldownValue) && cooldownValue > 0 ? cooldownValue : 60;
    const normalizedMaxCooldown = Number.isFinite(maxCooldownValue) && maxCooldownValue > 0 ? maxCooldownValue : 600;

    const recommendation = useMemo(() => {
        if (normalizedThreshold <= 3 || normalizedCooldown <= 30 || normalizedMaxCooldown <= 180) {
            return {
                tone: 'warning' as const,
                title: t('circuitBreaker.recommendation.aggressiveTitle'),
                description: t('circuitBreaker.recommendation.aggressiveDesc'),
            };
        }
        if (normalizedThreshold >= 8 || normalizedMaxCooldown >= 1800) {
            return {
                tone: 'muted' as const,
                title: t('circuitBreaker.recommendation.relaxedTitle'),
                description: t('circuitBreaker.recommendation.relaxedDesc'),
            };
        }
        return {
            tone: 'balanced' as const,
            title: t('circuitBreaker.recommendation.recommendedTitle'),
            description: t('circuitBreaker.recommendation.recommendedDesc'),
        };
    }, [normalizedCooldown, normalizedMaxCooldown, normalizedThreshold, t]);

    const statusToneClass = recommendation.tone === 'warning'
        ? 'border-amber-500/30 bg-amber-500/8 text-amber-700 dark:text-amber-300'
        : recommendation.tone === 'balanced'
            ? 'border-emerald-500/25 bg-emerald-500/8 text-emerald-700 dark:text-emerald-300'
            : 'border-border/60 bg-muted/30 text-muted-foreground';

    const summaryCards = [
        {
            label: t('circuitBreaker.status.trackedLabel'),
            value: tokenBreakdown?.circuit_tracked_count ?? 0,
            helper: t('circuitBreaker.status.trackedHint'),
        },
        {
            label: t('circuitBreaker.status.openLabel'),
            value: tokenBreakdown?.circuit_open_count ?? 0,
            helper: t('circuitBreaker.status.openHint'),
        },
        {
            label: t('circuitBreaker.status.halfOpenLabel'),
            value: tokenBreakdown?.circuit_half_open_count ?? 0,
            helper: t('circuitBreaker.status.halfOpenHint'),
        },
        {
            label: t('circuitBreaker.status.cooldownLabel'),
            value: formatSeconds(tokenBreakdown?.circuit_max_remaining_cooldown_sec ?? 0, t),
            helper: t('circuitBreaker.status.cooldownHint'),
        },
    ];

    const recoverySteps = [
        {
            title: t('circuitBreaker.recoveryStep1Title'),
            description: t('circuitBreaker.recoveryStep1Desc'),
        },
        {
            title: t('circuitBreaker.recoveryStep2Title'),
            description: t('circuitBreaker.recoveryStep2Desc'),
        },
        {
            title: t('circuitBreaker.recoveryStep3Title'),
            description: t('circuitBreaker.recoveryStep3Desc'),
        },
    ];

    const advancedFields: SettingField[] = [
        {
            key: SettingKey.CircuitBreakerThreshold,
            icon: Hash,
            label: t('circuitBreaker.threshold.label'),
            hint: t('circuitBreaker.thresholdHint'),
            placeholder: t('circuitBreaker.threshold.placeholder'),
            value: threshold,
            initialValue: thresholdSetting,
            onChange: setThresholdDraft,
        },
        {
            key: SettingKey.CircuitBreakerCooldown,
            icon: Timer,
            label: t('circuitBreaker.cooldown.label'),
            hint: t('circuitBreaker.cooldownHint'),
            placeholder: t('circuitBreaker.cooldown.placeholder'),
            value: cooldown,
            initialValue: cooldownSetting,
            onChange: setCooldownDraft,
        },
        {
            key: SettingKey.CircuitBreakerMaxCooldown,
            icon: TimerOff,
            label: t('circuitBreaker.maxCooldown.label'),
            hint: t('circuitBreaker.maxCooldownHint'),
            placeholder: t('circuitBreaker.maxCooldown.placeholder'),
            value: maxCooldown,
            initialValue: maxCooldownSetting,
            onChange: setMaxCooldownDraft,
        },
    ];

    return (
        <div data-testid="setting-circuit-breaker-card" className="octo-setting-card">
            <div className="space-y-2">
                <h2 className="octo-setting-heading">
                    <ShieldAlert className="size-4" />
                    {t('circuitBreaker.title')}
                </h2>
                <span className="sr-only">
                    {t('circuitBreaker.defaultPathTitle')}
                    {t('circuitBreaker.defaultPathDesc')}
                    {t('circuitBreaker.advancedDesc')}
                </span>
            </div>

            <div className={`rounded-2xl border px-3 py-2.5 ${statusToneClass}`}>
                <div className="flex flex-wrap items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                        <AlertTriangle className="size-4 shrink-0" />
                        <div className="text-sm font-semibold">{recommendation.title}</div>
                    </div>
                    <p className="text-xs leading-5">{t('circuitBreaker.recommendation.currentValues', {
                            threshold: normalizedThreshold,
                            cooldown: normalizedCooldown,
                            maxCooldown: normalizedMaxCooldown,
                        })}</p>
                </div>
            </div>

            <div className="grid grid-cols-2 gap-2 xl:grid-cols-4">
                {summaryCards.map((card) => (
                    <div key={card.label} className="octo-stat-card">
                        <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                            <span>{card.label}</span>
                            <HelpHint className="size-3.5">{card.helper}</HelpHint>
                        </div>
                        <div className="mt-1 text-sm font-semibold text-card-foreground">{card.value}</div>
                    </div>
                ))}
            </div>

            <Accordion type="single" collapsible className="w-full rounded-2xl border border-border/60 bg-muted/20 px-3">
                <AccordionItem value="circuit-breaker-recovery" className="border-none">
                    <AccordionTrigger
                        data-testid="setting-circuit-breaker-recovery-trigger"
                        className="py-4 text-left hover:no-underline"
                    >
                        <div className="space-y-1">
                            <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                                <Waves className="h-4 w-4 text-muted-foreground" />
                                <span>{t('circuitBreaker.recoveryTitle')}</span>
                            </div>
                        </div>
                    </AccordionTrigger>
                    <AccordionContent className="space-y-3 border-t pb-4 pt-4">
                        {recoverySteps.map((step) => (
                            <div key={step.title} className="rounded-2xl border border-border/60 bg-background/70 px-4 py-3">
                                <div className="text-sm font-medium text-card-foreground">{step.title}</div>
                                <p className="mt-1 text-xs leading-5 text-muted-foreground">{step.description}</p>
                            </div>
                        ))}
                    </AccordionContent>
                </AccordionItem>
            </Accordion>

            <Accordion type="single" collapsible className="w-full rounded-2xl border border-border/60 bg-background/60 px-3">
                <AccordionItem value="circuit-breaker-advanced" className="border-none">
                    <AccordionTrigger
                        data-testid="setting-circuit-breaker-advanced-trigger"
                        className="py-4 text-left hover:no-underline"
                        addon={<HelpHint className="mt-1 size-3.5">{t('circuitBreaker.advancedHint')}</HelpHint>}
                    >
                        <div className="space-y-1">
                            <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                                <span>{t('circuitBreaker.advancedTitle')}</span>
                            </div>
                            <p className="text-xs leading-5 text-muted-foreground">{t('circuitBreaker.advancedDesc')}</p>
                        </div>
                    </AccordionTrigger>
                    <AccordionContent className="space-y-4 border-t pb-4 pt-4">
                        {advancedFields.map((field) => {
                            const Icon = field.icon;
                            return (
                                <div key={field.key} className="flex flex-col gap-3 rounded-2xl border border-border/50 bg-card px-4 py-4 md:flex-row md:items-center md:justify-between">
                                    <div className="space-y-1.5">
                                        <div className="flex items-center gap-3 text-sm font-medium text-card-foreground">
                                            <Icon className="h-4 w-4 text-muted-foreground" />
                                            <span>{field.label}</span>
                                            <HelpHint>{field.hint}</HelpHint>
                                        </div>
                                    </div>
                                    <Input
                                        aria-label={field.label}
                                        type="number"
                                        value={field.value}
                                        onChange={(e) => field.onChange(e.target.value)}
                                        onBlur={() => handleSave(field.key, field.value, field.initialValue)}
                                        placeholder={field.placeholder}
                                        className="w-full rounded-xl md:w-52"
                                    />
                                </div>
                            );
                        })}
                    </AccordionContent>
                </AccordionItem>
            </Accordion>
        </div>
    );
}
