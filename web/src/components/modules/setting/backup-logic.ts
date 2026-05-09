export type SummaryTone = 'safe' | 'warning' | 'danger';

export type ImportMode = 'incremental' | 'map' | 'merge' | 'replace' | 'skip';

type CompatibilitySummaryLike = {
	missing_providers?: number;
	missing_models?: number;
	conflicts?: number;
	alias_conflicts?: number;
	credential_rebind_targets?: number;
	channel_key_rebind_targets?: number;
	api_key_rebind_targets?: number;
	model_mapping_previews?: number;
	used_model_mappings?: number;
	unused_model_mappings?: number;
	missing_mapping_targets?: number;
	alias_preview_mappings?: number;
	model_policy_diffs?: number;
	route_conflicts?: number;
	invalid_route_targets?: number;
	skipped_route_target_previews?: number;
	route_preview_warnings?: number;
	route_preview_diffs?: number;
	base_url_mismatches?: number;
	schema_mismatches?: number;
	skipped_targets?: number;
	replace_pruned_channels?: number;
	replace_pruned_groups?: number;
	replace_pruned_settings?: number;
	replace_pruned_llm_infos?: number;
	replace_pruned_api_keys?: number;
};

type CredentialRebindTargetLike = {
	target_type: string;
	snapshot_id?: number;
	channel_name?: string;
	key_name?: string;
	source_type?: string;
	remark?: string;
	models?: string[];
	affected_groups?: string[];
	contexts?: string[];
	warnings?: string[];
};

type RouteTargetPreviewIssueLike = {
	group_name?: string;
	channel_name?: string;
	model?: string;
	resolved_model?: string;
	key_id?: number;
	issue_type: string;
	reason?: string;
	action?: string;
};

type RoutePreviewCandidateLike = {
	channel_name?: string;
	model?: string;
	resolved_model?: string;
	priority?: number;
	weight?: number;
	reason?: string;
};

type RoutePreviewDiffLike = {
	group_name?: string;
	model?: string;
	before_candidates?: RoutePreviewCandidateLike[];
	after_candidates?: RoutePreviewCandidateLike[];
	removed_candidates?: RoutePreviewCandidateLike[];
	added_candidates?: RoutePreviewCandidateLike[];
	skip_reasons?: string[];
	fallback_changed?: boolean;
};

type ModelMappingPreviewLike = {
	used?: boolean;
	target_exists?: boolean;
};

type AliasPreviewMappingLike = {
	snapshot_model?: string;
	current_model?: string;
	canonical?: string;
	contexts?: string[];
};

type ModelPolicyStateLike = {
	billing_mode?: string;
	probe_policy?: string;
	probe_interval?: number;
	probe_concurrency?: number;
};

type ModelPolicyDiffLike = {
	model?: string;
	current_model?: string;
	impact_level?: string;
	changed_fields?: string[];
	before?: ModelPolicyStateLike;
	after?: ModelPolicyStateLike;
	contexts?: string[];
	warnings?: string[];
	skip_reasons?: string[];
};

type ModelMappingPreviewDisplayLike = ModelMappingPreviewLike & {
	source_model?: string;
	target_model?: string;
	contexts?: string[];
	touched_fields?: string[];
	usage_count?: number;
	warnings?: string[];
};

export type CompatibilityLike = {
	summary?: CompatibilitySummaryLike;
	conflicts?: unknown[];
	alias_conflicts?: unknown[];
	route_conflicts?: unknown[];
	credential_rebind_targets?: CredentialRebindTargetLike[];
	invalid_route_targets?: unknown[];
	skipped_route_target_previews?: unknown[];
	route_preview_warnings?: unknown[];
	route_preview_diffs?: unknown[];
	missing_providers?: unknown[];
	missing_models?: unknown[];
	base_url_mismatches?: unknown[];
	schema_mismatches?: unknown[];
	skipped_targets?: unknown[];
	model_mapping_previews?: ModelMappingPreviewDisplayLike[];
	alias_preview_mappings?: unknown[];
	model_policy_diffs?: unknown[];
	replace_pruned_channels?: unknown[];
	replace_pruned_groups?: unknown[];
	replace_pruned_settings?: unknown[];
	replace_pruned_llm_infos?: unknown[];
	replace_pruned_api_keys?: unknown[];
};

export type CompatibilityCounts = {
	conflicts: number;
	aliasConflicts: number;
	routeConflicts: number;
	credentialRebindTargets: number;
	channelKeyRebindTargets: number;
	apiKeyRebindTargets: number;
	invalidRouteTargets: number;
	skippedRouteTargetPreviews: number;
	routePreviewWarnings: number;
	routePreviewDiffs: number;
	missingProviders: number;
	missingModels: number;
	baseURLMismatches: number;
	schemaMismatches: number;
	skippedTargets: number;
	modelMappingPreviews: number;
	usedModelMappings: number;
	unusedModelMappings: number;
	missingMappingTargets: number;
	aliasPreviewMappings: number;
	modelPolicyDiffs: number;
	replacePrunedTargets: number;
};

export type CompatibilityGuidanceItem = {
	key: string;
	tone: SummaryTone;
	title: string;
	detail: string;
};

type ReplacePrunedBreakdownLike = {
	channels?: unknown[];
	groups?: unknown[];
	settings?: unknown[];
	models?: unknown[];
	apiKeys?: unknown[];
};

type CompatibilityOverviewInput = {
	counts: CompatibilityCounts;
	warningsCount: number;
	kind: 'import' | 'rollback';
	locale?: BackupLogicLocale;
};

type CompatibilitySignalItemsInput = {
	kind: 'import' | 'rollback';
	counts: CompatibilityCounts;
	locale?: BackupLogicLocale;
	warningsCount?: number;
	replacePruneCount?: number;
	structuredReplacePrunedCount?: number;
	includeWarningsCount?: boolean;
	includeReplaceModeRisk?: boolean;
	includeModelMappingPreviews?: boolean;
	includeUnusedModelMappings?: boolean;
	includeStructuredReplacePrunedCount?: boolean;
	effectiveMode?: ImportMode;
	importWarningsLabel?: string;
};

type CompatibilityGuidanceItemsInput = {
	compatibility: CompatibilityLike | undefined;
	counts: CompatibilityCounts;
	kind?: 'import' | 'rollback';
	effectiveMode?: ImportMode;
	locale?: BackupLogicLocale;
};

type PostImportValidationSummaryLike = {
	groups_scanned?: number;
	candidates_scanned?: number;
	degraded_groups?: number;
	empty_groups?: number;
	disabled_channels?: number;
	channels_without_keys?: number;
	stale_items_removed?: number;
	route_warnings?: number;
	price_rule_warnings?: number;
	alias_mappings?: number;
	alias_warnings?: number;
};

export type PostImportValidationLike = {
	summary?: PostImportValidationSummaryLike;
	degraded_groups?: unknown[];
	empty_groups?: unknown[];
	disabled_channels?: unknown[];
	channels_without_keys?: unknown[];
	stale_items_removed?: unknown[];
	route_warnings?: unknown[];
	price_rule_warnings?: unknown[];
	alias_mappings?: unknown[];
	alias_warnings?: unknown[];
};

export type PostImportValidationSummary = {
	degradedGroups: number;
	emptyGroups: number;
	disabledChannels: number;
	channelsWithoutKeys: number;
	staleItemsRemoved: number;
	routeWarnings: number;
	priceRuleWarnings: number;
	aliasMappings: number;
	aliasWarnings: number;
};

export type ExportSnapshotPresentation = {
	summary: string;
	warning: string;
	scopeBadges: string[];
	toggleLabel: string;
};

export type ApplySameImportGuardReason = 'missing_request' | 'missing_preview_token' | 'confirm_required' | null;

export type BackupLogicLocale = 'zh-Hans' | 'zh-Hant' | 'en' | 'ja';

function isZhHant(locale: BackupLogicLocale) {
	return locale === 'zh-Hant';
}

function isJa(locale: BackupLogicLocale) {
	return locale === 'ja';
}

function t(locale: BackupLogicLocale, text: { 'zh-Hans': string; 'zh-Hant'?: string; en?: string; ja?: string }) {
	if (locale === 'zh-Hans') return text['zh-Hans'];
	if (locale === 'zh-Hant') return text['zh-Hant'] ?? text['zh-Hans'];
	if (locale === 'ja') return text.ja ?? text.en ?? text['zh-Hans'];
	return text.en ?? text['zh-Hans'];
}

function getCompatibilityCount(summaryValue: number | undefined, items: unknown[] | undefined) {
	return getSummaryOrFallbackCount(summaryValue, Array.isArray(items) ? items.length : 0);
}

function getSummaryOrListCount(summaryValue: number | undefined, items: unknown[] | undefined) {
	return getSummaryOrFallbackCount(summaryValue, Array.isArray(items) ? items.length : 0);
}

function getSummaryOrFallbackCount(summaryValue: number | undefined, fallbackValue: number) {
	if (typeof summaryValue !== 'number') return fallbackValue;
	return Math.max(summaryValue, fallbackValue);
}

export function getEffectiveCredentialRebindCount(counts: {
	credentialRebindTargets: number;
	channelKeyRebindTargets: number;
	apiKeyRebindTargets: number;
}) {
	return Math.max(counts.credentialRebindTargets, counts.channelKeyRebindTargets + counts.apiKeyRebindTargets);
}

function countStructuredReplacePrunedItems(compatibility: CompatibilityLike | undefined) {
	if (!compatibility) return 0;
	return getSummaryOrFallbackCount(compatibility.summary?.replace_pruned_channels, compatibility.replace_pruned_channels?.length ?? 0)
		+ getSummaryOrFallbackCount(compatibility.summary?.replace_pruned_groups, compatibility.replace_pruned_groups?.length ?? 0)
		+ getSummaryOrFallbackCount(compatibility.summary?.replace_pruned_settings, compatibility.replace_pruned_settings?.length ?? 0)
		+ getSummaryOrFallbackCount(compatibility.summary?.replace_pruned_llm_infos, compatibility.replace_pruned_llm_infos?.length ?? 0)
		+ getSummaryOrFallbackCount(compatibility.summary?.replace_pruned_api_keys, compatibility.replace_pruned_api_keys?.length ?? 0);
}

function joinLocalizedList(items: string[] | undefined, locale: BackupLogicLocale) {
	const values = items ?? [];
	if (values.length === 0) return '';
	return values.join(locale === 'en' ? ', ' : '\u3001');
}

function normalizeLocalizedDetailText(text: string, locale: BackupLogicLocale) {
	if (locale !== 'en') return text;
	return text.replaceAll('\u3001', ', ');
}

function normalizeLocalizedPrimaryCopy(text: string, locale: BackupLogicLocale) {
	if (locale === 'zh-Hans') {
		return text
			.replaceAll('replace/map', '“替换导入 / 映射导入”')
			.replaceAll(' remap ', ' “快照模型=当前模型” ')
			.replaceAll(' remap。', ' “快照模型=当前模型”。')
			.replaceAll(' remap 的', ' “快照模型=当前模型” 的');
	}
	if (locale === 'zh-Hant') {
		return text
			.replaceAll('replace/map', '「替換導入 / 映射導入」')
			.replaceAll(' remap ', '「快照模型=目前模型」')
			.replaceAll(' remap。', '「快照模型=目前模型」。')
			.replaceAll(' remap 的', '「快照模型=目前模型」的');
	}
	return text;
}

type LocalizedValueMap = Record<string, { 'zh-Hans': string; 'zh-Hant'?: string; en?: string; ja?: string }>;

type LocalizedTokenMap = Record<string, { 'zh-Hans': string; 'zh-Hant'?: string; en?: string; ja?: string }>;

const BACKUP_DIAGNOSTIC_TEXT: LocalizedValueMap = {
	'current model not found': {
		'zh-Hans': '\u5f53\u524d\u9879\u76ee\u4e2d\u672a\u627e\u5230\u8be5\u6a21\u578b',
		'zh-Hant': '\u76ee\u524d\u5c08\u6848\u4e2d\u627e\u4e0d\u5230\u8a72\u6a21\u578b',
		en: 'current model not found',
		ja: '\u73fe\u5728\u306e\u30d7\u30ed\u30b8\u30a7\u30af\u30c8\u3067\u305d\u306e\u30e2\u30c7\u30eb\u304c\u898b\u3064\u304b\u308a\u307e\u305b\u3093',
	},
	'mapped target not found in current environment': {
		'zh-Hans': '\u5f53\u524d\u73af\u5883\u4e2d\u672a\u627e\u5230\u6620\u5c04\u76ee\u6807',
		'zh-Hant': '\u76ee\u524d\u74b0\u5883\u4e2d\u627e\u4e0d\u5230\u6620\u5c04\u76ee\u6a19',
		en: 'mapped target not found in current environment',
		ja: '\u73fe\u5728\u306e\u74b0\u5883\u3067\u30de\u30c3\u30d4\u30f3\u30b0\u5148\u304c\u898b\u3064\u304b\u308a\u307e\u305b\u3093',
	},
	'mapping source not referenced by selected import scopes': {
		'zh-Hans': '\u8be5\u6620\u5c04\u6765\u6e90\u672a\u88ab\u6240\u9009\u5bfc\u5165\u8303\u56f4\u5f15\u7528',
		'zh-Hant': '\u8a72\u6620\u5c04\u4f86\u6e90\u672a\u88ab\u6240\u9078\u5c0e\u5165\u7bc4\u570d\u5f15\u7528',
		en: 'mapping source not referenced by selected import scopes',
		ja: '\u3053\u306e\u30de\u30c3\u30d4\u30f3\u30b0\u5143\u306f\u9078\u629e\u3055\u308c\u305f\u30a4\u30f3\u30dd\u30fc\u30c8\u7bc4\u56f2\u304b\u3089\u53c2\u7167\u3055\u308c\u3066\u3044\u307e\u305b\u3093',
	},
	'policy drift': {
		'zh-Hans': '\u7b56\u7565\u5dee\u5f02',
		'zh-Hant': '\u7b56\u7565\u5dee\u7570',
		en: 'policy drift',
		ja: '\u30dd\u30ea\u30b7\u30fc\u5dee\u5206',
	},
	'route may degrade': {
		'zh-Hans': '\u8def\u7531\u5019\u9009\u94fe\u53ef\u80fd\u964d\u7ea7',
		'zh-Hant': '\u8def\u7531\u5019\u9078\u93c8\u53ef\u80fd\u964d\u7d1a',
		en: 'route may degrade',
		ja: '\u30eb\u30fc\u30c8\u5019\u88dc\u304c\u52a3\u5316\u3059\u308b\u53ef\u80fd\u6027\u304c\u3042\u308a\u307e\u3059',
	},
	'route preview needs manual review': {
		'zh-Hans': '\u8def\u7531\u9884\u89c8\u4ecd\u9700\u8981\u4eba\u5de5\u590d\u6838',
		'zh-Hant': '\u8def\u7531\u9810\u89bd\u4ecd\u9700\u8981\u4eba\u5de5\u8907\u6838',
		en: 'route preview needs manual review',
		ja: '\u30eb\u30fc\u30c8\u30d7\u30ec\u30d3\u30e5\u30fc\u306f\u307e\u3060\u624b\u52d5\u78ba\u8a8d\u304c\u5fc5\u8981\u3067\u3059',
	},
	'missing candidate': {
		'zh-Hans': '\u7f3a\u5c11\u5019\u9009\u9879',
		'zh-Hant': '\u7f3a\u5c11\u5019\u9078\u9805',
		en: 'missing candidate',
		ja: '\u5019\u88dc\u304c\u4e0d\u8db3\u3057\u3066\u3044\u307e\u3059',
	},
};

