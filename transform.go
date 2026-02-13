package toil

import (
	"runtime"
	"sync"
)

// TransformFunc defines the type of function that can be applied to each item in the input slice.
// It takes an input of type I and returns an output of type O along with an error if any occurs.
// If the function returns an error, this error is returned if AbortOnError is true.
type TransformFunc[I any, O any] func(I) (O, error)

// ParallelTransform takes a slice of input items of `I`, a function `f` and transforms each item
// This is very similar to the Python `multiprocessing.Pool.Map` -- just for Go.
// Order is preserved during the transformation.
func ParallelTransform[I any, O any](v *[]I, f TransformFunc[I, O], opts Options) ([]O, error) {
	if opts.workers <= 0 {
		opts.workers = runtime.NumCPU()
	}

	results := make([]O, len(*v))
	sem := make(chan struct{}, opts.workers) // Semaphore for worker limiting
	errChan := make(chan error, len(*v))     // Buffer for all potential errors

	var wg sync.WaitGroup

	for n, item := range *v {
		wg.Add(1)
		go func(i I, index int) {
			defer wg.Done()

			sem <- struct{}{}        // Acquire worker slot
			defer func() { <-sem }() // Release worker slot

			result, err := f(i)
			if err != nil {
				if opts.stopOnError {
					errChan <- err
					return
				}
				// Still record the error but continue
				select {
				case errChan <- err:
				default:
				}
			}
			results[index] = result
		}(item, n)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	select {
	case err := <-errChan:
		if opts.stopOnError {
			return nil, err
		}
		// If not aborting on error, you might want to collect all errors
		// or just return the first one found
		return results, err
	default:
	}

	return results, nil
}
