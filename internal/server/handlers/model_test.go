package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
)

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
