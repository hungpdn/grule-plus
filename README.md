# grule-plus

**grule-plus** is a high-performance, extensible rule engine built on top of [Grule Rule Engine](https://github.com/hyperjumptech/grule-rule-engine). It provides advanced caching, partitioning, and flexible configuration for scalable rule evaluation in Go applications.

---

## Features

- **Pluggable Cache Engines:** Supports LRU (Least Recently Used), LFU (Least Frequently Used), ARC (Adaptive Replacement Cache), and RANDOM cache strategies.
- **Partitioned Rule Engine:** Scale horizontally with partitioned engines for concurrent rule evaluation.
- **Flexible TTL & Cleanup:** Per-rule time-to-live and periodic cleanup for cache entries.
- **Structured Logging:** Integrated with Go's `slog` for context-aware, structured logs.
- **Runtime Stats:** Built-in runtime statistics for monitoring and debugging.
- **Thread-Safe:** Safe for concurrent use in multi-goroutine environments.

---

## Getting Started

### Installation

```sh
go get github.com/hungpdn/grule-plus
```

### Example Usage

```go
import (
    "context"
    "github.com/hungpdn/grule-plus/engine"
)

func main() {
    cfg := engine.Config{
        Type:            0, // LRU
        Size:            1000,
        CleanupInterval: 10,
        TTL:             60,
        Partition:       4,
    }
    grule := engine.NewPartitionEngine(cfg, nil)

    rule := "DiscountRule"
    statement := `rule DiscountRule "Apply discount" salience 10 { when DiscountFact.Amount > 100 then DiscountFact.Discount = 10; }`
    grule.AddRule(rule, statement, 60)

    fact := struct {
        Amount   int
        Discount int
    }{Amount: 150}

    grule.Execute(context.Background(), rule, &fact)
}
```

---

## Configuration

See `engine.Config` for all available options:

- `Type`: Cache type (LRU, LFU, ARC, RANDOM)
- `Size`: Maximum cache size
- `CleanupInterval`: Cache cleanup interval (seconds)
- `TTL`: Default time-to-live for rules (seconds)
- `Partition`: Number of partitions for parallelism

---

## TODO

- **internal/cache:** arc, 2q.
- **internal/consistenthash**.
