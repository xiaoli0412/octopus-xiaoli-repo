package model

import "time"

const (
	GovernanceSessionStatusDraft    = "draft"
	GovernanceSessionStatusPlanning = "planning"
	GovernanceSessionStatusReady    = "ready"
	GovernanceSessionStatusStale    = "stale"
	GovernanceSessionStatusApplying = "applying"
	GovernanceSessionStatusApplied  = "applied"
	GovernanceSessionStatusFailed   = "failed"
)

const (
	GovernanceStageIntentIntake    = "intent_intake"
	GovernanceStageContextAssembly = "context_assembly"
	GovernanceStagePlanGeneration  = "plan_generation"
	GovernanceStagePreviewCompile  = "preview_compilation"
	GovernanceStageGuardValidation = "guard_validation"
	GovernanceStageApplyExecution  = "apply_execution"
	GovernanceStageCompleted       = "completed"
)

const (
	GovernanceApplyRunStatusPending    = "pending"
	GovernanceApplyRunStatusValidating = "validating"
	GovernanceApplyRunStatusRunning    = "running"
	GovernanceApplyRunStatusSucceeded  = "succeeded"
	GovernanceApplyRunStatusFailed     = "failed"
	GovernanceApplyRunStatusRolledBack = "rolled_back"
)

const (
	StrategyProfileStatusDraft    = "draft"
	StrategyProfileStatusReady    = "ready"
	StrategyProfileStatusActive   = "active"
	StrategyProfileStatusImported = "imported"
)

const (
	GovernanceScopeRoutingGrouping = "routing_grouping"
)

const (
	GovernanceMutationTypeGroupUpsert               = "group_upsert"
	GovernanceMutationTypeGroupItemAttach           = "group_item_attach"
	GovernanceMutationTypeGroupItemDetach           = "group_item_detach"
	GovernanceMutationTypeGroupItemReorder          = "group_item_reorder"
	GovernanceMutationTypeRouteTargetOverrideUpsert = "route_target_override_upsert"
	GovernanceMutationTypeRouteTargetOverrideDelete = "route_target_override_delete"
	GovernanceMutationTypeLLMPriceUpsert            = "llm_price_upsert"
	GovernanceMutationTypeDynamicRoutingSettingSet  = "dynamic_routing_setting_set"
	GovernanceMutationTypeRuntimePolicySet          = "runtime_policy_set"
	GovernanceMutationTypeStrategyProfileActivate   = "strategy_profile_activate"
)

const (
	GovernanceExpertPresetConservative = "conservative"
	GovernanceExpertPresetBalanced     = "balanced"
	GovernanceExpertPresetDeepReview   = "deep_review"
)

