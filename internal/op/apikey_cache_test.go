package op

import (
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestAPIKeyDeleteRemovesAPIKeyLookupFromCache(t *testing.T) {
	ctx := setupOpTestDB(t)

	apiKey := &model.APIKey{APIKey: "sk-octopus-delete-lookup", Enabled: true}
	if err := APIKeyCreate(apiKey, ctx); err != nil {
		t.Fatalf("APIKeyCreate() error = %v", err)
	}
	if err := APIKeyDelete(apiKey.ID, ctx); err != nil {
		t.Fatalf("APIKeyDelete() error = %v", err)
	}

	if _, err := APIKeyGet(apiKey.ID, ctx); err == nil {
		t.Fatalf("APIKeyGet() expected deleted key to be absent")
	}
	if _, err := APIKeyGetByAPIKey(apiKey.APIKey, ctx); err == nil {
		t.Fatalf("APIKeyGetByAPIKey() expected deleted key token to be absent")
	}
}
func TestAPIKeyRefreshCacheRemovesDeletedKeyFromCache(t *testing.T) {
	ctx := setupOpTestDB(t)

	apiKey := &model.APIKey{APIKey: "sk-octopus-refresh-delete", Enabled: true}
	if err := APIKeyCreate(apiKey, ctx); err != nil {
		t.Fatalf("APIKeyCreate() error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Delete(&model.APIKey{}, apiKey.ID).Error; err != nil {
		t.Fatalf("delete api key error = %v", err)
	}

	if err := apiKeyRefreshCache(ctx); err != nil {
		t.Fatalf("apiKeyRefreshCache() error = %v", err)
	}
	if _, err := APIKeyGet(apiKey.ID, ctx); err == nil {
		t.Fatalf("APIKeyGet() expected deleted key to be absent after full refresh")
	}
	if _, err := APIKeyGetByAPIKey(apiKey.APIKey, ctx); err == nil {
		t.Fatalf("APIKeyGetByAPIKey() expected deleted key token to be absent after full refresh")
	}
}

func TestAPIKeyCreateNormalizesSupportedModels(t *testing.T) {
	ctx := setupOpTestDB(t)

	apiKey := &model.APIKey{
		Name:            "normalize-create",
		APIKey:          "sk-octopus-normalize-create",
		Enabled:         true,
		SupportedModels: " GPT-4O , o1-mini,GPT-4O ",
	}
	if err := APIKeyCreate(apiKey, ctx); err != nil {
		t.Fatalf("APIKeyCreate() error = %v", err)
	}

	cached, err := APIKeyGet(apiKey.ID, ctx)
	if err != nil {
		t.Fatalf("APIKeyGet() error = %v", err)
	}
	if cached.SupportedModels != "gpt-4o,o1-mini" {
		t.Fatalf("cached SupportedModels = %q, want %q", cached.SupportedModels, "gpt-4o,o1-mini")
	}

	var stored model.APIKey
	if err := db.GetDB().WithContext(ctx).First(&stored, apiKey.ID).Error; err != nil {
		t.Fatalf("query stored api key error = %v", err)
	}
	if stored.SupportedModels != "gpt-4o,o1-mini" {
		t.Fatalf("stored SupportedModels = %q, want %q", stored.SupportedModels, "gpt-4o,o1-mini")
	}
}

func TestAPIKeyUpdateNormalizesSupportedModels(t *testing.T) {
	ctx := setupOpTestDB(t)

	apiKey := &model.APIKey{
		Name:            "normalize-update",
		APIKey:          "sk-octopus-normalize-update",
		Enabled:         true,
		SupportedModels: "gpt-4o",
	}
	if err := APIKeyCreate(apiKey, ctx); err != nil {
		t.Fatalf("APIKeyCreate() error = %v", err)
	}

	apiKey.SupportedModels = " O1-mini , gpt-4o, o1-mini "
	if err := APIKeyUpdate(apiKey, ctx); err != nil {
		t.Fatalf("APIKeyUpdate() error = %v", err)
	}

	updated, err := APIKeyGet(apiKey.ID, ctx)
	if err != nil {
		t.Fatalf("APIKeyGet() after update error = %v", err)
	}
	if updated.SupportedModels != "gpt-4o,o1-mini" {
		t.Fatalf("updated SupportedModels = %q, want %q", updated.SupportedModels, "gpt-4o,o1-mini")
	}
}

func TestAPIKeyRefreshCacheNormalizesSupportedModels(t *testing.T) {
	ctx := setupOpTestDB(t)

	raw := model.APIKey{
		Name:            "normalize-refresh",
		APIKey:          "sk-octopus-normalize-refresh",
		Enabled:         true,
		SupportedModels: " GPT-4O , o1-mini , GPT-4O ",
	}
	if err := db.GetDB().WithContext(ctx).Create(&raw).Error; err != nil {
		t.Fatalf("direct create api key error = %v", err)
	}

	if err := apiKeyRefreshCache(ctx); err != nil {
		t.Fatalf("apiKeyRefreshCache() error = %v", err)
	}

	cached, err := APIKeyGetByAPIKey(raw.APIKey, ctx)
	if err != nil {
		t.Fatalf("APIKeyGetByAPIKey() error = %v", err)
	}
	if cached.SupportedModels != "gpt-4o,o1-mini" {
		t.Fatalf("cached SupportedModels = %q, want %q", cached.SupportedModels, "gpt-4o,o1-mini")
	}
}
