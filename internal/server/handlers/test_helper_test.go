package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/gin-gonic/gin"
)

type handlerResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func setupHandlerTest(t *testing.T) {
	t.Helper()
	t.Setenv(op.BootstrapAdminUsernameEnv(), op.BootstrapAdminDefaultUsername())
	t.Setenv(op.BootstrapAdminPasswordEnv(), "admin")
	gin.SetMode(gin.TestMode)
	setupHandlerTestDB(t)
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() error = %v", err)
	}
	if err := initializeHandlerUser(); err != nil {
		t.Fatalf("initializeHandlerUser() error = %v", err)
	}
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() after user init error = %v", err)
	}
}

func setupHandlerTestDB(t *testing.T) context.Context {
	t.Helper()
	resetHandlerTestState()
	if db.GetDB() != nil {
		_ = db.Close()
	}

	dbPath := filepath.Join(t.TempDir(), "octopus-handler-test.db")
	if err := db.InitDB("sqlite", dbPath, false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	t.Cleanup(func() {
		if db.GetDB() != nil {
			_ = db.Close()
		}
		resetHandlerTestState()
	})

	return context.Background()
}

func resetHandlerTestState() {
	resetLoginThrottleState()
	resetProvidersState()
}

func initializeHandlerUser() error {
	return op.UserInit()
}

func initializeHandlerCaches() error {
	return op.InitCache()
}

func performJSONHandlerRequest(t *testing.T, method string, target string, body any, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, target, bytes.NewReader(payload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	handler(ctx)
	return recorder
}

func performParamHandlerRequest(t *testing.T, method string, target string, body any, params map[string]string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	recorder := performJSONHandlerRequest(t, method, target, body, func(ctx *gin.Context) {
		ctx.Params = make(gin.Params, 0, len(params))
		for key, value := range params {
			ctx.Params = append(ctx.Params, gin.Param{Key: key, Value: value})
		}
		handler(ctx)
	})
	return recorder
}

func decodeHandlerResponse(t *testing.T, recorder *httptest.ResponseRecorder) handlerResponse {
	t.Helper()
	var res handlerResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &res); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, body = %s", err, recorder.Body.String())
	}
	return res
}
