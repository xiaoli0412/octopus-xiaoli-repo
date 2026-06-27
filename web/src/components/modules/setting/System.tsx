'use client';

import { useEffect, useMemo, useState, useRef } from 'react';
import { useTranslations } from 'next-intl';
import { Monitor, Globe, Clock, Shield, Link2 } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { usePublicAccess, useSettingList, useSetSetting, SettingKey } from '@/api/endpoints/setting';
import { cn } from '@/lib/utils';

export function SettingSystem() {
    const t = useTranslations('setting');
    const { data: settings } = useSettingList();
    const { data: publicAccess } = usePublicAccess();
    const setSetting = useSetSetting();

    const [proxyUrl, setProxyUrl] = useState('');
    const [apiBaseUrl, setApiBaseUrl] = useState('');
    const [apiAlternateBaseUrls, setApiAlternateBaseUrls] = useState('');
    const [trustedProxyCIDRs, setTrustedProxyCIDRs] = useState('');
    const [opsIPDisplayMode, setOpsIPDisplayMode] = useState<'masked' | 'full'>('masked');
    const [statsSaveInterval, setStatsSaveInterval] = useState('');
    const [corsAllowOrigins, setCorsAllowOrigins] = useState('');

    const initialProxyUrl = useRef('');
    const initialApiBaseUrl = useRef('');
    const initialApiAlternateBaseUrls = useRef('');
    const initialTrustedProxyCIDRs = useRef('');
    const initialOpsIPDisplayMode = useRef('masked');
    const initialStatsSaveInterval = useRef('');
    const initialCorsAllowOrigins = useRef('');

    useEffect(() => {
        if (settings) {
            const proxy = settings.find(s => s.key === SettingKey.ProxyURL);
            const apiBase = settings.find(s => s.key === SettingKey.ApiBaseUrl);
            const alternate = settings.find(s => s.key === SettingKey.ApiAlternateBaseUrls);
            const trusted = settings.find(s => s.key === SettingKey.TrustedProxyCIDRs);
            const displayMode = settings.find(s => s.key === SettingKey.OpsIPDisplayMode);
            const interval = settings.find(s => s.key === SettingKey.StatsSaveInterval);
            const cors = settings.find(s => s.key === SettingKey.CORSAllowOrigins);
            if (proxy) {
                queueMicrotask(() => setProxyUrl(proxy.value));
                initialProxyUrl.current = proxy.value;
            }
            if (apiBase) {
                queueMicrotask(() => setApiBaseUrl(apiBase.value));
                initialApiBaseUrl.current = apiBase.value;
            }
            if (alternate) {
                queueMicrotask(() => setApiAlternateBaseUrls(alternate.value));
                initialApiAlternateBaseUrls.current = alternate.value;
            }
            if (trusted) {
                queueMicrotask(() => setTrustedProxyCIDRs(trusted.value));
                initialTrustedProxyCIDRs.current = trusted.value;
            }
            if (displayMode) {
                const value = displayMode.value === 'full' ? 'full' : 'masked';
                queueMicrotask(() => setOpsIPDisplayMode(value));
                initialOpsIPDisplayMode.current = value;
            }
            if (interval) {
                queueMicrotask(() => setStatsSaveInterval(interval.value));
                initialStatsSaveInterval.current = interval.value;
            }
            if (cors) {
                queueMicrotask(() => setCorsAllowOrigins(cors.value));
                initialCorsAllowOrigins.current = cors.value;
            }
        }
    }, [settings]);

    const handleSave = (key: string, value: string, initialValue: string) => {
        if (value === initialValue) return;

        setSetting.mutate({ key, value }, {
            onSuccess: () => {
                if (key === SettingKey.ProxyURL) {
                    initialProxyUrl.current = value;
                } else if (key === SettingKey.ApiBaseUrl) {
                    initialApiBaseUrl.current = value;
                } else if (key === SettingKey.ApiAlternateBaseUrls) {
                    initialApiAlternateBaseUrls.current = value;
                } else if (key === SettingKey.TrustedProxyCIDRs) {
                    initialTrustedProxyCIDRs.current = value;
                } else if (key === SettingKey.OpsIPDisplayMode) {
                    initialOpsIPDisplayMode.current = value;
                } else if (key === SettingKey.StatsSaveInterval) {
                    initialStatsSaveInterval.current = value;
                } else if (key === SettingKey.CORSAllowOrigins) {
                    initialCorsAllowOrigins.current = value;
                }
            }
        });
    };

    const publicBaseSummary = useMemo(() => {
        const items = [
            publicAccess?.primary_base_url,
            ...(publicAccess?.alternate_base_urls ?? []),
            publicAccess?.current_base_url,
        ].map((item) => item?.trim()).filter(Boolean);
        return Array.from(new Set(items));
    }, [publicAccess]);

    return (
        <div className="octo-setting-card">
            <h2 className="octo-setting-heading">
                <Monitor className="size-4" />
                {t('system')}
            </h2>

            {/* 代理地址 */}
            <div className="octo-setting-row">
                <div className="octo-setting-label">
                    <Globe className="size-4 text-muted-foreground" />
                    <span className="text-sm font-medium">{t('proxyUrl.label')}</span>
                </div>
                <Input
                    value={proxyUrl}
                    onChange={(e) => setProxyUrl(e.target.value)}
                    onBlur={() => handleSave('proxy_url', proxyUrl, initialProxyUrl.current)}
                    placeholder={t('proxyUrl.placeholder')}
                    className="h-9 rounded-xl"
                />
            </div>

            <div className="rounded-xl border border-border/70 bg-muted/15 p-2.5">
                <div className="flex flex-wrap items-center justify-between gap-2">
                    <div>
                        <div className="flex items-center gap-2 text-sm font-semibold text-card-foreground">
                            <Link2 className="size-4 text-primary" />
                            {t('publicAccess.title')}
                        </div>
                    </div>
                    <div className="rounded-full border border-border/60 bg-background/60 px-2.5 py-1 text-[11px] text-muted-foreground">
                        {t('publicAccess.currentClientLabel', { label: publicAccess?.current_client_label || '-' })}
                    </div>
                </div>

                <div className="mt-2 grid gap-2 md:grid-cols-2">
                    <label className="grid gap-1 text-xs text-muted-foreground">
                        {t('publicAccess.primaryBaseUrl')}
                        <Input
                            value={apiBaseUrl}
                            onChange={(e) => setApiBaseUrl(e.target.value)}
                            onBlur={() => handleSave(SettingKey.ApiBaseUrl, apiBaseUrl.trim(), initialApiBaseUrl.current)}
                            placeholder={t('publicAccess.primaryBaseUrlPlaceholder')}
                            className="h-9 rounded-xl"
                        />
                    </label>
                    <label className="grid gap-1 text-xs text-muted-foreground">
                        {t('publicAccess.trustedProxyCIDRs')}
                        <Input
                            value={trustedProxyCIDRs}
                            onChange={(e) => setTrustedProxyCIDRs(e.target.value)}
                            onBlur={() => handleSave(SettingKey.TrustedProxyCIDRs, trustedProxyCIDRs.trim(), initialTrustedProxyCIDRs.current)}
                            placeholder={t('publicAccess.trustedProxyCIDRsPlaceholder')}
                            className="h-9 rounded-xl"
                        />
                    </label>
                    <label className="grid gap-1 text-xs text-muted-foreground md:col-span-2">
                        {t('publicAccess.alternateBaseUrls')}
                        <textarea
                            value={apiAlternateBaseUrls}
                            onChange={(e) => setApiAlternateBaseUrls(e.target.value)}
                            onBlur={() => handleSave(SettingKey.ApiAlternateBaseUrls, apiAlternateBaseUrls.trim(), initialApiAlternateBaseUrls.current)}
                            placeholder={t('publicAccess.alternateBaseUrlsPlaceholder')}
                            className="min-h-12 rounded-xl border border-input bg-transparent px-3 py-2 text-sm text-card-foreground outline-none transition-colors placeholder:text-muted-foreground focus-visible:border-ring"
                        />
                    </label>
                </div>

                <div className="mt-2 flex flex-wrap items-center gap-2">
                    {(['masked', 'full'] as const).map((mode) => (
                        <button
                            key={mode}
                            type="button"
                            onClick={() => {
                                setOpsIPDisplayMode(mode);
                                handleSave(SettingKey.OpsIPDisplayMode, mode, initialOpsIPDisplayMode.current);
                            }}
                            className={cn(
                                'rounded-full border px-2.5 py-1 text-xs transition',
                                opsIPDisplayMode === mode ? 'border-primary/30 bg-primary/10 text-primary' : 'border-border/60 bg-background/60 text-muted-foreground hover:text-foreground',
                            )}
                        >
                            {t(`publicAccess.ipDisplay.${mode}`)}
                        </button>
                    ))}
                </div>

                <div className="mt-2 flex min-w-0 flex-wrap gap-1.5">
                    {publicBaseSummary.map((item) => (
                        <span key={item} className="max-w-full break-all rounded-full border border-border/60 bg-background/60 px-2.5 py-1 font-mono text-[10px] text-muted-foreground">
                            {item}
                        </span>
                    ))}
                </div>
            </div>

            {/* 统计保存周期 */}
            <div className="octo-setting-row">
                <div className="octo-setting-label">
                    <Clock className="size-4 text-muted-foreground" />
                    <span className="text-sm font-medium">{t('statsSaveInterval.label')}</span>
                </div>
                <Input
                    type="number"
                    value={statsSaveInterval}
                    onChange={(e) => setStatsSaveInterval(e.target.value)}
                    onBlur={() => handleSave('stats_save_interval', statsSaveInterval, initialStatsSaveInterval.current)}
                    placeholder={t('statsSaveInterval.placeholder')}
                    className="h-9 rounded-xl"
                />
            </div>

            {/* CORS 跨域白名单 */}
            <div className="octo-setting-row">
                <div className="octo-setting-label">
                    <Shield className="size-4 text-muted-foreground" />
                    <span className="text-sm font-medium">{t('corsAllowOrigins.label')}</span>
                </div>
                <Input
                    value={corsAllowOrigins}
                    onChange={(e) => setCorsAllowOrigins(e.target.value)}
                    onBlur={() => handleSave(SettingKey.CORSAllowOrigins, corsAllowOrigins, initialCorsAllowOrigins.current)}
                    className="h-9 rounded-xl"
                />
            </div>
        </div>
    );
}
