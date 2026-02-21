// Product Search Service
//
// Modularized following Parnas (1972) "On the Criteria To Be Used
// in Decomposing Systems into Modules." Each package hides one
// design decision behind a stable interface:
//
//	model      → product data representation
//	store      → storage mechanism (sync.Map, concurrency)
//	seeddata   → seed catalog source and content
//	generator  → expansion strategy (seeds → 100K products)
//	search     → algorithm, iteration bounds, matching logic
//	handler    → HTTP transport, routing, serialization
//
// main is the composition root: it wires modules together but
// contains no domain logic itself.
package main

import (
	"log"
	"net/http"
	"os"

	"product-search/generator"
	"product-search/handler"
	"product-search/store"
)

func main() {
	// 1. Create the storage layer.
	productStore := store.New()

	// 2. Populate with generated products.
	generator.Populate(productStore)

	// 3. Wire HTTP handlers to the store.
	h := handler.New(productStore)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// 4. Start serving. Port is read from the PORT env var so that
	//    Docker / cloud orchestrators can inject it at runtime.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Product Search Service listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
