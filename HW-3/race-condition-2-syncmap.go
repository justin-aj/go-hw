package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// Create a sync.Map for concurrent access without explicit locking
	var m sync.Map

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start timing
	startTime := time.Now()

	// Spawn 50 goroutines
	for g := 0; g < 50; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// Run 1000 iterations in each goroutine
			for i := 0; i < 1000; i++ {
				m.Store(goroutineID*1000+i, i)
			}
		}(g)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Calculate elapsed time
	elapsed := time.Since(startTime)

	// Count entries using Range
	// Range calls the function for each key-value pair in the map
	// The function returns true to continue iteration, false to stop
	count := 0
	m.Range(func(key, value interface{}) bool {
		count++
		return true // continue iteration
	})

	// Print the count of map entries and time taken
	fmt.Printf("Length of map: %d\n", count)
	fmt.Printf("Total time taken: %v\n", elapsed)
	fmt.Printf("\nExample: First few entries:\n")

	// Demonstrate Range with early termination
	entryCount := 0
	m.Range(func(key, value interface{}) bool {
		if entryCount < 5 {
			fmt.Printf("  Key: %v, Value: %v\n", key, value)
			entryCount++
			return true // continue
		}
		return false // stop after 5 entries
	})
}
