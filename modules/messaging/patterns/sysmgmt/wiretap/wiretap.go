package wiretap

import (
	"context"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// wiretap is the Wire Tap implementation. It owns a single subscription
// on the source channel (registered in Start, cancelled in Stop) and
// forwards each received message to both the primary destination and
// the side-channel tap.
type wiretap[T any] struct {
	name         string
	src          messaging.Channel[T]
	dst          messaging.Channel[T]
	tap          messaging.Channel[T]
	errorHandler messaging.ErrorHandler

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu     sync.Mutex
	cancel messaging.Cancel
}

// NewWiretap constructs a Wire Tap that subscribes to src and forwards
// every Message[T] to both dst (primary flow) and tap (observability
// sink). The wiretap is not running on return; call lifecycle.Build (or
// Start directly) to register the subscription.
//
// name is used in lifecycle logs and must be non-empty. src, dst and
// tap are mandatory. The only optional behavior is the ErrorHandler
// hook installed via WithErrorHandler (defaulting to
// messaging.DefaultErrorHandler which logs via common/log).
//
// Order of operations: primary dst Send happens first; tap Send happens
// second regardless of whether the primary succeeded. Tap failures
// NEVER alter the primary flow — they are only reported through the
// ErrorHandler.
func NewWiretap[T any](name string, src messaging.Channel[T], dst messaging.Channel[T], tap messaging.Channel[T], opts ...Option) Wiretap[T] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(dst, "destination channel is nil")
	cassert.NotNil(tap, "tap channel is nil")

	options := NewOptions(opts...)

	return &wiretap[T]{
		name:         name,
		src:          src,
		dst:          dst,
		tap:          tap,
		errorHandler: options.errorHandler,
		done:         make(chan struct{}),
	}
}

// Name returns the wiretap's identity used in lifecycle logs.
func (w *wiretap[T]) Name() string {
	cassert.NotNil(w, "wiretap is nil")

	return w.name
}

// Start registers the wiretap handler as a subscriber on the source
// channel. It satisfies the lifecycle.Component worker-style contract:
// Start returns immediately after the subscription is in place; the
// actual dispatching runs in the source channel's goroutine model.
// Start is idempotent — a second invocation returns nil without
// re-subscribing.
func (w *wiretap[T]) Start(_ context.Context) error {
	cassert.NotNil(w, "wiretap is nil")

	var startErr error

	w.startOnce.Do(func() {
		cancel, err := w.src.Subscribe(w.handle)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		w.mu.Lock()
		w.cancel = cancel
		w.mu.Unlock()
	})

	return startErr
}

// Stop cancels the source-channel subscription and closes Done. Stop
// is idempotent per the lifecycle.Component contract. It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when ctx
// is already expired on entry; otherwise nil.
func (w *wiretap[T]) Stop(ctx context.Context) error {
	cassert.NotNil(w, "wiretap is nil")

	w.stopOnce.Do(func() {
		w.mu.Lock()
		cancel := w.cancel
		w.cancel = nil
		w.mu.Unlock()

		if cancel != nil {
			cancel()
		}

		w.doneOnce.Do(func() { close(w.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (w *wiretap[T]) Done() <-chan struct{} {
	cassert.NotNil(w, "wiretap is nil")

	return w.done
}

// handle is the Handler[T] subscribed on the source channel. It sends
// msg to the primary destination first, then to the tap regardless of
// the primary outcome. Both failures are reported through the
// configured ErrorHandler; tap failures never affect the primary flow.
// The function itself always returns nil so wiretap concerns never
// propagate to the source channel's Send caller.
func (w *wiretap[T]) handle(ctx context.Context, msg messaging.Message[T]) error {
	dstErr := w.dst.Send(ctx, msg)
	if dstErr != nil {
		w.report(ctx, msg, ErrWiretap(ErrForwardFailed, dstErr))
	}

	tapErr := w.tap.Send(ctx, msg)
	if tapErr != nil {
		w.report(ctx, msg, ErrWiretap(ErrTapSendFailed, tapErr))
	}

	return nil
}

// report forwards err to the configured ErrorHandler. ErrorHandler is
// guaranteed non-nil by NewOptions (defaults to
// messaging.DefaultErrorHandler), so the nil-guard is defensive only.
func (w *wiretap[T]) report(ctx context.Context, msg messaging.Message[T], err error) {
	if w.errorHandler == nil {
		return
	}

	w.errorHandler(ctx, msg, err)
}
