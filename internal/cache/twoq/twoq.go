// twoq implements a 2Q cache.
package twoq

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/hungpdn/grule-plus/internal/cache/common"
)

// Cache is a 2Q cache structure
type Cache struct {
	maxEntries int                   // The maximum number of cache entries before an entry is evicted, zero means no limit
	entries    map[any]*list.Element // Map for quick access to cache entries
	a1         *list.List            // A1: FIFO queue for new entries
	a2         *list.List            // A2: LRU queue for frequently accessed entries
	b          *list.List            // B: ghost queue for evicted entries
	kin        int                   // Size of A1 queue (typically maxEntries/4)
	mu         sync.RWMutex          // Mutex to ensure concurrent access safety
	onEvicted  common.EvictedFunc    // OnEvicted optionally specifies a callback function to be executed when an entry is purged from the cache
	// cleanup
	defaultTTL      time.Duration // default TTL for item expire
	cleanupInterval time.Duration // how often to run the expired entry cleaner
	stopChan        chan struct{} // Channel to stop cleanup goroutine
	closed          bool          // Flag to indicate if cache is closed
}

// entry represents an entry in the 2Q cache
type entry struct {
	key        any
	value      any
	expiration int64 // Unix timestamp (nanoseconds) when the item expires, 0 means never expires
}

// New creates a new 2Q cache
// maxEntries: the maximum number of cache entries before an entry is evicted, zero means no limit
// cleanupInterval: how often to run the expired entry cleaner
func New(maxEntries int, cleanupInterval time.Duration) *Cache {
	kin := maxEntries / 4
	if kin < 1 {
		kin = 1
	}

	cache := &Cache{
		maxEntries:      maxEntries,
		entries:         make(map[any]*list.Element),
		a1:              list.New(),
		a2:              list.New(),
		b:               list.New(),
		kin:             kin,
		cleanupInterval: cleanupInterval,
		stopChan:        make(chan struct{}),
	}
	if cache.cleanupInterval > 0 {
		go cache.startCleanup()
	}
	return cache
}

// Set inserts or updates the specified key-value pair with an expiration time
func (c *Cache) Set(key any, value any, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration int64
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	} else if c.defaultTTL > 0 {
		expiration = time.Now().Add(c.defaultTTL).UnixNano()
	}

	if ele, ok := c.entries[key]; ok {
		// Update existing entry
		ent := ele.Value.(*entry)
		ent.value = value
		ent.expiration = expiration
		// If in A1, move to front of A2
		if c.a1.Remove(ele) != nil {
			c.a2.PushFront(ent)
			c.entries[key] = c.a2.Front()
		} else {
			// Already in A2, move to front
			c.a2.MoveToFront(ele)
		}
		return
	}

	// New entry
	ent := &entry{key: key, value: value, expiration: expiration}
	c.entries[key] = c.a1.PushFront(ent)

	// Check if we need to evict
	if c.a1.Len()+c.a2.Len() > c.maxEntries {
		c.evict()
	}
}

// Get looks up a key's value from the cache
func (c *Cache) Get(key any) (value any, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, exists := c.entries[key]; exists {
		ent := ele.Value.(*entry)
		if ent.expiration > 0 && time.Now().UnixNano() > ent.expiration {
			// Expired, remove it
			c.removeElement(ele, common.ExpirationEvent)
			return nil, false
		}

		// Move from A1 to A2 if in A1
		if c.a1.Remove(ele) != nil {
			c.a2.PushFront(ent)
			c.entries[key] = c.a2.Front()
		} else {
			// Already in A2, move to front
			c.a2.MoveToFront(ele)
		}
		return ent.value, true
	}

	// Cache miss - check ghost queue
	if c.checkGhost(key) {
		// Was in B, don't add to cache (2Q policy)
		return nil, false
	}

	return nil, false
}

// Has returns true if the key exists in the cache
func (c *Cache) Has(key any) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ele, ok := c.entries[key]
	if !ok {
		return false
	}
	ent := ele.Value.(*entry)
	if ent.expiration > 0 && time.Now().UnixNano() > ent.expiration {
		return false
	}
	return true
}

