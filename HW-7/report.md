# HW-7 Report: When Your Startup's Flash Sale Almost Failed

## Part II: Synchronous vs Asynchronous Order Processing

---

## Phase 1: Synchronous System Baseline

### Architecture

```
POST /orders/sync → acquire payment slot (buffered channel, cap=1) → sleep 3s → release → 200 OK
```

The bottleneck is intentional: a Go buffered channel of size 1 serializes all payment processing, simulating a single-threaded payment processor. No matter how many users hit the endpoint concurrently, only one order can be "in payment" at a time.

---

### Test 1: Normal Operations (5 users, 30s, spawn rate 5/s)

**Results:**

| Metric | Value |
|---|---|
| Total Requests | 9 |
| Failures | 0 (0%) |
| Min Response Time | 3,003 ms |
| Median (P50) | 15,000 ms |
| P95 | 15,000 ms |
| Average | 11,670 ms |
| Max | 15,004 ms |
| Throughput | 0.4 req/s |

**Chart Observations:**
- RPS climbed from 0 to ~0.4 as users spawned, then held flat — the hard ceiling of 1 order / 3s
- Response times climbed steeply as each user joined the queue: user 1 waited 3s, user 5 waited 15s
- P50 and P95 converged at 15,000ms once the queue was full — every user was waiting for all others ahead
- 0 failures throughout — the system never crashed, it just made customers wait

**Key insight:** Min = 3,003ms (first user, no queue). Max = 15,004ms ≈ 5 × 3s (last user waited for 4 others). The queue length × 3s formula holds exactly.

---

### Test 2: Flash Sale (20 users, 60s, spawn rate 20/s)

**Results:**

| Metric | Value |
|---|---|
| Total Requests | 19 |
| Failures | 0 (0%) |
| Min Response Time | 3,017 ms |
| Median (P50) | 30,000 ms |
| P95 | 57,000 ms |
| P99 | 57,000 ms |
| Average | 30,017 ms |
| Max | 57,018 ms |
| Throughput | 0.4 req/s |

**Chart Observations:**
- RPS stayed flat at 0.4 req/s — identical to normal operations. 20 users changed nothing about throughput.
- Response times spiked immediately to 50,000–60,000ms as all 20 users queued behind the single payment slot
- P50 = 30,000ms (~10 users deep × 3s), P95 = 57,000ms (~19 users deep × 3s)
- Max = 57,018ms ≈ 19 × 3s — the last user in queue waited for all 19 ahead of them
- 0 failures — the system stayed alive, but customers waited up to 57 seconds for a response

**Key insight:** Throughput did not change (0.4 req/s) — only customer wait time exploded. From the customer's perspective this system had "failed" long before any HTTP error occurred. A 57-second checkout is an abandoned cart.

---

## Phase 2: Bottleneck Analysis

### The Math

| Variable | Value |
|---|---|
| Payment processor speed | 1 order / 3s = **0.33 orders/s** |
| Flash sale demand | ~60 orders/s |
| Actual throughput (measured) | 0.33 orders/s (unchanged) |
| Orders lost per second | 60 - 0.33 = **59.67/s** |
| Queue backlog after 60s | 59.67 × 60 = **~3,580 orders** |
| Time to drain with 1 worker | 3,580 × 3s = **10,740s ≈ 179 minutes** |

### Why more users don't help

The buffered channel (`make(chan struct{}, 1)`) is a hard serialization point. Adding concurrent users only adds goroutines waiting to acquire the slot — it does not increase throughput. This is the classic **fixed-capacity bottleneck**: throughput = min(supply, demand), where supply = 0.33/s regardless of demand.

### Workers needed to match flash sale demand

```
Workers = demand × processing_time
        = 60 orders/s × 3s/order
        = 180 goroutines
```

This is why Phase 5 scales the processor goroutines — each goroutine can process 1 order per 3s independently, so 180 goroutines × 0.33/s = ~60 orders/s throughput.

---

## Phase 3: Async Solution

> *To be filled after AWS deployment*

## Phase 4: Queue Problem

> *To be filled after AWS deployment*

## Phase 5: Worker Scaling

> *To be filled after AWS deployment*

---

## Analysis Questions

### How many times more orders did async accept vs sync?

> *To be filled after Phase 3 results*

### What causes queue buildup and how do you prevent it?

Queue buildup occurs when the **processing rate < arrival rate**. With 1 worker at 0.33 orders/s and 60 orders/s incoming, the queue grows at 59.67 messages/second. Prevention strategies:
1. **Scale workers** — add goroutines until processing rate >= arrival rate (need 180 for 60/s)
2. **Auto-scaling** — watch `ApproximateNumberOfMessagesVisible` in CloudWatch and scale ECS tasks when it climbs
3. **Lambda** (Part III) — AWS auto-scales concurrency automatically, no manual tuning

### When would you choose sync vs async in production?

**Choose sync when:**
- The operation must complete before you can respond (e.g., checking inventory before confirming a seat)
- Latency is low enough that users will wait (< 500ms)
- Simplicity matters more than throughput

**Choose async when:**
- Processing takes seconds (payment verification, image resizing, email sending)
- You want to decouple acceptance rate from processing rate
- You need resilience — SQS holds messages if the processor crashes
