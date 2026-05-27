package breaker

import (
	"errors"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

func TestErrBreakerRejected(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrBreakerRejectedFailed without causes", func(t *testing.T) {
		t.Parallel()

		err := ErrBreakerRejected()
		if !errors.Is(err, ErrBreakerRejectedFailed) {
			t.Fatalf("expected wrap of ErrBreakerRejectedFailed, got %v", err)
		}
	})

	t.Run("wraps additional causes", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("breaker open")
		err := ErrBreakerRejected(cause)

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}
		if !errors.Is(err, ErrBreakerRejectedFailed) {
			t.Fatalf("expected wrap of ErrBreakerRejectedFailed, got %v", err)
		}
	})

	t.Run("returns *Error with BreakerTransportType tag", func(t *testing.T) {
		t.Parallel()

		err := ErrBreakerRejected()
		var e *Error
		ok := errors.As(err, &e)
		if !ok {
			t.Fatalf("expected *Error, got %T", err)
		}
		if e.Type != BreakerTransportType {
			t.Fatalf("Type = %q, want %q", e.Type, BreakerTransportType)
		}
	})
}

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats type and inner error", func(t *testing.T) {
		t.Parallel()

		e := &Error{
			TypedError: cerrs.TypedError{
				Type: BreakerTransportType,
				Err:  errors.New("boom"),
			},
		}

		got := e.Error()
		want := BreakerTransportType + " error: boom"

		if got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})
}

func TestStatusCodeError_Error(t *testing.T) {
	t.Parallel()

	e := &StatusCodeError{StatusCode: 503}
	got := e.Error()
	want := "http status 503 reported as breaker failure"
	if got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}
