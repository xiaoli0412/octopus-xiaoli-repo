import { useCallback, useMemo, useState } from 'react';
import {
    Trash2,
    CheckCircle2,
    XCircle,
    FileText,
    DollarSign,
    Clock,
    Activity,
    TrendingUp,
    Globe,
    Key,
    Loader2,
    FlaskConical,
    PencilLine
} from 'lucide-react';
import { normalizeKeyManagementMode, normalizeKeyRoutingPolicy, useRouteTargetOverrideList, useTestChannelModelsByConfig, useUpdateChannel, useDeleteChannel, type Channel, type RouteTargetOverride, type TestModelResult, type UpdateChannelRequest } from '@/api/endpoints/channel';
import {
    MorphingDialogTitle,
    MorphingDialogDescription,
    MorphingDialogClose,
    useMorphingDialog,
} from '@/components/ui/morphing-dialog';
import { Tabs, TabsList, TabsTrigger, TabsContents, TabsContent } from '@/components/animate-ui/components/animate/tabs';
import { type StatsMetricsFormatted } from '@/api/endpoints/stats';
import { useTranslations } from 'next-intl';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { HelpHint } from '@/components/common/HelpHint';
import { ChannelForm, type ChannelFormData } from './Form';
import { ModelTabContent } from './ModelTabContent';
import { buildChannelKeyLabelMap, getChannelKeyLabel } from './key-label';
import { formatMoney } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { getBillingModeKey, getChannelKeySourceTypeKey, getProbePolicyKey } from '@/lib/ui-labels';
import { formatDateTimeByLocale } from '@/lib/locale';
import { useSettingStore } from '@/stores/setting';

