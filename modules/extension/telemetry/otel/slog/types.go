// Package slog provides OpenTelemetry-aware adapters for the project's
// log/slog package. It exposes:
//
//   - TraceExtractor: an attribute extractor that pulls trace_id / span_id /
//     trace_flags from an active OTel span context, surfacing them on every
//     log record emitted by a Logger configured with WithOtelTrace.
//   - Bridge: a slog.Handler that re-emits every record through the OTel
//     Logs API (OTLP export, via the global LoggerProvider).
//
// Keeping the OTel dependency out of log/slog is intentional: log/ must
// remain free of any telemetry SDK so it can be imported by very small
// services that do not run OTel.
//
// Recommended import alias by callers: otelslog (to disambiguate from the
// stdlib log/slog and from yarumo's log/slog).
package slog

import (
	"log/slog"

	cslog "github.com/guidomantilla/yarumo/extension/common/log/slog"
)

var (
	_ slog.Handler = (*bridge)(nil)

	_ TraceExtractorFn = TraceExtractor
	_ WithOtelTraceFn  = WithOtelTrace
)

// TraceExtractorFn is the function type for TraceExtractor.
type TraceExtractorFn func() cslog.AttrExtractor

// WithOtelTraceFn is the function type for WithOtelTrace.
type WithOtelTraceFn func() cslog.Option
