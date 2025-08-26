package engine

import (
	"context"

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
	Type            int // LRU, LFU, ARC, RANDOM
	Size            int // size of the cache
	CleanupInterval int // cleanup interval in seconds
	TTL             int // time-to-live in seconds
	Partition       int // number of partitions for the cache
}
