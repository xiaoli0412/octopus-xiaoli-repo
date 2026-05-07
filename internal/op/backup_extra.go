package op

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db/migrate"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/llmname"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type legacyExportChannel struct {
	ID            int                   `json:"id"`
	Name          string                `json:"name"`
	Type          outbound.OutboundType `json:"type,omitempty"`
	Enabled       bool                  `json:"enabled"`
	BaseUrls      []model.BaseUrl       `json:"base_urls,omitempty"`
	Keys          []model.ChannelKey    `json:"keys,omitempty"`
	Model         string                `json:"model,omitempty"`
	CustomModel   string                `json:"custom_model,omitempty"`
	Proxy         bool                  `json:"proxy"`
	AutoSync      bool                  `json:"auto_sync"`
	AutoGroup     model.AutoGroupType   `json:"auto_group"`
	CustomHeader  []model.CustomHeader  `json:"custom_header,omitempty"`
	ParamOverride *string               `json:"param_override,omitempty"`
	ChannelProxy  *string               `json:"channel_proxy,omitempty"`
	MatchRegex    *string               `json:"match_regex,omitempty"`
}

type legacyExportChannelKey struct {
	ID               int     `json:"id"`
	ChannelID        int     `json:"channel_id"`
	Enabled          bool    `json:"enabled"`
	ChannelKey       string  `json:"channel_key"`
	StatusCode       int     `json:"status_code"`
	LastUseTimeStamp int64   `json:"last_use_time_stamp"`
	TotalCost        float64 `json:"total_cost"`
	Remark           string  `json:"remark"`
}

type legacyExportGroup struct {
	ID                int             `json:"id"`
	Name              string          `json:"name"`
	Mode              model.GroupMode `json:"mode"`
	MatchRegex        string          `json:"match_regex,omitempty"`
	FirstTokenTimeOut int             `json:"first_token_time_out"`
	SessionKeepTime   int             `json:"session_keep_time"`
}

type legacyExportLLMInfo struct {
	Name string `json:"name"`
	model.LLMPrice
	model.OfficialLLMPrice
}

type legacyExportAPIKey struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	APIKey   string  `json:"api_key"`
	Enabled  bool    `json:"enabled"`
	ExpireAt int64   `json:"expire_at,omitempty"`
	MaxCost  float64 `json:"max_cost,omitempty"`
}

type legacyExportDump struct {
	Version      int                      `json:"version"`
	ExportedAt   time.Time                `json:"exported_at"`
	IncludeLogs  bool                     `json:"include_logs"`
	IncludeStats bool                     `json:"include_stats"`
	Channels     []legacyExportChannel    `json:"channels,omitempty"`
	ChannelKeys  []legacyExportChannelKey `json:"channel_keys,omitempty"`
	Groups       []legacyExportGroup      `json:"groups,omitempty"`
	GroupItems   []model.GroupItem        `json:"group_items,omitempty"`
	LLMInfos     []legacyExportLLMInfo    `json:"llm_infos,omitempty"`
	APIKeys      []legacyExportAPIKey     `json:"api_keys,omitempty"`
	Settings     []model.Setting          `json:"settings,omitempty"`
	StatsTotal   []model.StatsTotal       `json:"stats_total,omitempty"`
	StatsDaily   []model.StatsDaily       `json:"stats_daily,omitempty"`
	StatsHourly  []model.StatsHourly      `json:"stats_hourly,omitempty"`
	StatsModel   []model.StatsModel       `json:"stats_model,omitempty"`
	StatsChannel []model.StatsChannel     `json:"stats_channel,omitempty"`
	StatsAPIKey  []model.StatsAPIKey      `json:"stats_api_key,omitempty"`
}

func ensureDumpLegacyHints(dump *model.DBDump) *model.DBDumpLegacyHints {
	if dump == nil {
		return nil
	}
	if dump.LegacyHints == nil {
		dump.LegacyHints = &model.DBDumpLegacyHints{}
	}
	if dump.LegacyHints.ChannelsByName == nil {
		dump.LegacyHints.ChannelsByName = map[string]model.DBDumpLegacyChannelHint{}
	}
	if dump.LegacyHints.ChannelKeysBySnapshotID == nil {
		dump.LegacyHints.ChannelKeysBySnapshotID = map[int]model.DBDumpLegacyChannelKeyHint{}
	}
	if dump.LegacyHints.GroupsByName == nil {
		dump.LegacyHints.GroupsByName = map[string]model.DBDumpLegacyGroupHint{}
	}
	if dump.LegacyHints.LLMInfosByName == nil {
		dump.LegacyHints.LLMInfosByName = map[string]model.DBDumpLegacyLLMInfoHint{}
	}
	if dump.LegacyHints.APIKeysBySnapshotID == nil {
		dump.LegacyHints.APIKeysBySnapshotID = map[int]model.DBDumpLegacyAPIKeyHint{}
	}
	return dump.LegacyHints
}

func inferLegacyHintsFromDump(dump *model.DBDump) {
	if dump == nil {
		return
	}
	hints := ensureDumpLegacyHints(dump)
	if hints == nil {
		return
	}
	legacyLike := hints.Legacy || hints.MissingManifest || hints.MissingUsers || hints.MissingRouteTargetOverrides || hints.MissingMigrationRecords || hints.MissingRelayLogs
	if !legacyLike {
		exportSource := strings.ToLower(strings.TrimSpace(dump.Manifest.ExportSource))
		legacyLike = strings.Contains(exportSource, "legacy")
	}
	if !legacyLike {
		return
	}
	for _, row := range dump.Channels {
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		hint := hints.ChannelsByName[name]
		if row.KeyManagementMode == "" {
			hint.MissingKeyManagementMode = true
		}
		if row.KeyRoutingPolicy == "" {
			hint.MissingKeyRoutingPolicy = true
		}
		if hint.MissingKeyManagementMode || hint.MissingKeyRoutingPolicy {
			legacyLike = true
		}
		hints.ChannelsByName[name] = hint
	}
	for _, row := range dump.ChannelKeys {
		hint := hints.ChannelKeysBySnapshotID[row.ID]
		if strings.TrimSpace(row.SourceType) == "" {
			hint.MissingSourceType = true
		}
		if strings.TrimSpace(row.AllowedModels) == "" {
			hint.MissingAllowedModels = true
		}
		if hint.MissingSourceType || hint.MissingAllowedModels {
			legacyLike = true
		}
		hints.ChannelKeysBySnapshotID[row.ID] = hint
	}
	for _, row := range dump.Groups {
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		hint := hints.GroupsByName[name]
		if row.RetryRounds == 0 {
			hint.MissingRetryRounds = true
		}
		if row.RetryDelayMs == 0 {
			hint.MissingRetryDelayMs = true
		}
		if row.FailoverWindowSec == 0 {
			hint.MissingFailoverWindowSec = true
		}
		if row.RaceAfterFails == 0 {
			hint.MissingRaceAfterFails = true
		}
		if row.RaceConcurrency == 0 {
			hint.MissingRaceConcurrency = true
		}
		if hint.MissingRetryRounds || hint.MissingRetryDelayMs || hint.MissingFailoverWindowSec || hint.MissingRaceAfterFails || hint.MissingRaceConcurrency {
			legacyLike = true
		}
		hints.GroupsByName[name] = hint
	}
	for _, row := range dump.LLMInfos {
		name := normalizeModelRef(row.Name)
		if name == "" {
			continue
		}
		hint := hints.LLMInfosByName[name]
		if strings.TrimSpace(row.CanonicalName) == "" {
			hint.MissingCanonicalName = true
		}
		if strings.TrimSpace(string(row.BillingMode)) == "" {
			hint.MissingBillingMode = true
		}
		if strings.TrimSpace(string(row.ProbePolicy)) == "" {
			hint.MissingProbePolicy = true
		}
		if row.ProbeIntervalSeconds == 0 {
			hint.MissingProbeIntervalSeconds = true
		}
		if row.ProbeConcurrencyLimit == 0 {
			hint.MissingProbeConcurrencyLimit = true
		}
		if hint.MissingCanonicalName || hint.MissingBillingMode || hint.MissingProbePolicy || hint.MissingProbeIntervalSeconds || hint.MissingProbeConcurrencyLimit {
			legacyLike = true
		}
		hints.LLMInfosByName[name] = hint
	}
	for _, row := range dump.APIKeys {
		hint := hints.APIKeysBySnapshotID[row.ID]
		if strings.TrimSpace(row.SupportedModels) == "" {
			hint.MissingSupportedModels = true
		}
		if hint.MissingSupportedModels {
			legacyLike = true
		}
		hints.APIKeysBySnapshotID[row.ID] = hint
	}
	hints.Legacy = legacyLike
}

func NormalizeLegacyDump(dump *model.DBDump) {
	if dump == nil {
		return
	}
	hints := ensureDumpLegacyHints(dump)
	inferLegacyHintsFromDump(dump)
	if hints == nil {
		return
	}
	if hints.MissingManifest {
		dump.Manifest = model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus-legacy",
			Encrypted:       false,
			ContainsSecrets: dumpContainsSecrets(dump),
		}
	} else {
		if strings.TrimSpace(dump.Manifest.SchemaVersion) == "" {
			dump.Manifest.SchemaVersion = "v1"
		}
		if strings.TrimSpace(dump.Manifest.ExportSource) == "" {
			if hints.Legacy {
				dump.Manifest.ExportSource = "octopus-legacy"
			} else {
				dump.Manifest.ExportSource = "octopus"
			}
		}
		dump.Manifest.ContainsSecrets = dumpContainsSecrets(dump)
	}
	for i := range dump.Channels {
		name := strings.TrimSpace(dump.Channels[i].Name)
		if name == "" {
			continue
		}
		hint := hints.ChannelsByName[name]
		if hint.MissingKeyManagementMode {
			dump.Channels[i].KeyManagementMode = model.KeyManagementModePooled
		}
		if hint.MissingKeyRoutingPolicy {
			dump.Channels[i].KeyRoutingPolicy = model.KeyRoutingPolicyRoundRobin
		}
	}
	for i := range dump.ChannelKeys {
		hint := hints.ChannelKeysBySnapshotID[dump.ChannelKeys[i].ID]
		if hint.MissingSourceType {
			dump.ChannelKeys[i].SourceType = model.ChannelKeySourceTypeUnknown
		}
		if hint.MissingAllowedModels {
			dump.ChannelKeys[i].AllowedModels = ""
		}
	}
	for i := range dump.LLMInfos {
		name := normalizeModelRef(dump.LLMInfos[i].Name)
		if name == "" {
			continue
		}
		hint := hints.LLMInfosByName[name]
		if hint.MissingCanonicalName {
			dump.LLMInfos[i].CanonicalName = llmname.CanonicalModelName(dump.LLMInfos[i].Name)
		}
		if hint.MissingBillingMode {
			dump.LLMInfos[i].BillingMode = model.BillingModeUnknown
		}
		if hint.MissingProbePolicy {
			dump.LLMInfos[i].ProbePolicy = model.ProbePolicyPassiveOnly
		}
		if hint.MissingProbeIntervalSeconds {
			dump.LLMInfos[i].ProbeIntervalSeconds = 0
		}
		if hint.MissingProbeConcurrencyLimit {
			dump.LLMInfos[i].ProbeConcurrencyLimit = 0
		}
	}
	for i := range dump.APIKeys {
		hint := hints.APIKeysBySnapshotID[dump.APIKeys[i].ID]
		if hint.MissingSupportedModels {
			dump.APIKeys[i].SupportedModels = normalizeAPIKeySupportedModels(dump.APIKeys[i].SupportedModels)
		}
	}
}

func ExportDumpLegacyView(dump *model.DBDump) *legacyExportDump {
	if dump == nil {
		return nil
	}
	view := &legacyExportDump{
		Version:      dump.Version,
		ExportedAt:   dump.ExportedAt,
		IncludeLogs:  dump.IncludeLogs,
		IncludeStats: dump.IncludeStats,
		Channels:     make([]legacyExportChannel, 0, len(dump.Channels)),
		ChannelKeys:  make([]legacyExportChannelKey, 0, len(dump.ChannelKeys)),
		Groups:       make([]legacyExportGroup, 0, len(dump.Groups)),
		GroupItems:   append([]model.GroupItem(nil), dump.GroupItems...),
		LLMInfos:     make([]legacyExportLLMInfo, 0, len(dump.LLMInfos)),
		APIKeys:      make([]legacyExportAPIKey, 0, len(dump.APIKeys)),
		Settings:     append([]model.Setting(nil), dump.Settings...),
		StatsTotal:   append([]model.StatsTotal(nil), dump.StatsTotal...),
		StatsDaily:   append([]model.StatsDaily(nil), dump.StatsDaily...),
		StatsHourly:  append([]model.StatsHourly(nil), dump.StatsHourly...),
		StatsModel:   append([]model.StatsModel(nil), dump.StatsModel...),
		StatsChannel: append([]model.StatsChannel(nil), dump.StatsChannel...),
		StatsAPIKey:  append([]model.StatsAPIKey(nil), dump.StatsAPIKey...),
	}
	for _, row := range dump.Channels {
		view.Channels = append(view.Channels, legacyExportChannel{
			ID:            row.ID,
			Name:          row.Name,
			Type:          row.Type,
			Enabled:       row.Enabled,
			BaseUrls:      append([]model.BaseUrl(nil), row.BaseUrls...),
			Keys:          nil,
			Model:         row.Model,
			CustomModel:   row.CustomModel,
			Proxy:         row.Proxy,
			AutoSync:      row.AutoSync,
			AutoGroup:     row.AutoGroup,
			CustomHeader:  append([]model.CustomHeader(nil), row.CustomHeader...),
			ParamOverride: row.ParamOverride,
			ChannelProxy:  row.ChannelProxy,
			MatchRegex:    row.MatchRegex,
		})
	}
	for _, row := range dump.ChannelKeys {
		view.ChannelKeys = append(view.ChannelKeys, legacyExportChannelKey{
			ID:               row.ID,
			ChannelID:        row.ChannelID,
			Enabled:          row.Enabled,
			ChannelKey:       row.ChannelKey,
			StatusCode:       row.StatusCode,
			LastUseTimeStamp: row.LastUseTimeStamp,
			TotalCost:        row.TotalCost,
			Remark:           row.Remark,
		})
	}
	for _, row := range dump.Groups {
		view.Groups = append(view.Groups, legacyExportGroup{
			ID:                row.ID,
			Name:              row.Name,
			Mode:              row.Mode,
			MatchRegex:        row.MatchRegex,
			FirstTokenTimeOut: row.FirstTokenTimeOut,
			SessionKeepTime:   row.SessionKeepTime,
		})
	}
	for _, row := range dump.LLMInfos {
		view.LLMInfos = append(view.LLMInfos, legacyExportLLMInfo{
			Name:             row.Name,
			LLMPrice:         row.LLMPrice,
			OfficialLLMPrice: row.OfficialLLMPrice,
		})
	}
	for _, row := range dump.APIKeys {
		view.APIKeys = append(view.APIKeys, legacyExportAPIKey{
			ID:       row.ID,
			Name:     row.Name,
			APIKey:   row.APIKey,
			Enabled:  row.Enabled,
			ExpireAt: row.ExpireAt,
			MaxCost:  row.MaxCost,
		})
	}
	return view
}

func dumpContainsSecrets(dump *model.DBDump) bool {
	if dump == nil {
		return false
	}
	for _, row := range dump.Users {
		if strings.TrimSpace(row.Password) != "" {
			return true
		}
	}
	for _, row := range dump.ChannelKeys {
		if strings.TrimSpace(row.ChannelKey) != "" {
			return true
		}
	}
	for _, row := range dump.APIKeys {
		if strings.TrimSpace(row.APIKey) != "" {
			return true
		}
	}
	for _, channel := range dump.Channels {
		if channelProxyContainsSecret(channel.ChannelProxy) {
			return true
		}
		for _, header := range channel.CustomHeader {
			if customHeaderContainsSecret(header) {
				return true
			}
		}
	}
	return false
}

