package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/observability"
)

// metricsContextKey 用于在 request context 中注入 metrics registry
type metricsContextKey struct{}

// excludedMetricsPaths 这些端点不收集 RED 指标
var excludedMetricsPaths = map[string]struct{}{
	"/metrics":  {},
	"/healthz":  {},
	"/readyz":   {},
}

// MetricsMiddleware 收集 RED 指标（Rate, Errors, Duration）
func MetricsMiddleware(metrics *observability.MetricsRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 排除监控/健康检查端点
		if _, excluded := excludedMetricsPaths[path]; excluded {
			c.Next()
			return
		}

		start := time.Now()

		// 将 metrics registry 注入 request context
		if metrics != nil {
			ctx := observability.WithMetricsRegistry(c.Request.Context(), metrics)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()

		if metrics == nil {
			return
		}

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		// 归一化路径，避免高基数（将 :id 替换为 :id 占位符）
		normalizedPath := normalizePath(path)

		metrics.HTTPRequestsTotal.WithLabelValues(method, normalizedPath, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, normalizedPath).Observe(duration)
	}
}

// normalizePath 将路径中的数字 ID 替换为 :id，降低 Prometheus 指标基数
func normalizePath(path string) string {
	if len(path) == 0 {
		return path
	}
	// 简单归一化：对 /api/v1/xxx/数字 形式替换最后一段数字为 :id
	runes := []rune(path)
	n := len(runes)
	// 找到最后一个 / 之后的部分
	lastSlash := -1
	for i := n - 1; i >= 0; i-- {
		if runes[i] == '/' {
			lastSlash = i
			break
		}
	}
	if lastSlash < 0 || lastSlash == n-1 {
		return path
	}
	lastSeg := runes[lastSlash+1:]
	allDigit := true
	for _, r := range lastSeg {
		if r < '0' || r > '9' {
			allDigit = false
			break
		}
	}
	if allDigit {
		return string(runes[:lastSlash+1]) + ":id"
	}
	return path
}
