package sets

import "testing"

func TestNew_empty(t *testing.T) {
	t.Parallel()

	s := New[int]()

	if s.Len() != 0 {
		t.Fatalf("expected empty set, got len %d", s.Len())
	}
}

func TestNew_withItems(t *testing.T) {
	t.Parallel()

	s := New(1, 2, 3)

	if s.Len() != 3 {
		t.Fatalf("expected 3, got %d", s.Len())
	}
}

func TestNew_deduplicates(t *testing.T) {
	t.Parallel()

	s := New(1, 1, 2, 2, 3)

	if s.Len() != 3 {
		t.Fatalf("expected 3 unique, got %d", s.Len())
	}
}

func TestNew_strings(t *testing.T) {
	t.Parallel()

	s := New("a", "b", "c")

	if s.Len() != 3 {
		t.Fatalf("expected 3, got %d", s.Len())
	}

	if !s.Contains("a") {
		t.Fatal("expected set to contain 'a'")
	}
}
