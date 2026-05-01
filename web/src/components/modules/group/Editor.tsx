'use client';

import { useCallback, useMemo, useState, type FormEvent } from 'react';
import { Check, Plus, Sparkles } from 'lucide-react';
import { useTranslations } from 'next-intl';
import * as AccordionPrimitive from '@radix-ui/react-accordion';
import { useModelChannelList, type LLMChannel } from '@/api/endpoints/model';
import { Button } from '@/components/ui/button';
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { HelpHint } from '@/components/common/HelpHint';
import { cn } from '@/lib/utils';
import { getModelIcon } from '@/lib/model-icons';
import type { GroupMode } from '@/api/endpoints/group';
import type { SelectedMember } from './ItemList';
import { MemberList } from './ItemList';
import { matchesGroupName, memberKey, normalizeKey, MODE_LABELS } from './utils';

type GroupEditorCopy = {
    flowTitle: string;
    flowHint: string;
    flowDesc: string;
    stepLabel: (index: number) => string;
    flowSteps: {
        naming: string;
        mode: string;
        models: string;
    };
    namingSectionTitle: string;
    namingSectionHint: string;
    namingSectionDesc: string;
    modeSectionTitle: string;
    modeSectionHint: string;
    modeSectionDesc: string;
    nameHint: string;
    matchRegexHint: string;
    advancedStrategyHint: string;
    modelPickerTitle: string;
    modelPickerHint: string;
    noAvailableModelsTitle: string;
    noAvailableModelsHint: string;
    noFilteredModelsHint: string;
    itemsEmptyTitle: string;
    itemsEmptyHint: string;
    selectionSummaryTitle: string;
    selectionSummaryModels: (count: number) => string;
    selectionSummaryChannels: (count: number) => string;
    selectionSummaryWeighted: (count: number) => string;
    selectionSummaryEmpty: string;
    selectionSummaryHint: string;
    weightedModeActive: string;
    weightedModeInactive: string;
    retryRounds: string;
    retryRoundsHint: string;
    retryDelayMs: string;
    retryDelayMsHint: string;
    failoverWindowSec: string;
    failoverWindowSecHint: string;
    raceAfterFails: string;
    raceAfterFailsHint: string;
    raceConcurrency: string;
    raceConcurrencyHint: string;
    advancedStrategy: string;
    advancedStrategyDesc: string;
};

export type GroupEditorValues = {
    name: string;
    match_regex: string;
    mode: GroupMode;
    first_token_time_out: number;
    session_keep_time: number;
    retry_rounds: number;
    retry_delay_ms: number;
    failover_window_sec: number;
    race_after_fails: number;
    race_concurrency: number;
    members: SelectedMember[];
};

function parsePositiveInt(raw: string) {
    if (raw.trim() === '') {
        return 0;
    }
    const parsed = Number.parseInt(raw, 10);
    return Number.isFinite(parsed) && parsed > 0 ? parsed : 0;
}

