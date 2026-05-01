'use client';

import { memo, useCallback, useEffect, useId, useMemo, useRef, useState } from 'react';
import { Pencil, Trash2, ArrowDownToLine, ArrowUpFromLine } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';
import { useTranslations } from 'next-intl';
import { useUpdateModel, useDeleteModel, type LLMInfo } from '@/api/endpoints/model';
import { getModelIcon } from '@/lib/model-icons';
import { getBillingModeKey } from '@/lib/ui-labels';
import { toast } from '@/components/common/Toast';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/animate-ui/components/animate/tooltip';
import { ModelDeleteOverlay, ModelEditOverlay } from './ItemOverlays';
import { cn } from '@/lib/utils';
import { createPortal } from 'react-dom';

interface ModelItemProps {
    model: LLMInfo;
    layout: 'grid' | 'list';
    index: number;
}

export const ModelItem = memo(function ModelItem({ model, layout, index }: ModelItemProps) {
    const t = useTranslations('model');
    const tSetting = useTranslations('setting.llmRouteTarget');
    const [isEditOpen, setIsEditOpen] = useState(false);
    const [confirmDelete, setConfirmDelete] = useState(false);
    const [overlayRect, setOverlayRect] = useState<{ top: number; left: number; width: number } | null>(null);
    const instanceId = useId();
    const editLayoutId = `edit-btn-${model.name}-${instanceId}`;
    const deleteLayoutId = `delete-btn-${model.name}-${instanceId}`;
    const cardRef = useRef<HTMLElement | null>(null);
    const editButtonRef = useRef<HTMLButtonElement | null>(null);
    const editOverlayRef = useRef<HTMLDivElement | null>(null);
    const [editValues, setEditValues] = useState(() => ({
        input: model.input.toString(),
        output: model.output.toString(),
        cache_read: model.cache_read.toString(),
        cache_write: model.cache_write.toString(),
    }));

    const updateModel = useUpdateModel();
    const deleteModel = useDeleteModel();

    const { Avatar: ModelAvatar, color: brandColor } = useMemo(() => getModelIcon(model.name), [model.name]);

    const formatPrice = (value?: number) => (typeof value === 'number' ? value.toFixed(2) : '-');
    const metaItems = [
        {
            key: 'canonical',
            label: tSetting('canonicalName'),
            value: model.canonical_name ?? '-',
        },
        {
            key: 'billing',
            label: tSetting('billingMode'),
            value: tSetting(`billingModeOptions.${getBillingModeKey(model.billing_mode)}`),
        },
        {
            key: 'official',
            label: tSetting('officialPricePair'),
            value: model.official_input || model.official_output || model.official_cache_read || model.official_cache_write
                ? `${formatPrice(model.official_input)}/${formatPrice(model.official_output)} / ${formatPrice(model.official_cache_read)}/${formatPrice(model.official_cache_write)}`
                : '-'
        },
    ];

    const updateOverlayRect = useCallback(() => {
        const card = cardRef.current;
        if (!card) return;
        const rect = card.getBoundingClientRect();
        setOverlayRect((prev) => {
            if (prev && prev.top === rect.top && prev.left === rect.left && prev.width === rect.width) {
                return prev;
            }
            return { top: rect.top, left: rect.left, width: rect.width };
        });
    }, []);

    const closeEdit = useCallback(() => {
        setIsEditOpen(false);
    }, []);

    const handleEditClick = () => {
        setConfirmDelete(false);
        setEditValues({
            input: model.input.toString(),
            output: model.output.toString(),
            cache_read: model.cache_read.toString(),
            cache_write: model.cache_write.toString(),
        });
        // Ensure first open already has anchor geometry so layout animation can run.
        updateOverlayRect();
        setIsEditOpen(true);
    };

    const handleCancelEdit = () => {
        closeEdit();
    };

    const handleSaveEdit = () => {
        updateModel.mutate({
            name: model.name,
            input: parseFloat(editValues.input) || 0,
            output: parseFloat(editValues.output) || 0,
            cache_read: parseFloat(editValues.cache_read) || 0,
            cache_write: parseFloat(editValues.cache_write) || 0,
        }, {
            onSuccess: () => {
                closeEdit();
                toast.success(t('toast.updated'));
            },
            onError: (error) => {
                toast.error(t('toast.updateFailed'), { description: error.message });
            }
        });
    };

    const handleDeleteClick = () => {
        closeEdit();
        setConfirmDelete(true);
    };
    const handleCancelDelete = () => setConfirmDelete(false);
    const handleConfirmDelete = () => {
        deleteModel.mutate(model.name, {
            onSuccess: () => {
                setConfirmDelete(false);
                toast.success(t('toast.deleted'));
            },
            onError: (error) => {
                setConfirmDelete(false);
                toast.error(t('toast.deleteFailed'), { description: error.message });
            }
        });
    };

    useEffect(() => {
        if (!isEditOpen) return;

        const handlePointerDown = (event: PointerEvent) => {
            const target = event.target as Node | null;
            if (!target) return;
            if (editOverlayRef.current?.contains(target)) return;
            if (editButtonRef.current?.contains(target)) return;
            closeEdit();
        };

        const handleKeyDown = (event: KeyboardEvent) => {
            if (event.key === 'Escape') closeEdit();
        };

        updateOverlayRect();
        window.addEventListener('resize', updateOverlayRect);
        window.addEventListener('scroll', updateOverlayRect, true);
        document.addEventListener('pointerdown', handlePointerDown);
        document.addEventListener('keydown', handleKeyDown);

        return () => {
            window.removeEventListener('resize', updateOverlayRect);
            window.removeEventListener('scroll', updateOverlayRect, true);
            document.removeEventListener('pointerdown', handlePointerDown);
            document.removeEventListener('keydown', handleKeyDown);
        };
    }, [isEditOpen, updateOverlayRect, closeEdit]);

    const shouldRenderEditPortal = isEditOpen || overlayRect !== null;

    return (
        <article
            ref={cardRef}
            data-testid={typeof index === 'number' ? `model-card-${index}` : undefined}
            data-model-name={model.name}
            data-layout={layout}
            data-slot="model-card"
            className={cn(
                'group relative min-h-[7.25rem] rounded-3xl border border-border bg-card transition-all duration-300 flex items-center gap-3 px-3.5 py-3.5',
                (isEditOpen || confirmDelete) && 'z-50'
            )}
        >
            <ModelAvatar size={layout === 'grid' ? 46 : 52} />

            <div className="flex-1 min-w-0 flex flex-col justify-center gap-1.5">
                <Tooltip side="top" sideOffset={10} align="start">
                    <TooltipTrigger className='text-[15px] font-semibold text-card-foreground leading-tight truncate'>
                        {model.name}
                    </TooltipTrigger>
                    <TooltipContent key={model.name}>
                        {model.name}
                    </TooltipContent>
                </Tooltip>

                <p className="flex items-center gap-1.5 text-[13px] text-muted-foreground">
                    <ArrowDownToLine className="size-3.5" style={{ color: brandColor }} />
                    {t('card.inputCache')}
                    <span className="tabular-nums">{model.input.toFixed(2)}/{model.cache_read.toFixed(2)}$</span>
                </p>

                <p className="flex items-center gap-1.5 text-[13px] text-muted-foreground">
                    <ArrowUpFromLine className="size-3.5" style={{ color: brandColor }} />
                    {t('card.outputCache')}
                    <span className="tabular-nums">{model.output.toFixed(2)}/{model.cache_write.toFixed(2)}$</span>
                </p>

                {layout === 'grid' ? (
                    <div data-slot="model-card-meta" className="flex flex-wrap gap-1.5 pt-0.5 text-[10px] text-muted-foreground">
                        {metaItems.map((item) => (
                            <span key={item.key} className="rounded-full border border-border/60 bg-muted/30 px-1.5 py-0.5">
                                {item.label}: {item.value}
                            </span>
                        ))}
                    </div>
                ) : null}

                {layout === 'list' ? (
                    <div data-slot="model-card-meta-compact" className="flex flex-wrap gap-2 pt-1 text-[11px] text-muted-foreground">
                        {metaItems.map((item) => (
                            <span key={item.key} className="rounded-full border border-border/60 bg-muted/30 px-2 py-0.5">
                                {item.label}: {item.value}
                            </span>
                        ))}
                    </div>
                ) : null}
            </div>

            <div
                className={cn(
                    'shrink-0 flex flex-col justify-between self-stretch gap-2',
                    (isEditOpen || confirmDelete) && 'invisible pointer-events-none'
                )}
            >
                <motion.button
                    ref={editButtonRef}
                    layoutId={editLayoutId}
                    type="button"
                    onClick={handleEditClick}
                    disabled={isEditOpen || confirmDelete}
                    className="h-8 w-8 flex items-center justify-center rounded-lg bg-muted/60 text-muted-foreground transition-colors hover:bg-muted disabled:opacity-50"
                    title={t('card.edit')}
                >
                    <Pencil className="size-4" />
                </motion.button>

                <motion.button
                    layoutId={deleteLayoutId}
                    type="button"
                    onClick={handleDeleteClick}
                    disabled={isEditOpen || confirmDelete}
                    className="h-8 w-8 flex items-center justify-center rounded-lg bg-destructive/10 text-destructive transition-colors hover:bg-destructive hover:text-destructive-foreground disabled:opacity-50"
                    title={t('card.delete')}
                >
                    <Trash2 className="size-4" />
                </motion.button>
            </div>

            <AnimatePresence>
                {confirmDelete && (
                    <ModelDeleteOverlay
                        layoutId={deleteLayoutId}
                        isPending={deleteModel.isPending}
                        onCancel={handleCancelDelete}
                        onConfirm={handleConfirmDelete}
                    />
                )}
            </AnimatePresence>

            {shouldRenderEditPortal && typeof document !== 'undefined'
                ? createPortal(
                    <AnimatePresence onExitComplete={() => setOverlayRect(null)}>
                        {isEditOpen && overlayRect && (
                            <div
                                ref={editOverlayRef}
                                className="fixed z-[90]"
                                style={{
                                    top: `${overlayRect.top}px`,
                                    left: `${overlayRect.left}px`,
                                    width: `${overlayRect.width}px`,
                                }}
                            >
                                <div className="relative">
                                    <ModelEditOverlay
                                        layoutId={editLayoutId}
                                        modelName={model.name}
                                        brandColor={brandColor}
                                        editValues={editValues}
                                        isPending={updateModel.isPending}
                                        onChange={setEditValues}
                                        onCancel={handleCancelEdit}
                                        onSave={handleSaveEdit}
                                    />
                                </div>
                            </div>
                        )}
                    </AnimatePresence>,
                    document.body
                )
                : null}
        </article>
    );
});
