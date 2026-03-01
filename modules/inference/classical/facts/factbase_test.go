package facts

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"

	"github.com/guidomantilla/yarumo/inference/classical/explain"
)

func TestNewFactBase(t *testing.T) {
	t.Parallel()

	t.Run("creates empty fact base", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		if fb.Len() != 0 {
			t.Fatalf("expected 0, got %d", fb.Len())
		}
	})
}

func TestNewFactBaseFrom(t *testing.T) {
	t.Parallel()

	t.Run("pre-populates from initial facts", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBaseFrom(logic.Fact{"A": true, "B": false})

		if fb.Len() != 2 {
			t.Fatalf("expected 2, got %d", fb.Len())
		}

		val, known := fb.Get("A")
		if !known || !val {
			t.Fatal("expected A=true")
		}
	})

	t.Run("empty initial facts", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBaseFrom(logic.Fact{})
		if fb.Len() != 0 {
			t.Fatalf("expected 0, got %d", fb.Len())
		}
	})
}

func TestFactBase_Assert(t *testing.T) {
	t.Parallel()

	t.Run("asserts new fact", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		fb.Assert("A", true)

		val, known := fb.Get("A")
		if !known || !val {
			t.Fatal("expected A=true")
		}
	})

	t.Run("overwrites existing fact", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		fb.Assert("A", true)
		fb.Assert("A", false)

		val, known := fb.Get("A")
		if !known || val {
			t.Fatal("expected A=false")
		}
	})
}

func TestFactBase_AssertAll(t *testing.T) {
	t.Parallel()

	t.Run("asserts multiple facts", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		fb.AssertAll(logic.Fact{"A": true, "B": false})

		if fb.Len() != 2 {
			t.Fatalf("expected 2, got %d", fb.Len())
		}
	})
}

func TestFactBase_Derive(t *testing.T) {
	t.Parallel()

	t.Run("derives fact with rule info", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		fb.Derive("B", true, "rule1", 1)

		val, known := fb.Get("B")
		if !known || !val {
			t.Fatal("expected B=true")
		}

		prov, ok := fb.Provenance("B")
		if !ok {
			t.Fatal("expected provenance for B")
		}

		if prov.Origin != explain.Derived {
			t.Fatal("expected derived origin")
		}

		if prov.RuleName != "rule1" {
			t.Fatalf("expected rule1, got %s", prov.RuleName)
		}

		if prov.Step != 1 {
			t.Fatalf("expected step 1, got %d", prov.Step)
		}
	})
}

func TestFactBase_Retract(t *testing.T) {
	t.Parallel()

	t.Run("removes existing fact", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		fb.Assert("A", true)
		fb.Retract("A")

		_, known := fb.Get("A")
		if known {
			t.Fatal("expected A to be retracted")
		}
	})

	t.Run("retract nonexistent is no-op", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		fb.Retract("X")

		if fb.Len() != 0 {
			t.Fatalf("expected 0, got %d", fb.Len())
		}
	})
}

func TestFactBase_Get(t *testing.T) {
	t.Parallel()

	t.Run("known variable", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		fb.Assert("A", true)

		val, known := fb.Get("A")
		if !known {
			t.Fatal("expected known")
		}

		if !val {
			t.Fatal("expected true")
		}
	})

	t.Run("unknown variable", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()

		val, known := fb.Get("X")
		if known {
			t.Fatal("expected unknown")
		}

		if val {
			t.Fatal("expected false for unknown")
		}
	})
}

func TestFactBase_Snapshot(t *testing.T) {
	t.Parallel()

	t.Run("returns copy of facts", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		fb.Assert("A", true)
		fb.Assert("B", false)

		snap := fb.Snapshot()
		if len(snap) != 2 {
			t.Fatalf("expected 2, got %d", len(snap))
		}

		if !snap["A"] {
			t.Fatal("expected A=true in snapshot")
		}
	})

	t.Run("snapshot mutation does not affect fact base", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		fb.Assert("A", true)

		snap := fb.Snapshot()
		snap["A"] = false

		val, _ := fb.Get("A")
		if !val {
			t.Fatal("expected original fact base unchanged")
		}
	})
}

func TestFactBase_Provenance(t *testing.T) {
	t.Parallel()

	t.Run("asserted provenance", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		fb.Assert("A", true)

		prov, ok := fb.Provenance("A")
		if !ok {
			t.Fatal("expected provenance")
		}

		if prov.Origin != explain.Asserted {
			t.Fatal("expected asserted")
		}

		if prov.RuleName != "" {
			t.Fatalf("expected empty rule name, got %s", prov.RuleName)
		}
	})

	t.Run("unknown variable", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()

		_, ok := fb.Provenance("X")
		if ok {
			t.Fatal("expected not found")
		}
	})
}

func TestFactBase_AllProvenance(t *testing.T) {
	t.Parallel()

	t.Run("returns sorted provenance", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		fb.Assert("B", false)
		fb.Derive("A", true, "r1", 1)

		provs := fb.AllProvenance()
		if len(provs) != 2 {
			t.Fatalf("expected 2, got %d", len(provs))
		}

		if provs[0].Variable != "A" {
			t.Fatalf("expected A first, got %s", provs[0].Variable)
		}

		if provs[1].Variable != "B" {
			t.Fatalf("expected B second, got %s", provs[1].Variable)
		}
	})

	t.Run("empty fact base", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()

		provs := fb.AllProvenance()
		if len(provs) != 0 {
			t.Fatalf("expected 0, got %d", len(provs))
		}
	})
}

func TestFactBase_Len(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		if fb.Len() != 0 {
			t.Fatalf("expected 0, got %d", fb.Len())
		}
	})

	t.Run("after assertions", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		fb.Assert("A", true)
		fb.Assert("B", false)

		if fb.Len() != 2 {
			t.Fatalf("expected 2, got %d", fb.Len())
		}
	})
}

func TestFactBase_Clone(t *testing.T) {
	t.Parallel()

	t.Run("creates independent copy", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		fb.Assert("A", true)

		clone := fb.Clone()
		clone.Assert("B", false)

		if fb.Len() != 1 {
			t.Fatalf("expected original unchanged, got %d", fb.Len())
		}

		if clone.Len() != 2 {
			t.Fatalf("expected clone has 2, got %d", clone.Len())
		}
	})

	t.Run("clone preserves provenance", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		fb.Derive("A", true, "r1", 1)

		clone := fb.Clone()

		prov, ok := clone.Provenance("A")
		if !ok {
			t.Fatal("expected provenance in clone")
		}

		if prov.RuleName != "r1" {
			t.Fatalf("expected r1, got %s", prov.RuleName)
		}
	})
}
