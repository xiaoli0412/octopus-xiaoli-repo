package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 7,
		Up:      addChannelKeySourceType,
	})
}

// 007:
// - channel_keys: add source_type for paid/free/public route-target strategy inheritance
func addChannelKeySourceType(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	dialect := db.Dialector.Name()
	hasColumn := func(table, column string) bool {
		switch dialect {
		case "sqlite":
			var name string
			db.Raw("SELECT name FROM pragma_table_info(?) WHERE name = ? LIMIT 1", table, column).Scan(&name)
			return name == column
		case "mysql":
			var count int64
			db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?", table, column).Scan(&count)
			return count > 0
		case "postgres":
			var count int64
			db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = ? AND column_name = ?", table, column).Scan(&count)
			return count > 0
		default:
			return db.Migrator().HasColumn(table, column)
		}
	}

	if hasColumn("channel_keys", "source_type") {
		return nil
	}

	var sql string
	if dialect == "mysql" {
		sql = "ALTER TABLE `channel_keys` ADD COLUMN `source_type` TEXT"
	} else if dialect == "postgres" {
		sql = "ALTER TABLE channel_keys ADD COLUMN IF NOT EXISTS source_type TEXT"
	} else {
		sql = "ALTER TABLE channel_keys ADD COLUMN source_type TEXT DEFAULT ''"
	}

	if err := db.Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to add channel_keys.source_type: %w", err)
	}

	return nil
}
