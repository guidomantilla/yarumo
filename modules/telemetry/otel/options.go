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
	tracePropagators      []propagation.TextMapPropagator
	traceExporterOptions  []otlptracegrpc.Option
	traceProviderOptions  []sdktrace.TracerProviderOption
	profileOptions        []runtimemetrics.Option
	profileInternal       time.Duration
	metricExporterOptions []otlpmetricgrpc.Option
	metricProviderOptions []sdkmetric.Option
	metricInterval        time.Duration
	logExporterOptions    []otlploggrpc.Option
	logProviderOptions    []sdklog.LoggerProviderOption
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		tracePropagators:      []propagation.TextMapPropagator{},
		traceExporterOptions:  []otlptracegrpc.Option{},
		traceProviderOptions:  []sdktrace.TracerProviderOption{},
		profileInternal:       runtimemetrics.DefaultMinimumReadMemStatsInterval,
		profileOptions:        []runtimemetrics.Option{},
		metricExporterOptions: []otlpmetricgrpc.Option{},
		metricProviderOptions: []sdkmetric.Option{},
		metricInterval:        time.Millisecond * 60000,
		logExporterOptions:    []otlploggrpc.Option{},
		logProviderOptions:    []sdklog.LoggerProviderOption{},
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

func WithTracePropagators(propagators ...propagation.TextMapPropagator) Option {
	return func(opts *Options) {
		opts.tracePropagators = propagators
	}
}

func WithTraceExporterOptions(exporterOptions ...otlptracegrpc.Option) Option {
	return func(opts *Options) {
		opts.traceExporterOptions = exporterOptions
	}
}

func WithTraceProviderOptions(providerOptions ...sdktrace.TracerProviderOption) Option {
	return func(opts *Options) {
		opts.traceProviderOptions = providerOptions
	}
}

func WithProfileOptions(profileOptions ...runtimemetrics.Option) Option {
	return func(opts *Options) {
		opts.profileOptions = profileOptions
	}
}

func WithProfileInternal(profileInternal time.Duration) Option {
	return func(opts *Options) {
		opts.profileInternal = profileInternal
	}
}

func WithMetricExporterOptions(exporterOptions ...otlpmetricgrpc.Option) Option {
	return func(opts *Options) {
		opts.metricExporterOptions = exporterOptions
	}
}

func WithMetricProviderOptions(providerOptions ...sdkmetric.Option) Option {
	return func(opts *Options) {
		opts.metricProviderOptions = providerOptions
	}
}

func WithMetricInterval(metricInterval time.Duration) Option {
	return func(opts *Options) {
		opts.metricInterval = metricInterval
	}
}

func WithLogExporterOptions(exporterOptions ...otlploggrpc.Option) Option {
	return func(opts *Options) {
		opts.logExporterOptions = exporterOptions
	}
}

func WithLogProviderOptions(providerOptions ...sdklog.LoggerProviderOption) Option {
	return func(opts *Options) {
		opts.logProviderOptions = providerOptions
	}
}
