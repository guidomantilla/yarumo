package fuzzy

import (
	"errors"
	"testing"
)

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

func TestErrFuzzy(t *testing.T) {
	t.Parallel()

	err := ErrFuzzy(ErrEmptySamples)
	if !errors.Is(err, ErrEmptySamples) {
		t.Fatal("expected ErrEmptySamples")
	}
}

func TestErrFuzzy_type(t *testing.T) {
	t.Parallel()

	err := ErrFuzzy(ErrInvalidRange)

	var fuzzyErr *Error

	if !errors.As(err, &fuzzyErr) {
		t.Fatal("expected *Error type")
	}

	if fuzzyErr.Type != FuzzyType {
		t.Fatalf("expected type %s, got %s", FuzzyType, fuzzyErr.Type)
	}
}

func TestErrFuzzy_zeroArgs(t *testing.T) {
	t.Parallel()

	err := ErrFuzzy()
	if !errors.Is(err, ErrFuzzyFailed) {
		t.Fatal("expected ErrFuzzyFailed in chain")
	}
}

func TestErrFuzzyFailed(t *testing.T) {
	t.Parallel()

	if ErrFuzzyFailed == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrFuzzyFailed.Error() != "fuzzy operation failed" {
		t.Fatalf("unexpected message: %s", ErrFuzzyFailed.Error())
	}
}
