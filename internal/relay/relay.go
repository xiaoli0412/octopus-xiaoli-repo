package relay

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmaxmax/go-sse"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/helper"
	dbmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/relay/balancer"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/inbound"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/xstrings"
)

// Handler 处理入站请求并转发到上游服务
func Handler(inboundType inbound.InboundType, c *gin.Context) {
	// 解析请求
	internalRequest, inAdapter, err := parseRequest(inboundType, c)
	if err != nil {
		return
	}
	supportedModels := c.GetString("supported_models")
	if supportedModels != "" {
		if !slices.Contains(xstrings.SplitTrimCompact(",", supportedModels), strings.TrimSpace(internalRequest.Model)) {
			resp.Error(c, http.StatusBadRequest, "model not allowed for this API key")
			return
		}
	}

	requestModel := internalRequest.Model
	apiKeyID := c.GetInt("api_key_id")

	// 已禁用的模型不再参与路由
	if disabled, _ := op.IsModelDisabled(c.Request.Context(), requestModel); disabled {
		resp.Error(c, http.StatusNotFound, "model disabled")
		return
	}

	// 获取通道分组
	group, err := op.GroupGetMap(requestModel, c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusNotFound, "model not found")
		return
	}

	requestCapability := requestCapabilityForInternalRequest(internalRequest)
	modeState := initDynamicRoutingModeState(group, requestModel, requestCapability)

	// 创建迭代器（策略排序 + 粘性优先，可被动态模式覆盖）
	iter := dynamicIterator(group, apiKeyID, requestModel, modeState)
	if iter.Len() == 0 {
		resp.Error(c, http.StatusServiceUnavailable, "no available channel")
		return
	}

	// 初始化 Metrics
	metrics := NewRelayMetrics(apiKeyID, requestModel, internalRequest)
	metrics.SetClientIP(op.ClientIPFromRequest(c.Request))

	// 请求级上下文
	req := &relayRequest{
		c:               c,
		inAdapter:       inAdapter,
		internalRequest: internalRequest,
		isStreaming:     isStreamingRequest(internalRequest),
		isEmbedding:     internalRequest.IsEmbeddingRequest(),
		isChat:          internalRequest.IsChatRequest(),
		metrics:         metrics,
		apiKeyID:        apiKeyID,
		requestModel:    requestModel,
		iter:            iter,
	}
	modeState.apply(req)

	var lastErr error
	retryRounds := group.RetryRounds
	if retryRounds <= 0 {
		retryRounds = 1
	}
	failoverStartedAt := time.Now()
	consecutiveFails := 0

