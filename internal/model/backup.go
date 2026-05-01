package model

import (
	"strings"
	"time"
)

type DBDumpMigrationRecord struct {
	Version int `json:"version"`
	Status  int `json:"status"`
}

type DBImportMode string

const (
	DBImportModeIncremental DBImportMode = "incremental"
	DBImportModeMap         DBImportMode = "map"
	DBImportModeMerge       DBImportMode = "merge"
	DBImportModeReplace     DBImportMode = "replace"
	DBImportModeSkip        DBImportMode = "skip"
)

type DBImportOptions struct {
	ModelMappings         map[string]string `json:"model_mappings,omitempty"`
	ImportScopes          *DBImportScopes   `json:"import_scopes,omitempty"`
	SkipPreImportSnapshot bool              `json:"-"`
}

type DBImportScopes struct {
	Routing  bool `json:"routing,omitempty"`
	Models   bool `json:"models,omitempty"`
	APIKeys  bool `json:"api_keys,omitempty"`
	Settings bool `json:"settings,omitempty"`
	Stats    bool `json:"stats,omitempty"`
	Logs     bool `json:"logs,omitempty"`
}

type DBRollbackResult struct {
	SnapshotPath string          `json:"snapshot_path,omitempty"`
	SnapshotName string          `json:"snapshot_name,omitempty"`
	ImportedAt   time.Time       `json:"imported_at,omitempty"`
	AppliedScopes *DBImportScopes `json:"applied_scopes,omitempty"`
	Result       *DBImportResult `json:"result,omitempty"`
}

type DBImportSnapshotInfo struct {
	SnapshotPath string    `json:"snapshot_path,omitempty"`
	SnapshotName string    `json:"snapshot_name,omitempty"`
	ImportedAt   time.Time `json:"imported_at,omitempty"`
	SizeBytes    int64     `json:"size_bytes,omitempty"`
	IsLatest     bool      `json:"is_latest,omitempty"`
}

type DBRollbackPreviewResult struct {
	SnapshotPath    string                       `json:"snapshot_path,omitempty"`
	SnapshotName    string                       `json:"snapshot_name,omitempty"`
	ImportedAt      time.Time                    `json:"imported_at,omitempty"`
	AppliedScopes   *DBImportScopes              `json:"applied_scopes,omitempty"`
	Manifest        *DBDumpManifest              `json:"manifest,omitempty"`
	Compatibility   *DBImportCompatibilityReport `json:"compatibility,omitempty"`
	RowsSummary     map[string]int               `json:"rows_summary,omitempty"`
	PreviewWarnings []string                     `json:"preview_warnings,omitempty"`
}

func NormalizeDBImportMode(input string) DBImportMode {
	return DBImportMode(strings.ToLower(strings.TrimSpace(input)))
}

func DefaultDBImportMode(input string) DBImportMode {
	mode := NormalizeDBImportMode(input)
	if !IsValidDBImportMode(mode) {
		return DBImportModeIncremental
	}
	return mode
}

func IsValidDBImportMode(mode DBImportMode) bool {
	switch NormalizeDBImportMode(string(mode)) {
	case DBImportModeIncremental, DBImportModeMap, DBImportModeMerge, DBImportModeReplace, DBImportModeSkip:
		return true
	default:
		return false
	}
}

type DBDumpManifest struct {
	SchemaVersion   string `json:"schema_version"`
	ExportSource    string `json:"export_source"`
	Checksum        string `json:"checksum,omitempty"`
	Encrypted       bool   `json:"encrypted"`
	ContainsSecrets bool   `json:"contains_secrets"`
}

type DBDumpLegacyChannelHint struct {
	MissingKeyManagementMode bool `json:"-"`
	MissingKeyRoutingPolicy  bool `json:"-"`
}

type DBDumpLegacyChannelKeyHint struct {
	MissingSourceType    bool `json:"-"`
	MissingAllowedModels bool `json:"-"`
}

type DBDumpLegacyGroupHint struct {
	MissingRetryRounds       bool `json:"-"`
	MissingRetryDelayMs      bool `json:"-"`
	MissingFailoverWindowSec bool `json:"-"`
	MissingRaceAfterFails    bool `json:"-"`
	MissingRaceConcurrency   bool `json:"-"`
}

