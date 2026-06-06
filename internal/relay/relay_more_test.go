package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	dbmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/relay/balancer"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/inbound"
	tmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
)

func setupRelayTestDB(t *testing.T) context.Context {
	t.Helper()

	resetRelayTestState()
	if db.GetDB() != nil {
		_ = db.Close()
	}

	dbPath := filepath.Join(t.TempDir(), "octopus-relay-test.db")
	if err := db.InitDB("sqlite", dbPath, false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	if err := op.InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := op.RelayLogClear(context.Background()); err != nil {
		t.Fatalf("RelayLogClear() error = %v", err)
	}

	t.Cleanup(func() {
		if db.GetDB() != nil {
			_ = db.Close()
		}
		resetRelayTestState()
	})

	return context.Background()
}

func resetRelayTestState() {
	op.LLMBatchDelete([]string{}, context.Background())
	resetRaceBudgetState()
	balancer.ResetCircuitStateForTest()
}

type fakeInbound struct {
	request         *tmodel.InternalLLMRequest
	requestErr      error
	transformedBody []byte
	responseErr     error
	streamBody      []byte
	streamErr       error
	storedResponse  *tmodel.InternalLLMResponse
	storedErr       error
}

func (f *fakeInbound) TransformRequest(ctx context.Context, body []byte) (*tmodel.InternalLLMRequest, error) {
	if f.requestErr != nil {
		return nil, f.requestErr
	}
	if f.request != nil {
		return f.request, nil
	}
	var req tmodel.InternalLLMRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (f *fakeInbound) TransformResponse(ctx context.Context, response *tmodel.InternalLLMResponse) ([]byte, error) {
	if f.responseErr != nil {
		return nil, f.responseErr
	}
	if f.transformedBody != nil {
		return f.transformedBody, nil
	}
	return json.Marshal(response)
}

func (f *fakeInbound) TransformStream(ctx context.Context, stream *tmodel.InternalLLMResponse) ([]byte, error) {
	if f.streamErr != nil {
		return nil, f.streamErr
	}
	if f.streamBody != nil {
		return f.streamBody, nil
	}
	return []byte("data: stub\n\n"), nil
}

func (f *fakeInbound) GetInternalResponse(ctx context.Context) (*tmodel.InternalLLMResponse, error) {
	if f.storedErr != nil {
		return nil, f.storedErr
	}
	return f.storedResponse, nil
}

type fakeOutbound struct {
	request      *http.Request
	requestErr   error
	response     *tmodel.InternalLLMResponse
	responseErr  error
	streamResp   *tmodel.InternalLLMResponse
	streamErr    error
	capturedURL  string
	capturedKey  string
	capturedBody *tmodel.InternalLLMRequest
}

func (f *fakeOutbound) TransformRequest(ctx context.Context, request *tmodel.InternalLLMRequest, baseURL, key string) (*http.Request, error) {
	if f.requestErr != nil {
		return nil, f.requestErr
	}
	f.capturedURL = baseURL
	f.capturedKey = key
	f.capturedBody = request
	if f.request != nil {
		return f.request, nil
	}
	return http.NewRequestWithContext(ctx, http.MethodPost, baseURL, nil)
}

func (f *fakeOutbound) TransformResponse(ctx context.Context, response *http.Response) (*tmodel.InternalLLMResponse, error) {
	if f.responseErr != nil {
		return nil, f.responseErr
	}
	if f.response != nil {
		return f.response, nil
	}
	return &tmodel.InternalLLMResponse{}, nil
}

func (f *fakeOutbound) TransformStream(ctx context.Context, eventData []byte) (*tmodel.InternalLLMResponse, error) {
	if f.streamErr != nil {
		return nil, f.streamErr
	}
	if f.streamResp != nil {
		return f.streamResp, nil
	}
	return &tmodel.InternalLLMResponse{}, nil
}

func newRelayAttemptTestContext(method, target string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, target, bytes.NewReader(body))
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = request
	return ctx, recorder
}

func strPtr(s string) *string {
	return &s
}

func TestRelayRequestSelectsKeyByRawAPIFormat(t *testing.T) {
	channel := &dbmodel.Channel{
		ID:               77,
		Name:             "raw-format-key-channel",
		Type:             outbound.OutboundTypeOpenAIChat,
		Enabled:          true,
		KeyRoutingPolicy: dbmodel.KeyRoutingPolicyPriority,
		Model:            "gpt-4o",
		Keys: []dbmodel.ChannelKey{
			{ID: 1, Enabled: true, ChannelKey: "chat-key", AllowedModels: "gpt-4o", RequestCapabilities: dbmodel.RequestCapabilityOpenAIChat},
			{ID: 2, Enabled: true, ChannelKey: "responses-key", AllowedModels: "gpt-4o", RequestCapabilities: dbmodel.RequestCapabilityOpenAIResponses},
		},
	}

	chatReq := &relayRequest{internalRequest: &tmodel.InternalLLMRequest{Model: "gpt-4o", RawAPIFormat: tmodel.APIFormatOpenAIChatCompletion}}
	chatCapability := chatReq.requestCapabilityFor(channel, "gpt-4o")
	if chatCapability != dbmodel.RequestCapabilityOpenAIChat {
		t.Fatalf("chat capability = %q, want %q", chatCapability, dbmodel.RequestCapabilityOpenAIChat)
	}
	if key := channel.GetChannelKeyForRequest("gpt-4o", chatCapability); key.ChannelKey != "chat-key" {
		t.Fatalf("chat key = %#v, want chat-key", key)
	}

	responsesReq := &relayRequest{internalRequest: &tmodel.InternalLLMRequest{Model: "gpt-4o", RawAPIFormat: tmodel.APIFormatOpenAIResponse}}
	responsesCapability := responsesReq.requestCapabilityFor(channel, "gpt-4o")
	if responsesCapability != dbmodel.RequestCapabilityOpenAIResponses {
		t.Fatalf("responses capability = %q, want %q", responsesCapability, dbmodel.RequestCapabilityOpenAIResponses)
	}
	if key := channel.GetChannelKeyForRequest("gpt-4o", responsesCapability); key.ChannelKey != "responses-key" {
		t.Fatalf("responses key = %#v, want responses-key", key)
	}
}

func setupHandlerTestRoute(t *testing.T, ctx context.Context, modelName, baseURL string, retryRounds int, retryDelayMs int, keys []string) (*dbmodel.Channel, map[string]dbmodel.ChannelKey) {
	t.Helper()

	channel := &dbmodel.Channel{
		Name:              strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-")) + "-channel",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: baseURL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	addRequests := make([]dbmodel.ChannelKeyAddRequest, 0, len(keys))
	for _, key := range keys {
		addRequests = append(addRequests, dbmodel.ChannelKeyAddRequest{
			Enabled:       true,
			ChannelKey:    key,
			SourceType:    "public/free",
			AllowedModels: modelName,
		})
	}
	updated, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: channel.ID, KeysToAdd: addRequests}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(keys) error = %v", err)
	}

	group := &dbmodel.Group{
		Name:         modelName,
		Mode:         dbmodel.GroupModeFailover,
		RetryRounds:  retryRounds,
		RetryDelayMs: retryDelayMs,
	}
	if err := op.GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: channel.ID, ModelName: modelName, Priority: 1, Weight: 1}, ctx); err != nil {
		t.Fatalf("GroupItemAdd() error = %v", err)
	}

	keysByValue := make(map[string]dbmodel.ChannelKey, len(updated.Keys))
	for _, key := range updated.Keys {
		keysByValue[key.ChannelKey] = key
	}
	return updated, keysByValue
}

func insertRelayGroupItemDirect(t *testing.T, ctx context.Context, item *dbmodel.GroupItem) {
	t.Helper()
	if err := db.GetDB().WithContext(ctx).Create(item).Error; err != nil {
		t.Fatalf("direct Create(group item) error = %v", err)
	}
	if _, err := op.GroupUpdate(&dbmodel.GroupUpdateRequest{ID: item.GroupID}, ctx); err != nil {
		t.Fatalf("GroupUpdate(refresh) error = %v", err)
	}
}

func relayLogsForTest(t *testing.T) []dbmodel.RelayLog {
	t.Helper()
	logs, err := op.RelayLogList(context.Background(), nil, nil, 1, 20)
	if err != nil {
		t.Fatalf("RelayLogList() error = %v", err)
	}
	return logs
}

func TestAllowsRacingByModel(t *testing.T) {
	ctx := setupRelayTestDB(t)

	mustCreateModel := func(info dbmodel.LLMInfo) {
		t.Helper()
		if err := op.LLMCreate(info, ctx); err != nil {
			t.Fatalf("LLMCreate(%q) error = %v", info.Name, err)
		}
	}

	mustCreateModel(dbmodel.LLMInfo{
		Name:        "per-request-model",
		BillingMode: dbmodel.BillingModePerRequest,
	})
	mustCreateModel(dbmodel.LLMInfo{
		Name:                  "per-token-concurrent",
		BillingMode:           dbmodel.BillingModePerToken,
		ProbePolicy:           dbmodel.ProbePolicyConcurrent,
		ProbeConcurrencyLimit: 2,
	})
	mustCreateModel(dbmodel.LLMInfo{
		Name:                  "per-token-serial",
		BillingMode:           dbmodel.BillingModePerToken,
		ProbePolicy:           dbmodel.ProbePolicySequential,
		ProbeConcurrencyLimit: 5,
	})
	mustCreateModel(dbmodel.LLMInfo{
		Name:                  "unknown-concurrent",
		BillingMode:           dbmodel.BillingModeUnknown,
		ProbePolicy:           dbmodel.ProbePolicyConcurrent,
		ProbeConcurrencyLimit: 3,
	})
	mustCreateModel(dbmodel.LLMInfo{
		Name:        "free-model",
		BillingMode: dbmodel.BillingModeFree,
	})
	mustCreateModel(dbmodel.LLMInfo{
		Name:        "flat-model",
		BillingMode: dbmodel.BillingModeFlat,
	})

	cases := []struct {
		name      string
		modelName string
		want      bool
	}{
		{name: "empty model defaults true", modelName: "", want: true},
		{name: "missing model defaults true", modelName: "not-found", want: true},
		{name: "per request disabled", modelName: "per-request-model", want: false},
		{name: "per token concurrent enabled", modelName: "per-token-concurrent", want: true},
		{name: "per token non concurrent disabled", modelName: "per-token-serial", want: false},
		{name: "unknown concurrent enabled", modelName: "unknown-concurrent", want: true},
		{name: "free enabled", modelName: "free-model", want: true},
		{name: "flat defaults enabled", modelName: "flat-model", want: true},
		{name: "trim and lowercase lookup", modelName: "  PER-TOKEN-CONCURRENT  ", want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := allowsRacingByModel(tc.modelName)
			if got != tc.want {
				t.Fatalf("allowsRacingByModel(%q) = %t, want %t", tc.modelName, got, tc.want)
			}
		})
	}
}

func TestFinalChannelPrefersLastSuccessOtherwiseLastFailure(t *testing.T) {
	attempts := []dbmodel.ChannelAttempt{
		{ChannelID: 1, ChannelName: "channel-a", Status: dbmodel.AttemptFailed},
		{ChannelID: 2, ChannelName: "channel-b", Status: dbmodel.AttemptSkipped},
		{ChannelID: 3, ChannelName: "channel-c", Status: dbmodel.AttemptSuccess},
		{ChannelID: 4, ChannelName: "channel-d", Status: dbmodel.AttemptFailed},
	}

	channelID, channelName := finalChannel(attempts)
	if channelID != 3 || channelName != "channel-c" {
		t.Fatalf("finalChannel(success path) = (%d, %q), want (3, %q)", channelID, channelName, "channel-c")
	}

	failedOnly := []dbmodel.ChannelAttempt{
		{ChannelID: 10, ChannelName: "channel-x", Status: dbmodel.AttemptFailed},
		{ChannelID: 11, ChannelName: "channel-y", Status: dbmodel.AttemptSkipped},
		{ChannelID: 12, ChannelName: "channel-z", Status: dbmodel.AttemptFailed},
	}
	channelID, channelName = finalChannel(failedOnly)
	if channelID != 12 || channelName != "channel-z" {
		t.Fatalf("finalChannel(failure path) = (%d, %q), want (12, %q)", channelID, channelName, "channel-z")
	}

	skippedOnly := []dbmodel.ChannelAttempt{
		{ChannelID: 21, ChannelName: "channel-skip", Status: dbmodel.AttemptSkipped},
		{ChannelID: 22, ChannelName: "channel-circuit", Status: dbmodel.AttemptCircuitBreak},
	}
	channelID, channelName = finalChannel(skippedOnly)
	if channelID != 22 || channelName != "channel-circuit" {
		t.Fatalf("finalChannel(fallback path) = (%d, %q), want (22, %q)", channelID, channelName, "channel-circuit")
	}
}

