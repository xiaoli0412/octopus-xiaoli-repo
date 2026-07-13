package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/observability"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
)

var auditRecorderInstance observability.AuditRecorder

// SetAuditRecorder 设置审计日志记录器实例（由 server 在 DB 初始化后调用）
func SetAuditRecorder(recorder observability.AuditRecorder) {
	auditRecorderInstance = recorder
}

func init() {
	router.NewGroupRouter("/api/v1/audit").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(listAuditLogs),
		).
		AddRoute(
			router.NewRoute("/:id", http.MethodGet).
				Handle(getAuditLog),
		)
}

// listAuditLogs 分页查询审计日志，支持按 time_range/user_id/action/resource_type 过滤
func listAuditLogs(c *gin.Context) {
	page, _, err := parseOptionalBoundedIntQuery(c, "page", 1, 1, 0)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, "invalid page")
		return
	}
	pageSize, _, err := parseOptionalBoundedIntQuery(c, "page_size", 20, 1, 100)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, "invalid page_size")
		return
	}

	startTimeStr, startTimeExists := c.GetQuery("start_time")
	endTimeStr, endTimeExists := c.GetQuery("end_time")

	filter := observability.AuditQueryFilter{
		Page:     page,
		PageSize: pageSize,
	}

	if startTimeExists && startTimeStr != "" {
		if ts, err := strconv.ParseInt(startTimeStr, 10, 64); err == nil {
			filter.StartTime = time.Unix(ts, 0)
		} else if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = t
		}
	}
	if endTimeExists && endTimeStr != "" {
		if ts, err := strconv.ParseInt(endTimeStr, 10, 64); err == nil {
			filter.EndTime = time.Unix(ts, 0)
		} else if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = t
		}
	}

	if uid := c.Query("user_id"); uid != "" {
		if id, err := strconv.ParseUint(uid, 10, 64); err == nil {
			filter.UserID = uint(id)
		}
	}
	filter.Action = c.Query("action")
	filter.ResourceType = c.Query("resource_type")
	filter.ResourceID = c.Query("resource_id")

	logs, total, err := auditRecorderInstance.Query(c.Request.Context(), filter)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to list audit logs")
		return
	}

	resp.Success(c, gin.H{
		"list":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// getAuditLog 按 ID 查询单条审计日志
func getAuditLog(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		resp.Error(c, http.StatusBadRequest, "missing id")
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	log, err := auditRecorderInstance.GetByID(c.Request.Context(), uint(id))
	if err != nil || log == nil {
		resp.Error(c, http.StatusNotFound, "audit log not found")
		return
	}

	resp.Success(c, log)
}