type DBDumpLegacyLLMInfoHint struct {
	MissingCanonicalName         bool `json:"-"`
	MissingBillingMode           bool `json:"-"`
	MissingProbePolicy           bool `json:"-"`
	MissingProbeIntervalSeconds  bool `json:"-"`
	MissingProbeConcurrencyLimit bool `json:"-"`
}

type DBDumpLegacyAPIKeyHint struct {
	MissingSupportedModels bool `json:"-"`
}

type DBDumpLegacyHints struct {
	Legacy                      bool                               `json:"-"`
	MissingManifest             bool                               `json:"-"`
	MissingUsers                bool                               `json:"-"`
	MissingRouteTargetOverrides bool                               `json:"-"`
	MissingMigrationRecords     bool                               `json:"-"`
	MissingRelayLogs            bool                               `json:"-"`
	ChannelsByName              map[string]DBDumpLegacyChannelHint `json:"-"`
	ChannelKeysBySnapshotID     map[int]DBDumpLegacyChannelKeyHint `json:"-"`
	GroupsByName                map[string]DBDumpLegacyGroupHint   `json:"-"`
	LLMInfosByName              map[string]DBDumpLegacyLLMInfoHint `json:"-"`
	APIKeysBySnapshotID         map[int]DBDumpLegacyAPIKeyHint     `json:"-"`
}

type DBImportCompatibilityReport struct {
	Summary                 *DBImportCompatibilitySummary    `json:"summary,omitempty"`
	AffectedGroups          []string                         `json:"affected_groups,omitempty"`
	AffectedChannels        []string                         `json:"affected_channels,omitempty"`
	MissingProviders        []string                         `json:"missing_providers,omitempty"`
	MissingModels           []string                         `json:"missing_models,omitempty"`
	Conflicts               []string                         `json:"conflicts,omitempty"`
	AliasConflicts          []string                         `json:"alias_conflicts,omitempty"`
	CredentialRebindTargets []DBImportCredentialRebindTarget `json:"credential_rebind_targets,omitempty"`
	ModelMappingPreviews    []DBImportModelMappingPreview    `json:"model_mapping_previews,omitempty"`
	AliasPreviewMappings    []DBImportAliasPreviewMapping    `json:"alias_preview_mappings,omitempty"`
	ModelPolicyDiffs        []DBImportModelPolicyDiff        `json:"model_policy_diffs,omitempty"`
	RouteConflicts          []string                         `json:"route_conflicts,omitempty"`
	BaseURLMismatches       []string                         `json:"base_url_mismatches,omitempty"`
	SchemaMismatches        []string                         `json:"schema_mismatches,omitempty"`
	SkippedTargets          []string                         `json:"skipped_targets,omitempty"`
	ReplacePrunedChannels   []string                         `json:"replace_pruned_channels,omitempty"`
	ReplacePrunedGroups     []string                         `json:"replace_pruned_groups,omitempty"`
	ReplacePrunedSettings   []string                         `json:"replace_pruned_settings,omitempty"`
	ReplacePrunedLLMInfos   []string                         `json:"replace_pruned_llm_infos,omitempty"`
	ReplacePrunedAPIKeys    []string                         `json:"replace_pruned_api_keys,omitempty"`
	ReplacePrunePreview     *DBReplacePrunePreview           `json:"replace_prune_preview,omitempty"`
	RoutePreviewWarnings    []string                         `json:"route_preview_warnings,omitempty"`
	InvalidRouteTargets     []DBImportRoutePreviewIssue      `json:"invalid_route_targets,omitempty"`
	SkippedRoutePreviews    []DBImportRoutePreviewIssue      `json:"skipped_route_target_previews,omitempty"`
	RoutePreviewDiffs       []DBImportRoutePreviewDiff       `json:"route_preview_diffs,omitempty"`
}

