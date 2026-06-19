package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

func TestStreamLogConsumesTokenOnFirstUse(t *testing.T) {
	setupHandlerTest(t)

	token, err := op.RelayLogStreamTokenCreate()
	if err != nil {
		t.Fatalf("RelayLogStreamTokenCreate() error = %v", err)
	}

	first := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(first)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/log/stream?token="+token, nil)
	requestCtx, cancel := context.WithCancel(req.Context())
	ctx.Request = req.WithContext(requestCtx)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()
	streamLog(ctx)

	if first.Code != http.StatusOK {
		t.Fatalf("first stream status = %d, want %d, body = %s", first.Code, http.StatusOK, first.Body.String())
	}
	if !strings.HasPrefix(first.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatalf("first content-type = %q, want prefix %q", first.Header().Get("Content-Type"), "text/event-stream")
	}

	second := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/log/stream?token="+token, nil, streamLog)
	if second.Code != http.StatusUnauthorized {
		t.Fatalf("second stream status = %d, want %d, body = %s", second.Code, http.StatusUnauthorized, second.Body.String())
	}
}

func TestStreamLogRejectsMissingOrBlankToken(t *testing.T) {
	setupHandlerTest(t)

	for _, target := range []string{
		"/api/v1/log/stream",
		"/api/v1/log/stream?token=",
		"/api/v1/log/stream?token=%20",
	} {
		recorder := performJSONHandlerRequest(t, http.MethodGet, target, nil, streamLog)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("target %s status = %d, want %d, body = %s", target, recorder.Code, http.StatusUnauthorized, recorder.Body.String())
		}
		response := decodeHandlerResponse(t, recorder)
		if response.Message != "invalid stream token" {
			t.Fatalf("target %s message = %q, want %q", target, response.Message, "invalid stream token")
		}
	}
}

