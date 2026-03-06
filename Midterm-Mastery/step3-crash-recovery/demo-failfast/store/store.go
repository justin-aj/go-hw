package store

import (
	"sync"
	"sync/atomic"

	"demo-failfast/model"
)

// ProductStore manages the product catalog in memory.
type ProductStore struct {
	data  sync.Map
	count atomic.Int64
}

// New creates an empty ProductStore.
func New() *ProductStore {
	return &ProductStore{}
}

// Put adds or updates a product in the store.
func (s *ProductStore) Put(product model.Product) {
	s.data.Store(product.ID, product)
	s.count.Add(1)
}

// Get retrieves a product by ID.
func (s *ProductStore) Get(id int) (model.Product, bool) {
	val, ok := s.data.Load(id)
	if !ok {
		return model.Product{}, false
	}
	return val.(model.Product), true
}

// Count returns the total number of products in the store.
func (s *ProductStore) Count() int {
	return int(s.count.Load())
}

// Iterate calls fn for products starting at startID, up to maxCount products.
func (s *ProductStore) Iterate(startID, maxCount int, fn func(model.Product) bool) int {
	total := int(s.count.Load())
	visited := 0
	for id := startID; id <= total && visited < maxCount; id++ {
		val, ok := s.data.Load(id)
		if !ok {
			continue
		}
		visited++
		if !fn(val.(model.Product)) {
			break
		}
	}
	return visited
}
