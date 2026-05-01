package relay

import (
	"testing"
	"time"

	dbmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	tmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
)

func TestAllowsRacingByDefault(t *testing.T) {
	cases := []struct {
		name       string
		sourceType string
		want       bool
	}{
		{name: "empty now stays conservative", sourceType: "", want: false},
		{name: "paid disabled", sourceType: "paid", want: false},
		{name: "metered disabled", sourceType: "metered", want: false},
		{name: "paid metered disabled", sourceType: " paid/metered ", want: false},
		{name: "private internal disabled", sourceType: "private/internal", want: false},
		{name: "public allowed", sourceType: "public/free", want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := allowsRacingByDefault(tc.sourceType)
			if got != tc.want {
				t.Fatalf("allowsRacingByDefault(%q) = %t, want %t", tc.sourceType, got, tc.want)
			}
		})
	}
}

func TestIsStreamingRequest(t *testing.T) {
	streamTrue := true
	streamFalse := false

	cases := []struct {
		name string
		req  *tmodel.InternalLLMRequest
		want bool
	}{
		{name: "nil request", req: nil, want: false},
		{name: "nil stream pointer", req: &tmodel.InternalLLMRequest{}, want: false},
		{name: "stream false", req: &tmodel.InternalLLMRequest{Stream: &streamFalse}, want: false},
		{name: "stream true", req: &tmodel.InternalLLMRequest{Stream: &streamTrue}, want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isStreamingRequest(tc.req)
			if got != tc.want {
				t.Fatalf("isStreamingRequest() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestShouldEscalateToRace(t *testing.T) {
	baseGroup := dbmodel.Group{
		Mode:              dbmodel.GroupModeFailover,
		RaceAfterFails:    2,
		RaceConcurrency:   3,
		FailoverWindowSec: 360,
	}
	freeKey := dbmodel.ChannelKey{SourceType: "public/free"}
	paidKey := dbmodel.ChannelKey{SourceType: "paid"}

	cases := []struct {
		name             string
		group            dbmodel.Group
		channel          *dbmodel.Channel
		key              dbmodel.ChannelKey
		requestModel     string
		consecutiveFails int
		want             bool
	}{
		{
			name:             "below threshold does not race",
			group:            baseGroup,
			channel:          nil,
			key:              freeKey,
			requestModel:     "",
			consecutiveFails: 1,
			want:             false,
		},
		{
			name: "non failover mode does not race",
			group: dbmodel.Group{
				Mode:              dbmodel.GroupModeRoundRobin,
				RaceAfterFails:    2,
				RaceConcurrency:   3,
				FailoverWindowSec: 360,
			},
			channel:          nil,
			key:              freeKey,
			requestModel:     "",
			consecutiveFails: 2,
			want:             false,
		},
		{
			name:             "paid source type does not race by default",
			group:            baseGroup,
			channel:          nil,
			key:              paidKey,
			requestModel:     "",
			consecutiveFails: 2,
			want:             false,
		},
		{
			name: "missing concurrency does not race",
			group: dbmodel.Group{
				Mode:              dbmodel.GroupModeFailover,
				RaceAfterFails:    2,
				RaceConcurrency:   1,
				FailoverWindowSec: 360,
			},
			channel:          nil,
			key:              freeKey,
			requestModel:     "",
			consecutiveFails: 2,
			want:             false,
		},
		{
			name: "missing window does not race",
			group: dbmodel.Group{
				Mode:              dbmodel.GroupModeFailover,
				RaceAfterFails:    2,
				RaceConcurrency:   3,
				FailoverWindowSec: 0,
			},
			channel:          nil,
			key:              freeKey,
			requestModel:     "",
			consecutiveFails: 2,
			want:             false,
		},
		{
			name:             "free source with empty model can race",
			group:            baseGroup,
			channel:          nil,
			key:              freeKey,
			requestModel:     "",
			consecutiveFails: 2,
			want:             true,
		},
		{
			name: "default threshold falls back to two",
			group: dbmodel.Group{
				Mode:              dbmodel.GroupModeFailover,
				RaceAfterFails:    0,
				RaceConcurrency:   3,
				FailoverWindowSec: 360,
			},
			channel:          nil,
			key:              freeKey,
			requestModel:     "",
			consecutiveFails: 2,
			want:             true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tuning := effectiveDynamicRoutingTuning(tc.group, nil, tc.key, tc.requestModel)
			policy := op.ResolveRouteTargetPolicy(tc.channel, tc.key, tc.requestModel)
			got := shouldEscalateToRace(tc.group, policy, tc.consecutiveFails, tuning)
			if got != tc.want {
				t.Fatalf("shouldEscalateToRace() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestShouldEscalateToRaceHonorsRouteTargetOverride(t *testing.T) {
	ctx := setupRelayTestDB(t)

	channel := &dbmodel.Channel{Name: "race-override-channel", Enabled: true, Model: "gpt-4o"}
	if err := op.ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	updated, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: channel.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{
		Enabled:       true,
		ChannelKey:    "race-override-key",
		SourceType:    "public/free",
		AllowedModels: "gpt-4o",
	}}}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(keys) error = %v", err)
	}
	if len(updated.Keys) != 1 {
		t.Fatalf("updated.Keys len = %d, want 1", len(updated.Keys))
	}
	key := updated.Keys[0]

	if err := op.LLMCreate(dbmodel.LLMInfo{Name: "gpt-4o", BillingMode: dbmodel.BillingModePerRequest}, ctx); err != nil {
		t.Fatalf("LLMCreate() error = %v", err)
	}
	if _, err := op.RouteTargetOverrideUpsert(dbmodel.RouteTargetOverride{
		ChannelID:    updated.ID,
		ChannelKeyID: key.ID,
		ModelName:    "gpt-4o",
		BillingMode:  dbmodel.BillingModeFree,
	}, ctx); err != nil {
		t.Fatalf("RouteTargetOverrideUpsert() error = %v", err)
	}

	group := dbmodel.Group{Mode: dbmodel.GroupModeFailover, RaceAfterFails: 2, RaceConcurrency: 3, FailoverWindowSec: 360}
	tuning := effectiveDynamicRoutingTuning(group, updated, key, "gpt-4o")
	policy := op.ResolveRouteTargetPolicy(updated, key, "gpt-4o")
	if got := shouldEscalateToRace(group, policy, 2, tuning); !got {
		t.Fatalf("shouldEscalateToRace() = %t, want true when route-target override permits racing", got)
	}
}

func TestShouldEscalateToRaceRejectsRouteTargetOverrideThatForbidsRacing(t *testing.T) {
	ctx := setupRelayTestDB(t)

	channel := &dbmodel.Channel{Name: "race-override-block-channel", Enabled: true, Model: "gpt-4o"}
	if err := op.ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	updated, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: channel.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{
		Enabled:       true,
		ChannelKey:    "race-override-block-key",
		SourceType:    "public/free",
		AllowedModels: "gpt-4o",
	}}}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(keys) error = %v", err)
	}
	if len(updated.Keys) != 1 {
		t.Fatalf("updated.Keys len = %d, want 1", len(updated.Keys))
	}
	key := updated.Keys[0]

	if err := op.LLMCreate(dbmodel.LLMInfo{Name: "gpt-4o", BillingMode: dbmodel.BillingModePerRequest}, ctx); err != nil {
		t.Fatalf("LLMCreate() error = %v", err)
	}
	if _, err := op.RouteTargetOverrideUpsert(dbmodel.RouteTargetOverride{
		ChannelID:    updated.ID,
		ChannelKeyID: key.ID,
		ModelName:    "gpt-4o",
		BillingMode:  dbmodel.BillingModePerRequest,
	}, ctx); err != nil {
		t.Fatalf("RouteTargetOverrideUpsert() error = %v", err)
	}

	group := dbmodel.Group{Mode: dbmodel.GroupModeFailover, RaceAfterFails: 2, RaceConcurrency: 3, FailoverWindowSec: 360}
	tuning := effectiveDynamicRoutingTuning(group, updated, key, "gpt-4o")
	policy := op.ResolveRouteTargetPolicy(updated, key, "gpt-4o")
	if got := shouldEscalateToRace(group, policy, 2, tuning); got {
		t.Fatalf("shouldEscalateToRace() = %t, want false when route-target override forbids racing", got)
	}
}

func TestEffectiveDynamicRoutingTuningPreservesPriorityWhileAdjustingThresholds(t *testing.T) {
	group := dbmodel.Group{
		Mode:              dbmodel.GroupModeFailover,
		RaceAfterFails:    2,
		RaceConcurrency:   3,
		FailoverWindowSec: 360,
	}
	key := dbmodel.ChannelKey{SourceType: "public/free"}

	tuning := effectiveDynamicRoutingTuning(group, nil, key, "")
	if !tuning.PreservedPriorities {
		t.Fatal("dynamic tuning should never change user priority ordering")
	}
	if tuning.RaceAfterFails > group.RaceAfterFails {
		t.Fatalf("RaceAfterFails = %d, want <= %d for free/public routing", tuning.RaceAfterFails, group.RaceAfterFails)
	}
	if tuning.RaceConcurrency < group.RaceConcurrency {
		t.Fatalf("RaceConcurrency = %d, want >= %d for free/public routing", tuning.RaceConcurrency, group.RaceConcurrency)
	}
}

func TestEffectiveDynamicRoutingTuningWithPolicyMatchesPublicHelper(t *testing.T) {
	ctx := setupRelayTestDB(t)
	group := dbmodel.Group{
		Mode:              dbmodel.GroupModeFailover,
		RaceAfterFails:    2,
		RaceConcurrency:   3,
		FailoverWindowSec: 360,
	}
	channel := &dbmodel.Channel{Name: "tuning-match-channel", Enabled: true, Model: "gpt-4o"}
	if err := op.ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	updated, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: channel.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{
		Enabled:       true,
		ChannelKey:    "tuning-match-key",
		SourceType:    dbmodel.ChannelKeySourceTypePublicFree,
		AllowedModels: "gpt-4o",
	}}}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(keys) error = %v", err)
	}
	key := updated.Keys[0]
	if err := op.SettingSetString(dbmodel.SettingKeyDynamicRoutingHealthEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(dynamic routing health) error = %v", err)
	}

	policy := op.ResolveRouteTargetPolicy(updated, key, "gpt-4o")
	got := effectiveDynamicRoutingTuningWithPolicy(group, policy, true)
	want := effectiveDynamicRoutingTuning(group, updated, key, "gpt-4o")
	if got != want {
		t.Fatalf("effectiveDynamicRoutingTuningWithPolicy() = %#v, want %#v", got, want)
	}
}

