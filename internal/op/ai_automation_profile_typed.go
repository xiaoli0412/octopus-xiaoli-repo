package op

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm"
)

var errAIProfileTypedPayloadNotFound = errors.New("ai profile typed payload not found")

func isTypedAIProfileDomain(domain string) bool {
	switch model.NormalizeAIProfileDomain(domain) {
	case model.AIProfileDomainGrouping,
		model.AIProfileDomainChannelRecognition,
		model.AIProfileDomainPriceRecognition,
		model.AIProfileDomainModelClassification,
		model.AIProfileDomainConfigHealthCheck:
		return true
	default:
		return false
	}
}

func BackfillAIProfileTypedPayloads(ctx context.Context) error {
	profiles := []model.AIProfile{}
	if err := db.GetDB().WithContext(ctx).Preload("Versions", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("version desc")
	}).Find(&profiles).Error; err != nil {
		return err
	}
	for _, profile := range profiles {
		if !isTypedAIProfileDomain(profile.Domain) || aiProfileHasTypedPayload(profile, ctx) {
			continue
		}
		if len(profile.Versions) == 0 || strings.TrimSpace(profile.Versions[0].ContentJSON) == "" {
			_ = markAIProfileMigrationStatus(profile.ID, model.AIProfileMigrationStatusLegacyOnly, "profile has no legacy content", ctx)
			continue
		}
		if err := persistAIProfileTypedPayload(profile, profile.Versions[0].ContentJSON, ctx); err != nil {
			_ = markAIProfileMigrationStatus(profile.ID, model.AIProfileMigrationStatusLegacyOnly, err.Error(), ctx)
		}
	}
	return nil

}

func aiProfileHasTypedPayload(profile model.AIProfile, ctx context.Context) bool {
	_, _, _, err := loadAIProfileTypedPayload(profile, ctx)
	return err == nil
}

func enrichAIProfileDetail(profile model.AIProfile, ctx context.Context) model.AIProfile {
	payloadType, payload, migrationStatus, err := loadAIProfileTypedPayload(profile, ctx)
	if err == nil {
		profile.DomainPayloadType = payloadType
		profile.DomainPayload = payload
		if strings.TrimSpace(profile.MigrationStatus) == "" {
			profile.MigrationStatus = migrationStatus
		}
		return profile
	}
	if strings.TrimSpace(profile.MigrationStatus) == "" {
		profile.MigrationStatus = model.AIProfileMigrationStatusLegacyOnly
	}
	return profile
}

func persistAIProfileTypedPayload(profile model.AIProfile, contentJSON string, ctx context.Context) error {
	profile.Domain = model.NormalizeAIProfileDomain(profile.Domain)
	if !isTypedAIProfileDomain(profile.Domain) {
		return nil
	}
	if profile.ID <= 0 {
		return fmt.Errorf("profile id is required")
	}
	if aiProfileHasTypedPayload(profile, ctx) {
		return markAIProfileMigrationStatus(profile.ID, model.AIProfileMigrationStatusTypedBackfilled, "", ctx)
	}
	payload, err := buildAIProfileTypedPayload(profile.Domain, contentJSON, profile.Explanation)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	common := aiTypedProfileRecord{
		ProfileID:        profile.ID,
		Version:          profile.Version,
		TaskID:           profile.SourceTaskID,
		Status:           profile.Status,
		Confidence:       profile.Confidence,
		RiskLevel:        aiTypedPayloadRiskLevel(payload),
		Summary:          aiTypedPayloadSummary(payload, profile.Explanation),
		TypedPayloadJSON: string(raw),
		TypedPayloadHash: aiTypedPayloadHash(raw),
		CreatedAt:        time.Now(),
	}
	if common.Version <= 0 {
		common.Version = 1
	}
	if common.Status == "" {
		common.Status = model.AIProfileStatusReady
	}
	if err := createDomainTypedProfile(profile.Domain, common, ctx); err != nil {
		return err
	}
	return markAIProfileMigrationStatus(profile.ID, model.AIProfileMigrationStatusTypedBackfilled, "", ctx)
}

type aiTypedProfileRecord struct {
	ProfileID        int
	Version          int
	TaskID           *int
	Status           string
	Confidence       float64
	RiskLevel        string
	Summary          string
	TypedPayloadJSON string
	TypedPayloadHash string
	CreatedAt        time.Time
}

