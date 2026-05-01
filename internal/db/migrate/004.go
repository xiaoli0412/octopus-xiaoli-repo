package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 4,
		Up:      addLLMRouteTargetDefaults,
	})
}

// 004:
// - llm_infos: add model-level route-target defaults for billing/probe behavior
func addLLMRouteTargetDefaults(db *gorm.DB) error {
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

	addTextColumn := func(table, column, defaultValue string) error {
		if hasColumn(table, column) {
			return nil
		}
		var sql string
		switch dialect {
		case "postgres":
			sql = fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s TEXT DEFAULT '%s'", table, column, defaultValue)
		case "mysql":
			sql = fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` TEXT", table, column)
		default:
			sql = fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s TEXT DEFAULT '%s'", table, column, defaultValue)
		}
		return db.Exec(sql).Error
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

	if err := addTextColumn("llm_infos", "billing_mode", "unknown"); err != nil {
		return fmt.Errorf("failed to add llm_infos.billing_mode: %w", err)
	}
	if err := addTextColumn("llm_infos", "probe_policy", "passive_only"); err != nil {
		return fmt.Errorf("failed to add llm_infos.probe_policy: %w", err)
	}
	if err := addIntColumn("llm_infos", "probe_interval_seconds", 3600); err != nil {
		return fmt.Errorf("failed to add llm_infos.probe_interval_seconds: %w", err)
	}
	if err := addIntColumn("llm_infos", "probe_concurrency_limit", 1); err != nil {
		return fmt.Errorf("failed to add llm_infos.probe_concurrency_limit: %w", err)
	}

	return nil
}
