'use client';

import { useMemo, useState } from 'react';
import { Loader2, PlugZap, ScanSearch } from 'lucide-react';
import {
    useApplyUpstreamGateway,
    useChannelList,
    useInspectUpstreamGateway,
    type UpstreamAuthMode,
    type UpstreamInspectRequest,
    type UpstreamInspectResult,
    type UpstreamProviderType,
} from '@/api/endpoints/channel';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { toast } from '@/components/common/Toast';
import { cn } from '@/lib/utils';

const PROVIDER_OPTIONS: Array<{ key: UpstreamProviderType; label: string }> = [
    { key: 'newapi', label: 'New API' },
    { key: 'sub2api', label: 'sub2API' },
    { key: 'openai_compatible', label: 'OpenAI 兼容' },
];

const AUTH_MODE_OPTIONS: Array<{ key: UpstreamAuthMode; label: string }> = [
    { key: 'token', label: '管理令牌' },
    { key: 'access_key', label: '网关密钥' },
    { key: 'account_password', label: '账号密码' },
];

function compactList(values: string[] | undefined, limit: number) {
    const items = values ?? [];
    return { visible: items.slice(0, limit), hidden: Math.max(0, items.length - limit) };
}

function defaultChannelName(result: UpstreamInspectResult) {
    const provider = result.provider_type === 'sub2api' ? 'sub2API' : result.provider_type === 'newapi' ? 'New API' : '兼容上游';
    try {
        return `${provider} - ${new URL(result.base_url).hostname}`;
    } catch {
        return provider;
    }
}

function usageLabel(result: UpstreamInspectResult | null) {
    if (!result?.token_usage?.available) return '-';
    if (result.token_usage.unlimited) return '不限额';
    const used = result.token_usage.used_quota ?? 0;
    const remain = result.token_usage.remain_quota ?? 0;
    return `${used} / ${remain}`;
}

function subscriptionLabel(item: NonNullable<UpstreamInspectResult['subscriptions']>[number]) {
    return [item.name, item.plan, item.status].filter(Boolean).join(' · ') || item.source || '订阅';
}

function formatPriceValue(value?: number) {
    if (!value) return '-';
    if (value >= 1) return value.toFixed(3).replace(/0+$/, '').replace(/\.$/, '');
    return value.toPrecision(3);
}

function formatOptionalNumber(value?: number, empty = '-') {
    return typeof value === 'number' && Number.isFinite(value) ? String(value) : empty;
}

function groupLabel(item: NonNullable<UpstreamInspectResult['groups']>[number]) {
    return item.name || item.id || '分组';
}

function MiniPills({ values, limit = 3, tone = 'muted' }: { values?: string[]; limit?: number; tone?: 'muted' | 'primary' }) {
    const list = compactList(values, limit);
    if (!list.visible.length) return null;
    return (
        <div className="mt-2 flex flex-wrap gap-1">
            {list.visible.map((value) => (
                <span
                    key={value}
                    className={cn(
                        'max-w-full truncate rounded-full border px-2 py-0.5 text-[10px]',
                        tone === 'primary' ? 'border-primary/20 bg-primary/5 text-primary' : 'border-border/60 bg-background/60 text-muted-foreground',
                    )}
                >
                    {value}
                </span>
            ))}
            {list.hidden > 0 ? <span className="rounded-full border border-border/60 px-2 py-0.5 text-[10px] text-muted-foreground">+{list.hidden}</span> : null}
        </div>
    );
}

