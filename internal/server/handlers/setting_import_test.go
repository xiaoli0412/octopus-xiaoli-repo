package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

func TestExportDBDefaultsToFullSnapshotAndCanExplicitlyRedactSecrets(t *testing.T) {
	setupHandlerTest(t)

	proxyURL := `https://octopus:secret@example.com:8443`
	channel := model.Channel{
		Name:    `export-channel`,
		Enabled: true,
		CustomHeader: []model.CustomHeader{
			{HeaderKey: `Authorization`, HeaderValue: `Bearer upstream-secret`},
			{HeaderKey: `X-Workspace-ID`, HeaderValue: `workspace-1`},
		},
		ChannelProxy: &proxyURL,
	}
	if err := db.GetDB().Create(&channel).Error; err != nil {
		t.Fatalf(`create channel error = %v`, err)
	}
	channelKey := model.ChannelKey{ChannelID: channel.ID, Enabled: true, ChannelKey: `sk-channel-secret`}
	if err := db.GetDB().Create(&channelKey).Error; err != nil {
		t.Fatalf(`create channel key error = %v`, err)
	}
	apiKey := model.APIKey{Name: `export-api-key`, APIKey: `sk-api-secret`, Enabled: true}
	if err := db.GetDB().Create(&apiKey).Error; err != nil {
		t.Fatalf(`create api key error = %v`, err)
	}

	defaultRecorder := httptest.NewRecorder()
	defaultCtx, _ := gin.CreateTestContext(defaultRecorder)
	defaultCtx.Request = httptest.NewRequest(http.MethodGet, `/api/v1/setting/export`, nil)
	exportDB(defaultCtx)
	if defaultRecorder.Code != http.StatusOK {
		t.Fatalf(`default status = %d, want %d, body = %s`, defaultRecorder.Code, http.StatusOK, defaultRecorder.Body.String())
	}
	var defaultDump model.DBDump
	if err := json.Unmarshal(defaultRecorder.Body.Bytes(), &defaultDump); err != nil {
		t.Fatalf(`json.Unmarshal(defaultDump) error = %v`, err)
	}
	if !defaultDump.Manifest.ContainsSecrets {
		t.Fatalf(`Manifest.ContainsSecrets = false, want true by default`)
	}
	if len(defaultDump.ChannelKeys) == 0 || defaultDump.ChannelKeys[0].ChannelKey != `sk-channel-secret` {
		t.Fatalf(`default channel key = %#v, want exported credential`, defaultDump.ChannelKeys)
	}
	if len(defaultDump.APIKeys) == 0 || defaultDump.APIKeys[0].APIKey != `sk-api-secret` {
		t.Fatalf(`default api key = %#v, want exported credential`, defaultDump.APIKeys)
	}
	if got := defaultDump.Channels[0].CustomHeader[0].HeaderValue; got != `Bearer upstream-secret` {
		t.Fatalf(`default auth header value = %q, want exported credential`, got)
	}
	if got := defaultDump.Channels[0].CustomHeader[1].HeaderValue; got != `workspace-1` {
		t.Fatalf(`default workspace header value = %q, want preserved`, got)
	}
	if defaultDump.Channels[0].ChannelProxy == nil || *defaultDump.Channels[0].ChannelProxy != proxyURL {
		t.Fatalf(`default channel proxy = %#v, want preserved proxy URL`, defaultDump.Channels[0].ChannelProxy)
	}

	redactedRecorder := httptest.NewRecorder()
	redactedCtx, _ := gin.CreateTestContext(redactedRecorder)
	redactedCtx.Request = httptest.NewRequest(http.MethodGet, `/api/v1/setting/export?include_secrets=false`, nil)
	exportDB(redactedCtx)
	if redactedRecorder.Code != http.StatusOK {
		t.Fatalf(`redacted status = %d, want %d, body = %s`, redactedRecorder.Code, http.StatusOK, redactedRecorder.Body.String())
	}
	var redactedDump model.DBDump
	if err := json.Unmarshal(redactedRecorder.Body.Bytes(), &redactedDump); err != nil {
		t.Fatalf(`json.Unmarshal(redactedDump) error = %v`, err)
	}
	if redactedDump.Manifest.ContainsSecrets {
		t.Fatalf(`Manifest.ContainsSecrets = true, want false when include_secrets=false`)
	}
	if len(redactedDump.Users) == 0 {
		t.Fatalf(`redacted users = %#v, want exported admin user`, redactedDump.Users)
	}
	if redactedDump.Users[0].Password != `` {
		t.Fatalf(`redacted user password = %q, want empty credential`, redactedDump.Users[0].Password)
	}
	if len(redactedDump.ChannelKeys) == 0 || redactedDump.ChannelKeys[0].ChannelKey != `` {
		t.Fatalf(`redacted channel key = %#v, want empty credential`, redactedDump.ChannelKeys)
	}
	if len(redactedDump.APIKeys) == 0 || redactedDump.APIKeys[0].APIKey != `` {
		t.Fatalf(`redacted api key = %#v, want empty credential`, redactedDump.APIKeys)
	}
	if got := redactedDump.Channels[0].CustomHeader[0].HeaderValue; got != `` {
		t.Fatalf(`redacted auth header value = %q, want empty credential`, got)
	}
	if got := redactedDump.Channels[0].CustomHeader[1].HeaderValue; got != `workspace-1` {
		t.Fatalf(`redacted workspace header value = %q, want preserved non-secret value`, got)
	}
	if redactedDump.Channels[0].ChannelProxy == nil || *redactedDump.Channels[0].ChannelProxy != `https://example.com:8443` {
		t.Fatalf(`redacted channel proxy = %#v, want credentials stripped and endpoint preserved`, redactedDump.Channels[0].ChannelProxy)
	}
}

func TestExportDBSupportsLegacyFormat(t *testing.T) {
	setupHandlerTest(t)

	channel := model.Channel{Name: `legacy-export-channel`, Enabled: true, KeyManagementMode: model.KeyManagementModeClassified, KeyRoutingPolicy: model.KeyRoutingPolicyPriority}
	if err := db.GetDB().Create(&channel).Error; err != nil {
		t.Fatalf(`create channel error = %v`, err)
	}
	channelKey := model.ChannelKey{ChannelID: channel.ID, Enabled: true, ChannelKey: `sk-channel-secret`, SourceType: model.ChannelKeySourceTypePaidMetered, AllowedModels: `gpt-4o`}
	if err := db.GetDB().Create(&channelKey).Error; err != nil {
		t.Fatalf(`create channel key error = %v`, err)
	}
	llmInfo := model.LLMInfo{Name: `gpt-4o`, CanonicalName: `gpt-4o`, BillingMode: model.BillingModePerToken, ProbePolicy: model.ProbePolicyConcurrent, ProbeIntervalSeconds: 60, ProbeConcurrencyLimit: 2}
	if err := db.GetDB().Create(&llmInfo).Error; err != nil {
		t.Fatalf(`create llm info error = %v`, err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, `/api/v1/setting/export?format=legacy`, nil)
	exportDB(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf(`json.Unmarshal(payload) error = %v`, err)
	}
	if _, ok := payload[`manifest`]; ok {
		t.Fatalf(`legacy payload should not contain manifest: %#v`, payload)
	}
	channels, ok := payload[`channels`].([]any)
	if !ok || len(channels) != 1 {
		t.Fatalf(`channels = %#v, want one legacy channel`, payload[`channels`])
	}
	channelRow := channels[0].(map[string]any)
	if _, ok := channelRow[`key_management_mode`]; ok {
		t.Fatalf(`legacy channel row leaked key_management_mode: %#v`, channelRow)
	}
	if _, ok := channelRow[`key_routing_policy`]; ok {
		t.Fatalf(`legacy channel row leaked key_routing_policy: %#v`, channelRow)
	}
	channelKeys, ok := payload[`channel_keys`].([]any)
	if !ok || len(channelKeys) != 1 {
		t.Fatalf(`channel_keys = %#v, want one legacy key row`, payload[`channel_keys`])
	}
	channelKeyRow := channelKeys[0].(map[string]any)
	if _, ok := channelKeyRow[`source_type`]; ok {
		t.Fatalf(`legacy channel key row leaked source_type: %#v`, channelKeyRow)
	}
	if _, ok := channelKeyRow[`allowed_models`]; ok {
		t.Fatalf(`legacy channel key row leaked allowed_models: %#v`, channelKeyRow)
	}
	llmInfos, ok := payload[`llm_infos`].([]any)
	if !ok || len(llmInfos) != 1 {
		t.Fatalf(`llm_infos = %#v, want one legacy llm info row`, payload[`llm_infos`])
	}
	llmInfoRow := llmInfos[0].(map[string]any)
	if _, ok := llmInfoRow[`canonical_name`]; ok {
		t.Fatalf(`legacy llm info row leaked canonical_name: %#v`, llmInfoRow)
	}
	if _, ok := llmInfoRow[`billing_mode`]; ok {
		t.Fatalf(`legacy llm info row leaked billing_mode: %#v`, llmInfoRow)
	}
}

func TestExportDBRejectsInvalidBooleanQueryValues(t *testing.T) {
	setupHandlerTest(t)

	for _, target := range []string{
		`/api/v1/setting/export?include_logs=maybe`,
		`/api/v1/setting/export?include_logs=`,
		`/api/v1/setting/export?include_logs=%20`,
		`/api/v1/setting/export?include_stats=definitely-not`,
		`/api/v1/setting/export?include_stats=`,
		`/api/v1/setting/export?include_stats=%20`,
		`/api/v1/setting/export?include_secrets=redacted-please`,
		`/api/v1/setting/export?include_secrets=`,
		`/api/v1/setting/export?include_secrets=%20`,
	} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, target, nil)
		exportDB(ctx)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf(`target %s status = %d, want %d, body = %s`, target, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
	}
}

