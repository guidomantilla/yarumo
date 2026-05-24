// Package retry provides an http.RoundTripper that retries failed requests
// per a configurable strategy. The transport delegates the actual retry
// loop to github.com/avast/retry-go/v4 and exposes a yarumo-flavored
// Options pattern on top.
//
// Compose it like any other RoundTripper:
//
//	import (
//	    chttp      "github.com/guidomantilla/yarumo/common/http"
//	    chttpretry "github.com/guidomantilla/yarumo/extensions/common/http/retry"
//	)
//
//	transport := chttpretry.NewRetryTransport(
//	    http.DefaultTransport,
//	    chttpretry.WithAttempts(3),
//	    chttpretry.WithRetryOnResponse(chttpretry.RetryOn5xxAnd429),
//	)
//	client := chttp.NewClient(chttp.WithTransport(transport))
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
	_ RetryIfFn         = NoopRetryIf
	_ RetryIfFn         = RetryIfHttpError
	_ RetryHookFn       = NoopRetryHook

	_ ErrRetryFn               = ErrRetry
	_ ErrNonReplayableBodyFn   = ErrNonReplayableBody
)

// RetryOnResponseFn decides whether a successful HTTP response should
// trigger a retry. Typical use: retry on 5xx server errors and 429
// throttling.
type RetryOnResponseFn func(res *http.Response) bool

// RetryIfFn decides whether an error returned by RoundTrip should trigger
// a retry. Typical use: retry on network-level errors and on responses
// wrapped as *StatusCodeError by the retry transport itself.
type RetryIfFn func(err error) bool

// RetryHookFn is called before each retry attempt with the attempt index
// (zero-based; 0 is the first retry, i.e. the second total attempt) and
// the error that triggered the retry.
type RetryHookFn func(attempt uint, err error)

// ErrRetryFn is the function type for ErrRetry.
type ErrRetryFn func(causes ...error) error

// ErrNonReplayableBodyFn is the function type for ErrNonReplayableBody.
type ErrNonReplayableBodyFn func(causes ...error) error
