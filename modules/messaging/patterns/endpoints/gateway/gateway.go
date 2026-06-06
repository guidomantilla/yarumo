package gateway

import (
	"context"
	"sync"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	cuids "github.com/guidomantilla/yarumo/core/common/uids"
	"github.com/guidomantilla/yarumo/messaging"
)

// gateway is the Messaging Gateway implementation. It tracks pending
// Request invocations in a correlation-id → waiter map, publishes each
// request to requestChan with the matching CorrelationID, and routes
// incoming replies from replyChan back to the corresponding waiter.
type gateway[Req, Res any] struct {
	name           string
	requestChan    messaging.Channel[Req]
	replyChan      messaging.Channel[Res]
	uid            cuids.UID
	requestTimeout time.Duration
	errorHandler   messaging.ErrorHandler

	// pendingMu guards pending; the map carries one buffered chan per
	// in-flight Request so reply-routing never blocks on a slow caller.
	pendingMu sync.Mutex
	pending   map[string]chan Res

	// started flips true after Start completes successfully. Request
	// reads it under pendingMu to refuse work before Start.
	started bool

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu     sync.Mutex
	cancel messaging.Cancel
}

// NewGateway constructs a Gateway over the given request and reply
// channels. The Gateway is not running on return; call lifecycle.Build
// (or Start directly) to subscribe to the reply channel.
//
// name is used as the Headers.ReplyTo value stamped onto outgoing
// requests AND as the lifecycle log identity; it must be non-empty.
// requestChan and replyChan are mandatory. Callers MUST supply a
// cuids.UID generator via WithUIDGenerator — without it every Request
// fails with ErrCorrelationIDFailed.
//
// Optional behaviors:
//
//   - WithRequestTimeout overrides the default 5s per-Request timeout.
//   - WithErrorHandler overrides the default
//     messaging.DefaultErrorHandler (which logs via common/log).
func NewGateway[Req, Res any](name string, requestChan messaging.Channel[Req], replyChan messaging.Channel[Res], opts ...Option) Gateway[Req, Res] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(requestChan, "request channel is nil")
	cassert.NotNil(replyChan, "reply channel is nil")

	options := NewOptions(opts...)

	return &gateway[Req, Res]{
		name:           name,
		requestChan:    requestChan,
		replyChan:      replyChan,
		uid:            options.uid,
		requestTimeout: options.requestTimeout,
		errorHandler:   options.errorHandler,
		pending:        map[string]chan Res{},
		done:           make(chan struct{}),
	}
}

// Name returns the gateway's identity used in lifecycle logs and as
// the outgoing Headers.ReplyTo value.
func (g *gateway[Req, Res]) Name() string {
	cassert.NotNil(g, "gateway is nil")

	return g.name
}

// Start registers the reply-routing handler as a subscriber on the
// reply channel. It satisfies the lifecycle.Component worker-style
// contract: Start returns immediately after the subscription is in
// place; the actual dispatching runs in the reply channel's goroutine
// model. Start is idempotent — a second invocation returns nil
// without re-subscribing.
func (g *gateway[Req, Res]) Start(_ context.Context) error {
	cassert.NotNil(g, "gateway is nil")

	var startErr error

	g.startOnce.Do(func() {
		cancel, err := g.replyChan.Subscribe(g.handleReply)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		g.mu.Lock()
		g.cancel = cancel
		g.mu.Unlock()

		g.pendingMu.Lock()
		g.started = true
		g.pendingMu.Unlock()
	})

	return startErr
}

