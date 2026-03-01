package errs

import (
	"errors"
	"testing"
)

func TestNewTypedError(t *testing.T) {
	t.Parallel()

	t.Run("creates error with type and inner error", func(t *testing.T) {
		t.Parallel()

		inner := errors.New("boom")

		err := NewTypedError("IO", inner)
		if err == nil {
			t.Fatal("expected non-nil error")
		}

		got := err.Error()
		if got != "IO error: boom" {
			t.Fatalf("Error() = %q, want %q", got, "IO error: boom")
		}
	})

	t.Run("unwraps to inner error", func(t *testing.T) {
		t.Parallel()

		inner := errors.New("cause")
		err := NewTypedError("DB", inner)

		var unwrapper interface{ Unwrap() error }
		if !errors.As(err, &unwrapper) {
			t.Fatal("expected Unwrap interface")
		}

		if !errors.Is(unwrapper.Unwrap(), inner) {
			t.Fatal("Unwrap did not return inner error")
		}
	})

	t.Run("exposes ErrorType via interface", func(t *testing.T) {
		t.Parallel()

		err := NewTypedError("NET", errors.New("timeout"))

		te, ok := err.(interface{ ErrorType() string })
		if !ok {
			t.Fatal("expected ErrorType interface")
		}

		got := te.ErrorType()
		if got != "NET" {
			t.Fatalf("ErrorType() = %q, want %q", got, "NET")
		}
	})
}

func TestTypedError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats type and message", func(t *testing.T) {
		t.Parallel()

		e := &TypedError{Type: "IO", Err: errors.New("boom")}

		got := e.Error()
		if got != "IO error: boom" {
			t.Fatalf("Error() = %q, want %q", got, "IO error: boom")
		}
	})

	t.Run("formats with empty type", func(t *testing.T) {
		t.Parallel()

		e := &TypedError{Type: "", Err: errors.New("oops")}

		got := e.Error()
		if got != " error: oops" {
			t.Fatalf("Error() = %q, want %q", got, " error: oops")
		}
	})

	t.Run("formats with multi-line inner error", func(t *testing.T) {
		t.Parallel()

		inner := errors.Join(errors.New("a"), errors.New("b"))
		e := &TypedError{Type: "X", Err: inner}

		got := e.Error()

		want := "X error: a\nb"
		if got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})
}

func TestTypedError_Unwrap(t *testing.T) {
	t.Parallel()

	t.Run("returns inner error", func(t *testing.T) {
		t.Parallel()

		inner := errors.New("wrapped")
		e := &TypedError{Type: "X", Err: inner}

		got := e.Unwrap()
		if !errors.Is(got, inner) {
			t.Fatalf("Unwrap() = %v, want %v", got, inner)
		}
	})

	t.Run("returns nil when inner is nil", func(t *testing.T) {
		t.Parallel()

		e := &TypedError{Type: "X"}

		got := e.Unwrap()
		if got != nil {
			t.Fatalf("Unwrap() = %v, want nil", got)
		}
	})
}

func TestTypedError_ErrorType(t *testing.T) {
	t.Parallel()

	t.Run("returns the type string", func(t *testing.T) {
		t.Parallel()

		e := &TypedError{Type: "DB", Err: errors.New("fail")}

		got := e.ErrorType()
		if got != "DB" {
			t.Fatalf("ErrorType() = %q, want %q", got, "DB")
		}
	})

	t.Run("returns empty string when type is empty", func(t *testing.T) {
		t.Parallel()

		e := &TypedError{Type: "", Err: errors.New("fail")}

		got := e.ErrorType()
		if got != "" {
			t.Fatalf("ErrorType() = %q, want %q", got, "")
		}
	})
}
