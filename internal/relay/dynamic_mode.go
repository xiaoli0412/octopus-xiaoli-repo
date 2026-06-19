package relay

import (
	"fmt"
	"sort"
	"strings"

	dbmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/relay/balancer"
)

const (
	dynamicRoutingModeShadowAI          = "shadow-ai"
	dynamicRoutingModeHybrid            = "hybrid"
	dynamicRoutingModeMetricsOnly       = "metrics-only"
	dynamicRoutingModeStrict            = "strict-mechanism"
	dynamicRoutingModeIncidentSafe      = "incident-safe"
	dynamicRoutingDecisionDeterministic = "deterministic"
	dynamicRoutingDecisionRecommended   = "recommended"
	dynamicRoutingDecisionShadow        = "shadow"
	dynamicRoutingDecisionMetrics       = "metrics"
)

type relayDynamicAudit struct {
	Mode          string
	EffectiveMode string
	Decision      string
	Reason        string
	Confidence    float64
	Fallback      bool
	Recommended   []string
}

type dynamicRoutingModeState struct {
	Mode          string
	EffectiveMode string
	Decision      string
	Reason        string
	Confidence    float64
	Fallback      bool
	HealthEnabled bool
	AllowAdaptive bool
	AllowRace     bool
	Recommended   []dbmodel.GroupItem
}

type DynamicRoutingSummarySnapshot struct {
	CurrentMode    string
	HealthEnabled  bool
	EffectiveMode  string
	Decision       string
	DecisionReason string
}

type dynamicRoutingCandidateScore struct {
	Item         dbmodel.GroupItem
	ChannelName  string
	Reasons      []string
	Confidence   float64
	Score        float64
	Paid         bool
	CircuitOpen  bool
	FailRate     float64
	SuccessCount int64
	FailureCount int64
	AvgWaitMs    float64
	PriorityBias float64
	SupportsRace bool
}

func dynamicRoutingMode() string {
	mode, err := op.SettingGetString(dbmodel.SettingKeyDynamicRoutingMode)
	if err != nil {
		return dynamicRoutingModeHybrid
	}
	switch strings.TrimSpace(mode) {
	case dynamicRoutingModeShadowAI, dynamicRoutingModeHybrid, dynamicRoutingModeMetricsOnly, dynamicRoutingModeStrict, dynamicRoutingModeIncidentSafe:
		return strings.TrimSpace(mode)
	default:
		return dynamicRoutingModeHybrid
	}
}

func baseDynamicCandidates(group dbmodel.Group) []dbmodel.GroupItem {
	items := filterUpstreamSuppressedItems(group.Items)
	return balancer.GetBalancer(group.Mode).Candidates(items)
}

