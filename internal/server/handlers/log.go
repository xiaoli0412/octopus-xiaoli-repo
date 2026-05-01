package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
	"github.com/gin-gonic/gin"
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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var startTime, endTime *int
	if startTimeStr != "" && endTimeStr != "" {
		st, err := strconv.Atoi(startTimeStr)
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		et, err := strconv.Atoi(endTimeStr)
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		startTime = &st
		endTime = &et
	}

	logs, err := op.RelayLogList(c.Request.Context(), startTime, endTime, page, pageSize)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	resp.Success(c, logs)
}

func clearLog(c *gin.Context) {
	if err := op.RelayLogClear(c.Request.Context()); err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, nil)
}

func getStreamToken(c *gin.Context) {
	token, err := op.RelayLogStreamTokenCreate()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, gin.H{"token": token})
}

func streamLog(c *gin.Context) {
	token := c.Query("token")
	if token == "" || !op.RelayLogStreamTokenConsume(token) {
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
	format := c.DefaultQuery("format", "jsonl")
	if format != "json" && format != "jsonl" {
		resp.Error(c, http.StatusBadRequest, "unsupported format")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "2000"))
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	var startTime, endTime *int
	if startTimeStr != "" && endTimeStr != "" {
		st, err := strconv.Atoi(startTimeStr)
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		et, err := strconv.Atoi(endTimeStr)
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		startTime = &st
		endTime = &et
	}

	logs, err := op.RelayLogExport(c.Request.Context(), startTime, endTime, limit)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	fileName := fmt.Sprintf("octopus-logs-%s.%s", time.Now().Format("20060102-150405"), format)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))

	if format == "json" {
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
