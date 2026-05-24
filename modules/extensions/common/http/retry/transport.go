package retry

import (
	"net/http"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	rretry "github.com/guidomantilla/yarumo/extensions/common/resilience/retry"
)

// retryTransport wraps another http.RoundTripper and delegates the retry
// loop to the provided resilience.Retry. The only HTTP-specific behavior
// kept here is the request-clone-and-replay-body logic and the synthetic
// *StatusCodeError emitted when retryOnResponse returns true.
type retryTransport struct {
	base            http.RoundTripper
	retrier         rretry.Retry
	retryOnResponse RetryOnResponseFn
}

// NewRetryTransport wraps the given base RoundTripper with retry logic
// driven by the supplied retrier. A nil base falls back to
// http.DefaultTransport. A nil retrier is rejected at construction time —
// callers that want to disable retries should not wrap the transport in
// the first place.
//
// The retry policy (attempts, delay, backoff, retry-if predicate,
// per-attempt hook) is owned entirely by retrier; this transport adds
// only the HTTP-specific behavior configured through opts (currently:
// WithRetryOnResponse).
//
// The returned RoundTripper is safe for concurrent use as long as base
// and retrier are.
func NewRetryTransport(base http.RoundTripper, retrier rretry.Retry, opts ...Option) http.RoundTripper {
	cassert.NotNil(retrier, "retrier is nil")

	if base == nil {
		base = http.DefaultTransport
	}

	options := NewOptions(opts...)

	return &retryTransport{
		base:            base,
		retrier:         retrier,
		retryOnResponse: options.retryOnResponse,
	}
}

// RoundTrip executes the request through the retrier. If the request has
// a non-replayable body (Body != nil and GetBody == nil), it returns
// ErrNonReplayableBody synchronously — retries would silently send a
// consumed body.
//
// Each attempt clones the request via req.Clone (which shares fields but
// not Body) and rebuilds Body from GetBody when present. When the
// configured RetryOnResponseFn returns true, the response body is closed
// and a *StatusCodeError is returned from the closure so the retrier
// observes it as a retryable failure (when configured with
// WithRetryIf(RetryIfHttpError)).
//
// When the retrier gives up, RoundTrip returns the error it produced
// (typically wrapped as rretry.ErrRetryFailed with the last underlying
// error in the chain).
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cassert.NotNil(t, "retry transport receiver is nil")
	cassert.NotNil(req, "request is nil")

	if req.Body != nil && req.GetBody == nil {
		return nil, ErrNonReplayableBody()
	}

	var res *http.Response

	err := t.retrier.Do(req.Context(), func() error {
		clonedReq := req.Clone(req.Context())
		if req.Body != nil && req.GetBody != nil {
			body, getBodyErr := req.GetBody()
			if getBodyErr != nil {
				return getBodyErr
			}
			clonedReq.Body = body
		}

		var rtErr error
		res, rtErr = t.base.RoundTrip(clonedReq)
		if rtErr != nil {
			return rtErr
		}

		if t.retryOnResponse(res) {
			_ = res.Body.Close()
			return &StatusCodeError{StatusCode: res.StatusCode}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