func customHeaderContainsSecret(header model.CustomHeader) bool {
	if !isSensitiveCustomHeaderKey(header.HeaderKey) {
		return false
	}
	return strings.TrimSpace(header.HeaderValue) != ""
}

func isSensitiveCustomHeaderKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if normalized == "" {
		return false
	}
	switch normalized {
	case "authorization", "proxy-authorization", "x-api-key", "api-key", "x-auth-token", "x-access-token", "x-openai-api-key", "anthropic-api-key", "cf-access-client-secret":
		return true
	}
	if strings.Contains(normalized, "secret") || strings.Contains(normalized, "token") {
		return true
	}
	if strings.HasSuffix(normalized, "api-key") || strings.HasSuffix(normalized, "apikey") {
		return true
	}
	return false
}

func channelProxyContainsSecret(raw *string) bool {
	if raw == nil {
		return false
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return false
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return true
	}
	return parsed.User != nil
}

func sanitizeChannelProxyForRedactedExport(raw *string) *string {
	if raw == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		empty := ""
		return &empty
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		empty := ""
		return &empty
	}
	if parsed.User == nil {
		clean := parsed.String()
		return &clean
	}
	parsed.User = nil
	clean := parsed.String()
	return &clean
}

func redactDumpSecrets(dump *model.DBDump) {
	if dump == nil {
		return
	}
	for i := range dump.Users {
		dump.Users[i].Password = ""
	}
	for i := range dump.ChannelKeys {
		dump.ChannelKeys[i].ChannelKey = ""
	}
	for i := range dump.APIKeys {
		dump.APIKeys[i].APIKey = ""
	}
	for i := range dump.Channels {
		dump.Channels[i].ChannelProxy = sanitizeChannelProxyForRedactedExport(dump.Channels[i].ChannelProxy)
		for j := range dump.Channels[i].CustomHeader {
			if customHeaderContainsSecret(dump.Channels[i].CustomHeader[j]) {
				dump.Channels[i].CustomHeader[j].HeaderValue = ""
			}
		}
	}
}

func cloneDumpForImport(dump *model.DBDump) *model.DBDump {
	if dump == nil {
		return nil
	}
	payload, err := json.Marshal(dump)
	if err != nil {
		clone := *dump
		clone.LegacyHints = cloneLegacyHints(dump.LegacyHints)
		return &clone
	}
	var clone model.DBDump
	if err := json.Unmarshal(payload, &clone); err != nil {
		fallback := *dump
		fallback.LegacyHints = cloneLegacyHints(dump.LegacyHints)
		return &fallback
	}
	clone.LegacyHints = cloneLegacyHints(dump.LegacyHints)
	return &clone
}

func cloneLegacyHints(hints *model.DBDumpLegacyHints) *model.DBDumpLegacyHints {
	if hints == nil {
		return nil
	}
	clone := *hints
	if hints.ChannelsByName != nil {
		clone.ChannelsByName = make(map[string]model.DBDumpLegacyChannelHint, len(hints.ChannelsByName))
		for key, value := range hints.ChannelsByName {
			clone.ChannelsByName[key] = value
		}
	}
	if hints.ChannelKeysBySnapshotID != nil {
		clone.ChannelKeysBySnapshotID = make(map[int]model.DBDumpLegacyChannelKeyHint, len(hints.ChannelKeysBySnapshotID))
		for key, value := range hints.ChannelKeysBySnapshotID {
			clone.ChannelKeysBySnapshotID[key] = value
		}
	}
	if hints.GroupsByName != nil {
		clone.GroupsByName = make(map[string]model.DBDumpLegacyGroupHint, len(hints.GroupsByName))
		for key, value := range hints.GroupsByName {
			clone.GroupsByName[key] = value
		}
	}
	if hints.LLMInfosByName != nil {
		clone.LLMInfosByName = make(map[string]model.DBDumpLegacyLLMInfoHint, len(hints.LLMInfosByName))
		for key, value := range hints.LLMInfosByName {
			clone.LLMInfosByName[key] = value
		}
	}
	if hints.APIKeysBySnapshotID != nil {
		clone.APIKeysBySnapshotID = make(map[int]model.DBDumpLegacyAPIKeyHint, len(hints.APIKeysBySnapshotID))
		for key, value := range hints.APIKeysBySnapshotID {
			clone.APIKeysBySnapshotID[key] = value
		}
	}
	return &clone
}

func applyImportOptionsToDump(dump *model.DBDump, options model.DBImportOptions) {
	if dump == nil {
		return
	}
	applyImportScopesToDump(dump, options.ImportScopes)
	applyModelMappingsToDump(dump, options.ModelMappings)
}

func applyModelMappingsToDump(dump *model.DBDump, input map[string]string) {
	if dump == nil {
		return
	}
	mappings := normalizeModelMappings(input)
	if len(mappings) == 0 {
		return
	}
	for i := range dump.Channels {
		dump.Channels[i].Model = remapCSVWithModelMappings(dump.Channels[i].Model, mappings)
		dump.Channels[i].CustomModel = remapCSVWithModelMappings(dump.Channels[i].CustomModel, mappings)
	}
	for i := range dump.ChannelKeys {
		dump.ChannelKeys[i].AllowedModels = remapCSVWithModelMappings(dump.ChannelKeys[i].AllowedModels, mappings)
	}
	for i := range dump.RouteTargetOverrides {
		dump.RouteTargetOverrides[i].ModelName = remapSingleModelWithMappings(dump.RouteTargetOverrides[i].ModelName, mappings)
	}
	for i := range dump.GroupItems {
		dump.GroupItems[i].ModelName = remapSingleModelWithMappings(dump.GroupItems[i].ModelName, mappings)
	}
	for i := range dump.LLMInfos {
		if mapped := remapSingleModelWithMappings(dump.LLMInfos[i].Name, mappings); mapped != "" {
			dump.LLMInfos[i].Name = mapped
		}
		if mapped := remapSingleModelWithMappings(dump.LLMInfos[i].CanonicalName, mappings); mapped != "" {
			dump.LLMInfos[i].CanonicalName = mapped
		}
	}
	for i := range dump.APIKeys {
		dump.APIKeys[i].SupportedModels = remapCSVWithModelMappings(dump.APIKeys[i].SupportedModels, mappings)
	}
	for i := range dump.StatsModel {
		dump.StatsModel[i].Name = remapSingleModelWithMappings(dump.StatsModel[i].Name, mappings)
	}
	for i := range dump.RelayLogs {
		dump.RelayLogs[i].RequestModelName = remapSingleModelWithMappings(dump.RelayLogs[i].RequestModelName, mappings)
		dump.RelayLogs[i].ActualModelName = remapSingleModelWithMappings(dump.RelayLogs[i].ActualModelName, mappings)
		for j := range dump.RelayLogs[i].Attempts {
			dump.RelayLogs[i].Attempts[j].ModelName = remapSingleModelWithMappings(dump.RelayLogs[i].Attempts[j].ModelName, mappings)
		}
	}
}

func applyImportScopesToDump(dump *model.DBDump, scopes *model.DBImportScopes) {
	if dump == nil || scopes == nil {
		return
	}
	dump.Users = nil
	dump.MigrationRecords = nil
	if !scopes.Routing {
		dump.Channels = nil
		dump.ChannelKeys = nil
		dump.RouteTargetOverrides = nil
		dump.Groups = nil
		dump.GroupItems = nil
	}
	if !scopes.Models {
		dump.LLMInfos = nil
	}
	if !scopes.APIKeys {
		dump.APIKeys = nil
	}
	if !scopes.Settings {
		dump.Settings = nil
	}
	if !scopes.Stats {
		dump.IncludeStats = false
		dump.StatsTotal = nil
		dump.StatsDaily = nil
		dump.StatsHourly = nil
		dump.StatsModel = nil
		dump.StatsChannel = nil
		dump.StatsAPIKey = nil
	}
	if !scopes.Logs {
		dump.IncludeLogs = false
		dump.RelayLogs = nil
	}
}

func cloneImportScopes(scopes *model.DBImportScopes) *model.DBImportScopes {
	if scopes == nil {
		return nil
	}
	clone := *scopes
	return &clone
}

func isFullImportScopes(scopes *model.DBImportScopes) bool {
	if scopes == nil {
		return true
	}
	return scopes.Routing && scopes.Models && scopes.APIKeys && scopes.Settings && scopes.Stats && scopes.Logs
}

func validateImportScopes(scopes *model.DBImportScopes) error {
	if scopes == nil {
		return nil
	}
	return scopes.Validate()
}

func validateImportModelMappings(input map[string]string) error {
	if len(input) == 0 {
		return nil
	}
	for source, target := range input {
		if normalizeModelRef(source) == "" || normalizeModelRef(target) == "" {
			return fmt.Errorf("invalid model_mappings")
		}
	}
	return nil
}

func normalizeModelMappings(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for source, target := range input {
		s := normalizeModelRef(source)
		t := normalizeModelRef(target)
		if s == "" || t == "" {
			continue
		}
		out[s] = t
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func remapSingleModelWithMappings(modelName string, mappings map[string]string) string {
	name := normalizeModelRef(modelName)
	if name == "" {
		return ""
	}
	if mapped, ok := mappings[name]; ok && mapped != "" {
		return mapped
	}
	return name
}

func remapCSVWithModelMappings(input string, mappings map[string]string) string {
	parts := splitCSVModels(input)
	if len(parts) == 0 {
		return ""
	}
	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		mapped := remapSingleModelWithMappings(part, mappings)
		if mapped == "" {
			continue
		}
		if _, ok := seen[mapped]; ok {
			continue
		}
		seen[mapped] = struct{}{}
		out = append(out, mapped)
	}
	sort.Strings(out)
	return strings.Join(out, ",")
}

func importSnapshotDir() (string, error) {
	dbType := strings.ToLower(strings.TrimSpace(db.GetCurrentDBType()))
	if dbType == "sqlite" {
		dsn := strings.TrimSpace(db.GetCurrentDSN())
		if dsn != "" {
			if abs, err := filepath.Abs(dsn); err == nil {
				return filepath.Join(filepath.Dir(abs), importSnapshotDirName), nil
			}
			return filepath.Join(filepath.Dir(dsn), importSnapshotDirName), nil
		}
	}
	configured := strings.TrimSpace(conf.AppConfig.Database.Path)
	if configured != "" {
		if abs, err := filepath.Abs(configured); err == nil {
			return filepath.Join(filepath.Dir(abs), importSnapshotDirName), nil
		}
		return filepath.Join(filepath.Dir(configured), importSnapshotDirName), nil
	}
	return filepath.Abs(filepath.Join("data", importSnapshotDirName))
}

func loadCurrentImportState(conn *gorm.DB) (*currentImportState, error) {
	state := &currentImportState{
		channelsByName:            map[string]model.Channel{},
		channelsByID:              map[int]model.Channel{},
		groupsByName:              map[string]model.Group{},
		llmInfosByName:            map[string]model.LLMInfo{},
		llmInfosByCanonical:       map[string]model.LLMInfo{},
		settingsByKey:             map[string]model.Setting{},
		apiKeysByAPIKey:           map[string]model.APIKey{},
		routeTargetOverridesByKey: map[string]model.RouteTargetOverride{},
	}
	if conn == nil {
		return state, nil
	}
	var channels []model.Channel
	if err := conn.Preload("Keys").Find(&channels).Error; err != nil {
		return nil, fmt.Errorf("load current channels: %w", err)
	}
	for _, row := range channels {
		state.channelsByName[strings.TrimSpace(row.Name)] = row
		state.channelsByID[row.ID] = row
	}
	var groups []model.Group
	if err := conn.Preload("Items").Find(&groups).Error; err != nil {
		return nil, fmt.Errorf("load current groups: %w", err)
	}
	for _, row := range groups {
		state.groupsByName[strings.TrimSpace(row.Name)] = row
	}
	var llmInfos []model.LLMInfo
	if err := conn.Find(&llmInfos).Error; err != nil {
		return nil, fmt.Errorf("load current llm infos: %w", err)
	}
	for _, row := range llmInfos {
		nameKey := normalizeModelRef(row.Name)
		if nameKey != "" {
			state.llmInfosByName[nameKey] = row
		}
		canonicalKey := normalizeModelRef(row.CanonicalName)
		if canonicalKey == "" {
			canonicalKey = llmname.CanonicalModelName(row.Name)
		}
		if canonicalKey != "" {
			state.llmInfosByCanonical[canonicalKey] = row
		}
	}
	var settings []model.Setting
	if err := conn.Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("load current settings: %w", err)
	}
	for _, row := range settings {
		state.settingsByKey[strings.TrimSpace(string(row.Key))] = row
	}
	var apiKeys []model.APIKey
	if err := conn.Find(&apiKeys).Error; err != nil {
		return nil, fmt.Errorf("load current api keys: %w", err)
	}
	for _, row := range apiKeys {
		apiKey := strings.TrimSpace(row.APIKey)
		if apiKey == "" {
			continue
		}
		state.apiKeysByAPIKey[apiKey] = row
	}
	var overrides []model.RouteTargetOverride
	if err := conn.Find(&overrides).Error; err != nil {
		return nil, fmt.Errorf("load current route target overrides: %w", err)
	}
	for _, row := range overrides {
		state.routeTargetOverridesByKey[routeTargetOverrideLookupKey(row.ChannelID, row.ChannelKeyID, row.ModelName)] = row
	}
	return state, nil
}

func sameBaseURLs(left, right []model.BaseUrl) bool {
	if len(left) != len(right) {
		return false
	}
	leftCopy := append([]model.BaseUrl(nil), left...)
	rightCopy := append([]model.BaseUrl(nil), right...)
	sort.Slice(leftCopy, func(i, j int) bool { return leftCopy[i].URL < leftCopy[j].URL })
	sort.Slice(rightCopy, func(i, j int) bool { return rightCopy[i].URL < rightCopy[j].URL })
	for i := range leftCopy {
		if leftCopy[i].URL != rightCopy[i].URL || leftCopy[i].Delay != rightCopy[i].Delay {
			return false
		}
	}
	return true
}

