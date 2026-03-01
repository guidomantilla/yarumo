package passwords

// Default generator parameters.
const (
	DefaultPasswordLength = 26
	DefaultMinSpecialChar = 4
	DefaultMinNum         = 6
	DefaultMinUpperCase   = 6
)

// GeneratorOption is the functional option type for Generator.
type GeneratorOption func(opts *GeneratorOptions)

// GeneratorOptions holds the configurable fields for a Generator.
type GeneratorOptions struct {
	passwordLength int
	minSpecialChar int
	minNum         int
	minUpperCase   int
}

// NewGeneratorOptions creates GeneratorOptions with defaults.
func NewGeneratorOptions(opts ...GeneratorOption) *GeneratorOptions {
	options := &GeneratorOptions{
		passwordLength: DefaultPasswordLength,
		minSpecialChar: DefaultMinSpecialChar,
		minNum:         DefaultMinNum,
		minUpperCase:   DefaultMinUpperCase,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithPasswordLength sets the minimum password length.
func WithPasswordLength(length int) GeneratorOption {
	return func(opts *GeneratorOptions) {
		if length >= DefaultPasswordLength {
			opts.passwordLength = length
		}
	}
}

// WithMinSpecialChar sets the minimum number of special characters.
func WithMinSpecialChar(count int) GeneratorOption {
	return func(opts *GeneratorOptions) {
		if count >= DefaultMinSpecialChar {
			opts.minSpecialChar = count
		}
	}
}

// WithMinNum sets the minimum number of numeric characters.
func WithMinNum(count int) GeneratorOption {
	return func(opts *GeneratorOptions) {
		if count >= DefaultMinNum {
			opts.minNum = count
		}
	}
}

// WithMinUpperCase sets the minimum number of uppercase characters.
func WithMinUpperCase(count int) GeneratorOption {
	return func(opts *GeneratorOptions) {
		if count >= DefaultMinUpperCase {
			opts.minUpperCase = count
		}
	}
}
