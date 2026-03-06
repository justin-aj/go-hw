// ===================================================================
// BULKHEAD PATTERN
// ===================================================================
//
// CONCEPT: "Limit how many requests can be inside at the same time."
//
// Problem:  500 requests arrive at once. The server tries to handle
//           all 500 simultaneously. It runs out of memory and crashes.
//
// Solution: Set a maximum capacity (100). If 100 requests are already
//           being processed, reject new ones immediately with 429.
//
// In our demo: Max 100 concurrent search requests. This prevents
//              __slow queries from consuming all server resources.
//
// Implementation: A Go buffered channel acts as a "parking lot."
//   - Channel has 100 spaces.
//   - Each request tries to "park" (put an item in the channel).
//   - If full → rejected. If space → parked, served, then leaves.
// ===================================================================

package resilience

import (
	"encoding/json"
	"log"
	"net/http"
)

// Bulkhead limits concurrent requests using a channel-based semaphore.
type Bulkhead struct {
	next          http.Handler
	sem           chan struct{} // the "parking lot" — each item = one occupied slot
	totalRejected int
}

// NewBulkhead creates a bulkhead with maxConcurrent slots.
func NewBulkhead(next http.Handler, maxConcurrent int) *Bulkhead {
	return &Bulkhead{
		next: next,
		sem:  make(chan struct{}, maxConcurrent),
	}
}

// ServeHTTP runs on every request. It checks if there's a free slot.
func (b *Bulkhead) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	case b.sem <- struct{}{}:
		// Got a slot. Process the request, then free the slot.
		defer func() { <-b.sem }() // run this at the very end to free up the slot
		b.next.ServeHTTP(w, r)

	default:
		// All slots full. Reject immediately.
		b.totalRejected++
		log.Printf("[BULKHEAD] All %d slots busy — rejecting (total: %d)\n",
			cap(b.sem), b.totalRejected)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(429)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "too many concurrent requests",
			"pattern": "bulkhead",
		})
	}
}

// Stats returns current state for /metrics.
type BulkheadStats struct {
	MaxSlots      int `json:"max_slots"`
	InFlight      int `json:"in_flight"`
	TotalRejected int `json:"total_rejected"`
}

func (b *Bulkhead) Stats() BulkheadStats {
	return BulkheadStats{
		MaxSlots:      cap(b.sem),
		InFlight:      len(b.sem),
		TotalRejected: b.totalRejected,
	}
}
