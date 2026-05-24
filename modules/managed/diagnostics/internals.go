package diagnostics

import (
	"io"
	"runtime/pprof"
)

// captureNamedProfile is the shared implementation for snapshot-style profiles
// (heap, goroutine, block, mutex, threadcreate, allocs).
func captureNamedProfile(name string, w io.Writer) error {
	if w == nil {
		return ErrCaptureProfile(ErrWriterNil)
	}

	profile := pprof.Lookup(name)
	if profile == nil {
		return ErrCaptureProfile(ErrProfileNotFound)
	}

	err := profile.WriteTo(w, 0)
	if err != nil {
		return ErrCaptureProfile(err)
	}

	return nil
}
