// Package http provides a high-level HTTP server abstraction over the
// stdlib net/http.Server with managed-component lifecycle semantics.
//
// Server is created via NewServer with a name, network, host, port and an
// http.Handler. It implements common/lifecycle.Component (Name + Start +
// Stop + Done) with server-style semantics: Start blocks calling
// Serve/ServeTLS until shutdown; Done closes when Start returns.
//
// BuildServer wraps NewServer with the managed-component idiom: it
// returns the Server together with a lifecycle.CloseFn that performs
// graceful shutdown bounded by a timeout and blocks until the background
// goroutine has exited.
//
// Error contract: server operations wrap errors into a domain Error type
// with ServerType. Callers should prefer errors.Is/As instead of relying
// on string messages.
//
// Concurrency: Server implementations are safe for concurrent use by
// multiple goroutines.
package http

import (
	"context"
	nethttp "net/http"

	"github.com/guidomantilla/yarumo/common/lifecycle"
)

var (
	_ Server = (*server)(nil)

	_ BuildServerFn = BuildServer
	_ ErrServerFn   = ErrServer
)

// BuildServerFn is the function type for BuildServer.
type BuildServerFn func(ctx context.Context, name string, network string, host string, port string, handler nethttp.Handler, errChan lifecycle.ErrChan, options ...Option) (Server, lifecycle.CloseFn, error)

// ErrServerFn is the function type for ErrServer.
type ErrServerFn func(errs ...error) error

// Server defines the interface for an HTTP server with lifecycle support.
//
// The caller is responsible for calling Stop to release resources.
// Implementations must be safe for concurrent use by multiple goroutines.
type Server interface {
	lifecycle.Component
}
