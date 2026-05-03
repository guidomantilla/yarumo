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
// This custom implementation/extension of *http.Server provides a thin wrapper with secure defaults for timeouts and
// header sizes, configurable via Options.
//
// Error contract: implementations may wrap underlying errors. Callers should prefer errors.Is/As instead of relying on string messages.
// Responsibility: the caller must close res.Body when err == nil.
// Concurrency: implementations must be safe for concurrent use by multiple goroutines.
package http

import (
	"context"
	"net/http"

	retry "github.com/avast/retry-go/v4"
)

var (
	_ Client = (*client)(nil)
	_ Client = (*PluggableClient)(nil)
	_ Server = (*server)(nil)

	_ ErrDoFn           = ErrDo
	_ DoFn              = ErrorDo
	_ DoFn              = NoopDo
	_ DoFn              = Do
	_ LimiterEnabledFn  = EnabledLimiter
	_ LimiterEnabledFn  = DisabledLimiter
	_ RetrierEnabledFn  = EnabledRetrier
	_ RetrierEnabledFn  = DisabledRetrier
	_ retry.RetryIfFunc = NoopRetryIf
	_ retry.RetryIfFunc = RetryIfHttpError
	_ retry.OnRetryFunc = NoopRetryHook
	_ RetryOnResponseFn = NoopRetryOnResponse
	_ RetryOnResponseFn = RetryOn5xxAnd429Response

	_ DoFn             = DefaultClient.Do
	_ LimiterEnabledFn = DefaultClient.LimiterEnabled
	_ RetrierEnabledFn = DefaultClient.RetrierEnabled

	_ DoFn             = NoopClient.Do
	_ LimiterEnabledFn = NoopClient.LimiterEnabled
	_ RetrierEnabledFn = NoopClient.RetrierEnabled

	_ DoFn             = ErrorClient.Do
	_ LimiterEnabledFn = ErrorClient.LimiterEnabled
	_ RetrierEnabledFn = ErrorClient.RetrierEnabled
)

// ErrDoFn is the function type for ErrDo.
type ErrDoFn func(errs ...error) error

// DoFn is the function type for Do.
type DoFn func(req *http.Request) (*http.Response, error)

// RetryOnResponseFn is the function type for RetryOn5xxAnd429Response.
type RetryOnResponseFn func(res *http.Response) bool

// LimiterEnabledFn is the function type for LimiterEnabled.
type LimiterEnabledFn func() bool

// RetrierEnabledFn is the function type for RetrierEnabled.
type RetrierEnabledFn func() bool

// Client defines the interface for HTTP client operations.
//
// Cancellation and deadlines are controlled via req.Context().
// The caller is responsible for closing res.Body when err == nil.
// Implementations must be safe for concurrent use by multiple goroutines.
type Client interface {
	// Do sends the HTTP request and returns the response or an error.
	Do(req *http.Request) (*http.Response, error)
	// LimiterEnabled reports whether client-side rate limiting is active.
	LimiterEnabled() bool
	// RetrierEnabled reports whether automatic retries are active.
	RetrierEnabled() bool
}

// Server defines the interface for HTTP server lifecycle management.
//
// The caller is responsible for calling Shutdown or Close to release resources.
// Implementations must be safe for concurrent use by multiple goroutines.
type Server interface {
	// Address returns the network address the server is configured to listen on.
	Address() string
	// ListenAndServe starts the server on the configured address.
	ListenAndServe() error
	// ListenAndServeTLS starts the server with TLS using the provided certificate and key files.
	ListenAndServeTLS(certFile string, keyFile string) error
	// Shutdown gracefully shuts down the server without interrupting active connections.
	Shutdown(ctx context.Context) error
	// Close immediately closes the server.
	Close() error
}
