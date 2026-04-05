package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type replicateRequest struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Version int64  `json:"version"`
}

type Replicator struct {
	followers []string
	client    *http.Client
}

func NewReplicator(followers []string) *Replicator {
	return &Replicator{
		followers: followers,
		client:    &http.Client{Timeout: 5 * time.Second},
	}
}

// FanOut sends to all followers concurrently. waitsFor controls how many ACKs
// to block on before returning (0 = fire-and-forget, N = wait for N).
func (r *Replicator) FanOut(key, value string, version int64, waitsFor int) {
	// 1. Spawn one goroutine per follower
	ch := make(chan error, len(r.followers))
	for _, url := range r.followers {
		go func(followerURL string) {
			// 2. Sleep 200ms (simulates leader→follower network latency per spec)
			time.Sleep(200 * time.Millisecond)
			// 3. Send the replicate request
			ch <- r.sendReplicate(followerURL, key, value, version)
		}(url)
	}
	// 4. Block until waitsFor ACKs received (the rest finish in the background)
	for i := 0; i < waitsFor; i++ {
		if err := <-ch; err != nil {
			log.Printf("replicate error: %v", err)
		}
	}
}

func (r *Replicator) sendReplicate(followerURL, key, value string, version int64) error {
	body, _ := json.Marshal(replicateRequest{Key: key, Value: value, Version: version})
	resp, err := r.client.Post(followerURL+"/replicate", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%s /replicate: %w", followerURL, err)
	}
	resp.Body.Close()
	return nil
}

// ReadBest reads from count followers and returns the entry with the highest version.
func (r *Replicator) ReadBest(key string, count int) (kvResponse, bool) {
	// 1. Pick the first `count` followers to read from
	targets := r.followers
	if count < len(targets) {
		targets = targets[:count]
	}
	// 2. Read from each concurrently
	ch := make(chan kvResponse, len(targets))
	for _, url := range targets {
		go func(followerURL string) {
			entry, err := r.readFrom(followerURL, key)
			if err != nil {
				log.Printf("read from %s: %v", followerURL, err)
				ch <- kvResponse{Version: -1}
				return
			}
			ch <- entry
		}(url)
	}
	// 3. Collect responses and return the highest version
	best := kvResponse{Version: -1}
	for range targets {
		if e := <-ch; e.Version > best.Version {
			best = e
		}
	}
	return best, best.Version >= 0
}

func (r *Replicator) readFrom(followerURL, key string) (kvResponse, error) {
	resp, err := r.client.Get(followerURL + "/get/" + key)
	if err != nil {
		return kvResponse{Version: -1}, fmt.Errorf("get %s: %w", followerURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return kvResponse{Version: -1}, nil
	}
	var entry kvResponse
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		return kvResponse{Version: -1}, err
	}
	return entry, nil
}
