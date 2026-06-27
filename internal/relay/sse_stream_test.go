package relay

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	tmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
)

func TestUseGoSSE(t *testing.T) {
	cases := []struct {
		name string
		env  string
		want bool
	}{
		{name: "unset", env: "", want: false},
		{name: "1", env: "1", want: true},
		{name: "true lower", env: "true", want: true},
		{name: "true upper", env: "TRUE", want: true},
		{name: "true mixed", env: "True", want: true},
		{name: "0", env: "0", want: false},
		{name: "false", env: "false", want: false},
		{name: "whitespace", env: "  TRUE  ", want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("OCTOPUS_USE_GO_SSE", tc.env)
			if got := useGoSSE(); got != tc.want {
				t.Fatalf("useGoSSE() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestExtractSSEData(t *testing.T) {
	cases := []struct {
		name  string
		event string
		want  string
	}{
		{
			name:  "single data line",
			event: "data: {\"x\":1}",
			want:  "{\"x\":1}",
		},
		{
			name:  "multiple data lines joined",
			event: "data: line1\ndata: line2",
			want:  "line1\nline2",
		},
		{
			name:  "ignores other fields",
			event: "event: msg\nid: 1\nretry: 100\ndata: {\"x\":1}",
			want:  "{\"x\":1}",
		},
		{
			name:  "empty data line skipped",
			event: "data: a\ndata:\ndata: b",
			want:  "a\nb",
		},
		{
			name:  "done marker passed through",
			event: "data: [DONE]",
			want:  "[DONE]",
		},
		{
			name:  "crlf line endings",
			event: "event: msg\r\ndata: {\"x\":1}\r\n",
			want:  "{\"x\":1}",
		},
		{
			name:  "no data lines",
			event: "event: ping\nid: 1",
			want:  "",
		},
		{
			name:  "data field case insensitive",
			event: "DATA: {\"x\":1}",
			want:  "{\"x\":1}",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractSSEData(tc.event); got != tc.want {
				t.Fatalf("extractSSEData(%q) = %q, want %q", tc.event, got, tc.want)
			}
		})
	}
}

type streamCaptureOutbound struct{ data []string }

func (f *streamCaptureOutbound) TransformRequest(ctx context.Context, request *tmodel.InternalLLMRequest, baseURL, key string) (*http.Request, error) {
	return nil, nil
}

func (f *streamCaptureOutbound) TransformResponse(ctx context.Context, response *http.Response) (*tmodel.InternalLLMResponse, error) {
	return &tmodel.InternalLLMResponse{}, nil
}

func (f *streamCaptureOutbound) TransformStream(ctx context.Context, eventData []byte) (*tmodel.InternalLLMResponse, error) {
	f.data = append(f.data, string(eventData))
	return &tmodel.InternalLLMResponse{Object: "chat.completion.chunk"}, nil
}

type echoStreamInbound struct{ prefix string }

func (f *echoStreamInbound) TransformRequest(ctx context.Context, body []byte) (*tmodel.InternalLLMRequest, error) {
	return &tmodel.InternalLLMRequest{}, nil
}

func (f *echoStreamInbound) TransformResponse(ctx context.Context, response *tmodel.InternalLLMResponse) ([]byte, error) {
	return nil, nil
}

func (f *echoStreamInbound) TransformStream(ctx context.Context, stream *tmodel.InternalLLMResponse) ([]byte, error) {
	return []byte(f.prefix + stream.Object), nil
}

func (f *echoStreamInbound) GetInternalResponse(ctx context.Context) (*tmodel.InternalLLMResponse, error) {
	return nil, nil
}

func testHandleStreamResponse(t *testing.T, useGo string) {
	if useGo != "" {
		t.Setenv("OCTOPUS_USE_GO_SSE", useGo)
	} else {
		t.Setenv("OCTOPUS_USE_GO_SSE", "")
	}

	ctx, recorder := newRelayAttemptTestContext("POST", "http://example.com/v1/chat/completions", nil)
	body := "event: msg\n" +
		"id: 1\n" +
		"data: {\"chunk\":1}\n" +
		"\n" +
		"data: line1\n" +
		"data: line2\n" +
		"\n" +
		"data: [DONE]\n" +
		"\n"
	response := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	out := &streamCaptureOutbound{}
	ra := &relayAttempt{
		relayRequest: &relayRequest{
			c:               ctx,
			inAdapter:       &echoStreamInbound{prefix: "out: "},
			internalRequest: &tmodel.InternalLLMRequest{Model: "gpt-4o"},
			metrics:         NewRelayMetrics(0, "gpt-4o", &tmodel.InternalLLMRequest{Model: "gpt-4o"}),
		},
		outAdapter: out,
	}

	if err := ra.handleStreamResponse(context.Background(), response); err != nil {
		t.Fatalf("handleStreamResponse() error = %v", err)
	}

	want := []string{`{"chunk":1}`, "line1\nline2", "[DONE]"}
	if len(out.data) != len(want) {
		t.Fatalf("captured data = %#v, want %#v", out.data, want)
	}
	for i := range want {
		if out.data[i] != want[i] {
			t.Fatalf("data[%d] = %q, want %q", i, out.data[i], want[i])
		}
	}
	gotBody := recorder.Body.String()
	if !strings.Contains(gotBody, "out: chat.completion.chunk") {
		t.Fatalf("response body = %q, want streamed transformed output", gotBody)
	}
	if ct := recorder.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content-type = %q, want text/event-stream", ct)
	}
}

func TestHandleStreamResponseStreamBuffer(t *testing.T) {
	testHandleStreamResponse(t, "")
}

func TestHandleStreamResponseGoSSE(t *testing.T) {
	testHandleStreamResponse(t, "1")
}