func normalizeModelRef(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

func splitCSVModels(input string) []string {
	parts := strings.Split(input, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := normalizeModelRef(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func normalizeAPIKeySupportedModels(input string) string {
	parts := splitCSVModels(input)
	if len(parts) == 0 {
		return ""
	}
	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		out = append(out, part)
	}
	sort.Strings(out)
	return strings.Join(out, ",")
}

func sliceContainsSubstring(items []string, want string) bool {
	for _, item := range items {
		if strings.Contains(item, want) {
			return true
		}
	}
	return false
}

func filterImportableChannelKeys(rows []model.ChannelKey) ([]model.ChannelKey, int) {
	if len(rows) == 0 {
		return nil, 0
	}
	out := make([]model.ChannelKey, 0, len(rows))
	skipped := 0
	for _, row := range rows {
		if strings.TrimSpace(row.ChannelKey) == "" {
			skipped++
			continue
		}
		row.SourceType = model.NormalizeChannelKeySourceType(row.SourceType)
		row.AllowedModels = model.NormalizeChannelKeyAllowedModels(row.AllowedModels)
		out = append(out, row)
	}
	return out, skipped
}

func filterImportableAPIKeys(rows []model.APIKey) ([]model.APIKey, int) {
	if len(rows) == 0 {
		return nil, 0
	}
	out := make([]model.APIKey, 0, len(rows))
	skipped := 0
	for _, row := range rows {
		if strings.TrimSpace(row.APIKey) == "" {
			skipped++
			continue
		}
		row.SupportedModels = normalizeAPIKeySupportedModels(row.SupportedModels)
		out = append(out, row)
	}
	return out, skipped
}

func importChannels(tx *gorm.DB, rows []model.Channel, mode model.DBImportMode, legacyHints *model.DBDumpLegacyHints) (map[int]int, int64, error) {
	resultMap := map[int]int{}
	var affected int64
	for _, row := range rows {
		snapshotID := row.ID
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		row.ID = 0
		row.Name = name
		row.ChannelProxy = model.NormalizeChannelProxy(row.ChannelProxy)
		if err := model.ValidateChannelProxy(row.ChannelProxy); err != nil {
			return nil, 0, fmt.Errorf("invalid channel %s proxy: %w", name, err)
		}
		row.KeyManagementMode = model.NormalizeKeyManagementMode(row.KeyManagementMode)
		if !model.IsValidKeyManagementMode(row.KeyManagementMode) {
			row.KeyManagementMode = model.KeyManagementModePooled
		}
		row.KeyRoutingPolicy = model.NormalizeKeyRoutingPolicy(row.KeyRoutingPolicy)
		if !model.IsValidKeyRoutingPolicy(row.KeyRoutingPolicy) {
			row.KeyRoutingPolicy = model.KeyRoutingPolicyRoundRobin
		}
		var existing model.Channel
		err := tx.Where("name = ?", name).First(&existing).Error
		switch {
		case err == nil:
			resultMap[snapshotID] = existing.ID
			if mode == model.DBImportModeSkip {
				continue
			}
			if legacyHints != nil {
				if hint, ok := legacyHints.ChannelsByName[name]; ok {
					if hint.MissingKeyManagementMode {
						row.KeyManagementMode = existing.KeyManagementMode
					}
					if hint.MissingKeyRoutingPolicy {
						row.KeyRoutingPolicy = existing.KeyRoutingPolicy
					}
				}
			}
			updates := model.Channel{
				Type:              row.Type,
				Enabled:           row.Enabled,
				KeyManagementMode: row.KeyManagementMode,
				KeyRoutingPolicy:  row.KeyRoutingPolicy,
				BaseUrls:          row.BaseUrls,
				Model:             row.Model,
				CustomModel:       row.CustomModel,
				Proxy:             row.Proxy,
				AutoSync:          row.AutoSync,
				AutoGroup:         row.AutoGroup,
				CustomHeader:      row.CustomHeader,
				ParamOverride:     row.ParamOverride,
				ChannelProxy:      row.ChannelProxy,
				MatchRegex:        row.MatchRegex,
			}
			if err := tx.Model(&model.Channel{}).Where("id = ?", existing.ID).Select("Type", "Enabled", "KeyManagementMode", "KeyRoutingPolicy", "BaseUrls", "Model", "CustomModel", "Proxy", "AutoSync", "AutoGroup", "CustomHeader", "ParamOverride", "ChannelProxy", "MatchRegex").Updates(&updates).Error; err != nil {
				return nil, 0, err
			}
			affected++
		case err == gorm.ErrRecordNotFound:
			enabled := row.Enabled
			if err := tx.Select("Name", "Type", "Enabled", "KeyManagementMode", "KeyRoutingPolicy", "BaseUrls", "Model", "CustomModel", "Proxy", "AutoSync", "AutoGroup", "CustomHeader", "ParamOverride", "ChannelProxy", "MatchRegex").Create(&row).Error; err != nil {
				return nil, 0, err
			}
			if err := tx.Model(&model.Channel{}).Where("id = ?", row.ID).Update("enabled", enabled).Error; err != nil {
				return nil, 0, err
			}
			resultMap[snapshotID] = row.ID
			affected++
		default:
			return nil, 0, err
		}
	}
	return resultMap, affected, nil
}

func replaceChannels(tx *gorm.DB, rows []model.Channel) (int64, error) {
	keepNames := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		keepNames = append(keepNames, name)
	}
	query := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&model.Channel{})
	if len(keepNames) > 0 {
		query = query.Where("name NOT IN ?", keepNames)
	}
	var staleChannels []model.Channel
	if err := query.Find(&staleChannels).Error; err != nil {
		return 0, err
	}
	if len(staleChannels) == 0 {
		return 0, nil
	}
	channelIDs := make([]int, 0, len(staleChannels))
	for _, row := range staleChannels {
		channelIDs = append(channelIDs, row.ID)
	}
	if err := tx.Where("channel_id IN ?", channelIDs).Delete(&model.GroupItem{}).Error; err != nil {
		return 0, err
	}
	if err := tx.Where("channel_id IN ?", channelIDs).Delete(&model.RouteTargetOverride{}).Error; err != nil {
		return 0, err
	}
	if err := tx.Where("channel_id IN ?", channelIDs).Delete(&model.ChannelKey{}).Error; err != nil {
		return 0, err
	}
	result := tx.Where("id IN ?", channelIDs).Delete(&model.Channel{})
	return result.RowsAffected, result.Error
}

func replaceChannelKeys(tx *gorm.DB, channelIDMap map[int]int) (int64, error) {
	channelIDs := make([]int, 0, len(channelIDMap))
	seen := make(map[int]struct{}, len(channelIDMap))
	for _, id := range channelIDMap {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		channelIDs = append(channelIDs, id)
	}
	if len(channelIDs) == 0 {
		return 0, nil
	}
	if err := tx.Where("channel_id IN ?", channelIDs).Delete(&model.RouteTargetOverride{}).Error; err != nil {
		return 0, err
	}
	result := tx.Where("channel_id IN ?", channelIDs).Delete(&model.ChannelKey{})
	return result.RowsAffected, result.Error
}

func importGroups(tx *gorm.DB, rows []model.Group, mode model.DBImportMode, legacyHints *model.DBDumpLegacyHints) (map[int]int, int64, error) {
	resultMap := map[int]int{}
	var affected int64
	for _, row := range rows {
		snapshotID := row.ID
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		row.ID = 0
		row.Name = name
		var existing model.Group
		err := tx.Where("name = ?", name).First(&existing).Error
		switch {
		case err == nil:
			resultMap[snapshotID] = existing.ID
			if mode == model.DBImportModeSkip {
				continue
			}
			if legacyHints != nil {
				if hint, ok := legacyHints.GroupsByName[name]; ok {
					if hint.MissingRetryRounds {
						row.RetryRounds = existing.RetryRounds
					}
					if hint.MissingRetryDelayMs {
						row.RetryDelayMs = existing.RetryDelayMs
					}
					if hint.MissingFailoverWindowSec {
						row.FailoverWindowSec = existing.FailoverWindowSec
					}
					if hint.MissingRaceAfterFails {
						row.RaceAfterFails = existing.RaceAfterFails
					}
					if hint.MissingRaceConcurrency {
						row.RaceConcurrency = existing.RaceConcurrency
					}
				}
			}
			updates := model.Group{
				Mode:              row.Mode,
				MatchRegex:        row.MatchRegex,
				FirstTokenTimeOut: row.FirstTokenTimeOut,
				SessionKeepTime:   row.SessionKeepTime,
				RetryRounds:       row.RetryRounds,
				RetryDelayMs:      row.RetryDelayMs,
				FailoverWindowSec: row.FailoverWindowSec,
				RaceAfterFails:    row.RaceAfterFails,
				RaceConcurrency:   row.RaceConcurrency,
			}
			if err := tx.Model(&model.Group{}).Where("id = ?", existing.ID).Select("Mode", "MatchRegex", "FirstTokenTimeOut", "SessionKeepTime", "RetryRounds", "RetryDelayMs", "FailoverWindowSec", "RaceAfterFails", "RaceConcurrency").Updates(&updates).Error; err != nil {
				return nil, 0, err
			}
			affected++
		case err == gorm.ErrRecordNotFound:
			if err := tx.Create(&row).Error; err != nil {
				return nil, 0, err
			}
			resultMap[snapshotID] = row.ID
			affected++
		default:
			return nil, 0, err
		}
	}
	return resultMap, affected, nil
}

func replaceGroups(tx *gorm.DB, rows []model.Group) (int64, error) {
	keepNames := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		keepNames = append(keepNames, name)
	}
	query := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&model.Group{})
	if len(keepNames) > 0 {
		query = query.Where("name NOT IN ?", keepNames)
	}
	var staleGroups []model.Group
	if err := query.Find(&staleGroups).Error; err != nil {
		return 0, err
	}
	if len(staleGroups) == 0 {
		return 0, nil
	}
	groupIDs := make([]int, 0, len(staleGroups))
	for _, row := range staleGroups {
		groupIDs = append(groupIDs, row.ID)
	}
	if err := tx.Where("group_id IN ?", groupIDs).Delete(&model.GroupItem{}).Error; err != nil {
		return 0, err
	}
	result := tx.Where("id IN ?", groupIDs).Delete(&model.Group{})
	return result.RowsAffected, result.Error
}

func replaceGroupItems(tx *gorm.DB, groupIDMap map[int]int) (int64, error) {
	groupIDs := make([]int, 0, len(groupIDMap))
	seen := make(map[int]struct{}, len(groupIDMap))
	for _, id := range groupIDMap {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		groupIDs = append(groupIDs, id)
	}
	if len(groupIDs) == 0 {
		return 0, nil
	}
	result := tx.Where("group_id IN ?", groupIDs).Delete(&model.GroupItem{})
	return result.RowsAffected, result.Error
}

func prepareChannelKeysForImport(tx *gorm.DB, rows []model.ChannelKey, dumpChannels []model.Channel, channelIDMap map[int]int, state *currentImportState, mode model.DBImportMode, legacyHints *model.DBDumpLegacyHints) ([]preparedChannelKeyImport, []string, error) {
	_ = tx
	prepared := make([]preparedChannelKeyImport, 0, len(rows))
	warnings := make([]string, 0)
	channelNamesBySnapshotID := make(map[int]string, len(dumpChannels))
	for _, row := range dumpChannels {
		channelNamesBySnapshotID[row.ID] = strings.TrimSpace(row.Name)
	}
	for _, row := range rows {
		snapshotID := row.ID
		localChannelID, ok := channelIDMap[row.ChannelID]
		if !ok || localChannelID <= 0 {
			warnings = append(warnings, fmt.Sprintf("skipped channel_key:%d because channel_id:%d could not be resolved after import", snapshotID, row.ChannelID))
			continue
		}
		if mode == model.DBImportModeSkip {
			if _, exists := state.channelsByName[channelNamesBySnapshotID[row.ChannelID]]; exists {
				continue
			}
		}
		if legacyHints != nil {
			if hint, ok := legacyHints.ChannelKeysBySnapshotID[snapshotID]; ok {
				channelName := channelNamesBySnapshotID[row.ChannelID]
				if existingChannel, ok := state.channelsByName[channelName]; ok {
					for _, existingKey := range existingChannel.Keys {
						if strings.TrimSpace(existingKey.ChannelKey) != strings.TrimSpace(row.ChannelKey) {
							continue
						}
						if hint.MissingSourceType {
							row.SourceType = existingKey.SourceType
						}
						if hint.MissingAllowedModels {
							row.AllowedModels = existingKey.AllowedModels
						}
						break
					}
				}
			}
		}
		row.ID = 0
		row.ChannelID = localChannelID
		row.SourceType = model.NormalizeChannelKeySourceType(row.SourceType)
		row.AllowedModels = model.NormalizeChannelKeyAllowedModels(row.AllowedModels)
		prepared = append(prepared, preparedChannelKeyImport{SnapshotID: snapshotID, Row: row})
	}
	return prepared, warnings, nil
}

func importPreparedChannelKeys(tx *gorm.DB, rows []preparedChannelKeyImport, mode model.DBImportMode) (int64, map[int]int, error) {
	idMap := map[int]int{}
	var affected int64
	for _, prepared := range rows {
		row := prepared.Row
		if mode == model.DBImportModeSkip {
			if err := tx.Create(&row).Error; err != nil {
				return 0, nil, err
			}
			idMap[prepared.SnapshotID] = row.ID
			affected++
			continue
		}
		var existing model.ChannelKey
		err := tx.Where("channel_id = ? AND channel_key = ?", row.ChannelID, row.ChannelKey).First(&existing).Error
		switch {
		case err == nil:
			updates := model.ChannelKey{
				Enabled:       row.Enabled,
				SourceType:    row.SourceType,
				Remark:        row.Remark,
				AllowedModels: row.AllowedModels,
			}
			if err := tx.Model(&model.ChannelKey{}).Where("id = ?", existing.ID).Select("Enabled", "SourceType", "Remark", "AllowedModels").Updates(&updates).Error; err != nil {
				return 0, nil, err
			}
			idMap[prepared.SnapshotID] = existing.ID
			affected++
		case err == gorm.ErrRecordNotFound:
			enabled := row.Enabled
			if err := tx.Select("Enabled", "ChannelKey", "SourceType", "Remark", "AllowedModels", "ChannelID").Create(&row).Error; err != nil {
				return 0, nil, err
			}
			if err := tx.Model(&model.ChannelKey{}).Where("id = ?", row.ID).Update("enabled", enabled).Error; err != nil {
				return 0, nil, err
			}
			idMap[prepared.SnapshotID] = row.ID
			affected++
		default:
			return 0, nil, err
		}
	}
	return affected, idMap, nil
}

func prepareRouteTargetOverridesForImport(tx *gorm.DB, rows []model.RouteTargetOverride, dumpChannels []model.Channel, dumpChannelKeys []model.ChannelKey, channelIDMap map[int]int, channelKeyIDMap map[int]int, state *currentImportState, mode model.DBImportMode) ([]preparedRouteTargetOverrideImport, []string, error) {
	_ = tx
	prepared := make([]preparedRouteTargetOverrideImport, 0, len(rows))
	warnings := make([]string, 0)
	channelNamesBySnapshotID := make(map[int]string, len(dumpChannels))
	for _, row := range dumpChannels {
		channelNamesBySnapshotID[row.ID] = strings.TrimSpace(row.Name)
	}
	for _, row := range rows {
		snapshotID := row.ID
		localChannelID, ok := channelIDMap[row.ChannelID]
		if !ok || localChannelID <= 0 {
			warnings = append(warnings, fmt.Sprintf("skipped route_target_override:%d because channel_id:%d could not be resolved after import", snapshotID, row.ChannelID))
			continue
		}
		localChannelKeyID, ok := channelKeyIDMap[row.ChannelKeyID]
		if !ok || localChannelKeyID <= 0 {
			warnings = append(warnings, fmt.Sprintf("skipped route_target_override:%d because channel_key_id:%d could not be resolved after import", snapshotID, row.ChannelKeyID))
			continue
		}
		if mode == model.DBImportModeSkip {
			if _, exists := state.channelsByName[channelNamesBySnapshotID[row.ChannelID]]; exists {
				continue
			}
		}
		row.ID = 0
		row.ChannelID = localChannelID
		row.ChannelKeyID = localChannelKeyID
		row.ModelName = normalizeModelRef(row.ModelName)
		prepared = append(prepared, preparedRouteTargetOverrideImport{SnapshotID: snapshotID, Row: row})
	}
	return prepared, warnings, nil
}

func importPreparedRouteTargetOverrides(tx *gorm.DB, rows []preparedRouteTargetOverrideImport, mode model.DBImportMode) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	var affected int64
	for _, prepared := range rows {
		row := prepared.Row
		if mode == model.DBImportModeSkip {
			result := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "channel_id"}, {Name: "channel_key_id"}, {Name: "model_name"}},
				DoNothing: true,
			}).Create(&row)
			if result.Error != nil {
				return 0, result.Error
			}
			affected += result.RowsAffected
			continue
		}
		result := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "channel_id"}, {Name: "channel_key_id"}, {Name: "model_name"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"billing_mode",
				"probe_policy",
				"probe_interval_seconds",
				"probe_concurrency_limit",
			}),
		}).Create(&row)
		if result.Error != nil {
			return 0, result.Error
		}
		affected += result.RowsAffected
	}
	return affected, nil
}

