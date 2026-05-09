package op

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm"
)

type governanceRollbackSnapshot struct {
	Groups               []model.Group               `json:"groups"`
	GroupItems           []model.GroupItem           `json:"group_items"`
	RouteTargetOverrides []model.RouteTargetOverride `json:"route_target_overrides"`
	Models               []model.LLMInfo             `json:"models"`
	StrategyProfiles     []model.StrategyProfile     `json:"strategy_profiles"`
	Settings             []model.Setting             `json:"settings"`
}

func governanceCreateRollbackPoint(tx *gorm.DB, sessionID int, applyRunID *int, checksum string, ctx context.Context) (model.GovernanceRollbackPoint, error) {
	snapshot, err := governanceBuildRollbackSnapshot(tx, ctx)
	if err != nil {
		return model.GovernanceRollbackPoint{}, err
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return model.GovernanceRollbackPoint{}, err
	}
	now := time.Now()
	row := model.GovernanceRollbackPoint{
		SessionID:        sessionID,
		ApplyRunID:       applyRunID,
		SnapshotChecksum: checksum,
		SnapshotJSON:     string(payload),
		Summary:          "apply-preflight snapshot",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := tx.WithContext(ctx).Create(&row).Error; err != nil {
		return model.GovernanceRollbackPoint{}, err
	}
	return row, nil
}

func GovernanceRollbackPointList(sessionID int, ctx context.Context) ([]model.GovernanceRollbackPointView, error) {
	rows := make([]model.GovernanceRollbackPoint, 0)
	query := db.GetDB().WithContext(ctx).Order("created_at desc, id desc")
	if sessionID > 0 {
		query = query.Where("session_id = ?", sessionID)
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]model.GovernanceRollbackPointView, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.GovernanceRollbackPointView{
			ID:               row.ID,
			SessionID:        row.SessionID,
			ApplyRunID:       row.ApplyRunID,
			SnapshotChecksum: row.SnapshotChecksum,
			Summary:          row.Summary,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
		})
	}
	return items, nil
}

func GovernanceSessionRollback(sessionID int, rollbackPointID int, ctx context.Context) (model.GovernanceSessionDetail, error) {
	if sessionID <= 0 {
		return model.GovernanceSessionDetail{}, fmt.Errorf("invalid governance session id")
	}
	var point model.GovernanceRollbackPoint
	query := db.GetDB().WithContext(ctx).Where("session_id = ?", sessionID)
	if rollbackPointID > 0 {
		query = query.Where("id = ?", rollbackPointID)
	} else {
		query = query.Order("created_at desc, id desc")
	}
	if err := query.First(&point).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.GovernanceSessionDetail{}, fmt.Errorf("rollback point not found")
		}
		return model.GovernanceSessionDetail{}, err
	}
	var snapshot governanceRollbackSnapshot
	if err := json.Unmarshal([]byte(point.SnapshotJSON), &snapshot); err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	err := db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&model.GroupItem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("1 = 1").Delete(&model.Group{}).Error; err != nil {
			return err
		}
		if err := tx.Where("1 = 1").Delete(&model.RouteTargetOverride{}).Error; err != nil {
			return err
		}
		if err := tx.Where("1 = 1").Delete(&model.LLMInfo{}).Error; err != nil {
			return err
		}
		if err := tx.Where("1 = 1").Delete(&model.StrategyProfile{}).Error; err != nil {
			return err
		}
		if len(snapshot.StrategyProfiles) > 0 {
			if err := tx.Create(&snapshot.StrategyProfiles).Error; err != nil {
				return err
			}
		}
		for _, setting := range snapshot.Settings {
			if err := tx.Model(&model.Setting{}).Where("key = ?", setting.Key).Update("value", setting.Value).Error; err != nil {
				return err
			}
		}
		if len(snapshot.Groups) > 0 {
			if err := tx.Create(&snapshot.Groups).Error; err != nil {
				return err
			}
		}
		if len(snapshot.GroupItems) > 0 {
			if err := tx.Create(&snapshot.GroupItems).Error; err != nil {
				return err
			}
		}
		if len(snapshot.RouteTargetOverrides) > 0 {
			if err := tx.Create(&snapshot.RouteTargetOverrides).Error; err != nil {
				return err
			}
		}
		if len(snapshot.Models) > 0 {
			if err := tx.Create(&snapshot.Models).Error; err != nil {
				return err
			}
		}
		now := time.Now()
		return tx.Model(&model.GovernanceSession{}).Where("id = ?", sessionID).Updates(map[string]any{
			"status":        model.GovernanceSessionStatusApplied,
			"current_stage": model.GovernanceStageCompleted,
			"updated_at":    now,
			"error_message": "",
		}).Error
	})
	if err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	if err := InitCache(); err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	return GovernanceSessionGet(sessionID, ctx)
}

func governanceBuildRollbackSnapshot(tx *gorm.DB, ctx context.Context) (governanceRollbackSnapshot, error) {
	groups := make([]model.Group, 0)
	groupItems := make([]model.GroupItem, 0)
	overrides := make([]model.RouteTargetOverride, 0)
	models := make([]model.LLMInfo, 0)
	settings := make([]model.Setting, 0)
	if err := tx.WithContext(ctx).Find(&groups).Error; err != nil {
		return governanceRollbackSnapshot{}, err
	}
	if err := tx.WithContext(ctx).Find(&groupItems).Error; err != nil {
		return governanceRollbackSnapshot{}, err
	}
	if err := tx.WithContext(ctx).Find(&overrides).Error; err != nil {
		return governanceRollbackSnapshot{}, err
	}
	if err := tx.WithContext(ctx).Find(&models).Error; err != nil {
		return governanceRollbackSnapshot{}, err
	}
	strategyProfiles := make([]model.StrategyProfile, 0)
	if err := tx.WithContext(ctx).Find(&strategyProfiles).Error; err != nil {
		return governanceRollbackSnapshot{}, err
	}
	affectedSettingKeys := []model.SettingKey{
		model.SettingKeyDynamicRoutingMode,
		model.SettingKeyDynamicRoutingHealthEnabled,
		model.SettingKeyDynamicRoutingLearningEnabled,
		model.SettingKeyRaceGlobalBudget,
		model.SettingKeyRaceGroupBudget,
		model.SettingKeyRaceChannelBudget,
		model.SettingKeyRaceKeyBudget,
		model.SettingKeyRaceProbeBudget,
		model.SettingKeyActiveStrategyProfileID,
		model.SettingKeyAIRuntimeStrategy,
		model.SettingKeyAIRuntimeDispatchMode,
		model.SettingKeyAIRuntimeMaxParallelRuns,
		model.SettingKeyAIRuntimeDoubleReviewEnabled,
		model.SettingKeyAIRuntimeFallbackDeterministic,
	}
	if err := tx.WithContext(ctx).Where("key IN ?", affectedSettingKeys).Find(&settings).Error; err != nil {
		return governanceRollbackSnapshot{}, err
	}
	return governanceRollbackSnapshot{
		Groups:               groups,
		GroupItems:           groupItems,
		RouteTargetOverrides: overrides,
		Models:               models,
		StrategyProfiles:     strategyProfiles,
		Settings:             settings,
	}, nil
}
