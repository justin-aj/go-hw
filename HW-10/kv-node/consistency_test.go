package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ---- helpers ----------------------------------------------------------------

// newTestMux registers all four endpoints on a fresh ServeMux for a Handler.
func newTestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /set", h.Set)
	mux.HandleFunc("GET /get/{key}", h.Get)
	mux.HandleFunc("POST /replicate", h.Replicate)
	mux.HandleFunc("GET /local_read/{key}", h.LocalRead)
	return mux
}

// startFollowers spins up n in-process follower nodes. Returns their servers
// and base URLs. All servers are closed automatically when the test ends.
func startFollowers(t *testing.T, n int) ([]*httptest.Server, []string) {
	t.Helper()
	servers := make([]*httptest.Server, n)
	urls := make([]string, n)
	for i := range servers {
		h := NewHandler(NewKVStore(), "follower", 0, 0, nil)
		servers[i] = httptest.NewServer(newTestMux(h))
		urls[i] = servers[i].URL
		t.Cleanup(servers[i].Close)
	}
	return servers, urls
}

// startLeader spins up a leader node wired to the given follower URLs.
func startLeader(t *testing.T, w, r int, followerURLs []string) *httptest.Server {
	t.Helper()
	h := NewHandler(NewKVStore(), "leader", w, r, NewReplicator(followerURLs))
	srv := httptest.NewServer(newTestMux(h))
	t.Cleanup(srv.Close)
	return srv
}

// startLeaderlessCluster spins up n equal leaderless nodes. Because each node
// needs to know the other nodes' URLs (which aren't known until the servers
// start), we start them first without peers and then wire the replicators in.
// This works because the tests are in the same package as the main code.
func startLeaderlessCluster(t *testing.T, n int) []*httptest.Server {
	t.Helper()
	handlers := make([]*Handler, n)
	servers := make([]*httptest.Server, n)

	for i := range handlers {
		handlers[i] = NewHandler(NewKVStore(), "leaderless", 0, 0, nil)
		servers[i] = httptest.NewServer(newTestMux(handlers[i]))
		t.Cleanup(servers[i].Close)
	}
	// All servers are now running — URLs are known. Wire up peers.
	for i, h := range handlers {
		peers := make([]string, 0, n-1)
		for j, s := range servers {
			if j != i {
				peers = append(peers, s.URL)
			}
		}
		h.replicator = NewReplicator(peers)
	}
	return servers
}

// doSet sends POST /set and returns the HTTP status code.
func doSet(baseURL, key, value string) int {
	body, _ := json.Marshal(map[string]string{"key": key, "value": value})
	resp, err := http.Post(baseURL+"/set", "application/json", bytes.NewReader(body))
	if err != nil {
		return 0
	}
	resp.Body.Close()
	return resp.StatusCode
}

// doGet sends GET /get/{key} and returns the decoded response + status code.
func doGet(baseURL, key string) (kvResponse, int) {
	resp, err := http.Get(baseURL + "/get/" + key)
	if err != nil {
		return kvResponse{}, 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return kvResponse{}, resp.StatusCode
	}
	var kv kvResponse
	json.NewDecoder(resp.Body).Decode(&kv)
	return kv, resp.StatusCode
}

// doLocalRead sends GET /local_read/{key} and returns the decoded response + status code.
func doLocalRead(baseURL, key string) (kvResponse, int) {
	resp, err := http.Get(baseURL + "/local_read/" + key)
	if err != nil {
		return kvResponse{}, 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return kvResponse{}, resp.StatusCode
	}
	var kv kvResponse
	json.NewDecoder(resp.Body).Decode(&kv)
	return kv, resp.StatusCode
}

// ---- Leader-Follower tests --------------------------------------------------

// TestW5R1_ConsistencyAfterWrite: after the leader returns 201 (W=5 waited for
// all 4 followers), every node must have the value.
func TestW5R1_ConsistencyAfterWrite(t *testing.T) {
	_, followerURLs := startFollowers(t, 4)
	leader := startLeader(t, 5, 1, followerURLs)

	if status := doSet(leader.URL, "color", "blue"); status != http.StatusCreated {
		t.Fatalf("set: expected 201, got %d", status)
	}

	// Leader must have it
	kv, status := doGet(leader.URL, "color")
	if status != http.StatusOK || kv.Value != "blue" {
		t.Errorf("leader get: expected blue/200, got %q/%d", kv.Value, status)
	}

	// Every follower must have it — W=5 blocked until all confirmed
	for i, url := range followerURLs {
		kv, status := doLocalRead(url, "color")
		if status != http.StatusOK || kv.Value != "blue" {
			t.Errorf("follower%d local_read: expected blue/200, got %q/%d", i+1, kv.Value, status)
		}
	}
}

