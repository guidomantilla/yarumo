package retry

// Option is a functional option for configuring retry Options.
type Option func(opts *Options)

// Options holds the configuration for the retry transport. Fields are
// unexported; callers configure them through the With* functions.
type Options struct {
	attempts        uint
	retryIf         RetryIfFn
	retryHook       RetryHookFn
	retryOnResponse RetryOnResponseFn

	retryIfSet         bool
	retryOnResponseSet bool
}

// NewOptions creates Options with safe defaults and applies the provided
// functional options. Defaults:
//
//   - attempts: 3 (1 original + 2 retries)
//   - retryIf: NoopRetryIf (no error-based retries)
//   - retryHook: NoopRetryHook (no per-attempt hook)
//   - retryOnResponse: NoopRetryOnResponse (no response-based retries)
//
// Auto-wire: when WithRetryOnResponse is configured but WithRetryIf is
// not, retryIf is set to RetryIfHttpError so response-based retries
// actually trigger the retry loop (the response is converted to a
// *StatusCodeError that RetryIfHttpError matches).
func NewOptions(opts ...Option) *Options {
	options := &Options{
		attempts:        3,
		retryIf:         NoopRetryIf,
		retryHook:       NoopRetryHook,
		retryOnResponse: NoopRetryOnResponse,
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.retryOnResponseSet && !options.retryIfSet {
		options.retryIf = RetryIfHttpError
	}

	return options
}

// WithAttempts sets the total number of attempts (1 original + N-1 retries).
// Values less than 2 are ignored, preserving the default.
func WithAttempts(attempts uint) Option {
	return func(opts *Options) {
		if attempts > 1 {
			opts.attempts = attempts
		}
	}
}

// WithRetryIf sets the predicate that decides whether an error should
// trigger a retry. Nil values are ignored.
func WithRetryIf(retryIf RetryIfFn) Option {
	return func(opts *Options) {
		if retryIf != nil {
			opts.retryIf = retryIf
			opts.retryIfSet = true
		}
	}
}

// WithRetryHook sets the hook invoked before each retry attempt. Nil
// values are ignored.
func WithRetryHook(retryHook RetryHookFn) Option {
	return func(opts *Options) {
		if retryHook != nil {
			opts.retryHook = retryHook
		}
	}
}

// WithRetryOnResponse sets the predicate that decides whether a
// successful response should be treated as a retryable failure. When
// set, the transport synthesizes a *StatusCodeError and the retry loop
// picks it up via retryIf (auto-wired to RetryIfHttpError when not
// explicitly configured). Nil values are ignored.
func WithRetryOnResponse(retryOnResponse RetryOnResponseFn) Option {
	return func(opts *Options) {
		if retryOnResponse != nil {
			opts.retryOnResponse = retryOnResponse
			opts.retryOnResponseSet = true
		}
	}
}
