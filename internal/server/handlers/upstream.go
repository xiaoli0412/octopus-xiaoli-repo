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
	router.NewGroupRouter("/api/v1/upstream").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(listUpstreamSites),
		).
		AddRoute(
			router.NewRoute("/detail/:id", http.MethodGet).
				Handle(getUpstreamSiteDetail),
		).
		AddRoute(
			router.NewRoute("/delete/:id", http.MethodDelete).
				Handle(deleteUpstreamSite),
		)

	router.NewGroupRouter("/api/v1/upstream").
		Use(middleware.Auth()).
		Use(middleware.RequireJSON()).
		AddRoute(
			router.NewRoute("/inspect", http.MethodPost).
				Handle(inspectUpstreamSite),
		).
		AddRoute(
			router.NewRoute("/create", http.MethodPost).
				Handle(createUpstreamSite),
		).
		AddRoute(
			router.NewRoute("/update", http.MethodPost).
				Handle(updateUpstreamSite),
		).
		AddRoute(
			router.NewRoute("/refresh", http.MethodPost).
				Handle(refreshUpstreamSite),
		).
		AddRoute(
			router.NewRoute("/apply", http.MethodPost).
				Handle(applyUpstreamSite),
		)
}

func listUpstreamSites(c *gin.Context) {
	sites, err := op.UpstreamSiteList(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to list upstream sites")
		return
	}
	resp.Success(c, sites)
}

func getUpstreamSiteDetail(c *gin.Context) {
	id, ok := parsePositivePathIDValue(c, "id")
	if !ok {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidParam)
		return
	}
	detail, err := op.UpstreamSiteDetailGet(c.Request.Context(), id)
	if err != nil {
		resp.Error(c, http.StatusNotFound, err.Error())
		return
	}
	resp.Success(c, detail)
}

func inspectUpstreamSite(c *gin.Context) {
	var request model.UpstreamInspectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	result, err := op.UpstreamInspect(c.Request.Context(), request)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	resp.Success(c, result)
}

func createUpstreamSite(c *gin.Context) {
	var request model.UpstreamSiteCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	detail, err := op.UpstreamSiteCreate(c.Request.Context(), request)
	if err != nil {
		respondChannelOpError(c, err)
		return
	}
	if detail.Site.LinkedChannelID > 0 {
		if channel, getErr := op.ChannelGet(detail.Site.LinkedChannelID, c.Request.Context()); getErr == nil {
			scheduleChannelPostSaveTask(channel)
		}
	}
	resp.Success(c, detail)
}

func updateUpstreamSite(c *gin.Context) {
	var request model.UpstreamSiteUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	site, err := op.UpstreamSiteUpdate(c.Request.Context(), request)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	resp.Success(c, site)
}

func refreshUpstreamSite(c *gin.Context) {
	var request model.UpstreamRefreshRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	request.Manual = true
	detail, err := op.UpstreamSiteRefresh(c.Request.Context(), request)
	if err != nil {
		respondChannelOpError(c, err)
		return
	}
	if request.ApplyChannel && detail.Site.LinkedChannelID > 0 {
		if channel, getErr := op.ChannelGet(detail.Site.LinkedChannelID, c.Request.Context()); getErr == nil {
			scheduleChannelPostSaveTask(channel)
		}
	}
	resp.Success(c, detail)
}

func applyUpstreamSite(c *gin.Context) {
	var request struct {
		ID              int `json:"id" binding:"required"`
		TargetChannelID int `json:"target_channel_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	result, err := op.UpstreamSiteApply(c.Request.Context(), request.ID, request.TargetChannelID)
	if err != nil {
		respondChannelOpError(c, err)
		return
	}
	scheduleChannelPostSaveTask(&result.Channel)
	resp.Success(c, result)
}

func deleteUpstreamSite(c *gin.Context) {
	id, ok := parsePositivePathIDValue(c, "id")
	if !ok {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidParam)
		return
	}
	if err := op.UpstreamSiteDelete(c.Request.Context(), id); err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to delete upstream site")
		return
	}
	resp.Success(c, nil)
}
