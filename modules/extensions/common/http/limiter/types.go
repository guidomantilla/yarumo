// Package limiter provides an http.RoundTripper that gates outgoing
// requests through a golang.org/x/time/rate.Limiter. Compose it with any
// base RoundTripper:
//
//	import (
//	    chttp        "github.com/guidomantilla/yarumo/common/http"
//	    chttplimiter "github.com/guidomantilla/yarumo/extensions/common/http/limiter"
//	)
//
//	transport := chttplimiter.NewLimiterTransport(
//	    http.DefaultTransport,
//	    rate.NewLimiter(10, 5),
//	)
//	client := chttp.NewClient(chttp.WithTransport(transport))
//
// Composition order matters when stacking with the retry transport — see
// the package docs in modules/extensions/common/http/retry/ for the
// semantics of limiter-outside-retry vs limiter-inside-retry.
package limiter

import (
	"net/http"

	"golang.org/x/time/rate"
)

var (
	_ http.RoundTripper = (*limiterTransport)(nil)

	_ NewLimiterTransportFn = NewLimiterTransport

	_ ErrRateLimiterExceededFn = ErrRateLimiterExceeded
)

// NewLimiterTransportFn is the function type for NewLimiterTransport.
type NewLimiterTransportFn func(base http.RoundTripper, limiter *rate.Limiter) http.RoundTripper

// ErrRateLimiterExceededFn is the function type for ErrRateLimiterExceeded.
type ErrRateLimiterExceededFn func(causes ...error) error