function ModelPickerSection({
    modelChannels,
    selectedMembers,
    onAdd,
    onAutoAdd,
    autoAddDisabled,
    idPrefix,
    copy,
    modelFilter,
    onModelFilterChange,
}: {
    modelChannels: LLMChannel[];
    selectedMembers: SelectedMember[];
    onAdd: (channel: LLMChannel) => void;
    onAutoAdd: () => void;
    autoAddDisabled: boolean;
    idPrefix: string;
    copy: GroupEditorCopy;
    modelFilter: string;
    onModelFilterChange: (value: string) => void;
}) {
    const t = useTranslations('group');
    const selectedKeys = useMemo(() => new Set(selectedMembers.map(memberKey)), [selectedMembers]);
    const hasAnyModelChannels = modelChannels.length > 0;

    const filteredModelChannels = useMemo(() => {
        const keyword = modelFilter.trim().toLowerCase();
        if (!keyword) {
            return modelChannels;
        }
        return modelChannels.filter((mc) => {
            const channelName = (mc.channel_name || t('channelFallbackName', { id: mc.channel_id })).toLowerCase();
            return mc.name.toLowerCase().includes(keyword) || channelName.includes(keyword);
        });
    }, [modelChannels, modelFilter, t]);

    const channels = useMemo(() => {
        const byId = new Map<number, { id: number; name: string; models: LLMChannel[] }>();
        filteredModelChannels.forEach((mc) => {
            const existing = byId.get(mc.channel_id);
            if (existing) {
                existing.models.push(mc);
            } else {
                byId.set(mc.channel_id, {
                    id: mc.channel_id,
                    name: mc.channel_name || t('channelFallbackName', { id: mc.channel_id }),
                    models: [mc],
                });
            }
        });

        return Array.from(byId.values())
            .map((channel) => ({
                ...channel,
                models: [...channel.models].sort((a, b) => a.name.localeCompare(b.name)),
            }))
            .sort((a, b) => a.name.localeCompare(b.name));
    }, [filteredModelChannels, t]);

    const emptyTitle = hasAnyModelChannels ? t('form.noFilteredModels') : copy.noAvailableModelsTitle;
    const emptyHint = hasAnyModelChannels ? copy.noFilteredModelsHint : copy.noAvailableModelsHint;

    return (
        <section data-testid={`${idPrefix}-model-picker-section`} className="rounded-2xl border border-border/50 bg-muted/30 p-4">
            <div className="flex items-start justify-between gap-3">
                <div className="space-y-1">
                    <span className="truncate text-sm font-medium text-foreground">{copy.modelPickerTitle}</span>
                    <p className="text-xs text-muted-foreground">{copy.modelPickerHint}</p>
                </div>
                <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    data-testid={`${idPrefix}-auto-add`}
                    onClick={onAutoAdd}
                    disabled={autoAddDisabled}
                    className="shrink-0 rounded-xl"
                >
                    <Sparkles className="size-3.5" />
                    <span>{t('form.autoAdd')}</span>
                </Button>
            </div>

            <div className="mt-4 space-y-3">
                <Input
                    value={modelFilter}
                    onChange={(event) => onModelFilterChange(event.target.value)}
                    placeholder={t('form.modelFilterPlaceholder')}
                    data-testid={`${idPrefix}-model-filter`}
                    className="rounded-xl"
                />

                {channels.length > 0 ? (
                    <div className="grid max-h-56 grid-cols-1 gap-3 overflow-y-auto pr-1 md:grid-cols-2 xl:grid-cols-4">
                        {channels.map((channel) => {
                            const total = channel.models.length;
                            const selectedCount = channel.models.reduce(
                                (count, model) => count + (selectedKeys.has(memberKey(model)) ? 1 : 0),
                                0,
                            );

                            return (
                                <Accordion key={channel.id} type="single" collapsible className="rounded-xl border border-border/50 bg-background/80 px-3">
                                    <AccordionItem value={`channel-${channel.id}`} className="border-none">
                                        <AccordionTrigger className="py-3 text-sm font-medium">
                                            <div className="min-w-0 space-y-1">
                                                <div className="truncate text-sm font-medium text-foreground">{channel.name}</div>
                                                <div className="text-xs text-muted-foreground">{selectedCount}/{total}</div>
                                            </div>
                                        </AccordionTrigger>
                                        <AccordionContent>
                                            <div className="space-y-2">
                                                {channel.models.map((model) => {
                                                    const selected = selectedKeys.has(memberKey(model));
                                                    const { Avatar } = getModelIcon(model.name);
                                                    return (
                                                        <button
                                                            key={memberKey(model)}
                                                            type="button"
                                                            onClick={() => {
                                                                if (!selected) {
                                                                    onAdd(model);
                                                                }
                                                            }}
                                                            disabled={selected}
                                                            className={cn(
                                                                'flex w-full items-center justify-between gap-2 rounded-xl border border-border/50 px-2.5 py-2 text-left transition-colors',
                                                                selected ? 'cursor-not-allowed bg-muted/60 opacity-60' : 'bg-background hover:bg-muted',
                                                            )}
                                                        >
                                                            <span className="flex min-w-0 items-center gap-2">
                                                                <Avatar size={16} />
                                                                <span className="truncate text-sm font-medium">{model.name}</span>
                                                            </span>
                                                            {selected ? <Check className="size-4 text-primary" /> : <Plus className="size-4 text-muted-foreground" />}
                                                        </button>
                                                    );
                                                })}
                                            </div>
                                        </AccordionContent>
                                    </AccordionItem>
                                </Accordion>
                            );
                        })}
                    </div>
                ) : (
                    <div data-testid={`${idPrefix}-model-empty-state`} className="rounded-xl border border-dashed border-border/60 bg-background/70 p-4">
                        <div className="font-medium text-foreground">{emptyTitle}</div>
                        <p className="mt-1 text-xs text-muted-foreground">{emptyHint}</p>
                    </div>
                )}
            </div>
        </section>
    );
}