func TestRelayAttemptCopyHeadersFiltersHopByHopAndAppliesCustomHeaders(t *testing.T) {
	ctx, _ := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat", nil)
	ctx.Request.Header.Set("Authorization", "Bearer secret")
	ctx.Request.Header.Set("X-Api-Key", "secret")
	ctx.Request.Header.Set("Connection", "keep-alive")
	ctx.Request.Header.Set("X-Custom", "from-client")
	ctx.Request.Header.Set("X-Trace", "trace-1")

	outReq := httptest.NewRequest(http.MethodPost, "http://upstream.example.com", nil)
	ra := &relayAttempt{
		relayRequest: &relayRequest{c: ctx},
		channel:      &dbmodel.Channel{CustomHeader: []dbmodel.CustomHeader{{HeaderKey: "X-Custom", HeaderValue: "from-channel"}, {HeaderKey: "X-Channel", HeaderValue: "yes"}}},
	}

	ra.copyHeaders(outReq)

	if got := outReq.Header.Get("Authorization"); got != "" {
		t.Fatalf("Authorization header should be filtered, got %q", got)
	}
	if got := outReq.Header.Get("Connection"); got != "" {
		t.Fatalf("Connection header should be filtered, got %q", got)
	}
	if got := outReq.Header.Get("X-Api-Key"); got != "" {
		t.Fatalf("X-Api-Key header should be filtered, got %q", got)
	}
	if got := outReq.Header.Get("X-Trace"); got != "trace-1" {
		t.Fatalf("X-Trace header = %q, want %q", got, "trace-1")
	}
	if got := outReq.Header.Get("X-Custom"); got != "from-channel" {
		t.Fatalf("X-Custom header = %q, want channel override", got)
	}
	if got := outReq.Header.Get("X-Channel"); got != "yes" {
		t.Fatalf("X-Channel header = %q, want %q", got, "yes")
	}
}

func TestRelayAttemptHandleResponseWritesJSONBody(t *testing.T) {
	ctx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat", nil)
	body := []byte(`{"ok":true}`)
	outboundResp := &tmodel.InternalLLMResponse{ID: "resp_1", Model: "gpt-4o"}
	ra := &relayAttempt{
		relayRequest: &relayRequest{c: ctx, inAdapter: &fakeInbound{transformedBody: body}},
		outAdapter:   &fakeOutbound{response: outboundResp},
	}

	httpResp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: http.NoBody}
	if err := ra.handleResponse(context.Background(), httpResp); err != nil {
		t.Fatalf("handleResponse() error = %v", err)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d", recorder.Code, http.StatusOK)
	}
	if ct := recorder.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type = %q, want %q", ct, "application/json")
	}
	if got := recorder.Body.String(); got != string(body) {
		t.Fatalf("response body = %q, want %q", got, string(body))
	}
}

func TestRelayAttemptCollectResponseSetsMetricsFromInboundStorage(t *testing.T) {
	ctx, _ := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat", nil)
	stored := &tmodel.InternalLLMResponse{ID: "stored", Model: "gpt-4o", Usage: &tmodel.Usage{PromptTokens: 11, CompletionTokens: 7}}
	metrics := NewRelayMetrics(1, "requested-model", &tmodel.InternalLLMRequest{Model: "requested-model"})
	ra := &relayAttempt{
		relayRequest: &relayRequest{
			c:               ctx,
			inAdapter:       &fakeInbound{storedResponse: stored},
			internalRequest: &tmodel.InternalLLMRequest{Model: "actual-model"},
			metrics:         metrics,
		},
	}

	ra.collectResponse()

	if metrics.InternalResponse != stored {
		t.Fatal("collectResponse() should persist inbound stored response into metrics")
	}
	if metrics.ActualModel != "actual-model" {
		t.Fatalf("metrics actual model = %q, want %q", metrics.ActualModel, "actual-model")
	}
}

func TestRelayAttemptCollectResponseIgnoresInboundErrors(t *testing.T) {
	ctx, _ := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat", nil)
	metrics := NewRelayMetrics(1, "requested-model", &tmodel.InternalLLMRequest{Model: "requested-model"})
	ra := &relayAttempt{
		relayRequest: &relayRequest{
			c:               ctx,
			inAdapter:       &fakeInbound{storedErr: errors.New("boom")},
			internalRequest: &tmodel.InternalLLMRequest{Model: "actual-model"},
			metrics:         metrics,
		},
	}

	ra.collectResponse()

	if metrics.InternalResponse != nil {
		t.Fatal("collectResponse() should ignore inbound retrieval errors")
	}
}

func TestParseRequestPassesThroughQueryParameters(t *testing.T) {
	body := []byte(`{"model":"gpt-4o","messages":[{"role":"user","content":"hello"}]}`)
	ctx, _ := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat?foo=bar&baz=qux", body)

	internalReq, _, err := parseRequest(inbound.InboundTypeOpenAIChat, ctx)
	if err != nil {
		t.Fatalf("parseRequest() error = %v", err)
	}
	if internalReq.Model != "gpt-4o" {
		t.Fatalf("parsed model = %q, want %q", internalReq.Model, "gpt-4o")
	}
	if got := internalReq.Query.Get("foo"); got != "bar" {
		t.Fatalf("query foo = %q, want %q", got, "bar")
	}
	if got := internalReq.Query.Get("baz"); got != "qux" {
		t.Fatalf("query baz = %q, want %q", got, "qux")
	}
}

func TestParseRequestRejectsInvalidPayload(t *testing.T) {
	ctx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat", []byte(`{"model":""}`))

	internalReq, _, err := parseRequest(inbound.InboundTypeOpenAIChat, ctx)
	if err == nil {
		t.Fatal("parseRequest() should fail for invalid request body")
	}
	if internalReq != nil {
		t.Fatal("internal request should be nil on parse failure")
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("response code = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestParseRequestRejectsMalformedJSONAsBadRequest(t *testing.T) {
	ctx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat", []byte(`{"model":"gpt-4o",`))

	internalReq, _, err := parseRequest(inbound.InboundTypeOpenAIChat, ctx)
	if err == nil {
		t.Fatal("parseRequest() should fail for malformed json")
	}
	if internalReq != nil {
		t.Fatal("internal request should be nil on malformed json")
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("response code = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if strings.TrimSpace(recorder.Body.String()) == "" {
		t.Fatal("response body should contain parse failure details")
	}
}

func TestParseRequestPreservesOriginalQueryObject(t *testing.T) {
	body := []byte(`{"model":"embed-1","input":"hello"}`)
	ctx, _ := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/embeddings?encoding=float", body)
	ctx.Request.URL.RawQuery = url.Values{"encoding": []string{"float"}, "user": []string{"abc"}}.Encode()

	internalReq, _, err := parseRequest(inbound.InboundTypeOpenAIEmbedding, ctx)
	if err != nil {
		t.Fatalf("parseRequest() embedding error = %v", err)
	}
	if !internalReq.IsEmbeddingRequest() {
		t.Fatal("expected embedding request")
	}
	if got := internalReq.Query.Get("user"); got != "abc" {
		t.Fatalf("query user = %q, want %q", got, "abc")
	}
}

func TestParseRequestRejectsOversizedBody(t *testing.T) {
	originalLimit := maxRelayRequestBodyBytes
	defer func() { _ = originalLimit }()

	largeBody := bytes.Repeat([]byte("a"), int(maxRelayRequestBodyBytes)+1)
	ctx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", largeBody)

	_, _, err := parseRequest(inbound.InboundTypeOpenAIChat, ctx)
	if err == nil {
		t.Fatal("parseRequest() should fail for oversized body")
	}
	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("response code = %d, want %d", recorder.Code, http.StatusRequestEntityTooLarge)
	}
	if !strings.Contains(recorder.Body.String(), "request body too large") {
		t.Fatalf("response body = %q, want request body too large", recorder.Body.String())
	}
}

func TestHandlerRetriesWithNextKeyInSameChannel(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	responseBody, err := json.Marshal(&tmodel.InternalLLMResponse{
		ID:      "resp_fallback",
		Object:  "chat.completion",
		Created: 123,
		Model:   modelName,
		Choices: []tmodel.Choice{{Index: 0, Message: &tmodel.Message{Role: "assistant", Content: tmodel.MessageContent{Content: strPtr("ok from key2")}}}},
		Usage:   &tmodel.Usage{PromptTokens: 3, CompletionTokens: 5},
	})
	if err != nil {
		t.Fatalf("marshal response error = %v", err)
	}

	requestCounts := map[string]int{}
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		requestCounts[auth]++
		if auth == "key-1" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"key1 failed"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(responseBody)
	}))
	defer server.Close()

	_, keysByValue := setupHandlerTestRoute(t, ctxDB, modelName, server.URL, 1, 0, []string{"key-1", "key-2"})

	body := []byte(`{"model":"` + modelName + `","messages":[{"role":"user","content":"hello"}]}`)
	ginCtx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", body)
	ginCtx.Set("api_key_id", 1001)

	Handler(inbound.InboundTypeOpenAIChat, ginCtx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if requestCounts["key-1"] != 1 || requestCounts["key-2"] != 1 {
		t.Fatalf("request counts = %#v, want key-1=1 and key-2=1", requestCounts)
	}

	logs := relayLogsForTest(t)
	if len(logs) == 0 {
		t.Fatal("expected relay log entry")
	}
	attempts := logs[0].Attempts
	if len(attempts) != 2 {
		t.Fatalf("attempts len = %d, want 2", len(attempts))
	}
	if attempts[0].Status != dbmodel.AttemptFailed || attempts[0].ChannelKeyID != keysByValue["key-1"].ID || attempts[0].StatusCode != http.StatusInternalServerError {
		t.Fatalf("first attempt = %#v, want failed key-1", attempts[0])
	}
	if attempts[len(attempts)-1].Status != dbmodel.AttemptSuccess || attempts[len(attempts)-1].ChannelKeyID != keysByValue["key-2"].ID || attempts[len(attempts)-1].StatusCode != http.StatusOK {
		t.Fatalf("last attempt = %#v, want success key-2", attempts[len(attempts)-1])
	}
	if logs[0].TotalAttempts != 2 {
		t.Fatalf("TotalAttempts = %d, want 2", logs[0].TotalAttempts)
	}
	if logs[0].DynamicRoutingMode != dynamicRoutingModeHybrid {
		t.Fatalf("DynamicRoutingMode = %q, want hybrid", logs[0].DynamicRoutingMode)
	}
	if logs[0].DynamicRoutingEffectiveMode != dynamicRoutingModeStrict {
		t.Fatalf("DynamicRoutingEffectiveMode = %q, want strict fallback", logs[0].DynamicRoutingEffectiveMode)
	}
	if logs[0].DynamicRoutingDecision != dynamicRoutingDecisionDeterministic {
		t.Fatalf("DynamicRoutingDecision = %q, want deterministic", logs[0].DynamicRoutingDecision)
	}
	if !logs[0].DynamicRoutingFallback {
		t.Fatal("DynamicRoutingFallback = false, want true for low-confidence hybrid path")
	}
	if logs[0].DynamicRoutingRecommended == "" {
		t.Fatal("DynamicRoutingRecommended should be populated")
	}
}

