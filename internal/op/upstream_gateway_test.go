package op

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

// upstreamMockHandler returns a handler that simulates the three upstream
// protocol families supported by scripts/test-upstream-mock.
func upstreamMockHandler() http.Handler {
	const (
		newAPIToken  = "newapi-management-token"
		sub2APIToken = "sub2-management-token"
		openAIKey    = "sk-openai-compatible"
		cookieToken  = "cookie-session-token"
	)

	bearer := func(r *http.Request) string {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
		return ""
	}
	cookieSession := func(r *http.Request) string {
		for _, c := range strings.Split(r.Header.Get("Cookie"), ";") {
			parts := strings.SplitN(strings.TrimSpace(c), "=", 2)
			if len(parts) == 2 && parts[0] == "session" {
				return parts[1]
			}
		}
		return ""
	}
	writeJSON := func(w http.ResponseWriter, status int, payload any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(payload)
	}

	mux := http.NewServeMux()

	// new API endpoints.
	mux.HandleFunc("/api/user/me", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != newAPIToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": "u1", "username": "owner"}})
	})
	mux.HandleFunc("/api/user/self", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != newAPIToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"username": "owner", "used_quota": 10.0, "remain_quota": 990.0}})
	})
	mux.HandleFunc("/api/user/models", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != newAPIToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": []map[string]any{{"id": "gpt-4o"}, {"id": "gpt-4o-mini"}}})
	})
	mux.HandleFunc("/api/user/self/groups", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != newAPIToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": []map[string]any{{"id": 1, "name": "default", "models": []string{"gpt-4o"}, "rate_multiplier": 1.0}}})
	})
	mux.HandleFunc("/api/token/search", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != newAPIToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": []map[string]any{{"id": 1, "name": "default", "key": "sk-newapi-importable", "models": "gpt-4o"}}})
	})
	mux.HandleFunc("/api/token", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != newAPIToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": []map[string]any{{"id": 1, "name": "default", "key": "sk-newapi-importable", "models": "gpt-4o"}}})
	})
	mux.HandleFunc("/api/pricing", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != newAPIToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"model_ratio": map[string]float64{"gpt-4o": 0.5}, "completion_ratio": map[string]float64{"gpt-4o": 2.0}})
	})
	mux.HandleFunc("/api/user/checkin", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != newAPIToken && cookieSession(r) != cookieToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{"quota": 3.3}, "message": "签到成功"})
	})

	// sub2API endpoints.
	mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["username"] == "admin@example.com" && body["password"] == "admin123" {
			writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"access_token": sub2APIToken, "user_id": "sub2-user-1"}})
			return
		}
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid credentials"})
	})
	mux.HandleFunc("/api/profile", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != sub2APIToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": "sub2-user-1", "quota": 500.0, "used": 50.0}})
	})
	mux.HandleFunc("/api/v1/user/profile", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != sub2APIToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": "sub2-user-1", "quota": 500.0, "used": 50.0}})
	})
	mux.HandleFunc("/api/keys", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != sub2APIToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": []map[string]any{{"id": 1, "name": "main", "key": "sk-sub2-importable", "models": []string{"gpt-4o"}, "groups": []string{"vip"}, "status": "enabled"}}})
	})
	mux.HandleFunc("/api/v1/keys", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != sub2APIToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": []map[string]any{{"id": 1, "name": "main", "key": "sk-sub2-importable", "models": []string{"gpt-4o"}, "groups": []string{"vip"}, "status": "enabled"}}})
	})
	mux.HandleFunc("/api/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != sub2APIToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": []map[string]any{{"name": "pro", "plan": "monthly", "status": "active", "balance": 450.0}}})
	})
	mux.HandleFunc("/api/v1/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != sub2APIToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": []map[string]any{{"name": "pro", "plan": "monthly", "status": "active", "balance": 450.0}}})
	})
	mux.HandleFunc("/api/groups", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != sub2APIToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": []map[string]any{{"id": 1, "name": "vip", "models": []string{"gpt-4o"}, "rate_multiplier": 0.8, "status": "active"}}})
	})
	mux.HandleFunc("/api/v1/pricing", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != sub2APIToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": []map[string]any{{"model": "gpt-4o", "input_price": 0.6, "output_price": 1.2}}})
	})
	mux.HandleFunc("/api/v1/user/checkin", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != sub2APIToken && cookieSession(r) != cookieToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "amount": 7.7, "message": "sub2 checkin reward"})
	})

	// OpenAI Compatible endpoints.
	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != openAIKey && r.Header.Get("x-api-key") != openAIKey {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"object": "list", "data": []map[string]any{{"id": "gpt-3.5-turbo"}, {"id": "gpt-4o"}}})
	})
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		if bearer(r) != openAIKey && r.Header.Get("x-api-key") != openAIKey {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"id": "chatcmpl-test", "object": "chat.completion", "model": "gpt-3.5-turbo", "choices": []map[string]any{{"index": 0, "message": map[string]string{"role": "assistant", "content": "hi"}, "finish_reason": "stop"}}})
	})

	return mux
}

