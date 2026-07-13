package model

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/xurl"
)

const maxSettingDurationNanos int64 = 1<<63 - 1

type SettingKey string

const (
	SettingKeyAuthTokenSecret                SettingKey = "auth_token_secret"
	SettingKeyAuthTokenSecretSecondary       SettingKey = "auth_token_secret_secondary"
	SettingKeyProxyURL                       SettingKey = "proxy_url"
	SettingKeyAPIBaseURL                     SettingKey = "api_base_url" // 系统 API 地址（用于文档展示）
	SettingKeyAPIAlternateBaseURLs           SettingKey = "api_alternate_base_urls"
	SettingKeyTrustedProxyCIDRs              SettingKey = "trusted_proxy_cidrs"
	SettingKeyOpsIPDisplayMode               SettingKey = "ops_ip_display_mode"
	SettingKeyStatsSaveInterval              SettingKey = "stats_save_interval"            // 将统计信息写入数据库的周期(分钟)
	SettingKeyModelInfoUpdateInterval        SettingKey = "model_info_update_interval"     // 模型信息更新间隔(小时)
	SettingKeySyncLLMInterval                SettingKey = "sync_llm_interval"              // LLM 同步间隔(小时)
	SettingKeyRelayLogKeepPeriod             SettingKey = "relay_log_keep_period"          // 日志保存时间范围(天)
	SettingKeyRelayLogKeepEnabled            SettingKey = "relay_log_keep_enabled"         // 是否保留历史日志
	SettingKeyCORSAllowOrigins               SettingKey = "cors_allow_origins"             // 跨域白名单(逗号分隔, 如 "example.com,example2.com"). 为空不允许跨域, "*"允许所有
	SettingKeyCircuitBreakerThreshold        SettingKey = "circuit_breaker_threshold"      // 熔断触发阈值（连续失败次数）
	SettingKeyCircuitBreakerCooldown         SettingKey = "circuit_breaker_cooldown"       // 熔断基础冷却时间（秒）
	SettingKeyCircuitBreakerMaxCooldown      SettingKey = "circuit_breaker_max_cooldown"   // 熔断最大冷却时间（秒），指数退避上限
	SettingKeyDynamicRoutingMode             SettingKey = "dynamic_routing_mode"           // 动态路由运行模式
	SettingKeyDynamicRoutingHealthEnabled    SettingKey = "dynamic_routing_health_enabled" // 是否启用动态健康调节
	SettingKeyDynamicRoutingLearningEnabled  SettingKey = "dynamic_routing_learning_enabled"
	SettingKeyRaceGlobalBudget               SettingKey = "race_global_budget"  // 并发竞速全局预算
	SettingKeyRaceGroupBudget                SettingKey = "race_group_budget"   // 并发竞速分组预算
	SettingKeyRaceChannelBudget              SettingKey = "race_channel_budget" // 并发竞速渠道预算
	SettingKeyRaceKeyBudget                  SettingKey = "race_key_budget"     // 并发竞速 key 预算
	SettingKeyRaceProbeBudget                SettingKey = "race_probe_budget"   // 并发竞速 probe 预算
	SettingKeyAIAutomationEnabled            SettingKey = "ai_automation_enabled"
	SettingKeyAIAutomationBaseURL            SettingKey = "ai_automation_base_url"
	SettingKeyAIAutomationAPIKey             SettingKey = "ai_automation_api_key"
	SettingKeyAIAutomationChannelType        SettingKey = "ai_automation_channel_type"
	SettingKeyAIAutomationModel              SettingKey = "ai_automation_model"
	SettingKeyAIAutomationUseLocalDefault    SettingKey = "ai_automation_use_local_default"
	SettingKeyConfigSourceMode               SettingKey = "config_source_mode"
	SettingKeyActiveAIProfileID              SettingKey = "active_ai_profile_id"
	SettingKeyAIGovernanceManagedGroupName   SettingKey = "ai_governance_managed_group_name"
	SettingKeyActiveStrategyProfileID        SettingKey = "active_strategy_profile_id"
	SettingKeyAIRuntimeStrategy              SettingKey = "ai_runtime_strategy"
	SettingKeyAIRuntimeDispatchMode          SettingKey = "ai_runtime_dispatch_mode"
	SettingKeyAIRuntimeMaxParallelRuns       SettingKey = "ai_runtime_max_parallel_runs"
	SettingKeyAIRuntimeDoubleReviewEnabled   SettingKey = "ai_runtime_double_review_enabled"
	SettingKeyAIRuntimeFallbackDeterministic SettingKey = "ai_runtime_fallback_to_deterministic"
	SettingKeyForcePasswordChange            SettingKey = "force_password_change"
	SettingKeyHealthCheckInterval            SettingKey = "health_check_interval"  // 主动健康检查间隔(持续时间字符串, 如 "5m")
	SettingKeyHealthCheckTimeout             SettingKey = "health_check_timeout"   // 单次探测超时(持续时间字符串, 如 "10s")
	SettingKeyHealthCheckLLMProbe            SettingKey = "health_check_llm_probe" // 是否启用 LLM 1-token 探测
	SettingKeyMetricsAuthToken               SettingKey = "metrics_auth_token"      // /metrics 端点认证 token
	SettingKeyMetricsIPAllowlist             SettingKey = "metrics_ip_allowlist"    // /metrics 端点 IP 白名单(逗号分隔 CIDR 或 IP)
	SettingKeyResponseCacheEnabled           SettingKey = "response_cache_enabled"           // 响应缓存全局开关
	SettingKeyResponseCacheTTL               SettingKey = "response_cache_ttl"               // 响应缓存 TTL（秒）
	SettingKeyResponseCacheMaxEntries        SettingKey = "response_cache_max_entries"       // 响应缓存最大条目数
	SettingKeyCostAlertWebhookURL            SettingKey = "cost_alert_webhook_url"  // 成本告警 Webhook URL，空表示禁用
	SettingKeyCostAlertThresholds            SettingKey = "cost_alert_thresholds"   // 成本告警阈值（逗号分隔百分比，如 "0.5,0.8,1.0"）
	SettingKeyCostAlertFormat                SettingKey = "cost_alert_format"       // 成本告警消息格式：generic / slack / feishu / dingtalk
	SettingKeyBackupInterval                 SettingKey = "backup_interval"         // 定时备份间隔(持续时间字符串, 如 "24h")
	SettingKeyBackupKeepCount                SettingKey = "backup_keep_count"       // 定时备份保留数量
)

