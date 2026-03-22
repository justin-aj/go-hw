// Package store defines the storage interface for the shopping cart service.
// Both MySQL and DynamoDB backends implement this interface, allowing the
// handler layer to remain completely backend-agnostic.
package store

import (
	"context"
	"fmt"

	"cart-service/model"
)

// Store is the single abstraction both backends satisfy.
type Store interface {
	// CreateCart inserts a new cart for the given customer and returns its ID.
	CreateCart(ctx context.Context, customerID int) (cartID int, err error)

	// GetCart retrieves a cart and all its items by cart ID.
	// Returns ErrNotFound if the cart does not exist.
	GetCart(ctx context.Context, cartID int) (*model.Cart, []model.CartItem, error)

	// AddItem appends a product+quantity to an existing cart.
	// Returns an error that wraps ErrNotFound if the cart does not exist.
	AddItem(ctx context.Context, cartID, productID, quantity int) error

	// Checkout marks the cart as checked-out and returns a generated order ID.
	// Returns an error that wraps ErrNotFound if the cart is missing or already
	// checked out.
	Checkout(ctx context.Context, cartID int) (orderID int, err error)

	// Close releases any resources held by the backend (DB pool, etc.).
	Close()
}

// ErrNotFound is returned (wrapped) when a cart does not exist.
type ErrNotFound struct{ CartID int }

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("cart %d not found", e.CartID)
}