func TestExportDBRejectsInvalidFormatQueryValue(t *testing.T) {
	setupHandlerTest(t)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, `/api/v1/setting/export?format=surprise`, nil)

	exportDB(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, `unsupported export format`) {
		t.Fatalf(`message = %q, want unsupported export format`, response.Message)
	}
}

func TestExportDBRejectsBlankFormatQueryValue(t *testing.T) {
	setupHandlerTest(t)

	for _, target := range []string{
		`/api/v1/setting/export?format=`,
		`/api/v1/setting/export?format=%20`,
	} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, target, nil)
		exportDB(ctx)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf(`target %s status = %d, want %d, body = %s`, target, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
		response := decodeHandlerResponse(t, recorder)
		if !strings.Contains(response.Message, `unsupported export format`) {
			t.Fatalf(`target %s message = %q, want unsupported export format`, target, response.Message)
		}
	}
}

func TestDecodeDBDumpNormalizesLegacyRawStructure(t *testing.T) {
	legacyJSON := []byte(`{"version":1,"exported_at":"2026-04-21T17:45:36Z","include_logs":false,"include_stats":false,"channels":[{"id":1,"name":"legacy-channel","type":0,"enabled":true,"base_urls":[{"url":"https://legacy.example.com/v1","delay":1}],"keys":null,"model":"gpt-4o","custom_model":""}],"channel_keys":[{"id":2,"channel_id":1,"enabled":true,"channel_key":"sk-legacy"}],"groups":[{"id":3,"name":"legacy-group","mode":1}],"group_items":[{"id":4,"group_id":3,"channel_id":1,"model_name":"gpt-4o","priority":1,"weight":1}],"llm_infos":[{"name":"gpt-4o","input":1,"output":2}],"api_keys":[{"id":5,"name":"client","api_key":"sk-client","enabled":true}],"settings":[{"key":"api_base_url","value":"https://legacy.example.com"}]}`)
	var dump model.DBDump
	if err := decodeDBDump(legacyJSON, &dump); err != nil {
		t.Fatalf(`decodeDBDump() error = %v`, err)
	}
	if dump.Manifest.ExportSource != `octopus-legacy` {
		t.Fatalf(`manifest.export_source = %q, want octopus-legacy`, dump.Manifest.ExportSource)
	}
	if !dump.Manifest.ContainsSecrets {
		t.Fatalf(`manifest.contains_secrets = false, want true`)
	}
	if dump.LegacyHints == nil || !dump.LegacyHints.Legacy {
		t.Fatalf(`legacy_hints = %#v, want legacy=true`, dump.LegacyHints)
	}
	if dump.Channels[0].KeyManagementMode != model.KeyManagementModePooled {
		t.Fatalf(`channel.key_management_mode = %q, want pooled fallback`, dump.Channels[0].KeyManagementMode)
	}
	if dump.ChannelKeys[0].SourceType != model.ChannelKeySourceTypeUnknown {
		t.Fatalf(`channel_key.source_type = %q, want unknown fallback`, dump.ChannelKeys[0].SourceType)
	}
	if dump.LLMInfos[0].CanonicalName != `gpt-4o` {
		t.Fatalf(`llm_info.canonical_name = %q, want gpt-4o`, dump.LLMInfos[0].CanonicalName)
	}
	if dump.LLMInfos[0].ProbePolicy != model.ProbePolicyPassiveOnly {
		t.Fatalf(`llm_info.probe_policy = %q, want passive_only fallback`, dump.LLMInfos[0].ProbePolicy)
	}
}

func TestImportDBDryRunReturnsPreviewTokenAndApplyWithSameTokenSucceeds(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://preview-token.example.com`)

	dryRunRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=incremental`, dump, nil)
	if dryRunRecorder.Code != http.StatusOK {
		t.Fatalf(`dry-run status = %d, want %d, body = %s`, dryRunRecorder.Code, http.StatusOK, dryRunRecorder.Body.String())
	}
	dryRunResponse := decodeHandlerResponse(t, dryRunRecorder)
	var dryRunResult model.DBImportResult
	if err := json.Unmarshal(dryRunResponse.Data, &dryRunResult); err != nil {
		t.Fatalf(`json.Unmarshal(dryRunResult) error = %v`, err)
	}
	if strings.TrimSpace(dryRunResult.PreviewToken) == `` {
		t.Fatalf(`preview_token = %q, want non-empty`, dryRunResult.PreviewToken)
	}

	applyRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=incremental`, dump, map[string]string{
		`preview_token`: dryRunResult.PreviewToken,
	})
	if applyRecorder.Code != http.StatusOK {
		t.Fatalf(`apply status = %d, want %d, body = %s`, applyRecorder.Code, http.StatusOK, applyRecorder.Body.String())
	}
	applyResponse := decodeHandlerResponse(t, applyRecorder)
	var applyResult model.DBImportResult
	if err := json.Unmarshal(applyResponse.Data, &applyResult); err != nil {
		t.Fatalf(`json.Unmarshal(applyResult) error = %v`, err)
	}
	if applyResult.DryRun {
		t.Fatalf(`applyResult.DryRun = true, want false`)
	}
	if got := applyResult.RowsAffected[`settings`]; got != 1 {
		t.Fatalf(`rows_affected[settings] = %d, want 1`, got)
	}
	if strings.TrimSpace(applyResult.PreviewToken) != `` {
		t.Fatalf(`applyResult.PreviewToken = %q, want empty on apply`, applyResult.PreviewToken)
	}
	if !containsHandlerWarning(applyResult.Warnings, `dump contains no channels`) {
		t.Fatalf(`apply warnings = %#v, want standard import warnings`, applyResult.Warnings)
	}
}

func TestImportDBRejectsInvalidDryRunAndModeQueryValues(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://preview-token.example.com`)

	invalidDryRun := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=not-a-bool&mode=incremental`, dump, nil)
	if invalidDryRun.Code != http.StatusBadRequest {
		t.Fatalf(`invalid dry_run status = %d, want %d, body = %s`, invalidDryRun.Code, http.StatusBadRequest, invalidDryRun.Body.String())
	}

	blankDryRun := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=&mode=incremental`, dump, nil)
	if blankDryRun.Code != http.StatusBadRequest {
		t.Fatalf(`blank dry_run status = %d, want %d, body = %s`, blankDryRun.Code, http.StatusBadRequest, blankDryRun.Body.String())
	}

	whitespaceDryRun := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=%20&mode=incremental`, dump, nil)
	if whitespaceDryRun.Code != http.StatusBadRequest {
		t.Fatalf(`whitespace dry_run status = %d, want %d, body = %s`, whitespaceDryRun.Code, http.StatusBadRequest, whitespaceDryRun.Body.String())
	}

	invalidMode := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=surprise`, dump, nil)
	if invalidMode.Code != http.StatusBadRequest {
		t.Fatalf(`invalid mode status = %d, want %d, body = %s`, invalidMode.Code, http.StatusBadRequest, invalidMode.Body.String())
	}

	blankMode := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=`, dump, nil)
	if blankMode.Code != http.StatusBadRequest {
		t.Fatalf(`blank mode status = %d, want %d, body = %s`, blankMode.Code, http.StatusBadRequest, blankMode.Body.String())
	}

	whitespaceMode := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=%20`, dump, nil)
	if whitespaceMode.Code != http.StatusBadRequest {
		t.Fatalf(`whitespace mode status = %d, want %d, body = %s`, whitespaceMode.Code, http.StatusBadRequest, whitespaceMode.Body.String())
	}
}

func TestImportDBPreviewTokenSurvivesUsernameChange(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump("https://preview-token.example.com")

	dryRunRecorder := performImportMultipartHandlerRequest(t, "/api/v1/setting/import?dry_run=true&mode=incremental", dump, nil)
	if dryRunRecorder.Code != http.StatusOK {
		t.Fatalf("dry-run status = %d, want %d, body = %s", dryRunRecorder.Code, http.StatusOK, dryRunRecorder.Body.String())
	}
	dryRunResponse := decodeHandlerResponse(t, dryRunRecorder)
	var dryRunResult model.DBImportResult
	if err := json.Unmarshal(dryRunResponse.Data, &dryRunResult); err != nil {
		t.Fatalf("json.Unmarshal(dryRunResult) error = %v", err)
	}
	if strings.TrimSpace(dryRunResult.PreviewToken) == "" {
		t.Fatalf("preview_token = %q, want non-empty", dryRunResult.PreviewToken)
	}

	changeRecorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/change-username", map[string]any{
		"new_username":     "alice",
		"current_password": "admin",
	}, changeUsername)
	if changeRecorder.Code != http.StatusOK {
		t.Fatalf("change username status = %d, want %d, body = %s", changeRecorder.Code, http.StatusOK, changeRecorder.Body.String())
	}

	applyRecorder := performImportMultipartHandlerRequest(t, "/api/v1/setting/import?dry_run=false&mode=incremental", dump, map[string]string{
		"preview_token": dryRunResult.PreviewToken,
	})
	if applyRecorder.Code != http.StatusOK {
		t.Fatalf("apply status = %d, want %d, body = %s", applyRecorder.Code, http.StatusOK, applyRecorder.Body.String())
	}
}

