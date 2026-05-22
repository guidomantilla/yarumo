package grpc

import (
	"context"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/guidomantilla/yarumo/common/lifecycle"
	clog "github.com/guidomantilla/yarumo/common/log"
)

// BuildServer creates a managed gRPC Server, starts it in a background
// goroutine, and returns a CloseFn for graceful shutdown.
//
// Startup errors are logged and forwarded to errChan (non-blocking). The
// returned Server can be used to register additional services via
// RegisterService. The returned CloseFn must be called by the caller to
// release resources; it bounds the shutdown by the given timeout and
// blocks until the background goroutine has exited.
func BuildServer(ctx context.Context, name string, network string, host string, port string, errChan lifecycle.ErrChan, options ...Option) (Server, lifecycle.CloseFn, error) {
	clog.Info(ctx, "starting up", "stage", "startup", "component", name)

	component := NewServer(name, network, host, port, options...)

	spawned := make(chan struct{})

	closeFn := func(ctx context.Context, timeout time.Duration) {
		clog.Info(ctx, "stopping", "stage", "shutdown", "component", name)
		defer clog.Info(ctx, "stopped", "stage", "shutdown", "component", name)

		err := lifecycle.Stop(ctx, component, timeout)
		if err != nil {
			clog.Error(ctx, "shutdown failed", "stage", "shutdown", "component", name, "error", err)
		}

		<-spawned
	}

	go func() {
		defer close(spawned)

		err := lifecycle.Start(ctx, component, errChan)
		if err != nil {
			clog.Error(ctx, "failed to start", "stage", "startup", "component", name, "error", err)
		}
	}()

	return component, closeFn, nil
}

// RecoveryInterceptor returns a unary server interceptor that recovers from panics.
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			r := recover()
			if r != nil {
				clog.Error(ctx, "grpc handler panicked", "method", info.FullMethod, "panic", r, "stack", string(debug.Stack()))

				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

// StreamRecoveryInterceptor returns a stream server interceptor that recovers from panics.
func StreamRecoveryInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			r := recover()
			if r != nil {
				clog.Error(ss.Context(), "grpc stream handler panicked", "method", info.FullMethod, "panic", r, "stack", string(debug.Stack()))

				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(srv, ss)
	}
}

// LoggingInterceptor returns a unary server interceptor that logs request method and duration.
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		if err != nil {
			clog.Error(ctx, "grpc request failed", "method", info.FullMethod, "duration", duration, "error", err)
		} else {
			clog.Info(ctx, "grpc request completed", "method", info.FullMethod, "duration", duration)
		}

		return resp, err
	}
}

// StreamLoggingInterceptor returns a stream server interceptor that logs request method and duration.
func StreamLoggingInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		err := handler(srv, ss)

		duration := time.Since(start)

		if err != nil {
			clog.Error(ss.Context(), "grpc stream request failed", "method", info.FullMethod, "duration", duration, "error", err)
		} else {
			clog.Info(ss.Context(), "grpc stream request completed", "method", info.FullMethod, "duration", duration)
		}

		return err
	}
}
