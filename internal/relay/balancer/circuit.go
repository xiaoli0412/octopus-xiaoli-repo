package balancer

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

// CircuitState 熔断器状态
type CircuitState int

const (
	StateClosed   CircuitState = iota // 正常通行
	StateOpen                         // 熔断中，拒绝所有请求
	StateHalfOpen                     // 半开，仅允许单个试探请求
)

// circuitEntry 单个熔断器条目
type circuitEntry struct {
	State               CircuitState
	ConsecutiveFailures int64
	LastFailureTime     time.Time
	TripCount           int // 累计熔断触发次数（用于指数退避）
	mu                  sync.Mutex
}

type CircuitSummary struct {
	TrackedCount            int    `json:"tracked_count"`
	OpenCount               int    `json:"open_count"`
	HalfOpenCount           int    `json:"half_open_count"`
	ClosedCount             int    `json:"closed_count"`
	MaxRemainingCooldownSec int    `json:"max_remaining_cooldown_sec"`
	Basis                   string `json:"basis"`
}

// 全局熔断器存储
var globalBreaker sync.Map // key: string -> value: *circuitEntry

// circuitKey 生成熔断器键：channelID:channelKeyID:modelName
func circuitKey(channelID, keyID int, modelName string) string {
	return fmt.Sprintf("%d:%d:%s", channelID, keyID, modelName)
}

// getOrCreateEntry 获取或创建熔断器条目
func getOrCreateEntry(key string) *circuitEntry {
	if v, ok := globalBreaker.Load(key); ok {
		return v.(*circuitEntry)
	}
	entry := &circuitEntry{State: StateClosed}
	actual, _ := globalBreaker.LoadOrStore(key, entry)
	return actual.(*circuitEntry)
}

// getThreshold 获取熔断阈值配置
var RelayEffectiveCircuitThresholdShim = func(keyID int, modelName string) int64 {
	return 0
}

func getThreshold() int64 {
	v, err := op.SettingGetInt(model.SettingKeyCircuitBreakerThreshold)
	if err != nil || v <= 0 {
		return 5
	}
	return int64(v)
}

func effectiveThreshold(keyID int, modelName string) int64 {
	threshold := getThreshold()
	if override := RelayEffectiveCircuitThresholdShim(keyID, modelName); override > 0 {
		threshold = override
	}
	if threshold < 2 {
		threshold = 2
	}
	return threshold
}

// GetCooldown 获取当前冷却时间（带指数退避）
func GetCooldown(tripCount int) time.Duration {
	base, err := op.SettingGetInt(model.SettingKeyCircuitBreakerCooldown)
	if err != nil || base <= 0 {
		base = 60
	}
	maxCooldown, err := op.SettingGetInt(model.SettingKeyCircuitBreakerMaxCooldown)
	if err != nil || maxCooldown <= 0 {
		maxCooldown = 600
	}

	// 指数退避：baseCooldown * 2^(tripCount-1)
	cooldown := base
	if tripCount > 1 {
		shift := tripCount - 1
		if shift > 20 { // 防止溢出
			shift = 20
		}
		cooldown = base << shift
	}
	if cooldown > maxCooldown {
		cooldown = maxCooldown
	}

	return time.Duration(cooldown) * time.Second
}

// IsTripped 检查通道是否处于熔断状态
// 返回 tripped=true 表示该通道应被跳过，remaining 为剩余冷却时间
func IsTripped(channelID, keyID int, modelName string) (tripped bool, remaining time.Duration) {
	key := circuitKey(channelID, keyID, modelName)
	v, ok := globalBreaker.Load(key)
	if !ok {
		return false, 0 // 无记录，视为 Closed
	}
	entry := v.(*circuitEntry)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	switch entry.State {
	case StateClosed:
		return false, 0

	case StateOpen:
		cooldown := GetCooldown(entry.TripCount)
		elapsed := time.Since(entry.LastFailureTime)
		if elapsed >= cooldown {
			entry.State = StateHalfOpen
			log.Infof("circuit breaker [%s] Open -> HalfOpen (cooldown %v elapsed)", key, cooldown)
			return false, 0
		}
		// 仍在冷却中
		return true, cooldown - elapsed

	case StateHalfOpen:
		// 已有试探请求在进行中，拒绝其他请求
		return true, 0

	default:
		return false, 0
	}
}

// RecordSuccess 记录成功，重置熔断器状态
func RecordSuccess(channelID, keyID int, modelName string) {
	key := circuitKey(channelID, keyID, modelName)
	v, ok := globalBreaker.Load(key)
	if !ok {
		return
	}
	entry := v.(*circuitEntry)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if entry.State == StateHalfOpen {
		log.Infof("circuit breaker [%s] HalfOpen -> Closed (probe succeeded)", key)
	}

	// 重置全部状态
	entry.State = StateClosed
	entry.ConsecutiveFailures = 0
	entry.TripCount = 0
}

