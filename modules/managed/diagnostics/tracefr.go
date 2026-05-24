package diagnostics

import (
	"context"
	"io"
	"runtime/trace"
	"sync"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// tracefr is the canonical TraceFlightRecorder implementation. It
// embeds the stdlib *runtime/trace.FlightRecorder by pointer so the
// underlying buffer state is shared, not copied. The Start and Stop
// methods of trace.FlightRecorder are shadowed by lifecycle-aware
// versions; Enabled and WriteTo are promoted directly from the embed.
type tracefr struct {
	*trace.FlightRecorder

	name string
	done chan struct{}
	once sync.Once
}

// NewTraceFlightRecorder creates a TraceFlightRecorder with the given
// name and options. The name is used in logs and lifecycle events; it
// must be non-empty.
func NewTraceFlightRecorder(name string, options ...Option) TraceFlightRecorder {
	cassert.NotEmpty(name, "name is empty")

	opts := NewOptions(options...)

	internal := trace.NewFlightRecorder(trace.FlightRecorderConfig{
		MinAge:   opts.minAge,
		MaxBytes: opts.maxBytes,
	})

	return &tracefr{
		FlightRecorder: internal,
		name:           name,
		done:           make(chan struct{}),
	}
}

// Name returns the recorder's identity used in logs.
func (t *tracefr) Name() string {
	cassert.NotNil(t, "tracefr is nil")

	return t.name
}

// Start enables the runtime trace flight-recorder buffer and returns
// immediately. It satisfies the lifecycle.Component worker-style
// contract; Done is closed after Stop completes.
func (t *tracefr) Start(_ context.Context) error {
	cassert.NotNil(t, "tracefr is nil")

	return t.FlightRecorder.Start()
}

// Stop disables the runtime trace flight-recorder buffer and closes
// Done. It is idempotent. The stdlib FlightRecorder.Stop returns no
// error, so Stop always returns nil.
func (t *tracefr) Stop(_ context.Context) error {
	cassert.NotNil(t, "tracefr is nil")

	defer t.once.Do(func() { close(t.done) })

	t.FlightRecorder.Stop()

	return nil
}

// Done returns the channel that is closed after Stop has been called.
func (t *tracefr) Done() <-chan struct{} {
	cassert.NotNil(t, "tracefr is nil")

	return t.done
}

// Enabled reports whether the recorder is currently capturing.
func (t *tracefr) Enabled() bool {
	cassert.NotNil(t, "tracefr is nil")

	return t.FlightRecorder.Enabled()
}

// WriteTo writes the captured trace data to the given writer.
func (t *tracefr) WriteTo(w io.Writer) (int64, error) {
	cassert.NotNil(t, "tracefr is nil")

	return t.FlightRecorder.WriteTo(w)
}