const BACKUP_CONTEXT_TEXT: LocalizedTokenMap = {
	routing: { 'zh-Hans': '\u8def\u7531', 'zh-Hant': '\u8def\u7531', en: 'routing', ja: '\u30eb\u30fc\u30c6\u30a3\u30f3\u30b0' },
	api_keys: { 'zh-Hans': 'API\u5bc6\u94a5', 'zh-Hant': 'API\u91d1\u9470', en: 'api keys', ja: 'API\u30ad\u30fc' },
	fallback: { 'zh-Hans': '\u56de\u9000', 'zh-Hant': '\u5f8c\u5099', en: 'fallback', ja: '\u30d5\u30a9\u30fc\u30eb\u30d0\u30c3\u30af' },
	channel: { 'zh-Hans': '\u6e20\u9053', 'zh-Hant': '\u6e20\u9053', en: 'channel', ja: '\u30c1\u30e3\u30cd\u30eb' },
	channel_key: { 'zh-Hans': '\u6e20\u9053\u5bc6\u94a5', 'zh-Hant': '\u6e20\u9053\u91d1\u9470', en: 'channel key', ja: '\u30c1\u30e3\u30cd\u30eb\u30ad\u30fc' },
	group: { 'zh-Hans': '\u5206\u7ec4', 'zh-Hant': '\u5206\u7d44', en: 'group', ja: '\u30b0\u30eb\u30fc\u30d7' },
	group_route: { 'zh-Hans': '\u5206\u7ec4\u8def\u7531', 'zh-Hant': '\u5206\u7d44\u8def\u7531', en: 'group route', ja: '\u30b0\u30eb\u30fc\u30d7\u30eb\u30fc\u30c8' },
	api_key: { 'zh-Hans': 'API\u5bc6\u94a5', 'zh-Hant': 'API\u91d1\u9470', en: 'api key', ja: 'API\u30ad\u30fc' },
	api_key_model: { 'zh-Hans': 'API\u5bc6\u94a5\u6a21\u578b', 'zh-Hant': 'API\u91d1\u9470\u6a21\u578b', en: 'api key model', ja: 'API\u30ad\u30fc\u30e2\u30c7\u30eb' },
	route_target_override: { 'zh-Hans': '\u8def\u7531\u76ee\u6807\u8986\u76d6', 'zh-Hant': '\u8def\u7531\u76ee\u6a19\u8986\u84cb', en: 'route target override', ja: '\u30eb\u30fc\u30c8\u5bfe\u8c61\u4e0a\u66f8\u304d' },
	llm_info: { 'zh-Hans': '\u6a21\u578b\u4fe1\u606f', 'zh-Hant': '\u6a21\u578b\u8cc7\u8a0a', en: 'llm info', ja: '\u30e2\u30c7\u30eb\u60c5\u5831' },
	stats_model: { 'zh-Hans': '\u6a21\u578b\u7edf\u8ba1', 'zh-Hant': '\u6a21\u578b\u7d71\u8a08', en: 'stats model', ja: '\u30e2\u30c7\u30eb\u7d71\u8a08' },
	relay_log: { 'zh-Hans': '\u4e2d\u7ee7\u65e5\u5fd7', 'zh-Hant': '\u4e2d\u7e7c\u65e5\u8a8c', en: 'relay log', ja: '\u4e2d\u7d99\u30ed\u30b0' },
	attempt: { 'zh-Hans': '\u5c1d\u8bd5', 'zh-Hant': '\u5617\u8a66', en: 'attempt', ja: '\u8a66\u884c' },
};
const BACKUP_FIELD_TEXT: LocalizedTokenMap = {
	primary_model: { 'zh-Hans': '\u4e3b\u6a21\u578b', 'zh-Hant': '\u4e3b\u6a21\u578b', en: 'primary model', ja: '\u4e3b\u30e2\u30c7\u30eb' },
	fallback_model: { 'zh-Hans': '\u5907\u7528\u6a21\u578b', 'zh-Hant': '\u5099\u7528\u6a21\u578b', en: 'fallback model', ja: '\u4ee3\u66ff\u30e2\u30c7\u30eb' },
	billing_mode: { 'zh-Hans': '\u8ba1\u8d39\u6a21\u5f0f', 'zh-Hant': '\u8a08\u8cbb\u6a21\u5f0f', en: 'billing mode', ja: '\u8ab2\u91d1\u30e2\u30fc\u30c9' },
	probe_policy: { 'zh-Hans': '\u63a2\u6d4b\u7b56\u7565', 'zh-Hant': '\u63a2\u6e2c\u7b56\u7565', en: 'probe policy', ja: '\u30d7\u30ed\u30fc\u30d6\u65b9\u91dd' },
	model: { 'zh-Hans': '\u6a21\u578b', 'zh-Hant': '\u6a21\u578b', en: 'model', ja: '\u30e2\u30c7\u30eb' },
	name: { 'zh-Hans': '\u540d\u79f0', 'zh-Hant': '\u540d\u7a31', en: 'name', ja: '\u540d\u524d' },
	canonical_name: { 'zh-Hans': '\u89c4\u8303\u540d\u79f0', 'zh-Hant': '\u898f\u7bc4\u540d\u7a31', en: 'canonical name', ja: '\u6b63\u898f\u540d' },
	allowed_models: { 'zh-Hans': '\u5141\u8bb8\u6a21\u578b', 'zh-Hant': '\u5141\u8a31\u6a21\u578b', en: 'allowed models', ja: '\u8a31\u53ef\u30e2\u30c7\u30eb' },
	supported_models: { 'zh-Hans': '\u652f\u6301\u6a21\u578b', 'zh-Hant': '\u652f\u63f4\u6a21\u578b', en: 'supported models', ja: '\u5bfe\u5fdc\u30e2\u30c7\u30eb' },
	custom_model: { 'zh-Hans': '\u81ea\u5b9a\u4e49\u6a21\u578b', 'zh-Hant': '\u81ea\u8a02\u6a21\u578b', en: 'custom model', ja: '\u30ab\u30b9\u30bf\u30e0\u30e2\u30c7\u30eb' },
	model_name: { 'zh-Hans': '\u6a21\u578b\u540d\u79f0', 'zh-Hant': '\u6a21\u578b\u540d\u7a31', en: 'model name', ja: '\u30e2\u30c7\u30eb\u540d' },
	request_model_name: { 'zh-Hans': '\u8bf7\u6c42\u6a21\u578b', 'zh-Hant': '\u8acb\u6c42\u6a21\u578b', en: 'request model', ja: '\u30ea\u30af\u30a8\u30b9\u30c8\u30e2\u30c7\u30eb' },
	actual_model_name: { 'zh-Hans': '\u5b9e\u9645\u6a21\u578b', 'zh-Hant': '\u5be6\u969b\u6a21\u578b', en: 'actual model', ja: '\u5b9f\u969b\u306e\u30e2\u30c7\u30eb' },
	channels: { 'zh-Hans': '\u6e20\u9053', 'zh-Hant': '\u6e20\u9053', en: 'channels', ja: '\u30c1\u30e3\u30cd\u30eb' },
	channel_keys: { 'zh-Hans': '\u6e20\u9053\u5bc6\u94a5', 'zh-Hant': '\u6e20\u9053\u91d1\u9470', en: 'channel keys', ja: '\u30c1\u30e3\u30cd\u30eb\u30ad\u30fc' },
	group_items: { 'zh-Hans': '\u5206\u7ec4\u6761\u76ee', 'zh-Hant': '\u5206\u7d44\u689d\u76ee', en: 'group items', ja: '\u30b0\u30eb\u30fc\u30d7\u9805\u76ee' },
	api_keys: { 'zh-Hans': 'API\u5bc6\u94a5', 'zh-Hant': 'API\u91d1\u9470', en: 'api keys', ja: 'API\u30ad\u30fc' },
	route_target_overrides: { 'zh-Hans': '\u8def\u7531\u76ee\u6807\u8986\u76d6', 'zh-Hant': '\u8def\u7531\u76ee\u6a19\u8986\u84cb', en: 'route target overrides', ja: '\u30eb\u30fc\u30c8\u5bfe\u8c61\u4e0a\u66f8\u304d' },
	llm_infos: { 'zh-Hans': '\u6a21\u578b\u4fe1\u606f', 'zh-Hant': '\u6a21\u578b\u8cc7\u8a0a', en: 'llm infos', ja: '\u30e2\u30c7\u30eb\u60c5\u5831' },
	stats_model: { 'zh-Hans': '\u6a21\u578b\u7edf\u8ba1', 'zh-Hant': '\u6a21\u578b\u7d71\u8a08', en: 'stats model', ja: '\u30e2\u30c7\u30eb\u7d71\u8a08' },
	relay_logs: { 'zh-Hans': '\u4e2d\u7ee7\u65e5\u5fd7', 'zh-Hant': '\u4e2d\u7e7c\u65e5\u8a8c', en: 'relay logs', ja: '\u4e2d\u7d99\u30ed\u30b0' },
	attempts: { 'zh-Hans': '\u5c1d\u8bd5\u8bb0\u5f55', 'zh-Hant': '\u5617\u8a66\u8a18\u9304', en: 'attempts', ja: '\u8a66\u884c\u8a18\u9332' },
};
const BACKUP_ENUM_TEXT: LocalizedTokenMap = {
	high: { 'zh-Hans': '高', 'zh-Hant': '高', en: 'high', ja: '高' },
	medium: { 'zh-Hans': '中', 'zh-Hant': '中', en: 'medium', ja: '中' },
	low: { 'zh-Hans': '低', 'zh-Hant': '低', en: 'low', ja: '低' },
	paid: { 'zh-Hans': '付费', 'zh-Hant': '付費', en: 'paid', ja: '有料' },
	free: { 'zh-Hans': '免费', 'zh-Hant': '免費', en: 'free', ja: '無料' },
	manual: { 'zh-Hans': '手动', 'zh-Hant': '手動', en: 'manual', ja: '手動' },
	auto: { 'zh-Hans': '自动', 'zh-Hant': '自動', en: 'auto', ja: '自動' },
	per_request: { 'zh-Hans': '按次计费', 'zh-Hant': '按次計費', en: 'per request', ja: 'リクエスト毎' },
	per_token: { 'zh-Hans': '按令牌计费', 'zh-Hant': '按令牌計費', en: 'per token', ja: 'トークン毎' },
	per_quota: { 'zh-Hans': '按额度计费', 'zh-Hant': '按額度計費', en: 'per quota', ja: 'クォータ毎' },
	flat: { 'zh-Hans': '固定计费', 'zh-Hant': '固定計費', en: 'flat', ja: '固定' },
	unknown: { 'zh-Hans': '未知', 'zh-Hant': '未知', en: 'unknown', ja: '不明' },
	passive_only: { 'zh-Hans': '仅被动探测', 'zh-Hant': '僅被動探測', en: 'passive only', ja: '受動のみ' },
	sparse_single: { 'zh-Hans': '稀疏单探测', 'zh-Hant': '稀疏單探測', en: 'sparse single', ja: '疎な単独' },
	sequential: { 'zh-Hans': '顺序探测', 'zh-Hant': '順序探測', en: 'sequential', ja: '順次' },
	concurrent: { 'zh-Hans': '并发探测', 'zh-Hant': '並發探測', en: 'concurrent', ja: '並列' },
};

type BackupPolicyWarningPattern = {
	pattern: RegExp;
	format: (locale: BackupLogicLocale, ...captures: string[]) => string;
};

