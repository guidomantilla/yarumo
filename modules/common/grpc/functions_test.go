package grpc

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeServerStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (f *fakeServerStream) Context() context.Context {
	return f.ctx
}

func TestRecoveryInterceptor(t *testing.T) {
	t.Parallel()

	interceptor := RecoveryInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	t.Run("passes through on success", func(t *testing.T) {
		t.Parallel()

		handler := func(_ context.Context, _ any) (any, error) {
			return "ok", nil
		}

		resp, err := interceptor(context.Background(), nil, info, handler)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if resp != "ok" {
			t.Fatalf("expected response %q, got %v", "ok", resp)
		}
	})

	t.Run("passes through handler error", func(t *testing.T) {
		t.Parallel()

		handlerErr := errors.New("handler failed")
		handler := func(_ context.Context, _ any) (any, error) {
			return nil, handlerErr
		}

		resp, err := interceptor(context.Background(), nil, info, handler)
		if !errors.Is(err, handlerErr) {
			t.Fatalf("expected handler error, got %v", err)
		}

		if resp != nil {
			t.Fatalf("expected nil response, got %v", resp)
		}
	})

	t.Run("recovers from panic", func(t *testing.T) {
		t.Parallel()

		handler := func(_ context.Context, _ any) (any, error) {
			panic("test panic")
		}

		resp, err := interceptor(context.Background(), nil, info, handler)
		if err == nil {
			t.Fatal("expected error after panic")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected gRPC status error, got %T", err)
		}

		if st.Code() != codes.Internal {
			t.Fatalf("expected codes.Internal, got %v", st.Code())
		}

		if resp != nil {
			t.Fatalf("expected nil response, got %v", resp)
		}
	})
}

func TestStreamRecoveryInterceptor(t *testing.T) {
	t.Parallel()

	interceptor := StreamRecoveryInterceptor()
	info := &grpc.StreamServerInfo{FullMethod: "/test.Service/StreamMethod"}

	t.Run("passes through on success", func(t *testing.T) {
		t.Parallel()

		handler := func(_ any, _ grpc.ServerStream) error {
			return nil
		}

		ss := &fakeServerStream{ctx: context.Background()}

		err := interceptor(nil, ss, info, handler)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("passes through handler error", func(t *testing.T) {
		t.Parallel()

		handlerErr := errors.New("stream handler failed")
		handler := func(_ any, _ grpc.ServerStream) error {
			return handlerErr
		}

		ss := &fakeServerStream{ctx: context.Background()}

		err := interceptor(nil, ss, info, handler)
		if !errors.Is(err, handlerErr) {
			t.Fatalf("expected handler error, got %v", err)
		}
	})

	t.Run("recovers from panic", func(t *testing.T) {
		t.Parallel()

		handler := func(_ any, _ grpc.ServerStream) error {
			panic("stream panic")
		}

		ss := &fakeServerStream{ctx: context.Background()}

		err := interceptor(nil, ss, info, handler)
		if err == nil {
			t.Fatal("expected error after panic")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected gRPC status error, got %T", err)
		}

		if st.Code() != codes.Internal {
			t.Fatalf("expected codes.Internal, got %v", st.Code())
		}
	})
}

func TestLoggingInterceptor(t *testing.T) {
	t.Parallel()

	interceptor := LoggingInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	t.Run("logs success", func(t *testing.T) {
		t.Parallel()

		handler := func(_ context.Context, _ any) (any, error) {
			return "ok", nil
		}

		resp, err := interceptor(context.Background(), nil, info, handler)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if resp != "ok" {
			t.Fatalf("expected %q, got %v", "ok", resp)
		}
	})

	t.Run("logs error", func(t *testing.T) {
		t.Parallel()

		handlerErr := errors.New("fail")
		handler := func(_ context.Context, _ any) (any, error) {
			return nil, handlerErr
		}

		resp, err := interceptor(context.Background(), nil, info, handler)
		if !errors.Is(err, handlerErr) {
			t.Fatalf("expected handler error, got %v", err)
		}

		if resp != nil {
			t.Fatalf("expected nil response, got %v", resp)
		}
	})
}

func TestStreamLoggingInterceptor(t *testing.T) {
	t.Parallel()

	interceptor := StreamLoggingInterceptor()
	info := &grpc.StreamServerInfo{FullMethod: "/test.Service/StreamMethod"}

	t.Run("logs success", func(t *testing.T) {
		t.Parallel()

		handler := func(_ any, _ grpc.ServerStream) error {
			return nil
		}

		ss := &fakeServerStream{ctx: context.Background()}

		err := interceptor(nil, ss, info, handler)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("logs error", func(t *testing.T) {
		t.Parallel()

		handlerErr := errors.New("stream fail")
		handler := func(_ any, _ grpc.ServerStream) error {
			return handlerErr
		}

		ss := &fakeServerStream{ctx: context.Background()}

		err := interceptor(nil, ss, info, handler)
		if !errors.Is(err, handlerErr) {
			t.Fatalf("expected handler error, got %v", err)
		}
	})
}
