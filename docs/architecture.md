# Architecture

## Overview

Grule-plus is a high-performance rule engine built on top of the [Grule Rule Engine](https://github.com/hyperjumptech/grule-rule-engine). It provides advanced caching, partitioning, and consistent hashing for scalable rule evaluation.

## High-Level Architecture

```text
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                 Grule-Plus Engine                   │    │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐    │    │
│  │  │ Partition 1 │ │ Partition 2 │ │ Partition N │    │    │
│  │  │             │ │             │ │             │    │    │
│  │  │ ┌─────────┐ │ │ ┌─────────┐ │ │ ┌─────────┐ │    │    │
│  │  │ │  Cache  │ │ │ │  Cache  │ │ │ │  Cache  │ │    │    │
│  │  │ │ (LRU/   │ │ │ │ (ARC/   │ │ │ │ (LFU/   │ │    │    │
│  │  │ │  LFU)   │ │ │ │  TWOQ)  │ │ │ │  etc)   │ │    │    │
│  │  │ └─────────┘ │ │ └─────────┘ │ │ └─────────┘ │    │    │
│  │  │             │ │             │ │             │    │    │
│  │  │ ┌─────────┐ │ │ ┌─────────┐ │ │ ┌─────────┐ │    │    │
│  │  │ │  Grule  │ │ │ │  Grule  │ │ │ │  Grule  │ │    │    │
│  │  │ │ Engine  │ │ │ │ Engine  │ │ │ │ Engine  │ │    │    │
│  │  │ └─────────┘ │ │ └─────────┘ │ │ └─────────┘ │    │    │
│  │  └─────────────┘ └─────────────┘ └─────────────┘    │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                  Grule Rule Engine                          │
│  (https://github.com/hyperjumptech/grule-rule-engine)       │
└─────────────────────────────────────────────────────────────┘
```

## Component Architecture

### 1. Engine Layer

The engine layer provides the main API and orchestrates rule execution:

- **PartitionEngine**: Manages multiple single engines with partitioning
- **SingleEngine**: Individual rule engine instances with caching
- **Configuration**: Centralized configuration management

### 2. Cache Layer

Pluggable cache implementations for rule storage and retrieval:

- **LRU (Least Recently Used)**: Fast, simple eviction policy
- **LFU (Least Frequently Used)**: Frequency-based eviction
- **ARC (Adaptive Replacement Cache)**: Adaptive algorithm balancing recency/frequency
- **TWOQ (Two-Queue)**: Two-queue eviction policy
- **TTL Support**: Time-based expiration
- **Cleanup**: Background cleanup goroutines

### 3. Partitioning Layer

Scalable partitioning using consistent hashing:

- **ConsistentHash**: Key distribution across partitions
- **Virtual Nodes**: Better load distribution
- **Dynamic Scaling**: Add/remove partitions without rehashing

### 4. Supporting Components

- **Logger**: Structured logging with slog
- **Stats**: Runtime statistics collection
- **Utils**: Hash functions, math utilities, recover mechanisms

## Data Flow

### Rule Execution Flow

```text
Client Request
       │
       ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Partition   │────▶│   Cache     │────▶│ Rule Found? │
│ Selection   │     │   Lookup    │     └─────────────┘
│ (Hash)      │     └─────────────┘            │
└─────────────┘                                │
       │                                       ▼
       │                                ┌─────────────┐
       │                                │ Execute     │
       │                                │ Rule        │
       │                                └─────────────┘
       │                                      │
       ▼                                      ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Cache Miss  │────▶│ Load Rule   │────▶│ Execute &   │
│             │     │ from Source │     │ Cache       │
└─────────────┘     └─────────────┘     └─────────────┘
       │                                      │
       └──────────────────────────────────────┘
                      Response
```

### Cache Eviction Flow

```text
Cache Full?
     │
     ▼
┌─────────────┐     ┌─────────────┐
│   Eviction  │────▶│ Call        │
│   Policy    │     │ Eviction    │
│   (LRU/     │     │ Callback    │
│    LFU/     │     └─────────────┘
│    ARC)     │             │
└─────────────┘             ▼
     │              ┌─────────────┐
     │              │ Remove      │
     │              │ Expired     │
     │              │ Entries     │
     │              └─────────────┘
     ▼
┌─────────────┐
│ Add New     │
│ Entry       │
└─────────────┘
```

## Performance Characteristics

### Cache Performance

| Cache Type | Read Performance | Write Performance | Memory Usage |
|------------|------------------|-------------------|--------------|
| LRU        | High             | High              | Low          |
| LFU        | Medium           | Low               | High         |
| ARC        | High             | Medium            | Medium       |
| TWOQ       | High             | Medium            | Medium       |

### Scaling Characteristics

- **Partitions**: Linear scaling with number of partitions
- **Cache Size**: Larger caches improve hit rates but increase memory usage
- **Concurrent Access**: Efficient locking mechanisms for multi-threaded environments
