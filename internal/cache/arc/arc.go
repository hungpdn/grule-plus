// arc implements an ARC cache.
package arc

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/hungpdn/grule-plus/internal/cache/common"
)

// Cache is an ARC cache structure
type Cache struct {
	maxEntries int                   // The maximum number of cache entries before an entry is evicted, zero means no limit
	entries    map[any]*list.Element // Map for quick access to cache entries
	t1         *list.List            // T1: recently accessed items
	t2         *list.List            // T2: frequently accessed items
	b1         *list.List            // B1: ghost entries evicted from T1
	b2         *list.List            // B2: ghost entries evicted from T2
	p          int                   // Target size for T1, adapts based on access patterns
	mu         sync.RWMutex          // Mutex to ensure concurrent access safety
	onEvicted  common.EvictedFunc    // OnEvicted optionally specifies a callback function to be executed when an entry is purged from the cache
	// cleanup
	defaultTTL      time.Duration // default TTL for item expire
	cleanupInterval time.Duration // how often to run the expired entry cleaner
	stopChan        chan struct{} // Channel to stop cleanup goroutine
	closed          bool          // Flag to indicate if cache is closed
}

// entry represents an entry in the ARC cache
type entry struct {
	key        any
	value      any
	expiration int64 // Unix timestamp (nanoseconds) when the item expires, 0 means never expires
}

// New creates a new ARC cache
// maxEntries: the maximum number of cache entries before an entry is evicted, zero means no limit
// cleanupInterval: how often to run the expired entry cleaner
func New(maxEntries int, cleanupInterval time.Duration) *Cache {
	cache := &Cache{
		maxEntries:      maxEntries,
		entries:         make(map[any]*list.Element),
		t1:              list.New(),
		t2:              list.New(),
		b1:              list.New(),
		b2:              list.New(),
		p:               0, // Start with p = 0
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
		// Move to T2 if in T1, or move to front of T2 if already in T2
		if c.t1.Remove(ele) != nil {
			c.t2.PushFront(ent)
			c.entries[key] = c.t2.Front()
		} else {
			c.t2.MoveToFront(ele)
		}
		return
	}

	// New entry - check ghost lists first
	ent := &entry{key: key, value: value, expiration: expiration}

	// Check if in B1 or B2 (ghost entries)
	inB1 := c.checkGhost(c.b1, key)
	inB2 := c.checkGhost(c.b2, key)

	if inB1 {
		// Hit in B1, increase p
		if c.b1.Len() > 0 {
			c.p = min(c.p+max(1, c.b2.Len()/c.b1.Len()), c.maxEntries)
		} else {
			c.p = min(c.p+1, c.maxEntries)
		}
		c.replace(key) // This might be redundant, but follows ARC
	} else if inB2 {
		// Hit in B2, decrease p
		if c.b2.Len() > 0 {
			c.p = max(c.p-max(1, c.b1.Len()/c.b2.Len()), 0)
		} else {
			c.p = max(c.p-1, 0)
		}
		c.replace(key)
	}

	// Add to T1
	c.entries[key] = c.t1.PushFront(ent)

	// Check if we need to evict
	if c.t1.Len()+c.t2.Len() > c.maxEntries {
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
		// Move to T2
		if c.t1.Remove(ele) != nil {
			c.t2.PushFront(ent)
			c.entries[key] = c.t2.Front()
		} else {
			c.t2.MoveToFront(ele)
		}
		return ent.value, true
	}

	// Miss - check ghost lists
	inB1 := c.checkGhost(c.b1, key)
	inB2 := c.checkGhost(c.b2, key)

	if inB1 {
		if c.b1.Len() > 0 {
			c.p = min(c.p+max(1, c.b2.Len()/c.b1.Len()), c.maxEntries)
		} else {
			c.p = min(c.p+1, c.maxEntries)
		}
	} else if inB2 {
		if c.b2.Len() > 0 {
			c.p = max(c.p-max(1, c.b1.Len()/c.b2.Len()), 0)
		} else {
			c.p = max(c.p-1, 0)
		}
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
	c.t1.Init()
	c.t2.Init()
	c.b1.Init()
	c.b2.Init()
	c.p = 0
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

// evict implements the ARC eviction policy
func (c *Cache) evict() {
	if c.t1.Len() >= max(1, c.p) {
		// Evict from T1, add to B1
		ele := c.t1.Back()
		c.t1.Remove(ele)
		ent := ele.Value.(*entry)
		delete(c.entries, ent.key)
		c.b1.PushFront(ent)
		if c.b1.Len() > c.maxEntries {
			c.b1.Remove(c.b1.Back())
		}
		if c.onEvicted != nil {
			c.onEvicted(ent.key, ent.value, common.EvictionEvent)
		}
	} else {
		// Evict from T2, add to B2
		ele := c.t2.Back()
		c.t2.Remove(ele)
		ent := ele.Value.(*entry)
		delete(c.entries, ent.key)
		c.b2.PushFront(ent)
		if c.b2.Len() > c.maxEntries {
			c.b2.Remove(c.b2.Back())
		}
		if c.onEvicted != nil {
			c.onEvicted(ent.key, ent.value, common.EvictionEvent)
		}
	}
}

// replace implements the ARC replace policy (simplified)
func (c *Cache) replace(key any) {
	// ARC replace: if T1 is too big, evict from T1, else evict from T2
	if c.t1.Len() >= max(1, c.p) {
		// Evict from T1, add to B1
		ele := c.t1.Back()
		c.t1.Remove(ele)
		ent := ele.Value.(*entry)
		delete(c.entries, ent.key)
		c.b1.PushFront(ent)
		if c.b1.Len() > c.maxEntries {
			c.b1.Remove(c.b1.Back())
		}
		if c.onEvicted != nil {
			c.onEvicted(ent.key, ent.value, common.EvictionEvent)
		}
	} else {
		// Evict from T2, add to B2
		ele := c.t2.Back()
		c.t2.Remove(ele)
		ent := ele.Value.(*entry)
		delete(c.entries, ent.key)
		c.b2.PushFront(ent)
		if c.b2.Len() > c.maxEntries {
			c.b2.Remove(c.b2.Back())
		}
		if c.onEvicted != nil {
			c.onEvicted(ent.key, ent.value, common.EvictionEvent)
		}
	}
}

// checkGhost checks if key exists in ghost list and removes it if found
func (c *Cache) checkGhost(list *list.List, key any) bool {
	for ele := list.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(*entry).key == key {
			list.Remove(ele)
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

	// Check T1
	for ele := c.t1.Front(); ele != nil; ele = ele.Next() {
		ent := ele.Value.(*entry)
		if ent.expiration > 0 && now > ent.expiration {
			toRemove = append(toRemove, ele)
		}
	}

	// Check T2
	for ele := c.t2.Front(); ele != nil; ele = ele.Next() {
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
	if c.t1.Remove(ele) == nil {
		c.t2.Remove(ele)
	}

	if c.onEvicted != nil {
		c.onEvicted(ent.key, ent.value, event)
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
