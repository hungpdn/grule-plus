# Grule-Plus Documentation

Welcome to the official documentation for **grule-plus**, a high-performance, extensible rule engine built on top of Grule Rule Engine.

## Overview

Grule-plus provides advanced caching, partitioning, and flexible configuration for scalable rule evaluation in Go applications.

## Quick Start

```go
import "github.com/hungpdn/grule-plus/engine"

cfg := engine.Config{
    Type:            engine.LRU,
    Size:            1000,
    CleanupInterval: 10,
    TTL:             60,
    Partition:       1,
    FactName:        "DiscountFact",
}

grule := engine.NewPartitionEngine(cfg, nil)
```

## Table of Contents

- [API Reference](api.md)
- [Architecture](architecture.md)
- [Cache Types](cache-types.md)
- [Configuration](configuration.md)
- [Benchmarks](benchmarks.md)
- [Examples](examples.md)
