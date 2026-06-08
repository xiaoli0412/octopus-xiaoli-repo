package price

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func parseModelsDevPricePayload(payload []byte) (map[string]model.LLMPrice, error) {
	var rawPrice map[string]struct {
		Models map[string]struct {
			ID   string         `json:"id"`
			Cost model.LLMPrice `json:"cost"`
		} `json:"models"`
	}
	if err := json.Unmarshal(payload, &rawPrice); err != nil {
		return nil, fmt.Errorf("parse models.dev payload: %w", err)
	}
	out := make(map[string]model.LLMPrice, 1024)
	for _, provider := range Provider {
		for _, entry := range rawPrice[provider].Models {
			if strings.TrimSpace(entry.ID) == "" {
				continue
			}
			out[strings.ToLower(strings.TrimSpace(entry.ID))] = entry.Cost
		}
	}
	return out, nil
}

type liteLLMPriceEntry struct {
	InputCostPerToken           float64 `json:"input_cost_per_token"`
	OutputCostPerToken          float64 `json:"output_cost_per_token"`
	CacheReadInputTokenCost     float64 `json:"cache_read_input_token_cost"`
	CacheWriteInputTokenCost    float64 `json:"cache_write_input_token_cost"`
	CacheCreationInputTokenCost float64 `json:"cache_creation_input_token_cost"`
}

func parseLiteLLMPricePayload(payload []byte) (map[string]model.LLMPrice, error) {
	var raw map[string]liteLLMPriceEntry
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("parse LiteLLM payload: %w", err)
	}
	out := make(map[string]model.LLMPrice, len(raw))
	for name, entry := range raw {
		out[strings.ToLower(strings.TrimSpace(name))] = model.LLMPrice{
			Input:      entry.InputCostPerToken * 1_000_000,
			Output:     entry.OutputCostPerToken * 1_000_000,
			CacheRead:  entry.CacheReadInputTokenCost * 1_000_000,
			CacheWrite: maxFloat(entry.CacheWriteInputTokenCost, entry.CacheCreationInputTokenCost) * 1_000_000,
		}
	}
	return out, nil
}

type genAIPricesProvider struct {
	ID     string                  `json:"id"`
	Models []genAIPricesModelEntry `json:"models"`
}

type genAIPricesModelEntry struct {
	ID     string            `json:"id"`
	Match  genAIPricesMatch  `json:"match"`
	Prices genAIPriceOptions `json:"prices"`
}

type genAIPricesMatch struct {
	StartsWith string                 `json:"starts_with"`
	Equals     string                 `json:"equals"`
	Contains   string                 `json:"contains"`
	Or         []genAIPricesMatchRule `json:"or"`
}

type genAIPricesMatchRule struct {
	StartsWith string `json:"starts_with"`
	Equals     string `json:"equals"`
	Contains   string `json:"contains"`
}

type genAIPriceAmount float64

func (a *genAIPriceAmount) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*a = 0
		return nil
	}
	var direct float64
	if err := json.Unmarshal(data, &direct); err == nil {
		*a = genAIPriceAmount(direct)
		return nil
	}
	var tiered struct {
		Base  *float64 `json:"base"`
		Tiers []struct {
			Price float64 `json:"price"`
		} `json:"tiers"`
	}
	if err := json.Unmarshal(data, &tiered); err != nil {
		return err
	}
	if tiered.Base != nil {
		*a = genAIPriceAmount(*tiered.Base)
		return nil
	}
	if len(tiered.Tiers) > 0 {
		*a = genAIPriceAmount(tiered.Tiers[0].Price)
		return nil
	}
	*a = 0
	return nil
}

type genAIPriceValues struct {
	InputMtok      genAIPriceAmount `json:"input_mtok"`
	OutputMtok     genAIPriceAmount `json:"output_mtok"`
	CacheReadMtok  genAIPriceAmount `json:"cache_read_mtok"`
	CacheWriteMtok genAIPriceAmount `json:"cache_write_mtok"`
}

