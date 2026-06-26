package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
)

// relayTokenUsageKey 用于在 request context 中传递本次请求的实际 token 消耗。
type relayTokenUsageKey struct{}

// SetRelayTokensInContext 将实际 token 消耗写入 request context。
func SetRelayTokensInContext(ctx context.Context, tokens int64) context.Context {
	if tokens <= 0 {
		return ctx
	}
	return context.WithValue(ctx, relayTokenUsageKey{}, tokens)
}

func totalTokensFromContext(ctx context.Context) int64 {
	v, ok := ctx.Value(relayTokenUsageKey{}).(int64)
	if !ok {
		return 0
	}
	return v
}

// apiKeyRateLimiter 提供基于进程内存的 API Key 限流。
// 当前实现适用于单实例部署；多实例场景需切换为 Redis 等共享存储。
type apiKeyRateLimiter struct {
	mu sync.RWMutex

	// rpmBuckets 每个 API Key 的令牌桶，用于限制每分钟请求数。
	rpmBuckets map[int]*tokenBucket

	// tpmWindows 每个 API Key 的滑动窗口计数器，用于限制每分钟 token 数。
	tpmWindows map[int]*slidingWindow

	// dailyCounters 每个 API Key 的日请求计数器。
	dailyCounters map[int]*dailyWindow
}

// tokenBucket 简单令牌桶，limit 为每分钟额度。
type tokenBucket struct {
	limit   int64
	tokens  float64
	lastRef time.Time
}

// slidingWindow 滑动窗口计数器，windowSize 固定为 1 分钟。
type slidingWindow struct {
	windowSize time.Duration
	hits       []windowHit
}

type windowHit struct {
	ts     time.Time
	amount int64
}

// dailyWindow 日请求计数窗口，按自然日重置。
type dailyWindow struct {
	day   int
	count int64
}

var globalAPIKeyRateLimiter = &apiKeyRateLimiter{
	rpmBuckets:    make(map[int]*tokenBucket),
	tpmWindows:    make(map[int]*slidingWindow),
	dailyCounters: make(map[int]*dailyWindow),
}

// AllowRequest 检查指定 API Key 是否允许发起本次请求。
// 若任意限流维度超限，返回 false。
func (rl *apiKeyRateLimiter) AllowRequest(apiKeyID int, limits model.APIKey) bool {
	if limits.RateLimitRPM <= 0 && limits.RateLimitTPM <= 0 && limits.RateLimitDaily <= 0 {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// RPM
	if limits.RateLimitRPM > 0 {
		bucket, ok := rl.rpmBuckets[apiKeyID]
		if !ok {
			bucket = &tokenBucket{limit: limits.RateLimitRPM, tokens: float64(limits.RateLimitRPM), lastRef: now}
			rl.rpmBuckets[apiKeyID] = bucket
		}
		elapsed := now.Sub(bucket.lastRef).Minutes()
		bucket.tokens += elapsed * float64(bucket.limit)
		if bucket.tokens > float64(bucket.limit) {
			bucket.tokens = float64(bucket.limit)
		}
		bucket.lastRef = now
		if bucket.tokens < 1 {
			return false
		}
		bucket.tokens--
	}

	// Daily
	if limits.RateLimitDaily > 0 {
		currentDay := now.YearDay() + now.Year()*1000
		counter, ok := rl.dailyCounters[apiKeyID]
		if !ok || counter.day != currentDay {
			counter = &dailyWindow{day: currentDay, count: 0}
			rl.dailyCounters[apiKeyID] = counter
		}
		if counter.count >= limits.RateLimitDaily {
			return false
		}
		counter.count++
	}

	return true
}

// RecordTokens 记录请求实际消耗的 token 数，用于 TPM 限流。
func (rl *apiKeyRateLimiter) RecordTokens(apiKeyID int, tokens int64, limits model.APIKey) bool {
	if limits.RateLimitTPM <= 0 || tokens <= 0 {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	window, ok := rl.tpmWindows[apiKeyID]
	if !ok {
		window = &slidingWindow{windowSize: time.Minute}
		rl.tpmWindows[apiKeyID] = window
	}

	window.hits = append(window.hits, windowHit{ts: now, amount: tokens})

	var total int64
	cutoff := now.Add(-window.windowSize)
	filtered := make([]windowHit, 0, len(window.hits))
	for _, hit := range window.hits {
		if hit.ts.After(cutoff) {
			filtered = append(filtered, hit)
			total += hit.amount
		}
	}
	window.hits = filtered

	return total <= limits.RateLimitTPM
}

// APIKeyRateLimitCheck 在请求进入 relay 前检查 RPM / 日配额，
// 返回是否被限流以及当前 API Key 配置（供后续 TPM 记录使用）。
func APIKeyRateLimitCheck(c *gin.Context) (apiKey *model.APIKey, allowed bool) {
	apiKeyObj, exists := c.Get("api_key")
	if !exists {
		return nil, true
	}
	key, ok := apiKeyObj.(*model.APIKey)
	if !ok || key == nil {
		return nil, true
	}
	if !globalAPIKeyRateLimiter.AllowRequest(key.ID, *key) {
		resp.Error(c, http.StatusTooManyRequests, "API key rate limit exceeded")
		c.Abort()
		return key, false
	}
	return key, true
}

// APIKeyRateLimitRecordTokens 在请求结束后记录实际 token 消耗，检查 TPM。
func APIKeyRateLimitRecordTokens(c *gin.Context, tokens int64) bool {
	apiKeyObj, exists := c.Get("api_key")
	if !exists {
		return true
	}
	key, ok := apiKeyObj.(*model.APIKey)
	if !ok || key == nil {
		return true
	}
	return globalAPIKeyRateLimiter.RecordTokens(key.ID, tokens, *key)
}
