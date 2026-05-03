package fsm

import (
	"testing"
)

func TestState_String(t *testing.T) {
	t.Parallel()

	s := State{ID: "idle"}
	expected := "State(idle)"

	if s.String() != expected {
		t.Fatalf("expected %q, got %q", expected, s.String())
	}
}

func TestTransition_String_no_guard(t *testing.T) {
	t.Parallel()

	tr := Transition{ID: "t1", From: "idle", To: "active", Event: "start"}
	expected := "Transition(t1: idle --[start]--> active)"

	if tr.String() != expected {
		t.Fatalf("expected %q, got %q", expected, tr.String())
	}
}

func TestTransition_String_with_guard(t *testing.T) {
	t.Parallel()

	tr := Transition{
		ID:    "t1",
		From:  "idle",
		To:    "active",
		Event: "start",
		Guard: func(any) bool { return true },
	}
	expected := "Transition(t1: idle --[start]--> active, guarded)"

	if tr.String() != expected {
		t.Fatalf("expected %q, got %q", expected, tr.String())
	}
}
