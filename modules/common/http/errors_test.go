package http

import (
	"errors"
	"strings"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

func TestHTTPError_ErrorFormatting_NonNil(t *testing.T) {
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
}

func TestHTTPError_ErrorFormatting_NilInner(t *testing.T) {
	// When inner Err is nil, Error() should still be safe to call and return a non-empty string
	e := &Error{TypedError: cerrs.TypedError{Type: RequestType}}
	got := e.Error()
	wantPrefix := "http request " + RequestType + " error: "
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("Error() prefix = %q, want prefix %q", got, wantPrefix)
	}
	// Implementation uses %s with a nil error; depending on fmt, this typically renders as %!s(<nil>)
	// We only assert that the suffix is non-empty to avoid coupling to fmt details.
	if len(got) == len(wantPrefix) {
		t.Fatalf("Error() produced empty suffix for nil inner: %q", got)
	}
}

func TestErrDoCall_JoinAndType(t *testing.T) {
	e1 := errors.New("first")
	e2 := errors.New("second")

	err := ErrDoCall(e1, e2)
	if err == nil {
		t.Fatalf("ErrDoCall returned nil")
	}

	// It should be an *Error with the expected Type
	var he *Error
	if !errors.As(err, &he) || he == nil {
		t.Fatalf("errors.As to *Error failed: %T", err)
	}
	if he.Type != RequestType {
		t.Fatalf("Type = %q, want %q", he.Type, RequestType)
	}

	// And errors.It should match the joined components
	if !errors.Is(err, e1) || !errors.Is(err, e2) {
		t.Fatalf("joined error does not match components: %v", err)
	}

	// Calling Error() should include both parts in some form
	msg := err.Error()
	if !strings.Contains(msg, "first") || !strings.Contains(msg, "second") {
		t.Fatalf("Error() does not include components: %q", msg)
	}
}

func TestErrDoCall_NoArgs_NilInner(t *testing.T) {
	err := ErrDoCall()
	if err == nil {
		t.Fatalf("ErrDoCall() with no args should still return non-nil *Error")
	}

	// Unwrap should be nil because TypedError.Err is nil
	if u := errors.Unwrap(err); u != nil {
		t.Fatalf("errors.Unwrap() = %v, want nil", u)
	}

	// Error() should be well-formed with the expected prefix
	msg := err.Error()
	wantPrefix := "http request " + RequestType + " error: "
	if !strings.HasPrefix(msg, wantPrefix) {
		t.Fatalf("Error() prefix = %q, want prefix %q", msg, wantPrefix)
	}
}

func TestSentinelErrors(t *testing.T) {
	// Basic sanity to reference the sentinel errors; ensures usability with errors.Is
	joined := errors.Join(ErrRateLimiterExceeded, ErrHttpRequestFailed)
	if !errors.Is(joined, ErrRateLimiterExceeded) || !errors.Is(joined, ErrHttpRequestFailed) {
		t.Fatalf("sentinel errors are not matched via errors.Is")
	}
}
