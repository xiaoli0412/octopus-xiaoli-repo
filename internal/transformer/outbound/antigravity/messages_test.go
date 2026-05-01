package antigravity

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	tmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestLoadCodeAssistHTTPClientHasTimeout(t *testing.T) {
	if loadCodeAssistHTTPClient == nil {
		t.Fatal("loadCodeAssistHTTPClient should be initialized")
	}
	if loadCodeAssistHTTPClient == http.DefaultClient {
		t.Fatal("loadCodeAssistHTTPClient should not reuse http.DefaultClient")
	}
	if loadCodeAssistHTTPClient.Timeout != 15*time.Second {
		t.Fatalf("loadCodeAssistHTTPClient.Timeout = %v, want %v", loadCodeAssistHTTPClient.Timeout, 15*time.Second)
	}
}

func TestGetProjectIDRejectsOversizedLoadCodeAssistResponse(t *testing.T) {
	oldClient := loadCodeAssistHTTPClient
	loadCodeAssistHTTPClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(strings.Repeat("x", int(maxLoadCodeAssistResponseBytes)+1))),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})}
	defer func() {
		loadCodeAssistHTTPClient = oldClient
	}()

	_, err := getProjectID(context.Background(), "http://example.invalid", "token-oversized-response")
	if err == nil || !strings.Contains(err.Error(), "loadCodeAssist response too large") {
		t.Fatalf("getProjectID() error = %v, want loadCodeAssist response too large", err)
	}
}

func TestMessagesOutboundTransformResponseReturnsStructuredHTTPError(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader(`{"error":{"code":400,"message":"invalid project","status":"INVALID_ARGUMENT"}}`)),
	}

	_, err := (&MessagesOutbound{}).TransformResponse(context.Background(), resp)
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
	if respErr.Detail.Message != "invalid project" {
		t.Fatalf("ResponseError.Detail.Message = %q, want %q", respErr.Detail.Message, "invalid project")
	}
	if respErr.Detail.Type != "INVALID_ARGUMENT" {
		t.Fatalf("ResponseError.Detail.Type = %q, want %q", respErr.Detail.Type, "INVALID_ARGUMENT")
	}
}
