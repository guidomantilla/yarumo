package uids

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats uid error message", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("ABC")

		got := err.Error()
		if !strings.Contains(got, "uid") {
			t.Fatalf("expected 'uid' in error: %q", got)
		}

		if !strings.Contains(got, "ABC") {
			t.Fatalf("expected 'ABC' in error: %q", got)
		}

		if !strings.Contains(got, UidNotFound) {
			t.Fatalf("expected %q in error: %q", UidNotFound, got)
		}
	})
}

func TestErrAlgorithmNotSupported(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil error", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("XYZ")
		if err == nil {
			t.Fatal("expected non-nil error")
		}
	})

	t.Run("error is of type Error", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("XYZ")

		var e *Error
		if !errors.As(err, &e) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("contains algorithm name in message", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("UNKNOWN")
		if !strings.Contains(err.Error(), "UNKNOWN") {
			t.Fatalf("expected algorithm name in error: %q", err.Error())
		}
	})

	t.Run("uses uid algorithm wording", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("FOO")
		if !strings.Contains(err.Error(), "uid algorithm") {
			t.Fatalf("expected 'uid algorithm' in error: %q", err.Error())
		}
	})
}
