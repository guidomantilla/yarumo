package diagnostics

import (
	"context"
	"io"
	"runtime/pprof"
	"time"
)

// CaptureCPUProfile starts a CPU profile that writes to w for the given duration.
//
// The function blocks until either the duration elapses or the context is
// cancelled, whichever happens first. On context cancellation the CPU profile
// is stopped and the context error is wrapped into the returned domain error.
//
// CPU profiling has low overhead and can be enabled on demand in production.
func CaptureCPUProfile(ctx context.Context, w io.Writer, duration time.Duration) error {
	if ctx == nil {
		return ErrCaptureProfile(ErrContextNil)
	}

	if w == nil {
		return ErrCaptureProfile(ErrWriterNil)
	}

	if duration <= 0 {
		return ErrCaptureProfile(ErrDurationNonPositive)
	}

	err := pprof.StartCPUProfile(w)
	if err != nil {
		return ErrCaptureProfile(err)
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		pprof.StopCPUProfile()

		return nil
	case <-ctx.Done():
		pprof.StopCPUProfile()

		return ErrCaptureProfile(ctx.Err())
	}
}

// CaptureHeapProfile writes a heap profile snapshot to w. The heap profile is
// always sampled by the Go runtime, so this function has near-zero overhead
// beyond the cost of writing the snapshot.
func CaptureHeapProfile(w io.Writer) error {
	return captureNamedProfile("heap", w)
}

// CaptureGoroutineProfile writes a goroutine profile snapshot to w. This
// profile is always available; capture cost scales with the number of live
// goroutines.
func CaptureGoroutineProfile(w io.Writer) error {
	return captureNamedProfile("goroutine", w)
}

// CaptureBlockProfile writes a block profile snapshot to w.
//
// The block profile is only populated when runtime.SetBlockProfileRate has
// been called with a non-zero rate. Block profiling carries non-trivial
// overhead — every blocking event is sampled — so it is typically left off in
// production and toggled on for short investigation windows.
func CaptureBlockProfile(w io.Writer) error {
	return captureNamedProfile("block", w)
}
