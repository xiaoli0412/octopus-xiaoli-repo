package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	transformerOutbound "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
)

func TestUpsertRouteTargetOverrideSucceeds(t *testing.T) {
	setupHandlerTest(t)
	ctx := setupHandlerTestDB(t)
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() error = %v", err)
	}

	channel := &model.Channel{
		Name:              "handler-route-target-channel",
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
			ChannelKey:    "handler-route-target-key",
			SourceType:    "paid/metered",
			AllowedModels: "gpt-4o",
		}},
	}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate() error = %v", err)
	}
	if len(updated.Keys) != 1 {
		t.Fatalf("updated.Keys = %#v, want one key", updated.Keys)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/route-target/upsert", map[string]any{
		"channel_id":              updated.ID,
		"channel_key_id":          updated.Keys[0].ID,
		"model_name":              "gpt-4o",
		"billing_mode":            "per_request",
		"probe_policy":            "concurrent",
		"probe_interval_seconds":  60,
		"probe_concurrency_limit": 3,
	}, upsertRouteTargetOverride)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var row model.RouteTargetOverride
	if err := json.Unmarshal(res.Data, &row); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if row.ChannelID != updated.ID || row.ChannelKeyID != updated.Keys[0].ID || row.ModelName != "gpt-4o" {
		t.Fatalf("route target override = %#v, want persisted target identity", row)
	}
	if row.BillingMode != model.BillingModePerRequest || row.ProbePolicy != model.ProbePolicyConcurrent {
		t.Fatalf("route target override = %#v, want policy fields persisted", row)
	}
}

func TestUpsertRouteTargetOverrideRejectsForeignChannelKey(t *testing.T) {
	setupHandlerTest(t)
	ctx := setupHandlerTestDB(t)
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() error = %v", err)
	}

	createChannelWithKey := func(name, keyValue string) *model.Channel {
		channel := &model.Channel{
			Name:              name,
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
				ChannelKey:    keyValue,
				SourceType:    "paid/metered",
				AllowedModels: "gpt-4o",
			}},
		}, ctx)
		if err != nil {
			t.Fatalf("ChannelUpdate() error = %v", err)
		}
		return updated
	}

	updatedA := createChannelWithKey("handler-route-target-foreign-a", "handler-route-target-foreign-key-a")
	updatedB := createChannelWithKey("handler-route-target-foreign-b", "handler-route-target-foreign-key-b")

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/route-target/upsert", map[string]any{
		"channel_id":              updatedA.ID,
		"channel_key_id":          updatedB.Keys[0].ID,
		"model_name":              "gpt-4o",
		"billing_mode":            "per_request",
		"probe_policy":            "concurrent",
		"probe_interval_seconds":  60,
		"probe_concurrency_limit": 3,
	}, upsertRouteTargetOverride)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if res.Message != "invalid channel key id for channel" {
		t.Fatalf("message = %q, want %q", res.Message, "invalid channel key id for channel")
	}
}