func TestHandlerSelectsModelEligibleKeyBeforeForeignScopedKey(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	responseBody, err := json.Marshal(&tmodel.InternalLLMResponse{
		ID:      "resp_model_scoped_key",
		Object:  "chat.completion",
		Created: 123,
		Model:   modelName,
		Choices: []tmodel.Choice{{Index: 0, Message: &tmodel.Message{Role: "assistant", Content: tmodel.MessageContent{Content: strPtr("ok from eligible key")}}}},
		Usage:   &tmodel.Usage{PromptTokens: 3, CompletionTokens: 5},
	})
	if err != nil {
		t.Fatalf("marshal response error = %v", err)
	}

	requestCounts := map[string]int{}
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		requestCounts[auth]++
		if auth == "foreign-key" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"foreign model key should not be used"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(responseBody)
	}))
	defer server.Close()

	channel := &dbmodel.Channel{
		Name:              modelName + "-channel",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: server.URL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(channel, ctxDB); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	updated, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: channel.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{
		{Enabled: true, ChannelKey: "foreign-key", SourceType: "public/free", AllowedModels: "foreign-model"},
		{Enabled: true, ChannelKey: "eligible-key", SourceType: "public/free", AllowedModels: modelName},
	}}, ctxDB)
	if err != nil {
		t.Fatalf("ChannelUpdate(keys) error = %v", err)
	}
	keysByValue := make(map[string]dbmodel.ChannelKey, len(updated.Keys))
	for _, key := range updated.Keys {
		keysByValue[key.ChannelKey] = key
	}

	group := &dbmodel.Group{Name: modelName, Mode: dbmodel.GroupModeFailover}
	if err := op.GroupCreate(group, ctxDB); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: channel.ID, ModelName: modelName, Priority: 1, Weight: 1}, ctxDB); err != nil {
		t.Fatalf("GroupItemAdd() error = %v", err)
	}

	body := []byte(`{"model":"` + modelName + `","messages":[{"role":"user","content":"hello"}]}`)
	ginCtx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", body)
	ginCtx.Set("api_key_id", 1004)

	Handler(inbound.InboundTypeOpenAIChat, ginCtx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if requestCounts["foreign-key"] != 0 {
		t.Fatalf("foreign-key request count = %d, want 0", requestCounts["foreign-key"])
	}
	if requestCounts["eligible-key"] != 1 {
		t.Fatalf("eligible-key request count = %d, want 1", requestCounts["eligible-key"])
	}

	logs := relayLogsForTest(t)
	if len(logs) == 0 {
		t.Fatal("expected relay log entry")
	}
	attempts := logs[0].Attempts
	if len(attempts) != 1 {
		t.Fatalf("attempts len = %d, want 1", len(attempts))
	}
	if attempts[0].Status != dbmodel.AttemptSuccess || attempts[0].ChannelKeyID != keysByValue["eligible-key"].ID {
		t.Fatalf("attempt[0] = %#v, want success on eligible-key", attempts[0])
	}
}

func TestHandlerRetriesAcrossRoundsBeforeResponseWritten(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	responseBody, err := json.Marshal(&tmodel.InternalLLMResponse{
		ID:      "resp_retry_round",
		Object:  "chat.completion",
		Created: 123,
		Model:   modelName,
		Choices: []tmodel.Choice{{Index: 0, Message: &tmodel.Message{Role: "assistant", Content: tmodel.MessageContent{Content: strPtr("ok on second round")}}}},
	})
	if err != nil {
		t.Fatalf("marshal response error = %v", err)
	}

	requestCount := 0
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"first round failed"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(responseBody)
	}))
	defer server.Close()

	channel, keysByValue := setupHandlerTestRoute(t, ctxDB, modelName, server.URL, 2, 0, []string{"retry-key"})
	_ = channel

	body := []byte(`{"model":"` + modelName + `","messages":[{"role":"user","content":"hello"}]}`)
	ginCtx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", body)
	ginCtx.Set("api_key_id", 1002)

	Handler(inbound.InboundTypeOpenAIChat, ginCtx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if requestCount != 2 {
		t.Fatalf("requestCount = %d, want 2", requestCount)
	}

	logs := relayLogsForTest(t)
	if len(logs) == 0 {
		t.Fatal("expected relay log entry")
	}
	attempts := logs[0].Attempts
	if len(attempts) != 2 {
		t.Fatalf("attempts len = %d, want 2", len(attempts))
	}
	for i, status := range []dbmodel.AttemptStatus{dbmodel.AttemptFailed, dbmodel.AttemptSuccess} {
		if attempts[i].Status != status {
			t.Fatalf("attempt %d status = %s, want %s", i, attempts[i].Status, status)
		}
		if attempts[i].ChannelKeyID != keysByValue["retry-key"].ID {
			t.Fatalf("attempt %d key id = %d, want %d", i, attempts[i].ChannelKeyID, keysByValue["retry-key"].ID)
		}
	}
	if attempts[0].StatusCode != http.StatusInternalServerError || attempts[1].StatusCode != http.StatusOK {
		t.Fatalf("attempt status codes = [%d %d], want [%d %d]", attempts[0].StatusCode, attempts[1].StatusCode, http.StatusInternalServerError, http.StatusOK)
	}
	if logs[0].TotalAttempts != 2 {
		t.Fatalf("TotalAttempts = %d, want 2", logs[0].TotalAttempts)
	}
}

func TestHandlerSkips429KeyOnNextRequestDuringCooldown(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	responseBody, err := json.Marshal(&tmodel.InternalLLMResponse{
		ID:      "resp_cooldown",
		Object:  "chat.completion",
		Created: 123,
		Model:   modelName,
		Choices: []tmodel.Choice{{Index: 0, Message: &tmodel.Message{Role: "assistant", Content: tmodel.MessageContent{Content: strPtr("ok")}}}},
	})
	if err != nil {
		t.Fatalf("marshal response error = %v", err)
	}

	var requestMu sync.Mutex
	requestCounts := map[string]int{}
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		requestMu.Lock()
		requestCounts[auth]++
		count := requestCounts[auth]
		requestMu.Unlock()
		if auth == "key-1" && count == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limited"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(responseBody)
	}))
	defer server.Close()

	_, keysByValue := setupHandlerTestRoute(t, ctxDB, modelName, server.URL, 1, 0, []string{"key-1", "key-2"})

	makeRequest := func(apiKeyID int) *httptest.ResponseRecorder {
		body := []byte(`{"model":"` + modelName + `","messages":[{"role":"user","content":"hello"}]}`)
		ginCtx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", body)
		ginCtx.Set("api_key_id", apiKeyID)
		Handler(inbound.InboundTypeOpenAIChat, ginCtx)
		return recorder
	}

	first := makeRequest(2001)
	if first.Code != http.StatusOK {
		t.Fatalf("first response code = %d, want %d, body=%s", first.Code, http.StatusOK, first.Body.String())
	}
	second := makeRequest(2002)
	if second.Code != http.StatusOK {
		t.Fatalf("second response code = %d, want %d, body=%s", second.Code, http.StatusOK, second.Body.String())
	}

	requestMu.Lock()
	if requestCounts["key-1"] != 1 {
		requestMu.Unlock()
		t.Fatalf("key-1 request count = %d, want 1", requestCounts["key-1"])
	}
	if requestCounts["key-2"] != 2 {
		requestMu.Unlock()
		t.Fatalf("key-2 request count = %d, want 2", requestCounts["key-2"])
	}
	requestMu.Unlock()

	logs := relayLogsForTest(t)
	if len(logs) < 2 {
		t.Fatalf("relay log count = %d, want at least 2", len(logs))
	}

	secondAttempts := logs[0].Attempts
	if len(secondAttempts) != 1 {
		t.Fatalf("second request attempts len = %d, want 1", len(secondAttempts))
	}
	if secondAttempts[0].Status != dbmodel.AttemptSuccess || secondAttempts[0].ChannelKeyID != keysByValue["key-2"].ID {
		t.Fatalf("second request attempt = %#v, want direct success on key-2", secondAttempts[0])
	}

	firstAttempts := logs[1].Attempts
	if len(firstAttempts) != 2 {
		t.Fatalf("first request attempts len = %d, want 2", len(firstAttempts))
	}
	if firstAttempts[0].StatusCode != http.StatusTooManyRequests || firstAttempts[0].ChannelKeyID != keysByValue["key-1"].ID {
		t.Fatalf("first request first attempt = %#v, want 429 on key-1", firstAttempts[0])
	}
}

func TestHandlerUsesDynamicCircuitThresholdFromRouteTargetPolicy(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	if err := op.SettingSetString(dbmodel.SettingKeyCircuitBreakerThreshold, "5"); err != nil {
		t.Fatalf("SettingSetString(circuit threshold) error = %v", err)
	}

	requestCount := 0
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"always fail"}`))
	}))
	defer server.Close()

	channel := &dbmodel.Channel{
		Name:              modelName + "-channel",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: server.URL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(channel, ctxDB); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	updated, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: channel.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{
		Enabled:       true,
		ChannelKey:    "dynamic-threshold-key",
		SourceType:    dbmodel.ChannelKeySourceTypePublicFree,
		AllowedModels: modelName,
	}}}, ctxDB)
	if err != nil {
		t.Fatalf("ChannelUpdate(keys) error = %v", err)
	}
	if len(updated.Keys) != 1 {
		t.Fatalf("updated.Keys len = %d, want 1", len(updated.Keys))
	}
	key := updated.Keys[0]

	if _, err := op.RouteTargetOverrideUpsert(dbmodel.RouteTargetOverride{
		ChannelID:             updated.ID,
		ChannelKeyID:          key.ID,
		ModelName:             modelName,
		BillingMode:           dbmodel.BillingModePerRequest,
		ProbePolicy:           dbmodel.ProbePolicyPassiveOnly,
		ProbeIntervalSeconds:  3600,
		ProbeConcurrencyLimit: 1,
	}, ctxDB); err != nil {
		t.Fatalf("RouteTargetOverrideUpsert() error = %v", err)
	}

	group := &dbmodel.Group{Name: modelName, Mode: dbmodel.GroupModeFailover}
	if err := op.GroupCreate(group, ctxDB); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: updated.ID, ModelName: modelName, Priority: 1, Weight: 1}, ctxDB); err != nil {
		t.Fatalf("GroupItemAdd() error = %v", err)
	}

	body := []byte(`{"model":"` + modelName + `","messages":[{"role":"user","content":"hello"}]}`)
	makeRequest := func(apiKeyID int) *httptest.ResponseRecorder {
		ginCtx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", body)
		ginCtx.Set("api_key_id", apiKeyID)
		Handler(inbound.InboundTypeOpenAIChat, ginCtx)
		return recorder
	}

	for i := 0; i < 4; i++ {
		recorder := makeRequest(4100 + i)
		if recorder.Code != http.StatusBadGateway {
			t.Fatalf("request %d response code = %d, want %d", i+1, recorder.Code, http.StatusBadGateway)
		}
	}
	if requestCount != 4 {
		t.Fatalf("requestCount after four failing requests = %d, want 4", requestCount)
	}

	fifth := makeRequest(4205)
	if fifth.Code != http.StatusBadGateway {
		t.Fatalf("fifth response code = %d, want %d", fifth.Code, http.StatusBadGateway)
	}
	if requestCount != 4 {
		t.Fatalf("requestCount after fifth request = %d, want 4 because route-target policy should trip circuit at threshold 4", requestCount)
	}

	logs := relayLogsForTest(t)
	if len(logs) == 0 {
		t.Fatal("expected relay log entry")
	}
	attempts := logs[0].Attempts
	if len(attempts) != 1 {
		t.Fatalf("latest attempts len = %d, want 1", len(attempts))
	}
	if attempts[0].Status != dbmodel.AttemptCircuitBreak || attempts[0].ChannelKeyID != key.ID {
		t.Fatalf("latest attempt = %#v, want circuit-break record on the routed key", attempts[0])
	}
	if !strings.Contains(attempts[0].Msg, "circuit breaker tripped") {
		t.Fatalf("latest attempt msg = %q, want circuit breaker note", attempts[0].Msg)
	}
}

func TestHandlerFailoverWindowExceededDoesNotCreateSyntheticAttempt(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1200 * time.Millisecond)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"slow failure"}`))
	}))
	defer server.Close()

	channel, _ := setupHandlerTestRoute(t, ctxDB, modelName, server.URL, 2, 0, []string{"retry-key"})
	group, err := op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap() error = %v", err)
	}
	windowSec := 1
	if _, err := op.GroupUpdate(&dbmodel.GroupUpdateRequest{
		ID:                group.ID,
		FailoverWindowSec: &windowSec,
	}, ctxDB); err != nil {
		t.Fatalf("GroupUpdate() error = %v", err)
	}

	body := []byte(`{"model":"` + modelName + `","messages":[{"role":"user","content":"hello"}]}`)
	ginCtx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", body)
	ginCtx.Set("api_key_id", 3001)

	Handler(inbound.InboundTypeOpenAIChat, ginCtx)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("response code = %d, want %d, body=%s", recorder.Code, http.StatusBadGateway, recorder.Body.String())
	}

	logs := relayLogsForTest(t)
	if len(logs) == 0 {
		t.Fatal("expected relay log entry")
	}
	if got := logs[0].Attempts; len(got) != 1 {
		t.Fatalf("attempts len = %d, want 1 actual attempt only", len(got))
	}
	if strings.Contains(logs[0].Attempts[0].Msg, "failover window exceeded") {
		t.Fatalf("attempt should not contain synthetic failover-window record: %#v", logs[0].Attempts)
	}
	if !strings.Contains(logs[0].Error, "failover window exceeded") {
		t.Fatalf("relay log error = %q, want failover window exceeded", logs[0].Error)
	}
	if logs[0].ChannelId != channel.ID {
		t.Fatalf("relay log channel id = %d, want %d", logs[0].ChannelId, channel.ID)
	}
}

