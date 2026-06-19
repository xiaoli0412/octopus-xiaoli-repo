package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

func TestUpdateUpstreamSiteEndpointStoresConfigFields(t *testing.T) {
	setupHandlerTest(t)
	ctx := t.Context()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/user/self":
			_, _ = w.Write([]byte(`{"data":{"username":"owner","remain_quota":100,"used_quota":10,"group":"default"}}`))
		case "/api/user/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		case "/api/user/self/groups":
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/api/token/search":
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/api/pricing":
			_, _ = w.Write([]byte(`{"model_ratio":{}}`))
		case "/v1/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	detail, err := op.UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:         "handler config",
		ProviderType: model.UpstreamProviderNewAPI,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeToken,
		Token:        "management-token",
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/upstream/update", map[string]any{
		"id":                      detail.Site.ID,
		"auto_create_key":         true,
		"key_quota_limit":         150,
		"key_expire_days":         45,
		"auto_sync_group":         true,
		"auto_sync_price":         false,
		"auto_checkin":            true,
		"checkin_interval_secs":   10800,
		"balance_alert_threshold": 12.5,
	}, updateUpstreamSite)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	response := decodeHandlerResponse(t, recorder)
	var site model.UpstreamSite
	if err := json.Unmarshal(response.Data, &site); err != nil {
		t.Fatalf("json.Unmarshal(site) error = %v", err)
	}
	if !site.AutoCreateKey {
		t.Fatalf("AutoCreateKey = %v, want true", site.AutoCreateKey)
	}
	if site.KeyQuotaLimit != 150 {
		t.Fatalf("KeyQuotaLimit = %v, want 150", site.KeyQuotaLimit)
	}
	if site.KeyExpireDays != 45 {
		t.Fatalf("KeyExpireDays = %v, want 45", site.KeyExpireDays)
	}
	if !site.AutoSyncGroup {
		t.Fatalf("AutoSyncGroup = %v, want true", site.AutoSyncGroup)
	}
	if site.AutoSyncPrice {
		t.Fatalf("AutoSyncPrice = %v, want false", site.AutoSyncPrice)
	}
	if site.CheckinIntervalSecs != 10800 {
		t.Fatalf("CheckinIntervalSecs = %v, want 10800", site.CheckinIntervalSecs)
	}
	if site.BalanceAlertThreshold != 12.5 {
		t.Fatalf("BalanceAlertThreshold = %v, want 12.5", site.BalanceAlertThreshold)
	}
}

func TestUpstreamSiteHealthEndpoint(t *testing.T) {
	setupHandlerTest(t)
	ctx := t.Context()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/user/self":
			_, _ = w.Write([]byte(`{"data":{"username":"owner","remain_quota":100,"used_quota":10,"group":"default"}}`))
		case "/api/user/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		case "/api/user/self/groups":
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/api/token/search":
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/api/pricing":
			_, _ = w.Write([]byte(`{"model_ratio":{}}`))
		case "/v1/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	detail, err := op.UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:         "health-test",
		ProviderType: model.UpstreamProviderNewAPI,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeToken,
		Token:        "management-token",
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/upstream/health", nil, getUpstreamSiteHealth)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	response := decodeHandlerResponse(t, recorder)
	var items []model.UpstreamHealthItem
	if err := json.Unmarshal(response.Data, &items); err != nil {
		t.Fatalf("json.Unmarshal(health) error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].ID != detail.Site.ID {
		t.Fatalf("health item ID = %d, want %d", items[0].ID, detail.Site.ID)
	}
	if items[0].Status != model.UpstreamHealthStatusHealthy {
		t.Fatalf("status = %s, want healthy", items[0].Status)
	}
}

func TestUpstreamSiteUsageEndpoint(t *testing.T) {
	setupHandlerTest(t)
	ctx := t.Context()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/user/self":
			_, _ = w.Write([]byte(`{"data":{"username":"owner","remain_quota":100,"used_quota":10,"group":"default"}}`))
		case "/api/user/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		case "/api/user/self/groups":
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/api/token/search":
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/api/pricing":
			_, _ = w.Write([]byte(`{"model_ratio":{}}`))
		case "/v1/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	detail, err := op.UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:         "usage-test",
		ProviderType: model.UpstreamProviderNewAPI,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeToken,
		Token:        "management-token",
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}

	recorder := performParamHandlerRequest(t, http.MethodGet, "/api/v1/upstream/usage/"+strconv.Itoa(detail.Site.ID), nil, map[string]string{"id": strconv.Itoa(detail.Site.ID)}, getUpstreamSiteUsage)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	response := decodeHandlerResponse(t, recorder)
	var usage model.UpstreamUsageResponse
	if err := json.Unmarshal(response.Data, &usage); err != nil {
		t.Fatalf("json.Unmarshal(usage) error = %v", err)
	}
	if usage.SiteID != detail.Site.ID {
		t.Fatalf("usage.SiteID = %d, want %d", usage.SiteID, detail.Site.ID)
	}
	if len(usage.Points) != 7 {
		t.Fatalf("len(points) = %d, want 7", len(usage.Points))
	}
}

func TestRestoreUpstreamSitePriorityEndpoint(t *testing.T) {
	setupHandlerTest(t)
	ctx := t.Context()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/user/self":
			_, _ = w.Write([]byte(`{"data":{"username":"owner","remain_quota":100,"used_quota":10,"group":"default"}}`))
		case "/api/user/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		case "/api/user/self/groups":
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/api/token/search":
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/api/pricing":
			_, _ = w.Write([]byte(`{"model_ratio":{}}`))
		case "/v1/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	detail, err := op.UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:         "restore-test",
		ProviderType: model.UpstreamProviderNewAPI,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeToken,
		Token:        "management-token",
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}

	op.UpstreamSiteSuppress(detail.Site.ID, "test")
	recorder := performParamHandlerRequest(t, http.MethodPost, "/api/v1/upstream/restore-priority/"+strconv.Itoa(detail.Site.ID), nil, map[string]string{"id": strconv.Itoa(detail.Site.ID)}, restoreUpstreamSitePriority)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if op.UpstreamSiteIsSuppressed(detail.Site.ID) {
		t.Fatalf("site should not be suppressed after restore")
	}
}

func TestCheckinUpstreamSiteEndpoint(t *testing.T) {
	setupHandlerTest(t)
	ctx := t.Context()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/user/checkin":
			_, _ = w.Write([]byte(`{"success":true,"data":{"quota":3},"message":"endpoint reward"}`))
		case "/api/user/self":
			_, _ = w.Write([]byte(`{"data":{"username":"owner","remain_quota":100,"used_quota":10,"group":"default"}}`))
		case "/api/user/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		case "/api/user/self/groups":
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/api/token/search":
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/api/pricing":
			_, _ = w.Write([]byte(`{"model_ratio":{}}`))
		case "/v1/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	detail, err := op.UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:         "handler checkin",
		ProviderType: model.UpstreamProviderNewAPI,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeToken,
		Token:        "management-token",
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/upstream/checkin", map[string]int{"id": detail.Site.ID}, checkinUpstreamSite)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	response := decodeHandlerResponse(t, recorder)
	var result model.UpstreamCheckinResult
	if err := json.Unmarshal(response.Data, &result); err != nil {
		t.Fatalf("json.Unmarshal(result) error = %v", err)
	}
	if !result.Success {
		t.Fatalf("result.Success = false, want true")
	}
	if result.Amount != 3 {
		t.Fatalf("result.Amount = %v, want 3", result.Amount)
	}
}
