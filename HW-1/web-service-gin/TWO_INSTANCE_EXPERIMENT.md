# Two Instance Experiment - Stateful Application Problem

## Overview
This experiment demonstrates a fundamental problem in distributed systems: **state synchronization across multiple instances**.

## Experiment Setup

### Infrastructure
- **2 EC2 Instances** deployed using Terraform
  - Instance 1: `44.192.75.222`
  - Instance 2: `13.223.206.46`
- Both running identical Docker containers with the Go web service
- Each instance stores album data in-memory (in a Go slice)

### Test Script
The test performs the following operations:
1. GET albums from both instances (initial state)
2. POST a new album (Betty Carter) to Instance 2
3. GET albums from both instances again (final state)

## Experiment Results

```
Starting data test...
Instance 1 response: [
    {
        "id": "1",
        "title": "Blue Train",
        "artist": "John Coltrane",
        "price": 56.99
    },
    {
        "id": "2",
        "title": "Jeru",
        "artist": "Gerry Mulligan",
        "price": 17.99
    },
    {
        "id": "3",
        "title": "Sarah Vaughan and Clifford Brown",
        "artist": "Sarah Vaughan",
        "price": 39.99
    }
]


and...
Instance 2 response: [
    {
        "id": "1",
        "title": "Blue Train",
        "artist": "John Coltrane",
        "price": 56.99
    },
    {
        "id": "2",
        "title": "Jeru",
        "artist": "Gerry Mulligan",
        "price": 17.99
    },
    {
        "id": "3",
        "title": "Sarah Vaughan and Clifford Brown",
        "artist": "Sarah Vaughan",
        "price": 39.99
    }
]


and adding...
Instance 2 response: {
    "id": "4",
    "title": "The Modern Sound of Betty Carter",
    "artist": "Betty Carter",
    "price": 49.99
}
Instance 1 response: [
    {
        "id": "1",
        "title": "Blue Train",
        "artist": "John Coltrane",
        "price": 56.99
    },
    {
        "id": "2",
        "title": "Jeru",
        "artist": "Gerry Mulligan",
        "price": 17.99
    },
    {
        "id": "3",
        "title": "Sarah Vaughan and Clifford Brown",
        "artist": "Sarah Vaughan",
        "price": 39.99
    }
]


and...
Instance 2 response: [
    {
        "id": "1",
        "title": "Blue Train",
        "artist": "John Coltrane",
        "price": 56.99
    },
    {
        "id": "2",
        "title": "Jeru",
        "artist": "Gerry Mulligan",
        "price": 17.99
    },
    {
        "id": "3",
        "title": "Sarah Vaughan and Clifford Brown",
        "artist": "Sarah Vaughan",
        "price": 39.99
    },
    {
        "id": "4",
        "title": "The Modern Sound of Betty Carter",
        "artist": "Betty Carter",
        "price": 49.99
    }
]
uhoh... what happened?
```

## What Happened?

### Initial State (Before POST)
✅ **Both instances have identical data**: 3 albums

### After POST to Instance 2
- ✅ **Instance 2**: Now has **4 albums** (including Betty Carter)
- ❌ **Instance 1**: Still has only **3 albums** (no Betty Carter)

## Why Did This Happen?

### The Problem: In-Memory State
Each instance stores data **independently in its own memory**:

```go
var albums = []album{
    {ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
    {ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
    {ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}
```

When you POST to Instance 2:
1. Instance 2 appends the new album to **its own slice**
2. Instance 1 has **no idea** this happened
3. The instances have **no shared state**

### Visualization

```
Before POST:
┌──────────────┐         ┌──────────────┐
│  Instance 1  │         │  Instance 2  │
│              │         │              │
│  Albums: 3   │         │  Albums: 3   │
└──────────────┘         └──────────────┘

After POST to Instance 2:
┌──────────────┐         ┌──────────────┐
│  Instance 1  │         │  Instance 2  │
│              │         │              │
│  Albums: 3   │    ❌    │  Albums: 4   │
└──────────────┘         └──────────────┘
      ↑                         ↑
      |                         |
   No Change               Added Betty
```

## Real-World Implications

### With a Load Balancer
```
User → Load Balancer → Instance 1 or Instance 2?
```

