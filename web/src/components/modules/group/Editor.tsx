'use client';

import { useCallback, useMemo, useState, type FormEvent } from 'react';
import { Check, Plus, Sparkles } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { useModelChannelList, type LLMChannel } from '@/api/endpoints/model';
import { Button } from '@/components/ui/button';
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { cn } from '@/lib/utils';
import { getModelIcon } from '@/lib/model-icons';
import type { GroupMode } from '@/api/endpoints/group';
import type { SelectedMember } from './ItemList';
import { MemberList } from './ItemList';
import { matchesGroupName, memberKey, normalizeKey, MODE_LABELS } from './utils';

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
    modelFilter,
    onModelFilterChange,
}: {
    modelChannels: LLMChannel[];
    selectedMembers: SelectedMember[];
    onAdd: (channel: LLMChannel) => void;
    onAutoAdd: () => void;
    autoAddDisabled: boolean;
    idPrefix: string;
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

    const emptyTitle = hasAnyModelChannels ? t('form.noFilteredModels') : t('form.noAvailableModelsTitle');

    return (
        <section data-testid={`${idPrefix}-model-picker-section`} className="rounded-xl border border-border/50 bg-muted/20 p-3.5">
            <div className="flex items-center justify-between gap-3">
                <div className="truncate text-sm font-medium text-foreground">{t('form.modelPickerTitle')}</div>
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

            <div className="mt-3 space-y-3">
                <Input
                    value={modelFilter}
                    onChange={(event) => onModelFilterChange(event.target.value)}
                    placeholder={t('form.modelFilterPlaceholder')}
                    data-testid={`${idPrefix}-model-filter`}
                    className="rounded-xl"
                />

                {channels.length > 0 ? (
                    <div className="grid max-h-[24rem] grid-cols-1 gap-3 overflow-y-auto pr-1 md:grid-cols-2 xl:grid-cols-3">
                        {channels.map((channel) => {
                            const total = channel.models.length;
                            const selectedCount = channel.models.reduce(
                                (count, model) => count + (selectedKeys.has(memberKey(model)) ? 1 : 0),
                                0,
                            );

                            return (
                                <Accordion
                                    key={channel.id}
                                    type="single"
                                    collapsible
                                    className="rounded-xl border border-border/50 bg-background/80 px-3"
                                >
                                    <AccordionItem value={`channel-${channel.id}`} className="border-none">
                                        <AccordionTrigger className="py-2.5 text-sm font-medium">
                                            <div className="min-w-0 space-y-1">
                                                <div className="truncate text-sm font-medium text-foreground">{channel.name}</div>
                                                <div className="text-xs text-muted-foreground">{selectedCount}/{total}</div>
                                            </div>
                                        </AccordionTrigger>
                                        <AccordionContent>
                                            <div className="max-h-56 space-y-2 overflow-y-auto pr-1">
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
                    <div data-testid={`${idPrefix}-model-empty-state`} className="rounded-xl border border-dashed border-border/60 bg-background/70 px-3 py-4 text-sm text-muted-foreground">
                        {emptyTitle}
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
}: {
    members: SelectedMember[];
    onReorder: (members: SelectedMember[]) => void;
    onRemove: (id: string) => void;
    onWeightChange: (id: string, weight: number) => void;
    removingIds: Set<string>;
    showWeight: boolean;
    onClear: () => void;
    idPrefix: string;
}) {
    const t = useTranslations('group');
    const channelCount = useMemo(() => new Set(members.map((member) => member.channel_id)).size, [members]);
    const weightedCount = useMemo(() => members.filter((member) => (member.weight ?? 1) > 1).length, [members]);

    return (
        <section
            data-testid={`${idPrefix}-selected-section`}
            className="rounded-xl border border-border/50 bg-muted/20 p-3.5 xl:max-h-[calc(100vh-19rem)]"
        >
            <div className="flex items-center justify-between gap-2">
                <div className="text-sm font-medium text-foreground">{t('form.items')}</div>
                <Button type="button" variant="ghost" size="sm" onClick={onClear} disabled={members.length === 0} className="rounded-xl px-2">
                    {t('form.clear')}
                </Button>
            </div>

            <div className="mt-3 flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
                <span className="rounded-full bg-background px-2 py-1">{t('form.selectionSummaryModels', { count: members.length })}</span>
                <span className="rounded-full bg-background px-2 py-1">{t('form.selectionSummaryChannels', { count: channelCount })}</span>
                <span className="rounded-full bg-background px-2 py-1">{t('form.selectionSummaryWeighted', { count: weightedCount })}</span>
                <span className="rounded-full bg-background px-2 py-1">
                    {showWeight ? t('form.weightedModeActive') : t('form.weightedModeInactive')}
                </span>
            </div>

            <div className="mt-3 min-h-[16rem]">
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
                    <div data-testid={`${idPrefix}-selected-empty-state`} className="rounded-xl border border-dashed border-border/60 bg-background/70 px-3 py-4 text-sm text-muted-foreground">
                        {t('form.itemsEmptyTitle')}
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
            <div className="flex-1 min-h-0 overflow-y-auto pr-1">
                <FieldGroup className="flex h-full min-h-0 flex-col gap-3">
                    <section className="rounded-xl border border-border/50 bg-muted/20 p-3.5">
                        <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                            <Field>
                                <FieldLabel htmlFor={`${idPrefix}-name`}>{t('form.name')}</FieldLabel>
                                <Input id={`${idPrefix}-name`} value={groupName} onChange={(event) => setGroupName(event.target.value)} className="rounded-xl" />
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
                                {regexError ? <p className="mt-1 text-xs text-destructive">{t('form.matchRegexInvalid')}: {t('form.matchRegexInvalidHint')}</p> : null}
                            </Field>
                        </div>

                        <div className="mt-3">
                            <div className="mb-2 text-sm font-medium text-foreground">{t('form.modeSectionTitle')}</div>
                        <div className="grid grid-cols-2 gap-1.5 sm:grid-cols-4 lg:grid-cols-7">
                            {([1, 2, 3, 4, 5, 6, 7] as const).map((itemMode) => (
                                <button
                                    key={itemMode}
                                    type="button"
                                    onClick={() => setMode(itemMode)}
                                    className={cn('rounded-lg py-2 text-xs transition-colors', mode === itemMode ? 'bg-primary text-primary-foreground' : 'bg-background hover:bg-muted')}
                                >
                                    {t(`mode.${MODE_LABELS[itemMode]}`)}
                                </button>
                            ))}
                        </div>
                        </div>

                        <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-2">
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
                            </Field>
                        </div>
                    </section>

                    <section data-testid={`${idPrefix}-advanced-strategy-section`}>
                        <Accordion type="single" collapsible className="rounded-xl border border-border/50 bg-muted/10 px-3">
                            <AccordionItem data-testid={`${idPrefix}-advanced-strategy-item`} value="advanced-strategy" className="border-none">
                                <AccordionTrigger data-testid="group-advanced-strategy-trigger" className="py-3 text-sm font-medium text-foreground">
                                    {t('form.advancedStrategy')}
                                </AccordionTrigger>
                                <AccordionContent>
                                    <div className="grid grid-cols-1 gap-3 pb-1 md:grid-cols-2 xl:grid-cols-3">
                                        <Field>
                                            <FieldLabel htmlFor={`${idPrefix}-retry-rounds`}>{t('form.retryRounds')}</FieldLabel>
                                            <Input id={`${idPrefix}-retry-rounds`} type="number" inputMode="numeric" min={0} step={1} value={String(retryRounds)} onChange={(event) => setRetryRounds(parsePositiveInt(event.target.value))} className="rounded-xl" />
                                        </Field>
                                        <Field>
                                            <FieldLabel htmlFor={`${idPrefix}-retry-delay-ms`}>{t('form.retryDelayMs')}</FieldLabel>
                                            <Input id={`${idPrefix}-retry-delay-ms`} type="number" inputMode="numeric" min={0} step={1} value={String(retryDelayMs)} onChange={(event) => setRetryDelayMs(parsePositiveInt(event.target.value))} className="rounded-xl" />
                                        </Field>
                                        <Field>
                                            <FieldLabel htmlFor={`${idPrefix}-failover-window-sec`}>{t('form.failoverWindowSec')}</FieldLabel>
                                            <Input id={`${idPrefix}-failover-window-sec`} type="number" inputMode="numeric" min={0} step={1} value={String(failoverWindowSec)} onChange={(event) => setFailoverWindowSec(parsePositiveInt(event.target.value))} className="rounded-xl" />
                                        </Field>
                                        <Field>
                                            <FieldLabel htmlFor={`${idPrefix}-race-after-fails`}>{t('form.raceAfterFails')}</FieldLabel>
                                            <Input id={`${idPrefix}-race-after-fails`} type="number" inputMode="numeric" min={0} step={1} value={String(raceAfterFails)} onChange={(event) => setRaceAfterFails(parsePositiveInt(event.target.value))} className="rounded-xl" />
                                        </Field>
                                        <Field>
                                            <FieldLabel htmlFor={`${idPrefix}-race-concurrency`}>{t('form.raceConcurrency')}</FieldLabel>
                                            <Input id={`${idPrefix}-race-concurrency`} type="number" inputMode="numeric" min={0} step={1} value={String(raceConcurrency)} onChange={(event) => setRaceConcurrency(parsePositiveInt(event.target.value))} className="rounded-xl" />
                                        </Field>
                                    </div>
                                </AccordionContent>
                            </AccordionItem>
                        </Accordion>
                    </section>

                    <div className="flex-1 min-h-[18rem]">
                        <div className="grid h-full min-h-0 grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(0,1.05fr)_minmax(0,0.95fr)]">
                            <ModelPickerSection
                                modelChannels={modelChannels}
                                selectedMembers={selectedMembers}
                                onAdd={handleAddMember}
                                onAutoAdd={handleAutoAdd}
                                autoAddDisabled={autoAddDisabled}
                                idPrefix={idPrefix}
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
