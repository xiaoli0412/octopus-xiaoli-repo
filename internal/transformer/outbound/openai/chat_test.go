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
