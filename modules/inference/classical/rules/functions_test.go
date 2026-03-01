package rules

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"
)

func TestSortByPriority(t *testing.T) {
	t.Parallel()

	t.Run("sorts by priority ascending", func(t *testing.T) {
		t.Parallel()

		r1 := NewRule("low", logic.Var("A"), map[logic.Var]bool{"B": true}, WithPriority(10))
		r2 := NewRule("high", logic.Var("A"), map[logic.Var]bool{"C": true}, WithPriority(1))
		r3 := NewRule("mid", logic.Var("A"), map[logic.Var]bool{"D": true}, WithPriority(5))

		sorted := SortByPriority([]Rule{r1, r2, r3})

		if sorted[0].Name() != "high" {
			t.Fatalf("expected high first, got %s", sorted[0].Name())
		}

		if sorted[1].Name() != "mid" {
			t.Fatalf("expected mid second, got %s", sorted[1].Name())
		}

		if sorted[2].Name() != "low" {
			t.Fatalf("expected low third, got %s", sorted[2].Name())
		}
	})

	t.Run("stable sort preserves insertion order", func(t *testing.T) {
		t.Parallel()

		r1 := NewRule("first", logic.Var("A"), map[logic.Var]bool{"B": true})
		r2 := NewRule("second", logic.Var("A"), map[logic.Var]bool{"C": true})

		sorted := SortByPriority([]Rule{r1, r2})

		if sorted[0].Name() != "first" {
			t.Fatalf("expected first, got %s", sorted[0].Name())
		}

		if sorted[1].Name() != "second" {
			t.Fatalf("expected second, got %s", sorted[1].Name())
		}
	})

	t.Run("does not modify original slice", func(t *testing.T) {
		t.Parallel()

		r1 := NewRule("low", logic.Var("A"), map[logic.Var]bool{"B": true}, WithPriority(10))
		r2 := NewRule("high", logic.Var("A"), map[logic.Var]bool{"C": true}, WithPriority(1))
		original := []Rule{r1, r2}

		SortByPriority(original)

		if original[0].Name() != "low" {
			t.Fatal("expected original slice unchanged")
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		sorted := SortByPriority([]Rule{})
		if len(sorted) != 0 {
			t.Fatalf("expected empty, got %d", len(sorted))
		}
	})
}

func TestVariables(t *testing.T) {
	t.Parallel()

	t.Run("returns all variables sorted and deduplicated", func(t *testing.T) {
		t.Parallel()

		r := NewRule("r1",
			logic.AndF{L: logic.Var("B"), R: logic.Var("A")},
			map[logic.Var]bool{"C": true, "A": false},
		)

		vars := Variables(r)

		if len(vars) != 3 {
			t.Fatalf("expected 3 variables, got %d", len(vars))
		}

		if vars[0] != "A" || vars[1] != "B" || vars[2] != "C" {
			t.Fatalf("expected [A B C], got %v", vars)
		}
	})

	t.Run("single variable in condition and conclusion", func(t *testing.T) {
		t.Parallel()

		r := NewRule("r1", logic.Var("A"), map[logic.Var]bool{"A": true})

		vars := Variables(r)

		if len(vars) != 1 {
			t.Fatalf("expected 1 variable, got %d", len(vars))
		}

		if vars[0] != "A" {
			t.Fatalf("expected A, got %s", vars[0])
		}
	})
}
