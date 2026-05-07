package op

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db/migrate"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const dbDumpVersion = 1
const importSnapshotDirName = "import-snapshots"
const importSnapshotLatestFilename = "latest-import-snapshot.json"
const importSnapshotDirPerm = 0o700
const importSnapshotFilePerm = 0o600

const snapshotCredentialsRedactedWarning = "snapshot does not include plaintext credentials; rebind channel/API keys after import"

type importSnapshotMetadata struct {
	SnapshotPath string    `json:"snapshot_path"`
	SnapshotName string    `json:"snapshot_name"`
	ImportedAt   time.Time `json:"imported_at"`
}

type currentImportState struct {
	channelsByName            map[string]model.Channel
	channelsByID              map[int]model.Channel
	groupsByName              map[string]model.Group
	llmInfosByName            map[string]model.LLMInfo
	llmInfosByCanonical       map[string]model.LLMInfo
	settingsByKey             map[string]model.Setting
	apiKeysByAPIKey           map[string]model.APIKey
	routeTargetOverridesByKey map[string]model.RouteTargetOverride
}

type routePreviewChainKey struct {
	groupName string
	modelName string
}

type preparedChannelKeyImport struct {
	SnapshotID int
	Row        model.ChannelKey
}

type preparedAPIKeyImport struct {
	SnapshotID int
	Row        model.APIKey
}

type preparedRouteTargetOverrideImport struct {
	SnapshotID int
	Row        model.RouteTargetOverride
}

func DBExportAll(ctx context.Context, includeLogs, includeStats, includeSecrets bool) (*model.DBDump, error) {
	conn := db.GetDB().WithContext(ctx)

	d := &model.DBDump{
		Version:      dbDumpVersion,
		ExportedAt:   time.Now().UTC(),
		IncludeLogs:  includeLogs,
		IncludeStats: includeStats,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			Encrypted:       false,
			ContainsSecrets: false,
		},
	}

	if err := conn.Find(&d.Channels).Error; err != nil {
		return nil, fmt.Errorf("export channels: %w", err)
	}
	if err := conn.Find(&d.Users).Error; err != nil {
		return nil, fmt.Errorf("export users: %w", err)
	}
	if err := conn.Find(&d.ChannelKeys).Error; err != nil {
		return nil, fmt.Errorf("export channel_keys: %w", err)
	}
	if err := conn.Find(&d.RouteTargetOverrides).Error; err != nil {
		return nil, fmt.Errorf("export route_target_overrides: %w", err)
	}
	if err := conn.Find(&d.Groups).Error; err != nil {
		return nil, fmt.Errorf("export groups: %w", err)
	}
	if err := conn.Find(&d.GroupItems).Error; err != nil {
		return nil, fmt.Errorf("export group_items: %w", err)
	}
	if err := conn.Find(&d.LLMInfos).Error; err != nil {
		return nil, fmt.Errorf("export llm_infos: %w", err)
	}
	if err := conn.Find(&d.APIKeys).Error; err != nil {
		return nil, fmt.Errorf("export api_keys: %w", err)
	}
	if err := conn.Find(&d.Settings).Error; err != nil {
		return nil, fmt.Errorf("export settings: %w", err)
	}
	filteredSettings := make([]model.Setting, 0, len(d.Settings))
	for _, setting := range d.Settings {
		if IsSecretSettingKey(setting.Key) {
			continue
		}
		filteredSettings = append(filteredSettings, setting)
	}
	d.Settings = filteredSettings
	var migrationRecords []migrate.MigrationRecord
	if err := conn.Find(&migrationRecords).Error; err != nil {
		return nil, fmt.Errorf("export migration_records: %w", err)
	}
	if len(migrationRecords) > 0 {
		d.MigrationRecords = make([]model.DBDumpMigrationRecord, 0, len(migrationRecords))
		for _, row := range migrationRecords {
			d.MigrationRecords = append(d.MigrationRecords, model.DBDumpMigrationRecord{
				Version: row.Version,
				Status:  int(row.Status),
			})
		}
	}

	if includeStats {
		if err := conn.Find(&d.StatsTotal).Error; err != nil {
			return nil, fmt.Errorf("export stats_total: %w", err)
		}
		if err := conn.Find(&d.StatsDaily).Error; err != nil {
			return nil, fmt.Errorf("export stats_daily: %w", err)
		}
		if err := conn.Find(&d.StatsHourly).Error; err != nil {
			return nil, fmt.Errorf("export stats_hourly: %w", err)
		}
		if err := conn.Find(&d.StatsModel).Error; err != nil {
			return nil, fmt.Errorf("export stats_model: %w", err)
		}
		if err := conn.Find(&d.StatsChannel).Error; err != nil {
			return nil, fmt.Errorf("export stats_channel: %w", err)
		}
		if err := conn.Find(&d.StatsAPIKey).Error; err != nil {
			return nil, fmt.Errorf("export stats_api_key: %w", err)
		}
	}

	if includeLogs {
		if err := conn.Find(&d.RelayLogs).Error; err != nil {
			return nil, fmt.Errorf("export relay_logs: %w", err)
		}
	}

	if !includeSecrets {
		redactDumpSecrets(d)
	}
	d.Manifest.ContainsSecrets = dumpContainsSecrets(d)
	clone := *d
	clone.Manifest.Checksum = ""
	if payload, err := json.Marshal(clone); err == nil {
		sum := sha256.Sum256(payload)
		d.Manifest.Checksum = hex.EncodeToString(sum[:])
	}

	return d, nil
}