func TestImportDBPreviewTokenSurvivesPasswordChange(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump("https://preview-token.example.com")

	dryRunRecorder := performImportMultipartHandlerRequest(t, "/api/v1/setting/import?dry_run=true&mode=incremental", dump, nil)
	if dryRunRecorder.Code != http.StatusOK {
		t.Fatalf("dry-run status = %d, want %d, body = %s", dryRunRecorder.Code, http.StatusOK, dryRunRecorder.Body.String())
	}
	dryRunResponse := decodeHandlerResponse(t, dryRunRecorder)
	var dryRunResult model.DBImportResult
	if err := json.Unmarshal(dryRunResponse.Data, &dryRunResult); err != nil {
		t.Fatalf("json.Unmarshal(dryRunResult) error = %v", err)
	}
	if strings.TrimSpace(dryRunResult.PreviewToken) == "" {
		t.Fatalf("preview_token = %q, want non-empty", dryRunResult.PreviewToken)
	}

	changeRecorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/change-password", map[string]any{
		"old_password": "admin",
		"new_password": "new-admin-password",
	}, changePassword)
	if changeRecorder.Code != http.StatusOK {
		t.Fatalf("change password status = %d, want %d, body = %s", changeRecorder.Code, http.StatusOK, changeRecorder.Body.String())
	}

	applyRecorder := performImportMultipartHandlerRequest(t, "/api/v1/setting/import?dry_run=false&mode=incremental", dump, map[string]string{
		"preview_token": dryRunResult.PreviewToken,
	})
	if applyRecorder.Code != http.StatusOK {
		t.Fatalf("apply status = %d, want %d, body = %s", applyRecorder.Code, http.StatusOK, applyRecorder.Body.String())
	}
}

func TestImportDBApplyRejectsMismatchedPreviewToken(t *testing.T) {
	setupHandlerTest(t)
	dryRunDump := testImportPreviewDump(`https://preview-token.example.com`)

	dryRunRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=incremental`, dryRunDump, nil)
	if dryRunRecorder.Code != http.StatusOK {
		t.Fatalf(`dry-run status = %d, want %d, body = %s`, dryRunRecorder.Code, http.StatusOK, dryRunRecorder.Body.String())
	}
	dryRunResponse := decodeHandlerResponse(t, dryRunRecorder)
	var dryRunResult model.DBImportResult
	if err := json.Unmarshal(dryRunResponse.Data, &dryRunResult); err != nil {
		t.Fatalf(`json.Unmarshal(dryRunResult) error = %v`, err)
	}

	modifiedDump := testImportPreviewDump(`https://changed-preview-token.example.com`)
	applyRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=incremental`, modifiedDump, map[string]string{
		`preview_token`: dryRunResult.PreviewToken,
	})
	if applyRecorder.Code != http.StatusBadRequest {
		t.Fatalf(`apply status = %d, want %d, body = %s`, applyRecorder.Code, http.StatusBadRequest, applyRecorder.Body.String())
	}
	applyResponse := decodeHandlerResponse(t, applyRecorder)
	if !strings.Contains(applyResponse.Message, `preview_token does not match current import payload`) {
		t.Fatalf(`message = %q, want preview_token mismatch`, applyResponse.Message)
	}
}

func TestImportDBReplaceApplyRequiresPreviewToken(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://replace-preview-token.example.com`)

	applyRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=replace`, dump, nil)
	if applyRecorder.Code != http.StatusBadRequest {
		t.Fatalf(`apply status = %d, want %d, body = %s`, applyRecorder.Code, http.StatusBadRequest, applyRecorder.Body.String())
	}
	applyResponse := decodeHandlerResponse(t, applyRecorder)
	if !strings.Contains(applyResponse.Message, `preview_token is required`) {
		t.Fatalf(`message = %q, want preview_token required`, applyResponse.Message)
	}
}

func TestImportDBReplaceApplyRejectsPreviewTokenInQuery(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://replace-preview-token-query.example.com`)

	dryRunRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=replace`, dump, nil)
	if dryRunRecorder.Code != http.StatusOK {
		t.Fatalf(`dry-run status = %d, want %d, body = %s`, dryRunRecorder.Code, http.StatusOK, dryRunRecorder.Body.String())
	}
	dryRunResponse := decodeHandlerResponse(t, dryRunRecorder)
	var dryRunResult model.DBImportResult
	if err := json.Unmarshal(dryRunResponse.Data, &dryRunResult); err != nil {
		t.Fatalf(`json.Unmarshal(dryRunResult) error = %v`, err)
	}
	if strings.TrimSpace(dryRunResult.PreviewToken) == `` {
		t.Fatalf(`preview_token = %q, want non-empty`, dryRunResult.PreviewToken)
	}

	applyRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=replace&preview_token=`+dryRunResult.PreviewToken, dump, nil)
	if applyRecorder.Code != http.StatusBadRequest {
		t.Fatalf(`apply status = %d, want %d, body = %s`, applyRecorder.Code, http.StatusBadRequest, applyRecorder.Body.String())
	}
	applyResponse := decodeHandlerResponse(t, applyRecorder)
	if !strings.Contains(applyResponse.Message, `preview_token is required`) {
		t.Fatalf(`message = %q, want preview_token required`, applyResponse.Message)
	}
}

