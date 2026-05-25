package rbac

import (
	"github.com/guidomantilla/yarumo/security/authz"
)

// defaultPrincipalIDResolver returns ("", false) regardless of input.
// It is the safe default — the consumer must wire
// WithPrincipalIDResolver to teach RBAC how to extract the principal
// id from its authn shape.
func defaultPrincipalIDResolver(_ any) (string, bool) {
	return "", false
}

// Option is a functional option for configuring RBAC Options.
type Option func(opts *Options)

// Options holds the configuration for NewPolicy.
type Options struct {
	hierarchy           map[string][]string
	rolePermissions     map[string][]string
	store               RolesStore
	principalIDResolver PrincipalIDResolverFn
	auditHook           authz.AuditHookFn
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. By default the policy is empty (no roles, no
// permissions) and uses an in-memory RolesStore.
//
// The default audit hook is authz.DefaultAuditHook (log every
// Decision via common/log); pass WithAuditHook(authz.SilentAuditHook)
// to opt out, or any custom hook to redirect.
//
// The principal-id resolver defaults to a "no id available" stub —
// consumers MUST wire WithPrincipalIDResolver to teach RBAC how to
// extract the id from their authn Principal shape, otherwise every
// request denies.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		hierarchy:           map[string][]string{},
		rolePermissions:     map[string][]string{},
		store:               NewInMemoryRolesStore(),
		principalIDResolver: defaultPrincipalIDResolver,
		auditHook:           authz.DefaultAuditHook,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithRolePermissions registers permission strings against role. A
// permission is "<resource_type>.<action>" with "*" wildcards allowed
// in either segment (e.g. "orders.*", "*.read", "*"). Empty role names
// and empty permission strings are silently ignored. Repeated calls
// accumulate.
func WithRolePermissions(role string, permissions ...string) Option {
	return func(opts *Options) {
		if role == "" {
			return
		}

		filtered := make([]string, 0, len(permissions))
		for _, p := range permissions {
			if p != "" {
				filtered = append(filtered, p)
			}
		}

		if len(filtered) == 0 {
			return
		}

		opts.rolePermissions[role] = append(opts.rolePermissions[role], filtered...)
	}
}

// WithInheritance declares that child inherits every permission of
// parent (and transitively parent's parents). Empty role names are
// silently ignored. Repeated calls accumulate. Cycle detection
// happens at construction time (NewPolicy returns an Abstain-only
// policy in that case, which the audit hook surfaces).
func WithInheritance(child string, parents ...string) Option {
	return func(opts *Options) {
		if child == "" {
			return
		}

		filtered := make([]string, 0, len(parents))
		for _, p := range parents {
			if p != "" {
				filtered = append(filtered, p)
			}
		}

		if len(filtered) == 0 {
			return
		}

		opts.hierarchy[child] = append(opts.hierarchy[child], filtered...)
	}
}

// WithRolesStore overrides the default in-memory RolesStore with a
// custom implementation. Nil stores are ignored.
func WithRolesStore(store RolesStore) Option {
	return func(opts *Options) {
		if store != nil {
			opts.store = store
		}
	}
}

// WithPrincipalIDResolver installs the function the policy uses to
// extract a principal id from authz.Request.Principal. Nil resolvers
// are ignored.
//
// Consumers MUST wire a resolver — the default returns ("", false),
// which causes every request to deny. RBAC keeps this strict by
// default so an unwired Principal shape is loud at the first request,
// not silently bypassed.
func WithPrincipalIDResolver(resolver PrincipalIDResolverFn) Option {
	return func(opts *Options) {
		if resolver != nil {
			opts.principalIDResolver = resolver
		}
	}
}

// WithAuditHook installs an audit hook fired once per Decision on the
// underlying authz.Policy. The default (when WithAuditHook is not
// passed) is authz.DefaultAuditHook. Pass authz.SilentAuditHook to
// opt out, or any custom hook to redirect. Nil values are ignored.
func WithAuditHook(hook authz.AuditHookFn) Option {
	return func(opts *Options) {
		if hook != nil {
			opts.auditHook = hook
		}
	}
}
