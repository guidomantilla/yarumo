package claimcheck

import (
	"context"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
	"github.com/guidomantilla/yarumo/messaging/stores"
)

// claimCheckIn is the heavy-payload producer side of Claim Check. It
// owns a single subscription on the source Channel[T] (registered in
// Start, cancelled in Stop), stores each received Message[T] in the
// configured MessageStore[T] under a generated key, and forwards a
// fresh Message[ClaimCheckReference] to the configured downstream
// channel.
type claimCheckIn[T any] struct {
	name         string
	src          messaging.Channel[T]
	dst          messaging.Channel[ClaimCheckReference]
	msgStore     stores.MessageStore[T]
	keyGen       KeyGenFn
	errorHandler messaging.ErrorHandler

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu     sync.Mutex
	cancel messaging.Cancel
}

// NewClaimCheckIn constructs the producer-side endpoint of Claim
// Check. It subscribes to src (heavy Message[T]) and, for each
// received envelope, stores the original in msgStore under a
// generated key and forwards a lightweight
// Message[ClaimCheckReference]{Key: key} to dst. The reference message
// preserves Headers.CorrelationID from the original so downstream
// observers can correlate the reference hop back to the original
// flow.
//
// name is used in lifecycle logs and must be non-empty. src, dst and
// msgStore are mandatory. Optional behaviors:
//
//   - WithKeyGen overrides the default crypto/rand-based generator.
//   - WithErrorHandler overrides the default
//     messaging.DefaultErrorHandler (which logs via common/log) with a
//     custom hook for store-put errors and forward failures.
//
// WithDeleteAfterRetrieve is silently ignored — In never deletes.
func NewClaimCheckIn[T any](name string, src messaging.Channel[T], dst messaging.Channel[ClaimCheckReference], msgStore stores.MessageStore[T], opts ...Option) ClaimCheckIn[T] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(dst, "destination channel is nil")
	cassert.NotNil(msgStore, "message store is nil")

	options := NewOptions(opts...)

	return &claimCheckIn[T]{
		name:         name,
		src:          src,
		dst:          dst,
		msgStore:     msgStore,
		keyGen:       options.keyGen,
		errorHandler: options.errorHandler,
		done:         make(chan struct{}),
	}
}

// Name returns the endpoint's identity used in lifecycle logs.
func (c *claimCheckIn[T]) Name() string {
	cassert.NotNil(c, "claim check in is nil")

	return c.name
}

// Start registers the offload handler as a subscriber on the source
// channel. It satisfies the lifecycle.Component worker-style contract:
// Start returns immediately after the subscription is in place; the
// actual dispatching runs in the source channel's goroutine model.
// Start is idempotent — a second invocation returns nil without
// re-subscribing.
func (c *claimCheckIn[T]) Start(_ context.Context) error {
	cassert.NotNil(c, "claim check in is nil")

	var startErr error

	c.startOnce.Do(func() {
		cancel, err := c.src.Subscribe(c.handle)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		c.mu.Lock()
		c.cancel = cancel
		c.mu.Unlock()
	})

	return startErr
}

