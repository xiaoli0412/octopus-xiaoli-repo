package task

import (
	"fmt"
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

type taskEntry struct {
	name       string
	interval   time.Duration
	fn         func()
	runOnStart bool
	execMu     sync.Mutex
	execWG     sync.WaitGroup
	ticker     *time.Ticker
	stopCh     chan struct{}
	stopOnce   sync.Once
	updateCh   chan time.Duration
}

var (
	tasks           = make(map[string]*taskEntry)
	tasksMu         sync.RWMutex
	taskRunnerWG    sync.WaitGroup
	taskRunStopCh   = make(chan struct{})
	taskRunStopOnce sync.Once
)

const taskStopTimeout = 5 * time.Second

// Register 注册一个定时任务
// runOnStart: 是否在启动时立即执行一次
func Register(name string, interval time.Duration, runOnStart bool, fn func()) {
	if interval <= 0 {
		log.Debugf("task %s not registered: interval is 0", name)
		return
	}

	tasksMu.Lock()
	defer tasksMu.Unlock()

	if _, exists := tasks[name]; exists {
		log.Warnf("task %s already registered, skipping", name)
		return
	}

	tasks[name] = &taskEntry{
		name:       name,
		interval:   interval,
		fn:         fn,
		runOnStart: runOnStart,
		execMu:     sync.Mutex{},
		stopCh:     make(chan struct{}),
		updateCh:   make(chan time.Duration, 1),
	}
	log.Debugf("task %s registered with interval %v, runOnStart: %v", name, interval, runOnStart)
}

// Update 更新任务的执行间隔
// 当 interval 为 0 时，删除任务
func Update(name string, interval time.Duration) {
	tasksMu.Lock()
	entry, exists := tasks[name]
	if !exists {
		tasksMu.Unlock()
		log.Warnf("task %s not found", name)
		return
	}

	if interval <= 0 {
		delete(tasks, name)
		tasksMu.Unlock()
		stopTask(entry)
		log.Infof("task %s removed: interval is 0", name)
		return
	}
	tasksMu.Unlock()

	updated := false
	select {
	case entry.updateCh <- interval:
		updated = true
	default:
		select {
		case <-entry.updateCh:
		default:
		}
		select {
		case entry.updateCh <- interval:
			updated = true
		default:
		}
	}

	if updated {
		log.Infof("task %s interval updated to %v", name, interval)
		return
	}
	log.Warnf("task %s interval update could not be queued", name)
}

// RUN 启动所有注册的任务
func RUN() {
	tasksMu.RLock()
	entries := make([]*taskEntry, 0, len(tasks))
	for _, entry := range tasks {
		entries = append(entries, entry)
	}
	tasksMu.RUnlock()

	for _, entry := range entries {
		taskRunnerWG.Add(1)
		go runTask(entry)
	}

	<-taskRunStopCh
}

func runTask(entry *taskEntry) {
	defer taskRunnerWG.Done()
	// 根据配置决定是否在启动时立即执行
	if entry.runOnStart {
		startTaskExecution(entry)
	}

	entry.ticker = time.NewTicker(entry.interval)
	defer entry.ticker.Stop()

	for {
		select {
		case <-entry.ticker.C:
			startTaskExecution(entry)
		case newInterval := <-entry.updateCh:
			entry.ticker.Stop()
			entry.interval = newInterval
			entry.ticker = time.NewTicker(newInterval)
		case <-entry.stopCh:
			return
		case <-taskRunStopCh:
			return
		}
	}
}

func startTaskExecution(entry *taskEntry) bool {
	if entry == nil || entry.fn == nil {
		return false
	}
	if !entry.execMu.TryLock() {
		log.Debugf("task %s skipped overlapping execution", entry.name)
		return false
	}
	entry.execWG.Add(1)
	go func() {
		defer entry.execWG.Done()
		defer entry.execMu.Unlock()
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Errorf("task %s panicked: %v", entry.name, recovered)
			}
		}()
		entry.fn()
	}()
	return true
}

func StopAll() error {
	taskRunStopOnce.Do(func() {
		close(taskRunStopCh)
	})

	tasksMu.RLock()
	entries := make([]*taskEntry, 0, len(tasks))
	for _, entry := range tasks {
		entries = append(entries, entry)
	}
	tasksMu.RUnlock()

	for _, entry := range entries {
		stopTask(entry)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		taskRunnerWG.Wait()
		for _, entry := range entries {
			entry.execWG.Wait()
		}
	}()

	select {
	case <-done:
		return nil
	case <-time.After(taskStopTimeout):
		return fmt.Errorf("timed out waiting for background tasks to stop")
	}
}

func stopTask(entry *taskEntry) {
	if entry == nil {
		return
	}
	entry.stopOnce.Do(func() {
		close(entry.stopCh)
	})
}
