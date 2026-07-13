package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/relay"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/inbound"
)

// replayHeader 标记重放请求，relay 引擎据此跳过统计写入。
const replayHeader = "X-Octopus-Replay"

func init() {
	router.NewGroupRouter("/api/v1/log").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/replay/:id", http.MethodPost).
				Handle(replayLog),
		)
}

// replayOriginalSummary 原始请求响应摘要
type replayOriginalSummary struct {
	Model         string `json:"model"`
	Channel       string `json:"channel"`
	Content       string `json:"content"`
	Error         string `json:"error,omitempty"`
	InputTokens   int    `json:"input_tokens"`
	OutputTokens  int    `json:"output_tokens"`
	UseTime       int    `json:"use_time"`
}

// replayReplayResult 重放结果
type replayReplayResult struct {
	Status  int    `json:"status"`
	Content string `json:"content"`
}

// replayDiff 差异摘要
type replayDiff struct {
	Changed bool `json:"changed"`
}

// replayResponse 完整重放响应
type replayResponse struct {
	Original replayOriginalSummary `json:"original"`
	Replay   replayReplayResult    `json:"replay"`
	Diff     replayDiff            `json:"diff"`
}

func replayLog(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	if idStr == "" {
		resp.Error(c, http.StatusBadRequest, "invalid log id")
		return
	}

	// RelayLog.ID 为 Snowflake int64
	var id int64
	if err := json.Unmarshal([]byte(idStr), &id); err != nil {
		resp.Error(c, http.StatusBadRequest, "invalid log id")
		return
	}
	if id <= 0 {
		resp.Error(c, http.StatusBadRequest, "invalid log id")
		return
	}

	logEntry, err := op.RelayLogGetByID(c.Request.Context(), id)
	if err != nil {
		resp.Error(c, http.StatusNotFound, "log not found")
		return
	}

	if strings.TrimSpace(logEntry.RequestContent) == "" {
		resp.Error(c, http.StatusBadRequest, "original request content is empty")
		return
	}

	// 获取原始请求使用的 API Key
	if logEntry.APIKeyID == 0 {
		resp.Error(c, http.StatusBadRequest, "original request has no API key")
		return
	}
	apiKey, err := op.APIKeyGet(logEntry.APIKeyID, c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusBadRequest, "original API key not found")
		return
	}

	// 推断原始入站协议类型
	inboundType := detectInboundType(logEntry.RequestContent)
	relayPath := relayPathForInbound(inboundType)

	// 构造重放请求上下文
	recorder := httptest.NewRecorder()
	relayCtx, _ := gin.CreateTestContext(recorder)
	relayCtx.Request = httptest.NewRequest(http.MethodPost, relayPath, strings.NewReader(logEntry.RequestContent))
	relayCtx.Request.Header.Set("Content-Type", "application/json")
	relayCtx.Request.Header.Set("Authorization", "Bearer "+apiKey.APIKey)
	relayCtx.Request.Header.Set(replayHeader, "true")

	// 复制 APIKeyAuth 中间件设置的 context 值
	relayCtx.Set("request_type", "openai")
	relayCtx.Set("supported_models", apiKey.SupportedModels)
	relayCtx.Set("api_key_id", apiKey.ID)
	relayCtx.Set("api_key", apiKey)
	relayCtx.Set("api_key_allowed_channels", apiKey.APIKeyAllowedChannels())
	relayCtx.Set("api_key_allowed_groups", apiKey.APIKeyAllowedGroups())

	// 调用 relay 引擎
	relay.Handler(inboundType, relayCtx)

	replayBody := recorder.Body.String()
	replayStatus := recorder.Code

	original := replayOriginalSummary{
		Model:        logEntry.ActualModelName,
		Channel:      logEntry.ChannelName,
		Content:      logEntry.ResponseContent,
		Error:        logEntry.Error,
		InputTokens:  logEntry.InputTokens,
		OutputTokens: logEntry.OutputTokens,
		UseTime:      logEntry.UseTime,
	}

	replay := replayReplayResult{
		Status:  replayStatus,
		Content: replayBody,
	}

	diff := replayDiff{
		Changed: strings.TrimSpace(logEntry.ResponseContent) != strings.TrimSpace(replayBody),
	}

	resp.Success(c, replayResponse{
		Original: original,
		Replay:   replay,
		Diff:     diff,
	})
}

// detectInboundType 根据请求内容推断入站协议类型
func detectInboundType(requestContent string) inbound.InboundType {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(requestContent), &raw); err != nil {
		return inbound.InboundTypeOpenAIChat
	}
	if _, ok := raw["messages"]; ok {
		return inbound.InboundTypeOpenAIChat
	}
	if _, ok := raw["input"]; ok {
		return inbound.InboundTypeOpenAIResponse
	}
	if _, ok := raw["contents"]; ok {
		return inbound.InboundTypeGemini
	}
	return inbound.InboundTypeOpenAIChat
}

// relayPathForInbound 根据入站类型返回对应的 relay 路径
func relayPathForInbound(t inbound.InboundType) string {
	switch t {
	case inbound.InboundTypeOpenAIChat:
		return "/v1/chat/completions"
	case inbound.InboundTypeOpenAIResponse:
		return "/v1/responses"
	case inbound.InboundTypeAnthropic:
		return "/v1/messages"
	case inbound.InboundTypeOpenAIEmbedding:
		return "/v1/embeddings"
	default:
		return "/v1/chat/completions"
	}
}
