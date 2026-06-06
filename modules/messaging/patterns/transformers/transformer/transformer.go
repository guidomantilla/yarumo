package transformer

import (
	"context"
	"fmt"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// transformer is the Message Translator implementation. It owns a
// single subscription on the source channel (registered in Start,
// cancelled in Stop) and forwards each transformed message to the
// destination channel.
type transformer[T, U any] struct {
	name         string
	src          messaging.Channel[T]
	dst          messaging.Channel[U]
	transform    TransformFn[T, U]
	errorHandler messaging.ErrorHandler

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu     sync.Mutex
	cancel messaging.Cancel
}

// NewTransformer constructs a Message Translator that subscribes to src
// and forwards every transformed Message[U] to dst. The transformer is
// not running on return; call lifecycle.Build (or Start directly) to
// register the subscription.
//
// name is used in lifecycle logs and must be non-empty. src, dst and
// transform are mandatory. The only optional behavior is the
// ErrorHandler hook installed via WithErrorHandler (defaulting to
// messaging.DefaultErrorHandler which logs via common/log).
func NewTransformer[T, U any](name string, src messaging.Channel[T], dst messaging.Channel[U], transform TransformFn[T, U], opts ...Option) Transformer[T, U] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(dst, "destination channel is nil")
	cassert.NotNil(transform, "transform function is nil")

	options := NewOptions(opts...)

	return &transformer[T, U]{
		name:         name,
		src:          src,
		dst:          dst,
		transform:    transform,
		errorHandler: options.errorHandler,
		done:         make(chan struct{}),
	}
}

// Name returns the transformer's identity used in lifecycle logs.
func (x *transformer[T, U]) Name() string {
	cassert.NotNil(x, "transformer is nil")

	return x.name
}

// Start registers the transforming handler as a subscriber on the
// source channel. It satisfies the lifecycle.Component worker-style
// contract: Start returns immediately after the subscription is in
// place; the actual dispatching runs in the source channel's goroutine
// model. Start is idempotent — a second invocation returns nil without
// re-subscribing.
func (x *transformer[T, U]) Start(_ context.Context) error {
	cassert.NotNil(x, "transformer is nil")

	var startErr error

	x.startOnce.Do(func() {
		cancel, err := x.src.Subscribe(x.handle)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		x.mu.Lock()
		x.cancel = cancel
		x.mu.Unlock()
	})

	return startErr
}

// Stop cancels the source-channel subscription and closes Done. Stop
// is idempotent per the lifecycle.Component contract. It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when ctx
// is already expired on entry; otherwise nil.
func (x *transformer[T, U]) Stop(ctx context.Context) error {
	cassert.NotNil(x, "transformer is nil")

	x.stopOnce.Do(func() {
		x.mu.Lock()
		cancel := x.cancel
		x.cancel = nil
		x.mu.Unlock()

		if cancel != nil {
			cancel()
		}

		x.doneOnce.Do(func() { close(x.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (x *transformer[T, U]) Done() <-chan struct{} {
	cassert.NotNil(x, "transformer is nil")

	return x.done
}

// handle is the Handler[T] subscribed on the source channel. It runs
// the transform under panic recovery, then forwards the resulting
// Message[U] to the destination channel. Transformation and forward
// failures are reported through the configured ErrorHandler; the
// function itself always returns nil so transformer concerns never
// propagate to the source channel's Send caller.
func (x *transformer[T, U]) handle(ctx context.Context, msg messaging.Message[T]) error {
	out, err := x.transformWithRecover(ctx, msg)
	if err != nil {
		x.report(ctx, msg, err)

		return nil
	}

	err = x.dst.Send(ctx, out)
	if err != nil {
		x.report(ctx, msg, ErrTransformer(ErrForwardFailed, err))
	}

	return nil
}

// transformWithRecover invokes the user-supplied TransformFn under
// panic recovery. Panics become ErrTransformer(ErrTransformerPanic,
// ...) errors; normal errors become ErrTransformer(ErrTransformFailed,
// err).
func (x *transformer[T, U]) transformWithRecover(ctx context.Context, msg messaging.Message[T]) (out messaging.Message[U], err error) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}

		out = messaging.Message[U]{}
		err = ErrTransformer(ErrTransformerPanic, fmt.Errorf("%v", rec))
	}()

	out, err = x.transform(ctx, msg)
	if err != nil {
		return messaging.Message[U]{}, ErrTransformer(ErrTransformFailed, err)
	}

	return out, nil
}

// report forwards err to the configured ErrorHandler. ErrorHandler is
// guaranteed non-nil by NewOptions (defaults to
// messaging.DefaultErrorHandler), so the nil-guard is defensive only.
func (x *transformer[T, U]) report(ctx context.Context, msg messaging.Message[T], err error) {
	if x.errorHandler == nil {
		return
	}

	x.errorHandler(ctx, msg, err)
}
