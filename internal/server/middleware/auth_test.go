package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	serverauth "github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/auth"
	"github.com/gin-gonic/gin"
)

func TestBearerTokenFromHeaderRequiresBearerScheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header string
		want   string
		ok     bool
	}{
		{name: "standard bearer", header: "Bearer token-123", want: "token-123", ok: true},
		{name: "case insensitive scheme", header: "bearer token-123", want: "token-123", ok: true},
		{name: "missing scheme", header: "token-123", ok: false},
		{name: "empty token", header: "Bearer   ", ok: false},
		{name: "wrong scheme", header: "Basic token-123", ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := bearerTokenFromHeader(tc.header)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if got != tc.want {
				t.Fatalf("token = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAuthRejectsAuthorizationHeaderWithoutBearerScheme(t *testing.T) {
	setupCORSTestState(t)

	token, _, err := serverauth.GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}

	router := gin.New()
	router.Use(Auth())
	router.GET("/api/v1/secure", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secure", nil)
	req.Header.Set("Authorization", token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestAPIKeyAuthRejectsAuthorizationHeaderWithoutBearerScheme(t *testing.T) {
	setupCORSTestState(t)

	apiKey := &model.APIKey{
		Name:    "middleware-auth",
		APIKey:  serverauth.GenerateAPIKey(),
		Enabled: true,
	}
	if err := op.APIKeyCreate(apiKey, context.Background()); err != nil {
		t.Fatalf("APIKeyCreate() error = %v", err)
	}

	router := gin.New()
	router.Use(APIKeyAuth())
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("Authorization", apiKey.APIKey)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestAuthBlocksProtectedRoutesUntilForcedPasswordChangeCompletes(t *testing.T) {
	setupCORSTestState(t)
	if err := op.SettingSetString(model.SettingKeyForcePasswordChange, "true"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	token, _, err := serverauth.GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}

	router := gin.New()
	router.Use(Auth())
	router.GET("/api/v1/channel/list", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channel/list", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestAuthAllowsForcedPasswordChangeRoutesDuringFirstLoginGate(t *testing.T) {
	setupCORSTestState(t)
	if err := op.SettingSetString(model.SettingKeyForcePasswordChange, "true"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	token, _, err := serverauth.GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}

	router := gin.New()
	router.Use(Auth())
	router.POST("/api/v1/user/force-change-password", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	router.GET("/api/v1/user/status", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/api/v1/user/force-change-password"},
		{method: http.MethodGet, path: "/api/v1/user/status"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("%s %s status = %d, want %d, body = %s", tc.method, tc.path, rec.Code, http.StatusNoContent, rec.Body.String())
		}
	}
}