func TestImportDBRejectsBlankPreviewTokenFromHeader(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://blank-preview-token-header.example.com`)

	payload, err := json.Marshal(dump)
	if err != nil {
		t.Fatalf(`json.Marshal(dump) error = %v`, err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, `/api/v1/setting/import?dry_run=false&mode=replace`, bytes.NewReader(payload))
	ctx.Request.Header.Set(`Content-Type`, `application/json`)
	ctx.Request.Header.Set(`X-Octopus-Import-Preview-Token`, `   `)

	importDB(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, `invalid preview_token`) {
		t.Fatalf(`message = %q, want invalid preview_token`, response.Message)
	}
}

func TestImportDBRejectsBlankOptionalFormJSONFields(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://blank-form-fields.example.com`)

	tests := []struct {
		name        string
		fieldName   string
		fieldValue  string
		wantMessage string
	}{
		{name: `blank model_mappings`, fieldName: `model_mappings`, fieldValue: ``, wantMessage: `invalid model_mappings`},
		{name: `whitespace model_mappings`, fieldName: `model_mappings`, fieldValue: `   `, wantMessage: `invalid model_mappings`},
		{name: `blank import_scopes`, fieldName: `import_scopes`, fieldValue: ``, wantMessage: `invalid import_scopes`},
		{name: `whitespace import_scopes`, fieldName: `import_scopes`, fieldValue: `   `, wantMessage: `invalid import_scopes`},
	}

	for _, tt := range tests {
		caseData := tt
		t.Run(caseData.name, func(t *testing.T) {
			recorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=incremental`, dump, map[string]string{
				caseData.fieldName: caseData.fieldValue,
			})
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusBadRequest, recorder.Body.String())
			}
			response := decodeHandlerResponse(t, recorder)
			if !strings.Contains(response.Message, caseData.wantMessage) {
				t.Fatalf(`message = %q, want %q`, response.Message, caseData.wantMessage)
			}
		})
	}
}

func TestImportDBRejectsBlankModelMappingsEntries(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://blank-model-mappings.example.com`)

	tests := []struct {
		name    string
		mapping map[string]string
	}{
		{name: `blank source`, mapping: map[string]string{`   `: `gpt-4o`}},
		{name: `blank target`, mapping: map[string]string{`legacy-model`: `   `}},
	}

	for _, tt := range tests {
		caseData := tt
		t.Run(caseData.name, func(t *testing.T) {
			recorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=map`, dump, map[string]string{
				`model_mappings`: marshalHandlerJSONField(t, caseData.mapping),
			})
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusBadRequest, recorder.Body.String())
			}
			response := decodeHandlerResponse(t, recorder)
			if !strings.Contains(response.Message, `invalid model_mappings`) {
				t.Fatalf(`message = %q, want invalid model_mappings`, response.Message)
			}
		})
	}
}

func TestImportDBReplaceApplyWithSameTokenSucceeds(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://replace-preview-token.example.com`)

	dryRunRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=replace`, dump, nil)
	if dryRunRecorder.Code != http.StatusOK {
		t.Fatalf(`dry-run status = %d, want %d, body = %s`, dryRunRecorder.Code, http.StatusOK, dryRunRecorder.Body.String())
	}
	dryRunResponse := decodeHandlerResponse(t, dryRunRecorder)
	var dryRunResult model.DBImportResult
	if err := json.Unmarshal(dryRunResponse.Data, &dryRunResult); err != nil {
		t.Fatalf(`json.Unmarshal(dryRunResult) error = %v`, err)
	}
	if strings.TrimSpace(dryRunResult.PreviewToken) == `` {
		t.Fatalf(`preview_token = %q, want non-empty`, dryRunResult.PreviewToken)
	}

	applyRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=replace`, dump, map[string]string{
		`preview_token`: dryRunResult.PreviewToken,
	})
	if applyRecorder.Code != http.StatusOK {
		t.Fatalf(`apply status = %d, want %d, body = %s`, applyRecorder.Code, http.StatusOK, applyRecorder.Body.String())
	}
	applyResponse := decodeHandlerResponse(t, applyRecorder)
	var applyResult model.DBImportResult
	if err := json.Unmarshal(applyResponse.Data, &applyResult); err != nil {
		t.Fatalf(`json.Unmarshal(applyResult) error = %v`, err)
	}
	if applyResult.DryRun {
		t.Fatalf(`applyResult.DryRun = true, want false`)
	}
	if got := applyResult.RowsAffected[`settings`]; got != 1 {
		t.Fatalf(`rows_affected[settings] = %d, want 1`, got)
	}
}

func TestImportDBRejectsUnsupportedMediaType(t *testing.T) {
	setupHandlerTest(t)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, `/api/v1/setting/import?dry_run=true&mode=incremental`, strings.NewReader(`{"version":1}`))
	ctx.Request.Header.Set(`Content-Type`, `text/plain; charset=utf-8`)

	importDB(ctx)

	if recorder.Code != http.StatusUnsupportedMediaType {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusUnsupportedMediaType, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, `multipart/form-data or application/json`) {
		t.Fatalf(`message = %q, want content-type guidance`, response.Message)
	}
}

func TestImportDBAcceptsPlusJSONMediaType(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://plus-json.example.com`)
	payload, err := json.Marshal(dump)
	if err != nil {
		t.Fatalf(`json.Marshal(dump) error = %v`, err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, `/api/v1/setting/import?dry_run=true&mode=incremental`, bytes.NewReader(payload))
	ctx.Request.Header.Set(`Content-Type`, `application/problem+json`)

	importDB(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusOK, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	var result model.DBImportResult
	if err := json.Unmarshal(response.Data, &result); err != nil {
		t.Fatalf(`json.Unmarshal(result) error = %v`, err)
	}
	if !result.DryRun {
		t.Fatalf(`result.DryRun = false, want true`)
	}
	if strings.TrimSpace(result.PreviewToken) == `` {
		t.Fatalf(`preview_token = %q, want non-empty`, result.PreviewToken)
	}
}

func TestImportDBApplyRejectsPreviewTokenWhenModeChanges(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://mode-preview-token.example.com`)
	formFields := map[string]string{
		`model_mappings`: marshalHandlerJSONField(t, map[string]string{`legacy-model`: `gpt-4o`}),
		`import_scopes`:  marshalHandlerJSONField(t, model.DBImportScopes{Settings: true}),
	}

	dryRunRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=map`, dump, formFields)
	if dryRunRecorder.Code != http.StatusOK {
		t.Fatalf(`dry-run status = %d, want %d, body = %s`, dryRunRecorder.Code, http.StatusOK, dryRunRecorder.Body.String())
	}
	dryRunResponse := decodeHandlerResponse(t, dryRunRecorder)
	var dryRunResult model.DBImportResult
	if err := json.Unmarshal(dryRunResponse.Data, &dryRunResult); err != nil {
		t.Fatalf(`json.Unmarshal(dryRunResult) error = %v`, err)
	}

	applyFields := cloneHandlerFormFields(formFields)
	applyFields[`preview_token`] = dryRunResult.PreviewToken
	applyRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=merge`, dump, applyFields)
	if applyRecorder.Code != http.StatusBadRequest {
		t.Fatalf(`apply status = %d, want %d, body = %s`, applyRecorder.Code, http.StatusBadRequest, applyRecorder.Body.String())
	}
	applyResponse := decodeHandlerResponse(t, applyRecorder)
	if !strings.Contains(applyResponse.Message, `preview_token does not match current import payload`) {
		t.Fatalf(`message = %q, want preview_token mismatch`, applyResponse.Message)
	}
}

func TestImportDBApplyRejectsPreviewTokenWhenImportScopesChange(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://scope-preview-token.example.com`)
	formFields := map[string]string{
		`import_scopes`: marshalHandlerJSONField(t, model.DBImportScopes{Settings: true}),
	}

	dryRunRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=replace`, dump, formFields)
	if dryRunRecorder.Code != http.StatusOK {
		t.Fatalf(`dry-run status = %d, want %d, body = %s`, dryRunRecorder.Code, http.StatusOK, dryRunRecorder.Body.String())
	}
	dryRunResponse := decodeHandlerResponse(t, dryRunRecorder)
	var dryRunResult model.DBImportResult
	if err := json.Unmarshal(dryRunResponse.Data, &dryRunResult); err != nil {
		t.Fatalf(`json.Unmarshal(dryRunResult) error = %v`, err)
	}

	applyFields := cloneHandlerFormFields(formFields)
	applyFields[`preview_token`] = dryRunResult.PreviewToken
	applyFields[`import_scopes`] = marshalHandlerJSONField(t, model.DBImportScopes{Settings: true, Models: true})
	applyRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=replace`, dump, applyFields)
	if applyRecorder.Code != http.StatusBadRequest {
		t.Fatalf(`apply status = %d, want %d, body = %s`, applyRecorder.Code, http.StatusBadRequest, applyRecorder.Body.String())
	}
	applyResponse := decodeHandlerResponse(t, applyRecorder)
	if !strings.Contains(applyResponse.Message, `preview_token does not match current import payload`) {
		t.Fatalf(`message = %q, want preview_token mismatch`, applyResponse.Message)
	}
}

func TestImportDBApplyRejectsPreviewTokenWhenModelMappingsChange(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://mapping-preview-token.example.com`)
	formFields := map[string]string{
		`model_mappings`: marshalHandlerJSONField(t, map[string]string{`legacy-model`: `gpt-4o`}),
		`import_scopes`:  marshalHandlerJSONField(t, model.DBImportScopes{Settings: true}),
	}

	dryRunRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=map`, dump, formFields)
	if dryRunRecorder.Code != http.StatusOK {
		t.Fatalf(`dry-run status = %d, want %d, body = %s`, dryRunRecorder.Code, http.StatusOK, dryRunRecorder.Body.String())
	}
	dryRunResponse := decodeHandlerResponse(t, dryRunRecorder)
	var dryRunResult model.DBImportResult
	if err := json.Unmarshal(dryRunResponse.Data, &dryRunResult); err != nil {
		t.Fatalf(`json.Unmarshal(dryRunResult) error = %v`, err)
	}

	applyFields := cloneHandlerFormFields(formFields)
	applyFields[`preview_token`] = dryRunResult.PreviewToken
	applyFields[`model_mappings`] = marshalHandlerJSONField(t, map[string]string{`legacy-model`: `gpt-4.1`})
	applyRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=map`, dump, applyFields)
	if applyRecorder.Code != http.StatusBadRequest {
		t.Fatalf(`apply status = %d, want %d, body = %s`, applyRecorder.Code, http.StatusBadRequest, applyRecorder.Body.String())
	}
	applyResponse := decodeHandlerResponse(t, applyRecorder)
	if !strings.Contains(applyResponse.Message, `preview_token does not match current import payload`) {
		t.Fatalf(`message = %q, want preview_token mismatch`, applyResponse.Message)
	}
}

func TestImportDBApplyWithSameTokenAndOptionsInMapModeSucceeds(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://map-preview-token.example.com`)
	formFields := map[string]string{
		`model_mappings`: marshalHandlerJSONField(t, map[string]string{`legacy-model`: `gpt-4o`}),
		`import_scopes`:  marshalHandlerJSONField(t, model.DBImportScopes{Settings: true}),
	}

	dryRunRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=map`, dump, formFields)
	if dryRunRecorder.Code != http.StatusOK {
		t.Fatalf(`dry-run status = %d, want %d, body = %s`, dryRunRecorder.Code, http.StatusOK, dryRunRecorder.Body.String())
	}
	dryRunResponse := decodeHandlerResponse(t, dryRunRecorder)
	var dryRunResult model.DBImportResult
	if err := json.Unmarshal(dryRunResponse.Data, &dryRunResult); err != nil {
		t.Fatalf(`json.Unmarshal(dryRunResult) error = %v`, err)
	}
	if strings.TrimSpace(dryRunResult.PreviewToken) == `` {
		t.Fatalf(`preview_token = %q, want non-empty`, dryRunResult.PreviewToken)
	}

	applyFields := cloneHandlerFormFields(formFields)
	applyFields[`preview_token`] = dryRunResult.PreviewToken
	applyRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=map`, dump, applyFields)
	if applyRecorder.Code != http.StatusOK {
		t.Fatalf(`apply status = %d, want %d, body = %s`, applyRecorder.Code, http.StatusOK, applyRecorder.Body.String())
	}
	applyResponse := decodeHandlerResponse(t, applyRecorder)
	var applyResult model.DBImportResult
	if err := json.Unmarshal(applyResponse.Data, &applyResult); err != nil {
		t.Fatalf(`json.Unmarshal(applyResult) error = %v`, err)
	}
	if applyResult.DryRun {
		t.Fatalf(`applyResult.DryRun = true, want false`)
	}
	if got := applyResult.RowsAffected[`settings`]; got != 1 {
		t.Fatalf(`rows_affected[settings] = %d, want 1`, got)
	}
	if strings.TrimSpace(applyResult.PreviewToken) != `` {
		t.Fatalf(`applyResult.PreviewToken = %q, want empty on apply`, applyResult.PreviewToken)
	}
}

func TestImportDBDryRunWithJSONBodyAndQueryOptionsProducesMatchingPreviewDigest(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://json-query-preview-token.example.com`)
	modelMappings := marshalHandlerJSONField(t, map[string]string{`legacy-model`: `gpt-4o`})
	importScopes := marshalHandlerJSONField(t, model.DBImportScopes{Settings: true})

	dryRunTarget := `/api/v1/setting/import?dry_run=true&mode=map&model_mappings=` + url.QueryEscape(modelMappings) + `&import_scopes=` + url.QueryEscape(importScopes)
	dryRunRecorder := performJSONHandlerRequest(t, http.MethodPost, dryRunTarget, dump, importDB)
	if dryRunRecorder.Code != http.StatusOK {
		t.Fatalf(`dry-run status = %d, want %d, body = %s`, dryRunRecorder.Code, http.StatusOK, dryRunRecorder.Body.String())
	}
	dryRunResponse := decodeHandlerResponse(t, dryRunRecorder)
	var dryRunResult model.DBImportResult
	if err := json.Unmarshal(dryRunResponse.Data, &dryRunResult); err != nil {
		t.Fatalf(`json.Unmarshal(dryRunResult) error = %v`, err)
	}
	if strings.TrimSpace(dryRunResult.PreviewToken) == `` {
		t.Fatalf(`preview_token = %q, want non-empty`, dryRunResult.PreviewToken)
	}
	expectedDigest, err := buildImportPreviewDigest(dump, model.DBImportModeMap, model.DBImportOptions{
		ModelMappings: map[string]string{`legacy-model`: `gpt-4o`},
		ImportScopes:  &model.DBImportScopes{Settings: true},
	})
	if err != nil {
		t.Fatalf(`buildImportPreviewDigest() error = %v`, err)
	}
	if err := verifyImportPreviewToken(dryRunResult.PreviewToken, expectedDigest); err != nil {
		t.Fatalf(`verifyImportPreviewToken() error = %v, want nil`, err)
	}
	if stored, err := op.SettingGetString(model.SettingKeyAPIBaseURL); err != nil {
		t.Fatalf(`SettingGetString(api_base_url) error = %v`, err)
	} else if stored != `http://localhost:1088` {
		t.Fatalf(`api_base_url = %q, want unchanged after dry-run`, stored)
	}
}

func TestImportDBReplaceApplyAcceptsPreviewTokenFromJSONQuery(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://json-query-replace.example.com`)

	dryRunRecorder := performJSONHandlerRequest(t, http.MethodPost, `/api/v1/setting/import?dry_run=true&mode=replace`, dump, importDB)
	if dryRunRecorder.Code != http.StatusOK {
		t.Fatalf(`dry-run status = %d, want %d, body = %s`, dryRunRecorder.Code, http.StatusOK, dryRunRecorder.Body.String())
	}
	dryRunResponse := decodeHandlerResponse(t, dryRunRecorder)
	var dryRunResult model.DBImportResult
	if err := json.Unmarshal(dryRunResponse.Data, &dryRunResult); err != nil {
		t.Fatalf(`json.Unmarshal(dryRunResult) error = %v`, err)
	}
	if strings.TrimSpace(dryRunResult.PreviewToken) == `` {
		t.Fatalf(`preview_token = %q, want non-empty`, dryRunResult.PreviewToken)
	}

	applyTarget := `/api/v1/setting/import?dry_run=false&mode=replace&preview_token=` + url.QueryEscape(dryRunResult.PreviewToken)
	applyRecorder := performJSONHandlerRequest(t, http.MethodPost, applyTarget, dump, importDB)
	if applyRecorder.Code != http.StatusOK {
		t.Fatalf(`apply status = %d, want %d, body = %s`, applyRecorder.Code, http.StatusOK, applyRecorder.Body.String())
	}
	applyResponse := decodeHandlerResponse(t, applyRecorder)
	var applyResult model.DBImportResult
	if err := json.Unmarshal(applyResponse.Data, &applyResult); err != nil {
		t.Fatalf(`json.Unmarshal(applyResult) error = %v`, err)
	}
	if applyResult.DryRun {
		t.Fatalf(`applyResult.DryRun = true, want false`)
	}
	if got := applyResult.RowsAffected[`settings`]; got != 1 {
		t.Fatalf(`rows_affected[settings] = %d, want 1`, got)
	}
	if strings.TrimSpace(applyResult.PreviewToken) != `` {
		t.Fatalf(`applyResult.PreviewToken = %q, want empty on apply`, applyResult.PreviewToken)
	}
	stored, err := op.SettingGetString(model.SettingKeyAPIBaseURL)
	if err != nil {
		t.Fatalf(`SettingGetString(api_base_url) error = %v`, err)
	}
	if stored != `https://json-query-replace.example.com` {
		t.Fatalf(`api_base_url = %q, want imported JSON body value`, stored)
	}
}

func TestImportPreviewDigestStableForSameMapModeOptions(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://digest-stable.example.com`)
	options := model.DBImportOptions{
		ModelMappings: map[string]string{`legacy-model`: `gpt-4o`},
		ImportScopes:  &model.DBImportScopes{Settings: true},
	}

	first, err := buildImportPreviewDigest(dump, model.DBImportModeMap, options)
	if err != nil {
		t.Fatalf(`buildImportPreviewDigest(first) error = %v`, err)
	}
	second, err := buildImportPreviewDigest(dump, model.DBImportModeMap, options)
	if err != nil {
		t.Fatalf(`buildImportPreviewDigest(second) error = %v`, err)
	}
	if first != second {
		t.Fatalf(`digest mismatch for same map options: %q != %q`, first, second)
	}
}

func TestImportDBRejectsOversizedMultipartPayload(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://oversized-multipart.example.com`)

	originalLimit := maxDBImportPayloadBytes
	maxDBImportPayloadBytes = 64
	t.Cleanup(func() {
		maxDBImportPayloadBytes = originalLimit
	})

	recorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=incremental`, dump, nil)
	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusRequestEntityTooLarge, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, `import payload too large`) {
		t.Fatalf(`message = %q, want import payload too large`, response.Message)
	}
}

func TestImportDBRejectsOversizedMultipartFilePart(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://oversized-file-part.example.com`)
	payload, err := json.Marshal(dump)
	if err != nil {
		t.Fatalf(`json.Marshal(dump) error = %v`, err)
	}

	originalPayloadLimit := maxDBImportPayloadBytes
	originalFileLimit := maxDBImportFileBytes
	maxDBImportPayloadBytes = 1 << 20
	maxDBImportFileBytes = int64(len(payload) - 1)
	t.Cleanup(func() {
		maxDBImportPayloadBytes = originalPayloadLimit
		maxDBImportFileBytes = originalFileLimit
	})

	recorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=incremental`, dump, nil)
	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusRequestEntityTooLarge, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, `import payload too large`) {
		t.Fatalf(`message = %q, want import payload too large`, response.Message)
	}
}

func TestImportDBRejectsOversizedJSONPayload(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://oversized-json.example.com`)
	payload, err := json.Marshal(dump)
	if err != nil {
		t.Fatalf(`json.Marshal(dump) error = %v`, err)
	}

	originalLimit := maxDBImportPayloadBytes
	maxDBImportPayloadBytes = int64(len(payload) - 1)
	t.Cleanup(func() {
		maxDBImportPayloadBytes = originalLimit
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, `/api/v1/setting/import?dry_run=true&mode=incremental`, bytes.NewReader(payload))
	ctx.Request.Header.Set(`Content-Type`, `application/json`)
	importDB(ctx)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusRequestEntityTooLarge, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, `import payload too large`) {
		t.Fatalf(`message = %q, want import payload too large`, response.Message)
	}
}

