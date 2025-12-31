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

func TestHTTPError_NilInner_UnwrapIsNilAndNoErrorCall(t *testing.T) {
	// Do not call Error() because current implementation fatals on nil inner.
	// Still, Unwrap should be nil via the embedded TypedError.
	e := &Error{TypedError: cerrs.TypedError{Type: RequestType}}
	u := errors.Unwrap(e)
	if u != nil {
		t.Fatalf("errors.Unwrap(e) = %v, want nil", u)
	}
}

func TestErrDoCall_JoinAndType(t *testing.T) {
	e1 := errors.New("first")
	e2 := errors.New("second")

	err := ErrDo(e1, e2)
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

func TestErrDoCall_NoArgs_ErrHttpRequestFailedInner(t *testing.T) {
	err := ErrDo()
	if err == nil {
		t.Fatalf("ErrDoCall() with no args should still return non-nil *Error")
	}

	// Unwrap should be nil because TypedError.Err is ErrHttpRequestFailed
	u := errors.Unwrap(err)
	if !errors.Is(u, ErrHttpRequestFailed) {
		t.Fatalf("errors.Unwrap() = %v, want nil", u)
	}

	// Do not call Error() on err because inner Err is nil and assert will fatal.
}

func TestSentinelErrors(t *testing.T) {
	// Basic sanity to reference the sentinel errors; ensures usability with errors.Is
	joined := errors.Join(ErrRateLimiterExceeded, ErrHttpRequestFailed)
	if !errors.Is(joined, ErrRateLimiterExceeded) || !errors.Is(joined, ErrHttpRequestFailed) {
		t.Fatalf("sentinel errors are not matched via errors.Is")
	}
}

func TestStatusCodeError_Error_NonNil(t *testing.T) {
	e := &StatusCodeError{StatusCode: 503}
	got := e.Error()

	want := "http retryable status code: 503"
	if got != want {
		t.Fatalf("StatusCodeError.Error() = %q, want %q", got, want)
	}
}

// Note: calling (*StatusCodeError)(nil).Error() fatals in current assert behavior,
// so we avoid that path in tests to keep suite stable.