func createDomainTypedProfile(domain string, common aiTypedProfileRecord, ctx context.Context) error {
	switch model.NormalizeAIProfileDomain(domain) {
	case model.AIProfileDomainGrouping:
		return db.GetDB().WithContext(ctx).Create(&model.AIGroupingProfile{ProfileID: common.ProfileID, Version: common.Version, TaskID: common.TaskID, Status: common.Status, Confidence: common.Confidence, RiskLevel: common.RiskLevel, Summary: common.Summary, TypedPayloadJSON: common.TypedPayloadJSON, TypedPayloadHash: common.TypedPayloadHash, CreatedAt: common.CreatedAt}).Error
	case model.AIProfileDomainChannelRecognition:
		return db.GetDB().WithContext(ctx).Create(&model.AIChannelRecognitionProfile{ProfileID: common.ProfileID, Version: common.Version, TaskID: common.TaskID, Status: common.Status, Confidence: common.Confidence, RiskLevel: common.RiskLevel, Summary: common.Summary, TypedPayloadJSON: common.TypedPayloadJSON, TypedPayloadHash: common.TypedPayloadHash, CreatedAt: common.CreatedAt}).Error
	case model.AIProfileDomainPriceRecognition:
		return db.GetDB().WithContext(ctx).Create(&model.AIPriceRecognitionProfile{ProfileID: common.ProfileID, Version: common.Version, TaskID: common.TaskID, Status: common.Status, Confidence: common.Confidence, RiskLevel: common.RiskLevel, Summary: common.Summary, TypedPayloadJSON: common.TypedPayloadJSON, TypedPayloadHash: common.TypedPayloadHash, CreatedAt: common.CreatedAt}).Error
	case model.AIProfileDomainModelClassification:
		return db.GetDB().WithContext(ctx).Create(&model.AIModelClassificationProfile{ProfileID: common.ProfileID, Version: common.Version, TaskID: common.TaskID, Status: common.Status, Confidence: common.Confidence, RiskLevel: common.RiskLevel, Summary: common.Summary, TypedPayloadJSON: common.TypedPayloadJSON, TypedPayloadHash: common.TypedPayloadHash, CreatedAt: common.CreatedAt}).Error
	case model.AIProfileDomainConfigHealthCheck:
		return db.GetDB().WithContext(ctx).Create(&model.AIConfigHealthProfile{ProfileID: common.ProfileID, Version: common.Version, TaskID: common.TaskID, Status: common.Status, Confidence: common.Confidence, RiskLevel: common.RiskLevel, Summary: common.Summary, TypedPayloadJSON: common.TypedPayloadJSON, TypedPayloadHash: common.TypedPayloadHash, CreatedAt: common.CreatedAt}).Error
	default:
		return nil
	}
}

func loadAIProfileTypedPayload(profile model.AIProfile, ctx context.Context) (string, map[string]any, string, error) {
	var raw string
	payloadType := model.NormalizeAIProfileDomain(profile.Domain)
	switch payloadType {
	case model.AIProfileDomainGrouping:
		var row model.AIGroupingProfile
		if err := db.GetDB().WithContext(ctx).Where("profile_id = ?", profile.ID).Order("version desc, id desc").First(&row).Error; err != nil {
			return "", nil, "", normalizeTypedPayloadLoadError(err)
		}
		raw = row.TypedPayloadJSON
	case model.AIProfileDomainChannelRecognition:
		var row model.AIChannelRecognitionProfile
		if err := db.GetDB().WithContext(ctx).Where("profile_id = ?", profile.ID).Order("version desc, id desc").First(&row).Error; err != nil {
			return "", nil, "", normalizeTypedPayloadLoadError(err)
		}
		raw = row.TypedPayloadJSON
	case model.AIProfileDomainPriceRecognition:
		var row model.AIPriceRecognitionProfile
		if err := db.GetDB().WithContext(ctx).Where("profile_id = ?", profile.ID).Order("version desc, id desc").First(&row).Error; err != nil {
			return "", nil, "", normalizeTypedPayloadLoadError(err)
		}
		raw = row.TypedPayloadJSON
	case model.AIProfileDomainModelClassification:
		var row model.AIModelClassificationProfile
		if err := db.GetDB().WithContext(ctx).Where("profile_id = ?", profile.ID).Order("version desc, id desc").First(&row).Error; err != nil {
			return "", nil, "", normalizeTypedPayloadLoadError(err)
		}
		raw = row.TypedPayloadJSON
	case model.AIProfileDomainConfigHealthCheck:
		var row model.AIConfigHealthProfile
		if err := db.GetDB().WithContext(ctx).Where("profile_id = ?", profile.ID).Order("version desc, id desc").First(&row).Error; err != nil {
			return "", nil, "", normalizeTypedPayloadLoadError(err)
		}
		raw = row.TypedPayloadJSON
	default:
		return "", nil, "", errAIProfileTypedPayloadNotFound
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return "", nil, "", err
	}
	return payloadType, payload, model.AIProfileMigrationStatusTypedBackfilled, nil
}

