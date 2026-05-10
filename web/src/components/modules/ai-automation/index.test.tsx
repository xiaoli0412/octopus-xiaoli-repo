import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { AIAutomation } from './index';

const state = vi.hoisted(() => ({
	overview: {
		enabled: true,
		execution_source: { mode: 'manual', base_url: 'http://127.0.0.1:1088/v1', channel_type: 'openai-compatible', model: 'gpt-4o', use_local_default: true, label: 'Manual AI endpoint' },
		runtime_policy: { strategy: 'balanced_latency', dispatch_mode: 'single_best', max_parallel_runs: 2, double_review_enabled: false, fallback_to_deterministic: true, degraded_to_deterministic: false },
		managed_group_name: 'AI Governance Managed',
		learning: { enabled: true, sample_count: 6, top_target: 'gpt-4o / ch#1 / key#2', last_sample_at: 1715100000, top_score: 0.91 },
		recent_session: { id: 91, goal: '整理当前路由', scope: 'routing_grouping', expert_preset_id: 'balanced', status: 'ready', current_stage: 'completed', operator_summary: 'recent summary', risk_summary: 'recent risk', confidence: 0.88, mutation_count: 4, can_apply: true, created_at: '2026-05-09T08:00:00Z', updated_at: '2026-05-09T08:05:00Z' },
	},
	sessions: [
		{ id: 91, goal: '整理当前路由', scope: 'routing_grouping', expert_preset_id: 'balanced', status: 'ready', current_stage: 'completed', operator_summary: 'Normalize managed group', risk_summary: 'Explicit apply required', confidence: 0.88, mutation_count: 4, can_apply: true, created_at: '2026-05-09T08:00:00Z', updated_at: '2026-05-09T08:05:00Z' },
	],
	sessionDetail: {
		id: 91,
		goal: '整理当前路由',
		scope: 'routing_grouping',
		expert_preset_id: 'balanced',
		status: 'ready',
		current_stage: 'completed',
		operator_summary: 'Governance reviewed 3 channels and prepared 4 controlled mutations.',
		risk_summary: 'Explicit apply is required.',
		confidence: 0.88,
		mutation_count: 4,
		can_apply: true,
		created_at: '2026-05-09T08:00:00Z',
		updated_at: '2026-05-09T08:05:00Z',
		snapshot_checksum: 'checksum-91',
		snapshot_summary: { channels: 3, enabled_channels: 3, groups: 2, group_items: 5, route_target_overrides: 1, models: 2, active_source_mode: 'manual', active_source_label: 'Manual AI endpoint', highlights: ['3 channels / 3 enabled', '2 groups / 5 group items'] },
		plan: {
			findings: [
				{ severity: 'warning', title: 'Model gpt-4o is scattered', detail: 'Consolidate gpt-4o into the managed governance group.' },
			],
			decisions: [
				{ title: 'Curate managed routing group', summary: 'Refresh ordering and clean stale entries.' },
			],
			mutations: [
				{ type: 'group_upsert', summary: 'Create or update group AI Governance Managed', group_upsert: { group_name: 'AI Governance Managed', mode: 3 } },
				{ type: 'group_item_attach', summary: 'Attach gpt-4o from channel #1', group_item_attach: { group_name: 'AI Governance Managed', channel_id: 1, model_name: 'gpt-4o', priority: 1, weight: 1 } },
			],
			risk_summary: 'Explicit apply is required.',
			confidence: 0.88,
			operator_summary: 'Governance reviewed 3 channels and prepared 4 controlled mutations.',
		},
		preview: {
			headline: 'Governance reviewed 3 channels and prepared 4 controlled mutations.',
			summary_lines: ['1 findings / 1 decisions', '4 typed mutations staged'],
			impact_counts: { groups: 1, items: 3, overrides: 0, profiles: 0 },
			risk_notes: ['Apply is explicit and transactional.'],
			apply_blockers: [],
			can_apply: true,
			mutation_count: 2,
			mutations: [
				{ type: 'group_upsert', summary: 'Create or update group AI Governance Managed', group_upsert: { group_name: 'AI Governance Managed', mode: 3 } },
				{ type: 'group_item_attach', summary: 'Attach gpt-4o from channel #1', group_item_attach: { group_name: 'AI Governance Managed', channel_id: 1, model_name: 'gpt-4o', priority: 1, weight: 1 } },
			],
		},
		apply_runs: [
			{ id: 31, session_id: 91, status: 'succeeded', result_summary: 'Applied governance changes', created_at: '2026-05-09T08:10:00Z', updated_at: '2026-05-09T08:10:03Z', audit: { summary: 'Applied governance plan', items: [{ mutation_type: 'group_upsert', summary: 'Create or update group AI Governance Managed', status: 'succeeded', message: '' }] } },
		],
	},
	presets: [
		{ id: 'balanced', name: 'Balanced governance', description: 'Default routing and grouping governance.', review_depth: 'standard', create_managed_group: true, sync_bindings: true, cleanup_stale: true },
		{ id: 'deep_review', name: 'Deep review', description: 'Extra cleanup and drift visibility.', review_depth: 'deep', create_managed_group: true, sync_bindings: true, cleanup_stale: true },
	],
	profiles: [
		{ id: 1, name: 'Managed routing baseline', summary: 'Imported from session 91', status: 'ready', is_active: false, created_at: '2026-05-09T08:00:00Z', updated_at: '2026-05-09T08:00:00Z' },
	],
	rollbackPoints: [
		{ id: 501, session_id: 91, snapshot_checksum: 'rollback-checksum', summary: 'Rollback point before apply #31', created_at: '2026-05-09T08:09:00Z', updated_at: '2026-05-09T08:09:00Z' },
	],
	apiKeys: [
		{ id: 7, name: 'Primary API Key', api_key: 'sk-test', enabled: true },
	],
	learningSummary: { enabled: true, sample_count: 6, top_target: 'gpt-4o / ch#1 / key#2', last_sample_at: 1715100000, top_score: 0.91 },
	createSessionMutateAsync: vi.fn(),
	replanSessionMutateAsync: vi.fn(),
	applySessionMutateAsync: vi.fn(),
	rollbackSessionMutateAsync: vi.fn(),
	createProfileMutateAsync: vi.fn(),
	activateProfileMutateAsync: vi.fn(),
	updateRuntimePolicyMutateAsync: vi.fn(),
	setSettingMutateAsync: vi.fn(),
	toastSuccess: vi.fn(),
	toastError: vi.fn(),
}));