// Stop cancels the source-channel subscription and closes Done. Stop
// is idempotent per the lifecycle.Component contract. It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when
// ctx is already expired on entry; otherwise nil. Stop does NOT stop
// the underlying MessageStore — the store is caller-owned and may be
// shared with the matching ClaimCheckOut (and other endpoints).
func (c *claimCheckIn[T]) Stop(ctx context.Context) error {
	cassert.NotNil(c, "claim check in is nil")

	c.stopOnce.Do(func() {
		c.mu.Lock()
		cancel := c.cancel
		c.cancel = nil
		c.mu.Unlock()

		if cancel != nil {
			cancel()
		}

		c.doneOnce.Do(func() { close(c.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (c *claimCheckIn[T]) Done() <-chan struct{} {
	cassert.NotNil(c, "claim check in is nil")

	return c.done
}

// handle is the Handler[T] subscribed on the source channel. It
// stores the original envelope in the message store and forwards a
// reference envelope downstream. The function itself always returns
// nil so claim check concerns never propagate to the source channel's
// Send caller.
//
// Store-put failures are fail-CLOSED: when Put errors, the reference
// is NOT forwarded (the payload was never stored, downstream cannot
// recover) and the error is routed through WithErrorHandler.
func (c *claimCheckIn[T]) handle(ctx context.Context, msg messaging.Message[T]) error {
	key := c.keyGen()

	err := c.msgStore.Put(ctx, key, msg)
	if err != nil {
		c.reportError(ctx, msg, ErrClaimCheck(ErrStorePut, err))

		return nil
	}

	ref := messaging.Message[ClaimCheckReference]{
		Payload: ClaimCheckReference{Key: key},
		Headers: messaging.Headers{
			// CorrelationID is preserved so downstream observers can
			// correlate the reference hop (and the future
			// ClaimCheckOut Get) back to the original flow. MessageID
			// is intentionally left empty: the reference is a NEW
			// envelope; reusing the original MessageID would collide
			// on any downstream dedup path.
			CorrelationID: msg.Headers.CorrelationID,
			Timestamp:     msg.Headers.Timestamp,
		},
	}

	err = c.dst.Send(ctx, ref)
	if err != nil {
		c.reportError(ctx, msg, ErrClaimCheck(ErrForwardFailed, err))
	}

	return nil
}

// reportError forwards err to the configured ErrorHandler. ErrorHandler
// is guaranteed non-nil by NewOptions (defaults to
// messaging.DefaultErrorHandler), so the nil-guard is defensive only.
func (c *claimCheckIn[T]) reportError(ctx context.Context, msg messaging.Message[T], err error) {
	if c.errorHandler == nil {
		return
	}

	c.errorHandler(ctx, msg, err)
}

// claimCheckOut is the heavy-payload consumer side of Claim Check. It
// owns a single subscription on a source Channel[ClaimCheckReference]
// (registered in Start, cancelled in Stop), retrieves the original
// Message[T] from msgStore using the reference's key, optionally
// deletes the entry, and forwards the original to dst.
type claimCheckOut[T any] struct {
	name                string
	src                 messaging.Channel[ClaimCheckReference]
	dst                 messaging.Channel[T]
	msgStore            stores.MessageStore[T]
	deleteAfterRetrieve bool
	errorHandler        messaging.ErrorHandler

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu     sync.Mutex
	cancel messaging.Cancel
}

// NewClaimCheckOut constructs the consumer-side endpoint of Claim
// Check. It subscribes to src (lightweight
// Message[ClaimCheckReference]) and, for each received reference,
// retrieves the original Message[T] from msgStore and forwards it to
// dst. By default the store entry is deleted after a successful
// retrieval (configurable via WithDeleteAfterRetrieve).
//
// name is used in lifecycle logs and must be non-empty. src, dst and
// msgStore are mandatory. Optional behaviors:
//
//   - WithDeleteAfterRetrieve toggles cleanup after Get (default true).
//   - WithErrorHandler overrides the default
//     messaging.DefaultErrorHandler (which logs via common/log) with a
//     custom hook for store-get/delete errors and forward failures.
//
// WithKeyGen is silently ignored — Out never generates keys.
func NewClaimCheckOut[T any](name string, src messaging.Channel[ClaimCheckReference], dst messaging.Channel[T], msgStore stores.MessageStore[T], opts ...Option) ClaimCheckOut[T] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(dst, "destination channel is nil")
	cassert.NotNil(msgStore, "message store is nil")

	options := NewOptions(opts...)

	return &claimCheckOut[T]{
		name:                name,
		src:                 src,
		dst:                 dst,
		msgStore:            msgStore,
		deleteAfterRetrieve: options.deleteAfterRetrieve,
		errorHandler:        options.errorHandler,
		done:                make(chan struct{}),
	}
}

// Name returns the endpoint's identity used in lifecycle logs.
func (c *claimCheckOut[T]) Name() string {
	cassert.NotNil(c, "claim check out is nil")

	return c.name
}

// Start registers the retrieval handler as a subscriber on the source
// channel. It satisfies the lifecycle.Component worker-style contract:
// Start returns immediately after the subscription is in place; the
// actual dispatching runs in the source channel's goroutine model.
// Start is idempotent — a second invocation returns nil without
// re-subscribing.
func (c *claimCheckOut[T]) Start(_ context.Context) error {
	cassert.NotNil(c, "claim check out is nil")

	var startErr error

	c.startOnce.Do(func() {
		cancel, err := c.src.Subscribe(c.handle)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		c.mu.Lock()
		c.cancel = cancel
		c.mu.Unlock()
	})

	return startErr
}

// Stop cancels the source-channel subscription and closes Done. Stop
// is idempotent per the lifecycle.Component contract. It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when
// ctx is already expired on entry; otherwise nil. Stop does NOT stop
// the underlying MessageStore — the store is caller-owned and may be
// shared with the matching ClaimCheckIn (and other endpoints).
func (c *claimCheckOut[T]) Stop(ctx context.Context) error {
	cassert.NotNil(c, "claim check out is nil")

	c.stopOnce.Do(func() {
		c.mu.Lock()
		cancel := c.cancel
		c.cancel = nil
		c.mu.Unlock()

		if cancel != nil {
			cancel()
		}

		c.doneOnce.Do(func() { close(c.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (c *claimCheckOut[T]) Done() <-chan struct{} {
	cassert.NotNil(c, "claim check out is nil")

	return c.done
}

// handle is the Handler[ClaimCheckReference] subscribed on the source
// channel. It retrieves the original from the message store using the
// reference's key, optionally deletes the entry, and forwards the
// original to the destination. The function itself always returns nil
// so claim check concerns never propagate to the source channel's
// Send caller.
//
// Store-get failures (including ErrStoreNotFound surfaced by the
// store as an error) are fail-CLOSED: when Get errors, the original
// is NOT forwarded and the error is routed through WithErrorHandler.
// Delete failures are fail-OPEN: the original is still forwarded —
// losing the chance to clean up the store is preferable to losing the
// message.
func (c *claimCheckOut[T]) handle(ctx context.Context, ref messaging.Message[ClaimCheckReference]) error {
	original, err := c.msgStore.Get(ctx, ref.Payload.Key)
	if err != nil {
		c.reportRefError(ctx, ref, ErrClaimCheck(ErrStoreGet, err))

		return nil
	}

	if c.deleteAfterRetrieve {
		err = c.msgStore.Delete(ctx, ref.Payload.Key)
		if err != nil {
			// Fail-open: surface the delete failure but still
			// forward. Losing the chance to clean up the store is
			// preferable to losing the message.
			c.reportRefError(ctx, ref, ErrClaimCheck(ErrStoreDelete, err))
		}
	}

	err = c.dst.Send(ctx, original)
	if err != nil {
		c.reportError(ctx, original, ErrClaimCheck(ErrForwardFailed, err))
	}

	return nil
}

// reportRefError forwards err to the configured ErrorHandler, passing
// the reference envelope as the msg argument so observers can map the
// failure back to the upstream reference hop.
func (c *claimCheckOut[T]) reportRefError(ctx context.Context, ref messaging.Message[ClaimCheckReference], err error) {
	if c.errorHandler == nil {
		return
	}

	c.errorHandler(ctx, ref, err)
}

// reportError forwards err to the configured ErrorHandler, passing
// the retrieved original envelope as the msg argument so observers
// can map a forward failure back to the payload being dispatched.
func (c *claimCheckOut[T]) reportError(ctx context.Context, msg messaging.Message[T], err error) {
	if c.errorHandler == nil {
		return
	}

	c.errorHandler(ctx, msg, err)
}
