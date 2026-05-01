import assert from 'node:assert/strict';
import { createRequire } from 'node:module';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');
const require = createRequire(import.meta.url);
const { createJiti } = require(path.join(repoRoot, 'web/node_modules/.pnpm/jiti@2.6.1/node_modules/jiti/lib/jiti.cjs'));
const jiti = createJiti(path.join(scriptDir, 'verify-backup-logic.mjs'), {
	moduleCache: false,
	alias: {
		'@': path.join(repoRoot, 'web', 'src'),
	},
});
const {
	buildCompatibilitySignalItems,
	getAliasPreviewItems,
	getApplySameImportGuardReason,
	getCompatibilityCounts,
	getCompatibilityOverview,
	getExportSnapshotPresentation,
	getImportResultPresentation,
	getMissingModelMappingItems,
	getModelMappingPreviewItems,
	getModelPolicyDiffItems,
	getPostImportValidationSummary,
	getRemainingMigrationToolingItems,
	getRemainingMigrationToolingSections,
	getUnusedModelMappingItems,
} = jiti(path.join(repoRoot, 'web/src/components/modules/setting/backup-logic.ts'));

const locale = 'en';

const counts = getCompatibilityCounts({
	credential_rebind_targets: [
		{ target_type: 'channel_key' },
		{ target_type: 'channel_key' },
		{ target_type: 'api_key' },
	],
	model_mapping_previews: [
		{ used: true, target_exists: true },
		{ used: true, target_exists: false },
		{ used: false, target_exists: true },
	],
	conflicts: ['conflict-a'],
	missing_models: ['model-a', 'model-b'],
	replace_pruned_channels: ['channel-a'],
	replace_pruned_api_keys: ['api-key-a', 'api-key-b'],
});

assert.equal(counts.conflicts, 1);
assert.equal(counts.credentialRebindTargets, 3);
assert.equal(counts.channelKeyRebindTargets, 2);
assert.equal(counts.apiKeyRebindTargets, 1);
assert.equal(counts.modelMappingPreviews, 3);
assert.equal(counts.usedModelMappings, 2);
assert.equal(counts.unusedModelMappings, 1);
assert.equal(counts.missingMappingTargets, 1);
assert.equal(counts.missingModels, 2);
assert.equal(counts.replacePrunedTargets, 3);

const rollbackOverview = getCompatibilityOverview({
	kind: 'rollback',
	warningsCount: 0,
	locale,
	counts: {
		conflicts: 0,
		aliasConflicts: 0,
		routeConflicts: 1,
		credentialRebindTargets: 0,
		channelKeyRebindTargets: 0,
		apiKeyRebindTargets: 0,
		invalidRouteTargets: 0,
		skippedRouteTargetPreviews: 0,
		routePreviewWarnings: 0,
		routePreviewDiffs: 0,
		missingProviders: 0,
		missingModels: 0,
		baseURLMismatches: 0,
		schemaMismatches: 1,
		skippedTargets: 0,
		modelMappingPreviews: 0,
		usedModelMappings: 0,
		unusedModelMappings: 0,
		missingMappingTargets: 0,
		aliasPreviewMappings: 0,
		modelPolicyDiffs: 0,
		replacePrunedTargets: 0,
	},
});

assert.equal(rollbackOverview.tone, 'danger');
assert.equal(rollbackOverview.title, 'Rollback has blocking risks');
assert.match(rollbackOverview.description, /route or schema risks/i);

const applyRiskItems = buildCompatibilitySignalItems({
	kind: 'import',
	locale,
	effectiveMode: 'replace',
	includeReplaceModeRisk: true,
	includeWarningsCount: false,
	includeModelMappingPreviews: false,
	includeUnusedModelMappings: false,
	replacePruneCount: 4,
	counts: {
		conflicts: 2,
		aliasConflicts: 0,
		routeConflicts: 0,
		credentialRebindTargets: 1,
		channelKeyRebindTargets: 1,
		apiKeyRebindTargets: 0,
		invalidRouteTargets: 0,
		skippedRouteTargetPreviews: 0,
		routePreviewWarnings: 0,
		routePreviewDiffs: 0,
		missingProviders: 0,
		missingModels: 0,
		baseURLMismatches: 0,
		schemaMismatches: 0,
		skippedTargets: 0,
		modelMappingPreviews: 0,
		usedModelMappings: 0,
		unusedModelMappings: 0,
		missingMappingTargets: 0,
		aliasPreviewMappings: 0,
		modelPolicyDiffs: 0,
		replacePrunedTargets: 0,
	},
});

