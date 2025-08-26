// lfu implements an LFU cache.
package lfu

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/hungpdn/grule-plus/internal/cache/common"
)

// entry holds a key-value item, its frequency count, and expiration.
type entry struct {
	key        any
	value      any
	freq       int
	expiration int64
	node       *list.Element
}

// Cache is a fixed-maxEntries in-memory cache with LFU eviction and per-item TTL.
type Cache struct {
	maxEntries      int
	entries         map[any]*entry
	freqList        map[int]*list.List // maps frequency -> list of entries
	minFreq         int
	mu              sync.RWMutex
	onEvicted       common.EvictedFunc
	defaultTTL      time.Duration
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// NewLFUCache creates an Cache with given maxEntries and starts a background
// cleanup goroutine that runs every cleanupInterval.
func New(maxEntries int, cleanupInterval time.Duration) *Cache {
	cache := &Cache{
		maxEntries:      maxEntries,
		entries:         make(map[any]*entry),
		freqList:        make(map[int]*list.List),
		minFreq:         0,
		cleanupInterval: cleanupInterval,
		stopCleanup:     make(chan struct{}),
	}
	// Start background cleanup of expired entries
	if cache.cleanupInterval > 0 {
		go cache.startCleanup()
	}
	return cache
}

// NewWithEvictionFunc creates an Cache with given maxEntries and eviction callback function,
// and starts a background cleanup goroutine that runs every cleanupInterval.
func NewWithEvictionFunc(maxEntries int, cleanupInterval time.Duration, f common.EvictedFunc) *Cache {
	c := New(maxEntries, cleanupInterval)
	c.onEvicted = f
	return c
}

// SetEvictedFunc sets the eviction callback function.
func (c *Cache) SetEvictedFunc(f common.EvictedFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.onEvicted != nil {
		return fmt.Errorf("lfu cache eviction function is already set")
	}
	c.onEvicted = f
	return nil
}

// SetDefaultTTL sets the default TTL for items. A zero duration means no default TTL.
func (c *Cache) SetDefaultTTL(ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.defaultTTL = ttl
	return nil
}

// Set inserts or updates a key with the given value and TTL (in seconds).
// If the cache is at maxEntries, it evicts the least-frequently used item.
func (c *Cache) Set(key, value any, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

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

	// Update existing entry
	if entry, ok := c.entries[key]; ok {
		entry.value = value
		entry.expiration = expiration

		// Increase frequency
		c.incrementFrequency(entry)
		return
	}

	// Evict if necessary
	if len(c.entries) >= c.maxEntries {
		c.evict()
	}

	// Insert new entry at frequency 1
	entry := &entry{
		key:        key,
		value:      value,
		freq:       1,
		expiration: expiration,
	}

	c.entries[key] = entry
	if c.freqList[1] == nil {
		c.freqList[1] = list.New()
	}
	entry.node = c.freqList[1].PushBack(entry)
	c.minFreq = 1
}

// Get retrieves the value for a key, returning (nil,false) if not found or expired.
// On a hit, it increments the access frequency (LFU policy).
func (c *Cache) Get(key any) (value any, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		return
	}
	// Check expiration
	if entry.expiration > 0 && time.Now().UnixNano() > entry.expiration {
		// Remove expired entry
		c.removeEntry(entry, common.ExpirationEvent)
		delete(c.entries, key)
		return
	}
	// Increment frequency and return value
	c.incrementFrequency(entry)
	return entry.value, true
}

// incrementFrequency moves an entry from freq -> freq+1 list.
func (c *Cache) incrementFrequency(entry *entry) {
	freq := entry.freq
	// Remove from current frequency list
	c.freqList[freq].Remove(entry.node)
	if c.freqList[freq].Len() == 0 {
		delete(c.freqList, freq)
		if c.minFreq == freq {
			c.minFreq++
		}
	}
	// Add to next frequency list
	entry.freq++
	if c.freqList[entry.freq] == nil {
		c.freqList[entry.freq] = list.New()
	}
	entry.node = c.freqList[entry.freq].PushBack(entry)
}

// evict removes the least frequently used entry (and oldest among ties).
func (c *Cache) evict() {
	// Find list of entries with minFreq
	list := c.freqList[c.minFreq]
	if list == nil {
		return
	}
	// Remove oldest entry from this list
	oldest := list.Front().Value.(*entry)
	list.Remove(list.Front())
	delete(c.entries, oldest.key)
	if list.Len() == 0 {
		delete(c.freqList, c.minFreq)
		// next minFreq will reset on new insert
	}
	if c.onEvicted != nil {
		c.onEvicted(oldest.key, oldest.value, common.EvictionEvent)
	}
}

// removeEntry removes an entry from its frequency list (used on expiration).
func (c *Cache) removeEntry(entry *entry, event int) {
	list := c.freqList[entry.freq]
	if list != nil {
		list.Remove(entry.node)
		if list.Len() == 0 {
			delete(c.freqList, entry.freq)
			if entry.freq == c.minFreq {
				c.minFreq = 1 // reset; will be recomputed on next insert
			}
		}
		if c.onEvicted != nil {
			c.onEvicted(entry.key, entry.value, event)
		}
	}
}

// Has checks if a key exists and is not expired, without updating its frequency.
func (c *Cache) Has(key any) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.entries == nil {
		return false
	}

	if entry, hit := c.entries[key]; hit {
		if entry.expiration > 0 && time.Now().UnixNano() > entry.expiration {
			return false
		}
		return true
	}

	return false
}

// Delete removes a key from the cache. Returns true if the key was present.
func (c *Cache) Delete(key any) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, hit := c.entries[key]; hit {
		c.removeEntry(ele, common.DeleteEvent)
		delete(c.entries, key)
		return true
	}

	return false
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.entries == nil {
		return 0
	}
	return len(c.entries)
}

// Clear removes all items from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.onEvicted != nil {
		for _, entry := range c.entries {
			c.onEvicted(entry.key, entry.value, common.ClearEvent)
		}
	}

	c.entries = nil
	c.freqList = nil
	c.minFreq = 0
}

// startCleanup runs in background to delete all expired entries periodically.
// This uses a ticker to scan the map and remove outdated entries:contentReference[oaicite:2]{index=2}.
func (c *Cache) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("Cache: Running cleanup routine...")
			c.cleanupExpiredEntries()
		case <-c.stopCleanup:
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
	for key, entry := range c.entries {
		if entry.expiration > 0 && now > entry.expiration {
			c.removeEntry(entry, common.ExpirationEvent)
			delete(c.entries, key)
		}
	}
}

func (c *Cache) StopCleanup() {
	if c.stopCleanup != nil {
		close(c.stopCleanup)
	}
}

// Keys returns a slice of all keys in the cache.
func (c *Cache) Keys() []any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]interface{}, 0, len(c.entries))
	for k := range c.entries {
		keys = append(keys, k)
	}
	return keys
}

// Close stops the background cleanup goroutine.
func (c *Cache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.StopCleanup()

	if c.onEvicted != nil {
		for _, entry := range c.entries {
			c.onEvicted(entry.key, entry.value, common.ClearEvent)
		}
	}

	c.entries = nil
	c.freqList = nil
	c.minFreq = 0
}
