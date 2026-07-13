package observability

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// metricsRegistryContextKey 用于在 context 中传递 MetricsRegistry
type metricsRegistryContextKey struct{}

// WithMetricsRegistry 将 MetricsRegistry 注入到 context 中
func WithMetricsRegistry(ctx context.Context, mr *MetricsRegistry) context.Context {
	if mr == nil {
		return ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, metricsRegistryContextKey{}, mr)
}

// MetricsRegistryFromContext 从 context 中获取 MetricsRegistry，不存在时返回 nil
func MetricsRegistryFromContext(ctx context.Context) *MetricsRegistry {
	if ctx == nil {
		return nil
	}
	v, ok := ctx.Value(metricsRegistryContextKey{}).(*MetricsRegistry)
	if !ok {
		return nil
	}
	return v
}

// MetricsRegistry 封装所有 Prometheus 指标供外部访问
type MetricsRegistry struct {
	RelayRequestsTotal    *prometheus.CounterVec
	RelayDurationSeconds  *prometheus.HistogramVec
	ChannelHealth         *prometheus.GaugeVec
	CircuitBreakerState   *prometheus.GaugeVec
	TokenThroughputTotal  *prometheus.CounterVec
	HTTPClientPoolIdle    *prometheus.GaugeVec
	HTTPRequestsTotal     *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
}

var defaultRegistry *MetricsRegistry

// InitMetrics 注册所有指标到默认 registry，并返回封装后的 MetricsRegistry
func InitMetrics() *MetricsRegistry {
	mr := &MetricsRegistry{
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
			Help: "Channel health status (1=healthy, 0=unhealthy).",
		}, []string{"channel_id"}),
		CircuitBreakerState: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "octopus_circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half-open).",
		}, []string{"channel_id", "model"}),
		TokenThroughputTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "octopus_token_throughput_total",
			Help: "Total tokens processed by channel and model.",
		}, []string{"channel_id", "model", "type"}),
		HTTPClientPoolIdle: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "octopus_http_client_pool_idle",
			Help: "Number of idle HTTP client connections by client type.",
		}, []string{"client_type"}),
		HTTPRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "octopus_http_requests_total",
			Help: "Total number of HTTP requests processed by the server.",
		}, []string{"method", "path", "status"}),
		HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "octopus_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path"}),
	}

	prometheus.MustRegister(
		mr.RelayRequestsTotal,
		mr.RelayDurationSeconds,
		mr.ChannelHealth,
		mr.CircuitBreakerState,
		mr.TokenThroughputTotal,
		mr.HTTPClientPoolIdle,
		mr.HTTPRequestsTotal,
		mr.HTTPRequestDuration,
	)

	defaultRegistry = mr
	return mr
}

// DefaultRegistry 返回已初始化的全局 MetricsRegistry，未初始化时返回 nil
func DefaultRegistry() *MetricsRegistry {
	return defaultRegistry
}
