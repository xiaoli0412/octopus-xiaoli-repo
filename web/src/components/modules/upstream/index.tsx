'use client';

import { useEffect, useMemo, useState } from 'react';
import { CalendarCheck, Gift, History, KeyRound, Loader2, PlugZap, RefreshCw, Search, ShieldCheck, Trash2, Waypoints } from 'lucide-react';
import {
    useApplyUpstreamSite,
    useCheckinUpstreamSite,
    useCreateUpstreamKey,
    useCreateUpstreamSite,
    useDeleteUpstreamSite,
    useInspectUpstreamSite,
    useRefreshUpstreamSite,
    useUpdateUpstreamSite,
    useUpstreamSiteDetail,
    useUpstreamSiteList,
    type UpstreamCheckinLogEntry,
    type UpstreamSite,
} from '@/api/endpoints/upstream';
import { useChannelList, type UpstreamAuthMode, type UpstreamInspectRequest, type UpstreamProviderType } from '@/api/endpoints/channel';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { toast } from '@/components/common/Toast';
import { cn } from '@/lib/utils';

const PROVIDERS: Array<{ value: UpstreamProviderType; label: string }> = [
    { value: 'newapi', label: 'New API' },
    { value: 'sub2api', label: 'sub2API' },
    { value: 'openai_compatible', label: '兼容接口' },
];

const AUTH_MODES: Array<{ value: UpstreamAuthMode; label: string }> = [
    { value: 'token', label: '管理令牌' },
    { value: 'access_key', label: '网关 Key' },
    { value: 'account_password', label: '账号密码' },
];

type DetailTab = 'overview' | 'keys' | 'groups' | 'prices' | 'checkin' | 'config';

function splitList(value?: string) {
    return (value ?? '').split(',').map((item) => item.trim()).filter(Boolean);
}

function formatNumber(value?: number, fallback = '-') {
    return typeof value === 'number' && Number.isFinite(value) ? value.toLocaleString('zh-CN', { maximumFractionDigits: 4 }) : fallback;
}

function formatBalance(site?: UpstreamSite) {
    if (!site?.balance_available) return '-';
    if (site.balance_unlimited) return '不限额';
    return `${formatNumber(site.balance_used, '0')} / ${formatNumber(site.balance_remain, '0')}`;
}

function providerLabel(value?: string) {
    return PROVIDERS.find((item) => item.value === value)?.label ?? '上游';
}

function statusLabel(value?: string) {
    switch ((value ?? '').trim()) {
        case 'success':
            return '已同步';
        case 'failed':
            return '失败';
        default:
            return '未同步';
    }
}

function formatDateTime(value?: string) {
    if (!value) return '-';
    try {
        return new Date(value).toLocaleString('zh-CN');
    } catch {
        return '-';
    }
}

function parseCheckinLog(value?: string): UpstreamCheckinLogEntry[] {
    if (!value) return [];
    try {
        const parsed = JSON.parse(value);
        return Array.isArray(parsed) ? parsed : [];
    } catch {
        return [];
    }
}

function PriceCell({ label, value }: { label: string; value?: number }) {
    return (
        <span className="rounded-lg bg-background/70 px-2 py-1 text-[10px] text-muted-foreground">
            {label} {formatNumber(value, '-')}
        </span>
    );
}

