package model

import "strings"

type LLMPrice struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cache_read"`
	CacheWrite float64 `json:"cache_write"`
}

type OfficialLLMPrice struct {
	OfficialInput      float64 `json:"official_input"`
	OfficialOutput     float64 `json:"official_output"`
	OfficialCacheRead  float64 `json:"official_cache_read"`
	OfficialCacheWrite float64 `json:"official_cache_write"`
}

func (p LLMPrice) IsZero() bool {
	return p.Input == 0 && p.Output == 0 && p.CacheRead == 0 && p.CacheWrite == 0
}

func (p OfficialLLMPrice) IsZero() bool {
	return p.OfficialInput == 0 && p.OfficialOutput == 0 && p.OfficialCacheRead == 0 && p.OfficialCacheWrite == 0
}

func OfficialPriceFromLLMPrice(price LLMPrice) OfficialLLMPrice {
	return OfficialLLMPrice{
		OfficialInput:      price.Input,
		OfficialOutput:     price.Output,
		OfficialCacheRead:  price.CacheRead,
		OfficialCacheWrite: price.CacheWrite,
	}
}

type BillingMode string

const (
	BillingModeUnknown    BillingMode = "unknown"
	BillingModePerRequest BillingMode = "per_request"
	BillingModePerToken   BillingMode = "per_token"
	BillingModePerQuota   BillingMode = "per_quota"
	BillingModeFlat       BillingMode = "flat"
	BillingModeFree       BillingMode = "free"
)

type ProbePolicy string

const (
	ProbePolicyPassiveOnly  ProbePolicy = "passive_only"
	ProbePolicySparseSingle ProbePolicy = "sparse_single"
	ProbePolicySequential   ProbePolicy = "sequential"
	ProbePolicyConcurrent   ProbePolicy = "concurrent"
)

type CachePolicy string

const (
	CachePolicyUnknown     CachePolicy = "unknown"
	CachePolicySupported   CachePolicy = "supported"
	CachePolicyUnsupported CachePolicy = "unsupported"
)

func NormalizeBillingMode(input BillingMode) BillingMode {
	v := BillingMode(strings.ToLower(strings.TrimSpace(string(input))))
	if v == "" {
		return BillingModeUnknown
	}
	return v
}

func IsValidBillingMode(input BillingMode) bool {
	switch NormalizeBillingMode(input) {
	case BillingModeUnknown, BillingModePerRequest, BillingModePerToken, BillingModePerQuota, BillingModeFlat, BillingModeFree:
		return true
	default:
		return false
	}
}

func NormalizeProbePolicy(input ProbePolicy) ProbePolicy {
	v := ProbePolicy(strings.ToLower(strings.TrimSpace(string(input))))
	if v == "" {
		return ProbePolicyPassiveOnly
	}
	return v
}

func IsValidProbePolicy(input ProbePolicy) bool {
	switch NormalizeProbePolicy(input) {
	case ProbePolicyPassiveOnly, ProbePolicySparseSingle, ProbePolicySequential, ProbePolicyConcurrent:
		return true
	default:
		return false
	}
}

func NormalizeCachePolicy(input CachePolicy) CachePolicy {
	v := CachePolicy(strings.ToLower(strings.TrimSpace(string(input))))
	if v == "" {
		return CachePolicyUnknown
	}
	return v
}

func IsValidCachePolicy(input CachePolicy) bool {
	switch NormalizeCachePolicy(input) {
	case CachePolicyUnknown, CachePolicySupported, CachePolicyUnsupported:
		return true
	default:
		return false
	}
}

func InferCacheSupport(info LLMInfo) (bool, CachePolicy, string) {
	policy := NormalizeCachePolicy(info.CachePolicy)
	switch policy {
	case CachePolicySupported:
		return true, policy, strings.TrimSpace(info.CacheReason)
	case CachePolicyUnsupported:
		reason := strings.TrimSpace(info.CacheReason)
		if reason == "" {
			reason = "model metadata marks prompt cache as unsupported"
		}
		return false, policy, reason
	}

	lower := strings.ToLower(strings.TrimSpace(info.Name + " " + info.CanonicalName))
	for _, marker := range []string{
		"embedding", "embed", "rerank", "moderation", "whisper", "tts", "speech", "audio", "image", "sora",
	} {
		if strings.Contains(lower, marker) {
			return false, CachePolicyUnsupported, "model type does not use prompt cache"
		}
	}
	if info.CacheRead > 0 || info.CacheWrite > 0 || info.OfficialCacheRead > 0 || info.OfficialCacheWrite > 0 {
		return true, CachePolicySupported, "cache pricing metadata is available"
	}
	for _, marker := range []string{
		"gpt-", "o1", "o3", "o4", "claude-", "gemini-", "deepseek", "qwen", "glm", "kimi", "moonshot",
	} {
		if strings.Contains(lower, marker) {
			return true, CachePolicyUnknown, "chat model may support prompt cache; no explicit policy is set"
		}
	}
	return false, CachePolicyUnknown, "cache support is unknown"
}

