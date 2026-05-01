package model

import (
	"strings"
	"time"
)

const (
	AIAutomationTaskTypeNaturalLanguage      = "natural_language"
	AIAutomationTaskTypeGroupSuggestion      = "group_suggestion"
	AIAutomationTaskTypeChannelRecognition   = "channel_recognition"
	AIAutomationTaskTypePriceRecognition     = "price_recognition"
	AIAutomationTaskTypeModelClassification  = "model_classification"
	AIAutomationTaskTypeConfigHealthCheck    = "config_health_check"
	AIAutomationTaskTypeDynamicRoutingDigest = "dynamic_routing_digest"
)

const (
	AITaskStatusPending             = "pending"
	AITaskStatusRunning             = "running"
	AITaskStatusRecoverable         = "recoverable"
	AITaskStatusSucceeded           = "succeeded"
	AITaskStatusFailed              = "failed"
	AITaskStatusFailedUnrecoverable = "failed_unrecoverable"
	AITaskStatusCanceled            = "canceled"
)

const (
	AITaskResumeStateCollectContext  = "ready_to_collect_context"
	AITaskResumeStateSelectModel     = "ready_to_select_model"
	AITaskResumeStateCallAI          = "ready_to_call_ai"
	AITaskResumeStateParse           = "ready_to_parse"
	AITaskResumeStateGenerateProfile = "ready_to_generate_profile"
	AITaskResumeStateSaveResult      = "ready_to_save_result"
	AITaskResumeStateCompleted       = "completed"
)

const (
	AITaskStepStatusPending   = "pending"
	AITaskStepStatusRunning   = "running"
	AITaskStepStatusSucceeded = "succeeded"
	AITaskStepStatusFailed    = "failed"
	AITaskStepStatusSkipped   = "skipped"
)

const (
	AIPromptTemplateSourceBuiltin = "builtin"
	AIPromptTemplateSourceCustom  = "custom"
)

const (
	AIProfileDomainGrouping             = "grouping"
	AIProfileDomainChannelRecognition   = "channel_recognition"
	AIProfileDomainPriceRecognition     = "price_recognition"
	AIProfileDomainModelClassification  = "model_classification"
	AIProfileDomainConfigHealthCheck    = "config_health_check"
	AIProfileDomainDynamicRoutingDigest = "dynamic_routing_digest"
)

const (
	AIProfileStatusDraft    = "draft"
	AIProfileStatusReady    = "ready"
	AIProfileStatusActive   = "active"
	AIProfileStatusArchived = "archived"
	AIProfileStatusInvalid  = "invalid"
)

const (
	AIProfileMigrationStatusTypedBackfilled = "typed_backfilled"
	AIProfileMigrationStatusLegacyOnly      = "legacy_only"
)

const (
	AIAutomationDefaultSelectionPolicy     = "free_success_latency"
	AIAutomationModelSourceLocalCache      = "local_model_cache"
	AIAutomationModelSourceRemoteDiscovery = "remote_model_discovery"
)

const (
	AITaskToolKeyChannelInventory = "channel_inventory"
	AITaskToolKeyGroupTopology    = "group_topology"
	AITaskToolKeyPriceCatalog     = "price_catalog"
	AITaskToolKeyModelCatalog     = "model_catalog"
	AITaskToolKeyRouteOverrides   = "route_overrides"
	AITaskToolKeyLearningState    = "learning_state"
	AITaskToolKeyProfileWrite     = "profile_write"
	AITaskToolKeyProfileActivate  = "profile_activate"
	AITaskToolKeySnapshotGuard    = "snapshot_guard"
)

