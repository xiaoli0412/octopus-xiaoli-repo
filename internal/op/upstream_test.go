package op

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestUpstreamSiteCreateStoresSnapshotsAndEncryptedCredential(t *testing.T) {
	ctx := SetupOpTestDB(t)
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
	ctx := SetupOpTestDB(t)
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

func TestCreateUpstreamKeySub2API(t *testing.T) {
	ctx := SetupOpTestDB(t)
	var sawCreate bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/auth/login":
			_, _ = w.Write([]byte(`{"data":{"access_token":"jwt-owner"}}`))
		case "/api/v1/user/profile":
			_, _ = w.Write([]byte(`{"data":{"quota":100,"used":10}}`))
		case "/api/v1/keys":
			_, _ = w.Write([]byte(`{"data":[{"name":"main","key":"sk-existing-long","models":["gpt-4o"],"groups":["vip"],"status":"enabled"}]}`))
		case "/api/v1/api-keys":
			if r.Method == http.MethodPost {
				sawCreate = true
				if r.Header.Get("Authorization") != "Bearer jwt-owner" {
					t.Fatalf("Authorization = %q, want management bearer", r.Header.Get("Authorization"))
				}
				body, _ := io.ReadAll(r.Body)
				if !strings.Contains(string(body), `"name":"new-sub2-key"`) {
					t.Fatalf("request body missing name: %s", string(body))
				}
				_, _ = w.Write([]byte(`{"data":{"id":99,"name":"new-sub2-key","key":"sk-created-sub2","quota":50,"expires_at":"2026-12-31","models":"gpt-4o","groups":"vip"}}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":[{"name":"main","key":"sk-existing-long","models":["gpt-4o"],"groups":["vip"],"status":"enabled"}]}`))
		case "/api/v1/groups":
			_, _ = w.Write([]byte(`{"data":[{"name":"vip","models":["gpt-4o"],"rate_multiplier":0.7}]}`))
		case "/api/v1/subscriptions":
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/api/v1/pricing":
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/v1/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	detail, err := UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:          "sub2 create key",
		ProviderType:  model.UpstreamProviderSub2API,
		BaseURL:       server.URL,
		AuthMode:      model.UpstreamAuthModeAccountPassword,
		Username:      "owner@example.com",
		Password:      "password",
		AutoRefresh:   false,
		SyncToChannel: false,
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}

	result, err := CreateUpstreamKey(ctx, model.UpstreamCreateKeyRequest{
		SiteID:    detail.Site.ID,
		Name:      "new-sub2-key",
		Quota:     50,
		ExpiresAt: "2026-12-31",
		Models:    []string{"gpt-4o"},
		Groups:    []string{"vip"},
	})
	if err != nil {
		t.Fatalf("CreateUpstreamKey() error = %v", err)
	}
	if !sawCreate {
		t.Fatal("expected /api/v1/api-keys creation request")
	}
	if result.Key != "sk-created-sub2" {
		t.Fatalf("result.Key = %q, want sk-created-sub2", result.Key)
	}
	if !strings.Contains(result.MaskedKey, "...") {
		t.Fatalf("result.MaskedKey = %q, want masked", result.MaskedKey)
	}
}

func TestCreateUpstreamKeyNewAPI(t *testing.T) {
	ctx := SetupOpTestDB(t)
	var sawCreate bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/user/self":
			_, _ = w.Write([]byte(`{"data":{"username":"owner","remain_quota":100,"used_quota":10,"group":"default"}}`))
		case "/api/user/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		case "/api/user/self/groups":
			_, _ = w.Write([]byte(`{"data":[{"name":"default","models":["gpt-4o"],"rate_multiplier":1.0,"status":"active"}]}`))
		case "/api/token/search":
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/api/pricing":
			_, _ = w.Write([]byte(`{"model_ratio":{}}`))
		case "/api/token":
			sawCreate = true
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			if r.Header.Get("Authorization") != "Bearer management-token" {
				t.Fatalf("Authorization = %q, want management bearer", r.Header.Get("Authorization"))
			}
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"name":"new-newapi-key"`) {
				t.Fatalf("request body missing name: %s", string(body))
			}
			_, _ = w.Write([]byte(`{"success":true,"data":{"id":88,"name":"new-newapi-key","key":"sk-created-newapi"}}`))
		case "/v1/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	detail, err := UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:          "newapi create key",
		ProviderType:  model.UpstreamProviderNewAPI,
		BaseURL:       server.URL,
		AuthMode:      model.UpstreamAuthModeToken,
		Token:         "management-token",
		AutoRefresh:   false,
		SyncToChannel: false,
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}

	result, err := CreateUpstreamKey(ctx, model.UpstreamCreateKeyRequest{
		SiteID: detail.Site.ID,
		Name:   "new-newapi-key",
		Quota:  100,
		Models: []string{"gpt-4o"},
		Groups: []string{"default"},
	})
	if err != nil {
		t.Fatalf("CreateUpstreamKey() error = %v", err)
	}
	if !sawCreate {
		t.Fatal("expected /api/token creation request")
	}
	if result.Key != "sk-created-newapi" {
		t.Fatalf("result.Key = %q, want sk-created-newapi", result.Key)
	}
	if !strings.Contains(result.MaskedKey, "...") {
		t.Fatalf("result.MaskedKey = %q, want masked", result.MaskedKey)
	}
}

func TestCheckUpstreamBalanceTriggersAlertWhenBelowThreshold(t *testing.T) {
	ctx := SetupOpTestDB(t)

	site := model.UpstreamSite{
		Name:                  "alert-test",
		ProviderType:          model.UpstreamProviderNewAPI,
		BaseURL:               "http://example.com",
		Enabled:               true,
		BalanceAvailable:      true,
		BalanceRemain:         5,
		BalanceAlertThreshold: 10,
	}
	if err := db.GetDB().WithContext(ctx).Create(&site).Error; err != nil {
		t.Fatalf("create site error = %v", err)
	}

	alert, err := CheckUpstreamBalance(ctx, site.ID)
	if err != nil {
		t.Fatalf("CheckUpstreamBalance() error = %v", err)
	}
	if !alert.Alert {
		t.Fatalf("alert = %v, want true", alert.Alert)
	}
	if alert.Remain != 5 || alert.Threshold != 10 {
		t.Fatalf("alert values = %+v, want remain=5 threshold=10", alert)
	}

	var refreshed model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&refreshed, site.ID).Error; err != nil {
		t.Fatalf("query refreshed site error = %v", err)
	}
	if refreshed.LastBalanceCheckAt.IsZero() {
		t.Fatalf("LastBalanceCheckAt should be set")
	}
	if refreshed.LastBalanceValue != 5 {
		t.Fatalf("LastBalanceValue = %v, want 5", refreshed.LastBalanceValue)
	}
}

func TestCheckUpstreamBalanceNoAlertWhenUnlimited(t *testing.T) {
	ctx := SetupOpTestDB(t)

	site := model.UpstreamSite{
		Name:                  "unlimited-test",
		ProviderType:          model.UpstreamProviderNewAPI,
		BaseURL:               "http://example.com",
		Enabled:               true,
		BalanceAvailable:      true,
		BalanceUnlimited:      true,
		BalanceAlertThreshold: 10,
	}
	if err := db.GetDB().WithContext(ctx).Create(&site).Error; err != nil {
		t.Fatalf("create site error = %v", err)
	}

	alert, err := CheckUpstreamBalance(ctx, site.ID)
	if err != nil {
		t.Fatalf("CheckUpstreamBalance() error = %v", err)
	}
	if alert.Alert {
		t.Fatalf("alert = %v, want false for unlimited balance", alert.Alert)
	}
}

func TestCheckUpstreamBalanceNoAlertWhenThresholdDisabled(t *testing.T) {
	ctx := SetupOpTestDB(t)

	site := model.UpstreamSite{
		Name:             "no-threshold-test",
		ProviderType:     model.UpstreamProviderNewAPI,
		BaseURL:          "http://example.com",
		Enabled:          true,
		BalanceAvailable: true,
		BalanceRemain:    0,
	}
	if err := db.GetDB().WithContext(ctx).Create(&site).Error; err != nil {
		t.Fatalf("create site error = %v", err)
	}

	alert, err := CheckUpstreamBalance(ctx, site.ID)
	if err != nil {
		t.Fatalf("CheckUpstreamBalance() error = %v", err)
	}
	if alert.Alert {
		t.Fatalf("alert = %v, want false when threshold is zero", alert.Alert)
	}
}

func TestUpstreamSiteCreateStoresConfigFields(t *testing.T) {
	ctx := SetupOpTestDB(t)
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

	detail, err := UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:          "config create",
		ProviderType:  model.UpstreamProviderNewAPI,
		BaseURL:       server.URL,
		AuthMode:      model.UpstreamAuthModeToken,
		Token:         "management-token",
		AutoCreateKey: true,
		KeyQuotaLimit: 120,
		KeyExpireDays: 30,
		AutoSyncGroup: true,
		AutoSyncPrice: true,
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}
	if !detail.Site.AutoCreateKey {
		t.Fatalf("AutoCreateKey = %v, want true", detail.Site.AutoCreateKey)
	}
	if detail.Site.KeyQuotaLimit != 120 {
		t.Fatalf("KeyQuotaLimit = %v, want 120", detail.Site.KeyQuotaLimit)
	}
	if detail.Site.KeyExpireDays != 30 {
		t.Fatalf("KeyExpireDays = %v, want 30", detail.Site.KeyExpireDays)
	}
	if !detail.Site.AutoSyncGroup {
		t.Fatalf("AutoSyncGroup = %v, want true", detail.Site.AutoSyncGroup)
	}
	if !detail.Site.AutoSyncPrice {
		t.Fatalf("AutoSyncPrice = %v, want true", detail.Site.AutoSyncPrice)
	}

	var stored model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&stored, detail.Site.ID).Error; err != nil {
		t.Fatalf("query stored site error = %v", err)
	}
	if !stored.AutoCreateKey || stored.KeyQuotaLimit != 120 || stored.KeyExpireDays != 30 || !stored.AutoSyncGroup || !stored.AutoSyncPrice {
		t.Fatalf("stored site config = %#v, want persisted values", stored)
	}
}

func TestUpstreamSiteUpdateStoresConfigFields(t *testing.T) {
	ctx := SetupOpTestDB(t)

	site := model.UpstreamSite{
		Name:         "update-config",
		ProviderType: model.UpstreamProviderNewAPI,
		BaseURL:      "http://example.com",
		Enabled:      true,
	}
	if err := db.GetDB().WithContext(ctx).Create(&site).Error; err != nil {
		t.Fatalf("create site error = %v", err)
	}

	autoCreateKey := true
	keyQuotaLimit := 200.0
	keyExpireDays := 60
	autoSyncGroup := true
	autoSyncPrice := false
	autoCheckin := true
	checkinInterval := 7200
	balanceThreshold := 8.5
	updated, err := UpstreamSiteUpdate(ctx, model.UpstreamSiteUpdateRequest{
		ID:                    site.ID,
		AutoCreateKey:         &autoCreateKey,
		KeyQuotaLimit:         &keyQuotaLimit,
		KeyExpireDays:         &keyExpireDays,
		AutoSyncGroup:         &autoSyncGroup,
		AutoSyncPrice:         &autoSyncPrice,
		AutoCheckin:           &autoCheckin,
		CheckinIntervalSecs:   &checkinInterval,
		BalanceAlertThreshold: &balanceThreshold,
	})
	if err != nil {
		t.Fatalf("UpstreamSiteUpdate() error = %v", err)
	}
	if !updated.AutoCreateKey {
		t.Fatalf("AutoCreateKey = %v, want true", updated.AutoCreateKey)
	}
	if updated.KeyQuotaLimit != 200 {
		t.Fatalf("KeyQuotaLimit = %v, want 200", updated.KeyQuotaLimit)
	}
	if updated.KeyExpireDays != 60 {
		t.Fatalf("KeyExpireDays = %v, want 60", updated.KeyExpireDays)
	}
	if !updated.AutoSyncGroup {
		t.Fatalf("AutoSyncGroup = %v, want true", updated.AutoSyncGroup)
	}
	if updated.AutoSyncPrice {
		t.Fatalf("AutoSyncPrice = %v, want false", updated.AutoSyncPrice)
	}
	if updated.CheckinIntervalSecs != 7200 {
		t.Fatalf("CheckinIntervalSecs = %v, want 7200", updated.CheckinIntervalSecs)
	}
	if updated.BalanceAlertThreshold != 8.5 {
		t.Fatalf("BalanceAlertThreshold = %v, want 8.5", updated.BalanceAlertThreshold)
	}
}

func TestUpstreamSiteUpdateSetsBalanceAlertThreshold(t *testing.T) {
	ctx := SetupOpTestDB(t)

	site := model.UpstreamSite{
		Name:         "update-threshold-test",
		ProviderType: model.UpstreamProviderNewAPI,
		BaseURL:      "http://example.com",
		Enabled:      true,
	}
	if err := db.GetDB().WithContext(ctx).Create(&site).Error; err != nil {
		t.Fatalf("create site error = %v", err)
	}

	threshold := 20.5
	updated, err := UpstreamSiteUpdate(ctx, model.UpstreamSiteUpdateRequest{
		ID:                    site.ID,
		BalanceAlertThreshold: &threshold,
	})
	if err != nil {
		t.Fatalf("UpstreamSiteUpdate() error = %v", err)
	}
	if updated.BalanceAlertThreshold != 20.5 {
		t.Fatalf("BalanceAlertThreshold = %v, want 20.5", updated.BalanceAlertThreshold)
	}
}

func TestUpstreamPriceSummariesKeepOfficialAndGatewaySeparate(t *testing.T) {
	ctx := SetupOpTestDB(t)
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

func TestUpstreamSiteCheckinNewAPI(t *testing.T) {
	ctx := SetupOpTestDB(t)
	var sawCheckin bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/user/checkin":
			sawCheckin = true
			if r.Method != http.MethodPost {
				t.Errorf("checkin method = %s, want POST", r.Method)
			}
			if r.Header.Get("Authorization") != "Bearer management-token" {
				t.Errorf("Authorization = %q, want management bearer", r.Header.Get("Authorization"))
			}
			_, _ = w.Write([]byte(`{"success":true,"data":{"quota":5.5},"message":"签到成功"}`))
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

	detail, err := UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:         "newapi checkin",
		ProviderType: model.UpstreamProviderNewAPI,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeToken,
		Token:        "management-token",
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}

	result, err := UpstreamSiteCheckin(ctx, detail.Site.ID)
	if err != nil {
		t.Fatalf("UpstreamSiteCheckin() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("result.Success = false, want true")
	}
	if result.Amount != 5.5 {
		t.Fatalf("result.Amount = %v, want 5.5", result.Amount)
	}
	if !sawCheckin {
		t.Fatal("expected /api/user/checkin request")
	}

	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, detail.Site.ID).Error; err != nil {
		t.Fatalf("query site error = %v", err)
	}
	if site.LastCheckinAt.IsZero() {
		t.Fatal("LastCheckinAt should be set")
	}
	var entries []model.UpstreamCheckinLogEntry
	if err := json.Unmarshal([]byte(site.CheckinLog), &entries); err != nil {
		t.Fatalf("unmarshal checkin log error = %v", err)
	}
	if len(entries) != 1 || !entries[0].Success || entries[0].Amount != 5.5 {
		t.Fatalf("checkin log = %#v, want one success entry with amount 5.5", entries)
	}
}

func TestUpstreamSiteCheckinSub2API(t *testing.T) {
	ctx := SetupOpTestDB(t)
	var sawCheckin bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/user/checkin":
			sawCheckin = true
			if r.Method != http.MethodPost {
				t.Errorf("checkin method = %s, want POST", r.Method)
			}
			if r.Header.Get("Authorization") != "Bearer jwt-owner" {
				t.Errorf("Authorization = %q, want owner bearer", r.Header.Get("Authorization"))
			}
			_, _ = w.Write([]byte(`{"success":true,"amount":10,"message":"sub2 reward"}`))
		case "/api/v1/auth/login":
			_, _ = w.Write([]byte(`{"data":{"access_token":"jwt-owner"}}`))
		case "/api/v1/user/profile":
			_, _ = w.Write([]byte(`{"data":{"quota":100,"used":10}}`))
		case "/api/v1/keys":
			_, _ = w.Write([]byte(`{"data":[{"name":"main","key":"sk-existing-long","models":["gpt-4o"],"groups":["vip"],"status":"enabled"}]}`))
		case "/api/v1/groups":
			_, _ = w.Write([]byte(`{"data":[{"name":"vip","models":["gpt-4o"],"rate_multiplier":0.7}]}`))
		case "/api/v1/subscriptions":
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/api/v1/pricing":
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/v1/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	detail, err := UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:          "sub2 checkin",
		ProviderType:  model.UpstreamProviderSub2API,
		BaseURL:       server.URL,
		AuthMode:      model.UpstreamAuthModeAccountPassword,
		Username:      "owner@example.com",
		Password:      "password",
		SyncToChannel: false,
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}

	result, err := UpstreamSiteCheckin(ctx, detail.Site.ID)
	if err != nil {
		t.Fatalf("UpstreamSiteCheckin() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("result.Success = false, want true")
	}
	if result.Amount != 10 {
		t.Fatalf("result.Amount = %v, want 10", result.Amount)
	}
	if !sawCheckin {
		t.Fatal("expected /api/v1/user/checkin request")
	}

	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, detail.Site.ID).Error; err != nil {
		t.Fatalf("query site error = %v", err)
	}
	if site.LastCheckinAt.IsZero() {
		t.Fatal("LastCheckinAt should be set")
	}
}

func TestUpstreamSiteCheckinFailureRecordsLog(t *testing.T) {
	ctx := SetupOpTestDB(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/user/checkin":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"success":false,"message":"already checked in"}`))
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

	detail, err := UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:         "newapi checkin fail",
		ProviderType: model.UpstreamProviderNewAPI,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeToken,
		Token:        "management-token",
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}

	result, err := UpstreamSiteCheckin(ctx, detail.Site.ID)
	if err == nil {
		t.Fatal("UpstreamSiteCheckin() expected error")
	}
	if result.Success {
		t.Fatal("result.Success = true, want false")
	}

	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, detail.Site.ID).Error; err != nil {
		t.Fatalf("query site error = %v", err)
	}
	var entries []model.UpstreamCheckinLogEntry
	if err := json.Unmarshal([]byte(site.CheckinLog), &entries); err != nil {
		t.Fatalf("unmarshal checkin log error = %v", err)
	}
	if len(entries) != 1 || entries[0].Success {
		t.Fatalf("checkin log = %#v, want one failure entry", entries)
	}
}

