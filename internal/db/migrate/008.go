package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 8,
		Up:      addGroupAdvancedFailoverFields,
	})
}

// 008:
// - groups: add advanced failover configuration fields
func addGroupAdvancedFailoverFields(db *gorm.DB) error {
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

	addIntColumn := func(table, column string, defaultValue int) error {
		if hasColumn(table, column) {
			return nil
		}
		var sql string
		switch dialect {
		case "postgres":
			sql = fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s INTEGER DEFAULT %d", table, column, defaultValue)
		case "mysql":
			sql = fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` INT DEFAULT %d", table, column, defaultValue)
		default:
			sql = fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s INTEGER DEFAULT %d", table, column, defaultValue)
		}
		return db.Exec(sql).Error
	}

	if err := addIntColumn("groups", "failover_window_sec", 360); err != nil {
		return fmt.Errorf("failed to add groups.failover_window_sec: %w", err)
	}
	if err := addIntColumn("groups", "race_after_fails", 2); err != nil {
		return fmt.Errorf("failed to add groups.race_after_fails: %w", err)
	}
	if err := addIntColumn("groups", "race_concurrency", 2); err != nil {
		return fmt.Errorf("failed to add groups.race_concurrency: %w", err)
	}

	return nil
}