roundLoop:
	for round := 1; round <= retryRounds; round++ {
		if round > 1 {
			if failoverWindowExceeded(group, failoverStartedAt) {
				lastErr = fmt.Errorf("failover window exceeded")
				break
			}
			if !waitForRetryDelay(c.Request.Context(), time.Duration(group.RetryDelayMs)*time.Millisecond) {
				log.Infof("request context canceled while waiting for retry delay")
				metrics.Save(c.Request.Context(), false, context.Canceled, iter.Attempts())
				return
			}
			iter.Reset()
		}

		for iter.Next() {
			select {
			case <-c.Request.Context().Done():
				log.Infof("request context canceled, stopping retry")
				metrics.Save(c.Request.Context(), false, context.Canceled, iter.Attempts())
				return
			default:
			}

			if failoverWindowExceeded(group, failoverStartedAt) {
				lastErr = fmt.Errorf("failover window exceeded")
				break roundLoop
			}

			item := iter.Item()
			targetModel := strings.TrimSpace(item.ModelName)
			if targetModel == "" {
				targetModel = requestModel
			}

			// 获取通道
			channel, err := op.ChannelGet(item.ChannelID, c.Request.Context())
			if err != nil {
				log.Warnf("failed to get channel %d: %v", item.ChannelID, err)
				iter.Skip(item.ChannelID, 0, fmt.Sprintf("channel_%d", item.ChannelID), fmt.Sprintf("channel not found: %v", err))
				lastErr = err
				continue
			}
			if !channel.Enabled {
				iter.Skip(channel.ID, 0, channel.Name, "channel disabled")
				continue
			}
			requestFormat := req.requestCapabilityFor(channel, targetModel)
			if !channel.HasConfiguredKeyForRequest(targetModel, requestFormat) {
				iter.Skip(channel.ID, 0, channel.Name, fmt.Sprintf("stale route item: channel does not declare model or has no configured key for model %s and request format %s", targetModel, requestFormat))
				continue
			}

			outAdapter := outbound.GetForModel(channel.Type, targetModel)
			if outAdapter == nil {
				iter.Skip(channel.ID, 0, channel.Name, fmt.Sprintf("unsupported channel type: %d", channel.Type))
				continue
			}

			if req.isEmbedding && !outbound.IsEmbeddingChannelType(channel.Type) {
				iter.Skip(channel.ID, 0, channel.Name, "channel type not compatible with embedding request")
				continue
			}
			if req.isChat && !outbound.IsChatChannelType(channel.Type) {
				iter.Skip(channel.ID, 0, channel.Name, "channel type not compatible with chat request")
				continue
			}

			excludedKeys := make(map[int]struct{})
			for {
				if failoverWindowExceeded(group, failoverStartedAt) {
					lastErr = fmt.Errorf("failover window exceeded")
					break roundLoop
				}

				usedKey := channel.GetChannelKeyForRequestExcept(targetModel, requestFormat, excludedKeys)
				if strings.TrimSpace(usedKey.ChannelKey) == "" {
					if len(excludedKeys) == 0 {
						iter.Skip(channel.ID, 0, channel.Name, "no available key")
					}
					break
				}

				// 熔断检查
				if iter.SkipCircuitBreak(channel.ID, usedKey.ID, channel.Name) {
					excludedKeys[usedKey.ID] = struct{}{}
					continue
				}

				// 设置实际模型
				internalRequest.Model = targetModel

				log.Infof("request model %s, mode: %d, forwarding to channel: %s model: %s (attempt %d/%d, round %d/%d, sticky=%t)",
					requestModel, group.Mode, channel.Name, targetModel,
					iter.Index()+1, iter.Len(), round, retryRounds, iter.IsSticky())

				// 构造尝试级上下文 -- 只写变化的 4 个字段
				ra := &relayAttempt{
					relayRequest:         req,
					outAdapter:           outAdapter,
					channel:              channel,
					usedKey:              usedKey,
					firstTokenTimeOutSec: group.FirstTokenTimeOut,
				}

				result := ra.attempt()
				if result.Success {
					consecutiveFails = 0
					recordRelayTokensForRateLimit(c, metrics)
					metrics.Save(c.Request.Context(), true, nil, iter.Attempts())
					return
				}
				if result.Written {
					recordRelayTokensForRateLimit(c, metrics)
					metrics.Save(c.Request.Context(), false, result.Err, iter.Attempts())
					return
				}
				lastErr = result.Err
				consecutiveFails++
				excludedKeys[usedKey.ID] = struct{}{}

				nextKey := channel.GetChannelKeyForRequestExcept(targetModel, requestFormat, excludedKeys)
				if !req.isStreaming && strings.TrimSpace(nextKey.ChannelKey) == "" {
					policy := op.ResolveRouteTargetPolicy(channel, usedKey, targetModel)
					tuning := effectiveDynamicRoutingTuningForMode(group, policy, req.dynamicMode)
					if !shouldEscalateToRaceWithMode(group, policy, consecutiveFails, tuning, req.dynamicMode) {
						continue
					}
					var raceDeadline time.Time
					if group.FailoverWindowSec > 0 {
						raceDeadline = failoverStartedAt.Add(time.Duration(group.FailoverWindowSec) * time.Second)
					}
					raceOutcome := runRaceFallbackWindow(req, iter, raceDeadline, tuning.RaceConcurrency)
					if raceOutcome.executed {
						if raceOutcome.consumedToIndex > iter.Index() {
							for iter.Index() < raceOutcome.consumedToIndex {
								if !iter.Next() {
									break
								}
							}
						}
						if raceOutcome.result.Success {
							if err := req.finalizeRaceFallbackSuccess(raceOutcome.result); err == nil {
								consecutiveFails = 0
								recordRelayTokensForRateLimit(c, metrics)
								metrics.Save(c.Request.Context(), true, nil, iter.Attempts())
								return
							} else {
								lastErr = err
							}
						}
					}
				}
			}
		}
	}

	// 所有通道都失败
	if lastErr == nil && failoverWindowExceeded(group, failoverStartedAt) {
		lastErr = fmt.Errorf("failover window exceeded")
	}
	recordRelayTokensForRateLimit(c, metrics)
	metrics.Save(c.Request.Context(), false, lastErr, iter.Attempts())
	resp.Error(c, http.StatusBadGateway, "all channels failed")
}

