package predicate

import (
	"errors"
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/logic"
)

// --- ForAll tests ---

func TestForAll_allTrue(t *testing.T) {
	t.Parallel()

	docs := Collection{
		logic.Fact{logic.Var("vigente"): true, logic.Var("firmado"): true},
		logic.Fact{logic.Var("vigente"): true, logic.Var("firmado"): false},
	}

	got, err := ForAll(docs, logic.Var("vigente"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !got {
		t.Fatal("expected true when all elements satisfy predicate")
	}
}

func TestForAll_oneFalse(t *testing.T) {
	t.Parallel()

	docs := Collection{
		logic.Fact{logic.Var("vigente"): true, logic.Var("firmado"): true},
		logic.Fact{logic.Var("vigente"): false, logic.Var("firmado"): true},
	}

	got, err := ForAll(docs, logic.Var("vigente"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got {
		t.Fatal("expected false when one element fails predicate")
	}
}

func TestForAll_emptyCollection(t *testing.T) {
	t.Parallel()

	got, err := ForAll(Collection{}, logic.Var("vigente"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !got {
		t.Fatal("expected true (vacuous truth) for empty collection")
	}
}

func TestForAll_nilPredicate(t *testing.T) {
	t.Parallel()

	docs := Collection{
		logic.Fact{logic.Var("vigente"): true},
	}

	_, err := ForAll(docs, nil)
	if !errors.Is(err, ErrNilPredicate) {
		t.Fatalf("expected ErrNilPredicate, got %v", err)
	}
}

// --- Exists tests ---

func TestExists_oneTrue(t *testing.T) {
	t.Parallel()

	docs := Collection{
		logic.Fact{logic.Var("vigente"): false, logic.Var("firmado"): true},
		logic.Fact{logic.Var("vigente"): true, logic.Var("firmado"): false},
		logic.Fact{logic.Var("vigente"): false, logic.Var("firmado"): false},
	}

	got, err := Exists(docs, logic.Var("vigente"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !got {
		t.Fatal("expected true when at least one element satisfies predicate")
	}
}

func TestExists_noneSatisfy(t *testing.T) {
	t.Parallel()

	docs := Collection{
		logic.Fact{logic.Var("vigente"): false, logic.Var("firmado"): true},
		logic.Fact{logic.Var("vigente"): false, logic.Var("firmado"): false},
	}

	got, err := Exists(docs, logic.Var("vigente"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got {
		t.Fatal("expected false when no element satisfies predicate")
	}
}

func TestExists_emptyCollection(t *testing.T) {
	t.Parallel()

	got, err := Exists(Collection{}, logic.Var("vigente"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got {
		t.Fatal("expected false (vacuous falsity) for empty collection")
	}
}

func TestExists_nilPredicate(t *testing.T) {
	t.Parallel()

	docs := Collection{
		logic.Fact{logic.Var("vigente"): true},
	}

	_, err := Exists(docs, nil)
	if !errors.Is(err, ErrNilPredicate) {
		t.Fatalf("expected ErrNilPredicate, got %v", err)
	}
}

// --- Count tests ---

func TestCount_allMatch(t *testing.T) {
	t.Parallel()

	docs := Collection{
		logic.Fact{logic.Var("vigente"): true},
		logic.Fact{logic.Var("vigente"): true},
		logic.Fact{logic.Var("vigente"): true},
	}

	got, err := Count(docs, logic.Var("vigente"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}

func TestCount_someMatch(t *testing.T) {
	t.Parallel()

	docs := Collection{
		logic.Fact{logic.Var("vigente"): true, logic.Var("firmado"): true},
		logic.Fact{logic.Var("vigente"): false, logic.Var("firmado"): true},
		logic.Fact{logic.Var("vigente"): true, logic.Var("firmado"): false},
	}

	got, err := Count(docs, logic.Var("vigente"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestCount_noneMatch(t *testing.T) {
	t.Parallel()

	docs := Collection{
		logic.Fact{logic.Var("vigente"): false},
		logic.Fact{logic.Var("vigente"): false},
	}

	got, err := Count(docs, logic.Var("vigente"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestCount_emptyCollection(t *testing.T) {
	t.Parallel()

	got, err := Count(Collection{}, logic.Var("vigente"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != 0 {
		t.Fatalf("expected 0 for empty collection, got %d", got)
	}
}

func TestCount_nilPredicate(t *testing.T) {
	t.Parallel()

	docs := Collection{
		logic.Fact{logic.Var("vigente"): true},
	}

	_, err := Count(docs, nil)
	if !errors.Is(err, ErrNilPredicate) {
		t.Fatalf("expected ErrNilPredicate, got %v", err)
	}
}

// --- Filter tests ---

func TestFilter_someMatch(t *testing.T) {
	t.Parallel()

	docs := Collection{
		logic.Fact{logic.Var("vigente"): true, logic.Var("firmado"): true},
		logic.Fact{logic.Var("vigente"): false, logic.Var("firmado"): true},
		logic.Fact{logic.Var("vigente"): true, logic.Var("firmado"): false},
	}

	got, err := Filter(docs, logic.Var("vigente"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 matching elements, got %d", len(got))
	}
}

func TestFilter_noneMatch(t *testing.T) {
	t.Parallel()

	docs := Collection{
		logic.Fact{logic.Var("vigente"): false, logic.Var("firmado"): true},
		logic.Fact{logic.Var("vigente"): false, logic.Var("firmado"): false},
	}

	got, err := Filter(docs, logic.Var("vigente"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("expected 0 matching elements, got %d", len(got))
	}
}

func TestFilter_allMatch(t *testing.T) {
	t.Parallel()

	docs := Collection{
		logic.Fact{logic.Var("vigente"): true, logic.Var("firmado"): true},
		logic.Fact{logic.Var("vigente"): true, logic.Var("firmado"): false},
	}

	got, err := Filter(docs, logic.Var("vigente"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 matching elements, got %d", len(got))
	}
}

func TestFilter_emptyCollection(t *testing.T) {
	t.Parallel()

	got, err := Filter(Collection{}, logic.Var("vigente"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != nil {
		t.Fatalf("expected nil for empty collection, got %v", got)
	}
}

func TestFilter_nilPredicate(t *testing.T) {
	t.Parallel()

	docs := Collection{
		logic.Fact{logic.Var("vigente"): true},
	}

	_, err := Filter(docs, nil)
	if !errors.Is(err, ErrNilPredicate) {
		t.Fatalf("expected ErrNilPredicate, got %v", err)
	}
}
