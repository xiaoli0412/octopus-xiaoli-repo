'use client';

import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { SettingAIAutomationSource } from './AIAutomationSource';

const mocks = vi.hoisted(() => ({
	overview: {
		enabled: true,
		execution_source: { mode: 'manual', base_url: 'http://127.0.0.1:1088/v1', channel_type: 'openai-compatible', model: 'gpt-4o', use_local_default: true, label: 'Manual AI endpoint' },
		managed_group_name: 'AI Governance Managed',
		learning: { enabled: true, sample_count: 4, top_target: 'gpt-4o / ch#1 / key#2', last_sample_at: 1715100000, top_score: 0.91 },
		active_strategy_profile: { id: 11, name: 'Managed routing baseline', summary: 'Imported from current session', status: 'active', is_active: true, created_at: '2026-05-09T08:00:00Z', updated_at: '2026-05-09T08:02:00Z' },
		recent_session: { id: 91, goal: '整理当前路由', scope: 'routing_grouping', expert_preset_id: 'balanced', status: 'ready', current_stage: 'completed', operator_summary: 'Governance reviewed 3 channels and prepared 4 controlled mutations.', risk_summary: 'Explicit apply is required.', confidence: 0.88, mutation_count: 4, can_apply: true, created_at: '2026-05-09T08:00:00Z', updated_at: '2026-05-09T08:05:00Z' },
	},
	profiles: [
		{ id: 11, name: 'Managed routing baseline', summary: 'Imported from current session', status: 'active', is_active: true, created_at: '2026-05-09T08:00:00Z', updated_at: '2026-05-09T08:02:00Z' },
	],
	setActiveItemMock: vi.fn(),
}));

vi.mock('next-intl', () => ({
	useTranslations: () => (key: string) => key,
}));

vi.mock('@/api/endpoints/ai-automation', () => ({
	useAIGovernanceOverview: () => ({ data: mocks.overview }),
	useStrategyProfiles: () => ({ data: mocks.profiles }),
}));

vi.mock('@/components/modules/navbar', () => ({
	useNavStore: (selector: (state: { setActiveItem: typeof mocks.setActiveItemMock }) => unknown) => selector({ setActiveItem: mocks.setActiveItemMock }),
}));

vi.mock('@/components/ui/button', () => ({
	Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => <button {...props}>{children}</button>,
}));

describe('SettingAIAutomationSource V2', () => {
	beforeEach(() => {
		mocks.setActiveItemMock.mockReset();
	});

	it('renders governance summary instead of the old profile switcher', () => {
		render(<SettingAIAutomationSource />);

		expect(screen.getByText('aiAutomationSource.title')).toBeInTheDocument();
		expect(screen.getByText('手动配置')).toBeInTheDocument();
		expect(screen.getByText('样本 4')).toBeInTheDocument();
		expect(screen.getByText('可应用')).toBeInTheDocument();
		expect(screen.getByText('Managed routing baseline')).toBeInTheDocument();
		expect(screen.getByText('Governance reviewed 3 channels and prepared 4 controlled mutations.')).toBeInTheDocument();
		expect(screen.getByText('aiAutomationSource.manualSafety')).toBeInTheDocument();
		expect(screen.queryByRole('combobox')).not.toBeInTheDocument();
	});

	it('keeps a lightweight open-center action', () => {
		render(<SettingAIAutomationSource />);

		fireEvent.click(screen.getByRole('button', { name: 'aiAutomationSource.openCenter' }));
		expect(mocks.setActiveItemMock).toHaveBeenCalledWith('ai');
	});
});