vi.mock('next-intl', () => ({
	useTranslations: (namespace?: string) => (key: string) => namespace ? `${namespace}.${key}` : key,
}));

vi.mock('@/components/common/PageWrapper', () => ({
	PageWrapper: ({ children, className, 'data-testid': testId }: { children: React.ReactNode; className?: string; 'data-testid'?: string }) => <div className={className} data-testid={testId}>{children}</div>,
}));

vi.mock('@/components/ui/button', () => ({
	Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => <button {...props}>{children}</button>,
}));

vi.mock('@/components/ui/input', () => ({
	Input: (props: React.InputHTMLAttributes<HTMLInputElement>) => <input {...props} />,
}));

vi.mock('@/components/ui/dialog', () => ({
	Dialog: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
	DialogContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
	DialogDescription: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
	DialogHeader: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
	DialogTitle: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
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

	function SelectTrigger({ className }: { className?: string }) {
		const context = React.useContext(SelectContext);
		return (
			<select className={className} role="combobox" onChange={(event) => context.onValueChange?.(event.target.value)} value={context.value}>
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

vi.mock('@/components/ui/switch', () => ({
	Switch: ({ checked, onCheckedChange }: { checked?: boolean; onCheckedChange?: (value: boolean) => void }) => (
		<button role="switch" aria-checked={checked} onClick={() => onCheckedChange?.(!checked)} type="button" />
	),
}));

vi.mock('@/components/common/AnimatedNumber', () => ({
	AnimatedNumber: ({ value }: { value: number }) => <>{value}</>,
}));

vi.mock('@/stores/setting', () => ({
	useSettingStore: (selector: (state: { locale: string }) => unknown) => selector({ locale: 'zh-Hans' }),
}));

vi.mock('@/lib/locale', () => ({
	formatDateTimeByLocale: (value: string) => value,
}));

vi.mock('@/components/common/Toast', () => ({
	toast: {
		success: state.toastSuccess,
		error: state.toastError,
	},
}));

vi.mock('@/api/endpoints/ai-automation', () => ({
	useAIGovernanceOverview: () => ({ data: state.overview }),
	useGovernanceSessions: () => ({ data: state.sessions }),
	useGovernanceSession: () => ({ data: state.sessionDetail }),
	useGovernanceRollbackPoints: () => ({ data: state.rollbackPoints }),
	useGovernanceRuntimePolicy: () => ({ data: state.overview.runtime_policy }),
	useExpertPresets: () => ({ data: state.presets }),
	useStrategyProfiles: () => ({ data: state.profiles }),
	useAIGovernanceLearningSummary: () => ({ data: state.learningSummary }),
	useCreateGovernanceSession: () => ({ isPending: false, mutateAsync: state.createSessionMutateAsync }),
	useReplanGovernanceSession: () => ({ isPending: false, mutateAsync: state.replanSessionMutateAsync }),
	useApplyGovernanceSession: () => ({ isPending: false, mutateAsync: state.applySessionMutateAsync }),
	useRollbackGovernanceSession: () => ({ isPending: false, mutateAsync: state.rollbackSessionMutateAsync }),
	useCreateStrategyProfile: () => ({ mutateAsync: state.createProfileMutateAsync }),
	useActivateStrategyProfile: () => ({ mutateAsync: state.activateProfileMutateAsync }),
	useUpdateGovernanceRuntimePolicy: () => ({ isPending: false, mutateAsync: state.updateRuntimePolicyMutateAsync }),
}));

vi.mock('@/api/endpoints/apikey', () => ({
	useAPIKeyList: () => ({ data: state.apiKeys }),
}));

vi.mock('@/api/endpoints/setting', () => ({
	SettingKey: {
		AIAutomationBaseUrl: 'ai_automation_base_url',
		AIAutomationUseLocalDefault: 'ai_automation_use_local_default',
		AIAutomationAPIKey: 'ai_automation_api_key',
		AIAutomationModel: 'ai_automation_model',
	},
	useSetSetting: () => ({ isPending: false, mutateAsync: state.setSettingMutateAsync }),
	useSettingList: () => ({ data: [] }),
}));

describe('AIAutomation V2', () => {
	beforeEach(() => {
		window.sessionStorage.clear();
		state.createSessionMutateAsync.mockReset();
		state.replanSessionMutateAsync.mockReset();
		state.applySessionMutateAsync.mockReset();
		state.rollbackSessionMutateAsync.mockReset();
		state.createProfileMutateAsync.mockReset();
		state.activateProfileMutateAsync.mockReset();
		state.updateRuntimePolicyMutateAsync.mockReset();
		state.setSettingMutateAsync.mockReset();
		state.toastSuccess.mockReset();
		state.toastError.mockReset();
		state.createSessionMutateAsync.mockResolvedValue({ ...state.sessionDetail, id: 92 });
		state.replanSessionMutateAsync.mockResolvedValue(state.sessionDetail);
		state.applySessionMutateAsync.mockResolvedValue({ ...state.sessionDetail, status: 'applied' });
		state.rollbackSessionMutateAsync.mockResolvedValue({ success: true });
		state.createProfileMutateAsync.mockResolvedValue({ id: 2, name: 'Strategy 91' });
		state.activateProfileMutateAsync.mockResolvedValue({ id: 1, name: 'Managed routing baseline' });
		state.updateRuntimePolicyMutateAsync.mockResolvedValue(state.overview.runtime_policy);
		state.setSettingMutateAsync.mockResolvedValue({ ok: true });
	});

	it('renders the V2 single-goal governance workbench', () => {
		render(<AIAutomation />);

		expect(screen.getByTestId('ai-automation-page')).toBeInTheDocument();
		expect(screen.getByText('aiAutomationV2.hero.title')).toBeInTheDocument();
		expect(screen.getByText('aiAutomationV2.main.goalTitle')).toBeInTheDocument();
		expect(screen.getByText('aiAutomationV2.summary.currentGoal')).toBeInTheDocument();
		expect(screen.getByText('aiAutomationV2.summary.activePlan')).toBeInTheDocument();
		expect(screen.getByText('aiAutomationV2.summary.lastApply')).toBeInTheDocument();
		expect(screen.getByText('aiAutomationV2.sidebar.runtimePolicy')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'aiAutomationV2.workspace.preview' })).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'aiAutomationV2.workspace.history' })).toBeInTheDocument();
		expect(screen.queryByText('aiAutomation.hero.title')).not.toBeInTheDocument();
	});

	it('creates a governance session from the one-line goal input', async () => {
		render(<AIAutomation />);

		fireEvent.change(screen.getByPlaceholderText('aiAutomationV2.main.goalPlaceholder'), { target: { value: '请整理当前路由与分组' } });
		fireEvent.click(screen.getByRole('button', { name: 'aiAutomationV2.actions.create' }));

		await waitFor(() => {
			expect(state.createSessionMutateAsync).toHaveBeenCalledWith({ goal: '请整理当前路由与分组', expert_preset_id: 'balanced' });
		});
	});

	it('shows preview mutations and allows apply from the preview tab', async () => {
		render(<AIAutomation />);

		fireEvent.click(screen.getByRole('button', { name: 'aiAutomationV2.workspace.preview' }));
		expect(screen.getByTestId('ai-governance-workspace-preview')).toBeInTheDocument();
		expect(screen.getByText('Create or update group AI Governance Managed')).toBeInTheDocument();
		expect(screen.getByText('Attach gpt-4o from channel #1')).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: 'aiAutomationV2.actions.apply' }));

		await waitFor(() => {
			expect(state.applySessionMutateAsync).toHaveBeenCalledWith(91);
		});
	});

	it('creates and activates strategy profiles in the profiles tab', async () => {
		render(<AIAutomation />);

		fireEvent.click(screen.getByRole('button', { name: 'aiAutomationV2.workspace.profiles' }));
		fireEvent.change(screen.getByPlaceholderText('aiAutomationV2.profiles.namePlaceholder'), { target: { value: 'Managed routing baseline' } });
		fireEvent.click(screen.getByRole('button', { name: 'aiAutomationV2.profiles.create' }));

		await waitFor(() => {
			expect(state.createProfileMutateAsync).toHaveBeenCalledWith({ session_id: 91, name: 'Managed routing baseline' });
		});

		fireEvent.click(screen.getByRole('button', { name: 'aiAutomationV2.profiles.activate' }));
		await waitFor(() => {
			expect(state.activateProfileMutateAsync).toHaveBeenCalledWith(1);
		});
	});

	it('opens the history tab and consumes the learning focus target jump', async () => {
		window.sessionStorage.setItem('octopus-ai-automation-focus-target', 'learning');
		const scrollIntoView = vi.fn();
		Object.defineProperty(Element.prototype, 'scrollIntoView', { configurable: true, value: scrollIntoView });

		render(<AIAutomation />);

		await waitFor(() => {
			expect(screen.getByTestId('ai-governance-workspace-history')).toBeInTheDocument();
		});
		expect(scrollIntoView).toHaveBeenCalled();
		expect(screen.getByText('aiAutomationV2.history.learningTitle')).toBeInTheDocument();
	});

	it('saves automation settings with local default toggle and actual api key value', async () => {
		render(<AIAutomation />);

		fireEvent.click(screen.getByRole('button', { name: 'aiAutomationV2.settings.open' }));

		const combos = screen.getAllByRole('combobox');
		fireEvent.change(combos[0], { target: { value: 'sk-test' } });
		fireEvent.change(combos[1], { target: { value: 'gpt-4o' } });

		const switches = screen.getAllByRole('switch');
		fireEvent.click(switches[0]);

		fireEvent.click(screen.getByRole('button', { name: 'aiAutomationV2.settings.save' }));

		await waitFor(() => {
			expect(state.setSettingMutateAsync).toHaveBeenCalledWith({ key: 'ai_automation_use_local_default', value: 'false' });
			expect(state.setSettingMutateAsync).toHaveBeenCalledWith({ key: 'ai_automation_api_key', value: 'sk-test' });
			expect(state.setSettingMutateAsync).toHaveBeenCalledWith({ key: 'ai_automation_model', value: 'gpt-4o' });
		});
	});
});
