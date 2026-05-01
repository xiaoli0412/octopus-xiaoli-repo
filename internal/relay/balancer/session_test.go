package balancer

import (
	"sync"
	"testing"
	"time"
)

func TestGetStickyReturnsNilWhenMissingOrExpired(t *testing.T) {
	globalSession = sync.Map{}

	if got := GetSticky(1, "gpt-4o", time.Minute); got != nil {
		t.Fatalf("GetSticky() for missing session = %#v, want nil", got)
	}

	key := sessionKey(1, "gpt-4o")
	globalSession.Store(key, &SessionEntry{
		ChannelID:    9,
		ChannelKeyID: 99,
		Timestamp:    time.Now().Add(-2 * time.Minute),
	})

	if got := GetSticky(1, "gpt-4o", time.Minute); got != nil {
		t.Fatalf("GetSticky() for expired session = %#v, want nil", got)
	}
	if _, ok := globalSession.Load(key); ok {
		t.Fatal("expired sticky session should be lazily deleted")
	}
}

func TestSetStickyStoresLatestEntry(t *testing.T) {
	globalSession = sync.Map{}

	SetSticky(7, "gpt-4o", 12, 120)
	first := GetSticky(7, "gpt-4o", time.Minute)
	if first == nil {
		t.Fatal("GetSticky() should return stored entry")
	}
	if first.ChannelID != 12 || first.ChannelKeyID != 120 {
		t.Fatalf("first sticky entry = %#v, want channel 12 key 120", first)
	}

	SetSticky(7, "gpt-4o", 22, 220)
	second := GetSticky(7, "gpt-4o", time.Minute)
	if second == nil {
		t.Fatal("GetSticky() should return updated entry")
	}
	if second.ChannelID != 22 || second.ChannelKeyID != 220 {
		t.Fatalf("updated sticky entry = %#v, want channel 22 key 220", second)
	}
}
