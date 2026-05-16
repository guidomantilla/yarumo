// Package slog provides a structured logger implementation built on top of log/slog.
package slog

import (
	"context"
	"log/slog"
)

// Interface compliance for the package's slog.Handler implementations and
// the bundled AttrExtractor adapter.
var (
	_ slog.Handler = (*fanoutHandler)(nil)
	_ slog.Handler = (*contextHandler)(nil)

	_ AttrExtractor = SlogctxExtractor
)

// AttrExtractor is the function type for context-aware slog attribute
// extractors. Implementations must be safe for concurrent use and should
// return nil when no attrs are available so the caller can short-circuit.
type AttrExtractor func(ctx context.Context) []slog.Attr
