package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

func TestAntigravityHTTPClientHasTimeout(t *testing.T) {
	if antigravityHTTPClient == nil {
		t.Fatal("antigravityHTTPClient is nil")
	}
	if antigravityHTTPClient.Timeout < 15*time.Second {
		t.Fatalf("antigravityHTTPClient.Timeout = %v, want at least 15s", antigravityHTTPClient.Timeout)
	}
	if antigravityHTTPClient.Timeout != 15*time.Second {
		t.Fatalf("antigravityHTTPClient.Timeout = %v, want %v", antigravityHTTPClient.Timeout, 15*time.Second)
	}
	if antigravityHTTPClient == http.DefaultClient {
		t.Fatal("antigravityHTTPClient should not reuse http.DefaultClient")
	}
}

func TestBuildAPIPublicBaseUsesRequestHost(t *testing.T) {
	setupHandlerTest(t)
	if err := op.SettingSetString(model.SettingKeyAPIBaseURL, ""); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channel/antigravity/oauth/start", nil)
	req.Host = "octopus.example:8443"
	req.Header.Set("X-Forwarded-Host", "attacker.example")
	req.Header.Set("X-Forwarded-Proto", "http")
	ctx.Request = req

	if got := buildAPIPublicBase(ctx); got != "http://octopus.example:8443" {
		t.Fatalf("buildAPIPublicBase() = %q, want %q", got, "http://octopus.example:8443")
	}
}

