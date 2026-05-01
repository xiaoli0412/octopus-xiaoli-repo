import { AutoGroupType, CHANNEL_KEY_SOURCE_TYPES, ChannelType, normalizeKeyManagementMode, normalizeKeyRoutingPolicy, type Channel, type ChannelKeySourceType, type KeyManagementMode, type KeyRoutingPolicy, type RouteTargetOverride, useDeleteRouteTargetOverride, useFetchModel, useRouteTargetOverrideList, useTestChannelModelsByConfig, useUpsertRouteTargetOverride, type TestModelResult, useCopilotRequestDeviceCode, useCopilotPollToken, useAntigravityOAuthStart, useAntigravityOAuthPoll } from '@/api/endpoints/channel';
import { useProviders } from '@/api/endpoints/providers';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { toast } from '@/components/common/Toast';
import { useTranslations } from 'next-intl';
import { useCallback, useDeferredValue, useEffect, useMemo, useRef, useState } from 'react';
import { X, Plus, CheckCircle2, XCircle, Loader2, Info, Copy, ExternalLink, Check, Search, ChevronDown } from 'lucide-react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { BILLING_MODE_OPTIONS, PROBE_POLICY_OPTIONS } from '@/api/endpoints/model';
import { buildChannelKeyLabelMap, getChannelKeyLabel } from './key-label';
import { getBillingModeKey, getChannelKeySourceTypeKey, getProbePolicyKey } from '@/lib/ui-labels';
import { HelpHint } from '@/components/common/HelpHint';

export interface ChannelKeyFormItem {
    id?: number;
    enabled: boolean;
    channel_key: string;
    source_type?: ChannelKeySourceType | '';
    status_code?: number;
    last_use_time_stamp?: number;
    total_cost?: number;
    remark?: string;
    allowed_models?: string;
}
export interface ChannelFormData {
    name: string;
    type: ChannelType;
    key_management_mode?: KeyManagementMode;
    key_routing_policy?: KeyRoutingPolicy;
    base_urls: Channel['base_urls'];
    custom_header: Channel['custom_header'];
    channel_proxy: string;
    param_override: string;
    keys: ChannelKeyFormItem[];
    model: string;
    custom_model: string;
    enabled: boolean;
    proxy: boolean;
    auto_sync: boolean;
    auto_group: AutoGroupType;
    match_regex: string;
}
interface RouteTargetOverrideFormState {
    channel_key_id: string;
    model_name: string;
    billing_mode: string;
    probe_policy: string;
    probe_interval_seconds: string;
    probe_concurrency_limit: string;
}

export interface ChannelFormProps {
    formData: ChannelFormData;
    onFormDataChange: (data: ChannelFormData) => void;
    onSubmit: (event: React.FormEvent<HTMLFormElement>) => void;
    isPending: boolean;
    submitText: string;
    pendingText: string;
    onCancel?: () => void;
    cancelText?: string;
    idPrefix?: string;
    channelId?: number;
    focusKeyId?: number | null;
}

import {
    Accordion,
    AccordionContent,
    AccordionItem,
    AccordionTrigger,
} from "@/components/ui/accordion";

type KeyModelDialogState = {
    keyIndex: number;
    models: string[];
    selected: Set<string>;
};

function splitAllowedModels(value?: string) {
    return (value ?? '')
        .split(',')
        .map((model) => model.trim())
        .filter(Boolean);
}

function appendUniqueModels(existing: string[], incoming: string[]) {
    const next = [...existing];
    const seen = new Set(existing);

    for (const model of incoming) {
        const normalized = model.trim();
        if (!normalized || seen.has(normalized)) continue;
        seen.add(normalized);
        next.push(normalized);
    }

    return next;
}

function maskKeyPreview(value?: string) {
    const raw = value?.trim() ?? '';
    if (!raw) return '';
    return raw.length > 10 ? `${raw.slice(0, 4)}...${raw.slice(-4)}` : raw;
}

function getErrorMessage(error: unknown) {
    if (error instanceof Error && error.message.trim()) {
        return error.message;
    }
    if (error && typeof error === 'object') {
        if ('message' in error && typeof error.message === 'string' && error.message.trim()) {
            return error.message;
        }
        if ('code' in error && typeof error.code === 'number') {
            return `Request failed (${error.code})`;
        }
    }
    return String(error);
}

