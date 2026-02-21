// Package search implements bounded product search over the store.
//
// Design decision hidden: The search algorithm and iteration bounds.
// Currently does case-insensitive substring matching over exactly
// MaxCheck products, returning at most MaxResults. The matching
// strategy (substring, regex, fuzzy, inverted index) and the
// iteration bound are encapsulated here. Changing the algorithm
// from O(n) scan to an inverted index would only require changes
// in this module.
package search

import (
	"strings"
	"time"

	"product-search/model"
	"product-search/store"
)

const (
	// MaxCheck is the number of products inspected per search.
	// This simulates a fixed-cost computation (e.g., running an
	// AI model on each product). The assignment requires exactly 100.
	MaxCheck = 100

	// MaxResults caps the number of products returned in a response.
	MaxResults = 20
)

// Result holds the outcome of a single search operation.
type Result struct {
	Products   []model.Product `json:"products"`
	TotalFound int             `json:"total_found"`
	SearchTime string          `json:"search_time"`
	Checked    int             `json:"products_checked"`
}

// Execute runs a bounded search for the given query string.
// It checks exactly MaxCheck products and returns up to MaxResults matches.
func Execute(s *store.ProductStore, query string) Result {
	start := time.Now()
	queryLower := strings.ToLower(query)

	var matches []model.Product
	totalFound := 0

	// Iterate over exactly MaxCheck products via the store's iterator.
	// The callback receives every product; we count ALL visited, not
	// just matches (this is the "fixed computation" the assignment requires).
	checked := s.Iterate(1, MaxCheck, func(p model.Product) bool {
		nameLower := strings.ToLower(p.Name)
		catLower := strings.ToLower(p.Category)

		if strings.Contains(nameLower, queryLower) || strings.Contains(catLower, queryLower) {
			totalFound++
			if len(matches) < MaxResults {
				matches = append(matches, p)
			}
		}
		return true // always continue until MaxCheck is reached
	})

	return Result{
		Products:   matches,
		TotalFound: totalFound,
		SearchTime: time.Since(start).String(),
		Checked:    checked,
	}
}
