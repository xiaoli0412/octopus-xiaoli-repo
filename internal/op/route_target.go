package op

import (
	"context"
	"fmt"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/cache"
	"gorm.io/gorm/clause"
)

var routeTargetOverrideCache = cache.New[int, model.RouteTargetOverride](16)
var routeTargetOverrideIndex = cache.New[string, model.RouteTargetOverride](16)

func routeTargetOverrideLookupKey(channelID, channelKeyID int, modelName string) string {
	return fmt.Sprintf("%d|%d|%s", channelID, channelKeyID, model.NormalizeRouteTargetModelName(modelName))
}

func defaultRouteTargetResolvedPolicy(channelID, channelKeyID int, sourceType, modelName string) model.RouteTargetResolvedPolicy {
	basis := model.RouteTargetPolicyBasisChannelKeyInheritance
	return model.RouteTargetResolvedPolicy{
		ChannelID:             channelID,
		ChannelKeyID:          channelKeyID,
		ModelName:             model.NormalizeRouteTargetModelName(modelName),
		SourceType:            model.EffectiveChannelKeySourceType(sourceType),
		SourceTypeBasis:       basis,
		BillingMode:           model.BillingModeUnknown,
		BillingModeBasis:      basis,
		ProbePolicy:           model.ProbePolicyPassiveOnly,
		ProbePolicyBasis:      basis,
		ProbeIntervalSeconds:  3600,
		ProbeIntervalBasis:    basis,
		ProbeConcurrencyLimit: 1,
		ProbeConcurrencyBasis: basis,
	}
}

func routeTargetKeyAllowsModel(key model.ChannelKey, modelName string) bool {
	if modelName == "" {
		return true
	}
	allowed := strings.TrimSpace(key.AllowedModels)
	if allowed == "" {
		return true
	}
	for _, part := range strings.Split(allowed, ",") {
		if model.NormalizeRouteTargetModelName(part) == modelName {
			return true
		}
	}
	return false
}

func validateRouteTargetOverrideTarget(channelID, channelKeyID int, modelName string) error {
	channel, ok := channelCache.Get(channelID)
	if !ok {
		return fmt.Errorf("channel not found")
	}
	key, ok := channelKeyCache.Get(channelKeyID)
	if !ok {
		return fmt.Errorf("channel key not found")
	}
	if key.ChannelID != channelID {
		return fmt.Errorf("invalid channel key id for channel")
	}
	normalizedModelName := model.NormalizeRouteTargetModelName(modelName)
	if normalizedModelName == "" {
		return fmt.Errorf("model name is required")
	}
	if !channel.SupportsModel(normalizedModelName) {
		return fmt.Errorf("invalid model for channel")
	}
	if !routeTargetKeyAllowsModel(key, normalizedModelName) {
		return fmt.Errorf("invalid model for channel key")
	}
	return nil
}

func normalizeRouteTargetOverrideStrict(row model.RouteTargetOverride) (model.RouteTargetOverride, error) {
	row.ModelName = model.NormalizeRouteTargetModelName(row.ModelName)
	if row.ChannelID <= 0 {
		return row, fmt.Errorf("invalid channel id")
	}
	if row.ChannelKeyID <= 0 {
		return row, fmt.Errorf("invalid channel key id")
	}
	if row.ModelName == "" {
		return row, fmt.Errorf("model name is required")
	}
	row.BillingMode = model.NormalizeBillingMode(row.BillingMode)
	if !model.IsValidBillingMode(row.BillingMode) {
		return row, fmt.Errorf("invalid billing mode: %q", row.BillingMode)
	}
	row.ProbePolicy = model.NormalizeProbePolicy(row.ProbePolicy)
	if !model.IsValidProbePolicy(row.ProbePolicy) {
		return row, fmt.Errorf("invalid probe policy: %q", row.ProbePolicy)
	}
	if row.ProbeIntervalSeconds <= 0 {
		row.ProbeIntervalSeconds = 3600
	}
	if row.ProbeConcurrencyLimit <= 0 {
		row.ProbeConcurrencyLimit = 1
	}
	return row, nil
}

