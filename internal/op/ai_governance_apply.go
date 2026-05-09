package op

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm/clause"
	"gorm.io/gorm"
)

func GovernanceSessionApply(id int, ctx context.Context) (model.GovernanceSessionDetail, error) {
	var row model.GovernanceSession
	if err := db.GetDB().WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.GovernanceSessionDetail{}, fmt.Errorf("governance session not found")
		}
		return model.GovernanceSessionDetail{}, err
	}
	_, currentChecksum, err := governanceBuildSnapshot(ctx)
	if err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	if currentChecksum != row.SnapshotChecksum {
		if err := db.GetDB().WithContext(ctx).Model(&model.GovernanceSession{}).Where("id = ?", row.ID).Updates(map[string]any{"status": model.GovernanceSessionStatusStale, "updated_at": time.Now(), "error_message": errGovernanceSessionStale.Error()}).Error; err != nil {
			return model.GovernanceSessionDetail{}, err
		}
		return model.GovernanceSessionDetail{}, errGovernanceSessionStale
	}
	plan, preview, _, err := parseGovernanceArtifacts(row)
	if err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	if !preview.CanApply {
		return model.GovernanceSessionDetail{}, fmt.Errorf("governance preview is not applyable")
	}
	audit := model.GovernanceApplyAudit{Summary: "Applied governance plan", Items: make([]model.GovernanceApplyAuditItem, 0, len(plan.Mutations))}
	resultSummary := "Applied governance changes"
	var applyRun model.GovernanceApplyRun
	err = db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		applyRun = model.GovernanceApplyRun{SessionID: row.ID, Status: model.GovernanceApplyRunStatusRunning, ResultSummary: resultSummary, CreatedAt: now, UpdatedAt: now}
		if err := tx.Create(&applyRun).Error; err != nil {
			return err
		}
		if _, err := governanceCreateRollbackPoint(tx, row.ID, &applyRun.ID, row.SnapshotChecksum, ctx); err != nil {
			return err
		}
		if err := tx.Model(&model.GovernanceSession{}).Where("id = ?", row.ID).Updates(map[string]any{"status": model.GovernanceSessionStatusApplying, "current_stage": model.GovernanceStageApplyExecution, "last_apply_run_id": applyRun.ID, "updated_at": now}).Error; err != nil {
			return err
		}
		for _, mutation := range plan.Mutations {
			item := model.GovernanceApplyAuditItem{MutationType: mutation.Type, Summary: mutation.Summary, Status: "succeeded"}
			if err := governanceApplyMutation(tx, mutation, ctx); err != nil {
				item.Status = "failed"
				item.Message = err.Error()
				audit.Items = append(audit.Items, item)
				return err
			}
			audit.Items = append(audit.Items, item)
		}
		auditJSON, marshalErr := json.Marshal(audit)
		if marshalErr != nil {
			return marshalErr
		}
		if err := tx.Model(&model.GovernanceApplyRun{}).Where("id = ?", applyRun.ID).Updates(map[string]any{"status": model.GovernanceApplyRunStatusSucceeded, "result_summary": resultSummary, "audit_json": string(auditJSON), "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		appliedAt := time.Now()
		return tx.Model(&model.GovernanceSession{}).Where("id = ?", row.ID).Updates(map[string]any{"status": model.GovernanceSessionStatusApplied, "current_stage": model.GovernanceStageCompleted, "applied_at": appliedAt, "updated_at": appliedAt, "error_message": ""}).Error
	})
	if err != nil {
		auditJSON, _ := json.Marshal(audit)
		_ = db.GetDB().WithContext(ctx).Model(&model.GovernanceApplyRun{}).Where("id = ?", applyRun.ID).Updates(map[string]any{"status": model.GovernanceApplyRunStatusFailed, "result_summary": "Governance apply failed", "error_message": err.Error(), "audit_json": string(auditJSON), "updated_at": time.Now()})
		_ = db.GetDB().WithContext(ctx).Model(&model.GovernanceSession{}).Where("id = ?", row.ID).Updates(map[string]any{"status": model.GovernanceSessionStatusFailed, "current_stage": model.GovernanceStageApplyExecution, "error_message": err.Error(), "updated_at": time.Now()})
		return model.GovernanceSessionDetail{}, err
	}
	if err := InitCache(); err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	return GovernanceSessionGet(id, ctx)
}

