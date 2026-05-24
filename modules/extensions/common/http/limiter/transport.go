package limiter

import (
	"net/http"

	"golang.org/x/time/rate"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// limiterTransport gates each outgoing request on rate.Limiter.Wait
// before delegating to base. Constructed via NewLimiterTransport.
type limiterTransport struct {
	base    http.RoundTripper
	limiter *rate.Limiter
}

// NewLimiterTransport wraps the given base RoundTripper with a rate
// limiter that blocks each RoundTrip call until the provided Limiter
// grants a token (or the request context expires).
//
// A nil base falls back to http.DefaultTransport. A nil limiter is
// rejected at construction time — callers that want to disable gating
// should not wrap the transport in the first place; passing nil here
// almost always indicates a wiring bug. Use rate.NewLimiter with the
// desired rate and burst to construct the limiter.
//
// The returned RoundTripper is safe for concurrent use as long as base
// and limiter are.
func NewLimiterTransport(base http.RoundTripper, limiter *rate.Limiter) http.RoundTripper {
	cassert.NotNil(limiter, "limiter is nil")

	if base == nil {
		base = http.DefaultTransport
	}

	return &limiterTransport{
		base:    base,
		limiter: limiter,
	}
}

// RoundTrip waits for a token from limiter and then delegates to base.
// Returns an error wrapping ErrRateLimiterExceeded when the context
// expires while waiting for a token.
func (t *limiterTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cassert.NotNil(t, "limiter transport receiver is nil")
	cassert.NotNil(req, "request is nil")

	err := t.limiter.Wait(req.Context())
	if err != nil {
		return nil, ErrRateLimiterExceeded(err)
	}

	return t.base.RoundTrip(req)
}
