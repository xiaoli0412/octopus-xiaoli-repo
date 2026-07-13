package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/price"
	transformerModel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

const (
	relayLogMaxStringRunes  = 2048
	relayLogTruncatedSuffix = "...[truncated]"
)

// RelayMetrics 负责最终的日志收集与持久化
type RelayMetrics struct {
	APIKeyID     int
	ClientIP     string
	RequestModel string
	StartTime    time.Time

	// 首 Token 时间
	FirstTokenTime time.Time

	// 请求和响应内容
	InternalRequest  *transformerModel.InternalLLMRequest
	InternalResponse *transformerModel.InternalLLMResponse

	// 统计指标
	ActualModel      string
	Stats            model.StatsMetrics
	CacheReadTokens  int64
	CacheWriteTokens int64
	Dynamic          *relayDynamicAudit

	// skipStats 为 true 时跳过统计写入与日志持久化（用于请求重放）
	skipStats bool
}

var relayStatsDailyUpdate = op.StatsDailyUpdate

// calculateTokenCosts computes input and output costs given pricing, usage, and whether
// the request is Anthropic-style (where prompt_tokens already includes cached tokens).
func calculateTokenCosts(modelPrice model.LLMPrice, usage *transformerModel.Usage) (inputCost, outputCost float64) {
	cachedTokens := int64(0)
	if usage.PromptTokensDetails != nil {
		cachedTokens = usage.PromptTokensDetails.CachedTokens
	}

	if usage.AnthropicUsage {
		inputCost = (float64(cachedTokens)*modelPrice.CacheRead +
			float64(usage.PromptTokens)*modelPrice.Input +
			float64(usage.CacheCreationInputTokens)*modelPrice.CacheWrite) * 1e-6
	} else {
		nonCachedPromptTokens := usage.PromptTokens - cachedTokens
		if nonCachedPromptTokens < 0 {
			nonCachedPromptTokens = 0
		}
		inputCost = (float64(cachedTokens)*modelPrice.CacheRead + float64(nonCachedPromptTokens)*modelPrice.Input) * 1e-6
	}
	outputCost = float64(usage.CompletionTokens) * modelPrice.Output * 1e-6
	return inputCost, outputCost
}

func NewRelayMetrics(apiKeyID int, requestModel string, req *transformerModel.InternalLLMRequest) *RelayMetrics {
	return &RelayMetrics{
		APIKeyID:        apiKeyID,
		RequestModel:    requestModel,
		StartTime:       time.Now(),
		InternalRequest: req,
	}
}

func (m *RelayMetrics) SetClientIP(clientIP string) {
	m.ClientIP = strings.TrimSpace(clientIP)
}

func (m *RelayMetrics) SetFirstTokenTime(t time.Time) {
	m.FirstTokenTime = t
}

func (m *RelayMetrics) SetInternalResponse(resp *transformerModel.InternalLLMResponse, actualModel string) {
	m.InternalResponse = resp
	m.ActualModel = actualModel

	if resp == nil || resp.Usage == nil {
		return
	}

	usage := resp.Usage
	m.Stats.InputToken = usage.PromptTokens
	m.Stats.OutputToken = usage.CompletionTokens

	modelPrice := price.GetLLMPrice(actualModel)
	if modelPrice == nil {
		return
	}
	if usage.PromptTokensDetails == nil {
		usage.PromptTokensDetails = &transformerModel.PromptTokensDetails{
			CachedTokens: 0,
		}
	}
	m.CacheReadTokens = int64(usage.PromptTokensDetails.CachedTokens)
	m.CacheWriteTokens = int64(usage.CacheCreationInputTokens)
	m.Stats.InputCost, m.Stats.OutputCost = calculateTokenCosts(*modelPrice, usage)
}