func normalizeTypedPayloadLoadError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errAIProfileTypedPayloadNotFound
	}
	return err
}

func buildAIProfileTypedPayload(domain, contentJSON, fallbackSummary string) (map[string]any, error) {
	legacy := map[string]any{}
	if err := json.Unmarshal([]byte(contentJSON), &legacy); err != nil {
		return nil, err
	}
	if typed, ok := legacy["domain_payload"].(map[string]any); ok && len(typed) > 0 {
		payload := cloneAIProfilePayloadMap(typed)
		if summary := strings.TrimSpace(stringFromAny(payload["summary"])); summary == "" {
			payload["summary"] = strings.TrimSpace(fallbackSummary)
		}
		if _, ok := payload["findings"]; !ok {
			payload["findings"] = ensurePayloadArray(nil, legacy["raw_output"])
		}
		if _, ok := payload["recommendations"]; !ok {
			payload["recommendations"] = []any{}
		}
		if _, ok := payload["risks"]; !ok {
			payload["risks"] = []any{}
		}
		payload["legacy_snapshot"] = legacy
		return payload, nil
	}
	summary := strings.TrimSpace(stringFromAny(legacy["summary"]))
	if summary == "" {
		summary = strings.TrimSpace(fallbackSummary)
	}
	if summary == "" {
		summary = deriveAITaskResultSummary(stringFromAny(legacy["raw_output"]))
	}
	payload := map[string]any{
		"summary":         summary,
		"findings":        ensurePayloadArray(legacy["findings"], legacy["raw_output"]),
		"recommendations": ensurePayloadArray(legacy["recommendations"], nil),
		"risks":           ensurePayloadArray(legacy["risks"], nil),
		"legacy_snapshot": legacy,
	}
	if config, ok := legacy["config"].(map[string]any); ok {
		payload["config"] = config
	}
	addDomainTypedPayloadFields(model.NormalizeAIProfileDomain(domain), payload, legacy)
	return payload, nil
}

func cloneAIProfilePayloadMap(value map[string]any) map[string]any {
	out := make(map[string]any, len(value))
	for key, item := range value {
		out[key] = item
	}
	return out
}
func addDomainTypedPayloadFields(domain string, payload, legacy map[string]any) {
	switch domain {
	case model.AIProfileDomainGrouping:
		payload["grouping_suggestions"] = firstPayloadValue(legacy, "grouping_suggestions", "groups")
		payload["candidate_channel_model_mappings"] = firstPayloadValue(legacy, "candidate_channel_model_mappings", "channel_model_mappings")
		payload["conflicts"] = firstPayloadValue(legacy, "conflicts")
	case model.AIProfileDomainChannelRecognition:
		payload["channel_type"] = firstPayloadValue(legacy, "channel_type", "detected_channel_type")
		payload["source_type"] = firstPayloadValue(legacy, "source_type")
		payload["model_coverage"] = firstPayloadValue(legacy, "model_coverage", "models")
		payload["evidence"] = firstPayloadValue(legacy, "evidence")
	case model.AIProfileDomainPriceRecognition:
		payload["billing_mode"] = firstPayloadValue(legacy, "billing_mode")
		payload["price_items"] = firstPayloadValue(legacy, "price_items", "prices")
		payload["currency"] = firstPayloadValue(legacy, "currency")
		payload["unit"] = firstPayloadValue(legacy, "unit")
		payload["missing_items"] = firstPayloadValue(legacy, "missing_items")
	case model.AIProfileDomainModelClassification:
		payload["canonical_name"] = firstPayloadValue(legacy, "canonical_name", "model")
		payload["aliases"] = firstPayloadValue(legacy, "aliases")
		payload["classification"] = firstPayloadValue(legacy, "classification", "category")
		payload["source_type"] = firstPayloadValue(legacy, "source_type")
		payload["route_hints"] = firstPayloadValue(legacy, "route_hints")
	case model.AIProfileDomainConfigHealthCheck:
		payload["issues"] = firstPayloadValue(legacy, "issues", "findings")
		payload["severity"] = firstPayloadValue(legacy, "severity", "risk_level")
		payload["suggested_actions"] = firstPayloadValue(legacy, "suggested_actions", "recommendations")
		payload["blocking_activation"] = firstPayloadValue(legacy, "blocking_activation")
	}
}

