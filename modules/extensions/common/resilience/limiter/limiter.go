package limiter

import (
	"context"

	"golang.org/x/time/rate"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// limiter is the private implementation of the Limiter interface. It
// holds the underlying golang.org/x/time/rate.Limiter directly; methods
// delegate without any indirection layer.
type limiter struct {
	bucket *rate.Limiter
}

// NewLimiter constructs a Limiter configured via opts. Defaults: ~10 rps
// (one token every 100ms), burst 10. The returned Limiter is safe for
// concurrent use.
func NewLimiter(opts ...Option) Limiter {
	options := NewOptions(opts...)

	return &limiter{
		bucket: rate.NewLimiter(options.rateLimit(), options.burst),
	}
}

// Allow reports whether a token is available right now without blocking.
func (l *limiter) Allow() bool {
	cassert.NotNil(l, "limiter is nil")

	return l.bucket.Allow()
}

// Wait blocks until a token is available or ctx is canceled. Returns an
// error wrapping ErrWaitFailed when ctx is nil, when ctx expires, or when
// the underlying limiter rejects the request.
func (l *limiter) Wait(ctx context.Context) error {
	cassert.NotNil(l, "limiter is nil")

	if ctx == nil {
		return ErrWait(ErrContextNil)
	}

	err := l.bucket.Wait(ctx)
	if err != nil {
		return ErrWait(err)
	}

	return nil
}
