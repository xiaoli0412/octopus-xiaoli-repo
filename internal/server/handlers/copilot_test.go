package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestCopilotHTTPClientHasTimeout(t *testing.T) {
	if copilotHTTPClient == nil {
		t.Fatal("copilotHTTPClient is nil")
	}
	if copilotHTTPClient.Timeout < 15*time.Second {
		t.Fatalf("copilotHTTPClient.Timeout = %v, want at least 15s", copilotHTTPClient.Timeout)
	}
	if copilotHTTPClient.Timeout != 15*time.Second {
		t.Fatalf("copilotHTTPClient.Timeout = %v, want %v", copilotHTTPClient.Timeout, 15*time.Second)
	}
	if copilotHTTPClient == http.DefaultClient {
		t.Fatal("copilotHTTPClient should not reuse http.DefaultClient")
	}
}

func TestCopilotRequestDeviceCodeSendsJSONBody(t *testing.T) {
	setupHandlerTest(t)

	var capturedBody map[string]string
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("json decode error = %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"device_code":"device-123","user_code":"user-123","verification_uri":"https://example.com","expires_in":900,"interval":5}`))
	}))
	defer server.Close()

	t.Setenv("OCTOPUS_COPILOT_CLIENT_ID", "client-123")
	t.Setenv("OCTOPUS_COPILOT_SCOPE", "scope-a scope-b")
	t.Setenv("OCTOPUS_COPILOT_DEVICE_CODE_URL", server.URL)
	copilotHTTPClient = server.Client()
	t.Cleanup(func() {
		copilotHTTPClient = &http.Client{Timeout: 15 * time.Second}
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/channel/copilot/device-code", nil)
	copilotRequestDeviceCode(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if capturedBody["client_id"] != "client-123" || capturedBody["scope"] != "scope-a scope-b" {
		t.Fatalf("capturedBody = %#v, want JSON encoded credentials", capturedBody)
	}
}
func TestCopilotPollTokenSendsJSONBody(t *testing.T) {
	setupHandlerTest(t)

	var capturedBody map[string]string
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("json decode error = %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"token-123","token_type":"bearer"}`))
	}))
	defer server.Close()

	t.Setenv("OCTOPUS_COPILOT_CLIENT_ID", "client-123")
	t.Setenv("OCTOPUS_COPILOT_ACCESS_TOKEN_URL", server.URL)
	copilotHTTPClient = server.Client()
	t.Cleanup(func() {
		copilotHTTPClient = &http.Client{Timeout: 15 * time.Second}
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/channel/copilot/poll-token", strings.NewReader(`{"device_code":"device-123"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	copilotPollToken(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if capturedBody["client_id"] != "client-123" || capturedBody["device_code"] != "device-123" {
		t.Fatalf("capturedBody = %#v, want JSON encoded poll payload", capturedBody)
	}
}

func TestDecodeCopilotOAuthResponseRejectsOversizedBody(t *testing.T) {
	payload := strings.Repeat("a", int(maxCopilotOAuthResponseBytes)+1)

	err := decodeCopilotOAuthResponse(strings.NewReader(payload), &copilotPollResponse{})
	if err == nil || !strings.Contains(err.Error(), "copilot oauth response too large") {
		t.Fatalf("decodeCopilotOAuthResponse() error = %v, want size limit error", err)
	}
}

func TestCopilotPollTokenRejectsOversizedOAuthResponse(t *testing.T) {
	setupHandlerTest(t)

	originalClient := copilotHTTPClient
	defer func() { copilotHTTPClient = originalClient }()

	copilotHTTPClient = &http.Client{Transport: copilotRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(strings.Repeat("x", int(maxCopilotOAuthResponseBytes)+1))),
			Request:    req,
		}, nil
	})}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/channel/copilot/poll-token", strings.NewReader(`{"device_code":"device-123"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	copilotPollToken(ctx)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadGateway, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "copilot oauth response too large") {
		t.Fatalf("body = %q, want oversized response message", recorder.Body.String())
	}
}

func TestCopilotOAuthConfigRejectsInvalidOverrideURLs(t *testing.T) {
	t.Setenv("OCTOPUS_COPILOT_DEVICE_CODE_URL", "ftp://github.com/login/device/code")

	_, _, _, _, err := copilotOAuthConfig()
	if err == nil {
		t.Fatal("copilotOAuthConfig() error = nil, want invalid override URL error")
	}
	if got := err.Error(); got != "copilot device code url must be absolute http or https URL" {
		t.Fatalf("copilotOAuthConfig() error = %q, want %q", got, "copilot device code url must be absolute http or https URL")
	}
}

func TestCopilotRequestDeviceCodeRejectsCredentialBearingOverrideURL(t *testing.T) {
	setupHandlerTest(t)
	t.Setenv("OCTOPUS_COPILOT_DEVICE_CODE_URL", "https://user:pass@github.com/login/device/code")

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/channel/copilot/device-code", nil)

	copilotRequestDeviceCode(ctx)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusInternalServerError, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "copilot device code url must not include credentials") {
		t.Fatalf("body = %q, want credential validation message", recorder.Body.String())
	}
}

type copilotRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn copilotRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
