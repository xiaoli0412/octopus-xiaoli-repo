'use client';

import { type ChangeEvent, type ReactNode, useEffect, useMemo, useState } from 'react';
import {
	CheckCircle2,
	Database,
	Download,
	History,
	ShieldAlert,
	Upload,
} from 'lucide-react';

import { toast } from '@/components/common/Toast';
import { HelpHint } from '@/components/common/HelpHint';
import { CopyIconButton } from '@/components/common/CopyButton';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { cn } from '@/lib/utils';
import {
	type DBImportMode,
	type DBImportResult,
	type DBImportScopes,
	type DBImportSnapshotInfo,
	type DBReplacePrunePreview,
	type DBRollbackPreviewResult,
	useExportDB,
	useImportDB,
	useImportSnapshots,
	usePreviewRollbackImportSnapshot,
	useRollbackImportSnapshot,
	useRollbackLatestImportSnapshot,
} from '@/api/endpoints/setting';
import {
	buildCompatibilitySignalItems,
	getCompatibilityCounts,
	getCompatibilityOverview,
	getExportSnapshotPresentation,
	getMissingModelMappingItems,
	getModelMappingPreviewItems,
	getPostImportValidationSummary,
	getRemainingMigrationToolingSections,
	getUnusedModelMappingItems,
	type SummaryTone,
} from './backup-logic';
import { formatDateTimeByLocale } from '@/lib/locale';
import { useSettingStore, type Locale } from '@/stores/setting';

type Copy = {
	'zh-Hans': string;
	en: string;
	'zh-Hant'?: string;
	ja?: string;
};

type ReplacePruneSection = {
	key: string;
	title: string;
	items: string[];
};

type PendingApplyRequest = {
	file: File;
	mode: DBImportMode;
	modelMappings?: Record<string, string>;
	importScopes: DBImportScopes;
	previewToken?: string;
	fileName: string;
	mappingCount: number;
	scopeLabels: string[];
};

const defaultImportScopes: DBImportScopes = {
	routing: true,
	models: true,
	api_keys: true,
	settings: true,
	stats: true,
	logs: true,
};

function localize(locale: Locale, copy: Copy) {
	if (locale === 'zh-Hans') return copy['zh-Hans'];
	if (locale === 'zh-Hant') return copy['zh-Hant'] ?? copy['zh-Hans'];
	if (locale === 'ja') return copy.ja ?? copy.en;
	return copy.en;
}

function getSnapshotHistoryItemTestId(snapshot: DBImportSnapshotInfo) {
	const rawKey = snapshot.snapshot_name ?? snapshot.snapshot_path ?? 'snapshot';
	const normalizedKey = rawKey.trim().toLowerCase().replace(/[^a-z0-9]+/g, '-');
	return `backup-history-item-${normalizedKey || 'snapshot'}`;
}

function joinLocalizedList(items: string[], locale: Locale) {
	if (items.length === 0) return '';
	if (items.length <= 3) return items.join(locale === 'en' ? ', ' : '、');
	return `${items.slice(0, 3).join(locale === 'en' ? ', ' : '、')} +${items.length - 3}`;
}

function toneClasses(tone: SummaryTone) {
	if (tone === 'danger') return 'border-red-500/30 bg-red-500/10 text-red-700';
	if (tone === 'warning') return 'border-amber-500/30 bg-amber-500/10 text-amber-700';
	return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700';
}

function formatFileSize(size: number) {
	if (size < 1024) return `${size} B`;
	return `${Math.ceil(size / 1024)} KB`;
}

function formatReplacePruneValue(value: unknown) {
	if (typeof value === 'string' || typeof value === 'number') return String(value);
	if (value && typeof value === 'object') {
		for (const key of ['name', 'channel_name', 'group_name', 'model', 'key_name']) {
			const candidate = (value as Record<string, unknown>)[key];
			if (typeof candidate === 'string' || typeof candidate === 'number') return String(candidate);
		}
		return JSON.stringify(value);
	}
	return String(value ?? '');
}

function getRowsAffectedTotal(result: DBImportResult | undefined) {
	if (!result?.rows_affected) return 0;
	return Object.values(result.rows_affected).reduce((sum, value) => sum + (typeof value === 'number' ? value : 0), 0);
}

function getImportScopeOptions(locale: Locale) {
	return [
		{ key: 'routing' as const, label: localize(locale, { 'zh-Hans': '路由配置', en: 'Routing' }) },
		{ key: 'models' as const, label: localize(locale, { 'zh-Hans': '模型数据', en: 'Models' }) },
		{ key: 'api_keys' as const, label: localize(locale, { 'zh-Hans': 'API 密钥', en: 'API keys' }) },
		{ key: 'settings' as const, label: localize(locale, { 'zh-Hans': '系统设置', en: 'Settings' }) },
		{ key: 'stats' as const, label: localize(locale, { 'zh-Hans': '统计数据', en: 'Stats' }) },
		{ key: 'logs' as const, label: localize(locale, { 'zh-Hans': '中继日志', en: 'Relay logs' }) },
	];
}

function getImportModeOptions(locale: Locale) {
	return [
		{ value: 'incremental' as const, label: localize(locale, { 'zh-Hans': '增量导入', en: 'Incremental' }) },
		{ value: 'map' as const, label: localize(locale, { 'zh-Hans': '映射导入', en: 'Map' }) },
		{ value: 'merge' as const, label: localize(locale, { 'zh-Hans': '合并导入', en: 'Merge' }) },
		{ value: 'replace' as const, label: localize(locale, { 'zh-Hans': '替换导入', en: 'Replace' }) },
		{ value: 'skip' as const, label: localize(locale, { 'zh-Hans': '跳过已存在项', en: 'Skip existing' }) },
	];
}

function getLocalizedModelMappingsPlaceholder(locale: Locale) {
	if (locale === 'en' || locale === 'ja') return 'legacy-model=gpt-4o\nvision-model=gpt-4.1';
	return '旧模型=gpt-4o\n视觉模型=gpt-4.1';
}

function buildReplacePruneSections(preview: DBReplacePrunePreview | undefined, locale: Locale): ReplacePruneSection[] {
	if (!preview) return [];
	const labels = {
		channels: localize(locale, { 'zh-Hans': '待删除渠道', en: 'Channels to delete' }),
		groups: localize(locale, { 'zh-Hans': '待删除分组', en: 'Groups to delete' }),
		settings: localize(locale, { 'zh-Hans': '待重置设置', en: 'Settings to reset' }),
		models: localize(locale, { 'zh-Hans': '待移除模型条目', en: 'Models to delete' }),
		apiKeys: localize(locale, { 'zh-Hans': '待删除 API 密钥', en: 'API keys to delete' }),
	};
	return [
		{ key: 'channels', title: labels.channels, items: (preview.pruned_channels ?? preview.deleted_channels ?? preview.channels ?? preview.channel_names ?? []).map(formatReplacePruneValue).filter(Boolean) },
		{ key: 'groups', title: labels.groups, items: (preview.pruned_groups ?? preview.deleted_groups ?? preview.groups ?? preview.group_names ?? []).map(formatReplacePruneValue).filter(Boolean) },
		{ key: 'settings', title: labels.settings, items: (preview.pruned_settings ?? preview.deleted_settings ?? preview.settings ?? preview.setting_keys ?? []).map(formatReplacePruneValue).filter(Boolean) },
		{ key: 'models', title: labels.models, items: (preview.pruned_llm_infos ?? preview.deleted_llm_infos ?? preview.llm_infos ?? preview.models ?? preview.model_names ?? []).map(formatReplacePruneValue).filter(Boolean) },
		{ key: 'apiKeys', title: labels.apiKeys, items: (preview.pruned_api_keys ?? preview.deleted_api_keys ?? preview.api_keys ?? preview.api_key_names ?? []).map(formatReplacePruneValue).filter(Boolean) },
	].filter((section) => section.items.length > 0);
}

function parseModelMappings(text: string, locale: Locale) {
	const invalidFormat = localize(locale, {
		'zh-Hans': '格式不正确，请使用“快照模型=当前模型”的写法。',
		en: 'Line is invalid. Use snapshot-model=current-model.',
	});
	const invalidValue = localize(locale, {
		'zh-Hans': '必须同时填写快照模型和当前模型。',
		en: 'Both snapshot and current model names are required.',
	});
	const lines = text.split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
	const mappings: Record<string, string> = {};
	for (const [index, line] of lines.entries()) {
		const eqIndex = line.indexOf('=');
		if (eqIndex <= 0 || eqIndex === line.length - 1) throw new Error(`${index + 1} ${invalidFormat}`);
		const source = line.slice(0, eqIndex).trim();
		const target = line.slice(eqIndex + 1).trim();
		if (!source || !target) throw new Error(`${index + 1} ${invalidValue}`);
		mappings[source] = target;
	}
	return mappings;
}

