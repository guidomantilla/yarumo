package retry

import (
	"errors"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

func TestErrRetry(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrRetryFailed without causes", func(t *testing.T) {
		t.Parallel()

		err := ErrRetry()
		if !errors.Is(err, ErrRetryFailed) {
			t.Fatalf("expected wrap of ErrRetryFailed, got %v", err)
		}
	})

	t.Run("wraps additional causes", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("upstream-timeout")
		err := ErrRetry(cause)

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}
		if !errors.Is(err, ErrRetryFailed) {
			t.Fatalf("expected wrap of ErrRetryFailed, got %v", err)
		}
	})

	t.Run("returns *Error with RetryType tag", func(t *testing.T) {
		t.Parallel()

		err := ErrRetry()
		var e *Error
		ok := errors.As(err, &e)
		if !ok {
			t.Fatalf("expected *Error, got %T", err)
		}
		if e.Type != RetryType {
			t.Fatalf("Type = %q, want %q", e.Type, RetryType)
		}
	})
}

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats type and inner error", func(t *testing.T) {
		t.Parallel()

		e := &Error{
			TypedError: cerrs.TypedError{
				Type: RetryType,
				Err:  errors.New("boom"),
			},
		}

		got := e.Error()
		want := RetryType + " error: boom"

		if got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})
}
