package model

import (
	"encoding/json"
	"strings"
)

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

	// AllowedChannels 允许调用的渠道 ID 列表（JSON 数组，如 "[1,2,3]"）。空字符串表示不限制。
	AllowedChannels string `json:"allowed_channels,omitempty" gorm:"type:text"`
	// AllowedGroups 允许调用的分组 ID 列表（JSON 数组，如 "[1,2,3]"）。空字符串表示不限制。
	AllowedGroups string `json:"allowed_groups,omitempty" gorm:"type:text"`
	// AllowedCapabilities 允许调用的能力列表（JSON 数组，如 ["chat","embedding"]）。空字符串表示不限制。
	// 能力取值：chat / embedding / response / message
	AllowedCapabilities string `json:"allowed_capabilities,omitempty" gorm:"type:text"`
	// AllowedIPCIDRs 允许调用方 IP 的 CIDR 白名单（JSON 数组，如 ["10.0.0.0/8"]）。空字符串表示不限制。
	AllowedIPCIDRs string `json:"allowed_ip_cidrs,omitempty" gorm:"type:text"`

	// ResponseCacheEnabled 控制该 API Key 是否启用响应缓存（仅非流式请求）。
	// 需要全局 response_cache_enabled 也开启时才生效。
	ResponseCacheEnabled bool `json:"response_cache_enabled,omitempty" gorm:"default:false"`
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

// ParseAPIKeyIntList 解析 API Key 的 JSON 整数数组字段（如 AllowedChannels/AllowedGroups）。
// 空字符串、空白、解析失败或结果为空均返回 nil，调用方据此判定“不限制”。
func ParseAPIKeyIntList(raw string) []int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var out []int
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ParseAPIKeyStringList 解析 API Key 的 JSON 字符串数组字段（如 AllowedCapabilities/AllowedIPCIDRs）。
// 空字符串、空白、解析失败或结果为空均返回 nil，调用方据此判定“不限制”。
func ParseAPIKeyStringList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// NormalizeAPIKeyIntList 把整数切片序列化为 API Key 字段使用的 JSON 字符串。
// 空切片返回空字符串，便于“空值表示不限制”的向后兼容契约。
func NormalizeAPIKeyIntList(ids []int) string {
	if len(ids) == 0 {
		return ""
	}
	data, err := json.Marshal(ids)
	if err != nil {
		return ""
	}
	return string(data)
}

// NormalizeAPIKeyStringList 把字符串切片序列化为 API Key 字段使用的 JSON 字符串。
// 空切片返回空字符串，便于“空值表示不限制”的向后兼容契约。
func NormalizeAPIKeyStringList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	data, err := json.Marshal(items)
	if err != nil {
		return ""
	}
	return string(data)
}

// APIKeyAllowedChannels 解析 AllowedChannels 字段为渠道 ID 切片。nil 表示不限制。
func (k APIKey) APIKeyAllowedChannels() []int {
	return ParseAPIKeyIntList(k.AllowedChannels)
}

// APIKeyAllowedGroups 解析 AllowedGroups 字段为分组 ID 切片。nil 表示不限制。
func (k APIKey) APIKeyAllowedGroups() []int {
	return ParseAPIKeyIntList(k.AllowedGroups)
}

// APIKeyAllowedCapabilities 解析 AllowedCapabilities 字段为能力切片。nil 表示不限制。
func (k APIKey) APIKeyAllowedCapabilities() []string {
	return ParseAPIKeyStringList(k.AllowedCapabilities)
}

// APIKeyAllowedIPCIDRs 解析 AllowedIPCIDRs 字段为 CIDR 切片。nil 表示不限制。
func (k APIKey) APIKeyAllowedIPCIDRs() []string {
	return ParseAPIKeyStringList(k.AllowedIPCIDRs)
}
