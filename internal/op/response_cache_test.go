package op

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
)

func newTestResponseCache(maxEntries int) *ResponseCache {
	if maxEntries <= 0 {
		maxEntries = defaultResponseCacheMaxEntries
	}
	return &ResponseCache{
		entries:    make(map[string]*cacheEntry),
		maxEntries: maxEntries,
	}
}

func TestResponseCacheComputeCacheKeyDeterministic(t *testing.T) {
	model := "gpt-4"
	messages := json.RawMessage(`[{"role":"user","content":"hello"}]`)
	params := map[string]interface{}{"temperature": 0.7}

	key1 := ComputeCacheKey(model, messages, params)
	key2 := ComputeCacheKey(model, messages, params)
	if key1 != key2 {
		t.Fatalf("ComputeCacheKey not deterministic: %q vs %q", key1, key2)
	}
	if len(key1) != 64 {
		t.Fatalf("expected 64-char hex key, got %d chars: %q", len(key1), key1)
	}
}

func TestResponseCacheComputeCacheKeyDiffers(t *testing.T) {
	messages := json.RawMessage(`[{"role":"user","content":"hello"}]`)

	key1 := ComputeCacheKey("gpt-4", messages, nil)
	key2 := ComputeCacheKey("gpt-3.5-turbo", messages, nil)
	if key1 == key2 {
		t.Fatal("different models should produce different keys")
	}

	key3 := ComputeCacheKey("gpt-4", json.RawMessage(`[{"role":"user","content":"world"}]`), nil)
	if key1 == key3 {
		t.Fatal("different messages should produce different keys")
	}

	key4 := ComputeCacheKey("gpt-4", messages, map[string]interface{}{"temperature": 0.5})
	if key1 == key4 {
		t.Fatal("different params should produce different keys")
	}
}

func TestResponseCacheHitAndMiss(t *testing.T) {
	rc := newTestResponseCache(100)
	key := "test-key"
	resp := []byte(`{"id":"chatcmpl-1","choices":[]}`)

	// Miss before set
	got, ok := rc.Get(key)
	if ok {
		t.Fatalf("expected miss before Set, got %s", got)
	}

	// Set then hit
	rc.Set(key, resp, 10*time.Second)
	got, ok = rc.Get(key)
	if !ok {
		t.Fatal("expected hit after Set")
	}
	if string(got) != string(resp) {
		t.Fatalf("response mismatch: got %q, want %q", got, resp)
	}
}

func TestResponseCacheGetReturnsCopy(t *testing.T) {
	rc := newTestResponseCache(100)
	key := "copy-key"
	resp := []byte(`{"ok":true}`)

	rc.Set(key, resp, 10*time.Second)
	got, _ := rc.Get(key)
	got[0] = 'X'

	got2, _ := rc.Get(key)
	if string(got2) != `{"ok":true}` {
		t.Fatalf("cached response was mutated: %q", got2)
	}
}

func TestResponseCacheExpired(t *testing.T) {
	rc := newTestResponseCache(100)
	key := "expire-key"
	resp := []byte(`{"ok":true}`)

	rc.Set(key, resp, 50*time.Millisecond)

	// Hit immediately
	if _, ok := rc.Get(key); !ok {
		t.Fatal("expected hit before expiry")
	}

	time.Sleep(80 * time.Millisecond)

	// Miss after expiry
	if _, ok := rc.Get(key); ok {
		t.Fatal("expected miss after expiry")
	}
}

func TestResponseCacheEvictExpired(t *testing.T) {
	rc := newTestResponseCache(100)

	rc.Set("k1", []byte("v1"), 30*time.Millisecond)
	rc.Set("k2", []byte("v2"), 200*time.Millisecond)

	time.Sleep(60 * time.Millisecond)

	rc.EvictExpired()

	if rc.Len() != 1 {
		t.Fatalf("expected 1 entry after eviction, got %d", rc.Len())
	}
	if _, ok := rc.Get("k2"); !ok {
		t.Fatal("k2 should still be valid")
	}
}

