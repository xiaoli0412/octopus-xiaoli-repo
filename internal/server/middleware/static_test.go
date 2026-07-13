package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func TestStaticFallsThroughWhenFileMissing(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(static("/", http.FS(fstest.MapFS{})))
	r.GET("/dashboard", func(c *gin.Context) {
		c.String(http.StatusOK, "fallback")
	})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "fallback" {
		t.Fatalf("body = %q, want %q", body, "fallback")
	}
}

func TestStaticCacheControlByPath(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(static("/", http.FS(fstest.MapFS{
		"index.html":          &fstest.MapFile{Data: []byte("index")},
		"_next/static/app.js": &fstest.MapFile{Data: []byte("console.log('ok')")},
	})))

	t.Run("next static assets are immutable", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/_next/static/app.js", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
			t.Fatalf("Cache-Control = %q", got)
		}
	})

	t.Run("html shell is not immutable", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
			t.Fatalf("Cache-Control = %q", got)
		}
	})
}

func TestStaticSupportsPrefixedMounts(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(static("/console", http.FS(fstest.MapFS{
		"_next/static/app.js": &fstest.MapFile{Data: []byte("console.log('ok')")},
	})))
	r.GET("/console/fallback", func(c *gin.Context) {
		c.String(http.StatusOK, "fallback")
	})

	req := httptest.NewRequest(http.MethodGet, "/console/_next/static/app.js", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "console.log('ok')" {
		t.Fatalf("body = %q, want %q", body, "console.log('ok')")
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("Cache-Control = %q", got)
	}
}

func TestStaticSPAFallback(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(static("/", http.FS(fstest.MapFS{
		"index.html":          &fstest.MapFile{Data: []byte("<!doctype html><html><body>app</body></html>")},
		"_next/static/app.js": &fstest.MapFile{Data: []byte("console.log('ok')")},
	})))
	r.GET("/api/something", func(c *gin.Context) {
		c.String(http.StatusOK, "api-handler")
	})

	t.Run("unknown route serves index.html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/settings", nil)
		req.Header.Set("Accept", "text/html,application/xhtml+xml")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if body := rec.Body.String(); body != "<!doctype html><html><body>app</body></html>" {
			t.Fatalf("body = %q, want index.html content", body)
		}
		if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
			t.Fatalf("Content-Type = %q, want %q", got, "text/html; charset=utf-8")
		}
		if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
			t.Fatalf("Cache-Control = %q, want %q", got, "no-cache")
		}
	})

	t.Run("api route falls through", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/something", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if body := rec.Body.String(); body != "api-handler" {
			t.Fatalf("body = %q, want %q", body, "api-handler")
		}
	})
}
