package otel

import (
	"context"
	"time"
)

// unwindTimeout is the per-provider shutdown budget used when Observe tears
// down already-started providers after a later step fails. Chosen to be
// large enough for an OTLP collector flush but small enough that a startup
// failure does not block boot for long.
const unwindTimeout = 5 * time.Second

// noopStop is the StopFn returned when Observe fails before any provider has
// started successfully, or by each provider's setup function on error. It is
// a no-op that ignores its arguments and never panics.
func noopStop(_ context.Context, _ time.Duration) {}
