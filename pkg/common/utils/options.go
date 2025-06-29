package utils

import "golang.org/x/text/language"

type Options struct {
	Charset string
	Lang    language.Tag
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		Charset: AllCharset,
		Lang:    language.English,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type Option func(opts *Options)

func WithCharset(charset string) Option {
	return func(opts *Options) {
		opts.Charset = charset
	}
}

func WithLanguage(lang language.Tag) Option {
	return func(opts *Options) {
		opts.Lang = lang
	}
}

//

type OptionsChain struct {
	chain []Option
}

func Chain() *OptionsChain {
	return &OptionsChain{
		chain: make([]Option, 0),
	}
}

func (chain *OptionsChain) Build() Option {
	return func(opts *Options) {
		for _, option := range chain.chain {
			option(opts)
		}
	}
}

func (chain *OptionsChain) WithCharset(charset string) *OptionsChain {
	chain.chain = append(chain.chain, WithCharset(charset))
	return chain
}

func (chain *OptionsChain) WithLanguage(lang language.Tag) *OptionsChain {
	chain.chain = append(chain.chain, WithLanguage(lang))
	return chain
}