const LOCALIZED_BACKUP_POLICY_WARNING_PATTERNS: readonly BackupPolicyWarningPattern[] = [
	{
		pattern: /^channel_key:(\d+) empty credential$/,
		format: (locale: BackupLogicLocale, ...[id = '']) => t(locale, {
			'zh-Hans': `渠道密钥:${id.trim()} 缺少明文凭证`,
			'zh-Hant': `渠道金鑰:${id.trim()} 缺少明文憑證`,
			en: `channel_key:${id.trim()} empty credential`,
			ja: `チャネルキー:${id.trim()} に平文認証情報がありません`,
		}),
	},
	{
		pattern: /^api_key:(\d+) empty credential$/,
		format: (locale: BackupLogicLocale, ...[id = '']) => t(locale, {
			'zh-Hans': `API 密钥:${id.trim()} 缺少明文凭证`,
			'zh-Hant': `API 金鑰:${id.trim()} 缺少明文憑證`,
			en: `api_key:${id.trim()} empty credential`,
			ja: `API キー:${id.trim()} に平文認証情報がありません`,
		}),
	},
	{
		pattern: /^channel:(.+) existing row preserved by skip mode$/,
		format: (locale: BackupLogicLocale, ...[name = '']) => t(locale, {
			'zh-Hans': `渠道:${name.trim()} 因跳过模式而保留当前记录`,
			'zh-Hant': `渠道:${name.trim()} 因跳過模式而保留目前記錄`,
			en: `channel:${name.trim()} existing row preserved by skip mode`,
			ja: `チャネル:${name.trim()} は skip モードにより既存レコードを保持しました`,
		}),
	},
	{
		pattern: /^setting:(.+) existing row preserved by skip mode$/,
		format: (locale: BackupLogicLocale, ...[key = '']) => t(locale, {
			'zh-Hans': `系统设置:${key.trim()} 因跳过模式而保留当前记录`,
			'zh-Hant': `系統設定:${key.trim()} 因跳過模式而保留目前記錄`,
			en: `setting:${key.trim()} existing row preserved by skip mode`,
			ja: `設定:${key.trim()} は skip モードにより既存レコードを保持しました`,
		}),
	},
	{
		pattern: /^snapshot schema:(.+) differs$/,
		format: (locale: BackupLogicLocale, ...[schema = '']) => t(locale, {
			'zh-Hans': `快照结构版本 ${schema.trim()} 与当前导入链路不一致`,
			'zh-Hant': `快照結構版本 ${schema.trim()} 與目前導入鏈路不一致`,
			en: `snapshot schema:${schema.trim()} differs`,
			ja: `スナップショットのスキーマ ${schema.trim()} が現在の導入チェーンと一致しません`,
		}),
		},
	{
		pattern: /^billing_mode changed from (.+) to (.+)$/,
		format: (locale: BackupLogicLocale, ...[from = '', to = '']) => t(locale, {
			'zh-Hans': `\u8ba1\u8d39\u6a21\u5f0f\u4ece ${formatPolicyValue(from, locale)} \u53d8\u4e3a ${formatPolicyValue(to, locale)}`,
			'zh-Hant': `\u8a08\u8cbb\u6a21\u5f0f\u5f9e ${formatPolicyValue(from, locale)} \u8b8a\u70ba ${formatPolicyValue(to, locale)}`,
			en: `billing_mode changed from ${from.trim()} to ${to.trim()}`,
			ja: `\u8ab2\u91d1\u30e2\u30fc\u30c9\u304c ${formatPolicyValue(from, locale)} \u304b\u3089 ${formatPolicyValue(to, locale)} \u306b\u5909\u308f\u308a\u307e\u3057\u305f`,
		}),
	},
	{
		pattern: /^probe_policy changed from (.+) to (.+)$/,
		format: (locale: BackupLogicLocale, ...[from = '', to = '']) => t(locale, {
			'zh-Hans': `\u63a2\u6d4b\u7b56\u7565\u4ece ${formatPolicyValue(from, locale)} \u53d8\u4e3a ${formatPolicyValue(to, locale)}`,
			'zh-Hant': `\u63a2\u6e2c\u7b56\u7565\u5f9e ${formatPolicyValue(from, locale)} \u8b8a\u70ba ${formatPolicyValue(to, locale)}`,
			en: `probe_policy changed from ${from.trim()} to ${to.trim()}`,
			ja: `\u30d7\u30ed\u30fc\u30d6\u65b9\u91dd\u304c ${formatPolicyValue(from, locale)} \u304b\u3089 ${formatPolicyValue(to, locale)} \u306b\u5909\u308f\u308a\u307e\u3057\u305f`,
		}),
	},
	{
		pattern: /^probe_interval changed from (\d+) to (\d+)$/,
		format: (locale: BackupLogicLocale, ...[from = '', to = '']) => t(locale, {
			'zh-Hans': `\u63a2\u6d4b\u95f4\u9694\u4ece ${from.trim()} \u53d8\u4e3a ${to.trim()}`,
			'zh-Hant': `\u63a2\u6e2c\u9593\u9694\u5f9e ${from.trim()} \u8b8a\u70ba ${to.trim()}`,
			en: `probe_interval changed from ${from.trim()} to ${to.trim()}`,
			ja: `\u30d7\u30ed\u30fc\u30d6\u9593\u9694\u304c ${from.trim()} \u304b\u3089 ${to.trim()} \u306b\u5909\u308f\u308a\u307e\u3057\u305f`,
		}),
	},
	{
		pattern: /^probe_concurrency changed from (\d+) to (\d+)$/,
		format: (locale: BackupLogicLocale, ...[from = '', to = '']) => t(locale, {
			'zh-Hans': `\u63a2\u6d4b\u5e76\u53d1\u4ece ${from.trim()} \u53d8\u4e3a ${to.trim()}`,
			'zh-Hant': `\u63a2\u6e2c\u4e26\u767c\u5f9e ${from.trim()} \u8b8a\u70ba ${to.trim()}`,
			en: `probe_concurrency changed from ${from.trim()} to ${to.trim()}`,
			ja: `\u30d7\u30ed\u30fc\u30d6\u4e26\u5217\u6570\u304c ${from.trim()} \u304b\u3089 ${to.trim()} \u306b\u5909\u308f\u308a\u307e\u3057\u305f`,
		}),
	},
	{
		pattern: /^model:(.+) concurrent probe\/race may increase cost$/,
		format: (locale: BackupLogicLocale, ...[modelName = '']) => t(locale, {
			'zh-Hans': `\u6a21\u578b ${modelName.trim()} \u7684\u5e76\u53d1\u63a2\u6d4b\u6216\u7ade\u901f\u53ef\u80fd\u589e\u52a0\u6210\u672c`,
			'zh-Hant': `\u6a21\u578b ${modelName.trim()} \u7684\u4e26\u767c\u63a2\u6e2c\u6216\u7af6\u901f\u53ef\u80fd\u589e\u52a0\u6210\u672c`,
			en: `model:${modelName.trim()} concurrent probe/race may increase cost`,
			ja: `\u30e2\u30c7\u30eb ${modelName.trim()} \u306e\u4e26\u5217\u30d7\u30ed\u30fc\u30d6\u3084\u7af6\u4e89\u5b9f\u884c\u306f\u30b3\u30b9\u30c8\u5897\u306b\u3064\u306a\u304c\u308b\u53ef\u80fd\u6027\u304c\u3042\u308a\u307e\u3059`,
		}),
	},
	{
		pattern: /^route preview diffs: (\d+)$/,
		format: (locale: BackupLogicLocale, ...[count = '0']) => t(locale, {
			'zh-Hans': `\u8def\u7531\u9884\u89c8\u53d1\u73b0 ${count.trim()} \u5904\u5dee\u5f02`,
			'zh-Hant': `\u8def\u7531\u9810\u89bd\u767c\u73fe ${count.trim()} \u8655\u5dee\u7570`,
			en: `route preview diffs: ${count.trim()}`,
			ja: `\u30eb\u30fc\u30c8\u30d7\u30ec\u30d3\u30e5\u30fc\u3067 ${count.trim()} \u4ef6\u306e\u5dee\u5206\u304c\u898b\u3064\u304b\u308a\u307e\u3057\u305f`,
		}),
	},
] as const;

function localizeKnownValue(value: string | undefined, locale: BackupLogicLocale, map: LocalizedValueMap) {
	if (!value) return value ?? '';
	const normalized = value.trim();
	const entry = map[normalized];
	if (!entry) return normalized;
	return t(locale, entry);
}

function localizeKnownToken(value: string | undefined, locale: BackupLogicLocale, map: LocalizedTokenMap) {
	if (!value) return value ?? '';
	const normalized = value.trim();
	if (locale === 'en') return normalized;
	const entry = map[normalized];
	if (!entry) return normalized;
	return t(locale, entry);
}

function localizePolicyWarning(value: string, locale: BackupLogicLocale) {
	for (const item of LOCALIZED_BACKUP_POLICY_WARNING_PATTERNS) {
		const match = value.match(item.pattern);
		if (match) return item.format(locale, ...match.slice(1));
	}
	return value;
}

function localizeDiagnosticList(values: string[] | undefined, locale: BackupLogicLocale) {
	return (values ?? []).map((value) => {
		const normalized = value?.trim() ?? '';
		const localized = localizeKnownValue(normalized, locale, BACKUP_DIAGNOSTIC_TEXT);
		if (localized !== normalized) return localized;
		return localizePolicyWarning(normalized, locale);
	});
}

function localizeTokenList(values: string[] | undefined, locale: BackupLogicLocale, map: LocalizedTokenMap) {
	return (values ?? []).map((value) => localizeKnownToken(value, locale, map));
}

function localizeContextToken(value: string | undefined, locale: BackupLogicLocale) {
	if (!value) return value ?? '';
	const normalized = value.trim();
	if (locale === 'en') return normalized;
	const parts = normalized.split(':');
	if (parts.length <= 1) return localizeKnownToken(normalized, locale, BACKUP_CONTEXT_TEXT);
	const [head, ...rest] = parts;
	return [localizeKnownToken(head, locale, BACKUP_CONTEXT_TEXT), ...rest].join(':');
}

function localizeFieldToken(value: string | undefined, locale: BackupLogicLocale) {
	if (!value) return value ?? '';
	const normalized = value.trim();
	if (locale === 'en') return normalized;
	return normalized.split('.').map((segment) => localizeKnownToken(segment, locale, BACKUP_FIELD_TEXT)).join('.');
}

function formatContexts(contexts?: string[], locale: BackupLogicLocale = 'zh-Hans') {
	return joinLocalizedList((contexts ?? []).map((context) => localizeContextToken(context, locale)), locale) || t(locale, {
		'zh-Hans': '未限定范围',
		'zh-Hant': '未限定範圍',
		en: 'unscoped',
		ja: '未スコープ',
	});
}

function formatFieldList(fields?: string[], locale: BackupLogicLocale = 'zh-Hans') {
	return joinLocalizedList((fields ?? []).map((field) => localizeFieldToken(field, locale)), locale) || t(locale, {
		'zh-Hans': '无',
		'zh-Hant': '無',
		en: 'none',
		ja: 'なし',
	});
}

function formatPolicyValue(value: string | undefined, locale: BackupLogicLocale = 'zh-Hans') {
	return localizeKnownToken(value, locale, BACKUP_ENUM_TEXT) || t(locale, {
		'zh-Hans': '未知',
		'zh-Hant': '未知',
		en: 'unknown',
		ja: '不明',
	});
}

function formatWarnings(warnings: string[] | undefined, locale: BackupLogicLocale = 'zh-Hans') {
	return joinLocalizedList(localizeDiagnosticList(warnings, locale), locale);
}

function formatModelPolicyState(state?: ModelPolicyStateLike, locale: BackupLogicLocale = 'zh-Hans') {
	if (!state) return t(locale, {
		'zh-Hans': '无',
		'zh-Hant': '無',
		en: 'n/a',
		ja: 'なし',
	});
	return [
		`${t(locale, { 'zh-Hans': '计费', 'zh-Hant': '計費', en: 'billing', ja: '課金' })}:${formatPolicyValue(state.billing_mode, locale)}`,
		`${t(locale, { 'zh-Hans': '探测', 'zh-Hant': '探測', en: 'policy', ja: 'プローブ' })}:${formatPolicyValue(state.probe_policy, locale)}`,
		`${t(locale, { 'zh-Hans': '间隔', 'zh-Hant': '間隔', en: 'interval', ja: '間隔' })}:${state.probe_interval ?? 0}`,
		`${t(locale, { 'zh-Hans': '并发', 'zh-Hant': '並發', en: 'concurrency', ja: '並列' })}:${state.probe_concurrency ?? 0}`,
	].join(', ');
}

function formatAliasPreviewMapping(item: AliasPreviewMappingLike, locale: BackupLogicLocale = 'zh-Hans') {
	return [
		`${t(locale, { 'zh-Hans': '快照模型', 'zh-Hant': '快照模型', en: 'snapshot', ja: 'スナップショット' })}:${item.snapshot_model}`,
		`${t(locale, { 'zh-Hans': '当前模型', 'zh-Hant': '目前模型', en: 'current', ja: '現在' })}:${item.current_model}`,
		`${t(locale, { 'zh-Hans': '规范名', 'zh-Hant': '規範名', en: 'canonical', ja: '正規名' })}:${item.canonical || t(locale, { 'zh-Hans': '未知', 'zh-Hant': '未知', en: 'unknown', ja: '不明' })}`,
		`${t(locale, { 'zh-Hans': '作用范围', 'zh-Hant': '作用範圍', en: 'contexts', ja: '適用範囲' })}:${formatContexts(item.contexts, locale)}`,
	].join(' | ');
}

function formatModelMappingPreview(item: ModelMappingPreviewDisplayLike, locale: BackupLogicLocale = 'zh-Hans') {
	const parts = [
		`${t(locale, { 'zh-Hans': '快照模型', 'zh-Hant': '快照模型', en: 'snapshot', ja: 'スナップショット' })}:${item.source_model}`,
		`${t(locale, { 'zh-Hans': '当前模型', 'zh-Hant': '目前模型', en: 'current', ja: '現在' })}:${item.target_model}`,
		`${t(locale, { 'zh-Hans': '引用次数', 'zh-Hant': '引用次數', en: 'usage', ja: '使用回数' })}:${item.usage_count ?? 0}`,
		`${t(locale, { 'zh-Hans': '是否使用', 'zh-Hant': '是否使用', en: 'used', ja: '使用中' })}:${item.used ? t(locale, { 'zh-Hans': '是', 'zh-Hant': '是', en: 'yes', ja: 'はい' }) : t(locale, { 'zh-Hans': '否', 'zh-Hant': '否', en: 'no', ja: 'いいえ' })}`,
		`${t(locale, { 'zh-Hans': '目标状态', 'zh-Hant': '目標狀態', en: 'target', ja: 'ターゲット' })}:${item.target_exists ? t(locale, { 'zh-Hans': '存在', 'zh-Hant': '存在', en: 'present', ja: '存在' }) : t(locale, { 'zh-Hans': '缺失', 'zh-Hant': '缺失', en: 'missing', ja: '欠落' })}`,
		`${t(locale, { 'zh-Hans': '受影响字段', 'zh-Hant': '受影響欄位', en: 'fields', ja: '変更フィールド' })}:${formatFieldList(item.touched_fields, locale)}`,
		`${t(locale, { 'zh-Hans': '作用范围', 'zh-Hant': '作用範圍', en: 'contexts', ja: '適用範囲' })}:${formatContexts(item.contexts, locale)}`,
	];
	const warnings = formatWarnings(item.warnings, locale);
	if (warnings) parts.push(`${t(locale, { 'zh-Hans': '警告', 'zh-Hant': '警告', en: 'warnings', ja: '警告' })}:${warnings}`);
	return parts.join(' | ');
}

function formatModelMappingTarget(item: ModelMappingPreviewDisplayLike, locale: BackupLogicLocale = 'zh-Hans') {
	const parts = [
		`${t(locale, { 'zh-Hans': '快照模型', 'zh-Hant': '快照模型', en: 'snapshot', ja: 'スナップショット' })}:${item.source_model}`,
		`${t(locale, { 'zh-Hans': '当前模型', 'zh-Hant': '目前模型', en: 'current', ja: '現在' })}:${item.target_model}`,
		`${t(locale, { 'zh-Hans': '作用范围', 'zh-Hant': '作用範圍', en: 'contexts', ja: '適用範囲' })}:${formatContexts(item.contexts, locale)}`,
	];
	const warnings = formatWarnings(item.warnings, locale);
	if (warnings) parts.push(`${t(locale, { 'zh-Hans': '警告', 'zh-Hant': '警告', en: 'warnings', ja: '警告' })}:${warnings}`);
	return parts.join(' | ');
}

