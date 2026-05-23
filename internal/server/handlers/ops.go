package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
)

func init() {
	router.NewGroupRouter("/api/v1/ops").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/overview", http.MethodGet).
				Handle(getOpsOverview),
		).
		AddRoute(
			router.NewRoute("/entities", http.MethodGet).
				Handle(getOpsEntities),
		).
		AddRoute(
			router.NewRoute("/series", http.MethodGet).
				Handle(getOpsEntitySeries),
		).
		AddRoute(
			router.NewRoute("/details", http.MethodGet).
				Handle(getOpsRecentDetails),
		)
}

func getOpsOverview(c *gin.Context) {
	overview, err := op.OpsOverviewGet(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, overview)
}

func getOpsEntities(c *gin.Context) {
	scope, ok, err := parseOptionalNonEmptyTrimmedStringQuery(c, "scope")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if !ok {
		resp.Error(c, http.StatusBadRequest, "scope is required")
		return
	}
	limit, _, err := parseOptionalPositiveIntQuery(c, "limit")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	items, err := op.OpsEntityList(c.Request.Context(), scope, limit)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, items)
}

func getOpsEntitySeries(c *gin.Context) {
	scope, ok, err := parseOptionalNonEmptyTrimmedStringQuery(c, "scope")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if !ok {
		resp.Error(c, http.StatusBadRequest, "scope is required")
		return
	}
	entityKey, _, err := parseOptionalNonEmptyTrimmedStringQuery(c, "entity_key")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if scope == model.OpsScopeOverall && entityKey == "" {
		entityKey = model.OpsEntityOverall
	}
	items, err := op.OpsEntitySeries(c.Request.Context(), scope, entityKey)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, items)
}

func getOpsRecentDetails(c *gin.Context) {
	scope, ok, err := parseOptionalNonEmptyTrimmedStringQuery(c, "scope")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if !ok {
		resp.Error(c, http.StatusBadRequest, "scope is required")
		return
	}
	entityKey, _, err := parseOptionalNonEmptyTrimmedStringQuery(c, "entity_key")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	limit, _, err := parseOptionalPositiveIntQuery(c, "limit")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	items, err := op.OpsRecentDetails(c.Request.Context(), scope, entityKey, limit)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, items)
}
