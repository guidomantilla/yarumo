// Package limiter provides an http.RoundTripper that gates outgoing
// requests through a golang.org/x/time/rate.Limiter. Compose it with any
// base RoundTripper:
//
//	import (
//	    chttp        "github.com/guidomantilla/yarumo/core/common/http"
//	    chttplimiter "github.com/guidomantilla/yarumo/extension/common/http/limiter"
//	)
//
//	transport := chttplimiter.NewLimiterTransport(
//	    http.DefaultTransport,
//	    rate.NewLimiter(10, 5),
//	)
//	client := chttp.NewClient(chttp.WithTransport(transport))
//
// Composition order matters when stacking with the retry transport — see
// the package docs in modules/extension/common/http/retry/ for the
// semantics of limiter-outside-retry vs limiter-inside-retry.
package limiter

import (
	"net/http"
)

var (
	_ http.RoundTripper = (*limiterTransport)(nil)

	_ ErrRateLimiterExceededFn = ErrRateLimiterExceeded
)

// ErrRateLimiterExceededFn is the function type for ErrRateLimiterExceeded.
type ErrRateLimiterExceededFn func(causes ...error) error
