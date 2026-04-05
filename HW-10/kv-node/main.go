package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	// 1. Read config from environment variables
	role := envOr("ROLE", "standalone")
	port := envOr("PORT", "8080")
	w := envInt("W", 5)
	r := envInt("R", 1)

	// 2. Create the in-memory store
	store := NewKVStore()

	// 3. Wire up the replicator (only for leader and leaderless)
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

	// 4. Create the handler (store + role + quorums + replicator)
	h := NewHandler(store, role, w, r, replicator)

	// 5. Register routes and start the server
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
