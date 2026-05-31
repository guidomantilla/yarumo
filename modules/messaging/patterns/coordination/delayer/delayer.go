package delayer

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// delayer is the Delayer implementation. It owns one subscription on
// the source channel (registered in Start, cancelled in Stop) plus an
// internal messaging.ScheduledChannel[T] that holds messages until
// their deliverAt elapses, and a second subscription on that internal
// channel that forwards each due message to the real destination.
//
// pending tracks the number of messages currently in flight (scheduled
// but not yet delivered). It is incremented when the source handler
// enqueues into the internal channel and decremented when the internal
// forwarder picks the message up. WithMaxPending caps this counter; a
// new message that would exceed the bound is dropped via the
// DropHandler hook with ErrMaxPendingExceeded.
type delayer[T any] struct {
	name         string
	src          messaging.Channel[T]
	dst          messaging.Channel[T]
	fixedDelay   time.Duration
	delayFn      DelayFn[T]
	maxPending   int
	errorHandler messaging.ErrorHandler
	dropHandler  DropHandler

	internal          messaging.ScheduledChannel[T]
	internalLifecycle lifecycle.Component
	pending           atomic.Int64

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu          sync.Mutex
	srcCancel   messaging.Cancel
	innerCancel messaging.Cancel
}

// NewDelayer constructs a Delayer that subscribes to src and forwards
// every Message[T] to dst after a delay determined by the configured
// strategy. The delayer is not running on return; call lifecycle.Build
// (or Start directly) to boot the internal scheduled channel and
// register the source subscription.
//
// name is used in lifecycle logs and must be non-empty. src and dst are
// mandatory. The delay strategy is selected by options in this order:
//
//  1. WithFixedDelay — constant per-message duration.
//  2. WithDelayFn — caller computes the delay per message.
//  3. Default — uses Headers.ExpirationTime as deliver-at; a zero or
//     past ExpirationTime forwards immediately.
//
// Other optional behaviors:
//
//   - WithMaxPending bounds in-flight messages (default defaultMaxPending).
//   - WithErrorHandler / WithDropHandler observe failures and drops.
func NewDelayer[T any](name string, src messaging.Channel[T], dst messaging.Channel[T], opts ...Option[T]) Delayer[T] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(dst, "destination channel is nil")

	options := NewOptions(opts...)

	internal := messaging.NewScheduledChannel[T](name + "-internal")
	internalLifecycle, ok := internal.(lifecycle.Component)
	cassert.True(ok, "internal ScheduledChannel must implement lifecycle.Component")

	return &delayer[T]{
		name:              name,
		src:               src,
		dst:               dst,
		fixedDelay:        options.fixedDelay,
		delayFn:           options.delayFn,
		maxPending:        options.maxPending,
		errorHandler:      options.errorHandler,
		dropHandler:       options.dropHandler,
		internal:          internal,
		internalLifecycle: internalLifecycle,
		done:              make(chan struct{}),
	}
}

// Name returns the delayer's identity used in lifecycle logs.
func (d *delayer[T]) Name() string {
	cassert.NotNil(d, "delayer is nil")

	return d.name
}

// Start boots the internal ScheduledChannel, registers an internal
// subscriber that forwards due messages to the real destination, and
// registers the source-channel subscription. Start satisfies the
// lifecycle.Component worker-style contract: it returns immediately
// after wiring is in place; actual dispatching runs in the source
// channel's goroutine model plus the internal scheduler's worker. Start
// is idempotent — a second invocation returns nil without re-wiring.
func (d *delayer[T]) Start(ctx context.Context) error {
	cassert.NotNil(d, "delayer is nil")

	var startErr error

	d.startOnce.Do(func() {
		err := d.internalLifecycle.Start(ctx)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		innerCancel, err := d.internal.Subscribe(d.forward)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		srcCancel, err := d.src.Subscribe(d.handle)
		if err != nil {
			innerCancel()
			_ = d.internalLifecycle.Stop(ctx)
			startErr = lifecycle.ErrStart(err)

			return
		}

		d.mu.Lock()
		d.srcCancel = srcCancel
		d.innerCancel = innerCancel
		d.mu.Unlock()
	})

	return startErr
}

