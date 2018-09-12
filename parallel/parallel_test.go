package parallel

import (
	"fmt"
	"testing"
	"time"
)

func TestSomething(t *testing.T) {
	t.Parallel()
	time.Sleep(time.Second)
}

func TestA(t *testing.T) {
	t.Parallel()
	time.Sleep(time.Second)
}
func TestA2(t *testing.T) {
	t.Parallel()
	time.Sleep(time.Second)
}
func TestA3(t *testing.T) {
	t.Parallel()
	time.Sleep(time.Second)
}

func TestB(t *testing.T) {
	fmt.Println("setup")
	defer fmt.Println("deferred teardown")
	t.Run("group", func(t *testing.T) {
		t.Run("sub1", func(t *testing.T) {
			t.Parallel()
			// run sub1
			time.Sleep(time.Second)
			fmt.Println("sub1 done")
		})
		t.Run("sub2", func(t *testing.T) {
			t.Parallel()
			// run sub2
			time.Sleep(time.Second)
			fmt.Println("sub2 done")
		})
	})
	fmt.Println("teardown")
}

func TestGotcha(t *testing.T) {
	testCases := []struct {
		arg  int
		want int
	}{
		{2, 4},
		{3, 9},
		{4, 16},
	}
	for _, tc := range testCases {
		tc := tc // copy value for parallel tests - do not delete this!
		t.Run(fmt.Sprintf("arg=%d", tc.arg), func(t *testing.T) {
			t.Parallel()
			t.Logf("Testing with: arg=%d, want=%d", tc.arg, tc.want)
			if tc.arg*tc.arg != tc.want {
				t.Errorf("%d^2 != %d", tc.arg, tc.want)
			}
		})
	}
	// for i := 0; i < 10; i++ {
	// 	t.Run(fmt.Sprintf("i=%d", i), func(t *testing.T) {
	// 		t.Parallel()
	// 		t.Logf("Testing with i=%d", i)
	// 	})
	// }
}
