package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	// We'll use an atomic integer type to represent our (always-positive) counter.
	var atomicOps atomic.Uint64

	// Non-atomic counter for comparison - THIS WILL HAVE RACE CONDITIONS!
	var nonAtomicOps uint64

	// A WaitGroup will help us wait for all goroutines to finish their work.
	var wg sync.WaitGroup

	// We'll start 500 goroutines that each increment the counters exactly 1000 times.
	for range 500 {
		wg.Go(func() {
			for range 1000 {
				// Atomic increment - thread-safe
				atomicOps.Add(1)

				// Non-atomic increment - NOT thread-safe!
				// This is a read-modify-write operation that can be interleaved
				nonAtomicOps++
			}
		})
	}

	// Wait until all the goroutines are done.
	wg.Wait()

	// Print results
	fmt.Println("Expected value: 500000 (500 goroutines × 1000 increments)")
	fmt.Println("Atomic ops:    ", atomicOps.Load())
	fmt.Println("Non-atomic ops:", nonAtomicOps)

	// Calculate how many increments were lost due to race conditions
	expected := uint64(500000)
	atomicLost := expected - atomicOps.Load()
	nonAtomicLost := expected - nonAtomicOps

	fmt.Printf("\nIncrements lost:\n")
	fmt.Printf("  Atomic:     %d (%.2f%%)\n", atomicLost, float64(atomicLost)/float64(expected)*100)
	fmt.Printf("  Non-atomic: %d (%.2f%%)\n", nonAtomicLost, float64(nonAtomicLost)/float64(expected)*100)
}
