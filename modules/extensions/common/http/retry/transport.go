package retry

import (
	"net/http"

	retrygo "github.com/avast/retry-go/v4"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// retryTransport wraps another http.RoundTripper and retries failed
// requests according to the configured strategy. Constructed via
// NewRetryTransport.
type retryTransport struct {
	base            http.RoundTripper
	attempts        uint
	retryIf         RetryIfFn
	retryHook       RetryHookFn
	retryOnResponse RetryOnResponseFn
}

// NewRetryTransport wraps the given base RoundTripper with retry logic
// configured through opts. A nil base falls back to http.DefaultTransport.
//
// The returned RoundTripper is safe for concurrent use as long as base
// and the configured callbacks are.
func NewRetryTransport(base http.RoundTripper, opts ...Option) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	options := NewOptions(opts...)

	return &retryTransport{
		base:            base,
		attempts:        options.attempts,
		retryIf:         options.retryIf,
		retryHook:       options.retryHook,
		retryOnResponse: options.retryOnResponse,
	}
}

// RoundTrip executes the request with retries. If the request has a
// non-replayable body (Body != nil and GetBody == nil), it returns
// ErrNonReplayableBody without attempting any request — retries would
// silently send a consumed body.
//
// Each attempt clones the request via req.Clone (which shares fields but
// not Body) and rebuilds Body from GetBody when present. When the
// configured RetryOnResponseFn returns true, the response body is closed
// and a *StatusCodeError is returned so the retry loop can match it.
//
// When attempts is exhausted, RoundTrip returns the last error wrapped
// in ErrRetry.
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cassert.NotNil(t, "retry transport receiver is nil")
	cassert.NotNil(req, "request is nil")

	if req.Body != nil && req.GetBody == nil {
		return nil, ErrNonReplayableBody()
	}

	attempt := func() (*http.Response, error) {
		clonedReq := req.Clone(req.Context())
		if req.Body != nil && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}

			clonedReq.Body = body
		}

		res, err := t.base.RoundTrip(clonedReq)
		if err != nil {
			return nil, err
		}

		if t.retryOnResponse(res) {
			_ = res.Body.Close()
			return nil, &StatusCodeError{StatusCode: res.StatusCode}
		}

		return res, nil
	}

	res, err := retrygo.DoWithData(attempt,
		retrygo.Context(req.Context()),
		retrygo.Attempts(t.attempts),
		retrygo.RetryIf(retrygo.RetryIfFunc(t.retryIf)),
		retrygo.OnRetry(retrygo.OnRetryFunc(t.retryHook)),
	)
	if err != nil {
		return nil, ErrRetry(err)
	}

	return res, nil
}
