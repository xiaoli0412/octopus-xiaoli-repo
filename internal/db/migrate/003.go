package migrate

import (
    "fmt"
	"gorm.io/gorm"
)

func init() {
    RegisterAfterAutoMigration(Migration{
        Version: 3,
        Up:      addChannelKeyAllowedModelsAndGroupRetryFields,
    })
}

// 003:
// - channel_keys: add allowed_models for per-key model binding
// - groups: add retry_rounds and retry_delay_ms for multi-round retry
func addChannelKeyAllowedModelsAndGroupRetryFields(db *gorm.DB) error {
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

    // channel_keys.allowed_models
    if !hasColumn("channel_keys", "allowed_models") {
        var sql string
        switch dialect {
        case "postgres":
            sql = "ALTER TABLE channel_keys ADD COLUMN IF NOT EXISTS allowed_models TEXT DEFAULT ''"
        case "mysql":
            sql = "ALTER TABLE `channel_keys` ADD COLUMN `allowed_models` TEXT"
        default: // sqlite and others
            sql = "ALTER TABLE channel_keys ADD COLUMN allowed_models TEXT DEFAULT ''"
        }
        if err := db.Exec(sql).Error; err != nil {
            return fmt.Errorf("failed to add channel_keys.allowed_models: %w", err)
        }
    }

    // groups.retry_rounds
    if !hasColumn("groups", "retry_rounds") {
        var sql string
        switch dialect {
        case "postgres":
            sql = "ALTER TABLE groups ADD COLUMN IF NOT EXISTS retry_rounds INTEGER DEFAULT 1"
        case "mysql":
            sql = "ALTER TABLE `groups` ADD COLUMN `retry_rounds` INT DEFAULT 1"
        default:
            sql = "ALTER TABLE groups ADD COLUMN retry_rounds INTEGER DEFAULT 1"
        }
        if err := db.Exec(sql).Error; err != nil {
            return fmt.Errorf("failed to add groups.retry_rounds: %w", err)
        }
    }

    // groups.retry_delay_ms
    if !hasColumn("groups", "retry_delay_ms") {
        var sql string
        switch dialect {
        case "postgres":
            sql = "ALTER TABLE groups ADD COLUMN IF NOT EXISTS retry_delay_ms INTEGER DEFAULT 0"
        case "mysql":
            sql = "ALTER TABLE `groups` ADD COLUMN `retry_delay_ms` INT DEFAULT 0"
        default:
            sql = "ALTER TABLE groups ADD COLUMN retry_delay_ms INTEGER DEFAULT 0"
        }
        if err := db.Exec(sql).Error; err != nil {
            return fmt.Errorf("failed to add groups.retry_delay_ms: %w", err)
        }
    }

    return nil
}
