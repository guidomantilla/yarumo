package otelhttp

import (
	"fmt"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

// redactedValue is the placeholder used by the tracing transport for
// headers listed in WithHeaderRedaction. The raw header value never
// reaches the span.
const redactedValue = "<redacted>"

// tracingTransport wraps a base RoundTripper, opens a client-kind span
// per request, propagates W3C trace context via the configured
// propagator, records HTTP attributes, and maps the response status to
// the span status.
type tracingTransport struct {
	base       http.RoundTripper
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
	redaction  map[string]struct{}
	spanNameFn SpanNameFn
}

// NewTracingTransport wraps base with a RoundTripper that opens an OTel
// client-kind span per request, injects W3C trace context into the
// outgoing request headers via the configured propagator, records HTTP
// attributes on the span, and maps the response status to Error for 4xx
// / 5xx responses or transport failures.
//
// Only the tracerProvider + name + propagator + redaction + spanNameFn
// fields of the shared Options are consulted; the metrics-only fields
// are ignored.
//
// A nil base falls back to http.DefaultTransport. The returned
// RoundTripper is safe for concurrent use as long as base is.
func NewTracingTransport(base http.RoundTripper, opts ...Option) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	options := NewOptions(opts...)

	return &tracingTransport{
		base:       base,
		tracer:     options.tracerProvider.Tracer(options.name),
		propagator: options.propagator,
		redaction:  options.redaction,
		spanNameFn: options.spanNameFn,
	}
}

// RoundTrip starts a client span, propagates context into the request
// headers, delegates to base, records attributes and span status, and
// returns the response unchanged.
func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cassert.NotNil(t, "tracing transport receiver is nil")
	cassert.NotNil(req, "request is nil")

	ctx, span := t.tracer.Start(req.Context(), t.spanNameFn(req),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("http.method", req.Method),
			attribute.String("http.url", req.URL.String()),
			attribute.String("http.host", req.URL.Host),
			attribute.String("http.path", req.URL.Path),
		),
	)
	defer span.End()

	req = req.Clone(ctx)
	t.propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	for headerName := range t.redaction {
		_, present := req.Header[headerName]
		if !present {
			continue
		}
		span.SetAttributes(attribute.String("http.request.header."+headerName, redactedValue))
	}

	res, err := t.base.RoundTrip(req)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Bool("http.error", true))
		span.RecordError(err)
		return res, err
	}

	if res != nil {
		span.SetAttributes(attribute.String("http.status", strconv.Itoa(res.StatusCode)))
		if res.ContentLength >= 0 {
			span.SetAttributes(attribute.Int64("http.response.size", res.ContentLength))
		}
		if res.StatusCode >= http.StatusBadRequest {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", res.StatusCode))
		}
	}

	return res, nil
}
