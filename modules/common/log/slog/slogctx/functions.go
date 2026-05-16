package slogctx

import (
	"context"
	"log/slog"
)

// WithAttrs returns a child context whose bag carries the supplied attrs in
// addition to anything already present in ctx's bag. The parent bag is never
// mutated: callers can fork divergent branches off a shared ancestor without
// cross-contamination. When ctx has no bag yet, a new one is created.
func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	if ctx == nil {
		return ctx
	}

	parent := fromContext(ctx)

	child := newBag()
	if parent != nil {
		child.attrs = parent.snapshot()
	}

	child.append(attrs...)

	return context.WithValue(ctx, bagKey, child)
}

// SetAttrs appends the supplied attrs to the bag carried by ctx. If ctx has
// no bag yet, the call is a no-op (use WithAttrs first to establish one).
// The mutation is visible to every goroutine that shares ctx.
func SetAttrs(ctx context.Context, attrs ...slog.Attr) {
	carried := fromContext(ctx)
	if carried == nil {
		return
	}

	carried.append(attrs...)
}

// Attrs returns a snapshot of the attrs bound to ctx. The snapshot is safe
// to read concurrently with further SetAttrs calls on the same context.
// Returns nil when ctx has no bag.
func Attrs(ctx context.Context) []slog.Attr {
	carried := fromContext(ctx)
	if carried == nil {
		return nil
	}

	return carried.snapshot()
}
