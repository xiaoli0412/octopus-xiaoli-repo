package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

func setupMetricsAuthTestState(t *testing.T) {
	t.Helper()
	setupCORSTestState(t)
}

func newMetricsAuthRouter() *gin.Engine {
	r := gin.New()
	r.GET("/metrics", MetricsAuthMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func TestMetricsAuthNoConfigAllowsAccess(t *testing.T) {
	setupMetricsAuthTestState(t)

	router := newMetricsAuthRouter()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (no auth configured, should pass through)", rec.Code, http.StatusOK)
	}
}

func TestMetricsAuthTokenRejectsMissingToken(t *testing.T) {
	setupMetricsAuthTestState(t)
	if err := op.SettingSetString(model.SettingKeyMetricsAuthToken, "scraper-token"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	router := newMetricsAuthRouter()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestMetricsAuthTokenAcceptsBearerHeader(t *testing.T) {
	setupMetricsAuthTestState(t)
	if err := op.SettingSetString(model.SettingKeyMetricsAuthToken, "scraper-token"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	router := newMetricsAuthRouter()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer scraper-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestMetricsAuthTokenAcceptsQueryParam(t *testing.T) {
	setupMetricsAuthTestState(t)
	if err := op.SettingSetString(model.SettingKeyMetricsAuthToken, "scraper-token"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	router := newMetricsAuthRouter()
	req := httptest.NewRequest(http.MethodGet, "/metrics?token=scraper-token", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestMetricsAuthTokenRejectsWrongToken(t *testing.T) {
	setupMetricsAuthTestState(t)
	if err := op.SettingSetString(model.SettingKeyMetricsAuthToken, "scraper-token"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	router := newMetricsAuthRouter()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestMetricsIPAllowlistAcceptsAllowedIP(t *testing.T) {
	setupMetricsAuthTestState(t)
	if err := op.SettingSetString(model.SettingKeyMetricsIPAllowlist, "192.0.2.0/24"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	router := newMetricsAuthRouter()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.RemoteAddr = "192.0.2.10:1234"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (IP in allowlist)", rec.Code, http.StatusOK)
	}
}

func TestMetricsIPAllowlistRejectsNonAllowedIP(t *testing.T) {
	setupMetricsAuthTestState(t)
	if err := op.SettingSetString(model.SettingKeyMetricsIPAllowlist, "192.0.2.0/24"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	router := newMetricsAuthRouter()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (IP not in allowlist)", rec.Code, http.StatusForbidden)
	}
}

func TestMetricsIPAllowlistAcceptsBareIPEntry(t *testing.T) {
	setupMetricsAuthTestState(t)
	if err := op.SettingSetString(model.SettingKeyMetricsIPAllowlist, "127.0.0.1"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	router := newMetricsAuthRouter()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.RemoteAddr = "127.0.0.1:5678"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (bare IP match)", rec.Code, http.StatusOK)
	}
}

func TestMetricsBothTokenAndIPRequired(t *testing.T) {
	setupMetricsAuthTestState(t)
	if err := op.SettingSetString(model.SettingKeyMetricsAuthToken, "scraper-token"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}
	if err := op.SettingSetString(model.SettingKeyMetricsIPAllowlist, "192.0.2.0/24"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	router := newMetricsAuthRouter()

	// Correct token but wrong IP → 403
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer scraper-token")
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("correct token + wrong IP: status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	// Correct IP but wrong token → 403
	req2 := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req2.Header.Set("Authorization", "Bearer wrong-token")
	req2.RemoteAddr = "192.0.2.10:1234"
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusForbidden {
		t.Fatalf("wrong token + correct IP: status = %d, want %d", rec2.Code, http.StatusForbidden)
	}

	// Both correct → 200
	req3 := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req3.Header.Set("Authorization", "Bearer scraper-token")
	req3.RemoteAddr = "192.0.2.10:1234"
	rec3 := httptest.NewRecorder()
	router.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("correct token + correct IP: status = %d, want %d", rec3.Code, http.StatusOK)
	}
}