func TestBuildAPIPublicBaseIgnoresLoopbackConfiguredBaseURLForRemoteHost(t *testing.T) {
	setupHandlerTest(t)
	if err := op.SettingSetString(model.SettingKeyAPIBaseURL, "http://localhost:8080"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channel/antigravity/oauth/start", nil)
	req.Host = "gateway.example.com"
	ctx.Request = req

	if got := buildAPIPublicBase(ctx); got != "http://gateway.example.com" {
		t.Fatalf("buildAPIPublicBase() = %q, want %q", got, "http://gateway.example.com")
	}
}

func TestBuildAPIPublicBasePrefersConfiguredExternalBaseURL(t *testing.T) {
	setupHandlerTest(t)
	if err := op.SettingSetString(model.SettingKeyAPIBaseURL, "https://api.example.com/root/"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channel/antigravity/oauth/start", nil)
	req.Host = "gateway.example.com"
	ctx.Request = req

	if got := buildAPIPublicBase(ctx); got != "https://api.example.com/root" {
		t.Fatalf("buildAPIPublicBase() = %q, want %q", got, "https://api.example.com/root")
	}
}

func TestBuildAPIPublicBaseIgnoresInvalidConfiguredBaseURL(t *testing.T) {
	setupHandlerTest(t)
	if err := op.SettingSetString(model.SettingKeyAPIBaseURL, "ftp://gateway.example.com"); err == nil {
		t.Fatal("SettingSetString() error = nil, want invalid setting rejection")
	}
	if err := db.GetDB().Model(&model.Setting{}).Where("key = ?", model.SettingKeyAPIBaseURL).Update("value", "ftp://gateway.example.com").Error; err != nil {
		t.Fatalf("force invalid api_base_url error = %v", err)
	}
	if err := op.InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channel/antigravity/oauth/start", nil)
	req.Host = "gateway.example.com"
	ctx.Request = req

	if got := buildAPIPublicBase(ctx); got != "http://gateway.example.com" {
		t.Fatalf("buildAPIPublicBase() = %q, want %q", got, "http://gateway.example.com")
	}
}

func TestReadAntigravityTokenResponseRejectsOversizedBody(t *testing.T) {
	payload := strings.Repeat("a", int(maxAntigravityTokenResponseBytes)+1)

	_, _, err := readAntigravityTokenResponse(strings.NewReader(payload))
	if err == nil || !strings.Contains(err.Error(), "antigravity token response too large") {
		t.Fatalf("readAntigravityTokenResponse() error = %v, want size limit error", err)
	}
}

func TestCleanupAntigravitySessionsCapsDistinctStates(t *testing.T) {
	antigravityOAuthLock.Lock()
	defer antigravityOAuthLock.Unlock()

	antigravityOAuthSessions = map[string]*antigravityOAuthSession{}
	base := time.Now()
	for i := 0; i < antigravitySessionMaxEntries+10; i++ {
		state := fmt.Sprintf("state-%03d", i)
		antigravityOAuthSessions[state] = &antigravityOAuthSession{
			Status:    "pending",
			CreatedAt: base.Add(time.Duration(i) * time.Second),
		}
		trimAntigravitySessionsLocked(state)
	}

	if got := len(antigravityOAuthSessions); got != antigravitySessionMaxEntries {
		t.Fatalf("len(antigravityOAuthSessions) = %d, want %d", got, antigravitySessionMaxEntries)
	}
	if _, ok := antigravityOAuthSessions["state-000"]; ok {
		t.Fatal("oldest state still present after trim")
	}
	if _, ok := antigravityOAuthSessions[fmt.Sprintf("state-%03d", antigravitySessionMaxEntries+9)]; !ok {
		t.Fatal("latest state missing after trim")
	}
}

func TestAntigravityOAuthCallbackRejectsOversizedTokenResponse(t *testing.T) {
	setupHandlerTest(t)
	defer resetLoginThrottleState()
	t.Setenv("OCTOPUS_ANTIGRAVITY_CLIENT_ID", "test-client-id")
	t.Setenv("OCTOPUS_ANTIGRAVITY_CLIENT_SECRET", "test-client-secret")

	originalClient := antigravityHTTPClient
	defer func() { antigravityHTTPClient = originalClient }()

	antigravityHTTPClient = &http.Client{Transport: antigravityRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(strings.Repeat("x", int(maxAntigravityTokenResponseBytes)+1))),
			Request:    req,
		}, nil
	})}

	state := "oversized-state"
	antigravityOAuthLock.Lock()
	antigravityOAuthSessions = map[string]*antigravityOAuthSession{
		state: {Status: "pending", CreatedAt: time.Now()},
	}
	antigravityOAuthLock.Unlock()

	ctxRecorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(ctxRecorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/channel/antigravity/oauth/callback?state="+state+"&code=test-code", nil)
	ctx.Request.Host = "octopus.example"

	antigravityOAuthCallback(ctx)

	if ctxRecorder.Code != http.StatusOK {
		t.Fatalf("callback status = %d, want %d", ctxRecorder.Code, http.StatusOK)
	}
	if !strings.Contains(ctxRecorder.Body.String(), "Authorization failed") {
		t.Fatalf("callback body = %q, want Authorization failed page", ctxRecorder.Body.String())
	}

	antigravityOAuthLock.Lock()
	defer antigravityOAuthLock.Unlock()
	session := antigravityOAuthSessions[state]
	if session == nil {
		t.Fatalf("session = nil, want stored failed session")
	}
	if session.Status != "failed" {
		t.Fatalf("session.Status = %q, want failed", session.Status)
	}
	if !strings.Contains(session.Error, "antigravity token response too large") {
		t.Fatalf("session.Error = %q, want oversized response message", session.Error)
	}
}