type AIAutomationConfig struct {
	Enabled                       bool                     `json:"enabled"`
	BaseURL                       string                   `json:"base_url"`
	APIKey                        string                   `json:"api_key,omitempty"`
	ChannelType                   string                   `json:"channel_type"`
	Model                         string                   `json:"model"`
	UseLocalDefault               bool                     `json:"use_local_default"`
	DefaultSelectionPolicy        string                   `json:"default_selection_policy"`
	RequestedConfigSourceMode     string                   `json:"requested_config_source_mode"`
	ConfigSourceMode              string                   `json:"config_source_mode"`
	RequestedActiveAIProfileID    int                      `json:"requested_active_ai_profile_id"`
	ActiveAIProfileID             int                      `json:"active_ai_profile_id"`
	RequestedActiveAIProfile      *AIAutomationProfileRef  `json:"requested_active_ai_profile,omitempty"`
	ActiveAIProfile               *AIAutomationProfileRef  `json:"active_ai_profile,omitempty"`
	SourceFallbackReason          string                   `json:"source_fallback_reason,omitempty"`
	DynamicRoutingLearningEnabled bool                     `json:"dynamic_routing_learning_enabled"`
	ManualConfig                  AIAutomationConfigValues `json:"manual_config"`
	EffectiveConfig               AIAutomationConfigValues `json:"effective_config"`
}

type AIAutomationConfigValues struct {
	BaseURL         string `json:"base_url"`
	APIKey          string `json:"api_key,omitempty"`
	ChannelType     string `json:"channel_type"`
	Model           string `json:"model"`
	UseLocalDefault bool   `json:"use_local_default"`
}

type AIAutomationProfileRef struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Domain      string    `json:"domain"`
	Version     int       `json:"version"`
	Status      string    `json:"status"`
	Confidence  float64   `json:"confidence"`
	Explanation string    `json:"explanation,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AIAutomationConfigUpdateRequest struct {
	Enabled         *bool   `json:"enabled,omitempty"`
	BaseURL         *string `json:"base_url,omitempty"`
	APIKey          *string `json:"api_key,omitempty"`
	ChannelType     *string `json:"channel_type,omitempty"`
	Model           *string `json:"model,omitempty"`
	UseLocalDefault *bool   `json:"use_local_default,omitempty"`
}

type AIModelCandidate struct {
	Name         string  `json:"name"`
	Source       string  `json:"source"`
	ChannelID    int     `json:"channel_id,omitempty"`
	ChannelName  string  `json:"channel_name,omitempty"`
	Available    bool    `json:"available"`
	FreeLikely   bool    `json:"free_likely"`
	SuccessRate  float64 `json:"success_rate"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	Recommended  bool    `json:"recommended"`
	Reason       string  `json:"reason"`
}

type AIModelsFetchRequest struct {
	BaseURL         string `json:"base_url"`
	APIKey          string `json:"api_key"`
	ChannelType     string `json:"channel_type"`
	UseLocalDefault *bool  `json:"use_local_default,omitempty"`
}

type AIModelsFetchResult struct {
	Source       string             `json:"source"`
	Candidates   []AIModelCandidate `json:"candidates"`
	SelectedName string             `json:"selected_name"`
	Policy       string             `json:"policy"`
}

type AIAutomationTaskConfig struct {
	BaseURL         string   `json:"base_url,omitempty"`
	APIKey          string   `json:"api_key,omitempty"`
	ChannelType     string   `json:"channel_type,omitempty"`
	Model           string   `json:"model,omitempty"`
	UseLocalDefault *bool    `json:"use_local_default,omitempty"`
	ToolKeys        []string `json:"tool_keys,omitempty"`
}

