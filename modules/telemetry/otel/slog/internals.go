package slog

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	otelog "go.opentelemetry.io/otel/log"
)

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