// filterUpstreamSuppressedItems removes group items whose channel is linked to a
// temporarily suppressed upstream site. This implements automatic multi-channel
// failover without mutating persisted group configuration.
func filterUpstreamSuppressedItems(items []dbmodel.GroupItem) []dbmodel.GroupItem {
	if len(items) == 0 {
		return nil
	}
	filtered := make([]dbmodel.GroupItem, 0, len(items))
	for _, item := range items {
		channel, err := op.ChannelGet(item.ChannelID, nil)
		if err != nil {
			// Keep the item when we cannot resolve its channel; suppression is
			// only applied when the channel is confirmed to be upstream-linked
			// and its upstream site is currently suppressed.
			filtered = append(filtered, item)
			continue
		}
		if channel.UpstreamSiteID > 0 && op.UpstreamSiteIsSuppressed(channel.UpstreamSiteID) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func initAIDynamicModeState(group dbmodel.Group, requestModel string, requestCapability string) *dynamicRoutingModeState {
	state := &dynamicRoutingModeState{
		Mode:          "ai-dynamic",
		EffectiveMode: "ai-dynamic",
		Decision:      dynamicRoutingDecisionRecommended,
		Reason:        "ai_dynamic_runtime_ranking",
		AllowAdaptive: true,
		AllowRace:     true,
	}

	healthEnabled, err := op.SettingGetBool(dbmodel.SettingKeyDynamicRoutingHealthEnabled)
	if err == nil {
		state.HealthEnabled = healthEnabled
	}

	base := sortByDynamicScore(baseDynamicCandidates(group), requestModel, requestCapability)
	state.Recommended = append(state.Recommended, base...)
	if len(base) == 0 {
		state.Decision = dynamicRoutingDecisionDeterministic
		state.Reason = "no_candidates"
		state.AllowAdaptive = false
		return state
	}

	confidenceSum := 0.0
	for _, score := range scoreDynamicCandidates(base, requestModel, requestCapability) {
		confidenceSum += score.Confidence
	}
	state.Confidence = confidenceSum / float64(len(base))
	return state
}

func sortByDynamicScore(items []dbmodel.GroupItem, requestModel string, requestCapability string) []dbmodel.GroupItem {
	scored := scoreDynamicCandidates(items, requestModel, requestCapability)
	ordered := make([]dbmodel.GroupItem, 0, len(scored))
	for _, item := range scored {
		ordered = append(ordered, item.Item)
	}
	return ordered
}

func initDynamicRoutingModeState(group dbmodel.Group, requestModel string, requestCapability string) *dynamicRoutingModeState {
	if group.Mode == dbmodel.GroupModeAIDynamic {
		return initAIDynamicModeState(group, requestModel, requestCapability)
	}

	mode := dynamicRoutingMode()
	state := &dynamicRoutingModeState{
		Mode:          mode,
		EffectiveMode: mode,
		Decision:      dynamicRoutingDecisionDeterministic,
		Reason:        "deterministic_mechanism_default",
		Confidence:    0,
		AllowAdaptive: false,
		AllowRace:     mode != dynamicRoutingModeIncidentSafe,
	}

	healthEnabled, err := op.SettingGetBool(dbmodel.SettingKeyDynamicRoutingHealthEnabled)
	if err == nil {
		state.HealthEnabled = healthEnabled
	}

	base := baseDynamicCandidates(group)
	state.Recommended = append(state.Recommended, base...)

	if len(base) == 0 {
		state.Reason = "no_candidates"
		state.AllowRace = false
		return state
	}

	if mode == dynamicRoutingModeStrict {
		state.EffectiveMode = dynamicRoutingModeStrict
		state.Decision = dynamicRoutingDecisionDeterministic
		state.Reason = "strict_mechanism_forces_deterministic_path"
		return state
	}

	if mode == dynamicRoutingModeIncidentSafe {
		state.EffectiveMode = dynamicRoutingModeIncidentSafe
		state.Decision = dynamicRoutingDecisionDeterministic
		state.Reason = "incident_safe_forces_conservative_path"
		state.AllowRace = false
		return state
	}

	if !state.HealthEnabled {
		state.Fallback = mode == dynamicRoutingModeHybrid
		if state.Fallback {
			state.EffectiveMode = dynamicRoutingModeStrict
		}
		state.Decision = map[string]string{
			dynamicRoutingModeShadowAI:    dynamicRoutingDecisionShadow,
			dynamicRoutingModeMetricsOnly: dynamicRoutingDecisionMetrics,
		}[mode]
		if state.Decision == "" {
			state.Decision = dynamicRoutingDecisionDeterministic
		}
		state.Reason = "dynamic_health_disabled"
		return state
	}

	scored := scoreDynamicCandidates(base, requestModel, requestCapability)
	state.Recommended = make([]dbmodel.GroupItem, 0, len(scored))
	recommendedLabels := make([]string, 0, len(scored))
	confidenceSum := 0.0
	for _, score := range scored {
		state.Recommended = append(state.Recommended, score.Item)
		recommendedLabels = append(recommendedLabels, dynamicRoutingCandidateLabel(score.Item))
		confidenceSum += score.Confidence
	}
	state.Confidence = confidenceSum / float64(len(scored))
	state.Reason = strings.Join(recommendedLabels, ",")

	safeToAdopt := state.Confidence >= 0.55
	if mode == dynamicRoutingModeMetricsOnly {
		state.Decision = dynamicRoutingDecisionMetrics
		state.Reason = "metrics_only_records_recommendation_without_mutation"
		return state
	}
	if mode == dynamicRoutingModeShadowAI {
		state.Decision = dynamicRoutingDecisionShadow
		state.Reason = "shadow_ai_records_recommendation_without_mutation"
		return state
	}

	if safeToAdopt {
		state.Decision = dynamicRoutingDecisionRecommended
		state.Reason = "hybrid_adopted_runtime_recommendation"
		state.AllowAdaptive = true
		state.AllowRace = true
		return state
	}

	state.Fallback = true
	state.EffectiveMode = dynamicRoutingModeStrict
	state.Decision = dynamicRoutingDecisionDeterministic
	state.Reason = "hybrid_fallback_low_confidence"
	return state
}

func scoreDynamicCandidates(items []dbmodel.GroupItem, requestModel string, requestCapability string) []dynamicRoutingCandidateScore {
	scores := make([]dynamicRoutingCandidateScore, 0, len(items))
	for idx, item := range items {
		targetModel := strings.TrimSpace(item.ModelName)
		if targetModel == "" {
			targetModel = requestModel
		}
		channel, err := op.ChannelGet(item.ChannelID, nil)
		score := dynamicRoutingCandidateScore{Item: item, Confidence: 0.35}
		if err != nil || channel == nil {
			score.Score = -9999
			score.Reasons = append(score.Reasons, "channel_unavailable")
			scores = append(scores, score)
			continue
		}
		score.ChannelName = channel.Name
		if !channel.Enabled {
			score.Score = -9998
			score.Reasons = append(score.Reasons, "channel_disabled")
			scores = append(scores, score)
			continue
		}
		targetCapability := resolveRequestCapability(channel, targetModel, requestCapability)
		if !op.ChannelCanServeRequest(channel.ID, targetModel, targetCapability) {
			score.Score = -9997
			score.Reasons = append(score.Reasons, "stale_route_target")
			scores = append(scores, score)
			continue
		}

		key := dynamicRoutingScoreKey(channel, targetModel, targetCapability)
		if strings.TrimSpace(key.ChannelKey) == "" {
			score.Score = -9996
			score.Reasons = append(score.Reasons, "no_eligible_key")
			scores = append(scores, score)
			continue
		}
		policy := op.ResolveRouteTargetPolicy(channel, key, targetModel)
		stats, _ := op.StatsChannelSnapshot(channel.ID)
		total := stats.RequestSuccess + stats.RequestFailed
		if total > 0 {
			score.SuccessCount = stats.RequestSuccess
			score.FailureCount = stats.RequestFailed
			score.FailRate = float64(stats.RequestFailed) / float64(total)
			score.Confidence += 0.15
		}
		if total > 0 {
			score.AvgWaitMs = float64(stats.WaitTime) / float64(total)
		}
		score.SupportsRace = allowsRacingByPolicy(policy)
		score.Paid = normalizeSourceType(policy.SourceType) == dbmodel.ChannelKeySourceTypePaidMetered ||
			policy.BillingMode == dbmodel.BillingModePerRequest ||
			policy.BillingMode == dbmodel.BillingModePerQuota ||
			policy.BillingMode == dbmodel.BillingModeFlat
		tripped, _ := balancer.IsTripped(channel.ID, key.ID, targetModel)
		score.CircuitOpen = tripped
		score.PriorityBias = float64(len(items)-idx) * 0.05

		score.Score = score.PriorityBias
		if score.CircuitOpen {
			score.Score -= 4
			score.Reasons = append(score.Reasons, "circuit_open")
		} else {
			score.Score += 1.2
			score.Confidence += 0.1
		}
		if score.FailRate > 0 {
			score.Score -= score.FailRate * 3.5
			score.Reasons = append(score.Reasons, fmt.Sprintf("fail_rate_%.2f", score.FailRate))
		} else {
			score.Score += 0.6
			if total > 0 {
				score.Reasons = append(score.Reasons, "healthy_recent_successes")
			}
		}
		if score.AvgWaitMs > 0 {
			if score.AvgWaitMs < 1200 {
				score.Score += 0.7
				score.Reasons = append(score.Reasons, "low_wait")
			} else if score.AvgWaitMs > 6000 {
				score.Score -= 0.8
				score.Reasons = append(score.Reasons, "high_wait")
			}
		}
		if score.Paid {
			score.Score -= 0.3
			score.Reasons = append(score.Reasons, "paid_conservative_bias")
		} else {
			score.Score += 0.25
			score.Reasons = append(score.Reasons, "free_public_bias")
		}
		if score.SupportsRace {
			score.Score += 0.15
		}
		if learning, ok := dynamicRoutingLearningBias(channel.ID, key.ID, targetModel); ok {
			score.Score += learning.Score
			score.Confidence += learning.Confidence * 0.2
			score.Reasons = append(score.Reasons, "local_learning_signal")
		}
		if score.Confidence > 1 {
			score.Confidence = 1
		}
		scores = append(scores, score)
	}

	sort.SliceStable(scores, func(i, j int) bool {
		if scores[i].Score != scores[j].Score {
			return scores[i].Score > scores[j].Score
		}
		if scores[i].Item.Priority != scores[j].Item.Priority {
			return scores[i].Item.Priority < scores[j].Item.Priority
		}
		return scores[i].Item.ChannelID < scores[j].Item.ChannelID
	})
	return scores
}

func dynamicRoutingLearningBias(channelID, keyID int, modelName string) (dbmodel.DynamicRouteLearningState, bool) {
	enabled, err := op.SettingGetBool(dbmodel.SettingKeyDynamicRoutingLearningEnabled)
	if err != nil || !enabled {
		return dbmodel.DynamicRouteLearningState{}, false
	}
	return op.DynamicRouteLearningGet(channelID, keyID, modelName)
}

func dynamicRoutingScoreKey(channel *dbmodel.Channel, modelName string, requestCapability string) dbmodel.ChannelKey {
	if channel == nil {
		return dbmodel.ChannelKey{}
	}
	eligible := channel.EligibleChannelKeysForRequest(modelName, requestCapability)
	if len(eligible) == 0 {
		return dbmodel.ChannelKey{}
	}
	return eligible[0]
}

func dynamicRoutingCandidateLabel(item dbmodel.GroupItem) string {
	modelName := strings.TrimSpace(item.ModelName)
	if modelName == "" {
		modelName = "default"
	}
	return fmt.Sprintf("%d:%s", item.ChannelID, modelName)
}

func (s *dynamicRoutingModeState) recommendedLabels() []string {
	if s == nil || len(s.Recommended) == 0 {
		return nil
	}
	out := make([]string, 0, len(s.Recommended))
	for _, item := range s.Recommended {
		out = append(out, dynamicRoutingCandidateLabel(item))
	}
	return out
}

func (s *dynamicRoutingModeState) apply(req *relayRequest) {
	if req == nil || req.metrics == nil || s == nil {
		return
	}
	req.dynamicMode = s
	req.metrics.Dynamic = &relayDynamicAudit{
		Mode:          s.Mode,
		EffectiveMode: s.EffectiveMode,
		Decision:      s.Decision,
		Reason:        s.Reason,
		Confidence:    s.Confidence,
		Fallback:      s.Fallback,
		Recommended:   s.recommendedLabels(),
	}
}

func dynamicIterator(group dbmodel.Group, apiKeyID int, requestModel string, state *dynamicRoutingModeState) *balancer.Iterator {
	base := baseDynamicCandidates(group)
	if state == nil || !state.AllowAdaptive || len(state.Recommended) == 0 {
		return balancer.NewIteratorWithCandidates(group, apiKeyID, requestModel, base)
	}
	return balancer.NewIteratorWithCandidates(group, apiKeyID, requestModel, state.Recommended)
}

func effectiveDynamicRoutingTuningForMode(group dbmodel.Group, policy dbmodel.RouteTargetResolvedPolicy, state *dynamicRoutingModeState) dynamicRoutingTuning {
	healthEnabled := state == nil || state.HealthEnabled
	if state != nil {
		switch state.EffectiveMode {
		case dynamicRoutingModeStrict, dynamicRoutingModeMetricsOnly, dynamicRoutingModeShadowAI:
			return defaultDynamicRoutingTuning(group)
		}
	}
	tuning := effectiveDynamicRoutingTuningWithPolicy(group, policy, healthEnabled)
	if state == nil {
		return tuning
	}
	switch state.EffectiveMode {
	case dynamicRoutingModeIncidentSafe:
		tuning.RaceAfterFails = max(4, tuning.RaceAfterFails)
		tuning.RaceConcurrency = 1
	}
	return tuning
}

func shouldEscalateToRaceWithMode(group dbmodel.Group, policy dbmodel.RouteTargetResolvedPolicy, consecutiveFails int, tuning dynamicRoutingTuning, state *dynamicRoutingModeState) bool {
	if state != nil && !state.AllowRace {
		return false
	}
	return shouldEscalateToRace(group, policy, consecutiveFails, tuning)
}

func dynamicRoutingSummarySnapshot() DynamicRoutingSummarySnapshot {
	mode := dynamicRoutingMode()
	healthEnabled, _ := op.SettingGetBool(dbmodel.SettingKeyDynamicRoutingHealthEnabled)
	effectiveMode := mode
	decision := dynamicRoutingDecisionDeterministic
	if !healthEnabled && mode == dynamicRoutingModeHybrid {
		effectiveMode = dynamicRoutingModeStrict
	}
	switch mode {
	case dynamicRoutingModeHybrid:
		if healthEnabled {
			decision = dynamicRoutingDecisionRecommended
		}
	case dynamicRoutingModeShadowAI:
		decision = dynamicRoutingDecisionShadow
	case dynamicRoutingModeMetricsOnly:
		decision = dynamicRoutingDecisionMetrics
	}
	reason := "summary_snapshot_runtime_modes"
	if !healthEnabled {
		reason = "summary_snapshot_health_disabled"
	}
	return DynamicRoutingSummarySnapshot{
		CurrentMode:    mode,
		HealthEnabled:  healthEnabled,
		EffectiveMode:  effectiveMode,
		Decision:       decision,
		DecisionReason: reason,
	}
}

func DynamicRoutingSummarySnapshotForTask() DynamicRoutingSummarySnapshot {
	return dynamicRoutingSummarySnapshot()
}