func TestAntigravityOAuthCallbackRejectsBlankState(t *testing.T) {
	setupHandlerTest(t)
	defer resetLoginThrottleState()

	ctxRecorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(ctxRecorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/channel/antigravity/oauth/callback?state=%20", nil)

	antigravityOAuthCallback(ctx)

	if ctxRecorder.Code != http.StatusBadRequest {
		t.Fatalf("callback status = %d, want %d", ctxRecorder.Code, http.StatusBadRequest)
	}
	if body := strings.TrimSpace(ctxRecorder.Body.String()); body != "missing state" {
		t.Fatalf("callback body = %q, want %q", body, "missing state")
	}
}

func TestAntigravityConfigRejectsInvalidOverrideURLs(t *testing.T) {
	t.Setenv("OCTOPUS_ANTIGRAVITY_AUTHORIZE_URL", "ftp://accounts.google.com/o/oauth2/v2/auth")

	_, _, _, _, _, err := antigravityConfig()
	if err == nil {
		t.Fatal("antigravityConfig() error = nil, want invalid override URL error")
	}
	if got := err.Error(); got != "antigravity authorize url must be absolute http or https URL" {
		t.Fatalf("antigravityConfig() error = %q, want %q", got, "antigravity authorize url must be absolute http or https URL")
	}
}

func TestStartAntigravityOAuthRejectsCredentialBearingAuthorizeURL(t *testing.T) {
	setupHandlerTest(t)
	t.Setenv("OCTOPUS_ANTIGRAVITY_CLIENT_ID", "test-client-id")
	t.Setenv("OCTOPUS_ANTIGRAVITY_AUTHORIZE_URL", "https://user:pass@accounts.google.com/o/oauth2/v2/auth")

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/channel/antigravity/oauth/start", nil)

	startAntigravityOAuth(ctx)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusInternalServerError, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "antigravity authorize url must not include credentials") {
		t.Fatalf("body = %q, want credential validation message", recorder.Body.String())
	}
}

func TestAntigravityOAuthCallbackRejectsBlankCode(t *testing.T) {
	setupHandlerTest(t)
	defer resetLoginThrottleState()

	state := "blank-code-state"
	antigravityOAuthLock.Lock()
	antigravityOAuthSessions = map[string]*antigravityOAuthSession{
		state: {Status: "pending", CreatedAt: time.Now()},
	}
	antigravityOAuthLock.Unlock()

	ctxRecorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(ctxRecorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/channel/antigravity/oauth/callback?state="+state+"&code=%20", nil)

	antigravityOAuthCallback(ctx)

	if ctxRecorder.Code != http.StatusBadRequest {
		t.Fatalf("callback status = %d, want %d", ctxRecorder.Code, http.StatusBadRequest)
	}
	if body := strings.TrimSpace(ctxRecorder.Body.String()); body != "missing code" {
		t.Fatalf("callback body = %q, want %q", body, "missing code")
	}

	antigravityOAuthLock.Lock()
	defer antigravityOAuthLock.Unlock()
	if session := antigravityOAuthSessions[state]; session == nil || session.Status != "pending" {
		t.Fatalf("session = %#v, want pending session preserved", session)
	}
}

func TestAntigravityOAuthCallbackTrimsOAuthErrorFields(t *testing.T) {
	setupHandlerTest(t)
	defer resetLoginThrottleState()

	state := "trimmed-error-state"
	antigravityOAuthLock.Lock()
	antigravityOAuthSessions = map[string]*antigravityOAuthSession{
		state: {Status: "pending", CreatedAt: time.Now()},
	}
	antigravityOAuthLock.Unlock()

	ctxRecorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(ctxRecorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/channel/antigravity/oauth/callback?state=%20"+state+"%20&error=%20access_denied%20&error_description=%20user%20cancelled%20", nil)

	antigravityOAuthCallback(ctx)

	if ctxRecorder.Code != http.StatusOK {
		t.Fatalf("callback status = %d, want %d", ctxRecorder.Code, http.StatusOK)
	}
	if !strings.Contains(ctxRecorder.Body.String(), "Authorization failed") {
		t.Fatalf("callback body = %q, want Authorization failed page", ctxRecorder.Body.String())
	}

	antigravityOAuthLock.Lock()
	defer antigravityOAuthLock.Unlock()
	session := antigravityOAuthSessions[state]
	if session == nil {
		t.Fatalf("session = nil, want stored failed session")
	}
	if session.Status != "failed" {
		t.Fatalf("session.Status = %q, want failed", session.Status)
	}
	if session.Error != "user cancelled" {
		t.Fatalf("session.Error = %q, want %q", session.Error, "user cancelled")
	}
}

type antigravityRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn antigravityRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
