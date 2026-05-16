package diagnostics

import (
	"bytes"
	"testing"
	"time"
)

func TestNewTraceFlightRecorder(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil recorder", func(t *testing.T) {
		t.Parallel()

		r := NewTraceFlightRecorder()
		if r == nil {
			t.Fatal("expected non-nil recorder")
		}
	})

	t.Run("with custom option", func(t *testing.T) {
		t.Parallel()

		r := NewTraceFlightRecorder(WithMinAge(30 * time.Second))

		if r == nil {
			t.Fatal("expected non-nil recorder")
		}
	})

	t.Run("not enabled before start", func(t *testing.T) {
		t.Parallel()

		r := NewTraceFlightRecorder()
		if r.Enabled() {
			t.Fatal("expected recorder not to be enabled before start")
		}
	})
}

// TestTraceFlightRecorder_Lifecycle tests Start, Stop, Enabled, and WriteTo.
// These tests are sequential because only one trace.FlightRecorder can be active per process.
func TestTraceFlightRecorder_Lifecycle(t *testing.T) {
	t.Run("start and stop", func(t *testing.T) {
		r := NewTraceFlightRecorder()

		err := r.Start()
		if err != nil {
			t.Fatalf("unexpected start error: %v", err)
		}

		if !r.Enabled() {
			t.Fatal("expected recorder to be enabled after start")
		}

		r.Stop()

		if r.Enabled() {
			t.Fatal("expected recorder not to be enabled after stop")
		}
	})

	t.Run("write to while enabled", func(t *testing.T) {
		r := NewTraceFlightRecorder()

		err := r.Start()
		if err != nil {
			t.Fatalf("unexpected start error: %v", err)
		}

		defer r.Stop()

		var buf bytes.Buffer

		n, err := r.WriteTo(&buf)
		if err != nil {
			t.Fatalf("unexpected write error: %v", err)
		}

		if n == 0 {
			t.Fatal("expected non-zero bytes written")
		}
	})
}
