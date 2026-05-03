package managed

import (
	"context"
	"errors"
	"testing"
	"time"

	cdiagnostics "github.com/guidomantilla/yarumo/common/diagnostics"
)

func TestBuildTraceFlightRecorderWorker(t *testing.T) {
	t.Run("build succeeds and stop completes", func(t *testing.T) {
		errCh := make(chan error, 1)

		fr := &cdiagnostics.PluggableTraceFlightRecorder{
			StartFn: func() error { return nil },
			StopFn:  func() {},
		}

		component, stopFn, err := BuildTraceFlightRecorderWorker(context.Background(), "test-tfr", fr, errCh)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if component.name != "test-tfr" {
			t.Fatalf("expected name test-tfr, got %s", component.name)
		}

		if stopFn == nil {
			t.Fatal("expected non-nil stopFn")
		}

		time.Sleep(50 * time.Millisecond)

		stopFn(context.Background(), 5*time.Second)

		select {
		case err := <-errCh:
			t.Fatalf("unexpected error: %v", err)
		default:
		}
	})

	t.Run("start error is sent to errChan", func(t *testing.T) {
		errCh := make(chan error, 1)

		fr := &cdiagnostics.PluggableTraceFlightRecorder{
			StartFn: func() error { return errors.New("start failed") },
			StopFn:  func() {},
		}

		_, _, err := BuildTraceFlightRecorderWorker(context.Background(), "test-tfr-fail", fr, errCh)
		if err != nil {
			t.Fatalf("expected nil build error, got %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		select {
		case err := <-errCh:
			if err == nil {
				t.Fatal("expected non-nil error from errChan")
			}
		default:
			t.Fatal("expected error in errChan")
		}
	})

	t.Run("stop with short timeout logs error", func(t *testing.T) {
		errCh := make(chan error, 1)

		blockCh := make(chan struct{})
		fr := &cdiagnostics.PluggableTraceFlightRecorder{
			StartFn: func() error { return nil },
			StopFn:  func() { <-blockCh },
		}

		_, stopFn, err := BuildTraceFlightRecorderWorker(context.Background(), "test-tfr-timeout", fr, errCh)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		time.Sleep(50 * time.Millisecond)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		stopFn(ctx, time.Nanosecond)

		close(blockCh)
	})
}
