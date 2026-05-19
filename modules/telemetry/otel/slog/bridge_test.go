package slog

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	otelog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/embedded"
	"go.opentelemetry.io/otel/log/global"
)

// recordingProvider is a minimal otelog.LoggerProvider for tests; it
// captures every record emitted through any Logger it hands out.
type recordingProvider struct {
	embedded.LoggerProvider

	mu      sync.Mutex
	loggers map[string]*recordingLogger
}

func newRecordingProvider() *recordingProvider {
	return &recordingProvider{loggers: make(map[string]*recordingLogger)}
}

func (p *recordingProvider) Logger(name string, _ ...otelog.LoggerOption) otelog.Logger {
	p.mu.Lock()
	defer p.mu.Unlock()
	l, ok := p.loggers[name]
	if !ok {
		l = &recordingLogger{name: name}
		p.loggers[name] = l
	}
	return l
}

// recordingLogger captures records for inspection.
type recordingLogger struct {
	embedded.Logger
	name string

	mu      sync.Mutex
	records []otelog.Record
	ctxs    []context.Context
}

func (l *recordingLogger) Emit(ctx context.Context, rec otelog.Record) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.records = append(l.records, rec)
	l.ctxs = append(l.ctxs, ctx)
}

func (l *recordingLogger) Enabled(_ context.Context, _ otelog.EnabledParameters) bool {
	return true
}

func (l *recordingLogger) snapshot() []otelog.Record {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]otelog.Record, len(l.records))
	copy(out, l.records)
	return out
}

// installRecorder swaps the global LoggerProvider for a recorder and
// returns the per-logger handle plus a cleanup func. Tests using this
// helper cannot run in parallel because the global is shared.
func installRecorder(t *testing.T, name string) *recordingLogger {
	t.Helper()
	prev := global.GetLoggerProvider()
	p := newRecordingProvider()
	global.SetLoggerProvider(p)
	t.Cleanup(func() { global.SetLoggerProvider(prev) })
	return p.Logger(name, nil).(*recordingLogger)
}

// collectAttrs walks the otelog.Record attributes into a map keyed by
// attribute key for easy lookup.
func collectAttrs(rec otelog.Record) map[string]otelog.Value {
	out := map[string]otelog.Value{}
	rec.WalkAttributes(func(kv otelog.KeyValue) bool {
		out[kv.Key] = kv.Value
		return true
	})
	return out
}

func TestBridge_Enabled(t *testing.T) {
	t.Parallel()

	b := NewBridge("svc", slog.LevelInfo)

	t.Run("below threshold rejected", func(t *testing.T) {
		t.Parallel()
		if b.Enabled(context.Background(), slog.LevelDebug) {
			t.Fatalf("Debug must be below Info threshold")
		}
	})

	t.Run("at threshold accepted", func(t *testing.T) {
		t.Parallel()
		if !b.Enabled(context.Background(), slog.LevelInfo) {
			t.Fatalf("Info must pass Info threshold")
		}
	})

	t.Run("above threshold accepted", func(t *testing.T) {
		t.Parallel()
		if !b.Enabled(context.Background(), slog.LevelError) {
			t.Fatalf("Error must pass Info threshold")
		}
	})
}

func TestBridge_Handle_SeverityMapping(t *testing.T) {
	rec := installRecorder(t, "svc")
	b := NewBridge("svc", slog.LevelDebug-100)

	cases := []struct {
		name     string
		level    slog.Level
		wantSev  otelog.Severity
		wantText string
	}{
		{"error", slog.LevelError, otelog.SeverityError, "ERROR"},
		{"warn", slog.LevelWarn, otelog.SeverityWarn, "WARN"},
		{"info", slog.LevelInfo, otelog.SeverityInfo, "INFO"},
		{"debug", slog.LevelDebug, otelog.SeverityDebug, "DEBUG"},
		{"trace", slog.LevelDebug - 4, otelog.SeverityTrace, "TRACE"},
	}

	for _, tc := range cases {
		r := slog.NewRecord(time.Now(), tc.level, "msg", 0)
		err := b.Handle(context.Background(), r)
		if err != nil {
			t.Fatalf("%s: unexpected error %v", tc.name, err)
		}
	}

	got := rec.snapshot()
	if len(got) != len(cases) {
		t.Fatalf("expected %d records, got %d", len(cases), len(got))
	}
	for i, tc := range cases {
		if got[i].Severity() != tc.wantSev {
			t.Fatalf("%s: severity = %v, want %v", tc.name, got[i].Severity(), tc.wantSev)
		}
		if got[i].SeverityText() != tc.wantText {
			t.Fatalf("%s: severity text = %q, want %q", tc.name, got[i].SeverityText(), tc.wantText)
		}
	}
}

