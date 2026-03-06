package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"demo-circuitbreaker/model"
	"demo-circuitbreaker/store"
)

// DownstreamCaller is implemented by the circuit breaker.
type DownstreamCaller interface {
	Call(fn func() error) error
}

type Handler struct {
	store        *store.ProductStore
	downstreamCB DownstreamCaller // nil in broken mode, set in resilient mode
}

func New(s *store.ProductStore) *Handler { return &Handler{store: s} }

func (h *Handler) SetDownstreamCB(cb DownstreamCaller) { h.downstreamCB = cb }

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing q"})
		return
	}

	start := time.Now()
	circuitBlocked := false

	// FAULT: __slow simulates a 5-second downstream call
	if strings.ToLower(query) == "__slow" {
		if h.downstreamCB != nil {
			// Resilient mode: ask circuit breaker
			//   CLOSED  → runs the sleep
			//   OPEN    → skips it instantly
			err := h.downstreamCB.Call(func() error {
				time.Sleep(5 * time.Second)
				return fmt.Errorf("downstream timeout")
			})
			if err != nil {
				circuitBlocked = true
			}
		} else {
			// Broken mode: no protection, sleep full 5 seconds
			time.Sleep(5 * time.Second)
		}
		if !circuitBlocked {
			query = "beauty"
		}
	}

	if circuitBlocked {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(503)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":           "downstream unavailable (circuit open)",
			"search_time":     time.Since(start).String(),
			"circuit_blocked": true,
		})
		return
	}

	// Normal search
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
