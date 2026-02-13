# toil

A Go package for parallel processing of collections with error handling and worker management.

## Features

- **Parallel Transform**: Apply a function to each item in a slice concurrently
- **Parallel Reduce**: Reduce a slice to a single value using parallel binary operations
- **Worker Control**: Configure the number of concurrent workers
- **Error Handling**: Choose between stopping on first error or collecting all errors

## Usage

### Parallel Transform

Transform each element of a slice in parallel:

```go
package main

import (
	"fmt"
	"github.com/indrora/toil"
)

func main() {
	input := []int{1, 2, 3, 4, 5}
	
	// Double each number
	double := func(x int) (int, error) {
		return x * 2, nil
	}
	
	opts := toil.Options{}.WithWorkers(2)
	results, err := toil.ParallelTransform(input, double, opts)
	if err != nil {
		panic(err)
	}
	
	fmt.Println(results) // [2, 4, 6, 8, 10]
}
```

### Parallel Reduce

Reduce a slice to a single value in parallel:

```go
package main

import (
	"fmt"
	"github.com/indrora/toil"
)

func main() {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	
	// Sum all numbers
	sum := func(a, b int) (int, error) {
		return a + b, nil
	}
	
	opts := toil.Options{}.WithWorkers(3)
	result, err := toil.ParallelReduce(input, sum, opts)
	if err != nil {
		panic(err)
	}
	
	fmt.Println(result) // 55
}
```

### Options

Configure processing behavior:

```go
opts := toil.Options{}.
	WithWorkers(4).          // Use 4 workers (default: number of CPU cores)
	StopOnError(true)        // Stop on first error (default: false)
```

## Notes

- Order is preserved for `ParallelTransform` results
- The reduction function in `ParallelReduce` should be associative (order not guaranteed)
- Order of processing in `ParallelReduce` is *not* guaranteed or preserved.
- Be wary of side effects: if `StopOnError` is true, no further work will be scheduled *upon reporting of an error*; any functions which have not yet completed will still complete.
- If workers is 0 or negative, defaults to `runtime.NumCPU()`
- Reduce is a memory-heavy operation