func (v genAIPriceValues) isEmpty() bool {
	return v.InputMtok == 0 && v.OutputMtok == 0 && v.CacheReadMtok == 0 && v.CacheWriteMtok == 0
}

func (v genAIPriceValues) llmPrice() model.LLMPrice {
	return model.LLMPrice{
		Input:      float64(v.InputMtok),
		Output:     float64(v.OutputMtok),
		CacheRead:  float64(v.CacheReadMtok),
		CacheWrite: float64(v.CacheWriteMtok),
	}
}

type genAIPriceOptions []genAIPriceValues

func (o *genAIPriceOptions) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*o = nil
		return nil
	}
	var single genAIPriceValues
	if err := json.Unmarshal(data, &single); err == nil {
		if single.isEmpty() {
			*o = nil
			return nil
		}
		*o = []genAIPriceValues{single}
		return nil
	}
	var entries []struct {
		Prices genAIPriceValues `json:"prices"`
	}
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}
	out := make([]genAIPriceValues, 0, len(entries))
	for _, entry := range entries {
		if !entry.Prices.isEmpty() {
			out = append(out, entry.Prices)
		}
	}
	*o = out
	return nil
}

func (o genAIPriceOptions) primary() (genAIPriceValues, bool) {
	for _, price := range o {
		if !price.isEmpty() {
			return price, true
		}
	}
	return genAIPriceValues{}, false
}

func parseGenAIPricesPayload(payload []byte) (map[string]model.LLMPrice, error) {
	var providers []genAIPricesProvider
	if err := json.Unmarshal(payload, &providers); err != nil {
		return nil, fmt.Errorf("parse genai-prices payload: %w", err)
	}
	out := make(map[string]model.LLMPrice, 1024)
	for _, provider := range providers {
		for _, entry := range provider.Models {
			priceValues, ok := entry.Prices.primary()
			if !ok {
				continue
			}
			price := priceValues.llmPrice()
			if entry.ID != "" {
				out[strings.ToLower(strings.TrimSpace(entry.ID))] = price
			}
			for _, alias := range collectGenAIPriceAliases(entry.Match) {
				out[strings.ToLower(strings.TrimSpace(alias))] = price
			}
		}
	}
	return out, nil
}

func collectGenAIPriceAliases(match genAIPricesMatch) []string {
	aliases := make([]string, 0, len(match.Or)+2)
	if match.StartsWith != "" {
		aliases = append(aliases, match.StartsWith)
	}
	if match.Equals != "" {
		aliases = append(aliases, match.Equals)
	}
	if match.Contains != "" {
		aliases = append(aliases, match.Contains)
	}
	for _, rule := range match.Or {
		if rule.StartsWith != "" {
			aliases = append(aliases, rule.StartsWith)
		}
		if rule.Equals != "" {
			aliases = append(aliases, rule.Equals)
		}
		if rule.Contains != "" {
			aliases = append(aliases, rule.Contains)
		}
	}
	return aliases
}

func parseLLMPricesPayload(payload []byte) (map[string]model.LLMPrice, error) {
	var raw struct {
		Prices []struct {
			ID          string   `json:"id"`
			Input       float64  `json:"input"`
			Output      float64  `json:"output"`
			InputCached *float64 `json:"input_cached"`
		} `json:"prices"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("parse llm-prices payload: %w", err)
	}
	out := make(map[string]model.LLMPrice, len(raw.Prices))
	for _, entry := range raw.Prices {
		cacheRead := 0.0
		if entry.InputCached != nil {
			cacheRead = *entry.InputCached
		}
		out[strings.ToLower(strings.TrimSpace(entry.ID))] = model.LLMPrice{
			Input:     entry.Input,
			Output:    entry.Output,
			CacheRead: cacheRead,
		}
	}
	return out, nil
}

func maxFloat(a, b float64) float64 {
	if b > a {
		return b
	}
	return a
}
