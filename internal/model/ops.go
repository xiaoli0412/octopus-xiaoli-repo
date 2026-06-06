package model

const (
	OpsScopeOverall    = "overall"
	OpsScopeModel      = "model"
	OpsScopeChannel    = "channel"
	OpsScopeChannelKey = "channel_key"
	OpsScopeAPIKey     = "api_key"
	OpsScopeIP         = "ip"

	OpsEntityOverall = "all"
)

type OpsMetricBucket struct {
	Scope                string `json:"scope" gorm:"primaryKey;size:32;not null;index:idx_ops_bucket_scope_time,priority:1"`
	EntityKey            string `json:"entity_key" gorm:"primaryKey;size:191;not null;index:idx_ops_bucket_scope_time,priority:2"`
	BucketStart          int64  `json:"bucket_start" gorm:"primaryKey;not null;index:idx_ops_bucket_scope_time,priority:3"`
	EntityLabel          string `json:"entity_label"`
	SuccessCount         int64  `json:"success_count"`
	FailureCount         int64  `json:"failure_count"`
	SkippedCount         int64  `json:"skipped_count"`
	WaitTime             int64  `json:"wait_time"`
	InputToken           int64  `json:"input_token"`
	OutputToken          int64  `json:"output_token"`
	CacheReadToken       int64  `json:"cache_read_token"`
	CacheWriteToken      int64  `json:"cache_write_token"`
	CacheHitCount        int64  `json:"cache_hit_count"`
	CacheWriteCount      int64  `json:"cache_write_count"`
	CacheSuccessCount    int64  `json:"cache_success_count"`
	CacheEligibleCount   int64  `json:"cache_eligible_count"`
	CacheIneligibleCount int64  `json:"cache_ineligible_count"`
}

type OpsEntitySummary struct {
	Scope                string  `json:"scope"`
	EntityKey            string  `json:"entity_key"`
	EntityLabel          string  `json:"entity_label"`
	EntityDisplayLabel   string  `json:"entity_display_label,omitempty"`
	SuccessCount         int64   `json:"success_count"`
	FailureCount         int64   `json:"failure_count"`
	SkippedCount         int64   `json:"skipped_count"`
	WaitTime             int64   `json:"wait_time"`
	InputToken           int64   `json:"input_token"`
	OutputToken          int64   `json:"output_token"`
	CacheReadToken       int64   `json:"cache_read_token"`
	CacheWriteToken      int64   `json:"cache_write_token"`
	CacheHitCount        int64   `json:"cache_hit_count"`
	CacheWriteCount      int64   `json:"cache_write_count"`
	CacheCreateCount     int64   `json:"cache_create_count"`
	CacheSuccessCount    int64   `json:"cache_success_count"`
	CacheEligibleCount   int64   `json:"cache_eligible_count"`
	CacheIneligibleCount int64   `json:"cache_ineligible_count"`
	CacheSupported       bool    `json:"cache_supported"`
	SuccessRate          float64 `json:"success_rate"`
	CacheHitRate         float64 `json:"cache_hit_rate"`
	CacheCreateRate      float64 `json:"cache_create_rate"`
	CacheRate            float64 `json:"cache_rate"`
	AvgLatencyMS         float64 `json:"avg_latency_ms"`
}

type OpsSeriesPoint struct {
	BucketStart          int64   `json:"bucket_start"`
	Label                string  `json:"label"`
	SuccessCount         int64   `json:"success_count"`
	FailureCount         int64   `json:"failure_count"`
	SkippedCount         int64   `json:"skipped_count"`
	WaitTime             int64   `json:"wait_time"`
	InputToken           int64   `json:"input_token"`
	OutputToken          int64   `json:"output_token"`
	CacheReadToken       int64   `json:"cache_read_token"`
	CacheWriteToken      int64   `json:"cache_write_token"`
	CacheHitCount        int64   `json:"cache_hit_count"`
	CacheWriteCount      int64   `json:"cache_write_count"`
	CacheCreateCount     int64   `json:"cache_create_count"`
	CacheSuccessCount    int64   `json:"cache_success_count"`
	CacheEligibleCount   int64   `json:"cache_eligible_count"`
	CacheIneligibleCount int64   `json:"cache_ineligible_count"`
	CacheSupported       bool    `json:"cache_supported"`
	SuccessRate          float64 `json:"success_rate"`
	CacheHitRate         float64 `json:"cache_hit_rate"`
	CacheCreateRate      float64 `json:"cache_create_rate"`
	CacheRate            float64 `json:"cache_rate"`
	AvgLatencyMS         float64 `json:"avg_latency_ms"`
}

type OpsOverview struct {
	Window         string             `json:"window"`
	Total          OpsEntitySummary   `json:"total"`
	TopModels      []OpsEntitySummary `json:"top_models"`
	TopChannels    []OpsEntitySummary `json:"top_channels"`
	TopChannelKeys []OpsEntitySummary `json:"top_channel_keys"`
	TopAPIKeys     []OpsEntitySummary `json:"top_api_keys"`
	TopIPs         []OpsEntitySummary `json:"top_ips"`
}

type OpsRecentDetail struct {
	ID               int64  `json:"id"`
	Time             int64  `json:"time"`
	ClientIP         string `json:"client_ip"`
	ClientIPLabel    string `json:"client_ip_label"`
	RequestModelName string `json:"request_model_name"`
	ActualModelName  string `json:"actual_model_name"`
	APIKeyID         int    `json:"api_key_id"`
	ChannelID        int    `json:"channel_id"`
	ChannelName      string `json:"channel_name"`
	ChannelKeyID     int    `json:"channel_key_id"`
	InputTokens      int    `json:"input_tokens"`
	OutputTokens     int    `json:"output_tokens"`
	CacheReadTokens  int    `json:"cache_read_tokens"`
	CacheWriteTokens int    `json:"cache_write_tokens"`
	CacheSupported   bool   `json:"cache_supported"`
	UseTime          int    `json:"use_time"`
	Success          bool   `json:"success"`
	StatusCode       int    `json:"status_code"`
	Error            string `json:"error"`
	AttemptCount     int    `json:"attempt_count"`
}
