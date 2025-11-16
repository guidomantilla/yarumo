package errs

import (
	"errors"
	"testing"
)

func TestTypedError_Error(t *testing.T) {
	t.Run("formats with non-nil inner error", func(t *testing.T) {
		inner := errors.New("boom")
		e := &TypedError{Type: "IO", Err: inner}
		if got := e.Error(); got != "IO error: boom" {
			t.Fatalf("Error() = %q, want %q", got, "IO error: boom")
		}
	})
}

func TestTypedError_Unwrap(t *testing.T) {
	// non-nil returns inner error
	inner := errors.New("wrapped")
	e := &TypedError{Type: "X", Err: inner}
	if got := e.Unwrap(); !errors.Is(got, inner) {
		t.Fatalf("Unwrap() = %v, want %v", got, inner)
	}

	// non-nil with nil inner -> returns nil
	eNilInner := &TypedError{Type: "X"}
	if got := eNilInner.Unwrap(); got != nil {
		t.Fatalf("Unwrap() with nil inner = %v, want nil", got)
	}
}
