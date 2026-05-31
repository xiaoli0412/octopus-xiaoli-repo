'use client';

import { useState } from 'react';
import { ArrowDownUp, LayoutGrid, List, Plus, Search, SlidersHorizontal, X } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';
import {
    MorphingDialog,
    MorphingDialogTrigger,
    MorphingDialogContainer,
    MorphingDialogContent,
} from '@/components/ui/morphing-dialog';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { buttonVariants } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { useNavStore, type NavItem } from '@/components/modules/navbar';
import { CreateDialogContent as ChannelCreateContent } from '@/components/modules/channel/Create';
import { CreateDialogContent as GroupCreateContent } from '@/components/modules/group/Create';
import { CreateDialogContent as ModelCreateContent } from '@/components/modules/model/Create';
import { useTranslations } from 'next-intl';
import { useSearchStore } from './search-store';
import {
    useToolbarViewOptionsStore,
    TOOLBAR_PAGES,
    type ToolbarPage,
    type CardDensity,
    type ChannelFilter,
    type ChannelProviderFilter,
    type GroupFilter,
    type ModelFilter,
} from './view-options-store';

const CHANNEL_FILTER_OPTIONS: ChannelFilter[] = ['all', 'enabled', 'disabled'];
const CHANNEL_PROVIDER_FILTER_OPTIONS: ChannelProviderFilter[] = ['all', 'openai', 'anthropic', 'gemini', 'volcengine', 'github-copilot', 'antigravity', 'zen'];
const GROUP_FILTER_OPTIONS: GroupFilter[] = ['all', 'with-members', 'empty'];
const MODEL_FILTER_OPTIONS: ModelFilter[] = ['all', 'priced', 'free'];

function isToolbarPage(item: NavItem): item is ToolbarPage {
    return (TOOLBAR_PAGES as readonly NavItem[]).includes(item);
}

function CreateDialogContent({ activeItem }: { activeItem: ToolbarPage }) {
    switch (activeItem) {
        case 'channel':
            return <ChannelCreateContent />;
        case 'group':
            return <GroupCreateContent />;
        case 'model':
            return <ModelCreateContent />;
    }
}

