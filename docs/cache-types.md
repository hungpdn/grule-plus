# Cache Types

Grule-plus supports multiple cache eviction algorithms, each with different performance characteristics and use cases.

## Overview

| Cache Type | Best For | Memory Usage | Hit Ratio | Complexity |
| --- | --- | --- | --- | --- |
| LRU | General purpose | Low | Good | Low |
| LFU | Frequency-based | High | Excellent | Medium |
| ARC | Adaptive workloads | Medium | Very Good | High |
| TWOQ | Mixed workloads | Medium | Good | Medium |
| RANDOM | Simple eviction | Low | Poor | Low |

## LRU (Least Recently Used)

**Description:** Evicts the least recently accessed items first.

**When to Use:**

- General-purpose caching
- Most common access patterns
- Memory-constrained environments
- Simple, predictable behavior needed

**Performance:**

- Read: ~3.8M ops/sec
- Write: ~2.6M ops/sec
- Memory: 21 B/op read, 64 B/op write

**Example:**

```go
cfg := engine.Config{
    Type: engine.LRU,
    Size: 1000,
    // ... other config
}
```

## LFU (Least Frequently Used)

**Description:** Evicts the least frequently accessed items first.

**When to Use:**

- Workloads with clear access frequency patterns
- Long-running applications
- When some items are accessed much more than others
- Memory is not a primary constraint

**Performance:**

- Read: ~2.4M ops/sec
- Write: ~2.7M ops/sec
- Memory: 104 B/op read, 112 B/op write

**Example:**

```go
cfg := engine.Config{
    Type: engine.LFU,
    Size: 1000,
    // ... other config
}
```

## ARC (Adaptive Replacement Cache)

**Description:** Adaptive algorithm that balances between recency and frequency-based eviction.

**When to Use:**

- Workloads with changing access patterns
- Mixed read/write workloads
- When you need adaptive behavior
- Applications with varying access patterns over time

**Algorithm Details:**

- Maintains two LRU lists (T1, T2) and two ghost lists (B1, B2)
- Adapts replacement policy based on workload characteristics
- T1: Recent items, T2: Frequent items
- B1/B2: Ghost entries for adaptation

**Performance:**

- Read: ~3.3M ops/sec
- Write: ~1.9M ops/sec
- Memory: 69 B/op read, 112 B/op write

**Example:**

```go
cfg := engine.Config{
    Type: engine.ARC,
    Size: 1000,
    // ... other config
}
```

## TWOQ (Two-Queue Cache)

**Description:** Uses two queues to separate one-time and multiple-time accessed items.

**When to Use:**

- Workloads with many one-time accesses
- Scan-resistant caching needed
- Mixed workloads with temporal locality

**Algorithm Details:**

- A1: First-time accessed items (FIFO)
- A2: Frequently accessed items (LRU)
- B: Evicted items tracking

**Performance:**

- Read: ~3.3M ops/sec
- Write: ~2.0M ops/sec
- Memory: Similar to ARC

**Example:**

```go
cfg := engine.Config{
    Type: engine.TWOQ,
    Size: 1000,
    // ... other config
}
```

## RANDOM

**Description:** Random eviction policy.

**When to Use:**

- Simple implementation needed
- Memory is extremely constrained
- No specific access patterns
- Baseline performance comparison

**Performance:**

- Read: ~3.5M ops/sec
- Write: ~2.8M ops/sec
- Memory: 20 B/op read, 60 B/op write

**Example:**

```go
cfg := engine.Config{
    Type: engine.RANDOM,
    Size: 1000,
    // ... other config
}
```

## Configuration Options

All cache types support these configuration options:

```go
type Config struct {
    Type            CacheType // Cache algorithm
    Size            int       // Maximum cache size
    CleanupInterval int       // Cleanup interval in seconds
    TTL             int       // Default TTL in seconds
    Partition       int       // Number of partitions
    FactName        string    // Fact name for rules
}
```

## TTL (Time-To-Live) Support

All cache types support TTL for automatic expiration:

```go
// Set with TTL
cache.Set("key", "value", time.Hour)

// Set without TTL (uses default or no expiration)
cache.Set("key", "value", 0)
```

## Cleanup Behavior

- **Background Cleanup:** Runs in goroutines at specified intervals
- **TTL Expiration:** Automatic removal of expired items
- **Eviction Callbacks:** Optional callbacks when items are evicted

## Choosing a Cache Type

### For Most Applications

Use **LRU** - it provides good performance and predictable behavior.

### For Frequency-Based Caching

Use **LFU** when you have clear frequency patterns and memory isn't constrained.

### For Adaptive Behavior

Use **ARC** when access patterns change over time or you need the best hit ratios.

### For Scan-Resistant Caching

Use **TWOQ** when you have workloads with many one-time accesses.

### Performance Comparison

Based on benchmark results (Intel Core i7-9750H):

```text
Read-Heavy Workload (operations/sec):
LRU:  3,982,196
ARC:  3,251,343
TWOQ: 3,251,343
LFU:  2,401,837

Write-Heavy Workload (operations/sec):
LRU:  2,641,264
LFU:  2,700,162
ARC:  1,939,623
TWOQ: 1,939,623

Memory Usage (bytes per operation):
LRU:  21-64 B/op
ARC:  69-112 B/op
LFU:  104-112 B/op
TWOQ: Similar to ARC
```

## Thread Safety

All cache implementations are thread-safe and can be used concurrently from multiple goroutines.
