package model

import "time"

const (
	UpstreamRefreshManual    = "manual"
	UpstreamRefreshScheduled = "scheduled"

	UpstreamCredentialManagementToken = "management_token"
	UpstreamCredentialAccessKey       = "access_key"

	UpstreamPriceSourceGateway = "gateway"
)

type UpstreamSite struct {
	ID                    int       `json:"id" gorm:"primaryKey"`
	Name                  string    `json:"name" gorm:"not null;index"`
	ProviderType          string    `json:"provider_type" gorm:"index"`
	BaseURL               string    `json:"base_url"`
	APIBaseURL            string    `json:"api_base_url"`
	AuthMode              string    `json:"auth_mode"`
	Enabled               bool      `json:"enabled" gorm:"default:true"`
	AutoRefresh           bool      `json:"auto_refresh" gorm:"default:false"`
	RefreshIntervalSecs   int       `json:"refresh_interval_secs" gorm:"default:43200"`
	AutoCheckin           bool      `json:"auto_checkin" gorm:"default:false"`
	CheckinIntervalSecs   int       `json:"checkin_interval_secs" gorm:"default:86400"`
	LastCheckinAt         time.Time `json:"last_checkin_at,omitempty"`
	CheckinLog            string    `json:"checkin_log,omitempty" gorm:"type:text"`
	SyncToChannel         bool      `json:"sync_to_channel" gorm:"default:false"`
	LinkedChannelID       int       `json:"linked_channel_id" gorm:"index"`
	LastRefreshAt         time.Time `json:"last_refresh_at,omitempty"`
	LastRefreshStatus     string    `json:"last_refresh_status,omitempty"`
	LastRefreshMessage    string    `json:"last_refresh_message,omitempty" gorm:"type:text"`
	ModelCount            int       `json:"model_count"`
	KeyCount              int       `json:"key_count"`
	GroupCount            int       `json:"group_count"`
	PriceCount            int       `json:"price_count"`
	SubscriptionCount     int       `json:"subscription_count"`
	BalanceAvailable      bool      `json:"balance_available"`
	BalanceUsed           float64   `json:"balance_used"`
	BalanceRemain         float64   `json:"balance_remain"`
	BalanceUnlimited      bool      `json:"balance_unlimited"`
	BalanceAlertThreshold float64   `json:"balance_alert_threshold" gorm:"default:0"`
	LastBalanceCheckAt    time.Time `json:"last_balance_check_at,omitempty"`
	LastBalanceValue      float64   `json:"last_balance_value"`
	AutoCreateKey         bool      `json:"auto_create_key" gorm:"default:false"`
	KeyQuotaLimit         float64   `json:"key_quota_limit" gorm:"default:0"`
	KeyExpireDays         int       `json:"key_expire_days" gorm:"default:0"`
	AutoSyncGroup         bool      `json:"auto_sync_group" gorm:"default:false"`
	AutoSyncPrice         bool      `json:"auto_sync_price" gorm:"default:false"`
	SourceLabel           string    `json:"source_label,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type UpstreamCheckinLogEntry struct {
	Success bool      `json:"success"`
	Amount  float64   `json:"amount,omitempty"`
	Message string    `json:"message,omitempty"`
	At      time.Time `json:"at"`
}

type UpstreamCheckinResult struct {
	Success bool      `json:"success"`
	Amount  float64   `json:"amount,omitempty"`
	Message string    `json:"message,omitempty"`
	At      time.Time `json:"at"`
}

type UpstreamCredential struct {
	ID              int       `json:"id" gorm:"primaryKey"`
	UpstreamSiteID  int       `json:"upstream_site_id" gorm:"index"`
	CredentialType  string    `json:"credential_type" gorm:"index"`
	AuthMode        string    `json:"auth_mode"`
	DisplayName     string    `json:"display_name"`
	MaskedValue     string    `json:"masked_value"`
	EncryptedValue  string    `json:"-" gorm:"type:text"`
	UserID          string    `json:"user_id,omitempty"`
	Importable      bool      `json:"importable"`
	LastValidatedAt time.Time `json:"last_validated_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type UpstreamKeySnapshot struct {
	ID                  int       `json:"id" gorm:"primaryKey"`
	UpstreamSiteID      int       `json:"upstream_site_id" gorm:"index"`
	Name                string    `json:"name,omitempty"`
	MaskedKey           string    `json:"masked_key,omitempty"`
	AllowedModels       string    `json:"allowed_models,omitempty" gorm:"type:text"`
	RequestCapabilities string    `json:"request_capabilities,omitempty"`
	Groups              string    `json:"groups,omitempty" gorm:"type:text"`
	Status              string    `json:"status,omitempty"`
	Quota               float64   `json:"quota,omitempty"`
	QuotaUsed           float64   `json:"quota_used,omitempty"`
	ExpiresAt           string    `json:"expires_at,omitempty"`
	Importable          bool      `json:"importable"`
	SourceType          string    `json:"source_type,omitempty"`
	ChannelKeyID        int       `json:"channel_key_id,omitempty" gorm:"index"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type UpstreamGroupSnapshot struct {
	ID                  int       `json:"id" gorm:"primaryKey"`
	UpstreamSiteID      int       `json:"upstream_site_id" gorm:"index"`
	ExternalID          string    `json:"external_id,omitempty"`
	Name                string    `json:"name"`
	Description         string    `json:"description,omitempty" gorm:"type:text"`
	Platform            string    `json:"platform,omitempty"`
	Status              string    `json:"status,omitempty"`
	RateMultiplier      float64   `json:"rate_multiplier,omitempty"`
	Models              string    `json:"models,omitempty" gorm:"type:text"`
	RequestCapabilities string    `json:"request_capabilities,omitempty"`
	Source              string    `json:"source,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type UpstreamModelPrice struct {
	ID                 int       `json:"id" gorm:"primaryKey"`
	UpstreamSiteID     int       `json:"upstream_site_id" gorm:"index"`
	ChannelID          int       `json:"channel_id,omitempty" gorm:"index"`
	ModelName          string    `json:"model_name" gorm:"index"`
	CanonicalName      string    `json:"canonical_name,omitempty" gorm:"index"`
	PriceSource        string    `json:"price_source,omitempty"`
	PriceMatchedKey    string    `json:"price_matched_key,omitempty"`
	SourceLabel        string    `json:"source_label,omitempty"`
	CachePolicy        string    `json:"cache_policy,omitempty"`
	CacheReason        string    `json:"cache_reason,omitempty" gorm:"type:text"`
	CacheSupported     *bool     `json:"cache_supported,omitempty" gorm:"-"`
	Input              float64   `json:"input"`
	Output             float64   `json:"output"`
	CacheRead          float64   `json:"cache_read"`
	CacheWrite         float64   `json:"cache_write"`
	OfficialInput      float64   `json:"official_input"`
	OfficialOutput     float64   `json:"official_output"`
	OfficialCacheRead  float64   `json:"official_cache_read"`
	OfficialCacheWrite float64   `json:"official_cache_write"`
	UpdatedAt          time.Time `json:"updated_at"`
	CreatedAt          time.Time `json:"created_at"`
}

