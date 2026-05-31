package controlbus

import (
	"context"
	"fmt"
	"maps"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// controlBus is the Control Bus implementation. It owns a single
// subscription on the command channel (registered in Start, cancelled
// in Stop), dispatches each Command to the registered Handler by Verb,
// and publishes the resulting Result to the reply channel.
type controlBus struct {
	name               string
	cmdChan            messaging.Channel[Command]
	resChan            messaging.Channel[Result]
	handlers           map[string]Handler
	unknownVerbHandler Handler
	errorHandler       messaging.ErrorHandler

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu     sync.Mutex
	cancel messaging.Cancel
}

// NewControlBus constructs a ControlBus that subscribes to cmdChan and
// publishes each handler's Result to resChan. The bus is not running on
// return; call lifecycle.Build (or Start directly) to register the
// subscription.
//
// name is used in lifecycle logs and must be non-empty. cmdChan, resChan
// and a non-nil handlers map are mandatory. The handlers map is copied
// defensively so post-construction mutation of the caller's map does
// not affect dispatch. Optional behaviors:
//
//   - WithErrorHandler overrides the default
//     messaging.DefaultErrorHandler (which logs via common/log) with a
//     custom hook for handler panics and forward failures.
//   - WithUnknownVerbHandler overrides the default unknown-verb response
//     (Result{Success: false, Message: "unknown verb"}).
func NewControlBus(name string, cmdChan messaging.Channel[Command], resChan messaging.Channel[Result], handlers map[string]Handler, opts ...Option) ControlBus {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(cmdChan, "command channel is nil")
	cassert.NotNil(resChan, "reply channel is nil")
	cassert.NotNil(handlers, "handlers map is nil")

	options := NewOptions(opts...)

	return &controlBus{
		name:               name,
		cmdChan:            cmdChan,
		resChan:            resChan,
		handlers:           maps.Clone(handlers),
		unknownVerbHandler: options.unknownVerbHandler,
		errorHandler:       options.errorHandler,
		done:               make(chan struct{}),
	}
}

// Name returns the bus's identity used in lifecycle logs.
func (b *controlBus) Name() string {
	cassert.NotNil(b, "controlBus is nil")

	return b.name
}

// Start registers the dispatch handler as a subscriber on the command
// channel. It satisfies the lifecycle.Component worker-style contract:
// Start returns immediately after the subscription is in place; the
// actual dispatching runs in the command channel's goroutine model.
// Start is idempotent — a second invocation returns nil without
// re-subscribing.
func (b *controlBus) Start(_ context.Context) error {
	cassert.NotNil(b, "controlBus is nil")

	var startErr error

	b.startOnce.Do(func() {
		cancel, err := b.cmdChan.Subscribe(b.handle)
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

// Stop cancels the command-channel subscription and closes Done. Stop
// is idempotent per the lifecycle.Component contract. It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when ctx
// is already expired on entry; otherwise nil.
func (b *controlBus) Stop(ctx context.Context) error {
	cassert.NotNil(b, "controlBus is nil")

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
func (b *controlBus) Done() <-chan struct{} {
	cassert.NotNil(b, "controlBus is nil")

	return b.done
}

// handle is the Handler[Command] subscribed on the command channel. It
// resolves the verb-specific handler, invokes it under panic recovery,
// and publishes the resulting Result to the reply channel. The function
// itself always returns nil so bus concerns never propagate to the
// command channel's Send caller.
func (b *controlBus) handle(ctx context.Context, msg messaging.Message[Command]) error {
	handler := b.resolve(msg.Payload.Verb)

	result := b.invokeWithRecover(ctx, msg.Payload, handler)

	err := b.resChan.Send(ctx, messaging.Message[Result]{Payload: result})
	if err != nil {
		b.report(ctx, msg, ErrControlBus(ErrForwardFailed, err))
	}

	return nil
}

// resolve returns the registered Handler for verb, or the configured
// UnknownVerbHandler when verb is not in the registry. The registry
// was cloned at construction so concurrent mutation by the caller
// cannot affect dispatch.
func (b *controlBus) resolve(verb string) Handler {
	h, ok := b.handlers[verb]
	if !ok {
		return b.unknownVerbHandler
	}

	return h
}

// invokeWithRecover runs the chosen Handler under panic recovery. A
// panic becomes a Result{Success: false} whose Message records the
// panic value and triggers the ErrorHandler with ErrHandlerPanic.
func (b *controlBus) invokeWithRecover(ctx context.Context, cmd Command, handler Handler) (result Result) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}

		result = Result{
			Command: cmd,
			Success: false,
			Message: fmt.Sprintf("handler panicked: %v", rec),
		}

		b.report(ctx, messaging.Message[Command]{Payload: cmd}, ErrControlBus(ErrHandlerPanic, fmt.Errorf("%v", rec)))
	}()

	return handler(ctx, cmd)
}

// report forwards err to the configured ErrorHandler. ErrorHandler is
// guaranteed non-nil by NewOptions (defaults to
// messaging.DefaultErrorHandler), so the nil-guard is defensive only.
func (b *controlBus) report(ctx context.Context, msg messaging.Message[Command], err error) {
	if b.errorHandler == nil {
		return
	}

	b.errorHandler(ctx, msg, err)
}
