package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
)

func init() {
	router.NewGroupRouter("/api/v1/ai").
		Use(middleware.Auth()).
		AddRoute(router.NewRoute("/config", http.MethodGet).Handle(getAIAutomationConfig)).
		AddRoute(router.NewRoute("/config", http.MethodPost).Use(middleware.RequireJSON()).Handle(updateAIAutomationConfig)).
		AddRoute(router.NewRoute("/models/fetch", http.MethodPost).Use(middleware.RequireJSON()).Handle(fetchAIAutomationModels)).
		AddRoute(router.NewRoute("/prompt-templates", http.MethodGet).Handle(listAIPromptTemplates)).
		AddRoute(router.NewRoute("/prompt-templates", http.MethodPost).Use(middleware.RequireJSON()).Handle(createAIPromptTemplate)).
		AddRoute(router.NewRoute("/tasks", http.MethodGet).Handle(listAITasks)).
		AddRoute(router.NewRoute("/tasks", http.MethodPost).Use(middleware.RequireJSON()).Handle(createAITask)).
		AddRoute(router.NewRoute("/tasks/:id", http.MethodGet).Handle(getAITask)).
		AddRoute(router.NewRoute("/tasks/:id/artifacts", http.MethodGet).Handle(getAITaskArtifacts)).
		AddRoute(router.NewRoute("/tasks/:id/cancel", http.MethodPost).Use(middleware.RequireJSON()).Handle(cancelAITask)).
		AddRoute(router.NewRoute("/tasks/:id/retry", http.MethodPost).Use(middleware.RequireJSON()).Handle(retryAITask)).
		AddRoute(router.NewRoute("/profiles", http.MethodGet).Handle(listAIProfiles)).
		AddRoute(router.NewRoute("/profiles/:id", http.MethodGet).Handle(getAIProfile)).
		AddRoute(router.NewRoute("/profiles/:id/activate", http.MethodPost).Use(middleware.RequireJSON()).Handle(activateAIProfile))
}

func respondAIAutomationError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	message := err.Error()
	switch {
	case errors.Is(err, context.Canceled):
		resp.Error(c, http.StatusRequestTimeout, message)
	case errors.Is(err, op.ErrAIAutomationDisabled):
		resp.Error(c, http.StatusForbidden, message)
	case strings.Contains(message, "not found"):
		resp.Error(c, http.StatusNotFound, message)
	case strings.Contains(message, "invalid"), strings.Contains(message, "required"), strings.Contains(message, "cannot be activated"):
		resp.Error(c, http.StatusBadRequest, message)
	default:
		resp.Error(c, http.StatusInternalServerError, message)
	}
}

func getAIAutomationConfig(c *gin.Context) {
	config, err := op.AIAutomationConfigGet(c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, config)
}

func updateAIAutomationConfig(c *gin.Context) {
	var req model.AIAutomationConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	config, err := op.AIAutomationConfigUpdate(req, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, config)
}

func fetchAIAutomationModels(c *gin.Context) {
	var req model.AIModelsFetchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	result, err := op.AIAutomationFetchModels(req, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, result)
}

func listAIPromptTemplates(c *gin.Context) {
	rows, err := op.AIPromptTemplateList(c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, rows)
}

func createAIPromptTemplate(c *gin.Context) {
	var req model.AIPromptTemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	row, err := op.AIPromptTemplateCreate(req, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, row)
}

func createAITask(c *gin.Context) {
	var req model.AITaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	task, err := op.AITaskCreate(req, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, op.RedactAITaskForResponse(task))
}

func listAITasks(c *gin.Context) {
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
	if int64(page-1)*int64(pageSize) >= 10000 {
		resp.Error(c, http.StatusBadRequest, "invalid page")
		return
	}
	status, _, err := parseOptionalNonEmptyTrimmedStringQuery(c, "status")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	taskType, _, err := parseOptionalNonEmptyTrimmedStringQuery(c, "type")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	profileDomain, _, err := parseOptionalNonEmptyTrimmedStringQuery(c, "profile_domain")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	keyword, _, err := parseOptionalNonEmptyTrimmedStringQuery(c, "keyword")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	req := model.AITaskListRequest{
		Page:          page,
		PageSize:      pageSize,
		Status:        status,
		Type:          taskType,
		ProfileDomain: profileDomain,
		Keyword:       keyword,
	}
	createdFrom, err := parseOptionalRFC3339TimeQuery(c, "created_from")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if createdFrom != nil {
		req.CreatedFrom = *createdFrom
	}
	createdTo, err := parseOptionalRFC3339TimeQuery(c, "created_to")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if createdTo != nil {
		req.CreatedTo = *createdTo
	}
	result, err := op.AITaskList(req, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, result)
}

func getAITask(c *gin.Context) {
	id, ok := parsePositivePathID(c, "id")
	if !ok {
		return
	}
	task, err := op.AITaskGet(id, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, op.RedactAITaskForResponse(task))
}

func getAITaskArtifacts(c *gin.Context) {
	id, ok := parsePositivePathID(c, "id")
	if !ok {
		return
	}
	artifacts, err := op.AITaskArtifacts(id, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, op.RedactAITaskArtifactsForResponse(artifacts))
}

func retryAITask(c *gin.Context) {
	id, ok := parsePositivePathID(c, "id")
	if !ok {
		return
	}
	task, err := op.AITaskRetry(id, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, op.RedactAITaskForResponse(task))
}

func cancelAITask(c *gin.Context) {
	id, ok := parsePositivePathID(c, "id")
	if !ok {
		return
	}
	task, err := op.AITaskCancel(id, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, op.RedactAITaskForResponse(task))
}

func listAIProfiles(c *gin.Context) {
	rows, err := op.AIProfileList(c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, rows)
}

func getAIProfile(c *gin.Context) {
	id, ok := parsePositivePathID(c, "id")
	if !ok {
		return
	}
	profile, err := op.AIProfileGetRedacted(id, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, profile)
}

func activateAIProfile(c *gin.Context) {
	id, ok := parsePositivePathID(c, "id")
	if !ok {
		return
	}
	profile, err := op.AIProfileActivate(id, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, op.RedactAIProfileForResponse(profile))
}

func parsePositivePathID(c *gin.Context, name string) (int, bool) {
	id, ok := parsePositivePathIDValue(c, name)
	if !ok {
		resp.Error(c, http.StatusBadRequest, "invalid "+name)
		return 0, false
	}
	return id, true
}
