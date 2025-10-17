package main

import (
	"fmt"
	"sync"

	"github.com/almoatamed/go-conc/concpool"
)

// This example program demonstrates using the concpool.Pool package.
func main() {
	pool := concpool.New(10)
	var counter uint64
	var mu sync.Mutex

	const tasks = 10000

	for i := 0; i < tasks; i++ {
		pool.Run(func() error {
			mu.Lock()
			counter += 1
			mu.Unlock()
			return nil
		})
	}

	results := pool.Wait()

	fmt.Printf("Finished: counter=%d results=%d\n", counter, len(results))

	// Print a summary of failures (if any)
	failed := 0
	for _, r := range results {
		if !r.Success {
			failed++
		}
	}
	fmt.Printf("failed=%d\n", failed)
}
