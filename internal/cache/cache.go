package cache

import (
	"time"

	"github.com/hungpdn/grule-plus/internal/cache/common"
	"github.com/hungpdn/grule-plus/internal/cache/lfu"
	"github.com/hungpdn/grule-plus/internal/cache/lru"
)

// CacheType defines the type of cache to be used.
const (
	LRU = iota
	LFU
	ARC
	RANDOM
)

// ICache defines the interface for a cache system.
type ICache interface {
	// Set inserts or updates the specified key-value pair with an expiration time
	Set(key any, value any, duration time.Duration)
	// Get looks up a key's value from the cache
	Get(key any) (value any, ok bool)
	// Has returns true if the key exists in the cache
	Has(key any) bool
	// Keys returns a slice of the keys in the cache
	Keys() []any
	// Len returns the number of items in the cache
	Len() int
	// Clear purges all key-value pairs from the cache
	Clear()
	// Close purges all key-value pairs from the cache and stop cleanup
	Close()
	// SetEvictedFunc updates the eviction func
	SetEvictedFunc(f common.EvictedFunc) error
}

// Config holds the configuration for the cache.
type Config struct {
	Type            int
	Size            int
	CleanupInterval time.Duration
	DefaultTTL      time.Duration
	EvictedFunc     common.EvictedFunc
}

// New creates a new cache instance based on the provided configuration.
func New(config Config) ICache {
	switch config.Type {
	case LRU:
		cache := lru.New(config.Size, config.CleanupInterval)
		if config.EvictedFunc != nil {
			_ = cache.SetEvictedFunc(config.EvictedFunc)
		}
		if config.DefaultTTL > 0 {
			cache.SetDefaultTTL(config.DefaultTTL)
		}
		return cache
	case LFU:
		cache := lfu.New(config.Size, config.CleanupInterval)
		if config.EvictedFunc != nil {
			_ = cache.SetEvictedFunc(config.EvictedFunc)
		}
		if config.DefaultTTL > 0 {
			cache.SetDefaultTTL(config.DefaultTTL)
		}
		return cache
	default:
		panic("unknown type")
	}
}