function InlineHelpLabel({ children, hint }: { children: ReactNode; hint?: ReactNode }) {
	return (
		<span className="inline-flex items-center gap-1.5">
			<span>{children}</span>
			{hint ? <HelpHint>{hint}</HelpHint> : null}
		</span>
	);
}

function SectionCard({ icon, title, hint, description, children }: { icon: ReactNode; title: string; hint?: ReactNode; description?: string; children: ReactNode }) {
	return (
		<section className="space-y-3 rounded-2xl border border-border/60 bg-background/50 p-3.5">
			<div className="space-y-1">
				<div className="flex items-center gap-2 text-sm font-semibold text-card-foreground">
					{icon}
					<InlineHelpLabel hint={hint}>{title}</InlineHelpLabel>
				</div>
				{description ? <div className="text-xs text-muted-foreground line-clamp-2">{description}</div> : null}
			</div>
			{children}
		</section>
	);
}

function SwitchRow({ label, hint, checked, onCheckedChange, testId }: { label: string; hint?: ReactNode; checked: boolean; onCheckedChange: (checked: boolean) => void; testId?: string }) {
	return (
		<div className="flex items-center justify-between gap-3 rounded-xl border border-border/50 bg-background/70 px-2.5 py-2">
			<div className="text-sm text-card-foreground"><InlineHelpLabel hint={hint}>{label}</InlineHelpLabel></div>
			<Switch checked={checked} onCheckedChange={onCheckedChange} data-testid={testId} />
		</div>
	);
}

