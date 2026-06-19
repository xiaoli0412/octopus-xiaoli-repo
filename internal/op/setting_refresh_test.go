package op

import (
	"strings"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestSettingRefreshCacheRepairsEmptyAuthTokenSecret(t *testing.T) {
	ctx := SetupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Model(&model.Setting{Key: model.SettingKeyAuthTokenSecret}).Update("Value", "").Error; err != nil {
		t.Fatalf("update auth token secret error = %v", err)
	}

	if err := settingRefreshCache(ctx); err != nil {
		t.Fatalf("settingRefreshCache() error = %v", err)
	}

	var stored model.Setting
	if err := db.GetDB().WithContext(ctx).First(&stored, "key = ?", model.SettingKeyAuthTokenSecret).Error; err != nil {
		t.Fatalf("load stored auth token secret error = %v", err)
	}
	if strings.TrimSpace(stored.Value) == "" {
		t.Fatal("stored auth token secret is still empty")
	}

	cached, err := SettingGetString(model.SettingKeyAuthTokenSecret)
	if err != nil {
		t.Fatalf("SettingGetString() error = %v", err)
	}
	if strings.TrimSpace(cached) == "" {
		t.Fatal("cached auth token secret is still empty")
	}
	if cached != stored.Value {
		t.Fatalf("cached secret = %q, want %q", cached, stored.Value)
	}
}
