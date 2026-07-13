package op

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/snowflake"
)

const relayLogMaxSize = 20
const relayLogMaxSizeNoDB = 100 // 当不保存到数据库时，允许更大的缓存用于实时查询

var relayLogCache = make([]model.RelayLog, 0, relayLogMaxSize)
var relayLogCacheLock sync.Mutex

var relayLogFlushLock sync.Mutex

var relayLogSubscribers = make(map[chan model.RelayLog]struct{})
var relayLogSubscribersLock sync.RWMutex

type relayLogStreamToken struct {
	expiresAt time.Time
	sequence  uint64
}

var relayLogStreamTokens = make(map[string]relayLogStreamToken)
var relayLogStreamTokensLock sync.RWMutex

const relayLogStreamTokenMaxEntries = 4096

var relayLogStreamTokenTTL = 15 * time.Minute
var relayLogStreamTokenNow = time.Now
var relayLogStreamTokenSeq uint64

func RelayLogStreamTokenCreate() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)
	now := relayLogStreamTokenNow()

	relayLogStreamTokensLock.Lock()
	cleanupRelayLogStreamTokensLocked(now)
	relayLogStreamTokenSeq++
	relayLogStreamTokens[token] = relayLogStreamToken{expiresAt: now.Add(relayLogStreamTokenTTL), sequence: relayLogStreamTokenSeq}
	trimRelayLogStreamTokensLocked(token)
	relayLogStreamTokensLock.Unlock()

	return token, nil
}

func RelayLogStreamTokenVerify(token string) bool {
	relayLogStreamTokensLock.Lock()
	defer relayLogStreamTokensLock.Unlock()

	entry, ok := relayLogStreamTokens[token]
	if !ok {
		return false
	}
	if relayLogStreamTokenNow().After(entry.expiresAt) {
		delete(relayLogStreamTokens, token)
		return false
	}
	return true
}

func RelayLogStreamTokenConsume(token string) bool {
	relayLogStreamTokensLock.Lock()
	defer relayLogStreamTokensLock.Unlock()

	entry, ok := relayLogStreamTokens[token]
	if !ok {
		return false
	}
	if relayLogStreamTokenNow().After(entry.expiresAt) {
		delete(relayLogStreamTokens, token)
		return false
	}
	delete(relayLogStreamTokens, token)
	return true
}

func RelayLogStreamTokenRevoke(token string) {
	relayLogStreamTokensLock.Lock()
	delete(relayLogStreamTokens, token)
	relayLogStreamTokensLock.Unlock()
}

func cleanupRelayLogStreamTokens() {
	relayLogStreamTokensLock.Lock()
	defer relayLogStreamTokensLock.Unlock()

	cleanupRelayLogStreamTokensLocked(relayLogStreamTokenNow())
}

func cleanupRelayLogStreamTokensLocked(now time.Time) {
	for token, entry := range relayLogStreamTokens {
		if now.After(entry.expiresAt) {
			delete(relayLogStreamTokens, token)
		}
	}
}

func trimRelayLogStreamTokensLocked(currentToken string) {
	for len(relayLogStreamTokens) > relayLogStreamTokenMaxEntries {
		oldestToken := ""
		var oldestEntry relayLogStreamToken
		for token, entry := range relayLogStreamTokens {
			if token == currentToken && len(relayLogStreamTokens) > 1 {
				continue
			}
			if oldestToken == "" || entry.expiresAt.Before(oldestEntry.expiresAt) || (entry.expiresAt.Equal(oldestEntry.expiresAt) && entry.sequence < oldestEntry.sequence) {
				oldestToken = token
				oldestEntry = entry
			}
		}
		if oldestToken == "" {
			break
		}
		delete(relayLogStreamTokens, oldestToken)
	}
}

func RelayLogSubscribe() chan model.RelayLog {
	ch := make(chan model.RelayLog, 10)
	relayLogSubscribersLock.Lock()
	relayLogSubscribers[ch] = struct{}{}
	relayLogSubscribersLock.Unlock()
	return ch
}

func RelayLogUnsubscribe(ch chan model.RelayLog) {
	relayLogSubscribersLock.Lock()
	delete(relayLogSubscribers, ch)
	relayLogSubscribersLock.Unlock()
	close(ch)
}

func notifySubscribers(relayLog model.RelayLog) {
	relayLogSubscribersLock.RLock()
	defer relayLogSubscribersLock.RUnlock()

	for ch := range relayLogSubscribers {
		select {
		case ch <- relayLog:
		default:
		}
	}
}

