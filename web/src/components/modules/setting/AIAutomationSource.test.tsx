'use client';

import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { SettingAIAutomationSource } from './AIAutomationSource';
import type { AIAutomationProfileRef } from '@/api/endpoints/ai-automation';

type AIAutomationConfigContract = {
    requested_config_source_mode: 'manual' | 'ai_profile';
    config_source_mode: 'manual' | 'ai_profile';
    requested_active_ai_profile_id: number;
    active_ai_profile_id: number;
    requested_active_ai_profile?: AIAutomationProfileRef;
    active_ai_profile?: AIAutomationProfileRef;
    source_fallback_reason: string;
};

const mocks = vi.hoisted(() => ({
    settings: [
        { key: 'config_source_mode', value: 'manual' },
        { key: 'active_ai_profile_id', value: '11' },
    ],
    profiles: [
        {
            id: 11,
            domain: 'grouping',
            name: 'Routing Plan Alpha',
            version: 2,
            status: 'active',
            confidence: 0.87,
            explanation: 'Keeps manual config intact while switching runtime reads.',
            created_at: '2026-04-27T08:00:00Z',
            updated_at: '2026-04-27T09:00:00Z',
        },
        {
            id: 12,
            domain: 'price_recognition',
            name: 'Cost Review Beta',
            version: 1,
            status: 'ready',
            confidence: 0.61,
            explanation: 'Adds pricing hints without overwriting manual pricing data.',
            created_at: '2026-04-27T10:00:00Z',
            updated_at: '2026-04-27T11:00:00Z',
        },
    ],
    config: {
        requested_config_source_mode: 'manual' as 'manual' | 'ai_profile',
        config_source_mode: 'manual' as 'manual' | 'ai_profile',
        requested_active_ai_profile_id: 11,
        active_ai_profile_id: 11,
        requested_active_ai_profile: {
            id: 11,
            domain: 'grouping',
            name: 'Routing Plan Alpha',
            version: 2,
            status: 'active',
            confidence: 0.87,
            explanation: 'Keeps manual config intact while switching runtime reads.',
            updated_at: '2026-04-27T09:00:00Z',
        },
        active_ai_profile: {
            id: 11,
            domain: 'grouping',
            name: 'Routing Plan Alpha',
            version: 2,
            status: 'active',
            confidence: 0.87,
            explanation: 'Keeps manual config intact while switching runtime reads.',
            updated_at: '2026-04-27T09:00:00Z',
        },
        source_fallback_reason: '',
    } as AIAutomationConfigContract,
    setSettingMutateMock: vi.fn(),
    activateProfileMutateAsyncMock: vi.fn(async () => undefined),
    setActiveItemMock: vi.fn(),
    toastSuccessMock: vi.fn(),
    toastInfoMock: vi.fn(),
    toastErrorMock: vi.fn(),
}));

vi.mock('next-intl', () => ({
    useTranslations: () => (key: string, values?: Record<string, unknown>) => values ? `${key}:${JSON.stringify(values)}` : key,
}));

vi.mock('@/components/ui/select', async () => {
    const React = await import('react');
    type SelectItemLikeProps = { children?: React.ReactNode; value?: string };
    const SelectContext = React.createContext<{
        items: Array<{ label: string; value: string }>;
        onValueChange?: (value: string) => void;
        value?: string;
    }>({ items: [], onValueChange: undefined, value: undefined });

    function collectItems(children: React.ReactNode, items: Array<{ label: string; value: string }> = []) {
        React.Children.forEach(children, (child) => {
            if (!React.isValidElement(child)) return;
            const element = child as React.ReactElement<SelectItemLikeProps>;
            if ((element.type as { __testIsSelectItem?: boolean }).__testIsSelectItem) {
                items.push({
                    label: typeof element.props.children === 'string' ? element.props.children : String(element.props.value),
                    value: element.props.value ?? '',
                });
                return;
            }
            if (element.props?.children) {
                collectItems(element.props.children, items);
            }
        });
        return items;
    }

    function Select({ children, onValueChange, value }: { children: React.ReactNode; onValueChange?: (value: string) => void; value?: string }) {
        const items = React.useMemo(() => collectItems(children), [children]);
        const contextValue = React.useMemo(() => ({ items, onValueChange, value }), [items, onValueChange, value]);
        return <SelectContext.Provider value={contextValue}>{children}</SelectContext.Provider>;
    }

    function SelectTrigger({ className, 'aria-label': ariaLabel }: { className?: string; 'aria-label'?: string }) {
        const context = React.useContext(SelectContext);
        return (
            <select aria-label={ariaLabel} className={className} role="combobox" onChange={(event) => context.onValueChange?.(event.target.value)} value={context.value}>
                {context.items.map((item) => (
                    <option key={item.value} value={item.value}>{item.label}</option>
                ))}
            </select>
        );
    }

    function SelectValue() {
        return null;
    }

    function SelectContent() {
        return null;
    }

    function SelectItem() {
        return null;
    }

    (SelectItem as { __testIsSelectItem?: boolean }).__testIsSelectItem = true;

    return { Select, SelectContent, SelectItem, SelectTrigger, SelectValue };
});

