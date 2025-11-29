package uids

import (
	"errors"
	"strings"
	"testing"
)

func TestErrUIDFunctionNotFound(t *testing.T) {
	name := "generate"
	err := ErrUIDFunctionNotFound(name)
	if err == nil {
		t.Fatalf("ErrUIDFunctionNotFound returned nil")
	}

	// Type assertion to *UIDError
	ue, ok := err.(*Error)
	if !ok {
		t.Fatalf("error is not *UIDError: %T", err)
	}

	// Check exported constant and inner error message
	if ue.Type != UIDNotFound {
		t.Fatalf("Type = %q, want %q", ue.Type, UIDNotFound)
	}
	innerMsg := "uid function " + name + " not found"
	if ue.Err == nil || ue.Err.Error() != innerMsg {
		t.Fatalf("inner error message = %v, want %q", ue.Err, innerMsg)
	}

	// Check Error() formatting overrides the embedded type
	expected := "uid " + UIDNotFound + " error: " + innerMsg
	if got := ue.Error(); got != expected {
		t.Fatalf("Error() = %q, want %q", got, expected)
	}

	// errors.As should capture *Error
	var target *Error
	if !errors.As(err, &target) || target == nil {
		t.Fatalf("errors.As to *UIDError failed")
	}

	// errors.Unwrap should return the inner wrapped error via promoted Unwrap()
	u := errors.Unwrap(err)
	if u == nil || !strings.Contains(u.Error(), name) {
		t.Fatalf("errors.Unwrap() = %v, want message containing %q", u, name)
	}
}
