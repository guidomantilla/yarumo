package diagnostics

import (
	"bytes"
	"context"
	"testing"
	"time"
)

// TestNewTraceFlightRecorder constructs the recorder and checks basic
// invariants that do not require starting the runtime trace machinery.
//
// Subtests run in series — Enabled() and Start/Stop touch the
// process-global trace.FlightRecorder slot, which only allows one
// active recorder at a time.
func TestNewTraceFlightRecorder(t *testing.T) {
	t.Run("returns non-nil recorder", func(t *testing.T) {
		r := NewTraceFlightRecorder("tracefr-1")
		if r == nil {
			t.Fatal("expected non-nil recorder")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		r := NewTraceFlightRecorder("tracefr-named")
		if r.Name() != "tracefr-named" {
			t.Fatalf("expected name %q, got %q", "tracefr-named", r.Name())
		}
	})

	t.Run("with custom option", func(t *testing.T) {
		r := NewTraceFlightRecorder("tracefr-opts", WithMinAge(30*time.Second))
		if r == nil {
			t.Fatal("expected non-nil recorder")
		}
	})

	t.Run("not enabled before start", func(t *testing.T) {
		r := NewTraceFlightRecorder("tracefr-disabled")
		if r.Enabled() {
			t.Fatal("expected recorder not to be enabled before start")
		}
	})

	t.Run("done channel is open at construction", func(t *testing.T) {
		r := NewTraceFlightRecorder("tracefr-open-done")

		select {
		case <-r.Done():
			t.Fatal("expected Done channel to be open before Start/Stop")
		default:
		}
	})
}

// TestTraceFlightRecorder_Lifecycle exercises Start/Stop/Enabled/WriteTo.
// Only one trace.FlightRecorder can be active per process so subtests
// run sequentially.
func TestTraceFlightRecorder_Lifecycle(t *testing.T) {
	t.Run("start and stop", func(t *testing.T) {
		r := NewTraceFlightRecorder("tracefr-lifecycle")

		err := r.Start(context.Background())
		if err != nil {
			t.Fatalf("unexpected start error: %v", err)
		}

		if !r.Enabled() {
			t.Fatal("expected recorder to be enabled after start")
		}

		err = r.Stop(context.Background())
		if err != nil {
			t.Fatalf("unexpected stop error: %v", err)
		}

		if r.Enabled() {
			t.Fatal("expected recorder not to be enabled after stop")
		}

		select {
		case <-r.Done():
		default:
			t.Fatal("expected Done channel closed after Stop")
		}
	})

	t.Run("stop is idempotent", func(t *testing.T) {
		r := NewTraceFlightRecorder("tracefr-stop-twice")

		err := r.Start(context.Background())
		if err != nil {
			t.Fatalf("unexpected start error: %v", err)
		}

		err = r.Stop(context.Background())
		if err != nil {
			t.Fatalf("first Stop returned %v", err)
		}

		err = r.Stop(context.Background())
		if err != nil {
			t.Fatalf("second Stop returned %v", err)
		}
	})

	t.Run("write to while enabled", func(t *testing.T) {
		r := NewTraceFlightRecorder("tracefr-write")

		err := r.Start(context.Background())
		if err != nil {
			t.Fatalf("unexpected start error: %v", err)
		}

		defer func() {
			_ = r.Stop(context.Background())
		}()

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
