package price

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/llmname"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestDecodeLLMPriceResponseRejectsOversizedBody(t *testing.T) {
	payload := strings.Repeat("a", int(maxLLMPriceResponseBytes)+1)

	err := decodeLLMPriceResponse(strings.NewReader(payload), &map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "response body too large") {
		t.Fatalf("decodeLLMPriceResponse() error = %v, want size limit error", err)
	}
}

func TestDecodeLLMPriceResponseParsesValidPayload(t *testing.T) {
	type modelCostPayload struct {
		ID   string         `json:"id"`
		Cost model.LLMPrice `json:"cost"`
	}
	type providerPayload struct {
		Models map[string]modelCostPayload `json:"models"`
	}

	payload := map[string]providerPayload{}
	err := decodeLLMPriceResponse(strings.NewReader(`{"openai":{"models":{"gpt-4.1":{"id":"gpt-4.1","cost":{"input":1.23,"output":4.56}}}}}`), &payload)
	if err != nil {
		t.Fatalf("decodeLLMPriceResponse() error = %v", err)
	}
	if payload["openai"].Models["gpt-4.1"].Cost.Input != 1.23 {
		t.Fatalf("decoded input cost = %v, want %v", payload["openai"].Models["gpt-4.1"].Cost.Input, 1.23)
	}
	if payload["openai"].Models["gpt-4.1"].Cost.Output != 4.56 {
		t.Fatalf("decoded output cost = %v, want %v", payload["openai"].Models["gpt-4.1"].Cost.Output, 4.56)
	}
}

func TestParseLiteLLMPricePayloadConvertsPerTokenPricing(t *testing.T) {
	raw := []byte(`{
		"gpt-4o": {
			"input_cost_per_token": 0.0000025,
			"output_cost_per_token": 0.00001,
			"cache_read_input_token_cost": 0.00000125
		}
	}`)

	got, err := parseLiteLLMPricePayload(raw)
	if err != nil {
		t.Fatalf("parseLiteLLMPricePayload() error = %v", err)
	}
	price, ok := got["gpt-4o"]
	if !ok {
		t.Fatalf("parseLiteLLMPricePayload() missing gpt-4o in %#v", got)
	}
	if price.Input != 2.5 || price.Output != 10 || price.CacheRead != 1.25 {
		t.Fatalf("liteLLM price = %#v, want input=2.5 output=10 cacheRead=1.25", price)
	}
}

func TestParseGenAIPricesPayloadRegistersAliasMatches(t *testing.T) {
	raw := []byte(`[
		{
			"id":"anthropic",
			"name":"Anthropic",
			"models":[
				{
					"id":"claude-3-7-sonnet-latest",
					"match":{"or":[{"starts_with":"claude-3-7-sonnet"},{"starts_with":"claude-3.7-sonnet"}]},
					"prices":{"input_mtok":3,"cache_read_mtok":0.3,"cache_write_mtok":3.75,"output_mtok":15}
				}
			]
		}
	]`)

	got, err := parseGenAIPricesPayload(raw)
	if err != nil {
		t.Fatalf("parseGenAIPricesPayload() error = %v", err)
	}
	for _, key := range []string{"claude-3-7-sonnet-latest", "claude-3.7-sonnet"} {
		price, ok := got[key]
		if !ok {
			t.Fatalf("parseGenAIPricesPayload() missing %q in %#v", key, got)
		}
		if price.Input != 3 || price.Output != 15 || price.CacheRead != 0.3 || price.CacheWrite != 3.75 {
			t.Fatalf("parseGenAIPricesPayload() %q = %#v", key, price)
		}
	}
}