export function ChannelForm({
    formData,
    onFormDataChange,
    onSubmit,
    isPending,
    submitText,
    pendingText,
    onCancel,
    cancelText,
    idPrefix = 'channel',
    channelId,
    focusKeyId,
}: ChannelFormProps) {
    const t = useTranslations('channel.form');
    const tModels = useTranslations('channel.models');
    const tDetail = useTranslations('channel.detail');
    const tRouteTarget = useTranslations('setting.llmRouteTarget');

    const flowCopy = useMemo(() => ({
        createFlowTitle: t('flow.createFlowTitle'),
        createFlowHint: t('flow.createFlowHint'),
        createFlowDesc: t('flow.createFlowDesc'),
        stepLabel: (index: number) => t('flow.stepLabel', { index }),
        basicSectionTitle: t('flow.basicSectionTitle'),
        basicSectionDesc: t('flow.basicSectionDesc'),
        basicSectionHint: t('flow.basicSectionHint'),
        keySectionTitle: t('flow.keySectionTitle'),
        keySectionDesc: t('flow.keySectionDesc'),
        modelSectionTitle: t('flow.modelSectionTitle'),
        modelSectionDesc: t('flow.modelSectionDesc'),
        modelSectionHint: t('flow.modelSectionHint'),
        advancedSectionTitle: t('flow.advancedSectionTitle'),
        advancedSectionDesc: t('flow.advancedSectionDesc'),
        advancedSectionHint: t('flow.advancedSectionHint'),
        baseUrlsDesc: t('flow.baseUrlsDesc'),
    }), [t]);

    // Fetch providers for auto-fill base_url
    const { data: providers } = useProviders();

    // Test state
    const testByConfig = useTestChannelModelsByConfig();
    const routeTargetOverrides = useRouteTargetOverrideList(channelId);
    const upsertRouteTargetOverride = useUpsertRouteTargetOverride();
    const deleteRouteTargetOverride = useDeleteRouteTargetOverride();
    const [isTesting, setIsTesting] = useState(false);
    const [testResults, setTestResults] = useState<Map<string, TestModelResult>>(new Map());
    const [showRouteTargetDialog, setShowRouteTargetDialog] = useState(false);
    const [routeTargetForm, setRouteTargetForm] = useState<RouteTargetOverrideFormState>({
        channel_key_id: '',
        model_name: '',
        billing_mode: 'unknown',
        probe_policy: 'passive_only',
        probe_interval_seconds: '3600',
        probe_concurrency_limit: '1',
    });

    // Ensure the form always shows at least 1 row for base_urls / keys / custom_header.
    // This avoids "empty list" UI and also keeps URL + APIKEY layout consistent.
    useEffect(() => {
        if (!formData.base_urls || formData.base_urls.length === 0) {
            onFormDataChange({ ...formData, key_management_mode: normalizeKeyManagementMode(formData.key_management_mode), key_routing_policy: normalizeKeyRoutingPolicy(formData.key_routing_policy), base_urls: [{ url: '', delay: 0 }] });
            return;
        }
        if (!formData.keys || formData.keys.length === 0) {
            onFormDataChange({ ...formData, keys: [{ enabled: true, channel_key: '', source_type: '', remark: '', allowed_models: '' }] });
            return;
        }
        if (!formData.custom_header || formData.custom_header.length === 0) {
            onFormDataChange({ ...formData, custom_header: [{ header_key: '', header_value: '' }] });
        }
    }, [formData, onFormDataChange]);

    // Auto-fill base_url when type changes and base_url is empty
    useEffect(() => {
        if (!providers) return;

        const provider = providers.find((p) => p.channel_type === formData.type);
        // Only auto-fill if there's exactly one base_url and it's empty
        if (provider && formData.base_urls.length === 1 && formData.base_urls[0].url === '') {
            onFormDataChange({
                ...formData,
                base_urls: [{ url: provider.base_url, delay: 0 }],
            });
        }
    }, [formData, onFormDataChange, providers]);

    const autoModels = formData.model
        ? formData.model.split(',').map((m) => m.trim()).filter(Boolean)
        : [];
    const customModels = formData.custom_model
        ? formData.custom_model.split(',').map((m) => m.trim()).filter(Boolean)
        : [];
    const [inputValue, setInputValue] = useState('');
    const inputRef = useRef<HTMLInputElement>(null);
    const keyRowRefs = useRef<Record<string, HTMLDivElement | null>>({});
    const [expandedKeyItems, setExpandedKeyItems] = useState<string[]>(['key-0']);
    const [keyFilter, setKeyFilter] = useState('');
    const [showKeyFilterPanel, setShowKeyFilterPanel] = useState(false);
    const [keyModelDrafts, setKeyModelDrafts] = useState<Record<number, string>>({});
    const [showModelSelectDialog, setShowModelSelectDialog] = useState(false);
    const [keyModelDialog, setKeyModelDialog] = useState<KeyModelDialogState | null>(null);
    const [modelDialogKeyword, setModelDialogKeyword] = useState('');
    const deferredModelDialogKeyword = useDeferredValue(modelDialogKeyword);

    const keySetupSummary = useMemo(() => {
        const keys = formData.keys ?? [];
        const total = keys.length;
        const ready = keys.filter((key) => key.channel_key.trim()).length;
        const enabled = keys.filter((key) => key.enabled).length;
        return { total, ready, enabled, pending: Math.max(0, total - ready) };
    }, [formData.keys]);
    const showKeyFilter = keySetupSummary.total > 1 || keyFilter.trim().length > 0;
    const keyFilterExpanded = showKeyFilterPanel || keyFilter.trim().length > 0;

    useEffect(() => {
        const lastIndex = Math.max(0, (formData.keys?.length ?? 1) - 1);
        setExpandedKeyItems((current) => {
            const validKeys = new Set((formData.keys ?? []).map((_, idx) => `key-${idx}`));
            const next = current.filter((item) => validKeys.has(item));
            if (next.length > 0) return next;
            return [`key-${lastIndex}`];
        });
    }, [formData.keys]);

    useEffect(() => {
        if (focusKeyId == null) return;
        const node = keyRowRefs.current[String(focusKeyId)];
        if (!node) return;
        node.scrollIntoView({ behavior: 'smooth', block: 'center' });
        const input = node.querySelector('input');
        if (input instanceof HTMLInputElement) {
            input.focus();
            input.select();
        }
    }, [focusKeyId, formData.keys]);

    // ---- GitHub Copilot Device Flow ----
    const copilotDeviceCodeRef = useRef('');
    const copilotPollIntervalRef = useRef(5);
    const copilotTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const [copilotStatus, setCopilotStatus] = useState<
        'idle' | 'loading' | 'waiting' | 'authorized' | 'expired' | 'denied' | 'error'
    >('idle');
    const [copilotUserCode, setCopilotUserCode] = useState('');
    const [copilotVerificationUri, setCopilotVerificationUri] = useState('');

    // Keep stable refs to avoid stale closures in async poll callbacks
    const formDataRef = useRef(formData);
    useEffect(() => { formDataRef.current = formData; }, [formData]);
    const onFormDataChangeRef = useRef(onFormDataChange);
    useEffect(() => { onFormDataChangeRef.current = onFormDataChange; }, [onFormDataChange]);

    const copilotRequestDeviceCode = useCopilotRequestDeviceCode();
    const copilotPollToken = useCopilotPollToken();

    // ---- Antigravity OAuth Web Flow ----
    const antigravityStateRef = useRef('');
    const antigravityTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const [antigravityStatus, setAntigravityStatus] = useState<'idle' | 'loading' | 'waiting' | 'authorized' | 'error'>('idle');
    const [antigravityError, setAntigravityError] = useState('');
    const antigravityOAuthStart = useAntigravityOAuthStart();
    const antigravityOAuthPoll = useAntigravityOAuthPoll();

    // Cleanup timer on unmount
    useEffect(() => {
        return () => {
            if (copilotTimerRef.current) clearTimeout(copilotTimerRef.current);
            if (antigravityTimerRef.current) clearTimeout(antigravityTimerRef.current);
        };
    }, []);

    // Reset device flow when switching away from GitHub Copilot type
    useEffect(() => {
        if (formData.type !== ChannelType.GithubCopilot) {
            if (copilotTimerRef.current) {
                clearTimeout(copilotTimerRef.current);
                copilotTimerRef.current = null;
            }
            setCopilotStatus('idle');
            copilotDeviceCodeRef.current = '';
        }
    }, [formData.type]);

    useEffect(() => {
        if (formData.type !== ChannelType.Antigravity) {
            if (antigravityTimerRef.current) {
                clearTimeout(antigravityTimerRef.current);
                antigravityTimerRef.current = null;
            }
            antigravityStateRef.current = '';
            setAntigravityStatus('idle');
            setAntigravityError('');
        }
    }, [formData.type]);

    const startPollLoop = useCallback(() => {
        const poll = async () => {
            if (!copilotDeviceCodeRef.current) return;
            try {
                const result = await copilotPollToken.mutateAsync(copilotDeviceCodeRef.current);
                if (result.access_token) {
                    setCopilotStatus('authorized');
                    onFormDataChangeRef.current({
                        ...formDataRef.current,
                        base_urls: [{ url: 'https://api.githubcopilot.com', delay: 0 }],
                        keys: [{ enabled: true, channel_key: result.access_token, source_type: '', remark: '', allowed_models: '' }],
                    });
                    return; // Stop polling
                }
                if (result.error === 'slow_down') {
                    copilotPollIntervalRef.current += 5;
                } else if (result.error === 'expired_token') {
                    setCopilotStatus('expired');
                    return;
                } else if (result.error === 'access_denied') {
                    setCopilotStatus('denied');
                    return;
                } else if (result.error && result.error !== 'authorization_pending') {
                    setCopilotStatus('error');
                    return;
                }
            } catch {
                // network error, retry
            }
            copilotTimerRef.current = setTimeout(poll, copilotPollIntervalRef.current * 1000);
        };
        copilotTimerRef.current = setTimeout(poll, copilotPollIntervalRef.current * 1000);
    }, [copilotPollToken]);

    const handleCopilotStartAuth = async () => {
        if (copilotTimerRef.current) {
            clearTimeout(copilotTimerRef.current);
            copilotTimerRef.current = null;
        }
        copilotDeviceCodeRef.current = '';
        copilotPollIntervalRef.current = 5;
        setCopilotStatus('loading');
        try {
            const result = await copilotRequestDeviceCode.mutateAsync();
            copilotDeviceCodeRef.current = result.device_code;
            copilotPollIntervalRef.current = result.interval || 5;
            setCopilotUserCode(result.user_code);
            setCopilotVerificationUri(result.verification_uri);
            setCopilotStatus('waiting');
            startPollLoop();
        } catch {
            setCopilotStatus('error');
            toast.error(t('copilotError'));
        }
    };
    // ---- End GitHub Copilot Device Flow ----

    const startAntigravityPollLoop = useCallback(() => {
        const poll = async () => {
            if (!antigravityStateRef.current) return;
            try {
                const result = await antigravityOAuthPoll.mutateAsync(antigravityStateRef.current);
                if (result.status === 'authorized' && result.access_token) {
                    setAntigravityStatus('authorized');
                    const currentBaseUrls = formDataRef.current.base_urls?.filter((u) => u.url.trim()) ?? [];
                    onFormDataChangeRef.current({
                        ...formDataRef.current,
                        base_urls: currentBaseUrls.length > 0 ? currentBaseUrls : [{ url: 'https://cloudcode-pa.googleapis.com', delay: 0 }],
                        keys: [{ enabled: true, channel_key: result.access_token, source_type: '', remark: '', allowed_models: '' }],
                    });
                    return;
                }
                if (result.status === 'failed') {
                    setAntigravityStatus('error');
                    setAntigravityError(result.error || t('antigravityAuthFailed'));
                    return;
                }
            } catch {
                // keep polling on temporary failures
            }
            antigravityTimerRef.current = setTimeout(poll, 2000);
        };
        antigravityTimerRef.current = setTimeout(poll, 2000);
    }, [antigravityOAuthPoll, t]);

    const handleAntigravityStartAuth = async () => {
        if (antigravityTimerRef.current) {
            clearTimeout(antigravityTimerRef.current);
            antigravityTimerRef.current = null;
        }
        antigravityStateRef.current = '';
        setAntigravityError('');
        setAntigravityStatus('loading');
        try {
            const result = await antigravityOAuthStart.mutateAsync();
            antigravityStateRef.current = result.state;
            setAntigravityStatus('waiting');
            window.open(result.auth_url, '_blank', 'noopener,noreferrer');
            startAntigravityPollLoop();
        } catch (error) {
            const message = error instanceof Error ? error.message : t('antigravityAuthFailed');
            setAntigravityStatus('error');
            setAntigravityError(message);
            toast.error(t('antigravityAuthFailed'), { description: message });
        }
    };
    // ---- End Antigravity OAuth Web Flow ----

    const fetchModel = useFetchModel();

    const availableBaseUrl = formData.base_urls?.find((u) => u.url.trim())?.url.trim() || '';
    const availableRequestKey = formData.keys.find((k) => k.channel_key.trim())?.channel_key.trim() || '';
    const canOpenGlobalModelDialog = Boolean(availableBaseUrl && availableRequestKey);

    const buildModelFetchRequest = useCallback((requestKey: string) => ({
        type: formData.type,
        base_url: availableBaseUrl,
        key: requestKey.trim(),
        proxy: formData.proxy,
        channel_proxy: formData.channel_proxy?.trim() || null,
        custom_header: formData.custom_header?.filter((h) => h.header_key.trim()) || [],
    }), [availableBaseUrl, formData.channel_proxy, formData.custom_header, formData.proxy, formData.type]);

    const normalizeFetchedModels = useCallback((data: string[] | undefined) => {
        return Array.from(new Set((data ?? []).map((m) => m.trim()).filter(Boolean)));
    }, []);

    const updateModels = (nextAuto: string[], nextCustom: string[]) => {
        const model = nextAuto.join(',');
        const custom_model = nextCustom.join(',');
        if (formData.model === model && formData.custom_model === custom_model) return;
        onFormDataChange({ ...formData, model, custom_model });
    };
    const handleConfirmModelSelect = () => {
        if (!keyModelDialog) return;
        const selected = Array.from(keyModelDialog.selected);
        const dialogModelSet = new Set(keyModelDialog.models);

        // 保留当前请求结果之外的模型；当前请求结果内的勾选状态以弹窗选择为准。
        const nextAuto = autoModels.filter((model) => !dialogModelSet.has(model));
        const nextCustom = Array.from(new Set([
            ...customModels.filter((model) => !dialogModelSet.has(model)),
            ...selected,
        ]));

        updateModels(nextAuto, nextCustom);
        setKeyModelDialog(null);
        setShowModelSelectDialog(false);
    };

    const handleConfirmKeyModelSelect = () => {
        if (!keyModelDialog) return;
        const targetKey = formData.keys[keyModelDialog.keyIndex];
        if (!targetKey) return;
        const nextAllowedModels = Array.from(keyModelDialog.selected).join(',');
        handleUpdateKey(keyModelDialog.keyIndex, { allowed_models: nextAllowedModels });
        setKeyModelDialog(null);
        setShowModelSelectDialog(false);
    };

    const handleAddModel = (model: string) => {
        const trimmedModel = model.trim();
        if (trimmedModel && !customModels.includes(trimmedModel) && !autoModels.includes(trimmedModel)) {
            updateModels(autoModels, [...customModels, trimmedModel]);
        }
        setInputValue('');
    };

    const handleRemoveAutoModel = (model: string) => {
        updateModels(autoModels.filter(m => m !== model), customModels);
    };

    const handleRemoveCustomModel = (model: string) => {
        updateModels(autoModels, customModels.filter(m => m !== model));
    };

    const handleInputKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            if (inputValue.trim()) handleAddModel(inputValue);
        }
    };

    const openGlobalModelDialog = (models: string[]) => {
        setModelDialogKeyword('');
        setKeyModelDialog({
            keyIndex: -1,
            models,
            selected: new Set(models.filter((model) => allModels.includes(model))),
        });
        setShowModelSelectDialog(true);
    };

    const openKeyModelDialog = (keyIndex: number, models: string[]) => {
        const allowedModels = splitAllowedModels(formData.keys[keyIndex]?.allowed_models);
        setModelDialogKeyword('');
        setKeyModelDialog({
            keyIndex,
            models,
            selected: new Set(models.filter((model) => allowedModels.includes(model))),
        });
        setShowModelSelectDialog(true);
    };

    const handleFetchModelsForKey = async (keyIndex: number) => {
        const targetKey = formData.keys[keyIndex];
        if (!targetKey?.channel_key.trim() || !availableBaseUrl) {
            toast.warning(t('testNeedBaseUrlAndKey'));
            return;
        }
        try {
            const data = await fetchModel.mutateAsync(buildModelFetchRequest(targetKey.channel_key));

            const nextFetched = normalizeFetchedModels(data);
            if (nextFetched.length === 0) {
                toast.warning(t('modelRefreshEmpty'));
                return;
            }

            openKeyModelDialog(keyIndex, nextFetched);
        } catch (error) {
            const errorMessage = getErrorMessage(error);
            toast.error(t('modelRefreshFailed'), { description: errorMessage });
        }
    };

    const handleOpenGlobalModelDialog = async () => {
        setKeyModelDialog(null);
        const requestKey = availableRequestKey;
        if (!requestKey || !availableBaseUrl) {
            toast.warning(t('testNeedBaseUrlAndKey'));
            return;
        }
        try {
            const data = await fetchModel.mutateAsync(buildModelFetchRequest(requestKey));
            const nextFetched = normalizeFetchedModels(data);
            if (nextFetched.length === 0) {
                toast.warning(t('modelRefreshEmpty'));
                return;
            }
            openGlobalModelDialog(nextFetched);
        } catch (error) {
            const errorMessage = getErrorMessage(error);
            toast.error(t('modelRefreshFailed'), { description: errorMessage });
        }
    };

    const handleModelDialogOpenChange = (open: boolean) => {
        setShowModelSelectDialog(open);
        if (open) return;
        setModelDialogKeyword('');
        setKeyModelDialog(null);
    };

    const filteredDialogModels = useMemo(() => {
        const models = keyModelDialog?.models ?? [];
        const keyword = deferredModelDialogKeyword.trim().toLowerCase();
        if (!keyword) return models;
        return models.filter((model) => model.toLowerCase().includes(keyword));
    }, [deferredModelDialogKeyword, keyModelDialog?.models]);

    const modelDialogSelectedCount = keyModelDialog?.selected.size ?? 0;

    const handleAddKey = () => {
        onFormDataChange({
            ...formData,
            keys: [...formData.keys, { enabled: true, channel_key: '', source_type: '', remark: '', allowed_models: '' }],
        });
        setKeyFilter('');
        setExpandedKeyItems((current) => [...new Set([...current, `key-${formData.keys.length}`])]);
    };

    const handleToggleExpandedKey = useCallback((itemValue: string) => {
        setExpandedKeyItems((current) => (
            current.includes(itemValue)
                ? current.filter((item) => item !== itemValue)
                : [...current, itemValue]
        ));
    }, []);

    const handleToggleKeyFilterPanel = useCallback(() => {
        if (showKeyFilterPanel || keyFilter.trim().length > 0) {
            setShowKeyFilterPanel(false);
            if (keyFilter.trim().length > 0) {
                setKeyFilter('');
            }
            return;
        }

        setShowKeyFilterPanel(true);
    }, [keyFilter, showKeyFilterPanel]);

    const handleUpdateKey = (idx: number, patch: Partial<ChannelKeyFormItem>) => {
        const next = formData.keys.map((k, i) => (i === idx ? { ...k, ...patch } : k));
        onFormDataChange({ ...formData, keys: next });
    };

    const handleAddModelToKey = (keyIndex: number) => {
        const rawDraft = keyModelDrafts[keyIndex] ?? '';
        const nextModel = rawDraft.trim();
        if (!nextModel) return;

        const current = splitAllowedModels(formData.keys[keyIndex]?.allowed_models);
        const merged = appendUniqueModels(current, [nextModel]);
        handleUpdateKey(keyIndex, { allowed_models: merged.join(',') });
        setKeyModelDrafts((currentDrafts) => ({ ...currentDrafts, [keyIndex]: '' }));
    };

    const handleKeyModelDraftChange = (keyIndex: number, value: string) => {
        setKeyModelDrafts((currentDrafts) => ({ ...currentDrafts, [keyIndex]: value }));
    };

    const handleKeyModelDraftKeyDown = (event: React.KeyboardEvent<HTMLInputElement>, keyIndex: number) => {
        if (event.key !== 'Enter') return;
        event.preventDefault();
        handleAddModelToKey(keyIndex);
    };

    const handleRemoveKey = (idx: number) => {
        const curr = formData.keys ?? [];
        if (curr.length <= 1) return;
        const next = curr.filter((_, i) => i !== idx);
        onFormDataChange({ ...formData, keys: next });
        setKeyModelDrafts({});
        setExpandedKeyItems((current) => current
            .filter((item) => item !== `key-${idx}`)
            .map((item) => {
                const match = /^key-(\d+)$/.exec(item);
                if (!match) return item;
                const itemIndex = Number(match[1]);
                return itemIndex > idx ? `key-${itemIndex - 1}` : item;
            }));
    };

    const handleAddBaseUrl = () => {
        onFormDataChange({
            ...formData,
            base_urls: [...(formData.base_urls ?? []), { url: '', delay: 0 }],
        });
    };

    const handleUpdateBaseUrl = (idx: number, patch: Partial<Channel['base_urls'][number]>) => {
        const next = (formData.base_urls ?? []).map((u, i) => (i === idx ? { ...u, ...patch } : u));
        onFormDataChange({ ...formData, base_urls: next });
    };

    const handleRemoveBaseUrl = (idx: number) => {
        const curr = formData.base_urls ?? [];
        if (curr.length <= 1) return;
        onFormDataChange({ ...formData, base_urls: curr.filter((_, i) => i !== idx) });
    };

    const handleAddHeader = () => {
        onFormDataChange({
            ...formData,
            custom_header: [...(formData.custom_header ?? []), { header_key: '', header_value: '' }],
        });
    };

    const handleUpdateHeader = (idx: number, patch: Partial<Channel['custom_header'][number]>) => {
        const next = (formData.custom_header ?? []).map((h, i) => (i === idx ? { ...h, ...patch } : h));
        onFormDataChange({ ...formData, custom_header: next });
    };

    const handleRemoveHeader = (idx: number) => {
        const curr = formData.custom_header ?? [];
        if (curr.length <= 1) return;
        onFormDataChange({ ...formData, custom_header: curr.filter((_, i) => i !== idx) });
    };

    // All models (auto + custom)
    const allModels = [
        ...autoModels,
        ...customModels,
    ];

    const currentChannelKeyOptions = useMemo(() => {
        return (formData.keys ?? [])
            .filter((key): key is ChannelKeyFormItem & { id: number } => typeof key.id === 'number' && key.id > 0)
            .map((key) => ({
                id: key.id,
                label: getChannelKeyLabel(key, { fallbackLabel: t('keyFallbackLabel') }),
            }));
    }, [formData.keys, t]);

    const currentChannelKeyLabelByID = useMemo(() => {
        return buildChannelKeyLabelMap(formData.keys ?? [], { fallbackLabel: t('keyFallbackLabel') });
    }, [formData.keys, t]);

    const currentRouteTargetOverrides = useMemo(() => routeTargetOverrides.data ?? [], [routeTargetOverrides.data]);

    const formatSourceTypeLabel = useCallback((value?: string | null) => {
        return tRouteTarget(`sourceTypeOptions.${getChannelKeySourceTypeKey(value)}`);
    }, [tRouteTarget]);

    const getKeySummaryText = useCallback((key: ChannelKeyFormItem) => {
        const parts: string[] = [];
        const masked = maskKeyPreview(key.channel_key);
        const sourceTypeLabel = formatSourceTypeLabel(key.source_type);
        const allowedModels = splitAllowedModels(key.allowed_models);

        parts.push(masked || t('keyValueEmpty'));
        if (key.remark?.trim()) {
            parts.push(key.remark.trim());
        }
        if (key.source_type && key.source_type !== 'unknown' && sourceTypeLabel) {
            parts.push(sourceTypeLabel);
        } else {
            parts.push(t('sourceTypePlaceholder'));
        }
        parts.push(
            allowedModels.length > 0
                ? t('keySummaryAllowedModels', { count: allowedModels.length })
                : t('keySummaryAllModels')
        );

        return parts;
    }, [formatSourceTypeLabel, t]);

    const visibleFormKeys = useMemo(() => {
        const keyword = keyFilter.trim().toLowerCase();
        return (formData.keys ?? []).flatMap((key, index) => {
            if (!keyword) {
                return [{ key, index }];
            }

            const searchText = [
                t('keyRowTitle', { index: index + 1 }),
                getChannelKeyLabel(key, { fallbackLabel: t('keyFallbackLabel') }),
                ...getKeySummaryText(key),
                key.channel_key ?? '',
                key.remark ?? '',
                key.allowed_models ?? '',
                key.source_type ?? '',
            ].join(' ').toLowerCase();

            return searchText.includes(keyword) ? [{ key, index }] : [];
        });
    }, [formData.keys, getKeySummaryText, keyFilter, t]);

    const formatBillingModeLabel = useCallback((value?: string | null) => {
        return tRouteTarget(`billingModeOptions.${getBillingModeKey(value)}`);
    }, [tRouteTarget]);

    const formatProbePolicyLabel = useCallback((value?: string | null) => {
        return tRouteTarget(`probePolicyOptions.${getProbePolicyKey(value)}`);
    }, [tRouteTarget]);

    const formatRouteTargetSummary = useCallback((row: RouteTargetOverride) => {
        return t('routeTargetSummary', {
            key: currentChannelKeyLabelByID.get(row.channel_key_id) ?? `${t('keyFallbackLabel')} #${row.channel_key_id}`,
            model: row.model_name,
            billing: formatBillingModeLabel(row.billing_mode),
            probe: formatProbePolicyLabel(row.probe_policy),
            interval: row.probe_interval_seconds,
            concurrency: row.probe_concurrency_limit,
        });
    }, [currentChannelKeyLabelByID, formatBillingModeLabel, formatProbePolicyLabel, t]);

    const handleOpenRouteTargetDialog = () => {
        const defaultKey = currentChannelKeyOptions[0]?.id;
        const defaultModel = allModels[0] ?? '';
        setRouteTargetForm((current) => ({
            ...current,
            channel_key_id: current.channel_key_id || (defaultKey ? String(defaultKey) : ''),
            model_name: current.model_name || defaultModel,
        }));
        setShowRouteTargetDialog(true);
    };

    const handleApplyRouteTargetOverride = async () => {
        if (!channelId) {
            toast.error(t('routeTargetSaveFirst'));
            return;
        }
        if (!routeTargetForm.channel_key_id || !routeTargetForm.model_name.trim()) {
            toast.error(t('routeTargetRequired'));
            return;
        }
        await upsertRouteTargetOverride.mutateAsync({
            channel_id: channelId,
            channel_key_id: Number(routeTargetForm.channel_key_id),
            model_name: routeTargetForm.model_name.trim(),
            billing_mode: routeTargetForm.billing_mode as RouteTargetOverride['billing_mode'],
            probe_policy: routeTargetForm.probe_policy as RouteTargetOverride['probe_policy'],
            probe_interval_seconds: parseInt(routeTargetForm.probe_interval_seconds, 10) || 3600,
            probe_concurrency_limit: parseInt(routeTargetForm.probe_concurrency_limit, 10) || 1,
        });
        toast.success(t('routeTargetSaved'));
    };

    const handleDeleteRouteTargetOverrideRow = async (row: RouteTargetOverride) => {
        await deleteRouteTargetOverride.mutateAsync({
            channel_id: row.channel_id,
            channel_key_id: row.channel_key_id,
            model_name: row.model_name,
        });
        toast.success(t('routeTargetDeleted'));
    };

    const handleEditRouteTargetOverrideRow = (row: RouteTargetOverride) => {
        setRouteTargetForm({
            channel_key_id: String(row.channel_key_id),
            model_name: row.model_name,
            billing_mode: row.billing_mode,
            probe_policy: row.probe_policy,
            probe_interval_seconds: String(row.probe_interval_seconds ?? 3600),
            probe_concurrency_limit: String(row.probe_concurrency_limit ?? 1),
        });
        setShowRouteTargetDialog(true);
    };

    const handleTestModels = async (models: string[]) => {
        if (models.length === 0 || isTesting) return;
        const hasBaseUrl = formData.base_urls?.some((u) => u.url.trim());
        const hasKey = formData.keys?.some((k) => k.channel_key.trim());
        if (!hasBaseUrl || !hasKey) {
            toast.warning(t('testNeedBaseUrlAndKey'));
            return;
        }
        setIsTesting(true);
        try {
            const results = await testByConfig.mutateAsync({
                type: formData.type,
                enabled: formData.enabled,
                base_urls: formData.base_urls.filter((u) => u.url.trim()),
                keys: formData.keys.filter((k) => k.channel_key.trim()).map((k) => ({
                    enabled: k.enabled,
                    channel_key: k.channel_key.trim(),
                    source_type: (k.source_type ?? '').trim(),
                    allowed_models: (k.allowed_models ?? '').trim(),
                })),
                proxy: formData.proxy,
                channel_proxy: formData.channel_proxy?.trim() || null,
                custom_header: formData.custom_header?.filter((h) => h.header_key.trim()) || [],
                key_management_mode: normalizeKeyManagementMode(formData.key_management_mode),
                key_routing_policy: normalizeKeyRoutingPolicy(formData.key_routing_policy),
                models,
            });
            const map = new Map<string, TestModelResult>();
            for (const r of results) map.set(r.model, r);
            setTestResults(map);
        } catch {
            toast.error(t('testFailed'));
        } finally {
            setIsTesting(false);
        }
    };

    const handleTestFirst = () => {
        if (allModels.length > 0) handleTestModels([allModels[0]]);
    };

    const handleTestAll = () => {
        handleTestModels(allModels);
    };

    // Provider preset quick-select
    const handleProviderPreset = (providerName: string) => {
        if (!providers) return;
        const provider = providers.find((p) => p.name === providerName);
        if (!provider) return;
        onFormDataChange({
            ...formData,
            type: provider.channel_type as ChannelType,
            base_urls: [{ url: provider.base_url, delay: 0 }],
        });
    };

    const namePlaceholder = (() => {
        if (!providers) return t('namePlaceholder');
        const currentUrl = formData.base_urls?.[0]?.url?.trim();
        const p = providers.find((p) => currentUrl && p.base_url === currentUrl);
        return p ? `${t('namePlaceholderPrefix')}${p.name}` : t('namePlaceholder');
    })();

    const showManualConnectionFields = formData.type !== ChannelType.GithubCopilot && formData.type !== ChannelType.Antigravity;
    const currentKeyManagementMode = normalizeKeyManagementMode(formData.key_management_mode);
    const isClassifiedMode = currentKeyManagementMode === 'classified';
    const keyModeSummaryText = isClassifiedMode ? t('allowedModelsScopedHint') : t('allowedModelsPooledSummary');
    const createFlowSteps = [
        { key: 'basic', index: 1, title: flowCopy.basicSectionTitle, description: flowCopy.basicSectionDesc },
        { key: 'keys', index: 2, title: flowCopy.keySectionTitle, description: flowCopy.keySectionDesc },
        { key: 'models', index: 3, title: flowCopy.modelSectionTitle, description: flowCopy.modelSectionDesc },
        { key: 'advanced', index: 4, title: flowCopy.advancedSectionTitle, description: flowCopy.advancedSectionDesc },
    ];

    return (
        <>
        <form data-testid={`${idPrefix}-form`} onSubmit={onSubmit} className="space-y-4 px-1">
            <div data-testid={`${idPrefix}-flow-card`} className="rounded-2xl border border-border/70 bg-muted/15 p-3">
                <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                    <span>{flowCopy.createFlowTitle}</span>
                </div>
                <div className="mt-3 grid gap-2 md:grid-cols-2 xl:grid-cols-4">
                    {createFlowSteps.map((step) => (
                        <div key={step.key} className="rounded-xl border border-border/60 bg-background/85 px-3 py-2">
                            <div className="flex items-center gap-2">
                                <Badge variant="outline" className="h-6 rounded-full px-2 text-[11px] font-medium">
                                    {flowCopy.stepLabel(step.index)}
                                </Badge>
                                <span className="text-sm font-medium text-card-foreground">{step.title}</span>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
            {/* Provider 快速预设选择 */}
            {providers && providers.length > 0 && (
                <div className="space-y-2">
                    <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                        <span>{t('providerPreset')}</span>
                    </div>
                    <Select onValueChange={handleProviderPreset}>
                        <SelectTrigger className="rounded-xl w-full border border-border px-4 py-2 text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
                            <SelectValue placeholder={t('providerPresetPlaceholder')} />
                        </SelectTrigger>
                        <SelectContent className="rounded-xl">
                            {providers.map((p) => (
                                <SelectItem key={`${p.name}-${p.channel_type}`} className="rounded-xl" value={p.name}>
                                    {p.name}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>
            )}
            <div data-testid={`${idPrefix}-basic-section`} className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                    <span>{flowCopy.basicSectionTitle}</span>
                </div>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                    <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                        <label htmlFor={`${idPrefix}-name`}>{t('name')}</label>
                    </div>
                    <Input
                        className='rounded-xl'
                        id={`${idPrefix}-name`}
                        type="text"
                        value={formData.name}
                        onChange={(event) => onFormDataChange({ ...formData, name: event.target.value })}
                        placeholder={namePlaceholder}
                        required
                    />
                </div>

                <div className="space-y-2">
                    <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                        <label htmlFor={`${idPrefix}-type`}>{t('type')}</label>
                    </div>
                    <Select
                        value={String(formData.type)}
                        onValueChange={(value) => onFormDataChange({ ...formData, type: Number(value) as ChannelType })}
                    >
                        <SelectTrigger id={`${idPrefix}-type`} className="rounded-xl w-full border border-border px-4 py-2 text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent className='rounded-xl'>
                            <SelectItem className='rounded-xl' value={String(ChannelType.OpenAIChat)}>{t('typeOpenAIChat')}</SelectItem>
                            <SelectItem className='rounded-xl' value={String(ChannelType.OpenAIResponse)}>{t('typeOpenAIResponse')}</SelectItem>
                            <SelectItem className='rounded-xl' value={String(ChannelType.Anthropic)}>{t('typeAnthropic')}</SelectItem>
                            <SelectItem className='rounded-xl' value={String(ChannelType.Gemini)}>{t('typeGemini')}</SelectItem>
                            <SelectItem className='rounded-xl' value={String(ChannelType.Volcengine)}>{t('typeVolcengine')}</SelectItem>
                            <SelectItem className='rounded-xl' value={String(ChannelType.OpenAIEmbedding)}>{t('typeOpenAIEmbedding')}</SelectItem>
                            <SelectItem className='rounded-xl' value={String(ChannelType.GithubCopilot)}>{t('typeGithubCopilot')}</SelectItem>
                            <SelectItem className='rounded-xl' value={String(ChannelType.Antigravity)}>{t('typeAntigravity')}</SelectItem>
                            <SelectItem className='rounded-xl' value={String(ChannelType.Zen)}>{t('typeZen')}</SelectItem>
                        </SelectContent>
                    </Select>
                </div>

                <div className="space-y-2">
                    <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                        <label htmlFor={`${idPrefix}-key-management-mode`}>{t('keyManagementMode')}</label>
                        <HelpHint className="size-3.5">{t('keyManagementModeHint')}</HelpHint>
                    </div>
                    <Select
                        value={normalizeKeyManagementMode(formData.key_management_mode)}
                        onValueChange={(value) => onFormDataChange({ ...formData, key_management_mode: normalizeKeyManagementMode(value) })}
                    >
                        <SelectTrigger id={`${idPrefix}-key-management-mode`} className="rounded-xl w-full border border-border px-4 py-2 text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent className='rounded-xl'>
                            <SelectItem className='rounded-xl' value="pooled">{tDetail('mode.pooled')}</SelectItem>
                            <SelectItem className='rounded-xl' value="classified">{tDetail('mode.classified')}</SelectItem>
                        </SelectContent>
                    </Select>
                </div>

                <div className="space-y-2">
                    <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                        <label htmlFor={`${idPrefix}-key-routing-policy`}>{t('keyRoutingPolicy')}</label>
                        <HelpHint className="size-3.5">{t('keyRoutingPolicyHint')}</HelpHint>
                    </div>
                    <Select
                        value={normalizeKeyRoutingPolicy(formData.key_routing_policy)}
                        onValueChange={(value) => onFormDataChange({ ...formData, key_routing_policy: normalizeKeyRoutingPolicy(value) })}
                    >
                        <SelectTrigger id={`${idPrefix}-key-routing-policy`} className="rounded-xl w-full border border-border px-4 py-2 text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent className='rounded-xl'>
                            <SelectItem className='rounded-xl' value="round_robin">{tDetail('policy.round_robin')}</SelectItem>
                            <SelectItem className='rounded-xl' value="fill_priority">{tDetail('policy.fill_priority')}</SelectItem>
                            <SelectItem className='rounded-xl' value="priority_order">{tDetail('policy.priority_order')}</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
            </div>

            <div data-testid={`${idPrefix}-key-section-heading`} className="space-y-1">
                <div className="flex flex-wrap items-center justify-between gap-2">
                    <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                        <span>{flowCopy.keySectionTitle}</span>
                        <HelpHint className="size-3.5">{keyModeSummaryText}</HelpHint>
                    </div>
                    <Badge variant="secondary" className="rounded-full px-2 py-0 text-[11px]">
                        {isClassifiedMode ? tDetail('mode.classified') : tDetail('mode.pooled')}
                    </Badge>
                </div>
            </div>

            {/* GitHub Copilot Device Flow Panel */}
            {formData.type === ChannelType.GithubCopilot && (
                <div className="space-y-3 rounded-xl border border-blue-500/30 bg-blue-500/5 p-4">
                    <div className="flex items-center gap-2 text-sm font-medium text-blue-700 dark:text-blue-300">
                        <Info className="h-4 w-4 shrink-0" />
                        <span>{t('copilotDeviceFlow')}</span>
                    </div>

                    {copilotStatus === 'idle' && (
                        <Button
                            type="button"
                            onClick={handleCopilotStartAuth}
                            className="w-full rounded-xl h-11 gap-2 bg-blue-600 hover:bg-blue-700 text-white"
                        >
                            {t('copilotStartAuth')}
                        </Button>
                    )}

                    {copilotStatus === 'loading' && (
                        <div className="flex justify-center py-4">
                            <Loader2 className="h-6 w-6 animate-spin text-blue-500" />
                        </div>
                    )}

                    {copilotStatus === 'waiting' && (
                        <div className="space-y-3">
                            <p className="text-xs text-muted-foreground">{t('copilotUserCodeHint')}</p>
                            <div className="flex items-center gap-2">
                                <div className="flex-1 rounded-xl border-2 border-green-500/50 bg-green-500/10 px-4 py-3 text-center">
                                    <span className="font-mono text-2xl font-bold tracking-widest text-green-700 dark:text-green-400">
                                        {copilotUserCode}
                                    </span>
                                </div>
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="sm"
                                    onClick={() => {
                                        navigator.clipboard.writeText(copilotUserCode);
                                        toast.success(t('copilotCodeCopied'));
                                    }}
                                    className="rounded-xl h-11 w-11 p-0"
                                    title={t('copilotCodeCopied')}
                                >
                                    <Copy className="h-4 w-4" />
                                </Button>
                            </div>
                            <Button
                                type="button"
                                variant="outline"
                                className="w-full rounded-xl h-11 gap-2"
								onClick={() => window.open(copilotVerificationUri, '_blank', 'noopener,noreferrer')}
                            >
                                <ExternalLink className="h-4 w-4" />
                                {t('copilotOpenGitHub')}
                            </Button>
                            <div className="flex items-center gap-2 text-xs text-muted-foreground pt-1">
                                <Loader2 className="h-3.5 w-3.5 animate-spin" />
                                <span>{t('copilotWaiting')}</span>
                            </div>
                        </div>
                    )}

                    {copilotStatus === 'authorized' && (
                        <div className="flex items-center gap-2 text-sm text-green-600 dark:text-green-400">
                            <CheckCircle2 className="h-4 w-4" />
                            <span>{t('copilotSuccess')}</span>
                        </div>
                    )}

                    {(copilotStatus === 'expired' || copilotStatus === 'denied' || copilotStatus === 'error') && (
                        <div className="space-y-3">
                            <div className="flex items-center gap-2 text-sm text-destructive">
                                <XCircle className="h-4 w-4" />
                                <span>
                                    {copilotStatus === 'expired'
                                        ? t('copilotExpired')
                                        : copilotStatus === 'denied'
                                          ? t('copilotDenied')
                                          : t('copilotError')}
                                </span>
                            </div>
                            <Button
                                type="button"
                                variant="outline"
                                className="w-full rounded-xl h-11"
                                onClick={handleCopilotStartAuth}
                            >
                                {t('copilotRetry')}
                            </Button>
                        </div>
                    )}
                </div>
            )}

            {/* Antigravity OAuth Web Flow Panel */}
            {formData.type === ChannelType.Antigravity && (
                <div className="space-y-3 rounded-xl border border-purple-500/30 bg-purple-500/5 p-4">
                    <div className="flex items-center gap-2 text-sm font-medium text-purple-700 dark:text-purple-300">
                        <Info className="h-4 w-4 shrink-0" />
                        <span>{t('antigravityOAuthTitle')}</span>
                    </div>

                    {antigravityStatus === 'idle' && (
                        <>
                        <p className="text-xs text-muted-foreground">{t('antigravityConfigHint')}</p>
                        <Button
                            type="button"
                            onClick={handleAntigravityStartAuth}
                            className="w-full rounded-xl h-11 gap-2 bg-purple-600 hover:bg-purple-700 text-white"
                        >
                            {t('antigravityStartAuth')}
                        </Button>
                        </>
                    )}

                    {antigravityStatus === 'loading' && (
                        <div className="flex justify-center py-4">
                            <Loader2 className="h-6 w-6 animate-spin text-purple-500" />
                        </div>
                    )}

                    {antigravityStatus === 'waiting' && (
                        <div className="space-y-2 text-xs text-muted-foreground">
                            <div className="flex items-center gap-2">
                                <Loader2 className="h-3.5 w-3.5 animate-spin" />
                                <span>{t('antigravityWaiting')}</span>
                            </div>
                            <Button
                                type="button"
                                variant="outline"
                                className="w-full rounded-xl h-10"
                                onClick={handleAntigravityStartAuth}
                            >
                                {t('antigravityOpenAgain')}
                            </Button>
                        </div>
                    )}

                    {antigravityStatus === 'authorized' && (
                        <div className="flex items-center gap-2 text-sm text-green-600 dark:text-green-400">
                            <CheckCircle2 className="h-4 w-4" />
                            <span>{t('antigravitySuccess')}</span>
                        </div>
                    )}

                    {antigravityStatus === 'error' && (
                        <div className="space-y-3">
                            <div className="flex items-center gap-2 text-sm text-destructive">
                                <XCircle className="h-4 w-4" />
                                <span>{antigravityError || t('antigravityAuthFailed')}</span>
                            </div>
                            <Button
                                type="button"
                                variant="outline"
                                className="w-full rounded-xl h-11"
                                onClick={handleAntigravityStartAuth}
                            >
                                {t('antigravityRetry')}
                            </Button>
                        </div>
                    )}
                </div>
            )}

            {showManualConnectionFields && (
            <div data-testid={`${idPrefix}-key-section`} className="space-y-2">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-1">
                        <label className="text-sm font-medium text-card-foreground">
                            {t('baseUrls')} {formData.base_urls.length > 0 ? `(${formData.base_urls.length})` : ''}
                        </label>
                    </div>
                    <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={handleAddBaseUrl}
                        className="h-6 px-2 text-xs text-muted-foreground/70 hover:text-muted-foreground hover:bg-transparent"
                    >
                        <Plus className="h-3 w-3 mr-1" />
                        {t('add')}
                    </Button>
                </div>
                <div className="space-y-2">
                    {(formData.base_urls ?? []).map((u, idx) => (
                        <div key={`baseurl-${idx}`} className="flex items-center gap-2 max-[420px]:flex-col max-[420px]:items-stretch">
                            <Input
                                id={`${idPrefix}-base-${idx}`}
                                type="url"
                                value={u.url}
                                onChange={(e) => handleUpdateBaseUrl(idx, { url: e.target.value })}
                                placeholder={t('baseUrlUrl')}
                                required={idx === 0}
                                className="rounded-xl flex-1"
                            />
                            <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                onClick={() => handleRemoveBaseUrl(idx)}
                                disabled={(formData.base_urls ?? []).length <= 1}
                                className="h-8 w-8 p-0 rounded-xl text-muted-foreground hover:text-destructive disabled:opacity-40 hover:bg-transparent"
                                title={t('remove')}
                            >
                                <X className="h-4 w-4" />
                            </Button>
                        </div>
                    ))}
                </div>
            </div>
            )}

            {showManualConnectionFields && (
            <div data-testid={`${idPrefix}-key-section`} className="space-y-2">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                        <label>
                            {t('multiKeyTitle')} {formData.keys.length > 0 ? `(${formData.keys.length})` : ''}
                        </label>
                    </div>
                    <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={handleAddKey}
                        className="h-6 px-2 text-xs text-muted-foreground/70 hover:text-muted-foreground hover:bg-transparent"
                    >
                        <Plus className="h-3 w-3 mr-1" />
                        {t('add')}
                    </Button>
                </div>

                <div className="rounded-[1.15rem] border border-border/60 bg-background/75 px-3 py-2.5 sm:px-3.5">
                    <div className="flex flex-wrap items-center justify-between gap-2.5">
                        <div className="min-w-0 text-sm font-semibold leading-6 text-card-foreground">
                            {t('multiKeySummaryLine', {
                                total: keySetupSummary.total,
                                ready: keySetupSummary.ready,
                                enabled: keySetupSummary.enabled,
                                pending: keySetupSummary.pending,
                            })}
                        </div>
                        <div className="flex flex-wrap items-center gap-1.5 text-[11px] text-muted-foreground">
                            <Badge variant="outline" className="rounded-full px-2 py-0">{t('keySummaryTotalLabel')} {keySetupSummary.total}</Badge>
                            <Badge variant="outline" className="rounded-full px-2 py-0">{t('keySummaryReadyLabel')} {keySetupSummary.ready}</Badge>
                            <Badge variant="outline" className="rounded-full px-2 py-0">{t('keySummaryEnabledLabel')} {keySetupSummary.enabled}</Badge>
                        </div>
                    </div>
                </div>

                {showKeyFilter && (
                    <div className="space-y-2 rounded-[1.2rem] border border-border/60 bg-muted/15 p-3">
                        <div className="flex flex-wrap items-center justify-between gap-2">
                            <div className="text-xs font-medium text-card-foreground">{t('keyFilterLabel')}</div>
                            <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                onClick={handleToggleKeyFilterPanel}
                                className="h-8 rounded-xl px-3 text-[11px] text-muted-foreground hover:bg-background/80 hover:text-card-foreground"
                            >
                                <Search className="mr-1 h-3.5 w-3.5" />
                                {keyFilterExpanded ? t('collapseInline') : t('expandInline')}
                            </Button>
                        </div>
                        {keyFilterExpanded && (
                            <>
                                <p className="text-[11px] leading-5 text-muted-foreground">{t('keyFilterHint')}</p>
                                <div className="relative">
                                    <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                                    <Input
                                        data-testid={`${idPrefix}-key-filter`}
                                        value={keyFilter}
                                        onChange={(event) => setKeyFilter(event.target.value)}
                                        placeholder={t('keyFilterPlaceholder')}
                                        className="rounded-2xl bg-background pl-9"
                                    />
                                </div>
                                {keyFilter.trim() && (
                                    <div className="px-1 text-xs text-muted-foreground">
                                        {tDetail('keySummaryLine', {
                                            total: keySetupSummary.total,
                                            enabled: keySetupSummary.enabled,
                                            matched: visibleFormKeys.length,
                                        })}
                                    </div>
                                )}
                            </>
                        )}
                    </div>
                )}

                {visibleFormKeys.length > 0 ? (
                isClassifiedMode ? (
                <Accordion
                    type="multiple"
                    value={expandedKeyItems}
                    onValueChange={setExpandedKeyItems}
                    className="space-y-2"
                >
                    {visibleFormKeys.map(({ key: k, index: idx }) => {
                        const keyAccordionValue = `key-${idx}`;
                        const isExpanded = expandedKeyItems.includes(keyAccordionValue);
                        const allowedModels = splitAllowedModels(k.allowed_models);
                        const keyModelDraft = keyModelDrafts[idx] ?? '';
                        const canAddKeyModel = Boolean(keyModelDraft.trim()) && !allowedModels.includes(keyModelDraft.trim());
                        const hasRealKey = Boolean(k.channel_key.trim());
                        const maskedKey = maskKeyPreview(k.channel_key);
                        const sourceTypeSummary = k.source_type && k.source_type !== 'unknown'
                            ? formatSourceTypeLabel(k.source_type)
                            : t('sourceTypePlaceholder');
                        const scopeSummary = allowedModels.length > 0
                            ? t('keySummaryAllowedModels', { count: allowedModels.length })
                            : t('keySummaryAllModels');
                        const remarkSummary = k.remark?.trim() || t('keySummaryNoRemark');
                        return (
                            <AccordionItem
                                key={k.id ?? `new-${idx}`}
                                value={keyAccordionValue}
                                data-testid={`${idPrefix}-key-item-${idx}`}
                                ref={(node) => {
                                    if (typeof k.id === 'number') keyRowRefs.current[String(k.id)] = node;
                                }}
                                className="overflow-hidden rounded-2xl border border-border/70 bg-muted/20 transition-colors data-[state=open]:border-primary/20 data-[state=open]:bg-primary/[0.04]"
                            >
                                <AccordionTrigger
                                    data-testid={`${idPrefix}-key-trigger-${idx}`}
                                    className="rounded-2xl px-3.5 py-2.5 hover:no-underline hover:bg-background/50 focus-visible:ring-primary/20 sm:px-4"
                                    showIndicator={false}
                                    addon={(
                                        <div className="flex items-center gap-2 pt-1 sm:pt-0">
                                            <div
                                                className="flex items-center gap-2 rounded-full border border-border/60 bg-background/90 px-2.5 py-1 text-[11px] text-muted-foreground shadow-sm"
                                                onClick={(event) => event.stopPropagation()}
                                            >
                                                <span>{t('enabled')}</span>
                                                <Switch
                                                    checked={k.enabled}
                                                    onCheckedChange={(checked) => handleUpdateKey(idx, { enabled: checked })}
                                                    onClick={(event) => event.stopPropagation()}
                                                />
                                            </div>
                                            <span className="flex h-8 w-8 items-center justify-center rounded-xl border border-border/60 bg-background/80 text-muted-foreground transition-transform duration-200 data-[expanded=true]:rotate-180" data-expanded={isExpanded}>
                                                <ChevronDown className="h-4 w-4" />
                                            </span>
                                            <Button
                                                type="button"
                                                variant="ghost"
                                                size="sm"
                                                onClick={(event) => {
                                                    event.stopPropagation();
                                                    handleRemoveKey(idx);
                                                }}
                                                disabled={(formData.keys ?? []).length <= 1}
                                                className="h-8 w-8 rounded-xl p-0 text-muted-foreground hover:bg-transparent hover:text-destructive disabled:opacity-40"
                                                title={t('remove')}
                                            >
                                                <X className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    )}
                                    addonClassName="self-start sm:self-center"
                                >
                                    <div className="flex w-full flex-col gap-2 text-left">
                                        <div className="flex flex-wrap items-start gap-2.5">
                                            <div className="min-w-0 flex-1">
                                                <div className="flex flex-wrap items-center gap-2">
                                                    <span className="text-sm font-medium text-card-foreground">{t('keyRowTitle', { index: idx + 1 })}</span>
                                                    <Badge variant="outline" className="rounded-full px-2 py-0 text-[11px]">
                                                        {k.enabled ? t('keySummaryEnabled') : t('keySummaryDisabled')}
                                                    </Badge>
                                                    <Badge variant="secondary" className="rounded-full px-2 py-0 text-[11px]">
                                                        {hasRealKey ? t('keySetupReady') : t('keySetupPending')}
                                                    </Badge>

                                                </div>
                                                {!hasRealKey ? (
                                                    <p data-slot="channel-key-summary-banner" className="mt-1.5 text-xs leading-5 text-muted-foreground">
                                                        {t('keyCollapsedPendingHint')}
                                                    </p>
                                                ) : null}
                                            </div>
                                        </div>

                                        <div data-slot="channel-key-summary-grid" className="flex flex-wrap items-center gap-1.5">
                                            <span className="inline-flex min-w-0 max-w-full items-center rounded-full border border-border/60 bg-background/90 px-2.5 py-1 text-[11px] text-card-foreground">
                                                <span className="mr-1 shrink-0 text-muted-foreground">{t('keyValueLabel')}</span>
                                                <span className="truncate font-mono" title={maskedKey || t('keyValueEmpty')}>
                                                    {maskedKey || t('keyValueEmpty')}
                                                </span>
                                            </span>
                                            <span className="inline-flex min-w-0 max-w-full items-center rounded-full border border-border/60 bg-background/90 px-2.5 py-1 text-[11px] text-card-foreground">
                                                <span className="mr-1 shrink-0 text-muted-foreground">{t('allowedModelsLabel')}</span>
                                                <span className="truncate" title={scopeSummary}>{scopeSummary}</span>
                                            </span>
                                            {k.source_type && k.source_type !== 'unknown' && (
                                                <span className="inline-flex min-w-0 max-w-full items-center rounded-full border border-border/60 bg-background/90 px-2.5 py-1 text-[11px] text-card-foreground">
                                                    <span className="mr-1 shrink-0 text-muted-foreground">{t('sourceTypeLabel')}</span>
                                                    <span className="truncate" title={sourceTypeSummary}>{sourceTypeSummary}</span>
                                                </span>
                                            )}
                                        </div>
                                    </div>
                                </AccordionTrigger>

                                <AccordionContent className="border-t border-border/60 px-4 pb-4 pt-4">
                                    <div className="grid gap-4 xl:grid-cols-[minmax(0,1.24fr)_minmax(300px,0.76fr)] 2xl:grid-cols-[minmax(0,1.32fr)_minmax(320px,0.72fr)]">
                                        <div className="space-y-3 xl:order-1">
                                            <div data-testid={`${idPrefix}-key-primary-${idx}`} className="space-y-3 rounded-2xl border border-primary/20 bg-primary/5 p-4">
                                                <div className="flex flex-wrap items-center justify-between gap-3">
                                                    <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                                                        <label htmlFor={`${idPrefix}-key-value-${idx}`}>{t('keyPrimarySectionTitle')}</label>
                                                        <HelpHint className="size-3.5">{t('keyValueHint')}</HelpHint>
                                                    </div>
                                                    <Badge variant="secondary" className="rounded-full px-2 py-0.5 text-[11px]">
                                                        {hasRealKey ? t('keySetupReady') : t('keySetupPending')}
                                                    </Badge>
                                                </div>
                                                <p className="text-xs leading-5 text-muted-foreground">{t('keyValueLeadHint')}</p>
                                                <Input
                                                    id={`${idPrefix}-key-value-${idx}`}
                                                    type="text"
                                                    value={k.channel_key}
                                                    onChange={(e) => handleUpdateKey(idx, { channel_key: e.target.value })}
                                                    placeholder={t('keyValuePlaceholder')}
                                                    required={idx === 0}
                                                    className="h-11 rounded-xl border-border/70 bg-background"
                                                />
                                                <div className={`text-xs ${hasRealKey ? 'text-emerald-700 dark:text-emerald-200' : 'text-amber-700 dark:text-amber-200'}`}>
                                                    {hasRealKey ? t('keyPrimaryStatusReady') : t('keyPrimaryStatusPending')}
                                                </div>
                                                <div className="flex flex-wrap gap-2 text-[11px] text-muted-foreground">
                                                    <span className="rounded-full border border-border/60 bg-background/90 px-2.5 py-1 font-mono">
                                                        {maskedKey || t('keyValueEmpty')}
                                                    </span>
                                                </div>
                                            </div>

                                            <div className="space-y-3 rounded-2xl border border-border/60 bg-background/70 p-4">
                                                <div className="flex flex-wrap items-start justify-between gap-3">
                                                    <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                                                        <span>{t('keyAdvancedSectionTitle')}</span>
                                                        <HelpHint className="size-3.5">{t('keyAdvancedHint')}</HelpHint>
                                                    </div>
                                                    <Badge variant="outline" className="rounded-full px-2 py-0 text-[11px]">
                                                        {t('keyAdvancedOptionalBadge')}
                                                    </Badge>
                                                </div>

                                                <div className="grid gap-3 md:grid-cols-[minmax(156px,0.82fr)_minmax(0,1fr)]">
                                                    <div className="space-y-2">
                                                        <div className="text-sm font-medium text-card-foreground">{t('sourceTypeLabel')}</div>
                                                        <Select
                                                            value={k.source_type && CHANNEL_KEY_SOURCE_TYPES.includes(k.source_type as ChannelKeySourceType)
                                                                ? k.source_type
                                                                : 'unknown'}
                                                            onValueChange={(value) => handleUpdateKey(idx, { source_type: value as ChannelKeySourceType })}
                                                        >
                                                            <SelectTrigger className="rounded-xl w-full bg-background">
                                                                <SelectValue placeholder={t('sourceTypePlaceholder')} />
                                                            </SelectTrigger>
                                                            <SelectContent className='rounded-xl'>
                                                                {CHANNEL_KEY_SOURCE_TYPES.map((sourceType) => (
                                                                    <SelectItem key={sourceType} className='rounded-xl' value={sourceType}>
                                                                        {formatSourceTypeLabel(sourceType)}
                                                                    </SelectItem>
                                                                ))}
                                                            </SelectContent>
                                                        </Select>
                                                    </div>

                                                    <div className="space-y-2">
                                                        <div className="text-sm font-medium text-card-foreground">{t('remarkLabel')}</div>
                                                        <Input
                                                            type="text"
                                                            value={k.remark ?? ''}
                                                            onChange={(e) => handleUpdateKey(idx, { remark: e.target.value })}
                                                            placeholder={t('remarkPlaceholder')}
                                                            className="rounded-xl bg-background"
                                                        />
                                                    </div>
                                                </div>

                                                <div className="flex flex-wrap gap-2 text-[11px] text-muted-foreground">
                                                    {k.source_type && k.source_type !== 'unknown' ? (
                                                        <span className="rounded-full border border-border/60 bg-background/90 px-2.5 py-1">
                                                            {sourceTypeSummary}
                                                        </span>
                                                    ) : null}
                                                    {k.remark?.trim() ? (
                                                        <span className="rounded-full border border-border/60 bg-background/90 px-2.5 py-1">
                                                            {k.remark.trim()}
                                                        </span>
                                                    ) : null}
                                                </div>
                                            </div>
                                        </div>

                                        <div className="space-y-3 rounded-2xl border border-border/60 bg-background/70 p-4 md:p-5 xl:order-2">
                                            <div className="flex flex-wrap items-start justify-between gap-3">
                                                <div className="flex items-center gap-2">
                                                    <div className="text-sm font-medium text-card-foreground">{t('allowedModelsLabel')}</div>
                                                    <HelpHint className="size-3.5">{t('allowedModelsPerKeyLead')}</HelpHint>
                                                </div>
                                                <Badge variant="secondary" className="rounded-full px-2 py-0 text-[11px]">
                                                    {scopeSummary}
                                                </Badge>
                                            </div>

                                            <div className="grid gap-3">
                                                <Button
                                                    type="button"
                                                    variant="outline"
                                                    data-testid={`${idPrefix}-key-fetch-models-${idx}`}
                                                    disabled={fetchModel.isPending || !hasRealKey || !availableBaseUrl}
                                                    onClick={() => handleFetchModelsForKey(idx)}
                                                    className="h-11 w-full rounded-xl px-4 text-sm gap-2"
                                                >
                                                    {fetchModel.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Search className="h-4 w-4" />}
                                                    {t('selectModels')}
                                                </Button>
                                                <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center">
                                                    <Input
                                                        data-testid={`${idPrefix}-key-model-draft-${idx}`}
                                                        value={keyModelDraft}
                                                        onChange={(event) => handleKeyModelDraftChange(idx, event.target.value)}
                                                        onKeyDown={(event) => handleKeyModelDraftKeyDown(event, idx)}
                                                        placeholder={t('modelCustomPlaceholder')}
                                                        className="h-11 rounded-xl bg-background"
                                                    />
                                                    <Button
                                                        type="button"
                                                        variant="outline"
                                                        data-testid={`${idPrefix}-key-model-add-${idx}`}
                                                        disabled={!canAddKeyModel}
                                                        onClick={() => handleAddModelToKey(idx)}
                                                        className="h-11 rounded-xl px-4 text-sm gap-2"
                                                    >
                                                        <Plus className="h-4 w-4" />
                                                        {t('modelAdd')}
                                                    </Button>
                                                </div>
                                            </div>

                                            <div className="rounded-2xl border border-border/60 bg-background/85 p-3">
                                                <div className="flex min-h-40 flex-wrap content-start gap-1.5 rounded-xl border border-border/60 bg-background px-3 py-3">
                                                    {allowedModels.length > 0 ? (
                                                        allowedModels.map((item) => (
                                                            <Badge key={`${idx}-${item}`} className="bg-primary hover:bg-primary/90 text-primary-foreground">
                                                                {item}
                                                                <button
                                                                    type="button"
                                                                    onClick={() => handleUpdateKey(idx, { allowed_models: allowedModels.filter((model) => model !== item).join(',') })}
                                                                    className="ml-1 rounded-sm opacity-70 hover:opacity-100 focus:outline-none focus:ring-1 focus:ring-ring"
                                                                >
                                                                    <X className="h-3 w-3" />
                                                                </button>
                                                            </Badge>
                                                        ))
                                                    ) : (
                                                        <div className="flex min-h-24 items-center text-xs text-muted-foreground">{t('allowedModelsEmptyState')}</div>
                                                    )}
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </AccordionContent>
                            </AccordionItem>
                        );
                    })}
                </Accordion>
                ) : (
                <div className="space-y-2">
                    {visibleFormKeys.map(({ key: k, index: idx }) => {
                        const hasRealKey = Boolean(k.channel_key.trim());
                        const maskedKey = maskKeyPreview(k.channel_key);
                        const sourceTypeSummary = k.source_type && k.source_type !== 'unknown'
                            ? formatSourceTypeLabel(k.source_type)
                            : t('sourceTypePlaceholder');
                        const remarkSummary = k.remark?.trim() || t('keySummaryNoRemark');
                        return (
                            <div
                                key={k.id ?? `pooled-${idx}`}
                                data-testid={`${idPrefix}-pooled-key-item-${idx}`}
                                ref={(node) => {
                                    if (typeof k.id === 'number') keyRowRefs.current[String(k.id)] = node;
                                }}
                                className="rounded-2xl border border-border/70 bg-background/80 px-4 py-4"
                            >
                                <div className="flex flex-col gap-3 lg:flex-row lg:items-center">
                                    <div className="min-w-0 flex-1 space-y-3">
                                        <div className="flex flex-wrap items-center gap-2">
                                            <span className="text-sm font-medium text-card-foreground">{t('keyRowTitle', { index: idx + 1 })}</span>
                                            <Badge variant="outline" className="rounded-full px-2 py-0 text-[11px]">
                                                {k.enabled ? t('keySummaryEnabled') : t('keySummaryDisabled')}
                                            </Badge>
                                            <Badge variant="secondary" className="rounded-full px-2 py-0 text-[11px]">
                                                {hasRealKey ? t('keySetupReady') : t('keySetupPending')}
                                            </Badge>
                                        </div>
                                        <p className="text-xs leading-5 text-muted-foreground">
                                            {hasRealKey ? t('keyCollapsedReadyHint') : t('keyCollapsedPendingHint')}
                                        </p>

                                        <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto_auto] md:items-center">
                                            <Input
                                                id={`${idPrefix}-pooled-key-value-${idx}`}
                                                type="text"
                                                value={k.channel_key}
                                                onChange={(e) => handleUpdateKey(idx, { channel_key: e.target.value })}
                                                placeholder={t('keyValuePlaceholder')}
                                                required={idx === 0}
                                                className="h-11 rounded-xl border-border/70 bg-background"
                                            />
                                            <Select
                                                value={k.source_type && CHANNEL_KEY_SOURCE_TYPES.includes(k.source_type as ChannelKeySourceType)
                                                    ? k.source_type
                                                    : 'unknown'}
                                                onValueChange={(value) => handleUpdateKey(idx, { source_type: value as ChannelKeySourceType })}
                                            >
                                                <SelectTrigger className="h-11 min-w-40 rounded-xl bg-background md:w-44">
                                                    <SelectValue placeholder={t('sourceTypePlaceholder')} />
                                                </SelectTrigger>
                                                <SelectContent className='rounded-xl'>
                                                    {CHANNEL_KEY_SOURCE_TYPES.map((sourceType) => (
                                                        <SelectItem key={sourceType} className='rounded-xl' value={sourceType}>
                                                            {formatSourceTypeLabel(sourceType)}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                            <div className="flex items-center justify-between gap-2 rounded-xl border border-border/60 bg-muted/20 px-3 py-2">
                                                <span className="text-xs text-muted-foreground">{t('enabled')}</span>
                                                <Switch
                                                    checked={k.enabled}
                                                    onCheckedChange={(checked) => handleUpdateKey(idx, { enabled: checked })}
                                                />
                                            </div>
                                        </div>

                                        <div className="grid gap-2 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
                                            <Input
                                                type="text"
                                                value={k.remark ?? ''}
                                                onChange={(e) => handleUpdateKey(idx, { remark: e.target.value })}
                                                placeholder={t('remarkPlaceholder')}
                                                className="rounded-xl bg-background"
                                            />
                                            <div className="flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground md:justify-end">
                                                <span className="rounded-full border border-border/60 bg-muted/20 px-2.5 py-1">{maskedKey || t('keyValueEmpty')}</span>
                                                <span className="rounded-full border border-border/60 bg-muted/20 px-2.5 py-1">{sourceTypeSummary}</span>
                                                <span className="rounded-full border border-border/60 bg-muted/20 px-2.5 py-1">{remarkSummary}</span>
                                            </div>
                                        </div>
                                    </div>

                                    <div className="flex items-center justify-end gap-2 lg:self-start">
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="sm"
                                            onClick={() => handleRemoveKey(idx)}
                                            disabled={(formData.keys ?? []).length <= 1}
                                            className="h-9 w-9 rounded-xl p-0 text-muted-foreground hover:bg-transparent hover:text-destructive disabled:opacity-40"
                                            title={t('remove')}
                                        >
                                            <X className="h-4 w-4" />
                                        </Button>
                                    </div>
                                </div>
                            </div>
                        );
                    })}
                </div>
                )
                ) : (
                    <div className="rounded-2xl border border-dashed border-border/60 bg-background/60 px-4 py-6 text-sm text-muted-foreground">
                        {tDetail('noKeysMatched')}
                    </div>
                )}
            </div>
            )}

            {!isClassifiedMode && (
                <div className="space-y-2">
                    <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                        <div className="min-w-0 flex items-center gap-2 text-sm font-medium text-card-foreground">
                            <label>{flowCopy.modelSectionTitle}</label>
                        </div>
                        <div className="flex items-center gap-1 self-start">
                            {canOpenGlobalModelDialog && (
                                <Button
                                    type="button"
                                    variant="outline"
                                    data-testid={`${idPrefix}-global-fetch-models`}
                                    disabled={fetchModel.isPending}
                                    onClick={handleOpenGlobalModelDialog}
                                    className="h-10 rounded-xl px-4 text-sm gap-2"
                                >
                                    {fetchModel.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Search className="h-4 w-4" />}
                                    {t('selectModels')}
                                </Button>
                            )}
                        </div>
                    </div>
                    <input type="hidden" value={formData.model} />

                    <div className="relative">
                        <Input
                            ref={inputRef}
                            id={`${idPrefix}-model-custom`}
                            type="text"
                            value={inputValue}
                            onChange={(e) => setInputValue(e.target.value)}
                            onKeyDown={handleInputKeyDown}
                            placeholder={t('modelCustomPlaceholder')}
                            className="pr-10 rounded-xl"
                        />
                        {inputValue.trim() && !customModels.includes(inputValue.trim()) && !autoModels.includes(inputValue.trim()) && (
                            <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                onClick={() => handleAddModel(inputValue)}
                                className="absolute rounded-lg right-1 top-1/2 -translate-y-1/2 h-7 w-7 p-0 text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
                                title={t('modelAdd')}
                            >
                                <Plus className="size-4" />
                            </Button>
                        )}
                    </div>

                    <div className="space-y-2">
                        <div className="flex items-center justify-between">
                            <label className="text-xs font-medium text-card-foreground">
                                {t('modelSelected')} {(autoModels.length + customModels.length) > 0 && `(${autoModels.length + customModels.length})`}
                            </label>
                            {(autoModels.length + customModels.length) > 0 && (
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="sm"
                                    onClick={() => {
                                        updateModels([], []);
                                    }}
                                    className="h-6 px-2 text-xs text-muted-foreground/50 hover:text-muted-foreground hover:bg-transparent"
                                >
                                    {t('modelClearAll')}
                                </Button>
                            )}
                        </div>
                        <div className="rounded-xl border border-border bg-muted/30 p-2.5 max-h-40 min-h-12 overflow-y-auto">
                            {(autoModels.length + customModels.length) > 0 ? (
                                <div className="flex flex-wrap gap-1.5">
                                    {autoModels.map((model) => (
                                        <Badge key={model} className="bg-primary hover:bg-primary/90">
                                            {model}
                                            <button
                                                type="button"
                                                onClick={() => handleRemoveAutoModel(model)}
                                                className="ml-1 rounded-sm opacity-70 hover:opacity-100 focus:outline-none focus:ring-1 focus:ring-ring"
                                            >
                                                <X className="h-3 w-3" />
                                            </button>
                                        </Badge>
                                    ))}
                                    {customModels.map((model) => (
                                        <Badge key={model} className="bg-primary hover:bg-primary/90">
                                            {model}
                                            <button
                                                type="button"
                                                onClick={() => handleRemoveCustomModel(model)}
                                                className="ml-1 rounded-sm opacity-70 hover:opacity-100 focus:outline-none focus:ring-1 focus:ring-ring"
                                            >
                                                <X className="h-3 w-3" />
                                            </button>
                                        </Badge>
                                    ))}
                                </div>
                            ) : (
                                <div className="flex min-h-12 flex-col items-center justify-center gap-1 text-center text-xs text-muted-foreground">
                                    <div>{t('modelNoSelected')}</div>
                                    <div className="max-w-md text-xs leading-5 text-muted-foreground">{t('modelNoSelectedHint')}</div>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            )}

            <Accordion type="single" collapsible className="w-full border rounded-xl bg-card">
                <AccordionItem value="advanced" className="border-none">
                    <AccordionTrigger
                        className="text-sm font-medium text-card-foreground py-3 px-4 hover:no-underline hover:bg-muted/30 rounded-xl transition-colors"
                        addon={null}
                    >
                        <div>
                            <div>{flowCopy.advancedSectionTitle}</div>
                        </div>
                    </AccordionTrigger>
                    <AccordionContent className="pt-4 px-4 pb-4 space-y-4 border-t">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                                    <label htmlFor={`${idPrefix}-auto-group`}>{t('autoGroup')}</label>
                                </div>
                                <Select
                                    value={String(formData.auto_group)}
                                    onValueChange={(value) => onFormDataChange({ ...formData, auto_group: Number(value) as AutoGroupType })}
                                >
                                    <SelectTrigger id={`${idPrefix}-auto-group`} className="rounded-xl w-full border border-border px-4 py-2 text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent className='rounded-xl'>
                                        <SelectItem className='rounded-xl' value={String(AutoGroupType.None)}>{t('autoGroupNone')}</SelectItem>
                                        <SelectItem className='rounded-xl' value={String(AutoGroupType.Fuzzy)}>{t('autoGroupFuzzy')}</SelectItem>
                                        <SelectItem className='rounded-xl' value={String(AutoGroupType.Exact)}>{t('autoGroupExact')}</SelectItem>
                                        <SelectItem className='rounded-xl' value={String(AutoGroupType.Regex)}>{t('autoGroupRegex')}</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>

                            <div className="space-y-2">
                                <label htmlFor={`${idPrefix}-channel-proxy`} className="text-sm font-medium text-card-foreground">
                                    {t('channelProxy')}
                                </label>
                                <Input
                                    id={`${idPrefix}-channel-proxy`}
                                    type="text"
                                    value={formData.channel_proxy}
                                    onChange={(e) => onFormDataChange({ ...formData, channel_proxy: e.target.value })}
                                    placeholder={t('channelProxyPlaceholder')}
                                    className="rounded-xl"
                                />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <div className="flex items-center justify-between">
                                <label className="text-sm font-medium text-card-foreground">
                                    {t('customHeader')} {formData.custom_header.length > 0 ? `(${formData.custom_header.length})` : ''}
                                </label>
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="sm"
                                    onClick={handleAddHeader}
                                    className="h-6 px-2 text-xs text-muted-foreground/70 hover:text-muted-foreground hover:bg-transparent"
                                >
                                    <Plus className="h-3 w-3 mr-1" />
                                    {t('customHeaderAdd')}
                                </Button>
                            </div>
                            <div className="space-y-2">
                                {(formData.custom_header ?? []).map((h, idx) => (
                                    <div key={`hdr-${idx}`} className="flex items-center gap-2">
                                        <Input
                                            type="text"
                                            value={h.header_key}
                                            onChange={(e) => handleUpdateHeader(idx, { header_key: e.target.value })}
                                            placeholder={t('customHeaderKey')}
                                            className="rounded-xl flex-1"
                                        />
                                        <Input
                                            type="text"
                                            value={h.header_value}
                                            onChange={(e) => handleUpdateHeader(idx, { header_value: e.target.value })}
                                            placeholder={t('customHeaderValue')}
                                            className="rounded-xl flex-1"
                                        />
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="sm"
                                            onClick={() => handleRemoveHeader(idx)}
                                            disabled={(formData.custom_header ?? []).length <= 1}
                                            className="h-8 w-8 p-0 rounded-xl text-muted-foreground hover:text-destructive hover:bg-transparent disabled:opacity-40"
                                            title={t('remove')}
                                        >
                                            <X className="h-4 w-4" />
                                        </Button>
                                    </div>
                                ))}
                            </div>
                        </div>

                        <div className="space-y-2">
                            <label htmlFor={`${idPrefix}-match-regex`} className="text-sm font-medium text-card-foreground">
                                {t('matchRegex')}
                            </label>
                            <Input
                                id={`${idPrefix}-match-regex`}
                                type="text"
                                value={formData.match_regex}
                                onChange={(e) => onFormDataChange({ ...formData, match_regex: e.target.value })}
                                placeholder={t('matchRegexPlaceholder')}
                                className="rounded-xl"
                            />
                        </div>

                        <div className="space-y-2">
                            <label htmlFor={`${idPrefix}-param-override`} className="text-sm font-medium text-card-foreground">
                                {t('paramOverride')}
                            </label>
                            <textarea
                                id={`${idPrefix}-param-override`}
                                value={formData.param_override}
                                onChange={(e) => onFormDataChange({ ...formData, param_override: e.target.value })}
                                placeholder={t('paramOverridePlaceholder')}
                                className="min-h-28 w-full rounded-xl border border-border bg-background px-3 py-2 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                            />
                        </div>

                        <Accordion type="single" collapsible className="w-full border rounded-xl bg-muted/20">
                            <AccordionItem value="route-target-overrides" className="border-none">
                                <AccordionTrigger
                                    className="text-sm font-medium text-card-foreground px-4 py-3 hover:no-underline hover:bg-muted/30 rounded-xl transition-colors"
                                >
                                    <div className="w-full flex items-center justify-between gap-3">
                                        <div>
                                            <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                                                <span>{t('routeTargetTitle')}</span>
                                                <HelpHint className="size-3.5">{t('routeTargetHint')}</HelpHint>
                                            </div>
                                            <div className="mt-1 text-xs text-muted-foreground">
                                                <span>{t('routeTargetDesc')}</span>
                                            </div>
                                        </div>
                                        <Badge variant="outline" className="shrink-0">{t('routeTargetCount', { count: currentRouteTargetOverrides.length })}</Badge>
                                    </div>
                                </AccordionTrigger>
                                <AccordionContent className="pt-4 px-4 pb-4 border-t space-y-3">
                                    <div className="space-y-2 text-xs text-muted-foreground">
                                        {channelId ? currentRouteTargetOverrides.slice(0, 3).map((row) => (
                                            <div key={`${row.channel_id}-${row.channel_key_id}-${row.model_name}`} className="rounded-lg bg-background/80 px-3 py-2 break-all">
                                                {formatRouteTargetSummary(row)}
                                            </div>
                                        )) : (
                                            <div className="rounded-lg bg-background/80 px-3 py-2">{t('routeTargetSaveFirst')}</div>
                                        )}
                                        {channelId && currentRouteTargetOverrides.length > 3 && (
                                            <div className="text-[11px]">{t('routeTargetPreviewMore', { count: currentRouteTargetOverrides.length - 3 })}</div>
                                        )}
                                    </div>
                                    <div className="flex justify-end">
                                        <Button
                                            type="button"
                                            variant="outline"
                                            size="sm"
                                            className="rounded-xl"
                                            disabled={!channelId}
                                            onClick={handleOpenRouteTargetDialog}
                                        >
                                            {t('routeTargetManage')}
                                        </Button>
                                    </div>
                                </AccordionContent>
                            </AccordionItem>
                        </Accordion>
                    </AccordionContent>
                </AccordionItem>
            </Accordion>

            <div className="flex flex-wrap items-center justify-between gap-4 p-4 rounded-xl bg-muted/20 border border-border/50">
                <label className="flex items-center gap-2 cursor-pointer">
                    <Switch
                        checked={formData.enabled}
                        onCheckedChange={(checked) => onFormDataChange({ ...formData, enabled: checked })}
                    />
                    <span className="text-sm font-medium text-card-foreground">{t('enabled')}</span>
                </label>
                <div className="flex items-center gap-6">
                    <label className="flex items-center gap-2 cursor-pointer">
                        <Switch
                            checked={formData.proxy}
                            onCheckedChange={(checked) => onFormDataChange({ ...formData, proxy: checked })}
                        />
                        <span className="text-sm text-card-foreground">{t('proxy')}</span>
                    </label>
                    <label className="flex items-center gap-2 cursor-pointer">
                        <Switch
                            checked={formData.auto_sync}
                            onCheckedChange={(checked) => onFormDataChange({ ...formData, auto_sync: checked })}
                        />
                        <span className="text-sm text-card-foreground">{t('autoSync')}</span>
                    </label>
                </div>
            </div>

            <div className={`flex flex-col gap-3 pt-2 ${onCancel ? 'lg:flex-row' : ''}`}>
                {onCancel && cancelText && (
                    <Button
                        type="button"
                        variant="secondary"
                        onClick={onCancel}
                        className="h-12 w-full rounded-2xl lg:flex-1"
                    >
                        {cancelText}
                    </Button>
                )}
                <div className="grid w-full gap-2 sm:grid-cols-2 lg:flex lg:flex-1">
                    <Button
                        type="button"
                        variant="outline"
                        disabled={isTesting || allModels.length === 0}
                        onClick={handleTestFirst}
                        className="h-12 rounded-2xl lg:flex-1"
                        title={t('testFirstTitle')}
                    >
                        {isTesting ? (
                            <Loader2 className="h-4 w-4 mr-1 animate-spin" />
                        ) : null}
                        {t('testFirst')}
                    </Button>
                    <Button
                        type="button"
                        variant="outline"
                        disabled={isTesting || allModels.length === 0}
                        onClick={handleTestAll}
                        className="h-12 rounded-2xl lg:flex-1"
                        title={t('testAllTitle')}
                    >
                        {isTesting ? (
                            <Loader2 className="h-4 w-4 mr-1 animate-spin" />
                        ) : null}
                        {t('testAll')}
                    </Button>
                    <Button
                        type="submit"
                        disabled={isPending}
                        className="h-12 rounded-2xl sm:col-span-2 lg:flex-[1.35]"
                    >
                        {isPending ? pendingText : submitText}
                    </Button>
                </div>
            </div>

            {/* 测试结果摘要 */}
            {testResults.size > 0 && (
                <div className="rounded-xl border border-border bg-muted/20 p-3 space-y-2">
                    <div className="text-xs font-medium text-card-foreground">{t('testResultTitle')}</div>
                    <div className="space-y-1 max-h-40 overflow-y-auto">
                        {Array.from(testResults.entries()).map(([model, result]) => (
                            <div key={model} className="flex items-center gap-2 text-xs">
                                {result.passed ? (
                                    <CheckCircle2 className="h-3.5 w-3.5 text-green-500 shrink-0" />
                                ) : (
                                    <XCircle className="h-3.5 w-3.5 text-red-500 shrink-0" />
                                )}
                                <span className="font-mono flex-1 truncate">{model}</span>
                                {result.delay !== undefined && (
                                    <span className="text-muted-foreground">{result.delay}ms</span>
                                )}
                                {result.source_type && (
                                    <span className="text-muted-foreground">{formatSourceTypeLabel(result.source_type)}</span>
                                )}
                                {result.billing_mode && (
                                    <span className="text-muted-foreground">{formatBillingModeLabel(result.billing_mode)}</span>
                                )}
                                {result.probe_policy && (
                                    <span className="text-muted-foreground">{formatProbePolicyLabel(result.probe_policy)}</span>
                                )}
                                {result.policy_basis && (
                                    <span className="text-muted-foreground truncate max-w-40" title={result.policy_basis}>{t('testPolicyBasis', { basis: result.policy_basis })}</span>
                                )}
                                {result.error && (
                                    <span className="text-red-500 truncate max-w-32" title={result.error}>{result.error}</span>
                                )}
                            </div>
                        ))}
                    </div>
                    <div className="text-xs text-muted-foreground">
                        {t('testResultSummary', {
                            total: testResults.size,
                            passed: Array.from(testResults.values()).filter((r) => r.passed).length,
                        })}
                    </div>
                </div>
            )}
        </form>
        {/* 选择模型弹框 - 放在 form 外避免事件冒泡关闭外层弹框 */}
        <Dialog open={showModelSelectDialog} onOpenChange={handleModelDialogOpenChange}>
            <DialogContent className="max-w-md flex flex-col max-h-[80vh] overflow-hidden">
                <div data-testid={`${idPrefix}-model-select-dialog`} className="contents">
                <DialogHeader className="shrink-0">
                    <div className="flex items-center justify-between">
                        <div className="space-y-1">
                            <DialogTitle data-testid={`${idPrefix}-model-select-title`}>
                                {keyModelDialog?.keyIndex != null && keyModelDialog.keyIndex >= 0
                                    ? t('keyFetchedModelList', { index: keyModelDialog.keyIndex + 1 })
                                    : t('fetchedModelList')}
                            </DialogTitle>
                            <div className="text-xs text-muted-foreground">
                                {modelDialogSelectedCount > 0 ? `${t('modelSelected')} ${modelDialogSelectedCount}` : t('modelNoSelected')}
                            </div>
                        </div>
                        {fetchModel.isPending && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
                    </div>
                </DialogHeader>
                {/* 全选行 */}
                {fetchModel.isPending ? (
                    <div className="flex items-center justify-center py-8">
                        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                    </div>
                ) : (keyModelDialog?.models.length ?? 0) === 0 ? (
                    <div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
                        {tModels('noModels')}
                    </div>
                ) : (
                    <>
                    <div className="shrink-0 pb-2">
                        <Input
                            data-testid={`${idPrefix}-model-search-input`}
                            value={modelDialogKeyword}
                            onChange={(event) => setModelDialogKeyword(event.target.value)}
                            placeholder={t('modelSearchPlaceholder')}
                            className="rounded-xl"
                        />
                    </div>
                    <div
                        data-testid={`${idPrefix}-model-select-all`}
                        className="flex items-center gap-2 px-1 py-1 cursor-pointer hover:bg-accent/5 rounded-lg shrink-0"
                        onClick={() => {
                            if (!keyModelDialog) return;
                            const nextSelected = new Set(keyModelDialog.selected);
                            if (filteredDialogModels.length > 0 && filteredDialogModels.every((model) => keyModelDialog.selected.has(model))) {
                                filteredDialogModels.forEach((model) => nextSelected.delete(model));
                            } else {
                                filteredDialogModels.forEach((model) => nextSelected.add(model));
                            }
                            setKeyModelDialog({ ...keyModelDialog, selected: nextSelected });
                        }}
                    >
                        <div className="size-4 shrink-0 rounded border border-primary flex items-center justify-center">
                            {keyModelDialog && filteredDialogModels.length > 0 && filteredDialogModels.every((model) => keyModelDialog.selected.has(model)) && (
                                <Check className="h-3 w-3 text-primary" />
                            )}
                        </div>
                        <span className="text-sm font-medium">{t('modelSelectAllFetched')}</span>
                    </div>
                <div className="border-t shrink-0" />
                {/* 模型列表 - 直接作为 flex 子项，自身滑动 */}
                <div
                    data-testid={`${idPrefix}-model-select-list`}
                    className="flex-1 min-h-0 overflow-y-auto space-y-0.5 py-1 dialog-model-scrollbar"
                    style={{ scrollbarWidth: 'thin', msOverflowStyle: 'auto' }}
                >
                    {filteredDialogModels.map((model) => (
                        <div
                            key={model}
                            data-testid={`${idPrefix}-model-option-${model}`}
                            className="flex items-center gap-2 px-1 py-1.5 cursor-pointer hover:bg-accent/5 rounded-lg"
                            onClick={() => {
                                if (!keyModelDialog) return;
                                const next = new Set(keyModelDialog.selected);
                                if (next.has(model)) next.delete(model);
                                else next.add(model);
                                setKeyModelDialog({ ...keyModelDialog, selected: next });
                            }}
                        >
                            <div className="size-4 shrink-0 rounded border border-primary flex items-center justify-center">
                                {keyModelDialog?.selected.has(model) && (
                                    <Check className="h-3 w-3 text-primary" />
                                )}
                            </div>
                            <span className="font-mono text-sm">{model}</span>
                        </div>
                    ))}
                    {filteredDialogModels.length === 0 && (
                        <div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
                            {tModels('noModels')}
                        </div>
                    )}
                    </div>
                    </>
                )}
                <DialogFooter className="shrink-0">
                    <Button
                        type="button"
                        variant="outline"
                        data-testid={`${idPrefix}-model-select-cancel`}
                        onClick={() => handleModelDialogOpenChange(false)}
                        className="rounded-xl"
                    >
                        {t('selectModelsCancel')}
                    </Button>
                    <Button
                        type="button"
                        data-testid={`${idPrefix}-model-select-confirm`}
                        onClick={keyModelDialog?.keyIndex != null && keyModelDialog.keyIndex >= 0 ? handleConfirmKeyModelSelect : handleConfirmModelSelect}
                        className="rounded-xl"
                    >
                        {t('selectModelsConfirm')}
                    </Button>
                </DialogFooter>
                </div>
            </DialogContent>
        </Dialog>
        <Dialog open={showRouteTargetDialog} onOpenChange={setShowRouteTargetDialog}>
            <DialogContent className="max-w-2xl flex flex-col max-h-[85vh] overflow-hidden">
                <DialogHeader>
                    <DialogTitle>{t('routeTargetDialogTitle')}</DialogTitle>
                </DialogHeader>
                <div className="space-y-4 overflow-y-auto pr-1">
                    <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                        <div className="space-y-2">
                            <label className="text-sm font-medium text-card-foreground">{t('routeTargetChannelKey')}</label>
                            <Select value={routeTargetForm.channel_key_id} onValueChange={(value) => setRouteTargetForm((current) => ({ ...current, channel_key_id: value }))}>
                                <SelectTrigger className="rounded-xl bg-background">
                                    <SelectValue placeholder={t('routeTargetChannelKeyPlaceholder')} />
                                </SelectTrigger>
                                <SelectContent>
                                    {currentChannelKeyOptions.map((key) => (
                                        <SelectItem key={key.id} value={String(key.id)}>{key.label}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-medium text-card-foreground">{t('routeTargetModel')}</label>
                            <Input value={routeTargetForm.model_name} onChange={(e) => setRouteTargetForm((current) => ({ ...current, model_name: e.target.value }))} placeholder={t('routeTargetModelPlaceholder')} className="rounded-xl" />
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-medium text-card-foreground">{tRouteTarget('billingMode')}</label>
                            <Select value={routeTargetForm.billing_mode} onValueChange={(value) => setRouteTargetForm((current) => ({ ...current, billing_mode: value }))}>
                                <SelectTrigger className="rounded-xl bg-background">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    {BILLING_MODE_OPTIONS.map((option) => (
                                        <SelectItem key={option} value={option}>{formatBillingModeLabel(option)}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-medium text-card-foreground">{tRouteTarget('probePolicy')}</label>
                            <Select value={routeTargetForm.probe_policy} onValueChange={(value) => setRouteTargetForm((current) => ({ ...current, probe_policy: value }))}>
                                <SelectTrigger className="rounded-xl bg-background">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    {PROBE_POLICY_OPTIONS.map((option) => (
                                        <SelectItem key={option} value={option}>{formatProbePolicyLabel(option)}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-medium text-card-foreground">{tRouteTarget('probeInterval')}</label>
                            <Input value={routeTargetForm.probe_interval_seconds} onChange={(e) => setRouteTargetForm((current) => ({ ...current, probe_interval_seconds: e.target.value }))} className="rounded-xl" />
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-medium text-card-foreground">{tRouteTarget('probeConcurrency')}</label>
                            <Input value={routeTargetForm.probe_concurrency_limit} onChange={(e) => setRouteTargetForm((current) => ({ ...current, probe_concurrency_limit: e.target.value }))} className="rounded-xl" />
                        </div>
                    </div>
                    <div className="space-y-2">
                        <div className="text-sm font-medium text-card-foreground">{t('routeTargetExisting')}</div>
                        <div className="space-y-2 max-h-64 overflow-y-auto">
                            {currentRouteTargetOverrides.length > 0 ? currentRouteTargetOverrides.map((row) => (
                                <div key={`${row.channel_id}-${row.channel_key_id}-${row.model_name}`} className="flex flex-col gap-2 rounded-xl border border-border/70 bg-background px-3 py-3 text-xs">
                                    <div className="break-all">{formatRouteTargetSummary(row)}</div>
                                    <div className="flex items-center gap-2">
                                        <Button type="button" size="sm" variant="outline" className="rounded-xl" onClick={() => handleEditRouteTargetOverrideRow(row)}>
                                            {t('routeTargetEdit')}
                                        </Button>
                                        <Button type="button" size="sm" variant="destructive" className="rounded-xl" onClick={() => void handleDeleteRouteTargetOverrideRow(row)} disabled={deleteRouteTargetOverride.isPending}>
                                            {t('routeTargetDelete')}
                                        </Button>
                                    </div>
                                </div>
                            )) : (
                                <div className="rounded-xl border border-dashed border-border px-3 py-4 text-xs text-muted-foreground">{t('routeTargetEmpty')}</div>
                            )}
                        </div>
                    </div>
                </div>
                <DialogFooter>
                    <Button type="button" variant="outline" className="rounded-xl" onClick={() => setShowRouteTargetDialog(false)}>
                        {t('close')}
                    </Button>
                    <Button type="button" className="rounded-xl" onClick={() => void handleApplyRouteTargetOverride()} disabled={upsertRouteTargetOverride.isPending || !channelId}>
                        {upsertRouteTargetOverride.isPending ? t('routeTargetSaving') : t('routeTargetSave')}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
        </>
    );
}
