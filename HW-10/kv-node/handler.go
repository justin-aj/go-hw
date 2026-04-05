package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handler wires the KVStore to HTTP endpoints.
// role controls which artificial delays are applied:
//
//	"follower" — 100ms on set (simulates write latency), 50ms on get (simulates read latency)
//	"leader"   — no delays here; the 200ms-per-follower delay lives in the replication layer
//	anything else — no delays (useful for standalone / unit-test mode)
type Handler struct {
	store *KVStore
	role  string
}

func NewHandler(store *KVStore, role string) *Handler {
	return &Handler{store: store, role: role}
}

// --- request / response types ---

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
// Body: {"key":"...", "value":"..."}
// Returns 201 on success, 400 if key is empty or body is malformed.
//
// Follower delay: 100ms before writing (spec: "Follower should sleep for 100ms
// when it receives the update before responding").
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

	if h.role == "follower" {
		time.Sleep(100 * time.Millisecond)
	}

	h.store.Set(req.Key, req.Value)
	w.WriteHeader(http.StatusCreated)
}

// Get handles GET /get/{key}.
// Returns 200 + JSON body on hit, 404 on miss.
//
// Follower delay: 50ms before reading (spec: "any Follower that receives a read
// request from the Leader should sleep for 50ms before responding").
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "key cannot be empty", http.StatusBadRequest)
		return
	}

	if h.role == "follower" {
		time.Sleep(50 * time.Millisecond)
	}

	entry, ok := h.store.Get(key)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kvResponse{Key: key, Value: entry.Value, Version: entry.Version})
}

// LocalRead handles GET /local_read/{key}.
// Identical to Get but with NO delays and NO replication forwarding.
// Used by unit tests to inspect a node's local state mid-replication window.
func (h *Handler) LocalRead(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "key cannot be empty", http.StatusBadRequest)
		return
	}

	entry, ok := h.store.Get(key)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kvResponse{Key: key, Value: entry.Value, Version: entry.Version})
}
