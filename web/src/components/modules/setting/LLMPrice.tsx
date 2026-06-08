'use client';

import { useEffect, useState, useRef } from 'react';
import { useTranslations } from 'next-intl';
import { DollarSign, Clock, RefreshCw } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { useSettingList, useSetSetting, SettingKey } from '@/api/endpoints/setting';
import { useUpdateModelPrice, useLastUpdateTime } from '@/api/endpoints/model';
import { toast } from '@/components/common/Toast';

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
        <div data-testid="setting-llm-price-card" className="octo-setting-card">
            <div className="space-y-2">
                <h2 className="octo-setting-heading">
                    <DollarSign className="size-4" />
                    {t('llmPrice.title')}
                </h2>

                <div className="grid grid-cols-1 gap-2 md:grid-cols-3">
                    {scopeCards.map((card) => (
                        <div key={card.label} className="octo-stat-card">
                            <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                                <span>{card.label}</span>
                            </div>
                            <div className="mt-1 text-sm font-semibold text-card-foreground">{card.value}</div>
                        </div>
                    ))}
                </div>
            </div>

            {/* 更新间隔 */}
            <div className="octo-setting-row rounded-2xl border border-border/60 bg-background/60 p-3">
                <div className="octo-setting-label">
                    <Clock className="size-4 text-muted-foreground" />
                    <span>{t('llmPrice.updateInterval.label')}</span>
                </div>
                <Input
                    type="number"
                    value={updateInterval}
                    onChange={(e) => setUpdateInterval(e.target.value)}
                    onBlur={() => handleSave(SettingKey.ModelInfoUpdateInterval, updateInterval, initialUpdateInterval.current)}
                    placeholder={t('llmPrice.updateInterval.placeholder')}
                    className="h-9 rounded-xl"
                />
            </div>

            {/* 手动更新 */}
            <div className="flex flex-col gap-3 rounded-2xl border border-border/60 bg-background/60 p-3 md:flex-row md:items-center md:justify-between">
                <div className="flex flex-col gap-1">
                    <div className="flex items-center gap-3">
                        <RefreshCw className="size-4 text-muted-foreground" />
                        <span className="text-sm font-medium">{t('llmPrice.manualUpdate.label')}</span>
                    </div>
                    <span className="ml-7 text-xs text-muted-foreground">
                        {t('llmPrice.lastUpdate')}: {formatLastUpdateTime(lastUpdateTime)}
                    </span>
                </div>
                <Button
                    variant="outline"
                    size="sm"
                    onClick={handleManualUpdate}
                    disabled={updatePrice.isPending}
                    className="h-9 rounded-xl"
                >
                    {updatePrice.isPending ? t('llmPrice.manualUpdate.updating') : t('llmPrice.manualUpdate.button')}
                </Button>
            </div>
        </div>
    );
}

