package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	transformerOutbound "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
)

func TestCreateChannelReturnsBadRequestForInvalidSourceType(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/create", map[string]any{
		"name":      "handler-invalid-source-type",
		"type":      int(transformerOutbound.OutboundTypeOpenAIChat),
		"base_urls": []map[string]any{{"url": "https://example.com/v1", "delay": 0}},
		"keys":      []map[string]any{{"enabled": true, "channel_key": "sk-test", "source_type": "enterprise"}},
		"model":     "gpt-4o",
	}, createChannel)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if res.Message == "" {
		t.Fatalf("expected non-empty error message")
	}
}

func TestUpdateChannelReturnsNotFoundForMissingChannel(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/update", map[string]any{
		"id":   999999,
		"name": "missing-channel",
	}, updateChannel)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}

func TestDeleteChannelRejectsNonPositivePathIDs(t *testing.T) {
	setupHandlerTest(t)

	for _, id := range []string{"0", "-1"} {
		recorder := performParamHandlerRequest(t, http.MethodDelete, "/api/v1/channel/delete/"+id, nil, map[string]string{"id": id}, deleteChannel)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("id %s status = %d, want %d, body = %s", id, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
		res := decodeHandlerResponse(t, recorder)
		if res.Message != "Invalid parameter" {
			t.Fatalf("id %s message = %q, want %q", id, res.Message, "Invalid parameter")
		}
	}
}

func TestTestChannelModelsByConfigReturnsBadRequestForEmptyModels(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/test-models-by-config", map[string]any{
		"type":      int(transformerOutbound.OutboundTypeOpenAIChat),
		"base_urls": []map[string]any{{"url": "https://example.com/v1", "delay": 0}},
		"keys":      []map[string]any{{"enabled": true, "channel_key": "sk-test", "source_type": "public/free"}},
		"models":    []string{},
	}, testChannelModelsByConfig)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if res.Message != "models is required" {
		t.Fatalf("message = %q, want %q", res.Message, "models is required")
	}
}

func TestFetchModelAcceptsHTTPBaseURL(t *testing.T) {
	setupHandlerTest(t)

	upstream := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-http" {
			t.Fatalf("authorization = %q, want %q", got, "Bearer sk-http")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"http-model"}]}`))
	}))
	defer upstream.Close()

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/fetch-model", map[string]any{
		"type":     int(transformerOutbound.OutboundTypeOpenAIResponse),
		"base_url": upstream.URL,
		"key":      "sk-http",
	}, fetchModel)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	res := decodeHandlerResponse(t, recorder)
	var models []string
	if err := json.Unmarshal(res.Data, &models); err != nil {
		t.Fatalf("json.Unmarshal(response.data) error = %v, body = %s", err, recorder.Body.String())
	}
	if len(models) != 1 || models[0] != "http-model" {
		t.Fatalf("models = %#v, want [http-model]", models)
	}
}

func TestFetchModelRejectsInvalidBaseURL(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/fetch-model", map[string]any{
		"type":     int(transformerOutbound.OutboundTypeOpenAIResponse),
		"base_url": "ftp://example.com/v1",
		"key":      "sk-http",
	}, fetchModel)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if res.Message != "base_url must be absolute http or https URL" {
		t.Fatalf("message = %q, want %q", res.Message, "base_url must be absolute http or https URL")
	}
}

func TestFetchModelAcceptsZeroOutboundTypeForOpenAIChat(t *testing.T) {
	setupHandlerTest(t)

	upstream := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-zero" {
			t.Fatalf("authorization = %q, want %q", got, "Bearer sk-zero")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"zero-type-model"}]}`))
	}))
	defer upstream.Close()

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/fetch-model", map[string]any{
		"type":     0,
		"base_url": upstream.URL,
		"key":      "sk-zero",
	}, fetchModel)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	res := decodeHandlerResponse(t, recorder)
	var models []string
	if err := json.Unmarshal(res.Data, &models); err != nil {
		t.Fatalf("json.Unmarshal(response.data) error = %v, body = %s", err, recorder.Body.String())
	}
	if len(models) != 1 || models[0] != "zero-type-model" {
		t.Fatalf("models = %#v, want [zero-type-model]", models)
	}
}

func TestTestChannelModelsByConfigRejectsInvalidBaseURL(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/test-models-by-config", map[string]any{
		"type":      int(transformerOutbound.OutboundTypeOpenAIChat),
		"base_urls": []map[string]any{{"url": "ftp://example.com/v1", "delay": 0}},
		"keys":      []map[string]any{{"enabled": true, "channel_key": "sk-test", "source_type": "public/free"}},
		"models":    []string{"gpt-4o"},
	}, testChannelModelsByConfig)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if res.Message != "base_url must be absolute http or https URL" {
		t.Fatalf("message = %q, want %q", res.Message, "base_url must be absolute http or https URL")
	}
}
