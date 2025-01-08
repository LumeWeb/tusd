package internal

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

func TraceContext(ctx context.Context, name string) context.Context {
	traceDone := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			fmt.Printf("Context %s cancelled due to: %v\nStack trace:\n%s\n",
				name, ctx.Err(), string(debug.Stack()))
		case <-traceDone:
			// Context is being cleaned up normally
		}
	}()

	return &TracedContext{
		Context: ctx,
		cleanup: func() {
			close(traceDone)
		},
		name: name,
	}
}

// Fixed TraceMultiContext implementation
func TraceMultiContext(ctx context.Context, name string, parents ...context.Context) context.Context {
	traceDone := make(chan struct{})

	for i, parent := range parents {
		go func(i int, parent context.Context) {
			select {
			case <-parent.Done():
				fmt.Printf("Parent context %d for %s cancelled due to: %v\nStack trace:\n%s\n",
					i, name, parent.Err(), string(debug.Stack()))
			case <-traceDone:
				// Context is being cleaned up normally
			}
		}(i, parent)
	}

	return &TracedContext{
		Context: ctx,
		cleanup: func() {
			close(traceDone)
		},
		name: name,
	}
}

type TracedContext struct {
	context.Context
	cleanup     func()
	once        sync.Once
	cleanupOnce sync.Once
	name        string
}

func (c *TracedContext) Done() <-chan struct{} {
	return c.Context.Done()
}

func (c *TracedContext) Err() error {
	return c.Context.Err()
}

func (c *TracedContext) Deadline() (time.Time, bool) {
	return c.Context.Deadline()
}

func (c *TracedContext) Value(key interface{}) interface{} {
	return c.Context.Value(key)
}

func (c *TracedContext) Cleanup() {
	c.cleanupOnce.Do(c.cleanup)
}