func (m *RelayMetrics) Save(ctx context.Context, success bool, err error, attempts []model.ChannelAttempt) {
	if m != nil && m.skipStats {
		return
	}
	duration := time.Since(m.StartTime)
	channelID, channelName := finalChannel(attempts)
	m.recalculateCostsForChannel(channelID)

	globalStats := model.StatsMetrics{
		WaitTime:    duration.Milliseconds(),
		InputToken:  m.Stats.InputToken,
		OutputToken: m.Stats.OutputToken,
		InputCost:   m.Stats.InputCost,
		OutputCost:  m.Stats.OutputCost,
	}
	if success {
		globalStats.RequestSuccess = 1
	} else {
		globalStats.RequestFailed = 1
	}

	op.StatsTotalUpdate(globalStats)
	op.StatsHourlyUpdate(globalStats)
	relayStatsDailyUpdate(ctx, globalStats)
	op.StatsAPIKeyUpdate(m.APIKeyID, globalStats)
	op.StatsChannelUpdate(channelID, globalStats)

	// Async cost alert check: fire-and-forget, never blocks relay response.
	go triggerCostAlert(m.APIKeyID)
	if err := op.OpsRecordRelay(ctx, op.OpsRelayEvent{
		Time:             m.StartTime,
		APIKeyID:         m.APIKeyID,
		ClientIP:         m.ClientIP,
		RequestModel:     m.RequestModel,
		ActualModel:      m.ActualModel,
		Success:          success,
		DurationMS:       duration.Milliseconds(),
		InputTokens:      m.Stats.InputToken,
		OutputTokens:     m.Stats.OutputToken,
		CacheReadTokens:  m.CacheReadTokens,
		CacheWriteTokens: m.CacheWriteTokens,
		Attempts:         attempts,
	}); err != nil {
		log.Warnf("failed to record ops metrics: %v", err)
	}

	log.Infof("relay complete: model=%s, channel=%d(%s), success=%t, duration=%dms, input_token=%d, output_token=%d, input_cost=%f, output_cost=%f, total_cost=%f, attempts=%d",
		m.RequestModel, channelID, channelName, success, duration.Milliseconds(),
		m.Stats.InputToken, m.Stats.OutputToken,
		m.Stats.InputCost, m.Stats.OutputCost, m.Stats.InputCost+m.Stats.OutputCost,
		len(attempts))

	m.saveLog(ctx, err, duration, attempts, channelID, channelName)
}

func (m *RelayMetrics) recalculateCostsForChannel(channelID int) {
	if m == nil || m.InternalResponse == nil || m.InternalResponse.Usage == nil {
		return
	}
	actualModel := m.ActualModel
	if strings.TrimSpace(actualModel) == "" {
		actualModel = m.RequestModel
	}
	modelPrice, ok := op.ResolveGatewayLLMPrice(actualModel, channelID)
	if !ok {
		return
	}
	usage := m.InternalResponse.Usage
	if usage.PromptTokensDetails == nil {
		usage.PromptTokensDetails = &transformerModel.PromptTokensDetails{CachedTokens: 0}
	}
	m.CacheReadTokens = int64(usage.PromptTokensDetails.CachedTokens)
	m.CacheWriteTokens = int64(usage.CacheCreationInputTokens)
	m.Stats.InputCost, m.Stats.OutputCost = calculateTokenCosts(modelPrice, usage)
}

func finalChannel(attempts []model.ChannelAttempt) (int, string) {
	var lastID int
	var lastName string
	var fallbackID int
	var fallbackName string
	for i := len(attempts) - 1; i >= 0; i-- {
		a := attempts[i]
		if a.ChannelID != 0 && fallbackID == 0 {
			fallbackID = a.ChannelID
			fallbackName = a.ChannelName
		}
		if a.Status == model.AttemptSuccess && a.ChannelID != 0 {
			return a.ChannelID, a.ChannelName
		}
		if a.Status == model.AttemptFailed && a.ChannelID != 0 && lastID == 0 {
			lastID = a.ChannelID
			lastName = a.ChannelName
		}
	}
	if lastID == 0 {
		return fallbackID, fallbackName
	}
	return lastID, lastName
}

func (m *RelayMetrics) saveLog(ctx context.Context, err error, duration time.Duration, attempts []model.ChannelAttempt, channelID int, channelName string) {
	actualModel := m.ActualModel
	if actualModel == "" {
		actualModel = m.RequestModel
	}

	relayLog := model.RelayLog{
		Time:             m.StartTime.Unix(),
		APIKeyID:         m.APIKeyID,
		RequestModelName: m.RequestModel,
		ChannelName:      channelName,
		ChannelId:        channelID,
		ActualModelName:  actualModel,
		UseTime:          int(duration.Milliseconds()),
		Attempts:         attempts,
		TotalAttempts:    len(attempts),
	}
	if m.Dynamic != nil {
		relayLog.DynamicRoutingMode = m.Dynamic.Mode
		relayLog.DynamicRoutingEffectiveMode = m.Dynamic.EffectiveMode
		relayLog.DynamicRoutingDecision = m.Dynamic.Decision
		relayLog.DynamicRoutingReason = m.Dynamic.Reason
		relayLog.DynamicRoutingConfidence = m.Dynamic.Confidence
		relayLog.DynamicRoutingFallback = m.Dynamic.Fallback
		relayLog.DynamicRoutingRecommended = strings.Join(m.Dynamic.Recommended, ",")
	}

	// 首字时间
	if !m.FirstTokenTime.IsZero() {
		relayLog.Ftut = int(m.FirstTokenTime.Sub(m.StartTime).Milliseconds())
	}

	// Usage
	if m.InternalResponse != nil && m.InternalResponse.Usage != nil {
		usage := m.InternalResponse.Usage
		relayLog.InputTokens = int(usage.PromptTokens)
		relayLog.OutputTokens = int(usage.CompletionTokens)
		if usage.PromptTokensDetails != nil {
			relayLog.CacheReadTokens = int(usage.PromptTokensDetails.CachedTokens)
		}
		relayLog.CacheWriteTokens = int(usage.CacheCreationInputTokens)
		relayLog.Cost = m.Stats.InputCost + m.Stats.OutputCost
	}
	relayLog.ClientIP = m.ClientIP

	// 请求内容
	if m.InternalRequest != nil {
		if reqJSON, jsonErr := json.Marshal(m.InternalRequest); jsonErr == nil {
			relayLog.RequestContent = truncateRelayLogJSON(reqJSON)
		}
	}

	// 响应内容
	if m.InternalResponse != nil {
		respForLog := m.filterResponseForLog(m.InternalResponse)
		if respJSON, jsonErr := json.Marshal(respForLog); jsonErr == nil {
			if m.InternalResponse.Usage != nil && m.InternalResponse.Usage.CacheCreationInputTokens > 0 {
				respStr := string(respJSON)
				old := `"usage":{`
				insert := fmt.Sprintf(`"usage":{"cache_creation_input_tokens":%d,`, m.InternalResponse.Usage.CacheCreationInputTokens)
				respJSON = []byte(strings.Replace(respStr, old, insert, 1))
			}
			relayLog.ResponseContent = truncateRelayLogJSON(respJSON)
		}
	}

	// 错误信息
	if err != nil {
		relayLog.Error = err.Error()
	}

	if logErr := op.RelayLogAdd(ctx, relayLog); logErr != nil {
		log.Warnf("failed to save relay log: %v", logErr)
	}
}

