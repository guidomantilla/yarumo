// Package scattergather provides the Scatter-Gather EIP pattern over
// messaging.Channel[T] and messaging.Channel[U].
//
// Scatter-Gather is the composition of two integration patterns:
//
//   - Scatter — a Recipient List fans out one source Message[T] to N
//     worker destinations chosen by a caller-supplied SelectorFn[T].
//   - Gather — an Aggregator collects each worker's reply Message[T]
//     into a group keyed by CorrelationID and folds the complete group
//     into a single Message[U] via the caller-supplied AggregateFn[T,U],
//     which is then forwarded to a final destination Channel[U].
//
// The ScatterGather component owns both halves and a per-correlation
// expected-size map so each in-flight gather knows exactly how many
// worker replies to wait for (the selector can return a different
// number of destinations per request). The map is updated at scatter
// time (when the selector fires) and consulted by a custom completion
// function injected into the internal Aggregator.
//
// # Worker contract
//
// Workers consume from their per-key destination Channel[T] and publish
// their reply to the shared replyChan with the SAME CorrelationID as
// the incoming request. This is a CONTRACT the caller honors — the
// Aggregator groups replies by CorrelationID and the pattern cannot
// distinguish replies that drop the header from unrelated traffic.
//
// # Bounds and observability
//
// Two safeguards keep an in-flight gather from leaking memory or
// stalling forever:
//
//   - WithMaxConcurrentScatters caps the number of simultaneously
//     tracked correlations (default 1000). A request that would create
//     the n+1-th entry is rejected via WithErrorHandler with
//     ErrMaxScattersExceeded; existing in-flight gathers are untouched.
//   - WithGroupTimeout is REQUIRED and bounds the lifetime of any
//     single gather: a partial group released by the internal
//     Aggregator's sweeper drops via WithDropHandler and the
//     correlation entry is cleaned up.
//
// Two observability hooks separate failures from intentional drops:
//
//   - WithErrorHandler routes real failures (AggregateFn failed,
//     missing worker key, forward Send failed, MaxScatters exceeded).
//   - WithDropHandler routes intentional drops (empty selector,
//     timed-out partial gather).
//
// # Lifecycle
//
// ScatterGather implements common/lifecycle.Component (worker-style).
// Start spawns the internal Aggregator first (so it is ready to
// receive replies before any scatter happens) and then the internal
// Recipient List. Stop reverses the order: it stops the source
// interceptor first (no new requests), then the internal Aggregator
// (which drains in-flight gathers via the standard release path —
// timed-out groups land in WithDropHandler). Done closes after both
// halves have drained.
//
// # Error handling
//
// The handler installed on the source channel always returns nil per
// the package-wide policy in modules/messaging/CODING_STANDARDS.md.
// Scatter-Gather concerns flow through WithErrorHandler /
// WithDropHandler; nothing propagates to the source channel's Send
// caller.
package scattergather

import (
	"context"
	"sync"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
	"github.com/guidomantilla/yarumo/messaging/patterns/routers/aggregator"
	"github.com/guidomantilla/yarumo/messaging/patterns/routers/recipientlist"
)

var (
	_ ScatterGather[any, any] = (*scatterGather[any, any])(nil)

	_ ErrScatterGatherFn = ErrScatterGather
)

// ScatterGather is the public interface for a Scatter-Gather pattern.
// It embeds lifecycle.Component so callers wire it up with
// lifecycle.Build. The interface exists (rather than returning
// lifecycle.Component directly) so the consumer's API surface
// preserves "this is a ScatterGather" semantics and the type stays
// open to future scatter-gather-specific methods without breaking
// callers.
type ScatterGather[T, U any] interface {
	lifecycle.Component
}

// SelectorFn is the scatter rule: it returns the ordered list of
// worker destination keys for the incoming request msg. Each key is
// looked up in the workers map passed to NewScatterGather. An empty
// slice means "no recipients" and routes the message to the
// DropHandler with no scatter. Reuses recipientlist.SelectorFn so the
// caller writes one selector type for both patterns.
type SelectorFn[T any] = recipientlist.SelectorFn[T]

// AggregateFn is the gather rule: it folds the collected worker
// replies into a single Message[U] for publication to the aggregated
// destination channel. Reuses aggregator.AggregateFn so the caller
// writes one aggregate type for both patterns.
type AggregateFn[T, U any] = aggregator.AggregateFn[T, U]

// DropHandler is the optional observability hook invoked once per
// intentional drop. msg is type-erased; cast it inside the hook when
// payload-specific behavior is needed. The hook is invoked
// synchronously from the source channel's dispatcher or from the
// internal Aggregator's sweeper and must not block — long
// observability work should be dispatched asynchronously by the
// implementer.
//
// DropHandler fires on two deliberate-drop events: empty selector
// result (no workers chosen for the request) and partial gather
// released by the Aggregator's sweeper after the group timeout
// elapsed without all expected replies arriving. It does NOT fire on
// real failures (missing worker key, AggregateFn failed, forward Send
// failed, MaxScatters exceeded), which route to the ErrorHandler
// instead.
type DropHandler func(ctx context.Context, msg any)

// ErrScatterGatherFn is the function type for ErrScatterGather.
type ErrScatterGatherFn func(causes ...error) error

// expectation tracks one in-flight gather: the number of replies the
// internal Aggregator must collect before releasing the group, and
// the moment the scatter happened so the orphan sweeper can evict
// entries whose workers never replied at all.
type expectation struct {
	count       int
	scatteredAt time.Time
}

// scatterGather is the Scatter-Gather pattern implementation. It owns
// an internal Recipient List (the scatter half), an internal
// Aggregator (the gather half), and a per-correlation expected-size
// map so each in-flight gather knows how many replies to wait for.
// A background orphan sweeper evicts entries whose workers never
// replied so the map cannot grow unbounded when MaxConcurrentScatters
// has not been hit yet.
type scatterGather[T, U any] struct {
	name         string
	src          messaging.Channel[T]
	workers      map[string]messaging.Channel[T]
	replyChan    messaging.Channel[T]
	aggregateDst messaging.Channel[U]
	selector     SelectorFn[T]
	aggregate    AggregateFn[T, U]

	groupTimeout          time.Duration
	maxConcurrentScatters int
	errorHandler          messaging.ErrorHandler
	dropHandler           DropHandler

	scatterer recipientlist.RecipientList[T]
	gatherer  aggregator.Aggregator[T, U]

	done         chan struct{}
	startOnce    sync.Once
	stopOnce     sync.Once
	doneOnce     sync.Once
	workerWG     sync.WaitGroup
	workerCancel context.CancelFunc

	mu       sync.Mutex
	expected map[string]expectation
}
