# sync.Map Experiment

## Overview
This experiment explores `sync.Map`, Go's built-in concurrent map type, and compares it with Mutex and RWMutex approaches for handling concurrent map operations.

## Experiment Setup

### Parameters
- **Goroutines**: 50
- **Iterations per goroutine**: 1,000
- **Total writes**: 50,000
- **Access pattern**: 100% concurrent writes

---

## sync.Map Implementation

### Key Differences from Regular Maps

**Regular map with Mutex:**
```go
type SafeMap struct {
    mu sync.Mutex
    m  map[int]int
}
safeMap.Write(key, value)  // Manual locking required
```

**sync.Map:**
```go
var m sync.Map              // No explicit locking needed
m.Store(key, value)         // Built-in concurrency safety
```

### Implementation Details

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    // Create a sync.Map - no initialization needed
    var m sync.Map
    
    var wg sync.WaitGroup
    startTime := time.Now()
    
    // Spawn 50 goroutines
    for g := 0; g < 50; g++ {
        wg.Add(1)
        go func(goroutineID int) {
            defer wg.Done()
            for i := 0; i < 1000; i++ {
                // Store without explicit locking
                m.Store(goroutineID*1000+i, i)
            }
        }(g)
    }
    
    wg.Wait()
    elapsed := time.Since(startTime)
    
    // Count entries using Range
    count := 0
    m.Range(func(key, value interface{}) bool {
        count++
        return true  // continue iteration
    })
    
    fmt.Printf("Length of map: %d\n", count)
    fmt.Printf("Total time taken: %v\n", elapsed)
}
```

### Key sync.Map Methods

1. **Store(key, value interface{})** - Store a key-value pair
2. **Load(key interface{}) (value interface{}, ok bool)** - Retrieve a value
3. **LoadOrStore(key, value interface{}) (actual interface{}, loaded bool)** - Atomic get-or-create
4. **Delete(key interface{})** - Remove a key
5. **Range(func(key, value interface{}) bool)** - Iterate over all entries

---

## Test Results

### Run 1
```
Length of map: 50000
Total time taken: 3.3608ms
```

### Run 2
```
Length of map: 50000
Total time taken: 2.636ms
```

### Run 3
```
Length of map: 50000
Total time taken: 2.6383ms
```

**Mean time**: ~**2.87ms**

---

## Three-Way Performance Comparison

| Approach | Mean Time | Relative Speed | Correctness |
|----------|-----------|----------------|-------------|
| **Plain Map** | N/A | N/A | ❌ Crashes (race condition) |
| **Mutex** | ~7.37ms | Baseline (1.0x) | ✅ Safe |
| **RWMutex** | ~7.87ms | 0.94x (slower) | ✅ Safe |
| **sync.Map** | ~2.87ms | **2.57x faster** | ✅ Safe |

### Visual Comparison
```
sync.Map:  ████████████████ (2.87ms) ⚡ FASTEST
Mutex:     ████████████████████████████████████████ (7.37ms)
RWMutex:   ███████████████████████████████████████████ (7.87ms)
```

### Performance Improvement
- **sync.Map vs Mutex**: ~61% faster (4.5ms improvement)
- **sync.Map vs RWMutex**: ~64% faster (5.0ms improvement)

---

## Why sync.Map is Faster (Write-Heavy Workload)

### 1. Optimized Internal Design
sync.Map uses a two-map strategy:
```
┌─────────────────────────────────────┐
│          sync.Map                   │
├─────────────────────────────────────┤
│  read (atomic.Value)                │  ← Lock-free reads
│  • Immutable map                    │  ← Copy-on-write
│  • Atomic pointer swaps             │
├─────────────────────────────────────┤
│  dirty (regular map + mutex)        │  ← New writes go here
│  • Protected by mutex               │
│  • Promoted to read map             │
└─────────────────────────────────────┘
```

### 2. Lock-Free Fast Path
- First access attempts lock-free atomic read from `read` map
- Only falls back to locked `dirty` map if key not found
- Reduces lock contention compared to always acquiring mutex

### 3. Amortized Lock Costs
- New keys written to `dirty` map under lock
- When `dirty` map grows large enough, it's promoted to `read` map
- Subsequent reads of those keys become lock-free

### 4. Fine-Grained Locking
Unlike a single global mutex:
- Different keys may have reduced contention
- Internal optimizations for disjoint key access patterns

---

## Tradeoffs: When to Use Each Approach

### 1. sync.Map ⚡

**Use when:**
- ✅ Keys are written once, read many times (append-only pattern)
- ✅ Disjoint key sets accessed by different goroutines
- ✅ Read-heavy workloads with occasional writes
- ✅ Don't need strong type safety (okay with `interface{}`)

**Avoid when:**
- ❌ Need type safety (sync.Map uses `interface{}`)
- ❌ Heavy, continuous writes to same keys
- ❌ Need to iterate frequently (Range has overhead)
- ❌ Simple use case where regular map + mutex is clearer

**Performance characteristics:**
```
Write-heavy (this test): ⚡⚡⚡ (2.57x faster)
Read-heavy:              ⚡⚡⚡⚡⚡ (even better!)
Mixed workload:          ⚡⚡⚡
```

### 2. sync.Mutex 🔒

**Use when:**
- ✅ **Best default choice** - simple and predictable
- ✅ Mixed read/write workloads
- ✅ Need full control over locking granularity
- ✅ Type safety is important
- ✅ Simple, maintainable code preferred

**Characteristics:**
```
Write-heavy: ⚡ (baseline)
Read-heavy:  ⚡ (all reads serialized)
Mixed:       ⚡⚡
```

### 3. sync.RWMutex 📖

**Use when:**
- ✅ 90%+ reads, <10% writes
- ✅ Multiple goroutines need to read simultaneously
- ✅ Read operations are expensive
- ✅ Profiling shows read contention

**Avoid when:**
- ❌ Write-dominated workloads (this test showed 6.8% slower)
- ❌ Starting a new project (use Mutex first, profile later)

**Characteristics:**
```
Write-heavy: 💤 (slower than Mutex due to overhead)
Read-heavy:  ⚡⚡⚡⚡ (multiple concurrent readers)
Mixed:       ⚡⚡⚡
```

---

## Detailed Comparison Table

| Feature | sync.Map | Mutex | RWMutex |
|---------|----------|-------|---------|
| **Type Safety** | ❌ interface{} | ✅ Generic | ✅ Generic |
| **Write Speed** | ⚡⚡⚡ | ⚡ | ⚡ |
| **Read Speed** | ⚡⚡⚡⚡ | ⚡ | ⚡⚡⚡⚡ |
| **Memory** | Higher | Low | Low |
| **Code Complexity** | Low | Low | Medium |
| **Learning Curve** | Medium | Easy | Medium |
| **Iteration** | Range() | Direct | Direct |
| **Length Check** | Range count | len() | len() |

---

## Read-Heavy Workload Analysis

**Question**: What if reads dominate instead of writes?

### Hypothetical Read-Heavy Test (90% reads, 10% writes)

**Expected Results:**
```
sync.Map:   ⚡⚡⚡⚡⚡  (~1-2ms)    - Lock-free reads shine
RWMutex:    ⚡⚡⚡⚡    (~3-4ms)    - Concurrent reads help
Mutex:      ⚡         (~15-20ms) - All reads serialized
```

### Why Each Performs Differently

**sync.Map:**
- Most reads hit the lock-free `read` map
- Atomic operations are very fast
- No lock contention for reads

**RWMutex:**
- Multiple goroutines can hold RLock simultaneously
- Reads don't block each other
- Writers still need exclusive lock

**Mutex:**
- Every read acquires exclusive lock
- All 50 goroutines serialize on every read
- Massive contention bottleneck

---

## sync.Map Internals Explained

### The Two-Map Strategy

```go
type Map struct {
    mu     Mutex
    read   atomic.Value  // readOnly struct
    dirty  map[interface{}]*entry
    misses int
}
```

### Operation Flow

**Store Operation:**
```
1. Try to store in read map (if exists)
2. If not in read map:
   ├─ Lock mutex
   ├─ Store in dirty map
   └─ Unlock mutex
