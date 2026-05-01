package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/auth"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/gin-gonic/gin"
)

func bearerTokenFromHeader(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	const bearerPrefix = "Bearer "
	if len(value) < len(bearerPrefix) || !strings.EqualFold(value[:len(bearerPrefix)], bearerPrefix) {
		return "", false
	}
	token := strings.TrimSpace(value[len(bearerPrefix):])
	if token == "" {
		return "", false
	}
	return token, true
}

func allowRequestDuringForcedPasswordChange(c *gin.Context) bool {
	if c == nil || c.Request == nil || !op.UserMustChangePassword() {
		return true
	}

	path := strings.TrimSpace(c.Request.URL.Path)
	switch {
	case c.Request.Method == http.MethodGet && path == "/api/v1/user/status":
		return true
	case c.Request.Method == http.MethodPost && path == "/api/v1/user/force-change-password":
		return true
	default:
		return false
	}
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerTokenFromHeader(c.GetHeader("Authorization"))
		if !ok {
			resp.Error(c, http.StatusBadRequest, resp.ErrBadRequest)
			c.Abort()
			return
		}
		if !auth.VerifyJWTToken(token) {
			resp.Error(c, http.StatusUnauthorized, resp.ErrUnauthorized)
			c.Abort()
			return
		}
		if !allowRequestDuringForcedPasswordChange(c) {
			resp.Error(c, http.StatusForbidden, "password change required before accessing other management routes")
			c.Abort()
			return
		}
		c.Next()
	}
}

func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var apiKey string
		var requestType string

		if key := c.Request.Header.Get("x-api-key"); key != "" {
			apiKey = key
			requestType = "anthropic"
		} else if authHeader := c.Request.Header.Get("Authorization"); authHeader != "" {
			var ok bool
			apiKey, ok = bearerTokenFromHeader(authHeader)
			if !ok {
				resp.Error(c, http.StatusUnauthorized, resp.ErrUnauthorized)
				c.Abort()
				return
			}
			requestType = "openai"
		}

		if apiKey == "" {
			resp.Error(c, http.StatusUnauthorized, resp.ErrUnauthorized)
			c.Abort()
			return
		}

		if !strings.HasPrefix(apiKey, "sk-"+conf.APP_NAME+"-") {
			resp.Error(c, http.StatusUnauthorized, resp.ErrUnauthorized)
			c.Abort()
			return
		}
		apiKeyObj, err := op.APIKeyGetByAPIKey(apiKey, c.Request.Context())
		if err != nil {
			resp.Error(c, http.StatusUnauthorized, resp.ErrUnauthorized)
			c.Abort()
			return
		}
		if !apiKeyObj.Enabled {
			resp.Error(c, http.StatusUnauthorized, "API key is disabled")
			c.Abort()
			return
		}
		if apiKeyObj.ExpireAt > 0 && apiKeyObj.ExpireAt < time.Now().Unix() {
			resp.Error(c, http.StatusUnauthorized, "API key has expired")
			c.Abort()
			return
		}
		statsAPIKey := op.StatsAPIKeyGet(apiKeyObj.ID)
		if apiKeyObj.MaxCost > 0 && apiKeyObj.MaxCost <= statsAPIKey.StatsMetrics.OutputCost+statsAPIKey.StatsMetrics.InputCost {
			resp.Error(c, http.StatusUnauthorized, "API key has reached the max cost")
			c.Abort()
			return
		}
		c.Set("request_type", requestType)
		c.Set("supported_models", apiKeyObj.SupportedModels)
		c.Set("api_key_id", apiKeyObj.ID)
		c.Next()
	}
}
