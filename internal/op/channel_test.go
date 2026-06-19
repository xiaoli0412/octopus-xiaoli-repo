package op

import (
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
)

func TestChannelUpdateRefreshesCacheAndNormalizesKeys(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channel := &model.Channel{
		Name:              "channel-before",
		Enabled:           true,
		KeyManagementMode: model.KeyManagementModePooled,
		KeyRoutingPolicy:  model.KeyRoutingPolicyRoundRobin,
		BaseUrls:          []model.BaseUrl{{URL: "https://old.example.com", Delay: 20}},
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	existingKey := model.ChannelKey{
		ChannelID:     channel.ID,
		Enabled:       true,
		ChannelKey:    "old-key",
		SourceType:    "unknown",
		Remark:        "before",
		AllowedModels: "gpt-4o",
	}
	if err := db.GetDB().WithContext(ctx).Create(&existingKey).Error; err != nil {
		t.Fatalf("create existing key error = %v", err)
	}
	if err := channelRefreshCacheByID(channel.ID, ctx); err != nil {
		t.Fatalf("channelRefreshCacheByID() error = %v", err)
	}

	updatedName := "channel-after"
	updatedMode := model.KeyManagementModeClassified
	updatedPolicy := model.KeyRoutingPolicyFillPriority
	updatedBaseURLs := []model.BaseUrl{{URL: "https://fast.example.com", Delay: 3}}
	updatedSourceType := " Paid/Metered "
	updatedRemark := "after"
	updatedAllowedModels := " beta , alpha, beta "

	updated, err := ChannelUpdate(&model.ChannelUpdateRequest{
		ID:                channel.ID,
		Name:              &updatedName,
		KeyManagementMode: &updatedMode,
		KeyRoutingPolicy:  &updatedPolicy,
		BaseUrls:          &updatedBaseURLs,
		KeysToUpdate: []model.ChannelKeyUpdateRequest{{
			ID:            existingKey.ID,
			SourceType:    &updatedSourceType,
			Remark:        &updatedRemark,
			AllowedModels: &updatedAllowedModels,
		}},
		KeysToAdd: []model.ChannelKeyAddRequest{{
			Enabled:       true,
			ChannelKey:    "new-key",
			SourceType:    " Public/Free ",
			Remark:        "new",
			AllowedModels: " zeta, alpha, zeta ",
		}},
	}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate() error = %v", err)
	}

	if updated.Name != updatedName {
		t.Fatalf("updated channel name = %q, want %q", updated.Name, updatedName)
	}
	if updated.KeyManagementMode != updatedMode {
		t.Fatalf("updated key management mode = %q, want %q", updated.KeyManagementMode, updatedMode)
	}
	if updated.KeyRoutingPolicy != updatedPolicy {
		t.Fatalf("updated key routing policy = %q, want %q", updated.KeyRoutingPolicy, updatedPolicy)
	}
	if len(updated.BaseUrls) != 1 || updated.BaseUrls[0].URL != "https://fast.example.com" {
		t.Fatalf("updated base urls = %#v, want fast.example.com", updated.BaseUrls)
	}

	var stored model.Channel
	if err := db.GetDB().WithContext(ctx).Preload("Keys").First(&stored, channel.ID).Error; err != nil {
		t.Fatalf("query updated channel error = %v", err)
	}
	if stored.Name != updatedName {
		t.Fatalf("stored channel name = %q, want %q", stored.Name, updatedName)
	}

	keysByValue := make(map[string]model.ChannelKey, len(stored.Keys))
	for _, key := range stored.Keys {
		keysByValue[key.ChannelKey] = key
	}

	if len(keysByValue) != 2 {
		t.Fatalf("stored keys len = %d, want 2", len(keysByValue))
	}
	if got := keysByValue["old-key"].SourceType; got != "paid/metered" {
		t.Fatalf("updated key source_type = %q, want %q", got, "paid/metered")
	}
	if got := keysByValue["old-key"].AllowedModels; got != "alpha,beta" {
		t.Fatalf("updated key allowed_models = %q, want %q", got, "alpha,beta")
	}
	if got := keysByValue["new-key"].SourceType; got != "public/free" {
		t.Fatalf("new key source_type = %q, want %q", got, "public/free")
	}
	if got := keysByValue["new-key"].AllowedModels; got != "alpha,zeta" {
		t.Fatalf("new key allowed_models = %q, want %q", got, "alpha,zeta")
	}

	cached, err := ChannelGet(channel.ID, ctx)
	if err != nil {
		t.Fatalf("ChannelGet() error = %v", err)
	}
	if cached.Name != updatedName || len(cached.Keys) != 2 {
		t.Fatalf("cached channel = %#v, want updated name and 2 keys", cached)
	}
	if _, ok := channelKeyCache.Get(existingKey.ID); !ok {
		t.Fatalf("channelKeyCache missing updated key id %d", existingKey.ID)
	}
}

func TestChannelBaseUrlUpdateCopiesInputAndChannelKeySaveDBPersistsCacheUpdates(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channel := &model.Channel{Name: "channel-copy-test", Enabled: true}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	key := model.ChannelKey{
		ChannelID:     channel.ID,
		Enabled:       true,
		ChannelKey:    "save-me",
		Remark:        "before",
		AllowedModels: "gpt-4o",
	}
	if err := db.GetDB().WithContext(ctx).Create(&key).Error; err != nil {
		t.Fatalf("create channel key error = %v", err)
	}
	if err := channelRefreshCacheByID(channel.ID, ctx); err != nil {
		t.Fatalf("channelRefreshCacheByID() error = %v", err)
	}

	baseURLs := []model.BaseUrl{{URL: "https://origin.example.com", Delay: 11}}
	if err := ChannelBaseUrlUpdate(channel.ID, baseURLs, ctx); err != nil {
		t.Fatalf("ChannelBaseUrlUpdate() error = %v", err)
	}
	baseURLs[0].URL = "https://mutated.example.com"

	cached, err := ChannelGet(channel.ID, ctx)
	if err != nil {
		t.Fatalf("ChannelGet() error = %v", err)
	}
	if len(cached.BaseUrls) != 1 || cached.BaseUrls[0].URL != "https://origin.example.com" {
		t.Fatalf("cached base urls = %#v, want preserved copy", cached.BaseUrls)
	}

	var storedChannel model.Channel
	if err := db.GetDB().WithContext(ctx).First(&storedChannel, channel.ID).Error; err != nil {
		t.Fatalf("query stored channel error = %v", err)
	}
	if len(storedChannel.BaseUrls) != 1 || storedChannel.BaseUrls[0].URL != "https://origin.example.com" || storedChannel.BaseUrls[0].Delay != 11 {
		t.Fatalf("stored base urls = %#v, want persisted copy", storedChannel.BaseUrls)
	}

	updatedKey := cached.Keys[0]
	updatedKey.Remark = "after"
	updatedKey.AllowedModels = "gpt-4o,claude-3-5-sonnet"
	if err := ChannelKeyUpdate(updatedKey); err != nil {
		t.Fatalf("ChannelKeyUpdate() error = %v", err)
	}
	if err := ChannelKeySaveDB(ctx); err != nil {
		t.Fatalf("ChannelKeySaveDB() error = %v", err)
	}

	var storedKey model.ChannelKey
	if err := db.GetDB().WithContext(ctx).First(&storedKey, updatedKey.ID).Error; err != nil {
		t.Fatalf("query stored key error = %v", err)
	}
	if storedKey.Remark != "after" {
		t.Fatalf("stored key remark = %q, want %q", storedKey.Remark, "after")
	}
	if storedKey.AllowedModels != "claude-3-5-sonnet,gpt-4o" {
		t.Fatalf("stored key allowed_models = %q, want normalized updated value", storedKey.AllowedModels)
	}

	refreshed, err := ChannelGet(channel.ID, ctx)
	if err != nil {
		t.Fatalf("ChannelGet() after key update error = %v", err)
	}
	if refreshed.Keys[0].Remark != "after" {
		t.Fatalf("cached key remark = %q, want %q", refreshed.Keys[0].Remark, "after")
	}

	if err := ChannelKeySaveDB(ctx); err != nil {
		t.Fatalf("second ChannelKeySaveDB() error = %v", err)
	}
}

func TestChannelBaseUrlUpdateRejectsInvalidURL(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channel := &model.Channel{Name: "channel-invalid-base-url-update", Enabled: true}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	if err := ChannelBaseUrlUpdate(channel.ID, []model.BaseUrl{{URL: "ftp://example.com/v1", Delay: 0}}, ctx); err == nil {
		t.Fatal("ChannelBaseUrlUpdate() expected invalid base url error")
	}

	cached, err := ChannelGet(channel.ID, ctx)
	if err != nil {
		t.Fatalf("ChannelGet() error = %v", err)
	}
	if len(cached.BaseUrls) != 0 {
		t.Fatalf("cached base urls = %#v, want empty after rejected update", cached.BaseUrls)
	}
}

func TestChannelRefreshCacheRemovesDeletedChannelFromCache(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channel := &model.Channel{Name: "channel-refresh-delete", Enabled: true}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Delete(&model.Channel{}, channel.ID).Error; err != nil {
		t.Fatalf("delete channel error = %v", err)
	}

	if err := channelRefreshCache(ctx); err != nil {
		t.Fatalf("channelRefreshCache() error = %v", err)
	}
	if _, err := ChannelGet(channel.ID, ctx); err == nil {
		t.Fatalf("ChannelGet() expected deleted channel to be absent after full refresh")
	}
}

func TestChannelDelRemovesDependentGroupItemsAndCaches(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channel := &model.Channel{Name: "channel-delete", Enabled: true}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	key := model.ChannelKey{ChannelID: channel.ID, Enabled: true, ChannelKey: "delete-key"}
	if err := db.GetDB().WithContext(ctx).Create(&key).Error; err != nil {
		t.Fatalf("create channel key error = %v", err)
	}
	if err := channelRefreshCacheByID(channel.ID, ctx); err != nil {
		t.Fatalf("channelRefreshCacheByID() error = %v", err)
	}

	group := &model.Group{Name: "group-delete", Mode: model.GroupModeFailover}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	if err := GroupItemAdd(&model.GroupItem{GroupID: group.ID, ChannelID: channel.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1}, ctx); err != nil {
		t.Fatalf("GroupItemAdd() error = %v", err)
	}

	stats := model.StatsChannel{ChannelID: channel.ID}
	if err := db.GetDB().WithContext(ctx).Create(&stats).Error; err != nil {
		t.Fatalf("create stats channel error = %v", err)
	}
	statsChannelCache.Set(channel.ID, stats)

	if err := ChannelDel(channel.ID, ctx); err != nil {
		t.Fatalf("ChannelDel() error = %v", err)
	}

	if _, err := ChannelGet(channel.ID, ctx); err == nil {
		t.Fatalf("ChannelGet() after delete expected error")
	}
	if _, ok := channelKeyCache.Get(key.ID); ok {
		t.Fatalf("channelKeyCache still contains deleted key id %d", key.ID)
	}
	if _, ok := statsChannelCache.Get(channel.ID); ok {
		t.Fatalf("statsChannelCache still contains deleted channel id %d", channel.ID)
	}

	refreshedGroup, err := GroupGet(group.ID, ctx)
	if err != nil {
		t.Fatalf("GroupGet() error = %v", err)
	}
	if len(refreshedGroup.Items) != 0 {
		t.Fatalf("group items after channel delete = %d, want 0", len(refreshedGroup.Items))
	}

	var channelCount, keyCount, groupItemCount, statsCount int64
	if err := db.GetDB().WithContext(ctx).Model(&model.Channel{}).Where("id = ?", channel.ID).Count(&channelCount).Error; err != nil {
		t.Fatalf("count channels error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Model(&model.ChannelKey{}).Where("channel_id = ?", channel.ID).Count(&keyCount).Error; err != nil {
		t.Fatalf("count channel keys error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Model(&model.GroupItem{}).Where("channel_id = ?", channel.ID).Count(&groupItemCount).Error; err != nil {
		t.Fatalf("count group items error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Model(&model.StatsChannel{}).Where("channel_id = ?", channel.ID).Count(&statsCount).Error; err != nil {
		t.Fatalf("count stats channels error = %v", err)
	}
	if channelCount != 0 || keyCount != 0 || groupItemCount != 0 || statsCount != 0 {
		t.Fatalf("counts after delete = channel:%d key:%d groupItem:%d stats:%d, want all zero", channelCount, keyCount, groupItemCount, statsCount)
	}
}

func TestChannelCreateNormalizesAndValidatesKeyRoutingSettings(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channel := &model.Channel{
		Name:              "channel-create-normalize",
		Enabled:           true,
		KeyManagementMode: model.KeyManagementMode(" Classified "),
		KeyRoutingPolicy:  model.KeyRoutingPolicy(" Fill_Priority "),
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	if channel.KeyManagementMode != model.KeyManagementModeClassified {
		t.Fatalf("created channel key_management_mode = %q, want %q", channel.KeyManagementMode, model.KeyManagementModeClassified)
	}
	if channel.KeyRoutingPolicy != model.KeyRoutingPolicyFillPriority {
		t.Fatalf("created channel key_routing_policy = %q, want %q", channel.KeyRoutingPolicy, model.KeyRoutingPolicyFillPriority)
	}

	invalid := &model.Channel{
		Name:              "channel-create-invalid",
		Enabled:           true,
		KeyManagementMode: model.KeyManagementMode("invalid-mode"),
		KeyRoutingPolicy:  model.KeyRoutingPolicyRoundRobin,
	}
	if err := ChannelCreate(invalid, ctx); err == nil {
		t.Fatalf("ChannelCreate() expected invalid key management mode error")
	}
	invalid.KeyManagementMode = model.KeyManagementModePooled
	invalid.KeyRoutingPolicy = model.KeyRoutingPolicy("invalid-policy")
	if err := ChannelCreate(invalid, ctx); err == nil {
		t.Fatalf("ChannelCreate() expected invalid key routing policy error")
	}
}

func TestChannelUpdateValidatesAndNormalizesKeyRoutingSettings(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channel := &model.Channel{
		Name:              "channel-update-validate",
		Enabled:           true,
		KeyManagementMode: model.KeyManagementModePooled,
		KeyRoutingPolicy:  model.KeyRoutingPolicyRoundRobin,
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	mode := model.KeyManagementMode(" Classified ")
	policy := model.KeyRoutingPolicy(" Priority_Order ")
	updated, err := ChannelUpdate(&model.ChannelUpdateRequest{
		ID:                channel.ID,
		KeyManagementMode: &mode,
		KeyRoutingPolicy:  &policy,
	}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate() error = %v", err)
	}
	if updated.KeyManagementMode != model.KeyManagementModeClassified {
		t.Fatalf("updated key_management_mode = %q, want %q", updated.KeyManagementMode, model.KeyManagementModeClassified)
	}
	if updated.KeyRoutingPolicy != model.KeyRoutingPolicyPriority {
		t.Fatalf("updated key_routing_policy = %q, want %q", updated.KeyRoutingPolicy, model.KeyRoutingPolicyPriority)
	}

	invalidMode := model.KeyManagementMode("boom")
	if _, err := ChannelUpdate(&model.ChannelUpdateRequest{ID: channel.ID, KeyManagementMode: &invalidMode}, ctx); err == nil {
		t.Fatalf("ChannelUpdate() expected invalid key management mode error")
	}
	invalidPolicy := model.KeyRoutingPolicy("boom")
	if _, err := ChannelUpdate(&model.ChannelUpdateRequest{ID: channel.ID, KeyRoutingPolicy: &invalidPolicy}, ctx); err == nil {
		t.Fatalf("ChannelUpdate() expected invalid key routing policy error")
	}

	refreshed, err := ChannelGet(channel.ID, ctx)
	if err != nil {
		t.Fatalf("ChannelGet() error = %v", err)
	}
	if refreshed.KeyManagementMode != model.KeyManagementModeClassified {
		t.Fatalf("refreshed key_management_mode = %q, want preserved classified", refreshed.KeyManagementMode)
	}
	if refreshed.KeyRoutingPolicy != model.KeyRoutingPolicyPriority {
		t.Fatalf("refreshed key_routing_policy = %q, want preserved priority_order", refreshed.KeyRoutingPolicy)
	}
}

func TestChannelCreateAndUpdateValidateChannelKeySourceType(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channel := &model.Channel{
		Name:              "channel-source-type-validate",
		Enabled:           true,
		KeyManagementMode: model.KeyManagementModePooled,
		KeyRoutingPolicy:  model.KeyRoutingPolicyRoundRobin,
		Keys: []model.ChannelKey{{
			Enabled:       true,
			ChannelKey:    "bad-key",
			SourceType:    "enterprise",
			AllowedModels: "gpt-4o",
		}},
	}
	if err := ChannelCreate(channel, ctx); err == nil {
		t.Fatalf("ChannelCreate() expected invalid channel key source type error")
	}

	valid := &model.Channel{
		Name:              "channel-source-type-ok",
		Enabled:           true,
		KeyManagementMode: model.KeyManagementModePooled,
		KeyRoutingPolicy:  model.KeyRoutingPolicyRoundRobin,
	}
	if err := ChannelCreate(valid, ctx); err != nil {
		t.Fatalf("ChannelCreate(valid) error = %v", err)
	}

	_, err := ChannelUpdate(&model.ChannelUpdateRequest{
		ID: valid.ID,
		KeysToAdd: []model.ChannelKeyAddRequest{{
			Enabled:       true,
			ChannelKey:    "bad-add",
			SourceType:    "enterprise",
			AllowedModels: "gpt-4o",
		}},
	}, ctx)
	if err == nil {
		t.Fatalf("ChannelUpdate(KeysToAdd) expected invalid channel key source type error")
	}

	updated, err := ChannelUpdate(&model.ChannelUpdateRequest{
		ID: valid.ID,
		KeysToAdd: []model.ChannelKeyAddRequest{{
			Enabled:       true,
			ChannelKey:    "good-add",
			SourceType:    "free",
			AllowedModels: "gpt-4o",
		}},
	}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(valid add) error = %v", err)
	}
	if len(updated.Keys) != 1 || updated.Keys[0].SourceType != "public/free" {
		t.Fatalf("updated keys = %#v, want normalized public/free source type", updated.Keys)
	}
}

func TestChannelCreateAndUpdateRejectUnsupportedBaseURLScheme(t *testing.T) {
	ctx := SetupOpTestDB(t)

	invalid := &model.Channel{
		Name:     "channel-invalid-base-url",
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: "ftp://example.com/v1", Delay: 0}},
	}
	if err := ChannelCreate(invalid, ctx); err == nil {
		t.Fatalf("ChannelCreate() expected invalid base url error")
	}

	channel := &model.Channel{Name: "channel-valid-base-url", Enabled: true, BaseUrls: []model.BaseUrl{{URL: "https://example.com/v1", Delay: 0}}}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate(valid) error = %v", err)
	}

	updatedBaseURLs := []model.BaseUrl{{URL: "ftp://example.com/v1", Delay: 1}}
	if _, err := ChannelUpdate(&model.ChannelUpdateRequest{ID: channel.ID, BaseUrls: &updatedBaseURLs}, ctx); err == nil {
		t.Fatalf("ChannelUpdate() expected invalid base url error")
	}

	refreshed, err := ChannelGet(channel.ID, ctx)
	if err != nil {
		t.Fatalf("ChannelGet() error = %v", err)
	}
	if len(refreshed.BaseUrls) != 1 || refreshed.BaseUrls[0].URL != "https://example.com/v1" {
		t.Fatalf("refreshed base urls = %#v, want preserved valid base url", refreshed.BaseUrls)
	}
}

func TestChannelCreateAndUpdateRejectChannelProxyWithCredentials(t *testing.T) {
	ctx := SetupOpTestDB(t)

	proxy := "https://user:pass@example.com:8443"
	invalid := &model.Channel{
		Name:         "channel-invalid-proxy-create",
		Enabled:      true,
		ChannelProxy: &proxy,
	}
	if err := ChannelCreate(invalid, ctx); err == nil {
		t.Fatalf("ChannelCreate() expected invalid channel proxy error")
	}

	channel := &model.Channel{Name: "channel-valid-proxy-update", Enabled: true}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate(valid) error = %v", err)
	}

	if _, err := ChannelUpdate(&model.ChannelUpdateRequest{ID: channel.ID, ChannelProxy: &proxy}, ctx); err == nil {
		t.Fatalf("ChannelUpdate() expected invalid channel proxy error")
	}

	blank := "   "
	updated, err := ChannelUpdate(&model.ChannelUpdateRequest{ID: channel.ID, ChannelProxy: &blank}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(blank proxy) error = %v", err)
	}
	if updated.ChannelProxy != nil {
		t.Fatalf("updated channel proxy = %#v, want nil after blank normalization", updated.ChannelProxy)
	}
}

func TestChannelCreateAndUpdatePersistAllowedModelsInBothModes(t *testing.T) {
	ctx := SetupOpTestDB(t)

	classified := &model.Channel{
		Name:              "channel-allowed-models-classified",
		Enabled:           true,
		KeyManagementMode: model.KeyManagementModeClassified,
		KeyRoutingPolicy:  model.KeyRoutingPolicyRoundRobin,
		Keys: []model.ChannelKey{{
			Enabled:       true,
			ChannelKey:    "classified-key",
			SourceType:    model.ChannelKeySourceTypePublicFree,
			AllowedModels: "gpt-4o,claude-3-5-sonnet",
		}},
	}
	if err := ChannelCreate(classified, ctx); err != nil {
		t.Fatalf("ChannelCreate(classified) error = %v", err)
	}
	if len(classified.Keys) != 1 || classified.Keys[0].AllowedModels != "claude-3-5-sonnet,gpt-4o" {
		t.Fatalf("classified key allowed_models = %#v, want normalized sorted models", classified.Keys)
	}

	updatedAllowed := "gemini-2.5-pro, gpt-4o"
	updated, err := ChannelUpdate(&model.ChannelUpdateRequest{
		ID: classified.ID,
		KeysToUpdate: []model.ChannelKeyUpdateRequest{{
			ID:            classified.Keys[0].ID,
			AllowedModels: &updatedAllowed,
		}},
	}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(classified) error = %v", err)
	}
	if len(updated.Keys) != 1 || updated.Keys[0].AllowedModels != "gemini-2.5-pro,gpt-4o" {
		t.Fatalf("updated classified key allowed_models = %#v, want normalized update", updated.Keys)
	}

	pooled := &model.Channel{
		Name:              "channel-allowed-models-pooled",
		Enabled:           true,
		KeyManagementMode: model.KeyManagementModePooled,
		Model:             "gpt-4o,claude-3-5-sonnet",
		KeyRoutingPolicy:  model.KeyRoutingPolicyRoundRobin,
		Keys: []model.ChannelKey{{
			Enabled:       true,
			ChannelKey:    "pooled-key",
			SourceType:    model.ChannelKeySourceTypePublicFree,
			AllowedModels: "gpt-4o",
		}},
	}
	if err := ChannelCreate(pooled, ctx); err != nil {
		t.Fatalf("ChannelCreate(pooled) error = %v", err)
	}
	if !pooled.HasConfiguredKeyForModel("gpt-4o") {
		t.Fatalf("pooled channel should keep configured key for gpt-4o")
	}
	if pooled.HasConfiguredKeyForModel("gemini-2.5-pro") {
		t.Fatalf("pooled channel should not expose unsupported model as configured")
	}
}

func TestChannelUpdateRemovesGroupItemsWhenDeletingLastEligibleKeyForModel(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channel := &model.Channel{
		Name:        "channel-delete-last-key",
		Enabled:     true,
		Model:       "gpt-4o",
		CustomModel: "",
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	key := model.ChannelKey{ChannelID: channel.ID, Enabled: true, ChannelKey: "k1", AllowedModels: "gpt-4o"}
	if err := db.GetDB().WithContext(ctx).Create(&key).Error; err != nil {
		t.Fatalf("create key error = %v", err)
	}
	if err := channelRefreshCacheByID(channel.ID, ctx); err != nil {
		t.Fatalf("channelRefreshCacheByID() error = %v", err)
	}
	group := &model.Group{Name: "group-delete-last-key", Mode: model.GroupModeFailover}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	if err := GroupItemAdd(&model.GroupItem{GroupID: group.ID, ChannelID: channel.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1}, ctx); err != nil {
		t.Fatalf("GroupItemAdd() error = %v", err)
	}

	updated, err := ChannelUpdate(&model.ChannelUpdateRequest{ID: channel.ID, KeysToDelete: []int{key.ID}}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate() error = %v", err)
	}
	if updated.HasConfiguredKeyForModel("gpt-4o") {
		t.Fatalf("updated channel still reports configured key for gpt-4o")
	}
	refreshedGroup, err := GroupGet(group.ID, ctx)
	if err != nil {
		t.Fatalf("GroupGet() error = %v", err)
	}
	if len(refreshedGroup.Items) != 0 {
		t.Fatalf("group items len = %d, want 0", len(refreshedGroup.Items))
	}
	var overrideCount int64
	if err := db.GetDB().WithContext(ctx).Model(&model.RouteTargetOverride{}).Where("channel_id = ? AND model_name = ?", channel.ID, "gpt-4o").Count(&overrideCount).Error; err != nil {
		t.Fatalf("count route target overrides error = %v", err)
	}
	if overrideCount != 0 {
		t.Fatalf("route target override count = %d, want 0", overrideCount)
	}
}

func TestChannelUpdateKeepsGroupItemsWhenAnotherKeyStillServesModel(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channel := &model.Channel{Name: "channel-keep-group-item", Enabled: true, Model: "gpt-4o"}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	key1 := model.ChannelKey{ChannelID: channel.ID, Enabled: true, ChannelKey: "k1", AllowedModels: "gpt-4o"}
	key2 := model.ChannelKey{ChannelID: channel.ID, Enabled: true, ChannelKey: "k2", AllowedModels: "gpt-4o"}
	if err := db.GetDB().WithContext(ctx).Create(&key1).Error; err != nil {
		t.Fatalf("create key1 error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&key2).Error; err != nil {
		t.Fatalf("create key2 error = %v", err)
	}
	if err := channelRefreshCacheByID(channel.ID, ctx); err != nil {
		t.Fatalf("channelRefreshCacheByID() error = %v", err)
	}
	group := &model.Group{Name: "group-keep-group-item", Mode: model.GroupModeFailover}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	if err := GroupItemAdd(&model.GroupItem{GroupID: group.ID, ChannelID: channel.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1}, ctx); err != nil {
		t.Fatalf("GroupItemAdd() error = %v", err)
	}

	updated, err := ChannelUpdate(&model.ChannelUpdateRequest{ID: channel.ID, KeysToDelete: []int{key1.ID}}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate() error = %v", err)
	}
	if !updated.HasConfiguredKeyForModel("gpt-4o") {
		t.Fatalf("updated channel should still have configured key for gpt-4o")
	}
	refreshedGroup, err := GroupGet(group.ID, ctx)
	if err != nil {
		t.Fatalf("GroupGet() error = %v", err)
	}
	if len(refreshedGroup.Items) != 1 {
		t.Fatalf("group items len = %d, want 1", len(refreshedGroup.Items))
	}
}

func TestChannelLLMListUsesKeyLimitedCapabilityInventory(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channel := &model.Channel{
		Name:    "channel-key-inventory",
		Enabled: true,
		Type:    outbound.OutboundTypeOpenAIChat,
		Keys: []model.ChannelKey{
			{Enabled: true, ChannelKey: "key-gpt", AllowedModels: "gpt-4o", RequestCapabilities: "openai_chat"},
			{Enabled: true, ChannelKey: "key-claude", AllowedModels: "claude-3-5-sonnet", RequestCapabilities: "openai_chat"},
			{Enabled: true, ChannelKey: "key-gemini-wrong-protocol", AllowedModels: "gemini-2.5-pro", RequestCapabilities: "gemini_contents"},
		},
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	items, err := ChannelLLMList(ctx)
	if err != nil {
		t.Fatalf("ChannelLLMList() error = %v", err)
	}
	byName := make(map[string]model.LLMChannel, len(items))
	for _, item := range items {
		if item.ChannelID == channel.ID {
			byName[item.Name] = item
		}
	}

	if got := byName["gpt-4o"]; got.KeyCount != 1 || got.InventorySource != "channel_key_allowed" {
		t.Fatalf("gpt-4o inventory = %#v, want key_count=1 source=channel_key_allowed", got)
	}
	if got := byName["claude-3-5-sonnet"]; got.KeyCount != 1 {
		t.Fatalf("claude inventory = %#v, want key_count=1", got)
	}
	if _, ok := byName["gemini-2.5-pro"]; ok {
		t.Fatalf("gemini-2.5-pro should be hidden because request capability mismatches OpenAI chat channel")
	}
}

func TestCapabilityInventoryDoesNotPromoteStaleAPIKeyBindings(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channel := &model.Channel{
		Name:    "capability-inventory-serviceable",
		Enabled: true,
		Type:    outbound.OutboundTypeOpenAIChat,
		Keys: []model.ChannelKey{{
			Enabled:       true,
			ChannelKey:    "key-gpt",
			AllowedModels: "gpt-4o",
		}},
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	apiKey := &model.APIKey{
		Name:            "client-stale-supported-model",
		APIKey:          "sk-client-stale-supported-model",
		Enabled:         true,
		SupportedModels: "gpt-4o,stale-only-model",
	}
	if err := APIKeyCreate(apiKey, ctx); err != nil {
		t.Fatalf("APIKeyCreate() error = %v", err)
	}

	inventory, err := CapabilityInventory(ctx)
	if err != nil {
		t.Fatalf("CapabilityInventory() error = %v", err)
	}
	selectable := make(map[string]model.SelectableGroupModelInventoryItem, len(inventory.SelectableModels))
	for _, item := range inventory.SelectableModels {
		selectable[item.Name] = item
	}
	if _, ok := selectable["gpt-4o"]; !ok {
		t.Fatalf("selectable models = %#v, want serviceable gpt-4o", inventory.SelectableModels)
	}
	if stale, ok := selectable["stale-only-model"]; ok {
		t.Fatalf("stale API key binding leaked into selectable inventory: %#v", stale)
	}
}

func TestGroupItemAddAllowsUndeclaredKeyLimitedModel(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channel := &model.Channel{
		Name:    "channel-undeclared-key-model",
		Enabled: true,
		Type:    outbound.OutboundTypeOpenAIChat,
		Keys: []model.ChannelKey{{
			Enabled:       true,
			ChannelKey:    "key-gpt",
			AllowedModels: "gpt-4o",
		}},
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	group := &model.Group{Name: "group-undeclared-key-model", Mode: model.GroupModeRoundRobin}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	if err := GroupItemAdd(&model.GroupItem{GroupID: group.ID, ChannelID: channel.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1}, ctx); err != nil {
		t.Fatalf("GroupItemAdd() should allow key-limited undeclared model: %v", err)
	}
}

func TestGroupItemAddRejectsProtocolMismatchedKey(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channel := &model.Channel{
		Name:    "channel-protocol-mismatch",
		Enabled: true,
		Type:    outbound.OutboundTypeOpenAIChat,
		Keys: []model.ChannelKey{{
			Enabled:             true,
			ChannelKey:          "key-gemini-only",
			AllowedModels:       "gpt-4o",
			RequestCapabilities: "gemini_contents",
		}},
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	group := &model.Group{Name: "group-protocol-mismatch", Mode: model.GroupModeRoundRobin}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	if err := GroupItemAdd(&model.GroupItem{GroupID: group.ID, ChannelID: channel.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1}, ctx); err == nil {
		t.Fatalf("GroupItemAdd() expected protocol mismatch error")
	}
}
