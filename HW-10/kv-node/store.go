package main

import (
	"sync"
	"sync/atomic"
)

// Entry holds a value and a logical version number.
// Version is a monotonically increasing counter shared across the node,
// so every write gets a unique, comparable version.
type Entry struct {
	Value   string
	Version int64
}

// KVStore is a thread-safe in-memory key-value store.
type KVStore struct {
	mu      sync.RWMutex
	data    map[string]Entry
	counter int64 // global version counter, incremented atomically on every write
}

func NewKVStore() *KVStore {
	return &KVStore{data: make(map[string]Entry)}
}

// Set stores value under key and returns the new version number.
func (s *KVStore) Set(key, value string) int64 {
	version := atomic.AddInt64(&s.counter, 1)
	s.mu.Lock()
	s.data[key] = Entry{Value: value, Version: version}
	s.mu.Unlock()
	return version
}

// Get retrieves the entry for key. Returns (entry, true) or (zero, false).
func (s *KVStore) Get(key string) (Entry, bool) {
	s.mu.RLock()
	e, ok := s.data[key]
	s.mu.RUnlock()
	return e, ok
}
