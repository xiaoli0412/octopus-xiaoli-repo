package relay

import (
	"strings"

	dbmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/relay/balancer"
)

func init() {
	balancer.RelayEffectiveCircuitThresholdShim = func(keyID int, modelName string) int64 {
		key, ok := op.ChannelKeyGet(keyID)
		if !ok {
			return 0
		}
		return effectiveCircuitThresholdForRelay(key, modelName)
	}
}

type dynamicRoutingTuning struct {
	RaceAfterFails      int
	RaceConcurrency     int
	PreservedPriorities bool
}

func effectiveDynamicRoutingTuning(group dbmodel.Group, channel *dbmodel.Channel, key dbmodel.ChannelKey, requestModel string) dynamicRoutingTuning {
	enabled, err := op.SettingGetBool(dbmodel.SettingKeyDynamicRoutingHealthEnabled)
	if err != nil || !enabled {
		return defaultDynamicRoutingTuning(group)
	}
	return effectiveDynamicRoutingTuningWithPolicy(group, op.ResolveRouteTargetPolicy(channel, key, requestModel), true)
}

func defaultDynamicRoutingTuning(group dbmodel.Group) dynamicRoutingTuning {
	tuning := dynamicRoutingTuning{
		RaceAfterFails:      defaultPositiveInt(group.RaceAfterFails, 2),
		RaceConcurrency:     defaultPositiveInt(group.RaceConcurrency, 2),
		PreservedPriorities: true,
	}
	return tuning
}

func effectiveDynamicRoutingTuningWithPolicy(group dbmodel.Group, policy dbmodel.RouteTargetResolvedPolicy, healthEnabled bool) dynamicRoutingTuning {
	tuning := defaultDynamicRoutingTuning(group)
	if !healthEnabled {
		return tuning
	}
	tuning.RaceAfterFails = effectiveRaceAfterFails(group, policy)
	tuning.RaceConcurrency = effectiveRaceConcurrency(group, policy)
	return tuning
}

func defaultPositiveInt(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func normalizeSourceType(sourceType string) string {
	return dbmodel.EffectiveChannelKeySourceType(sourceType)
}

func effectiveRaceAfterFails(group dbmodel.Group, policy dbmodel.RouteTargetResolvedPolicy) int {
	threshold := defaultPositiveInt(group.RaceAfterFails, 2)
	billingMode := policy.BillingMode
	sourceType := normalizeSourceType(policy.SourceType)

	switch {
	case sourceType == dbmodel.ChannelKeySourceTypePaidMetered:
		if threshold < 3 {
			threshold = 3
		}
	case billingMode == dbmodel.BillingModePerRequest || billingMode == dbmodel.BillingModePerQuota || billingMode == dbmodel.BillingModeFlat:
		if threshold < 3 {
			threshold = 3
		}
	case billingMode == dbmodel.BillingModePerToken && allowsRacingByPolicy(policy):
		if threshold > 2 {
			threshold = 2
		}
	}

	return threshold
}

func effectiveRaceConcurrency(group dbmodel.Group, policy dbmodel.RouteTargetResolvedPolicy) int {
	concurrency := defaultPositiveInt(group.RaceConcurrency, 2)
	billingMode := policy.BillingMode
	sourceType := normalizeSourceType(policy.SourceType)

	switch {
	case sourceType == dbmodel.ChannelKeySourceTypePaidMetered:
		if concurrency > 1 {
			concurrency = 1
		}
	case billingMode == dbmodel.BillingModePerRequest || billingMode == dbmodel.BillingModePerQuota || billingMode == dbmodel.BillingModeFlat:
		if concurrency > 1 {
			concurrency = 1
		}
	case billingMode == dbmodel.BillingModeFree || strings.Contains(sourceType, "free") || strings.Contains(sourceType, "public"):
		if group.RaceConcurrency == 0 && concurrency < 2 {
			concurrency = 2
		}
	}

	if policy.ProbeConcurrencyLimit > 1 && concurrency > policy.ProbeConcurrencyLimit {
		concurrency = policy.ProbeConcurrencyLimit
	}
	if concurrency < 1 {
		concurrency = 1
	}
	return concurrency
}

func effectiveCircuitThresholdForRelay(key dbmodel.ChannelKey, requestModel string) int64 {
	return effectiveCircuitThresholdForRelayPolicy(op.ResolveRouteTargetPolicy(nil, key, requestModel))
}

func effectiveCircuitThresholdForRelayPolicy(policy dbmodel.RouteTargetResolvedPolicy) int64 {
	base := int64(5)
	if configured, err := op.SettingGetInt(dbmodel.SettingKeyCircuitBreakerThreshold); err == nil && configured > 0 {
		base = int64(configured)
	}
	billingMode := policy.BillingMode
	sourceType := normalizeSourceType(policy.SourceType)

	switch {
	case sourceType == dbmodel.ChannelKeySourceTypePaidMetered:
		if base > 4 {
			base = 4
		}
	case billingMode == dbmodel.BillingModePerRequest || billingMode == dbmodel.BillingModePerQuota || billingMode == dbmodel.BillingModeFlat:
		if base > 4 {
			base = 4
		}
	case billingMode == dbmodel.BillingModeFree || strings.Contains(sourceType, "free") || strings.Contains(sourceType, "public"):
		base++
	}
	if base < 2 {
		base = 2
	}
	return base
}
