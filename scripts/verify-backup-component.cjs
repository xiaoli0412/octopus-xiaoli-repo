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
	const textarea = container.querySelector('textarea');
	if (!(textarea instanceof HTMLTextAreaElement)) throw new Error('Model mappings textarea not found');
	return textarea;
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
	assert(document.querySelector('[data-testid="backup-remaining-migration-trigger"]'), 'Expected remaining migration outer trigger');
	clickAccordionByTestId('backup-remaining-migration-trigger', fireEvent);
	assert(document.querySelector('[data-testid="backup-remaining-migration-panel"]'), 'Expected remaining migration panel selector after opening remaining migration section');
	assert(document.querySelector('[data-testid="backup-remaining-migration-section-trigger-0"]'), 'Expected first remaining migration section trigger selector after opening remaining migration section');
	clickAccordionByTestId('backup-remaining-migration-section-trigger-0', fireEvent);
	assert(document.querySelector('[data-testid="backup-remaining-migration-section-panel-0"]'), 'Expected first remaining migration section panel selector after opening remaining migration section');
	assert(document.querySelector('[data-testid="backup-remaining-migration-section-item-rollback-tooling-compare-workflow"]'), 'Expected first remaining migration item selector after opening remaining migration section');
	assert(document.querySelector('[data-testid="backup-remaining-migration-section-item-rollback-tooling-compare-workflow-label"]'), 'Expected first remaining migration item label selector after opening remaining migration section');
	assert(document.querySelector('[data-testid="backup-remaining-migration-section-item-rollback-tooling-compare-workflow-text"]'), 'Expected first remaining migration item text selector after opening remaining migration section');
	assert(getHelpHintButtons().length >= 9, 'Expected rollback view to keep backup help-hint buttons visible');
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
	assert(previewRollbackCalls.length === 1, 'Expected one rollback preview call');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-scope"]')?.textContent?.includes('Rollback Scope：Unknown'), 'Expected rollback preview scope summary selector content');
	assert(!document.querySelector('[data-testid="backup-rollback-preview-meta-scope"]')?.hasAttribute('data-raw-value'), 'Expected rollback preview scope raw-value attribute to stay absent when scopes are unknown');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-encrypted"]')?.textContent?.includes('Encryption：Unknown'), 'Expected rollback preview encrypted summary selector content');
	assert(!document.querySelector('[data-testid="backup-rollback-preview-meta-encrypted"]')?.hasAttribute('data-raw-value'), 'Expected rollback preview encrypted raw-value attribute to stay absent when encryption is unknown');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-contains-secrets"]')?.textContent?.includes('Contains Credentials：Yes'), 'Expected rollback preview contains-secrets summary selector content');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-contains-secrets"]')?.getAttribute('data-raw-value') === 'true', 'Expected rollback preview contains-secrets raw-value attribute');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-schema-version"]')?.textContent?.includes('Schema Version：10'), 'Expected rollback preview schema-version summary selector content');
	assert(document.querySelector('[data-testid="backup-rollback-preview-meta-schema-version"]')?.getAttribute('data-raw-value') === '10', 'Expected rollback preview schema-version raw-value attribute');

	const previousConfirm = global.window.confirm;
	global.window.confirm = () => true;
	fireEvent.click(document.querySelector('[data-testid="backup-history-item-snapshot-1"] [data-testid="backup-history-rollback-button"]'));
	await waitFor(() => {
		if (rollbackImportCalls.length !== 1) throw new Error('Expected one rollback apply call');
	});
	global.window.confirm = previousConfirm;
	assert(toastSuccessCalls.some((message) => String(message).includes('snapshot-1')), 'Expected rollback success toast containing snapshot name');
	assert(document.querySelector('[data-testid="backup-advanced-pending-title"]')?.textContent?.includes('Advanced migration tooling still pending'), 'Expected advanced-pending title selector content after opening snapshot history');
	assert(document.querySelector('[data-testid="backup-advanced-pending-summary"]')?.textContent?.includes('Collapsed by default. Open only when you need the still-manual migration gaps.'), 'Expected advanced-pending summary selector content after opening snapshot history');

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
	assert(getTextarea(zhView.container).placeholder !== 'legacy-model=gpt-4o\nvision-model=gpt-4.1', 'Expected zh-Hans map placeholder without English visible sample text');
	fireEvent.click(getByRoleName(screen, 'button', ['展开']));
	assert(document.querySelector('[data-testid="backup-import-remaining-migration-trigger"]'), 'Expected import remaining migration outer trigger');
	assert(document.querySelector('[data-testid="backup-import-remaining-migration-title"]')?.textContent?.includes('导入补强项'), 'Expected import remaining migration title selector content');
	assert(document.querySelector('[data-testid="backup-import-remaining-migration-summary"]')?.textContent?.includes('默认收起，按需查看仍需手动处理的迁移能力。'), 'Expected import remaining migration summary selector content');
	assert(!getVisibleText(zhView.container).includes('导入工具补强'), 'Expected remaining migration section titles to stay collapsed by default');
	clickAccordionByTestId('backup-import-remaining-migration-trigger', fireEvent);
	assert(Array.from(document.querySelectorAll('[data-slot="accordion-trigger"]')).some((node) => node.textContent?.includes('导入工具补强')), 'Expected remaining migration tooling summary section title after expanding outer section');
	assert(!getVisibleText(zhView.container).includes('替换导入 / 映射导入'), 'Expected zh-Hans detailed pending copy to stay collapsed by default');
	clickAccordionByText(screen, fireEvent, '导入工具补强');
	const zhHansVisibleText = getVisibleText(zhView.container);
	assert(zhHansVisibleText.includes('替换导入 / 映射导入'), 'Expected zh-Hans remaining migration tooling copy to use localized replace/map wording');
	assert(zhHansVisibleText.includes('快照模型=当前模型'), 'Expected zh-Hans remaining migration tooling copy to use localized mapping example wording');
	assert(!/replace\/map/i.test(zhHansVisibleText), 'Expected zh-Hans visible backup copy to avoid raw replace/map text');
	assert(!/\bremap\b/i.test(zhHansVisibleText), 'Expected zh-Hans visible backup copy to avoid raw remap text');
	cleanup();

	resetState();
	setLocale('en');
	const { SettingBackup: MapBackup } = jiti('../web/src/components/modules/setting/Backup.tsx');
	const mapFile = new File(['{"snapshot":true}'], 'snapshot-map.json', { type: 'application/json' });
	const view = render(React.createElement(MapBackup));
	await selectImportMode(screen, fireEvent, waitFor, 'map');
	assert(document.querySelector('[data-testid="backup-map-preview-root"]'), 'Expected map-preview root selector in map mode');
	assert(getHelpHintButtons().length >= 9, 'Expected map mode to keep backup help-hint buttons visible');
	fireEvent.change(getTextarea(view.container), { target: { value: 'legacy-model=gpt-4o\nmissing-model=gpt-4.1-mini\nunused-model=gpt-4.1' } });
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
	assert(document.querySelector('[data-testid="backup-compatibility-signal-list"]')?.textContent?.includes('Compatibility report found 1 missing mapping targets.'), 'Expected missing-mapping signal selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-signal-list"]')?.textContent?.includes('Compatibility report found 1 unused model mappings.'), 'Expected unused-mapping signal selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-missing-providers"]'), 'Expected missing-providers detail selector in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-missing-providers-title"]')?.textContent?.includes('Missing Providers / Channels'), 'Expected missing-providers detail title selector content in map mode');
	assert(document.querySelector('[data-testid="backup-compatibility-missing-providers-item-0"]')?.textContent?.includes('legacy-provider'), 'Expected missing-providers detail item selector content in map mode');
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
	assert(document.querySelector('[data-testid="backup-replace-prune-panel"]'), 'Expected replace-prune panel selector in replace mode');
	assert(document.querySelector('[data-testid="backup-import-remaining-migration-trigger"]'), 'Expected import remaining-migration outer trigger selector in replace mode');
	assert(document.querySelector('[data-testid="backup-import-remaining-migration-title"]')?.textContent?.includes('Import migration tooling'), 'Expected import remaining migration title selector content in replace mode');
	assert(document.querySelector('[data-testid="backup-import-remaining-migration-summary"]')?.textContent?.includes('Collapsed by default. Open only when you need the still-manual migration gaps.'), 'Expected import remaining migration summary selector content in replace mode');
	clickAccordionByTestId('backup-import-remaining-migration-trigger', fireEvent);
	assert(document.querySelector('[data-testid="backup-import-remaining-migration-panel"]'), 'Expected import remaining-migration panel selector in replace mode');
	assert(document.querySelector('[data-testid="backup-import-remaining-migration-section-trigger-0"]'), 'Expected import remaining-migration section trigger selector in replace mode');
	clickAccordionByTestId('backup-import-remaining-migration-section-trigger-0', fireEvent);
	assert(document.querySelector('[data-testid="backup-import-remaining-migration-section-panel-0"]'), 'Expected import remaining-migration section panel selector in replace mode');
	assert(document.querySelector('[data-testid="backup-import-remaining-migration-section-item-import-tooling-conflict-handling"]'), 'Expected import remaining-migration item selector in replace mode');
	assert(document.querySelector('[data-testid="backup-import-remaining-migration-section-item-import-tooling-conflict-handling-label"]'), 'Expected import remaining-migration item label selector in replace mode');
	assert(document.querySelector('[data-testid="backup-import-remaining-migration-section-item-import-tooling-conflict-handling-text"]'), 'Expected import remaining-migration item text selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-title"]')?.textContent?.includes('Replace-prune preview'), 'Expected replace-prune preview section selector');
	assert(document.querySelector('[data-testid="backup-replace-prune-summary"]')?.textContent?.includes('records are hidden by default'), 'Expected replace-prune summary selector to stay collapsed before expanding');
	clickAccordionByTestId('backup-replace-prune-trigger', fireEvent);
	assert(document.querySelector('[data-testid="backup-replace-prune-section-channels"]'), 'Expected replace-prune channels section selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-title-channels"]'), 'Expected replace-prune channels title selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-item-channels-0"]'), 'Expected replace-prune channels item selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-apiKeys"]'), 'Expected replace-prune api-keys section selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-title-apiKeys"]'), 'Expected replace-prune api-keys title selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-item-apiKeys-0"]'), 'Expected replace-prune api-keys item selector in replace mode');
	assert(document.querySelector('[data-testid="backup-replace-prune-section-title-channels"]')?.textContent?.includes('Channels to delete'), 'Expected replace-prune channels section');
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
