// Order Receiver Service
//
// Exposes two HTTP endpoints:
//   POST /orders/sync   — synchronous (Phase 1)
//   POST /orders/async  — async via SNS (Phase 3)
//   GET  /health        — ALB health probe
package main

import (
	"log"
	"net/http"
	"os"

	"order-receiver/handler"
)

func main() {
	h := handler.New()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Order Receiver listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
