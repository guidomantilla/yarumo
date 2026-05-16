// Package slogctx provides context-bound slog attribute propagation. The
// package stores a thread-safe bag of slog.Attr values inside a
// context.Context; callers add entries with WithAttrs / SetAttrs, and a
// reader (typically a slog handler) consumes the snapshot via Attrs.
package slogctx

import (
	"context"
	"log/slog"
)

var (
	_ WithAttrsFn = WithAttrs
	_ SetAttrsFn  = SetAttrs
	_ AttrsFn     = Attrs
)

// WithAttrsFn is the function type for WithAttrs.
type WithAttrsFn func(ctx context.Context, attrs ...slog.Attr) context.Context

// SetAttrsFn is the function type for SetAttrs.
type SetAttrsFn func(ctx context.Context, attrs ...slog.Attr)

// AttrsFn is the function type for Attrs.
type AttrsFn func(ctx context.Context) []slog.Attr