function formatModelPolicyDiff(item: ModelPolicyDiffLike, locale: BackupLogicLocale = 'zh-Hans') {
	const parts = [
		`${t(locale, { 'zh-Hans': '模型', 'zh-Hant': '模型', en: 'model', ja: 'モデル' })}:${item.model}`,
		`${t(locale, { 'zh-Hans': '当前模型', 'zh-Hant': '目前模型', en: 'current', ja: '現在' })}:${item.current_model || item.model}`,
		`${t(locale, { 'zh-Hans': '影响级别', 'zh-Hant': '影響等級', en: 'impact', ja: '影響' })}:${formatPolicyValue(item.impact_level, locale)}`,
		`${t(locale, { 'zh-Hans': '变更字段', 'zh-Hant': '變更欄位', en: 'changed', ja: '変更項目' })}:${formatFieldList(item.changed_fields, locale)}`,
		`${t(locale, { 'zh-Hans': '变更前', 'zh-Hant': '變更前', en: 'before', ja: '変更前' })}:${formatModelPolicyState(item.before, locale)}`,
		`${t(locale, { 'zh-Hans': '变更后', 'zh-Hant': '變更後', en: 'after', ja: '変更後' })}:${formatModelPolicyState(item.after, locale)}`,
		`${t(locale, { 'zh-Hans': '作用范围', 'zh-Hant': '作用範圍', en: 'contexts', ja: '適用範囲' })}:${formatContexts(item.contexts, locale)}`,
	];
	const warnings = formatWarnings(item.warnings, locale);
	const skipReasons = joinLocalizedList(localizeDiagnosticList(item.skip_reasons, locale), locale);
	if (warnings) parts.push(`${t(locale, { 'zh-Hans': '警告', 'zh-Hant': '警告', en: 'warnings', ja: '警告' })}:${warnings}`);
	if (skipReasons) parts.push(`${t(locale, { 'zh-Hans': '跳过原因', 'zh-Hant': '略過原因', en: 'skip', ja: 'スキップ理由' })}:${skipReasons}`);
	return parts.join(' | ');
}

function formatCredentialTargetType(value: string | undefined, locale: BackupLogicLocale = 'zh-Hans') {
	switch ((value ?? '').trim()) {
		case 'channel_key':
			return t(locale, { 'zh-Hans': '渠道密钥', 'zh-Hant': '渠道金鑰', en: 'channel key', ja: 'チャネルキー' });
		case 'api_key':
			return t(locale, { 'zh-Hans': 'API 密钥', 'zh-Hant': 'API 金鑰', en: 'api key', ja: 'API キー' });
		default:
			return t(locale, { 'zh-Hans': '凭证', 'zh-Hant': '憑證', en: 'credential', ja: '認証情報' });
	}
}

function formatCredentialRebindTarget(item: CredentialRebindTargetLike, locale: BackupLogicLocale = 'zh-Hans') {
	const parts: string[] = [];
	parts.push(`${t(locale, { 'zh-Hans': '目标类型', 'zh-Hant': '目標類型', en: 'target', ja: '対象' })}:${formatCredentialTargetType(item.target_type, locale)}`);
	if (item.channel_name) parts.push(`${t(locale, { 'zh-Hans': '渠道', 'zh-Hant': '渠道', en: 'channel', ja: 'チャネル' })}:${item.channel_name}`);
	if (item.key_name) parts.push(`${t(locale, { 'zh-Hans': '密钥', 'zh-Hant': '金鑰', en: 'key', ja: 'キー' })}:${item.key_name}`);
	if (item.source_type) parts.push(`${t(locale, { 'zh-Hans': '来源', 'zh-Hant': '來源', en: 'source', ja: 'ソース' })}:${item.source_type}`);
	if (item.models?.length) parts.push(`${t(locale, { 'zh-Hans': '模型', 'zh-Hant': '模型', en: 'models', ja: 'モデル' })}:${joinLocalizedList(item.models, locale)}`);
	if (item.affected_groups?.length) parts.push(`${t(locale, { 'zh-Hans': '影响分组', 'zh-Hant': '影響分組', en: 'groups', ja: 'グループ' })}:${joinLocalizedList(item.affected_groups, locale)}`);
	if (item.contexts?.length) parts.push(`${t(locale, { 'zh-Hans': '作用范围', 'zh-Hant': '作用範圍', en: 'contexts', ja: '適用範囲' })}:${formatContexts(item.contexts, locale)}`);
	const warnings = formatWarnings(item.warnings, locale);
	if (warnings) parts.push(`${t(locale, { 'zh-Hans': '警告', 'zh-Hant': '警告', en: 'warnings', ja: '警告' })}:${warnings}`);
	return parts.join(' | ');
}

function formatRouteTargetIssue(item: RouteTargetPreviewIssueLike, locale: BackupLogicLocale = 'zh-Hans') {
	const parts: string[] = [];
	if (item.group_name) parts.push(`${t(locale, { 'zh-Hans': '分组', 'zh-Hant': '分組', en: 'group', ja: 'グループ' })}:${item.group_name}`);
	if (item.channel_name) parts.push(`${t(locale, { 'zh-Hans': '渠道', 'zh-Hant': '渠道', en: 'channel', ja: 'チャネル' })}:${item.channel_name}`);
	if (item.model) parts.push(`${t(locale, { 'zh-Hans': '模型', 'zh-Hant': '模型', en: 'model', ja: 'モデル' })}:${item.model}`);
	if (item.resolved_model && item.resolved_model !== item.model) parts.push(`${t(locale, { 'zh-Hans': '解析模型', 'zh-Hant': '解析模型', en: 'resolved', ja: '解決後モデル' })}:${item.resolved_model}`);
	parts.push(`${t(locale, { 'zh-Hans': '问题类型', 'zh-Hant': '問題類型', en: 'issue', ja: '問題種別' })}:${item.issue_type}`);
	if (item.reason) parts.push(`${t(locale, { 'zh-Hans': '原因', 'zh-Hant': '原因', en: 'reason', ja: '理由' })}:${item.reason}`);
	if (item.action) parts.push(`${t(locale, { 'zh-Hans': '建议动作', 'zh-Hant': '建議動作', en: 'action', ja: '対応' })}:${item.action}`);
	return parts.join(' | ');
}

function formatRoutePreviewCandidate(item: RoutePreviewCandidateLike, locale: BackupLogicLocale = 'zh-Hans') {
	const channel = item.channel_name?.trim() || t(locale, {
		'zh-Hans': '未知渠道',
		'zh-Hant': '未知渠道',
		en: 'unknown channel',
		ja: '不明なチャネル',
	});
	const model = (item.resolved_model || item.model || '').trim() || t(locale, {
		'zh-Hans': '未知模型',
		'zh-Hant': '未知模型',
		en: 'unknown model',
		ja: '不明なモデル',
	});
	const parts = [`${channel}:${model}`];
	if (typeof item.priority === 'number') parts.push(`${t(locale, { 'zh-Hans': '优先级', 'zh-Hant': '優先級', en: 'priority', ja: '優先度' })}:${item.priority}`);
	if (typeof item.weight === 'number') parts.push(`${t(locale, { 'zh-Hans': '权重', 'zh-Hant': '權重', en: 'weight', ja: '重み' })}:${item.weight}`);
	if (item.reason) parts.push(`${t(locale, { 'zh-Hans': '原因', 'zh-Hant': '原因', en: 'reason', ja: '理由' })}:${localizeDiagnosticList([item.reason], locale)[0] ?? item.reason}`);
	return parts.join(' | ');
}

function formatRoutePreviewCandidateList(candidates: RoutePreviewCandidateLike[] | undefined, locale: BackupLogicLocale = 'zh-Hans') {
	const items = (candidates ?? []).map((item) => formatRoutePreviewCandidate(item, locale));
	return joinLocalizedList(items, locale) || t(locale, {
		'zh-Hans': '无',
		'zh-Hant': '無',
		en: 'none',
		ja: 'なし',
	});
}

function formatRoutePreviewDiff(item: RoutePreviewDiffLike, locale: BackupLogicLocale = 'zh-Hans') {
	const parts: string[] = [];
	if (item.group_name) parts.push(`${t(locale, { 'zh-Hans': '分组', 'zh-Hant': '分組', en: 'group', ja: 'グループ' })}:${item.group_name}`);
	if (item.model) parts.push(`${t(locale, { 'zh-Hans': '模型', 'zh-Hant': '模型', en: 'model', ja: 'モデル' })}:${item.model}`);
	parts.push(`${t(locale, { 'zh-Hans': '当前候选', 'zh-Hant': '目前候選', en: 'before', ja: '現在候補' })}:${formatRoutePreviewCandidateList(item.before_candidates, locale)}`);
	parts.push(`${t(locale, { 'zh-Hans': '快照候选', 'zh-Hant': '快照候選', en: 'after', ja: 'スナップショット候補' })}:${formatRoutePreviewCandidateList(item.after_candidates, locale)}`);
	if ((item.removed_candidates ?? []).length > 0) parts.push(`${t(locale, { 'zh-Hans': '将被移除', 'zh-Hant': '將被移除', en: 'removed', ja: '削除候補' })}:${formatRoutePreviewCandidateList(item.removed_candidates, locale)}`);
	if ((item.added_candidates ?? []).length > 0) parts.push(`${t(locale, { 'zh-Hans': '将被新增', 'zh-Hant': '將被新增', en: 'added', ja: '追加候補' })}:${formatRoutePreviewCandidateList(item.added_candidates, locale)}`);
	if (item.fallback_changed) parts.push(`${t(locale, { 'zh-Hans': '回退链变化', 'zh-Hant': '回退鏈變化', en: 'fallback changed', ja: 'フォールバック変化' })}:${t(locale, { 'zh-Hans': '是', 'zh-Hant': '是', en: 'yes', ja: 'はい' })}`);
	const skipReasons = joinLocalizedList(localizeDiagnosticList(item.skip_reasons, locale), locale);
	if (skipReasons) parts.push(`${t(locale, { 'zh-Hans': '跳过原因', 'zh-Hant': '略過原因', en: 'skip', ja: 'スキップ理由' })}:${skipReasons}`);
	return parts.join(' | ');
}

function collectExampleText(items: string[], locale: BackupLogicLocale = 'zh-Hans') {
	const visible = items.map((item) => item.trim()).filter(Boolean);
	if (visible.length === 0) return '';
	return visible.slice(0, 2).join(locale === 'en' ? '; ' : '；');
}

export function getAliasPreviewItems(previewMappings?: readonly unknown[], locale: BackupLogicLocale = 'zh-Hans') {
	return (previewMappings ?? []).map((item) => normalizeLocalizedDetailText(formatAliasPreviewMapping(item as AliasPreviewMappingLike, locale), locale));
}

export function getModelMappingPreviewItems(previewMappings?: readonly unknown[], locale: BackupLogicLocale = 'zh-Hans') {
	return (previewMappings ?? []).map((item) => normalizeLocalizedDetailText(formatModelMappingPreview(item as ModelMappingPreviewDisplayLike, locale), locale));
}

export function getMissingModelMappingItems(previewMappings?: readonly unknown[], locale: BackupLogicLocale = 'zh-Hans') {
	return (previewMappings ?? [])
		.filter((item) => (item as ModelMappingPreviewDisplayLike).used && (item as ModelMappingPreviewDisplayLike).target_exists === false)
		.map((item) => normalizeLocalizedDetailText(formatModelMappingTarget(item as ModelMappingPreviewDisplayLike, locale), locale));
}

export function getUnusedModelMappingItems(previewMappings?: readonly unknown[], locale: BackupLogicLocale = 'zh-Hans') {
	return (previewMappings ?? [])
		.filter((item) => !(item as ModelMappingPreviewDisplayLike).used)
		.map((item) => normalizeLocalizedDetailText(formatModelMappingTarget(item as ModelMappingPreviewDisplayLike, locale), locale));
}

export function getCredentialRebindTargetItems(rebindTargets?: readonly unknown[], locale: BackupLogicLocale = 'zh-Hans') {
	return (rebindTargets ?? []).map((item) => normalizeLocalizedDetailText(formatCredentialRebindTarget(item as CredentialRebindTargetLike, locale), locale));
}

export function getRouteTargetIssueItems(routeTargetIssues?: readonly unknown[], locale: BackupLogicLocale = 'zh-Hans') {
	return (routeTargetIssues ?? []).map((item) => normalizeLocalizedDetailText(formatRouteTargetIssue(item as RouteTargetPreviewIssueLike, locale), locale));
}

export function getRoutePreviewWarningItems(warnings?: readonly unknown[], locale: BackupLogicLocale = 'zh-Hans') {
	return localizeDiagnosticList((warnings ?? []).map((item) => String(item ?? '')), locale).map((item) => normalizeLocalizedDetailText(item, locale));
}

export function getMergedRoutePreviewWarningItems(
	warningGroups: ReadonlyArray<readonly unknown[] | undefined>,
	locale: BackupLogicLocale = 'zh-Hans',
) {
	const merged: string[] = [];
	const seen = new Set<string>();

	for (const warnings of warningGroups) {
		for (const item of warnings ?? []) {
			const raw = String(item ?? '').trim();
			if (!raw || seen.has(raw)) {
				continue;
			}
			seen.add(raw);
			merged.push(raw);
		}
	}

	return getRoutePreviewWarningItems(merged, locale);
}

export function getRoutePreviewDiffItems(diffs?: readonly unknown[], locale: BackupLogicLocale = 'zh-Hans') {
	return (diffs ?? []).map((item) => normalizeLocalizedDetailText(formatRoutePreviewDiff(item as RoutePreviewDiffLike, locale), locale));
}

export function getCompatibilityDiagnosticItems(items?: readonly unknown[], locale: BackupLogicLocale = 'zh-Hans') {
	return localizeDiagnosticList((items ?? []).map((item) => String(item ?? '').trim()).filter(Boolean), locale)
		.map((item) => normalizeLocalizedDetailText(item, locale));
}

export function getCompatibilityNameItems(items?: readonly unknown[], locale: BackupLogicLocale = 'zh-Hans') {
	return (items ?? [])
		.map((item) => normalizeLocalizedDetailText(String(item ?? '').trim(), locale))
		.filter(Boolean);
}

export function getModelPolicyDiffItems(diffs?: readonly unknown[], locale: BackupLogicLocale = 'zh-Hans') {
	return (diffs ?? []).map((item) => normalizeLocalizedDetailText(formatModelPolicyDiff(item as ModelPolicyDiffLike, locale), locale));
}

export function getReplacePrunedBreakdownItems(breakdown: ReplacePrunedBreakdownLike | undefined, locale: BackupLogicLocale = 'zh-Hans') {
	if (!breakdown) return {
		channels: [] as string[],
		groups: [] as string[],
		settings: [] as string[],
		llmInfos: [] as string[],
		apiKeys: [] as string[],
	};
	return {
		channels: getCompatibilityNameItems(breakdown.channels, locale),
		groups: getCompatibilityNameItems(breakdown.groups, locale),
		settings: getCompatibilityDiagnosticItems(breakdown.settings, locale),
		llmInfos: getCompatibilityNameItems(breakdown.models, locale),
		apiKeys: getCompatibilityNameItems(breakdown.apiKeys, locale),
	};
}

