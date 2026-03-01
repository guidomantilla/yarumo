package diagnostics

import (
	"io"
	"runtime/trace"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// tracefr is a wrapper around trace.FlightRecorder.
type tracefr struct {
	recorder *trace.FlightRecorder
}

// NewTraceFlightRecorder returns a new TraceFlightRecorder.
func NewTraceFlightRecorder(options ...Option) TraceFlightRecorder {
	opts := NewOptions(options...)

	return &tracefr{
		recorder: trace.NewFlightRecorder(trace.FlightRecorderConfig{
			MinAge:   opts.minAge,
			MaxBytes: opts.maxBytes,
		}),
	}
}

// Start begins capturing execution traces.
func (t *tracefr) Start() error {
	cassert.NotNil(t, "tracefr is nil")

	return t.recorder.Start()
}

// Stop halts trace capture.
func (t *tracefr) Stop() {
	cassert.NotNil(t, "tracefr is nil")
	t.recorder.Stop()
}

// Enabled reports whether the recorder is currently capturing.
func (t *tracefr) Enabled() bool {
	cassert.NotNil(t, "tracefr is nil")

	return t.recorder.Enabled()
}

// WriteTo writes the captured trace data to the given writer.
func (t *tracefr) WriteTo(w io.Writer) (int64, error) {
	cassert.NotNil(t, "tracefr is nil")

	return t.recorder.WriteTo(w)
}
