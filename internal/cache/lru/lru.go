// lru implements an LRU cache
package lru

import (
	"container/list"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/hungpdn/grule-plus/internal/cache/common"
)

// Cache is an LRU cache structure
type Cache struct {
	maxEntries int                   // The maximum number of cache entries before an entry is evicte, zero means no limit
	entries    map[any]*list.Element // Map for quick access to cache entries
	ll         *list.List            // Doubly linked list to track LRU order
	mu         sync.RWMutex          // Mutex to ensure concurrent access safety
	onEvicted  common.EvictedFunc    // OnEvicted optionally specifies a callback function to be executed when an entry is purged from the cache
	// cleanup
	defaultTTL      time.Duration // default TTL for item expire
	cleanupInterval time.Duration // how often to run the expired entry cleaner
	stopChan        chan struct{} // Channel to stop cleanup goroutine
}

// entry represents an entry in the LRU cache
type entry struct {
	key        any
	value      any
	expiration int64 // Unix timestamp (nanoseconds) when the item expires, 0 means never expires
}

// New creates a new LRU cache
// maxEntries: the maximum number of cache entries before an entry is evicted, zero means no limit
// cleanupInterval: how often to run the expired entry cleaner
func New(maxEntries int, cleanupInterval time.Duration) *Cache {
	cache := &Cache{
		maxEntries:      maxEntries,
		entries:         make(map[any]*list.Element),
		ll:              list.New(),
		cleanupInterval: cleanupInterval,
		stopChan:        make(chan struct{}),
	}
	if cache.cleanupInterval > 0 {
		go cache.startCleanup()
	}
	return cache
}

// NewWithEvictionFunc creates an LRU of the given size with the given eviction func
func NewWithEvictionFunc(maxEntries int, cleanupInterval time.Duration, f common.EvictedFunc) *Cache {
	c := New(maxEntries, cleanupInterval)
	c.onEvicted = f
	return c
}

// SetEvictedFunc updates the eviction func
func (c *Cache) SetEvictedFunc(f common.EvictedFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.onEvicted != nil {
		return fmt.Errorf("lru cache eviction function is already set")
	}
	c.onEvicted = f
	return nil
}

// SetDefaultTTL updates the defaultTTL
func (c *Cache) SetDefaultTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.defaultTTL = ttl
}

// Add adds or updates a value to the cache
func (c *Cache) Set(key any, value any, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.entries == nil {
		c.entries = make(map[any]*list.Element)
		c.ll = list.New()
	}

	expiration := int64(0)
	if duration > 0 {
		if c.defaultTTL > 0 && duration > c.defaultTTL {
			expiration = time.Now().Add(c.defaultTTL).UnixNano()
		} else {
			expiration = time.Now().Add(duration).UnixNano()
		}
	} else {
		if c.defaultTTL > 0 {
			expiration = time.Now().Add(c.defaultTTL).UnixNano()
		}
	}

	if ele, ok := c.entries[key]; ok {
		c.ll.MoveToFront(ele)
		entry := ele.Value.(*entry)
		entry.value = value
		entry.expiration = expiration
		return
	}

	if c.maxEntries != 0 && c.ll.Len() >= c.maxEntries {
		c.RemoveOldest()
	}

	entry := &entry{
		key:        key,
		value:      value,
		expiration: expiration,
	}
	ele := c.ll.PushFront(entry)
	c.entries[key] = ele
}

// Get looks up a key's value from the cache
func (c *Cache) Get(key any) (value any, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.entries == nil {
		return
	}
	if ele, hit := c.entries[key]; hit {
		entry := ele.Value.(*entry)
		if entry.expiration > 0 && time.Now().UnixNano() > entry.expiration {
			c.removeElement(ele, common.ExpirationEvent)
			return
		}
		c.ll.MoveToFront(ele)
		return entry.value, true
	}
	return
}

// Has returns true if the key exists in the cache.
func (c *Cache) Has(key any) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.entries == nil {
		return false
	}
	if ele, hit := c.entries[key]; hit {
		entry := ele.Value.(*entry)
		if entry.expiration > 0 && time.Now().UnixNano() > entry.expiration {
			return false
		}
		return true
	}
	return false
}

// Delete deletes a key-value from the cache
func (c *Cache) Delete(key any) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, hit := c.entries[key]; hit {
		c.removeElement(ele, common.DeleteEvent)
		return true
	}
	return false
}

// Len returns the number of items in the cache
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.entries == nil {
		return 0
	}
	return c.ll.Len()
}

// Clear purges all stored items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.onEvicted != nil {
		for _, e := range c.entries {
			entry := e.Value.(*entry)
			c.onEvicted(entry.key, entry.value, common.ClearEvent)
		}
	}
	c.ll = nil
	c.entries = nil
}

// RemoveOldest removes the oldest item from the cache
func (c *Cache) RemoveOldest() {
	if c.entries == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele, common.EvictionEvent)
	}
}

// removeElement removes the a item from the cache
func (c *Cache) removeElement(e *list.Element, event int) {
	c.ll.Remove(e)
	entry := e.Value.(*entry)
	delete(c.entries, entry.key)
	if c.onEvicted != nil {
		c.onEvicted(entry.key, entry.value, event)
	}
}

// startCleanup cleanup expired entry periodically
func (c *Cache) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("Cache: Running cleanup routine...")
			c.cleanupExpiredEntries()
			runtime.GC()
		case <-c.stopChan:
			fmt.Println("Cache: Stopping cleanup routine...")
			return
		}
	}
}

// cleanupExpiredEntries cleanup expired entry
func (c *Cache) cleanupExpiredEntries() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixNano()
	ele := c.ll.Back()
	for ele != nil {
		prev := ele.Prev()
		entry := ele.Value.(*entry)
		if entry.expiration > 0 && now > entry.expiration {
			c.removeElement(ele, common.ExpirationEvent)
		}
		ele = prev
	}
}

// StopCleanup stops goroutine cleanup
func (c *Cache) StopCleanup() {
	if c.stopChan != nil {
		close(c.stopChan)
	}
}

// Keys returns a slice of the keys in the cache
func (c *Cache) Keys() []any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]interface{}, 0, len(c.entries))
	for k := range c.entries {
		keys = append(keys, k)
	}
	return keys
}

// Close purges all key-value pairs from the cache and stop cleanup
func (c *Cache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.StopCleanup()

	if c.onEvicted != nil {
		for _, e := range c.entries {
			entry := e.Value.(*entry)
			c.onEvicted(entry.key, entry.value, common.ClearEvent)
		}
	}
	c.ll = nil
	c.entries = nil
}