func governanceApplyMutation(tx *gorm.DB, mutation model.GovernanceMutation, ctx context.Context) error {
	switch mutation.Type {
	case model.GovernanceMutationTypeGroupUpsert:
		if mutation.GroupUpsert == nil {
			return fmt.Errorf("group_upsert payload is required")
		}
		var row model.Group
		err := tx.WithContext(ctx).Where("name = ?", mutation.GroupUpsert.GroupName).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			create := model.Group{Name: mutation.GroupUpsert.GroupName, Mode: mutation.GroupUpsert.Mode}
			return tx.WithContext(ctx).Create(&create).Error
		}
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Model(&model.Group{}).Where("id = ?", row.ID).Updates(map[string]any{"mode": mutation.GroupUpsert.Mode}).Error
	case model.GovernanceMutationTypeGroupItemAttach:
		if mutation.GroupItemAttach == nil {
			return fmt.Errorf("group_item_attach payload is required")
		}
		groupID, err := governanceGroupIDByName(tx, mutation.GroupItemAttach.GroupName)
		if err != nil {
			return err
		}
		row := model.GroupItem{GroupID: groupID, ChannelID: mutation.GroupItemAttach.ChannelID, ModelName: mutation.GroupItemAttach.ModelName, Priority: mutation.GroupItemAttach.Priority, Weight: mutation.GroupItemAttach.Weight}
		return tx.WithContext(ctx).Where("group_id = ? AND channel_id = ? AND model_name = ?", row.GroupID, row.ChannelID, row.ModelName).Assign(row).FirstOrCreate(&row).Error
	case model.GovernanceMutationTypeGroupItemDetach:
		if mutation.GroupItemDetach == nil {
			return fmt.Errorf("group_item_detach payload is required")
		}
		groupID, err := governanceGroupIDByName(tx, mutation.GroupItemDetach.GroupName)
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Where("group_id = ? AND channel_id = ? AND model_name = ?", groupID, mutation.GroupItemDetach.ChannelID, mutation.GroupItemDetach.ModelName).Delete(&model.GroupItem{}).Error
	case model.GovernanceMutationTypeGroupItemReorder:
		if mutation.GroupItemReorder == nil {
			return fmt.Errorf("group_item_reorder payload is required")
		}
		groupID, err := governanceGroupIDByName(tx, mutation.GroupItemReorder.GroupName)
		if err != nil {
			return err
		}
		for _, item := range mutation.GroupItemReorder.Items {
			if err := tx.WithContext(ctx).Model(&model.GroupItem{}).Where("group_id = ? AND channel_id = ? AND model_name = ?", groupID, item.ChannelID, item.ModelName).Updates(map[string]any{"priority": item.Priority, "weight": item.Weight}).Error; err != nil {
				return err
			}
		}
		return nil
	case model.GovernanceMutationTypeRouteTargetOverrideUpsert:
		if mutation.RouteTargetUpsert == nil {
			return fmt.Errorf("route_target_override_upsert payload is required")
		}
		row := model.RouteTargetOverride{ChannelID: mutation.RouteTargetUpsert.ChannelID, ChannelKeyID: mutation.RouteTargetUpsert.ChannelKeyID, ModelName: mutation.RouteTargetUpsert.ModelName, BillingMode: mutation.RouteTargetUpsert.BillingMode, ProbePolicy: mutation.RouteTargetUpsert.ProbePolicy, ProbeIntervalSeconds: mutation.RouteTargetUpsert.ProbeIntervalSeconds, ProbeConcurrencyLimit: mutation.RouteTargetUpsert.ProbeConcurrencyLimit}
		if _, err := governanceRouteTargetOverrideUpsertTx(tx, row, ctx); err != nil {
			return err
		}
		return nil
	case model.GovernanceMutationTypeRouteTargetOverrideDelete:
		if mutation.RouteTargetDelete == nil {
			return fmt.Errorf("route_target_override_delete payload is required")
		}
		return tx.WithContext(ctx).Where("channel_id = ? AND channel_key_id = ? AND model_name = ?", mutation.RouteTargetDelete.ChannelID, mutation.RouteTargetDelete.ChannelKeyID, model.NormalizeRouteTargetModelName(mutation.RouteTargetDelete.ModelName)).Delete(&model.RouteTargetOverride{}).Error
	case model.GovernanceMutationTypeLLMPriceUpsert:
		if mutation.LLMPriceUpsert == nil {
			return fmt.Errorf("llm_price_upsert payload is required")
		}
		row := model.LLMInfo{
			Name:                  mutation.LLMPriceUpsert.Name,
			CanonicalName:         mutation.LLMPriceUpsert.CanonicalName,
			LLMPrice:              model.LLMPrice{Input: mutation.LLMPriceUpsert.Input, Output: mutation.LLMPriceUpsert.Output, CacheRead: mutation.LLMPriceUpsert.CacheRead, CacheWrite: mutation.LLMPriceUpsert.CacheWrite},
			OfficialLLMPrice:      model.OfficialLLMPrice{OfficialInput: mutation.LLMPriceUpsert.OfficialInput, OfficialOutput: mutation.LLMPriceUpsert.OfficialOutput, OfficialCacheRead: mutation.LLMPriceUpsert.OfficialCacheRead, OfficialCacheWrite: mutation.LLMPriceUpsert.OfficialCacheWrite},
			BillingMode:           mutation.LLMPriceUpsert.BillingMode,
			ProbePolicy:           mutation.LLMPriceUpsert.ProbePolicy,
			ProbeIntervalSeconds:  mutation.LLMPriceUpsert.ProbeIntervalSeconds,
			ProbeConcurrencyLimit: mutation.LLMPriceUpsert.ProbeConcurrencyLimit,
		}
		row, err := normalizeLLMPolicyFields(row)
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoUpdates: clause.AssignmentColumns([]string{"input", "output", "cache_read", "cache_write", "official_input", "official_output", "official_cache_read", "official_cache_write", "canonical_name", "billing_mode", "probe_policy", "probe_interval_seconds", "probe_concurrency_limit"})}).Create(&row).Error
	case model.GovernanceMutationTypeDynamicRoutingSettingSet:
		if mutation.DynamicRoutingSettingSet == nil {
			return fmt.Errorf("dynamic_routing_setting_set payload is required")
		}
		return tx.WithContext(ctx).Model(&model.Setting{}).Where("key = ?", mutation.DynamicRoutingSettingSet.Key).Update("value", mutation.DynamicRoutingSettingSet.Value).Error
	case model.GovernanceMutationTypeRuntimePolicySet:
		if mutation.RuntimePolicySet == nil {
			return fmt.Errorf("runtime_policy_set payload is required")
		}
		updates := map[model.SettingKey]string{
			model.SettingKeyAIRuntimeStrategy:              mutation.RuntimePolicySet.Strategy,
			model.SettingKeyAIRuntimeDispatchMode:          mutation.RuntimePolicySet.DispatchMode,
			model.SettingKeyAIRuntimeMaxParallelRuns:       fmt.Sprintf("%d", mutation.RuntimePolicySet.MaxParallelRuns),
			model.SettingKeyAIRuntimeDoubleReviewEnabled:   fmt.Sprintf("%t", mutation.RuntimePolicySet.DoubleReviewEnabled),
			model.SettingKeyAIRuntimeFallbackDeterministic: fmt.Sprintf("%t", mutation.RuntimePolicySet.FallbackToDeterministic),
		}
		for key, value := range updates {
			if err := tx.WithContext(ctx).Model(&model.Setting{}).Where("key = ?", key).Update("value", value).Error; err != nil {
				return err
			}
		}
		return nil
	case model.GovernanceMutationTypeStrategyProfileActivate:
		if mutation.StrategyProfileActivate == nil {
			return fmt.Errorf("strategy_profile_activate payload is required")
		}
		_, err := governanceStrategyProfileActivateTx(tx, mutation.StrategyProfileActivate.StrategyProfileID, ctx)
		return err
	default:
		return fmt.Errorf("unsupported governance mutation type: %s", mutation.Type)
	}
}

