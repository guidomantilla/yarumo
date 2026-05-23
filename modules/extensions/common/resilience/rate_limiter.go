package resilience

import (
	"context"

	"golang.org/x/time/rate"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// rateLimiter is the private implementation of the RateLimiter interface.
//
// Behaviour is pluggable via function fields populated at construction time
// by newRateLimiter. The struct itself contains no backend-specific state —
// each method just delegates to its function field after asserting the
// receiver is non-nil. This follows criterion 4 Exception 3 (Pluggable struct
// pattern) from the common coding standards and mirrors the cache refactor
// (commit cd0804f) and crypto's *Method.
type rateLimiter struct {
	// Pluggable function fields — the limiter's behaviour IS these closures.
	allowFn func() bool
	waitFn  func(ctx context.Context) error
}

// newRateLimiter builds a rateLimiter from the configured options. The
// underlying *rate.Limiter is captured in the allowFn / waitFn closures
// rather than stored as a struct field.
func newRateLimiter(opts *Options) *rateLimiter {
	cassert.NotNil(opts, "options is nil")

	limiter := rate.NewLimiter(opts.rateLimit(), opts.rateBurst)

	return &rateLimiter{
		allowFn: func() bool {
			return limiter.Allow()
		},
		waitFn: func(ctx context.Context) error {
			err := validateWait(ctx)
			if err != nil {
				return ErrRateLimiterWait(err)
			}

			err = limiter.Wait(ctx)
			if err != nil {
				return ErrRateLimiterWait(cerrs.Wrap(err))
			}

			return nil
		},
	}
}

// Allow reports whether a token is available right now without blocking.
func (l *rateLimiter) Allow() bool {
	cassert.NotNil(l, "rate limiter is nil")

	return l.allowFn()
}

// Wait blocks until a token is available or ctx is canceled.
func (l *rateLimiter) Wait(ctx context.Context) error {
	cassert.NotNil(l, "rate limiter is nil")

	return l.waitFn(ctx)
}
