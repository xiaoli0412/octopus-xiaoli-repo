package op

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

const (
	inventorySourceChannelDeclared = "channel_declared"
	inventorySourceChannelKey      = "channel_key_allowed"
	inventorySourceGroupBound      = "group_bound"
	inventorySourceCombined        = "combined"
)

type serviceableAccumulator struct {
	item                model.ServiceableModelInventoryItem
	sources             map[string]struct{}
	capability          map[string]struct{}
	requestUnrestricted bool
}

type selectableAccumulator struct {
	item                model.SelectableGroupModelInventoryItem
	sources             map[string]struct{}
	capability          map[string]struct{}
	channels            map[int]struct{}
	enabled             map[int]struct{}
	requestUnrestricted bool
}

type routableAccumulator struct {
	item                model.RoutableModelInventoryItem
	sources             map[string]struct{}
	capability          map[string]struct{}
	channels            map[int]struct{}
	enabled             map[int]struct{}
	requestUnrestricted bool
}

func addSource(sources map[string]struct{}, source string) {
	if source == "" {
		return
	}
	sources[source] = struct{}{}
}

func inventorySourceLabel(sources map[string]struct{}) string {
	if len(sources) == 0 {
		return ""
	}
	if len(sources) > 1 {
		return inventorySourceCombined
	}
	for source := range sources {
		return source
	}
	return ""
}

func sortedSetValues(values map[string]struct{}) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func declaredChannelModels(channel model.Channel) []string {
	return splitModelNames(channel.Model, channel.CustomModel)
}

func serviceableModelsForChannel(channel model.Channel) []model.ServiceableModelInventoryItem {
	declaredModels := declaredChannelModels(channel)
	declaredSet := make(map[string]struct{}, len(declaredModels))
	for _, name := range declaredModels {
		declaredSet[strings.ToLower(strings.TrimSpace(name))] = struct{}{}
	}

	byName := make(map[string]*serviceableAccumulator)
	ensure := func(modelName string) *serviceableAccumulator {
		key := strings.ToLower(strings.TrimSpace(modelName))
		if key == "" {
			return nil
		}
		if existing, ok := byName[key]; ok {
			return existing
		}
		acc := &serviceableAccumulator{
			item: model.ServiceableModelInventoryItem{
				Name:        strings.TrimSpace(modelName),
				Enabled:     channel.Enabled,
				ChannelID:   channel.ID,
				ChannelName: channel.Name,
			},
			sources:    make(map[string]struct{}),
			capability: make(map[string]struct{}),
		}
		byName[key] = acc
		return acc
	}

	for _, key := range channel.Keys {
		if !key.Enabled || strings.TrimSpace(key.ChannelKey) == "" {
			continue
		}

		allowedModels := model.ChannelKeyAllowedModelsList(key.AllowedModels)
		source := inventorySourceChannelKey
		if len(allowedModels) == 0 {
			allowedModels = declaredModels
			source = inventorySourceChannelDeclared
		}
		if len(allowedModels) == 0 {
			continue
		}

		for _, modelName := range allowedModels {
			trimmed := strings.TrimSpace(modelName)
			if trimmed == "" {
				continue
			}
			if len(model.ChannelKeyAllowedModelsList(key.AllowedModels)) == 0 {
				if _, declared := declaredSet[strings.ToLower(trimmed)]; !declared {
					continue
				}
			}
			requestCapability := channel.RequestCapabilityForModel(trimmed)
			if !channel.KeyCanServeRequest(key, trimmed, requestCapability) {
				continue
			}
			acc := ensure(trimmed)
			if acc == nil {
				continue
			}
			acc.item.KeyCount++
			addSource(acc.sources, source)
			if requestCapability == "" {
				acc.requestUnrestricted = true
			} else {
				acc.capability[requestCapability] = struct{}{}
			}
		}
	}

	items := make([]model.ServiceableModelInventoryItem, 0, len(byName))
	for _, acc := range byName {
		if acc.item.KeyCount <= 0 {
			continue
		}
		acc.item.InventorySource = inventorySourceLabel(acc.sources)
		if !acc.requestUnrestricted {
			acc.item.RequestCapabilities = sortedSetValues(acc.capability)
		}
		items = append(items, acc.item)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Name != items[j].Name {
			return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		}
		return items[i].ChannelID < items[j].ChannelID
	})
	return items
}

func addSelectable(selectable map[string]*selectableAccumulator, name string, source string) *selectableAccumulator {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil
	}
	key := strings.ToLower(trimmed)
	if existing, ok := selectable[key]; ok {
		addSource(existing.sources, source)
		return existing
	}
	acc := &selectableAccumulator{
		item: model.SelectableGroupModelInventoryItem{
			Name: trimmed,
		},
		sources:    make(map[string]struct{}),
		capability: make(map[string]struct{}),
		channels:   make(map[int]struct{}),
		enabled:    make(map[int]struct{}),
	}
	addSource(acc.sources, source)
	selectable[key] = acc
	return acc
}

