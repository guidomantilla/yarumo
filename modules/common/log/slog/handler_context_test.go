package slog

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

// recordingHandler captures the attrs of every record it handles.
type recordingHandler struct {
	enabled bool
	err     error
	records []slog.Record
	groups  []string
	attrs   []slog.Attr
}

func (h *recordingHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return h.enabled
}

func (h *recordingHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)

	return h.err
}

func (h *recordingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	merged = append(merged, h.attrs...)
	merged = append(merged, attrs...)

	return &recordingHandler{enabled: h.enabled, err: h.err, groups: h.groups, attrs: merged}
}

func (h *recordingHandler) WithGroup(name string) slog.Handler {
	groups := make([]string, 0, len(h.groups)+1)
	groups = append(groups, h.groups...)
	groups = append(groups, name)

	return &recordingHandler{enabled: h.enabled, err: h.err, groups: groups, attrs: h.attrs}
}

func collectAttrs(r slog.Record) map[string]any {
	out := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		out[a.Key] = a.Value.Any()

		return true
	})

	return out
}

func TestNewContextHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil inner returns nil", func(t *testing.T) {
		t.Parallel()

		if got := NewContextHandler(nil, func(context.Context) []slog.Attr { return nil }); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("no extractors returns inner unchanged", func(t *testing.T) {
		t.Parallel()

		inner := &recordingHandler{enabled: true}
		if got := NewContextHandler(inner); got != inner {
			t.Fatal("expected the inner handler to be returned unchanged when no extractors are configured")
		}
	})

	t.Run("nil extractors filtered out", func(t *testing.T) {
		t.Parallel()

		inner := &recordingHandler{enabled: true}
		if got := NewContextHandler(inner, nil, nil); got != inner {
			t.Fatal("expected the inner handler to be returned unchanged when all extractors are nil")
		}
	})
}

func TestContextHandler_Enabled(t *testing.T) {
	t.Parallel()

	t.Run("delegates to inner", func(t *testing.T) {
		t.Parallel()

		inner := &recordingHandler{enabled: true}
		h := NewContextHandler(inner, func(context.Context) []slog.Attr { return nil })

		if !h.Enabled(context.Background(), slog.LevelInfo) {
			t.Fatal("expected enabled to be true when inner is enabled")
		}

		inner.enabled = false

		if h.Enabled(context.Background(), slog.LevelInfo) {
			t.Fatal("expected enabled to be false when inner is disabled")
		}
	})
}

func TestContextHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("injects extractor attrs into record", func(t *testing.T) {
		t.Parallel()

		inner := &recordingHandler{enabled: true}
		extractor := func(_ context.Context) []slog.Attr {
			return []slog.Attr{slog.String("request_id", "abc"), slog.String("user_id", "u1")}
		}

		h := NewContextHandler(inner, extractor)

		err := h.Handle(context.Background(), slog.Record{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(inner.records) != 1 {
			t.Fatalf("got %d records, want 1", len(inner.records))
		}

		got := collectAttrs(inner.records[0])
		if got["request_id"] != "abc" || got["user_id"] != "u1" {
			t.Fatalf("got attrs %v, want request_id=abc + user_id=u1", got)
		}
	})

	t.Run("nil extractor result is skipped", func(t *testing.T) {
		t.Parallel()

		inner := &recordingHandler{enabled: true}
		nilExtractor := func(_ context.Context) []slog.Attr { return nil }
		nonNilExtractor := func(_ context.Context) []slog.Attr {
			return []slog.Attr{slog.String("k", "v")}
		}

		h := NewContextHandler(inner, nilExtractor, nonNilExtractor)

		err := h.Handle(context.Background(), slog.Record{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got := collectAttrs(inner.records[0])
		if got["k"] != "v" {
			t.Fatalf("got %v, want k=v", got)
		}
	})

	t.Run("empty context produces no extra attrs", func(t *testing.T) {
		t.Parallel()

		inner := &recordingHandler{enabled: true}
		extractor := func(_ context.Context) []slog.Attr { return nil }

		h := NewContextHandler(inner, extractor)

		err := h.Handle(context.Background(), slog.Record{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := collectAttrs(inner.records[0]); len(got) != 0 {
			t.Fatalf("got attrs %v, want empty", got)
		}
	})

	t.Run("multiple extractors compose in order", func(t *testing.T) {
		t.Parallel()

		inner := &recordingHandler{enabled: true}
		first := func(_ context.Context) []slog.Attr { return []slog.Attr{slog.String("a", "1")} }
		second := func(_ context.Context) []slog.Attr { return []slog.Attr{slog.String("b", "2")} }

		h := NewContextHandler(inner, first, second)

		err := h.Handle(context.Background(), slog.Record{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got := collectAttrs(inner.records[0])
		if got["a"] != "1" || got["b"] != "2" {
			t.Fatalf("got %v, want a=1 + b=2", got)
		}
	})

	t.Run("propagates inner errors", func(t *testing.T) {
		t.Parallel()

		want := errors.New("boom")
		inner := &recordingHandler{enabled: true, err: want}
		h := NewContextHandler(inner, func(context.Context) []slog.Attr { return nil })

		err := h.Handle(context.Background(), slog.Record{})
		if !errors.Is(err, want) {
			t.Fatalf("got %v, want %v", err, want)
		}
	})
}

func TestContextHandler_WithAttrs(t *testing.T) {
	t.Parallel()

	t.Run("returns wrapped handler that still extracts", func(t *testing.T) {
		t.Parallel()

		inner := &recordingHandler{enabled: true}
		extractor := func(_ context.Context) []slog.Attr {
			return []slog.Attr{slog.String("ctx", "yes")}
		}

		h := NewContextHandler(inner, extractor).WithAttrs([]slog.Attr{slog.String("static", "ok")})

		err := h.Handle(context.Background(), slog.Record{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// inner.records is on the original recorder; the WithAttrs child copies it,
		// so the static attrs land in a separate recorder. We only need to verify the
		// ctx attrs flowed through the new wrapper.
		newInner, ok := h.(*contextHandler).inner.(*recordingHandler)
		if !ok {
			t.Fatal("expected inner to remain a recordingHandler")
		}

		if got := collectAttrs(newInner.records[0]); got["ctx"] != "yes" {
			t.Fatalf("got %v, want ctx=yes", got)
		}

		if got := newInner.attrs; len(got) != 1 || got[0].Key != "static" {
			t.Fatalf("got attrs %v, want one static attr", got)
		}
	})
}

func TestContextHandler_WithGroup(t *testing.T) {
	t.Parallel()

	t.Run("empty name returns same handler", func(t *testing.T) {
		t.Parallel()

		inner := &recordingHandler{enabled: true}
		h := NewContextHandler(inner, func(context.Context) []slog.Attr { return nil })

		if got := h.(*contextHandler).WithGroup(""); got != h {
			t.Fatal("expected same handler for empty group name")
		}
	})

	t.Run("non-empty name returns new handler that still extracts", func(t *testing.T) {
		t.Parallel()

		inner := &recordingHandler{enabled: true}
		extractor := func(_ context.Context) []slog.Attr {
			return []slog.Attr{slog.String("ctx", "yes")}
		}

		base, ok := NewContextHandler(inner, extractor).(*contextHandler)
		if !ok {
			t.Fatal("expected NewContextHandler to return *contextHandler")
		}

		h := base.WithGroup("g")

		err := h.Handle(context.Background(), slog.Record{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		child, ok := h.(*contextHandler).inner.(*recordingHandler)
		if !ok {
			t.Fatal("expected inner to remain a recordingHandler")
		}

		if len(child.groups) != 1 || child.groups[0] != "g" {
			t.Fatalf("got groups %v, want [g]", child.groups)
		}

		if got := collectAttrs(child.records[0]); got["ctx"] != "yes" {
			t.Fatalf("got %v, want ctx=yes", got)
		}
	})
}