3. Eventually promote dirty → read
```

**Load Operation:**
```
1. Check read map (lock-free!) ⚡
2. If found: return immediately
3. If not found:
   ├─ Lock mutex
   ├─ Check dirty map
   └─ Unlock mutex
```

### Why It's Fast for Our Test

In our write-heavy test with unique keys:
- **First write**: Goes to dirty map (locked)
- **Promotion**: After enough operations, dirty → read
- **Reduced contention**: Better than single global lock
- **Optimized internals**: Hand-tuned by Go team for common patterns

---

## Important Caveats

### 1. Type Safety
```go
// sync.Map - no type safety
m.Store(1, "value")         // OK
m.Store("key", 42)          // Also OK
value, _ := m.Load(1)       // Returns interface{}
str := value.(string)       // Manual type assertion needed!
```

### 2. No Built-in Length
```go
// Must iterate to count
count := 0
m.Range(func(k, v interface{}) bool {
    count++
    return true
})
```

### 3. Range Iterator Snapshot Inconsistency
- Range may or may not reflect concurrent modifications
- Not guaranteed to be a consistent snapshot
- Fine for most use cases, but be aware

---

## Real-World Use Cases

### When sync.Map Excels

**1. Cache Implementation**
```go
var cache sync.Map

// Write once (cache miss)
cache.Store(key, expensiveComputation())

