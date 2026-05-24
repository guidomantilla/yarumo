package otelhttp

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// defaultScopeName is the instrumentation scope name used by both
// transports when the caller does not override it via WithName.
const defaultScopeName = "github.com/guidomantilla/yarumo/extensions/telemetry/otel/http"

// Option is a functional option for configuring Options. Both
// NewMetricsTransport and NewTracingTransport accept the same Option
// type; each constructor reads only the fields it cares about.
type Option func(opts *Options)

// Options holds the configuration shared by both transports in this
// package. The metrics transport consumes meterProvider + name; the
// tracing transport consumes tracerProvider + name + propagator +
// redaction + spanNameFn. The name field doubles as the instrumentation
// scope passed to MeterProvider.Meter and TracerProvider.Tracer. Fields
// the constructor does not care about are ignored — callers can configure
// a single Options bag and pass it to either or both transports.
type Options struct {
	name string

	meterProvider metric.MeterProvider

	tracerProvider trace.TracerProvider
	propagator     propagation.TextMapPropagator
	redaction      map[string]struct{}
	spanNameFn     SpanNameFn
}

// NewOptions creates Options with safe defaults and applies the given
// functional options. Defaults pull the global meter / tracer / propagator
// from OTel; the instrumentation scope name resolves to this package's
// import path; the default span-name function emits "HTTP <method>".
func NewOptions(opts ...Option) *Options {
	options := &Options{
		name:           defaultScopeName,
		meterProvider:  otel.GetMeterProvider(),
		tracerProvider: otel.GetTracerProvider(),
		propagator:     otel.GetTextMapPropagator(),
		spanNameFn:     defaultSpanName,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithName overrides the instrumentation scope name passed to both
// MeterProvider.Meter and TracerProvider.Tracer. Empty values are ignored,
// preserving the package-path default.
func WithName(name string) Option {
	return func(opts *Options) {
		if name != "" {
			opts.name = name
		}
	}
}

// WithMeterProvider overrides the meter provider used by the metrics
// transport. Nil values are ignored, preserving the global default.
func WithMeterProvider(provider metric.MeterProvider) Option {
	return func(opts *Options) {
		if provider != nil {
			opts.meterProvider = provider
		}
	}
}

// WithTracerProvider overrides the tracer provider used by the tracing
// transport. Nil values are ignored, preserving the global default.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return func(opts *Options) {
		if provider != nil {
			opts.tracerProvider = provider
		}
	}
}

// WithPropagator overrides the text-map propagator used by the tracing
// transport to inject W3C trace context (or any other configured carrier)
// into outgoing request headers. Nil values are ignored.
func WithPropagator(propagator propagation.TextMapPropagator) Option {
	return func(opts *Options) {
		if propagator != nil {
			opts.propagator = propagator
		}
	}
}

// WithHeaderRedaction marks request headers whose values must be masked
// as <redacted> on the span emitted by the tracing transport. Header
// names are matched case-insensitively (canonicalized to MIME header
// form). Empty argument list is a no-op; subsequent calls accumulate.
func WithHeaderRedaction(names ...string) Option {
	return func(opts *Options) {
		if len(names) == 0 {
			return
		}
		if opts.redaction == nil {
			opts.redaction = make(map[string]struct{}, len(names))
		}
		for _, name := range names {
			opts.redaction[http.CanonicalHeaderKey(name)] = struct{}{}
		}
	}
}

// WithSpanNameFn overrides the span-name function used by the tracing
// transport. Nil values are ignored, preserving the default
// ("HTTP <method>").
func WithSpanNameFn(fn SpanNameFn) Option {
	return func(opts *Options) {
		if fn != nil {
			opts.spanNameFn = fn
		}
	}
}

// defaultSpanName is the default SpanNameFn. It emits the request method
// prefixed with "HTTP " to keep span cardinality bounded (the URL goes on
// the span as an attribute, not in the name).
func defaultSpanName(req *http.Request) string {
	return "HTTP " + req.Method
}