type AITask struct {
	ID                 int          `json:"id" gorm:"primaryKey"`
	Type               string       `json:"type" gorm:"not null;index"`
	InputText          string       `json:"input_text" gorm:"type:text"`
	ContextScope       string       `json:"context_scope"`
	PromptTemplateIDs  string       `json:"prompt_template_ids"`
	CustomPrompt       string       `json:"custom_prompt" gorm:"type:text"`
	Status             string       `json:"status" gorm:"not null;index"`
	Progress           int          `json:"progress"`
	ErrorMessage       string       `json:"error_message" gorm:"type:text"`
	ResultProfileID    *int         `json:"result_profile_id,omitempty"`
	ResultSummary      string       `json:"result_summary" gorm:"type:text"`
	ResultJSON         string       `json:"result_json" gorm:"type:text"`
	ConfigSnapshotJSON string       `json:"config_snapshot_json" gorm:"type:text"`
	ContextPayloadJSON string       `json:"context_payload_json" gorm:"type:text"`
	PromptText         string       `json:"prompt_text" gorm:"type:text"`
	SelectedModel      string       `json:"selected_model" gorm:"index"`
	ModelReason        string       `json:"model_reason" gorm:"type:text"`
	ResumeToken        string       `json:"resume_token" gorm:"index"`
	ResumeState        string       `json:"resume_state" gorm:"index"`
	ExecutorVersion    string       `json:"executor_version"`
	LastHeartbeatAt    *time.Time   `json:"last_heartbeat_at,omitempty"`
	AttemptCount       int          `json:"attempt_count"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
	StartedAt          *time.Time   `json:"started_at,omitempty"`
	FinishedAt         *time.Time   `json:"finished_at,omitempty"`
	Steps              []AITaskStep `json:"steps,omitempty" gorm:"foreignKey:TaskID"`
}

type AITaskStep struct {
	ID              int        `json:"id" gorm:"primaryKey"`
	TaskID          int        `json:"task_id" gorm:"not null;index"`
	StepKey         string     `json:"step_key" gorm:"not null"`
	Name            string     `json:"name"`
	Status          string     `json:"status" gorm:"not null"`
	Message         string     `json:"message" gorm:"type:text"`
	InputJSON       string     `json:"input_json" gorm:"type:text"`
	OutputJSON      string     `json:"output_json" gorm:"type:text"`
	CheckpointState string     `json:"checkpoint_state" gorm:"index"`
	RetryCount      int        `json:"retry_count"`
	SortOrder       int        `json:"sort_order"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
}

type AITaskCreateRequest struct {
	Type              string                  `json:"type" binding:"required"`
	InputText         string                  `json:"input_text"`
	ContextScope      string                  `json:"context_scope"`
	PromptTemplateIDs []int                   `json:"prompt_template_ids"`
	CustomPrompt      string                  `json:"custom_prompt"`
	ConfigSnapshot    *AIAutomationTaskConfig `json:"config_snapshot,omitempty"`
}

type AITaskListRequest struct {
	Page          int       `json:"page"`
	PageSize      int       `json:"page_size"`
	Status        string    `json:"status"`
	Type          string    `json:"type"`
	ProfileDomain string    `json:"profile_domain"`
	Keyword       string    `json:"keyword"`
	CreatedFrom   time.Time `json:"created_from"`
	CreatedTo     time.Time `json:"created_to"`
}

