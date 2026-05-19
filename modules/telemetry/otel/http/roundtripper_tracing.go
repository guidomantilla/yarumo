package http

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// tracingRoundTripper wraps an http.RoundTripper and opens a client-kind
// span around each RoundTrip. Trace context is injected into the request
// headers via the configured propagator before the call.
type tracingRoundTripper struct {
	base       http.RoundTripper
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
	opts       *TracingOptions
}

// NewTracingRoundTripper returns an http.RoundTripper that emits an OTel
// span for every request before delegating to base. If base is nil it
// falls back to http.DefaultTransport. The tracer is resolved via the
// provided TracerProvider (default: otel.GetTracerProvider()).
func NewTracingRoundTripper(base http.RoundTripper, opts ...TracingOption) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	o := NewTracingOptions(opts...)
	tracer := o.tracerProvider.Tracer(o.tracerName)

	return &tracingRoundTripper{
		base:       base,
		tracer:     tracer,
		propagator: o.propagator,
		opts:       o,
	}
}

// RoundTrip wraps the call with a client span, injects trace context into
// the outgoing headers, records HTTP attributes and the final status.
func (t *tracingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	sr := &spanRequest{
		method: req.Method,
		host:   hostOrEmpty(req),
		path:   pathOrEmpty(req),
	}
	name := t.opts.spanNameFn(sr)

	ctx, span := t.tracer.Start(req.Context(), name, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", sr.method),
		attribute.String("http.host", sr.host),
		attribute.String("http.path", sr.path),
	)

	headerAttrs := collectHeaderAttributes(req.Header, t.opts.redactedHeaders)
	if len(headerAttrs) > 0 {
		span.SetAttributes(headerAttrs...)
	}

	// Inject trace context into a clone so we don't mutate caller state.
	reqWithCtx := req.WithContext(ctx)
	t.propagator.Inject(ctx, propagation.HeaderCarrier(reqWithCtx.Header))

	resp, err := t.base.RoundTrip(reqWithCtx)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return resp, err
	}

	if resp != nil {
		span.SetAttributes(attribute.Int("http.status", resp.StatusCode))
		if resp.StatusCode >= 400 {
			span.SetStatus(codes.Error, http.StatusText(resp.StatusCode))
		}
	}

	return resp, err
}
