import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { SettingDynamicRouting } from './DynamicRouting';
import { AI_AUTOMATION_FOCUS_TARGET_STORAGE_KEY } from '../ai-automation/focus-target';

const { mutateMock, resetLearningMutateMock, toastSuccessMock, dynamicSummaryState, setActiveItemMock } = vi.hoisted(() => ({
	mutateMock: vi.fn(),
	resetLearningMutateMock: vi.fn(),
	toastSuccessMock: vi.fn(),
	setActiveItemMock: vi.fn(),
		dynamicSummaryState: {
		data: {
			last_run_at: '2026-04-21T03:10:00Z',
			last_success_at: '2026-04-21T03:10:00Z',
			last_status: 'skipped',
			last_message: 'dynamic routing health disabled; summary scan skipped',
			current_mode: 'hybrid',
			effective_mode: 'strict-mechanism',
			decision: 'deterministic',
			decision_reason: 'summary_snapshot_health_disabled',
			health_enabled: false,
			channel_count: 4,
			enabled_channels: 3,
			group_count: 2,
			failover_groups: 1,
			free_public_keys: 5,
			paid_metered_keys: 2,
			private_internal_keys: 1,
			unknown_keys: 0,
			basis: 'daily_summary_scan_no_runtime_mutation',
		},
		isLoading: false,
		isError: false,
	},
}));

vi.mock('next-intl', () => ({
	useTranslations: () => (key: string) => key,
}));

vi.mock('@/api/endpoints/setting', () => ({
	SettingKey: {
		DynamicRoutingMode: 'dynamic_routing_mode',
		DynamicRoutingHealthEnabled: 'dynamic_routing_health_enabled',
		DynamicRoutingLearningEnabled: 'dynamic_routing_learning_enabled',
		RaceGlobalBudget: 'race_global_budget',
		RaceGroupBudget: 'race_group_budget',
		RaceChannelBudget: 'race_channel_budget',
		RaceKeyBudget: 'race_key_budget',
		RaceProbeBudget: 'race_probe_budget',
	},
	useSettingList: () => ({
		data: [
			{ key: 'dynamic_routing_mode', value: 'hybrid' },
			{ key: 'dynamic_routing_health_enabled', value: 'true' },
			{ key: 'dynamic_routing_learning_enabled', value: 'true' },
			{ key: 'race_global_budget', value: '64' },
			{ key: 'race_group_budget', value: '8' },
			{ key: 'race_channel_budget', value: '4' },
			{ key: 'race_key_budget', value: '2' },
			{ key: 'race_probe_budget', value: '16' },
		],
	}),
	useSetSetting: () => ({
		mutate: mutateMock,
	}),
}));

vi.mock('@/components/modules/navbar', () => ({
	useNavStore: (selector: (state: { setActiveItem: typeof setActiveItemMock }) => unknown) =>
		selector({ setActiveItem: setActiveItemMock }),
}));

vi.mock('@/api/endpoints/stats', () => ({
	useStatsDynamicRoutingSummary: () => dynamicSummaryState,
}));

vi.mock('@/api/endpoints/ai-automation', () => ({
	useDynamicRouteLearning: () => ({
		data: {
			enabled: true,
			states: [
				{
					id: 1,
					channel_id: 7,
					channel_key_id: 3,
					model_name: 'gpt-free',
					success_count: 10,
					failure_count: 1,
					fallback_count: 2,
					race_winner_count: 4,
					latency_ms_ewma: 120,
					score: 0.91,
					confidence: 0.84,
					last_sample_at: 1713669000,
				},
				{
					id: 2,
					channel_id: 8,
					channel_key_id: 4,
					model_name: 'gpt-safe',
					success_count: 4,
					failure_count: 0,
					fallback_count: 0,
					race_winner_count: 1,
					latency_ms_ewma: 98,
					score: 0.62,
					confidence: 0.55,
					last_sample_at: 1713600000,
				},
			],
		},
		refetch: vi.fn(),
	}),
	useResetDynamicRouteLearning: () => ({
		isPending: false,
		mutate: resetLearningMutateMock,
	}),
}));

vi.mock('@/components/common/Toast', () => ({
	toast: {
		success: toastSuccessMock,
	},
}));

