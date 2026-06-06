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

func TestUpstreamSiteCreateStoresSnapshotsAndEncryptedCredential(t *testing.T) {
	ctx := setupOpTestDB(t)
	var sawLogin bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/auth/login":
			sawLogin = true
			_, _ = w.Write([]byte(`{"data":{"access_token":"jwt-owner"}}`))
		case "/api/v1/user/profile":
			_, _ = w.Write([]byte(`{"data":{"quota":100,"used":10}}`))
		case "/api/v1/keys":
			_, _ = w.Write([]byte(`{"data":[{"name":"main","key":"sk-upstream-test-key","models":["gpt-4o"],"groups":["vip"],"status":"enabled"}]}`))
		case "/api/v1/groups":
			_, _ = w.Write([]byte(`{"data":[{"name":"vip","models":["gpt-4o"],"rate_multiplier":0.7}]}`))
		case "/api/v1/subscriptions":
			_, _ = w.Write([]byte(`{"data":[{"name":"pro","balance":90}]}`))
		case "/api/v1/pricing":
			_, _ = w.Write([]byte(`{"data":[{"model":"gpt-4o","input_price":0.7,"output_price":1.4}]}`))
		case "/v1/models":
			if r.Header.Get("x-api-key") != "sk-upstream-test-key" {
				http.Error(w, "missing gateway key", http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		case "/v1/usage":
			_, _ = w.Write([]byte(`{"used":10,"remaining":90}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	detail, err := UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:          "sub2 import",
		ProviderType:  model.UpstreamProviderSub2API,
		BaseURL:       server.URL,
		AuthMode:      model.UpstreamAuthModeAccountPassword,
		Username:      "owner@example.com",
		Password:      "password-only-once",
		AutoRefresh:   true,
		SyncToChannel: true,
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}
	if !sawLogin {
		t.Fatalf("expected account login")
	}
	if detail.Site.ID == 0 || detail.Site.LinkedChannelID == 0 || detail.Site.ModelCount != 1 || detail.Site.PriceCount != 1 {
		t.Fatalf("detail.Site = %#v, want persisted linked upstream with snapshots", detail.Site)
	}
	if len(detail.Credentials) == 0 || detail.Credentials[0].EncryptedValue != "" {
		t.Fatalf("Credentials = %#v, want redacted public response", detail.Credentials)
	}
	var storedCredentials []model.UpstreamCredential
	if err := db.GetDB().WithContext(ctx).Where("upstream_site_id = ?", detail.Site.ID).Find(&storedCredentials).Error; err != nil {
		t.Fatalf("query credentials error = %v", err)
	}
	raw, _ := json.Marshal(storedCredentials)
	if strings.Contains(string(raw), "password-only-once") || strings.Contains(string(raw), "sk-upstream") || strings.Contains(string(raw), "jwt-owner") {
		t.Fatalf("stored credentials leaked raw secret: %s", raw)
	}
	var storedChannel model.Channel
	if err := db.GetDB().WithContext(ctx).Preload("Keys").First(&storedChannel, detail.Site.LinkedChannelID).Error; err != nil {
		t.Fatalf("query linked channel error = %v", err)
	}
	if storedChannel.UpstreamSiteID != detail.Site.ID || storedChannel.UpstreamSource == "" {
		t.Fatalf("linked channel upstream fields = %#v, want source linkage", storedChannel)
	}
	if len(storedChannel.Keys) != 1 || storedChannel.Keys[0].UpstreamSiteID != detail.Site.ID || storedChannel.Keys[0].AllowedModels != "gpt-4o" {
		t.Fatalf("linked channel keys = %#v, want upstream-linked model-limited key", storedChannel.Keys)
	}
	info, err := LLMGet("gpt-4o")
	if err != nil {
		t.Fatalf("LLMGet(gpt-4o) error = %v", err)
	}
	if info.OfficialInput != 0 || info.OfficialOutput != 0 {
		t.Fatalf("official price = %.2f/%.2f, want upstream price not to become official", info.OfficialInput, info.OfficialOutput)
	}
	if price, ok := ResolveGatewayLLMPrice("gpt-4o", detail.Site.LinkedChannelID); !ok || price.Input != 0.7 || price.Output != 1.4 {
		t.Fatalf("ResolveGatewayLLMPrice() = %#v/%v, want upstream gateway price", price, ok)
	}
}

func TestUpstreamSiteRefreshUsesSavedCredentialWithoutPassword(t *testing.T) {
	ctx := setupOpTestDB(t)
	modelCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/models":
			if r.Header.Get("x-api-key") != "sk-refresh" {
				http.Error(w, "missing x-api-key", http.StatusUnauthorized)
				return
			}
			modelCount++
			if modelCount == 1 {
				_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"},{"id":"deepseek-chat"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	detail, err := UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:         "refreshable",
		ProviderType: model.UpstreamProviderOpenAICompatible,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeAccessKey,
		AccessKey:    "sk-refresh",
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}
	refreshed, err := UpstreamSiteRefresh(ctx, model.UpstreamRefreshRequest{ID: detail.Site.ID, Manual: true})
	if err != nil {
		t.Fatalf("UpstreamSiteRefresh() error = %v", err)
	}
	if refreshed.Site.ModelCount != 2 {
		t.Fatalf("refreshed model count = %d, want 2", refreshed.Site.ModelCount)
	}
}

func TestUpstreamPriceSummariesKeepOfficialAndGatewaySeparate(t *testing.T) {
	ctx := setupOpTestDB(t)
	if err := LLMCreate(model.LLMInfo{
		Name:             "gpt-4o",
		OfficialLLMPrice: model.OfficialLLMPrice{OfficialInput: 5, OfficialOutput: 15},
	}, ctx); err != nil {
		t.Fatalf("LLMCreate() error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.UpstreamModelPrice{
		UpstreamSiteID: 1,
		ChannelID:      7,
		ModelName:      "gpt-4o",
		CanonicalName:  "gpt-4o",
		Input:          1.2,
		Output:         2.4,
		SourceLabel:    "中转站",
	}).Error; err != nil {
		t.Fatalf("create upstream price error = %v", err)
	}
	summaries, err := UpstreamPriceSummaries(ctx)
	if err != nil {
		t.Fatalf("UpstreamPriceSummaries() error = %v", err)
	}
	if len(summaries) != 1 || len(summaries[0].GatewayPrices) != 1 {
		t.Fatalf("summaries = %#v, want one gateway price", summaries)
	}
	if summaries[0].OfficialPrice.OfficialInput != 5 || summaries[0].GatewayPrices[0].Input != 1.2 {
		t.Fatalf("summary prices = %#v, want official and gateway separated", summaries[0])
	}
}
