package middleware

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/gin-gonic/gin"
)

func TestCorsDeniesLoopbackOutsideDebug(t *testing.T) {
	setupCORSTestState(t)

	router := gin.New()
	router.Use(Cors())
	router.OPTIONS("/api/v1/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := performCORSPreflight(router, "http://localhost:3000")
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

func TestCorsAllowsLoopbackInDebug(t *testing.T) {
	t.Setenv("OCTOPUS_DEBUG", "true")
	setupCORSTestState(t)

	router := gin.New()
	router.Use(Cors())
	router.OPTIONS("/api/v1/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for _, origin := range []string{
		"http://127.0.0.1:3000",
		"http://[::1]:3000",
	} {
		recorder := performCORSPreflight(router, origin)
		if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != origin {
			t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, origin)
		}
	}
}

func TestCorsAllowsWildcardAllowlist(t *testing.T) {
	setupCORSTestState(t)
	if err := op.SettingSetString(model.SettingKeyCORSAllowOrigins, "*"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	router := gin.New()
	router.Use(Cors())
	router.OPTIONS("/api/v1/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	origin := "https://evil.example"
	recorder := performCORSPreflight(router, origin)
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, origin)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("Access-Control-Allow-Credentials = %q, want empty", got)
	}
}

func TestCorsAllowsConfiguredOriginHost(t *testing.T) {
	setupCORSTestState(t)
	if err := op.SettingSetString(model.SettingKeyCORSAllowOrigins, "example.com"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	router := gin.New()
	router.Use(Cors())
	router.OPTIONS("/api/v1/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	origin := "https://example.com"
	recorder := performCORSPreflight(router, origin)
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, origin)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("Access-Control-Allow-Credentials = %q, want empty", got)
	}
}

func TestCorsAllowsConfiguredLoopbackOutsideDebug(t *testing.T) {
	setupCORSTestState(t)
	if err := op.SettingSetString(model.SettingKeyCORSAllowOrigins, "127.0.0.1"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	router := gin.New()
	router.Use(Cors())
	router.OPTIONS("/api/v1/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	origin := "http://127.0.0.1:3000"
	recorder := performCORSPreflight(router, origin)
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, origin)
	}
}

func setupCORSTestState(t *testing.T) {
	t.Helper()
	t.Setenv(op.BootstrapAdminUsernameEnv(), op.BootstrapAdminDefaultUsername())
	t.Setenv(op.BootstrapAdminPasswordEnv(), "admin")

	if db.GetDB() != nil {
		if err := db.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}

	dbPath := filepath.Join(t.TempDir(), "octopus-cors.db")
	if err := db.InitDB("sqlite", dbPath, false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	if err := op.InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}

	t.Cleanup(func() {
		if db.GetDB() != nil {
			if err := db.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}
		}
	})
}

func performCORSPreflight(router *gin.Engine, origin string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/ping", nil)
	req.Host = "octopus.test"
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	router.ServeHTTP(recorder, req)
	return recorder
}
