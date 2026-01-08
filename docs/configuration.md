# Configuration

This document describes all configuration options available in grule-plus.

## Config Struct

```go
type Config struct {
    Type            CacheType // Cache type: lru, lfu, arc, twoq, random
    Size            int       // Cache size, 0 means unlimited
    CleanupInterval int       // Cleanup interval in seconds, 0 means no cleanup
    TTL             int       // Time-to-live in seconds, 0 means no expiration
    Partition       int       // Number of partitions for the engine
    FactName        string    // Name of the fact to be used in rules
}
```

## Configuration Options

### Cache Type (`Type`)

**Type:** `CacheType` (string)

**Default:** `LRU`

**Options:**

- `"lru"` - Least Recently Used
- `"lfu"` - Least Frequently Used
- `"arc"` - Adaptive Replacement Cache
- `"twoq"` - Two-Queue Cache
- `"random"` - Random eviction

**Description:** Specifies which cache eviction algorithm to use.

```go
cfg := engine.Config{
    Type: engine.LRU, // or engine.LFU, engine.ARC, etc.
}
```

### Cache Size (`Size`)

**Type:** `int`

**Default:** `0` (unlimited)

**Description:** Maximum number of items to store in the cache. When set to 0, the cache has no size limit.

```go
cfg := engine.Config{
    Size: 1000, // Cache up to 1000 rules
}
```

### Cleanup Interval (`CleanupInterval`)

**Type:** `int` (seconds)

**Default:** `0` (no cleanup)

**Description:** How often to run the background cleanup goroutine to remove expired items. Set to 0 to disable automatic cleanup.

```go
cfg := engine.Config{
    CleanupInterval: 60, // Clean up every 60 seconds
}
```

### Time-To-Live (`TTL`)

**Type:** `int` (seconds)

**Default:** `0` (no expiration)

**Description:** Default time-to-live for cached items. Individual items can override this when added.

```go
cfg := engine.Config{
    TTL: 300, // 5 minutes default TTL
}
```

### Partition Count (`Partition`)

**Type:** `int`

**Default:** `1`

**Description:** Number of partitions to use for the rule engine. Should typically match the number of CPU cores for optimal performance.

```go
cfg := engine.Config{
    Partition: runtime.NumCPU(), // Use all available CPU cores
}
```

### Fact Name (`FactName`)

**Type:** `string`

**Default:** `"Fact"`

**Description:** Name of the fact object used in Grule rules.

```go
cfg := engine.Config{
    FactName: "DiscountFact", // Custom fact name
}
```

## Example Configurations

### Basic Configuration

```go
cfg := engine.Config{
    Type:            engine.LRU,
    Size:            1000,
    CleanupInterval: 60,
    TTL:             300,
    Partition:       1,
    FactName:        "Fact",
}
```

### High-Performance Configuration

```go
cfg := engine.Config{
    Type:            engine.ARC,
    Size:            10000,
    CleanupInterval: 30,
    TTL:             600,
    Partition:       runtime.NumCPU(),
    FactName:        "BusinessFact",
}
```

### Memory-Constrained Configuration

```go
cfg := engine.Config{
    Type:            engine.LRU,
    Size:            100,
    CleanupInterval: 300, // Less frequent cleanup
    TTL:             60,   // Shorter TTL
    Partition:       1,
    FactName:        "Fact",
}
```

### No Expiration Configuration

```go
cfg := engine.Config{
    Type:            engine.LFU,
    Size:            5000,
    CleanupInterval: 0, // No automatic cleanup
    TTL:             0, // No default expiration
    Partition:       4,
    FactName:        "RuleFact",
}
```

## Advanced Configuration

### Custom Hash Function for Partitioning

```go
hashFunc := func(rule string) int {
    // Custom partitioning logic
    h := fnv.New32a()
    h.Write([]byte(rule))
    return int(h.Sum32()) % partitionCount
}

grule := engine.NewPartitionEngine(cfg, hashFunc)
```

### Eviction Callbacks

```go
cache := cache.New(cache.Config{
    Type: cache.LRU,
    Size: 1000,
    EvictedFunc: func(key, value any) {
        log.Printf("Evicted key: %v", key)
    },
})
```

## Performance Tuning

### Cache Size Tuning

- **Small caches (100-1000):** Fast access, lower memory usage
- **Medium caches (1000-10000):** Good balance of speed and hit ratio
- **Large caches (10000+):** Better hit ratios, slower access

### Partition Tuning

- **Single partition:** Simple, good for low concurrency
- **Multiple partitions:** Better concurrency, matches CPU cores
- **Too many partitions:** Increased contention, diminishing returns

### Cleanup Tuning

- **Frequent cleanup (10-30s):** Responsive expiration, higher CPU usage
- **Infrequent cleanup (60-300s):** Lower CPU usage, less responsive expiration
- **No cleanup:** Lowest CPU usage, manual expiration only

## Environment-Specific Configuration

### Development

```go
cfg := engine.Config{
    Type:            engine.LRU,
    Size:            100,
    CleanupInterval: 10,
    TTL:             60,
    Partition:       1,
    FactName:        "DevFact",
}
```

### Production

```go
cfg := engine.Config{
    Type:            engine.ARC,
    Size:            10000,
    CleanupInterval: 30,
    TTL:             3600,
    Partition:       runtime.NumCPU(),
    FactName:        "ProdFact",
}
```

### Testing

```go
cfg := engine.Config{
    Type:            engine.LRU,
    Size:            10,
    CleanupInterval: 1,
    TTL:             5,
    Partition:       1,
    FactName:        "TestFact",
}
```

## Validation

The configuration is validated when creating the engine:

```go
grule := engine.NewPartitionEngine(cfg, nil)
if grule == nil {
    log.Fatal("Invalid configuration")
}
```

## Monitoring Configuration

Use the `Debug()` method to inspect current configuration:

```go
debug := grule.Debug()
fmt.Printf("Config: %+v\n", debug["partition_config"])
fmt.Printf("Stats: %+v\n", debug["stats"])
```
