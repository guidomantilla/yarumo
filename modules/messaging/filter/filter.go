package filter

import (
	"context"
	"fmt"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// filter is the Message Filter implementation. It owns a single
// subscription on the source channel (registered in Start, cancelled
// in Stop) and forwards each received message to the destination only
// when the configured PredicateFn returns true.
type filter[T any] struct {
	name         string
	src          messaging.Channel[T]
	dst          messaging.Channel[T]
	predicate    PredicateFn[T]
	errorHandler messaging.ErrorHandler
	dropHandler  DropHandler

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu     sync.Mutex
	cancel messaging.Cancel
}

// NewFilter constructs a Message Filter that subscribes to src and
// forwards every Message[T] for which predicate returns true to dst.
// Messages where predicate returns false are dropped (observable via
// WithDropHandler when wired).
//
// name is used in lifecycle logs and must be non-empty. src, dst and
// predicate are mandatory. Optional behaviors:
//
//   - WithErrorHandler overrides the default
//     messaging.DefaultErrorHandler (which logs via common/log) with a
//     custom hook for predicate errors/panics and forward failures.
//   - WithDropHandler installs an optional hook for observing
//     intentional drops; nil by default (silent drop).
func NewFilter[T any](name string, src messaging.Channel[T], dst messaging.Channel[T], predicate PredicateFn[T], opts ...Option) lifecycle.Component {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(dst, "destination channel is nil")
	cassert.NotNil(predicate, "predicate is nil")

	options := NewOptions(opts...)

	return &filter[T]{
		name:         name,
		src:          src,
		dst:          dst,
		predicate:    predicate,
		errorHandler: options.errorHandler,
		dropHandler:  options.dropHandler,
		done:         make(chan struct{}),
	}
}

// Name returns the filter's identity used in lifecycle logs.
func (f *filter[T]) Name() string {
	cassert.NotNil(f, "filter is nil")

	return f.name
}

// Start registers the filtering handler as a subscriber on the source
// channel. It satisfies the lifecycle.Component worker-style contract:
// Start returns immediately after the subscription is in place; the
// actual dispatching runs in the source channel's goroutine model.
// Start is idempotent — a second invocation returns nil without
// re-subscribing.
func (f *filter[T]) Start(_ context.Context) error {
	cassert.NotNil(f, "filter is nil")

	var startErr error

	f.startOnce.Do(func() {
		cancel, err := f.src.Subscribe(f.handle)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		f.mu.Lock()
		f.cancel = cancel
		f.mu.Unlock()
	})

	return startErr
}

// Stop cancels the source-channel subscription and closes Done. Stop
// is idempotent per the lifecycle.Component contract. It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when ctx
// is already expired on entry; otherwise nil.
func (f *filter[T]) Stop(ctx context.Context) error {
	cassert.NotNil(f, "filter is nil")

	f.stopOnce.Do(func() {
		f.mu.Lock()
		cancel := f.cancel
		f.cancel = nil
		f.mu.Unlock()

		if cancel != nil {
			cancel()
		}

		f.doneOnce.Do(func() { close(f.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (f *filter[T]) Done() <-chan struct{} {
	cassert.NotNil(f, "filter is nil")

	return f.done
}

// handle is the Handler[T] subscribed on the source channel. It runs
// the predicate under panic recovery, then either forwards the message
// or routes it to the DropHandler. The function itself always returns
// nil so filter concerns never propagate to the source channel's Send
// caller.
func (f *filter[T]) handle(ctx context.Context, msg messaging.Message[T]) error {
	pass, err := f.evalWithRecover(ctx, msg)
	if err != nil {
		f.reportError(ctx, msg, err)

		return nil
	}

	if !pass {
		f.reportDrop(ctx, msg)

		return nil
	}

	err = f.dst.Send(ctx, msg)
	if err != nil {
		f.reportError(ctx, msg, ErrFilter(ErrForwardFailed, err))
	}

	return nil
}

// evalWithRecover invokes the user-supplied PredicateFn under panic
// recovery. Panics become ErrFilter(ErrPredicatePanic, ...) errors;
// normal errors become ErrFilter(ErrPredicateFailed, err).
func (f *filter[T]) evalWithRecover(ctx context.Context, msg messaging.Message[T]) (pass bool, err error) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}

		pass = false
		err = ErrFilter(ErrPredicatePanic, fmt.Errorf("%v", rec))
	}()

	pass, err = f.predicate(ctx, msg)
	if err != nil {
		return false, ErrFilter(ErrPredicateFailed, err)
	}

	return pass, nil
}

// reportError forwards err to the configured ErrorHandler. ErrorHandler
// is guaranteed non-nil by NewOptions (defaults to
// messaging.DefaultErrorHandler), so the nil-guard is defensive only.
func (f *filter[T]) reportError(ctx context.Context, msg messaging.Message[T], err error) {
	if f.errorHandler == nil {
		return
	}

	f.errorHandler(ctx, msg, err)
}

// reportDrop forwards msg to the configured DropHandler. DropHandler
// is nil by default (silent drops); the guard skips invocation in that
// case.
func (f *filter[T]) reportDrop(ctx context.Context, msg messaging.Message[T]) {
	if f.dropHandler == nil {
		return
	}

	f.dropHandler(ctx, msg)
}
