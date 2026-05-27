package breaker

import (
	"errors"
	"net/http"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	rbreaker "github.com/guidomantilla/yarumo/extension/common/resilience/breaker"
)

// breakerTransport wraps another http.RoundTripper and gates each
// request through a resilience.Breaker. Only HTTP-specific behavior
// kept here is the synthetic *StatusCodeError that lets the breaker
// observe successful-but-bad responses as failures.
type breakerTransport struct {
	base           http.RoundTripper
	breaker        rbreaker.Breaker
	failOnResponse FailOnResponseFn
}

// NewBreakerTransport wraps the given base RoundTripper with the
// supplied breaker. A nil base falls back to http.DefaultTransport. A
// nil breaker is rejected at construction time — callers that want to
// disable the breaker should not wrap the transport in the first place.
//
// The breaker policy (consecutive failures threshold, open-state
// timeout, half-open probe budget, state-change hook) is owned entirely
// by the breaker; this transport adds only the HTTP-specific behavior
// configured through opts (currently: WithFailOnResponse).
//
// The returned RoundTripper is safe for concurrent use as long as base
// and breaker are.
func NewBreakerTransport(base http.RoundTripper, breaker rbreaker.Breaker, opts ...Option) http.RoundTripper {
	cassert.NotNil(breaker, "breaker is nil")

	if base == nil {
		base = http.DefaultTransport
	}

	options := NewOptions(opts...)

	return &breakerTransport{
		base:           base,
		breaker:        breaker,
		failOnResponse: options.failOnResponse,
	}
}

// RoundTrip executes the request through the breaker.
//
// When the breaker is open, Execute returns immediately with
// rbreaker.ErrBreakerOpen and the request is NOT sent. In half-open
// state beyond the probe budget the breaker returns
// rbreaker.ErrBreakerTooManyRequests. Both surface to the caller as an
// error wrapping ErrBreakerRejectedFailed.
//
// When the breaker admits the call:
//   - Transport errors are reported to the breaker as failures (so they
//     count toward the consecutive-failures threshold).
//   - When the configured FailOnResponseFn returns true on the response,
//     the response body is closed, a *StatusCodeError is synthesized and
//     returned to the breaker so the response counts as a failure.
//     RoundTrip then returns nil response + the *StatusCodeError to the
//     caller (the caller can inspect the status via errors.As).
//   - Otherwise the response is returned unchanged and the breaker
//     counts the call as a success.
func (t *breakerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cassert.NotNil(t, "breaker transport receiver is nil")
	cassert.NotNil(req, "request is nil")

	var res *http.Response

	err := t.breaker.Execute(req.Context(), func() error {
		var rtErr error
		res, rtErr = t.base.RoundTrip(req)
		if rtErr != nil {
			return rtErr
		}

		if t.failOnResponse(res) {
			_ = res.Body.Close()
			statusErr := &StatusCodeError{StatusCode: res.StatusCode}
			res = nil
			return statusErr
		}

		return nil
	})
	if err != nil {
		// Pass through the underlying status-code synthesis without
		// re-wrapping so callers can errors.As(err, &StatusCodeError).
		var statusErr *StatusCodeError
		if errors.As(err, &statusErr) {
			return nil, err
		}
		return nil, ErrBreakerRejected(err)
	}

	return res, nil
}
