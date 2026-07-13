package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

func createTestGroupForBatch(t *testing.T, name string) *model.Group {
	t.Helper()
	group := &model.Group{Name: name}
	if err := op.GroupCreate(group, context.Background()); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	return group
}

func createTestAPIKeyForBatch(t *testing.T, name string) *model.APIKey {
	t.Helper()
	apiKey := &model.APIKey{
		Name:    name,
		APIKey:  "sk-octopus-test-batch-" + name,
		Enabled: true,
	}
	if err := op.APIKeyCreate(apiKey, context.Background()); err != nil {
		t.Fatalf("APIKeyCreate() error = %v", err)
	}
	return apiKey
}

func TestBatchEnableChannelSuccess(t *testing.T) {
	setupHandlerTest(t)

	ch1 := createTestChannel(t, "batch-enable-ch1")
	ch2 := createTestChannel(t, "batch-enable-ch2")
	// 先禁用
	if err := op.ChannelEnabled(ch1.ID, false, context.Background()); err != nil {
		t.Fatalf("ChannelEnabled(false) error = %v", err)
	}
	if err := op.ChannelEnabled(ch2.ID, false, context.Background()); err != nil {
		t.Fatalf("ChannelEnabled(false) error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/batch-enable",
		BatchOperationRequest{IDs: []int{ch1.ID, ch2.ID}}, batchEnableChannel)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	res := decodeHandlerResponse(t, recorder)
	var result BatchOperationResult
	if err := json.Unmarshal(res.Data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if result.SuccessCount != 2 || result.FailedCount != 0 {
		t.Fatalf("result = %+v, want success=2 failed=0", result)
	}

	ch1After, err := op.ChannelGet(ch1.ID, context.Background())
	if err != nil {
		t.Fatalf("ChannelGet() error = %v", err)
	}
	if !ch1After.Enabled {
		t.Fatalf("channel %d not enabled after batch operation", ch1.ID)
	}
}

func TestBatchDisableChannelSuccess(t *testing.T) {
	setupHandlerTest(t)

	ch1 := createTestChannel(t, "batch-disable-ch1")
	ch2 := createTestChannel(t, "batch-disable-ch2")

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/batch-disable",
		BatchOperationRequest{IDs: []int{ch1.ID, ch2.ID}}, batchDisableChannel)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	res := decodeHandlerResponse(t, recorder)
	var result BatchOperationResult
	if err := json.Unmarshal(res.Data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if result.SuccessCount != 2 || result.FailedCount != 0 {
		t.Fatalf("result = %+v, want success=2 failed=0", result)
	}

	ch1After, err := op.ChannelGet(ch1.ID, context.Background())
	if err != nil {
		t.Fatalf("ChannelGet() error = %v", err)
	}
	if ch1After.Enabled {
		t.Fatalf("channel %d not disabled after batch operation", ch1.ID)
	}
}

func TestBatchDeleteChannelSuccess(t *testing.T) {
	setupHandlerTest(t)

	ch1 := createTestChannel(t, "batch-delete-ch1")
	ch2 := createTestChannel(t, "batch-delete-ch2")

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/batch-delete",
		BatchOperationRequest{IDs: []int{ch1.ID, ch2.ID}}, batchDeleteChannel)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	res := decodeHandlerResponse(t, recorder)
	var result BatchOperationResult
	if err := json.Unmarshal(res.Data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if result.SuccessCount != 2 || result.FailedCount != 0 {
		t.Fatalf("result = %+v, want success=2 failed=0", result)
	}

	if _, err := op.ChannelGet(ch1.ID, context.Background()); err == nil {
		t.Fatalf("channel %d still exists after batch delete", ch1.ID)
	}
}

func TestBatchDeleteGroupSuccess(t *testing.T) {
	setupHandlerTest(t)

	g1 := createTestGroupForBatch(t, "batch-delete-group1")
	g2 := createTestGroupForBatch(t, "batch-delete-group2")

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/group/batch-delete",
		BatchOperationRequest{IDs: []int{g1.ID, g2.ID}}, batchDeleteGroup)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	res := decodeHandlerResponse(t, recorder)
	var result BatchOperationResult
	if err := json.Unmarshal(res.Data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if result.SuccessCount != 2 || result.FailedCount != 0 {
		t.Fatalf("result = %+v, want success=2 failed=0", result)
	}

	if _, err := op.GroupGet(g1.ID, context.Background()); err == nil {
		t.Fatalf("group %d still exists after batch delete", g1.ID)
	}
}

func TestBatchEnableAPIKeySuccess(t *testing.T) {
	setupHandlerTest(t)

	key1 := createTestAPIKeyForBatch(t, "batch-enable-key1")
	key2 := createTestAPIKeyForBatch(t, "batch-enable-key2")
	// 先禁用
	key1.Enabled = false
	if err := op.APIKeyUpdate(key1, context.Background()); err != nil {
		t.Fatalf("APIKeyUpdate() error = %v", err)
	}
	key2.Enabled = false
	if err := op.APIKeyUpdate(key2, context.Background()); err != nil {
		t.Fatalf("APIKeyUpdate() error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/apikey/batch-enable",
		BatchOperationRequest{IDs: []int{key1.ID, key2.ID}}, batchEnableAPIKey)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	res := decodeHandlerResponse(t, recorder)
	var result BatchOperationResult
	if err := json.Unmarshal(res.Data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if result.SuccessCount != 2 || result.FailedCount != 0 {
		t.Fatalf("result = %+v, want success=2 failed=0", result)
	}

	key1After, err := op.APIKeyGet(key1.ID, context.Background())
	if err != nil {
		t.Fatalf("APIKeyGet() error = %v", err)
	}
	if !key1After.Enabled {
		t.Fatalf("apikey %d not enabled after batch operation", key1.ID)
	}
}

func TestBatchDisableAPIKeySuccess(t *testing.T) {
	setupHandlerTest(t)

	key1 := createTestAPIKeyForBatch(t, "batch-disable-key1")
	key2 := createTestAPIKeyForBatch(t, "batch-disable-key2")

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/apikey/batch-disable",
		BatchOperationRequest{IDs: []int{key1.ID, key2.ID}}, batchDisableAPIKey)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	res := decodeHandlerResponse(t, recorder)
	var result BatchOperationResult
	if err := json.Unmarshal(res.Data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if result.SuccessCount != 2 || result.FailedCount != 0 {
		t.Fatalf("result = %+v, want success=2 failed=0", result)
	}

	key1After, err := op.APIKeyGet(key1.ID, context.Background())
	if err != nil {
		t.Fatalf("APIKeyGet() error = %v", err)
	}
	if key1After.Enabled {
		t.Fatalf("apikey %d not disabled after batch operation", key1.ID)
	}
}

func TestBatchDeleteAPIKeySuccess(t *testing.T) {
	setupHandlerTest(t)

	key1 := createTestAPIKeyForBatch(t, "batch-delete-key1")
	key2 := createTestAPIKeyForBatch(t, "batch-delete-key2")

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/apikey/batch-delete",
		BatchOperationRequest{IDs: []int{key1.ID, key2.ID}}, batchDeleteAPIKey)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	res := decodeHandlerResponse(t, recorder)
	var result BatchOperationResult
	if err := json.Unmarshal(res.Data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if result.SuccessCount != 2 || result.FailedCount != 0 {
		t.Fatalf("result = %+v, want success=2 failed=0", result)
	}

	if _, err := op.APIKeyGet(key1.ID, context.Background()); err == nil {
		t.Fatalf("apikey %d still exists after batch delete", key1.ID)
	}
}

func TestBatchOperationMixedSuccessAndFailure(t *testing.T) {
	setupHandlerTest(t)

	ch1 := createTestChannel(t, "batch-mixed-ch1")
	// ch2 不存在（使用一个很大的 ID）

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/batch-delete",
		BatchOperationRequest{IDs: []int{ch1.ID, 999999}}, batchDeleteChannel)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	res := decodeHandlerResponse(t, recorder)
	var result BatchOperationResult
	if err := json.Unmarshal(res.Data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if result.SuccessCount != 1 || result.FailedCount != 1 {
		t.Fatalf("result = %+v, want success=1 failed=1", result)
	}
	if len(result.Errors) != 1 {
		t.Fatalf("errors length = %d, want 1", len(result.Errors))
	}
}

func TestBatchRequestRejectsEmptyIDs(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/batch-enable",
		BatchOperationRequest{IDs: []int{}}, batchEnableChannel)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestBatchRequestRejectsTooManyIDs(t *testing.T) {
	setupHandlerTest(t)

	ids := make([]int, batchMaxIDs+1)
	for i := range ids {
		ids[i] = i + 1
	}
	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/batch-enable",
		BatchOperationRequest{IDs: ids}, batchEnableChannel)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestBatchRequestRejectsNonPositiveIDs(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/batch-enable",
		BatchOperationRequest{IDs: []int{0, -1, -2}}, batchEnableChannel)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestBatchRequestDeduplicatesIDs(t *testing.T) {
	setupHandlerTest(t)

	ch1 := createTestChannel(t, "batch-dedup-ch1")
	if err := op.ChannelEnabled(ch1.ID, false, context.Background()); err != nil {
		t.Fatalf("ChannelEnabled(false) error = %v", err)
	}

	// 重复 ID 应被去重
	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/batch-enable",
		BatchOperationRequest{IDs: []int{ch1.ID, ch1.ID, ch1.ID}}, batchEnableChannel)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	res := decodeHandlerResponse(t, recorder)
	var result BatchOperationResult
	if err := json.Unmarshal(res.Data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if result.SuccessCount != 1 {
		t.Fatalf("result.SuccessCount = %d, want 1 (after dedup)", result.SuccessCount)
	}
}

func TestBatchRequestRejectsInvalidJSON(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/channel/batch-enable",
		"not-json", batchEnableChannel)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