type AITaskListResult struct {
	Items    []AITask `json:"items"`
	Total    int64    `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
}

type AITaskArtifacts struct {
	TaskID             int                     `json:"task_id"`
	ConfigSnapshotJSON string                  `json:"config_snapshot_json,omitempty"`
	ConfigSnapshot     *AIAutomationTaskConfig `json:"config_snapshot,omitempty"`
	ContextPayloadJSON string                  `json:"context_payload_json,omitempty"`
	ContextPayload     any                     `json:"context_payload,omitempty"`
	ResultJSON         string                  `json:"result_json,omitempty"`
	ResultPayload      any                     `json:"result_payload,omitempty"`
	PromptText         string                  `json:"prompt_text,omitempty"`
	SelectedModel      string                  `json:"selected_model,omitempty"`
	ModelReason        string                  `json:"model_reason,omitempty"`
	ResumeState        string                  `json:"resume_state,omitempty"`
	Steps              []AITaskStep            `json:"steps,omitempty"`
}
type AIPromptTemplate struct {
	ID              int       `json:"id" gorm:"primaryKey"`
	Name            string    `json:"name" gorm:"not null"`
	Source          string    `json:"source" gorm:"not null;index"`
	TaskType        string    `json:"task_type" gorm:"not null;index"`
	Domain          string    `json:"domain" gorm:"index"`
	Prompt          string    `json:"prompt" gorm:"type:text"`
	WorkRequirement string    `json:"work_requirement" gorm:"type:text"`
	Enabled         bool      `json:"enabled" gorm:"default:true"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AIPromptTemplateCreateRequest struct {
	Name            string `json:"name" binding:"required"`
	TaskType        string `json:"task_type" binding:"required"`
	Domain          string `json:"domain"`
	Prompt          string `json:"prompt" binding:"required"`
	WorkRequirement string `json:"work_requirement"`
	Enabled         *bool  `json:"enabled,omitempty"`
}

type AIProfile struct {
	ID                int                `json:"id" gorm:"primaryKey"`
	Domain            string             `json:"domain" gorm:"not null;index"`
	Name              string             `json:"name" gorm:"not null"`
	Version           int                `json:"version" gorm:"not null;default:1"`
	Status            string             `json:"status" gorm:"not null;index"`
	Confidence        float64            `json:"confidence"`
	Explanation       string             `json:"explanation" gorm:"type:text"`
	SourceTaskID      *int               `json:"source_task_id,omitempty"`
	MigrationStatus   string             `json:"migration_status" gorm:"index"`
	MigrationError    string             `json:"migration_error,omitempty" gorm:"type:text"`
	DomainPayloadType string             `json:"domain_payload_type,omitempty" gorm:"-"`
	DomainPayload     any                `json:"domain_payload,omitempty" gorm:"-"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
	Versions          []AIProfileVersion `json:"versions,omitempty" gorm:"foreignKey:ProfileID"`
}

type AIProfileVersion struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	ProfileID   int       `json:"profile_id" gorm:"not null;index"`
	Version     int       `json:"version" gorm:"not null"`
	ContentJSON string    `json:"content_json" gorm:"type:text"`
	Explanation string    `json:"explanation" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
}

type AIGroupingProfile struct {
	ID               int       `json:"id" gorm:"primaryKey"`
	ProfileID        int       `json:"profile_id" gorm:"not null;index"`
	Version          int       `json:"version" gorm:"not null"`
	TaskID           *int      `json:"task_id,omitempty" gorm:"index"`
	Status           string    `json:"status" gorm:"not null;index"`
	Confidence       float64   `json:"confidence"`
	RiskLevel        string    `json:"risk_level" gorm:"index"`
	Summary          string    `json:"summary" gorm:"type:text"`
	TypedPayloadJSON string    `json:"typed_payload_json" gorm:"type:text"`
	TypedPayloadHash string    `json:"typed_payload_hash" gorm:"index"`
	CreatedAt        time.Time `json:"created_at"`
}

type AIChannelRecognitionProfile struct {
	ID               int       `json:"id" gorm:"primaryKey"`
	ProfileID        int       `json:"profile_id" gorm:"not null;index"`
	Version          int       `json:"version" gorm:"not null"`
	TaskID           *int      `json:"task_id,omitempty" gorm:"index"`
	Status           string    `json:"status" gorm:"not null;index"`
	Confidence       float64   `json:"confidence"`
	RiskLevel        string    `json:"risk_level" gorm:"index"`
	Summary          string    `json:"summary" gorm:"type:text"`
	TypedPayloadJSON string    `json:"typed_payload_json" gorm:"type:text"`
	TypedPayloadHash string    `json:"typed_payload_hash" gorm:"index"`
	CreatedAt        time.Time `json:"created_at"`
}

type AIPriceRecognitionProfile struct {
	ID               int       `json:"id" gorm:"primaryKey"`
	ProfileID        int       `json:"profile_id" gorm:"not null;index"`
	Version          int       `json:"version" gorm:"not null"`
	TaskID           *int      `json:"task_id,omitempty" gorm:"index"`
	Status           string    `json:"status" gorm:"not null;index"`
	Confidence       float64   `json:"confidence"`
	RiskLevel        string    `json:"risk_level" gorm:"index"`
	Summary          string    `json:"summary" gorm:"type:text"`
	TypedPayloadJSON string    `json:"typed_payload_json" gorm:"type:text"`
	TypedPayloadHash string    `json:"typed_payload_hash" gorm:"index"`
	CreatedAt        time.Time `json:"created_at"`
}

type AIModelClassificationProfile struct {
	ID               int       `json:"id" gorm:"primaryKey"`
	ProfileID        int       `json:"profile_id" gorm:"not null;index"`
	Version          int       `json:"version" gorm:"not null"`
	TaskID           *int      `json:"task_id,omitempty" gorm:"index"`
	Status           string    `json:"status" gorm:"not null;index"`
	Confidence       float64   `json:"confidence"`
	RiskLevel        string    `json:"risk_level" gorm:"index"`
	Summary          string    `json:"summary" gorm:"type:text"`
	TypedPayloadJSON string    `json:"typed_payload_json" gorm:"type:text"`
	TypedPayloadHash string    `json:"typed_payload_hash" gorm:"index"`
	CreatedAt        time.Time `json:"created_at"`
}

type AIConfigHealthProfile struct {
	ID               int       `json:"id" gorm:"primaryKey"`
	ProfileID        int       `json:"profile_id" gorm:"not null;index"`
	Version          int       `json:"version" gorm:"not null"`
	TaskID           *int      `json:"task_id,omitempty" gorm:"index"`
	Status           string    `json:"status" gorm:"not null;index"`
	Confidence       float64   `json:"confidence"`
	RiskLevel        string    `json:"risk_level" gorm:"index"`
	Summary          string    `json:"summary" gorm:"type:text"`
	TypedPayloadJSON string    `json:"typed_payload_json" gorm:"type:text"`
	TypedPayloadHash string    `json:"typed_payload_hash" gorm:"index"`
	CreatedAt        time.Time `json:"created_at"`
}

type DynamicRouteLearningState struct {
	ID              int       `json:"id" gorm:"primaryKey"`
	ChannelID       int       `json:"channel_id" gorm:"not null;uniqueIndex:idx_dynamic_route_learning_target"`
	ChannelKeyID    int       `json:"channel_key_id" gorm:"not null;uniqueIndex:idx_dynamic_route_learning_target"`
	ModelName       string    `json:"model_name" gorm:"not null;uniqueIndex:idx_dynamic_route_learning_target"`
	SuccessCount    int64     `json:"success_count"`
	FailureCount    int64     `json:"failure_count"`
	FallbackCount   int64     `json:"fallback_count"`
	RaceWinnerCount int64     `json:"race_winner_count"`
	LatencyMsEWMA   float64   `json:"latency_ms_ewma"`
	Score           float64   `json:"score"`
	Confidence      float64   `json:"confidence"`
	RecentSamples   string    `json:"recent_samples" gorm:"type:text"`
	LastSampleAt    int64     `json:"last_sample_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type DynamicRouteLearningListResult struct {
	Enabled bool                        `json:"enabled"`
	States  []DynamicRouteLearningState `json:"states"`
}

func NormalizeAITaskType(input string) string {
	v := strings.ToLower(strings.TrimSpace(input))
	if v == "" {
		return AIAutomationTaskTypeNaturalLanguage
	}
	return v
}

func IsValidAITaskType(input string) bool {
	switch NormalizeAITaskType(input) {
	case AIAutomationTaskTypeNaturalLanguage,
		AIAutomationTaskTypeGroupSuggestion,
		AIAutomationTaskTypeChannelRecognition,
		AIAutomationTaskTypePriceRecognition,
		AIAutomationTaskTypeModelClassification,
		AIAutomationTaskTypeConfigHealthCheck,
		AIAutomationTaskTypeDynamicRoutingDigest:
		return true
	default:
		return false
	}
}

func NormalizeAIProfileDomain(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

func IsValidAIProfileDomain(input string) bool {
	switch NormalizeAIProfileDomain(input) {
	case AIProfileDomainGrouping,
		AIProfileDomainChannelRecognition,
		AIProfileDomainPriceRecognition,
		AIProfileDomainModelClassification,
		AIProfileDomainConfigHealthCheck,
		AIProfileDomainDynamicRoutingDigest:
		return true
	default:
		return false
	}
}

func NormalizeAITaskToolKey(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

func IsValidAITaskToolKey(input string) bool {
	switch NormalizeAITaskToolKey(input) {
	case AITaskToolKeyChannelInventory,
		AITaskToolKeyGroupTopology,
		AITaskToolKeyPriceCatalog,
		AITaskToolKeyModelCatalog,
		AITaskToolKeyRouteOverrides,
		AITaskToolKeyLearningState,
		AITaskToolKeyProfileWrite,
		AITaskToolKeyProfileActivate,
		AITaskToolKeySnapshotGuard:
		return true
	default:
		return false
	}
}