func governanceRouteTargetOverrideUpsertTx(tx *gorm.DB, row model.RouteTargetOverride, ctx context.Context) (model.RouteTargetOverride, error) {
	normalized, err := normalizeRouteTargetOverrideStrict(row)
	if err != nil {
		return model.RouteTargetOverride{}, err
	}
	if err := validateRouteTargetOverrideTarget(normalized.ChannelID, normalized.ChannelKeyID, normalized.ModelName); err != nil {
		return model.RouteTargetOverride{}, err
	}
	if err := tx.WithContext(ctx).Clauses(clause.OnConflict{
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
	if err := tx.WithContext(ctx).
		Where("channel_id = ? AND channel_key_id = ? AND model_name = ?", normalized.ChannelID, normalized.ChannelKeyID, normalized.ModelName).
		First(&stored).Error; err != nil {
		return model.RouteTargetOverride{}, err
	}
	stored, _ = normalizeRouteTargetOverrideLenient(stored)
	return stored, nil
}

func governanceStrategyProfileActivateTx(tx *gorm.DB, id int, ctx context.Context) (model.StrategyProfileSummary, error) {
	if id <= 0 {
		return model.StrategyProfileSummary{}, fmt.Errorf("invalid strategy profile id")
	}
	now := time.Now()
	if err := tx.WithContext(ctx).Model(&model.StrategyProfile{}).Where("status = ?", model.StrategyProfileStatusActive).Updates(map[string]any{"status": model.StrategyProfileStatusReady, "updated_at": now}).Error; err != nil {
		return model.StrategyProfileSummary{}, err
	}
	result := tx.WithContext(ctx).Model(&model.StrategyProfile{}).Where("id = ?", id).Updates(map[string]any{"status": model.StrategyProfileStatusActive, "activated_at": now, "updated_at": now})
	if result.Error != nil {
		return model.StrategyProfileSummary{}, result.Error
	}
	if result.RowsAffected == 0 {
		return model.StrategyProfileSummary{}, fmt.Errorf("invalid strategy profile id")
	}
	if err := tx.WithContext(ctx).Model(&model.Setting{}).Where("key = ?", model.SettingKeyActiveStrategyProfileID).Update("value", fmt.Sprintf("%d", id)).Error; err != nil {
		return model.StrategyProfileSummary{}, err
	}
	var row model.StrategyProfile
	if err := tx.WithContext(ctx).First(&row, id).Error; err != nil {
		return model.StrategyProfileSummary{}, err
	}
	return strategyProfileSummaryFromRecord(row, id), nil
}

func governanceGroupIDByName(tx *gorm.DB, name string) (int, error) {
	var row model.Group
	if err := tx.Where("name = ?", name).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("group %s not found", name)
		}
		return 0, err
	}
	return row.ID, nil
}

