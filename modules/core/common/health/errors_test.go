package health

import (
	"errors"
	"strings"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats with type and cause", func(t *testing.T) {
		t.Parallel()

		inner := errors.New("boom")
		e := &Error{TypedError: cerrs.TypedError{Type: HealthType, Err: inner}}

		got := e.Error()

		wantPrefix := "health " + HealthType + " error: "
		if !strings.HasPrefix(got, wantPrefix) {
			t.Fatalf("Error() prefix = %q, want prefix %q", got, wantPrefix)
		}

		if !strings.Contains(got, "boom") {
			t.Fatalf("Error() should contain inner error message; got %q", got)
		}
	})

	t.Run("nil inner unwraps to nil", func(t *testing.T) {
		t.Parallel()

		e := &Error{TypedError: cerrs.TypedError{Type: HealthType}}

		u := errors.Unwrap(e)
		if u != nil {
			t.Fatalf("errors.Unwrap(e) = %v, want nil", u)
		}
	})
}

func TestErrHealth(t *testing.T) {
	t.Parallel()

	t.Run("joins errors with type", func(t *testing.T) {
		t.Parallel()

		e1 := errors.New("first")
		e2 := errors.New("second")

		err := ErrHealth(e1, e2)
		if err == nil {
			t.Fatalf("ErrHealth returned nil")
		}

		var he *Error

		ok := errors.As(err, &he)
		if !ok || he == nil {
			t.Fatalf("errors.As to *Error failed: %T", err)
		}

		if he.Type != HealthType {
			t.Fatalf("Type = %q, want %q", he.Type, HealthType)
		}

		if !errors.Is(err, e1) || !errors.Is(err, e2) {
			t.Fatalf("joined error does not match components: %v", err)
		}

		if !errors.Is(err, ErrHealthFailed) {
			t.Fatalf("joined error must wrap ErrHealthFailed; got %v", err)
		}

		msg := err.Error()
		if !strings.Contains(msg, "first") || !strings.Contains(msg, "second") {
			t.Fatalf("Error() does not include components: %q", msg)
		}
	})

	t.Run("no args wraps ErrHealthFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrHealth()
		if err == nil {
			t.Fatalf("ErrHealth() with no args should still return non-nil *Error")
		}

		if !errors.Is(err, ErrHealthFailed) {
			t.Fatalf("errors.Is(err, ErrHealthFailed) = false")
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("matched via errors.Is", func(t *testing.T) {
		t.Parallel()

		joined := errors.Join(ErrCheckNil, ErrContextNil, ErrHealthFailed)
		if !errors.Is(joined, ErrCheckNil) {
			t.Fatalf("errors.Is(joined, ErrCheckNil) = false")
		}

		if !errors.Is(joined, ErrContextNil) {
			t.Fatalf("errors.Is(joined, ErrContextNil) = false")
		}

		if !errors.Is(joined, ErrHealthFailed) {
			t.Fatalf("errors.Is(joined, ErrHealthFailed) = false")
		}
	})
}
