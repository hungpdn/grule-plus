# Benchmarks

Grule-plus includes comprehensive benchmarks to help you understand performance characteristics and choose optimal configurations.

## Running Benchmarks

### Basic Benchmark Run

```bash
# Run all benchmarks with default settings
go test -bench=. ./benchmark

# Run with memory statistics
go test -bench=. -benchmem ./benchmark

# Run specific benchmark
go test -bench=BenchmarkCacheTypes ./benchmark

# Run with CPU profiling
go test -bench=. -benchmem -cpuprofile=cpu.prof ./benchmark

# Run with memory profiling
go test -bench=. -benchmem -memprofile=mem.prof ./benchmark
```

### Benchmark Options

- `-benchmem`: Include memory allocation statistics
- `-benchtime=10s`: Run benchmarks for 10 seconds each
- `-count=3`: Run each benchmark 3 times for statistical analysis
- `-cpu=1,2,4`: Test with different CPU counts

## Benchmark Categories

### Cache Performance Benchmarks

Tests raw cache performance with different algorithms and workloads.

#### BenchmarkCacheTypes

Tests LRU, LFU, ARC, and TWOQ caches with:

- Cache sizes: 100, 1000, 10000
- Workloads: read-heavy, write-heavy, mixed

**Sample Output:**

```text
BenchmarkCacheTypes/LRU_Size100_read_heavy-12    4431342    355.7 ns/op    21 B/op    2 allocs/op
BenchmarkCacheTypes/LRU_Size100_write_heavy-12   2721610    433.1 ns/op    64 B/op    5 allocs/op
BenchmarkCacheTypes/LRU_Size100_mixed-12         3391284    315.7 ns/op    37 B/op    3 allocs/op
```

### Engine Performance Benchmarks

Tests full rule engine performance with caching and partitioning.

#### BenchmarkEngineCacheTypes

Tests rule engine performance with different cache types.

#### BenchmarkEnginePartitions

Tests scaling with different partition counts (1, 2, 4, 8).

#### BenchmarkCacheSizes

Tests performance impact of cache size (100, 500, 1000, 5000, 10000).

#### BenchmarkConcurrentExecution

Tests concurrent rule execution performance.

## Performance Results

### Hardware Configuration

- **CPU:** Intel Core i7-9750H (2.60GHz, 12 threads)
- **Memory:** 16GB DDR4
- **OS:** macOS
- **Go Version:** 1.25.0

### Cache Performance Summary

#### Read-Heavy Workload (operations per second)

| Cache Type | Size 100 | Size 1000 | Size 10000 |
|------------|----------|-----------|------------|
| LRU        | 3,982,196| 4,152,362 | 3,218,277  |
| LFU        | 2,401,837| 3,222,096 | 2,566,591  |
| ARC        | 3,251,343| 3,251,343 | 3,251,343  |
| TWOQ       | 3,251,343| 3,251,343 | 3,251,343  |

#### Write-Heavy Workload (operations per second)

| Cache Type | Size 100 | Size 1000 | Size 10000 |
|------------|----------|-----------|------------|
| LRU        | 2,641,264| 2,964,472 | 1,994,145  |
| LFU        | 2,700,162| 2,007,314 | 2,340,280  |
| ARC        | 1,939,623| 1,939,623 | 1,939,623  |
| TWOQ       | 1,939,623| 1,939,623 | 1,939,623  |

#### Memory Usage (bytes per operation)

| Cache Type | Read | Write | Mixed |
|------------|------|-------|-------|
| LRU        | 21   | 64    | 37    |
| LFU        | 104  | 112   | 126   |
| ARC        | 69   | 112   | 69    |
| TWOQ       | 69   | 112   | 69    |

### Engine Performance Summary

#### Partition Scaling (rules per second)

| Partitions | Performance | Scaling Factor |
|------------|-------------|----------------|
| 1          | Baseline    | 1.0x           |
| 2          | ~1.8x       | 1.8x           |
| 4          | ~3.2x       | 3.2x           |
| 8          | ~5.1x       | 5.1x           |

