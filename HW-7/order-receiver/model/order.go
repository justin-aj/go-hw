package model

import "time"

// Item is a single product in an order.
type Item struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// Order is the core domain object.
// Status transitions: pending → processing → completed
type Order struct {
	OrderID    string    `json:"order_id"`
	CustomerID int       `json:"customer_id"`
	Status     string    `json:"status"` // pending | processing | completed
	Items      []Item    `json:"items"`
	CreatedAt  time.Time `json:"created_at"`
}