export function Upstream() {
    const sites = useUpstreamSiteList();
    const channels = useChannelList();
    const createSite = useCreateUpstreamSite();
    const inspectSite = useInspectUpstreamSite();
    const refreshSite = useRefreshUpstreamSite();
    const applySite = useApplyUpstreamSite();
    const createKey = useCreateUpstreamKey();
    const updateSite = useUpdateUpstreamSite();
    const checkinSite = useCheckinUpstreamSite();
    const deleteSite = useDeleteUpstreamSite();

    const [selectedID, setSelectedID] = useState<number | undefined>();
    const [pendingSiteId, setPendingSiteId] = useState<number | undefined>();
    const [selectedSiteIDs, setSelectedSiteIDs] = useState<Set<number>>(new Set());
    const [provider, setProvider] = useState<UpstreamProviderType>('newapi');
    const [authMode, setAuthMode] = useState<UpstreamAuthMode>('token');
    const [name, setName] = useState('');
    const [baseUrl, setBaseUrl] = useState('');
    const [token, setToken] = useState('');
    const [accessKey, setAccessKey] = useState('');
    const [userID, setUserID] = useState('');
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [syncToChannel, setSyncToChannel] = useState(true);
    const [targetChannelID, setTargetChannelID] = useState('new');
    const [tab, setTab] = useState<DetailTab>('overview');
    const [detailSearch, setDetailSearch] = useState('');
    const [preview, setPreview] = useState<{ model_count: number; key_count: number; group_count: number; price_count: number } | null>(null);
    const [newKeyName, setNewKeyName] = useState('');
    const [newKeyQuota, setNewKeyQuota] = useState('');
    const [newKeyExpiresAt, setNewKeyExpiresAt] = useState('');
    const [newKeyModels, setNewKeyModels] = useState('');
    const [newKeyGroups, setNewKeyGroups] = useState('');
    const [alertThreshold, setAlertThreshold] = useState('');
    const [checkinAuto, setCheckinAuto] = useState(false);
    const [checkinInterval, setCheckinInterval] = useState('86400');
    const [configAutoCreateKey, setConfigAutoCreateKey] = useState(false);
    const [configKeyQuotaLimit, setConfigKeyQuotaLimit] = useState('');
    const [configKeyExpireDays, setConfigKeyExpireDays] = useState('');
    const [configAutoSyncGroup, setConfigAutoSyncGroup] = useState(false);
    const [configAutoSyncPrice, setConfigAutoSyncPrice] = useState(false);
    const hasSites = (sites.data?.length ?? 0) > 0;

    const activeID = selectedID ?? sites.data?.[0]?.id;
    const detail = useUpstreamSiteDetail(activeID);
    const activeSite = detail.data?.site ?? sites.data?.find((item) => item.id === activeID);

    useEffect(() => {
        if (activeSite?.balance_alert_threshold !== undefined) {
            setAlertThreshold(String(activeSite.balance_alert_threshold));
        } else {
            setAlertThreshold('');
        }
    }, [activeSite?.balance_alert_threshold]);

    useEffect(() => {
        setCheckinAuto(activeSite?.auto_checkin ?? false);
        setCheckinInterval(String(activeSite?.checkin_interval_secs ?? 86400));
    }, [activeSite?.auto_checkin, activeSite?.checkin_interval_secs]);

    useEffect(() => {
        setConfigAutoCreateKey(activeSite?.auto_create_key ?? false);
        setConfigKeyQuotaLimit(activeSite?.key_quota_limit !== undefined ? String(activeSite.key_quota_limit) : '');
        setConfigKeyExpireDays(activeSite?.key_expire_days !== undefined ? String(activeSite.key_expire_days) : '');
        setConfigAutoSyncGroup(activeSite?.auto_sync_group ?? false);
        setConfigAutoSyncPrice(activeSite?.auto_sync_price ?? false);
    }, [
        activeSite?.auto_create_key,
        activeSite?.key_quota_limit,
        activeSite?.key_expire_days,
        activeSite?.auto_sync_group,
        activeSite?.auto_sync_price,
    ]);

    const inspectRequest = useMemo<UpstreamInspectRequest>(() => {
        const req: UpstreamInspectRequest = {
            provider_type: provider,
            base_url: baseUrl.trim(),
            auth_mode: authMode,
        };
        if (authMode === 'access_key') {
            req.access_key = accessKey.trim();
        } else if (authMode === 'account_password') {
            req.username = username.trim();
            req.password = password;
            req.user_id = userID.trim();
        } else {
            req.token = token.trim();
            req.user_id = userID.trim();
        }
        return req;
    }, [accessKey, authMode, baseUrl, password, provider, token, userID, username]);

    const canSubmit = Boolean(baseUrl.trim()) && (
        authMode === 'account_password'
            ? Boolean(username.trim() && password)
            : authMode === 'access_key'
                ? Boolean(accessKey.trim())
                : Boolean(token.trim())
    );

    const filteredKeys = useMemo(() => {
        const term = detailSearch.trim().toLowerCase();
        const values = detail.data?.keys ?? [];
        if (!term) return values;
        return values.filter((item) => [item.name, item.masked_key, item.allowed_models, item.groups, item.status].some((value) => value?.toLowerCase().includes(term)));
    }, [detail.data?.keys, detailSearch]);

    const filteredGroups = useMemo(() => {
        const term = detailSearch.trim().toLowerCase();
        const values = detail.data?.groups ?? [];
        if (!term) return values;
        return values.filter((item) => [item.name, item.platform, item.models, item.status, item.source].some((value) => value?.toLowerCase().includes(term)));
    }, [detail.data?.groups, detailSearch]);

    const filteredPrices = useMemo(() => {
        const term = detailSearch.trim().toLowerCase();
        const values = detail.data?.prices ?? [];
        if (!term) return values;
        return values.filter((item) => [item.model_name, item.canonical_name, item.price_source, item.source_label].some((value) => value?.toLowerCase().includes(term)));
    }, [detail.data?.prices, detailSearch]);

    const handleInspect = async () => {
        try {
            const result = await inspectSite.mutateAsync(inspectRequest);
            setPreview({
                model_count: result.model_count,
                key_count: result.keys?.length ?? 0,
                group_count: result.groups?.length ?? 0,
                price_count: result.price_candidates?.length ?? 0,
            });
        } catch (error) {
            setPreview(null);
            toast.error('上游检查失败', { description: error instanceof Error ? error.message : undefined });
        }
    };

    const handleCreate = async () => {
        try {
            const result = await createSite.mutateAsync({
                ...inspectRequest,
                name: name.trim() || undefined,
                auto_refresh: true,
                refresh_interval_secs: 12 * 60 * 60,
                sync_to_channel: syncToChannel,
                target_channel_id: targetChannelID === 'new' ? undefined : Number(targetChannelID),
                channel_name: name.trim() || undefined,
            });
            setSelectedID(result.site.id);
            setPreview(null);
            toast.success('已保存上游');
        } catch (error) {
            toast.error('保存上游失败', { description: error instanceof Error ? error.message : undefined });
        }
    };

    const handleRefresh = async (siteId?: number, applyChannel = false) => {
        const id = siteId ?? activeID;
        if (!id) return;
        setPendingSiteId(id);
        try {
            await refreshSite.mutateAsync({ id, apply_channel: applyChannel });
            toast.success(applyChannel ? '已刷新并同步渠道' : '已刷新上游');
        } catch (error) {
            toast.error('刷新失败', { description: error instanceof Error ? error.message : undefined });
        } finally {
            setPendingSiteId(undefined);
        }
    };

    const handleDelete = async (siteId: number) => {
        if (!window.confirm('确定删除该上游站点？关联的渠道与普通 Key 不会被删除。')) return;
        setPendingSiteId(siteId);
        try {
            await deleteSite.mutateAsync(siteId);
            if (selectedID === siteId) {
                setSelectedID(undefined);
            }
            setSelectedSiteIDs((prev) => {
                const next = new Set(prev);
                next.delete(siteId);
                return next;
            });
            toast.success('已删除上游站点');
        } catch (error) {
            toast.error('删除失败', { description: error instanceof Error ? error.message : undefined });
        } finally {
            setPendingSiteId(undefined);
        }
    };

    const toggleSiteSelection = (siteId: number) => {
        setSelectedSiteIDs((prev) => {
            const next = new Set(prev);
            if (next.has(siteId)) {
                next.delete(siteId);
            } else {
                next.add(siteId);
            }
            return next;
        });
    };

    const toggleSelectAll = () => {
        if (selectedSiteIDs.size === (sites.data?.length ?? 0)) {
            setSelectedSiteIDs(new Set());
        } else {
            setSelectedSiteIDs(new Set((sites.data ?? []).map((site) => site.id)));
        }
    };

    const handleBatchRefresh = async (applyChannel = false) => {
        if (selectedSiteIDs.size === 0) return;
        if (applyChannel && !window.confirm(`确定将选中的 ${selectedSiteIDs.size} 个上游站点刷新并同步到渠道？`)) return;
        let success = 0;
        let failed = 0;
        for (const siteId of Array.from(selectedSiteIDs)) {
            try {
                await refreshSite.mutateAsync({ id: siteId, apply_channel: applyChannel });
                success++;
            } catch {
                failed++;
            }
        }
        const actionLabel = applyChannel ? '批量刷新并应用渠道' : '批量刷新';
        if (failed === 0) {
            toast.success(`已${actionLabel} ${success} 个上游站点`);
        } else {
            toast.error(`${actionLabel}完成`, { description: `成功 ${success} 个，失败 ${failed} 个` });
        }
    };

    const handleBatchDelete = async () => {
        if (selectedSiteIDs.size === 0) return;
        if (!window.confirm(`确定删除选中的 ${selectedSiteIDs.size} 个上游站点？关联的渠道与普通 Key 不会被删除。`)) return;
        let success = 0;
        let failed = 0;
        for (const siteId of Array.from(selectedSiteIDs)) {
            try {
                await deleteSite.mutateAsync(siteId);
                success++;
            } catch {
                failed++;
            }
        }
        if (selectedID !== undefined && selectedSiteIDs.has(selectedID)) {
            setSelectedID(undefined);
        }
        setSelectedSiteIDs(new Set());
        if (failed === 0) {
            toast.success(`已删除 ${success} 个上游站点`);
        } else {
            toast.error('批量删除完成', { description: `成功 ${success} 个，失败 ${failed} 个` });
        }
    };

    const handleApply = async () => {
        if (!activeID) return;
        try {
            await applySite.mutateAsync({ id: activeID });
            toast.success('已应用到渠道');
        } catch (error) {
            toast.error('应用失败', { description: error instanceof Error ? error.message : undefined });
        }
    };

    const handleCreateKey = async () => {
        if (!activeID || !newKeyName.trim()) return;
        try {
            const result = await createKey.mutateAsync({
                site_id: activeID,
                name: newKeyName.trim(),
                quota: newKeyQuota.trim() ? Number(newKeyQuota.trim()) : undefined,
                expires_at: newKeyExpiresAt.trim() || undefined,
                models: newKeyModels.split(',').map((item) => item.trim()).filter(Boolean),
                groups: newKeyGroups.split(',').map((item) => item.trim()).filter(Boolean),
            });
            setNewKeyName('');
            setNewKeyQuota('');
            setNewKeyExpiresAt('');
            setNewKeyModels('');
            setNewKeyGroups('');
            toast.success('已创建上游 Key', { description: `${result.name} · ${result.masked_key}` });
        } catch (error) {
            toast.error('创建 Key 失败', { description: error instanceof Error ? error.message : undefined });
        }
    };

    const canCreateKey = Boolean(activeID && newKeyName.trim());

    const handleSaveConfig = async () => {
        if (!activeID) return;
        const threshold = alertThreshold.trim() === '' ? 0 : Number(alertThreshold.trim());
        const interval = Number(checkinInterval.trim());
        const keyQuotaLimit = configKeyQuotaLimit.trim() === '' ? 0 : Number(configKeyQuotaLimit.trim());
        const keyExpireDays = configKeyExpireDays.trim() === '' ? 0 : Number(configKeyExpireDays.trim());
        if (Number.isNaN(threshold) || threshold < 0) {
            toast.error('预警阈值必须是大于等于 0 的数字');
            return;
        }
        if (Number.isNaN(interval) || interval < 3600) {
            toast.error('签到间隔不能小于 3600 秒');
            return;
        }
        if (Number.isNaN(keyQuotaLimit) || keyQuotaLimit < 0) {
            toast.error('Key 配额限制必须是大于等于 0 的数字');
            return;
        }
        if (Number.isNaN(keyExpireDays) || keyExpireDays < 0) {
            toast.error('Key 过期天数必须是大于等于 0 的数字');
            return;
        }
        try {
            await updateSite.mutateAsync({
                id: activeID,
                auto_checkin: checkinAuto,
                checkin_interval_secs: interval,
                balance_alert_threshold: threshold,
                auto_create_key: configAutoCreateKey,
                key_quota_limit: keyQuotaLimit,
                key_expire_days: keyExpireDays,
                auto_sync_group: configAutoSyncGroup,
                auto_sync_price: configAutoSyncPrice,
            });
            toast.success('已保存配置');
        } catch (error) {
            toast.error('保存配置失败', { description: error instanceof Error ? error.message : undefined });
        }
    };

    const handleCheckin = async () => {
        if (!activeID) return;
        try {
            const result = await checkinSite.mutateAsync(activeID);
            if (result.success) {
                toast.success('签到成功', { description: `获得额度 ${formatNumber(result.amount, '0')}${result.message ? ` · ${result.message}` : ''}` });
            } else {
                toast.error('签到失败', { description: result.message });
            }
        } catch (error) {
            toast.error('签到失败', { description: error instanceof Error ? error.message : undefined });
        }
    };

    const isBalanceAlert = Boolean(
        activeSite?.balance_available &&
        !activeSite?.balance_unlimited &&
        (activeSite?.balance_alert_threshold ?? 0) > 0 &&
        (activeSite?.balance_remain ?? Number.POSITIVE_INFINITY) <= (activeSite?.balance_alert_threshold ?? 0),
    );

    return (
        <div
            data-testid="upstream-page"
            className={cn(
                'grid h-full min-h-0 items-start gap-3',
                hasSites ? 'lg:grid-cols-[16.5rem_minmax(0,1fr)] 2xl:grid-cols-[17.5rem_minmax(0,1fr)]' : 'lg:grid-cols-1',
            )}
        >
            <aside className={cn('octo-section max-h-full self-start overflow-y-auto', !hasSites && 'hidden')}>
                <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2 text-sm font-semibold">
                        <Waypoints className="size-4 text-primary" />
                        上游
                    </div>
                    <div className="flex items-center gap-2">
                        {hasSites ? (
                            <label className="flex cursor-pointer items-center gap-1.5 text-xs text-muted-foreground transition hover:text-foreground">
                                <input
                                    type="checkbox"
                                    checked={selectedSiteIDs.size === (sites.data?.length ?? 0) && (sites.data?.length ?? 0) > 0}
                                    onChange={toggleSelectAll}
                                />
                                全选
                            </label>
                        ) : null}
                        <span className="text-xs text-muted-foreground">{sites.data?.length ?? 0} 个站点</span>
                    </div>
                </div>

                {selectedSiteIDs.size > 0 ? (
                    <div className="mt-2 flex flex-wrap items-center justify-between gap-2 rounded-xl border border-border/70 bg-background/60 p-2">
                        <span className="px-1 text-xs text-muted-foreground">已选 {selectedSiteIDs.size} 个</span>
                        <div className="flex items-center gap-1.5">
                            <Button type="button" variant="secondary" size="sm" onClick={() => handleBatchRefresh(false)} disabled={refreshSite.isPending} className="h-7 rounded-lg px-2.5 text-xs">
                                <RefreshCw className="mr-1 size-3" />
                                批量刷新
                            </Button>
                            <Button type="button" variant="secondary" size="sm" onClick={() => handleBatchRefresh(true)} disabled={refreshSite.isPending} className="h-7 rounded-lg px-2.5 text-xs">
                                <Waypoints className="mr-1 size-3" />
                                批量刷新并应用
                            </Button>
                            <Button type="button" variant="destructive" size="sm" onClick={handleBatchDelete} disabled={deleteSite.isPending} className="h-7 rounded-lg px-2.5 text-xs">
                                <Trash2 className="mr-1 size-3" />
                                批量删除
                            </Button>
                        </div>
                    </div>
                ) : null}

                <div className="mt-3 grid gap-2">
                    {(sites.data ?? []).map((site) => (
                        <div
                            key={site.id}
                            className={cn(
                                'group rounded-2xl border p-2.5 transition',
                                activeID === site.id ? 'border-primary/40 bg-primary/8' : 'border-border/70 bg-background/50 hover:bg-muted/35',
                            )}
                        >
                            <button
                                type="button"
                                onClick={() => setSelectedID(site.id)}
                                className="w-full text-left"
                            >
                                <div className="flex items-center justify-between gap-2">
                                    <span className="flex items-center gap-2 truncate text-sm font-semibold text-card-foreground">
                                        <input
                                            type="checkbox"
                                            checked={selectedSiteIDs.has(site.id)}
                                            onClick={(event) => event.stopPropagation()}
                                            onChange={() => toggleSiteSelection(site.id)}
                                            className="shrink-0"
                                        />
                                        {site.name}
                                    </span>
                                    <span className="shrink-0 rounded-full border border-border/60 px-2 py-0.5 text-[10px] text-muted-foreground">{providerLabel(site.provider_type)}</span>
                                </div>
                                <div className="mt-2 grid grid-cols-3 gap-1 text-center text-[10px] text-muted-foreground">
                                    <span className="rounded-lg bg-muted/35 px-1 py-1">模型 {site.model_count}</span>
                                    <span className="rounded-lg bg-muted/35 px-1 py-1">Key {site.key_count}</span>
                                    <span className="rounded-lg bg-muted/35 px-1 py-1">{statusLabel(site.last_refresh_status)}</span>
                                </div>
                            </button>
                            <div className="mt-2 flex items-center justify-end gap-1 opacity-100 transition sm:opacity-0 sm:group-hover:opacity-100 sm:focus-within:opacity-100">
                                <button
                                    type="button"
                                    title="刷新"
                                    onClick={() => handleRefresh(site.id, false)}
                                    disabled={Boolean(pendingSiteId)}
                                    className="inline-flex size-7 items-center justify-center rounded-lg text-muted-foreground transition hover:bg-primary/10 hover:text-primary disabled:opacity-50"
                                >
                                    {pendingSiteId === site.id ? <Loader2 className="size-3.5 animate-spin" /> : <RefreshCw className="size-3.5" />}
                                </button>
                                <button
                                    type="button"
                                    title="删除"
                                    onClick={() => handleDelete(site.id)}
                                    disabled={Boolean(pendingSiteId)}
                                    className="inline-flex size-7 items-center justify-center rounded-lg text-muted-foreground transition hover:bg-destructive/10 hover:text-destructive disabled:opacity-50"
                                >
                                    {pendingSiteId === site.id ? <Loader2 className="size-3.5 animate-spin" /> : <Trash2 className="size-3.5" />}
                                </button>
                            </div>
                        </div>
                    ))}
                    {(sites.data?.length ?? 0) === 0 ? (
                        <div className="rounded-xl border border-dashed border-border/70 p-3 text-xs leading-5 text-muted-foreground">暂无站点。右侧接入后会在这里切换管理。</div>
                    ) : null}
                </div>
            </aside>

            <main className="octo-workbench">
                <section className="octo-section">
                    <div className="octo-toolbar">
                        <div>
                            <div className="text-base font-semibold text-card-foreground">接入站点</div>
                        </div>
                        <div className="flex flex-wrap gap-1.5">
                            {!hasSites ? <span className="octo-chip">暂无站点</span> : null}
                            {PROVIDERS.map((item) => (
                                <button
                                    key={item.value}
                                    type="button"
                                    onClick={() => setProvider(item.value)}
                                    className={cn('rounded-full border px-3 py-1 text-xs transition', provider === item.value ? 'border-primary/35 bg-primary/10 text-primary' : 'border-border/70 text-muted-foreground hover:text-foreground')}
                                >
                                    {item.label}
                                </button>
                            ))}
                        </div>
                    </div>

                    <div className="mt-2.5 grid gap-2 md:grid-cols-2 xl:grid-cols-[0.8fr_1.35fr_0.68fr]">
                        <Input value={name} onChange={(event) => setName(event.target.value)} placeholder="名称（可选）" className="h-9 rounded-xl" />
                        <Input value={baseUrl} onChange={(event) => setBaseUrl(event.target.value)} placeholder={provider === 'sub2api' ? 'https://sub2api.org' : 'https://newapi.example.com'} className="h-9 rounded-xl" />
                        <Select value={authMode} onValueChange={(value) => setAuthMode(value as UpstreamAuthMode)}>
                            <SelectTrigger className="h-9 w-full rounded-xl border-input bg-background/45 text-card-foreground">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent className="max-h-72 rounded-xl border-border bg-popover text-popover-foreground shadow-xl">
                                {AUTH_MODES.map((item) => (
                                    <SelectItem key={item.value} value={item.value} className="rounded-lg">
                                        {item.label}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>

                    <div className="mt-2 grid gap-2 xl:grid-cols-[minmax(0,1fr)_auto_auto]">
                        {authMode === 'account_password' ? (
                            <div className="grid gap-2 md:grid-cols-3">
                                <Input value={username} onChange={(event) => setUsername(event.target.value)} placeholder="账号 / 邮箱" className="h-9 rounded-xl" />
                                <Input value={password} onChange={(event) => setPassword(event.target.value)} type="password" placeholder="密码" className="h-9 rounded-xl" />
                                <Input value={userID} onChange={(event) => setUserID(event.target.value)} placeholder="用户 ID（可选）" className="h-9 rounded-xl" />
                            </div>
                        ) : authMode === 'access_key' ? (
                            <Input value={accessKey} onChange={(event) => setAccessKey(event.target.value)} type="password" placeholder="网关 Key" className="h-9 rounded-xl" />
                        ) : (
                            <div className="grid gap-2 md:grid-cols-2">
                                <Input value={token} onChange={(event) => setToken(event.target.value)} type="password" placeholder="管理令牌" className="h-9 rounded-xl" />
                                <Input value={userID} onChange={(event) => setUserID(event.target.value)} placeholder="用户 ID（可选）" className="h-9 rounded-xl" />
                            </div>
                        )}
                        <Button type="button" variant="secondary" onClick={handleInspect} disabled={!canSubmit || inspectSite.isPending} className="h-9 rounded-xl px-3">
                            {inspectSite.isPending ? <Loader2 className="size-4 animate-spin" /> : <Search className="size-4" />}
                            检查
                        </Button>
                        <Button type="button" onClick={handleCreate} disabled={!canSubmit || createSite.isPending} className="h-9 rounded-xl px-3">
                            {createSite.isPending ? <Loader2 className="size-4 animate-spin" /> : <ShieldCheck className="size-4" />}
                            保存
                        </Button>
                    </div>

                    <div className="mt-2 flex flex-wrap items-center gap-2">
                        <label className="flex h-8 items-center gap-2 rounded-xl border border-border/70 px-3 text-xs text-muted-foreground">
                            <input type="checkbox" checked={syncToChannel} onChange={(event) => setSyncToChannel(event.target.checked)} />
                            保存后同步到渠道
                        </label>
                        <div className="min-w-[14rem] max-w-full">
                            <Select value={targetChannelID} onValueChange={setTargetChannelID}>
                                <SelectTrigger className="h-8 w-full rounded-xl border-input bg-background/45 text-xs text-card-foreground">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent className="max-h-80 rounded-xl border-border bg-popover text-popover-foreground shadow-xl">
                                    <SelectItem value="new" className="rounded-lg">新建渠道</SelectItem>
                                    {(channels.data ?? []).map((item) => (
                                        <SelectItem key={item.raw.id} value={String(item.raw.id)} className="rounded-lg">
                                            追加到：{item.raw.name}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        {preview ? (
                            <div className="flex flex-wrap gap-1.5 text-[11px] text-muted-foreground">
                                <span className="octo-chip">模型 {preview.model_count}</span>
                                <span className="octo-chip">Key {preview.key_count}</span>
                                <span className="octo-chip">分组 {preview.group_count}</span>
                                <span className="octo-chip">中转价 {preview.price_count}</span>
                            </div>
                        ) : null}
                    </div>
                </section>

                {!activeSite ? (
                <section className="mt-3 octo-section">
                    <div className="octo-toolbar">
                        <div className="min-w-0">
                            <div className="text-base font-semibold text-card-foreground">等待接入</div>
                            <div className="mt-1 text-xs text-muted-foreground">填好上方站点与凭据后，先检查，再保存并同步到渠道。</div>
                        </div>
                        {preview ? (
                            <div className="flex flex-wrap gap-1.5">
                                <span className="octo-chip">模型 {preview.model_count}</span>
                                <span className="octo-chip">Key {preview.key_count}</span>
                                <span className="octo-chip">分组 {preview.group_count}</span>
                                <span className="octo-chip">中转价 {preview.price_count}</span>
                            </div>
                        ) : null}
                    </div>
                    <div className="mt-2.5 grid gap-2 md:grid-cols-2 xl:grid-cols-4">
                        {[
                            ['终端', '规范化站点与 API Base'],
                            ['凭据', '令牌、网关 Key 或一次性登录'],
                            ['快照', '模型、Key、分组和中转价格'],
                            ['渠道', '保存后可直接出现在渠道列表'],
                        ].map(([title, desc]) => (
                            <div key={title} className="octo-stat-card">
                                <div className="text-sm font-semibold text-card-foreground">{title}</div>
                                <div className="mt-1 text-xs leading-5 text-muted-foreground">{desc}</div>
                            </div>
                        ))}
                    </div>
                </section>
                ) : (
                <section className="mt-3 octo-section">
                    <div className="octo-toolbar">
                        <div className="min-w-0">
                            <div className="truncate text-base font-semibold text-card-foreground">{activeSite?.name ?? '未选择上游'}</div>
                            <div className="mt-1 truncate text-xs text-muted-foreground">{activeSite?.api_base_url || activeSite?.base_url || '保存站点后显示终端地址'}</div>
                        </div>
                        <div className="flex flex-wrap gap-2">
                            <Button type="button" variant="secondary" onClick={() => handleRefresh(activeID, false)} disabled={!activeID || Boolean(pendingSiteId)} className="h-8 rounded-xl px-3">
                                {pendingSiteId === activeID ? <Loader2 className="size-4 animate-spin" /> : <RefreshCw className="size-4" />}
                                刷新
                            </Button>
                            <Button type="button" variant="secondary" onClick={() => handleRefresh(activeID, true)} disabled={!activeID || Boolean(pendingSiteId)} className="h-8 rounded-xl px-3">
                                {pendingSiteId === activeID ? <Loader2 className="size-4 animate-spin" /> : <RefreshCw className="size-4" />}
                                立即同步
                            </Button>
                            <Button type="button" onClick={handleApply} disabled={!activeID || applySite.isPending || Boolean(pendingSiteId)} className="h-8 rounded-xl px-3">
                                {applySite.isPending ? <Loader2 className="size-4 animate-spin" /> : <PlugZap className="size-4" />}
                                应用渠道
                            </Button>
                        </div>
                    </div>

                    <div className="mt-2.5 grid grid-cols-2 gap-2 md:grid-cols-3 xl:grid-cols-6">
                        {[
                            ['余额', formatBalance(activeSite)],
                            ['模型', String(activeSite?.model_count ?? 0)],
                            ['Key', String(activeSite?.key_count ?? 0)],
                            ['分组', String(activeSite?.group_count ?? 0)],
                            ['中转价', String(activeSite?.price_count ?? 0)],
                            ['刷新', statusLabel(activeSite?.last_refresh_status)],
                        ].map(([label, value]) => (
                            <div key={label} className="octo-stat-card">
                                <div className="text-[11px] text-muted-foreground">{label}</div>
                                <div className="mt-1 truncate text-sm font-semibold text-card-foreground">{value}</div>
                            </div>
                        ))}
                    </div>

                    <div className="mt-2.5 flex flex-wrap items-center gap-2">
                        {[
                            ['overview', '概览'],
                            ['keys', 'Key'],
                            ['groups', '分组'],
                            ['prices', '中转价格'],
                            ['checkin', '签到'],
                            ['config', '配置'],
                        ].map(([key, label]) => (
                            <button
                                key={key}
                                type="button"
                                onClick={() => setTab(key as DetailTab)}
                                className={cn('h-8 rounded-xl px-3 text-xs font-medium transition', tab === key ? 'bg-primary text-primary-foreground' : 'border border-border/70 bg-background/60 text-muted-foreground hover:bg-muted/50')}
                            >
                                {label}
                            </button>
                        ))}
                        {tab !== 'overview' ? (
                            <div className="ml-auto min-w-[12rem] max-w-full">
                                <Input value={detailSearch} onChange={(event) => setDetailSearch(event.target.value)} placeholder="搜索明细" className="h-8 rounded-xl text-xs" />
                            </div>
                        ) : null}
                    </div>

                    {tab === 'overview' ? (
                        <div className="mt-2.5 grid gap-2 xl:grid-cols-4">
                            <div className="octo-subsection text-xs">
                                <div className="font-medium text-card-foreground">终端</div>
                                <div className="mt-2 space-y-1 text-muted-foreground">
                                    <div className="truncate">站点：{activeSite?.base_url || '-'}</div>
                                    <div className="truncate">接口：{activeSite?.api_base_url || '-'}</div>
                                    <div>方式：{AUTH_MODES.find((item) => item.value === activeSite?.auth_mode)?.label ?? '-'}</div>
                                </div>
                            </div>
                            <div className="octo-subsection text-xs">
                                <div className="font-medium text-card-foreground">凭据</div>
                                <div className="mt-2 flex flex-wrap gap-1.5">
                                    {(detail.data?.credentials ?? []).slice(0, 4).map((item) => (
                                        <span key={item.id} className="rounded-full border border-border/70 px-2 py-1 text-muted-foreground">{item.display_name}: {item.masked_value}</span>
                                    ))}
                                    {(detail.data?.credentials?.length ?? 0) === 0 ? <span className="text-muted-foreground">-</span> : null}
                                </div>
                            </div>
                            <div className="octo-subsection text-xs">
                                <div className="font-medium text-card-foreground">渠道</div>
                                <div className="mt-2 text-muted-foreground">
                                    {detail.data?.linked_channel ? `${detail.data.linked_channel.name} / Key ${detail.data.linked_channel.key_count}` : '尚未应用到渠道'}
                                </div>
                            </div>
                            <div className="octo-subsection text-xs">
                                <div className="flex items-center gap-2 font-medium text-card-foreground">
                                    余额预警
                                    {isBalanceAlert ? <span className="rounded-full bg-destructive/15 px-2 py-0.5 text-[10px] text-destructive">已触发</span> : null}
                                </div>
                                <div className="mt-2 space-y-1 text-muted-foreground">
                                    <div>当前：{formatBalance(activeSite)}</div>
                                    <div>阈值：{formatNumber(activeSite?.balance_alert_threshold, '未设置')}</div>
                                </div>
                            </div>
                        </div>
                    ) : null}

                    {tab === 'keys' ? (
                        <div className="mt-2.5 max-h-[24rem] overflow-y-auto pr-1">
                            <div className="octo-subsection mb-3 text-xs">
                                <div className="font-medium text-card-foreground">新建 Key</div>
                                <div className="mt-2 grid gap-2 md:grid-cols-2 xl:grid-cols-4">
                                    <Input value={newKeyName} onChange={(event) => setNewKeyName(event.target.value)} placeholder="名称 *" className="h-8 rounded-xl text-xs" />
                                    <Input value={newKeyQuota} onChange={(event) => setNewKeyQuota(event.target.value)} placeholder="配额" className="h-8 rounded-xl text-xs" />
                                    <Input value={newKeyExpiresAt} onChange={(event) => setNewKeyExpiresAt(event.target.value)} placeholder="过期时间" className="h-8 rounded-xl text-xs" />
                                    <Input value={newKeyModels} onChange={(event) => setNewKeyModels(event.target.value)} placeholder="模型（逗号分隔）" className="h-8 rounded-xl text-xs" />
                                    <Input value={newKeyGroups} onChange={(event) => setNewKeyGroups(event.target.value)} placeholder="分组（逗号分隔）" className="h-8 rounded-xl text-xs" />
                                    <Button type="button" onClick={handleCreateKey} disabled={!canCreateKey || createKey.isPending} className="h-8 rounded-xl px-3 text-xs md:col-span-1 xl:col-span-3">
                                        {createKey.isPending ? <Loader2 className="size-4 animate-spin" /> : <KeyRound className="size-4" />}
                                        创建 Key
                                    </Button>
                                </div>
                            </div>
                            <div className="grid gap-2 xl:grid-cols-2">
                                {filteredKeys.map((item) => (
                                    <div key={item.id} className="octo-subsection text-xs">
                                        <div className="flex items-center justify-between gap-2">
                                            <span className="truncate font-medium text-card-foreground">{item.name || `Key ${item.id}`}</span>
                                            <span className="font-mono text-[11px] text-muted-foreground">{item.masked_key || '-'}</span>
                                        </div>
                                        <div className="mt-2 flex flex-wrap gap-1.5 text-[11px] text-muted-foreground">
                                            <span className="rounded-full border border-border/70 px-2 py-0.5">{item.importable ? '可导入' : '仅展示'}</span>
                                            <span className="rounded-full border border-border/70 px-2 py-0.5">模型 {splitList(item.allowed_models).length || '全部'}</span>
                                            <span className="rounded-full border border-border/70 px-2 py-0.5">分组 {splitList(item.groups).length}</span>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    ) : null}

                    {tab === 'groups' ? (
                        <div className="mt-2.5 max-h-[24rem] overflow-y-auto pr-1">
                            <div className="grid gap-2 xl:grid-cols-2">
                                {filteredGroups.map((item) => (
                                    <div key={item.id} className="octo-subsection text-xs">
                                        <div className="flex items-center justify-between gap-2">
                                            <span className="truncate font-medium text-card-foreground">{item.name}</span>
                                            <span className="text-[11px] text-muted-foreground">{item.platform || item.source || '-'}</span>
                                        </div>
                                        <div className="mt-2 flex flex-wrap gap-1.5 text-[11px] text-muted-foreground">
                                            <span className="rounded-full border border-border/70 px-2 py-0.5">{item.status || '状态 -'}</span>
                                            <span className="rounded-full border border-border/70 px-2 py-0.5">倍率 {formatNumber(item.rate_multiplier)}</span>
                                            <span className="rounded-full border border-border/70 px-2 py-0.5">模型 {splitList(item.models).length}</span>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    ) : null}

                    {tab === 'prices' ? (
                        <div className="mt-2.5">
                            <div className="mb-3 flex flex-wrap gap-2">
                                {filteredPrices.slice(0, 3).map((item) => (
                                    <div key={`preview-${item.id}`} className="octo-stat-card text-xs">
                                        <div className="max-w-[12rem] truncate font-medium text-card-foreground">{item.model_name}</div>
                                        <div className="mt-1 text-[11px] text-muted-foreground">入 {formatNumber(item.input)} / 出 {formatNumber(item.output)}</div>
                                    </div>
                                ))}
                            </div>
                            <div className="max-h-[24rem] overflow-y-auto pr-1">
                                <div className="grid gap-2 xl:grid-cols-2">
                                    {filteredPrices.map((item) => (
                                        <div key={item.id} className="octo-subsection text-xs">
                                            <div className="flex items-center justify-between gap-2">
                                                <span className="truncate font-medium text-card-foreground">{item.model_name}</span>
                                                <span className="shrink-0 text-[11px] text-muted-foreground">{item.source_label || '中转价'}</span>
                                            </div>
                                            <div className="mt-2 grid grid-cols-4 gap-1 text-center">
                                                <PriceCell label="入" value={item.input} />
                                                <PriceCell label="出" value={item.output} />
                                                <PriceCell label="读" value={item.cache_read} />
                                                <PriceCell label="写" value={item.cache_write} />
                                            </div>
                                            <div className="mt-2 truncate text-[11px] text-muted-foreground">{item.price_source || item.price_matched_key || '上游价格'}</div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    ) : null}

                    {tab === 'checkin' ? (
                        <div className="mt-2.5 grid gap-2 xl:grid-cols-2">
                            <div className="octo-subsection text-xs">
                                <div className="flex items-center gap-2 font-medium text-card-foreground">
                                    <CalendarCheck className="size-3.5 text-primary" />
                                    手动签到
                                </div>
                                <div className="mt-2 space-y-3">
                                    <div className="text-[11px] text-muted-foreground">
                                        上次签到：{formatDateTime(activeSite?.last_checkin_at)}
                                    </div>
                                    <Button
                                        type="button"
                                        onClick={handleCheckin}
                                        disabled={!activeID || checkinSite.isPending}
                                        className="h-8 rounded-xl px-3 text-xs"
                                    >
                                        {checkinSite.isPending ? <Loader2 className="size-3.5 animate-spin" /> : <Gift className="size-3.5" />}
                                        立即签到
                                    </Button>
                                </div>
                            </div>
                            <div className="octo-subsection text-xs">
                                <div className="flex items-center gap-2 font-medium text-card-foreground">
                                    <History className="size-3.5 text-primary" />
                                    签到历史
                                </div>
                                <div className="mt-2 max-h-[24rem] overflow-y-auto pr-1">
                                    {(() => {
                                        const logs = parseCheckinLog(activeSite?.checkin_log).slice().reverse();
                                        if (logs.length === 0) {
                                            return <div className="text-muted-foreground">暂无签到记录</div>;
                                        }
                                        return (
                                            <div className="grid gap-2">
                                                {logs.map((entry, index) => (
                                                    <div key={index} className="flex items-start justify-between gap-2 rounded-lg border border-border/70 bg-background/50 p-2">
                                                        <div className="min-w-0">
                                                            <div className="flex items-center gap-2">
                                                                <span className={cn('rounded-full px-2 py-0.5 text-[10px]', entry.success ? 'bg-primary/10 text-primary' : 'bg-destructive/10 text-destructive')}>
                                                                    {entry.success ? '成功' : '失败'}
                                                                </span>
                                                                {entry.amount !== undefined ? <span className="text-[10px] text-muted-foreground">+{formatNumber(entry.amount)}</span> : null}
                                                            </div>
                                                            {entry.message ? <div className="mt-1 truncate text-[11px] text-muted-foreground">{entry.message}</div> : null}
                                                        </div>
                                                        <div className="shrink-0 text-[10px] text-muted-foreground">{formatDateTime(entry.at)}</div>
                                                    </div>
                                                ))}
                                            </div>
                                        );
                                    })()}
                                </div>
                            </div>
                        </div>
                    ) : null}

                    {tab === 'config' ? (
                        <div className="mt-2.5 grid gap-2 xl:grid-cols-2">
                            <div className="octo-subsection text-xs">
                                <div className="flex items-center gap-2 font-medium text-card-foreground">
                                    <CalendarCheck className="size-3.5 text-primary" />
                                    自动签到
                                </div>
                                <div className="mt-2 space-y-3">
                                    <label className="flex items-center gap-2 text-muted-foreground">
                                        <input
                                            type="checkbox"
                                            checked={checkinAuto}
                                            onChange={(event) => setCheckinAuto(event.target.checked)}
                                        />
                                        启用自动签到
                                    </label>
                                    <Input
                                        type="number"
                                        min={3600}
                                        step={1}
                                        value={checkinInterval}
                                        onChange={(event) => setCheckinInterval(event.target.value)}
                                        placeholder="间隔秒数（默认 86400）"
                                        className="h-8 rounded-xl text-xs"
                                    />
                                </div>
                            </div>
                            <div className="octo-subsection text-xs">
                                <div className="flex items-center gap-2 font-medium text-card-foreground">
                                    余额预警
                                    {isBalanceAlert ? <span className="rounded-full bg-destructive/15 px-2 py-0.5 text-[10px] text-destructive">已触发</span> : null}
                                </div>
                                <div className="mt-2 space-y-2">
                                    <div className="text-muted-foreground">
                                        当前：{formatBalance(activeSite)}
                                    </div>
                                    <Input
                                        type="number"
                                        min={0}
                                        step="any"
                                        value={alertThreshold}
                                        onChange={(event) => setAlertThreshold(event.target.value)}
                                        placeholder="阈值（0 为关闭）"
                                        className="h-8 rounded-xl text-xs"
                                    />
                                </div>
                            </div>
                            <div className="octo-subsection text-xs">
                                <div className="font-medium text-card-foreground">Key 创建策略</div>
                                <div className="mt-2 space-y-3">
                                    <label className="flex items-center gap-2 text-muted-foreground">
                                        <input
                                            type="checkbox"
                                            checked={configAutoCreateKey}
                                            onChange={(event) => setConfigAutoCreateKey(event.target.checked)}
                                        />
                                        自动创建 Key
                                    </label>
                                    <Input
                                        type="number"
                                        min={0}
                                        step="any"
                                        value={configKeyQuotaLimit}
                                        onChange={(event) => setConfigKeyQuotaLimit(event.target.value)}
                                        placeholder="默认配额限制（0 为无限制）"
                                        className="h-8 rounded-xl text-xs"
                                    />
                                    <Input
                                        type="number"
                                        min={0}
                                        step={1}
                                        value={configKeyExpireDays}
                                        onChange={(event) => setConfigKeyExpireDays(event.target.value)}
                                        placeholder="默认过期天数（0 为不过期）"
                                        className="h-8 rounded-xl text-xs"
                                    />
                                </div>
                            </div>
                            <div className="octo-subsection text-xs">
                                <div className="font-medium text-card-foreground">同步策略</div>
                                <div className="mt-2 space-y-3">
                                    <label className="flex items-center gap-2 text-muted-foreground">
                                        <input
                                            type="checkbox"
                                            checked={activeSite?.sync_to_channel}
                                            disabled
                                        />
                                        同步到渠道
                                    </label>
                                    <label className="flex items-center gap-2 text-muted-foreground">
                                        <input
                                            type="checkbox"
                                            checked={configAutoSyncGroup}
                                            onChange={(event) => setConfigAutoSyncGroup(event.target.checked)}
                                        />
                                        自动同步分组
                                    </label>
                                    <label className="flex items-center gap-2 text-muted-foreground">
                                        <input
                                            type="checkbox"
                                            checked={configAutoSyncPrice}
                                            onChange={(event) => setConfigAutoSyncPrice(event.target.checked)}
                                        />
                                        自动同步价格
                                    </label>
                                </div>
                            </div>
                            <div className="octo-subsection flex items-center justify-end gap-2 xl:col-span-2">
                                <Button
                                    type="button"
                                    size="sm"
                                    onClick={handleSaveConfig}
                                    disabled={updateSite.isPending}
                                    className="h-8 rounded-xl px-4 text-xs"
                                >
                                    {updateSite.isPending ? <Loader2 className="size-3.5 animate-spin" /> : '保存配置'}
                                </Button>
                            </div>
                        </div>
                    ) : null}
                </section>
                )}
            </main>
        </div>
    );
}
