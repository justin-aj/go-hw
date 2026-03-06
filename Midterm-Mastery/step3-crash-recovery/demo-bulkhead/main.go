// ===================================================================
// BULKHEAD PATTERN — STANDALONE DEMO
// ===================================================================
//
// CONCEPT: "Limit how many requests can be inside at the same time."
//
// This demo shows ONLY the bulkhead pattern. No fail-fast, no circuit
// breaker. Run with and without RESILIENCE_MODE=resilient to compare:
//
//   Broken:    unlimited concurrency → server overloaded → crash
//   Resilient: max 10 concurrent → excess gets 429 → server stays alive
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

	"demo-bulkhead/generator"
	"demo-bulkhead/handler"
	"demo-bulkhead/resilience"
	"demo-bulkhead/store"
)

func main() {
	mode := os.Getenv("RESILIENCE_MODE")
	if mode == "" {
		mode = "broken"
	}

	log.Printf("=== BULKHEAD DEMO ===")
	log.Printf("Mode: %s", mode)

	productStore := store.New()
	generator.Populate(productStore)
	h := handler.New(productStore)

	mux := http.NewServeMux()

	if mode == "resilient" {
		// ONLY bulkhead — limit to 10 concurrent search requests
		log.Println("Bulkhead ENABLED: max 10 concurrent search requests")
		searchHandler := http.HandlerFunc(h.Search)
		bh := resilience.NewBulkhead(searchHandler, 10)
		mux.Handle("/products/search", bh)
	} else {
		log.Println("Bulkhead DISABLED — unlimited concurrency")
		mux.HandleFunc("/products/search", h.Search)
	}

	mux.HandleFunc("/health", h.Health)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