func TestEffectiveRaceConcurrencyPaidAndFreeProfiles(t *testing.T) {
	group := dbmodel.Group{RaceConcurrency: 4}
	if got := effectiveRaceConcurrency(group, dbmodel.RouteTargetResolvedPolicy{SourceType: dbmodel.ChannelKeySourceTypePaidMetered}); got != 1 {
		t.Fatalf("paid race concurrency = %d, want 1", got)
	}
	if got := effectiveRaceConcurrency(group, dbmodel.RouteTargetResolvedPolicy{SourceType: dbmodel.ChannelKeySourceTypePublicFree, BillingMode: dbmodel.BillingModeFree}); got < 2 {
		t.Fatalf("free/public race concurrency = %d, want >= 2", got)
	}
}

func TestEffectiveRaceConcurrencyRespectsExplicitUserLimit(t *testing.T) {
	group := dbmodel.Group{RaceConcurrency: 1}
	if got := effectiveRaceConcurrency(group, dbmodel.RouteTargetResolvedPolicy{SourceType: dbmodel.ChannelKeySourceTypePublicFree, BillingMode: dbmodel.BillingModeFree}); got != 1 {
		t.Fatalf("explicit user race concurrency = %d, want 1", got)
	}
}

func TestEffectiveCircuitThresholdForRelayPaidAndFreeProfiles(t *testing.T) {
	paid := effectiveCircuitThresholdForRelayPolicy(dbmodel.RouteTargetResolvedPolicy{SourceType: dbmodel.ChannelKeySourceTypePaidMetered})
	free := effectiveCircuitThresholdForRelayPolicy(dbmodel.RouteTargetResolvedPolicy{SourceType: dbmodel.ChannelKeySourceTypePublicFree, BillingMode: dbmodel.BillingModeFree})
	if paid >= free {
		t.Fatalf("paid threshold = %d, free/public threshold = %d, want paid more conservative than free/public", paid, free)
	}
}