func DBImportIncremental(ctx context.Context, dump *model.DBDump, mode model.DBImportMode, dryRun bool) (*model.DBImportResult, error) {
	return DBImportIncrementalWithOptions(ctx, dump, mode, dryRun, model.DBImportOptions{})
}

func DBImportIncrementalWithOptions(ctx context.Context, dump *model.DBDump, mode model.DBImportMode, dryRun bool, options model.DBImportOptions) (*model.DBImportResult, error) {
	if dump == nil {
		return nil, fmt.Errorf("empty dump")
	}
	if err := validateImportScopes(options.ImportScopes); err != nil {
		return nil, err
	}
	if err := validateImportModelMappings(options.ModelMappings); err != nil {
		return nil, err
	}

	originalDump := cloneDumpForImport(dump)
	originalManifestContainsSecrets := originalDump.Manifest.ContainsSecrets
	NormalizeLegacyDump(originalDump)
	if originalManifestContainsSecrets {
		originalDump.Manifest.ContainsSecrets = true
	}
	applyImportScopesToDump(originalDump, options.ImportScopes)
	dump = cloneDumpForImport(originalDump)
	applyModelMappingsToDump(dump, options.ModelMappings)

	if dump.Version != 0 && dump.Version != dbDumpVersion {
		return nil, fmt.Errorf("unsupported dump version: %d", dump.Version)
	}
	if err := validateImportChannels(dump.Channels); err != nil {
		return nil, err
	}
	if err := validateImportSettings(dump.Settings); err != nil {
		return nil, err
	}

	rawMode := strings.TrimSpace(string(mode))
	if rawMode != "" {
		normalizedMode := model.NormalizeDBImportMode(rawMode)
		if !model.IsValidDBImportMode(normalizedMode) {
			return nil, fmt.Errorf("unsupported import mode: %s", rawMode)
		}
	}
	mode = model.DefaultDBImportMode(rawMode)
	routingScopeEnabled := options.ImportScopes == nil || options.ImportScopes.Routing
	settingsScopeEnabled := options.ImportScopes == nil || options.ImportScopes.Settings
	modelsScopeEnabled := options.ImportScopes == nil || options.ImportScopes.Models
	apiKeysScopeEnabled := options.ImportScopes == nil || options.ImportScopes.APIKeys

	conn := db.GetDB().WithContext(ctx)
	state, err := loadCurrentImportState(conn)
	if err != nil {
		return nil, err
	}
	var postImportHealthTargets []ChannelModelHealthCheckTarget
	res := &model.DBImportResult{
		RowsAffected:  map[string]int64{},
		DryRun:        dryRun,
		Mode:          mode,
		Manifest:      &dump.Manifest,
		Compatibility: buildImportCompatibility(originalDump, dump, state, mode, options.ImportScopes, options.ModelMappings),
	}
	filteredChannelKeys, skippedChannelKeys := filterImportableChannelKeys(dump.ChannelKeys)
	filteredAPIKeys, skippedAPIKeys := filterImportableAPIKeys(dump.APIKeys)
	if skippedChannelKeys > 0 {
		res.Warnings = append(res.Warnings, fmt.Sprintf("skipped %d channel keys without credentials", skippedChannelKeys))
	}
	if skippedAPIKeys > 0 {
		res.Warnings = append(res.Warnings, fmt.Sprintf("skipped %d API keys without credentials", skippedAPIKeys))
	}
	if !dump.Manifest.ContainsSecrets && (skippedChannelKeys > 0 || skippedAPIKeys > 0) {
		res.Warnings = append(res.Warnings, snapshotCredentialsRedactedWarning)
	}
	if dump.Manifest.SchemaVersion != "" && dump.Manifest.SchemaVersion != "v1" {
		res.Warnings = append(res.Warnings, fmt.Sprintf("schema version %s may require compatibility handling", dump.Manifest.SchemaVersion))
	}
	if dump.Manifest.ExportSource != "" && dump.Manifest.ExportSource != "octopus" {
		res.Warnings = append(res.Warnings, fmt.Sprintf("export source is %s", dump.Manifest.ExportSource))
	}
	if originalDump.LegacyHints != nil && originalDump.LegacyHints.Legacy {
		res.Warnings = append(res.Warnings, "legacy backup compatibility mode enabled")
	}
	if len(dump.Channels) == 0 {
		res.Warnings = append(res.Warnings, "dump contains no channels")
	}
	if len(dump.Groups) == 0 {
		res.Warnings = append(res.Warnings, "dump contains no groups")
	}
	if len(dump.ChannelKeys) == 0 {
		res.Warnings = append(res.Warnings, "dump contains no channel keys")
	}

	if dryRun {
		sort.Strings(res.Warnings)
		return res, nil
	}

	if !options.SkipPreImportSnapshot {
		if _, _, err := savePreImportSnapshot(ctx); err != nil {
			return nil, fmt.Errorf("save pre-import snapshot: %w", err)
		}
	}

	err = conn.Transaction(func(tx *gorm.DB) error {
		var n int64
		if mode == model.DBImportModeReplace && routingScopeEnabled {
			if n, err = replaceGroups(tx, dump.Groups); err != nil {
				return fmt.Errorf("replace groups: %w", err)
			}
			res.RowsAffected["replaced_groups"] = n
			if n, err = replaceChannels(tx, dump.Channels); err != nil {
				return fmt.Errorf("replace channels: %w", err)
			}
			res.RowsAffected["replaced_channels"] = n
		}

		channelIDMap, n, err := importChannels(tx, dump.Channels, mode, dump.LegacyHints)
		if err != nil {
			return fmt.Errorf("import channels: %w", err)
		}
		res.RowsAffected["channels"] = n

		if n, err = importUsers(tx, dump.Users, mode); err != nil {
			return fmt.Errorf("import users: %w", err)
		}
		res.RowsAffected["users"] = n
		if mode == model.DBImportModeSkip {
			for _, row := range dump.Channels {
				if _, ok := state.channelsByName[strings.TrimSpace(row.Name)]; ok {
					res.Warnings = append(res.Warnings, fmt.Sprintf("skip mode preserved existing channel:%s", strings.TrimSpace(row.Name)))
				}
			}
		}

		if mode == model.DBImportModeReplace {
			if n, err = replaceChannelKeys(tx, channelIDMap); err != nil {
				return fmt.Errorf("replace channel_keys: %w", err)
			}
			res.RowsAffected["replaced_channel_keys"] = n
		}

		groupIDMap, n, err := importGroups(tx, dump.Groups, mode, dump.LegacyHints)
		if err != nil {
			return fmt.Errorf("import groups: %w", err)
		}
		res.RowsAffected["groups"] = n
		if mode == model.DBImportModeSkip {
			for _, row := range dump.Groups {
				if _, ok := state.groupsByName[strings.TrimSpace(row.Name)]; ok {
					res.Warnings = append(res.Warnings, fmt.Sprintf("skip mode preserved existing group:%s", strings.TrimSpace(row.Name)))
				}
			}
		}

		if mode == model.DBImportModeReplace {
			if n, err = replaceGroupItems(tx, groupIDMap); err != nil {
				return fmt.Errorf("replace group_items: %w", err)
			}
			res.RowsAffected["replaced_group_items"] = n
		}

		preparedChannelKeys, keyWarnings, err := prepareChannelKeysForImport(tx, filteredChannelKeys, dump.Channels, channelIDMap, state, mode, dump.LegacyHints)
		if err != nil {
			return fmt.Errorf("prepare channel_keys: %w", err)
		}
		res.Warnings = append(res.Warnings, keyWarnings...)
		channelKeyIDMap := map[int]int{}
		if n, channelKeyIDMap, err = importPreparedChannelKeys(tx, preparedChannelKeys, mode); err != nil {
			return fmt.Errorf("import channel_keys: %w", err)
		}
		res.RowsAffected["channel_keys"] = n

		preparedRouteTargetOverrides, overrideWarnings, err := prepareRouteTargetOverridesForImport(tx, dump.RouteTargetOverrides, dump.Channels, filteredChannelKeys, channelIDMap, channelKeyIDMap, state, mode)
		if err != nil {
			return fmt.Errorf("prepare route_target_overrides: %w", err)
		}
		res.Warnings = append(res.Warnings, overrideWarnings...)
		if n, err = importPreparedRouteTargetOverrides(tx, preparedRouteTargetOverrides, mode); err != nil {
			return fmt.Errorf("import route_target_overrides: %w", err)
		}
		res.RowsAffected["route_target_overrides"] = n

		remappedGroupItems, itemWarnings := remapGroupItemsForImport(dump.GroupItems, dump.Groups, dump.Channels, groupIDMap, channelIDMap, state, mode)
		res.Warnings = append(res.Warnings, itemWarnings...)
		postImportHealthTargets = buildImportedHealthCheckTargets(dump.Groups, groupIDMap, remappedGroupItems)
		if n, err = importGroupItems(tx, remappedGroupItems, mode); err != nil {
			return fmt.Errorf("import group_items: %w", err)
		}
		res.RowsAffected["group_items"] = n

		if mode == model.DBImportModeSkip {
			if n, err = createDoNothingOnColumns(tx, dump.LLMInfos, []clause.Column{{Name: "name"}}); err != nil {
				return fmt.Errorf("import llm_infos: %w", err)
			}
		} else {
			if mode == model.DBImportModeReplace && modelsScopeEnabled {
				if n, err = replaceLLMInfos(tx, dump.LLMInfos); err != nil {
					return fmt.Errorf("replace llm_infos: %w", err)
				}
				res.RowsAffected["replaced_llm_infos"] = n
			}
			if n, err = importLLMInfos(tx, dump.LLMInfos, state, dump.LegacyHints); err != nil {
				return fmt.Errorf("import llm_infos: %w", err)
			}
		}
		res.RowsAffected["llm_infos"] = n

		preparedAPIKeys, apiKeyWarnings, err := prepareAPIKeysForImport(tx, filteredAPIKeys)
		if err != nil {
			return fmt.Errorf("prepare api_keys: %w", err)
		}
		res.Warnings = append(res.Warnings, apiKeyWarnings...)
		if mode == model.DBImportModeReplace && apiKeysScopeEnabled && dump.Manifest.ContainsSecrets {
			if n, err = replaceAPIKeys(tx, filteredAPIKeys); err != nil {
				return fmt.Errorf("replace api_keys: %w", err)
			}
			res.RowsAffected["replaced_api_keys"] = n
		}
		apiKeyIDMap := map[int]int{}
		if n, apiKeyIDMap, err = importPreparedAPIKeys(tx, preparedAPIKeys, mode); err != nil {
			return fmt.Errorf("import api_keys: %w", err)
		}
		res.RowsAffected["api_keys"] = n

		if mode == model.DBImportModeSkip {
			if n, err = createDoNothingOnColumns(tx, dump.Settings, []clause.Column{{Name: "key"}}); err != nil {
				return fmt.Errorf("import settings: %w", err)
			}
		} else {
			if mode == model.DBImportModeReplace && settingsScopeEnabled {
				if n, err = replaceSettings(tx, dump.Settings); err != nil {
					return fmt.Errorf("replace settings: %w", err)
				}
				res.RowsAffected["replaced_settings"] = n
			}
			if n, err = createUpsertSettings(tx, dump.Settings); err != nil {
				return fmt.Errorf("import settings: %w", err)
			}
		}
		res.RowsAffected["settings"] = n

		migrationRecordsAffected := int64(0)
		if mode == model.DBImportModeReplace && options.ImportScopes == nil {
			if n, err = replaceMigrationRecords(tx, dump.MigrationRecords); err != nil {
				return fmt.Errorf("replace migration_records: %w", err)
			}
			res.RowsAffected["replaced_migration_records"] = n
			migrationRecordsAffected = n
		}
		if options.ImportScopes == nil {
			if n, err = importMigrationRecords(tx, dump.MigrationRecords, mode); err != nil {
				return fmt.Errorf("import migration_records: %w", err)
			}
			migrationRecordsAffected = n
		}
		res.RowsAffected["migration_records"] = migrationRecordsAffected

		if dump.IncludeStats {
			remappedStatsModel, statsModelWarnings := remapStatsModelsForImport(dump.StatsModel, channelIDMap)
			res.Warnings = append(res.Warnings, statsModelWarnings...)
			remappedStatsChannel, statsChannelWarnings := remapStatsChannelsForImport(dump.StatsChannel, channelIDMap)
			res.Warnings = append(res.Warnings, statsChannelWarnings...)
			remappedStatsAPIKey, statsAPIKeyWarnings := remapStatsAPIKeysForImport(dump.StatsAPIKey, apiKeyIDMap)
			res.Warnings = append(res.Warnings, statsAPIKeyWarnings...)
			if mode == model.DBImportModeSkip {
				if n, err = createDoNothingOnColumns(tx, dump.StatsTotal, []clause.Column{{Name: "id"}}); err != nil {
					return fmt.Errorf("import stats_total: %w", err)
				}
				res.RowsAffected["stats_total"] = n
				if n, err = createDoNothingOnColumns(tx, dump.StatsDaily, []clause.Column{{Name: "date"}}); err != nil {
					return fmt.Errorf("import stats_daily: %w", err)
				}
				res.RowsAffected["stats_daily"] = n
				if n, err = createDoNothingOnColumns(tx, dump.StatsHourly, []clause.Column{{Name: "hour"}}); err != nil {
					return fmt.Errorf("import stats_hourly: %w", err)
				}
				res.RowsAffected["stats_hourly"] = n
				if n, err = createDoNothingOnColumns(tx, remappedStatsModel, []clause.Column{{Name: "id"}}); err != nil {
					return fmt.Errorf("import stats_model: %w", err)
				}
				res.RowsAffected["stats_model"] = n
				if n, err = createDoNothingOnColumns(tx, remappedStatsChannel, []clause.Column{{Name: "channel_id"}}); err != nil {
					return fmt.Errorf("import stats_channel: %w", err)
				}
				res.RowsAffected["stats_channel"] = n
				if n, err = createDoNothingOnColumns(tx, remappedStatsAPIKey, []clause.Column{{Name: "api_key_id"}}); err != nil {
					return fmt.Errorf("import stats_api_key: %w", err)
				}
				res.RowsAffected["stats_api_key"] = n
			} else {
				if n, err = createUpsertAll(tx, dump.StatsTotal, []clause.Column{{Name: "id"}}); err != nil {
					return fmt.Errorf("import stats_total: %w", err)
				}
				res.RowsAffected["stats_total"] = n
				if n, err = createUpsertAll(tx, dump.StatsDaily, []clause.Column{{Name: "date"}}); err != nil {
					return fmt.Errorf("import stats_daily: %w", err)
				}
				res.RowsAffected["stats_daily"] = n
				if n, err = createUpsertAll(tx, dump.StatsHourly, []clause.Column{{Name: "hour"}}); err != nil {
					return fmt.Errorf("import stats_hourly: %w", err)
				}
				res.RowsAffected["stats_hourly"] = n
				if n, err = createUpsertAll(tx, remappedStatsModel, []clause.Column{{Name: "id"}}); err != nil {
					return fmt.Errorf("import stats_model: %w", err)
				}
				res.RowsAffected["stats_model"] = n
				if n, err = createUpsertAll(tx, remappedStatsChannel, []clause.Column{{Name: "channel_id"}}); err != nil {
					return fmt.Errorf("import stats_channel: %w", err)
				}
				res.RowsAffected["stats_channel"] = n
				if n, err = createUpsertAll(tx, remappedStatsAPIKey, []clause.Column{{Name: "api_key_id"}}); err != nil {
					return fmt.Errorf("import stats_api_key: %w", err)
				}
				res.RowsAffected["stats_api_key"] = n
			}
		}

		if dump.IncludeLogs {
			remappedRelayLogs, relayLogWarnings := remapRelayLogsForImport(dump.RelayLogs, channelIDMap, channelKeyIDMap)
			res.Warnings = append(res.Warnings, relayLogWarnings...)
			if n, err = createDoNothing(tx, remappedRelayLogs); err != nil {
				return fmt.Errorf("import relay_logs: %w", err)
			}
			res.RowsAffected["relay_logs"] = n
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	validation, cleanedCount, validationWarnings, err := buildPostImportValidationReport(ctx, dump)
	if err != nil {
		return nil, err
	}
	if validation != nil {
		validation.HealthCheck = buildPostImportHealthCheck(ctx, res.Compatibility, postImportHealthTargets)
	}
	res.PostImportValidation = validation
	if cleanedCount > 0 {
		res.RowsAffected["cleaned_group_items"] = cleanedCount
	}
	res.Warnings = append(res.Warnings, validationWarnings...)
	sort.Strings(res.Warnings)
	return res, nil
}

func validateImportSettings(rows []model.Setting) error {
	if len(rows) == 0 {
		return nil
	}
	knownKeys := make(map[model.SettingKey]struct{}, len(model.DefaultSettings()))
	for _, setting := range model.DefaultSettings() {
		knownKeys[setting.Key] = struct{}{}
	}
	for _, row := range rows {
		if _, ok := knownKeys[row.Key]; !ok {
			return fmt.Errorf("unknown setting key: %s", row.Key)
		}
		if err := row.Validate(); err != nil {
			return fmt.Errorf("invalid setting %s: %w", row.Key, err)
		}
	}
	return nil
}

func validateImportChannels(rows []model.Channel) error {
	for _, row := range rows {
		if err := model.ValidateChannelProxy(row.ChannelProxy); err != nil {
			name := strings.TrimSpace(row.Name)
			if name == "" {
				name = fmt.Sprintf("id:%d", row.ID)
			}
			return fmt.Errorf("invalid channel %s proxy: %w", name, err)
		}
	}
	return nil
}

func DBRollbackLatestImportSnapshot(ctx context.Context) (*model.DBRollbackResult, error) {
	metadata, dump, err := loadLatestImportSnapshot()
	if err != nil {
		return nil, err
	}
	return rollbackToImportSnapshot(ctx, metadata, dump, nil)
}

func DBListImportSnapshots() ([]model.DBImportSnapshotInfo, error) {
	dir, err := importSnapshotDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.DBImportSnapshotInfo{}, nil
		}
		return nil, err
	}
	latestMetadata, _, latestErr := loadLatestImportSnapshot()
	if latestErr != nil && !strings.Contains(latestErr.Error(), "latest import snapshot not found") {
		return nil, latestErr
	}
	latestSnapshotPath := ""
	if latestMetadata != nil {
		latestSnapshotPath = strings.TrimSpace(latestMetadata.SnapshotPath)
	}
	items := make([]model.DBImportSnapshotInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSpace(entry.Name())
		if name == "" || name == importSnapshotLatestFilename || !strings.HasSuffix(strings.ToLower(name), ".json") {
			continue
		}
		path := filepath.Join(dir, name)
		fileInfo, err := entry.Info()
		if err != nil {
			return nil, err
		}
		metadata, err := loadImportSnapshotMetadataFromFile(path, fileInfo)
		if err != nil {
			return nil, err
		}
		items = append(items, model.DBImportSnapshotInfo{
			SnapshotPath: metadata.SnapshotPath,
			SnapshotName: metadata.SnapshotName,
			ImportedAt:   metadata.ImportedAt,
			SizeBytes:    fileInfo.Size(),
			IsLatest:     sameImportSnapshotPath(metadata.SnapshotPath, latestSnapshotPath),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if !items[i].ImportedAt.Equal(items[j].ImportedAt) {
			return items[i].ImportedAt.After(items[j].ImportedAt)
		}
		return items[i].SnapshotName > items[j].SnapshotName
	})
	return items, nil
}

func DBRollbackImportSnapshot(ctx context.Context, snapshotName string, scopes *model.DBImportScopes) (*model.DBRollbackResult, error) {
	if err := validateImportScopes(scopes); err != nil {
		return nil, err
	}
	metadata, dump, err := loadImportSnapshotByName(snapshotName)
	if err != nil {
		return nil, err
	}
	return rollbackToImportSnapshot(ctx, metadata, dump, scopes)
}

func DBPreviewRollbackImportSnapshot(ctx context.Context, snapshotName string, scopes *model.DBImportScopes) (*model.DBRollbackPreviewResult, error) {
	if err := validateImportScopes(scopes); err != nil {
		return nil, err
	}
	metadata, dump, err := loadImportSnapshotByName(snapshotName)
	if err != nil {
		return nil, err
	}
	workingDump := cloneDumpForImport(dump)
	applyImportScopesToDump(workingDump, scopes)
	conn := db.GetDB().WithContext(ctx)
	state, err := loadCurrentImportState(conn)
	if err != nil {
		return nil, err
	}
	compatibility := buildImportCompatibility(nil, workingDump, state, model.DBImportModeReplace, scopes, nil)
	rowsSummary := buildDumpRowsSummary(workingDump)
	previewWarnings := buildRollbackPreviewWarnings(workingDump, compatibility)
	return &model.DBRollbackPreviewResult{
		SnapshotPath:    metadata.SnapshotPath,
		SnapshotName:    metadata.SnapshotName,
		ImportedAt:      metadata.ImportedAt,
		AppliedScopes:   cloneImportScopes(scopes),
		Manifest:        &workingDump.Manifest,
		Compatibility:   compatibility,
		RowsSummary:     rowsSummary,
		PreviewWarnings: previewWarnings,
	}, nil
}

func rollbackToImportSnapshot(ctx context.Context, metadata *importSnapshotMetadata, dump *model.DBDump, scopes *model.DBImportScopes) (*model.DBRollbackResult, error) {
	if metadata == nil || dump == nil {
		return nil, fmt.Errorf("import snapshot is empty")
	}
	workingDump := cloneDumpForImport(dump)
	applyImportScopesToDump(workingDump, scopes)
	if isFullImportScopes(scopes) {
		if err := resetDatabaseForRollback(ctx); err != nil {
			return nil, err
		}
	}
	result, err := DBImportIncrementalWithOptions(ctx, workingDump, model.DBImportModeReplace, false, model.DBImportOptions{
		ImportScopes:          cloneImportScopes(scopes),
		SkipPreImportSnapshot: true,
	})
	if err != nil {
		return nil, err
	}
	return &model.DBRollbackResult{
		SnapshotPath:  metadata.SnapshotPath,
		SnapshotName:  metadata.SnapshotName,
		ImportedAt:    metadata.ImportedAt,
		AppliedScopes: cloneImportScopes(scopes),
		Result:        result,
	}, nil
}

func resetDatabaseForRollback(ctx context.Context) error {
	conn := db.GetDB().WithContext(ctx)
	return conn.Transaction(func(tx *gorm.DB) error {
		orderedDeletes := []any{
			&model.RelayLog{},
			&model.StatsAPIKey{},
			&model.StatsChannel{},
			&model.StatsModel{},
			&model.StatsHourly{},
			&model.StatsDaily{},
			&model.StatsTotal{},
			&model.GroupItem{},
			&model.Group{},
			&model.RouteTargetOverride{},
			&model.ChannelKey{},
			&model.Channel{},
			&model.User{},
			&model.APIKey{},
			&model.LLMInfo{},
			&model.Setting{},
			&migrate.MigrationRecord{},
		}
		for _, table := range orderedDeletes {
			if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func savePreImportSnapshot(ctx context.Context) (string, string, error) {
	dump, err := DBExportAll(ctx, true, true, true)
	if err != nil {
		return "", "", err
	}
	payload, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		return "", "", err
	}
	dir, err := importSnapshotDir()
	if err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(dir, importSnapshotDirPerm); err != nil {
		return "", "", err
	}
	if err := os.Chmod(dir, importSnapshotDirPerm); err != nil {
		return "", "", err
	}
	name := buildImportSnapshotFilename(time.Now().UTC())
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, payload, importSnapshotFilePerm); err != nil {
		return "", "", err
	}
	if err := os.Chmod(path, importSnapshotFilePerm); err != nil {
		return "", "", err
	}
	metadata := importSnapshotMetadata{SnapshotPath: path, SnapshotName: name, ImportedAt: time.Now().UTC()}
	metadataPayload, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", "", err
	}
	latestPath := filepath.Join(dir, importSnapshotLatestFilename)
	if err := os.WriteFile(latestPath, metadataPayload, importSnapshotFilePerm); err != nil {
		return "", "", err
	}
	if err := os.Chmod(latestPath, importSnapshotFilePerm); err != nil {
		return "", "", err
	}
	return path, name, nil
}

func buildImportSnapshotFilename(ts time.Time) string {
	utc := ts.UTC()
	return fmt.Sprintf("pre-import-%s-%09d.json", utc.Format("20060102T150405Z"), utc.Nanosecond())
}

func loadLatestImportSnapshot() (*importSnapshotMetadata, *model.DBDump, error) {
	dir, err := importSnapshotDir()
	if err != nil {
		return nil, nil, err
	}
	metadataPath := filepath.Join(dir, importSnapshotLatestFilename)
	metadataPayload, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("latest import snapshot not found")
		}
		return nil, nil, err
	}
	var metadata importSnapshotMetadata
	if err := json.Unmarshal(metadataPayload, &metadata); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(metadata.SnapshotPath) == "" {
		return nil, nil, fmt.Errorf("latest import snapshot metadata is missing snapshot_path")
	}
	resolvedSnapshotPath, err := resolveImportSnapshotPath(metadata.SnapshotPath)
	if err != nil {
		return nil, nil, err
	}
	metadata.SnapshotPath = resolvedSnapshotPath
	dumpPayload, err := os.ReadFile(resolvedSnapshotPath)
	if err != nil {
		return nil, nil, err
	}
	var dump model.DBDump
	if err := json.Unmarshal(dumpPayload, &dump); err != nil {
		return nil, nil, err
	}
	return &metadata, &dump, nil
}

func loadImportSnapshotByName(snapshotName string) (*importSnapshotMetadata, *model.DBDump, error) {
	trimmedName := strings.TrimSpace(snapshotName)
	if trimmedName == "" {
		return nil, nil, fmt.Errorf("snapshot_name is required")
	}
	if filepath.Base(trimmedName) != trimmedName || strings.Contains(trimmedName, "..") {
		return nil, nil, fmt.Errorf("invalid snapshot_name")
	}
	dir, err := importSnapshotDir()
	if err != nil {
		return nil, nil, err
	}
	path := filepath.Join(dir, trimmedName)
	return loadImportSnapshotByPath(path)
}

func loadImportSnapshotByPath(path string) (*importSnapshotMetadata, *model.DBDump, error) {
	resolvedPath, err := resolveImportSnapshotPath(path)
	if err != nil {
		return nil, nil, err
	}
	dumpPayload, err := os.ReadFile(resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("import snapshot not found: %s", filepath.Base(resolvedPath))
		}
		return nil, nil, err
	}
	var dump model.DBDump
	if err := json.Unmarshal(dumpPayload, &dump); err != nil {
		return nil, nil, err
	}
	fileInfo, err := os.Stat(resolvedPath)
	if err != nil {
		return nil, nil, err
	}
	metadata, err := loadImportSnapshotMetadataFromFile(resolvedPath, fileInfo)
	if err != nil {
		return nil, nil, err
	}
	return metadata, &dump, nil
}

