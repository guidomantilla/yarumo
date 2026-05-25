package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/security/authn"
)

// NewUnaryInterceptor returns a grpc.UnaryServerInterceptor that
// terminates Bearer authentication for unary RPCs. On success it
// injects the resulting *Principal into the per-RPC ctx via
// authn.WithPrincipal and forwards to the handler. On failure it
// short-circuits with status.Error(codes.Unauthenticated, ...).
//
// A nil authenticator panics via cassert.NotNil so construction-time
// wiring mistakes surface immediately.
func NewUnaryInterceptor(authenticator authn.Authenticator, options ...Option) grpc.UnaryServerInterceptor {
	cassert.NotNil(authenticator, "authenticator is nil")

	opts := NewOptions(options...)

	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		principal, err := authenticatePrincipal(ctx, authenticator, opts)
		if err != nil {
			return nil, err
		}

		return handler(authn.WithPrincipal(ctx, principal), req)
	}
}

// NewStreamInterceptor returns a grpc.StreamServerInterceptor that
// terminates Bearer authentication for streaming RPCs. On success it
// wraps the inbound grpc.ServerStream so its Context() carries the
// validated *Principal, then forwards to the handler. On failure it
// short-circuits with status.Error(codes.Unauthenticated, ...).
func NewStreamInterceptor(authenticator authn.Authenticator, options ...Option) grpc.StreamServerInterceptor {
	cassert.NotNil(authenticator, "authenticator is nil")

	opts := NewOptions(options...)

	return func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		principal, err := authenticatePrincipal(ss.Context(), authenticator, opts)
		if err != nil {
			return err
		}

		wrapped := &authenticatedStream{
			ServerStream: ss,
			ctx:          authn.WithPrincipal(ss.Context(), principal),
		}

		return handler(srv, wrapped)
	}
}

// authenticatePrincipal extracts the bearer token from the gRPC
// metadata, validates it via authenticator, and returns the resulting
// *Principal. All failures are translated to status.Error with
// codes.Unauthenticated so they propagate as a gRPC-native error.
func authenticatePrincipal(ctx context.Context, authenticator authn.Authenticator, opts *Options) (*authn.Principal, error) {
	token, err := extractBearerToken(ctx, opts.metadataKey, opts.scheme)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	principal, err := authenticator.Validate(ctx, token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if principal == nil {
		return nil, status.Error(codes.Unauthenticated, authn.ErrPrincipalNil.Error())
	}

	return principal, nil
}

// extractBearerToken returns the credential portion of an
// "authorization: Bearer <token>" metadata pair. Missing metadata or
// missing key returns ErrHeaderMissing; any other shape (multiple
// values, wrong scheme, empty credential) returns ErrHeaderMalformed.
//
// gRPC stores metadata keys lower-cased and may carry multiple values
// per key; this helper accepts exactly one entry.
func extractBearerToken(ctx context.Context, metadataKey, scheme string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", authn.ErrHeaderMissing
	}

	values := md.Get(metadataKey)
	if len(values) == 0 {
		return "", authn.ErrHeaderMissing
	}

	if len(values) != 1 {
		return "", authn.ErrHeaderMalformed
	}

	parts := strings.SplitN(values[0], " ", 2)
	if len(parts) != 2 {
		return "", authn.ErrHeaderMalformed
	}

	if !strings.EqualFold(parts[0], scheme) {
		return "", authn.ErrHeaderMalformed
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", authn.ErrHeaderMalformed
	}

	return token, nil
}
