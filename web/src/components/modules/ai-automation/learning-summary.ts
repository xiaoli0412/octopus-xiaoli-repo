import { formatDateTimeByLocale } from '@/lib/locale';
import type { DynamicRouteLearningState } from '@/api/endpoints/ai-automation';

export type LearningSummary = {
    hasSamples: boolean;
    sampleCount: number;
    canReset: boolean;
    latestState: DynamicRouteLearningState | null;
    topState: DynamicRouteLearningState | null;
};

export type LearningNoticeState = 'enabled_with_samples' | 'enabled_without_samples' | 'disabled_with_samples' | 'disabled_without_samples';

export type LearningDisplayState = {
    enabled: boolean;
    sampleCount: number;
    hasSamples: boolean;
    canReset: boolean;
    runtimeKey: 'runtimeEnabled' | 'runtimeDisabled';
    noticeState: LearningNoticeState;
};

export type LearningSummaryCardItem = {
    key: string;
    label: string;
    value: string | number;
    testId?: string;
    valueClassName?: string;
};

export type LearningSummaryItemLabels = {
    statusLabel: string;
    enabledLabel: string;
    disabledLabel: string;
    samplesLabel: string;
    runtimeLabel: string;
    runtimeEnabledLabel: string;
    runtimeDisabledLabel: string;
    latestSampleLabel: string;
    topTargetLabel: string;
};

export type LearningNoticeValues<T> = {
    enabledWithSamples: T;
    enabledWithoutSamples: T;
    disabledWithSamples: T;
    disabledWithoutSamples: T;
};

export type LearningSummaryViewModel = {
    summary: LearningSummary;
    display: LearningDisplayState;
    latestSampleLabel: string;
    topTargetLabel: string;
    notice: string;
    items: LearningSummaryCardItem[];
};

export type LearningSummaryItemKey = LearningSummaryCardItem['key'];

export type LearningSummarySectionEntry = LearningSummaryItemKey | LearningSummaryCardItem;

export type LearningSummarySectionEntries = {
    primaryEntries: LearningSummarySectionEntry[];
    secondaryEntries: LearningSummarySectionEntry[];
};

export const DEFAULT_LEARNING_SUMMARY_SECTION_ENTRIES = {
    primaryEntries: ['status', 'samples', 'runtime'],
    secondaryEntries: ['latest-sample', 'top-target'],
} satisfies LearningSummarySectionEntries;

export type LearningSummarySections = {
    primaryItems: LearningSummaryCardItem[];
    secondaryItems: LearningSummaryCardItem[];
};

export type BuildLearningSummaryViewModelOptions = {
    states: DynamicRouteLearningState[];
    enabled: boolean;
    locale: Parameters<typeof formatDateTimeByLocale>[1];
    emptyLabel: string;
    topTargetFormatter: (state: DynamicRouteLearningState) => string;
    itemLabels: LearningSummaryItemLabels;
    noticeValues: LearningNoticeValues<string>;
    runtimeTestId?: string;
    topTargetValueClassName?: string;
};

export function summarizeDynamicRouteLearning(states: DynamicRouteLearningState[]): LearningSummary {
    const latestState = states.reduce<DynamicRouteLearningState | null>((latest, item) => {
        if (!latest) return item;
        return (item.last_sample_at ?? 0) > (latest.last_sample_at ?? 0) ? item : latest;
    }, null);

    const topState = states.reduce<DynamicRouteLearningState | null>((best, item) => {
        if (!best) return item;
        return (item.score ?? 0) > (best.score ?? 0) ? item : best;
    }, null);

    return {
        hasSamples: states.length > 0,
        sampleCount: states.length,
        canReset: states.length > 0,
        latestState,
        topState,
    };
}

export function deriveLearningDisplayState(enabled: boolean, summary: LearningSummary): LearningDisplayState {
    const noticeState: LearningNoticeState = enabled
        ? summary.hasSamples
            ? 'enabled_with_samples'
            : 'enabled_without_samples'
        : summary.hasSamples
            ? 'disabled_with_samples'
            : 'disabled_without_samples';

    return {
        enabled,
        sampleCount: summary.sampleCount,
        hasSamples: summary.hasSamples,
        canReset: summary.canReset,
        runtimeKey: enabled ? 'runtimeEnabled' : 'runtimeDisabled',
        noticeState,
    };
}