vi.mock('@/api/endpoints/setting', () => ({
    SettingKey: {
        ConfigSourceMode: 'config_source_mode',
        ActiveAIProfileID: 'active_ai_profile_id',
    },
    useSettingList: () => ({ data: mocks.settings }),
    useSetSetting: () => ({
        mutate: mocks.setSettingMutateMock,
    }),
}));

vi.mock('@/api/endpoints/ai-automation', () => ({
    useAIAutomationConfig: () => ({ data: mocks.config }),
    useAIProfiles: () => ({ data: mocks.profiles }),
    useActivateAIProfile: () => ({
        isPending: false,
        mutateAsync: mocks.activateProfileMutateAsyncMock,
    }),
}));

vi.mock('@/components/modules/navbar', () => ({
    useNavStore: (selector: (state: { setActiveItem: typeof mocks.setActiveItemMock }) => unknown) => selector({ setActiveItem: mocks.setActiveItemMock }),
}));

vi.mock('@/components/common/Toast', () => ({
    toast: {
        success: mocks.toastSuccessMock,
        info: mocks.toastInfoMock,
        error: mocks.toastErrorMock,
    },
}));

describe('SettingAIAutomationSource', () => {
    beforeEach(() => {
        mocks.settings = [
            { key: 'config_source_mode', value: 'manual' },
            { key: 'active_ai_profile_id', value: '11' },
        ];
        mocks.profiles = [
            {
                id: 11,
                domain: 'grouping',
                name: 'Routing Plan Alpha',
                version: 2,
                status: 'active',
                confidence: 0.87,
                explanation: 'Keeps manual config intact while switching runtime reads.',
                created_at: '2026-04-27T08:00:00Z',
                updated_at: '2026-04-27T09:00:00Z',
            },
            {
                id: 12,
                domain: 'price_recognition',
                name: 'Cost Review Beta',
                version: 1,
                status: 'ready',
                confidence: 0.61,
                explanation: 'Adds pricing hints without overwriting manual pricing data.',
                created_at: '2026-04-27T10:00:00Z',
                updated_at: '2026-04-27T11:00:00Z',
            },
        ];
        mocks.config = {
            requested_config_source_mode: 'manual',
            config_source_mode: 'manual',
            requested_active_ai_profile_id: 11,
            active_ai_profile_id: 11,
            requested_active_ai_profile: {
                id: 11,
                domain: 'grouping',
                name: 'Routing Plan Alpha',
                version: 2,
                status: 'active',
                confidence: 0.87,
                explanation: 'Keeps manual config intact while switching runtime reads.',
                updated_at: '2026-04-27T09:00:00Z',
            },
            active_ai_profile: {
                id: 11,
                domain: 'grouping',
                name: 'Routing Plan Alpha',
                version: 2,
                status: 'active',
                confidence: 0.87,
                explanation: 'Keeps manual config intact while switching runtime reads.',
                updated_at: '2026-04-27T09:00:00Z',
            },
            source_fallback_reason: '',
        };
        mocks.setSettingMutateMock.mockReset();
        mocks.activateProfileMutateAsyncMock.mockReset();
        mocks.activateProfileMutateAsyncMock.mockResolvedValue(undefined);
        mocks.setActiveItemMock.mockReset();
        mocks.toastSuccessMock.mockReset();
        mocks.toastInfoMock.mockReset();
        mocks.toastErrorMock.mockReset();
        mocks.setSettingMutateMock.mockImplementation((_payload: unknown, options?: { onSuccess?: () => void }) => {
            options?.onSuccess?.();
        });
    });

    it('renders selected profile metadata and safety hints', () => {
        render(<SettingAIAutomationSource />);

        expect(screen.getAllByText('Routing Plan Alpha').length).toBeGreaterThan(0);
        expect(screen.getByText('aiAutomationSource.selectedProfile')).toBeInTheDocument();
        expect(screen.getAllByText('aiAutomationSource.statusValues.active').length).toBeGreaterThan(0);
        expect(screen.getByText('aiAutomationSource.profileConfidence')).toBeInTheDocument();
        expect(screen.getByText('aiAutomationSource.readyHint')).toBeInTheDocument();
        expect(screen.getByText('aiAutomationSource.manualSafety')).toBeInTheDocument();
        expect(screen.getByText('aiAutomationSource.fallbackHint')).toBeInTheDocument();
        expect(screen.getByText('Keeps manual config intact while switching runtime reads.')).toBeInTheDocument();
    });

    it('activates the selected profile when switching to AI profile mode', async () => {
        render(<SettingAIAutomationSource />);

        fireEvent.change(screen.getByRole('combobox', { name: 'aiAutomationSource.profileLabel' }), { target: { value: '12' } });
        fireEvent.change(screen.getByRole('combobox', { name: 'aiAutomationSource.mode' }), { target: { value: 'ai_profile' } });

        await waitFor(() => {
            expect(mocks.activateProfileMutateAsyncMock).toHaveBeenCalledWith(12);
        });
    });

    it('falls back to a simple setting update when switching back to manual mode', async () => {
        mocks.settings = [
            { key: 'config_source_mode', value: 'ai_profile' },
            { key: 'active_ai_profile_id', value: '11' },
        ];

        render(<SettingAIAutomationSource />);

        fireEvent.change(screen.getByRole('combobox', { name: 'aiAutomationSource.mode' }), { target: { value: 'manual' } });

        await waitFor(() => {
            expect(mocks.setSettingMutateMock).toHaveBeenCalledWith(
                { key: 'config_source_mode', value: 'manual' },
                expect.objectContaining({ onSuccess: expect.any(Function) }),
            );
        });
    });

    it('shows runtime fallback notice when backend already resolved source back to manual', () => {
        mocks.settings = [
            { key: 'config_source_mode', value: 'ai_profile' },
            { key: 'active_ai_profile_id', value: '11' },
        ];
        mocks.config = {
            requested_config_source_mode: 'ai_profile',
            config_source_mode: 'manual',
            requested_active_ai_profile_id: 11,
            active_ai_profile_id: 0,
            requested_active_ai_profile: {
                id: 11,
                domain: 'grouping',
                name: 'Routing Plan Alpha',
                version: 2,
                status: 'invalid',
                confidence: 0.87,
                explanation: 'Keeps manual config intact while switching runtime reads.',
                updated_at: '2026-04-27T09:00:00Z',
            },
            active_ai_profile: undefined,
            source_fallback_reason: 'profile_missing',
        };

        render(<SettingAIAutomationSource />);

        expect(screen.getByText('aiAutomationSource.runtimeFallbackTitle')).toBeInTheDocument();
        expect(screen.getByText('aiAutomationSource.runtimeFallbackNotice')).toBeInTheDocument();
        expect(screen.getByText('aiAutomationSource.fallbackReasons.profile_missing')).toBeInTheDocument();
        expect(screen.getByRole('combobox', { name: 'aiAutomationSource.mode' })).toHaveValue('manual');
    });

    it('uses config profile summary when profile list is temporarily unavailable', () => {
        mocks.profiles = [];
        mocks.settings = [
            { key: 'config_source_mode', value: 'ai_profile' },
            { key: 'active_ai_profile_id', value: '11' },
        ];
        mocks.config = {
            requested_config_source_mode: 'ai_profile',
            config_source_mode: 'manual',
            requested_active_ai_profile_id: 11,
            active_ai_profile_id: 0,
            requested_active_ai_profile: {
                id: 11,
                domain: 'grouping',
                name: 'Routing Plan Alpha',
                version: 2,
                status: 'ready',
                confidence: 0.87,
                explanation: 'Keeps manual config intact while switching runtime reads.',
                updated_at: '2026-04-27T09:00:00Z',
            },
            active_ai_profile: undefined,
            source_fallback_reason: 'profile_invalid_content',
        };

        render(<SettingAIAutomationSource />);

        expect(screen.getByText('aiAutomationSource.noActiveProfile')).toBeInTheDocument();
        expect(screen.getByText('Routing Plan Alpha')).toBeInTheDocument();
        expect(screen.getByText('aiAutomationSource.profileEmpty')).toBeInTheDocument();
        expect(screen.getByText('aiAutomationSource.runtimeFallbackTitle')).toBeInTheDocument();
        expect(screen.getByText('aiAutomationSource.fallbackReasons.profile_invalid_content')).toBeInTheDocument();
    });

    it('blocks switching to invalid profiles and surfaces risk hints', async () => {
        mocks.profiles = [
            {
                id: 11,
                domain: 'grouping',
                name: 'Routing Plan Alpha',
                version: 2,
                status: 'active',
                confidence: 0.87,
                explanation: 'Keeps manual config intact while switching runtime reads.',
                created_at: '2026-04-27T08:00:00Z',
                updated_at: '2026-04-27T09:00:00Z',
            },
            {
                id: 12,
                domain: 'price_recognition',
                name: 'Broken Plan',
                version: 3,
                status: 'invalid',
                confidence: 0.41,
                explanation: 'This plan failed validation.',
                created_at: '2026-04-27T10:00:00Z',
                updated_at: '2026-04-27T11:00:00Z',
            },
        ];

        render(<SettingAIAutomationSource />);

        fireEvent.change(screen.getByRole('combobox', { name: 'aiAutomationSource.profileLabel' }), { target: { value: '12' } });

        expect(screen.getByText('aiAutomationSource.riskInvalid')).toBeInTheDocument();
        expect(screen.getByText('aiAutomationSource.riskLowConfidence:{"confidence":"41%"}')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'aiAutomationSource.useSelected' })).toBeDisabled();

        fireEvent.change(screen.getByRole('combobox', { name: 'aiAutomationSource.mode' }), { target: { value: 'ai_profile' } });

        await waitFor(() => {
            expect(mocks.toastInfoMock).toHaveBeenCalledWith('aiAutomationSource.profileBlocked');
        });
        expect(mocks.activateProfileMutateAsyncMock).not.toHaveBeenCalled();
    });
});
