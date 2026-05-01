package op

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm/clause"
)

func setupOpTestDB(t *testing.T) context.Context {
	t.Helper()
	t.Setenv(bootstrapAdminUsernameEnv, bootstrapAdminDefaultUsername)
	t.Setenv(bootstrapAdminPasswordEnv, "admin")

	stopOpTestAsyncState(t)
	resetOpTestState()
	if db.GetDB() != nil {
		_ = db.Close()
	}

	dbPath := filepath.Join(t.TempDir(), "octopus-test.db")
	if err := db.InitDB("sqlite", dbPath, false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}

	t.Cleanup(func() {
		stopOpTestAsyncState(t)
		if db.GetDB() != nil {
			_ = db.Close()
		}
		resetOpTestState()
	})

	return context.Background()
}

func stopOpTestAsyncState(t *testing.T) {
	t.Helper()
	if err := CancelAllAITasks(); err != nil {
		t.Fatalf("CancelAllAITasks() error = %v", err)
	}
}

func resetOpTestState() {
	channelCache.Clear()
	channelKeyCache.Clear()
	channelKeyCacheNeedUpdateLock.Lock()
	channelKeyCacheNeedUpdate = make(map[int]struct{})
	channelKeyCacheNeedUpdateLock.Unlock()

	groupCache.Clear()
	groupMap.Clear()

	apiKeyCache.Clear()
	apiKeyIDMap.Clear()

	llmModelCache.Clear()
	routeTargetOverrideCache.Clear()
	routeTargetOverrideIndex.Clear()
	settingCache.Clear()
	relayLogCache = make([]model.RelayLog, 0, relayLogMaxSize)
	userCache = model.User{}

	statsDailyCacheLock.Lock()
	statsDailyCache = model.StatsDaily{}
	statsDailyCacheLock.Unlock()

	statsTotalCacheLock.Lock()
	statsTotalCache = model.StatsTotal{}
	statsTotalCacheLock.Unlock()

	statsHourlyCacheLock.Lock()
	statsHourlyCache = [24]model.StatsHourly{}
	statsHourlyCacheLock.Unlock()

	statsChannelCache.Clear()
	statsChannelCacheNeedUpdateLock.Lock()
	statsChannelCacheNeedUpdate = make(map[int]struct{})
	statsChannelCacheNeedUpdateLock.Unlock()

	statsModelCache.Clear()
	statsModelCacheNeedUpdateLock.Lock()
	statsModelCacheNeedUpdate = make(map[int]struct{})
	statsModelCacheNeedUpdateLock.Unlock()

	statsAPIKeyCache.Clear()
	statsAPIKeyCacheNeedUpdateLock.Lock()
	statsAPIKeyCacheNeedUpdate = make(map[int]struct{})
	statsAPIKeyCacheNeedUpdateLock.Unlock()

	probeEventCacheLock.Lock()
	probeEventCache = make([]model.ProbeEvent, 0, probeEventMaxSize)
	probeEventCacheLock.Unlock()

	relayLogStreamTokensLock.Lock()
	relayLogStreamTokens = make(map[string]relayLogStreamToken)
	relayLogStreamTokensLock.Unlock()
	relayLogStreamTokenTTL = 15 * time.Minute
	relayLogStreamTokenNow = time.Now

	aiTaskStartGroup.Range(func(key, value any) bool {
		aiTaskStartGroup.Delete(key)
		return true
	})
	aiTaskCancelFuncs.Range(func(key, value any) bool {
		aiTaskCancelFuncs.Delete(key)
		return true
	})
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