export function Toolbar() {
    const t = useTranslations('toolbar');
    const { activeItem } = useNavStore();
    const toolbarItem = isToolbarPage(activeItem) ? activeItem : null;
    const searchTerm = useSearchStore((s) => (toolbarItem ? s.searchTerms[toolbarItem] || '' : ''));
    const setSearchTerm = useSearchStore((s) => s.setSearchTerm);
    const layout = useToolbarViewOptionsStore((s) => (toolbarItem ? s.getLayout(toolbarItem) : 'grid'));
    const channelDensity = useToolbarViewOptionsStore((s) => s.channelDensity);
    const modelDensity = useToolbarViewOptionsStore((s) => s.modelDensity);
    const sortOrder = useToolbarViewOptionsStore((s) => (toolbarItem ? s.getSortOrder(toolbarItem) : 'asc'));
    const setLayout = useToolbarViewOptionsStore((s) => s.setLayout);
    const setChannelDensity = useToolbarViewOptionsStore((s) => s.setChannelDensity);
    const setModelDensity = useToolbarViewOptionsStore((s) => s.setModelDensity);
    const setSortOrder = useToolbarViewOptionsStore((s) => s.setSortOrder);
    const channelFilter = useToolbarViewOptionsStore((s) => s.channelFilter);
    const channelProviderFilter = useToolbarViewOptionsStore((s) => s.channelProviderFilter);
    const channelModelKeyword = useToolbarViewOptionsStore((s) => s.channelModelKeyword);
    const channelKeyKeyword = useToolbarViewOptionsStore((s) => s.channelKeyKeyword);
    const groupFilter = useToolbarViewOptionsStore((s) => s.groupFilter);
    const modelFilter = useToolbarViewOptionsStore((s) => s.modelFilter);
    const setChannelFilter = useToolbarViewOptionsStore((s) => s.setChannelFilter);
    const setChannelProviderFilter = useToolbarViewOptionsStore((s) => s.setChannelProviderFilter);
    const setChannelModelKeyword = useToolbarViewOptionsStore((s) => s.setChannelModelKeyword);
    const setChannelKeyKeyword = useToolbarViewOptionsStore((s) => s.setChannelKeyKeyword);
    const clearChannelDetailFilters = useToolbarViewOptionsStore((s) => s.clearChannelDetailFilters);
    const setGroupFilter = useToolbarViewOptionsStore((s) => s.setGroupFilter);
    const setModelFilter = useToolbarViewOptionsStore((s) => s.setModelFilter);
    const [expandedSearchItem, setExpandedSearchItem] = useState<ToolbarPage | null>(null);
    const searchExpanded = expandedSearchItem === toolbarItem;

    if (!toolbarItem) return null;

    const channelFilterLabelKeys: Record<ChannelFilter, string> = {
        all: 'popover.filter.channel.all',
        enabled: 'popover.filter.channel.enabled',
        disabled: 'popover.filter.channel.disabled',
    };
    const channelProviderFilterLabelKeys: Record<ChannelProviderFilter, string> = {
        all: 'popover.filter.channelProvider.all',
        openai: 'popover.filter.channelProvider.openai',
        anthropic: 'popover.filter.channelProvider.anthropic',
        gemini: 'popover.filter.channelProvider.gemini',
        volcengine: 'popover.filter.channelProvider.volcengine',
        'github-copilot': 'popover.filter.channelProvider.githubCopilot',
        antigravity: 'popover.filter.channelProvider.antigravity',
        zen: 'popover.filter.channelProvider.zen',
    };
    const groupFilterLabelKeys: Record<GroupFilter, string> = {
        all: 'popover.filter.group.all',
        'with-members': 'popover.filter.group.withMembers',
        empty: 'popover.filter.group.empty',
    };
    const modelFilterLabelKeys: Record<ModelFilter, string> = {
        all: 'popover.filter.model.all',
        priced: 'popover.filter.model.priced',
        free: 'popover.filter.model.free',
    };

    const filterOptions = toolbarItem === 'channel'
        ? CHANNEL_FILTER_OPTIONS.map((value) => ({
            value,
            label: t(channelFilterLabelKeys[value]),
        }))
        : toolbarItem === 'group'
            ? GROUP_FILTER_OPTIONS.map((value) => ({
                value,
                label: t(groupFilterLabelKeys[value]),
            }))
            : MODEL_FILTER_OPTIONS.map((value) => ({
                value,
                label: t(modelFilterLabelKeys[value]),
            }));

    const activeFilter = toolbarItem === 'channel'
        ? channelFilter
        : toolbarItem === 'group'
            ? groupFilter
            : modelFilter;
    const channelProviderOptions = CHANNEL_PROVIDER_FILTER_OPTIONS.map((value) => ({
        value,
        label: t(channelProviderFilterLabelKeys[value]),
    }));
    const searchPlaceholderKey = toolbarItem === 'channel'
        ? 'searchPlaceholder.channel'
        : toolbarItem === 'group'
            ? 'searchPlaceholder.group'
            : 'searchPlaceholder.model';
    const searchInputWidthClass = toolbarItem === 'channel'
        ? 'w-28 sm:w-44 md:w-52'
        : toolbarItem === 'group'
            ? 'w-24 sm:w-36 md:w-40'
            : 'w-24 sm:w-36 md:w-40';
    const isDensityToolbar = toolbarItem === 'channel' || toolbarItem === 'model';
    const activeDensity: CardDensity = toolbarItem === 'channel' ? channelDensity : modelDensity;
    const layoutSectionLabel = isDensityToolbar ? t('popover.density') : t('popover.layout');
    const layoutLabels = isDensityToolbar
        ? { grid: t('popover.normal'), list: t('popover.compact') }
        : { grid: t('popover.grid'), list: t('popover.list') };
    const densityButtonMap: Record<'grid' | 'list', CardDensity> = {
        grid: 'normal',
        list: 'compact',
    };

    const handleLayoutOptionChange = (value: 'grid' | 'list') => {
        if (toolbarItem === 'channel') {
            setChannelDensity(densityButtonMap[value]);
            return;
        }

        if (toolbarItem === 'model') {
            setModelDensity(densityButtonMap[value]);
            return;
        }

        setLayout(toolbarItem, value);
    };

    const isLayoutOptionActive = (value: 'grid' | 'list') => {
        if (isDensityToolbar) {
            return activeDensity === densityButtonMap[value];
        }

        return layout === value;
    };

    const handleFilterChange = (value: string) => {
        switch (toolbarItem) {
            case 'channel':
                setChannelFilter(value as ChannelFilter);
                break;
            case 'group':
                setGroupFilter(value as GroupFilter);
                break;
            case 'model':
                setModelFilter(value as ModelFilter);
                break;
        }
    };

    return (
        <AnimatePresence mode="wait">
            <motion.div
                key="toolbar"
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.9 }}
                transition={{ duration: 0.2 }}
                className="flex items-center gap-1.5 sm:gap-2"
            >
                {/* 搜索按钮/展开框 */}
                <div className="relative h-9 w-9 shrink-0">
                    {!searchExpanded ? (
                        <motion.button
                            layoutId="search-box"
                            onClick={() => setExpandedSearchItem(toolbarItem)}
                            className={buttonVariants({ variant: "ghost", size: "icon", className: "absolute inset-0 rounded-xl transition-none hover:bg-transparent text-muted-foreground hover:text-foreground focus-visible:ring-2 focus-visible:ring-primary/20" })}
                        >
                            <motion.span layout="position"><Search className="size-4 transition-colors duration-300" /></motion.span>
                        </motion.button>
                    ) : (
                        <motion.div
                            layoutId="search-box"
                            data-slot="toolbar-search-expanded"
                            data-page={toolbarItem}
                            className="absolute right-0 top-0 flex h-9 max-w-[calc(100vw-8.5rem)] items-center gap-2 rounded-xl border border-border/70 bg-background/95 px-3 shadow-sm sm:max-w-none"
                            transition={{ type: 'spring', stiffness: 400, damping: 30 }}
                        >
                            <motion.span layout="position"><Search className="size-4 text-muted-foreground shrink-0" /></motion.span>
                            <input
                                type="text"
                                value={searchTerm}
                                onChange={(e) => setSearchTerm(toolbarItem, e.target.value)}
                                placeholder={t(searchPlaceholderKey)}
                                autoFocus
                                className={cn('min-w-0 bg-transparent text-sm outline-none placeholder:text-muted-foreground', searchInputWidthClass)}
                            />
                            <button
                                onClick={() => {
                                    setSearchTerm(toolbarItem, '');
                                    setExpandedSearchItem(null);
                                }}
                                className="rounded p-0.5 shrink-0 text-muted-foreground transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-primary/20"
                            >
                                <X className="size-3.5" />
                            </button>
                        </motion.div>
                    )}
                </div>

                <Popover>
                    <PopoverTrigger asChild>
                        <button
                            type="button"
                            data-testid={`toolbar-view-options-trigger-${toolbarItem}`}
                            aria-label={t('popover.ariaLabel')}
                            className={buttonVariants({
                                variant: 'ghost',
                                size: 'icon',
                                className: 'rounded-xl transition-none hover:bg-transparent text-muted-foreground hover:text-foreground focus-visible:ring-2 focus-visible:ring-primary/20',
                            })}
                        >
                            <SlidersHorizontal className="size-4 transition-colors duration-300" />
                        </button>
                    </PopoverTrigger>
                    <PopoverContent
                        data-testid={`toolbar-view-options-content-${toolbarItem}`}
                        align="center"
                        side="bottom"
                        sideOffset={8}
                        className="w-[min(21rem,calc(100vw-1rem))] rounded-2xl border border-border/60 bg-card p-3 shadow-xl sm:w-72"
                    >
                        <div className="grid gap-2.5">
                            <div className="grid gap-2">
                                <p className="text-xs font-medium text-muted-foreground">{layoutSectionLabel}</p>
                                <div className="grid grid-cols-2 gap-2">
                                    <button
                                        type="button"
                                        data-testid={`toolbar-layout-grid-${toolbarItem}`}
                                        onClick={() => handleLayoutOptionChange('grid')}
                                        className={cn(
                                            'inline-flex h-8 items-center justify-center gap-1.5 rounded-lg border text-xs font-medium transition-colors focus-visible:ring-2 focus-visible:ring-primary/20',
                                            isLayoutOptionActive('grid')
                                                ? 'border-primary/30 bg-primary text-primary-foreground'
                                                : 'border-border bg-muted/20 text-foreground hover:bg-muted/30'
                                        )}
                                    >
                                        <LayoutGrid className="size-3.5" />
                                        {layoutLabels.grid}
                                    </button>
                                    <button
                                        type="button"
                                        data-testid={`toolbar-layout-list-${toolbarItem}`}
                                        onClick={() => handleLayoutOptionChange('list')}
                                        className={cn(
                                            'inline-flex h-8 items-center justify-center gap-1.5 rounded-lg border text-xs font-medium transition-colors focus-visible:ring-2 focus-visible:ring-primary/20',
                                            isLayoutOptionActive('list')
                                                ? 'border-primary/30 bg-primary text-primary-foreground'
                                                : 'border-border bg-muted/20 text-foreground hover:bg-muted/30'
                                        )}
                                    >
                                        <List className="size-3.5" />
                                        {layoutLabels.list}
                                    </button>
                                </div>
                            </div>

                            <div className="grid gap-2">
                                <p className="text-xs font-medium text-muted-foreground">{t('popover.sort')}</p>
                                <div className="grid grid-cols-2 gap-2">
                                    <button
                                        type="button"
                                        onClick={() => setSortOrder(toolbarItem, 'asc')}
                                        className={cn(
                                            'inline-flex h-8 items-center justify-center gap-1.5 rounded-lg border text-xs font-medium transition-colors focus-visible:ring-2 focus-visible:ring-primary/20',
                                            sortOrder === 'asc'
                                                ? 'border-primary/30 bg-primary text-primary-foreground'
                                                : 'border-border bg-muted/20 text-foreground hover:bg-muted/30'
                                        )}
                                    >
                                        <ArrowDownUp className="size-3.5" />
                                        {t('popover.asc')}
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => setSortOrder(toolbarItem, 'desc')}
                                        className={cn(
                                            'inline-flex h-8 items-center justify-center gap-1.5 rounded-lg border text-xs font-medium transition-colors focus-visible:ring-2 focus-visible:ring-primary/20',
                                            sortOrder === 'desc'
                                                ? 'border-primary/30 bg-primary text-primary-foreground'
                                                : 'border-border bg-muted/20 text-foreground hover:bg-muted/30'
                                        )}
                                    >
                                        <ArrowDownUp className="size-3.5 rotate-180" />
                                        {t('popover.desc')}
                                    </button>
                                </div>
                            </div>

                            <div className="grid gap-2">
                                <p className="text-xs font-medium text-muted-foreground">{t('popover.filter.title')}</p>
                                <div className="grid gap-2">
                                    {filterOptions.map((option) => (
                                        <button
                                            key={option.value}
                                            type="button"
                                            data-testid={toolbarItem === 'channel' ? `toolbar-channel-filter-${option.value}` : undefined}
                                            onClick={() => handleFilterChange(option.value)}
                                            className={cn(
                                                'h-8 rounded-lg border px-2 text-left text-xs font-medium transition-colors focus-visible:ring-2 focus-visible:ring-primary/20',
                                                activeFilter === option.value
                                                    ? 'border-primary/30 bg-primary text-primary-foreground'
                                                    : 'border-border bg-muted/20 text-foreground hover:bg-muted/30'
                                            )}
                                        >
                                            {option.label}
                                        </button>
                                    ))}
                                </div>
                            </div>

                            {toolbarItem === 'channel' && (
                                <div className="grid gap-2.5 border-t border-border/60 pt-2.5">
                                    <div className="grid gap-2">
                                        <p className="text-xs font-medium text-muted-foreground">{t('popover.filter.channelProviderTitle')}</p>
                                        <div className="grid grid-cols-2 gap-2">
                                            {channelProviderOptions.map((option) => (
                                                <button
                                                    key={option.value}
                                                    type="button"
                                                    data-testid={`toolbar-channel-provider-${option.value}`}
                                                    onClick={() => setChannelProviderFilter(option.value)}
                                                    className={cn(
                                                        'h-8 rounded-lg border px-2 text-left text-[11px] font-medium transition-colors focus-visible:ring-2 focus-visible:ring-primary/20 sm:text-xs',
                                                        channelProviderFilter === option.value
                                                            ? 'border-primary/30 bg-primary text-primary-foreground'
                                                            : 'border-border bg-muted/20 text-foreground hover:bg-muted/30'
                                                    )}
                                                >
                                                    <span className="block truncate">{option.label}</span>
                                                </button>
                                            ))}
                                        </div>
                                    </div>

                                    <div className="grid gap-2">
                                        <p className="text-xs font-medium text-muted-foreground">{t('popover.filter.channelModelTitle')}</p>
                                        <input
                                            type="text"
                                            data-testid="toolbar-channel-model-keyword"
                                            value={channelModelKeyword}
                                            onChange={(e) => setChannelModelKeyword(e.target.value)}
                                            placeholder={t('popover.filter.channelModelPlaceholder')}
                                            className="h-9 rounded-lg border border-border bg-background px-3 text-xs text-foreground outline-none placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-primary/20"
                                        />
                                    </div>

                                    <div className="grid gap-2">
                                        <p className="text-xs font-medium text-muted-foreground">{t('popover.filter.channelKeyTitle')}</p>
                                        <input
                                            type="text"
                                            data-testid="toolbar-channel-key-keyword"
                                            value={channelKeyKeyword}
                                            onChange={(e) => setChannelKeyKeyword(e.target.value)}
                                            placeholder={t('popover.filter.channelKeyPlaceholder')}
                                            className="h-9 rounded-lg border border-border bg-background px-3 text-xs text-foreground outline-none placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-primary/20"
                                        />
                                    </div>

                                    <button
                                        type="button"
                                        data-testid="toolbar-channel-clear-detail-filters"
                                        onClick={() => clearChannelDetailFilters()}
                                        className="h-8 rounded-lg border border-border bg-muted/20 px-2 text-xs font-medium text-foreground transition-colors hover:bg-muted/30 focus-visible:ring-2 focus-visible:ring-primary/20"
                                    >
                                        {t('popover.filter.channelClearDetailFilters')}
                                    </button>
                                </div>
                            )}
                        </div>
                    </PopoverContent>
                </Popover>

                {/* 创建按钮 */}
                <MorphingDialog>
                    <MorphingDialogTrigger
                        data-slot="toolbar-create-trigger"
                        data-page={toolbarItem}
                        data-testid={`toolbar-create-trigger-${toolbarItem}`}
                        className={buttonVariants({ variant: "ghost", size: "icon", className: "rounded-xl transition-none hover:bg-transparent text-muted-foreground hover:text-foreground focus-visible:ring-2 focus-visible:ring-primary/20" })}
                    >
                        <Plus className="size-4 transition-colors duration-300" />
                    </MorphingDialogTrigger>

                    <MorphingDialogContainer>
                        <MorphingDialogContent className="w-fit max-w-full bg-card text-card-foreground px-6 py-4 rounded-3xl custom-shadow max-h-[calc(100vh-2rem)] flex flex-col overflow-hidden">
                            <CreateDialogContent activeItem={toolbarItem} />
                        </MorphingDialogContent>
                    </MorphingDialogContainer>
                </MorphingDialog>
            </motion.div>
        </AnimatePresence>
    );
}

export { useSearchStore } from './search-store';
export { useToolbarViewOptionsStore } from './view-options-store';
