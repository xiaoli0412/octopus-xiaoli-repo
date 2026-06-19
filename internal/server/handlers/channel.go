package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/helper"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/task"
	transformerOutbound "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

const channelPostSaveTaskMaxConcurrent = 4

var channelPostSaveTaskSlots = make(chan struct{}, channelPostSaveTaskMaxConcurrent)
var channelPostSaveTaskTimeout = 5 * time.Minute
var channelPostSaveTaskRunner = func(channel *model.Channel, ctx context.Context) {
	if err := helper.EnsureReferencedLLMInfos(ctx); err != nil {
		log.Warnf("post-save llm price update failed (channel=%d): %v", channel.ID, err)
	}
	helper.ChannelBaseUrlDelayUpdate(channel, ctx)
	helper.ChannelAutoGroup(channel, ctx)
}

func scheduleChannelPostSaveTask(channel *model.Channel) bool {
	if channel == nil {
		return false
	}
	select {
	case channelPostSaveTaskSlots <- struct{}{}:
		go func(channel *model.Channel) {
			runner := channelPostSaveTaskRunner
			timeout := channelPostSaveTaskTimeout
			slots := channelPostSaveTaskSlots
			defer func() { <-slots }()
			ctx, cancel := task.DetachedContextWithTimeout(timeout)
			defer cancel()
			runner(channel, ctx)
		}(channel)
		return true
	default:
		log.Warnf("post-save channel maintenance skipped because the background queue is full (channel=%d)", channel.ID)
		return false
	}
}

func respondChannelOpError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	message := err.Error()
	switch {
	case errors.Is(err, context.Canceled):
		resp.Error(c, http.StatusRequestTimeout, message)
	case strings.Contains(message, "not found"):
		resp.Error(c, http.StatusNotFound, message)
	case strings.Contains(message, "invalid"), strings.Contains(message, "must be"), strings.Contains(message, "must not"):
		resp.Error(c, http.StatusBadRequest, message)
	default:
		resp.Error(c, http.StatusInternalServerError, "internal server error")
	}
}

func validateChannelMatchRegex(matchRegex *string) error {
	if matchRegex == nil || strings.TrimSpace(*matchRegex) == "" {
		return nil
	}
	_, err := regexp2.Compile(*matchRegex, regexp2.ECMAScript)
	return err
}

func init() {
	router.NewGroupRouter("/api/v1/channel").
		Use(middleware.Auth()).
		Use(middleware.RequireJSON()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(listChannel),
		).
		AddRoute(
			router.NewRoute("/create", http.MethodPost).
				Handle(createChannel),
		).
		AddRoute(
			router.NewRoute("/update", http.MethodPost).
				Handle(updateChannel),
		).
		AddRoute(
			router.NewRoute("/enable", http.MethodPost).
				Handle(enableChannel),
		).
		AddRoute(
			router.NewRoute("/delete/:id", http.MethodDelete).
				Handle(deleteChannel),
		).
		AddRoute(
			router.NewRoute("/fetch-model", http.MethodPost).
				Handle(fetchModel),
		).
		AddRoute(
			router.NewRoute("/newapi/inspect", http.MethodPost).
				Handle(inspectNewAPI),
		).
		AddRoute(
			router.NewRoute("/upstream/inspect", http.MethodPost).
				Handle(inspectUpstreamGateway),
		).
		AddRoute(
			router.NewRoute("/upstream/apply", http.MethodPost).
				Handle(applyUpstreamGateway),
		)
	router.NewGroupRouter("/api/v1/channel").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/sync", http.MethodPost).
				Use(middleware.RequireJSON()).
				Handle(syncChannel),
		).
		AddRoute(
			router.NewRoute("/last-sync-time", http.MethodGet).
				Handle(getLastSyncTime),
		).
		AddRoute(
			router.NewRoute("/test-models", http.MethodPost).
				Use(middleware.RequireJSON()).
				Handle(testChannelModels),
		).
		AddRoute(
			router.NewRoute("/test-models-by-config", http.MethodPost).
				Use(middleware.RequireJSON()).
				Handle(testChannelModelsByConfig),
		)
}

func listChannel(c *gin.Context) {
	channels, err := op.ChannelList(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to list channels")
		return
	}
	for i, channel := range channels {
		stats := op.StatsChannelGet(channel.ID)
		channels[i].Stats = &stats
	}
	resp.Success(c, channels)
}

func createChannel(c *gin.Context) {
	var channel model.Channel
	if err := c.ShouldBindJSON(&channel); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := validateChannelMatchRegex(channel.MatchRegex); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := op.ChannelCreate(&channel, c.Request.Context()); err != nil {
		respondChannelOpError(c, err)
		return
	}
	stats := op.StatsChannelGet(channel.ID)
	channel.Stats = &stats
	scheduleChannelPostSaveTask(&channel)
	resp.Success(c, channel)
}

func updateChannel(c *gin.Context) {
	var req model.ChannelUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := validateChannelMatchRegex(req.MatchRegex); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	channel, err := op.ChannelUpdate(&req, c.Request.Context())
	if err != nil {
		respondChannelOpError(c, err)
		return
	}
	stats := op.StatsChannelGet(channel.ID)
	channel.Stats = &stats
	scheduleChannelPostSaveTask(channel)
	resp.Success(c, channel)
}

func enableChannel(c *gin.Context) {
	var request struct {
		ID      int  `json:"id"`
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if request.ID <= 0 {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidParam)
		return
	}
	if err := op.ChannelEnabled(request.ID, request.Enabled, c.Request.Context()); err != nil {
		respondChannelOpError(c, err)
		return
	}
	resp.Success(c, nil)
}