func remapGroupItemsForImport(rows []model.GroupItem, dumpGroups []model.Group, dumpChannels []model.Channel, groupIDMap map[int]int, channelIDMap map[int]int, state *currentImportState, mode model.DBImportMode) ([]model.GroupItem, []string) {
	if len(rows) == 0 {
		return nil, nil
	}
	groupNamesBySnapshotID := make(map[int]string, len(dumpGroups))
	for _, row := range dumpGroups {
		groupNamesBySnapshotID[row.ID] = strings.TrimSpace(row.Name)
	}
	channelsBySnapshotID := make(map[int]model.Channel, len(dumpChannels))
	for _, row := range dumpChannels {
		channelsBySnapshotID[row.ID] = row
	}
	out := make([]model.GroupItem, 0, len(rows))
	warnings := make([]string, 0)
	for _, row := range rows {
		if mode == model.DBImportModeSkip {
			if _, exists := state.groupsByName[groupNamesBySnapshotID[row.GroupID]]; exists {
				continue
			}
		}
		localGroupID, ok := groupIDMap[row.GroupID]
		if !ok || localGroupID <= 0 {
			warnings = append(warnings, fmt.Sprintf("skipped group_item:%d because group_id:%d could not be resolved after import", row.ID, row.GroupID))
			continue
		}
		localChannelID, ok := channelIDMap[row.ChannelID]
		if !ok || localChannelID <= 0 {
			warnings = append(warnings, fmt.Sprintf("skipped group_item:%d because channel_id:%d could not be resolved after import", row.ID, row.ChannelID))
			continue
		}
		channel := channelsBySnapshotID[row.ChannelID]
		if !channel.SupportsModel(row.ModelName) {
			warnings = append(warnings, fmt.Sprintf("skipped group_item:%d because channel:%s does not declare model:%s", row.ID, channel.Name, row.ModelName))
			continue
		}
		row.ID = 0
		row.GroupID = localGroupID
		row.ChannelID = localChannelID
		row.ModelName = normalizeModelRef(row.ModelName)
		out = append(out, row)
	}
	return out, warnings
}

func buildImportedHealthCheckTargets(dumpGroups []model.Group, groupIDMap map[int]int, rows []model.GroupItem) []ChannelModelHealthCheckTarget {
	if len(rows) == 0 {
		return nil
	}
	groupNamesByLocalID := make(map[int]string, len(groupIDMap))
	for _, row := range dumpGroups {
		if localID, ok := groupIDMap[row.ID]; ok && localID > 0 {
			groupNamesByLocalID[localID] = strings.TrimSpace(row.Name)
		}
	}
	targets := make([]ChannelModelHealthCheckTarget, 0, len(rows))
	for _, row := range rows {
		targets = append(targets, ChannelModelHealthCheckTarget{
			GroupID:   row.GroupID,
			GroupName: groupNamesByLocalID[row.GroupID],
			ChannelID: row.ChannelID,
			ModelName: row.ModelName,
		})
	}
	return dedupeAndSortHealthCheckTargets(targets)
}

func importGroupItems(tx *gorm.DB, rows []model.GroupItem, mode model.DBImportMode) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	if mode == model.DBImportModeSkip {
		return createDoNothingOnColumns(tx, rows, []clause.Column{{Name: "group_id"}, {Name: "channel_id"}, {Name: "model_name"}})
	}
	return createUpsertAll(tx, rows, []clause.Column{{Name: "group_id"}, {Name: "channel_id"}, {Name: "model_name"}})
}

func prepareAPIKeysForImport(tx *gorm.DB, rows []model.APIKey) ([]preparedAPIKeyImport, []string, error) {
	_ = tx
	prepared := make([]preparedAPIKeyImport, 0, len(rows))
	for _, row := range rows {
		snapshotID := row.ID
		row.ID = 0
		row.SupportedModels = normalizeAPIKeySupportedModels(row.SupportedModels)
		prepared = append(prepared, preparedAPIKeyImport{SnapshotID: snapshotID, Row: row})
	}
	return prepared, nil, nil
}

func importPreparedAPIKeys(tx *gorm.DB, rows []preparedAPIKeyImport, mode model.DBImportMode) (int64, map[int]int, error) {
	idMap := map[int]int{}
	var affected int64
	for _, prepared := range rows {
		row := prepared.Row
		if mode == model.DBImportModeSkip {
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
			if result.Error != nil {
				return 0, nil, result.Error
			}
			if row.ID != 0 {
				idMap[prepared.SnapshotID] = row.ID
			}
			affected += result.RowsAffected
			continue
		}
		var existing model.APIKey
		err := tx.Where("api_key = ?", row.APIKey).First(&existing).Error
		switch {
		case err == nil:
			updates := model.APIKey{
				Name:            row.Name,
				Enabled:         row.Enabled,
				ExpireAt:        row.ExpireAt,
				MaxCost:         row.MaxCost,
				SupportedModels: row.SupportedModels,
			}
			if err := tx.Model(&model.APIKey{}).Where("id = ?", existing.ID).Select("Name", "Enabled", "ExpireAt", "MaxCost", "SupportedModels").Updates(&updates).Error; err != nil {
				return 0, nil, err
			}
			idMap[prepared.SnapshotID] = existing.ID
			affected++
		case err == gorm.ErrRecordNotFound:
			enabled := row.Enabled
			if err := tx.Select("Name", "Enabled", "ExpireAt", "MaxCost", "SupportedModels", "APIKey").Create(&row).Error; err != nil {
				return 0, nil, err
			}
			if err := tx.Model(&model.APIKey{}).Where("id = ?", row.ID).Update("enabled", enabled).Error; err != nil {
				return 0, nil, err
			}
			idMap[prepared.SnapshotID] = row.ID
			affected++
		default:
			return 0, nil, err
		}
	}
	return affected, idMap, nil
}

func importLLMInfos(tx *gorm.DB, rows []model.LLMInfo, state *currentImportState, legacyHints *model.DBDumpLegacyHints) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	prepared := make([]model.LLMInfo, 0, len(rows))
	for _, row := range rows {
		name := normalizeModelRef(row.Name)
		if name == "" {
			continue
		}
		row.Name = name
		row.CanonicalName = normalizeModelRef(row.CanonicalName)
		if row.CanonicalName == "" {
			row.CanonicalName = llmname.CanonicalModelName(row.Name)
		}
		row.BillingMode = model.NormalizeBillingMode(row.BillingMode)
		row.ProbePolicy = model.NormalizeProbePolicy(row.ProbePolicy)
		if legacyHints != nil {
			if hint, ok := legacyHints.LLMInfosByName[name]; ok {
				if current, found := findCurrentLLMForSnapshot(row, state); found {
					if hint.MissingCanonicalName {
						row.CanonicalName = current.CanonicalName
						if row.CanonicalName == "" {
							row.CanonicalName = llmname.CanonicalModelName(current.Name)
						}
					}
					if hint.MissingBillingMode {
						row.BillingMode = current.BillingMode
					}
					if hint.MissingProbePolicy {
						row.ProbePolicy = current.ProbePolicy
					}
					if hint.MissingProbeIntervalSeconds {
						row.ProbeIntervalSeconds = current.ProbeIntervalSeconds
					}
					if hint.MissingProbeConcurrencyLimit {
						row.ProbeConcurrencyLimit = current.ProbeConcurrencyLimit
					}
				}
			}
		}
		prepared = append(prepared, row)
	}
	if len(prepared) == 0 {
		return 0, nil
	}
	result := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"input",
			"output",
			"cache_read",
			"cache_write",
			"official_input",
			"official_output",
			"official_cache_read",
			"official_cache_write",
			"canonical_name",
			"billing_mode",
			"probe_policy",
			"probe_interval_seconds",
			"probe_concurrency_limit",
		}),
	}).Create(&prepared)
	return result.RowsAffected, result.Error
}

func importUsers(tx *gorm.DB, rows []model.User, mode model.DBImportMode) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	var affected int64
	for _, row := range rows {
		row.Username = strings.TrimSpace(row.Username)
		if row.Username == "" || strings.TrimSpace(row.Password) == "" {
			continue
		}
		row.ID = 0
		if mode == model.DBImportModeSkip {
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
			if result.Error != nil {
				return 0, result.Error
			}
			affected += result.RowsAffected
			continue
		}
		var existing model.User
		err := tx.Where("username = ?", row.Username).First(&existing).Error
		switch {
		case err == nil:
			if err := tx.Model(&model.User{}).Where("id = ?", existing.ID).Select("Password").Updates(&model.User{Password: row.Password}).Error; err != nil {
				return 0, err
			}
			affected++
		case err == gorm.ErrRecordNotFound:
			if err := tx.Select("Username", "Password").Create(&row).Error; err != nil {
				return 0, err
			}
			affected++
		default:
			return 0, err
		}
	}
	return affected, nil
}

func remapStatsModelsForImport(rows []model.StatsModel, channelIDMap map[int]int) ([]model.StatsModel, []string) {
	if len(rows) == 0 {
		return nil, nil
	}
	out := make([]model.StatsModel, 0, len(rows))
	warnings := make([]string, 0)
	for _, row := range rows {
		localChannelID, ok := channelIDMap[row.ChannelID]
		if !ok || localChannelID <= 0 {
			warnings = append(warnings, fmt.Sprintf("skipped stats_model for snapshot channel_id=%d because it could not be resolved after import", row.ChannelID))
			continue
		}
		row.ChannelID = localChannelID
		out = append(out, row)
	}
	return out, warnings
}

func remapStatsChannelsForImport(rows []model.StatsChannel, channelIDMap map[int]int) ([]model.StatsChannel, []string) {
	if len(rows) == 0 {
		return nil, nil
	}
	out := make([]model.StatsChannel, 0, len(rows))
	warnings := make([]string, 0)
	for _, row := range rows {
		localChannelID, ok := channelIDMap[row.ChannelID]
		if !ok || localChannelID <= 0 {
			warnings = append(warnings, fmt.Sprintf("skipped stats_channel for snapshot channel_id=%d because it could not be resolved after import", row.ChannelID))
			continue
		}
		row.ChannelID = localChannelID
		out = append(out, row)
	}
	return out, warnings
}

func remapStatsAPIKeysForImport(rows []model.StatsAPIKey, apiKeyIDMap map[int]int) ([]model.StatsAPIKey, []string) {
	if len(rows) == 0 {
		return nil, nil
	}
	out := make([]model.StatsAPIKey, 0, len(rows))
	warnings := make([]string, 0)
	for _, row := range rows {
		localAPIKeyID, ok := apiKeyIDMap[row.APIKeyID]
		if !ok || localAPIKeyID <= 0 {
			warnings = append(warnings, fmt.Sprintf("skipped stats_api_key for snapshot api_key_id=%d because it could not be resolved after import", row.APIKeyID))
			continue
		}
		row.APIKeyID = localAPIKeyID
		out = append(out, row)
	}
	return out, warnings
}

func remapRelayLogsForImport(rows []model.RelayLog, channelIDMap map[int]int, channelKeyIDMap map[int]int) ([]model.RelayLog, []string) {
	if len(rows) == 0 {
		return nil, nil
	}
	out := make([]model.RelayLog, 0, len(rows))
	warnings := make([]string, 0)
	resetCount := 0
	for _, row := range rows {
		localChannelID, ok := channelIDMap[row.ChannelId]
		if !ok || localChannelID <= 0 {
			warnings = append(warnings, fmt.Sprintf("skipped relay_log:%d because channel_id:%d could not be resolved after import", row.ID, row.ChannelId))
			continue
		}
		row.ChannelId = localChannelID
		for i := range row.Attempts {
			if mappedChannelID, ok := channelIDMap[row.Attempts[i].ChannelID]; ok && mappedChannelID > 0 {
				row.Attempts[i].ChannelID = mappedChannelID
			}
			if row.Attempts[i].ChannelKeyID == 0 {
				continue
			}
			if mappedKeyID, ok := channelKeyIDMap[row.Attempts[i].ChannelKeyID]; ok && mappedKeyID > 0 {
				row.Attempts[i].ChannelKeyID = mappedKeyID
				continue
			}
			row.Attempts[i].ChannelKeyID = 0
			resetCount++
		}
		out = append(out, row)
	}
	if resetCount > 0 {
		warnings = append(warnings, fmt.Sprintf("reset %d relay log attempt channel key references that could not be resolved after import", resetCount))
	}
	return out, warnings
}

func createDoNothing[T any](tx *gorm.DB, rows []T) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows)
	return result.RowsAffected, result.Error
}

func createDoNothingOnColumns[T any](tx *gorm.DB, rows []T, columns []clause.Column) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	result := tx.Clauses(clause.OnConflict{Columns: columns, DoNothing: true}).Create(&rows)
	return result.RowsAffected, result.Error
}

func createUpsertAll[T any](tx *gorm.DB, rows []T, columns []clause.Column) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	result := tx.Clauses(clause.OnConflict{Columns: columns, UpdateAll: true}).Create(&rows)
	return result.RowsAffected, result.Error
}

func createUpsertSettings(tx *gorm.DB, rows []model.Setting) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	result := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&rows)
	return result.RowsAffected, result.Error
}

func replaceSettings(tx *gorm.DB, rows []model.Setting) (int64, error) {
	keepKeys := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(string(row.Key))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keepKeys = append(keepKeys, key)
	}
	query := tx.Session(&gorm.Session{AllowGlobalUpdate: true})
	if len(keepKeys) > 0 {
		query = query.Where("key NOT IN ?", keepKeys)
	}
	result := query.Delete(&model.Setting{})
	return result.RowsAffected, result.Error
}

func replaceLLMInfos(tx *gorm.DB, rows []model.LLMInfo) (int64, error) {
	keepNames := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		keepNames = append(keepNames, name)
	}
	query := tx.Session(&gorm.Session{AllowGlobalUpdate: true})
	if len(keepNames) > 0 {
		query = query.Where("name NOT IN ?", keepNames)
	}
	result := query.Delete(&model.LLMInfo{})
	return result.RowsAffected, result.Error
}

func replaceAPIKeys(tx *gorm.DB, rows []model.APIKey) (int64, error) {
	keepKeys := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		apiKey := strings.TrimSpace(row.APIKey)
		if apiKey == "" {
			continue
		}
		if _, ok := seen[apiKey]; ok {
			continue
		}
		seen[apiKey] = struct{}{}
		keepKeys = append(keepKeys, apiKey)
	}
	query := tx.Session(&gorm.Session{AllowGlobalUpdate: true})
	if len(keepKeys) > 0 {
		query = query.Where("api_key NOT IN ?", keepKeys)
	}
	result := query.Delete(&model.APIKey{})
	return result.RowsAffected, result.Error
}

func replaceMigrationRecords(tx *gorm.DB, rows []model.DBDumpMigrationRecord) (int64, error) {
	keepVersions := make([]int, 0, len(rows))
	seen := make(map[int]struct{}, len(rows))
	for _, row := range rows {
		if row.Version <= 0 {
			continue
		}
		if _, ok := seen[row.Version]; ok {
			continue
		}
		seen[row.Version] = struct{}{}
		keepVersions = append(keepVersions, row.Version)
	}
	query := tx.Session(&gorm.Session{AllowGlobalUpdate: true})
	if len(keepVersions) > 0 {
		query = query.Where("version NOT IN ?", keepVersions)
	}
	result := query.Delete(&migrate.MigrationRecord{})
	return result.RowsAffected, result.Error
}

func importMigrationRecords(tx *gorm.DB, rows []model.DBDumpMigrationRecord, mode model.DBImportMode) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	records := make([]migrate.MigrationRecord, 0, len(rows))
	for _, row := range rows {
		if row.Version <= 0 {
			continue
		}
		records = append(records, migrate.MigrationRecord{
			Version: row.Version,
			Status:  migrate.MigrationRecordStatus(row.Status),
		})
	}
	if len(records) == 0 {
		return 0, nil
	}
	if mode == model.DBImportModeSkip {
		return createDoNothingOnColumns(tx, records, []clause.Column{{Name: "version"}})
	}
	result := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "version"}},
		DoUpdates: clause.AssignmentColumns([]string{"status"}),
	}).Create(&records)
	return result.RowsAffected, result.Error
}

