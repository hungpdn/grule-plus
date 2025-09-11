package engine

import (
	"context"
	"testing"
)

func TestNewSingleEngine(t *testing.T) {
	cfg := Config{Type: LRU, Size: 10, CleanupInterval: 1, TTL: 1}
	se := NewSingleEngine(cfg)
	if se == nil {
		t.Fatalf("NewSingleEngine returned nil")
	}
	if se.cfg != cfg {
		t.Fatalf("Config not set correctly")
	}
}

func TestRemoveRule(t *testing.T) {
	se := NewSingleEngine(Config{})
	se.knowledgeLibraries["r1"] = nil
	se.RemoveRule("r1")
	if _, ok := se.knowledgeLibraries["r1"]; ok {
		t.Fatalf("RemoveRule did not remove rule")
	}
}

func TestDebug(t *testing.T) {
	se := NewSingleEngine(Config{})
	se.knowledgeLibraries["r1"] = nil
	se.localCache.Set("r1", nil, 0)
	dbg := se.Debug()
	if dbg["local_cache"] == nil || dbg["libraries"] == nil {
		t.Fatalf("Debug missing keys")
	}
}

func TestClose(t *testing.T) {
	se := NewSingleEngine(Config{})
	se.knowledgeLibraries["r1"] = nil
	se.localCache.Set("r1", nil, 0)
	se.Close()
	if len(se.knowledgeLibraries) != 0 {
		t.Fatalf("Close did not clear knowledgeLibraries")
	}
	if se.localCache.Len() != 0 {
		t.Fatalf("Close did not clear localCache")
	}
}

func TestContainsRule(t *testing.T) {
	se := NewSingleEngine(Config{})
	se.knowledgeLibraries["r1"] = nil
	se.localCache.Set("r1", nil, 0)
	if !se.ContainsRule("r1") {
		t.Fatalf("ContainsRule should return true for present rule")
	}
	if se.ContainsRule("r2") {
		t.Fatalf("ContainsRule should return false for absent rule")
	}
}

func TestAddRuleAndBuildRule(t *testing.T) {
	se := NewSingleEngine(Config{})
	// AddRule should add rule and cache
	statement := `rule DiscountRule "Apply discount" salience 10 { 
				when 
					DiscountFact.Amount > 100 
				then 
					DiscountFact.Discount = 10; }
				`
	err := se.AddRule("r1", statement, 0)
	if err != nil {
		t.Fatalf("AddRule error: %v", err)
	}
	if _, ok := se.knowledgeLibraries["r1"]; !ok {
		t.Fatalf("AddRule did not add rule to knowledgeLibraries")
	}
	if !se.localCache.Has("r1") {
		t.Fatalf("AddRule did not add rule to localCache")
	}
	// BuildRule should not overwrite existing rule
	err = se.BuildRule("r1", statement, 1)
	if err != nil {
		t.Fatalf("BuildRule error: %v", err)
	}
	if _, ok := se.knowledgeLibraries["r1"]; !ok {
		t.Fatalf("BuildRule did not add rule to knowledgeLibraries")
	}
}

func TestExecuteAndFetchMatching(t *testing.T) {
	se := NewSingleEngine(Config{})
	// Add a rule so Execute/FetchMatching can find it
	se.knowledgeLibraries["r1"] = nil
	// Should error because knowledgeLibrary is nil
	err := se.Execute(context.Background(), "r1", struct{}{})
	if err == nil {
		t.Fatalf("Execute should error if knowledgeLibrary is nil")
	}
	_, err = se.FetchMatching(context.Background(), "r1", struct{}{})
	if err == nil {
		t.Fatalf("FetchMatching should error if knowledgeLibrary is nil")
	}
	// Should error for missing rule
	err = se.Execute(context.Background(), "r2", struct{}{})
	if err == nil {
		t.Fatalf("Execute should error for missing rule")
	}
}
