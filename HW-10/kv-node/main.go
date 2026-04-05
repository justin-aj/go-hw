// kv-node — in-memory Key-Value store node
//
// This is the base building block for the Leader-Follower and Leaderless
// distributed databases built in HW-10.  Each node is a single instance of
// this binary; replication logic is layered on top in Phase 2.
//
// Environment variables:
//
//	ROLE   "leader" | "follower" | "standalone" (default: standalone)
//	       Controls which artificial sleep delays are applied to simulate
//	       real-world storage latency and make inconsistency windows visible.
//	PORT   HTTP listen port (default: 8080)
//
// Endpoints:
//
//	POST /set            {"key":"...","value":"..."} → 201 Created
//	GET  /get/{key}      → 200 {"key","value","version"} | 404
//	GET  /local_read/{key} → same as /get but no delays, no forwarding
package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	role := envOr("ROLE", "standalone")
	port := envOr("PORT", "8080")

	store := NewKVStore()
	h := NewHandler(store, role)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /set", h.Set)
	mux.HandleFunc("GET /get/{key}", h.Get)
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
