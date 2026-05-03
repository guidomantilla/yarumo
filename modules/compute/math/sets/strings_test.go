package sets

import "testing"

func TestString_empty(t *testing.T) {
	t.Parallel()

	s := New[int]()

	if s.String() != "{}" {
		t.Fatalf("expected {}, got %s", s.String())
	}
}

func TestString_ints(t *testing.T) {
	t.Parallel()

	s := New(3, 1, 2)
	str := s.String()

	if str != "{1, 2, 3}" {
		t.Fatalf("expected {1, 2, 3}, got %s", str)
	}
}

func TestString_strings(t *testing.T) {
	t.Parallel()

	s := New("c", "a", "b")
	str := s.String()

	if str != "{a, b, c}" {
		t.Fatalf("expected {a, b, c}, got %s", str)
	}
}

func TestString_single(t *testing.T) {
	t.Parallel()

	s := New(42)
	str := s.String()

	if str != "{42}" {
		t.Fatalf("expected {42}, got %s", str)
	}
}
