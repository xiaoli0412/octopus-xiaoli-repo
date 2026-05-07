package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/gin-gonic/gin"
)

func TestDeleteAPIKeyRejectsNonPositivePathIDs(t *testing.T) {
	setupHandlerTest(t)

	for _, id := range []string{"0", "-1"} {
		recorder := performParamHandlerRequest(t, http.MethodDelete, "/api/v1/apikey/delete/"+id, nil, map[string]string{"id": id}, deleteAPIKey)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("id %s status = %d, want %d, body = %s", id, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
		res := decodeHandlerResponse(t, recorder)
		if res.Message != "Invalid parameter" {
			t.Fatalf("id %s message = %q, want %q", id, res.Message, "Invalid parameter")
		}
	}
}

func TestLoginAPIKeyReturnsStructuredStatus(t *testing.T) {
	setupHandlerTest(t)

	apiKey := model.APIKey{
		Name:            "dashboard-key",
		APIKey:          "sk-octopus-dashboard-123",
		Enabled:         true,
		ExpireAt:        time.Now().Add(24 * time.Hour).Unix(),
		SupportedModels: "gpt-4o,claude-3-5-sonnet",
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/apikey/login", nil)
	if err := op.APIKeyCreate(&apiKey, ctx.Request.Context()); err != nil {
		t.Fatalf("op.APIKeyCreate() error = %v", err)
	}
	ctx.Request.Header.Set("Authorization", "Bearer "+apiKey.APIKey)
	ctx.Set("api_key_id", apiKey.ID)
	loginAPIKey(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	response := decodeHandlerResponse(t, recorder)
	var payload model.APIKeyAuthStatus
	if err := json.Unmarshal(response.Data, &payload); err != nil {
		t.Fatalf("json.Unmarshal(payload) error = %v", err)
	}
	if !payload.OK {
		t.Fatalf("payload.OK = false, want true")
	}
	if payload.APIKeyID != apiKey.ID || payload.Name != apiKey.Name {
		t.Fatalf("payload = %#v, want matching api key identity", payload)
	}
	if payload.AuthMode != "api_key" {
		t.Fatalf("payload.AuthMode = %q, want %q", payload.AuthMode, "api_key")
	}
	if payload.SupportedModels != apiKey.SupportedModels {
		t.Fatalf("payload.SupportedModels = %q, want %q", payload.SupportedModels, apiKey.SupportedModels)
	}
}
