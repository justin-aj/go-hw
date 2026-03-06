# Step III — Crash & Recovery Demo

Three standalone demos, each showing **one resilience pattern** from Sam Newman's *Building Microservices*. Each demo has a **broken mode** (no protection) and a **resilient mode** (pattern enabled) so you can compare.

```
step3-crash-recovery/
├── demo-failfast/          ← Pattern 1: Fail-Fast
├── demo-bulkhead/          ← Pattern 2: Bulkhead
└── demo-circuitbreaker/    ← Pattern 3: Circuit Breaker
```

---

## The Shared Problem

All three demos use the same product-search service with a fault injection:
- Normal query (`?q=beauty`) → searches products → returns in ~7ms
- Faulty query (`?q=__slow`) → simulates a broken downstream call → sleeps **5 seconds**

Without protection, the `__slow` queries pile up and bring down the entire service.

---

## Demo 1: Fail-Fast (`demo-failfast/`)

**Concept:** *"Don't wait forever for a slow response."*

Sets a **500ms timeout** on search requests. If the handler doesn't finish in time → killed immediately → `503`.

| | Broken | With Fail-Fast |
|--|--------|----------------|
| `__slow` latency | 5,000ms (full sleep) | 500ms (killed by timer) |
| Resources wasted | 5s per stuck request | 0.5s per stuck request |

```bash
cd demo-failfast
docker compose --profile resilient up --build
locust -f locustfile.py --headless -u 50 -r 10 --run-time 60s --host http://localhost:8080
```

**Key file:** `resilience/failfast.go` — races the handler against a timer using `context.WithTimeout`.

---

## Demo 2: Bulkhead (`demo-bulkhead/`)

**Concept:** *"Limit how many requests can be inside at the same time."*

Only **10 requests** can run concurrently. Request #11 → rejected instantly → `429`.

| | Broken | With Bulkhead |
|--|--------|---------------|
| Concurrent requests | Unlimited → server overloaded | Max 10 → excess rejected |
| Server stability | Crashes under load | Stays alive |

```bash
cd demo-bulkhead
docker compose --profile resilient up --build
locust -f locustfile.py --headless -u 50 -r 10 --run-time 60s --host http://localhost:8080
```

**Key file:** `resilience/bulkhead.go` — uses a buffered Go channel as a semaphore (parking lot with 10 spaces).

---

## Demo 3: Circuit Breaker (`demo-circuitbreaker/`)

**Concept:** *"Stop calling something that's clearly broken."*

After **5 consecutive failures**, the circuit trips OPEN → `__slow` queries are skipped instantly (no 5s sleep). Normal queries (`?q=beauty`) **bypass the circuit breaker entirely** and always work.

| | Broken | With Circuit Breaker |
|--|--------|----------------------|
| `__slow` after 5 failures | Still sleeps 5s every time | Skipped instantly (0.001ms) |
| Normal queries | Starved by slow queries | Always work (bypass CB) |
| Recovery | Never | Auto-probes after 10s cooldown |

```bash
cd demo-circuitbreaker
docker compose --profile resilient up --build
locust -f locustfile.py --headless -u 50 -r 10 --run-time 60s --host http://localhost:8080
```

**Key file:** `resilience/circuitbreaker.go` — three states: CLOSED → OPEN → HALF-OPEN.

---

## How They're Different

| | Fail-Fast | Bulkhead | Circuit Breaker |
|--|-----------|----------|-----------------|
| **Question** | "How long should I wait?" | "How many at once?" | "Should I even try?" |
| **Trigger** | Time (500ms) | Count (10 concurrent) | Failures (5 in a row) |
| **Response** | 503 (timeout) | 429 (too busy) | Skips the downstream call |
| **Protects against** | Slow responses | Resource exhaustion | Repeated failures |

---

## How It Works: Broken vs Resilient

### Broken Mode (no protection)

```
__slow request #1:  → time.Sleep(5s) → 5,000ms wasted → goroutine stuck
__slow request #2:  → time.Sleep(5s) → 5,000ms wasted → goroutine stuck
...hundreds pile up...
Normal request:     → BLOCKED — server choked → latency spikes → CRASH
```

### With All Three Patterns (combined)

```
__slow #1:  → Bulkhead: slot ✓ → Fail-Fast: 500ms timer → sleeps → killed at 500ms → failureCount=1
__slow #2–4: Same → failureCount = 2, 3, 4
__slow #5:  → failureCount=5 → CIRCUIT TRIPS OPEN!
__slow #6+: → Circuit Breaker: OPEN → skipped instantly (0.001ms) → 503
Normal:     → Bulkhead: slot ✓ → Fail-Fast: finishes in 7ms → BYPASSES circuit breaker → 200 OK
```

---

## Prerequisites

- Docker & Docker Compose
- Python 3 with `locust` (`pip install locust`)

Each demo has `--profile broken` and `--profile resilient` modes.
