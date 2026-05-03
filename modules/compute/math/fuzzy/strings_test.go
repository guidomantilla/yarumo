package fuzzy

import "testing"

func TestDegree_String(t *testing.T) {
	t.Parallel()

	d := Degree(0.75)

	if d.String() != "0.75" {
		t.Fatalf("expected 0.75, got %q", d.String())
	}
}

func TestDegree_String_zero(t *testing.T) {
	t.Parallel()

	d := Degree(0)

	if d.String() != "0" {
		t.Fatalf("expected 0, got %q", d.String())
	}
}

func TestDegree_String_one(t *testing.T) {
	t.Parallel()

	d := Degree(1)

	if d.String() != "1" {
		t.Fatalf("expected 1, got %q", d.String())
	}
}

func TestSet_String(t *testing.T) {
	t.Parallel()

	s := Set{Name: "cold"}

	if s.String() != "Set(cold)" {
		t.Fatalf("expected Set(cold), got %q", s.String())
	}
}

func TestSet_String_empty(t *testing.T) {
	t.Parallel()

	s := Set{}

	if s.String() != "Set()" {
		t.Fatalf("expected Set(), got %q", s.String())
	}
}