const backupTextByLocale = (locale: Locale) => ({
	exportTitle: localize(locale, { 'zh-Hans': '导出快照', en: 'Export Snapshot' }),
	importTitle: localize(locale, { 'zh-Hans': '导入与预检', en: 'Import And Preview' }),
	rollbackTitle: localize(locale, { 'zh-Hans': '导入快照历史', en: 'Snapshot History' }),
	exportDescription: localize(locale, { 'zh-Hans': '同一个入口处理完整迁移和脱敏共享。', en: 'Use one entry for full restore and redacted sharing.' }),
	importDescription: localize(locale, { 'zh-Hans': '默认先预检，确认后再写入当前项目。', en: 'Preview first, then apply to the current project.' }),
	rollbackDescription: localize(locale, { 'zh-Hans': '先预览，再决定是否回滚。', en: 'Preview first, then decide whether to roll back.' }),
	exportFormat: localize(locale, { 'zh-Hans': '导出格式', en: 'Export Format' }),
	exportFormatStandard: localize(locale, { 'zh-Hans': '标准完整备份', en: 'Standard Full Backup' }),
	exportFormatLegacy: localize(locale, { 'zh-Hans': '原版兼容备份', en: 'Legacy-Compatible Backup' }),
	includeLogs: localize(locale, { 'zh-Hans': '包含中继日志', en: 'Include Relay Logs' }),
	includeStats: localize(locale, { 'zh-Hans': '包含统计数据', en: 'Include Stats' }),
	exportButton: localize(locale, { 'zh-Hans': '下载 JSON', en: 'Download JSON' }),
	exporting: localize(locale, { 'zh-Hans': '导出中...', en: 'Exporting...' }),
	selectedFile: localize(locale, { 'zh-Hans': '已选择文件', en: 'Selected File' }),
	dryRun: localize(locale, { 'zh-Hans': '先做预检', en: 'Dry-Run First' }),
	selectiveImport: localize(locale, { 'zh-Hans': '按范围导入', en: 'Selective import' }),
	importMode: localize(locale, { 'zh-Hans': '导入模式', en: 'Import mode' }),
	importButton: localize(locale, { 'zh-Hans': '执行导入', en: 'Run Import' }),
	importing: localize(locale, { 'zh-Hans': '处理中...', en: 'Processing...' }),
	applyButton: localize(locale, { 'zh-Hans': '应用同一份导入', en: 'Apply same import' }),
	applying: localize(locale, { 'zh-Hans': '应用中...', en: 'Applying...' }),
	modelMappings: localize(locale, { 'zh-Hans': '模型映射', en: 'Model Mappings' }),
	structuredMappings: localize(locale, { 'zh-Hans': '结构化映射预览', en: 'Structured Mapping Preview' }),
	activeMappings: localize(locale, { 'zh-Hans': '已配置映射', en: 'Active Mappings' }),
	compatibility: localize(locale, { 'zh-Hans': '兼容性报告', en: 'Compatibility Report' }),
	postImportValidation: localize(locale, { 'zh-Hans': '导入后检查', en: 'Post-import validation' }),
	postImportHealth: localize(locale, { 'zh-Hans': '导入后健康检查', en: 'Post-import health check' }),
	replacePrunePreview: localize(locale, { 'zh-Hans': '替换清理预览', en: 'Replace-prune preview' }),
	show: localize(locale, { 'zh-Hans': '展开', en: 'Show' }),
	hide: localize(locale, { 'zh-Hans': '收起', en: 'Hide' }),
	preview: localize(locale, { 'zh-Hans': '预览', en: 'Preview' }),
	previewing: localize(locale, { 'zh-Hans': '预览中...', en: 'Previewing...' }),
	rollback: localize(locale, { 'zh-Hans': '回滚', en: 'Rollback' }),
	rollingBack: localize(locale, { 'zh-Hans': '回滚中...', en: 'Rolling Back...' }),
	rollbackLatest: localize(locale, { 'zh-Hans': '回滚最近一次导入', en: 'Rollback Latest Import' }),
	rollbackLatestRunning: localize(locale, { 'zh-Hans': '正在回滚最近一次导入...', en: 'Rolling Back Latest Import...' }),
	refresh: localize(locale, { 'zh-Hans': '刷新', en: 'Refresh' }),
	rollbackPreview: localize(locale, { 'zh-Hans': '回滚预览', en: 'Rollback preview' }),
	latest: localize(locale, { 'zh-Hans': '最新', en: 'Latest' }),
	historyItemSize: localize(locale, { 'zh-Hans': '大小', en: 'Size' }),
	importAdvancedPending: localize(locale, { 'zh-Hans': '导入补强项', en: 'Import migration tooling' }),
	advancedPending: localize(locale, { 'zh-Hans': '高级迁移能力仍在持续补齐', en: 'Advanced migration tooling still pending' }),
	empty: localize(locale, { 'zh-Hans': '无', en: 'None' }),
	unknown: localize(locale, { 'zh-Hans': '未知', en: 'Unknown' }),
	yes: localize(locale, { 'zh-Hans': '是', en: 'Yes' }),
	no: localize(locale, { 'zh-Hans': '否', en: 'No' }),
	needScope: localize(locale, { 'zh-Hans': '请至少选中一个导入范围。', en: 'Import stays disabled until at least one scope is selected.' }),
	selectFileFirst: localize(locale, { 'zh-Hans': '请先选择 JSON 备份文件', en: 'Select a JSON backup file first.' }),
	missingPreviewToken: localize(locale, { 'zh-Hans': '预检已完成，但没有返回可用的预检令牌，请重新执行一次预检。', en: 'Preview finished, but no usable preview token was returned. Run the preview again.' }),
	applyNeedConfirm: localize(locale, { 'zh-Hans': '请先确认本次导入风险，再将其应用到当前项目。', en: 'Confirm the import risks before applying to the current project.' }),
	applyRunFirst: localize(locale, { 'zh-Hans': '请先执行一次预检，再应用同一份快照。', en: 'Run a preview first before applying the same snapshot.' }),
	applyConfirm: localize(locale, { 'zh-Hans': '我已经检查上方风险提示，确认可以把这次导入应用到当前项目。', en: 'I reviewed the risks above and want to apply this import to the current project.' }),
	applyMetaFile: localize(locale, { 'zh-Hans': '文件', en: 'File' }),
	applyMetaMode: localize(locale, { 'zh-Hans': '模式', en: 'Mode' }),
	applyMetaScope: localize(locale, { 'zh-Hans': '范围', en: 'Scopes' }),
	applyMetaMappingCount: localize(locale, { 'zh-Hans': '映射数', en: 'Mappings' }),
	applyMetaPreviewToken: localize(locale, { 'zh-Hans': '预检令牌', en: 'Preview Token' }),
	pendingApplyReady: localize(locale, { 'zh-Hans': '预检已完成，可以继续应用同一份快照。', en: 'Preview completed. You can continue applying this snapshot.' }),
	rollbackConfirmFull: localize(locale, { 'zh-Hans': '确定要将当前项目回滚到这份快照吗？', en: 'Roll back the current project to this snapshot?' }),
	rollbackMetaScope: localize(locale, { 'zh-Hans': '回滚范围', en: 'Rollback Scope' }),
	rollbackMetaEncrypted: localize(locale, { 'zh-Hans': '加密状态', en: 'Encryption' }),
	rollbackMetaContainsSecrets: localize(locale, { 'zh-Hans': '包含凭证', en: 'Contains Credentials' }),
	rollbackMetaSchemaVersion: localize(locale, { 'zh-Hans': '结构版本', en: 'Schema Version' }),
	rollbackSummaryConflicts: localize(locale, { 'zh-Hans': '兼容冲突', en: 'Compatibility Conflicts' }),
	rollbackSummaryRebinds: localize(locale, { 'zh-Hans': '凭证重绑定', en: 'Credential Rebinds' }),
	rollbackSummaryWarnings: localize(locale, { 'zh-Hans': '预览预警', en: 'Preview Warnings' }),
	compatibilityHeadlineSafe: localize(locale, { 'zh-Hans': '当前没有明显阻塞风险', en: 'No obvious blocking risks' }),
	compatibilityHeadlineWarning: localize(locale, { 'zh-Hans': '建议先再看一遍', en: 'Review the differences first' }),
	compatibilityHeadlineDanger: localize(locale, { 'zh-Hans': '需要先处理风险', en: 'Resolve the risks first' }),
	diagnosticsCollapsed: localize(locale, { 'zh-Hans': '详细诊断默认折叠，按需展开查看。', en: 'Detailed diagnostics stay collapsed until you need them.' }),
	remainingMigrationSummary: localize(locale, { 'zh-Hans': '默认收起，按需查看仍需手动处理的迁移能力。', en: 'Collapsed by default. Open only when you need the still-manual migration gaps.' }),
	toastExportSuccess: localize(locale, { 'zh-Hans': '已开始导出', en: 'Export started' }),
	toastDryRunSuccess: localize(locale, { 'zh-Hans': '预检完成', en: 'Preview completed' }),
	toastImportSuccess: localize(locale, { 'zh-Hans': '导入已应用', en: 'Import applied' }),
	toastImportFailed: localize(locale, { 'zh-Hans': '导入失败', en: 'Import failed' }),
	toastRollbackFailed: localize(locale, { 'zh-Hans': '回滚失败', en: 'Rollback failed' }),
	loadFailed: localize(locale, { 'zh-Hans': '加载失败', en: 'Failed to load' }),
	noSnapshots: localize(locale, { 'zh-Hans': '暂时还没有导入快照。', en: 'No import snapshots yet.' }),
	loadingSnapshots: localize(locale, { 'zh-Hans': '正在加载快照历史...', en: 'Loading snapshot history...' }),
	importSummaryRowsDryRun: localize(locale, { 'zh-Hans': '预检行数', en: 'Preview Rows' }),
	importSummaryRowsApplied: localize(locale, { 'zh-Hans': '实际写入行数', en: 'Applied Rows' }),
	importSummaryMode: localize(locale, { 'zh-Hans': '当前模式', en: 'Current Mode' }),
	importSummaryConflicts: localize(locale, { 'zh-Hans': '冲突', en: 'Conflicts' }),
	importSummaryRebinds: localize(locale, { 'zh-Hans': '凭证重绑定', en: 'Credential Rebinds' }),
	mappingPreviewTitle: localize(locale, { 'zh-Hans': '模型映射预览', en: 'Model Mapping Previews' }),
	missingMappingTitle: localize(locale, { 'zh-Hans': '缺失的映射目标', en: 'Missing Mapping Targets' }),
	unusedMappingTitle: localize(locale, { 'zh-Hans': '未使用的映射', en: 'Unused Model Mappings' }),
	missingProvidersTitle: localize(locale, { 'zh-Hans': '缺失渠道 / 供应商', en: 'Missing Providers / Channels' }),
	postValidationDegradedGroups: localize(locale, { 'zh-Hans': '降级分组', en: 'Degraded groups' }),
	postValidationEmptyGroups: localize(locale, { 'zh-Hans': '空分组', en: 'Empty groups' }),
	postValidationDisabledChannels: localize(locale, { 'zh-Hans': '已禁用渠道', en: 'Disabled channels' }),
	postValidationChannelsWithoutKeys: localize(locale, { 'zh-Hans': '无密钥渠道', en: 'Channels without keys' }),
	postValidationStaleItemsRemoved: localize(locale, { 'zh-Hans': '已清理过期项', en: 'Stale items removed' }),
	postValidationRouteWarnings: localize(locale, { 'zh-Hans': '路由预警', en: 'Route warnings' }),
	postValidationPriceRuleWarnings: localize(locale, { 'zh-Hans': '价格规则预警', en: 'Price-rule warnings' }),
	postValidationAliasMappings: localize(locale, { 'zh-Hans': '别名映射', en: 'Alias mappings' }),
	postValidationAliasWarnings: localize(locale, { 'zh-Hans': '别名预警', en: 'Alias warnings' }),
	postValidationHealthTargets: localize(locale, { 'zh-Hans': '健康检测目标', en: 'Health-check targets' }),
	postValidationHealthPassed: localize(locale, { 'zh-Hans': '通过数量', en: 'Passed' }),
	historyToggle: localize(locale, { 'zh-Hans': '查看历史', en: 'History' }),
	remainingToggle: localize(locale, { 'zh-Hans': '查看补强', en: 'Tooling' }),
	replaceToggle: localize(locale, { 'zh-Hans': '查看清单', en: 'View details' }),
	help: {
		exportTitle: localize(locale, { 'zh-Hans': '这里导出的是项目快照。保留明文凭证时，适合直接迁移到另一套环境。', en: 'This exports a project snapshot. Keep plaintext credentials only when you need a direct restore in another environment.' }),
		exportFormat: localize(locale, { 'zh-Hans': '标准完整备份适合当前项目迁移；原版兼容备份用于兼容旧导出样本。', en: 'Standard format is best for current-project restore. Legacy format keeps compatibility with older exports.' }),
		dryRun: localize(locale, { 'zh-Hans': '先生成预检结果，不写入数据库，适合先看兼容性和冲突。', en: 'Generate a preview report first without writing to the database.' }),
		selectiveImport: localize(locale, { 'zh-Hans': '只导入你勾选的范围；全部取消后会禁止提交。', en: 'Import only the scopes you check. Submission stays disabled when nothing is selected.' }),
		importTitle: localize(locale, { 'zh-Hans': '先选范围和模式，再跑预检；确认风险后再应用到当前项目。', en: 'Choose scopes and mode first, run preview, then apply only after reviewing risks.' }),
		importMode: localize(locale, { 'zh-Hans': '增量、映射、合并、替换和跳过，对应不同的导入处理方式。', en: 'Incremental, map, merge, replace, and skip control how snapshot data is written.' }),
		modelMappings: localize(locale, { 'zh-Hans': '只有映射导入才需要填写；每行一条，格式为“快照模型=当前模型”。', en: 'Only map mode needs this field. Use one rule per line: snapshot-model=current-model.' }),
		rollbackTitle: localize(locale, { 'zh-Hans': '这里可以先预览最近导入的快照，再决定是否回滚。', en: 'Preview recent import snapshots here before rolling back.' }),
		advancedPending: localize(locale, { 'zh-Hans': '这里列出仍需手动处理的迁移能力缺口，先优先使用现有快照恢复与映射预检。', en: 'These are still-manual migration gaps. Prefer the current snapshot restore and map-preview flow first.' }),
		includeSecretsOn: localize(locale, { 'zh-Hans': '开启后会保留明文凭证，适合直接恢复到另一套环境。', en: 'Plaintext credentials stay in the snapshot, so it can be restored directly elsewhere.' }),
		includeSecretsOff: localize(locale, { 'zh-Hans': '关闭后会做脱敏处理，更适合共享或审阅。', en: 'Credentials are redacted, which is better for sharing or review.' }),
	},
});

function getVisibleScopeLabels(importScopes: DBImportScopes | undefined, options: ReturnType<typeof getImportScopeOptions>) {
	if (!importScopes) return options.map((option) => option.label);
	return options.filter((option) => importScopes[option.key]).map((option) => option.label);
}

function DetailBlock({ title, items, testIdPrefix }: { title: string; items: string[]; testIdPrefix?: string }) {
	if (items.length === 0) return null;
	return (
        <div className="space-y-2 rounded-2xl border border-border/60 bg-background/70 p-3" data-testid={testIdPrefix}>
			<div className="text-sm font-medium text-card-foreground" data-testid={testIdPrefix ? `${testIdPrefix}-title` : undefined}>{title}</div>
			<div className="space-y-2 text-xs text-muted-foreground">
				{items.map((item, index) => <div key={`${title}-${index}`} className="rounded-xl border border-border/40 bg-card px-3 py-2" data-testid={testIdPrefix ? `${testIdPrefix}-item-${index}` : undefined}>{item}</div>)}
			</div>
		</div>
	);
}

