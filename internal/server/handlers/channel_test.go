package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/task"
	transformerOutbound "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
)

func createTestChannel(t *testing.T, name string) *model.Channel {
	t.Helper()
	channel := &model.Channel{Name: name, Enabled: true}
	if err := op.ChannelCreate(channel, context.Background()); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	return channel
}

func TestScheduleChannelPostSaveTaskDropsWhenQueueIsFull(t *testing.T) {
	originalSlots := channelPostSaveTaskSlots
	originalTimeout := channelPostSaveTaskTimeout
	originalRunner := channelPostSaveTaskRunner
	channelPostSaveTaskSlots = make(chan struct{}, 1)
	channelPostSaveTaskTimeout = time.Second
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	var releaseOnce sync.Once
	var runs atomic.Int32
	channelPostSaveTaskRunner = func(channel *model.Channel, ctx context.Context) {
		runs.Add(1)
		started <- struct{}{}
		<-release
	}
	t.Cleanup(func() {
		releaseOnce.Do(func() { close(release) })
		channelPostSaveTaskSlots = originalSlots
		channelPostSaveTaskTimeout = originalTimeout
		channelPostSaveTaskRunner = originalRunner
	})

	if ok := scheduleChannelPostSaveTask(&model.Channel{ID: 1}); !ok {
		t.Fatal("first scheduleChannelPostSaveTask() = false, want true")
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first post-save task did not start")
	}

	if ok := scheduleChannelPostSaveTask(&model.Channel{ID: 2}); ok {
		t.Fatal("second scheduleChannelPostSaveTask() = true, want false when queue is full")
	}
	if got := runs.Load(); got != 1 {
		t.Fatalf("runs after rejected schedule = %d, want 1", got)
	}

	releaseOnce.Do(func() { close(release) })
	time.Sleep(20 * time.Millisecond)
	if got := runs.Load(); got != 1 {
		t.Fatalf("runs after release = %d, want 1", got)
	}
}

func TestScheduleChannelPostSaveTaskUsesShutdownAwareContext(t *testing.T) {
	originalSlots := channelPostSaveTaskSlots
	originalTimeout := channelPostSaveTaskTimeout
	originalRunner := channelPostSaveTaskRunner
	channelPostSaveTaskSlots = make(chan struct{}, 1)
	channelPostSaveTaskTimeout = time.Minute
	task.StopAll()
	started := make(chan struct{}, 1)
	ctxDone := make(chan struct{}, 1)
	channelPostSaveTaskRunner = func(channel *model.Channel, ctx context.Context) {
		started <- struct{}{}
		<-ctx.Done()
		ctxDone <- struct{}{}
	}
	t.Cleanup(func() {
		channelPostSaveTaskSlots = originalSlots
		channelPostSaveTaskTimeout = originalTimeout
		channelPostSaveTaskRunner = originalRunner
		_ = task.StopAll()
	})

	if ok := scheduleChannelPostSaveTask(&model.Channel{ID: 3}); !ok {
		t.Fatal("scheduleChannelPostSaveTask() = false, want true")
	}

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("post-save task did not start")
	}

	if err := task.StopAll(); err != nil {
		t.Fatalf("StopAll() error = %v", err)
	}

	select {
	case <-ctxDone:
	case <-time.After(time.Second):
		t.Fatal("post-save task context was not canceled on StopAll()")
	}
}

func TestCreateChannelReturnsBadRequestForInvalidSourceType(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/create", map[string]any{
		"name":      "handler-invalid-source-type",
		"type":      int(transformerOutbound.OutboundTypeOpenAIChat),
		"base_urls": []map[string]any{{"url": "https://example.com/v1", "delay": 0}},
		"keys":      []map[string]any{{"enabled": true, "channel_key": "sk-test", "source_type": "enterprise"}},
		"model":     "gpt-4o",
	}, createChannel)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if res.Message == "" {
		t.Fatalf("expected non-empty error message")
	}
}

func TestCreateChannelReturnsBadRequestForCredentialBearingChannelProxy(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/create", map[string]any{
		"name":          "handler-invalid-channel-proxy",
		"type":          int(transformerOutbound.OutboundTypeOpenAIChat),
		"base_urls":     []map[string]any{{"url": "https://example.com/v1", "delay": 0}},
		"channel_proxy": "https://user:pass@example.com:8443",
		"keys":          []map[string]any{{"enabled": true, "channel_key": "sk-test", "source_type": "public/free"}},
		"model":         "gpt-4o",
	}, createChannel)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if res.Message != "channel_proxy must not include credentials" {
		t.Fatalf("message = %q, want %q", res.Message, "channel_proxy must not include credentials")
	}
}

func TestUpdateChannelReturnsNotFoundForMissingChannel(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/update", map[string]any{
		"id":   999999,
		"name": "missing-channel",
	}, updateChannel)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}

func TestUpdateChannelReturnsBadRequestForCredentialBearingChannelProxy(t *testing.T) {
	setupHandlerTest(t)

	channel := createTestChannel(t, "update-invalid-channel-proxy")
	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/update", map[string]any{
		"id":            channel.ID,
		"channel_proxy": "https://user:pass@example.com:8443",
	}, updateChannel)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if res.Message != "channel_proxy must not include credentials" {
		t.Fatalf("message = %q, want %q", res.Message, "channel_proxy must not include credentials")
	}
}