type DBImportCompatibilitySummary struct {
	MissingProviders        int `json:"missing_providers"`
	MissingModels           int `json:"missing_models"`
	Conflicts               int `json:"conflicts"`
	AliasConflicts          int `json:"alias_conflicts"`
	CredentialRebindTargets int `json:"credential_rebind_targets"`
	ChannelKeyRebindTargets int `json:"channel_key_rebind_targets"`
	APIKeyRebindTargets     int `json:"api_key_rebind_targets"`
	ModelMappingPreviews    int `json:"model_mapping_previews"`
	UsedModelMappings       int `json:"used_model_mappings"`
	UnusedModelMappings     int `json:"unused_model_mappings"`
	MissingMappingTargets   int `json:"missing_mapping_targets"`
	AliasPreviewMaps        int `json:"alias_preview_mappings"`
	ModelPolicyDiffs        int `json:"model_policy_diffs"`
	RouteConflicts          int `json:"route_conflicts"`
	InvalidRouteTargets     int `json:"invalid_route_targets"`
	SkippedRoutePreviews    int `json:"skipped_route_target_previews"`
	RoutePreviewDiffs       int `json:"route_preview_diffs"`
	BaseURLMismatches       int `json:"base_url_mismatches"`
	SchemaMismatches        int `json:"schema_mismatches"`
	SkippedTargets          int `json:"skipped_targets"`
	ReplacePrunedChannels   int `json:"replace_pruned_channels"`
	ReplacePrunedGroups     int `json:"replace_pruned_groups"`
	ReplacePrunedSettings   int `json:"replace_pruned_settings"`
	ReplacePrunedLLMInfos   int `json:"replace_pruned_llm_infos"`
	ReplacePrunedAPIKeys    int `json:"replace_pruned_api_keys"`
}

type DBReplacePrunePreview struct {
	Channels []string `json:"channels,omitempty"`
	Groups   []string `json:"groups,omitempty"`
	Settings []string `json:"settings,omitempty"`
	LLMInfos []string `json:"llm_infos,omitempty"`
	APIKeys  []string `json:"api_keys,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

type DBImportCredentialRebindTarget struct {
	TargetType     string   `json:"target_type"`
	SnapshotID     int      `json:"snapshot_id,omitempty"`
	ChannelName    string   `json:"channel_name,omitempty"`
	KeyName        string   `json:"key_name,omitempty"`
	SourceType     string   `json:"source_type,omitempty"`
	Remark         string   `json:"remark,omitempty"`
	Models         []string `json:"models,omitempty"`
	AffectedGroups []string `json:"affected_groups,omitempty"`
	Contexts       []string `json:"contexts,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
}

type DBImportModelMappingPreview struct {
	SourceModel   string   `json:"source_model"`
	TargetModel   string   `json:"target_model"`
	Contexts      []string `json:"contexts,omitempty"`
	TouchedFields []string `json:"touched_fields,omitempty"`
	UsageCount    int      `json:"usage_count,omitempty"`
	Used          bool     `json:"used,omitempty"`
	TargetExists  bool     `json:"target_exists,omitempty"`
	Warnings      []string `json:"warnings,omitempty"`
}

type DBImportAliasPreviewMapping struct {
	SnapshotModel string   `json:"snapshot_model"`
	CurrentModel  string   `json:"current_model"`
	Canonical     string   `json:"canonical,omitempty"`
	Contexts      []string `json:"contexts,omitempty"`
}

type DBImportModelPolicyState struct {
	BillingMode      string `json:"billing_mode,omitempty"`
	ProbePolicy      string `json:"probe_policy,omitempty"`
	ProbeInterval    int    `json:"probe_interval,omitempty"`
	ProbeConcurrency int    `json:"probe_concurrency,omitempty"`
}

type DBImportModelPolicyDiff struct {
	Model         string                   `json:"model"`
	CurrentModel  string                   `json:"current_model,omitempty"`
	Canonical     string                   `json:"canonical,omitempty"`
	Before        DBImportModelPolicyState `json:"before,omitempty"`
	After         DBImportModelPolicyState `json:"after,omitempty"`
	ChangedFields []string                 `json:"changed_fields,omitempty"`
	ImpactLevel   string                   `json:"impact_level,omitempty"`
	Warnings      []string                 `json:"warnings,omitempty"`
	Contexts      []string                 `json:"contexts,omitempty"`
	SkipReasons   []string                 `json:"skip_reasons,omitempty"`
}

