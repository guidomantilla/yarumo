package breaker

import (
	"context"

	"github.com/sony/gobreaker"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

// breaker is the private implementation of the Breaker interface. It
// holds the underlying gobreaker.CircuitBreaker directly; methods
// delegate without indirection.
type breaker struct {
	cb *gobreaker.CircuitBreaker
}

// NewBreaker constructs a Breaker configured via opts. Defaults: 5
// consecutive failures trip closed → open, 15s open timeout before
// half-open, 1 probe in half-open, 60s counter reset cycle in closed,
// no state-change hook. The returned Breaker is safe for concurrent
// use.
func NewBreaker(opts ...Option) Breaker {
	options := NewOptions(opts...)

	return &breaker{
		cb: gobreaker.NewCircuitBreaker(settingsFor(options)),
	}
}

// Execute invokes fn through the breaker. Returns nil when fn returns
// nil; otherwise returns an error wrapping ErrBreakerFailed plus the
// underlying error (ErrBreakerOpen / ErrBreakerTooManyRequests when the
// breaker rejected the call, or whatever fn returned when it ran).
func (b *breaker) Execute(ctx context.Context, fn func() error) error {
	cassert.NotNil(b, "breaker receiver is nil")

	if ctx == nil {
		return ErrBreaker(ErrContextNil)
	}
	if fn == nil {
		return ErrBreaker(ErrFnNil)
	}

	_, err := b.cb.Execute(func() (any, error) {
		ctxErr := ctx.Err()
		if ctxErr != nil {
			return nil, ctxErr
		}
		return nil, fn()
	})
	if err != nil {
		return ErrBreaker(translateBreakerError(err))
	}

	return nil
}

// State returns the current operating state of the breaker.
func (b *breaker) State() State {
	cassert.NotNil(b, "breaker receiver is nil")

	return fromGobreakerState(b.cb.State())
}
