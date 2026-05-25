package authn

// Default claim keys used to map a token payload onto a Principal.
// They match the most common naming conventions (RFC 7519 "sub" for
// the subject, and the de-facto "name" / "roles" extras).
const (
	defaultSubjectClaim = "sub"
	defaultNameClaim    = "name"
	defaultRolesClaim   = "roles"
)

// Option is a functional option for configuring token Options.
type Option func(opts *Options)

// Options holds the configuration for a TokenAuthenticator. Each field
// names the token payload key from which the corresponding Principal
// attribute is sourced.
type Options struct {
	subjectClaim string
	nameClaim    string
	rolesClaim   string
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		subjectClaim: defaultSubjectClaim,
		nameClaim:    defaultNameClaim,
		rolesClaim:   defaultRolesClaim,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithSubjectClaim sets the token payload key from which Principal.ID
// is sourced. Empty values are ignored (the default "sub" is preserved).
func WithSubjectClaim(claim string) Option {
	return func(opts *Options) {
		if claim != "" {
			opts.subjectClaim = claim
		}
	}
}

// WithNameClaim sets the token payload key from which Principal.Name
// is sourced. Empty values are ignored (the default "name" is preserved).
func WithNameClaim(claim string) Option {
	return func(opts *Options) {
		if claim != "" {
			opts.nameClaim = claim
		}
	}
}

// WithRolesClaim sets the token payload key from which Principal.Roles
// is sourced. Empty values are ignored (the default "roles" is preserved).
func WithRolesClaim(claim string) Option {
	return func(opts *Options) {
		if claim != "" {
			opts.rolesClaim = claim
		}
	}
}
