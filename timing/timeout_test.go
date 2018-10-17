package timing

import (
	"flag"
	"testing"
	"time"
)

var (
	timeoutFlag int64
)

func init() {
	flag.Int64Var(&timeoutFlag, "timeoutMultiplier", 1, "set this to a higher value if you are on a slower pc and experiencing test failure due to unexpected timeouts")
}

// This test intentionally fails
func TestDoThing(t *testing.T) {
	timeoutMultiplier = time.Duration(timeoutFlag)
	err := DoThing()
	if err != nil {
		t.Errorf("DoThing() err = %s; want nil", err)
	}
}

func TestDoThing_inject(t *testing.T) {
	timeAfter = func(time.Duration) <-chan time.Time {
		ch := make(chan time.Time)
		go func() { ch <- time.Now() }()
		return ch
	}
	err := DoThing()
	if err == nil {
		t.Errorf("DoThing() err = nil; want %s", ErrTimeout)
	}

	timeAfter = func(time.Duration) <-chan time.Time {
		ch := make(chan time.Time)
		return ch
		// return time.After(10 * time.Minute)
	}
	err = DoThing()
	if err != nil {
		t.Errorf("DoThing() err = %s; want nil", ErrTimeout)
	}

	timeAfter = time.After
}
