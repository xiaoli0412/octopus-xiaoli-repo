package op

import (
	"context"
	"slices"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func createTestLLM(t *testing.T, ctx context.Context, name string) model.LLMInfo {
	t.Helper()
	info := model.LLMInfo{
		Name:        name,
		BillingMode: model.BillingModePerToken,
		ProbePolicy: model.ProbePolicyConcurrent,
		CachePolicy: model.CachePolicyUnsupported,
	}
	if err := LLMCreate(info, ctx); err != nil {
		t.Fatalf("LLMCreate(%q) error = %v", name, err)
	}
	return info
}

func TestDisableEnableModel(t *testing.T) {
	ctx := SetupOpTestDB(t)
	createTestLLM(t, ctx, "disable-test-model")

	if disabled, err := IsModelDisabled(ctx, "disable-test-model"); err != nil || disabled {
		t.Fatalf("IsModelDisabled() before disable = %v, err = %v, want false, nil", disabled, err)
	}

	if err := DisableModel(ctx, "disable-test-model"); err != nil {
		t.Fatalf("DisableModel() error = %v", err)
	}

	if disabled, err := IsModelDisabled(ctx, "disable-test-model"); err != nil || !disabled {
		t.Fatalf("IsModelDisabled() after disable = %v, err = %v, want true, nil", disabled, err)
	}

	disabledModels, err := ListDisabledModels(ctx)
	if err != nil {
		t.Fatalf("ListDisabledModels() error = %v", err)
	}
	if len(disabledModels) != 1 || disabledModels[0] != "disable-test-model" {
		t.Fatalf("ListDisabledModels() = %#v, want [disable-test-model]", disabledModels)
	}

	if err := EnableModel(ctx, "disable-test-model"); err != nil {
		t.Fatalf("EnableModel() error = %v", err)
	}

	if disabled, err := IsModelDisabled(ctx, "disable-test-model"); err != nil || disabled {
		t.Fatalf("IsModelDisabled() after enable = %v, err = %v, want false, nil", disabled, err)
	}
}

func TestDisabledModelNotFoundForUnknown(t *testing.T) {
	ctx := SetupOpTestDB(t)
	createTestLLM(t, ctx, "known-model")

	if disabled, err := IsModelDisabled(ctx, "unknown-model"); err != nil || disabled {
		t.Fatalf("IsModelDisabled(unknown) = %v, err = %v, want false, nil", disabled, err)
	}

	if err := DisableModel(ctx, "unknown-model"); err == nil {
		t.Fatalf("DisableModel(unknown) error = nil, want model not found")
	}

	if err := EnableModel(ctx, "unknown-model"); err == nil {
		t.Fatalf("EnableModel(unknown) error = nil, want model not found")
	}
}

func TestDisabledModelExcludedFromCapabilityInventory(t *testing.T) {
	ctx := SetupOpTestDB(t)
	createTestLLM(t, ctx, "gpt-4o-disabled")
	createConfiguredTestChannel(t, ctx, "cap-channel", "gpt-4o-disabled", "")

	before, err := CapabilityInventory(ctx)
	if err != nil {
		t.Fatalf("CapabilityInventory() before disable error = %v", err)
	}
	if !slices.ContainsFunc(before.SelectableModels, func(item model.SelectableGroupModelInventoryItem) bool {
		return item.Name == "gpt-4o-disabled"
	}) {
		t.Fatalf("CapabilityInventory() before disable missing selectable model gpt-4o-disabled")
	}

	if err := DisableModel(ctx, "gpt-4o-disabled"); err != nil {
		t.Fatalf("DisableModel() error = %v", err)
	}

	after, err := CapabilityInventory(ctx)
	if err != nil {
		t.Fatalf("CapabilityInventory() after disable error = %v", err)
	}
	if slices.ContainsFunc(after.SelectableModels, func(item model.SelectableGroupModelInventoryItem) bool {
		return item.Name == "gpt-4o-disabled"
	}) {
		t.Fatalf("CapabilityInventory() after disable still contains selectable model gpt-4o-disabled")
	}
	if slices.ContainsFunc(after.ServiceableModels, func(item model.ServiceableModelInventoryItem) bool {
		return item.Name == "gpt-4o-disabled"
	}) {
		t.Fatalf("CapabilityInventory() after disable still contains serviceable model gpt-4o-disabled")
	}
}

func TestDisabledModelExcludedFromGroupListModel(t *testing.T) {
	ctx := SetupOpTestDB(t)
	createTestLLM(t, ctx, "group-disabled-model")
	channel := createConfiguredTestChannel(t, ctx, "group-list-channel", "group-disabled-model", "")
	group := &model.Group{
		Name: "group-disabled-model",
		Mode: model.GroupModeRoundRobin,
		Items: []model.GroupItem{
			{ChannelID: channel.ID, ModelName: "group-disabled-model", Priority: 1, Weight: 1},
		},
	}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}

	before, err := GroupListModel(ctx)
	if err != nil {
		t.Fatalf("GroupListModel() before disable error = %v", err)
	}
	if !slices.Contains(before, "group-disabled-model") {
		t.Fatalf("GroupListModel() before disable missing group-disabled-model")
	}

	if err := DisableModel(ctx, "group-disabled-model"); err != nil {
		t.Fatalf("DisableModel() error = %v", err)
	}

	after, err := GroupListModel(ctx)
	if err != nil {
		t.Fatalf("GroupListModel() after disable error = %v", err)
	}
	if slices.Contains(after, "group-disabled-model") {
		t.Fatalf("GroupListModel() after disable still contains group-disabled-model")
	}
}

func TestLLMUpdateKeepsDisabledState(t *testing.T) {
	ctx := SetupOpTestDB(t)
	createTestLLM(t, ctx, "update-disabled-model")
	if err := DisableModel(ctx, "update-disabled-model"); err != nil {
		t.Fatalf("DisableModel() error = %v", err)
	}

	info, err := LLMGet("update-disabled-model")
	if err != nil {
		t.Fatalf("LLMGet() error = %v", err)
	}
	info.Input = 0.01
	info.Output = 0.02
	if err := LLMUpdate(info, ctx); err != nil {
		t.Fatalf("LLMUpdate() error = %v", err)
	}

	updated, err := LLMGet("update-disabled-model")
	if err != nil {
		t.Fatalf("LLMGet() after update error = %v", err)
	}
	if !updated.Disabled {
		t.Fatalf("LLMGet() after update Disabled = %v, want true", updated.Disabled)
	}
}
