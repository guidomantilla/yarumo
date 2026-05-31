package recipientlist

import (
	"context"
	"fmt"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// recipientList is the Recipient List implementation. It owns a single
// subscription on the source channel (registered in Start, cancelled in
// Stop) and forwards each received message to every destination channel
// resolved from the keys returned by SelectorFn.
type recipientList[T any] struct {
	name         string
	src          messaging.Channel[T]
	selector     SelectorFn[T]
	routes       map[string]messaging.Channel[T]
	errorHandler messaging.ErrorHandler
	dropHandler  DropHandler

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu     sync.Mutex
	cancel messaging.Cancel
}

// NewRecipientList constructs a Recipient List that subscribes to src and
// forwards each Message[T] to every routes[k] for k in selector(msg).
// The recipient list is not running on return; call lifecycle.Build (or
// Start directly) to register the subscription.
//
// name is used in lifecycle logs and must be non-empty. src, selector
// and a non-empty routes map are mandatory. Optional behaviors:
//
//   - WithErrorHandler overrides the default
//     messaging.DefaultErrorHandler (which logs via common/log) with a
//     custom hook for per-recipient errors (missing key, forward fail,
//     selector error/panic).
//   - WithDropHandler installs an optional hook for observing
//     intentional drops (SelectorFn returned an empty slice); nil by
//     default (silent drop).
func NewRecipientList[T any](name string, src messaging.Channel[T], selector SelectorFn[T], routes map[string]messaging.Channel[T], opts ...Option) RecipientList[T] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(selector, "selector function is nil")
	cassert.NotNil(routes, "routes map is nil")

	options := NewOptions(opts...)

	return &recipientList[T]{
		name:         name,
		src:          src,
		selector:     selector,
		routes:       routes,
		errorHandler: options.errorHandler,
		dropHandler:  options.dropHandler,
		done:         make(chan struct{}),
	}
}

// Name returns the recipient list's identity used in lifecycle logs.
func (r *recipientList[T]) Name() string {
	cassert.NotNil(r, "recipient list is nil")

	return r.name
}

// Start registers the dispatching handler as a subscriber on the source
// channel. It satisfies the lifecycle.Component worker-style contract:
// Start returns immediately after the subscription is in place; the
// actual dispatching runs in the source channel's goroutine model.
// Start is idempotent — a second invocation returns nil without
// re-subscribing.
func (r *recipientList[T]) Start(_ context.Context) error {
	cassert.NotNil(r, "recipient list is nil")

	var startErr error

	r.startOnce.Do(func() {
		cancel, err := r.src.Subscribe(r.handle)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		r.mu.Lock()
		r.cancel = cancel
		r.mu.Unlock()
	})

	return startErr
}

// Stop cancels the source-channel subscription and closes Done. Stop
// is idempotent per the lifecycle.Component contract. It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when ctx
// is already expired on entry; otherwise nil.
func (r *recipientList[T]) Stop(ctx context.Context) error {
	cassert.NotNil(r, "recipient list is nil")

	r.stopOnce.Do(func() {
		r.mu.Lock()
		cancel := r.cancel
		r.cancel = nil
		r.mu.Unlock()

		if cancel != nil {
			cancel()
		}

		r.doneOnce.Do(func() { close(r.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (r *recipientList[T]) Done() <-chan struct{} {
	cassert.NotNil(r, "recipient list is nil")

	return r.done
}

// handle is the Handler[T] subscribed on the source channel. It runs
// SelectorFn under panic recovery, then forwards the message to every
// resolved recipient. Per-recipient errors are reported through the
// configured ErrorHandler individually so a single failure does not
// abort delivery to the others. The function itself always returns nil
// so recipient list concerns never propagate to the source channel's
// Send caller.
func (r *recipientList[T]) handle(ctx context.Context, msg messaging.Message[T]) error {
	keys, err := r.selectWithRecover(ctx, msg)
	if err != nil {
		r.reportError(ctx, msg, err)

		return nil
	}

	if len(keys) == 0 {
		r.reportDrop(ctx, msg)

		return nil
	}

	for _, key := range keys {
		r.dispatchOne(ctx, msg, key)
	}

	return nil
}

// dispatchOne resolves a single key and forwards msg to it. Missing
// keys and forward failures are reported individually to the configured
// ErrorHandler.
func (r *recipientList[T]) dispatchOne(ctx context.Context, msg messaging.Message[T], key string) {
	dst, ok := r.routes[key]
	if !ok {
		r.reportError(ctx, msg, ErrRecipientList(ErrNoRoute, fmt.Errorf("key=%q", key)))

		return
	}

	err := dst.Send(ctx, msg)
	if err != nil {
		r.reportError(ctx, msg, ErrRecipientList(ErrForwardFailed, fmt.Errorf("key=%q", key), err))
	}
}

// selectWithRecover invokes the user-supplied SelectorFn under panic
// recovery. Panics become ErrRecipientList(ErrSelectorPanic, ...)
// errors; normal errors become ErrRecipientList(ErrSelectorFnFailed,
// err).
func (r *recipientList[T]) selectWithRecover(ctx context.Context, msg messaging.Message[T]) (keys []string, err error) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}

		keys = nil
		err = ErrRecipientList(ErrSelectorPanic, fmt.Errorf("%v", rec))
	}()

	keys, err = r.selector(ctx, msg)
	if err != nil {
		return nil, ErrRecipientList(ErrSelectorFnFailed, err)
	}

	return keys, nil
}

// reportError forwards err to the configured ErrorHandler. ErrorHandler
// is guaranteed non-nil by NewOptions (defaults to
// messaging.DefaultErrorHandler), so the nil-guard is defensive only.
func (r *recipientList[T]) reportError(ctx context.Context, msg messaging.Message[T], err error) {
	if r.errorHandler == nil {
		return
	}

	r.errorHandler(ctx, msg, err)
}

// reportDrop forwards msg to the configured DropHandler. DropHandler is
// nil by default (silent drops); the guard skips invocation in that
// case.
func (r *recipientList[T]) reportDrop(ctx context.Context, msg messaging.Message[T]) {
	if r.dropHandler == nil {
		return
	}

	r.dropHandler(ctx, msg)
}
