package bridge

import (
	"context"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/core/common/messaging"
)

// bridge is the one-to-one channel forwarder implementation. It owns a
// single subscription on the source channel (registered in Start,
// cancelled in Stop) and forwards each received message to the
// configured destination channel unchanged.
type bridge[T any] struct {
	name         string
	src          messaging.Channel[T]
	dst          messaging.Channel[T]
	errorHandler messaging.ErrorHandler

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu     sync.Mutex
	cancel messaging.Cancel
}

// NewBridge constructs a Bridge that subscribes to src and forwards
// every received Message[T] to dst. The bridge is not running on
// return; call lifecycle.Build (or Start directly) to register the
// subscription.
//
// name is used in lifecycle logs and must be non-empty. src and dst
// are mandatory. The only optional behavior is the ErrorHandler hook
// installed via WithErrorHandler (defaulting to
// messaging.DefaultErrorHandler which logs via common/log).
func NewBridge[T any](name string, src messaging.Channel[T], dst messaging.Channel[T], opts ...Option) lifecycle.Component {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(dst, "destination channel is nil")

	options := NewOptions(opts...)

	return &bridge[T]{
		name:         name,
		src:          src,
		dst:          dst,
		errorHandler: options.errorHandler,
		done:         make(chan struct{}),
	}
}

// Name returns the bridge's identity used in lifecycle logs.
func (b *bridge[T]) Name() string {
	cassert.NotNil(b, "bridge is nil")

	return b.name
}

// Start registers the forwarding handler as a subscriber on the source
// channel. It satisfies the lifecycle.Component worker-style contract:
// Start returns immediately after the subscription is in place; the
// actual dispatching runs in the source channel's goroutine model.
// Start is idempotent — a second invocation returns nil without
// re-subscribing.
func (b *bridge[T]) Start(_ context.Context) error {
	cassert.NotNil(b, "bridge is nil")

	var startErr error

	b.startOnce.Do(func() {
		cancel, err := b.src.Subscribe(b.handle)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		b.mu.Lock()
		b.cancel = cancel
		b.mu.Unlock()
	})

	return startErr
}

// Stop cancels the source-channel subscription and closes Done. Stop
// is idempotent per the lifecycle.Component contract. It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when ctx
// is already expired on entry; otherwise nil.
func (b *bridge[T]) Stop(ctx context.Context) error {
	cassert.NotNil(b, "bridge is nil")

	b.stopOnce.Do(func() {
		b.mu.Lock()
		cancel := b.cancel
		b.cancel = nil
		b.mu.Unlock()

		if cancel != nil {
			cancel()
		}

		b.doneOnce.Do(func() { close(b.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (b *bridge[T]) Done() <-chan struct{} {
	cassert.NotNil(b, "bridge is nil")

	return b.done
}

// handle is the Handler[T] subscribed on the source channel. It
// forwards msg to the destination channel; forward failures are
// reported through the configured ErrorHandler. The function itself
// always returns nil so bridge concerns never propagate to the source
// channel's Send caller.
func (b *bridge[T]) handle(ctx context.Context, msg messaging.Message[T]) error {
	err := b.dst.Send(ctx, msg)
	if err != nil && b.errorHandler != nil {
		b.errorHandler(ctx, msg, ErrBridge(ErrForwardFailed, err))
	}

	return nil
}