func TestHandlerEscalatesToLiveRaceAfterSequentialFailuresAndDoesNotDuplicateWinnerProbe(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	if err := op.SettingSetString(dbmodel.SettingKeyDynamicRoutingHealthEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(dynamic routing health) error = %v", err)
	}

	failureServer := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"sequential failure"}`))
	}))
	defer failureServer.Close()

	winnerHits := 0
	winnerServer := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		winnerHits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"race-winner","object":"chat.completion","created":1,"model":"` + modelName + `","choices":[{"index":0,"message":{"role":"assistant","content":"winner"}}],"usage":{"prompt_tokens":3,"completion_tokens":4,"total_tokens":7}}`))
	}))
	defer winnerServer.Close()

	primary, _ := setupHandlerTestRoute(t, ctxDB, modelName, failureServer.URL, 1, 0, []string{"primary-key"})
	group, err := op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap() error = %v", err)
	}
	windowSec := 360
	raceAfterFails := 2
	raceConcurrency := 2
	if _, err := op.GroupUpdate(&dbmodel.GroupUpdateRequest{
		ID:                group.ID,
		FailoverWindowSec: &windowSec,
		RaceAfterFails:    &raceAfterFails,
		RaceConcurrency:   &raceConcurrency,
	}, ctxDB); err != nil {
		t.Fatalf("GroupUpdate(primary group) error = %v", err)
	}

	second := &dbmodel.Channel{
		Name:              primary.Name + "-second",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: failureServer.URL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(second, ctxDB); err != nil {
		t.Fatalf("ChannelCreate(second) error = %v", err)
	}
	updatedSecond, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: second.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{Enabled: true, ChannelKey: "second-key", SourceType: dbmodel.ChannelKeySourceTypePublicFree, AllowedModels: modelName}}}, ctxDB)
	if err != nil {
		t.Fatalf("ChannelUpdate(second) error = %v", err)
	}

	winner := &dbmodel.Channel{
		Name:              primary.Name + "-winner",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: winnerServer.URL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(winner, ctxDB); err != nil {
		t.Fatalf("ChannelCreate(winner) error = %v", err)
	}
	updatedWinner, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: winner.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{Enabled: true, ChannelKey: "winner-key", SourceType: dbmodel.ChannelKeySourceTypePublicFree, AllowedModels: modelName}}}, ctxDB)
	if err != nil {
		t.Fatalf("ChannelUpdate(winner) error = %v", err)
	}

	group, err = op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap(refresh before add) error = %v", err)
	}
	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: second.ID, ModelName: modelName, Priority: 2, Weight: 1}, ctxDB); err != nil {
		t.Fatalf("GroupItemAdd(second) error = %v", err)
	}
	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: winner.ID, ModelName: modelName, Priority: 3, Weight: 1}, ctxDB); err != nil {
		t.Fatalf("GroupItemAdd(winner) error = %v", err)
	}

	body := []byte(`{"model":"` + modelName + `","messages":[{"role":"user","content":"hello"}]}`)
	ginCtx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", body)
	ginCtx.Set("api_key_id", 5001)

	Handler(inbound.InboundTypeOpenAIChat, ginCtx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if winnerHits != 1 {
		t.Fatalf("winnerHits = %d, want 1 so live race winner is not probed again sequentially", winnerHits)
	}

	logs := relayLogsForTest(t)
	if len(logs) == 0 {
		t.Fatal("expected relay log entry")
	}
	attempts := logs[0].Attempts
	if len(attempts) != 3 {
		t.Fatalf("attempts len = %d, want 3 (two sequential failures + one race winner)", len(attempts))
	}
	if attempts[0].Status != dbmodel.AttemptFailed || attempts[0].ChannelID != primary.ID {
		t.Fatalf("attempt[0] = %#v, want primary failure", attempts[0])
	}
	if attempts[1].Status != dbmodel.AttemptFailed || attempts[1].ChannelID != updatedSecond.ID {
		t.Fatalf("attempt[1] = %#v, want second sequential failure", attempts[1])
	}
	if attempts[2].Status != dbmodel.AttemptSuccess || attempts[2].ChannelID != updatedWinner.ID || attempts[2].ChannelKeyID != updatedWinner.Keys[0].ID {
		t.Fatalf("attempt[2] = %#v, want race winner success", attempts[2])
	}
	if attempts[2].Msg != "race fallback winner" {
		t.Fatalf("winner attempt msg = %q, want %q", attempts[2].Msg, "race fallback winner")
	}
	if logs[0].TotalAttempts != 3 {
		t.Fatalf("TotalAttempts = %d, want 3", logs[0].TotalAttempts)
	}
	if logs[0].ChannelId != updatedWinner.ID {
		t.Fatalf("relay log channel id = %d, want %d", logs[0].ChannelId, updatedWinner.ID)
	}
	if logs[0].ActualModelName != modelName {
		t.Fatalf("ActualModelName = %q, want %q", logs[0].ActualModelName, modelName)
	}
	if !strings.Contains(logs[0].ResponseContent, "race-winner") {
		t.Fatalf("response content = %q, want race winner payload", logs[0].ResponseContent)
	}
	if logs[0].Error != "" {
		t.Fatalf("relay log error = %q, want empty", logs[0].Error)
	}
}

