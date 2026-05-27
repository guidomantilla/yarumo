package limiter

import (
	"errors"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

func TestErrWait(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrWaitFailed without causes", func(t *testing.T) {
		t.Parallel()

		err := ErrWait()
		if !errors.Is(err, ErrWaitFailed) {
			t.Fatalf("expected wrap of ErrWaitFailed, got %v", err)
		}
	})

	t.Run("wraps additional causes", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("ctx deadline exceeded")
		err := ErrWait(cause)

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}
		if !errors.Is(err, ErrWaitFailed) {
			t.Fatalf("expected wrap of ErrWaitFailed, got %v", err)
		}
	})

	t.Run("returns *Error with LimiterType tag", func(t *testing.T) {
		t.Parallel()

		err := ErrWait()
		var e *Error
		ok := errors.As(err, &e)
		if !ok {
			t.Fatalf("expected *Error, got %T", err)
		}
		if e.Type != LimiterType {
			t.Fatalf("Type = %q, want %q", e.Type, LimiterType)
		}
	})
}

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats type and inner error", func(t *testing.T) {
		t.Parallel()

		e := &Error{
			TypedError: cerrs.TypedError{
				Type: LimiterType,
				Err:  errors.New("boom"),
			},
		}

		got := e.Error()
		want := LimiterType + " error: boom"

		if got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})
}
