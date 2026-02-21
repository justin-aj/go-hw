// Package seeddata fetches the base product catalog from DummyJSON API.
//
// Design decision hidden: Where seed data comes from and how it's parsed.
// Currently pulls all products from https://dummyjson.com/products at startup.
// Could be swapped to read from a file, database, or other API without
// affecting any other module.
package seeddata

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// SeedProduct holds the template data for generating product variants.
type SeedProduct struct {
	Name        string
	Category    string
	Description string
	Brand       string
	Price       float64
}

// dummyJSONResponse matches the DummyJSON API response shape.
type dummyJSONResponse struct {
	Products []dummyJSONProduct `json:"products"`
	Total    int                `json:"total"`
}

// dummyJSONProduct matches a single product from DummyJSON.
type dummyJSONProduct struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Brand       string  `json:"brand"`
}

// apiURL fetches all products with only the fields we need.
const apiURL = "https://dummyjson.com/products?limit=0&select=title,description,category,price,brand"

// Load fetches all products from DummyJSON and returns them as seeds.
func Load() []SeedProduct {
	log.Println("Fetching seed products from DummyJSON API...")

	resp, err := http.Get(apiURL)
	if err != nil {
		log.Fatalf("Failed to fetch from DummyJSON: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("DummyJSON returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var apiResp dummyJSONResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		log.Fatalf("Failed to parse DummyJSON response: %v", err)
	}

	seeds := make([]SeedProduct, 0, len(apiResp.Products))
	for _, p := range apiResp.Products {
		brand := p.Brand
		if brand == "" {
			brand = "Generic"
		}
		seeds = append(seeds, SeedProduct{
			Name:        p.Title,
			Category:    p.Category,
			Description: p.Description,
			Brand:       brand,
			Price:       p.Price,
		})
	}

	log.Printf("Fetched %d seed products from DummyJSON (total available: %d)\n", len(seeds), apiResp.Total)
	if len(seeds) == 0 {
		log.Fatal("No seed products fetched — cannot start service")
	}

	// Log category distribution — use log.Printf for consistency
	categories := make(map[string]int)
	for _, s := range seeds {
		categories[s.Category]++
	}
	log.Printf("Categories: %v\n", categories)

	return seeds
}
