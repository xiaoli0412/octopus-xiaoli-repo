package op

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/cache"
	"gorm.io/gorm"
)

var settingCache = cache.New[model.SettingKey, string](16)

func IsSecretSettingKey(key model.SettingKey) bool {
	switch key {
	case model.SettingKeyAuthTokenSecret,
		model.SettingKeyAuthTokenSecretSecondary,
		model.SettingKeyAIAutomationAPIKey:
		return true
	default:
		return false
	}
}

func generateSettingSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate setting secret: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

func SettingList(ctx context.Context) ([]model.Setting, error) {
	settings := make([]model.Setting, 0, settingCache.Len())
	for key, value := range settingCache.GetAll() {
		if IsSecretSettingKey(key) {
			continue
		}
		settings = append(settings, model.Setting{
			Key:   key,
			Value: value,
		})
	}
	return settings, nil
}

func SettingGetString(key model.SettingKey) (string, error) {
	setting, ok := settingCache.Get(key)
	if !ok {
		return "", fmt.Errorf("setting not found")
	}
	return setting, nil
}

func SettingSetString(key model.SettingKey, value string) error {
	valueCache, ok := settingCache.Get(key)
	if !ok {
		return fmt.Errorf("setting not found")
	}
	setting := model.Setting{Key: key, Value: value}
	if err := setting.Validate(); err != nil {
		return err
	}
	if valueCache == value {
		return nil
	}
	result := db.GetDB().Model(&model.Setting{Key: key}).Update("Value", value)
	if result.Error != nil {
		return fmt.Errorf("failed to set setting: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("failed to set setting, key not found")
	}
	settingCache.Set(key, value)
	return nil
}

func SettingGetInt(key model.SettingKey) (int, error) {
	setting, ok := settingCache.Get(key)
	if !ok {
		return 0, fmt.Errorf("setting not found")
	}
	return strconv.Atoi(setting)
}

func SettingGetBool(key model.SettingKey) (bool, error) {
	setting, ok := settingCache.Get(key)
	if !ok {
		return false, fmt.Errorf("setting not found")
	}
	return strconv.ParseBool(setting)
}

func SettingSetInt(key model.SettingKey, value int) error {
	valueCache, ok := settingCache.Get(key)
	if !ok {
		return fmt.Errorf("setting not found")
	}
	setting := model.Setting{Key: key, Value: strconv.Itoa(value)}
	if err := setting.Validate(); err != nil {
		return err
	}
	if valueCache == setting.Value {
		return nil
	}
	result := db.GetDB().Model(&model.Setting{Key: key}).Update("Value", value)
	if result.Error != nil {
		return fmt.Errorf("failed to set setting: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("failed to set setting, key not found")
	}
	settingCache.Set(key, strconv.Itoa(value))
	return nil
}

// RotateAuthTokenSecret performs JWT secret rotation by moving the current
// primary secret to secondary and generating a new primary. After rotation,
// tokens signed with the old primary remain valid via the secondary until
// secondary is cleared or overwritten by a subsequent rotation.
func RotateAuthTokenSecret(ctx context.Context) error {
	currentPrimary, ok := settingCache.Get(model.SettingKeyAuthTokenSecret)
	if !ok || strings.TrimSpace(currentPrimary) == "" {
		return fmt.Errorf("auth token secret is not initialized")
	}

	newSecret, err := generateSettingSecret()
	if err != nil {
		return fmt.Errorf("generate new auth token secret: %w", err)
	}

	conn := db.GetDB().WithContext(ctx)
	err = conn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Setting{Key: model.SettingKeyAuthTokenSecretSecondary}).
			Update("Value", currentPrimary).Error; err != nil {
			return fmt.Errorf("update secondary secret: %w", err)
		}
		if err := tx.Model(&model.Setting{Key: model.SettingKeyAuthTokenSecret}).
			Update("Value", newSecret).Error; err != nil {
			return fmt.Errorf("update primary secret: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	settingCache.Set(model.SettingKeyAuthTokenSecretSecondary, currentPrimary)
	settingCache.Set(model.SettingKeyAuthTokenSecret, newSecret)
	return nil
}

func settingRefreshCache(ctx context.Context) error {
	conn := db.GetDB().WithContext(ctx)

	var settings []model.Setting
	if err := conn.Find(&settings).Error; err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}

	existingKeys := make(map[model.SettingKey]bool)
	settingIndex := make(map[model.SettingKey]int, len(settings))
	for idx, setting := range settings {
		existingKeys[setting.Key] = true
		settingIndex[setting.Key] = idx
	}

	defaultSettings := model.DefaultSettings()
	missingSettings := make([]model.Setting, 0, len(defaultSettings))

	for _, defaultSetting := range defaultSettings {
		if defaultSetting.Key == model.SettingKeyAuthTokenSecret && defaultSetting.Value == "" {
			if idx, ok := settingIndex[defaultSetting.Key]; ok {
				if strings.TrimSpace(settings[idx].Value) == "" {
					secret, err := generateSettingSecret()
					if err != nil {
						return fmt.Errorf("failed to generate auth token secret: %w", err)
					}
					settings[idx].Value = secret
					if err := conn.Model(&model.Setting{Key: defaultSetting.Key}).Update("Value", secret).Error; err != nil {
						return fmt.Errorf("failed to persist auth token secret: %w", err)
					}
				}
				continue
			}
			secret, err := generateSettingSecret()
			if err != nil {
				return fmt.Errorf("failed to generate auth token secret: %w", err)
			}
			defaultSetting.Value = secret
		}
		if !existingKeys[defaultSetting.Key] {
			missingSettings = append(missingSettings, defaultSetting)
		}
	}

	if len(missingSettings) > 0 {
		if err := conn.CreateInBatches(missingSettings, len(missingSettings)).Error; err != nil {
			return fmt.Errorf("failed to create missing settings: %w", err)
		}
		settings = append(settings, missingSettings...)
	}
	settingCache.Clear()
	for _, setting := range settings {
		settingCache.Set(setting.Key, setting.Value)
	}
	return nil
}
