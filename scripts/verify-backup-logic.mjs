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
	buildImportCompatibilityGuidanceItems,
	getCompatibilityDiagnosticItems,
	getCompatibilityNameItems,
	getCredentialRebindTargetItems,
	getAliasPreviewItems,
	getApplySameImportGuardReason,
	getCompatibilityCounts,
	getCompatibilityOverview,
	getExportSnapshotPresentation,
	getImportResultPresentation,
	getMergedRoutePreviewWarningItems,
	getMissingModelMappingItems,
	getModelMappingPreviewItems,
	getModelPolicyDiffItems,
	getPostImportValidationSummary,
	getReplacePrunedBreakdownItems,
	getRoutePreviewDiffItems,
	getRoutePreviewWarningItems,
	getRouteTargetIssueItems,
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

const staleSummaryCounts = getCompatibilityCounts({
	summary: {
		conflicts: 0,
		used_model_mappings: 0,
		unused_model_mappings: 0,
		missing_mapping_targets: 0,
		route_preview_warnings: 0,
		replace_pruned_channels: 0,
		replace_pruned_groups: 0,
		replace_pruned_settings: 0,
		replace_pruned_llm_infos: 0,
		replace_pruned_api_keys: 0,
	},
	conflicts: ['conflict-a', 'conflict-b'],
	route_preview_warnings: ['route may degrade', 'route preview diffs: 2'],
	model_mapping_previews: [
		{ used: true, target_exists: true },
		{ used: true, target_exists: false },
		{ used: false, target_exists: true },
	],
	replace_pruned_channels: ['legacy-channel'],
	replace_pruned_groups: ['legacy-group'],
	replace_pruned_settings: ['proxy_url'],
	replace_pruned_llm_infos: ['legacy-model'],
	replace_pruned_api_keys: ['client-key'],
});

assert.equal(staleSummaryCounts.conflicts, 2);
assert.equal(staleSummaryCounts.usedModelMappings, 2);
assert.equal(staleSummaryCounts.unusedModelMappings, 1);
assert.equal(staleSummaryCounts.missingMappingTargets, 1);
assert.equal(staleSummaryCounts.routePreviewWarnings, 2);
assert.equal(staleSummaryCounts.replacePrunedTargets, 5);

const staleRebindSummaryCounts = getCompatibilityCounts({
	summary: {
		credential_rebind_targets: 1,
		channel_key_rebind_targets: 1,
		api_key_rebind_targets: 1,
	},
});

assert.equal(staleRebindSummaryCounts.credentialRebindTargets, 2);
assert.equal(staleRebindSummaryCounts.channelKeyRebindTargets, 1);
assert.equal(staleRebindSummaryCounts.apiKeyRebindTargets, 1);

const mergedRollbackWarnings = getMergedRoutePreviewWarningItems([
	['route preview needs manual review'],
	['rollback route may degrade', 'route preview needs manual review'],
], locale);

assert.deepEqual(mergedRollbackWarnings, [
	'route preview needs manual review',
	'rollback route may degrade',
]);

