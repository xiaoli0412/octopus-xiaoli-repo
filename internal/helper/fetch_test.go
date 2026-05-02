package helper

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func responseJSON(req *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: req}
}

func TestDecodeJSONResponseRejectsNon2xx(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusBadGateway,
		Body:       io.NopCloser(strings.NewReader(`{"error":"upstream down"}`)),
	}

	var out model.OpenAIModelList
	err := decodeJSONResponse(resp, &out)
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
	if !strings.Contains(err.Error(), "unexpected status 502") {
		t.Fatalf("expected status in error, got %v", err)
	}
	if !strings.Contains(err.Error(), "upstream down") {
		t.Fatalf("expected upstream body in error, got %v", err)
	}
}

func TestDecodeJSONResponseRejectsOversizedSuccessBody(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(strings.Repeat("x", int(maxModelFetchSuccessBodyBytes)+1))),
	}

	var out model.OpenAIModelList
	err := decodeJSONResponse(resp, &out)
	if err == nil || !strings.Contains(err.Error(), "model fetch response too large") {
		t.Fatalf("decodeJSONResponse() error = %v, want model fetch response too large", err)
	}
}

func TestFetchOpenAIModelsReturnsStatusError(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/models" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		if got := req.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization header %q", got)
		}
		return responseJSON(req, http.StatusUnauthorized, `{"error":"bad key"}`), nil
	})}

	request := model.Channel{
		Type:     outbound.OutboundTypeOpenAIChat,
		BaseUrls: []model.BaseUrl{{URL: "http://example.invalid"}},
		Keys:     []model.ChannelKey{{Enabled: true, ChannelKey: "test-key"}},
	}

	_, err := fetchOpenAIModels(client, context.Background(), request, request.Keys[0])
	if err == nil {
		t.Fatal("expected fetchOpenAIModels to fail on non-2xx response")
	}
	if !strings.Contains(err.Error(), "unexpected status 401") {
		t.Fatalf("expected status error, got %v", err)
	}
	if !strings.Contains(err.Error(), "bad key") {
		t.Fatalf("expected upstream body in error, got %v", err)
	}
}

func TestFetchGeminiModelsAggregatesPages(t *testing.T) {
	t.Parallel()

	requests := 0
	client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		if req.URL.Path != "/models" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		if got := req.Header.Get("X-Goog-Api-Key"); got != "gemini-key" {
			t.Fatalf("unexpected api key header %q", got)
		}
		switch req.URL.Query().Get("pageToken") {
		case "":
			return responseJSON(req, http.StatusOK, `{"models":[{"name":"models/gemini-1.5-pro"}],"nextPageToken":"page-2"}`), nil
		case "page-2":
			return responseJSON(req, http.StatusOK, `{"models":[{"name":"models/gemini-1.5-flash"}]}`), nil
		default:
			t.Fatalf("unexpected page token %q", req.URL.Query().Get("pageToken"))
			return nil, nil
		}
	})}

	request := model.Channel{
		Type:     outbound.OutboundTypeGemini,
		BaseUrls: []model.BaseUrl{{URL: "http://example.invalid"}},
		Keys:     []model.ChannelKey{{Enabled: true, ChannelKey: "gemini-key"}},
	}

	models, err := fetchGeminiModels(client, context.Background(), request, request.Keys[0])
	if err != nil {
		t.Fatalf("fetchGeminiModels returned error: %v", err)
	}
	if requests != 2 {
		t.Fatalf("expected 2 paged requests, got %d", requests)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}
	if models[0] != "gemini-1.5-pro" || models[1] != "gemini-1.5-flash" {
		t.Fatalf("unexpected models: %#v", models)
	}
}

func TestFetchGeminiModelsPinsSingleKeyAcrossPages(t *testing.T) {
	t.Parallel()

	requests := 0
	client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		if got := req.Header.Get("X-Goog-Api-Key"); got != "first-key" {
			t.Fatalf("unexpected api key header %q", got)
		}
		switch req.URL.Query().Get("pageToken") {
		case "":
			return responseJSON(req, http.StatusOK, `{"models":[{"name":"models/gemini-1.5-pro"}],"nextPageToken":"page-2"}`), nil
		case "page-2":
			return responseJSON(req, http.StatusOK, `{"models":[{"name":"models/gemini-1.5-flash"}]}`), nil
		default:
			t.Fatalf("unexpected page token %q", req.URL.Query().Get("pageToken"))
			return nil, nil
		}
	})}

	request := model.Channel{
		Type:             outbound.OutboundTypeGemini,
		KeyRoutingPolicy: model.KeyRoutingPolicyRoundRobin,
		BaseUrls:         []model.BaseUrl{{URL: "http://example.invalid"}},
		Keys: []model.ChannelKey{
			{ID: 1, Enabled: true, ChannelKey: "first-key"},
			{ID: 2, Enabled: true, ChannelKey: "second-key"},
		},
	}

	models, err := fetchGeminiModels(client, context.Background(), request, request.Keys[0])
	if err != nil {
		t.Fatalf("fetchGeminiModels returned error: %v", err)
	}
	if requests != 2 {
		t.Fatalf("expected 2 paged requests, got %d", requests)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}
	if models[0] != "gemini-1.5-pro" || models[1] != "gemini-1.5-flash" {
		t.Fatalf("unexpected models: %#v", models)
	}
}