type UpstreamSiteCreateRequest struct {
	Name                string  `json:"name"`
	ProviderType        string  `json:"provider_type"`
	BaseURL             string  `json:"base_url"`
	AuthMode            string  `json:"auth_mode"`
	Token               string  `json:"token,omitempty"`
	AccessKey           string  `json:"access_key,omitempty"`
	UserID              string  `json:"user_id,omitempty"`
	Username            string  `json:"username,omitempty"`
	Password            string  `json:"password,omitempty"`
	AutoRefresh         bool    `json:"auto_refresh"`
	RefreshIntervalSecs int     `json:"refresh_interval_secs"`
	AutoCheckin         bool    `json:"auto_checkin"`
	CheckinIntervalSecs int     `json:"checkin_interval_secs"`
	SyncToChannel       bool    `json:"sync_to_channel"`
	AutoSyncGroup       bool    `json:"auto_sync_group"`
	AutoSyncPrice       bool    `json:"auto_sync_price"`
	AutoCreateKey       bool    `json:"auto_create_key"`
	KeyQuotaLimit       float64 `json:"key_quota_limit"`
	KeyExpireDays       int     `json:"key_expire_days"`
	ChannelName         string  `json:"channel_name,omitempty"`
	TargetChannelID     int     `json:"target_channel_id,omitempty"`
}

