package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireJSONRejectsOversizedManagementBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalLimit := maxManagementJSONBodyBytes
	maxManagementJSONBodyBytes = 32
	t.Cleanup(func() {
		maxManagementJSONBodyBytes = originalLimit
	})

	handlerCalled := false
	router := gin.New()
	router.Use(RequireJSON())
	router.POST("/api/v1/user/login", func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusNoContent)
	})

	payload := fmt.Sprintf(`{"username":"%s","password":"admin"}`, strings.Repeat("a", 64))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/login", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
	if handlerCalled {
		t.Fatalf("handlerCalled = true, want false")
	}
	if !strings.Contains(strings.ToLower(rec.Body.String()), "request body too large") {
		t.Fatalf("body = %s, want request body too large", rec.Body.String())
	}
}

func TestRequireJSONDoesNotCapRelayBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalLimit := maxManagementJSONBodyBytes
	maxManagementJSONBodyBytes = 32
	t.Cleanup(func() {
		maxManagementJSONBodyBytes = originalLimit
	})

	handlerCalled := false
	router := gin.New()
	router.Use(RequireJSON())
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusNoContent)
	})

	payload := fmt.Sprintf(`{"model":"gpt-4o","input":"%s"}`, strings.Repeat("a", 64))
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if !handlerCalled {
		t.Fatalf("handlerCalled = false, want true")
	}
}

func TestRequireJSONRejectsJSONSubstringContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handlerCalled := false
	router := gin.New()
	router.Use(RequireJSON())
	router.POST("/api/v1/user/login", func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/login", strings.NewReader(`{"username":"admin"}`))
	req.Header.Set("Content-Type", "application/jsonp")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusUnsupportedMediaType, rec.Body.String())
	}
	if handlerCalled {
		t.Fatalf("handlerCalled = true, want false")
	}
}

func TestRequireJSONAcceptsStructuredJSONContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handlerCalled := false
	router := gin.New()
	router.Use(RequireJSON())
	router.POST("/api/v1/user/login", func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/login", strings.NewReader(`{"username":"admin"}`))
	req.Header.Set("Content-Type", "application/problem+json; charset=utf-8")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if !handlerCalled {
		t.Fatalf("handlerCalled = false, want true")
	}
}
