package enricher

import (
	"context"
	"fmt"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// enricher is the Content/Header Enricher implementation. It owns a
// single subscription on the source channel (registered in Start,
// cancelled in Stop) and forwards each received message to the
// destination after applying the user-supplied EnrichFn.
type enricher[T any] struct {
	name         string
	src          messaging.Channel[T]
	dst          messaging.Channel[T]
	enrich       EnrichFn[T]
	errorHandler messaging.ErrorHandler

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu     sync.Mutex
	cancel messaging.Cancel
}

// NewEnricher constructs an Enricher that subscribes to src and forwards
// each enrich(msg) to dst. The enricher is not running on return; call
// lifecycle.Build (or Start directly) to register the subscription.
//
// name is used in lifecycle logs and must be non-empty. src, dst and
// enrich are mandatory. The only optional behavior is the ErrorHandler
// hook installed via WithErrorHandler (defaulting to
// messaging.DefaultErrorHandler which logs via common/log).
func NewEnricher[T any](name string, src messaging.Channel[T], dst messaging.Channel[T], enrich EnrichFn[T], opts ...Option) Enricher[T] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(dst, "destination channel is nil")
	cassert.NotNil(enrich, "enrich function is nil")

	options := NewOptions(opts...)

	return &enricher[T]{
		name:         name,
		src:          src,
		dst:          dst,
		enrich:       enrich,
		errorHandler: options.errorHandler,
		done:         make(chan struct{}),
	}
}

// Name returns the enricher's identity used in lifecycle logs.
func (e *enricher[T]) Name() string {
	cassert.NotNil(e, "enricher is nil")

	return e.name
}

// Start registers the enriching handler as a subscriber on the source
// channel. It satisfies the lifecycle.Component worker-style contract:
// Start returns immediately after the subscription is in place; the
// actual dispatching runs in the source channel's goroutine model.
// Start is idempotent — a second invocation returns nil without
// re-subscribing.
func (e *enricher[T]) Start(_ context.Context) error {
	cassert.NotNil(e, "enricher is nil")

	var startErr error

	e.startOnce.Do(func() {
		cancel, err := e.src.Subscribe(e.handle)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		e.mu.Lock()
		e.cancel = cancel
		e.mu.Unlock()
	})

	return startErr
}

// Stop cancels the source-channel subscription and closes Done. Stop
// is idempotent per the lifecycle.Component contract. It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when ctx
// is already expired on entry; otherwise nil.
func (e *enricher[T]) Stop(ctx context.Context) error {
	cassert.NotNil(e, "enricher is nil")

	e.stopOnce.Do(func() {
		e.mu.Lock()
		cancel := e.cancel
		e.cancel = nil
		e.mu.Unlock()

		if cancel != nil {
			cancel()
		}

		e.doneOnce.Do(func() { close(e.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (e *enricher[T]) Done() <-chan struct{} {
	cassert.NotNil(e, "enricher is nil")

	return e.done
}

// handle is the Handler[T] subscribed on the source channel. It runs
// the enrich callback under panic recovery, then forwards the enriched
// message. EnrichFn failures and forward failures are reported through
// the configured ErrorHandler; the function itself always returns nil
// so enricher concerns never propagate to the source channel's Send
// caller.
func (e *enricher[T]) handle(ctx context.Context, msg messaging.Message[T]) error {
	enriched, err := e.enrichWithRecover(ctx, msg)
	if err != nil {
		e.reportError(ctx, msg, err)

		return nil
	}

	err = e.dst.Send(ctx, enriched)
	if err != nil {
		e.reportError(ctx, msg, ErrEnricher(ErrForwardFailed, err))
	}

	return nil
}

// enrichWithRecover invokes the user-supplied EnrichFn under panic
// recovery. Panics become ErrEnricher(ErrEnrichPanic, ...) errors;
// normal errors become ErrEnricher(ErrEnrichFnFailed, err).
func (e *enricher[T]) enrichWithRecover(ctx context.Context, msg messaging.Message[T]) (out messaging.Message[T], err error) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}

		out = messaging.Message[T]{}
		err = ErrEnricher(ErrEnrichPanic, fmt.Errorf("%v", rec))
	}()

	out, err = e.enrich(ctx, msg)
	if err != nil {
		return messaging.Message[T]{}, ErrEnricher(ErrEnrichFnFailed, err)
	}

	return out, nil
}

// reportError forwards err to the configured ErrorHandler. ErrorHandler
// is guaranteed non-nil by NewOptions (defaults to
// messaging.DefaultErrorHandler), so the nil-guard is defensive only.
func (e *enricher[T]) reportError(ctx context.Context, msg messaging.Message[T], err error) {
	if e.errorHandler == nil {
		return
	}

	e.errorHandler(ctx, msg, err)
}
