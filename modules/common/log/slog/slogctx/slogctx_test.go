package slogctx

import (
	"context"
	"log/slog"
	"sync"
	"testing"
)

func TestNewBag(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil empty bag", func(t *testing.T) {
		t.Parallel()

		bag := NewBag()
		if bag == nil {
			t.Fatal("expected non-nil bag")
		}

		if got := bag.Snapshot(); got != nil {
			t.Fatalf("got %v, want nil snapshot", got)
		}
	})
}

func TestBag_Append(t *testing.T) {
	t.Parallel()

	t.Run("appends attributes", func(t *testing.T) {
		t.Parallel()

		bag := NewBag()
		bag.Append(slog.String("k1", "v1"), slog.Int("k2", 2))

		got := bag.Snapshot()
		if len(got) != 2 {
			t.Fatalf("got %d attrs, want 2", len(got))
		}

		if got[0].Key != "k1" || got[1].Key != "k2" {
			t.Fatalf("got keys %q, %q, want k1, k2", got[0].Key, got[1].Key)
		}
	})

	t.Run("empty input is no-op", func(t *testing.T) {
		t.Parallel()

		bag := NewBag()
		bag.Append()

		if got := bag.Snapshot(); got != nil {
			t.Fatalf("got %v, want nil snapshot", got)
		}
	})

	t.Run("nil receiver is no-op", func(t *testing.T) {
		t.Parallel()

		var bag *Bag
		bag.Append(slog.String("k", "v"))
	})

	t.Run("concurrent appends are safe", func(t *testing.T) {
		t.Parallel()

		bag := NewBag()

		var wg sync.WaitGroup

		for i := range 100 {
			wg.Add(1)

			go func(idx int) {
				defer wg.Done()
				bag.Append(slog.Int("k", idx))
			}(i)
		}

		wg.Wait()

		if got := len(bag.Snapshot()); got != 100 {
			t.Fatalf("got %d attrs, want 100", got)
		}
	})
}

func TestBag_Snapshot(t *testing.T) {
	t.Parallel()

	t.Run("nil receiver returns nil", func(t *testing.T) {
		t.Parallel()

		var bag *Bag
		if got := bag.Snapshot(); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("returned slice is a copy", func(t *testing.T) {
		t.Parallel()

		bag := NewBag()
		bag.Append(slog.String("k", "v"))

		snap := bag.Snapshot()
		snap[0] = slog.String("modified", "x")

		if got := bag.Snapshot(); got[0].Key != "k" {
			t.Fatalf("internal state mutated via snapshot: got %q, want %q", got[0].Key, "k")
		}
	})
}

func TestInject(t *testing.T) {
	t.Parallel()

	t.Run("binds bag to context", func(t *testing.T) {
		t.Parallel()

		bag := NewBag()
		ctx := Inject(context.Background(), bag)

		if got := FromContext(ctx); got != bag {
			t.Fatalf("got %v, want %v", got, bag)
		}
	})

	t.Run("nil ctx returned as is", func(t *testing.T) {
		t.Parallel()

		bag := NewBag()

		ctx := Inject(context.TODO(), bag)
		if ctx == nil {
			t.Fatal("expected non-nil ctx")
		}

		var nilCtx context.Context

		if got := Inject(nilCtx, bag); got != nil {
			t.Fatalf("got %v, want nil ctx", got)
		}
	})

	t.Run("nil bag is no-op", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		if got := Inject(ctx, nil); got != ctx {
			t.Fatal("expected same ctx when bag is nil")
		}
	})
}

func TestFromContext(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for empty context", func(t *testing.T) {
		t.Parallel()

		if got := FromContext(context.Background()); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("returns nil for nil context", func(t *testing.T) {
		t.Parallel()

		var nilCtx context.Context
		if got := FromContext(nilCtx); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})
}

func TestWithAttrs(t *testing.T) {
	t.Parallel()

	t.Run("attaches attrs to a fresh context", func(t *testing.T) {
		t.Parallel()

		ctx := WithAttrs(context.Background(), slog.String("request_id", "abc"))

		attrs := Attrs(ctx)
		if len(attrs) != 1 || attrs[0].Key != "request_id" {
			t.Fatalf("got %v, want one request_id attr", attrs)
		}
	})

	t.Run("merges parent attrs with child attrs", func(t *testing.T) {
		t.Parallel()

		parent := WithAttrs(context.Background(), slog.String("request_id", "abc"))
		child := WithAttrs(parent, slog.String("user_id", "u1"))

		attrs := Attrs(child)
		if len(attrs) != 2 {
			t.Fatalf("got %d attrs, want 2", len(attrs))
		}
	})

	t.Run("child mutations do not leak to parent", func(t *testing.T) {
		t.Parallel()

		parent := WithAttrs(context.Background(), slog.String("a", "1"))
		child := WithAttrs(parent, slog.String("b", "2"))

		SetAttrs(child, slog.String("c", "3"))

		if got := len(Attrs(parent)); got != 1 {
			t.Fatalf("parent got %d attrs, want 1 — child mutations leaked", got)
		}

		if got := len(Attrs(child)); got != 3 {
			t.Fatalf("child got %d attrs, want 3", got)
		}
	})

	t.Run("nil ctx falls back to background", func(t *testing.T) {
		t.Parallel()

		var nilCtx context.Context

		ctx := WithAttrs(nilCtx, slog.String("k", "v"))
		if ctx == nil {
			t.Fatal("expected non-nil ctx")
		}

		if got := Attrs(ctx); len(got) != 1 {
			t.Fatalf("got %d attrs, want 1", len(got))
		}
	})

	t.Run("no attrs still creates isolated bag", func(t *testing.T) {
		t.Parallel()

		ctx := WithAttrs(context.Background())

		if got := FromContext(ctx); got == nil {
			t.Fatal("expected an attached bag even with empty input")
		}
	})
}

func TestSetAttrs(t *testing.T) {
	t.Parallel()

	t.Run("appends to bound bag", func(t *testing.T) {
		t.Parallel()

		ctx := WithAttrs(context.Background())
		SetAttrs(ctx, slog.String("user_id", "u1"))

		attrs := Attrs(ctx)
		if len(attrs) != 1 || attrs[0].Key != "user_id" {
			t.Fatalf("got %v, want one user_id attr", attrs)
		}
	})

	t.Run("no bag is no-op", func(t *testing.T) {
		t.Parallel()

		SetAttrs(context.Background(), slog.String("k", "v"))
	})

	t.Run("nil ctx is no-op", func(t *testing.T) {
		t.Parallel()

		var nilCtx context.Context
		SetAttrs(nilCtx, slog.String("k", "v"))
	})

	t.Run("empty input is no-op", func(t *testing.T) {
		t.Parallel()

		ctx := WithAttrs(context.Background())
		SetAttrs(ctx)

		if got := Attrs(ctx); got != nil {
			t.Fatalf("got %v, want nil attrs", got)
		}
	})
}

func TestAttrs(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for empty context", func(t *testing.T) {
		t.Parallel()

		if got := Attrs(context.Background()); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("returns snapshot independent of bag", func(t *testing.T) {
		t.Parallel()

		ctx := WithAttrs(context.Background(), slog.String("k", "v"))
		snap := Attrs(ctx)

		SetAttrs(ctx, slog.String("k2", "v2"))

		if len(snap) != 1 {
			t.Fatalf("snapshot mutated after later SetAttrs: got %d, want 1", len(snap))
		}
	})
}
