package op

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestInspectNewAPIReadsModelsAndTokenUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}
		switch r.URL.Path {
		case "/v1/models":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"gpt-4o"},{"id":"gemini-2.5-pro"},{"id":"text-embedding-3-small"}]}`))
		case "/api/usage/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"total_used":12.5,"total_available":87.5,"unlimited_quota":false}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	result, err := InspectNewAPI(context.Background(), model.NewAPIInspectRequest{
		BaseURL: server.URL,
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("InspectNewAPI() error = %v", err)
	}
	if result.APIBaseURL != server.URL+"/v1" {
		t.Fatalf("APIBaseURL = %q, want %s/v1", result.APIBaseURL, server.URL)
	}
	if result.ModelCount != 3 {
		t.Fatalf("ModelCount = %d, want 3", result.ModelCount)
	}
	if !result.TokenUsage.Available || result.TokenUsage.UsedQuota != 12.5 || result.TokenUsage.RemainQuota != 87.5 {
		t.Fatalf("TokenUsage = %#v, want parsed New API usage", result.TokenUsage)
	}
	if !newAPITestContainsString(result.RequestCapabilities, model.RequestCapabilityGeminiContents) {
		t.Fatalf("RequestCapabilities = %#v, want gemini capability", result.RequestCapabilities)
	}
	if !newAPITestContainsString(result.RequestCapabilities, model.RequestCapabilityOpenAIEmbeddings) {
		t.Fatalf("RequestCapabilities = %#v, want embeddings capability", result.RequestCapabilities)
	}
}

func TestInspectNewAPIWarnsWhenUsageUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	result, err := InspectNewAPI(context.Background(), model.NewAPIInspectRequest{
		BaseURL: server.URL + "/v1",
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("InspectNewAPI() error = %v", err)
	}
	if result.BaseURL != server.URL || result.APIBaseURL != server.URL+"/v1" {
		t.Fatalf("base URLs = %#v, want normalized site and API base", result)
	}
	if result.TokenUsage.Available {
		t.Fatalf("TokenUsage.Available = true, want false")
	}
	if len(result.Warnings) == 0 {
		t.Fatalf("Warnings = empty, want usage warning")
	}
}

func TestInspectUpstreamGatewaySub2APILoginImportsGatewayKey(t *testing.T) {
	var sawLogin bool
	var sawModelsWithXAPIKey bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/auth/login":
			sawLogin = true
			_, _ = w.Write([]byte(`{"data":{"access_token":"jwt-management"}}`))
		case "/api/v1/user/profile":
			if r.Header.Get("Authorization") != "Bearer jwt-management" {
				t.Fatalf("profile Authorization = %q, want management JWT", r.Header.Get("Authorization"))
			}
			_, _ = w.Write([]byte(`{"data":{"quota":42,"used":8}}`))
		case "/api/v1/keys":
			if r.Header.Get("Authorization") != "Bearer jwt-management" {
				t.Fatalf("keys Authorization = %q, want management JWT", r.Header.Get("Authorization"))
			}
			_, _ = w.Write([]byte(`{"data":[{"name":"primary","key":"sk-sub2-gateway","models":["gpt-4o","qwen3-coder"],"groups":["vip-package"]}]}`))
		case "/v1/models":
			if r.Header.Get("Authorization") != "" {
				http.Error(w, "bearer rejected", http.StatusUnauthorized)
				return
			}
			if r.Header.Get("x-api-key") == "sk-sub2-gateway" {
				sawModelsWithXAPIKey = true
			}
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"},{"id":"qwen3-coder"}]}`))
		case "/v1/usage":
			_, _ = w.Write([]byte(`{"remaining":34,"used":8}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	result, err := InspectUpstreamGateway(context.Background(), model.UpstreamInspectRequest{
		ProviderType: model.UpstreamProviderSub2API,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeAccountPassword,
		Username:     "owner@example.com",
		Password:     "secret",
	})
	if err != nil {
		t.Fatalf("InspectUpstreamGateway() error = %v", err)
	}
	if !sawLogin || !sawModelsWithXAPIKey {
		t.Fatalf("expected login and x-api-key model probe, got login=%v x-api-key=%v", sawLogin, sawModelsWithXAPIKey)
	}
	if result.ProviderType != model.UpstreamProviderSub2API || result.ModelCount != 2 {
		t.Fatalf("result = %#v, want sub2api with 2 models", result)
	}
	if len(result.Keys) != 1 || result.Keys[0].MaskedKey == "" || result.Keys[0].Key != "sk-sub2-gateway" {
		t.Fatalf("Keys = %#v, want imported gateway key with masked label", result.Keys)
	}
	if len(result.Keys[0].RequestCapabilities) != 0 {
		t.Fatalf("key request capabilities = %#v, want no protocol restriction when upstream does not declare one", result.Keys[0].RequestCapabilities)
	}
	if len(result.Keys[0].Groups) != 1 || result.Keys[0].Groups[0] != "vip-package" {
		t.Fatalf("key groups = %#v, want upstream group metadata preserved for display", result.Keys[0].Groups)
	}
	if newAPITestContainsString(result.Keys[0].AllowedModels, "vip-package") {
		t.Fatalf("AllowedModels = %#v, want upstream groups excluded from model restrictions", result.Keys[0].AllowedModels)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal inspect result error = %v", err)
	}
	if strings.Contains(string(payload), "sk-sub2-gateway") {
		t.Fatalf("serialized inspect result leaked raw gateway key: %s", payload)
	}
	if !result.TokenUsage.Available || result.TokenUsage.RemainQuota != 34 || result.TokenUsage.UsedQuota != 8 {
		t.Fatalf("TokenUsage = %#v, want gateway usage to override profile usage", result.TokenUsage)
	}
	if len(result.PriceCandidates) != 2 {
		t.Fatalf("PriceCandidates len = %d, want 2", len(result.PriceCandidates))
	}
}

func TestInspectUpstreamGatewayNewAPIManagementReadsCatalog(t *testing.T) {
	var sawUserHeader bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			if r.Header.Get("Authorization") != "Bearer management-token" {
				t.Fatalf("Authorization = %q, want management bearer", r.Header.Get("Authorization"))
			}
			if r.Header.Get("New-Api-User") == "42" {
				sawUserHeader = true
			}
		}
		switch r.URL.Path {
		case "/api/user/self":
			_, _ = w.Write([]byte(`{"data":{"username":"owner","remain_quota":88,"used_quota":12,"group":"vip"}}`))
		case "/api/user/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"},{"id":"deepseek-chat"},{"id":"qwen3-coder"},{"id":"text-embedding-3-small"}]}`))
		case "/api/user/self/groups":
			_, _ = w.Write([]byte(`{"data":[{"name":"vip","models":["gpt-4o"],"rate_multiplier":0.8,"status":"active"}]}`))
		case "/api/token/search":
			_, _ = w.Write([]byte(`{"data":[{"id":7,"name":"masked","key":"reHR********OspA","group":"vip","status":"enabled","remain_quota":0,"used_quota":0}]}`))
		case "/api/token/self":
			_, _ = w.Write([]byte(`{"data":{"used_quota":99,"remain_quota":1,"status":"token-self"}}`))
		case "/api/pricing":
			_, _ = w.Write([]byte(`{"model_ratio":{"gpt-4o":0.5},"completion_ratio":{"gpt-4o":2}}`))
		case "/v1/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	result, err := InspectUpstreamGateway(context.Background(), model.UpstreamInspectRequest{
		ProviderType: model.UpstreamProviderNewAPI,
		BaseURL:      server.URL,
		AuthMode:     model.UpstreamAuthModeToken,
		Token:        "management-token",
		UserID:       "42",
	})
	if err != nil {
		t.Fatalf("InspectUpstreamGateway() error = %v", err)
	}
	if !sawUserHeader {
		t.Fatalf("expected New-Api-User header to be sent")
	}
	if result.ModelCount != 4 || !newAPITestContainsString(result.Models, "qwen3-coder") || !newAPITestContainsString(result.Models, "text-embedding-3-small") {
		t.Fatalf("Models = %#v, want merged management model catalog plus gateway probe", result.Models)
	}
	if !newAPITestContainsString(result.RequestCapabilities, model.RequestCapabilityOpenAIEmbeddings) {
		t.Fatalf("RequestCapabilities = %#v, want embedding capability from management model catalog", result.RequestCapabilities)
	}
	if !result.TokenUsage.Available || result.TokenUsage.RemainQuota != 88 || result.TokenUsage.UsedQuota != 12 {
		t.Fatalf("TokenUsage = %#v, want account balance to remain preferred over token self records", result.TokenUsage)
	}
	if len(result.Groups) != 1 || result.Groups[0].Name != "vip" || result.Groups[0].RateMultiplier != 0.8 {
		t.Fatalf("Groups = %#v, want parsed New API group", result.Groups)
	}
	if len(result.Keys) != 1 || result.Keys[0].Importable || result.Keys[0].Key != "" || !strings.HasPrefix(result.Keys[0].MaskedKey, "sk-") {
		t.Fatalf("Keys = %#v, want masked non-importable token list entry", result.Keys)
	}
	if len(result.PriceCandidates) != 4 {
		t.Fatalf("PriceCandidates len = %d, want model candidates", len(result.PriceCandidates))
	}
	var gptPrice model.UpstreamPriceCandidate
	for _, item := range result.PriceCandidates {
		if item.Name == "gpt-4o" {
			gptPrice = item
			break
		}
	}
	if gptPrice.Input != 0.5 || gptPrice.Output != 1 || !strings.Contains(gptPrice.PriceSource, "ratio") {
		t.Fatalf("gpt-4o price candidate = %#v, want upstream ratio price", gptPrice)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal result error = %v", err)
	}
	if strings.Contains(string(payload), "reHR********OspA") && !strings.Contains(string(payload), "sk-reHR********OspA") {
		t.Fatalf("serialized result contains unnormalized masked token: %s", payload)
	}
	if strings.Contains(string(payload), "management-token") {
		t.Fatalf("serialized result leaked management token: %s", payload)
	}
}

