package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// replicateRequest is the body sent from leader to follower on POST /replicate.
type replicateRequest struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Version int64  `json:"version"`
}

// Replicator handles fan-out writes and quorum reads on behalf of the leader.
type Replicator struct {
	followers []string // follower base URLs, e.g. "http://follower1:8080"
	client    *http.Client
}

func NewReplicator(followers []string) *Replicator {
	return &Replicator{
		followers: followers,
		client:    &http.Client{Timeout: 5 * time.Second},
	}
}

// FanOut sends key/value/version to all followers concurrently.
//
// waitsFor: how many follower ACKs to wait for before returning.
//
//	0  = fire-and-forget (goroutines keep running in the background)
//	N  = block until N followers have responded
//
// Each goroutine sleeps 200ms before sending — this simulates the network/disk
// latency described in the spec ("Leader should sleep 200ms following each
// message to a Follower") and widens the inconsistency window for testing.
func (r *Replicator) FanOut(key, value string, version int64, waitsFor int) {
	ch := make(chan error, len(r.followers))
	for _, url := range r.followers {
		go func(followerURL string) {
			time.Sleep(200 * time.Millisecond)
			ch <- r.sendReplicate(followerURL, key, value, version)
		}(url)
	}
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

// ReadBest fetches key from up to count followers concurrently and returns the
// entry with the highest version number.
// Returns (entry, true) if at least one follower had the key, (zero, false) otherwise.
func (r *Replicator) ReadBest(key string, count int) (kvResponse, bool) {
	targets := r.followers
	if count < len(targets) {
		targets = targets[:count]
	}
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
