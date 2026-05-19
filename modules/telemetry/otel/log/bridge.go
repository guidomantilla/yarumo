package log

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	otelog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
)

// Bridge implements slog.Handler by re-emitting every record through the
// OpenTelemetry Logs API. The lookup of the global LoggerProvider is lazy
// (performed on each Handle call) so the bridge can be installed before
// telemetry.Observe runs: pre-Observe records drop silently on the noop
// provider, post-Observe records export via OTLP.
//
// Bridge is safe for concurrent use by multiple goroutines. Its only state
// is the immutable name, minLevel, baseAttrs slice, and groupPath slice —
// the latter two are copied on WithAttrs and WithGroup so the receiver and
// its derivatives never share mutable storage.
type Bridge struct {
	name      string
	minLevel  slog.Level
	baseAttrs []slog.Attr
	groupPath []string
}

// NewBridge constructs a slog.Handler that forwards every record to the
// OpenTelemetry Logs API under the given instrumentation scope name.
// minLevel filters records before any translation work happens.
func NewBridge(name string, minLevel slog.Level) *Bridge {
	return &Bridge{name: name, minLevel: minLevel}
}

// Enabled reports whether level passes the bridge's minimum threshold.
func (b *Bridge) Enabled(_ context.Context, level slog.Level) bool {
	return level >= b.minLevel
}

// Handle translates r into an otelog.Record and emits it via the current
// global LoggerProvider. Inherited attrs (from WithAttrs) and groups (from
// WithGroup) are flattened into dotted keys and prepended to the record.
func (b *Bridge) Handle(ctx context.Context, r slog.Record) error {
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

// WithAttrs returns a new Bridge that prepends attrs to every subsequent
// record. The receiver is not mutated.
func (b *Bridge) WithAttrs(attrs []slog.Attr) slog.Handler {
	cp := *b
	cp.baseAttrs = append(append([]slog.Attr{}, b.baseAttrs...), attrs...)
	return &cp
}

// WithGroup returns a new Bridge that prefixes every attribute key with
// name. Successive WithGroup calls compose as dotted prefixes:
// WithGroup("a").WithGroup("b") flattens keys as "a.b.<key>".
func (b *Bridge) WithGroup(name string) slog.Handler {
	cp := *b
	cp.groupPath = append(append([]string{}, b.groupPath...), name)
	return &cp
}

// mapSeverity translates an slog level into OpenTelemetry severity + text.
func mapSeverity(l slog.Level) (otelog.Severity, string) {
	switch {
	case l >= slog.LevelError:
		return otelog.SeverityError, "ERROR"
	case l >= slog.LevelWarn:
		return otelog.SeverityWarn, "WARN"
	case l >= slog.LevelInfo:
		return otelog.SeverityInfo, "INFO"
	case l >= slog.LevelDebug:
		return otelog.SeverityDebug, "DEBUG"
	default:
		return otelog.SeverityTrace, "TRACE"
	}
}

// toKV converts a resolved slog.Value into the matching otelog.KeyValue.
// KindGroup recurses; KindAny falls back to a string formatted with %v.
func toKV(key string, v slog.Value) otelog.KeyValue {
	v = v.Resolve()
	switch v.Kind() {
	case slog.KindString:
		return otelog.String(key, v.String())
	case slog.KindInt64:
		return otelog.Int64(key, v.Int64())
	case slog.KindUint64:
		return otelog.Int64(key, int64(v.Uint64()))
	case slog.KindFloat64:
		return otelog.Float64(key, v.Float64())
	case slog.KindBool:
		return otelog.Bool(key, v.Bool())
	case slog.KindDuration:
		return otelog.Int64(key, v.Duration().Nanoseconds())
	case slog.KindTime:
		return otelog.String(key, v.Time().Format(time.RFC3339Nano))
	case slog.KindGroup:
		kvs := make([]otelog.KeyValue, 0, len(v.Group()))
		for _, a := range v.Group() {
			kvs = append(kvs, toKV(a.Key, a.Value))
		}
		return otelog.Map(key, kvs...)
	default:
		return otelog.String(key, fmt.Sprintf("%v", v.Any()))
	}
}

// prefixKey joins groups with dots and appends key. With no groups it
// returns key unchanged.
func prefixKey(groups []string, key string) string {
	if len(groups) == 0 {
		return key
	}
	return strings.Join(groups, ".") + "." + key
}
