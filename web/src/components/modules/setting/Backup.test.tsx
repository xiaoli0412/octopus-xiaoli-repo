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
	const textarea = container.querySelector('textarea');
	if (!(textarea instanceof HTMLTextAreaElement)) {
		throw new Error('Model mapping textarea not found');
	}
	return textarea;
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
		expect(screen.getByTestId('backup-compatibility-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-overview')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-toggle')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-details')).toBeInTheDocument();
		expect(screen.getByTestId('backup-apply-confirm-switch')).toBeInTheDocument();
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
		const textarea = getModelMappingsTextarea(container);
		expect(textarea.placeholder).toBe('旧模型=gpt-4o\n视觉模型=gpt-4.1');
	});

	it('shows English model-mapping placeholder after locale switch', async () => {
		setLocale('en');
		const { container } = render(<SettingBackup />);
		await selectImportMode('map');
		const textarea = getModelMappingsTextarea(container);
		expect(textarea.placeholder).toBe('legacy-model=gpt-4o\nvision-model=gpt-4.1');
	});

	it('supports map mode previews and explicit model mappings', async () => {
		const selectedFile = new File(['{"snapshot":true}'], 'snapshot-map.json', { type: 'application/json' });
		mocks.importMutateAsyncMock.mockResolvedValue({
			rows_affected: { channels: 1, groups: 1 },
			preview_token: 'preview-token-map',
			dry_run: true,
			mode: 'map',
			compatibility: {
				alias_preview_mappings: [
					{ snapshot_model: 'legacy-vision', current_model: 'gpt-4.1', canonical: 'gpt-4.1', contexts: ['routing'] },
				],
				model_mapping_previews: [
					{ source_model: 'legacy-model', target_model: 'gpt-4o', contexts: ['routing'], touched_fields: ['primary_model'], usage_count: 2, used: true, target_exists: true },
					{ source_model: 'missing-model', target_model: 'gpt-4.1-mini', contexts: ['fallback'], touched_fields: ['fallback_model'], usage_count: 1, used: true, target_exists: false, warnings: ['current model not found'] },
					{ source_model: 'unused-model', target_model: 'gpt-4.1', contexts: ['api_keys'], touched_fields: ['model'], usage_count: 0, used: false, target_exists: true },
				],
				credential_rebind_targets: [{ target_type: 'channel_key', channel_name: 'Primary', key_name: 'key-1', models: ['legacy-model'], affected_groups: ['group-a'] }],
			},
		});

		const { container } = render(<SettingBackup />);
		await selectImportMode('map');
		expect(screen.getByTestId('backup-map-preview-root')).toBeInTheDocument();
		expect(screen.getByTestId('backup-import-remaining-migration-title')).toHaveTextContent('导入补强项');
		expect(screen.getByTestId('backup-import-remaining-migration-summary')).toHaveTextContent('默认收起，按需查看仍需手动处理的迁移能力。');
		expect(screen.getByTestId('backup-import-remaining-migration-trigger')).toBeInTheDocument();
		expect(screen.getByTestId('backup-import-remaining-migration-panel')).toHaveAttribute('aria-hidden', 'true');
		expect(screen.getByTestId('backup-import-remaining-migration-panel')).toHaveClass('hidden');
		expect(getHelpHintButtons().length).toBeGreaterThanOrEqual(9);
		fireEvent.change(getModelMappingsTextarea(container), { target: { value: 'legacy-model=gpt-4o\nmissing-model=gpt-4.1-mini\nunused-model=gpt-4.1' } });
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
		expect(screen.getByTestId('backup-compatibility-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-overview')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-toggle')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-details')).toBeInTheDocument();
		expect(screen.getByTestId('backup-compatibility-summary')).toHaveTextContent('详细诊断默认折叠，按需展开查看。');
		fireEvent.click(screen.getByTestId('backup-import-remaining-migration-trigger'));
		expect(screen.getByTestId('backup-import-remaining-migration-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-import-remaining-migration-section-trigger-0')).toBeInTheDocument();
		expect(screen.queryByTestId('backup-import-remaining-migration-section-item-import-tooling-conflict-handling')).not.toBeInTheDocument();
		fireEvent.click(screen.getByTestId('backup-import-remaining-migration-section-trigger-0'));
		expect(screen.getByTestId('backup-import-remaining-migration-section-panel-0')).toBeInTheDocument();
		expect(screen.getByTestId('backup-import-remaining-migration-section-item-import-tooling-conflict-handling')).toBeInTheDocument();
		expect(screen.queryByTestId('backup-compatibility-mapping-preview-title')).not.toBeInTheDocument();
		fireEvent.click(getByRoleName('button', [/^展开\s+\d+\s*项?$/]));
		const mapSignalList = screen.getByTestId('backup-compatibility-signal-list');
		expect(mapSignalList).toBeInTheDocument();
		expect(mapSignalList).toHaveTextContent('1 个导入目标需要重新绑定渠道密钥凭证。');
		expect(mapSignalList).toHaveTextContent('兼容性报告发现 1 个缺失的映射目标。');
		expect(mapSignalList).toHaveTextContent('兼容性报告发现 1 条未使用的模型映射。');
		expect(screen.queryByTestId('backup-compatibility-missing-providers')).not.toBeInTheDocument();
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
					replace_prune_preview: {
						pruned_channels: ['legacy-channel'],
						pruned_api_keys: ['client-key'],
					},
					compatibility: {
						conflicts: ['replace conflict'],
						credential_rebind_targets: [{ target_type: 'channel_key' }],
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
		expect(screen.getByTestId('backup-replace-prune-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-import-remaining-migration-title')).toHaveTextContent('导入补强项');
		expect(screen.getByTestId('backup-import-remaining-migration-summary')).toHaveTextContent('默认收起，按需查看仍需手动处理的迁移能力。');
		expect(screen.getByTestId('backup-import-remaining-migration-trigger')).toBeInTheDocument();
		fireEvent.click(screen.getByTestId('backup-import-remaining-migration-trigger'));
		expect(screen.getByTestId('backup-import-remaining-migration-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-import-remaining-migration-section-trigger-0')).toBeInTheDocument();
		fireEvent.click(screen.getByTestId('backup-import-remaining-migration-section-trigger-0'));
		expect(screen.getByTestId('backup-import-remaining-migration-section-panel-0')).toBeInTheDocument();
		fireEvent.click(screen.getByTestId('backup-replace-prune-trigger'));
		expect(screen.getByTestId('backup-replace-prune-section-channels')).toBeInTheDocument();
		expect(screen.getByTestId('backup-replace-prune-section-title-channels')).toHaveTextContent('待删除渠道');
		expect(screen.getByTestId('backup-replace-prune-section-item-channels-0')).toHaveTextContent('legacy-channel');
		expect(screen.getByTestId('backup-replace-prune-section-apiKeys')).toBeInTheDocument();
		expect(screen.getByTestId('backup-replace-prune-section-title-apiKeys')).toHaveTextContent('待删除 API 密钥');
		expect(screen.getByTestId('backup-replace-prune-section-item-apiKeys-0')).toHaveTextContent('client-key');
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
		mocks.importSnapshotsState.data = [{
			snapshot_name: 'snapshot-selected-lock',
			snapshot_path: 'snapshots/snapshot-selected-lock.json',
			imported_at: '2026-04-21T10:00:00Z',
			size_bytes: 2048,
			is_latest: true,
		}];
		mocks.previewRollbackMutateAsyncMock.mockResolvedValue({
			snapshot_name: 'snapshot-selected-lock',
			manifest: { encrypted: undefined, contains_secrets: true, schema_version: '10' },
			rows_summary: { channels: 1 },
			compatibility: { conflicts: ['channel conflict'], credential_rebind_targets: ['channel-key-1'] },
			preview_warnings: ['route preview needs manual review'],
		});
		mocks.rollbackImportSnapshotMutateAsyncMock.mockResolvedValue({ snapshot_name: 'snapshot-selected-lock' });

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
		expect(getHelpHintButtons().length).toBeGreaterThanOrEqual(9);
		fireEvent.click(within(historyItem).getByTestId('backup-history-preview-button'));
		await screen.findByText('Rollback preview');
		expect(screen.getByTestId('backup-rollback-preview-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-preview-header')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-preview-title')).toHaveTextContent('Rollback preview');
		expect(screen.getByTestId('backup-rollback-preview-name')).toHaveTextContent('snapshot-selected-lock');
		expect(screen.getByTestId('backup-rollback-preview-name')).toHaveAttribute('data-raw-value', 'snapshot-selected-lock');
		expect(screen.getByTestId('backup-rollback-preview-overview')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-preview-summary-grid')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-preview-summary-conflicts')).toHaveTextContent('Compatibility Conflicts：1');
		expect(screen.getByTestId('backup-rollback-preview-summary-conflicts')).toHaveAttribute('data-raw-value', '1');
		expect(screen.getByTestId('backup-rollback-preview-summary-rebinds')).toHaveTextContent('Credential Rebinds：1');
		expect(screen.getByTestId('backup-rollback-preview-summary-rebinds')).toHaveAttribute('data-raw-value', '1');
		expect(screen.getByTestId('backup-rollback-preview-summary-warnings')).toHaveTextContent('Preview Warnings：1');
		expect(screen.getByTestId('backup-rollback-preview-summary-warnings')).toHaveAttribute('data-raw-value', '1');
		expect(screen.getByTestId('backup-rollback-preview-meta-grid')).toBeInTheDocument();
		expect(screen.getByTestId('backup-rollback-preview-meta-scope')).toHaveTextContent('Rollback Scope：Unknown');
		expect(screen.getByTestId('backup-rollback-preview-meta-scope')).not.toHaveAttribute('data-raw-value');
		expect(screen.getByTestId('backup-rollback-preview-meta-encrypted')).toHaveTextContent('Encryption：Unknown');
		expect(screen.getByTestId('backup-rollback-preview-meta-encrypted')).not.toHaveAttribute('data-raw-value');
		expect(screen.getByTestId('backup-rollback-preview-meta-contains-secrets')).toHaveTextContent('Contains Credentials：Yes');
		expect(screen.getByTestId('backup-rollback-preview-meta-contains-secrets')).toHaveAttribute('data-raw-value', 'true');
		expect(screen.getByTestId('backup-rollback-preview-meta-schema-version')).toHaveTextContent('Schema Version：10');
		expect(screen.getByTestId('backup-rollback-preview-meta-schema-version')).toHaveAttribute('data-raw-value', '10');

		const previousConfirm = window.confirm;
		window.confirm = () => true;
		fireEvent.click(within(historyItem).getByTestId('backup-history-rollback-button'));
		await waitFor(() => {
		expect(mocks.rollbackImportSnapshotMutateAsyncMock).toHaveBeenCalledTimes(1);
		});
		window.confirm = previousConfirm;
		expect(mocks.toastSuccessMock).toHaveBeenCalledWith(expect.stringContaining('snapshot-selected-lock'));
		expect(screen.getByTestId('backup-advanced-pending-title')).toHaveTextContent('Advanced migration tooling still pending');
		expect(screen.getByTestId('backup-advanced-pending-summary')).toHaveTextContent('Collapsed by default. Open only when you need the still-manual migration gaps.');
		fireEvent.click(screen.getByTestId('backup-remaining-migration-trigger'));
		expect(screen.getByTestId('backup-remaining-migration-panel')).toBeInTheDocument();
		expect(screen.getByTestId('backup-remaining-migration-section-trigger-0')).toBeInTheDocument();
		expect(screen.queryByTestId('backup-remaining-migration-section-item-rollback-tooling-compare-workflow')).not.toBeInTheDocument();
		fireEvent.click(screen.getByTestId('backup-remaining-migration-section-trigger-0'));
		expect(screen.getByTestId('backup-remaining-migration-section-panel-0')).toBeInTheDocument();
		expect(screen.getByTestId('backup-remaining-migration-section-item-rollback-tooling-compare-workflow')).toBeInTheDocument();
		expect(screen.getByTestId('backup-remaining-migration-section-item-rollback-tooling-compare-workflow-label')).toHaveTextContent('Compare workflow');
		expect(screen.getByTestId('backup-remaining-migration-section-item-rollback-tooling-compare-workflow-text')).toHaveTextContent('Multi-snapshot compare workflow with richer diff navigation beyond the current snapshot history list and preview panel.');
	});
});
