package log

import (
	"time"

	"github.com/rs/zerolog"
)

type Options struct {
	Caller bool
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		Caller: false,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

//

type Option func(options *Options)

func WithCaller(enabled bool) Option {
	return func(options *Options) {
		options.Caller = enabled
	}
}

func WithGlobalLevel(level zerolog.Level) Option {
	return func(_ *Options) {
		zerolog.SetGlobalLevel(level)
	}
}

func WithDisableSampling(disable bool) Option {
	return func(_ *Options) {
		zerolog.DisableSampling(disable)
	}
}

func WithTimestampFieldName(name string) Option {
	return func(_ *Options) {
		zerolog.TimestampFieldName = name
	}
}

func WithLevelFieldName(name string) Option {
	return func(_ *Options) {
		zerolog.LevelFieldName = name
	}
}

func WithMessageFieldName(name string) Option {
	return func(_ *Options) {
		zerolog.MessageFieldName = name
	}
}

func WithErrorFieldName(name string) Option {
	return func(_ *Options) {
		zerolog.ErrorFieldName = name
	}
}

func WithTimeFieldFormat(format string) Option {
	return func(_ *Options) {
		zerolog.TimeFieldFormat = format
	}
}

func WithDurationFieldUnit(unit time.Duration) Option {
	return func(_ *Options) {
		zerolog.DurationFieldUnit = unit
	}
}

func WithDurationFieldInteger(integer bool) Option {
	return func(_ *Options) {
		zerolog.DurationFieldInteger = integer
	}
}

func WithErrorHandler(handler func(err error)) Option {
	return func(_ *Options) {
		zerolog.ErrorHandler = handler
	}
}

func WithFloatingPointPrecision(precision int) Option {
	return func(_ *Options) {
		zerolog.FloatingPointPrecision = precision
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
	return func(options *Options) {
		for _, option := range chain.chain {
			option(options)
		}
	}
}

func (chain *OptionsChain) WithCaller(enabled bool) *OptionsChain {
	chain.chain = append(chain.chain, WithCaller(enabled))
	return chain
}

func (chain *OptionsChain) WithGlobalLevel(level zerolog.Level) *OptionsChain {
	chain.chain = append(chain.chain, WithGlobalLevel(level))
	return chain
}

func (chain *OptionsChain) WithDisableSampling(disable bool) *OptionsChain {
	chain.chain = append(chain.chain, WithDisableSampling(disable))
	return chain
}

func (chain *OptionsChain) WithTimestampFieldName(name string) *OptionsChain {
	chain.chain = append(chain.chain, WithTimestampFieldName(name))
	return chain
}

func (chain *OptionsChain) WithLevelFieldName(name string) *OptionsChain {
	chain.chain = append(chain.chain, WithLevelFieldName(name))
	return chain
}

func (chain *OptionsChain) WithMessageFieldName(name string) *OptionsChain {
	chain.chain = append(chain.chain, WithMessageFieldName(name))
	return chain
}

func (chain *OptionsChain) WithErrorFieldName(name string) *OptionsChain {
	chain.chain = append(chain.chain, WithErrorFieldName(name))
	return chain
}

func (chain *OptionsChain) WithTimeFieldFormat(format string) *OptionsChain {
	chain.chain = append(chain.chain, WithTimeFieldFormat(format))
	return chain
}

func (chain *OptionsChain) WithDurationFieldUnit(unit time.Duration) *OptionsChain {
	chain.chain = append(chain.chain, WithDurationFieldUnit(unit))
	return chain
}

func (chain *OptionsChain) WithDurationFieldInteger(integer bool) *OptionsChain {
	chain.chain = append(chain.chain, WithDurationFieldInteger(integer))
	return chain
}

func (chain *OptionsChain) WithErrorHandler(handler func(err error)) *OptionsChain {
	chain.chain = append(chain.chain, WithErrorHandler(handler))
	return chain
}

func (chain *OptionsChain) WithFloatingPointPrecision(precision int) *OptionsChain {
	chain.chain = append(chain.chain, WithFloatingPointPrecision(precision))
	return chain
}
