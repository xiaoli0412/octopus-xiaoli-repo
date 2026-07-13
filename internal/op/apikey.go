package op

import (
	"context"
	"fmt"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/cache"
)

var apiKeyCache = cache.New[int, model.APIKey](16)
var apiKeyIDMap = cache.New[string, int](16)

func APIKeyCreate(key *model.APIKey, ctx context.Context) error {
	key.SupportedModels = normalizeAPIKeySupportedModels(key.SupportedModels)
	if err := db.GetDB().WithContext(ctx).Create(key).Error; err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}
	apiKeyCache.Set(key.ID, *key)
	apiKeyIDMap.Set(key.APIKey, key.ID)
	return nil
}

func APIKeyUpdate(key *model.APIKey, ctx context.Context) error {
	existing, ok := apiKeyCache.Get(key.ID)
	key.SupportedModels = normalizeAPIKeySupportedModels(key.SupportedModels)
	if !ok {
		return fmt.Errorf("API key not found")
	}
	if err := db.GetDB().WithContext(ctx).Omit("api_key").Save(key).Error; err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}
	key.APIKey = existing.APIKey
	apiKeyCache.Set(key.ID, *key)
	return nil
}

func APIKeyList(ctx context.Context) ([]model.APIKey, error) {
	keys := apiKeyCache.Values()
	return keys, nil
}

func APIKeyGet(id int, ctx context.Context) (model.APIKey, error) {
	apiKey, ok := apiKeyCache.Get(id)
	if !ok {
		return model.APIKey{}, fmt.Errorf("API key not found")
	}
	return apiKey, nil
}

func APIKeyGetByAPIKey(apiKey string, ctx context.Context) (model.APIKey, error) {
	id, ok := apiKeyIDMap.Get(apiKey)
	if !ok {
		return model.APIKey{}, fmt.Errorf("API key not found")
	}
	return APIKeyGet(id, ctx)
}

func APIKeyDelete(id int, ctx context.Context) error {
	k := model.APIKey{ID: id}
	if cached, ok := apiKeyCache.Get(id); ok {
		k = cached
	} else if err := db.GetDB().WithContext(ctx).First(&k, id).Error; err != nil {
		return fmt.Errorf("API key not found")
	}
	if err := StatsAPIKeyDel(id); err != nil {
		return fmt.Errorf("failed to delete stats API key: %v", err)
	}
	ResetCostAlerts(uint(id))
	result := db.GetDB().WithContext(ctx).Delete(&k)
	if result.RowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}
	if result.Error != nil {
		return fmt.Errorf("failed to delete API key: %w", result.Error)
	}
	apiKeyCache.Del(k.ID)
	if strings.TrimSpace(k.APIKey) != "" {
		apiKeyIDMap.Del(k.APIKey)
	}
	return nil
}

func apiKeyRefreshCache(ctx context.Context) error {
	apiKeys := []model.APIKey{}
	if err := db.GetDB().WithContext(ctx).Find(&apiKeys).Error; err != nil {
		return err
	}
	apiKeyCache.Clear()
	apiKeyIDMap.Clear()
	for _, apiKey := range apiKeys {
		apiKey.SupportedModels = normalizeAPIKeySupportedModels(apiKey.SupportedModels)
		apiKeyCache.Set(apiKey.ID, apiKey)
		apiKeyIDMap.Set(apiKey.APIKey, apiKey.ID)
	}
	return nil
}
