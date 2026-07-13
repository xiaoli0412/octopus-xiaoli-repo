package op

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

const (
	defaultResponseCacheTTL        = 300 * time.Second
	defaultResponseCacheMaxEntries = 1000
)

// cacheEntry 存储单个缓存响应及其过期时间与最近访问时间（用于 LRU 淘汰）。
type cacheEntry struct {
	response   []byte
	expiry     time.Time
	lastAccess time.Time
}

// ResponseCache 是基于内存的 LRU + TTL 响应缓存。
// 仅用于非流式请求，不同 API Key 的相同请求可共享缓存（key 不包含 API Key ID）。
type ResponseCache struct {
	mu         sync.RWMutex
	entries    map[string]*cacheEntry
	maxEntries int
}

var (
	globalResponseCache     *ResponseCache
	globalResponseCacheOnce sync.Once
)

// GetResponseCache 返回全局 ResponseCache 单例。
func GetResponseCache() *ResponseCache {
	globalResponseCacheOnce.Do(func() {
		globalResponseCache = &ResponseCache{
			entries:    make(map[string]*cacheEntry),
			maxEntries: defaultResponseCacheMaxEntries,
		}
	})
	return globalResponseCache
}

// ComputeCacheKey 根据 model + messages + params 计算缓存的 SHA256 hex key。
// messages 通常是请求体中 messages 字段的规范 JSON；
// params 为生成参数（temperature、max_tokens 等），可为 nil。
func ComputeCacheKey(model string, messages json.RawMessage, params map[string]interface{}) string {
	payload := struct {
		Model    string                 `json:"model"`
		Messages json.RawMessage        `json:"messages"`
		Params   map[string]interface{} `json:"params,omitempty"`
	}{
		Model:    model,
		Messages: messages,
		Params:   params,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		// 序列化失败时退化为仅用 model + messages 原始字节
		data = append([]byte(model), messages...)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// Get 按 key 查找缓存。命中且未过期时返回响应字节并更新 lastAccess。
func (rc *ResponseCache) Get(key string) ([]byte, bool) {
	rc.mu.RLock()
	entry, ok := rc.entries[key]
	rc.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiry) {
		return nil, false
	}
	rc.mu.Lock()
	entry.lastAccess = time.Now()
	rc.mu.Unlock()
	// 返回副本，避免调用方修改缓存内容
	out := make([]byte, len(entry.response))
	copy(out, entry.response)
	return out, true
}

// Set 写入缓存条目。若超过 maxEntries，则淘汰 lastAccess 最旧的条目（LRU）。
func (rc *ResponseCache) Set(key string, response []byte, ttl time.Duration) {
	if ttl <= 0 {
		ttl = defaultResponseCacheTTL
	}
	respCopy := make([]byte, len(response))
	copy(respCopy, response)
	now := time.Now()

	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.entries[key] = &cacheEntry{
		response:   respCopy,
		expiry:     now.Add(ttl),
		lastAccess: now,
	}

	maxEntries := rc.maxEntries
	if maxEntries <= 0 {
		maxEntries = defaultResponseCacheMaxEntries
	}
	for len(rc.entries) > maxEntries {
		var oldestKey string
		var oldestTime time.Time
		first := true
		for k, e := range rc.entries {
			if first || e.lastAccess.Before(oldestTime) {
				oldestKey = k
				oldestTime = e.lastAccess
				first = false
			}
		}
		if oldestKey == "" {
			break
		}
		delete(rc.entries, oldestKey)
	}
}

// EvictExpired 遍历并删除所有已过期条目。
func (rc *ResponseCache) EvictExpired() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	now := time.Now()
	for k, e := range rc.entries {
		if now.After(e.expiry) {
			delete(rc.entries, k)
		}
	}
}

// SetMaxEntries 更新最大条目数。仅影响后续 Set 操作的淘汰判断。
func (rc *ResponseCache) SetMaxEntries(n int) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if n <= 0 {
		n = defaultResponseCacheMaxEntries
	}
	rc.maxEntries = n
}

// Len 返回当前缓存条目数（含可能已过期但尚未清理的条目）。
func (rc *ResponseCache) Len() int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return len(rc.entries)
}

// Clear 清空所有缓存条目。
func (rc *ResponseCache) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.entries = make(map[string]*cacheEntry)
}

// ResponseCacheTTL 从系统设置读取缓存 TTL，失败时返回默认值。
func ResponseCacheTTL() time.Duration {
	seconds, err := SettingGetInt(model.SettingKeyResponseCacheTTL)
	if err != nil || seconds <= 0 {
		return defaultResponseCacheTTL
	}
	return time.Duration(seconds) * time.Second
}

// ResponseCacheMaxEntries 从系统设置读取最大条目数，失败时返回默认值。
func ResponseCacheMaxEntries() int {
	n, err := SettingGetInt(model.SettingKeyResponseCacheMaxEntries)
	if err != nil || n <= 0 {
		return defaultResponseCacheMaxEntries
	}
	return n
}

// ResponseCacheEnabled 从系统设置读取全局缓存开关，失败时返回 false。
func ResponseCacheEnabled() bool {
	v, err := SettingGetBool(model.SettingKeyResponseCacheEnabled)
	if err != nil {
		return false
	}
	return v
}

// SyncResponseCacheConfig 根据系统设置同步缓存的最大条目数。
// 应在设置变更后或服务启动时调用。
func SyncResponseCacheConfig() {
	GetResponseCache().SetMaxEntries(ResponseCacheMaxEntries())
}
