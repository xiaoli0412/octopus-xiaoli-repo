package balancer

import (
	"sync"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func resetStickySessions() {
	globalSession = sync.Map{}
}

func TestNewIteratorMovesStickyCandidateToFront(t *testing.T) {
	resetStickySessions()
	SetSticky(1001, "gpt-4o", 22, 220)

	group := model.Group{
		ID:              9,
		Mode:            model.GroupModeRoundRobin,
		SessionKeepTime: 60,
		Items: []model.GroupItem{
			{ChannelID: 11, ModelName: "gpt-4o", Priority: 1},
			{ChannelID: 22, ModelName: "gpt-4o", Priority: 2},
			{ChannelID: 33, ModelName: "gpt-4o", Priority: 3},
		},
	}

	it := NewIterator(group, 1001, "gpt-4o")
	if it.Len() != 3 {
		t.Fatalf("iterator length = %d, want 3", it.Len())
	}

	first, ok := it.CandidateAt(0)
	if !ok {
		t.Fatal("CandidateAt(0) should succeed")
	}
	if first.ChannelID != 22 {
		t.Fatalf("sticky candidate ChannelID = %d, want 22", first.ChannelID)
	}

	if !it.Next() {
		t.Fatal("Next() should advance to first item")
	}
	if !it.IsSticky() {
		t.Fatal("first iterator item should be sticky")
	}
	if it.Item().ChannelID != 22 {
		t.Fatalf("Item().ChannelID = %d, want 22", it.Item().ChannelID)
	}
}

func TestIteratorResetReappliesStickyOrdering(t *testing.T) {
	resetStickySessions()
	SetSticky(1002, "gpt-4o", 22, 220)

	group := model.Group{
		ID:              10,
		Mode:            model.GroupModeRoundRobin,
		SessionKeepTime: 60,
		Items: []model.GroupItem{
			{ChannelID: 11, ModelName: "gpt-4o", Priority: 1},
			{ChannelID: 22, ModelName: "gpt-4o", Priority: 2},
			{ChannelID: 33, ModelName: "gpt-4o", Priority: 3},
		},
	}

	it := NewIterator(group, 1002, "gpt-4o")
	it.Reset()

	first, ok := it.CandidateAt(0)
	if !ok {
		t.Fatal("CandidateAt(0) should succeed after reset")
	}
	if first.ChannelID != 22 {
		t.Fatalf("reset should preserve sticky ordering, got ChannelID %d, want 22", first.ChannelID)
	}
	if it.Index() != -1 {
		t.Fatalf("iterator index after reset = %d, want -1", it.Index())
	}
}

func TestIteratorRecordsAttemptsInOrder(t *testing.T) {
	resetStickySessions()
	SetSticky(2001, "gpt-4o", 11, 111)

	group := model.Group{
		ID:              11,
		Mode:            model.GroupModeRoundRobin,
		SessionKeepTime: 60,
		Items: []model.GroupItem{
			{ChannelID: 11, ModelName: "gpt-4o", Priority: 1},
		},
	}

	it := NewIterator(group, 2001, "gpt-4o")
	if !it.Next() {
		t.Fatal("Next() should move to first item")
	}

	span := it.StartAttempt(11, 111, "channel-a")
	time.Sleep(2 * time.Millisecond)
	span.End(model.AttemptFailed, 500, "upstream failed")
	it.Skip(11, 111, "channel-a", "skip after failure")
	it.Record(11, 111, "channel-a", "gpt-4o", model.AttemptSuccess, 200, 25*time.Millisecond, "recovered")

	attempts := it.Attempts()
	if len(attempts) != 3 {
		t.Fatalf("Attempts() len = %d, want 3", len(attempts))
	}

	wantStatuses := []model.AttemptStatus{
		model.AttemptFailed,
		model.AttemptSkipped,
		model.AttemptSuccess,
	}
	for i := range wantStatuses {
		if attempts[i].AttemptNum != i+1 {
			t.Fatalf("attempt %d number = %d, want %d", i, attempts[i].AttemptNum, i+1)
		}
		if attempts[i].Status != wantStatuses[i] {
			t.Fatalf("attempt %d status = %s, want %s", i, attempts[i].Status, wantStatuses[i])
		}
	}
	if attempts[0].StatusCode != 500 {
		t.Fatalf("attempt[0] status code = %d, want 500", attempts[0].StatusCode)
	}
	if attempts[0].Duration <= 0 {
		t.Fatalf("attempt[0] duration = %d, want > 0", attempts[0].Duration)
	}
	if !attempts[0].Sticky || !attempts[1].Sticky || !attempts[2].Sticky {
		t.Fatalf("attempt sticky flags = %#v, want all true", attempts)
	}
	if attempts[2].StatusCode != 200 {
		t.Fatalf("attempt[2] status code = %d, want 200", attempts[2].StatusCode)
	}
	if attempts[2].Duration != 25 {
		t.Fatalf("attempt[2] duration = %d, want 25", attempts[2].Duration)
	}
}

func TestIteratorConcurrentRecordsKeepSequentialAttemptNumbers(t *testing.T) {
	resetStickySessions()

	group := model.Group{
		ID:   12,
		Mode: model.GroupModeRoundRobin,
		Items: []model.GroupItem{
			{ChannelID: 11, ModelName: "gpt-4o", Priority: 1},
		},
	}

	it := NewIterator(group, 0, "gpt-4o")
	if !it.Next() {
		t.Fatal("Next() should move to first item")
	}

	const total = 24
	var wg sync.WaitGroup
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			it.Record(11, 111, "channel-a", "gpt-4o", model.AttemptSkipped, 0, 0, "parallel record")
		}()
	}
	wg.Wait()

	attempts := it.Attempts()
	if len(attempts) != total {
		t.Fatalf("Attempts() len = %d, want %d", len(attempts), total)
	}

	seen := make(map[int]struct{}, total)
	for _, attempt := range attempts {
		if attempt.Status != model.AttemptSkipped {
			t.Fatalf("attempt status = %s, want %s", attempt.Status, model.AttemptSkipped)
		}
		seen[attempt.AttemptNum] = struct{}{}
	}
	for i := 1; i <= total; i++ {
		if _, ok := seen[i]; !ok {
			t.Fatalf("missing attempt number %d in %#v", i, attempts)
		}
	}
}

