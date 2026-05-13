// Package slogctx provides helpers to bind slog attributes to a context.Context
// so that downstream handlers can pull them and emit them on every log record.
//
// The package exposes two complementary APIs:
//
//   - WithAttrs returns a new context that carries the supplied attributes in
//     addition to any attributes already attached to the parent context.
//   - SetAttrs mutates an existing per-context bag (created by NewBag) so that
//     attributes set after a context has been propagated through the call stack
//     are still visible to subsequent log records.
//
// Both flavors are useful: WithAttrs is the canonical immutable-context style;
// SetAttrs is useful for handler middleware (e.g. HTTP) that wants to enrich
// the bag at a single point with values resolved later in the request.
package slogctx

import (
	"context"
	"log/slog"
	"sync"
)

// contextKey is the unexported type for context keys defined by this package.
// Empty-struct values of this type are equal across calls (the type has zero
// size and structural equality), so attrsKey returns one each time it is
// called without leaking a mutable package-level variable.
type contextKey struct{}

// attrsKey returns the context.Context key used to store the per-context
// attribute bag.
func attrsKey() contextKey {
	return contextKey{}
}

// Bag is a goroutine-safe container of slog.Attr values bound to a context.
// Use NewBag to obtain one, then either attach it via Inject or rely on
// SetAttrs to attach it on first use.
type Bag struct {
	mu    sync.RWMutex
	attrs []slog.Attr
}

// NewBag returns a fresh, empty Bag.
func NewBag() *Bag {
	return &Bag{}
}

// Snapshot returns a copy of the attributes currently stored in the bag.
// The returned slice is safe to retain and mutate independently of the bag.
func (b *Bag) Snapshot() []slog.Attr {
	if b == nil {
		return nil
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.attrs) == 0 {
		return nil
	}

	out := make([]slog.Attr, len(b.attrs))
	copy(out, b.attrs)

	return out
}

// Append adds the supplied attributes to the bag. Empty input is a no-op.
func (b *Bag) Append(attrs ...slog.Attr) {
	if b == nil || len(attrs) == 0 {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.attrs = append(b.attrs, attrs...)
}

// Inject returns a new context with the supplied bag bound under the package key.
// Passing a nil bag returns ctx unchanged.
func Inject(ctx context.Context, bag *Bag) context.Context {
	if ctx == nil || bag == nil {
		return ctx
	}

	return context.WithValue(ctx, attrsKey(), bag)
}

// FromContext returns the bag bound to ctx, or nil if none is bound.
func FromContext(ctx context.Context) *Bag {
	if ctx == nil {
		return nil
	}

	bag, ok := ctx.Value(attrsKey()).(*Bag)
	if !ok {
		return nil
	}

	return bag
}

// WithAttrs returns a child context that carries the union of ctx's existing
// attributes (if any) plus the supplied attrs. The returned context owns a
// fresh Bag, so subsequent SetAttrs calls on the child do not leak back to ctx.
// Empty input still returns a context with an isolated empty bag, which keeps
// behavior predictable for callers that always shadow before logging.
func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	parent := FromContext(ctx)
	child := NewBag()

	if parent != nil {
		child.Append(parent.Snapshot()...)
	}

	child.Append(attrs...)

	return Inject(ctx, child)
}

// SetAttrs appends the supplied attributes to the bag bound to ctx. If no bag
// is bound, SetAttrs is a no-op (use WithAttrs to seed one first or call
// Inject explicitly). This signature mirrors slog's Logger.With ergonomics
// for cases where the caller cannot replace the context value (e.g. inside
// middleware that already chose a context).
func SetAttrs(ctx context.Context, attrs ...slog.Attr) {
	if ctx == nil || len(attrs) == 0 {
		return
	}

	bag := FromContext(ctx)
	if bag == nil {
		return
	}

	bag.Append(attrs...)
}

// Attrs returns the attributes currently bound to ctx, or nil if none.
// The returned slice is a snapshot and safe to mutate.
func Attrs(ctx context.Context) []slog.Attr {
	return FromContext(ctx).Snapshot()
}
