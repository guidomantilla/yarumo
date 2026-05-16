package slog

import (
	"io"
	"log/slog"
	"os"
)

// Option is a functional option for configuring logger Options.
type Option func(opts *Options)

// Options holds the configuration for the slog Logger.
type Options struct {
	level      Level
	writer     io.Writer
	handlers   []slog.Handler
	extractors []AttrExtractor
}

// NewOptions creates a new Options with sensible defaults and applies the given options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		level:  LevelOff,
		writer: os.Stderr,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithLevel sets the minimum logging level. Invalid levels are ignored.
func WithLevel(level Level) Option {
	return func(opts *Options) {
		switch level {
		case LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal, LevelOff:
			opts.level = level
		}
	}
}

// WithWriter sets the output writer. Nil values are ignored.
func WithWriter(writer io.Writer) Option {
	return func(opts *Options) {
		if writer != nil {
			opts.writer = writer
		}
	}
}

// WithHandlers appends custom slog handlers to the logger pipeline.
func WithHandlers(handlers ...slog.Handler) Option {
	return func(opts *Options) {
		if len(handlers) > 0 {
			opts.handlers = append(opts.handlers, handlers...)
		}
	}
}

// WithContextExtractors registers context-aware attribute extractors on the
// Logger. The extractors run on every record produced by the Logger; their
// returned attrs are merged into the record before fanout to the configured
// handlers. Nil extractors are filtered out. Multiple calls accumulate.
func WithContextExtractors(extractors ...AttrExtractor) Option {
	return func(opts *Options) {
		for _, extractor := range extractors {
			if extractor != nil {
				opts.extractors = append(opts.extractors, extractor)
			}
		}
	}
}
