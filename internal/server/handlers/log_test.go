package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/gin-gonic/gin"
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
	done := make(chan struct{})
	go func() {
		defer close(done)
		streamLog(ctx)
	}()
	deadline := time.Now().Add(500 * time.Millisecond)
	for !strings.HasPrefix(first.Header().Get("Content-Type"), "text/event-stream") && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if first.Code != http.StatusOK {
		cancel()
		<-done
		t.Fatalf("first stream status = %d, want %d, body = %s", first.Code, http.StatusOK, first.Body.String())
	}
	if !strings.HasPrefix(first.Header().Get("Content-Type"), "text/event-stream") {
		cancel()
		<-done
		t.Fatalf("first content-type = %q, want prefix %q", first.Header().Get("Content-Type"), "text/event-stream")
	}
	cancel()
	<-done

	second := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/log/stream?token="+token, nil, streamLog)
	if second.Code != http.StatusUnauthorized {
		t.Fatalf("second stream status = %d, want %d, body = %s", second.Code, http.StatusUnauthorized, second.Body.String())
	}
}