func TestDynamicRoutingModeStateHybridFallbackWhenHealthDisabled(t *testing.T) {
	ctx := setupRelayTestDB(t)
	if err := op.SettingSetString(dbmodel.SettingKeyDynamicRoutingMode, "hybrid"); err != nil {
		t.Fatalf("SettingSetString(mode) error = %v", err)
	}
	if err := op.SettingSetString(dbmodel.SettingKeyDynamicRoutingHealthEnabled, "false"); err != nil {
		t.Fatalf("SettingSetString(health) error = %v", err)
	}

	group := dbmodel.Group{ID: 1, Mode: dbmodel.GroupModeFailover, Items: []dbmodel.GroupItem{{ChannelID: 1, ModelName: "gpt-4o", Priority: 1}}}
	_ = ctx
	state := initDynamicRoutingModeState(group, "gpt-4o")
	if state.Mode != dynamicRoutingModeHybrid {
		t.Fatalf("Mode = %q, want hybrid", state.Mode)
	}
	if state.EffectiveMode != dynamicRoutingModeStrict {
		t.Fatalf("EffectiveMode = %q, want strict-mechanism fallback", state.EffectiveMode)
	}
	if !state.Fallback {
		t.Fatal("Fallback = false, want true")
	}
	if state.Decision != dynamicRoutingDecisionDeterministic {
		t.Fatalf("Decision = %q, want deterministic", state.Decision)
	}
}

