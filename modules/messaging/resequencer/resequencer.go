package resequencer

import (
	"context"
	"slices"
	"sync"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// seqGroup holds the per-CorrelationID buffering state for a
// sequence. msgs is a sparse map seqNumber → message; expected is the
// total sequence size (captured from the first arriving message's
// Headers.SequenceSize); nextEmit is the cursor of the next position
// to forward; firstSeen and lastSeen drive the sweeper's eviction
// decision (lastSeen so an actively-arriving sequence is not killed
// mid-stream).
type seqGroup[T any] struct {
	msgs      map[int]messaging.Message[T]
	expected  int
	nextEmit  int
	firstSeen time.Time
	lastSeen  time.Time
}

// resequencer is the Resequencer implementation. It owns a single
// subscription on the source channel (registered in Start, cancelled
// in Stop), buffers out-of-order messages per Headers.CorrelationID,
// and emits them in SequenceNumber order. A background sweeper
// goroutine evicts groups whose missing position never arrives.
type resequencer[T any] struct {
	name          string
	src           messaging.Channel[T]
	dst           messaging.Channel[T]
	groupTimeout  time.Duration
	maxGroups     int
	sweepInterval time.Duration
	errorHandler  messaging.ErrorHandler
	dropHandler   DropHandler

	groupsMu sync.Mutex
	groups   map[string]*seqGroup[T]

	sweeperWG sync.WaitGroup

	done        chan struct{}
	sweeperDone chan struct{}
	sweeperOnce sync.Once
	startOnce   sync.Once
	stopOnce    sync.Once
	doneOnce    sync.Once

	subMu     sync.Mutex
	subCancel messaging.Cancel
}

// NewResequencer constructs a Resequencer that subscribes to src,
// buffers out-of-order messages per Headers.CorrelationID, and emits
// them to dst in SequenceNumber order. The Resequencer is not running
// on return; call lifecycle.Build (or Start directly) to register the
// subscription and spawn the sweeper.
//
// name is used in lifecycle logs and must be non-empty. src and dst
// are mandatory. WithGroupTimeout is REQUIRED (a positive duration)
// to bound memory — the constructor asserts on it via cassert.
//
// Optional behaviors:
//
//   - WithMaxGroups caps the number of distinct correlations tracked
//     at any one time (default DefaultMaxGroups).
//   - WithSweepInterval tunes the sweeper cadence (default
//     DefaultSweepInterval).
//   - WithErrorHandler overrides the default
//     messaging.DefaultErrorHandler (which logs via common/log) with a
//     custom hook for forward Send failures during emit.
//   - WithDropHandler installs an optional hook for observing
//     intentional drops; nil by default (silent drop).
func NewResequencer[T any](name string, src messaging.Channel[T], dst messaging.Channel[T], opts ...Option) Resequencer[T] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(dst, "destination channel is nil")

	options := NewOptions(opts...)
	cassert.True(options.groupTimeout > 0, "WithGroupTimeout is required (positive duration)")

	return &resequencer[T]{
		name:          name,
		src:           src,
		dst:           dst,
		groupTimeout:  options.groupTimeout,
		maxGroups:     options.maxGroups,
		sweepInterval: options.sweepInterval,
		errorHandler:  options.errorHandler,
		dropHandler:   options.dropHandler,
		groups:        map[string]*seqGroup[T]{},
		done:          make(chan struct{}),
		sweeperDone:   make(chan struct{}),
	}
}

// Name returns the resequencer's identity used in lifecycle logs.
func (r *resequencer[T]) Name() string {
	cassert.NotNil(r, "resequencer is nil")

	return r.name
}

// Start registers the buffering handler as a subscriber on the source
// channel and spawns the timeout sweeper goroutine. It satisfies the
// lifecycle.Component worker-style contract: Start returns immediately
// after the subscription is in place. Start is idempotent — a second
// invocation returns nil without re-subscribing.
func (r *resequencer[T]) Start(_ context.Context) error {
	cassert.NotNil(r, "resequencer is nil")

	var startErr error

	r.startOnce.Do(func() {
		cancel, err := r.src.Subscribe(r.handle)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		r.subMu.Lock()
		r.subCancel = cancel
		r.subMu.Unlock()

		r.sweeperWG.Go(r.sweep)
	})

	return startErr
}

