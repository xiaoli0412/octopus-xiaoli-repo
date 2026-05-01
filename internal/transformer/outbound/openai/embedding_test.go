package openai

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	tmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
)

func TestEmbeddingOutboundTransformResponseReturnsStructuredHTTPError(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"invalid input","type":"invalid_request_error","code":"bad_input"}}`)),
	}

	_, err := (&EmbeddingOutbound{}).TransformResponse(context.Background(), resp)
	if err == nil {
		t.Fatal("TransformResponse() error = nil, want structured response error")
	}
	respErr, ok := err.(*tmodel.ResponseError)
	if !ok {
		t.Fatalf("TransformResponse() error type = %T, want *ResponseError", err)
	}
	if respErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("ResponseError.StatusCode = %d, want %d", respErr.StatusCode, http.StatusBadRequest)
	}
	if respErr.Detail.Message != "invalid input" {
		t.Fatalf("ResponseError.Detail.Message = %q, want %q", respErr.Detail.Message, "invalid input")
	}
}
