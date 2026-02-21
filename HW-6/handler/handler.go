// Package handler provides HTTP request handlers for the product API.
//
// Design decision hidden: The transport protocol, URL structure,
// response format, and error handling strategy. Currently serves
// JSON over HTTP. Could be replaced with gRPC, GraphQL, or any
// other transport without modifying the search or store modules.
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"product-search/search"
	"product-search/store"
)

// ProductHandler holds dependencies for HTTP handlers.
type ProductHandler struct {
	store *store.ProductStore
}

// New creates a ProductHandler with the given store.
func New(s *store.ProductStore) *ProductHandler {
	return &ProductHandler{store: s}
}

// RegisterRoutes wires up all HTTP endpoints.
func (h *ProductHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/products/search", h.Search)
	mux.HandleFunc("/health", h.Health)
}

// Search handles GET /products/search?q={query}
func (h *ProductHandler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "only GET is supported",
		})
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "query parameter 'q' is required",
		})
		return
	}

	result := search.Execute(h.store, query)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Health handles GET /health for ALB health checks.
func (h *ProductHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "healthy",
		"products": strconv.Itoa(h.store.Count()),
	})
}
