package cache

import (
	"time"

	"github.com/hungpdn/grule-plus/internal/cache/arc"
	"github.com/hungpdn/grule-plus/internal/cache/common"
	"github.com/hungpdn/grule-plus/internal/cache/lfu"
	"github.com/hungpdn/grule-plus/internal/cache/lru"
	"github.com/hungpdn/grule-plus/internal/cache/random"
	"github.com/hungpdn/grule-plus/internal/cache/twoq"
)

// CacheType defines the type of cache to be used.
const (
	LRU = iota
	LFU
	ARC
	TWOQ
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
	// SetDefaultTTL sets the default TTL for cache entries
	SetDefaultTTL(ttl time.Duration)
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
	factories := map[int]func() ICache{
		LRU:    func() ICache { return lru.New(config.Size, config.CleanupInterval) },
		LFU:    func() ICache { return lfu.New(config.Size, config.CleanupInterval) },
		ARC:    func() ICache { return arc.New(config.Size, config.CleanupInterval) },
		TWOQ:   func() ICache { return twoq.New(config.Size, config.CleanupInterval) },
		RANDOM: func() ICache { return random.New(config.Size, config.CleanupInterval) },
	}

	factory, ok := factories[config.Type]
	if !ok {
		panic("unknown type")
	}

	cache := factory()
	if config.EvictedFunc != nil {
		_ = cache.SetEvictedFunc(config.EvictedFunc)
	}
	if config.DefaultTTL > 0 {
		cache.SetDefaultTTL(config.DefaultTTL)
	}
	return cache
}