func TestPreviewRollbackImportSnapshotReturnsStructuredReport(t *testing.T) {
	setupHandlerTest(t)

	currentChannel := model.Channel{
		Name:     `preview-rollback-channel`,
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: `https://current-preview.example.com/v1`, Delay: 0}},
		Model:    `gpt-4o`,
	}
	if err := db.GetDB().Create(&currentChannel).Error; err != nil {
		t.Fatalf(`create current channel error = %v`, err)
	}
	currentGroup := model.Group{Name: `preview-rollback-group`, Mode: model.GroupModeRoundRobin}
	if err := db.GetDB().Create(&currentGroup).Error; err != nil {
		t.Fatalf(`create current group error = %v`, err)
	}
	if err := db.GetDB().Create(&model.GroupItem{GroupID: currentGroup.ID, ChannelID: currentChannel.ID, ModelName: `gpt-4o`, Priority: 1, Weight: 1}).Error; err != nil {
		t.Fatalf(`create current group item error = %v`, err)
	}
	if err := db.GetDB().Create(&model.LLMInfo{Name: `gpt-4o`, CanonicalName: `gpt-4o`}).Error; err != nil {
		t.Fatalf(`create current llm info error = %v`, err)
	}
	if err := db.GetDB().Create(&model.StatsTotal{ID: 1, StatsMetrics: model.StatsMetrics{InputToken: 42}}).Error; err != nil {
		t.Fatalf(`create stats_total error = %v`, err)
	}
	if err := db.GetDB().Create(&model.RelayLog{ID: 9201, Time: 1710000010, RequestModelName: `gpt-4o`, ChannelId: currentChannel.ID, ChannelName: currentChannel.Name, ActualModelName: `gpt-4o`, TotalAttempts: 1}).Error; err != nil {
		t.Fatalf(`create relay log error = %v`, err)
	}
	if err := db.GetDB().Create(&model.Channel{Name: `preview-before-import`, Enabled: true, Model: `gpt-4o`}).Error; err != nil {
		t.Fatalf(`create pre-import channel error = %v`, err)
	}

	importDump := &model.DBDump{
		Version: 1,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   `v1`,
			ExportSource:    `octopus`,
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{ID: 3501, Name: `preview-imported-channel`, Enabled: true, Model: `gpt-4o`}},
	}
	importRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=incremental`, importDump, nil)
	if importRecorder.Code != http.StatusOK {
		t.Fatalf(`import status = %d, want %d, body = %s`, importRecorder.Code, http.StatusOK, importRecorder.Body.String())
	}

	if err := db.GetDB().Delete(&model.Channel{}, `name = ?`, `preview-before-import`).Error; err != nil {
		t.Fatalf(`delete pre-import channel error = %v`, err)
	}
	if err := db.GetDB().Delete(&model.LLMInfo{}, `name = ?`, `gpt-4o`).Error; err != nil {
		t.Fatalf(`delete current llm info error = %v`, err)
	}
	currentChannel.BaseUrls = []model.BaseUrl{{URL: `https://current-mutated.example.com/v1`, Delay: 0}}
	currentChannel.Model = `gpt-4o,o1-mini`
	if err := db.GetDB().Save(&currentChannel).Error; err != nil {
		t.Fatalf(`mutate current channel error = %v`, err)
	}
	if err := db.GetDB().Create(&model.GroupItem{GroupID: currentGroup.ID, ChannelID: currentChannel.ID, ModelName: `o1-mini`, Priority: 2, Weight: 1}).Error; err != nil {
		t.Fatalf(`create mutated group item error = %v`, err)
	}

	listRecorder := performJSONHandlerRequest(t, http.MethodGet, `/api/v1/setting/import-snapshots`, nil, listImportSnapshots)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf(`list status = %d, want %d, body = %s`, listRecorder.Code, http.StatusOK, listRecorder.Body.String())
	}
	listResponse := decodeHandlerResponse(t, listRecorder)
	var snapshots []model.DBImportSnapshotInfo
	if err := json.Unmarshal(listResponse.Data, &snapshots); err != nil {
		t.Fatalf(`json.Unmarshal(snapshots) error = %v`, err)
	}
	if len(snapshots) == 0 {
		t.Fatalf(`snapshots = %#v, want at least one snapshot`, snapshots)
	}
	for _, snapshot := range snapshots {
		if strings.Contains(snapshot.SnapshotPath, `:\\`) || strings.HasPrefix(snapshot.SnapshotPath, `/`) {
			t.Fatalf(`snapshot_path = %q, want display-safe relative value`, snapshot.SnapshotPath)
		}
		if want := `import-snapshots/` + snapshot.SnapshotName; snapshot.SnapshotPath != want {
			t.Fatalf(`snapshot_path = %q, want %q`, snapshot.SnapshotPath, want)
		}
	}

	previewRecorder := performJSONHandlerRequest(t, http.MethodPost, `/api/v1/setting/preview-rollback-import-snapshot`, map[string]string{
		`snapshot_name`: snapshots[0].SnapshotName,
	}, previewRollbackImportSnapshot)
	if previewRecorder.Code != http.StatusOK {
		t.Fatalf(`preview status = %d, want %d, body = %s`, previewRecorder.Code, http.StatusOK, previewRecorder.Body.String())
	}
	previewResponse := decodeHandlerResponse(t, previewRecorder)
	var preview model.DBRollbackPreviewResult
	if err := json.Unmarshal(previewResponse.Data, &preview); err != nil {
		t.Fatalf(`json.Unmarshal(preview) error = %v`, err)
	}
	if preview.SnapshotName != snapshots[0].SnapshotName {
		t.Fatalf(`snapshot_name = %q, want %q`, preview.SnapshotName, snapshots[0].SnapshotName)
	}
	if preview.SnapshotPath != snapshots[0].SnapshotPath {
		t.Fatalf(`snapshot_path = %q, want %q`, preview.SnapshotPath, snapshots[0].SnapshotPath)
	}
	if preview.ImportedAt.IsZero() {
		t.Fatalf(`imported_at = %v, want non-zero`, preview.ImportedAt)
	}
	if !snapshots[0].ImportedAt.IsZero() && !preview.ImportedAt.Equal(snapshots[0].ImportedAt) {
		t.Fatalf(`imported_at = %v, want %v`, preview.ImportedAt, snapshots[0].ImportedAt)
	}
	if preview.Manifest == nil || preview.Manifest.SchemaVersion != `v1` {
		t.Fatalf(`manifest = %#v, want schema v1`, preview.Manifest)
	}
	if preview.RowsSummary == nil || preview.RowsSummary[`channels`] == 0 {
		t.Fatalf(`rows_summary = %#v, want populated channels summary`, preview.RowsSummary)
	}
	if preview.Compatibility == nil || preview.Compatibility.Summary == nil {
		t.Fatalf(`compatibility = %#v, want populated summary`, preview.Compatibility)
	}
	if got := preview.Compatibility.Summary.RoutePreviewDiffs; got == 0 {
		t.Fatalf(`compatibility.summary.route_preview_diffs = %d, want > 0`, got)
	}
	if !containsHandlerWarning(preview.Compatibility.BaseURLMismatches, `preview-rollback-channel`) {
		t.Fatalf(`base_url_mismatches = %#v, want preview-rollback-channel`, preview.Compatibility.BaseURLMismatches)
	}
	if !containsHandlerWarning(preview.Compatibility.MissingProviders, `preview-before-import`) {
		t.Fatalf(`missing_providers = %#v, want preview-before-import`, preview.Compatibility.MissingProviders)
	}
	if !containsHandlerWarning(preview.Compatibility.MissingModels, `gpt-4o`) {
		t.Fatalf(`missing_models = %#v, want gpt-4o`, preview.Compatibility.MissingModels)
	}
	if !containsHandlerWarning(preview.Compatibility.ReplacePrunedChannels, `preview-imported-channel`) {
		t.Fatalf(`replace_pruned_channels = %#v, want preview-imported-channel`, preview.Compatibility.ReplacePrunedChannels)
	}
	if !containsHandlerWarning(preview.PreviewWarnings, `route preview diffs`) {
		t.Fatalf(`preview_warnings = %#v, want route preview diffs warning`, preview.PreviewWarnings)
	}
	if !containsHandlerWarning(preview.PreviewWarnings, `base URL mismatches`) {
		t.Fatalf(`preview_warnings = %#v, want base URL mismatches warning`, preview.PreviewWarnings)
	}
	if !containsHandlerWarning(preview.PreviewWarnings, `includes stats tables`) {
		t.Fatalf(`preview_warnings = %#v, want includes stats tables warning`, preview.PreviewWarnings)
	}
	if !containsHandlerWarning(preview.PreviewWarnings, `includes relay logs`) {
		t.Fatalf(`preview_warnings = %#v, want includes relay logs warning`, preview.PreviewWarnings)
	}
}

