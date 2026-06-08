package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
)

const relayLogExportMaxLimit = 10000
const relayLogListMaxOffset = 10000

type relayLogExportFormat string

const (
	relayLogExportFormatJSON  relayLogExportFormat = "json"
	relayLogExportFormatJSONL relayLogExportFormat = "jsonl"
)

func init() {
	router.NewGroupRouter("/api/v1/log").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(listLog),
		).
		AddRoute(
			router.NewRoute("/clear", http.MethodDelete).
				Handle(clearLog),
		).
		AddRoute(
			router.NewRoute("/stream-token", http.MethodGet).
				Handle(getStreamToken),
		).
		AddRoute(
			router.NewRoute("/export", http.MethodGet).
				Handle(exportLog),
		)

	router.NewGroupRouter("/api/v1/log").
		AddRoute(
			router.NewRoute("/stream", http.MethodGet).
				Handle(streamLog),
		)
}

func listLog(c *gin.Context) {
	page, _, err := parseOptionalBoundedIntQuery(c, "page", 1, 1, 0)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, "invalid page")
		return
	}
	pageSize, _, err := parseOptionalBoundedIntQuery(c, "page_size", 20, 1, 100)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, "invalid page_size")
		return
	}
	if int64(page-1)*int64(pageSize) >= relayLogListMaxOffset {
		resp.Error(c, http.StatusBadRequest, "invalid page")
		return
	}

	startTime, endTime, err := parseOptionalIntRangeQuery(c, "start_time", "end_time")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	logs, err := op.RelayLogList(c.Request.Context(), startTime, endTime, page, pageSize)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to list logs")
		return
	}

	resp.Success(c, logs)
}

func clearLog(c *gin.Context) {
	if err := op.RelayLogClear(c.Request.Context()); err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to clear logs")
		return
	}
	resp.Success(c, nil)
}

func getStreamToken(c *gin.Context) {
	token, err := op.RelayLogStreamTokenCreate()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to create stream token")
		return
	}
	resp.Success(c, gin.H{"token": token})
}

func streamLog(c *gin.Context) {
	token, _, err := parseOptionalNonEmptyTrimmedStringQuery(c, "token")
	if err != nil || !op.RelayLogStreamTokenConsume(token) {
		resp.Error(c, http.StatusUnauthorized, "invalid stream token")
		return
	}

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
		case log, ok := <-logChan:
			if !ok {
				return
			}
			data, err := json.Marshal(log)
			if err != nil {
				continue
			}
			c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
			c.Writer.Flush()
		}
	}
}

func exportLog(c *gin.Context) {
	format, err := parseOptionalRelayLogExportFormat(c, "format", relayLogExportFormatJSONL)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	limit, _, err := parseOptionalBoundedIntQuery(c, "limit", 2000, 1, relayLogExportMaxLimit)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, "invalid limit")
		return
	}
	startTime, endTime, err := parseOptionalIntRangeQuery(c, "start_time", "end_time")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	logs, err := op.RelayLogExport(c.Request.Context(), startTime, endTime, limit)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to export logs")
		return
	}

	fileName := fmt.Sprintf("octopus-logs-%s.%s", time.Now().Format("20060102-150405"), format)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))

	if format == relayLogExportFormatJSON {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.JSON(http.StatusOK, logs)
		return
	}

	c.Header("Content-Type", "application/x-ndjson; charset=utf-8")
	writer := bufio.NewWriter(c.Writer)
	defer writer.Flush()
	for _, item := range logs {
		line, err := json.Marshal(item)
		if err != nil {
			continue
		}
		_, _ = writer.Write(line)
		_, _ = writer.WriteString("\n")
	}
}
