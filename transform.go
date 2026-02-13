package toil

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// TransformFunc defines the type of function that can be applied to each item in the input slice.
// It takes an input of type I and returns an output of type O along with an error if any occurs.
// If the function returns an error, this error is returned if AbortOnError is true.
type TransformFunc[I any, O any] func(I) (O, error)

// transformJob represents a single transform operation
type transformJob[I, O any] struct {
	item  I
	index int
}

// ParallelTransform takes a slice of input items of `I`, a function `f` and transforms each item
// This is very similar to the Python `multiprocessing.Pool.Map` -- just for Go.
// Order is preserved during the transformation.
func ParallelTransform[I any, O any](v []I, f TransformFunc[I, O], opts Options) ([]O, error) {
	if opts.workers <= 0 {
		opts.workers = runtime.NumCPU()
	}

	results := make([]O, len(v))
	
	// Early return for empty input
	if len(v) == 0 {
		return results, nil
	}

	var (
		wg       sync.WaitGroup
		firstErr atomic.Pointer[error]  // Lock-free error storage
	)

	// Create job channel with buffer to avoid blocking
	jobs := make(chan transformJob[I, O], len(v))
	
	// Start worker pool
	for i := 0; i < opts.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				result, err := f(job.item)
				if err != nil {
					// Lock-free error handling - first error wins
					if opts.stopOnError {
						firstErr.CompareAndSwap(nil, &err)
						// Drain remaining jobs on error if stopping
						go func() {
							for range jobs {
								// Consume remaining jobs to prevent deadlock
							}
						}()
						return
					}
					// Record error but continue processing
					firstErr.CompareAndSwap(nil, &err)
				}
				// Direct indexed write - no mutex needed
				results[job.index] = result
			}
		}()
	}

	// Send all jobs to workers
	for i, item := range v {
		jobs <- transformJob[I, O]{item: item, index: i}
	}
	close(jobs)

	wg.Wait()

	// Check for errors
	if errPtr := firstErr.Load(); errPtr != nil {
		if opts.stopOnError {
			return nil, *errPtr
		}
		return results, *errPtr
	}

	return results, nil
}
