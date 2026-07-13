package middleware

import (
	"crypto/subtle"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
)

// MetricsAuthMiddleware authenticates requests to the /metrics endpoint.
//
// When both metrics_auth_token and metrics_ip_allowlist are empty the
// middleware is a no-op, preserving backward compatibility. When either
// is configured the corresponding check is enforced; if both are
// configured the request must pass both checks. Failures produce 403.
func MetricsAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := op.SettingGetString(model.SettingKeyMetricsAuthToken)
		token = strings.TrimSpace(token)
		allowlistRaw, _ := op.SettingGetString(model.SettingKeyMetricsIPAllowlist)
		allowlistRaw = strings.TrimSpace(allowlistRaw)

		if token == "" && allowlistRaw == "" {
			c.Next()
			return
		}

		if token != "" && !metricsTokenOK(c, token) {
			resp.Error(c, http.StatusForbidden, "metrics token authentication failed")
			c.Abort()
			return
		}

		if allowlistRaw != "" && !metricsIPAllowed(c, allowlistRaw) {
			resp.Error(c, http.StatusForbidden, "metrics IP not allowed")
			c.Abort()
			return
		}

		c.Next()
	}
}

func metricsTokenOK(c *gin.Context, expected string) bool {
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		if token, ok := bearerTokenFromHeader(authHeader); ok {
			return subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
		}
	}
	if q := strings.TrimSpace(c.Query("token")); q != "" {
		return subtle.ConstantTimeCompare([]byte(q), []byte(expected)) == 1
	}
	return false
}

func metricsIPAllowed(c *gin.Context, allowlistRaw string) bool {
	clientIP := net.ParseIP(c.ClientIP())
	if clientIP == nil {
		return false
	}
	for _, entry := range splitMetricsAllowlist(allowlistRaw) {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			_, cidr, err := net.ParseCIDR(entry)
			if err != nil {
				continue
			}
			if cidr.Contains(clientIP) {
				return true
			}
			continue
		}
		if ip := net.ParseIP(entry); ip != nil && ip.Equal(clientIP) {
			return true
		}
	}
	return false
}

func splitMetricsAllowlist(value string) []string {
	normalized := strings.ReplaceAll(value, "\n", ",")
	normalized = strings.ReplaceAll(normalized, ";", ",")
	parts := strings.Split(normalized, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
