// Package http provides a high-level HTTP server abstraction over the
// stdlib net/http.Server with managed-component lifecycle semantics.
//
// Server is created via NewServer with a name, network, host, port and an
// http.Handler. It implements common/lifecycle.Component (Name + Start +
// Stop + Done) with server-style semantics: Start blocks calling
// Serve/ServeTLS until shutdown; Done closes when Start returns.
//
// Consumers wire the Server into the lifecycle pipeline via
// lifecycle.Build(ctx, server, errChan), which returns the CloseFn for
// graceful shutdown.
//
// Error contract: server operations wrap errors into a domain Error type
// with ServerType. Callers should prefer errors.Is/As instead of relying
// on string messages.
//
// Concurrency: Server implementations are safe for concurrent use by
// multiple goroutines.
package http

import (
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

var (
	_ Server = (*server)(nil)

	_ ErrServerFn = ErrServer
)

// ErrServerFn is the function type for ErrServer.
type ErrServerFn func(errs ...error) error

// Server defines the interface for an HTTP server with lifecycle support.
//
// The caller is responsible for calling Stop to release resources.
// Implementations must be safe for concurrent use by multiple goroutines.
type Server interface {
	lifecycle.Component
}
