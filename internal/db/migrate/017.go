package migrate

import (
	"fmt"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 17,
		Up:      addUpstreamManagementTables,
	})
}

func addUpstreamManagementTables(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if err := db.AutoMigrate(
		&model.UpstreamSite{},
		&model.UpstreamCredential{},
		&model.UpstreamKeySnapshot{},
		&model.UpstreamGroupSnapshot{},
		&model.UpstreamModelPrice{},
	); err != nil {
		return fmt.Errorf("failed to migrate upstream management tables: %w", err)
	}

	channelTable := "channels"
	channelKeyTable := "channel_keys"
	channelSQLTable := channelTable
	channelKeySQLTable := channelKeyTable
	if db.Dialector.Name() == "mysql" {
		channelSQLTable = "`channels`"
		channelKeySQLTable = "`channel_keys`"
	}
	for _, col := range []struct {
		table      string
		sqlTable   string
		column     string
		definition string
	}{
		{channelTable, channelSQLTable, "upstream_site_id", "INTEGER DEFAULT 0"},
		{channelTable, channelSQLTable, "upstream_source", "TEXT"},
		{channelKeyTable, channelKeySQLTable, "upstream_site_id", "INTEGER DEFAULT 0"},
		{channelKeyTable, channelKeySQLTable, "upstream_key_name", "TEXT"},
	} {
		if err := addColumnIfMissing(db, col.table, col.sqlTable, col.column, col.definition); err != nil {
			return err
		}
	}
	return nil
}
