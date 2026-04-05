package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handler wires the KVStore to HTTP endpoints.
//
// role controls which delays and replication behaviors are applied:
//
//	"leader"     — fans out writes to followers, reads from R nodes
//	"follower"   — sleeps 50ms on /get (simulates read latency)
//	"leaderless" — any node can receive a write and become write coordinator;
//	               fans out to all peers (W=N), reads are local only (R=1)
//	"standalone" — no delays, no replication (unit tests / local dev)
//
// w and r are the write/read quorums (meaningful only on the leader).
// replicator is non-nil when role == "leader" or role == "leaderless".
type Handler struct {
	store      *KVStore
	role       string
	w          int
	r          int
	replicator *Replicator
}

func NewHandler(store *KVStore, role string, w, r int, replicator *Replicator) *Handler {
	return &Handler{store: store, role: role, w: w, r: r, replicator: replicator}
}

// --- shared request / response types ---

type setRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type kvResponse struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Version int64  `json:"version"`
}

// --- handlers ---

// Set handles POST /set.
//
// On leader:
//   - Write locally (gets a version number from the atomic counter).
//   - W=1: return 201 immediately, replicate async to all followers.
//   - W=3: block until 2 followers confirm (leader + 2 = 3 total), then 201.
//   - W=5: block until all 4 followers confirm, then 201.
//
// On leaderless node (write coordinator):
//   - This node becomes the coordinator for this write.
//   - Write locally, then fan out to ALL peers and wait for all (W=N).
//   - Only return 201 after every peer confirms.
//   - Inconsistency window: peers that haven't been reached yet will serve
//     stale data if read during the ~300ms replication window.
//
// On follower: write locally, return 201 (not called directly in Phase 2).
func (h *Handler) Set(w http.ResponseWriter, r *http.Request) {
	var req setRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Key == "" {
		http.Error(w, "key cannot be empty", http.StatusBadRequest)
		return
	}

	// Leaderless nodes use a nanosecond timestamp as the version so that
	// writes from different coordinators are comparable (last-write-wins).
	// Leader-follower nodes use the leader's atomic counter instead.
	var version int64
	if h.role == "leaderless" {
		version = h.store.SetTS(req.Key, req.Value)
	} else {
		version = h.store.Set(req.Key, req.Value)
	}

	if h.role == "leader" && h.replicator != nil {
		waitsFor := h.w - 1 // leader itself counts as 1
		if waitsFor <= 0 {
			// W=1: client gets 201 immediately; replication runs in the background.
			go h.replicator.FanOut(req.Key, req.Value, version, 0)
		} else {
			if waitsFor > len(h.replicator.followers) {
				waitsFor = len(h.replicator.followers)
			}
			h.replicator.FanOut(req.Key, req.Value, version, waitsFor)
		}
	}

	if h.role == "leaderless" && h.replicator != nil {
		// W=N: wait for every peer to confirm before returning.
		h.replicator.FanOut(req.Key, req.Value, version, len(h.replicator.followers))
	}

	// Return the version so the load tester can track staleness.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Version int64 `json:"version"`
	}{version})
}

// Get handles GET /get/{key}.
//
// On leader with R=1: read from local store only (W=5 guarantees it is up to date).
// On leader with R>1: read from local store + (R-1) followers concurrently,
//
//	return the entry with the highest version number.
//
// On follower: sleep 50ms (spec), then read local store.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "key cannot be empty", http.StatusBadRequest)
		return
	}

	if h.role == "follower" || h.role == "leaderless" {
		// Followers and leaderless nodes: R=1, just return local value.
		// Leaderless nodes sleep 50ms to simulate read latency consistently.
		time.Sleep(50 * time.Millisecond)
		h.serveLocal(w, key)
		return
	}

	// Leader: R=1 → serve locally (fast path).
	if h.r == 1 {
		h.serveLocal(w, key)
		return
	}

	// Leader: R>1 → collect local + follower reads, return max version.
	localEntry, localOk := h.store.Get(key)
	followerBest, followerFound := h.replicator.ReadBest(key, h.r-1)

	best := kvResponse{Key: key, Version: -1}
	if localOk && localEntry.Version > best.Version {
		best.Value = localEntry.Value
		best.Version = localEntry.Version
	}
	if followerFound && followerBest.Version > best.Version {
		best.Value = followerBest.Value
		best.Version = followerBest.Version
	}

	if best.Version < 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(best)
}

// Replicate handles POST /replicate.
// Called internally by the leader to propagate a write to this node.
// Sleeps 100ms before writing (spec: "Follower should sleep 100ms when it
// receives the update before responding").
// Uses SetWithVersion so followers always store the leader-assigned version
// and out-of-order messages don't overwrite newer data.
func (h *Handler) Replicate(w http.ResponseWriter, r *http.Request) {
	var req replicateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	time.Sleep(100 * time.Millisecond)
	h.store.SetWithVersion(req.Key, req.Value, req.Version)
	w.WriteHeader(http.StatusCreated)
}

// LocalRead handles GET /local_read/{key}.
// No delays, no forwarding. Used by unit tests to inspect raw node state
// mid-replication window to catch stale reads.
func (h *Handler) LocalRead(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "key cannot be empty", http.StatusBadRequest)
		return
	}
	h.serveLocal(w, key)
}

// serveLocal reads from the local store and writes the HTTP response.
func (h *Handler) serveLocal(w http.ResponseWriter, key string) {
	entry, ok := h.store.Get(key)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kvResponse{Key: key, Value: entry.Value, Version: entry.Version})
}
