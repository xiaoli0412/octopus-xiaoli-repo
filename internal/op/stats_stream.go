package op

import (
	"sync"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

// StatsSnapshot is the aggregated payload pushed to SSE subscribers after a
// relay completes. It mirrors the data served by the REST stats endpoints so
// the frontend can update its TanStack Query caches directly from the event.
type StatsSnapshot struct {
	Total  model.StatsTotal  `json:"total"`
	Today  model.StatsDaily  `json:"today"`
	Hourly model.StatsHourly `json:"hourly"`
}

var statsSubscribers = make(map[chan StatsSnapshot]struct{})
var statsSubscribersLock sync.RWMutex

// StatsSubscribe returns a buffered channel that receives a StatsSnapshot every
// time a relay completes. Callers must call StatsUnsubscribe to release the
// channel when the SSE connection closes.
func StatsSubscribe() chan StatsSnapshot {
	ch := make(chan StatsSnapshot, 16)
	statsSubscribersLock.Lock()
	statsSubscribers[ch] = struct{}{}
	statsSubscribersLock.Unlock()
	return ch
}

// StatsUnsubscribe removes the channel from the subscriber set and closes it.
func StatsUnsubscribe(ch chan StatsSnapshot) {
	statsSubscribersLock.Lock()
	delete(statsSubscribers, ch)
	statsSubscribersLock.Unlock()
	close(ch)
}

// PublishStatsSnapshot fans out a snapshot to all subscribers. It is
// non-blocking: when a subscriber's buffer is full the event is dropped so a
// slow consumer cannot stall the relay hot path.
func PublishStatsSnapshot(snapshot StatsSnapshot) {
	statsSubscribersLock.RLock()
	defer statsSubscribersLock.RUnlock()

	for ch := range statsSubscribers {
		select {
		case ch <- snapshot:
		default:
		}
	}
}

// StatsSnapshotBuild collects the current in-memory stats caches into a
// snapshot suitable for SSE publishing. It performs cheap RLock reads and is
// safe to call from the relay completion path.
func StatsSnapshotBuild() StatsSnapshot {
	total := StatsTotalGet()
	today := StatsTodayGet()

	nowHour := -1
	hourlyAll := StatsHourlyGet()
	if len(hourlyAll) > 0 {
		nowHour = len(hourlyAll) - 1
	}
	var hourly model.StatsHourly
	if nowHour >= 0 && nowHour < len(hourlyAll) {
		hourly = hourlyAll[nowHour]
	}
	return StatsSnapshot{
		Total:  total,
		Today:  today,
		Hourly: hourly,
	}
}