const mergedWarningRollbackGuidanceItems = buildImportCompatibilityGuidanceItems({
	kind: 'rollback',
	locale,
	counts: {
		conflicts: 0,
		aliasConflicts: 0,
		routeConflicts: 0,
		credentialRebindTargets: 0,
		channelKeyRebindTargets: 0,
		apiKeyRebindTargets: 0,
		invalidRouteTargets: 0,
		skippedRouteTargetPreviews: 1,
		routePreviewWarnings: 2,
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
	compatibility: {
		route_preview_warnings: ['rollback route may degrade'],
		skipped_route_target_previews: [{ issue_type: 'skipped_preview' }],
	},
});

assert.deepEqual(mergedWarningRollbackGuidanceItems, [
	{
		key: 'route-and-policy',
		tone: 'warning',
		title: 'Review post-rollback route and policy drift',
		detail: 'The rollback preview also surfaced 2 route warnings, 1 skipped route previews. Review the route and policy details below before restoring.',
	},
]);

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

const rollbackSignalItems = buildCompatibilitySignalItems({
	kind: 'rollback',
	locale,
	warningsCount: 1,
	counts: {
		conflicts: 1,
		aliasConflicts: 0,
		routeConflicts: 0,
		credentialRebindTargets: 1,
		channelKeyRebindTargets: 1,
		apiKeyRebindTargets: 0,
		invalidRouteTargets: 0,
		skippedRouteTargetPreviews: 0,
		routePreviewWarnings: 1,
		routePreviewDiffs: 0,
		missingProviders: 1,
		missingModels: 1,
		baseURLMismatches: 0,
		schemaMismatches: 1,
		skippedTargets: 1,
		modelMappingPreviews: 0,
		usedModelMappings: 0,
		unusedModelMappings: 0,
		missingMappingTargets: 0,
		aliasPreviewMappings: 0,
		modelPolicyDiffs: 1,
		replacePrunedTargets: 0,
	},
});

assert.deepEqual(rollbackSignalItems, [
	'Rollback preview emitted 1 warnings.',
	'Rollback preview found 1 conflicts.',
	'Channel-key credential rebind is required for 1 restored targets.',
	'Compatibility report found 1 missing providers.',
	'Compatibility report found 1 missing models.',
	'Compatibility report found 1 schema mismatches.',
	'Compatibility report skipped 1 targets.',
	'Route preview emitted 1 warnings.',
	'Compatibility report found 1 model-policy diffs.',
]);

const guidanceItems = buildImportCompatibilityGuidanceItems({
	effectiveMode: 'map',
	locale: 'en',
	counts: {
		conflicts: 1,
		aliasConflicts: 0,
		routeConflicts: 1,
		credentialRebindTargets: 1,
		channelKeyRebindTargets: 1,
		apiKeyRebindTargets: 0,
		invalidRouteTargets: 1,
		skippedRouteTargetPreviews: 1,
		routePreviewWarnings: 1,
		routePreviewDiffs: 1,
		missingProviders: 1,
		missingModels: 0,
		baseURLMismatches: 0,
		schemaMismatches: 0,
		skippedTargets: 2,
		modelMappingPreviews: 2,
		usedModelMappings: 1,
		unusedModelMappings: 1,
		missingMappingTargets: 1,
		aliasPreviewMappings: 0,
		modelPolicyDiffs: 1,
		replacePrunedTargets: 0,
	},
	compatibility: {
		missing_providers: ['legacy-provider'],
		skipped_targets: ['channel_key:201 empty credential', 'setting:api_base_url existing row preserved by skip mode'],
		credential_rebind_targets: [{ target_type: 'channel_key', channel_name: 'Primary', key_name: 'key-1', models: ['legacy-model'] }],
		skipped_route_target_previews: [{ group_name: 'group-b', channel_name: 'Backup', model: 'legacy-fallback', resolved_model: 'gpt-4.1', issue_type: 'skipped_preview', reason: 'model not declared on current channel', action: 'review mapping' }],
		model_mapping_previews: [
			{ source_model: 'missing-model', target_model: 'gpt-4.1-mini', used: true, target_exists: false, contexts: ['routing'], warnings: ['current model not found'] },
			{ source_model: 'unused-model', target_model: 'gpt-4.1', used: false, contexts: ['api_keys'] },
		],
	},
});

assert.deepEqual(guidanceItems, [
	{
		key: 'blocking-risks',
		tone: 'danger',
		title: 'Resolve blocking risks first',
		detail: 'The current preview still contains 1 compatibility conflicts, 1 route conflicts, 1 invalid route targets. Expand the conflict and route details below, fix them, then apply again.',
	},
	{
		key: 'missing-targets',
		tone: 'warning',
		title: 'Restore missing providers or models',
		detail: 'The preview already marked the missing objects. Start with these examples: legacy-provider; snapshot:missing-model | current:gpt-4.1-mini | contexts:routing | warnings:current model not found.',
	},
	{
		key: 'credential-rebind',
		tone: 'warning',
		title: 'Prepare credential rebinds',
		detail: 'These targets will need credential rebinds after apply: target:channel key | channel:Primary | key:key-1 | models:legacy-model.',
	},
	{
		key: 'skipped-targets',
		tone: 'warning',
		title: 'Review which targets are skipped',
		detail: 'This mode will preserve or skip these targets: channel_key:201 empty credential; setting:api_base_url existing row preserved by skip mode. Confirm that this matches what you want before applying.',
	},
	{
		key: 'model-mappings',
		tone: 'warning',
		title: 'Fix model mappings before apply',
		detail: 'The current mapping set already shows items to fix or remove: snapshot:missing-model | current:gpt-4.1-mini | contexts:routing | warnings:current model not found; snapshot:unused-model | current:gpt-4.1 | contexts:api_keys.',
	},
	{
		key: 'route-and-policy',
		tone: 'warning',
		title: 'Review route and policy drift',
		detail: 'The preview also surfaced 1 route diffs, 1 route warnings, 1 skipped route previews, 1 model-policy diffs. Review the route and policy details below before applying.',
	},
]);

const replaceGuidanceItems = buildImportCompatibilityGuidanceItems({
	effectiveMode: 'replace',
	locale: 'en',
	counts: {
		conflicts: 1,
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
		replacePrunedTargets: 2,
	},
	compatibility: {
		credential_rebind_targets: [{ target_type: 'channel_key', channel_name: 'Primary' }],
		replace_pruned_channels: ['legacy-channel'],
		replace_pruned_api_keys: ['client-key'],
	},
});

assert.deepEqual(replaceGuidanceItems, [
	{
		key: 'blocking-risks',
		tone: 'danger',
		title: 'Resolve blocking risks first',
		detail: 'The current preview still contains 1 compatibility conflicts. Expand the conflict and route details below, fix them, then apply again.',
	},
	{
		key: 'credential-rebind',
		tone: 'warning',
		title: 'Prepare credential rebinds',
		detail: 'These targets will need credential rebinds after apply: target:channel key | channel:Primary.',
	},
	{
		key: 'replace-prune',
		tone: 'warning',
		title: 'Review which current records replace mode removes',
		detail: 'Replace mode will prune these current records: legacy-channel; client-key. Expand the structured prune details below before you apply it.',
	},
]);

const rollbackGuidanceItems = buildImportCompatibilityGuidanceItems({
	kind: 'rollback',
	locale: 'en',
	counts: {
		conflicts: 1,
		aliasConflicts: 0,
		routeConflicts: 1,
		credentialRebindTargets: 1,
		channelKeyRebindTargets: 1,
		apiKeyRebindTargets: 0,
		invalidRouteTargets: 1,
		skippedRouteTargetPreviews: 1,
		routePreviewWarnings: 1,
		routePreviewDiffs: 1,
		missingProviders: 1,
		missingModels: 0,
		baseURLMismatches: 0,
		schemaMismatches: 1,
		skippedTargets: 1,
		modelMappingPreviews: 0,
		usedModelMappings: 0,
		unusedModelMappings: 0,
		missingMappingTargets: 0,
		aliasPreviewMappings: 0,
		modelPolicyDiffs: 1,
		replacePrunedTargets: 0,
	},
	compatibility: {
		missing_providers: ['rollback-provider'],
		skipped_targets: ['channel_key:101 empty credential'],
		credential_rebind_targets: [{ target_type: 'channel_key', channel_name: 'Primary', key_name: 'key-1', models: ['legacy-model'] }],
	},
});

assert.deepEqual(rollbackGuidanceItems, [
	{
		key: 'blocking-risks',
		tone: 'danger',
		title: 'Resolve rollback risks first',
		detail: 'The current rollback preview still contains 1 compatibility conflicts, 1 route conflicts, 1 invalid route targets, 1 schema mismatches. Expand the conflict and route details below, fix them, then run the rollback again.',
	},
	{
		key: 'missing-targets',
		tone: 'warning',
		title: 'Restore missing targets before rollback',
		detail: 'The rollback preview already marked the missing objects. Start with these examples: rollback-provider.',
	},
	{
		key: 'credential-rebind',
		tone: 'warning',
		title: 'Prepare post-rollback credential rebinds',
		detail: 'These targets will need credential rebinds after rollback: target:channel key | channel:Primary | key:key-1 | models:legacy-model.',
	},
	{
		key: 'skipped-targets',
		tone: 'warning',
		title: 'Review which targets rollback keeps',
		detail: 'This rollback will preserve or skip these targets: channel_key:101 empty credential. Confirm that this matches what you want before restoring.',
	},
	{
		key: 'route-and-policy',
		tone: 'warning',
		title: 'Review post-rollback route and policy drift',
		detail: 'The rollback preview also surfaced 1 route diffs, 1 route warnings, 1 skipped route previews, 1 model-policy diffs. Review the route and policy details below before restoring.',
	},
]);

const rollbackReplacePruneGuidanceItems = buildImportCompatibilityGuidanceItems({
	kind: 'rollback',
	locale: 'en',
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
		replacePrunedTargets: 2,
	},
	compatibility: {
		replace_pruned_channels: ['legacy-channel'],
		replace_pruned_api_keys: ['client-key'],
	},
});

