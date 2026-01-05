package engine

import (
	"context"

	"github.com/hungpdn/grule-plus/internal/cache"
	"github.com/hyperjumptech/grule-rule-engine/ast"
)

// IGruleEngine defines the interface for the Grule rule engine.
type IGruleEngine interface {
	// Execute runs the rule engine with the given context, rule name, and fact.
	Execute(ctx context.Context, rule string, fact any) error
	// FetchMatching retrieves rules matching the given rule name and fact.
	FetchMatching(ctx context.Context, rule string, fact any) ([]*ast.RuleEntry, error)
	// AddRule adds a new rule to the engine with an optional duration for caching.
	AddRule(rule, statement string, duration int64) error
	// BuildRule builds or updates an existing rule in the engine with an optional duration for caching.
	BuildRule(rule, statement string, duration int64) error
	// ContainsRule checks if a rule exists in the engine.
	ContainsRule(rule string) bool
	// Debug provides internal state information for debugging purposes.
	Debug() map[string]any
	// Close cleans up resources used by the engine.
	Close()
}

// Config holds the configuration for the Grule engine.
type Config struct {
	Type            CacheType // type of cache: lru, lfu, arc, random
	Size            int       // size of the cache, 0 means unlimited
	CleanupInterval int       // cleanup interval in seconds, 0 means no cleanup
	TTL             int       // time-to-live in seconds, 0 means no expiration
	Partition       int       // number of partitions for the engine
	FactName        string    // name of the fact to be used in rules, default is "Fact"
}

// CacheType represents the type of cache to be used.
type CacheType string

const (
	LRU    CacheType = "lru"
	LFU    CacheType = "lfu"
	ARC    CacheType = "arc"
	TWOQ   CacheType = "twoq"
	RANDOM CacheType = "random"
)

// GetCacheType returns the corresponding cache type constant.
func (c Config) GetCacheType() int {
	switch c.Type {
	case LRU:
		return cache.LRU
	case LFU:
		return cache.LFU
	case ARC:
		return cache.ARC
	case TWOQ:
		return cache.TWOQ
	case RANDOM:
		return cache.RANDOM
	default:
		return cache.LRU
	}
}

// GetFactName returns the configured fact name or a default value if not set.
func (c Config) GetFactName() string {
	if c.FactName == "" {
		return "Fact"
	}
	return c.FactName
}
