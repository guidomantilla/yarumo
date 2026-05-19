package log

import (
	"context"
	"log/slog"

	cslog "github.com/guidomantilla/yarumo/common/log/slog"
	"go.opentelemetry.io/otel/trace"
)

// Field keys used for OTel trace correlation. They follow the OpenTelemetry
// log semantic conventions so that downstream collectors and dashboards can
// link log records to their parent trace.
const (
	// AttrTraceID is the attribute key for the OTel trace identifier.
	AttrTraceID = "trace_id"
	// AttrSpanID is the attribute key for the OTel span identifier.
	AttrSpanID = "span_id"
	// AttrTraceFlags is the attribute key for the OTel trace flags (sampling bits).
	AttrTraceFlags = "trace_flags"
)

// TraceExtractor returns a slog attribute extractor that, when a span is
// active in the supplied context, emits trace_id, span_id, and trace_flags
// attributes per the OTel logging semantic conventions. If no span is active
// (or the span context is invalid), the extractor returns nil and the
// record is left unchanged.
//
// The extractor is safe for concurrent use and incurs no allocation when no
// span is active.
func TraceExtractor() cslog.AttrExtractor {
	return func(ctx context.Context) []slog.Attr {
		if ctx == nil {
			return nil
		}

		sc := trace.SpanContextFromContext(ctx)
		if !sc.IsValid() {
			return nil
		}

		return []slog.Attr{
			slog.String(AttrTraceID, sc.TraceID().String()),
			slog.String(AttrSpanID, sc.SpanID().String()),
			slog.String(AttrTraceFlags, sc.TraceFlags().String()),
		}
	}
}

// WithOtelTrace returns a slog Option that registers TraceExtractor on the
// resulting Logger. It is the convenience wrapper most callers want: pass it
// to slog.NewLogger to enable trace ↔ log correlation in a single line.
//
//	logger := slog.NewLogger(slog.WithLevel(slog.LevelInfo), otellog.WithOtelTrace())
func WithOtelTrace() cslog.Option {
	return cslog.WithContextExtractors(TraceExtractor())
}
