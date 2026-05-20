// Package http provides a thin wrapper around the standard *http.Server with
// secure defaults for timeouts and header sizes, configurable via Options.
//
// This package hosts only the server side of the HTTP abstraction; the client
// side (rate-limited, retrying *http.Client wrapper) lives in
// modules/common/http and stays in modules/common/ because it is stateless.
// The server lives here because it carries lifecycle (listener goroutines,
// graceful shutdown), which violates modules/common/'s no-lifecycle clause.
//
// Concurrency: Server implementations are safe for concurrent use by multiple
// goroutines. The caller is responsible for calling Shutdown or Close to
// release resources.
package http

import (
	"context"
)

var (
	_ Server      = (*server)(nil)
	_ ErrServerFn = ErrServer
)

// ErrServerFn is the function type for ErrServer.
type ErrServerFn func(errs ...error) error

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