func TestPreviewRollbackImportSnapshotHonorsImportScopes(t *testing.T) {
	setupHandlerTest(t)

	channel := model.Channel{Name: `scope-preview-channel`, Enabled: true, Model: `gpt-4o`}
	if err := db.GetDB().Create(&channel).Error; err != nil {
		t.Fatalf(`create channel error = %v`, err)
	}
	if err := db.GetDB().Model(&model.Setting{}).Where(`key = ?`, model.SettingKeyAPIBaseURL).Update(`value`, `https://before.example.com`).Error; err != nil {
		t.Fatalf(`update setting error = %v`, err)
	}

	importDump := &model.DBDump{
		Version: 1,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   `v1`,
			ExportSource:    `octopus`,
			ContainsSecrets: true,
		},
		Settings: []model.Setting{{Key: model.SettingKeyAPIBaseURL, Value: `https://snapshot.example.com`}},
	}
	importRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=incremental`, importDump, nil)
	if importRecorder.Code != http.StatusOK {
		t.Fatalf(`import status = %d, want %d, body = %s`, importRecorder.Code, http.StatusOK, importRecorder.Body.String())
	}

	listRecorder := performJSONHandlerRequest(t, http.MethodGet, `/api/v1/setting/import-snapshots`, nil, listImportSnapshots)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf(`list status = %d, want %d, body = %s`, listRecorder.Code, http.StatusOK, listRecorder.Body.String())
	}
	listResponse := decodeHandlerResponse(t, listRecorder)
	var snapshots []model.DBImportSnapshotInfo
	if err := json.Unmarshal(listResponse.Data, &snapshots); err != nil {
		t.Fatalf(`json.Unmarshal(snapshots) error = %v`, err)
	}
	if len(snapshots) == 0 {
		t.Fatalf(`snapshots = %#v, want at least one snapshot`, snapshots)
	}
	if snapshots[0].SnapshotPath != `import-snapshots/`+snapshots[0].SnapshotName {
		t.Fatalf(`snapshot_path = %q, want display-safe relative path`, snapshots[0].SnapshotPath)
	}

	previewRecorder := performJSONHandlerRequest(t, http.MethodPost, `/api/v1/setting/preview-rollback-import-snapshot`, map[string]any{
		`snapshot_name`: snapshots[0].SnapshotName,
		`import_scopes`: model.DBImportScopes{Settings: true},
	}, previewRollbackImportSnapshot)
	if previewRecorder.Code != http.StatusOK {
		t.Fatalf(`preview status = %d, want %d, body = %s`, previewRecorder.Code, http.StatusOK, previewRecorder.Body.String())
	}
	previewResponse := decodeHandlerResponse(t, previewRecorder)
	var preview model.DBRollbackPreviewResult
	if err := json.Unmarshal(previewResponse.Data, &preview); err != nil {
		t.Fatalf(`json.Unmarshal(preview) error = %v`, err)
	}
	if preview.AppliedScopes == nil || !preview.AppliedScopes.Settings || preview.AppliedScopes.Routing {
		t.Fatalf(`applied_scopes = %#v, want settings-only scope`, preview.AppliedScopes)
	}
	if preview.RowsSummary[`settings`] == 0 {
		t.Fatalf(`rows_summary = %#v, want settings summary`, preview.RowsSummary)
	}
	if preview.RowsSummary[`channels`] != 0 {
		t.Fatalf(`rows_summary = %#v, want channels summary removed by settings-only scope`, preview.RowsSummary)
	}
}

