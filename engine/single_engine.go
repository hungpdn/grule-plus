package engine

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/hungpdn/grule-plus/internal/cache"
	"github.com/hungpdn/grule-plus/internal/cache/common"
	"github.com/hungpdn/grule-plus/internal/logger"
	"github.com/hungpdn/grule-plus/internal/utils"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

const (
	DiscountFact   = "DiscountFact"
	LibraryName    = "library"
	LibraryVersion = "0.0.1"
)

type singleEngine struct {
	cfg                Config
	engine             *engine.GruleEngine
	knowledgeLibraries map[string]*ast.KnowledgeLibrary
	localCache         cache.ICache
	mu                 sync.RWMutex // protect knowledgeLibraries
}

func NewSingleEngine(cfg Config) *singleEngine {

	singleEngine := &singleEngine{
		cfg:                cfg,
		engine:             engine.NewGruleEngine(),
		knowledgeLibraries: make(map[string]*ast.KnowledgeLibrary),
	}

	localCache := cache.New(cache.Config{
		Type:            cfg.Type,
		Size:            cfg.Size,
		CleanupInterval: time.Duration(cfg.CleanupInterval) * time.Second,
		DefaultTTL:      time.Duration(cfg.TTL) * time.Second,
	})
	localCache.SetEvictedFunc(func(key, value any, event int) {
		go func() {
			switch event {
			case common.ExpirationEvent, common.EvictionEvent:
				singleEngine.RemoveRule(key.(string))
			default:
				// do nothing
			}
		}()
	})

	singleEngine.localCache = localCache
	return singleEngine
}

func (s *singleEngine) RemoveRule(rule string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.knowledgeLibraries, rule)
}

func (s *singleEngine) Debug() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rulesInLocalCache := s.localCache.Keys()
	rulesInLibraries := make([]string, 0, len(s.knowledgeLibraries))
	for rule := range s.knowledgeLibraries {
		rulesInLibraries = append(rulesInLibraries, rule)
	}

	return map[string]any{
		"local_cache": map[string]any{
			"config": s.cfg,
			"rules":  rulesInLocalCache,
			"len":    len(rulesInLocalCache),
		},
		"libraries": map[string]any{
			"rules": rulesInLibraries,
			"len":   len(s.knowledgeLibraries),
		},
		"stats": utils.GetStats(),
	}
}

func (s *singleEngine) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.knowledgeLibraries = make(map[string]*ast.KnowledgeLibrary)
	s.localCache.Clear()
	runtime.GC()
}

func (s *singleEngine) ContainsRule(rule string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.localCache.Get(rule)
	_, ok := s.knowledgeLibraries[rule]

	return ok
}

// Note: must use with Mutex
func (s *singleEngine) addRule(rule, statement string) error {

	library := ast.NewKnowledgeLibrary()
	rb := builder.NewRuleBuilder(library)
	err := rb.BuildRuleFromResource(LibraryName, LibraryVersion, pkg.NewBytesResource([]byte(statement)))
	if err != nil {
		return err
	}

	s.knowledgeLibraries[rule] = library

	return nil
}

// AddRule add rule if not exists, update if exists
func (s *singleEngine) AddRule(rule, statement string, duration int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.addRule(rule, statement)
	if err != nil {
		return err
	}

	s.localCache.Set(rule, nil, time.Duration(duration))

	return nil
}

// BuildRule add rule if not exists
func (s *singleEngine) BuildRule(rule, statement string, duration int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.knowledgeLibraries[rule]; !ok {
		err := s.addRule(rule, statement)
		if err != nil {
			return err
		}
	}
	s.localCache.Set(rule, nil, time.Duration(duration))

	return nil
}

// Note: must rules exists
func (s *singleEngine) Execute(ctx context.Context, rule string, fact any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dataContext := ast.NewDataContext()
	if err := dataContext.Add(DiscountFact, fact); err != nil {
		logger.WithContext(ctx).Errorf("[singleEngine][Execute] add fact %v has error : %v", fact, err)
		return err
	}

	knowledgeLibrary, ok := s.knowledgeLibraries[rule]
	if knowledgeLibrary == nil {
		logger.WithContext(ctx).Errorf("[singleEngine][Execute] knowledge library empty, %v cache hit %v", rule, ok)
		return errors.New("knowledge library empty")
	}

	kb, err := knowledgeLibrary.NewKnowledgeBaseInstance(LibraryName, LibraryVersion)
	if kb == nil {
		logger.WithContext(ctx).Errorf("[singleEngine][Execute] knowledge base instance empty")
		return errors.New("knowledge base instance empty")
	}
	if err != nil {
		logger.WithContext(ctx).Errorf("[singleEngine][Execute] knowledge base instance error %v", err)
		return err
	}

	err = s.engine.Execute(dataContext, kb)
	if err != nil {
		logger.WithContext(ctx).Errorf("[singleEngine][Execute] execute data context fact %v has error : %v", fact, err)
		return err
	}

	return nil
}

// Note: must rules exists
func (s *singleEngine) FetchMatching(ctx context.Context, rule string, fact any) ([]*ast.RuleEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dataContext := ast.NewDataContext()
	if err := dataContext.Add(DiscountFact, fact); err != nil {
		logger.WithContext(ctx).Errorf("[singleEngine][FetchMatching] add fact %v has error : %v", fact, err)
		return nil, err
	}

	knowledgeLibrary, ok := s.knowledgeLibraries[rule]
	if knowledgeLibrary == nil {
		logger.WithContext(ctx).Errorf("[singleEngine][FetchMatching] knowledge library empty, %v cache hit %v", rule, ok)
		return nil, errors.New("knowledge library empty")
	}

	kb, err := knowledgeLibrary.NewKnowledgeBaseInstance(LibraryName, LibraryVersion)
	if kb == nil {
		logger.WithContext(ctx).Errorf("[singleEngine][FetchMatching] knowledge base instance empty")
		return nil, errors.New("knowledge base instance empty")
	}
	if err != nil {
		logger.WithContext(ctx).Errorf("[singleEngine][FetchMatching] knowledge base instance error %v", err)
		return nil, err
	}

	ruleEntries, err := s.engine.FetchMatchingRules(dataContext, kb)
	if err != nil {
		logger.WithContext(ctx).Errorf("[singleEngine][FetchMatching] execute data context fact %v has error : %v", fact, err)
		return nil, err
	}

	return ruleEntries, nil
}