export function getCompatibilityCounts(compatibility: CompatibilityLike | undefined): CompatibilityCounts {
	const modelMappingPreviews = compatibility?.model_mapping_previews ?? [];
	const credentialRebindTargets = compatibility?.credential_rebind_targets ?? [];
	const rawCredentialRebindTargets = getCompatibilityCount(compatibility?.summary?.credential_rebind_targets, compatibility?.credential_rebind_targets);
	const channelKeyRebindTargets = getSummaryOrFallbackCount(
		compatibility?.summary?.channel_key_rebind_targets,
		credentialRebindTargets.filter((item) => item.target_type === 'channel_key').length,
	);
	const apiKeyRebindTargets = getSummaryOrFallbackCount(
		compatibility?.summary?.api_key_rebind_targets,
		credentialRebindTargets.filter((item) => item.target_type === 'api_key').length,
	);

	return {
		conflicts: getCompatibilityCount(compatibility?.summary?.conflicts, compatibility?.conflicts),
		aliasConflicts: getCompatibilityCount(compatibility?.summary?.alias_conflicts, compatibility?.alias_conflicts),
		routeConflicts: getCompatibilityCount(compatibility?.summary?.route_conflicts, compatibility?.route_conflicts),
		credentialRebindTargets: getEffectiveCredentialRebindCount({
			credentialRebindTargets: rawCredentialRebindTargets,
			channelKeyRebindTargets,
			apiKeyRebindTargets,
		}),
		channelKeyRebindTargets,
		apiKeyRebindTargets,
		invalidRouteTargets: getCompatibilityCount(compatibility?.summary?.invalid_route_targets, compatibility?.invalid_route_targets),
		skippedRouteTargetPreviews: getCompatibilityCount(compatibility?.summary?.skipped_route_target_previews, compatibility?.skipped_route_target_previews),
		routePreviewWarnings: getCompatibilityCount(compatibility?.summary?.route_preview_warnings, compatibility?.route_preview_warnings),
		routePreviewDiffs: getCompatibilityCount(compatibility?.summary?.route_preview_diffs, compatibility?.route_preview_diffs),
		missingProviders: getCompatibilityCount(compatibility?.summary?.missing_providers, compatibility?.missing_providers),
		missingModels: getCompatibilityCount(compatibility?.summary?.missing_models, compatibility?.missing_models),
		baseURLMismatches: getCompatibilityCount(compatibility?.summary?.base_url_mismatches, compatibility?.base_url_mismatches),
		schemaMismatches: getCompatibilityCount(compatibility?.summary?.schema_mismatches, compatibility?.schema_mismatches),
		skippedTargets: getCompatibilityCount(compatibility?.summary?.skipped_targets, compatibility?.skipped_targets),
		modelMappingPreviews: getCompatibilityCount(compatibility?.summary?.model_mapping_previews, compatibility?.model_mapping_previews),
		usedModelMappings: getSummaryOrFallbackCount(
			compatibility?.summary?.used_model_mappings,
			modelMappingPreviews.filter((item) => item.used).length,
		),
		unusedModelMappings: getSummaryOrFallbackCount(
			compatibility?.summary?.unused_model_mappings,
			modelMappingPreviews.filter((item) => !item.used).length,
		),
		missingMappingTargets: getSummaryOrFallbackCount(
			compatibility?.summary?.missing_mapping_targets,
			modelMappingPreviews.filter((item) => item.used && item.target_exists === false).length,
		),
		aliasPreviewMappings: getCompatibilityCount(compatibility?.summary?.alias_preview_mappings, compatibility?.alias_preview_mappings),
		modelPolicyDiffs: getCompatibilityCount(compatibility?.summary?.model_policy_diffs, compatibility?.model_policy_diffs),
		replacePrunedTargets: countStructuredReplacePrunedItems(compatibility),
	};
}

export function getCompatibilityOverview(input: CompatibilityOverviewInput) {
	const counts = input.counts;
	const locale = (input as CompatibilityOverviewInput & { locale?: BackupLogicLocale }).locale ?? 'zh-Hans';
	const totalCredentialRebindTargets = getEffectiveCredentialRebindCount(counts);
	let tone: SummaryTone = 'safe';

	if (
		counts.conflicts > 0
		|| counts.routeConflicts > 0
		|| counts.baseURLMismatches > 0
		|| counts.schemaMismatches > 0
		|| counts.invalidRouteTargets > 0
	) {
		tone = 'danger';
	} else if (
		counts.aliasConflicts > 0
		|| totalCredentialRebindTargets > 0
		|| counts.missingProviders > 0
		|| counts.missingModels > 0
		|| counts.skippedTargets > 0
		|| counts.replacePrunedTargets > 0
		|| counts.routePreviewWarnings > 0
		|| counts.skippedRouteTargetPreviews > 0
		|| counts.routePreviewDiffs > 0
		|| counts.aliasPreviewMappings > 0
		|| counts.modelPolicyDiffs > 0
		|| input.warningsCount > 0
	) {
		tone = 'warning';
	}

	if (input.kind === 'rollback') {
		if (tone === 'danger') {
			return {
				...counts,
				tone,
				title: t(locale, {
					'zh-Hans': '回滚存在阻断风险',
					'zh-Hant': '回滾存在阻斷風險',
					en: 'Rollback has blocking risks',
					ja: 'ロールバックに阻害リスクがあります',
				}),
				description: t(locale, {
					'zh-Hans': '回滚预检发现了路由或结构层面的风险，可能导致恢复行为变化，或留下未解决的目标。',
					'zh-Hant': '回滾預檢發現路由或結構層面的風險，可能導致恢復行為改變，或留下未解決的目標。',
					en: 'Rollback preview found route or schema risks that can change restore behavior or leave targets unresolved.',
					ja: 'ロールバックのプレビューで、復元動作に影響したり未解決の対象を残したりするルートまたはスキーマのリスクが見つかりました。',
				}),
			};
		}
		if (tone === 'warning') {
			return {
				...counts,
				tone,
				title: t(locale, {
					'zh-Hans': '回滚前需要复核',
					'zh-Hant': '回滾前需要複核',
					en: 'Rollback needs review',
					ja: 'ロールバック前に確認が必要です',
				}),
				description: t(locale, {
					'zh-Hans': '回滚预检发现了兼容性差异，建议在恢复前先确认。',
					'zh-Hant': '回滾預檢發現相容性差異，建議在恢復前先確認。',
					en: 'Rollback preview found compatibility differences that should be reviewed before restore.',
					ja: 'ロールバックのプレビューで、復元前に確認すべき互換差分が見つかりました。',
				}),
			};
		}
		return {
			...counts,
			tone,
			title: t(locale, {
				'zh-Hans': '回滚预检整体安全',
				'zh-Hant': '回滾預檢整體安全',
				en: 'Rollback looks safe',
				ja: 'ロールバックは概ね安全です',
			}),
			description: t(locale, {
				'zh-Hans': '回滚预检未发现明显的路由或结构风险，但执行后仍会恢复为快照中的状态。',
				'zh-Hant': '回滾預檢未發現明顯的路由或結構風險，但執行後仍會恢復為快照中的狀態。',
				en: 'Rollback preview did not find obvious route or schema risks. It still restores snapshot state.',
				ja: 'ロールバックのプレビューでは明確なルートやスキーマのリスクは見つかりませんでしたが、実行するとスナップショットの状態へ復元されます。',
			}),
		};
	}

	if (tone === 'danger') {
		return {
			...counts,
			tone,
			title: t(locale, {
				'zh-Hans': '导入存在阻断风险',
				'zh-Hant': '導入存在阻斷風險',
				en: 'Import has blocking risks',
				ja: 'インポートに阻害リスクがあります',
			}),
			description: t(locale, {
				'zh-Hans': '预检发现了阻断性的冲突，或路由/结构层面的风险。建议先修正，再执行导入。',
				'zh-Hant': '預檢發現阻斷性的衝突，或路由/結構層面的風險。建議先修正，再執行導入。',
				en: 'Dry-run found blocking conflicts or route/schema risks. Fix them before applying.',
				ja: 'ドライランで、適用前に修正すべき競合やルート/スキーマのリスクが見つかりました。',
			}),
		};
	}
	if (tone === 'warning') {
		return {
			...counts,
			tone,
			title: t(locale, {
				'zh-Hans': '导入前需要复核',
				'zh-Hant': '導入前需要複核',
				en: 'Import needs review',
				ja: 'インポート前に確認が必要です',
			}),
			description: t(locale, {
				'zh-Hans': '预检发现了一些兼容性提示。建议在应用前复核模型、提供商和路由目标。',
				'zh-Hant': '預檢發現一些相容性提示。建議在套用前複核模型、供應商與路由目標。',
				en: 'Dry-run found compatibility hints. Review models, providers, and route targets before applying.',
				ja: 'ドライランで互換性に関する注意が見つかりました。適用前にモデル、プロバイダー、ルート対象を確認してください。',
			}),
		};
	}
	return {
		...counts,
		tone,
		title: t(locale, {
			'zh-Hans': '未发现阻断问题',
			'zh-Hant': '未發現阻斷問題',
			en: 'No blocking issues found',
			ja: '阻害問題は見つかりませんでした',
		}),
		description: t(locale, {
			'zh-Hans': '预检未发现阻断性的兼容风险，但应用后仍会把快照内容写入当前项目。',
			'zh-Hant': '預檢未發現阻斷性的相容風險，但套用後仍會把快照內容寫入目前專案。',
			en: 'Dry-run did not find blocking compatibility risks. Applying will still write the snapshot into the current project.',
			ja: 'ドライランでは阻害的な互換リスクは見つかりませんでしたが、適用すると現在のプロジェクトへスナップショット内容が書き込まれます。',
		}),
	};
}

