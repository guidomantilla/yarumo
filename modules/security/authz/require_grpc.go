package authz

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// RequireUnary returns a gRPC unary server interceptor that evaluates
// policy against the inbound RPC for the given action. Allow forwards
// to the handler; Deny / Abstain short-circuits with
// codes.PermissionDenied and Decision.Reason as the status message.
//
// Action and policy validation mirrors RequireHTTP: nil policy or
// empty action panics at construction time (fail closed loud, not at
// request time).
func RequireUnary(policy Policy, action string, opts ...Option) grpc.UnaryServerInterceptor {
	if policy == nil {
		panic(ErrAuthz(ErrPolicyNil))
	}

	if action == "" {
		panic(ErrAuthz(ErrActionEmpty))
	}

	options := NewOptions(opts...)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		method := ""
		if info != nil {
			method = info.FullMethod
		}

		principal, ok := readPrincipalGRPC(ctx, options.principalReader)
		if !ok {
			dec := Deny("principal not present in context")
			authzReq := Request{Action: action}
			options.auditHook(ctx, authzReq, dec)

			return nil, denyStatus(dec)
		}

		authzReq := buildGRPCRequest(ctx, principal, action, method, req, options.grpcResourceFn)
		dec := policy.Evaluate(ctx, authzReq)
		options.auditHook(ctx, authzReq, dec)

		if dec.Effect != EffectAllow {
			return nil, denyStatus(dec)
		}

		return handler(ctx, req)
	}
}

// RequireStream returns a gRPC stream server interceptor with the
// same evaluation logic as RequireUnary, but applied to streaming
// RPCs.
//
// The principal and resource resolution happen once per stream
// (before the handler returns). Per-message checks are not performed
// — callers that need per-message authorization can call Evaluate
// directly inside their handler.
func RequireStream(policy Policy, action string, opts ...Option) grpc.StreamServerInterceptor {
	if policy == nil {
		panic(ErrAuthz(ErrPolicyNil))
	}

	if action == "" {
		panic(ErrAuthz(ErrActionEmpty))
	}

	options := NewOptions(opts...)

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		method := ""
		if info != nil {
			method = info.FullMethod
		}

		principal, ok := readPrincipalGRPC(ctx, options.principalReader)
		if !ok {
			dec := Deny("principal not present in context")
			authzReq := Request{Action: action}
			options.auditHook(ctx, authzReq, dec)

			return denyStatus(dec)
		}

		authzReq := buildGRPCRequest(ctx, principal, action, method, nil, options.grpcResourceFn)
		dec := policy.Evaluate(ctx, authzReq)
		options.auditHook(ctx, authzReq, dec)

		if dec.Effect != EffectAllow {
			return denyStatus(dec)
		}

		return handler(srv, ss)
	}
}

// readPrincipalGRPC delegates to the configured PrincipalReader. A nil
// reader returns ok=false (caller treats as deny).
func readPrincipalGRPC(ctx context.Context, reader PrincipalReader) (any, bool) {
	if reader == nil {
		return nil, false
	}

	return reader.Read(ctx)
}

// buildGRPCRequest assembles a Request from the gRPC ctx, the
// principal, the action, the gRPC FullMethod, the typed request
// message (unary only) and the optional resource resolver.
func buildGRPCRequest(ctx context.Context, principal any, action string, method string, req any, resolver GRPCResourceResolverFn) Request {
	resource := Resource{}
	if resolver != nil {
		resource = resolver(ctx, method, req)
	}

	env := Environment{IP: peerIP(ctx)}

	authzReq := NewRequest(principal, action, resource, env)
	authzReq.Environment.Attrs = grpcMethodAttr(method, authzReq.Environment.Attrs)

	return authzReq
}

// grpcMethodAttr inserts the gRPC FullMethod into the env.Attrs map so
// policies can pattern-match on the RPC name without a transport
// dependency. A nil attrs map is allocated lazily.
func grpcMethodAttr(method string, attrs map[string]any) map[string]any {
	if method == "" {
		return attrs
	}

	if attrs == nil {
		attrs = map[string]any{}
	}

	attrs["grpc_method"] = method

	return attrs
}

// peerIP extracts the caller IP from the gRPC peer in ctx. Returns nil
// when no peer is bound or the address is not a recognized IP.
func peerIP(ctx context.Context) net.IP {
	p, ok := peer.FromContext(ctx)
	if !ok || p == nil || p.Addr == nil {
		return nil
	}

	addr := p.Addr.String()
	if addr == "" {
		return nil
	}

	host, _, splitErr := net.SplitHostPort(addr)
	if splitErr != nil {
		host = addr
	}

	return net.ParseIP(host)
}

// denyStatus translates a Deny / Abstain Decision into a gRPC status
// error with codes.PermissionDenied and Reason as the message.
func denyStatus(dec Decision) error {
	return status.Error(codes.PermissionDenied, dec.Reason)
}