func firstPayloadValue(payload map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := payload[key]; ok && value != nil {
			return value
		}
	}
	return []any{}
}

func ensurePayloadArray(value any, fallback any) []any {
	switch typed := value.(type) {
	case []any:
		return typed
	case []string:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, item)
		}
		return out
	case string:
		if strings.TrimSpace(typed) != "" {
			return []any{strings.TrimSpace(typed)}
		}
	}
	if text := strings.TrimSpace(stringFromAny(fallback)); text != "" {
		return []any{map[string]any{"type": "ai_output", "description": trimTextWithSuffix(text, aiTaskResultSummaryLimit, "...")}}
	}
	return []any{}
}

func aiTypedPayloadSummary(payload map[string]any, fallback string) string {
	if summary := strings.TrimSpace(stringFromAny(payload["summary"])); summary != "" {
		return trimTextWithSuffix(summary, aiTaskResultSummaryLimit, "...")
	}
	return trimTextWithSuffix(fallback, aiTaskResultSummaryLimit, "...")
}

func aiTypedPayloadRiskLevel(payload map[string]any) string {
	if level := strings.TrimSpace(stringFromAny(payload["risk_level"])); level != "" {
		return level
	}
	if level := strings.TrimSpace(stringFromAny(payload["severity"])); level != "" {
		return level
	}
	if risks, ok := payload["risks"].([]any); ok && len(risks) > 0 {
		return "medium"
	}
	return "low"
}

func aiTypedPayloadHash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func markAIProfileMigrationStatus(profileID int, status, message string, ctx context.Context) error {
	if profileID <= 0 {
		return nil
	}
	return db.GetDB().WithContext(ctx).Model(&model.AIProfile{}).Where("id = ?", profileID).Updates(map[string]any{
		"migration_status": status,
		"migration_error":  trimTextWithSuffix(message, aiTaskResultSummaryLimit, "..."),
	}).Error
}

func extractAIAutomationConfigFromTypedProfile(profile model.AIProfile, ctx context.Context) (model.AIAutomationTaskConfig, bool, error) {
	_, payload, _, err := loadAIProfileTypedPayload(profile, ctx)
	if err != nil {
		if errors.Is(err, errAIProfileTypedPayloadNotFound) {
			return model.AIAutomationTaskConfig{}, false, nil
		}
		return model.AIAutomationTaskConfig{}, false, err
	}
	parsed, ok := extractAIAutomationConfigFromPayload(payload)
	return parsed, ok, nil
}

func extractAIAutomationConfigFromPayload(payload map[string]any) (model.AIAutomationTaskConfig, bool) {
	sections := []map[string]any{}
	for _, key := range []string{"typed_config", "config", "runtime"} {
		if section, ok := payload[key].(map[string]any); ok {
			sections = append(sections, section)
		}
	}
	for _, section := range sections {
		parsed := parseAIAutomationConfigSection(section)
		if !isEmptyAITaskConfigSnapshot(parsed) {
			return parsed, true
		}
	}
	return model.AIAutomationTaskConfig{}, false
}

func parseAIAutomationConfigSection(section map[string]any) model.AIAutomationTaskConfig {
	parsed := model.AIAutomationTaskConfig{}
	if baseURL, ok := section["base_url"].(string); ok {
		parsed.BaseURL = strings.TrimSpace(baseURL)
	}
	if apiKey, ok := section["api_key"].(string); ok {
		parsed.APIKey = strings.TrimSpace(apiKey)
	}
	if channelType, ok := section["channel_type"].(string); ok {
		parsed.ChannelType = strings.TrimSpace(channelType)
	}
	if channelType, ok := section["ai_channel_type"].(string); ok && parsed.ChannelType == "" {
		parsed.ChannelType = strings.TrimSpace(channelType)
	}
	if modelName, ok := section["model"].(string); ok {
		parsed.Model = strings.TrimSpace(modelName)
	}
	if modelName, ok := section["ai_configured_model"].(string); ok && parsed.Model == "" {
		parsed.Model = strings.TrimSpace(modelName)
	}
	if useLocalDefault, ok := section["use_local_default"].(bool); ok {
		value := useLocalDefault
		parsed.UseLocalDefault = &value
	}
	if useLocalDefault, ok := section["ai_use_local_default"].(bool); ok && parsed.UseLocalDefault == nil {
		value := useLocalDefault
		parsed.UseLocalDefault = &value
	}
	return parsed
}

func stringFromAny(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprint(value)
}
