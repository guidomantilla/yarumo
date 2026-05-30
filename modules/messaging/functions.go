package messaging

import (
	"context"

	clog "github.com/guidomantilla/yarumo/core/common/log"
)

// DefaultErrorHandler is the hook installed by NewOptions when the
// caller does not pass WithErrorHandler. It logs every failure via
// common/log at Error level so handler bugs surface in standard
// telemetry without explicit caller wiring.
func DefaultErrorHandler(ctx context.Context, _ any, err error) {
	clog.Error(ctx, "messaging handler failed",
		"error", err.Error(),
	)
}

// SilentErrorHandler is a no-op ErrorHandler. Use it when the caller
// genuinely wants to suppress error logging — for example, in tests
// that intentionally drive failure paths.
func SilentErrorHandler(_ context.Context, _ any, _ error) {}
