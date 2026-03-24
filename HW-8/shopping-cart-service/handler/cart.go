// Package handler wires the HTTP transport layer to the store interface.
// All three shopping-cart endpoints from the HW-5 API spec are implemented here.
package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"cart-service/store"
)

type Handler struct {
	store store.Store
}

func New(s store.Store) *Handler {
	return &Handler{store: s}
}

// RegisterRoutes attaches all routes to mux.
// Uses Go 1.22 method+path patterns so no router dependency is needed.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("POST /shopping-carts", h.createCart)
	mux.HandleFunc("GET /shopping-carts/{cartID}", h.getCart)
	mux.HandleFunc("POST /shopping-carts/{cartID}/items", h.addItems)
	mux.HandleFunc("POST /shopping-carts/{cartID}/checkout", h.checkout)
}

// POST /shopping-carts
// Body: {"customer_id": 42}
// Response 201: {"shopping_cart_id": 7}
func (h *Handler) createCart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID int `json:"customer_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CustomerID < 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid input"})
		return
	}
	cartID, err := h.store.CreateCart(r.Context(), req.CustomerID)
	if err != nil {
		log.Printf("createCart error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal error"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int{"shopping_cart_id": cartID})
}

// GET /shopping-carts/{cartID}
// Response 200: {"shopping_cart_id": N, "customer_id": N, "status": "active", "items": [...]}
func (h *Handler) getCart(w http.ResponseWriter, r *http.Request) {
	cartID, ok := pathInt(r, "cartID")
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid cart id"})
		return
	}
	cart, items, err := h.store.GetCart(r.Context(), cartID)
	if err != nil {
		var nf *store.ErrNotFound
		if errors.As(err, &nf) {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"shopping_cart_id": cart.ID,
		"customer_id":      cart.CustomerID,
		"status":           cart.Status,
		"items":            items,
	})
}

// POST /shopping-carts/{cartID}/items
// Body: {"product_id": 5, "quantity": 3}
// Response 204
func (h *Handler) addItems(w http.ResponseWriter, r *http.Request) {
	cartID, ok := pathInt(r, "cartID")
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid cart id"})
		return
	}
	var req struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProductID < 1 || req.Quantity < 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid input"})
		return
	}
	if err := h.store.AddItem(r.Context(), cartID, req.ProductID, req.Quantity); err != nil {
		var nf *store.ErrNotFound
		if errors.As(err, &nf) {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal error"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /shopping-carts/{cartID}/checkout
// Response 200: {"order_id": 12345}
func (h *Handler) checkout(w http.ResponseWriter, r *http.Request) {
	cartID, ok := pathInt(r, "cartID")
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid cart id"})
		return
	}
	orderID, err := h.store.Checkout(r.Context(), cartID)
	if err != nil {
		var nf *store.ErrNotFound
		if errors.As(err, &nf) {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"order_id": orderID})
}

// GET /health — ALB health check
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func pathInt(r *http.Request, key string) (int, bool) {
	v, err := strconv.Atoi(r.PathValue(key))
	return v, err == nil && v > 0
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
