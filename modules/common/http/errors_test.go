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
		e := &Error{TypedError: cerrs.TypedError{Type: RequestType, Err: inner}}

		got := e.Error()

		wantPrefix := "http request " + RequestType + " error: "
		if !strings.HasPrefix(got, wantPrefix) {
			t.Fatalf("Error() prefix = %q, want prefix %q", got, wantPrefix)
		}

		if !strings.Contains(got, "boom") {
			t.Fatalf("Error() should contain inner error message; got %q", got)
		}
	})

	t.Run("nil inner unwraps to nil", func(t *testing.T) {
		t.Parallel()

		e := &Error{TypedError: cerrs.TypedError{Type: RequestType}}

		u := errors.Unwrap(e)
		if u != nil {
			t.Fatalf("errors.Unwrap(e) = %v, want nil", u)
		}
	})
}

func TestStatusCodeError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats status code", func(t *testing.T) {
		t.Parallel()

		e := &StatusCodeError{StatusCode: 503}
		got := e.Error()

		want := "http retryable status code: 503"
		if got != want {
			t.Fatalf("StatusCodeError.Error() = %q, want %q", got, want)
		}
	})
}

func TestErrDo(t *testing.T) {
	t.Parallel()

	t.Run("joins errors with type", func(t *testing.T) {
		t.Parallel()

		e1 := errors.New("first")
		e2 := errors.New("second")

		err := ErrDo(e1, e2)
		if err == nil {
			t.Fatalf("ErrDo returned nil")
		}

		var he *Error

		ok := errors.As(err, &he)
		if !ok || he == nil {
			t.Fatalf("errors.As to *Error failed: %T", err)
		}

		if he.Type != RequestType {
			t.Fatalf("Type = %q, want %q", he.Type, RequestType)
		}

		if !errors.Is(err, e1) || !errors.Is(err, e2) {
			t.Fatalf("joined error does not match components: %v", err)
		}

		msg := err.Error()
		if !strings.Contains(msg, "first") || !strings.Contains(msg, "second") {
			t.Fatalf("Error() does not include components: %q", msg)
		}
	})

	t.Run("no args wraps ErrHttpRequestFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrDo()
		if err == nil {
			t.Fatalf("ErrDo() with no args should still return non-nil *Error")
		}

		u := errors.Unwrap(err)
		if !errors.Is(u, ErrHttpRequestFailed) {
			t.Fatalf("errors.Unwrap() = %v, want ErrHttpRequestFailed", u)
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("matched via errors.Is", func(t *testing.T) {
		t.Parallel()

		joined := errors.Join(ErrRateLimiterExceeded, ErrHttpRequestFailed)
		if !errors.Is(joined, ErrRateLimiterExceeded) || !errors.Is(joined, ErrHttpRequestFailed) {
			t.Fatalf("sentinel errors are not matched via errors.Is")
		}
	})
}
