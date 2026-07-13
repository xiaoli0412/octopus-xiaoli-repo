package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TraceMiddleware 从 W3C TraceContext header 解析父 span context 并创建 root span
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 request header 中提取父 span context
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		tracer := otel.GetTracerProvider().Tracer("octopus")

		route := c.Request.URL.Path
		if route == "" {
			route = "/"
		}

		spanName := fmt.Sprintf("%s %s", c.Request.Method, route)
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.route", c.Request.URL.Path),
				attribute.String("http.user_agent", c.Request.Header.Get("User-Agent")),
			),
		)

		defer span.End()

		// 将带 span 的 context 注入 request
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// 设置响应状态码属性
		status := c.Writer.Status()
		span.SetAttributes(attribute.Int("http.status_code", status))

		// 记录错误（4xx/5xx）
		if status >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	}
}
