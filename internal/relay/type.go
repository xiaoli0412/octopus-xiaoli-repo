package relay

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	dbmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/relay/balancer"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
)

// maxSSEEventSize 定义 SSE 事件的最大大小。
// 对于图像生成模型（如 gemini-3-pro-image-preview），返回的 base64 编码图像数据
// 可能非常大（高分辨率图像可能超过 10MB），因此需要设置足够大的缓冲区。
// 默认 32MB，可通过环境变量 OCTOPUS_RELAY_MAX_SSE_EVENT_SIZE 覆盖。
const (
	maxUpstreamErrorBodyBytes int64 = 8 * 1024
	maxRelayRequestBodyBytes  int64 = 8 * 1024 * 1024
)

var maxSSEEventSize = 32 * 1024 * 1024

func init() {
	if raw := strings.TrimSpace(os.Getenv(strings.ToUpper(conf.APP_NAME) + "_RELAY_MAX_SSE_EVENT_SIZE")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			maxSSEEventSize = v
		}
	}
}

// hopByHopHeaders 定义不应转发的 HTTP 头。
var hopByHopHeaders = map[string]bool{
	"authorization":       true,
	"x-api-key":           true,
	"connection":          true,
	"keep-alive":          true,
	"proxy-authenticate":  true,
	"proxy-authorization": true,
	"te":                  true,
	"trailer":             true,
	"transfer-encoding":   true,
	"upgrade":             true,
	"content-length":      true,
	"host":                true,
	"accept-encoding":     true,
	"x-forwarded-for":     true,
	"x-forwarded-host":    true,
	"x-forwarded-proto":   true,
	"x-forwarded-port":    true,
	"x-real-ip":           true,
	"forwarded":           true,
	"cf-connecting-ip":    true,
	"true-client-ip":      true,
	"x-client-ip":         true,
	"x-cluster-client-ip": true,
}

type relayRequest struct {
	c               *gin.Context
	inAdapter       model.Inbound
	internalRequest *model.InternalLLMRequest
	isStreaming     bool
	isEmbedding     bool
	isChat          bool
	metrics         *RelayMetrics
	apiKeyID        int
	requestModel    string
	iter            *balancer.Iterator
	dynamicMode     *dynamicRoutingModeState
}

func requestCapabilityForAPIFormat(format model.APIFormat) string {
	switch format {
	case model.APIFormatOpenAIChatCompletion, model.APIFormatAiSDKText, model.APIFormatAiSDKDataStream:
		return dbmodel.RequestCapabilityOpenAIChat
	case model.APIFormatOpenAIResponse, model.APIFormatOpenAIImageGeneration:
		return dbmodel.RequestCapabilityOpenAIResponses
	case model.APIFormatOpenAIEmbedding:
		return dbmodel.RequestCapabilityOpenAIEmbeddings
	case model.APIFormatAnthropicMessage:
		return dbmodel.RequestCapabilityAnthropicMessages
	case model.APIFormatGeminiContents:
		return dbmodel.RequestCapabilityGeminiContents
	default:
		return ""
	}
}

func requestCapabilityForInternalRequest(req *model.InternalLLMRequest) string {
	if req == nil {
		return ""
	}
	return requestCapabilityForAPIFormat(req.RawAPIFormat)
}

func resolveRequestCapability(channel *dbmodel.Channel, targetModel string, rawRequestCapability string) string {
	if rawRequestCapability != "" {
		return rawRequestCapability
	}
	if channel == nil {
		return ""
	}
	return channel.RequestCapabilityForModel(targetModel)
}

func (req *relayRequest) requestCapabilityFor(channel *dbmodel.Channel, targetModel string) string {
	if req == nil {
		return resolveRequestCapability(channel, targetModel, "")
	}
	return resolveRequestCapability(channel, targetModel, requestCapabilityForInternalRequest(req.internalRequest))
}

// relayAttempt 尝试级上下文
type relayAttempt struct {
	*relayRequest

	outAdapter           model.Outbound
	channel              *dbmodel.Channel
	usedKey              dbmodel.ChannelKey
	firstTokenTimeOutSec int
}

// attemptResult 封装单次尝试的结果
type attemptResult struct {
	Success          bool
	Written          bool
	Err              error
	AttemptRecorded  bool
	StatusCode       int
	Duration         time.Duration
	Channel          *dbmodel.Channel
	UsedKey          dbmodel.ChannelKey
	InternalResponse *model.InternalLLMResponse
	ActualModel      string
}

type raceOutcome struct {
	index  int
	result attemptResult
}

type raceAttemptRecorder interface {
	Record(channelID, channelKeyID int, channelName, modelName string, status dbmodel.AttemptStatus, statusCode int, duration time.Duration, msg string)
}
