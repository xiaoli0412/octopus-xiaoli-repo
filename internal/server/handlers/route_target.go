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

func respondRouteTargetOpError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	message := err.Error()
	switch {
	case errors.Is(err, context.Canceled):
		resp.Error(c, http.StatusRequestTimeout, message)
	case strings.Contains(message, "not found"):
		resp.Error(c, http.StatusNotFound, message)
	case strings.Contains(message, "invalid"), strings.Contains(message, "required"):
		resp.Error(c, http.StatusBadRequest, message)
	default:
		resp.Error(c, http.StatusInternalServerError, message)
	}
}

func init() {
	router.NewGroupRouter("/api/v1/route-target").
		Use(middleware.Auth()).
		Use(middleware.RequireJSON()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(listRouteTargetOverrides),
		).
		AddRoute(
			router.NewRoute("/upsert", http.MethodPost).
				Handle(upsertRouteTargetOverride),
		).
		AddRoute(
			router.NewRoute("/delete", http.MethodPost).
				Handle(deleteRouteTargetOverride),
		)
}

func listRouteTargetOverrides(c *gin.Context) {
	channelIDParam := strings.TrimSpace(c.Query("channel_id"))
	var (
		rows []model.RouteTargetOverride
		err  error
	)
	if channelIDParam != "" {
		channelID, ok, parseErr := parseOptionalPositiveIntQuery(c, "channel_id")
		if parseErr != nil || !ok {
			resp.Error(c, http.StatusBadRequest, "invalid channel id")
			return
		}
		rows, err = op.RouteTargetOverrideListByChannel(channelID, c.Request.Context())
	} else {
		rows, err = op.RouteTargetOverrideList(c.Request.Context())
	}
	if err != nil {
		respondRouteTargetOpError(c, err)
		return
	}
	resp.Success(c, rows)
}

func upsertRouteTargetOverride(c *gin.Context) {
	var req model.RouteTargetOverrideUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	row, err := op.RouteTargetOverrideUpsert(model.RouteTargetOverride{
		ChannelID:             req.ChannelID,
		ChannelKeyID:          req.ChannelKeyID,
		ModelName:             req.ModelName,
		BillingMode:           req.BillingMode,
		ProbePolicy:           req.ProbePolicy,
		ProbeIntervalSeconds:  req.ProbeIntervalSeconds,
		ProbeConcurrencyLimit: req.ProbeConcurrencyLimit,
	}, c.Request.Context())
	if err != nil {
		respondRouteTargetOpError(c, err)
		return
	}
	resp.Success(c, row)
}

func deleteRouteTargetOverride(c *gin.Context) {
	var req model.RouteTargetOverrideDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := op.RouteTargetOverrideDelete(req.ChannelID, req.ChannelKeyID, req.ModelName, c.Request.Context()); err != nil {
		respondRouteTargetOpError(c, err)
		return
	}
	resp.Success(c, nil)
}
