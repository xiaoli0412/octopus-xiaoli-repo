package op

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestGovernancePlanDetachesStaleItemFromOriginalGroup(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIGovernanceManagedGroupName, "AI Governance Managed"); err != nil {
		t.Fatalf("SettingSetString(managed_group) error = %v", err)
	}

	channel := createConfiguredTestChannel(t, ctx, "governance-stale-channel", "gpt-4o", "")
	staleGroup := &model.Group{Name: "legacy-group", Mode: model.GroupModeRoundRobin}
	if err := GroupCreate(staleGroup, ctx); err != nil {
		t.Fatalf("GroupCreate(stale) error = %v", err)
	}
	staleItem := model.GroupItem{GroupID: staleGroup.ID, ChannelID: channel.ID, ModelName: "removed-model", Priority: 1, Weight: 1}
	if err := db.GetDB().WithContext(ctx).Create(&staleItem).Error; err != nil {
		t.Fatalf("seed stale group item error = %v", err)
	}
	if err := groupRefreshCacheByID(staleGroup.ID, ctx); err != nil {
		t.Fatalf("groupRefreshCacheByID() error = %v", err)
	}

	snapshot, _, err := governanceBuildSnapshot(ctx)
	if err != nil {
		t.Fatalf("governanceBuildSnapshot() error = %v", err)
	}
	plan := governancePlanFromSnapshot("cleanup stale routes", model.GovernanceExpertPresetBalanced, snapshot)

	for _, mutation := range plan.Mutations {
		if mutation.Type != model.GovernanceMutationTypeGroupItemDetach || mutation.GroupItemDetach == nil {
			continue
		}
		if mutation.GroupItemDetach.ModelName == "removed-model" {
			if mutation.GroupItemDetach.GroupName != staleGroup.Name {
				t.Fatalf("detach group = %q, want %q", mutation.GroupItemDetach.GroupName, staleGroup.Name)
			}
			return
		}
	}
	t.Fatal("expected stale group item detach mutation")
}