// Stop cancels the source-channel subscription, stops the sweeper,
// drains every still-buffered message through WithDropHandler, and
// closes Done. Stop is idempotent per the lifecycle.Component
// contract. It returns lifecycle.ErrShutdown wrapping
// lifecycle.ErrShutdownTimeout when ctx expires before the sweeper
// goroutine has exited.
func (r *resequencer[T]) Stop(ctx context.Context) error {
	cassert.NotNil(r, "resequencer is nil")

	r.stopOnce.Do(func() {
		r.subMu.Lock()
		cancel := r.subCancel
		r.subCancel = nil
		r.subMu.Unlock()

		if cancel != nil {
			cancel()
		}

		r.sweeperOnce.Do(func() { close(r.sweeperDone) })

		r.sweeperWG.Wait()

		r.drainAll(ctx)

		r.doneOnce.Do(func() { close(r.done) })
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
func (r *resequencer[T]) Done() <-chan struct{} {
	cassert.NotNil(r, "resequencer is nil")

	return r.done
}

// handle is the Handler[T] subscribed on the source channel. It
// validates the message's sequence metadata, buffers it into its
// correlation group, drains as many consecutive positions as it can,
// and emits them to dst in order. Failures and drops flow through the
// configured hooks; the function itself always returns nil so
// resequencer concerns never propagate to the source channel's Send
// caller.
func (r *resequencer[T]) handle(ctx context.Context, msg messaging.Message[T]) error {
	correlation := msg.Headers.CorrelationID
	if correlation == "" {
		r.reportDrop(ctx, msg)

		return nil
	}

	size := msg.Headers.SequenceSize
	if size <= 0 {
		r.reportDrop(ctx, msg)

		return nil
	}

	seqNumber := msg.Headers.SequenceNumber
	if seqNumber < 0 || seqNumber >= size {
		r.reportDrop(ctx, msg)

		return nil
	}

	emit, dropped := r.bufferAndDrain(msg, correlation, size, seqNumber)
	if dropped {
		r.reportDrop(ctx, msg)

		return nil
	}

	for _, m := range emit {
		err := r.dst.Send(ctx, m)
		if err != nil {
			r.reportError(ctx, m, ErrResequencer(ErrForwardFailed, err))
		}
	}

	return nil
}

// bufferAndDrain stores msg in its correlation group at the given
// seqNumber and drains as many consecutive positions starting from
// nextEmit as possible, returning the contiguous slice to be emitted
// in order. dropped is true when the message must be dropped
// (MaxGroups cap, SequenceSize mismatch, duplicate position); in that
// case emit is nil and the caller must fire the drop hook outside the
// lock. When the group's cursor reaches expected, the group is removed.
// All map mutation happens under groupsMu.
func (r *resequencer[T]) bufferAndDrain(msg messaging.Message[T], correlation string, size, seqNumber int) (emit []messaging.Message[T], dropped bool) {
	r.groupsMu.Lock()
	defer r.groupsMu.Unlock()

	now := time.Now()

	group, exists := r.groups[correlation]
	if !exists {
		if len(r.groups) >= r.maxGroups {
			return nil, true
		}

		group = &seqGroup[T]{
			msgs:      map[int]messaging.Message[T]{},
			expected:  size,
			nextEmit:  0,
			firstSeen: now,
		}
		r.groups[correlation] = group
	}

	if group.expected != size {
		return nil, true
	}

	if seqNumber < group.nextEmit {
		return nil, true
	}

	_, dup := group.msgs[seqNumber]
	if dup {
		return nil, true
	}

	group.msgs[seqNumber] = msg
	group.lastSeen = now

	for {
		next, ok := group.msgs[group.nextEmit]
		if !ok {
			break
		}

		emit = append(emit, next)
		delete(group.msgs, group.nextEmit)
		group.nextEmit++
	}

	if group.nextEmit >= group.expected {
		delete(r.groups, correlation)
	}

	return emit, false
}

// reportError forwards err to the configured ErrorHandler.
// ErrorHandler is guaranteed non-nil by NewOptions, so the nil-guard
// is defensive only.
func (r *resequencer[T]) reportError(ctx context.Context, msg messaging.Message[T], err error) {
	if r.errorHandler == nil {
		return
	}

	r.errorHandler(ctx, msg, err)
}

// reportDrop forwards msg to the configured DropHandler. DropHandler
// is nil by default (silent drops); the guard skips invocation in
// that case.
func (r *resequencer[T]) reportDrop(ctx context.Context, msg messaging.Message[T]) {
	if r.dropHandler == nil {
		return
	}

	r.dropHandler(ctx, msg)
}

// sweep is the timeout sweeper goroutine. It wakes on sweepInterval
// and evicts any group whose lastSeen is older than groupTimeout,
// firing the drop hook for every still-buffered (unforwarded)
// message. The goroutine exits when sweeperDone is closed (Stop).
func (r *resequencer[T]) sweep() {
	ticker := time.NewTicker(r.sweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.sweeperDone:
			return
		case <-ticker.C:
			r.sweepOnce()
		}
	}
}

// sweepOnce performs one eviction pass. Groups are collected under
// groupsMu but the drop hook is invoked outside the lock so a slow
// hook does not block incoming Send.
func (r *resequencer[T]) sweepOnce() {
	now := time.Now()

	type evicted struct {
		msgs []messaging.Message[T]
	}

	var doomed []evicted

	r.groupsMu.Lock()
	for correlation, group := range r.groups {
		if now.Sub(group.lastSeen) < r.groupTimeout {
			continue
		}

		doomed = append(doomed, evicted{msgs: collectGroupMessages(group)})
		delete(r.groups, correlation)
	}
	r.groupsMu.Unlock()

	for _, e := range doomed {
		for _, m := range e.msgs {
			r.reportDrop(context.Background(), m)
		}
	}
}

// drainAll evicts every in-flight group (Stop). All still-buffered
// messages fire the drop hook. drainAll is called from Stop after
// the sweeper has exited, so no race against concurrent sweepOnce.
func (r *resequencer[T]) drainAll(ctx context.Context) {
	r.groupsMu.Lock()
	pending := r.groups
	r.groups = map[string]*seqGroup[T]{}
	r.groupsMu.Unlock()

	for _, group := range pending {
		for _, m := range collectGroupMessages(group) {
			r.reportDrop(ctx, m)
		}
	}
}

// collectGroupMessages returns the still-buffered messages of a
// group, sorted by their stored SequenceNumber. The slice contains
// every map entry; positions already emitted are not present (they
// were removed during bufferAndDrain). Sorting yields a stable order
// for the drop hook (helps deterministic tests and audit trails).
func collectGroupMessages[T any](group *seqGroup[T]) []messaging.Message[T] {
	if len(group.msgs) == 0 {
		return nil
	}

	keys := make([]int, 0, len(group.msgs))
	for k := range group.msgs {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	out := make([]messaging.Message[T], 0, len(keys))
	for _, k := range keys {
		out = append(out, group.msgs[k])
	}

	return out
}
