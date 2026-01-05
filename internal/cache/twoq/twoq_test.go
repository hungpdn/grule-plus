package twoq

import (
	"testing"
	"time"

	"github.com/hungpdn/grule-plus/internal/cache/common"
)

func TestNewAndLen(t *testing.T) {
	cache := New(10, 0)
	if cache.Len() != 0 {
		t.Errorf("expected length 0, got %d", cache.Len())
	}
}

func TestSetAndGet(t *testing.T) {
	cache := New(10, 0)

	// Test Set and Get
	cache.Set("key1", "value1", 0)
	if value, ok := cache.Get("key1"); !ok || value != "value1" {
		t.Errorf("expected value1, got %v", value)
	}

	// Test update
	cache.Set("key1", "value2", 0)
	if value, ok := cache.Get("key1"); !ok || value != "value2" {
		t.Errorf("expected value2, got %v", value)
	}
}

func TestHas(t *testing.T) {
	cache := New(10, 0)

	cache.Set("key1", "value1", 0)
	if !cache.Has("key1") {
		t.Error("expected key1 to exist")
	}

	if cache.Has("key2") {
		t.Error("expected key2 to not exist")
	}
}

func TestEvictionPolicy(t *testing.T) {
	cache := New(3, 0)

	// Fill cache
	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)
	cache.Set("key3", "value3", 0)

	// Access key1 to move it to A2
	cache.Get("key1")

	// Add new key, should evict key2 from A1
	cache.Set("key4", "value4", 0)

	if cache.Has("key2") {
		t.Error("expected key2 to be evicted")
	}

	if !cache.Has("key1") || !cache.Has("key3") || !cache.Has("key4") {
		t.Error("expected key1, key3, key4 to remain")
	}
}

func TestDefaultTTLAndExpiration(t *testing.T) {
	cache := New(10, 100*time.Millisecond)
	cache.SetDefaultTTL(200 * time.Millisecond)

	cache.Set("key1", "value1", 0) // Should use default TTL

	// Should exist immediately
	if !cache.Has("key1") {
		t.Error("expected key1 to exist immediately")
	}

	// Wait for expiration
	time.Sleep(250 * time.Millisecond)

	if cache.Has("key1") {
		t.Error("expected key1 to be expired")
	}
}

func TestCleanupGoroutine(t *testing.T) {
	cache := New(10, 100*time.Millisecond)
	cache.SetDefaultTTL(150 * time.Millisecond)

	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	if cache.Len() != 0 {
		t.Errorf("expected cache to be empty after cleanup, got %d items", cache.Len())
	}

	cache.Close()
}

func TestEvictedFuncAndSetEvictedFunc(t *testing.T) {
	cache := New(2, 0)

	var evictedKey any
	var evictedValue any
	var evictedEvent int

	cache.SetEvictedFunc(func(key, value any, event int) {
		evictedKey = key
		evictedValue = value
		evictedEvent = event
	})

	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)
	cache.Set("key3", "value3", 0) // Should evict key1

	if evictedKey != "key1" || evictedValue != "value1" || evictedEvent != common.EvictionEvent {
		t.Errorf("expected eviction of key1, got key=%v, value=%v, event=%d", evictedKey, evictedValue, evictedEvent)
	}
}

func TestKeysAndClear(t *testing.T) {
	cache := New(10, 0)

	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)

	keys := cache.Keys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}

	cache.Clear()

	if cache.Len() != 0 {
		t.Error("expected cache to be empty after clear")
	}
}

func Test2QPromotion(t *testing.T) {
	cache := New(4, 0)

	// Add items to A1
	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)

	// Access key1 to promote to A2
	cache.Get("key1")

	// Add more items
	cache.Set("key3", "value3", 0)
	cache.Set("key4", "value4", 0)

	// Access key2 to promote to A2
	cache.Get("key2")

	// Add item that should evict from A1
	cache.Set("key5", "value5", 0)

	// key1 and key2 should be in A2, key3 should be evicted
	if !cache.Has("key1") || !cache.Has("key2") {
		t.Error("expected key1 and key2 to remain")
	}

	if cache.Has("key3") {
		t.Error("expected key3 to be evicted")
	}
}