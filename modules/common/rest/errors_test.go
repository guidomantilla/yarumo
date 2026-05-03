package rest

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats error message with type and cause", func(t *testing.T) {
		t.Parallel()

		err := ErrCall(errors.New("root-cause"))
		if !strings.Contains(err.Error(), "rest request rest-request error: root-cause") {
			t.Fatalf("unexpected error string: %s", err.Error())
		}
	})

	t.Run("wraps with ErrRestCallFailed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrCall(errors.New("inner"))
		if !errors.Is(err, ErrRestCallFailed) {
			t.Fatal("expected error to wrap ErrRestCallFailed")
		}
	})
}

func TestHTTPError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats status code and status text", func(t *testing.T) {
		t.Parallel()

		he := &HTTPError{StatusCode: 418, Status: "I'm a teapot"}

		got := he.Error()
		if !strings.Contains(got, "unexpected status code 418") {
			t.Fatalf("unexpected http error string: %s", got)
		}
	})
}

func TestDecodeResponseError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats content type and target type", func(t *testing.T) {
		t.Parallel()

		de := &DecodeResponseError[sample]{ContentType: "text/plain", T: sample{}}

		got := de.Error()
		if !strings.Contains(got, "content type text/plain not supported") {
			t.Fatalf("unexpected decode error string: %s", got)
		}
	})
}

func TestErrCall(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error wrapping provided errors", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("cause")

		err := ErrCall(cause)

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !errors.Is(err, cause) {
			t.Fatal("expected error to wrap cause")
		}

		if !errors.Is(err, ErrRestCallFailed) {
			t.Fatal("expected error to wrap ErrRestCallFailed")
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrRestCallFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrRestCallFailed == nil {
			t.Fatal("ErrRestCallFailed should not be nil")
		}
	})

	t.Run("ErrContextNil is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrContextNil == nil {
			t.Fatal("ErrContextNil should not be nil")
		}
	})

	t.Run("ErrRequestSpecNil is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrRequestSpecNil == nil {
			t.Fatal("ErrRequestSpecNil should not be nil")
		}
	})

	t.Run("ErrResponseTooLarge is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrResponseTooLarge == nil {
			t.Fatal("ErrResponseTooLarge should not be nil")
		}
	})
}
