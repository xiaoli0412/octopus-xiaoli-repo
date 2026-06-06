'use client';

import { useMemo } from 'react';
import { useModelList, useUpstreamPriceSummaries, type UpstreamPriceSummary } from '@/api/endpoints/model';
import { ModelItem } from './Item';
import { useSearchStore, useToolbarViewOptionsStore } from '@/components/modules/toolbar';
import { VirtualizedGrid } from '@/components/common/VirtualizedGrid';

export function Model() {
    const { data: models } = useModelList();
    const { data: upstreamPrices } = useUpstreamPriceSummaries();
    const pageKey = 'model' as const;
    const searchTerm = useSearchStore((s) => s.getSearchTerm(pageKey));
    const modelDensity = useToolbarViewOptionsStore((s) => s.modelDensity);
    const sortOrder = useToolbarViewOptionsStore((s) => s.getSortOrder(pageKey));
    const filter = useToolbarViewOptionsStore((s) => s.modelFilter);

    const sortedModels = useMemo(() => {
        if (!models) return [];
        return [...models].sort((a, b) =>
            sortOrder === 'asc' ? a.name.localeCompare(b.name) : b.name.localeCompare(a.name)
        );
    }, [models, sortOrder]);

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

        if (filter === 'priced') {
            return bySearch.filter(hasPricing);
        }
        if (filter === 'free') {
            return bySearch.filter((m) => !hasPricing(m));
        }

        return bySearch;
    }, [sortedModels, searchTerm, filter]);

    const estimateItemHeight = modelDensity === 'compact' ? 126 : 156;
    const upstreamByModel = useMemo(() => {
        const map = new Map<string, UpstreamPriceSummary>();
        (upstreamPrices ?? []).forEach((item) => {
            map.set(item.model_name.toLowerCase(), item);
        });
        return map;
    }, [upstreamPrices]);

    return (
        <div data-testid="model-page" data-layout="grid" data-density={modelDensity} className="h-full min-h-0">
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
