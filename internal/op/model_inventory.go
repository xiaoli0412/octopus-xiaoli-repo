package op

import (
	"context"
	"sort"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func addReferencedModelNames(seen map[string]struct{}, out *[]string, values ...string) {
	for _, value := range splitModelNames(values...) {
		name := strings.ToLower(strings.TrimSpace(value))
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		*out = append(*out, name)
	}
}

func ReferencedModelNames(ctx context.Context) ([]string, error) {
	seen := make(map[string]struct{}, 128)
	names := make([]string, 0, 128)

	channels, err := ChannelList(ctx)
	if err != nil {
		return nil, err
	}
	for _, channel := range channels {
		addReferencedModelNames(seen, &names, channel.Model, channel.CustomModel)
		for _, key := range channel.Keys {
			addReferencedModelNames(seen, &names, key.AllowedModels)
		}
	}

	groups, err := GroupList(ctx)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		for _, item := range group.Items {
			addReferencedModelNames(seen, &names, item.ModelName)
		}
	}

	overrides, err := RouteTargetOverrideList(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range overrides {
		addReferencedModelNames(seen, &names, row.ModelName)
	}

	apiKeys, err := APIKeyList(ctx)
	if err != nil {
		return nil, err
	}
	for _, apiKey := range apiKeys {
		addReferencedModelNames(seen, &names, apiKey.SupportedModels)
	}

	addReferencedModelNames(seen, &names, settingStringOrDefault(model.SettingKeyAIAutomationModel, ""))

	sort.Strings(names)
	return names, nil
}
