package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 9,
		Up:      addMilestoneThreeDefaults,
	})
}

// 009:
// - channels: add key_management_mode and key_routing_policy
// - settings: ensure milestone-3 runtime defaults exist
func addMilestoneThreeDefaults(db *gorm.DB) error {
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

	if err := addTextColumn("channels", "key_management_mode", "pooled"); err != nil {
		return fmt.Errorf("failed to add channels.key_management_mode: %w", err)
	}
	if err := addTextColumn("channels", "key_routing_policy", "round_robin"); err != nil {
		return fmt.Errorf("failed to add channels.key_routing_policy: %w", err)
	}

	settingDefaults := map[string]string{
		"dynamic_routing_health_enabled": "true",
		"race_global_budget":             "64",
		"race_group_budget":              "8",
		"race_channel_budget":            "4",
		"race_key_budget":                "2",
		"race_probe_budget":              "16",
	}
	for key, value := range settingDefaults {
		if err := db.Exec("INSERT INTO settings (key, value) SELECT ?, ? WHERE NOT EXISTS (SELECT 1 FROM settings WHERE key = ?)", key, value, key).Error; err != nil {
			return fmt.Errorf("failed to ensure setting %s: %w", key, err)
		}
	}

	return nil
}
