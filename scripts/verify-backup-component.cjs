const path = require('node:path');
const Module = require('node:module');
const React = require('../web/node_modules/react');
const { JSDOM } = require('../web/node_modules/jsdom');
const { createJiti } = require('../web/node_modules/.pnpm/jiti@2.6.1/node_modules/jiti/lib/jiti.cjs');

const webRoot = path.resolve(__dirname, '..', 'web');
const originalLoad = Module._load;

const importDBState = { data: undefined };
const previewRollbackState = { data: undefined };
const importSnapshotsState = { data: [], isLoading: false, isError: false, isFetching: false };

const exportMutateAsyncCalls = [];
const importMutateAsyncCalls = [];
const previewRollbackCalls = [];
const rollbackImportCalls = [];
const rollbackLatestCalls = [];
const importSnapshotsRefetchCalls = [];
const toastSuccessCalls = [];
const toastErrorCalls = [];

global.__backupComponentVerifyState = {
	importDBState,
	previewRollbackState,
	importSnapshotsState,
	exportMutateAsyncCalls,
	importMutateAsyncCalls,
	previewRollbackCalls,
	rollbackImportCalls,
	rollbackLatestCalls,
	importSnapshotsRefetchCalls,
	toastSuccessCalls,
	toastErrorCalls,
};

const dom = new JSDOM('<!doctype html><html><body></body></html>', { url: 'http://localhost/' });
global.window = dom.window;
global.document = dom.window.document;
global.navigator = dom.window.navigator;
global.React = React;
global.HTMLElement = dom.window.HTMLElement;
global.HTMLInputElement = dom.window.HTMLInputElement;
global.HTMLTextAreaElement = dom.window.HTMLTextAreaElement;
global.File = dom.window.File;
global.Node = dom.window.Node;
global.MutationObserver = dom.window.MutationObserver;
global.getComputedStyle = dom.window.getComputedStyle.bind(dom.window);
global.requestAnimationFrame = global.requestAnimationFrame || ((callback) => setTimeout(() => callback(Date.now()), 0));
global.cancelAnimationFrame = global.cancelAnimationFrame || ((id) => clearTimeout(id));
dom.window.requestAnimationFrame = dom.window.requestAnimationFrame || global.requestAnimationFrame;
dom.window.cancelAnimationFrame = dom.window.cancelAnimationFrame || global.cancelAnimationFrame;
global.IS_REACT_ACT_ENVIRONMENT = true;

Object.getOwnPropertyNames(dom.window).forEach((key) => {
	if (!(key in global)) {
		global[key] = dom.window[key];
	}
});

Module._load = function patchedLoad(request, parent, isMain) {
	const normalized = typeof request === 'string' ? request.replace(/\\/g, '/') : '';
	if (request === 'next-intl' || normalized.includes('/next-intl/')) {
		return { useTranslations: () => (key) => key };
	}
	if (request === '@/components/common/Toast' || normalized.includes('/src/components/common/Toast')) {
		return {
			toast: {
				success: (message) => toastSuccessCalls.push(message),
				error: (message) => toastErrorCalls.push(message),
			},
		};
	}
	if (request === '@/api/endpoints/setting' || normalized.includes('/src/api/endpoints/setting')) {
		return require('./verify-backup-component.setting-mock.cjs');
	}
	if (request === '@/components/ui/select' || normalized.includes('/src/components/ui/select')) {
		return require('./verify-backup-component.select-mock.cjs');
	}
	return originalLoad(request, parent, isMain);
};

const jiti = createJiti(__filename, {
	jsx: true,
	moduleCache: false,
	alias: {
		'@': path.join(webRoot, 'src'),
		'next-intl': path.resolve(__dirname, 'verify-backup-component.next-intl-mock.cjs'),
		'@/components/common/Toast': path.resolve(__dirname, 'verify-backup-component.toast-mock.cjs'),
		'@/api/endpoints/setting': path.resolve(__dirname, 'verify-backup-component.setting-mock.cjs'),
		'@/components/ui/select': path.resolve(__dirname, 'verify-backup-component.select-mock.cjs'),
	},
});

function setLocale(locale) {
	const { useSettingStore } = jiti('../web/src/stores/setting.ts');
	useSettingStore.setState({ locale });
}

function resetState() {
	importDBState.data = undefined;
	previewRollbackState.data = undefined;
	importSnapshotsState.data = [];
	importSnapshotsState.isLoading = false;
	importSnapshotsState.isError = false;
	importSnapshotsState.isFetching = false;
	exportMutateAsyncCalls.length = 0;
	importMutateAsyncCalls.length = 0;
	previewRollbackCalls.length = 0;
	rollbackImportCalls.length = 0;
	rollbackLatestCalls.length = 0;
	importSnapshotsRefetchCalls.length = 0;
	toastSuccessCalls.length = 0;
	toastErrorCalls.length = 0;
}

function assert(condition, message) {
	if (!condition) throw new Error(message);
}

function getHelpHintButtons(container = document.body) {
	return Array.from(container.querySelectorAll('button[data-slot="help-hint-trigger"], button[aria-label="View help"], button[aria-label="ariaLabel"]'));
}

function getFileInput(container) {
	const input = container.querySelector('input[type="file"]');
	if (!(input instanceof HTMLInputElement)) throw new Error('Backup file input not found');
	return input;
}

function getTextarea(container) {
	const textarea = container.querySelector('[data-testid="backup-model-mappings-textarea"]');
	if (!(textarea instanceof HTMLTextAreaElement)) throw new Error('Model mappings textarea not found');
	return textarea;
}

function setStructuredMappingRows(fireEvent, rows) {
	rows.forEach((row, index) => {
		if (index > 0) {
			fireEvent.click(document.querySelector('[data-testid="backup-structured-mapping-add"]'));
		}
		fireEvent.change(document.querySelector(`[data-testid="backup-structured-mapping-source-${index}"]`), { target: { value: row.source } });
		fireEvent.change(document.querySelector(`[data-testid="backup-structured-mapping-target-${index}"]`), { target: { value: row.target } });
	});
}

function getSwitchForLabel(screen, within, label) {
	const labelNode = screen.getByText(label);
	let row = labelNode.parentElement;
	while (row) {
		if (row.querySelectorAll('[role="switch"]').length === 1) {
			return within(row).getByRole('switch');
		}
		row = row.parentElement;
	}
	throw new Error(`Switch row for ${label} not found`);
}

function getByRoleName(screen, role, names) {
	for (const name of names) {
		const match = screen.queryByRole(role, { name });
		if (match) return match;
		const textMatches = screen.queryAllByText(name);
		for (const textMatch of textMatches) {
			const fallback = textMatch.closest(role === 'button' ? 'button,[role="button"]' : `[role="${role}"]`);
			if (fallback instanceof HTMLElement) return fallback;
		}
	}
	throw new Error(`Unable to find ${role} with names: ${names.map((item) => item.toString()).join(', ')}`);
}

