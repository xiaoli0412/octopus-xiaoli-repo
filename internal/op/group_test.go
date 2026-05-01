package op

import (
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestGroupItemBatchAddDeduplicatesAndAppendsPriority(t *testing.T) {
	ctx := setupOpTestDB(t)
	existingChannel := createConfiguredTestChannel(t, ctx, "group-batch-existing-channel", "existing", "")
	gptChannel := createConfiguredTestChannel(t, ctx, "group-batch-gpt-channel", "gpt-4o", "")
	claudeChannel := createConfiguredTestChannel(t, ctx, "group-batch-claude-channel", "claude-3-5-sonnet", "")

	group := &model.Group{Name: "group-batch", Mode: model.GroupModeFailover}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	if err := GroupItemAdd(&model.GroupItem{GroupID: group.ID, ChannelID: existingChannel.ID, ModelName: "existing", Priority: 3, Weight: 2}, ctx); err != nil {
		t.Fatalf("GroupItemAdd() error = %v", err)
	}

	if err := GroupItemBatchAdd(group.ID, []model.GroupIDAndLLMName{
		{ChannelID: gptChannel.ID, ModelName: "gpt-4o"},
		{ChannelID: gptChannel.ID, ModelName: "gpt-4o"},
		{ChannelID: claudeChannel.ID, ModelName: "claude-3-5-sonnet"},
		{ChannelID: 0, ModelName: "ignored"},
		{ChannelID: gptChannel.ID, ModelName: ""},
	}, ctx); err != nil {
		t.Fatalf("GroupItemBatchAdd() error = %v", err)
	}

	updated, err := GroupGet(group.ID, ctx)
	if err != nil {
		t.Fatalf("GroupGet() error = %v", err)
	}
	if len(updated.Items) != 3 {
		t.Fatalf("group items len = %d, want 3", len(updated.Items))
	}

	items := make(map[string]model.GroupItem, len(updated.Items))
	for _, item := range updated.Items {
		items[item.ModelName] = item
	}

	if items["existing"].Priority != 3 {
		t.Fatalf("existing item priority = %d, want 3", items["existing"].Priority)
	}
	if items["gpt-4o"].Priority != 4 || items["gpt-4o"].Weight != 1 {
		t.Fatalf("gpt-4o item = %#v, want priority 4 weight 1", items["gpt-4o"])
	}
	if items["claude-3-5-sonnet"].Priority != 5 || items["claude-3-5-sonnet"].Weight != 1 {
		t.Fatalf("claude item = %#v, want priority 5 weight 1", items["claude-3-5-sonnet"])
	}

	var count int64
	if err := db.GetDB().WithContext(ctx).Model(&model.GroupItem{}).Where("group_id = ?", group.ID).Count(&count).Error; err != nil {
		t.Fatalf("count group items error = %v", err)
	}
	if count != 3 {
		t.Fatalf("stored group item count = %d, want 3", count)
	}
}

func TestGroupUpdateRefreshesNameMapAndItems(t *testing.T) {
	ctx := setupOpTestDB(t)
	firstChannel := createConfiguredTestChannel(t, ctx, "group-update-first-channel", "gpt-4o", "")
	secondChannel := createConfiguredTestChannel(t, ctx, "group-update-second-channel", "claude-3-5-sonnet", "")
	thirdChannel := createConfiguredTestChannel(t, ctx, "group-update-third-channel", "gemini-2.5-pro", "")

	group := &model.Group{Name: "group-before", Mode: model.GroupModeRoundRobin}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	if err := GroupItemAdd(&model.GroupItem{GroupID: group.ID, ChannelID: firstChannel.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1}, ctx); err != nil {
		t.Fatalf("GroupItemAdd(first) error = %v", err)
	}
	if err := GroupItemAdd(&model.GroupItem{GroupID: group.ID, ChannelID: secondChannel.ID, ModelName: "claude-3-5-sonnet", Priority: 2, Weight: 2}, ctx); err != nil {
		t.Fatalf("GroupItemAdd(second) error = %v", err)
	}

	before, err := GroupGet(group.ID, ctx)
	if err != nil {
		t.Fatalf("GroupGet() before update error = %v", err)
	}
	if len(before.Items) != 2 {
		t.Fatalf("before update items len = %d, want 2", len(before.Items))
	}

	var firstItem, secondItem model.GroupItem
	for _, item := range before.Items {
		switch item.ModelName {
		case "gpt-4o":
			firstItem = item
		case "claude-3-5-sonnet":
			secondItem = item
		}
	}

	updatedName := "group-after"
	updatedMode := model.GroupModeFailover
	updatedRegex := "^gpt"
	firstTokenTimeout := 25
	sessionKeepTime := 60
	retryRounds := 3
	retryDelayMs := 150
	failoverWindowSec := 360
	raceAfterFails := 2
	raceConcurrency := 4

	updated, err := GroupUpdate(&model.GroupUpdateRequest{
		ID:                group.ID,
		Name:              &updatedName,
		Mode:              &updatedMode,
		MatchRegex:        &updatedRegex,
		FirstTokenTimeOut: &firstTokenTimeout,
		SessionKeepTime:   &sessionKeepTime,
		RetryRounds:       &retryRounds,
		RetryDelayMs:      &retryDelayMs,
		FailoverWindowSec: &failoverWindowSec,
		RaceAfterFails:    &raceAfterFails,
		RaceConcurrency:   &raceConcurrency,
		ItemsToUpdate: []model.GroupItemUpdateRequest{{
			ID:       firstItem.ID,
			Priority: 9,
			Weight:   7,
		}},
		ItemsToDelete: []int{secondItem.ID},
		ItemsToAdd: []model.GroupItemAddRequest{{
			ChannelID: thirdChannel.ID,
			ModelName: "gemini-2.5-pro",
			Priority:  5,
			Weight:    6,
		}},
	}, ctx)
	if err != nil {
		t.Fatalf("GroupUpdate() error = %v", err)
	}

	if updated.Name != updatedName || updated.Mode != updatedMode {
		t.Fatalf("updated group = %#v, want updated name and mode", updated)
	}
	if _, err := GroupGetMap("group-before", ctx); err == nil {
		t.Fatalf("old group name should be removed from groupMap")
	}
	if mapped, err := GroupGetMap(updatedName, ctx); err != nil || mapped.ID != group.ID {
		t.Fatalf("GroupGetMap(new name) = %#v, %v, want id %d", mapped, err, group.ID)
	}

	var stored model.Group
	if err := db.GetDB().WithContext(ctx).Preload("Items").First(&stored, group.ID).Error; err != nil {
		t.Fatalf("query updated group error = %v", err)
	}
	if stored.MatchRegex != updatedRegex || stored.RetryRounds != retryRounds || stored.RaceConcurrency != raceConcurrency {
		t.Fatalf("stored group fields = %#v, want updated retry/failover fields", stored)
	}
	if len(stored.Items) != 2 {
		t.Fatalf("stored items len = %d, want 2", len(stored.Items))
	}

	items := make(map[string]model.GroupItem, len(stored.Items))
	for _, item := range stored.Items {
		items[item.ModelName] = item
	}

	if updatedItem := items["gpt-4o"]; updatedItem.Priority != 9 || updatedItem.Weight != 7 {
		t.Fatalf("updated gpt-4o item = %#v, want priority 9 weight 7", updatedItem)
	}
	if _, ok := items["claude-3-5-sonnet"]; ok {
		t.Fatalf("deleted item still exists in stored items")
	}
	if addedItem := items["gemini-2.5-pro"]; addedItem.Priority != 5 || addedItem.Weight != 6 {
		t.Fatalf("added gemini item = %#v, want priority 5 weight 6", addedItem)
	}
}

func TestGroupCreateValidatesRuntimeConfig(t *testing.T) {
	ctx := setupOpTestDB(t)

	invalid := &model.Group{
		Name:            "group-invalid-runtime",
		Mode:            model.GroupModeFailover,
		RetryRounds:     -1,
		RaceAfterFails:  0,
		RaceConcurrency: 2,
	}
	if err := GroupCreate(invalid, ctx); err == nil {
		t.Fatalf("GroupCreate() expected runtime config validation error")
	}

	valid := &model.Group{
		Name:              "group-valid-runtime",
		Mode:              model.GroupModeFailover,
		RetryRounds:       1,
		RetryDelayMs:      0,
		FailoverWindowSec: 360,
		RaceAfterFails:    2,
		RaceConcurrency:   2,
	}
	if err := GroupCreate(valid, ctx); err != nil {
		t.Fatalf("GroupCreate() valid config error = %v", err)
	}
}

func TestGroupUpdateValidatesRuntimeConfig(t *testing.T) {
	ctx := setupOpTestDB(t)

	group := &model.Group{Name: "group-update-validate-runtime", Mode: model.GroupModeFailover}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}

	negativeRetry := -1
	if _, err := GroupUpdate(&model.GroupUpdateRequest{ID: group.ID, RetryRounds: &negativeRetry}, ctx); err == nil {
		t.Fatalf("GroupUpdate() expected negative retry_rounds validation error")
	}

	raceConcurrency := 2
	raceAfterFails := 0
	if _, err := GroupUpdate(&model.GroupUpdateRequest{ID: group.ID, RaceConcurrency: &raceConcurrency, RaceAfterFails: &raceAfterFails}, ctx); err == nil {
		t.Fatalf("GroupUpdate() expected race_after_fails validation error when race_concurrency is enabled")
	}

	validRaceAfterFails := 2
	updated, err := GroupUpdate(&model.GroupUpdateRequest{ID: group.ID, RaceConcurrency: &raceConcurrency, RaceAfterFails: &validRaceAfterFails}, ctx)
	if err != nil {
		t.Fatalf("GroupUpdate() valid runtime config error = %v", err)
	}
	if updated.RaceConcurrency != 2 || updated.RaceAfterFails != 2 {
		t.Fatalf("updated group race settings = %#v, want race_concurrency=2 race_after_fails=2", updated)
	}
}

func TestGroupRefreshCacheRemovesDeletedGroupFromCache(t *testing.T) {
	ctx := setupOpTestDB(t)

	group := &model.Group{Name: "group-refresh-delete", Mode: model.GroupModeFailover}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Delete(&model.Group{}, group.ID).Error; err != nil {
		t.Fatalf("delete group error = %v", err)
	}

	if err := groupRefreshCache(ctx); err != nil {
		t.Fatalf("groupRefreshCache() error = %v", err)
	}
	if _, err := GroupGet(group.ID, ctx); err == nil {
		t.Fatalf("GroupGet() expected deleted group to be absent after full refresh")
	}
	if _, err := GroupGetMap(group.Name, ctx); err == nil {
		t.Fatalf("GroupGetMap() expected deleted group name to be absent after full refresh")
	}
}
