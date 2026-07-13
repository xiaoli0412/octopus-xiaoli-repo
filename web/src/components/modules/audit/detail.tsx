'use client';

import { type ReactNode } from 'react';
import { useTranslations } from 'next-intl';
import { Loader2 } from 'lucide-react';
import { useAuditDetail, type AuditLog } from '@/api/endpoints/audit';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { JsonDiff } from './diff';
import { cn } from '@/lib/utils';

export interface AuditDetailDialogProps {
    /** 选中的审计日志 ID，传 null 关闭。 */
    auditId: number | null;
    /** 列表中已有的记录，用于在详情加载完成前即时展示基础字段。 */
    preview?: AuditLog | null;
    onClose: () => void;
}

function formatTime(value: string): string {
    if (!value) return '-';
    const d = new Date(value);
    if (Number.isNaN(d.getTime())) return value;
    return d.toLocaleString();
}

function InfoField({ label, value }: { label: string; value?: ReactNode }) {
    return (
        <div className="flex flex-col gap-0.5">
            <span className="text-xs text-muted-foreground">{label}</span>
            <span className="break-all text-sm text-foreground">{value || '-'}</span>
        </div>
    );
}

function ActionBadge({ action }: { action: string }) {
    const tone =
        action === 'create' ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
            : action === 'delete' ? 'bg-rose-500/10 text-rose-600 dark:text-rose-400'
                : action === 'update' ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400'
                    : action === 'login' ? 'bg-sky-500/10 text-sky-600 dark:text-sky-400'
                        : 'bg-muted text-muted-foreground';
    return (
        <span className={cn('inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium', tone)}>
            {action}
        </span>
    );
}

export function AuditDetailDialog({ auditId, preview, onClose }: AuditDetailDialogProps) {
    const t = useTranslations('audit');
    const { data, isLoading, error } = useAuditDetail(auditId);
    const record = data ?? preview ?? null;
    const open = auditId !== null;

    return (
        <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
            <DialogContent className="max-h-[88vh] overflow-y-auto sm:max-w-2xl">
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        {t('title')}
                        {record && <span className="text-sm font-normal text-muted-foreground">#{record.id}</span>}
                    </DialogTitle>
                    <DialogDescription className="sr-only">{t('title')}</DialogDescription>
                </DialogHeader>

                {isLoading && !record && (
                    <div className="flex items-center justify-center py-12">
                        <Loader2 className="size-5 animate-spin text-muted-foreground" />
                    </div>
                )}

                {error && !record && (
                    <div className="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive">
                        {error instanceof Error ? error.message : String(error)}
                    </div>
                )}

                {record && (
                    <div className="space-y-4">
                        <div className="grid grid-cols-2 gap-3 rounded-xl border border-border/60 bg-muted/20 p-3 md:grid-cols-3">
                            <InfoField label={t('user')} value={record.username || `#${record.user_id}`} />
                            <InfoField label={t('action')} value={<ActionBadge action={record.action} />} />
                            <InfoField label={t('resourceType')} value={record.resource_type} />
                            <InfoField label={t('resourceId')} value={record.resource_id} />
                            <InfoField label={t('ip')} value={record.ip} />
                            <InfoField label={t('time')} value={formatTime(record.created_at)} />
                            <div className="col-span-2 md:col-span-3">
                                <InfoField label="User-Agent" value={record.user_agent} />
                            </div>
                            {record.resource_name && (
                                <div className="col-span-2 md:col-span-3">
                                    <InfoField label={t('resourceName')} value={record.resource_name} />
                                </div>
                            )}
                        </div>

                        <div>
                            <div className="mb-1.5 text-xs font-medium text-muted-foreground">{t('diff')}</div>
                            <JsonDiff before={record.before_json} after={record.after_json} />
                        </div>

                        <div className="grid gap-3 md:grid-cols-2">
                            <div>
                                <div className="mb-1.5 text-xs font-medium text-muted-foreground">{t('before')}</div>
                                <pre className="max-h-48 overflow-auto rounded-md border border-border/60 bg-muted/30 p-2 text-xs text-foreground/80">
                                    {record.before_json || '(empty)'}
                                </pre>
                            </div>
                            <div>
                                <div className="mb-1.5 text-xs font-medium text-muted-foreground">{t('after')}</div>
                                <pre className="max-h-48 overflow-auto rounded-md border border-border/60 bg-muted/30 p-2 text-xs text-foreground/80">
                                    {record.after_json || '(empty)'}
                                </pre>
                            </div>
                        </div>
                    </div>
                )}
            </DialogContent>
        </Dialog>
    );
}
