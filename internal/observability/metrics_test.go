package observability

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// newTestMetricsRegistry 创建一个不使用全局 registry 的 MetricsRegistry，用于测试
func newTestMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{
		RelayRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "octopus_relay_requests_total",
			Help: "Total number of relay requests processed.",
		}, []string{"channel_id", "model", "provider_type", "status"}),
		RelayDurationSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "octopus_relay_duration_seconds",
			Help:    "Duration of relay requests in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"channel_id", "model", "provider_type"}),
		ChannelHealth: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "octopus_channel_health",
			Help: "Channel health status.",
		}, []string{"channel_id"}),
		CircuitBreakerState: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "octopus_circuit_breaker_state",
			Help: "Circuit breaker state.",
		}, []string{"channel_id", "model"}),
		TokenThroughputTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "octopus_token_throughput_total",
			Help: "Total tokens processed.",
		}, []string{"channel_id", "model", "type"}),
		HTTPClientPoolIdle: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "octopus_http_client_pool_idle",
			Help: "Number of idle HTTP client connections.",
		}, []string{"client_type"}),
		HTTPRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "octopus_http_requests_total",
			Help: "Total HTTP requests.",
		}, []string{"method", "path", "status"}),
		HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "octopus_http_request_duration_seconds",
			Help:    "HTTP request duration.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path"}),
	}
}

func TestMetricsRegistryFields(t *testing.T) {
	mr := newTestMetricsRegistry()
	if mr == nil {
		t.Fatal("MetricsRegistry is nil")
	}
	if mr.RelayRequestsTotal == nil {
		t.Error("RelayRequestsTotal is nil")
	}
	if mr.RelayDurationSeconds == nil {
		t.Error("RelayDurationSeconds is nil")
	}
	if mr.ChannelHealth == nil {
		t.Error("ChannelHealth is nil")
	}
	if mr.CircuitBreakerState == nil {
		t.Error("CircuitBreakerState is nil")
	}
	if mr.TokenThroughputTotal == nil {
		t.Error("TokenThroughputTotal is nil")
	}
	if mr.HTTPClientPoolIdle == nil {
		t.Error("HTTPClientPoolIdle is nil")
	}
	if mr.HTTPRequestsTotal == nil {
		t.Error("HTTPRequestsTotal is nil")
	}
	if mr.HTTPRequestDuration == nil {
		t.Error("HTTPRequestDuration is nil")
	}
}

func TestRelayRequestsTotalCounter(t *testing.T) {
	mr := newTestMetricsRegistry()

	mr.RelayRequestsTotal.WithLabelValues("1", "gpt-4", "openai", "success").Inc()
	mr.RelayRequestsTotal.WithLabelValues("1", "gpt-4", "openai", "success").Inc()
	mr.RelayRequestsTotal.WithLabelValues("2", "claude-3", "anthropic", "failed").Inc()

	if got := testutil.ToFloat64(mr.RelayRequestsTotal.WithLabelValues("1", "gpt-4", "openai", "success")); got != 2 {
		t.Errorf("RelayRequestsTotal for channel 1 = %v, want 2", got)
	}
	if got := testutil.ToFloat64(mr.RelayRequestsTotal.WithLabelValues("2", "claude-3", "anthropic", "failed")); got != 1 {
		t.Errorf("RelayRequestsTotal for channel 2 = %v, want 1", got)
	}
}

func TestChannelHealthGauge(t *testing.T) {
	mr := newTestMetricsRegistry()

	mr.ChannelHealth.WithLabelValues("1").Set(1)
	mr.ChannelHealth.WithLabelValues("2").Set(0)

	if got := testutil.ToFloat64(mr.ChannelHealth.WithLabelValues("1")); got != 1 {
		t.Errorf("ChannelHealth for channel 1 = %v, want 1", got)
	}
	if got := testutil.ToFloat64(mr.ChannelHealth.WithLabelValues("2")); got != 0 {
		t.Errorf("ChannelHealth for channel 2 = %v, want 0", got)
	}
}

