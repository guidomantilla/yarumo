package http

import (
	"github.com/guidomantilla/yarumo/security/authz"
)

// Option is a functional option for configuring http Options.
type Option func(opts *Options)

// Options holds the configuration for the RequireHTTP middleware.
type Options struct {
	principalReader authz.PrincipalReader
	auditHook       authz.AuditHookFn
	resourceFn      HTTPResourceResolverFn
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. The default audit hook (authz.DefaultAuditHook)
// logs every Decision via common/log; pass
// WithAuditHook(authz.SilentAuditHook) to opt out. The PrincipalReader
// is nil by default — RequireHTTP fails closed when invoked without an
// explicitly wired PrincipalReader.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		auditHook: authz.DefaultAuditHook,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithPrincipalReader installs the reader RequireHTTP uses to fetch
// the authenticated principal from ctx. Nil readers are ignored (the
// default — no reader configured — causes RequireHTTP to deny every
// request, since there is no principal to evaluate).
func WithPrincipalReader(reader authz.PrincipalReader) Option {
	return func(opts *Options) {
		if reader != nil {
			opts.principalReader = reader
		}
	}
}

// WithAuditHook installs an observability hook fired once per Decision.
// The default (when WithAuditHook is not passed) is
// authz.DefaultAuditHook, which logs every Decision via common/log so
// consumers that forget to wire observability still see authz
// outcomes. Pass authz.SilentAuditHook to opt out, or any custom hook
// to redirect. Nil values are ignored.
func WithAuditHook(hook authz.AuditHookFn) Option {
	return func(opts *Options) {
		if hook != nil {
			opts.auditHook = hook
		}
	}
}

// WithResourceResolver installs the resolver the middleware uses to
// compute the Resource an HTTP request targets. Nil resolvers are
// ignored. When no resolver is configured, the middleware populates
// Request.Resource with the zero value — useful for action-only checks
// where resource fields carry no meaning.
func WithResourceResolver(resolver HTTPResourceResolverFn) Option {
	return func(opts *Options) {
		if resolver != nil {
			opts.resourceFn = resolver
		}
	}
}
