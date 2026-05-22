// Package grpc provides a high-level gRPC server abstraction with built-in interceptors
// for panic recovery and request logging.
//
// The server is created via NewServer and configured through functional Options including
// service registration (WithService) and gRPC server options (WithServerOption). It
// implements common/lifecycle.Component (Name + Start + Stop + Done), with Start
// blocking until shutdown (server-style lifecycle).
//
// Built-in interceptors:
//   - RecoveryInterceptor / StreamRecoveryInterceptor: recover from handler panics and return codes.Internal.
//   - LoggingInterceptor / StreamLoggingInterceptor: log method name, duration, and errors.
//
// Error contract: server operations wrap errors into a domain Error type with ServerType.
// Callers should prefer errors.Is/As instead of relying on string messages.
//
// Concurrency: Server implementations are safe for concurrent use by multiple goroutines.
package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/guidomantilla/yarumo/common/lifecycle"
)

var (
	_ Server = (*server)(nil)

	_ BuildServerFn       = BuildServer
	_ ErrServerFn         = ErrServer
	_ UnaryInterceptorFn  = RecoveryInterceptor
	_ UnaryInterceptorFn  = LoggingInterceptor
	_ StreamInterceptorFn = StreamRecoveryInterceptor
	_ StreamInterceptorFn = StreamLoggingInterceptor
)

// BuildServerFn is the function type for BuildServer.
type BuildServerFn func(ctx context.Context, name string, network string, host string, port string, errChan lifecycle.ErrChan, options ...Option) (Server, lifecycle.CloseFn, error)

// ErrServerFn is the function type for ErrServer.
type ErrServerFn func(errs ...error) error

// UnaryInterceptorFn is the function type for unary interceptor factories.
type UnaryInterceptorFn func() grpc.UnaryServerInterceptor

// StreamInterceptorFn is the function type for stream interceptor factories.
type StreamInterceptorFn func() grpc.StreamServerInterceptor

// Server defines the interface for a gRPC server with lifecycle support.
//
// The caller is responsible for calling Stop to release resources.
// Implementations must be safe for concurrent use by multiple goroutines.
type Server interface {
	lifecycle.Component
	grpc.ServiceRegistrar
}