// Keys returns a slice of the keys in the cache
func (c *Cache) Keys() []any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]any, 0, len(c.entries))
	now := time.Now().UnixNano()

	for key, ele := range c.entries {
		ent := ele.Value.(*entry)
		if ent.expiration == 0 || now <= ent.expiration {
			keys = append(keys, key)
		}
	}
	return keys
}

// Len returns the number of items in the cache
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	now := time.Now().UnixNano()

	for _, ele := range c.entries {
		ent := ele.Value.(*entry)
		if ent.expiration == 0 || now <= ent.expiration {
			count++
		}
	}
	return count
}

// Clear purges all key-value pairs from the cache
func (c *Cache) Clear() {
	// Note: This function assumes the caller has already acquired the mutex
	for key, ele := range c.entries {
		if c.onEvicted != nil {
			ent := ele.Value.(*entry)
			c.onEvicted(key, ent.value, common.ClearEvent)
		}
	}

	c.entries = make(map[any]*list.Element)
	c.a1.Init()
	c.a2.Init()
	c.b.Init()
}

// Close purges all key-value pairs from the cache and stop cleanup
func (c *Cache) Close() {
	// Stop cleanup goroutine first
	if c.stopChan != nil {
		close(c.stopChan)
		c.stopChan = nil
	}

	c.mu.Lock()
	c.closed = true
	c.Clear()
	c.mu.Unlock()
}

// SetEvictedFunc updates the eviction func
func (c *Cache) SetEvictedFunc(f common.EvictedFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEvicted = f
	return nil
}

// SetDefaultTTL sets the default TTL for cache entries
func (c *Cache) SetDefaultTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.defaultTTL = ttl
}

// evict implements the 2Q eviction policy
func (c *Cache) evict() {
	// First try to evict from A1 (FIFO)
	if c.a1.Len() > 0 {
		ele := c.a1.Back()
		c.a1.Remove(ele)
		ent := ele.Value.(*entry)
		delete(c.entries, ent.key)

		// Add to ghost queue B
		c.b.PushFront(ent)
		if c.b.Len() > c.maxEntries {
			c.b.Remove(c.b.Back())
		}

		if c.onEvicted != nil {
			c.onEvicted(ent.key, ent.value, common.EvictionEvent)
		}
		return
	}

	// If A1 is empty, evict from A2 (LRU)
	if c.a2.Len() > 0 {
		ele := c.a2.Back()
		c.a2.Remove(ele)
		ent := ele.Value.(*entry)
		delete(c.entries, ent.key)

		if c.onEvicted != nil {
			c.onEvicted(ent.key, ent.value, common.EvictionEvent)
		}
	}
}

// checkGhost checks if key exists in ghost queue B and removes it if found
func (c *Cache) checkGhost(key any) bool {
	for ele := c.b.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(*entry).key == key {
			c.b.Remove(ele)
			return true
		}
	}
	return false
}

// startCleanup starts the cleanup goroutine
func (c *Cache) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopChan:
			return
		}
	}
}

// cleanup removes expired entries
func (c *Cache) cleanup() {
	c.mu.RLock()
	closed := c.closed
	c.mu.RUnlock()

	if closed {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double check after locking
	if c.closed {
		return
	}

	now := time.Now().UnixNano()
	toRemove := make([]*list.Element, 0)

	// Check A1
	for ele := c.a1.Front(); ele != nil; ele = ele.Next() {
		ent := ele.Value.(*entry)
		if ent.expiration > 0 && now > ent.expiration {
			toRemove = append(toRemove, ele)
		}
	}

	// Check A2
	for ele := c.a2.Front(); ele != nil; ele = ele.Next() {
		ent := ele.Value.(*entry)
		if ent.expiration > 0 && now > ent.expiration {
			toRemove = append(toRemove, ele)
		}
	}

	for _, ele := range toRemove {
		c.removeElement(ele, common.ExpirationEvent)
	}

	if len(toRemove) > 0 {
		fmt.Printf("Cache: Running cleanup routine, removed %d expired entries\n", len(toRemove))
	}
}

// removeElement removes an element from the cache
func (c *Cache) removeElement(ele *list.Element, event int) {
	ent := ele.Value.(*entry)
	delete(c.entries, ent.key)

	// Remove from whichever list it's in
	if c.a1.Remove(ele) == nil {
		c.a2.Remove(ele)
	}

	if c.onEvicted != nil {
		c.onEvicted(ent.key, ent.value, event)
	}
}