package passwords

const (
	PasswordLength = 26
	MinSpecialChar = 4
	MinNum         = 6
	MinUpperCase   = 6
)

type GeneratorOptions struct {
	passwordLength int
	minSpecialChar int
	minNum         int
	minUpperCase   int
}

func NewGeneratorOptions(opts ...GeneratorOption) *GeneratorOptions {
	options := &GeneratorOptions{
		passwordLength: PasswordLength,
		minSpecialChar: MinSpecialChar,
		minNum:         MinNum,
		minUpperCase:   MinUpperCase,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type GeneratorOption func(opts *GeneratorOptions)

func WithPasswordLength(length int) GeneratorOption {
	return func(opts *GeneratorOptions) {
		if length > PasswordLength {
			opts.passwordLength = length
		}
	}
}

func WithMinSpecialChar(min int) GeneratorOption {
	return func(opts *GeneratorOptions) {
		if min > MinSpecialChar {
			opts.minSpecialChar = min
		}
	}
}

func WithMinNum(min int) GeneratorOption {
	return func(opts *GeneratorOptions) {
		if min > MinNum {
			opts.minNum = min
		}
	}
}

func WithMinUpperCase(min int) GeneratorOption {
	return func(opts *GeneratorOptions) {
		if min > MinUpperCase {
			opts.minUpperCase = min
		}
	}
}
