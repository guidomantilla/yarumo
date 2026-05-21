// Package lifecycle provides primitives for entities tied to the
// application's lifecycle: things that hold resources (sockets, file
// handles, goroutines, external SDK lifetimes) which cannot be reclaimed
// by the garbage collector and must be explicitly stopped.
//
// A Component is born with NewComponent (or any package-specific
// constructor) and dies when Stop is called. Plain objects without
// lifecycle are not Components — they live and die with the goroutine
// that constructed them.
//
// Implementations choose how Start behaves:
//   - Server-style: Start blocks until shutdown (e.g., ListenAndServe).
//     Done MUST be closed when Start returns.
//   - Worker-style: Start kicks off internal goroutines and returns
//     immediately. Done MUST be closed after Stop completes.
//
// Callers should always wait on Done() for true completion regardless of
// Start's flavor. The Start helper in this package does exactly that.
package lifecycle

import (
	"context"
	"time"
)

var (
	_ Component = (*component)(nil)

	_ StartFn       = Start
	_ StopFn        = Stop
	_ ErrStartFn    = ErrStart
	_ ErrShutdownFn = ErrShutdown
)

// StartFn is the function type for Start.
type StartFn func(ctx context.Context, component Component, errChan ErrChan) error

// StopFn is the function type for Stop.
type StopFn func(ctx context.Context, component Component, timeout time.Duration) error

// ErrStartFn is the function type for ErrStart.
type ErrStartFn func(errs ...error) error

// ErrShutdownFn is the function type for ErrShutdown.
type ErrShutdownFn func(errs ...error) error

// ErrChan is a send-only error channel used by callers to report runtime errors.
type ErrChan chan<- error

// Component is the contract for entities tied to the application's lifecycle.
//
// Implementations must be safe for concurrent use by multiple goroutines.
// See the package documentation for the Start blocking-vs-non-blocking contract.
type Component interface {
	// Name returns the component's identity used in logs.
	Name() string
	// Start begins the component execution. See package docs for the
	// blocking-vs-non-blocking contract.
	Start(ctx context.Context) error
	// Stop gracefully stops the component bounded by ctx's deadline.
	Stop(ctx context.Context) error
	// Done returns a channel that is closed when the component has stopped.
	Done() <-chan struct{}
}
