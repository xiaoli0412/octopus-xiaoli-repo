package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

func TestFetchAIAutomationModelsHandlerReturnsFreeRecommendedCandidate(t *testing.T) {
	setupHandlerTest(t)
	ctx := setupHandlerTestDB(t)
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() error = %v", err)
	}
	if err := op.SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}

	if err := op.LLMCreate(model.LLMInfo{Name: "paid-model", BillingMode: model.BillingModePerToken}, ctx); err != nil {
		t.Fatalf("LLMCreate(paid-model) error = %v", err)
	}
	if err := op.LLMCreate(model.LLMInfo{Name: "free-model", BillingMode: model.BillingModeFree}, ctx); err != nil {
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

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/ai/models/fetch", map[string]any{
		"base_url":          server.URL,
		"channel_type":      "openai",
		"use_local_default": false,
	}, fetchAIAutomationModels)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var result model.AIModelsFetchResult
	if err := json.Unmarshal(res.Data, &result); err != nil {
		t.Fatalf("json.Unmarshal(result) error = %v", err)
	}
	if result.SelectedName != "free-model" {
		t.Fatalf("result.SelectedName = %q, want free-model", result.SelectedName)
	}
	if len(result.Candidates) < 2 || !result.Candidates[0].Recommended || result.Candidates[0].Name != "free-model" {
		t.Fatalf("result.Candidates = %#v, want recommended free-model first", result.Candidates)
	}
}

func TestAIAutomationConfigHandlers(t *testing.T) {
	setupHandlerTest(t)
	if err := op.SettingSetString(model.SettingKeyAIAutomationAPIKey, "handler-manual-secret"); err != nil {
		t.Fatalf("SettingSetString(api_key) error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/ai/config", nil, getAIAutomationConfig)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var config model.AIAutomationConfig
	if err := json.Unmarshal(res.Data, &config); err != nil {
		t.Fatalf("json.Unmarshal(config) error = %v", err)
	}
	if config.BaseURL != model.DefaultAIAutomationBaseURL || !config.UseLocalDefault {
		t.Fatalf("config = %#v, want local default endpoint", config)
	}
	if config.APIKey != "[redacted]" || config.ManualConfig.APIKey != "[redacted]" {
		t.Fatalf("config api keys = %#v, want redacted response secrets", config)
	}
	if config.RequestedConfigSourceMode != model.ConfigSourceModeManual || config.RequestedActiveAIProfileID != 0 || config.SourceFallbackReason != "" {
		t.Fatalf("initial source semantics = %#v, want clean manual defaults", config)
	}
	if config.RequestedActiveAIProfile != nil || config.ActiveAIProfile != nil {
		t.Fatalf("initial profile refs = requested %#v active %#v, want nil refs for manual defaults", config.RequestedActiveAIProfile, config.ActiveAIProfile)
	}

	customURL := "https://ai.example.com/v1"
	recorder = performJSONHandlerRequest(t, http.MethodPost, "/api/v1/ai/config", map[string]any{
		"enabled":           true,
		"base_url":          customURL,
		"api_key":           "new-handler-secret",
		"model":             "free-model",
		"use_local_default": false,
	}, updateAIAutomationConfig)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res = decodeHandlerResponse(t, recorder)
	if err := json.Unmarshal(res.Data, &config); err != nil {
		t.Fatalf("json.Unmarshal(updated config) error = %v", err)
	}
	if !config.Enabled || config.BaseURL != customURL || config.Model != "free-model" || config.UseLocalDefault {
		t.Fatalf("updated config = %#v, want custom AI automation config", config)
	}
	if config.APIKey != "[redacted]" || config.ManualConfig.APIKey != "[redacted]" {
		t.Fatalf("updated config api keys = %#v, want redacted response secrets", config)
	}
}

func TestAIAutomationConfigHandlerReturnsExplicitFallbackSemantics(t *testing.T) {
	setupHandlerTest(t)
	_ = setupHandlerTestDB(t)
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() error = %v", err)
	}
	if err := op.SettingSetString(model.SettingKeyAIAutomationBaseURL, "http://127.0.0.1:8080/v1"); err != nil {
		t.Fatalf("SettingSetString(base_url) error = %v", err)
	}
	if err := op.SettingSetString(model.SettingKeyAIAutomationChannelType, "openai-compatible"); err != nil {
		t.Fatalf("SettingSetString(channel_type) error = %v", err)
	}
	if err := op.SettingSetString(model.SettingKeyAIAutomationModel, "manual-model"); err != nil {
		t.Fatalf("SettingSetString(model) error = %v", err)
	}
	if err := op.SettingSetString(model.SettingKeyConfigSourceMode, model.ConfigSourceModeAIProfile); err != nil {
		t.Fatalf("SettingSetString(config_source_mode) error = %v", err)
	}
	if err := op.SettingSetString(model.SettingKeyActiveAIProfileID, "999"); err != nil {
		t.Fatalf("SettingSetString(active_ai_profile_id) error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/ai/config", nil, getAIAutomationConfig)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var config model.AIAutomationConfig
	if err := json.Unmarshal(res.Data, &config); err != nil {
		t.Fatalf("json.Unmarshal(config) error = %v", err)
	}
	if config.RequestedConfigSourceMode != model.ConfigSourceModeAIProfile || config.RequestedActiveAIProfileID != 999 {
		t.Fatalf("requested source = %#v, want preserved ai_profile request", config)
	}
	if config.ConfigSourceMode != model.ConfigSourceModeManual || config.ActiveAIProfileID != 0 {
		t.Fatalf("effective source = %#v, want manual fallback", config)
	}
	if config.SourceFallbackReason != "profile_missing" {
		t.Fatalf("config.SourceFallbackReason = %q, want profile_missing", config.SourceFallbackReason)
	}
	if config.RequestedActiveAIProfile != nil || config.ActiveAIProfile != nil {
		t.Fatalf("fallback profile refs = requested %#v active %#v, want nil refs when requested profile is missing", config.RequestedActiveAIProfile, config.ActiveAIProfile)
	}
}

func TestAITaskHandlerCreatesProgressSteps(t *testing.T) {
	setupHandlerTest(t)
	if err := op.SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/ai/tasks", map[string]any{
		"type":       model.AIAutomationTaskTypeGroupSuggestion,
		"input_text": "帮我整理分组",
	}, createAITask)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var task model.AITask
	if err := json.Unmarshal(res.Data, &task); err != nil {
		t.Fatalf("json.Unmarshal(task) error = %v", err)
	}
	if task.ID == 0 || task.Status != model.AITaskStatusPending || len(task.Steps) != 6 {
		t.Fatalf("task = %#v, want pending task with six progress steps", task)
	}
}

