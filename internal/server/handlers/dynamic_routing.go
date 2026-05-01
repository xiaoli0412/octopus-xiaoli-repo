package handlers

import (
	"net/http"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/dynamic-routing").
		Use(middleware.Auth()).
		AddRoute(router.NewRoute("/learning", http.MethodGet).Handle(getDynamicRouteLearning)).
		AddRoute(router.NewRoute("/learning/reset", http.MethodPost).Use(middleware.RequireJSON()).Handle(resetDynamicRouteLearning))
}

func getDynamicRouteLearning(c *gin.Context) {
	result, err := op.DynamicRouteLearningList(c.Request.Context())
	if err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, result)
}

func resetDynamicRouteLearning(c *gin.Context) {
	if err := op.DynamicRouteLearningReset(c.Request.Context()); err != nil {
		respondAIAutomationError(c, err)
		return
	}
	resp.Success(c, nil)
}