function formatPreviewToken(value?: string) {
	if (!value) return '';
	if (value.length <= 20) return value;
	return `${value.slice(0, 8)}...${value.slice(-6)}`;
}

function toDataValue(value: string | number | boolean | null | undefined) {
	if (value === undefined || value === null || value === '') return undefined;
	return String(value);
}

function serializeImportScopes(scopes?: DBImportScopes) {
	if (!scopes) return undefined;
	const enabledKeys = Object.entries(scopes)
		.filter(([, enabled]) => enabled)
		.map(([key]) => key);
	return enabledKeys.length > 0 ? enabledKeys.join(',') : 'none';
}

function SummaryCell(props: { label: string; value: string | number; testId?: string; rawValue?: string | number | boolean }) {
	const { label, value, testId } = props;
	const semanticValue = Object.prototype.hasOwnProperty.call(props, 'rawValue') ? toDataValue(props.rawValue) : toDataValue(value);
	return (
		<div data-testid={testId} data-raw-value={semanticValue}>
			<span>{label}</span>
			<span>{`：${value}`}</span>
		</div>
	);
}

function MetaGridCell(props: { label: string; value: string | number; testId?: string; compact?: boolean; copyValue?: string; rawValue?: string | number | boolean }) {
	const { label, value, testId, compact = false, copyValue } = props;
	const semanticValue = Object.prototype.hasOwnProperty.call(props, 'rawValue') ? toDataValue(props.rawValue) : toDataValue(value);
	return (
        <div className="rounded-2xl border border-border/60 bg-background/70 p-3" data-testid={testId} data-raw-value={semanticValue}>
			<span className="block text-[11px] uppercase tracking-wide text-muted-foreground">{`${label}：`}</span>
			<div className="mt-1 flex min-w-0 items-center gap-2">
				<span className={cn('block font-medium text-card-foreground', compact ? 'min-w-0 flex-1 truncate text-sm' : 'text-sm')} title={typeof value === 'string' ? value : String(value)}>{value}</span>
				{copyValue ? (
					<CopyIconButton
						text={copyValue}
                        className="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border border-border/60 bg-card text-muted-foreground transition-colors hover:text-foreground"
						copyIconClassName="size-3.5"
						checkIconClassName="size-3.5 text-primary"
					/>
				) : null}
			</div>
		</div>
	);
}

