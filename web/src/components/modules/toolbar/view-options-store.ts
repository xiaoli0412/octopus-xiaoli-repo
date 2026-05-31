import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';

export type ToolbarLayout = 'grid' | 'list';
export type ToolbarSortOrder = 'asc' | 'desc';
export type CardDensity = 'normal' | 'compact';
export type ChannelCardDensity = CardDensity;
export type ModelCardDensity = CardDensity;
export const TOOLBAR_PAGES = ['channel', 'group', 'model'] as const;
export type ToolbarPage = (typeof TOOLBAR_PAGES)[number];
export type ChannelFilter = 'all' | 'enabled' | 'disabled';
export type ChannelProviderFilter = 'all' | 'openai' | 'anthropic' | 'gemini' | 'volcengine' | 'github-copilot' | 'antigravity' | 'zen';
export type GroupFilter = 'all' | 'with-members' | 'empty';
export type ModelFilter = 'all' | 'priced' | 'free';

interface ToolbarViewOptionsState {
    layouts: Partial<Record<ToolbarPage, ToolbarLayout>>;
    sortOrders: Partial<Record<ToolbarPage, ToolbarSortOrder>>;
    channelDensity: ChannelCardDensity;
    modelDensity: ModelCardDensity;
    channelFilter: ChannelFilter;
    channelProviderFilter: ChannelProviderFilter;
    channelModelKeyword: string;
    channelKeyKeyword: string;
    groupFilter: GroupFilter;
    modelFilter: ModelFilter;

    getLayout: (item: ToolbarPage) => ToolbarLayout;
    setLayout: (item: ToolbarPage, value: ToolbarLayout) => void;
    setChannelDensity: (value: ChannelCardDensity) => void;
    setModelDensity: (value: ModelCardDensity) => void;

    getSortOrder: (item: ToolbarPage) => ToolbarSortOrder;
    setSortOrder: (item: ToolbarPage, value: ToolbarSortOrder) => void;

    setChannelFilter: (value: ChannelFilter) => void;
    setChannelProviderFilter: (value: ChannelProviderFilter) => void;
    setChannelModelKeyword: (value: string) => void;
    setChannelKeyKeyword: (value: string) => void;
    clearChannelDetailFilters: () => void;
    setGroupFilter: (value: GroupFilter) => void;
    setModelFilter: (value: ModelFilter) => void;
}

export const useToolbarViewOptionsStore = create<ToolbarViewOptionsState>()(
    persist(
        (set, get) => ({
            layouts: {},
            sortOrders: {},
            channelDensity: 'normal',
            modelDensity: 'compact',
            channelFilter: 'all',
            channelProviderFilter: 'all',
            channelModelKeyword: '',
            channelKeyKeyword: '',
            groupFilter: 'all',
            modelFilter: 'all',

            getLayout: (item) => get().layouts[item] || (item === 'model' ? 'list' : 'grid'),
            setLayout: (item, value) => {
                set((state) => ({ layouts: { ...state.layouts, [item]: value } }));
            },
            setChannelDensity: (value) => set({ channelDensity: value }),
            setModelDensity: (value) => set({ modelDensity: value }),

            getSortOrder: (item) => get().sortOrders[item] || 'asc',
            setSortOrder: (item, value) => {
                set((state) => ({ sortOrders: { ...state.sortOrders, [item]: value } }));
            },

            setChannelFilter: (value) => set({ channelFilter: value }),
            setChannelProviderFilter: (value) => set({ channelProviderFilter: value }),
            setChannelModelKeyword: (value) => set({ channelModelKeyword: value }),
            setChannelKeyKeyword: (value) => set({ channelKeyKeyword: value }),
            clearChannelDetailFilters: () => set({
                channelProviderFilter: 'all',
                channelModelKeyword: '',
                channelKeyKeyword: '',
            }),
            setGroupFilter: (value) => set({ groupFilter: value }),
            setModelFilter: (value) => set({ modelFilter: value }),
        }),
        {
            name: 'octopus-toolbar-view-options',
            storage: createJSONStorage(() => localStorage),
            partialize: (state) => ({
                channelDensity: state.channelDensity,
                modelDensity: state.modelDensity,
            }),
        }
    )
);
