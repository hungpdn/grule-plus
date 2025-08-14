package engine

import (
	"context"

	"github.com/hyperjumptech/grule-rule-engine/ast"
)

// IGruleEngine defines the interface for the Grule rule engine.
type IGruleEngine interface {
	Execute(ctx context.Context, rule string, fact any) error
	FetchMatching(ctx context.Context, rule string, fact any) ([]*ast.RuleEntry, error)
	AddRule(rule, statement string, duration int64) error
	BuildRule(rule, statement string, duration int64) error
	ContainsRule(rule string) bool
	Debug() map[string]any
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
