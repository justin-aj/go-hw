// ===================================================================
// FAIL-FAST PATTERN — STANDALONE DEMO
// ===================================================================
//
// CONCEPT: "Don't wait forever for a slow response."
//
// This demo shows ONLY the fail-fast pattern. No bulkhead, no circuit
// breaker. Run with and without RESILIENCE_MODE=resilient to compare:
//
//   Broken:    __slow query → waits 5 FULL seconds
//   Resilient: __slow query → killed at 500ms → 503 immediately
//
// Run:
//   docker compose --profile broken up --build      # no timeout
//   docker compose --profile resilient up --build    # 500ms timeout
//   locust -f locustfile.py --headless -u 50 -r 10 --run-time 60s --host http://localhost:8080
// ===================================================================

package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"demo-failfast/generator"
	"demo-failfast/handler"
	"demo-failfast/resilience"
	"demo-failfast/store"
)

func main() {
	mode := os.Getenv("RESILIENCE_MODE")
	if mode == "" {
		mode = "broken"
	}

	log.Printf("=== FAIL-FAST DEMO ===")
	log.Printf("Mode: %s", mode)

	productStore := store.New()
	generator.Populate(productStore)
	h := handler.New(productStore)

	mux := http.NewServeMux()

	if mode == "resilient" {
		// ONLY fail-fast — wrap search handler with a 500ms timeout
		log.Println("Fail-Fast ENABLED: 500ms timeout on /products/search")
		searchHandler := http.HandlerFunc(h.Search)
		ff := resilience.NewFailFast(searchHandler, 500*time.Millisecond)
		mux.Handle("/products/search", ff)
	} else {
		log.Println("Fail-Fast DISABLED — no timeout protection")
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
