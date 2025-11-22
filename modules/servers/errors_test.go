package servers

import (
	"errors"
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