assert.deepEqual(applyRiskItems, [
	'Replace mode can remove current project records that are not kept by the snapshot.',
	'Replace-prune preview will delete or reset 4 current records.',
	'Compatibility report found 2 conflicts.',
	'Channel-key credential rebind is required for 1 imported targets.',
]);

const importWarningItems = buildCompatibilitySignalItems({
	kind: 'import',
	locale,
	warningsCount: 2,
	counts: {
		conflicts: 0,
		aliasConflicts: 0,
		routeConflicts: 0,
		credentialRebindTargets: 0,
		channelKeyRebindTargets: 0,
		apiKeyRebindTargets: 0,
		invalidRouteTargets: 0,
		skippedRouteTargetPreviews: 0,
		routePreviewWarnings: 0,
		routePreviewDiffs: 0,
		missingProviders: 0,
		missingModels: 0,
		baseURLMismatches: 0,
		schemaMismatches: 0,
		skippedTargets: 0,
		modelMappingPreviews: 0,
		usedModelMappings: 0,
		unusedModelMappings: 0,
		missingMappingTargets: 0,
		aliasPreviewMappings: 0,
		modelPolicyDiffs: 0,
		replacePrunedTargets: 0,
	},
});

assert.deepEqual(importWarningItems, ['Import report emitted 2 warnings.']);

assert.equal(getApplySameImportGuardReason({ hasPendingApplyRequest: false, previewToken: 'token', requiresConfirm: false, confirmed: false }), 'missing_request');
assert.equal(getApplySameImportGuardReason({ hasPendingApplyRequest: true, previewToken: '   ', requiresConfirm: false, confirmed: false }), 'missing_preview_token');
assert.equal(getApplySameImportGuardReason({ hasPendingApplyRequest: true, previewToken: 'token', requiresConfirm: true, confirmed: false }), 'confirm_required');
assert.equal(getApplySameImportGuardReason({ hasPendingApplyRequest: true, previewToken: 'token', requiresConfirm: true, confirmed: true }), null);

const validationSummary = getPostImportValidationSummary({
	degraded_groups: ['group-a'],
	empty_groups: ['group-b', 'group-c'],
	disabled_channels: [],
	channels_without_keys: ['channel-a'],
	stale_items_removed: ['item-a'],
	route_warnings: ['warning-a', 'warning-b'],
	price_rule_warnings: ['drift-a'],
	alias_mappings: ['alias-a'],
	alias_warnings: ['alias-warning-a', 'alias-warning-b'],
});

assert.deepEqual(validationSummary, {
	degradedGroups: 1,
	emptyGroups: 2,
	disabledChannels: 0,
	channelsWithoutKeys: 1,
	staleItemsRemoved: 1,
	routeWarnings: 2,
	priceRuleWarnings: 1,
	aliasMappings: 1,
	aliasWarnings: 2,
});

assert.deepEqual(getImportResultPresentation({ resultIsDryRun: true, modeLabel: 'Replace', locale }), {
	title: 'Dry-run report',
	description: 'Dry-run report only analyzes compatibility and route-impact hints. It does not write to the database.',
	modeLabel: 'Dry-run mode used',
});

const appliedPresentation = getImportResultPresentation({ resultIsDryRun: false, modeLabel: 'Merge', locale });
assert.equal(appliedPresentation.title, 'Import applied');
assert.equal(appliedPresentation.modeLabel, 'Applied mode used');
assert.match(appliedPresentation.description, /Import applied in Merge mode/i);

