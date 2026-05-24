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
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Option configures an Options instance.
type Option func(opts *Options)

// Options holds configuration for OpenTelemetry providers.
type Options struct {
	endpoint                    string
	secure                      bool
	resource                    *resource.Resource
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

// NewOptions creates a new Options with defaults and applies the given options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		endpoint:                    "localhost:4317",
		secure:                      true,
		resource:                    &resource.Resource{},
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

// WithEndpoint sets the OTLP endpoint. Empty values are ignored.
func WithEndpoint(endpoint string) Option {
	return func(opts *Options) {
		if endpoint != "" {
			opts.endpoint = endpoint
		}
	}
}

// WithInsecure disables TLS for the OTLP connection.
func WithInsecure() Option {
	return func(opts *Options) {
		opts.secure = false
	}
}

// WithResource sets the OpenTelemetry resource. Nil values are ignored.
func WithResource(resource *resource.Resource) Option {
	return func(opts *Options) {
		if resource != nil {
			opts.resource = resource
		}
	}
}

// WithTracerPropagators sets the tracer propagators.
func WithTracerPropagators(propagators ...propagation.TextMapPropagator) Option {
	return func(opts *Options) {
		opts.tracerPropagators = propagators
	}
}

// WithTracerExporterOptions sets additional tracer exporter options.
func WithTracerExporterOptions(options ...otlptracegrpc.Option) Option {
	return func(opts *Options) {
		opts.tracerExporterOptions = options
	}
}

// WithTracerProviderOptions sets additional tracer provider options.
func WithTracerProviderOptions(options ...sdktrace.TracerProviderOption) Option {
	return func(opts *Options) {
		opts.tracerProviderOptions = options
	}
}

// WithMeterExporterOptions sets additional meter exporter options.
func WithMeterExporterOptions(options ...otlpmetricgrpc.Option) Option {
	return func(opts *Options) {
		opts.meterExporterOptions = options
	}
}

// WithMeterProviderOptions sets additional meter provider options.
func WithMeterProviderOptions(options ...sdkmetric.Option) Option {
	return func(opts *Options) {
		opts.meterProviderOptions = options
	}
}

// WithMeterInterval sets the periodic reader interval. Values less than or equal to zero are ignored.
func WithMeterInterval(interval time.Duration) Option {
	return func(opts *Options) {
		if interval > 0 {
			opts.meterInterval = interval
		}
	}
}

// WithMeterRuntimeMetricsEnabled enables or disables runtime metrics collection.
func WithMeterRuntimeMetricsEnabled(enabled bool) Option {
	return func(opts *Options) {
		opts.meterRuntimeMetricsEnabled = enabled
	}
}

// WithMeterRuntimeMetricsInterval sets the runtime metrics read interval. Values less than or equal to zero are ignored.
func WithMeterRuntimeMetricsInterval(interval time.Duration) Option {
	return func(opts *Options) {
		if interval > 0 {
			opts.meterRuntimeMetricsInterval = interval
		}
	}
}

// WithMeterRuntimeMetricsOptions sets additional runtime metrics options.
func WithMeterRuntimeMetricsOptions(options ...runtimemetrics.Option) Option {
	return func(opts *Options) {
		opts.meterRuntimeMetricsOptions = options
	}
}

// WithLoggerExporterOptions sets additional logger exporter options.
func WithLoggerExporterOptions(options ...otlploggrpc.Option) Option {
	return func(opts *Options) {
		opts.loggerExporterOptions = options
	}
}

// WithLoggerProviderOptions sets additional logger provider options.
func WithLoggerProviderOptions(options ...sdklog.LoggerProviderOption) Option {
	return func(opts *Options) {
		opts.loggerProviderOptions = options
	}
}
