'use client';

import { useState, useMemo, useEffect } from 'react';
import { useTranslations } from 'next-intl';
import { Copy, Check, BookOpen, X, HelpCircle, ExternalLink } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { useGroupList } from '@/api/endpoints/group';
import { useAPIKeyList } from '@/api/endpoints/apikey';
import { useSettingList, SettingKey } from '@/api/endpoints/setting';
import { motion, AnimatePresence } from 'motion/react';
import { cn } from '@/lib/utils';
import { HelpHint } from '@/components/common/HelpHint';

type ApiType = 'openai-chat' | 'openai-responses' | 'anthropic';
type ContentTab = 'curl' | 'ccswitch';
type CCSwitchAppType = 'claude' | 'codex';

const API_PATHS: Record<ApiType, string> = {
    'openai-chat': '/v1/chat/completions',
    'openai-responses': '/v1/responses',
    'anthropic': '/v1/messages',
};

function generateCurl(baseUrl: string, apiKey: string, model: string, apiType: ApiType): string {
    const path = API_PATHS[apiType];
    const url = `${baseUrl}${path}`;

    if (apiType === 'anthropic') {
        return `curl -X POST '${url}' \\
  -H 'Content-Type: application/json' \\
  -H 'x-api-key: ${apiKey || 'YOUR_API_KEY'}' \\
  -H 'anthropic-version: 2023-06-01' \\
  -d '{
    "model": "${model || 'YOUR_MODEL'}",
    "max_tokens": 1024,
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'`;
    }

    if (apiType === 'openai-responses') {
        return `curl -X POST '${url}' \\
  -H 'Content-Type: application/json' \\
  -H 'Authorization: Bearer ${apiKey || 'YOUR_API_KEY'}' \\
  -d '{
    "model": "${model || 'YOUR_MODEL'}",
    "input": "Hello!"
  }'`;
    }

    return `curl -X POST '${url}' \\
  -H 'Content-Type: application/json' \\
  -H 'Authorization: Bearer ${apiKey || 'YOUR_API_KEY'}' \\
  -d '{
    "model": "${model || 'YOUR_MODEL'}",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'`;
}

interface CCSwitchForm {
    appType: CCSwitchAppType;
    name: string;
    model: string;
    haikuModel: string;
    sonnetModel: string;
    opusModel: string;
}

function buildCCSwitchUrl(baseUrl: string, apiKey: string, form: CCSwitchForm): string {
    const params = new URLSearchParams();
    params.set('resource', 'provider');
    params.set('app', form.appType);
    params.set('name', form.name);
    params.set('endpoint', form.appType === 'codex' ? `${baseUrl}/v1` : baseUrl);
    params.set('apiKey', apiKey);
    params.set('model', form.model);
    params.set('homepage', baseUrl);
    params.set('enabled', 'true');
    if (form.appType === 'claude') {
        if (form.haikuModel) params.set('haikuModel', form.haikuModel);
        if (form.sonnetModel) params.set('sonnetModel', form.sonnetModel);
        if (form.opusModel) params.set('opusModel', form.opusModel);
    }
    return `ccswitch://v1/import?${params.toString()}`;
}

interface DocModalProps {
    isOpen: boolean;
    onClose: () => void;
    onGoSetting?: () => void;
}

