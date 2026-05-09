package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	serverauth "github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/auth"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
	"github.com/gin-gonic/gin"
)

func TestAdminPostRoutesRequireJSONForEmptyBody(t *testing.T) {
	setupHandlerTest(t)

	token, _, err := serverauth.GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}

	engine := gin.New()
	if err := router.RegisterAll(engine); err != nil {
		t.Fatalf("RegisterAll() error = %v", err)
	}

	for _, target := range []string{
		"/api/v1/ai/sessions/1/replan",
		"/api/v1/ai/sessions/1/apply",
		"/api/v1/ai/strategy-profiles/1/activate",
		"/api/v1/channel/sync",
		"/api/v1/channel/copilot/device-code",
		"/api/v1/channel/antigravity/oauth/start",
		"/api/v1/dynamic-routing/learning/reset",
		"/api/v1/update",
		"/api/v1/setting/rollback-latest-import",
	} {
		t.Run(target, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, target, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			engine.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusUnsupportedMediaType {
				t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusUnsupportedMediaType, recorder.Body.String())
			}
		})
	}
}
