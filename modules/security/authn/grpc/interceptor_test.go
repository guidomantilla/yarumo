package grpc_test

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/guidomantilla/yarumo/security/authn"
	authngrpc "github.com/guidomantilla/yarumo/security/authn/grpc"
)

// fakeAuthenticator is a test double for authn.Authenticator.
type fakeAuthenticator struct {
	validateFn func(ctx context.Context, token string) (*authn.Principal, error)
}

func (a *fakeAuthenticator) Validate(ctx context.Context, token string) (*authn.Principal, error) {
	return a.validateFn(ctx, token)
}

func newOKAuthenticator(p *authn.Principal) authn.Authenticator {
	return &fakeAuthenticator{
		validateFn: func(_ context.Context, _ string) (*authn.Principal, error) {
			return p, nil
		},
	}
}

func newRejectingAuthenticator(err error) authn.Authenticator {
	return &fakeAuthenticator{
		validateFn: func(_ context.Context, _ string) (*authn.Principal, error) {
			return nil, err
		},
	}
}

func ctxWithAuth(value string) context.Context {
	md := metadata.New(map[string]string{"authorization": value})
	return metadata.NewIncomingContext(context.Background(), md)
}

func unaryInfo() *grpc.UnaryServerInfo {
	return &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
}

func streamInfo() *grpc.StreamServerInfo {
	return &grpc.StreamServerInfo{FullMethod: "/test.Service/Method"}
}

// fakeStream implements grpc.ServerStream for the stream interceptor
// tests. Only Context() is meaningful here; the other methods are
// satisfied via embedding the nil-valued grpc.ServerStream.
type fakeStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (s *fakeStream) Context() context.Context { return s.ctx }

func TestNewUnaryInterceptor(t *testing.T) {
	t.Parallel()

	t.Run("happy path injects principal", func(t *testing.T) {
		t.Parallel()

		want := &authn.Principal{ID: "u-1"}
		interceptor := authngrpc.NewUnaryInterceptor(newOKAuthenticator(want))

		var got *authn.Principal

		var saw bool

		handler := func(ctx context.Context, _ any) (any, error) {
			got, saw = authn.FromContext(ctx)
			return "ok", nil
		}

		resp, err := interceptor(ctxWithAuth("Bearer t-1"), nil, unaryInfo(), handler)
		if err != nil {
			t.Fatalf("interceptor returned err: %v", err)
		}

		if resp != "ok" {
			t.Fatalf("response = %v, want ok", resp)
		}

		if !saw {
			t.Fatal("handler did not see principal in ctx")
		}

		if got != want {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("missing metadata returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		interceptor := authngrpc.NewUnaryInterceptor(newOKAuthenticator(&authn.Principal{ID: "x"}))

		_, err := interceptor(context.Background(), nil, unaryInfo(), func(_ context.Context, _ any) (any, error) {
			t.Fatal("handler must not be called")
			return nil, nil
		})

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("err %v is not a status", err)
		}

		if st.Code() != codes.Unauthenticated {
			t.Fatalf("code = %v, want Unauthenticated", st.Code())
		}
	})

	t.Run("missing authorization key returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		interceptor := authngrpc.NewUnaryInterceptor(newOKAuthenticator(&authn.Principal{ID: "x"}))

		md := metadata.New(map[string]string{"x-other": "v"})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		_, err := interceptor(ctx, nil, unaryInfo(), func(_ context.Context, _ any) (any, error) {
			return nil, nil
		})

		st, _ := status.FromError(err)

		if st.Code() != codes.Unauthenticated {
			t.Fatalf("code = %v, want Unauthenticated", st.Code())
		}
	})

	t.Run("malformed scheme returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		interceptor := authngrpc.NewUnaryInterceptor(newOKAuthenticator(&authn.Principal{ID: "x"}))

		_, err := interceptor(ctxWithAuth("Basic xxx"), nil, unaryInfo(), func(_ context.Context, _ any) (any, error) {
			return nil, nil
		})

		st, _ := status.FromError(err)
		if st.Code() != codes.Unauthenticated {
			t.Fatalf("code = %v, want Unauthenticated", st.Code())
		}
	})

	t.Run("malformed value (no space) returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		interceptor := authngrpc.NewUnaryInterceptor(newOKAuthenticator(&authn.Principal{ID: "x"}))

		_, err := interceptor(ctxWithAuth("Bearer"), nil, unaryInfo(), func(_ context.Context, _ any) (any, error) {
			return nil, nil
		})

		st, _ := status.FromError(err)
		if st.Code() != codes.Unauthenticated {
			t.Fatalf("code = %v, want Unauthenticated", st.Code())
		}
	})

	t.Run("empty bearer credential returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		interceptor := authngrpc.NewUnaryInterceptor(newOKAuthenticator(&authn.Principal{ID: "x"}))

		_, err := interceptor(ctxWithAuth("Bearer   "), nil, unaryInfo(), func(_ context.Context, _ any) (any, error) {
			return nil, nil
		})

		st, _ := status.FromError(err)
		if st.Code() != codes.Unauthenticated {
			t.Fatalf("code = %v, want Unauthenticated", st.Code())
		}
	})

	t.Run("multiple authorization values returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		interceptor := authngrpc.NewUnaryInterceptor(newOKAuthenticator(&authn.Principal{ID: "x"}))

		md := metadata.MD{}
		md.Append("authorization", "Bearer a", "Bearer b")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		_, err := interceptor(ctx, nil, unaryInfo(), func(_ context.Context, _ any) (any, error) {
			return nil, nil
		})

		st, _ := status.FromError(err)
		if st.Code() != codes.Unauthenticated {
			t.Fatalf("code = %v, want Unauthenticated", st.Code())
		}
	})

	t.Run("authenticator error returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		sentinel := authn.ErrAuthentication(authn.ErrTokenInvalid)
		interceptor := authngrpc.NewUnaryInterceptor(newRejectingAuthenticator(sentinel))

		_, err := interceptor(ctxWithAuth("Bearer bad"), nil, unaryInfo(), func(_ context.Context, _ any) (any, error) {
			t.Fatal("handler must not run")
			return nil, nil
		})

		st, _ := status.FromError(err)
		if st.Code() != codes.Unauthenticated {
			t.Fatalf("code = %v, want Unauthenticated", st.Code())
		}
	})

	t.Run("nil principal with no error returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		interceptor := authngrpc.NewUnaryInterceptor(newOKAuthenticator(nil))

		_, err := interceptor(ctxWithAuth("Bearer t"), nil, unaryInfo(), func(_ context.Context, _ any) (any, error) {
			t.Fatal("handler must not run")
			return nil, nil
		})

		st, _ := status.FromError(err)
		if st.Code() != codes.Unauthenticated {
			t.Fatalf("code = %v, want Unauthenticated", st.Code())
		}
	})

	t.Run("scheme case insensitive", func(t *testing.T) {
		t.Parallel()

		want := &authn.Principal{ID: "u"}
		interceptor := authngrpc.NewUnaryInterceptor(newOKAuthenticator(want))

		var saw bool

		_, err := interceptor(ctxWithAuth("bearer t"), nil, unaryInfo(), func(ctx context.Context, _ any) (any, error) {
			_, saw = authn.FromContext(ctx)
			return nil, nil
		})

		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}

		if !saw {
			t.Fatal("principal not propagated with lower-case scheme")
		}
	})

	t.Run("custom metadata key", func(t *testing.T) {
		t.Parallel()

		want := &authn.Principal{ID: "u"}
		interceptor := authngrpc.NewUnaryInterceptor(newOKAuthenticator(want),
			authngrpc.WithMetadataKey("X-Auth-Token"),
		)

		md := metadata.New(map[string]string{"x-auth-token": "Bearer t"})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		var saw bool

		_, err := interceptor(ctx, nil, unaryInfo(), func(c context.Context, _ any) (any, error) {
			_, saw = authn.FromContext(c)
			return nil, nil
		})

		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}

		if !saw {
			t.Fatal("principal not propagated via custom metadata key")
		}
	})

	t.Run("custom scheme", func(t *testing.T) {
		t.Parallel()

		want := &authn.Principal{ID: "u"}
		interceptor := authngrpc.NewUnaryInterceptor(newOKAuthenticator(want),
			authngrpc.WithScheme("Token"),
		)

		var saw bool

		_, err := interceptor(ctxWithAuth("Token abc"), nil, unaryInfo(), func(c context.Context, _ any) (any, error) {
			_, saw = authn.FromContext(c)
			return nil, nil
		})

		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}

		if !saw {
			t.Fatal("principal not propagated with custom scheme")
		}
	})
}