func failoverWindowExceeded(group dbmodel.Group, startedAt time.Time) bool {
	if group.FailoverWindowSec <= 0 {
		return false
	}
	return time.Since(startedAt) >= time.Duration(group.FailoverWindowSec)*time.Second
}

// recordRelayTokensForRateLimit 把本次请求实际消耗的 token 写回 request context，
// 供 auth 中间件在请求结束后进行 TPM 限流统计。
func recordRelayTokensForRateLimit(c *gin.Context, metrics *RelayMetrics) {
	if metrics == nil || metrics.InternalResponse == nil || metrics.InternalResponse.Usage == nil {
		return
	}
	usage := metrics.InternalResponse.Usage
	total := usage.PromptTokens + usage.CompletionTokens
	if total <= 0 {
		return
	}
	c.Request = c.Request.WithContext(
		middleware.SetRelayTokensInContext(c.Request.Context(), total),
	)
}

// attempt 统一管理一次通道尝试的完整生命周期
func (ra *relayAttempt) attempt() attemptResult {
	span := ra.iter.StartAttempt(ra.channel.ID, ra.usedKey.ID, ra.channel.Name)

	// 转发请求
	statusCode, fwdErr := ra.forward()

	// 更新 channel key 状态
	ra.usedKey.StatusCode = statusCode
	ra.usedKey.LastUseTimeStamp = time.Now().Unix()

	if fwdErr == nil {
		// ====== 成功 ======
		ra.collectResponse()
		ra.usedKey.TotalCost += ra.metrics.Stats.InputCost + ra.metrics.Stats.OutputCost
		op.ChannelKeyUpdate(ra.usedKey)

		span.End(dbmodel.AttemptSuccess, statusCode, "")

		// Channel 维度统计
		op.StatsChannelUpdate(ra.channel.ID, dbmodel.StatsMetrics{
			WaitTime:       span.Duration().Milliseconds(),
			RequestSuccess: 1,
		})

		// 熔断器：记录成功
		balancer.RecordSuccess(ra.channel.ID, ra.usedKey.ID, ra.internalRequest.Model)
		recordDynamicRouteLearning(ra.c.Request.Context(), op.DynamicRouteLearningObservation{
			ChannelID: ra.channel.ID,
			KeyID:     ra.usedKey.ID,
			ModelName: ra.internalRequest.Model,
			Success:   true,
			LatencyMs: span.Duration().Milliseconds(),
		})
		// 会话保持：更新粘性记录
		balancer.SetSticky(ra.apiKeyID, ra.requestModel, ra.channel.ID, ra.usedKey.ID)

		return attemptResult{Success: true}
	}

	// ====== 失败 ======
	op.ChannelKeyUpdate(ra.usedKey)
	span.End(dbmodel.AttemptFailed, statusCode, fwdErr.Error())

	// Channel 维度统计
	op.StatsChannelUpdate(ra.channel.ID, dbmodel.StatsMetrics{
		WaitTime:      span.Duration().Milliseconds(),
		RequestFailed: 1,
	})

	// 熔断器：记录失败
	balancer.RecordFailure(ra.channel.ID, ra.usedKey.ID, ra.internalRequest.Model)
	recordDynamicRouteLearning(ra.c.Request.Context(), op.DynamicRouteLearningObservation{
		ChannelID: ra.channel.ID,
		KeyID:     ra.usedKey.ID,
		ModelName: ra.internalRequest.Model,
		Success:   false,
		LatencyMs: span.Duration().Milliseconds(),
		Message:   fwdErr.Error(),
	})

	written := ra.c.Writer.Written()
	if written {
		ra.collectResponse()
	}
	return attemptResult{
		Success: false,
		Written: written,
		Err:     fmt.Errorf("channel %s failed: %v", ra.channel.Name, fwdErr),
	}
}

