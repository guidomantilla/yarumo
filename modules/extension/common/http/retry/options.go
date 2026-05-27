package retry

// Option is a functional option for configuring retry transport Options.
type Option func(opts *Options)

// Options holds the HTTP-specific configuration of the retry transport.
// Retry-policy knobs (attempts, delay, backoff, retry-if predicate,
// per-attempt hook) live in the underlying retry.Retry instance and are
// not duplicated here.
type Options struct {
	retryOnResponse RetryOnResponseFn
}

// NewOptions creates Options with safe defaults and applies the provided
// functional options. Defaults:
//
//   - retryOnResponse: NoopRetryOnResponse (no response-based retries;
//     only transport errors triger the retrier).
func NewOptions(opts ...Option) *Options {
	options := &Options{
		retryOnResponse: NoopRetryOnResponse,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithRetryOnResponse sets the predicate that decides whether a
// successful response should be treated as a retryable failure. When the
// predicate returns true, the transport synthesizes a *StatusCodeError
// and returns it to the retrier; callers should configure the retrier
// with WithRetryIf(RetryIfHttpError) so the synthetic error triggers a
// retry. Nil values are ignored, preserving the default
// (NoopRetryOnResponse).
func WithRetryOnResponse(retryOnResponse RetryOnResponseFn) Option {
	return func(opts *Options) {
		if retryOnResponse != nil {
			opts.retryOnResponse = retryOnResponse
		}
	}
}
