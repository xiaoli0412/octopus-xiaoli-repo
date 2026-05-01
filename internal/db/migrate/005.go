package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 5,
		Up:      addLLMCanonicalName,
	})
}

// 005:
// - llm_infos: add canonical_name for price normalization and alias resolution
func addLLMCanonicalName(db *gorm.DB) error {
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

	if hasColumn("llm_infos", "canonical_name") {
		return nil
	}

	var sql string
	if dialect == "mysql" {
		sql = "ALTER TABLE `llm_infos` ADD COLUMN `canonical_name` TEXT"
	} else if dialect == "postgres" {
		sql = "ALTER TABLE llm_infos ADD COLUMN IF NOT EXISTS canonical_name TEXT"
	} else {
		sql = "ALTER TABLE llm_infos ADD COLUMN canonical_name TEXT DEFAULT ''"
	}

	if err := db.Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to add llm_infos.canonical_name: %w", err)
	}

	return nil
}
