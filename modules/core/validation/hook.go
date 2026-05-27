package validation

import (
	"context"
	"time"

	clog "github.com/guidomantilla/yarumo/core/common/log"
)

// Hook is the observability extension point. Implementations are notified
// immediately before and after every leaf rule evaluation; the duration
// passed to AfterRule covers the leaf invocation only (not the path
// resolution or context evaluation).
type Hook interface {
	// BeforeRule fires before the engine invokes a leaf.
	BeforeRule(ctx context.Context, path, rule string, params []any)

	// AfterRule fires immediately after the leaf returns. err is nil on
	// success, the violation otherwise; took is the leaf evaluation latency.
	AfterRule(ctx context.Context, path, rule string, params []any, err error, took time.Duration)
}

// NoopHook is the zero-cost default. Engines without WithHook see this
// hook so the dispatch path never has to nil-check.
type NoopHook struct{}

// BeforeRule does nothing.
func (NoopHook) BeforeRule(_ context.Context, _, _ string, _ []any) {}

// AfterRule does nothing.
func (NoopHook) AfterRule(_ context.Context, _, _ string, _ []any, _ error, _ time.Duration) {}

// MultiHook fans out events to every wrapped hook in order. A nil entry is
// silently skipped so callers can pass `MultiHook(h1, nil, h2)` from
// option-driven construction.
type MultiHook []Hook

// BeforeRule fans out the call to every non-nil hook.
func (m MultiHook) BeforeRule(ctx context.Context, path, rule string, params []any) {
	for _, h := range m {
		if h == nil {
			continue
		}

		h.BeforeRule(ctx, path, rule, params)
	}
}

// AfterRule fans out the call to every non-nil hook.
func (m MultiHook) AfterRule(ctx context.Context, path, rule string, params []any, err error, took time.Duration) {
	for _, h := range m {
		if h == nil {
			continue
		}

		h.AfterRule(ctx, path, rule, params, err, took)
	}
}

// LoggingHook is the reference Hook backed by common/log. Each rule
// evaluation produces one debug line on AfterRule with path/rule/duration
// fields so production telemetry can grep / filter.
type LoggingHook struct {
	logger clog.Logger
}

// NewLoggingHook returns a LoggingHook bound to the given logger.
func NewLoggingHook(logger clog.Logger) *LoggingHook {
	return &LoggingHook{logger: logger}
}

// BeforeRule is intentionally a no-op — the after-line covers the same
// information plus the outcome.
func (h *LoggingHook) BeforeRule(_ context.Context, _, _ string, _ []any) {}

// AfterRule emits one debug log line per evaluated rule.
func (h *LoggingHook) AfterRule(ctx context.Context, path, rule string, _ []any, err error, took time.Duration) {
	if h == nil || h.logger == nil {
		return
	}

	if err != nil {
		h.logger.Debug(ctx, "validation rule failed",
			"path", path,
			"rule", rule,
			"took_ns", took.Nanoseconds(),
			"error", err.Error(),
		)

		return
	}

	h.logger.Debug(ctx, "validation rule passed",
		"path", path,
		"rule", rule,
		"took_ns", took.Nanoseconds(),
	)
}