func TestListLogRejectsInvalidPagingQueryValues(t *testing.T) {
	setupHandlerTest(t)

	for _, target := range []string{
		`/api/v1/log/list?page=nan`,
		`/api/v1/log/list?page=`,
		`/api/v1/log/list?page=%20`,
		`/api/v1/log/list?page=0`,
		`/api/v1/log/list?page=-1`,
		`/api/v1/log/list?page=1001&page_size=10`,
		`/api/v1/log/list?page_size=huge`,
		`/api/v1/log/list?page_size=`,
		`/api/v1/log/list?page_size=%20`,
		`/api/v1/log/list?page_size=0`,
		`/api/v1/log/list?page_size=-3`,
		`/api/v1/log/list?page_size=101`,
	} {
		recorder := performJSONHandlerRequest(t, http.MethodGet, target, nil, listLog)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("target %s status = %d, want %d, body = %s", target, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
	}
}

func TestListLogRejectsPartialTimeRangeQueryValues(t *testing.T) {
	setupHandlerTest(t)

	for _, target := range []string{
		`/api/v1/log/list?start_time=1700000000`,
		`/api/v1/log/list?end_time=1700000100`,
	} {
		recorder := performJSONHandlerRequest(t, http.MethodGet, target, nil, listLog)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("target %s status = %d, want %d, body = %s", target, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
		response := decodeHandlerResponse(t, recorder)
		if !strings.Contains(response.Message, `must be provided together`) {
			t.Fatalf("message = %q, want partial time range rejection", response.Message)
		}
	}
}

func TestListLogRejectsBlankTimeRangeQueryValues(t *testing.T) {
	setupHandlerTest(t)

	for _, target := range []string{
		`/api/v1/log/list?start_time=&end_time=1700000100`,
		`/api/v1/log/list?start_time=1700000000&end_time=`,
		`/api/v1/log/list?start_time=%20&end_time=1700000100`,
		`/api/v1/log/list?start_time=1700000000&end_time=%20`,
	} {
		recorder := performJSONHandlerRequest(t, http.MethodGet, target, nil, listLog)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("target %s status = %d, want %d, body = %s", target, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
		response := decodeHandlerResponse(t, recorder)
		if !strings.Contains(response.Message, `invalid start_time`) {
			t.Fatalf("target %s message = %q, want invalid start_time", target, response.Message)
		}
	}
}

func TestExportLogRejectsInvalidLimitQueryValue(t *testing.T) {
	setupHandlerTest(t)

	for _, target := range []string{
		"/api/v1/log/export?limit=lots",
		"/api/v1/log/export?limit=",
		"/api/v1/log/export?limit=%20",
		"/api/v1/log/export?limit=0",
		"/api/v1/log/export?limit=-5",
		"/api/v1/log/export?limit=10001",
	} {
		recorder := performJSONHandlerRequest(t, http.MethodGet, target, nil, exportLog)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("target %s status = %d, want %d, body = %s", target, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
	}
}

func TestExportLogDefaultsFormatAndTrimsWhitespace(t *testing.T) {
	setupHandlerTest(t)

	if err := op.RelayLogAdd(context.Background(), model.RelayLog{Time: 1700000001, RequestModelName: "format-trim"}); err != nil {
		t.Fatalf("RelayLogAdd() error = %v", err)
	}

	for _, target := range []string{
		"/api/v1/log/export?limit=1",
		"/api/v1/log/export?format=%20json%20&limit=1",
	} {
		recorder := performJSONHandlerRequest(t, http.MethodGet, target, nil, exportLog)
		if recorder.Code != http.StatusOK {
			t.Fatalf("target %s status = %d, want %d, body = %s", target, recorder.Code, http.StatusOK, recorder.Body.String())
		}
	}
}

func TestExportLogRejectsInvalidFormatQueryValue(t *testing.T) {
	setupHandlerTest(t)

	for _, target := range []string{
		"/api/v1/log/export?format=&limit=1",
		"/api/v1/log/export?format=%20&limit=1",
		"/api/v1/log/export?format=surprise&limit=1",
	} {
		recorder := performJSONHandlerRequest(t, http.MethodGet, target, nil, exportLog)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("target %s status = %d, want %d, body = %s", target, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
		response := decodeHandlerResponse(t, recorder)
		if !strings.Contains(response.Message, "unsupported format") {
			t.Fatalf("target %s message = %q, want unsupported format", target, response.Message)
		}
	}
}

func TestExportLogSupportsValidPositiveLimit(t *testing.T) {
	setupHandlerTest(t)

	if err := op.RelayLogAdd(context.Background(), model.RelayLog{Time: 1700000000, RequestModelName: "limit-positive"}); err != nil {
		t.Fatalf("RelayLogAdd() error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/log/export?format=json&limit=1", nil, exportLog)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var logs []map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &logs); err != nil {
		t.Fatalf("json.Unmarshal(recorder.Body.Bytes()) error = %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("logs len = %d, want 1 for limit=1", len(logs))
	}
	if got := logs[0]["request_model_name"]; got != "limit-positive" {
		t.Fatalf("request_model_name = %v, want limit-positive", got)
	}
}

func TestExportLogRejectsPartialTimeRangeQueryValues(t *testing.T) {
	setupHandlerTest(t)

	for _, target := range []string{
		`/api/v1/log/export?start_time=1700000000`,
		`/api/v1/log/export?end_time=1700000100`,
	} {
		recorder := performJSONHandlerRequest(t, http.MethodGet, target, nil, exportLog)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("target %s status = %d, want %d, body = %s", target, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
		response := decodeHandlerResponse(t, recorder)
		if !strings.Contains(response.Message, `must be provided together`) {
			t.Fatalf("message = %q, want partial time range rejection", response.Message)
		}
	}
}

func TestExportLogRejectsBlankTimeRangeQueryValues(t *testing.T) {
	setupHandlerTest(t)

	for _, target := range []string{
		`/api/v1/log/export?start_time=&end_time=1700000100`,
		`/api/v1/log/export?start_time=1700000000&end_time=`,
		`/api/v1/log/export?start_time=%20&end_time=1700000100`,
		`/api/v1/log/export?start_time=1700000000&end_time=%20`,
	} {
		recorder := performJSONHandlerRequest(t, http.MethodGet, target, nil, exportLog)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("target %s status = %d, want %d, body = %s", target, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
		response := decodeHandlerResponse(t, recorder)
		if !strings.Contains(response.Message, `invalid start_time`) {
			t.Fatalf("target %s message = %q, want invalid start_time", target, response.Message)
		}
	}
}