func buildImportCompatibility(originalDump, dump *model.DBDump, state *currentImportState, mode model.DBImportMode, scopes *model.DBImportScopes, modelMappings map[string]string) *model.DBImportCompatibilityReport {
	report := &model.DBImportCompatibilityReport{Summary: &model.DBImportCompatibilitySummary{}}
	if dump == nil {
		return report
	}
	routingScopeEnabled := scopes == nil || scopes.Routing
	settingsScopeEnabled := scopes == nil || scopes.Settings
	modelsScopeEnabled := scopes == nil || scopes.Models
	apiKeysScopeEnabled := scopes == nil || scopes.APIKeys

	for _, row := range dump.ChannelKeys {
		if strings.TrimSpace(row.ChannelKey) == "" {
			report.SkippedTargets = append(report.SkippedTargets, fmt.Sprintf("channel_key:%d empty credential", row.ID))
		}
	}
	for _, row := range dump.APIKeys {
		if strings.TrimSpace(row.APIKey) == "" {
			report.SkippedTargets = append(report.SkippedTargets, fmt.Sprintf("api_key:%d empty credential", row.ID))
		}
	}
	if schema := strings.TrimSpace(dump.Manifest.SchemaVersion); schema != "" && schema != "v1" {
		report.SchemaMismatches = append(report.SchemaMismatches, fmt.Sprintf("snapshot schema:%s differs", schema))
	}
	channelNamesBySnapshotID := make(map[int]string, len(dump.Channels))
	groupNamesBySnapshotID := make(map[int]string, len(dump.Groups))
	if routingScopeEnabled {
		for _, row := range dump.Channels {
			name := strings.TrimSpace(row.Name)
			channelNamesBySnapshotID[row.ID] = name
			if existing, ok := state.channelsByName[name]; !ok {
				report.MissingProviders = append(report.MissingProviders, name)
			} else if !sameBaseURLs(existing.BaseUrls, row.BaseUrls) {
				report.BaseURLMismatches = append(report.BaseURLMismatches, name)
			}
			if mode == model.DBImportModeSkip {
				if _, ok := state.channelsByName[name]; ok {
					report.SkippedTargets = append(report.SkippedTargets, fmt.Sprintf("channel:%s existing row preserved by skip mode", name))
				}
			}
		}
		for _, row := range dump.Groups {
			name := strings.TrimSpace(row.Name)
			groupNamesBySnapshotID[row.ID] = name
		}
	}
	if settingsScopeEnabled {
		for _, row := range dump.Settings {
			if mode == model.DBImportModeSkip {
				if _, ok := state.settingsByKey[strings.TrimSpace(string(row.Key))]; ok {
					report.SkippedTargets = append(report.SkippedTargets, fmt.Sprintf("setting:%s existing row preserved by skip mode", row.Key))
				}
			}
		}
	}
	if routingScopeEnabled || modelsScopeEnabled || apiKeysScopeEnabled {
		modelContexts := collectDumpModelContexts(dump)
		for modelName, contexts := range modelContexts {
			if _, ok := state.llmInfosByName[modelName]; ok {
				continue
			}
			if aliasTarget, ok := findAliasTarget(modelName, dump, state); ok {
				report.AliasConflicts = append(report.AliasConflicts, modelName)
				report.AliasPreviewMappings = append(report.AliasPreviewMappings, model.DBImportAliasPreviewMapping{
					SnapshotModel: modelName,
					CurrentModel:  aliasTarget.Name,
					Canonical:     aliasTarget.CanonicalName,
					Contexts:      append([]string(nil), contexts...),
				})
				continue
			}
			report.MissingModels = append(report.MissingModels, modelName)
		}
	}
	report.ModelMappingPreviews = buildModelMappingPreviews(originalDump, modelMappings, state)
	report.CredentialRebindTargets = buildCredentialRebindTargets(dump)
	if modelsScopeEnabled {
		report.ModelPolicyDiffs = buildModelPolicyDiffs(dump, state)
	}
	if routingScopeEnabled && mode == model.DBImportModeSkip {
		for _, row := range dump.GroupItems {
			if _, ok := state.groupsByName[groupNamesBySnapshotID[row.GroupID]]; ok {
				report.RouteConflicts = append(report.RouteConflicts, groupNamesBySnapshotID[row.GroupID])
			}
		}
	}
	if routingScopeEnabled {
		report.RoutePreviewDiffs = buildRoutePreviewDiffs(dump, state, mode)
		report.InvalidRouteTargets, report.SkippedRoutePreviews = buildRoutePreviewIssues(report.RoutePreviewDiffs)
		if len(report.RoutePreviewDiffs) > 0 {
			report.RoutePreviewWarnings = append(report.RoutePreviewWarnings, fmt.Sprintf("route preview diffs: %d", len(report.RoutePreviewDiffs)))
		}
		if len(report.InvalidRouteTargets) > 0 {
			report.RoutePreviewWarnings = append(report.RoutePreviewWarnings, fmt.Sprintf("invalid route targets: %d", len(report.InvalidRouteTargets)))
		}
		if len(report.SkippedRoutePreviews) > 0 {
			report.RoutePreviewWarnings = append(report.RoutePreviewWarnings, fmt.Sprintf("skipped route target previews: %d", len(report.SkippedRoutePreviews)))
		}
	}
	if mode == model.DBImportModeReplace {
		if routingScopeEnabled {
			report.ReplacePrunedChannels = buildReplacePrunedChannelNames(dump, state)
			report.ReplacePrunedGroups = buildReplacePrunedGroupNames(dump, state)
		}
		if settingsScopeEnabled {
			report.ReplacePrunedSettings = buildReplacePrunedSettingKeys(dump, state)
		}
		if modelsScopeEnabled {
			report.ReplacePrunedLLMInfos = buildReplacePrunedLLMInfoNames(dump, state)
		}
		if apiKeysScopeEnabled && dump.Manifest.ContainsSecrets {
			keepAPIKeys := dump.APIKeys
			if len(keepAPIKeys) > 0 {
				filteredAPIKeys, _ := filterImportableAPIKeys(dump.APIKeys)
				keepAPIKeys = filteredAPIKeys
			}
			report.ReplacePrunedAPIKeys = buildReplacePrunedAPIKeyNames(state.apiKeysByAPIKey, keepAPIKeys)
		}
		if len(report.ReplacePrunedChannels) > 0 || len(report.ReplacePrunedGroups) > 0 || len(report.ReplacePrunedSettings) > 0 || len(report.ReplacePrunedLLMInfos) > 0 || len(report.ReplacePrunedAPIKeys) > 0 {
			report.ReplacePrunePreview = &model.DBReplacePrunePreview{
				Channels: append([]string(nil), report.ReplacePrunedChannels...),
				Groups:   append([]string(nil), report.ReplacePrunedGroups...),
				Settings: append([]string(nil), report.ReplacePrunedSettings...),
				LLMInfos: append([]string(nil), report.ReplacePrunedLLMInfos...),
				APIKeys:  append([]string(nil), report.ReplacePrunedAPIKeys...),
			}
		}
	}
	for _, items := range [][]string{
		report.MissingProviders,
		report.MissingModels,
		report.AliasConflicts,
		report.RouteConflicts,
		report.BaseURLMismatches,
		report.SchemaMismatches,
		report.SkippedTargets,
		report.ReplacePrunedChannels,
		report.ReplacePrunedGroups,
		report.ReplacePrunedSettings,
		report.ReplacePrunedLLMInfos,
		report.ReplacePrunedAPIKeys,
	} {
		sort.Strings(items)
	}
	for i := range report.AliasPreviewMappings {
		sort.Strings(report.AliasPreviewMappings[i].Contexts)
	}
	report.Summary.MissingProviders = len(report.MissingProviders)
	report.Summary.MissingModels = len(report.MissingModels)
	report.Summary.Conflicts = len(report.Conflicts)
	report.Summary.AliasConflicts = len(report.AliasConflicts)
	report.Summary.CredentialRebindTargets = len(report.CredentialRebindTargets)
	for _, target := range report.CredentialRebindTargets {
		switch target.TargetType {
		case "channel_key":
			report.Summary.ChannelKeyRebindTargets++
		case "api_key":
			report.Summary.APIKeyRebindTargets++
		}
	}
	report.Summary.ModelMappingPreviews = len(report.ModelMappingPreviews)
	for _, preview := range report.ModelMappingPreviews {
		if preview.Used {
			report.Summary.UsedModelMappings++
			if !preview.TargetExists {
				report.Summary.MissingMappingTargets++
			}
			continue
		}
		report.Summary.UnusedModelMappings++
	}
	report.Summary.AliasPreviewMaps = len(report.AliasPreviewMappings)
	report.Summary.ModelPolicyDiffs = len(report.ModelPolicyDiffs)
	report.Summary.RouteConflicts = len(report.RouteConflicts)
	report.Summary.InvalidRouteTargets = len(report.InvalidRouteTargets)
	report.Summary.SkippedRoutePreviews = len(report.SkippedRoutePreviews)
	report.Summary.RoutePreviewDiffs = len(report.RoutePreviewDiffs)
	report.Summary.BaseURLMismatches = len(report.BaseURLMismatches)
	report.Summary.SchemaMismatches = len(report.SchemaMismatches)
	report.Summary.SkippedTargets = len(report.SkippedTargets)
	report.Summary.ReplacePrunedChannels = len(report.ReplacePrunedChannels)
	report.Summary.ReplacePrunedGroups = len(report.ReplacePrunedGroups)
	report.Summary.ReplacePrunedSettings = len(report.ReplacePrunedSettings)
	report.Summary.ReplacePrunedLLMInfos = len(report.ReplacePrunedLLMInfos)
	report.Summary.ReplacePrunedAPIKeys = len(report.ReplacePrunedAPIKeys)
	return report
}

func buildCredentialRebindTargets(dump *model.DBDump) []model.DBImportCredentialRebindTarget {
	if dump == nil {
		return nil
	}
	groupNamesByID := make(map[int]string, len(dump.Groups))
	for _, group := range dump.Groups {
		groupNamesByID[group.ID] = strings.TrimSpace(group.Name)
	}
	channelNamesByID := make(map[int]string, len(dump.Channels))
	for _, channel := range dump.Channels {
		channelNamesByID[channel.ID] = strings.TrimSpace(channel.Name)
	}
	groupRefsByChannelAndModel := collectSnapshotRouteRefsByChannelAndModel(dump, groupNamesByID, channelNamesByID)
	targets := make([]model.DBImportCredentialRebindTarget, 0)
	for _, row := range dump.ChannelKeys {
		if strings.TrimSpace(row.ChannelKey) != "" {
			continue
		}
		models := splitCSVModels(row.AllowedModels)
		if len(models) == 0 {
			if channelName := strings.TrimSpace(channelNamesByID[row.ChannelID]); channelName != "" {
				for _, channel := range dump.Channels {
					if channel.ID != row.ChannelID {
						continue
					}
					models = append(models, splitCSVModels(channel.Model)...)
					models = append(models, splitCSVModels(channel.CustomModel)...)
					break
				}
			}
		}
		models = dedupeSortedStrings(models)
		affectedGroups, contexts := summarizeCredentialRouteRefs(models, groupRefsByChannelAndModel, row.ChannelID)
		warnings := []string{"snapshot does not include this channel key credential; rebind required after import"}
		if len(models) == 0 {
			warnings = append(warnings, "no explicit allowed models found; verify channel-level model coverage before rebind")
		}
		targets = append(targets, model.DBImportCredentialRebindTarget{
			TargetType:     "channel_key",
			SnapshotID:     row.ID,
			ChannelName:    strings.TrimSpace(channelNamesByID[row.ChannelID]),
			KeyName:        firstNonEmpty(strings.TrimSpace(row.Remark), fmt.Sprintf("channel_key:%d", row.ID)),
			SourceType:     strings.TrimSpace(row.SourceType),
			Remark:         strings.TrimSpace(row.Remark),
			Models:         models,
			AffectedGroups: affectedGroups,
			Contexts:       contexts,
			Warnings:       warnings,
		})
	}
	for _, row := range dump.APIKeys {
		if strings.TrimSpace(row.APIKey) != "" {
			continue
		}
		models := dedupeSortedStrings(splitCSVModels(row.SupportedModels))
		warnings := []string{"snapshot does not include this API key credential; rebind required after import"}
		if len(models) == 0 {
			warnings = append(warnings, "API key currently has no supported_models binding in snapshot")
		}
		targets = append(targets, model.DBImportCredentialRebindTarget{
			TargetType: "api_key",
			SnapshotID: row.ID,
			KeyName:    firstNonEmpty(strings.TrimSpace(row.Name), fmt.Sprintf("api_key:%d", row.ID)),
			Models:     models,
			Contexts:   buildAPIKeyRebindContexts(strings.TrimSpace(row.Name), models),
			Warnings:   warnings,
		})
	}
	sort.Slice(targets, func(i, j int) bool {
		if targets[i].TargetType != targets[j].TargetType {
			return targets[i].TargetType < targets[j].TargetType
		}
		if targets[i].ChannelName != targets[j].ChannelName {
			return targets[i].ChannelName < targets[j].ChannelName
		}
		if targets[i].KeyName != targets[j].KeyName {
			return targets[i].KeyName < targets[j].KeyName
		}
		return targets[i].SnapshotID < targets[j].SnapshotID
	})
	return targets
}

func buildRoutePreviewIssues(diffs []model.DBImportRoutePreviewDiff) ([]model.DBImportRoutePreviewIssue, []model.DBImportRoutePreviewIssue) {
	if len(diffs) == 0 {
		return nil, nil
	}
	invalid := make([]model.DBImportRoutePreviewIssue, 0)
	skipped := make([]model.DBImportRoutePreviewIssue, 0)
	for _, diff := range diffs {
		for _, reason := range diff.SkipReasons {
			reason = strings.TrimSpace(reason)
			if reason == "" {
				continue
			}
			if strings.HasPrefix(reason, "skip_mode_preserved_existing_group:") {
				skipped = append(skipped, model.DBImportRoutePreviewIssue{
					GroupName: diff.GroupName,
					Model:     diff.Model,
					IssueType: "skipped_route_target_preview",
					Reason:    reason,
					Action:    "skip_mode_preserved_existing_group",
				})
			}
		}
		for _, candidate := range diff.AfterCandidates {
			if issue, ok := classifyRoutePreviewIssue(diff.GroupName, candidate); ok {
				invalid = append(invalid, issue)
			}
		}
		if len(diff.AfterCandidates) == 0 {
			for _, candidate := range diff.AddedCandidates {
				if issue, ok := classifyRoutePreviewIssue(diff.GroupName, candidate); ok {
					invalid = append(invalid, issue)
				}
			}
		}
	}
	return dedupeRoutePreviewIssues(invalid), dedupeRoutePreviewIssues(skipped)
}

func classifyRoutePreviewIssue(groupName string, candidate model.DBImportRoutePreviewCandidate) (model.DBImportRoutePreviewIssue, bool) {
	issue := model.DBImportRoutePreviewIssue{
		GroupName:     strings.TrimSpace(groupName),
		ChannelName:   strings.TrimSpace(candidate.ChannelName),
		Model:         normalizeModelRef(candidate.Model),
		ResolvedModel: normalizeModelRef(candidate.ResolvedModel),
		KeyID:         candidate.KeyID,
		IssueType:     "invalid_route_target",
	}
	if issue.ResolvedModel == "" {
		issue.ResolvedModel = issue.Model
	}
	if !candidate.Enabled {
		issue.Reason = firstNonEmpty(candidate.Reason, "channel_disabled")
		issue.Action = "channel_disabled"
		return issue, true
	}
	if !candidate.Declared {
		issue.Reason = firstNonEmpty(candidate.Reason, "undeclared_model")
		issue.Action = "undeclared_model"
		return issue, true
	}
	if !candidate.HasKey {
		issue.Reason = firstNonEmpty(candidate.Reason, "missing_key")
		issue.Action = "missing_key"
		return issue, true
	}
	if strings.Contains(candidate.Reason, "missing_model") {
		issue.Reason = candidate.Reason
		issue.Action = "missing_model"
		return issue, true
	}
	return model.DBImportRoutePreviewIssue{}, false
}

