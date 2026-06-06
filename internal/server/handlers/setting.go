package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/task"
)

type dbExportFormat string

const (
	dbExportFormatStandard       dbExportFormat = "standard"
	dbExportFormatLegacy         dbExportFormat = "legacy"
	importSnapshotDisplayDirName                = "import-snapshots"
)

var maxDBImportPayloadBytes int64 = 128 << 20
var maxDBImportFileBytes int64 = 128 << 20
var refreshCachesAfterMutableSettingOperation = op.InitCache
var syncMutableSettingTasksAfterCacheRefresh = syncMutableSettingTasksFromCache

func respondMutableSettingOperationRefreshFailure(c *gin.Context, operation string, err error) {
	resp.Error(c, http.StatusInternalServerError, fmt.Sprintf("%s succeeded but cache refresh failed: %v", operation, err))
}

func respondMutableSettingOperationTaskSyncFailure(c *gin.Context, operation string, err error) {
	resp.Error(c, http.StatusInternalServerError, fmt.Sprintf("%s succeeded but task schedule sync failed: %v", operation, err))
}

func syncMutableSettingTasksFromCache() error {
	statsSaveIntervalMinutes, err := op.SettingGetInt(model.SettingKeyStatsSaveInterval)
	if err != nil {
		return fmt.Errorf("load %s: %w", model.SettingKeyStatsSaveInterval, err)
	}
	task.Update(string(model.SettingKeyStatsSaveInterval), time.Duration(statsSaveIntervalMinutes)*time.Minute)

	modelInfoUpdateIntervalHours, err := op.SettingGetInt(model.SettingKeyModelInfoUpdateInterval)
	if err != nil {
		return fmt.Errorf("load %s: %w", model.SettingKeyModelInfoUpdateInterval, err)
	}
	task.Update(string(model.SettingKeyModelInfoUpdateInterval), time.Duration(modelInfoUpdateIntervalHours)*time.Hour)

	syncLLMIntervalHours, err := op.SettingGetInt(model.SettingKeySyncLLMInterval)
	if err != nil {
		return fmt.Errorf("load %s: %w", model.SettingKeySyncLLMInterval, err)
	}
	task.Update(string(model.SettingKeySyncLLMInterval), time.Duration(syncLLMIntervalHours)*time.Hour)

	return nil
}

func init() {
	router.NewGroupRouter("/api/v1/setting").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(getSettingList),
		).
		AddRoute(
			router.NewRoute("/public-access", http.MethodGet).
				Handle(getPublicAccess),
		).
		AddRoute(
			router.NewRoute("/set", http.MethodPost).
				Use(middleware.RequireJSON()).
				Handle(setSetting),
		).
		AddRoute(
			router.NewRoute("/export", http.MethodGet).
				Handle(exportDB),
		).
		AddRoute(
			router.NewRoute("/import", http.MethodPost).
				Handle(importDB),
		).
		AddRoute(
			router.NewRoute("/import-snapshots", http.MethodGet).
				Handle(listImportSnapshots),
		).
		AddRoute(
			router.NewRoute("/rollback-latest-import", http.MethodPost).
				Use(middleware.RequireJSON()).
				Handle(rollbackLatestImportSnapshot),
		).
		AddRoute(
			router.NewRoute("/rollback-import-snapshot", http.MethodPost).
				Use(middleware.RequireJSON()).
				Handle(rollbackImportSnapshot),
		).
		AddRoute(
			router.NewRoute("/preview-rollback-import-snapshot", http.MethodPost).
				Use(middleware.RequireJSON()).
				Handle(previewRollbackImportSnapshot),
		)
}

func getSettingList(c *gin.Context) {
	settings, err := op.SettingList(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, settings)
}

func getPublicAccess(c *gin.Context) {
	resp.Success(c, op.PublicAccessInfo(c.Request))
}

func setSetting(c *gin.Context) {
	var setting model.Setting
	if err := c.ShouldBindJSON(&setting); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := setting.Validate(); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := op.SettingSetString(setting.Key, setting.Value); err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	switch setting.Key {
	case model.SettingKeyStatsSaveInterval:
		minutes, err := strconv.Atoi(setting.Value)
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		task.Update(string(setting.Key), time.Duration(minutes)*time.Minute)
	case model.SettingKeyModelInfoUpdateInterval:
		hours, err := strconv.Atoi(setting.Value)
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		task.Update(string(setting.Key), time.Duration(hours)*time.Hour)
	case model.SettingKeySyncLLMInterval:
		hours, err := strconv.Atoi(setting.Value)
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		task.Update(string(setting.Key), time.Duration(hours)*time.Hour)
	}
	resp.Success(c, setting)
}

