package authz

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func TestRequireUnary_Allow(t *testing.T) {
	t.Parallel()

	t.Run("forwards to handler on allow", func(t *testing.T) {
		t.Parallel()

		reader := PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		interceptor := RequireUnary(allowPolicy{}, "read",
			WithPrincipalReader(reader),
			WithAuditHook(SilentAuditHook),
		)

		var called bool

		handler := func(_ context.Context, _ any) (any, error) {
			called = true

			return "response", nil
		}

		resp, err := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{FullMethod: "/svc/M"}, handler)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !called {
			t.Fatal("expected handler called")
		}

		s, ok := resp.(string)
		if !ok || s != "response" {
			t.Fatalf("expected 'response', got %#v", resp)
		}
	})
}

func TestRequireUnary_Deny(t *testing.T) {
	t.Parallel()

	t.Run("returns PermissionDenied", func(t *testing.T) {
		t.Parallel()

		reader := PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		interceptor := RequireUnary(denyPolicy{reason: "no role"}, "read",
			WithPrincipalReader(reader),
			WithAuditHook(SilentAuditHook),
		)

		var called bool

		handler := func(_ context.Context, _ any) (any, error) {
			called = true

			return nil, nil //nolint:nilnil
		}

		_, err := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{}, handler)
		if err == nil {
			t.Fatal("expected error")
		}

		if called {
			t.Fatal("expected handler NOT called")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected gRPC status error, got %#v", err)
		}

		if st.Code() != codes.PermissionDenied {
			t.Fatalf("expected PermissionDenied, got %s", st.Code())
		}

		if st.Message() != "no role" {
			t.Fatalf("expected 'no role', got %q", st.Message())
		}
	})

	t.Run("denies when principal missing", func(t *testing.T) {
		t.Parallel()

		interceptor := RequireUnary(allowPolicy{}, "read",
			WithAuditHook(SilentAuditHook),
		)

		_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(_ context.Context, _ any) (any, error) {
			return nil, nil //nolint:nilnil
		})

		if err == nil {
			t.Fatal("expected denial error")
		}

		st, _ := status.FromError(err)
		if st.Code() != codes.PermissionDenied {
			t.Fatalf("expected PermissionDenied, got %s", st.Code())
		}
	})
}

func TestRequireUnary_AuditHook(t *testing.T) {
	t.Parallel()

	t.Run("hook fires with decision", func(t *testing.T) {
		t.Parallel()

		reader := PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		var calls int
		var lastDec Decision

		hook := AuditHookFn(func(_ context.Context, _ Request, d Decision) {
			calls++
			lastDec = d
		})

		interceptor := RequireUnary(denyPolicy{reason: "x"}, "read",
			WithPrincipalReader(reader),
			WithAuditHook(hook),
		)

		_, _ = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(_ context.Context, _ any) (any, error) {
			return nil, nil //nolint:nilnil
		})

		if calls != 1 {
			t.Fatalf("expected 1 hook call, got %d", calls)
		}

		if lastDec.Effect != EffectDeny {
			t.Fatalf("expected EffectDeny, got %q", lastDec.Effect)
		}
	})
}

func TestRequireUnary_ResourceResolver(t *testing.T) {
	t.Parallel()

	t.Run("resolver populates resource", func(t *testing.T) {
		t.Parallel()

		reader := PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		policy := &captureRequestPolicy{}

		resolver := GRPCResourceResolverFn(func(_ context.Context, method string, _ any) Resource {
			return Resource{Type: "rpc", ID: method}
		})

		interceptor := RequireUnary(policy, "invoke",
			WithPrincipalReader(reader),
			WithGRPCResourceResolver(resolver),
			WithAuditHook(SilentAuditHook),
		)

		_, _ = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/X"}, func(_ context.Context, _ any) (any, error) {
			return nil, nil //nolint:nilnil
		})

		if policy.last.Resource.Type != "rpc" {
			t.Fatalf("expected 'rpc', got %q", policy.last.Resource.Type)
		}

		if policy.last.Resource.ID != "/svc/X" {
			t.Fatalf("expected '/svc/X', got %q", policy.last.Resource.ID)
		}
	})

	t.Run("method attr populated even without resolver", func(t *testing.T) {
		t.Parallel()

		reader := PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		policy := &captureRequestPolicy{}

		interceptor := RequireUnary(policy, "read",
			WithPrincipalReader(reader),
			WithAuditHook(SilentAuditHook),
		)

		_, _ = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/Y"}, func(_ context.Context, _ any) (any, error) {
			return nil, nil //nolint:nilnil
		})

		got, ok := policy.last.Environment.Attrs["grpc_method"]
		if !ok {
			t.Fatal("expected grpc_method attr")
		}

		s, isStr := got.(string)
		if !isStr || s != "/svc/Y" {
			t.Fatalf("expected '/svc/Y', got %#v", got)
		}
	})
}

