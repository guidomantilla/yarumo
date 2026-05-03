package sets

import (
	"slices"
	"testing"
)

// --- Add ---

func TestAdd(t *testing.T) {
	t.Parallel()

	s := New(1, 2)
	s.Add(3, 4)

	if s.Len() != 4 {
		t.Fatalf("expected 4, got %d", s.Len())
	}
}

func TestAdd_duplicate(t *testing.T) {
	t.Parallel()

	s := New(1, 2)
	s.Add(2)

	if s.Len() != 2 {
		t.Fatalf("expected 2 (no duplicate), got %d", s.Len())
	}
}

// --- Remove ---

func TestRemove(t *testing.T) {
	t.Parallel()

	s := New(1, 2, 3)
	s.Remove(2)

	if s.Len() != 2 {
		t.Fatalf("expected 2, got %d", s.Len())
	}

	if s.Contains(2) {
		t.Fatal("expected 2 to be removed")
	}
}

func TestRemove_nonexistent(t *testing.T) {
	t.Parallel()

	s := New(1, 2)
	s.Remove(5)

	if s.Len() != 2 {
		t.Fatalf("expected 2, got %d", s.Len())
	}
}

// --- Contains ---

func TestContains_true(t *testing.T) {
	t.Parallel()

	s := New(1, 2, 3)

	if !s.Contains(2) {
		t.Fatal("expected true for existing element")
	}
}

func TestContains_false(t *testing.T) {
	t.Parallel()

	s := New(1, 2, 3)

	if s.Contains(5) {
		t.Fatal("expected false for missing element")
	}
}

// --- Len ---

func TestLen_empty(t *testing.T) {
	t.Parallel()

	s := New[int]()

	if s.Len() != 0 {
		t.Fatalf("expected 0, got %d", s.Len())
	}
}

// --- Items ---

func TestItems(t *testing.T) {
	t.Parallel()

	s := New(3, 1, 2)
	items := s.Items()

	slices.Sort(items)

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	if items[0] != 1 || items[1] != 2 || items[2] != 3 {
		t.Fatalf("unexpected items: %v", items)
	}
}

// --- IsEmpty ---

func TestIsEmpty_true(t *testing.T) {
	t.Parallel()

	s := New[int]()

	if !s.IsEmpty() {
		t.Fatal("expected empty")
	}
}

func TestIsEmpty_false(t *testing.T) {
	t.Parallel()

	s := New(1)

	if s.IsEmpty() {
		t.Fatal("expected non-empty")
	}
}

// --- Clone ---

func TestClone(t *testing.T) {
	t.Parallel()

	s := New(1, 2, 3)
	c := s.Clone()

	if !Equal(s, c) {
		t.Fatal("clone should equal original")
	}

	c.Add(4)

	if s.Contains(4) {
		t.Fatal("modifying clone should not affect original")
	}
}

// --- Union ---

func TestUnion(t *testing.T) {
	t.Parallel()

	a := New(1, 2, 3)
	b := New(3, 4, 5)
	u := Union(a, b)

	if u.Len() != 5 {
		t.Fatalf("expected 5, got %d", u.Len())
	}

	for _, v := range []int{1, 2, 3, 4, 5} {
		if !u.Contains(v) {
			t.Fatalf("expected %d in union", v)
		}
	}
}

func TestUnion_disjoint(t *testing.T) {
	t.Parallel()

	a := New(1, 2)
	b := New(3, 4)
	u := Union(a, b)

	if u.Len() != 4 {
		t.Fatalf("expected 4, got %d", u.Len())
	}
}

func TestUnion_empty(t *testing.T) {
	t.Parallel()

	a := New(1, 2)
	b := New[int]()
	u := Union(a, b)

	if !Equal(u, a) {
		t.Fatal("union with empty should equal original")
	}
}

// --- Intersection ---

func TestIntersection(t *testing.T) {
	t.Parallel()

	a := New(1, 2, 3)
	b := New(2, 3, 4)
	inter := Intersection(a, b)

	if inter.Len() != 2 {
		t.Fatalf("expected 2, got %d", inter.Len())
	}

	if !inter.Contains(2) || !inter.Contains(3) {
		t.Fatal("expected {2, 3}")
	}
}

