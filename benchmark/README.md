# Grule-Plus Benchmarks

This directory contains comprehensive benchmarks for the grule-plus rule engine, testing various cache types, configurations, and workloads.

## Benchmark Categories

### Cache Performance Benchmarks

- **BenchmarkCacheTypes**: Tests LRU, LFU, ARC, and TWOQ cache implementations with different sizes (100, 1000, 10000) and workloads (read-heavy, write-heavy, mixed).

### Engine Performance Benchmarks

- **BenchmarkEngineCacheTypes**: Tests rule engine performance with different cache types.
- **BenchmarkEnginePartitions**: Tests partitioned engine performance with 1, 2, 4, and 8 partitions.
- **BenchmarkCacheSizes**: Tests performance with different cache sizes (100, 500, 1000, 5000, 10000).
- **BenchmarkConcurrentExecution**: Tests concurrent rule execution performance.
- **BenchmarkRuleLoading**: Tests rule loading performance.
- **BenchmarkMemoryUsage**: Tests memory usage patterns with large rule sets.
- **BenchmarkTTLEffects**: Tests the impact of TTL settings on performance.

## Running Benchmarks

```bash
# Run all benchmarks with memory statistics
go test -bench=. -benchmem -count=1 ./benchmark

# Run specific benchmark
go test -bench=BenchmarkCacheTypes -benchmem ./benchmark

# Run benchmarks with CPU profiling
go test -bench=. -benchmem -cpuprofile=cpu.prof ./benchmark

# Run benchmarks with memory profiling
go test -bench=. -benchmem -memprofile=mem.prof ./benchmark
```

## Sample Results

Based on testing on Intel Core i7-9750H:

### Cache Performance (operations per second, higher is better)

- LRU Read-heavy (size 100): ~3.8M ops/sec
- LRU Write-heavy (size 100): ~2.6M ops/sec
- ARC Read-heavy (size 100): ~3.3M ops/sec
- LFU Read-heavy (size 100): ~2.4M ops/sec

### Memory Usage (bytes per operation)

- LRU Read: 21 B/op, 2 allocs/op
- LRU Write: 64 B/op, 5 allocs/op
- ARC Read: 69 B/op, 3 allocs/op
- LFU Read: 104 B/op, 3 allocs/op

## Key Findings

1. **LRU** provides the best overall performance for most workloads
2. **ARC** offers good adaptive behavior but with higher memory usage
3. **LFU** has higher memory overhead due to frequency tracking
4. **Partitioning** significantly improves concurrent performance
5. **Cache size** impacts performance - larger caches have higher latency but better hit rates

## Data Files

- `rule_template.txt`: Template for generating test rules
- `sample_rules.grl`: Sample Grule rules for testing
