package slogctx

import (
	"context"
	"log/slog"
	"sync"
	"testing"
)

func TestAttrs(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when no bag", func(t *testing.T) {
		t.Parallel()

		got := Attrs(context.Background())
		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("returns nil for nil ctx", func(t *testing.T) {
		t.Parallel()

		got := Attrs(nil) //nolint:staticcheck // intentional nil for input validation test
		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})
}

func TestWithAttrs(t *testing.T) {
	t.Parallel()

	t.Run("creates new bag with attrs", func(t *testing.T) {
		t.Parallel()

		ctx := WithAttrs(context.Background(), slog.String("k", "v"))

		got := Attrs(ctx)
		if len(got) != 1 || got[0].Key != "k" {
			t.Fatalf("got %v, want [k=v]", got)
		}
	})

	t.Run("forking parent leaves it untouched", func(t *testing.T) {
		t.Parallel()

		parent := WithAttrs(context.Background(), slog.String("a", "1"))
		_ = WithAttrs(parent, slog.String("b", "2"))

		got := Attrs(parent)
		if len(got) != 1 || got[0].Key != "a" {
			t.Fatalf("parent mutated: %v", got)
		}
	})

	t.Run("returns ctx unchanged when ctx is nil", func(t *testing.T) {
		t.Parallel()

		got := WithAttrs(nil, slog.String("k", "v")) //nolint:staticcheck // intentional nil for input validation test
		if got != nil {
			t.Fatalf("got non-nil ctx for nil input")
		}
	})
}

func TestSetAttrs(t *testing.T) {
	t.Parallel()

	t.Run("appends to existing bag", func(t *testing.T) {
		t.Parallel()

		ctx := WithAttrs(context.Background(), slog.String("a", "1"))
		SetAttrs(ctx, slog.String("b", "2"))

		got := Attrs(ctx)
		if len(got) != 2 {
			t.Fatalf("got %v, want 2 attrs", got)
		}
	})

	t.Run("noop when no bag", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		SetAttrs(ctx, slog.String("a", "1"))

		got := Attrs(ctx)
		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("concurrent appends are safe", func(t *testing.T) {
		t.Parallel()

		ctx := WithAttrs(context.Background())

		var wg sync.WaitGroup
		for i := range 50 {
			wg.Go(func() {
				SetAttrs(ctx, slog.Int("n", i))
			})
		}

		wg.Wait()

		got := Attrs(ctx)
		if len(got) != 50 {
			t.Fatalf("got %d attrs, want 50", len(got))
		}
	})
}
