package diagnostics

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats error with type and cause", func(t *testing.T) {
		t.Parallel()

		err := ErrCaptureProfile(ErrWriterNil)

		got := err.Error()
		if !strings.Contains(got, "diagnostics") {
			t.Fatalf("expected 'diagnostics' in error, got %q", got)
		}

		if !strings.Contains(got, ProfileCapture) {
			t.Fatalf("expected type %q in error, got %q", ProfileCapture, got)
		}

		if !strings.Contains(got, "writer is nil") {
			t.Fatalf("expected cause message, got %q", got)
		}
	})
}

func TestErrCaptureProfile(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error", func(t *testing.T) {
		t.Parallel()

		err := ErrCaptureProfile(ErrWriterNil)

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("sets type to ProfileCapture", func(t *testing.T) {
		t.Parallel()

		err := ErrCaptureProfile(ErrWriterNil)

		var domErr *Error

		ok := errors.As(err, &domErr)
		if !ok {
			t.Fatal("expected *Error")
		}

		if domErr.Type != ProfileCapture {
			t.Fatalf("expected type %q, got %q", ProfileCapture, domErr.Type)
		}
	})

	t.Run("chains ErrCaptureFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrCaptureProfile(ErrWriterNil)

		if !errors.Is(err, ErrCaptureFailed) {
			t.Fatal("expected error to wrap ErrCaptureFailed")
		}
	})

	t.Run("preserves cause in chain", func(t *testing.T) {
		t.Parallel()

		err := ErrCaptureProfile(ErrWriterNil)

		if !errors.Is(err, ErrWriterNil) {
			t.Fatal("expected error to wrap ErrWriterNil")
		}
	})

	t.Run("supports multiple causes", func(t *testing.T) {
		t.Parallel()

		sentinel := errors.New("synthetic cause")

		err := ErrCaptureProfile(sentinel, ErrWriterNil)

		if !errors.Is(err, sentinel) {
			t.Fatal("expected error to wrap the synthetic cause")
		}

		if !errors.Is(err, ErrWriterNil) {
			t.Fatal("expected error to wrap ErrWriterNil")
		}

		if !errors.Is(err, ErrCaptureFailed) {
			t.Fatal("expected error to wrap ErrCaptureFailed")
		}
	})
}
