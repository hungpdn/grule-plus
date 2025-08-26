package lru

import (
	"fmt"
	"testing"
	"time"

	"github.com/hungpdn/grule-plus/internal/cache/common"
)

func TestBasicSetGetDelete(t *testing.T) {
	c := New(0, 0)
	defer c.StopCleanup()

	if got := c.Len(); got != 0 {
		t.Fatalf("expected len 0 got %d", got)
	}

	c.Set("a", "va", 0)
	c.Set("b", "vb", 0)

	if v, ok := c.Get("a"); !ok || v != "va" {
		t.Fatalf("Get a failed: %v %v", v, ok)
	}

	if !c.Has("b") {
		t.Fatalf("Has b false")
	}

	if c.Len() != 2 {
		t.Fatalf("Len want 2 got %d", c.Len())
	}

	if !c.Delete("a") {
		t.Fatalf("Delete a failed")
	}

	if c.Has("a") {
		t.Fatalf("a still present")
	}
}

func TestLRUEviction(t *testing.T) {
	c := New(2, 0)
	defer c.StopCleanup()

	c.Set("k1", "v1", 0)
	c.Set("k2", "v2", 0)

	// Access k1 so k2 becomes LRU
	if _, ok := c.Get("k1"); !ok {
		t.Fatalf("expected k1 present")
	}

	c.Set("k3", "v3", 0)

	if c.Has("k2") {
		t.Fatalf("k2 should be evicted (LRU)")
	}
	if !c.Has("k1") || !c.Has("k3") {
		t.Fatalf("expected k1 and k3 present")
	}
}

func TestExpirationAndDefaultTTL(t *testing.T) {
	c := New(0, 0)
	defer c.StopCleanup()

	// default TTL should cap longer durations
	c.SetDefaultTTL(20 * time.Millisecond)
	c.Set("x", "vx", 50*time.Millisecond) // should use default 20ms
	time.Sleep(30 * time.Millisecond)
	if c.Has("x") {
		t.Fatalf("x should be expired by default TTL")
	}

	// when duration is zero and defaultTTL is zero => no-expire
	c.SetDefaultTTL(0)
	c.Set("y", "vy", 0)
	if !c.Has("y") {
		t.Fatalf("y should exist (no-expire)")
	}
}

func TestCleanupGoroutine(t *testing.T) {
	c := New(0, 15*time.Millisecond)
	defer c.StopCleanup()

	c.Set("z", "vz", 10*time.Millisecond)

	// Give cleanup goroutine time to run
	time.Sleep(60 * time.Millisecond)

	if c.Has("z") {
		t.Fatalf("z should be cleaned up by goroutine")
	}
}

