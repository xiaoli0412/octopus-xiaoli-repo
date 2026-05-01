package task

import (
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

func resetTasksForTest() {
	tasksMu.Lock()
	defer tasksMu.Unlock()
	for name, entry := range tasks {
		stopTask(entry)
		delete(tasks, name)
	}
	taskRunnerWG = sync.WaitGroup{}
	taskRunStopCh = make(chan struct{})
	taskRunStopOnce = sync.Once{}
}

func taskIntervalForTest(name string) (time.Duration, bool) {
	tasksMu.RLock()
	defer tasksMu.RUnlock()
	entry, ok := tasks[name]
	if !ok {
		return 0, false
	}
	return entry.interval, true
}

func setupTaskTestDB(t *testing.T) {
	t.Helper()
	resetTasksForTest()
	if db.GetDB() != nil {
		_ = db.Close()
	}
	dbPath := filepath.Join(t.TempDir(), "octopus-task-test.db")
	if err := db.InitDB("sqlite", dbPath, false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() {
		if db.GetDB() != nil {
			_ = db.Close()
		}
	})
	if err := op.InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	setDynamicRoutingSummaryScanSummary(DynamicRoutingSummaryScanSummary{})
	t.Cleanup(resetTasksForTest)
}

func TestDynamicRoutingSummaryScanTaskBuildsDailySummary(t *testing.T) {
	setupTaskTestDB(t)

	channel := &model.Channel{
		Name:    "dynamic-task-channel",
		Enabled: true,
		Keys: []model.ChannelKey{
			{Enabled: true, ChannelKey: "free-key", SourceType: "public/free"},
			{Enabled: true, ChannelKey: "paid-key", SourceType: "paid"},
			{Enabled: true, ChannelKey: "private-key", SourceType: "private/internal"},
			{Enabled: true, ChannelKey: "unknown-key", SourceType: ""},
		},
	}
	if err := op.ChannelCreate(channel, t.Context()); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	group := &model.Group{Name: "dynamic-task-group", Mode: model.GroupModeFailover}
	if err := op.GroupCreate(group, t.Context()); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}

	DynamicRoutingSummaryScanTask()
	summary := GetDynamicRoutingSummaryScanSummary()

	if summary.LastStatus != "ok" {
		t.Fatalf("LastStatus = %q, want ok", summary.LastStatus)
	}
	if summary.ChannelCount != 1 || summary.EnabledChannels != 1 {
		t.Fatalf("channel summary = %#v, want 1 channel and 1 enabled channel", summary)
	}
	if summary.GroupCount != 1 || summary.FailoverGroups != 1 {
		t.Fatalf("group summary = %#v, want 1 group and 1 failover group", summary)
	}
	if summary.FreePublicKeys != 1 || summary.PaidMeteredKeys != 1 || summary.PrivateInnerKeys != 1 || summary.UnknownKeys != 1 {
		t.Fatalf("key summary = %#v, want one key in each category", summary)
	}
	if summary.LastRunAt.IsZero() || summary.LastSuccessAt.IsZero() {
		t.Fatalf("summary timestamps should be set: %#v", summary)
	}
	if time.Since(summary.LastRunAt) > time.Minute {
		t.Fatalf("LastRunAt = %v, want recent timestamp", summary.LastRunAt)
	}
	if summary.Basis == "" {
		t.Fatalf("Basis should be populated")
	}
	if summary.CurrentMode != "hybrid" {
		t.Fatalf("CurrentMode = %q, want hybrid", summary.CurrentMode)
	}
	if summary.EffectiveMode == "" || summary.Decision == "" {
		t.Fatalf("effective mode and decision should be populated: %#v", summary)
	}
}

func TestDynamicRoutingSummaryScanTaskMarksDisabledStateAsSkipped(t *testing.T) {
	setupTaskTestDB(t)

	if err := op.SettingSetString(model.SettingKeyDynamicRoutingHealthEnabled, "false"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}

	DynamicRoutingSummaryScanTask()
	summary := GetDynamicRoutingSummaryScanSummary()

	if summary.LastStatus != "skipped" {
		t.Fatalf("LastStatus = %q, want skipped", summary.LastStatus)
	}
	if summary.LastMessage != dynamicRoutingSummaryMessageHealthDisabledScanSkipped {
		t.Fatalf("LastMessage = %q, want disabled message", summary.LastMessage)
	}
	if !summary.LastSuccessAt.Equal(summary.LastRunAt) {
		t.Fatalf("LastSuccessAt = %v, want same as LastRunAt = %v", summary.LastSuccessAt, summary.LastRunAt)
	}
}

func TestInitRegistersStatsSaveTaskWithSettingKeyName(t *testing.T) {
	setupTaskTestDB(t)

	Init()

	interval, ok := taskIntervalForTest(string(model.SettingKeyStatsSaveInterval))
	if !ok {
		t.Fatalf("task %q not registered", model.SettingKeyStatsSaveInterval)
	}
	if interval != 10*time.Minute {
		t.Fatalf("interval = %v, want %v", interval, 10*time.Minute)
	}
}

