package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 18,
		Up:      addLLMInfoDisabledColumn,
	})
}

func addLLMInfoDisabledColumn(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	llmTable := "llm_infos"
	llmSQLTable := llmTable
	if db.Dialector.Name() == "mysql" {
		llmSQLTable = "`llm_infos`"
	}

	var definition string
	switch db.Dialector.Name() {
	case "postgres":
		definition = "BOOLEAN DEFAULT FALSE"
	case "mysql":
		definition = "TINYINT(1) DEFAULT 0"
	default:
		definition = "INTEGER DEFAULT 0"
	}

	return addColumnIfMissing(db, llmTable, llmSQLTable, "disabled", definition)
}
