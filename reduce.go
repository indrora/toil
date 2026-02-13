package toil

import (
	"runtime"
	"sync"
)

// a ReduceFunc is a function that takes two values of T and returns the "sum" of those values.
// For example, if T is int, a ReduceFunc could be a function that adds two integers together.
// If this function returns an error, the reduction will stop, the error is returned.
type ReduceFunc[T any] func(T, T) (T, error)

// ParallelReduce applies a binary function to reduce a slice to a single value in parallel.
// The function f should be associative for correct results. The reduction is performed in parallel
// using the number of workers specified in opts. If the slice is empty, returns an error.

func ParallelReduce[T any](v *[]T, f ReduceFunc[T], opts Options) (T, error) {
	var zero T
	n := len(*v)
	if n == 0 {
		return zero, nil // or return error if you want to disallow empty input
	}
	if n == 1 {
		return (*v)[0], nil
	}
	if opts.workers <= 0 {
		opts.workers = runtime.NumCPU()
	}

	items := *v
	for len(items) > 1 {
		var (
			wg   sync.WaitGroup
			mu   sync.Mutex
			next []T
			err  error
		)
		sem := make(chan struct{}, opts.workers)

		for i := 0; i < len(items)-1; i += 2 {
			wg.Add(1)
			sem <- struct{}{}
			go func(a, b T) {
				defer wg.Done()
				res, e := f(a, b)
				mu.Lock()
				if e != nil && err == nil {
					err = e
				}
				next = append(next, res)
				mu.Unlock()
				<-sem
			}(items[i], items[i+1])
		}
		// If odd, carry last item forward
		if len(items)%2 == 1 {
			mu.Lock()
			next = append(next, items[len(items)-1])
			mu.Unlock()
		}
		wg.Wait()
		if err != nil {
			return zero, err
		}
		items = next
	}
	return items[0], nil
}
