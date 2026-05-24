// Package retry provides an http.RoundTripper that retries failed requests
// by delegating the retry loop to a generic retry.Retry instance from
// modules/extensions/common/resilience/retry/. This module is a thin
// adapter: it accepts a pre-configured retry policy and wraps each
// RoundTrip call in retrier.Do.
//
// Compose it like any other RoundTripper:
//
//	import (
//	    chttp      "github.com/guidomantilla/yarumo/common/http"
//	    chttpretry "github.com/guidomantilla/yarumo/extensions/common/http/retry"
//	    rretry     "github.com/guidomantilla/yarumo/extensions/common/resilience/retry"
//	)
//
//	r := rretry.NewRetry(
//	    rretry.WithAttempts(3),
//	    rretry.WithRetryIf(chttpretry.RetryIfHttpError),
//	)
//
//	transport := chttpretry.NewRetryTransport(
//	    http.DefaultTransport,
//	    r,
//	    chttpretry.WithRetryOnResponse(chttpretry.RetryOn5xxAnd429),
//	)
//	client := chttp.NewClient(chttp.WithTransport(transport))
//
// The HTTP-specific concerns kept here are:
//
//   - WithRetryOnResponse: predicate that turns a successful response
//     (e.g. 5xx, 429) into a synthetic *StatusCodeError so the retrier
//     observes it as a retryable failure.
//   - RetryIfHttpError: helper predicate that recognizes *StatusCodeError
//     for use with rretry.WithRetryIf.
//   - Replayable-body guard (ErrNonReplayableBody): refuses to retry
//     requests whose body can be consumed only once.
//
// Order matters: when the retry transport wraps the limiter transport
// (retry outside), each retry attempt consumes a token from the limiter.
// When the limiter wraps retry (limiter outside), only one token is
// consumed per request regardless of retries. See package chttp docs.
package retry

import "net/http"

var (
	_ http.RoundTripper = (*retryTransport)(nil)
	_ error             = (*StatusCodeError)(nil)
	_ error             = (*Error)(nil)

	_ RetryOnResponseFn = NoopRetryOnResponse
	_ RetryOnResponseFn = RetryOn5xxAnd429

	_ ErrNonReplayableBodyFn = ErrNonReplayableBody
)

// RetryOnResponseFn decides whether a successful HTTP response should
// trigger a retry. Typical use: retry on 5xx server errors and 429
// throttling. Implementations must be safe for concurrent use.
type RetryOnResponseFn func(res *http.Response) bool

// ErrNonReplayableBodyFn is the function type for ErrNonReplayableBody.
type ErrNonReplayableBodyFn func(causes ...error) error
