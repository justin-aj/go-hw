// Package generator populates the product store from seed data.
//
// Design decision hidden: The strategy for expanding seed products
// into a full catalog. Currently uses variant numbering with slight
// price adjustments. Could be changed to use random generation,
// Markov chains, or any other expansion strategy without affecting
// the store, search, or handler modules.
package generator

import (
	"fmt"
	"log"

	"product-search/model"
	"product-search/seeddata"
	"product-search/store"
)

const TotalProducts = 100000

// Populate fills the store with TotalProducts products derived from seeds.
func Populate(s *store.ProductStore) {
	seeds := seeddata.Load()
	numSeeds := len(seeds)

	log.Printf("Generating %d products from %d real product seeds...\n", TotalProducts, numSeeds)

	for i := 1; i <= TotalProducts; i++ {
		seed := seeds[(i-1)%numSeeds]

		// Variant numbering: first cycle uses original name, subsequent
		// cycles append "Edition N" to differentiate.
		variantNum := (i-1)/numSeeds + 1
		name := seed.Name
		if variantNum > 1 {
			name = fmt.Sprintf("%s Edition %d", seed.Name, variantNum)
		}

		// Small price drift per variant to simulate real-world variation.
		price := seed.Price * (1.0 + float64(variantNum-1)*0.01)

		s.Put(model.Product{
			ID:          i,
			Name:        name,
			Category:    seed.Category,
			Description: seed.Description,
			Brand:       seed.Brand,
			Price:       price,
		})
	}

	log.Printf("Product catalog ready: %d products loaded\n", s.Count())
}
