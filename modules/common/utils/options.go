package utils

import "golang.org/x/text/language"

// Options holds configuration for utility functions that support customization.
type Options struct {
	charset string
	lang    language.Tag
}

// NewOptions creates a new Options with sensible defaults and applies the given options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		charset: AllCharset,
		lang:    language.English,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// Option is a functional option for configuring utils Options.
type Option func(opts *Options)

// WithCharset sets the character set used for random string generation.
func WithCharset(charset string) Option {
	return func(opts *Options) {
		if NotEmpty(charset) {
			opts.charset = charset
		}
	}
}

// WithLanguage sets the language tag used for locale-aware operations.
func WithLanguage(lang language.Tag) Option {
	return func(opts *Options) {
		if lang != language.Und {
			opts.lang = lang
		}
	}
}
