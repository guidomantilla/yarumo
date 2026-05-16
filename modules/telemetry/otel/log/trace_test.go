package log

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	cslog "github.com/guidomantilla/yarumo/common/log/slog"
	"go.opentelemetry.io/otel/trace"
)

// newSpanContext returns a deterministic, valid trace.SpanContext for tests.
func newSpanContext(t *testing.T) trace.SpanContext {
	t.Helper()

	traceID, err := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	if err != nil {
		t.Fatalf("failed to build trace id: %v", err)
	}

	spanID, err := trace.SpanIDFromHex("1112131415161718")
	if err != nil {
		t.Fatalf("failed to build span id: %v", err)
	}

	return trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     false,
	})
}

func TestTraceExtractor_NoSpan(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when no span is active", func(t *testing.T) {
		t.Parallel()

		extractor := TraceExtractor()
		if got := extractor(context.Background()); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("returns nil when ctx is nil", func(t *testing.T) {
		t.Parallel()

		extractor := TraceExtractor()

		var nilCtx context.Context
		if got := extractor(nilCtx); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("returns nil when span context is invalid", func(t *testing.T) {
		t.Parallel()

		ctx := trace.ContextWithSpanContext(context.Background(), trace.SpanContext{})

		extractor := TraceExtractor()
		if got := extractor(ctx); got != nil {
			t.Fatalf("got %v, want nil for invalid span context", got)
		}
	})
}

func TestTraceExtractor_ActiveSpan(t *testing.T) {
	t.Parallel()

	t.Run("emits trace_id, span_id and trace_flags", func(t *testing.T) {
		t.Parallel()

		sc := newSpanContext(t)
		ctx := trace.ContextWithSpanContext(context.Background(), sc)

		extractor := TraceExtractor()
		attrs := extractor(ctx)

		if len(attrs) != 3 {
			t.Fatalf("got %d attrs, want 3", len(attrs))
		}

		got := make(map[string]string, len(attrs))
		for _, a := range attrs {
			got[a.Key] = a.Value.String()
		}

		if got[AttrTraceID] != sc.TraceID().String() {
			t.Fatalf("got %q, want %q for %s", got[AttrTraceID], sc.TraceID().String(), AttrTraceID)
		}

		if got[AttrSpanID] != sc.SpanID().String() {
			t.Fatalf("got %q, want %q for %s", got[AttrSpanID], sc.SpanID().String(), AttrSpanID)
		}

		if got[AttrTraceFlags] != sc.TraceFlags().String() {
			t.Fatalf("got %q, want %q for %s", got[AttrTraceFlags], sc.TraceFlags().String(), AttrTraceFlags)
		}
	})
}

func TestWithOtelTraceFn_LoggerIntegration(t *testing.T) {
	t.Parallel()

	t.Run("trace ids land on every record", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		logger := cslog.NewLogger(
			cslog.WithWriter(buf),
			cslog.WithLevel(cslog.LevelInfo),
			WithOtelTraceFn(),
		)

		sc := newSpanContext(t)
		ctx := trace.ContextWithSpanContext(context.Background(), sc)

		logger.Info(ctx, "request received", "method", "GET")

		var got map[string]any

		err := json.Unmarshal(buf.Bytes(), &got)
		if err != nil {
			t.Fatalf("invalid JSON output %q: %v", buf.String(), err)
		}

		if got[AttrTraceID] != sc.TraceID().String() {
			t.Fatalf("missing trace_id in output: %v", got)
		}

		if got[AttrSpanID] != sc.SpanID().String() {
			t.Fatalf("missing span_id in output: %v", got)
		}

		if got["method"] != "GET" {
			t.Fatalf("missing inline attr in output: %v", got)
		}
	})

	t.Run("no-active-span emits no trace fields", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		logger := cslog.NewLogger(
			cslog.WithWriter(buf),
			cslog.WithLevel(cslog.LevelInfo),
			WithOtelTraceFn(),
		)

		logger.Info(context.Background(), "no span")

		var got map[string]any

		err := json.Unmarshal(buf.Bytes(), &got)
		if err != nil {
			t.Fatalf("invalid JSON output %q: %v", buf.String(), err)
		}

		if _, ok := got[AttrTraceID]; ok {
			t.Fatalf("did not expect %s in output: %v", AttrTraceID, got)
		}
	})
}
