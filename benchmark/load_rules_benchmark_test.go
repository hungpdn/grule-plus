package benchmark

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hungpdn/grule-plus/engine"
	"github.com/hungpdn/grule-plus/internal/cache"
)

type DiscountFact struct {
	Amount   int
	Discount int
}

// BenchmarkCacheTypes benchmarks different cache types with various workloads
func BenchmarkCacheTypes(b *testing.B) {
	cacheTypes := []struct {
		name string
		typ  int
	}{
		{"LRU", cache.LRU},
		{"LFU", cache.LFU},
		{"ARC", cache.ARC},
		{"TWOQ", cache.TWOQ},
	}

	sizes := []int{100, 1000, 10000}
	workloads := []string{"read_heavy", "write_heavy", "mixed"}

	for _, cacheType := range cacheTypes {
		for _, size := range sizes {
			for _, workload := range workloads {
				b.Run(fmt.Sprintf("%s_Size%d_%s", cacheType.name, size, workload), func(b *testing.B) {
					c := cache.New(cache.Config{
						Type:            cacheType.typ,
						Size:            size,
						CleanupInterval: time.Minute,
						DefaultTTL:      time.Hour,
					})
					defer c.Close()

					b.ResetTimer()
					runCacheWorkload(b, c, workload, size)
				})
			}
		}
	}
}

func runCacheWorkload(b *testing.B, c cache.ICache, workload string, size int) {
	switch workload {
	case "read_heavy":
		// Pre-populate cache
		for i := 0; i < size; i++ {
			c.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i), 0)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key%d", rand.Intn(size))
			c.Get(key)
		}
	case "write_heavy":
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key%d", i%size)
			value := fmt.Sprintf("value%d", i)
			c.Set(key, value, 0)
		}
	case "mixed":
		// Pre-populate half the cache
		for i := 0; i < size/2; i++ {
			c.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i), 0)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if i%3 == 0 {
				// Write operation
				key := fmt.Sprintf("key%d", rand.Intn(size))
				value := fmt.Sprintf("value%d", i)
				c.Set(key, value, 0)
			} else {
				// Read operation
				key := fmt.Sprintf("key%d", rand.Intn(size/2))
				c.Get(key)
			}
		}
	}
}

// BenchmarkEngineCacheTypes benchmarks engine performance with different cache types
func BenchmarkEngineCacheTypes(b *testing.B) {
	cacheTypes := []engine.CacheType{
		engine.LRU,
		engine.LFU,
		engine.ARC,
		engine.TWOQ,
	}

	for _, cacheType := range cacheTypes {
		b.Run(string(cacheType), func(b *testing.B) {
			cfg := engine.Config{
				Type:            cacheType,
				Size:            1000,
				CleanupInterval: 60,
				TTL:             300,
				Partition:       1,
				FactName:        "DiscountFact",
			}
			grule := engine.NewPartitionEngine(cfg, nil)
			defer grule.Close()

			// Pre-load some rules
			for i := 0; i < 100; i++ {
				rule := fmt.Sprintf("Rule%d", i)
				statement := fmt.Sprintf(`rule %s "Test rule %d" salience 10 {
					when
						DiscountFact.Amount > %d
					then
						DiscountFact.Discount = %d;
						Retract("%s");
				}`, rule, i, i*10, i, rule)
				_ = grule.AddRule(rule, statement, 300)
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					rule := fmt.Sprintf("Rule%d", i%100)
					fact := &DiscountFact{Amount: (i%100)*10 + 50}
					_ = grule.Execute(context.Background(), rule, fact)
					i++
				}
			})
		})
	}
}

// BenchmarkEnginePartitions benchmarks engine performance with different partition counts
func BenchmarkEnginePartitions(b *testing.B) {
	partitions := []int{1, 2, 4, 8}

	for _, partition := range partitions {
		b.Run(fmt.Sprintf("Partitions%d", partition), func(b *testing.B) {
			cfg := engine.Config{
				Type:            engine.LRU,
				Size:            1000,
				CleanupInterval: 60,
				TTL:             300,
				Partition:       partition,
				FactName:        "DiscountFact",
			}
			grule := engine.NewPartitionEngine(cfg, nil)
			defer grule.Close()

			// Pre-load rules
			for i := 0; i < 100; i++ {
				rule := fmt.Sprintf("Rule%d", i)
				statement := fmt.Sprintf(`rule %s "Test rule %d" salience 10 {
					when
						DiscountFact.Amount > %d
					then
						DiscountFact.Discount = %d;
						Retract("%s");
				}`, rule, i, i*10, i, rule)
				_ = grule.AddRule(rule, statement, 300)
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					rule := fmt.Sprintf("Rule%d", i%100)
					fact := &DiscountFact{Amount: (i%100)*10 + 50}
					_ = grule.Execute(context.Background(), rule, fact)
					i++
				}
			})
		})
	}
}