func dedupeRoutePreviewIssues(items []model.DBImportRoutePreviewIssue) []model.DBImportRoutePreviewIssue {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]model.DBImportRoutePreviewIssue, 0, len(items))
	for _, item := range items {
		sig := fmt.Sprintf("%s|%s|%s|%s|%d|%s|%s|%s", item.GroupName, item.ChannelName, item.Model, item.ResolvedModel, item.KeyID, item.IssueType, item.Reason, item.Action)
		if _, ok := seen[sig]; ok {
			continue
		}
		seen[sig] = struct{}{}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].IssueType != out[j].IssueType {
			return out[i].IssueType < out[j].IssueType
		}
		if out[i].GroupName != out[j].GroupName {
			return out[i].GroupName < out[j].GroupName
		}
		if out[i].ChannelName != out[j].ChannelName {
			return out[i].ChannelName < out[j].ChannelName
		}
		if out[i].Model != out[j].Model {
			return out[i].Model < out[j].Model
		}
		return out[i].Action < out[j].Action
	})
	return out
}

func collectSnapshotRouteRefsByChannelAndModel(dump *model.DBDump, groupNamesByID map[int]string, channelNamesByID map[int]string) map[string]map[string]struct{} {
	refs := make(map[string]map[string]struct{})
	if dump == nil {
		return refs
	}
	for _, item := range dump.GroupItems {
		modelName := normalizeModelRef(item.ModelName)
		if modelName == "" {
			continue
		}
		lookupKey := fmt.Sprintf("%d|%s", item.ChannelID, modelName)
		if refs[lookupKey] == nil {
			refs[lookupKey] = make(map[string]struct{})
		}
		groupName := strings.TrimSpace(groupNamesByID[item.GroupID])
		channelName := strings.TrimSpace(channelNamesByID[item.ChannelID])
		if groupName != "" {
			refs[lookupKey][fmt.Sprintf("group:%s", groupName)] = struct{}{}
			refs[lookupKey][fmt.Sprintf("group_route:%s", groupName)] = struct{}{}
		}
		if channelName != "" {
			refs[lookupKey][fmt.Sprintf("channel:%s", channelName)] = struct{}{}
		}
	}
	return refs
}

func summarizeCredentialRouteRefs(models []string, refs map[string]map[string]struct{}, channelID int) ([]string, []string) {
	groupSet := make(map[string]struct{})
	contextSet := make(map[string]struct{})
	for _, modelName := range models {
		lookupKey := fmt.Sprintf("%d|%s", channelID, normalizeModelRef(modelName))
		for context := range refs[lookupKey] {
			contextSet[context] = struct{}{}
			if strings.HasPrefix(context, "group:") {
				groupSet[strings.TrimPrefix(context, "group:")] = struct{}{}
			}
		}
	}
	return sortedStringSet(groupSet), sortedStringSet(contextSet)
}

func buildAPIKeyRebindContexts(name string, models []string) []string {
	contexts := make([]string, 0, 1+len(models))
	if strings.TrimSpace(name) != "" {
		contexts = append(contexts, fmt.Sprintf("api_key:%s", strings.TrimSpace(name)))
	}
	for _, modelName := range models {
		contexts = append(contexts, fmt.Sprintf("api_key_model:%s", modelName))
	}
	return dedupeSortedStrings(contexts)
}

func dedupeSortedStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

type dumpModelMappingUsage struct {
	UsageCount    int
	Contexts      map[string]struct{}
	TouchedFields map[string]struct{}
}

