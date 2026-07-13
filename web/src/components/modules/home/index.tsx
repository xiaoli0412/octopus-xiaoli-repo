'use client';

import { useState, useCallback } from 'react';
import { Download, Loader2 } from 'lucide-react';
import { Activity } from './activity';
import { Total } from './total';
import { StatsChart } from './chart';
import { Rank } from './rank';
import { TokenBreakdown } from './token-breakdown';
import { PageWrapper } from '@/components/common/PageWrapper';
import { useTranslations } from 'next-intl';
import { exportStatsCsv, type StatsExportDimension } from '@/api/endpoints/stats';
import { toast } from '@/components/common/Toast';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { cn } from '@/lib/utils';

export function Home() {
    const t = useTranslations('home.total');
    const [exporting, setExporting] = useState(false);
    const [exportOpen, setExportOpen] = useState(false);

    const handleExport = useCallback(async (dimension: StatsExportDimension) => {
        setExportOpen(false);
        setExporting(true);
        try {
            await exportStatsCsv({ dimension });
            toast.success(t('exportSuccess'));
        } catch (error) {
            const msg = error instanceof Error ? error.message : String(error);
            toast.error(t('exportFailed'), { description: msg });
        } finally {
            setExporting(false);
        }
    }, [t]);

    return (
        <PageWrapper data-testid="home-page" className="h-full min-h-0 overflow-y-auto overscroll-contain space-y-5 pb-24 md:pb-4 rounded-t-3xl md:space-y-6">
            <div className="flex items-center justify-end">
                <Popover open={exportOpen} onOpenChange={setExportOpen}>
                    <PopoverTrigger asChild>
                        <button
                            type="button"
                            disabled={exporting}
                            className={cn(
                                'flex h-8 items-center gap-1.5 rounded-xl border border-border/60 bg-card/80 px-3 text-xs font-medium text-muted-foreground transition hover:bg-muted/50 hover:text-foreground disabled:opacity-50'
                            )}
                        >
                            {exporting ? <Loader2 className="size-4 animate-spin" /> : <Download className="size-4" />}
                            {t('exportCsv')}
                        </button>
                    </PopoverTrigger>
                    <PopoverContent align="end" sideOffset={8} className="w-44 rounded-xl border border-border/60 p-1">
                        <button
                            type="button"
                            onClick={() => handleExport('channel')}
                            className="flex w-full items-center rounded-lg px-3 py-2 text-xs text-foreground transition hover:bg-muted/50"
                        >
                            {t('exportDimensionChannel')}
                        </button>
                        <button
                            type="button"
                            onClick={() => handleExport('model')}
                            className="flex w-full items-center rounded-lg px-3 py-2 text-xs text-foreground transition hover:bg-muted/50"
                        >
                            {t('exportDimensionModel')}
                        </button>
                        <button
                            type="button"
                            onClick={() => handleExport('apikey')}
                            className="flex w-full items-center rounded-lg px-3 py-2 text-xs text-foreground transition hover:bg-muted/50"
                        >
                            {t('exportDimensionApiKey')}
                        </button>
                    </PopoverContent>
                </Popover>
            </div>
            <Total />
            <div data-testid="home-main-grid" className="space-y-5 md:space-y-6">
                <Activity />
                <StatsChart />
                <Rank />
                <TokenBreakdown />
            </div>
        </PageWrapper>
    );
}
