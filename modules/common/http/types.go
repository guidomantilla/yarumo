// Package http provides a small wrapper around the standard *http.Client and *http.Server.
//
// This custom implementation/extension of *http.Client optionally adds rate limiting and retry capabilities while keeping
// full API compatibility through a minimal Client interface.
//
// Timeouts alignment:
//   - The underlying http.Client.Timeout is configurable via Options and acts as an overall per-request timeout.
//   - The internal rate limiter wait is bounded by the effective deadline, which is the minimum between req.Context() deadline and the client Timeout.
//   - When a *http.Transport is provided, selected transport timeouts are capped to not exceed the client Timeout
//     (TLSHandshakeTimeout, ResponseHeaderTimeout, ExpectContinueTimeout). Stricter values provided by the transport are kept.
//
// Error contract: implementations may wrap underlying errors. Callers should prefer errors.Is/As instead of relying on string messages.
// Responsibility: the caller must close res.Body when err == nil.
// Concurrency: implementations must be safe for concurrent use by multiple goroutines.
//
// This custom implementation/extension of *http.Server ??
package http

import (
	"context"
	"net/http"
)

var (
	_ Client = (*client)(nil)
	_ Client = (*PluggableClient)(nil)
)

// Client represents the minimal contract compatible with *http.Client.
//
// Semantics:
//   - Do sends the HTTP request and returns the response or an error.
//   - Cancellation and deadlines are controlled via req.Context().
//   - Implementations may apply internal policies (e.g., rate limiting, retries, instrumentation) without changing responsibilities.
//   - The caller is responsible for closing res.Body when err == nil.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
	LimiterEnabled() bool
	RetrierEnabled() bool
}

type Server interface {
	ListenAndServe() error
	ListenAndServeTLS(certFile string, keyFile string) error
	Shutdown(ctx context.Context) error
	Close() error
}