function SortSection({
    members,
    onReorder,
    onRemove,
    onWeightChange,
    removingIds,
    showWeight,
    onClear,
    idPrefix,
    copy,
}: {
    members: SelectedMember[];
    onReorder: (members: SelectedMember[]) => void;
    onRemove: (id: string) => void;
    onWeightChange: (id: string, weight: number) => void;
    removingIds: Set<string>;
    showWeight: boolean;
    onClear: () => void;
    idPrefix: string;
    copy: GroupEditorCopy;
}) {
    const t = useTranslations('group');
    const channelCount = useMemo(() => new Set(members.map((member) => member.channel_id)).size, [members]);
    const weightedCount = useMemo(() => members.filter((member) => (member.weight ?? 1) > 1).length, [members]);

    return (
        <section
            data-testid={`${idPrefix}-selected-section`}
            className="rounded-2xl border border-border/50 bg-muted/30 p-4 xl:sticky xl:top-0 xl:max-h-[calc(100vh-16rem)]"
        >
            <div className="space-y-1">
                <div className="text-sm font-medium text-foreground">{copy.selectionSummaryTitle}</div>
                <p className="text-xs text-muted-foreground">{copy.selectionSummaryHint}</p>
            </div>

            <div className="mt-3">
                <div className="flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
                    <span className="rounded-full bg-background px-2 py-1">{copy.selectionSummaryModels(members.length)}</span>
                    <span className="rounded-full bg-background px-2 py-1">{copy.selectionSummaryChannels(channelCount)}</span>
                    <span className="rounded-full bg-background px-2 py-1">{copy.selectionSummaryWeighted(weightedCount)}</span>
                    <span className="rounded-full bg-background px-2 py-1">
                        {showWeight ? copy.weightedModeActive : copy.weightedModeInactive}
                    </span>
                </div>
            </div>

            <div className="mt-3 flex items-center justify-between gap-2">
                <div className="text-sm font-medium text-foreground">{t('form.items')}</div>
                <Button type="button" variant="ghost" size="sm" onClick={onClear} disabled={members.length === 0} className="rounded-xl px-2">
                    {t('form.clear')}
                </Button>
            </div>

            <div className="mt-3 min-h-[12rem]">
                {members.length > 0 ? (
                    <MemberList
                        members={members}
                        onReorder={onReorder}
                        onRemove={onRemove}
                        onWeightChange={onWeightChange}
                        removingIds={removingIds}
                        showWeight={showWeight}
                        showConfirmDelete={false}
                        layoutScope={`${idPrefix}-members`}
                    />
                ) : (
                    <div data-testid={`${idPrefix}-selected-empty-state`} className="rounded-xl border border-dashed border-border/60 bg-background/70 p-4">
                        <div className="font-medium text-foreground">{copy.itemsEmptyTitle}</div>
                        <p className="mt-1 text-xs text-muted-foreground">{copy.itemsEmptyHint}</p>
                    </div>
                )}
            </div>
        </section>
    );
}

