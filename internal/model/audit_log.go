package model

import "time"

// AuditLog 对应 audit_log 表，记录管理操作的审计日志
type AuditLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"index"`
	Username     string    `json:"username" gorm:"size:128"`
	Action       string    `json:"action" gorm:"size:32;index"`
	ResourceType string    `json:"resource_type" gorm:"size:32;index"`
	ResourceID   string    `json:"resource_id" gorm:"size:128"`
	ResourceName string    `json:"resource_name" gorm:"size:256"`
	BeforeJSON   string    `json:"before_json" gorm:"type:text"`
	AfterJSON    string    `json:"after_json" gorm:"type:text"`
	IP           string    `json:"ip" gorm:"size:64"`
	UserAgent    string    `json:"user_agent" gorm:"size:512"`
	CreatedAt    time.Time `json:"created_at" gorm:"index"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}
