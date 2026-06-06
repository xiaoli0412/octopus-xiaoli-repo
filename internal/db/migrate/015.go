package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 15,
		Up:      addChannelKeyRequestCapabilities,
	})
}

func addChannelKeyRequestCapabilities(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if db.Migrator().HasColumn("channel_keys", "request_capabilities") {
		return nil
	}

	var sql string
	switch db.Dialector.Name() {
	case "postgres":
		sql = "ALTER TABLE channel_keys ADD COLUMN IF NOT EXISTS request_capabilities TEXT DEFAULT ''"
	case "mysql":
		sql = "ALTER TABLE `channel_keys` ADD COLUMN `request_capabilities` TEXT"
	default:
		sql = "ALTER TABLE channel_keys ADD COLUMN request_capabilities TEXT DEFAULT ''"
	}
	if err := db.Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to add channel_keys.request_capabilities: %w", err)
	}
	return nil
}
