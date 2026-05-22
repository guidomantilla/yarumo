package grpc

import (
	"context"
	"errors"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	clog "github.com/guidomantilla/yarumo/common/log"
	cslog "github.com/guidomantilla/yarumo/log/slog"
)

type fakeServerStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (f *fakeServerStream) Context() context.Context {
	return f.ctx
}

var _ clog.Logger = (*stackCapturingLogger)(nil)

// stackCapturingLogger records the "stack" attribute from Error calls.
// Used to validate that panic recovery interceptors capture the full
// goroutine stack rather than a truncated 4 KiB buffer.
type stackCapturingLogger struct {
	stack string
}

func (l *stackCapturingLogger) Trace(_ context.Context, _ string, _ ...any) {}

func (l *stackCapturingLogger) Debug(_ context.Context, _ string, _ ...any) {}

func (l *stackCapturingLogger) Info(_ context.Context, _ string, _ ...any) {}

func (l *stackCapturingLogger) Warn(_ context.Context, _ string, _ ...any) {}

func (l *stackCapturingLogger) Fatal(_ context.Context, _ string, _ ...any) {}

func (l *stackCapturingLogger) Error(_ context.Context, _ string, args ...any) {
	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}

		if key != "stack" {
			continue
		}

		value, ok := args[i+1].(string)
		if !ok {
			continue
		}

		l.stack = value
	}
}

// outermostFrameSentinel_YA0037 is a sentinel function whose name appears in
// the panic stack trace below the recursion frames. With ≥10 frames of
// recursion plus all the deferred + gRPC + test infrastructure frames, the
// stack output exceeds 4 KiB, and this sentinel sits past the truncation
// boundary that the old fixed-buffer implementation would have applied.
func outermostFrameSentinel_YA0037(fn func()) {
	fn()
}

// recurseThenPanic_YA0037 produces a controlled depth of stack frames and
// then panics, simulating a deeply-nested handler.
func recurseThenPanic_YA0037(depth int) {
	if depth <= 0 {
		panic("deep panic")
	}

	recurseThenPanic_YA0037(depth - 1)
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

// TestRecoveryInterceptor_StackCaptureNotTruncated validates that recovery
// interceptors capture the full goroutine stack on panic, not a truncated
// 4 KiB prefix. It is the regression test for YA-0037.
//
// The test does not call t.Parallel because it mutates the package-global
// logger via clog.Use. The same applies to the subtests.
func TestRecoveryInterceptor_StackCaptureNotTruncated(t *testing.T) {
	const recursionDepth = 100

	t.Run("unary interceptor captures outermost recursion frames", func(t *testing.T) {
		spy := &stackCapturingLogger{}

		clog.Use(spy)

		defer clog.Use(cslog.NewLogger())

		interceptor := RecoveryInterceptor()
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/DeepPanic"}

		handler := func(_ context.Context, _ any) (any, error) {
			outermostFrameSentinel_YA0037(func() {
				recurseThenPanic_YA0037(recursionDepth)
			})

			return "unreachable", nil
		}

		_, err := interceptor(context.Background(), nil, info, handler)
		if err == nil {
			t.Fatal("expected error after panic")
		}

		if spy.stack == "" {
			t.Fatal("expected stack trace to be captured in log call")
		}

		if len(spy.stack) <= 4096 {
			t.Fatalf("expected captured stack to exceed 4096 bytes (proves dynamic sizing), got %d bytes", len(spy.stack))
		}

		contains := strings.Contains(spy.stack, "outermostFrameSentinel_YA0037")
		if !contains {
			t.Fatalf("expected stack trace to contain %q (the sentinel frame past the 4 KiB boundary); stack length=%d", "outermostFrameSentinel_YA0037", len(spy.stack))
		}
	})

	t.Run("stream interceptor captures outermost recursion frames", func(t *testing.T) {
		spy := &stackCapturingLogger{}

		clog.Use(spy)

		defer clog.Use(cslog.NewLogger())

		interceptor := StreamRecoveryInterceptor()
		info := &grpc.StreamServerInfo{FullMethod: "/test.Service/StreamDeepPanic"}

		handler := func(_ any, _ grpc.ServerStream) error {
			outermostFrameSentinel_YA0037(func() {
				recurseThenPanic_YA0037(recursionDepth)
			})

			return nil
		}

		ss := &fakeServerStream{ctx: context.Background()}

		err := interceptor(nil, ss, info, handler)
		if err == nil {
			t.Fatal("expected error after panic")
		}

		if spy.stack == "" {
			t.Fatal("expected stack trace to be captured in log call")
		}

		if len(spy.stack) <= 4096 {
			t.Fatalf("expected captured stack to exceed 4096 bytes (proves dynamic sizing), got %d bytes", len(spy.stack))
		}

		contains := strings.Contains(spy.stack, "outermostFrameSentinel_YA0037")
		if !contains {
			t.Fatalf("expected stack trace to contain %q (the sentinel frame past the 4 KiB boundary); stack length=%d", "outermostFrameSentinel_YA0037", len(spy.stack))
		}
	})
}