func relayLogFlushToDB(ctx context.Context) error {
	relayLogFlushLock.Lock()
	defer relayLogFlushLock.Unlock()

	relayLogCacheLock.Lock()
	if len(relayLogCache) == 0 {
		relayLogCacheLock.Unlock()
		return nil
	}
	batch := make([]model.RelayLog, len(relayLogCache))
	copy(batch, relayLogCache)
	flushedUpto := len(batch)
	relayLogCacheLock.Unlock()

	result := db.GetDB().WithContext(ctx).Create(&batch)
	if result.Error != nil {
		return result.Error
	}

	relayLogCacheLock.Lock()
	if len(relayLogCache) >= flushedUpto {
		relayLogCache = relayLogCache[flushedUpto:]
	} else {
		relayLogCache = relayLogCache[:0]
	}
	if len(relayLogCache) == 0 {
		relayLogCache = make([]model.RelayLog, 0, relayLogMaxSize)
	}
	relayLogCacheLock.Unlock()

	return nil
}

func RelayLogAdd(ctx context.Context, relayLog model.RelayLog) error {
	enabled, err := SettingGetBool(model.SettingKeyRelayLogKeepEnabled)
	if err != nil {
		return err
	}
	maxSize := relayLogMaxSize
	if !enabled {
		maxSize = relayLogMaxSizeNoDB
	}
	relayLog.ID = snowflake.GenerateID()
	notifySubscribers(relayLog)

	relayLogCacheLock.Lock()
	relayLogCache = append(relayLogCache, relayLog)
	if len(relayLogCache) >= maxSize {
		if enabled {
			relayLogCacheLock.Unlock()
			return relayLogFlushToDB(ctx)
		}
		// 如果未启用日志保存，移除最旧的日志，保留最新的日志用于实时查询
		keepSize := maxSize / 2
		if len(relayLogCache) > keepSize {
			relayLogCache = relayLogCache[len(relayLogCache)-keepSize:]
		}
	}
	relayLogCacheLock.Unlock()
	return nil
}

func RelayLogSaveDBTask(ctx context.Context) error {
	log.Debugf("relay log save db task started")
	startTime := time.Now()
	defer func() {
		log.Debugf("relay log save db task finished, save time: %s", time.Since(startTime))
	}()
	enabled, err := SettingGetBool(model.SettingKeyRelayLogKeepEnabled)
	if err != nil {
		return err
	}

	if enabled {
		if err := relayLogFlushToDB(ctx); err != nil {
			return err
		}
		return relayLogCleanup(ctx)
	}

	// 如果未启用日志保存，检查缓存大小，如果超过限制则清理旧日志
	relayLogCacheLock.Lock()
	if len(relayLogCache) > relayLogMaxSizeNoDB {
		keepSize := relayLogMaxSizeNoDB / 2
		relayLogCache = relayLogCache[len(relayLogCache)-keepSize:]
	}
	relayLogCacheLock.Unlock()

	return nil
}

func relayLogCleanup(ctx context.Context) error {
	keepPeriod, err := SettingGetInt(model.SettingKeyRelayLogKeepPeriod)
	if err != nil {
		return err
	}

	if keepPeriod <= 0 {
		return nil
	}

	cutoffTime := time.Now().Add(-time.Duration(keepPeriod) * 24 * time.Hour).Unix()
	return db.GetDB().WithContext(ctx).Where("time < ?", cutoffTime).Delete(&model.RelayLog{}).Error
}