**Scenario:**
1. User POSTs album → routed to Instance 2 ✅
2. User GETs albums → routed to Instance 1 ❌
3. User sees the album is missing! 😱

### The User Experience
- **Data appears to "disappear"** depending on which instance handles the request
- **Inconsistent responses** from the same API
- **Lost data** if an instance crashes (all in-memory data is gone)

## Solutions

### 1. **Shared Database** (Recommended)
```
┌──────────────┐         ┌──────────────┐
│  Instance 1  │────┐    │  Instance 2  │────┐
└──────────────┘    │    └──────────────┘    │
                    ↓                         ↓
              ┌────────────────────┐
              │   PostgreSQL DB    │
              │   (Shared State)   │
              └────────────────────┘
```

**Pros:**
- True single source of truth
- Data persists across restarts
- ACID guarantees

**Cons:**
- Database becomes bottleneck
- Need to handle DB failures

### 2. **Distributed Cache (Redis, Memcached)**
```
┌──────────────┐         ┌──────────────┐
│  Instance 1  │────┐    │  Instance 2  │────┐
└──────────────┘    │    └──────────────┘    │
                    ↓                         ↓
              ┌────────────────────┐
              │   Redis Cluster    │
              │   (Shared Cache)   │
              └────────────────────┘
```

**Pros:**
- Very fast (in-memory)
- Good for session data
- Can handle high throughput

**Cons:**
- Not persistent by default
- Need cache invalidation strategy

### 3. **Sticky Sessions (Session Affinity)**
```
User → Load Balancer → Always Route to Same Instance
```

**Pros:**
- Simple to implement
- Works with stateful apps

**Cons:**
- Poor load distribution
- User stuck to failed instance
- Not truly scalable

### 4. **Event Sourcing / Message Queue**
```
Instance 2 → POST → Kafka/RabbitMQ → Instance 1 subscribes
```

**Pros:**
- Eventual consistency
- Audit trail of all changes
- Decoupled architecture

**Cons:**
- Complex to implement
- Eventual consistency (not immediate)

## Key Distributed Systems Concepts

### 1. **Stateless vs Stateful**
- **Stateless**: No server-side state, all data in DB/cache
- **Stateful**: Server holds state in memory (our current problem)

### 2. **CAP Theorem**
Can only guarantee 2 of 3:
- **Consistency**: All nodes see same data
- **Availability**: System always responds
- **Partition Tolerance**: Works despite network failures

### 3. **Horizontal Scaling**
Adding more instances doesn't help if state isn't shared!

## Interview Discussion Points

1. **Why is this a problem?**
   - Data inconsistency across instances
   - Poor user experience
   - Data loss on crashes

2. **How would you fix it?**
   - Add PostgreSQL/MySQL database
   - Use Redis for caching
   - Implement proper state management

3. **Trade-offs?**
   - Database adds latency
   - Single point of failure (need replication)
   - Increased complexity

4. **Production patterns?**
   - Stateless application servers
   - Separate data tier
   - Read replicas for scaling
   - Cache layer (Redis)
   - CDN for static content

5. **What about sessions?**
   - JWT tokens (stateless)
   - Redis session store
   - Database session table

## Conclusion

**The Lesson**: Storing state in application memory doesn't scale in distributed systems!

**Best Practice**: Always design applications to be **stateless** with external state management (databases, caches, etc.)

This experiment perfectly demonstrates why modern architectures separate:
- **Compute Layer** (stateless app servers)
- **Data Layer** (databases, caches)
- **Load Balancing Layer** (route traffic)

## Related Concepts to Explore

- **Microservices Architecture**
- **Database Replication** (Master-Slave, Multi-Master)
- **Caching Strategies** (Write-through, Write-behind, Cache-aside)
- **Consistency Models** (Strong, Eventual, Causal)
- **Distributed Transactions** (2PC, Saga pattern)
- **Service Mesh** (Istio, Linkerd)
- **Circuit Breakers** (Resilience patterns)

---

**Date**: January 26, 2026  
**Infrastructure**: 2x t3.small EC2 instances (Amazon Linux 2023)  
**Application**: Go web service with Gin framework  
**Deployment**: Docker containers via Terraform
