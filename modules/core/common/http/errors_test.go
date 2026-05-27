package http

import (
	"errors"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

func TestErrTransport(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrTransportFailed without causes", func(t *testing.T) {
		t.Parallel()

		err := ErrTransport()
		if !errors.Is(err, ErrTransportFailed) {
			t.Fatalf("expected wrap of ErrTransportFailed, got %v", err)
		}
	})

	t.Run("wraps additional causes", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("boom")
		err := ErrTransport(cause)

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}

		if !errors.Is(err, ErrTransportFailed) {
			t.Fatalf("expected wrap of ErrTransportFailed, got %v", err)
		}
	})

	t.Run("returns *Error with HTTPType tag", func(t *testing.T) {
		t.Parallel()

		err := ErrTransport()

		var e *Error
		ok := errors.As(err, &e)
		if !ok {
			t.Fatalf("expected *Error, got %T", err)
		}

		if e.Type != HTTPType {
			t.Fatalf("Type = %q, want %q", e.Type, HTTPType)
		}
	})
}

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats type and inner error", func(t *testing.T) {
		t.Parallel()

		e := &Error{
			TypedError: cerrs.TypedError{
				Type: HTTPType,
				Err:  errors.New("boom"),
			},
		}

		got := e.Error()
		want := "http error: boom"
		if got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})
}
