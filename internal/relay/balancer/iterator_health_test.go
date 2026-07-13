package balancer

import (
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

func TestHealthScore(t *testing.T) {
	allInfos := []healthInfo{
		{latency: 100, successRate: 0.9},
		{latency: 50, successRate: 0.5},
		{latency: 200, successRate: 0.8},
	}

	// score = successRate*0.6 + (1/latency)*0.4
	// A: 0.9*0.6 + (1/100)*0.4 = 0.54 + 0.004 = 0.544
	// B: 0.5*0.6 + (1/50)*0.4  = 0.30 + 0.008 = 0.308
	// C: 0.8*0.6 + (1/200)*0.4 = 0.48 + 0.002 = 0.482
	scoreA := healthScore(allInfos[0], allInfos)
	scoreB := healthScore(allInfos[1], allInfos)
	scoreC := healthScore(allInfos[2], allInfos)

	if scoreA <= scoreB || scoreA <= scoreC {
		t.Fatalf("A should have highest score: A=%f B=%f C=%f", scoreA, scoreB, scoreC)
	}
	if scoreC <= scoreB {
		t.Fatalf("C should outscore B: C=%f B=%f", scoreC, scoreB)
	}
}

func TestHealthScoreZeroLatencyFallback(t *testing.T) {
	// When latency is 0, healthScore should use max latency from allInfos
	allInfos := []healthInfo{
		{latency: 0, successRate: 1.0},
		{latency: 100, successRate: 0.5},
	}
	// For info[0]: latency=0 → fallback to max(0, 100)=100
	// score = 1.0*0.6 + (1/100)*0.4 = 0.6 + 0.004 = 0.604
	// For info[1]: latency=100
	// score = 0.5*0.6 + (1/100)*0.4 = 0.3 + 0.004 = 0.304
	score0 := healthScore(allInfos[0], allInfos)
	score1 := healthScore(allInfos[1], allInfos)
	if score0 <= score1 {
		t.Fatalf("zero-latency candidate with perfect success should win: %f vs %f", score0, score1)
	}
}

func TestHealthScoreAllZeroLatency(t *testing.T) {
	// When all latencies are 0, should not panic (uses latency=1)
	allInfos := []healthInfo{
		{latency: 0, successRate: 0.5},
		{latency: 0, successRate: 0.9},
	}
	score0 := healthScore(allInfos[0], allInfos)
	score1 := healthScore(allInfos[1], allInfos)
	// Both use latency=1: score = successRate*0.6 + (1/1)*0.4
	// score0 = 0.5*0.6 + 0.4 = 0.7
	// score1 = 0.9*0.6 + 0.4 = 0.94
	if score1 <= score0 {
		t.Fatalf("higher success rate should win when all latencies equal: %f vs %f", score1, score0)
	}
}

func TestChannelCircuitStateNoBreaker(t *testing.T) {
	ResetCircuitStateForTest()
	state := ChannelCircuitState(99901, "gpt-4o")
	if state != "closed" {
		t.Fatalf("expected closed for channel with no breaker, got %s", state)
	}
}

func TestChannelCircuitStateOpenAfterFailures(t *testing.T) {
	ResetCircuitStateForTest()
	// Trip the circuit breaker for channel 99902, keyID 1, model "gpt-4o"
	// Need enough failures to exceed threshold (default 5, but let's use the shim)
	for i := 0; i < 10; i++ {
		RecordFailure(99902, 1, "gpt-4o")
	}
	state := ChannelCircuitState(99902, "gpt-4o")
	if state != "open" {
		t.Fatalf("expected open after repeated failures, got %s", state)
	}
}

func TestChannelCircuitStateIgnoresOtherChannels(t *testing.T) {
	ResetCircuitStateForTest()
	for i := 0; i < 10; i++ {
		RecordFailure(99903, 1, "gpt-4o")
	}
	// Channel 99904 should be unaffected
	state := ChannelCircuitState(99904, "gpt-4o")
	if state != "closed" {
		t.Fatalf("expected closed for unaffected channel, got %s", state)
	}
}

func TestChannelCircuitStateIgnoresOtherModel(t *testing.T) {
	ResetCircuitStateForTest()
	for i := 0; i < 10; i++ {
		RecordFailure(99905, 1, "claude-3")
	}
	// Different model on same channel should be unaffected
	state := ChannelCircuitState(99905, "gpt-4o")
	if state != "closed" {
		t.Fatalf("expected closed for different model, got %s", state)
	}
}

func TestIteratorGoFallbackLeastLatency(t *testing.T) {
	resetStickySessions()
	ResetCircuitStateForTest()
	t.Setenv("OCTOPUS_RUST_BALANCER", "0")

	// Set up stats for three channels with different average latencies.
	// StatsChannelUpdate uses Add, so we call once with the desired totals.
	// latency = WaitTime / (RequestSuccess + RequestFailed)
	op.StatsChannelUpdate(99910, model.StatsMetrics{RequestSuccess: 10, RequestFailed: 0, WaitTime: 1000}) // avg 100ms
	op.StatsChannelUpdate(99911, model.StatsMetrics{RequestSuccess: 10, RequestFailed: 0, WaitTime: 2000}) // avg 200ms
	op.StatsChannelUpdate(99912, model.StatsMetrics{RequestSuccess: 10, RequestFailed: 0, WaitTime: 500})  // avg 50ms

	group := model.Group{
		ID:   901,
		Mode: model.GroupModeLeastLatency,
		Items: []model.GroupItem{
			{ChannelID: 99910, ModelName: "gpt-4o", Priority: 1},
			{ChannelID: 99911, ModelName: "gpt-4o", Priority: 2},
			{ChannelID: 99912, ModelName: "gpt-4o", Priority: 3},
		},
	}

	it := NewIterator(group, 0, "gpt-4o")
	want := []int{99912, 99910, 99911} // 50ms, 100ms, 200ms
	got := collectIteratorChannelIDs(it)
	if len(got) != len(want) {
		t.Fatalf("iterator length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("least_latency order = %v, want %v", got, want)
		}
	}
}

func TestIteratorGoFallbackHealthAware(t *testing.T) {
	resetStickySessions()
	ResetCircuitStateForTest()
	t.Setenv("OCTOPUS_RUST_BALANCER", "0")

	// Channel A: 90% success, 100ms avg latency
	//   score = 0.9*0.6 + (1/100)*0.4 = 0.544
	// Channel B: 50% success, 50ms avg latency
	//   score = 0.5*0.6 + (1/50)*0.4 = 0.308
	// Channel C: 80% success, 200ms avg latency
	//   score = 0.8*0.6 + (1/200)*0.4 = 0.482
	op.StatsChannelUpdate(99920, model.StatsMetrics{RequestSuccess: 9, RequestFailed: 1, WaitTime: 1000})  // 90%, 100ms
	op.StatsChannelUpdate(99921, model.StatsMetrics{RequestSuccess: 5, RequestFailed: 5, WaitTime: 500})   // 50%, 50ms
	op.StatsChannelUpdate(99922, model.StatsMetrics{RequestSuccess: 8, RequestFailed: 2, WaitTime: 2000})  // 80%, 200ms

	group := model.Group{
		ID:   902,
		Mode: model.GroupModeHealthAware,
		Items: []model.GroupItem{
			{ChannelID: 99920, ModelName: "gpt-4o", Priority: 1},
			{ChannelID: 99921, ModelName: "gpt-4o", Priority: 2},
			{ChannelID: 99922, ModelName: "gpt-4o", Priority: 3},
		},
	}

	it := NewIterator(group, 0, "gpt-4o")
	want := []int{99920, 99922, 99921} // A(0.544), C(0.482), B(0.308)
	got := collectIteratorChannelIDs(it)
	if len(got) != len(want) {
		t.Fatalf("iterator length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("health_aware order = %v, want %v", got, want)
		}
	}
}

func TestIteratorGoFallbackDefaultsForNewChannels(t *testing.T) {
	resetStickySessions()
	ResetCircuitStateForTest()
	t.Setenv("OCTOPUS_RUST_BALANCER", "0")

	// Channels with no stats data: all defaults (successRate=1.0, latency=0).
	// All candidates are equal → original priority order preserved.
	group := model.Group{
		ID:   903,
		Mode: model.GroupModeLeastLatency,
		Items: []model.GroupItem{
			{ChannelID: 99930, ModelName: "gpt-4o", Priority: 3},
			{ChannelID: 99931, ModelName: "gpt-4o", Priority: 1},
			{ChannelID: 99932, ModelName: "gpt-4o", Priority: 2},
		},
	}

	it := NewIterator(group, 0, "gpt-4o")
	want := []int{99931, 99932, 99930} // priority order
	got := collectIteratorChannelIDs(it)
	if len(got) != len(want) {
		t.Fatalf("iterator length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("default order = %v, want %v (priority preserved when no data)", got, want)
		}
	}
}

func TestIteratorGoFallbackCircuitOpenDeprioritized(t *testing.T) {
	resetStickySessions()
	ResetCircuitStateForTest()
	t.Setenv("OCTOPUS_RUST_BALANCER", "0")

	// Channel 99940 has an open circuit breaker.
	// In Go fallback, the circuit state is collected but applyGoFallback
	// doesn't explicitly skip open circuits — it uses latency/score.
	// However, the stats for the tripped channel would show low success rate.
	for i := 0; i < 10; i++ {
		RecordFailure(99940, 1, "gpt-4o")
	}

	// Give 99940 some stats showing failures (low success rate).
	op.StatsChannelUpdate(99940, model.StatsMetrics{RequestSuccess: 1, RequestFailed: 9, WaitTime: 1000})  // 10%, 100ms
	op.StatsChannelUpdate(99941, model.StatsMetrics{RequestSuccess: 10, RequestFailed: 0, WaitTime: 1000}) // 100%, 100ms

	group := model.Group{
		ID:   904,
		Mode: model.GroupModeHealthAware,
		Items: []model.GroupItem{
			{ChannelID: 99940, ModelName: "gpt-4o", Priority: 1},
			{ChannelID: 99941, ModelName: "gpt-4o", Priority: 2},
		},
	}

	it := NewIterator(group, 0, "gpt-4o")
	got := collectIteratorChannelIDs(it)
	if len(got) != 2 {
		t.Fatalf("iterator length = %d, want 2", len(got))
	}
	// Channel 99941 (100% success) should be first.
	if got[0] != 99941 {
		t.Fatalf("expected healthy channel first, got %d", got[0])
	}
}

func TestIteratorCircuitStatePassedToRustSelection(t *testing.T) {
	resetStickySessions()
	ResetCircuitStateForTest()

	// Trip circuit for channel 99950.
	for i := 0; i < 10; i++ {
		RecordFailure(99950, 1, "gpt-4o")
	}

	// Collect health info — circuit state should be "open".
	info := collectChannelHealth(99950, "gpt-4o")
	if info.circuitState != "open" {
		t.Fatalf("expected circuit state 'open' for tripped channel, got %s", info.circuitState)
	}
	if info.healthy {
		t.Fatalf("expected healthy=false for tripped channel")
	}

	// Channel 99951 should be closed.
	info2 := collectChannelHealth(99951, "gpt-4o")
	if info2.circuitState != "closed" {
		t.Fatalf("expected circuit state 'closed' for healthy channel, got %s", info2.circuitState)
	}
	if !info2.healthy {
		t.Fatalf("expected healthy=true for healthy channel")
	}
}