func buildCapabilityInventory(ctx context.Context) (model.CapabilityInventory, error) {
	inventory := model.CapabilityInventory{
		ServiceableModels: make([]model.ServiceableModelInventoryItem, 0),
		SelectableModels:  make([]model.SelectableGroupModelInventoryItem, 0),
		RoutableModels:    make([]model.RoutableModelInventoryItem, 0),
	}

	selectable := make(map[string]*selectableAccumulator)
	serviceableByChannelModel := make(map[string]model.ServiceableModelInventoryItem)

	for _, channel := range channelCache.Values() {
		items := serviceableModelsForChannel(channel)
		inventory.ServiceableModels = append(inventory.ServiceableModels, items...)
		for _, item := range items {
			serviceableByChannelModel[fmt.Sprintf("%d|%s", item.ChannelID, strings.ToLower(strings.TrimSpace(item.Name)))] = item
			acc := addSelectable(selectable, item.Name, item.InventorySource)
			if acc == nil {
				continue
			}
			acc.item.KeyCount += item.KeyCount
			acc.channels[item.ChannelID] = struct{}{}
			if item.Enabled {
				acc.enabled[item.ChannelID] = struct{}{}
			}
			if len(item.RequestCapabilities) == 0 {
				acc.requestUnrestricted = true
			}
			for _, capability := range item.RequestCapabilities {
				acc.capability[capability] = struct{}{}
			}
		}
	}

	groups, err := GroupList(ctx)
	if err != nil {
		return model.CapabilityInventory{}, err
	}
	routable := make(map[string]*routableAccumulator)
	for _, group := range groups {
		addSelectable(selectable, group.Name, inventorySourceGroupBound)
		for _, item := range group.Items {
			targetModel := strings.TrimSpace(item.ModelName)
			if targetModel == "" {
				targetModel = strings.TrimSpace(group.Name)
			}
			addSelectable(selectable, targetModel, inventorySourceGroupBound)
			serviceable, ok := serviceableByChannelModel[fmt.Sprintf("%d|%s", item.ChannelID, strings.ToLower(targetModel))]
			if !ok || serviceable.KeyCount <= 0 {
				continue
			}
			groupName := strings.TrimSpace(group.Name)
			if groupName == "" {
				continue
			}
			key := strings.ToLower(groupName)
			acc, ok := routable[key]
			if !ok {
				acc = &routableAccumulator{
					item: model.RoutableModelInventoryItem{
						Name:      groupName,
						GroupID:   group.ID,
						GroupName: groupName,
					},
					sources:    make(map[string]struct{}),
					capability: make(map[string]struct{}),
					channels:   make(map[int]struct{}),
					enabled:    make(map[int]struct{}),
				}
				routable[key] = acc
			}
			acc.item.KeyCount += serviceable.KeyCount
			acc.channels[serviceable.ChannelID] = struct{}{}
			if serviceable.Enabled {
				acc.enabled[serviceable.ChannelID] = struct{}{}
			}
			addSource(acc.sources, inventorySourceGroupBound)
			addSource(acc.sources, serviceable.InventorySource)
			if len(serviceable.RequestCapabilities) == 0 {
				acc.requestUnrestricted = true
			}
			for _, capability := range serviceable.RequestCapabilities {
				acc.capability[capability] = struct{}{}
			}
		}
	}

	for _, acc := range selectable {
		acc.item.ChannelCount = len(acc.channels)
		acc.item.EnabledChannelCount = len(acc.enabled)
		acc.item.InventorySource = inventorySourceLabel(acc.sources)
		if !acc.requestUnrestricted {
			acc.item.RequestCapabilities = sortedSetValues(acc.capability)
		}
		inventory.SelectableModels = append(inventory.SelectableModels, acc.item)
	}
	for _, acc := range routable {
		acc.item.ChannelCount = len(acc.channels)
		acc.item.EnabledChannelCount = len(acc.enabled)
		acc.item.InventorySource = inventorySourceLabel(acc.sources)
		if !acc.requestUnrestricted {
			acc.item.RequestCapabilities = sortedSetValues(acc.capability)
		}
		inventory.RoutableModels = append(inventory.RoutableModels, acc.item)
	}

	sort.SliceStable(inventory.ServiceableModels, func(i, j int) bool {
		left := strings.ToLower(inventory.ServiceableModels[i].Name)
		right := strings.ToLower(inventory.ServiceableModels[j].Name)
		if left != right {
			return left < right
		}
		return inventory.ServiceableModels[i].ChannelID < inventory.ServiceableModels[j].ChannelID
	})
	sort.SliceStable(inventory.SelectableModels, func(i, j int) bool {
		return strings.ToLower(inventory.SelectableModels[i].Name) < strings.ToLower(inventory.SelectableModels[j].Name)
	})
	sort.SliceStable(inventory.RoutableModels, func(i, j int) bool {
		return strings.ToLower(inventory.RoutableModels[i].Name) < strings.ToLower(inventory.RoutableModels[j].Name)
	})

	return inventory, nil
}

func CapabilityInventory(ctx context.Context) (model.CapabilityInventory, error) {
	return buildCapabilityInventory(ctx)
}

func ChannelCanServeModel(channelID int, modelName string) bool {
	channel, ok := channelCache.Get(channelID)
	if !ok {
		return false
	}
	return channel.HasConfiguredKeyForRequest(modelName, channel.RequestCapabilityForModel(modelName))
}

func ChannelCanServeRequest(channelID int, modelName string, requestFormat string) bool {
	channel, ok := channelCache.Get(channelID)
	if !ok {
		return false
	}
	return channel.HasConfiguredKeyForRequest(modelName, requestFormat)
}
