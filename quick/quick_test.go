package quick

import (
	"testing"
	"testing/quick"
)

func TestSquareAndAdd(t *testing.T) {
	f := func(a, b int) bool {
		got := SquareAndAdd(a, b)
		return got >= 0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
