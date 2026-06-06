package op

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/cache"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/diff"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/xstrings"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/xurl"
)

var channelCache = cache.New[int, model.Channel](16)
var channelKeyCache = cache.New[int, model.ChannelKey](16)
var channelKeyCacheNeedUpdate = make(map[int]struct{})
var channelKeyCacheNeedUpdateLock sync.Mutex

func normalizeAndValidateBaseURLs(baseURLs []model.BaseUrl) ([]model.BaseUrl, error) {
	if len(baseURLs) == 0 {
		return nil, nil
	}
	validated := make([]model.BaseUrl, len(baseURLs))
	for i, baseURL := range baseURLs {
		if err := xurl.ValidateAbsoluteHTTPURL(baseURL.URL, "base_url"); err != nil {
			return nil, err
		}
		validated[i] = model.BaseUrl{URL: strings.TrimSpace(baseURL.URL), Delay: baseURL.Delay}
	}
	return validated, nil
}

func ChannelList(ctx context.Context) ([]model.Channel, error) {
	channels := channelCache.Values()
	return channels, nil
}

func ChannelCreate(channel *model.Channel, ctx context.Context) error {
	channel.KeyManagementMode = model.NormalizeKeyManagementMode(channel.KeyManagementMode)
	if !model.IsValidKeyManagementMode(channel.KeyManagementMode) {
		return fmt.Errorf("invalid key management mode: %q", channel.KeyManagementMode)
	}
	channel.KeyRoutingPolicy = model.NormalizeKeyRoutingPolicy(channel.KeyRoutingPolicy)
	if !model.IsValidKeyRoutingPolicy(channel.KeyRoutingPolicy) {
		return fmt.Errorf("invalid key routing policy: %q", channel.KeyRoutingPolicy)
	}
	for i := range channel.Keys {
		normalizedSourceType := model.NormalizeChannelKeySourceType(channel.Keys[i].SourceType)
		if !model.IsValidChannelKeySourceType(normalizedSourceType) {
			return fmt.Errorf("invalid channel key source type: %q", channel.Keys[i].SourceType)
		}
		channel.Keys[i].SourceType = normalizedSourceType
		channel.Keys[i].AllowedModels = model.NormalizeChannelKeyAllowedModels(channel.Keys[i].AllowedModels)
		channel.Keys[i].RequestCapabilities = model.NormalizeChannelKeyRequestCapabilities(channel.Keys[i].RequestCapabilities)
	}
	validatedBaseURLs, err := normalizeAndValidateBaseURLs(channel.BaseUrls)
	if err != nil {
		return err
	}
	channel.BaseUrls = validatedBaseURLs
	channel.ChannelProxy = model.NormalizeChannelProxy(channel.ChannelProxy)
	if err := model.ValidateChannelProxy(channel.ChannelProxy); err != nil {
		return err
	}
	if err := db.GetDB().WithContext(ctx).Create(channel).Error; err != nil {
		return err
	}
	channelCache.Set(channel.ID, *channel)
	for _, k := range channel.Keys {
		if k.ID != 0 {
			channelKeyCache.Set(k.ID, k)
		}
	}
	return nil
}

func ChannelKeyUpdate(key model.ChannelKey) error {
	if key.ID == 0 || key.ChannelID == 0 {
		return fmt.Errorf("invalid channel key")
	}
	key.AllowedModels = model.NormalizeChannelKeyAllowedModels(key.AllowedModels)
	key.RequestCapabilities = model.NormalizeChannelKeyRequestCapabilities(key.RequestCapabilities)
	ch, ok := channelCache.Get(key.ChannelID)
	if !ok {
		return fmt.Errorf("channel not found")
	}
	if len(ch.Keys) > 0 {
		keys := make([]model.ChannelKey, len(ch.Keys))
		copy(keys, ch.Keys)
		for i := range keys {
			if keys[i].ID == key.ID {
				keys[i] = key
				break
			}
		}
		ch.Keys = keys
	}
	channelCache.Set(key.ChannelID, ch)
	channelKeyCache.Set(key.ID, key)
	channelKeyCacheNeedUpdateLock.Lock()
	channelKeyCacheNeedUpdate[key.ID] = struct{}{}
	channelKeyCacheNeedUpdateLock.Unlock()
	return nil
}