func TestNewStreamInterceptor(t *testing.T) {
	t.Parallel()

	t.Run("happy path injects principal", func(t *testing.T) {
		t.Parallel()

		want := &authn.Principal{ID: "u-1"}
		interceptor := authngrpc.NewStreamInterceptor(newOKAuthenticator(want))

		var got *authn.Principal

		var saw bool

		handler := func(_ any, ss grpc.ServerStream) error {
			got, saw = authn.FromContext(ss.Context())
			return nil
		}

		ss := &fakeStream{ctx: ctxWithAuth("Bearer t")}

		err := interceptor(nil, ss, streamInfo(), handler)
		if err != nil {
			t.Fatalf("interceptor returned err: %v", err)
		}

		if !saw {
			t.Fatal("handler did not see principal in ctx")
		}

		if got != want {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("missing metadata returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		interceptor := authngrpc.NewStreamInterceptor(newOKAuthenticator(&authn.Principal{ID: "x"}))

		ss := &fakeStream{ctx: context.Background()}

		err := interceptor(nil, ss, streamInfo(), func(_ any, _ grpc.ServerStream) error {
			t.Fatal("handler must not run")
			return nil
		})

		st, _ := status.FromError(err)
		if st.Code() != codes.Unauthenticated {
			t.Fatalf("code = %v, want Unauthenticated", st.Code())
		}
	})

	t.Run("authenticator error returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		interceptor := authngrpc.NewStreamInterceptor(newRejectingAuthenticator(errors.New("bad")))

		ss := &fakeStream{ctx: ctxWithAuth("Bearer t")}

		err := interceptor(nil, ss, streamInfo(), func(_ any, _ grpc.ServerStream) error {
			t.Fatal("handler must not run")
			return nil
		})

		st, _ := status.FromError(err)
		if st.Code() != codes.Unauthenticated {
			t.Fatalf("code = %v, want Unauthenticated", st.Code())
		}
	})

	t.Run("nil principal with no error returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		interceptor := authngrpc.NewStreamInterceptor(newOKAuthenticator(nil))

		ss := &fakeStream{ctx: ctxWithAuth("Bearer t")}

		err := interceptor(nil, ss, streamInfo(), func(_ any, _ grpc.ServerStream) error {
			t.Fatal("handler must not run")
			return nil
		})

		st, _ := status.FromError(err)
		if st.Code() != codes.Unauthenticated {
			t.Fatalf("code = %v, want Unauthenticated", st.Code())
		}
	})
}
