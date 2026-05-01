package op

import (
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

const probeEventMaxSize = 256

var probeEventCache = make([]model.ProbeEvent, 0, probeEventMaxSize)
var probeEventCacheLock sync.RWMutex

type ProbeSummary struct {
	WindowSeconds       int     `json:"window_seconds"`
	TotalCount          int64   `json:"total_count"`
	SuccessCount        int64   `json:"success_count"`
	FailedCount         int64   `json:"failed_count"`
	LastAt              int64   `json:"last_at"`
	LastStatus          string  `json:"last_status,omitempty"`
	LastChannel         string  `json:"last_channel,omitempty"`
	LastModel           string  `json:"last_model,omitempty"`
	LastMessage         string  `json:"last_message,omitempty"`
	EstimatedInputCost  float64 `json:"estimated_input_cost"`
	EstimatedOutputCost float64 `json:"estimated_output_cost"`
	EstimatedTotalCost  float64 `json:"estimated_total_cost"`
	Basis               string  `json:"basis"`
}

func ProbeEventAdd(event model.ProbeEvent) {
	if event.Time == 0 {
		event.Time = time.Now().Unix()
	}

	probeEventCacheLock.Lock()
	defer probeEventCacheLock.Unlock()

	probeEventCache = append(probeEventCache, event)
	if len(probeEventCache) <= probeEventMaxSize {
		return
	}

	trimmed := make([]model.ProbeEvent, probeEventMaxSize)
	copy(trimmed, probeEventCache[len(probeEventCache)-probeEventMaxSize:])
	probeEventCache = trimmed
}

func ProbeSummaryGet(window time.Duration) ProbeSummary {
	if window <= 0 {
		window = 24 * time.Hour
	}

	cutoff := time.Now().Add(-window).Unix()
	summary := ProbeSummary{
		WindowSeconds: int(window / time.Second),
		Basis:         "runtime_race_probe_events_recent_window",
	}

	probeEventCacheLock.RLock()
	defer probeEventCacheLock.RUnlock()

	for _, event := range probeEventCache {
		if event.Time < cutoff {
			continue
		}

		summary.TotalCount++
		summary.EstimatedInputCost += event.EstimatedInputCost
		summary.EstimatedOutputCost += event.EstimatedOutputCost

		switch event.Status {
		case model.ProbeEventFailed:
			summary.FailedCount++
		case model.ProbeEventSuccess, model.ProbeEventSelected:
			summary.SuccessCount++
		}

		if event.Time >= summary.LastAt {
			summary.LastAt = event.Time
			summary.LastStatus = string(event.Status)
			summary.LastChannel = event.ChannelName
			summary.LastModel = event.ModelName
			summary.LastMessage = event.Message
		}
	}

	summary.EstimatedTotalCost = summary.EstimatedInputCost + summary.EstimatedOutputCost
	return summary
}