// parseRequest 解析并验证入站请求
func parseRequest(inboundType inbound.InboundType, c *gin.Context) (*model.InternalLLMRequest, model.Inbound, error) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxRelayRequestBodyBytes+1))
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "internal server error")
		return nil, nil, err
	}
	if int64(len(body)) > maxRelayRequestBodyBytes {
		err = fmt.Errorf("request body too large")
		resp.Error(c, http.StatusRequestEntityTooLarge, err.Error())
		return nil, nil, err
	}

	inAdapter := inbound.Get(inboundType)
	if inAdapter == nil {
		err = fmt.Errorf("unsupported inbound type: %d", inboundType)
		resp.Error(c, http.StatusBadRequest, err.Error())
		return nil, nil, err
	}
	internalRequest, err := inAdapter.TransformRequest(c.Request.Context(), body)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return nil, nil, err
	}

	// Pass through the original query parameters
	internalRequest.Query = c.Request.URL.Query()
	if internalRequest.RawAPIFormat == "" {
		internalRequest.RawAPIFormat = rawAPIFormatForInbound(inboundType)
	}

	if err := internalRequest.Validate(); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return nil, nil, err
	}

	return internalRequest, inAdapter, nil
}

func rawAPIFormatForInbound(inboundType inbound.InboundType) model.APIFormat {
	switch inboundType {
	case inbound.InboundTypeOpenAIChat:
		return model.APIFormatOpenAIChatCompletion
	case inbound.InboundTypeOpenAIResponse:
		return model.APIFormatOpenAIResponse
	case inbound.InboundTypeOpenAIEmbedding:
		return model.APIFormatOpenAIEmbedding
	case inbound.InboundTypeAnthropic:
		return model.APIFormatAnthropicMessage
	case inbound.InboundTypeGemini:
		return model.APIFormatGeminiContents
	default:
		return ""
	}
}

// forward 转发请求到上游服务
func (ra *relayAttempt) forward() (int, error) {
	ctx := ra.c.Request.Context()

	// 构建出站请求
	outboundRequest, err := ra.outAdapter.TransformRequest(
		ctx,
		ra.internalRequest,
		ra.channel.GetBaseUrl(),
		ra.usedKey.ChannelKey,
	)
	if err != nil {
		log.Warnf("failed to create request: %v", err)
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// 复制请求头
	ra.copyHeaders(outboundRequest)

	// 发送请求
	response, err := ra.sendRequest(outboundRequest)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	// 检查响应状态
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, err := io.ReadAll(io.LimitReader(response.Body, maxUpstreamErrorBodyBytes))
		if err != nil {
			return response.StatusCode, fmt.Errorf("failed to read response body: %w", err)
		}
		return response.StatusCode, fmt.Errorf("upstream error: %d: %s", response.StatusCode, string(body))
	}

	// 处理响应
	if ra.internalRequest.Stream != nil && *ra.internalRequest.Stream {
		if err := ra.handleStreamResponse(ctx, response); err != nil {
			return response.StatusCode, err
		}
		return response.StatusCode, nil
	}
	if err := ra.handleResponse(ctx, response); err != nil {
		return response.StatusCode, err
	}
	return response.StatusCode, nil
}

// copyHeaders 复制请求头，过滤 hop-by-hop 头
func (ra *relayAttempt) copyHeaders(outboundRequest *http.Request) {
	for key, values := range ra.c.Request.Header {
		if hopByHopHeaders[strings.ToLower(key)] {
			continue
		}
		for _, value := range values {
			outboundRequest.Header.Set(key, value)
		}
	}
	if len(ra.channel.CustomHeader) > 0 {
		for _, header := range ra.channel.CustomHeader {
			outboundRequest.Header.Set(header.HeaderKey, header.HeaderValue)
		}
	}
}

// sendRequest 发送 HTTP 请求，并注入 overall timeout 以治理上游请求全生命周期。
func (ra *relayAttempt) sendRequest(req *http.Request) (*http.Response, error) {
	timeout := defaultUpstreamTimeout
	if ra.isStreaming {
		timeout = defaultStreamingUpstreamTimeout
	}

	httpClient, err := helper.ChannelHttpClientWithTimeout(ra.channel, timeout)
	if err != nil {
		log.Warnf("failed to get http client: %v", err)
		return nil, err
	}

	response, err := httpClient.Do(req)
	if err != nil {
		log.Warnf("failed to send request: %v", err)
		return nil, err
	}

	return response, nil
}

