// Package slog provides OpenTelemetry-aware adapters for the project's
// common/log/slog package. It exposes:
//
//   - TraceExtractor: an attribute extractor that pulls trace_id / span_id /
//     trace_flags from an active OTel span context, surfacing them on every
//     log record emitted by a Logger configured with WithOtelTrace.
//   - Bridge: a slog.Handler that re-emits every record through the OTel
//     Logs API (OTLP export, via the global LoggerProvider).
//
// Keeping the OTel dependency out of common/log/slog is intentional: common/
// must remain free of any telemetry SDK so it can be imported by very small
// services that do not run OTel.
//
// Recommended import alias by callers: otelslog (to disambiguate from the
// stdlib log/slog and from common/log/slog).
package slog

import (
	"log/slog"

	cslog "github.com/guidomantilla/yarumo/common/log/slog"
)

var (
	_ slog.Handler = (*Bridge)(nil)

	_ TraceExtractorFn = TraceExtractor
	_ WithOtelTraceFn  = WithOtelTrace
	_ NewBridgeFn      = NewBridge
)

// TraceExtractorFn is the function type for TraceExtractor.
type TraceExtractorFn func() cslog.AttrExtractor

// WithOtelTraceFn is the function type for WithOtelTrace.
type WithOtelTraceFn func() cslog.Option

// NewBridgeFn is the function type for NewBridge.
type NewBridgeFn func(name string, minLevel slog.Level) *Bridge
