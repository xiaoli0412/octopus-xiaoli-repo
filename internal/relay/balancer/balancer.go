package balancer

import (
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

// rngPool avoids global rand lock contention.
var rngPool = sync.Pool{New: func() any { return rand.New(rand.NewSource(time.Now().UnixNano())) }}

// Balancer 根据负载均衡模式选择通道
type Balancer interface {
	// Candidates 返回按策略排序的候选列表
	// 调用方在遍历候选列表时自行检查熔断状态
	Candidates(items []model.GroupItem) []model.GroupItem
}

// GetBalancer 根据模式返回对应的负载均衡器
func GetBalancer(mode model.GroupMode) Balancer {
	switch mode {
	case model.GroupModeRoundRobin:
		return &RoundRobin{}
	case model.GroupModeRandom:
		return &Random{}
	case model.GroupModeFailover:
		return &Failover{}
	case model.GroupModeWeighted:
		return &Weighted{}
	case model.GroupModeAIDynamic:
		return &RoundRobin{}
	default:
		return &RoundRobin{}
	}
}

// RoundRobin 严格顺序：每次从第 1 个开始的稳定顺序遍历。
type RoundRobin struct{}

func (b *RoundRobin) Candidates(items []model.GroupItem) []model.GroupItem {
	return sortByPriority(items)
}

// Random 随机：随机打乱所有 items
type Random struct{}

func (b *Random) Candidates(items []model.GroupItem) []model.GroupItem {
	n := len(items)
	if n == 0 {
		return nil
	}
	result := make([]model.GroupItem, n)
	copy(result, items)
	r := rngPool.Get().(*rand.Rand)
	defer rngPool.Put(r)
	r.Shuffle(n, func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	return result
}

// Failover 故障转移：按稳定优先级顺序排序。
type Failover struct{}

func (b *Failover) Candidates(items []model.GroupItem) []model.GroupItem {
	if len(items) == 0 {
		return nil
	}
	return sortByPriority(items)
}

// Weighted 加权分配：按严格权重序列展开，而不是加权随机。
type Weighted struct{}

func (b *Weighted) Candidates(items []model.GroupItem) []model.GroupItem {
	n := len(items)
	if n == 0 {
		return nil
	}

	itemsCopy := make([]model.GroupItem, n)
	copy(itemsCopy, items)

	sortGroupItems(itemsCopy)

	// Build strict cycle sequence by repeating each candidate exactly weight times.
	strictSeq := make([]model.GroupItem, 0)
	for _, item := range itemsCopy {
		w := item.Weight
		if w <= 0 {
			w = 1
		}
		for i := 0; i < w; i++ {
			strictSeq = append(strictSeq, item)
		}
	}
	return strictSeq
}

func sortByPriority(items []model.GroupItem) []model.GroupItem {
	sorted := make([]model.GroupItem, len(items))
	copy(sorted, items)
	sortGroupItems(sorted)
	return sorted
}

func sortGroupItems(items []model.GroupItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Priority != items[j].Priority {
			return items[i].Priority < items[j].Priority
		}
		if items[i].Weight != items[j].Weight {
			return items[i].Weight > items[j].Weight
		}
		if items[i].ChannelID != items[j].ChannelID {
			return items[i].ChannelID < items[j].ChannelID
		}
		if items[i].ModelName != items[j].ModelName {
			return items[i].ModelName < items[j].ModelName
		}
		return items[i].ID < items[j].ID
	})
}
