package slog

import (
	"context"
	"log/slog"

	cslogctx "github.com/guidomantilla/yarumo/common/log/slog/slogctx"
)

// SlogctxExtractor is an AttrExtractor that reads the attribute bag bound to
// ctx by the slogctx subpackage. It is the canonical bridge between
// slogctx.WithAttrs / SetAttrs and a Logger configured via
// WithContextExtractors.
func SlogctxExtractor(ctx context.Context) []slog.Attr {
	return cslogctx.Attrs(ctx)
}

// ReplaceLevel replaces the level attribute with a more readable value.
func ReplaceLevel(_ []string, a slog.Attr) slog.Attr {
	if a.Key != slog.LevelKey {
		return a
	}

	level, ok := a.Value.Any().(slog.Level)
	if !ok {
		return a
	}

	switch Level(level) {
	case LevelTrace:
		a.Value = slog.StringValue("TRACE")
	case LevelFatal:
		a.Value = slog.StringValue("FATAL")
	default:
	}

	return a
}
