package model

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/xurl"
)

const maxSettingDurationNanos int64 = 1<<63 - 1

type SettingKey string

const (
	SettingKeyAuthTokenSecret               SettingKey = "auth_token_secret"
	SettingKeyProxyURL                      SettingKey = "proxy_url"
	SettingKeyAPIBaseURL                    SettingKey = "api_base_url"                   // 绯荤粺 API 鍦板潃锛堢敤浜庢枃妗ｅ睍绀猴級
	SettingKeyStatsSaveInterval             SettingKey = "stats_save_interval"            // 灏嗙粺璁′俊鎭啓鍏ユ暟鎹簱鐨勫懆鏈?鍒嗛挓)
	SettingKeyModelInfoUpdateInterval       SettingKey = "model_info_update_interval"     // 妯″瀷淇℃伅鏇存柊闂撮殧(灏忔椂)
	SettingKeySyncLLMInterval               SettingKey = "sync_llm_interval"              // LLM 鍚屾闂撮殧(灏忔椂)
	SettingKeyRelayLogKeepPeriod            SettingKey = "relay_log_keep_period"          // 鏃ュ織淇濆瓨鏃堕棿鑼冨洿(澶?
	SettingKeyRelayLogKeepEnabled           SettingKey = "relay_log_keep_enabled"         // 鏄惁淇濈暀鍘嗗彶鏃ュ織
	SettingKeyCORSAllowOrigins              SettingKey = "cors_allow_origins"             // 璺ㄥ煙鐧藉悕鍗?閫楀彿鍒嗛殧, 濡?"example.com,example2.com"). 涓虹┖涓嶅厑璁歌法鍩? "*"鍏佽鎵€鏈?
	SettingKeyCircuitBreakerThreshold       SettingKey = "circuit_breaker_threshold"      // 鐔旀柇瑙﹀彂闃堝€硷紙杩炵画澶辫触娆℃暟锛?
	SettingKeyCircuitBreakerCooldown        SettingKey = "circuit_breaker_cooldown"       // 鐔旀柇鍩虹鍐峰嵈鏃堕棿锛堢锛?
	SettingKeyCircuitBreakerMaxCooldown     SettingKey = "circuit_breaker_max_cooldown"   // 鐔旀柇鏈€澶у喎鍗存椂闂达紙绉掞級锛屾寚鏁伴€€閬夸笂闄?
	SettingKeyDynamicRoutingMode            SettingKey = "dynamic_routing_mode"           // 鍔ㄦ€佽矾鐢辫繍琛屾ā寮?
	SettingKeyDynamicRoutingHealthEnabled   SettingKey = "dynamic_routing_health_enabled" // 鏄惁鍚敤鍔ㄦ€佸仴搴疯皟鑺?
	SettingKeyDynamicRoutingLearningEnabled SettingKey = "dynamic_routing_learning_enabled"
	SettingKeyRaceGlobalBudget              SettingKey = "race_global_budget"  // 骞跺彂绔為€熷叏灞€棰勭畻
	SettingKeyRaceGroupBudget               SettingKey = "race_group_budget"   // 骞跺彂绔為€熷垎缁勯绠?
	SettingKeyRaceChannelBudget             SettingKey = "race_channel_budget" // 骞跺彂绔為€熸笭閬撻绠?
	SettingKeyRaceKeyBudget                 SettingKey = "race_key_budget"     // 骞跺彂绔為€?key 棰勭畻
	SettingKeyRaceProbeBudget               SettingKey = "race_probe_budget"   // 骞跺彂绔為€?probe 棰勭畻
	SettingKeyAIAutomationEnabled           SettingKey = "ai_automation_enabled"
	SettingKeyAIAutomationBaseURL           SettingKey = "ai_automation_base_url"
	SettingKeyAIAutomationAPIKey            SettingKey = "ai_automation_api_key"
	SettingKeyAIAutomationChannelType       SettingKey = "ai_automation_channel_type"
	SettingKeyAIAutomationModel             SettingKey = "ai_automation_model"
	SettingKeyAIAutomationUseLocalDefault   SettingKey = "ai_automation_use_local_default"
	SettingKeyConfigSourceMode              SettingKey = "config_source_mode"
	SettingKeyActiveAIProfileID             SettingKey = "active_ai_profile_id"
	SettingKeyAIGovernanceManagedGroupName  SettingKey = "ai_governance_managed_group_name"
	SettingKeyActiveStrategyProfileID       SettingKey = "active_strategy_profile_id"
	SettingKeyAIRuntimeStrategy             SettingKey = "ai_runtime_strategy"
	SettingKeyAIRuntimeDispatchMode         SettingKey = "ai_runtime_dispatch_mode"
	SettingKeyAIRuntimeMaxParallelRuns      SettingKey = "ai_runtime_max_parallel_runs"
	SettingKeyAIRuntimeDoubleReviewEnabled  SettingKey = "ai_runtime_double_review_enabled"
	SettingKeyAIRuntimeFallbackDeterministic SettingKey = "ai_runtime_fallback_to_deterministic"
	SettingKeyForcePasswordChange           SettingKey = "force_password_change"
)

