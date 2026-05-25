package authz

import (
	"context"
	"net/http"
)

// HTTPResourceResolverFn is the function type for resolving the
// Resource an HTTP request targets. Require invokes the resolver once
// per inbound HTTP request before evaluating the policy.
//
// Returning the zero Resource is valid for action-only checks where
// resource type/id are irrelevant.
type HTTPResourceResolverFn func(r *http.Request) Resource

// GRPCResourceResolverFn is the function type for resolving the
// Resource a gRPC RPC targets. Require invokes the resolver once per
// inbound RPC before evaluating the policy. method is the gRPC
// FullMethod ("/pkg.Service/Method"), req is the typed request message
// (any) for unary calls or nil for stream calls.
//
// Returning the zero Resource is valid for action-only checks where
// resource type/id are irrelevant.
type GRPCResourceResolverFn func(ctx context.Context, method string, req any) Resource

// Option is a functional option for configuring authz Options.
type Option func(opts *Options)

// Options holds the configuration for the Require HTTP and gRPC
// middleware.
type Options struct {
	principalReader PrincipalReader
	auditHook       AuditHookFn
	httpResourceFn  HTTPResourceResolverFn
	grpcResourceFn  GRPCResourceResolverFn
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. The default audit hook (DefaultAuditHook) logs
// every Decision via common/log; pass WithAuditHook(SilentAuditHook)
// to opt out. The PrincipalReader and resource resolvers are nil by
// default — Require fails closed when invoked without an explicitly
// wired PrincipalReader.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		auditHook: DefaultAuditHook,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithPrincipalReader installs the reader Require uses to fetch the
// authenticated principal from ctx. Nil readers are ignored (the
// default — no reader configured — causes Require to deny every
// request, since there is no principal to evaluate).
func WithPrincipalReader(reader PrincipalReader) Option {
	return func(opts *Options) {
		if reader != nil {
			opts.principalReader = reader
		}
	}
}

// WithAuditHook installs an observability hook fired once per Decision.
// The default (when WithAuditHook is not passed) is DefaultAuditHook,
// which logs every Decision via common/log so consumers that forget to
// wire observability still see authz outcomes. Pass SilentAuditHook to
// opt out, or any custom hook to redirect. Nil values are ignored (the
// previously installed handler is preserved).
func WithAuditHook(hook AuditHookFn) Option {
	return func(opts *Options) {
		if hook != nil {
			opts.auditHook = hook
		}
	}
}

// WithHTTPResourceResolver installs the resolver Require uses to
// compute the Resource an HTTP request targets. Nil resolvers are
// ignored. When no resolver is configured, Require populates
// Request.Resource with the zero value — useful for action-only checks
// where resource fields carry no meaning.
func WithHTTPResourceResolver(resolver HTTPResourceResolverFn) Option {
	return func(opts *Options) {
		if resolver != nil {
			opts.httpResourceFn = resolver
		}
	}
}

// WithGRPCResourceResolver installs the resolver Require uses to
// compute the Resource a gRPC RPC targets. Nil resolvers are ignored.
// When no resolver is configured, Require populates Request.Resource
// with the zero value — useful for action-only checks where resource
// fields carry no meaning.
func WithGRPCResourceResolver(resolver GRPCResourceResolverFn) Option {
	return func(opts *Options) {
		if resolver != nil {
			opts.grpcResourceFn = resolver
		}
	}
}