export function CardContent({ channel, stats }: { channel: Channel; stats: StatsMetricsFormatted }) {
    const { setIsOpen } = useMorphingDialog();
    const locale = useSettingStore((state) => state.locale);
    const updateChannel = useUpdateChannel();
    const deleteChannel = useDeleteChannel();
    const routeTargetOverrides = useRouteTargetOverrideList(channel.id);
    const testModelsByConfig = useTestChannelModelsByConfig();
    const keyManagementMode = normalizeKeyManagementMode(channel.key_management_mode);
    const keyRoutingPolicy = normalizeKeyRoutingPolicy(channel.key_routing_policy);
    const [isEditing, setIsEditing] = useState(false);
    const [isConfirmingDelete, setIsConfirmingDelete] = useState(false);
    const [keyFilter, setKeyFilter] = useState('');
    const [testingKeyId, setTestingKeyId] = useState<number | null>(null);
    const [keyTestResults, setKeyTestResults] = useState<Record<number, TestModelResult[]>>({});
    const [focusKeyId, setFocusKeyId] = useState<number | null>(null);
    const [formData, setFormData] = useState<ChannelFormData>({
        name: channel.name,
        type: channel.type,
        enabled: channel.enabled,
        key_management_mode: normalizeKeyManagementMode(channel.key_management_mode),
        key_routing_policy: normalizeKeyRoutingPolicy(channel.key_routing_policy),
        base_urls: channel.base_urls?.length ? channel.base_urls : [{ url: '', delay: 0 }],
        custom_header: channel.custom_header ?? [],
        channel_proxy: channel.channel_proxy ?? '',
        param_override: channel.param_override ?? '',
        keys: channel.keys.length > 0
            ? channel.keys.map((k) => ({
                id: k.id,
                enabled: k.enabled,
                channel_key: k.channel_key,
                status_code: k.status_code,
                last_use_time_stamp: k.last_use_time_stamp,
                total_cost: k.total_cost,
                source_type: (k.source_type ?? 'unknown') as ChannelFormData['keys'][number]['source_type'],
                remark: k.remark,
                allowed_models: k.allowed_models ?? '',
                request_capabilities: k.request_capabilities ?? '',
            }))
            : [{ enabled: true, channel_key: '', source_type: 'unknown', remark: '', allowed_models: '', request_capabilities: '' }],
        model: channel.model,
        custom_model: channel.custom_model,
        proxy: channel.proxy,
        auto_sync: channel.auto_sync,
        auto_group: channel.auto_group,
        match_regex: channel.match_regex ?? '',
    });
    const t = useTranslations('channel.detail');
    const tRouteTarget = useTranslations('setting.llmRouteTarget');
    const keyLabelByID = useMemo(() => {
        return buildChannelKeyLabelMap(channel.keys ?? [], { fallbackLabel: t('keyFallbackLabel') });
    }, [channel.keys, t]);

    const formatSourceTypeLabel = useCallback((value?: string | null) => {
        return tRouteTarget(`sourceTypeOptions.${getChannelKeySourceTypeKey(value)}`);
    }, [tRouteTarget]);

    const formatBillingModeLabel = useCallback((value?: string | null) => {
        return tRouteTarget(`billingModeOptions.${getBillingModeKey(value)}`);
    }, [tRouteTarget]);

    const formatProbePolicyLabel = useCallback((value?: string | null) => {
        return tRouteTarget(`probePolicyOptions.${getProbePolicyKey(value)}`);
    }, [tRouteTarget]);

    const formatRouteTargetSummary = useCallback((row: { channel_key_id: number; model_name: string; billing_mode: string; probe_policy: string; probe_interval_seconds: number; probe_concurrency_limit: number; }) => {
        return t('actions.routeTargetSummary', {
            key: keyLabelByID.get(row.channel_key_id) ?? `${t('keyFallbackLabel')} #${row.channel_key_id}`,
            model: row.model_name,
            billing: formatBillingModeLabel(row.billing_mode),
            probe: formatProbePolicyLabel(row.probe_policy),
            interval: row.probe_interval_seconds,
            concurrency: row.probe_concurrency_limit,
        });
    }, [formatBillingModeLabel, formatProbePolicyLabel, keyLabelByID, t]);

    const enabledKeyCount = useMemo(() => (channel.keys ?? []).filter((key) => key.enabled).length, [channel.keys]);
    const aiDynamicScopedModelsCount = useMemo(() => {
        const keys = channel.keys ?? [];
        const scoped = new Set<string>();
        keys.forEach((key) => {
            const models = (key.allowed_models ?? '')
                .split(',')
                .map((item) => item.trim())
                .filter(Boolean);
            models.forEach((model) => scoped.add(model));
        });
        return scoped.size;
    }, [channel.keys]);
    const keyReadinessSummary = useMemo(() => {
        const keys = channel.keys ?? [];
        return {
            ready: keys.filter((key) => key.channel_key.trim()).length,
            pending: keys.filter((key) => !key.channel_key.trim()).length,
            unchecked: keys.filter((key) => !key.status_code).length,
            attention: keys.filter((key) => key.status_code > 0 && key.status_code !== 200).length,
        };
    }, [channel.keys]);

    const getKeyStatusMeta = useCallback((statusCode: number) => {
        if (!statusCode) {
            return {
                label: t('statusBadge.notChecked'),
                className: 'bg-muted text-muted-foreground border-muted-foreground/20',
            };
        }
        if (statusCode === 200) {
            return {
                label: t('statusBadge.available'),
                className: 'bg-green-500/15 text-green-700 dark:text-green-400 border-green-500/20',
            };
        }
        if (statusCode === 401 || statusCode === 403) {
            return {
                label: t('statusBadge.authFailed'),
                className: 'bg-red-500/15 text-red-700 dark:text-red-400 border-red-500/20',
            };
        }
        if (statusCode === 429) {
            return {
                label: t('statusBadge.rateLimited'),
                className: 'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/20',
            };
        }
        if (statusCode >= 500) {
            return {
                label: t('statusBadge.upstreamError'),
                className: 'bg-red-500/15 text-red-700 dark:text-red-400 border-red-500/20',
            };
        }
        if (statusCode >= 400) {
            return {
                label: t('statusBadge.requestError'),
                className: 'bg-orange-500/15 text-orange-700 dark:text-orange-400 border-orange-500/20',
            };
        }
        return {
            label: t('statusBadge.warning'),
            className: 'bg-orange-500/15 text-orange-700 dark:text-orange-400 border-orange-500/20',
        };
    }, [t]);

    const [detailTab, setDetailTab] = useState<'stats' | 'models'>('stats');
    const currentView = isEditing ? 'editing' : 'viewing';
    const modeBadgeClass = keyManagementMode === 'classified'
        ? 'bg-sky-500/10 text-sky-700 dark:text-sky-300 border-sky-500/30'
        : 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border-emerald-500/30';
    const policyBadgeClass = keyRoutingPolicy === 'fill_priority'
        ? 'bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/30'
        : keyRoutingPolicy === 'priority_order'
            ? 'bg-violet-500/10 text-violet-700 dark:text-violet-300 border-violet-500/30'
            : 'bg-muted text-muted-foreground border-muted-foreground/20';

    const baseUrlsEqual = (a: Channel['base_urls'] | undefined, b: Channel['base_urls'] | undefined) =>
        JSON.stringify(a ?? []) === JSON.stringify(b ?? []);
    const headersEqual = (a: Channel['custom_header'] | undefined, b: Channel['custom_header'] | undefined) =>
        JSON.stringify(a ?? []) === JSON.stringify(b ?? []);

    const visibleRouteTargetOverrides = useMemo(() => {
        return (routeTargetOverrides.data ?? []).slice().sort((a, b) => {
            if (a.channel_key_id !== b.channel_key_id) return a.channel_key_id - b.channel_key_id;
            return a.model_name.localeCompare(b.model_name);
        });
    }, [routeTargetOverrides.data]);

    const routeTargetOverridesByKeyId = useMemo(() => {
        const grouped = new Map<number, RouteTargetOverride[]>();
        for (const row of visibleRouteTargetOverrides) {
            const current = grouped.get(row.channel_key_id) ?? [];
            current.push(row);
            grouped.set(row.channel_key_id, current);
        }
        return grouped;
    }, [visibleRouteTargetOverrides]);

    const visibleKeys = useMemo(() => {
        const keyword = keyFilter.trim().toLowerCase();
        if (!keyword) return channel.keys ?? [];

        return (channel.keys ?? []).filter((key) => {
            const keyRouteTargetOverrides = routeTargetOverridesByKeyId.get(key.id) ?? [];
            const statusMeta = getKeyStatusMeta(key.status_code);
            const searchableParts = [
                key.channel_key,
                key.remark,
                key.source_type,
                formatSourceTypeLabel(key.source_type),
                key.allowed_models,
                key.request_capabilities,
                statusMeta.label,
                key.enabled ? t('labels.enabledOn') : t('labels.enabledOff'),
                String(key.status_code),
                ...keyRouteTargetOverrides.flatMap((row) => [
                    row.model_name,
                    row.billing_mode,
                    formatBillingModeLabel(row.billing_mode),
                    row.probe_policy,
                    formatProbePolicyLabel(row.probe_policy),
                    String(row.probe_interval_seconds),
                    String(row.probe_concurrency_limit),
                ]),
            ];

            return searchableParts.some((part) => part?.toLowerCase().includes(keyword));
        });
    }, [channel.keys, formatBillingModeLabel, formatProbePolicyLabel, formatSourceTypeLabel, getKeyStatusMeta, keyFilter, routeTargetOverridesByKeyId, t]);
    const showKeyFilterPanel = (channel.keys?.length ?? 0) > 1 || keyFilter.trim().length > 0;

    const handleUpdate = (event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault();
        const req: UpdateChannelRequest = { id: channel.id };

        // only send changed fields to avoid accidental clears
        if (formData.name !== channel.name) req.name = formData.name;
        if (formData.type !== channel.type) req.type = formData.type;
        if (formData.enabled !== channel.enabled) req.enabled = formData.enabled;
        if (normalizeKeyManagementMode(formData.key_management_mode) !== normalizeKeyManagementMode(channel.key_management_mode)) {
            req.key_management_mode = normalizeKeyManagementMode(formData.key_management_mode);
        }
        if (normalizeKeyRoutingPolicy(formData.key_routing_policy) !== normalizeKeyRoutingPolicy(channel.key_routing_policy)) {
            req.key_routing_policy = normalizeKeyRoutingPolicy(formData.key_routing_policy);
        }
        if (!baseUrlsEqual(formData.base_urls, channel.base_urls)) {
            req.base_urls = (formData.base_urls ?? []).filter((u) => u.url.trim()).map((u) => ({
                url: u.url.trim(),
                delay: Number(u.delay || 0),
            }));
        }
        if (formData.model !== channel.model) req.model = formData.model;
        if (formData.custom_model !== channel.custom_model) req.custom_model = formData.custom_model;
        if (formData.proxy !== channel.proxy) req.proxy = formData.proxy;
        if (formData.auto_sync !== channel.auto_sync) req.auto_sync = formData.auto_sync;
        if (formData.auto_group !== channel.auto_group) req.auto_group = formData.auto_group;

        if (!headersEqual(formData.custom_header, channel.custom_header)) {
            req.custom_header = (formData.custom_header ?? [])
                .map((h) => ({ header_key: h.header_key.trim(), header_value: h.header_value }))
                .filter((h) => h.header_key && h.header_value !== '');
        }

        const nextChannelProxy = formData.channel_proxy.trim();
        const curChannelProxy = channel.channel_proxy ?? '';
        if (nextChannelProxy !== curChannelProxy) {
            // Empty string means "clear" for patch semantics; backend maps it to NULL.
            req.channel_proxy = nextChannelProxy;
        }

        const nextParamOverride = formData.param_override.trim();
        const curParamOverride = channel.param_override ?? '';
        if (nextParamOverride !== curParamOverride) {
            // Empty string means "clear" for patch semantics; backend maps it to NULL.
            req.param_override = nextParamOverride;
        }

        const nextMatchRegex = formData.match_regex.trim();
        const curMatchRegex = channel.match_regex ?? '';
        if (nextMatchRegex !== curMatchRegex) {
            // Empty string means "clear" for patch semantics; backend maps it to NULL.
            req.match_regex = nextMatchRegex;
        }

        const originalKeys = channel.keys;
        const originalByID = new Map(originalKeys.map((k) => [k.id, k]));
        const nextKeys = formData.keys ?? [];

        const nextIDs = new Set(nextKeys.filter((k) => typeof k.id === 'number').map((k) => k.id as number));
        const keys_to_delete = originalKeys.filter((k) => !nextIDs.has(k.id)).map((k) => k.id);

        const keys_to_add = nextKeys
            .filter((k) => !k.id && k.channel_key.trim())
            .map((k) => ({ enabled: k.enabled, channel_key: k.channel_key, source_type: (k.source_type ?? '').trim(), remark: k.remark ?? '', allowed_models: (k.allowed_models ?? '').trim(), request_capabilities: (k.request_capabilities ?? '').trim() }));

        const keys_to_update = nextKeys
            .filter((k) => typeof k.id === 'number' && originalByID.has(k.id as number))
            .map((k) => {
                const orig = originalByID.get(k.id as number)!;
                const u: { id: number; enabled?: boolean; channel_key?: string; source_type?: string; remark?: string; allowed_models?: string; request_capabilities?: string } = { id: k.id as number };
                if (k.enabled !== orig.enabled) u.enabled = k.enabled;
                if (k.channel_key !== orig.channel_key) u.channel_key = k.channel_key;
                if ((k.source_type ?? '') !== (orig.source_type ?? '')) u.source_type = (k.source_type ?? '').trim();
                if ((k.remark ?? '') !== orig.remark) u.remark = k.remark ?? '';
                if ((k.allowed_models ?? '') !== (orig.allowed_models ?? '')) u.allowed_models = (k.allowed_models ?? '').trim();
                if ((k.request_capabilities ?? '') !== (orig.request_capabilities ?? '')) u.request_capabilities = (k.request_capabilities ?? '').trim();
                return Object.keys(u).length > 1 ? u : null;
            })
            .filter((u) => u !== null) as Array<{ id: number; enabled?: boolean; channel_key?: string; source_type?: string; remark?: string; allowed_models?: string; request_capabilities?: string }>;

        if (keys_to_add.length > 0) req.keys_to_add = keys_to_add;
        if (keys_to_update.length > 0) req.keys_to_update = keys_to_update;
        if (keys_to_delete.length > 0) req.keys_to_delete = keys_to_delete;

        updateChannel.mutate(req, {
            onSuccess: () => {
                setIsEditing(false);
                setIsOpen(false);
            }
        });
    };

    const handleDeleteClick = () => {
        if (!isConfirmingDelete) {
            setIsConfirmingDelete(true);
            return;
        }

        setIsOpen(false);
        setTimeout(() => {
            deleteChannel.mutate(channel.id);
        }, 300);
    };

    const handleEditKey = (keyId: number) => {
        setFocusKeyId(keyId);
        setIsEditing(true);
    };

    const handleTestKey = async (key: Channel['keys'][number]) => {
        const allowedModels = (key.allowed_models ?? '')
            .split(',')
            .map((model) => model.trim())
            .filter(Boolean);
        const fallbackModels = [
            ...channel.model.split(',').map((model) => model.trim()).filter(Boolean),
            ...channel.custom_model.split(',').map((model) => model.trim()).filter(Boolean),
        ];
        const models = allowedModels.length > 0 ? allowedModels : fallbackModels;

        if (models.length === 0) {
            setKeyTestResults((prev) => ({
                ...prev,
                [key.id]: [{ model: t('test.noModelLabel'), passed: false, error: t('test.noModels') }],
            }));
            return;
        }

        setTestingKeyId(key.id);
        try {
            const results = await testModelsByConfig.mutateAsync({
                type: channel.type,
                enabled: channel.enabled,
                base_urls: (channel.base_urls ?? []).filter((url) => url.url.trim()),
                keys: [{
                    enabled: key.enabled,
                    channel_key: key.channel_key,
                    source_type: key.source_type ?? '',
                    allowed_models: key.allowed_models ?? '',
                    request_capabilities: key.request_capabilities ?? '',
                }],
                proxy: channel.proxy,
                channel_proxy: channel.channel_proxy ?? null,
                custom_header: channel.custom_header ?? [],
                key_management_mode: normalizeKeyManagementMode(channel.key_management_mode),
                key_routing_policy: normalizeKeyRoutingPolicy(channel.key_routing_policy),
                models,
            });

            setKeyTestResults((prev) => ({
                ...prev,
                [key.id]: results,
            }));
        } catch (error) {
            const message = error instanceof Error ? error.message : t('test.failed');
            setKeyTestResults((prev) => ({
                ...prev,
                [key.id]: [{ model: t('test.requestLabel'), passed: false, error: message }],
            }));
        } finally {
            setTestingKeyId((current) => (current === key.id ? null : current));
        }
    };

    return (
        <>
            <MorphingDialogTitle>
                <header className="mb-6 flex items-center justify-between">
                    <h2 className="text-2xl font-bold text-card-foreground">
                        {isEditing ? t('title.edit') : t('title.view')}
                    </h2>
                    <MorphingDialogClose
                        className="relative top-0 right-0"
                        variants={{
                            initial: { opacity: 0, scale: 0.8 },
                            animate: { opacity: 1, scale: 1 },
                            exit: { opacity: 0, scale: 0.8 }
                        }}
                    />
                </header>
            </MorphingDialogTitle>

            <MorphingDialogDescription>
                <Tabs value={isEditing ? 'editing' : currentView}>
                    {!isEditing && (
                        <div className="mb-4 flex justify-center lg:mb-5 lg:justify-start">
                            <TabsList className="relative w-full sm:w-auto">
                                <TabsTrigger
                                    value="stats"
                                    onClick={() => setDetailTab('stats')}
                                    data-state={detailTab === 'stats' ? 'active' : 'inactive'}
                                >
                                    {t('tabs.stats')}
                                </TabsTrigger>
                                <TabsTrigger
                                    value="models"
                                    onClick={() => setDetailTab('models')}
                                    data-state={detailTab === 'models' ? 'active' : 'inactive'}
                                >
                                    {t('tabs.models')}
                                </TabsTrigger>
                            </TabsList>
                        </div>
                    )}
                    <TabsContents>
                        <TabsContent value="viewing">
                            {detailTab === 'stats' ? (
                                <div data-testid={`channel-detail-view-${channel.id}`} className="max-h-[68vh] space-y-2.5 overflow-y-auto pr-1">
                                <dl className="grid grid-cols-1 gap-2 sm:grid-cols-2 xl:grid-cols-4">
                                    <div className="rounded-2xl border bg-linear-to-br from-chart-1/10 to-chart-1/5 p-2.5 sm:p-3">
                                        <dt className="mb-1.5 flex items-center gap-2 text-xs font-medium text-muted-foreground">
                                            <Activity className="size-4 text-chart-1" />
                                            {t('metrics.totalRequests')}
                                        </dt>
                                        <dd className="text-lg sm:text-xl font-bold text-chart-1 xl:text-[1.35rem]">
                                            {stats.request_count.formatted.value}
                                            <span className="text-xs font-normal ml-1 text-muted-foreground">{stats.request_count.formatted.unit}</span>
                                        </dd>
                                    </div>

                                    <div className="rounded-2xl border bg-linear-to-br from-chart-3/10 to-chart-3/5 p-2.5 sm:p-3">
                                        <dt className="mb-1.5 flex items-center gap-2 text-xs font-medium text-muted-foreground">
                                            <FileText className="size-4 text-chart-3" />
                                            {t('metrics.totalToken')}
                                        </dt>
                                        <dd className="text-lg sm:text-xl font-bold text-chart-3 xl:text-[1.35rem]">
                                            {stats.total_token.formatted.value}
                                            <span className="text-xs font-normal ml-1 text-muted-foreground">{stats.total_token.formatted.unit}</span>
                                        </dd>
                                    </div>

                                    <div className="rounded-2xl border bg-linear-to-br from-chart-5/10 to-chart-5/5 p-2.5 sm:p-3">
                                        <dt className="mb-1.5 flex items-center gap-2 text-xs font-medium text-muted-foreground">
                                            <DollarSign className="size-4 text-chart-5" />
                                            {t('metrics.totalCost')}
                                        </dt>
                                        <dd className="text-lg sm:text-xl font-bold text-chart-5 xl:text-[1.35rem]">
                                            {stats.total_cost.formatted.value}
                                            <span className="text-xs font-normal ml-1 text-muted-foreground">{stats.total_cost.formatted.unit}</span>
                                        </dd>
                                    </div>

                                    <div className="rounded-2xl border bg-linear-to-br from-primary/10 to-primary/5 p-2.5 sm:p-3">
                                        <dt className="mb-1.5 flex items-center gap-2 text-xs font-medium text-muted-foreground">
                                            <Clock className="size-4 text-primary" />
                                            {t('metrics.avgWaitTime')}
                                        </dt>
                                        <dd className="text-lg sm:text-xl font-bold text-primary xl:text-[1.35rem]">
                                            {stats.wait_time.formatted.value}
                                            <span className="text-xs font-normal ml-1 text-muted-foreground">{stats.wait_time.formatted.unit}</span>
                                        </dd>
                                    </div>
                                </dl>

                                <section data-testid={`channel-routing-section-${channel.id}`} className="space-y-2 rounded-3xl border border-border/70 bg-card p-3 shadow-sm sm:p-3 ">
                                    <h4 className="flex items-center gap-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                        <Key className="size-3.5" />
                                        {t('sections.routing')}
                                        <HelpHint ariaLabel={t('sections.routing')}>
                                            <div className="space-y-1">
                                                <div>{t(`descriptions.mode.${keyManagementMode}`)}</div>
                                                <div>{t(`descriptions.policy.${keyRoutingPolicy}`)}</div>
                                            </div>
                                        </HelpHint>
                                    </h4>
                                    <div className="space-y-2 rounded-2xl border border-border/60 bg-muted/20 p-2.5 transition-colors hover:bg-accent/5">
                                        <div className="flex flex-wrap items-center gap-2">
                                            <Badge variant="outline" className={cn('h-6 rounded-full px-2.5 text-[11px] font-medium', modeBadgeClass)}>
                                                {t('labels.keyManagementMode')}: {t(`mode.${keyManagementMode}`)}
                                            </Badge>
                                            <Badge variant="outline" className={cn('h-6 rounded-full px-2.5 text-[11px] font-medium', policyBadgeClass)}>
                                                {t('labels.keyRoutingPolicy')}: {t(`policy.${keyRoutingPolicy}`)}
                                            </Badge>
                                        </div>
                                        <div data-testid={`channel-route-target-summary-${channel.id}`} className="space-y-2 rounded-2xl border border-border/60 bg-background/80 px-3 py-2.5">
                                            <div className="flex flex-wrap items-center justify-between gap-2">
                                                <div className="text-xs font-medium text-card-foreground">{t('actions.routeTargetTitle')}</div>
                                                <Badge variant="outline">{visibleRouteTargetOverrides.length}</Badge>
                                            </div>
                                            {routeTargetOverrides.isLoading ? (
                                                <div className="text-xs text-muted-foreground">{t('actions.routeTargetLoading')}</div>
                                            ) : visibleRouteTargetOverrides.length > 0 ? (
                                                <div className="space-y-1.5 text-xs text-muted-foreground">
                                                    {visibleRouteTargetOverrides.slice(0, 3).map((row) => (
                                                        <div key={`${row.channel_id}-${row.channel_key_id}-${row.model_name}`} className="rounded-xl bg-muted/30 px-3 py-2 break-all">
                                                            {formatRouteTargetSummary(row)}
                                                        </div>
                                                    ))}
                                                    {visibleRouteTargetOverrides.length > 3 && (
                                                        <div className="text-[11px]">{t('actions.routeTargetPreviewMore', { total: visibleRouteTargetOverrides.length })}</div>
                                                    )}
                                                </div>
                                            ) : (
                                                <div className="text-xs text-muted-foreground">{t('actions.routeTargetEmpty')}</div>
                                            )}
                                        </div>
                                    </div>
                                </section>

                                <section data-testid={`channel-ai-dynamic-section-${channel.id}`} className="space-y-2 rounded-3xl border border-border/70 bg-card p-3 shadow-sm sm:p-3">
                                    <h4 className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                                        <Activity className="size-3.5" />
                                        {t('sections.aiDynamic')}
                                        <HelpHint ariaLabel={t('sections.aiDynamic')}>{t('actions.aiDynamicHint')}</HelpHint>
                                    </h4>
                                    <div className="rounded-2xl border border-border/60 bg-muted/20 p-3">
                                        <div className="flex items-center justify-between gap-3">
                                            <div className="text-sm font-medium text-card-foreground">{t('actions.aiDynamicSummary')}</div>
                                            <Badge variant="outline">{t(`mode.${keyManagementMode}`)}</Badge>
                                        </div>
                                        <div className="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-2">
                                            <div className="rounded-xl border border-border/60 bg-background/80 px-3 py-2">
                                                <div className="text-[11px] text-muted-foreground">{t('labels.aiDynamicEnabledKeys')}</div>
                                                <div className="mt-1 text-sm font-medium text-card-foreground">{enabledKeyCount}</div>
                                            </div>
                                            <div className="rounded-xl border border-border/60 bg-background/80 px-3 py-2">
                                                <div className="text-[11px] text-muted-foreground">{t('labels.aiDynamicScopedModels')}</div>
                                                <div className="mt-1 text-sm font-medium text-card-foreground">{aiDynamicScopedModelsCount > 0 ? t('actions.aiDynamicCoverage', { count: aiDynamicScopedModelsCount }) : t('actions.aiDynamicScopeAll')}</div>
                                            </div>
                                            <div className="rounded-xl border border-border/60 bg-background/80 px-3 py-2">
                                                <div className="text-[11px] text-muted-foreground">{t('labels.aiDynamicOverrides')}</div>
                                                <div className="mt-1 text-sm font-medium text-card-foreground">{t('actions.aiDynamicOverridesValue', { count: visibleRouteTargetOverrides.length })}</div>
                                            </div>
                                            <div className="rounded-xl border border-border/60 bg-background/80 px-3 py-2">
                                                <div className="text-[11px] text-muted-foreground">{t('labels.aiDynamicHealth')}</div>
                                                <div className="mt-1 text-sm font-medium text-card-foreground">{channel.enabled ? t('actions.aiDynamicHealthy') : t('actions.aiDynamicPaused')}</div>
                                            </div>
                                        </div>
                                    </div>
                                </section>

                                {/* 请求详情 */}
                                <section className="space-y-2 rounded-3xl border border-border/70 bg-card p-3 shadow-sm sm:p-3.5 ">
                                    <h4 className="flex items-center gap-2 text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                                        <TrendingUp className="size-3.5" />
                                        {t('sections.requests')}
                                    </h4>
                                    <dl className="grid gap-2 grid-cols-1">
                                        <div className="rounded-2xl border border-border/60 bg-background/80 p-3 transition-colors hover:bg-accent/5 sm:p-4">
                                            <dt className="flex items-center gap-2 mb-2 text-xs text-muted-foreground">
                                                <CheckCircle2 className="size-4 text-accent" />
                                                {t('metrics.successRequests')}
                                            </dt>
                                            <dd className="text-2xl font-bold text-accent">
                                                {stats.request_success.formatted.value}
                                                <span className="text-sm font-normal ml-1 text-muted-foreground">{stats.request_success.formatted.unit}</span>
                                            </dd>
                                        </div>

                                        <div className="rounded-2xl border border-border/60 bg-background/80 p-3 transition-colors hover:bg-accent/5 sm:p-4">
                                            <dt className="flex items-center gap-2 mb-2 text-xs text-muted-foreground">
                                                <XCircle className="size-4 text-destructive" />
                                                {t('metrics.failedRequests')}
                                            </dt>
                                            <dd className="text-2xl font-bold text-destructive">
                                                {stats.request_failed.formatted.value}
                                                <span className="text-sm font-normal ml-1 text-muted-foreground">{stats.request_failed.formatted.unit}</span>
                                            </dd>
                                        </div>
                                    </dl>
                                </section>

                                {/* Token 使用 */}
                                <section className="space-y-2 rounded-3xl border border-border/70 bg-card p-3 shadow-sm sm:p-3.5 ">
                                    <h4 className="flex items-center gap-2 text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                                        <FileText className="size-3.5" />
                                        {t('sections.tokens')}
                                    </h4>
                                    <dl className="grid gap-2 grid-cols-1">
                                        <div className="rounded-2xl border border-border/60 bg-background/80 p-3 transition-colors hover:bg-accent/5 sm:p-4">
                                            <dt className="flex items-center gap-2 mb-2 text-xs text-muted-foreground">
                                                <div className="size-2 rounded-full bg-chart-1" />
                                                {t('metrics.inputToken')}
                                            </dt>
                                            <dd className="text-2xl font-bold text-card-foreground">
                                                {stats.input_token.formatted.value}
                                                <span className="text-sm font-normal ml-1 text-muted-foreground">{stats.input_token.formatted.unit}</span>
                                            </dd>
                                        </div>

                                        <div className="rounded-2xl border border-border/60 bg-background/80 p-3 transition-colors hover:bg-accent/5 sm:p-4">
                                            <dt className="flex items-center gap-2 mb-2 text-xs text-muted-foreground">
                                                <div className="size-2 rounded-full bg-chart-3" />
                                                {t('metrics.outputToken')}
                                            </dt>
                                            <dd className="text-2xl font-bold text-card-foreground">
                                                {stats.output_token.formatted.value}
                                                <span className="text-sm font-normal ml-1 text-muted-foreground">{stats.output_token.formatted.unit}</span>
                                            </dd>
                                        </div>
                                    </dl>
                                </section>

                                {/* 成本详情 */}
                                <section className="space-y-2 rounded-3xl border border-border/70 bg-card p-3 shadow-sm sm:p-3.5 ">
                                    <h4 className="flex items-center gap-2 text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                                        <DollarSign className="size-3.5" />
                                        {t('sections.costs')}
                                    </h4>
                                    <dl className="grid gap-2 grid-cols-1">
                                        <div className="rounded-2xl border border-border/60 bg-background/80 p-3 transition-colors hover:bg-accent/5 sm:p-4">
                                            <dt className="flex items-center gap-2 mb-2 text-xs text-muted-foreground">
                                                <div className="size-2 rounded-full bg-chart-2" />
                                                {t('metrics.inputCost')}
                                            </dt>
                                            <dd className="text-2xl font-bold text-card-foreground">
                                                {stats.input_cost.formatted.value}
                                                <span className="text-sm font-normal ml-1 text-muted-foreground">{stats.input_cost.formatted.unit}</span>
                                            </dd>
                                        </div>

                                        <div className="rounded-2xl border border-border/60 bg-background/80 p-3 transition-colors hover:bg-accent/5 sm:p-4">
                                            <dt className="flex items-center gap-2 mb-2 text-xs text-muted-foreground">
                                                <div className="size-2 rounded-full bg-chart-5" />
                                                {t('metrics.outputCost')}
                                            </dt>
                                            <dd className="text-2xl font-bold text-card-foreground">
                                                {stats.output_cost.formatted.value}
                                                <span className="text-sm font-normal ml-1 text-muted-foreground">{stats.output_cost.formatted.unit}</span>
                                            </dd>
                                        </div>
                                    </dl>
                                </section>

                                {/* Base URLs */}
                                <section className="space-y-2 rounded-3xl border border-border/70 bg-card p-3 shadow-sm sm:p-3.5 ">
                                    <h4 className="flex items-center gap-2 text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                                        <Globe className="size-3.5" />
                                        {t('sections.baseUrls')}
                                    </h4>
                                    <div className="overflow-hidden rounded-2xl border border-border/60 bg-background/80">
                                        {channel.base_urls?.map((url, i) => (
                                            <div key={i} className="flex items-center justify-between gap-3 border-b border-border/60 p-3 transition-colors hover:bg-accent/5 last:border-0 sm:p-4">
                                                <div className="flex flex-col gap-1 min-w-0">
                                                    <span className="font-mono text-sm truncate select-all">{url.url}</span>
                                                </div>
                                                <Badge
                                                    variant="secondary"
                                                    className={cn(
                                                        "h-5 px-1.5 text-xs",
                                                        url.delay < 300
                                                            ? "bg-green-500/15 text-green-700 dark:text-green-400"
                                                            : url.delay < 1000
                                                                ? "bg-orange-500/15 text-orange-700 dark:text-orange-400"
                                                                : "bg-red-500/15 text-red-700 dark:text-red-400"
                                                    )}
                                                >
                                                    {url.delay}ms
                                                </Badge>
                                            </div>
                                        ))}
                                        {(!channel.base_urls || channel.base_urls.length === 0) && (
                                            <div className="p-4 text-sm text-muted-foreground text-center">{t('noBaseUrls')}</div>
                                        )}
                                    </div>
                                </section>

                                {/* Keys */}
                                <section className="space-y-3">
                                    <h4 className="flex items-center gap-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                        <Key className="size-3.5" />
                                        {t('sections.keys')}
                                    </h4>
                                    <div className="space-y-3 rounded-3xl border border-border/70 bg-linear-to-b from-background to-muted/20 p-3 shadow-sm sm:p-3.5 xl:p-3.5">
                                    <div className="space-y-2.5 rounded-2xl border border-border/60 bg-background/80 p-3">
                                            <div className="space-y-2">
                                                <div className="flex flex-wrap items-center gap-2">
                                                    <Badge variant="outline" className="h-5 px-1.5 text-[10px]">{t('labels.enabledOn')}: {enabledKeyCount}</Badge>
                                                    <Badge variant="outline" className="h-5 px-1.5 text-[10px]">{t('labels.allowedModelsCount')}: {visibleKeys.length}/{channel.keys?.length ?? 0}</Badge>
                                                    <Badge variant="outline" className="h-5 px-1.5 text-[10px]">{t('labels.routeTargetOverridesCount')}: {visibleRouteTargetOverrides.length}</Badge>
                                                </div>
                                                <p className="text-sm font-semibold text-card-foreground">
                                                    {t('keySummaryLine', {
                                                        total: channel.keys?.length ?? 0,
                                                        enabled: enabledKeyCount,
                                                        matched: visibleKeys.length,
                                                    })}
                                                </p>
                                            </div>

                                            {showKeyFilterPanel && <div className="space-y-2">
                                                <Input
                                                    data-testid={`channel-key-filter-${channel.id}`}
                                                    value={keyFilter}
                                                    onChange={(e) => setKeyFilter(e.target.value)}
                                                    placeholder={t('keyFilterPlaceholder')}
                                                    className="h-9 rounded-2xl bg-background"
                                                />
                                            </div>}

                                                <div className="grid grid-cols-2 gap-2 xl:grid-cols-4">
                                                    <div className="rounded-xl border border-border/60 bg-muted/20 px-3 py-2">
                                                        <div className="text-[11px] text-muted-foreground">{t('keyReadiness.ready')}</div>
                                                        <div className="mt-1 text-base font-semibold text-card-foreground">{keyReadinessSummary.ready}</div>
                                                    </div>
                                                    <div className="rounded-xl border border-border/60 bg-muted/20 px-3 py-2">
                                                        <div className="text-[11px] text-muted-foreground">{t('keyReadiness.pending')}</div>
                                                        <div className="mt-1 text-base font-semibold text-card-foreground">{keyReadinessSummary.pending}</div>
                                                    </div>
                                                    <div className="rounded-xl border border-border/60 bg-muted/20 px-3 py-2">
                                                        <div className="text-[11px] text-muted-foreground">{t('keyReadiness.unchecked')}</div>
                                                        <div className="mt-1 text-base font-semibold text-card-foreground">{keyReadinessSummary.unchecked}</div>
                                                    </div>
                                                    <div className="rounded-xl border border-border/60 bg-muted/20 px-3 py-2">
                                                        <div className="text-[11px] text-muted-foreground">{t('keyReadiness.attention')}</div>
                                                        <div className="mt-1 text-base font-semibold text-card-foreground">{keyReadinessSummary.attention}</div>
                                                    </div>
                                                </div>
                                            </div>

                                        <Accordion data-testid={`channel-key-accordion-${channel.id}`} type="multiple" className="rounded-2xl border border-border/60 bg-background/75 px-3 sm:px-4">
                                        {visibleKeys.map((key) => (
                                            <AccordionItem data-testid={`channel-key-item-${key.id}`} key={key.id} value={`key-${key.id}`} className="border-b last:border-0">
                                                {(() => {
                                                    const allowedModels = (key.allowed_models ?? '').trim()
                                                        ? (key.allowed_models ?? '')
                                                            .split(',')
                                                            .map((model) => model.trim())
                                                            .filter(Boolean)
                                                        : [];
                                                    const keyRouteTargetOverrides = routeTargetOverridesByKeyId.get(key.id) ?? [];
                                                    const keyLabel = getChannelKeyLabel(key, { fallbackLabel: t('keyFallbackLabel') });
                                                    const maskedKey = key.channel_key.length > 10
                                                        ? `${key.channel_key.slice(0, 4)}...${key.channel_key.slice(-4)}`
                                                        : key.channel_key;
                                                    const testResults = keyTestResults[key.id] ?? [];
                                                    const isTestingThisKey = testingKeyId === key.id;
                                                    const statusMeta = getKeyStatusMeta(key.status_code);
                                                    const costSummary = formatMoney(key.total_cost).formatted;

                                                    return (
                                                        <>
                                                            <AccordionTrigger data-testid={`channel-key-trigger-${key.id}`} className="rounded-2xl py-2 hover:no-underline hover:bg-accent/5 focus-visible:ring-primary/20">
                                                                <div className="flex min-w-0 flex-1 flex-col gap-2 text-left">
                                                                    <div className="flex min-w-0 items-start gap-3 text-left">
                                                                        <div className={cn("mt-1 size-2 shrink-0 rounded-full", key.enabled ? "bg-emerald-500" : "bg-destructive")} />
                                                                        <div className="min-w-0 flex-1 space-y-2">
                                                                            <div className="flex flex-wrap items-center gap-2">
                                                                                <span className="min-w-0 truncate text-sm font-semibold text-card-foreground">{keyLabel}</span>
                                                                                <Badge
                                                                                    variant="outline"
                                                                                    aria-label={`${t('labels.enabledState')}: ${key.enabled ? t('labels.enabledOn') : t('labels.enabledOff')}`}
                                                                                    title={`${t('labels.enabledState')}: ${key.enabled ? t('labels.enabledOn') : t('labels.enabledOff')}`}
                                                                                    className={cn('h-5 px-1.5 text-[10px] font-medium', key.enabled ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300' : 'border-border bg-muted text-muted-foreground')}
                                                                                >
                                                                                    {key.enabled ? t('labels.enabledOn') : t('labels.enabledOff')}
                                                                                </Badge>
                                                                                <Badge variant="secondary" className={cn('h-5 px-1.5 text-[10px] font-medium', statusMeta.className)}>
                                                                                    {statusMeta.label}
                                                                                    {key.status_code > 0 ? ` (${key.status_code})` : ''}
                                                                                </Badge>
                                                                                {key.source_type && (
                                                                                    <Badge variant="outline" className="h-5 px-1.5 text-[10px]">
                                                                                        {formatSourceTypeLabel(key.source_type)}
                                                                                    </Badge>
                                                                                )}
                                                                            </div>
                                                                            <div className="flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
                                                                                <span>{t('labels.maskedKey')}: {maskedKey || t('keyFallbackLabel')}</span>
                                                                            </div>
                                                                        </div>
                                                                    </div>

                                                                    <div className="flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
                                                                        <span className="inline-flex items-center rounded-full border border-border/60 bg-muted/20 px-2.5 py-1">
                                                                            <span className="mr-1">{t('labels.cost')}</span>
                                                                            <span className="font-medium text-card-foreground">{costSummary.value}{costSummary.unit}</span>
                                                                        </span>
                                                                        <span className="inline-flex items-center rounded-full border border-border/60 bg-muted/20 px-2.5 py-1">
                                                                            <span className="mr-1">{t('labels.allowedModelsCount')}</span>
                                                                            <span className="font-medium text-card-foreground">{allowedModels.length > 0 ? allowedModels.length : t('allModelsShort')}</span>
                                                                        </span>
                                                                        <span className="inline-flex items-center rounded-full border border-border/60 bg-muted/20 px-2.5 py-1">
                                                                            <span className="mr-1">{t('labels.routeTargetOverridesCount')}</span>
                                                                            <span className="font-medium text-card-foreground">{keyRouteTargetOverrides.length}</span>
                                                                        </span>
                                                                    </div>
                                                                </div>
                                                            </AccordionTrigger>

                                                            <AccordionContent className="pb-2">
                                                                <div className="space-y-2">
                                                                    <div className="space-y-2">
                                                                        <div data-testid={`channel-key-models-${key.id}`} className="space-y-2">
                                                                        <div className="text-xs font-medium text-muted-foreground">
                                                                            {t('labels.allowedModels')}
                                                                        </div>
                                                                        <div className="flex flex-wrap gap-1.5">
                                                                            {allowedModels.length > 0
                                                                                ? allowedModels.map((model) => (
                                                                                        <Badge key={`${key.id}-${model}`} variant="outline" className="max-w-full text-[10px] leading-4">
                                                                                            <span className="truncate">{model}</span>
                                                                                        </Badge>
                                                                                    ))
                                                                                : (
                                                                                    <Badge variant="outline" className="text-[10px] text-muted-foreground">
                                                                                        {t('allModels')}
                                                                                    </Badge>
                                                                                )}
                                                                        </div>
                                                                        </div>

                                                                        <div className="flex flex-wrap items-center gap-2">
                                                                            <Button
                                                                                type="button"
                                                                                variant="outline"
                                                                                size="sm"
                                                                                className="h-8 w-full rounded-xl sm:w-auto"
                                                                                onClick={() => handleEditKey(key.id)}
                                                                            >
                                                                                <PencilLine className="size-4" />
                                                                                {t('test.editThisKey')}
                                                                            </Button>
                                                                            <Button
                                                                                type="button"
                                                                                variant="outline"
                                                                                size="sm"
                                                                                className="h-8 w-full rounded-xl sm:w-auto"
                                                                                onClick={() => void handleTestKey(key)}
                                                                                disabled={isTestingThisKey || testModelsByConfig.isPending}
                                                                            >
                                                                                {isTestingThisKey ? (
                                                                                    <Loader2 className="size-4 animate-spin" />
                                                                                ) : (
                                                                                    <FlaskConical className="size-4" />
                                                                                )}
                                                                                {isTestingThisKey ? t('test.testing') : t('test.runForKey')}
                                                                            </Button>
                                                                            <span className="text-[11px] text-muted-foreground sm:ml-auto">
                                                                                {t('labels.lastUsed')}: {key.last_use_time_stamp > 0 ? formatDateTimeByLocale(new Date(key.last_use_time_stamp * 1000), locale) : t('labels.neverUsed')}
                                                                            </span>
                                                                        </div>
                                                                    </div>

                                                                    <div data-testid={`channel-key-route-target-${key.id}`} className="space-y-2">
                                                                        <div className="flex flex-wrap items-center justify-between gap-2">
                                                                            <div className="text-xs font-medium text-muted-foreground">{t('actions.routeTargetTitle')}</div>
                                                                            <Badge variant="outline" className="text-[10px]">{t('labels.routeTargetOverridesCount')}: {keyRouteTargetOverrides.length}</Badge>
                                                                        </div>
                                                                        {keyRouteTargetOverrides.length > 0 ? (
                                                                            <div className="space-y-1.5 rounded-2xl border border-border/60 bg-muted/20 p-2">
                                                                                {keyRouteTargetOverrides.slice(0, 2).map((row) => (
                                                                                    <div
                                                                                        data-testid={`channel-key-route-target-row-${key.id}-${row.model_name}`}
                                                                                        key={`${row.channel_id}-${row.channel_key_id}-${row.model_name}`}
                                                                                        className="rounded-xl bg-background/80 px-3 py-1.5 text-xs text-muted-foreground break-all"
                                                                                    >
                                                                                        {formatRouteTargetSummary(row)}
                                                                                    </div>
                                                                                ))}
                                                                                {keyRouteTargetOverrides.length > 2 && (
                                                                                    <div className="text-[11px] text-muted-foreground">{t('actions.routeTargetKeyPreviewMore', { total: keyRouteTargetOverrides.length })}</div>
                                                                                )}
                                                                            </div>
                                                                        ) : (
                                                                            <div className="rounded-xl border border-dashed border-border/60 bg-background/80 px-3 py-2 text-xs text-muted-foreground">
                                                                                {t('actions.routeTargetKeyEmpty')}
                                                                            </div>
                                                                        )}
                                                                    </div>

                                                                    {testResults.length > 0 && (
                                                                        <div data-testid={`channel-key-test-results-${key.id}`} className="space-y-2 rounded-2xl border border-border/60 bg-background/80 p-2 lg:col-span-2">
                                                                            <div className="text-xs font-medium text-muted-foreground">{t('test.resultsTitle')}</div>
                                                                            <div className="space-y-2">
                                                                                {testResults.map((result) => (
                                                                                    <div key={`${key.id}-${result.model}`} className="flex flex-wrap items-center gap-2 text-xs">
                                                                                        <span className="font-mono break-all text-card-foreground">{result.model}</span>
                                                                                        <Badge
                                                                                            variant="secondary"
                                                                                            className={result.passed
                                                                                                ? 'bg-green-500/15 text-green-700 dark:text-green-400'
                                                                                                : 'bg-red-500/15 text-red-700 dark:text-red-400'}
                                                                                        >
                                                                                            {result.passed ? t('test.passed') : t('test.failed')}
                                                                                        </Badge>
                                                                                        {typeof result.delay === 'number' && (
                                                                                            <span className="text-muted-foreground">{result.delay}ms</span>
                                                                                        )}
                                                                                        {result.error && (
                                                                                            <span className="text-muted-foreground break-all">{result.error}</span>
                                                                                        )}
                                                                                    </div>
                                                                                ))}
                                                                            </div>
                                                                        </div>
                                                                    )}

                                                                </div>
                                                            </AccordionContent>
                                                        </>
                                                    );
                                                })()}
                                            </AccordionItem>
                                        ))}
                                        </Accordion>
                                        {visibleKeys.length === 0 && (channel.keys?.length ?? 0) > 0 && (
                                            <div className="p-4 text-sm text-muted-foreground text-center rounded-2xl border bg-card">{t('noKeysMatched')}</div>
                                        )}
                                        {(!channel.keys || channel.keys.length === 0) && (
                                            <div className="p-4 text-sm text-muted-foreground text-center">{t('noKeys')}</div>
                                        )}
                                    </div>
                                </section>
                            </div>
                            ) : (
                                <ModelTabContent channel={channel} />
                            )}

                            {/* 操作按钮 */}
                            <div className="grid gap-2 pt-2 sm:grid-cols-2">
                                <Button
                                    onClick={() => (isConfirmingDelete ? setIsConfirmingDelete(false) : setIsEditing(true))}
                                    variant={isConfirmingDelete ? 'secondary' : 'default'}
                                    className="h-10 w-full rounded-2xl"
                                >
                                    {isConfirmingDelete ? t('actions.cancel') : t('actions.edit')}
                                </Button>
                                <Button
                                    onClick={handleDeleteClick}
                                    disabled={deleteChannel.isPending}
                                    variant="destructive"
                                    className="h-10 w-full rounded-2xl"
                                >
                                    <Trash2 className={`size-4 transition-transform ${isConfirmingDelete ? 'scale-110' : ''}`} />
                                    {deleteChannel.isPending
                                        ? t('actions.deleting')
                                        : isConfirmingDelete
                                            ? t('actions.confirmDelete')
                                            : t('actions.delete')}
                                </Button>
                            </div>
                        </TabsContent>

                        <TabsContent value="editing">
                            <ChannelForm
                                formData={formData}
                                onFormDataChange={setFormData}
                                onSubmit={handleUpdate}
                                isPending={updateChannel.isPending}
                                submitText={t('actions.save')}
                                pendingText={t('actions.saving')}
                                onCancel={() => setIsEditing(false)}
                                cancelText={t('actions.cancel')}
                                idPrefix="channel"
                                focusKeyId={focusKeyId}
                            />
                        </TabsContent>
                    </TabsContents>
                </Tabs>
            </MorphingDialogDescription>
        </>
    );
}