func truncateRelayLogJSON(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}

	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return truncateRelayLogString(string(raw))
	}

	truncated := truncateRelayLogValue(payload)
	encoded, err := json.Marshal(truncated)
	if err != nil {
		return truncateRelayLogString(string(raw))
	}
	return string(encoded)
}

func truncateRelayLogValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		truncated := make(map[string]any, len(typed))
		for key, nested := range typed {
			truncated[key] = truncateRelayLogValue(nested)
		}
		return truncated
	case []any:
		truncated := make([]any, 0, len(typed))
		for _, item := range typed {
			truncated = append(truncated, truncateRelayLogValue(item))
		}
		return truncated
	case string:
		return truncateRelayLogString(typed)
	default:
		return value
	}
}

func truncateRelayLogString(value string) string {
	if value == "" {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= relayLogMaxStringRunes {
		return value
	}
	return string(runes[:relayLogMaxStringRunes]) + relayLogTruncatedSuffix
}

// filterResponseForLog 创建响应的浅拷贝，过滤掉 images、MultipleContent 中的图片数据和 Audio.Data 以减少存储压力
func (m *RelayMetrics) filterResponseForLog(resp *transformerModel.InternalLLMResponse) *transformerModel.InternalLLMResponse {
	if resp == nil {
		return nil
	}

	filterMsg := func(msg *transformerModel.Message) *transformerModel.Message {
		if msg == nil {
			return nil
		}
		c := *msg
		c.Images = nil
		if len(c.Content.MultipleContent) > 0 {
			parts := make([]transformerModel.MessageContentPart, 0, len(c.Content.MultipleContent))
			for _, p := range c.Content.MultipleContent {
				if p.Type == "image_url" && p.ImageURL != nil {
					parts = append(parts, transformerModel.MessageContentPart{
						Type:     "image_url",
						ImageURL: &transformerModel.ImageURL{URL: "[image data omitted for storage]"},
					})
				} else {
					parts = append(parts, p)
				}
			}
			c.Content = transformerModel.MessageContent{Content: c.Content.Content, MultipleContent: parts}
		}
		if c.Audio != nil && c.Audio.Data != "" {
			a := *c.Audio
			a.Data = "[audio data omitted for storage]"
			c.Audio = &a
		}
		return &c
	}

	filtered := *resp
	filtered.Choices = make([]transformerModel.Choice, len(resp.Choices))
	for i, choice := range resp.Choices {
		filtered.Choices[i] = choice
		filtered.Choices[i].Message = filterMsg(choice.Message)
		filtered.Choices[i].Delta = filterMsg(choice.Delta)
	}
	return &filtered
}

// triggerCostAlert loads the API key and its accumulated cost stats, then
// invokes op.CheckCostAlert asynchronously. It uses context.Background()
// because the original request context may be canceled after the response
// is sent. Any error is logged inside CheckCostAlert and never propagated.
func triggerCostAlert(apiKeyID int) {
	apiKey, err := op.APIKeyGet(apiKeyID, context.Background())
	if err != nil {
		return
	}
	if apiKey.MaxCost <= 0 {
		return
	}
	stats := op.StatsAPIKeyGet(apiKeyID)
	currentCost := stats.InputCost + stats.OutputCost
	op.CheckCostAlert(uint(apiKeyID), apiKey.Name, currentCost, apiKey.MaxCost)
}