export function buildCompatibilitySignalItems(input: CompatibilitySignalItemsInput) {
	const {
		kind,
		counts,
		warningsCount = 0,
		replacePruneCount = 0,
		structuredReplacePrunedCount = 0,
		includeWarningsCount = true,
		includeReplaceModeRisk = false,
		includeModelMappingPreviews = true,
		includeUnusedModelMappings = true,
		includeStructuredReplacePrunedCount = false,
		effectiveMode,
		importWarningsLabel: rawImportWarningsLabel,
	} = input;
	const locale = (input as CompatibilitySignalItemsInput & { locale?: BackupLogicLocale }).locale ?? 'zh-Hans';
	const importWarningsLabel = rawImportWarningsLabel?.trim() || t(locale, {
		'zh-Hans': '导入报告',
		'zh-Hant': '導入報告',
		en: 'Import report',
		ja: 'インポートレポート',
	});

	const targetLabel = kind === 'rollback'
		? t(locale, { 'zh-Hans': '回滚目标', 'zh-Hant': '回滾目標', en: 'restored targets', ja: '復元対象' })
		: t(locale, { 'zh-Hans': '导入目标', 'zh-Hant': '導入目標', en: 'imported targets', ja: 'インポート対象' });
	const totalCredentialRebindTargets = getEffectiveCredentialRebindCount(counts);
	const genericCredentialRebindTargets = Math.max(
		totalCredentialRebindTargets - counts.channelKeyRebindTargets - counts.apiKeyRebindTargets,
		0,
	);
	const items: string[] = [];

	if (includeReplaceModeRisk && effectiveMode === 'replace') {
		items.push(t(locale, {
			'zh-Hans': '替换模式会移除当前项目中那些未被快照保留的记录。',
			'zh-Hant': '替換模式會移除目前專案中那些未被快照保留的記錄。',
			en: 'Replace mode can remove current project records that are not kept by the snapshot.',
			ja: '置換モードでは、スナップショットに含まれない現在のプロジェクト記録が削除される可能性があります。',
		}));
	}
	if (includeWarningsCount && warningsCount > 0) {
		items.push(kind === 'rollback'
			? t(locale, {
				'zh-Hans': `回滚预检产生了 ${warningsCount} 条警告。`,
				'zh-Hant': `回滾預檢產生了 ${warningsCount} 條警告。`,
				en: `Rollback preview emitted ${warningsCount} warnings.`,
				ja: `ロールバックのプレビューで ${warningsCount} 件の警告が出ました。`,
			})
			: t(locale, {
				'zh-Hans': `${importWarningsLabel}产生了 ${warningsCount} 条警告。`,
				'zh-Hant': `${importWarningsLabel}產生了 ${warningsCount} 條警告。`,
				en: `${importWarningsLabel} emitted ${warningsCount} warnings.`,
				ja: `${importWarningsLabel} で ${warningsCount} 件の警告が出ました。`,
			}));
	}
	if (replacePruneCount > 0) {
		items.push(kind === 'rollback'
			? t(locale, {
				'zh-Hans': `回滚清理预览会删除或重置当前项目中的 ${replacePruneCount} 条记录。`,
				'zh-Hant': `回滾清理預覽會刪除或重設目前專案中的 ${replacePruneCount} 條記錄。`,
				en: `Rollback prune preview will delete or reset ${replacePruneCount} current records.`,
				ja: `ロールバックの整理プレビューでは、現在の記録 ${replacePruneCount} 件が削除またはリセットされます。`,
			})
			: t(locale, {
				'zh-Hans': `替换清理预览会删除或重置当前项目中的 ${replacePruneCount} 条记录。`,
				'zh-Hant': `替換清理預覽會刪除或重設目前專案中的 ${replacePruneCount} 條記錄。`,
				en: `Replace-prune preview will delete or reset ${replacePruneCount} current records.`,
				ja: `置換クリーンプレビューでは、現在の記録 ${replacePruneCount} 件が削除またはリセットされます。`,
			}));
	}
	if (includeStructuredReplacePrunedCount && structuredReplacePrunedCount > 0) {
		items.push(kind === 'rollback'
			? t(locale, {
				'zh-Hans': `回滚兼容性明细标记了 ${structuredReplacePrunedCount} 条会被删除或重置的当前记录。`,
				'zh-Hant': `回滾相容性明細標記了 ${structuredReplacePrunedCount} 條會被刪除或重設的目前記錄。`,
				en: `Rollback diagnostics marked ${structuredReplacePrunedCount} current records for removal or reset.`,
				ja: `ロールバック診断で、削除またはリセット対象の現在の記録 ${structuredReplacePrunedCount} 件が示されました。`,
			})
			: t(locale, {
				'zh-Hans': `结构化清理预览还发现了 ${structuredReplacePrunedCount} 条额外记录。`,
				'zh-Hant': `結構化清理預覽還發現了 ${structuredReplacePrunedCount} 條額外記錄。`,
				en: `Structured prune preview found ${structuredReplacePrunedCount} additional records.`,
				ja: `構造化された整理プレビューで追加の ${structuredReplacePrunedCount} 件が見つかりました。`,
			}));
	}
	if (counts.conflicts > 0) {
		items.push(kind === 'rollback'
			? t(locale, {
				'zh-Hans': `回滚预检发现 ${counts.conflicts} 个冲突。`,
				'zh-Hant': `回滾預檢發現 ${counts.conflicts} 個衝突。`,
				en: `Rollback preview found ${counts.conflicts} conflicts.`,
				ja: `ロールバックのプレビューで ${counts.conflicts} 件の競合が見つかりました。`,
			})
			: t(locale, {
				'zh-Hans': `兼容性报告发现 ${counts.conflicts} 个冲突。`,
				'zh-Hant': `相容性報告發現 ${counts.conflicts} 個衝突。`,
				en: `Compatibility report found ${counts.conflicts} conflicts.`,
				ja: `互換レポートで ${counts.conflicts} 件の競合が見つかりました。`,
			}));
	}
	if (counts.aliasConflicts > 0) {
		items.push(kind === 'rollback'
			? t(locale, {
				'zh-Hans': `回滚预检发现 ${counts.aliasConflicts} 个别名冲突。`,
				'zh-Hant': `回滾預檢發現 ${counts.aliasConflicts} 個別名衝突。`,
				en: `Rollback preview found ${counts.aliasConflicts} alias conflicts.`,
				ja: `ロールバックのプレビューで ${counts.aliasConflicts} 件のエイリアス競合が見つかりました。`,
			})
			: t(locale, {
				'zh-Hans': `兼容性报告发现 ${counts.aliasConflicts} 个别名冲突。`,
				'zh-Hant': `相容性報告發現 ${counts.aliasConflicts} 個別名衝突。`,
				en: `Compatibility report found ${counts.aliasConflicts} alias conflicts.`,
				ja: `互換レポートで ${counts.aliasConflicts} 件のエイリアス競合が見つかりました。`,
			}));
	}
	if (counts.routeConflicts > 0) {
		items.push(kind === 'rollback'
			? t(locale, {
				'zh-Hans': `回滚预检发现 ${counts.routeConflicts} 个路由冲突。`,
				'zh-Hant': `回滾預檢發現 ${counts.routeConflicts} 個路由衝突。`,
				en: `Rollback preview found ${counts.routeConflicts} route conflicts.`,
				ja: `ロールバックのプレビューで ${counts.routeConflicts} 件のルート競合が見つかりました。`,
			})
			: t(locale, {
				'zh-Hans': `兼容性报告发现 ${counts.routeConflicts} 个路由冲突。`,
				'zh-Hant': `相容性報告發現 ${counts.routeConflicts} 個路由衝突。`,
				en: `Compatibility report found ${counts.routeConflicts} route conflicts.`,
				ja: `互換レポートで ${counts.routeConflicts} 件のルート競合が見つかりました。`,
			}));
	}
	if (genericCredentialRebindTargets > 0) {
		items.push(t(locale, {
			'zh-Hans': `${genericCredentialRebindTargets} 个${targetLabel}需要重新绑定凭证。`,
			'zh-Hant': `${genericCredentialRebindTargets} 個${targetLabel}需要重新綁定憑證。`,
			en: `Credential rebind is required for ${genericCredentialRebindTargets} ${targetLabel}.`,
			ja: `${targetLabel} ${genericCredentialRebindTargets} 件で認証情報の再バインドが必要です。`,
		}));
	}
	if (counts.channelKeyRebindTargets > 0) {
		items.push(t(locale, {
			'zh-Hans': `${counts.channelKeyRebindTargets} 个${targetLabel}需要重新绑定渠道密钥凭证。`,
			'zh-Hant': `${counts.channelKeyRebindTargets} 個${targetLabel}需要重新綁定渠道金鑰憑證。`,
			en: `Channel-key credential rebind is required for ${counts.channelKeyRebindTargets} ${targetLabel}.`,
			ja: `${targetLabel} ${counts.channelKeyRebindTargets} 件でチャネルキー認証の再バインドが必要です。`,
		}));
	}
	if (counts.apiKeyRebindTargets > 0) {
		items.push(t(locale, {
			'zh-Hans': `${counts.apiKeyRebindTargets} 个${targetLabel}需要重新绑定 API 密钥凭证。`,
			'zh-Hant': `${counts.apiKeyRebindTargets} 個${targetLabel}需要重新綁定 API 金鑰憑證。`,
			en: `API-key credential rebind is required for ${counts.apiKeyRebindTargets} ${targetLabel}.`,
			ja: `${targetLabel} ${counts.apiKeyRebindTargets} 件で API キー認証の再バインドが必要です。`,
		}));
	}
	if (counts.missingProviders > 0) {
		items.push(t(locale, {
			'zh-Hans': `兼容性报告发现 ${counts.missingProviders} 个缺失的提供商。`,
			'zh-Hant': `相容性報告發現 ${counts.missingProviders} 個缺失的供應商。`,
			en: `Compatibility report found ${counts.missingProviders} missing providers.`,
			ja: `互換レポートで ${counts.missingProviders} 件の不足プロバイダーが見つかりました。`,
		}));
	}
	if (counts.missingModels > 0) {
		items.push(t(locale, {
			'zh-Hans': `兼容性报告发现 ${counts.missingModels} 个缺失的模型。`,
			'zh-Hant': `相容性報告發現 ${counts.missingModels} 個缺失的模型。`,
			en: `Compatibility report found ${counts.missingModels} missing models.`,
			ja: `互換レポートで ${counts.missingModels} 件の不足モデルが見つかりました。`,
		}));
	}
	if (counts.baseURLMismatches > 0) {
		items.push(t(locale, {
			'zh-Hans': `兼容性报告发现 ${counts.baseURLMismatches} 个基础地址不匹配。`,
			'zh-Hant': `相容性報告發現 ${counts.baseURLMismatches} 個基礎位址不匹配。`,
			en: `Compatibility report found ${counts.baseURLMismatches} base-URL mismatches.`,
			ja: `互換レポートで ${counts.baseURLMismatches} 件のベース URL 不一致が見つかりました。`,
		}));
	}
	if (counts.schemaMismatches > 0) {
		items.push(t(locale, {
			'zh-Hans': `兼容性报告发现 ${counts.schemaMismatches} 个结构不匹配。`,
			'zh-Hant': `相容性報告發現 ${counts.schemaMismatches} 個結構不匹配。`,
			en: `Compatibility report found ${counts.schemaMismatches} schema mismatches.`,
			ja: `互換レポートで ${counts.schemaMismatches} 件のスキーマ不一致が見つかりました。`,
		}));
	}
	if (counts.skippedTargets > 0) {
		items.push(t(locale, {
			'zh-Hans': `兼容性报告跳过了 ${counts.skippedTargets} 个目标。`,
			'zh-Hant': `相容性報告略過了 ${counts.skippedTargets} 個目標。`,
			en: `Compatibility report skipped ${counts.skippedTargets} targets.`,
			ja: `互換レポートで ${counts.skippedTargets} 件の対象がスキップされました。`,
		}));
	}
	if (counts.invalidRouteTargets > 0) {
		items.push(t(locale, {
			'zh-Hans': `兼容性报告发现 ${counts.invalidRouteTargets} 个无效的路由目标。`,
			'zh-Hant': `相容性報告發現 ${counts.invalidRouteTargets} 個無效的路由目標。`,
			en: `Compatibility report found ${counts.invalidRouteTargets} invalid route targets.`,
			ja: `互換レポートで ${counts.invalidRouteTargets} 件の無効なルート対象が見つかりました。`,
		}));
	}
	if (counts.skippedRouteTargetPreviews > 0) {
		items.push(t(locale, {
			'zh-Hans': `兼容性报告跳过了 ${counts.skippedRouteTargetPreviews} 个路由目标预览。`,
			'zh-Hant': `相容性報告略過了 ${counts.skippedRouteTargetPreviews} 個路由目標預覽。`,
			en: `Compatibility report skipped ${counts.skippedRouteTargetPreviews} route-target previews.`,
			ja: `互換レポートで ${counts.skippedRouteTargetPreviews} 件のルート対象プレビューがスキップされました。`,
		}));
	}
	if (counts.routePreviewWarnings > 0) {
		items.push(t(locale, {
			'zh-Hans': `路由预览产生了 ${counts.routePreviewWarnings} 条警告。`,
			'zh-Hant': `路由預覽產生了 ${counts.routePreviewWarnings} 條警告。`,
			en: `Route preview emitted ${counts.routePreviewWarnings} warnings.`,
			ja: `ルートプレビューで ${counts.routePreviewWarnings} 件の警告が出ました。`,
		}));
	}
	if (counts.routePreviewDiffs > 0) {
		items.push(t(locale, {
			'zh-Hans': `路由预览发现 ${counts.routePreviewDiffs} 处差异。`,
			'zh-Hant': `路由預覽發現 ${counts.routePreviewDiffs} 處差異。`,
			en: `Route preview found ${counts.routePreviewDiffs} diffs.`,
			ja: `ルートプレビューで ${counts.routePreviewDiffs} 件の差分が見つかりました。`,
		}));
	}
	if (includeModelMappingPreviews && counts.modelMappingPreviews > 0) {
		items.push(t(locale, {
			'zh-Hans': `兼容性报告发现 ${counts.modelMappingPreviews} 条模型映射预览。`,
			'zh-Hant': `相容性報告發現 ${counts.modelMappingPreviews} 條模型映射預覽。`,
			en: `Compatibility report found ${counts.modelMappingPreviews} model-mapping previews.`,
			ja: `互換レポートで ${counts.modelMappingPreviews} 件のモデルマッピングプレビューが見つかりました。`,
		}));
	}
	if (counts.missingMappingTargets > 0) {
		items.push(t(locale, {
			'zh-Hans': `兼容性报告发现 ${counts.missingMappingTargets} 个缺失的映射目标。`,
			'zh-Hant': `相容性報告發現 ${counts.missingMappingTargets} 個缺失的映射目標。`,
			en: `Compatibility report found ${counts.missingMappingTargets} missing mapping targets.`,
			ja: `互換レポートで ${counts.missingMappingTargets} 件の不足マッピング先が見つかりました。`,
		}));
	}
	if (includeUnusedModelMappings && counts.unusedModelMappings > 0) {
		items.push(t(locale, {
			'zh-Hans': `兼容性报告发现 ${counts.unusedModelMappings} 条未使用的模型映射。`,
			'zh-Hant': `相容性報告發現 ${counts.unusedModelMappings} 條未使用的模型映射。`,
			en: `Compatibility report found ${counts.unusedModelMappings} unused model mappings.`,
			ja: `互換レポートで ${counts.unusedModelMappings} 件の未使用モデルマッピングが見つかりました。`,
		}));
	}
	if (counts.aliasPreviewMappings > 0) {
		items.push(t(locale, {
			'zh-Hans': `兼容性报告发现 ${counts.aliasPreviewMappings} 条别名预览映射。`,
			'zh-Hant': `相容性報告發現 ${counts.aliasPreviewMappings} 條別名預覽映射。`,
			en: `Compatibility report found ${counts.aliasPreviewMappings} alias preview mappings.`,
			ja: `互換レポートで ${counts.aliasPreviewMappings} 件のエイリアスプレビューが見つかりました。`,
		}));
	}
	if (counts.modelPolicyDiffs > 0) {
		items.push(t(locale, {
			'zh-Hans': `兼容性报告发现 ${counts.modelPolicyDiffs} 条模型策略差异。`,
			'zh-Hant': `相容性報告發現 ${counts.modelPolicyDiffs} 條模型策略差異。`,
			en: `Compatibility report found ${counts.modelPolicyDiffs} model-policy diffs.`,
			ja: `互換レポートで ${counts.modelPolicyDiffs} 件のモデルポリシー差分が見つかりました。`,
		}));
	}

	return items;
}

