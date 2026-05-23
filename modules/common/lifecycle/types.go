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
//
// # Goroutine-dispatching exception within common/
//
// This package is the **single sanctioned exception** to the
// modules/common/ rule "no goroutines spawned by package functions, no
// log calls at the boundary, no side-effecting builders". It hosts the
// canonical lifecycle wiring helpers (Start, Stop, and the Build*
// family), so it must operate at the lifecycle boundary — that means
// dispatching background goroutines for Component starts and emitting
// "starting up" / "stopping" / "stopped" / "failed to start" /
// "shutdown failed" log lines via common/log.
//
// Every other package under modules/common/ MUST remain side-effect
// free at the function-call boundary. If a feature needs lifecycle of
// its own, it belongs in a top-level module (modules/http/,
// modules/cron/, modules/grpc/, modules/diagnostics/, etc.), not under
// common/. See modules/common/CODING_STANDARDS.md (section
// "common/lifecycle/ — Lone Goroutine-Dispatching Exception") for the
// full rule and review guidance.
package lifecycle

import (
	"context"
	"time"
)

var (
	_ Component = (*component)(nil)

	_ StartFn       = Start
	_ StopFn        = Stop
	_ BuildFn       = Build
	_ ErrStartFn    = ErrStart
	_ ErrShutdownFn = ErrShutdown
)

// StartFn is the function type for Start.
type StartFn func(ctx context.Context, component Component, errChan ErrChan) error

// StopFn is the function type for Stop.
type StopFn func(ctx context.Context, component Component, timeout time.Duration) error

// BuildFn is the function type for Build.
type BuildFn func(ctx context.Context, component Component, errChan ErrChan) (CloseFn, error)

// ErrStartFn is the function type for ErrStart.
type ErrStartFn func(errs ...error) error

// ErrShutdownFn is the function type for ErrShutdown.
type ErrShutdownFn func(errs ...error) error

// ErrChan is a send-only error channel used by callers to report runtime errors.
type ErrChan chan<- error

// CloseFn is the teardown callback returned by component builders. Callers
// invoke it to drain in-flight work bounded by the given timeout. It does
// not return an error: shutdown errors are logged at the builder boundary.
type CloseFn func(ctx context.Context, timeout time.Duration)

// Component is the contract for entities tied to the application's
// lifecycle.
//
// Implementations must be safe for concurrent use by multiple goroutines.
// Implementations must observe the five invariants below; deviations break
// the lifecycle.Build / Start / Stop helpers and the consumer contract.
//
// # Invariants
//
//  1. Stop is idempotent. Calling Stop more than once is safe: no panic,
//     no goroutine leak, no double-close of Done. The first call performs
//     the work; subsequent calls return without re-entering the shutdown
//     path. The returned error on the second call is implementation-
//     defined and may be non-nil (e.g. wrapped "already closed" from a
//     driver) — callers MUST NOT depend on it.
//  2. Stop does not block waiting on Done. Stop initiates shutdown and
//     returns (bounded by ctx). Use the Stop helper in this package and
//     wait on Done() separately if you need post-shutdown synchronization.
//  3. Start is server-style XOR worker-style.
//     - Server-style (e.g. http, grpc): Start blocks until the server has
//     shut down. Done MUST be closed when Start returns.
//     - Worker-style (e.g. cron, diagnostics): Start enables internal
//     state or dispatches goroutines and returns immediately. Done MUST
//     be closed after Stop completes.
//     The flavor is fixed per implementation and documented on its Start
//     method.
//  4. Done is closed exactly once. The closer is whichever runs first:
//     Stop, or the end of Start in server-style implementations. Use
//     sync.Once.Do(close(done)) — see component.go for the canonical
//     pattern.
//  5. Re-Start is not supported. A Component is single-use: once Stopped,
//     it MUST NOT be Started again. Construct a new Component instead.
//     Implementations are not required to guard against re-Start.
//
// # Testing the contract
//
// The lifecycle/tests subpackage publishes AssertIdempotentStop, which
// verifies invariants 1, 2, and 4 against any Component implementation.
// Every implementation under modules/ MUST have at least one test that
// invokes it.
type Component interface {
	// Name returns the component's identity used in logs.
	Name() string
	// Start begins the component execution. Per invariant 3, Start is
	// either server-style (blocks until shutdown) or worker-style (returns
	// after enabling internal state). The flavor is documented on each
	// concrete implementation.
	Start(ctx context.Context) error
	// Stop initiates graceful shutdown bounded by ctx's deadline.
	// Per invariant 1, Stop is safe to call multiple times. Per invariant
	// 2, Stop returns once shutdown has been initiated; callers wait on
	// Done() for completion.
	Stop(ctx context.Context) error
	// Done returns a channel that is closed exactly once when the
	// component has stopped. The closer is Stop, or the end of Start in
	// server-style implementations, whichever runs first.
	Done() <-chan struct{}
}