export function resolveLearningNoticeValue<T>(noticeState: LearningNoticeState, values: LearningNoticeValues<T>): T {
    switch (noticeState) {
        case 'enabled_with_samples':
            return values.enabledWithSamples;
        case 'enabled_without_samples':
            return values.enabledWithoutSamples;
        case 'disabled_with_samples':
            return values.disabledWithSamples;
        default:
            return values.disabledWithoutSamples;
    }
}

export function formatLearningTopTarget(
    state: DynamicRouteLearningState | null,
    formatter: (state: DynamicRouteLearningState) => string,
    emptyLabel: string,
) {
    if (!state) return emptyLabel;
    return formatter(state);
}

export function buildLearningSummaryItems({
    display,
    latestSampleLabel,
    topTargetLabel,
    labels,
    runtimeTestId,
    topTargetValueClassName,
}: {
    display: LearningDisplayState;
    latestSampleLabel: string;
    topTargetLabel: string;
    labels: LearningSummaryItemLabels;
    runtimeTestId?: string;
    topTargetValueClassName?: string;
}): LearningSummaryCardItem[] {
    return [
        {
            key: 'status',
            label: labels.statusLabel,
            value: display.enabled ? labels.enabledLabel : labels.disabledLabel,
        },
        {
            key: 'samples',
            label: labels.samplesLabel,
            value: display.sampleCount,
        },
        {
            key: 'runtime',
            label: labels.runtimeLabel,
            value: display.runtimeKey === 'runtimeEnabled' ? labels.runtimeEnabledLabel : labels.runtimeDisabledLabel,
            testId: runtimeTestId,
        },
        {
            key: 'latest-sample',
            label: labels.latestSampleLabel,
            value: latestSampleLabel,
        },
        {
            key: 'top-target',
            label: labels.topTargetLabel,
            value: topTargetLabel,
            valueClassName: topTargetValueClassName,
        },
    ];
}

export function buildLearningSummaryViewModel({
    states,
    enabled,
    locale,
    emptyLabel,
    topTargetFormatter,
    itemLabels,
    noticeValues,
    runtimeTestId,
    topTargetValueClassName,
}: BuildLearningSummaryViewModelOptions): LearningSummaryViewModel {
    const summary = summarizeDynamicRouteLearning(states);
    const display = deriveLearningDisplayState(enabled, summary);
    const latestSampleLabel = formatLearningSampleTime(summary.latestState?.last_sample_at, locale, emptyLabel);
    const topTargetLabel = formatLearningTopTarget(summary.topState, topTargetFormatter, emptyLabel);
    const notice = resolveLearningNoticeValue(display.noticeState, noticeValues);
    const items = buildLearningSummaryItems({
        display,
        latestSampleLabel,
        topTargetLabel,
        labels: itemLabels,
        runtimeTestId,
        topTargetValueClassName,
    });

    return {
        summary,
        display,
        latestSampleLabel,
        topTargetLabel,
        notice,
        items,
    };
}

export function indexLearningSummaryItems(items: LearningSummaryCardItem[]) {
    return items.reduce<Partial<Record<LearningSummaryItemKey, LearningSummaryCardItem>>>((indexed, item) => {
        indexed[item.key] = item;
        return indexed;
    }, {});
}

function resolveLearningSummarySectionItems(
    entries: LearningSummarySectionEntry[],
    indexed: Partial<Record<LearningSummaryItemKey, LearningSummaryCardItem>>,
    errorScope: string,
) {
    return entries.map((entry) => {
        if (typeof entry !== 'string') return entry;

        const item = indexed[entry];
        if (!item) {
            throw new Error(`${errorScope} summary items are incomplete`);
        }

        return item;
    });
}

export function buildLearningSummarySections({
    items,
    primaryEntries,
    secondaryEntries,
    errorScope,
}: {
    items: LearningSummaryCardItem[];
    primaryEntries: LearningSummarySectionEntry[];
    secondaryEntries: LearningSummarySectionEntry[];
    errorScope: string;
}): LearningSummarySections {
    const indexed = indexLearningSummaryItems(items);

    return {
        primaryItems: resolveLearningSummarySectionItems(primaryEntries, indexed, errorScope),
        secondaryItems: resolveLearningSummarySectionItems(secondaryEntries, indexed, errorScope),
    };
}

export function formatLearningSampleTime(
    value: number | undefined,
    locale: Parameters<typeof formatDateTimeByLocale>[1],
    emptyLabel: string,
) {
    if (!value || value <= 0) return emptyLabel;
    return formatDateTimeByLocale(new Date(value * 1000), locale);
}