func normalizeRouteTargetOverrideLenient(row model.RouteTargetOverride) (model.RouteTargetOverride, bool) {
	row.ModelName = model.NormalizeRouteTargetModelName(row.ModelName)
	if row.ChannelID <= 0 || row.ChannelKeyID <= 0 || row.ModelName == "" {
		return row, false
	}
	row.BillingMode = model.NormalizeBillingMode(row.BillingMode)
	if !model.IsValidBillingMode(row.BillingMode) {
		row.BillingMode = model.BillingModeUnknown
	}
	row.ProbePolicy = model.NormalizeProbePolicy(row.ProbePolicy)
	if !model.IsValidProbePolicy(row.ProbePolicy) {
		row.ProbePolicy = model.ProbePolicyPassiveOnly
	}
	if row.ProbeIntervalSeconds <= 0 {
		row.ProbeIntervalSeconds = 3600
	}
	if row.ProbeConcurrencyLimit <= 0 {
		row.ProbeConcurrencyLimit = 1
	}
	return row, true
}

func RouteTargetOverrideList(ctx context.Context) ([]model.RouteTargetOverride, error) {
	rows := routeTargetOverrideCache.Values()
	_ = ctx
	return rows, nil
}

func RouteTargetOverrideListByChannel(channelID int, ctx context.Context) ([]model.RouteTargetOverride, error) {
	if channelID <= 0 {
		return nil, fmt.Errorf("invalid channel id")
	}
	rows := make([]model.RouteTargetOverride, 0)
	for _, row := range routeTargetOverrideCache.Values() {
		if row.ChannelID != channelID {
			continue
		}
		rows = append(rows, row)
	}
	_ = ctx
	return rows, nil
}

func RouteTargetOverrideGet(channelID, channelKeyID int, modelName string) (model.RouteTargetOverride, bool) {
	return routeTargetOverrideIndex.Get(routeTargetOverrideLookupKey(channelID, channelKeyID, modelName))
}

func RouteTargetOverrideUpsert(row model.RouteTargetOverride, ctx context.Context) (model.RouteTargetOverride, error) {
	normalized, err := normalizeRouteTargetOverrideStrict(row)
	if err != nil {
		return model.RouteTargetOverride{}, err
	}
	if err := validateRouteTargetOverrideTarget(normalized.ChannelID, normalized.ChannelKeyID, normalized.ModelName); err != nil {
		return model.RouteTargetOverride{}, err
	}
	if err := db.GetDB().WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "channel_id"}, {Name: "channel_key_id"}, {Name: "model_name"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"billing_mode",
			"probe_policy",
			"probe_interval_seconds",
			"probe_concurrency_limit",
		}),
	}).Create(&normalized).Error; err != nil {
		return model.RouteTargetOverride{}, err
	}

	var stored model.RouteTargetOverride
	if err := db.GetDB().WithContext(ctx).
		Where("channel_id = ? AND channel_key_id = ? AND model_name = ?", normalized.ChannelID, normalized.ChannelKeyID, normalized.ModelName).
		First(&stored).Error; err != nil {
		return model.RouteTargetOverride{}, err
	}
	stored, _ = normalizeRouteTargetOverrideLenient(stored)
	routeTargetOverrideCache.Set(stored.ID, stored)
	routeTargetOverrideIndex.Set(routeTargetOverrideLookupKey(stored.ChannelID, stored.ChannelKeyID, stored.ModelName), stored)
	return stored, nil
}

func RouteTargetOverrideDeleteByModels(modelNames []string, ctx context.Context) error {
	if len(modelNames) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(modelNames))
	seen := make(map[string]struct{}, len(modelNames))
	for _, modelName := range modelNames {
		name := model.NormalizeRouteTargetModelName(modelName)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		normalized = append(normalized, name)
	}
	if len(normalized) == 0 {
		return nil
	}
	if err := db.GetDB().WithContext(ctx).Where("model_name IN ?", normalized).Delete(&model.RouteTargetOverride{}).Error; err != nil {
		return err
	}
	return routeTargetRefreshCache(ctx)
}

