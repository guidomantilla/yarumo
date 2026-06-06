package barrier

import (
	"context"
	"sync"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// barrierGroup holds the per-CorrelationID accumulation state. msgs
// accumulates in arrival order so the release fan-out preserves the
// upstream sequence. firstSeen and lastSeen drive the sweeper's
// timeout decision (firstSeen for absolute deadlines, lastSeen for
// inactivity-based eviction — the sweeper uses lastSeen so a still-
// active group is not killed mid-stream).
type barrierGroup[T any] struct {
	msgs      []messaging.Message[T]
	firstSeen time.Time
	lastSeen  time.Time
}

// barrier is the Barrier implementation. It owns a single
// subscription on the source channel (registered in Start, cancelled
// in Stop), accumulates messages per Headers.CorrelationID, and
// releases each group as-is when the configured quorum is reached.
// A background sweeper goroutine evicts groups that exceed
// WithGroupTimeout without reaching quorum.
type barrier[T any] struct {
	name          string
	src           messaging.Channel[T]
	dst           messaging.Channel[T]
	quorum        int
	groupTimeout  time.Duration
	maxGroups     int
	sweepInterval time.Duration
	errorHandler  messaging.ErrorHandler
	dropHandler   DropHandler

	groupsMu sync.Mutex
	groups   map[string]*barrierGroup[T]

	sweeperWG sync.WaitGroup

	done         chan struct{}
	sweeperDone  chan struct{}
	sweeperOnce  sync.Once
	startOnce    sync.Once
	stopOnce     sync.Once
	doneOnce     sync.Once

	subMu     sync.Mutex
	subCancel messaging.Cancel
}

// NewBarrier constructs a Barrier that subscribes to src and releases
// quorum-sized groups (correlated by Headers.CorrelationID) to dst.
// The Barrier is not running on return; call lifecycle.Build (or
// Start directly) to register the subscription and spawn the sweeper.
//
// name is used in lifecycle logs and must be non-empty. src and dst
// are mandatory; quorum must be positive. WithGroupTimeout is REQUIRED
// (a positive duration) to bound memory — the constructor asserts on
// it via cassert.
//
// Optional behaviors:
//
//   - WithMaxGroups caps the number of distinct correlations tracked
//     at any one time (default DefaultMaxGroups).
//   - WithSweepInterval tunes the sweeper cadence (default
//     DefaultSweepInterval).
//   - WithErrorHandler overrides the default
//     messaging.DefaultErrorHandler (which logs via common/log) with a
//     custom hook for forward Send failures during release.
//   - WithDropHandler installs an optional hook for observing
//     intentional drops; nil by default (silent drop).
func NewBarrier[T any](name string, src messaging.Channel[T], dst messaging.Channel[T], quorum int, opts ...Option) Barrier[T] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(dst, "destination channel is nil")
	cassert.True(quorum > 0, "quorum must be positive")

	options := NewOptions(opts...)
	cassert.True(options.groupTimeout > 0, "WithGroupTimeout is required (positive duration)")

	return &barrier[T]{
		name:          name,
		src:           src,
		dst:           dst,
		quorum:        quorum,
		groupTimeout:  options.groupTimeout,
		maxGroups:     options.maxGroups,
		sweepInterval: options.sweepInterval,
		errorHandler:  options.errorHandler,
		dropHandler:   options.dropHandler,
		groups:        map[string]*barrierGroup[T]{},
		done:          make(chan struct{}),
		sweeperDone:   make(chan struct{}),
	}
}

// Name returns the barrier's identity used in lifecycle logs.
func (b *barrier[T]) Name() string {
	cassert.NotNil(b, "barrier is nil")

	return b.name
}

// Start registers the accumulating handler as a subscriber on the
// source channel and spawns the timeout sweeper goroutine. It
// satisfies the lifecycle.Component worker-style contract: Start
// returns immediately after the subscription is in place. Start is
// idempotent — a second invocation returns nil without re-subscribing.
func (b *barrier[T]) Start(_ context.Context) error {
	cassert.NotNil(b, "barrier is nil")

	var startErr error

	b.startOnce.Do(func() {
		cancel, err := b.src.Subscribe(b.handle)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		b.subMu.Lock()
		b.subCancel = cancel
		b.subMu.Unlock()

		b.sweeperWG.Go(b.sweep)
	})

	return startErr
}

