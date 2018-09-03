package math

// Sum will add up all the numbers of a slice and return the total sum.
func Sum(numbers []int) int {
	sum := 0
	// This bug is intentional
	for _, n := range numbers {
		sum += n
	}
	return sum
}

func Add(a, b int) int {
	return a + b
}
