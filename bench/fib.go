package bench

// The first few Fibonacci numbers are:
// 1, 1, 2, 3, 5, 8, 13, ... where F(1) = 1, F(2) = 1, and
// F(3) = F(2) + F(1) = 2, and more generally
// F(N) = F(N-1) + F(N-2)
// F(0) isn't really a value, but if we make it equal to 0 we can also say
// that F(2) = F(1) + F(0) = 1 + 0 = 1. This simplifies our recursive
// solution a bit.

// FibRecursive will calculate the Fibonacci number recursively.
func FibRecursive(n int) int {
	if n < 2 {
		return n
	}
	return FibRecursive(n-1) + FibRecursive(n-2)
}

// FibIterative will calculate the Fibonacci number iteratively.
func FibIterative(n int) int {
	if n == 0 {
		return 0
	}
	a, b := 0, 1
	for i := 1; i < n; i++ {
		a, b = b, a+b
	}
	return b
}

var (
	memo = []int{0, 1}
)

// FibMemo will calculate the Fibonacci number recursively using a pkg
// level memo variable. This is NOT threadsafe.
func FibMemo(n int) int {
	if len(memo) <= n {
		FibMemo(n - 1)                                 // make sure the memo is filled in up to n-1
		memo = append(memo, FibMemo(n-1)+FibMemo(n-2)) // append the nth value
	}
	return memo[n]
}
