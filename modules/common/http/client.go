package http

import "net/http"

// NewClient returns a Client wired with the given options. The concrete
// type is the stdlib *http.Client (which satisfies Client via Do); the
// interface return follows the workspace convention so consumers can
// depend on the abstraction and substitute test doubles without touching
// production code.
//
// NewClient does not compose any middleware on its own. Whatever
// RoundTripper the caller passes via WithTransport is what the client
// uses verbatim, so the order of any limiter/retry/tracing wrappers is
// always explicit at the call site:
//
//	transport := chttpretry.NewRetryTransport(
//	    chttp.NewLimiterTransport(http.DefaultTransport, rate.NewLimiter(10, 5)),
//	    chttpretry.WithAttempts(3),
//	)
//	client := chttp.NewClient(chttp.WithTransport(transport))
//
// Defaults applied: a 30s overall timeout and, when the configured
// transport is a *http.Transport, internal-timeout capping to the
// overall timeout (see NewOptions).
func NewClient(opts ...Option) Client {
	options := NewOptions(opts...)

	return &http.Client{
		Transport: options.transport,
		Timeout:   options.timeout,
	}
}
