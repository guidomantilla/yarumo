package zerolog

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Default values for Logger configuration. The defaults match the slog
// sibling: the Logger writes JSON to os.Stderr and is silent until the
// caller explicitly raises the level.
const (
	// DefaultTimeFormat is the timestamp format applied to every record
	// when the caller does not configure it explicitly.
	DefaultTimeFormat = time.RFC3339Nano
	// DefaultSampling is the no-sampling default: every record emitted at
	// or above the configured level is written.
	DefaultSampling uint32 = 1
)

// Option is a functional option for configuring zerolog Logger Options.
type Option func(opts *Options)

// Options holds the configuration for the zerolog Logger.
type Options struct {
	level      Level
	writer     io.Writer
	console    bool
	timeFormat string
	sampling   uint32
}

// NewOptions creates a new Options with sensible defaults and applies the
// given options. Defaults: LevelOff (silent), os.Stderr writer, JSON
// format, RFC3339Nano timestamps, no sampling.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		level:      LevelOff,
		writer:     os.Stderr,
		console:    false,
		timeFormat: DefaultTimeFormat,
		sampling:   DefaultSampling,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithLevel sets the minimum logging level. Invalid levels are ignored,
// preserving the default (LevelOff).
func WithLevel(level Level) Option {
	return func(opts *Options) {
		switch level {
		case LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal, LevelOff:
			opts.level = level
		}
	}
}

// WithWriter sets the output writer. Nil values are ignored, preserving
// the default (os.Stderr).
func WithWriter(writer io.Writer) Option {
	return func(opts *Options) {
		if writer != nil {
			opts.writer = writer
		}
	}
}

// WithConsole toggles the human-readable zerolog ConsoleWriter on top of
// the configured writer. Defaults to false, meaning records are emitted
// as JSON.
func WithConsole(enabled bool) Option {
	return func(opts *Options) {
		opts.console = enabled
	}
}

// WithTimeFormat sets the timestamp format applied to every record. Empty
// values are ignored, preserving the default (RFC3339Nano).
func WithTimeFormat(format string) Option {
	return func(opts *Options) {
		if format != "" {
			opts.timeFormat = format
		}
	}
}

// WithSampling sets the basic 1-in-N sampling rate. A value of 1 emits
// every record; a value of N emits one record out of every N. Zero is
// ignored, preserving the default (1, no sampling).
func WithSampling(every uint32) Option {
	return func(opts *Options) {
		if every > 0 {
			opts.sampling = every
		}
	}
}

// writer returns the io.Writer the Logger should write to, wrapping the
// configured writer in zerolog.ConsoleWriter when console mode is on.
func (o *Options) effectiveWriter() io.Writer {
	if o.console {
		return zerolog.ConsoleWriter{Out: o.writer, TimeFormat: o.timeFormat}
	}

	return o.writer
}
