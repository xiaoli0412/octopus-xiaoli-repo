package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/observability"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/auth"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

// auditSensitivePrefixes 需要记录审计日志的敏感路由前缀
var auditSensitivePrefixes = []string{
	"/api/v1/channel",
	"/api/v1/group",
	"/api/v1/apikey",
	"/api/v1/setting",
	"/api/v1/user",
	"/api/v1/backup",
}

// isAuditSensitiveRequest 判断请求是否需要记录审计日志
func isAuditSensitiveRequest(method, path string) bool {
	if method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete {
		return false
	}
	for _, prefix := range auditSensitivePrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// inferAuditAction 从请求方法与路径推断审计操作类型
func inferAuditAction(method, path string) string {
	switch {
	case method == http.MethodPost && strings.Contains(path, "/delete"):
		return string(observability.AuditActionDelete)
	case method == http.MethodDelete:
		return string(observability.AuditActionDelete)
	case method == http.MethodPost && (strings.Contains(path, "/enable") || strings.Contains(path, "/disable")):
		if strings.Contains(path, "/disable") {
			return string(observability.AuditActionDisable)
		}
		return string(observability.AuditActionEnable)
	case method == http.MethodPost && strings.Contains(path, "/restore"):
		return string(observability.AuditActionRestore)
	case method == http.MethodPost && strings.Contains(path, "/backup"):
		return string(observability.AuditActionBackup)
	case method == http.MethodPost && strings.Contains(path, "/login"):
		return string(observability.AuditActionLogin)
	case method == http.MethodPost && strings.Contains(path, "/create"):
		return string(observability.AuditActionCreate)
	case method == http.MethodPost || method == http.MethodPut:
		return string(observability.AuditActionUpdate)
	default:
		return string(observability.AuditActionUpdate)
	}
}

// inferResourceType 从路径推断资源类型
func inferResourceType(path string) string {
	switch {
	case strings.HasPrefix(path, "/api/v1/channel"):
		return string(observability.ResourceTypeChannel)
	case strings.HasPrefix(path, "/api/v1/group"):
		return string(observability.ResourceTypeGroup)
	case strings.HasPrefix(path, "/api/v1/apikey"):
		return string(observability.ResourceTypeAPIKey)
	case strings.HasPrefix(path, "/api/v1/setting"):
		return string(observability.ResourceTypeSetting)
	case strings.HasPrefix(path, "/api/v1/user"):
		return string(observability.ResourceTypeUser)
	default:
		return ""
	}
}

// shouldFetchBeforeSnapshot 判断是否需要在 c.Next() 前获取变更前快照。
// 仅对更新、删除、启用、禁用类操作需要记录 before 快照。
func shouldFetchBeforeSnapshot(action, resourceType string) bool {
	if resourceType == "" {
		return false
	}
	switch action {
	case string(observability.AuditActionUpdate),
		string(observability.AuditActionDelete),
		string(observability.AuditActionEnable),
		string(observability.AuditActionDisable):
		return true
	default:
		return false
	}
}

// extractResourceID 从请求体 JSON 或路径参数中推断资源 ID。
// 优先解析请求体中的 id 字段（setting 资源使用 key 字段），
// 其次从路径末尾提取数字 ID，最后对单例 user 资源回退到 JWT 用户 ID。
func extractResourceID(path, resourceType string, body []byte, userID uint) string {
	// 优先从请求体 JSON 中提取
	if len(body) > 0 {
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(body, &bodyMap); err == nil {
			// setting 资源使用 key 字段作为资源 ID
			if resourceType == string(observability.ResourceTypeSetting) {
				if key, ok := bodyMap["key"].(string); ok && key != "" {
					return key
				}
			}
			// 其他资源使用 id 字段
			if id, ok := bodyMap["id"]; ok {
				switch v := id.(type) {
				case float64:
					if v != 0 {
						return strconv.FormatFloat(v, 'f', -1, 64)
					}
				case string:
					if v != "" {
						return v
					}
				}
			}
		}
	}

	// user 资源是单例，使用 JWT 中的用户 ID
	if resourceType == string(observability.ResourceTypeUser) && userID > 0 {
		return strconv.FormatUint(uint64(userID), 10)
	}

	// 从路径末尾提取数字 ID（如 /api/v1/apikey/delete/1）
	segments := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(segments) > 0 {
		last := segments[len(segments)-1]
		if _, err := strconv.Atoi(last); err == nil {
			return last
		}
	}

	return ""
}

// fetchResourceSnapshot 按 resourceType 分发查询当前记录快照，返回 JSON 字符串。
// 查询失败时返回错误，调用方负责降级处理，不阻塞请求。
func fetchResourceSnapshot(resourceType string, resourceID string) (string, error) {
	if resourceID == "" {
		return "", nil
	}

	database := db.GetDB()
	if database == nil {
		return "", fmt.Errorf("database not initialized")
	}

	switch resourceType {
	case string(observability.ResourceTypeChannel):
		id, err := strconv.Atoi(resourceID)
		if err != nil {
			return "", fmt.Errorf("invalid channel id: %s", resourceID)
		}
		var channel model.Channel
		if err := database.First(&channel, id).Error; err != nil {
			return "", err
		}
		data, err := json.Marshal(channel)
		if err != nil {
			return "", err
		}
		return string(data), nil

	case string(observability.ResourceTypeGroup):
		id, err := strconv.Atoi(resourceID)
		if err != nil {
			return "", fmt.Errorf("invalid group id: %s", resourceID)
		}
		var group model.Group
		if err := database.First(&group, id).Error; err != nil {
			return "", err
		}
		data, err := json.Marshal(group)
		if err != nil {
			return "", err
		}
		return string(data), nil

	case string(observability.ResourceTypeAPIKey):
		id, err := strconv.Atoi(resourceID)
		if err != nil {
			return "", fmt.Errorf("invalid apikey id: %s", resourceID)
		}
		var apiKey model.APIKey
		if err := database.First(&apiKey, id).Error; err != nil {
			return "", err
		}
		data, err := json.Marshal(apiKey)
		if err != nil {
			return "", err
		}
		return string(data), nil

	case string(observability.ResourceTypeSetting):
		setting := model.Setting{Key: model.SettingKey(resourceID)}
		if err := database.First(&setting).Error; err != nil {
			return "", err
		}
		data, err := json.Marshal(setting)
		if err != nil {
			return "", err
		}
		return string(data), nil

	case string(observability.ResourceTypeUser):
		id, err := strconv.Atoi(resourceID)
		if err != nil {
			return "", fmt.Errorf("invalid user id: %s", resourceID)
		}
		var user model.User
		if err := database.First(&user, id).Error; err != nil {
			return "", err
		}
		data, err := json.Marshal(user)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	return "", nil
}

// extractUsernameFromJWT 从 JWT token 的 Subject claim 中提取用户名
func extractUsernameFromJWT(c *gin.Context) (uint, string) {
	authHeader := c.GetHeader("Authorization")
	token, ok := bearerTokenFromHeader(authHeader)
	if !ok {
		return 0, ""
	}

	// 复用 auth 包的验证逻辑：先验证 token 有效
	if !auth.VerifyJWTToken(token) {
		return 0, ""
	}

	// 重新解析获取 Subject（用户名）
	claims := &jwt.RegisteredClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(token, claims)
	if err != nil {
		return 0, ""
	}
	username := strings.TrimSpace(claims.Subject)

	// 通过 op.UserGet 获取用户 ID
	user := op.UserGet()
	return user.ID, username
}

// AuditMiddleware 审计日志中间件，对敏感路由的写操作异步记录审计日志。
// 对更新/删除/启停类操作，在 c.Next() 前查询变更前快照并填入 BeforeJSON。
func AuditMiddleware(recorder observability.AuditRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		if !isAuditSensitiveRequest(method, path) || recorder == nil {
			c.Next()
			return
		}

		// 读取请求 body 用于记录 after 值（针对 POST/PUT）
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		// 提取用户信息
		userID, username := extractUsernameFromJWT(c)

		// 推断 action 与 resourceType
		action := inferAuditAction(method, path)
		resourceType := inferResourceType(path)

		// 提取资源 ID（用于审计日志字段与 before 快照查询）
		resourceID := extractResourceID(path, resourceType, bodyBytes, userID)

		// 对更新/删除/启停类操作，在 c.Next() 前查询当前快照
		beforeJSON := ""
		if shouldFetchBeforeSnapshot(action, resourceType) && resourceID != "" {
			snapshot, err := fetchResourceSnapshot(resourceType, resourceID)
			if err != nil {
				beforeJSON = fmt.Sprintf("[error: %s]", err.Error())
			} else {
				beforeJSON = snapshot
			}
		}

		// 使用 response body writer 捕获响应
		writer := &auditResponseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		c.Next()

		// 计算 afterJSON
		afterJSON := ""
		// 删除操作：after 设为 [deleted]
		if action == string(observability.AuditActionDelete) {
			afterJSON = "[deleted]"
		} else if writer.body.Len() > 0 {
			afterJSON = writer.body.String()
		} else if len(bodyBytes) > 0 {
			afterJSON = string(bodyBytes)
		}

		// 脱敏：移除 password/key/token/secret 等敏感字段值
		beforeJSON = observability.SanitizeAuditJSON(beforeJSON)
		afterJSON = observability.SanitizeAuditJSON(afterJSON)

		auditLog := observability.AuditLog{
			UserID:       userID,
			Username:     username,
			Action:       action,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			BeforeJSON:   truncateAuditJSON(beforeJSON),
			AfterJSON:    truncateAuditJSON(afterJSON),
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			CreatedAt:    time.Now(),
		}

		// 异步写入，不阻塞请求
		go func(al observability.AuditLog) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := recorder.Record(ctx, al); err != nil {
				log.Warnf("failed to record audit log: %v", err)
			}
		}(auditLog)
	}
}

// auditResponseWriter 包装 gin.ResponseWriter 以捕获响应 body
type auditResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *auditResponseWriter) Write(b []byte) (int, error) {
	// 限制捕获大小，避免大响应占用内存
	if w.body.Len() < maxAuditResponseBodyBytes {
		w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

const maxAuditResponseBodyBytes = 64 * 1024

func truncateAuditJSON(s string) string {
	if len(s) > maxAuditResponseBodyBytes {
		return s[:maxAuditResponseBodyBytes] + "...[truncated]"
	}
	return s
}