assert.deepEqual(rollbackReplacePruneGuidanceItems, [
	{
		key: 'replace-prune',
		tone: 'warning',
		title: 'Review which current records rollback removes',
		detail: 'Rollback restore will prune these current records: legacy-channel; client-key. Expand the structured cleanup details below before restoring.',
	},
]);

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
assert.deepEqual(getReplacePrunedBreakdownItems({
	channels: ['legacy-channel'],
	settings: ['setting:api_base_url existing row preserved by skip mode'],
	apiKeys: ['client-key'],
}, 'zh-Hans'), {
	channels: ['legacy-channel'],
	groups: [],
	settings: ['系统设置:api_base_url 因跳过模式而保留当前记录'],
	llmInfos: [],
	apiKeys: ['client-key'],
});
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
assert.deepEqual(getCredentialRebindTargetItems([{ target_type: 'channel_key', channel_name: 'Primary', key_name: 'key-1', source_type: 'oauth', models: ['legacy-model'], affected_groups: ['group-a'], contexts: ['routing'] }], 'zh-Hans'), [
	'目标类型:渠道密钥 | 渠道:Primary | 密钥:key-1 | 来源:oauth | 模型:legacy-model | 影响分组:group-a | 作用范围:路由',
]);
assert.deepEqual(getCompatibilityNameItems(['group-a', 'group-b'], 'en'), ['group-a', 'group-b']);
assert.deepEqual(getCompatibilityDiagnosticItems(['channel_key:201 empty credential', 'setting:api_base_url existing row preserved by skip mode', 'snapshot schema:v2 differs'], 'zh-Hans'), [
	'渠道密钥:201 缺少明文凭证',
	'系统设置:api_base_url 因跳过模式而保留当前记录',
	'快照结构版本 v2 与当前导入链路不一致',
]);
assert.deepEqual(getRouteTargetIssueItems([{ group_name: 'group-a', channel_name: 'Primary', model: 'gpt-4o', resolved_model: 'gpt-4.1', issue_type: 'missing_target', reason: 'channel key missing', action: 'rebind credential' }], 'zh-Hans'), [
	'分组:group-a | 渠道:Primary | 模型:gpt-4o | 解析模型:gpt-4.1 | 问题类型:missing_target | 原因:channel key missing | 建议动作:rebind credential',
]);
assert.deepEqual(getRoutePreviewWarningItems(['route may degrade', 'route preview diffs: 2'], 'zh-Hans'), [
	'路由候选链可能降级',
	'路由预览发现 2 处差异',
]);
assert.deepEqual(getRoutePreviewDiffItems([{
	group_name: 'group-a',
	model: 'legacy-model',
	before_candidates: [{ channel_name: 'Primary', model: 'legacy-model', resolved_model: 'gpt-4o', priority: 1, weight: 100 }],
	after_candidates: [{ channel_name: 'Backup', model: 'legacy-model', resolved_model: 'gpt-4.1', priority: 2, weight: 50 }],
	removed_candidates: [{ channel_name: 'Primary', model: 'legacy-model', resolved_model: 'gpt-4o', priority: 1, weight: 100 }],
	added_candidates: [{ channel_name: 'Backup', model: 'legacy-model', resolved_model: 'gpt-4.1', priority: 2, weight: 50 }],
	fallback_changed: true,
	skip_reasons: ['missing candidate'],
}], 'zh-Hans'), [
	'分组:group-a | 模型:legacy-model | 当前候选:Primary:gpt-4o | 优先级:1 | 权重:100 | 快照候选:Backup:gpt-4.1 | 优先级:2 | 权重:50 | 将被移除:Primary:gpt-4o | 优先级:1 | 权重:100 | 将被新增:Backup:gpt-4.1 | 优先级:2 | 权重:50 | 回退链变化:是 | 跳过原因:缺少候选项',
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