func TestFetchAntigravityModelsEscapesProjectID(t *testing.T) {
	t.Parallel()

	projectID := `proj"quote\\slash`
	client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/v1internal:retrieveUserQuota" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		if got := req.Header.Get("Authorization"); got != "Bearer oauth-token" {
			t.Fatalf("unexpected authorization header %q", got)
		}

		var payload struct {
			Project string `json:"project"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatalf("expected valid JSON payload: %v", err)
		}
		if payload.Project != projectID {
			t.Fatalf("project = %q, want %q", payload.Project, projectID)
		}
		return responseJSON(req, http.StatusOK, `{"buckets":[{"modelId":"gemini-code-assist"}]}`), nil
	})}

	request := model.Channel{
		Type:     outbound.OutboundTypeAntigravity,
		BaseUrls: []model.BaseUrl{{URL: "http://example.invalid"}},
		Keys:     []model.ChannelKey{{Enabled: true, ChannelKey: "oauth-token|" + projectID}},
	}

	models, err := fetchAntigravityModels(client, context.Background(), request, request.Keys[0])
	if err != nil {
		t.Fatalf("fetchAntigravityModels returned error: %v", err)
	}
	if len(models) != 1 || models[0] != "gemini-code-assist" {
		t.Fatalf("unexpected models: %#v", models)
	}
}

func TestFetchOpenAIModelsRejectsInvalidBaseURL(t *testing.T) {
	t.Parallel()

	request := model.Channel{
		Type:     outbound.OutboundTypeOpenAIChat,
		BaseUrls: []model.BaseUrl{{URL: "://bad-url"}},
		Keys:     []model.ChannelKey{{Enabled: true, ChannelKey: "test-key"}},
	}

	_, err := fetchOpenAIModels(http.DefaultClient, context.Background(), request, request.Keys[0])
	if err == nil {
		t.Fatal("expected fetchOpenAIModels to reject invalid base url")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "missing protocol scheme") {
		t.Fatalf("expected invalid URL error, got %v", err)
	}
}

func TestFetchModelsAppliesManagementTimeout(t *testing.T) {
	t.Parallel()

	original := channelHTTPClientForManagement
	channelHTTPClientForManagement = func(channel *model.Channel) (*http.Client, error) {
		return &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			<-req.Context().Done()
			return nil, req.Context().Err()
		})}, nil
	}
	t.Cleanup(func() {
		channelHTTPClientForManagement = original
	})

	request := model.Channel{
		Type:     outbound.OutboundTypeOpenAIChat,
		BaseUrls: []model.BaseUrl{{URL: "http://example.invalid"}},
		Keys:     []model.ChannelKey{{Enabled: true, ChannelKey: "test-key"}},
	}

	start := time.Now()
	_, err := FetchModels(context.Background(), request)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("FetchModels() error = %v, want context deadline exceeded", err)
	}
	if elapsed := time.Since(start); elapsed >= 25*time.Second {
		t.Fatalf("FetchModels() elapsed = %v, want < 25s management timeout", elapsed)
	}
}

func TestManagementHTTPClientNoRedirectBlocksCrossHostRedirects(t *testing.T) {
	t.Parallel()

	var redirectTargetHit atomic.Bool
	redirectTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectTargetHit.Store(true)
		w.WriteHeader(http.StatusOK)
	}))
	defer redirectTarget.Close()

	redirectSource := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, redirectTarget.URL+"/models", http.StatusFound)
	}))
	defer redirectSource.Close()

	client, err := managementHTTPClientNoRedirect(&model.Channel{BaseUrls: []model.BaseUrl{{URL: redirectSource.URL}}})
	if err != nil {
		t.Fatalf("managementHTTPClientNoRedirect() error = %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, redirectSource.URL+"/models", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusFound)
	}
	if redirectTargetHit.Load() {
		t.Fatal("redirect target was reached, want redirect to be blocked")
	}
}

func TestManagementHTTPClientNoRedirectAllowsSameHostRedirects(t *testing.T) {
	t.Parallel()

	var redirectTargetHit atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/models":
			http.Redirect(w, r, "/v1/models", http.StatusFound)
		case "/v1/models":
			redirectTargetHit.Store(true)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := managementHTTPClientNoRedirect(&model.Channel{BaseUrls: []model.BaseUrl{{URL: server.URL}}})
	if err != nil {
		t.Fatalf("managementHTTPClientNoRedirect() error = %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, server.URL+"/models", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if !redirectTargetHit.Load() {
		t.Fatal("redirect target was not reached, want same-host redirect to be followed")
	}
}

func TestGetUrlDelayHonorsContextDeadline(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	})}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := GetUrlDelay(httpClient, "http://example.invalid", ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("GetUrlDelay() error = %v, want context deadline exceeded", err)
	}
}

func TestGetUrlDelayRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	_, err := GetUrlDelay(http.DefaultClient, "://bad-url", context.Background())
	if err == nil {
		t.Fatal("expected GetUrlDelay to reject invalid url")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "missing protocol scheme") {
		t.Fatalf("expected invalid URL error, got %v", err)
	}
}
