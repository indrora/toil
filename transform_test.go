package toil

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestToilOptions_WithWorkers(t *testing.T) {
	opts := Options{}
	newOpts := opts.WithWorkers(5)

	if newOpts.workers != 5 {
		t.Errorf("Expected Workers to be 5, got %d", newOpts.workers)
	}

	// Ensure original options are unchanged
	if opts.workers != 0 {
		t.Errorf("Original options should not be modified, got Workers=%d", opts.workers)
	}
}

func TestToilOptions_WithAbortOnError(t *testing.T) {
	opts := Options{}
	newOpts := opts.StopOnError(true)

	if !newOpts.stopOnError {
		t.Errorf("Expected AbortOnError to be true, got %v", newOpts.stopOnError)
	}

	// Ensure original options are unchanged
	if opts.stopOnError {
		t.Errorf("Original options should not be modified, got AbortOnError=%v", opts.stopOnError)
	}
}

func TestToil_BasicFunctionality(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}

	// Simple doubling function
	double := func(x int) (int, error) {
		fmt.Fprintf(t.Output(), "Processing %v\n", x)
		return x * 2, nil
	}

	opts := Options{}.WithWorkers(2)
	results, err := ParallelTransform(&input, double, opts)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []int{2, 4, 6, 8, 10}
	if len(results) != len(expected) {
		t.Fatalf("Expected %d results, got %d", len(expected), len(results))
	}

	for i, result := range results {
		if result != expected[i] {
			t.Errorf("Expected result[%d] to be %d, got %d", i, expected[i], result)
		}
	}
}

func TestToil_DefaultWorkers(t *testing.T) {
	input := []int{1, 2, 3}

	identity := func(x int) (int, error) {
		return x, nil
	}

	// Test with zero workers (should default to runtime.NumCPU())
	opts := Options{}.WithWorkers(0)
	results, err := ParallelTransform(&input, identity, opts)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != len(input) {
		t.Fatalf("Expected %d results, got %d", len(input), len(results))
	}

	// Test with negative workers (should also default to runtime.NumCPU())
	opts.workers = -1
	results, err = ParallelTransform(&input, identity, opts)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != len(input) {
		t.Fatalf("Expected %d results, got %d", len(input), len(results))
	}
}

func TestToil_EmptyInput(t *testing.T) {
	input := []int{}

	identity := func(x int) (int, error) {
		return x, nil
	}

	opts := Options{workers: 2}
	results, err := ParallelTransform(&input, identity, opts)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("Expected 0 results for empty input, got %d", len(results))
	}
}

func TestToil_WithError_AbortOnError(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}

	// Function that errors on even numbers
	errorOnEven := func(x int) (int, error) {
		if x%2 == 0 {
			fmt.Fprintf(t.Output(), "Throwing error: %v is even\n", x)
			return 0, errors.New("even number error")
		}

		return x, nil
	}

	opts := Options{}.WithWorkers(2).StopOnError(true)
	results, err := ParallelTransform(&input, errorOnEven, opts)

	if err == nil {
		t.Fatal("Expected error but got none")
	}

	if results != nil {
		t.Errorf("Expected nil results when aborting on error, got %v", results)
	}
	fmt.Fprintf(t.Output(), "Error: %v\n", err)
	if err.Error() != "even number error" {
		t.Errorf("Expected 'even number error', got %v", err)
	}
}

func TestToil_WithError_ContinueOnError(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}

	// Function that errors on even numbers
	errorOnEven := func(x int) (int, error) {
		if x%2 == 0 {
			fmt.Fprintf(t.Output(), "Throwing error: %v is even\n", x)
			return 0, errors.New("even number error")
		}
		return x * 2, nil
	}

	opts := Options{}.WithWorkers(2).StopOnError(false)
	results, err := ParallelTransform(&input, errorOnEven, opts)

	// Should return an error but also results
	if err == nil {
		t.Fatal("Expected error but got none")
	}

	if results == nil {
		t.Fatal("Expected results even with error")
	}

	if len(results) != len(input) {
		t.Fatalf("Expected %d results, got %d", len(input), len(results))
	}

	// Check that odd numbers were processed correctly
	expected := []int{2, 0, 6, 0, 10} // Even indices should have zero values
	for i, result := range results {
		if result != expected[i] {
			t.Errorf("Expected result[%d] to be %d, got %d", i, expected[i], result)
		}
	}
}