export function UpstreamImport() {
    const channels = useChannelList();
    const inspectUpstream = useInspectUpstreamGateway();
    const applyUpstream = useApplyUpstreamGateway();

    const [provider, setProvider] = useState<UpstreamProviderType>('newapi');
    const [authMode, setAuthMode] = useState<UpstreamAuthMode>('token');
    const [baseUrl, setBaseUrl] = useState('');
    const [token, setToken] = useState('');
    const [accessKey, setAccessKey] = useState('');
    const [userID, setUserID] = useState('');
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [channelName, setChannelName] = useState('');
    const [targetChannelID, setTargetChannelID] = useState('new');
    const [result, setResult] = useState<UpstreamInspectResult | null>(null);
    const [showDetails, setShowDetails] = useState(false);

    const request = useMemo<UpstreamInspectRequest>(() => {
        const base: UpstreamInspectRequest = {
            provider_type: provider,
            base_url: baseUrl.trim(),
            auth_mode: authMode,
        };
        if (authMode === 'access_key') {
            base.access_key = accessKey.trim() || undefined;
        } else if (authMode === 'account_password') {
            base.username = username.trim() || undefined;
            base.password = password || undefined;
            base.user_id = userID.trim() || undefined;
        } else {
            base.token = token.trim() || undefined;
            base.user_id = userID.trim() || undefined;
        }
        return base;
    }, [accessKey, authMode, baseUrl, password, provider, token, userID, username]);

    const canInspect = Boolean(baseUrl.trim())
        && (authMode === 'account_password'
            ? Boolean(username.trim() && password)
            : authMode === 'access_key'
                ? Boolean(accessKey.trim())
                : Boolean(token.trim()));
    const importedKeyCount = result?.keys?.length ?? 0;
    const importableKeyCount = result?.keys?.filter((key) => key.importable).length ?? 0;
    const displayKeyCount = importedKeyCount || (result && authMode === 'access_key' && accessKey.trim() ? 1 : 0);
    const canApply = Boolean(result?.model_count)
        && (importableKeyCount > 0 || (authMode === 'access_key' && Boolean(accessKey.trim())));

    const visibleModels = compactList(result?.models, 12);
    const visiblePrices = compactList(result?.price_candidates?.map((item) => item.name), 8);
    const visibleCapabilities = compactList(result?.request_capabilities, 5);
    const visibleKeys = compactList(result?.keys?.map((item, index) => item.masked_key || item.name || `Key ${index + 1}`), 8);
    const visibleSubscriptions = compactList(result?.subscriptions?.map((item) => subscriptionLabel(item)), 6);
    const visibleGroups = compactList(result?.groups?.map((item) => groupLabel(item)), 8);
    const showNewAPIUser = provider === 'newapi' && authMode !== 'access_key';

    const handleInspect = async () => {
        try {
            const next = await inspectUpstream.mutateAsync(request);
            setResult(next);
            setShowDetails(true);
            if (!channelName.trim()) {
                setChannelName(defaultChannelName(next));
            }
        } catch (error) {
            setResult(null);
            toast.error('上游检查失败', { description: error instanceof Error ? error.message : undefined });
        }
    };

    const handleApply = async () => {
        if (!result) return;
        try {
            const applied = await applyUpstream.mutateAsync({
                inspect: request,
                target_channel_id: targetChannelID === 'new' ? undefined : Number(targetChannelID),
                channel_name: targetChannelID === 'new' ? (channelName.trim() || defaultChannelName(result)) : undefined,
                append_keys: true,
                overwrite_models: true,
                enable_channel: true,
            });
            setResult(applied.inspect);
            toast.success(applied.created ? '已创建渠道' : '已更新渠道');
        } catch (error) {
            toast.error('应用到渠道失败', { description: error instanceof Error ? error.message : undefined });
        }
    };

    return (
        <div className="h-full min-h-0 overflow-y-auto rounded-t-3xl">
            <div className="space-y-3 pb-28 md:pb-6">
                <section className="rounded-3xl border border-border bg-card p-4">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                        <div className="flex items-center gap-2">
                            <ScanSearch className="size-4 text-primary" />
                            <div className="text-base font-semibold text-card-foreground">上游接入</div>
                        </div>
                        <div className="flex flex-wrap gap-1.5">
                            {PROVIDER_OPTIONS.map((item) => (
                                <button
                                    key={item.key}
                                    type="button"
                                    onClick={() => {
                                        setProvider(item.key);
                                        setResult(null);
                                    }}
                                    className={cn(
                                        'rounded-full border px-3 py-1 text-xs transition',
                                        provider === item.key ? 'border-primary/30 bg-primary/10 text-primary' : 'border-border/60 bg-background/60 text-muted-foreground hover:text-foreground',
                                    )}
                                >
                                    {item.label}
                                </button>
                            ))}
                        </div>
                    </div>

                    <div className="mt-4 grid gap-2 md:grid-cols-2 xl:grid-cols-[1.1fr_0.65fr_1.25fr_auto]">
                        <Input
                            value={baseUrl}
                            onChange={(event) => setBaseUrl(event.target.value)}
                            placeholder={provider === 'sub2api' ? 'https://sub2api.org' : 'https://newapi.example.com'}
                            className="h-10 rounded-xl"
                        />
                        <Select
                            value={authMode}
                            onValueChange={(value) => {
                                setAuthMode(value as UpstreamAuthMode);
                                setResult(null);
                            }}
                        >
                            <SelectTrigger className="h-10 w-full rounded-xl border-input bg-background/45 text-card-foreground">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent className="max-h-72 rounded-xl border-border bg-popover text-popover-foreground shadow-xl">
                                {AUTH_MODE_OPTIONS.map((item) => (
                                    <SelectItem key={item.key} value={item.key} className="rounded-lg">
                                        {item.label}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                        {authMode === 'account_password' ? (
                            <div className={cn('grid gap-2 sm:grid-cols-2', showNewAPIUser ? 'xl:grid-cols-3' : 'xl:grid-cols-2')}>
                                <Input value={username} onChange={(event) => setUsername(event.target.value)} placeholder="账号 / 邮箱" className="h-10 rounded-xl" />
                                <Input value={password} onChange={(event) => setPassword(event.target.value)} placeholder="密码（不保存）" type="password" className="h-10 rounded-xl" />
                                {showNewAPIUser ? <Input value={userID} onChange={(event) => setUserID(event.target.value)} placeholder="User ID（可选）" className="h-10 rounded-xl" /> : null}
                            </div>
                        ) : authMode === 'access_key' ? (
                            <Input value={accessKey} onChange={(event) => setAccessKey(event.target.value)} placeholder="网关密钥" type="password" className="h-10 rounded-xl" />
                        ) : (
                            <div className={cn('grid gap-2', showNewAPIUser ? 'sm:grid-cols-2' : '')}>
                                <Input value={token} onChange={(event) => setToken(event.target.value)} placeholder="管理令牌（模型、余额、密钥）" type="password" className="h-10 rounded-xl" />
                                {showNewAPIUser ? <Input value={userID} onChange={(event) => setUserID(event.target.value)} placeholder="User ID（可选）" className="h-10 rounded-xl" /> : null}
                            </div>
                        )}
                        <Button type="button" onClick={handleInspect} disabled={inspectUpstream.isPending || !canInspect} className="h-10 rounded-xl">
                            {inspectUpstream.isPending ? <Loader2 className="size-4 animate-spin" /> : null}
                            {inspectUpstream.isPending ? '检查中' : '检查'}
                        </Button>
                    </div>
                </section>

                {result ? (
                    <section className="rounded-3xl border border-border bg-card p-4">
                        <div className="grid grid-cols-2 gap-2 lg:grid-cols-3 xl:grid-cols-6">
                            {[
                                ['模型', String(result.model_count)],
                                ['Key', `${displayKeyCount || 0} / 可导入 ${importableKeyCount || (authMode === 'access_key' && accessKey.trim() ? 1 : 0)}`],
                                ['分组', String(result.groups?.length ?? 0)],
                                ['余额', usageLabel(result)],
                                ['价格候选', String(result.price_candidates?.length ?? 0)],
                                ['订阅', String(result.subscriptions?.length ?? 0)],
                            ].map(([label, value]) => (
                                <div key={label} className="rounded-2xl border border-border/60 bg-muted/15 px-3 py-2">
                                    <div className="text-[11px] text-muted-foreground">{label}</div>
                                    <div className="mt-1 truncate text-sm font-semibold text-card-foreground">{value}</div>
                                </div>
                            ))}
                        </div>

                        <div className="mt-3 rounded-2xl border border-border/60 bg-background/55 p-3">
                            <div className="flex flex-wrap gap-1.5">
                                {visibleCapabilities.visible.map((item) => <span key={item} className="rounded-full border border-primary/20 bg-primary/5 px-2 py-1 text-[11px] text-primary">{item}</span>)}
                                {visibleModels.visible.map((item) => <span key={item} className="max-w-full break-all rounded-full border border-border/60 bg-muted/25 px-2 py-1 font-mono text-[11px] text-muted-foreground">{item}</span>)}
                                {visibleModels.hidden > 0 ? <span className="rounded-full border border-border/60 px-2 py-1 text-[11px] text-muted-foreground">模型 +{visibleModels.hidden}</span> : null}
                                {visibleGroups.hidden > 0 ? <span className="rounded-full border border-border/60 px-2 py-1 text-[11px] text-muted-foreground">分组 +{visibleGroups.hidden}</span> : null}
                                {visibleKeys.hidden > 0 ? <span className="rounded-full border border-border/60 px-2 py-1 text-[11px] text-muted-foreground">Key +{visibleKeys.hidden}</span> : null}
                                {visiblePrices.hidden > 0 ? <span className="rounded-full border border-border/60 px-2 py-1 text-[11px] text-muted-foreground">价格 +{visiblePrices.hidden}</span> : null}
                                {visibleSubscriptions.hidden > 0 ? <span className="rounded-full border border-border/60 px-2 py-1 text-[11px] text-muted-foreground">订阅 +{visibleSubscriptions.hidden}</span> : null}
                            </div>
                        </div>

                        <div className="mt-3 grid gap-2 lg:grid-cols-[1fr_1fr_auto]">
                            <Select
                                value={targetChannelID}
                                onValueChange={setTargetChannelID}
                            >
                                <SelectTrigger className="h-10 w-full rounded-xl border-input bg-background/45 text-card-foreground">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent className="max-h-80 rounded-xl border-border bg-popover text-popover-foreground shadow-xl">
                                    <SelectItem value="new" className="rounded-lg">新建普通渠道</SelectItem>
                                {(channels.data ?? []).map((item) => (
                                        <SelectItem key={item.raw.id} value={String(item.raw.id)} className="rounded-lg">
                                            追加到：{item.raw.name}
                                        </SelectItem>
                                ))}
                                </SelectContent>
                            </Select>
                            <Input
                                value={channelName}
                                onChange={(event) => setChannelName(event.target.value)}
                                disabled={targetChannelID !== 'new'}
                                placeholder="新建渠道名称"
                                className="h-10 rounded-xl"
                            />
                            <Button type="button" onClick={handleApply} disabled={applyUpstream.isPending || !canApply} className="h-10 rounded-xl">
                                {applyUpstream.isPending ? <Loader2 className="size-4 animate-spin" /> : <PlugZap className="size-4" />}
                                {applyUpstream.isPending ? '应用中' : '应用到渠道'}
                            </Button>
                        </div>

                        <button
                            type="button"
                            onClick={() => setShowDetails((current) => !current)}
                            className="mt-3 rounded-xl border border-border/60 px-3 py-1.5 text-xs text-muted-foreground transition hover:text-foreground"
                        >
                            {showDetails ? '收起明细' : '查看 Key / 订阅明细'}
                        </button>

                        {showDetails ? (
                            <div className="mt-3 grid gap-3 xl:grid-cols-2">
                                <div className="space-y-2">
                                    <div className="flex items-center justify-between text-xs font-medium text-muted-foreground">
                                        <span>上游 Key</span>
                                        {(result.keys?.length ?? 0) > 8 ? <span>+{(result.keys?.length ?? 0) - 8}</span> : null}
                                    </div>
                                    <div className="max-h-72 overflow-y-auto pr-1">
                                        <div className="grid gap-2 md:grid-cols-2">
                                            {(result.keys ?? []).map((key, index) => (
                                                <div key={`${key.masked_key || key.name}-${index}`} className="rounded-2xl border border-border/60 bg-muted/15 p-3 text-xs">
                                                    <div className="flex items-center justify-between gap-2">
                                                        <span className="truncate font-medium text-card-foreground">{key.name || `Key ${index + 1}`}</span>
                                                        <span className="shrink-0 font-mono text-[11px] text-muted-foreground">{key.masked_key || '-'}</span>
                                                    </div>
                                                    <div className="mt-2 flex flex-wrap gap-1.5 text-[11px] text-muted-foreground">
                                                        <span className={cn('rounded-full border px-2 py-0.5', key.importable ? 'border-emerald-500/25 text-emerald-600 dark:text-emerald-300' : 'border-border/60')}>
                                                            {key.importable ? '可导入' : '仅展示'}
                                                        </span>
                                                        <span className="rounded-full border border-border/60 px-2 py-0.5">{key.status || '状态 -'}</span>
                                                        <span className="rounded-full border border-border/60 px-2 py-0.5">模型 {key.allowed_models?.length || '全部'}</span>
                                                        <span className="rounded-full border border-border/60 px-2 py-0.5">分组 {key.groups?.length || 0}</span>
                                                        <span className="rounded-full border border-border/60 px-2 py-0.5">协议 {key.request_capabilities?.length || '不限'}</span>
                                                    </div>
                                                    <MiniPills values={key.allowed_models} limit={3} tone="primary" />
                                                    <MiniPills values={key.groups} limit={3} />
                                                    {(typeof key.quota === 'number' || typeof key.quota_used === 'number' || key.expires_at) ? (
                                                        <div className="mt-2 truncate text-[11px] text-muted-foreground">
                                                            已用 {formatOptionalNumber(key.quota_used)} / 额度 {formatOptionalNumber(key.quota)}{key.expires_at ? ` · ${key.expires_at}` : ''}
                                                        </div>
                                                    ) : null}
                                                </div>
                                            ))}
                                            {(result.keys?.length ?? 0) === 0 ? (
                                                <div className="rounded-2xl border border-dashed border-border/60 p-4 text-xs text-muted-foreground">当前检查结果没有返回可导入 Key。</div>
                                            ) : null}
                                        </div>
                                    </div>
                                </div>

                                <div className="space-y-2">
                                    <div className="flex items-center justify-between text-xs font-medium text-muted-foreground">
                                        <span>服务商分组</span>
                                        {(result.groups?.length ?? 0) > 8 ? <span>+{(result.groups?.length ?? 0) - 8}</span> : null}
                                    </div>
                                    <div className="max-h-72 overflow-y-auto pr-1">
                                        <div className="grid gap-2 md:grid-cols-2">
                                            {(result.groups ?? []).map((group, index) => (
                                                <div key={`${group.id || group.name}-${index}`} className="rounded-2xl border border-border/60 bg-muted/15 p-3 text-xs">
                                                    <div className="flex items-center justify-between gap-2">
                                                        <span className="truncate font-medium text-card-foreground">{groupLabel(group)}</span>
                                                        <span className="shrink-0 text-[11px] text-muted-foreground">{group.platform || group.source || '-'}</span>
                                                    </div>
                                                    <div className="mt-2 flex flex-wrap gap-1.5 text-[11px] text-muted-foreground">
                                                        <span className="rounded-full border border-border/60 px-2 py-0.5">{group.status || '状态 -'}</span>
                                                        <span className="rounded-full border border-border/60 px-2 py-0.5">倍率 {group.rate_multiplier || '-'}</span>
                                                        <span className="rounded-full border border-border/60 px-2 py-0.5">模型 {group.models?.length || 0}</span>
                                                    </div>
                                                    <MiniPills values={group.models} limit={4} tone="primary" />
                                                    <MiniPills values={group.request_capabilities} limit={3} />
                                                </div>
                                            ))}
                                            {(result.groups?.length ?? 0) === 0 ? (
                                                <div className="rounded-2xl border border-dashed border-border/60 p-4 text-xs text-muted-foreground">未读取到分组目录。</div>
                                            ) : null}
                                        </div>
                                    </div>
                                </div>

                                <div className="space-y-2">
                                    <div className="flex items-center justify-between text-xs font-medium text-muted-foreground">
                                        <span>价格覆盖</span>
                                        {(result.price_candidates?.length ?? 0) > 10 ? <span>+{(result.price_candidates?.length ?? 0) - 10}</span> : null}
                                    </div>
                                    <div className="max-h-72 overflow-y-auto pr-1">
                                        <div className="grid gap-2 md:grid-cols-2">
                                            {(result.price_candidates ?? []).map((item, index) => (
                                                <div key={`${item.name}-${index}`} className="rounded-2xl border border-border/60 bg-muted/15 p-3 text-xs">
                                                    <div className="truncate font-medium text-card-foreground">{item.name}</div>
                                                    <div className="mt-2 grid grid-cols-4 gap-1 text-center text-[10px] text-muted-foreground">
                                                        <span className="rounded-lg bg-background/60 px-1 py-1">入 {formatPriceValue(item.input)}</span>
                                                        <span className="rounded-lg bg-background/60 px-1 py-1">出 {formatPriceValue(item.output)}</span>
                                                        <span className="rounded-lg bg-background/60 px-1 py-1">读 {formatPriceValue(item.cache_read)}</span>
                                                        <span className="rounded-lg bg-background/60 px-1 py-1">写 {formatPriceValue(item.cache_write)}</span>
                                                    </div>
                                                    <div className="mt-2 truncate text-[11px] text-muted-foreground">{item.price_source || item.cache_policy || '候选'}</div>
                                                </div>
                                            ))}
                                            {(result.price_candidates?.length ?? 0) === 0 ? (
                                                <div className="rounded-2xl border border-dashed border-border/60 p-4 text-xs text-muted-foreground">未读取到价格候选。</div>
                                            ) : null}
                                        </div>
                                    </div>
                                </div>

                                <div className="space-y-2">
                                    <div className="flex items-center justify-between text-xs font-medium text-muted-foreground">
                                        <span>订阅 / 余额</span>
                                        {(result.subscriptions?.length ?? 0) > 8 ? <span>+{(result.subscriptions?.length ?? 0) - 8}</span> : null}
                                    </div>
                                    <div className="max-h-72 overflow-y-auto pr-1">
                                        <div className="grid gap-2 md:grid-cols-2">
                                            {(result.subscriptions ?? []).map((item, index) => (
                                                <div key={`${subscriptionLabel(item)}-${index}`} className="rounded-2xl border border-border/60 bg-muted/15 p-3 text-xs">
                                                    <div className="truncate font-medium text-card-foreground">{subscriptionLabel(item)}</div>
                                                    <div className="mt-2 grid grid-cols-3 gap-1 text-[10px] text-muted-foreground">
                                                        <span className="truncate rounded-lg bg-background/60 px-2 py-1">{item.status || '-'}</span>
                                                        <span className="truncate rounded-lg bg-background/60 px-2 py-1">余额 {formatOptionalNumber(item.balance)}</span>
                                                        <span className="truncate rounded-lg bg-background/60 px-2 py-1">{item.expires_at || '到期 -'}</span>
                                                    </div>
                                                </div>
                                            ))}
                                            {(result.subscriptions?.length ?? 0) === 0 ? (
                                                <div className="rounded-2xl border border-dashed border-border/60 p-4 text-xs text-muted-foreground">未读取到订阅明细。</div>
                                            ) : null}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        ) : null}
                    </section>
                ) : null}
            </div>
        </div>
    );
}
