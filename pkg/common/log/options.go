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