func exportDB(c *gin.Context) {
	includeLogs, err := parseOptionalBoolQuery(c, "include_logs", false)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	includeStats, err := parseOptionalBoolQuery(c, "include_stats", false)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	exportFormat, err := parseOptionalDBExportFormat(c, "format", dbExportFormatStandard)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	includeSecrets, err := parseOptionalBoolQuery(c, "include_secrets", true)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	dump, err := op.DBExportAll(c.Request.Context(), includeLogs, includeStats, includeSecrets)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Header("Content-Type", "application/json")
	filename := "octopus-export-" + time.Now().Format("20060102150405") + ".json"
	if exportFormat == dbExportFormatLegacy {
		filename = "octopus-export-legacy-" + time.Now().Format("20060102150405") + ".json"
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	if exportFormat == dbExportFormatLegacy {
		c.JSON(http.StatusOK, op.ExportDumpLegacyView(dump))
		return
	}
	c.JSON(http.StatusOK, dump)
}

func importDB(c *gin.Context) {
	var dump model.DBDump
	dryRun, err := parseOptionalBoolQuery(c, "dry_run", false)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	mode, _, err := parseOptionalDBImportModeQuery(c, "mode", model.DBImportModeIncremental)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, "unsupported import mode")
		return
	}
	options := model.DBImportOptions{}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxDBImportPayloadBytes)
	contentType := c.GetHeader("Content-Type")
	previewToken := ""
	rawModelMappings := ""
	rawImportScopes := ""

	if strings.Contains(contentType, "multipart/form-data") {
		previewToken, _, err = parseOptionalNonEmptyTrimmedPostForm(c, "preview_token")
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		rawModelMappings, _, err = parseOptionalNonEmptyTrimmedPostForm(c, "model_mappings")
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		rawImportScopes, _, err = parseOptionalNonEmptyTrimmedPostForm(c, "import_scopes")
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		fh, err := c.FormFile("file")
		if err != nil {
			respondImportReadError(c, err, "missing upload file field 'file'")
			return
		}
		f, err := fh.Open()
		if err != nil {
			respondImportReadError(c, err, err.Error())
			return
		}
		defer f.Close()
		body, err := io.ReadAll(io.LimitReader(f, maxDBImportFileBytes+1))
		if err != nil {
			respondImportReadError(c, err, err.Error())
			return
		}
		if int64(len(body)) > maxDBImportFileBytes {
			resp.Error(c, http.StatusRequestEntityTooLarge, "import payload too large")
			return
		}
		if err := decodeDBDump(body, &dump); err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	} else if isJSONMediaType(contentType) {
		previewToken, _, err = parseOptionalNonEmptyTrimmedStringQuery(c, "preview_token")
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		rawModelMappings, _, err = parseOptionalNonEmptyTrimmedStringQuery(c, "model_mappings")
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		rawImportScopes, _, err = parseOptionalNonEmptyTrimmedStringQuery(c, "import_scopes")
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			respondImportReadError(c, err, err.Error())
			return
		}
		if err := decodeDBDump(body, &dump); err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	} else {
		resp.Error(c, http.StatusUnsupportedMediaType, "import content type must be multipart/form-data or application/json")
		return
	}
	if previewToken == "" {
		previewToken, _, err = parseOptionalNonEmptyTrimmedHeader(c, "X-Octopus-Import-Preview-Token")
		if err != nil {
			resp.Error(c, http.StatusBadRequest, "invalid preview_token")
			return
		}
	}
	if rawModelMappings != "" {
		if err := json.Unmarshal([]byte(rawModelMappings), &options.ModelMappings); err != nil {
			resp.Error(c, http.StatusBadRequest, "invalid model_mappings json")
			return
		}
		if err := validateImportModelMappings(options.ModelMappings); err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	}
	if rawImportScopes != "" {
		var scopes model.DBImportScopes
		if err := json.Unmarshal([]byte(rawImportScopes), &scopes); err != nil {
			resp.Error(c, http.StatusBadRequest, "invalid import_scopes json")
			return
		}
		if err := validateImportScopes(&scopes); err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		options.ImportScopes = &scopes
	}

	previewDigest, err := buildImportPreviewDigest(&dump, mode, options)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, fmt.Sprintf("build import preview digest: %v", err))
		return
	}
	if !dryRun && mode == model.DBImportModeReplace {
		if err := verifyImportPreviewToken(previewToken, previewDigest); err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	}
	if !dryRun && previewToken != "" && mode != model.DBImportModeReplace {
		if err := verifyImportPreviewToken(previewToken, previewDigest); err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	}

	result, err := op.DBImportIncrementalWithOptions(c.Request.Context(), &dump, mode, dryRun, options)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if dryRun {
		token, err := signImportPreviewToken(previewDigest)
		if err != nil {
			resp.Error(c, http.StatusInternalServerError, fmt.Sprintf("sign import preview token: %v", err))
			return
		}
		result.PreviewToken = token
	}

	if !dryRun {
		if err := refreshCachesAfterMutableSettingOperation(); err != nil {
			respondMutableSettingOperationRefreshFailure(c, "import", err)
			return
		}
		if err := syncMutableSettingTasksAfterCacheRefresh(); err != nil {
			respondMutableSettingOperationTaskSyncFailure(c, "import", err)
			return
		}
	}

	resp.Success(c, result)
}

