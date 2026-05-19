package http

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Default instrumentation scope names. Callers may override via the
// WithMetricsMeterName / WithTracingTracerName options.
const (
	defaultMeterName  = "github.com/guidomantilla/yarumo/telemetry/otel/http"
	defaultTracerName = "github.com/guidomantilla/yarumo/telemetry/otel/http"
)

// MetricsOptions holds configuration for the metrics RoundTripper.
type MetricsOptions struct {
	meterProvider metric.MeterProvider
	meterName     string
}

// MetricsOption is a functional option for MetricsOptions.
type MetricsOption func(*MetricsOptions)

// NewMetricsOptions creates MetricsOptions with safe defaults and applies the
// given functional options.
func NewMetricsOptions(opts ...MetricsOption) *MetricsOptions {
	o := &MetricsOptions{
		meterProvider: otel.GetMeterProvider(),
		meterName:     defaultMeterName,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithMetricsMeterProvider overrides the meter provider used to obtain the
// instruments. Nil values are silently ignored, preserving the default.
func WithMetricsMeterProvider(p metric.MeterProvider) MetricsOption {
	return func(o *MetricsOptions) {
		if p != nil {
			o.meterProvider = p
		}
	}
}

// WithMetricsMeterName overrides the instrumentation scope name. Empty
// values are silently ignored.
func WithMetricsMeterName(name string) MetricsOption {
	return func(o *MetricsOptions) {
		if name != "" {
			o.meterName = name
		}
	}
}

// TracingOptions holds configuration for the tracing RoundTripper.
type TracingOptions struct {
	tracerProvider     trace.TracerProvider
	tracerName         string
	propagator         propagation.TextMapPropagator
	redactedHeaders    map[string]struct{}
	spanNameFn         func(*spanRequest) string
}

// TracingOption is a functional option for TracingOptions.
type TracingOption func(*TracingOptions)

// NewTracingOptions creates TracingOptions with safe defaults and applies the
// given functional options. Defaults: global TracerProvider, the package
// default tracer name, the global text map propagator, no redacted headers,
// a span name of "HTTP <method>".
func NewTracingOptions(opts ...TracingOption) *TracingOptions {
	o := &TracingOptions{
		tracerProvider:  otel.GetTracerProvider(),
		tracerName:      defaultTracerName,
		propagator:      otel.GetTextMapPropagator(),
		redactedHeaders: map[string]struct{}{},
		spanNameFn:      defaultSpanName,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithTracingTracerProvider overrides the tracer provider used to obtain
// the tracer. Nil values are silently ignored.
func WithTracingTracerProvider(p trace.TracerProvider) TracingOption {
	return func(o *TracingOptions) {
		if p != nil {
			o.tracerProvider = p
		}
	}
}

// WithTracingTracerName overrides the instrumentation scope name. Empty
// values are silently ignored.
func WithTracingTracerName(name string) TracingOption {
	return func(o *TracingOptions) {
		if name != "" {
			o.tracerName = name
		}
	}
}

// WithTracingPropagator overrides the propagator used to inject trace
// context into outgoing request headers. Nil values are silently ignored.
func WithTracingPropagator(p propagation.TextMapPropagator) TracingOption {
	return func(o *TracingOptions) {
		if p != nil {
			o.propagator = p
		}
	}
}

// WithTracingHeaderRedaction adds the supplied header names to the set
// whose values must be masked when recorded on the span as attributes.
// Names are case-insensitive (canonicalised via http.CanonicalHeaderKey
// internally). Empty input is a no-op.
func WithTracingHeaderRedaction(names ...string) TracingOption {
	return func(o *TracingOptions) {
		for _, n := range names {
			if n == "" {
				continue
			}
			o.redactedHeaders[canonicalHeader(n)] = struct{}{}
		}
	}
}

// WithTracingSpanNameFn overrides the function that builds the span name
// from the outgoing request. Nil functions are silently ignored.
func WithTracingSpanNameFn(fn func(method, host, path string) string) TracingOption {
	return func(o *TracingOptions) {
		if fn != nil {
			o.spanNameFn = func(r *spanRequest) string {
				return fn(r.method, r.host, r.path)
			}
		}
	}
}
