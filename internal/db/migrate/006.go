package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 6,
		Up:      addOfficialLLMPrices,
	})
}

// 006:
// - llm_infos: add official price fields for dual-view display (official vs gateway)
func addOfficialLLMPrices(db *gorm.DB) error {
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

	addFloatColumn := func(table, column string) error {
		if hasColumn(table, column) {
			return nil
		}
		var sql string
		switch dialect {
		case "postgres":
			sql = fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s DOUBLE PRECISION DEFAULT 0", table, column)
		case "mysql":
			sql = fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` DOUBLE DEFAULT 0", table, column)
		default:
			sql = fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s REAL DEFAULT 0", table, column)
		}
		return db.Exec(sql).Error
	}

	columns := []string{"official_input", "official_output", "official_cache_read", "official_cache_write"}
	for _, col := range columns {
		if err := addFloatColumn("llm_infos", col); err != nil {
			return fmt.Errorf("failed to add llm_infos.%s: %w", col, err)
		}
	}

	return nil
}
