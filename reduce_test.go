package toil

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"testing"
)

func TestParallelReduce_Sum(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sumFunc := func(a, b int) (int, error) { return a + b, nil }
	opts := Options{}.WithWorkers(3)
	result, err := ParallelReduce(&input, sumFunc, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != 55 {
		t.Errorf("Expected sum 55, got %d", result)
	}
}

func TestParallelReduce_Product(t *testing.T) {
	input := []int{1, 2, 3, 4}
	prodFunc := func(a, b int) (int, error) { return a * b, nil }
	opts := Options{}.WithWorkers(2)
	result, err := ParallelReduce(&input, prodFunc, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != 24 {
		t.Errorf("Expected product 24, got %d", result)
	}
}

func TestParallelReduce_Empty(t *testing.T) {
	input := []int{}
	sumFunc := func(a, b int) (int, error) { return a + b, nil }
	opts := Options{}.WithWorkers(2)
	result, err := ParallelReduce(&input, sumFunc, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != 0 {
		t.Errorf("Expected zero value for empty input, got %d", result)
	}
}

func TestParallelReduce_SingleElement(t *testing.T) {
	input := []int{42}
	sumFunc := func(a, b int) (int, error) { return a + b, nil }
	opts := Options{}.WithWorkers(2)
	result, err := ParallelReduce(&input, sumFunc, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != 42 {
		t.Errorf("Expected 42 for single element, got %d", result)
	}
}

func TestParallelReduce_Error(t *testing.T) {
	input := []int{1, 2, 3, 4}
	errFunc := func(a, b int) (int, error) {
		if a == 2 || b == 2 {
			return 0, errors.New("fail on 2")
		}
		return a + b, nil
	}
	opts := Options{}.WithWorkers(2)
	_, err := ParallelReduce(&input, errFunc, opts)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestParallelReduce_NonAssociative(t *testing.T) {
	input := []int{1, 2, 3, 4}
	subFunc := func(a, b int) (int, error) { return a - b, nil }
	opts := Options{}.WithWorkers(2)
	// Result is not well-defined for non-associative functions, but should not panic or deadlock
	_, err := ParallelReduce(&input, subFunc, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestParallelReduce_ParallelCorrectness(t *testing.T) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = 1
	}
	sumFunc := func(a, b int) (int, error) { return a + b, nil }
	opts := Options{}.WithWorkers(8)
	result, err := ParallelReduce(&input, sumFunc, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != 1000 {
		t.Errorf("Expected sum 1000, got %d", result)
	}
}

func BenchmarkParallelReduce_HeavySum(b *testing.B) {
	sizes := []int{1000, 10000, 100000, 1000000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
			input := make([]int, size)
			for i := range input {
				input[i] = i + 1
			}

			// Heavy computation in sum function
			heavySum := func(a, b int) (int, error) {
				// Simulate expensive computation
				for i := 0; i < 100; i++ {
					a = (a * 31) % 1000000007
					b = (b * 37) % 1000000007
				}
				return a + b, nil
			}

			opts := Options{}.WithWorkers(runtime.NumCPU())
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := ParallelReduce(&input, heavySum, opts)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkParallelReduce_WorkerScaling(b *testing.B) {
	size := 50000
	input := make([]int, size)
	for i := range input {
		input[i] = i + 1
	}

	// Very heavy computation function
	heavyReduce := func(a, b int) (int, error) {
		// Simulate cryptographic-like operations
		result := a ^ b
		for i := 0; i < 500; i++ {
			result = ((result * 1103515245) + 12345) & 0x7fffffff
			result ^= (result >> 16)
			result = ((result * 1103515245) + 12345) & 0x7fffffff
		}
		return result, nil
	}

	workerCounts := []int{1, 2, 4, 8, 16, runtime.NumCPU()}

	for _, workers := range workerCounts {
		b.Run(fmt.Sprintf("Workers%d", workers), func(b *testing.B) {
			opts := Options{}.WithWorkers(workers)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := ParallelReduce(&input, heavyReduce, opts)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkParallelReduce_FloatPrecision(b *testing.B) {
	size := 100000
	input := make([]float64, size)
	for i := range input {
		input[i] = float64(i+1) * 0.00001
	}

	// Heavy floating point operations
	precisionSum := func(a, b float64) (float64, error) {
		// Simulate expensive math operations
		for i := 0; i < 200; i++ {
			a = math.Sin(a) + math.Cos(b)
			b = math.Sqrt(math.Abs(a*b)) + math.Log(math.Abs(a)+1)
		}
		return a + b, nil
	}

	opts := Options{}.WithWorkers(runtime.NumCPU())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ParallelReduce(&input, precisionSum, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParallelReduce_StringConcat(b *testing.B) {
	size := 10000
	input := make([]string, size)
	for i := range input {
		input[i] = fmt.Sprintf("chunk_%d_", i)
	}

	// Heavy string processing
	heavyConcat := func(a, b string) (string, error) {
		// Simulate complex string operations
		result := a + b
		for i := 0; i < 50; i++ {
			result = fmt.Sprintf("%x", []byte(result))[:len(result)]
			if len(result) > 1000 {
				result = result[:1000] // Prevent excessive growth
			}
		}
		return result, nil
	}

	opts := Options{}.WithWorkers(runtime.NumCPU())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ParallelReduce(&input, heavyConcat, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}
