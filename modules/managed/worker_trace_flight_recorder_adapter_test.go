package managed

import (
	"context"
	"errors"
	"testing"

	cdiagnostics "github.com/guidomantilla/yarumo/common/diagnostics"
)

func TestNewTraceFlightRecorderWorker(t *testing.T) {
	t.Parallel()

	fr := &cdiagnostics.PluggableTraceFlightRecorder{
		StartFn: func() error { return nil },
		StopFn:  func() {},
	}
	worker := NewTraceFlightRecorderWorker(fr)
	if worker == nil {
		t.Fatal("expected non-nil worker")
	}
}

func Test_traceFlightRecorderWorker_Start(t *testing.T) {
	t.Parallel()

	t.Run("start succeeds", func(t *testing.T) {
		t.Parallel()

		fr := &cdiagnostics.PluggableTraceFlightRecorder{
			StartFn: func() error { return nil },
		}
		worker := NewTraceFlightRecorderWorker(fr)

		err := worker.Start(context.Background())
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("start returns error", func(t *testing.T) {
		t.Parallel()

		fr := &cdiagnostics.PluggableTraceFlightRecorder{
			StartFn: func() error { return errors.New("start failed") },
		}
		worker := NewTraceFlightRecorderWorker(fr)

		err := worker.Start(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func Test_traceFlightRecorderWorker_Stop(t *testing.T) {
	t.Parallel()

	t.Run("stop completes successfully", func(t *testing.T) {
		t.Parallel()

		fr := &cdiagnostics.PluggableTraceFlightRecorder{
			StopFn: func() {},
		}
		worker := NewTraceFlightRecorderWorker(fr)

		err := worker.Stop(context.Background())
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		select {
		case <-worker.Done():
		default:
			t.Fatal("expected done channel to be closed")
		}
	})

	t.Run("stop with canceled context returns error and closes done", func(t *testing.T) {
		t.Parallel()

		stopCh := make(chan struct{})
		fr := &cdiagnostics.PluggableTraceFlightRecorder{
			StopFn: func() { <-stopCh },
		}
		worker := NewTraceFlightRecorderWorker(fr)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := worker.Stop(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		close(stopCh)

		select {
		case <-worker.Done():
		default:
			t.Fatal("expected done channel to be closed after timeout")
		}
	})
}

func Test_traceFlightRecorderWorker_Done(t *testing.T) {
	t.Parallel()

	fr := &cdiagnostics.PluggableTraceFlightRecorder{
		StopFn: func() {},
	}
	worker := NewTraceFlightRecorderWorker(fr)

	ch := worker.Done()
	if ch == nil {
		t.Fatal("expected non-nil done channel")
	}

	select {
	case <-ch:
		t.Fatal("expected done channel to be open")
	default:
	}
}