func TestGovernanceSessionApplyRollsBackTransactionScopedMutations(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIGovernanceManagedGroupName, "AI Governance Managed"); err != nil {
		t.Fatalf("SettingSetString(managed_group) error = %v", err)
	}

	channel := &model.Channel{Name: "governance-apply-channel", Enabled: true, Model: "gpt-4o"}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	key := model.ChannelKey{ChannelID: channel.ID, Enabled: true, ChannelKey: "governance-apply-key", AllowedModels: "gpt-4o"}
	if err := db.GetDB().WithContext(ctx).Create(&key).Error; err != nil {
		t.Fatalf("create key error = %v", err)
	}
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() after key seed error = %v", err)
	}
	if err := LLMCreate(model.LLMInfo{Name: "gpt-4o", BillingMode: model.BillingModePerToken, ProbePolicy: model.ProbePolicySequential, ProbeConcurrencyLimit: 1}, ctx); err != nil {
		t.Fatalf("LLMCreate() error = %v", err)
	}

	profile := model.StrategyProfile{Name: "pre-existing-active", Summary: "baseline", Status: model.StrategyProfileStatusActive}
	if err := db.GetDB().WithContext(ctx).Create(&profile).Error; err != nil {
		t.Fatalf("create active profile error = %v", err)
	}
	if err := SettingSetInt(model.SettingKeyActiveStrategyProfileID, profile.ID); err != nil {
		t.Fatalf("SettingSetInt(active profile) error = %v", err)
	}

	missingProfileID := profile.ID + 999
	plan := model.GovernancePlanView{
		Mutations: []model.GovernanceMutation{
			{
				Type:    model.GovernanceMutationTypeRouteTargetOverrideUpsert,
				Summary: "create override",
				RouteTargetUpsert: &model.GovernanceRouteTargetOverrideMutation{
					ChannelID:             channel.ID,
					ChannelKeyID:          key.ID,
					ModelName:             "gpt-4o",
					BillingMode:           model.BillingModePerRequest,
					ProbePolicy:           model.ProbePolicyConcurrent,
					ProbeIntervalSeconds:  1800,
					ProbeConcurrencyLimit: 2,
				},
			},
			{
				Type:                    model.GovernanceMutationTypeStrategyProfileActivate,
				Summary:                 "activate missing profile",
				StrategyProfileActivate: &model.GovernanceStrategyProfileActivateMutation{StrategyProfileID: missingProfileID},
			},
		},
	}
	preview := governancePreviewFromPlan(plan)
	snapshot, checksum, err := governanceBuildSnapshot(ctx)
	if err != nil {
		t.Fatalf("governanceBuildSnapshot() error = %v", err)
	}
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot error = %v", err)
	}
	planJSON, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal plan error = %v", err)
	}
	previewJSON, err := json.Marshal(preview)
	if err != nil {
		t.Fatalf("marshal preview error = %v", err)
	}
	row := model.GovernanceSession{
		Goal:             "test rollback",
		Scope:            model.GovernanceScopeRoutingGrouping,
		ExpertPresetID:   model.GovernanceExpertPresetBalanced,
		Status:           model.GovernanceSessionStatusReady,
		CurrentStage:     model.GovernanceStageCompleted,
		OperatorSummary:  "rollback test",
		RiskSummary:      "rollback test",
		Confidence:       0.5,
		SnapshotChecksum: checksum,
		SnapshotJSON:     string(snapshotJSON),
		PlanJSON:         string(planJSON),
		PreviewJSON:      string(previewJSON),
		CreatedAt:        nowForTest(),
		UpdatedAt:        nowForTest(),
	}
	if err := db.GetDB().WithContext(ctx).Create(&row).Error; err != nil {
		t.Fatalf("create governance session error = %v", err)
	}

	if _, err := GovernanceSessionApply(row.ID, ctx); err == nil {
		t.Fatal("GovernanceSessionApply() expected error")
	}

	if _, ok := RouteTargetOverrideGet(channel.ID, key.ID, "gpt-4o"); ok {
		t.Fatal("route target override persisted despite failed governance apply")
	}
	var storedProfile model.StrategyProfile
	if err := db.GetDB().WithContext(ctx).First(&storedProfile, profile.ID).Error; err != nil {
		t.Fatalf("reload profile error = %v", err)
	}
	if storedProfile.Status != model.StrategyProfileStatusActive {
		t.Fatalf("active profile status = %q, want %q", storedProfile.Status, model.StrategyProfileStatusActive)
	}
	activeID, err := SettingGetInt(model.SettingKeyActiveStrategyProfileID)
	if err != nil {
		t.Fatalf("SettingGetInt(active profile) error = %v", err)
	}
	if activeID != profile.ID {
		t.Fatalf("active strategy profile id = %d, want %d", activeID, profile.ID)
	}
}

