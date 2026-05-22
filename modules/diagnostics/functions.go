package diagnostics

import (
	"context"
	"io"
	"runtime/pprof"
	"time"

	"github.com/guidomantilla/yarumo/common/lifecycle"
	clog "github.com/guidomantilla/yarumo/common/log"
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
// been called with a non-zero rate. Use BlockProfiling (or call
// runtime.SetBlockProfileRate directly) to enable sampling. Block profiling
// carries non-trivial overhead — every blocking event is sampled — so it is
// typically left off in production and toggled on for short investigation
// windows.
func CaptureBlockProfile(w io.Writer) error {
	return captureNamedProfile("block", w)
}

// BuildTraceFlightRecorder creates a managed TraceFlightRecorder, starts
// it in a background goroutine, and returns a CloseFn for graceful shutdown.
//
// Startup errors are logged and forwarded to errChan (non-blocking). The
// returned recorder can be used to inspect Enabled() and dump the buffer
// via WriteTo. The returned CloseFn must be called by the caller to
// release runtime resources; it bounds the shutdown by the given
// timeout and blocks until the background goroutine has exited.
func BuildTraceFlightRecorder(ctx context.Context, name string, errChan lifecycle.ErrChan, options ...Option) (TraceFlightRecorder, lifecycle.CloseFn, error) {
	clog.Info(ctx, "starting up", "stage", "startup", "component", name)

	component := NewTraceFlightRecorder(name, options...)

	spawned := make(chan struct{})

	closeFn := func(ctx context.Context, timeout time.Duration) {
		clog.Info(ctx, "stopping", "stage", "shutdown", "component", name)
		defer clog.Info(ctx, "stopped", "stage", "shutdown", "component", name)

		err := lifecycle.Stop(ctx, component, timeout)
		if err != nil {
			clog.Error(ctx, "shutdown failed", "stage", "shutdown", "component", name, "error", err)
		}

		<-spawned
	}

	go func() {
		defer close(spawned)

		err := lifecycle.Start(ctx, component, errChan)
		if err != nil {
			clog.Error(ctx, "failed to start", "stage", "startup", "component", name, "error", err)
		}
	}()

	return component, closeFn, nil
}

// BuildBlockProfiling creates a managed BlockProfiling sampler, starts
// it in a background goroutine, and returns a CloseFn for graceful
// shutdown.
//
// Startup errors are logged and forwarded to errChan (non-blocking).
// The returned sampler can be used to inspect Rate(). The returned
// CloseFn must be called by the caller to disable sampling
// (runtime.SetBlockProfileRate(0)); it bounds the shutdown by the
// given timeout and blocks until the background goroutine has exited.
func BuildBlockProfiling(ctx context.Context, name string, errChan lifecycle.ErrChan, options ...Option) (BlockProfiling, lifecycle.CloseFn, error) {
	clog.Info(ctx, "starting up", "stage", "startup", "component", name)

	component := NewBlockProfiling(name, options...)

	spawned := make(chan struct{})

	closeFn := func(ctx context.Context, timeout time.Duration) {
		clog.Info(ctx, "stopping", "stage", "shutdown", "component", name)
		defer clog.Info(ctx, "stopped", "stage", "shutdown", "component", name)

		err := lifecycle.Stop(ctx, component, timeout)
		if err != nil {
			clog.Error(ctx, "shutdown failed", "stage", "shutdown", "component", name, "error", err)
		}

		<-spawned
	}

	go func() {
		defer close(spawned)

		err := lifecycle.Start(ctx, component, errChan)
		if err != nil {
			clog.Error(ctx, "failed to start", "stage", "startup", "component", name, "error", err)
		}
	}()

	return component, closeFn, nil
}
