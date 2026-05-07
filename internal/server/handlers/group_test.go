package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	transformerOutbound "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
)

func TestCreateGroupReturnsBadRequestForInvalidRuntimeConfig(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/group/create", map[string]any{
		"name":                "group-invalid-runtime",
		"mode":                int(model.GroupModeFailover),
		"race_concurrency":    2,
		"race_after_fails":    0,
		"failover_window_sec": 360,
	}, createGroup)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
}

func TestUpdateGroupReturnsNotFoundForMissingGroup(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/group/update", map[string]any{
		"id":   999999,
		"name": "missing-group",
	}, updateGroup)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}

func TestDeleteGroupReturnsNotFoundForMissingGroup(t *testing.T) {
	setupHandlerTest(t)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/group/delete/999999", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "999999"}}
	deleteGroup(ctx)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}

func TestDeleteGroupRejectsNonPositivePathIDs(t *testing.T) {
	setupHandlerTest(t)

	for _, id := range []string{"0", "-1"} {
		recorder := performParamHandlerRequest(t, http.MethodDelete, "/api/v1/group/delete/"+id, nil, map[string]string{"id": id}, deleteGroup)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("id %s status = %d, want %d, body = %s", id, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
		res := decodeHandlerResponse(t, recorder)
		if res.Message != "Invalid parameter" {
			t.Fatalf("id %s message = %q, want %q", id, res.Message, "Invalid parameter")
		}
	}
}

func TestCreateGroupRejectsBlankName(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/group/create", map[string]any{
		"name": "   ",
		"mode": int(model.GroupModeRoundRobin),
	}, createGroup)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
}

func TestCreateGroupTrimsNameBeforePersisting(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/group/create", map[string]any{
		"name": "  trimmed-group  ",
		"mode": int(model.GroupModeRoundRobin),
	}, createGroup)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	group, err := model.Group{}, error(nil)
	_ = err
	if err := json.Unmarshal(decodeHandlerResponse(t, recorder).Data, &group); err != nil {
		t.Fatalf("json.Unmarshal(group) error = %v", err)
	}
	if group.Name != "trimmed-group" {
		t.Fatalf("group.Name = %q, want %q", group.Name, "trimmed-group")
	}
}

func TestCreateGroupPersistsSubmittedItems(t *testing.T) {
	setupHandlerTest(t)
	ctx := setupHandlerTestDB(t)
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() error = %v", err)
	}

	channel := &model.Channel{
		Name:              "handler-group-create-items-channel",
		Type:              transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: model.KeyManagementModeClassified,
		BaseUrls:          []model.BaseUrl{{URL: "https://example.com/v1", Delay: 0}},
		Model:             "gpt-4o",
	}
	if err := op.ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	updated, err := op.ChannelUpdate(&model.ChannelUpdateRequest{
		ID: channel.ID,
		KeysToAdd: []model.ChannelKeyAddRequest{{
			Enabled:       true,
			ChannelKey:    "handler-group-create-items-key",
			SourceType:    "paid/metered",
			AllowedModels: "gpt-4o",
		}},
	}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate() error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/group/create", map[string]any{
		"name": "group-create-with-items",
		"mode": int(model.GroupModeRoundRobin),
		"items": []map[string]any{{
			"channel_id": updated.ID,
			"model_name": "gpt-4o",
			"priority":   1,
			"weight":     2,
		}},
	}, createGroup)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var group model.Group
	if err := json.Unmarshal(decodeHandlerResponse(t, recorder).Data, &group); err != nil {
		t.Fatalf("json.Unmarshal(group) error = %v", err)
	}
	if len(group.Items) != 1 {
		t.Fatalf("group.Items = %#v, want one persisted item", group.Items)
	}
	if group.Items[0].ID <= 0 || group.Items[0].GroupID != group.ID {
		t.Fatalf("group.Items[0] = %#v, want persisted item identity", group.Items[0])
	}
	if group.Items[0].ChannelID != updated.ID || group.Items[0].ModelName != "gpt-4o" {
		t.Fatalf("group.Items[0] = %#v, want submitted channel/model", group.Items[0])
	}
}

func TestCreateGroupRejectsSubmittedItemsReferencingMissingChannel(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/group/create", map[string]any{
		"name": "group-create-invalid-items",
		"mode": int(model.GroupModeRoundRobin),
		"items": []map[string]any{{
			"channel_id": 99999,
			"model_name": "gpt-4o",
			"priority":   1,
			"weight":     1,
		}},
	}, createGroup)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}
