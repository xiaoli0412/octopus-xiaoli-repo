'use client';

import { memo, useCallback, useEffect, useId, useMemo, useRef, useState } from 'react';
import { Pencil, Trash2, ArrowDownToLine, ArrowUpFromLine, Power, PowerOff } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';
import { useTranslations } from 'next-intl';
import { useUpdateModel, useDeleteModel, useDisableModel, useEnableModel, type LLMInfo, type UpstreamPriceSummary } from '@/api/endpoints/model';
import { getModelIcon } from '@/lib/model-icons';
import { getBillingModeKey } from '@/lib/ui-labels';
import { toast } from '@/components/common/Toast';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/animate-ui/components/animate/tooltip';
import { ModelDeleteOverlay, ModelEditOverlay } from './ItemOverlays';
import { cn } from '@/lib/utils';
import { createPortal } from 'react-dom';
import type { ModelCardDensity } from '@/components/modules/toolbar/view-options-store';

interface ModelItemProps {
    model: LLMInfo;
    upstreamPrice?: UpstreamPriceSummary;
    density: ModelCardDensity;
    index: number;
}

interface PriceRowProps {
    isCompact: boolean;
    label: string;
    labelClass: string;
    input?: number;
    output?: number;
    cacheRead?: number;
    cacheWrite?: number;
    brandColor: string;
}

function PriceRow({ isCompact, label, labelClass, input, output, cacheRead, cacheWrite, brandColor }: PriceRowProps) {
    const t = useTranslations('model');
    const formatPrice = (value?: number) => (typeof value === 'number' ? value.toFixed(2) : '-');
    const values = [input, output, cacheRead, cacheWrite];
    const isEmpty = values.every((v) => typeof v !== 'number' || v === 0);

    return (
        <div className={cn('flex min-w-0 items-center gap-1.5', isCompact ? 'text-[11px]' : 'text-[13px]')}>
            <span className={cn('shrink-0 font-medium', labelClass)}>{label}</span>
            {isEmpty ? (
                <span className="text-muted-foreground">-</span>
            ) : (
                <span className="min-w-0 truncate tabular-nums text-muted-foreground">
                    <span className="inline-flex items-center gap-1" style={{ color: brandColor }}>
                        <ArrowDownToLine className={cn(isCompact ? 'size-3' : 'size-3.5')} />
                        {t('card.inputCache')}
                    </span>
                    <span className="ml-0.5">{formatPrice(input)}/{formatPrice(cacheRead)}$</span>
                    <span className="mx-1 text-border">·</span>
                    <span className="inline-flex items-center gap-1" style={{ color: brandColor }}>
                        <ArrowUpFromLine className={cn(isCompact ? 'size-3' : 'size-3.5')} />
                        {t('card.outputCache')}
                    </span>
                    <span className="ml-0.5">{formatPrice(output)}/{formatPrice(cacheWrite)}$</span>
                </span>
            )}
        </div>
    );
}