func TestPreviewRollbackImportSnapshotRejectsEmptyImportScopesObject(t *testing.T) {
	setupHandlerTest(t)

	snapshotDir := filepath.Join(filepath.Dir(filepath.Clean(db.GetCurrentDSN())), `import-snapshots`)
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		t.Fatalf(`MkdirAll(snapshotDir) error = %v`, err)
	}
	snapshotName := `empty-scope-preview.json`
	snapshotPath := filepath.Join(snapshotDir, snapshotName)
	dump := &model.DBDump{
		Version: 1,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   `v1`,
			ExportSource:    `octopus`,
			ContainsSecrets: false,
		},
		Settings: []model.Setting{{Key: model.SettingKeyAPIBaseURL, Value: `https://empty-scope.example.com`}},
	}
	payload, err := json.MarshalIndent(dump, ``, `  `)
	if err != nil {
		t.Fatalf(`json.MarshalIndent(dump) error = %v`, err)
	}
	if err := os.WriteFile(snapshotPath, payload, 0o644); err != nil {
		t.Fatalf(`WriteFile(snapshotPath) error = %v`, err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, `/api/v1/setting/preview-rollback-import-snapshot`, map[string]any{
		`snapshot_name`: snapshotName,
		`import_scopes`: model.DBImportScopes{},
	}, previewRollbackImportSnapshot)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, `at least one import scope must be enabled`) {
		t.Fatalf(`message = %q, want at least one import scope must be enabled`, response.Message)
	}
}

func TestImportDBRejectsEmptyImportScopesObject(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://empty-import-scope.example.com`)

	recorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=incremental`, dump, map[string]string{
		`import_scopes`: marshalHandlerJSONField(t, model.DBImportScopes{}),
	})
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, `at least one import scope must be enabled`) {
		t.Fatalf(`message = %q, want at least one import scope must be enabled`, response.Message)
	}
}

func TestImportDBDryRunWithoutImportScopesStillDefaultsToFullScope(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://empty-import-scope.example.com`)

	recorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=incremental`, dump, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusOK, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	var result model.DBImportResult
	if err := json.Unmarshal(response.Data, &result); err != nil {
		t.Fatalf(`json.Unmarshal(result) error = %v`, err)
	}
	if result.Compatibility == nil || result.Compatibility.Summary == nil {
		t.Fatalf(`compatibility = %#v, want populated summary`, result.Compatibility)
	}
	if result.PreviewToken == `` {
		t.Fatal(`preview_token = empty, want dry-run preview token`)
	}
}

func TestImportDBReturnsErrorWhenCacheRefreshFailsAfterApply(t *testing.T) {
	setupHandlerTest(t)
	dump := testImportPreviewDump(`https://cache-refresh-fail.example.com`)

	originalRefresh := refreshCachesAfterMutableSettingOperation
	refreshCachesAfterMutableSettingOperation = func() error {
		return errors.New(`cache refresh boom`)
	}
	t.Cleanup(func() {
		refreshCachesAfterMutableSettingOperation = originalRefresh
	})

	recorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=incremental`, dump, nil)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusInternalServerError, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, `import succeeded but cache refresh failed`) {
		t.Fatalf(`message = %q, want import cache refresh failure`, response.Message)
	}

	stored, err := op.SettingGetString(model.SettingKeyAPIBaseURL)
	if err != nil {
		t.Fatalf(`SettingGetString(api_base_url) error = %v`, err)
	}
	if stored != `https://cache-refresh-fail.example.com` {
		t.Fatalf(`api_base_url = %q, want persisted import change`, stored)
	}
}

func TestImportDBSyncsMutableSettingTasksAfterApply(t *testing.T) {
	setupHandlerTest(t)

	dump := &model.DBDump{
		Version: 1,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   `v1`,
			ExportSource:    `octopus`,
			ContainsSecrets: true,
		},
		Settings: []model.Setting{
			{Key: model.SettingKeyStatsSaveInterval, Value: `17`},
			{Key: model.SettingKeyModelInfoUpdateInterval, Value: `31`},
			{Key: model.SettingKeySyncLLMInterval, Value: `29`},
		},
	}

	called := 0
	originalSync := syncMutableSettingTasksAfterCacheRefresh
	syncMutableSettingTasksAfterCacheRefresh = func() error {
		called++
		assertStoredSettingValue(t, model.SettingKeyStatsSaveInterval, `17`)
		assertStoredSettingValue(t, model.SettingKeyModelInfoUpdateInterval, `31`)
		assertStoredSettingValue(t, model.SettingKeySyncLLMInterval, `29`)
		return nil
	}
	t.Cleanup(func() {
		syncMutableSettingTasksAfterCacheRefresh = originalSync
	})

	recorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=incremental`, dump, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if called != 1 {
		t.Fatalf(`task sync hook called %d times, want 1`, called)
	}
}

func TestRollbackImportSnapshotRejectsEmptyImportScopesObject(t *testing.T) {
	setupHandlerTest(t)

	snapshotDir := filepath.Join(filepath.Dir(filepath.Clean(db.GetCurrentDSN())), `import-snapshots`)
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		t.Fatalf(`MkdirAll(snapshotDir) error = %v`, err)
	}
	snapshotName := `empty-rollback-scope.json`
	snapshotPath := filepath.Join(snapshotDir, snapshotName)
	dump := &model.DBDump{
		Version: 1,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   `v1`,
			ExportSource:    `octopus`,
			ContainsSecrets: false,
		},
		Settings: []model.Setting{{Key: model.SettingKeyAPIBaseURL, Value: `https://empty-rollback.example.com`}},
	}
	payload, err := json.MarshalIndent(dump, ``, `  `)
	if err != nil {
		t.Fatalf(`json.MarshalIndent(dump) error = %v`, err)
	}
	if err := os.WriteFile(snapshotPath, payload, 0o644); err != nil {
		t.Fatalf(`WriteFile(snapshotPath) error = %v`, err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, `/api/v1/setting/rollback-import-snapshot`, map[string]any{
		`snapshot_name`: snapshotName,
		`import_scopes`: model.DBImportScopes{},
	}, rollbackImportSnapshot)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, `at least one import scope must be enabled`) {
		t.Fatalf(`message = %q, want at least one import scope must be enabled`, response.Message)
	}
}

