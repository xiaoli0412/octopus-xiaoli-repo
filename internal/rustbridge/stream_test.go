package rustbridge

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestStreamBufferFeedCompleteEvents(t *testing.T) {
	buf := NewStreamBuffer()
	defer buf.Close()

	events, err := buf.Feed("data: hello\n\ndata: world\n\n")
	if err != nil {
		t.Fatalf("feed error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d: %v", len(events), events)
	}
}

func TestStreamBufferPartialEvent(t *testing.T) {
	buf := NewStreamBuffer()
	defer buf.Close()

	events, err := buf.Feed("data: part")
	if err != nil {
		t.Fatalf("feed error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}

	events, err = buf.Feed("ial\n\ndata: next\n\n")
	if err != nil {
		t.Fatalf("feed error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d: %v", len(events), events)
	}
}

func TestStreamBufferTake(t *testing.T) {
	buf := NewStreamBuffer()
	defer buf.Close()

	feedEvents, err := buf.Feed("data: a\n\ndata: b\n\n")
	if err != nil {
		t.Fatalf("feed error: %v", err)
	}
	if len(feedEvents) != 2 {
		t.Fatalf("expected 2 events from feed, got %d", len(feedEvents))
	}

	// After Feed has already extracted complete events, Take should return none.
	takeEvents, err := buf.Take()
	if err != nil {
		t.Fatalf("take error: %v", err)
	}
	if len(takeEvents) != 0 {
		t.Fatalf("expected 0 events from take, got %d", len(takeEvents))
	}
}

func TestStreamBufferTakePartial(t *testing.T) {
	buf := NewStreamBuffer()
	defer buf.Close()

	feedEvents, err := buf.Feed("data: partial")
	if err != nil {
		t.Fatalf("feed error: %v", err)
	}
	if len(feedEvents) != 0 {
		t.Fatalf("expected 0 events from first feed, got %d", len(feedEvents))
	}

	takeEvents, err := buf.Take()
	if err != nil {
		t.Fatalf("take error: %v", err)
	}
	if len(takeEvents) != 0 {
		t.Fatalf("expected 0 events from take, got %d", len(takeEvents))
	}

	feedEvents, err = buf.Feed(" event\n\n")
	if err != nil {
		t.Fatalf("feed error: %v", err)
	}
	if len(feedEvents) != 1 {
		t.Fatalf("expected 1 event from second feed, got %d", len(feedEvents))
	}
}

func TestStreamBufferConcurrency(t *testing.T) {
	if !Enabled() {
		t.Skip("rust backend not enabled")
	}
	buf := NewStreamBuffer()
	defer buf.Close()

	var wg sync.WaitGroup
	var totalFed atomic.Int32
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			chunk := "data: " + strings.Repeat("x", idx) + "\n\n"
			if evts, err := buf.Feed(chunk); err == nil {
				totalFed.Add(int32(len(evts)))
			}
		}(i)
	}
	wg.Wait()
	takeEvents, err := buf.Take()
	if err != nil {
		t.Fatalf("take error: %v", err)
	}
	total := int(totalFed.Load()) + len(takeEvents)
	if total != 100 {
		t.Fatalf("expected 100 events total, got %d", total)
	}
}

func TestStreamEnvSwitch(t *testing.T) {
	if !Enabled() {
		t.Skip("rust backend not enabled")
	}
	bufRust := NewStreamBuffer()
	defer bufRust.Close()
	rustEvents, err := bufRust.Feed("data: x\n\n")
	if err != nil {
		t.Fatalf("rust feed error: %v", err)
	}

	t.Setenv(envDisableRustStream, "0")
	bufGo := NewStreamBuffer()
	defer bufGo.Close()
	goEvents, err := bufGo.Feed("data: x\n\n")
	if err != nil {
		t.Fatalf("go feed error: %v", err)
	}

	if len(rustEvents) != len(goEvents) {
		t.Fatalf("env switch changed event count: rust=%d go=%d", len(rustEvents), len(goEvents))
	}
	if len(rustEvents) > 0 && rustEvents[0] != goEvents[0] {
		t.Fatalf("env switch changed event content: rust=%q go=%q", rustEvents[0], goEvents[0])
	}
}
