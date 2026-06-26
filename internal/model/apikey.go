package model

type APIKey struct {
	ID              int     `json:"id" gorm:"primaryKey"`
	Name            string  `json:"name" gorm:"not null"`
	APIKey          string  `json:"api_key" gorm:"not null"`
	Enabled         bool    `json:"enabled" gorm:"default:true"`
	ExpireAt        int64   `json:"expire_at,omitempty"`
	MaxCost         float64 `json:"max_cost,omitempty"`
	SupportedModels string  `json:"supported_models,omitempty"`

	// RateLimitRPM  每分钟请求数上限（0 为不限制）
	RateLimitRPM int64 `json:"rate_limit_rpm,omitempty"`
	// RateLimitTPM  每分钟 token 数上限（0 为不限制）
	RateLimitTPM int64 `json:"rate_limit_tpm,omitempty"`
	// RateLimitDaily 每日请求数上限（0 为不限制）
	RateLimitDaily int64 `json:"rate_limit_daily,omitempty"`
}

type APIKeyAuthStatus struct {
	OK              bool   `json:"ok"`
	APIKeyID        int    `json:"api_key_id"`
	Name            string `json:"name"`
	Enabled         bool   `json:"enabled"`
	ExpireAt        int64  `json:"expire_at,omitempty"`
	SupportedModels string `json:"supported_models,omitempty"`
	AuthMode        string `json:"auth_mode"`
}
