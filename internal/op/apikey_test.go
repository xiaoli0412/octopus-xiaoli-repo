package op

import (
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestAPIKeyCreateAndUpdateNormalizeSupportedModels(t *testing.T) {
	ctx := SetupOpTestDB(t)

	apiKey := &model.APIKey{
		Name:            "client-normalize-supported-models",
		APIKey:          "sk-client-normalize-supported-models",
		Enabled:         true,
		SupportedModels: " beta,alpha,beta ",
	}
	if err := APIKeyCreate(apiKey, ctx); err != nil {
		t.Fatalf("APIKeyCreate() error = %v", err)
	}
	if apiKey.SupportedModels != "alpha,beta" {
		t.Fatalf("created supported_models = %q, want alpha,beta", apiKey.SupportedModels)
	}

	apiKey.SupportedModels = " gpt-4o,alpha,gpt-4o "
	if err := APIKeyUpdate(apiKey, ctx); err != nil {
		t.Fatalf("APIKeyUpdate() error = %v", err)
	}
	refreshed, err := APIKeyGet(apiKey.ID, ctx)
	if err != nil {
		t.Fatalf("APIKeyGet() error = %v", err)
	}
	if refreshed.SupportedModels != "alpha,gpt-4o" {
		t.Fatalf("updated supported_models = %q, want alpha,gpt-4o", refreshed.SupportedModels)
	}
}