func TestRollbackImportSnapshotWithoutImportScopesStillDefaultsToFullScope(t *testing.T) {
	setupHandlerTest(t)

	snapshotDir := filepath.Join(filepath.Dir(filepath.Clean(db.GetCurrentDSN())), `import-snapshots`)
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		t.Fatalf(`MkdirAll(snapshotDir) error = %v`, err)
	}
	snapshotName := `empty-rollback-scope-full.json`
	snapshotPath := filepath.Join(snapshotDir, snapshotName)
	dump := &model.DBDump{
		Version: 1,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   `v1`,
			ExportSource:    `octopus`,
			ContainsSecrets: false,
		},
		Settings: []model.Setting{{Key: model.SettingKeyAPIBaseURL, Value: `https://empty-rollback.example.com`}},
	}
	payload, err := json.MarshalIndent(dump, ``, `  `)
	if err != nil {
		t.Fatalf(`json.MarshalIndent(dump) error = %v`, err)
	}
	if err := os.WriteFile(snapshotPath, payload, 0o644); err != nil {
		t.Fatalf(`WriteFile(snapshotPath) error = %v`, err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, `/api/v1/setting/rollback-import-snapshot`, map[string]any{
		`snapshot_name`: snapshotName,
	}, rollbackImportSnapshot)
	if recorder.Code != http.StatusOK {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusOK, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	var result map[string]any
	if err := json.Unmarshal(response.Data, &result); err != nil {
		t.Fatalf(`json.Unmarshal(result) error = %v`, err)
	}
	if _, ok := result[`applied_scopes`]; ok {
		t.Fatalf(`applied_scopes = %#v, want omitted full-scope default`, result[`applied_scopes`])
	}
}

func TestRollbackImportSnapshotReturnsErrorWhenCacheRefreshFails(t *testing.T) {
	setupHandlerTest(t)

	importDump := &model.DBDump{
		Version: 1,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   `v1`,
			ExportSource:    `octopus`,
			ContainsSecrets: true,
		},
		Settings: []model.Setting{{Key: model.SettingKeyAPIBaseURL, Value: `https://snapshot-before-refresh-fail.example.com`}},
	}
	importRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=incremental`, importDump, nil)
	if importRecorder.Code != http.StatusOK {
		t.Fatalf(`import status = %d, want %d, body = %s`, importRecorder.Code, http.StatusOK, importRecorder.Body.String())
	}
	if err := db.GetDB().Model(&model.Setting{}).Where(`key = ?`, model.SettingKeyAPIBaseURL).Update(`value`, `https://mutated-after-import.example.com`).Error; err != nil {
		t.Fatalf(`mutate setting error = %v`, err)
	}

	listRecorder := performJSONHandlerRequest(t, http.MethodGet, `/api/v1/setting/import-snapshots`, nil, listImportSnapshots)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf(`list status = %d, want %d, body = %s`, listRecorder.Code, http.StatusOK, listRecorder.Body.String())
	}
	listResponse := decodeHandlerResponse(t, listRecorder)
	var snapshots []model.DBImportSnapshotInfo
	if err := json.Unmarshal(listResponse.Data, &snapshots); err != nil {
		t.Fatalf(`json.Unmarshal(snapshots) error = %v`, err)
	}
	if len(snapshots) == 0 {
		t.Fatalf(`snapshots = %#v, want at least one snapshot`, snapshots)
	}

	originalRefresh := refreshCachesAfterMutableSettingOperation
	refreshCachesAfterMutableSettingOperation = func() error {
		return errors.New(`cache refresh boom`)
	}
	t.Cleanup(func() {
		refreshCachesAfterMutableSettingOperation = originalRefresh
	})

	recorder := performJSONHandlerRequest(t, http.MethodPost, `/api/v1/setting/rollback-import-snapshot`, map[string]any{
		`snapshot_name`: snapshots[0].SnapshotName,
	}, rollbackImportSnapshot)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusInternalServerError, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, `rollback import snapshot succeeded but cache refresh failed`) {
		t.Fatalf(`message = %q, want rollback cache refresh failure`, response.Message)
	}

	stored, err := op.SettingGetString(model.SettingKeyAPIBaseURL)
	if err != nil {
		t.Fatalf(`SettingGetString(api_base_url) error = %v`, err)
	}
	if stored != `http://localhost:1088` {
		t.Fatalf(`api_base_url = %q, want rollback-applied pre-import value`, stored)
	}
}

func TestRollbackImportSnapshotSyncsMutableSettingTasksAfterApply(t *testing.T) {
	setupHandlerTest(t)

	seedDump := &model.DBDump{
		Version: 1,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   `v1`,
			ExportSource:    `octopus`,
			ContainsSecrets: true,
		},
		Settings: []model.Setting{
			{Key: model.SettingKeyStatsSaveInterval, Value: `17`},
			{Key: model.SettingKeyModelInfoUpdateInterval, Value: `31`},
			{Key: model.SettingKeySyncLLMInterval, Value: `29`},
		},
	}
	seedRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=incremental`, seedDump, nil)
	if seedRecorder.Code != http.StatusOK {
		t.Fatalf(`seed import status = %d, want %d, body = %s`, seedRecorder.Code, http.StatusOK, seedRecorder.Body.String())
	}

	rollbackDump := &model.DBDump{
		Version: 1,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   `v1`,
			ExportSource:    `octopus`,
			ContainsSecrets: true,
		},
		Settings: []model.Setting{
			{Key: model.SettingKeyStatsSaveInterval, Value: `7`},
			{Key: model.SettingKeyModelInfoUpdateInterval, Value: `13`},
			{Key: model.SettingKeySyncLLMInterval, Value: `11`},
		},
	}
	rollbackImportRecorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=false&mode=incremental`, rollbackDump, nil)
	if rollbackImportRecorder.Code != http.StatusOK {
		t.Fatalf(`rollback import status = %d, want %d, body = %s`, rollbackImportRecorder.Code, http.StatusOK, rollbackImportRecorder.Body.String())
	}

	called := 0
	originalSync := syncMutableSettingTasksAfterCacheRefresh
	syncMutableSettingTasksAfterCacheRefresh = func() error {
		called++
		assertStoredSettingValue(t, model.SettingKeyStatsSaveInterval, `17`)
		assertStoredSettingValue(t, model.SettingKeyModelInfoUpdateInterval, `31`)
		assertStoredSettingValue(t, model.SettingKeySyncLLMInterval, `29`)
		return nil
	}
	t.Cleanup(func() {
		syncMutableSettingTasksAfterCacheRefresh = originalSync
	})

	listRecorder := performJSONHandlerRequest(t, http.MethodGet, `/api/v1/setting/import-snapshots`, nil, listImportSnapshots)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf(`list status = %d, want %d, body = %s`, listRecorder.Code, http.StatusOK, listRecorder.Body.String())
	}
	listResponse := decodeHandlerResponse(t, listRecorder)
	var snapshots []model.DBImportSnapshotInfo
	if err := json.Unmarshal(listResponse.Data, &snapshots); err != nil {
		t.Fatalf(`json.Unmarshal(snapshots) error = %v`, err)
	}
	if len(snapshots) == 0 {
		t.Fatalf(`snapshots = %#v, want at least one snapshot`, snapshots)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, `/api/v1/setting/rollback-import-snapshot`, map[string]any{
		`snapshot_name`: snapshots[0].SnapshotName,
	}, rollbackImportSnapshot)
	if recorder.Code != http.StatusOK {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if called != 1 {
		t.Fatalf(`task sync hook called %d times, want 1`, called)
	}
}

func TestPreviewRollbackImportSnapshotRejectsInvalidSnapshotName(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, `/api/v1/setting/preview-rollback-import-snapshot`, map[string]string{
		`snapshot_name`: `..\\outside.json`,
	}, previewRollbackImportSnapshot)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, `invalid snapshot_name`) {
		t.Fatalf(`message = %q, want invalid snapshot_name`, response.Message)
	}
}

func TestPreviewRollbackImportSnapshotRequiresSnapshotName(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, `/api/v1/setting/preview-rollback-import-snapshot`, map[string]string{
		`snapshot_name`: ``,
	}, previewRollbackImportSnapshot)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, `snapshot_name is required`) {
		t.Fatalf(`message = %q, want snapshot_name is required`, response.Message)
	}
}

func TestSanitizeSnapshotDisplayPathDropsDirectorySegments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantPath string
	}{
		{name: `unix style`, input: `import-snapshots/nested/snapshot-a.json`, wantPath: `import-snapshots/snapshot-a.json`},
		{name: `windows style`, input: `import-snapshots\\nested\\snapshot-b.json`, wantPath: `import-snapshots/snapshot-b.json`},
		{name: `bare file`, input: `snapshot-c.json`, wantPath: `import-snapshots/snapshot-c.json`},
		{name: `empty`, input: `   `, wantPath: ``},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeSnapshotDisplayPath(tt.input); got != tt.wantPath {
				t.Fatalf(`sanitizeSnapshotDisplayPath(%q) = %q, want %q`, tt.input, got, tt.wantPath)
			}
		})
	}
}

func TestPreviewRollbackImportSnapshotRejectsLatestMetadataFile(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, `/api/v1/setting/preview-rollback-import-snapshot`, map[string]string{
		`snapshot_name`: `latest-import-snapshot.json`,
	}, previewRollbackImportSnapshot)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(`status = %d, want %d, body = %s`, recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	if !strings.Contains(response.Message, `snapshot metadata file cannot be used as rollback target`) {
		t.Fatalf(`message = %q, want metadata-file guard`, response.Message)
	}
}

func performImportMultipartHandlerRequest(t *testing.T, target string, dump *model.DBDump, formFields map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(dump)
	if err != nil {
		t.Fatalf(`json.Marshal(dump) error = %v`, err)
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fileWriter, err := writer.CreateFormFile(`file`, `import.json`)
	if err != nil {
		t.Fatalf(`CreateFormFile() error = %v`, err)
	}
	if _, err := fileWriter.Write(payload); err != nil {
		t.Fatalf(`fileWriter.Write() error = %v`, err)
	}
	for key, value := range formFields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf(`WriteField(%s) error = %v`, key, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf(`writer.Close() error = %v`, err)
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, target, &body)
	ctx.Request.Header.Set(`Content-Type`, writer.FormDataContentType())
	importDB(ctx)
	return recorder
}

func testImportPreviewDump(apiBaseURL string) *model.DBDump {
	return &model.DBDump{
		Version: 1,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   `v1`,
			ExportSource:    `octopus`,
			ContainsSecrets: false,
		},
		Settings: []model.Setting{{
			Key:   model.SettingKeyAPIBaseURL,
			Value: apiBaseURL,
		}},
	}
}

func containsHandlerWarning(items []string, want string) bool {
	for _, item := range items {
		if strings.Contains(item, want) {
			return true
		}
	}
	return false
}

func marshalHandlerJSONField(t *testing.T, value any) string {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf(`json.Marshal(field) error = %v`, err)
	}
	return string(payload)
}

func cloneHandlerFormFields(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	cloned := make(map[string]string, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}

func assertStoredSettingValue(t *testing.T, key model.SettingKey, want string) {
	t.Helper()

	stored, err := op.SettingGetString(key)
	if err != nil {
		t.Fatalf(`SettingGetString(%s) error = %v`, key, err)
	}
	if stored != want {
		t.Fatalf(`setting %s = %q, want %q`, key, stored, want)
	}
}
