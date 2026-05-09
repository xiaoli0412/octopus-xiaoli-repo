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
          <div className={cn('text-[10px] uppercase tracking-[0.16em] text-slate-500', labelClassName)}>{item.label}</div>
          <div className={cn('mt-1.5 text-sm font-medium leading-5 text-slate-100', valueClassName, item.valueClassName)}>{item.value}</div>
        </div>
      ))}
    </div>
  );
}
export function LearningSummaryActionBar({ actions, hint, className }: LearningSummaryActionBarProps) {
  if (actions.length === 0 && !hint) return null;

  return (
    <div className={cn('space-y-2', className)}>
      {actions.length > 0 ? (
        <div className="flex flex-wrap items-center gap-2">
          {actions.map((action) => (
            <Button
              key={action.key}
              type="button"
              variant="outline"
              size="sm"
              className="rounded-[14px] border-slate-700 bg-slate-950/40 text-slate-100 hover:bg-slate-900"
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
      {hint ? <div className="text-[11px] leading-5 text-slate-500">{hint}</div> : null}
    </div>
  );
}

export function LearningSummaryPanel({ primaryGrid, secondaryGrid, noticeTitle, noticeBody, noticeClassName, footer }: LearningSummaryPanelProps) {
  return (
    <div className="space-y-3">
      <SummaryGrid {...primaryGrid} />
      {secondaryGrid ? <SummaryGrid {...secondaryGrid} /> : null}
      <div className={noticeClassName}>
        <div className="font-medium text-slate-100">{noticeTitle}</div>
        <div className="mt-1.5">{noticeBody}</div>
      </div>
      {footer}
    </div>
  );
}
