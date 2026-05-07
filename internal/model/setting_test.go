package model

import (
	"testing"
)

func TestSettingValidateRejectsNegativeNumericValues(t *testing.T) {
	cases := []struct {
		name string
		key  SettingKey
		want string
	}{
		{name: "stats save interval", key: SettingKeyStatsSaveInterval, want: "setting value must be a non-negative integer"},
		{name: "relay log keep period", key: SettingKeyRelayLogKeepPeriod, want: "setting value must be a non-negative integer"},
		{name: "race budget", key: SettingKeyRaceGlobalBudget, want: "setting value must be a non-negative integer"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := Setting{Key: tc.key, Value: "-1"}
			if err := s.Validate(); err == nil || err.Error() != tc.want {
				t.Fatalf("Validate() error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestSettingValidateAllowsZeroNumericValues(t *testing.T) {
	cases := []SettingKey{
		SettingKeyStatsSaveInterval,
		SettingKeyRelayLogKeepPeriod,
		SettingKeyRaceGlobalBudget,
	}

	for _, key := range cases {
		t.Run(string(key), func(t *testing.T) {
			s := Setting{Key: key, Value: "0"}
			if err := s.Validate(); err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
		})
	}
}

func TestSettingValidateAllowsSocksProxyURL(t *testing.T) {
	s := Setting{Key: SettingKeyProxyURL, Value: "socks://127.0.0.1:1080"}
	if err := s.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestSettingValidateRejectsUnsupportedProxyScheme(t *testing.T) {
	s := Setting{Key: SettingKeyProxyURL, Value: "ftp://127.0.0.1:21"}
	err := s.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want unsupported proxy scheme error")
	}
	if got, want := err.Error(), "proxy URL scheme must be http, https, socks, or socks5"; got != want {
		t.Fatalf("Validate() error = %q, want %q", got, want)
	}
}

func TestSettingValidateRejectsProxyURLWithCredentials(t *testing.T) {
	s := Setting{Key: SettingKeyProxyURL, Value: "http://user:pass@127.0.0.1:8080"}
	err := s.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want credential rejection")
	}
	if got, want := err.Error(), "proxy URL must not include credentials"; got != want {
		t.Fatalf("Validate() error = %q, want %q", got, want)
	}
}

func TestSettingValidateRejectsBaseURLWithCredentials(t *testing.T) {
	cases := []SettingKey{SettingKeyAPIBaseURL, SettingKeyAIAutomationBaseURL}
	for _, key := range cases {
		t.Run(string(key), func(t *testing.T) {
			s := Setting{Key: key, Value: "https://user:pass@example.com/v1"}
			err := s.Validate()
			if err == nil {
				t.Fatal("Validate() error = nil, want credential rejection")
			}
			if got := err.Error(); got != "api base URL must not include credentials" {
				t.Fatalf("Validate() error = %q, want %q", got, "api base URL must not include credentials")
			}
		})
	}
}

func TestSettingValidateAllowsDynamicRoutingModes(t *testing.T) {
	modes := []string{"shadow-ai", "hybrid", "metrics-only", "strict-mechanism", "incident-safe"}

	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			s := Setting{Key: SettingKeyDynamicRoutingMode, Value: mode}
			if err := s.Validate(); err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
		})
	}
}

func TestSettingValidateRejectsUnknownDynamicRoutingMode(t *testing.T) {
	s := Setting{Key: SettingKeyDynamicRoutingMode, Value: "ai-super-mode"}
	err := s.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want invalid mode error")
	}
	if got, want := err.Error(), "setting value must be one of shadow-ai, hybrid, metrics-only, strict-mechanism, or incident-safe"; got != want {
		t.Fatalf("Validate() error = %q, want %q", got, want)
	}
}