const fullExportPresentation = getExportSnapshotPresentation({ includeSecrets: true, includeLogs: false, includeStats: true, locale });
assert.equal(fullExportPresentation.toggleLabel, 'Include plaintext credentials in the snapshot');
assert.deepEqual(fullExportPresentation.scopeBadges, ['Project snapshot', 'Channels / groups / routing', 'Plaintext credentials', 'Stats']);

const redactedExportPresentation = getExportSnapshotPresentation({ includeSecrets: false, includeLogs: true, includeStats: false, locale });
assert.deepEqual(redactedExportPresentation.scopeBadges, ['Project snapshot', 'Channels / groups / routing', 'Redacted credentials', 'Relay logs']);
assert.match(redactedExportPresentation.summary, /redacted project snapshot/i);

const zhHansFullExportPresentation = getExportSnapshotPresentation({ includeSecrets: true, includeLogs: false, includeStats: true, locale: 'zh-Hans' });
assert.deepEqual(zhHansFullExportPresentation.scopeBadges, ['项目快照', '渠道 / 分组 / 路由', '明文凭证', '统计数据']);

const zhHansRedactedExportPresentation = getExportSnapshotPresentation({ includeSecrets: false, includeLogs: true, includeStats: false, locale: 'zh-Hans' });
assert.deepEqual(zhHansRedactedExportPresentation.scopeBadges, ['项目快照', '渠道 / 分组 / 路由', '脱敏凭证', '中继日志']);
const remainingMigrationToolingItems = getRemainingMigrationToolingItems(locale);
assert.deepEqual(remainingMigrationToolingItems, [
	{ key: 'conflict-handling', label: 'Conflict handling', text: 'Guided conflict handling for richer replace/map edge cases beyond the current structured preview.' },
	{ key: 'mapping-editor', label: 'Mapping editor', text: 'Richer model-mapping editor beyond the current line-based remap input.' },
	{ key: 'compare-workflow', label: 'Compare workflow', text: 'Multi-snapshot compare workflow with richer diff navigation beyond the current snapshot history list and preview panel.' },
	{ key: 'rollback-domains', label: 'Rollback domains', text: 'Granular rollback-domain editing beyond the current full snapshot restore and selective-scope override flow.' },
	{ key: 'route-diff', label: 'Route diff', text: 'Side-by-side route diff tooling beyond the current compact summary cards and detail lists.' },
]);
assert.deepEqual(getRemainingMigrationToolingSections(locale), [
	{
		key: 'import-tooling',
		title: 'Import tooling',
		summary: 'These gaps still need guided import conflict resolution and remap editing.',
		items: remainingMigrationToolingItems.slice(0, 2),
	},
	{
		key: 'rollback-tooling',
		title: 'Rollback tooling',
		summary: 'These gaps still need richer snapshot recovery and compare navigation.',
		items: remainingMigrationToolingItems.slice(2, 4),
	},
	{
		key: 'route-analysis',
		title: 'Route analysis',
		summary: 'This gap still needs richer side-by-side route diff inspection.',
		items: remainingMigrationToolingItems.slice(4),
	},
]);

const zhHansRemainingMigrationToolingItems = getRemainingMigrationToolingItems('zh-Hans');
assert.equal(zhHansRemainingMigrationToolingItems[0]?.label, '冲突处理');
assert.match(zhHansRemainingMigrationToolingItems[0]?.text ?? '', /替换导入 \/ 映射导入/);
assert.doesNotMatch(zhHansRemainingMigrationToolingItems[0]?.text ?? '', /replace\/map/i);
assert.match(zhHansRemainingMigrationToolingItems[1]?.text ?? '', /快照模型=当前模型/);
assert.doesNotMatch(zhHansRemainingMigrationToolingItems[1]?.text ?? '', /\bremap\b/i);

const zhHantRemainingMigrationToolingItems = getRemainingMigrationToolingItems('zh-Hant');
assert.equal(zhHantRemainingMigrationToolingItems[0]?.label, '衝突處理');
assert.match(zhHantRemainingMigrationToolingItems[0]?.text ?? '', /替換導入 \/ 映射導入/);
assert.doesNotMatch(zhHantRemainingMigrationToolingItems[0]?.text ?? '', /replace\/map/i);
assert.match(zhHantRemainingMigrationToolingItems[1]?.text ?? '', /快照模型=目前模型/);
assert.doesNotMatch(zhHantRemainingMigrationToolingItems[1]?.text ?? '', /\bremap\b/i);

