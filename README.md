# grule-plus

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://golang.org/)
[![CI](https://github.com/hungpdn/grule-plus/workflows/CI/badge.svg)](https://github.com/hungpdn/grule-plus/actions)
[![codecov](https://codecov.io/gh/hungpdn/grule-plus/branch/main/graph/badge.svg)](https://codecov.io/gh/hungpdn/grule-plus)
[![Go Report Card](https://goreportcard.com/badge/github.com/hungpdn/grule-plus)](https://goreportcard.com/report/github.com/hungpdn/grule-plus)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

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
        Type:            engine.LRU,
        Size:            1000,
        CleanupInterval: 10,
        TTL:             60,
        Partition:       1,
        FactName:        "DiscountFact",
    }
    grule := engine.NewPartitionEngine(cfg, nil)

    rule := "DiscountRule"
    statement := `rule DiscountRule "Apply discount" salience 10 { 
                when 
                    DiscountFact.Amount > 100 
                then 
                    DiscountFact.Discount = 10; 
                    Retract("DiscountRule");
                }`
    _ = grule.AddRule(rule, statement, 60)

    fact := struct {
        Amount   int
        Discount int
    }{Amount: 150}

    _ = grule.Execute(context.Background(), rule, &fact)
    
    fmt.Printf("fact.Discount = 10: %v", fact.Discount)
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

- **internal/cache:** 2q.
- **internal/consistenthash**.
- **benchmark**.
- **protobuf**.

---

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