export function buildImportCompatibilityGuidanceItems(input: CompatibilityGuidanceItemsInput): CompatibilityGuidanceItem[] {
	const { compatibility, counts, effectiveMode, kind = 'import' } = input;
	const locale = input.locale ?? 'zh-Hans';
	const items: CompatibilityGuidanceItem[] = [];
	const totalCredentialRebindTargets = getEffectiveCredentialRebindCount(counts);
	const missingProviderNames = compatibility?.missing_providers ?? [];
	const missingModelNames = getMissingModelMappingItems(compatibility?.model_mapping_previews, locale);
	const unusedModelNames = getUnusedModelMappingItems(compatibility?.model_mapping_previews, locale);
	const credentialTargets = getCredentialRebindTargetItems(compatibility?.credential_rebind_targets, locale);
	const skippedTargetItems = getCompatibilityDiagnosticItems(compatibility?.skipped_targets, locale);
	const replacePrunedItems = getReplacePrunedBreakdownItems({
		channels: compatibility?.replace_pruned_channels,
		groups: compatibility?.replace_pruned_groups,
		settings: compatibility?.replace_pruned_settings,
		models: compatibility?.replace_pruned_llm_infos,
		apiKeys: compatibility?.replace_pruned_api_keys,
	}, locale);
	const replacePrunedExamples = collectExampleText([
		...replacePrunedItems.channels,
		...replacePrunedItems.groups,
		...replacePrunedItems.settings,
		...replacePrunedItems.llmInfos,
		...replacePrunedItems.apiKeys,
	], locale);

	if (
		counts.conflicts > 0
		|| counts.routeConflicts > 0
		|| counts.invalidRouteTargets > 0
		|| counts.baseURLMismatches > 0
		|| counts.schemaMismatches > 0
	) {
		const segments: string[] = [];
		if (counts.conflicts > 0) segments.push(t(locale, {
			'zh-Hans': `${counts.conflicts} 个兼容冲突`,
			'zh-Hant': `${counts.conflicts} 個相容衝突`,
			en: `${counts.conflicts} compatibility conflicts`,
			ja: `${counts.conflicts} 件の互換競合`,
		}));
		if (counts.routeConflicts > 0) segments.push(t(locale, {
			'zh-Hans': `${counts.routeConflicts} 个路由冲突`,
			'zh-Hant': `${counts.routeConflicts} 個路由衝突`,
			en: `${counts.routeConflicts} route conflicts`,
			ja: `${counts.routeConflicts} 件のルート競合`,
		}));
		if (counts.invalidRouteTargets > 0) segments.push(t(locale, {
			'zh-Hans': `${counts.invalidRouteTargets} 个无效路由目标`,
			'zh-Hant': `${counts.invalidRouteTargets} 個無效路由目標`,
			en: `${counts.invalidRouteTargets} invalid route targets`,
			ja: `${counts.invalidRouteTargets} 件の無効なルート対象`,
		}));
		if (counts.baseURLMismatches > 0) segments.push(t(locale, {
			'zh-Hans': `${counts.baseURLMismatches} 个基础地址不匹配`,
			'zh-Hant': `${counts.baseURLMismatches} 個基礎位址不匹配`,
			en: `${counts.baseURLMismatches} base-URL mismatches`,
			ja: `${counts.baseURLMismatches} 件のベース URL 不一致`,
		}));
		if (counts.schemaMismatches > 0) segments.push(t(locale, {
			'zh-Hans': `${counts.schemaMismatches} 个结构不匹配`,
			'zh-Hant': `${counts.schemaMismatches} 個結構不匹配`,
			en: `${counts.schemaMismatches} schema mismatches`,
			ja: `${counts.schemaMismatches} 件のスキーマ不一致`,
		}));
		items.push({
			key: 'blocking-risks',
			tone: 'danger',
			title: t(locale, {
				'zh-Hans': kind === 'rollback' ? '先处理回滚风险' : '先处理阻断风险',
				'zh-Hant': kind === 'rollback' ? '先處理回滾風險' : '先處理阻斷風險',
				en: kind === 'rollback' ? 'Resolve rollback risks first' : 'Resolve blocking risks first',
				ja: kind === 'rollback' ? 'まずロールバックのリスクを解消する' : 'まず阻害リスクを解消する',
			}),
			detail: t(locale, {
				'zh-Hans': kind === 'rollback'
					? `当前回滚预览里有 ${segments.join('、')}。建议先展开下方冲突和路由明细，修正后再执行回滚。`
					: `当前预检里有 ${segments.join('、')}。建议先展开下方冲突和路由明细，修正后再应用。`,
				'zh-Hant': kind === 'rollback'
					? `目前回滾預覽裡有 ${segments.join('、')}。建議先展開下方衝突與路由明細，修正後再執行回滾。`
					: `目前預檢裡有 ${segments.join('、')}。建議先展開下方衝突與路由明細，修正後再套用。`,
				en: kind === 'rollback'
					? `The current rollback preview still contains ${segments.join(', ')}. Expand the conflict and route details below, fix them, then run the rollback again.`
					: `The current preview still contains ${segments.join(', ')}. Expand the conflict and route details below, fix them, then apply again.`,
				ja: kind === 'rollback'
					? `現在のロールバックプレビューには ${segments.join('、')} が残っています。下の競合とルート詳細を確認し、修正してからロールバックを実行してください。`
					: `現在のプレビューには ${segments.join('、')} が残っています。下の競合とルート詳細を確認し、修正してから適用してください。`,
			}),
		});
	}

	if (counts.missingProviders > 0 || counts.missingModels > 0) {
		const examples = collectExampleText([
			...missingProviderNames.map((item) => String(item)),
			...missingModelNames,
		], locale);
		items.push({
			key: 'missing-targets',
			tone: 'warning',
			title: t(locale, {
				'zh-Hans': kind === 'rollback' ? '先补齐缺失对象再回滚' : '补齐缺失的渠道或模型',
				'zh-Hant': kind === 'rollback' ? '先補齊缺失對象再回滾' : '補齊缺失的渠道或模型',
				en: kind === 'rollback' ? 'Restore missing targets before rollback' : 'Restore missing providers or models',
				ja: kind === 'rollback' ? 'ロールバック前に不足対象を補う' : '不足しているチャネルやモデルを補う',
			}),
			detail: examples
				? t(locale, {
					'zh-Hans': kind === 'rollback'
						? `回滚预览已标出缺失对象，优先处理这些名称：${examples}。`
						: `预检已标出缺失对象，优先处理这些名称：${examples}。`,
					'zh-Hant': kind === 'rollback'
						? `回滾預覽已標出缺失對象，優先處理這些名稱：${examples}。`
						: `預檢已標出缺失對象，優先處理這些名稱：${examples}。`,
					en: kind === 'rollback'
						? `The rollback preview already marked the missing objects. Start with these examples: ${examples}.`
						: `The preview already marked the missing objects. Start with these examples: ${examples}.`,
					ja: kind === 'rollback'
						? `ロールバックプレビューで不足対象が示されています。まずは次の例から対応してください: ${examples}。`
						: `プレビューで不足対象が示されています。まずは次の例から対応してください: ${examples}。`,
				})
				: t(locale, {
					'zh-Hans': kind === 'rollback' ? '先补齐缺失的渠道、提供商或模型，再重新执行一次回滚预览。' : '先补齐缺失的渠道、提供商或模型，再重新跑一次预检。',
					'zh-Hant': kind === 'rollback' ? '先補齊缺失的渠道、供應商或模型，再重新執行一次回滾預覽。' : '先補齊缺失的渠道、供應商或模型，再重新跑一次預檢。',
					en: kind === 'rollback' ? 'Add the missing providers, channels, or models first, then run the rollback preview again.' : 'Add the missing providers, channels, or models first, then run the preview again.',
					ja: kind === 'rollback' ? '不足しているプロバイダー、チャネル、モデルを補ってから、ロールバックプレビューを再実行してください。' : '不足しているプロバイダー、チャネル、モデルを補ってから、プレビューを再実行してください。',
				}),
		});
	}

	if (totalCredentialRebindTargets > 0) {
		const examples = collectExampleText(credentialTargets, locale);
		items.push({
			key: 'credential-rebind',
			tone: 'warning',
			title: t(locale, {
				'zh-Hans': kind === 'rollback' ? '提前准备回滚后的凭证重绑定' : '提前准备凭证重绑定',
				'zh-Hant': kind === 'rollback' ? '提前準備回滾後的憑證重綁定' : '提前準備憑證重綁定',
				en: kind === 'rollback' ? 'Prepare post-rollback credential rebinds' : 'Prepare credential rebinds',
				ja: kind === 'rollback' ? 'ロールバック後の認証情報再バインドを準備する' : '認証情報の再バインドを準備する',
			}),
			detail: examples
				? t(locale, {
					'zh-Hans': kind === 'rollback' ? `这些目标在回滚后需要重新绑定凭证：${examples}。` : `这些目标在应用后需要重新绑定凭证：${examples}。`,
					'zh-Hant': kind === 'rollback' ? `這些目標在回滾後需要重新綁定憑證：${examples}。` : `這些目標在套用後需要重新綁定憑證：${examples}。`,
					en: kind === 'rollback' ? `These targets will need credential rebinds after rollback: ${examples}.` : `These targets will need credential rebinds after apply: ${examples}.`,
					ja: kind === 'rollback' ? `ロールバック後に認証情報を再バインドする必要がある対象: ${examples}。` : `適用後に認証情報を再バインドする必要がある対象: ${examples}。`,
				})
				: t(locale, {
					'zh-Hans': kind === 'rollback' ? '执行回滚前先确认哪些渠道密钥或 API 密钥需要重新绑定，避免恢复后目标失效。' : '应用前先确认哪些渠道密钥或 API 密钥需要重新绑定，避免导入后目标失效。',
					'zh-Hant': kind === 'rollback' ? '執行回滾前先確認哪些渠道金鑰或 API 金鑰需要重新綁定，避免恢復後目標失效。' : '套用前先確認哪些渠道金鑰或 API 金鑰需要重新綁定，避免導入後目標失效。',
					en: kind === 'rollback' ? 'Review which channel-key or API-key targets need rebinds before rollback, so the restore does not leave them unusable.' : 'Review which channel-key or API-key targets need rebinds before applying, so the import does not leave them unusable.',
					ja: kind === 'rollback' ? 'ロールバック前に、どのチャネルキーや API キーが再バインドを要するか確認し、復元後に対象が使えなくならないようにしてください。' : '適用前に、どのチャネルキーや API キーが再バインドを要するか確認し、導入後に対象が使えなくならないようにしてください。',
				}),
		});
	}

	if (counts.skippedTargets > 0) {
		const examples = collectExampleText(skippedTargetItems, locale);
		items.push({
			key: 'skipped-targets',
			tone: 'warning',
			title: t(locale, {
				'zh-Hans': kind === 'rollback' ? '确认回滚会保留哪些对象' : '确认哪些对象会被跳过',
				'zh-Hant': kind === 'rollback' ? '確認回滾會保留哪些對象' : '確認哪些目標會被跳過',
				en: kind === 'rollback' ? 'Review which targets rollback keeps' : 'Review which targets are skipped',
				ja: kind === 'rollback' ? 'ロールバックで保持される対象を確認する' : 'どの対象がスキップされるか確認する',
			}),
			detail: examples
				? t(locale, {
					'zh-Hans': kind === 'rollback' ? `当前回滚会保留或跳过这些对象：${examples}。建议先确认这是否符合本次恢复预期。` : `当前模式会保留或跳过这些对象：${examples}。建议先确认这是否符合本次导入预期。`,
					'zh-Hant': kind === 'rollback' ? `目前回滾會保留或跳過這些對象：${examples}。建議先確認這是否符合本次恢復預期。` : `目前模式會保留或跳過這些目標：${examples}。建議先確認這是否符合本次導入預期。`,
					en: kind === 'rollback' ? `This rollback will preserve or skip these targets: ${examples}. Confirm that this matches what you want before restoring.` : `This mode will preserve or skip these targets: ${examples}. Confirm that this matches what you want before applying.`,
					ja: kind === 'rollback' ? `このロールバックでは次の対象が保持またはスキップされます: ${examples}。復元前に意図どおりか確認してください。` : `このモードでは次の対象が保持またはスキップされます: ${examples}。適用前に意図どおりか確認してください。`,
				})
				: t(locale, {
					'zh-Hans': kind === 'rollback' ? '回滚预览已标出会被保留或跳过的对象，建议先展开下方明细确认再执行回滚。' : '预检已标出会被保留或跳过的对象，建议先展开下方明细确认再应用。',
					'zh-Hant': kind === 'rollback' ? '回滾預覽已標出會被保留或跳過的對象，建議先展開下方明細確認再執行回滾。' : '預檢已標出會被保留或跳過的目標，建議先展開下方明細確認再套用。',
					en: kind === 'rollback' ? 'The rollback preview already marked which targets will be preserved or skipped. Review the details below before running rollback.' : 'The preview already marked which targets will be preserved or skipped. Review the details below before applying.',
					ja: kind === 'rollback' ? 'ロールバックプレビューには保持またはスキップされる対象が示されています。ロールバック前に下の詳細を確認してください。' : 'プレビューには保持またはスキップされる対象が示されています。適用前に下の詳細を確認してください。',
				}),
		});
	}

	if ((effectiveMode === 'replace' || kind === 'rollback') && counts.replacePrunedTargets > 0) {
		items.push({
			key: 'replace-prune',
			tone: 'warning',
			title: t(locale, {
				'zh-Hans': kind === 'rollback' ? '确认回滚会清理哪些当前记录' : '确认替换会清理哪些当前记录',
				'zh-Hant': kind === 'rollback' ? '確認回滾會清理哪些目前記錄' : '確認替換會清理哪些目前記錄',
				en: kind === 'rollback' ? 'Review which current records rollback removes' : 'Review which current records replace mode removes',
				ja: kind === 'rollback' ? 'ロールバックで整理される現在の記録を確認する' : '置換で削除される現在の記録を確認する',
			}),
			detail: replacePrunedExamples
				? t(locale, {
					'zh-Hans': kind === 'rollback'
						? `回滚恢复会清理这些当前记录：${replacePrunedExamples}。建议先展开下方结构化清理明细，再决定是否执行回滚。`
						: `替换导入会清理这些当前记录：${replacePrunedExamples}。建议先展开下方结构化清理明细，再决定是否应用。`,
					'zh-Hant': kind === 'rollback'
						? `回滾恢復會清理這些目前記錄：${replacePrunedExamples}。建議先展開下方結構化清理明細，再決定是否執行回滾。`
						: `替換導入會清理這些目前記錄：${replacePrunedExamples}。建議先展開下方結構化清理明細，再決定是否套用。`,
					en: kind === 'rollback'
						? `Rollback restore will prune these current records: ${replacePrunedExamples}. Expand the structured cleanup details below before restoring.`
						: `Replace mode will prune these current records: ${replacePrunedExamples}. Expand the structured prune details below before you apply it.`,
					ja: kind === 'rollback'
						? `ロールバック復元では次の現在記録が整理対象になります: ${replacePrunedExamples}。復元前に下の構造化された整理詳細を確認してください。`
						: `置換モードでは次の現在記録が整理対象になります: ${replacePrunedExamples}。適用前に下の構造化プレビューを確認してください。`,
				})
				: t(locale, {
					'zh-Hans': kind === 'rollback'
						? '回滚恢复会清理当前项目中未被目标快照保留的记录，建议先展开下方结构化清理明细确认范围。'
						: '替换导入会清理当前项目中未被快照保留的记录，建议先展开下方结构化清理明细确认范围。',
					'zh-Hant': kind === 'rollback'
						? '回滾恢復會清理目前專案中未被目標快照保留的記錄，建議先展開下方結構化清理明細確認範圍。'
						: '替換導入會清理目前專案中未被快照保留的記錄，建議先展開下方結構化清理明細確認範圍。',
					en: kind === 'rollback'
						? 'Rollback restore prunes current records that the target snapshot does not keep. Review the structured cleanup details below before restoring.'
						: 'Replace mode prunes current records that the snapshot does not keep. Review the structured prune details below before applying.',
					ja: kind === 'rollback'
						? 'ロールバック復元では対象スナップショットに含まれない現在の記録が整理されます。復元前に下の構造化された整理詳細を確認してください。'
						: '置換モードではスナップショットに含まれない現在の記録が整理されます。適用前に下の構造化プレビューを確認してください。',
				}),
		});
	}

	if (counts.missingMappingTargets > 0 || counts.unusedModelMappings > 0) {
		const mappingExamples = collectExampleText([
			...missingModelNames,
			...unusedModelNames,
		], locale);
		items.push({
			key: 'model-mappings',
			tone: 'warning',
			title: t(locale, {
				'zh-Hans': effectiveMode === 'map'
					? (kind === 'rollback' ? '修正模型映射后再回滚' : '修正模型映射后再应用')
					: (kind === 'rollback' ? '复核回滚用模型映射规则' : '复核模型映射规则'),
				'zh-Hant': effectiveMode === 'map'
					? (kind === 'rollback' ? '修正模型映射後再回滾' : '修正模型映射後再套用')
					: (kind === 'rollback' ? '複核回滾用模型映射規則' : '複核模型映射規則'),
				en: effectiveMode === 'map'
					? (kind === 'rollback' ? 'Fix model mappings before rollback' : 'Fix model mappings before apply')
					: (kind === 'rollback' ? 'Review rollback model mapping rules' : 'Review model mapping rules'),
				ja: effectiveMode === 'map'
					? (kind === 'rollback' ? 'ロールバック前にモデルマッピングを修正する' : '適用前にモデルマッピングを修正する')
					: (kind === 'rollback' ? 'ロールバック用のモデルマッピング規則を確認する' : 'モデルマッピング規則を確認する'),
			}),
			detail: mappingExamples
				? t(locale, {
					'zh-Hans': `当前映射里已经暴露出需要修正或清理的项目：${mappingExamples}。`,
					'zh-Hant': `目前映射裡已經暴露出需要修正或清理的項目：${mappingExamples}。`,
					en: `The current mapping set already shows items to fix or remove: ${mappingExamples}.`,
					ja: `現在のマッピングには修正または整理が必要な項目があります: ${mappingExamples}。`,
				})
				: t(locale, {
					'zh-Hans': '先把缺失目标和未使用映射清干净，再继续应用当前快照。',
					'zh-Hant': '先把缺失目標和未使用映射清乾淨，再繼續套用目前快照。',
					en: 'Clean up missing targets and unused mappings first, then continue with this snapshot.',
					ja: '不足マッピング先と未使用マッピングを整理してから、このスナップショットを続行してください。',
				}),
		});
	}

	if (
		counts.routePreviewDiffs > 0
		|| counts.routePreviewWarnings > 0
		|| counts.skippedRouteTargetPreviews > 0
		|| counts.modelPolicyDiffs > 0
		|| counts.aliasPreviewMappings > 0
	) {
		const signals: string[] = [];
		if (counts.routePreviewDiffs > 0) signals.push(t(locale, {
			'zh-Hans': `${counts.routePreviewDiffs} 处路由差异`,
			'zh-Hant': `${counts.routePreviewDiffs} 處路由差異`,
			en: `${counts.routePreviewDiffs} route diffs`,
			ja: `${counts.routePreviewDiffs} 件のルート差分`,
		}));
		if (counts.routePreviewWarnings > 0) signals.push(t(locale, {
			'zh-Hans': `${counts.routePreviewWarnings} 条路由预警`,
			'zh-Hant': `${counts.routePreviewWarnings} 條路由預警`,
			en: `${counts.routePreviewWarnings} route warnings`,
			ja: `${counts.routePreviewWarnings} 件のルート警告`,
		}));
		if (counts.skippedRouteTargetPreviews > 0) signals.push(t(locale, {
			'zh-Hans': `${counts.skippedRouteTargetPreviews} 个跳过的路由预览`,
			'zh-Hant': `${counts.skippedRouteTargetPreviews} 個跳過的路由預覽`,
			en: `${counts.skippedRouteTargetPreviews} skipped route previews`,
			ja: `${counts.skippedRouteTargetPreviews} 件のスキップされたルートプレビュー`,
		}));
		if (counts.modelPolicyDiffs > 0) signals.push(t(locale, {
			'zh-Hans': `${counts.modelPolicyDiffs} 条模型策略差异`,
			'zh-Hant': `${counts.modelPolicyDiffs} 條模型策略差異`,
			en: `${counts.modelPolicyDiffs} model-policy diffs`,
			ja: `${counts.modelPolicyDiffs} 件のモデルポリシー差分`,
		}));
		if (counts.aliasPreviewMappings > 0) signals.push(t(locale, {
			'zh-Hans': `${counts.aliasPreviewMappings} 条别名映射预览`,
			'zh-Hant': `${counts.aliasPreviewMappings} 條別名映射預覽`,
			en: `${counts.aliasPreviewMappings} alias preview mappings`,
			ja: `${counts.aliasPreviewMappings} 件のエイリアスプレビュー`,
		}));
		items.push({
			key: 'route-and-policy',
			tone: 'warning',
			title: t(locale, {
				'zh-Hans': kind === 'rollback' ? '复核回滚后的路由与策略差异' : '复核路由与策略差异',
				'zh-Hant': kind === 'rollback' ? '複核回滾後的路由與策略差異' : '複核路由與策略差異',
				en: kind === 'rollback' ? 'Review post-rollback route and policy drift' : 'Review route and policy drift',
				ja: kind === 'rollback' ? 'ロールバック後のルートとポリシー差分を確認する' : 'ルートとポリシーの差分を確認する',
			}),
			detail: t(locale, {
				'zh-Hans': kind === 'rollback' ? `当前回滚预览还提示了 ${signals.join('、')}。建议在执行回滚前把下方路由与策略明细过一遍。` : `当前预检还提示了 ${signals.join('、')}。建议在应用前把下方路由与策略明细过一遍。`,
				'zh-Hant': kind === 'rollback' ? `目前回滾預覽還提示了 ${signals.join('、')}。建議在執行回滾前把下方路由與策略明細過一遍。` : `目前預檢還提示了 ${signals.join('、')}。建議在套用前把下方路由與策略明細過一遍。`,
				en: kind === 'rollback' ? `The rollback preview also surfaced ${signals.join(', ')}. Review the route and policy details below before restoring.` : `The preview also surfaced ${signals.join(', ')}. Review the route and policy details below before applying.`,
				ja: kind === 'rollback' ? `現在のロールバックプレビューでは ${signals.join('、')} も示されています。復元前に下のルートとポリシー詳細を確認してください。` : `現在のプレビューでは ${signals.join('、')} も示されています。適用前に下のルートとポリシー詳細を確認してください。`,
			}),
		});
	}

	return items;
}