const zhHansRemainingMigrationToolingSections = getRemainingMigrationToolingSections('zh-Hans');
assert.equal(zhHansRemainingMigrationToolingSections[0]?.title, '导入工具补强');
assert.equal(zhHansRemainingMigrationToolingSections[0]?.summary, '这些缺口主要集中在导入时的冲突引导和模型映射编辑能力。');

const zhHantRemainingMigrationToolingSections = getRemainingMigrationToolingSections('zh-Hant');
assert.equal(zhHantRemainingMigrationToolingSections[0]?.title, '導入工具補強');
assert.equal(zhHantRemainingMigrationToolingSections[0]?.summary, '這些缺口主要集中在導入時的衝突引導與模型映射編輯能力。');
assert.deepEqual(getAliasPreviewItems([{ snapshot_model: 'legacy-model', current_model: 'gpt-4o', canonical: 'gpt-4o', contexts: ['routing', 'fallback'] }], locale), [
	'snapshot:legacy-model | current:gpt-4o | canonical:gpt-4o | contexts:routing, fallback',
]);
assert.deepEqual(getAliasPreviewItems([{ snapshot_model: 'legacy-model', current_model: 'gpt-4o', canonical: 'gpt-4o', contexts: ['channel:preview-channel', 'group:preview-group'] }], 'zh-Hans'), [
	'\u5feb\u7167\u6a21\u578b:legacy-model | \u5f53\u524d\u6a21\u578b:gpt-4o | \u89c4\u8303\u540d:gpt-4o | \u4f5c\u7528\u8303\u56f4:\u6e20\u9053:preview-channel\u3001\u5206\u7ec4:preview-group',
]);
assert.deepEqual(getModelMappingPreviewItems([{ source_model: 'legacy-model', target_model: 'gpt-4.1', usage_count: 2, used: true, target_exists: false, touched_fields: ['primary_model', 'fallback_model'], contexts: ['routing'], warnings: ['current model not found'] }], locale), [
	'snapshot:legacy-model | current:gpt-4.1 | usage:2 | used:yes | target:missing | fields:primary_model, fallback_model | contexts:routing | warnings:current model not found',
]);
assert.deepEqual(getModelMappingPreviewItems([{ source_model: 'legacy-model', target_model: 'gpt-4.1', usage_count: 3, used: true, target_exists: true, touched_fields: ['channels.model', 'group_items.model_name', 'api_keys.supported_models'], contexts: ['channel:mapped-channel', 'group_route:mapped-group', 'api_key:preview-client'] }], 'zh-Hans'), [
	'\u5feb\u7167\u6a21\u578b:legacy-model | \u5f53\u524d\u6a21\u578b:gpt-4.1 | \u5f15\u7528\u6b21\u6570:3 | \u662f\u5426\u4f7f\u7528:\u662f | \u76ee\u6807\u72b6\u6001:\u5b58\u5728 | \u53d7\u5f71\u54cd\u5b57\u6bb5:\u6e20\u9053.\u6a21\u578b\u3001\u5206\u7ec4\u6761\u76ee.\u6a21\u578b\u540d\u79f0\u3001API\u5bc6\u94a5.\u652f\u6301\u6a21\u578b | \u4f5c\u7528\u8303\u56f4:\u6e20\u9053:mapped-channel\u3001\u5206\u7ec4\u8def\u7531:mapped-group\u3001API\u5bc6\u94a5:preview-client',
]);
assert.deepEqual(getMissingModelMappingItems([{ source_model: 'legacy-model', target_model: 'gpt-4.1', used: true, target_exists: false, contexts: ['routing'], warnings: ['current model not found'] }], locale), [
	'snapshot:legacy-model | current:gpt-4.1 | contexts:routing | warnings:current model not found',
]);
assert.deepEqual(getUnusedModelMappingItems([{ source_model: 'unused-model', target_model: 'gpt-4.1', used: false, contexts: ['api_keys'] }], locale), [
	'snapshot:unused-model | current:gpt-4.1 | contexts:api_keys',
]);
assert.deepEqual(getModelPolicyDiffItems([{
	model: 'legacy-model',
	current_model: 'gpt-4.1',
	impact_level: 'high',
	changed_fields: ['billing_mode', 'probe_policy'],
	before: { billing_mode: 'paid', probe_policy: 'manual', probe_interval: 30, probe_concurrency: 1 },
	after: { billing_mode: 'free', probe_policy: 'auto', probe_interval: 60, probe_concurrency: 2 },
	contexts: ['routing'],
	warnings: [
		'billing_mode changed from paid to free',
		'probe_policy changed from manual to auto',
		'probe_interval changed from 30 to 60',
		'probe_concurrency changed from 1 to 2',
		'model:legacy-model concurrent probe/race may increase cost',
	],
}], 'zh-Hans'), [
	'\u6a21\u578b:legacy-model | \u5f53\u524d\u6a21\u578b:gpt-4.1 | \u5f71\u54cd\u7ea7\u522b:\u9ad8 | \u53d8\u66f4\u5b57\u6bb5:\u8ba1\u8d39\u6a21\u5f0f\u3001\u63a2\u6d4b\u7b56\u7565 | \u53d8\u66f4\u524d:\u8ba1\u8d39:\u4ed8\u8d39, \u63a2\u6d4b:\u624b\u52a8, \u95f4\u9694:30, \u5e76\u53d1:1 | \u53d8\u66f4\u540e:\u8ba1\u8d39:\u514d\u8d39, \u63a2\u6d4b:\u81ea\u52a8, \u95f4\u9694:60, \u5e76\u53d1:2 | \u4f5c\u7528\u8303\u56f4:\u8def\u7531 | \u8b66\u544a:\u8ba1\u8d39\u6a21\u5f0f\u4ece \u4ed8\u8d39 \u53d8\u4e3a \u514d\u8d39\u3001\u63a2\u6d4b\u7b56\u7565\u4ece \u624b\u52a8 \u53d8\u4e3a \u81ea\u52a8\u3001\u63a2\u6d4b\u95f4\u9694\u4ece 30 \u53d8\u4e3a 60\u3001\u63a2\u6d4b\u5e76\u53d1\u4ece 1 \u53d8\u4e3a 2\u3001\u6a21\u578b legacy-model \u7684\u5e76\u53d1\u63a2\u6d4b\u6216\u7ade\u901f\u53ef\u80fd\u589e\u52a0\u6210\u672c',
]);

