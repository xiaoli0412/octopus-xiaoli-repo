package handlers

import (
	"errors"
	"net/http"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/update"
)

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
