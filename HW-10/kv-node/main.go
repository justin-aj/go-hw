// kv-node — in-memory Key-Value store node
//
// Single binary that runs as a leader, follower, or standalone node.
// Role and quorum settings are controlled entirely via environment variables,
// so the same Docker image serves all roles in the cluster.
//
// Environment variables:
//
//	ROLE       "leader" | "follower" | "standalone" (default: standalone)
//	PORT       HTTP listen port (default: 8080)
//	W          Write quorum — how many nodes must confirm a write (default: 5)
//	            Leader only. Leader itself always counts as 1.
//	            W=5 → wait for all 4 followers before responding.
//	            W=3 → wait for 2 followers (quorum).
//	            W=1 → return immediately, replicate async.
//	R          Read quorum — how many nodes to read from (default: 1)
//	            Leader only.
//	            R=1 → read from leader only.
//	            R=3 → read from leader + 2 followers, return max version.
//	            R=5 → read from all 5 nodes, return max version.
//	FOLLOWERS  Comma-separated follower base URLs (required when ROLE=leader)
//	            e.g. "http://follower1:8080,http://follower2:8080,..."
//
// Endpoints:
//
//	POST /set              {"key":"...","value":"..."} → 201
//	GET  /get/{key}        → 200 {"key","value","version"} | 404
//	POST /replicate        {"key","value","version"} → 201  (internal, leader→follower)
//	GET  /local_read/{key} → 200 {"key","value","version"} | 404  (test hook, no delays)
package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	role := envOr("ROLE", "standalone")
	port := envOr("PORT", "8080")
	w := envInt("W", 5)
	r := envInt("R", 1)

	store := NewKVStore()

	var replicator *Replicator
	switch role {
	case "leader":
		followersStr := os.Getenv("FOLLOWERS")
		if followersStr == "" {
			log.Fatal("ROLE=leader requires FOLLOWERS env var")
		}
		followers := strings.Split(followersStr, ",")
		replicator = NewReplicator(followers)
		log.Printf("leader mode: W=%d R=%d followers=%v", w, r, followers)
	case "leaderless":
		peersStr := os.Getenv("PEERS")
		if peersStr == "" {
			log.Fatal("ROLE=leaderless requires PEERS env var")
		}
		peers := strings.Split(peersStr, ",")
		replicator = NewReplicator(peers)
		log.Printf("leaderless mode: peers=%v", peers)
	}

	h := NewHandler(store, role, w, r, replicator)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /set", h.Set)
	mux.HandleFunc("GET /get/{key}", h.Get)
	mux.HandleFunc("POST /replicate", h.Replicate)
	mux.HandleFunc("GET /local_read/{key}", h.LocalRead)

	log.Printf("kv-node listening on :%s  role=%s", port, role)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("invalid %s=%q, using default %d", key, v, def)
		return def
	}
	return n
}
