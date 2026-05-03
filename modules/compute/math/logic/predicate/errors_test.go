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
