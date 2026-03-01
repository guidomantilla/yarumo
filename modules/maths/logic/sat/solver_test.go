package sat

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"
)

func TestSolver(t *testing.T) {
	t.Parallel()

	t.Run("satisfiable formula", func(t *testing.T) {
		t.Parallel()

		solver := Solver()

		sat, assignment := solver(logic.Var("A"))

		if !sat {
			t.Fatal("expected satisfiable")
		}

		if assignment == nil {
			t.Fatal("expected non-nil assignment")
		}
	})

	t.Run("unsatisfiable formula", func(t *testing.T) {
		t.Parallel()

		solver := Solver()

		f := logic.AndF{L: logic.Var("A"), R: logic.NotF{F: logic.Var("A")}}

		sat, _ := solver(f)

		if sat {
			t.Fatal("expected unsatisfiable")
		}
	})

	t.Run("assignment satisfies formula", func(t *testing.T) {
		t.Parallel()

		solver := Solver()

		f := logic.AndF{L: logic.Var("A"), R: logic.Var("B")}

		sat, assignment := solver(f)

		if !sat {
			t.Fatal("expected satisfiable")
		}

		if !f.Eval(assignment) {
			t.Fatalf("assignment does not satisfy formula: %v", assignment)
		}
	})

	t.Run("complex formula", func(t *testing.T) {
		t.Parallel()

		solver := Solver()

		// (A => B) & (B => C) & A
		f := logic.AndF{
			L: logic.AndF{
				L: logic.ImplF{L: logic.Var("A"), R: logic.Var("B")},
				R: logic.ImplF{L: logic.Var("B"), R: logic.Var("C")},
			},
			R: logic.Var("A"),
		}

		sat, assignment := solver(f)

		if !sat {
			t.Fatal("expected satisfiable")
		}

		if !f.Eval(assignment) {
			t.Fatalf("assignment does not satisfy formula: %v", assignment)
		}
	})
}
