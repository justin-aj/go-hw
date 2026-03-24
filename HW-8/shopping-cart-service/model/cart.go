package model

// Cart represents a shopping cart in either backend.
type Cart struct {
	ID         int    `json:"shopping_cart_id"`
	CustomerID int    `json:"customer_id"`
	Status     string `json:"status"`
}

// CartItem is a product + quantity inside a cart.
type CartItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}
