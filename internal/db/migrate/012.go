package migrate

import (
	"fmt"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 12,
		Up:      addAIAutomationCenterTables,
	})
}

// 012:
// - AI automation center task/profile/prompt tables
// - dynamic routing local learning state table
// - non-destructive settings for AI Profile dual-track switching
func addAIAutomationCenterTables(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if err := db.AutoMigrate(
		&model.AITask{},
		&model.AITaskStep{},
		&model.AIPromptTemplate{},
		&model.AIProfile{},
		&model.AIProfileVersion{},
		&model.DynamicRouteLearningState{},
	); err != nil {
		return fmt.Errorf("auto migrate ai automation center tables: %w", err)
	}
	for _, setting := range model.DefaultSettings() {
		switch setting.Key {
		case model.SettingKeyAIAutomationEnabled,
			model.SettingKeyAIAutomationBaseURL,
			model.SettingKeyAIAutomationAPIKey,
			model.SettingKeyAIAutomationChannelType,
			model.SettingKeyAIAutomationModel,
			model.SettingKeyAIAutomationUseLocalDefault,
			model.SettingKeyConfigSourceMode,
			model.SettingKeyActiveAIProfileID,
			model.SettingKeyDynamicRoutingLearningEnabled:
			if err := db.Exec("INSERT INTO settings (key, value) SELECT ?, ? WHERE NOT EXISTS (SELECT 1 FROM settings WHERE key = ?)", setting.Key, setting.Value, setting.Key).Error; err != nil {
				return fmt.Errorf("failed to ensure setting %s: %w", setting.Key, err)
			}
		}
	}
	return nil
}