func TestTokenThroughputTotal(t *testing.T) {
	mr := newTestMetricsRegistry()

	mr.TokenThroughputTotal.WithLabelValues("1", "gpt-4", "prompt").Add(100)
	mr.TokenThroughputTotal.WithLabelValues("1", "gpt-4", "completion").Add(50)

	if got := testutil.ToFloat64(mr.TokenThroughputTotal.WithLabelValues("1", "gpt-4", "prompt")); got != 100 {
		t.Errorf("TokenThroughputTotal prompt = %v, want 100", got)
	}
	if got := testutil.ToFloat64(mr.TokenThroughputTotal.WithLabelValues("1", "gpt-4", "completion")); got != 50 {
		t.Errorf("TokenThroughputTotal completion = %v, want 50", got)
	}
}

func TestHTTPMetricsCollection(t *testing.T) {
	mr := newTestMetricsRegistry()

	mr.HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/channel/list", "200").Inc()
	mr.HTTPRequestDuration.WithLabelValues("GET", "/api/v1/channel/list").Observe(0.5)

	if got := testutil.ToFloat64(mr.HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/channel/list", "200")); got != 1 {
		t.Errorf("HTTPRequestsTotal = %v, want 1", got)
	}
}

func TestMetricsRegistryFromContext(t *testing.T) {
	ctx := WithMetricsRegistry(nil, nil)
	if got := MetricsRegistryFromContext(ctx); got != nil {
		t.Error("expected nil for nil registry")
	}

	mr := newTestMetricsRegistry()
	ctx = WithMetricsRegistry(nil, mr)
	if got := MetricsRegistryFromContext(ctx); got != mr {
		t.Error("expected same MetricsRegistry from context")
	}

	if got := MetricsRegistryFromContext(nil); got != nil {
		t.Error("expected nil for nil context")
	}
}

func TestRelayDurationHistogram(t *testing.T) {
	mr := newTestMetricsRegistry()

	mr.RelayDurationSeconds.WithLabelValues("1", "gpt-4", "openai").Observe(0.1)
	mr.RelayDurationSeconds.WithLabelValues("1", "gpt-4", "openai").Observe(0.5)
	mr.RelayDurationSeconds.WithLabelValues("1", "gpt-4", "openai").Observe(1.0)

	count := testutil.CollectAndCount(mr.RelayDurationSeconds)
	if count != 1 {
		t.Errorf("expected 1 metric series, got %d", count)
	}
}

func TestCircuitBreakerStateGauge(t *testing.T) {
	mr := newTestMetricsRegistry()

	mr.CircuitBreakerState.WithLabelValues("1", "gpt-4").Set(0) // closed
	mr.CircuitBreakerState.WithLabelValues("2", "claude-3").Set(1) // open

	if got := testutil.ToFloat64(mr.CircuitBreakerState.WithLabelValues("1", "gpt-4")); got != 0 {
		t.Errorf("CircuitBreakerState channel 1 = %v, want 0", got)
	}
	if got := testutil.ToFloat64(mr.CircuitBreakerState.WithLabelValues("2", "claude-3")); got != 1 {
		t.Errorf("CircuitBreakerState channel 2 = %v, want 1", got)
	}
}

func TestInitMetricsReturnsRegistry(t *testing.T) {
	// InitMetrics uses the global registry; calling it once should succeed.
	// Note: if called multiple times in the same process, MustRegister panics
	// due to duplicate registration. We test it once here.
	mr := InitMetrics()
	if mr == nil {
		t.Fatal("InitMetrics returned nil")
	}
	if mr.RelayRequestsTotal == nil {
		t.Error("RelayRequestsTotal is nil after InitMetrics")
	}
}

func TestDefaultRegistry(t *testing.T) {
	// After InitMetrics, DefaultRegistry should be non-nil
	if got := DefaultRegistry(); got == nil {
		t.Error("DefaultRegistry() returned nil after InitMetrics")
	}
}

// 确保指标名称符合预期
func TestMetricNames(t *testing.T) {
	mr := newTestMetricsRegistry()

	// 通过 Desc() 检查指标名称
	desc := mr.RelayRequestsTotal.WithLabelValues("1", "gpt-4", "openai", "success").Desc().String()
	if !strings.Contains(desc, "octopus_relay_requests_total") {
		t.Errorf("metric desc = %q, want to contain octopus_relay_requests_total", desc)
	}
}