type DBImportRoutePreviewDiff struct {
	GroupName         string                          `json:"group_name"`
	Model             string                          `json:"model"`
	BeforeCandidates  []DBImportRoutePreviewCandidate `json:"before_candidates,omitempty"`
	AfterCandidates   []DBImportRoutePreviewCandidate `json:"after_candidates,omitempty"`
	RemovedCandidates []DBImportRoutePreviewCandidate `json:"removed_candidates,omitempty"`
	AddedCandidates   []DBImportRoutePreviewCandidate `json:"added_candidates,omitempty"`
	SkipReasons       []string                        `json:"skip_reasons,omitempty"`
	FallbackChanged   bool                            `json:"fallback_changed,omitempty"`
}

type DBImportRoutePreviewIssue struct {
	GroupName     string `json:"group_name,omitempty"`
	ChannelName   string `json:"channel_name,omitempty"`
	Model         string `json:"model,omitempty"`
	ResolvedModel string `json:"resolved_model,omitempty"`
	KeyID         int    `json:"key_id,omitempty"`
	IssueType     string `json:"issue_type"`
	Reason        string `json:"reason,omitempty"`
	Action        string `json:"action,omitempty"`
}

type DBImportRoutePreviewCandidate struct {
	ChannelName           string `json:"channel_name"`
	Model                 string `json:"model"`
	ResolvedModel         string `json:"resolved_model,omitempty"`
	Priority              int    `json:"priority"`
	Weight                int    `json:"weight"`
	Enabled               bool   `json:"enabled"`
	Declared              bool   `json:"declared"`
	HasKey                bool   `json:"has_key"`
	KeyID                 int    `json:"key_id,omitempty"`
	KeySourceType         string `json:"key_source_type,omitempty"`
	KeyRemark             string `json:"key_remark,omitempty"`
	BillingMode           string `json:"billing_mode,omitempty"`
	ProbePolicy           string `json:"probe_policy,omitempty"`
	ProbeIntervalSeconds  int    `json:"probe_interval_seconds,omitempty"`
	ProbeConcurrencyLimit int    `json:"probe_concurrency_limit,omitempty"`
	BillingModeBasis      string `json:"billing_mode_basis,omitempty"`
	ProbePolicyBasis      string `json:"probe_policy_basis,omitempty"`
	ProbeIntervalBasis    string `json:"probe_interval_basis,omitempty"`
	ProbeConcurrencyBasis string `json:"probe_concurrency_basis,omitempty"`
	PolicyBasis           string `json:"policy_basis,omitempty"`
	Reason                string `json:"reason,omitempty"`
}

type DBDump struct {
	Version      int            `json:"version"`
	ExportedAt   time.Time      `json:"exported_at"`
	IncludeLogs  bool           `json:"include_logs"`
	IncludeStats bool           `json:"include_stats"`
	Manifest     DBDumpManifest `json:"manifest"`
	LegacyHints  *DBDumpLegacyHints `json:"-"`

	Channels             []Channel               `json:"channels,omitempty"`
	Users                []User                  `json:"users,omitempty"`
	ChannelKeys          []ChannelKey            `json:"channel_keys,omitempty"`
	RouteTargetOverrides []RouteTargetOverride   `json:"route_target_overrides,omitempty"`
	Groups               []Group                 `json:"groups,omitempty"`
	GroupItems           []GroupItem             `json:"group_items,omitempty"`
	LLMInfos             []LLMInfo               `json:"llm_infos,omitempty"`
	APIKeys              []APIKey                `json:"api_keys,omitempty"`
	Settings             []Setting               `json:"settings,omitempty"`
	MigrationRecords     []DBDumpMigrationRecord `json:"migration_records,omitempty"`

	StatsTotal   []StatsTotal   `json:"stats_total,omitempty"`
	StatsDaily   []StatsDaily   `json:"stats_daily,omitempty"`
	StatsHourly  []StatsHourly  `json:"stats_hourly,omitempty"`
	StatsModel   []StatsModel   `json:"stats_model,omitempty"`
	StatsChannel []StatsChannel `json:"stats_channel,omitempty"`
	StatsAPIKey  []StatsAPIKey  `json:"stats_api_key,omitempty"`

	RelayLogs []RelayLog `json:"relay_logs,omitempty"`
}

