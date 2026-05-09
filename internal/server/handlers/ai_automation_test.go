package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

func seedGovernanceRoutingFixture(t *testing.T) {
	t.Helper()
	ctx := setupHandlerTestDB(t)
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() error = %v", err)
	}
	if err := op.SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}
	if err := op.SettingSetString(model.SettingKeyAIGovernanceManagedGroupName, "AI Governance Managed"); err != nil {
		t.Fatalf("SettingSetString(managed_group) error = %v", err)
	}
	if err := op.SettingSetString(model.SettingKeyDynamicRoutingLearningEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(dynamic_routing_learning_enabled) error = %v", err)
	}
	channel := &model.Channel{Name: "governance-openai", Enabled: true, Model: "gpt-4o,gpt-4.1"}
	if err := op.ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	key := model.ChannelKey{ChannelID: channel.ID, Enabled: true, ChannelKey: "k-test", AllowedModels: "gpt-4o,gpt-4.1"}
	if err := db.GetDB().WithContext(ctx).Create(&key).Error; err != nil {
		t.Fatalf("create key error = %v", err)
	}
	if err := op.InitCache(); err != nil {
		t.Fatalf("InitCache() after key seed error = %v", err)
	}
	if err := op.LLMCreate(model.LLMInfo{Name: "gpt-4o", BillingMode: model.BillingModePerToken, ProbePolicy: model.ProbePolicyConcurrent, ProbeConcurrencyLimit: 2}, ctx); err != nil {
		t.Fatalf("LLMCreate(gpt-4o) error = %v", err)
	}
	if err := op.LLMCreate(model.LLMInfo{Name: "gpt-4.1", BillingMode: model.BillingModePerToken, ProbePolicy: model.ProbePolicySequential, ProbeConcurrencyLimit: 1}, ctx); err != nil {
		t.Fatalf("LLMCreate(gpt-4.1) error = %v", err)
	}
	group := &model.Group{Name: "legacy-group", Mode: model.GroupModeRoundRobin}
	if err := op.GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	if err := op.GroupItemAdd(&model.GroupItem{GroupID: group.ID, ChannelID: channel.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1}, ctx); err != nil {
		t.Fatalf("GroupItemAdd() error = %v", err)
	}
	if err := op.DynamicRouteLearningRecord(ctx, op.DynamicRouteLearningObservation{ChannelID: channel.ID, KeyID: key.ID, ModelName: "gpt-4o", Success: true, LatencyMs: 90}); err != nil {
		t.Fatalf("DynamicRouteLearningRecord() error = %v", err)
	}
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() after seed error = %v", err)
	}
}

func TestAIGovernanceOverviewHandler(t *testing.T) {
	setupHandlerTest(t)
	seedGovernanceRoutingFixture(t)

	recorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/ai/overview", nil, getAIGovernanceOverview)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var overview model.AIGovernanceOverview
	if err := json.Unmarshal(res.Data, &overview); err != nil {
		t.Fatalf("json.Unmarshal(overview) error = %v", err)
	}
	if !overview.Enabled || overview.ManagedGroupName != "AI Governance Managed" {
		t.Fatalf("overview = %#v, want enabled managed group summary", overview)
	}
	if overview.Learning.SampleCount == 0 {
		t.Fatalf("overview learning = %#v, want sample count > 0", overview.Learning)
	}
}

