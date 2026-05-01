package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 11,
		Up:      addDynamicRoutingModeDefaults,
	})
}

// 011:
// - relay_logs: add dynamic routing audit columns
// - settings: ensure dynamic routing mode default exists
func addDynamicRoutingModeDefaults(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	settingDefaults := map[string]string{
		"dynamic_routing_mode": "hybrid",
	}
	for key, value := range settingDefaults {
		if err := db.Exec("INSERT INTO settings (key, value) SELECT ?, ? WHERE NOT EXISTS (SELECT 1 FROM settings WHERE key = ?)", key, value, key).Error; err != nil {
			return fmt.Errorf("failed to ensure setting %s: %w", key, err)
		}
	}

	if err := db.AutoMigrate(&relayLogDynamicRoutingColumns{}); err != nil {
		return fmt.Errorf("auto migrate relay log dynamic routing columns: %w", err)
	}

	return nil
}

type relayLogDynamicRoutingColumns struct {
	ID                          int64   `gorm:"primaryKey;autoIncrement:false"`
	DynamicRoutingMode          string  `gorm:"column:dynamic_routing_mode"`
	DynamicRoutingEffectiveMode string  `gorm:"column:dynamic_routing_effective_mode"`
	DynamicRoutingDecision      string  `gorm:"column:dynamic_routing_decision"`
	DynamicRoutingReason        string  `gorm:"column:dynamic_routing_reason"`
	DynamicRoutingConfidence    float64 `gorm:"column:dynamic_routing_confidence"`
	DynamicRoutingFallback      bool    `gorm:"column:dynamic_routing_fallback"`
	DynamicRoutingRecommended   string  `gorm:"column:dynamic_routing_recommended"`
}

func (relayLogDynamicRoutingColumns) TableName() string {
	return "relay_logs"
}