// Stop cancels the source-channel subscription, stops the sweeper,
// drains every still-incomplete group through WithDropHandler, and
// closes Done. Stop is idempotent per the lifecycle.Component
// contract. It returns lifecycle.ErrShutdown wrapping
// lifecycle.ErrShutdownTimeout when ctx expires before the sweeper
// goroutine has exited.
func (b *barrier[T]) Stop(ctx context.Context) error {
	cassert.NotNil(b, "barrier is nil")

	b.stopOnce.Do(func() {
		b.subMu.Lock()
		cancel := b.subCancel
		b.subCancel = nil
		b.subMu.Unlock()

		if cancel != nil {
			cancel()
		}

		b.sweeperOnce.Do(func() { close(b.sweeperDone) })

		b.sweeperWG.Wait()

		b.drainAll(ctx)

		b.doneOnce.Do(func() { close(b.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has drained
// every in-flight group.
func (b *barrier[T]) Done() <-chan struct{} {
	cassert.NotNil(b, "barrier is nil")

	return b.done
}

// handle is the Handler[T] subscribed on the source channel. It
// accumulates msg into its CorrelationID group and, on quorum,
// releases the whole group to dst in arrival order. Failures and
// drops flow through the configured hooks; the function itself
// always returns nil so barrier concerns never propagate to the
// source channel's Send caller.
func (b *barrier[T]) handle(ctx context.Context, msg messaging.Message[T]) error {
	correlation := msg.Headers.CorrelationID
	if correlation == "" {
		b.reportDrop(ctx, msg)

		return nil
	}

	released, dropped := b.appendAndMaybeRelease(msg, correlation)
	if dropped {
		b.reportDrop(ctx, msg)

		return nil
	}

	for _, m := range released {
		err := b.dst.Send(ctx, m)
		if err != nil {
			b.reportError(ctx, m, ErrBarrier(ErrForwardFailed, err))
		}
	}

	return nil
}

// appendAndMaybeRelease appends msg to its correlation group and, if
// the quorum is reached, removes and returns the accumulated slice
// for release. dropped is true when the message could not be accepted
// (currently only the WithMaxGroups cap); in that case released is
// nil and the caller must fire the drop hook outside the lock.
// released may be nil with dropped=false when the group is not yet
// at quorum. All map mutation happens under groupsMu.
func (b *barrier[T]) appendAndMaybeRelease(msg messaging.Message[T], correlation string) (released []messaging.Message[T], dropped bool) {
	b.groupsMu.Lock()
	defer b.groupsMu.Unlock()

	now := time.Now()

	group, exists := b.groups[correlation]
	if !exists {
		if len(b.groups) >= b.maxGroups {
			return nil, true
		}

		group = &barrierGroup[T]{
			msgs:      []messaging.Message[T]{},
			firstSeen: now,
		}
		b.groups[correlation] = group
	}

	group.msgs = append(group.msgs, msg)
	group.lastSeen = now

	if len(group.msgs) < b.quorum {
		return nil, false
	}

	delete(b.groups, correlation)

	return group.msgs, false
}

// reportError forwards err to the configured ErrorHandler.
// ErrorHandler is guaranteed non-nil by NewOptions, so the nil-guard
// is defensive only.
func (b *barrier[T]) reportError(ctx context.Context, msg messaging.Message[T], err error) {
	if b.errorHandler == nil {
		return
	}

	b.errorHandler(ctx, msg, err)
}

// reportDrop forwards msg to the configured DropHandler. DropHandler
// is nil by default (silent drops); the guard skips invocation in
// that case.
func (b *barrier[T]) reportDrop(ctx context.Context, msg messaging.Message[T]) {
	if b.dropHandler == nil {
		return
	}

	b.dropHandler(ctx, msg)
}

// sweep is the timeout sweeper goroutine. It wakes on sweepInterval
// and evicts any group whose lastSeen is older than groupTimeout,
// firing the drop hook for every still-accumulated message. The
// goroutine exits when sweeperDone is closed (Stop).
func (b *barrier[T]) sweep() {
	ticker := time.NewTicker(b.sweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.sweeperDone:
			return
		case <-ticker.C:
			b.sweepOnce()
		}
	}
}

// sweepOnce performs one eviction pass. Groups are collected under
// groupsMu but the drop hook is invoked outside the lock so a slow
// hook does not block incoming Send.
func (b *barrier[T]) sweepOnce() {
	now := time.Now()

	type evicted struct {
		correlation string
		msgs        []messaging.Message[T]
	}

	var doomed []evicted

	b.groupsMu.Lock()
	for correlation, group := range b.groups {
		if now.Sub(group.lastSeen) < b.groupTimeout {
			continue
		}

		doomed = append(doomed, evicted{correlation: correlation, msgs: group.msgs})
		delete(b.groups, correlation)
	}
	b.groupsMu.Unlock()

	for _, e := range doomed {
		for _, m := range e.msgs {
			b.reportDrop(context.Background(), m)
		}
	}
}

// drainAll evicts every in-flight group (Stop). All messages still
// accumulated fire the drop hook. drainAll is called from Stop after
// the sweeper has exited, so no race against concurrent sweepOnce.
func (b *barrier[T]) drainAll(ctx context.Context) {
	b.groupsMu.Lock()
	pending := b.groups
	b.groups = map[string]*barrierGroup[T]{}
	b.groupsMu.Unlock()

	for _, group := range pending {
		for _, m := range group.msgs {
			b.reportDrop(ctx, m)
		}
	}
}
