'use client';

import { useCallback, useMemo, useState } from 'react';
import { useTranslations } from 'next-intl';
import { Loader2, RotateCw } from 'lucide-react';
import { AUDIT_ACTIONS, AUDIT_RESOURCE_TYPES, useAuditList, type AuditLog } from '@/api/endpoints/audit';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import { EmptyState } from '@/components/common/EmptyState';
import { AuditDetailDialog } from './detail';
import { cn } from '@/lib/utils';

const PAGE_SIZE = 20;

function formatDateTimeLocal(value?: number) {
    if (!value) return '';
    const date = new Date(value);
    const pad = (n: number) => String(n).padStart(2, '0');
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function parseDateTimeLocal(value: string): number | undefined {
    if (!value) return undefined;
    const ts = new Date(value).getTime();
    return Number.isFinite(ts) ? ts : undefined;
}

function formatTime(value: string): string {
    if (!value) return '-';
    const d = new Date(value);
    if (Number.isNaN(d.getTime())) return value;
    return d.toLocaleString();
}

function ActionTag({ action }: { action: string }) {
    const tone =
        action === 'create' ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
            : action === 'delete' ? 'bg-rose-500/10 text-rose-600 dark:text-rose-400'
                : action === 'update' ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400'
                    : action === 'login' ? 'bg-sky-500/10 text-sky-600 dark:text-sky-400'
                        : 'bg-muted text-muted-foreground';
    return (
        <span className={cn('inline-flex items-center rounded-md px-1.5 py-0.5 text-xs font-medium', tone)}>
            {action}
        </span>
    );
}

export function Audit() {
    const t = useTranslations('audit');
    const tCommon = useTranslations('common.state');

    const [page, setPage] = useState(1);
    const [actionFilter, setActionFilter] = useState<string>('');
    const [resourceTypeFilter, setResourceTypeFilter] = useState<string>('');
    const [startTimeStr, setStartTimeStr] = useState('');
    const [endTimeStr, setEndTimeStr] = useState('');
    const [appliedFilters, setAppliedFilters] = useState<{
        start_time?: number;
        end_time?: number;
        action?: string;
        resource_type?: string;
    }>({});
    const [selectedId, setSelectedId] = useState<number | null>(null);
    const [selectedPreview, setSelectedPreview] = useState<AuditLog | null>(null);

    const queryParams = useMemo(() => ({
        page,
        page_size: PAGE_SIZE,
        ...appliedFilters,
    }), [page, appliedFilters]);

    const { data, isLoading, isError, error, refetch, isFetching } = useAuditList(queryParams);

    const list = data?.list ?? [];
    const total = data?.total ?? 0;
    const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

    const handleApplyFilters = useCallback(() => {
        const start = parseDateTimeLocal(startTimeStr);
        const end = parseDateTimeLocal(endTimeStr);
        setAppliedFilters({
            start_time: start,
            end_time: end,
            action: actionFilter || undefined,
            resource_type: resourceTypeFilter || undefined,
        });
        setPage(1);
    }, [startTimeStr, endTimeStr, actionFilter, resourceTypeFilter]);

    const handleReset = useCallback(() => {
        setStartTimeStr('');
        setEndTimeStr('');
        setActionFilter('');
        setResourceTypeFilter('');
        setAppliedFilters({});
        setPage(1);
    }, []);

    const handleRowClick = useCallback((log: AuditLog) => {
        setSelectedPreview(log);
        setSelectedId(log.id);
    }, []);

    const handleClose = useCallback(() => {
        setSelectedId(null);
        setSelectedPreview(null);
    }, []);

    return (
        <div data-testid="audit-page" className="flex h-full min-h-0 flex-col gap-3">
            <div className="shrink-0 rounded-2xl border border-border/60 bg-card/80 p-3">
                <div className="flex flex-wrap items-end gap-2">
                    <label className="grid gap-1 text-xs text-muted-foreground">
                        <span>{t('timeRange')}</span>
                        <Input
                            type="datetime-local"
                            value={startTimeStr}
                            onChange={(e) => setStartTimeStr(e.target.value)}
                            className="h-8 rounded-xl text-sm"
                        />
                    </label>
                    <label className="grid gap-1 text-xs text-muted-foreground">
                        <span>—</span>
                        <Input
                            type="datetime-local"
                            value={endTimeStr}
                            onChange={(e) => setEndTimeStr(e.target.value)}
                            className="h-8 rounded-xl text-sm"
                        />
                    </label>
                    <label className="grid gap-1 text-xs text-muted-foreground">
                        <span>{t('action')}</span>
                        <Select value={actionFilter} onValueChange={setActionFilter}>
                            <SelectTrigger size="sm" className="h-8 w-[10rem] rounded-xl">
                                <SelectValue placeholder={t('allActions')} />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="">{t('allActions')}</SelectItem>
                                {AUDIT_ACTIONS.map((a) => (
                                    <SelectItem key={a} value={a}>{a}</SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </label>
                    <label className="grid gap-1 text-xs text-muted-foreground">
                        <span>{t('resourceType')}</span>
                        <Select value={resourceTypeFilter} onValueChange={setResourceTypeFilter}>
                            <SelectTrigger size="sm" className="h-8 w-[10rem] rounded-xl">
                                <SelectValue placeholder={t('allTypes')} />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="">{t('allTypes')}</SelectItem>
                                {AUDIT_RESOURCE_TYPES.map((r) => (
                                    <SelectItem key={r} value={r}>{r}</SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </label>
                    <div className="ml-auto flex items-center gap-2">
                        <Button type="button" variant="ghost" size="sm" className="rounded-xl" onClick={handleReset} disabled={isFetching}>
                            {t('reset')}
                        </Button>
                        <Button type="button" size="sm" className="rounded-xl" onClick={handleApplyFilters} disabled={isFetching}>
                            {t('filter')}
                        </Button>
                    </div>
                </div>
            </div>

            <div className="min-h-0 flex-1 overflow-auto rounded-2xl border border-border/60 bg-card/60">
                {isLoading && list.length === 0 ? (
                    <div className="flex items-center justify-center py-12">
                        <Loader2 className="size-5 animate-spin text-muted-foreground" />
                    </div>
                ) : isError ? (
                    <div className="flex flex-col items-center justify-center gap-3 py-12">
                        <span className="text-sm text-destructive">{tCommon('error')}</span>
                        <Button type="button" variant="outline" size="sm" className="rounded-xl" onClick={() => refetch()}>
                            <RotateCw className="size-4" />
                            <span className="ml-1">{tCommon('retry')}</span>
                        </Button>
                    </div>
                ) : list.length === 0 ? (
                    <div className="flex items-center justify-center py-12">
                        <EmptyState title={t('empty')} />
                    </div>
                ) : (
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-[10rem]">{t('time')}</TableHead>
                                <TableHead className="w-[8rem]">{t('user')}</TableHead>
                                <TableHead className="w-[6rem]">{t('action')}</TableHead>
                                <TableHead className="w-[6rem]">{t('resourceType')}</TableHead>
                                <TableHead className="w-[8rem]">{t('resourceId')}</TableHead>
                                <TableHead className="w-[8rem]">{t('ip')}</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {list.map((log) => (
                                <TableRow
                                    key={log.id}
                                    className="cursor-pointer"
                                    onClick={() => handleRowClick(log)}
                                    data-testid={`audit-row-${log.id}`}
                                >
                                    <TableCell className="whitespace-nowrap text-xs text-muted-foreground">
                                        {formatTime(log.created_at)}
                                    </TableCell>
                                    <TableCell className="text-sm">{log.username || `#${log.user_id}`}</TableCell>
                                    <TableCell><ActionTag action={log.action} /></TableCell>
                                    <TableCell className="text-sm">{log.resource_type}</TableCell>
                                    <TableCell className="text-sm text-muted-foreground">{log.resource_id || '-'}</TableCell>
                                    <TableCell className="text-xs text-muted-foreground">{log.ip || '-'}</TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                )}
            </div>

            {total > 0 && (
                <div className="flex shrink-0 items-center justify-between rounded-2xl border border-border/60 bg-card/80 px-3 py-2 text-xs text-muted-foreground">
                    <span>{t('total', { count: total })}</span>
                    <div className="flex items-center gap-2">
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            className="rounded-xl"
                            disabled={page <= 1 || isFetching}
                            onClick={() => setPage((p) => Math.max(1, p - 1))}
                        >
                            {t('prev')}
                        </Button>
                        <span>{t('page', { page, pages: totalPages })}</span>
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            className="rounded-xl"
                            disabled={page >= totalPages || isFetching}
                            onClick={() => setPage((p) => p + 1)}
                        >
                            {t('next')}
                        </Button>
                    </div>
                </div>
            )}

            <AuditDetailDialog auditId={selectedId} preview={selectedPreview} onClose={handleClose} />
        </div>
    );
}
