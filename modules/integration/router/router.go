package router

import (
	"context"
	"fmt"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/core/common/messaging"
)

// router is the Content-Based Router implementation. It owns a single
// subscription on the source channel (registered in Start, cancelled in
// Stop) and forwards each received message to one of the configured
// destination channels based on the RouteFn decision.
type router[T any] struct {
	name           string
	src            messaging.Channel[T]
	decide         RouteFn[T]
	routes         map[string]messaging.Channel[T]
	defaultChannel messaging.Channel[T]
	errorHandler   messaging.ErrorHandler

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu     sync.Mutex
	cancel messaging.Cancel
}

// NewRouter constructs a Content-Based Router that subscribes to src and
// forwards each Message[T] to routes[decide(msg)]. The router is not
// running on return; call lifecycle.Build (or Start directly) to
// register the subscription.
//
// name is used in lifecycle logs and must be non-empty. src, decide and
// a non-empty routes map are mandatory. Optional behaviors:
//
//   - WithDefaultChannel installs a fallback destination for NoRoute
//     messages; without it, NoRoute messages are dropped and reported
//     to the ErrorHandler.
//   - WithErrorHandler overrides the default messaging.DefaultErrorHandler
//     (which logs via common/log) with a custom hook or
//     messaging.SilentErrorHandler.
func NewRouter[T any](name string, src messaging.Channel[T], decide RouteFn[T], routes map[string]messaging.Channel[T], opts ...Option[T]) lifecycle.Component {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(decide, "route function is nil")
	cassert.NotNil(routes, "routes map is nil")

	options := NewOptions(opts...)

	return &router[T]{
		name:           name,
		src:            src,
		decide:         decide,
		routes:         routes,
		defaultChannel: options.defaultChannel,
		errorHandler:   options.errorHandler,
		done:           make(chan struct{}),
	}
}

// Name returns the router's identity used in lifecycle logs.
func (r *router[T]) Name() string {
	cassert.NotNil(r, "router is nil")

	return r.name
}

// Start registers the routing handler as a subscriber on the source
// channel. It satisfies the lifecycle.Component worker-style contract:
// Start returns immediately after the subscription is in place; the
// actual dispatching runs in the source channel's goroutine model.
// Start is idempotent — a second invocation returns nil without
// re-subscribing.
func (r *router[T]) Start(_ context.Context) error {
	cassert.NotNil(r, "router is nil")

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
func (r *router[T]) Stop(ctx context.Context) error {
	cassert.NotNil(r, "router is nil")

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
func (r *router[T]) Done() <-chan struct{} {
	cassert.NotNil(r, "router is nil")

	return r.done
}

// handle is the Handler[T] subscribed on the source channel. It runs
// RouteFn under panic recovery, resolves the destination, and forwards
// the message. Routing failures are reported through the configured
// ErrorHandler; the function itself always returns nil so routing
// concerns never propagate to the source channel's Send caller.
func (r *router[T]) handle(ctx context.Context, msg messaging.Message[T]) error {
	key, err := r.decideWithRecover(ctx, msg)
	if err != nil {
		r.report(ctx, msg, err)

		return nil
	}

	dst, ok := r.routes[key]
	if !ok {
		r.handleNoRoute(ctx, msg, key)

		return nil
	}

	err = dst.Send(ctx, msg)
	if err != nil {
		r.report(ctx, msg, ErrRoute(ErrForwardFailed, err))
	}

	return nil
}

// decideWithRecover invokes the user-supplied RouteFn under panic
// recovery. Panics become ErrRoute(ErrRoutePanic, ...) errors; normal
// errors become ErrRoute(ErrRouteFnFailed, err).
func (r *router[T]) decideWithRecover(ctx context.Context, msg messaging.Message[T]) (key string, err error) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}

		err = ErrRoute(ErrRoutePanic, fmt.Errorf("%v", rec))
	}()

	key, err = r.decide(ctx, msg)
	if err != nil {
		return "", ErrRoute(ErrRouteFnFailed, err)
	}

	return key, nil
}

// handleNoRoute applies the NoRoute policy: forward to the default
// channel when configured, else report ErrNoRoute via the ErrorHandler.
// A failure to send to the default channel is itself reported as
// ErrForwardFailed.
func (r *router[T]) handleNoRoute(ctx context.Context, msg messaging.Message[T], key string) {
	if r.defaultChannel != nil {
		err := r.defaultChannel.Send(ctx, msg)
		if err != nil {
			r.report(ctx, msg, ErrRoute(ErrForwardFailed, err))
		}

		return
	}

	r.report(ctx, msg, ErrRoute(ErrNoRoute, fmt.Errorf("key=%q", key)))
}

// report forwards err to the configured ErrorHandler. ErrorHandler is
// guaranteed non-nil by NewOptions (defaults to
// messaging.DefaultErrorHandler), so the nil-guard is defensive only.
func (r *router[T]) report(ctx context.Context, msg messaging.Message[T], err error) {
	if r.errorHandler == nil {
		return
	}

	r.errorHandler(ctx, msg, err)
}
