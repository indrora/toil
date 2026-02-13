package toil

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// a ReduceFunc is a function that takes two values of T and returns the "sum" of those values.
// For example, if T is int, a ReduceFunc could be a function that adds two integers together.
// If this function returns an error, the reduction will stop, the error is returned.
type ReduceFunc[T any] func(T, T) (T, error)

// ParallelReduce applies a binary function to reduce a slice to a single value in parallel.
// The function f should be associative for correct results. The reduction is performed in parallel
// using the number of workers specified in opts. If the slice is empty, returns an error.

func ParallelReduce[T any](v []T, f ReduceFunc[T], opts Options) (T, error) {
	var zero T
	n := len(v)
	if n == 0 {
		return zero, nil // or return error if you want to disallow empty input
	}
	if n == 1 {
		return v[0], nil
	}
	if opts.workers <= 0 {
		opts.workers = runtime.NumCPU()
	}

	items := v
	for len(items) > 1 {
		// Pre-allocate next slice with exact capacity to eliminate reallocations
		nextCap := (len(items) + 1) / 2  // Ceiling division for pair count
		next := make([]T, nextCap)       // Pre-allocated with exact size (not just capacity)
		
		var (
			wg       sync.WaitGroup
			firstErr atomic.Pointer[error]  // Lock-free error storage
		)
		sem := make(chan struct{}, opts.workers)

		for i := 0; i < len(items)-1; i += 2 {
			wg.Add(1)
			sem <- struct{}{}
			
			go func(a, b T, resultIndex int) {
				defer wg.Done()
				defer func() { <-sem }()
				
				res, err := f(a, b)
				if err != nil {
					// Lock-free: only first error wins, others ignored
					firstErr.CompareAndSwap(nil, &err)
				}
				// Lock-free: direct indexed write, no contention
				next[resultIndex] = res
				
			}(items[i], items[i+1], i/2)
		}
		
		// Handle odd element outside goroutines (no mutex needed)
		if len(items)%2 == 1 {
			next[nextCap-1] = items[len(items)-1]
		}
		
		wg.Wait()
		
		// Check for any errors after all work complete
		if errPtr := firstErr.Load(); errPtr != nil {
			return zero, *errPtr
		}
		
		items = next
	}
	return items[0], nil
}
