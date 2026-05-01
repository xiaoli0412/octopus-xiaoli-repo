'use client';

import { startTransition, useMemo, useState } from 'react';
import { Activity, Gauge, Search, ShieldCheck, Waves } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { toast } from '@/components/common/Toast';
import { HelpHint } from '@/components/common/HelpHint';
import { PROBE_POLICY_OPTIONS, useModelList, useUpdateModel, type LLMInfo, type ProbePolicy } from '@/api/endpoints/model';
import { getProbePolicyKey } from '@/lib/ui-labels';

type ProbeDraft = {
    probe_policy: ProbePolicy;
    probe_interval_seconds: string;
    probe_concurrency_limit: string;
};

const DEFAULT_VISIBLE_MODEL_COUNT = 12;

function buildDraft(model: LLMInfo): ProbeDraft {
    return {
        probe_policy: model.probe_policy ?? 'passive_only',
        probe_interval_seconds: String(model.probe_interval_seconds ?? 3600),
        probe_concurrency_limit: String(model.probe_concurrency_limit ?? 1),
    };
}

function ModelProbeRow({ model }: { model: LLMInfo }) {
    const t = useTranslations('setting.modelProbe');
    const tRouteTarget = useTranslations('setting.llmRouteTarget');
    const updateModel = useUpdateModel();
    const [draft, setDraft] = useState<ProbeDraft>(() => buildDraft(model));

    const syncedDraft = useMemo(() => buildDraft(model), [model]);

    const currentPolicy = model.probe_policy ?? 'passive_only';
    const currentInterval = String(model.probe_interval_seconds ?? 3600);
    const currentConcurrency = String(model.probe_concurrency_limit ?? 1);

    const effectiveDraft =
        draft.probe_policy === currentPolicy &&
        draft.probe_interval_seconds === currentInterval &&
        draft.probe_concurrency_limit === currentConcurrency
            ? syncedDraft
            : draft;

    const hasChanges =
        effectiveDraft.probe_policy !== currentPolicy ||
        effectiveDraft.probe_interval_seconds !== currentInterval ||
        effectiveDraft.probe_concurrency_limit !== currentConcurrency;

    const handleSave = () => {
        updateModel.mutate(
            {
                ...model,
                probe_policy: effectiveDraft.probe_policy,
                probe_interval_seconds: parseInt(effectiveDraft.probe_interval_seconds, 10) || 3600,
                probe_concurrency_limit: parseInt(effectiveDraft.probe_concurrency_limit, 10) || 1,
            },
            {
                onSuccess: () => {
                    toast.success(t('saveSuccess', { model: model.name }));
                },
                onError: (error) => {
                    toast.error(t('saveFailed'), { description: error.message });
                },
            }
        );
    };

    const summaryItems = [
        {
            label: t('summaryPolicy'),
            value: tRouteTarget(`probePolicyOptions.${getProbePolicyKey(model.probe_policy)}`),
        },
        {
            label: t('summaryInterval'),
            value: t('summaryIntervalValue', { seconds: model.probe_interval_seconds ?? 3600 }),
        },
        {
            label: t('summaryConcurrency'),
            value: model.probe_concurrency_limit ?? 1,
        },
    ];

    return (
        <AccordionItem value={model.name} className="rounded-2xl border border-border/60 bg-background/60 px-4">
            <AccordionTrigger className="py-4 text-left hover:no-underline">
                <div className="min-w-0 space-y-3">
                    <div className="truncate text-sm font-medium text-card-foreground">{model.name}</div>
                    <div className="flex flex-wrap gap-2 text-[11px] text-muted-foreground">
                        {model.canonical_name ? <span>{tRouteTarget('canonicalName')}: {model.canonical_name}</span> : null}
                        <span>{t('defaultBadge')}</span>
                    </div>
                    <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
                        {summaryItems.map((item) => (
                            <div key={item.label} className="rounded-xl border border-border/60 bg-muted/20 px-3 py-2">
                                <div className="text-[11px] text-muted-foreground">{item.label}</div>
                                <div className="mt-1 text-sm font-medium text-card-foreground">{item.value}</div>
                            </div>
                        ))}
                    </div>
                </div>
            </AccordionTrigger>
            <AccordionContent className="space-y-4 pb-4">
                <div className="flex items-center gap-2 rounded-xl border border-border/60 bg-muted/20 px-3 py-2 text-xs text-muted-foreground">
                    <span className="font-medium text-card-foreground">{t('advancedPanelTitle')}</span>
                    <HelpHint>{t('advancedPanelDesc')}</HelpHint>
                </div>

                <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
                    <div className="space-y-2">
                        <label className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                            {tRouteTarget('probePolicy')}
                            <HelpHint>{t('policyHint')}</HelpHint>
                        </label>
                        <Select
                            value={effectiveDraft.probe_policy}
                            onValueChange={(value) => setDraft((current) => ({ ...current, probe_policy: value as ProbePolicy }))}
                        >
                            <SelectTrigger className="rounded-xl">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent className="rounded-xl">
                                {PROBE_POLICY_OPTIONS.map((policy) => (
                                    <SelectItem key={policy} value={policy} className="rounded-xl">
                                        {tRouteTarget(`probePolicyOptions.${getProbePolicyKey(policy)}`)}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>

                    <div className="space-y-2">
                        <label className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                            {tRouteTarget('probeInterval')}
                            <HelpHint>{t('intervalHint')}</HelpHint>
                        </label>
                        <Input
                            type="number"
                            min={1}
                            value={effectiveDraft.probe_interval_seconds}
                            onChange={(e) => setDraft((current) => ({ ...current, probe_interval_seconds: e.target.value }))}
                            className="rounded-xl"
                        />
                    </div>

                    <div className="space-y-2">
                        <label className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                            {tRouteTarget('probeConcurrency')}
                            <HelpHint>{t('concurrencyHint')}</HelpHint>
                        </label>
                        <Input
                            type="number"
                            min={1}
                            value={effectiveDraft.probe_concurrency_limit}
                            onChange={(e) => setDraft((current) => ({ ...current, probe_concurrency_limit: e.target.value }))}
                            className="rounded-xl"
                        />
                    </div>
                </div>

                <div className="rounded-xl border border-border/60 bg-muted/20 px-3 py-2 text-xs leading-5 text-muted-foreground">
                    {t('routeOverrideHint')}
                </div>

                <div className="flex justify-end">
                    <Button
                        type="button"
                        size="sm"
                        onClick={handleSave}
                        disabled={!hasChanges || updateModel.isPending}
                        className="rounded-xl"
                    >
                        {updateModel.isPending ? t('saving') : t('save')}
                    </Button>
                </div>
            </AccordionContent>
        </AccordionItem>
    );
}

export function SettingModelProbe() {
    const t = useTranslations('setting.modelProbe');
    const [keyword, setKeyword] = useState('');
    const [showModelList, setShowModelList] = useState(false);
    const [visibleCount, setVisibleCount] = useState(DEFAULT_VISIBLE_MODEL_COUNT);
    const hasKeyword = keyword.trim().length > 0;
    const shouldFetchModels = showModelList || hasKeyword;
    const { data: models } = useModelList({ enabled: shouldFetchModels });

    const visibleModels = useMemo(() => {
        const term = keyword.trim().toLowerCase();
        return [...(models ?? [])]
            .sort((a, b) => a.name.localeCompare(b.name))
            .filter((model) => {
                if (!term) return true;
                return model.name.toLowerCase().includes(term) || (model.canonical_name ?? '').toLowerCase().includes(term);
            });
    }, [keyword, models]);

    const shouldRenderModelRows = showModelList || hasKeyword;
    const renderedModels = shouldRenderModelRows && !hasKeyword
        ? visibleModels.slice(0, visibleCount)
        : visibleModels;
    const hasMoreModels = shouldRenderModelRows && !hasKeyword && visibleModels.length > renderedModels.length;

    const summaryModels = visibleModels.slice(0, 3);
    const policySummary = summaryModels.map((model) => t(`policySummary.${getProbePolicyKey(model.probe_policy)}`));
    const uniquePolicies = [...new Set(policySummary)];

    const summaryCards = [
        {
            icon: ShieldCheck,
            label: t('summaryCards.modelCountLabel'),
            value: visibleModels.length,
            helper: t('summaryCards.modelCountHint'),
        },
        {
            icon: Activity,
            label: t('summaryCards.defaultPolicyLabel'),
            value: summaryModels.length === 0 ? '-' : uniquePolicies.length === 1 ? uniquePolicies[0] : t('summaryCards.mixedPoliciesValue'),
            helper: t('summaryCards.defaultPolicyHint'),
        },
        {
            icon: Gauge,
            label: t('summaryCards.intervalLabel'),
            value: summaryModels.length > 0
                ? t('summaryIntervalValue', { seconds: summaryModels[0]?.probe_interval_seconds ?? 3600 })
                : '-',
            helper: t('summaryCards.intervalHint'),
        },
        {
            icon: Waves,
            label: t('summaryCards.overrideLabel'),
            value: t('summaryCards.overrideValue'),
            helper: t('summaryCards.overrideHint'),
        },
    ];

    const handleKeywordChange = (value: string) => {
        setKeyword(value);
        startTransition(() => {
            setVisibleCount(DEFAULT_VISIBLE_MODEL_COUNT);
        });
    };

    const handleToggleModelList = () => {
        startTransition(() => {
            setShowModelList((current) => {
                const next = !current;
                if (!next) {
                    setVisibleCount(DEFAULT_VISIBLE_MODEL_COUNT);
                }
                return next;
            });
        });
    };

    const handleShowMore = () => {
        startTransition(() => {
            setVisibleCount((current) => current + DEFAULT_VISIBLE_MODEL_COUNT);
        });
    };

    return (
        <div data-testid="setting-model-probe-card" className="rounded-3xl border border-border bg-card p-6 space-y-5">
            <div className="space-y-2">
                <h2 className="flex items-center gap-2 text-lg font-bold text-card-foreground">
                    <Activity className="h-5 w-5" />
                    {t('title')}
                    <HelpHint>{t('hint')}</HelpHint>
                </h2>
                <p className="text-sm text-muted-foreground">{t('defaultPathDesc')}</p>
            </div>

            <div data-testid="setting-model-probe-default-path" className="rounded-2xl border border-border/60 bg-muted/20 px-4 py-3">
                <div className="flex flex-wrap items-start justify-between gap-3">
                    <div className="min-w-0 space-y-1">
                        <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                            <span>{t('defaultPathTitle')}</span>
                            <HelpHint>{t('defaultPathHint')}</HelpHint>
                        </div>
                        <p className="text-xs text-muted-foreground">{t('guidanceCompactDesc')}</p>
                    </div>
                    <span className="rounded-full border border-border/60 bg-background/80 px-2.5 py-1 text-[11px] font-medium text-muted-foreground">
                        {t('defaultBadge')}
                    </span>
                </div>
            </div>

            <div data-testid="setting-model-probe-search" className="relative">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                    value={keyword}
                    onChange={(e) => handleKeywordChange(e.target.value)}
                    placeholder={t('searchPlaceholder')}
                    className="rounded-xl pl-9"
                />
            </div>

            <div className="grid grid-cols-2 gap-3 xl:grid-cols-4">
                {summaryCards.map((card) => {
                    const Icon = card.icon;

                    return (
                        <div key={card.label} className="rounded-2xl border border-border/60 bg-muted/20 px-4 py-3">
                            <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                                <Icon className="h-3.5 w-3.5" />
                                <span>{card.label}</span>
                            </div>
                            <div className="mt-1 text-sm font-semibold text-card-foreground">{card.value}</div>
                        </div>
                    );
                })}
            </div>

            <div className="rounded-2xl border border-border/60 bg-muted/20 px-4 py-3">
                <div className="flex flex-wrap items-center justify-between gap-3">
                    <div className="min-w-0">
                        <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                            <span>{t('guidanceTitle')}</span>
                            <HelpHint>{t('guidanceHint')}</HelpHint>
                        </div>
                    </div>
                    <Button
                        data-testid="setting-model-probe-toggle"
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={handleToggleModelList}
                        className="rounded-xl"
                    >
                        {shouldRenderModelRows ? t('collapseModels') : t('showModels')}
                    </Button>
                </div>
            </div>

            {shouldRenderModelRows && visibleModels.length > 0 ? (
                <div className="space-y-3">
                    <div data-testid="setting-model-probe-scroll-region" className="max-h-[32rem] overflow-y-auto overscroll-contain pr-1">
                        <Accordion data-testid="setting-model-probe-model-list" type="multiple" className="space-y-3">
                            {renderedModels.map((model) => (
                                <ModelProbeRow key={model.name} model={model} />
                            ))}
                        </Accordion>
                    </div>
                    {hasMoreModels ? (
                        <div className="flex justify-center">
                            <Button data-testid="setting-model-probe-show-more" type="button" variant="outline" size="sm" onClick={handleShowMore} className="rounded-xl">
                                {t('showMoreModels', { count: Math.min(DEFAULT_VISIBLE_MODEL_COUNT, visibleModels.length - renderedModels.length) })}
                            </Button>
                        </div>
                    ) : null}
                </div>
            ) : !shouldRenderModelRows ? (
                <div data-testid="setting-model-probe-collapsed-state" className="rounded-2xl border border-dashed border-border px-4 py-6 text-sm text-muted-foreground">
                    {t('collapsedState')}
                </div>
            ) : (
                <div data-testid="setting-model-probe-empty-state" className="rounded-2xl border border-dashed border-border px-4 py-6 text-sm text-muted-foreground">
                    {t('empty')}
                </div>
            )}
        </div>
    );
}
