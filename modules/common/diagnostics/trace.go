package diagnostics

import (
	"io"
	"runtime/trace"
	"time"
)

type TraceFlightRecorder interface {
	Start() error
	Stop()
	Enabled() bool
	WriteTo(w io.Writer) (n int64, err error)
}

type tracefr struct {
	*trace.FlightRecorder
}

func NewTraceFlightRecorder(options ...Option) TraceFlightRecorder {
	opts := NewOptions(options...)
	return &tracefr{
		FlightRecorder: trace.NewFlightRecorder(trace.FlightRecorderConfig{
			MinAge:   opts.minAge,
			MaxBytes: opts.maxBytes,
		}),
	}
}

type Option func(opts *Options)

type Options struct {
	minAge   time.Duration
	maxBytes uint64
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		minAge:   10 * time.Second,
		maxBytes: 10 << 2,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}
