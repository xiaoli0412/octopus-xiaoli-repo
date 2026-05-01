package relay

import (
	"context"
	"fmt"
	"sync"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

type raceBudgetCounter struct {
	mu     sync.Mutex
	counts map[string]int
}

var globalRaceBudgetCounter = &raceBudgetCounter{counts: make(map[string]int)}

type raceBudgetRelease func()

func (c *raceBudgetCounter) acquire(key string, limit int) bool {
	if limit <= 0 {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.counts[key] >= limit {
		return false
	}
	c.counts[key]++
	return true
}

func (c *raceBudgetCounter) release(key string) {
	if key == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.counts[key] <= 1 {
		delete(c.counts, key)
		return
	}
	c.counts[key]--
}

func resetRaceBudgetState() {
	globalRaceBudgetCounter = &raceBudgetCounter{counts: make(map[string]int)}
}

func effectiveBudget(settingKey model.SettingKey, fallback int) int {
	if v, err := op.SettingGetInt(settingKey); err == nil && v > 0 {
		return v
	}
	return fallback
}

func acquireRaceBudgets(ctx context.Context, groupID, channelID, keyID int) (raceBudgetRelease, error) {
	type acquiredBudget struct {
		key string
	}

	acquired := make([]acquiredBudget, 0, 5)
	tryAcquire := func(key string, limit int, scope string) error {
		if !globalRaceBudgetCounter.acquire(key, limit) {
			return fmt.Errorf("race %s budget exhausted", scope)
		}
		acquired = append(acquired, acquiredBudget{key: key})
		return nil
	}
	releaseAll := func() {
		for i := len(acquired) - 1; i >= 0; i-- {
			globalRaceBudgetCounter.release(acquired[i].key)
		}
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := tryAcquire("global", effectiveBudget(model.SettingKeyRaceGlobalBudget, 64), "global"); err != nil {
		return nil, err
	}
	if groupID > 0 {
		if err := tryAcquire(fmt.Sprintf("group:%d", groupID), effectiveBudget(model.SettingKeyRaceGroupBudget, 8), "group"); err != nil {
			releaseAll()
			return nil, err
		}
	}
	if channelID > 0 {
		if err := tryAcquire(fmt.Sprintf("channel:%d", channelID), effectiveBudget(model.SettingKeyRaceChannelBudget, 4), "channel"); err != nil {
			releaseAll()
			return nil, err
		}
	}
	if keyID > 0 {
		if err := tryAcquire(fmt.Sprintf("key:%d", keyID), effectiveBudget(model.SettingKeyRaceKeyBudget, 2), "key"); err != nil {
			releaseAll()
			return nil, err
		}
	}
	if err := tryAcquire("probe", effectiveBudget(model.SettingKeyRaceProbeBudget, 16), "probe"); err != nil {
		releaseAll()
		return nil, err
	}

	return func() {
		releaseAll()
	}, nil
}
