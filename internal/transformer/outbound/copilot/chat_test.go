package copilot

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

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestCopilotTokenHTTPClientHasTimeout(t *testing.T) {
	if copilotTokenHTTPClient == nil {
		t.Fatal("copilotTokenHTTPClient should be initialized")
	}
	if copilotTokenHTTPClient == http.DefaultClient {
		t.Fatal("copilotTokenHTTPClient should not reuse http.DefaultClient")
	}
	if copilotTokenHTTPClient.Timeout != 15*time.Second {
		t.Fatalf("copilotTokenHTTPClient.Timeout = %v, want %v", copilotTokenHTTPClient.Timeout, 15*time.Second)
	}
}

func TestExchangeTokenRejectsOversizedResponse(t *testing.T) {
	oldClient := copilotTokenHTTPClient
	defer func() {
		copilotTokenHTTPClient = oldClient
	}()

	copilotTokenHTTPClient = &http.Client{
		Timeout: 15 * time.Second,
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(strings.Repeat("x", int(maxCopilotTokenResponseBytes)+1))),
				Request:    r,
			}, nil
		}),
	}

	tokenCacheMu.Lock()
	delete(tokenCache, "github-token-oversized")
	tokenCacheMu.Unlock()

	_, err := ExchangeToken(context.Background(), "github-token-oversized")
	if err == nil || !strings.Contains(err.Error(), "copilot token response too large") {
		t.Fatalf("ExchangeToken() error = %v, want copilot token response too large", err)
	}
}

func TestChatOutboundTransformResponseReturnsStructuredHTTPError(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusForbidden,
		Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"copilot denied","type":"permission_error","code":"forbidden"}}`)),
	}

	_, err := (&ChatOutbound{}).TransformResponse(context.Background(), resp)
	if err == nil {
		t.Fatal("TransformResponse() error = nil, want structured response error")
	}
	respErr, ok := err.(*tmodel.ResponseError)
	if !ok {
		t.Fatalf("TransformResponse() error type = %T, want *ResponseError", err)
	}
	if respErr.StatusCode != http.StatusForbidden {
		t.Fatalf("ResponseError.StatusCode = %d, want %d", respErr.StatusCode, http.StatusForbidden)
	}
	if respErr.Detail.Message != "copilot denied" {
		t.Fatalf("ResponseError.Detail.Message = %q, want %q", respErr.Detail.Message, "copilot denied")
	}
}