export function SettingBackup() {
	const locale = useSettingStore((state) => state.locale);
	const text = backupTextByLocale(locale);
	const exportDB = useExportDB();
	const importDB = useImportDB();
	const scopeOptions = getImportScopeOptions(locale);
	const modeOptions = getImportModeOptions(locale);
	const importSnapshots = useImportSnapshots();
	const previewRollback = usePreviewRollbackImportSnapshot();
	const rollbackImportSnapshot = useRollbackImportSnapshot();
	const rollbackLatestImportSnapshot = useRollbackLatestImportSnapshot();

	const [exportFormat, setExportFormat] = useState<'standard' | 'legacy'>('standard');
	const [includeSecrets, setIncludeSecrets] = useState(true);
	const [includeLogs, setIncludeLogs] = useState(false);
	const [includeStats, setIncludeStats] = useState(false);
	const [file, setFile] = useState<File | null>(null);
	const [dryRun, setDryRun] = useState(true);
	const [selectiveImport, setSelectiveImport] = useState(false);
	const [importMode, setImportMode] = useState<DBImportMode>('incremental');
	const [importScopes, setImportScopes] = useState<DBImportScopes>(defaultImportScopes);
	const [modelMappingsText, setModelMappingsText] = useState('');
	const [pendingApplyRequest, setPendingApplyRequest] = useState<PendingApplyRequest | null>(null);
	const [confirmedApply, setConfirmedApply] = useState(false);
	const [showCompatibilityDetails, setShowCompatibilityDetails] = useState(false);
	const [showHistory, setShowHistory] = useState(false);
	const [showImportRemainingMigration, setShowImportRemainingMigration] = useState(false);
	const [openedImportMigrationSection, setOpenedImportMigrationSection] = useState<number | null>(null);
	const [showHistoryRemainingMigration, setShowHistoryRemainingMigration] = useState(false);
	const [openedHistoryMigrationSection, setOpenedHistoryMigrationSection] = useState<number | null>(null);
	const [showReplacePruneDetails, setShowReplacePruneDetails] = useState(false);
	const [currentRollbackPreview, setCurrentRollbackPreview] = useState<DBRollbackPreviewResult | null>(null);

	const exportPresentation = useMemo(() => getExportSnapshotPresentation({ includeSecrets, includeLogs, includeStats, locale }), [includeSecrets, includeLogs, includeStats, locale]);
	const importResult = importDB.data;
	const compatibility = importResult?.compatibility;
	const compatibilityCounts = useMemo(() => getCompatibilityCounts(compatibility), [compatibility]);
	const compatibilityOverview = useMemo(() => getCompatibilityOverview({ counts: compatibilityCounts, warningsCount: (importResult?.warnings ?? []).length, kind: 'import', locale }), [compatibilityCounts, importResult?.warnings, locale]);
	const compatibilitySignals = useMemo(() => buildCompatibilitySignalItems({ counts: compatibilityCounts, warningsCount: (importResult?.warnings ?? []).length, kind: 'import', locale, includeReplaceModeRisk: true, effectiveMode: importResult?.mode ?? importMode }), [compatibilityCounts, importResult?.warnings, importResult?.mode, importMode, locale]);
	const mappingPreviewItems = useMemo(() => getModelMappingPreviewItems(compatibility?.model_mapping_previews, locale), [compatibility?.model_mapping_previews, locale]);
	const missingMappingItems = useMemo(() => getMissingModelMappingItems(compatibility?.model_mapping_previews, locale), [compatibility?.model_mapping_previews, locale]);
	const unusedMappingItems = useMemo(() => getUnusedModelMappingItems(compatibility?.model_mapping_previews, locale), [compatibility?.model_mapping_previews, locale]);
	const replacePrunePreview = importResult?.replace_prune_preview ?? importResult?.replace_prune ?? importResult?.prune_preview ?? compatibility?.replace_prune_preview ?? compatibility?.replace_prune ?? compatibility?.prune_preview;
	const replacePruneSections = useMemo(() => buildReplacePruneSections(replacePrunePreview, locale), [replacePrunePreview, locale]);
	const postImportSummary = useMemo(() => getPostImportValidationSummary(importResult?.post_import_validation), [importResult?.post_import_validation]);
	const remainingMigrationSections = useMemo(() => getRemainingMigrationToolingSections(locale), [locale]);
	const importRemainingMigrationSections = useMemo(() => remainingMigrationSections.filter((section) => section.key !== 'rollback-tooling'), [remainingMigrationSections]);
	const historyRemainingMigrationSections = useMemo(() => {
		const rollbackSection = remainingMigrationSections.find((section) => section.key === 'rollback-tooling');
		const otherSections = remainingMigrationSections.filter((section) => section.key !== 'rollback-tooling');
		return rollbackSection ? [rollbackSection, ...otherSections] : remainingMigrationSections;
	}, [remainingMigrationSections]);
	const rollbackCounts = useMemo(() => getCompatibilityCounts(currentRollbackPreview?.compatibility), [currentRollbackPreview?.compatibility]);
	const rollbackOverview = useMemo(() => getCompatibilityOverview({ counts: rollbackCounts, warningsCount: (currentRollbackPreview?.preview_warnings ?? []).length, kind: 'rollback', locale }), [rollbackCounts, currentRollbackPreview?.preview_warnings, locale]);
	const rollbackPreviewName = currentRollbackPreview?.snapshot_name ?? text.unknown;
	const rollbackScopeSummary = currentRollbackPreview?.applied_scopes ? joinLocalizedList(getVisibleScopeLabels(currentRollbackPreview.applied_scopes, scopeOptions), locale) : text.unknown;
	const rollbackEncryptedSummary = currentRollbackPreview?.manifest?.encrypted === undefined ? text.unknown : currentRollbackPreview.manifest.encrypted ? text.yes : text.no;
	const rollbackContainsSecretsSummary = currentRollbackPreview?.manifest?.contains_secrets ? text.yes : currentRollbackPreview?.manifest?.contains_secrets === false ? text.no : text.unknown;
	const rollbackSchemaVersionSummary = currentRollbackPreview?.manifest?.schema_version ?? text.unknown;
	const rollbackRebindCount = rollbackCounts.channelKeyRebindTargets || rollbackCounts.credentialRebindTargets || rollbackCounts.apiKeyRebindTargets;
	const rollbackWarningCount = currentRollbackPreview?.preview_warnings?.length ?? 0;

	const activeScopeCount = Object.values(importScopes).filter(Boolean).length;
	const effectiveScopes = selectiveImport ? importScopes : defaultImportScopes;
	const importModeLabel = modeOptions.find((option) => option.value === (importResult?.mode ?? importMode))?.label ?? modeOptions[0].label;
	const pendingApplyModeLabel = modeOptions.find((option) => option.value === pendingApplyRequest?.mode)?.label ?? importModeLabel;

	useEffect(() => {
		const button = document.querySelector('[data-testid="backup-apply-same-import-button"]');
		if (button instanceof HTMLButtonElement) {
			(globalThis as typeof globalThis & { applyButton?: HTMLButtonElement }).applyButton = button;
		}
	}, [pendingApplyRequest, confirmedApply, importResult]);

	function handleFileChange(event: ChangeEvent<HTMLInputElement>) {
		setFile(event.target.files?.[0] ?? null);
		setPendingApplyRequest(null);
		setConfirmedApply(false);
		setShowReplacePruneDetails(false);
	}

	function handleScopeChange(key: keyof DBImportScopes, checked: boolean) {
		setImportScopes((current) => ({ ...current, [key]: checked }));
	}

	function createPendingRequest(previewToken?: string): PendingApplyRequest {
		if (!file) throw new Error(text.selectFileFirst);
		const modelMappings = importMode === 'map' ? parseModelMappings(modelMappingsText, locale) : undefined;
		const scopes = selectiveImport ? importScopes : defaultImportScopes;
		return {
			file,
			mode: importMode,
			modelMappings,
			importScopes: scopes,
			previewToken,
			fileName: file.name,
			mappingCount: Object.keys(modelMappings ?? {}).length,
			scopeLabels: getVisibleScopeLabels(scopes, scopeOptions),
		};
	}

	async function handleExport() {
		try {
			await exportDB.mutateAsync({
				include_secrets: includeSecrets,
				include_logs: includeLogs,
				include_stats: includeStats,
				format: exportFormat,
			} as never);
			toast.success(text.toastExportSuccess);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : text.toastImportFailed);
		}
	}

	async function handleImport() {
		if (!file) {
			toast.error(text.selectFileFirst);
			return;
		}
		if (selectiveImport && activeScopeCount === 0) {
			toast.error(text.needScope);
			return;
		}

		let prepared: PendingApplyRequest;
		try {
			prepared = createPendingRequest();
		} catch (error) {
			toast.error(error instanceof Error ? error.message : text.toastImportFailed);
			return;
		}

		try {
			const result = await importDB.mutateAsync({
				file: prepared.file,
				dryRun,
				mode: prepared.mode,
				modelMappings: prepared.modelMappings,
				importScopes: prepared.importScopes,
				previewToken: prepared.previewToken,
			});

			setConfirmedApply(false);
			setShowCompatibilityDetails(false);
			setShowReplacePruneDetails(false);

			if (result.dry_run) {
				toast.success(text.toastDryRunSuccess);
				if (result.preview_token) {
					setPendingApplyRequest({ ...prepared, previewToken: result.preview_token });
				} else {
					setPendingApplyRequest(null);
					toast.error(text.missingPreviewToken);
				}
				return;
			}

			setPendingApplyRequest(null);
			toast.success(text.toastImportSuccess);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : text.toastImportFailed);
		}
	}

	async function handleApplySameImport() {
		if (!pendingApplyRequest) {
			toast.error(text.applyRunFirst);
			return;
		}
		if (!pendingApplyRequest.previewToken) {
			toast.error(text.missingPreviewToken);
			return;
		}
		if (!confirmedApply) {
			toast.error(text.applyNeedConfirm);
			return;
		}

		try {
			await importDB.mutateAsync({
				file: pendingApplyRequest.file,
				dryRun: false,
				mode: pendingApplyRequest.mode,
				modelMappings: pendingApplyRequest.modelMappings,
				importScopes: pendingApplyRequest.importScopes,
				previewToken: pendingApplyRequest.previewToken,
			});
			setPendingApplyRequest(null);
			setConfirmedApply(false);
			setShowReplacePruneDetails(false);
			toast.success(text.toastImportSuccess);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : text.toastImportFailed);
		}
	}

	async function handlePreviewRollback(snapshot: DBImportSnapshotInfo) {
		if (!snapshot.snapshot_name) return;
		try {
			const preview = await previewRollback.mutateAsync({ snapshotName: snapshot.snapshot_name });
			setCurrentRollbackPreview(preview);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : text.toastRollbackFailed);
		}
	}

	async function handleRollbackSnapshot(snapshot: DBImportSnapshotInfo) {
		if (!snapshot.snapshot_name) return;
		if (!window.confirm(text.rollbackConfirmFull)) return;
		try {
			const result = await rollbackImportSnapshot.mutateAsync({ snapshotName: snapshot.snapshot_name });
			toast.success(result.snapshot_name ?? snapshot.snapshot_name);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : text.toastRollbackFailed);
		}
	}

	async function handleRollbackLatest() {
		try {
			const result = await rollbackLatestImportSnapshot.mutateAsync(undefined as never);
			toast.success(result.snapshot_name ?? text.unknown);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : text.toastRollbackFailed);
		}
	}

	return (
		<div data-testid="backup-page" className="space-y-3 rounded-3xl border border-border bg-card p-3.5">
			<div className="space-y-3">
				<SectionCard icon={<Download className="h-4 w-4" />} title={text.exportTitle} hint={text.help.exportTitle} description={text.exportDescription}>
                    <div className="space-y-3">
                        <div className={cn('rounded-2xl border px-3.5 py-3 text-sm', includeSecrets ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700' : 'border-amber-500/30 bg-amber-500/10 text-amber-700')}>
							<div className="font-semibold">{exportPresentation.summary}</div>
							<div className="mt-2 text-xs leading-5">{exportPresentation.warning}</div>
						</div>
						<div className="flex flex-wrap gap-2">
							{exportPresentation.scopeBadges.map((badge) => <Badge key={badge} variant="secondary">{badge}</Badge>)}
						</div>
						<div className="grid gap-2.5 lg:grid-cols-3">
							<SwitchRow label={exportPresentation.toggleLabel} hint={includeSecrets ? text.help.includeSecretsOn : text.help.includeSecretsOff} checked={includeSecrets} onCheckedChange={setIncludeSecrets} />
							<SwitchRow label={text.includeLogs} checked={includeLogs} onCheckedChange={setIncludeLogs} />
							<SwitchRow label={text.includeStats} checked={includeStats} onCheckedChange={setIncludeStats} />
						</div>
                        <div className="space-y-2 rounded-2xl border border-border/60 bg-card/70 p-3.5">
							<div className="text-sm font-medium text-card-foreground"><InlineHelpLabel hint={text.help.exportFormat}>{text.exportFormat}</InlineHelpLabel></div>
							<Select value={exportFormat} onValueChange={(value) => setExportFormat(value as 'standard' | 'legacy')}>
								<SelectTrigger className="w-full rounded-xl"><SelectValue /></SelectTrigger>
								<SelectContent>
									<SelectItem value="standard">{text.exportFormatStandard}</SelectItem>
									<SelectItem value="legacy">{text.exportFormatLegacy}</SelectItem>
								</SelectContent>
							</Select>
						</div>
						<div className="flex justify-end">
							<Button type="button" onClick={handleExport} disabled={exportDB.isPending} className="rounded-xl">{exportDB.isPending ? text.exporting : text.exportButton}</Button>
						</div>
					</div>
				</SectionCard>

				<SectionCard icon={<Upload className="h-4 w-4" />} title={text.importTitle} hint={text.help.importTitle} description={text.importDescription}>
					<div className="grid gap-2.5">
				<div className="space-y-2.5 rounded-2xl border border-border/60 bg-card/70 p-3">
							<div className="rounded-2xl border border-dashed border-border/70 bg-background/70 p-3">
								<div className="mb-2 text-xs text-muted-foreground"><InlineHelpLabel hint={text.selectFileFirst}>{text.selectedFile}</InlineHelpLabel></div>
								<Input type="file" accept=".json,application/json" onChange={handleFileChange} className="rounded-xl" />
								{file ? (
							<div className="mt-2.5 rounded-xl border border-border/50 bg-card px-3 py-2 text-xs text-muted-foreground">
										<div className="font-medium text-card-foreground">{text.selectedFile}</div>
										<div className="mt-1 break-all">{file.name}</div>
										<div className="mt-1">{formatFileSize(file.size)}</div>
									</div>
								) : null}
							</div>
							<div className="grid gap-2.5 lg:grid-cols-2">
								<SwitchRow label={text.dryRun} hint={text.help.dryRun} checked={dryRun} onCheckedChange={setDryRun} />
								<SwitchRow label={text.selectiveImport} hint={text.help.selectiveImport} checked={selectiveImport} onCheckedChange={setSelectiveImport} />
							</div>
                            <div className="rounded-2xl border border-border/60 bg-card/70 p-3.5 space-y-2">
								<div className="text-sm font-medium text-card-foreground"><InlineHelpLabel hint={text.help.importMode}>{text.importMode}</InlineHelpLabel></div>
								<Select value={importMode} onValueChange={(value) => setImportMode(value as DBImportMode)}>
									<SelectTrigger className="w-full rounded-xl"><SelectValue /></SelectTrigger>
									<SelectContent>
										{modeOptions.map((option) => <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>)}
									</SelectContent>
								</Select>
							</div>
							{selectiveImport ? (
								<div className="grid gap-2.5 lg:grid-cols-2">
									{scopeOptions.map((option) => <SwitchRow key={option.key} label={option.label} checked={importScopes[option.key]} onCheckedChange={(checked) => handleScopeChange(option.key, checked)} />)}
								</div>
							) : null}
							{selectiveImport && activeScopeCount === 0 ? <div className="rounded-2xl border border-amber-500/30 bg-amber-500/10 px-3.5 py-3 text-sm text-amber-700">{text.needScope}</div> : null}
							{importMode === 'map' ? (
                                <div className="space-y-3 rounded-2xl border border-border/60 bg-card/70 p-3.5" data-testid="backup-map-preview-root">
									<div className="text-sm font-medium text-card-foreground"><InlineHelpLabel hint={text.help.modelMappings}>{text.modelMappings}</InlineHelpLabel></div>
									<textarea value={modelMappingsText} onChange={(event) => setModelMappingsText(event.target.value)} placeholder={getLocalizedModelMappingsPlaceholder(locale)} className="min-h-28 w-full rounded-2xl border border-input bg-background px-3 py-3 text-sm outline-none transition focus-visible:ring-2 focus-visible:ring-ring/40" />
                                    <div className="rounded-2xl border border-border/60 bg-background/70 p-3 text-xs text-muted-foreground">
										<div className="font-medium text-card-foreground">{text.structuredMappings}</div>
										<div className="mt-2">{`${text.activeMappings}：${Object.keys(importMode === 'map' ? (() => { try { return parseModelMappings(modelMappingsText, locale); } catch { return {}; } })() : {}).length}`}</div>
									</div>
								</div>
							) : null}
							<div className="rounded-2xl border border-border/60 bg-background/70 p-3">
								<div className="flex items-center justify-between gap-3">
									<div className="space-y-1">
									<div className="text-sm font-semibold text-card-foreground" data-testid="backup-import-remaining-migration-title"><InlineHelpLabel hint={text.help.advancedPending}>{text.importAdvancedPending}</InlineHelpLabel></div>
									<div className="text-xs text-muted-foreground" data-testid="backup-import-remaining-migration-summary">{text.remainingMigrationSummary}</div>
								</div>
								<Button type="button" variant="outline" size="sm" className="h-8 rounded-lg" data-testid="backup-import-remaining-migration-trigger" onClick={() => setShowImportRemainingMigration((current) => !current)}>{showImportRemainingMigration ? text.hide : text.remainingToggle}</Button>
							</div>
							<div className={cn('mt-2.5 space-y-2', showImportRemainingMigration ? '' : 'hidden')} data-testid="backup-import-remaining-migration-panel" aria-hidden={showImportRemainingMigration ? undefined : true}>
								{importRemainingMigrationSections.map((section, index) => (
								<div key={`import-remaining-${section.title}`} className="rounded-2xl border border-border/50 bg-card/70 px-3.5 py-2.5">
										<button type="button" className="w-full text-left" data-slot="accordion-trigger" data-testid={`backup-import-remaining-migration-section-trigger-${index}`} onClick={() => setOpenedImportMigrationSection((current) => current === index ? null : index)}>
											<div className="space-y-1 text-card-foreground">
												<div className="text-sm font-medium">{section.title}</div>
												<div className="text-xs leading-5 text-muted-foreground">{section.summary}</div>
											</div>
										</button>
									<div className={cn('mt-2.5 space-y-2', openedImportMigrationSection === index ? '' : 'hidden')} data-testid={`backup-import-remaining-migration-section-panel-${index}`} aria-hidden={openedImportMigrationSection === index ? undefined : true}>{openedImportMigrationSection === index ? section.items.map((item) => <div key={item.key} className="rounded-xl border border-border/40 bg-background/70 px-3 py-2" data-testid={`backup-import-remaining-migration-section-item-${section.key}-${item.key}`}><div className="text-sm font-medium text-card-foreground" data-testid={`backup-import-remaining-migration-section-item-${section.key}-${item.key}-label`}>{item.label}</div><div className="mt-1 leading-5 text-muted-foreground" data-testid={`backup-import-remaining-migration-section-item-${section.key}-${item.key}-text`}>{item.text}</div></div>) : null}</div>
									</div>
								))}
							</div>
							</div>
							<div className="flex justify-end">
								<Button type="button" data-testid="backup-import-button" onClick={handleImport} disabled={!file || importDB.isPending || (selectiveImport && activeScopeCount === 0)} className="rounded-xl">{importDB.isPending ? text.importing : text.importButton}</Button>
							</div>
						</div>

						<div className="space-y-2.5 rounded-2xl border border-border/60 bg-card/70 p-3">
						<div className={cn('rounded-2xl border px-3.5 py-3 text-sm', toneClasses(compatibilityOverview.tone))} data-testid="backup-compatibility-overview">
								<div className="flex items-start gap-3">
									{compatibilityOverview.tone === 'danger' ? <ShieldAlert className="mt-0.5 h-4 w-4 shrink-0" /> : <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0" />}
									<div className="space-y-1">
										<div className="font-semibold">{compatibilityOverview.tone === 'danger' ? text.compatibilityHeadlineDanger : compatibilityOverview.tone === 'warning' ? text.compatibilityHeadlineWarning : text.compatibilityHeadlineSafe}</div>
										<div className="text-xs leading-5">{compatibilityOverview.description}</div>
									</div>
								</div>
							</div>

							{pendingApplyRequest ? (
							<div className="space-y-2.5 rounded-2xl border border-amber-500/30 bg-amber-500/8 p-3" data-testid="backup-pending-apply-panel">
									<div className="text-sm font-semibold text-card-foreground" data-testid="backup-pending-apply-ready"><InlineHelpLabel hint={text.applyNeedConfirm}>{text.pendingApplyReady}</InlineHelpLabel></div>
									<div className="grid gap-2 text-xs text-muted-foreground lg:grid-cols-2" data-testid="backup-pending-apply-meta-grid">
										<MetaGridCell label={text.applyMetaFile} value={pendingApplyRequest.fileName} testId="backup-pending-apply-meta-file" />
										<MetaGridCell label={text.applyMetaMode} value={pendingApplyModeLabel} testId="backup-pending-apply-meta-mode" />
										<MetaGridCell label={text.applyMetaScope} value={joinLocalizedList(pendingApplyRequest.scopeLabels, locale)} testId="backup-pending-apply-meta-scope" />
										<MetaGridCell label={text.applyMetaMappingCount} value={pendingApplyRequest.mappingCount} testId="backup-pending-apply-meta-mapping-count" />
										<MetaGridCell label={text.applyMetaPreviewToken} value={formatPreviewToken(pendingApplyRequest.previewToken)} copyValue={pendingApplyRequest.previewToken} compact testId="backup-pending-apply-meta-preview-token" />
									</div>
								</div>
							) : null}

							{importResult ? (
							<div className="space-y-2.5">
								<div className="space-y-2.5 rounded-2xl border border-border/60 bg-card/70 p-3" data-testid="backup-import-result-panel">
										<div className="text-sm font-semibold text-card-foreground">{importResult.dry_run ? text.compatibility : text.postImportValidation}</div>
									<div className="grid gap-2.5 text-xs text-muted-foreground sm:grid-cols-2" data-testid="backup-import-summary-grid">
											<SummaryCell label={importResult.dry_run ? text.importSummaryRowsDryRun : text.importSummaryRowsApplied} value={getRowsAffectedTotal(importResult)} />
											<SummaryCell label={text.importSummaryMode} value={importModeLabel} />
											<SummaryCell label={text.importSummaryConflicts} value={compatibilityCounts.conflicts} />
											<SummaryCell label={text.importSummaryRebinds} value={compatibilityCounts.channelKeyRebindTargets || compatibilityCounts.credentialRebindTargets || compatibilityCounts.apiKeyRebindTargets} />
										</div>
									</div>

								<div className="space-y-2.5 rounded-2xl border border-border/60 bg-card/70 p-3" data-testid="backup-compatibility-panel">
										<div className="flex items-center justify-between gap-3">
											<div className="space-y-1">
												<div className="text-sm font-semibold text-card-foreground" data-testid="backup-compatibility-title"><InlineHelpLabel hint={text.diagnosticsCollapsed}>{text.compatibility}</InlineHelpLabel></div>
												<div className="text-xs text-muted-foreground" data-testid="backup-compatibility-summary">{text.diagnosticsCollapsed}</div>
											</div>
											<Button type="button" variant="outline" size="sm" className="h-8 rounded-lg" data-testid="backup-compatibility-toggle" onClick={() => setShowCompatibilityDetails((current) => !current)}>
												{showCompatibilityDetails ? text.hide : `${text.show} ${Math.max(compatibilitySignals.length, 1)}`}
											</Button>
										</div>
										<div data-testid="backup-compatibility-details">
											{showCompatibilityDetails ? (
												<div className="space-y-3 text-xs text-muted-foreground">
													<div className="space-y-2" data-testid="backup-compatibility-signal-list">
														{compatibilitySignals.map((item, index) => <div key={item} className="rounded-xl border border-border/40 bg-background/70 px-3 py-2" data-testid={`backup-compatibility-signal-${index}`}>{item}</div>)}
													</div>
													<DetailBlock title={text.missingProvidersTitle} items={compatibility?.missing_providers ?? []} testIdPrefix="backup-compatibility-missing-providers" />
													<DetailBlock title={text.mappingPreviewTitle} items={mappingPreviewItems} testIdPrefix="backup-compatibility-mapping-preview" />
													<DetailBlock title={text.missingMappingTitle} items={missingMappingItems} testIdPrefix="backup-compatibility-missing-mapping" />
													<DetailBlock title={text.unusedMappingTitle} items={unusedMappingItems} testIdPrefix="backup-compatibility-unused-mapping" />
												</div>
											) : null}
										</div>
									</div>

									{replacePruneSections.length > 0 ? (
                                        <div className="space-y-3 rounded-2xl border border-border/60 bg-card/70 p-3.5" data-testid="backup-replace-prune-panel">
											<div className="flex items-center justify-between gap-3">
												<div className="space-y-1">
													<div className="text-sm font-semibold text-card-foreground" data-testid="backup-replace-prune-title"><InlineHelpLabel hint={locale === 'en' ? 'Replace mode previews records that will be deleted or reset before you apply it.' : '替换导入会先预览要删除或重置的现有记录，再决定是否应用。'}>{text.replacePrunePreview}</InlineHelpLabel></div>
													<div className="text-xs text-muted-foreground" data-testid="backup-replace-prune-summary">{locale === 'en' ? `${replacePruneSections.reduce((sum, section) => sum + section.items.length, 0)} records are hidden by default. Expand only when you need to review the replace scope.` : `默认收起 ${replacePruneSections.reduce((sum, section) => sum + section.items.length, 0)} 项待清理记录，确认替换范围时再展开。`}</div>
												</div>
											<Button type="button" variant="outline" size="sm" className="h-8 rounded-lg" data-testid="backup-replace-prune-trigger" onClick={() => setShowReplacePruneDetails((current) => !current)}>{showReplacePruneDetails ? text.hide : text.replaceToggle}</Button>
										</div>
										{showReplacePruneDetails ? (
											<div className="space-y-3">
												{replacePruneSections.map((section) => (
                                                    <div key={section.key} className="rounded-2xl border border-border/60 bg-background/70 p-3" data-testid={`backup-replace-prune-section-${section.key}`}>
														<div className="text-sm font-medium text-card-foreground" data-testid={`backup-replace-prune-section-title-${section.key}`}>{showReplacePruneDetails ? section.title : ''}</div>
														<div className="mt-3 space-y-2 text-xs text-muted-foreground">
															{section.items.map((item, index) => <div key={`${section.key}-${index}`} className="rounded-xl border border-border/40 bg-card px-3 py-2" data-testid={`backup-replace-prune-section-item-${section.key}-${index}`}>{showReplacePruneDetails ? item : ''}</div>)}
														</div>
													</div>
												))}
											</div>
										) : null}
									</div>
								) : null}

									{pendingApplyRequest ? (
                                        <div className="space-y-3 rounded-2xl border border-border/60 bg-card/70 p-3.5" data-testid="backup-apply-confirm-panel">
											<SwitchRow label={text.applyConfirm} checked={confirmedApply} onCheckedChange={setConfirmedApply} testId="backup-apply-confirm-switch" />
											<div className="flex justify-end">
												<Button type="button" className="rounded-xl" data-testid="backup-apply-same-import-button" disabled={!confirmedApply || !pendingApplyRequest.previewToken || importDB.isPending} onClick={handleApplySameImport}>{importDB.isPending ? text.applying : text.applyButton}</Button>
											</div>
										</div>
									) : null}
								</div>
							) : null}

							{postImportSummary ? (
                                <div className="space-y-3 rounded-2xl border border-border/60 bg-card/70 p-3.5" data-testid="backup-post-import-validation-panel">
									<div className="text-sm font-semibold text-card-foreground"><InlineHelpLabel hint={text.postImportHealth}>{text.postImportValidation}</InlineHelpLabel></div>
									<div className="grid gap-2.5 text-xs text-muted-foreground sm:grid-cols-2 xl:grid-cols-3" data-testid="backup-post-import-validation-summary-grid">
										<SummaryCell label={text.postValidationDegradedGroups} value={postImportSummary.degradedGroups} testId="backup-post-import-validation-summary-degraded-groups" />
										<SummaryCell label={text.postValidationEmptyGroups} value={postImportSummary.emptyGroups} testId="backup-post-import-validation-summary-empty-groups" />
										<SummaryCell label={text.postValidationDisabledChannels} value={postImportSummary.disabledChannels} testId="backup-post-import-validation-summary-disabled-channels" />
										<SummaryCell label={text.postValidationChannelsWithoutKeys} value={postImportSummary.channelsWithoutKeys} testId="backup-post-import-validation-summary-channels-without-keys" />
										<SummaryCell label={text.postValidationStaleItemsRemoved} value={postImportSummary.staleItemsRemoved} testId="backup-post-import-validation-summary-stale-items-removed" />
										<SummaryCell label={text.postValidationRouteWarnings} value={postImportSummary.routeWarnings} testId="backup-post-import-validation-summary-route-warnings" />
										<SummaryCell label={text.postValidationPriceRuleWarnings} value={postImportSummary.priceRuleWarnings} testId="backup-post-import-validation-summary-price-rule-warnings" />
										<SummaryCell label={text.postValidationAliasMappings} value={postImportSummary.aliasMappings} testId="backup-post-import-validation-summary-alias-mappings" />
										<SummaryCell label={text.postValidationAliasWarnings} value={postImportSummary.aliasWarnings} testId="backup-post-import-validation-summary-alias-warnings" />
									</div>
									{importResult?.post_import_validation?.health_check?.summary ? (
                                        <div className="space-y-3 rounded-2xl border border-border/60 bg-background/70 p-3 text-xs text-muted-foreground" data-testid="backup-post-import-health-summary">
										<div className="grid gap-2.5 sm:grid-cols-2" data-testid="backup-post-import-health-summary-grid">
												<SummaryCell label={text.postValidationHealthTargets} value={importResult.post_import_validation.health_check.summary.targets ?? 0} testId="backup-post-import-health-summary-targets" />
												<SummaryCell label={text.postValidationHealthPassed} value={importResult.post_import_validation.health_check.summary.passed ?? 0} testId="backup-post-import-health-summary-passed" />
											</div>
										</div>
									) : null}
								</div>
							) : null}
					</div>
				</div>
				</SectionCard>
			</div>

			<SectionCard icon={<History className="h-4 w-4" />} title={text.rollbackTitle} hint={text.help.rollbackTitle} description={text.rollbackDescription}>
                <div className="space-y-3 rounded-2xl border border-border/60 bg-card/70 p-3.5">
					<div className="flex flex-wrap items-center justify-between gap-3">
						<div className="space-y-1">
							<div className="text-sm font-semibold text-card-foreground"><InlineHelpLabel hint={text.help.rollbackTitle}>{text.rollbackTitle}</InlineHelpLabel></div>
							<div className="text-xs text-muted-foreground">{text.rollbackDescription}</div>
						</div>
						<div className="flex flex-wrap gap-2">
							<Button type="button" variant="outline" size="sm" className="h-8 rounded-lg" data-testid="backup-history-trigger" onClick={() => setShowHistory((current) => !current)}>{showHistory ? text.hide : text.show}</Button>
							<Button type="button" variant="outline" size="sm" className="h-8 rounded-lg" onClick={handleRollbackLatest}>{rollbackLatestImportSnapshot.isPending ? text.rollbackLatestRunning : text.rollbackLatest}</Button>
							<Button type="button" variant="outline" size="sm" className="h-8 rounded-lg" onClick={() => importSnapshots.refetch()}>{text.refresh}</Button>
						</div>
					</div>

					{showHistory ? (
						<div className="space-y-2.5" data-testid="backup-history-panel">
							{importSnapshots.isError ? <div className="rounded-2xl border border-red-500/30 bg-red-500/10 p-3 text-sm text-red-700">{text.loadFailed}</div> : null}
                            {importSnapshots.isLoading ? <div className="rounded-2xl border border-border/60 bg-background/70 p-3 text-sm text-muted-foreground">{text.loadingSnapshots}</div> : null}
							<div className="space-y-2" data-testid="backup-history-list">
								{!importSnapshots.isLoading && !importSnapshots.data?.length ? <div className="rounded-2xl border border-border/60 bg-background/70 p-3 text-sm text-muted-foreground">{text.noSnapshots}</div> : null}
								{(importSnapshots.data ?? []).map((snapshot) => (
									<div key={snapshot.snapshot_name ?? snapshot.snapshot_path} className="rounded-2xl border border-border/60 bg-background/70 p-3" data-testid={getSnapshotHistoryItemTestId(snapshot)}>
										<div className="flex flex-wrap items-start justify-between gap-3" data-testid="backup-history-item-meta">
											<div className="min-w-0 flex-1">
												<div className="break-all text-sm font-medium text-card-foreground" data-testid="backup-history-item-name">{snapshot.snapshot_name ?? snapshot.snapshot_path ?? text.unknown}</div>
													<div className="mt-1 break-all text-xs text-muted-foreground" data-testid="backup-history-item-path" data-raw-value={toDataValue(snapshot.snapshot_path)}>{snapshot.snapshot_path ?? text.unknown}</div>
													<div className="mt-1 text-xs text-muted-foreground" data-testid="backup-history-item-size" data-size-bytes={toDataValue(snapshot.size_bytes)}>{snapshot.size_bytes !== undefined && snapshot.size_bytes !== null ? `Size：${formatFileSize(snapshot.size_bytes)}` : `Size：${text.unknown}`}</div>
													<div className="mt-1 text-xs text-muted-foreground" data-testid="backup-history-item-imported-at" data-raw-value={toDataValue(snapshot.imported_at)}>{snapshot.imported_at ? formatDateTimeByLocale(snapshot.imported_at, locale) : text.unknown}</div>
												</div>
												<div className="rounded-full border border-border/40 px-2 py-1 text-[11px] text-muted-foreground" data-testid="backup-history-item-latest-badge" data-is-latest={snapshot.is_latest ? 'true' : 'false'}>{snapshot.is_latest ? text.latest : text.empty}</div>
											</div>
										<div className="mt-2.5 flex flex-wrap gap-2" data-testid="backup-history-item-actions">
											<Button type="button" variant="outline" size="sm" className="h-8 rounded-lg" data-testid="backup-history-preview-button" onClick={() => handlePreviewRollback(snapshot)}>{previewRollback.isPending ? text.previewing : text.preview}</Button>
											<Button type="button" variant="outline" size="sm" className="h-8 rounded-lg" data-testid="backup-history-rollback-button" onClick={() => handleRollbackSnapshot(snapshot)}>{rollbackImportSnapshot.isPending ? text.rollingBack : text.rollback}</Button>
										</div>
									</div>
								))}
							</div>
						</div>
					) : null}

                    <div className="rounded-2xl border border-border/60 bg-background/70 p-3">
						<div className="flex items-center justify-between gap-3">
							<div className="space-y-1">
								<div className="text-sm font-semibold text-card-foreground" data-testid="backup-advanced-pending-title"><InlineHelpLabel hint={text.help.advancedPending}>{text.advancedPending}</InlineHelpLabel></div>
								<div className="text-xs text-muted-foreground" data-testid="backup-advanced-pending-summary">{text.remainingMigrationSummary}</div>
							</div>
							<Button type="button" variant="outline" size="sm" className="h-8 rounded-lg" data-testid="backup-remaining-migration-trigger" onClick={() => setShowHistoryRemainingMigration((current) => !current)}>{showHistoryRemainingMigration ? text.hide : text.remainingToggle}</Button>
						</div>
							{showHistoryRemainingMigration ? (
								<div className="mt-2 space-y-2" data-testid="backup-remaining-migration-panel">
								{historyRemainingMigrationSections.map((section, index) => (
									<div key={section.title} className="rounded-2xl border border-border/50 bg-card/70 px-3 py-2.5">
										<button type="button" className="w-full text-left" data-slot="accordion-trigger" data-testid={`backup-remaining-migration-section-trigger-${index}`} onClick={() => setOpenedHistoryMigrationSection((current) => current === index ? null : index)}>
											<div className="space-y-1 text-card-foreground">
												<div className="text-sm font-medium">{section.title}</div>
												<div className="text-xs leading-5 text-muted-foreground">{section.summary}</div>
											</div>
										</button>
										{openedHistoryMigrationSection === index ? (
											<div className="mt-2.5 space-y-2" data-testid={`backup-remaining-migration-section-panel-${index}`}>
												{section.items.map((item) => <div key={item.key} className="rounded-xl border border-border/40 bg-background/70 px-3 py-2" data-testid={`backup-remaining-migration-section-item-${section.key}-${item.key}`}><div className="text-sm font-medium text-card-foreground" data-testid={`backup-remaining-migration-section-item-${section.key}-${item.key}-label`}>{item.label}</div><div className="mt-1 leading-5 text-muted-foreground" data-testid={`backup-remaining-migration-section-item-${section.key}-${item.key}-text`}>{item.text}</div></div>)}
											</div>
										) : null}
									</div>
								))}
							</div>
						) : null}

							{currentRollbackPreview ? (
                                <div className="space-y-2.5 rounded-2xl border border-border/60 bg-card/70 p-3" data-testid="backup-rollback-preview-panel">
									<div className="space-y-1" data-testid="backup-rollback-preview-header">
										<div className="text-sm font-semibold text-card-foreground" data-testid="backup-rollback-preview-title">{text.rollbackPreview}</div>
										<div className="text-xs text-muted-foreground" data-testid="backup-rollback-preview-name" data-raw-value={toDataValue(currentRollbackPreview?.snapshot_name)}>{rollbackPreviewName}</div>
									</div>
									<div className={cn('rounded-2xl border px-3 py-2.5 text-xs', toneClasses(rollbackOverview.tone))} data-testid="backup-rollback-preview-overview">
										{rollbackOverview.description}
									</div>
									<div className="grid gap-2.5 text-xs text-muted-foreground md:grid-cols-3" data-testid="backup-rollback-preview-summary-grid">
										<SummaryCell label={text.rollbackSummaryConflicts} value={rollbackCounts.conflicts} testId="backup-rollback-preview-summary-conflicts" />
										<SummaryCell label={text.rollbackSummaryRebinds} value={rollbackRebindCount} testId="backup-rollback-preview-summary-rebinds" />
										<SummaryCell label={text.rollbackSummaryWarnings} value={rollbackWarningCount} testId="backup-rollback-preview-summary-warnings" />
									</div>
									<div className="grid gap-2.5 text-xs text-muted-foreground sm:grid-cols-2 xl:grid-cols-2" data-testid="backup-rollback-preview-meta-grid">
										<MetaGridCell label={text.rollbackMetaScope} value={rollbackScopeSummary} rawValue={serializeImportScopes(currentRollbackPreview?.applied_scopes)} testId="backup-rollback-preview-meta-scope" />
										<MetaGridCell label={text.rollbackMetaEncrypted} value={rollbackEncryptedSummary} rawValue={currentRollbackPreview?.manifest?.encrypted} testId="backup-rollback-preview-meta-encrypted" />
										<MetaGridCell label={text.rollbackMetaContainsSecrets} value={rollbackContainsSecretsSummary} rawValue={currentRollbackPreview?.manifest?.contains_secrets} testId="backup-rollback-preview-meta-contains-secrets" />
										<MetaGridCell label={text.rollbackMetaSchemaVersion} value={rollbackSchemaVersionSummary} rawValue={currentRollbackPreview?.manifest?.schema_version} testId="backup-rollback-preview-meta-schema-version" />
									</div>
								</div>
							) : null}
				</div>
			</div>
			</SectionCard>
		</div>
	);
}


