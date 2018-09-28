package main

import (
	"context"
	"fmt"
	"time"
)

type Context struct {
	context.Context
	doneCh chan struct{}
	done   bool
}

func (c *Context) Done() <-chan struct{} {
	return c.doneCh
}

func (c *Context) Err() error {

}

var contextWithTimeout = context.WithTimeout

func main() {
	doneCh := make(chan struct{})
	testCtx := &Context{
		doneCh: doneCh,
	}
	contextWithTimeout = func(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
		testCtx.Context = ctx
		return context.WithCancel(testCtx)
	}

	// ctx, cancel := contextWithTimeout(context.Background(), 50*time.Millisecond)
	// cancel()
	// select {
	// case <-time.After(100 * time.Millisecond):
	// 	fmt.Println("overslept")
	// case <-ctx.Done():
	// 	fmt.Println(ctx.Err()) // prints "context cancelled"
	// }

	go func() { doneCh <- struct{}{} }()
	ctx, cancel := contextWithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	select {
	case <-time.After(100 * time.Millisecond):
		fmt.Println("overslept")
	case <-ctx.Done():
		fmt.Println(ctx.Err()) // prints "context cancelled"
	}
}
