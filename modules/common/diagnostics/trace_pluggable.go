package diagnostics

import "io"

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