// BenchmarkRuleLoading benchmarks the performance of loading rules
func BenchmarkRuleLoading(b *testing.B) {
	cfg := engine.Config{
		Type:            engine.LRU,
		Size:            1000,
		CleanupInterval: 60,
		TTL:             300,
		Partition:       1,
		FactName:        "DiscountFact",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		grule := engine.NewPartitionEngine(cfg, nil)

		// Load multiple rules
		for j := 0; j < 50; j++ {
			rule := fmt.Sprintf("Rule%d_%d", i, j)
			statement := fmt.Sprintf(`rule %s "Test rule %d-%d" salience 10 {
				when
					DiscountFact.Amount > %d
				then
					DiscountFact.Discount = %d;
					Retract("%s");
			}`, rule, i, j, j*10, j, rule)
			_ = grule.AddRule(rule, statement, 300)
		}

		grule.Close()
	}
}

// BenchmarkConcurrentExecution benchmarks concurrent rule execution
func BenchmarkConcurrentExecution(b *testing.B) {
	cfg := engine.Config{
		Type:            engine.LRU,
		Size:            1000,
		CleanupInterval: 60,
		TTL:             300,
		Partition:       4,
		FactName:        "DiscountFact",
	}
	grule := engine.NewPartitionEngine(cfg, nil)
	defer grule.Close()

	// Pre-load rules
	for i := 0; i < 200; i++ {
		rule := fmt.Sprintf("Rule%d", i)
		statement := fmt.Sprintf(`rule %s "Test rule %d" salience 10 {
			when
				DiscountFact.Amount > %d
			then
				DiscountFact.Discount = %d;
				Retract("%s");
		}`, rule, i, i*5, i, rule)
		_ = grule.AddRule(rule, statement, 300)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		localRand := rand.New(rand.NewSource(time.Now().UnixNano()))
		i := 0
		for pb.Next() {
			rule := fmt.Sprintf("Rule%d", localRand.Intn(200))
			fact := &DiscountFact{Amount: localRand.Intn(1000) + 100}
			_ = grule.Execute(context.Background(), rule, fact)
			i++
		}
	})
}

// BenchmarkCacheSizes benchmarks performance with different cache sizes
func BenchmarkCacheSizes(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
			cfg := engine.Config{
				Type:            engine.LRU,
				Size:            size,
				CleanupInterval: 60,
				TTL:             300,
				Partition:       2,
				FactName:        "DiscountFact",
			}
			grule := engine.NewPartitionEngine(cfg, nil)
			defer grule.Close()

			// Pre-load rules equal to cache size
			for i := 0; i < size/10; i++ {
				rule := fmt.Sprintf("Rule%d", i)
				statement := fmt.Sprintf(`rule %s "Test rule %d" salience 10 {
					when
						DiscountFact.Amount > %d
					then
						DiscountFact.Discount = %d;
						Retract("%s");
				}`, rule, i, i*10, i, rule)
				_ = grule.AddRule(rule, statement, 300)
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					rule := fmt.Sprintf("Rule%d", i%(size/10))
					fact := &DiscountFact{Amount: (i%(size/10))*10 + 50}
					_ = grule.Execute(context.Background(), rule, fact)
					i++
				}
			})
		})
	}
}

// BenchmarkMemoryUsage benchmarks memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	cfg := engine.Config{
		Type:            engine.LRU,
		Size:            10000,
		CleanupInterval: 60,
		TTL:             300,
		Partition:       4,
		FactName:        "DiscountFact",
	}
	grule := engine.NewPartitionEngine(cfg, nil)
	defer grule.Close()

	// Load many rules to stress memory
	for i := 0; i < 1000; i++ {
		rule := fmt.Sprintf("MemoryRule%d", i)
		statement := fmt.Sprintf(`rule %s "Memory test rule %d" salience 10 {
			when
				DiscountFact.Amount > %d
			then
				DiscountFact.Discount = %d;
				Retract("%s");
		}`, rule, i, i, i%100, rule)
		_ = grule.AddRule(rule, statement, 300)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rule := fmt.Sprintf("MemoryRule%d", i%1000)
		fact := &DiscountFact{Amount: (i % 1000) + 1}
		_ = grule.Execute(context.Background(), rule, fact)
	}
}

// BenchmarkTTLEffects benchmarks the impact of TTL on performance
func BenchmarkTTLEffects(b *testing.B) {
	ttls := []int{0, 30, 300, 3600} // 0 = no TTL, others in seconds

	for _, ttl := range ttls {
		b.Run(fmt.Sprintf("TTL%d", ttl), func(b *testing.B) {
			cfg := engine.Config{
				Type:            engine.LRU,
				Size:            1000,
				CleanupInterval: 10,
				TTL:             ttl,
				Partition:       1,
				FactName:        "DiscountFact",
			}
			grule := engine.NewPartitionEngine(cfg, nil)
			defer grule.Close()

			// Pre-load rules
			for i := 0; i < 100; i++ {
				rule := fmt.Sprintf("TTLRule%d", i)
				statement := fmt.Sprintf(`rule %s "TTL test rule %d" salience 10 {
					when
						DiscountFact.Amount > %d
					then
						DiscountFact.Discount = %d;
						Retract("%s");
				}`, rule, i, i*10, i, rule)
				_ = grule.AddRule(rule, statement, int64(ttl))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rule := fmt.Sprintf("TTLRule%d", i%100)
				fact := &DiscountFact{Amount: (i%100)*10 + 50}
				_ = grule.Execute(context.Background(), rule, fact)
			}
		})
	}
}
