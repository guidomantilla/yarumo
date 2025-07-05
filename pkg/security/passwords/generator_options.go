package passwords

type GeneratorOptions struct {
	passwordLength int
	minSpecialChar int
	minNum         int
	minUpperCase   int
}

func NewGeneratorOptions(opts ...GeneratorOption) *GeneratorOptions {
	options := &GeneratorOptions{
		passwordLength: 16,
		minSpecialChar: 2,
		minNum:         2,
		minUpperCase:   2,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type GeneratorOption func(opts *GeneratorOptions)

func WithPasswordLength(length int) GeneratorOption {
	return func(opts *GeneratorOptions) {
		opts.passwordLength = length
	}
}

func WithMinSpecialChar(min int) GeneratorOption {
	return func(opts *GeneratorOptions) {
		opts.minSpecialChar = min
	}
}

func WithMinNum(min int) GeneratorOption {
	return func(opts *GeneratorOptions) {
		opts.minNum = min
	}
}

func WithMinUpperCase(min int) GeneratorOption {
	return func(opts *GeneratorOptions) {
		opts.minUpperCase = min
	}
}
