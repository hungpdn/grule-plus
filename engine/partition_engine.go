package engine

import (
	"context"
	"runtime"

	"github.com/hungpdn/grule-plus/internal/utils"
	"github.com/hyperjumptech/grule-rule-engine/ast"
)

type HashFunc = func(rule string) int

type partitionEngine struct {
	cfg       Config
	partition int
	engines   map[int]*singleEngine
	hash      HashFunc
}

func NewPartitionEngine(cfg Config, hashFunc HashFunc) *partitionEngine {
	partition := utils.MaxInt(runtime.NumCPU(), cfg.Partition)
	partitionEngine := &partitionEngine{
		cfg:       cfg,
		partition: partition,
		engines:   make(map[int]*singleEngine),
		hash:      hashFunc,
	}

	if hashFunc == nil {
		partitionEngine.hash = func(rule string) int {
			random := utils.HashStringToRange(rule, 1, int64(partition))
			return int(random)
		}
	}

	sizeE := cfg.Size / partition
	for i := 0; i < partition; i++ {
		cfgE := Config{
			Type:            cfg.Type,
			Size:            sizeE,
			CleanupInterval: cfg.CleanupInterval,
			TTL:             cfg.TTL,
		}
		partitionEngine.engines[i+1] = NewSingleEngine(cfgE)
	}

	return partitionEngine
}

func (s *partitionEngine) Execute(ctx context.Context, rule string, fact any) error {
	return s.engines[s.hash(rule)].Execute(ctx, rule, fact)
}

func (s *partitionEngine) FetchMatching(ctx context.Context, rule string, fact any) ([]*ast.RuleEntry, error) {
	return s.engines[s.hash(rule)].FetchMatching(ctx, rule, fact)
}

func (s *partitionEngine) AddRule(rule, statement string, duration int64) error {
	return s.engines[s.hash(rule)].AddRule(rule, statement, duration)
}

func (s *partitionEngine) BuildRule(rule, statement string, duration int64) error {
	return s.engines[s.hash(rule)].BuildRule(rule, statement, duration)
}

func (s *partitionEngine) ContainsRule(rule string) bool {
	return s.engines[s.hash(rule)].ContainsRule(rule)
}

func (s *partitionEngine) Debug() map[string]any {
	engines := make(map[int]map[string]any)
	for k, v := range s.engines {
		if v != nil {
			engines[k] = v.Debug()
			delete(engines[k], "stats")
		}
	}
	return map[string]any{
		"partition_config": s.cfg,
		"engines":          engines,
		"stats":            utils.GetStats(),
	}
}

func (s *partitionEngine) Close() {
	for _, v := range s.engines {
		if v != nil {
			v.Close()
		}
	}
}
