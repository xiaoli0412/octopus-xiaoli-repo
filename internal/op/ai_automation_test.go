package op

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	transformerOutbound "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
)

func boolPtr(v bool) *bool {
	return &v
}

func TestAIAutomationConfigGetUsesActiveProfileEffectiveConfig(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationBaseURL, "http://127.0.0.1:8080/v1"); err != nil {
		t.Fatalf("SettingSetString(base_url) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationChannelType, "openai-compatible"); err != nil {
		t.Fatalf("SettingSetString(channel_type) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationModel, "manual-model"); err != nil {
		t.Fatalf("SettingSetString(model) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationUseLocalDefault, "true"); err != nil {
		t.Fatalf("SettingSetString(use_local_default) error = %v", err)
	}
	profile, err := AIProfileCreate(model.AIProfile{Domain: model.AIProfileDomainGrouping, Name: "effective profile", Status: model.AIProfileStatusReady}, `{"config":{"base_url":"https://profile.example/v1","channel_type":"anthropic","model":"profile-model","use_local_default":false}}`, ctx)
	if err != nil {
		t.Fatalf("AIProfileCreate() error = %v", err)
	}
	if _, err := AIProfileActivate(profile.ID, ctx); err != nil {
		t.Fatalf("AIProfileActivate() error = %v", err)
	}

	config, err := AIAutomationConfigGet(ctx)
	if err != nil {
		t.Fatalf("AIAutomationConfigGet() error = %v", err)
	}
	if config.ManualConfig.BaseURL != "http://127.0.0.1:8080/v1" || config.ManualConfig.Model != "manual-model" || !config.ManualConfig.UseLocalDefault {
		t.Fatalf("manual config = %#v, want preserved manual values", config.ManualConfig)
	}
	if config.RequestedConfigSourceMode != model.ConfigSourceModeAIProfile || config.RequestedActiveAIProfileID != profile.ID {
		t.Fatalf("requested source = %#v, want requested ai_profile with profile id", config)
	}
	if config.RequestedActiveAIProfile == nil || config.RequestedActiveAIProfile.ID != profile.ID || config.RequestedActiveAIProfile.Name != "effective profile" {
		t.Fatalf("requested profile ref = %#v, want populated requested profile summary", config.RequestedActiveAIProfile)
	}
	if config.EffectiveConfig.BaseURL != "https://profile.example/v1" || config.EffectiveConfig.ChannelType != "anthropic" || config.EffectiveConfig.Model != "profile-model" || config.EffectiveConfig.UseLocalDefault {
		t.Fatalf("effective config = %#v, want profile-backed values", config.EffectiveConfig)
	}
	if config.ConfigSourceMode != model.ConfigSourceModeAIProfile || config.ActiveAIProfileID != profile.ID || config.SourceFallbackReason != "" {
		t.Fatalf("effective source = %#v, want active ai_profile without fallback", config)
	}
	if config.ActiveAIProfile == nil || config.ActiveAIProfile.ID != profile.ID || config.ActiveAIProfile.Name != "effective profile" {
		t.Fatalf("active profile ref = %#v, want populated active profile summary", config.ActiveAIProfile)
	}
	if config.BaseURL != config.EffectiveConfig.BaseURL || config.Model != config.EffectiveConfig.Model || config.ChannelType != config.EffectiveConfig.ChannelType || config.UseLocalDefault != config.EffectiveConfig.UseLocalDefault {
		t.Fatalf("top-level config = %#v, want effective config mirrored", config)
	}
}

func TestAIAutomationConfigGetRedactsSecretsForResponseWhileRawConfigKeepsThem(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationBaseURL, "http://127.0.0.1:8080/v1"); err != nil {
		t.Fatalf("SettingSetString(base_url) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationAPIKey, "manual-secret-key"); err != nil {
		t.Fatalf("SettingSetString(api_key) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationChannelType, "openai-compatible"); err != nil {
		t.Fatalf("SettingSetString(channel_type) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationModel, "manual-model"); err != nil {
		t.Fatalf("SettingSetString(model) error = %v", err)
	}
	profile, err := AIProfileCreate(model.AIProfile{Domain: model.AIProfileDomainGrouping, Name: "secret profile", Status: model.AIProfileStatusReady}, `{"config":{"base_url":"https://profile.example/v1","api_key":"profile-secret-key","channel_type":"anthropic","model":"profile-model","use_local_default":false}}`, ctx)
	if err != nil {
		t.Fatalf("AIProfileCreate() error = %v", err)
	}
	if _, err := AIProfileActivate(profile.ID, ctx); err != nil {
		t.Fatalf("AIProfileActivate() error = %v", err)
	}

	rawConfig, err := aiAutomationConfigGetRaw(ctx)
	if err != nil {
		t.Fatalf("aiAutomationConfigGetRaw() error = %v", err)
	}
	if rawConfig.ManualConfig.APIKey != "manual-secret-key" {
		t.Fatalf("raw manual api key = %q, want manual-secret-key", rawConfig.ManualConfig.APIKey)
	}
	if rawConfig.EffectiveConfig.APIKey != "profile-secret-key" || rawConfig.APIKey != "profile-secret-key" {
		t.Fatalf("raw effective config = %#v, want preserved profile secret", rawConfig)
	}

	responseConfig, err := AIAutomationConfigGet(ctx)
	if err != nil {
		t.Fatalf("AIAutomationConfigGet() error = %v", err)
	}
	if responseConfig.ManualConfig.APIKey != aiAutomationRedactedSecret {
		t.Fatalf("response manual api key = %q, want redacted", responseConfig.ManualConfig.APIKey)
	}
	if responseConfig.EffectiveConfig.APIKey != aiAutomationRedactedSecret || responseConfig.APIKey != aiAutomationRedactedSecret {
		t.Fatalf("response effective config = %#v, want redacted secrets", responseConfig)
	}
}

func TestAIAutomationConfigGetFallsBackToManualWhenActiveProfileInvalid(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationBaseURL, "http://127.0.0.1:8080/v1"); err != nil {
		t.Fatalf("SettingSetString(base_url) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationChannelType, "openai-compatible"); err != nil {
		t.Fatalf("SettingSetString(channel_type) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationModel, "manual-model"); err != nil {
		t.Fatalf("SettingSetString(model) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationUseLocalDefault, "true"); err != nil {
		t.Fatalf("SettingSetString(use_local_default) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyConfigSourceMode, model.ConfigSourceModeAIProfile); err != nil {
		t.Fatalf("SettingSetString(config_source_mode) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyActiveAIProfileID, "999"); err != nil {
		t.Fatalf("SettingSetString(active_ai_profile_id) error = %v", err)
	}

	config, err := AIAutomationConfigGet(ctx)
	if err != nil {
		t.Fatalf("AIAutomationConfigGet() error = %v", err)
	}
	if config.ConfigSourceMode != model.ConfigSourceModeManual || config.ActiveAIProfileID != 0 {
		t.Fatalf("config source fallback = %#v, want manual mode and cleared active profile", config)
	}
	if config.RequestedConfigSourceMode != model.ConfigSourceModeAIProfile || config.RequestedActiveAIProfileID != 999 {
		t.Fatalf("requested source after fallback = %#v, want preserved ai_profile request", config)
	}
	if config.RequestedActiveAIProfile != nil {
		t.Fatalf("requested profile ref = %#v, want nil when requested profile is missing", config.RequestedActiveAIProfile)
	}
	if config.SourceFallbackReason != "profile_missing" {
		t.Fatalf("config.SourceFallbackReason = %q, want profile_missing", config.SourceFallbackReason)
	}
	if config.ManualConfig.BaseURL != "http://127.0.0.1:8080/v1" || config.EffectiveConfig.BaseURL != "http://127.0.0.1:8080/v1" {
		t.Fatalf("manual/effective config = %#v, want manual fallback", config)
	}
	mode, err := SettingGetString(model.SettingKeyConfigSourceMode)
	if err != nil || mode != model.ConfigSourceModeAIProfile {
		t.Fatalf("persisted config_source_mode = %q, err = %v, want original ai_profile request preserved", mode, err)
	}
	activeID, err := SettingGetInt(model.SettingKeyActiveAIProfileID)
	if err != nil || activeID != 999 {
		t.Fatalf("persisted active_ai_profile_id = %d, err = %v, want original invalid request preserved", activeID, err)
	}
}

func TestAIAutomationConfigGetFallsBackToManualWhenProfileContentInvalid(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationBaseURL, "http://127.0.0.1:8080/v1"); err != nil {
		t.Fatalf("SettingSetString(base_url) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationChannelType, "openai-compatible"); err != nil {
		t.Fatalf("SettingSetString(channel_type) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationModel, "manual-model"); err != nil {
		t.Fatalf("SettingSetString(model) error = %v", err)
	}
	profile, err := AIProfileCreate(model.AIProfile{Domain: model.AIProfileDomainGrouping, Name: "invalid content profile", Status: model.AIProfileStatusReady}, `{"groups":[]}`, ctx)
	if err != nil {
		t.Fatalf("AIProfileCreate() error = %v", err)
	}
	if _, err := AIProfileActivate(profile.ID, ctx); err != nil {
		t.Fatalf("AIProfileActivate() error = %v", err)
	}

	config, err := AIAutomationConfigGet(ctx)
	if err != nil {
		t.Fatalf("AIAutomationConfigGet() error = %v", err)
	}
	if config.ConfigSourceMode != model.ConfigSourceModeManual || config.ActiveAIProfileID != 0 {
		t.Fatalf("config source fallback = %#v, want manual mode and cleared active profile", config)
	}
	if config.RequestedConfigSourceMode != model.ConfigSourceModeAIProfile || config.RequestedActiveAIProfileID != profile.ID {
		t.Fatalf("requested source after invalid content fallback = %#v, want preserved ai_profile request", config)
	}
	if config.RequestedActiveAIProfile == nil || config.RequestedActiveAIProfile.ID != profile.ID || config.RequestedActiveAIProfile.Name != "invalid content profile" {
		t.Fatalf("requested profile ref = %#v, want requested invalid-content profile summary preserved", config.RequestedActiveAIProfile)
	}
	if config.SourceFallbackReason != "profile_invalid_content" {
		t.Fatalf("config.SourceFallbackReason = %q, want profile_invalid_content", config.SourceFallbackReason)
	}
	if config.ActiveAIProfile != nil {
		t.Fatalf("active profile ref = %#v, want nil after runtime fallback", config.ActiveAIProfile)
	}
}

func TestAIProfileCreatePersistsTypedPayloadAndDetail(t *testing.T) {
	ctx := SetupOpTestDB(t)
	content := `{"summary":"grouping typed summary","domain_payload":{"summary":"grouping typed summary","grouping_suggestions":[{"group_name":"free","item_count":2}],"candidate_channel_model_mappings":[{"channel_name":"alpha","models":["gpt-free"]}],"conflicts":[{"type":"empty_group"}],"typed_config":{"base_url":"https://typed.example/v1","channel_type":"openai","model":"typed-model","use_local_default":false}},"findings":[{"title":"coverage"}],"recommendations":["review groups"],"risks":["manual review"],"config":{"base_url":"https://typed.example/v1","channel_type":"openai","model":"typed-model","use_local_default":false},"groups":[{"name":"free"}]}`
	profile, err := AIProfileCreate(model.AIProfile{Domain: model.AIProfileDomainGrouping, Name: "typed grouping", Status: model.AIProfileStatusReady, Confidence: 0.8}, content, ctx)
	if err != nil {
		t.Fatalf("AIProfileCreate() error = %v", err)
	}
	if profile.MigrationStatus != model.AIProfileMigrationStatusTypedBackfilled {
		t.Fatalf("MigrationStatus = %q, want typed_backfilled", profile.MigrationStatus)
	}
	if profile.DomainPayloadType != model.AIProfileDomainGrouping {
		t.Fatalf("DomainPayloadType = %q, want grouping", profile.DomainPayloadType)
	}
	payload, ok := profile.DomainPayload.(map[string]any)
	if !ok {
		t.Fatalf("DomainPayload = %#v, want object", profile.DomainPayload)
	}
	if payload["summary"] != "grouping typed summary" {
		t.Fatalf("payload summary = %#v, want typed summary", payload["summary"])
	}
	groupingSuggestions, ok := payload["grouping_suggestions"].([]any)
	if !ok || len(groupingSuggestions) != 1 {
		t.Fatalf("grouping_suggestions = %#v, want one structured suggestion", payload["grouping_suggestions"])
	}
	var typedCount int64
	if err := db.GetDB().WithContext(ctx).Model(&model.AIGroupingProfile{}).Where("profile_id = ?", profile.ID).Count(&typedCount).Error; err != nil {
		t.Fatalf("count typed profile error = %v", err)
	}
	if typedCount != 1 {
		t.Fatalf("typed profile count = %d, want 1", typedCount)
	}
}

func TestAITaskListFiltersAndPaginates(t *testing.T) {
	ctx := SetupOpTestDB(t)
	older := time.Now().Add(-time.Hour)
	tasks := []model.AITask{
		{Type: model.AIAutomationTaskTypeNaturalLanguage, InputText: "first task", Status: model.AITaskStatusSucceeded, ResultSummary: "alpha", CreatedAt: older},
		{Type: model.AIAutomationTaskTypeConfigHealthCheck, InputText: "second task", Status: model.AITaskStatusFailedUnrecoverable, ErrorMessage: "missing config", SelectedModel: "health-model"},
		{Type: model.AIAutomationTaskTypeConfigHealthCheck, InputText: "third task", Status: model.AITaskStatusSucceeded, ResultSummary: "health ok", SelectedModel: "health-model"},
	}
	if err := db.GetDB().WithContext(ctx).Create(&tasks).Error; err != nil {
		t.Fatalf("create tasks error = %v", err)
	}
	result, err := AITaskList(model.AITaskListRequest{Page: 1, PageSize: 1, Type: model.AIAutomationTaskTypeConfigHealthCheck, Keyword: "health"}, ctx)
	if err != nil {
		t.Fatalf("AITaskList() error = %v", err)
	}
	if result.Total != 2 || len(result.Items) != 1 || result.Page != 1 || result.PageSize != 1 {
		t.Fatalf("list result = %#v, want total 2 and one item page", result)
	}
	if result.Items[0].Type != model.AIAutomationTaskTypeConfigHealthCheck {
		t.Fatalf("item type = %q, want config_health_check", result.Items[0].Type)
	}
}

func TestAITaskRetryCreatesNewPendingTaskFromSnapshot(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}
	useLocalDefault := false
	snapshotRaw, err := json.Marshal(model.AIAutomationTaskConfig{
		BaseURL:         "https://retry.example/v1",
		APIKey:          "retry-key",
		ChannelType:     "openai",
		Model:           "retry-model",
		UseLocalDefault: &useLocalDefault,
		ToolKeys:        []string{model.AITaskToolKeyChannelInventory, model.AITaskToolKeySnapshotGuard},
	})
	if err != nil {
		t.Fatalf("json.Marshal(snapshot) error = %v", err)
	}
	original := model.AITask{Type: model.AIAutomationTaskTypeGroupSuggestion, InputText: "retry me", ContextScope: "channels", PromptTemplateIDs: "7,2", CustomPrompt: "focus on compact retries", Status: model.AITaskStatusSucceeded, ConfigSnapshotJSON: string(snapshotRaw)}
	if err := db.GetDB().WithContext(ctx).Create(&original).Error; err != nil {
		t.Fatalf("create original task error = %v", err)
	}
	retried, err := AITaskRetry(original.ID, ctx)
	if err != nil {
		t.Fatalf("AITaskRetry() error = %v", err)
	}
	if retried.ID == original.ID || retried.Status != model.AITaskStatusPending {
		t.Fatalf("retried task = %#v, want a new pending task", retried)
	}
	if retried.Type != original.Type || retried.InputText != original.InputText || retried.ContextScope != original.ContextScope || retried.CustomPrompt != original.CustomPrompt {
		t.Fatalf("retried task fields = %#v, want original input fields preserved", retried)
	}
	if retried.PromptTemplateIDs != "7,2" || len(retried.Steps) != 6 {
		t.Fatalf("retried templates/steps = %q/%d, want 7,2 and 6 steps", retried.PromptTemplateIDs, len(retried.Steps))
	}
	var snapshot model.AIAutomationTaskConfig
	if err := json.Unmarshal([]byte(retried.ConfigSnapshotJSON), &snapshot); err != nil {
		t.Fatalf("json.Unmarshal(retried snapshot) error = %v", err)
	}
	if snapshot.BaseURL != "https://retry.example/v1" || snapshot.APIKey != "retry-key" || snapshot.Model != "retry-model" || snapshot.UseLocalDefault == nil || *snapshot.UseLocalDefault {
		t.Fatalf("retried snapshot = %#v, want preserved original snapshot", snapshot)
	}
	if len(snapshot.ToolKeys) != 2 {
		t.Fatalf("snapshot.ToolKeys = %#v, want preserved tool keys", snapshot.ToolKeys)
	}
}

func TestAITaskGetRedactsConfigSnapshotSecrets(t *testing.T) {
	ctx := SetupOpTestDB(t)
	useLocalDefault := false
	snapshotRaw, err := json.Marshal(model.AIAutomationTaskConfig{
		BaseURL:         "https://redact.example/v1",
		APIKey:          "super-secret-key",
		ChannelType:     "openai",
		Model:           "redact-model",
		UseLocalDefault: &useLocalDefault,
	})
	if err != nil {
		t.Fatalf("json.Marshal(snapshot) error = %v", err)
	}
	task := model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, Status: model.AITaskStatusSucceeded, ConfigSnapshotJSON: string(snapshotRaw)}
	if err := db.GetDB().WithContext(ctx).Create(&task).Error; err != nil {
		t.Fatalf("create task error = %v", err)
	}
	got, err := AITaskGet(task.ID, ctx)
	if err != nil {
		t.Fatalf("AITaskGet() error = %v", err)
	}
	if strings.Contains(got.ConfigSnapshotJSON, "super-secret-key") {
		t.Fatalf("ConfigSnapshotJSON leaked API key: %s", got.ConfigSnapshotJSON)
	}
	if !strings.Contains(got.ConfigSnapshotJSON, aiAutomationRedactedSecret) {
		t.Fatalf("ConfigSnapshotJSON = %s, want redacted marker", got.ConfigSnapshotJSON)
	}
	var redacted model.AIAutomationTaskConfig
	if err := json.Unmarshal([]byte(got.ConfigSnapshotJSON), &redacted); err != nil {
		t.Fatalf("json.Unmarshal(redacted snapshot) error = %v", err)
	}
	if redacted.APIKey != aiAutomationRedactedSecret {
		t.Fatalf("redacted.APIKey = %q, want %q", redacted.APIKey, aiAutomationRedactedSecret)
	}
}

func TestAITaskListRedactsConfigSnapshotSecrets(t *testing.T) {
	ctx := SetupOpTestDB(t)
	useLocalDefault := false
	snapshotRaw, err := json.Marshal(model.AIAutomationTaskConfig{
		BaseURL:         "https://list-redact.example/v1",
		APIKey:          "list-secret-key",
		ChannelType:     "openai",
		Model:           "list-model",
		UseLocalDefault: &useLocalDefault,
	})
	if err != nil {
		t.Fatalf("json.Marshal(snapshot) error = %v", err)
	}
	task := model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, Status: model.AITaskStatusSucceeded, ConfigSnapshotJSON: string(snapshotRaw)}
	if err := db.GetDB().WithContext(ctx).Create(&task).Error; err != nil {
		t.Fatalf("create task error = %v", err)
	}
	result, err := AITaskList(model.AITaskListRequest{Page: 1, PageSize: 20}, ctx)
	if err != nil {
		t.Fatalf("AITaskList() error = %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("len(result.Items) = %d, want 1", len(result.Items))
	}
	if strings.Contains(result.Items[0].ConfigSnapshotJSON, "list-secret-key") {
		t.Fatalf("ConfigSnapshotJSON leaked API key: %s", result.Items[0].ConfigSnapshotJSON)
	}
	if !strings.Contains(result.Items[0].ConfigSnapshotJSON, aiAutomationRedactedSecret) {
		t.Fatalf("ConfigSnapshotJSON = %s, want redacted marker", result.Items[0].ConfigSnapshotJSON)
	}
}

func TestAITaskArtifactsReturnsParsedPayloads(t *testing.T) {
	ctx := SetupOpTestDB(t)
	useLocalDefault := false
	snapshotRaw, err := json.Marshal(model.AIAutomationTaskConfig{BaseURL: "https://artifact.example/v1", ChannelType: "anthropic", Model: "artifact-model", UseLocalDefault: &useLocalDefault})
	if err != nil {
		t.Fatalf("json.Marshal(snapshot) error = %v", err)
	}
	task := model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, Status: model.AITaskStatusSucceeded, ConfigSnapshotJSON: string(snapshotRaw), ContextPayloadJSON: `{"counts":{"channels":1},"scope":"history"}`, PromptText: "artifact prompt", SelectedModel: "artifact-model", ModelReason: "history replay", ResumeState: model.AITaskResumeStateGenerateProfile, ResultJSON: `{"summary":"ok","findings":["a"]}`}
	if err := db.GetDB().WithContext(ctx).Create(&task).Error; err != nil {
		t.Fatalf("create task error = %v", err)
	}
	steps := defaultAITaskSteps(task.ID)
	steps[0].Status = model.AITaskStepStatusSucceeded
	steps[0].OutputJSON = `{"counts":{"channels":1}}`
	if err := db.GetDB().WithContext(ctx).Create(&steps).Error; err != nil {
		t.Fatalf("create steps error = %v", err)
	}
	artifacts, err := AITaskArtifacts(task.ID, ctx)
	if err != nil {
		t.Fatalf("AITaskArtifacts() error = %v", err)
	}
	if artifacts.TaskID != task.ID || artifacts.ConfigSnapshot == nil || artifacts.ConfigSnapshot.BaseURL != "https://artifact.example/v1" || artifacts.PromptText != "artifact prompt" || artifacts.SelectedModel != "artifact-model" || len(artifacts.Steps) != 6 {
		t.Fatalf("artifacts = %#v, want parsed snapshot, prompt metadata, and steps", artifacts)
	}
	contextPayload, ok := artifacts.ContextPayload.(map[string]any)
	if !ok || contextPayload["scope"] != "history" {
		t.Fatalf("context payload = %#v, want parsed history payload", artifacts.ContextPayload)
	}
	resultPayload, ok := artifacts.ResultPayload.(map[string]any)
	if !ok || resultPayload["summary"] != "ok" {
		t.Fatalf("result payload = %#v, want parsed result payload", artifacts.ResultPayload)
	}
}

func TestAITaskArtifactsRedactsConfigSnapshotSecrets(t *testing.T) {
	ctx := SetupOpTestDB(t)
	useLocalDefault := false
	snapshotRaw, err := json.Marshal(model.AIAutomationTaskConfig{
		BaseURL:         "https://artifact-redact.example/v1",
		APIKey:          "artifact-secret-key",
		ChannelType:     "anthropic",
		Model:           "artifact-model",
		UseLocalDefault: &useLocalDefault,
	})
	if err != nil {
		t.Fatalf("json.Marshal(snapshot) error = %v", err)
	}
	task := model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, Status: model.AITaskStatusSucceeded, ConfigSnapshotJSON: string(snapshotRaw), ResultJSON: `{"summary":"ok"}`}
	if err := db.GetDB().WithContext(ctx).Create(&task).Error; err != nil {
		t.Fatalf("create task error = %v", err)
	}
	artifacts, err := AITaskArtifacts(task.ID, ctx)
	if err != nil {
		t.Fatalf("AITaskArtifacts() error = %v", err)
	}
	if strings.Contains(artifacts.ConfigSnapshotJSON, "artifact-secret-key") {
		t.Fatalf("ConfigSnapshotJSON leaked API key: %s", artifacts.ConfigSnapshotJSON)
	}
	if artifacts.ConfigSnapshot == nil || artifacts.ConfigSnapshot.APIKey != aiAutomationRedactedSecret {
		t.Fatalf("ConfigSnapshot = %#v, want redacted api key", artifacts.ConfigSnapshot)
	}
}

func TestRedactAIProfileForResponseRedactsVersionAndDomainPayloadSecrets(t *testing.T) {
	ctx := SetupOpTestDB(t)
	content := `{"summary":"grouping typed summary","domain_payload":{"summary":"grouping typed summary","typed_config":{"base_url":"https://typed.example/v1","api_key":"profile-secret-key","channel_type":"openai","model":"typed-model","use_local_default":false},"config":{"api_key":"profile-secret-key"},"findings":[],"recommendations":[],"risks":[]},"config":{"base_url":"https://typed.example/v1","api_key":"profile-secret-key","channel_type":"openai","model":"typed-model","use_local_default":false}}`
	profile, err := AIProfileCreate(model.AIProfile{Domain: model.AIProfileDomainGrouping, Name: "typed grouping redacted", Status: model.AIProfileStatusReady, Confidence: 0.8}, content, ctx)
	if err != nil {
		t.Fatalf("AIProfileCreate() error = %v", err)
	}
	fetched, err := AIProfileGet(profile.ID, ctx)
	if err != nil {
		t.Fatalf("AIProfileGet() error = %v", err)
	}
	redacted := RedactAIProfileForResponse(fetched)
	if strings.Contains(fetched.Versions[0].ContentJSON, "profile-secret-key") == false {
		t.Fatalf("AIProfileGet() should preserve internal content, got %s", fetched.Versions[0].ContentJSON)
	}
	if len(redacted.Versions) != 1 {
		t.Fatalf("len(redacted.Versions) = %d, want 1", len(redacted.Versions))
	}
	if strings.Contains(redacted.Versions[0].ContentJSON, "profile-secret-key") {
		t.Fatalf("ContentJSON leaked API key: %s", redacted.Versions[0].ContentJSON)
	}
	if !strings.Contains(redacted.Versions[0].ContentJSON, aiAutomationRedactedSecret) {
		t.Fatalf("ContentJSON = %s, want redacted marker", redacted.Versions[0].ContentJSON)
	}
	payload, ok := redacted.DomainPayload.(map[string]any)
	if !ok {
		t.Fatalf("DomainPayload = %#v, want object", redacted.DomainPayload)
	}
	typedConfig, ok := payload["typed_config"].(map[string]any)
	if !ok || typedConfig["api_key"] != aiAutomationRedactedSecret {
		t.Fatalf("typed_config = %#v, want redacted api_key", payload["typed_config"])
	}
}

func TestAITaskListRejectsOversizedPageSize(t *testing.T) {
	ctx := SetupOpTestDB(t)

	if _, err := AITaskList(model.AITaskListRequest{Page: 1, PageSize: 101}, ctx); err == nil {
		t.Fatal("AITaskList() expected invalid page_size error")
	} else if err.Error() != "invalid page_size" {
		t.Fatalf("AITaskList() error = %v, want invalid page_size", err)
	}
}

func TestAITaskListRejectsOversizedOffset(t *testing.T) {
	ctx := SetupOpTestDB(t)

	if _, err := AITaskList(model.AITaskListRequest{Page: 1001, PageSize: 10}, ctx); err == nil {
		t.Fatal("AITaskList() expected invalid page error")
	} else if err.Error() != "invalid page" {
		t.Fatalf("AITaskList() error = %v, want invalid page", err)
	}
}

func TestAIAutomationFetchModelsPrefersRemoteFreeCandidate(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}

	if err := LLMCreate(model.LLMInfo{Name: "paid-model", BillingMode: model.BillingModePerToken}, ctx); err != nil {
		t.Fatalf("LLMCreate(paid-model) error = %v", err)
	}
	if err := LLMCreate(model.LLMInfo{Name: "free-model", BillingMode: model.BillingModeFree}, ctx); err != nil {
		t.Fatalf("LLMCreate(free-model) error = %v", err)
	}

	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Fatalf("path = %s, want /models", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "paid-model"},
				{"id": "free-model"},
			},
		})
	}))
	defer server.Close()

	result, err := AIAutomationFetchModels(model.AIModelsFetchRequest{
		BaseURL:         server.URL,
		ChannelType:     "openai",
		UseLocalDefault: boolPtr(false),
	}, ctx)
	if err != nil {
		t.Fatalf("AIAutomationFetchModels() error = %v", err)
	}
	if result.Source != model.AIAutomationModelSourceRemoteDiscovery {
		t.Fatalf("result.Source = %q, want %q", result.Source, model.AIAutomationModelSourceRemoteDiscovery)
	}
	if result.SelectedName != "free-model" {
		t.Fatalf("result.SelectedName = %q, want free-model", result.SelectedName)
	}
	if len(result.Candidates) != 2 {
		t.Fatalf("candidate count = %d, want 2", len(result.Candidates))
	}
	if !result.Candidates[0].Recommended || result.Candidates[0].Name != "free-model" {
		t.Fatalf("top candidate = %#v, want recommended free-model", result.Candidates[0])
	}
	if result.Policy != model.AIAutomationDefaultSelectionPolicy {
		t.Fatalf("result.Policy = %q, want %q", result.Policy, model.AIAutomationDefaultSelectionPolicy)
	}
}

