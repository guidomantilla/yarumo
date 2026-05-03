package bayesian

import (
	"errors"
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/stats"
)

func TestNewCPT(t *testing.T) {
	t.Parallel()

	c := NewCPT("WetGrass", []stats.Var{"Rain", "Sprinkler"})

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

	c := NewCPT("WetGrass", []stats.Var{"Rain"})
	dist := stats.Distribution{"true": 0.9, "false": 0.1}

	c.Set(stats.Assignment{"Rain": "true"}, dist)

	result, err := c.Lookup(stats.Assignment{"Rain": "true"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["true"] != 0.9 {
		t.Fatalf("expected 0.9, got %f", float64(result["true"]))
	}
}

func TestCPT_Lookup_notFound(t *testing.T) {
	t.Parallel()

	c := NewCPT("WetGrass", []stats.Var{"Rain"})

	_, err := c.Lookup(stats.Assignment{"Rain": "true"})
	if !errors.Is(err, stats.ErrOutcomeNotFound) {
		t.Fatalf("expected ErrOutcomeNotFound, got %v", err)
	}
}

func TestCPT_SetCopiesDistribution(t *testing.T) {
	t.Parallel()

	c := NewCPT("X", []stats.Var{"Y"})
	dist := stats.Distribution{"a": 0.5, "b": 0.5}

	c.Set(stats.Assignment{"Y": "y1"}, dist)

	// Modify original - should not affect CPT.
	dist["a"] = 0.9

	result, err := c.Lookup(stats.Assignment{"Y": "y1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["a"] != 0.5 {
		t.Fatalf("expected 0.5 (original value), got %f", float64(result["a"]))
	}
}

func TestCPT_LookupCopiesDistribution(t *testing.T) {
	t.Parallel()

	c := NewCPT("X", []stats.Var{"Y"})
	c.Set(stats.Assignment{"Y": "y1"}, stats.Distribution{"a": 0.5, "b": 0.5})

	result, err := c.Lookup(stats.Assignment{"Y": "y1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Modify result - should not affect CPT.
	result["a"] = 0.9

	result2, err := c.Lookup(stats.Assignment{"Y": "y1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result2["a"] != 0.5 {
		t.Fatalf("expected 0.5 (original value), got %f", float64(result2["a"]))
	}
}

func TestCPT_Validate_valid(t *testing.T) {
	t.Parallel()

	c := NewCPT("X", []stats.Var{"Y"})
	c.Set(stats.Assignment{"Y": "y1"}, stats.Distribution{"a": 0.6, "b": 0.4})
	c.Set(stats.Assignment{"Y": "y2"}, stats.Distribution{"a": 0.3, "b": 0.7})

	err := c.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCPT_Validate_empty(t *testing.T) {
	t.Parallel()

	c := NewCPT("X", nil)

	err := c.Validate()
	if !errors.Is(err, stats.ErrEmptyDist) {
		t.Fatalf("expected ErrEmptyDist, got %v", err)
	}
}

func TestCPT_Validate_notNormalized(t *testing.T) {
	t.Parallel()

	c := NewCPT("X", []stats.Var{"Y"})
	c.Set(stats.Assignment{"Y": "y1"}, stats.Distribution{"a": 0.3, "b": 0.3})

	err := c.Validate()
	if !errors.Is(err, stats.ErrNotNormalized) {
		t.Fatalf("expected ErrNotNormalized, got %v", err)
	}
}

func TestNewCPT_copiesParents(t *testing.T) {
	t.Parallel()

	parents := []stats.Var{"A", "B"}

	c := NewCPT("X", parents)

	// Modify original - should not affect CPT.
	parents[0] = "Z"

	if c.Parents[0] != "A" {
		t.Fatalf("expected A (original value), got %s", string(c.Parents[0]))
	}
}

func TestSerializeAssignment_empty(t *testing.T) {
	t.Parallel()

	result := serializeAssignment(stats.Assignment{}, nil)

	if result != "" {
		t.Fatalf("expected empty string, got %q", result)
	}
}

func TestSerializeAssignment_ordered(t *testing.T) {
	t.Parallel()

	config := stats.Assignment{"B": "b1", "A": "a1"}
	order := []stats.Var{"A", "B"}

	result := serializeAssignment(config, order)

	if result != "A=a1,B=b1" {
		t.Fatalf("expected A=a1,B=b1, got %q", result)
	}
}

func TestSerializeAssignmentSorted(t *testing.T) {
	t.Parallel()

	config := stats.Assignment{"C": "c1", "A": "a1", "B": "b1"}

	result := SerializeAssignmentSorted(config)

	if result != "A=a1,B=b1,C=c1" {
		t.Fatalf("expected A=a1,B=b1,C=c1, got %q", result)
	}
}

func TestSerializeAssignmentSorted_empty(t *testing.T) {
	t.Parallel()

	result := SerializeAssignmentSorted(stats.Assignment{})

	if result != "" {
		t.Fatalf("expected empty string, got %q", result)
	}
}
