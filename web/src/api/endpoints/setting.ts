import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient, buildApiUrl, getResolvedAuthToken } from '../client';
import { logger } from '@/lib/logger';
import { useAuthStore } from './user';

/**
 * Setting 閺佺増宓?
 */
export interface Setting {
    key: string;
    value: string;
}

export const SettingKey = {
    ProxyURL: 'proxy_url',
    StatsSaveInterval: 'stats_save_interval',
    ModelInfoUpdateInterval: 'model_info_update_interval',
    SyncLLMInterval: 'sync_llm_interval',
    RelayLogKeepEnabled: 'relay_log_keep_enabled',
    RelayLogKeepPeriod: 'relay_log_keep_period',
    CORSAllowOrigins: 'cors_allow_origins',
    CircuitBreakerThreshold: 'circuit_breaker_threshold',
    CircuitBreakerCooldown: 'circuit_breaker_cooldown',
    CircuitBreakerMaxCooldown: 'circuit_breaker_max_cooldown',
    DynamicRoutingMode: 'dynamic_routing_mode',
    DynamicRoutingHealthEnabled: 'dynamic_routing_health_enabled',
    DynamicRoutingLearningEnabled: 'dynamic_routing_learning_enabled',
    RaceGlobalBudget: 'race_global_budget',
    RaceGroupBudget: 'race_group_budget',
    RaceChannelBudget: 'race_channel_budget',
    RaceKeyBudget: 'race_key_budget',
    RaceProbeBudget: 'race_probe_budget',
    AIAutomationEnabled: 'ai_automation_enabled',
    AIAutomationBaseUrl: 'ai_automation_base_url',
    AIAutomationAPIKey: 'ai_automation_api_key',
    AIAutomationChannelType: 'ai_automation_channel_type',
    AIAutomationModel: 'ai_automation_model',
    AIAutomationUseLocalDefault: 'ai_automation_use_local_default',
    ConfigSourceMode: 'config_source_mode',
    ActiveAIProfileID: 'active_ai_profile_id',
    ApiBaseUrl: 'api_base_url',
    ApiAlternateBaseUrls: 'api_alternate_base_urls',
    TrustedProxyCIDRs: 'trusted_proxy_cidrs',
    OpsIPDisplayMode: 'ops_ip_display_mode',
} as const;

export interface PublicAccessInfo {
    primary_base_url: string;
    alternate_base_urls: string[];
    current_base_url: string;
    trusted_proxy_cidrs: string[];
    ops_ip_display_mode: 'masked' | 'full';
    current_client_ip: string;
    current_client_label: string;
}

/**
 * 閼惧嘲褰?Setting 閸掓銆?Hook
 * 
 * @example
 * const { data: settings, isLoading, error } = useSettingList();
 * 
 * if (isLoading) return <Loading />;
 * if (error) return <Error message={error.message} />;
 * 
 * settings?.forEach(setting => console.log(setting.key, setting.value));
 */
