package op

import (
	"encoding/json"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

const aiAutomationRedactedSecret = "[redacted]"

func redactAIAutomationConfigValues(values model.AIAutomationConfigValues) model.AIAutomationConfigValues {
	if strings.TrimSpace(values.APIKey) != "" {
		values.APIKey = aiAutomationRedactedSecret
	}
	return values
}

func redactAIAutomationConfig(config model.AIAutomationConfig) model.AIAutomationConfig {
	if strings.TrimSpace(config.APIKey) != "" {
		config.APIKey = aiAutomationRedactedSecret
	}
	config.ManualConfig = redactAIAutomationConfigValues(config.ManualConfig)
	config.EffectiveConfig = redactAIAutomationConfigValues(config.EffectiveConfig)
	return config
}

func RedactAIAutomationConfigForResponse(config model.AIAutomationConfig) model.AIAutomationConfig {
	return redactAIAutomationConfig(config)
}

func redactAIAutomationTask(task model.AITask) model.AITask {
	task.ConfigSnapshotJSON = redactAIAutomationConfigSnapshotJSON(task.ConfigSnapshotJSON)
	return task
}

func RedactAITaskForResponse(task model.AITask) model.AITask {
	return redactAIAutomationTask(task)
}

func redactAIAutomationArtifacts(artifacts model.AITaskArtifacts) model.AITaskArtifacts {
	artifacts.ConfigSnapshotJSON = redactAIAutomationConfigSnapshotJSON(artifacts.ConfigSnapshotJSON)
	if artifacts.ConfigSnapshot != nil {
		redacted := redactAIAutomationTaskConfigSnapshot(*artifacts.ConfigSnapshot)
		artifacts.ConfigSnapshot = &redacted
	}
	return artifacts
}

func RedactAITaskArtifactsForResponse(artifacts model.AITaskArtifacts) model.AITaskArtifacts {
	return redactAIAutomationArtifacts(artifacts)
}

func redactAIProfile(profile model.AIProfile) model.AIProfile {
	if len(profile.Versions) > 0 {
		versions := make([]model.AIProfileVersion, len(profile.Versions))
		copy(versions, profile.Versions)
		profile.Versions = versions
	}
	for i := range profile.Versions {
		profile.Versions[i].ContentJSON = redactAIAutomationSecretJSON(profile.Versions[i].ContentJSON)
	}
	if profile.DomainPayload != nil {
		profile.DomainPayload = redactAIAutomationSecretsValue(profile.DomainPayload)
	}
	return profile
}

func RedactAIProfileForResponse(profile model.AIProfile) model.AIProfile {
	return redactAIProfile(profile)
}

func redactAIAutomationConfigSnapshotJSON(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	var snapshot model.AIAutomationTaskConfig
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return redactAIAutomationSecretJSON(raw)
	}
	redacted := redactAIAutomationTaskConfigSnapshot(snapshot)
	encoded, err := json.Marshal(redacted)
	if err != nil {
		return redactAIAutomationSecretJSON(raw)
	}
	return string(encoded)
}

func redactAIAutomationTaskConfigSnapshot(snapshot model.AIAutomationTaskConfig) model.AIAutomationTaskConfig {
	snapshot = normalizeAITaskConfigSnapshot(snapshot)
	if strings.TrimSpace(snapshot.APIKey) != "" {
		snapshot.APIKey = aiAutomationRedactedSecret
	}
	return snapshot
}

func redactAIAutomationSecretJSON(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	var payload any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return raw
	}
	redacted := redactAIAutomationSecretsValue(payload)
	encoded, err := json.Marshal(redacted)
	if err != nil {
		return raw
	}
	return string(encoded)
}

func redactAIAutomationSecretsValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		redacted := make(map[string]any, len(typed))
		for key, nested := range typed {
			if isAIAutomationSecretKey(key) && strings.TrimSpace(stringFromAny(nested)) != "" {
				redacted[key] = aiAutomationRedactedSecret
				continue
			}
			redacted[key] = redactAIAutomationSecretsValue(nested)
		}
		return redacted
	case []any:
		redacted := make([]any, 0, len(typed))
		for _, item := range typed {
			redacted = append(redacted, redactAIAutomationSecretsValue(item))
		}
		return redacted
	default:
		return value
	}
}

func isAIAutomationSecretKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	return normalized == "api_key" || normalized == "apikey"
}