func TestHandlerStreamingRequestDoesNotEscalateToLiveRace(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	if err := op.SettingSetString(dbmodel.SettingKeyDynamicRoutingHealthEnabled, "true"); err != nil {
		t.Fatalf("SettingSetString(dynamic routing health) error = %v", err)
	}

	firstChunk, err := json.Marshal(&tmodel.InternalLLMResponse{
		ID:      "chunk-1",
		Object:  "chat.completion.chunk",
		Created: 123,
		Model:   modelName,
		Choices: []tmodel.Choice{{Index: 0, Delta: &tmodel.Message{Role: "assistant", Content: tmodel.MessageContent{Content: strPtr("hello")}}}},
	})
	if err != nil {
		t.Fatalf("marshal chunk error = %v", err)
	}

	primaryHits := 0
	streamFailServer := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		primaryHits++
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		_, _ = w.Write([]byte("data: " + string(firstChunk) + "\n\n"))
		if flusher != nil {
			flusher.Flush()
		}
		_, _ = w.Write([]byte("data: {invalid-json}\n\n"))
		if flusher != nil {
			flusher.Flush()
		}
	}))
	defer streamFailServer.Close()

	raceHits := 0
	raceServer := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raceHits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"should-not-run","object":"chat.completion","created":1,"model":"` + modelName + `","choices":[{"index":0,"message":{"role":"assistant","content":"race"}}]}`))
	}))
	defer raceServer.Close()

	primary, _ := setupHandlerTestRoute(t, ctxDB, modelName, streamFailServer.URL, 1, 0, []string{"stream-key"})
	group, err := op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap() error = %v", err)
	}
	windowSec := 360
	raceAfterFails := 1
	raceConcurrency := 2
	if _, err := op.GroupUpdate(&dbmodel.GroupUpdateRequest{
		ID:                group.ID,
		FailoverWindowSec: &windowSec,
		RaceAfterFails:    &raceAfterFails,
		RaceConcurrency:   &raceConcurrency,
	}, ctxDB); err != nil {
		t.Fatalf("GroupUpdate() error = %v", err)
	}

	raceChannel := &dbmodel.Channel{
		Name:              primary.Name + "-race",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: raceServer.URL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(raceChannel, ctxDB); err != nil {
		t.Fatalf("ChannelCreate(race) error = %v", err)
	}
	updatedRace, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: raceChannel.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{Enabled: true, ChannelKey: "race-key", SourceType: dbmodel.ChannelKeySourceTypePublicFree, AllowedModels: modelName}}}, ctxDB)
	if err != nil {
		t.Fatalf("ChannelUpdate(race) error = %v", err)
	}
	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: raceChannel.ID, ModelName: modelName, Priority: 2, Weight: 1}, ctxDB); err != nil {
		t.Fatalf("GroupItemAdd(race) error = %v", err)
	}

	body := []byte(`{"model":"` + modelName + `","stream":true,"messages":[{"role":"user","content":"hello"}]}`)
	ginCtx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", body)
	ginCtx.Set("api_key_id", 5002)

	Handler(inbound.InboundTypeOpenAIChat, ginCtx)

	if primaryHits != 1 {
		t.Fatalf("primaryHits = %d, want 1", primaryHits)
	}
	if raceHits != 0 {
		t.Fatalf("raceHits = %d, want 0 because streaming requests should not escalate to live race", raceHits)
	}
	if !strings.Contains(recorder.Body.String(), "data: ") || !strings.Contains(recorder.Body.String(), "hello") {
		t.Fatalf("stream body = %q, want first streamed chunk", recorder.Body.String())
	}

	logs := relayLogsForTest(t)
	if len(logs) == 0 {
		t.Fatal("expected relay log entry")
	}
	attempts := logs[0].Attempts
	if len(attempts) != 1 {
		t.Fatalf("attempts len = %d, want 1", len(attempts))
	}
	if attempts[0].Status != dbmodel.AttemptFailed || attempts[0].ChannelID != primary.ID {
		t.Fatalf("attempt[0] = %#v, want only primary stream failure", attempts[0])
	}
	for _, attempt := range attempts {
		if attempt.ChannelID == updatedRace.ID || attempt.ChannelKeyID == updatedRace.Keys[0].ID {
			t.Fatalf("streaming request should not create race attempt records: %#v", attempts)
		}
	}
}

func TestHandlerSkipsStaleGroupItemWhenChannelDoesNotDeclareModel(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	requestModel := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))
	realModel := requestModel + "-real"

	responseBody, err := json.Marshal(&tmodel.InternalLLMResponse{
		ID:      "resp_declared_model",
		Object:  "chat.completion",
		Created: 123,
		Model:   realModel,
		Choices: []tmodel.Choice{{Index: 0, Message: &tmodel.Message{Role: "assistant", Content: tmodel.MessageContent{Content: strPtr("ok from declared model")}}}},
	})
	if err != nil {
		t.Fatalf("marshal response error = %v", err)
	}

	requestCounts := map[string]int{}
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		requestCounts[auth]++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(responseBody)
	}))
	defer server.Close()

	channel, keysByValue := setupHandlerTestRoute(t, ctxDB, realModel, server.URL, 1, 0, []string{"declared-key"})
	group, err := op.GroupGetMap(realModel, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap() error = %v", err)
	}
	groupName := requestModel
	if _, err := op.GroupUpdate(&dbmodel.GroupUpdateRequest{ID: group.ID, Name: &groupName}, ctxDB); err != nil {
		t.Fatalf("GroupUpdate(rename) error = %v", err)
	}
	insertRelayGroupItemDirect(t, ctxDB, &dbmodel.GroupItem{GroupID: group.ID, ChannelID: channel.ID, ModelName: requestModel, Priority: 0, Weight: 1})

	body := []byte(`{"model":"` + requestModel + `","messages":[{"role":"user","content":"hello"}]}`)
	ginCtx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", body)
	ginCtx.Set("api_key_id", 3001)

	Handler(inbound.InboundTypeOpenAIChat, ginCtx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if requestCounts["declared-key"] != 1 {
		t.Fatalf("declared-key request count = %d, want 1", requestCounts["declared-key"])
	}

	logs := relayLogsForTest(t)
	if len(logs) == 0 {
		t.Fatal("expected relay log entry")
	}
	attempts := logs[0].Attempts
	if len(attempts) != 2 {
		t.Fatalf("attempts len = %d, want 2", len(attempts))
	}
	if attempts[0].Status != dbmodel.AttemptSkipped || !strings.Contains(attempts[0].Msg, "does not declare model") {
		t.Fatalf("first attempt = %#v, want skipped stale route item", attempts[0])
	}
	if attempts[1].Status != dbmodel.AttemptSuccess || attempts[1].ChannelKeyID != keysByValue["declared-key"].ID {
		t.Fatalf("second attempt = %#v, want success on declared route item", attempts[1])
	}
}

func TestHandlerStopsRetryingAfterStreamWrite(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	firstChunk, err := json.Marshal(&tmodel.InternalLLMResponse{
		ID:      "chunk-1",
		Object:  "chat.completion.chunk",
		Created: 123,
		Model:   modelName,
		Choices: []tmodel.Choice{{Index: 0, Delta: &tmodel.Message{Role: "assistant", Content: tmodel.MessageContent{Content: strPtr("hello")}}}},
	})
	if err != nil {
		t.Fatalf("marshal chunk error = %v", err)
	}

	requestCounts := map[string]int{}
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		requestCounts[auth]++
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		_, _ = w.Write([]byte("data: " + string(firstChunk) + "\n\n"))
		if flusher != nil {
			flusher.Flush()
		}
		_, _ = w.Write([]byte("data: {invalid-json}\n\n"))
		if flusher != nil {
			flusher.Flush()
		}
	}))
	defer server.Close()

	_, keysByValue := setupHandlerTestRoute(t, ctxDB, modelName, server.URL, 2, 0, []string{"stream-key-1", "stream-key-2"})

	body := []byte(`{"model":"` + modelName + `","stream":true,"messages":[{"role":"user","content":"hello"}]}`)
	ginCtx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", body)
	ginCtx.Set("api_key_id", 1003)

	Handler(inbound.InboundTypeOpenAIChat, ginCtx)

	if requestCounts["stream-key-1"] != 1 {
		t.Fatalf("stream-key-1 request count = %d, want 1", requestCounts["stream-key-1"])
	}
	if requestCounts["stream-key-2"] != 0 {
		t.Fatalf("stream-key-2 request count = %d, want 0", requestCounts["stream-key-2"])
	}
	if !strings.Contains(recorder.Body.String(), "data: ") || !strings.Contains(recorder.Body.String(), "hello") {
		t.Fatalf("stream body = %q, want first streamed chunk", recorder.Body.String())
	}

	logs := relayLogsForTest(t)
	if len(logs) == 0 {
		t.Fatal("expected relay log entry")
	}
	attempts := logs[0].Attempts
	if len(attempts) != 1 {
		t.Fatalf("attempts len = %d, want 1", len(attempts))
	}
	if attempts[0].Status != dbmodel.AttemptFailed || attempts[0].ChannelKeyID != keysByValue["stream-key-1"].ID {
		t.Fatalf("attempt[0] = %#v, want failed first key only", attempts[0])
	}
}

func TestRunRaceFallbackRecordsUnavailableCandidates(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"race-win","object":"chat.completion","created":1,"model":"` + modelName + `","choices":[{"index":0,"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer server.Close()

	primary, _ := setupHandlerTestRoute(t, ctxDB, modelName, server.URL, 1, 0, []string{"primary-key"})
	group, err := op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap() error = %v", err)
	}
	insertRelayGroupItemDirect(t, ctxDB, &dbmodel.GroupItem{GroupID: group.ID, ChannelID: 999999, ModelName: modelName, Priority: 2, Weight: 1})

	fallback := &dbmodel.Channel{
		Name:              primary.Name + "-fallback",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: server.URL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(fallback, ctxDB); err != nil {
		t.Fatalf("ChannelCreate(fallback) error = %v", err)
	}
	updatedFallback, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: fallback.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{Enabled: true, ChannelKey: "fallback-key", SourceType: "public/free", AllowedModels: modelName}}}, ctxDB)
	if err != nil {
		t.Fatalf("ChannelUpdate(fallback) error = %v", err)
	}
	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: fallback.ID, ModelName: modelName, Priority: 3, Weight: 1}, ctxDB); err != nil {
		t.Fatalf("GroupItemAdd(fallback) error = %v", err)
	}

	group, err = op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap(refresh) error = %v", err)
	}

	iter := balancer.NewIterator(group, 0, modelName)
	if !iter.Next() {
		t.Fatal("iterator should have primary candidate")
	}
	ctx, _ := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", []byte(`{"model":"`+modelName+`","messages":[{"role":"user","content":"hello"}]}`))
	req := &relayRequest{
		c:               ctx,
		inAdapter:       &fakeInbound{},
		internalRequest: &tmodel.InternalLLMRequest{Model: modelName, Messages: []tmodel.Message{{Role: "user", Content: tmodel.MessageContent{Content: strPtr("hello")}}}},
		metrics:         NewRelayMetrics(0, modelName, &tmodel.InternalLLMRequest{Model: modelName}),
		requestModel:    modelName,
		iter:            iter,
	}

	result, ok := runRaceFallback(req, iter, time.Now().Add(10*time.Second), 3)
	if !ok {
		t.Fatal("runRaceFallback should execute when next candidates exist")
	}
	if !result.Success {
		t.Fatalf("runRaceFallback result = %#v, want success", result)
	}
	if result.Channel == nil || result.Channel.ID != updatedFallback.ID {
		t.Fatalf("race result channel = %#v, want fallback channel %d", result.Channel, updatedFallback.ID)
	}
	if len(updatedFallback.Keys) == 0 || result.UsedKey.ID != updatedFallback.Keys[0].ID {
		t.Fatalf("race result key id = %d, want fallback key id", result.UsedKey.ID)
	}

	attempts := iter.Attempts()
	foundUnavailable := false
	for _, attempt := range attempts {
		if attempt.ChannelID == 999999 && attempt.Status == dbmodel.AttemptSkipped && strings.Contains(attempt.Msg, "race candidate channel unavailable") {
			foundUnavailable = true
		}
	}
	if !foundUnavailable {
		t.Fatalf("attempts = %#v, want unavailable race candidate record", attempts)
	}
}

func TestRunRaceFallbackPrefersEarlierSuccessfulCandidate(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))
	if err := op.LLMCreate(dbmodel.LLMInfo{
		Name:                  modelName,
		BillingMode:           dbmodel.BillingModePerToken,
		ProbePolicy:           dbmodel.ProbePolicyConcurrent,
		ProbeConcurrencyLimit: 2,
		LLMPrice: dbmodel.LLMPrice{
			Input:  2,
			Output: 4,
		},
	}, ctxDB); err != nil {
		t.Fatalf("LLMCreate() error = %v", err)
	}

	fastSecond := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"earlier-success","object":"chat.completion","created":1,"model":"` + modelName + `","choices":[{"index":0,"message":{"role":"assistant","content":"earlier"}}],"usage":{"prompt_tokens":5,"completion_tokens":7,"total_tokens":12}}`))
	}))
	defer fastSecond.Close()

	slowerThird := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"later-success","object":"chat.completion","created":1,"model":"` + modelName + `","choices":[{"index":0,"message":{"role":"assistant","content":"later"}}],"usage":{"prompt_tokens":11,"completion_tokens":13,"total_tokens":24}}`))
	}))
	defer slowerThird.Close()

	primary, _ := setupHandlerTestRoute(t, ctxDB, modelName, fastSecond.URL, 1, 0, []string{"primary-key"})
	group, err := op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap() error = %v", err)
	}

	second := &dbmodel.Channel{
		Name:              primary.Name + "-second",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: fastSecond.URL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(second, ctxDB); err != nil {
		t.Fatalf("ChannelCreate(second) error = %v", err)
	}
	if _, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: second.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{Enabled: true, ChannelKey: "second-key", SourceType: "public/free", AllowedModels: modelName}}}, ctxDB); err != nil {
		t.Fatalf("ChannelUpdate(second) error = %v", err)
	}

	third := &dbmodel.Channel{
		Name:              primary.Name + "-third",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: slowerThird.URL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(third, ctxDB); err != nil {
		t.Fatalf("ChannelCreate(third) error = %v", err)
	}
	if _, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: third.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{Enabled: true, ChannelKey: "third-key", SourceType: "public/free", AllowedModels: modelName}}}, ctxDB); err != nil {
		t.Fatalf("ChannelUpdate(third) error = %v", err)
	}

	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: second.ID, ModelName: modelName, Priority: 2, Weight: 1}, ctxDB); err != nil {
		t.Fatalf("GroupItemAdd(second) error = %v", err)
	}
	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: third.ID, ModelName: modelName, Priority: 3, Weight: 1}, ctxDB); err != nil {
		t.Fatalf("GroupItemAdd(third) error = %v", err)
	}
	group, err = op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap(refresh) error = %v", err)
	}

	iter := balancer.NewIterator(group, 0, modelName)
	if !iter.Next() {
		t.Fatal("iterator should have primary candidate")
	}
	ctx, _ := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", []byte(`{"model":"`+modelName+`","messages":[{"role":"user","content":"hello"}]}`))
	req := &relayRequest{
		c:               ctx,
		inAdapter:       &fakeInbound{},
		internalRequest: &tmodel.InternalLLMRequest{Model: modelName, Messages: []tmodel.Message{{Role: "user", Content: tmodel.MessageContent{Content: strPtr("hello")}}}},
		metrics:         NewRelayMetrics(0, modelName, &tmodel.InternalLLMRequest{Model: modelName}),
		requestModel:    modelName,
		iter:            iter,
	}

	result, ok := runRaceFallback(req, iter, time.Now().Add(10*time.Second), 3)
	if !ok {
		t.Fatal("runRaceFallback should execute")
	}
	if !result.Success || result.Channel == nil {
		t.Fatalf("result = %#v, want successful race result", result)
	}
	if result.Channel.ID != second.ID {
		t.Fatalf("race winner channel id = %d, want earlier candidate %d", result.Channel.ID, second.ID)
	}

	attempts := iter.Attempts()
	foundWinner := false
	foundLaterSuccess := false
	for _, attempt := range attempts {
		if attempt.ChannelID == second.ID && attempt.Status == dbmodel.AttemptSuccess {
			foundWinner = true
		}
		if attempt.ChannelID == third.ID && attempt.Status == dbmodel.AttemptCanceled {
			foundLaterSuccess = true
		}
	}
	if !foundWinner {
		t.Fatalf("attempts = %#v, want winner attempt record for second candidate", attempts)
	}
	if !foundLaterSuccess {
		t.Fatalf("attempts = %#v, want canceled record for later successful probe", attempts)
	}

	refreshedThird, err := op.ChannelGet(third.ID, ctxDB)
	if err != nil {
		t.Fatalf("ChannelGet(third) error = %v", err)
	}
	if len(refreshedThird.Keys) != 1 {
		t.Fatalf("refreshedThird.Keys len = %d, want 1", len(refreshedThird.Keys))
	}
	if got := refreshedThird.Keys[0].TotalCost; got <= 0 {
		t.Fatalf("third key total_cost = %f, want > 0 after successful non-selected probe", got)
	}

	probeSummary := op.ProbeSummaryGet(24 * time.Hour)
	if probeSummary.SuccessCount < 2 {
		t.Fatalf("ProbeSummary success count = %d, want at least 2", probeSummary.SuccessCount)
	}
	if probeSummary.EstimatedTotalCost <= 0 {
		t.Fatalf("ProbeSummary estimated total cost = %f, want > 0 for successful non-selected probe", probeSummary.EstimatedTotalCost)
	}
}

func TestRunRaceFallbackCapsConcurrencyToRemainingCandidates(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	requestCounts := map[string]int{}
	var requestMu sync.Mutex
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		requestMu.Lock()
		requestCounts[auth]++
		requestMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"race-win","object":"chat.completion","created":1,"model":"` + modelName + `","choices":[{"index":0,"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer server.Close()

	primary, _ := setupHandlerTestRoute(t, ctxDB, modelName, server.URL, 1, 0, []string{"primary-key"})
	group, err := op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap() error = %v", err)
	}

	second := &dbmodel.Channel{
		Name:              primary.Name + "-second",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: server.URL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(second, ctxDB); err != nil {
		t.Fatalf("ChannelCreate(second) error = %v", err)
	}
	if _, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: second.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{Enabled: true, ChannelKey: "second-key", SourceType: "public/free", AllowedModels: modelName}}}, ctxDB); err != nil {
		t.Fatalf("ChannelUpdate(second) error = %v", err)
	}
	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: second.ID, ModelName: modelName, Priority: 2, Weight: 1}, ctxDB); err != nil {
		t.Fatalf("GroupItemAdd(second) error = %v", err)
	}
	group, err = op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap(refresh) error = %v", err)
	}

	iter := balancer.NewIterator(group, 0, modelName)
	if !iter.Next() {
		t.Fatal("iterator should have primary candidate")
	}
	ctx, _ := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", []byte(`{"model":"`+modelName+`","messages":[{"role":"user","content":"hello"}]}`))
	req := &relayRequest{
		c:               ctx,
		inAdapter:       &fakeInbound{},
		internalRequest: &tmodel.InternalLLMRequest{Model: modelName, Messages: []tmodel.Message{{Role: "user", Content: tmodel.MessageContent{Content: strPtr("hello")}}}},
		metrics:         NewRelayMetrics(0, modelName, &tmodel.InternalLLMRequest{Model: modelName}),
		requestModel:    modelName,
		iter:            iter,
	}

	result, ok := runRaceFallback(req, iter, time.Now().Add(10*time.Second), 5)
	if !ok {
		t.Fatal("runRaceFallback should execute when one remaining candidate exists")
	}
	if !result.Success || result.Channel == nil {
		t.Fatalf("result = %#v, want successful race result", result)
	}
	if result.Channel.ID != second.ID {
		t.Fatalf("race winner channel id = %d, want %d", result.Channel.ID, second.ID)
	}

	requestMu.Lock()
	defer requestMu.Unlock()
	if requestCounts["second-key"] != 1 {
		t.Fatalf("second-key request count = %d, want 1", requestCounts["second-key"])
	}
	if requestCounts["primary-key"] != 0 {
		t.Fatalf("primary-key request count = %d, want 0", requestCounts["primary-key"])
	}
}

func TestFinalizeRaceFallbackSuccessReturnsErrorWhenInboundTransformFails(t *testing.T) {
	ctx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", nil)
	iter := balancer.NewIterator(dbmodel.Group{
		ID:   1,
		Mode: dbmodel.GroupModeFailover,
		Items: []dbmodel.GroupItem{
			{ChannelID: 11, ModelName: "gpt-4o", Priority: 1},
		},
	}, 0, "gpt-4o")
	if !iter.Next() {
		t.Fatal("iterator should have one item")
	}
	metrics := NewRelayMetrics(0, "gpt-4o", &tmodel.InternalLLMRequest{Model: "gpt-4o"})
	req := &relayRequest{
		c:            ctx,
		inAdapter:    &fakeInbound{responseErr: errors.New("transform failed")},
		metrics:      metrics,
		apiKeyID:     7,
		requestModel: "gpt-4o",
		iter:         iter,
	}
	channel := &dbmodel.Channel{ID: 11, Name: "race-channel"}
	result := attemptResult{
		Success:          true,
		StatusCode:       http.StatusOK,
		Duration:         35 * time.Millisecond,
		Channel:          channel,
		UsedKey:          dbmodel.ChannelKey{ID: 101, ChannelID: 11, ChannelKey: "key-1"},
		InternalResponse: &tmodel.InternalLLMResponse{ID: "resp_1", Model: "gpt-4o"},
		ActualModel:      "gpt-4o",
	}

	err := req.finalizeRaceFallbackSuccess(result)
	if err == nil {
		t.Fatal("finalizeRaceFallbackSuccess() should fail when inbound transform fails")
	}
	if recorder.Code != http.StatusOK && recorder.Code != 0 {
		t.Fatalf("unexpected response code = %d", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("response body should stay empty on transform failure, got %q", recorder.Body.String())
	}
	attempts := iter.Attempts()
	if len(attempts) != 1 {
		t.Fatalf("attempts len = %d, want 1", len(attempts))
	}
	if attempts[0].Status != dbmodel.AttemptFailed {
		t.Fatalf("attempt status = %s, want %s", attempts[0].Status, dbmodel.AttemptFailed)
	}
	if attempts[0].StatusCode != http.StatusOK {
		t.Fatalf("attempt status code = %d, want %d", attempts[0].StatusCode, http.StatusOK)
	}
	if attempts[0].Duration != 35 {
		t.Fatalf("attempt duration = %d, want 35", attempts[0].Duration)
	}
	if !strings.Contains(attempts[0].Msg, "race fallback inbound transform failed") {
		t.Fatalf("attempt msg = %q, want transform failure note", attempts[0].Msg)
	}
}

func TestFinalizeRaceFallbackSuccessWritesResponseAndRecordsWinner(t *testing.T) {
	ctx, recorder := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", nil)
	iter := balancer.NewIterator(dbmodel.Group{
		ID:   1,
		Mode: dbmodel.GroupModeFailover,
		Items: []dbmodel.GroupItem{
			{ChannelID: 11, ModelName: "gpt-4o", Priority: 1},
		},
	}, 0, "gpt-4o")
	if !iter.Next() {
		t.Fatal("iterator should have one item")
	}
	metrics := NewRelayMetrics(0, "gpt-4o", &tmodel.InternalLLMRequest{Model: "gpt-4o"})
	metrics.Stats.InputCost = 1.2
	metrics.Stats.OutputCost = 0.8
	req := &relayRequest{
		c:            ctx,
		inAdapter:    &fakeInbound{transformedBody: []byte(`{"ok":true}`)},
		metrics:      metrics,
		apiKeyID:     7,
		requestModel: "gpt-4o",
		iter:         iter,
	}
	channel := &dbmodel.Channel{ID: 11, Name: "race-channel"}
	usedKey := dbmodel.ChannelKey{ID: 101, ChannelID: 11, ChannelKey: "key-1"}
	result := attemptResult{
		Success:          true,
		StatusCode:       http.StatusOK,
		Duration:         35 * time.Millisecond,
		Channel:          channel,
		UsedKey:          usedKey,
		InternalResponse: &tmodel.InternalLLMResponse{ID: "resp_1", Model: "gpt-4o"},
		ActualModel:      "gpt-4o",
	}

	if err := req.finalizeRaceFallbackSuccess(result); err != nil {
		t.Fatalf("finalizeRaceFallbackSuccess() error = %v", err)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Body.String(); got != `{"ok":true}` {
		t.Fatalf("response body = %q, want %q", got, `{"ok":true}`)
	}
	attempts := iter.Attempts()
	if len(attempts) != 1 {
		t.Fatalf("attempts len = %d, want 1", len(attempts))
	}
	if attempts[0].Status != dbmodel.AttemptSuccess {
		t.Fatalf("attempt status = %s, want %s", attempts[0].Status, dbmodel.AttemptSuccess)
	}
	if attempts[0].StatusCode != http.StatusOK {
		t.Fatalf("attempt status code = %d, want %d", attempts[0].StatusCode, http.StatusOK)
	}
	if attempts[0].Duration != 35 {
		t.Fatalf("attempt duration = %d, want 35", attempts[0].Duration)
	}
	if attempts[0].Msg != "race fallback winner" {
		t.Fatalf("attempt msg = %q, want %q", attempts[0].Msg, "race fallback winner")
	}
	if metrics.InternalResponse == nil || metrics.ActualModel != "gpt-4o" {
		t.Fatal("metrics should capture race fallback internal response and actual model")
	}
}

func TestRelayLogJSONContentIsTruncatedPerStringField(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	longText := strings.Repeat("a", relayLogMaxStringRunes+128)
	longReasoning := strings.Repeat("b", relayLogMaxStringRunes+64)
	longArgument := strings.Repeat("c", relayLogMaxStringRunes+96)
	longResponse := strings.Repeat("d", relayLogMaxStringRunes+160)

	metrics := NewRelayMetrics(1, "gpt-4o", &tmodel.InternalLLMRequest{
		Model: "gpt-4o",
		Messages: []tmodel.Message{{
			Role:    "user",
			Content: tmodel.MessageContent{Content: strPtr(longText)},
			ToolCalls: []tmodel.ToolCall{{
				ID:   "tool-1",
				Type: "function",
				Function: tmodel.FunctionCall{
					Name:      "search",
					Arguments: longArgument,
				},
			}},
			ReasoningContent: strPtr(longReasoning),
		}},
	})
	metrics.SetInternalResponse(&tmodel.InternalLLMResponse{
		ID:      "resp_1",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gpt-4o",
		Choices: []tmodel.Choice{{
			Index: 0,
			Message: &tmodel.Message{
				Role:    "assistant",
				Content: tmodel.MessageContent{Content: strPtr(longResponse)},
			},
		}},
	}, "gpt-4o")

	metrics.saveLog(ctxDB, nil, 10*time.Millisecond, nil, 0, "")

	logs := relayLogsForTest(t)
	if len(logs) != 1 {
		t.Fatalf("logs len = %d, want 1", len(logs))
	}

	var requestPayload map[string]any
	if err := json.Unmarshal([]byte(logs[0].RequestContent), &requestPayload); err != nil {
		t.Fatalf("request content should remain valid JSON: %v", err)
	}
	messages, ok := requestPayload["messages"].([]any)
	if !ok || len(messages) != 1 {
		t.Fatalf("request messages = %#v, want single message", requestPayload["messages"])
	}
	message, ok := messages[0].(map[string]any)
	if !ok {
		t.Fatalf("request message = %#v, want object", messages[0])
	}
	if got := message["content"].(string); got != truncateRelayLogString(longText) {
		t.Fatalf("request content length = %d, want truncated length %d", len([]rune(got)), len([]rune(truncateRelayLogString(longText))))
	}
	if got := message["reasoning_content"].(string); got != truncateRelayLogString(longReasoning) {
		t.Fatalf("reasoning content not truncated as expected")
	}
	toolCalls, ok := message["tool_calls"].([]any)
	if !ok || len(toolCalls) != 1 {
		t.Fatalf("tool_calls = %#v, want single call", message["tool_calls"])
	}
	toolCall, ok := toolCalls[0].(map[string]any)
	if !ok {
		t.Fatalf("tool_call = %#v, want object", toolCalls[0])
	}
	function, ok := toolCall["function"].(map[string]any)
	if !ok {
		t.Fatalf("function = %#v, want object", toolCall["function"])
	}
	if got := function["arguments"].(string); got != truncateRelayLogString(longArgument) {
		t.Fatalf("tool arguments not truncated as expected")
	}

	var responsePayload map[string]any
	if err := json.Unmarshal([]byte(logs[0].ResponseContent), &responsePayload); err != nil {
		t.Fatalf("response content should remain valid JSON: %v", err)
	}
	choices, ok := responsePayload["choices"].([]any)
	if !ok || len(choices) != 1 {
		t.Fatalf("response choices = %#v, want single choice", responsePayload["choices"])
	}
	choice, ok := choices[0].(map[string]any)
	if !ok {
		t.Fatalf("response choice = %#v, want object", choices[0])
	}
	respMessage, ok := choice["message"].(map[string]any)
	if !ok {
		t.Fatalf("response message = %#v, want object", choice["message"])
	}
	if got := respMessage["content"].(string); got != truncateRelayLogString(longResponse) {
		t.Fatalf("response content not truncated as expected")
	}
	if !strings.Contains(logs[0].RequestContent, relayLogTruncatedSuffix) {
		t.Fatalf("request content = %q, want truncation marker", logs[0].RequestContent)
	}
	if !strings.Contains(logs[0].ResponseContent, relayLogTruncatedSuffix) {
		t.Fatalf("response content = %q, want truncation marker", logs[0].ResponseContent)
	}
}

func TestRelayMetricsSavePersistsCacheTokenUsage(t *testing.T) {
	ctxDB := setupRelayTestDB(t)

	metrics := NewRelayMetrics(1, "gpt-4o", &tmodel.InternalLLMRequest{
		Model: "gpt-4o",
		Messages: []tmodel.Message{{
			Role:    "user",
			Content: tmodel.MessageContent{Content: strPtr("hello")},
		}},
	})
	metrics.SetInternalResponse(&tmodel.InternalLLMResponse{
		ID:      "resp_cache_usage",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gpt-4o",
		Usage: &tmodel.Usage{
			PromptTokens:     120,
			CompletionTokens: 45,
			TotalTokens:      181,
			PromptTokensDetails: &tmodel.PromptTokensDetails{
				CachedTokens: 30,
			},
			CacheCreationInputTokens: 16,
		},
		Choices: []tmodel.Choice{{
			Index: 0,
			Message: &tmodel.Message{
				Role:    "assistant",
				Content: tmodel.MessageContent{Content: strPtr("ok")},
			},
		}},
	}, "gpt-4o")

	metrics.saveLog(ctxDB, nil, 10*time.Millisecond, nil, 0, "")

	logs := relayLogsForTest(t)
	if len(logs) != 1 {
		t.Fatalf("logs len = %d, want 1", len(logs))
	}
	if logs[0].CacheReadTokens != 30 {
		t.Fatalf("cache read tokens = %d, want 30", logs[0].CacheReadTokens)
	}
	if logs[0].CacheWriteTokens != 16 {
		t.Fatalf("cache write tokens = %d, want 16", logs[0].CacheWriteTokens)
	}
	wantCost := metrics.Stats.InputCost + metrics.Stats.OutputCost
	if logs[0].Cost != wantCost {
		t.Fatalf("cost = %f, want %f", logs[0].Cost, wantCost)
	}
	if !strings.Contains(logs[0].ResponseContent, `"cache_creation_input_tokens":16`) {
		t.Fatalf("response content = %q, want cache_creation_input_tokens injected", logs[0].ResponseContent)
	}
}

func TestRelayMetricsSaveUsesRequestContextForDailyStats(t *testing.T) {
	setupRelayTestDB(t)

	originalUpdate := relayStatsDailyUpdate
	relayStatsDailyUpdate = func(ctx context.Context, metrics dbmodel.StatsMetrics) error {
		if ctx == nil {
			t.Fatal("stats daily update ctx = nil, want request context")
		}
		if ctx.Err() == nil {
			t.Fatal("stats daily update ctx should be canceled after request cancellation")
		}
		return nil
	}
	t.Cleanup(func() { relayStatsDailyUpdate = originalUpdate })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	metrics := NewRelayMetrics(1, "gpt-4o", &tmodel.InternalLLMRequest{Model: "gpt-4o"})
	metrics.Save(ctx, true, nil, nil)
}

func TestRunRaceFallbackCancelsSlowerLowerPriorityProbeAfterWinner(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	winnerRelease := make(chan struct{})
	fastServer := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-winnerRelease
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"winner","object":"chat.completion","created":1,"model":"` + modelName + `","choices":[{"index":0,"message":{"role":"assistant","content":"winner"}}]}`))
	}))
	defer fastServer.Close()

	slowStarted := make(chan struct{}, 1)
	slowServer := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case slowStarted <- struct{}{}:
		default:
		}
		select {
		case <-r.Context().Done():
			return
		case <-time.After(2 * time.Second):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"late","object":"chat.completion","created":1,"model":"` + modelName + `","choices":[{"index":0,"message":{"role":"assistant","content":"late"}}]}`))
		}
	}))
	defer slowServer.Close()

	primary, _ := setupHandlerTestRoute(t, ctxDB, modelName, fastServer.URL, 1, 0, []string{"primary-key"})
	group, err := op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap() error = %v", err)
	}

	second := &dbmodel.Channel{
		Name:              primary.Name + "-second",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: fastServer.URL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(second, ctxDB); err != nil {
		t.Fatalf("ChannelCreate(second) error = %v", err)
	}
	if _, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: second.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{Enabled: true, ChannelKey: "second-key", SourceType: "public/free", AllowedModels: modelName}}}, ctxDB); err != nil {
		t.Fatalf("ChannelUpdate(second) error = %v", err)
	}

	third := &dbmodel.Channel{
		Name:              primary.Name + "-third",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: slowServer.URL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(third, ctxDB); err != nil {
		t.Fatalf("ChannelCreate(third) error = %v", err)
	}
	if _, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: third.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{Enabled: true, ChannelKey: "third-key", SourceType: "public/free", AllowedModels: modelName}}}, ctxDB); err != nil {
		t.Fatalf("ChannelUpdate(third) error = %v", err)
	}

	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: second.ID, ModelName: modelName, Priority: 2, Weight: 1}, ctxDB); err != nil {
		t.Fatalf("GroupItemAdd(second) error = %v", err)
	}
	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: third.ID, ModelName: modelName, Priority: 3, Weight: 1}, ctxDB); err != nil {
		t.Fatalf("GroupItemAdd(third) error = %v", err)
	}
	group, err = op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap(refresh) error = %v", err)
	}

	iter := balancer.NewIterator(group, 0, modelName)
	if !iter.Next() {
		t.Fatal("iterator should have primary candidate")
	}
	ctx, _ := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", []byte(`{"model":"`+modelName+`","messages":[{"role":"user","content":"hello"}]}`))
	req := &relayRequest{
		c:               ctx,
		inAdapter:       &fakeInbound{},
		internalRequest: &tmodel.InternalLLMRequest{Model: modelName, Messages: []tmodel.Message{{Role: "user", Content: tmodel.MessageContent{Content: strPtr("hello")}}}},
		metrics:         NewRelayMetrics(0, modelName, &tmodel.InternalLLMRequest{Model: modelName}),
		requestModel:    modelName,
		iter:            iter,
	}

	go func() {
		select {
		case <-slowStarted:
			close(winnerRelease)
		case <-time.After(2 * time.Second):
		}
	}()

	start := time.Now()
	result, ok := runRaceFallback(req, iter, time.Now().Add(10*time.Second), 3)
	if !ok {
		t.Fatal("runRaceFallback should execute")
	}
	if !result.Success || result.Channel == nil || result.Channel.ID != second.ID {
		t.Fatalf("result = %#v, want second channel winner", result)
	}
	if elapsed := time.Since(start); elapsed >= 1500*time.Millisecond {
		t.Fatalf("runRaceFallback elapsed = %s, want early return before slow lower-priority probe finishes", elapsed)
	}
	attempts := iter.Attempts()
	foundWinner := false
	foundCanceled := false
	for _, attempt := range attempts {
		if attempt.ChannelID == second.ID && attempt.Status == dbmodel.AttemptSuccess {
			foundWinner = true
		}
		if attempt.ChannelID == third.ID && attempt.Status == dbmodel.AttemptFailed {
			t.Fatalf("lower-priority canceled race probe should not be recorded as failed: %#v", attempts)
		}
		if attempt.ChannelID == third.ID && attempt.Status == dbmodel.AttemptCanceled {
			foundCanceled = true
		}
	}
	if !foundWinner {
		t.Fatalf("attempts = %#v, want winner attempt record for the selected race candidate", attempts)
	}
	if !foundCanceled {
		t.Logf("attempts = %#v; slower loser did not surface a canceled attempt before the race returned, but it also was not mis-recorded as failed", attempts)
	}
}

func TestRunRaceFallbackRecordsProbeFailureAttempt(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	failingServer := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"probe failed"}`))
	}))
	defer failingServer.Close()

	primary, _ := setupHandlerTestRoute(t, ctxDB, modelName, failingServer.URL, 1, 0, []string{"primary-key"})
	group, err := op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap() error = %v", err)
	}

	second := &dbmodel.Channel{
		Name:              primary.Name + "-second",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: failingServer.URL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(second, ctxDB); err != nil {
		t.Fatalf("ChannelCreate(second) error = %v", err)
	}
	updatedSecond, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: second.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{Enabled: true, ChannelKey: "second-key", SourceType: "public/free", AllowedModels: modelName}}}, ctxDB)
	if err != nil {
		t.Fatalf("ChannelUpdate(second) error = %v", err)
	}
	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: second.ID, ModelName: modelName, Priority: 2, Weight: 1}, ctxDB); err != nil {
		t.Fatalf("GroupItemAdd(second) error = %v", err)
	}
	group, err = op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap(refresh) error = %v", err)
	}

	iter := balancer.NewIterator(group, 0, modelName)
	if !iter.Next() {
		t.Fatal("iterator should have primary candidate")
	}
	ctx, _ := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", []byte(`{"model":"`+modelName+`","messages":[{"role":"user","content":"hello"}]}`))
	req := &relayRequest{
		c:               ctx,
		inAdapter:       &fakeInbound{},
		internalRequest: &tmodel.InternalLLMRequest{Model: modelName, Messages: []tmodel.Message{{Role: "user", Content: tmodel.MessageContent{Content: strPtr("hello")}}}},
		metrics:         NewRelayMetrics(0, modelName, &tmodel.InternalLLMRequest{Model: modelName}),
		requestModel:    modelName,
		iter:            iter,
	}

	result, ok := runRaceFallback(req, iter, time.Now().Add(10*time.Second), 2)
	if !ok {
		t.Fatal("runRaceFallback should execute")
	}
	if result.Success {
		t.Fatalf("result = %#v, want failure", result)
	}

	attempts := iter.Attempts()
	foundFailure := false
	for _, attempt := range attempts {
		if attempt.ChannelID == updatedSecond.ID && attempt.ChannelKeyID == updatedSecond.Keys[0].ID && attempt.Status == dbmodel.AttemptFailed && attempt.StatusCode == http.StatusBadGateway && strings.Contains(attempt.Msg, "race probe upstream error") {
			foundFailure = true
		}
	}
	if !foundFailure {
		t.Fatalf("attempts = %#v, want recorded race probe failure for second candidate", attempts)
	}
}