func TestRequireUnary_ConstructorFailsClosed(t *testing.T) {
	t.Parallel()

	t.Run("panics on nil policy", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic")
			}
		}()

		_ = RequireUnary(nil, "read")
	})

	t.Run("panics on empty action", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic")
			}
		}()

		_ = RequireUnary(allowPolicy{}, "")
	})
}

// fakeServerStream implements grpc.ServerStream with a configurable ctx
// for stream interceptor tests.
type fakeServerStream struct {
	grpc.ServerStream
	ctx context.Context //nolint:containedctx
}

func (f *fakeServerStream) Context() context.Context {
	return f.ctx
}

func TestRequireStream_Allow(t *testing.T) {
	t.Parallel()

	t.Run("forwards to handler on allow", func(t *testing.T) {
		t.Parallel()

		reader := PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		interceptor := RequireStream(allowPolicy{}, "stream",
			WithPrincipalReader(reader),
			WithAuditHook(SilentAuditHook),
		)

		var called bool

		handler := func(_ any, _ grpc.ServerStream) error {
			called = true

			return nil
		}

		ss := &fakeServerStream{ctx: context.Background()}
		err := interceptor(nil, ss, &grpc.StreamServerInfo{FullMethod: "/svc/S"}, handler)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !called {
			t.Fatal("expected handler called")
		}
	})
}

func TestRequireStream_Deny(t *testing.T) {
	t.Parallel()

	t.Run("returns PermissionDenied", func(t *testing.T) {
		t.Parallel()

		reader := PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		interceptor := RequireStream(denyPolicy{reason: "denied"}, "stream",
			WithPrincipalReader(reader),
			WithAuditHook(SilentAuditHook),
		)

		ss := &fakeServerStream{ctx: context.Background()}
		err := interceptor(nil, ss, &grpc.StreamServerInfo{}, func(_ any, _ grpc.ServerStream) error {
			t.Fatal("handler should not run")

			return nil
		})

		if err == nil {
			t.Fatal("expected error")
		}

		st, _ := status.FromError(err)
		if st.Code() != codes.PermissionDenied {
			t.Fatalf("expected PermissionDenied, got %s", st.Code())
		}
	})

	t.Run("denies when principal missing", func(t *testing.T) {
		t.Parallel()

		interceptor := RequireStream(allowPolicy{}, "stream",
			WithAuditHook(SilentAuditHook),
		)

		ss := &fakeServerStream{ctx: context.Background()}
		err := interceptor(nil, ss, &grpc.StreamServerInfo{}, func(_ any, _ grpc.ServerStream) error {
			return nil
		})

		if err == nil {
			t.Fatal("expected denial")
		}
	})
}

func TestRequireStream_ConstructorFailsClosed(t *testing.T) {
	t.Parallel()

	t.Run("panics on nil policy", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic")
			}
		}()

		_ = RequireStream(nil, "stream")
	})

	t.Run("panics on empty action", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic")
			}
		}()

		_ = RequireStream(allowPolicy{}, "")
	})
}

func TestPeerIP_ExtractedFromContext(t *testing.T) {
	t.Parallel()

	t.Run("populates env.IP from peer", func(t *testing.T) {
		t.Parallel()

		reader := PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		policy := &captureRequestPolicy{}

		interceptor := RequireUnary(policy, "read",
			WithPrincipalReader(reader),
			WithAuditHook(SilentAuditHook),
		)

		addr := &net.TCPAddr{IP: net.ParseIP("10.0.0.7"), Port: 42}
		ctx := peer.NewContext(context.Background(), &peer.Peer{Addr: addr})

		_, _ = interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/svc/X"}, func(_ context.Context, _ any) (any, error) {
			return nil, nil //nolint:nilnil
		})

		if policy.last.Environment.IP == nil {
			t.Fatal("expected non-nil IP")
		}

		if policy.last.Environment.IP.String() != "10.0.0.7" {
			t.Fatalf("expected 10.0.0.7, got %s", policy.last.Environment.IP)
		}
	})

	t.Run("no peer leaves IP nil", func(t *testing.T) {
		t.Parallel()

		reader := PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		policy := &captureRequestPolicy{}

		interceptor := RequireUnary(policy, "read",
			WithPrincipalReader(reader),
			WithAuditHook(SilentAuditHook),
		)

		_, _ = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(_ context.Context, _ any) (any, error) {
			return nil, nil //nolint:nilnil
		})

		if policy.last.Environment.IP != nil {
			t.Fatalf("expected nil IP, got %s", policy.last.Environment.IP)
		}
	})
}
