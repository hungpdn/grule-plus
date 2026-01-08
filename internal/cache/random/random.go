// random implements a random eviction cache
package random

import (
	"math/rand"
	"sync"
	"time"

	"github.com/hungpdn/grule-plus/internal/cache/common"
)

// Cache is a random eviction cache structure
type Cache struct {
	maxEntries int                // The maximum number of cache entries before an entry is evicted, zero means no limit
	entries    map[any]*entry     // Map for quick access to cache entries
	keys       []any              // Slice of keys for random selection
	mu         sync.RWMutex       // Mutex to ensure concurrent access safety
	onEvicted  common.EvictedFunc // OnEvicted optionally specifies a callback function to be executed when an entry is purged from the cache
	// cleanup
	defaultTTL      time.Duration // default TTL for item expire
	cleanupInterval time.Duration // how often to run the expired entry cleaner
	stopChan        chan struct{} // Channel to stop cleanup goroutine
}

// entry represents an entry in the random cache
type entry struct {
	key        any
	value      any
	expiration int64 // Unix timestamp (nanoseconds) when the item expires, 0 means never expires
}

// New creates a new random eviction cache
func New(maxEntries int, cleanupInterval time.Duration) *Cache {
	cache := &Cache{
		maxEntries:      maxEntries,
		entries:         make(map[any]*entry),
		keys:            make([]any, 0),
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

	if ent, exists := c.entries[key]; exists {
		// Update existing entry
		ent.value = value
		ent.expiration = expiration
	} else {
		// Add new entry
		ent := &entry{
			key:        key,
			value:      value,
			expiration: expiration,
		}
		c.entries[key] = ent
		c.keys = append(c.keys, key)

		// Evict if over capacity
		if c.maxEntries > 0 && len(c.entries) > c.maxEntries {
			c.evictRandom()
		}
	}
}

// Get looks up a key's value from the cache
func (c *Cache) Get(key any) (value any, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if ent, exists := c.entries[key]; exists {
		if ent.expiration > 0 && time.Now().UnixNano() > ent.expiration {
			return nil, false
		}
		return ent.value, true
	}
	return nil, false
}

// Has returns true if the key exists in the cache
func (c *Cache) Has(key any) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if ent, exists := c.entries[key]; exists {
		if ent.expiration > 0 && time.Now().UnixNano() > ent.expiration {
			return false
		}
		return true
	}
	return false
}

// Keys returns a slice of the keys in the cache
func (c *Cache) Keys() []any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]any, 0, len(c.entries))
	now := time.Now().UnixNano()

	for _, key := range c.keys {
		if ent, exists := c.entries[key]; exists {
			if ent.expiration == 0 || now <= ent.expiration {
				keys = append(keys, key)
			}
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

	for _, ent := range c.entries {
		if ent.expiration == 0 || now <= ent.expiration {
			count++
		}
	}
	return count
}

// Clear purges all key-value pairs from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.onEvicted != nil {
		for _, ent := range c.entries {
			c.onEvicted(ent.key, ent.value, common.ClearEvent)
		}
	}

	c.entries = make(map[any]*entry)
	c.keys = make([]any, 0)
}

// Close purges all key-value pairs from the cache and stop cleanup
func (c *Cache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stopCleanup()

	// Clear all entries
	c.Clear()
}

// stopCleanup stops the cleanup goroutine
func (c *Cache) stopCleanup() {
	if c.cleanupInterval > 0 && c.stopChan != nil {
		close(c.stopChan)
		c.stopChan = nil
	}
}

// StopCleanup stops the cleanup goroutine (for testing)
func (c *Cache) StopCleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stopCleanup()
}

// SetEvictedFunc updates the eviction callback function
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

// evictRandom randomly evicts one entry from the cache
func (c *Cache) evictRandom() {
	if len(c.keys) == 0 {
		return
	}

	// Pick a random key
	randomIndex := rand.Intn(len(c.keys))
	keyToEvict := c.keys[randomIndex]

	if ent, exists := c.entries[keyToEvict]; exists {
		// Call eviction callback if set
		if c.onEvicted != nil {
			c.onEvicted(ent.key, ent.value, common.EvictionEvent)
		}

		// Remove from map
		delete(c.entries, keyToEvict)

		// Remove from keys slice (swap with last element for efficiency)
		c.keys[randomIndex] = c.keys[len(c.keys)-1]
		c.keys = c.keys[:len(c.keys)-1]
	}
}

// startCleanup starts the cleanup goroutine that periodically removes expired entries
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

// cleanup removes expired entries from the cache
func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixNano()
	expiredKeys := make([]any, 0)

	// Find expired keys
	for key, ent := range c.entries {
		if ent.expiration > 0 && now > ent.expiration {
			expiredKeys = append(expiredKeys, key)
		}
	}

	// Remove expired entries
	for _, key := range expiredKeys {
		if ent, exists := c.entries[key]; exists {
			if c.onEvicted != nil {
				c.onEvicted(ent.key, ent.value, common.ExpirationEvent)
			}
			delete(c.entries, key)

			// Remove from keys slice
			for i, k := range c.keys {
				if k == key {
					c.keys = append(c.keys[:i], c.keys[i+1:]...)
					break
				}
			}
		}
	}
}