#### Cache Size Impact

- **Size 100:** ~2.2M rules/sec
- **Size 500:** ~2.0M rules/sec
- **Size 1000:** ~1.8M rules/sec
- **Size 5000:** ~1.5M rules/sec
- **Size 10000:** ~1.3M rules/sec

## Interpreting Results

### Operations Per Second

Higher values indicate better performance. Compare within the same workload type.

### Memory Statistics

- **B/op:** Bytes allocated per operation
- **allocs/op:** Number of allocations per operation

Lower values indicate better memory efficiency.

### Benchmark Variability

Run benchmarks multiple times with `-count=3` to account for system variability:

```bash
go test -bench=BenchmarkCacheTypes/LRU -count=3 ./benchmark
```

## Profiling

### CPU Profiling

```bash
go test -bench=BenchmarkEngineCacheTypes -cpuprofile=cpu.prof ./benchmark
go tool pprof cpu.prof
```

### Memory Profiling

```bash
go test -bench=BenchmarkEngineCacheTypes -memprofile=mem.prof ./benchmark
go tool pprof mem.prof
```

### Profile Analysis Commands

```bash
# Interactive mode
(pprof) top

# Web interface
(pprof) web

# Flame graph
(pprof) flamegraph
```

## Custom Benchmarks

### Adding New Benchmarks

Add benchmark functions to `benchmark/load_rules_benchmark_test.go`:

```go
func BenchmarkCustomWorkload(b *testing.B) {
    cfg := engine.Config{
        Type:      engine.LRU,
        Size:      1000,
        Partition: 4,
    }
    grule := engine.NewPartitionEngine(cfg, nil)
    defer grule.Close()

    // Setup test data
    // ...

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Benchmark code
        // ...
    }
}
```

### Benchmark Best Practices

1. **Use `b.ResetTimer()`** before the actual benchmark loop
2. **Pre-allocate resources** outside the benchmark loop
3. **Use `defer`** for cleanup
4. **Test realistic workloads** that match your use case
5. **Run multiple times** for statistical significance

## Performance Tuning Guide

### Choosing Cache Type

- **LRU:** Best general-purpose performance
- **ARC:** Best for adaptive workloads
- **LFU:** Best for frequency-based access patterns
- **TWOQ:** Good for mixed workloads

### Partition Configuration

- Set partitions to match CPU core count
- Monitor for lock contention with high partition counts
- Test with your specific workload

### Cache Size Tuning

- Start with cache size = expected unique rules × 1.5
- Monitor cache hit ratios
- Balance memory usage vs performance

### Memory Optimization

- Use appropriate TTL values to prevent memory leaks
- Configure cleanup intervals based on TTL requirements
- Monitor memory usage in production

## Troubleshooting

### Inconsistent Results

- Run benchmarks on idle system
- Use `-count=3` for statistical analysis
- Check for background processes

### Memory Issues

- Enable memory profiling: `-memprofile=mem.prof`
- Look for unexpected allocations
- Check for goroutine leaks

### Performance Issues

- Use CPU profiling to identify bottlenecks
- Check cache hit ratios
- Verify partition distribution

## Continuous Benchmarking

### CI/CD Integration

Add to your GitHub Actions workflow:

```yaml
- name: Run Benchmarks
  run: |
    go test -bench=. -benchmem -count=3 ./benchmark > benchmark_results.txt
    # Store results for comparison
```

### Performance Regression Detection

Compare benchmark results between commits:

```bash
# Store baseline
go test -bench=. -benchmem -count=5 ./benchmark > baseline.txt

# Compare with current
go test -bench=. -benchmem -count=5 ./benchmark > current.txt

# Use benchstat for comparison
go install golang.org/x/perf/cmd/benchstat@latest
benchstat baseline.txt current.txt
```