// TestW1R5_InconsistencyWindow: with W=1 the leader returns 201 immediately
// and replicates in the background. Followers are stale during the ~300ms
// replication window. R=5 reads on the leader still return the correct value
// by fetching from all nodes and picking the highest version.
func TestW1R5_InconsistencyWindow(t *testing.T) {
	_, followerURLs := startFollowers(t, 4)
	leader := startLeader(t, 1, 5, followerURLs)

	// W=1: returns in ~1ms, replication is async
	if status := doSet(leader.URL, "color", "red"); status != http.StatusCreated {
		t.Fatalf("set: expected 201, got %d", status)
	}

	// Followers haven't been reached yet (replication sleeps 200ms before sending)
	staleCount := 0
	for i, url := range followerURLs {
		_, status := doLocalRead(url, "color")
		if status == http.StatusNotFound {
			staleCount++
			t.Logf("follower%d local_read: stale (404) ✓", i+1)
		}
	}
	t.Logf("stale reads immediately after W=1 write: %d/4", staleCount)
	if staleCount == 0 {
		t.Log("WARNING: caught no stale reads — machine may be too fast; run under load")
	}

	// R=5 read from leader fetches all nodes, returns max version — must be correct
	kv, status := doGet(leader.URL, "color")
	if status != http.StatusOK || kv.Value != "red" {
		t.Errorf("leader R=5 get: expected red/200, got %q/%d", kv.Value, status)
	}

	// After replication window, followers must also be consistent
	time.Sleep(400 * time.Millisecond)
	for i, url := range followerURLs {
		kv, status := doLocalRead(url, "color")
		if status != http.StatusOK || kv.Value != "red" {
			t.Errorf("follower%d after replication: expected red/200, got %q/%d", i+1, kv.Value, status)
		}
	}
}

// TestW3R3_QuorumConsistency: after a quorum write (W=3), an R=3 read must
// return the latest value because the write and read sets overlap by ≥1 node.
func TestW3R3_QuorumConsistency(t *testing.T) {
	_, followerURLs := startFollowers(t, 4)
	leader := startLeader(t, 3, 3, followerURLs)

	if status := doSet(leader.URL, "color", "green"); status != http.StatusCreated {
		t.Fatalf("set: expected 201, got %d", status)
	}

	kv, status := doGet(leader.URL, "color")
	if status != http.StatusOK || kv.Value != "green" {
		t.Errorf("quorum R=3 get: expected green/200, got %q/%d", kv.Value, status)
	}
}

// TestW1_LocalReadShowsInconsistency: hammers the leader with W=1 writes while
// concurrently local_reading followers. Expects to catch at least one stale read
// across the burst — demonstrating the inconsistency window under load.
func TestW1_LocalReadShowsInconsistency(t *testing.T) {
	followerServers, followerURLs := startFollowers(t, 4)
	_ = followerServers
	leader := startLeader(t, 1, 5, followerURLs)

	totalStale := 0
	for i := 0; i < 10; i++ {
		key := "key"
		doSet(leader.URL, key, "val")
		for _, url := range followerURLs {
			_, status := doLocalRead(url, key)
			if status == http.StatusNotFound {
				totalStale++
			}
		}
	}
	t.Logf("total stale local_reads across 10 writes: %d", totalStale)
	if totalStale == 0 {
		t.Log("WARNING: no stale reads caught — try running with -count=10 or under load")
	}
}

// ---- Leaderless tests -------------------------------------------------------

// TestLeaderless_InconsistencyWindow: write to node0 (coordinator), then
// immediately GET from other nodes while replication is in flight.
// The coordinator fans out concurrently with a 200ms sleep per peer, so other
// nodes return 404 for ~200ms after the write starts.
func TestLeaderless_InconsistencyWindow(t *testing.T) {
	servers := startLeaderlessCluster(t, 5)

	staleCount := 0
	done := make(chan struct{})

	// Write to node0 in background — takes ~300ms (W=N)
	go func() {
		doSet(servers[0].URL, "item", "newval")
		close(done)
	}()

	// Read from other nodes at ~50ms — replication goroutines are sleeping (200ms)
	time.Sleep(50 * time.Millisecond)
	for i, srv := range servers[1:] {
		_, status := doGet(srv.URL, "item")
		if status == http.StatusNotFound {
			staleCount++
			t.Logf("node%d get during window: stale (404) ✓", i+1)
		}
	}
	t.Logf("stale reads during leaderless write window: %d/4", staleCount)
	if staleCount == 0 {
		t.Log("WARNING: no stale reads caught — try running under load")
	}

	<-done // wait for coordinator to return 201

	// After 201: coordinator must be consistent
	kv, status := doGet(servers[0].URL, "item")
	if status != http.StatusOK || kv.Value != "newval" {
		t.Errorf("coordinator after write: expected newval/200, got %q/%d", kv.Value, status)
	}

	// After 201: all other nodes must also be consistent (W=N waited for all)
	for i, srv := range servers[1:] {
		kv, status := doGet(srv.URL, "item")
		if status != http.StatusOK || kv.Value != "newval" {
			t.Errorf("node%d after write: expected newval/200, got %q/%d", i+1, kv.Value, status)
		}
	}
}

// TestLeaderless_MultipleCoordinators: two sequential writes from different
// coordinators. Last-write-wins (timestamp) ensures the second write is visible
// everywhere after both complete.
func TestLeaderless_MultipleCoordinators(t *testing.T) {
	servers := startLeaderlessCluster(t, 5)

	// First write via node0
	if status := doSet(servers[0].URL, "item", "first"); status != http.StatusCreated {
		t.Fatalf("first set: expected 201, got %d", status)
	}
	// Second write via node3 (different coordinator)
	if status := doSet(servers[3].URL, "item", "second"); status != http.StatusCreated {
		t.Fatalf("second set: expected 201, got %d", status)
	}

	// All nodes must reflect the second write
	for i, srv := range servers {
		kv, status := doGet(srv.URL, "item")
		if status != http.StatusOK || kv.Value != "second" {
			t.Errorf("node%d: expected second/200, got %q/%d", i, kv.Value, status)
		}
	}
}
