'use client';

import { useId, type ReactNode } from 'react';
import { HelpCircle } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/animate-ui/components/animate/tooltip';
import { cn } from '@/lib/utils';

type HelpHintProps = {
    children: ReactNode;
    className?: string;
    ariaLabel?: string;
};

export function HelpHint({ children, className, ariaLabel }: HelpHintProps) {
    const t = useTranslations('common.helpHint');
    const resolvedAriaLabel = ariaLabel ?? t('ariaLabel');
    const hintId = useId();

    return (
        <TooltipProvider>
            <Tooltip>
                <TooltipTrigger asChild>
                    <button
                        type="button"
                        aria-label={resolvedAriaLabel}
                        data-slot="help-hint-trigger"
                        data-help-hint-trigger="true"
                        data-help-hint-id={hintId}
                        className={cn(
                            'inline-flex h-4 w-4 cursor-help items-center justify-center rounded-full border-0 bg-transparent p-0 text-muted-foreground outline-none transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background',
                            className
                        )}
                    >
                        <HelpCircle aria-hidden="true" className="size-full" />
                    </button>
                </TooltipTrigger>
                <TooltipContent
                    id={hintId}
                    className="max-w-72 text-xs leading-5"
                    data-slot="help-hint-content"
                    data-help-hint-id={hintId}
                >
                    {children}
                </TooltipContent>
            </Tooltip>
        </TooltipProvider>
    );
}
