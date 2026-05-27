package limiter

import (
	"errors"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

func TestErrRateLimiterExceeded(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrRateLimiterFailed without causes", func(t *testing.T) {
		t.Parallel()

		err := ErrRateLimiterExceeded()
		if !errors.Is(err, ErrRateLimiterFailed) {
			t.Fatalf("expected wrap of ErrRateLimiterFailed, got %v", err)
		}
	})

	t.Run("wraps additional causes", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("ctx expired")
		err := ErrRateLimiterExceeded(cause)

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}
	})

	t.Run("returns *Error with LimiterType tag", func(t *testing.T) {
		t.Parallel()

		err := ErrRateLimiterExceeded()

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
		want := "http-limiter error: boom"
		if got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})
}