func TestRunRaceFallbackSkipsCandidateWhenRaceBudgetExhausted(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	if err := op.SettingSetString(dbmodel.SettingKeyRaceGlobalBudget, "1"); err != nil {
		t.Fatalf("SettingSetString(global budget) error = %v", err)
	}
	if err := op.SettingSetString(dbmodel.SettingKeyRaceProbeBudget, "1"); err != nil {
		t.Fatalf("SettingSetString(probe budget) error = %v", err)
	}

	releaseHeld, err := acquireRaceBudgets(context.Background(), 999, 999, 999)
	if err != nil {
		t.Fatalf("acquireRaceBudgets(seed) error = %v", err)
	}
	defer releaseHeld()

	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"race-win","object":"chat.completion","created":1,"model":"` + modelName + `","choices":[{"index":0,"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer server.Close()

	primary, _ := setupHandlerTestRoute(t, ctxDB, modelName, server.URL, 1, 0, []string{"primary-key"})
	group, err := op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap() error = %v", err)
	}

	second := &dbmodel.Channel{
		Name:              primary.Name + "-second",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: server.URL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(second, ctxDB); err != nil {
		t.Fatalf("ChannelCreate(second) error = %v", err)
	}
	updatedSecond, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: second.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{Enabled: true, ChannelKey: "second-key", SourceType: "public/free", AllowedModels: modelName}}}, ctxDB)
	if err != nil {
		t.Fatalf("ChannelUpdate(second) error = %v", err)
	}
	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: second.ID, ModelName: modelName, Priority: 2, Weight: 1}, ctxDB); err != nil {
		t.Fatalf("GroupItemAdd(second) error = %v", err)
	}
	group, err = op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap(refresh) error = %v", err)
	}

	iter := balancer.NewIterator(group, 0, modelName)
	if !iter.Next() {
		t.Fatal("iterator should have primary candidate")
	}
	ctx, _ := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", []byte(`{"model":"`+modelName+`","messages":[{"role":"user","content":"hello"}]}`))
	req := &relayRequest{
		c:               ctx,
		inAdapter:       &fakeInbound{},
		internalRequest: &tmodel.InternalLLMRequest{Model: modelName, Messages: []tmodel.Message{{Role: "user", Content: tmodel.MessageContent{Content: strPtr("hello")}}}},
		metrics:         NewRelayMetrics(0, modelName, &tmodel.InternalLLMRequest{Model: modelName}),
		requestModel:    modelName,
		iter:            iter,
	}

	result, ok := runRaceFallback(req, iter, time.Now().Add(10*time.Second), 2)
	if !ok {
		t.Fatal("runRaceFallback should execute")
	}
	if result.Success {
		t.Fatalf("result = %#v, want no race success when budget exhausted", result)
	}

	attempts := iter.Attempts()
	foundBudgetSkip := false
	for _, attempt := range attempts {
		if attempt.ChannelID == updatedSecond.ID && attempt.ChannelKeyID == updatedSecond.Keys[0].ID && attempt.Status == dbmodel.AttemptSkipped && strings.Contains(attempt.Msg, "budget exhausted") {
			foundBudgetSkip = true
		}
	}
	if !foundBudgetSkip {
		t.Fatalf("attempts = %#v, want budget exhaustion skip record", attempts)
	}
}