func resolveImportSnapshotPath(path string) (string, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return "", fmt.Errorf("snapshot path is required")
	}
	baseDir, err := importSnapshotDir()
	if err != nil {
		return "", err
	}
	resolvedBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	if realBaseDir, err := filepath.EvalSymlinks(resolvedBaseDir); err == nil {
		resolvedBaseDir = realBaseDir
	}
	resolvedPath, err := filepath.Abs(trimmedPath)
	if err != nil {
		return "", err
	}
	if realPath, err := filepath.EvalSymlinks(resolvedPath); err == nil {
		resolvedPath = realPath
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if !strings.HasPrefix(strings.ToLower(resolvedPath), strings.ToLower(resolvedBaseDir+string(os.PathSeparator))) && !strings.EqualFold(resolvedPath, resolvedBaseDir) {
		return "", fmt.Errorf("snapshot path is outside import snapshot directory")
	}
	if strings.EqualFold(filepath.Base(resolvedPath), importSnapshotLatestFilename) {
		return "", fmt.Errorf("snapshot metadata file cannot be used as rollback target")
	}
	return resolvedPath, nil
}

func loadImportSnapshotMetadataFromFile(path string, fileInfo os.FileInfo) (*importSnapshotMetadata, error) {
	resolvedPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	metadata := &importSnapshotMetadata{SnapshotPath: resolvedPath, SnapshotName: filepath.Base(resolvedPath)}
	if fileInfo != nil {
		metadata.ImportedAt = fileInfo.ModTime().UTC()
	}
	if strings.TrimSpace(metadata.SnapshotName) == "" {
		return nil, fmt.Errorf("snapshot file name is empty")
	}
	return metadata, nil
}

