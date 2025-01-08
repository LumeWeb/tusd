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

	// Capture the stack trace when the context is created
	creationStack := string(debug.Stack())
	creationTime := time.Now()

	go func() {
		select {
		case <-ctx.Done():
			cancelTime := time.Now()
			fmt.Printf(`
Context %s cancelled:
Created at: %s
Cancelled at: %s
Time alive: %s
Creation stack:
%s
Cancellation stack:
%s
Error: %v
`, name,
				creationTime.Format(time.RFC3339Nano),
				cancelTime.Format(time.RFC3339Nano),
				cancelTime.Sub(creationTime),
				creationStack,
				string(debug.Stack()),
				ctx.Err())
		case <-traceDone:
			// Context is being cleaned up normally
		}
	}()

	return &TracedContext{
		Context: ctx,
		cleanup: func() {
			close(traceDone)
		},
		name:      name,
		createdAt: creationTime,
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
	createdAt   time.Time
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
func NewCancelContext(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)

	wrappedCancel := func() {
		// Only print stack trace if we're not already cancelled
		select {
		case <-ctx.Done():
			// Context is already cancelled, just propagate the cancellation
			cancel()
		default:
			// This is a new cancellation, print the stack trace
			fmt.Printf("Cancel called for context from:\n%s\n", string(debug.Stack()))
			cancel()
		}
	}

	return ctx, wrappedCancel
}
