package v1

import (
	"github.com/DataDog/datadog-go/v5/statsd"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

type Option func(opts *Options)

type Options struct {
	tracerOptions   []tracer.StartOption
	profilerOptions []profiler.Option
	metricsAddr     string
	metricsOptions  []statsd.Option
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		tracerOptions:   []tracer.StartOption{},
		profilerOptions: []profiler.Option{},
		metricsAddr:     "",
		metricsOptions:  []statsd.Option{},
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

func WithTracerOptions(tracerOpts ...tracer.StartOption) Option {
	return func(opts *Options) {
		opts.tracerOptions = tracerOpts
	}
}

func WithProfilerOptions(profilerOpts ...profiler.Option) Option {
	return func(opts *Options) {
		opts.profilerOptions = profilerOpts
	}
}

func WithMetricsAddr(metricsAddr string) Option {
	return func(opts *Options) {
		opts.metricsAddr = metricsAddr
	}
}

func WithMetricsOptions(metricsOpts ...statsd.Option) Option {
	return func(opts *Options) {
		opts.metricsOptions = metricsOpts
	}
}
