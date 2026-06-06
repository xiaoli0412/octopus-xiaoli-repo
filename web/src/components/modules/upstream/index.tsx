'use client';

import { useMemo, useState } from 'react';
import { Loader2, PlugZap, RefreshCw, Search, ShieldCheck, Waypoints } from 'lucide-react';
import {
    useApplyUpstreamSite,
    useCreateUpstreamSite,
    useInspectUpstreamSite,
    useRefreshUpstreamSite,
    useUpstreamSiteDetail,
    useUpstreamSiteList,
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

type DetailTab = 'overview' | 'keys' | 'groups' | 'prices';

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

    const [selectedID, setSelectedID] = useState<number | undefined>();
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

    const activeID = selectedID ?? sites.data?.[0]?.id;
    const detail = useUpstreamSiteDetail(activeID);
    const activeSite = detail.data?.site ?? sites.data?.find((item) => item.id === activeID);

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

    const handleRefresh = async (applyChannel = false) => {
        if (!activeID) return;
        try {
            await refreshSite.mutateAsync({ id: activeID, apply_channel: applyChannel });
            toast.success(applyChannel ? '已刷新并同步渠道' : '已刷新上游');
        } catch (error) {
            toast.error('刷新失败', { description: error instanceof Error ? error.message : undefined });
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

    return (
        <div data-testid="upstream-page" className="grid h-full min-h-0 gap-3 lg:grid-cols-[20rem_minmax(0,1fr)]">
            <aside className="max-h-full self-start overflow-y-auto rounded-3xl border border-border bg-card p-3">
                <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2 text-sm font-semibold">
                        <Waypoints className="size-4 text-primary" />
                        上游
                    </div>
                    <span className="text-xs text-muted-foreground">{sites.data?.length ?? 0} 个站点</span>
                </div>

                <div className="mt-3 grid gap-2">
                    {(sites.data ?? []).map((site) => (
                        <button
                            key={site.id}
                            type="button"
                            onClick={() => setSelectedID(site.id)}
                            className={cn(
                                'rounded-2xl border p-3 text-left transition',
                                activeID === site.id ? 'border-primary/40 bg-primary/8' : 'border-border/70 bg-background/50 hover:bg-muted/35',
                            )}
                        >
                            <div className="flex items-center justify-between gap-2">
                                <span className="truncate text-sm font-semibold text-card-foreground">{site.name}</span>
                                <span className="shrink-0 rounded-full border border-border/60 px-2 py-0.5 text-[10px] text-muted-foreground">{providerLabel(site.provider_type)}</span>
                            </div>
                            <div className="mt-2 grid grid-cols-3 gap-1 text-center text-[10px] text-muted-foreground">
                                <span className="rounded-lg bg-muted/35 px-1 py-1">模型 {site.model_count}</span>
                                <span className="rounded-lg bg-muted/35 px-1 py-1">Key {site.key_count}</span>
                                <span className="rounded-lg bg-muted/35 px-1 py-1">{statusLabel(site.last_refresh_status)}</span>
                            </div>
                        </button>
                    ))}
                    {(sites.data?.length ?? 0) === 0 ? (
                        <div className="rounded-2xl border border-dashed border-border/70 p-4 text-xs text-muted-foreground">还没有上游。先在右侧添加一个站点。</div>
                    ) : null}
                </div>
            </aside>

            <main className="min-h-0 overflow-y-auto rounded-t-3xl pb-28 md:pb-4">
                <section className="rounded-3xl border border-border bg-card p-4">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                        <div>
                            <div className="text-base font-semibold text-card-foreground">接入站点</div>
                            <div className="mt-1 text-xs text-muted-foreground">账号密码只用于换取令牌，不长期保存密码。</div>
                        </div>
                        <div className="flex flex-wrap gap-1.5">
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

                    <div className="mt-4 grid gap-2 xl:grid-cols-[0.8fr_1fr_0.75fr_1.3fr_auto_auto]">
                        <Input value={name} onChange={(event) => setName(event.target.value)} placeholder="名称（可选）" className="h-10 rounded-xl" />
                        <Input value={baseUrl} onChange={(event) => setBaseUrl(event.target.value)} placeholder={provider === 'sub2api' ? 'https://sub2api.org' : 'https://newapi.example.com'} className="h-10 rounded-xl" />
                        <Select value={authMode} onValueChange={(value) => setAuthMode(value as UpstreamAuthMode)}>
                            <SelectTrigger className="h-10 w-full rounded-xl border-input bg-background/45 text-card-foreground">
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
                        {authMode === 'account_password' ? (
                            <div className="grid gap-2 md:grid-cols-3">
                                <Input value={username} onChange={(event) => setUsername(event.target.value)} placeholder="账号 / 邮箱" className="h-10 rounded-xl" />
                                <Input value={password} onChange={(event) => setPassword(event.target.value)} type="password" placeholder="密码" className="h-10 rounded-xl" />
                                <Input value={userID} onChange={(event) => setUserID(event.target.value)} placeholder="用户 ID（可选）" className="h-10 rounded-xl" />
                            </div>
                        ) : authMode === 'access_key' ? (
                            <Input value={accessKey} onChange={(event) => setAccessKey(event.target.value)} type="password" placeholder="网关 Key" className="h-10 rounded-xl" />
                        ) : (
                            <div className="grid gap-2 md:grid-cols-2">
                                <Input value={token} onChange={(event) => setToken(event.target.value)} type="password" placeholder="管理令牌" className="h-10 rounded-xl" />
                                <Input value={userID} onChange={(event) => setUserID(event.target.value)} placeholder="用户 ID（可选）" className="h-10 rounded-xl" />
                            </div>
                        )}
                        <Button type="button" variant="secondary" onClick={handleInspect} disabled={!canSubmit || inspectSite.isPending} className="h-10 rounded-xl">
                            {inspectSite.isPending ? <Loader2 className="size-4 animate-spin" /> : <Search className="size-4" />}
                            检查
                        </Button>
                        <Button type="button" onClick={handleCreate} disabled={!canSubmit || createSite.isPending} className="h-10 rounded-xl">
                            {createSite.isPending ? <Loader2 className="size-4 animate-spin" /> : <ShieldCheck className="size-4" />}
                            保存
                        </Button>
                    </div>

                    <div className="mt-3 flex flex-wrap items-center gap-2">
                        <label className="flex h-9 items-center gap-2 rounded-xl border border-border/70 px-3 text-xs text-muted-foreground">
                            <input type="checkbox" checked={syncToChannel} onChange={(event) => setSyncToChannel(event.target.checked)} />
                            保存后同步到渠道
                        </label>
                        <div className="min-w-[14rem] max-w-full">
                            <Select value={targetChannelID} onValueChange={setTargetChannelID}>
                                <SelectTrigger className="h-9 w-full rounded-xl border-input bg-background/45 text-xs text-card-foreground">
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
                                <span className="rounded-full border border-border/70 px-2 py-1">模型 {preview.model_count}</span>
                                <span className="rounded-full border border-border/70 px-2 py-1">Key {preview.key_count}</span>
                                <span className="rounded-full border border-border/70 px-2 py-1">分组 {preview.group_count}</span>
                                <span className="rounded-full border border-border/70 px-2 py-1">中转价 {preview.price_count}</span>
                            </div>
                        ) : null}
                    </div>
                </section>

                <section className="mt-3 rounded-3xl border border-border bg-card p-4">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                        <div className="min-w-0">
                            <div className="truncate text-base font-semibold text-card-foreground">{activeSite?.name ?? '未选择上游'}</div>
                            <div className="mt-1 truncate text-xs text-muted-foreground">{activeSite?.api_base_url || activeSite?.base_url || '保存站点后显示终端地址'}</div>
                        </div>
                        <div className="flex flex-wrap gap-2">
                            <Button type="button" variant="secondary" onClick={() => handleRefresh(false)} disabled={!activeID || refreshSite.isPending} className="h-9 rounded-xl">
                                {refreshSite.isPending ? <Loader2 className="size-4 animate-spin" /> : <RefreshCw className="size-4" />}
                                刷新
                            </Button>
                            <Button type="button" onClick={handleApply} disabled={!activeID || applySite.isPending} className="h-9 rounded-xl">
                                {applySite.isPending ? <Loader2 className="size-4 animate-spin" /> : <PlugZap className="size-4" />}
                                应用渠道
                            </Button>
                        </div>
                    </div>

                    <div className="mt-4 grid grid-cols-2 gap-2 md:grid-cols-3 xl:grid-cols-6">
                        {[
                            ['余额', formatBalance(activeSite)],
                            ['模型', String(activeSite?.model_count ?? 0)],
                            ['Key', String(activeSite?.key_count ?? 0)],
                            ['分组', String(activeSite?.group_count ?? 0)],
                            ['中转价', String(activeSite?.price_count ?? 0)],
                            ['刷新', statusLabel(activeSite?.last_refresh_status)],
                        ].map(([label, value]) => (
                            <div key={label} className="rounded-2xl border border-border/60 bg-muted/15 px-3 py-2">
                                <div className="text-[11px] text-muted-foreground">{label}</div>
                                <div className="mt-1 truncate text-sm font-semibold text-card-foreground">{value}</div>
                            </div>
                        ))}
                    </div>

                    <div className="mt-4 flex flex-wrap items-center gap-2">
                        {[
                            ['overview', '概览'],
                            ['keys', 'Key'],
                            ['groups', '分组'],
                            ['prices', '中转价格'],
                        ].map(([key, label]) => (
                            <button
                                key={key}
                                type="button"
                                onClick={() => setTab(key as DetailTab)}
                                className={cn('h-8 rounded-xl px-3 text-xs font-medium transition', tab === key ? 'bg-primary text-primary-foreground' : 'border border-border/70 text-muted-foreground hover:bg-muted/50')}
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
                        <div className="mt-4 grid gap-3 xl:grid-cols-3">
                            <div className="rounded-2xl border border-border/60 bg-background/55 p-3 text-xs">
                                <div className="font-medium text-card-foreground">终端</div>
                                <div className="mt-2 space-y-1 text-muted-foreground">
                                    <div className="truncate">站点：{activeSite?.base_url || '-'}</div>
                                    <div className="truncate">接口：{activeSite?.api_base_url || '-'}</div>
                                    <div>方式：{AUTH_MODES.find((item) => item.value === activeSite?.auth_mode)?.label ?? '-'}</div>
                                </div>
                            </div>
                            <div className="rounded-2xl border border-border/60 bg-background/55 p-3 text-xs">
                                <div className="font-medium text-card-foreground">凭据</div>
                                <div className="mt-2 flex flex-wrap gap-1.5">
                                    {(detail.data?.credentials ?? []).slice(0, 4).map((item) => (
                                        <span key={item.id} className="rounded-full border border-border/70 px-2 py-1 text-muted-foreground">{item.display_name}: {item.masked_value}</span>
                                    ))}
                                    {(detail.data?.credentials?.length ?? 0) === 0 ? <span className="text-muted-foreground">-</span> : null}
                                </div>
                            </div>
                            <div className="rounded-2xl border border-border/60 bg-background/55 p-3 text-xs">
                                <div className="font-medium text-card-foreground">渠道</div>
                                <div className="mt-2 text-muted-foreground">
                                    {detail.data?.linked_channel ? `${detail.data.linked_channel.name} / Key ${detail.data.linked_channel.key_count}` : '尚未应用到渠道'}
                                </div>
                            </div>
                        </div>
                    ) : null}

                    {tab === 'keys' ? (
                        <div className="mt-4 max-h-[28rem] overflow-y-auto pr-1">
                            <div className="grid gap-2 xl:grid-cols-2">
                                {filteredKeys.map((item) => (
                                    <div key={item.id} className="rounded-2xl border border-border/60 bg-background/55 p-3 text-xs">
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
                        <div className="mt-4 max-h-[28rem] overflow-y-auto pr-1">
                            <div className="grid gap-2 xl:grid-cols-2">
                                {filteredGroups.map((item) => (
                                    <div key={item.id} className="rounded-2xl border border-border/60 bg-background/55 p-3 text-xs">
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
                        <div className="mt-4">
                            <div className="mb-3 flex flex-wrap gap-2">
                                {filteredPrices.slice(0, 3).map((item) => (
                                    <div key={`preview-${item.id}`} className="rounded-2xl border border-border/60 bg-background/70 px-3 py-2 text-xs">
                                        <div className="max-w-[12rem] truncate font-medium text-card-foreground">{item.model_name}</div>
                                        <div className="mt-1 text-[11px] text-muted-foreground">入 {formatNumber(item.input)} / 出 {formatNumber(item.output)}</div>
                                    </div>
                                ))}
                            </div>
                            <div className="max-h-[28rem] overflow-y-auto pr-1">
                                <div className="grid gap-2 xl:grid-cols-2">
                                    {filteredPrices.map((item) => (
                                        <div key={item.id} className="rounded-2xl border border-border/60 bg-background/55 p-3 text-xs">
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
                </section>
            </main>
        </div>
    );
}
