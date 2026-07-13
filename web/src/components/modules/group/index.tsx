'use client';

import { useCallback, useMemo, useState } from 'react';
import { GroupCard } from './Card';
import { useGroupList, useBatchDeleteGroup, type BatchOperationResult } from '@/api/endpoints/group';
import { useSearchStore, useToolbarViewOptionsStore } from '@/components/modules/toolbar';
import { VirtualizedGrid } from '@/components/common/VirtualizedGrid';
import { BatchToolbar, type BatchAction } from '@/components/common/BatchToolbar';
import { useTranslations } from 'next-intl';
import { toast } from '@/components/common/Toast';

export function Group() {
    const t = useTranslations('group');
    const { data: groups, isLoading, error, refetch } = useGroupList();
    const pageKey = 'group' as const;
    const searchTerm = useSearchStore((s) => s.getSearchTerm(pageKey));
    const layout = useToolbarViewOptionsStore((s) => s.getLayout(pageKey));
    const sortOrder = useToolbarViewOptionsStore((s) => s.getSortOrder(pageKey));
    const filter = useToolbarViewOptionsStore((s) => s.groupFilter);

    const batchDelete = useBatchDeleteGroup();
    const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());

    const sortedGroups = useMemo(() => {
        if (!groups) return [];
        return [...groups].sort((a, b) => {
            const aId = a.id || 0;
            const bId = b.id || 0;
            return sortOrder === 'asc' ? aId - bId : bId - aId;
        });
    }, [groups, sortOrder]);

    const visibleGroups = useMemo(() => {
        const term = searchTerm.toLowerCase().trim();
        const byName = !term ? sortedGroups : sortedGroups.filter((g) => g.name.toLowerCase().includes(term));

        if (filter === 'with-members') return byName.filter((g) => (g.items?.length || 0) > 0);
        if (filter === 'empty') return byName.filter((g) => (g.items?.length || 0) === 0);

        return byName;
    }, [sortedGroups, searchTerm, filter]);

    const handleToggleSelect = useCallback((id: number) => {
        setSelectedIds((prev) => {
            const next = new Set(prev);
            if (next.has(id)) {
                next.delete(id);
            } else {
                next.add(id);
            }
            return next;
        });
    }, []);

    const handleClearSelection = useCallback(() => {
        setSelectedIds(new Set());
    }, []);

    const handleSelectAll = useCallback(() => {
        setSelectedIds(new Set(visibleGroups.map((g) => g.id).filter((id): id is number => typeof id === 'number')));
    }, [visibleGroups]);

    const handleBatchResult = useCallback((result: BatchOperationResult, successMsg: string) => {
        if (result.failed_count > 0) {
            toast.warning(`${successMsg}：成功 ${result.success_count}，失败 ${result.failed_count}`, {
                description: result.errors.slice(0, 3).join('\n'),
            });
        } else {
            toast.success(`${successMsg}：${result.success_count} 项`);
        }
        setSelectedIds(new Set());
    }, []);

    const handleBatchDelete = useCallback(() => {
        const ids = Array.from(selectedIds);
        batchDelete.mutate(ids, {
            onSuccess: (result) => handleBatchResult(result, t('batch.deleteSuccess')),
            onError: (error) => toast.error(error.message),
        });
    }, [selectedIds, batchDelete, handleBatchResult, t]);

    const selectedIdArray = Array.from(selectedIds);
    const allVisibleSelected = visibleGroups.length > 0 && visibleGroups.every((g) => typeof g.id === 'number' && selectedIds.has(g.id));

    const batchActions: BatchAction[] = [
        {
            label: t('batch.selectAll'),
            onClick: () => (allVisibleSelected ? handleClearSelection() : handleSelectAll()),
        },
        {
            label: t('batch.delete'),
            onClick: handleBatchDelete,
            variant: 'destructive',
            requireConfirm: true,
            confirmText: t('batch.deleteConfirm', { count: selectedIdArray.length }),
        },
    ];

    return (
        <div className="flex h-full min-h-0 flex-col gap-3">
            <div className="flex shrink-0 items-center gap-2 rounded-2xl border border-border/60 bg-card/80 p-1.5">
                <div className="flex-1 px-2 text-xs font-medium text-muted-foreground">{t('listTitle')}</div>
                <BatchToolbar
                    selectedIds={selectedIdArray}
                    onClearSelection={handleClearSelection}
                    actions={batchActions}
                />
            </div>
            <div className="min-h-0 flex-1">
                <VirtualizedGrid
                    items={visibleGroups}
                    layout={layout}
                    columns={{ default: 1, md: 2, lg: 3 }}
                    estimateItemHeight={520}
                    getItemKey={(group, index) => group.id ?? `group-${index}`}
                    renderItem={(group) => (
                        <GroupCard
                            group={group}
                            selected={typeof group.id === 'number' ? selectedIds.has(group.id) : false}
                            onToggleSelect={() => typeof group.id === 'number' && handleToggleSelect(group.id)}
                        />
                    )}
                    isLoading={isLoading}
                    error={error}
                    onRetry={() => refetch()}
                    emptyHint={t('empty')}
                />
            </div>
        </div>
    );
}
