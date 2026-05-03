// Package grpc provides a high-level gRPC server abstraction with built-in interceptors
// for panic recovery and request logging.
//
// The server is created via NewServer and configured through functional Options including
// service registration (WithService) and gRPC server options (WithServerOption).
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
	"net"

	"google.golang.org/grpc"
)

var (
	_ Server              = (*server)(nil)
	_ ErrServerFn         = ErrServer
	_ UnaryInterceptorFn  = RecoveryInterceptor
	_ UnaryInterceptorFn  = LoggingInterceptor
	_ StreamInterceptorFn = StreamRecoveryInterceptor
	_ StreamInterceptorFn = StreamLoggingInterceptor
)

// ErrServerFn is the function type for ErrServer.
type ErrServerFn func(errs ...error) error

// UnaryInterceptorFn is the function type for unary interceptor factories.
type UnaryInterceptorFn func() grpc.UnaryServerInterceptor

// StreamInterceptorFn is the function type for stream interceptor factories.
type StreamInterceptorFn func() grpc.StreamServerInterceptor

// Server defines the interface for a gRPC server.
//
// The caller is responsible for calling Stop or GracefulStop to release resources.
// Implementations must be safe for concurrent use by multiple goroutines.
type Server interface {
	grpc.ServiceRegistrar
	// Address returns the network address the server is configured to listen on.
	Address() string
	// Stop stops the gRPC server immediately.
	Stop()
	// GracefulStop stops the gRPC server gracefully.
	GracefulStop()
	// Serve accepts incoming connections on the listener.
	Serve(lis net.Listener) error
}
