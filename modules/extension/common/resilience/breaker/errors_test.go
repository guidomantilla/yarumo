package breaker

import (
	"errors"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

func TestErrBreaker(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrBreakerFailed without causes", func(t *testing.T) {
		t.Parallel()

		err := ErrBreaker()
		if !errors.Is(err, ErrBreakerFailed) {
			t.Fatalf("expected wrap of ErrBreakerFailed, got %v", err)
		}
	})

	t.Run("wraps additional causes", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("upstream-500")
		err := ErrBreaker(cause)

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}
		if !errors.Is(err, ErrBreakerFailed) {
			t.Fatalf("expected wrap of ErrBreakerFailed, got %v", err)
		}
	})

	t.Run("returns *Error with BreakerType tag", func(t *testing.T) {
		t.Parallel()

		err := ErrBreaker()
		var e *Error
		ok := errors.As(err, &e)
		if !ok {
			t.Fatalf("expected *Error, got %T", err)
		}
		if e.Type != BreakerType {
			t.Fatalf("Type = %q, want %q", e.Type, BreakerType)
		}
	})
}

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats type and inner error", func(t *testing.T) {
		t.Parallel()

		e := &Error{
			TypedError: cerrs.TypedError{
				Type: BreakerType,
				Err:  errors.New("boom"),
			},
		}

		got := e.Error()
		want := BreakerType + " error: boom"

		if got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})
}
