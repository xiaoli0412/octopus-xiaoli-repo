package op

import (
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestResolveRouteTargetPolicyPrecedence(t *testing.T) {
	ctx := setupOpTestDB(t)

	channel := &model.Channel{Name: "route-target-policy-channel", Enabled: true, Model: "gpt-4o"}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	updated, err := ChannelUpdate(&model.ChannelUpdateRequest{ID: channel.ID, KeysToAdd: []model.ChannelKeyAddRequest{{
		Enabled:       true,
		ChannelKey:    "route-target-key",
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

	policy := ResolveRouteTargetPolicy(updated, key, "gpt-4o")
	if policy.BillingMode != model.BillingModeUnknown {
		t.Fatalf("inheritance billing_mode = %q, want unknown", policy.BillingMode)
	}
	if policy.BillingModeBasis != model.RouteTargetPolicyBasisChannelKeyInheritance {
		t.Fatalf("inheritance billing basis = %q, want %q", policy.BillingModeBasis, model.RouteTargetPolicyBasisChannelKeyInheritance)
	}
	if policy.SourceType != model.ChannelKeySourceTypePublicFree {
		t.Fatalf("source_type = %q, want %q", policy.SourceType, model.ChannelKeySourceTypePublicFree)
	}

	if err := LLMCreate(model.LLMInfo{
		Name:                  "gpt-4o",
		BillingMode:           model.BillingModePerToken,
		ProbePolicy:           model.ProbePolicySequential,
		ProbeIntervalSeconds:  7200,
		ProbeConcurrencyLimit: 3,
	}, ctx); err != nil {
		t.Fatalf("LLMCreate() error = %v", err)
	}

	policy = ResolveRouteTargetPolicy(updated, key, "gpt-4o")
	if policy.BillingMode != model.BillingModePerToken || policy.BillingModeBasis != model.RouteTargetPolicyBasisModelDefault {
		t.Fatalf("model-default billing = (%q,%q), want per_token/model_default", policy.BillingMode, policy.BillingModeBasis)
	}
	if policy.ProbePolicy != model.ProbePolicySequential || policy.ProbePolicyBasis != model.RouteTargetPolicyBasisModelDefault {
		t.Fatalf("model-default probe = (%q,%q), want sequential/model_default", policy.ProbePolicy, policy.ProbePolicyBasis)
	}
	if policy.ProbeConcurrencyLimit != 3 || policy.ProbeConcurrencyBasis != model.RouteTargetPolicyBasisModelDefault {
		t.Fatalf("model-default concurrency = (%d,%q), want 3/model_default", policy.ProbeConcurrencyLimit, policy.ProbeConcurrencyBasis)
	}

	if _, err := RouteTargetOverrideUpsert(model.RouteTargetOverride{
		ChannelID:             updated.ID,
		ChannelKeyID:          key.ID,
		ModelName:             "gpt-4o",
		BillingMode:           model.BillingModePerRequest,
		ProbePolicy:           model.ProbePolicyConcurrent,
		ProbeIntervalSeconds:  1800,
		ProbeConcurrencyLimit: 2,
	}, ctx); err != nil {
		t.Fatalf("RouteTargetOverrideUpsert() error = %v", err)
	}

	policy = ResolveRouteTargetPolicy(updated, key, "gpt-4o")
	if policy.BillingMode != model.BillingModePerRequest || policy.BillingModeBasis != model.RouteTargetPolicyBasisExplicitOverride {
		t.Fatalf("override billing = (%q,%q), want per_request/route_target_override", policy.BillingMode, policy.BillingModeBasis)
	}
	if policy.ProbePolicy != model.ProbePolicyConcurrent || policy.ProbePolicyBasis != model.RouteTargetPolicyBasisExplicitOverride {
		t.Fatalf("override probe = (%q,%q), want concurrent/route_target_override", policy.ProbePolicy, policy.ProbePolicyBasis)
	}
	if policy.ProbeIntervalSeconds != 1800 || policy.ProbeIntervalBasis != model.RouteTargetPolicyBasisExplicitOverride {
		t.Fatalf("override interval = (%d,%q), want 1800/route_target_override", policy.ProbeIntervalSeconds, policy.ProbeIntervalBasis)
	}
	if policy.ProbeConcurrencyLimit != 2 || policy.ProbeConcurrencyBasis != model.RouteTargetPolicyBasisExplicitOverride {
		t.Fatalf("override concurrency = (%d,%q), want 2/route_target_override", policy.ProbeConcurrencyLimit, policy.ProbeConcurrencyBasis)
	}
	if got := policy.PolicyBasisSummary(); got != "route_target_override+channel_key_inheritance" && got != "route_target_override+model_default+channel_key_inheritance" {
		t.Fatalf("PolicyBasisSummary() = %q, want override-led summary", got)
	}
}

func TestRouteTargetOverrideDeleteByModels(t *testing.T) {
	ctx := setupOpTestDB(t)

	channel := &model.Channel{Name: "route-target-delete-channel", Enabled: true, Model: "gpt-4o,claude-3-5-sonnet"}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	updated, err := ChannelUpdate(&model.ChannelUpdateRequest{ID: channel.ID, KeysToAdd: []model.ChannelKeyAddRequest{{
		Enabled:       true,
		ChannelKey:    "route-target-delete-key",
		SourceType:    "public/free",
		AllowedModels: "gpt-4o,claude-3-5-sonnet",
	}}}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(keys) error = %v", err)
	}
	key := updated.Keys[0]

	if _, err := RouteTargetOverrideUpsert(model.RouteTargetOverride{ChannelID: updated.ID, ChannelKeyID: key.ID, ModelName: "gpt-4o"}, ctx); err != nil {
		t.Fatalf("RouteTargetOverrideUpsert(gpt-4o) error = %v", err)
	}
	if _, err := RouteTargetOverrideUpsert(model.RouteTargetOverride{ChannelID: updated.ID, ChannelKeyID: key.ID, ModelName: "claude-3-5-sonnet"}, ctx); err != nil {
		t.Fatalf("RouteTargetOverrideUpsert(claude) error = %v", err)
	}
	if err := RouteTargetOverrideDeleteByModels([]string{"gpt-4o"}, ctx); err != nil {
		t.Fatalf("RouteTargetOverrideDeleteByModels() error = %v", err)
	}
	if _, ok := RouteTargetOverrideGet(updated.ID, key.ID, "gpt-4o"); ok {
		t.Fatalf("override for gpt-4o should be deleted")
	}
	if _, ok := RouteTargetOverrideGet(updated.ID, key.ID, "claude-3-5-sonnet"); !ok {
		t.Fatalf("override for claude should remain")
	}
}

func TestRouteTargetOverrideUpsertRejectsChannelKeyFromAnotherChannel(t *testing.T) {
	ctx := setupOpTestDB(t)

	channelA := createConfiguredTestChannel(t, ctx, "route-target-ownership-a", "gpt-4o", "")
	channelB := createConfiguredTestChannel(t, ctx, "route-target-ownership-b", "gpt-4o", "")

	if _, err := RouteTargetOverrideUpsert(model.RouteTargetOverride{
		ChannelID:    channelA.ID,
		ChannelKeyID: channelB.Keys[0].ID,
		ModelName:    "gpt-4o",
	}, ctx); err == nil {
		t.Fatal("RouteTargetOverrideUpsert() expected channel/key ownership error")
	} else if err.Error() != "invalid channel key id for channel" {
		t.Fatalf("RouteTargetOverrideUpsert() error = %v, want invalid channel key id for channel", err)
	}

	rows, err := RouteTargetOverrideList(ctx)
	if err != nil {
		t.Fatalf("RouteTargetOverrideList() error = %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("route target overrides = %#v, want no persisted rows", rows)
	}
}

func TestRouteTargetOverrideUpsertRejectsModelOutsideKeyScope(t *testing.T) {
	ctx := setupOpTestDB(t)

	channel := &model.Channel{Name: "route-target-model-scope", Enabled: true, Model: "gpt-4o,claude-3-5-sonnet"}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	updated, err := ChannelUpdate(&model.ChannelUpdateRequest{ID: channel.ID, KeysToAdd: []model.ChannelKeyAddRequest{{
		Enabled:       true,
		ChannelKey:    "route-target-model-scope-key",
		SourceType:    "paid/metered",
		AllowedModels: "gpt-4o",
	}}}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(keys) error = %v", err)
	}

	if _, err := RouteTargetOverrideUpsert(model.RouteTargetOverride{
		ChannelID:    updated.ID,
		ChannelKeyID: updated.Keys[0].ID,
		ModelName:    "claude-3-5-sonnet",
	}, ctx); err == nil {
		t.Fatal("RouteTargetOverrideUpsert() expected key model-scope validation error")
	} else if err.Error() != "invalid model for channel key" {
		t.Fatalf("RouteTargetOverrideUpsert() error = %v, want invalid model for channel key", err)
	}

	rows, err := RouteTargetOverrideList(ctx)
	if err != nil {
		t.Fatalf("RouteTargetOverrideList() error = %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("route target overrides = %#v, want no persisted rows", rows)
	}
}
