package migrate

import (
	"fmt"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 19,
		Up:      createAuditLogTable,
	})
}

// 019:
// - 创建 audit_logs 表用于记录管理操作审计日志
func createAuditLogTable(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if err := db.AutoMigrate(&model.AuditLog{}); err != nil {
		return fmt.Errorf("auto migrate audit_logs: %w", err)
	}
	return nil
}