func TestInitContinuesWhenOneSettingLookupFails(t *testing.T) {
	setupTaskTestDB(t)

	initTasks(func(key model.SettingKey) (int, error) {
		switch key {
		case model.SettingKeyModelInfoUpdateInterval:
			return 0, fmt.Errorf("boom")
		case model.SettingKeySyncLLMInterval:
			return 12, nil
		case model.SettingKeyStatsSaveInterval:
			return 15, nil
		default:
			return 0, fmt.Errorf("unexpected key %s", key)
		}
	})

	if _, ok := taskIntervalForTest(string(model.SettingKeySyncLLMInterval)); !ok {
		t.Fatalf("sync LLM task should still be registered")
	}
	if _, ok := taskIntervalForTest(string(model.SettingKeyStatsSaveInterval)); !ok {
		t.Fatalf("stats save task should still be registered")
	}
	if _, ok := taskIntervalForTest(string(model.SettingKeyModelInfoUpdateInterval)); ok {
		t.Fatalf("price update task should be skipped when its setting lookup fails")
	}
}

func TestUpdateKeepsLatestPendingIntervalBeforeRun(t *testing.T) {
	resetTasksForTest()
	t.Cleanup(resetTasksForTest)

	Register("update-test", time.Minute, false, func() {})
	Update("update-test", 2*time.Minute)
	Update("update-test", 3*time.Minute)

	tasksMu.RLock()
	entry, ok := tasks["update-test"]
	tasksMu.RUnlock()
	if !ok || entry == nil {
		t.Fatalf("task update-test not registered")
	}

	select {
	case got := <-entry.updateCh:
		if got != 3*time.Minute {
			t.Fatalf("pending interval = %v, want %v", got, 3*time.Minute)
		}
	default:
		t.Fatalf("expected pending interval update")
	}
}

func TestStartTaskExecutionRecoversFromPanicAndAllowsFutureRuns(t *testing.T) {
	resetTasksForTest()
	t.Cleanup(resetTasksForTest)

	started := make(chan struct{}, 2)
	done := make(chan struct{}, 1)
	panicOnce := true
	entry := &taskEntry{
		name: "panic-recovery-test",
		fn: func() {
			started <- struct{}{}
			if panicOnce {
				panicOnce = false
				panic("boom")
			}
			done <- struct{}{}
		},
	}

	if !startTaskExecution(entry) {
		t.Fatal("first execution should start")
	}

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first execution did not start in time")
	}

	deadline := time.Now().Add(time.Second)
	for {
		if startTaskExecution(entry) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("task should be runnable again after panic recovery")
		}
		time.Sleep(10 * time.Millisecond)
	}

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("second execution did not start in time")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("second execution did not finish in time")
	}
}

func TestStartTaskExecutionSkipsOverlappingRuns(t *testing.T) {
	resetTasksForTest()
	t.Cleanup(resetTasksForTest)

	started := make(chan struct{}, 1)
	release := make(chan struct{})
	done := make(chan struct{})
	entry := &taskEntry{
		name: "overlap-test",
		fn: func() {
			started <- struct{}{}
			defer close(done)
			<-release
		},
	}

	if !startTaskExecution(entry) {
		t.Fatal("first execution should start")
	}

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first execution did not start in time")
	}

	if startTaskExecution(entry) {
		t.Fatal("second execution should be skipped while the first is running")
	}

	close(release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("first execution did not finish in time")
	}
}

func TestGetLastSyncModelsTimeRoundTrip(t *testing.T) {
	want := time.Date(2026, time.April, 23, 4, 0, 0, 123456789, time.FixedZone("UTC+8", 8*60*60))
	setLastSyncModelsTime(want)
	t.Cleanup(func() {
		setLastSyncModelsTime(time.Now())
	})

	got := GetLastSyncModelsTime()
	if !got.Equal(want.UTC()) {
		t.Fatalf("GetLastSyncModelsTime() = %v, want %v", got, want.UTC())
	}
	if got.Location() != time.UTC {
		t.Fatalf("GetLastSyncModelsTime() location = %v, want UTC", got.Location())
	}
	if got.Nanosecond() != want.UTC().Nanosecond() {
		t.Fatalf("GetLastSyncModelsTime() nanosecond = %d, want %d", got.Nanosecond(), want.UTC().Nanosecond())
	}
	if got.IsZero() {
		t.Fatal("GetLastSyncModelsTime() should not be zero")
	}
}

func TestStopAllStopsRunningTaskLoopAndWaitsForExecution(t *testing.T) {
	resetTasksForTest()
	t.Cleanup(resetTasksForTest)

	started := make(chan struct{}, 1)
	release := make(chan struct{})
	var releaseOnce sync.Once
	var runs int32

	Register("stop-all-test", 10*time.Millisecond, true, func() {
		if atomic.AddInt32(&runs, 1) != 1 {
			return
		}
		started <- struct{}{}
		<-release
	})

	runDone := make(chan struct{})
	go func() {
		defer close(runDone)
		RUN()
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		releaseOnce.Do(func() { close(release) })
		t.Fatal("task execution did not start in time")
	}

	stopDone := make(chan error, 1)
	go func() {
		stopDone <- StopAll()
	}()

	select {
	case err := <-stopDone:
		releaseOnce.Do(func() { close(release) })
		if err != nil {
			t.Fatalf("StopAll() error = %v", err)
		}
		if got := atomic.LoadInt32(&runs); got != 1 {
			t.Fatalf("runs = %d, want 1", got)
		}
		<-runDone
	case <-time.After(200 * time.Millisecond):
	}

	releaseOnce.Do(func() { close(release) })

	select {
	case err := <-stopDone:
		if err != nil {
			t.Fatalf("StopAll() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("StopAll() did not return after task release")
	}

	select {
	case <-runDone:
	case <-time.After(time.Second):
		t.Fatal("RUN() did not exit after StopAll()")
	}
}
