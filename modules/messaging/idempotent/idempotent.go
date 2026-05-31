package idempotent

import (
	"context"
	"sync"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
	"github.com/guidomantilla/yarumo/messaging/store"
)

// idempotent is the Idempotent Receiver implementation. It owns a
// single subscription on the source channel (registered in Start,
// cancelled in Stop) and forwards each received message to the
// destination only when the configured KeyFn yields a non-empty key
// AND the MetadataStore reports the key as unseen within the TTL
// window.
type idempotent[T any] struct {
	name         string
	src          messaging.Channel[T]
	dst          messaging.Channel[T]
	metaStore    store.MetadataStore
	ttl          time.Duration
	keyFn        KeyFn[T]
	errorHandler messaging.ErrorHandler
	dropHandler  DropHandler

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu     sync.Mutex
	cancel messaging.Cancel
}

// NewIdempotent constructs an Idempotent Receiver that subscribes to
// src and forwards each Message[T] to dst only when its dedup key has
// not been recorded in metaStore within the TTL window. Duplicates
// (key already recorded) and keyless messages (KeyFn returned empty)
// are dropped (observable via WithDropHandler when wired).
//
// name is used in lifecycle logs and must be non-empty. src, dst and
// metaStore are mandatory. Optional behaviors:
//
//   - WithTTL overrides the default 24h dedup window.
//   - WithKeyFn overrides the default Headers.MessageID extractor.
//   - WithErrorHandler overrides the default
//     messaging.DefaultErrorHandler (which logs via common/log) with a
//     custom hook for store errors and forward failures.
//   - WithDropHandler installs an optional hook for observing
//     intentional drops; nil by default (silent drop).
func NewIdempotent[T any](name string, src messaging.Channel[T], dst messaging.Channel[T], metaStore store.MetadataStore, opts ...Option[T]) Idempotent[T] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(dst, "destination channel is nil")
	cassert.NotNil(metaStore, "metadata store is nil")

	options := NewOptions(opts...)

	return &idempotent[T]{
		name:         name,
		src:          src,
		dst:          dst,
		metaStore:    metaStore,
		ttl:          options.ttl,
		keyFn:        options.keyFn,
		errorHandler: options.errorHandler,
		dropHandler:  options.dropHandler,
		done:         make(chan struct{}),
	}
}

// Name returns the receiver's identity used in lifecycle logs.
func (i *idempotent[T]) Name() string {
	cassert.NotNil(i, "idempotent is nil")

	return i.name
}

// Start registers the dedup handler as a subscriber on the source
// channel. It satisfies the lifecycle.Component worker-style contract:
// Start returns immediately after the subscription is in place; the
// actual dispatching runs in the source channel's goroutine model.
// Start is idempotent — a second invocation returns nil without
// re-subscribing.
func (i *idempotent[T]) Start(_ context.Context) error {
	cassert.NotNil(i, "idempotent is nil")

	var startErr error

	i.startOnce.Do(func() {
		cancel, err := i.src.Subscribe(i.handle)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		i.mu.Lock()
		i.cancel = cancel
		i.mu.Unlock()
	})

	return startErr
}

// Stop cancels the source-channel subscription and closes Done. Stop
// is idempotent per the lifecycle.Component contract. It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when ctx
// is already expired on entry; otherwise nil. Stop does NOT stop the
// underlying MetadataStore — the store is caller-owned and may be
// shared across receivers.
func (i *idempotent[T]) Stop(ctx context.Context) error {
	cassert.NotNil(i, "idempotent is nil")

	i.stopOnce.Do(func() {
		i.mu.Lock()
		cancel := i.cancel
		i.cancel = nil
		i.mu.Unlock()

		if cancel != nil {
			cancel()
		}

		i.doneOnce.Do(func() { close(i.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (i *idempotent[T]) Done() <-chan struct{} {
	cassert.NotNil(i, "idempotent is nil")

	return i.done
}

// handle is the Handler[T] subscribed on the source channel. It
// extracts the dedup key, checks the metadata store, records the key
// on first sight, and forwards the message to the destination. The
// function itself always returns nil so idempotent concerns never
// propagate to the source channel's Send caller.
func (i *idempotent[T]) handle(ctx context.Context, msg messaging.Message[T]) error {
	key := i.keyFn(msg)
	if key == "" {
		i.reportDrop(ctx, msg, DropReasonNoKey)

		return nil
	}

	seen, err := i.metaStore.Has(ctx, key)
	if err != nil {
		i.reportError(ctx, msg, ErrIdempotent(ErrStoreCheck, err))

		return nil
	}

	if seen {
		i.reportDrop(ctx, msg, DropReasonDuplicate)

		return nil
	}

	err = i.metaStore.Add(ctx, key, i.ttl)
	if err != nil {
		// Fail-open: surface the recording failure but still forward.
		// A future duplicate is preferred over a known drop.
		i.reportError(ctx, msg, ErrIdempotent(ErrStoreAdd, err))
	}

	err = i.dst.Send(ctx, msg)
	if err != nil {
		i.reportError(ctx, msg, ErrIdempotent(ErrForwardFailed, err))
	}

	return nil
}

// reportError forwards err to the configured ErrorHandler. ErrorHandler
// is guaranteed non-nil by NewOptions (defaults to
// messaging.DefaultErrorHandler), so the nil-guard is defensive only.
func (i *idempotent[T]) reportError(ctx context.Context, msg messaging.Message[T], err error) {
	if i.errorHandler == nil {
		return
	}

	i.errorHandler(ctx, msg, err)
}

// reportDrop forwards msg + reason to the configured DropHandler.
// DropHandler is nil by default (silent drops); the guard skips
// invocation in that case.
func (i *idempotent[T]) reportDrop(ctx context.Context, msg messaging.Message[T], reason DropReason) {
	if i.dropHandler == nil {
		return
	}

	i.dropHandler(ctx, msg, reason)
}
