# Two Instances Experiment: Understanding Stateful Applications

## Overview

This experiment demonstrates a critical problem in distributed systems: **state synchronization across multiple instances**. By deploying the same Go web service to two separate EC2 instances and testing data consistency, we reveal why in-memory state doesn't scale in distributed environments.

---

## Experiment Setup

### Infrastructure
- **2 EC2 Instances** (t3.small, Amazon Linux 2023)
- **Same Docker Container** running on both instances
- **No shared database** - each instance stores data in-memory
- **Security Group** allows HTTP traffic on port 8080

### Instance Details
- **Instance 1**: `44.192.75.222`
- **Instance 2**: `13.223.206.46`

### Initial Data
Both instances start with 3 albums in memory:
```json
[
  {"id": "1", "title": "Blue Train", "artist": "John Coltrane", "price": 56.99},
  {"id": "2", "title": "Jeru", "artist": "Gerry Mulligan", "price": 17.99},
  {"id": "3", "title": "Sarah Vaughan and Clifford Brown", "artist": "Sarah Vaughan", "price": 39.99}
]
```

---

## Test Procedure

### Step 1: Check Initial State
```python
GET http://44.192.75.222:8080/albums  # Instance 1
GET http://13.223.206.46:8080/albums  # Instance 2
```
**Result**: Both return 3 albums ✅

### Step 2: Add Album to Instance 2
```python
POST http://13.223.206.46:8080/albums
{
  "id": "4",
  "title": "The Modern Sound of Betty Carter",
  "artist": "Betty Carter",
  "price": 49.99
}
```
**Result**: Instance 2 confirms album added ✅

### Step 3: Check Final State
```python
GET http://44.192.75.222:8080/albums  # Instance 1
GET http://13.223.206.46:8080/albums  # Instance 2
```

---

## Results

### Instance 1 Response (after POST)
```json
[
  {"id": "1", "title": "Blue Train", "artist": "John Coltrane", "price": 56.99},
  {"id": "2", "title": "Jeru", "artist": "Gerry Mulligan", "price": 17.99},
  {"id": "3", "title": "Sarah Vaughan and Clifford Brown", "artist": "Sarah Vaughan", "price": 39.99}
]
```
**Count**: 3 albums (unchanged) ❌

### Instance 2 Response (after POST)
```json
[
  {"id": "1", "title": "Blue Train", "artist": "John Coltrane", "price": 56.99},
  {"id": "2", "title": "Jeru", "artist": "Gerry Mulligan", "price": 17.99},
  {"id": "3", "title": "Sarah Vaughan and Clifford Brown", "artist": "Sarah Vaughan", "price": 39.99},
  {"id": "4", "title": "The Modern Sound of Betty Carter", "artist": "Betty Carter", "price": 49.99}
]
```
**Count**: 4 albums ✅

---

## What Happened?

### The Problem: Isolated State

Each instance maintains its own **in-memory state**:

```go
// In main.go
var albums = []album{
    {ID: "1", Title: "Blue Train", ...},
    {ID: "2", Title: "Jeru", ...},
    {ID: "3", Title: "Sarah Vaughan...", ...},
}
```

When a POST request adds an album:
1. The data is appended to the **local** `albums` slice
2. This change exists **only in that instance's memory**
3. Other instances have **no knowledge** of this change
4. Each instance operates in **complete isolation**

### Visual Representation

```
┌─────────────────────┐          ┌─────────────────────┐
│   Instance 1        │          │   Instance 2        │
│   44.192.75.222     │          │   13.223.206.46     │
├─────────────────────┤          ├─────────────────────┤
│ Memory:             │          │ Memory:             │
│  albums[] = [       │          │  albums[] = [       │
│    Album 1,         │          │    Album 1,         │
│    Album 2,         │          │    Album 2,         │
│    Album 3          │          │    Album 3,         │
│  ]                  │   POST   │    Album 4 ⭐      │
│                     │  ────►   │  ]                  │
│ Still has 3 albums  │          │ Now has 4 albums    │
└─────────────────────┘          └─────────────────────┘
```

---

## Why This Is a Problem

### 1. **Data Inconsistency**
Different instances return different results for the same API endpoint. Users get inconsistent experiences depending on which server handles their request.

### 2. **Load Balancer Issues**
```
User → POST to Instance 2 (adds album)
     ↓
User → GET from Instance 1 (album not found!)
```
A load balancer distributing requests randomly will cause confusion and bugs.

