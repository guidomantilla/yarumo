package tracing

import (
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type HttpTracingRoundTripper struct {
	Tracer trace.Tracer
	Next   http.RoundTripper
}

func (tripper *HttpTracingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	ctx, span := tripper.Tracer.Start(ctx, fmt.Sprintf("%s %s", req.Method, req.URL), trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	newReq := req.Clone(req.Context())
	otel.GetTextMapPropagator().Inject(ctx, propagationHeaderCarrier(newReq.Header))

	start := time.Now()
	resp, err := tripper.Next.RoundTrip(req)
	duration := time.Since(start)

	span.SetAttributes(
		attribute.String("http.method", newReq.Method),
		attribute.String("http.url", newReq.URL.String()),
		attribute.String("http.host", newReq.URL.Host),
		attribute.Float64("http.duration_ms", float64(duration.Milliseconds())),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
	return resp, nil
}

type propagationHeaderCarrier http.Header

func (c propagationHeaderCarrier) Get(key string) string {
	return http.Header(c).Get(key)
}

func (c propagationHeaderCarrier) Set(key, value string) {
	http.Header(c).Set(key, value)
}

func (c propagationHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}
