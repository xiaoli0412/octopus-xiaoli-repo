package balancer

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/rustbridge"
)

type Iterator struct {
	candidates            []model.GroupItem
	baseCandidates        []model.GroupItem
	index                 int
	stickyChannelID       int
	modelName             string
	groupID               int
	groupMode             model.GroupMode
	strategy              string
	rustSelectionDisabled bool

	attemptMu sync.Mutex
	attempts  []model.ChannelAttempt
	count     int
}

func NewIterator(group model.Group, apiKeyID int, requestModel string) *Iterator {
	b := GetBalancer(group.Mode)
	base := b.Candidates(group.Items)
	return NewIteratorWithCandidates(group, apiKeyID, requestModel, base)
}

func NewIteratorWithCandidates(group model.Group, apiKeyID int, requestModel string, base []model.GroupItem) *Iterator {
	candidates := make([]model.GroupItem, len(base))
	copy(candidates, base)

	baseCandidates := make([]model.GroupItem, len(base))
	copy(baseCandidates, base)

	stickyChannelID := 0
	if group.SessionKeepTime > 0 {
		stickyTTL := time.Duration(group.SessionKeepTime) * time.Second
		if sticky := GetSticky(apiKeyID, requestModel, stickyTTL); sticky != nil {
			for i, item := range candidates {
				if item.ChannelID != sticky.ChannelID {
					continue
				}
				if i > 0 {
					stickyItem := candidates[i]
					copy(candidates[1:i+1], candidates[0:i])
					candidates[0] = stickyItem
				}
				stickyChannelID = sticky.ChannelID
				break
			}
		}
	}

	return &Iterator{
		candidates:      candidates,
		baseCandidates:  baseCandidates,
		index:           -1,
		stickyChannelID: stickyChannelID,
		modelName:       requestModel,
		groupID:         group.ID,
		groupMode:       group.Mode,
		strategy:        strategyForMode(group.Mode),
	}
}

func (it *Iterator) Reset() {
	it.index = -1
	it.candidates = make([]model.GroupItem, len(it.baseCandidates))
	copy(it.candidates, it.baseCandidates)
	if it.stickyChannelID > 0 {
		for i, item := range it.candidates {
			if item.ChannelID != it.stickyChannelID {
				continue
			}
			if i > 0 {
				stickyItem := it.candidates[i]
				copy(it.candidates[1:i+1], it.candidates[0:i])
				it.candidates[0] = stickyItem
			}
			return
		}
	}
}

func (it *Iterator) Next() bool {
	it.index++
	if it.index >= len(it.candidates) {
		return false
	}
	// Sticky session has already pinned the first candidate; do not override it.
	if it.index == 0 && it.stickyChannelID > 0 {
		return true
	}
	remaining := it.candidates[it.index:]
	if len(remaining) <= 1 {
		return true
	}
	if rustBalancerEnabled() && it.isStrategyMode() {
		it.applyRustSelection(remaining)
	} else if it.isHealthMode() {
		it.applyGoFallback(remaining)
	}
	return true
}

// healthInfo aggregates runtime health signals for a candidate channel.
type healthInfo struct {
	latency      int64  // milliseconds (primary: BaseUrl.Delay, fallback: stats avg)
	healthy      bool
	circuitState string // "closed", "open", "half-open"
	successRate  float64
	avgLatency   int64 // from StatsChannel, milliseconds
}

// collectHealthInfo aggregates BaseUrl.Delay, circuit breaker state, and
// StatsChannel success rate / average latency for the given candidates.
func (it *Iterator) collectHealthInfo(items []model.GroupItem) []healthInfo {
	infos := make([]healthInfo, len(items))
	for i, item := range items {
		modelName := item.ModelName
		if modelName == "" {
			modelName = it.modelName
		}
		infos[i] = collectChannelHealth(item.ChannelID, modelName)
	}
	return infos
}

func collectChannelHealth(channelID int, modelName string) healthInfo {
	info := healthInfo{
		healthy:      true,
		circuitState: "closed",
		successRate:  1.0, // default: assume healthy for new channels
	}

	// Circuit breaker state (channel-level, read-only)
	cs := ChannelCircuitState(channelID, modelName)
	info.circuitState = cs
	if cs == "open" {
		info.healthy = false
	}

	// BaseUrl delay (min across configured base URLs)
	if ch, err := op.ChannelGet(channelID, nil); err == nil && ch != nil {
		minDelay := 0
		for _, bu := range ch.BaseUrls {
			if bu.URL == "" {
				continue
			}
			if minDelay == 0 || bu.Delay < minDelay {
				minDelay = bu.Delay
			}
		}
		info.latency = int64(minDelay)
	}

	// Stats: success rate and average latency
	stats := op.StatsChannelGet(channelID)
	total := stats.RequestSuccess + stats.RequestFailed
	if total > 0 {
		info.successRate = float64(stats.RequestSuccess) / float64(total)
		info.avgLatency = stats.WaitTime / total
		// If no BaseUrl delay, use stats average latency
		if info.latency == 0 {
			info.latency = info.avgLatency
		}
	}

	return info
}