func TestBridge_Handle_AttrKinds(t *testing.T) {
	rec := installRecorder(t, "svc")
	b := NewBridge("svc", slog.LevelInfo)

	r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
	r.AddAttrs(
		slog.String("s", "hello"),
		slog.Int64("i", 42),
		slog.Float64("f", 3.14),
		slog.Bool("b", true),
		slog.Duration("d", 100*time.Millisecond),
		slog.Time("t", time.Unix(0, 0).UTC()),
		slog.Group("g", slog.String("nested", "v")),
		slog.Any("any", struct{ X int }{X: 1}),
	)

	err := b.Handle(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	got := rec.snapshot()
	if len(got) != 1 {
		t.Fatalf("expected 1 record, got %d", len(got))
	}

	attrs := collectAttrs(got[0])

	if v, ok := attrs["s"]; !ok || v.Kind() != otelog.KindString || v.AsString() != "hello" {
		t.Fatalf("attr s wrong: %v", v)
	}
	if v, ok := attrs["i"]; !ok || v.Kind() != otelog.KindInt64 || v.AsInt64() != 42 {
		t.Fatalf("attr i wrong: %v", v)
	}
	if v, ok := attrs["f"]; !ok || v.Kind() != otelog.KindFloat64 || v.AsFloat64() != 3.14 {
		t.Fatalf("attr f wrong: %v", v)
	}
	if v, ok := attrs["b"]; !ok || v.Kind() != otelog.KindBool || !v.AsBool() {
		t.Fatalf("attr b wrong: %v", v)
	}
	if v, ok := attrs["d"]; !ok || v.Kind() != otelog.KindInt64 || v.AsInt64() != (100*time.Millisecond).Nanoseconds() {
		t.Fatalf("attr d wrong: %v", v)
	}
	if v, ok := attrs["t"]; !ok || v.Kind() != otelog.KindString {
		t.Fatalf("attr t wrong: %v", v)
	}
	if v, ok := attrs["g"]; !ok || v.Kind() != otelog.KindMap {
		t.Fatalf("attr g wrong: %v", v)
	} else {
		nested := false
		for _, kv := range v.AsMap() {
			if kv.Key == "nested" && kv.Value.AsString() == "v" {
				nested = true
			}
		}
		if !nested {
			t.Fatalf("group attr did not contain nested entry")
		}
	}
	if v, ok := attrs["any"]; !ok || v.Kind() != otelog.KindString {
		t.Fatalf("attr any wrong: %v", v)
	}
}

func TestBridge_Handle_Body(t *testing.T) {
	rec := installRecorder(t, "svc")
	b := NewBridge("svc", slog.LevelInfo)

	r := slog.NewRecord(time.Now(), slog.LevelInfo, "hello world", 0)
	err := b.Handle(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	got := rec.snapshot()
	if got[0].Body().AsString() != "hello world" {
		t.Fatalf("body = %q, want hello world", got[0].Body().AsString())
	}
}

func TestBridge_WithAttrs_Inherits(t *testing.T) {
	rec := installRecorder(t, "svc")
	b := NewBridge("svc", slog.LevelInfo)

	bound := b.WithAttrs([]slog.Attr{slog.String("service", "users")})
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
	r.AddAttrs(slog.Int64("rid", 7))

	err := bound.Handle(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	got := rec.snapshot()
	attrs := collectAttrs(got[0])
	if attrs["service"].AsString() != "users" {
		t.Fatalf("inherited attr missing: %+v", attrs)
	}
	if attrs["rid"].AsInt64() != 7 {
		t.Fatalf("record attr missing: %+v", attrs)
	}
}

func TestBridge_WithAttrs_NoMutationOnReceiver(t *testing.T) {
	rec := installRecorder(t, "svc")
	b := NewBridge("svc", slog.LevelInfo)

	_ = b.WithAttrs([]slog.Attr{slog.String("x", "1")})

	// Emit through the original — the derived attr must NOT appear.
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
	err := b.Handle(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	got := rec.snapshot()
	attrs := collectAttrs(got[0])
	if _, leaked := attrs["x"]; leaked {
		t.Fatalf("derivative attr leaked into receiver: %+v", attrs)
	}
}

func TestBridge_WithGroup_DottedKeys(t *testing.T) {
	rec := installRecorder(t, "svc")
	b := NewBridge("svc", slog.LevelInfo)

	grouped := b.WithGroup("auth").WithGroup("session")
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
	r.AddAttrs(slog.Int64("uid", 99))

	err := grouped.Handle(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	got := rec.snapshot()
	attrs := collectAttrs(got[0])
	if v, ok := attrs["auth.session.uid"]; !ok || v.AsInt64() != 99 {
		t.Fatalf("expected dotted key auth.session.uid=99, got %+v", attrs)
	}
}

func TestBridge_LazyLookup_TolerateNoopBeforeProvider(t *testing.T) {
	// Reset to the default (noop) provider; do NOT install a recorder.
	prev := global.GetLoggerProvider()
	t.Cleanup(func() { global.SetLoggerProvider(prev) })

	b := NewBridge("svc", slog.LevelInfo)
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)

	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("Handle panicked with noop provider: %v", rec)
		}
	}()

	err := b.Handle(context.Background(), r)
	if err != nil {
		t.Fatalf("Handle errored with noop provider: %v", err)
	}
}

func TestBridge_LazyLookup_PicksUpNewProvider(t *testing.T) {
	// Construct the bridge BEFORE the provider exists; verify it picks up
	// the new provider on the next Emit (per-call lookup).
	prev := global.GetLoggerProvider()
	t.Cleanup(func() { global.SetLoggerProvider(prev) })

	b := NewBridge("svc", slog.LevelInfo)

	// Emit once with the default noop provider — record dropped.
	r1 := slog.NewRecord(time.Now(), slog.LevelInfo, "before", 0)
	err := b.Handle(context.Background(), r1)
	if err != nil {
		t.Fatalf("first Handle errored: %v", err)
	}

	// Now install a recorder.
	p := newRecordingProvider()
	global.SetLoggerProvider(p)

	r2 := slog.NewRecord(time.Now(), slog.LevelInfo, "after", 0)
	err = b.Handle(context.Background(), r2)
	if err != nil {
		t.Fatalf("second Handle errored: %v", err)
	}

	rec := p.Logger("svc", nil).(*recordingLogger)
	got := rec.snapshot()
	if len(got) != 1 {
		t.Fatalf("expected 1 record after provider install, got %d", len(got))
	}
	if got[0].Body().AsString() != "after" {
		t.Fatalf("record body = %q, want 'after'", got[0].Body().AsString())
	}
}

// Compile-time assertion that *Bridge satisfies slog.Handler.
var _ slog.Handler = (*Bridge)(nil)