func rollbackLatestImportSnapshot(c *gin.Context) {
	result, err := op.DBRollbackLatestImportSnapshot(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := refreshCachesAfterMutableSettingOperation(); err != nil {
		respondMutableSettingOperationRefreshFailure(c, "rollback latest import snapshot", err)
		return
	}
	if err := syncMutableSettingTasksAfterCacheRefresh(); err != nil {
		respondMutableSettingOperationTaskSyncFailure(c, "rollback latest import snapshot", err)
		return
	}
	resp.Success(c, sanitizeRollbackResult(result))
}

func listImportSnapshots(c *gin.Context) {
	items, err := op.DBListImportSnapshots()
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	sanitized := make([]model.DBImportSnapshotInfo, 0, len(items))
	for _, item := range items {
		sanitized = append(sanitized, sanitizeImportSnapshotInfo(item))
	}
	resp.Success(c, sanitized)
}

func rollbackImportSnapshot(c *gin.Context) {
	var payload struct {
		SnapshotName string                `json:"snapshot_name"`
		ImportScopes *model.DBImportScopes `json:"import_scopes"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateImportScopes(payload.ImportScopes); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := op.DBRollbackImportSnapshot(c.Request.Context(), payload.SnapshotName, payload.ImportScopes)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := refreshCachesAfterMutableSettingOperation(); err != nil {
		respondMutableSettingOperationRefreshFailure(c, "rollback import snapshot", err)
		return
	}
	if err := syncMutableSettingTasksAfterCacheRefresh(); err != nil {
		respondMutableSettingOperationTaskSyncFailure(c, "rollback import snapshot", err)
		return
	}
	resp.Success(c, sanitizeRollbackResult(result))
}

func previewRollbackImportSnapshot(c *gin.Context) {
	var payload struct {
		SnapshotName string                `json:"snapshot_name"`
		ImportScopes *model.DBImportScopes `json:"import_scopes"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateImportScopes(payload.ImportScopes); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := op.DBPreviewRollbackImportSnapshot(c.Request.Context(), payload.SnapshotName, payload.ImportScopes)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	resp.Success(c, sanitizeRollbackPreviewResult(result))
}

func sanitizeSnapshotDisplayPath(snapshotName string) string {
	trimmed := strings.TrimSpace(snapshotName)
	if trimmed == "" {
		return ""
	}
	displayName := filepath.Base(strings.ReplaceAll(trimmed, `\`, `/`))
	if displayName == "." || displayName == "/" || strings.TrimSpace(displayName) == "" {
		return ""
	}
	return importSnapshotDisplayDirName + "/" + displayName
}

func sanitizeImportSnapshotInfo(item model.DBImportSnapshotInfo) model.DBImportSnapshotInfo {
	item.SnapshotPath = sanitizeSnapshotDisplayPath(item.SnapshotName)
	return item
}

func sanitizeRollbackResult(result *model.DBRollbackResult) *model.DBRollbackResult {
	if result == nil {
		return nil
	}
	clone := *result
	clone.SnapshotPath = sanitizeSnapshotDisplayPath(clone.SnapshotName)
	return &clone
}

func sanitizeRollbackPreviewResult(result *model.DBRollbackPreviewResult) *model.DBRollbackPreviewResult {
	if result == nil {
		return nil
	}
	clone := *result
	clone.SnapshotPath = sanitizeSnapshotDisplayPath(clone.SnapshotName)
	return &clone
}

func respondImportReadError(c *gin.Context, err error, fallback string) {
	if isImportPayloadTooLarge(err) {
		resp.Error(c, http.StatusRequestEntityTooLarge, "import payload too large")
		return
	}
	resp.Error(c, http.StatusBadRequest, fallback)
}

func isImportPayloadTooLarge(err error) bool {
	if err == nil {
		return false
	}
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "request body too large")
}

func isJSONMediaType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}

func decodeDBDump(body []byte, dump *model.DBDump) error {
	if dump == nil {
		return json.Unmarshal(body, &struct{}{})
	}

	if err := json.Unmarshal(body, dump); err != nil {
		return err
	}

	if dump.Version == 0 &&
		len(dump.Channels) == 0 &&
		len(dump.Groups) == 0 &&
		len(dump.GroupItems) == 0 &&
		len(dump.Settings) == 0 &&
		len(dump.APIKeys) == 0 &&
		len(dump.LLMInfos) == 0 &&
		len(dump.RelayLogs) == 0 &&
		len(dump.StatsDaily) == 0 &&
		len(dump.StatsHourly) == 0 &&
		len(dump.StatsTotal) == 0 &&
		len(dump.StatsChannel) == 0 &&
		len(dump.StatsModel) == 0 &&
		len(dump.StatsAPIKey) == 0 {
		var wrapper struct {
			Code    int             `json:"code"`
			Message string          `json:"message"`
			Data    json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(body, &wrapper); err == nil && len(wrapper.Data) > 0 {
			if err := json.Unmarshal(wrapper.Data, dump); err != nil {
				return err
			}
		}
	}

	attachLegacyHints(body, dump)
	op.NormalizeLegacyDump(dump)
	return nil
}

func attachLegacyHints(body []byte, dump *model.DBDump) {
	if dump == nil {
		return
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return
	}
	if dataRaw, ok := raw["data"]; ok && len(dataRaw) > 0 {
		var wrapped map[string]json.RawMessage
		if err := json.Unmarshal(dataRaw, &wrapped); err == nil {
			raw = wrapped
		}
	}
	dump.LegacyHints = &model.DBDumpLegacyHints{
		MissingManifest:             !hasRawJSONField(raw, "manifest"),
		MissingUsers:                !hasRawJSONField(raw, "users"),
		MissingRouteTargetOverrides: !hasRawJSONField(raw, "route_target_overrides"),
		MissingMigrationRecords:     !hasRawJSONField(raw, "migration_records"),
		MissingRelayLogs:            !hasRawJSONField(raw, "relay_logs"),
		ChannelsByName:              collectLegacyChannelHints(raw),
		ChannelKeysBySnapshotID:     collectLegacyChannelKeyHints(raw),
		GroupsByName:                collectLegacyGroupHints(raw),
		LLMInfosByName:              collectLegacyLLMInfoHints(raw),
		APIKeysBySnapshotID:         collectLegacyAPIKeyHints(raw),
	}
}

func hasRawJSONField(raw map[string]json.RawMessage, key string) bool {
	if len(raw) == 0 {
		return false
	}
	value, ok := raw[key]
	if !ok {
		return false
	}
	return len(value) > 0 && strings.TrimSpace(string(value)) != "null"
}

func collectLegacyChannelHints(raw map[string]json.RawMessage) map[string]model.DBDumpLegacyChannelHint {
	items := make(map[string]model.DBDumpLegacyChannelHint)
	value, ok := raw["channels"]
	if !ok || len(value) == 0 {
		return items
	}
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(value, &rows); err != nil {
		return items
	}
	for _, row := range rows {
		name := extractJSONString(row, "name")
		if name == "" {
			continue
		}
		items[name] = model.DBDumpLegacyChannelHint{
			MissingKeyManagementMode: !hasRawJSONField(row, "key_management_mode"),
			MissingKeyRoutingPolicy:  !hasRawJSONField(row, "key_routing_policy"),
		}
	}
	return items
}

func collectLegacyChannelKeyHints(raw map[string]json.RawMessage) map[int]model.DBDumpLegacyChannelKeyHint {
	items := make(map[int]model.DBDumpLegacyChannelKeyHint)
	value, ok := raw["channel_keys"]
	if !ok || len(value) == 0 {
		return items
	}
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(value, &rows); err != nil {
		return items
	}
	for _, row := range rows {
		id := extractJSONInt(row, "id")
		items[id] = model.DBDumpLegacyChannelKeyHint{
			MissingSourceType:          !hasRawJSONField(row, "source_type"),
			MissingAllowedModels:       !hasRawJSONField(row, "allowed_models"),
			MissingRequestCapabilities: !hasRawJSONField(row, "request_capabilities"),
		}
	}
	return items
}

func collectLegacyGroupHints(raw map[string]json.RawMessage) map[string]model.DBDumpLegacyGroupHint {
	items := make(map[string]model.DBDumpLegacyGroupHint)
	value, ok := raw["groups"]
	if !ok || len(value) == 0 {
		return items
	}
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(value, &rows); err != nil {
		return items
	}
	for _, row := range rows {
		name := extractJSONString(row, "name")
		if name == "" {
			continue
		}
		items[name] = model.DBDumpLegacyGroupHint{
			MissingRetryRounds:       !hasRawJSONField(row, "retry_rounds"),
			MissingRetryDelayMs:      !hasRawJSONField(row, "retry_delay_ms"),
			MissingFailoverWindowSec: !hasRawJSONField(row, "failover_window_sec"),
			MissingRaceAfterFails:    !hasRawJSONField(row, "race_after_fails"),
			MissingRaceConcurrency:   !hasRawJSONField(row, "race_concurrency"),
		}
	}
	return items
}

func collectLegacyLLMInfoHints(raw map[string]json.RawMessage) map[string]model.DBDumpLegacyLLMInfoHint {
	items := make(map[string]model.DBDumpLegacyLLMInfoHint)
	value, ok := raw["llm_infos"]
	if !ok || len(value) == 0 {
		return items
	}
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(value, &rows); err != nil {
		return items
	}
	for _, row := range rows {
		name := strings.ToLower(strings.TrimSpace(extractJSONString(row, "name")))
		if name == "" {
			continue
		}
		items[name] = model.DBDumpLegacyLLMInfoHint{
			MissingCanonicalName:         !hasRawJSONField(row, "canonical_name"),
			MissingBillingMode:           !hasRawJSONField(row, "billing_mode"),
			MissingProbePolicy:           !hasRawJSONField(row, "probe_policy"),
			MissingProbeIntervalSeconds:  !hasRawJSONField(row, "probe_interval_seconds"),
			MissingProbeConcurrencyLimit: !hasRawJSONField(row, "probe_concurrency_limit"),
		}
	}
	return items
}

func collectLegacyAPIKeyHints(raw map[string]json.RawMessage) map[int]model.DBDumpLegacyAPIKeyHint {
	items := make(map[int]model.DBDumpLegacyAPIKeyHint)
	value, ok := raw["api_keys"]
	if !ok || len(value) == 0 {
		return items
	}
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(value, &rows); err != nil {
		return items
	}
	for _, row := range rows {
		id := extractJSONInt(row, "id")
		items[id] = model.DBDumpLegacyAPIKeyHint{MissingSupportedModels: !hasRawJSONField(row, "supported_models")}
	}
	return items
}

func extractJSONString(raw map[string]json.RawMessage, key string) string {
	value, ok := raw[key]
	if !ok || len(value) == 0 {
		return ""
	}
	var out string
	if err := json.Unmarshal(value, &out); err == nil {
		return strings.TrimSpace(out)
	}
	return ""
}

func extractJSONInt(raw map[string]json.RawMessage, key string) int {
	value, ok := raw[key]
	if !ok || len(value) == 0 {
		return 0
	}
	var out int
	if err := json.Unmarshal(value, &out); err == nil {
		return out
	}
	var asFloat float64
	if err := json.Unmarshal(value, &asFloat); err == nil {
		return int(asFloat)
	}
	return 0
}
