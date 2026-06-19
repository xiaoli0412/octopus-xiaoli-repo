package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
)

func TestTestModelConnectivityHandler(t *testing.T) {
	setupHandlerTest(t)

	channel := &model.Channel{
		Name:     "handler-test-connectivity",
		Enabled:  true,
		Type:     outbound.OutboundTypeOpenAIChat,
		BaseUrls: []model.BaseUrl{{URL: "http://localhost:1"}},
		Model:    "gpt-4o",
		Keys: []model.ChannelKey{{
			Enabled:    true,
			ChannelKey: "sk-test",
		}},
	}
	if err := op.ChannelCreate(channel, nil); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/model/test", map[string]any{
		"channel_id": channel.ID,
		"model":      "gpt-4o",
	}, testModelConnectivity)
	if recorder.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	var result model.ModelTestResult
	if err := json.Unmarshal(response.Data, &result); err != nil {
		t.Fatalf("json.Unmarshal(data) error = %v, data=%s", err, string(response.Data))
	}
	if result.Success {
		t.Fatalf("expected connectivity failure for unreachable upstream, got success")
	}
}

func TestTestModelConnectivityHandlerValidation(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/model/test", map[string]any{
		"model": "gpt-4o",
	}, testModelConnectivity)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("response code = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestGetCapabilityInventoryReturnsServiceableFields(t *testing.T) {
	setupHandlerTest(t)

	channel := &model.Channel{
		Name:    "handler-capability-inventory",
		Enabled: true,
		Type:    outbound.OutboundTypeOpenAIChat,
		Keys: []model.ChannelKey{{
			Enabled:             true,
			ChannelKey:          "sk-handler-capability",
			AllowedModels:       "gpt-4o",
			RequestCapabilities: model.RequestCapabilityOpenAIChat,
		}},
	}
	if err := op.ChannelCreate(channel, nil); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/model/capability-inventory", nil, getCapabilityInventory)
	if recorder.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	var inventory model.CapabilityInventory
	if err := json.Unmarshal(response.Data, &inventory); err != nil {
		t.Fatalf("json.Unmarshal(data) error = %v, data=%s", err, string(response.Data))
	}

	if len(inventory.ServiceableModels) != 1 {
		t.Fatalf("serviceable models = %#v, want one item", inventory.ServiceableModels)
	}
	serviceable := inventory.ServiceableModels[0]
	if serviceable.Name != "gpt-4o" || serviceable.KeyCount != 1 || serviceable.InventorySource != "channel_key_allowed" {
		t.Fatalf("serviceable item = %#v, want gpt-4o key_count=1 source=channel_key_allowed", serviceable)
	}
	if len(serviceable.RequestCapabilities) != 1 || serviceable.RequestCapabilities[0] != model.RequestCapabilityOpenAIChat {
		t.Fatalf("serviceable request capabilities = %#v, want openai_chat", serviceable.RequestCapabilities)
	}

	if len(inventory.SelectableModels) != 1 {
		t.Fatalf("selectable models = %#v, want one item", inventory.SelectableModels)
	}
	selectable := inventory.SelectableModels[0]
	if selectable.Name != "gpt-4o" || selectable.KeyCount != 1 || selectable.ChannelCount != 1 {
		t.Fatalf("selectable item = %#v, want gpt-4o key_count=1 channel_count=1", selectable)
	}
}

func createTestLLMForHandler(t *testing.T, name string) {
	t.Helper()
	if err := op.LLMCreate(model.LLMInfo{
		Name:        name,
		BillingMode: model.BillingModePerToken,
		ProbePolicy: model.ProbePolicyConcurrent,
		CachePolicy: model.CachePolicyUnsupported,
	}, nil); err != nil {
		t.Fatalf("LLMCreate(%q) error = %v", name, err)
	}
}

func TestDisableLLMHandler(t *testing.T) {
	setupHandlerTest(t)
	createTestLLMForHandler(t, "handler-disable-model")

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/model/disable", map[string]any{
		"name": "handler-disable-model",
	}, disableLLM)
	if recorder.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	disabled, err := op.IsModelDisabled(nil, "handler-disable-model")
	if err != nil {
		t.Fatalf("IsModelDisabled() error = %v", err)
	}
	if !disabled {
		t.Fatalf("IsModelDisabled() = %v, want true", disabled)
	}
}

func TestEnableLLMHandler(t *testing.T) {
	setupHandlerTest(t)
	createTestLLMForHandler(t, "handler-enable-model")
	if err := op.DisableModel(nil, "handler-enable-model"); err != nil {
		t.Fatalf("DisableModel() error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/model/enable", map[string]any{
		"name": "handler-enable-model",
	}, enableLLM)
	if recorder.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	disabled, err := op.IsModelDisabled(nil, "handler-enable-model")
	if err != nil {
		t.Fatalf("IsModelDisabled() error = %v", err)
	}
	if disabled {
		t.Fatalf("IsModelDisabled() = %v, want false", disabled)
	}
}

func TestListDisabledLLMsHandler(t *testing.T) {
	setupHandlerTest(t)
	createTestLLMForHandler(t, "handler-list-disabled-model")
	if err := op.DisableModel(nil, "handler-list-disabled-model"); err != nil {
		t.Fatalf("DisableModel() error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/model/disabled", nil, listDisabledLLMs)
	if recorder.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	var names []string
	if err := json.Unmarshal(response.Data, &names); err != nil {
		t.Fatalf("json.Unmarshal(data) error = %v, data=%s", err, string(response.Data))
	}
	if len(names) != 1 || names[0] != "handler-list-disabled-model" {
		t.Fatalf("disabled models = %#v, want [handler-list-disabled-model]", names)
	}
}