func TestDynamicRoutingModeStateIncidentSafeDisablesRace(t *testing.T) {
	ctx := setupRelayTestDB(t)
	if err := op.SettingSetString(dbmodel.SettingKeyDynamicRoutingMode, "incident-safe"); err != nil {
		t.Fatalf("SettingSetString(mode) error = %v", err)
	}
	if err := op.SettingSetString(dbmodel.SettingKeyDynamicRoutingHealthEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(health) error = %v", err)
	}

	group := dbmodel.Group{ID: 1, Mode: dbmodel.GroupModeFailover, RaceAfterFails: 2, RaceConcurrency: 3, FailoverWindowSec: 360}
	state := initDynamicRoutingModeState(group, "gpt-4o")
	tuning := effectiveDynamicRoutingTuningForMode(group, dbmodel.RouteTargetResolvedPolicy{SourceType: dbmodel.ChannelKeySourceTypePublicFree, BillingMode: dbmodel.BillingModeFree}, state)
	if state.EffectiveMode != dynamicRoutingModeIncidentSafe {
		t.Fatalf("EffectiveMode = %q, want incident-safe", state.EffectiveMode)
	}
	if state.AllowRace {
		t.Fatal("AllowRace = true, want false")
	}
	if tuning.RaceConcurrency != 1 {
		t.Fatalf("RaceConcurrency = %d, want 1", tuning.RaceConcurrency)
	}
	_ = ctx
}

func TestDynamicRoutingScoringDoesNotAdvanceRoundRobinKeyState(t *testing.T) {
	ctx := setupRelayTestDB(t)
	if err := op.SettingSetString(dbmodel.SettingKeyDynamicRoutingMode, "hybrid"); err != nil {
		t.Fatalf("SettingSetString(mode) error = %v", err)
	}
	if err := op.SettingSetString(dbmodel.SettingKeyDynamicRoutingHealthEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(health) error = %v", err)
	}

	channel := &dbmodel.Channel{
		Name:             "dynamic-score-rr-channel",
		Enabled:          true,
		Model:            "gpt-4o",
		KeyRoutingPolicy: dbmodel.KeyRoutingPolicyRoundRobin,
	}
	if err := op.ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	updated, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: channel.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{
		{Enabled: true, ChannelKey: "rr-key-a", SourceType: dbmodel.ChannelKeySourceTypePublicFree, AllowedModels: "gpt-4o"},
		{Enabled: true, ChannelKey: "rr-key-b", SourceType: dbmodel.ChannelKeySourceTypePublicFree, AllowedModels: "gpt-4o"},
	}}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(keys) error = %v", err)
	}
	if len(updated.Keys) != 2 {
		t.Fatalf("updated keys = %d, want 2", len(updated.Keys))
	}

	group := dbmodel.Group{
		ID:    1,
		Mode:  dbmodel.GroupModeFailover,
		Items: []dbmodel.GroupItem{{ChannelID: updated.ID, ModelName: "gpt-4o", Priority: 1}},
	}
	_ = initDynamicRoutingModeState(group, "gpt-4o")

	refreshed, err := op.ChannelGet(updated.ID, ctx)
	if err != nil {
		t.Fatalf("ChannelGet() error = %v", err)
	}
	firstRuntimeKey := refreshed.GetChannelKeyForModel("gpt-4o")
	if firstRuntimeKey.ChannelKey != "rr-key-a" {
		t.Fatalf("first runtime key after scoring = %q, want rr-key-a", firstRuntimeKey.ChannelKey)
	}
}

