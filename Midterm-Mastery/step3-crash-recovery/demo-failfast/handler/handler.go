package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"demo-failfast/model"
	"demo-failfast/store"
)

type Handler struct{ store *store.ProductStore }

func New(s *store.ProductStore) *Handler { return &Handler{store: s} }

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing q"})
		return
	}

	start := time.Now()

	// FAULT: __slow simulates a 5-second downstream call
	if strings.ToLower(query) == "__slow" {
		time.Sleep(5 * time.Second)
		query = "beauty"
	}

	queryLower := strings.ToLower(query)
	var results []model.Product
	h.store.Iterate(1, 100, func(p model.Product) bool {
		if strings.Contains(strings.ToLower(p.Name), queryLower) ||
			strings.Contains(strings.ToLower(p.Category), queryLower) {
			results = append(results, p)
		}
		return len(results) < 20
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"products":    results,
		"total_found": len(results),
		"search_time": time.Since(start).String(),
	})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
