package op

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

type governanceSnapshot struct {
	GeneratedAt          time.Time                         `json:"generated_at"`
	ManagedGroupName     string                            `json:"managed_group_name"`
	Channels             []governanceSnapshotChannel       `json:"channels"`
	Groups               []governanceSnapshotGroup         `json:"groups"`
	RouteTargetOverrides []governanceSnapshotRouteTarget   `json:"route_target_overrides"`
	Models               []governanceSnapshotModel         `json:"models"`
	DynamicRouting       governanceSnapshotDynamicRouting  `json:"dynamic_routing"`
	RuntimePolicy        model.GovernanceRuntimePolicyView `json:"runtime_policy"`
	Learning             model.AIGovernanceLearningSummary `json:"learning"`
	ExecutionSource      model.AIGovernanceExecutionSource `json:"execution_source"`
	SnapshotSummary      model.GovernanceSnapshotSummary   `json:"snapshot_summary"`
}

type governanceSnapshotChannel struct {
	ID                int      `json:"id"`
	Name              string   `json:"name"`
	Enabled           bool     `json:"enabled"`
	DeclaredModels    []string `json:"declared_models"`
	ConfiguredModels  []string `json:"configured_models"`
	KeyRoutingPolicy  string   `json:"key_routing_policy"`
	KeyManagementMode string   `json:"key_management_mode"`
	KeyCount          int      `json:"key_count"`
}

type governanceSnapshotGroup struct {
	ID          int                           `json:"id"`
	Name        string                        `json:"name"`
	Mode        model.GroupMode               `json:"mode"`
	Items       []governanceSnapshotGroupItem `json:"items"`
	ModelCounts map[string]int                `json:"model_counts"`
	Issues      []string                      `json:"issues,omitempty"`
}

type governanceSnapshotGroupItem struct {
	ID        int    `json:"id"`
	GroupName string `json:"group_name"`
	ChannelID int    `json:"channel_id"`
	ModelName string `json:"model_name"`
	Priority  int    `json:"priority"`
	Weight    int    `json:"weight"`
	Valid     bool   `json:"valid"`
	Issue     string `json:"issue,omitempty"`
}

type governanceSnapshotRouteTarget struct {
	ChannelID             int    `json:"channel_id"`
	ChannelKeyID          int    `json:"channel_key_id"`
	ModelName             string `json:"model_name"`
	BillingMode           string `json:"billing_mode"`
	ProbePolicy           string `json:"probe_policy"`
	ProbeIntervalSeconds  int    `json:"probe_interval_seconds"`
	ProbeConcurrencyLimit int    `json:"probe_concurrency_limit"`
	Declared              bool   `json:"declared"`
	Issue                 string `json:"issue,omitempty"`
}

type governanceSnapshotModel struct {
	Name                  string `json:"name"`
	BillingMode           string `json:"billing_mode"`
	ProbePolicy           string `json:"probe_policy"`
	ProbeIntervalSeconds  int    `json:"probe_interval_seconds"`
	ProbeConcurrencyLimit int    `json:"probe_concurrency_limit"`
	HasPrice              bool   `json:"has_price"`
}

type governanceSnapshotDynamicRouting struct {
	Mode              string `json:"mode"`
	HealthEnabled     bool   `json:"health_enabled"`
	LearningEnabled   bool   `json:"learning_enabled"`
	RaceGlobalBudget  string `json:"race_global_budget"`
	RaceGroupBudget   string `json:"race_group_budget"`
	RaceChannelBudget string `json:"race_channel_budget"`
	RaceKeyBudget     string `json:"race_key_budget"`
	RaceProbeBudget   string `json:"race_probe_budget"`
}

