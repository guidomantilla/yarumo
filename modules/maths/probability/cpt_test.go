package probability

import (
	"errors"
	"testing"
)

func TestNewCPT(t *testing.T) {
	t.Parallel()

	c := NewCPT("WetGrass", []Var{"Rain", "Sprinkler"})

	if c.Variable != "WetGrass" {
		t.Fatalf("expected WetGrass, got %s", string(c.Variable))
	}

	if len(c.Parents) != 2 {
		t.Fatalf("expected 2 parents, got %d", len(c.Parents))
	}

	if len(c.Entries) != 0 {
		t.Fatalf("expected empty entries, got %d", len(c.Entries))
	}
}

func TestNewCPT_noParents(t *testing.T) {
	t.Parallel()

	c := NewCPT("Rain", nil)

	if c.Variable != "Rain" {
		t.Fatalf("expected Rain, got %s", string(c.Variable))
	}

	if len(c.Parents) != 0 {
		t.Fatalf("expected 0 parents, got %d", len(c.Parents))
	}
}

func TestCPT_SetAndLookup(t *testing.T) {
	t.Parallel()

	c := NewCPT("WetGrass", []Var{"Rain"})
	dist := Distribution{"true": 0.9, "false": 0.1}

	c.Set(Assignment{"Rain": "true"}, dist)

	result, err := c.Lookup(Assignment{"Rain": "true"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["true"] != 0.9 {
		t.Fatalf("expected 0.9, got %f", float64(result["true"]))
	}
}

func TestCPT_Lookup_notFound(t *testing.T) {
	t.Parallel()

	c := NewCPT("WetGrass", []Var{"Rain"})

	_, err := c.Lookup(Assignment{"Rain": "true"})
	if !errors.Is(err, ErrOutcomeNotFound) {
		t.Fatalf("expected ErrOutcomeNotFound, got %v", err)
	}
}

func TestCPT_SetCopiesDistribution(t *testing.T) {
	t.Parallel()

	c := NewCPT("X", []Var{"Y"})
	dist := Distribution{"a": 0.5, "b": 0.5}

	c.Set(Assignment{"Y": "y1"}, dist)

	// Modify original - should not affect CPT.
	dist["a"] = 0.9

	result, err := c.Lookup(Assignment{"Y": "y1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["a"] != 0.5 {
		t.Fatalf("expected 0.5 (original value), got %f", float64(result["a"]))
	}
}

func TestCPT_LookupCopiesDistribution(t *testing.T) {
	t.Parallel()

	c := NewCPT("X", []Var{"Y"})
	c.Set(Assignment{"Y": "y1"}, Distribution{"a": 0.5, "b": 0.5})

	result, err := c.Lookup(Assignment{"Y": "y1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Modify result - should not affect CPT.
	result["a"] = 0.9

	result2, err := c.Lookup(Assignment{"Y": "y1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result2["a"] != 0.5 {
		t.Fatalf("expected 0.5 (original value), got %f", float64(result2["a"]))
	}
}

func TestCPT_Validate_valid(t *testing.T) {
	t.Parallel()

	c := NewCPT("X", []Var{"Y"})
	c.Set(Assignment{"Y": "y1"}, Distribution{"a": 0.6, "b": 0.4})
	c.Set(Assignment{"Y": "y2"}, Distribution{"a": 0.3, "b": 0.7})

	err := c.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCPT_Validate_empty(t *testing.T) {
	t.Parallel()

	c := NewCPT("X", nil)

	err := c.Validate()
	if !errors.Is(err, ErrEmptyDist) {
		t.Fatalf("expected ErrEmptyDist, got %v", err)
	}
}

func TestCPT_Validate_notNormalized(t *testing.T) {
	t.Parallel()

	c := NewCPT("X", []Var{"Y"})
	c.Set(Assignment{"Y": "y1"}, Distribution{"a": 0.3, "b": 0.3})

	err := c.Validate()
	if !errors.Is(err, ErrNotNormalized) {
		t.Fatalf("expected ErrNotNormalized, got %v", err)
	}
}

func TestNewCPT_copiesParents(t *testing.T) {
	t.Parallel()

	parents := []Var{"A", "B"}

	c := NewCPT("X", parents)

	// Modify original - should not affect CPT.
	parents[0] = "Z"

	if c.Parents[0] != "A" {
		t.Fatalf("expected A (original value), got %s", string(c.Parents[0]))
	}
}

func TestSerializeAssignment_empty(t *testing.T) {
	t.Parallel()

	result := serializeAssignment(Assignment{}, nil)

	if result != "" {
		t.Fatalf("expected empty string, got %q", result)
	}
}

func TestSerializeAssignment_ordered(t *testing.T) {
	t.Parallel()

	config := Assignment{"B": "b1", "A": "a1"}
	order := []Var{"A", "B"}

	result := serializeAssignment(config, order)

	if result != "A=a1,B=b1" {
		t.Fatalf("expected A=a1,B=b1, got %q", result)
	}
}

func TestSerializeAssignmentSorted(t *testing.T) {
	t.Parallel()

	config := Assignment{"C": "c1", "A": "a1", "B": "b1"}

	result := SerializeAssignmentSorted(config)

	if result != "A=a1,B=b1,C=c1" {
		t.Fatalf("expected A=a1,B=b1,C=c1, got %q", result)
	}
}

func TestSerializeAssignmentSorted_empty(t *testing.T) {
	t.Parallel()

	result := SerializeAssignmentSorted(Assignment{})

	if result != "" {
		t.Fatalf("expected empty string, got %q", result)
	}
}