type UpstreamSiteUpdateRequest struct {
	ID                    int      `json:"id" binding:"required"`
	Name                  *string  `json:"name,omitempty"`
	Enabled               *bool    `json:"enabled,omitempty"`
	AutoRefresh           *bool    `json:"auto_refresh,omitempty"`
	RefreshIntervalSecs   *int     `json:"refresh_interval_secs,omitempty"`
	AutoCheckin           *bool    `json:"auto_checkin,omitempty"`
	CheckinIntervalSecs   *int     `json:"checkin_interval_secs,omitempty"`
	SyncToChannel         *bool    `json:"sync_to_channel,omitempty"`
	AutoSyncGroup         *bool    `json:"auto_sync_group,omitempty"`
	AutoSyncPrice         *bool    `json:"auto_sync_price,omitempty"`
	LinkedChannelID       *int     `json:"linked_channel_id,omitempty"`
	BalanceAlertThreshold *float64 `json:"balance_alert_threshold,omitempty"`
	AutoCreateKey         *bool    `json:"auto_create_key,omitempty"`
	KeyQuotaLimit         *float64 `json:"key_quota_limit,omitempty"`
	KeyExpireDays         *int     `json:"key_expire_days,omitempty"`
}

type UpstreamBalanceCheckResult struct {
	Alert     bool    `json:"alert"`
	Threshold float64 `json:"threshold"`
	Remain    float64 `json:"remain"`
	Message   string  `json:"message,omitempty"`
}

type UpstreamSiteDetail struct {
	Site          UpstreamSite            `json:"site"`
	Credentials   []UpstreamCredential    `json:"credentials"`
	Keys          []UpstreamKeySnapshot   `json:"keys"`
	Groups        []UpstreamGroupSnapshot `json:"groups"`
	Prices        []UpstreamModelPrice    `json:"prices"`
	Subscriptions []UpstreamSubscription  `json:"subscriptions,omitempty"`
	LinkedChannel *UpstreamAppliedChannel `json:"linked_channel,omitempty"`
}

type UpstreamRefreshRequest struct {
	ID           int  `json:"id" binding:"required"`
	ApplyChannel bool `json:"apply_channel"`
	Manual       bool `json:"manual"`
}

type UpstreamCreateKeyRequest struct {
	SiteID    int      `json:"site_id" binding:"required"`
	Name      string   `json:"name" binding:"required"`
	Quota     float64  `json:"quota"`
	ExpiresAt string   `json:"expires_at,omitempty"`
	Models    []string `json:"models,omitempty"`
	Groups    []string `json:"groups,omitempty"`
}

type UpstreamCreateKeyResult struct {
	Name      string `json:"name"`
	Key       string `json:"key"`
	MaskedKey string `json:"masked_key"`
}

type UpstreamPriceSummary struct {
	ModelName        string               `json:"model_name"`
	OfficialPrice    OfficialLLMPrice     `json:"official_price"`
	GatewayPrices    []UpstreamModelPrice `json:"gateway_prices"`
	EffectiveGateway *UpstreamModelPrice  `json:"effective_gateway,omitempty"`
}

type UpstreamHealthStatus string

const (
	UpstreamHealthStatusHealthy   UpstreamHealthStatus = "healthy"
	UpstreamHealthStatusDegraded  UpstreamHealthStatus = "degraded"
	UpstreamHealthStatusUnhealthy UpstreamHealthStatus = "unhealthy"
)

type UpstreamHealthItem struct {
	ID                int                  `json:"id"`
	Name              string               `json:"name"`
	Status            UpstreamHealthStatus `json:"status"`
	Enabled           bool                 `json:"enabled"`
	LastRefreshAt     time.Time            `json:"last_refresh_at,omitempty"`
	LastRefreshStatus string               `json:"last_refresh_status,omitempty"`
	ModelCount        int                  `json:"model_count"`
	BalanceAvailable  bool                 `json:"balance_available"`
	BalanceRemain     float64              `json:"balance_remain"`
	BalanceUnlimited  bool                 `json:"balance_unlimited"`
	BalanceAlert      bool                 `json:"balance_alert"`
	ErrorRate         float64              `json:"error_rate"`
	Suppressed        bool                 `json:"suppressed"`
	Reasons           []string             `json:"reasons,omitempty"`
}

type UpstreamUsagePoint struct {
	Date         string  `json:"date"`
	Timestamp    int64   `json:"timestamp"`
	RequestCount int64   `json:"request_count"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	TotalTokens  int64   `json:"total_tokens"`
	Cost         float64 `json:"cost"`
	SuccessCount int64   `json:"success_count"`
	FailureCount int64   `json:"failure_count"`
}

type UpstreamUsageResponse struct {
	SiteID     int                  `json:"site_id"`
	Days       int                  `json:"days"`
	Points     []UpstreamUsagePoint `json:"points"`
	ChannelIDs []int                `json:"channel_ids,omitempty"`
}