const (
	ConfigSourceModeManual    = "manual"
	ConfigSourceModeAIProfile = "ai_profile"
)

const (
	DefaultAIAutomationBaseURL     = "http://127.0.0.1:1088/v1"
	DefaultAIAutomationChannelType = "openai-compatible"
)

type Setting struct {
	Key   SettingKey `json:"key" gorm:"primaryKey"`
	Value string     `json:"value" gorm:"not null"`
}

func DefaultSettings() []Setting {
	return []Setting{
		{Key: SettingKeyProxyURL, Value: ""},
		{Key: SettingKeyAuthTokenSecret, Value: ""},
		{Key: SettingKeyAPIBaseURL, Value: "http://localhost:1088"}, // 榛樿绯荤粺 API 鍦板潃
		{Key: SettingKeyStatsSaveInterval, Value: "10"},             // 榛樿10鍒嗛挓淇濆瓨涓€娆＄粺璁′俊鎭?
		{Key: SettingKeyCORSAllowOrigins, Value: ""},                // CORS 榛樿涓嶅厑璁歌法鍩燂紝璁剧疆涓?"*" 鎵嶅厑璁告墍鏈夋潵婧?
		{Key: SettingKeyModelInfoUpdateInterval, Value: "24"},       // 榛樿24灏忔椂鏇存柊涓€娆℃ā鍨嬩俊鎭?
		{Key: SettingKeySyncLLMInterval, Value: "24"},               // 榛樿24灏忔椂鍚屾涓€娆LM
		{Key: SettingKeyRelayLogKeepPeriod, Value: "7"},             // 榛樿鏃ュ織淇濆瓨7澶?
		{Key: SettingKeyRelayLogKeepEnabled, Value: "true"},         // 榛樿淇濈暀鍘嗗彶鏃ュ織
		{Key: SettingKeyCircuitBreakerThreshold, Value: "5"},        // 榛樿杩炵画澶辫触5娆¤Е鍙戠啍鏂?
		{Key: SettingKeyCircuitBreakerCooldown, Value: "60"},        // 榛樿鍩虹鍐峰嵈60绉?
		{Key: SettingKeyCircuitBreakerMaxCooldown, Value: "600"},    // 榛樿鏈€澶у喎鍗?00绉掞紙10鍒嗛挓锛?
		{Key: SettingKeyDynamicRoutingMode, Value: "hybrid"},        // 榛樿浣跨敤娣峰悎鍔ㄦ€佽矾鐢辨ā寮?
		{Key: SettingKeyDynamicRoutingHealthEnabled, Value: "true"}, // 榛樿寮€鍚姩鎬佸仴搴疯皟鑺?
		{Key: SettingKeyDynamicRoutingLearningEnabled, Value: "false"},
		{Key: SettingKeyRaceGlobalBudget, Value: "64"}, // 榛樿鍏ㄥ眬骞跺彂绔為€熼绠?
		{Key: SettingKeyRaceGroupBudget, Value: "8"},   // 榛樿鍒嗙粍骞跺彂绔為€熼绠?
		{Key: SettingKeyRaceChannelBudget, Value: "4"}, // 榛樿娓犻亾骞跺彂绔為€熼绠?
		{Key: SettingKeyRaceKeyBudget, Value: "2"},     // 榛樿 key 骞跺彂绔為€熼绠?
		{Key: SettingKeyRaceProbeBudget, Value: "16"},  // 榛樿 probe 骞跺彂绔為€熼绠?
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
		SettingKeyAIRuntimeMaxParallelRuns:
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
		SettingKeyForcePasswordChange:
		if s.Value != "true" && s.Value != "false" {
			return fmt.Errorf("setting value must be true or false")
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
	}

	return nil
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