// applyRustSelection builds BalanceCandidate with health data and calls
// rustbridge.BalanceSelect to pick the best remaining candidate.
func (it *Iterator) applyRustSelection(remaining []model.GroupItem) {
	infos := it.collectHealthInfo(remaining)
	rustCands := make([]rustbridge.BalanceCandidate, len(remaining))
	for i, item := range remaining {
		rustCands[i] = rustbridge.BalanceCandidate{
			ID:           int64(item.ChannelID),
			Weight:       int64(item.Weight),
			Latency:      infos[i].latency,
			Priority:     int64(item.Priority),
			Healthy:      infos[i].healthy,
			CircuitState: infos[i].circuitState,
		}
	}
	if result, err := rustbridge.BalanceSelect(rustCands, it.strategy, 0); err == nil {
		for i, item := range remaining {
			if int64(item.ChannelID) == result.ID {
				if i != 0 {
					remaining[0], remaining[i] = remaining[i], remaining[0]
				}
				break
			}
		}
	}
}

// applyGoFallback reorders remaining candidates in pure Go when Rust
// selection is disabled. Used for LeastLatency and HealthAware modes.
func (it *Iterator) applyGoFallback(remaining []model.GroupItem) {
	infos := it.collectHealthInfo(remaining)
	switch it.groupMode {
	case model.GroupModeLeastLatency:
		bestIdx := 0
		bestLatency := infos[0].latency
		for i := 1; i < len(infos); i++ {
			if infos[i].latency < bestLatency {
				bestLatency = infos[i].latency
				bestIdx = i
			}
		}
		if bestIdx != 0 {
			remaining[0], remaining[bestIdx] = remaining[bestIdx], remaining[0]
		}
	case model.GroupModeHealthAware:
		bestIdx := 0
		bestScore := healthScore(infos[0], infos)
		for i := 1; i < len(infos); i++ {
			score := healthScore(infos[i], infos)
			if score > bestScore {
				bestScore = score
				bestIdx = i
			}
		}
		if bestIdx != 0 {
			remaining[0], remaining[bestIdx] = remaining[bestIdx], remaining[0]
		}
	}
}

// healthScore computes the HealthAware score:
// score = successRate * 0.6 + (1/latency) * 0.4
// Candidates with no latency data get a neutral score based on success rate.
func healthScore(info healthInfo, allInfos []healthInfo) float64 {
	latency := info.latency
	if latency <= 0 {
		// No latency data: use the max latency among candidates as fallback
		// so the channel isn't unfairly penalized or rewarded.
		for _, oi := range allInfos {
			if oi.latency > latency {
				latency = oi.latency
			}
		}
		if latency <= 0 {
			latency = 1
		}
	}
	return info.successRate*0.6 + (1.0/float64(latency))*0.4
}

func (it *Iterator) Item() model.GroupItem {
	return it.candidates[it.index]
}

func strategyForMode(mode model.GroupMode) string {
	switch mode {
	case model.GroupModeRoundRobin:
		return "round_robin"
	case model.GroupModeRandom:
		return "random"
	case model.GroupModeFailover:
		return "failover"
	case model.GroupModeWeighted:
		return "weighted"
	case model.GroupModeLeastLatency:
		return "least_latency"
	case model.GroupModeHealthAware:
		return "health_aware"
	default:
		return ""
	}
}

func rustBalancerEnabled() bool {
	v := os.Getenv("OCTOPUS_RUST_BALANCER")
	return v != "0" && v != "false" && v != "FALSE" && v != "False"
}

func (it *Iterator) isStrategyMode() bool {
	if it.rustSelectionDisabled {
		return false
	}
	switch it.groupMode {
	case model.GroupModeRoundRobin, model.GroupModeRandom, model.GroupModeFailover, model.GroupModeWeighted, model.GroupModeLeastLatency, model.GroupModeHealthAware:
		return true
	default:
		return false
	}
}

// isHealthMode reports whether the mode needs health-aware Go fallback
// when Rust selection is unavailable.
func (it *Iterator) isHealthMode() bool {
	switch it.groupMode {
	case model.GroupModeLeastLatency, model.GroupModeHealthAware:
		return true
	default:
		return false
	}
}

// DisableRustSelection disables the optional rustbridge.BalanceSelect path for
// this iterator. Callers that have already ordered candidates externally (e.g.
// dynamic routing recommendations) should use this to preserve their ordering.
func (it *Iterator) DisableRustSelection() {
	it.rustSelectionDisabled = true
}