// Stop cancels the source-channel subscription, stops the internal
// ScheduledChannel (dropping any undelivered pending messages per the
// scheduled-channel best-effort semantics) and closes Done. Stop is
// idempotent per the lifecycle.Component contract.
func (d *delayer[T]) Stop(ctx context.Context) error {
	cassert.NotNil(d, "delayer is nil")

	var stopErr error

	d.stopOnce.Do(func() {
		d.mu.Lock()
		srcCancel := d.srcCancel
		innerCancel := d.innerCancel
		d.srcCancel = nil
		d.innerCancel = nil
		d.mu.Unlock()

		if srcCancel != nil {
			srcCancel()
		}

		if innerCancel != nil {
			innerCancel()
		}

		stopErr = d.internalLifecycle.Stop(ctx)

		d.doneOnce.Do(func() { close(d.done) })
	})

	if stopErr != nil {
		return stopErr
	}

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (d *delayer[T]) Done() <-chan struct{} {
	cassert.NotNil(d, "delayer is nil")

	return d.done
}

// handle is the Handler[T] subscribed on the source channel. It
// computes the per-message delay, enforces the WithMaxPending bound and
// either forwards immediately (delay <= 0) or schedules deferred
// delivery on the internal scheduled channel. The function always
// returns nil so delayer concerns never propagate to the source
// channel's Send caller.
func (d *delayer[T]) handle(ctx context.Context, msg messaging.Message[T]) error {
	delay := d.computeDelay(ctx, msg)

	if delay <= 0 {
		err := d.dst.Send(ctx, msg)
		if err != nil {
			d.reportError(ctx, msg, ErrDelayer(ErrForwardFailed, err))
		}

		return nil
	}

	current := d.pending.Add(1)
	if current > int64(d.maxPending) {
		d.pending.Add(-1)
		d.reportDrop(ctx, msg)

		return nil
	}

	err := d.internal.SendAfter(ctx, delay, msg)
	if err != nil {
		d.pending.Add(-1)
		d.reportError(ctx, msg, ErrDelayer(ErrScheduleFailed, err))
	}

	return nil
}

// forward is the Handler[T] subscribed on the internal scheduled
// channel. It decrements the pending counter and forwards the now-due
// message to the real destination. Forward failures are reported via
// the configured ErrorHandler.
func (d *delayer[T]) forward(ctx context.Context, msg messaging.Message[T]) error {
	d.pending.Add(-1)

	err := d.dst.Send(ctx, msg)
	if err != nil {
		d.reportError(ctx, msg, ErrDelayer(ErrForwardFailed, err))
	}

	return nil
}

// computeDelay selects the configured delay strategy and returns the
// duration to wait before forwarding msg. Order of precedence:
//
//  1. WithFixedDelay (if > 0).
//  2. WithDelayFn (if set).
//  3. Headers.ExpirationTime fallback — delivers when the deadline
//     elapses; zero/past values forward immediately.
func (d *delayer[T]) computeDelay(ctx context.Context, msg messaging.Message[T]) time.Duration {
	if d.fixedDelay > 0 {
		return d.fixedDelay
	}

	if d.delayFn != nil {
		return d.delayFn(ctx, msg)
	}

	if msg.Headers.ExpirationTime.IsZero() {
		return 0
	}

	return time.Until(msg.Headers.ExpirationTime)
}

// reportError forwards err to the configured ErrorHandler. ErrorHandler
// is guaranteed non-nil by NewOptions (defaults to
// messaging.DefaultErrorHandler), so the nil-guard is defensive only.
func (d *delayer[T]) reportError(ctx context.Context, msg messaging.Message[T], err error) {
	if d.errorHandler == nil {
		return
	}

	d.errorHandler(ctx, msg, err)
}

// reportDrop invokes the configured DropHandler with
// ErrMaxPendingExceeded. DropHandler is nil by default (silent drop);
// the guard skips invocation in that case.
func (d *delayer[T]) reportDrop(ctx context.Context, msg messaging.Message[T]) {
	if d.dropHandler == nil {
		return
	}

	d.dropHandler(ctx, msg, ErrDelayer(ErrMaxPendingExceeded))
}
