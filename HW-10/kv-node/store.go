package main

import (
	"sync"
	"sync/atomic"
	"time"
)

// Entry holds a value and a logical version number.
// Version is a monotonically increasing counter assigned by the leader,
// so every write gets a unique, comparable version across all nodes.
type Entry struct {
	Value   string
	Version int64
}

// KVStore is a thread-safe in-memory key-value store.
type KVStore struct {
	mu      sync.RWMutex
	data    map[string]Entry
	counter int64 // global version counter, incremented atomically on every leader write
}

func NewKVStore() *KVStore {
	return &KVStore{data: make(map[string]Entry)}
}

// Set stores value under key with an auto-incremented version. Called by the leader.
func (s *KVStore) Set(key, value string) int64 {
	version := atomic.AddInt64(&s.counter, 1)
	s.mu.Lock()
	s.data[key] = Entry{Value: value, Version: version}
	s.mu.Unlock()
	return version
}

// SetTS stores value under key using the current Unix nanosecond timestamp as
// the version. Used by leaderless write coordinators so that each independent
// write gets a globally comparable version without a shared counter.
// Last-write-wins: only updates if the incoming timestamp is newer.
func (s *KVStore) SetTS(key, value string) int64 {
	version := time.Now().UnixNano()
	s.mu.Lock()
	if existing, ok := s.data[key]; !ok || version > existing.Version {
		s.data[key] = Entry{Value: value, Version: version}
	}
	s.mu.Unlock()
	return version
}

// SetWithVersion stores value under key with the leader-assigned version.
// Only updates if the incoming version is newer (handles out-of-order replication).
// Called by followers receiving a replication message.
func (s *KVStore) SetWithVersion(key, value string, version int64) {
	s.mu.Lock()
	if existing, ok := s.data[key]; !ok || version > existing.Version {
		s.data[key] = Entry{Value: value, Version: version}
	}
	s.mu.Unlock()
}

// Get retrieves the entry for key. Returns (entry, true) or (zero, false).
func (s *KVStore) Get(key string) (Entry, bool) {
	s.mu.RLock()
	e, ok := s.data[key]
	s.mu.RUnlock()
	return e, ok
}
