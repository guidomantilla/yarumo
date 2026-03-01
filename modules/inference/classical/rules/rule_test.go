package rules

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"
)

func TestNewRule(t *testing.T) {
	t.Parallel()

	t.Run("creates rule with defaults", func(t *testing.T) {
		t.Parallel()

		r := NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})

		if r.Name() != "r1" {
			t.Fatalf("expected r1, got %s", r.Name())
		}

		if r.Priority() != 0 {
			t.Fatalf("expected priority 0, got %d", r.Priority())
		}
	})

	t.Run("creates rule with priority", func(t *testing.T) {
		t.Parallel()

		r := NewRule("r2", logic.Var("A"), map[logic.Var]bool{"B": true}, WithPriority(5))

		if r.Priority() != 5 {
			t.Fatalf("expected priority 5, got %d", r.Priority())
		}
	})
}

func TestRule_Condition(t *testing.T) {
	t.Parallel()

	t.Run("returns the formula", func(t *testing.T) {
		t.Parallel()

		cond := logic.AndF{L: logic.Var("A"), R: logic.Var("B")}
		r := NewRule("r1", cond, map[logic.Var]bool{"C": true})

		got := r.Condition().String()
		if got != "(A & B)" {
			t.Fatalf("expected (A & B), got %s", got)
		}
	})
}

func TestRule_Conclusion(t *testing.T) {
	t.Parallel()

	t.Run("returns copy of conclusion", func(t *testing.T) {
		t.Parallel()

		original := map[logic.Var]bool{"B": true}
		r := NewRule("r1", logic.Var("A"), original)
		c := r.Conclusion()

		c["B"] = false

		c2 := r.Conclusion()
		if !c2["B"] {
			t.Fatal("expected original conclusion unchanged")
		}
	})

	t.Run("constructor copies conclusion", func(t *testing.T) {
		t.Parallel()

		original := map[logic.Var]bool{"B": true}
		r := NewRule("r1", logic.Var("A"), original)

		original["B"] = false

		c := r.Conclusion()
		if !c["B"] {
			t.Fatal("expected stored conclusion unchanged after mutating original")
		}
	})
}

func TestRule_Fires(t *testing.T) {
	t.Parallel()

	t.Run("fires when condition is true", func(t *testing.T) {
		t.Parallel()

		r := NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})

		got := r.Fires(logic.Fact{"A": true})
		if !got {
			t.Fatal("expected rule to fire")
		}
	})

	t.Run("does not fire when condition is false", func(t *testing.T) {
		t.Parallel()

		r := NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})

		got := r.Fires(logic.Fact{"A": false})
		if got {
			t.Fatal("expected rule not to fire")
		}
	})

	t.Run("does not fire when variable missing", func(t *testing.T) {
		t.Parallel()

		r := NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})

		got := r.Fires(logic.Fact{})
		if got {
			t.Fatal("expected rule not to fire with empty facts")
		}
	})
}

func TestRule_Produces(t *testing.T) {
	t.Parallel()

	t.Run("produces when conclusion differs", func(t *testing.T) {
		t.Parallel()

		r := NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})

		got := r.Produces(logic.Fact{"A": true})
		if !got {
			t.Fatal("expected rule to produce new info")
		}
	})

	t.Run("does not produce when conclusion already known", func(t *testing.T) {
		t.Parallel()

		r := NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})

		got := r.Produces(logic.Fact{"A": true, "B": true})
		if got {
			t.Fatal("expected rule not to produce when conclusion matches")
		}
	})

	t.Run("does not produce when condition is false", func(t *testing.T) {
		t.Parallel()

		r := NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})

		got := r.Produces(logic.Fact{"A": false})
		if got {
			t.Fatal("expected rule not to produce when condition is false")
		}
	})

	t.Run("produces when value differs", func(t *testing.T) {
		t.Parallel()

		r := NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})

		got := r.Produces(logic.Fact{"A": true, "B": false})
		if !got {
			t.Fatal("expected rule to produce when value differs")
		}
	})
}
