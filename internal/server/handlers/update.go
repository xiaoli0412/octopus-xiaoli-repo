package handlers

import (
	"errors"
	"net/http"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/update"
	"github.com/gin-gonic/gin"
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
			router.NewRoute("", http.MethodPost).
				Use(middleware.RequireJSON()).
				Handle(updateFunc),
		)
}

func latest(c *gin.Context) {
	latestInfo, err := update.GetLatestInfo()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, *latestInfo)
}

func getNowVersion(c *gin.Context) {
	resp.Success(c, conf.Version)
}

func updateFunc(c *gin.Context) {
	err := runUpdateCore()
	if err != nil {
		if errors.Is(err, update.ErrUpdateInProgress) {
			resp.Error(c, http.StatusConflict, err.Error())
			return
		}
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, "update success")
}