func TestResponseCacheLRUEviction(t *testing.T) {
	rc := newTestResponseCache(3)

	rc.Set("k1", []byte("v1"), 10*time.Second)
	time.Sleep(time.Millisecond)
	rc.Set("k2", []byte("v2"), 10*time.Second)
	time.Sleep(time.Millisecond)
	rc.Set("k3", []byte("v3"), 10*time.Second)
	time.Sleep(time.Millisecond)

	// Access k1 to make it most-recently-used
	rc.Get("k1")
	time.Sleep(time.Millisecond)

	// Adding k4 should evict k2 (least recently used)
	rc.Set("k4", []byte("v4"), 10*time.Second)

	if _, ok := rc.Get("k2"); ok {
		t.Fatal("k2 should have been evicted (LRU)")
	}
	if _, ok := rc.Get("k1"); !ok {
		t.Fatal("k1 should still be present")
	}
	if _, ok := rc.Get("k3"); !ok {
		t.Fatal("k3 should still be present")
	}
	if _, ok := rc.Get("k4"); !ok {
		t.Fatal("k4 should still be present")
	}
}

func TestResponseCacheLRUEvictionOrder(t *testing.T) {
	rc := newTestResponseCache(2)

	rc.Set("a", []byte("1"), 10*time.Second)
	time.Sleep(time.Millisecond)
	rc.Set("b", []byte("2"), 10*time.Second)
	time.Sleep(time.Millisecond)

	// Access a -> b becomes LRU
	rc.Get("a")
	time.Sleep(time.Millisecond)

	rc.Set("c", []byte("3"), 10*time.Second)
	// b should be evicted
	if _, ok := rc.Get("b"); ok {
		t.Fatal("b should be evicted")
	}

	// Now a and c exist. a is LRU (accessed before c was set).
	time.Sleep(time.Millisecond)
	rc.Set("d", []byte("4"), 10*time.Second)
	// a should be evicted
	if _, ok := rc.Get("a"); ok {
		t.Fatal("a should be evicted")
	}
	if _, ok := rc.Get("c"); !ok {
		t.Fatal("c should still exist")
	}
	if _, ok := rc.Get("d"); !ok {
		t.Fatal("d should still exist")
	}
}

func TestResponseCacheSetMaxEntries(t *testing.T) {
	rc := newTestResponseCache(10)

	rc.SetMaxEntries(2)
	rc.Set("k1", []byte("v1"), 10*time.Second)
	time.Sleep(time.Millisecond)
	rc.Set("k2", []byte("v2"), 10*time.Second)
	time.Sleep(time.Millisecond)
	rc.Set("k3", []byte("v3"), 10*time.Second)

	if rc.Len() != 2 {
		t.Fatalf("expected 2 entries after SetMaxEntries(2), got %d", rc.Len())
	}
	if _, ok := rc.Get("k1"); ok {
		t.Fatal("k1 should have been evicted")
	}
}

func TestResponseCacheClear(t *testing.T) {
	rc := newTestResponseCache(100)
	rc.Set("k1", []byte("v1"), 10*time.Second)
	rc.Set("k2", []byte("v2"), 10*time.Second)

	rc.Clear()
	if rc.Len() != 0 {
		t.Fatalf("expected 0 entries after Clear, got %d", rc.Len())
	}
}

func TestResponseCacheGlobalSingleton(t *testing.T) {
	c1 := GetResponseCache()
	c2 := GetResponseCache()
	if c1 != c2 {
		t.Fatal("GetResponseCache should return the same instance")
	}
}

// TestResponseCacheConcurrentSafety exercises concurrent reads and writes
// under the race detector to ensure sync.RWMutex protects all access.
func TestResponseCacheConcurrentSafety(t *testing.T) {
	rc := newTestResponseCache(50)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j%5)
				rc.Set(key, []byte(fmt.Sprintf("val-%d-%d", id, j)), 5*time.Second)
				rc.Get(key)
			}
		}(i)
	}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				rc.EvictExpired()
				rc.Len()
			}
		}()
	}
	wg.Wait()
}
