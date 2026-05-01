package migrate

import (
	"fmt"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 10,
		Up:      addRouteTargetOverrides,
	})
}

// 010:
// - route_target_overrides: add explicit (channel, key, model) override table
func addRouteTargetOverrides(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if err := db.AutoMigrate(&model.RouteTargetOverride{}); err != nil {
		return fmt.Errorf("auto migrate route_target_overrides: %w", err)
	}
	return nil
}