func TestEffectiveDynamicRoutingTuningForModeStrictUsesDefaultMechanismTuning(t *testing.T) {
	group := dbmodel.Group{RaceAfterFails: 5, RaceConcurrency: 4}
	state := &dynamicRoutingModeState{EffectiveMode: dynamicRoutingModeStrict, HealthEnabled: true}
	policy := dbmodel.RouteTargetResolvedPolicy{SourceType: dbmodel.ChannelKeySourceTypePaidMetered, BillingMode: dbmodel.BillingModePerRequest}

	got := effectiveDynamicRoutingTuningForMode(group, policy, state)
	want := defaultDynamicRoutingTuning(group)
	if got != want {
		t.Fatalf("strict tuning = %#v, want default mechanism tuning %#v", got, want)
	}
}

func TestDynamicRoutingScoringDemotesCandidateWithoutEligibleKey(t *testing.T) {
	ctx := setupRelayTestDB(t)
	if err := op.SettingSetString(dbmodel.SettingKeyDynamicRoutingMode, dynamicRoutingModeHybrid); err != nil {
		t.Fatalf("SettingSetString(mode) error = %v", err)
	}
	if err := op.SettingSetString(dbmodel.SettingKeyDynamicRoutingHealthEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(health) error = %v", err)
	}

	cooling := &dbmodel.Channel{Name: "dynamic-score-cooling", Enabled: true, Model: "gpt-4o"}
	if err := op.ChannelCreate(cooling, ctx); err != nil {
		t.Fatalf("ChannelCreate(cooling) error = %v", err)
	}
	cooling, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: cooling.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{
		Enabled:       true,
		ChannelKey:    "cooling-key",
		SourceType:    dbmodel.ChannelKeySourceTypePublicFree,
		AllowedModels: "gpt-4o",
	}}}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(cooling) error = %v", err)
	}
	coolingKey := cooling.Keys[0]
	coolingKey.StatusCode = 429
	coolingKey.LastUseTimeStamp = time.Now().Unix()
	if err := op.ChannelKeyUpdate(coolingKey); err != nil {
		t.Fatalf("ChannelKeyUpdate(cooling key) error = %v", err)
	}

	healthy := &dbmodel.Channel{Name: "dynamic-score-eligible", Enabled: true, Model: "gpt-4o"}
	if err := op.ChannelCreate(healthy, ctx); err != nil {
		t.Fatalf("ChannelCreate(healthy) error = %v", err)
	}
	healthy, err = op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: healthy.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{
		Enabled:       true,
		ChannelKey:    "eligible-key",
		SourceType:    dbmodel.ChannelKeySourceTypePublicFree,
		AllowedModels: "gpt-4o",
	}}}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(healthy) error = %v", err)
	}

	group := dbmodel.Group{
		ID:   1,
		Mode: dbmodel.GroupModeFailover,
		Items: []dbmodel.GroupItem{
			{ID: 1, ChannelID: cooling.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1},
			{ID: 2, ChannelID: healthy.ID, ModelName: "gpt-4o", Priority: 2, Weight: 1},
		},
	}
	state := initDynamicRoutingModeState(group, "gpt-4o")
	if len(state.Recommended) == 0 || state.Recommended[0].ChannelID != healthy.ID {
		t.Fatalf("recommended order = %#v, want eligible channel %d first", state.Recommended, healthy.ID)
	}
}

func TestDynamicRoutingHybridAdoptsHealthyRecommendationOrder(t *testing.T) {
	group, unhealthyID, healthyID := setupDynamicRoutingRecommendationTest(t, dynamicRoutingModeHybrid)

	state := initDynamicRoutingModeState(group, "gpt-4o")
	if state.Decision != dynamicRoutingDecisionRecommended {
		t.Fatalf("Decision = %q, want recommended", state.Decision)
	}
	if !state.AllowAdaptive {
		t.Fatal("AllowAdaptive = false, want true")
	}
	if len(state.Recommended) == 0 || state.Recommended[0].ChannelID != healthyID {
		t.Fatalf("recommended order = %#v, want healthy channel %d first", state.Recommended, healthyID)
	}

	iter := dynamicIterator(group, 0, "gpt-4o", state)
	if !iter.Next() {
		t.Fatal("iterator should have first candidate")
	}
	if got := iter.Item().ChannelID; got != healthyID {
		t.Fatalf("hybrid iterator first channel = %d, want healthy channel %d", got, healthyID)
	}
	if unhealthyID == healthyID {
		t.Fatalf("test setup invalid: unhealthyID == healthyID == %d", unhealthyID)
	}
}