export function useSettingList() {
    return useQuery({
        queryKey: ['settings', 'list'],
        queryFn: async () => {
            return apiClient.get<Setting[]>('/api/v1/setting/list');
        },
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

export function usePublicAccess() {
    return useQuery({
        queryKey: ['settings', 'public-access'],
        queryFn: async () => apiClient.get<PublicAccessInfo>('/api/v1/setting/public-access'),
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

/**
 * 鐠佸墽鐤?Setting Hook
 * 
 * @example
 * const setSetting = useSetSetting();
 * 
 * setSetting.mutate({
 *   key: 'theme',
 *   value: 'dark',
 * });
 */
export function useSetSetting() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: Setting) => {
            return apiClient.post<Setting>('/api/v1/setting/set', data);
        },
        onSuccess: (data) => {
            logger.log('Setting 鐠佸墽鐤嗛幋鎰:', data);
            queryClient.invalidateQueries({ queryKey: ['settings', 'list'] });
            queryClient.invalidateQueries({ queryKey: ['settings', 'public-access'] });
        },
        onError: (error) => {
            logger.error('Setting 鐠佸墽鐤嗘径杈Е:', error);
        },
    });
}

/**
 * 閺佺増宓佹惔鎾愁嚤閸?鐎电厧鍤?
 */
export interface DBImportResult {
    rows_affected: Record<string, number>;
    preview_token?: string;
    replace_prune_preview?: DBReplacePrunePreview;
    replace_prune?: DBReplacePrunePreview;
    prune_preview?: DBReplacePrunePreview;
    warnings?: string[];
    dry_run?: boolean;
    mode?: 'incremental' | 'map' | 'merge' | 'replace' | 'skip';
    manifest?: {
        schema_version?: string;
        export_source?: string;
        checksum?: string;
        encrypted?: boolean;
        contains_secrets?: boolean;
    };
    compatibility?: {
        summary?: {
            missing_providers: number;
            missing_models: number;
            conflicts: number;
            alias_conflicts: number;
            credential_rebind_targets: number;
            channel_key_rebind_targets: number;
            api_key_rebind_targets: number;
            model_mapping_previews: number;
            used_model_mappings: number;
            unused_model_mappings: number;
            missing_mapping_targets: number;
            alias_preview_mappings: number;
            model_policy_diffs: number;
            route_conflicts: number;
            invalid_route_targets: number;
            skipped_route_target_previews: number;
            route_preview_warnings: number;
            route_preview_diffs: number;
            base_url_mismatches: number;
            schema_mismatches: number;
            skipped_targets: number;
            replace_pruned_channels: number;
            replace_pruned_groups: number;
            replace_pruned_settings: number;
            replace_pruned_llm_infos: number;
            replace_pruned_api_keys: number;
        };
        affected_groups?: string[];
        affected_channels?: string[];
        missing_providers?: string[];
        missing_models?: string[];
        conflicts?: string[];
        alias_conflicts?: string[];
        credential_rebind_targets?: Array<{
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
        }>;
        model_mapping_previews?: Array<{
            source_model: string;
            target_model: string;
            contexts?: string[];
            touched_fields?: string[];
            usage_count?: number;
            used?: boolean;
            target_exists?: boolean;
            warnings?: string[];
        }>;
        alias_preview_mappings?: Array<{
            snapshot_model: string;
            current_model: string;
            canonical?: string;
            contexts?: string[];
        }>;
        model_policy_diffs?: Array<{
            model: string;
            current_model?: string;
            canonical?: string;
            before?: {
                billing_mode?: string;
                probe_policy?: string;
                probe_interval?: number;
                probe_concurrency?: number;
            };
            after?: {
                billing_mode?: string;
                probe_policy?: string;
                probe_interval?: number;
                probe_concurrency?: number;
            };
            changed_fields?: string[];
            impact_level?: string;
            warnings?: string[];
            contexts?: string[];
            skip_reasons?: string[];
        }>;
        route_conflicts?: string[];
        base_url_mismatches?: string[];
        schema_mismatches?: string[];
        skipped_targets?: string[];
        replace_pruned_channels?: string[];
        replace_pruned_groups?: string[];
        replace_pruned_settings?: string[];
        replace_pruned_llm_infos?: string[];
        replace_pruned_api_keys?: string[];
        route_preview_warnings?: string[];
        invalid_route_targets?: Array<{
            group_name?: string;
            channel_name?: string;
            model?: string;
            resolved_model?: string;
            key_id?: number;
            issue_type: string;
            reason?: string;
            action?: string;
        }>;
        skipped_route_target_previews?: Array<{
            group_name?: string;
            channel_name?: string;
            model?: string;
            resolved_model?: string;
            key_id?: number;
            issue_type: string;
            reason?: string;
            action?: string;
        }>;
        route_preview_diffs?: DBRoutePreviewDiff[];
        replace_prune_preview?: DBReplacePrunePreview;
        replace_prune?: DBReplacePrunePreview;
        prune_preview?: DBReplacePrunePreview;
    };
    post_import_validation?: {
        summary?: {
            groups_scanned: number;
            candidates_scanned: number;
            degraded_groups: number;
            empty_groups: number;
            disabled_channels: number;
            channels_without_keys: number;
            stale_items_removed: number;
            route_warnings: number;
            price_rule_warnings: number;
            alias_mappings: number;
            alias_warnings: number;
        };
        degraded_groups?: string[];
        empty_groups?: string[];
        disabled_channels?: string[];
        channels_without_keys?: string[];
        stale_items_removed?: string[];
        route_warnings?: string[];
        price_rule_warnings?: string[];
        alias_mappings?: string[];
        alias_warnings?: string[];
        health_check?: {
            summary?: {
                target_groups: number;
                targets: number;
                passed: number;
                failed: number;
                skipped: number;
                rate_limited: number;
                connectivity_only: number;
            };
            checks?: Array<{
                group_name?: string;
                channel_id: number;
                channel_name?: string;
                model: string;
                passed: boolean;
                skipped?: boolean;
                rate_limited?: boolean;
                delay?: number;
                status_code?: number;
                error?: string;
                check_stage?: string;
            }>;
        };
    };
}

export type DBReplacePrunePreviewEntry = string | number | Record<string, unknown>;

export interface DBReplacePrunePreview {
    channels?: DBReplacePrunePreviewEntry[];
    channel_names?: DBReplacePrunePreviewEntry[];
    deleted_channels?: DBReplacePrunePreviewEntry[];
    pruned_channels?: DBReplacePrunePreviewEntry[];
    groups?: DBReplacePrunePreviewEntry[];
    group_names?: DBReplacePrunePreviewEntry[];
    deleted_groups?: DBReplacePrunePreviewEntry[];
    pruned_groups?: DBReplacePrunePreviewEntry[];
    settings?: DBReplacePrunePreviewEntry[];
    setting_keys?: DBReplacePrunePreviewEntry[];
    deleted_settings?: DBReplacePrunePreviewEntry[];
    pruned_settings?: DBReplacePrunePreviewEntry[];
    llm_infos?: DBReplacePrunePreviewEntry[];
    models?: DBReplacePrunePreviewEntry[];
    model_names?: DBReplacePrunePreviewEntry[];
    deleted_llm_infos?: DBReplacePrunePreviewEntry[];
    pruned_llm_infos?: DBReplacePrunePreviewEntry[];
    api_keys?: DBReplacePrunePreviewEntry[];
    api_key_names?: DBReplacePrunePreviewEntry[];
    deleted_api_keys?: DBReplacePrunePreviewEntry[];
    pruned_api_keys?: DBReplacePrunePreviewEntry[];
    warnings?: string[];
    preview_warnings?: string[];
}

export interface DBRoutePreviewCandidate {
    channel_name: string;
    model: string;
    resolved_model?: string;
    priority: number;
    weight: number;
    enabled: boolean;
    declared: boolean;
    has_key: boolean;
    key_id?: number;
    key_source_type?: string;
    key_remark?: string;
    reason?: string;
    billing_mode?: string;
    probe_policy?: string;
    probe_interval_seconds?: number;
    probe_concurrency_limit?: number;
    policy_basis?: string;
    billing_mode_basis?: string;
    probe_policy_basis?: string;
    probe_interval_basis?: string;
    probe_concurrency_basis?: string;
}

export interface DBRoutePreviewDiff {
    group_name: string;
    model: string;
    before_candidates?: DBRoutePreviewCandidate[];
    after_candidates?: DBRoutePreviewCandidate[];
    removed_candidates?: DBRoutePreviewCandidate[];
    added_candidates?: DBRoutePreviewCandidate[];
    skip_reasons?: string[];
    fallback_changed?: boolean;
}

export interface DBRollbackResult {
	snapshot_path?: string;
	snapshot_name?: string;
	imported_at?: string;
	applied_scopes?: DBImportScopes;
	result?: DBImportResult;
}

export interface DBImportSnapshotInfo {
	snapshot_path?: string;
	snapshot_name?: string;
	imported_at?: string;
	size_bytes?: number;
	is_latest?: boolean;
}

export interface DBRollbackPreviewResult {
	snapshot_path?: string;
	snapshot_name?: string;
	imported_at?: string;
	applied_scopes?: DBImportScopes;
	manifest?: {
		schema_version?: string;
		export_source?: string;
		checksum?: string;
		encrypted?: boolean;
		contains_secrets?: boolean;
	};
	compatibility?: DBImportResult['compatibility'];
	rows_summary?: Record<string, number>;
	preview_warnings?: string[];
}

export type DBImportMode = 'incremental' | 'map' | 'merge' | 'replace' | 'skip';

export interface DBImportScopes {
    routing: boolean;
    models: boolean;
    api_keys: boolean;
    settings: boolean;
    stats: boolean;
    logs: boolean;
}

export interface DBExportOptions {
    include_logs?: boolean;
    include_stats?: boolean;
    include_secrets?: boolean;
    format?: 'standard' | 'legacy';
}

type ApiResponse<T> = {
    code?: number;
    message?: string;
    data?: T;
};

function isRecord(value: unknown): value is Record<string, unknown> {
    return typeof value === 'object' && value !== null;
}

function getMessageField(value: unknown): string | undefined {
    if (!isRecord(value)) return undefined;
    const msg = value.message;
    return typeof msg === 'string' ? msg : undefined;
}

function getDataField<T>(value: unknown): T | undefined {
    if (!isRecord(value)) return undefined;
    return (value as ApiResponse<T>).data;
}

function buildAuthHeaders(): Headers {
	const token = useAuthStore.getState().token || getResolvedAuthToken();
	if (!token) throw new Error('Not authenticated');
	const headers = new Headers();
	headers.set('Authorization', `Bearer ${token}`);
	return headers;
}

function parseFilename(contentDisposition: string | null): string | null {
    if (!contentDisposition) return null;
    // e.g. attachment; filename="octopus-export-20250101120000.json"
    const match = contentDisposition.match(/filename="([^"]+)"/i);
    return match?.[1] ?? null;
}

function exportFallbackFilename() {
    const d = new Date();
    const pad = (n: number) => String(n).padStart(2, '0');
    const ts = `${d.getFullYear()}${pad(d.getMonth() + 1)}${pad(d.getDate())}${pad(d.getHours())}${pad(d.getMinutes())}${pad(d.getSeconds())}`;
    return `octopus-export-${ts}.json`;
}

async function downloadBlob(blob: Blob, filename: string) {
    const url = URL.createObjectURL(blob);
    try {
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        a.remove();
    } finally {
        URL.revokeObjectURL(url);
    }
}

/**
 * 鐎电厧鍤弫鐗堝祦鎼存搫绱欐稉瀣祰 JSON 閺傚洣娆㈤敍?
 */
export function useExportDB() {
    return useMutation({
        mutationFn: async (options: DBExportOptions = {}) => {
            const params = new URLSearchParams();
            params.set('include_logs', String(!!options.include_logs));
            params.set('include_stats', String(!!options.include_stats));
            params.set('include_secrets', String(options.include_secrets ?? true));
            params.set('format', options.format ?? 'standard');

			const res = await fetch(buildApiUrl('/api/v1/setting/export', {
				include_logs: !!options.include_logs,
				include_stats: !!options.include_stats,
				include_secrets: options.include_secrets ?? true,
				format: options.format ?? 'standard',
			}), {
				method: 'GET',
				headers: buildAuthHeaders(),
			});

            if (!res.ok) {
                const text = await res.text();
                throw new Error(text || res.statusText);
            }

            const blob = await res.blob();
            const filename = parseFilename(res.headers.get('content-disposition')) || exportFallbackFilename();
            await downloadBlob(blob, filename);
            return { filename };
        },
        onError: (error) => {
            logger.error('鐎电厧鍤弫鐗堝祦鎼存挸銇戠拹?', error);
        },
    });
}

/**
 * 鐎电厧鍙嗛弫鐗堝祦鎼存搫绱欐稉濠佺炊 JSON 閺傚洣娆㈤敍灞筋杻闁插繐顕遍崗銉礆
 */
export function useImportDB() {
    return useMutation({
        mutationFn: async ({
            file,
            dryRun = false,
            mode = 'incremental',
            modelMappings,
            importScopes,
            previewToken,
        }: {
            file: File;
            dryRun?: boolean;
            mode?: DBImportMode;
            modelMappings?: Record<string, string>;
            importScopes?: DBImportScopes;
            previewToken?: string;
        }) => {
            const form = new FormData();
            form.append('file', file);
            if (modelMappings && Object.keys(modelMappings).length > 0) {
                form.append('model_mappings', JSON.stringify(modelMappings));
            }
            if (importScopes) {
                form.append('import_scopes', JSON.stringify(importScopes));
            }
            if (previewToken) {
                form.append('preview_token', previewToken);
            }

			const res = await fetch(buildApiUrl('/api/v1/setting/import', {
				dry_run: dryRun,
				mode,
			}), {
				method: 'POST',
				headers: buildAuthHeaders(),
				body: form,
			});

            const contentType = res.headers.get('content-type') || '';
            const isJson = contentType.includes('application/json');
            const data = isJson ? await res.json() : await res.text();

            if (!res.ok) {
                const message = getMessageField(data) ?? (typeof data === 'string' ? data : res.statusText);
                throw new Error(message);
            }

            // 閺€顖涘瘮閸氬海顏弽鍥у櫙 ApiResponse閿涙code,message,data:{...}}
            const nested = getDataField<DBImportResult>(data);
            return nested ?? (data as DBImportResult);
        },
        onError: (error) => {
            logger.error('鐎电厧鍙嗛弫鐗堝祦鎼存挸銇戠拹?', error);
        },
    });
}

export function useRollbackLatestImportSnapshot() {
	return useMutation({
		mutationFn: async () => {
			return apiClient.post<DBRollbackResult>('/api/v1/setting/rollback-latest-import', {});
		},
		onError: (error) => {
			logger.error('閸ョ偞绮撮張鈧潻鎴滅濞嗏€愁嚤閸忋儱鎻╅悡褍銇戠拹?', error);
		},
	});
}

export function useImportSnapshots() {
	return useQuery({
		queryKey: ['settings', 'import-snapshots'],
		queryFn: async () => {
			return apiClient.get<DBImportSnapshotInfo[]>('/api/v1/setting/import-snapshots');
		},
	});
}

export function useRollbackImportSnapshot() {
	return useMutation({
		mutationFn: async ({ snapshotName, importScopes }: { snapshotName: string; importScopes?: DBImportScopes }) => {
			return apiClient.post<DBRollbackResult>('/api/v1/setting/rollback-import-snapshot', {
				snapshot_name: snapshotName,
				import_scopes: importScopes,
			});
		},
		onError: (error) => {
			logger.error('闁搞儳鍋炵划鎾箰閸パ呮毎閻庣數鍘ч崣鍡氱疀椤愩倕寮惧鎯扮簿鐟?', error);
		},
	});
}

export function usePreviewRollbackImportSnapshot() {
	return useMutation({
		mutationFn: async ({ snapshotName, importScopes }: { snapshotName: string; importScopes?: DBImportScopes }) => {
			return apiClient.post<DBRollbackPreviewResult>('/api/v1/setting/preview-rollback-import-snapshot', {
				snapshot_name: snapshotName,
				import_scopes: importScopes,
			});
		},
		onError: (error) => {
			logger.error('闁搞儳鍋炵划纾嬬疀椤愩倕寮惧Λ鏉垮椤秵寰勬潏顐バ?', error);
		},
	});
}

