package main

import (
	"sync"
	"sync/atomic"
	"time"
)

type Entry struct {
	Value   string
	Version int64
}

type KVStore struct {
	mu      sync.RWMutex
	data    map[string]Entry
	counter int64
}

func NewKVStore() *KVStore {
	return &KVStore{data: make(map[string]Entry)}
}

// Set uses an atomic counter — for the leader, where one node assigns all versions.
func (s *KVStore) Set(key, value string) int64 {
	// 1. Increment the global counter to get a unique version
	version := atomic.AddInt64(&s.counter, 1)
	// 2. Write to the map
	s.mu.Lock()
	s.data[key] = Entry{Value: value, Version: version}
	s.mu.Unlock()
	return version
}

// SetTS uses a nanosecond timestamp — for leaderless, where any node can write.
// Last-write-wins: only updates if the incoming version is newer.
func (s *KVStore) SetTS(key, value string) int64 {
	// 1. Use current time as the version (unique across independent coordinators)
	version := time.Now().UnixNano()
	// 2. Only write if this version is newer than what's stored
	s.mu.Lock()
	if existing, ok := s.data[key]; !ok || version > existing.Version {
		s.data[key] = Entry{Value: value, Version: version}
	}
	s.mu.Unlock()
	return version
}

// SetWithVersion only updates if the incoming version is newer (handles out-of-order replication).
func (s *KVStore) SetWithVersion(key, value string, version int64) {
	// 1. Only write if this version beats what's already stored
	s.mu.Lock()
	if existing, ok := s.data[key]; !ok || version > existing.Version {
		s.data[key] = Entry{Value: value, Version: version}
	}
	s.mu.Unlock()
}

func (s *KVStore) Get(key string) (Entry, bool) {
	s.mu.RLock()
	e, ok := s.data[key]
	s.mu.RUnlock()
	return e, ok
}
