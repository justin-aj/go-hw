package main

import (
	"encoding/json"
	"net/http"
	"time"
)

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

type setRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type kvResponse struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Version int64  `json:"version"`
}

func (h *Handler) Set(w http.ResponseWriter, r *http.Request) {
	// 1. Decode the request body
	var req setRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Key == "" {
		http.Error(w, "key cannot be empty", http.StatusBadRequest)
		return
	}

	// 2. Write to local store and get a version number
	var version int64
	if h.role == "leaderless" {
		version = h.store.SetTS(req.Key, req.Value)
	} else {
		version = h.store.Set(req.Key, req.Value)
	}

	// 3. Fan out to peers (leader: up to W-1 followers; leaderless: all peers)
	if h.role == "leader" && h.replicator != nil {
		waitsFor := h.w - 1 // leader counts as 1
		if waitsFor <= 0 {
			go h.replicator.FanOut(req.Key, req.Value, version, 0)
		} else {
			if waitsFor > len(h.replicator.followers) {
				waitsFor = len(h.replicator.followers)
			}
			h.replicator.FanOut(req.Key, req.Value, version, waitsFor)
		}
	}
	if h.role == "leaderless" && h.replicator != nil {
		h.replicator.FanOut(req.Key, req.Value, version, len(h.replicator.followers))
	}

	// 4. Return 201 with the version number
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Version int64 `json:"version"`
	}{version})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	// 1. Extract key from URL
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "key cannot be empty", http.StatusBadRequest)
		return
	}

	// 2. Follower / leaderless: sleep 50ms then return local value (R=1)
	if h.role == "follower" || h.role == "leaderless" {
		time.Sleep(50 * time.Millisecond)
		h.serveLocal(w, key)
		return
	}

	// 3. Leader R=1: return local value directly
	if h.r == 1 {
		h.serveLocal(w, key)
		return
	}

	// 4. Leader R>1: read local + (R-1) followers, return the highest version
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

func (h *Handler) Replicate(w http.ResponseWriter, r *http.Request) {
	// 1. Decode the replication message from the leader
	var req replicateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	// 2. Sleep 100ms (simulates follower write latency per spec)
	time.Sleep(100 * time.Millisecond)
	// 3. Write to local store (only if version is newer)
	h.store.SetWithVersion(req.Key, req.Value, req.Version)
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) LocalRead(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "key cannot be empty", http.StatusBadRequest)
		return
	}
	h.serveLocal(w, key)
}

func (h *Handler) serveLocal(w http.ResponseWriter, key string) {
	entry, ok := h.store.Get(key)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kvResponse{Key: key, Value: entry.Value, Version: entry.Version})
}
