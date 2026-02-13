package toil

// The Options struct defines the configuration for parallel processing in the toil package.
type Options struct {
	workers     int
	stopOnError bool
}

// Define the numberof workers to use. If this value is 0 or a negative value, the number of CPU cores will be used.
func (o Options) WithWorkers(workers int) Options {
	o.workers = workers
	return o
}

// Define whether to stop on error. If true, the processing will stop immediately if any error occurs.
// If false, the processing will continue even if errors occur, and the first error encountered will be returned after all processing is complete.
func (o Options) StopOnError(stopOnError bool) Options {
	o.stopOnError = stopOnError
	return o
}
