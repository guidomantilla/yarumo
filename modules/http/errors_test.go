package http

import (
	"errors"
	"strings"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats with type and cause", func(t *testing.T) {
		t.Parallel()

		inner := errors.New("boom")
		e := &Error{TypedError: cerrs.TypedError{Type: ServerType, Err: inner}}

		got := e.Error()

		wantPrefix := "http server " + ServerType + " error: "
		if !strings.HasPrefix(got, wantPrefix) {
			t.Fatalf("Error() prefix = %q, want prefix %q", got, wantPrefix)
		}

		if !strings.Contains(got, "boom") {
			t.Fatalf("Error() = %q, want it to contain inner cause %q", got, "boom")
		}
	})
}

func TestErrServer(t *testing.T) {
	t.Parallel()

	t.Run("wraps cause with sentinel", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("listen failed")
		err := ErrServer(cause)

		if err == nil {
			t.Fatal("ErrServer returned nil")
		}

		if !errors.Is(err, ErrHttpServerFailed) {
			t.Fatalf("expected errors.Is(err, ErrHttpServerFailed) to be true; got false")
		}

		if !errors.Is(err, cause) {
			t.Fatalf("expected errors.Is(err, cause) to be true; got false")
		}
	})

	t.Run("wraps no causes", func(t *testing.T) {
		t.Parallel()

		err := ErrServer()

		if err == nil {
			t.Fatal("ErrServer returned nil")
		}

		if !errors.Is(err, ErrHttpServerFailed) {
			t.Fatalf("expected errors.Is(err, ErrHttpServerFailed) to be true; got false")
		}
	})

	t.Run("returns *Error domain type", func(t *testing.T) {
		t.Parallel()

		err := ErrServer(errors.New("x"))

		var de *Error
		if !errors.As(err, &de) {
			t.Fatalf("expected errors.As to *Error; got false")
		}

		if de.Type != ServerType {
			t.Fatalf("Type = %q, want %q", de.Type, ServerType)
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrHttpServerFailed has expected message", func(t *testing.T) {
		t.Parallel()

		if got := ErrHttpServerFailed.Error(); got != "http server failed" {
			t.Fatalf("ErrHttpServerFailed.Error() = %q", got)
		}
	})
}
