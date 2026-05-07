import { describe, expect, it } from 'vitest';

import {
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
	getMissingModelMappingItems,
	getModelMappingPreviewItems,
	getModelPolicyDiffItems,
	getPostImportValidationSummary,
	getRoutePreviewDiffItems,
	getRoutePreviewWarningItems,
	getRouteTargetIssueItems,
	getUnusedModelMappingItems,
} from './backup-logic';

describe('backup-logic', () => {
	it('derives compatibility counters from mixed summary inputs', () => {
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

		expect(counts).toMatchObject({
			conflicts: 1,
			credentialRebindTargets: 3,
			channelKeyRebindTargets: 2,
			apiKeyRebindTargets: 1,
			modelMappingPreviews: 3,
			usedModelMappings: 2,
			unusedModelMappings: 1,
			missingMappingTargets: 1,
			missingModels: 2,
			replacePrunedTargets: 3,
		});
	});

	it('marks rollback overview as dangerous when route or schema drift exists', () => {
		const overview = getCompatibilityOverview({
			kind: 'rollback',
			warningsCount: 0,
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

		expect(overview.tone).toBe('danger');
		expect(overview.title).toBe('回滚存在阻断风险');
		expect(overview.description).toContain('路由或结构层面的风险');
	});

	it('builds apply risk items for replace mode imports', () => {
		const items = buildCompatibilitySignalItems({
			kind: 'import',
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

		expect(items).toEqual([
			'替换模式会移除当前项目中那些未被快照保留的记录。',
			'替换清理预览会删除或重置当前项目中的 4 条记录。',
			'兼容性报告发现 2 个冲突。',
			'1 个导入目标需要重新绑定渠道密钥凭证。',
		]);
	});

	it('builds finer-grained import guidance items from compatibility details', () => {
		const guidance = buildImportCompatibilityGuidanceItems({
			effectiveMode: 'map',
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

		expect(guidance).toEqual([
			{
				key: 'blocking-risks',
				tone: 'danger',
				title: '先处理阻断风险',
				detail: '当前预检里有 1 个兼容冲突、1 个路由冲突、1 个无效路由目标。建议先展开下方冲突和路由明细，修正后再应用。',
			},
			{
				key: 'missing-targets',
				tone: 'warning',
				title: '补齐缺失的渠道或模型',
				detail: '预检已标出缺失对象，优先处理这些名称：legacy-provider；快照模型:missing-model | 当前模型:gpt-4.1-mini | 作用范围:路由 | 警告:当前项目中未找到该模型。',
			},
			{
				key: 'credential-rebind',
				tone: 'warning',
				title: '提前准备凭证重绑定',
				detail: '这些目标在应用后需要重新绑定凭证：目标类型:渠道密钥 | 渠道:Primary | 密钥:key-1 | 模型:legacy-model。',
			},
			{
				key: 'skipped-targets',
				tone: 'warning',
				title: '确认哪些对象会被跳过',
				detail: '当前模式会保留或跳过这些对象：渠道密钥:201 缺少明文凭证；系统设置:api_base_url 因跳过模式而保留当前记录。建议先确认这是否符合本次导入预期。',
			},
			{
				key: 'model-mappings',
				tone: 'warning',
				title: '修正模型映射后再应用',
				detail: '当前映射里已经暴露出需要修正或清理的项目：快照模型:missing-model | 当前模型:gpt-4.1-mini | 作用范围:路由 | 警告:当前项目中未找到该模型；快照模型:unused-model | 当前模型:gpt-4.1 | 作用范围:API密钥。',
			},
			{
				key: 'route-and-policy',
				tone: 'warning',
				title: '复核路由与策略差异',
				detail: '当前预检还提示了 1 处路由差异、1 条路由预警、1 个跳过的路由预览、1 条模型策略差异。建议在应用前把下方路由与策略明细过一遍。',
			},
		]);
	});

	it('adds replace-prune guidance when structured replace deletions are present', () => {
		const guidance = buildImportCompatibilityGuidanceItems({
			effectiveMode: 'replace',
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

		expect(guidance).toEqual([
			{
				key: 'blocking-risks',
				tone: 'danger',
				title: '先处理阻断风险',
				detail: '当前预检里有 1 个兼容冲突。建议先展开下方冲突和路由明细，修正后再应用。',
			},
			{
				key: 'credential-rebind',
				tone: 'warning',
				title: '提前准备凭证重绑定',
				detail: '这些目标在应用后需要重新绑定凭证：目标类型:渠道密钥 | 渠道:Primary。',
			},
			{
				key: 'replace-prune',
				tone: 'warning',
				title: '确认替换会清理哪些当前记录',
				detail: '替换导入会清理这些当前记录：legacy-channel；client-key。建议先展开下方结构化清理明细，再决定是否应用。',
			},
		]);
	});

	it('adds rollback replace-prune guidance when structured rollback cleanup is present', () => {
		const guidance = buildImportCompatibilityGuidanceItems({
			kind: 'rollback',
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

		expect(guidance).toEqual([
			{
				key: 'replace-prune',
				tone: 'warning',
				title: '确认回滚会清理哪些当前记录',
				detail: '回滚恢复会清理这些当前记录：legacy-channel；client-key。建议先展开下方结构化清理明细，再决定是否执行回滚。',
			},
		]);
	});

	it('uses import-report wording by default when import warnings are present', () => {
		const items = buildCompatibilitySignalItems({
			kind: 'import',
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

		expect(items).toEqual(['导入报告产生了 2 条警告。']);
	});

	it('guards apply-same-import with the expected reasons', () => {
		expect(getApplySameImportGuardReason({ hasPendingApplyRequest: false, previewToken: 'token', requiresConfirm: false, confirmed: false })).toBe('missing_request');
		expect(getApplySameImportGuardReason({ hasPendingApplyRequest: true, previewToken: '   ', requiresConfirm: false, confirmed: false })).toBe('missing_preview_token');
		expect(getApplySameImportGuardReason({ hasPendingApplyRequest: true, previewToken: 'token', requiresConfirm: true, confirmed: false })).toBe('confirm_required');
		expect(getApplySameImportGuardReason({ hasPendingApplyRequest: true, previewToken: 'token', requiresConfirm: true, confirmed: true })).toBeNull();
	});

	it('summarizes post-import validation and presentation helpers', () => {
		expect(getPostImportValidationSummary({
			degraded_groups: ['group-a'],
			empty_groups: ['group-b', 'group-c'],
			disabled_channels: [],
			channels_without_keys: ['channel-a'],
			stale_items_removed: ['item-a'],
			route_warnings: ['warning-a', 'warning-b'],
			price_rule_warnings: ['drift-a'],
			alias_mappings: ['alias-a'],
			alias_warnings: ['alias-warning-a', 'alias-warning-b'],
		})).toEqual({
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

		expect(getImportResultPresentation({ resultIsDryRun: true, modeLabel: 'Replace' })).toEqual({
			title: '预检结果',
			description: '预检只会分析兼容性和路由影响提示，不会写入数据库。',
			modeLabel: '本次使用的预检模式',
		});

		const applied = getImportResultPresentation({ resultIsDryRun: false, modeLabel: 'Merge' });
		expect(applied.title).toBe('导入已应用');
		expect(applied.modeLabel).toBe('本次应用的导入模式');
		expect(applied.description).toContain('已按 Merge 模式应用导入');
	});

	it('describes export snapshots for full and redacted modes', () => {
		expect(getExportSnapshotPresentation({ includeSecrets: true, includeLogs: false, includeStats: true })).toEqual({
			summary: '导出一份完整项目迁移快照，默认包含渠道、分组、路由条目、模型绑定、明文凭证，以及可选的统计与日志。只有在你明确需要脱敏审阅版时，才建议关闭凭证导出。',
			warning: '默认导出会包含明文凭证，这样快照才能直接恢复到另一套环境。只有在你明确要分享或审阅脱敏版时，才建议关闭。',
			scopeBadges: ['项目快照', '渠道 / 分组 / 路由', '明文凭证', '统计数据'],
			toggleLabel: '在快照中包含明文凭证',
		});

		expect(getExportSnapshotPresentation({ includeSecrets: false, includeLogs: true, includeStats: false })).toEqual({
			summary: '导出一份脱敏后的项目快照，包含渠道、分组、路由条目、模型绑定，以及可选的统计与日志。如果你需要一份可以直接恢复的迁移快照，请重新打开明文凭证导出。',
			warning: '这份导出不会包含明文凭证，适合分享或审阅，不适合直接恢复。',
			scopeBadges: ['项目快照', '渠道 / 分组 / 路由', '脱敏凭证', '中继日志'],
			toggleLabel: '在快照中包含明文凭证',
		});
		expect(getExportSnapshotPresentation({ includeSecrets: false, includeLogs: true, includeStats: false, locale: 'en' })).toEqual({
			summary: 'Export a redacted project snapshot with channels, groups, route items, model bindings, and optional stats/logs. Turn secrets back on when you need a restore-ready migration snapshot.',
			warning: 'This export omits plaintext credentials. Use it for sharing or review when you do not need a directly restorable snapshot.',
			scopeBadges: ['Project snapshot', 'Channels / groups / routing', 'Redacted credentials', 'Relay logs'],
			toggleLabel: 'Include plaintext credentials in the snapshot',
		});
	});
	it('formats preview detail helpers through a shared source', () => {
		expect(getAliasPreviewItems([{ snapshot_model: 'legacy-model', current_model: 'gpt-4o', canonical: 'gpt-4o', contexts: ['channel:preview-channel', 'group:preview-group'] }])).toEqual([
			'\u5feb\u7167\u6a21\u578b:legacy-model | \u5f53\u524d\u6a21\u578b:gpt-4o | \u89c4\u8303\u540d:gpt-4o | \u4f5c\u7528\u8303\u56f4:\u6e20\u9053:preview-channel\u3001\u5206\u7ec4:preview-group',
		]);
		expect(getModelMappingPreviewItems([{ source_model: 'legacy-model', target_model: 'gpt-4.1', usage_count: 3, used: true, target_exists: true, touched_fields: ['channels.model', 'group_items.model_name', 'api_keys.supported_models'], contexts: ['channel:mapped-channel', 'group_route:mapped-group', 'api_key:preview-client'] }])).toEqual([
			'\u5feb\u7167\u6a21\u578b:legacy-model | \u5f53\u524d\u6a21\u578b:gpt-4.1 | \u5f15\u7528\u6b21\u6570:3 | \u662f\u5426\u4f7f\u7528:\u662f | \u76ee\u6807\u72b6\u6001:\u5b58\u5728 | \u53d7\u5f71\u54cd\u5b57\u6bb5:\u6e20\u9053.\u6a21\u578b\u3001\u5206\u7ec4\u6761\u76ee.\u6a21\u578b\u540d\u79f0\u3001API\u5bc6\u94a5.\u652f\u6301\u6a21\u578b | \u4f5c\u7528\u8303\u56f4:\u6e20\u9053:mapped-channel\u3001\u5206\u7ec4\u8def\u7531:mapped-group\u3001API\u5bc6\u94a5:preview-client',
		]);
		expect(getAliasPreviewItems([{ snapshot_model: 'legacy-model', current_model: 'gpt-4o', canonical: 'gpt-4o', contexts: ['routing', 'fallback'] }])).toEqual([
			'快照模型:legacy-model | 当前模型:gpt-4o | 规范名:gpt-4o | 作用范围:路由、回退',
		]);
		expect(getModelMappingPreviewItems([{ source_model: 'legacy-model', target_model: 'gpt-4.1', usage_count: 2, used: true, target_exists: false, touched_fields: ['primary_model', 'fallback_model'], contexts: ['routing'], warnings: ['current model not found'] }])).toEqual([
			'快照模型:legacy-model | 当前模型:gpt-4.1 | 引用次数:2 | 是否使用:是 | 目标状态:缺失 | 受影响字段:主模型、备用模型 | 作用范围:路由 | 警告:当前项目中未找到该模型',
		]);
		expect(getMissingModelMappingItems([{ source_model: 'legacy-model', target_model: 'gpt-4.1', used: true, target_exists: false, contexts: ['routing'], warnings: ['current model not found'] }])).toEqual([
			'快照模型:legacy-model | 当前模型:gpt-4.1 | 作用范围:路由 | 警告:当前项目中未找到该模型',
		]);
		expect(getUnusedModelMappingItems([{ source_model: 'unused-model', target_model: 'gpt-4.1', used: false, contexts: ['api_keys'] }])).toEqual([
			'快照模型:unused-model | 当前模型:gpt-4.1 | 作用范围:API密钥',
		]);
		expect(getModelPolicyDiffItems([{ model: 'legacy-model', current_model: 'gpt-4.1', impact_level: 'high', changed_fields: ['billing_mode', 'probe_policy'], before: { billing_mode: 'paid', probe_policy: 'manual', probe_interval: 30, probe_concurrency: 1 }, after: { billing_mode: 'free', probe_policy: 'auto', probe_interval: 60, probe_concurrency: 2 }, contexts: ['routing'], warnings: ['policy drift'], skip_reasons: ['missing candidate'] }])).toEqual([
			'模型:legacy-model | 当前模型:gpt-4.1 | 影响级别:高 | 变更字段:计费模式、探测策略 | 变更前:计费:付费, 探测:手动, 间隔:30, 并发:1 | 变更后:计费:免费, 探测:自动, 间隔:60, 并发:2 | 作用范围:路由 | 警告:策略差异 | 跳过原因:缺少候选项',
		]);
		expect(getCredentialRebindTargetItems([{ target_type: 'channel_key', channel_name: 'Primary', key_name: 'key-1', source_type: 'oauth', models: ['legacy-model'], affected_groups: ['group-a'], contexts: ['routing'] }])).toEqual([
			'目标类型:渠道密钥 | 渠道:Primary | 密钥:key-1 | 来源:oauth | 模型:legacy-model | 影响分组:group-a | 作用范围:路由',
		]);
		expect(getCompatibilityNameItems(['group-a', 'group-b'])).toEqual(['group-a', 'group-b']);
		expect(getCompatibilityDiagnosticItems(['channel_key:201 empty credential', 'setting:api_base_url existing row preserved by skip mode', 'snapshot schema:v2 differs'])).toEqual([
			'渠道密钥:201 缺少明文凭证',
			'系统设置:api_base_url 因跳过模式而保留当前记录',
			'快照结构版本 v2 与当前导入链路不一致',
		]);
		expect(getRouteTargetIssueItems([{ group_name: 'group-a', channel_name: 'Primary', model: 'gpt-4o', resolved_model: 'gpt-4.1', issue_type: 'missing_target', reason: 'channel key missing', action: 'rebind credential' }])).toEqual([
			'分组:group-a | 渠道:Primary | 模型:gpt-4o | 解析模型:gpt-4.1 | 问题类型:missing_target | 原因:channel key missing | 建议动作:rebind credential',
		]);
		expect(getRoutePreviewWarningItems(['route may degrade', 'route preview diffs: 2'])).toEqual([
			'路由候选链可能降级',
			'路由预览发现 2 处差异',
		]);
		expect(getRoutePreviewDiffItems([{
			group_name: 'group-a',
			model: 'legacy-model',
			before_candidates: [{ channel_name: 'Primary', model: 'legacy-model', resolved_model: 'gpt-4o', priority: 1, weight: 100 }],
			after_candidates: [{ channel_name: 'Backup', model: 'legacy-model', resolved_model: 'gpt-4.1', priority: 2, weight: 50 }],
			removed_candidates: [{ channel_name: 'Primary', model: 'legacy-model', resolved_model: 'gpt-4o', priority: 1, weight: 100 }],
			added_candidates: [{ channel_name: 'Backup', model: 'legacy-model', resolved_model: 'gpt-4.1', priority: 2, weight: 50 }],
			fallback_changed: true,
			skip_reasons: ['missing candidate'],
		}])).toEqual([
			'分组:group-a | 模型:legacy-model | 当前候选:Primary:gpt-4o | 优先级:1 | 权重:100 | 快照候选:Backup:gpt-4.1 | 优先级:2 | 权重:50 | 将被移除:Primary:gpt-4o | 优先级:1 | 权重:100 | 将被新增:Backup:gpt-4.1 | 优先级:2 | 权重:50 | 回退链变化:是 | 跳过原因:缺少候选项',
		]);
	});

	it('localizes additional mapping warning strings for non-English locales', () => {
		expect(getMissingModelMappingItems([{ source_model: 'legacy-model', target_model: 'gpt-4.1', used: true, target_exists: false, contexts: ['routing'], warnings: ['mapped target not found in current environment'] }])).toEqual([
			'\u5feb\u7167\u6a21\u578b:legacy-model | \u5f53\u524d\u6a21\u578b:gpt-4.1 | \u4f5c\u7528\u8303\u56f4:\u8def\u7531 | \u8b66\u544a:\u5f53\u524d\u73af\u5883\u4e2d\u672a\u627e\u5230\u6620\u5c04\u76ee\u6807',
		]);
		expect(getUnusedModelMappingItems([{ source_model: 'unused-model', target_model: 'gpt-4.1', used: false, contexts: ['api_keys'], warnings: ['mapping source not referenced by selected import scopes'] }])).toEqual([
			'\u5feb\u7167\u6a21\u578b:unused-model | \u5f53\u524d\u6a21\u578b:gpt-4.1 | \u4f5c\u7528\u8303\u56f4:API\u5bc6\u94a5 | \u8b66\u544a:\u8be5\u6620\u5c04\u6765\u6e90\u672a\u88ab\u6240\u9009\u5bfc\u5165\u8303\u56f4\u5f15\u7528',
		]);
	});

	it('localizes model policy sentence warnings for non-English locales', () => {
		expect(getModelPolicyDiffItems([{
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
		}])).toEqual([
			'\u6a21\u578b:legacy-model | \u5f53\u524d\u6a21\u578b:gpt-4.1 | \u5f71\u54cd\u7ea7\u522b:\u9ad8 | \u53d8\u66f4\u5b57\u6bb5:\u8ba1\u8d39\u6a21\u5f0f\u3001\u63a2\u6d4b\u7b56\u7565 | \u53d8\u66f4\u524d:\u8ba1\u8d39:\u4ed8\u8d39, \u63a2\u6d4b:\u624b\u52a8, \u95f4\u9694:30, \u5e76\u53d1:1 | \u53d8\u66f4\u540e:\u8ba1\u8d39:\u514d\u8d39, \u63a2\u6d4b:\u81ea\u52a8, \u95f4\u9694:60, \u5e76\u53d1:2 | \u4f5c\u7528\u8303\u56f4:\u8def\u7531 | \u8b66\u544a:\u8ba1\u8d39\u6a21\u5f0f\u4ece \u4ed8\u8d39 \u53d8\u4e3a \u514d\u8d39\u3001\u63a2\u6d4b\u7b56\u7565\u4ece \u624b\u52a8 \u53d8\u4e3a \u81ea\u52a8\u3001\u63a2\u6d4b\u95f4\u9694\u4ece 30 \u53d8\u4e3a 60\u3001\u63a2\u6d4b\u5e76\u53d1\u4ece 1 \u53d8\u4e3a 2\u3001\u6a21\u578b legacy-model \u7684\u5e76\u53d1\u63a2\u6d4b\u6216\u7ade\u901f\u53ef\u80fd\u589e\u52a0\u6210\u672c',
		]);
	});
});