func ImportLegacyActiveAIProfileAsStrategyProfile(ctx context.Context) error {
	activeProfileID, _ := SettingGetInt(model.SettingKeyActiveAIProfileID)
	if activeProfileID <= 0 {
		return nil
	}
	var existing int64
	if err := db.GetDB().WithContext(ctx).Model(&model.StrategyProfile{}).Where("legacy_ai_profile_id = ?", activeProfileID).Count(&existing).Error; err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}
	profile, err := AIProfileGet(activeProfileID, ctx)
	if err != nil {
		return nil
	}
	name := strings.TrimSpace(profile.Name)
	if name == "" {
		name = fmt.Sprintf("Imported AI Profile %d", profile.ID)
	}
	now := time.Now()
	row := model.StrategyProfile{Name: name, Summary: strings.TrimSpace(profile.Explanation), Status: model.StrategyProfileStatusImported, LegacyAIProfileID: &profile.ID, CreatedAt: now, UpdatedAt: now, ActivatedAt: &now}
	if err := db.GetDB().WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	if err := SettingSetInt(model.SettingKeyActiveStrategyProfileID, row.ID); err != nil {
		return err
	}
	return db.GetDB().WithContext(ctx).Model(&model.StrategyProfile{}).Where("id = ?", row.ID).Update("status", model.StrategyProfileStatusActive).Error
}