async function selectImportMode(screen, fireEvent, waitFor, value) {
	const comboboxes = Array.from(document.querySelectorAll('[role="combobox"]'));
	for (const combobox of comboboxes) {
		if (!(combobox instanceof HTMLSelectElement)) continue;
		const optionValues = Array.from(combobox.options).map((option) => option.value);
		if (!['incremental', 'map', 'merge', 'replace', 'skip'].every((option) => optionValues.includes(option))) continue;
		fireEvent.change(combobox, { target: { value } });
		await waitFor(() => {
			if (combobox.value !== value) throw new Error(`Import mode did not switch to ${value}`);
		});
		return;
	}
	throw new Error('Import mode select not found');
}

function getConfirmSwitch(screen, within) {
	const marker = screen.getByText((text) => [
		'我已经检查上方风险提示，确认可以把这次导入应用到当前项目。',
		'I reviewed the risks above and want to apply this import to the current project.',
	].includes(text));
	let row = marker.parentElement;
	while (row) {
		if (row.querySelectorAll('[role="switch"]').length === 1) {
			return within(row).getByRole('switch');
		}
		row = row.parentElement;
	}
	throw new Error('Apply confirmation switch not found');
}

function getConfirmSwitchBySelector(screen) {
	const toggle = screen.queryByTestId('backup-apply-confirm-switch');
	if (!toggle) {
		throw new Error('Apply confirmation switch not found');
	}
	return toggle;
}

function getImportButton(screen) {
	const button = screen.queryByTestId('backup-import-button');
	if (!button) {
		throw new Error('Backup import button not found');
	}
	return button;
}

function getVisibleText(container) {
	const clone = container.cloneNode(true);
	for (const hidden of clone.querySelectorAll('.sr-only, [aria-hidden="true"]')) {
		hidden.remove();
	}
	return clone.textContent || '';
}

function clickAccordionByText(screen, fireEvent, text) {
	const trigger = Array.from(document.querySelectorAll('[data-slot="accordion-trigger"]')).find((node) => node.textContent?.includes(text));
	if (!(trigger instanceof HTMLElement)) {
		throw new Error(`Accordion trigger for ${text} not found`);
	}
	fireEvent.click(trigger);
}

function clickAccordionByTestId(testId, fireEvent) {
	const trigger = document.querySelector(`[data-testid="${testId}"]`);
	if (!(trigger instanceof HTMLElement)) {
		throw new Error(`Accordion trigger ${testId} not found`);
	}
	fireEvent.click(trigger);
}

async function verifyExportFlow({ render, screen, fireEvent, waitFor, cleanup, within }) {
	resetState();
	setLocale('en');
	const { SettingBackup } = jiti('../web/src/components/modules/setting/Backup.tsx');
	render(React.createElement(SettingBackup));
	assert(document.querySelector('[data-testid="backup-page"]'), 'Expected backup page root test id');
	assert(getHelpHintButtons().length >= 8, 'Expected backup default view to keep key help-hint buttons visible');

	fireEvent.click(getByRoleName(screen, 'button', ['Download JSON']));
	await waitFor(() => {
		if (exportMutateAsyncCalls.length !== 1) throw new Error('Expected one export call');
	});
	assert(exportMutateAsyncCalls[0].include_secrets === true, 'Default export should include plaintext credentials');
	assert(exportMutateAsyncCalls[0].include_logs === false, 'Default export should disable logs');
	assert(exportMutateAsyncCalls[0].include_stats === false, 'Default export should disable stats');
	assert(toastSuccessCalls.length === 1, 'Expected one export success toast');

	cleanup();
	resetState();
	setLocale('en');
	render(React.createElement(SettingBackup));
	fireEvent.click(getSwitchForLabel(screen, within, 'Include plaintext credentials in the snapshot'));
	assert(screen.queryAllByText('Redacted credentials').length >= 1, 'Expected redacted export badge after disabling secrets');
	fireEvent.click(getByRoleName(screen, 'button', ['Download JSON']));
	await waitFor(() => {
		if (exportMutateAsyncCalls.length !== 1) throw new Error('Expected one redacted export call');
	});
	assert(exportMutateAsyncCalls[0].include_secrets === false, 'Redacted export should disable secrets');

	cleanup();
}

