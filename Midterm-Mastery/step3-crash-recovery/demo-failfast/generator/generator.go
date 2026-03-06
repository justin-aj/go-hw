package generator

import (
	"fmt"
	"log"

	"demo-failfast/model"
	"demo-failfast/seeddata"
	"demo-failfast/store"
)

// TotalProducts is smaller than HW-6 (10K vs 100K) for faster startup in demos.
const TotalProducts = 10000

// Populate fills the store with TotalProducts products derived from seeds.
func Populate(s *store.ProductStore) {
	seeds := seeddata.Load()
	numSeeds := len(seeds)

	log.Printf("Generating %d products from %d real product seeds...\n", TotalProducts, numSeeds)

	for i := 1; i <= TotalProducts; i++ {
		seed := seeds[(i-1)%numSeeds]

		variantNum := (i-1)/numSeeds + 1
		name := seed.Name
		if variantNum > 1 {
			name = fmt.Sprintf("%s Edition %d", seed.Name, variantNum)
		}

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