// handleStreamResponse 处理流式响应
func (ra *relayAttempt) handleStreamResponse(ctx context.Context, response *http.Response) error {
	if ct := response.Header.Get("Content-Type"); ct != "" && !strings.Contains(strings.ToLower(ct), "text/event-stream") {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 16*1024))
		return fmt.Errorf("upstream returned non-SSE content-type %q for stream request: %s", ct, string(body))
	}

	// 设置 SSE 响应头
	ra.c.Header("Content-Type", "text/event-stream")
	ra.c.Header("Cache-Control", "no-cache")
	ra.c.Header("Connection", "keep-alive")
	ra.c.Header("X-Accel-Buffering", "no")

	firstToken := true

	type sseReadResult struct {
		data string
		err  error
	}
	results := make(chan sseReadResult, 1)
	stopReading := make(chan struct{})
	go func() {
		defer close(results)
		readCfg := &sse.ReadConfig{MaxEventSize: maxSSEEventSize}
		for ev, err := range sse.Read(response.Body, readCfg) {
			if err != nil {
				select {
				case results <- sseReadResult{err: err}:
				case <-stopReading:
				}
				return
			}
			select {
			case results <- sseReadResult{data: ev.Data}:
			case <-stopReading:
				return
			}
		}
	}()
	defer close(stopReading)

	var firstTokenTimer *time.Timer
	var firstTokenC <-chan time.Time
	if firstToken && ra.firstTokenTimeOutSec > 0 {
		firstTokenTimer = time.NewTimer(time.Duration(ra.firstTokenTimeOutSec) * time.Second)
		firstTokenC = firstTokenTimer.C
		defer func() {
			if firstTokenTimer != nil {
				firstTokenTimer.Stop()
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			log.Infof("client disconnected, stopping stream")
			return nil
		case <-firstTokenC:
			log.Warnf("first token timeout (%ds), switching channel", ra.firstTokenTimeOutSec)
			_ = response.Body.Close()
			return fmt.Errorf("first token timeout (%ds)", ra.firstTokenTimeOutSec)
		case r, ok := <-results:
			if !ok {
				log.Infof("stream end")
				return nil
			}
			if r.err != nil {
				log.Warnf("failed to read event: %v", r.err)
				return fmt.Errorf("failed to read stream event: %w", r.err)
			}

			data, err := ra.transformStreamData(ctx, r.data)
			if err != nil {
				log.Warnf("failed to transform stream data: %v", err)
				return fmt.Errorf("failed to transform stream data: %w", err)
			}
			if len(data) == 0 {
				continue
			}
			if firstToken {
				ra.metrics.SetFirstTokenTime(time.Now())
				firstToken = false
				if firstTokenTimer != nil {
					if !firstTokenTimer.Stop() {
						select {
						case <-firstTokenTimer.C:
						default:
						}
					}
					firstTokenTimer = nil
					firstTokenC = nil
				}
			}

			if _, err := ra.c.Writer.Write(data); err != nil {
				log.Infof("client disconnected during stream write: %v", err)
				return nil
			}
			ra.c.Writer.Flush()
		}
	}
}

// transformStreamData 转换流式数据
func (ra *relayAttempt) transformStreamData(ctx context.Context, data string) ([]byte, error) {
	internalStream, err := ra.outAdapter.TransformStream(ctx, []byte(data))
	if err != nil {
		log.Warnf("failed to transform stream: %v", err)
		return nil, err
	}
	if internalStream == nil {
		return nil, nil
	}

	inStream, err := ra.inAdapter.TransformStream(ctx, internalStream)
	if err != nil {
		log.Warnf("failed to transform stream: %v", err)
		return nil, err
	}

	return inStream, nil
}

// handleResponse 处理非流式响应
func (ra *relayAttempt) handleResponse(ctx context.Context, response *http.Response) error {
	internalResponse, err := ra.outAdapter.TransformResponse(ctx, response)
	if err != nil {
		log.Warnf("failed to transform response: %v", err)
		return fmt.Errorf("failed to transform outbound response: %w", err)
	}

	inResponse, err := ra.inAdapter.TransformResponse(ctx, internalResponse)
	if err != nil {
		log.Warnf("failed to transform response: %v", err)
		return fmt.Errorf("failed to transform inbound response: %w", err)
	}

	ra.c.Data(http.StatusOK, "application/json", inResponse)
	return nil
}

// collectResponse 收集响应信息
func (ra *relayAttempt) collectResponse() {
	internalResponse, err := ra.inAdapter.GetInternalResponse(ra.c.Request.Context())
	if err != nil || internalResponse == nil {
		return
	}

	ra.metrics.SetInternalResponse(internalResponse, ra.internalRequest.Model)
}