func TestGovernanceSessionRollbackRestoresStrategyProfilesAndActiveSetting(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := SettingSetString(model.SettingKeyAIAutomationEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(enabled) error = %v", err)
	}
	if err := SettingSetString(model.SettingKeyAIGovernanceManagedGroupName, "AI Governance Managed"); err != nil {
		t.Fatalf("SettingSetString(managed_group) error = %v", err)
	}

	original := model.StrategyProfile{Name: "original-active", Summary: "baseline", Status: model.StrategyProfileStatusActive}
	if err := db.GetDB().WithContext(ctx).Create(&original).Error; err != nil {
		t.Fatalf("create original profile error = %v", err)
	}
	target := model.StrategyProfile{Name: "target-ready", Summary: "candidate", Status: model.StrategyProfileStatusReady}
	if err := db.GetDB().WithContext(ctx).Create(&target).Error; err != nil {
		t.Fatalf("create target profile error = %v", err)
	}
	if err := SettingSetInt(model.SettingKeyActiveStrategyProfileID, original.ID); err != nil {
		t.Fatalf("SettingSetInt(active profile) error = %v", err)
	}

	snapshot, checksum, err := governanceBuildSnapshot(ctx)
	if err != nil {
		t.Fatalf("governanceBuildSnapshot() error = %v", err)
	}
	plan := model.GovernancePlanView{Mutations: []model.GovernanceMutation{{
		Type:                    model.GovernanceMutationTypeStrategyProfileActivate,
		Summary:                 "activate target profile",
		StrategyProfileActivate: &model.GovernanceStrategyProfileActivateMutation{StrategyProfileID: target.ID},
	}}}
	preview := governancePreviewFromPlan(plan)
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot error = %v", err)
	}
	planJSON, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal plan error = %v", err)
	}
	previewJSON, err := json.Marshal(preview)
	if err != nil {
		t.Fatalf("marshal preview error = %v", err)
	}
	row := model.GovernanceSession{
		Goal:             "test rollback profiles",
		Scope:            model.GovernanceScopeRoutingGrouping,
		ExpertPresetID:   model.GovernanceExpertPresetBalanced,
		Status:           model.GovernanceSessionStatusReady,
		CurrentStage:     model.GovernanceStageCompleted,
		OperatorSummary:  "rollback profile test",
		RiskSummary:      "rollback profile test",
		Confidence:       0.9,
		SnapshotChecksum: checksum,
		SnapshotJSON:     string(snapshotJSON),
		PlanJSON:         string(planJSON),
		PreviewJSON:      string(previewJSON),
		CreatedAt:        nowForTest(),
		UpdatedAt:        nowForTest(),
	}
	if err := db.GetDB().WithContext(ctx).Create(&row).Error; err != nil {
		t.Fatalf("create governance session error = %v", err)
	}

	applyResult, err := GovernanceSessionApply(row.ID, ctx)
	if err != nil {
		t.Fatalf("GovernanceSessionApply() error = %v", err)
	}
	if len(applyResult.RollbackPoints) == 0 {
		t.Fatal("expected rollback point to be created")
	}

	var activeAfterApply model.StrategyProfile
	if err := db.GetDB().WithContext(ctx).First(&activeAfterApply, target.ID).Error; err != nil {
		t.Fatalf("reload target profile after apply error = %v", err)
	}
	if activeAfterApply.Status != model.StrategyProfileStatusActive {
		t.Fatalf("target profile status after apply = %q, want %q", activeAfterApply.Status, model.StrategyProfileStatusActive)
	}
	var originalAfterApply model.StrategyProfile
	if err := db.GetDB().WithContext(ctx).First(&originalAfterApply, original.ID).Error; err != nil {
		t.Fatalf("reload original profile after apply error = %v", err)
	}
	if originalAfterApply.Status != model.StrategyProfileStatusReady {
		t.Fatalf("original profile status after apply = %q, want %q", originalAfterApply.Status, model.StrategyProfileStatusReady)
	}
	activeIDAfterApply, err := SettingGetInt(model.SettingKeyActiveStrategyProfileID)
	if err != nil {
		t.Fatalf("SettingGetInt(active profile after apply) error = %v", err)
	}
	if activeIDAfterApply != target.ID {
		t.Fatalf("active strategy profile id after apply = %d, want %d", activeIDAfterApply, target.ID)
	}

	rollbackPointID := applyResult.RollbackPoints[0].ID
	rollbackResult, err := GovernanceSessionRollback(row.ID, rollbackPointID, ctx)
	if err != nil {
		t.Fatalf("GovernanceSessionRollback() error = %v", err)
	}
	if rollbackResult.ID != row.ID {
		t.Fatalf("rollback result session id = %d, want %d", rollbackResult.ID, row.ID)
	}

	var originalAfterRollback model.StrategyProfile
	if err := db.GetDB().WithContext(ctx).First(&originalAfterRollback, original.ID).Error; err != nil {
		t.Fatalf("reload original profile after rollback error = %v", err)
	}
	if originalAfterRollback.Status != model.StrategyProfileStatusActive {
		t.Fatalf("original profile status after rollback = %q, want %q", originalAfterRollback.Status, model.StrategyProfileStatusActive)
	}
	var targetAfterRollback model.StrategyProfile
	if err := db.GetDB().WithContext(ctx).First(&targetAfterRollback, target.ID).Error; err != nil {
		t.Fatalf("reload target profile after rollback error = %v", err)
	}
	if targetAfterRollback.Status != model.StrategyProfileStatusReady {
		t.Fatalf("target profile status after rollback = %q, want %q", targetAfterRollback.Status, model.StrategyProfileStatusReady)
	}
	activeIDAfterRollback, err := SettingGetInt(model.SettingKeyActiveStrategyProfileID)
	if err != nil {
		t.Fatalf("SettingGetInt(active profile after rollback) error = %v", err)
	}
	if activeIDAfterRollback != original.ID {
		t.Fatalf("active strategy profile id after rollback = %d, want %d", activeIDAfterRollback, original.ID)
	}
}

func nowForTest() time.Time {
	return time.Unix(1, 0).UTC()
}