func TestIntersection_disjoint(t *testing.T) {
	t.Parallel()

	a := New(1, 2)
	b := New(3, 4)
	inter := Intersection(a, b)

	if inter.Len() != 0 {
		t.Fatalf("expected empty, got %d", inter.Len())
	}
}

// --- Difference ---

func TestDifference(t *testing.T) {
	t.Parallel()

	a := New(1, 2, 3)
	b := New(2, 3, 4)
	diff := Difference(a, b)

	if diff.Len() != 1 {
		t.Fatalf("expected 1, got %d", diff.Len())
	}

	if !diff.Contains(1) {
		t.Fatal("expected {1}")
	}
}

func TestDifference_noOverlap(t *testing.T) {
	t.Parallel()

	a := New(1, 2)
	b := New(3, 4)
	diff := Difference(a, b)

	if !Equal(diff, a) {
		t.Fatal("difference with no overlap should equal original")
	}
}

// --- SymmetricDifference ---

func TestSymmetricDifference(t *testing.T) {
	t.Parallel()

	a := New(1, 2, 3)
	b := New(2, 3, 4)
	sym := SymmetricDifference(a, b)

	if sym.Len() != 2 {
		t.Fatalf("expected 2, got %d", sym.Len())
	}

	if !sym.Contains(1) || !sym.Contains(4) {
		t.Fatal("expected {1, 4}")
	}
}

func TestSymmetricDifference_equal(t *testing.T) {
	t.Parallel()

	a := New(1, 2)
	b := New(1, 2)
	sym := SymmetricDifference(a, b)

	if !sym.IsEmpty() {
		t.Fatal("symmetric difference of equal sets should be empty")
	}
}

// --- IsSubset ---

func TestIsSubset_true(t *testing.T) {
	t.Parallel()

	a := New(1, 2)
	b := New(1, 2, 3)

	if !IsSubset(a, b) {
		t.Fatal("expected {1,2} to be subset of {1,2,3}")
	}
}

func TestIsSubset_false(t *testing.T) {
	t.Parallel()

	a := New(1, 4)
	b := New(1, 2, 3)

	if IsSubset(a, b) {
		t.Fatal("expected {1,4} to not be subset of {1,2,3}")
	}
}

func TestIsSubset_empty(t *testing.T) {
	t.Parallel()

	a := New[int]()
	b := New(1, 2, 3)

	if !IsSubset(a, b) {
		t.Fatal("empty set should be subset of any set")
	}
}

// --- IsSuperset ---

func TestIsSuperset_true(t *testing.T) {
	t.Parallel()

	a := New(1, 2, 3)
	b := New(1, 2)

	if !IsSuperset(a, b) {
		t.Fatal("expected {1,2,3} to be superset of {1,2}")
	}
}

func TestIsSuperset_false(t *testing.T) {
	t.Parallel()

	a := New(1, 2)
	b := New(1, 2, 3)

	if IsSuperset(a, b) {
		t.Fatal("expected {1,2} to not be superset of {1,2,3}")
	}
}

// --- Equal ---

func TestEqual_true(t *testing.T) {
	t.Parallel()

	a := New(1, 2, 3)
	b := New(3, 2, 1)

	if !Equal(a, b) {
		t.Fatal("expected equal sets")
	}
}

func TestEqual_false(t *testing.T) {
	t.Parallel()

	a := New(1, 2)
	b := New(1, 3)

	if Equal(a, b) {
		t.Fatal("expected unequal sets")
	}
}

func TestEqual_differentSizes(t *testing.T) {
	t.Parallel()

	a := New(1, 2)
	b := New(1, 2, 3)

	if Equal(a, b) {
		t.Fatal("expected unequal sets")
	}
}

func TestEqual_bothEmpty(t *testing.T) {
	t.Parallel()

	a := New[int]()
	b := New[int]()

	if !Equal(a, b) {
		t.Fatal("expected empty sets to be equal")
	}
}