func buildModelMappingPreviews(originalDump *model.DBDump, input map[string]string, state *currentImportState) []model.DBImportModelMappingPreview {
	mappings := normalizeModelMappings(input)
	if originalDump == nil || len(mappings) == 0 {
		return nil
	}
	usageByModel := collectDumpModelMappingUsage(originalDump)
	sources := make([]string, 0, len(mappings))
	for source := range mappings {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	previews := make([]model.DBImportModelMappingPreview, 0, len(sources))
	for _, source := range sources {
		target := mappings[source]
		usage := usageByModel[source]
		preview := model.DBImportModelMappingPreview{
			SourceModel:  source,
			TargetModel:  target,
			TargetExists: currentStateHasModel(target, state),
		}
		if usage != nil {
			preview.UsageCount = usage.UsageCount
			preview.Used = usage.UsageCount > 0 || len(usage.Contexts) > 0
			preview.Contexts = sortedStringSet(usage.Contexts)
			preview.TouchedFields = sortedStringSet(usage.TouchedFields)
		}
		if !preview.Used {
			preview.Warnings = append(preview.Warnings, "mapping source not referenced by selected import scopes")
		}
		if preview.Used && !preview.TargetExists {
			preview.Warnings = append(preview.Warnings, "mapped target not found in current environment")
		}
		previews = append(previews, preview)
	}
	return previews
}

func collectDumpModelMappingUsage(dump *model.DBDump) map[string]*dumpModelMappingUsage {
	usages := make(map[string]*dumpModelMappingUsage)
	if dump == nil {
		return usages
	}
	channelNamesByID := make(map[int]string, len(dump.Channels))
	for _, channel := range dump.Channels {
		channelNamesByID[channel.ID] = strings.TrimSpace(channel.Name)
		for _, modelName := range splitCSVModels(channel.Model) {
			addDumpModelMappingUsage(usages, modelName, fmt.Sprintf("channel:%s", channel.Name), "channels.model")
		}
		for _, modelName := range splitCSVModels(channel.CustomModel) {
			addDumpModelMappingUsage(usages, modelName, fmt.Sprintf("channel:%s", channel.Name), "channels.custom_model")
		}
	}
	groupNamesByID := make(map[int]string, len(dump.Groups))
	for _, group := range dump.Groups {
		groupNamesByID[group.ID] = strings.TrimSpace(group.Name)
	}
	for _, key := range dump.ChannelKeys {
		for _, modelName := range splitCSVModels(key.AllowedModels) {
			addDumpModelMappingUsage(usages, modelName, fmt.Sprintf("channel_key:%d", key.ID), "channel_keys.allowed_models")
		}
	}
	for _, override := range dump.RouteTargetOverrides {
		channelRef := channelNamesByID[override.ChannelID]
		if channelRef == "" {
			channelRef = fmt.Sprintf("channel_id:%d", override.ChannelID)
		}
		context := fmt.Sprintf("route_target_override:%s", channelRef)
		if override.ChannelKeyID > 0 {
			context = fmt.Sprintf("%s:key:%d", context, override.ChannelKeyID)
		}
		addDumpModelMappingUsage(usages, override.ModelName, context, "route_target_overrides.model_name")
	}
	for _, item := range dump.GroupItems {
		groupName := groupNamesByID[item.GroupID]
		addDumpModelMappingUsage(usages, item.ModelName, fmt.Sprintf("group:%s", groupName), "group_items.model_name")
		addDumpModelMappingUsage(usages, item.ModelName, fmt.Sprintf("group_route:%s", groupName), "group_items.model_name")
	}
	for _, row := range dump.LLMInfos {
		context := strings.TrimSpace(row.Name)
		if context == "" {
			context = strings.TrimSpace(row.CanonicalName)
		}
		context = fmt.Sprintf("llm_info:%s", context)
		addDumpModelMappingUsage(usages, row.Name, context, "llm_infos.name")
		addDumpModelMappingUsage(usages, row.CanonicalName, context, "llm_infos.canonical_name")
	}
	for _, key := range dump.APIKeys {
		for _, modelName := range splitCSVModels(key.SupportedModels) {
			addDumpModelMappingUsage(usages, modelName, fmt.Sprintf("api_key:%s", key.Name), "api_keys.supported_models")
		}
	}
	for _, row := range dump.StatsModel {
		addDumpModelMappingUsage(usages, row.Name, fmt.Sprintf("stats_model:%s", row.Name), "stats_model.name")
	}
	for i, row := range dump.RelayLogs {
		context := fmt.Sprintf("relay_log:%d", i+1)
		addDumpModelMappingUsage(usages, row.RequestModelName, context, "relay_logs.request_model_name")
		addDumpModelMappingUsage(usages, row.ActualModelName, context, "relay_logs.actual_model_name")
		for j, attempt := range row.Attempts {
			addDumpModelMappingUsage(usages, attempt.ModelName, fmt.Sprintf("%s:attempt:%d", context, j+1), "relay_logs.attempts.model_name")
		}
	}
	return usages
}

func addDumpModelMappingUsage(usages map[string]*dumpModelMappingUsage, modelName, context, field string) {
	modelName = normalizeModelRef(modelName)
	if modelName == "" {
		return
	}
	usage, ok := usages[modelName]
	if !ok {
		usage = &dumpModelMappingUsage{
			Contexts:      make(map[string]struct{}),
			TouchedFields: make(map[string]struct{}),
		}
		usages[modelName] = usage
	}
	usage.UsageCount++
	context = strings.TrimSpace(context)
	if context != "" {
		usage.Contexts[context] = struct{}{}
	}
	field = strings.TrimSpace(field)
	if field != "" {
		usage.TouchedFields[field] = struct{}{}
	}
}

func currentStateHasModel(modelName string, state *currentImportState) bool {
	modelName = normalizeModelRef(modelName)
	if modelName == "" || state == nil {
		return false
	}
	if _, ok := state.llmInfosByName[modelName]; ok {
		return true
	}
	if _, ok := state.llmInfosByCanonical[modelName]; ok {
		return true
	}
	canonical := normalizeModelRef(llmname.CanonicalModelName(modelName))
	if canonical == "" {
		return false
	}
	if _, ok := state.llmInfosByName[canonical]; ok {
		return true
	}
	if _, ok := state.llmInfosByCanonical[canonical]; ok {
		return true
	}
	return false
}

func sortedStringSet(items map[string]struct{}) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for item := range items {
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func buildReplacePrunedChannelNames(dump *model.DBDump, state *currentImportState) []string {
	if dump == nil || state == nil {
		return nil
	}
	keep := make(map[string]struct{}, len(dump.Channels))
	for _, row := range dump.Channels {
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		keep[name] = struct{}{}
	}
	out := make([]string, 0)
	for name := range state.channelsByName {
		if _, ok := keep[name]; ok {
			continue
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func buildReplacePrunedGroupNames(dump *model.DBDump, state *currentImportState) []string {
	if dump == nil || state == nil {
		return nil
	}
	keep := make(map[string]struct{}, len(dump.Groups))
	for _, row := range dump.Groups {
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		keep[name] = struct{}{}
	}
	out := make([]string, 0)
	for name := range state.groupsByName {
		if _, ok := keep[name]; ok {
			continue
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func buildReplacePrunedSettingKeys(dump *model.DBDump, state *currentImportState) []string {
	if dump == nil || state == nil {
		return nil
	}
	keep := make(map[string]struct{}, len(dump.Settings))
	for _, row := range dump.Settings {
		key := strings.TrimSpace(string(row.Key))
		if key == "" {
			continue
		}
		keep[key] = struct{}{}
	}
	out := make([]string, 0)
	for key := range state.settingsByKey {
		if IsSecretSettingKey(model.SettingKey(key)) {
			continue
		}
		if _, ok := keep[key]; ok {
			continue
		}
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func buildReplacePrunedLLMInfoNames(dump *model.DBDump, state *currentImportState) []string {
	if dump == nil || state == nil {
		return nil
	}
	keep := make(map[string]struct{}, len(dump.LLMInfos))
	for _, row := range dump.LLMInfos {
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		keep[name] = struct{}{}
	}
	out := make([]string, 0)
	seen := make(map[string]struct{}, len(state.llmInfosByName))
	for _, row := range state.llmInfosByName {
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		if _, ok := keep[name]; ok {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func buildReplacePrunedAPIKeyNames(currentAPIKeys map[string]model.APIKey, keepAPIKeys []model.APIKey) []string {
	if len(currentAPIKeys) == 0 {
		return nil
	}
	keep := make(map[string]struct{}, len(keepAPIKeys))
	for _, row := range keepAPIKeys {
		apiKey := strings.TrimSpace(row.APIKey)
		if apiKey == "" {
			continue
		}
		keep[apiKey] = struct{}{}
	}
	out := make([]string, 0)
	for apiKey, row := range currentAPIKeys {
		if _, ok := keep[apiKey]; ok {
			continue
		}
		name := strings.TrimSpace(row.Name)
		if name == "" {
			name = fmt.Sprintf("#%d", row.ID)
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func buildModelPolicyDiffs(dump *model.DBDump, state *currentImportState) []model.DBImportModelPolicyDiff {
	if dump == nil || len(dump.LLMInfos) == 0 {
		return nil
	}
	diffs := make([]model.DBImportModelPolicyDiff, 0)
	contextsByModel := collectDumpModelContexts(dump)
	for _, row := range dump.LLMInfos {
		current, ok := findCurrentLLMForSnapshot(row, state)
		if !ok {
			continue
		}
		changedFields := make([]string, 0, 4)
		warnings := make([]string, 0, 4)
		if model.NormalizeBillingMode(row.BillingMode) != model.NormalizeBillingMode(current.BillingMode) {
			changedFields = append(changedFields, "billing_mode")
			warnings = append(warnings, fmt.Sprintf("billing_mode changed from %s to %s", current.BillingMode, row.BillingMode))
		}
		if model.NormalizeProbePolicy(row.ProbePolicy) != model.NormalizeProbePolicy(current.ProbePolicy) {
			changedFields = append(changedFields, "probe_policy")
			warnings = append(warnings, fmt.Sprintf("probe_policy changed from %s to %s", current.ProbePolicy, row.ProbePolicy))
		}
		if row.ProbeIntervalSeconds != current.ProbeIntervalSeconds {
			changedFields = append(changedFields, "probe_interval")
			warnings = append(warnings, fmt.Sprintf("probe_interval changed from %d to %d", current.ProbeIntervalSeconds, row.ProbeIntervalSeconds))
		}
		if row.ProbeConcurrencyLimit != current.ProbeConcurrencyLimit {
			changedFields = append(changedFields, "probe_concurrency")
			warnings = append(warnings, fmt.Sprintf("probe_concurrency changed from %d to %d", current.ProbeConcurrencyLimit, row.ProbeConcurrencyLimit))
		}
		if len(changedFields) == 0 {
			continue
		}
		diffs = append(diffs, model.DBImportModelPolicyDiff{
			Model:        normalizeModelRef(row.Name),
			CurrentModel: current.Name,
			Canonical:    current.CanonicalName,
			Before: model.DBImportModelPolicyState{
				BillingMode:      string(current.BillingMode),
				ProbePolicy:      string(current.ProbePolicy),
				ProbeInterval:    current.ProbeIntervalSeconds,
				ProbeConcurrency: current.ProbeConcurrencyLimit,
			},
			After: model.DBImportModelPolicyState{
				BillingMode:      string(row.BillingMode),
				ProbePolicy:      string(row.ProbePolicy),
				ProbeInterval:    row.ProbeIntervalSeconds,
				ProbeConcurrency: row.ProbeConcurrencyLimit,
			},
			ChangedFields: changedFields,
			ImpactLevel:   "high",
			Warnings:      warnings,
			Contexts:      collectContextsForModel(normalizeModelRef(row.Name), contextsByModel),
		})
	}
	return diffs
}

func buildRoutePreviewDiffs(dump *model.DBDump, state *currentImportState, mode model.DBImportMode) []model.DBImportRoutePreviewDiff {
	if dump == nil {
		return nil
	}
	beforeMap := buildCurrentRoutePreviewMap(state)
	afterMap := buildImportedRoutePreviewMap(dump, state)
	keys := make([]routePreviewChainKey, 0, len(beforeMap)+len(afterMap))
	seen := make(map[routePreviewChainKey]struct{}, len(beforeMap)+len(afterMap))
	for key := range beforeMap {
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	for key := range afterMap {
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].groupName != keys[j].groupName {
			return keys[i].groupName < keys[j].groupName
		}
		return keys[i].modelName < keys[j].modelName
	})
	diffs := make([]model.DBImportRoutePreviewDiff, 0, len(keys))
	for _, key := range keys {
		before := append([]model.DBImportRoutePreviewCandidate(nil), beforeMap[key]...)
		after := append([]model.DBImportRoutePreviewCandidate(nil), afterMap[key]...)
		diff := model.DBImportRoutePreviewDiff{GroupName: key.groupName, Model: key.modelName}
		if mode == model.DBImportModeSkip {
			if _, ok := state.groupsByName[key.groupName]; ok && len(after) > 0 {
				diff.BeforeCandidates = before
				diff.RemovedCandidates = after
				diff.SkipReasons = []string{fmt.Sprintf("skip_mode_preserved_existing_group:%s", key.groupName)}
				diffs = append(diffs, diff)
				continue
			}
		}
		removed, added := diffRoutePreviewCandidates(before, after)
		if len(removed) == 0 && len(added) == 0 && len(before) == 0 && len(after) == 0 {
			continue
		}
		diff.BeforeCandidates = before
		diff.AfterCandidates = after
		diff.RemovedCandidates = removed
		diff.AddedCandidates = added
		if len(diff.SkipReasons) == 0 && (len(removed) > 0 || len(added) > 0) {
			diff.SkipReasons = append(diff.SkipReasons, "route_candidates_changed")
		}
		if len(removed) > 0 || len(added) > 0 || len(before) != len(after) {
			diff.FallbackChanged = true
		}
		diffs = append(diffs, diff)
	}
	return diffs
}

func buildCurrentRoutePreviewMap(state *currentImportState) map[routePreviewChainKey][]model.DBImportRoutePreviewCandidate {
	out := make(map[routePreviewChainKey][]model.DBImportRoutePreviewCandidate)
	if state == nil {
		return out
	}
	for _, group := range state.groupsByName {
		for _, item := range group.Items {
			channel, ok := state.channelsByID[item.ChannelID]
			if !ok {
				continue
			}
			candidate := buildRoutePreviewCandidate(channel, item, state.llmInfosByName, false, state.routeTargetOverridesByKey)
			key := routePreviewChainKey{groupName: group.Name, modelName: item.ModelName}
			out[key] = append(out[key], candidate)
		}
	}
	for key := range out {
		sortRoutePreviewCandidates(out[key])
	}
	return out
}

func buildImportedRoutePreviewMap(dump *model.DBDump, state *currentImportState) map[routePreviewChainKey][]model.DBImportRoutePreviewCandidate {
	out := make(map[routePreviewChainKey][]model.DBImportRoutePreviewCandidate)
	if dump == nil {
		return out
	}
	dumpChannelNamesByID := make(map[int]string, len(dump.Channels))
	for _, row := range dump.Channels {
		dumpChannelNamesByID[row.ID] = strings.TrimSpace(row.Name)
	}
	dumpChannelsByID := make(map[int]model.Channel, len(dump.Channels))
	for _, row := range dump.Channels {
		channel := row
		channel.Keys = nil
		for _, key := range dump.ChannelKeys {
			if key.ChannelID == row.ID && strings.TrimSpace(key.ChannelKey) != "" {
				channel.Keys = append(channel.Keys, key)
			}
		}
		if existing, ok := state.channelsByName[strings.TrimSpace(row.Name)]; ok {
			channel.ID = existing.ID
			if len(channel.Keys) == 0 {
				channel.Keys = existing.Keys
			} else {
				for i := range channel.Keys {
					for _, existingKey := range existing.Keys {
						if strings.TrimSpace(existingKey.ChannelKey) != strings.TrimSpace(channel.Keys[i].ChannelKey) {
							continue
						}
						channel.Keys[i].ID = existingKey.ID
						break
					}
				}
			}
		}
		dumpChannelsByID[row.ID] = channel
	}
	dumpKeyLocalIDBySnapshotID := make(map[int]int, len(dump.ChannelKeys))
	for _, key := range dump.ChannelKeys {
		localID := key.ID
		channelName := dumpChannelNamesByID[key.ChannelID]
		if existing, ok := state.channelsByName[channelName]; ok {
			for _, existingKey := range existing.Keys {
				if strings.TrimSpace(existingKey.ChannelKey) != strings.TrimSpace(key.ChannelKey) {
					continue
				}
				localID = existingKey.ID
				break
			}
		}
		dumpKeyLocalIDBySnapshotID[key.ID] = localID
	}
	overridesByLookupKey := make(map[string]model.RouteTargetOverride, len(dump.RouteTargetOverrides))
	for _, row := range dump.RouteTargetOverrides {
		channelID := row.ChannelID
		if channel, ok := dumpChannelsByID[row.ChannelID]; ok && channel.ID > 0 {
			channelID = channel.ID
		}
		channelKeyID := row.ChannelKeyID
		if mappedKeyID, ok := dumpKeyLocalIDBySnapshotID[row.ChannelKeyID]; ok && mappedKeyID > 0 {
			channelKeyID = mappedKeyID
		}
		row.ChannelID = channelID
		row.ChannelKeyID = channelKeyID
		overridesByLookupKey[routeTargetOverrideLookupKey(row.ChannelID, row.ChannelKeyID, row.ModelName)] = row
	}
	dumpLLMs := make(map[string]model.LLMInfo, len(dump.LLMInfos))
	for _, row := range dump.LLMInfos {
		dumpLLMs[normalizeModelRef(row.Name)] = row
	}
	for _, row := range state.llmInfosByName {
		key := normalizeModelRef(row.Name)
		if _, ok := dumpLLMs[key]; !ok {
			dumpLLMs[key] = row
		}
	}
	dumpGroupsByID := make(map[int]model.Group, len(dump.Groups))
	for _, row := range dump.Groups {
		dumpGroupsByID[row.ID] = row
	}
	for _, item := range dump.GroupItems {
		group, ok := dumpGroupsByID[item.GroupID]
		if !ok {
			continue
		}
		channel, ok := dumpChannelsByID[item.ChannelID]
		if !ok {
			continue
		}
		candidate := buildRoutePreviewCandidate(channel, item, dumpLLMs, true, overridesByLookupKey)
		key := routePreviewChainKey{groupName: group.Name, modelName: item.ModelName}
		out[key] = append(out[key], candidate)
	}
	for key := range out {
		sortRoutePreviewCandidates(out[key])
	}
	return out
}

func buildPreviewRouteTargetPolicy(channel model.Channel, key model.ChannelKey, resolvedModel string, llmInfos map[string]model.LLMInfo, previewOverrides map[string]model.RouteTargetOverride) model.RouteTargetResolvedPolicy {
	policy := defaultRouteTargetResolvedPolicy(channel.ID, key.ID, key.SourceType, resolvedModel)
	if info, ok := llmInfos[normalizeModelRef(resolvedModel)]; ok {
		policy.BillingMode = info.BillingMode
		policy.BillingModeBasis = model.RouteTargetPolicyBasisModelDefault
		policy.ProbePolicy = info.ProbePolicy
		policy.ProbePolicyBasis = model.RouteTargetPolicyBasisModelDefault
		if info.ProbeIntervalSeconds > 0 {
			policy.ProbeIntervalSeconds = info.ProbeIntervalSeconds
		}
		policy.ProbeIntervalBasis = model.RouteTargetPolicyBasisModelDefault
		if info.ProbeConcurrencyLimit > 0 {
			policy.ProbeConcurrencyLimit = info.ProbeConcurrencyLimit
		}
		policy.ProbeConcurrencyBasis = model.RouteTargetPolicyBasisModelDefault
	}
	if previewOverrides != nil {
		if override, ok := previewOverrides[routeTargetOverrideLookupKey(channel.ID, key.ID, resolvedModel)]; ok {
			policy.BillingMode = override.BillingMode
			policy.BillingModeBasis = model.RouteTargetPolicyBasisExplicitOverride
			policy.ProbePolicy = override.ProbePolicy
			policy.ProbePolicyBasis = model.RouteTargetPolicyBasisExplicitOverride
			policy.ProbeIntervalSeconds = override.ProbeIntervalSeconds
			policy.ProbeIntervalBasis = model.RouteTargetPolicyBasisExplicitOverride
			policy.ProbeConcurrencyLimit = override.ProbeConcurrencyLimit
			policy.ProbeConcurrencyBasis = model.RouteTargetPolicyBasisExplicitOverride
		}
		return policy
	}
	if override, ok := RouteTargetOverrideGet(channel.ID, key.ID, resolvedModel); ok {
		policy.BillingMode = override.BillingMode
		policy.BillingModeBasis = model.RouteTargetPolicyBasisExplicitOverride
		policy.ProbePolicy = override.ProbePolicy
		policy.ProbePolicyBasis = model.RouteTargetPolicyBasisExplicitOverride
		policy.ProbeIntervalSeconds = override.ProbeIntervalSeconds
		policy.ProbeIntervalBasis = model.RouteTargetPolicyBasisExplicitOverride
		policy.ProbeConcurrencyLimit = override.ProbeConcurrencyLimit
		policy.ProbeConcurrencyBasis = model.RouteTargetPolicyBasisExplicitOverride
	}
	return policy
}

func buildRoutePreviewCandidate(channel model.Channel, item model.GroupItem, llmInfos map[string]model.LLMInfo, allowCanonicalFallback bool, previewOverrides map[string]model.RouteTargetOverride) model.DBImportRoutePreviewCandidate {
	originalModel := normalizeModelRef(item.ModelName)
	resolvedModel := originalModel
	reasonParts := make([]string, 0, 2)
	resolvedInfo, hasResolvedInfo := llmInfos[originalModel]
	if !hasResolvedInfo && allowCanonicalFallback {
		if aliasInfo, ok := findInfoByCanonical(llmInfos, llmname.CanonicalModelName(originalModel)); ok {
			resolvedInfo = aliasInfo
			hasResolvedInfo = true
		}
	}
	if hasResolvedInfo {
		if canonical := normalizeModelRef(resolvedInfo.CanonicalName); canonical != "" && canonical != originalModel {
			if findInfoByName(llmInfos, canonical) {
				resolvedModel = canonical
				reasonParts = append(reasonParts, "alias_remapped")
			}
		}
	} else {
		reasonParts = append(reasonParts, "missing_model")
	}
	declared := channel.SupportsModel(resolvedModel)
	key := channel.GetChannelKeyForModel(resolvedModel)
	if key.ID == 0 && resolvedModel != originalModel {
		key = channel.GetChannelKeyForModel(originalModel)
	}
	candidate := model.DBImportRoutePreviewCandidate{
		ChannelName:   channel.Name,
		Model:         originalModel,
		ResolvedModel: resolvedModel,
		Priority:      item.Priority,
		Weight:        item.Weight,
		Enabled:       channel.Enabled,
		Declared:      declared,
		HasKey:        strings.TrimSpace(key.ChannelKey) != "",
		KeyID:         key.ID,
		KeySourceType: strings.TrimSpace(key.SourceType),
		KeyRemark:     key.Remark,
		Reason:        strings.Join(reasonParts, ","),
	}
	if candidate.HasKey {
		policy := buildPreviewRouteTargetPolicy(channel, key, resolvedModel, llmInfos, previewOverrides)
		candidate.BillingMode = string(policy.BillingMode)
		candidate.ProbePolicy = string(policy.ProbePolicy)
		candidate.ProbeIntervalSeconds = policy.ProbeIntervalSeconds
		candidate.ProbeConcurrencyLimit = policy.ProbeConcurrencyLimit
		candidate.BillingModeBasis = policy.BillingModeBasis
		candidate.ProbePolicyBasis = policy.ProbePolicyBasis
		candidate.ProbeIntervalBasis = policy.ProbeIntervalBasis
		candidate.ProbeConcurrencyBasis = policy.ProbeConcurrencyBasis
		candidate.PolicyBasis = policy.PolicyBasisSummary()
	}
	return candidate
}

func findInfoByCanonical(llmInfos map[string]model.LLMInfo, canonical string) (model.LLMInfo, bool) {
	canonical = normalizeModelRef(canonical)
	if canonical == "" {
		return model.LLMInfo{}, false
	}
	for _, info := range llmInfos {
		if normalizeModelRef(info.CanonicalName) == canonical {
			return info, true
		}
	}
	return model.LLMInfo{}, false
}

func findInfoByName(llmInfos map[string]model.LLMInfo, name string) bool {
	_, ok := llmInfos[normalizeModelRef(name)]
	return ok
}

func diffRoutePreviewCandidates(before, after []model.DBImportRoutePreviewCandidate) ([]model.DBImportRoutePreviewCandidate, []model.DBImportRoutePreviewCandidate) {
	beforeMap := make(map[string]model.DBImportRoutePreviewCandidate, len(before))
	afterMap := make(map[string]model.DBImportRoutePreviewCandidate, len(after))
	for _, row := range before {
		beforeMap[routePreviewCandidateSignature(row)] = row
	}
	for _, row := range after {
		afterMap[routePreviewCandidateSignature(row)] = row
	}
	removed := make([]model.DBImportRoutePreviewCandidate, 0)
	added := make([]model.DBImportRoutePreviewCandidate, 0)
	for sig, row := range beforeMap {
		if _, ok := afterMap[sig]; !ok {
			removed = append(removed, row)
		}
	}
	for sig, row := range afterMap {
		if _, ok := beforeMap[sig]; !ok {
			added = append(added, row)
		}
	}
	sortRoutePreviewCandidates(removed)
	sortRoutePreviewCandidates(added)
	return removed, added
}

func routePreviewCandidateSignature(row model.DBImportRoutePreviewCandidate) string {
	return fmt.Sprintf("%s|%s|%s|%d|%d|%t|%t|%d|%s|%s|%s|%s|%d|%d|%s|%s|%s|%s|%s", row.ChannelName, row.Model, row.ResolvedModel, row.Priority, row.Weight, row.Enabled, row.Declared, row.KeyID, row.KeySourceType, row.Reason, row.BillingMode, row.ProbePolicy, row.ProbeIntervalSeconds, row.ProbeConcurrencyLimit, row.PolicyBasis, row.BillingModeBasis, row.ProbePolicyBasis, row.ProbeIntervalBasis, row.ProbeConcurrencyBasis)
}

func sortRoutePreviewCandidates(rows []model.DBImportRoutePreviewCandidate) {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ChannelName != rows[j].ChannelName {
			return rows[i].ChannelName < rows[j].ChannelName
		}
		if rows[i].ResolvedModel != rows[j].ResolvedModel {
			return rows[i].ResolvedModel < rows[j].ResolvedModel
		}
		if rows[i].Priority != rows[j].Priority {
			return rows[i].Priority < rows[j].Priority
		}
		return rows[i].KeyID < rows[j].KeyID
	})
}

func buildPostImportValidationReport(ctx context.Context, dump *model.DBDump) (*model.DBImportPostValidationReport, int64, []string, error) {
	if err := InitCache(); err != nil {
		return nil, 0, nil, err
	}
	groups, err := GroupList(ctx)
	if err != nil {
		return nil, 0, nil, err
	}
	channels, err := ChannelList(ctx)
	if err != nil {
		return nil, 0, nil, err
	}
	channelByID := make(map[int]model.Channel, len(channels))
	for _, row := range channels {
		channelByID[row.ID] = row
	}
	report := &model.DBImportPostValidationReport{Summary: &model.DBImportPostValidationSummary{GroupsScanned: len(groups)}}
	cleanedCount := int64(0)
	warnings := make([]string, 0)
	staleGroupIDs := make([]int, 0)
	seenDisabledChannels := make(map[string]struct{})
	seenNoKeyChannels := make(map[string]struct{})
	for _, group := range groups {
		degraded := false
		if len(group.Items) == 0 {
			report.EmptyGroups = append(report.EmptyGroups, fmt.Sprintf("group:%s has no route items", group.Name))
			report.DegradedGroups = append(report.DegradedGroups, fmt.Sprintf("group:%s has no usable routes", group.Name))
			continue
		}
		for _, item := range group.Items {
			report.Summary.CandidatesScanned++
			channel, ok := channelByID[item.ChannelID]
			if !ok || !channel.SupportsModel(item.ModelName) {
				if err := db.GetDB().WithContext(ctx).Delete(&model.GroupItem{}, item.ID).Error; err != nil {
					return nil, cleanedCount, nil, err
				}
				cleanedCount++
				staleGroupIDs = append(staleGroupIDs, group.ID)
				report.StaleItemsRemoved = append(report.StaleItemsRemoved, fmt.Sprintf("removed stale group item group:%s channel_id:%d model:%s", group.Name, item.ChannelID, item.ModelName))
				continue
			}
			if !channel.Enabled {
				degraded = true
				if _, ok := seenDisabledChannels[channel.Name]; !ok {
					seenDisabledChannels[channel.Name] = struct{}{}
					report.DisabledChannels = append(report.DisabledChannels, fmt.Sprintf("channel:%s is disabled", channel.Name))
				}
				continue
			}
			if strings.TrimSpace(channel.GetChannelKeyForModel(item.ModelName).ChannelKey) == "" {
				degraded = true
				if _, ok := seenNoKeyChannels[channel.Name]; !ok {
					seenNoKeyChannels[channel.Name] = struct{}{}
					report.ChannelsWithoutKeys = append(report.ChannelsWithoutKeys, fmt.Sprintf("channel:%s has no available keys", channel.Name))
				}
			}
		}
		if degraded {
			report.DegradedGroups = append(report.DegradedGroups, fmt.Sprintf("group:%s has degraded routes", group.Name))
		}
	}
	for _, row := range report.EmptyGroups {
		groupName := strings.TrimPrefix(strings.TrimSuffix(row, " has no route items"), "group:")
		if !sliceContainsSubstring(report.DegradedGroups, groupName) {
			report.DegradedGroups = append(report.DegradedGroups, fmt.Sprintf("group:%s has no usable routes", groupName))
		}
	}
	if len(staleGroupIDs) > 0 {
		if err := groupRefreshCacheByIDs(staleGroupIDs, ctx); err != nil && err != gorm.ErrRecordNotFound {
			return nil, cleanedCount, nil, err
		}
	}
	currentState, err := loadCurrentImportState(db.GetDB().WithContext(ctx))
	if err != nil {
		return nil, cleanedCount, nil, err
	}
	report.RouteWarnings = buildPostImportRouteWarnings(dump, currentState)
	report.PriceRuleWarnings = buildPriceRuleWarnings(dump, ctx)
	report.AliasMappings, report.AliasWarnings = buildAliasValidationWarnings(dump, ctx)
	report.Summary.DegradedGroups = len(report.DegradedGroups)
	report.Summary.EmptyGroups = len(report.EmptyGroups)
	report.Summary.DisabledChannels = len(report.DisabledChannels)
	report.Summary.ChannelsWithoutKeys = len(report.ChannelsWithoutKeys)
	report.Summary.StaleItemsRemoved = len(report.StaleItemsRemoved)
	report.Summary.RouteWarnings = len(report.RouteWarnings)
	report.Summary.PriceRuleWarnings = len(report.PriceRuleWarnings)
	report.Summary.AliasMappings = len(report.AliasMappings)
	report.Summary.AliasWarnings = len(report.AliasWarnings)
	if report.Summary.DegradedGroups > 0 {
		warnings = append(warnings, fmt.Sprintf("post-import validation found %d degraded groups", report.Summary.DegradedGroups))
	}
	if report.Summary.StaleItemsRemoved > 0 {
		warnings = append(warnings, fmt.Sprintf("cleaned %d stale group items after import", report.Summary.StaleItemsRemoved))
	}
	if report.Summary.RouteWarnings > 0 {
		warnings = append(warnings, fmt.Sprintf("post-import validation found %d route warnings", report.Summary.RouteWarnings))
	}
	if report.Summary.PriceRuleWarnings > 0 {
		warnings = append(warnings, fmt.Sprintf("post-import validation found %d price rule warnings", report.Summary.PriceRuleWarnings))
	}
	if report.Summary.AliasWarnings > 0 {
		warnings = append(warnings, fmt.Sprintf("post-import validation found %d alias warnings", report.Summary.AliasWarnings))
	}
	return report, cleanedCount, warnings, nil
}

func buildPostImportRouteWarnings(dump *model.DBDump, state *currentImportState) []string {
	if dump == nil || state == nil {
		return nil
	}
	expectedMap := buildImportedRoutePreviewMap(dump, state)
	actualMap := buildCurrentRoutePreviewMap(state)
	keys := collectPostImportRouteValidationKeys(dump, expectedMap, actualMap)
	if len(keys) == 0 {
		return nil
	}
	warnings := make([]string, 0, len(keys))
	for _, key := range keys {
		expected := append([]model.DBImportRoutePreviewCandidate(nil), expectedMap[key]...)
		actual := append([]model.DBImportRoutePreviewCandidate(nil), actualMap[key]...)
		removed, added := diffRoutePreviewCandidates(expected, actual)
		if len(expected) == 0 && len(actual) == 0 {
			continue
		}
		if len(removed) == 0 && len(added) == 0 && len(expected) == len(actual) {
			continue
		}
		parts := []string{
			fmt.Sprintf("group:%s model:%s route verification mismatch", key.groupName, key.modelName),
			fmt.Sprintf("expected:%s", summarizeRoutePreviewCandidates(expected)),
			fmt.Sprintf("actual:%s", summarizeRoutePreviewCandidates(actual)),
		}
		if len(removed) > 0 {
			parts = append(parts, fmt.Sprintf("missing:%s", summarizeRoutePreviewCandidates(removed)))
		}
		if len(added) > 0 {
			parts = append(parts, fmt.Sprintf("extra:%s", summarizeRoutePreviewCandidates(added)))
		}
		if len(expected) != len(actual) {
			parts = append(parts, fmt.Sprintf("candidate_count:%d->%d", len(expected), len(actual)))
		}
		warnings = append(warnings, strings.Join(parts, " | "))
	}
	sort.Strings(warnings)
	return warnings
}

func collectPostImportRouteValidationKeys(dump *model.DBDump, expectedMap, actualMap map[routePreviewChainKey][]model.DBImportRoutePreviewCandidate) []routePreviewChainKey {
	if dump == nil || len(dump.Groups) == 0 {
		return nil
	}
	relevantGroups := make(map[string]struct{}, len(dump.Groups))
	for _, group := range dump.Groups {
		name := strings.TrimSpace(group.Name)
		if name == "" {
			continue
		}
		relevantGroups[name] = struct{}{}
	}
	if len(relevantGroups) == 0 {
		return nil
	}
	keys := make([]routePreviewChainKey, 0, len(expectedMap)+len(actualMap))
	seen := make(map[routePreviewChainKey]struct{}, len(expectedMap)+len(actualMap))
	appendKey := func(key routePreviewChainKey) {
		if _, ok := relevantGroups[key.groupName]; !ok {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	for key := range expectedMap {
		appendKey(key)
	}
	for key := range actualMap {
		appendKey(key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].groupName != keys[j].groupName {
			return keys[i].groupName < keys[j].groupName
		}
		return keys[i].modelName < keys[j].modelName
	})
	return keys
}

func summarizeRoutePreviewCandidates(rows []model.DBImportRoutePreviewCandidate) string {
	if len(rows) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(rows))
	for _, row := range rows {
		modelName := row.Model
		if row.ResolvedModel != "" && row.ResolvedModel != row.Model {
			modelName = fmt.Sprintf("%s->%s", row.Model, row.ResolvedModel)
		}
		keyRef := "no-key"
		if row.KeyID > 0 {
			keyRef = fmt.Sprintf("key:%d", row.KeyID)
		}
		part := fmt.Sprintf("%s:%s@p%d/w%d/%s", row.ChannelName, modelName, row.Priority, row.Weight, keyRef)
		if row.Reason != "" {
			part = fmt.Sprintf("%s:%s", part, row.Reason)
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, ", ")
}

func buildPostImportHealthCheck(ctx context.Context, compatibility *model.DBImportCompatibilityReport, targets []ChannelModelHealthCheckTarget) *model.DBImportHealthCheckReport {
	_ = compatibility
	if len(targets) == 0 {
		return nil
	}
	return RunImportHealthChecks(ctx, targets)
}

func buildRollbackPreviewWarnings(dump *model.DBDump, compatibility *model.DBImportCompatibilityReport) []string {
	warnings := make([]string, 0)
	if compatibility != nil && compatibility.Summary != nil {
		if compatibility.Summary.RoutePreviewDiffs > 0 {
			warnings = append(warnings, fmt.Sprintf("route preview diffs: %d", compatibility.Summary.RoutePreviewDiffs))
		}
		if compatibility.Summary.BaseURLMismatches > 0 {
			warnings = append(warnings, fmt.Sprintf("base URL mismatches: %d", compatibility.Summary.BaseURLMismatches))
		}
	}
	if dump != nil && dump.IncludeStats {
		warnings = append(warnings, "includes stats tables")
	}
	if dump != nil && dump.IncludeLogs {
		warnings = append(warnings, "includes relay logs")
	}
	return warnings
}

func collectDumpModelContexts(dump *model.DBDump) map[string][]string {
	contexts := make(map[string][]string)
	add := func(modelName, context string) {
		modelName = normalizeModelRef(modelName)
		context = strings.TrimSpace(context)
		if modelName == "" || context == "" {
			return
		}
		contexts[modelName] = append(contexts[modelName], context)
	}
	if dump == nil {
		return contexts
	}
	for _, channel := range dump.Channels {
		for _, modelName := range splitCSVModels(channel.Model) {
			add(modelName, fmt.Sprintf("channel:%s", channel.Name))
		}
		for _, modelName := range splitCSVModels(channel.CustomModel) {
			add(modelName, fmt.Sprintf("channel:%s", channel.Name))
		}
	}
	groupNamesByID := make(map[int]string, len(dump.Groups))
	for _, group := range dump.Groups {
		groupNamesByID[group.ID] = group.Name
	}
	for _, item := range dump.GroupItems {
		add(item.ModelName, fmt.Sprintf("group:%s", groupNamesByID[item.GroupID]))
		add(item.ModelName, fmt.Sprintf("group_route:%s", groupNamesByID[item.GroupID]))
	}
	for _, key := range dump.ChannelKeys {
		for _, modelName := range splitCSVModels(key.AllowedModels) {
			add(modelName, fmt.Sprintf("channel_key:%d", key.ID))
		}
	}
	for _, key := range dump.APIKeys {
		for _, modelName := range splitCSVModels(key.SupportedModels) {
			add(modelName, fmt.Sprintf("api_key:%s", key.Name))
		}
	}
	for modelName, items := range contexts {
		seen := map[string]struct{}{}
		uniq := make([]string, 0, len(items))
		for _, item := range items {
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			uniq = append(uniq, item)
		}
		sort.Strings(uniq)
		contexts[modelName] = uniq
	}
	return contexts
}

func collectContextsForModel(modelName string, contexts map[string][]string) []string {
	items := append([]string(nil), contexts[normalizeModelRef(modelName)]...)
	sort.Strings(items)
	return items
}

func findAliasTarget(modelName string, dump *model.DBDump, state *currentImportState) (model.LLMInfo, bool) {
	modelName = normalizeModelRef(modelName)
	if modelName == "" || state == nil {
		return model.LLMInfo{}, false
	}
	for _, row := range dump.LLMInfos {
		if normalizeModelRef(row.Name) != modelName {
			continue
		}
		canonical := normalizeModelRef(row.CanonicalName)
		if canonical == "" {
			break
		}
		if current, ok := state.llmInfosByName[canonical]; ok {
			return current, true
		}
		if current, ok := state.llmInfosByCanonical[canonical]; ok {
			return current, true
		}
	}
	canonical := llmname.CanonicalModelName(modelName)
	if canonical != "" && canonical != modelName {
		if current, ok := state.llmInfosByName[canonical]; ok {
			return current, true
		}
		if current, ok := state.llmInfosByCanonical[canonical]; ok {
			return current, true
		}
	}
	return model.LLMInfo{}, false
}

func findCurrentLLMForSnapshot(row model.LLMInfo, state *currentImportState) (model.LLMInfo, bool) {
	if state == nil {
		return model.LLMInfo{}, false
	}
	canonical := normalizeModelRef(row.CanonicalName)
	if canonical != "" {
		if current, ok := state.llmInfosByName[canonical]; ok {
			return current, true
		}
		if current, ok := state.llmInfosByCanonical[canonical]; ok {
			return current, true
		}
	}
	if current, ok := state.llmInfosByName[normalizeModelRef(row.Name)]; ok {
		return current, true
	}
	return model.LLMInfo{}, false
}

func buildPriceRuleWarnings(dump *model.DBDump, ctx context.Context) []string {
	if dump == nil || len(dump.LLMInfos) == 0 {
		return nil
	}
	state, err := loadCurrentImportState(db.GetDB().WithContext(ctx))
	if err != nil {
		return nil
	}
	warnings := make([]string, 0)
	for _, row := range dump.LLMInfos {
		current, ok := findCurrentLLMForSnapshot(row, state)
		if !ok {
			continue
		}
		if model.NormalizeBillingMode(row.BillingMode) != model.NormalizeBillingMode(current.BillingMode) {
			warnings = append(warnings, fmt.Sprintf("model:%s billing_mode changed from %s to %s", row.Name, current.BillingMode, row.BillingMode))
		}
		if model.NormalizeProbePolicy(row.ProbePolicy) != model.NormalizeProbePolicy(current.ProbePolicy) {
			warnings = append(warnings, fmt.Sprintf("model:%s probe_policy changed from %s to %s", row.Name, current.ProbePolicy, row.ProbePolicy))
		}
		if current.ProbeConcurrencyLimit > 1 || row.ProbeConcurrencyLimit > 1 || current.ProbePolicy == model.ProbePolicyConcurrent || row.ProbePolicy == model.ProbePolicyConcurrent {
			warnings = append(warnings, fmt.Sprintf("model:%s concurrent probe/race may increase cost", row.Name))
		}
	}
	return warnings
}

func buildAliasValidationWarnings(dump *model.DBDump, ctx context.Context) ([]string, []string) {
	state, err := loadCurrentImportState(db.GetDB().WithContext(ctx))
	if err != nil || dump == nil {
		return nil, nil
	}
	mappings := make([]string, 0)
	warnings := make([]string, 0)
	for _, row := range dump.LLMInfos {
		if current, ok := findAliasTarget(row.Name, dump, state); ok {
			mappings = append(mappings, fmt.Sprintf("model:%s remapped to current alias:%s", row.Name, current.Name))
		}
	}
	for _, item := range dump.GroupItems {
		if current, ok := findAliasTarget(item.ModelName, dump, state); ok {
			mappings = append(mappings, fmt.Sprintf("group route model:%s can map to current alias:%s", item.ModelName, current.Name))
		}
	}
	for _, channel := range dump.Channels {
		for _, modelName := range splitCSVModels(channel.Model) {
			if current, ok := findAliasTarget(modelName, dump, state); ok {
				warnings = append(warnings, fmt.Sprintf("channel:%s model:%s resolves to current alias:%s", channel.Name, modelName, current.Name))
			}
		}
	}
	for _, key := range dump.ChannelKeys {
		for _, modelName := range splitCSVModels(key.AllowedModels) {
			if current, ok := findAliasTarget(modelName, dump, state); ok {
				warnings = append(warnings, fmt.Sprintf("channel_key:%d model:%s resolves to current alias:%s", key.ID, modelName, current.Name))
			}
		}
	}
	for _, key := range dump.APIKeys {
		for _, modelName := range splitCSVModels(key.SupportedModels) {
			if current, ok := findAliasTarget(modelName, dump, state); ok {
				warnings = append(warnings, fmt.Sprintf("api_key:%s model:%s resolves to current alias:%s", key.Name, modelName, current.Name))
			}
		}
	}
	sort.Strings(mappings)
	sort.Strings(warnings)
	return mappings, warnings
}