func TestEvictedFuncAndSetEvictedFunc(t *testing.T) {
	events := make(chan int, 4)
	f := func(key, value any, event int) {
		events <- event
	}

	c := NewWithEvictionFunc(2, 0, f)
	// trigger eviction
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Set("c", 3, 0) // should evict oldest (a)

	select {
	case ev := <-events:
		if ev != common.EvictionEvent {
			t.Fatalf("expected EvictionEvent got %d", ev)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("eviction event not received")
	}

	// SetEvictedFunc should return error if called twice
	c2 := New(0, 0)
	defer c2.StopCleanup()
	if err := c2.SetEvictedFunc(f); err != nil {
		t.Fatalf("unexpected error setting eviction func: %v", err)
	}
	if err := c2.SetEvictedFunc(f); err == nil {
		t.Fatalf("expected error when setting eviction func twice")
	}

	// Closing cache will send ClearEvent for remaining entries
	c.Close()
	select {
	case ev := <-events:
		if ev != common.ClearEvent {
			t.Fatalf("expected ClearEvent got %d", ev)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("clear event not received on close")
	}
}

func TestNewAndLen(t *testing.T) {
	c := New(0, 0)
	defer c.StopCleanup()

	if c == nil {
		t.Fatalf("New returned nil cache")
	}
	if c.Len() != 0 {
		t.Fatalf("expected len 0 got %d", c.Len())
	}
}

func TestKeysAndClear(t *testing.T) {
	c := New(0, 0)
	defer c.StopCleanup()

	c.Set("a", 1, 0)
	c.Set("b", 2, 0)

	keys := c.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}

	c.Clear()
	if c.Len() != 0 {
		t.Fatalf("expected len 0 after Clear, got %d", c.Len())
	}

	keys = c.Keys()
	if len(keys) != 0 {
		t.Fatalf("expected 0 keys after Clear, got %d", len(keys))
	}
}

func TestRemoveOldest(t *testing.T) {
	c := New(2, 0)
	defer c.StopCleanup()

	c.Set("k1", 1, 0)
	c.Set("k2", 2, 0)

	// Remove oldest should remove k1
	c.RemoveOldest()
	if c.Has("k1") {
		t.Fatalf("k1 should have been removed by RemoveOldest")
	}
	if !c.Has("k2") {
		t.Fatalf("k2 should remain after RemoveOldest")
	}
}

func TestSetUpdatesExistingEntry(t *testing.T) {
	c := New(0, 0)
	defer c.StopCleanup()

	c.Set("u", 1, 0)
	c.Set("u", 2, 0) // update

	v, ok := c.Get("u")
	if !ok {
		t.Fatalf("expected key 'u' present after update")
	}
	if v.(int) != 2 {
		t.Fatalf("expected updated value 2 got %v", v)
	}
	if c.Len() != 1 {
		t.Fatalf("expected len 1 after update got %d", c.Len())
	}
}

func TestKeysReflectDelete(t *testing.T) {
	c := New(0, 0)
	defer c.StopCleanup()

	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	if !c.Delete("a") {
		t.Fatalf("expected Delete to return true for existing key")
	}

	keys := c.Keys()
	// keys should contain only "b"
	for _, k := range keys {
		if k == "a" {
			t.Fatalf("deleted key 'a' still present in Keys()")
		}
	}
	if !c.Has("b") {
		t.Fatalf("expected key 'b' to remain present")
	}
}

func TestMain(t *testing.T) {
	cache := New(3, 5*time.Second)
	defer cache.StopCleanup()

	fmt.Println("Setting initial entries...")
	cache.Set("itemcode1", nil, 20*time.Second)
	cache.Set("itemcode2", nil, 10*time.Second)
	cache.Set("itemcode3", nil, 0)

	fmt.Printf("Cache length: %d\n", cache.Len())

	if val, found := cache.Get("itemcode1"); found {
		fmt.Printf("Get itemcode1: %s\n", val)
	}

	fmt.Println("\nSetting itemcode4, expecting itemcode2 to be evicted (LRU)...")
	cache.Set("itemcode4", nil, 10*time.Second)
	fmt.Printf("Cache length: %d\n", cache.Len())

	if _, found := cache.Get("itemcode2"); !found {
		fmt.Println("itemcode2 correctly evicted or expired.")
	}

	fmt.Println("\nWaiting for itemcode4 TTL 10s and itemcode1 TTL 20s to potentially expire...")
	fmt.Println("itemcode3 should not expire.")

	time.Sleep(12 * time.Second)

	if val, found := cache.Get("itemcode1"); found {
		fmt.Printf("Get itemcode1 after 12s: %s (Still alive)\n", val)
	} else {
		fmt.Println("itemcode1 expired or evicted.")
	}

	if _, found := cache.Get("itemcode4"); !found {
		fmt.Println("itemcode4 correctly expired.")
	} else {
		fmt.Printf("itemcode4 is still present (should have expired or been cleaned up)\n")
	}

	if val, found := cache.Get("itemcode3"); found {
		fmt.Printf("Get itemcode3 after 12s: %s (Should always be present)\n", val)
	} else {
		fmt.Println("itemcode3 (no-expire) was unexpectedly removed.")
	}

	fmt.Printf("Cache length after 12s: %d\n", cache.Len())

	fmt.Println("\nWaiting for remaining items to expire (itemcode1)...")
	time.Sleep(10 * time.Second) // Tổng cộng 22 giây, itemcode1 (20s TTL) sẽ hết hạn

	if _, found := cache.Get("itemcode1"); !found {
		fmt.Println("itemcode1 correctly expired after ~22s.")
	} else {
		fmt.Printf("itemcode1 is still present (should have expired)\n")
	}
	fmt.Printf("Cache length after ~22s: %d\n", cache.Len())
}