func TestDeleteChannelRejectsNonPositivePathIDs(t *testing.T) {
	setupHandlerTest(t)

	for _, id := range []string{"0", "-1"} {
		recorder := performParamHandlerRequest(t, http.MethodDelete, "/api/v1/channel/delete/"+id, nil, map[string]string{"id": id}, deleteChannel)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("id %s status = %d, want %d, body = %s", id, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
		res := decodeHandlerResponse(t, recorder)
		if res.Message != "Invalid parameter" {
			t.Fatalf("id %s message = %q, want %q", id, res.Message, "Invalid parameter")
		}
	}
}

func TestTestChannelModelsByConfigReturnsBadRequestForEmptyModels(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/test-models-by-config", map[string]any{
		"type":      int(transformerOutbound.OutboundTypeOpenAIChat),
		"base_urls": []map[string]any{{"url": "https://example.com/v1", "delay": 0}},
		"keys":      []map[string]any{{"enabled": true, "channel_key": "sk-test", "source_type": "public/free"}},
		"models":    []string{},
	}, testChannelModelsByConfig)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if res.Message != "models is required" {
		t.Fatalf("message = %q, want %q", res.Message, "models is required")
	}
}

func TestTestChannelModelsRejectsNonPositiveChannelID(t *testing.T) {
	setupHandlerTest(t)

	for _, channelID := range []int{0, -1} {
		recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/test-models", map[string]any{
			"channel_id": channelID,
			"models":     []string{"gpt-4o"},
		}, testChannelModels)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("channel_id %d status = %d, want %d, body = %s", channelID, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
		res := decodeHandlerResponse(t, recorder)
		if res.Message != resp.ErrInvalidParam {
			t.Fatalf("channel_id %d message = %q, want %q", channelID, res.Message, resp.ErrInvalidParam)
		}
	}
}

func TestFetchModelAcceptsHTTPBaseURL(t *testing.T) {
	setupHandlerTest(t)

	upstream := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-http" {
			t.Fatalf("authorization = %q, want %q", got, "Bearer sk-http")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"http-model"}]}`))
	}))
	defer upstream.Close()

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/fetch-model", map[string]any{
		"type":     int(transformerOutbound.OutboundTypeOpenAIResponse),
		"base_url": upstream.URL,
		"key":      "sk-http",
	}, fetchModel)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	res := decodeHandlerResponse(t, recorder)
	var models []string
	if err := json.Unmarshal(res.Data, &models); err != nil {
		t.Fatalf("json.Unmarshal(response.data) error = %v, body = %s", err, recorder.Body.String())
	}
	if len(models) != 1 || models[0] != "http-model" {
		t.Fatalf("models = %#v, want [http-model]", models)
	}
}

func TestFetchModelRejectsInvalidBaseURL(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/fetch-model", map[string]any{
		"type":     int(transformerOutbound.OutboundTypeOpenAIResponse),
		"base_url": "ftp://example.com/v1",
		"key":      "sk-http",
	}, fetchModel)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if res.Message != "base_url must be absolute http or https URL" {
		t.Fatalf("message = %q, want %q", res.Message, "base_url must be absolute http or https URL")
	}
}

func TestFetchModelRejectsChannelProxyWithCredentials(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/fetch-model", map[string]any{
		"type":          int(transformerOutbound.OutboundTypeOpenAIResponse),
		"base_url":      "https://example.com/v1",
		"key":           "sk-http",
		"proxy":         true,
		"channel_proxy": "https://user:pass@example.com:8443",
	}, fetchModel)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if res.Message != "channel_proxy must not include credentials" {
		t.Fatalf("message = %q, want %q", res.Message, "channel_proxy must not include credentials")
	}
}

func TestFetchModelAcceptsZeroOutboundTypeForOpenAIChat(t *testing.T) {
	setupHandlerTest(t)

	upstream := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-zero" {
			t.Fatalf("authorization = %q, want %q", got, "Bearer sk-zero")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"zero-type-model"}]}`))
	}))
	defer upstream.Close()

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/fetch-model", map[string]any{
		"type":     0,
		"base_url": upstream.URL,
		"key":      "sk-zero",
	}, fetchModel)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	res := decodeHandlerResponse(t, recorder)
	var models []string
	if err := json.Unmarshal(res.Data, &models); err != nil {
		t.Fatalf("json.Unmarshal(response.data) error = %v, body = %s", err, recorder.Body.String())
	}
	if len(models) != 1 || models[0] != "zero-type-model" {
		t.Fatalf("models = %#v, want [zero-type-model]", models)
	}
}

func TestTestChannelModelsByConfigRejectsInvalidBaseURL(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/test-models-by-config", map[string]any{
		"type":      int(transformerOutbound.OutboundTypeOpenAIChat),
		"base_urls": []map[string]any{{"url": "ftp://example.com/v1", "delay": 0}},
		"keys":      []map[string]any{{"enabled": true, "channel_key": "sk-test", "source_type": "public/free"}},
		"models":    []string{"gpt-4o"},
	}, testChannelModelsByConfig)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if res.Message != "base_url must be absolute http or https URL" {
		t.Fatalf("message = %q, want %q", res.Message, "base_url must be absolute http or https URL")
	}
}

func TestTestChannelModelsByConfigRejectsChannelProxyWithCredentials(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/test-models-by-config", map[string]any{
		"type":          int(transformerOutbound.OutboundTypeOpenAIChat),
		"base_urls":     []map[string]any{{"url": "https://example.com/v1", "delay": 0}},
		"keys":          []map[string]any{{"enabled": true, "channel_key": "sk-test", "source_type": "public/free"}},
		"models":        []string{"gpt-4o"},
		"proxy":         true,
		"channel_proxy": "https://user:pass@example.com:8443",
	}, testChannelModelsByConfig)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if res.Message != "channel_proxy must not include credentials" {
		t.Fatalf("message = %q, want %q", res.Message, "channel_proxy must not include credentials")
	}
}
