package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

func TestDynamicRouteLearningHandlersListAndReset(t *testing.T) {
	setupHandlerTest(t)
	ctx := setupHandlerTestDB(t)
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() error = %v", err)
	}
	if err := op.SettingSetString(model.SettingKeyDynamicRoutingLearningEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}
	if err := op.DynamicRouteLearningRecord(ctx, op.DynamicRouteLearningObservation{ChannelID: 1, KeyID: 2, ModelName: "gpt-4o", Success: true, LatencyMs: 100}); err != nil {
		t.Fatalf("DynamicRouteLearningRecord() error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/dynamic-routing/learning", nil, getDynamicRouteLearning)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var result model.DynamicRouteLearningListResult
	if err := json.Unmarshal(res.Data, &result); err != nil {
		t.Fatalf("json.Unmarshal(result) error = %v", err)
	}
	if !result.Enabled || len(result.States) != 1 {
		t.Fatalf("result = %#v, want enabled with one state", result)
	}

	recorder = performJSONHandlerRequest(t, http.MethodPost, "/api/v1/dynamic-routing/learning/reset", nil, resetDynamicRouteLearning)
	if recorder.Code != http.StatusOK {
		t.Fatalf("reset status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	result, err := op.DynamicRouteLearningList(ctx)
	if err != nil {
		t.Fatalf("DynamicRouteLearningList() error = %v", err)
	}
	if len(result.States) != 0 {
		t.Fatalf("states after reset = %#v, want empty", result.States)
	}
}
