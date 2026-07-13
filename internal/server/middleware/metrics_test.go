package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/observability"
)

func newTestMetricsRegistry() *observability.MetricsRegistry {
	return &observability.MetricsRegistry{
		HTTPRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "octopus_http_requests_total_test",
			Help: "Total HTTP requests test.",
		}, []string{"method", "path", "status"}),
		HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "octopus_http_request_duration_seconds_test",
			Help:    "HTTP request duration test.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path"}),
	}
}

func TestMetricsMiddlewareCollectsMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mr := newTestMetricsRegistry()

	router := gin.New()
	router.Use(MetricsMiddleware(mr))
	router.GET("/api/v1/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %v, want %v", recorder.Code, http.StatusOK)
	}

	// 验证指标被收集
	count := testutil.ToFloat64(mr.HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/test", "200"))
	if count != 1 {
		t.Errorf("HTTPRequestsTotal = %v, want 1", count)
	}
}

func TestMetricsMiddlewareExcludesHealthz(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mr := newTestMetricsRegistry()

	router := gin.New()
	router.Use(MetricsMiddleware(mr))
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// /healthz 不应记录指标
	count := testutil.ToFloat64(mr.HTTPRequestsTotal.WithLabelValues("GET", "/healthz", "200"))
	if count != 0 {
		t.Errorf("HTTPRequestsTotal for /healthz = %v, want 0 (excluded)", count)
	}
}

func TestMetricsMiddlewareExcludesMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mr := newTestMetricsRegistry()

	router := gin.New()
	router.Use(MetricsMiddleware(mr))
	router.GET("/metrics", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// /metrics 不应记录指标
	count := testutil.ToFloat64(mr.HTTPRequestsTotal.WithLabelValues("GET", "/metrics", "200"))
	if count != 0 {
		t.Errorf("HTTPRequestsTotal for /metrics = %v, want 0 (excluded)", count)
	}
}

func TestMetricsMiddlewareNilRegistry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(MetricsMiddleware(nil))
	router.GET("/api/v1/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %v, want %v", recorder.Code, http.StatusOK)
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/api/v1/channel/list", "/api/v1/channel/list"},
		{"/api/v1/audit/123", "/api/v1/audit/:id"},
		{"/api/v1/audit/abc", "/api/v1/audit/abc"},
		{"/healthz", "/healthz"},
		{"", ""},
		{"/", "/"},
	}

	for _, tt := range tests {
		got := normalizePath(tt.input)
		if got != tt.expected {
			t.Errorf("normalizePath(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMetricsMiddlewareRecordsDuration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mr := newTestMetricsRegistry()

	router := gin.New()
	router.Use(MetricsMiddleware(mr))
	router.GET("/api/v1/slow", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/slow", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// 验证 duration histogram 被观测
	count := testutil.CollectAndCount(mr.HTTPRequestDuration)
	if count != 1 {
		t.Errorf("HTTPRequestDuration metric count = %v, want 1", count)
	}
}
