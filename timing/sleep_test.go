package timing

import (
	"testing"
	"time"
)

func TestPollUntil(t *testing.T) {
	timeSleep = func(d time.Duration) {}
	defer func() {
		timeSleep = time.Sleep
	}()
	fn := func() bool {
		return false
	}
	err := PollUntil(fn, 2)
	if err != ErrExceededMaxTries {
		t.Errorf("PollUntil() err = %s, want %s", err, ErrExceededMaxTries)
	}
}

// See https://youtu.be/ndmB0bj7eyw?t=1917 as well

func TestPoller_Until(t *testing.T) {
	fn := func() bool {
		return false
	}
	p := Poller{
		sleep: func(time.Duration) {},
	}
	err := p.Until(fn, 2)
	if err != ErrExceededMaxTries {
		t.Errorf("PollUntil() err = %s, want %s", err, ErrExceededMaxTries)
	}
}
