package generator

// Default generator parameters. These are floor values used by NewOptions
// when the caller does not provide explicit With<Field> overrides.
//
// IMPORTANT: validation thresholds are NOT duplicated here. The Generator
// validates raw passwords against the configured values on the *Options*,
// not against these defaults. Callers may override any field down to zero;
// the only invariant enforced at construction time is that the sum of
// minimums does not exceed the total password length.
const (
	DefaultPasswordLength = 26
	DefaultMinSpecialChar = 4
	DefaultMinNumber      = 6
	DefaultMinUpperCase   = 6
	DefaultMinLowerCase   = 6
)

// Option is the functional option type for Generator.
type Option func(opts *Options)

// Options holds the configurable fields for a Generator.
type Options struct {
	passwordLength int
	minSpecialChar int
	minNumber      int
	minUpperCase   int
	minLowerCase   int
}

// NewOptions creates Options with defaults and applies the provided overrides.
// All With<Field> options accept any non-negative integer; consistency
// (sum of minimums <= passwordLength) is enforced by NewGenerator, not here.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		passwordLength: DefaultPasswordLength,
		minSpecialChar: DefaultMinSpecialChar,
		minNumber:      DefaultMinNumber,
		minUpperCase:   DefaultMinUpperCase,
		minLowerCase:   DefaultMinLowerCase,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithPasswordLength sets the total password length. Negative values are
// clamped to zero; NewGenerator rejects configurations where the sum of
// character-class minimums exceeds this value.
func WithPasswordLength(length int) Option {
	return func(opts *Options) {
		if length < 0 {
			length = 0
		}
		opts.passwordLength = length
	}
}

// WithMinSpecialChar sets the minimum number of special characters required
// in the generated/validated password. Negative values are clamped to zero.
func WithMinSpecialChar(count int) Option {
	return func(opts *Options) {
		if count < 0 {
			count = 0
		}
		opts.minSpecialChar = count
	}
}

// WithMinNumber sets the minimum number of numeric characters required
// in the generated/validated password. Negative values are clamped to zero.
func WithMinNumber(count int) Option {
	return func(opts *Options) {
		if count < 0 {
			count = 0
		}
		opts.minNumber = count
	}
}

// WithMinUpperCase sets the minimum number of uppercase characters required
// in the generated/validated password. Negative values are clamped to zero.
func WithMinUpperCase(count int) Option {
	return func(opts *Options) {
		if count < 0 {
			count = 0
		}
		opts.minUpperCase = count
	}
}

// WithMinLowerCase sets the minimum number of lowercase characters required
// in the generated/validated password. Negative values are clamped to zero.
func WithMinLowerCase(count int) Option {
	return func(opts *Options) {
		if count < 0 {
			count = 0
		}
		opts.minLowerCase = count
	}
}