func TestApplyUpstreamGatewayCreatesChannelAndPriceMetadata(t *testing.T) {
	ctx := setupOpTestDB(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/models":
			if r.Header.Get("x-api-key") != "sk-apply" {
				t.Fatalf("x-api-key = %q, want gateway access key", r.Header.Get("x-api-key"))
			}
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"},{"id":"text-embedding-3-small"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	applied, err := ApplyUpstreamGateway(ctx, model.UpstreamApplyRequest{
		Inspect: model.UpstreamInspectRequest{
			ProviderType: model.UpstreamProviderNewAPI,
			BaseURL:      server.URL,
			AuthMode:     model.UpstreamAuthModeAccessKey,
			AccessKey:    "sk-apply",
		},
		ChannelName: "New API imported",
	})
	if err != nil {
		t.Fatalf("ApplyUpstreamGateway() error = %v", err)
	}
	if !applied.Created || applied.Channel.ID == 0 {
		t.Fatalf("ApplyUpstreamGateway() result = %#v, want created channel", applied)
	}
	var stored model.Channel
	if err := db.GetDB().WithContext(ctx).Preload("Keys").First(&stored, applied.Channel.ID).Error; err != nil {
		t.Fatalf("query channel error = %v", err)
	}
	if stored.Name != "New API imported" || stored.CustomModel != "gpt-4o,text-embedding-3-small" {
		t.Fatalf("stored channel = %#v, want normalized imported models", stored)
	}
	if len(stored.BaseUrls) != 1 || stored.BaseUrls[0].URL != server.URL+"/v1" {
		t.Fatalf("BaseUrls = %#v, want normalized /v1 API base", stored.BaseUrls)
	}
	if len(stored.Keys) != 1 || stored.Keys[0].AllowedModels != "gpt-4o,text-embedding-3-small" {
		t.Fatalf("stored keys = %#v, want one key with imported allowed models", stored.Keys)
	}
	payload, err := json.Marshal(applied)
	if err != nil {
		t.Fatalf("marshal apply result error = %v", err)
	}
	if strings.Contains(string(payload), "sk-apply") {
		t.Fatalf("serialized apply result leaked raw channel key: %s", payload)
	}
	embedding, err := LLMGet("text-embedding-3-small")
	if err != nil {
		t.Fatalf("LLMGet(embedding) error = %v", err)
	}
	supported, policy, _ := model.InferCacheSupport(embedding)
	if supported || policy != model.CachePolicyUnsupported {
		t.Fatalf("embedding cache support = %v/%s, want unsupported", supported, policy)
	}
}