func TestAITaskHandlerRejectsWhenDisabled(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/ai/tasks", map[string]any{
		"type":       model.AIAutomationTaskTypeGroupSuggestion,
		"input_text": "甯垜鏁寸悊鍒嗙粍",
	}, createAITask)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusForbidden, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if res.Message != op.ErrAIAutomationDisabled.Error() {
		t.Fatalf("message = %q, want %q", res.Message, op.ErrAIAutomationDisabled.Error())
	}
}

func TestAITaskHandlerEventuallyReturnsSavedProfile(t *testing.T) {
	setupHandlerTest(t)
	if err := op.SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}

	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]any{"content": "Summary: generated profile from handler test"},
			}},
		})
	}))
	defer server.Close()
	if err := op.SettingSetString(model.SettingKeyAIAutomationBaseURL, server.URL); err != nil {
		t.Fatalf("SettingSetString(base_url) error = %v", err)
	}
	if err := op.SettingSetString(model.SettingKeyAIAutomationModel, "free-model"); err != nil {
		t.Fatalf("SettingSetString(model) error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/ai/tasks", map[string]any{
		"type":       model.AIAutomationTaskTypeNaturalLanguage,
		"input_text": "帮我整理分组",
	}, createAITask)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var created model.AITask
	if err := json.Unmarshal(res.Data, &created); err != nil {
		t.Fatalf("json.Unmarshal(created) error = %v", err)
	}

	var polled model.AITask
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		recorder = performParamHandlerRequest(t, http.MethodGet, "/api/v1/ai/tasks/1", nil, map[string]string{"id": "1"}, getAITask)
		if recorder.Code != http.StatusOK {
			t.Fatalf("poll status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
		}
		res = decodeHandlerResponse(t, recorder)
		if err := json.Unmarshal(res.Data, &polled); err != nil {
			t.Fatalf("json.Unmarshal(polled) error = %v", err)
		}
		if polled.Status == model.AITaskStatusSucceeded || polled.Status == model.AITaskStatusFailed {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if polled.Status != model.AITaskStatusSucceeded {
		t.Fatalf("polled status = %s, error = %s", polled.Status, polled.ErrorMessage)
	}
	if polled.ResultProfileID == nil || *polled.ResultProfileID <= 0 {
		t.Fatalf("result_profile_id = %#v, want saved profile", polled.ResultProfileID)
	}
	if polled.ID != created.ID {
		t.Fatalf("polled ID = %d, want %d", polled.ID, created.ID)
	}
}

func TestCancelAITaskHandlerMarksTaskCanceled(t *testing.T) {
	setupHandlerTest(t)
	if err := op.SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/ai/tasks", map[string]any{
		"type":       model.AIAutomationTaskTypeGroupSuggestion,
		"input_text": "帮我整理分组",
	}, createAITask)
	if recorder.Code != http.StatusOK {
		t.Fatalf("create status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	recorder = performParamHandlerRequest(t, http.MethodPost, "/api/v1/ai/tasks/1/cancel", nil, map[string]string{"id": "1"}, cancelAITask)
	if recorder.Code != http.StatusOK {
		t.Fatalf("cancel status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var task model.AITask
	if err := json.Unmarshal(res.Data, &task); err != nil {
		t.Fatalf("json.Unmarshal(task) error = %v", err)
	}
	if task.Status != model.AITaskStatusCanceled {
		t.Fatalf("task.Status = %q, want %q", task.Status, model.AITaskStatusCanceled)
	}
	if task.Progress != 100 {
		t.Fatalf("task.Progress = %d, want 100", task.Progress)
	}
	if task.FinishedAt == nil {
		t.Fatal("task.FinishedAt is nil, want cancel timestamp")
	}
	if len(task.Steps) != 6 {
		t.Fatalf("task.Steps = %d, want 6", len(task.Steps))
	}
	for _, step := range task.Steps {
		if step.Status == model.AITaskStepStatusRunning {
			t.Fatalf("step %s status = %q, want not running after cancel", step.StepKey, step.Status)
		}
	}
}

func TestRetryAITaskHandlerCreatesNewTask(t *testing.T) {
	setupHandlerTest(t)
	if err := op.SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}
	useLocalDefault := false
	snapshotRaw, err := json.Marshal(model.AIAutomationTaskConfig{BaseURL: "https://handler-retry.example/v1", ChannelType: "openai", Model: "handler-retry-model", UseLocalDefault: &useLocalDefault})
	if err != nil {
		t.Fatalf("json.Marshal(snapshot) error = %v", err)
	}
	original := model.AITask{Type: model.AIAutomationTaskTypeGroupSuggestion, InputText: "retry from handler", ContextScope: "history", PromptTemplateIDs: "4,5", CustomPrompt: "retry handler prompt", Status: model.AITaskStatusSucceeded, ConfigSnapshotJSON: string(snapshotRaw)}
	if err := db.GetDB().Create(&original).Error; err != nil {
		t.Fatalf("create original task error = %v", err)
	}
	recorder := performParamHandlerRequest(t, http.MethodPost, "/api/v1/ai/tasks/1/retry", nil, map[string]string{"id": "1"}, retryAITask)
	if recorder.Code != http.StatusOK {
		t.Fatalf("retry status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var task model.AITask
	if err := json.Unmarshal(res.Data, &task); err != nil {
		t.Fatalf("json.Unmarshal(task) error = %v", err)
	}
	if task.ID == original.ID || task.Status != model.AITaskStatusPending || task.Type != original.Type || task.InputText != original.InputText || len(task.Steps) != 6 {
		t.Fatalf("retried task = %#v, want new pending task with preserved inputs", task)
	}
	if task.PromptTemplateIDs != "4,5" {
		t.Fatalf("task.PromptTemplateIDs = %q, want 4,5", task.PromptTemplateIDs)
	}
}

func TestGetAITaskArtifactsHandlerReturnsParsedPayloads(t *testing.T) {
	setupHandlerTest(t)
	useLocalDefault := false
	snapshotRaw, err := json.Marshal(model.AIAutomationTaskConfig{BaseURL: "https://handler-artifacts.example/v1", APIKey: "handler-secret-key", ChannelType: "anthropic", Model: "artifact-model", UseLocalDefault: &useLocalDefault})
	if err != nil {
		t.Fatalf("json.Marshal(snapshot) error = %v", err)
	}
	task := model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, Status: model.AITaskStatusSucceeded, ConfigSnapshotJSON: string(snapshotRaw), ContextPayloadJSON: `{"counts":{"channels":2},"scope":"handler"}`, PromptText: "handler prompt", SelectedModel: "artifact-model", ModelReason: "history", ResumeState: model.AITaskResumeStateParse, ResultJSON: `{"summary":"handler ok"}`}
	if err := db.GetDB().Create(&task).Error; err != nil {
		t.Fatalf("create task error = %v", err)
	}
	step := model.AITaskStep{TaskID: task.ID, StepKey: "collect_context", Name: "收集上下文", Status: model.AITaskStepStatusSucceeded, OutputJSON: `{"counts":{"channels":2}}`, SortOrder: 1}
	if err := db.GetDB().Create(&step).Error; err != nil {
		t.Fatalf("create step error = %v", err)
	}
	recorder := performParamHandlerRequest(t, http.MethodGet, "/api/v1/ai/tasks/1/artifacts", nil, map[string]string{"id": "1"}, getAITaskArtifacts)
	if recorder.Code != http.StatusOK {
		t.Fatalf("artifacts status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var artifacts model.AITaskArtifacts
	if err := json.Unmarshal(res.Data, &artifacts); err != nil {
		t.Fatalf("json.Unmarshal(artifacts) error = %v", err)
	}
	if artifacts.TaskID != task.ID || artifacts.ConfigSnapshot == nil || artifacts.ConfigSnapshot.BaseURL != "https://handler-artifacts.example/v1" || artifacts.PromptText != "handler prompt" || len(artifacts.Steps) != 1 {
		t.Fatalf("artifacts = %#v, want parsed artifacts payload", artifacts)
	}
	if artifacts.ConfigSnapshot.APIKey != "[redacted]" {
		t.Fatalf("artifacts.ConfigSnapshot.APIKey = %q, want [redacted]", artifacts.ConfigSnapshot.APIKey)
	}
	if strings.Contains(artifacts.ConfigSnapshotJSON, "handler-secret-key") {
		t.Fatalf("ConfigSnapshotJSON leaked API key: %s", artifacts.ConfigSnapshotJSON)
	}
	contextPayload, ok := artifacts.ContextPayload.(map[string]any)
	if !ok || contextPayload["scope"] != "handler" {
		t.Fatalf("context payload = %#v, want parsed handler payload", artifacts.ContextPayload)
	}
	resultPayload, ok := artifacts.ResultPayload.(map[string]any)
	if !ok || resultPayload["summary"] != "handler ok" {
		t.Fatalf("result payload = %#v, want parsed handler result", artifacts.ResultPayload)
	}
}

func TestCreateAIPromptTemplateHandlerRejectsInvalidTaskType(t *testing.T) {
	setupHandlerTest(t)
	if err := op.SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/ai/prompt-templates", map[string]any{
		"name":      "bad-template",
		"task_type": "not-a-real-task",
		"prompt":    "help me",
	}, createAIPromptTemplate)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
}

func TestGetAIProfileHandlerReturnsVersions(t *testing.T) {
	setupHandlerTest(t)
	ctx := setupHandlerTestDB(t)
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() error = %v", err)
	}
	profile, err := op.AIProfileCreate(model.AIProfile{
		Domain:      model.AIProfileDomainGrouping,
		Name:        "profile-with-version",
		Status:      model.AIProfileStatusReady,
		Explanation: "preview me",
	}, `{"config":{"api_key":"handler-profile-secret"},"domain_payload":{"typed_config":{"api_key":"handler-profile-secret"},"findings":[],"recommendations":[],"risks":[]},"groups":[{"name":"gpt-4o"}]}`, ctx)
	if err != nil {
		t.Fatalf("AIProfileCreate() error = %v", err)
	}

	recorder := performParamHandlerRequest(t, http.MethodGet, "/api/v1/ai/profiles/1", nil, map[string]string{"id": "1"}, getAIProfile)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var fetched model.AIProfile
	if err := json.Unmarshal(res.Data, &fetched); err != nil {
		t.Fatalf("json.Unmarshal(profile) error = %v", err)
	}
	if fetched.ID != profile.ID {
		t.Fatalf("fetched.ID = %d, want %d", fetched.ID, profile.ID)
	}
	if len(fetched.Versions) != 1 {
		t.Fatalf("len(fetched.Versions) = %d, want 1", len(fetched.Versions))
	}
	if fetched.Versions[0].ContentJSON == "" {
		t.Fatal("fetched version content_json is empty")
	}
	if strings.Contains(fetched.Versions[0].ContentJSON, "handler-profile-secret") {
		t.Fatalf("ContentJSON leaked API key: %s", fetched.Versions[0].ContentJSON)
	}
}

func TestAIProfileActivateHandlerSwitchesSettingsOnly(t *testing.T) {
	setupHandlerTest(t)
	ctx := setupHandlerTestDB(t)
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() error = %v", err)
	}
	profile, err := op.AIProfileCreate(model.AIProfile{Domain: model.AIProfileDomainGrouping, Name: "handler profile", Status: model.AIProfileStatusReady}, `{"groups":[]}`, ctx)
	if err != nil {
		t.Fatalf("AIProfileCreate() error = %v", err)
	}

	recorder := performParamHandlerRequest(t, http.MethodPost, "/api/v1/ai/profiles/1/activate", nil, map[string]string{"id": "1"}, activateAIProfile)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	activeID, err := op.SettingGetInt(model.SettingKeyActiveAIProfileID)
	if err != nil || activeID != profile.ID {
		t.Fatalf("active profile id = %d, err = %v, want %d", activeID, err, profile.ID)
	}
}
