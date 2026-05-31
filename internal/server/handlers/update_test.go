package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/update"
)

func TestGetUpdateStatusReturnsCurrentCapability(t *testing.T) {
	setupHandlerTest(t)

	originalGOOS := update.CurrentRuntimeGOOSForTest()
	update.SetCurrentRuntimeGOOSForTest("windows")
	t.Cleanup(func() {
		update.SetCurrentRuntimeGOOSForTest(originalGOOS)
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/update/status", nil)
	getUpdateStatus(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	var payload update.StatusInfo
	if err := json.Unmarshal(response.Data, &payload); err != nil {
		t.Fatalf("json.Unmarshal(status) error = %v", err)
	}
	if payload.Version != conf.Version {
		t.Fatalf("version = %q, want %q", payload.Version, conf.Version)
	}
	if payload.SelfUpdateSupported {
		t.Fatal("self_update_supported = true, want false on windows")
	}
	if payload.SelfUpdateUnsupportedReason != update.ErrUpdateUnsupportedPlatform.Error() {
		t.Fatalf("self_update_unsupported_reason = %q, want %q", payload.SelfUpdateUnsupportedReason, update.ErrUpdateUnsupportedPlatform.Error())
	}
}
func TestUpdateRouteAllowsEmptyPostBody(t *testing.T) {
	setupHandlerTest(t)

	original := runUpdateCore
	runUpdateCore = func() error { return update.ErrUpdateInProgress }
	t.Cleanup(func() {
		runUpdateCore = original
	})

	engine := gin.New()
	engine.POST("/api/v1/update", func(ctx *gin.Context) {
		updateFunc(ctx)
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/update", nil)

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusConflict, recorder.Body.String())
	}
}

func TestUpdateFuncReturnsConflictWhenUpdateAlreadyRunning(t *testing.T) {
	setupHandlerTest(t)

	original := runUpdateCore
	runUpdateCore = func() error {
		return update.ErrUpdateInProgress
	}
	t.Cleanup(func() {
		runUpdateCore = original
	})

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/update", map[string]any{}, updateFunc)
	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusConflict, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if response.Message != update.ErrUpdateInProgress.Error() {
		t.Fatalf("message = %q, want %q", response.Message, update.ErrUpdateInProgress.Error())
	}
}

func TestUpdateFuncReturnsInternalServerErrorForUnexpectedFailure(t *testing.T) {
	setupHandlerTest(t)

	original := runUpdateCore
	runUpdateCore = func() error {
		return errors.New("boom")
	}
	t.Cleanup(func() {
		runUpdateCore = original
	})

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/update", map[string]any{}, updateFunc)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusInternalServerError, recorder.Body.String())
	}
}

func TestUpdateFuncReturnsNotImplementedForUnsupportedPlatform(t *testing.T) {
	setupHandlerTest(t)

	original := runUpdateCore
	runUpdateCore = func() error {
		return update.ErrUpdateUnsupportedPlatform
	}
	t.Cleanup(func() {
		runUpdateCore = original
	})

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/update", map[string]any{}, updateFunc)
	if recorder.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusNotImplemented, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if response.Message != update.ErrUpdateUnsupportedPlatform.Error() {
		t.Fatalf("message = %q, want %q", response.Message, update.ErrUpdateUnsupportedPlatform.Error())
	}
}
