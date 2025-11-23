package hashes

import (
	"errors"
	"strings"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

func TestErrHashFunctionNotFound(t *testing.T) {
	name := "generate"
	err := ErrHashFunctionNotFound(name)
	if err == nil {
		t.Fatalf("ErrHashFunctionNotFound returned nil")
	}

	// Type assertion to *UIDError
	ue, ok := err.(*Error)
	if !ok {
		t.Fatalf("error is not *UIDError: %T", err)
	}

	// Check exported constant and inner error message
	if ue.Type != HashNotFound {
		t.Fatalf("Type = %q, want %q", ue.Type, HashNotFound)
	}
	innerMsg := "hash function " + name + " not found"
	if ue.Err == nil || ue.Err.Error() != innerMsg {
		t.Fatalf("inner error message = %v, want %q", ue.Err, innerMsg)
	}

	// Check Error() formatting overrides the embedded type
	expected := "hash " + HashNotFound + " error: " + innerMsg
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

func TestUIDError_ErrorVariants(t *testing.T) {
	t.Run("panics when receiver is nil", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("expected panic when calling Error() on nil receiver")
			}
		}()
		var ue *Error
		_ = ue.Error() // should trigger assert.NotEmpty(e, "error is nil")
	})

	t.Run("formats when inner error is nil", func(t *testing.T) {
		ue := &Error{TypedError: cerrs.TypedError{Type: "custom"}}
		got := ue.Error()
		// With Err == nil and %s formatting, fmt prints %!s(<nil>)
		want := "hash custom error: %!s(<nil>)"
		if got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
		// Unwrap should return nil when inner Err is nil
		if u := errors.Unwrap(ue); u != nil {
			t.Fatalf("Unwrap() = %v, want <nil>", u)
		}
	})
}
