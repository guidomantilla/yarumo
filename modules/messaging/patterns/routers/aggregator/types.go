// Package aggregator provides the Aggregator EIP pattern over
// messaging.Channel[T] and messaging.Channel[U].
//
// An Aggregator subscribes to a source Channel[T] and collects every
// received Message[T] into a group keyed by a correlation key extracted
// from the message (Headers.CorrelationID by default). When a group is
// declared complete by the configured CompletionStrategy, the Aggregator
// folds the accumulated messages into a single Message[U] via the
// caller-supplied AggregateFn and forwards it to a destination
// Channel[U]. The pattern is the canonical N→1 inverse of Splitter.
//
// # Completion strategies
//
// Three orthogonal completion strategies are supported and may be
// combined freely. A group is released the first time ANY enabled
// strategy fires:
//
//   - WithCompletionSize(n): release when the group reaches n messages.
//     The simplest strategy and the one most consumers want for batch
//     pagination, fixed-fan-in scatter/gather and similar shapes.
//   - WithCompletionFn(fn): release when fn(group) returns true. Use it
//     for predicate-based completion that depends on payload contents
//     (e.g. "saw an END marker", "total bytes ≥ threshold").
//   - WithGroupTimeout(d): release any group that sits idle for d since
//     the last message arrived. Required to bound memory when partial
//     groups may never reach size or predicate completion. Implemented
//     by a background sweeper goroutine spawned in Start.
//
// At least one strategy MUST be configured. Constructing an Aggregator
// without any strategy is a caller bug and panics in NewAggregator.
//
// # Memory bounding
//
// Two safeguards keep an Aggregator from leaking memory under abuse or
// pathological traffic:
//
//   - WithMaxGroups(n) caps the number of concurrently tracked groups
//     (default 1000). A message that would create the n+1-th group is
//     rejected via the ErrorHandler with ErrMaxGroupsExceeded; existing
//     groups remain untouched.
//   - WithGroupTimeout(d) caps the lifetime of any single group; the
//     sweeper releases (or drops) expired groups regardless of their
//     completion status.
//
// # Two observability hooks (shape B)
//
// The Aggregator separates "the pattern could not do its job" from
// "the pattern dropped a message on purpose":
//
//   - WithErrorHandler routes real failures (AggregateFn returned error,
//     AggregateFn panicked, MaxGroups exceeded, forward Send failed).
//     The default is messaging.DefaultErrorHandler which logs via
//     common/log — wire messaging.SilentErrorHandler to opt out.
//   - WithDropHandler routes intentional drops (empty correlation key,
//     expired groups released by the sweeper when consumers want to
//     audit "timed-out groups"). Default is nil — silent drop.
//
// # Lifecycle
//
// Aggregator implements common/lifecycle.Component (worker-style). Start
// registers the subscription on the source channel and spawns the
// sweeper goroutine when a group timeout is configured. Stop cancels
// the subscription, signals the sweeper to exit, waits for it, and
// drains remaining groups by releasing them through the same code path
// as a normal completion so partial groups land in the destination
// (rather than being silently dropped) — failures during drain are
// surfaced via WithErrorHandler. Done closes after drain completes.
//
// # Error handling
//
// The handler installed on the source channel always returns nil.
// Aggregator concerns (AggregateFn failure/panic, MaxGroups exceeded,
// forward Send failure, sweeper-released expired groups) flow through
// WithErrorHandler / WithDropHandler; nothing propagates to the source
// channel's Send caller error path. Consistent with the package-wide
// policy in modules/messaging/CODING_STANDARDS.md.
package aggregator

import (
	"context"
	"sync"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

var (
	_ Aggregator[any, any] = (*aggregator[any, any])(nil)

	_ ErrAggregatorFn = ErrAggregator
)

// Aggregator is the public interface for an Aggregator pattern. It
// embeds lifecycle.Component so callers wire it up with lifecycle.Build.
// The interface exists (rather than returning lifecycle.Component
// directly) so the consumer's API surface preserves "this is an
// Aggregator" semantics and the type stays open to future aggregator-
// specific methods without breaking callers.
type Aggregator[T, U any] interface {
	lifecycle.Component
}

// CorrelationFn extracts the correlation key from msg. The default
// implementation (used when WithCorrelationFn is not passed) returns
// msg.Headers.CorrelationID. An empty string skips aggregation: the
// message is reported via WithDropHandler and discarded — this lets
// upstream code that forgets to set a correlation id surface the bug
// instead of having every uncorrelated message land in the same
// global bucket.
type CorrelationFn[T any] func(msg messaging.Message[T]) string

// CompletionFn decides whether group is complete given its current
// accumulated messages. Returns true to release the group. The
// function runs under the Aggregator's lock while iterating the group;
// it must be cheap and side-effect-free.
type CompletionFn[T any] func(group []messaging.Message[T]) bool

// AggregateFn combines a complete group of messages into a single
// Message[U] for publication to the destination channel. The function
// receives a snapshot of the group's messages (safe to retain). A
// non-nil error is wrapped in ErrAggregator(ErrAggregateFnFailed, err)
// and forwarded to the ErrorHandler; a panic is recovered and wrapped
// in ErrAggregator(ErrAggregateFnFailed, ...) so it cannot kill the
// caller's goroutine.
type AggregateFn[T, U any] func(group []messaging.Message[T]) (messaging.Message[U], error)

// DropHandler is the optional observability hook invoked once per
// intentional drop. msg is type-erased; cast it inside the hook when
// payload-specific behavior is needed. The hook is invoked from the
// source channel's dispatcher goroutine or from the sweeper goroutine
// and must not block — long observability work should be dispatched
// asynchronously by the implementer.
//
// DropHandler fires on three deliberate-drop events: empty correlation
// key from CorrelationFn, group expired by the sweeper without
// completing (one DropHandler call per expired-and-empty group), and
// — for visibility — the drained partial groups released during Stop.
// It does NOT fire on real failures (AggregateFn error/panic, forward
// Send failure, MaxGroups exceeded), which route to the ErrorHandler
// instead.
type DropHandler func(ctx context.Context, msg any)

// ErrAggregatorFn is the function type for ErrAggregator.
type ErrAggregatorFn func(causes ...error) error

// group is the internal accumulator for a single correlation key. Its
// firstSeen / lastSeen fields drive the timeout sweeper. The mu of the
// owning aggregator protects all reads and writes of group; group has
// no per-instance lock so the aggregator's single map+lock pair is the
// only synchronisation point.
type group[T any] struct {
	msgs      []messaging.Message[T]
	firstSeen time.Time
	lastSeen  time.Time
}

// aggregator is the Aggregator pattern implementation. It owns a single
// subscription on the source channel (registered in Start, cancelled
// in Stop), an optional background sweeper goroutine for group-timeout
// completion, and a single map+lock protecting concurrent access to
// the in-flight groups.
type aggregator[T, U any] struct {
	name           string
	src            messaging.Channel[T]
	dst            messaging.Channel[U]
	aggregate      AggregateFn[T, U]
	correlation    CorrelationFn[T]
	completion     CompletionFn[T]
	completionSize int
	groupTimeout   time.Duration
	maxGroups      int
	errorHandler   messaging.ErrorHandler
	dropHandler    DropHandler

	done         chan struct{}
	workerCancel context.CancelFunc
	startOnce    sync.Once
	stopOnce     sync.Once
	doneOnce     sync.Once
	workerWG     sync.WaitGroup

	mu     sync.Mutex
	cancel messaging.Cancel
	groups map[string]*group[T]
}