func TestDynamicRoutingObservationModesKeepMechanismOrder(t *testing.T) {
	cases := []struct {
		name         string
		mode         string
		wantDecision string
	}{
		{name: "metrics only", mode: dynamicRoutingModeMetricsOnly, wantDecision: dynamicRoutingDecisionMetrics},
		{name: "shadow ai", mode: dynamicRoutingModeShadowAI, wantDecision: dynamicRoutingDecisionShadow},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			group, unhealthyID, healthyID := setupDynamicRoutingRecommendationTest(t, tc.mode)

			state := initDynamicRoutingModeState(group, "gpt-4o")
			if state.Decision != tc.wantDecision {
				t.Fatalf("Decision = %q, want %q", state.Decision, tc.wantDecision)
			}
			if state.AllowAdaptive {
				t.Fatal("AllowAdaptive = true, want false")
			}
			if len(state.Recommended) == 0 || state.Recommended[0].ChannelID != healthyID {
				t.Fatalf("recommended order = %#v, want healthy channel %d first", state.Recommended, healthyID)
			}

			iter := dynamicIterator(group, 0, "gpt-4o", state)
			if !iter.Next() {
				t.Fatal("iterator should have first candidate")
			}
			if got := iter.Item().ChannelID; got != unhealthyID {
				t.Fatalf("observation-mode iterator first channel = %d, want original mechanism channel %d", got, unhealthyID)
			}
		})
	}
}

func setupDynamicRoutingRecommendationTest(t *testing.T, mode string) (dbmodel.Group, int, int) {
	t.Helper()
	ctx := setupRelayTestDB(t)
	if err := op.SettingSetString(dbmodel.SettingKeyDynamicRoutingMode, mode); err != nil {
		t.Fatalf("SettingSetString(mode) error = %v", err)
	}
	if err := op.SettingSetString(dbmodel.SettingKeyDynamicRoutingHealthEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(health) error = %v", err)
	}

	unhealthy := &dbmodel.Channel{Name: "dynamic-recommend-unhealthy", Enabled: true, Model: "gpt-4o"}
	if err := op.ChannelCreate(unhealthy, ctx); err != nil {
		t.Fatalf("ChannelCreate(unhealthy) error = %v", err)
	}
	unhealthy, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: unhealthy.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{
		Enabled:       true,
		ChannelKey:    "unhealthy-key",
		SourceType:    dbmodel.ChannelKeySourceTypePublicFree,
		AllowedModels: "gpt-4o",
	}}}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(unhealthy) error = %v", err)
	}

	healthy := &dbmodel.Channel{Name: "dynamic-recommend-healthy", Enabled: true, Model: "gpt-4o"}
	if err := op.ChannelCreate(healthy, ctx); err != nil {
		t.Fatalf("ChannelCreate(healthy) error = %v", err)
	}
	healthy, err = op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: healthy.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{
		Enabled:       true,
		ChannelKey:    "healthy-key",
		SourceType:    dbmodel.ChannelKeySourceTypePublicFree,
		AllowedModels: "gpt-4o",
	}}}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(healthy) error = %v", err)
	}

	if err := op.StatsChannelUpdate(unhealthy.ID, dbmodel.StatsMetrics{RequestFailed: 10, WaitTime: 20000}); err != nil {
		t.Fatalf("StatsChannelUpdate(unhealthy) error = %v", err)
	}
	if err := op.StatsChannelUpdate(healthy.ID, dbmodel.StatsMetrics{RequestSuccess: 10, WaitTime: 500}); err != nil {
		t.Fatalf("StatsChannelUpdate(healthy) error = %v", err)
	}

	group := dbmodel.Group{
		ID:   1,
		Mode: dbmodel.GroupModeFailover,
		Items: []dbmodel.GroupItem{
			{ID: 1, ChannelID: unhealthy.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1},
			{ID: 2, ChannelID: healthy.ID, ModelName: "gpt-4o", Priority: 2, Weight: 1},
		},
	}
	return group, unhealthy.ID, healthy.ID
}

func TestAPIKeyAllowsModelTrimsWhitespace(t *testing.T) {
	t.Parallel()

	if !apiKeyAllowsModel("gpt-4o, gpt-4.1 ,  gpt-5", "gpt-4.1") {
		t.Fatal("apiKeyAllowsModel should accept trimmed supported model entries")
	}
	if apiKeyAllowsModel("gpt-4o, gpt-5", "gpt-4.1") {
		t.Fatal("apiKeyAllowsModel should reject unsupported models")
	}
}