const (
	ConfigSourceModeManual    = "manual"
	ConfigSourceModeAIProfile = "ai_profile"
)

const (
	DefaultAIAutomationBaseURL     = "http://127.0.0.1:1088/v1"
	DefaultAIAutomationChannelType = "openai-compatible"
)

const (
	OpsIPDisplayModeMasked = "masked"
	OpsIPDisplayModeFull   = "full"
)

type Setting struct {
	Key   SettingKey `json:"key" gorm:"primaryKey"`
	Value string     `json:"value" gorm:"not null"`
}

type PublicAccessInfo struct {
	PrimaryBaseURL     string   `json:"primary_base_url"`
	AlternateBaseURLs  []string `json:"alternate_base_urls"`
	CurrentBaseURL     string   `json:"current_base_url"`
	TrustedProxyCIDRs  []string `json:"trusted_proxy_cidrs"`
	OpsIPDisplayMode   string   `json:"ops_ip_display_mode"`
	CurrentClientIP    string   `json:"current_client_ip"`
	CurrentClientLabel string   `json:"current_client_label"`
}

func DefaultSettings() []Setting {
	return []Setting{
		{Key: SettingKeyProxyURL, Value: ""},
		{Key: SettingKeyAuthTokenSecret, Value: ""},
		{Key: SettingKeyAPIBaseURL, Value: "http://localhost:1088"}, // 默认系统 API 地址
		{Key: SettingKeyAPIAlternateBaseURLs, Value: ""},
		{Key: SettingKeyTrustedProxyCIDRs, Value: ""},
		{Key: SettingKeyOpsIPDisplayMode, Value: OpsIPDisplayModeMasked},
		{Key: SettingKeyStatsSaveInterval, Value: "10"},             // 默认10分钟保存一次统计信息
		{Key: SettingKeyCORSAllowOrigins, Value: ""},                // CORS 默认不允许跨域，设置为 "*" 才允许所有来源
		{Key: SettingKeyModelInfoUpdateInterval, Value: "24"},       // 默认24小时更新一次模型信息
		{Key: SettingKeySyncLLMInterval, Value: "24"},               // 默认24小时同步一次LLM
		{Key: SettingKeyRelayLogKeepPeriod, Value: "7"},             // 默认日志保存7天
		{Key: SettingKeyRelayLogKeepEnabled, Value: "true"},         // 默认保留历史日志
		{Key: SettingKeyCircuitBreakerThreshold, Value: "5"},        // 默认连续失败5次触发熔断
		{Key: SettingKeyCircuitBreakerCooldown, Value: "60"},        // 默认基础冷却60秒
		{Key: SettingKeyCircuitBreakerMaxCooldown, Value: "600"},    // 默认最大冷却600秒（10分钟）
		{Key: SettingKeyDynamicRoutingMode, Value: "hybrid"},        // 默认使用混合动态路由模式
		{Key: SettingKeyDynamicRoutingHealthEnabled, Value: "true"}, // 默认开启动态健康调节
		{Key: SettingKeyDynamicRoutingLearningEnabled, Value: "false"},
		{Key: SettingKeyRaceGlobalBudget, Value: "64"}, // 默认全局并发竞速预算
		{Key: SettingKeyRaceGroupBudget, Value: "8"},   // 默认分组并发竞速预算
		{Key: SettingKeyRaceChannelBudget, Value: "4"}, // 默认渠道并发竞速预算
		{Key: SettingKeyRaceKeyBudget, Value: "2"},     // 默认 key 并发竞速预算
		{Key: SettingKeyRaceProbeBudget, Value: "16"},  // 默认 probe 并发竞速预算
		{Key: SettingKeyAIAutomationEnabled, Value: "false"},
		{Key: SettingKeyAIAutomationBaseURL, Value: DefaultAIAutomationBaseURL},
		{Key: SettingKeyAIAutomationAPIKey, Value: ""},
		{Key: SettingKeyAIAutomationChannelType, Value: DefaultAIAutomationChannelType},
		{Key: SettingKeyAIAutomationModel, Value: ""},
		{Key: SettingKeyAIAutomationUseLocalDefault, Value: "true"},
		{Key: SettingKeyConfigSourceMode, Value: ConfigSourceModeManual},
		{Key: SettingKeyActiveAIProfileID, Value: "0"},
		{Key: SettingKeyAIGovernanceManagedGroupName, Value: "AI Governance Managed"},
		{Key: SettingKeyActiveStrategyProfileID, Value: "0"},
		{Key: SettingKeyAIRuntimeStrategy, Value: "highest_success_rate"},
		{Key: SettingKeyAIRuntimeDispatchMode, Value: "single_best"},
		{Key: SettingKeyAIRuntimeMaxParallelRuns, Value: "2"},
		{Key: SettingKeyAIRuntimeDoubleReviewEnabled, Value: "false"},
		{Key: SettingKeyAIRuntimeFallbackDeterministic, Value: "true"},
		{Key: SettingKeyForcePasswordChange, Value: "false"},
		{Key: SettingKeyHealthCheckInterval, Value: "5m"},    // 默认5分钟执行一次主动健康检查
		{Key: SettingKeyHealthCheckTimeout, Value: "10s"},    // 默认单次探测超时10秒
		{Key: SettingKeyHealthCheckLLMProbe, Value: "false"},  // 默认不启用 LLM 探测
		{Key: SettingKeyMetricsAuthToken, Value: ""},          // /metrics 鉴权 token，默认空（无鉴权）
		{Key: SettingKeyMetricsIPAllowlist, Value: ""},        // /metrics IP 白名单，默认空（无限制）
		{Key: SettingKeyAuthTokenSecretSecondary, Value: ""},  // JWT secondary 密钥，默认空（仅 primary 生效）
		{Key: SettingKeyCostAlertWebhookURL, Value: ""},       // 成本告警 Webhook URL，默认空（禁用）
		{Key: SettingKeyCostAlertThresholds, Value: "0.5,0.8,1.0"}, // 默认阈值 50% / 80% / 100%
		{Key: SettingKeyCostAlertFormat, Value: "generic"},    // 默认消息格式 generic
		{Key: SettingKeyResponseCacheEnabled, Value: "false"}, // 默认关闭响应缓存
		{Key: SettingKeyResponseCacheTTL, Value: "300"},       // 默认缓存 300 秒
		{Key: SettingKeyResponseCacheMaxEntries, Value: "1000"}, // 默认最多 1000 条
		{Key: SettingKeyBackupInterval, Value: "24h"},  // 默认每 24 小时备份一次
		{Key: SettingKeyBackupKeepCount, Value: "7"},   // 默认保留 7 个备份
	}
}