async function verifyDryRunApplyAndRollback({ render, screen, fireEvent, waitFor, cleanup, within }) {
	resetState();
	setLocale('en');
	const { SettingBackup } = jiti('../web/src/components/modules/setting/Backup.tsx');
	importSnapshotsState.data = [{ snapshot_name: 'snapshot-1', snapshot_path: 'snapshots/snapshot-1.json', imported_at: '2026-04-21T05:00:00Z', size_bytes: 2048, is_latest: true }];
	const file = new File(['{"snapshot":true}'], 'snapshot.json', { type: 'application/json' });

	const view = render(React.createElement(SettingBackup));
	assert(document.querySelector('[data-testid="backup-page"]'), 'Expected backup page root test id');
	fireEvent.click(getSwitchForLabel(screen, within, 'Selective import'));
	fireEvent.click(getSwitchForLabel(screen, within, 'Routing'));
	fireEvent.change(getFileInput(view.container), { target: { files: [file] } });
	fireEvent.click(getImportButton(screen));

	await screen.findByTestId('backup-pending-apply-ready');
	assert(document.querySelector('[data-testid="backup-pending-apply-panel"]'), 'Expected pending-apply panel selector after dry-run');
	assert(document.querySelector('[data-testid="backup-pending-apply-ready"]'), 'Expected pending-apply ready selector after dry-run');
	assert(document.querySelector('[data-testid="backup-import-result-panel"]'), 'Expected import result panel selector after dry-run');
	assert(document.querySelector('[data-testid="backup-import-summary-grid"]'), 'Expected import summary grid selector after dry-run');
	assert(document.querySelector('[data-testid="backup-compatibility-panel"]'), 'Expected compatibility panel selector after dry-run');
	assert(document.querySelector('[data-testid="backup-compatibility-overview"]'), 'Expected compatibility overview selector after dry-run');
	assert(document.querySelector('[data-testid="backup-compatibility-toggle"]'), 'Expected compatibility toggle selector after dry-run');
	assert(document.querySelector('[data-testid="backup-compatibility-details"]'), 'Expected compatibility details selector after dry-run');
	assert(document.querySelector('[data-testid="backup-apply-confirm-switch"]'), 'Expected apply confirmation switch selector after dry-run');
	assert(document.querySelector('[data-testid="backup-pending-apply-meta-grid"]'), 'Expected pending-apply meta grid selector after dry-run');
	assert(importMutateAsyncCalls.length === 1, 'Expected one dry-run import call');
	assert(importMutateAsyncCalls[0].dryRun === true, 'First import call should be dry-run');
	assert(importMutateAsyncCalls[0].previewToken === undefined, 'Dry-run should not send preview token');
	assert(document.querySelector('[data-testid="backup-pending-apply-meta-grid"]')?.textContent?.includes('preview-token-1'), 'Expected captured preview token summary');

	const applyButton = screen.getByTestId('backup-apply-same-import-button');
	assert(applyButton.disabled, 'Apply button should be disabled before confirmation');
	assert(document.querySelector('[data-testid="backup-apply-confirm-panel"]'), 'Expected apply confirmation panel selector after dry-run');
	fireEvent.click(getConfirmSwitchBySelector(screen));
	await waitFor(() => {
		if (applyButton.disabled) throw new Error('Apply button stayed disabled after confirmation');
	});

	fireEvent.click(applyButton);
	await waitFor(() => {
		if (!document.querySelector('[data-testid="backup-post-import-validation-panel"]')) {
			throw new Error('Post-import validation selector did not appear after apply');
		}
	});
	assert(document.querySelector('[data-testid="backup-post-import-validation-panel"]'), 'Expected post-import validation selector after apply');
	assert(document.querySelector('[data-testid="backup-post-import-validation-summary-grid"]'), 'Expected post-import validation summary grid selector after apply');
	assert(document.querySelector('[data-testid="backup-post-import-validation-summary-degraded-groups"]')?.textContent?.includes('Degraded groups：1'), 'Expected degraded-groups summary selector content after apply');
	assert(document.querySelector('[data-testid="backup-post-import-validation-summary-empty-groups"]')?.textContent?.includes('Empty groups：0'), 'Expected empty-groups summary selector content after apply');
	assert(document.querySelector('[data-testid="backup-post-import-validation-summary-disabled-channels"]')?.textContent?.includes('Disabled channels：0'), 'Expected disabled-channels summary selector content after apply');
	assert(document.querySelector('[data-testid="backup-post-import-validation-summary-channels-without-keys"]')?.textContent?.includes('Channels without keys：0'), 'Expected channels-without-keys summary selector content after apply');
	assert(document.querySelector('[data-testid="backup-post-import-validation-summary-stale-items-removed"]')?.textContent?.includes('Stale items removed：0'), 'Expected stale-items-removed summary selector content after apply');
	assert(document.querySelector('[data-testid="backup-post-import-validation-summary-route-warnings"]')?.textContent?.includes('Route warnings：0'), 'Expected route-warnings summary selector content after apply');
	assert(document.querySelector('[data-testid="backup-post-import-validation-summary-price-rule-warnings"]')?.textContent?.includes('Price-rule warnings：0'), 'Expected price-rule-warnings summary selector content after apply');
	assert(document.querySelector('[data-testid="backup-post-import-validation-summary-alias-mappings"]')?.textContent?.includes('Alias mappings：0'), 'Expected alias-mappings summary selector content after apply');
	assert(document.querySelector('[data-testid="backup-post-import-validation-summary-alias-warnings"]')?.textContent?.includes('Alias warnings：0'), 'Expected alias-warnings summary selector content after apply');
	assert(document.querySelector('[data-testid="backup-post-import-health-summary"]'), 'Expected post-import health summary selector after apply');
	assert(document.querySelector('[data-testid="backup-post-import-health-summary-grid"]'), 'Expected post-import health summary grid selector after apply');
	assert(document.querySelector('[data-testid="backup-post-import-health-summary-targets"]')?.textContent?.includes('Health-check targets：3'), 'Expected post-import health targets selector content after apply');
	assert(document.querySelector('[data-testid="backup-post-import-health-summary-passed"]')?.textContent?.includes('Passed：2'), 'Expected post-import health passed selector content after apply');
	assert(importMutateAsyncCalls.length === 2, 'Expected second import call for apply');
	assert(importMutateAsyncCalls[1].dryRun === false, 'Apply should execute a real import');
	assert(importMutateAsyncCalls[1].previewToken === 'preview-token-1', 'Apply should reuse preview token');

	fireEvent.click(screen.getByTestId('backup-history-trigger'));
	assert(document.querySelector('[data-testid="backup-history-panel"]'), 'Expected backup history panel selector after opening snapshot history');
	assert(document.querySelector('[data-testid="backup-page"]'), 'Expected backup page root selector after opening snapshot history');
	assert(document.querySelector('[data-testid="backup-history-list"]'), 'Expected backup history list selector after opening snapshot history');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"]'), 'Expected snapshot history item selector after opening snapshot history');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-actions"]'), 'Expected snapshot history action-group selector after opening snapshot history');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-meta"]'), 'Expected snapshot history meta selector after opening snapshot history');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-name"]'), 'Expected snapshot history name selector after opening snapshot history');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-path"]'), 'Expected snapshot history path selector after opening snapshot history');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-size"]'), 'Expected snapshot history size selector after opening snapshot history');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-imported-at"]'), 'Expected snapshot history imported-at selector after opening snapshot history');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-latest-badge"]'), 'Expected snapshot history latest-badge selector after opening snapshot history');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-name"]').textContent.includes('snapshot-1'), 'Expected snapshot history name content');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-path"]').textContent.includes('snapshots/snapshot-1.json'), 'Expected snapshot history path content');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-path"]')?.getAttribute('data-raw-value') === 'snapshots/snapshot-1.json', 'Expected snapshot history path raw-value attribute');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-size"]')?.textContent?.includes('Size：2 KB'), 'Expected snapshot history size content');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-size"]')?.getAttribute('data-size-bytes') === '2048', 'Expected snapshot history size-bytes attribute');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-imported-at"]').textContent.includes('2026'), 'Expected snapshot history imported-at content');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-imported-at"]')?.getAttribute('data-raw-value') === '2026-04-21T05:00:00Z', 'Expected snapshot history imported-at raw-value attribute');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-latest-badge"]').textContent.includes('Latest'), 'Expected snapshot history latest badge content');
	assert(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-item-latest-badge"]')?.getAttribute('data-is-latest') === 'true', 'Expected snapshot history latest-badge attribute');
	assert(document.querySelector('[data-testid="backup-rollback-scope-editor"]'), 'Expected rollback scope editor selector after opening snapshot history');
	assert(document.querySelector('[data-testid="backup-rollback-scope-editor-title"]')?.textContent?.includes('Rollback domains'), 'Expected rollback scope editor title selector content');
	assert(document.querySelector('[data-testid="backup-rollback-scope-current-summary"]')?.textContent?.includes('Rollback Scope：Full snapshot restore'), 'Expected full-restore rollback scope summary before enabling selective rollback');
	assert(!document.querySelector('[data-testid="backup-remaining-migration-trigger"]'), 'Expected stale rollback pending trigger to stay hidden once rollback-domain editing is wired');
	assert(!document.querySelector('[data-testid="backup-remaining-migration-panel"]'), 'Expected stale rollback pending panel to stay hidden once rollback-domain editing is wired');
	assert(getHelpHintButtons().length >= 9, 'Expected rollback view to keep backup help-hint buttons visible');
	fireEvent.click(document.querySelector('[data-testid="backup-rollback-selective-switch"]'));
	assert(document.querySelector('[data-testid="backup-rollback-scope-grid"]'), 'Expected rollback scope grid selector after enabling selective rollback');
	fireEvent.click(document.querySelector('[data-testid="backup-rollback-scope-routing"]'));
	fireEvent.click(document.querySelector('[data-testid="backup-rollback-scope-stats"]'));
	fireEvent.click(document.querySelector('[data-testid="backup-rollback-scope-logs"]'));
	assert(document.querySelector('[data-testid="backup-rollback-scope-current-summary"]')?.textContent?.includes('Rollback Scope：Models, API keys, Settings'), 'Expected narrowed rollback scope summary after disabling routing/stats/logs');
	fireEvent.click(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-preview-button"]'));
	await screen.findByText('Rollback preview');
	assert(document.querySelector('[data-testid="backup-rollback-preview-panel"]'), 'Expected rollback preview panel selector after previewing a snapshot');
	assert(document.querySelector('[data-testid="backup-rollback-preview-header"]'), 'Expected rollback preview header selector after previewing a snapshot');
	assert(document.querySelector('[data-testid="backup-rollback-preview-title"]')?.textContent?.includes('Rollback preview'), 'Expected rollback preview title selector content');
	assert(document.querySelector('[data-testid="backup-rollback-preview-name"]')?.textContent?.includes('snapshot-1'), 'Expected rollback preview snapshot-name selector content');
	assert(document.querySelector('[data-testid="backup-rollback-preview-name"]')?.getAttribute('data-raw-value') === 'snapshot-1', 'Expected rollback preview snapshot-name raw-value attribute');
	assert(document.querySelector('[data-testid="backup-rollback-preview-overview"]'), 'Expected rollback preview overview selector after previewing a snapshot');
	assert(document.querySelector('[data-testid="backup-rollback-preview-summary-grid"]'), 'Expected rollback preview summary-grid selector after previewing a snapshot');
	assert(document.querySelector('[data-testid="backup-rollback-preview-summary-conflicts"]')?.textContent?.includes('Compatibility Conflicts：1'), 'Expected rollback preview conflicts summary selector content');
	assert(document.querySelector('[data-testid="backup-rollback-preview-summary-conflicts"]')?.getAttribute('data-raw-value') === '1', 'Expected rollback preview conflicts raw-value attribute');
	assert(document.querySelector('[data-testid="backup-rollback-preview-summary-rebinds"]')?.textContent?.includes('Credential Rebinds：1'), 'Expected rollback preview rebind summary selector content');
	assert(document.querySelector('[data-testid="backup-rollback-preview-summary-rebinds"]')?.getAttribute('data-raw-value') === '1', 'Expected rollback preview rebind raw-value attribute');
	assert(document.querySelector('[data-testid="backup-rollback-preview-summary-warnings"]')?.textContent?.includes('Preview Warnings：1'), 'Expected rollback preview warnings summary selector content');
	assert(document.querySelector('[data-testid="backup-rollback-preview-summary-warnings"]')?.getAttribute('data-raw-value') === '1', 'Expected rollback preview warnings raw-value attribute');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-grid"]'), 'Expected rollback preview meta-grid selector after previewing a snapshot');
	assert(document.querySelector('[data-testid="backup-rollback-route-diff-panel"]'), 'Expected rollback route-diff compare panel selector after previewing a snapshot');
	assert(document.querySelector('[data-testid="backup-rollback-route-diff-row-title-0"]')?.textContent?.includes('group-a / gpt-4o'), 'Expected rollback route-diff row title selector content');
	assert(document.querySelector('[data-testid="backup-rollback-route-diff-current-0"]')?.textContent?.includes('current-primary:gpt-4o'), 'Expected rollback route-diff current state selector content');
	assert(document.querySelector('[data-testid="backup-rollback-route-diff-snapshot-0"]')?.textContent?.includes('snapshot-primary:gpt-4o'), 'Expected rollback route-diff snapshot state selector content');
	assert(document.querySelector('[data-testid="backup-rollback-route-diff-added-0"]')?.textContent?.includes('snapshot-primary:gpt-4o'), 'Expected rollback route-diff added selector content');
	assert(document.querySelector('[data-testid="backup-rollback-route-diff-removed-0"]')?.textContent?.includes('current-primary:gpt-4o'), 'Expected rollback route-diff removed selector content');
	assert(previewRollbackCalls.length === 1, 'Expected one rollback preview call');
	assert(previewRollbackCalls[0]?.snapshotName === 'snapshot-1', 'Expected rollback preview payload to keep snapshot name');
	assert(JSON.stringify(previewRollbackCalls[0]?.importScopes) === JSON.stringify({ routing: false, models: true, api_keys: true, settings: true, stats: false, logs: false }), 'Expected first rollback preview payload to carry narrowed rollback scopes');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-scope"]')?.textContent?.includes('Rollback Scope：Models, API keys, Settings'), 'Expected rollback preview scope summary selector content for narrowed scopes');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-scope"]')?.getAttribute('data-raw-value') === 'models,api_keys,settings', 'Expected rollback preview scope raw-value attribute for narrowed scopes');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-encrypted"]')?.textContent?.includes('Encryption：Unknown'), 'Expected rollback preview encrypted summary selector content');
	assert(!document.querySelector('[data-testid="backup-rollback-preview-meta-encrypted"]')?.hasAttribute('data-raw-value'), 'Expected rollback preview encrypted raw-value attribute to stay absent when encryption is unknown');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-contains-secrets"]')?.textContent?.includes('Contains Credentials：Yes'), 'Expected rollback preview contains-secrets summary selector content');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-contains-secrets"]')?.getAttribute('data-raw-value') === 'true', 'Expected rollback preview contains-secrets raw-value attribute');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-schema-version"]')?.textContent?.includes('Schema Version：10'), 'Expected rollback preview schema-version summary selector content');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-schema-version"]')?.getAttribute('data-raw-value') === '10', 'Expected rollback preview schema-version raw-value attribute');
	assert(document.querySelector('[data-testid="backup-rollback-signal-panel"]'), 'Expected rollback signal panel selector after previewing a snapshot');
	assert(document.querySelector('[data-testid="backup-rollback-signal-title"]')?.textContent?.includes('Recommended Rollback Steps'), 'Expected rollback signal title selector content');
	assert(document.querySelector('[data-testid="backup-rollback-signal-summary"]')?.textContent?.includes('Start with the summary signals'), 'Expected rollback signal summary selector content');
	assert(document.querySelector('[data-testid="backup-rollback-signal-list"]')?.textContent?.includes('Rollback preview emitted 1 warnings.'), 'Expected rollback signal list to include preview warnings summary');
	assert(document.querySelector('[data-testid="backup-rollback-signal-list"]')?.textContent?.includes('Rollback preview found 1 conflicts.'), 'Expected rollback signal list to include conflicts summary');
	assert(document.querySelector('[data-testid="backup-rollback-signal-list"]')?.textContent?.includes('Channel-key credential rebind is required for 1 restored targets.'), 'Expected rollback signal list to include rebind summary');
	assert(document.querySelector('[data-testid="backup-rollback-signal-list"]')?.textContent?.includes('Rollback diagnostics marked 5 current records for removal or reset.'), 'Expected rollback signal list to include structured replace-prune summary');
	assert(document.querySelector('[data-testid="backup-rollback-guidance"]'), 'Expected rollback guidance detail selector after previewing a snapshot');
	assert(document.querySelector('[data-testid="backup-rollback-guidance-title"]')?.textContent?.includes('Recommended Rollback Steps'), 'Expected rollback guidance title selector content');
	assert(document.querySelector('[data-testid="backup-rollback-guidance-item-0"]')?.textContent?.includes('Resolve rollback risks first'), 'Expected rollback guidance blocking-risk item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-guidance-item-4"]')?.textContent?.includes('Review which current records rollback removes'), 'Expected rollback guidance replace-prune item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-guidance-item-5"]')?.textContent?.includes('Review post-rollback route and policy drift'), 'Expected rollback guidance route-and-policy item selector content after replace-prune guidance');
	assert(document.querySelector('[data-testid="backup-rollback-preview-warnings-panel"]'), 'Expected rollback preview warnings panel selector after previewing a snapshot');
	assert(document.querySelector('[data-testid="backup-rollback-preview-warnings-title"]')?.textContent?.includes('Rollback Preview Warnings'), 'Expected rollback preview warnings title selector content');
	assert(document.querySelector('[data-testid="backup-rollback-preview-warnings-list-item-0"]')?.textContent?.includes('route preview needs manual review'), 'Expected rollback preview warnings detail selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-panel"]'), 'Expected rollback compatibility detail panel selector after previewing a snapshot');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-title"]')?.textContent?.includes('Rollback compatibility details'), 'Expected rollback compatibility detail title selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-conflicts-item-0"]')?.textContent?.includes('channel conflict'), 'Expected rollback compatibility conflicts item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-alias-conflicts-item-0"]')?.textContent?.includes('alias conflict: rollback-vision -> gpt-4.1'), 'Expected rollback compatibility alias-conflict item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-route-conflicts-item-0"]')?.textContent?.includes('route conflict: rollback-group -> legacy-model'), 'Expected rollback compatibility route-conflict item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-affected-groups-item-0"]')?.textContent?.includes('group-a'), 'Expected rollback compatibility affected-groups item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-affected-channels-item-0"]')?.textContent?.includes('Primary'), 'Expected rollback compatibility affected-channels item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-missing-providers-item-0"]')?.textContent?.includes('rollback-provider'), 'Expected rollback compatibility missing-provider item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-missing-models-item-0"]')?.textContent?.includes('legacy-model'), 'Expected rollback compatibility missing-model item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-base-url-mismatches-item-0"]')?.textContent?.includes('rollback-channel'), 'Expected rollback compatibility base-url mismatch item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-schema-mismatches-item-0"]')?.textContent?.includes('snapshot schema:v2 differs'), 'Expected rollback compatibility schema-mismatch item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-skipped-targets-item-0"]')?.textContent?.includes('channel_key:101 empty credential'), 'Expected rollback compatibility skipped-target item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-credential-rebind-item-0"]')?.textContent?.includes('Primary'), 'Expected rollback compatibility credential-rebind item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-replace-pruned-channels-title"]')?.textContent?.includes('Current channels removed by rollback'), 'Expected rollback replace-pruned channels title selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-replace-pruned-channels-item-0"]')?.textContent?.includes('current-channel-a'), 'Expected rollback replace-pruned channels item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-replace-pruned-groups-item-0"]')?.textContent?.includes('current-group-a'), 'Expected rollback replace-pruned groups item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-replace-pruned-settings-item-0"]')?.textContent?.includes('proxy_url'), 'Expected rollback replace-pruned settings item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-replace-pruned-llm-infos-item-0"]')?.textContent?.includes('current-legacy-model'), 'Expected rollback replace-pruned llm-info item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-replace-pruned-api-keys-item-0"]')?.textContent?.includes('current-client-key'), 'Expected rollback replace-pruned api-key item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-invalid-route-targets-item-0"]')?.textContent?.includes('missing_target'), 'Expected rollback compatibility invalid-route-target item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-skipped-route-targets-item-0"]')?.textContent?.includes('review mapping'), 'Expected rollback compatibility skipped-route-target item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-route-preview-warnings-item-0"]')?.textContent?.includes('rollback route may degrade'), 'Expected rollback compatibility route-preview-warning item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-mapping-preview-item-0"]')?.textContent?.includes('legacy-model'), 'Expected rollback compatibility mapping-preview item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-alias-preview-item-0"]')?.textContent?.includes('rollback-vision'), 'Expected rollback compatibility alias-preview item selector content');
	assert(document.querySelector('[data-testid="backup-rollback-compatibility-model-policy-diffs-item-0"]')?.textContent?.includes('legacy-model'), 'Expected rollback compatibility model-policy item selector content');
	fireEvent.click(document.querySelector('[data-testid="backup-rollback-scope-settings"]'));
	assert(!document.querySelector('[data-testid="backup-rollback-preview-panel"]'), 'Expected rollback preview panel to clear after rollback scope change');
	assert(previewRollbackCalls.length === 1, 'Expected rollback scope change to invalidate preview without issuing another preview call');
	assert(document.querySelector('[data-testid="backup-rollback-scope-current-summary"]')?.textContent?.includes('Rollback Scope：Models, API keys'), 'Expected rollback scope summary after narrowing settings away');
	fireEvent.click(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-preview-button"]'));
	await screen.findByText('Rollback preview');
	assert(previewRollbackCalls.length === 2, 'Expected second rollback preview call after scope change');
	assert(JSON.stringify(previewRollbackCalls[1]?.importScopes) === JSON.stringify({ routing: false, models: true, api_keys: true, settings: false, stats: false, logs: false }), 'Expected second rollback preview payload to carry updated rollback scopes');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-scope"]')?.textContent?.includes('Rollback Scope：Models, API keys'), 'Expected rollback preview scope summary selector content after scope narrowing');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-scope"]')?.getAttribute('data-raw-value') === 'models,api_keys', 'Expected rollback preview scope raw-value attribute after scope narrowing');

	const previousConfirm = global.window.confirm;
	global.window.confirm = () => true;
	fireEvent.click(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-rollback-button"]'));
	await waitFor(() => {
		if (rollbackImportCalls.length !== 1) throw new Error('Expected one rollback apply call');
	});
	global.window.confirm = previousConfirm;
	assert(JSON.stringify(rollbackImportCalls[0]?.importScopes) === JSON.stringify({ routing: false, models: true, api_keys: true, settings: false, stats: false, logs: false }), 'Expected rollback apply payload to carry the latest rollback scopes');
	assert(toastSuccessCalls.some((message) => String(message).includes('snapshot-1')), 'Expected rollback success toast containing snapshot name');
	assert(!document.querySelector('[data-testid="backup-advanced-pending-title"]'), 'Expected stale advanced-pending title to stay hidden after rollback-domain editing closure');
	assert(!document.querySelector('[data-testid="backup-advanced-pending-summary"]'), 'Expected stale advanced-pending summary to stay hidden after rollback-domain editing closure');
	for (const scopeKey of ['models', 'api_keys'] ) {
		fireEvent.click(document.querySelector(`[data-testid="backup-rollback-scope-${scopeKey}"]`));
	}
	assert(document.querySelector('[data-testid="backup-rollback-scope-current-summary"]')?.textContent?.includes('Rollback Scope：Full snapshot restore'), 'Expected empty selective rollback scopes to fall back to full restore summary');
	assert(document.querySelector('[data-testid="backup-rollback-scope-mode-note"]')?.textContent?.includes('No rollback domains are selected. Preview and rollback will fall back to a full snapshot restore.'), 'Expected fallback note for empty selective rollback scopes');
	fireEvent.click(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-preview-button"]'));
	await screen.findByText('Rollback preview');
	assert(previewRollbackCalls.length === 3, 'Expected third rollback preview call for full-restore fallback');
	assert(previewRollbackCalls[2]?.importScopes === undefined, 'Expected empty selective rollback scopes to omit importScopes in preview payload');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-scope"]')?.textContent?.includes('Rollback Scope：Unknown'), 'Expected fallback rollback preview scope summary selector content');
	assert(!document.querySelector('[data-testid="backup-rollback-preview-meta-scope"]')?.hasAttribute('data-raw-value'), 'Expected fallback rollback preview scope raw-value attribute to stay absent when scopes are omitted');

	cleanup();
}

async function verifySelectiveImportGuard({ render, screen, fireEvent, cleanup, within }) {
	resetState();
	setLocale('en');
	const { SettingBackup } = jiti('../web/src/components/modules/setting/Backup.tsx');
	const file = new File(['{"snapshot":true}'], 'snapshot-empty-selective.json', { type: 'application/json' });
	const view = render(React.createElement(SettingBackup));
	fireEvent.click(getSwitchForLabel(screen, within, 'Selective import'));
	for (const label of ['Routing', 'Models', 'API keys', 'Settings', 'Stats', 'Relay logs']) {
		fireEvent.click(getSwitchForLabel(screen, within, label));
	}
	fireEvent.change(getFileInput(view.container), { target: { files: [file] } });
	const importButton = getImportButton(screen);
	assert(importButton.disabled, 'Import button should be disabled when no scopes are selected');
	assert(screen.getByText('Import stays disabled until at least one scope is selected.'), 'Expected selective-import warning');
	cleanup();
}

async function verifyMapAndReplaceFlows({ render, screen, fireEvent, waitFor, cleanup, within }) {
	resetState();
	setLocale('zh-Hans');
	const { SettingBackup } = jiti('../web/src/components/modules/setting/Backup.tsx');
	const zhView = render(React.createElement(SettingBackup));
	assert(document.querySelector('[data-testid="backup-page"]'), 'Expected backup page root test id');
	await selectImportMode(screen, fireEvent, waitFor, 'map');
	assert(document.querySelector('[data-testid="backup-structured-mapping-source-0"]')?.getAttribute('placeholder') === '旧模型=gpt-4o', 'Expected zh-Hans source placeholder');
	assert(document.querySelector('[data-testid="backup-structured-mapping-target-0"]')?.getAttribute('placeholder') === '视觉模型=gpt-4.1', 'Expected zh-Hans target placeholder');
	fireEvent.click(getByRoleName(screen, 'button', ['展开']));
	assert(!document.querySelector('[data-testid="backup-import-remaining-migration-trigger"]'), 'Expected stale import remaining-migration entry to stay hidden after compatibility closure');
	cleanup();

	resetState();
	setLocale('en');
	const { SettingBackup: MapBackup } = jiti('../web/src/components/modules/setting/Backup.tsx');
	const mapFile = new File(['{"snapshot":true}'], 'snapshot-map.json', { type: 'application/json' });
	const view = render(React.createElement(MapBackup));
	await selectImportMode(screen, fireEvent, waitFor, 'map');
	assert(document.querySelector('[data-testid="backup-map-preview-root"]'), 'Expected map-preview root selector in map mode');
	assert(document.querySelector('[data-testid="backup-structured-mapping-panel"]'), 'Expected structured mapping panel selector in map mode');
	assert(document.querySelector('[data-testid="backup-structured-mapping-empty"]'), 'Expected structured mapping empty hint before rows are filled');
	assert(getHelpHintButtons().length >= 9, 'Expected map mode to keep backup help-hint buttons visible');
	setStructuredMappingRows(fireEvent, [
		{ source: 'legacy-model', target: 'gpt-4o' },
		{ source: 'missing-model', target: 'gpt-4.1-mini' },
		{ source: 'unused-model', target: 'gpt-4.1' },
	]);
	assert(document.querySelector('[data-testid="backup-structured-mapping-count"]')?.textContent?.includes('Active Mappings：3'), 'Expected structured mapping count after row edits');
	assert(getTextarea(view.container).value === 'legacy-model=gpt-4o\nmissing-model=gpt-4.1-mini\nunused-model=gpt-4.1', 'Expected hidden model-mappings payload to stay line-based');
	fireEvent.change(getFileInput(view.container), { target: { files: [mapFile] } });
	fireEvent.click(getImportButton(screen));
	await screen.findByTestId('backup-pending-apply-ready');
	assert(document.querySelector('[data-testid="backup-pending-apply-panel"]'), 'Expected pending-apply panel selector in map mode');
	assert(document.querySelector('[data-testid="backup-pending-apply-ready"]'), 'Expected pending-apply ready selector in map mode');
	assert(document.querySelector('[data-testid="backup-pending-apply-meta-grid"]'), 'Expected pending-apply meta grid selector in map mode');
	assert(document.querySelector('[data-testid="backup-import-result-panel"]'), 'Expected import result panel selector in map mode');
	assert(document.querySelector('[data-testid="backup-import-summary-grid"]'), 'Expected import summary grid selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-panel"]'), 'Expected compatibility panel selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-overview"]'), 'Expected compatibility overview selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-toggle"]'), 'Expected compatibility toggle selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-details"]'), 'Expected compatibility details selector in map mode');
	assert(importMutateAsyncCalls.length === 1, 'Expected one map dry-run call');
	assert(importMutateAsyncCalls[0].mode === 'map', 'Expected map mode import');
	assert(document.querySelector('[data-testid="backup-compatibility-summary"]')?.textContent?.includes('Detailed diagnostics stay collapsed until you need them.'), 'Expected import diagnostics summary selector to stay collapsed by default');
	assert(!document.querySelector('[data-testid="backup-compatibility-mapping-preview-title"]'), 'Expected model mapping previews selector to stay hidden before expanding diagnostics');
	fireEvent.click(screen.getByTestId('backup-compatibility-toggle'));
	assert(document.querySelector('[data-testid="backup-compatibility-signal-list"]'), 'Expected compatibility signal list selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-signal-list"]')?.textContent?.includes('Compatibility report found 1 missing providers.'), 'Expected missing-provider signal selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-signal-list"]')?.textContent?.includes('Channel-key credential rebind is required for 1 imported targets.'), 'Expected channel-key rebind signal selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-signal-list"]')?.textContent?.includes('Compatibility report skipped 1 route-target previews.'), 'Expected skipped-route-preview signal selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-signal-list"]')?.textContent?.includes('Compatibility report found 1 missing mapping targets.'), 'Expected missing-mapping signal selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-signal-list"]')?.textContent?.includes('Compatibility report found 1 unused model mappings.'), 'Expected unused-mapping signal selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-guidance"]'), 'Expected compatibility guidance selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-guidance-title"]')?.textContent?.includes('Recommended Next Steps'), 'Expected compatibility guidance title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-guidance-item-0"]')?.textContent?.includes('Resolve blocking risks first'), 'Expected blocking-risk guidance selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-guidance-item-1"]')?.textContent?.includes('Restore missing providers or models'), 'Expected missing-target guidance selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-guidance-item-2"]')?.textContent?.includes('Prepare credential rebinds'), 'Expected credential-rebind guidance selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-guidance-item-3"]')?.textContent?.includes('Review which targets are skipped'), 'Expected skipped-target guidance selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-guidance-item-4"]')?.textContent?.includes('Fix model mappings before apply'), 'Expected model-mapping guidance selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-guidance-item-5"]')?.textContent?.includes('Review route and policy drift'), 'Expected route/policy guidance selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-conflicts"]'), 'Expected compatibility-conflicts detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-conflicts-title"]')?.textContent?.includes('Compatibility Conflicts'), 'Expected compatibility-conflicts detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-conflicts-item-0"]')?.textContent?.includes('channel conflict'), 'Expected compatibility-conflicts detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-alias-conflicts"]'), 'Expected alias-conflicts detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-alias-conflicts-title"]')?.textContent?.includes('Alias Conflicts'), 'Expected alias-conflicts detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-alias-conflicts-item-0"]')?.textContent?.includes('alias conflict: legacy-vision -> gpt-4.1'), 'Expected alias-conflicts detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-route-conflicts"]'), 'Expected route-conflicts detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-route-conflicts-title"]')?.textContent?.includes('Route Conflicts'), 'Expected route-conflicts detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-route-conflicts-item-0"]')?.textContent?.includes('route conflict: group-a -> legacy-model'), 'Expected route-conflicts detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-affected-groups"]'), 'Expected affected-groups detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-affected-groups-title"]')?.textContent?.includes('Affected Groups'), 'Expected affected-groups detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-affected-groups-item-0"]')?.textContent?.includes('group-a'), 'Expected affected-groups detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-affected-channels"]'), 'Expected affected-channels detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-affected-channels-title"]')?.textContent?.includes('Affected Channels'), 'Expected affected-channels detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-affected-channels-item-0"]')?.textContent?.includes('Primary'), 'Expected affected-channels detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-missing-providers"]'), 'Expected missing-providers detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-missing-providers-title"]')?.textContent?.includes('Missing Providers / Channels'), 'Expected missing-providers detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-missing-providers-item-0"]')?.textContent?.includes('legacy-provider'), 'Expected missing-providers detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-missing-models"]'), 'Expected missing-models detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-missing-models-title"]')?.textContent?.includes('Missing Models'), 'Expected missing-models detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-missing-models-item-0"]')?.textContent?.includes('legacy-text-preview'), 'Expected missing-models detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-base-url-mismatches"]'), 'Expected base-url-mismatches detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-base-url-mismatches-title"]')?.textContent?.includes('Base-URL Mismatches'), 'Expected base-url-mismatches detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-base-url-mismatches-item-0"]')?.textContent?.includes('preview-channel'), 'Expected base-url-mismatches detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-schema-mismatches"]'), 'Expected schema-mismatches detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-schema-mismatches-title"]')?.textContent?.includes('Schema Mismatches'), 'Expected schema-mismatches detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-schema-mismatches-item-0"]')?.textContent?.includes('snapshot schema:v2 differs'), 'Expected schema-mismatches detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-skipped-targets"]'), 'Expected skipped-targets detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-skipped-targets-title"]')?.textContent?.includes('Preserved / Skipped Targets'), 'Expected skipped-targets detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-skipped-targets-item-0"]')?.textContent?.includes('channel_key:201 empty credential'), 'Expected skipped-targets credential detail selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-skipped-targets-item-1"]')?.textContent?.includes('setting:api_base_url existing row preserved by skip mode'), 'Expected skipped-targets skip-mode detail selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-credential-rebind"]'), 'Expected credential-rebind detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-credential-rebind-title"]')?.textContent?.includes('Credential Rebind Targets'), 'Expected credential-rebind detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-credential-rebind-item-0"]')?.textContent?.includes('Primary'), 'Expected credential-rebind detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-invalid-route-targets"]'), 'Expected invalid-route-target detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-invalid-route-targets-title"]')?.textContent?.includes('Route Target Risks'), 'Expected invalid-route-target detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-invalid-route-targets-item-0"]')?.textContent?.includes('missing_target'), 'Expected invalid-route-target detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-skipped-route-targets"]'), 'Expected skipped-route-target detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-skipped-route-targets-title"]')?.textContent?.includes('Skipped Route Previews'), 'Expected skipped-route-target detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-skipped-route-targets-item-0"]')?.textContent?.includes('skipped_preview'), 'Expected skipped-route-target detail issue selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-skipped-route-targets-item-0"]')?.textContent?.includes('review mapping'), 'Expected skipped-route-target detail action selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-route-preview-warnings"]'), 'Expected route-preview-warning detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-route-preview-warnings-title"]')?.textContent?.includes('Route Preview Warnings'), 'Expected route-preview-warning detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-route-preview-warnings-item-0"]')?.textContent?.includes('route may degrade'), 'Expected route-preview-warning detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-route-preview-diffs"]'), 'Expected route-preview-diff detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-route-preview-diffs-title"]')?.textContent?.includes('Route Preview Diffs'), 'Expected route-preview-diff detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-route-preview-diffs-item-0"]')?.textContent?.includes('group-a'), 'Expected route-preview-diff detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-mapping-preview"]'), 'Expected mapping-preview detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-mapping-preview-title"]')?.textContent?.includes('Model Mapping Previews'), 'Expected mapping-preview detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-mapping-preview-item-0"]')?.textContent?.includes('legacy-model'), 'Expected first mapping-preview detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-mapping-preview-item-1"]')?.textContent?.includes('missing-model'), 'Expected second mapping-preview detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-mapping-preview-item-2"]')?.textContent?.includes('unused-model'), 'Expected third mapping-preview detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-missing-mapping"]'), 'Expected missing-mapping detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-missing-mapping-title"]')?.textContent?.includes('Missing Mapping Targets'), 'Expected missing-mapping detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-missing-mapping-item-0"]')?.textContent?.includes('missing-model'), 'Expected missing-mapping detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-unused-mapping"]'), 'Expected unused-mapping detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-unused-mapping-title"]')?.textContent?.includes('Unused Model Mappings'), 'Expected unused-mapping detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-unused-mapping-item-0"]')?.textContent?.includes('unused-model'), 'Expected unused-mapping detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-alias-preview"]'), 'Expected alias-preview detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-alias-preview-title"]')?.textContent?.includes('Alias Preview Mappings'), 'Expected alias-preview detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-alias-preview-item-0"]')?.textContent?.includes('legacy-vision'), 'Expected alias-preview detail item selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-model-policy-diffs"]'), 'Expected model-policy-diffs detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-model-policy-diffs-title"]')?.textContent?.includes('Model Policy Diffs'), 'Expected model-policy-diffs detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-model-policy-diffs-item-0"]')?.textContent?.includes('legacy-model'), 'Expected model-policy-diffs detail item selector content in map mode');

	cleanup();
	resetState();
	setLocale('en');
	const { SettingBackup: ReplaceBackup } = jiti('../web/src/components/modules/setting/Backup.tsx');
	const replaceView = render(React.createElement(ReplaceBackup));
	const replaceFile = new File(['{"snapshot":true}'], 'snapshot-replace.json', { type: 'application/json' });
	await selectImportMode(screen, fireEvent, waitFor, 'replace');
	fireEvent.change(getFileInput(replaceView.container), { target: { files: [replaceFile] } });
	fireEvent.click(getImportButton(screen));
	await screen.findByTestId('backup-pending-apply-ready');
	assert(document.querySelector('[data-testid="backup-pending-apply-panel"]'), 'Expected pending-apply panel selector in replace mode');
	assert(document.querySelector('[data-testid="backup-pending-apply-ready"]'), 'Expected pending-apply ready selector in replace mode');
	assert(document.querySelector('[data-testid="backup-pending-apply-meta-grid"]'), 'Expected pending-apply meta grid selector in replace mode');
	assert(document.querySelector('[data-testid="backup-import-result-panel"]'), 'Expected import result panel selector in replace mode');
	assert(document.querySelector('[data-testid="backup-import-summary-grid"]'), 'Expected import summary grid selector in replace mode');
	assert(document.querySelector('[data-testid="backup-compatibility-panel"]'), 'Expected compatibility panel selector in replace mode');
	assert(document.querySelector('[data-testid="backup-compatibility-overview"]'), 'Expected compatibility overview selector in replace mode');
	assert(document.querySelector('[data-testid="backup-compatibility-toggle"]'), 'Expected compatibility toggle selector in replace mode');
	assert(document.querySelector('[data-testid="backup-compatibility-details"]'), 'Expected compatibility details selector in replace mode');
	fireEvent.click(screen.getByTestId('backup-compatibility-toggle'));
	assert(document.querySelector('[data-testid="backup-compatibility-signal-list"]'), 'Expected compatibility signal list selector in replace mode');
	assert(document.querySelector('[data-testid="backup-compatibility-signal-list"]')?.textContent?.includes('Replace mode can remove current project records that are not kept by the snapshot.'), 'Expected replace-mode risk signal selector content in replace mode');
	assert(document.querySelector('[data-testid="backup-compatibility-signal-list"]')?.textContent?.includes('Compatibility report found 1 conflicts.'), 'Expected replace-mode conflict signal selector content in replace mode');
	assert(document.querySelector('[data-testid="backup-compatibility-signal-list"]')?.textContent?.includes('Channel-key credential rebind is required for 1 imported targets.'), 'Expected replace-mode channel-key rebind signal selector content in replace mode');
	assert(document.querySelector('[data-testid="backup-compatibility-signal-list"]')?.textContent?.includes('Structured prune preview found 5 additional records.'), 'Expected replace-mode structured-prune signal selector content in replace mode');
	assert(document.querySelector('[data-testid="backup-compatibility-guidance-item-2"]')?.textContent?.includes('Review which current records replace mode removes'), 'Expected replace-prune guidance selector content in replace mode');
	assert(document.querySelector('[data-testid="backup-compatibility-replace-pruned-channels"]'), 'Expected replace-pruned channels detail selector in replace mode');
	assert(document.querySelector('[data-testid="backup-compatibility-replace-pruned-channels-title"]')?.textContent?.includes('Channels removed by replace mode'), 'Expected replace-pruned channels title selector in replace mode');
	assert(document.querySelector('[data-testid="backup-compatibility-replace-pruned-channels-item-0"]')?.textContent?.includes('legacy-channel'), 'Expected replace-pruned channels item selector in replace mode');
	assert(document.querySelector('[data-testid="backup-compatibility-replace-pruned-api-keys"]'), 'Expected replace-pruned api-keys detail selector in replace mode');
	assert(document.querySelector('[data-testid="backup-compatibility-replace-pruned-api-keys-title"]')?.textContent?.includes('API keys removed by replace mode'), 'Expected replace-pruned api-keys title selector in replace mode');
	assert(document.querySelector('[data-testid="backup-compatibility-replace-pruned-api-keys-item-0"]')?.textContent?.includes('client-key'), 'Expected replace-pruned api-keys item selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-panel"]'), 'Expected replace-prune panel selector in replace mode');
	assert(!document.querySelector('[data-testid="backup-import-remaining-migration-trigger"]'), 'Expected import remaining-migration entry to stay removed in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-title"]')?.textContent?.includes('Replace-prune preview'), 'Expected replace-prune preview section selector');
	assert(document.querySelector('[data-testid="backup-replace-prune-summary"]')?.textContent?.includes('records are hidden by default'), 'Expected replace-prune summary selector to stay collapsed before expanding');
	clickAccordionByTestId('backup-replace-prune-trigger', fireEvent);
	assert(document.querySelector('[data-testid="backup-replace-prune-section-channels"]'), 'Expected replace-prune channels section selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-title-channels"]'), 'Expected replace-prune channels title selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-item-channels-0"]'), 'Expected replace-prune channels item selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-groups"]'), 'Expected replace-prune groups section selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-title-groups"]'), 'Expected replace-prune groups title selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-item-groups-0"]'), 'Expected replace-prune groups item selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-settings"]'), 'Expected replace-prune settings section selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-title-settings"]'), 'Expected replace-prune settings title selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-item-settings-0"]'), 'Expected replace-prune settings item selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-models"]'), 'Expected replace-prune models section selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-title-models"]'), 'Expected replace-prune models title selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-item-models-0"]'), 'Expected replace-prune models item selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-apiKeys"]'), 'Expected replace-prune api-keys section selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-title-apiKeys"]'), 'Expected replace-prune api-keys title selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-item-apiKeys-0"]'), 'Expected replace-prune api-keys item selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-title-channels"]')?.textContent?.includes('Channels to delete'), 'Expected replace-prune channels section');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-title-groups"]')?.textContent?.includes('Groups to delete'), 'Expected replace-prune groups section');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-title-settings"]')?.textContent?.includes('Settings to reset'), 'Expected replace-prune settings section');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-title-models"]')?.textContent?.includes('Models to delete'), 'Expected replace-prune models section');
	assert(screen.getByTestId('backup-replace-prune-section-title-apiKeys').textContent?.includes('API keys to delete'), 'Expected replace-prune api-keys section');
	const applyButton = screen.getByTestId('backup-apply-same-import-button');
	assert(applyButton.disabled, 'Replace apply should require confirmation');
	assert(document.querySelector('[data-testid="backup-apply-confirm-panel"]'), 'Expected apply confirmation panel selector in replace mode');
	fireEvent.click(getConfirmSwitchBySelector(screen));
	await waitFor(() => {
		if (applyButton.disabled) throw new Error('Replace apply button stayed disabled after confirmation');
	});
	cleanup();
}

async function main() {
	const { cleanup, render, screen, fireEvent, waitFor, within } = require('../web/node_modules/@testing-library/react');
	await verifyExportFlow({ render, screen, fireEvent, waitFor, cleanup, within });
	await verifyDryRunApplyAndRollback({ render, screen, fireEvent, waitFor, cleanup, within });
	await verifySelectiveImportGuard({ render, screen, fireEvent, cleanup, within });
	await verifyMapAndReplaceFlows({ render, screen, fireEvent, waitFor, cleanup, within });
	console.log('backup-component verification passed');
}

main().catch((error) => {
	console.error(error && error.stack ? error.stack : error);
	process.exit(1);
}).finally(() => {
	Module._load = originalLoad;
	delete global.__backupComponentVerifyState;
	if (global.window?.close) global.window.close();
});
