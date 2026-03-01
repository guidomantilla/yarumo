package servers

import (
	"errors"
	"strings"
	"testing"
)

func TestServerErrors_WrappingAndType(t *testing.T) {
	base := errors.New("boom")

	e1 := ErrServerFailedToStart("x", base)
	if se, ok := e1.(*Error); !ok || se.Type != ServerStartType {
		t.Fatalf("expected start type, got %#v", e1)
	}
	if !errors.Is(e1, base) {
		t.Fatal("expected error to wrap base")
	}

	e2 := ErrServerFailedToStop("x", base)
	if se, ok := e2.(*Error); !ok || se.Type != ServerStopType {
		t.Fatalf("expected stop type, got %#v", e2)
	}
	if !errors.Is(e2, base) {
		t.Fatal("expected error to wrap base")
	}
}

func TestError_ErrorMessage_StartAndStop(t *testing.T) {
	base := errors.New("root-cause")

	// Start error message
	e1 := ErrServerFailedToStart("api", base)
	if e1 == nil {
		t.Fatal("nil error")
	}
	// Call the Error() method explicitly to cover it and validate the message text
	msg1 := e1.Error()
	if msg1 == "" || !containsAll(msg1, []string{"server", "api", "failed to start", "root-cause"}) {
		t.Fatalf("unexpected start error message: %q", msg1)
	}

	// Stop error message
	e2 := ErrServerFailedToStop("api", base)
	if e2 == nil {
		t.Fatal("nil error")
	}
	msg2 := e2.Error()
	if msg2 == "" || !containsAll(msg2, []string{"server", "api", "failed to stop", "root-cause"}) {
		t.Fatalf("unexpected stop error message: %q", msg2)
	}
}

// containsAll checks that s contains every substring from parts.
func containsAll(s string, parts []string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}
