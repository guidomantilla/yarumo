package evidence

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/probability"
)

func TestNewEvidenceBase(t *testing.T) {
	t.Parallel()

	eb := NewEvidenceBase()

	if eb.Len() != 0 {
		t.Fatalf("expected empty, got %d", eb.Len())
	}
}

func TestNewEvidenceBaseFrom(t *testing.T) {
	t.Parallel()

	eb := NewEvidenceBaseFrom(probability.Assignment{"X": "true", "Y": "false"})

	if eb.Len() != 2 {
		t.Fatalf("expected 2, got %d", eb.Len())
	}

	val, ok := eb.Get("X")
	if !ok {
		t.Fatal("expected X")
	}

	if val != "true" {
		t.Fatalf("expected true, got %s", string(val))
	}
}

func TestEvidenceBase_Observe(t *testing.T) {
	t.Parallel()

	eb := NewEvidenceBase()
	eb.Observe("Rain", "true")

	val, ok := eb.Get("Rain")
	if !ok {
		t.Fatal("expected Rain")
	}

	if val != "true" {
		t.Fatalf("expected true, got %s", string(val))
	}
}

func TestEvidenceBase_Retract(t *testing.T) {
	t.Parallel()

	eb := NewEvidenceBase()
	eb.Observe("Rain", "true")
	eb.Retract("Rain")

	_, ok := eb.Get("Rain")
	if ok {
		t.Fatal("expected not found after retract")
	}
}

func TestEvidenceBase_Get_notFound(t *testing.T) {
	t.Parallel()

	eb := NewEvidenceBase()

	_, ok := eb.Get("Unknown")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestEvidenceBase_Observed(t *testing.T) {
	t.Parallel()

	eb := NewEvidenceBase()
	eb.Observe("A", "a1")
	eb.Observe("B", "b1")

	observed := eb.Observed()
	if len(observed) != 2 {
		t.Fatalf("expected 2, got %d", len(observed))
	}

	if observed["A"] != "a1" {
		t.Fatalf("expected a1, got %s", string(observed["A"]))
	}
}

func TestEvidenceBase_Clone(t *testing.T) {
	t.Parallel()

	eb := NewEvidenceBase()
	eb.Observe("X", "x1")

	cloned := eb.Clone()
	cloned.Observe("Y", "y1")

	if eb.Len() != 1 {
		t.Fatalf("expected original unchanged, got %d", eb.Len())
	}

	if cloned.Len() != 2 {
		t.Fatalf("expected clone to have 2, got %d", cloned.Len())
	}
}

func TestEvidenceBase_Len(t *testing.T) {
	t.Parallel()

	eb := NewEvidenceBase()

	if eb.Len() != 0 {
		t.Fatalf("expected 0, got %d", eb.Len())
	}

	eb.Observe("A", "a1")

	if eb.Len() != 1 {
		t.Fatalf("expected 1, got %d", eb.Len())
	}
}

func TestEvidenceBase_ObserveOverwrite(t *testing.T) {
	t.Parallel()

	eb := NewEvidenceBase()
	eb.Observe("X", "old")
	eb.Observe("X", "new")

	val, ok := eb.Get("X")
	if !ok {
		t.Fatal("expected X")
	}

	if val != "new" {
		t.Fatalf("expected new, got %s", string(val))
	}
}
