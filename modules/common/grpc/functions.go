package grpc

import (
	"context"
	"runtime"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	clog "github.com/guidomantilla/yarumo/common/log"
)

// RecoveryInterceptor returns a unary server interceptor that recovers from panics.
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			r := recover()
			if r != nil {
				stack := make([]byte, 4096)
				n := runtime.Stack(stack, false)

				clog.Error(ctx, "grpc handler panicked", "method", info.FullMethod, "panic", r, "stack", string(stack[:n]))

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
				stack := make([]byte, 4096)
				n := runtime.Stack(stack, false)

				clog.Error(ss.Context(), "grpc stream handler panicked", "method", info.FullMethod, "panic", r, "stack", string(stack[:n]))

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
