package migrate

import (
	"fmt"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 14,
		Up:      addAIGovernanceV3Tables,
	})
}

func addAIGovernanceV3Tables(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if err := db.AutoMigrate(&model.GovernanceRollbackPoint{}); err != nil {
		return fmt.Errorf("auto migrate ai governance v3 tables: %w", err)
	}
	for _, setting := range model.DefaultSettings() {
		switch setting.Key {
		case model.SettingKeyAIRuntimeStrategy,
			model.SettingKeyAIRuntimeDispatchMode,
			model.SettingKeyAIRuntimeMaxParallelRuns,
			model.SettingKeyAIRuntimeDoubleReviewEnabled,
			model.SettingKeyAIRuntimeFallbackDeterministic:
			if err := db.Exec("INSERT INTO settings (key, value) SELECT ?, ? WHERE NOT EXISTS (SELECT 1 FROM settings WHERE key = ?)", setting.Key, setting.Value, setting.Key).Error; err != nil {
				return fmt.Errorf("failed to ensure setting %s: %w", setting.Key, err)
			}
		}
	}
	return nil
}