type LLMInfo struct {
	Name string `json:"name" gorm:"primaryKey;not null"`
	LLMPrice
	OfficialLLMPrice
	CanonicalName         string      `json:"canonical_name"`
	BillingMode           BillingMode `json:"billing_mode"`
	ProbePolicy           ProbePolicy `json:"probe_policy"`
	ProbeIntervalSeconds  int         `json:"probe_interval_seconds"`
	ProbeConcurrencyLimit int         `json:"probe_concurrency_limit"`
	CachePolicy           CachePolicy `json:"cache_policy"`
	CacheReason           string      `json:"cache_reason"`
	UpstreamProviderType  string      `json:"upstream_provider_type"`
	UpstreamSource        string      `json:"upstream_source"`
	CacheSupported        *bool       `json:"cache_supported,omitempty" gorm:"-"`
	ParsedVendor          string      `json:"parsed_vendor,omitempty" gorm:"-"`
	ParsedVersion         string      `json:"parsed_version,omitempty" gorm:"-"`
	ParsedSuffix          string      `json:"parsed_suffix,omitempty" gorm:"-"`
	PriceSource           string      `json:"price_source,omitempty" gorm:"-"`
	PriceMatchedKey       string      `json:"price_matched_key,omitempty" gorm:"-"`
}

type LLMChannel struct {
	Name                string   `json:"name"`
	Enabled             bool     `json:"enabled"`
	ChannelID           int      `json:"channel_id"`
	ChannelName         string   `json:"channel_name"`
	KeyCount            int      `json:"key_count,omitempty"`
	RequestCapabilities []string `json:"request_capabilities,omitempty"`
	InventorySource     string   `json:"inventory_source,omitempty"`
}

type ServiceableModelInventoryItem struct {
	Name                string   `json:"name"`
	Enabled             bool     `json:"enabled"`
	ChannelID           int      `json:"channel_id"`
	ChannelName         string   `json:"channel_name"`
	KeyCount            int      `json:"key_count"`
	RequestCapabilities []string `json:"request_capabilities,omitempty"`
	InventorySource     string   `json:"inventory_source"`
}

type SelectableGroupModelInventoryItem struct {
	Name                string   `json:"name"`
	ChannelCount        int      `json:"channel_count"`
	EnabledChannelCount int      `json:"enabled_channel_count"`
	KeyCount            int      `json:"key_count"`
	RequestCapabilities []string `json:"request_capabilities,omitempty"`
	InventorySource     string   `json:"inventory_source"`
}

type RoutableModelInventoryItem struct {
	Name                string   `json:"name"`
	GroupID             int      `json:"group_id"`
	GroupName           string   `json:"group_name"`
	ChannelCount        int      `json:"channel_count"`
	EnabledChannelCount int      `json:"enabled_channel_count"`
	KeyCount            int      `json:"key_count"`
	RequestCapabilities []string `json:"request_capabilities,omitempty"`
	InventorySource     string   `json:"inventory_source"`
}

type CapabilityInventory struct {
	ServiceableModels []ServiceableModelInventoryItem     `json:"serviceable_models"`
	SelectableModels  []SelectableGroupModelInventoryItem `json:"selectable_models"`
	RoutableModels    []RoutableModelInventoryItem        `json:"routable_models"`
}

type GeminiModel struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

type GeminiModelList struct {
	Models        []GeminiModel `json:"models"`
	NextPageToken string        `json:"nextPageToken"`
}

type OpenAIModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type OpenAIModelList struct {
	Object string        `json:"object"`
	Data   []OpenAIModel `json:"data"`
}
type AnthropicModel struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"created_at"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
}

type AnthropicModelList struct {
	Data    []AnthropicModel `json:"data"`
	FirstID string           `json:"first_id"`
	HasMore bool             `json:"has_more"`
	LastID  string           `json:"last_id"`
}
