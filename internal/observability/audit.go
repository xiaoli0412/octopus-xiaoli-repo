package observability

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// sensitiveKeys 需要脱敏的 JSON 字段名（小写匹配）
var sensitiveKeys = map[string]bool{
	"password":      true,
	"key":           true,
	"token":         true,
	"secret":        true,
	"api_key":       true,
	"apikey":        true,
	"access_token":  true,
	"refresh_token": true,
	"private_key":   true,
	"passphrase":    true,
	"authorization": true,
	"cookie":        true,
}

// maskValue 脱敏后的占位值
const maskValue = "***"

// SanitizeAuditJSON 对 JSON 字符串中的敏感字段值进行脱敏。
// 先尝试解析为 JSON 递归处理；解析失败时用正则兜底替换。
func SanitizeAuditJSON(s string) string {
	if s == "" {
		return s
	}

	// 尝试 JSON 解析
	var data interface{}
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		// 非 JSON，用正则兜底
		return sanitizeWithRegex(s)
	}

	sanitizeValue(data)
	masked, err := json.Marshal(data)
	if err != nil {
		return sanitizeWithRegex(s)
	}
	return string(masked)
}

// sanitizeValue 递归脱敏 map/slice 中的敏感字段
func sanitizeValue(v interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, child := range val {
			if sensitiveKeys[strings.ToLower(k)] {
				val[k] = maskValue
			} else {
				sanitizeValue(child)
			}
		}
	case []interface{}:
		for _, item := range val {
			sanitizeValue(item)
		}
	}
}

// sanitizeWithRegex 使用正则对非 JSON 文本中的敏感字段进行兜底脱敏
func sanitizeWithRegex(s string) string {
	for key := range sensitiveKeys {
		// 匹配 "key":"value" 或 "key": "value"（不区分大小写）
		re := regexp.MustCompile(`(?i)"` + regexp.QuoteMeta(key) + `"\s*:\s*"[^"]*"`)
		s = re.ReplaceAllString(s, `"`+key+`":"`+maskValue+`"`)
	}
	return s
}

// AuditAction 审计操作类型
type AuditAction string

const (
	AuditActionCreate  AuditAction = "create"
	AuditActionUpdate  AuditAction = "update"
	AuditActionDelete  AuditAction = "delete"
	AuditActionEnable  AuditAction = "enable"
	AuditActionDisable AuditAction = "disable"
	AuditActionLogin   AuditAction = "login"
	AuditActionBackup  AuditAction = "backup"
	AuditActionRestore AuditAction = "restore"
)

// ResourceType 审计资源类型
type ResourceType string

const (
	ResourceTypeChannel ResourceType = "channel"
	ResourceTypeGroup   ResourceType = "group"
	ResourceTypeAPIKey  ResourceType = "apikey"
	ResourceTypeSetting ResourceType = "setting"
	ResourceTypeUser    ResourceType = "user"
)

// AuditLog 对应 audit_log 表
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

// AuditQueryFilter 审计日志查询过滤条件
type AuditQueryFilter struct {
	StartTime     time.Time
	EndTime       time.Time
	UserID        uint
	Action        string
	ResourceType  string
	ResourceID    string
	Page          int
	PageSize      int
}

// AuditRecorder 审计日志记录器接口
type AuditRecorder interface {
	Record(ctx context.Context, log AuditLog) error
	Query(ctx context.Context, filter AuditQueryFilter) ([]AuditLog, int64, error)
	GetByID(ctx context.Context, id uint) (*AuditLog, error)
}

// auditRecorder 基于 GORM 的审计日志记录器实现
type auditRecorder struct {
	db *gorm.DB
}

// NewAuditRecorder 创建一个新的 AuditRecorder
func NewAuditRecorder(db *gorm.DB) AuditRecorder {
	return &auditRecorder{db: db}
}

// Record 异步写入一条审计日志
func (r *auditRecorder) Record(ctx context.Context, log AuditLog) error {
	if r.db == nil {
		return nil
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(&log).Error
}

// Query 分页查询审计日志
func (r *auditRecorder) Query(ctx context.Context, filter AuditQueryFilter) ([]AuditLog, int64, error) {
	if r.db == nil {
		return nil, 0, nil
	}

	query := r.db.WithContext(ctx).Model(&AuditLog{})

	if !filter.StartTime.IsZero() {
		query = query.Where("created_at >= ?", filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		query = query.Where("created_at <= ?", filter.EndTime)
	}
	if filter.UserID > 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.ResourceType != "" {
		query = query.Where("resource_type = ?", filter.ResourceType)
	}
	if filter.ResourceID != "" {
		query = query.Where("resource_id = ?", filter.ResourceID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	var logs []AuditLog
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByID 按 ID 查询单条审计日志
func (r *auditRecorder) GetByID(ctx context.Context, id uint) (*AuditLog, error) {
	if r.db == nil {
		return nil, nil
	}
	var log AuditLog
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}
