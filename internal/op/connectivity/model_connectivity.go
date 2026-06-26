package connectivity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/helper"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

const modelConnectivityTestTimeout = 60 * time.Second
const maxTestResponseBodyBytes = 256 * 1024

// TestModelConnectivity sends a minimal chat completion request to the given
// channel to verify the model is reachable and checks whether the local price
// matches the latest upstream price for that channel.
// The request is not recorded in relay statistics.
func TestModelConnectivity(ctx context.Context, req model.ModelTestRequest) (model.ModelTestResult, error) {
	channel, err := op.ChannelGet(req.ChannelID, ctx)
	if err != nil {
		return model.ModelTestResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("channel not found: %v", err),
		}, nil
	}

	if !channel.Enabled {
		return model.ModelTestResult{
			Success:      false,
			ErrorMessage: "channel is disabled",
		}, nil
	}

	modelName := strings.ToLower(strings.TrimSpace(req.Model))
	if modelName == "" {
		return model.ModelTestResult{
			Success:      false,
			ErrorMessage: "model is required",
		}, nil
	}

	if !channel.SupportsModel(modelName) {
		return model.ModelTestResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("channel does not support model %s", modelName),
		}, nil
	}

	key := channel.GetChannelKeyForModel(modelName)
	if strings.TrimSpace(key.ChannelKey) == "" {
		return model.ModelTestResult{
			Success:      false,
			ErrorMessage: "no available key for model",
		}, nil
	}

	baseURL := channel.GetBaseUrl()
	if strings.TrimSpace(baseURL) == "" {
		return model.ModelTestResult{
			Success:      false,
			ErrorMessage: "channel has no base url",
		}, nil
	}

	message := strings.TrimSpace(req.Message)
	if message == "" {
		message = "hi"
	}

	testReqBody := map[string]any{
		"model": modelName,
		"messages": []map[string]string{
			{"role": "user", "content": message},
		},
	}
	bodyBytes, err := json.Marshal(testReqBody)
	if err != nil {
		return model.ModelTestResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to marshal request: %v", err),
		}, nil
	}

	testURL := strings.TrimSuffix(baseURL, "/") + "/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return model.ModelTestResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key.ChannelKey)
	for _, h := range channel.CustomHeader {
		if strings.TrimSpace(h.HeaderKey) != "" {
			httpReq.Header.Set(h.HeaderKey, h.HeaderValue)
		}
	}

	httpClient, err := helper.ChannelHttpClientWithTimeout(channel, modelConnectivityTestTimeout)
	if err != nil {
		return model.ModelTestResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get http client: %v", err),
		}, nil
	}

	start := time.Now()
	resp, err := httpClient.Do(httpReq)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		return model.ModelTestResult{
			Success:      false,
			LatencyMs:    latencyMs,
			ErrorMessage: fmt.Sprintf("failed to send request: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxTestResponseBodyBytes+1))
	if err != nil {
		return model.ModelTestResult{
			Success:      false,
			LatencyMs:    latencyMs,
			ErrorMessage: fmt.Sprintf("failed to read response body: %v", err),
		}, nil
	}
	if int64(len(respBody)) > maxTestResponseBodyBytes {
		return model.ModelTestResult{
			Success:      false,
			LatencyMs:    latencyMs,
			ErrorMessage: "response body too large",
		}, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.ModelTestResult{
			Success:      false,
			LatencyMs:    latencyMs,
			ErrorMessage: fmt.Sprintf("upstream error %d: %s", resp.StatusCode, string(respBody)),
		}, nil
	}

	var parsed struct {
		Choices []struct {
			Message *struct {
				Content string `json:"content"`
			} `json:"message,omitempty"`
		} `json:"choices,omitempty"`
		Model string `json:"model"`
	}
	responseText := ""
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		log.Warnf("model connectivity test failed to parse response: %v", err)
		responseText = string(respBody)
	} else {
		if len(parsed.Choices) > 0 && parsed.Choices[0].Message != nil {
			responseText = parsed.Choices[0].Message.Content
		}
		if responseText == "" {
			responseText = string(respBody)
		}
	}

	priceMatch := checkModelPriceMatch(modelName, channel.ID)

	return model.ModelTestResult{
		Success:      true,
		LatencyMs:    latencyMs,
		ResponseText: responseText,
		PriceMatch:   priceMatch,
	}, nil
}

func checkModelPriceMatch(modelName string, channelID int) bool {
	localInfo, err := op.LLMGet(modelName)
	if err != nil {
		return false
	}

	upstreamPrice, ok := op.ResolveGatewayLLMPrice(modelName, channelID)
	if !ok {
		return false
	}

	return priceWithinTolerance(localInfo.Input, upstreamPrice.Input) &&
		priceWithinTolerance(localInfo.Output, upstreamPrice.Output) &&
		priceWithinTolerance(localInfo.CacheRead, upstreamPrice.CacheRead) &&
		priceWithinTolerance(localInfo.CacheWrite, upstreamPrice.CacheWrite)
}

func priceWithinTolerance(a, b float64) bool {
	const epsilon = 1e-9
	if a == 0 && b == 0 {
		return true
	}
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	if diff <= epsilon {
		return true
	}
	max := a
	if b > max {
		max = b
	}
	if max <= epsilon {
		return false
	}
	return diff/max <= 0.05
}