func TestIteratorRustBalancerDisabledFallsBackToGoOrder(t *testing.T) {
	resetStickySessions()
	t.Setenv("OCTOPUS_RUST_BALANCER", "0")

	group := model.Group{
		ID:   20,
		Mode: model.GroupModeRoundRobin,
		Items: []model.GroupItem{
			{ChannelID: 3, ModelName: "gpt-4o", Priority: 30},
			{ChannelID: 1, ModelName: "gpt-4o", Priority: 10},
			{ChannelID: 2, ModelName: "gpt-4o", Priority: 20},
		},
	}

	it := NewIterator(group, 0, "gpt-4o")
	want := []int{1, 2, 3}
	got := collectIteratorChannelIDs(it)
	if len(got) != len(want) {
		t.Fatalf("iterator length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("iterator order = %v, want %v", got, want)
		}
	}
}

func TestIteratorRustBalancerRoundRobinOrder(t *testing.T) {
	resetStickySessions()

	group := model.Group{
		ID:   21,
		Mode: model.GroupModeRoundRobin,
		Items: []model.GroupItem{
			{ChannelID: 3, ModelName: "gpt-4o", Priority: 30},
			{ChannelID: 1, ModelName: "gpt-4o", Priority: 10},
			{ChannelID: 2, ModelName: "gpt-4o", Priority: 20},
		},
	}

	it := NewIterator(group, 0, "gpt-4o")
	want := []int{1, 2, 3}
	got := collectIteratorChannelIDs(it)
	if len(got) != len(want) {
		t.Fatalf("iterator length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("iterator order = %v, want %v", got, want)
		}
	}
}

func TestIteratorRustBalancerWeightedOrder(t *testing.T) {
	resetStickySessions()

	group := model.Group{
		ID:   22,
		Mode: model.GroupModeWeighted,
		Items: []model.GroupItem{
			{ID: 1, ChannelID: 2, ModelName: "gpt-4o", Priority: 2, Weight: 3},
			{ID: 2, ChannelID: 1, ModelName: "gpt-4o", Priority: 1, Weight: 2},
			{ID: 3, ChannelID: 3, ModelName: "gpt-4o", Priority: 3, Weight: 1},
		},
	}

	it := NewIterator(group, 0, "gpt-4o")
	want := []int{1, 1, 2, 2, 2, 3}
	got := collectIteratorChannelIDs(it)
	if len(got) != len(want) {
		t.Fatalf("iterator length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("weighted sequence mismatch at %d: got %d want %d", i, got[i], want[i])
		}
	}
}

func TestIteratorRustBalancerSkipsStickyFirstCandidate(t *testing.T) {
	resetStickySessions()
	SetSticky(3001, "gpt-4o", 22, 220)

	group := model.Group{
		ID:              23,
		Mode:            model.GroupModeRoundRobin,
		SessionKeepTime: 60,
		Items: []model.GroupItem{
			{ChannelID: 11, ModelName: "gpt-4o", Priority: 1},
			{ChannelID: 22, ModelName: "gpt-4o", Priority: 2},
			{ChannelID: 33, ModelName: "gpt-4o", Priority: 3},
		},
	}

	it := NewIterator(group, 3001, "gpt-4o")
	if !it.Next() {
		t.Fatal("Next() should advance to first item")
	}
	if it.Item().ChannelID != 22 {
		t.Fatalf("sticky first item = %d, want 22", it.Item().ChannelID)
	}

	// Advance through the rest; the remaining order should still be valid.
	for it.Next() {
		if it.Item().ChannelID == 0 {
			t.Fatal("unexpected empty candidate")
		}
	}
}

func TestIteratorRustBalancerNotAppliedForAIDynamic(t *testing.T) {
	resetStickySessions()

	group := model.Group{
		ID:   24,
		Mode: model.GroupModeAIDynamic,
		Items: []model.GroupItem{
			{ChannelID: 3, ModelName: "gpt-4o", Priority: 30},
			{ChannelID: 1, ModelName: "gpt-4o", Priority: 10},
			{ChannelID: 2, ModelName: "gpt-4o", Priority: 20},
		},
	}

	it := NewIterator(group, 0, "gpt-4o")
	want := []int{1, 2, 3}
	got := collectIteratorChannelIDs(it)
	if len(got) != len(want) {
		t.Fatalf("iterator length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("AI dynamic fallback order = %v, want %v", got, want)
		}
	}
}

func collectIteratorChannelIDs(it *Iterator) []int {
	var ids []int
	for it.Next() {
		ids = append(ids, it.Item().ChannelID)
	}
	return ids
}