func ChannelBaseUrlUpdate(channelID int, baseUrl []model.BaseUrl, ctx context.Context) error {
	ch, ok := channelCache.Get(channelID)
	if !ok {
		return fmt.Errorf("channel not found")
	}
	validatedBaseURLs, err := normalizeAndValidateBaseURLs(baseUrl)
	if err != nil {
		return err
	}
	result := db.GetDB().WithContext(ctx).Model(&model.Channel{}).Where("id = ?", channelID).Select("base_urls").Updates(&model.Channel{BaseUrls: validatedBaseURLs})
	if result.Error != nil {
		return fmt.Errorf("failed to persist base urls: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("channel not found")
	}
	ch.BaseUrls = validatedBaseURLs
	channelCache.Set(channelID, ch)
	return nil
}

func ChannelKeySaveDB(ctx context.Context) error {
	channelKeyCacheNeedUpdateLock.Lock()
	keyIDs := make([]int, 0, len(channelKeyCacheNeedUpdate))
	for id := range channelKeyCacheNeedUpdate {
		keyIDs = append(keyIDs, id)
	}
	channelKeyCacheNeedUpdate = make(map[int]struct{})
	channelKeyCacheNeedUpdateLock.Unlock()

	if len(keyIDs) == 0 {
		return nil
	}

	dbConn := db.GetDB().WithContext(ctx)
	for _, id := range keyIDs {
		k, ok := channelKeyCache.Get(id)
		if !ok {
			continue
		}
		if err := dbConn.Save(&k).Error; err != nil {
			return err
		}
	}
	return nil
}

func ChannelUpdate(req *model.ChannelUpdateRequest, ctx context.Context) (*model.Channel, error) {
	oldChannel, ok := channelCache.Get(req.ID)
	if !ok {
		return nil, fmt.Errorf("channel not found")
	}
	oldReferencedModels := xstrings.SplitTrimCompact(",", oldChannel.Model, oldChannel.CustomModel)
	for _, key := range oldChannel.Keys {
		oldReferencedModels = append(oldReferencedModels, xstrings.SplitTrimCompact(",", key.AllowedModels)...)
	}
	var existingGroupItems []model.GroupItem
	if err := db.GetDB().WithContext(ctx).Where("channel_id = ?", req.ID).Find(&existingGroupItems).Error; err != nil {
		return nil, fmt.Errorf("failed to list existing group items for channel: %w", err)
	}
	for _, item := range existingGroupItems {
		oldReferencedModels = append(oldReferencedModels, item.ModelName)
	}
	oldConfiguredKeyModels := make(map[string]struct{})
	for _, modelName := range oldReferencedModels {
		if oldChannel.HasConfiguredKeyForRequest(modelName, oldChannel.RequestCapabilityForModel(modelName)) {
			oldConfiguredKeyModels[modelName] = struct{}{}
		}
	}
	oldModelSet := xstrings.SplitTrimCompact(",", oldChannel.Model, oldChannel.CustomModel)
	modelsChanged := req.Model != nil || req.CustomModel != nil

	tx := db.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var selectFields []string
	updates := model.Channel{ID: req.ID}

	if req.Name != nil {
		selectFields = append(selectFields, "name")
		updates.Name = *req.Name
	}
	if req.Type != nil {
		selectFields = append(selectFields, "type")
		updates.Type = *req.Type
	}
	if req.Enabled != nil {
		selectFields = append(selectFields, "enabled")
		updates.Enabled = *req.Enabled
	}
	if req.KeyManagementMode != nil {
		normalizedMode := model.NormalizeKeyManagementMode(*req.KeyManagementMode)
		if !model.IsValidKeyManagementMode(normalizedMode) {
			tx.Rollback()
			return nil, fmt.Errorf("invalid key management mode: %q", *req.KeyManagementMode)
		}
		selectFields = append(selectFields, "key_management_mode")
		updates.KeyManagementMode = normalizedMode
	}
	if req.KeyRoutingPolicy != nil {
		normalizedPolicy := model.NormalizeKeyRoutingPolicy(*req.KeyRoutingPolicy)
		if !model.IsValidKeyRoutingPolicy(normalizedPolicy) {
			tx.Rollback()
			return nil, fmt.Errorf("invalid key routing policy: %q", *req.KeyRoutingPolicy)
		}
		selectFields = append(selectFields, "key_routing_policy")
		updates.KeyRoutingPolicy = normalizedPolicy
	}
	if req.BaseUrls != nil {
		validatedBaseURLs, err := normalizeAndValidateBaseURLs(*req.BaseUrls)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		selectFields = append(selectFields, "base_urls")
		updates.BaseUrls = validatedBaseURLs
	}
	if req.Model != nil {
		selectFields = append(selectFields, "model")
		updates.Model = *req.Model
	}
	if req.CustomModel != nil {
		selectFields = append(selectFields, "custom_model")
		updates.CustomModel = *req.CustomModel
	}
	if req.Proxy != nil {
		selectFields = append(selectFields, "proxy")
		updates.Proxy = *req.Proxy
	}
	if req.AutoSync != nil {
		selectFields = append(selectFields, "auto_sync")
		updates.AutoSync = *req.AutoSync
	}
	if req.AutoGroup != nil {
		selectFields = append(selectFields, "auto_group")
		updates.AutoGroup = *req.AutoGroup
	}
	if req.CustomHeader != nil {
		selectFields = append(selectFields, "custom_header")
		updates.CustomHeader = *req.CustomHeader
	}
	if req.ChannelProxy != nil {
		normalizedProxy := model.NormalizeChannelProxy(req.ChannelProxy)
		if err := model.ValidateChannelProxy(normalizedProxy); err != nil {
			tx.Rollback()
			return nil, err
		}
		selectFields = append(selectFields, "channel_proxy")
		updates.ChannelProxy = normalizedProxy
	}
	if req.ParamOverride != nil {
		selectFields = append(selectFields, "param_override")
		updates.ParamOverride = req.ParamOverride
	}
	if req.MatchRegex != nil {
		selectFields = append(selectFields, "match_regex")
		updates.MatchRegex = req.MatchRegex
	}
	if req.UpstreamSiteID != nil {
		selectFields = append(selectFields, "upstream_site_id")
		updates.UpstreamSiteID = *req.UpstreamSiteID
	}
	if req.UpstreamSource != nil {
		selectFields = append(selectFields, "upstream_source")
		updates.UpstreamSource = *req.UpstreamSource
	}

	if len(selectFields) > 0 {
		if err := tx.Model(&model.Channel{}).Where("id = ?", req.ID).Select(selectFields).Updates(&updates).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update channel: %w", err)
		}
	}

	if len(req.KeysToDelete) > 0 {
		if err := tx.Where("channel_key_id IN ? AND channel_id = ?", req.KeysToDelete, req.ID).Delete(&model.RouteTargetOverride{}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to delete route target overrides for channel keys: %w", err)
		}
		if err := tx.Where("id IN ? AND channel_id = ?", req.KeysToDelete, req.ID).Delete(&model.ChannelKey{}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to delete channel keys: %w", err)
		}
	}

	if len(req.KeysToUpdate) > 0 {
		for _, ku := range req.KeysToUpdate {
			updates := map[string]interface{}{}
			if ku.Enabled != nil {
				updates["enabled"] = *ku.Enabled
			}
			if ku.ChannelKey != nil {
				updates["channel_key"] = *ku.ChannelKey
			}
			if ku.SourceType != nil {
				normalizedSourceType := model.NormalizeChannelKeySourceType(*ku.SourceType)
				if !model.IsValidChannelKeySourceType(normalizedSourceType) {
					tx.Rollback()
					return nil, fmt.Errorf("invalid channel key source type: %q", *ku.SourceType)
				}
				updates["source_type"] = normalizedSourceType
			}
			if ku.Remark != nil {
				updates["remark"] = *ku.Remark
			}
			if ku.AllowedModels != nil {
				updates["allowed_models"] = model.NormalizeChannelKeyAllowedModels(*ku.AllowedModels)
			}
			if ku.RequestCapabilities != nil {
				updates["request_capabilities"] = model.NormalizeChannelKeyRequestCapabilities(*ku.RequestCapabilities)
			}
			if ku.UpstreamSiteID != nil {
				updates["upstream_site_id"] = *ku.UpstreamSiteID
			}
			if ku.UpstreamKeyName != nil {
				updates["upstream_key_name"] = *ku.UpstreamKeyName
			}
			if len(updates) == 0 {
				continue
			}
			if err := tx.Model(&model.ChannelKey{}).
				Where("id = ? AND channel_id = ?", ku.ID, req.ID).
				Updates(updates).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to update channel key %d: %w", ku.ID, err)
			}
		}
	}

	if len(req.KeysToAdd) > 0 {
		newKeys := make([]model.ChannelKey, 0, len(req.KeysToAdd))
		for _, ka := range req.KeysToAdd {
			allowedModels := model.NormalizeChannelKeyAllowedModels(ka.AllowedModels)
			requestCapabilities := model.NormalizeChannelKeyRequestCapabilities(ka.RequestCapabilities)
			normalizedSourceType := model.NormalizeChannelKeySourceType(ka.SourceType)
			if !model.IsValidChannelKeySourceType(normalizedSourceType) {
				tx.Rollback()
				return nil, fmt.Errorf("invalid channel key source type: %q", ka.SourceType)
			}
			newKeys = append(newKeys, model.ChannelKey{
				ChannelID:           req.ID,
				Enabled:             ka.Enabled,
				ChannelKey:          ka.ChannelKey,
				SourceType:          normalizedSourceType,
				Remark:              ka.Remark,
				AllowedModels:       allowedModels,
				RequestCapabilities: requestCapabilities,
				UpstreamSiteID:      ka.UpstreamSiteID,
				UpstreamKeyName:     ka.UpstreamKeyName,
			})
		}
		if err := tx.Create(&newKeys).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create channel keys: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	if err := channelRefreshCacheByID(req.ID, ctx); err != nil {
		return nil, err
	}

	channel, _ := channelCache.Get(req.ID)
	lostConfiguredKeyModels := make([]string, 0)
	for modelName := range oldConfiguredKeyModels {
		if channel.HasConfiguredKeyForRequest(modelName, channel.RequestCapabilityForModel(modelName)) {
			continue
		}
		lostConfiguredKeyModels = append(lostConfiguredKeyModels, modelName)
	}
	if len(lostConfiguredKeyModels) > 0 {
		keys := make([]model.GroupIDAndLLMName, 0, len(lostConfiguredKeyModels))
		for _, modelName := range lostConfiguredKeyModels {
			keys = append(keys, model.GroupIDAndLLMName{ChannelID: req.ID, ModelName: modelName})
		}
		if err := GroupItemBatchDelByChannelAndModels(keys, ctx); err != nil {
			return nil, fmt.Errorf("failed to delete stale group items after key changes: %w", err)
		}
		if err := RouteTargetOverrideDeleteByChannelAndModels(req.ID, lostConfiguredKeyModels, ctx); err != nil {
			return nil, fmt.Errorf("failed to delete stale route target overrides after key changes: %w", err)
		}
		if err := channelRefreshCacheByID(req.ID, ctx); err != nil {
			return nil, err
		}
		channel, _ = channelCache.Get(req.ID)
	}

	if modelsChanged {
		newModel := oldChannel.Model
		if req.Model != nil {
			newModel = *req.Model
		}
		newCustomModel := oldChannel.CustomModel
		if req.CustomModel != nil {
			newCustomModel = *req.CustomModel
		}
		newModelSet := xstrings.SplitTrimCompact(",", newModel, newCustomModel)
		deletedModels, _ := diff.Diff(oldModelSet, newModelSet)
		if len(deletedModels) > 0 {
			keys := make([]model.GroupIDAndLLMName, 0, len(deletedModels))
			seen := make(map[string]struct{}, len(deletedModels))
			for _, modelName := range deletedModels {
				if modelName == "" {
					continue
				}
				cacheKey := fmt.Sprintf("%d|%s", req.ID, modelName)
				if _, ok := seen[cacheKey]; ok {
					continue
				}
				seen[cacheKey] = struct{}{}
				keys = append(keys, model.GroupIDAndLLMName{ChannelID: req.ID, ModelName: modelName})
			}
			if len(keys) > 0 {
				if err := GroupItemBatchDelByChannelAndModels(keys, ctx); err != nil {
					return nil, fmt.Errorf("failed to delete stale group items: %w", err)
				}
			}
			if err := RouteTargetOverrideDeleteByChannelAndModels(req.ID, deletedModels, ctx); err != nil {
				return nil, fmt.Errorf("failed to delete stale route target overrides: %w", err)
			}
		}
	}

	if modelsChanged {
		if err := channelRefreshCacheByID(req.ID, ctx); err != nil {
			return nil, err
		}
		channel, _ = channelCache.Get(req.ID)
	}
	return &channel, nil
}

