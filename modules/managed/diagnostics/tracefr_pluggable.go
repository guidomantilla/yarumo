package diagnostics

import (
	"context"
	"io"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

// PluggableTraceFlightRecorder is a TraceFlightRecorder implementation
// with pluggable function fields for testing and composition. Every
// behavior is configurable: when a function field is nil the method
// returns a safe zero value.
type PluggableTraceFlightRecorder struct {
	NameFn    func() string
	StartFn   func(ctx context.Context) error
	StopFn    func(ctx context.Context) error
	DoneFn    func() <-chan struct{}
	EnabledFn func() bool
	WriteToFn func(w io.Writer) (int64, error)

	done chan struct{}
	once sync.Once
}

// Name returns the configured name; falls back to an empty string when
// NameFn is nil.
func (p *PluggableTraceFlightRecorder) Name() string {
	cassert.NotNil(p, "PluggableTraceFlightRecorder is nil")

	if p.NameFn != nil {
		return p.NameFn()
	}

	return ""
}

// Start delegates to StartFn; returns nil when StartFn is nil.
func (p *PluggableTraceFlightRecorder) Start(ctx context.Context) error {
	cassert.NotNil(p, "PluggableTraceFlightRecorder is nil")

	if p.StartFn != nil {
		return p.StartFn(ctx)
	}

	return nil
}

// Stop delegates to StopFn and closes the internal Done channel; both
// are idempotent across multiple calls.
func (p *PluggableTraceFlightRecorder) Stop(ctx context.Context) error {
	cassert.NotNil(p, "PluggableTraceFlightRecorder is nil")

	defer p.once.Do(func() {
		if p.done != nil {
			close(p.done)
		}
	})

	if p.StopFn != nil {
		return p.StopFn(ctx)
	}

	return nil
}

// Done returns DoneFn's channel when configured, otherwise an internal
// channel closed after the first Stop call.
func (p *PluggableTraceFlightRecorder) Done() <-chan struct{} {
	cassert.NotNil(p, "PluggableTraceFlightRecorder is nil")

	if p.DoneFn != nil {
		return p.DoneFn()
	}

	if p.done == nil {
		p.done = make(chan struct{})
	}

	return p.done
}

// Enabled delegates to EnabledFn; returns false when EnabledFn is nil.
func (p *PluggableTraceFlightRecorder) Enabled() bool {
	cassert.NotNil(p, "PluggableTraceFlightRecorder is nil")

	if p.EnabledFn != nil {
		return p.EnabledFn()
	}

	return false
}

// WriteTo delegates to WriteToFn; returns (0, nil) when WriteToFn is nil.
func (p *PluggableTraceFlightRecorder) WriteTo(w io.Writer) (int64, error) {
	cassert.NotNil(p, "PluggableTraceFlightRecorder is nil")

	if p.WriteToFn != nil {
		return p.WriteToFn(w)
	}

	return 0, nil
}
