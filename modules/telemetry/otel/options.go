package otel

import (
	"time"

	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Option func(opts *Options)

type Options struct {
	tracerPropagators           []propagation.TextMapPropagator
	tracerExporterOptions       []otlptracegrpc.Option
	tracerProviderOptions       []sdktrace.TracerProviderOption
	meterExporterOptions        []otlpmetricgrpc.Option
	meterProviderOptions        []sdkmetric.Option
	meterInterval               time.Duration
	meterRuntimeMetricsEnabled  bool
	meterRuntimeMetricsInterval time.Duration
	meterRuntimeMetricsOptions  []runtimemetrics.Option
	loggerExporterOptions       []otlploggrpc.Option
	loggerProviderOptions       []sdklog.LoggerProviderOption
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		tracerPropagators:           []propagation.TextMapPropagator{},
		tracerExporterOptions:       []otlptracegrpc.Option{},
		tracerProviderOptions:       []sdktrace.TracerProviderOption{},
		meterExporterOptions:        []otlpmetricgrpc.Option{},
		meterProviderOptions:        []sdkmetric.Option{},
		meterInterval:               time.Millisecond * 60000,
		meterRuntimeMetricsEnabled:  false,
		meterRuntimeMetricsInterval: runtimemetrics.DefaultMinimumReadMemStatsInterval,
		meterRuntimeMetricsOptions:  []runtimemetrics.Option{},
		loggerExporterOptions:       []otlploggrpc.Option{},
		loggerProviderOptions:       []sdklog.LoggerProviderOption{},
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

func WithTracerPropagators(propagators ...propagation.TextMapPropagator) Option {
	return func(opts *Options) {
		opts.tracerPropagators = propagators
	}
}

func WithTracerExporterOptions(options ...otlptracegrpc.Option) Option {
	return func(opts *Options) {
		opts.tracerExporterOptions = options
	}
}

func WithTracerProviderOptions(options ...sdktrace.TracerProviderOption) Option {
	return func(opts *Options) {
		opts.tracerProviderOptions = options
	}
}

func WithMeterExporterOptions(options ...otlpmetricgrpc.Option) Option {
	return func(opts *Options) {
		opts.meterExporterOptions = options
	}
}

func WithMeterProviderOptions(options ...sdkmetric.Option) Option {
	return func(opts *Options) {
		opts.meterProviderOptions = options
	}
}

func WithMeterInterval(interval time.Duration) Option {
	return func(opts *Options) {
		opts.meterInterval = interval
	}
}

func WithMeterRuntimeMetricsEnabled(enabled bool) Option {
	return func(opts *Options) {
		opts.meterRuntimeMetricsEnabled = enabled
	}
}

func WithMeterRuntimeMetricsInterval(interval time.Duration) Option {
	return func(opts *Options) {
		opts.meterRuntimeMetricsInterval = interval
	}
}

func WithMeterRuntimeMetricsOptions(options ...runtimemetrics.Option) Option {
	return func(opts *Options) {
		opts.meterRuntimeMetricsOptions = options
	}
}

func WithLoggerExporterOptions(options ...otlploggrpc.Option) Option {
	return func(opts *Options) {
		opts.loggerExporterOptions = options
	}
}

func WithLoggerProviderOptions(options ...sdklog.LoggerProviderOption) Option {
	return func(opts *Options) {
		opts.loggerProviderOptions = options
	}
}
