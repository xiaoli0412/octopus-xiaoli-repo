package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
)

const aiAutomationLegacyGoneMessage = "ai automation v1 endpoint is retired; use governance sessions endpoints"

func init() {
	router.NewGroupRouter("/api/v1/ai").
		Use(middleware.Auth()).
		AddRoute(router.NewRoute("/overview", http.MethodGet).Handle(getAIGovernanceOverview)).
		AddRoute(router.NewRoute("/sessions", http.MethodGet).Handle(listGovernanceSessions)).
		AddRoute(router.NewRoute("/sessions", http.MethodPost).Use(middleware.RequireJSON()).Handle(createGovernanceSession)).
		AddRoute(router.NewRoute("/sessions/:id", http.MethodGet).Handle(getGovernanceSession)).
		AddRoute(router.NewRoute("/sessions/:id/replan", http.MethodPost).Use(middleware.RequireJSON()).Handle(replanGovernanceSession)).
		AddRoute(router.NewRoute("/sessions/:id/apply", http.MethodPost).Use(middleware.RequireJSON()).Handle(applyGovernanceSession)).
		AddRoute(router.NewRoute("/sessions/:id/rollback", http.MethodPost).Use(middleware.RequireJSON()).Handle(rollbackGovernanceSession)).
		AddRoute(router.NewRoute("/sessions/:id/apply-runs", http.MethodGet).Handle(listGovernanceApplyRuns)).
		AddRoute(router.NewRoute("/rollback-points", http.MethodGet).Handle(listGovernanceRollbackPoints)).
		AddRoute(router.NewRoute("/runtime-policy", http.MethodGet).Handle(getGovernanceRuntimePolicy)).
		AddRoute(router.NewRoute("/runtime-policy", http.MethodPost).Use(middleware.RequireJSON()).Handle(updateGovernanceRuntimePolicy)).
		AddRoute(router.NewRoute("/strategy-profiles", http.MethodGet).Handle(listGovernanceStrategyProfiles)).
		AddRoute(router.NewRoute("/strategy-profiles", http.MethodPost).Use(middleware.RequireJSON()).Handle(createGovernanceStrategyProfile)).
		AddRoute(router.NewRoute("/saved-plans", http.MethodGet).Handle(listGovernanceStrategyProfiles)).
		AddRoute(router.NewRoute("/saved-plans", http.MethodPost).Use(middleware.RequireJSON()).Handle(createGovernanceStrategyProfile)).
		AddRoute(router.NewRoute("/saved-plans/:id/apply", http.MethodPost).Use(middleware.RequireJSON()).Handle(activateGovernanceStrategyProfile)).
		AddRoute(router.NewRoute("/strategy-profiles/:id/activate", http.MethodPost).Use(middleware.RequireJSON()).Handle(activateGovernanceStrategyProfile)).
		AddRoute(router.NewRoute("/expert-presets", http.MethodGet).Handle(listGovernanceExpertPresets)).
		AddRoute(router.NewRoute("/learning/summary", http.MethodGet).Handle(getAIGovernanceLearningSummary)).
		AddRoute(router.NewRoute("/config", http.MethodGet).Handle(goneAIAutomationV1)).
		AddRoute(router.NewRoute("/config", http.MethodPost).Use(middleware.RequireJSON()).Handle(goneAIAutomationV1)).
		AddRoute(router.NewRoute("/models/fetch", http.MethodPost).Use(middleware.RequireJSON()).Handle(goneAIAutomationV1)).
		AddRoute(router.NewRoute("/prompt-templates", http.MethodGet).Handle(goneAIAutomationV1)).
		AddRoute(router.NewRoute("/prompt-templates", http.MethodPost).Use(middleware.RequireJSON()).Handle(goneAIAutomationV1)).
		AddRoute(router.NewRoute("/tasks", http.MethodGet).Handle(goneAIAutomationV1)).
		AddRoute(router.NewRoute("/tasks", http.MethodPost).Use(middleware.RequireJSON()).Handle(goneAIAutomationV1)).
		AddRoute(router.NewRoute("/tasks/:id", http.MethodGet).Handle(goneAIAutomationV1)).
		AddRoute(router.NewRoute("/tasks/:id/artifacts", http.MethodGet).Handle(goneAIAutomationV1)).
		AddRoute(router.NewRoute("/tasks/:id/cancel", http.MethodPost).Use(middleware.RequireJSON()).Handle(goneAIAutomationV1)).
		AddRoute(router.NewRoute("/tasks/:id/retry", http.MethodPost).Use(middleware.RequireJSON()).Handle(goneAIAutomationV1)).
		AddRoute(router.NewRoute("/profiles", http.MethodGet).Handle(goneAIAutomationV1)).
		AddRoute(router.NewRoute("/profiles/:id", http.MethodGet).Handle(goneAIAutomationV1)).
		AddRoute(router.NewRoute("/profiles/:id/activate", http.MethodPost).Use(middleware.RequireJSON()).Handle(goneAIAutomationV1))
}

