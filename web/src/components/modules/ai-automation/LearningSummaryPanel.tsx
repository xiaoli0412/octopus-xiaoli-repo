import type { ReactNode } from 'react';

import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

export type LearningSummaryItem = {
    key: string;
    label: string;
    value: ReactNode;
    testId?: string;
    className?: string;
    valueClassName?: string;
};

export type LearningSummaryAction = {
    key: string;
    label: string;
    onClick: () => void;
    disabled?: boolean;
    testId?: string;
    icon?: ReactNode;
};

type LearningSummaryGrid = {
    items: LearningSummaryItem[];
    className: string;
    cardClassName: string;
    testId?: string;
    labelClassName?: string;
    valueClassName?: string;
};

type LearningSummaryPanelProps = {
    primaryGrid: LearningSummaryGrid;
    secondaryGrid?: LearningSummaryGrid;
    noticeTitle: string;
    noticeBody: ReactNode;
    noticeClassName: string;
    footer?: ReactNode;
};

type LearningSummaryActionBarProps = {
    actions: LearningSummaryAction[];
    hint?: ReactNode;
    className?: string;
};

function SummaryGrid({ items, className, cardClassName, testId, labelClassName, valueClassName }: LearningSummaryGrid) {
    if (items.length === 0) return null;

    return (
        <div data-testid={testId} className={className}>
            {items.map((item) => (
                <div key={item.key} data-testid={item.testId} className={cn(cardClassName, item.className)}>
                    <div className={cn('text-[11px] text-muted-foreground', labelClassName)}>{item.label}</div>
                    <div className={cn('mt-1 text-sm font-medium text-foreground', valueClassName, item.valueClassName)}>{item.value}</div>
                </div>
            ))}
        </div>
    );
}

export function LearningSummaryActionBar({ actions, hint, className }: LearningSummaryActionBarProps) {
    if (actions.length === 0 && !hint) return null;

    return (
        <div className={cn('space-y-1.5', className)}>
            {actions.length > 0 ? (
                <div className="flex flex-wrap items-center gap-2">
                    {actions.map((action) => (
                        <Button
                            key={action.key}
                            type="button"
                            variant="outline"
                            size="sm"
                            data-testid={action.testId}
                            onClick={action.onClick}
                            disabled={action.disabled}
                        >
                            {action.icon}
                            {action.label}
                        </Button>
                    ))}
                </div>
            ) : null}
            {hint ? <div className="text-[11px] text-muted-foreground">{hint}</div> : null}
        </div>
    );
}

export function LearningSummaryPanel({ primaryGrid, secondaryGrid, noticeTitle, noticeBody, noticeClassName, footer }: LearningSummaryPanelProps) {
    return (
        <div className="space-y-2.5">
            <SummaryGrid {...primaryGrid} />
            {secondaryGrid ? <SummaryGrid {...secondaryGrid} /> : null}
            <div className={noticeClassName}>
                <div className="font-medium text-foreground">{noticeTitle}</div>
                <div className="mt-1.5">{noticeBody}</div>
            </div>
            {footer}
        </div>
    );
}
