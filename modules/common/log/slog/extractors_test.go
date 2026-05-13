package slog

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/guidomantilla/yarumo/common/log/slog/slogctx"
)

func TestSlogctxExtractor(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when no bag is bound", func(t *testing.T) {
		t.Parallel()

		if got := SlogctxExtractor(context.Background()); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("returns slogctx attrs", func(t *testing.T) {
		t.Parallel()

		ctx := slogctx.WithAttrs(context.Background(), slog.String("k", "v"))

		got := SlogctxExtractor(ctx)
		if len(got) != 1 || got[0].Key != "k" {
			t.Fatalf("got %v, want one k attr", got)
		}
	})
}

func TestWithContextExtractors_LoggerIntegration(t *testing.T) {
	t.Parallel()

	t.Run("attrs from context land on every record", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelInfo), WithContextExtractors(SlogctxExtractor))

		ctx := slogctx.WithAttrs(context.Background(),
			slog.String("request_id", "abc"),
			slog.String("user_id", "u1"),
		)

		l.Info(ctx, "request received", "method", "GET")

		var got map[string]any

		err := json.Unmarshal(buf.Bytes(), &got)
		if err != nil {
			t.Fatalf("invalid JSON output %q: %v", buf.String(), err)
		}

		if got["request_id"] != "abc" || got["user_id"] != "u1" {
			t.Fatalf("missing ctx attrs in output: %v", got)
		}

		if got["method"] != "GET" {
			t.Fatalf("missing inline attr in output: %v", got)
		}
	})

	t.Run("no-context-attrs is a no-op", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelInfo), WithContextExtractors(SlogctxExtractor))

		l.Info(context.Background(), "no ctx")

		var got map[string]any

		err := json.Unmarshal(buf.Bytes(), &got)
		if err != nil {
			t.Fatalf("invalid JSON output %q: %v", buf.String(), err)
		}

		if _, ok := got["request_id"]; ok {
			t.Fatalf("did not expect request_id in output: %v", got)
		}
	})

	t.Run("nil extractor filtered out", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithContextExtractors(nil, nil))
		if got := len(opts.extractors); got != 0 {
			t.Fatalf("got %d extractors, want 0", got)
		}
	})

	t.Run("multiple calls accumulate", func(t *testing.T) {
		t.Parallel()

		first := func(context.Context) []slog.Attr { return nil }
		second := func(context.Context) []slog.Attr { return nil }

		opts := NewOptions(WithContextExtractors(first), WithContextExtractors(second))
		if got := len(opts.extractors); got != 2 {
			t.Fatalf("got %d extractors, want 2", got)
		}
	})
}
