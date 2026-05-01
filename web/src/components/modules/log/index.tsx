'use client';

import { useCallback, useMemo, useState } from 'react';
import { exportLogs, useLogs } from '@/api/endpoints/log';
import { LogCard } from './Item';
import { CalendarRange, Download, Loader2, SlidersHorizontal } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { VirtualizedGrid } from '@/components/common/VirtualizedGrid';
import { Button } from '@/components/ui/button';
import { toast } from '@/components/common/Toast';
import { Input } from '@/components/ui/input';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';

type ExportFormat = 'json' | 'jsonl';

function formatDateTimeLocal(value?: number) {
    if (!value) return '';
    const date = new Date(value);
    const pad = (n: number) => String(n).padStart(2, '0');
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function parseDateTimeLocal(value: string) {
    if (!value) return undefined;
    const timestamp = new Date(value).getTime();
    return Number.isFinite(timestamp) ? timestamp : undefined;
}

/**
 * 日志页面组件
 * - 初始加载 pageSize 条历史日志
 * - SSE 实时推送新日志
 * - 滚动自动加载更多
 */
export function Log() {
    const t = useTranslations('log');
    const { logs, hasMore, isLoading, isLoadingMore, loadMore } = useLogs({ pageSize: 10 });
    const [isExporting, setIsExporting] = useState(false);
    const [exportOpen, setExportOpen] = useState(false);
    const [exportFormat, setExportFormat] = useState<ExportFormat>('jsonl');
    const [exportLimit, setExportLimit] = useState('3000');
    const [startTime, setStartTime] = useState('');
    const [endTime, setEndTime] = useState('');

    const canLoadMore = hasMore && !isLoading && !isLoadingMore && logs.length > 0;
    const handleReachEnd = useCallback(() => {
        if (!canLoadMore) return;
        void loadMore();
    }, [canLoadMore, loadMore]);

    const footer = useMemo(() => {
        if (hasMore && (isLoading || isLoadingMore)) {
            return (
                <div className="flex justify-center py-4">
                    <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                </div>
            );
        }
        if (!hasMore && logs.length > 0) {
            return (
                <div className="flex justify-center py-4">
                    <span className="text-sm text-muted-foreground">{t('list.noMore')}</span>
                </div>
            );
        }
        return null;
    }, [hasMore, isLoading, isLoadingMore, logs.length, t]);

    const exportSummary = useMemo(() => {
        const limitLabel = exportLimit.trim() ? t('list.summaryLimit', { count: exportLimit.trim() }) : t('list.summaryAll');
        if (!startTime && !endTime) {
            return `${exportFormat.toUpperCase()} / ${limitLabel}`;
        }

        if (startTime && endTime) {
            return t('list.summaryRange', { start: startTime.replace('T', ' '), end: endTime.replace('T', ' ') });
        }

        return startTime
            ? t('list.summaryFrom', { start: startTime.replace('T', ' ') })
            : t('list.summaryTo', { end: endTime.replace('T', ' ') });
    }, [endTime, exportFormat, exportLimit, startTime, t]);

    const invalidRange = useMemo(() => {
        const parsedStart = parseDateTimeLocal(startTime);
        const parsedEnd = parseDateTimeLocal(endTime);
        return parsedStart !== undefined && parsedEnd !== undefined && parsedStart > parsedEnd;
    }, [endTime, startTime]);

    const handleExport = useCallback(async () => {
        const parsedStart = parseDateTimeLocal(startTime);
        const parsedEnd = parseDateTimeLocal(endTime);
        if (invalidRange) {
            toast.error(t('list.invalidRange'));
            return;
        }

        setIsExporting(true);
        try {
            const parsedLimit = Number.parseInt(exportLimit.trim(), 10);
            const { blob, fileName } = await exportLogs({
                format: exportFormat,
                limit: Number.isFinite(parsedLimit) && parsedLimit > 0 ? parsedLimit : undefined,
                start_time: parsedStart,
                end_time: parsedEnd,
            });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = fileName;
            document.body.appendChild(a);
            a.click();
            a.remove();
            URL.revokeObjectURL(url);
            toast.success(t('list.exportSuccess'));
            setExportOpen(false);
        } catch (error) {
            toast.error(t('list.exportFailed'), {
                description: error instanceof Error ? error.message : undefined,
            });
        } finally {
            setIsExporting(false);
        }
    }, [endTime, exportFormat, exportLimit, invalidRange, startTime, t]);

    const handleQuickRange = useCallback((hours: number) => {
        const now = Date.now();
        setEndTime(formatDateTimeLocal(now));
        setStartTime(formatDateTimeLocal(now - hours * 60 * 60 * 1000));
    }, []);

    const handleResetExport = useCallback(() => {
        setExportFormat('jsonl');
        setExportLimit('3000');
        setStartTime('');
        setEndTime('');
    }, []);

    return (
        <div className="h-full min-h-0 flex flex-col gap-3">
            <div className="flex items-center justify-end px-1">
                <Popover open={exportOpen} onOpenChange={setExportOpen}>
                    <PopoverTrigger asChild>
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            className="rounded-xl"
                            disabled={isExporting}
                        >
                            {isExporting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Download className="h-4 w-4" />}
                            <span className="ml-1">{t('list.export')}</span>
                        </Button>
                    </PopoverTrigger>
                    <PopoverContent align="end" sideOffset={10} className="w-[min(92vw,360px)] rounded-2xl border border-border/60 bg-card p-4 shadow-xl">
                        <div className="space-y-4">
                            <div>
                                <div className="text-sm font-semibold text-foreground">{t('list.exportTitle')}</div>
                                <div className="mt-1 text-xs text-muted-foreground">{exportSummary}</div>
                            </div>

                            <div className="space-y-2">
                                <div className="text-xs font-medium text-muted-foreground">{t('list.exportFormatLabel')}</div>
                                <div className="grid grid-cols-2 gap-2">
                                    {(['jsonl', 'json'] as const).map((format) => (
                                        <button
                                            key={format}
                                            type="button"
                                            onClick={() => setExportFormat(format)}
                                            className={`rounded-xl border px-3 py-2 text-sm transition-colors ${exportFormat === format ? 'border-primary bg-primary/10 text-foreground' : 'border-border bg-muted/20 text-muted-foreground hover:bg-muted/40 hover:text-foreground'}`}
                                        >
                                            {format.toUpperCase()}
                                        </button>
                                    ))}
                                </div>
                            </div>

                            <div className="space-y-2">
                                <div className="text-xs font-medium text-muted-foreground">{t('list.exportRangeLabel')}</div>
                                <div className="grid grid-cols-2 gap-2">
                                    <button type="button" onClick={() => handleQuickRange(24)} className="rounded-xl border border-border bg-muted/20 px-3 py-2 text-sm text-foreground transition-colors hover:bg-muted/40">{t('list.last24Hours')}</button>
                                    <button type="button" onClick={() => handleQuickRange(24 * 7)} className="rounded-xl border border-border bg-muted/20 px-3 py-2 text-sm text-foreground transition-colors hover:bg-muted/40">{t('list.last7Days')}</button>
                                </div>

                                <div className="grid gap-2">
                                    <label className="grid gap-1 text-xs text-muted-foreground">
                                        <span>{t('list.startTime')}</span>
                                        <Input type="datetime-local" value={startTime} onChange={(e) => setStartTime(e.target.value)} className="rounded-xl text-sm" />
                                    </label>
                                    <label className="grid gap-1 text-xs text-muted-foreground">
                                        <span>{t('list.endTime')}</span>
                                        <Input type="datetime-local" value={endTime} onChange={(e) => setEndTime(e.target.value)} className="rounded-xl text-sm" />
                                    </label>
                                </div>
                                {invalidRange && (
                                    <div className="text-xs text-destructive">{t('list.invalidRange')}</div>
                                )}
                            </div>

                            <label className="grid gap-1 text-xs text-muted-foreground">
                                <span>{t('list.exportLimitLabel')}</span>
                                <Input
                                    inputMode="numeric"
                                    value={exportLimit}
                                    onChange={(e) => setExportLimit(e.target.value.replace(/[^\d]/g, ''))}
                                    placeholder={t('list.exportLimitPlaceholder')}
                                    className="rounded-xl text-sm"
                                />
                            </label>

                            <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                                <Button type="button" variant="ghost" size="sm" className="rounded-xl" onClick={handleResetExport} disabled={isExporting}>
                                    <SlidersHorizontal className="h-4 w-4" />
                                    <span className="ml-1">{t('list.reset')}</span>
                                </Button>
                                <Button type="button" size="sm" className="rounded-xl" disabled={isExporting || invalidRange} onClick={() => void handleExport()}>
                                    {isExporting ? <Loader2 className="h-4 w-4 animate-spin" /> : <CalendarRange className="h-4 w-4" />}
                                    <span className="ml-1">{t('list.confirmExport')}</span>
                                </Button>
                            </div>
                        </div>
                    </PopoverContent>
                </Popover>
            </div>

            <VirtualizedGrid
                items={logs}
                layout="list"
                columns={{ default: 1 }}
                estimateItemHeight={80}
                overscan={8}
                getItemKey={(log) => `log-${log.id}`}
                renderItem={(log) => <LogCard log={log} />}
                footer={footer}
                onReachEnd={handleReachEnd}
                reachEndEnabled={canLoadMore}
                reachEndOffset={2}
            />
        </div>
    );
}
