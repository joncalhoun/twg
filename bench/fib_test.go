package bench

import (
	"sync"
	"testing"
)

func TestFib(t *testing.T) {
	tests := []struct {
		arg  int
		want int
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{3, 2},
		{4, 3},
		{5, 5},
		{6, 8},
		{7, 13},
		{8, 21},
		{9, 34},
		{10, 55},
		{11, 89},
		{12, 144},
		{13, 233},
		{14, 377},
		{15, 610},
		{16, 987},
		{17, 1597},
		{18, 2584},
		{19, 4181},
		{20, 6765},
	}
	funcs := map[string]func(int) int{
		"Recursive": FibRecursive,
		"Iterative": FibIterative,
		"Memo":      FibMemo,
	}
	for name, fn := range funcs {
		t.Run(name, func(t *testing.T) {
			for _, tc := range tests {
				got := fn(tc.arg)
				if got != tc.want {
					t.Errorf("Fib(%d) = %d; want %d", tc.arg, got, tc.want)
				}
			}
		})
	}
}

// Run this with the -race flag
func TestFibMemoThreadsafe_threadsafe(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(20)
	for i := 0; i < 20; i++ {
		go func() {
			FibMemoThreadsafe(125)
			wg.Done()
		}()
	}
	wg.Wait()
}

func benchmarkFib(b *testing.B, fib func(int) int, n int) {
	// setup that takes some time
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fib(n)
	}
	b.StopTimer()
	// teardown that takes time
}

func BenchmarkFibRecursive5(b *testing.B)  { benchmarkFib(b, FibRecursive, 5) }
func BenchmarkFibRecursive10(b *testing.B) { benchmarkFib(b, FibRecursive, 10) }
func BenchmarkFibRecursive20(b *testing.B) { benchmarkFib(b, FibRecursive, 20) }

func BenchmarkFibIterative5(b *testing.B)   { benchmarkFib(b, FibIterative, 5) }
func BenchmarkFibIterative10(b *testing.B)  { benchmarkFib(b, FibIterative, 10) }
func BenchmarkFibIterative20(b *testing.B)  { benchmarkFib(b, FibIterative, 20) }
func BenchmarkFibIterative100(b *testing.B) { benchmarkFib(b, FibIterative, 100) }
func BenchmarkFibIterative500(b *testing.B) { benchmarkFib(b, FibIterative, 500) }

func BenchmarkFibMemo5(b *testing.B)   { benchmarkFib(b, FibMemo, 5) }
func BenchmarkFibMemo10(b *testing.B)  { benchmarkFib(b, FibMemo, 10) }
func BenchmarkFibMemo20(b *testing.B)  { benchmarkFib(b, FibMemo, 20) }
func BenchmarkFibMemo100(b *testing.B) { benchmarkFib(b, FibMemo, 100) }
func BenchmarkFibMemo500(b *testing.B) { benchmarkFib(b, FibMemo, 500) }

// I added a threadsafe implementation to demonstrate that adding
// thread-safety still results in a constant benchmark like the original
// memo, but that benchmark is much slower due to the addition of mutexes.
func BenchmarkFibMemoThreadsafe5(b *testing.B)   { benchmarkFib(b, FibMemoThreadsafe, 5) }
func BenchmarkFibMemoThreadsafe10(b *testing.B)  { benchmarkFib(b, FibMemoThreadsafe, 10) }
func BenchmarkFibMemoThreadsafe20(b *testing.B)  { benchmarkFib(b, FibMemoThreadsafe, 20) }
func BenchmarkFibMemoThreadsafe100(b *testing.B) { benchmarkFib(b, FibMemoThreadsafe, 100) }
func BenchmarkFibMemoThreadsafe500(b *testing.B) { benchmarkFib(b, FibMemoThreadsafe, 500) }
