package diagnostics

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error type constants for the diagnostics package.
const (
	// ProfileCapture is the error type for pprof capture operations.
	ProfileCapture = "profile_capture"
)

// Type compliance.
var (
	_ error = (*Error)(nil)
)

// Error is the domain error for the diagnostics package.
type Error struct {
	cerrs.TypedError
}

// Error returns a formatted error string including the error type and cause.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("diagnostics %s error: %s", e.Type, e.Err)
}

// Sentinel errors for the diagnostics package.
var (
	// ErrCaptureFailed indicates a pprof profile capture operation failed.
	ErrCaptureFailed = errors.New("profile capture failed")
	// ErrWriterNil indicates a nil io.Writer was passed to a capture function.
	ErrWriterNil = errors.New("writer is nil")
	// ErrContextNil indicates a nil context.Context was passed to a capture function.
	ErrContextNil = errors.New("context is nil")
	// ErrDurationNonPositive indicates a non-positive duration was passed to CaptureCPUProfile.
	ErrDurationNonPositive = errors.New("duration must be positive")
	// ErrProfileNotFound indicates the requested pprof profile is not registered.
	ErrProfileNotFound = errors.New("profile not found")
)

// ErrCaptureProfile wraps the given errors into a domain Error for pprof capture failures.
func ErrCaptureProfile(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ProfileCapture,
			Err:  errors.Join(append(causes, ErrCaptureFailed)...),
		},
	}
}
