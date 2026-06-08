package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/helper"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/price"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
)

func init() {
	router.NewGroupRouter("/api/v1/model").
		Use(middleware.Auth()).
		Use(middleware.RequireJSON()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(listLLM),
		).
		AddRoute(
			router.NewRoute("/create", http.MethodPost).
				Handle(createLLM),
		).
		AddRoute(
			router.NewRoute("/channel", http.MethodGet).
				Handle(listLLMByChannel),
		).
		AddRoute(
			router.NewRoute("/capability-inventory", http.MethodGet).
				Handle(getCapabilityInventory),
		).
		AddRoute(
			router.NewRoute("/upstream-prices", http.MethodGet).
				Handle(listUpstreamPrices),
		).
		AddRoute(
			router.NewRoute("/update", http.MethodPost).
				Handle(updateLLM),
		).
		AddRoute(
			router.NewRoute("/delete", http.MethodPost).
				Handle(deleteLLM),
		).
		AddRoute(
			router.NewRoute("/update-price", http.MethodPost).
				Handle(updateLLMPrice),
		).
		AddRoute(
			router.NewRoute("/last-update-time", http.MethodGet).
				Handle(getLastUpdateTime),
		)
	router.NewGroupRouter("/v1").
		Use(middleware.APIKeyAuth()).
		AddRoute(
			router.NewRoute("/models", http.MethodGet).
				Handle(getModelList),
		)
}

func getModelList(c *gin.Context) {
	models, err := op.GroupListModel(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to list models")
		return
	}
	apiKeyId := c.GetInt("api_key_id")
	apiKey, err := op.APIKeyGet(apiKeyId, c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to get api key")
		return
	}
	if apiKey.SupportedModels != "" {
		supportedModels := lo.Map(strings.Split(apiKey.SupportedModels, ","), func(s string, _ int) string {
			return strings.TrimSpace(s)
		})
		models = lo.Filter(models, func(m string, _ int) bool {
			return lo.Contains(supportedModels, m)
		})
	}

	if c.GetString("request_type") == "anthropic" {
		var anthropicModels []model.AnthropicModel
		for _, m := range models {
			anthropicModels = append(anthropicModels, model.AnthropicModel{
				ID:          m,
				CreatedAt:   "2024-01-01T00:00:00Z",
				DisplayName: m,
				Type:        "model",
			})
		}
		response := gin.H{
			"data":     anthropicModels,
			"has_more": false,
		}
		if len(anthropicModels) > 0 {
			response["first_id"] = anthropicModels[0].ID
			response["last_id"] = anthropicModels[len(anthropicModels)-1].ID
		}
		c.JSON(200, response)
	} else {
		var openAIModels []model.OpenAIModel
		for _, m := range models {
			openAIModels = append(openAIModels, model.OpenAIModel{
				ID:      m,
				Object:  "model",
				Created: 1763395200,
				OwnedBy: "octopus",
			})
		}
		c.JSON(200, gin.H{
			"success": true,
			"data":    openAIModels,
			"object":  "list",
		})
	}
}

func listLLM(c *gin.Context) {
	if err := helper.EnsureReferencedLLMInfos(c.Request.Context()); err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to sync model information")
		return
	}
	models, err := op.LLMList(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to list models")
		return
	}
	upstreamSummaries, _ := op.UpstreamPriceSummaries(c.Request.Context())
	upstreamByModel := make(map[string]model.UpstreamPriceSummary, len(upstreamSummaries))
	for _, summary := range upstreamSummaries {
		upstreamByModel[strings.ToLower(strings.TrimSpace(summary.ModelName))] = summary
	}
	for i := range models {
		models[i] = price.EnrichLLMInfo(models[i])
		if summary, ok := upstreamByModel[strings.ToLower(strings.TrimSpace(models[i].Name))]; ok {
			models[i].UpstreamPriceCount = len(summary.GatewayPrices)
			limit := 3
			if len(summary.GatewayPrices) < limit {
				limit = len(summary.GatewayPrices)
			}
			if limit > 0 {
				models[i].UpstreamPricePreview = summary.GatewayPrices[:limit]
			}
		}
	}
	resp.Success(c, models)
}

func listLLMByChannel(c *gin.Context) {
	channels, err := op.ChannelLLMList(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to list channel models")
		return
	}
	resp.Success(c, channels)
}

func getCapabilityInventory(c *gin.Context) {
	inventory, err := op.CapabilityInventory(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to get capability inventory")
		return
	}
	resp.Success(c, inventory)
}

func listUpstreamPrices(c *gin.Context) {
	summaries, err := op.UpstreamPriceSummaries(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to list upstream prices")
		return
	}
	resp.Success(c, summaries)
}

func createLLM(c *gin.Context) {
	var model model.LLMInfo
	if err := c.ShouldBindJSON(&model); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := op.LLMCreate(model, c.Request.Context()); err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to create model")
		return
	}
	resp.Success(c, model)
}

func updateLLM(c *gin.Context) {
	var model model.LLMInfo
	if err := c.ShouldBindJSON(&model); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := op.LLMUpdate(model, c.Request.Context()); err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to update model")
		return
	}
	resp.Success(c, model)
}

func deleteLLM(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := op.LLMDelete(req.Name, c.Request.Context()); err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to delete model")
		return
	}
	resp.Success(c, nil)
}

func updateLLMPrice(c *gin.Context) {
	err := price.UpdateLLMPrice(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to update model prices")
		return
	}
	resp.Success(c, nil)
}

func getLastUpdateTime(c *gin.Context) {
	time := price.GetLastUpdateTime()
	resp.Success(c, time)
}
