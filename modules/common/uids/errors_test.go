package uids

import (
	"fmt"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

func TestErrUIDFunctionNotFound(t *testing.T) {
	name := "generate"
	err := ErrUIDFunctionNotFound(name)
	if err == nil {
		t.Fatalf("ErrUIDFunctionNotFound returned nil")
	}

	// Since ErrUIDFunctionNotFound now returns a plain error, just check the message
	expected := "uid function " + name + " not found"
	if got := err.Error(); got != expected {
		t.Fatalf("Error() = %q, want %q", got, expected)
	}
}

func TestUIDError_ErrorFormatting(t *testing.T) {
	// Build a valid *Error with a non-nil inner error to avoid assert fatals
	inner := fmt.Errorf("boom")
	e := &Error{TypedError: cerrs.TypedError{Type: UIDNotFound, Err: inner}}

	got := e.Error()
	want := fmt.Sprintf("uid %s error: %s", UIDNotFound, inner)
	if got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}
