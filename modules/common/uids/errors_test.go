package uids

import (
	"testing"
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