func TestParseGenAIPricesPayloadSupportsTieredPriceArrays(t *testing.T) {
	raw := []byte(`[
		{
			"id":"anthropic",
			"models":[
				{
					"id":"claude-opus-4-6",
					"match":{"or":[{"starts_with":"claude-opus-4-6"},{"contains":"claude-4.6-opus"}]},
					"prices":[
						{
							"prices":{
								"input_mtok":{"base":5,"tiers":[{"start":200000,"price":10}]},
								"cache_write_mtok":{"base":6.25,"tiers":[{"start":200000,"price":12.5}]},
								"cache_read_mtok":{"base":0.5,"tiers":[{"start":200000,"price":1}]},
								"output_mtok":{"base":25,"tiers":[{"start":200000,"price":37.5}]}
							}
						},
						{
							"constraint":{"start_date":"2026-03-13"},
							"prices":{"input_mtok":5,"cache_write_mtok":6.25,"cache_read_mtok":0.5,"output_mtok":25}
						}
					]
				}
			]
		}
	]`)

	got, err := parseGenAIPricesPayload(raw)
	if err != nil {
		t.Fatalf("parseGenAIPricesPayload() error = %v", err)
	}
	for _, key := range []string{"claude-opus-4-6", "claude-4.6-opus"} {
		price, ok := got[key]
		if !ok {
			t.Fatalf("parseGenAIPricesPayload() missing %q in %#v", key, got)
		}
		if price.Input != 5 || price.Output != 25 || price.CacheRead != 0.5 || price.CacheWrite != 6.25 {
			t.Fatalf("parseGenAIPricesPayload() %q = %#v, want tier base prices", key, price)
		}
	}
}

func TestEnrichLLMInfoUsesResolvedPriceAndParsedMetadata(t *testing.T) {
	llmPriceLock.Lock()
	originalPrices := llmPrice
	originalMeta := llmPriceMetadata
	llmPrice = map[string]model.LLMPrice{
		"glm-4.7-pro": {Input: 1.2, Output: 4.8, CacheRead: 0.3, CacheWrite: 0.9},
	}
	llmPriceMetadata = map[string]PriceMetadata{
		"glm-4.7-pro": {PrimarySource: "litellm", Sources: []string{"litellm"}},
	}
	llmPriceLock.Unlock()
	t.Cleanup(func() {
		llmPriceLock.Lock()
		llmPrice = originalPrices
		llmPriceMetadata = originalMeta
		llmPriceLock.Unlock()
	})

	info := EnrichLLMInfo(model.LLMInfo{Name: "GLM 4.7 Pro"})
	if info.Input != 1.2 || info.Output != 4.8 {
		t.Fatalf("EnrichLLMInfo() price = %#v, want input=1.2 output=4.8", info.LLMPrice)
	}
	if info.OfficialInput != 1.2 || info.OfficialOutput != 4.8 {
		t.Fatalf("EnrichLLMInfo() official price = %#v, want official input/output fallback", info.OfficialLLMPrice)
	}
	if info.ParsedVendor != "glm" || info.ParsedVersion != "4.7" || info.ParsedSuffix != "pro" {
		t.Fatalf("EnrichLLMInfo() parsed fields = vendor=%q version=%q suffix=%q", info.ParsedVendor, info.ParsedVersion, info.ParsedSuffix)
	}
	if info.PriceSource != "litellm" {
		t.Fatalf("EnrichLLMInfo() PriceSource = %q, want litellm", info.PriceSource)
	}
	if info.PriceMatchedKey != "glm-4.7-pro" {
		t.Fatalf("EnrichLLMInfo() PriceMatchedKey = %q, want glm-4.7-pro", info.PriceMatchedKey)
	}
	if info.CanonicalName != llmname.CanonicalModelName("GLM 4.7 Pro") {
		t.Fatalf("EnrichLLMInfo() CanonicalName = %q, want %q", info.CanonicalName, llmname.CanonicalModelName("GLM 4.7 Pro"))
	}
}

func TestParseLLMPricesPayloadParsesVendorArray(t *testing.T) {
	payload := struct {
		UpdatedAt string `json:"updated_at"`
		Prices    []struct {
			ID          string   `json:"id"`
			Vendor      string   `json:"vendor"`
			Input       float64  `json:"input"`
			Output      float64  `json:"output"`
			InputCached *float64 `json:"input_cached"`
		} `json:"prices"`
	}{}
	if err := json.Unmarshal([]byte(`{"updated_at":"2026-05-19","prices":[{"id":"deepseek-v3.1","vendor":"deepseek","input":0.27,"output":1.1,"input_cached":0.07}]}`), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if payload.Prices[0].ID != "deepseek-v3.1" || payload.Prices[0].Input != 0.27 {
		t.Fatalf("payload = %#v", payload)
	}
}
