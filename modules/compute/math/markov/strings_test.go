package markov

import (
	"testing"
)

func TestStateClass_String_transient(t *testing.T) {
	t.Parallel()

	if Transient.String() != "Transient" {
		t.Fatalf("expected %q, got %q", "Transient", Transient.String())
	}
}

func TestStateClass_String_recurrent(t *testing.T) {
	t.Parallel()

	if Recurrent.String() != "Recurrent" {
		t.Fatalf("expected %q, got %q", "Recurrent", Recurrent.String())
	}
}

func TestStateClass_String_absorbing(t *testing.T) {
	t.Parallel()

	if Absorbing.String() != "Absorbing" {
		t.Fatalf("expected %q, got %q", "Absorbing", Absorbing.String())
	}
}

func TestStateClass_String_unknown(t *testing.T) {
	t.Parallel()

	unknown := StateClass(99)

	if unknown.String() != "Unknown" {
		t.Fatalf("expected %q, got %q", "Unknown", unknown.String())
	}
}
