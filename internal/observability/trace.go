package observability

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Tracer 全局变量供业务代码获取 span
var Tracer trace.Tracer

var (
	tracerInitOnce sync.Once
	tracerReady    bool
)

// InitTracer 初始化 TracerProvider。
// 当 OTEL_EXPORTER_OTLP_ENDPOINT 环境变量未配置时返回 noop tracer（不影响性能）。
// endpoint 参数可显式指定 OTLP gRPC 导出地址；为空时从环境变量读取。
// 返回的 shutdown 函数用于优雅关闭 TracerProvider。
func InitTracer(ctx context.Context, endpoint string) (func(context.Context) error, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	}

	// 未配置 endpoint 时使用 noop tracer
	if endpoint == "" {
		noopTP := trace.NewNoopTracerProvider()
		otel.SetTracerProvider(noopTP)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
		Tracer = noopTP.Tracer("octopus")
		tracerReady = true
		return func(context.Context) error { return nil }, nil
	}

	// 创建 OTLP gRPC exporter
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("octopus"),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	Tracer = tp.Tracer("octopus")
	tracerReady = true

	shutdown := func(ctx context.Context) error {
		var errs []error
		if tp != nil {
			if err := tp.Shutdown(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		if exporter != nil {
			if err := exporter.Shutdown(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	}

	return shutdown, nil
}

// IsTracerReady 返回 tracer 是否已初始化
func IsTracerReady() bool {
	return tracerReady
}

// ensureTracer 确保 tracer 至少被初始化为 noop，避免 nil panic
func ensureTracer() {
	tracerInitOnce.Do(func() {
		if Tracer == nil {
			noopTP := trace.NewNoopTracerProvider()
			Tracer = noopTP.Tracer("octopus")
		}
	})
}

// init 确保包加载时 Tracer 至少为 noop，防止未调用 InitTracer 时出现 nil panic
func init() {
	ensureTracer()
}