// Read many times (cache hits) - lock-free!
value, ok := cache.Load(key)
```

**2. Connection Pool**
```go
var connections sync.Map

// Register connection (rare)
connections.Store(connID, conn)

// Look up connection (frequent) - fast!
conn, ok := connections.Load(connID)
```

**3. Registry Pattern**
```go
var handlers sync.Map

// Register handlers at startup
handlers.Store("GET /api/users", userHandler)

// Look up during request handling - very fast!
handler, ok := handlers.Load(request.Route)
```

---

## Key Lessons Learned

### 1. No Silver Bullet
- sync.Map is fastest in our test, but not always best choice
- Type safety loss is a real cost
- Code complexity varies by use case

### 2. Write-Heavy Results (This Experiment)
```
Winner: sync.Map (2.87ms) ⚡
- 2.57x faster than Mutex
- Built for concurrent access
- Optimized internals pay off
```

### 3. Access Pattern Matters
| Pattern | Best Choice |
|---------|-------------|
| 100% writes | sync.Map or Mutex |
| 90% reads, 10% writes | sync.Map or RWMutex |
| 50% reads, 50% writes | Mutex or sync.Map |
| Complex logic | Mutex (clearest) |

### 4. When in Doubt
1. Start with `sync.Mutex` (simple, predictable)
2. Profile if performance matters
3. Consider `sync.Map` for cache-like patterns
4. Use `sync.RWMutex` only for read-heavy proven scenarios

### 5. The Right Tool
- **sync.Map**: Specialized concurrent map, great for specific patterns
- **Mutex**: General-purpose, simple, predictable
- **RWMutex**: Read-heavy optimization
- All have valid use cases!

---

## Summary Comparison Chart

```
┌────────────────────────────────────────────────────────────┐
│                    PERFORMANCE SUMMARY                     │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  Write-Heavy (This Test):                                 │
│  ┌────────────────────────────────────────────┐           │
│  │ sync.Map:  ████████ 2.87ms  ⚡ FASTEST      │           │
│  │ Mutex:     ████████████████████ 7.37ms     │           │
│  │ RWMutex:   █████████████████████ 7.87ms    │           │
│  └────────────────────────────────────────────┘           │
│                                                            │
│  Estimated Read-Heavy (90% reads):                        │
│  ┌────────────────────────────────────────────┐           │
│  │ sync.Map:  ███ 1-2ms  ⚡⚡ FASTEST           │           │
│  │ RWMutex:   ██████ 3-4ms                    │           │
│  │ Mutex:     ████████████████████ 15-20ms    │           │
│  └────────────────────────────────────────────┘           │
│                                                            │
├────────────────────────────────────────────────────────────┤
│                    TRADE-OFFS                              │
├────────────────────────────────────────────────────────────┤
│  Simplicity:     Mutex ✅  >  RWMutex  >  sync.Map        │
│  Type Safety:    Mutex ✅  =  RWMutex  >  sync.Map ❌      │
│  Write Perf:     sync.Map ⚡  >  Mutex  >  RWMutex        │
│  Read Perf:      sync.Map ⚡  =  RWMutex  >  Mutex        │
│  Memory:         Mutex ✅  =  RWMutex  >  sync.Map        │
└────────────────────────────────────────────────────────────┘
```

---

## Conclusion

**sync.Map delivered exceptional performance** in our write-heavy concurrent test, running **2.57x faster** than Mutex and **2.74x faster** than RWMutex. However, this speed comes with tradeoffs:

**Pros:**
- ⚡ Significantly faster for concurrent operations
- Built-in concurrency safety
- Optimized for common patterns (cache, registry)
- No manual lock management

**Cons:**
- No type safety (uses `interface{}`)
- Higher memory overhead
- No direct length/iteration support
- Best for specific access patterns

**Final Recommendation:**
- **Default**: Use `sync.Mutex` (simple, safe, predictable)
- **Cache/Registry**: Use `sync.Map` (optimized for this)
- **Read-heavy**: Use `sync.RWMutex` (concurrent reads)
- **Always**: Profile before optimizing!

The best concurrent map approach depends on your specific access patterns, type safety requirements, and performance needs. Understanding these tradeoffs enables you to choose the right tool for each situation.
