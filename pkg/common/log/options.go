package log

import (
	"time"

	"github.com/rs/zerolog"

	"github.com/guidomantilla/yarumo/pkg/common/utils"
)

type Options struct {
	caller bool
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		caller: false,
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
		options.caller = enabled
	}
}

func WithGlobalLevel(level zerolog.Level) Option {
	return func(_ *Options) {
		if level >= zerolog.TraceLevel && level < zerolog.Disabled {
			zerolog.SetGlobalLevel(level)
		}
	}
}

func WithDisableSampling(disable bool) Option {
	return func(_ *Options) {
		zerolog.DisableSampling(disable)
	}
}

func WithTimestampFieldName(name string) Option {
	return func(_ *Options) {
		if utils.NotEmpty(name) {
			zerolog.TimestampFieldName = name
		}

	}
}

func WithLevelFieldName(name string) Option {
	return func(_ *Options) {
		if utils.NotEmpty(name) {
			zerolog.LevelFieldName = name
		}
	}
}

func WithMessageFieldName(name string) Option {
	return func(_ *Options) {
		if utils.NotEmpty(name) {
			zerolog.MessageFieldName = name
		}
	}
}

func WithErrorFieldName(name string) Option {
	return func(_ *Options) {
		if utils.NotEmpty(name) {
			zerolog.ErrorFieldName = name
		}
	}
}

func WithTimeFieldFormat(format string) Option {
	return func(_ *Options) {
		if utils.NotEmpty(format) {
			zerolog.TimeFieldFormat = format
		}
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
		if utils.NotNil(handler) {
			zerolog.ErrorHandler = handler
		}
	}
}

func WithFloatingPointPrecision(precision int) Option {
	return func(_ *Options) {
		if precision >= zerolog.FloatingPointPrecision {
			zerolog.FloatingPointPrecision = precision
		}
	}
}
