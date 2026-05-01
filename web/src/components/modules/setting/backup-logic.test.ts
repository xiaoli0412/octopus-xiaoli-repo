import { describe, expect, it } from 'vitest';

import {
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
	it('shares the remaining migration tooling gaps through a single source', () => {
		expect(getRemainingMigrationToolingItems()).toEqual([
			{ key: 'conflict-handling', label: '冲突处理', text: '在当前结构化预览基础上，进一步补充更细致的 “替换导入 / 映射导入” 场景冲突引导。' },
			{ key: 'mapping-editor', label: '映射编辑器', text: '在当前按行填写 “快照模型=当前模型” 的基础上，补充更完整的模型映射编辑能力。' },
			{ key: 'compare-workflow', label: '对比工作流', text: '在当前快照历史与预览面板基础上，补充多快照对比和更顺畅的差异导航。' },
			{ key: 'rollback-domains', label: '回滚域控制', text: '在当前整包快照恢复和选择性范围覆盖之外，补充更细粒度的回滚域编辑。' },
			{ key: 'route-diff', label: '路由差异对比', text: '在当前摘要卡和详情列表基础上，补充并排式的路由差异查看能力。' },
		]);
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