func respondAIAutomationError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	message := err.Error()
	switch {
	case errors.Is(err, op.ErrAIAutomationDisabled):
		resp.Error(c, http.StatusForbidden, message)
	case errors.Is(err, op.ErrGovernanceSessionStale()):
		resp.Error(c, http.StatusConflict, message)
	case message == aiAutomationLegacyGoneMessage:
		resp.Error(c, http.StatusGone, message)
	case message == "governance session not found" || message == "group not found":
		resp.Error(c, http.StatusNotFound, message)
	case message == "invalid governance session id" || message == "invalid expert preset" || message == "goal is required" || message == "session_id is required" || message == "name is required" || message == "invalid strategy profile id" || message == "governance preview is not applyable":
		resp.Error(c, http.StatusBadRequest, message)
	default:
		resp.Error(c, http.StatusInternalServerError, message)
	}
}

func getAIGovernanceOverview(c *gin.Context) {
	result, err := op.AIGovernanceOverviewGet(c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, result)
}

func listGovernanceSessions(c *gin.Context) {
	items, err := op.GovernanceSessionList(c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, items)
}

func createGovernanceSession(c *gin.Context) {
	var req model.GovernanceSessionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	result, err := op.GovernanceSessionCreate(req, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, result)
}

func getGovernanceSession(c *gin.Context) {
	id, ok := parsePositivePathIDValue(c, "id")
	if !ok {
		return
	}
	result, err := op.GovernanceSessionGet(id, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, result)
}

func replanGovernanceSession(c *gin.Context) {
	id, ok := parsePositivePathIDValue(c, "id")
	if !ok {
		return
	}
	result, err := op.GovernanceSessionReplan(id, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, result)
}

func applyGovernanceSession(c *gin.Context) {
	id, ok := parsePositivePathIDValue(c, "id")
	if !ok {
		return
	}
	result, err := op.GovernanceSessionApply(id, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, result)
}

func listGovernanceApplyRuns(c *gin.Context) {
	id, ok := parsePositivePathIDValue(c, "id")
	if !ok {
		return
	}
	items, err := op.GovernanceApplyRunList(id, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, items)
}

func rollbackGovernanceSession(c *gin.Context) {
	id, ok := parsePositivePathIDValue(c, "id")
	if !ok {
		return
	}
	var req model.GovernanceRollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	result, err := op.GovernanceSessionRollback(id, req.RollbackPointID, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, result)
}

func listGovernanceRollbackPoints(c *gin.Context) {
	sessionID, _, err := parseOptionalPositiveIntQuery(c, "session_id")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	items, err := op.GovernanceRollbackPointList(sessionID, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, items)
}

func getGovernanceRuntimePolicy(c *gin.Context) {
	resp.Success(c, op.GovernanceRuntimePolicyGet())
}

func updateGovernanceRuntimePolicy(c *gin.Context) {
	var req model.GovernanceRuntimePolicyView
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	result, err := op.GovernanceRuntimePolicyUpdate(req)
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, result)
}

func listGovernanceStrategyProfiles(c *gin.Context) {
	items, err := op.StrategyProfileList(c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, items)
}

func createGovernanceStrategyProfile(c *gin.Context) {
	var req model.GovernanceStrategyProfileCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	result, err := op.StrategyProfileCreate(req, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, result)
}

func activateGovernanceStrategyProfile(c *gin.Context) {
	id, ok := parsePositivePathIDValue(c, "id")
	if !ok {
		return
	}
	result, err := op.StrategyProfileActivate(id, c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, result)
}

func listGovernanceExpertPresets(c *gin.Context) {
	resp.Success(c, op.AIGovernanceExpertPresetList())
}

func getAIGovernanceLearningSummary(c *gin.Context) {
	result, err := op.AIGovernanceLearningSummaryGet(c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, result)
}

func goneAIAutomationV1(c *gin.Context) {
	resp.Error(c, http.StatusGone, aiAutomationLegacyGoneMessage)
}
