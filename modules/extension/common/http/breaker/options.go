package breaker

// Option is a functional option for configuring breaker transport Options.
type Option func(opts *Options)

// Options holds the HTTP-specific configuration of the breaker transport.
// The breaker policy (consecutive failures threshold, timeout, probe
// budget, state-change hook) lives in the underlying
// resilience.Breaker instance and is not duplicated here.
type Options struct {
	failOnResponse FailOnResponseFn
}

// NewOptions creates Options with safe defaults and applies the provided
// functional options. Defaults:
//
//   - failOnResponse: NoopFailOnResponse (only transport errors count as
//     failures against the breaker).
func NewOptions(opts ...Option) *Options {
	options := &Options{
		failOnResponse: NoopFailOnResponse,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithFailOnResponse sets the predicate that decides whether a
// successful response should be reported to the breaker as a failure.
// When the predicate returns true, the transport closes the response
// body and returns a synthetic *StatusCodeError so the breaker counts
// the response toward its trip threshold. The synthetic error is then
// translated back into the response on the caller side (see RoundTrip
// docs for the reconstruction protocol). Nil values are ignored,
// preserving the default (NoopFailOnResponse).
func WithFailOnResponse(failOnResponse FailOnResponseFn) Option {
	return func(opts *Options) {
		if failOnResponse != nil {
			opts.failOnResponse = failOnResponse
		}
	}
}