func (s *Setting) Validate() error {
	switch s.Key {
	case SettingKeyStatsSaveInterval,
		SettingKeyModelInfoUpdateInterval,
		SettingKeySyncLLMInterval,
		SettingKeyRelayLogKeepPeriod,
		SettingKeyCircuitBreakerThreshold,
		SettingKeyCircuitBreakerCooldown,
		SettingKeyCircuitBreakerMaxCooldown,
		SettingKeyRaceGlobalBudget,
		SettingKeyRaceGroupBudget,
		SettingKeyRaceChannelBudget,
		SettingKeyRaceKeyBudget,
		SettingKeyRaceProbeBudget,
		SettingKeyActiveAIProfileID,
		SettingKeyActiveStrategyProfileID,
		SettingKeyAIRuntimeMaxParallelRuns,
		SettingKeyResponseCacheTTL,
		SettingKeyResponseCacheMaxEntries,
		SettingKeyBackupKeepCount:
		v, err := validateSettingNonNegativeInt(s.Value)
		if err != nil {
			return err
		}
		switch s.Key {
		case SettingKeyStatsSaveInterval:
			return validateSettingDurationUnits(v, time.Minute)
		case SettingKeyModelInfoUpdateInterval, SettingKeySyncLLMInterval:
			return validateSettingDurationUnits(v, time.Hour)
		case SettingKeyRelayLogKeepPeriod:
			return validateSettingDurationUnits(v, 24*time.Hour)
		case SettingKeyCircuitBreakerCooldown, SettingKeyCircuitBreakerMaxCooldown:
			return validateSettingDurationUnits(v, time.Second)
		default:
			return nil
		}
	case SettingKeyRelayLogKeepEnabled,
		SettingKeyDynamicRoutingHealthEnabled,
		SettingKeyDynamicRoutingLearningEnabled,
		SettingKeyAIAutomationEnabled,
		SettingKeyAIAutomationUseLocalDefault,
		SettingKeyAIRuntimeDoubleReviewEnabled,
		SettingKeyAIRuntimeFallbackDeterministic,
		SettingKeyForcePasswordChange,
		SettingKeyHealthCheckLLMProbe,
		SettingKeyResponseCacheEnabled:
		if s.Value != "true" && s.Value != "false" {
			return fmt.Errorf("setting value must be true or false")
		}
		return nil
	case SettingKeyHealthCheckInterval, SettingKeyHealthCheckTimeout, SettingKeyBackupInterval:
		d, err := time.ParseDuration(s.Value)
		if err != nil {
			return fmt.Errorf("setting value must be a valid duration (e.g. 5m, 10s): %w", err)
		}
		if d <= 0 {
			return fmt.Errorf("setting value must be a positive duration")
		}
		return nil
	case SettingKeyConfigSourceMode:
		switch s.Value {
		case ConfigSourceModeManual, ConfigSourceModeAIProfile:
			return nil
		default:
			return fmt.Errorf("setting value must be manual or ai_profile")
		}
	case SettingKeyDynamicRoutingMode:
		switch s.Value {
		case "shadow-ai", "hybrid", "metrics-only", "strict-mechanism", "incident-safe":
			return nil
		default:
			return fmt.Errorf("setting value must be one of shadow-ai, hybrid, metrics-only, strict-mechanism, or incident-safe")
		}
	case SettingKeyAIRuntimeStrategy:
		switch s.Value {
		case "highest_success_rate", "balanced_latency", "cost_first":
			return nil
		default:
			return fmt.Errorf("setting value must be one of highest_success_rate, balanced_latency, or cost_first")
		}
	case SettingKeyAIRuntimeDispatchMode:
		switch s.Value {
		case "single_best", "bounded_parallel", "round_robin_review":
			return nil
		default:
			return fmt.Errorf("setting value must be one of single_best, bounded_parallel, or round_robin_review")
		}
	case SettingKeyProxyURL:
		if s.Value == "" {
			return nil
		}
		if err := xurl.ValidateProxyURL(s.Value, "proxy URL"); err != nil {
			return err
		}
		return nil
	case SettingKeyAPIBaseURL, SettingKeyAIAutomationBaseURL:
		if s.Value == "" {
			return nil
		}
		if err := xurl.ValidateAbsoluteHTTPURL(s.Value, "api base URL"); err != nil {
			return err
		}
		parsedURL, err := url.Parse(s.Value)
		if err != nil {
			return fmt.Errorf("api base URL is invalid: %w", err)
		}
		if parsedURL.User != nil {
			return fmt.Errorf("api base URL must not include credentials")
		}
		return nil
	case SettingKeyAPIAlternateBaseURLs:
		for _, value := range splitSettingList(s.Value) {
			if err := xurl.ValidateAbsoluteHTTPURL(value, "alternate api base URL"); err != nil {
				return err
			}
			parsedURL, err := url.Parse(value)
			if err != nil {
				return fmt.Errorf("alternate api base URL is invalid: %w", err)
			}
			if parsedURL.User != nil {
				return fmt.Errorf("alternate api base URL must not include credentials")
			}
		}
		return nil
	case SettingKeyTrustedProxyCIDRs:
		for _, value := range splitSettingList(s.Value) {
			if strings.Contains(value, "/") {
				if _, _, err := net.ParseCIDR(value); err != nil {
					return fmt.Errorf("trusted proxy CIDR is invalid: %w", err)
				}
				continue
			}
			if ip := net.ParseIP(value); ip == nil {
				return fmt.Errorf("trusted proxy CIDR must be an IP or CIDR")
			}
		}
		return nil
	case SettingKeyOpsIPDisplayMode:
		switch strings.ToLower(strings.TrimSpace(s.Value)) {
		case OpsIPDisplayModeMasked, OpsIPDisplayModeFull:
			return nil
		default:
			return fmt.Errorf("setting value must be masked or full")
		}
	case SettingKeyAIAutomationChannelType:
		switch s.Value {
		case "", "openai-compatible", "openai", "anthropic", "gemini":
			return nil
		default:
			return fmt.Errorf("setting value must be openai-compatible, openai, anthropic, or gemini")
		}
	case SettingKeyAIAutomationModel, SettingKeyAIAutomationAPIKey:
		return nil
	case SettingKeyAIGovernanceManagedGroupName:
		if strings.TrimSpace(s.Value) == "" {
			return fmt.Errorf("managed group name cannot be empty")
		}
		return nil
	case SettingKeyAuthTokenSecret:
		if s.Value == "" {
			return fmt.Errorf("auth token secret cannot be empty")
		}
		return nil
	case SettingKeyAuthTokenSecretSecondary:
		// secondary 可为空（向后兼容），非空时无额外约束
		return nil
	case SettingKeyMetricsAuthToken:
		// metrics token 可为空（向后兼容），非空时无额外约束
		return nil
	case SettingKeyMetricsIPAllowlist:
		for _, value := range splitSettingList(s.Value) {
			if strings.Contains(value, "/") {
				if _, _, err := net.ParseCIDR(value); err != nil {
					return fmt.Errorf("metrics IP allowlist entry is invalid: %w", err)
				}
				continue
			}
			if ip := net.ParseIP(value); ip == nil {
				return fmt.Errorf("metrics IP allowlist entry must be an IP or CIDR")
			}
		}
		return nil
	case SettingKeyCostAlertWebhookURL:
		// 空值表示禁用，非空时校验为合法 HTTP(S) URL
		if s.Value == "" {
			return nil
		}
		if err := xurl.ValidateAbsoluteHTTPURL(s.Value, "cost alert webhook URL"); err != nil {
			return err
		}
		return nil
	case SettingKeyCostAlertThresholds:
		// 空值表示使用默认阈值，非空时校验每个值为 (0, 2] 范围内的浮点数
		if strings.TrimSpace(s.Value) == "" {
			return nil
		}
		for _, part := range splitSettingList(s.Value) {
			v, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return fmt.Errorf("cost alert threshold must be a float: %q", part)
			}
			if v <= 0 || v > 2 {
				return fmt.Errorf("cost alert threshold must be in (0, 2]: %q", part)
			}
		}
		return nil
	case SettingKeyCostAlertFormat:
		switch strings.ToLower(strings.TrimSpace(s.Value)) {
		case "", "generic", "slack", "feishu", "dingtalk":
			return nil
		default:
			return fmt.Errorf("cost alert format must be one of generic, slack, feishu, dingtalk")
		}
	}

	return nil
}

func splitSettingList(value string) []string {
	normalized := strings.ReplaceAll(value, "\n", ",")
	normalized = strings.ReplaceAll(normalized, ";", ",")
	parts := strings.Split(normalized, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func validateSettingNonNegativeInt(value string) (int, error) {
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("setting value must be an integer")
	}
	if v < 0 {
		return 0, fmt.Errorf("setting value must be a non-negative integer")
	}
	return v, nil
}

func validateSettingDurationUnits(value int, unit time.Duration) error {
	if unit <= 0 {
		return nil
	}
	maxUnits := maxSettingDurationNanos / int64(unit)
	if int64(value) > maxUnits {
		return fmt.Errorf("setting value exceeds supported duration range")
	}
	return nil
}
