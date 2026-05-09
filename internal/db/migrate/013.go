package migrate

import (
	"fmt"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 13,
		Up:      addAIGovernanceV2Tables,
	})
}

func addAIGovernanceV2Tables(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if err := db.AutoMigrate(
		&model.GovernanceSession{},
		&model.GovernanceApplyRun{},
		&model.StrategyProfile{},
	); err != nil {
		return fmt.Errorf("auto migrate ai governance v2 tables: %w", err)
	}
	for _, setting := range model.DefaultSettings() {
		switch setting.Key {
		case model.SettingKeyAIGovernanceManagedGroupName,
			model.SettingKeyActiveStrategyProfileID:
			if err := db.Exec("INSERT INTO settings (key, value) SELECT ?, ? WHERE NOT EXISTS (SELECT 1 FROM settings WHERE key = ?)", setting.Key, setting.Value, setting.Key).Error; err != nil {
				return fmt.Errorf("failed to ensure setting %s: %w", setting.Key, err)
			}
		}
	}
	return nil
}
