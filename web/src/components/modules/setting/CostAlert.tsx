'use client';

import { useCallback, useState } from 'react';
import { useTranslations } from 'next-intl';
import { Bell, Link2, Percent, MessageSquare } from 'lucide-react';
import { Input } from '@/components/ui/input';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { useSettingList, useSetSetting, SettingKey } from '@/api/endpoints/setting';
import { toast } from '@/components/common/Toast';
import { HelpHint } from '@/components/common/HelpHint';

const FORMAT_OPTIONS = [
    { value: 'generic', label: 'Generic JSON' },
    { value: 'slack', label: 'Slack' },
    { value: 'feishu', label: 'Feishu / Lark' },
    { value: 'dingtalk', label: 'DingTalk' },
] as const;

export function SettingCostAlert() {
    const t = useTranslations('setting');
    const { data: settings } = useSettingList();
    const setSetting = useSetSetting();

    const webhookURLSetting = settings?.find((s) => s.key === SettingKey.CostAlertWebhookURL)?.value ?? '';
    const thresholdsSetting = settings?.find((s) => s.key === SettingKey.CostAlertThresholds)?.value ?? '0.5,0.8,1.0';
    const formatSetting = settings?.find((s) => s.key === SettingKey.CostAlertFormat)?.value ?? 'generic';

    const [webhookURLDraft, setWebhookURLDraft] = useState<string | null>(null);
    const [thresholdsDraft, setThresholdsDraft] = useState<string | null>(null);
    const [formatDraft, setFormatDraft] = useState<string | null>(null);

    const webhookURL = webhookURLDraft ?? webhookURLSetting;
    const thresholds = thresholdsDraft ?? thresholdsSetting;
    const format = formatDraft ?? formatSetting;

    const handleSave = useCallback((key: string, value: string, initialValue: string) => {
        if (value === initialValue) return;
        setSetting.mutate({ key, value }, {
            onSuccess: () => {
                toast.success(t('saved'));
                if (key === SettingKey.CostAlertWebhookURL) {
                    setWebhookURLDraft(null);
                } else if (key === SettingKey.CostAlertThresholds) {
                    setThresholdsDraft(null);
                } else if (key === SettingKey.CostAlertFormat) {
                    setFormatDraft(null);
                }
            },
        });
    }, [setSetting, t]);

    const handleFormatChange = useCallback((value: string) => {
        setFormatDraft(value);
        handleSave(SettingKey.CostAlertFormat, value, formatSetting);
    }, [formatSetting, handleSave]);

    return (
        <div data-testid="setting-cost-alert-card" className="octo-setting-card">
            <div className="space-y-2">
                <h2 className="octo-setting-heading">
                    <Bell className="size-4" />
                    {t('costAlert.title')}
                </h2>
                <p className="text-xs leading-5 text-muted-foreground">{t('costAlert.description')}</p>
            </div>

            <div className="grid gap-3">
                {/* Webhook URL */}
                <div className="flex flex-col gap-3 rounded-2xl border border-border/50 bg-card px-4 py-4 md:flex-row md:items-center md:justify-between">
                    <div className="space-y-1.5">
                        <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                            <Link2 className="h-4 w-4 text-muted-foreground" />
                            <span>{t('costAlert.webhookURL.label')}</span>
                            <HelpHint>{t('costAlert.webhookURL.hint')}</HelpHint>
                        </div>
                    </div>
                    <Input
                        aria-label={t('costAlert.webhookURL.label')}
                        type="url"
                        value={webhookURL}
                        onChange={(e) => setWebhookURLDraft(e.target.value)}
                        onBlur={() => handleSave(SettingKey.CostAlertWebhookURL, webhookURL, webhookURLSetting)}
                        placeholder={t('costAlert.webhookURL.placeholder')}
                        className="w-full rounded-xl md:w-72"
                    />
                </div>

                {/* Thresholds */}
                <div className="flex flex-col gap-3 rounded-2xl border border-border/50 bg-card px-4 py-4 md:flex-row md:items-center md:justify-between">
                    <div className="space-y-1.5">
                        <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                            <Percent className="h-4 w-4 text-muted-foreground" />
                            <span>{t('costAlert.thresholds.label')}</span>
                            <HelpHint>{t('costAlert.thresholds.hint')}</HelpHint>
                        </div>
                    </div>
                    <Input
                        aria-label={t('costAlert.thresholds.label')}
                        type="text"
                        value={thresholds}
                        onChange={(e) => setThresholdsDraft(e.target.value)}
                        onBlur={() => handleSave(SettingKey.CostAlertThresholds, thresholds, thresholdsSetting)}
                        placeholder="0.5,0.8,1.0"
                        className="w-full rounded-xl md:w-72"
                    />
                </div>

                {/* Format */}
                <div className="flex flex-col gap-3 rounded-2xl border border-border/50 bg-card px-4 py-4 md:flex-row md:items-center md:justify-between">
                    <div className="space-y-1.5">
                        <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                            <MessageSquare className="h-4 w-4 text-muted-foreground" />
                            <span>{t('costAlert.format.label')}</span>
                            <HelpHint>{t('costAlert.format.hint')}</HelpHint>
                        </div>
                    </div>
                    <Select value={format} onValueChange={handleFormatChange}>
                        <SelectTrigger className="w-full rounded-xl md:w-72" aria-label={t('costAlert.format.label')}>
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            {FORMAT_OPTIONS.map((opt) => (
                                <SelectItem key={opt.value} value={opt.value}>
                                    {opt.label}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>
            </div>
        </div>
    );
}
