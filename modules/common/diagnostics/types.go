// Package diagnostics provides profiling and tracing tools for Go applications.
package diagnostics

import (
	"context"
	"io"
	"time"
)

// CaptureCPUFn is the signature of CaptureCPUProfile.
type CaptureCPUFn func(ctx context.Context, w io.Writer, duration time.Duration) error

// CaptureHeapFn is the signature of CaptureHeapProfile.
type CaptureHeapFn func(w io.Writer) error

// CaptureGoroutineFn is the signature of CaptureGoroutineProfile.
type CaptureGoroutineFn func(w io.Writer) error

// CaptureBlockFn is the signature of CaptureBlockProfile.
type CaptureBlockFn func(w io.Writer) error

var (
	_ TraceFlightRecorder = (*tracefr)(nil)
	_ TraceFlightRecorder = (*PluggableTraceFlightRecorder)(nil)

	_ CaptureCPUFn       = CaptureCPUProfile
	_ CaptureHeapFn      = CaptureHeapProfile
	_ CaptureGoroutineFn = CaptureGoroutineProfile
	_ CaptureBlockFn     = CaptureBlockProfile
)

// TraceFlightRecorder defines the interface for a flight recorder that captures execution traces.
// Only one flight recorder may be active at a time. The caller must call Stop to release
// resources when the recorder is no longer needed. Implementations must be safe for concurrent use.
type TraceFlightRecorder interface {
	// Start begins capturing execution traces.
	Start() error
	// Stop halts trace capture.
	Stop()
	// Enabled reports whether the recorder is currently capturing.
	Enabled() bool
	// WriteTo writes the captured trace data to the given writer.
	WriteTo(w io.Writer) (n int64, err error)
}

// PluggableTraceFlightRecorder is a TraceFlightRecorder implementation with pluggable function fields for testing and composition.
type PluggableTraceFlightRecorder struct {
	StartFn   func() error
	StopFn    func()
	EnabledFn func() bool
	WriteToFn func(w io.Writer) (int64, error)
}

// Start begins capturing execution traces.
func (p *PluggableTraceFlightRecorder) Start() error {
	if p.StartFn != nil {
		return p.StartFn()
	}

	return nil
}

// Stop halts trace capture.
func (p *PluggableTraceFlightRecorder) Stop() {
	if p.StopFn != nil {
		p.StopFn()
	}
}

// Enabled reports whether the recorder is currently capturing.
func (p *PluggableTraceFlightRecorder) Enabled() bool {
	if p.EnabledFn != nil {
		return p.EnabledFn()
	}

	return false
}

// WriteTo writes the captured trace data to the given writer.
func (p *PluggableTraceFlightRecorder) WriteTo(w io.Writer) (int64, error) {
	if p.WriteToFn != nil {
		return p.WriteToFn(w)
	}

	return 0, nil
}
