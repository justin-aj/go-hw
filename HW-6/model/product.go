// Package model defines the core data types for the product catalog.
//
// Design decision hidden: The product data representation and JSON
// serialization tags. Changing fields, types, or JSON keys only
// requires changes here, not in the store, search, or handler layers.
package model

// Product represents a single item in the product catalog.
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Brand       string  `json:"brand"`
	Price       float64 `json:"price"`
}
