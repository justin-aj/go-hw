// ===================================================================
// CIRCUIT BREAKER PATTERN — STANDALONE DEMO
// ===================================================================
//
// CONCEPT: "Stop calling something that's clearly broken."
//
// This demo shows ONLY the circuit breaker pattern. No fail-fast, no
// bulkhead. Run with and without RESILIENCE_MODE=resilient to compare:
//
//   Broken:    every __slow query waits 5 seconds, over and over
//   Resilient: after 5 failures, circuit trips → __slow queries skipped instantly
//              normal queries (?q=beauty) are NEVER affected
//
// Run:
//   docker compose --profile broken up --build
//   docker compose --profile resilient up --build
//   locust -f locustfile.py --headless -u 50 -r 10 --run-time 60s --host http://localhost:8080
// ===================================================================

package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"demo-circuitbreaker/generator"
	"demo-circuitbreaker/handler"
	"demo-circuitbreaker/resilience"
	"demo-circuitbreaker/store"
)

func main() {
	mode := os.Getenv("RESILIENCE_MODE")
	if mode == "" {
		mode = "broken"
	}

	log.Printf("=== CIRCUIT BREAKER DEMO ===")
	log.Printf("Mode: %s", mode)

	productStore := store.New()
	generator.Populate(productStore)
	h := handler.New(productStore)

	if mode == "resilient" {
		// ONLY circuit breaker — wraps the downstream call inside search
		log.Println("Circuit Breaker ENABLED: trips after 5 failures, 10s cooldown")
		cb := resilience.NewCircuitBreaker(5, 3, 10*time.Second)
		h.SetDownstreamCB(cb)
	} else {
		log.Println("Circuit Breaker DISABLED — every __slow query waits 5 seconds")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/products/search", h.Search)
	mux.HandleFunc("/health", h.Health)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
