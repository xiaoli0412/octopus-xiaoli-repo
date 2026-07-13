package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
)

func init() {
	router.NewGroupRouter("/api/v1/stream").
		Use(middleware.AuthSSE()).
		AddRoute(
			router.NewRoute("/stats", http.MethodGet).
				Handle(streamStats),
		).
		AddRoute(
			router.NewRoute("/logs", http.MethodGet).
				Handle(streamLogs),
		)
}

// streamStats pushes an aggregated stats snapshot to the client every time a
// relay completes. It is protected by AuthSSE so the frontend can subscribe
// with the same JWT used for the management API (passed via the "token" query
// parameter because EventSource cannot set custom headers).
func streamStats(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Send an initial snapshot so the client has data immediately on connect
	// without waiting for the next relay to complete.
	initial := op.StatsSnapshotBuild()
	if data, err := json.Marshal(initial); err == nil {
		c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
		c.Writer.Flush()
	}

	ch := op.StatsSubscribe()
	defer op.StatsUnsubscribe(ch)

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case snapshot, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(snapshot)
			if err != nil {
				continue
			}
			if _, err := c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data))); err != nil {
				return
			}
			c.Writer.Flush()
		}
	}
}

// streamLogs pushes each new RelayLog entry to the client as it is recorded.
// It reuses the existing op.RelayLogSubscribe event bus and is protected by
// AuthSSE so the frontend no longer needs the one-time stream token flow.
func streamLogs(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	logChan := op.RelayLogSubscribe()
	defer op.RelayLogUnsubscribe(logChan)

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case logEntry, ok := <-logChan:
			if !ok {
				return
			}
			data, err := json.Marshal(logEntry)
			if err != nil {
				continue
			}
			if _, err := c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data))); err != nil {
				return
			}
			c.Writer.Flush()
		}
	}
}