assert.deepEqual(getModelMappingPreviewItems([{ source_model: 'legacy-model', target_model: 'gpt-4.1', usage_count: 2, used: true, target_exists: false, touched_fields: ['primary_model', 'fallback_model'], contexts: ['routing'], warnings: ['current model not found'] }], 'zh-Hans'), [
	'快照模型:legacy-model | 当前模型:gpt-4.1 | 引用次数:2 | 是否使用:是 | 目标状态:缺失 | 受影响字段:主模型、备用模型 | 作用范围:路由 | 警告:当前项目中未找到该模型',
]);

assert.deepEqual(getMissingModelMappingItems([{ source_model: 'legacy-model', target_model: 'gpt-4.1', used: true, target_exists: false, contexts: ['routing'], warnings: ['mapped target not found in current environment'] }], 'zh-Hans'), [
	'\u5feb\u7167\u6a21\u578b:legacy-model | \u5f53\u524d\u6a21\u578b:gpt-4.1 | \u4f5c\u7528\u8303\u56f4:\u8def\u7531 | \u8b66\u544a:\u5f53\u524d\u73af\u5883\u4e2d\u672a\u627e\u5230\u6620\u5c04\u76ee\u6807',
]);
assert.deepEqual(getUnusedModelMappingItems([{ source_model: 'unused-model', target_model: 'gpt-4.1', used: false, contexts: ['api_keys'], warnings: ['mapping source not referenced by selected import scopes'] }], 'zh-Hans'), [
	'\u5feb\u7167\u6a21\u578b:unused-model | \u5f53\u524d\u6a21\u578b:gpt-4.1 | \u4f5c\u7528\u8303\u56f4:API\u5bc6\u94a5 | \u8b66\u544a:\u8be5\u6620\u5c04\u6765\u6e90\u672a\u88ab\u6240\u9009\u5bfc\u5165\u8303\u56f4\u5f15\u7528',
]);

