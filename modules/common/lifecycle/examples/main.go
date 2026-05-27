// Package main demonstrates common/lifecycle: a worker-style Component
// implementation, the two-step Build / closeFn shutdown pattern, the
// errChan error reporting path, and the Done() synchronization
// guarantee. The Build helper emits structured log lines via
// common/log; the example does not register a logger so the default
// noop logger swallows the "starting up" / "stopping" / "stopped"
// breadcrumbs — what you see printed is exclusively from the demo
// component itself.
package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	clifecycle "github.com/guidomantilla/yarumo/common/lifecycle"
)

// Ticker is a worker-style Component that prints a heartbeat every interval
// until Stop is called. Start returns immediately after launching the
// background goroutine; Done is closed exactly once after Stop completes.
type Ticker struct {
	name     string
	interval time.Duration
	done     chan struct{}
	cancel   chan struct{}
	once     sync.Once
}

// NewTicker constructs a Ticker with the given name and interval.
func NewTicker(name string, interval time.Duration) *Ticker {
	return &Ticker{
		name:     name,
		interval: interval,
		done:     make(chan struct{}),
		cancel:   make(chan struct{}),
	}
}

// Name returns the component identity used in lifecycle log lines.
func (t *Ticker) Name() string { return t.name }

// Start launches the heartbeat loop in a background goroutine and returns.
func (t *Ticker) Start(_ context.Context) error {
	go func() {
		ticker := time.NewTicker(t.interval)
		defer ticker.Stop()

		count := 0

		for {
			select {
			case <-t.cancel:
				return
			case <-ticker.C:
				count++
				fmt.Printf("  [%s] tick %d\n", t.name, count)
			}
		}
	}()

	return nil
}

// Stop signals the heartbeat to terminate. It is safe to call multiple times.
func (t *Ticker) Stop(_ context.Context) error {
	t.once.Do(func() {
		close(t.cancel)
		close(t.done)
	})

	return nil
}

// Done returns a channel closed exactly once after Stop completes.
func (t *Ticker) Done() <-chan struct{} { return t.done }

func main() {
	ctx := context.Background()

	demoBuildAndClose(ctx)
	demoErrChan(ctx)
}

// demoBuildAndClose builds the worker, lets it tick a few times, and
// shuts it down through the returned closeFn.
func demoBuildAndClose(ctx context.Context) {
	fmt.Println("=== Build + closeFn ===")

	errChan := make(chan error, 1)

	ticker := NewTicker("heartbeat", 50*time.Millisecond)

	closeFn, err := clifecycle.Build(ctx, ticker, errChan)
	if err != nil {
		fmt.Printf("  build failed: %v\n", err)
		return
	}

	time.Sleep(180 * time.Millisecond)

	closeFn(ctx, time.Second)

	fmt.Println("  closeFn returned — Done is closed")
}

// demoErrChan starts a Component that fails Start and observes the error
// being delivered through errChan via lifecycle.Start.
func demoErrChan(ctx context.Context) {
	fmt.Println("=== errChan delivery ===")

	errChan := make(chan error, 1)

	failing := &failingComponent{name: "broken"}

	go func() {
		_ = clifecycle.Start(ctx, failing, errChan)
	}()

	err := <-errChan
	fmt.Printf("  errChan delivered: %v\n", err)
}

// failingComponent fails Start immediately. Used only to exercise the
// errChan path inside lifecycle.Start.
type failingComponent struct {
	name string
	done chan struct{}
	once sync.Once
}

func (f *failingComponent) Name() string { return f.name }

func (f *failingComponent) Start(_ context.Context) error {
	return fmt.Errorf("simulated startup failure")
}

func (f *failingComponent) Stop(_ context.Context) error {
	f.once.Do(func() {
		if f.done != nil {
			close(f.done)
		}
	})

	return nil
}

func (f *failingComponent) Done() <-chan struct{} {
	if f.done == nil {
		f.done = make(chan struct{})
	}

	return f.done
}