func RouteTargetOverrideDeleteByChannelAndModels(channelID int, modelNames []string, ctx context.Context) error {
	if channelID <= 0 || len(modelNames) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(modelNames))
	seen := make(map[string]struct{}, len(modelNames))
	for _, modelName := range modelNames {
		name := model.NormalizeRouteTargetModelName(modelName)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		normalized = append(normalized, name)
	}
	if len(normalized) == 0 {
		return nil
	}
	if err := db.GetDB().WithContext(ctx).Where("channel_id = ? AND model_name IN ?", channelID, normalized).Delete(&model.RouteTargetOverride{}).Error; err != nil {
		return err
	}
	return routeTargetRefreshCache(ctx)
}

func RouteTargetOverrideDelete(channelID, channelKeyID int, modelName string, ctx context.Context) error {
	if channelID <= 0 {
		return fmt.Errorf("invalid channel id")
	}
	if channelKeyID <= 0 {
		return fmt.Errorf("invalid channel key id")
	}
	modelName = model.NormalizeRouteTargetModelName(modelName)
	if modelName == "" {
		return fmt.Errorf("model name is required")
	}
	result := db.GetDB().WithContext(ctx).
		Where("channel_id = ? AND channel_key_id = ? AND model_name = ?", channelID, channelKeyID, modelName).
		Delete(&model.RouteTargetOverride{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("route target override not found")
	}
	return routeTargetRefreshCache(ctx)
}

func ResolveRouteTargetPolicy(channel *model.Channel, key model.ChannelKey, modelName string) model.RouteTargetResolvedPolicy {
	channelID := key.ChannelID
	if channel != nil && channel.ID > 0 {
		channelID = channel.ID
	}
	policy := defaultRouteTargetResolvedPolicy(channelID, key.ID, key.SourceType, modelName)
	if strings.TrimSpace(policy.ModelName) == "" {
		return policy
	}
	if info, err := LLMGet(policy.ModelName); err == nil {
		policy.BillingMode = info.BillingMode
		policy.BillingModeBasis = model.RouteTargetPolicyBasisModelDefault
		policy.ProbePolicy = info.ProbePolicy
		policy.ProbePolicyBasis = model.RouteTargetPolicyBasisModelDefault
		if info.ProbeIntervalSeconds > 0 {
			policy.ProbeIntervalSeconds = info.ProbeIntervalSeconds
		}
		policy.ProbeIntervalBasis = model.RouteTargetPolicyBasisModelDefault
		if info.ProbeConcurrencyLimit > 0 {
			policy.ProbeConcurrencyLimit = info.ProbeConcurrencyLimit
		}
		policy.ProbeConcurrencyBasis = model.RouteTargetPolicyBasisModelDefault
	}
	if override, ok := RouteTargetOverrideGet(channelID, key.ID, policy.ModelName); ok {
		policy.BillingMode = override.BillingMode
		policy.BillingModeBasis = model.RouteTargetPolicyBasisExplicitOverride
		policy.ProbePolicy = override.ProbePolicy
		policy.ProbePolicyBasis = model.RouteTargetPolicyBasisExplicitOverride
		policy.ProbeIntervalSeconds = override.ProbeIntervalSeconds
		policy.ProbeIntervalBasis = model.RouteTargetPolicyBasisExplicitOverride
		policy.ProbeConcurrencyLimit = override.ProbeConcurrencyLimit
		policy.ProbeConcurrencyBasis = model.RouteTargetPolicyBasisExplicitOverride
	}
	return policy
}

func routeTargetRefreshCache(ctx context.Context) error {
	rows := []model.RouteTargetOverride{}
	if err := db.GetDB().WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	routeTargetOverrideCache.Clear()
	routeTargetOverrideIndex.Clear()
	for _, row := range rows {
		normalized, ok := normalizeRouteTargetOverrideLenient(row)
		if !ok {
			continue
		}
		routeTargetOverrideCache.Set(normalized.ID, normalized)
		routeTargetOverrideIndex.Set(routeTargetOverrideLookupKey(normalized.ChannelID, normalized.ChannelKeyID, normalized.ModelName), normalized)
	}
	return nil
}
