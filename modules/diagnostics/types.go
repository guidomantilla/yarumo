// Package diagnostics provides profiling and tracing tools for Go
// applications. It bundles one-shot pprof capture helpers (`CaptureCPU/
// Heap/Goroutine/BlockProfile`) and stateless HTTP exposition
// (`NewPprofHandler`) for ad-hoc collection.
//
// It also exposes lifecycle-aware diagnostics components that
// implement `common/lifecycle.Component`: `TraceFlightRecorder` wraps
// the Go 1.25+ `runtime/trace.FlightRecorder` (continuous trace
// buffering, dumped on demand), and `BlockProfiling` owns the
// `runtime.SetBlockProfileRate` lifecycle so block-profile sampling is
// enabled by `Start` and disabled by `Stop`. Builders
// `BuildTraceFlightRecorder` and `BuildBlockProfiling` wrap
// construction with the managed-component idiom and return a
// `lifecycle.CloseFn` for graceful shutdown.
//
// Error contract: capture operations wrap errors into a domain `Error`
// type with `ProfileCapture` as the type constant. Callers should
// prefer `errors.Is`/`As` over string matching. Lifecycle errors come
// from `common/lifecycle` (`ErrStart`, `ErrShutdown`). Concurrency:
// all public types and free functions are safe for concurrent use by
// multiple goroutines.
package diagnostics

import (
	"context"
	"io"
	"time"

	"github.com/guidomantilla/yarumo/common/lifecycle"
)

var (
	_ TraceFlightRecorder = (*tracefr)(nil)
	_ TraceFlightRecorder = (*PluggableTraceFlightRecorder)(nil)
	_ BlockProfiling      = (*blockprof)(nil)

	_ BuildTraceFlightRecorderFn = BuildTraceFlightRecorder
	_ BuildBlockProfilingFn      = BuildBlockProfiling
	_ ErrCaptureProfileFn        = ErrCaptureProfile

	_ CaptureCPUFn       = CaptureCPUProfile
	_ CaptureHeapFn      = CaptureHeapProfile
	_ CaptureGoroutineFn = CaptureGoroutineProfile
	_ CaptureBlockFn     = CaptureBlockProfile
)

// BuildTraceFlightRecorderFn is the function type for BuildTraceFlightRecorder.
type BuildTraceFlightRecorderFn func(ctx context.Context, name string, errChan lifecycle.ErrChan, options ...Option) (TraceFlightRecorder, lifecycle.CloseFn, error)

// BuildBlockProfilingFn is the function type for BuildBlockProfiling.
type BuildBlockProfilingFn func(ctx context.Context, name string, errChan lifecycle.ErrChan, options ...Option) (BlockProfiling, lifecycle.CloseFn, error)

// ErrCaptureProfileFn is the function type for ErrCaptureProfile.
type ErrCaptureProfileFn func(causes ...error) error

// CaptureCPUFn is the function type for CaptureCPUProfile.
type CaptureCPUFn func(ctx context.Context, w io.Writer, duration time.Duration) error

// CaptureHeapFn is the function type for CaptureHeapProfile.
type CaptureHeapFn func(w io.Writer) error

// CaptureGoroutineFn is the function type for CaptureGoroutineProfile.
type CaptureGoroutineFn func(w io.Writer) error

// CaptureBlockFn is the function type for CaptureBlockProfile.
type CaptureBlockFn func(w io.Writer) error

// TraceFlightRecorder defines the interface for a flight recorder that
// captures Go execution traces continuously into an in-memory buffer.
//
// Only one flight recorder may be active per process at a time. The
// recorder is a lifecycle component: Start enables the runtime trace
// buffer, Stop disables it, Done closes after Stop. Enabled reports
// whether the recorder is currently capturing. WriteTo dumps the
// current buffer contents to the given writer. The caller is
// responsible for calling Stop to release runtime resources.
// Implementations must be safe for concurrent use by multiple goroutines.
type TraceFlightRecorder interface {
	lifecycle.Component
	// Enabled reports whether the recorder is currently capturing.
	Enabled() bool
	// WriteTo writes the captured trace data to the given writer.
	WriteTo(w io.Writer) (n int64, err error)
}

// BlockProfiling defines the interface for a managed block-profile
// sampler.
//
// Block profiling carries non-trivial overhead — every blocking event
// is sampled — so it is typically left off in production and toggled
// on for short investigation windows. BlockProfiling owns that
// toggle: Start invokes `runtime.SetBlockProfileRate(rate)` and Stop
// resets it to zero. Done closes after Stop. The caller is responsible
// for calling Stop to disable sampling. Implementations must be safe
// for concurrent use by multiple goroutines.
type BlockProfiling interface {
	lifecycle.Component
	// Rate returns the configured block-profile sampling rate.
	Rate() int
}