func governanceBuildSnapshot(ctx context.Context) (governanceSnapshot, string, error) {
	config, err := AIAutomationConfigGet(ctx)
	if err != nil {
		return governanceSnapshot{}, "", err
	}
	managedGroupName := settingStringOrDefault(model.SettingKeyAIGovernanceManagedGroupName, "AI Governance Managed")
	channels, err := ChannelList(ctx)
	if err != nil {
		return governanceSnapshot{}, "", err
	}
	sort.SliceStable(channels, func(i, j int) bool { return channels[i].ID < channels[j].ID })
	channelModels := make(map[int]map[string]struct{}, len(channels))
	snapshotChannels := make([]governanceSnapshotChannel, 0, len(channels))
	enabledChannels := 0
	modelSet := make(map[string]struct{})
	for _, channel := range channels {
		declared := splitModelNames(channel.Model, channel.CustomModel)
		configuredSet := make(map[string]struct{})
		for _, item := range declared {
			name := strings.ToLower(strings.TrimSpace(item))
			if name == "" {
				continue
			}
			configuredSet[name] = struct{}{}
			modelSet[name] = struct{}{}
		}
		for _, key := range channel.Keys {
			if !key.Enabled || strings.TrimSpace(key.ChannelKey) == "" || strings.TrimSpace(key.AllowedModels) == "" {
				continue
			}
			for _, part := range strings.Split(key.AllowedModels, ",") {
				name := strings.ToLower(strings.TrimSpace(part))
				if name == "" {
					continue
				}
				configuredSet[name] = struct{}{}
				modelSet[name] = struct{}{}
			}
		}
		configured := make([]string, 0, len(configuredSet))
		for name := range configuredSet {
			configured = append(configured, name)
		}
		sort.Strings(declared)
		sort.Strings(configured)
		channelModels[channel.ID] = configuredSet
		if channel.Enabled {
			enabledChannels++
		}
		snapshotChannels = append(snapshotChannels, governanceSnapshotChannel{ID: channel.ID, Name: channel.Name, Enabled: channel.Enabled, DeclaredModels: declared, ConfiguredModels: configured, KeyRoutingPolicy: string(channel.KeyRoutingPolicy), KeyManagementMode: string(channel.KeyManagementMode), KeyCount: len(channel.Keys)})
	}
	groups, err := GroupList(ctx)
	if err != nil {
		return governanceSnapshot{}, "", err
	}
	sort.SliceStable(groups, func(i, j int) bool { return groups[i].ID < groups[j].ID })
	snapshotGroups := make([]governanceSnapshotGroup, 0, len(groups))
	groupItemsCount := 0
	for _, group := range groups {
		items := make([]governanceSnapshotGroupItem, 0, len(group.Items))
		issues := make([]string, 0)
		modelCounts := make(map[string]int)
		sortedItems := append([]model.GroupItem(nil), group.Items...)
		sort.SliceStable(sortedItems, func(i, j int) bool {
			if sortedItems[i].Priority == sortedItems[j].Priority {
				if sortedItems[i].ChannelID == sortedItems[j].ChannelID {
					return sortedItems[i].ModelName < sortedItems[j].ModelName
				}
				return sortedItems[i].ChannelID < sortedItems[j].ChannelID
			}
			return sortedItems[i].Priority < sortedItems[j].Priority
		})
		for _, item := range sortedItems {
			groupItemsCount++
			modelName := strings.ToLower(strings.TrimSpace(item.ModelName))
			modelCounts[modelName]++
			valid := true
			issue := ""
			if channelModelMap, ok := channelModels[item.ChannelID]; !ok {
				valid = false
				issue = "channel_missing"
			} else if _, ok := channelModelMap[modelName]; !ok {
				valid = false
				issue = "model_not_declared"
			}
			if !valid {
				issues = append(issues, fmt.Sprintf("%s/%s", item.ModelName, issue))
			}
			items = append(items, governanceSnapshotGroupItem{ID: item.ID, GroupName: group.Name, ChannelID: item.ChannelID, ModelName: item.ModelName, Priority: item.Priority, Weight: item.Weight, Valid: valid, Issue: issue})
		}
		if len(items) == 0 {
			issues = append(issues, "group_empty")
		}
		snapshotGroups = append(snapshotGroups, governanceSnapshotGroup{ID: group.ID, Name: group.Name, Mode: group.Mode, Items: items, ModelCounts: modelCounts, Issues: issues})
	}
	overrides, err := RouteTargetOverrideList(ctx)
	if err != nil {
		return governanceSnapshot{}, "", err
	}
	sort.SliceStable(overrides, func(i, j int) bool {
		if overrides[i].ChannelID == overrides[j].ChannelID {
			if overrides[i].ChannelKeyID == overrides[j].ChannelKeyID {
				return overrides[i].ModelName < overrides[j].ModelName
			}
			return overrides[i].ChannelKeyID < overrides[j].ChannelKeyID
		}
		return overrides[i].ChannelID < overrides[j].ChannelID
	})
	snapshotOverrides := make([]governanceSnapshotRouteTarget, 0, len(overrides))
	for _, row := range overrides {
		declared := false
		issue := ""
		if channelModelMap, ok := channelModels[row.ChannelID]; ok {
			_, declared = channelModelMap[strings.ToLower(strings.TrimSpace(row.ModelName))]
			if !declared {
				issue = "override_model_not_declared"
			}
		} else {
			issue = "override_channel_missing"
		}
		snapshotOverrides = append(snapshotOverrides, governanceSnapshotRouteTarget{ChannelID: row.ChannelID, ChannelKeyID: row.ChannelKeyID, ModelName: row.ModelName, BillingMode: string(row.BillingMode), ProbePolicy: string(row.ProbePolicy), ProbeIntervalSeconds: row.ProbeIntervalSeconds, ProbeConcurrencyLimit: row.ProbeConcurrencyLimit, Declared: declared, Issue: issue})
	}
	modelNames := make([]string, 0, len(modelSet))
	for name := range modelSet {
		modelNames = append(modelNames, name)
	}
	sort.Strings(modelNames)
	snapshotModels := make([]governanceSnapshotModel, 0, len(modelNames))
	missingPrices := 0
	for _, name := range modelNames {
		info, err := LLMGet(name)
		if err != nil {
			missingPrices++
			snapshotModels = append(snapshotModels, governanceSnapshotModel{Name: name, BillingMode: string(model.BillingModeUnknown), ProbePolicy: string(model.ProbePolicyPassiveOnly), ProbeIntervalSeconds: 3600, ProbeConcurrencyLimit: 1, HasPrice: false})
			continue
		}
		hasPrice := info.Input > 0 || info.Output > 0 || info.CacheRead > 0 || info.CacheWrite > 0 || info.OfficialInput > 0 || info.OfficialOutput > 0 || info.OfficialCacheRead > 0 || info.OfficialCacheWrite > 0 || info.BillingMode == model.BillingModeFree
		if !hasPrice {
			missingPrices++
		}
		snapshotModels = append(snapshotModels, governanceSnapshotModel{Name: name, BillingMode: string(info.BillingMode), ProbePolicy: string(info.ProbePolicy), ProbeIntervalSeconds: info.ProbeIntervalSeconds, ProbeConcurrencyLimit: info.ProbeConcurrencyLimit, HasPrice: hasPrice})
	}
	learningSummary, err := AIGovernanceLearningSummaryGet(ctx)
	if err != nil {
		return governanceSnapshot{}, "", err
	}
	dynamicRouting := governanceSnapshotDynamicRouting{
		Mode:              settingStringOrDefault(model.SettingKeyDynamicRoutingMode, "hybrid"),
		HealthEnabled:     settingBoolOrDefault(model.SettingKeyDynamicRoutingHealthEnabled, true),
		LearningEnabled:   settingBoolOrDefault(model.SettingKeyDynamicRoutingLearningEnabled, false),
		RaceGlobalBudget:  settingStringOrDefault(model.SettingKeyRaceGlobalBudget, "64"),
		RaceGroupBudget:   settingStringOrDefault(model.SettingKeyRaceGroupBudget, "8"),
		RaceChannelBudget: settingStringOrDefault(model.SettingKeyRaceChannelBudget, "4"),
		RaceKeyBudget:     settingStringOrDefault(model.SettingKeyRaceKeyBudget, "2"),
		RaceProbeBudget:   settingStringOrDefault(model.SettingKeyRaceProbeBudget, "16"),
	}
	runtimePolicy := GovernanceRuntimePolicyGet()
	highlights := []string{
		fmt.Sprintf("%d channels / %d enabled", len(snapshotChannels), enabledChannels),
		fmt.Sprintf("%d groups / %d group items", len(snapshotGroups), groupItemsCount),
		fmt.Sprintf("%d route target overrides", len(snapshotOverrides)),
	}
	if missingPrices > 0 {
		highlights = append(highlights, fmt.Sprintf("%d models missing price", missingPrices))
	}
	if learningSummary.SampleCount > 0 {
		highlights = append(highlights, fmt.Sprintf("learning samples %d", learningSummary.SampleCount))
	}
	snapshot := governanceSnapshot{
		GeneratedAt:          time.Now(),
		ManagedGroupName:     managedGroupName,
		Channels:             snapshotChannels,
		Groups:               snapshotGroups,
		RouteTargetOverrides: snapshotOverrides,
		Models:               snapshotModels,
		DynamicRouting:       dynamicRouting,
		RuntimePolicy:        runtimePolicy,
		Learning:             learningSummary,
		ExecutionSource:      governanceExecutionSourceFromConfig(config),
		SnapshotSummary:      model.GovernanceSnapshotSummary{Channels: len(snapshotChannels), EnabledChannels: enabledChannels, Groups: len(snapshotGroups), GroupItems: groupItemsCount, RouteTargetOverrides: len(snapshotOverrides), Models: len(snapshotModels), MissingPrices: missingPrices, ActiveSourceMode: config.ConfigSourceMode, ActiveSourceLabel: governanceExecutionSourceFromConfig(config).Label, Highlights: highlights},
	}
	checksum, err := governanceSnapshotChecksum(snapshot)
	if err != nil {
		return governanceSnapshot{}, "", err
	}
	return snapshot, checksum, nil
}

func governanceSnapshotChecksum(snapshot governanceSnapshot) (string, error) {
	snapshot.GeneratedAt = time.Time{}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}