func TestRunRaceFallbackSkipsRouteTargetBlockedCandidateAndUsesLaterWinner(t *testing.T) {
	ctxDB := setupRelayTestDB(t)
	modelName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	blockedServer := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"blocked","object":"chat.completion","created":1,"model":"` + modelName + `","choices":[{"index":0,"message":{"role":"assistant","content":"blocked"}}]}`))
	}))
	defer blockedServer.Close()

	winnerServer := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"winner","object":"chat.completion","created":1,"model":"` + modelName + `","choices":[{"index":0,"message":{"role":"assistant","content":"winner"}}]}`))
	}))
	defer winnerServer.Close()

	primary, _ := setupHandlerTestRoute(t, ctxDB, modelName, blockedServer.URL, 1, 0, []string{"primary-key"})
	group, err := op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap() error = %v", err)
	}

	blocked := &dbmodel.Channel{
		Name:              primary.Name + "-blocked",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: blockedServer.URL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(blocked, ctxDB); err != nil {
		t.Fatalf("ChannelCreate(blocked) error = %v", err)
	}
	blockedUpdated, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: blocked.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{Enabled: true, ChannelKey: "blocked-key", SourceType: "public/free", AllowedModels: modelName}}}, ctxDB)
	if err != nil {
		t.Fatalf("ChannelUpdate(blocked) error = %v", err)
	}
	if _, err := op.RouteTargetOverrideUpsert(dbmodel.RouteTargetOverride{ChannelID: blockedUpdated.ID, ChannelKeyID: blockedUpdated.Keys[0].ID, ModelName: modelName, BillingMode: dbmodel.BillingModePerRequest}, ctxDB); err != nil {
		t.Fatalf("RouteTargetOverrideUpsert(blocked) error = %v", err)
	}
	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: blocked.ID, ModelName: modelName, Priority: 2, Weight: 1}, ctxDB); err != nil {
		t.Fatalf("GroupItemAdd(blocked) error = %v", err)
	}

	winner := &dbmodel.Channel{
		Name:              primary.Name + "-winner",
		Type:              outbound.OutboundTypeOpenAIChat,
		Enabled:           true,
		KeyManagementMode: dbmodel.KeyManagementModeClassified,
		KeyRoutingPolicy:  dbmodel.KeyRoutingPolicyFillPriority,
		BaseUrls:          []dbmodel.BaseUrl{{URL: winnerServer.URL, Delay: 1}},
		Model:             modelName,
	}
	if err := op.ChannelCreate(winner, ctxDB); err != nil {
		t.Fatalf("ChannelCreate(winner) error = %v", err)
	}
	winnerUpdated, err := op.ChannelUpdate(&dbmodel.ChannelUpdateRequest{ID: winner.ID, KeysToAdd: []dbmodel.ChannelKeyAddRequest{{Enabled: true, ChannelKey: "winner-key", SourceType: "public/free", AllowedModels: modelName}}}, ctxDB)
	if err != nil {
		t.Fatalf("ChannelUpdate(winner) error = %v", err)
	}
	if err := op.GroupItemAdd(&dbmodel.GroupItem{GroupID: group.ID, ChannelID: winner.ID, ModelName: modelName, Priority: 3, Weight: 1}, ctxDB); err != nil {
		t.Fatalf("GroupItemAdd(winner) error = %v", err)
	}

	group, err = op.GroupGetMap(modelName, ctxDB)
	if err != nil {
		t.Fatalf("GroupGetMap(refresh) error = %v", err)
	}

	iter := balancer.NewIterator(group, 0, modelName)
	if !iter.Next() {
		t.Fatal("iterator should have primary candidate")
	}
	ctx, _ := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", []byte(`{"model":"`+modelName+`","messages":[{"role":"user","content":"hello"}]}`))
	req := &relayRequest{
		c:               ctx,
		inAdapter:       &fakeInbound{},
		internalRequest: &tmodel.InternalLLMRequest{Model: modelName, Messages: []tmodel.Message{{Role: "user", Content: tmodel.MessageContent{Content: strPtr("hello")}}}},
		metrics:         NewRelayMetrics(0, modelName, &tmodel.InternalLLMRequest{Model: modelName}),
		requestModel:    modelName,
		iter:            iter,
	}

	result, ok := runRaceFallback(req, iter, time.Now().Add(10*time.Second), 3)
	if !ok {
		t.Fatal("runRaceFallback should execute")
	}
	if !result.Success {
		t.Fatalf("result = %#v, want later winner success", result)
	}
	if result.Channel == nil || result.Channel.ID != winnerUpdated.ID {
		t.Fatalf("result channel = %#v, want winner channel %d", result.Channel, winnerUpdated.ID)
	}

	attempts := iter.Attempts()
	foundBlockedSkip := false
	foundWinner := false
	for _, attempt := range attempts {
		if attempt.ChannelID == blockedUpdated.ID && attempt.Status == dbmodel.AttemptSkipped && strings.Contains(attempt.Msg, "route-target forbids racing") {
			foundBlockedSkip = true
		}
		if attempt.ChannelID == winnerUpdated.ID && attempt.Status == dbmodel.AttemptSuccess {
			foundWinner = true
		}
	}
	if !foundBlockedSkip {
		t.Fatalf("attempts = %#v, want route-target skip record", attempts)
	}
	if !foundWinner {
		t.Fatalf("attempts = %#v, want later winner attempt record", attempts)
	}
}

