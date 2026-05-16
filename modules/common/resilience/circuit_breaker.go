package resilience

import (
	"context"

	"github.com/sony/gobreaker"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// circuitBreaker is the private implementation of the CircuitBreaker interface.
//
// Behaviour is pluggable via function fields populated at construction time by
// newCircuitBreaker. The struct itself contains no backend-specific state —
// each method just delegates to its function field after asserting the
// receiver is non-nil. This follows criterion 4 Exception 3 (Pluggable struct
// pattern) from the common coding standards and mirrors the cache refactor
// (commit cd0804f) and crypto's *Method.
type circuitBreaker struct {
	// Pluggable function fields — the breaker's behaviour IS these closures.
	executeFn func(ctx context.Context, fn func() (any, error)) (any, error)
	stateFn   func() State
}

// newCircuitBreaker builds a circuitBreaker from the configured options. The
// underlying *gobreaker.CircuitBreaker is captured in the executeFn / stateFn
// closures rather than stored as a struct field.
func newCircuitBreaker(name string, opts *Options) *circuitBreaker {
	cassert.NotNil(opts, "options is nil")

	cb := gobreaker.NewCircuitBreaker(settingsFor(name, opts))

	return &circuitBreaker{
		executeFn: func(ctx context.Context, fn func() (any, error)) (any, error) {
			err := validateExecute(ctx, fn)
			if err != nil {
				return nil, ErrCircuitBreakerExecute(err)
			}

			wrapped := func() (any, error) {
				// Surface ctx cancellation as the failure before invoking fn so
				// the breaker records the cancellation as a failure.
				ctxErr := ctx.Err()
				if ctxErr != nil {
					return nil, cerrs.Wrap(ctxErr)
				}

				return fn()
			}

			result, runErr := cb.Execute(wrapped)
			if runErr != nil {
				return nil, ErrCircuitBreakerExecute(translateBreakerError(runErr))
			}

			return result, nil
		},
		stateFn: func() State {
			return fromGobreakerState(cb.State())
		},
	}
}

// Execute invokes fn through the breaker. If the breaker is open the call is
// rejected with ErrCircuitBreakerOpen. If it is half-open and the probe budget
// is exhausted the call is rejected with ErrCircuitBreakerTooManyRequests.
// Failures returned by fn are recorded and may trip the breaker.
func (c *circuitBreaker) Execute(ctx context.Context, fn func() (any, error)) (any, error) {
	cassert.NotNil(c, "circuit breaker is nil")

	return c.executeFn(ctx, fn)
}

// State returns the current operating state of the breaker.
func (c *circuitBreaker) State() State {
	cassert.NotNil(c, "circuit breaker is nil")

	return c.stateFn()
}