// Stop cancels the reply-channel subscription, fails every pending
// Request with ErrGatewayShuttingDown, and closes Done. Stop is
// idempotent per the lifecycle.Component contract. It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when ctx
// is already expired on entry; otherwise nil.
func (g *gateway[Req, Res]) Stop(ctx context.Context) error {
	cassert.NotNil(g, "gateway is nil")

	g.stopOnce.Do(func() {
		g.mu.Lock()
		cancel := g.cancel
		g.cancel = nil
		g.mu.Unlock()

		if cancel != nil {
			cancel()
		}

		g.pendingMu.Lock()
		g.started = false

		// Close every pending waiter so the corresponding Request
		// returns ErrGatewayShuttingDown via its select. Closing the
		// chan signals "no value will arrive"; Request distinguishes
		// close-vs-value via the ok bool.
		for id, ch := range g.pending {
			close(ch)
			delete(g.pending, id)
		}
		g.pendingMu.Unlock()

		g.doneOnce.Do(func() { close(g.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (g *gateway[Req, Res]) Done() <-chan struct{} {
	cassert.NotNil(g, "gateway is nil")

	return g.done
}

// Request publishes req to the configured request channel with an
// auto-generated CorrelationID and Headers.ReplyTo = the Gateway's
// name, then waits for a matching Message[Res] on the reply channel.
// The first matching reply is returned. Request honours the caller's
// ctx and the configured per-request timeout (whichever fires first).
//
// Errors:
//
//   - ErrGatewayNotStarted when Start has not yet been called (or Stop
//     has already run).
//   - ErrCorrelationIDFailed when the configured uid generator fails
//     to mint a correlation id.
//   - ErrRequestSendFailed when the request channel's Send returns an
//     error.
//   - ErrRequestTimeout on timeout.
//   - ErrRequestCancelled (joined with ctx.Err()) on caller cancel.
//   - ErrGatewayShuttingDown when Stop runs while waiting.
func (g *gateway[Req, Res]) Request(ctx context.Context, req Req) (Res, error) {
	cassert.NotNil(g, "gateway is nil")

	var zero Res

	if ctx == nil {
		return zero, ErrGateway(messaging.ErrContextNil)
	}

	if g.uid == nil {
		return zero, ErrGateway(ErrCorrelationIDFailed)
	}

	corrID, err := g.uid.Generate()
	if err != nil {
		return zero, ErrGateway(ErrCorrelationIDFailed, err)
	}

	respCh, regErr := g.register(corrID)
	if regErr != nil {
		return zero, regErr
	}

	defer g.unregister(corrID)

	msg := messaging.Message[Req]{
		Payload: req,
		Headers: messaging.Headers{
			MessageID:     corrID,
			CorrelationID: corrID,
			ReplyTo:       g.name,
			Timestamp:     time.Now(),
		},
	}

	err = g.requestChan.Send(ctx, msg)
	if err != nil {
		return zero, ErrGateway(ErrRequestSendFailed, err)
	}

	return g.wait(ctx, respCh)
}

// register installs a buffered waiter chan under corrID in the pending
// map. Returns ErrGatewayNotStarted when the gateway is not running.
// The buffer of 1 lets handleReply deliver without blocking even when
// the caller has already moved past the Request select (e.g. ctx
// cancel raced with reply arrival).
func (g *gateway[Req, Res]) register(corrID string) (chan Res, error) {
	g.pendingMu.Lock()
	defer g.pendingMu.Unlock()

	if !g.started {
		return nil, ErrGateway(ErrGatewayNotStarted)
	}

	ch := make(chan Res, 1)
	g.pending[corrID] = ch

	return ch, nil
}

// unregister removes corrID from the pending map. Safe to call
// unconditionally from Request's defer: a missing entry (Stop already
// closed the waiter, or handleReply already delivered) is a no-op.
func (g *gateway[Req, Res]) unregister(corrID string) {
	g.pendingMu.Lock()
	defer g.pendingMu.Unlock()

	delete(g.pending, corrID)
}

// wait blocks until one of: a reply arrives on respCh, the gateway
// shuts down (respCh closed), the caller's ctx fires, or the per-
// request timeout fires. Returns the appropriate error per Request's
// documented contract.
func (g *gateway[Req, Res]) wait(ctx context.Context, respCh chan Res) (Res, error) {
	var zero Res

	timeoutCtx, cancel := g.deriveCtx(ctx)
	defer cancel()

	select {
	case res, ok := <-respCh:
		if !ok {
			return zero, ErrGateway(ErrGatewayShuttingDown)
		}

		return res, nil
	case <-ctx.Done():
		return zero, ErrGateway(ErrRequestCancelled, ctx.Err())
	case <-timeoutCtx.Done():
		// timeoutCtx fires either because ctx fired (handled above
		// in normal racing) or because the per-request timeout
		// expired. Distinguish the two via ctx.Err() — if the
		// caller's ctx is still live, this is a true timeout.
		if ctx.Err() == nil {
			return zero, ErrGateway(ErrRequestTimeout)
		}

		return zero, ErrGateway(ErrRequestCancelled, ctx.Err())
	}
}

// deriveCtx returns a context derived from ctx that fires after the
// configured request timeout. The caller's ctx deadline always wins
// when it is tighter (context.WithTimeout semantics).
func (g *gateway[Req, Res]) deriveCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, g.requestTimeout)
}

// handleReply is the Handler[Res] subscribed on the reply channel. It
// looks up the pending waiter by Headers.CorrelationID and forwards
// the payload. Replies for unknown CorrelationIDs are dropped and
// surfaced via the ErrorHandler with ErrUnknownCorrelationID. The
// function itself always returns nil so gateway concerns never
// propagate to the reply channel's Send caller.
func (g *gateway[Req, Res]) handleReply(ctx context.Context, msg messaging.Message[Res]) error {
	corrID := msg.Headers.CorrelationID

	g.pendingMu.Lock()
	ch, ok := g.pending[corrID]
	if ok {
		delete(g.pending, corrID)
	}
	g.pendingMu.Unlock()

	if !ok {
		g.report(ctx, msg, ErrGateway(ErrUnknownCorrelationID))

		return nil
	}

	// The waiter chan is buffered (size 1) so this send never blocks.
	ch <- msg.Payload

	return nil
}

// report forwards err to the configured ErrorHandler. ErrorHandler is
// guaranteed non-nil by NewOptions (defaults to
// messaging.DefaultErrorHandler), so the nil-guard is defensive only.
func (g *gateway[Req, Res]) report(ctx context.Context, msg messaging.Message[Res], err error) {
	if g.errorHandler == nil {
		return
	}

	g.errorHandler(ctx, msg, err)
}
