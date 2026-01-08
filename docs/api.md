# API Reference

## Engine Package

### Types

#### `IGruleEngine` Interface

```go
type IGruleEngine interface {
    Execute(ctx context.Context, rule string, fact any) error
    FetchMatching(ctx context.Context, rule string, fact any) ([]*ast.RuleEntry, error)
    AddRule(rule, statement string, duration int64) error
    BuildRule(rule, statement string, duration int64) error
    ContainsRule(rule string) bool
    Debug() map[string]any
    Close()
}
```

The main interface for rule engine operations.

#### `Config` Struct

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

Configuration structure for the rule engine.

#### `CacheType` Type

```go
type CacheType string

const (
    LRU    CacheType = "lru"
    LFU    CacheType = "lfu"
    ARC    CacheType = "arc"
    TWOQ   CacheType = "twoq"
    RANDOM CacheType = "random"
)
```

Supported cache types.

### Functions

#### `NewPartitionEngine`

```go
func NewPartitionEngine(cfg Config, hashFunc HashFunc) *partitionEngine
```

Creates a new partitioned rule engine instance.

**Parameters:**

- `cfg`: Engine configuration
- `hashFunc`: Optional hash function for partitioning (nil uses default)

**Returns:** Pointer to partitionEngine instance

#### `GetCacheType`

```go
func (c Config) GetCacheType() int
```

Converts CacheType string to internal cache type constant.

#### `GetFactName`

```go
func (c Config) GetFactName() string
```

Returns the configured fact name or default "Fact".

## Cache Package

### Interface

#### `ICache` Interface

```go
type ICache interface {
    Set(key any, value any, duration time.Duration)
    Get(key any) (value any, ok bool)
    Has(key any) bool
    Keys() []any
    Len() int
    Clear()
    Close()
    SetEvictedFunc(f common.EvictedFunc) error
}
```

Cache interface for all cache implementations.

### Functions

#### `New`

```go
func New(config Config) ICache
```

Creates a new cache instance based on configuration.

## ConsistentHash Package

### Types

#### `ConsistentHash` Struct

```go
type ConsistentHash struct {
    // Contains filtered or unexported fields
}
```

Consistent hashing implementation for key distribution.

#### `HashFunc` Type

```go
type HashFunc func(data []byte) uint32
```

Function signature for hash functions.

### Functions

#### `New`

```go
func New(replicas int, hashFunc HashFunc) *ConsistentHash
```

Creates a new consistent hash instance.

#### `AddNode`

```go
func (c *ConsistentHash) AddNode(node string)
```

Adds a node to the hash ring.

#### `RemoveNode`

```go
func (c *ConsistentHash) RemoveNode(node string)
```

Removes a node from the hash ring.

#### `GetNode`

```go
func (c *ConsistentHash) GetNode(key string) string
```

Returns the node responsible for the given key.

#### `GetNodes`

```go
func (c *ConsistentHash) GetNodes() []string
```

Returns all nodes in the ring.

## Error Handling

All functions return appropriate errors that should be checked:

```go
err := grule.Execute(ctx, rule, fact)
if err != nil {
    log.Printf("Rule execution failed: %v", err)
}
```

## Thread Safety

All grule-plus components are thread-safe and can be used concurrently from multiple goroutines.
