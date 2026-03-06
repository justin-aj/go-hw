# HW-3: Go Concurrency & Synchronization Experiments

Comprehensive exploration of concurrency patterns, synchronization primitives, and their performance tradeoffs in Go. This assignment examines race conditions, mutex strategies, concurrent data structures, and I/O patterns.

## Overview

This project implements various Go concurrency scenarios to understand:
- Race conditions and detection
- Mutex vs RWMutex performance tradeoffs
- sync.Map for concurrent access
- Context switching overhead
- File I/O blocking behavior
- Load testing with various concurrency models

## Project Structure

```
HW-3/
├── race-condition.go                     # Basic race condition demonstration
├── race-condition-2.go                   # Race condition with increment
├── race-condition-2-mutex.go             # Fixing with sync.Mutex
├── race-condition-2-rwmutex.go           # Fixing with sync.RWMutex
├── race-condition-2-syncmap.go           # Concurrent-safe sync.Map
├── atomic-counters.go                    # Atomic operations
├── context-switching-experiment.go       # Context switching analysis
├── file-io-experiment.go                 # Buffered vs unbuffered I/O
├── locustfile.py                         # Load testing (standard)
├── locustfile-fast.py                    # Load testing (optimized)
├── requirements.txt                      # Python dependencies
├── docker-compose-locust.yml             # Locust testing (standard)
├── docker-compose-locust-fast.yml        # Locust testing (optimized)
├── buffered_output.txt                   # Buffered I/O results
├── unbuffered_output.txt                 # Unbuffered I/O results
├── CONTEXT_SWITCHING_EXPERIMENT.md       # Context switching analysis
├── FILE_IO_EXPERIMENT.md                 # I/O experiments & findings
├── LOCUST_EXPERIMENT.md                  # Load testing results
├── MAP_RACE_CONDITION_EXPERIMENT.md      # Mapping concurrent access
├── RWMUTEX_EXPERIMENT.md                 # RWMutex performance analysis
├── SYNCMAP_EXPERIMENT.md                 # sync.Map performance findings
├── TRADEOFFS_COMPARISON.md               # Comprehensive tradeoffs analysis
└── README.md                             # This file

```

## Key Experiments

### 1. Race Condition Detection

**Program**: `race-condition.go`, `race-condition-2.go`
- Demonstrates unsynchronized access to shared data
- Multiple goroutines reading/writing simultaneously
- Results in unpredictable behavior

**Run with race detector**:
```bash
go run -race race-condition.go
```

### 2. Synchronization Strategies

#### Mutex (Binary Lock)
**Program**: `race-condition-2-mutex.go`
- Exclusive access to resource
- One goroutine at a time
- Best for frequent writes

**Performance**: Higher contention with concurrent reads

#### RWMutex (Reader-Writer Lock)
**Program**: `race-condition-2-rwmutex.go`
- Multiple concurrent readers
- Exclusive writer access
- Overhead: writer must wait for all readers

**Best for**: Read-heavy workloads

#### sync.Map (Concurrent Map)
**Program**: `race-condition-2-syncmap.go`
- Optimized for concurrent access
- No explicit locking needed
- Best for high-concurrency scenarios

**Tradeoff**: Slightly slower for sequential access

#### Atomic Operations
**Program**: `atomic-counters.go`
- Lock-free counter increments
- Lowest overhead synchronization
- Limited use cases

### 3. Context Switching Analysis

**Program**: `context-switching-experiment.go`
**Analysis**: `CONTEXT_SWITCHING_EXPERIMENT.md`

Measures overhead of context switching with different numbers of goroutines:
- 1, 10, 100, 1000, 10000 concurrent operations
- Impact on latency and throughput
- Optimal goroutine count analysis

### 4. File I/O Experiments

**Program**: `file-io-experiment.go`
**Results**: `FILE_IO_EXPERIMENT.md`

Compares I/O patterns:
- Buffered writes (performance optimized)
- Unbuffered writes (immediate I/O)
- Latency vs throughput tradeoffs

**Output Files**:
- `buffered_output.txt` - Results from buffered I/O
- `unbuffered_output.txt` - Results from unbuffered I/O

### 5. Load Testing with Locust

**Standard Load Test**:
```bash
docker-compose -f docker-compose-locust.yml up
```

**Fast/Optimized Load Test**:
```bash
docker-compose -f docker-compose-locust-fast.yml up
```

**Results**: `LOCUST_EXPERIMENT.md`

Metrics measured:
- Requests per second (RPS)
- Response time (p50, p95, p99)
- Error rates
- Connection efficiency

