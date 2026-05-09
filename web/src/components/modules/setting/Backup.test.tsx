'use client';

import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import * as React from 'react';

import { SettingBackup } from './Backup';
import { useSettingStore } from '@/stores/setting';

const mocks = vi.hoisted(() => {
	const importDBState: { data: any } = { data: undefined };
	const previewRollbackState: { data: any } = { data: undefined };
	const importSnapshotsState: {
		data: any[];
		isLoading: boolean;
		isError: boolean;
		isFetching: boolean;
	} = {
		data: [],
		isLoading: false,
		isError: false,
		isFetching: false,
	};

	return {
		importDBState,
		previewRollbackState,
		importSnapshotsState,
		importMutateAsyncMock: vi.fn(),
		exportMutateAsyncMock: vi.fn(),
		previewRollbackMutateAsyncMock: vi.fn(),
		rollbackLatestImportSnapshotMutateAsyncMock: vi.fn(),
		rollbackImportSnapshotMutateAsyncMock: vi.fn(),
		importSnapshotsRefetchMock: vi.fn(async () => ({ data: importSnapshotsState.data })),
		toastSuccessMock: vi.fn(),
		toastErrorMock: vi.fn(),
	};
});

vi.mock('next-intl', () => ({
	useTranslations: () => (key: string) => key,
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

vi.mock('@/components/common/Toast', () => ({
	toast: {
		success: mocks.toastSuccessMock,
		error: mocks.toastErrorMock,
	},
}));

vi.mock('@/api/endpoints/setting', async () => {
	const React = await import('react');
	return {
		useExportDB: () => ({
			isPending: false,
			mutateAsync: async (payload: any = {}) => {
				mocks.exportMutateAsyncMock(payload);
				return { filename: 'ignored.json' };
			},
		}),
		useImportDB: () => {
			const [data, setData] = React.useState(mocks.importDBState.data);
			return {
				data,
				isPending: false,
				reset() {
					mocks.importDBState.data = undefined;
					setData(undefined);
				},
				mutateAsync: async (payload: any) => {
					const result = await mocks.importMutateAsyncMock(payload);
					mocks.importDBState.data = result;
					setData(result);
					return result;
				},
			};
		},
		useImportSnapshots: () => ({
			data: mocks.importSnapshotsState.data,
			isLoading: mocks.importSnapshotsState.isLoading,
			isError: mocks.importSnapshotsState.isError,
			isFetching: mocks.importSnapshotsState.isFetching,
			refetch: mocks.importSnapshotsRefetchMock,
		}),
		usePreviewRollbackImportSnapshot: () => {
			const [data, setData] = React.useState(mocks.previewRollbackState.data);
			return {
				data,
				isPending: false,
				reset() {
					mocks.previewRollbackState.data = undefined;
					setData(undefined);
				},
				mutateAsync: async (payload: any) => {
					const result = await mocks.previewRollbackMutateAsyncMock(payload);
					mocks.previewRollbackState.data = result;
					setData(result);
					return result;
				},
			};
		},
		useRollbackLatestImportSnapshot: () => ({
			isPending: false,
			mutateAsync: mocks.rollbackLatestImportSnapshotMutateAsyncMock,
		}),
		useRollbackImportSnapshot: () => ({
			isPending: false,
			mutateAsync: mocks.rollbackImportSnapshotMutateAsyncMock,
		}),
	};
});

function getFileInput(container: HTMLElement) {
	const input = container.querySelector('input[type="file"]');
	if (!(input instanceof HTMLInputElement)) {
		throw new Error('Backup file input not found');
	}
	return input;
}

function setLocale(locale: 'zh-Hans' | 'zh-Hant' | 'en' | 'ja') {
	useSettingStore.setState({ locale });
	localStorage.setItem('octopus-settings', JSON.stringify({ state: { locale }, version: 0 }));
}

function getModelMappingsTextarea(container: HTMLElement) {
	const textarea = container.querySelector('[data-testid="backup-model-mappings-textarea"]');
	if (!(textarea instanceof HTMLTextAreaElement)) {
		throw new Error('Model mapping textarea not found');
	}
	return textarea;
}

function setStructuredMappingRows(rows: Array<{ source: string; target: string }>) {
	rows.forEach((row, index) => {
		if (index > 0) {
			fireEvent.click(screen.getByTestId('backup-structured-mapping-add'));
		}
		fireEvent.change(screen.getByTestId(`backup-structured-mapping-source-${index}`), { target: { value: row.source } });
		fireEvent.change(screen.getByTestId(`backup-structured-mapping-target-${index}`), { target: { value: row.target } });
	});
}

function getSwitchForLabel(label: string) {
	const labelNode = screen.getByText(label);
	let row: HTMLElement | null = labelNode.parentElement;
	while (row) {
		if (row.querySelectorAll('[role="switch"]').length === 1) {
			return within(row).getByRole('switch');
		}
		row = row.parentElement;
	}
	throw new Error(`Switch row for ${label} not found`);
}

function getByRoleName(role: Parameters<typeof screen.getByRole>[0], names: Array<string | RegExp>) {
	for (const name of names) {
		const match = screen.queryByRole(role, { name });
		if (match) return match;
		const textMatches = screen.queryAllByText(name as never);
		for (const textMatch of textMatches) {
			const fallback = textMatch.closest(role === 'button' ? 'button,[role="button"]' : `[role="${role}"]`);
			if (fallback instanceof HTMLElement) return fallback;
		}
	}
	throw new Error(`Unable to find ${role} with names: ${names.map((item) => item.toString()).join(', ')}`);
}

function getApplyConfirmSwitch() {
	const marker = screen.getByText((text) => [
		'我已经检查上方风险提示，确认可以把这次导入应用到当前项目。',
		'I reviewed the risks above and want to apply this import to the current project.',
	].includes(text));
	let row: HTMLElement | null = marker.parentElement;
	while (row) {
		if (row.querySelectorAll('[role="switch"]').length === 1) {
			return within(row).getByRole('switch');
		}
		row = row.parentElement;
	}
	throw new Error('Apply confirmation switch not found');
}

function getApplyConfirmSwitchBySelector() {
	return screen.getByTestId('backup-apply-confirm-switch');
}

function getHelpHintButtons(container: HTMLElement = document.body) {
	return Array.from(container.querySelectorAll('button[data-help-hint-trigger="true"]'));
}

function clickAccordionByText(text: string) {
	const trigger = screen.getByText(text).closest('[data-slot="accordion-trigger"]');
	if (!(trigger instanceof HTMLElement)) {
		throw new Error(`Accordion trigger for ${text} not found`);
	}
	fireEvent.click(trigger);
}

async function selectImportMode(value: string) {
	const comboboxes = Array.from(document.querySelectorAll('[role="combobox"]'));
	for (const combobox of comboboxes) {
		if (!(combobox instanceof HTMLSelectElement)) continue;
		const optionValues = Array.from(combobox.options).map((option) => option.value);
		if (!['incremental', 'map', 'merge', 'replace', 'skip'].every((option) => optionValues.includes(option))) continue;
		fireEvent.change(combobox, { target: { value } });
		await waitFor(() => {
			expect(combobox.value).toBe(value);
		});
		return;
	}
	throw new Error('Import mode combobox not found');
}

describe('SettingBackup', () => {
	beforeEach(() => {
		localStorage.clear();
		setLocale('zh-Hans');
		mocks.importDBState.data = undefined;
		mocks.previewRollbackState.data = undefined;
		mocks.importSnapshotsState.data = [];
		mocks.importSnapshotsState.isLoading = false;
		mocks.importSnapshotsState.isError = false;
		mocks.importSnapshotsState.isFetching = false;
		mocks.importMutateAsyncMock.mockReset();
		mocks.exportMutateAsyncMock.mockReset();
		mocks.previewRollbackMutateAsyncMock.mockReset();
		mocks.rollbackLatestImportSnapshotMutateAsyncMock.mockReset();
		mocks.rollbackImportSnapshotMutateAsyncMock.mockReset();
		mocks.importSnapshotsRefetchMock.mockReset();
		mocks.toastSuccessMock.mockReset();
		mocks.toastErrorMock.mockReset();
	});

	it('exports full and redacted snapshots', async () => {
		setLocale('en');
		render(<SettingBackup />);
		expect(document.querySelector('[data-testid="backup-page"]')).toBeInTheDocument();
		expect(getHelpHintButtons().length).toBeGreaterThanOrEqual(8);

		fireEvent.click(getByRoleName('button', ['Download JSON']));
		await waitFor(() => {
			expect(mocks.exportMutateAsyncMock).toHaveBeenCalledTimes(1);
		});
		expect(mocks.exportMutateAsyncMock).toHaveBeenCalledWith(expect.objectContaining({
			include_secrets: true,
			include_logs: false,
			include_stats: false,
		}));
		expect(mocks.toastSuccessMock).toHaveBeenCalledWith('Export started');

		mocks.exportMutateAsyncMock.mockClear();
		mocks.toastSuccessMock.mockClear();
		fireEvent.click(getSwitchForLabel('Include plaintext credentials in the snapshot'));
		expect(screen.getByText('Redacted credentials')).toBeInTheDocument();

		fireEvent.click(getByRoleName('button', ['Download JSON']));
		await waitFor(() => {
			expect(mocks.exportMutateAsyncMock).toHaveBeenCalledTimes(1);
		});
		expect(mocks.exportMutateAsyncMock).toHaveBeenCalledWith(expect.objectContaining({
			include_secrets: false,
		}));
	});

	it('runs dry-run, requires confirmation, and applies the same import', async () => {
		const selectedFile = new File(['{"snapshot":true}'], 'snapshot.json', { type: 'application/json' });
		const expectedScopes = {
			routing: false,
			models: true,
			api_keys: true,
			settings: true,
			stats: true,
			logs: true,
		};

		mocks.importMutateAsyncMock.mockImplementation(async (payload: any) => {
			if (payload.dryRun) {
				return {
					rows_affected: { channels: 1 },
					preview_token: 'preview-token-1',
					dry_run: true,
					mode: payload.mode,
					warnings: ['legacy warning: review before apply'],
					compatibility: { conflicts: ['channel conflict'] },
				};
			}
			return {
				rows_affected: { channels: 1, groups: 1 },
				dry_run: false,
				mode: payload.mode,
				post_import_validation: {
					degraded_groups: ['group-a'],
					health_check: {
						summary: { targets: 1, passed: 1, failed: 0, skipped: 0, rate_limited: 0 },
						checks: [{ channel_id: 7, channel_name: 'Primary', model: 'gpt-4o', passed: true }],
					},
				},
			};
		});

		const { container } = render(<SettingBackup />);
		expect(document.querySelector('[data-testid="backup-page"]')).toBeInTheDocument();
		fireEvent.click(getSwitchForLabel('按范围导入'));
		fireEvent.click(getSwitchForLabel('路由配置'));
		fireEvent.change(getFileInput(container), { target: { files: [selectedFile] } });

		fireEvent.click(screen.getByTestId('backup-import-button'));
		await screen.findByTestId('backup-pending-apply-ready');
		expect(screen.getByTestId('backup-pending-apply-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-pending-apply-meta-grid')).toBeInTheDocument();
		expect(screen.getByTestId('backup-import-result-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-import-summary-grid')).toBeInTheDocument();
		expect(screen.getByTestId('backup-import-summary-rebinds')).toHaveTextContent('凭证重绑定：0');
		expect(screen.getByTestId('backup-import-summary-rebinds')).toHaveAttribute('data-raw-value', '0');
		expect(screen.getByTestId('backup-compatibility-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-overview')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-toggle')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-details')).toBeInTheDocument();
		expect(screen.getByTestId('backup-apply-confirm-switch')).toBeInTheDocument();
		fireEvent.click(screen.getByTestId('backup-compatibility-toggle'));
		expect(screen.getByTestId('backup-import-warnings')).toBeInTheDocument();
		expect(screen.getByTestId('backup-import-warnings-title')).toHaveTextContent('导入预警');
		expect(screen.getByTestId('backup-import-warnings-item-0')).toHaveTextContent('legacy warning: review before apply');
		expect(mocks.importMutateAsyncMock).toHaveBeenNthCalledWith(1, expect.objectContaining({
			file: selectedFile,
			dryRun: true,
			mode: 'incremental',
			importScopes: expectedScopes,
			previewToken: undefined,
		}));
		expect(screen.getByTestId('backup-pending-apply-meta-file')).toHaveAttribute('data-raw-value', 'snapshot.json');
		expect(screen.getByTestId('backup-pending-apply-meta-mapping-count')).toHaveAttribute('data-raw-value', '0');
		expect(screen.getByTestId('backup-pending-apply-meta-preview-token')).toHaveAttribute('data-raw-value', 'preview-token-1');

		const applyButton = screen.getByTestId('backup-apply-same-import-button');
		expect(applyButton).toBeDisabled();
		expect(screen.getByTestId('backup-apply-confirm-panel')).toBeInTheDocument();
		fireEvent.click(getApplyConfirmSwitchBySelector());
		await waitFor(() => {
			expect(applyButton).toBeEnabled();
		});

		fireEvent.click(applyButton);
		await screen.findByTestId('backup-post-import-validation-panel');
		expect(screen.getByTestId('backup-post-import-validation-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-post-import-validation-summary-grid')).toBeInTheDocument();
		expect(screen.getByTestId('backup-post-import-validation-summary-degraded-groups')).toHaveTextContent('降级分组：1');
		expect(screen.getByTestId('backup-post-import-validation-summary-empty-groups')).toHaveTextContent('空分组：0');
		expect(screen.getByTestId('backup-post-import-validation-summary-disabled-channels')).toHaveTextContent('已禁用渠道：0');
		expect(screen.getByTestId('backup-post-import-validation-summary-channels-without-keys')).toHaveTextContent('无密钥渠道：0');
		expect(screen.getByTestId('backup-post-import-validation-summary-stale-items-removed')).toHaveTextContent('已清理过期项：0');
		expect(screen.getByTestId('backup-post-import-validation-summary-route-warnings')).toHaveTextContent('路由预警：0');
		expect(screen.getByTestId('backup-post-import-validation-summary-price-rule-warnings')).toHaveTextContent('价格规则预警：0');
		expect(screen.getByTestId('backup-post-import-validation-summary-alias-mappings')).toHaveTextContent('别名映射：0');
		expect(screen.getByTestId('backup-post-import-validation-summary-alias-warnings')).toHaveTextContent('别名预警：0');
		expect(screen.getByTestId('backup-post-import-health-summary')).toBeInTheDocument();
		expect(screen.getByTestId('backup-post-import-health-summary-grid')).toBeInTheDocument();
		expect(screen.getByTestId('backup-post-import-health-summary-targets')).toHaveTextContent('健康检测目标：1');
		expect(screen.getByTestId('backup-post-import-health-summary-passed')).toHaveTextContent('通过数量：1');
		expect(mocks.importMutateAsyncMock).toHaveBeenNthCalledWith(2, expect.objectContaining({
			file: selectedFile,
			dryRun: false,
			mode: 'incremental',
			importScopes: expectedScopes,
			previewToken: 'preview-token-1',
		}));
		expect(mocks.toastSuccessMock).toHaveBeenCalledWith('\u9884\u68c0\u5b8c\u6210');
		expect(mocks.toastSuccessMock).toHaveBeenCalledWith('\u5bfc\u5165\u5df2\u5e94\u7528');
	});

	it('disables import when selective import has no scopes', async () => {
		const selectedFile = new File(['{"snapshot":true}'], 'snapshot-empty-selective.json', { type: 'application/json' });
		const { container } = render(<SettingBackup />);
		expect(document.querySelector('[data-testid="backup-page"]')).toBeInTheDocument();
		fireEvent.click(getSwitchForLabel('按范围导入'));
		for (const label of ['路由配置', '模型数据', 'API 密钥', '系统设置', '统计数据', '中继日志']) {
			fireEvent.click(getSwitchForLabel(label));
		}
		fireEvent.change(getFileInput(container), { target: { files: [selectedFile] } });

		expect(screen.getByTestId('backup-import-button')).toBeDisabled();
		expect(screen.getByText('请至少选中一个导入范围。')).toBeInTheDocument();
	});

	it('does not keep pending apply when preview token is missing', async () => {
		const selectedFile = new File(['{"snapshot":true}'], 'snapshot-missing-preview-token.json', { type: 'application/json' });
		mocks.importMutateAsyncMock.mockResolvedValue({
			rows_affected: { channels: 1 },
			dry_run: true,
			mode: 'incremental',
			compatibility: { conflicts: ['channel conflict'] },
		});

		const { container } = render(<SettingBackup />);
		fireEvent.change(getFileInput(container), { target: { files: [selectedFile] } });
		fireEvent.click(screen.getByTestId('backup-import-button'));

		await screen.findByText('预检行数');
		expect(screen.queryByText('预检已完成，可以继续应用同一份快照。')).not.toBeInTheDocument();
		expect(mocks.toastErrorMock).toHaveBeenCalledWith('\u9884\u68c0\u5df2\u5b8c\u6210\uff0c\u4f46\u6ca1\u6709\u8fd4\u56de\u53ef\u7528\u7684\u9884\u68c0\u4ee4\u724c\uff0c\u8bf7\u91cd\u65b0\u6267\u884c\u4e00\u6b21\u9884\u68c0\u3002');
	});

	it('shows localized model-mapping placeholder in zh-Hans map mode', async () => {
		const { container } = render(<SettingBackup />);
		await selectImportMode('map');
		expect(screen.getByTestId('backup-structured-mapping-source-0')).toHaveAttribute('placeholder', '旧模型=gpt-4o');
		expect(screen.getByTestId('backup-structured-mapping-target-0')).toHaveAttribute('placeholder', '视觉模型=gpt-4.1');
		expect(getModelMappingsTextarea(container).value).toBe('');
	});

	it('shows English model-mapping placeholder after locale switch', async () => {
		setLocale('en');
		const { container } = render(<SettingBackup />);
		await selectImportMode('map');
		expect(screen.getByTestId('backup-structured-mapping-source-0')).toHaveAttribute('placeholder', 'legacy-model=gpt-4o');
		expect(screen.getByTestId('backup-structured-mapping-target-0')).toHaveAttribute('placeholder', 'vision-model=gpt-4.1');
		expect(getModelMappingsTextarea(container).value).toBe('');
	});

	it('supports map mode previews and explicit model mappings', async () => {
		const selectedFile = new File(['{"snapshot":true}'], 'snapshot-map.json', { type: 'application/json' });
		mocks.importMutateAsyncMock.mockResolvedValue({
			rows_affected: { channels: 1, groups: 1 },
			preview_token: 'preview-token-map',
			dry_run: true,
			mode: 'map',
			compatibility: {
				conflicts: ['channel conflict'],
				alias_conflicts: ['alias conflict: legacy-vision -> gpt-4.1'],
				route_conflicts: ['route conflict: group-a -> legacy-model'],
				affected_groups: ['group-a', 'group-b'],
				affected_channels: ['Primary', 'Backup'],
				alias_preview_mappings: [
					{ snapshot_model: 'legacy-vision', current_model: 'gpt-4.1', canonical: 'gpt-4.1', contexts: ['routing'] },
				],
				missing_providers: ['legacy-provider'],
				missing_models: ['legacy-text-preview'],
				base_url_mismatches: ['preview-channel'],
				schema_mismatches: ['snapshot schema:v2 differs'],
				skipped_targets: ['channel_key:201 empty credential', 'setting:api_base_url existing row preserved by skip mode'],
				invalid_route_targets: [
					{ group_name: 'group-a', channel_name: 'Primary', model: 'legacy-model', resolved_model: 'gpt-4o', issue_type: 'missing_target', reason: 'channel key missing', action: 'rebind credential' },
				],
				skipped_route_target_previews: [
					{ group_name: 'group-b', channel_name: 'Backup', model: 'legacy-fallback', resolved_model: 'gpt-4.1', issue_type: 'skipped_preview', reason: 'model not declared on current channel', action: 'review mapping' },
				],
				model_policy_diffs: [
					{ model: 'legacy-model', current_model: 'gpt-4o', impact_level: 'high', changed_fields: ['billing_mode'], before: { billing_mode: 'paid' }, after: { billing_mode: 'free' }, contexts: ['routing'], warnings: ['policy drift'] },
				],
				route_preview_warnings: ['route may degrade'],
				route_preview_diffs: [{ group_name: 'group-a', model: 'legacy-model' }],
				model_mapping_previews: [
					{ source_model: 'legacy-model', target_model: 'gpt-4o', contexts: ['routing'], touched_fields: ['primary_model'], usage_count: 2, used: true, target_exists: true },
					{ source_model: 'missing-model', target_model: 'gpt-4.1-mini', contexts: ['fallback'], touched_fields: ['fallback_model'], usage_count: 1, used: true, target_exists: false, warnings: ['current model not found'] },
					{ source_model: 'unused-model', target_model: 'gpt-4.1', contexts: ['api_keys'], touched_fields: ['model'], usage_count: 0, used: false, target_exists: true },
				],
				credential_rebind_targets: [
					{ target_type: 'channel_key', channel_name: 'Primary', key_name: 'key-1', models: ['legacy-model'], affected_groups: ['group-a'] },
					{ target_type: 'api_key', key_name: 'client-key-1', affected_groups: ['group-b'] },
				],
				summary: {
					credential_rebind_targets: 1,
					channel_key_rebind_targets: 1,
					api_key_rebind_targets: 1,
				},
			},
		});

		const { container } = render(<SettingBackup />);
		await selectImportMode('map');
		expect(screen.getByTestId('backup-map-preview-root')).toBeInTheDocument();
		expect(screen.getByTestId('backup-structured-mapping-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-structured-mapping-count')).toHaveTextContent('已配置映射：0');
		expect(getHelpHintButtons().length).toBeGreaterThanOrEqual(9);
		expect(screen.getByTestId('backup-structured-mapping-empty')).toBeInTheDocument();
		setStructuredMappingRows([
			{ source: 'legacy-model', target: 'gpt-4o' },
			{ source: 'missing-model', target: 'gpt-4.1-mini' },
			{ source: 'unused-model', target: 'gpt-4.1' },
		]);
		expect(screen.getByTestId('backup-structured-mapping-count')).toHaveTextContent('已配置映射：3');
		expect(getModelMappingsTextarea(container).value).toBe('legacy-model=gpt-4o\nmissing-model=gpt-4.1-mini\nunused-model=gpt-4.1');
		fireEvent.change(getFileInput(container), { target: { files: [selectedFile] } });
		fireEvent.click(screen.getByTestId('backup-import-button'));

		await screen.findByTestId('backup-pending-apply-ready');
		expect(screen.getByTestId('backup-pending-apply-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-pending-apply-meta-grid')).toBeInTheDocument();
		expect(mocks.importMutateAsyncMock).toHaveBeenCalledWith(expect.objectContaining({
			file: selectedFile,
			mode: 'map',
			modelMappings: {
				'legacy-model': 'gpt-4o',
				'missing-model': 'gpt-4.1-mini',
				'unused-model': 'gpt-4.1',
			},
		}));
		expect(screen.getByTestId('backup-import-result-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-import-summary-grid')).toBeInTheDocument();
		expect(screen.getByTestId('backup-import-summary-rebinds')).toHaveTextContent('凭证重绑定：2');
		expect(screen.getByTestId('backup-import-summary-rebinds')).toHaveAttribute('data-raw-value', '2');
		expect(screen.getByTestId('backup-compatibility-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-overview')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-toggle')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-details')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-summary')).toHaveTextContent('详细诊断默认折叠，按需展开查看。');
		expect(screen.queryByTestId('backup-compatibility-mapping-preview-title')).not.toBeInTheDocument();
		fireEvent.click(getByRoleName('button', [/^展开\s+\d+\s*项?$/]));
		const mapSignalList = screen.getByTestId('backup-compatibility-signal-list');
		expect(mapSignalList).toBeInTheDocument();
		expect(mapSignalList).toHaveTextContent('兼容性报告发现 1 个缺失的提供商。');
		expect(mapSignalList).toHaveTextContent('1 个导入目标需要重新绑定渠道密钥凭证。');
		expect(mapSignalList).toHaveTextContent('兼容性报告跳过了 1 个路由目标预览。');
		expect(mapSignalList).toHaveTextContent('兼容性报告发现 1 个缺失的映射目标。');
		expect(mapSignalList).toHaveTextContent('兼容性报告发现 1 条未使用的模型映射。');
		expect(screen.getByTestId('backup-compatibility-guidance')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-guidance-title')).toHaveTextContent('下一步处理建议');
		expect(screen.getByTestId('backup-compatibility-guidance-item-0')).toHaveTextContent('先处理阻断风险');
		expect(screen.getByTestId('backup-compatibility-guidance-item-1')).toHaveTextContent('补齐缺失的渠道或模型');
		expect(screen.getByTestId('backup-compatibility-guidance-item-2')).toHaveTextContent('提前准备凭证重绑定');
		expect(screen.getByTestId('backup-compatibility-guidance-item-3')).toHaveTextContent('确认哪些对象会被跳过');
		expect(screen.getByTestId('backup-compatibility-guidance-item-4')).toHaveTextContent('修正模型映射后再应用');
		expect(screen.getByTestId('backup-compatibility-guidance-item-5')).toHaveTextContent('复核路由与策略差异');
		expect(screen.getByTestId('backup-compatibility-conflicts')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-conflicts-title')).toHaveTextContent('兼容冲突明细');
		expect(screen.getByTestId('backup-compatibility-conflicts-item-0')).toHaveTextContent('channel conflict');
		expect(screen.getByTestId('backup-compatibility-alias-conflicts')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-alias-conflicts-title')).toHaveTextContent('别名冲突');
		expect(screen.getByTestId('backup-compatibility-alias-conflicts-item-0')).toHaveTextContent('alias conflict: legacy-vision -> gpt-4.1');
		expect(screen.getByTestId('backup-compatibility-route-conflicts')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-route-conflicts-title')).toHaveTextContent('路由冲突');
		expect(screen.getByTestId('backup-compatibility-route-conflicts-item-0')).toHaveTextContent('route conflict: group-a -> legacy-model');
		expect(screen.getByTestId('backup-compatibility-affected-groups')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-affected-groups-title')).toHaveTextContent('受影响分组');
		expect(screen.getByTestId('backup-compatibility-affected-groups-item-0')).toHaveTextContent('group-a');
		expect(screen.getByTestId('backup-compatibility-affected-channels')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-affected-channels-title')).toHaveTextContent('受影响渠道');
		expect(screen.getByTestId('backup-compatibility-affected-channels-item-0')).toHaveTextContent('Primary');
		expect(screen.getByTestId('backup-compatibility-missing-providers')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-missing-providers-title')).toHaveTextContent('缺失渠道 / 供应商');
		expect(screen.getByTestId('backup-compatibility-missing-providers-item-0')).toHaveTextContent('legacy-provider');
		expect(screen.getByTestId('backup-compatibility-missing-models')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-missing-models-title')).toHaveTextContent('缺失模型');
		expect(screen.getByTestId('backup-compatibility-missing-models-item-0')).toHaveTextContent('legacy-text-preview');
		expect(screen.getByTestId('backup-compatibility-base-url-mismatches')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-base-url-mismatches-title')).toHaveTextContent('基础地址不匹配');
		expect(screen.getByTestId('backup-compatibility-base-url-mismatches-item-0')).toHaveTextContent('preview-channel');
		expect(screen.getByTestId('backup-compatibility-schema-mismatches')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-schema-mismatches-title')).toHaveTextContent('结构版本不匹配');
		expect(screen.getByTestId('backup-compatibility-schema-mismatches-item-0')).toHaveTextContent('快照结构版本 v2 与当前导入链路不一致');
		expect(screen.getByTestId('backup-compatibility-skipped-targets')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-skipped-targets-title')).toHaveTextContent('被保留或跳过的对象');
		expect(screen.getByTestId('backup-compatibility-skipped-targets-item-0')).toHaveTextContent('渠道密钥:201 缺少明文凭证');
		expect(screen.getByTestId('backup-compatibility-skipped-targets-item-1')).toHaveTextContent('系统设置:api_base_url 因跳过模式而保留当前记录');
		expect(screen.getByTestId('backup-compatibility-credential-rebind')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-credential-rebind-title')).toHaveTextContent('凭证重绑定目标');
		expect(screen.getByTestId('backup-compatibility-credential-rebind-item-0')).toHaveTextContent('Primary');
		expect(screen.getByTestId('backup-compatibility-invalid-route-targets')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-invalid-route-targets-title')).toHaveTextContent('路由目标风险');
		expect(screen.getByTestId('backup-compatibility-invalid-route-targets-item-0')).toHaveTextContent('missing_target');
		expect(screen.getByTestId('backup-compatibility-skipped-route-targets')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-skipped-route-targets-title')).toHaveTextContent('被跳过的路由预览');
		expect(screen.getByTestId('backup-compatibility-skipped-route-targets-item-0')).toHaveTextContent('skipped_preview');
		expect(screen.getByTestId('backup-compatibility-skipped-route-targets-item-0')).toHaveTextContent('review mapping');
		expect(screen.getByTestId('backup-compatibility-route-preview-warnings')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-route-preview-warnings-title')).toHaveTextContent('路由预警');
		expect(screen.getByTestId('backup-compatibility-route-preview-warnings-item-0')).toHaveTextContent('路由候选链可能降级');
		expect(screen.getByTestId('backup-compatibility-route-preview-diffs')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-route-preview-diffs-title')).toHaveTextContent('路由差异预览');
		expect(screen.getByTestId('backup-compatibility-route-preview-diffs-item-0')).toHaveTextContent('group-a');
		expect(screen.getByTestId('backup-compatibility-route-preview-diffs-item-0')).toHaveTextContent('legacy-model');
		expect(screen.getByTestId('backup-compatibility-mapping-preview')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-mapping-preview-title')).toHaveTextContent('模型映射预览');
		expect(screen.getByTestId('backup-compatibility-mapping-preview-item-0')).toHaveTextContent('legacy-model');
		expect(screen.getByTestId('backup-compatibility-mapping-preview-item-1')).toHaveTextContent('missing-model');
		expect(screen.getByTestId('backup-compatibility-mapping-preview-item-2')).toHaveTextContent('unused-model');
		expect(screen.getByTestId('backup-compatibility-missing-mapping')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-missing-mapping-title')).toHaveTextContent('缺失的映射目标');
		expect(screen.getByTestId('backup-compatibility-missing-mapping-item-0')).toHaveTextContent('missing-model');
		expect(screen.getByTestId('backup-compatibility-unused-mapping')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-unused-mapping-title')).toHaveTextContent('未使用的映射');
		expect(screen.getByTestId('backup-compatibility-unused-mapping-item-0')).toHaveTextContent('unused-model');
		expect(screen.getByTestId('backup-compatibility-alias-preview')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-alias-preview-title')).toHaveTextContent('别名映射预览');
		expect(screen.getByTestId('backup-compatibility-alias-preview-item-0')).toHaveTextContent('legacy-vision');
		expect(screen.getByTestId('backup-compatibility-model-policy-diffs')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-model-policy-diffs-title')).toHaveTextContent('模型策略差异');
		expect(screen.getByTestId('backup-compatibility-model-policy-diffs-item-0')).toHaveTextContent('legacy-model');
	});

	it('keeps import rebind summary correct when only split summary counts are returned', async () => {
		const selectedFile = new File(['{"snapshot":true}'], 'snapshot-map-summary-only.json', { type: 'application/json' });
		mocks.importMutateAsyncMock.mockResolvedValue({
			rows_affected: { channels: 1, groups: 1 },
			preview_token: 'preview-token-map-summary-only',
			dry_run: true,
			mode: 'map',
			compatibility: {
				summary: {
					credential_rebind_targets: 1,
					channel_key_rebind_targets: 1,
					api_key_rebind_targets: 1,
				},
			},
		});

		const { container } = render(<SettingBackup />);
		await selectImportMode('map');
		fireEvent.change(getFileInput(container), { target: { files: [selectedFile] } });
		fireEvent.click(screen.getByTestId('backup-import-button'));

		await screen.findByTestId('backup-pending-apply-ready');
		expect(screen.getByTestId('backup-import-summary-rebinds')).toHaveTextContent('凭证重绑定：2');
		expect(screen.getByTestId('backup-import-summary-rebinds')).toHaveAttribute('data-raw-value', '2');
	});

	it('shows replace-prune preview and requires confirmation in replace mode', async () => {
		const selectedFile = new File(['{"snapshot":true}'], 'snapshot-replace.json', { type: 'application/json' });
		mocks.importMutateAsyncMock.mockImplementation(async (payload: any) => {
			if (payload.dryRun) {
				return {
					rows_affected: { channels: 1, groups: 1, api_keys: 1 },
					preview_token: 'preview-token-replace',
					dry_run: true,
					mode: 'replace',
					compatibility: {
						conflicts: ['replace conflict'],
						credential_rebind_targets: [{ target_type: 'channel_key' }],
						replace_prune_preview: {
							warnings: ['API key cleanup preview excludes credentials that are absent from this snapshot'],
						},
						replace_pruned_channels: ['legacy-channel'],
						replace_pruned_groups: ['legacy-group'],
						replace_pruned_settings: ['proxy_url'],
						replace_pruned_llm_infos: ['legacy-model'],
						replace_pruned_api_keys: ['client-key'],
					},
				};
			}
			return {
				rows_affected: { channels: 1 },
				dry_run: false,
				mode: 'replace',
			};
		});

		const { container } = render(<SettingBackup />);
		expect(document.querySelector('[data-testid="backup-page"]')).toBeInTheDocument();
		await selectImportMode('replace');
		fireEvent.change(getFileInput(container), { target: { files: [selectedFile] } });
		fireEvent.click(screen.getByTestId('backup-import-button'));

		await screen.findByTestId('backup-pending-apply-ready');
		expect(screen.getByTestId('backup-pending-apply-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-pending-apply-meta-grid')).toBeInTheDocument();
		expect(screen.getByTestId('backup-import-result-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-import-summary-grid')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-overview')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-toggle')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-details')).toBeInTheDocument();
		fireEvent.click(screen.getByTestId('backup-compatibility-toggle'));
		const replaceSignalList = screen.getByTestId('backup-compatibility-signal-list');
		expect(replaceSignalList).toBeInTheDocument();
		expect(replaceSignalList).toHaveTextContent('替换模式会移除当前项目中那些未被快照保留的记录。');
		expect(replaceSignalList).toHaveTextContent('1 个冲突');
		expect(replaceSignalList).toHaveTextContent('1 个导入目标需要重新绑定渠道密钥凭证。');
		expect(replaceSignalList).toHaveTextContent('结构化清理预览还发现了 5 条额外记录。');
		expect(screen.getByTestId('backup-compatibility-guidance-item-2')).toHaveTextContent('确认替换会清理哪些当前记录');
		expect(screen.getByTestId('backup-compatibility-replace-pruned-channels')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-replace-pruned-channels-title')).toHaveTextContent('替换后会移除的渠道');
		expect(screen.getByTestId('backup-compatibility-replace-pruned-channels-item-0')).toHaveTextContent('legacy-channel');
		expect(screen.getByTestId('backup-compatibility-replace-pruned-groups')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-replace-pruned-groups-title')).toHaveTextContent('替换后会移除的分组');
		expect(screen.getByTestId('backup-compatibility-replace-pruned-groups-item-0')).toHaveTextContent('legacy-group');
		expect(screen.getByTestId('backup-compatibility-replace-pruned-settings')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-replace-pruned-settings-title')).toHaveTextContent('替换后会重置的设置');
		expect(screen.getByTestId('backup-compatibility-replace-pruned-settings-item-0')).toHaveTextContent('proxy_url');
		expect(screen.getByTestId('backup-compatibility-replace-pruned-llm-infos')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-replace-pruned-llm-infos-title')).toHaveTextContent('替换后会移除的模型信息');
		expect(screen.getByTestId('backup-compatibility-replace-pruned-llm-infos-item-0')).toHaveTextContent('legacy-model');
		expect(screen.getByTestId('backup-compatibility-replace-pruned-api-keys')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-replace-pruned-api-keys-title')).toHaveTextContent('替换后会移除的 API 密钥');
		expect(screen.getByTestId('backup-compatibility-replace-pruned-api-keys-item-0')).toHaveTextContent('client-key');
		expect(screen.getByTestId('backup-replace-prune-panel')).toBeInTheDocument();
		expect(screen.queryByTestId('backup-import-remaining-migration-trigger')).not.toBeInTheDocument();
		expect(screen.queryByTestId('backup-import-remaining-migration-panel')).not.toBeInTheDocument();
		fireEvent.click(screen.getByTestId('backup-replace-prune-trigger'));
		expect(screen.getByTestId('backup-replace-prune-section-channels')).toBeInTheDocument();
		expect(screen.getByTestId('backup-replace-prune-section-title-channels')).toHaveTextContent('待删除渠道');
		expect(screen.getByTestId('backup-replace-prune-section-item-channels-0')).toHaveTextContent('legacy-channel');
		expect(screen.getByTestId('backup-replace-prune-section-groups')).toBeInTheDocument();
		expect(screen.getByTestId('backup-replace-prune-section-title-groups')).toHaveTextContent('待删除分组');
		expect(screen.getByTestId('backup-replace-prune-section-item-groups-0')).toHaveTextContent('legacy-group');
		expect(screen.getByTestId('backup-replace-prune-section-settings')).toBeInTheDocument();
		expect(screen.getByTestId('backup-replace-prune-section-title-settings')).toHaveTextContent('待重置设置');
		expect(screen.getByTestId('backup-replace-prune-section-item-settings-0')).toHaveTextContent('proxy_url');
		expect(screen.getByTestId('backup-replace-prune-section-models')).toBeInTheDocument();
		expect(screen.getByTestId('backup-replace-prune-section-title-models')).toHaveTextContent('待移除模型条目');
		expect(screen.getByTestId('backup-replace-prune-section-item-models-0')).toHaveTextContent('legacy-model');
		expect(screen.getByTestId('backup-replace-prune-section-apiKeys')).toBeInTheDocument();
		expect(screen.getByTestId('backup-replace-prune-section-title-apiKeys')).toHaveTextContent('待删除 API 密钥');
		expect(screen.getByTestId('backup-replace-prune-section-item-apiKeys-0')).toHaveTextContent('client-key');
		expect(screen.getByTestId('backup-replace-prune-section-warnings')).toBeInTheDocument();
		expect(screen.getByTestId('backup-replace-prune-section-title-warnings')).toHaveTextContent('额外提示');
		expect(screen.getByTestId('backup-replace-prune-section-item-warnings-0')).toHaveTextContent('API key cleanup preview excludes credentials that are absent from this snapshot');
		const applyButton = screen.getByTestId('backup-apply-same-import-button');
		expect(applyButton).toBeDisabled();
		expect(screen.getByTestId('backup-apply-confirm-panel')).toBeInTheDocument();
		fireEvent.click(getApplyConfirmSwitchBySelector());
		await waitFor(() => {
			expect(applyButton).toBeEnabled();
		});
		fireEvent.click(applyButton);
		await waitFor(() => {
			expect(mocks.importMutateAsyncMock).toHaveBeenCalledTimes(2);
		});
		expect(mocks.importMutateAsyncMock).toHaveBeenNthCalledWith(2, expect.objectContaining({
			dryRun: false,
			previewToken: 'preview-token-replace',
			mode: 'replace',
		}));
	});

	it('previews and applies rollback snapshots', async () => {
		setLocale('en');
		const initialRollbackScopes = {
			routing: false,
			models: true,
			api_keys: true,
			settings: true,
			stats: false,
			logs: false,
		};
		const narrowedRollbackScopes = {
			routing: false,
			models: true,
			api_keys: true,
			settings: false,
			stats: false,
			logs: false,
		};
		mocks.importSnapshotsState.data = [{
			snapshot_name: 'snapshot-selected-lock',
			snapshot_path: 'snapshots/snapshot-selected-lock.json',
			imported_at: '2026-04-21T10:00:00Z',
			size_bytes: 2048,
			is_latest: true,
		}];
		mocks.previewRollbackMutateAsyncMock.mockImplementation(async (payload: any) => ({
			snapshot_name: 'snapshot-selected-lock',
			applied_scopes: payload.importScopes,
			manifest: { encrypted: undefined, contains_secrets: true, schema_version: '10' },
			rows_summary: {
				channels: 1,
				users: 2,
				migration_records: 1,
				ai_tasks: 1,
				ai_prompt_templates: 1,
				dynamic_route_learning_states: 1,
			},
			compatibility: {
				conflicts: ['channel conflict'],
				alias_conflicts: ['alias conflict: rollback-vision -> gpt-4.1'],
				route_conflicts: ['route conflict: rollback-group -> legacy-model'],
				credential_rebind_targets: [
					{ target_type: 'channel_key', channel_name: 'Primary', key_name: 'key-1', models: ['gpt-4o'], affected_groups: ['group-a'] },
					{ target_type: 'api_key', key_name: 'rollback-client-key', affected_groups: ['group-b'] },
				],
				summary: {
					credential_rebind_targets: 1,
					channel_key_rebind_targets: 1,
					api_key_rebind_targets: 1,
				},
				affected_groups: ['group-a'],
				affected_channels: ['Primary'],
				missing_providers: ['rollback-provider'],
				missing_models: ['legacy-model'],
				base_url_mismatches: ['rollback-channel'],
				schema_mismatches: ['snapshot schema:v2 differs'],
				skipped_targets: ['channel_key:101 empty credential'],
				replace_pruned_channels: ['current-channel-a'],
				replace_pruned_groups: ['current-group-a'],
				replace_pruned_settings: ['proxy_url'],
				replace_pruned_llm_infos: ['current-legacy-model'],
				replace_pruned_api_keys: ['current-client-key'],
				invalid_route_targets: [{ group_name: 'rollback-group', channel_name: 'Primary', model: 'legacy-model', issue_type: 'missing_target', reason: 'channel removed', action: 'rebind channel' }],
				skipped_route_target_previews: [{ group_name: 'rollback-group', channel_name: 'Primary', model: 'legacy-model', issue_type: 'skipped_preview', reason: 'preview omitted', action: 'review mapping' }],
				route_preview_warnings: ['rollback route may degrade'],
				model_mapping_previews: [{ source_model: 'legacy-model', target_model: 'gpt-4o', contexts: ['routing'], touched_fields: ['model'], usage_count: 1, used: true, target_exists: true }],
				alias_preview_mappings: [{ snapshot_model: 'rollback-vision', current_model: 'gpt-4.1', canonical: 'gpt-4.1', contexts: ['routing'] }],
				model_policy_diffs: [{ model: 'legacy-model', current_model: 'gpt-4o', impact_level: 'high', changed_fields: ['billing_mode'], before: { billing_mode: 'paid' }, after: { billing_mode: 'free' }, contexts: ['routing'], warnings: ['policy drift'] }],
				route_preview_diffs: [{
					group_name: 'group-a',
					model: 'gpt-4o',
					before_candidates: [{ channel_name: 'current-primary', model: 'gpt-4o', priority: 1, weight: 100, enabled: true, declared: true, has_key: true }],
					after_candidates: [{ channel_name: 'snapshot-primary', model: 'gpt-4o', priority: 1, weight: 100, enabled: true, declared: true, has_key: true }],
					removed_candidates: [{ channel_name: 'current-primary', model: 'gpt-4o', priority: 1, weight: 100, enabled: true, declared: true, has_key: true }],
					added_candidates: [{ channel_name: 'snapshot-primary', model: 'gpt-4o', priority: 1, weight: 100, enabled: true, declared: true, has_key: true }],
					fallback_changed: true,
				}],
			},
			preview_warnings: ['route preview needs manual review'],
		}));
		mocks.rollbackImportSnapshotMutateAsyncMock.mockImplementation(async (payload: any) => ({ snapshot_name: 'snapshot-selected-lock', applied_scopes: payload.importScopes }));

		render(<SettingBackup />);
		fireEvent.click(screen.getByTestId('backup-history-trigger'));
		expect(screen.getByTestId('backup-history-panel')).toBeInTheDocument();
		expect(document.querySelector('[data-testid="backup-page"]')).toBeInTheDocument();
		expect(screen.getByTestId('backup-history-list')).toBeInTheDocument();
		const historyItem = screen.getByTestId('backup-history-item-snapshot-selected-lock');
		expect(historyItem).toBeInTheDocument();
		expect(within(historyItem).getByTestId('backup-history-item-actions')).toBeInTheDocument();
		expect(within(historyItem).getByTestId('backup-history-item-meta')).toBeInTheDocument();
		expect(within(historyItem).getByTestId('backup-history-item-name')).toHaveTextContent('snapshot-selected-lock');
		expect(within(historyItem).getByTestId('backup-history-item-path')).toHaveTextContent('snapshots/snapshot-selected-lock.json');
		expect(within(historyItem).getByTestId('backup-history-item-path')).toHaveAttribute('data-raw-value', 'snapshots/snapshot-selected-lock.json');
		expect(within(historyItem).getByTestId('backup-history-item-size')).toHaveTextContent('Size：2 KB');
		expect(within(historyItem).getByTestId('backup-history-item-size')).toHaveAttribute('data-size-bytes', '2048');
		expect(within(historyItem).getByTestId('backup-history-item-imported-at')).toHaveTextContent('2026');
		expect(within(historyItem).getByTestId('backup-history-item-imported-at')).toHaveAttribute('data-raw-value', '2026-04-21T10:00:00Z');
		expect(within(historyItem).getByTestId('backup-history-item-latest-badge')).toHaveTextContent('Latest');
		expect(within(historyItem).getByTestId('backup-history-item-latest-badge')).toHaveAttribute('data-is-latest', 'true');
		expect(screen.getByTestId('backup-rollback-scope-editor')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-scope-editor-title')).toHaveTextContent('Rollback domains');
		expect(screen.getByTestId('backup-rollback-scope-current-summary')).toHaveTextContent('Rollback Scope：Full snapshot restore');
		expect(getHelpHintButtons().length).toBeGreaterThanOrEqual(9);
		fireEvent.click(screen.getByTestId('backup-rollback-selective-switch'));
		expect(screen.getByTestId('backup-rollback-scope-grid')).toBeInTheDocument();
		fireEvent.click(screen.getByTestId('backup-rollback-scope-routing'));
		fireEvent.click(screen.getByTestId('backup-rollback-scope-stats'));
		fireEvent.click(screen.getByTestId('backup-rollback-scope-logs'));
		expect(screen.getByTestId('backup-rollback-scope-current-summary')).toHaveTextContent('Rollback Scope：Models, API keys, Settings');
		fireEvent.click(within(historyItem).getByTestId('backup-history-preview-button'));
		await screen.findByText('Rollback preview');
		expect(mocks.previewRollbackMutateAsyncMock).toHaveBeenNthCalledWith(1, expect.objectContaining({
			snapshotName: 'snapshot-selected-lock',
			importScopes: initialRollbackScopes,
		}));
		expect(screen.getByTestId('backup-rollback-preview-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-preview-header')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-preview-title')).toHaveTextContent('Rollback preview');
		expect(screen.getByTestId('backup-rollback-preview-name')).toHaveTextContent('snapshot-selected-lock');
		expect(screen.getByTestId('backup-rollback-preview-name')).toHaveAttribute('data-raw-value', 'snapshot-selected-lock');
		expect(screen.getByTestId('backup-rollback-preview-overview')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-preview-summary-grid')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-preview-summary-conflicts')).toHaveTextContent('Compatibility Conflicts：1');
		expect(screen.getByTestId('backup-rollback-preview-summary-conflicts')).toHaveAttribute('data-raw-value', '1');
		expect(screen.getByTestId('backup-rollback-preview-summary-rebinds')).toHaveTextContent('Credential Rebinds：2');
		expect(screen.getByTestId('backup-rollback-preview-summary-rebinds')).toHaveAttribute('data-raw-value', '2');
		expect(screen.getByTestId('backup-rollback-preview-summary-warnings')).toHaveTextContent('Preview Warnings：2');
		expect(screen.getByTestId('backup-rollback-preview-summary-warnings')).toHaveAttribute('data-raw-value', '2');
		expect(screen.getByTestId('backup-rollback-preview-meta-grid')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-preview-meta-scope')).toHaveTextContent('Rollback Scope：Models, API keys, Settings');
		expect(screen.getByTestId('backup-rollback-preview-meta-scope')).toHaveAttribute('data-raw-value', 'models,api_keys,settings');
		expect(screen.getByTestId('backup-rollback-preview-meta-encrypted')).toHaveTextContent('Encryption：Unknown');
		expect(screen.getByTestId('backup-rollback-preview-meta-encrypted')).not.toHaveAttribute('data-raw-value');
		expect(screen.getByTestId('backup-rollback-preview-meta-contains-secrets')).toHaveTextContent('Contains Credentials：Yes');
		expect(screen.getByTestId('backup-rollback-preview-meta-contains-secrets')).toHaveAttribute('data-raw-value', 'true');
		expect(screen.getByTestId('backup-rollback-preview-meta-schema-version')).toHaveTextContent('Schema Version：10');
		expect(screen.getByTestId('backup-rollback-preview-meta-schema-version')).toHaveAttribute('data-raw-value', '10');
		expect(screen.getByTestId('backup-rollback-signal-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-signal-title')).toHaveTextContent('Recommended Rollback Steps');
		expect(screen.getByTestId('backup-rollback-signal-summary')).toHaveTextContent('Start with the summary signals');
		expect(screen.getByTestId('backup-rollback-signal-list')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-signal-list')).toHaveTextContent('Rollback preview emitted 2 warnings.');
		expect(screen.getByTestId('backup-rollback-signal-list')).toHaveTextContent('Rollback preview found 1 conflicts.');
		expect(screen.getByTestId('backup-rollback-signal-list')).toHaveTextContent('Channel-key credential rebind is required for 1 restored targets.');
		expect(screen.getByTestId('backup-rollback-signal-list')).toHaveTextContent('Rollback diagnostics marked 5 current records for removal or reset.');
		expect(screen.getByTestId('backup-rollback-guidance')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-guidance-title')).toHaveTextContent('Recommended Rollback Steps');
		expect(screen.getByTestId('backup-rollback-guidance-item-0')).toHaveTextContent('Resolve rollback risks first');
		expect(screen.getByTestId('backup-rollback-guidance-item-1')).toHaveTextContent('Restore missing targets before rollback');
		expect(screen.getByTestId('backup-rollback-guidance-item-2')).toHaveTextContent('Prepare post-rollback credential rebinds');
		expect(screen.getByTestId('backup-rollback-guidance-item-3')).toHaveTextContent('Review which targets rollback keeps');
		expect(screen.getByTestId('backup-rollback-guidance-item-4')).toHaveTextContent('Review which current records rollback removes');
		expect(screen.getByTestId('backup-rollback-guidance-item-5')).toHaveTextContent('Review post-rollback route and policy drift');
		expect(screen.getByTestId('backup-rollback-preview-warnings-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-preview-warnings-title')).toHaveTextContent('Rollback Preview Warnings');
		expect(screen.getByTestId('backup-rollback-preview-warnings-summary')).toHaveTextContent('These warnings come from the rollback preview itself');
		expect(screen.getByTestId('backup-rollback-preview-warnings-list-item-0')).toHaveTextContent('route preview needs manual review');
		expect(screen.getByTestId('backup-rollback-preview-warnings-list-item-1')).toHaveTextContent('rollback route may degrade');
		expect(screen.getByTestId('backup-rollback-compatibility-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-compatibility-title')).toHaveTextContent('Rollback compatibility details');
		expect(screen.getByTestId('backup-rollback-compatibility-conflicts-item-0')).toHaveTextContent('channel conflict');
		expect(screen.getByTestId('backup-rollback-compatibility-alias-conflicts-item-0')).toHaveTextContent('alias conflict: rollback-vision -> gpt-4.1');
		expect(screen.getByTestId('backup-rollback-compatibility-route-conflicts-item-0')).toHaveTextContent('route conflict: rollback-group -> legacy-model');
		expect(screen.getByTestId('backup-rollback-compatibility-affected-groups-item-0')).toHaveTextContent('group-a');
		expect(screen.getByTestId('backup-rollback-compatibility-affected-channels-item-0')).toHaveTextContent('Primary');
		expect(screen.getByTestId('backup-rollback-compatibility-missing-providers-item-0')).toHaveTextContent('rollback-provider');
		expect(screen.getByTestId('backup-rollback-compatibility-missing-models-item-0')).toHaveTextContent('legacy-model');
		expect(screen.getByTestId('backup-rollback-compatibility-base-url-mismatches-item-0')).toHaveTextContent('rollback-channel');
		expect(screen.getByTestId('backup-rollback-compatibility-schema-mismatches-item-0')).toHaveTextContent('snapshot schema:v2 differs');
		expect(screen.getByTestId('backup-rollback-compatibility-skipped-targets-item-0')).toHaveTextContent('channel_key:101 empty credential');
		expect(screen.getByTestId('backup-rollback-compatibility-rows-summary-title')).toHaveTextContent('High-risk restored objects');
		expect(screen.getByTestId('backup-rollback-compatibility-rows-summary-item-0')).toHaveTextContent('Admin users: 2');
		expect(screen.getByTestId('backup-rollback-compatibility-rows-summary-item-1')).toHaveTextContent('Migration records: 1');
		expect(screen.getByTestId('backup-rollback-compatibility-rows-summary-item-2')).toHaveTextContent('AI tasks: 1');
		expect(screen.getByTestId('backup-rollback-compatibility-rows-summary-item-3')).toHaveTextContent('AI prompt templates: 1');
		expect(screen.getByTestId('backup-rollback-compatibility-rows-summary-item-4')).toHaveTextContent('Dynamic route learning states: 1');
		expect(screen.getByTestId('backup-rollback-compatibility-credential-rebind-item-0')).toHaveTextContent('Primary');
		expect(screen.getByTestId('backup-rollback-compatibility-replace-pruned-channels-title')).toHaveTextContent('Current channels removed by rollback');
		expect(screen.getByTestId('backup-rollback-compatibility-replace-pruned-channels-item-0')).toHaveTextContent('current-channel-a');
		expect(screen.getByTestId('backup-rollback-compatibility-replace-pruned-groups-item-0')).toHaveTextContent('current-group-a');
		expect(screen.getByTestId('backup-rollback-compatibility-replace-pruned-settings-item-0')).toHaveTextContent('proxy_url');
		expect(screen.getByTestId('backup-rollback-compatibility-replace-pruned-llm-infos-item-0')).toHaveTextContent('current-legacy-model');
		expect(screen.getByTestId('backup-rollback-compatibility-replace-pruned-api-keys-item-0')).toHaveTextContent('current-client-key');
		expect(screen.getByTestId('backup-rollback-compatibility-invalid-route-targets-item-0')).toHaveTextContent('missing_target');
		expect(screen.getByTestId('backup-rollback-compatibility-skipped-route-targets-item-0')).toHaveTextContent('review mapping');
		expect(screen.getByTestId('backup-rollback-compatibility-route-preview-warnings-item-0')).toHaveTextContent('rollback route may degrade');
		expect(screen.getByTestId('backup-rollback-compatibility-mapping-preview-item-0')).toHaveTextContent('legacy-model');
		expect(screen.getByTestId('backup-rollback-compatibility-alias-preview-item-0')).toHaveTextContent('rollback-vision');
		expect(screen.getByTestId('backup-rollback-compatibility-model-policy-diffs-item-0')).toHaveTextContent('legacy-model');
		expect(screen.getByTestId('backup-rollback-route-diff-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-route-diff-title')).toHaveTextContent('Route diff compare');
		expect(screen.getByTestId('backup-rollback-route-diff-row-title-0')).toHaveTextContent('group-a / gpt-4o');
		expect(screen.getByTestId('backup-rollback-route-diff-current-0')).toHaveTextContent('current-primary:gpt-4o');
		expect(screen.getByTestId('backup-rollback-route-diff-snapshot-0')).toHaveTextContent('snapshot-primary:gpt-4o');
		expect(screen.getByTestId('backup-rollback-route-diff-removed-0')).toHaveTextContent('current-primary:gpt-4o');
		expect(screen.getByTestId('backup-rollback-route-diff-added-0')).toHaveTextContent('snapshot-primary:gpt-4o');
		expect(screen.getByTestId('backup-rollback-route-diff-row-fallback-0')).toHaveTextContent('Yes');

		fireEvent.click(screen.getByTestId('backup-rollback-scope-settings'));
		expect(screen.queryByTestId('backup-rollback-preview-panel')).not.toBeInTheDocument();
		expect(mocks.previewRollbackMutateAsyncMock).toHaveBeenCalledTimes(1);
		expect(screen.getByTestId('backup-rollback-scope-current-summary')).toHaveTextContent('Rollback Scope：Models, API keys');
		fireEvent.click(within(historyItem).getByTestId('backup-history-preview-button'));
		await screen.findByText('Rollback preview');
		expect(mocks.previewRollbackMutateAsyncMock).toHaveBeenNthCalledWith(2, expect.objectContaining({
			snapshotName: 'snapshot-selected-lock',
			importScopes: narrowedRollbackScopes,
		}));
		expect(screen.getByTestId('backup-rollback-preview-meta-scope')).toHaveTextContent('Rollback Scope：Models, API keys');
		expect(screen.getByTestId('backup-rollback-preview-meta-scope')).toHaveAttribute('data-raw-value', 'models,api_keys');

		const previousConfirm = window.confirm;
		window.confirm = () => true;
		fireEvent.click(within(historyItem).getByTestId('backup-history-rollback-button'));
		await waitFor(() => {
		expect(mocks.rollbackImportSnapshotMutateAsyncMock).toHaveBeenCalledTimes(1);
		});
		window.confirm = previousConfirm;
		expect(mocks.rollbackImportSnapshotMutateAsyncMock).toHaveBeenNthCalledWith(1, expect.objectContaining({
			snapshotName: 'snapshot-selected-lock',
			importScopes: narrowedRollbackScopes,
		}));
		expect(mocks.toastSuccessMock).toHaveBeenCalledWith(expect.stringContaining('snapshot-selected-lock'));
		expect(screen.queryByTestId('backup-advanced-pending-title')).not.toBeInTheDocument();
		expect(screen.queryByTestId('backup-advanced-pending-summary')).not.toBeInTheDocument();
		expect(screen.queryByTestId('backup-remaining-migration-trigger')).not.toBeInTheDocument();
		expect(screen.queryByTestId('backup-remaining-migration-panel')).not.toBeInTheDocument();
	});

	it('falls back to full rollback when selective rollback scopes are empty', async () => {
		setLocale('en');
		mocks.importSnapshotsState.data = [{
			snapshot_name: 'snapshot-fallback',
			snapshot_path: 'snapshots/snapshot-fallback.json',
			imported_at: '2026-04-21T11:00:00Z',
			size_bytes: 1024,
			is_latest: true,
		}];
		mocks.previewRollbackMutateAsyncMock.mockImplementation(async (payload: any) => ({
			snapshot_name: 'snapshot-fallback',
			applied_scopes: payload.importScopes,
			manifest: { contains_secrets: true, schema_version: '10' },
			compatibility: {},
			preview_warnings: [],
		}));
		mocks.rollbackImportSnapshotMutateAsyncMock.mockResolvedValue({ snapshot_name: 'snapshot-fallback' });

		render(<SettingBackup />);
		fireEvent.click(screen.getByTestId('backup-history-trigger'));
		const historyItem = screen.getByTestId('backup-history-item-snapshot-fallback');
		fireEvent.click(screen.getByTestId('backup-rollback-selective-switch'));
		for (const scopeKey of ['routing', 'models', 'api_keys', 'settings', 'stats', 'logs'] as const) {
			fireEvent.click(screen.getByTestId(`backup-rollback-scope-${scopeKey}`));
		}
		expect(screen.getByTestId('backup-rollback-scope-current-summary')).toHaveTextContent('Rollback Scope：Full snapshot restore');
		expect(screen.getByTestId('backup-rollback-scope-mode-note')).toHaveTextContent('No rollback domains are selected. Preview and rollback will fall back to a full snapshot restore.');

		fireEvent.click(within(historyItem).getByTestId('backup-history-preview-button'));
		await screen.findByText('Rollback preview');
		expect(mocks.previewRollbackMutateAsyncMock).toHaveBeenCalledWith(expect.objectContaining({
			snapshotName: 'snapshot-fallback',
			importScopes: undefined,
		}));
		expect(screen.getByTestId('backup-rollback-preview-meta-scope')).toHaveTextContent('Rollback Scope：Unknown');
		expect(screen.getByTestId('backup-rollback-preview-meta-scope')).not.toHaveAttribute('data-raw-value');

		const previousConfirm = window.confirm;
		window.confirm = () => true;
		fireEvent.click(within(historyItem).getByTestId('backup-history-rollback-button'));
		await waitFor(() => {
			expect(mocks.rollbackImportSnapshotMutateAsyncMock).toHaveBeenCalledTimes(1);
		});
		window.confirm = previousConfirm;
		expect(mocks.rollbackImportSnapshotMutateAsyncMock).toHaveBeenCalledWith(expect.objectContaining({
			snapshotName: 'snapshot-fallback',
			importScopes: undefined,
		}));
	});

	it('keeps rollback rebind summary correct when only split summary counts are returned', async () => {
		setLocale('en');
		mocks.importSnapshotsState.data = [{
			snapshot_name: 'snapshot-summary-only',
			snapshot_path: 'snapshots/snapshot-summary-only.json',
			imported_at: '2026-04-21T10:00:00Z',
			size_bytes: 1024,
			is_latest: true,
		}];
		mocks.previewRollbackMutateAsyncMock.mockImplementation(async (payload: any) => ({
			snapshot_name: 'snapshot-summary-only',
			applied_scopes: payload.importScopes,
			manifest: { contains_secrets: true, schema_version: '10' },
			rows_summary: { channels: 1 },
			preview_warnings: [],
			compatibility: {
				summary: {
					credential_rebind_targets: 1,
					channel_key_rebind_targets: 1,
					api_key_rebind_targets: 1,
				},
			},
		}));

		render(<SettingBackup />);
		fireEvent.click(screen.getByTestId('backup-history-trigger'));
		const historyItem = await screen.findByTestId('backup-history-item-snapshot-summary-only');
		fireEvent.click(within(historyItem).getByTestId('backup-history-preview-button'));

		await screen.findByText('Rollback preview');
		expect(screen.getByTestId('backup-rollback-preview-summary-rebinds')).toHaveTextContent('Credential Rebinds：2');
		expect(screen.getByTestId('backup-rollback-preview-summary-rebinds')).toHaveAttribute('data-raw-value', '2');
	});

	it('merges rollback preview warnings with compatibility route-preview warnings without double counting', async () => {
		setLocale('en');
		mocks.importSnapshotsState.data = [{
			snapshot_name: 'snapshot-warning-merge',
			snapshot_path: 'snapshots/snapshot-warning-merge.json',
			imported_at: '2026-04-21T10:00:00Z',
			size_bytes: 1024,
			is_latest: true,
		}];
		mocks.previewRollbackMutateAsyncMock.mockImplementation(async (payload: any) => ({
			snapshot_name: 'snapshot-warning-merge',
			applied_scopes: payload.importScopes,
			manifest: { contains_secrets: true, schema_version: '10' },
			rows_summary: { channels: 1 },
			preview_warnings: ['route preview needs manual review'],
			compatibility: {
				route_preview_warnings: ['rollback route may degrade', 'route preview needs manual review'],
				summary: {
					route_preview_warnings: 1,
				},
			},
		}));

		render(<SettingBackup />);
		fireEvent.click(screen.getByTestId('backup-history-trigger'));
		const historyItem = await screen.findByTestId('backup-history-item-snapshot-warning-merge');
		fireEvent.click(within(historyItem).getByTestId('backup-history-preview-button'));

		await screen.findByText('Rollback preview');
		expect(screen.getByTestId('backup-rollback-preview-summary-warnings')).toHaveTextContent('Preview Warnings：2');
		expect(screen.getByTestId('backup-rollback-preview-summary-warnings')).toHaveAttribute('data-raw-value', '2');
		expect(screen.getByTestId('backup-rollback-signal-list')).toHaveTextContent('Rollback preview emitted 2 warnings.');
		expect(screen.getByTestId('backup-rollback-guidance-item-0')).toHaveTextContent('Review post-rollback route and policy drift');
		expect(screen.getByTestId('backup-rollback-guidance-item-0')).toHaveTextContent('2 route warnings');
		expect(screen.getByTestId('backup-rollback-preview-warnings-list-item-0')).toHaveTextContent('route preview needs manual review');
		expect(screen.getByTestId('backup-rollback-preview-warnings-list-item-1')).toHaveTextContent('rollback route may degrade');
		expect(screen.queryByTestId('backup-rollback-preview-warnings-list-item-2')).not.toBeInTheDocument();
	});
});
