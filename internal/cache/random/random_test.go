package random

import (
	"testing"
	"time"
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

	keys := c.Keys()
	if len(keys) != 2 {
		t.Fatalf("Keys want 2 got %d", len(keys))
	}

	c.Clear()

	if c.Len() != 0 {
		t.Fatalf("Clear failed")
	}
}

func TestRandomEviction(t *testing.T) {
	c := New(2, 0)
	defer c.StopCleanup()

	c.Set("k1", "v1", 0)
	c.Set("k2", "v2", 0)

	// Should not evict yet
	if c.Len() != 2 {
		t.Fatalf("Len want 2 got %d", c.Len())
	}

	// Add third item, should trigger random eviction
	c.Set("k3", "v3", 0)

	// Should still have 2 items (one was evicted randomly)
	if c.Len() != 2 {
		t.Fatalf("Len want 2 got %d after eviction", c.Len())
	}
}

func TestExpirationAndDefaultTTL(t *testing.T) {
	c := New(0, time.Millisecond*10)
	defer c.StopCleanup()

	// Set default TTL
	c.SetDefaultTTL(time.Millisecond * 50)

	// Set item with no explicit TTL (should use default)
	c.Set("a", "va", 0)
	if !c.Has("a") {
		t.Fatalf("Item should exist immediately after set")
	}

	// Wait for expiration
	time.Sleep(time.Millisecond * 60)

	// Item should be expired
	if c.Has("a") {
		t.Fatalf("Item should have expired")
	}
}

func TestEvictedFuncAndSetEvictedFunc(t *testing.T) {
	var evictedKey any
	var evictedValue any

	c := New(1, 0)
	defer c.StopCleanup()

	err := c.SetEvictedFunc(func(key, value any, event int) {
		evictedKey = key
		evictedValue = value
	})
	if err != nil {
		t.Fatalf("SetEvictedFunc failed: %v", err)
	}

	c.Set("k1", "v1", 0)
	c.Set("k1", "v2", 0) // This should evict k1

	if evictedKey != nil || evictedValue != nil {
		t.Fatalf("Eviction callback not called correctly: got key=%v value=%v, expected key=k1 value=v1", evictedKey, evictedValue)
	}
}
