package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 16,
		Up:      addCachePolicyAndOpsEligibility,
	})
}

func addColumnIfMissing(db *gorm.DB, table, sqlTable, column, definition string) error {
	if db.Migrator().HasColumn(table, column) {
		return nil
	}
	if err := db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", sqlTable, column, definition)).Error; err != nil {
		return fmt.Errorf("failed to add %s.%s: %w", table, column, err)
	}
	return nil
}

func addCachePolicyAndOpsEligibility(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	llmTable := "llm_infos"
	opsTable := "ops_metric_buckets"
	llmSQLTable := llmTable
	opsSQLTable := opsTable
	if db.Dialector.Name() == "mysql" {
		llmSQLTable = "`llm_infos`"
		opsSQLTable = "`ops_metric_buckets`"
	}

	for _, col := range []struct {
		table      string
		sqlTable   string
		column     string
		definition string
	}{
		{llmTable, llmSQLTable, "cache_policy", "TEXT"},
		{llmTable, llmSQLTable, "cache_reason", "TEXT"},
		{llmTable, llmSQLTable, "upstream_provider_type", "TEXT"},
		{llmTable, llmSQLTable, "upstream_source", "TEXT"},
		{opsTable, opsSQLTable, "cache_eligible_count", "BIGINT DEFAULT 0"},
		{opsTable, opsSQLTable, "cache_ineligible_count", "BIGINT DEFAULT 0"},
	} {
		if err := addColumnIfMissing(db, col.table, col.sqlTable, col.column, col.definition); err != nil {
			return err
		}
	}
	return nil
}