// RecordFailure 记录失败，可能触发熔断
func RecordFailure(channelID, keyID int, modelName string) {
	key := circuitKey(channelID, keyID, modelName)
	entry := getOrCreateEntry(key)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	entry.LastFailureTime = time.Now()

	switch entry.State {
	case StateClosed:
		entry.ConsecutiveFailures++
		threshold := effectiveThreshold(keyID, modelName)
		if entry.ConsecutiveFailures >= threshold {
			entry.State = StateOpen
			entry.TripCount++
			log.Warnf("circuit breaker [%s] Closed -> Open (failures=%d >= threshold=%d, tripCount=%d, cooldown=%v)",
				key, entry.ConsecutiveFailures, threshold, entry.TripCount, GetCooldown(entry.TripCount))
		}

	case StateHalfOpen:
		// 试探失败，重新进入 Open 状态，TripCount 递增（冷却时间翻倍）
		entry.State = StateOpen
		entry.TripCount++
		entry.ConsecutiveFailures = 0 // 重新开始计数
		log.Warnf("circuit breaker [%s] HalfOpen -> Open (probe failed, tripCount=%d, cooldown=%v)",
			key, entry.TripCount, GetCooldown(entry.TripCount))

	case StateOpen:
		// 理论上不应该在 Open 状态下接收到失败记录（请求应被拒绝），
		// 但为安全起见仍更新失败时间
	}
}

func ResetCircuitStateForTest() {
	globalBreaker = sync.Map{}
}

// ChannelCircuitState returns the worst circuit breaker state for a channel
// across all keys for the given model. Returns "closed", "open", or "half-open".
// This is a read-only inspection (unlike IsTripped which may transition Open→HalfOpen).
func ChannelCircuitState(channelID int, modelName string) string {
	chIDStr := fmt.Sprintf("%d", channelID)
	worst := StateClosed
	globalBreaker.Range(func(key, value any) bool {
		keyStr, ok := key.(string)
		if !ok {
			return true
		}
		parts := strings.SplitN(keyStr, ":", 3)
		if len(parts) < 3 || parts[0] != chIDStr {
			return true
		}
		if modelName != "" && parts[2] != modelName {
			return true
		}
		entry, ok := value.(*circuitEntry)
		if !ok || entry == nil {
			return true
		}
		entry.mu.Lock()
		state := entry.State
		tripCount := entry.TripCount
		lastFailure := entry.LastFailureTime
		entry.mu.Unlock()

		effectiveState := state
		if state == StateOpen {
			cooldown := GetCooldown(tripCount)
			if time.Since(lastFailure) >= cooldown {
				effectiveState = StateHalfOpen
			}
		}
		if effectiveState == StateOpen {
			worst = StateOpen
			return false // can stop scanning, worst found
		}
		if effectiveState == StateHalfOpen && worst != StateOpen {
			worst = StateHalfOpen
		}
		return true
	})
	switch worst {
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "closed"
	}
}

// ResetCircuitBreakersForChannel resets all circuit breakers associated with
// the given channel ID. It returns the number of breakers that were reset
// from a non-closed state. This is used by the proactive health check task
// to auto-recover channels that have been probed as healthy.
func ResetCircuitBreakersForChannel(channelID int) int {
	prefix := fmt.Sprintf("%d:", channelID)
	reset := 0
	globalBreaker.Range(func(key, value any) bool {
		keyStr, ok := key.(string)
		if !ok || !strings.HasPrefix(keyStr, prefix) {
			return true
		}
		entry, ok := value.(*circuitEntry)
		if !ok || entry == nil {
			return true
		}
		entry.mu.Lock()
		if entry.State != StateClosed {
			log.Infof("circuit breaker [%s] auto-recovered to Closed (health check probe succeeded)", keyStr)
			entry.State = StateClosed
			entry.ConsecutiveFailures = 0
			entry.TripCount = 0
			reset++
		}
		entry.mu.Unlock()
		return true
	})
	return reset
}

func init() {
	op.SetCircuitBreakerResetForChannel(ResetCircuitBreakersForChannel)
}

func SnapshotSummary(now time.Time) CircuitSummary {
	if now.IsZero() {
		now = time.Now()
	}

	summary := CircuitSummary{Basis: "runtime_breaker_snapshot"}
	globalBreaker.Range(func(_, value any) bool {
		entry, ok := value.(*circuitEntry)
		if !ok || entry == nil {
			return true
		}

		summary.TrackedCount++

		entry.mu.Lock()
		state := entry.State
		tripCount := entry.TripCount
		lastFailureTime := entry.LastFailureTime
		entry.mu.Unlock()

		switch state {
		case StateOpen:
			summary.OpenCount++
			if !lastFailureTime.IsZero() {
				remaining := GetCooldown(tripCount) - now.Sub(lastFailureTime)
				if remaining > 0 {
					remainingSec := int((remaining + time.Second - 1) / time.Second)
					if remainingSec > summary.MaxRemainingCooldownSec {
						summary.MaxRemainingCooldownSec = remainingSec
					}
				}
			}
		case StateHalfOpen:
			summary.HalfOpenCount++
		default:
			summary.ClosedCount++
		}

		return true
	})

	return summary
}
