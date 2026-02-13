// Package toil provides utilities for parallel processing of collections with error
// handling and worker management. Functionally, it provides a simple way to process
// a slice of input items in parallel, applying a user-defined function to each item,
// while controlling the number of concurrent workers and how errors are handled.
//
// CAVEATS:
// - The order of results is preserved, but the processing is done in parallel.
// - If AbortOnError is true, the first returned error will stop processing.
//   If multiple errors occur, only the first will be returned, and the rest will be ignored.
// - The reduction function in ParallelReduce should be associative -- Order is *not* guaranteed.

package toil