func TestUpstreamSiteCreateAutoCreatesKeyWhenEnabled(t *testing.T) {
	ctx := SetupOpTestDB(t)
	var sawCreate bool
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
		case "/api/token":
			if r.Method == http.MethodPost {
				sawCreate = true
				_, _ = w.Write([]byte(`{"success":true,"data":{"id":88,"name":"auto-created","key":"sk-auto-created"}}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/v1/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	detail, err := UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:          "auto-create-key",
		ProviderType:  model.UpstreamProviderNewAPI,
		BaseURL:       server.URL,
		AuthMode:      model.UpstreamAuthModeToken,
		Token:         "management-token",
		SyncToChannel: true,
		AutoCreateKey: true,
		KeyQuotaLimit: 100,
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}
	if !sawCreate {
		t.Fatal("expected /api/token creation request")
	}
	if detail.Site.LinkedChannelID == 0 {
		t.Fatalf("expected linked channel, got %d", detail.Site.LinkedChannelID)
	}
	channel, err := ChannelGet(detail.Site.LinkedChannelID, ctx)
	if err != nil {
		t.Fatalf("ChannelGet() error = %v", err)
	}
	if len(channel.Keys) != 1 || channel.Keys[0].ChannelKey != "sk-auto-created" {
		t.Fatalf("channel keys = %#v, want one auto-created key", channel.Keys)
	}
}

func TestUpstreamSiteCreateAutoSyncsGroups(t *testing.T) {
	ctx := SetupOpTestDB(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/user/self":
			_, _ = w.Write([]byte(`{"data":{"username":"owner","remain_quota":100,"used_quota":10,"group":"default"}}`))
		case "/api/user/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		case "/api/user/self/groups":
			_, _ = w.Write([]byte(`{"data":[{"name":"vip","models":["gpt-4o"],"rate_multiplier":0.7,"status":"active"}]}`))
		case "/api/token/search":
			_, _ = w.Write([]byte(`{"data":[{"name":"main","key":"sk-existing-long","models":["gpt-4o"],"groups":["vip"],"status":"enabled"}]}`))
		case "/api/pricing":
			_, _ = w.Write([]byte(`{"model_ratio":{}}`))
		case "/v1/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	detail, err := UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:          "auto-sync-group",
		ProviderType:  model.UpstreamProviderNewAPI,
		BaseURL:       server.URL,
		AuthMode:      model.UpstreamAuthModeToken,
		Token:         "management-token",
		SyncToChannel: true,
		AutoSyncGroup: true,
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}
	if detail.Site.LinkedChannelID == 0 {
		t.Fatalf("expected linked channel")
	}
	groups, err := GroupList(ctx)
	if err != nil {
		t.Fatalf("GroupList() error = %v", err)
	}
	var found bool
	for _, group := range groups {
		if group.Name == "vip" {
			found = true
			if len(group.Items) != 1 || group.Items[0].ModelName != "gpt-4o" || group.Items[0].ChannelID != detail.Site.LinkedChannelID {
				t.Fatalf("group items = %#v, want one item for gpt-4o in linked channel", group.Items)
			}
		}
	}
	if !found {
		t.Fatal("expected vip group synced from upstream")
	}
}

func TestUpstreamSiteCreateAutoSyncsPrice(t *testing.T) {
	ctx := SetupOpTestDB(t)
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
			_, _ = w.Write([]byte(`{"data":[{"name":"main","key":"sk-existing-long","models":["gpt-4o"],"status":"enabled"}]}`))
		case "/api/pricing":
			_, _ = w.Write([]byte(`{"model_ratio":{"gpt-4o":0.7},"completion_ratio":{"gpt-4o":2.0}}`))
		case "/v1/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	detail, err := UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:          "auto-sync-price",
		ProviderType:  model.UpstreamProviderNewAPI,
		BaseURL:       server.URL,
		AuthMode:      model.UpstreamAuthModeToken,
		Token:         "management-token",
		SyncToChannel: true,
		AutoSyncPrice: true,
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}
	if detail.Site.LinkedChannelID == 0 {
		t.Fatalf("expected linked channel")
	}
	info, err := LLMGet("gpt-4o")
	if err != nil {
		t.Fatalf("LLMGet(gpt-4o) error = %v", err)
	}
	if info.Input != 0.7 || info.Output != 1.4 {
		t.Fatalf("LLMPrice = %.4f/%.4f, want 0.7/1.4", info.Input, info.Output)
	}
	if info.OfficialInput != 0 || info.OfficialOutput != 0 {
		t.Fatalf("official price should not be set by upstream sync")
	}
}

func TestUpstreamSiteCreateDoesNotSyncPriceWhenDisabled(t *testing.T) {
	ctx := SetupOpTestDB(t)
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
			_, _ = w.Write([]byte(`{"data":[{"name":"main","key":"sk-existing-long","models":["gpt-4o"],"status":"enabled"}]}`))
		case "/api/pricing":
			_, _ = w.Write([]byte(`{"model_ratio":{"gpt-4o":0.7},"completion_ratio":{"gpt-4o":2.0}}`))
		case "/v1/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	_, err := UpstreamSiteCreate(ctx, model.UpstreamSiteCreateRequest{
		Name:          "no-auto-sync-price",
		ProviderType:  model.UpstreamProviderNewAPI,
		BaseURL:       server.URL,
		AuthMode:      model.UpstreamAuthModeToken,
		Token:         "management-token",
		SyncToChannel: true,
		AutoSyncPrice: false,
	})
	if err != nil {
		t.Fatalf("UpstreamSiteCreate() error = %v", err)
	}
	info, err := LLMGet("gpt-4o")
	if err != nil {
		t.Fatalf("LLMGet(gpt-4o) error = %v", err)
	}
	if info.Input != 0 || info.Output != 0 {
		t.Fatalf("LLMPrice = %.4f/%.4f, want zero when auto sync price is disabled", info.Input, info.Output)
	}
}
