package slog

// Attribute keys used for OTel trace correlation. They follow the OpenTelemetry
// log semantic conventions so that downstream collectors and dashboards can
// link log records to their parent trace.
const (
	AttrTraceID    = "trace_id"
	AttrSpanID     = "span_id"
	AttrTraceFlags = "trace_flags"
)
