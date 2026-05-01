package op

import (
	"strings"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestSettingSetStringRejectsInvalidValueBeforePersisting(t *testing.T) {
	ctx := setupOpTestDB(t)

	var before model.Setting
	if err := db.GetDB().WithContext(ctx).First(&before, "key = ?", model.SettingKeyDynamicRoutingMode).Error; err != nil {
		t.Fatalf("load before setting error = %v", err)
	}

	err := SettingSetString(model.SettingKeyDynamicRoutingMode, "broken-mode")
	if err == nil || !strings.Contains(err.Error(), "setting value must be one of") {
		t.Fatalf("SettingSetString() error = %v, want invalid dynamic routing mode", err)
	}

	var after model.Setting
	if err := db.GetDB().WithContext(ctx).First(&after, "key = ?", model.SettingKeyDynamicRoutingMode).Error; err != nil {
		t.Fatalf("load after setting error = %v", err)
	}
	if after.Value != before.Value {
		t.Fatalf("stored value = %q, want unchanged %q", after.Value, before.Value)
	}
}

func TestSettingSetIntRejectsNegativeValueBeforePersisting(t *testing.T) {
	ctx := setupOpTestDB(t)

	var before model.Setting
	if err := db.GetDB().WithContext(ctx).First(&before, "key = ?", model.SettingKeyStatsSaveInterval).Error; err != nil {
		t.Fatalf("load before setting error = %v", err)
	}

	err := SettingSetInt(model.SettingKeyStatsSaveInterval, -1)
	if err == nil || !strings.Contains(err.Error(), "non-negative integer") {
		t.Fatalf("SettingSetInt() error = %v, want non-negative integer validation", err)
	}

	var after model.Setting
	if err := db.GetDB().WithContext(ctx).First(&after, "key = ?", model.SettingKeyStatsSaveInterval).Error; err != nil {
		t.Fatalf("load after setting error = %v", err)
	}
	if after.Value != before.Value {
		t.Fatalf("stored value = %q, want unchanged %q", after.Value, before.Value)
	}
}

func TestSettingSetIntCanRepairCorruptedCachedValue(t *testing.T) {
	ctx := setupOpTestDB(t)
	settingCache.Set(model.SettingKeyStatsSaveInterval, "broken")
	if err := SettingSetInt(model.SettingKeyStatsSaveInterval, 15); err != nil {
		t.Fatalf("SettingSetInt() error = %v, want nil", err)
	}

	stored, err := SettingGetString(model.SettingKeyStatsSaveInterval)
	if err != nil {
		t.Fatalf("SettingGetString() error = %v", err)
	}
	if stored != "15" {
		t.Fatalf("stored setting = %q, want 15", stored)
	}

	var row model.Setting
	if err := db.GetDB().WithContext(ctx).First(&row, "key = ?", model.SettingKeyStatsSaveInterval).Error; err != nil {
		t.Fatalf("load persisted setting error = %v", err)
	}
	if row.Value != "15" {
		t.Fatalf("persisted setting = %q, want 15", row.Value)
	}
}