export function GroupEditor({
    initial,
    submitText,
    submittingText,
    isSubmitting,
    onSubmit,
    onCancel,
    idPrefix = 'group-editor',
}: {
    initial?: Partial<GroupEditorValues>;
    submitText: string;
    submittingText: string;
    isSubmitting: boolean;
    onSubmit: (values: GroupEditorValues) => void;
    onCancel?: () => void;
    idPrefix?: string;
}) {
    const t = useTranslations('group');
    const { data: modelChannels = [] } = useModelChannelList();

    const copy = useMemo<GroupEditorCopy>(() => ({
        flowTitle: t('form.flowTitle'),
        flowHint: t('form.flowHint'),
        flowDesc: t('form.flowDesc'),
        stepLabel: (index) => t('form.stepLabel', { index }),
        flowSteps: {
            naming: t('form.flowSteps.naming'),
            mode: t('form.flowSteps.mode'),
            models: t('form.flowSteps.models'),
        },
        namingSectionTitle: t('form.namingSectionTitle'),
        namingSectionHint: t('form.namingSectionHint'),
        namingSectionDesc: t('form.namingSectionDesc'),
        modeSectionTitle: t('form.modeSectionTitle'),
        modeSectionHint: t('form.modeSectionHint'),
        modeSectionDesc: t('form.modeSectionDesc'),
        nameHint: t('form.nameHint'),
        matchRegexHint: t('form.matchRegexHint'),
        advancedStrategyHint: t('form.advancedStrategyHint'),
        modelPickerTitle: t('form.modelPickerTitle'),
        modelPickerHint: t('form.modelPickerHint'),
        noAvailableModelsTitle: t('form.noAvailableModelsTitle'),
        noAvailableModelsHint: t('form.noAvailableModelsHint'),
        noFilteredModelsHint: t('form.noFilteredModelsHint'),
        itemsEmptyTitle: t('form.itemsEmptyTitle'),
        itemsEmptyHint: t('form.itemsEmptyHint'),
        selectionSummaryTitle: t('form.selectionSummaryTitle'),
        selectionSummaryModels: (count) => t('form.selectionSummaryModels', { count }),
        selectionSummaryChannels: (count) => t('form.selectionSummaryChannels', { count }),
        selectionSummaryWeighted: (count) => t('form.selectionSummaryWeighted', { count }),
        selectionSummaryEmpty: t('form.selectionSummaryEmpty'),
        selectionSummaryHint: t('form.selectionSummaryHint'),
        weightedModeActive: t('form.weightedModeActive'),
        weightedModeInactive: t('form.weightedModeInactive'),
        retryRounds: t('form.retryRounds'),
        retryRoundsHint: t('form.retryRoundsHint'),
        retryDelayMs: t('form.retryDelayMs'),
        retryDelayMsHint: t('form.retryDelayMsHint'),
        failoverWindowSec: t('form.failoverWindowSec'),
        failoverWindowSecHint: t('form.failoverWindowSecHint'),
        raceAfterFails: t('form.raceAfterFails'),
        raceAfterFailsHint: t('form.raceAfterFailsHint'),
        raceConcurrency: t('form.raceConcurrency'),
        raceConcurrencyHint: t('form.raceConcurrencyHint'),
        advancedStrategy: t('form.advancedStrategy'),
        advancedStrategyDesc: t('form.advancedStrategyDesc'),
    }), [t]);

    const [groupName, setGroupName] = useState(initial?.name ?? '');
    const [matchRegex, setMatchRegex] = useState(initial?.match_regex ?? '');
    const [mode, setMode] = useState<GroupMode>((initial?.mode ?? 1) as GroupMode);
    const [firstTokenTimeOut, setFirstTokenTimeOut] = useState<number>(initial?.first_token_time_out ?? 0);
    const [sessionKeepTime, setSessionKeepTime] = useState<number>(initial?.session_keep_time ?? 0);
    const [retryRounds, setRetryRounds] = useState<number>(initial?.retry_rounds ?? 0);
    const [retryDelayMs, setRetryDelayMs] = useState<number>(initial?.retry_delay_ms ?? 0);
    const [failoverWindowSec, setFailoverWindowSec] = useState<number>(initial?.failover_window_sec ?? 0);
    const [raceAfterFails, setRaceAfterFails] = useState<number>(initial?.race_after_fails ?? 0);
    const [raceConcurrency, setRaceConcurrency] = useState<number>(initial?.race_concurrency ?? 0);
    const [selectedMembers, setSelectedMembers] = useState<SelectedMember[]>(initial?.members ?? []);
    const [removingIds, setRemovingIds] = useState<Set<string>>(new Set());
    const [modelFilter, setModelFilter] = useState('');

    const groupKey = normalizeKey(groupName);
    const regexKey = matchRegex.trim();
    const invalidGroupName = /[:：\s]/.test(groupName.trim());

    const { matchedModelChannels, regexError } = useMemo(() => {
        const parseRegex = (input: string): RegExp => {
            const inlineMatch = input.match(/^\(\?([ism]+)\)(.+)$/);
            if (inlineMatch) {
                const flagMap: Record<string, string> = { i: 'i', s: 's', m: 'm' };
                const flags = inlineMatch[1].split('').map((flag) => flagMap[flag] || '').join('');
                return new RegExp(inlineMatch[2], flags);
            }
            return new RegExp(input);
        };

        if (regexKey) {
            try {
                const expression = parseRegex(regexKey);
                return {
                    matchedModelChannels: modelChannels.filter((channel) => expression.test(channel.name)),
                    regexError: '',
                };
            } catch {
                return { matchedModelChannels: [], regexError: 'invalid' };
            }
        }

        if (!groupKey) {
            return { matchedModelChannels: [], regexError: '' };
        }

        return {
            matchedModelChannels: modelChannels.filter((channel) => matchesGroupName(channel.name, groupKey)),
            regexError: '',
        };
    }, [groupKey, regexKey, modelChannels]);

    const autoAddDisabled = useMemo(() => {
        if ((!regexKey && !groupKey) || regexError || matchedModelChannels.length === 0) {
            return true;
        }
        const selectedKeys = new Set(selectedMembers.map((member) => member.id));
        return matchedModelChannels.every((channel) => selectedKeys.has(memberKey(channel)));
    }, [groupKey, matchedModelChannels, regexError, regexKey, selectedMembers]);

    const handleAddMember = useCallback((channel: LLMChannel) => {
        const id = memberKey(channel);
        setSelectedMembers((current) => {
            if (current.some((member) => member.id === id)) {
                return current;
            }
            return [...current, { ...channel, id, weight: 1 }];
        });
    }, []);

    const handleAutoAdd = useCallback(() => {
        if (matchedModelChannels.length === 0) {
            return;
        }
        setSelectedMembers((current) => {
            const selectedKeys = new Set(current.map((member) => member.id));
            const additions = matchedModelChannels
                .filter((channel) => !selectedKeys.has(memberKey(channel)))
                .map((channel) => ({ ...channel, id: memberKey(channel), weight: 1 }));
            return additions.length > 0 ? [...current, ...additions] : current;
        });
    }, [matchedModelChannels]);

    const handleWeightChange = useCallback((id: string, weight: number) => {
        setSelectedMembers((current) => current.map((member) => member.id === id ? { ...member, weight } : member));
    }, []);

    const handleRemoveMember = useCallback((id: string) => {
        setRemovingIds((current) => new Set(current).add(id));
        setTimeout(() => {
            setSelectedMembers((current) => current.filter((member) => member.id !== id));
            setRemovingIds((current) => {
                const next = new Set(current);
                next.delete(id);
                return next;
            });
        }, 200);
    }, []);

    const handleClearMembers = useCallback(() => {
        setSelectedMembers([]);
        setRemovingIds(new Set());
    }, []);

    const isValid = groupKey.length > 0 && selectedMembers.length > 0 && !regexError && !invalidGroupName;

    const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
        event.preventDefault();
        if (!isValid) {
            return;
        }

        onSubmit({
            name: groupName,
            match_regex: regexKey,
            mode,
            first_token_time_out: firstTokenTimeOut,
            session_keep_time: sessionKeepTime,
            retry_rounds: retryRounds,
            retry_delay_ms: retryDelayMs,
            failover_window_sec: failoverWindowSec,
            race_after_fails: raceAfterFails,
            race_concurrency: raceConcurrency,
            members: selectedMembers,
        });
    };

    return (
        <form onSubmit={handleSubmit} data-testid={`${idPrefix}-form`} className="flex h-full min-h-0 flex-col">
            <div className="flex-1 min-h-0 overflow-hidden pr-1">
                <FieldGroup className="flex h-full min-h-0 flex-col gap-4">
                    <section data-testid={`${idPrefix}-flow-card`} className="rounded-2xl border border-border/50 bg-muted/30 p-4">
                        <div className="space-y-1">
                            <div className="text-sm font-medium text-foreground">{copy.flowTitle}</div>
                            <p className="text-xs text-muted-foreground">{copy.flowHint}</p>
                            <p className="text-xs text-muted-foreground">{copy.flowDesc}</p>
                        </div>
                        <div className="mt-3 flex flex-wrap gap-2 text-xs text-muted-foreground">
                            {(['naming', 'mode', 'models'] as const).map((step, index) => (
                                <span key={step} className="rounded-full bg-background px-2 py-1">
                                    {copy.stepLabel(index + 1)} {copy.flowSteps[step as keyof GroupEditorCopy['flowSteps']]}
                                </span>
                            ))}
                        </div>
                    </section>

                    <section data-testid={`${idPrefix}-naming-section`} className="rounded-2xl border border-border/50 bg-muted/30 p-4">
                        <div className="space-y-1">
                            <div className="text-sm font-medium text-foreground">{copy.namingSectionTitle}</div>
                            <p className="text-xs text-muted-foreground">{copy.namingSectionHint}</p>
                            <p className="text-xs text-muted-foreground">{copy.namingSectionDesc}</p>
                        </div>

                        <div className="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] 2xl:grid-cols-[minmax(0,1.02fr)_minmax(0,0.98fr)]">
                            <Field>
                                <FieldLabel htmlFor={`${idPrefix}-name`}>{t('form.name')}</FieldLabel>
                                <Input id={`${idPrefix}-name`} value={groupName} onChange={(event) => setGroupName(event.target.value)} className="rounded-xl" />
                                <p className="mt-1 text-xs text-muted-foreground">{copy.nameHint}</p>
                                {invalidGroupName ? <p className="mt-1 text-xs text-destructive">{t('form.nameRule')}</p> : null}
                            </Field>

                            <Field>
                                <FieldLabel htmlFor={`${idPrefix}-match-regex`}>{t('form.matchRegex')}</FieldLabel>
                                <Input
                                    id={`${idPrefix}-match-regex`}
                                    value={matchRegex}
                                    onChange={(event) => setMatchRegex(event.target.value)}
                                    className="rounded-xl"
                                    placeholder={t('form.matchRegexPlaceholder')}
                                />
                                <p className="mt-1 text-xs text-muted-foreground">{copy.matchRegexHint}</p>
                                {regexError ? <p className="mt-1 text-xs text-destructive">{t('form.matchRegexInvalid')}: {t('form.matchRegexInvalidHint')}</p> : null}
                            </Field>
                        </div>
                    </section>

                    <section data-testid={`${idPrefix}-mode-section`} className="rounded-2xl border border-border/50 bg-muted/30 p-4">
                        <div className="space-y-1">
                            <div className="text-sm font-medium text-foreground">{copy.modeSectionTitle}</div>
                            <p className="text-xs text-muted-foreground">{copy.modeSectionHint}</p>
                            <p className="text-xs text-muted-foreground">{copy.modeSectionDesc}</p>
                        </div>

                        <div className="mt-4 flex gap-1">
                            {([1, 2, 3, 4, 5] as const).map((itemMode) => (
                                <button
                                    key={itemMode}
                                    type="button"
                                    onClick={() => setMode(itemMode)}
                                    className={cn('flex-1 rounded-lg py-1 text-xs transition-colors', mode === itemMode ? 'bg-primary text-primary-foreground' : 'bg-muted hover:bg-muted/80')}
                                >
                                    {t(`mode.${MODE_LABELS[itemMode]}`)}
                                </button>
                            ))}
                        </div>

                        <div className="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
                            <Field>
                                <FieldLabel htmlFor={`${idPrefix}-first-token-timeout`}>{t('form.firstTokenTimeOut')}</FieldLabel>
                                <Input
                                    id={`${idPrefix}-first-token-timeout`}
                                    type="number"
                                    inputMode="numeric"
                                    min={0}
                                    step={1}
                                    value={String(firstTokenTimeOut)}
                                    onChange={(event) => setFirstTokenTimeOut(parsePositiveInt(event.target.value))}
                                    className="rounded-xl"
                                />
                                <p className="mt-1 text-xs text-muted-foreground">{t('form.firstTokenTimeOutHint')}</p>
                            </Field>

                            <Field>
                                <FieldLabel htmlFor={`${idPrefix}-session-keep-time`}>{t('form.sessionKeepTime')}</FieldLabel>
                                <Input
                                    id={`${idPrefix}-session-keep-time`}
                                    type="number"
                                    inputMode="numeric"
                                    min={0}
                                    step={1}
                                    value={String(sessionKeepTime)}
                                    onChange={(event) => setSessionKeepTime(parsePositiveInt(event.target.value))}
                                    className="rounded-xl"
                                />
                                <p className="mt-1 text-xs text-muted-foreground">{t('form.sessionKeepTimeHint')}</p>
                            </Field>
                        </div>
                    </section>

                    <section data-testid={`${idPrefix}-advanced-strategy-section`}>
                        <Accordion type="single" collapsible className="rounded-2xl border border-border/50 bg-muted/20 px-4">
                            <AccordionItem data-testid={`${idPrefix}-advanced-strategy-item`} value="advanced-strategy" className="border-none">
                                <AccordionTrigger data-testid="group-advanced-strategy-trigger" className="py-3" addon={<HelpHint className="mt-1 size-3.5">{copy.advancedStrategyHint}</HelpHint>}>
                                    <div className="space-y-1">
                                        <div className="text-sm font-medium text-foreground">{copy.advancedStrategy}</div>
                                        <div className="text-xs text-muted-foreground">{copy.advancedStrategyDesc}</div>
                                    </div>
                                </AccordionTrigger>
                                <AccordionContent>
                                    <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
                                        <Field>
                                            <FieldLabel htmlFor={`${idPrefix}-retry-rounds`}>{copy.retryRounds}</FieldLabel>
                                            <Input id={`${idPrefix}-retry-rounds`} type="number" inputMode="numeric" min={0} step={1} value={String(retryRounds)} onChange={(event) => setRetryRounds(parsePositiveInt(event.target.value))} className="rounded-xl" />
                                            <p className="mt-1 text-xs text-muted-foreground">{copy.retryRoundsHint}</p>
                                        </Field>
                                        <Field>
                                            <FieldLabel htmlFor={`${idPrefix}-retry-delay-ms`}>{copy.retryDelayMs}</FieldLabel>
                                            <Input id={`${idPrefix}-retry-delay-ms`} type="number" inputMode="numeric" min={0} step={1} value={String(retryDelayMs)} onChange={(event) => setRetryDelayMs(parsePositiveInt(event.target.value))} className="rounded-xl" />
                                            <p className="mt-1 text-xs text-muted-foreground">{copy.retryDelayMsHint}</p>
                                        </Field>
                                        <Field>
                                            <FieldLabel htmlFor={`${idPrefix}-failover-window-sec`}>{copy.failoverWindowSec}</FieldLabel>
                                            <Input id={`${idPrefix}-failover-window-sec`} type="number" inputMode="numeric" min={0} step={1} value={String(failoverWindowSec)} onChange={(event) => setFailoverWindowSec(parsePositiveInt(event.target.value))} className="rounded-xl" />
                                            <p className="mt-1 text-xs text-muted-foreground">{copy.failoverWindowSecHint}</p>
                                        </Field>
                                        <Field>
                                            <FieldLabel htmlFor={`${idPrefix}-race-after-fails`}>{copy.raceAfterFails}</FieldLabel>
                                            <Input id={`${idPrefix}-race-after-fails`} type="number" inputMode="numeric" min={0} step={1} value={String(raceAfterFails)} onChange={(event) => setRaceAfterFails(parsePositiveInt(event.target.value))} className="rounded-xl" />
                                            <p className="mt-1 text-xs text-muted-foreground">{copy.raceAfterFailsHint}</p>
                                        </Field>
                                        <Field>
                                            <FieldLabel htmlFor={`${idPrefix}-race-concurrency`}>{copy.raceConcurrency}</FieldLabel>
                                            <Input id={`${idPrefix}-race-concurrency`} type="number" inputMode="numeric" min={0} step={1} value={String(raceConcurrency)} onChange={(event) => setRaceConcurrency(parsePositiveInt(event.target.value))} className="rounded-xl" />
                                            <p className="mt-1 text-xs text-muted-foreground">{copy.raceConcurrencyHint}</p>
                                        </Field>
                                    </div>
                                </AccordionContent>
                            </AccordionItem>
                        </Accordion>
                    </section>

                    <div className="flex-1 min-h-0">
                        <div className="grid h-full min-h-0 grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] 2xl:grid-cols-[minmax(0,1.02fr)_minmax(0,0.98fr)]">
                            <ModelPickerSection
                                modelChannels={matchedModelChannels}
                                selectedMembers={selectedMembers}
                                onAdd={handleAddMember}
                                onAutoAdd={handleAutoAdd}
                                autoAddDisabled={autoAddDisabled}
                                idPrefix={idPrefix}
                                copy={copy}
                                modelFilter={modelFilter}
                                onModelFilterChange={setModelFilter}
                            />

                            <SortSection
                                members={selectedMembers}
                                onReorder={setSelectedMembers}
                                onRemove={handleRemoveMember}
                                onWeightChange={handleWeightChange}
                                removingIds={removingIds}
                                showWeight={mode === 4}
                                onClear={handleClearMembers}
                                idPrefix={idPrefix}
                                copy={copy}
                            />
                        </div>
                    </div>
                </FieldGroup>
            </div>

            <div className="mt-auto shrink-0 pt-4">
                <div className="flex gap-2">
                    {onCancel ? <Button type="button" variant="secondary" className="h-11 flex-1 rounded-xl" onClick={onCancel}>{t('detail.actions.cancel')}</Button> : null}
                    <Button type="submit" disabled={!isValid || isSubmitting} className="h-11 flex-1 rounded-xl">{isSubmitting ? submittingText : submitText}</Button>
                </div>
            </div>
        </form>
    );
}