func TestApplyUpstreamGatewayRejectsManagementTokenAsChannelKey(t *testing.T) {
	ctx := setupOpTestDB(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/models":
			if r.Header.Get("Authorization") != "Bearer admin-token" {
				t.Fatalf("Authorization = %q, want management token probe", r.Header.Get("Authorization"))
			}
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	_, err := ApplyUpstreamGateway(ctx, model.UpstreamApplyRequest{
		Inspect: model.UpstreamInspectRequest{
			ProviderType: model.UpstreamProviderNewAPI,
			BaseURL:      server.URL,
			AuthMode:     model.UpstreamAuthModeToken,
			Token:        "admin-token",
		},
		ChannelName: "should-not-import-admin-token",
	})
	if err == nil || !strings.Contains(err.Error(), "gateway key") {
		t.Fatalf("ApplyUpstreamGateway() error = %v, want gateway key rejection", err)
	}
}

func TestOpsCacheUnsupportedModelsExcludedFromCacheRateDenominator(t *testing.T) {
	ctx := setupOpTestDB(t)
	if err := LLMCreate(model.LLMInfo{Name: "gpt-4o", CachePolicy: model.CachePolicySupported}, ctx); err != nil {
		t.Fatalf("LLMCreate(gpt-4o) error = %v", err)
	}
	if err := LLMCreate(model.LLMInfo{Name: "text-embedding-3-small", CachePolicy: model.CachePolicyUnsupported}, ctx); err != nil {
		t.Fatalf("LLMCreate(embedding) error = %v", err)
	}
	if err := OpsRecordRelay(ctx, OpsRelayEvent{
		RequestModel:     "gpt-4o",
		ActualModel:      "gpt-4o",
		Success:          true,
		CacheReadTokens:  128,
		CacheWriteTokens: 64,
	}); err != nil {
		t.Fatalf("OpsRecordRelay(gpt-4o) error = %v", err)
	}
	if err := OpsRecordRelay(ctx, OpsRelayEvent{
		RequestModel: "text-embedding-3-small",
		ActualModel:  "text-embedding-3-small",
		Success:      true,
	}); err != nil {
		t.Fatalf("OpsRecordRelay(embedding) error = %v", err)
	}

	total, err := opsOverallSummary(ctx)
	if err != nil {
		t.Fatalf("opsOverallSummary() error = %v", err)
	}
	if total.CacheEligibleCount != 1 || total.CacheIneligibleCount != 1 {
		t.Fatalf("cache eligibility = %d/%d, want 1 eligible and 1 ineligible", total.CacheEligibleCount, total.CacheIneligibleCount)
	}
	if total.CacheHitRate != 1 || total.CacheCreateRate != 1 || total.CacheRate != 1 {
		t.Fatalf("cache rates = hit %.2f create %.2f rate %.2f, want denominator to exclude unsupported model", total.CacheHitRate, total.CacheCreateRate, total.CacheRate)
	}

	models, err := OpsEntityList(ctx, model.OpsScopeModel, 10)
	if err != nil {
		t.Fatalf("OpsEntityList(model) error = %v", err)
	}
	byLabel := make(map[string]model.OpsEntitySummary, len(models))
	for _, item := range models {
		byLabel[item.EntityLabel] = item
	}
	embedding := byLabel["text-embedding-3-small"]
	if embedding.CacheSupported || embedding.CacheEligibleCount != 0 || embedding.CacheIneligibleCount != 1 {
		payload, _ := json.Marshal(models)
		t.Fatalf("embedding summary = %#v in %s, want unsupported and ineligible", embedding, payload)
	}
}

func newAPITestContainsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