assert.deepEqual(getModelPolicyDiffItems([{
	model: 'legacy-model',
	current_model: 'gpt-4.1',
	impact_level: 'high',
	changed_fields: ['billing_mode', 'probe_policy'],
	before: { billing_mode: 'paid', probe_policy: 'manual', probe_interval: 30, probe_concurrency: 1 },
	after: { billing_mode: 'free', probe_policy: 'auto', probe_interval: 60, probe_concurrency: 2 },
	contexts: ['routing'],
	warnings: [
		'billing_mode changed from paid to free',
		'probe_policy changed from manual to auto',
		'probe_interval changed from 30 to 60',
		'probe_concurrency changed from 1 to 2',
		'model:legacy-model concurrent probe/race may increase cost',
	],
}], 'zh-Hans'), [
	'\u6a21\u578b:legacy-model | \u5f53\u524d\u6a21\u578b:gpt-4.1 | \u5f71\u54cd\u7ea7\u522b:\u9ad8 | \u53d8\u66f4\u5b57\u6bb5:\u8ba1\u8d39\u6a21\u5f0f\u3001\u63a2\u6d4b\u7b56\u7565 | \u53d8\u66f4\u524d:\u8ba1\u8d39:\u4ed8\u8d39, \u63a2\u6d4b:\u624b\u52a8, \u95f4\u9694:30, \u5e76\u53d1:1 | \u53d8\u66f4\u540e:\u8ba1\u8d39:\u514d\u8d39, \u63a2\u6d4b:\u81ea\u52a8, \u95f4\u9694:60, \u5e76\u53d1:2 | \u4f5c\u7528\u8303\u56f4:\u8def\u7531 | \u8b66\u544a:\u8ba1\u8d39\u6a21\u5f0f\u4ece \u4ed8\u8d39 \u53d8\u4e3a \u514d\u8d39\u3001\u63a2\u6d4b\u7b56\u7565\u4ece \u624b\u52a8 \u53d8\u4e3a \u81ea\u52a8\u3001\u63a2\u6d4b\u95f4\u9694\u4ece 30 \u53d8\u4e3a 60\u3001\u63a2\u6d4b\u5e76\u53d1\u4ece 1 \u53d8\u4e3a 2\u3001\u6a21\u578b legacy-model \u7684\u5e76\u53d1\u63a2\u6d4b\u6216\u7ade\u901f\u53ef\u80fd\u589e\u52a0\u6210\u672c',
]);

assert.deepEqual(getModelPolicyDiffItems([{
	model: 'legacy-model',
	current_model: 'gpt-4.1',
	impact_level: 'high',
	changed_fields: ['billing_mode', 'probe_policy'],
	before: { billing_mode: 'paid', probe_policy: 'manual', probe_interval: 30, probe_concurrency: 1 },
	after: { billing_mode: 'free', probe_policy: 'auto', probe_interval: 60, probe_concurrency: 2 },
	contexts: ['routing'],
	warnings: [
		'billing_mode changed from paid to free',
		'probe_policy changed from manual to auto',
		'probe_interval changed from 30 to 60',
		'probe_concurrency changed from 1 to 2',
		'model:legacy-model concurrent probe/race may increase cost',
	],
}], 'en'), [
	'model:legacy-model | current:gpt-4.1 | impact:high | changed:billing_mode, probe_policy | before:billing:paid, policy:manual, interval:30, concurrency:1 | after:billing:free, policy:auto, interval:60, concurrency:2 | contexts:routing | warnings:billing_mode changed from paid to free, probe_policy changed from manual to auto, probe_interval changed from 30 to 60, probe_concurrency changed from 1 to 2, model:legacy-model concurrent probe/race may increase cost',
]);

console.log('backup-logic verification passed');

