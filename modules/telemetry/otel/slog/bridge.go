package slog

import (
	"context"
	"log/slog"
	"time"

	otelog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
)

// bridge implements slog.Handler by re-emitting every record through the
// OpenTelemetry Logs API. The lookup of the global LoggerProvider is lazy
// (performed on each Handle call) so the bridge can be installed before
// telemetry.Observe runs: pre-Observe records drop silently on the noop
// provider, post-Observe records export via OTLP.
//
// bridge is safe for concurrent use by multiple goroutines. Its only state
// is the immutable name, minLevel, baseAttrs slice, and groupPath slice —
// the latter two are copied on WithAttrs and WithGroup so the receiver and
// its derivatives never share mutable storage.
type bridge struct {
	name      string
	minLevel  slog.Level
	baseAttrs []slog.Attr
	groupPath []string
}

// NewBridge returns a slog.Handler that forwards every record to the
// OpenTelemetry Logs API under the given instrumentation scope name.
// minLevel filters records before any translation work happens.
func NewBridge(name string, minLevel slog.Level) slog.Handler {
	return &bridge{name: name, minLevel: minLevel}
}

// Enabled reports whether level passes the bridge's minimum threshold.
func (b *bridge) Enabled(_ context.Context, level slog.Level) bool {
	return level >= b.minLevel
}

// Handle translates r into an otelog.Record and emits it via the current
// global LoggerProvider. Inherited attrs (from WithAttrs) and groups (from
// WithGroup) are flattened into dotted keys and prepended to the record.
func (b *bridge) Handle(ctx context.Context, r slog.Record) error {
	var rec otelog.Record
	rec.SetTimestamp(r.Time)
	rec.SetObservedTimestamp(time.Now())

	sev, text := mapSeverity(r.Level)
	rec.SetSeverity(sev)
	rec.SetSeverityText(text)
	rec.SetBody(otelog.StringValue(r.Message))

	for _, a := range b.baseAttrs {
		rec.AddAttributes(toKV(prefixKey(b.groupPath, a.Key), a.Value))
	}
	r.Attrs(func(a slog.Attr) bool {
		rec.AddAttributes(toKV(prefixKey(b.groupPath, a.Key), a.Value))
		return true
	})

	global.GetLoggerProvider().Logger(b.name).Emit(ctx, rec)
	return nil
}

// WithAttrs returns a new bridge that prepends attrs to every subsequent
// record. The receiver is not mutated.
func (b *bridge) WithAttrs(attrs []slog.Attr) slog.Handler {
	cp := *b
	cp.baseAttrs = append(append([]slog.Attr{}, b.baseAttrs...), attrs...)
	return &cp
}

// WithGroup returns a new bridge that prefixes every attribute key with
// name. Successive WithGroup calls compose as dotted prefixes:
// WithGroup("a").WithGroup("b") flattens keys as "a.b.<key>".
func (b *bridge) WithGroup(name string) slog.Handler {
	cp := *b
	cp.groupPath = append(append([]string{}, b.groupPath...), name)
	return &cp
}
