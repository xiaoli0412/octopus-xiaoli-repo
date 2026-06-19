package op

import (
	"context"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm/clause"
)

func setupOpTestDB(t *testing.T) context.Context {
	return SetupOpTestDB(t)
}

func ptr[T any](v T) *T {
	return &v
}

func createConfiguredTestChannel(t *testing.T, ctx context.Context, name, models, customModel string) *model.Channel {
	t.Helper()

	channel := &model.Channel{
		Name:        name,
		Enabled:     true,
		Model:       models,
		CustomModel: customModel,
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	key := model.ChannelKey{
		ChannelID:  channel.ID,
		Enabled:    true,
		ChannelKey: name + "-key",
	}
	if err := db.GetDB().WithContext(ctx).Create(&key).Error; err != nil {
		t.Fatalf("create channel key error = %v", err)
	}
	if err := channelRefreshCacheByID(channel.ID, ctx); err != nil {
		t.Fatalf("channelRefreshCacheByID() error = %v", err)
	}

	refreshed, err := ChannelGet(channel.ID, ctx)
	if err != nil {
		t.Fatalf("ChannelGet() error = %v", err)
	}
	return refreshed
}

func createTestUser(t *testing.T, ctx context.Context, username, password string) model.User {
	t.Helper()

	user := model.User{Username: username, Password: password}
	if err := user.HashPassword(); err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&user).Error; err != nil {
		t.Fatalf("create user error = %v", err)
	}
	return user
}

func setTestSetting(t *testing.T, ctx context.Context, key model.SettingKey, value string) {
	t.Helper()

	setting := model.Setting{Key: key, Value: value}
	if err := db.GetDB().WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).
		Create(&setting).Error; err != nil {
		t.Fatalf("set setting %s error = %v", key, err)
	}
}
