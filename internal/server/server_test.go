package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewHTTPServerUsesSecurityTimeouts(t *testing.T) {
    srv := newHTTPServer("127.0.0.1:1088", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	if srv.ReadHeaderTimeout != httpReadHeaderTimeout {
		t.Fatalf("ReadHeaderTimeout = %v, want %v", srv.ReadHeaderTimeout, httpReadHeaderTimeout)
	}
	if srv.ReadTimeout != httpReadTimeout {
		t.Fatalf("ReadTimeout = %v, want %v", srv.ReadTimeout, httpReadTimeout)
	}
	if srv.IdleTimeout != httpIdleTimeout {
		t.Fatalf("IdleTimeout = %v, want %v", srv.IdleTimeout, httpIdleTimeout)
	}
	if srv.Handler == nil {
		t.Fatalf("Handler = nil, want non-nil")
	}
}

func TestNewEngineDoesNotTrustForwardedClientIPHeadersByDefault(t *testing.T) {
	engine, err := newEngine()
	if err != nil {
		t.Fatalf("newEngine() error = %v", err)
	}
	engine.GET("/client-ip", func(c *gin.Context) {
		c.String(http.StatusOK, c.ClientIP())
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.RemoteAddr = "203.0.113.10:4567"
	request.Header.Set("X-Forwarded-For", "198.51.100.77")
	request.URL.Path = "/client-ip"

	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("client-ip status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Body.String(); got != "203.0.113.10" {
		t.Fatalf("ClientIP() = %q, want remote address when proxies are untrusted", got)
	}
}
