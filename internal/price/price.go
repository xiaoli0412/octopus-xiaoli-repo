package price

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/client"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/llmname"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

const llmPriceUrl = "https://models.dev/api.json"
const liteLLMPriceURL = "https://raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json"
const genAIPricesURL = "https://raw.githubusercontent.com/pydantic/genai-prices/main/prices/data_slim.json"
const llmPricesURL = "https://www.llm-prices.com/current-v1.json"

const maxLLMPriceResponseBytes int64 = 8 << 20

var Provider = []string{
	"openai",     // GPT 系列
	"anthropic",  // Claude 系列
	"google",     // Gemini 系列
	"deepseek",   // DeepSeek 系列
	"xai",        // Grok 系列
	"alibaba",    // Qwen 系列
	"zhipuai",    // GLM 系列
	"minimax",    // MiniMax 系列
	"moonshotai", // Kimi/Moonshot
	"v0",         // v0 系列
}

var lastUpdateTime time.Time
var llmPriceMetadata = map[string]PriceMetadata{}

type PriceMetadata struct {
	PrimarySource string
	Sources       []string
}

type PriceResolution struct {
	MatchedKey    string
	PrimarySource string
	Sources       []string
}

type pricingSource interface {
	Name() string
	Fetch(ctx context.Context, httpClient *http.Client) (map[string]model.LLMPrice, error)
}

type staticPricingSource struct {
	name  string
	url   string
	fetch func([]byte) (map[string]model.LLMPrice, error)
}

func (s staticPricingSource) Name() string {
	return s.name
}

func (s staticPricingSource) Fetch(ctx context.Context, httpClient *http.Client) (map[string]model.LLMPrice, error) {
	payload, err := fetchPricingPayload(ctx, httpClient, s.url)
	if err != nil {
		return nil, err
	}
	return s.fetch(payload)
}

var pricingSources = []pricingSource{
	staticPricingSource{name: "models.dev", url: llmPriceUrl, fetch: parseModelsDevPricePayload},
	staticPricingSource{name: "llm-prices", url: llmPricesURL, fetch: parseLLMPricesPayload},
	staticPricingSource{name: "litellm", url: liteLLMPriceURL, fetch: parseLiteLLMPricePayload},
	staticPricingSource{name: "genai-prices", url: genAIPricesURL, fetch: parseGenAIPricesPayload},
}

func UpdateLLMPrice(ctx context.Context) error {
	log.Debugf("update LLM price task started")
	startTime := time.Now()
	defer func() {
		log.Debugf("update LLM price task finished, update time: %s", time.Since(startTime))
	}()
	client, err := client.GetHTTPClientSystemProxy(false)
	if err != nil {
		return err
	}
	mergedPrices := make(map[string]model.LLMPrice, len(llmPrice))
	mergedMeta := make(map[string]PriceMetadata, len(llmPriceMetadata))
	failures := make([]string, 0, len(pricingSources))
	successes := 0
	for _, source := range pricingSources {
		sourcePrices, fetchErr := source.Fetch(ctx, client)
		if fetchErr != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", source.Name(), fetchErr))
			log.Warnf("price source %s sync failed: %v", source.Name(), fetchErr)
			continue
		}
		successes++
		for key, value := range sourcePrices {
			mergeSourcePrice(mergedPrices, mergedMeta, key, value, source.Name())
		}
		log.Debugf("price source %s synced %d model aliases", source.Name(), len(sourcePrices))
	}
	if successes == 0 {
		return fmt.Errorf("failed to sync all pricing sources: %s", strings.Join(failures, "; "))
	}
	llmPriceLock.Lock()
	llmPrice = mergedPrices
	llmPriceMetadata = mergedMeta
	llmPriceLock.Unlock()
	lastUpdateTime = time.Now()
	return nil
}

func fetchPricingPayload(ctx context.Context, httpClient *http.Client, targetURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxLLMPriceResponseBytes+1))
}

func decodeLLMPriceResponse(r io.Reader, target any) error {
	payload, err := io.ReadAll(io.LimitReader(r, maxLLMPriceResponseBytes+1))
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	if int64(len(payload)) > maxLLMPriceResponseBytes {
		return fmt.Errorf("response body too large")
	}
	if err := json.Unmarshal(payload, target); err != nil {
		return err
	}
	return nil
}

func GetLastUpdateTime() time.Time {
	return lastUpdateTime
}

