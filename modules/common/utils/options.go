package utils

import "golang.org/x/text/language"

type Options struct {
	charset string
	lang    language.Tag
}

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

type Option func(opts *Options)

func WithCharset(charset string) Option {
	return func(opts *Options) {
		if NotEmpty(charset) {
			opts.charset = charset
		}

	}
}

func WithLanguage(lang language.Tag) Option {
	return func(opts *Options) {
		opts.lang = lang
	}
}