type GovernanceSession struct {
	ID               int        `json:"id" gorm:"primaryKey"`
	Goal             string     `json:"goal" gorm:"type:text;not null"`
	Scope            string     `json:"scope" gorm:"not null;index"`
	ExpertPresetID   string     `json:"expert_preset_id" gorm:"not null;index"`
	Status           string     `json:"status" gorm:"not null;index"`
	CurrentStage     string     `json:"current_stage" gorm:"not null;index"`
	OperatorSummary  string     `json:"operator_summary" gorm:"type:text"`
	RiskSummary      string     `json:"risk_summary" gorm:"type:text"`
	Confidence       float64    `json:"confidence"`
	SnapshotChecksum string     `json:"snapshot_checksum" gorm:"size:128;index"`
	SnapshotJSON     string     `json:"snapshot_json,omitempty" gorm:"type:text"`
	PlanJSON         string     `json:"plan_json,omitempty" gorm:"type:text"`
	PreviewJSON      string     `json:"preview_json,omitempty" gorm:"type:text"`
	ErrorMessage     string     `json:"error_message,omitempty" gorm:"type:text"`
	LastApplyRunID   *int       `json:"last_apply_run_id,omitempty" gorm:"index"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	AppliedAt        *time.Time `json:"applied_at,omitempty"`
}

type GovernanceApplyRun struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	SessionID     int       `json:"session_id" gorm:"not null;index"`
	Status        string    `json:"status" gorm:"not null;index"`
	ResultSummary string    `json:"result_summary" gorm:"type:text"`
	AuditJSON     string    `json:"audit_json,omitempty" gorm:"type:text"`
	ErrorMessage  string    `json:"error_message,omitempty" gorm:"type:text"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type GovernanceRollbackPoint struct {
	ID               int       `json:"id" gorm:"primaryKey"`
	SessionID        int       `json:"session_id" gorm:"not null;index"`
	ApplyRunID       *int      `json:"apply_run_id,omitempty" gorm:"index"`
	SnapshotChecksum string    `json:"snapshot_checksum" gorm:"size:128;index"`
	SnapshotJSON     string    `json:"snapshot_json,omitempty" gorm:"type:text"`
	Summary          string    `json:"summary" gorm:"type:text"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type StrategyProfile struct {
	ID                int        `json:"id" gorm:"primaryKey"`
	Name              string     `json:"name" gorm:"not null;index"`
	Summary           string     `json:"summary" gorm:"type:text"`
	Status            string     `json:"status" gorm:"not null;index"`
	SourceSessionID   *int       `json:"source_session_id,omitempty" gorm:"index"`
	LegacyAIProfileID *int       `json:"legacy_ai_profile_id,omitempty" gorm:"index"`
	MutationsJSON     string     `json:"mutations_json,omitempty" gorm:"type:text"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	ActivatedAt       *time.Time `json:"activated_at,omitempty"`
}

type GovernanceFindingView struct {
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

type GovernanceDecisionView struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

type GovernanceGroupUpsertMutation struct {
	GroupName string    `json:"group_name"`
	Mode      GroupMode `json:"mode"`
}

type GovernanceGroupItemMutation struct {
	GroupName  string `json:"group_name"`
	ChannelID  int    `json:"channel_id"`
	ModelName  string `json:"model_name"`
	Priority   int    `json:"priority,omitempty"`
	Weight     int    `json:"weight,omitempty"`
	ChannelKey int    `json:"channel_key_id,omitempty"`
}

type GovernanceGroupItemReorderMutation struct {
	GroupName string                        `json:"group_name"`
	Items     []GovernanceGroupItemMutation `json:"items"`
}

type GovernanceRouteTargetOverrideMutation struct {
	ChannelID             int         `json:"channel_id"`
	ChannelKeyID          int         `json:"channel_key_id"`
	ModelName             string      `json:"model_name"`
	BillingMode           BillingMode `json:"billing_mode,omitempty"`
	ProbePolicy           ProbePolicy `json:"probe_policy,omitempty"`
	ProbeIntervalSeconds  int         `json:"probe_interval_seconds,omitempty"`
	ProbeConcurrencyLimit int         `json:"probe_concurrency_limit,omitempty"`
}

type GovernanceStrategyProfileActivateMutation struct {
	StrategyProfileID int `json:"strategy_profile_id"`
}

type GovernanceLLMPriceUpsertMutation struct {
	Name                  string      `json:"name"`
	CanonicalName         string      `json:"canonical_name,omitempty"`
	Input                 float64     `json:"input"`
	Output                float64     `json:"output"`
	CacheRead             float64     `json:"cache_read"`
	CacheWrite            float64     `json:"cache_write"`
	OfficialInput         float64     `json:"official_input"`
	OfficialOutput        float64     `json:"official_output"`
	OfficialCacheRead     float64     `json:"official_cache_read"`
	OfficialCacheWrite    float64     `json:"official_cache_write"`
	BillingMode           BillingMode `json:"billing_mode,omitempty"`
	ProbePolicy           ProbePolicy `json:"probe_policy,omitempty"`
	ProbeIntervalSeconds  int         `json:"probe_interval_seconds,omitempty"`
	ProbeConcurrencyLimit int         `json:"probe_concurrency_limit,omitempty"`
	Source                string      `json:"source,omitempty"`
}

type GovernanceSettingMutation struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GovernanceRuntimePolicyView struct {
	Strategy                string `json:"strategy"`
	DispatchMode            string `json:"dispatch_mode"`
	MaxParallelRuns         int    `json:"max_parallel_runs"`
	DoubleReviewEnabled     bool   `json:"double_review_enabled"`
	FallbackToDeterministic bool   `json:"fallback_to_deterministic"`
	DegradedToDeterministic bool   `json:"degraded_to_deterministic"`
	Label                   string `json:"label,omitempty"`
}

type GovernanceMutation struct {
	Type                    string                                     `json:"type"`
	Summary                 string                                     `json:"summary"`
	GroupUpsert             *GovernanceGroupUpsertMutation             `json:"group_upsert,omitempty"`
	GroupItemAttach         *GovernanceGroupItemMutation               `json:"group_item_attach,omitempty"`
	GroupItemDetach         *GovernanceGroupItemMutation               `json:"group_item_detach,omitempty"`
	GroupItemReorder        *GovernanceGroupItemReorderMutation        `json:"group_item_reorder,omitempty"`
	RouteTargetUpsert       *GovernanceRouteTargetOverrideMutation     `json:"route_target_override_upsert,omitempty"`
	RouteTargetDelete       *GovernanceRouteTargetOverrideMutation     `json:"route_target_override_delete,omitempty"`
	LLMPriceUpsert          *GovernanceLLMPriceUpsertMutation          `json:"llm_price_upsert,omitempty"`
	DynamicRoutingSettingSet *GovernanceSettingMutation                `json:"dynamic_routing_setting_set,omitempty"`
	RuntimePolicySet        *GovernanceRuntimePolicyView               `json:"runtime_policy_set,omitempty"`
	StrategyProfileActivate *GovernanceStrategyProfileActivateMutation `json:"strategy_profile_activate,omitempty"`
}

type GovernanceDomainPlanView struct {
	Key           string                  `json:"key"`
	Title         string                  `json:"title"`
	Summary       string                  `json:"summary"`
	Status        string                  `json:"status"`
	FindingCount  int                     `json:"finding_count"`
	MutationCount int                     `json:"mutation_count"`
	Findings      []GovernanceFindingView `json:"findings,omitempty"`
	Mutations     []GovernanceMutation    `json:"mutations,omitempty"`
}

type GovernancePlanView struct {
	Findings        []GovernanceFindingView  `json:"findings"`
	Decisions       []GovernanceDecisionView `json:"decisions"`
	Mutations       []GovernanceMutation     `json:"mutations"`
	Domains         []GovernanceDomainPlanView `json:"domains,omitempty"`
	RiskSummary     string                   `json:"risk_summary"`
	Confidence      float64                  `json:"confidence"`
	OperatorSummary string                   `json:"operator_summary"`
}

type GovernancePreviewImpactCounts struct {
	Groups    int `json:"groups"`
	Items     int `json:"items"`
	Overrides int `json:"overrides"`
	Profiles  int `json:"profiles"`
}

type GovernancePreviewView struct {
	Headline      string                        `json:"headline"`
	SummaryLines  []string                      `json:"summary_lines"`
	ImpactCounts  GovernancePreviewImpactCounts `json:"impact_counts"`
	RiskNotes     []string                      `json:"risk_notes"`
	ApplyBlockers []string                      `json:"apply_blockers"`
	CanApply      bool                          `json:"can_apply"`
	MutationCount int                           `json:"mutation_count"`
	Mutations     []GovernanceMutation          `json:"mutations"`
}

type GovernanceApplyAuditItem struct {
	MutationType string `json:"mutation_type"`
	Summary      string `json:"summary"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

type GovernanceApplyAudit struct {
	Summary string                     `json:"summary"`
	Items   []GovernanceApplyAuditItem `json:"items"`
}

type AIGovernanceExecutionSource struct {
	Mode            string `json:"mode"`
	BaseURL         string `json:"base_url"`
	ChannelType     string `json:"channel_type"`
	Model           string `json:"model"`
	UseLocalDefault bool   `json:"use_local_default"`
	Label           string `json:"label"`
}

type AIGovernanceLearningSummary struct {
	Enabled      bool    `json:"enabled"`
	SampleCount  int     `json:"sample_count"`
	TopTarget    string  `json:"top_target,omitempty"`
	LastSampleAt *int64  `json:"last_sample_at,omitempty"`
	TopScore     float64 `json:"top_score,omitempty"`
}

type GovernanceSessionSummary struct {
	ID              int        `json:"id"`
	Goal            string     `json:"goal"`
	Scope           string     `json:"scope"`
	ExpertPresetID  string     `json:"expert_preset_id"`
	Status          string     `json:"status"`
	CurrentStage    string     `json:"current_stage"`
	OperatorSummary string     `json:"operator_summary"`
	RiskSummary     string     `json:"risk_summary"`
	Confidence      float64    `json:"confidence"`
	MutationCount   int        `json:"mutation_count"`
	CanApply        bool       `json:"can_apply"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	AppliedAt       *time.Time `json:"applied_at,omitempty"`
}

type GovernanceSessionDetail struct {
	GovernanceSessionSummary
	Plan             GovernancePlanView        `json:"plan"`
	Preview          GovernancePreviewView     `json:"preview"`
	SnapshotChecksum string                    `json:"snapshot_checksum"`
	ApplyRuns        []GovernanceApplyRunView  `json:"apply_runs,omitempty"`
	RollbackPoints   []GovernanceRollbackPointView `json:"rollback_points,omitempty"`
	SnapshotSummary  GovernanceSnapshotSummary `json:"snapshot_summary"`
}

type GovernanceApplyRunView struct {
	ID            int                  `json:"id"`
	SessionID     int                  `json:"session_id"`
	Status        string               `json:"status"`
	ResultSummary string               `json:"result_summary"`
	ErrorMessage  string               `json:"error_message,omitempty"`
	Audit         GovernanceApplyAudit `json:"audit"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

type GovernanceRollbackPointView struct {
	ID               int       `json:"id"`
	SessionID        int       `json:"session_id"`
	ApplyRunID       *int      `json:"apply_run_id,omitempty"`
	SnapshotChecksum string    `json:"snapshot_checksum"`
	Summary          string    `json:"summary"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type GovernanceSnapshotSummary struct {
	Channels             int      `json:"channels"`
	EnabledChannels      int      `json:"enabled_channels"`
	Groups               int      `json:"groups"`
	GroupItems           int      `json:"group_items"`
	RouteTargetOverrides int      `json:"route_target_overrides"`
	Models               int      `json:"models"`
	MissingPrices        int      `json:"missing_prices"`
	ActiveSourceMode     string   `json:"active_source_mode"`
	ActiveSourceLabel    string   `json:"active_source_label"`
	Highlights           []string `json:"highlights"`
}

type StrategyProfileSummary struct {
	ID              int        `json:"id"`
	Name            string     `json:"name"`
	Summary         string     `json:"summary"`
	Status          string     `json:"status"`
	SourceSessionID *int       `json:"source_session_id,omitempty"`
	ActivatedAt     *time.Time `json:"activated_at,omitempty"`
	IsActive        bool       `json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ExpertPresetView struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	ReviewDepth        string `json:"review_depth"`
	CreateManagedGroup bool   `json:"create_managed_group"`
	SyncBindings       bool   `json:"sync_bindings"`
	CleanupStale       bool   `json:"cleanup_stale"`
}

type AIGovernanceOverview struct {
	Enabled               bool                        `json:"enabled"`
	ExecutionSource       AIGovernanceExecutionSource `json:"execution_source"`
	RuntimePolicy         GovernanceRuntimePolicyView `json:"runtime_policy"`
	ManagedGroupName      string                      `json:"managed_group_name"`
	Learning              AIGovernanceLearningSummary `json:"learning"`
	ActiveStrategyProfile *StrategyProfileSummary     `json:"active_strategy_profile,omitempty"`
	RecentSession         *GovernanceSessionSummary   `json:"recent_session,omitempty"`
}

type GovernanceSessionCreateRequest struct {
	Goal           string `json:"goal" binding:"required"`
	ExpertPresetID string `json:"expert_preset_id,omitempty"`
}

type GovernanceStrategyProfileCreateRequest struct {
	SessionID int    `json:"session_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type GovernanceRollbackRequest struct {
	RollbackPointID int `json:"rollback_point_id,omitempty"`
}