func TestInspectUpstreamGatewayNewAPI(t *testing.T) {
	ctx := SetupOpTestDB(t)
	server := httptest.NewServer(upstreamMockHandler())
	defer server.Close()

	result, err := InspectUpstreamGateway(ctx, model.UpstreamInspectRequest{
		ProviderType: model.UpstreamProviderNewAPI,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeToken,
		Token:        "newapi-management-token",
	})
	if err != nil {
		t.Fatalf("InspectUpstreamGateway(newapi) error = %v", err)
	}
	if result.ProviderType != model.UpstreamProviderNewAPI {
		t.Fatalf("provider type = %q, want newapi", result.ProviderType)
	}
	if result.ModelCount < 1 {
		t.Fatalf("expected at least one model, got %d", result.ModelCount)
	}
	if !result.TokenUsage.Available {
		t.Fatalf("expected balance available")
	}
	if result.TokenUsage.RemainQuota != 990.0 {
		t.Fatalf("remain quota = %v, want 990", result.TokenUsage.RemainQuota)
	}
	if len(result.Keys) == 0 {
		t.Fatalf("expected at least one key")
	}
	found := false
	for _, m := range result.Models {
		if m == "gpt-4o" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected gpt-4o in models, got %v", result.Models)
	}
}

func TestInspectUpstreamGatewaySub2API(t *testing.T) {
	ctx := SetupOpTestDB(t)
	server := httptest.NewServer(upstreamMockHandler())
	defer server.Close()

	result, err := InspectUpstreamGateway(ctx, model.UpstreamInspectRequest{
		ProviderType: model.UpstreamProviderSub2API,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeAccountPassword,
		Username:     "admin@example.com",
		Password:     "admin123",
	})
	if err != nil {
		t.Fatalf("InspectUpstreamGateway(sub2api) error = %v", err)
	}
	if result.ProviderType != model.UpstreamProviderSub2API {
		t.Fatalf("provider type = %q, want sub2api", result.ProviderType)
	}
	if result.ModelCount < 1 {
		t.Fatalf("expected at least one model, got %d", result.ModelCount)
	}
	if !result.TokenUsage.Available {
		t.Fatalf("expected balance available")
	}
	if len(result.Keys) == 0 {
		t.Fatalf("expected at least one key")
	}
	if len(result.Subscriptions) == 0 {
		t.Fatalf("expected at least one subscription")
	}
}

func TestInspectUpstreamGatewayOpenAICompatible(t *testing.T) {
	ctx := SetupOpTestDB(t)
	server := httptest.NewServer(upstreamMockHandler())
	defer server.Close()

	result, err := InspectUpstreamGateway(ctx, model.UpstreamInspectRequest{
		ProviderType: model.UpstreamProviderOpenAICompatible,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeAccessKey,
		AccessKey:    "sk-openai-compatible",
	})
	if err != nil {
		t.Fatalf("InspectUpstreamGateway(openai_compatible) error = %v", err)
	}
	if result.ProviderType != model.UpstreamProviderOpenAICompatible {
		t.Fatalf("provider type = %q, want openai_compatible", result.ProviderType)
	}
	if result.ModelCount != 2 {
		t.Fatalf("model count = %d, want 2", result.ModelCount)
	}
}

func TestInspectUpstreamGatewayNewAPIInvalidToken(t *testing.T) {
	ctx := SetupOpTestDB(t)
	server := httptest.NewServer(upstreamMockHandler())
	defer server.Close()

	_, err := InspectUpstreamGateway(ctx, model.UpstreamInspectRequest{
		ProviderType: model.UpstreamProviderNewAPI,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeToken,
		Token:        "wrong-token",
	})
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestUpstreamSiteCheckinCookieFallback(t *testing.T) {
	ctx := SetupOpTestDB(t)
	server := httptest.NewServer(upstreamMockHandler())
	defer server.Close()

	// Create a new API site with a valid management token so inspection succeeds.
	detail, err := UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:         "newapi cookie checkin",
		ProviderType: model.UpstreamProviderNewAPI,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeToken,
		Token:        "newapi-management-token",
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}

	// Replace the stored management token with a cookie-only token to exercise checkin fallback.
	encrypted, err := encryptUpstreamSecret("cookie-session-token")
	if err != nil {
		t.Fatalf("encryptUpstreamSecret() error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Model(&model.UpstreamCredential{}).
		Where("upstream_site_id = ? AND credential_type = ?", detail.Site.ID, model.UpstreamCredentialManagementToken).
		Update("encrypted_value", encrypted).Error; err != nil {
		t.Fatalf("update credential error = %v", err)
	}

	result, err := UpstreamSiteCheckin(ctx, detail.Site.ID)
	if err != nil {
		t.Fatalf("UpstreamSiteCheckin() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("checkin failed")
	}
	if result.Amount != 3.3 {
		t.Fatalf("amount = %v, want 3.3", result.Amount)
	}
}

func TestUpstreamSiteRefreshRecordsBalanceLog(t *testing.T) {
	ctx := SetupOpTestDB(t)
	server := httptest.NewServer(upstreamMockHandler())
	defer server.Close()

	detail, err := UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:         "newapi balance log",
		ProviderType: model.UpstreamProviderNewAPI,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeToken,
		Token:        "newapi-management-token",
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}

	if _, err := UpstreamSiteRefresh(ctx, model.UpstreamRefreshRequest{ID: detail.Site.ID, Manual: true}); err != nil {
		t.Fatalf("UpstreamSiteRefresh() error = %v", err)
	}

	refreshed, err := UpstreamSiteDetailGet(ctx, detail.Site.ID)
	if err != nil {
		t.Fatalf("UpstreamSiteDetailGet() error = %v", err)
	}
	if refreshed.Site.BalanceRefreshLog == "" {
		t.Fatal("expected balance refresh log to be recorded")
	}
	var entries []model.UpstreamBalanceLogEntry
	if err := json.Unmarshal([]byte(refreshed.Site.BalanceRefreshLog), &entries); err != nil {
		t.Fatalf("unmarshal balance log error = %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one balance log entry")
	}
	if entries[len(entries)-1].Remain != 990.0 {
		t.Fatalf("last balance remain = %v, want 990", entries[len(entries)-1].Remain)
	}
}