func TestRelayAttemptForwardLimitsUpstreamErrorBody(t *testing.T) {
	largeBody := strings.Repeat("x", int(maxUpstreamErrorBodyBytes)+1024)
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(largeBody))
	}))
	defer server.Close()

	ctx, _ := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", nil)
	ra := &relayAttempt{
		relayRequest: &relayRequest{
			c:               ctx,
			inAdapter:       &fakeInbound{},
			internalRequest: &tmodel.InternalLLMRequest{Model: "gpt-4o"},
			metrics:         NewRelayMetrics(0, "gpt-4o", &tmodel.InternalLLMRequest{Model: "gpt-4o"}),
		},
		outAdapter: &fakeOutbound{},
		channel: &dbmodel.Channel{
			Type:     outbound.OutboundTypeOpenAIChat,
			Enabled:  true,
			BaseUrls: []dbmodel.BaseUrl{{URL: server.URL, Delay: 1}},
		},
		usedKey: dbmodel.ChannelKey{ChannelKey: "test-key"},
	}

	statusCode, err := ra.forward()
	if err == nil {
		t.Fatal("forward() expected upstream error")
	}
	if statusCode != http.StatusBadGateway {
		t.Fatalf("statusCode = %d, want %d", statusCode, http.StatusBadGateway)
	}
	errText := err.Error()
	if !strings.Contains(errText, "upstream error: 502: ") {
		t.Fatalf("unexpected error text: %q", errText)
	}
	prefixLen := len("upstream error: 502: ")
	if got := len(errText) - prefixLen; got != int(maxUpstreamErrorBodyBytes) {
		t.Fatalf("captured error body bytes = %d, want %d", got, maxUpstreamErrorBodyBytes)
	}
}

func TestRunRaceProbeLimitsUpstreamErrorBody(t *testing.T) {
	largeBody := strings.Repeat("y", int(maxUpstreamErrorBodyBytes)+2048)
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(largeBody))
	}))
	defer server.Close()

	ctx, _ := newRelayAttemptTestContext(http.MethodPost, "http://example.com/v1/chat/completions", nil)
	req := &relayRequest{
		c:               ctx,
		inAdapter:       &fakeInbound{},
		internalRequest: &tmodel.InternalLLMRequest{Model: "gpt-4o"},
		metrics:         NewRelayMetrics(0, "gpt-4o", &tmodel.InternalLLMRequest{Model: "gpt-4o"}),
	}
	channel := &dbmodel.Channel{
		Type:     outbound.OutboundTypeOpenAIChat,
		Enabled:  true,
		BaseUrls: []dbmodel.BaseUrl{{URL: server.URL, Delay: 1}},
	}
	usedKey := dbmodel.ChannelKey{ChannelKey: "race-key"}
	internalReq := &tmodel.InternalLLMRequest{Model: "gpt-4o"}

	result := runRaceProbe(context.Background(), req, channel, usedKey, &fakeOutbound{}, internalReq)
	if result.Success {
		t.Fatalf("result = %#v, want failure", result)
	}
	if result.StatusCode != http.StatusBadGateway {
		t.Fatalf("result status code = %d, want %d", result.StatusCode, http.StatusBadGateway)
	}
	if result.Err == nil {
		t.Fatal("runRaceProbe() expected upstream error")
	}
	errText := result.Err.Error()
	if !strings.Contains(errText, "race probe upstream error: 502: ") {
		t.Fatalf("unexpected error text: %q", errText)
	}
	prefixLen := len("race probe upstream error: 502: ")
	if got := len(errText) - prefixLen; got != int(maxUpstreamErrorBodyBytes) {
		t.Fatalf("captured race-probe error body bytes = %d, want %d", got, maxUpstreamErrorBodyBytes)
	}
}
