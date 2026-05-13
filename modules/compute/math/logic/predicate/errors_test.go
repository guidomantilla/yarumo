package predicate

import (
	"errors"
	"testing"
)

func TestErrEmptyCollection(t *testing.T) {
	t.Parallel()

	got := ErrEmptyCollection.Error()
	if got != "collection is empty" {
		t.Fatalf("expected %q, got %q", "collection is empty", got)
	}
}

func TestErrNilPredicate(t *testing.T) {
	t.Parallel()

	got := ErrNilPredicate.Error()
	if got != "predicate is nil" {
		t.Fatalf("expected %q, got %q", "predicate is nil", got)
	}
}

func TestErrPredicate(t *testing.T) {
	t.Parallel()

	err := ErrPredicate(ErrEmptyCollection)
	if !errors.Is(err, ErrEmptyCollection) {
		t.Fatal("expected ErrEmptyCollection")
	}
}

func TestErrPredicate_type(t *testing.T) {
	t.Parallel()

	err := ErrPredicate(ErrNilPredicate)

	var predErr *Error

	if !errors.As(err, &predErr) {
		t.Fatal("expected *Error type")
	}

	if predErr.Type != PredicateType {
		t.Fatalf("expected type %s, got %s", PredicateType, predErr.Type)
	}
}

func TestErrPredicate_zeroArgs(t *testing.T) {
	t.Parallel()

	err := ErrPredicate()
	if !errors.Is(err, ErrPredicateFailed) {
		t.Fatal("expected ErrPredicateFailed in chain")
	}
}

func TestErrPredicateFailed(t *testing.T) {
	t.Parallel()

	if ErrPredicateFailed == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrPredicateFailed.Error() != "predicate operation failed" {
		t.Fatalf("unexpected message: %s", ErrPredicateFailed.Error())
	}
}