func ChannelEnabled(id int, enabled bool, ctx context.Context) error {
	oldChannel, ok := channelCache.Get(id)
	if !ok {
		return fmt.Errorf("channel not found")
	}
	if err := db.GetDB().WithContext(ctx).Model(&model.Channel{}).Where("id = ?", id).Update("enabled", enabled).Error; err != nil {
		return err
	}
	oldChannel.Enabled = enabled
	channelCache.Set(id, oldChannel)
	return nil
}

func ChannelDel(id int, ctx context.Context) error {
	ch, ok := channelCache.Get(id)
	if !ok {
		return fmt.Errorf("channel not found")
	}

	tx := db.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var affectedGroupIDs []int
	if err := tx.Model(&model.GroupItem{}).
		Where("channel_id = ?", id).
		Pluck("group_id", &affectedGroupIDs).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get affected groups: %w", err)
	}

	if err := tx.Where("channel_id = ?", id).Delete(&model.GroupItem{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete group items: %w", err)
	}
	if err := tx.Where("channel_id = ?", id).Delete(&model.RouteTargetOverride{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete route target overrides: %w", err)
	}
	if err := tx.Where("channel_id = ?", id).Delete(&model.ChannelKey{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete channel keys: %w", err)
	}
	if err := tx.Where("channel_id = ?", id).Delete(&model.StatsChannel{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete channel stats: %w", err)
	}
	if err := tx.Delete(&model.Channel{}, id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	channelCache.Del(id)
	for _, k := range ch.Keys {
		if k.ID != 0 {
			channelKeyCache.Del(k.ID)
		}
	}
	StatsChannelDel(id)

	for _, groupID := range affectedGroupIDs {
		if err := groupRefreshCacheByID(groupID, ctx); err != nil {
			log.Warnf("failed to refresh group cache for group %d: %v", groupID, err)
		}
	}

	return nil
}

func ChannelLLMList(ctx context.Context) ([]model.LLMChannel, error) {
	inventory, err := CapabilityInventory(ctx)
	if err != nil {
		return nil, err
	}
	models := make([]model.LLMChannel, 0, len(inventory.ServiceableModels))
	for _, item := range inventory.ServiceableModels {
		models = append(models, model.LLMChannel{
			Name:                item.Name,
			Enabled:             item.Enabled,
			ChannelID:           item.ChannelID,
			ChannelName:         item.ChannelName,
			KeyCount:            item.KeyCount,
			RequestCapabilities: item.RequestCapabilities,
			InventorySource:     item.InventorySource,
		})
	}
	return models, nil
}

func ChannelGet(id int, ctx context.Context) (*model.Channel, error) {
	channel, ok := channelCache.Get(id)
	if !ok {
		return nil, fmt.Errorf("channel not found")
	}
	return &channel, nil
}

func ChannelKeyGet(id int) (model.ChannelKey, bool) {
	if id <= 0 {
		return model.ChannelKey{}, false
	}
	return channelKeyCache.Get(id)
}

func channelRefreshCache(ctx context.Context) error {
	channels := []model.Channel{}
	if err := db.GetDB().WithContext(ctx).
		Preload("Keys").
		Preload("Stats").
		Find(&channels).Error; err != nil {
		log.Warnf("failed to get channels: %v", err)
		return err
	}
	channelCache.Clear()
	channelKeyCache.Clear()
	channelKeyCacheNeedUpdateLock.Lock()
	channelKeyCacheNeedUpdate = make(map[int]struct{})
	channelKeyCacheNeedUpdateLock.Unlock()
	for _, channel := range channels {
		channelCache.Set(channel.ID, channel)
		for _, k := range channel.Keys {
			if k.ID != 0 {
				channelKeyCache.Set(k.ID, k)
			}
		}
	}
	return nil
}

func channelRefreshCacheByID(id int, ctx context.Context) error {
	if old, ok := channelCache.Get(id); ok {
		for _, k := range old.Keys {
			if k.ID != 0 {
				channelKeyCache.Del(k.ID)
			}
		}
	}
	var channel model.Channel
	if err := db.GetDB().WithContext(ctx).
		Preload("Keys").
		Preload("Stats").
		First(&channel, id).Error; err != nil {
		return err
	}
	channelCache.Set(channel.ID, channel)
	for _, k := range channel.Keys {
		if k.ID != 0 {
			channelKeyCache.Set(k.ID, k)
		}
	}
	return nil
}
