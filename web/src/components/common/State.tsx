'use client';

import { useTranslations } from 'next-intl';
import { Loader2, Inbox, AlertTriangle } from 'lucide-react';
import { cn } from '@/lib/utils';

export function LoadingState({ className }: { className?: string }) {
    const t = useTranslations('common.state');
    return (
        <div className={cn('octo-loading', className)}>
            <Loader2 className="size-4 animate-spin" />
            {t('loading')}
        </div>
    );
}

export function EmptyState({ className, children }: { className?: string; children?: React.ReactNode }) {
    const t = useTranslations('common.state');
    return (
        <div className={cn('octo-empty', className)}>
            <Inbox className="size-4 opacity-70" />
            {children ?? t('empty')}
        </div>
    );
}

export function ErrorState({ className, children, onRetry }: { className?: string; children?: React.ReactNode; onRetry?: () => void }) {
    const t = useTranslations('common.state');
    return (
        <div className={cn('octo-error', className)}>
            <AlertTriangle className="size-4" />
            <span>{children ?? t('error')}</span>
            {onRetry ? (
                <button
                    type="button"
                    onClick={onRetry}
                    className="rounded-lg border border-destructive/40 px-3 py-1 text-xs font-medium text-destructive transition hover:bg-destructive/10"
                >
                    {t('retry')}
                </button>
            ) : null}
        </div>
    );
}
