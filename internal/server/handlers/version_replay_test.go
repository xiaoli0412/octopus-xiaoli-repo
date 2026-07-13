package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

func TestVersionReturnsBuildMetadata(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/version", nil)
	version(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var info versionInfo
	if err := json.Unmarshal(recorder.Body.Bytes(), &info); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, body = %s", err, recorder.Body.String())
	}
	if info.Version != conf.Version {
		t.Fatalf("version = %q, want %q", info.Version, conf.Version)
	}
	if info.Commit != conf.Commit {
		t.Fatalf("commit = %q, want %q", info.Commit, conf.Commit)
	}
	if info.BuildTime != conf.BuildTime {
		t.Fatalf("buildTime = %q, want %q", info.BuildTime, conf.BuildTime)
	}
	if !strings.HasPrefix(info.GoVersion, "go") {
		t.Fatalf("goVersion = %q, want prefix %q", info.GoVersion, "go")
	}
}

func TestVersionResponseContainsExpectedFields(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/version", nil)
	version(ctx)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &raw); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	for _, field := range []string{"version", "commit", "buildTime", "goVersion"} {
		if _, ok := raw[field]; !ok {
			t.Fatalf("response missing required field %q, body = %s", field, recorder.Body.String())
		}
	}
}

func TestReplayRejectsInvalidLogID(t *testing.T) {
	setupHandlerTest(t)

	for _, id := range []string{"abc", "0", "-1", ""} {
		recorder := performParamHandlerRequest(t, http.MethodPost, "/api/v1/log/replay/"+id, nil, map[string]string{"id": id}, replayLog)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("id %q status = %d, want %d, body = %s", id, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
	}
}

func TestReplayReturnsNotFoundForMissingLog(t *testing.T) {
	setupHandlerTest(t)

	// Use a valid int64 that is extremely unlikely to exist in the test DB
	const missingID = "9223372036854775807"
	recorder := performParamHandlerRequest(t, http.MethodPost, "/api/v1/log/replay/"+missingID, nil, map[string]string{"id": missingID}, replayLog)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, "log not found") {
		t.Fatalf("message = %q, want %q", response.Message, "log not found")
	}
}

func TestReplayRejectsLogWithoutAPIKey(t *testing.T) {
	setupHandlerTest(t)

	logEntry := model.RelayLog{
		Time:             1700000000,
		RequestModelName: "gpt-4o",
		RequestContent:   `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`,
		ResponseContent:  `{"id":"chatcmpl-1"}`,
	}
	if err := op.RelayLogAdd(setupHandlerTestDB(t), logEntry); err != nil {
		t.Fatalf("op.RelayLogAdd() error = %v", err)
	}

	// Fetch the log to get its generated ID
	logs, err := op.RelayLogList(setupHandlerTestDB(t), nil, nil, 1, 10)
	if err != nil {
		t.Fatalf("op.RelayLogList() error = %v", err)
	}
	if len(logs) == 0 {
		t.Fatalf("expected at least one log entry")
	}
	createdID := logs[0].ID

	idStr := idToString(createdID)
	recorder := performParamHandlerRequest(t, http.MethodPost, "/api/v1/log/replay/"+idStr, nil, map[string]string{"id": idStr}, replayLog)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
}

func TestReplayRejectsLogWithEmptyRequestContent(t *testing.T) {
	setupHandlerTest(t)

	apiKey := model.APIKey{
		Name:    "replay-key",
		APIKey:  "sk-octopus-replay-empty",
		Enabled: true,
	}
	if err := op.APIKeyCreate(&apiKey, setupHandlerTestDB(t)); err != nil {
		t.Fatalf("op.APIKeyCreate() error = %v", err)
	}

	logEntry := model.RelayLog{
		Time:             1700000000,
		RequestModelName: "gpt-4o",
		RequestContent:   "",
		APIKeyID:         apiKey.ID,
	}
	if err := op.RelayLogAdd(setupHandlerTestDB(t), logEntry); err != nil {
		t.Fatalf("op.RelayLogAdd() error = %v", err)
	}

	logs, err := op.RelayLogList(setupHandlerTestDB(t), nil, nil, 1, 10)
	if err != nil {
		t.Fatalf("op.RelayLogList() error = %v", err)
	}
	if len(logs) == 0 {
		t.Fatalf("expected at least one log entry")
	}
	createdID := logs[0].ID

	idStr := idToString(createdID)
	recorder := performParamHandlerRequest(t, http.MethodPost, "/api/v1/log/replay/"+idStr, nil, map[string]string{"id": idStr}, replayLog)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, "request content is empty") {
		t.Fatalf("message = %q, want %q", response.Message, "request content is empty")
	}
}

func idToString(id int64) string {
	b, _ := json.Marshal(id)
	return strings.Trim(string(b), "\"")
}