func sameImportSnapshotPath(left, right string) bool {
	trimmedLeft := strings.TrimSpace(left)
	trimmedRight := strings.TrimSpace(right)
	if trimmedLeft == "" || trimmedRight == "" {
		return false
	}
	resolvedLeft, err := filepath.Abs(trimmedLeft)
	if err != nil {
		resolvedLeft = trimmedLeft
	}
	resolvedRight, err := filepath.Abs(trimmedRight)
	if err != nil {
		resolvedRight = trimmedRight
	}
	return strings.EqualFold(resolvedLeft, resolvedRight)
}

func buildDumpRowsSummary(dump *model.DBDump) map[string]int {
	if dump == nil {
		return map[string]int{}
	}
	return map[string]int{
		"channels":               len(dump.Channels),
		"channel_keys":           len(dump.ChannelKeys),
		"route_target_overrides": len(dump.RouteTargetOverrides),
		"groups":                 len(dump.Groups),
		"group_items":            len(dump.GroupItems),
		"llm_infos":              len(dump.LLMInfos),
		"api_keys":               len(dump.APIKeys),
		"settings":               len(dump.Settings),
		"users":                  len(dump.Users),
		"migration_records":      len(dump.MigrationRecords),
		"stats_total":            len(dump.StatsTotal),
		"stats_daily":            len(dump.StatsDaily),
		"stats_hourly":           len(dump.StatsHourly),
		"stats_model":            len(dump.StatsModel),
		"stats_channel":          len(dump.StatsChannel),
		"stats_api_key":          len(dump.StatsAPIKey),
		"relay_logs":             len(dump.RelayLogs),
	}
}
