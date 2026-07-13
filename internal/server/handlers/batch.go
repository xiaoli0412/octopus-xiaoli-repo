package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/observability"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/auth"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

// batchMaxIDs 限制单次批量操作的最大条目数，防止误操作。
const batchMaxIDs = 100

// auditBatchRecordTimeout 批量审计日志写入的超时时间。
const auditBatchRecordTimeout = 5 * time.Second

// BatchOperationRequest 批量操作的通用请求体。
type BatchOperationRequest struct {
	IDs []int `json:"ids"`
}

// BatchOperationResult 批量操作的统一返回结构。
type BatchOperationResult struct {
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	Errors       []string `json:"errors"`
}

// BatchAction 描述一个针对单条记录的操作。
// 操作成功时返回 nil；失败时返回 error，其 Error() 内容会被追加到 Errors 列表。
type BatchAction func(ctx context.Context, id int) error

// parseBatchRequest 解析批量操作请求体，并对 ID 数量做上限保护。
// 解析失败或超出上限时返回错误并通过 resp.Error 写回响应。
func parseBatchRequest(c *gin.Context) (*BatchOperationRequest, bool) {
	var req BatchOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return nil, false
	}
	if len(req.IDs) == 0 {
		resp.Error(c, http.StatusBadRequest, "ids is required")
		return nil, false
	}
	if len(req.IDs) > batchMaxIDs {
		resp.Error(c, http.StatusBadRequest, "too many ids, max "+strconv.Itoa(batchMaxIDs))
		return nil, false
	}
	seen := make(map[int]struct{}, len(req.IDs))
	deduped := make([]int, 0, len(req.IDs))
	for _, id := range req.IDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		deduped = append(deduped, id)
	}
	if len(deduped) == 0 {
		resp.Error(c, http.StatusBadRequest, "ids must be positive integers")
		return nil, false
	}
	req.IDs = deduped
	return &req, true
}

// runBatchOperation 顺序执行批量操作，收集成功/失败计数与错误信息。
// 审计日志由 AuditMiddleware 自动记录（基于路由前缀与请求体），此处不再重复写入。
func runBatchOperation(ctx context.Context, req *BatchOperationRequest, action BatchAction) BatchOperationResult {
	result := BatchOperationResult{
		Errors: make([]string, 0),
	}
	for _, id := range req.IDs {
		if err := action(ctx, id); err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, "id "+strconv.Itoa(id)+": "+err.Error())
			log.Warnf("batch operation failed for id=%d: %v", id, err)
			continue
		}
		result.SuccessCount++
	}
	return result
}

// recordBatchAudit 异步记录单条批量操作的审计日志。
// 由于 AuditMiddleware 对批量操作只能记录到一个 resource_id（取请求体中的 id 字段，
// 但批量操作请求体是 ids 数组），因此这里针对每条记录单独写一条审计日志，便于追溯。
func recordBatchAudit(c *gin.Context, action observability.AuditAction, resourceType observability.ResourceType, id int, resourceName string) {
	if auditRecorderInstance == nil {
		return
	}
	userID, username := extractUserIDAndUsernameFromContext(c)
	al := observability.AuditLog{
		UserID:       userID,
		Username:     username,
		Action:       string(action),
		ResourceType: string(resourceType),
		ResourceID:   strconv.Itoa(id),
		ResourceName: resourceName,
		IP:           c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	}
	go func(al observability.AuditLog) {
		bgCtx, cancel := context.WithTimeout(context.Background(), auditBatchRecordTimeout)
		defer cancel()
		if err := auditRecorderInstance.Record(bgCtx, al); err != nil {
			log.Warnf("failed to record batch audit log for %s id=%d: %v", resourceType, id, err)
		}
	}(al)
}

// extractUserIDAndUsernameFromContext 从 gin 上下文的 JWT 中提取用户 ID 与用户名。
// 复用 auth.VerifyJWTToken 验证 token 有效性后，再解析 Subject claim。
func extractUserIDAndUsernameFromContext(c *gin.Context) (uint, string) {
	authHeader := c.GetHeader("Authorization")
	token := extractBearerToken(authHeader)
	if token == "" {
		return 0, ""
	}
	if !auth.VerifyJWTToken(token) {
		return 0, ""
	}
	claims := &jwt.RegisteredClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(token, claims); err != nil {
		return 0, ""
	}
	username := strings.TrimSpace(claims.Subject)
	user := op.UserGet()
	return user.ID, username
}

// extractBearerToken 从 Authorization 头中提取 Bearer token。
func extractBearerToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	const bearerPrefix = "Bearer "
	if len(value) < len(bearerPrefix) || !strings.EqualFold(value[:len(bearerPrefix)], bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(value[len(bearerPrefix):])
}