// RelayLogList 查询日志列表，支持可选的时间范围过滤
// startTime 和 endTime 为 nil 时表示不限制时间范围
func RelayLogList(ctx context.Context, startTime, endTime *int, page, pageSize int) ([]model.RelayLog, error) {
	enabled, err := SettingGetBool(model.SettingKeyRelayLogKeepEnabled)
	if err != nil {
		return nil, err
	}
	hasTimeFilter := startTime != nil && endTime != nil
	seen := make(map[int64]struct{}, pageSize)

	// 获取缓存中符合条件的日志
	relayLogCacheLock.Lock()
	var cachedLogs []model.RelayLog
	for _, log := range relayLogCache {
		if hasTimeFilter {
			if log.Time >= int64(*startTime) && log.Time <= int64(*endTime) {
				cachedLogs = append(cachedLogs, log)
			}
		} else {
			cachedLogs = append(cachedLogs, log)
		}
	}
	relayLogCacheLock.Unlock()

	// 反转缓存日志顺序（原本新的在末尾，反转后新的在前面，方便分页）
	for i, j := 0, len(cachedLogs)-1; i < j; i, j = i+1, j-1 {
		cachedLogs[i], cachedLogs[j] = cachedLogs[j], cachedLogs[i]
	}

	cacheCount := len(cachedLogs)
	offset := (page - 1) * pageSize

	var result []model.RelayLog

	// 先从缓存中取（缓存是最新的日志）
	if offset < cacheCount {
		cacheEnd := offset + pageSize
		if cacheEnd > cacheCount {
			cacheEnd = cacheCount
		}
		result = append(result, cachedLogs[offset:cacheEnd]...)
		for _, item := range cachedLogs[offset:cacheEnd] {
			seen[item.ID] = struct{}{}
		}
	}

	// 如果启用了日志保存，缓存不够时从数据库补充
	if enabled {
		remaining := pageSize - len(result)
		dbOffset := 0
		if offset > cacheCount {
			dbOffset = offset - cacheCount
		}

		for remaining > 0 {
			query := db.GetDB().WithContext(ctx)
			if hasTimeFilter {
				query = query.Where("time >= ? AND time <= ?", *startTime, *endTime)
			}

			var dbLogs []model.RelayLog
			if err := query.Order("id DESC").Offset(dbOffset).Limit(remaining).Find(&dbLogs).Error; err != nil {
				return nil, err
			}
			if len(dbLogs) == 0 {
				break
			}

			dbOffset += len(dbLogs)
			appended := 0
			for _, item := range dbLogs {
				if _, ok := seen[item.ID]; ok {
					continue
				}
				result = append(result, item)
				seen[item.ID] = struct{}{}
				remaining--
				appended++
				if remaining == 0 {
					break
				}
			}
			if appended == 0 && len(dbLogs) < remaining {
				break
			}
		}
	}

	return result, nil
}

func RelayLogClear(ctx context.Context) error {
	relayLogCacheLock.Lock()
	relayLogCache = make([]model.RelayLog, 0, relayLogMaxSize)
	relayLogCacheLock.Unlock()
	return db.GetDB().WithContext(ctx).Where("1 = 1").Delete(&model.RelayLog{}).Error
}

// RelayLogGetByID 按 Snowflake ID 查询单条 RelayLog，优先查缓存，未命中时回源数据库。
func RelayLogGetByID(ctx context.Context, id int64) (model.RelayLog, error) {
	relayLogCacheLock.Lock()
	for _, entry := range relayLogCache {
		if entry.ID == id {
			result := entry
			relayLogCacheLock.Unlock()
			return result, nil
		}
	}
	relayLogCacheLock.Unlock()

	var entry model.RelayLog
	if err := db.GetDB().WithContext(ctx).Where("id = ?", id).First(&entry).Error; err != nil {
		return model.RelayLog{}, err
	}
	return entry, nil
}

// RelayLogExport returns logs in descending id order without page slicing.
// This is intended for explicit export use, so callers should always provide a sane limit.
func RelayLogExport(ctx context.Context, startTime, endTime *int, limit int) ([]model.RelayLog, error) {
	if limit <= 0 {
		limit = 2000
	}
	if limit > 10000 {
		limit = 10000
	}

	enabled, err := SettingGetBool(model.SettingKeyRelayLogKeepEnabled)
	if err != nil {
		return nil, err
	}
	hasTimeFilter := startTime != nil && endTime != nil

	result := make([]model.RelayLog, 0, limit)

	// Pull from cache first (newest first)
	relayLogCacheLock.Lock()
	for i := len(relayLogCache) - 1; i >= 0 && len(result) < limit; i-- {
		logItem := relayLogCache[i]
		if hasTimeFilter {
			if logItem.Time < int64(*startTime) || logItem.Time > int64(*endTime) {
				continue
			}
		}
		result = append(result, logItem)
	}
	relayLogCacheLock.Unlock()

	if !enabled || len(result) >= limit {
		return result, nil
	}

	query := db.GetDB().WithContext(ctx)
	if hasTimeFilter {
		query = query.Where("time >= ? AND time <= ?", *startTime, *endTime)
	}

	var dbLogs []model.RelayLog
	if err := query.Order("id DESC").Limit(limit).Find(&dbLogs).Error; err != nil {
		return nil, fmt.Errorf("failed to export relay logs: %w", err)
	}

	seen := make(map[int64]struct{}, len(result))
	for _, item := range result {
		seen[item.ID] = struct{}{}
	}
	for _, item := range dbLogs {
		if _, ok := seen[item.ID]; ok {
			continue
		}
		result = append(result, item)
		if len(result) >= limit {
			break
		}
	}

	return result, nil
}
