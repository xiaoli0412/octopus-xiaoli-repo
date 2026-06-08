package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/update"
)

var runUpdateCore = update.UpdateCore

func init() {
	router.NewGroupRouter("/api/v1/update").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("", http.MethodGet).
				Handle(latest),
		).
		AddRoute(
			router.NewRoute("/now-version", http.MethodGet).
				Handle(getNowVersion),
		).
		AddRoute(
			router.NewRoute("/status", http.MethodGet).
				Handle(getUpdateStatus),
		).
		AddRoute(
			router.NewRoute("", http.MethodPost).
				Handle(updateFunc),
		)
}

func latest(c *gin.Context) {
	latestInfo, err := update.GetLatestInfo()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to check for updates")
		return
	}
	resp.Success(c, *latestInfo)
}

func getNowVersion(c *gin.Context) {
	resp.Success(c, conf.Version)
}

func getUpdateStatus(c *gin.Context) {
	resp.Success(c, update.GetStatusInfo())
}

func updateFunc(c *gin.Context) {
	err := runUpdateCore()
	if err != nil {
		if errors.Is(err, update.ErrUpdateInProgress) {
			resp.Error(c, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, update.ErrUpdateUnsupportedPlatform) {
			resp.Error(c, http.StatusNotImplemented, err.Error())
			return
		}
		resp.Error(c, http.StatusInternalServerError, "update failed")
		return
	}
	resp.Success(c, "update success")
}
