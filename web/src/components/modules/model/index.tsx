'use client';

import { useMemo, useState } from 'react';
import { Activity, Play } from 'lucide-react';
import {
    useModelList,
    useModelChannelList,
    useUpstreamPriceSummaries,
    useTestModelConnectivity,
    type UpstreamPriceSummary,
    type ModelTestResult,
} from '@/api/endpoints/model';
import { ModelItem } from './Item';
import { useSearchStore, useToolbarViewOptionsStore } from '@/components/modules/toolbar';
import { VirtualizedGrid } from '@/components/common/VirtualizedGrid';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { toast } from '@/components/common/Toast';
import { useTranslations } from 'next-intl';

export function Model() {
    const t = useTranslations('model');
    const { data: models } = useModelList();
    const { data: channelModels } = useModelChannelList();
    const { data: upstreamPrices } = useUpstreamPriceSummaries();
    const pageKey = 'model' as const;
    const searchTerm = useSearchStore((s) => s.getSearchTerm(pageKey));
    const modelDensity = useToolbarViewOptionsStore((s) => s.modelDensity);
    const sortOrder = useToolbarViewOptionsStore((s) => s.getSortOrder(pageKey));
    const sortBy = useToolbarViewOptionsStore((s) => s.modelSortBy);
    const filter = useToolbarViewOptionsStore((s) => s.modelFilter);
    const minPrice = useToolbarViewOptionsStore((s) => s.modelMinPrice);
    const maxPrice = useToolbarViewOptionsStore((s) => s.modelMaxPrice);
    const hasUpstreamPrice = useToolbarViewOptionsStore((s) => s.modelHasUpstreamPrice);

    const [dialogOpen, setDialogOpen] = useState(false);
    const [selectedModel, setSelectedModel] = useState('');
    const [selectedChannel, setSelectedChannel] = useState('');
    const [testResult, setTestResult] = useState<ModelTestResult | null>(null);
    const testModel = useTestModelConnectivity();

    const upstreamByModel = useMemo(() => {
        const map = new Map<string, UpstreamPriceSummary>();
        (upstreamPrices ?? []).forEach((item) => {
            map.set(item.model_name.toLowerCase(), item);
        });
        return map;
    }, [upstreamPrices]);

    const sortedModels = useMemo(() => {
        if (!models) return [];
        const arr = [...models];
        const getGatewayInput = (m: (typeof models)[number]) =>
            upstreamByModel.get(m.name.toLowerCase())?.effective_gateway?.input ?? m.input;
        const getGatewayOutput = (m: (typeof models)[number]) =>
            upstreamByModel.get(m.name.toLowerCase())?.effective_gateway?.output ?? m.output;

        switch (sortBy) {
            case 'input':
                arr.sort((a, b) =>
                    sortOrder === 'asc'
                        ? getGatewayInput(a) - getGatewayInput(b)
                        : getGatewayInput(b) - getGatewayInput(a)
                );
                break;
            case 'output':
                arr.sort((a, b) =>
                    sortOrder === 'asc'
                        ? getGatewayOutput(a) - getGatewayOutput(b)
                        : getGatewayOutput(b) - getGatewayOutput(a)
                );
                break;
            default:
                arr.sort((a, b) =>
                    sortOrder === 'asc' ? a.name.localeCompare(b.name) : b.name.localeCompare(a.name)
                );
        }
        return arr;
    }, [models, sortBy, sortOrder, upstreamByModel]);

    const visibleModels = useMemo(() => {
        const term = searchTerm.toLowerCase().trim();
        const bySearch = !term
            ? sortedModels
            : sortedModels.filter((m) => {
                const canonicalName = (m.canonical_name ?? '').toLowerCase();
                return m.name.toLowerCase().includes(term) || canonicalName.includes(term);
            });
        const hasPricing = (model: (typeof bySearch)[number]) =>
            model.input + model.output + model.cache_read + model.cache_write > 0;

        const min = parseFloat(minPrice);
        const max = parseFloat(maxPrice);
        const hasRange = !Number.isNaN(min) || !Number.isNaN(max);

        return bySearch.filter((m) => {
            if (filter === 'priced' && !hasPricing(m)) return false;
            if (filter === 'free' && hasPricing(m)) return false;

            if (hasUpstreamPrice) {
                const summary = upstreamByModel.get(m.name.toLowerCase());
                if (!summary || summary.gateway_prices.length === 0) return false;
            }

            if (hasRange) {
                const price = upstreamByModel.get(m.name.toLowerCase())?.effective_gateway?.input ?? m.input;
                if (!Number.isNaN(min) && price < min) return false;
                if (!Number.isNaN(max) && price > max) return false;
            }

            return true;
        });
    }, [sortedModels, searchTerm, filter, minPrice, maxPrice, hasUpstreamPrice, upstreamByModel]);

    const estimateItemHeight = modelDensity === 'compact' ? 126 : 156;

    const channelsForSelectedModel = useMemo(() => {
        if (!selectedModel) return [];
        return (channelModels ?? []).filter(
            (item) => item.name.toLowerCase() === selectedModel.toLowerCase() && item.enabled
        );
    }, [channelModels, selectedModel]);

    const handleRunTest = () => {
        const channelId = parseInt(selectedChannel, 10);
        if (!selectedModel || !channelId) {
            toast.error(t('test.invalidSelection'));
            return;
        }
        setTestResult(null);
        testModel.mutate(
            { channel_id: channelId, model: selectedModel },
            {
                onSuccess: (data) => {
                    setTestResult(data);
                    if (data.success) {
                        toast.success(t('test.success'));
                    } else {
                        toast.error(t('test.failed'), { description: data.error_message });
                    }
                },
                onError: (error) => {
                    toast.error(t('test.failed'), { description: error.message });
                },
            }
        );
    };

    const handleModelChange = (value: string) => {
        setSelectedModel(value);
        setSelectedChannel('');
        setTestResult(null);
    };

    return (
        <div data-testid="model-page" data-layout="grid" data-density={modelDensity} className="h-full min-h-0">
            <div className="mb-3 flex items-center justify-end gap-2 px-1">
                <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
                    <DialogTrigger asChild>
                        <Button variant="outline" size="sm" className="gap-1.5 rounded-xl">
                            <Activity className="size-4" />
                            {t('test.title')}
                        </Button>
                    </DialogTrigger>
                    <DialogContent className="sm:max-w-md">
                        <DialogHeader>
                            <DialogTitle>{t('test.title')}</DialogTitle>
                            <DialogDescription>{t('test.description')}</DialogDescription>
                        </DialogHeader>
                        <div className="grid gap-4 py-2">
                            <div className="grid gap-2">
                                <label className="text-xs font-medium text-muted-foreground">{t('test.modelLabel')}</label>
                                <Select value={selectedModel} onValueChange={handleModelChange}>
                                    <SelectTrigger className="w-full">
                                        <SelectValue placeholder={t('test.modelPlaceholder')} />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {visibleModels.map((model) => (
                                            <SelectItem key={model.name} value={model.name}>
                                                {model.name}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>

                            <div className="grid gap-2">
                                <label className="text-xs font-medium text-muted-foreground">{t('test.channelLabel')}</label>
                                <Select value={selectedChannel} onValueChange={setSelectedChannel} disabled={!selectedModel || channelsForSelectedModel.length === 0}>
                                    <SelectTrigger className="w-full">
                                        <SelectValue placeholder={t('test.channelPlaceholder')} />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {channelsForSelectedModel.map((item) => (
                                            <SelectItem key={item.channel_id} value={String(item.channel_id)}>
                                                {item.channel_name} (#{item.channel_id})
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>

                            <Button
                                onClick={handleRunTest}
                                disabled={!selectedModel || !selectedChannel || testModel.isPending}
                                className="gap-1.5"
                            >
                                {testModel.isPending ? (
                                    <span className="size-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                                ) : (
                                    <Play className="size-4" />
                                )}
                                {t('test.run')}
                            </Button>

                            {testResult && (
                                <div className="grid gap-2 rounded-xl border border-border bg-muted/30 p-3">
                                    <div className="flex items-center gap-2">
                                        <span className="text-xs font-medium text-muted-foreground">{t('test.status')}:</span>
                                        <Badge
                                            variant={testResult.success ? 'default' : 'destructive'}
                                            className={cn(
                                                'text-xs',
                                                testResult.success && 'bg-green-600 hover:bg-green-700'
                                            )}
                                        >
                                            {testResult.success ? t('test.passed') : t('test.failed')}
                                        </Badge>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <span className="text-xs font-medium text-muted-foreground">{t('test.latency')}:</span>
                                        <span className="text-xs tabular-nums">{testResult.latency_ms} ms</span>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <span className="text-xs font-medium text-muted-foreground">{t('test.priceMatch')}:</span>
                                        <Badge variant={testResult.price_match ? 'default' : 'secondary'} className="text-xs">
                                            {testResult.price_match ? t('test.matched') : t('test.unmatched')}
                                        </Badge>
                                    </div>
                                    {!testResult.success && testResult.error_message && (
                                        <div className="grid gap-1">
                                            <span className="text-xs font-medium text-destructive">{t('test.error')}:</span>
                                            <p className="max-h-32 overflow-auto rounded-lg border border-destructive/20 bg-destructive/5 p-2 text-xs text-destructive">
                                                {testResult.error_message}
                                            </p>
                                        </div>
                                    )}
                                    {testResult.success && testResult.response_text && (
                                        <div className="grid gap-1">
                                            <span className="text-xs font-medium text-muted-foreground">{t('test.response')}:</span>
                                            <p className="max-h-32 overflow-auto rounded-lg border border-border bg-background p-2 text-xs">
                                                {testResult.response_text}
                                            </p>
                                        </div>
                                    )}
                                </div>
                            )}
                        </div>
                    </DialogContent>
                </Dialog>
            </div>

            <VirtualizedGrid
                items={visibleModels}
                layout="grid"
                columns={{ default: 1, md: 2, lg: 3, xl: 3, '2xl': 3 }}
                estimateItemHeight={estimateItemHeight}
                getItemKey={(model) => `model-${model.name}`}
                gap={modelDensity === 'compact' ? 10 : 12}
                renderItem={(model, index) => <ModelItem model={model} upstreamPrice={upstreamByModel.get(model.name.toLowerCase())} density={modelDensity} index={index} />}
            />
        </div>
    );
}