export function getExportSnapshotPresentation(input: {
	includeSecrets: boolean;
	includeLogs: boolean;
	includeStats: boolean;
	locale?: BackupLogicLocale;
}): ExportSnapshotPresentation {
	const locale = (input as typeof input & { locale?: BackupLogicLocale }).locale ?? 'zh-Hans';
	const scopeBadges = [
		t(locale, { 'zh-Hans': '项目快照', 'zh-Hant': '專案快照', en: 'Project snapshot', ja: 'スナップショット' }),
		t(locale, { 'zh-Hans': '渠道 / 分组 / 路由', 'zh-Hant': '渠道 / 分組 / 路由', en: 'Channels / groups / routing', ja: 'チャネル / グループ / ルーティング' }),
		input.includeSecrets
			? t(locale, { 'zh-Hans': '明文凭证', 'zh-Hant': '明文憑證', en: 'Plaintext credentials', ja: '平文認証情報' })
			: t(locale, { 'zh-Hans': '脱敏凭证', 'zh-Hant': '脫敏憑證', en: 'Redacted credentials', ja: 'マスク済み認証情報' }),
	];
	if (input.includeStats) scopeBadges.push(t(locale, { 'zh-Hans': '统计数据', 'zh-Hant': '統計資料', en: 'Stats', ja: '統計' }));
	if (input.includeLogs) scopeBadges.push(t(locale, { 'zh-Hans': '中继日志', 'zh-Hant': '中繼日誌', en: 'Relay logs', ja: '中継ログ' }));

	if (input.includeSecrets) {
		return {
			summary: t(locale, {
				'zh-Hans': '导出一份完整项目迁移快照，默认包含渠道、分组、路由条目、模型绑定、明文凭证，以及可选的统计与日志。只有在你明确需要脱敏审阅版时，才建议关闭凭证导出。',
				'zh-Hant': '匯出一份完整專案遷移快照，預設包含渠道、分組、路由條目、模型綁定、明文憑證，以及可選的統計與日誌。只有在你明確需要脫敏審閱版時，才建議關閉憑證匯出。',
				en: 'Export a full-project migration snapshot with channels, groups, route items, model bindings, plaintext credentials by default, and optional stats/logs. Turn secrets off only when you intentionally need a redacted review snapshot.',
				ja: 'チャネル、グループ、ルート項目、モデル紐付け、平文の認証情報、必要に応じた統計とログを含む完全な移行スナップショットをエクスポートします。共有や確認用のマスク済み版が必要な場合のみ認証情報をオフにしてください。',
			}),
			warning: t(locale, {
				'zh-Hans': '默认导出会包含明文凭证，这样快照才能直接恢复到另一套环境。只有在你明确要分享或审阅脱敏版时，才建议关闭。',
				'zh-Hant': '預設匯出會包含明文憑證，這樣快照才能直接恢復到另一套環境。只有在你明確要分享或審閱脫敏版時，才建議關閉。',
				en: 'Default exports include plaintext credentials so the snapshot can restore directly into another environment. Turn this off only when you intentionally need a redacted snapshot for sharing or review.',
				ja: '既定のエクスポートには平文の認証情報が含まれるため、別環境へそのまま復元できます。共有やレビュー用のマスク済みスナップショットが必要な場合のみオフにしてください。',
			}),
			scopeBadges,
			toggleLabel: t(locale, {
				'zh-Hans': '在快照中包含明文凭证',
				'zh-Hant': '在快照中包含明文憑證',
				en: 'Include plaintext credentials in the snapshot',
				ja: 'スナップショットに平文の認証情報を含める',
			}),
		};
	}

	return {
		summary: t(locale, {
			'zh-Hans': '导出一份脱敏后的项目快照，包含渠道、分组、路由条目、模型绑定，以及可选的统计与日志。如果你需要一份可以直接恢复的迁移快照，请重新打开明文凭证导出。',
			'zh-Hant': '匯出一份脫敏後的專案快照，包含渠道、分組、路由條目、模型綁定，以及可選的統計與日誌。如果你需要一份可直接恢復的遷移快照，請重新開啟明文憑證匯出。',
			en: 'Export a redacted project snapshot with channels, groups, route items, model bindings, and optional stats/logs. Turn secrets back on when you need a restore-ready migration snapshot.',
			ja: 'チャネル、グループ、ルート項目、モデル紐付け、必要に応じた統計とログを含むマスク済みのプロジェクトスナップショットをエクスポートします。直接復元できる移行スナップショットが必要な場合は認証情報を再度オンにしてください。',
		}),
		warning: t(locale, {
			'zh-Hans': '这份导出不会包含明文凭证，适合分享或审阅，不适合直接恢复。',
			'zh-Hant': '這份匯出不會包含明文憑證，適合分享或審閱，不適合直接恢復。',
			en: 'This export omits plaintext credentials. Use it for sharing or review when you do not need a directly restorable snapshot.',
			ja: 'このエクスポートには平文の認証情報が含まれません。共有やレビューには向いていますが、そのままの復元には向きません。',
		}),
		scopeBadges,
		toggleLabel: t(locale, {
			'zh-Hans': '在快照中包含明文凭证',
			'zh-Hant': '在快照中包含明文憑證',
			en: 'Include plaintext credentials in the snapshot',
			ja: 'スナップショットに平文の認証情報を含める',
		}),
	};
}

export function getPostImportValidationSummary(postImportValidation: PostImportValidationLike | undefined): PostImportValidationSummary | null {
	if (!postImportValidation) return null;
	const hasAnyData = !!postImportValidation.summary
		|| getSummaryOrListCount(undefined, postImportValidation.degraded_groups) > 0
		|| getSummaryOrListCount(undefined, postImportValidation.empty_groups) > 0
		|| getSummaryOrListCount(undefined, postImportValidation.disabled_channels) > 0
		|| getSummaryOrListCount(undefined, postImportValidation.channels_without_keys) > 0
		|| getSummaryOrListCount(undefined, postImportValidation.stale_items_removed) > 0
		|| getSummaryOrListCount(undefined, postImportValidation.route_warnings) > 0
		|| getSummaryOrListCount(undefined, postImportValidation.price_rule_warnings) > 0
		|| getSummaryOrListCount(undefined, postImportValidation.alias_mappings) > 0
		|| getSummaryOrListCount(undefined, postImportValidation.alias_warnings) > 0;

	if (!hasAnyData) return null;

	return {
		degradedGroups: getSummaryOrListCount(postImportValidation.summary?.degraded_groups, postImportValidation.degraded_groups),
		emptyGroups: getSummaryOrListCount(postImportValidation.summary?.empty_groups, postImportValidation.empty_groups),
		disabledChannels: getSummaryOrListCount(postImportValidation.summary?.disabled_channels, postImportValidation.disabled_channels),
		channelsWithoutKeys: getSummaryOrListCount(postImportValidation.summary?.channels_without_keys, postImportValidation.channels_without_keys),
		staleItemsRemoved: getSummaryOrListCount(postImportValidation.summary?.stale_items_removed, postImportValidation.stale_items_removed),
		routeWarnings: getSummaryOrListCount(postImportValidation.summary?.route_warnings, postImportValidation.route_warnings),
		priceRuleWarnings: getSummaryOrListCount(postImportValidation.summary?.price_rule_warnings, postImportValidation.price_rule_warnings),
		aliasMappings: getSummaryOrListCount(postImportValidation.summary?.alias_mappings, postImportValidation.alias_mappings),
		aliasWarnings: getSummaryOrListCount(postImportValidation.summary?.alias_warnings, postImportValidation.alias_warnings),
	};
}

export function getApplySameImportGuardReason(input: {
	hasPendingApplyRequest: boolean;
	previewToken?: string;
	requiresConfirm: boolean;
	confirmed: boolean;
}): ApplySameImportGuardReason {
	if (!input.hasPendingApplyRequest) return 'missing_request';
	if (!input.previewToken?.trim()) return 'missing_preview_token';
	if (input.requiresConfirm && !input.confirmed) return 'confirm_required';
	return null;
}

export function getImportResultPresentation(input: {
	resultIsDryRun: boolean;
	modeLabel: string;
	locale?: BackupLogicLocale;
}) {
	const locale = (input as typeof input & { locale?: BackupLogicLocale }).locale ?? 'zh-Hans';
	if (input.resultIsDryRun) {
		return {
			title: t(locale, { 'zh-Hans': '预检结果', 'zh-Hant': '預檢結果', en: 'Dry-run report', ja: 'ドライラン結果' }),
			description: t(locale, { 'zh-Hans': '预检只会分析兼容性和路由影响提示，不会写入数据库。', 'zh-Hant': '預檢只會分析相容性與路由影響提示，不會寫入資料庫。', en: 'Dry-run report only analyzes compatibility and route-impact hints. It does not write to the database.', ja: 'ドライランでは互換性とルート影響の確認のみを行い、データベースには書き込みません。' }),
			modeLabel: t(locale, { 'zh-Hans': '本次使用的预检模式', 'zh-Hant': '本次使用的預檢模式', en: 'Dry-run mode used', ja: '今回のドライランモード' }),
		};
	}

	return {
		title: t(locale, { 'zh-Hans': '导入已应用', 'zh-Hant': '導入已套用', en: 'Import applied', ja: 'インポート完了' }),
		description: t(locale, { 'zh-Hans': `已按 ${input.modeLabel} 模式应用导入。导入后校验仍可能清理过期路由引用，或提示后续需要继续检查的项目。`, 'zh-Hant': `已按 ${input.modeLabel} 模式套用導入。導入後校驗仍可能清理過期路由引用，或提示後續仍需檢查的項目。`, en: `Import applied in ${input.modeLabel} mode. Post-import validation can still remove stale route references or surface follow-up checks.`, ja: `${input.modeLabel} モードでインポートを適用しました。インポート後の検証では、古いルート参照の整理や追加確認項目の通知が行われる場合があります。` }),
		modeLabel: t(locale, { 'zh-Hans': '本次应用的导入模式', 'zh-Hant': '本次套用的導入模式', en: 'Applied mode used', ja: '今回の適用モード' }),
	};
}

