package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/observability"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/auth"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
)

func init() {
	router.NewGroupRouter("/api/v1/apikey").
		Use(middleware.Auth()).
		Use(middleware.RequireJSON()).
		AddRoute(
			router.NewRoute("/create", http.MethodPost).
				Handle(createAPIKey),
		).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(listAPIKey),
		).
		AddRoute(
			router.NewRoute("/update", http.MethodPost).
				Handle(updateAPIKey),
		).
		AddRoute(
			router.NewRoute("/delete/:id", http.MethodDelete).
				Handle(deleteAPIKey),
		).
		AddRoute(
			router.NewRoute("/batch-enable", http.MethodPost).
				Handle(batchEnableAPIKey),
		).
		AddRoute(
			router.NewRoute("/batch-disable", http.MethodPost).
				Handle(batchDisableAPIKey),
		).
		AddRoute(
			router.NewRoute("/batch-delete", http.MethodPost).
				Handle(batchDeleteAPIKey),
		)
	router.NewGroupRouter("/api/v1/apikey").
		Use(middleware.APIKeyAuth()).
		AddRoute(
			router.NewRoute("/stats", http.MethodGet).
				Handle(getStatsAPIKeyById),
		).
		AddRoute(
			router.NewRoute("/login", http.MethodGet).
				Handle(loginAPIKey),
		)
}

func createAPIKey(c *gin.Context) {
	var req model.APIKey
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	req.APIKey = auth.GenerateAPIKey()
	if err := op.APIKeyCreate(&req, c.Request.Context()); err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to create api key")
		return
	}
	resp.Success(c, req)
}

func listAPIKey(c *gin.Context) {
	apiKeys, err := op.APIKeyList(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to list api keys")
		return
	}
	resp.Success(c, apiKeys)
}

func updateAPIKey(c *gin.Context) {
	var req model.APIKey
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := op.APIKeyUpdate(&req, c.Request.Context()); err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to update api key")
		return
	}
	resp.Success(c, req)
}

func deleteAPIKey(c *gin.Context) {
	idNum, ok := parsePositivePathIDValue(c, "id")
	if !ok {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidParam)
		return
	}
	if err := op.APIKeyDelete(idNum, c.Request.Context()); err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to delete api key")
		return
	}
	resp.Success(c, nil)
}

// batchEnableAPIKey 批量启用 API Key
func batchEnableAPIKey(c *gin.Context) {
	req, ok := parseBatchRequest(c)
	if !ok {
		return
	}
	result := runBatchOperation(c.Request.Context(), req, func(ctx context.Context, id int) error {
		key, err := op.APIKeyGet(id, ctx)
		if err != nil {
			return err
		}
		key.Enabled = true
		if err := op.APIKeyUpdate(&key, ctx); err != nil {
			return err
		}
		recordBatchAudit(c, observability.AuditActionEnable, observability.ResourceTypeAPIKey, id, key.Name)
		return nil
	})
	resp.Success(c, result)
}

// batchDisableAPIKey 批量禁用 API Key
func batchDisableAPIKey(c *gin.Context) {
	req, ok := parseBatchRequest(c)
	if !ok {
		return
	}
	result := runBatchOperation(c.Request.Context(), req, func(ctx context.Context, id int) error {
		key, err := op.APIKeyGet(id, ctx)
		if err != nil {
			return err
		}
		key.Enabled = false
		if err := op.APIKeyUpdate(&key, ctx); err != nil {
			return err
		}
		recordBatchAudit(c, observability.AuditActionDisable, observability.ResourceTypeAPIKey, id, key.Name)
		return nil
	})
	resp.Success(c, result)
}

// batchDeleteAPIKey 批量删除 API Key
func batchDeleteAPIKey(c *gin.Context) {
	req, ok := parseBatchRequest(c)
	if !ok {
		return
	}
	result := runBatchOperation(c.Request.Context(), req, func(ctx context.Context, id int) error {
		var keyName string
		if key, err := op.APIKeyGet(id, ctx); err == nil {
			keyName = key.Name
		}
		if err := op.APIKeyDelete(id, ctx); err != nil {
			return err
		}
		recordBatchAudit(c, observability.AuditActionDelete, observability.ResourceTypeAPIKey, id, keyName)
		return nil
	})
	resp.Success(c, result)
}

func getStatsAPIKeyById(c *gin.Context) {
	id := c.GetInt("api_key_id")
	stats := op.StatsAPIKeyGet(id)
	info, err := op.APIKeyGet(id, c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to get api key")
		return
	}
	models, err := op.GroupListModel(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to list models")
		return
	}
	var modelsString string
	if info.SupportedModels == "" {
		modelsString = strings.Join(models, ", ")
	} else {
		supportedModels := lo.Map(strings.Split(info.SupportedModels, ","), func(s string, _ int) string {
			return strings.TrimSpace(s)
		})
		models = lo.Filter(models, func(m string, _ int) bool {
			return lo.Contains(supportedModels, m)
		})
		modelsString = strings.Join(models, ", ")
	}
	info.SupportedModels = modelsString
	resp.Success(c, map[string]any{
		"stats": stats,
		"info":  info,
	})
}

func loginAPIKey(c *gin.Context) {
	id := c.GetInt("api_key_id")
	info, err := op.APIKeyGet(id, c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to get api key")
		return
	}
	resp.Success(c, model.APIKeyAuthStatus{
		OK:              true,
		APIKeyID:        info.ID,
		Name:            info.Name,
		Enabled:         info.Enabled,
		ExpireAt:        info.ExpireAt,
		SupportedModels: info.SupportedModels,
		AuthMode:        "api_key",
	})
}
