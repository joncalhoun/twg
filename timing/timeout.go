package timing

import (
	"errors"
	"time"
)

var (
	timeoutMultiplier time.Duration = 1
	timeAfter                       = time.After
)

var (
	ErrTimeout = errors.New("timing: operation timed out")
)

func DoThing() error {
	done := make(chan bool)
	go func() {
		// do some work
		// In the real code, this woudl be an interaction with an external API
		// or a database, or something else
		time.Sleep(500 * time.Millisecond) // pretend this hangs
		done <- true
	}()

	select {
	case <-done:
		return nil
	case <-timeAfter(100 * time.Millisecond * timeoutMultiplier):
		return ErrTimeout
	}
}
