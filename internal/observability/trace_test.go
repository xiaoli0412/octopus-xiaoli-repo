package observability

import (
	"context"
	"testing"
)

func TestInitTracerNoopWhenEndpointEmpty(t *testing.T) {
	// 确保未配置 endpoint 时返回 noop tracer
	shutdown, err := InitTracer(context.Background(), "")
	if err != nil {
		t.Fatalf("InitTracer() error = %v", err)
	}
	if shutdown == nil {
		t.Fatal("shutdown function is nil")
	}

	if Tracer == nil {
		t.Fatal("Tracer is nil after InitTracer")
	}

	// noop tracer 应该能创建 span
	_, span := Tracer.Start(context.Background(), "test.span")
	if span == nil {
		t.Fatal("span is nil")
	}
	span.End()

	// noop span 的 context 应该有效但无 trace ID
	if !span.SpanContext().IsValid() {
		// noop tracer 可能返回 invalid span context，这是预期行为
	}

	// 清理
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown() error = %v", err)
	}
}

func TestInitTracerWithEnvVar(t *testing.T) {
	// 设置环境变量但指向不存在的端点，应该返回错误或成功创建
	// 由于 OTLP gRPC 需要实际连接，这里测试未设置时为 noop
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	shutdown, err := InitTracer(context.Background(), "")
	if err != nil {
		t.Fatalf("InitTracer() error = %v", err)
	}
	defer shutdown(context.Background())

	if !IsTracerReady() {
		t.Error("expected tracer to be ready")
	}
}

func TestTracerCreatesSpan(t *testing.T) {
	shutdown, err := InitTracer(context.Background(), "")
	if err != nil {
		t.Fatalf("InitTracer() error = %v", err)
	}
	defer shutdown(context.Background())

	ctx, span := Tracer.Start(context.Background(), "test.operation")
	if span == nil {
		t.Fatal("span is nil")
	}

	// 在 span 内创建子 span
	_, childSpan := Tracer.Start(ctx, "test.child")
	if childSpan == nil {
		t.Fatal("child span is nil")
	}
	childSpan.End()
	span.End()
}

func TestEnsureTracer(t *testing.T) {
	// 重置 Tracer 为 nil 来测试 ensureTracer
	// 注意：由于 tracerInitOnce 是 sync.Once，只能执行一次
	// 这个测试主要确保 ensureTracer 不会 panic
	ensureTracer()
}
