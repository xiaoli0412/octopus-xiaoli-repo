package openai

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	tmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
)

func TestChatOutboundTransformResponseRejectsOversizedBody(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(strings.Repeat("a", int(maxOpenAIResponseBodyBytes)+1))),
	}

	_, err := (&ChatOutbound{}).TransformResponse(context.Background(), resp)
	if err == nil || !strings.Contains(err.Error(), "openai response body too large") {
		t.Fatalf("TransformResponse() error = %v, want openai response body too large", err)
	}
}

func TestChatOutboundTransformResponseReturnsStructuredHTTPError(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusUnauthorized,
		Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"bad key","type":"invalid_request_error","code":"invalid_api_key"}}`)),
	}

	_, err := (&ChatOutbound{}).TransformResponse(context.Background(), resp)
	if err == nil {
		t.Fatal("TransformResponse() error = nil, want structured response error")
	}
	respErr, ok := err.(*tmodel.ResponseError)
	if !ok {
		t.Fatalf("TransformResponse() error type = %T, want *ResponseError", err)
	}
	if respErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("ResponseError.StatusCode = %d, want %d", respErr.StatusCode, http.StatusUnauthorized)
	}
	if respErr.Detail.Message != "bad key" {
		t.Fatalf("ResponseError.Detail.Message = %q, want %q", respErr.Detail.Message, "bad key")
	}
}

func TestChatOutboundTransformResponseParsesCacheUsageDetails(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(strings.NewReader(`{
			"id":"chatcmpl-cache",
			"object":"chat.completion",
			"created":1,
			"model":"gpt-4.1",
			"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],
			"usage":{
				"prompt_tokens":120,
				"completion_tokens":45,
				"total_tokens":181,
				"prompt_tokens_details":{"cached_tokens":30},
				"cache_creation_input_tokens":16
			}
		}`)),
	}

	got, err := (&ChatOutbound{}).TransformResponse(context.Background(), resp)
	if err != nil {
		t.Fatalf("TransformResponse() error = %v", err)
	}
	if got == nil || got.Usage == nil {
		t.Fatalf("usage = %#v, want parsed usage", got)
	}
	if got.Usage.PromptTokensDetails == nil || got.Usage.PromptTokensDetails.CachedTokens != 30 {
		t.Fatalf("cached tokens = %#v, want 30", got.Usage.PromptTokensDetails)
	}
	if got.Usage.CacheCreationInputTokens != 16 {
		t.Fatalf("cache creation input tokens = %d, want 16", got.Usage.CacheCreationInputTokens)
	}
}