export function DocModal({ isOpen, onClose, onGoSetting }: DocModalProps) {
    const t = useTranslations('doc');
    const [contentTab, setContentTab] = useState<ContentTab>('curl');
    const [apiType, setApiType] = useState<ApiType>('openai-chat');
    const [selectedApiKey, setSelectedApiKey] = useState<string>('');
    const [selectedModel, setSelectedModel] = useState<string>('');
    const [copied, setCopied] = useState(false);
    const [nameEdited, setNameEdited] = useState(false);
    const [ccswitchForm, setCcswitchForm] = useState<CCSwitchForm>({
        appType: 'claude',
        name: '',
        model: '',
        haikuModel: '',
        sonnetModel: '',
        opusModel: '',
    });

    const { data: groups } = useGroupList();
    const { data: apiKeys } = useAPIKeyList();
    const { data: settings } = useSettingList();

    const baseUrl = useMemo(() => {
        const setting = settings?.find(s => s.key === SettingKey.ApiBaseUrl);
        return setting?.value?.trim() || 'http://127.0.0.1:1088';
    }, [settings]);

    const curlCode = useMemo(
        () => generateCurl(baseUrl, selectedApiKey, selectedModel, apiType),
        [baseUrl, selectedApiKey, selectedModel, apiType]
    );

    const invalidGroupNames = useMemo(
        () => (groups ?? []).filter((g) => /[:：\s]/.test(g.name)).map((g) => g.name),
        [groups]
    );

    const selectedApiKeyRecord = useMemo(
        () => (apiKeys ?? []).find((k) => k.api_key === selectedApiKey),
        [apiKeys, selectedApiKey]
    );

    const allowedGroupsByKey = useMemo(() => {
        const raw = selectedApiKeyRecord?.supported_models?.trim();
        if (!raw) return null;
        return new Set(raw.split(',').map((m) => m.trim()).filter(Boolean));
    }, [selectedApiKeyRecord]);

    const groupOptions = useMemo(() => {
        const all = Array.from(new Set((groups ?? []).map((g) => g.name).filter(Boolean)));
        if (!allowedGroupsByKey) {
            return all.map((name) => ({ value: name, label: name }));
        }
        return all
            .filter((name) => allowedGroupsByKey.has(name))
            .map((name) => ({ value: name, label: name }));
    }, [groups, allowedGroupsByKey]);

    const hasGroupOption = useMemo(() => new Set(groupOptions.map((o) => o.value)), [groupOptions]);

    useEffect(() => {
        if (selectedModel && !hasGroupOption.has(selectedModel)) {
            setSelectedModel('');
        }
        setCcswitchForm((prev) => {
            const next = { ...prev };
            if (next.model && !hasGroupOption.has(next.model)) next.model = '';
            if (next.haikuModel && !hasGroupOption.has(next.haikuModel)) next.haikuModel = '';
            if (next.sonnetModel && !hasGroupOption.has(next.sonnetModel)) next.sonnetModel = '';
            if (next.opusModel && !hasGroupOption.has(next.opusModel)) next.opusModel = '';
            return next;
        });
    }, [hasGroupOption, selectedModel]);

    const handleCopy = async () => {
        try {
            await navigator.clipboard.writeText(curlCode);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch {
            // fallback
        }
    };

    const updateCCSwitch = (patch: Partial<CCSwitchForm>) =>
        setCcswitchForm(prev => ({ ...prev, ...patch }));

    const autoName = useMemo(() => {
        if (!selectedApiKey || !ccswitchForm.model) return '';
        return `octopus_${ccswitchForm.appType}_${ccswitchForm.model}`;
    }, [selectedApiKey, ccswitchForm.appType, ccswitchForm.model]);

    useEffect(() => {
        if (!autoName) return;
        if (!nameEdited || ccswitchForm.name === '' || ccswitchForm.name.startsWith('octopus_')) {
            setCcswitchForm((prev) => ({ ...prev, name: autoName }));
        }
    }, [autoName, nameEdited, ccswitchForm.name]);

    const ccswitchReady =
        ccswitchForm.name.trim() !== '' &&
        ccswitchForm.model !== '' &&
        selectedApiKey !== '';
    const ccswitchHasGroupOptions = groupOptions.length > 0;
    const ccswitchCanChooseModel = selectedApiKey !== '' && ccswitchHasGroupOptions;
    const ccswitchCanConfirmName = ccswitchCanChooseModel && ccswitchForm.model.trim() !== '';
    const ccswitchCanOpenAdvanced =
        ccswitchForm.appType === 'claude' &&
        ccswitchCanConfirmName &&
        ccswitchForm.name.trim() !== '';
    const ccswitchModelBlockedHint =
        selectedApiKey === ''
            ? t('ccswitchModelLockedHint')
            : t('ccswitchModelUnavailableHint');
    const ccswitchImportBlockedHint =
        selectedApiKey === ''
            ? t('ccswitchImportLockedHintKey')
            : !ccswitchHasGroupOptions
              ? t('ccswitchImportLockedHintGroup')
              : ccswitchForm.model.trim() === ''
                ? t('ccswitchImportLockedHintModel')
                : ccswitchForm.name.trim() === ''
                  ? t('ccswitchImportLockedHintName')
                  : '';

    const ccswitchProgress = useMemo(
        () => [
            {
                key: 'client',
                label: t('ccswitchStepClient'),
                done: true,
            },
            {
                key: 'key',
                label: t('ccswitchStepKey'),
                done: selectedApiKey !== '',
            },
            {
                key: 'model',
                label: t('ccswitchStepModel'),
                done: ccswitchForm.model.trim() !== '',
            },
            {
                key: 'name',
                label: t('ccswitchStepName'),
                done: ccswitchForm.name.trim() !== '',
            },
        ],
        [ccswitchForm.model, ccswitchForm.name, selectedApiKey, t]
    );

    const handleCCSwitchImport = () => {
        if (!ccswitchReady) return;
        const url = buildCCSwitchUrl(baseUrl, selectedApiKey, ccswitchForm);
        window.open(url, '_blank', 'noopener,noreferrer');
    };

    return (
        <AnimatePresence>
            {isOpen && (
                <>
                    {/* Backdrop */}
                    <motion.div
                        className="fixed inset-0 bg-black/40 z-50"
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        onClick={onClose}
                    />
                    {/* Modal */}
                    <motion.div
                        data-testid="doc-modal"
                        className="fixed inset-x-4 bottom-4 top-4 md:inset-auto md:left-1/2 md:top-1/2 md:-translate-x-1/2 md:-translate-y-1/2 md:w-[640px] md:max-h-[80vh] z-50 flex flex-col bg-card rounded-3xl border border-border shadow-2xl overflow-hidden"
                        initial={{ opacity: 0, scale: 0.95, y: 20 }}
                        animate={{ opacity: 1, scale: 1, y: 0 }}
                        exit={{ opacity: 0, scale: 0.95, y: 20 }}
                        transition={{ type: 'spring', stiffness: 300, damping: 30 }}
                    >
                        {/* Header */}
                        <div data-testid="doc-modal-header" className="flex items-center justify-between border-b border-border p-4 shrink-0 sm:p-6">
                            <div className="flex items-center gap-2">
                                <BookOpen className="h-5 w-5 text-primary" />
                                <h2 className="text-lg font-bold text-card-foreground">{t('title')}</h2>
                            </div>
                            <button
                                onClick={onClose}
                                className="p-1.5 rounded-xl text-muted-foreground hover:text-card-foreground hover:bg-muted/50 transition-colors"
                            >
                                <X className="h-5 w-5" />
                            </button>
                        </div>

                        {/* Tabs */}
                        <div data-testid="doc-modal-tabs" className="flex shrink-0 overflow-x-auto border-b border-border px-4 sm:px-6">
                            {(['curl', 'ccswitch'] as ContentTab[]).map((tab) => (
                                <button
                                    key={tab}
                                    onClick={() => setContentTab(tab)}
                                    data-testid={tab === 'curl' ? 'doc-modal-tab-curl' : 'doc-modal-tab-ccswitch'}
                                    className={cn(
                                        'whitespace-nowrap px-4 py-3 text-sm font-medium border-b-2 transition-colors -mb-px',
                                        contentTab === tab
                                            ? 'border-primary text-card-foreground'
                                            : 'border-transparent text-muted-foreground hover:text-card-foreground'
                                    )}
                                >
                                    {tab === 'curl' ? t('tabCurl') : t('tabCCSwitch')}
                                </button>
                            ))}
                        </div>

                        {/* Content */}
                        <div data-testid="doc-modal-content" className="flex-1 overflow-y-auto p-4 space-y-5 sm:p-6">
                            {contentTab === 'curl' && (
                                <>
                                    {/* 第一行：API 地址 + API 类型 */}
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                        <div className="space-y-1">
                                            <div className="flex items-center gap-1.5">
                                                <label className="text-sm font-medium text-muted-foreground">{t('baseUrl')}</label>
                                                <button
                                                    type="button"
                                                    onClick={() => {
                                                        onClose();
                                                        onGoSetting?.();
                                                    }}
                                                    className="inline-flex items-center gap-1 text-xs text-primary hover:underline"
                                                >
                                                    <HelpCircle className="h-3.5 w-3.5" />
                                                    {t('baseUrlTip')}
                                                </button>
                                            </div>
                                            <div className="font-mono text-sm bg-muted/30 rounded-xl px-3 py-2 text-card-foreground break-all truncate">{baseUrl}</div>
                                        </div>
                                        <div className="space-y-2">
                                            <label className="text-sm font-medium text-card-foreground">{t('apiType')}</label>
                                            <Select value={apiType} onValueChange={(v) => setApiType(v as ApiType)}>
                                                <SelectTrigger className="rounded-xl">
                                                    <SelectValue />
                                                </SelectTrigger>
                                                <SelectContent className="rounded-xl">
                                                    <SelectItem className="rounded-xl" value="openai-chat">{t('typeOpenAIChat')}</SelectItem>
                                                    <SelectItem className="rounded-xl" value="openai-responses">{t('typeOpenAIResponses')}</SelectItem>
                                                    <SelectItem className="rounded-xl" value="anthropic">{t('typeAnthropic')}</SelectItem>
                                                </SelectContent>
                                            </Select>
                                        </div>
                                    </div>

                                    {/* 第二行：API 密钥 + 分组 */}
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                        <div className="space-y-2">
                                            <label className="text-sm font-medium text-card-foreground">{t('apiKey')}</label>
                                            <Select value={selectedApiKey} onValueChange={setSelectedApiKey}>
                                                <SelectTrigger className="rounded-xl w-full">
                                                    <SelectValue placeholder={t('apiKeyPlaceholder')} />
                                                </SelectTrigger>
                                                <SelectContent className="rounded-xl">
                                                    {(apiKeys ?? []).map((key) => (
                                                        <SelectItem key={key.id} className="rounded-xl" value={key.api_key}>
                                                            {key.name} <span className="text-muted-foreground text-xs">{key.api_key.slice(0, 16)}...</span>
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                        </div>
                                        <div className="space-y-2">
                                            <label className="text-sm font-medium text-card-foreground">{t('group')}</label>
                                            <Select value={selectedModel} onValueChange={setSelectedModel}>
                                                <SelectTrigger className="rounded-xl w-full">
                                                    <SelectValue placeholder={t('groupPlaceholder')} />
                                                </SelectTrigger>
                                                <SelectContent className="rounded-xl">
                                                    {groupOptions.map((opt) => (
                                                        <SelectItem key={opt.value} className="rounded-xl" value={opt.value}>
                                                            {opt.label}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                        </div>
                                    </div>

                                    {invalidGroupNames.length > 0 && (
                                        <div className="rounded-xl border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-xs text-amber-700 dark:text-amber-300">
                                            {t('groupNameRule')}: {invalidGroupNames.join(', ')}
                                        </div>
                                    )}

                                    {/* curl 代码 */}
                                    <div className="space-y-2">
                                        <div className="flex items-center justify-between">
                                            <label className="text-sm font-medium text-card-foreground">{t('curlCode')}</label>
                                            <Button
                                                type="button"
                                                size="sm"
                                                variant="ghost"
                                                onClick={handleCopy}
                                                className="h-7 px-2 gap-1 text-xs"
                                            >
                                                {copied ? (
                                                    <><Check className="h-3.5 w-3.5 text-green-500" />{t('copied')}</>
                                                ) : (
                                                    <><Copy className="h-3.5 w-3.5" />{t('copy')}</>
                                                )}
                                            </Button>
                                        </div>
                                        <pre className="rounded-xl bg-muted/50 border border-border p-4 text-xs font-mono text-card-foreground overflow-x-auto whitespace-pre-wrap break-all">
                                            {curlCode}
                                        </pre>
                                    </div>

                                    {/* 端点路径说明 */}
                                    <div className="rounded-xl border border-border bg-muted/10 p-4 space-y-2">
                                        <div className="text-sm font-medium text-card-foreground">{t('endpoints')}</div>
                                        <div className="space-y-1 text-xs text-muted-foreground font-mono">
                                            <div><span className="text-primary">POST</span> {baseUrl}/v1/chat/completions — {t('endpointOpenAIChat')}</div>
                                            <div><span className="text-primary">POST</span> {baseUrl}/v1/responses — {t('endpointOpenAIResponses')}</div>
                                            <div><span className="text-primary">POST</span> {baseUrl}/v1/messages — {t('endpointAnthropic')}</div>
                                            <div><span className="text-primary">POST</span> {baseUrl}/v1/embeddings — {t('endpointEmbeddings')}</div>
                                        </div>
                                    </div>
                                </>
                            )}

                            {contentTab === 'ccswitch' && (
                                <>
                                    <div data-testid="ccswitch-panel" className="rounded-2xl border border-border/70 bg-muted/20 px-4 py-3 text-xs leading-5 text-muted-foreground">
                                        {t('ccswitchIntro')}
                                    </div>

                                    <div data-testid="ccswitch-progress-card" className="rounded-2xl border border-border/70 bg-muted/20 px-4 py-3 space-y-3">
                                        <div className="flex items-start justify-between gap-3">
                                            <div className="space-y-1">
                                                <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                                                    <span>{t('ccswitchProgressTitle')}</span>
                                                    <HelpHint>{t('ccswitchProgressHint')}</HelpHint>
                                                </div>
                                                <p className="text-xs leading-5 text-muted-foreground">
                                                    {t('ccswitchProgressFootnote')}
                                                </p>
                                            </div>
                                        </div>

                                        <div data-testid="ccswitch-progress-grid" className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                                            {ccswitchProgress.map((step, index) => (
                                                <div
                                                    key={step.key}
                                                    data-testid={`ccswitch-progress-step-${step.key}`}
                                                    className={cn(
                                                        'flex items-center gap-2 rounded-xl border px-3 py-2 text-xs transition-colors',
                                                        step.done
                                                            ? 'border-primary/30 bg-primary/10 text-card-foreground'
                                                            : 'border-border/70 bg-background/60 text-muted-foreground'
                                                    )}
                                                >
                                                    {step.done ? (
                                                        <Check className="h-3.5 w-3.5 text-primary" />
                                                    ) : (
                                                        <span className="flex h-4 w-4 items-center justify-center rounded-full border border-current/60 text-[10px] font-medium">
                                                            {index + 1}
                                                        </span>
                                                    )}
                                                    <span className="font-medium">{step.label}</span>
                                                    <span className="ml-auto text-[11px] text-muted-foreground">
                                                        {step.done ? t('ccswitchStepDone') : t('ccswitchStepPending')}
                                                    </span>
                                                </div>
                                            ))}
                                        </div>
                                    </div>

                                    <div data-testid="ccswitch-client-section" className="space-y-2">
                                        <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                                            <label>{t('ccswitchStepClient')}</label>
                                            <HelpHint>{t('ccswitchClientHint')}</HelpHint>
                                        </div>
                                        <div data-testid="ccswitch-client-options" className="grid grid-cols-2 gap-2">
                                            {(['claude', 'codex'] as CCSwitchAppType[]).map((app) => (
                                                <Button
                                                    key={app}
                                                    type="button"
                                                    onClick={() => updateCCSwitch({ appType: app })}
                                                    data-testid={`ccswitch-client-${app}`}
                                                    variant={ccswitchForm.appType === app ? 'default' : 'outline'}
                                                    size="sm"
                                                    className={cn(
                                                        'h-auto min-h-11 rounded-xl px-3 py-2 text-center text-xs whitespace-normal capitalize sm:text-sm',
                                                        ccswitchForm.appType !== app && 'text-muted-foreground'
                                                    )}
                                                >
                                                    {app === 'claude' ? t('ccswitchAppClaude') : t('ccswitchAppCodex')}
                                                </Button>
                                            ))}
                                        </div>
                                    </div>

                                    <div data-testid="ccswitch-key-section" className="space-y-2">
                                        <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                                            <label>{t('ccswitchStepKey')}</label>
                                            <HelpHint>{t('ccswitchKeyHint')}</HelpHint>
                                        </div>
                                        <Select value={selectedApiKey} onValueChange={setSelectedApiKey}>
                                            <SelectTrigger data-testid="ccswitch-key-trigger" className="rounded-xl w-full">
                                                <SelectValue placeholder={t('apiKeyPlaceholder')} />
                                            </SelectTrigger>
                                            <SelectContent data-testid="ccswitch-key-content" className="rounded-xl">
                                                {(apiKeys ?? []).map((key) => (
                                                    <SelectItem key={key.id} className="rounded-xl" value={key.api_key}>
                                                        {key.name} <span className="text-muted-foreground text-xs">{key.api_key.slice(0, 16)}...</span>
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                    </div>

                                    <div data-testid="ccswitch-model-section" className="space-y-2">
                                        <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                                            <label>
                                                {t('ccswitchStepModel')} <span className="text-destructive">*</span>
                                            </label>
                                            <HelpHint>{t('ccswitchModelHint')}</HelpHint>
                                        </div>
                                        <Select value={ccswitchForm.model} onValueChange={v => updateCCSwitch({ model: v })} disabled={!ccswitchCanChooseModel}>
                                            <SelectTrigger data-testid="ccswitch-model-trigger" className="rounded-xl w-full">
                                                <SelectValue placeholder={t('groupPlaceholder')} />
                                            </SelectTrigger>
                                            <SelectContent data-testid="ccswitch-model-content" className="rounded-xl">
                                                {groupOptions.map((opt) => (
                                                    <SelectItem key={opt.value} className="rounded-xl" value={opt.value}>
                                                        {opt.value}
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                        {!ccswitchCanChooseModel ? (
                                            <div data-testid="ccswitch-model-locked-hint" className="rounded-xl border border-dashed border-border/70 bg-background/60 px-3 py-2 text-xs leading-5 text-muted-foreground">
                                                {ccswitchModelBlockedHint}
                                            </div>
                                        ) : null}
                                    </div>

                                    {ccswitchCanConfirmName ? (
                                        <div data-testid="ccswitch-name-section" className="space-y-2">
                                            <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                                                <label>
                                                    {t('ccswitchStepName')} <span className="text-destructive">*</span>
                                                </label>
                                                <HelpHint>{t('ccswitchNameHint')}</HelpHint>
                                            </div>
                                            <Input
                                                data-testid="ccswitch-name-input"
                                                value={ccswitchForm.name}
                                                onChange={e => {
                                                    setNameEdited(true);
                                                    updateCCSwitch({ name: e.target.value });
                                                }}
                                                placeholder={t('ccswitchNamePlaceholder')}
                                                className="rounded-xl"
                                            />
                                        </div>
                                    ) : (
                                        <div data-testid="ccswitch-name-locked-hint" className="rounded-2xl border border-dashed border-border/70 bg-background/60 px-4 py-3">
                                            <div className="text-sm font-medium text-card-foreground">{t('ccswitchStepName')}</div>
                                            <p className="mt-1 text-xs leading-5 text-muted-foreground">{t('ccswitchNameLockedHint')}</p>
                                        </div>
                                    )}

                                    {ccswitchForm.appType === 'claude' ? (
                                        ccswitchCanOpenAdvanced ? (
                                            <Accordion data-testid="ccswitch-advanced-section" type="single" collapsible className="rounded-2xl border border-border/70 bg-muted/20 px-4">
                                                <AccordionItem value="ccswitch-advanced" className="border-none">
                                                    <AccordionTrigger
                                                        data-testid="ccswitch-advanced-trigger"
                                                        className="py-3 text-sm font-medium hover:no-underline"
                                                        addon={<HelpHint className="mt-1 size-3.5">{t('ccswitchAdvancedHint')}</HelpHint>}
                                                    >
                                                        <div>
                                                            <div>{t('ccswitchAdvanced')}</div>
                                                            <div className="mt-1 text-xs font-normal text-muted-foreground">
                                                                <span>{t('ccswitchAdvancedDesc')}</span>
                                                            </div>
                                                        </div>
                                                    </AccordionTrigger>
                                                    <AccordionContent className="pb-3">
                                                        <div data-testid="ccswitch-advanced-grid" className="grid grid-cols-1 gap-3 md:grid-cols-3">
                                                            {(
                                                                [
                                                                    { field: 'haikuModel' as const, labelKey: 'ccswitchHaikuModel' as const },
                                                                    { field: 'sonnetModel' as const, labelKey: 'ccswitchSonnetModel' as const },
                                                                    { field: 'opusModel' as const, labelKey: 'ccswitchOpusModel' as const },
                                                                ]
                                                            ).map(({ field, labelKey }) => (
                                                                <div key={field} className="space-y-2">
                                                                    <label className="text-sm font-medium text-card-foreground">{t(labelKey)}</label>
                                                                    <Select
                                                                        value={ccswitchForm[field]}
                                                                        onValueChange={v => updateCCSwitch({ [field]: v })}
                                                                    >
                                                                        <SelectTrigger data-testid={`ccswitch-advanced-trigger-${field}`} className="rounded-xl w-full">
                                                                            <SelectValue placeholder={t('groupPlaceholder')} />
                                                                        </SelectTrigger>
                                                                        <SelectContent data-testid={`ccswitch-advanced-content-${field}`} className="rounded-xl">
                                                                            {groupOptions.map((opt) => (
                                                                                <SelectItem key={opt.value} className="rounded-xl" value={opt.value}>
                                                                                    {opt.value}
                                                                                </SelectItem>
                                                                            ))}
                                                                        </SelectContent>
                                                                    </Select>
                                                                </div>
                                                            ))}
                                                        </div>
                                                    </AccordionContent>
                                                </AccordionItem>
                                            </Accordion>
                                        ) : (
                                            <div data-testid="ccswitch-advanced-locked-hint" className="rounded-2xl border border-dashed border-border/70 bg-background/60 px-4 py-3">
                                                <div className="flex items-center gap-2 text-sm font-medium text-card-foreground">
                                                    <span>{t('ccswitchAdvanced')}</span>
                                                    <HelpHint>{t('ccswitchAdvancedHint')}</HelpHint>
                                                </div>
                                                <p className="mt-1 text-xs leading-5 text-muted-foreground">{t('ccswitchAdvancedLockedHint')}</p>
                                            </div>
                                        )
                                    ) : null}

                                    <Button
                                        type="button"
                                        onClick={handleCCSwitchImport}
                                        data-testid="ccswitch-import-button"
                                        disabled={!ccswitchReady}
                                        className="h-auto min-h-11 w-full gap-2 rounded-xl px-4 py-3 text-sm whitespace-normal"
                                    >
                                        <ExternalLink className="h-4 w-4" />
                                        {t('ccswitchImport')}
                                    </Button>
                                    {!ccswitchReady && ccswitchImportBlockedHint ? (
                                        <div data-testid="ccswitch-import-locked-hint" className="rounded-2xl border border-dashed border-border/70 bg-background/60 px-4 py-3 text-xs leading-5 text-muted-foreground">
                                            {ccswitchImportBlockedHint}
                                        </div>
                                    ) : null}
                                </>
                            )}
                        </div>
                    </motion.div>
                </>
            )}
        </AnimatePresence>
    );
}