func TestToil_NoError_ContinueOnErrorOption(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}

	// Function that never errors
	double := func(x int) (int, error) {
		return x * 2, nil
	}

	opts := Options{}.WithWorkers(2).StopOnError(false)
	results, err := ParallelTransform(&input, double, opts)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []int{2, 4, 6, 8, 10}
	if len(results) != len(expected) {
		t.Fatalf("Expected %d results, got %d", len(expected), len(results))
	}

	for i, result := range results {
		if result != expected[i] {
			t.Errorf("Expected result[%d] to be %d, got %d", i, expected[i], result)
		}
	}
}

func TestToil_WithSlowFunctions(t *testing.T) {
	input := []int{1, 2, 3}

	slowFunc := func(x int) (int, error) {
		time.Sleep(10 * time.Millisecond) // Small delay
		return x * 2, nil
	}

	start := time.Now()
	opts := Options{workers: 3} // Parallel execution
	results, err := ParallelTransform(&input, slowFunc, opts)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []int{2, 4, 6}
	for i, result := range results {
		if result != expected[i] {
			t.Errorf("Expected result[%d] to be %d, got %d", i, expected[i], result)
		}
	}

	// Should complete faster than sequential execution (which would be ~30ms)
	if duration > 25*time.Millisecond {
		t.Errorf("Parallel execution took too long: %v", duration)
	}
}

func TestToil_WorkerLimiting(t *testing.T) {
	input := make([]int, 100)
	for i := range input {
		input[i] = i + 1
	}

	maxWorkers := 3
	currentWorkers := 0
	maxConcurrent := 0
	var mu sync.Mutex

	countingFunc := func(x int) (int, error) {
		mu.Lock()
		currentWorkers++
		if currentWorkers > maxConcurrent {
			maxConcurrent = currentWorkers
		}
		mu.Unlock()

		time.Sleep(5 * time.Millisecond) // Small delay to ensure concurrency

		mu.Lock()
		currentWorkers--
		mu.Unlock()

		return x, nil
	}

	opts := Options{workers: maxWorkers}
	_, err := ParallelTransform(&input, countingFunc, opts)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if maxConcurrent > maxWorkers {
		t.Errorf("Worker limit exceeded: max concurrent was %d, limit was %d", maxConcurrent, maxWorkers)
	}
}

func TestToil_DifferentTypes(t *testing.T) {
	// Test string to int conversion
	input := []string{"1", "2", "3", "4"}

	stringToInt := func(s string) (int, error) {
		switch s {
		case "1":
			return 1, nil
		case "2":
			return 2, nil
		case "3":
			return 3, nil
		case "4":
			return 4, nil
		default:
			return 0, errors.New("invalid number")
		}
	}

	opts := Options{}.WithWorkers(2)
	results, err := ParallelTransform(&input, stringToInt, opts)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []int{1, 2, 3, 4}
	for i, result := range results {
		if result != expected[i] {
			t.Errorf("Expected result[%d] to be %d, got %d", i, expected[i], result)
		}
	}
}

func BenchmarkToil_Sequential(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = i + 1
	}

	square := func(x int) (int, error) {
		return x * x, nil
	}

	opts := Options{workers: 1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParallelTransform(&input, square, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkToil_Parallel(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = i + 1
	}

	square := func(x int) (int, error) {
		return x * x, nil
	}

	opts := Options{}.WithWorkers(runtime.NumCPU())
	b.ResetTimer()

	for b.Loop() {
		_, err := ParallelTransform(&input, square, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkToil_Sized(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000, 100000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
			input := make([]int, size)
			for i := range input {
				input[i] = i + 1
			}

			square := func(x int) (int, error) {
				return x * x, nil
			}

			opts := Options{}.WithWorkers(runtime.NumCPU())
			b.ResetTimer()

			for b.Loop() {
				start := time.Now()
				_, err := ParallelTransform(&input, square, opts)
				duration := time.Since(start)
				b.ReportMetric((float64(size) / float64(duration.Nanoseconds())), "item/ns")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
