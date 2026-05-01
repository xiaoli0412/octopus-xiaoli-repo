'use client';

import { useEffect, useState, useRef } from 'react';
import { useTranslations } from 'next-intl';
import { DollarSign, Clock, RefreshCw } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { useSettingList, useSetSetting, SettingKey } from '@/api/endpoints/setting';
import { useUpdateModelPrice, useLastUpdateTime } from '@/api/endpoints/model';
import { toast } from '@/components/common/Toast';
import { HelpHint } from '@/components/common/HelpHint';

export function SettingLLMPrice() {
    const t = useTranslations('setting');
    const { data: settings } = useSettingList();
    const setSetting = useSetSetting();
    const updatePrice = useUpdateModelPrice();
    const { data: lastUpdateTime } = useLastUpdateTime();

    const [updateInterval, setUpdateInterval] = useState('');
    const initialUpdateInterval = useRef('');

    useEffect(() => {
        if (settings) {
            const interval = settings.find(s => s.key === SettingKey.ModelInfoUpdateInterval);
            if (interval) {
                queueMicrotask(() => setUpdateInterval(interval.value));
                initialUpdateInterval.current = interval.value;
            }
        }
    }, [settings]);

    const handleSave = (key: string, value: string, initialValue: string) => {
        if (value === initialValue) return;

        setSetting.mutate({ key, value }, {
            onSuccess: () => {
                toast.success(t('saved'));
                initialUpdateInterval.current = value;
            }
        });
    };

    const handleManualUpdate = () => {
        updatePrice.mutate(undefined, {
            onSuccess: () => {
                toast.success(t('llmPrice.updateSuccess'));
            },
            onError: () => {
                toast.error(t('llmPrice.updateFailed'));
            }
        });
    };

    const formatLastUpdateTime = (timeStr: string | undefined) => {
        if (!timeStr) return t('llmPrice.neverUpdated');
        const date = new Date(timeStr);
        if (date.getFullYear() === 1) return t('llmPrice.neverUpdated');
        return date.toLocaleString();
    };

    const scopeCards = [
        {
            label: t('llmPrice.scopeCards.pricingLabel'),
            value: t('llmPrice.scopeCards.pricingValue'),
            hint: t('llmPrice.scopeCards.pricingHint'),
        },
        {
            label: t('llmPrice.scopeCards.probeLabel'),
            value: t('llmPrice.scopeCards.probeValue'),
            hint: t('llmPrice.scopeCards.probeHint'),
        },
        {
            label: t('llmPrice.scopeCards.syncLabel'),
            value: formatLastUpdateTime(lastUpdateTime),
            hint: t('llmPrice.scopeCards.syncHint'),
        },
    ];

    return (
        <div data-testid="setting-llm-price-card" className="rounded-3xl border border-border bg-card p-6 space-y-5">
            <div className="space-y-3">
                <h2 className="text-lg font-bold text-card-foreground flex items-center gap-2">
                    <DollarSign className="h-5 w-5" />
                    {t('llmPrice.title')}
                    <HelpHint>{t('llmPrice.hint')}</HelpHint>
                </h2>

                <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
                    {scopeCards.map((card) => (
                        <div key={card.label} className="rounded-2xl border border-border/60 bg-muted/20 px-4 py-3">
                            <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                                <span>{card.label}</span>
                                <HelpHint className="size-3.5">{card.hint}</HelpHint>
                            </div>
                            <div className="mt-1 text-sm font-semibold text-card-foreground">{card.value}</div>
                        </div>
                    ))}
                </div>

                <div className="rounded-2xl border border-border/70 bg-muted/20 px-4 py-3 text-sm text-muted-foreground">
                    <div className="flex items-center justify-between gap-3">
                        <div className="font-medium text-card-foreground">{t('llmPrice.defaultPathTitle')}</div>
                        <HelpHint className="size-3.5">{t('llmPrice.defaultPathDesc')}</HelpHint>
                    </div>
                    <div className="mt-2 text-xs leading-5">{t('llmPrice.probeRedirectTitle')}</div>
                </div>
            </div>

            {/* 更新间隔 */}
            <div className="flex flex-col gap-3 rounded-2xl border border-border/60 bg-background/60 px-4 py-4 md:flex-row md:items-center md:justify-between">
                <div className="flex items-center gap-3">
                    <Clock className="h-5 w-5 text-muted-foreground" />
                    <div className="space-y-0.5">
                        <div className="flex items-center gap-2">
                            <span className="text-sm font-medium">{t('llmPrice.updateInterval.label')}</span>
                            <HelpHint className="size-3.5">{t('llmPrice.updateInterval.hint')}</HelpHint>
                        </div>
                        <p className="text-xs text-muted-foreground">{t('llmPrice.updateInterval.hint')}</p>
                    </div>
                </div>
                <Input
                    type="number"
                    value={updateInterval}
                    onChange={(e) => setUpdateInterval(e.target.value)}
                    onBlur={() => handleSave(SettingKey.ModelInfoUpdateInterval, updateInterval, initialUpdateInterval.current)}
                    placeholder={t('llmPrice.updateInterval.placeholder')}
                    className="w-48 rounded-xl"
                />
            </div>

            {/* 手动更新 */}
            <div className="flex flex-col gap-3 rounded-2xl border border-border/60 bg-background/60 px-4 py-4 md:flex-row md:items-center md:justify-between">
                <div className="flex flex-col gap-1">
                    <div className="flex items-center gap-3">
                        <RefreshCw className="h-5 w-5 text-muted-foreground" />
                        <span className="text-sm font-medium">{t('llmPrice.manualUpdate.label')}</span>
                        <HelpHint className="size-3.5">{t('llmPrice.manualUpdate.hint')}</HelpHint>
                    </div>
                    <span className="text-xs text-muted-foreground ml-8">
                        {t('llmPrice.lastUpdate')}: {formatLastUpdateTime(lastUpdateTime)}
                    </span>
                </div>
                <Button
                    variant="outline"
                    size="sm"
                    onClick={handleManualUpdate}
                    disabled={updatePrice.isPending}
                    className="rounded-xl"
                >
                    {updatePrice.isPending ? t('llmPrice.manualUpdate.updating') : t('llmPrice.manualUpdate.button')}
                </Button>
            </div>
        </div>
    );
}

