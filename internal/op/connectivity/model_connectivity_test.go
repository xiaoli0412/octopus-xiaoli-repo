package connectivity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
)

func setupConnectivityTestDB(t *testing.T) context.Context {
	t.Helper()
	return op.SetupOpTestDB(t)
}

func TestTestModelConnectivitySuccess(t *testing.T) {
	ctx := setupConnectivityTestDB(t)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer sk-test" {
			t.Errorf("unexpected authorization: %s", auth)
		}

		resp := map[string]any{
			"id":      "chatcmpl-test",
			"object":  "chat.completion",
			"created": 1234567890,
			"model":   "gpt-4o",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "hello back",
					},
					"finish_reason": "stop",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer upstream.Close()

	channel := &model.Channel{
		Name:     "test-connectivity-channel",
		Enabled:  true,
		Type:     outbound.OutboundTypeOpenAIChat,
		BaseUrls: []model.BaseUrl{{URL: upstream.URL}},
		Model:    "gpt-4o",
		Keys: []model.ChannelKey{{
			Enabled:    true,
			ChannelKey: "sk-test",
		}},
	}
	if err := op.ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	if err := op.LLMCreate(model.LLMInfo{Name: "gpt-4o", LLMPrice: model.LLMPrice{Input: 0.01, Output: 0.02}}, ctx); err != nil {
		t.Fatalf("LLMCreate() error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.UpstreamModelPrice{
		UpstreamSiteID: 1,
		ChannelID:      channel.ID,
		ModelName:      "gpt-4o",
		Input:          0.01,
		Output:         0.02,
	}).Error; err != nil {
		t.Fatalf("create upstream price error = %v", err)
	}

	result, err := TestModelConnectivity(ctx, model.ModelTestRequest{ChannelID: channel.ID, Model: "gpt-4o"})
	if err != nil {
		t.Fatalf("TestModelConnectivity() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.ErrorMessage)
	}
	if result.LatencyMs < 0 {
		t.Fatalf("latency should be non-negative, got %d", result.LatencyMs)
	}
	if !strings.Contains(result.ResponseText, "hello back") {
		t.Fatalf("response text = %q, want hello back", result.ResponseText)
	}
	if !result.PriceMatch {
		t.Fatalf("expected price match")
	}
}

func TestTestModelConnectivityChannelNotFound(t *testing.T) {
	ctx := setupConnectivityTestDB(t)

	result, err := TestModelConnectivity(ctx, model.ModelTestRequest{ChannelID: 999, Model: "gpt-4o"})
	if err != nil {
		t.Fatalf("TestModelConnectivity() error = %v", err)
	}
	if result.Success {
		t.Fatalf("expected failure for missing channel")
	}
	if !strings.Contains(result.ErrorMessage, "channel not found") {
		t.Fatalf("error message = %q, want channel not found", result.ErrorMessage)
	}
}

func TestTestModelConnectivityChannelDisabled(t *testing.T) {
	ctx := setupConnectivityTestDB(t)

	channel := &model.Channel{
		Name:     "disabled-channel",
		Enabled:  true,
		Type:     outbound.OutboundTypeOpenAIChat,
		BaseUrls: []model.BaseUrl{{URL: "http://localhost"}},
		Model:    "gpt-4o",
		Keys: []model.ChannelKey{{
			Enabled:    true,
			ChannelKey: "sk-test",
		}},
	}
	if err := op.ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	disabled := false
	if _, err := op.ChannelUpdate(&model.ChannelUpdateRequest{
		ID:      channel.ID,
		Enabled: &disabled,
	}, ctx); err != nil {
		t.Fatalf("ChannelUpdate() error = %v", err)
	}

	result, err := TestModelConnectivity(ctx, model.ModelTestRequest{ChannelID: channel.ID, Model: "gpt-4o"})
	if err != nil {
		t.Fatalf("TestModelConnectivity() error = %v", err)
	}
	if result.Success {
		t.Fatalf("expected failure for disabled channel")
	}
	if result.ErrorMessage != "channel is disabled" {
		t.Fatalf("error message = %q, want channel is disabled", result.ErrorMessage)
	}
}

func TestTestModelConnectivityUpstreamError(t *testing.T) {
	ctx := setupConnectivityTestDB(t)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"invalid key"}}`))
	}))
	defer upstream.Close()

	channel := &model.Channel{
		Name:     "test-upstream-error",
		Enabled:  true,
		Type:     outbound.OutboundTypeOpenAIChat,
		BaseUrls: []model.BaseUrl{{URL: upstream.URL}},
		Model:    "gpt-4o",
		Keys: []model.ChannelKey{{
			Enabled:    true,
			ChannelKey: "sk-test",
		}},
	}
	if err := op.ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	result, err := TestModelConnectivity(ctx, model.ModelTestRequest{ChannelID: channel.ID, Model: "gpt-4o"})
	if err != nil {
		t.Fatalf("TestModelConnectivity() error = %v", err)
	}
	if result.Success {
		t.Fatalf("expected failure for upstream error")
	}
	if !strings.Contains(result.ErrorMessage, "invalid key") {
		t.Fatalf("error message = %q, want invalid key", result.ErrorMessage)
	}
}

func TestPriceWithinTolerance(t *testing.T) {
	cases := []struct {
		a, b   float64
		expect bool
	}{
		{0, 0, true},
		{1.0, 1.0, true},
		{1.0, 1.04, true},
		{1.0, 1.06, false},
		{0.01, 0.0105, true},
		{0.01, 0.011, false},
	}
	for _, tc := range cases {
		got := priceWithinTolerance(tc.a, tc.b)
		if got != tc.expect {
			t.Errorf("priceWithinTolerance(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.expect)
		}
	}
}