func TestSettingValidateAllowsAIAutomationSettings(t *testing.T) {
	cases := []Setting{
		{Key: SettingKeyAIAutomationEnabled, Value: "true"},
		{Key: SettingKeyAIAutomationUseLocalDefault, Value: "false"},
		{Key: SettingKeyAIAutomationBaseURL, Value: "http://127.0.0.1:8080/v1"},
		{Key: SettingKeyAIAutomationChannelType, Value: "openai-compatible"},
		{Key: SettingKeyAIAutomationModel, Value: "gpt-4o-mini"},
		{Key: SettingKeyConfigSourceMode, Value: ConfigSourceModeAIProfile},
		{Key: SettingKeyActiveAIProfileID, Value: "12"},
		{Key: SettingKeyDynamicRoutingLearningEnabled, Value: "true"},
	}

	for _, setting := range cases {
		t.Run(string(setting.Key), func(t *testing.T) {
			if err := setting.Validate(); err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
		})
	}
}

func TestSettingValidateRejectsInvalidAIAutomationSettings(t *testing.T) {
	cases := []struct {
		setting Setting
		want    string
	}{
		{setting: Setting{Key: SettingKeyAIAutomationEnabled, Value: "yes"}, want: "setting value must be true or false"},
		{setting: Setting{Key: SettingKeyConfigSourceMode, Value: "overwrite"}, want: "setting value must be manual or ai_profile"},
		{setting: Setting{Key: SettingKeyActiveAIProfileID, Value: "-1"}, want: "setting value must be a non-negative integer"},
		{setting: Setting{Key: SettingKeyAIAutomationBaseURL, Value: "ftp://127.0.0.1"}, want: "api base URL must be absolute http or https URL"},
		{setting: Setting{Key: SettingKeyAIAutomationChannelType, Value: "unknown"}, want: "setting value must be openai-compatible, openai, anthropic, or gemini"},
	}

	for _, tc := range cases {
		t.Run(string(tc.setting.Key), func(t *testing.T) {
			err := tc.setting.Validate()
			if err == nil || err.Error() != tc.want {
				t.Fatalf("Validate() error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestSettingValidateRejectsDurationValuesThatOverflowRuntimeConversions(t *testing.T) {
	tests := []struct {
		name string
		key  SettingKey
		unit string
		bad  string
	}{
		{name: "stats save interval minutes overflow", key: SettingKeyStatsSaveInterval, unit: "minutes", bad: "153722868"},
		{name: "model info update hours overflow", key: SettingKeyModelInfoUpdateInterval, unit: "hours", bad: "2562048"},
		{name: "sync llm hours overflow", key: SettingKeySyncLLMInterval, unit: "hours", bad: "2562048"},
		{name: "relay log keep period days overflow", key: SettingKeyRelayLogKeepPeriod, unit: "days", bad: "106752"},
		{name: "circuit breaker cooldown seconds overflow", key: SettingKeyCircuitBreakerCooldown, unit: "seconds", bad: "9223372037"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setting := &Setting{Key: tc.key, Value: tc.bad}
			err := setting.Validate()
			if err == nil || err.Error() != "setting value exceeds supported duration range" {
				t.Fatalf("Validate() error = %v, want duration range error", err)
			}
		})
	}
}

func TestSettingValidateAcceptsLargestSupportedRuntimeDurations(t *testing.T) {
	tests := []struct {
		name string
		key  SettingKey
		max  string
	}{
		{name: "stats save interval max", key: SettingKeyStatsSaveInterval, max: "153722867"},
		{name: "model info update max", key: SettingKeyModelInfoUpdateInterval, max: "2562047"},
		{name: "sync llm max", key: SettingKeySyncLLMInterval, max: "2562047"},
		{name: "relay log keep period max", key: SettingKeyRelayLogKeepPeriod, max: "106751"},
		{name: "circuit breaker cooldown max", key: SettingKeyCircuitBreakerCooldown, max: "9223372036"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setting := &Setting{Key: tc.key, Value: tc.max}
			if err := setting.Validate(); err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
		})
	}
}
