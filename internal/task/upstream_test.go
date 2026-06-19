package task

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

func TestUpstreamRefreshTaskChecksBalanceAndSkipsWhenNotDue(t *testing.T) {
	setupTaskTestDB(t)
	ctx := t.Context()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/user/self":
			_, _ = w.Write([]byte(`{"data":{"username":"owner","remain_quota":8,"used_quota":2,"group":"default"}}`))
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
		Name:                "task-balance-check",
		ProviderType:        model.UpstreamProviderNewAPI,
		BaseURL:             server.URL,
		AuthMode:            model.UpstreamAuthModeToken,
		Token:               "management-token",
		AutoRefresh:         true,
		RefreshIntervalSecs: 3600,
		SyncToChannel:       false,
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}

	threshold := 10.0
	if _, err := op.UpstreamSiteUpdate(ctx, model.UpstreamSiteUpdateRequest{
		ID:                    detail.Site.ID,
		BalanceAlertThreshold: &threshold,
	}); err != nil {
		t.Fatalf("UpstreamSiteUpdate() error = %v", err)
	}

	// Move last refresh far into the past so the task will run.
	past := time.Now().Add(-2 * time.Hour)
	if err := db.GetDB().WithContext(ctx).Model(&model.UpstreamSite{}).Where("id = ?", detail.Site.ID).Update("last_refresh_at", past).Error; err != nil {
		t.Fatalf("update last_refresh_at error = %v", err)
	}
	UpstreamRefreshTask()

	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, detail.Site.ID).Error; err != nil {
		t.Fatalf("query site error = %v", err)
	}
	if site.LastRefreshStatus != "success" {
		t.Fatalf("LastRefreshStatus = %q, want success", site.LastRefreshStatus)
	}
	if site.LastBalanceCheckAt.IsZero() {
		t.Fatalf("LastBalanceCheckAt should be set after refresh task")
	}
	if site.LastBalanceValue != 8 {
		t.Fatalf("LastBalanceValue = %v, want 8", site.LastBalanceValue)
	}
}

func TestUpstreamCheckinTaskRunsForDueSite(t *testing.T) {
	setupTaskTestDB(t)
	ctx := t.Context()

	var sawCheckin bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/user/checkin":
			sawCheckin = true
			_, _ = w.Write([]byte(`{"success":true,"data":{"quota":7},"message":"task reward"}`))
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
		Name:                "task-checkin",
		ProviderType:        model.UpstreamProviderNewAPI,
		BaseURL:             server.URL,
		AuthMode:            model.UpstreamAuthModeToken,
		Token:               "management-token",
		AutoCheckin:         true,
		CheckinIntervalSecs: 3600,
		SyncToChannel:       false,
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}

	past := time.Now().Add(-2 * time.Hour)
	if err := db.GetDB().WithContext(ctx).Model(&model.UpstreamSite{}).Where("id = ?", detail.Site.ID).Update("last_checkin_at", past).Error; err != nil {
		t.Fatalf("update last_checkin_at error = %v", err)
	}

	UpstreamCheckinTask()

	if !sawCheckin {
		t.Fatal("expected /api/user/checkin request from task")
	}

	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, detail.Site.ID).Error; err != nil {
		t.Fatalf("query site error = %v", err)
	}
	if site.LastCheckinAt.IsZero() || !site.LastCheckinAt.After(past) {
		t.Fatalf("LastCheckinAt = %v, should be updated after task", site.LastCheckinAt)
	}

	var entries []model.UpstreamCheckinLogEntry
	if err := json.Unmarshal([]byte(site.CheckinLog), &entries); err != nil {
		t.Fatalf("unmarshal checkin log error = %v", err)
	}
	if len(entries) != 1 || !entries[0].Success || entries[0].Amount != 7 {
		t.Fatalf("checkin log = %#v, want one success entry with amount 7", entries)
	}
}
