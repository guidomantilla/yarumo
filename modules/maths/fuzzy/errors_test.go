package fuzzy

import "testing"

func TestErrInvalidDegree(t *testing.T) {
	t.Parallel()

	if ErrInvalidDegree == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrInvalidDegree.Error() != "degree must be in [0,1]" {
		t.Fatalf("unexpected message: %s", ErrInvalidDegree.Error())
	}
}

func TestErrEmptySamples(t *testing.T) {
	t.Parallel()

	if ErrEmptySamples == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrEmptySamples.Error() != "empty sample set" {
		t.Fatalf("unexpected message: %s", ErrEmptySamples.Error())
	}
}

func TestErrInvalidRange(t *testing.T) {
	t.Parallel()

	if ErrInvalidRange == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrInvalidRange.Error() != "invalid range" {
		t.Fatalf("unexpected message: %s", ErrInvalidRange.Error())
	}
}
