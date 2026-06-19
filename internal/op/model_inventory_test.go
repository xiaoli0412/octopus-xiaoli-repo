package op

import (
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestReferencedModelNamesCollectsAllConfiguredSources(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channelA := createConfiguredTestChannel(t, ctx, "inventory-a", "gpt-4o,deepseek-v3.1", "glm-4.7")
	keyA := model.ChannelKey{
		ChannelID:     channelA.ID,
		Enabled:       true,
		ChannelKey:    "inventory-a-key-allowed",
		AllowedModels: "claude-3.7-sonnet,gpt-4o",
	}
	if err := db.GetDB().WithContext(ctx).Create(&keyA).Error; err != nil {
		t.Fatalf("create keyA error = %v", err)
	}
	if err := channelRefreshCacheByID(channelA.ID, ctx); err != nil {
		t.Fatalf("channelRefreshCacheByID(channelA) error = %v", err)
	}

	group := &model.Group{
		Name: "inventory-group",
		Mode: model.GroupModeRoundRobin,
		Items: []model.GroupItem{
			{ChannelID: channelA.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1},
		},
	}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}

	channelB := createConfiguredTestChannel(t, ctx, "inventory-b", "qwen-2.5-max", "")
	keyB := model.ChannelKey{
		ChannelID:     channelB.ID,
		Enabled:       true,
		ChannelKey:    "inventory-b-key-allowed",
		AllowedModels: "qwen-2.5-max",
	}
	if err := db.GetDB().WithContext(ctx).Create(&keyB).Error; err != nil {
		t.Fatalf("create keyB error = %v", err)
	}
	if err := channelRefreshCacheByID(channelB.ID, ctx); err != nil {
		t.Fatalf("channelRefreshCacheByID(channelB) error = %v", err)
	}
	if _, err := RouteTargetOverrideUpsert(model.RouteTargetOverride{
		ChannelID:             channelB.ID,
		ChannelKeyID:          keyB.ID,
		ModelName:             "qwen-2.5-max",
		BillingMode:           model.BillingModePerToken,
		ProbePolicy:           model.ProbePolicyConcurrent,
		ProbeIntervalSeconds:  30,
		ProbeConcurrencyLimit: 2,
	}, ctx); err != nil {
		t.Fatalf("RouteTargetOverrideUpsert() error = %v", err)
	}

	if err := APIKeyCreate(&model.APIKey{
		Name:            "inventory-client",
		APIKey:          "sk-inventory-client",
		Enabled:         true,
		SupportedModels: "moonshot-v1-8k",
	}, ctx); err != nil {
		t.Fatalf("APIKeyCreate() error = %v", err)
	}

	names, err := ReferencedModelNames(ctx)
	if err != nil {
		t.Fatalf("ReferencedModelNames() error = %v", err)
	}
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		seen[name] = struct{}{}
	}

	for _, want := range []string{
		"gpt-4o",
		"deepseek-v3.1",
		"glm-4.7",
		"claude-3.7-sonnet",
		"qwen-2.5-max",
		"moonshot-v1-8k",
	} {
		if _, ok := seen[want]; !ok {
			t.Fatalf("ReferencedModelNames() missing %q in %#v", want, names)
		}
	}
}

