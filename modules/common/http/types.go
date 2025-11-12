// Package http provides a small wrapper around the standard *http.Client
// that optionally adds rate limiting and retry capabilities while keeping
// full API compatibility through a minimal Client interface.
//
// Error contract: implementations may wrap underlying errors. Callers should
// prefer errors.Is/As instead of relying on string messages.
// Responsibility: the caller must close res.Body when err == nil.
// Concurrency: implementations must be safe for concurrent use by multiple goroutines.
package http

import "net/http"

var (
	_ Client = (*http.Client)(nil)
	_ Client = (*client)(nil)
	_ Client = (*MockClient)(nil)
)

// Client represents the minimal contract compatible with *http.Client.
//
// Semantics:
//   - Do sends the HTTP request and returns the response or an error.
//   - Cancellation and deadlines are controlled via req.Context().
//   - Implementations may apply internal policies (e.g., rate limiting,
//     retries, instrumentation) without changing responsibilities.
//   - The caller is responsible for closing res.Body when err == nil.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}
