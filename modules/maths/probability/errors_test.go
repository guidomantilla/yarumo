package probability

import (
	"testing"
)

func TestErrInvalidProb(t *testing.T) {
	t.Parallel()

	if ErrInvalidProb == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrInvalidProb.Error() != "probability must be in [0,1]" {
		t.Fatalf("unexpected message: %s", ErrInvalidProb.Error())
	}
}

func TestErrNotNormalized(t *testing.T) {
	t.Parallel()

	if ErrNotNormalized == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrNotNormalized.Error() != "distribution does not sum to 1" {
		t.Fatalf("unexpected message: %s", ErrNotNormalized.Error())
	}
}

func TestErrEmptyDist(t *testing.T) {
	t.Parallel()

	if ErrEmptyDist == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrEmptyDist.Error() != "distribution is empty" {
		t.Fatalf("unexpected message: %s", ErrEmptyDist.Error())
	}
}

func TestErrOutcomeNotFound(t *testing.T) {
	t.Parallel()

	if ErrOutcomeNotFound == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrOutcomeNotFound.Error() != "outcome not found in distribution" {
		t.Fatalf("unexpected message: %s", ErrOutcomeNotFound.Error())
	}
}

func TestErrVariableNotFound(t *testing.T) {
	t.Parallel()

	if ErrVariableNotFound == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrVariableNotFound.Error() != "variable not found" {
		t.Fatalf("unexpected message: %s", ErrVariableNotFound.Error())
	}
}