## Running Individual Experiments

### Race Condition Detection
```bash
go run -race race-condition-2.go
```

### Mutex vs RWMutex Comparison
```bash
go run race-condition-2-mutex.go
go run race-condition-2-rwmutex.go
```

### Concurrent Map Performance
```bash
go run race-condition-2-syncmap.go
```

### Context Switching Overhead
```bash
go run context-switching-experiment.go
```

### File I/O Patterns
```bash
go run file-io-experiment.go
```

### Atomic Operations
```bash
go run atomic-counters.go
```

## Load Testing Setup

### Install Python Dependencies
```bash
pip install -r requirements.txt
```

### Run Locust Locally
```bash
locust -f locustfile.py --host=http://localhost:8080
```

Access dashboard at `http://localhost:8089`

### Run with Docker Compose
```bash
docker-compose -f docker-compose-locust.yml up
```

## Detailed Findings

See experiment-specific markdown files for comprehensive analysis:

- **Race Conditions**: `MAP_RACE_CONDITION_EXPERIMENT.md`
- **Context Switching**: `CONTEXT_SWITCHING_EXPERIMENT.md`
- **I/O Performance**: `FILE_IO_EXPERIMENT.md`
- **Synchronization Tradeoffs**: `RWMUTEX_EXPERIMENT.md`, `SYNCMAP_EXPERIMENT.md`
- **Comprehensive Analysis**: `TRADEOFFS_COMPARISON.md`

## Key Findings Summary

### When to Use Each Synchronization Primitive

| Pattern | Read-Heavy | Write-Heavy | No Lock | When |
|---------|-----------|-----------|---------|------|
| **Mutex** | ❌ Bad | ✅ Good | ❌ | Simple, balanced r/w |
| **RWMutex** | ✅ Excellent | ❌ Fair | ❌ | Many readers, few writers |
| **sync.Map** | ✅ Good | ✅ Good | ✅ | Concurrent updates, lock-free |
| **Atomic** | N/A | N/A | ✅ | Simple counters only |

### Optimization Strategies

1. **Minimize Critical Section**: Keep locked code as small as possible
2. **Choose Right Primitive**: Match access pattern to synchronization type
3. **Batch Operations**: Group multiple operations before acquiring lock
4. **Lock-Free Where Possible**: Use atomic operations and queues
5. **Profile First**: Measure before optimizing

## Concepts Covered

- **Race Conditions**: Unsafe concurrent access
- **Mutual Exclusion**: Preventing simultaneous access
- **Reader-Writer Locks**: Optimizing read-heavy workloads
- **Atomic Operations**: Lock-free synchronization
- **Concurrent Data Structures**: sync.Map, channels
- **Context Switching**: Goroutine scheduling overhead
- **I/O Blocking**: Synchronous vs asynchronous patterns
- **Performance Measurement**: Load testing and metrics

## Performance Considerations

### Goroutine Overhead
- Each goroutine uses ~2KB memory
- Context switching cost increases with goroutines
- Optimal count depends on workload

### Lock Contention
- High contention degrades performance
- RWMutex helps with read-heavy loads
- sync.Map best for mixed concurrent access

### I/O Patterns
- Buffered I/O improves throughput
- Unbuffered I/O reduces latency
- Choose based on requirements

## Technologies Used

- **Go 1.16+**: Programming language
- **sync package**: Synchronization primitives
- **sync/atomic**: Lock-free operations
- **Python 3.x**: Load testing scripts
- **Locust**: Load testing framework
- **Docker & Docker Compose**: Containerization

## Best Practices

1. **Always use `-race` during development**
   ```bash
   go test -race ./...
   ```

2. **Profile before optimizing**
   - Use pprof to identify bottlenecks
   - Measure impact of changes

3. **Document lock semantics**
   - Clarify which data is protected
   - Document synchronization strategy

4. **Prefer channels for communication**
   - "Share memory by communicating"
   - Cleaner than shared data + locks

5. **Test concurrent code thoroughly**
   - Use `-race` flag
   - Run tests multiple times
   - Test under load

## References

- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Context Package](https://pkg.go.dev/context)
- [sync Package](https://pkg.go.dev/sync)
- [Effective Go - Concurrency](https://go.dev/doc/effective_go#concurrency)
- [Locust Documentation](https://locust.io/)

## Notes

- Race detector has ~10% performance overhead
- Not all race conditions are detected by `-race`
- Larger goroutine counts don't always mean better performance
- I/O blocking is CPU-dependent; test on target hardware
