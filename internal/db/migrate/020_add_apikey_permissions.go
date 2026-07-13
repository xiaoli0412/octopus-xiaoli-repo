package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 20,
		Up:      addAPIKeyPermissionColumns,
	})
}

// 020:
// - 为 api_keys 表增加 API Key 权限粒度字段：
//   - allowed_channels      允许调用的渠道 ID 列表（JSON 数组）
//   - allowed_groups        允许调用的分组 ID 列表（JSON 数组）
//   - allowed_capabilities  允许调用的能力列表（JSON 数组：chat/embedding/response/message）
//   - allowed_ip_cidrs      允许调用方 IP 的 CIDR 白名单（JSON 数组）
//
// 所有字段为空表示不限制，保证向后兼容。
func addAPIKeyPermissionColumns(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	table := "api_keys"
	sqlTable := table
	if db.Dialector.Name() == "mysql" {
		sqlTable = "`api_keys`"
	}

	for _, col := range []struct {
		column     string
		definition string
	}{
		{"allowed_channels", "TEXT"},
		{"allowed_groups", "TEXT"},
		{"allowed_capabilities", "TEXT"},
		{"allowed_ip_cidrs", "TEXT"},
	} {
		if err := addColumnIfMissing(db, table, sqlTable, col.column, col.definition); err != nil {
			return err
		}
	}
	return nil
}