describe('SettingDynamicRouting', () => {
	beforeEach(() => {
		window.sessionStorage.clear();
		Object.defineProperty(Element.prototype, 'scrollIntoView', {
			configurable: true,
			value: vi.fn(),
		});
		mutateMock.mockReset();
		resetLearningMutateMock.mockReset();
		toastSuccessMock.mockReset();
		setActiveItemMock.mockReset();
		dynamicSummaryState.data = {
			last_run_at: '2026-04-21T03:10:00Z',
			last_success_at: '2026-04-21T03:10:00Z',
			last_status: 'skipped',
			last_message: 'dynamic routing health disabled; summary scan skipped',
			current_mode: 'hybrid',
			effective_mode: 'strict-mechanism',
			decision: 'deterministic',
			decision_reason: 'summary_snapshot_health_disabled',
			health_enabled: false,
			channel_count: 4,
			enabled_channels: 3,
			group_count: 2,
			failover_groups: 1,
			free_public_keys: 5,
			paid_metered_keys: 2,
			private_internal_keys: 1,
			unknown_keys: 0,
			basis: 'daily_summary_scan_no_runtime_mutation',
		};
		dynamicSummaryState.isLoading = false;
		dynamicSummaryState.isError = false;
		resetLearningMutateMock.mockImplementation((_payload: unknown, options?: { onSuccess?: () => void }) => {
			options?.onSuccess?.();
		});
		mutateMock.mockImplementation((_payload: unknown, options?: { onSuccess?: () => void }) => {
			options?.onSuccess?.();
		});
	});

	it('renders the daily summary snapshot', () => {
		render(<SettingDynamicRouting />);

		expect(screen.getByText('dynamicRouting.defaultPathTitle')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.learning.title')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.learning.scopeValue')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.learning.samplesLabel')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.learning.runtimeLabel')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.learning.runtimeEnabled')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.learning.latestSampleLabel')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.learning.topTargetLabel')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.learning.summaryTitle')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.learning.summaryEnabledWithSamples')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.learning.topTargetValue')).toBeInTheDocument();
		expect(screen.queryByText('dynamicRouting.defaultPathDesc')).not.toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.budgetSummaryTitle')).toBeInTheDocument();
		expect(screen.queryByText('dynamicRouting.healthEnabledDesc')).not.toBeInTheDocument();
		expect(screen.queryByText('dynamicRouting.budgetSummaryDesc')).not.toBeInTheDocument();
		expect(screen.getByText('64')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.summary.title')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.summary.statusValues.skipped')).toBeInTheDocument();
		expect(screen.getByText('3/4')).toBeInTheDocument();
		expect(screen.getByText('1/2')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.summary.detailsTitle')).toBeInTheDocument();
		expect(screen.queryByText('dynamicRouting.modeOptions.strict-mechanism')).not.toBeInTheDocument();
		expect(screen.queryByText((text) => text.includes('dynamicRouting.summary.messageValues.health_disabled_scan_skipped'))).not.toBeInTheDocument();
		expect(screen.queryByText('dynamic routing health disabled; summary scan skipped')).not.toBeInTheDocument();
	});

	it('saves the dynamic routing learning switch and keeps a lightweight AI center entry', async () => {
		render(<SettingDynamicRouting />);

		fireEvent.click(screen.getByRole('switch', { name: 'dynamicRouting.learning.switchTitle' }));

		await waitFor(() => {
			expect(mutateMock).toHaveBeenCalledWith(
				{ key: 'dynamic_routing_learning_enabled', value: 'false' },
				expect.objectContaining({ onSuccess: expect.any(Function) }),
			);
		});
		expect(screen.getByTestId('setting-dynamic-routing-learning-runtime')).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: 'dynamicRouting.learning.open' }));
		expect(setActiveItemMock).toHaveBeenCalledWith('ai');
		expect(window.sessionStorage.getItem(AI_AUTOMATION_FOCUS_TARGET_STORAGE_KEY)).toBe('learning');
	});

	it('resets learning state from the lightweight settings card', async () => {
		render(<SettingDynamicRouting />);

		fireEvent.click(screen.getByRole('button', { name: 'dynamicRouting.learning.reset' }));

		await waitFor(() => {
			expect(resetLearningMutateMock).toHaveBeenCalledWith(undefined, expect.objectContaining({ onSuccess: expect.any(Function) }));
		});
		expect(toastSuccessMock).toHaveBeenCalledWith('dynamicRouting.learning.resetSuccess');
	});

	it('keeps mode, key mix, and raw summary details behind the summary accordion by default', async () => {
		render(<SettingDynamicRouting />);

		expect(screen.queryByText('dynamicRouting.summary.modeSnapshotTitle')).not.toBeInTheDocument();
		expect(screen.queryByText('dynamicRouting.summary.keyMixTitle')).not.toBeInTheDocument();
		expect(screen.queryByText((text) => text.includes('dynamicRouting.summary.messageValues.health_disabled_scan_skipped'))).not.toBeInTheDocument();

		fireEvent.click(screen.getByText('dynamicRouting.summary.detailsTitle'));

		expect(await screen.findByText('dynamicRouting.summary.modeSnapshotTitle')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.summary.keyMixTitle')).toBeInTheDocument();
		expect(screen.getAllByText('dynamicRouting.modeOptions.hybrid').length).toBeGreaterThan(0);
		expect(screen.getByText('dynamicRouting.modeOptions.strict-mechanism')).toBeInTheDocument();
		expect(screen.getAllByText('dynamicRouting.summary.decisionValues.deterministic').length).toBeGreaterThan(0);
		expect(screen.getByText((text) => text.includes('dynamicRouting.summary.decisionReasonValues.summary_snapshot_health_disabled'))).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.summary.basisLabel')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.summary.basisValues.daily_summary_scan_no_runtime_mutation')).toBeInTheDocument();
		expect(screen.getByText((text) => text.includes('dynamicRouting.summary.messageValues.health_disabled_scan_skipped'))).toBeInTheDocument();
	});

	it('keeps unmapped backend summary messages out of the primary localized line', () => {
		dynamicSummaryState.data = {
			last_run_at: '2026-04-21T03:10:00Z',
			last_success_at: '2026-04-21T03:10:00Z',
			last_status: 'ok',
			last_message: 'backend emitted a brand new summary message',
			current_mode: 'shadow-ai',
			effective_mode: 'shadow-ai',
			decision: 'shadow',
			decision_reason: 'summary_snapshot_runtime_modes',
			health_enabled: true,
			channel_count: 4,
			enabled_channels: 4,
			group_count: 2,
			failover_groups: 1,
			free_public_keys: 5,
			paid_metered_keys: 2,
			private_internal_keys: 1,
			unknown_keys: 0,
			basis: 'daily_summary_scan_no_runtime_mutation',
		};

		render(<SettingDynamicRouting />);

		fireEvent.click(screen.getByText('dynamicRouting.summary.detailsTitle'));

		expect(screen.getByText('dynamicRouting.summary.messageValues.unmapped')).toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.summary.messageDetailLabel')).toBeInTheDocument();
		expect(screen.getByText('backend emitted a brand new summary message')).toBeInTheDocument();
		expect(screen.queryByText('dynamicRouting.summary.messageLabel: backend emitted a brand new summary message')).not.toBeInTheDocument();
	});

	it('keeps budget inputs behind the advanced accordion by default', () => {
		render(<SettingDynamicRouting />);

		expect(screen.queryByDisplayValue('64')).not.toBeInTheDocument();
		expect(screen.getByText('dynamicRouting.advancedTitle')).toBeInTheDocument();
		expect(screen.queryByText('dynamicRouting.advancedDesc')).not.toBeInTheDocument();
	});

	it('saves changed routing mode', async () => {
		render(<SettingDynamicRouting />);

		fireEvent.click(screen.getByRole('combobox', { name: 'dynamicRouting.mode' }));
		fireEvent.click(await screen.findByText('dynamicRouting.modeOptions.incident-safe'));

		await waitFor(() => {
			expect(mutateMock).toHaveBeenCalledWith(
				{ key: 'dynamic_routing_mode', value: 'incident-safe' },
				expect.objectContaining({ onSuccess: expect.any(Function) }),
			);
		});
		expect(toastSuccessMock).toHaveBeenCalledWith('saved');
	});

	it('saves changed numeric budgets on blur', async () => {
		render(<SettingDynamicRouting />);

		fireEvent.click(screen.getByText('dynamicRouting.advancedTitle'));

		const input = await screen.findByDisplayValue('64');
		fireEvent.change(input, { target: { value: '48' } });
		fireEvent.blur(input);

		await waitFor(() => {
			expect(mutateMock).toHaveBeenCalledWith(
				{ key: 'race_global_budget', value: '48' },
				expect.objectContaining({ onSuccess: expect.any(Function) }),
			);
		});
		expect(toastSuccessMock).toHaveBeenCalledWith('saved');
	});
});