func TestListRouteTargetOverridesFiltersByChannel(t *testing.T) {
	setupHandlerTest(t)
	ctx := setupHandlerTestDB(t)
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() error = %v", err)
	}

	createChannelWithOverride := func(name, keyValue string) (*model.Channel, *model.Channel, error) {
		channel := &model.Channel{
			Name:              name,
			Type:              transformerOutbound.OutboundTypeOpenAIChat,
			Enabled:           true,
			KeyManagementMode: model.KeyManagementModeClassified,
			BaseUrls:          []model.BaseUrl{{URL: "https://example.com/v1", Delay: 0}},
			Model:             "gpt-4o",
		}
		if err := op.ChannelCreate(channel, ctx); err != nil {
			return nil, nil, err
		}
		updated, err := op.ChannelUpdate(&model.ChannelUpdateRequest{
			ID: channel.ID,
			KeysToAdd: []model.ChannelKeyAddRequest{{
				Enabled:       true,
				ChannelKey:    keyValue,
				SourceType:    "paid/metered",
				AllowedModels: "gpt-4o",
			}},
		}, ctx)
		if err != nil {
			return nil, nil, err
		}
		if _, err := op.RouteTargetOverrideUpsert(model.RouteTargetOverride{
			ChannelID:             updated.ID,
			ChannelKeyID:          updated.Keys[0].ID,
			ModelName:             "gpt-4o",
			BillingMode:           model.BillingModePerRequest,
			ProbePolicy:           model.ProbePolicySequential,
			ProbeIntervalSeconds:  120,
			ProbeConcurrencyLimit: 2,
		}, ctx); err != nil {
			return nil, nil, err
		}
		return channel, updated, nil
	}

	_, updatedA, err := createChannelWithOverride("handler-route-target-filter-a", "handler-route-target-filter-key-a")
	if err != nil {
		t.Fatalf("createChannelWithOverride(a) error = %v", err)
	}
	_, _, err = createChannelWithOverride("handler-route-target-filter-b", "handler-route-target-filter-key-b")
	if err != nil {
		t.Fatalf("createChannelWithOverride(b) error = %v", err)
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/route-target/list?channel_id="+strconv.Itoa(updatedA.ID), nil)
	listRouteTargetOverrides(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var rows []model.RouteTargetOverride
	if err := json.Unmarshal(res.Data, &rows); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(rows) != 1 || rows[0].ChannelID != updatedA.ID {
		t.Fatalf("rows = %#v, want only channel_id=%d", rows, updatedA.ID)
	}
}

func TestListRouteTargetOverridesRejectsNonPositiveChannelIDFilter(t *testing.T) {
	setupHandlerTest(t)

	for _, target := range []string{
		"/api/v1/route-target/list?channel_id=0",
		"/api/v1/route-target/list?channel_id=-1",
		"/api/v1/route-target/list?channel_id=",
		"/api/v1/route-target/list?channel_id=%20",
	} {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodGet, target, nil)
		listRouteTargetOverrides(c)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("target %s status = %d, want %d, body = %s", target, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
		res := decodeHandlerResponse(t, recorder)
		if res.Message != "invalid channel id" {
			t.Fatalf("target %s message = %q, want invalid channel id", target, res.Message)
		}
	}
}

func TestDeleteRouteTargetOverrideReturnsNotFoundForMissingRow(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/route-target/delete", map[string]any{
		"channel_id":     1,
		"channel_key_id": 1,
		"model_name":     "gpt-4o",
	}, deleteRouteTargetOverride)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}

func TestListRouteTargetOverridesReturnsRows(t *testing.T) {
	setupHandlerTest(t)
	ctx := setupHandlerTestDB(t)
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() error = %v", err)
	}

	channel := &model.Channel{
		Name:              "handler-route-target-list-channel",
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
			ChannelKey:    "handler-route-target-list-key",
			SourceType:    "paid/metered",
			AllowedModels: "gpt-4o",
		}},
	}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate() error = %v", err)
	}
	if _, err := op.RouteTargetOverrideUpsert(model.RouteTargetOverride{
		ChannelID:             updated.ID,
		ChannelKeyID:          updated.Keys[0].ID,
		ModelName:             "gpt-4o",
		BillingMode:           model.BillingModePerRequest,
		ProbePolicy:           model.ProbePolicySequential,
		ProbeIntervalSeconds:  120,
		ProbeConcurrencyLimit: 2,
	}, ctx); err != nil {
		t.Fatalf("RouteTargetOverrideUpsert() error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/route-target/list", nil, listRouteTargetOverrides)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var rows []model.RouteTargetOverride
	if err := json.Unmarshal(res.Data, &rows); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(rows) == 0 {
		t.Fatalf("rows = %#v, want non-empty route target overrides", rows)
	}
}

func TestListRouteTargetOverridesRejectsInvalidChannelIDFilter(t *testing.T) {
	setupHandlerTest(t)

	for _, target := range []string{
		"/api/v1/route-target/list?channel_id=0",
		"/api/v1/route-target/list?channel_id=-1",
		"/api/v1/route-target/list?channel_id=",
		"/api/v1/route-target/list?channel_id=%20",
	} {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodGet, target, nil)
		listRouteTargetOverrides(c)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("target %s status = %d, want %d, body = %s", target, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
		res := decodeHandlerResponse(t, recorder)
		if res.Message != "invalid channel id" {
			t.Fatalf("target %s message = %q, want invalid channel id", target, res.Message)
		}
	}
}