func mergePriceValues(current model.LLMPrice, incoming model.LLMPrice) model.LLMPrice {
	if current.Input == 0 && incoming.Input != 0 {
		current.Input = incoming.Input
	}
	if current.Output == 0 && incoming.Output != 0 {
		current.Output = incoming.Output
	}
	if current.CacheRead == 0 && incoming.CacheRead != 0 {
		current.CacheRead = incoming.CacheRead
	}
	if current.CacheWrite == 0 && incoming.CacheWrite != 0 {
		current.CacheWrite = incoming.CacheWrite
	}
	return current
}

func mergeSourcePrice(prices map[string]model.LLMPrice, meta map[string]PriceMetadata, modelName string, price model.LLMPrice, source string) {
	for _, alias := range sourcePriceAliases(modelName) {
		normalized := llmname.NormalizeModelName(alias)
		if normalized == "" {
			continue
		}
		prices[normalized] = mergePriceValues(prices[normalized], price)
		currentMeta := meta[normalized]
		if currentMeta.PrimarySource == "" {
			currentMeta.PrimarySource = source
		}
		if !containsString(currentMeta.Sources, source) {
			currentMeta.Sources = append(currentMeta.Sources, source)
		}
		meta[normalized] = currentMeta
	}
}

func sourcePriceAliases(modelName string) []string {
	seen := make(map[string]struct{}, 12)
	out := make([]string, 0, 12)
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}

	add(modelName)
	add(llmname.NormalizeModelName(modelName))
	for _, key := range llmname.CandidateModelKeys(modelName) {
		add(key)
	}
	if idx := strings.Index(modelName, "/"); idx >= 0 && idx+1 < len(modelName) {
		add(modelName[idx+1:])
	}
	if idx := strings.LastIndex(modelName, "/"); idx >= 0 && idx+1 < len(modelName) {
		add(modelName[idx+1:])
	}
	return out
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func ResolveLLMPrice(modelName string) (model.LLMPrice, PriceResolution, bool) {
	for _, candidate := range llmname.CandidateModelKeys(modelName) {
		llmPriceLock.RLock()
		price, ok := llmPrice[candidate]
		meta := llmPriceMetadata[candidate]
		llmPriceLock.RUnlock()
		if !ok {
			continue
		}
		return price, PriceResolution{
			MatchedKey:    candidate,
			PrimarySource: meta.PrimarySource,
			Sources:       append([]string(nil), meta.Sources...),
		}, true
	}
	return model.LLMPrice{}, PriceResolution{}, false
}

func EnrichLLMInfo(info model.LLMInfo) model.LLMInfo {
	if info.CanonicalName == "" {
		info.CanonicalName = llmname.CanonicalModelName(info.Name)
	}
	parsed := llmname.ParseNormalizedModelInfo(info.Name)
	info.ParsedVendor = parsed.Vendor
	info.ParsedVersion = parsed.Version
	info.ParsedSuffix = parsed.Suffix
	resolvedPrice, resolution, ok := ResolveLLMPrice(info.Name)
	if !ok {
		supported, policy, reason := model.InferCacheSupport(info)
		info.CacheSupported = &supported
		if info.CachePolicy == "" {
			info.CachePolicy = policy
		}
		if info.CacheReason == "" {
			info.CacheReason = reason
		}
		return info
	}
	if info.LLMPrice.IsZero() {
		info.LLMPrice = resolvedPrice
	}
	if info.OfficialLLMPrice.IsZero() {
		info.OfficialLLMPrice = model.OfficialPriceFromLLMPrice(resolvedPrice)
	}
	info.PriceSource = resolution.PrimarySource
	info.PriceMatchedKey = resolution.MatchedKey
	supported, policy, reason := model.InferCacheSupport(info)
	info.CacheSupported = &supported
	if info.CachePolicy == "" {
		info.CachePolicy = policy
	}
	if info.CacheReason == "" {
		info.CacheReason = reason
	}
	return info
}

func GetLLMPrice(modelName string) *model.LLMPrice {
	for _, candidate := range llmname.CandidateModelKeys(modelName) {
		info, err := op.LLMGet(candidate)
		if err == nil {
			if info.CanonicalName != "" && info.CanonicalName != candidate {
				if canonicalInfo, canonicalErr := op.LLMGetByCanonical(info.CanonicalName); canonicalErr == nil {
					price := canonicalInfo.LLMPrice
					return &price
				}
			}
			price := info.LLMPrice
			return &price
		}
		if canonicalInfo, canonicalErr := op.LLMGetByCanonical(candidate); canonicalErr == nil {
			price := canonicalInfo.LLMPrice
			return &price
		}
	}
	if price, _, ok := ResolveLLMPrice(modelName); ok {
		return &price
	}
	return nil
}