func deleteChannel(c *gin.Context) {
	idNum, ok := parsePositivePathIDValue(c, "id")
	if !ok {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidParam)
		return
	}
	if err := op.ChannelDel(idNum, c.Request.Context()); err != nil {
		respondChannelOpError(c, err)
		return
	}
	resp.Success(c, nil)
}

func fetchModel(c *gin.Context) {
	var request model.ChannelFetchModelRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := validateAbsoluteHTTPURL(request.BaseURL, "base_url"); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	request.ChannelProxy = model.NormalizeChannelProxy(request.ChannelProxy)
	if err := model.ValidateChannelProxy(request.ChannelProxy); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	channel := model.Channel{
		Type:         request.Type,
		BaseUrls:     []model.BaseUrl{{URL: strings.TrimSpace(request.BaseURL), Delay: 0}},
		Proxy:        request.Proxy,
		ChannelProxy: request.ChannelProxy,
		CustomHeader: request.CustomHeader,
		MatchRegex:   request.MatchRegex,
		Keys: []model.ChannelKey{{
			Enabled:    true,
			ChannelKey: strings.TrimSpace(request.Key),
		}},
	}
	models, err := helper.FetchModels(c.Request.Context(), channel)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to fetch models from upstream")
		return
	}
	resp.Success(c, models)
}

func inspectNewAPI(c *gin.Context) {
	var request model.NewAPIInspectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	result, err := op.InspectNewAPI(c.Request.Context(), request)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	resp.Success(c, result)
}

func inspectUpstreamGateway(c *gin.Context) {
	var request model.UpstreamInspectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	result, err := op.InspectUpstreamGateway(c.Request.Context(), request)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	resp.Success(c, result)
}

func applyUpstreamGateway(c *gin.Context) {
	var request model.UpstreamApplyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	result, err := op.ApplyUpstreamGateway(c.Request.Context(), request)
	if err != nil {
		respondChannelOpError(c, err)
		return
	}
	scheduleChannelPostSaveTask(&result.Channel)
	resp.Success(c, result)
}

func syncChannel(c *gin.Context) {
	task.SyncModelsTask()
	resp.Success(c, nil)
}

func getLastSyncTime(c *gin.Context) {
	time := task.GetLastSyncModelsTime()
	resp.Success(c, time)
}

func testChannelModels(c *gin.Context) {
	type TestModelRequest struct {
		ChannelID int      `json:"channel_id"`
		Models    []string `json:"models"`
	}
	var req TestModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if req.ChannelID <= 0 {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidParam)
		return
	}

	if len(req.Models) == 0 {
		resp.Error(c, http.StatusBadRequest, "models is required")
		return
	}

	channel, err := op.ChannelGet(req.ChannelID, c.Request.Context())
	if err != nil {
		respondChannelOpError(c, err)
		return
	}
	results := make([]model.DBImportHealthCheckItem, 0, len(req.Models))
	for _, modelName := range req.Models {
		results = append(results, op.CheckChannelModelHealth(c.Request.Context(), channel, modelName))
	}
	c.JSON(http.StatusOK, results)
}

func testChannelModelsByConfig(c *gin.Context) {
	type TestModelByConfigRequest struct {
		Type     transformerOutbound.OutboundType `json:"type"`
		Enabled  *bool                            `json:"enabled"`
		BaseUrls []model.BaseUrl                  `json:"base_urls"`
		Keys     []struct {
			Enabled             bool   `json:"enabled"`
			ChannelKey          string `json:"channel_key"`
			SourceType          string `json:"source_type"`
			AllowedModels       string `json:"allowed_models"`
			RequestCapabilities string `json:"request_capabilities"`
		} `json:"keys"`
		Proxy             bool                    `json:"proxy"`
		ChannelProxy      *string                 `json:"channel_proxy"`
		CustomHeader      []model.CustomHeader    `json:"custom_header"`
		KeyManagementMode model.KeyManagementMode `json:"key_management_mode"`
		KeyRoutingPolicy  model.KeyRoutingPolicy  `json:"key_routing_policy"`
		Models            []string                `json:"models"`
	}

	var req TestModelByConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	for _, baseURL := range req.BaseUrls {
		if err := validateAbsoluteHTTPURL(baseURL.URL, "base_url"); err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	}
	req.ChannelProxy = model.NormalizeChannelProxy(req.ChannelProxy)
	if err := model.ValidateChannelProxy(req.ChannelProxy); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if len(req.Models) == 0 {
		resp.Error(c, http.StatusBadRequest, "models is required")
		return
	}

	channel := &model.Channel{
		Type:              req.Type,
		Enabled:           req.Enabled == nil || *req.Enabled,
		BaseUrls:          req.BaseUrls,
		Proxy:             req.Proxy,
		ChannelProxy:      req.ChannelProxy,
		CustomHeader:      req.CustomHeader,
		KeyManagementMode: req.KeyManagementMode,
		KeyRoutingPolicy:  req.KeyRoutingPolicy,
	}
	for _, k := range req.Keys {
		channel.Keys = append(channel.Keys, model.ChannelKey{
			Enabled:             k.Enabled,
			ChannelKey:          k.ChannelKey,
			SourceType:          k.SourceType,
			AllowedModels:       model.NormalizeChannelKeyAllowedModels(k.AllowedModels),
			RequestCapabilities: model.NormalizeChannelKeyRequestCapabilities(k.RequestCapabilities),
		})
	}

	results := make([]model.DBImportHealthCheckItem, 0, len(req.Models))
	for _, modelName := range req.Models {
		results = append(results, op.CheckChannelModelHealthForConfig(c.Request.Context(), channel, modelName))
	}

	c.JSON(http.StatusOK, results)
}
