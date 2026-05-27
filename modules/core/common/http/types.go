// Package http provides the base abstractions for the workspace's HTTP
// client stack: the Client interface, a NewClient convenience that wires
// stdlib *http.Client with safe defaults, the RoundTripperFn adapter for
// turning plain functions into http.RoundTripper, and a domain error
// type that other transports can wrap.
//
// All concrete middleware transports (rate limiting, retry, tracing,
// metrics, auth, ...) live in their own modules under
// modules/extension/common/http/<name>/ so consumers pull only what
// they use. Each is an http.RoundTripper that wraps another, and the
// caller composes the chain explicitly:
//
//	import (
//	    chttp        "github.com/guidomantilla/yarumo/core/common/http"
//	    chttplimiter "github.com/guidomantilla/yarumo/extension/common/http/limiter"
//	    chttpretry   "github.com/guidomantilla/yarumo/extension/common/http/retry"
//	)
//
//	transport := chttpretry.NewRetryTransport(
//	    chttplimiter.NewLimiterTransport(http.DefaultTransport, rate.NewLimiter(10, 5)),
//	    chttpretry.WithAttempts(3),
//	)
//	client := chttp.NewClient(chttp.WithTransport(transport))
package http

import "net/http"

var (
	_ http.RoundTripper = RoundTripperFn(nil)
	_ Client            = (*http.Client)(nil)

	_ ErrTransportFn = ErrTransport
)

// Client is the minimal contract for executing HTTP requests. It is
// satisfied by the stdlib *http.Client out of the box, so consumers that
// need an abstraction (for testing or alternative implementations) can
// accept Client and still pass the canonical type. Mocks typically
// implement Client directly or wrap *http.Client with a fake
// http.RoundTripper, which is the idiomatic Go testing pattern.
type Client interface {
	// Do executes the request and returns the response. The caller must
	// close res.Body when err == nil. Implementations must be safe for
	// concurrent use by multiple goroutines.
	Do(req *http.Request) (*http.Response, error)
}

// ErrTransportFn is the function type for ErrTransport.
type ErrTransportFn func(causes ...error) error