type DBImportResult struct {
	RowsAffected         map[string]int64              `json:"rows_affected"`
	Warnings             []string                      `json:"warnings,omitempty"`
	DryRun               bool                          `json:"dry_run,omitempty"`
	Mode                 DBImportMode                  `json:"mode,omitempty"`
	PreviewToken         string                        `json:"preview_token,omitempty"`
	Manifest             *DBDumpManifest               `json:"manifest,omitempty"`
	Compatibility        *DBImportCompatibilityReport  `json:"compatibility,omitempty"`
	PostImportValidation *DBImportPostValidationReport `json:"post_import_validation,omitempty"`
}

type DBImportPostValidationReport struct {
	Summary             *DBImportPostValidationSummary `json:"summary,omitempty"`
	DegradedGroups      []string                       `json:"degraded_groups,omitempty"`
	EmptyGroups         []string                       `json:"empty_groups,omitempty"`
	DisabledChannels    []string                       `json:"disabled_channels,omitempty"`
	ChannelsWithoutKeys []string                       `json:"channels_without_keys,omitempty"`
	StaleItemsRemoved   []string                       `json:"stale_items_removed,omitempty"`
	RouteWarnings       []string                       `json:"route_warnings,omitempty"`
	PriceRuleWarnings   []string                       `json:"price_rule_warnings,omitempty"`
	AliasMappings       []string                       `json:"alias_mappings,omitempty"`
	AliasWarnings       []string                       `json:"alias_warnings,omitempty"`
	HealthCheck         *DBImportHealthCheckReport     `json:"health_check,omitempty"`
}

type DBImportPostValidationSummary struct {
	GroupsScanned       int `json:"groups_scanned"`
	CandidatesScanned   int `json:"candidates_scanned"`
	DegradedGroups      int `json:"degraded_groups"`
	EmptyGroups         int `json:"empty_groups"`
	DisabledChannels    int `json:"disabled_channels"`
	ChannelsWithoutKeys int `json:"channels_without_keys"`
	StaleItemsRemoved   int `json:"stale_items_removed"`
	RouteWarnings       int `json:"route_warnings"`
	PriceRuleWarnings   int `json:"price_rule_warnings"`
	AliasMappings       int `json:"alias_mappings"`
	AliasWarnings       int `json:"alias_warnings"`
}

type DBImportHealthCheckReport struct {
	Summary *DBImportHealthCheckSummary `json:"summary,omitempty"`
	Checks  []DBImportHealthCheckItem   `json:"checks,omitempty"`
}

type DBImportHealthCheckSummary struct {
	TargetGroups     int `json:"target_groups"`
	Targets          int `json:"targets"`
	Passed           int `json:"passed"`
	Failed           int `json:"failed"`
	Skipped          int `json:"skipped"`
	RateLimited      int `json:"rate_limited"`
	ConnectivityOnly int `json:"connectivity_only"`
}

type DBImportHealthCheckItem struct {
	GroupName   string `json:"group_name,omitempty"`
	ChannelID   int    `json:"channel_id"`
	ChannelName string `json:"channel_name,omitempty"`
	Model       string `json:"model"`
	SourceType  string `json:"source_type,omitempty"`
	BillingMode string `json:"billing_mode,omitempty"`
	ProbePolicy string `json:"probe_policy,omitempty"`
	PolicyBasis string `json:"policy_basis,omitempty"`
	Passed      bool   `json:"passed"`
	Skipped     bool   `json:"skipped,omitempty"`
	RateLimited bool   `json:"rate_limited,omitempty"`
	Delay       int    `json:"delay,omitempty"`
	StatusCode  int    `json:"status_code,omitempty"`
	Error       string `json:"error,omitempty"`
	CheckStage  string `json:"check_stage,omitempty"`
}
