package cryptos

type AesCipherOptions struct {
	key []byte
}

func NewAesCipherOptions(opts ...AesCipherOption) *AesCipherOptions {
	options := &AesCipherOptions{
		key: []byte("a-valid-string-secret-that-is-at-least-512-bits-long-which-is-very-long"),
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type AesCipherOption func(opts *AesCipherOptions)

func WithAesCipherKey(key string) AesCipherOption {
	return func(opts *AesCipherOptions) {
		opts.key = []byte(key)
	}
}
