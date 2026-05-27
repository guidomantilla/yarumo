// Package breaker provides an http.RoundTripper that wraps each request
// in a resilience.Breaker (from
// modules/extensions/common/resilience/breaker/). This module is a thin
// adapter: it accepts a pre-configured Breaker and lets the breaker
// decide whether to admit, reject (open state), or probe (half-open
// state) each RoundTrip.
//
// Compose it like any other RoundTripper:
//
//	import (
//	    chttp        "github.com/guidomantilla/yarumo/common/http"
//	    chttpbreaker "github.com/guidomantilla/yarumo/extensions/common/http/breaker"
//	    rbreaker     "github.com/guidomantilla/yarumo/extensions/common/resilience/breaker"
//	)
//
//	b := rbreaker.NewBreaker(
//	    rbreaker.WithConsecutiveFailures(5),
//	    rbreaker.WithTimeout(15*time.Second),
//	)
//
//	transport := chttpbreaker.NewBreakerTransport(
//	    http.DefaultTransport,
//	    b,
//	    chttpbreaker.WithFailOnResponse(chttpbreaker.FailOn5xxAnd429),
//	)
//	client := chttp.NewClient(chttp.WithTransport(transport))
//
// The HTTP-specific concerns kept here are:
//
//   - WithFailOnResponse: predicate that turns a successful response
//     (e.g. 5xx, 429) into a synthetic *StatusCodeError so the breaker
//     counts it as a failure.
//   - Translation of breaker domain sentinels (ErrBreakerOpen,
//     ErrBreakerTooManyRequests) back to the caller through the standard
//     error chain.
//
// Order in a middleware stack matters: when the breaker wraps retry
// (breaker outside), each retry attempt consumes a probe in half-open
// state — likely undesirable. When retry wraps breaker (retry outside),
// retries see a clean fail-fast from the breaker and can back off
// accordingly. The recommended outer-to-inner order is retry → breaker →
// limiter → base.
package breaker

import "net/http"

var (
	_ http.RoundTripper = (*breakerTransport)(nil)
	_ error             = (*StatusCodeError)(nil)
	_ error             = (*Error)(nil)

	_ FailOnResponseFn = NoopFailOnResponse
	_ FailOnResponseFn = FailOn5xxAnd429

	_ ErrBreakerRejectedFn = ErrBreakerRejected
)

// FailOnResponseFn decides whether a successful HTTP response should be
// reported to the breaker as a failure. Typical use: count 5xx server
// errors and 429 throttling against the consecutive-failures threshold so
// a sustained upstream incident trips the breaker. Implementations must
// be safe for concurrent use.
type FailOnResponseFn func(res *http.Response) bool

// ErrBreakerRejectedFn is the function type for ErrBreakerRejected.
type ErrBreakerRejectedFn func(causes ...error) error
