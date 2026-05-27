package retry

import (
	"context"
	"time"

	retrygo "github.com/avast/retry-go/v4"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cretry "github.com/guidomantilla/yarumo/core/common/resilience/retry"
)

// retry is the private implementation of the Retry interface. It captures
// the configured options at construction time and translates them into
// retry-go options on each Do call.
type retry struct {
	attempts uint
	delay    time.Duration
	maxDelay time.Duration
	backoff  cretry.Backoff
	retryIf  cretry.RetryIfFn
	onRetry  cretry.OnRetryFn
}

// NewRetry constructs a Retry configured via opts. Defaults: 3 attempts
// (1 original + 2 retries), 100ms base delay with exponential backoff
// capped at 5s, always retry on non-nil error, no per-attempt hook. The
// returned Retry is safe for concurrent use.
func NewRetry(opts ...Option) cretry.Retry {
	options := NewOptions(opts...)

	return &retry{
		attempts: options.attempts,
		delay:    options.delay,
		maxDelay: options.maxDelay,
		backoff:  options.backoff,
		retryIf:  options.retryIf,
		onRetry:  options.onRetry,
	}
}

// Do invokes fn under the configured retry policy. Returns nil when fn
// eventually returns nil; otherwise returns an error wrapping
// ErrRetryFailed and the last underlying error.
func (r *retry) Do(ctx context.Context, fn func() error) error {
	cassert.NotNil(r, "retry receiver is nil")

	if ctx == nil {
		return cretry.ErrRetry(cretry.ErrContextNil)
	}
	if fn == nil {
		return cretry.ErrRetry(cretry.ErrFnNil)
	}

	err := retrygo.Do(fn,
		retrygo.Context(ctx),
		retrygo.Attempts(r.attempts),
		retrygo.Delay(r.delay),
		retrygo.MaxDelay(r.maxDelay),
		retrygo.DelayType(delayTypeFor(r.backoff)),
		retrygo.RetryIf(retrygo.RetryIfFunc(r.retryIf)),
		retrygo.OnRetry(retrygo.OnRetryFunc(r.onRetry)),
		retrygo.LastErrorOnly(true),
	)
	if err != nil {
		return cretry.ErrRetry(err)
	}

	return nil
}
