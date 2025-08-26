package lfu

import (
	"testing"
	"time"

	"github.com/hungpdn/grule-plus/internal/cache/common"
)

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

func TestSetAndGet(t *testing.T) {
	c := New(0, 0)
	defer c.StopCleanup()
	c.Set("a", "va", 0)
	v, ok := c.Get("a")
	if !ok || v != "va" {
		t.Fatalf("Get a failed: %v %v", v, ok)
	}
}

func TestHasAndDelete(t *testing.T) {
	c := New(2, 0)
	defer c.StopCleanup()
	c.Set("b", "vb", 0)
	if !c.Has("b") {
		t.Fatalf("Has b false")
	}
	if !c.Delete("b") {
		t.Fatalf("Delete b failed")
	}
	if c.Has("b") {
		t.Fatalf("b still present after delete")
	}
}

func TestEvictionPolicy(t *testing.T) {
	c := New(2, 0)
	defer c.StopCleanup()
	c.Set("x", 1, 0)
	c.Set("y", 2, 0)
	// Access x to increase its freq
	c.Get("x")
	c.Set("z", 3, 0) // should evict y (lowest freq)
	if c.Has("y") {
		t.Fatalf("y should be evicted (LFU)")
	}
	if !c.Has("x") || !c.Has("z") {
		t.Fatalf("expected x and z present")
	}
}

func TestDefaultTTLAndExpiration(t *testing.T) {
	c := New(0, 0)
	defer c.StopCleanup()
	c.SetDefaultTTL(20 * time.Millisecond)
	c.Set("ttl", "vttl", 50*time.Millisecond) // should use default 20ms
	time.Sleep(30 * time.Millisecond)
	if c.Has("ttl") {
		t.Fatalf("ttl should be expired by default TTL")
	}
	c.SetDefaultTTL(0)
	c.Set("noexpire", "vne", 0)
	if !c.Has("noexpire") {
		t.Fatalf("noexpire should exist (no-expire)")
	}
}

func TestCleanupGoroutine(t *testing.T) {
	c := New(0, 15*time.Millisecond)
	defer c.StopCleanup()
	c.Set("z", "vz", 10*time.Millisecond)
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
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Set("c", 3, 0) // should evict a
	select {
	case ev := <-events:
		if ev != common.EvictionEvent {
			t.Fatalf("expected EvictionEvent got %d", ev)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("eviction event not received")
	}
	c2 := New(0, 0)
	defer c2.StopCleanup()
	if err := c2.SetEvictedFunc(f); err != nil {
		t.Fatalf("unexpected error setting eviction func: %v", err)
	}
	if err := c2.SetEvictedFunc(f); err == nil {
		t.Fatalf("expected error when setting eviction func twice")
	}
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

func TestKeysAndClear(t *testing.T) {
	c := New(2, 0)
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