### 3. **Lost Data on Instance Restart**
If Instance 2 crashes and restarts, the Betty Carter album is **permanently lost** - it was never persisted anywhere.

### 4. **No Scalability**
You can't add more instances to handle increased load without solving the state synchronization problem first.

---

## Solutions

### Option 1: Shared Database ⭐ (Recommended)
Replace in-memory storage with a centralized database:

```go
// Instead of:
var albums = []album{...}

// Use:
db, _ := sql.Open("postgres", "connectionString")
// Query database for each request
```

**Pros**: 
- Single source of truth
- Data persists across restarts
- All instances see same data

**Cons**: 
- Database becomes bottleneck
- Additional infrastructure
- Network latency

### Option 2: Distributed Cache
Use Redis or Memcached for shared, fast in-memory storage:

```go
redis := redis.NewClient(&redis.Options{Addr: "cache:6379"})
// Store/retrieve albums from Redis
```

**Pros**:
- Fast (in-memory)
- Shared across instances
- Built-in replication

**Cons**:
- Still need persistence layer
- Cache invalidation complexity

### Option 3: Session Affinity / Sticky Sessions
Load balancer routes all requests from a user to the same instance.

**Pros**:
- Simple to implement
- No code changes needed

**Cons**:
- Poor load distribution
- Doesn't solve data loss
- Not truly scalable

### Option 4: Stateless Design ⭐⭐ (Best Practice)
Move all state to external services (databases, caches, queues):

```
┌─────────┐    ┌─────────┐    ┌─────────┐
│Instance1│───▶│         │◀───│Instance2│
└─────────┘    │Database │    └─────────┘
               │         │
               └─────────┘
```

**Pros**:
- Infinitely scalable
- No state synchronization issues
- Instances are interchangeable

**Cons**:
- More complex architecture
- External dependencies

---

## Real-World Implications

### E-Commerce Example
Imagine this scenario with shopping carts:
1. User adds item on Instance 1
2. Load balancer routes next request to Instance 2
3. Cart appears empty!
4. User abandons purchase 💰❌

### Banking Example
1. User transfers $100 on Instance 1
2. Balance check goes to Instance 2
3. Transfer not reflected!
4. User thinks transaction failed and tries again
5. Double charge! 🏦❌

---

## Key Takeaways

1. **In-memory state doesn't scale** in distributed systems
2. **Stateless applications** are easier to scale horizontally
3. **External data stores** are essential for consistency
4. **Load balancers** require careful consideration of state
5. **CAP Theorem** applies: Consistency, Availability, Partition tolerance

---

## CAP Theorem Context

This experiment demonstrates the **Consistency** challenge:

- **Consistency (C)**: All nodes see the same data at the same time
  - ❌ Failed in our experiment
  
- **Availability (A)**: Every request gets a response
  - ✅ Both instances respond successfully
  
- **Partition Tolerance (P)**: System works despite network failures
  - ✅ Instances operate independently

**Tradeoff**: Our system chose Availability over Consistency. In production, you typically choose 2 of 3, and in cloud environments, Partition tolerance is mandatory, so you choose between C and A.

---

## Interview Discussion Points

### Technical Understanding
- Explain the difference between stateful and stateless applications
- Describe database replication strategies (master-slave, multi-master)
- Discuss eventual consistency vs strong consistency

### System Design
- How would you architect a scalable web service?
- What database would you choose and why?
- How do you handle database failover?

### Trade-offs
- Performance vs consistency
- Cost vs reliability
- Complexity vs simplicity

### Real-World Examples
- How does Netflix/Amazon/YouTube handle this?
- Microservices architecture
- Event-driven systems

---

## Next Steps to Fix This

1. **Add PostgreSQL/MySQL**: Deploy a database instance
2. **Update Go code**: Replace in-memory slice with database queries
3. **Connection pooling**: Handle multiple concurrent connections
4. **Caching layer**: Add Redis for frequently accessed data
5. **Read replicas**: Scale read operations
6. **Write queues**: Handle write operations asynchronously

---

## Conclusion

This simple experiment reveals a fundamental challenge in distributed systems: **maintaining consistent state across multiple independent instances**. While in-memory storage works fine for a single instance, it breaks down immediately when you scale horizontally.

The solution is to embrace **stateless architecture** with external data stores, treating application instances as disposable, interchangeable workers that all read from and write to a shared source of truth.

**Remember**: If you can't afford to lose it, don't store it in memory! 🎯