func TestAIAutomationFetchModelsRejectsWhenDisabled(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}

	_, err := AIAutomationFetchModels(model.AIModelsFetchRequest{BaseURL: "http://127.0.0.1:8080/v1", ChannelType: "openai"}, ctx)
	if !errors.Is(err, ErrAIAutomationDisabled) {
		t.Fatalf("AIAutomationFetchModels() error = %v, want ErrAIAutomationDisabled", err)
	}
}

func TestAIAutomationFetchModelsReturnsRemoteErrorInsteadOfSilentLocalFallback(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}
	if err := LLMCreate(model.LLMInfo{Name: "cached-model", BillingMode: model.BillingModeFree}, ctx); err != nil {
		t.Fatalf("LLMCreate(cached-model) error = %v", err)
	}

	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"bad key"}}`))
	}))
	defer server.Close()

	_, err := AIAutomationFetchModels(model.AIModelsFetchRequest{
		BaseURL:         server.URL,
		APIKey:          "bad-key",
		ChannelType:     "openai",
		UseLocalDefault: boolPtr(false),
	}, ctx)
	if err == nil {
		t.Fatal("AIAutomationFetchModels() error = nil, want remote discovery failure")
	}
	if !strings.Contains(err.Error(), "remote model discovery returned status 401") {
		t.Fatalf("AIAutomationFetchModels() error = %v, want upstream 401 failure", err)
	}
}

func TestAIAutomationFetchModelsFallsBackToLocalOnlyWhenRemoteNotRequested(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationBaseURL, ""); err != nil {
		t.Fatalf("SettingSetString(base_url) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationAPIKey, ""); err != nil {
		t.Fatalf("SettingSetString(api_key) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationChannelType, ""); err != nil {
		t.Fatalf("SettingSetString(channel_type) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationUseLocalDefault, "false"); err != nil {
		t.Fatalf("SettingSetString(use_local_default) error = %v", err)
	}
	if err := LLMCreate(model.LLMInfo{Name: "cached-model", BillingMode: model.BillingModeFree}, ctx); err != nil {
		t.Fatalf("LLMCreate(cached-model) error = %v", err)
	}

	result, err := AIAutomationFetchModels(model.AIModelsFetchRequest{}, ctx)
	if err != nil {
		t.Fatalf("AIAutomationFetchModels() error = %v", err)
	}
	if result.Source != model.AIAutomationModelSourceLocalCache {
		t.Fatalf("result.Source = %q, want %q", result.Source, model.AIAutomationModelSourceLocalCache)
	}
	if result.SelectedName != "cached-model" {
		t.Fatalf("result.SelectedName = %q, want cached-model", result.SelectedName)
	}
}

func TestDecodeRemoteAIAutomationModelsRejectsOversizedResponse(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(&oversizedReader{remaining: maxAIAutomationModelDiscoveryResponseBytes + 1})}

	_, err := decodeRemoteAIAutomationModels(resp, "openai")
	if err == nil {
		t.Fatal("expected oversized response error, got nil")
	}
	if err.Error() != "ai automation model discovery response too large" {
		t.Fatalf("error = %v, want size limit error", err)
	}
}

type oversizedReader struct {
	remaining int64
}

func (r *oversizedReader) Read(p []byte) (int, error) {
	if r.remaining <= 0 {
		return 0, io.EOF
	}
	chunk := int64(len(p))
	if chunk > r.remaining {
		chunk = r.remaining
	}
	for i := int64(0); i < chunk; i++ {
		p[i] = 'a'
	}
	r.remaining -= chunk
	if r.remaining == 0 {
		return int(chunk), io.EOF
	}
	return int(chunk), nil
}

func TestAITaskCreateExecutesAndSavesProfile(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}

	var authHeader string
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %s, want /v1/chat/completions", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]any{"content": "Summary: generated grouping profile\n- keep manual config intact"},
			}},
		})
	}))
	defer server.Close()

	if err := SettingSetString(model.SettingKeyAIAutomationBaseURL, server.URL+"/v1"); err != nil {
		t.Fatalf("SettingSetString(base_url) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationAPIKey, "test-key"); err != nil {
		t.Fatalf("SettingSetString(api_key) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationModel, "free-model"); err != nil {
		t.Fatalf("SettingSetString(model) error = %v", err)
	}

	task, err := AITaskCreate(model.AITaskCreateRequest{Type: model.AIAutomationTaskTypeGroupSuggestion, InputText: "帮我整理分组"}, ctx)
	if err != nil {
		t.Fatalf("AITaskCreate() error = %v", err)
	}

	var latest model.AITask
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		latest, err = AITaskGet(task.ID, ctx)
		if err != nil {
			t.Fatalf("AITaskGet() error = %v", err)
		}
		if latest.Status == model.AITaskStatusSucceeded || latest.Status == model.AITaskStatusFailed {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if latest.Status != model.AITaskStatusSucceeded {
		t.Fatalf("task status = %s, error = %s", latest.Status, latest.ErrorMessage)
	}
	if latest.ResultProfileID == nil || *latest.ResultProfileID <= 0 {
		t.Fatalf("result_profile_id = %#v, want saved profile", latest.ResultProfileID)
	}
	if authHeader != "Bearer test-key" {
		t.Fatalf("authorization header = %q, want Bearer test-key", authHeader)
	}
	profile, err := AIProfileGet(*latest.ResultProfileID, ctx)
	if err != nil {
		t.Fatalf("AIProfileGet() error = %v", err)
	}
	if profile.SourceTaskID == nil || *profile.SourceTaskID != task.ID {
		t.Fatalf("profile.SourceTaskID = %#v, want %d", profile.SourceTaskID, task.ID)
	}
	if len(profile.Versions) != 1 {
		t.Fatalf("profile versions = %d, want 1", len(profile.Versions))
	}
	domainPayload, ok := profile.DomainPayload.(map[string]any)
	if !ok {
		t.Fatalf("profile.DomainPayload = %#v, want structured object", profile.DomainPayload)
	}
	groupingSuggestions, ok := domainPayload["grouping_suggestions"].([]any)
	if !ok || len(groupingSuggestions) == 0 {
		t.Fatalf("grouping_suggestions = %#v, want generated structured grouping suggestions", domainPayload["grouping_suggestions"])
	}
	mappings, ok := domainPayload["candidate_channel_model_mappings"].([]any)
	if !ok {
		t.Fatalf("candidate_channel_model_mappings = %#v, want structured mappings array", domainPayload["candidate_channel_model_mappings"])
	}
	if len(mappings) != 0 {
		t.Fatalf("candidate_channel_model_mappings = %#v, want empty when no channels exist", mappings)
	}
	var channelCount, groupCount, groupItemCount, llmCount, overrideCount int64
	db.GetDB().Model(&model.Channel{}).Count(&channelCount)
	db.GetDB().Model(&model.Group{}).Count(&groupCount)
	db.GetDB().Model(&model.GroupItem{}).Count(&groupItemCount)
	db.GetDB().Model(&model.LLMInfo{}).Count(&llmCount)
	db.GetDB().Model(&model.RouteTargetOverride{}).Count(&overrideCount)
	if channelCount != 0 || groupCount != 0 || groupItemCount != 0 || llmCount != 0 || overrideCount != 0 {
		t.Fatalf("manual tables changed unexpectedly: channels=%d groups=%d items=%d llms=%d overrides=%d", channelCount, groupCount, groupItemCount, llmCount, overrideCount)
	}
}

func TestAITaskCreateConfigSnapshotOverridesRuntimeOnly(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}

	type requestCapture struct {
		Path  string
		Auth  string
		Model string
	}
	captures := make(chan requestCapture, 1)
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("json decode request error = %v", err)
		}
		modelName, _ := payload["model"].(string)
		captures <- requestCapture{Path: r.URL.Path, Auth: r.Header.Get("Authorization"), Model: modelName}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]any{"content": "Summary: snapshot config executed successfully"},
			}},
		})
	}))
	defer server.Close()

	if err := SettingSetString(model.SettingKeyAIAutomationBaseURL, server.URL+"/global-v1"); err != nil {
		t.Fatalf("SettingSetString(global base_url) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationAPIKey, "global-key"); err != nil {
		t.Fatalf("SettingSetString(global api_key) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationModel, "global-model"); err != nil {
		t.Fatalf("SettingSetString(global model) error = %v", err)
	}

	useLocalDefault := false
	task, err := AITaskCreate(model.AITaskCreateRequest{
		Type:      model.AIAutomationTaskTypeNaturalLanguage,
		InputText: "浣跨敤浠诲姟绾?AI 閰嶇疆杩愯",
		ConfigSnapshot: &model.AIAutomationTaskConfig{
			BaseURL:         server.URL + "/snapshot-v1",
			APIKey:          "snapshot-key",
			ChannelType:     "openai",
			Model:           "snapshot-model",
			UseLocalDefault: &useLocalDefault,
		},
	}, ctx)
	if err != nil {
		t.Fatalf("AITaskCreate() error = %v", err)
	}

	var latest model.AITask
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		latest, err = AITaskGet(task.ID, ctx)
		if err != nil {
			t.Fatalf("AITaskGet() error = %v", err)
		}
		if latest.Status == model.AITaskStatusSucceeded || latest.Status == model.AITaskStatusFailed {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if latest.Status != model.AITaskStatusSucceeded {
		t.Fatalf("task status = %s, error = %s", latest.Status, latest.ErrorMessage)
	}

	select {
	case capture := <-captures:
		if capture.Path != "/snapshot-v1/chat/completions" {
			t.Fatalf("request path = %q, want /snapshot-v1/chat/completions", capture.Path)
		}
		if capture.Auth != "Bearer snapshot-key" {
			t.Fatalf("authorization header = %q, want Bearer snapshot-key", capture.Auth)
		}
		if capture.Model != "snapshot-model" {
			t.Fatalf("request model = %q, want snapshot-model", capture.Model)
		}
	default:
		t.Fatalf("expected AI request capture, got none")
	}

	globalBaseURL, err := SettingGetString(model.SettingKeyAIAutomationBaseURL)
	if err != nil {
		t.Fatalf("SettingGetString(global base_url) error = %v", err)
	}
	if globalBaseURL != server.URL+"/global-v1" {
		t.Fatalf("global base_url = %q, want %q", globalBaseURL, server.URL+"/global-v1")
	}
	globalAPIKey, err := SettingGetString(model.SettingKeyAIAutomationAPIKey)
	if err != nil {
		t.Fatalf("SettingGetString(global api_key) error = %v", err)
	}
	if globalAPIKey != "global-key" {
		t.Fatalf("global api_key = %q, want global-key", globalAPIKey)
	}
	globalModel, err := SettingGetString(model.SettingKeyAIAutomationModel)
	if err != nil {
		t.Fatalf("SettingGetString(global model) error = %v", err)
	}
	if globalModel != "global-model" {
		t.Fatalf("global model = %q, want global-model", globalModel)
	}
}

func TestAITaskCreateToolKeysLimitContextAndSkipProfileWrite(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationModel, "tool-model"); err != nil {
		t.Fatalf("SettingSetString(model) error = %v", err)
	}

	channel := &model.Channel{Name: "tool-channel", Type: transformerOutbound.OutboundTypeOpenAIChat, Enabled: true, Model: "tool-model", BaseUrls: []model.BaseUrl{{URL: "https://example.com/v1", Delay: 120}}}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	channelWithKey, err := ChannelUpdate(&model.ChannelUpdateRequest{
		ID:        channel.ID,
		KeysToAdd: []model.ChannelKeyAddRequest{{Enabled: true, ChannelKey: "tool-key", AllowedModels: "tool-model"}},
	}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate() error = %v", err)
	}
	group := &model.Group{Name: "tool-group", Mode: model.GroupModeFailover}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	if err := GroupItemAdd(&model.GroupItem{GroupID: group.ID, ChannelID: channelWithKey.ID, ModelName: "tool-model", Priority: 1, Weight: 1}, ctx); err != nil {
		t.Fatalf("GroupItemAdd() error = %v", err)
	}
	if err := LLMCreate(model.LLMInfo{Name: "tool-model", BillingMode: model.BillingModeFree}, ctx); err != nil {
		t.Fatalf("LLMCreate() error = %v", err)
	}

	var requestBody map[string]any
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("json decode request error = %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]any{"content": "Summary: limited context run"},
			}},
		})
	}))
	defer server.Close()
	if err := SettingSetString(model.SettingKeyAIAutomationBaseURL, server.URL+"/v1"); err != nil {
		t.Fatalf("SettingSetString(base_url) error = %v", err)
	}

	useLocalDefault := false
	task, err := AITaskCreate(model.AITaskCreateRequest{
		Type:      model.AIAutomationTaskTypeNaturalLanguage,
		InputText: "只读取渠道，不要生成 profile",
		ConfigSnapshot: &model.AIAutomationTaskConfig{
			UseLocalDefault: &useLocalDefault,
			ToolKeys:        []string{model.AITaskToolKeyChannelInventory, model.AITaskToolKeySnapshotGuard},
		},
	}, ctx)
	if err != nil {
		t.Fatalf("AITaskCreate() error = %v", err)
	}

	var latest model.AITask
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		latest, err = AITaskGet(task.ID, ctx)
		if err != nil {
			t.Fatalf("AITaskGet() error = %v", err)
		}
		if latest.Status == model.AITaskStatusSucceeded || latest.Status == model.AITaskStatusFailed {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if latest.Status != model.AITaskStatusSucceeded {
		t.Fatalf("task status = %s, error = %s", latest.Status, latest.ErrorMessage)
	}
	if latest.ResultProfileID != nil {
		t.Fatalf("result_profile_id = %#v, want nil when profile_write is disabled", latest.ResultProfileID)
	}

	messages, _ := requestBody["messages"].([]any)
	if len(messages) < 2 {
		t.Fatalf("messages = %#v, want at least system and user", requestBody["messages"])
	}
	userMessage, _ := messages[1].(map[string]any)
	contextText, _ := userMessage["content"].(string)
	if !strings.Contains(contextText, "channel_inventory") {
		t.Fatalf("context text = %q, want tool key list", contextText)
	}
	if !strings.Contains(contextText, "\"channels\"") {
		t.Fatalf("context text = %q, want channels payload", contextText)
	}
	if strings.Contains(contextText, "\n  \"groups\":") {
		t.Fatalf("context text = %q, want groups payload omitted when tool is disabled", contextText)
	}
	if strings.Contains(contextText, "\n  \"models\":") {
		t.Fatalf("context text = %q, want models payload omitted when tool is disabled", contextText)
	}
	if !strings.Contains(latest.ResultJSON, "\"profile_writable\":false") {
		t.Fatalf("result json = %s, want profile_writable false", latest.ResultJSON)
	}
}

func TestAITaskCreateRejectsWhenDisabled(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}

	_, err := AITaskCreate(model.AITaskCreateRequest{Type: model.AIAutomationTaskTypeNaturalLanguage, InputText: "disabled"}, ctx)
	if !errors.Is(err, ErrAIAutomationDisabled) {
		t.Fatalf("AITaskCreate() error = %v, want ErrAIAutomationDisabled", err)
	}
}

func TestAIProfileActivateDoesNotOverwriteManualConfig(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}

	channel := &model.Channel{
		Name:     "ai-profile-manual-channel",
		Type:     transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: "https://example.com/v1", Delay: 100}},
		Model:    "gpt-4o",
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	updated, err := ChannelUpdate(&model.ChannelUpdateRequest{
		ID:        channel.ID,
		KeysToAdd: []model.ChannelKeyAddRequest{{Enabled: true, ChannelKey: "manual-key", AllowedModels: "gpt-4o"}},
	}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate() error = %v", err)
	}
	group := &model.Group{Name: "gpt-4o", Mode: model.GroupModeFailover}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	if err := GroupItemAdd(&model.GroupItem{GroupID: group.ID, ChannelID: updated.ID, ModelName: "gpt-4o", Priority: 7, Weight: 3}, ctx); err != nil {
		t.Fatalf("GroupItemAdd() error = %v", err)
	}
	if err := LLMCreate(model.LLMInfo{Name: "gpt-4o", BillingMode: model.BillingModePerToken, ProbePolicy: model.ProbePolicyConcurrent, ProbeConcurrencyLimit: 2}, ctx); err != nil {
		t.Fatalf("LLMCreate() error = %v", err)
	}
	if _, err := RouteTargetOverrideUpsert(model.RouteTargetOverride{ChannelID: updated.ID, ChannelKeyID: updated.Keys[0].ID, ModelName: "gpt-4o", BillingMode: model.BillingModePerToken, ProbePolicy: model.ProbePolicyConcurrent, ProbeIntervalSeconds: 60, ProbeConcurrencyLimit: 2}, ctx); err != nil {
		t.Fatalf("RouteTargetOverrideUpsert() error = %v", err)
	}

	profile, err := AIProfileCreate(model.AIProfile{Domain: model.AIProfileDomainGrouping, Name: "AI generated grouping", Status: model.AIProfileStatusReady, Confidence: 0.91}, `{"groups":[]}`, ctx)
	if err != nil {
		t.Fatalf("AIProfileCreate() error = %v", err)
	}
	if _, err := AIProfileActivate(profile.ID, ctx); err != nil {
		t.Fatalf("AIProfileActivate() error = %v", err)
	}

	mode, err := SettingGetString(model.SettingKeyConfigSourceMode)
	if err != nil || mode != model.ConfigSourceModeAIProfile {
		t.Fatalf("config_source_mode = %q, err = %v, want ai_profile", mode, err)
	}
	activeID, err := SettingGetInt(model.SettingKeyActiveAIProfileID)
	if err != nil || activeID != profile.ID {
		t.Fatalf("active_ai_profile_id = %d, err = %v, want %d", activeID, err, profile.ID)
	}

	var channelCount, groupCount, groupItemCount, llmCount, overrideCount int64
	db.GetDB().Model(&model.Channel{}).Count(&channelCount)
	db.GetDB().Model(&model.Group{}).Count(&groupCount)
	db.GetDB().Model(&model.GroupItem{}).Count(&groupItemCount)
	db.GetDB().Model(&model.LLMInfo{}).Count(&llmCount)
	db.GetDB().Model(&model.RouteTargetOverride{}).Count(&overrideCount)
	if channelCount != 1 || groupCount != 1 || groupItemCount != 1 || llmCount != 1 || overrideCount != 1 {
		t.Fatalf("manual counts changed: channels=%d groups=%d group_items=%d llms=%d overrides=%d", channelCount, groupCount, groupItemCount, llmCount, overrideCount)
	}
}

func TestAIProfileActivateResetsPreviousActiveProfile(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}

	first, err := AIProfileCreate(model.AIProfile{
		Domain: model.AIProfileDomainGrouping,
		Name:   "first profile",
		Status: model.AIProfileStatusReady,
	}, `{"groups":["first"]}`, ctx)
	if err != nil {
		t.Fatalf("AIProfileCreate(first) error = %v", err)
	}
	second, err := AIProfileCreate(model.AIProfile{
		Domain: model.AIProfileDomainGrouping,
		Name:   "second profile",
		Status: model.AIProfileStatusReady,
	}, `{"groups":["second"]}`, ctx)
	if err != nil {
		t.Fatalf("AIProfileCreate(second) error = %v", err)
	}

	if _, err := AIProfileActivate(first.ID, ctx); err != nil {
		t.Fatalf("AIProfileActivate(first) error = %v", err)
	}
	if _, err := AIProfileActivate(second.ID, ctx); err != nil {
		t.Fatalf("AIProfileActivate(second) error = %v", err)
	}

	firstAfter, err := AIProfileGet(first.ID, ctx)
	if err != nil {
		t.Fatalf("AIProfileGet(first) error = %v", err)
	}
	if firstAfter.Status != model.AIProfileStatusReady {
		t.Fatalf("first profile status = %q, want %q", firstAfter.Status, model.AIProfileStatusReady)
	}

	secondAfter, err := AIProfileGet(second.ID, ctx)
	if err != nil {
		t.Fatalf("AIProfileGet(second) error = %v", err)
	}
	if secondAfter.Status != model.AIProfileStatusActive {
		t.Fatalf("second profile status = %q, want %q", secondAfter.Status, model.AIProfileStatusActive)
	}

	var activeCount int64
	if err := db.GetDB().Model(&model.AIProfile{}).Where("status = ?", model.AIProfileStatusActive).Count(&activeCount).Error; err != nil {
		t.Fatalf("count active profiles error = %v", err)
	}
	if activeCount != 1 {
		t.Fatalf("active profile count = %d, want 1", activeCount)
	}
}

func TestDynamicRouteLearningRecordHonorsEnabledSwitch(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}

	obs := DynamicRouteLearningObservation{ChannelID: 1, KeyID: 2, ModelName: "gpt-4o", Success: true, LatencyMs: 120}
	if err := DynamicRouteLearningRecord(ctx, obs); err != nil {
		t.Fatalf("DynamicRouteLearningRecord(disabled) error = %v", err)
	}
	result, err := DynamicRouteLearningList(ctx)
	if err != nil {
		t.Fatalf("DynamicRouteLearningList() error = %v", err)
	}
	if result.Enabled || len(result.States) != 0 {
		t.Fatalf("learning result when disabled = %#v, want disabled and empty", result)
	}

	if err := SettingSetString(model.SettingKeyDynamicRoutingLearningEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}
	if err := DynamicRouteLearningRecord(ctx, obs); err != nil {
		t.Fatalf("DynamicRouteLearningRecord(enabled) error = %v", err)
	}
	result, err = DynamicRouteLearningList(ctx)
	if err != nil {
		t.Fatalf("DynamicRouteLearningList(enabled) error = %v", err)
	}
	if !result.Enabled || len(result.States) != 1 || result.States[0].SuccessCount != 1 {
		t.Fatalf("learning result when enabled = %#v, want one success", result)
	}

	if err := SettingSetString(model.SettingKeyDynamicRoutingLearningEnabled, "false"); err != nil {
		t.Fatalf("SettingSetString(false) error = %v", err)
	}
	result, err = DynamicRouteLearningList(ctx)
	if err != nil {
		t.Fatalf("DynamicRouteLearningList(disabled again) error = %v", err)
	}
	if result.Enabled || len(result.States) != 1 {
		t.Fatalf("learning result after disabling = %#v, want retained data but disabled", result)
	}
}

func TestAIAutomationConfigUpdateRejectsInvalidFieldWithoutPartialPersist(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}

	initialEnabled, err := SettingGetString(model.SettingKeyAIAutomationEnabled)
	if err != nil {
		t.Fatalf("SettingGetString(enabled) error = %v", err)
	}
	initialBaseURL, err := SettingGetString(model.SettingKeyAIAutomationBaseURL)
	if err != nil {
		t.Fatalf("SettingGetString(base_url) error = %v", err)
	}

	_, err = AIAutomationConfigUpdate(model.AIAutomationConfigUpdateRequest{
		Enabled: ptr(true),
		BaseURL: ptr("ftp://invalid.example.com"),
		Model:   ptr("should-not-apply"),
	}, ctx)
	if err == nil {
		t.Fatal("AIAutomationConfigUpdate() error = nil, want validation failure")
	}
	if !strings.Contains(err.Error(), "absolute http or https URL") {
		t.Fatalf("AIAutomationConfigUpdate() error = %v, want base URL validation failure", err)
	}

	afterEnabled, err := SettingGetString(model.SettingKeyAIAutomationEnabled)
	if err != nil {
		t.Fatalf("SettingGetString(enabled) after error = %v", err)
	}
	if afterEnabled != initialEnabled {
		t.Fatalf("enabled setting = %q, want unchanged %q", afterEnabled, initialEnabled)
	}
	afterBaseURL, err := SettingGetString(model.SettingKeyAIAutomationBaseURL)
	if err != nil {
		t.Fatalf("SettingGetString(base_url) after error = %v", err)
	}
	if afterBaseURL != initialBaseURL {
		t.Fatalf("base_url setting = %q, want unchanged %q", afterBaseURL, initialBaseURL)
	}
}

func TestAIAutomationConfigUpdateRejectsExplicitPrivateBaseURLWithoutPartialPersist(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}

	initialEnabled, err := SettingGetString(model.SettingKeyAIAutomationEnabled)
	if err != nil {
		t.Fatalf("SettingGetString(enabled) error = %v", err)
	}
	initialBaseURL, err := SettingGetString(model.SettingKeyAIAutomationBaseURL)
	if err != nil {
		t.Fatalf("SettingGetString(base_url) error = %v", err)
	}

	_, err = AIAutomationConfigUpdate(model.AIAutomationConfigUpdateRequest{
		Enabled: ptr(true),
		BaseURL: ptr("http://127.0.0.1:9090/v1"),
		Model:   ptr("should-not-apply"),
	}, ctx)
	if err == nil {
		t.Fatal("AIAutomationConfigUpdate() error = nil, want private base URL validation failure")
	}
	if !strings.Contains(err.Error(), "must not target loopback or private IP addresses") {
		t.Fatalf("AIAutomationConfigUpdate() error = %v, want private base URL validation failure", err)
	}

	afterEnabled, err := SettingGetString(model.SettingKeyAIAutomationEnabled)
	if err != nil {
		t.Fatalf("SettingGetString(enabled) after error = %v", err)
	}
	if afterEnabled != initialEnabled {
		t.Fatalf("enabled setting = %q, want unchanged %q", afterEnabled, initialEnabled)
	}
	afterBaseURL, err := SettingGetString(model.SettingKeyAIAutomationBaseURL)
	if err != nil {
		t.Fatalf("SettingGetString(base_url) after error = %v", err)
	}
	if afterBaseURL != initialBaseURL {
		t.Fatalf("base_url setting = %q, want unchanged %q", afterBaseURL, initialBaseURL)
	}
}

func TestAITaskCreateExecutesProtectedActionsWhenAuthorized(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}

	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]any{"content": "Summary: protected actions executed"},
			}},
		})
	}))
	defer server.Close()

	if err := SettingSetString(model.SettingKeyAIAutomationBaseURL, server.URL+"/v1"); err != nil {
		t.Fatalf("SettingSetString(base_url) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationModel, "guarded-model"); err != nil {
		t.Fatalf("SettingSetString(model) error = %v", err)
	}

	useLocalDefault := false
	task, err := AITaskCreate(model.AITaskCreateRequest{
		Type:      model.AIAutomationTaskTypeConfigHealthCheck,
		InputText: "请生成配置巡检结果并执行保护动作",
		ConfigSnapshot: &model.AIAutomationTaskConfig{
			UseLocalDefault: &useLocalDefault,
			ToolKeys: []string{
				model.AITaskToolKeyChannelInventory,
				model.AITaskToolKeyProfileWrite,
				model.AITaskToolKeyProfileActivate,
				model.AITaskToolKeySnapshotGuard,
			},
		},
	}, ctx)
	if err != nil {
		t.Fatalf("AITaskCreate() error = %v", err)
	}

	var latest model.AITask
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		latest, err = AITaskGet(task.ID, ctx)
		if err != nil {
			t.Fatalf("AITaskGet() error = %v", err)
		}
		if latest.Status == model.AITaskStatusSucceeded || latest.Status == model.AITaskStatusFailed {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if latest.Status != model.AITaskStatusSucceeded {
		t.Fatalf("task status = %s, error = %s", latest.Status, latest.ErrorMessage)
	}
	if latest.ResultProfileID == nil || *latest.ResultProfileID <= 0 {
		t.Fatalf("result_profile_id = %#v, want saved profile", latest.ResultProfileID)
	}

	mode, err := SettingGetString(model.SettingKeyConfigSourceMode)
	if err != nil {
		t.Fatalf("SettingGetString(config_source_mode) error = %v", err)
	}
	if mode != model.ConfigSourceModeAIProfile {
		t.Fatalf("config_source_mode = %q, want %q", mode, model.ConfigSourceModeAIProfile)
	}
	activeID, err := SettingGetInt(model.SettingKeyActiveAIProfileID)
	if err != nil {
		t.Fatalf("SettingGetInt(active_ai_profile_id) error = %v", err)
	}
	if activeID != *latest.ResultProfileID {
		t.Fatalf("active_ai_profile_id = %d, want %d", activeID, *latest.ResultProfileID)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(latest.ResultJSON), &result); err != nil {
		t.Fatalf("json.Unmarshal(result_json) error = %v", err)
	}
	toolExecution, ok := result["tool_execution"].(map[string]any)
	if !ok {
		t.Fatalf("tool_execution = %#v, want object", result["tool_execution"])
	}
	protectedActions, ok := toolExecution["protected_actions"].([]any)
	if !ok || len(protectedActions) != 2 {
		t.Fatalf("protected_actions = %#v, want two actions", toolExecution["protected_actions"])
	}
	toolSummary, ok := result["tool_execution_summary"].(map[string]any)
	if !ok {
		t.Fatalf("tool_execution_summary = %#v, want object", result["tool_execution_summary"])
	}
	if toolSummary["protected_actions_executed"] != float64(2) {
		t.Fatalf("protected_actions_executed = %#v, want 2", toolSummary["protected_actions_executed"])
	}

	snapshotDir, err := aiTaskSnapshotDir()
	if err != nil {
		t.Fatalf("aiTaskSnapshotDir() error = %v", err)
	}
	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		t.Fatalf("os.ReadDir(snapshotDir) error = %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("snapshot dir %s is empty, want ai task snapshot", snapshotDir)
	}
	if !strings.Contains(entries[0].Name(), "ai-task-") {
		t.Fatalf("snapshot file = %q, want ai-task-*", entries[0].Name())
	}

	importSnapshotMeta := filepath.Join(filepath.Dir(snapshotDir), importSnapshotLatestFilename)
	if _, err := os.Stat(importSnapshotMeta); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no import snapshot metadata side effect, got err = %v", err)
	}
	profile, err := AIProfileGet(*latest.ResultProfileID, ctx)
	if err != nil {
		t.Fatalf("AIProfileGet() error = %v", err)
	}
	if profile.Status != model.AIProfileStatusActive {
		t.Fatalf("profile.Status = %q, want %q", profile.Status, model.AIProfileStatusActive)
	}
}

func TestAIAutomationFetchModelsRejectsExplicitPrivateBaseURL(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}

	_, err := AIAutomationFetchModels(model.AIModelsFetchRequest{
		BaseURL:         "http://127.0.0.1:9090/v1",
		ChannelType:     "openai",
		UseLocalDefault: boolPtr(false),
	}, ctx)
	if err == nil {
		t.Fatal("AIAutomationFetchModels() error = nil, want private base URL validation failure")
	}
	if !strings.Contains(err.Error(), "must not target loopback or private IP addresses") {
		t.Fatalf("AIAutomationFetchModels() error = %v, want private base URL validation failure", err)
	}
}

func TestAIAutomationFetchModelsRejectsBaseURLWithQueryOrFragment(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}

	_, err := AIAutomationFetchModels(model.AIModelsFetchRequest{
		BaseURL:         "https://models.example/v1?debug=true#frag",
		ChannelType:     "openai",
		UseLocalDefault: boolPtr(false),
	}, ctx)
	if err == nil {
		t.Fatal("AIAutomationFetchModels() error = nil, want query/fragment rejection")
	}
	if !strings.Contains(err.Error(), "must not contain query or fragment") {
		t.Fatalf("AIAutomationFetchModels() error = %v, want query/fragment rejection", err)
	}
}

func TestAIAutomationConfigGetFallsBackWhenActiveProfileUsesForbiddenBaseURL(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationBaseURL, "https://manual.example/v1"); err != nil {
		t.Fatalf("SettingSetString(base_url) error = %v", err)
	}
	profile, err := AIProfileCreate(model.AIProfile{Domain: model.AIProfileDomainGrouping, Name: "forbidden profile", Status: model.AIProfileStatusReady}, `{"config":{"base_url":"http://127.0.0.1:9090/v1","channel_type":"openai","model":"profile-model","use_local_default":false}}`, ctx)
	if err != nil {
		t.Fatalf("AIProfileCreate() error = %v", err)
	}
	if _, err := AIProfileActivate(profile.ID, ctx); err != nil {
		t.Fatalf("AIProfileActivate() error = %v", err)
	}

	config, err := AIAutomationConfigGet(ctx)
	if err != nil {
		t.Fatalf("AIAutomationConfigGet() error = %v", err)
	}
	if config.ConfigSourceMode != model.ConfigSourceModeManual || config.ActiveAIProfileID != 0 {
		t.Fatalf("config source fallback = %#v, want manual mode and cleared active profile", config)
	}
	if config.SourceFallbackReason != "profile_forbidden_base_url" {
		t.Fatalf("config.SourceFallbackReason = %q, want profile_forbidden_base_url", config.SourceFallbackReason)
	}
	if config.BaseURL != "https://manual.example/v1" {
		t.Fatalf("config.BaseURL = %q, want manual fallback", config.BaseURL)
	}
}

func TestAITaskCancelStopsInFlightExecution(t *testing.T) {
	ctx := SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIAutomationModel, "cancel-model"); err != nil {
		t.Fatalf("SettingSetString(model) error = %v", err)
	}

	requestStarted := make(chan struct{}, 1)
	releaseRequest := make(chan struct{})
	released := false
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case requestStarted <- struct{}{}:
		default:
		}
		<-releaseRequest
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]any{"content": "Summary: should not be persisted after cancel"},
			}},
		})
	}))
	t.Cleanup(func() {
		if !released {
			close(releaseRequest)
		}
	})
	defer server.Close()

	if err := SettingSetString(model.SettingKeyAIAutomationBaseURL, server.URL+"/v1"); err != nil {
		t.Fatalf("SettingSetString(base_url) error = %v", err)
	}

	task, err := AITaskCreate(model.AITaskCreateRequest{Type: model.AIAutomationTaskTypeNaturalLanguage, InputText: "cancel me"}, ctx)
	if err != nil {
		t.Fatalf("AITaskCreate() error = %v", err)
	}

	select {
	case <-requestStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("AI request did not start before cancel")
	}

	canceled, err := AITaskCancel(task.ID, ctx)
	if err != nil {
		t.Fatalf("AITaskCancel() error = %v", err)
	}
	if canceled.Status != model.AITaskStatusCanceled {
		t.Fatalf("canceled.Status = %q, want %q", canceled.Status, model.AITaskStatusCanceled)
	}
	close(releaseRequest)
	released = true

	var latest model.AITask
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		latest, err = AITaskGet(task.ID, ctx)
		if err != nil {
			t.Fatalf("AITaskGet() error = %v", err)
		}
		if latest.Status == model.AITaskStatusCanceled {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if latest.Status != model.AITaskStatusCanceled {
		t.Fatalf("latest.Status = %q, want canceled", latest.Status)
	}
	if latest.ResultProfileID != nil {
		t.Fatalf("result_profile_id = %#v, want nil after cancel", latest.ResultProfileID)
	}
	if latest.ResultJSON != "" {
		t.Fatalf("result_json = %q, want empty after cancel", latest.ResultJSON)
	}
	if latest.ErrorMessage != "" {
		t.Fatalf("error_message = %q, want empty after cancel", latest.ErrorMessage)
	}

	stepByKey := make(map[string]model.AITaskStep, len(latest.Steps))
	for _, step := range latest.Steps {
		stepByKey[step.StepKey] = step
	}
	callStep := stepByKey["call_ai"]
	if callStep.Status != model.AITaskStepStatusFailed || !strings.Contains(callStep.Message, "task canceled") {
		t.Fatalf("call_ai step = %#v, want failed task canceled", callStep)
	}
	for _, key := range []string{"parse_output", "generate_profile", "save_result"} {
		if stepByKey[key].Status != model.AITaskStepStatusSkipped {
			t.Fatalf("step %s status = %q, want skipped", key, stepByKey[key].Status)
		}
	}
	profiles, err := AIProfileList(ctx)
	if err != nil {
		t.Fatalf("AIProfileList() error = %v", err)
	}
	if len(profiles) != 0 {
		t.Fatalf("profile count = %d, want 0 after cancel", len(profiles))
	}
}

func TestCancelAllAITasksCancelsRunningContext(t *testing.T) {
	SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}

	baseCtx, cancel := context.WithCancel(context.Background())
	storeAITaskCancel(42, cancel)
	t.Cleanup(func() {
		deleteAITaskCancel(42)
	})

	if err := CancelAllAITasks(); err != nil {
		t.Fatalf("CancelAllAITasks() error = %v", err)
	}
	select {
	case <-baseCtx.Done():
		if !errors.Is(baseCtx.Err(), context.Canceled) {
			t.Fatalf("baseCtx.Err() = %v, want context.Canceled", baseCtx.Err())
		}
	case <-time.After(time.Second):
		t.Fatal("CancelAllAITasks() did not cancel registered task context")
	}
	if _, ok := aiTaskCancelFuncs.Load(42); !ok {
		t.Fatal("cancel function removed unexpectedly before task cleanup")
	}
}

func TestCancelAllAITasksWaitsForStartGroupEntriesWithoutRegisteredCancel(t *testing.T) {
	ctx := SetupOpTestDB(t)

	task := model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, InputText: "pending cleanup", Status: model.AITaskStatusPending}
	if err := db.GetDB().WithContext(ctx).Create(&task).Error; err != nil {
		t.Fatalf("create task error = %v", err)
	}
	steps := defaultAITaskSteps(task.ID)
	if err := db.GetDB().WithContext(ctx).Create(&steps).Error; err != nil {
		t.Fatalf("create steps error = %v", err)
	}

	startCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := make(chan struct{})
	finished := make(chan struct{})
	aiTaskStartGroup.Store(task.ID, struct{}{})
	go func() {
		defer close(finished)
		defer aiTaskStartGroup.Delete(task.ID)
		close(started)
		<-startCtx.Done()
	}()

	<-started
	cancel()
	if err := CancelAllAITasks(); err != nil {
		t.Fatalf("CancelAllAITasks() error = %v", err)
	}

	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("CancelAllAITasks() returned before start-group worker exited")
	}
	if _, ok := aiTaskStartGroup.Load(task.ID); ok {
		t.Fatal("aiTaskStartGroup still contains task after cancellation")
	}
}

func TestCancelAllAITasksReturnsErrorWhenTaskDoesNotStop(t *testing.T) {
	SetupOpTestDB(t)
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}

	originalTimeout := aiTaskCancelWaitTimeout
	originalPoll := aiTaskCancelPollInterval
	aiTaskCancelWaitTimeout = 20 * time.Millisecond
	aiTaskCancelPollInterval = 5 * time.Millisecond
	t.Cleanup(func() {
		aiTaskCancelWaitTimeout = originalTimeout
		aiTaskCancelPollInterval = originalPoll
		aiTaskStartGroup.Delete(77)
		deleteAITaskCancel(77)
	})

	baseCtx, cancel := context.WithCancel(context.Background())
	storeAITaskCancel(77, cancel)
	aiTaskStartGroup.Store(77, struct{}{})

	err := CancelAllAITasks()
	if err == nil {
		t.Fatal("CancelAllAITasks() error = nil, want timeout")
	}
	if !strings.Contains(err.Error(), "77") {
		t.Fatalf("CancelAllAITasks() error = %v, want task id in timeout message", err)
	}
	select {
	case <-baseCtx.Done():
		if !errors.Is(baseCtx.Err(), context.Canceled) {
			t.Fatalf("baseCtx.Err() = %v, want context.Canceled", baseCtx.Err())
		}
	case <-time.After(time.Second):
		t.Fatal("CancelAllAITasks() did not cancel registered task context before timing out")
	}
}

func TestInitCacheRecoversInterruptedAITasks(t *testing.T) {
	ctx := SetupOpTestDB(t)

	useLocalDefault := false
	snapshot := model.AIAutomationTaskConfig{BaseURL: "http://127.0.0.1:1/v1", ChannelType: "openai", Model: "resume-model", UseLocalDefault: &useLocalDefault}
	snapshotRaw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("json.Marshal(snapshot) error = %v", err)
	}

	pendingTask := model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, InputText: "pending", Status: model.AITaskStatusPending, ConfigSnapshotJSON: string(snapshotRaw), ResumeState: model.AITaskResumeStateCollectContext}
	if err := db.GetDB().WithContext(ctx).Create(&pendingTask).Error; err != nil {
		t.Fatalf("create pending task error = %v", err)
	}
	pendingSteps := defaultAITaskSteps(pendingTask.ID)
	if err := db.GetDB().WithContext(ctx).Create(&pendingSteps).Error; err != nil {
		t.Fatalf("create pending steps error = %v", err)
	}

	runningTask := model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, InputText: "running", Status: model.AITaskStatusRunning, Progress: aiTaskProgressCallAI, ConfigSnapshotJSON: string(snapshotRaw), ContextPayloadJSON: `{"counts":{"channels":0,"groups":0,"llms":0,"route_target_overrides":0,"dynamic_learning_states":0}}`, PromptText: "resume prompt", SelectedModel: "resume-model", ResumeState: model.AITaskResumeStateCallAI}
	startedAt := time.Now().Add(-time.Minute)
	runningTask.StartedAt = &startedAt
	if err := db.GetDB().WithContext(ctx).Create(&runningTask).Error; err != nil {
		t.Fatalf("create running task error = %v", err)
	}
	runningSteps := defaultAITaskSteps(runningTask.ID)
	for i := range runningSteps {
		switch runningSteps[i].StepKey {
		case "collect_context", "select_model":
			runningSteps[i].Status = model.AITaskStepStatusSucceeded
			runningSteps[i].StartedAt = &startedAt
			runningSteps[i].FinishedAt = &startedAt
		case "call_ai":
			runningSteps[i].Status = model.AITaskStepStatusRunning
			runningSteps[i].StartedAt = &startedAt
		}
	}
	if err := db.GetDB().WithContext(ctx).Create(&runningSteps).Error; err != nil {
		t.Fatalf("create running steps error = %v", err)
	}

	succeededTask := model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, InputText: "done", Status: model.AITaskStatusSucceeded, Progress: 100}
	finishedAt := time.Now().Add(-30 * time.Second)
	succeededTask.FinishedAt = &finishedAt
	if err := db.GetDB().WithContext(ctx).Create(&succeededTask).Error; err != nil {
		t.Fatalf("create succeeded task error = %v", err)
	}

	brokenTask := model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, InputText: "broken", Status: model.AITaskStatusRunning, ResumeState: model.AITaskResumeStateCallAI}
	if err := db.GetDB().WithContext(ctx).Create(&brokenTask).Error; err != nil {
		t.Fatalf("create broken task error = %v", err)
	}
	brokenSteps := defaultAITaskSteps(brokenTask.ID)
	if err := db.GetDB().WithContext(ctx).Create(&brokenSteps).Error; err != nil {
		t.Fatalf("create broken steps error = %v", err)
	}

	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}

	latestPending, err := AITaskGet(pendingTask.ID, ctx)
	if err != nil {
		t.Fatalf("AITaskGet(pending) error = %v", err)
	}
	if latestPending.Status != model.AITaskStatusRecoverable && latestPending.Status != model.AITaskStatusRunning && latestPending.Status != model.AITaskStatusFailed {
		t.Fatalf("pending task status = %q, want recovery scheduling or execution result", latestPending.Status)
	}
	if latestPending.ResumeState != model.AITaskResumeStateCollectContext {
		t.Fatalf("pending resume_state = %q, want collect_context", latestPending.ResumeState)
	}
	if latestPending.FinishedAt != nil && latestPending.Status != model.AITaskStatusFailed {
		t.Fatal("pending recoverable task unexpectedly has finished_at")
	}

	latestRunning, err := AITaskGet(runningTask.ID, ctx)
	if err != nil {
		t.Fatalf("AITaskGet(running) error = %v", err)
	}
	if latestRunning.Status != model.AITaskStatusRecoverable && latestRunning.Status != model.AITaskStatusRunning && latestRunning.Status != model.AITaskStatusFailed {
		t.Fatalf("running task status = %q, want recovery scheduling or execution result", latestRunning.Status)
	}
	if latestRunning.ResumeState != model.AITaskResumeStateCallAI {
		t.Fatalf("running resume_state = %q, want call_ai", latestRunning.ResumeState)
	}

	latestBroken, err := AITaskGet(brokenTask.ID, ctx)
	if err != nil {
		t.Fatalf("AITaskGet(broken) error = %v", err)
	}
	if latestBroken.Status != model.AITaskStatusFailedUnrecoverable {
		t.Fatalf("broken task status = %q, want failed_unrecoverable", latestBroken.Status)
	}
	if !strings.Contains(latestBroken.ErrorMessage, "missing config snapshot") {
		t.Fatalf("broken task error = %q, want missing config snapshot", latestBroken.ErrorMessage)
	}

	latestSucceeded, err := AITaskGet(succeededTask.ID, ctx)
	if err != nil {
		t.Fatalf("AITaskGet(succeeded) error = %v", err)
	}
	if latestSucceeded.Status != model.AITaskStatusSucceeded {
		t.Fatalf("succeeded task status = %q, want succeeded", latestSucceeded.Status)
	}
}
