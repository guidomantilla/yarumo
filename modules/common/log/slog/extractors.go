package slog

import (
	"context"
	"log/slog"

	"github.com/guidomantilla/yarumo/common/log/slog/slogctx"
)

// SlogctxExtractor is an AttrExtractor that returns the attributes bound to
// ctx via the slogctx subpackage (slogctx.WithAttrs / slogctx.SetAttrs).
//
// Register it on a logger with WithContextExtractors(SlogctxExtractor) to have
// every record automatically enriched with context-bound attributes.
func SlogctxExtractor(ctx context.Context) []slog.Attr {
	return slogctx.Attrs(ctx)
}