func (it *Iterator) CandidateAt(index int) (model.GroupItem, bool) {
	if index < 0 || index >= len(it.candidates) {
		return model.GroupItem{}, false
	}
	return it.candidates[index], true
}

func (it *Iterator) IsSticky() bool {
	return it.stickyChannelID > 0 && it.index >= 0 && it.index < len(it.candidates) && it.candidates[it.index].ChannelID == it.stickyChannelID
}

func (it *Iterator) Len() int {
	return len(it.candidates)
}

func (it *Iterator) Index() int {
	return it.index
}

func (it *Iterator) GroupID() int {
	return it.groupID
}

func (it *Iterator) Skip(channelID, channelKeyID int, channelName, msg string) {
	it.recordAttempt(model.ChannelAttempt{
		ChannelID:    channelID,
		ChannelKeyID: channelKeyID,
		ChannelName:  channelName,
		ModelName:    it.currentModelName(),
		Status:       model.AttemptSkipped,
		Sticky:       it.isStickyChannel(channelID),
		Msg:          msg,
	})
}

func (it *Iterator) Record(channelID, channelKeyID int, channelName, modelName string, status model.AttemptStatus, statusCode int, duration time.Duration, msg string) {
	durationMs := 0
	if duration > 0 {
		durationMs = int(duration.Milliseconds())
	}
	if modelName == "" {
		modelName = it.currentModelName()
	}
	it.recordAttempt(model.ChannelAttempt{
		ChannelID:    channelID,
		ChannelKeyID: channelKeyID,
		ChannelName:  channelName,
		ModelName:    modelName,
		Status:       status,
		StatusCode:   statusCode,
		Duration:     durationMs,
		Sticky:       it.isStickyChannel(channelID),
		Msg:          msg,
	})
}

func (it *Iterator) SkipCircuitBreak(channelID, channelKeyID int, channelName string) bool {
	modelName := it.currentModelName()
	tripped, remaining := IsTripped(channelID, channelKeyID, modelName)
	if !tripped {
		return false
	}
	msg := "circuit breaker tripped"
	if remaining > 0 {
		msg = fmt.Sprintf("circuit breaker tripped, remaining cooldown: %ds", int(remaining.Seconds()))
	}
	it.recordAttempt(model.ChannelAttempt{
		ChannelID:    channelID,
		ChannelKeyID: channelKeyID,
		ChannelName:  channelName,
		ModelName:    modelName,
		Status:       model.AttemptCircuitBreak,
		Sticky:       it.isStickyChannel(channelID),
		Msg:          msg,
	})
	return true
}

func (it *Iterator) StartAttempt(channelID, channelKeyID int, channelName string) *AttemptSpan {
	return &AttemptSpan{
		attempt: model.ChannelAttempt{
			ChannelID:    channelID,
			ChannelKeyID: channelKeyID,
			ChannelName:  channelName,
			ModelName:    it.currentModelName(),
			Sticky:       it.isStickyChannel(channelID),
		},
		startTime: time.Now(),
		iter:      it,
	}
}

func (it *Iterator) Attempts() []model.ChannelAttempt {
	it.attemptMu.Lock()
	defer it.attemptMu.Unlock()
	out := make([]model.ChannelAttempt, len(it.attempts))
	copy(out, it.attempts)
	return out
}

type AttemptSpan struct {
	attempt   model.ChannelAttempt
	startTime time.Time
	iter      *Iterator
	ended     bool
}

func (s *AttemptSpan) End(status model.AttemptStatus, statusCode int, msg string) {
	if s.ended {
		return
	}
	s.ended = true
	s.attempt.Status = status
	s.attempt.StatusCode = statusCode
	s.attempt.Duration = int(time.Since(s.startTime).Milliseconds())
	s.attempt.Msg = msg
	s.iter.recordAttempt(s.attempt)
}

func (s *AttemptSpan) Duration() time.Duration {
	return time.Since(s.startTime)
}

func (it *Iterator) recordAttempt(attempt model.ChannelAttempt) {
	it.attemptMu.Lock()
	defer it.attemptMu.Unlock()
	it.count++
	attempt.AttemptNum = it.count
	it.attempts = append(it.attempts, attempt)
}

func (it *Iterator) currentModelName() string {
	if it.index >= 0 && it.index < len(it.candidates) {
		if modelName := it.candidates[it.index].ModelName; modelName != "" {
			return modelName
		}
	}
	return it.modelName
}

func (it *Iterator) isStickyChannel(channelID int) bool {
	if it.stickyChannelID <= 0 {
		return false
	}
	return it.stickyChannelID == channelID
}