func TestGovernanceSessionLifecycleHandlers(t *testing.T) {
	setupHandlerTest(t)
	seedGovernanceRoutingFixture(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/ai/sessions", map[string]any{
		"goal":             "整理路由与分组",
		"expert_preset_id": model.GovernanceExpertPresetBalanced,
	}, createGovernanceSession)
	if recorder.Code != http.StatusOK {
		t.Fatalf("create status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var detail model.GovernanceSessionDetail
	if err := json.Unmarshal(res.Data, &detail); err != nil {
		t.Fatalf("json.Unmarshal(detail) error = %v", err)
	}
	if detail.ID == 0 || detail.Status != model.GovernanceSessionStatusReady {
		t.Fatalf("detail = %#v, want ready governance session", detail)
	}
	if !detail.Preview.CanApply || detail.Preview.MutationCount == 0 {
		t.Fatalf("preview = %#v, want applyable typed mutation plan", detail.Preview)
	}

	listRecorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/ai/sessions", nil, listGovernanceSessions)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d, body = %s", listRecorder.Code, http.StatusOK, listRecorder.Body.String())
	}

	getRecorder := performParamHandlerRequest(t, http.MethodGet, "/api/v1/ai/sessions/1", nil, map[string]string{"id": "1"}, getGovernanceSession)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d, body = %s", getRecorder.Code, http.StatusOK, getRecorder.Body.String())
	}

	applyRecorder := performParamHandlerRequest(t, http.MethodPost, "/api/v1/ai/sessions/1/apply", map[string]any{}, map[string]string{"id": "1"}, applyGovernanceSession)
	if applyRecorder.Code != http.StatusOK {
		t.Fatalf("apply status = %d, want %d, body = %s", applyRecorder.Code, http.StatusOK, applyRecorder.Body.String())
	}
	applyRes := decodeHandlerResponse(t, applyRecorder)
	if err := json.Unmarshal(applyRes.Data, &detail); err != nil {
		t.Fatalf("json.Unmarshal(applied detail) error = %v", err)
	}
	if detail.Status != model.GovernanceSessionStatusApplied {
		t.Fatalf("applied detail = %#v, want applied session", detail)
	}
	if len(detail.ApplyRuns) == 0 || detail.ApplyRuns[0].Status != model.GovernanceApplyRunStatusSucceeded {
		t.Fatalf("apply runs = %#v, want succeeded apply run", detail.ApplyRuns)
	}

	runsRecorder := performParamHandlerRequest(t, http.MethodGet, "/api/v1/ai/sessions/1/apply-runs", nil, map[string]string{"id": "1"}, listGovernanceApplyRuns)
	if runsRecorder.Code != http.StatusOK {
		t.Fatalf("apply runs status = %d, want %d, body = %s", runsRecorder.Code, http.StatusOK, runsRecorder.Body.String())
	}
}

func TestGovernanceStrategyProfileHandlers(t *testing.T) {
	setupHandlerTest(t)
	seedGovernanceRoutingFixture(t)
	performJSONHandlerRequest(t, http.MethodPost, "/api/v1/ai/sessions", map[string]any{"goal": "整理路由与分组"}, createGovernanceSession)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/ai/strategy-profiles", map[string]any{
		"session_id": 1,
		"name":       "Managed routing baseline",
	}, createGovernanceStrategyProfile)
	if recorder.Code != http.StatusOK {
		t.Fatalf("create strategy profile status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var profile model.StrategyProfileSummary
	if err := json.Unmarshal(res.Data, &profile); err != nil {
		t.Fatalf("json.Unmarshal(profile) error = %v", err)
	}
	if profile.ID == 0 || profile.Name != "Managed routing baseline" {
		t.Fatalf("profile = %#v, want created strategy profile", profile)
	}

	activateRecorder := performParamHandlerRequest(t, http.MethodPost, "/api/v1/ai/strategy-profiles/1/activate", map[string]any{}, map[string]string{"id": "1"}, activateGovernanceStrategyProfile)
	if activateRecorder.Code != http.StatusOK {
		t.Fatalf("activate status = %d, want %d, body = %s", activateRecorder.Code, http.StatusOK, activateRecorder.Body.String())
	}
	listRecorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/ai/strategy-profiles", nil, listGovernanceStrategyProfiles)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("list strategy profiles status = %d, want %d, body = %s", listRecorder.Code, http.StatusOK, listRecorder.Body.String())
	}

	presetsRecorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/ai/expert-presets", nil, listGovernanceExpertPresets)
	if presetsRecorder.Code != http.StatusOK {
		t.Fatalf("expert presets status = %d, want %d, body = %s", presetsRecorder.Code, http.StatusOK, presetsRecorder.Body.String())
	}
}

func TestAIAutomationLegacyEndpointsReturnGone(t *testing.T) {
	setupHandlerTest(t)
	recorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/ai/config", nil, goneAIAutomationV1)
	if recorder.Code != http.StatusGone {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusGone, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if response.Message != aiAutomationLegacyGoneMessage {
		t.Fatalf("message = %q, want %q", response.Message, aiAutomationLegacyGoneMessage)
	}
}