export const ModelItem = memo(function ModelItem({ model, upstreamPrice, density, index }: ModelItemProps) {
    const t = useTranslations('model');
    const tSetting = useTranslations('setting.llmRouteTarget');
    const isCompact = density === 'compact';
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
    const disableModel = useDisableModel();
    const enableModel = useEnableModel();

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
    ];
    const gatewayPrices = upstreamPrice?.gateway_prices ?? [];
    const effectiveGateway = upstreamPrice?.effective_gateway;
    const gatewayPrice = effectiveGateway ?? model;
    const gatewaySourceLabel = effectiveGateway?.source_label;
    const extraGatewayCount = gatewayPrices.length > 1 ? gatewayPrices.length - 1 : 0;

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

    const handleToggleDisable = () => {
        if (model.disabled) {
            enableModel.mutate(model.name, {
                onSuccess: () => toast.success(t('toast.enabled')),
                onError: (error) => toast.error(t('toast.enableFailed'), { description: error.message }),
            });
        } else {
            disableModel.mutate(model.name, {
                onSuccess: () => toast.success(t('toast.disabled')),
                onError: (error) => toast.error(t('toast.disableFailed'), { description: error.message }),
            });
        }
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
            data-layout="grid"
            data-density={density}
            data-slot="model-card"
            className={cn(
                'group relative flex items-center border border-border bg-card transition-all duration-300',
                isCompact
                    ? 'min-h-[6.25rem] gap-2.5 rounded-[1.6rem] px-3 py-3'
                    : 'min-h-[7.25rem] gap-3 rounded-3xl px-3.5 py-3.5',
                (isEditOpen || confirmDelete) && 'z-50',
                model.disabled && 'opacity-70'
            )}
        >
            <ModelAvatar size={isCompact ? 40 : 46} />

            <div className={cn('flex-1 min-w-0 flex flex-col justify-center', isCompact ? 'gap-1' : 'gap-1.5')}>
                <Tooltip side="top" sideOffset={10} align="start">
                    <TooltipTrigger className={cn('font-semibold text-card-foreground leading-tight truncate text-left', isCompact ? 'text-[13px]' : 'text-[15px]')}>
                        {model.name}
                    </TooltipTrigger>
                    <TooltipContent key={model.name}>
                        {model.name}
                    </TooltipContent>
                </Tooltip>

                <PriceRow
                    isCompact={isCompact}
                    label={t('card.official')}
                    labelClass="text-muted-foreground"
                    input={model.official_input}
                    output={model.official_output}
                    cacheRead={model.official_cache_read}
                    cacheWrite={model.official_cache_write}
                    brandColor={brandColor}
                />

                <div className="flex items-center gap-1.5 min-w-0">
                    <PriceRow
                        isCompact={isCompact}
                        label={t('card.gateway')}
                        labelClass="text-cyan-700 dark:text-cyan-300"
                        input={gatewayPrice.input}
                        output={gatewayPrice.output}
                        cacheRead={gatewayPrice.cache_read}
                        cacheWrite={gatewayPrice.cache_write}
                        brandColor={brandColor}
                    />
                    {gatewaySourceLabel ? (
                        <span
                            className={cn(
                                'shrink-0 truncate rounded-full border border-cyan-500/25 bg-cyan-500/10 px-1.5 py-0.5 text-cyan-700 dark:text-cyan-300',
                                isCompact ? 'max-w-[4.5rem] text-[9px]' : 'max-w-[6rem] text-[10px]'
                            )}
                            title={gatewaySourceLabel}
                        >
                            {gatewaySourceLabel}
                        </span>
                    ) : null}
                    {extraGatewayCount > 0 ? (
                        <Tooltip side="top" align="center">
                            <TooltipTrigger asChild>
                                <span
                                    className={cn(
                                        'shrink-0 cursor-help rounded-full border border-border/60 bg-muted/25 px-1.5 py-0.5 text-muted-foreground',
                                        isCompact ? 'text-[9px]' : 'text-[10px]'
                                    )}
                                >
                                    +{extraGatewayCount}
                                </span>
                            </TooltipTrigger>
                            <TooltipContent className="max-w-xs">
                                <div className="grid gap-1">
                                    {gatewayPrices.slice(1).map((item) => (
                                        <div key={item.id} className="text-xs">
                                            <span className="text-muted-foreground">{item.source_label || `#${item.upstream_site_id}`}:</span>{' '}
                                            <span className="tabular-nums">{formatPrice(item.input)}/{formatPrice(item.cache_read)} · {formatPrice(item.output)}/{formatPrice(item.cache_write)}$</span>
                                        </div>
                                    ))}
                                </div>
                            </TooltipContent>
                        </Tooltip>
                    ) : null}
                </div>

                <div
                    data-slot={isCompact ? 'model-card-meta-compact' : 'model-card-meta'}
                    className={cn(
                        'flex flex-wrap text-muted-foreground',
                        isCompact ? 'gap-1 pt-0.5 text-[9px]' : 'gap-1.5 pt-0.5 text-[10px]'
                    )}
                >
                    {metaItems.map((item) => (
                        <span
                            key={item.key}
                            className={cn(
                                'rounded-full border border-border/60 bg-muted/30',
                                isCompact ? 'px-1.5 py-0.5' : 'px-1.5 py-0.5'
                            )}
                        >
                            {item.label}: {item.value}
                        </span>
                    ))}
                </div>
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
                    className={cn(
                        'flex items-center justify-center bg-muted/60 text-muted-foreground transition-colors hover:bg-muted disabled:opacity-50',
                        isCompact ? 'h-7 w-7 rounded-lg' : 'h-8 w-8 rounded-lg'
                    )}
                    title={t('card.edit')}
                >
                    <Pencil className={cn(isCompact ? 'size-3.5' : 'size-4')} />
                </motion.button>

                <motion.button
                    type="button"
                    onClick={handleToggleDisable}
                    disabled={isEditOpen || confirmDelete || disableModel.isPending || enableModel.isPending}
                    className={cn(
                        'flex items-center justify-center transition-colors disabled:opacity-50',
                        model.disabled
                            ? 'bg-emerald-500/10 text-emerald-600 hover:bg-emerald-500 hover:text-white'
                            : 'bg-orange-500/10 text-orange-600 hover:bg-orange-500 hover:text-white',
                        isCompact ? 'h-7 w-7 rounded-lg' : 'h-8 w-8 rounded-lg'
                    )}
                    title={model.disabled ? t('card.enable') : t('card.disable')}
                >
                    <Power className={cn(isCompact ? 'size-3.5' : 'size-4')} />
                </motion.button>

                <motion.button
                    layoutId={deleteLayoutId}
                    type="button"
                    onClick={handleDeleteClick}
                    disabled={isEditOpen || confirmDelete}
                    className={cn(
                        'flex items-center justify-center bg-destructive/10 text-destructive transition-colors hover:bg-destructive hover:text-destructive-foreground disabled:opacity-50',
                        isCompact ? 'h-7 w-7 rounded-lg' : 'h-8 w-8 rounded-lg'
                    )}
                    title={t('card.delete')}
                >
                    <Trash2 className={cn(isCompact ? 'size-3.5' : 'size-4')} />
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
