package op

import (
	"context"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestRelayLogStreamTokenCreateVerifyAndExpire(t *testing.T) {
	originalTTL := relayLogStreamTokenTTL
	originalNow := relayLogStreamTokenNow
	relayLogStreamTokenTTL = time.Minute
	base := time.Unix(1700000000, 0)
	relayLogStreamTokenNow = func() time.Time { return base }
	t.Cleanup(func() {
		relayLogStreamTokenTTL = originalTTL
		relayLogStreamTokenNow = originalNow
		relayLogStreamTokensLock.Lock()
		relayLogStreamTokens = make(map[string]relayLogStreamToken)
		relayLogStreamTokensLock.Unlock()
	})

	token, err := RelayLogStreamTokenCreate()
	if err != nil {
		t.Fatalf("RelayLogStreamTokenCreate() error = %v", err)
	}
	if token == "" {
		t.Fatal("RelayLogStreamTokenCreate() returned empty token")
	}
	if !RelayLogStreamTokenVerify(token) {
		t.Fatal("RelayLogStreamTokenVerify() = false, want true before expiry")
	}

	relayLogStreamTokenNow = func() time.Time { return base.Add(2 * time.Minute) }
	if RelayLogStreamTokenVerify(token) {
		t.Fatal("RelayLogStreamTokenVerify() = true, want false after expiry")
	}

	relayLogStreamTokensLock.RLock()
	_, ok := relayLogStreamTokens[token]
	relayLogStreamTokensLock.RUnlock()
	if ok {
		t.Fatal("expired stream token still present in cache")
	}
}

func TestRelayLogStreamTokenConsumeIsSingleUse(t *testing.T) {
	originalTTL := relayLogStreamTokenTTL
	originalNow := relayLogStreamTokenNow
	relayLogStreamTokenTTL = time.Minute
	base := time.Unix(1700000000, 0)
	relayLogStreamTokenNow = func() time.Time { return base }
	t.Cleanup(func() {
		relayLogStreamTokenTTL = originalTTL
		relayLogStreamTokenNow = originalNow
		relayLogStreamTokensLock.Lock()
		relayLogStreamTokens = make(map[string]relayLogStreamToken)
		relayLogStreamTokensLock.Unlock()
	})

	token, err := RelayLogStreamTokenCreate()
	if err != nil {
		t.Fatalf("RelayLogStreamTokenCreate() error = %v", err)
	}
	if !RelayLogStreamTokenConsume(token) {
		t.Fatal("RelayLogStreamTokenConsume() = false, want true on first use")
	}
	if RelayLogStreamTokenConsume(token) {
		t.Fatal("RelayLogStreamTokenConsume() = true, want false after token is consumed")
	}
	if RelayLogStreamTokenVerify(token) {
		t.Fatal("RelayLogStreamTokenVerify() = true, want false after consume")
	}
	relayLogStreamTokensLock.RLock()
	_, ok := relayLogStreamTokens[token]
	relayLogStreamTokensLock.RUnlock()
	if ok {
		t.Fatal("consumed stream token still present in cache")
	}
}

func TestRelayLogStreamTokenCreateCapsDistinctEntries(t *testing.T) {
	originalTTL := relayLogStreamTokenTTL
	originalNow := relayLogStreamTokenNow
	relayLogStreamTokenTTL = time.Hour
	base := time.Unix(1700000000, 0)
	relayLogStreamTokenNow = func() time.Time { return base }
	t.Cleanup(func() {
		relayLogStreamTokenTTL = originalTTL
		relayLogStreamTokenNow = originalNow
		relayLogStreamTokensLock.Lock()
		relayLogStreamTokens = make(map[string]relayLogStreamToken)
		relayLogStreamTokenSeq = 0
		relayLogStreamTokensLock.Unlock()
	})

	created := make([]string, 0, relayLogStreamTokenMaxEntries+32)
	for i := 0; i < relayLogStreamTokenMaxEntries+32; i++ {
		token, err := RelayLogStreamTokenCreate()
		if err != nil {
			t.Fatalf("RelayLogStreamTokenCreate() error = %v", err)
		}
		created = append(created, token)
	}

	relayLogStreamTokensLock.RLock()
	gotLen := len(relayLogStreamTokens)
	relayLogStreamTokensLock.RUnlock()
	if gotLen != relayLogStreamTokenMaxEntries {
		t.Fatalf("len(relayLogStreamTokens) = %d, want %d", gotLen, relayLogStreamTokenMaxEntries)
	}

	if RelayLogStreamTokenVerify(created[0]) {
		t.Fatal("oldest token still valid after cache cap eviction")
	}
	if !RelayLogStreamTokenVerify(created[len(created)-1]) {
		t.Fatal("newest token invalid after cache cap eviction")
	}
}

func TestRelayLogAddNotifiesSubscribers(t *testing.T) {
	ctx := setupOpTestDB(t)

	ch := RelayLogSubscribe()
	t.Cleanup(func() {
		RelayLogUnsubscribe(ch)
	})

	entry := model.RelayLog{Time: 1700000000, RequestModelName: "gpt-test"}
	if err := RelayLogAdd(ctx, entry); err != nil {
		t.Fatalf("RelayLogAdd() error = %v", err)
	}

	select {
	case got := <-ch:
		if got.RequestModelName != entry.RequestModelName {
			t.Fatalf("subscriber log model = %q, want %q", got.RequestModelName, entry.RequestModelName)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for relay log subscriber notification")
	}
}

func TestRelayLogAddDoesNotBlockWhenSubscriberBufferIsFull(t *testing.T) {
	setupOpTestDB(t)

	ch := RelayLogSubscribe()
	t.Cleanup(func() {
		RelayLogUnsubscribe(ch)
	})

	for i := 0; i < cap(ch); i++ {
		ch <- model.RelayLog{Time: int64(i + 1)}
	}

	done := make(chan error, 1)
	go func() {
		done <- RelayLogAdd(context.Background(), model.RelayLog{Time: 1700000001, RequestModelName: "buffer-full"})
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RelayLogAdd() error = %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("RelayLogAdd() blocked with a full subscriber buffer")
	}
}
