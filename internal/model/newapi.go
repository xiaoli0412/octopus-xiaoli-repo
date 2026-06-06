package model

type NewAPIInspectRequest struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
}

type NewAPITokenUsage struct {
	Available     bool    `json:"available"`
	UsedQuota     float64 `json:"used_quota,omitempty"`
	RemainQuota   float64 `json:"remain_quota,omitempty"`
	Unlimited     bool    `json:"unlimited,omitempty"`
	RawStatusText string  `json:"raw_status_text,omitempty"`
}

type NewAPIInspectResult struct {
	BaseURL             string           `json:"base_url"`
	APIBaseURL          string           `json:"api_base_url"`
	ModelCount          int              `json:"model_count"`
	Models              []string         `json:"models"`
	RequestCapabilities []string         `json:"request_capabilities"`
	TokenUsage          NewAPITokenUsage `json:"token_usage"`
	Warnings            []string         `json:"warnings,omitempty"`
}

const (
	UpstreamProviderNewAPI           = "newapi"
	UpstreamProviderSub2API          = "sub2api"
	UpstreamProviderOpenAICompatible = "openai_compatible"

	UpstreamAuthModeToken           = "token"
	UpstreamAuthModeAccessKey       = "access_key"
	UpstreamAuthModeAccountPassword = "account_password"
)

type UpstreamInspectRequest struct {
	ProviderType string `json:"provider_type"`
	BaseURL      string `json:"base_url"`
	AuthMode     string `json:"auth_mode"`
	Token        string `json:"token,omitempty"`
	AccessKey    string `json:"access_key,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
}

type UpstreamGatewayKey struct {
	Name                string   `json:"name,omitempty"`
	Key                 string   `json:"-"`
	MaskedKey           string   `json:"masked_key,omitempty"`
	AllowedModels       []string `json:"allowed_models,omitempty"`
	RequestCapabilities []string `json:"request_capabilities,omitempty"`
	Groups              []string `json:"groups,omitempty"`
	Status              string   `json:"status,omitempty"`
	Quota               float64  `json:"quota,omitempty"`
	QuotaUsed           float64  `json:"quota_used,omitempty"`
	ExpiresAt           string   `json:"expires_at,omitempty"`
	Importable          bool     `json:"importable"`
	SourceType          string   `json:"source_type,omitempty"`
}

type UpstreamGroup struct {
	ID                  string   `json:"id,omitempty"`
	Name                string   `json:"name"`
	Description         string   `json:"description,omitempty"`
	Platform            string   `json:"platform,omitempty"`
	Status              string   `json:"status,omitempty"`
	RateMultiplier      float64  `json:"rate_multiplier,omitempty"`
	Models              []string `json:"models,omitempty"`
	RequestCapabilities []string `json:"request_capabilities,omitempty"`
	Source              string   `json:"source,omitempty"`
}

type UpstreamSubscription struct {
	Name      string  `json:"name,omitempty"`
	Plan      string  `json:"plan,omitempty"`
	Status    string  `json:"status,omitempty"`
	Balance   float64 `json:"balance,omitempty"`
	ExpiresAt string  `json:"expires_at,omitempty"`
	Source    string  `json:"source,omitempty"`
}

type UpstreamPriceCandidate struct {
	Name            string   `json:"name"`
	CanonicalName   string   `json:"canonical_name,omitempty"`
	CacheSupported  *bool    `json:"cache_supported,omitempty"`
	CachePolicy     string   `json:"cache_policy,omitempty"`
	CacheReason     string   `json:"cache_reason,omitempty"`
	PriceSource     string   `json:"price_source,omitempty"`
	PriceMatchedKey string   `json:"price_matched_key,omitempty"`
	Sources         []string `json:"sources,omitempty"`
	LLMPrice
	OfficialLLMPrice
}

type UpstreamInspectResult struct {
	ProviderType        string                   `json:"provider_type"`
	AuthMode            string                   `json:"auth_mode"`
	BaseURL             string                   `json:"base_url"`
	APIBaseURL          string                   `json:"api_base_url"`
	OfficialSource      bool                     `json:"official_source"`
	SourceLabel         string                   `json:"source_label,omitempty"`
	ModelCount          int                      `json:"model_count"`
	Models              []string                 `json:"models"`
	RequestCapabilities []string                 `json:"request_capabilities"`
	TokenUsage          NewAPITokenUsage         `json:"token_usage"`
	Keys                []UpstreamGatewayKey     `json:"keys,omitempty"`
	Groups              []UpstreamGroup          `json:"groups,omitempty"`
	Subscriptions       []UpstreamSubscription   `json:"subscriptions,omitempty"`
	PriceCandidates     []UpstreamPriceCandidate `json:"price_candidates,omitempty"`
	Warnings            []string                 `json:"warnings,omitempty"`
}

type UpstreamApplyRequest struct {
	Inspect         UpstreamInspectRequest `json:"inspect"`
	TargetChannelID int                    `json:"target_channel_id,omitempty"`
	ChannelName     string                 `json:"channel_name,omitempty"`
	AppendKeys      *bool                  `json:"append_keys,omitempty"`
	OverwriteModels *bool                  `json:"overwrite_models,omitempty"`
	EnableChannel   *bool                  `json:"enable_channel,omitempty"`
}

type UpstreamAppliedChannel struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Type        int       `json:"type"`
	Enabled     bool      `json:"enabled"`
	BaseUrls    []BaseUrl `json:"base_urls,omitempty"`
	CustomModel string    `json:"custom_model,omitempty"`
	KeyCount    int       `json:"key_count"`
}

type UpstreamApplyResult struct {
	Channel Channel                `json:"-"`
	Summary UpstreamAppliedChannel `json:"channel"`
	Inspect UpstreamInspectResult  `json:"inspect"`
	Created bool                   `json:"created"`
}
