package sat

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"
)

func TestFromFormula(t *testing.T) {
	t.Parallel()

	t.Run("single variable", func(t *testing.T) {
		t.Parallel()

		cnf := FromFormula(logic.Var("A"))

		if len(cnf) != 1 {
			t.Fatalf("expected 1 clause, got %d", len(cnf))
		}

		if len(cnf[0]) != 1 {
			t.Fatalf("expected 1 literal, got %d", len(cnf[0]))
		}

		if cnf[0][0].V != "A" || cnf[0][0].Neg {
			t.Fatalf("expected positive A, got %v", cnf[0][0])
		}
	})

	t.Run("negated variable", func(t *testing.T) {
		t.Parallel()

		cnf := FromFormula(logic.NotF{F: logic.Var("A")})

		if len(cnf) != 1 {
			t.Fatalf("expected 1 clause, got %d", len(cnf))
		}

		if len(cnf[0]) != 1 {
			t.Fatalf("expected 1 literal, got %d", len(cnf[0]))
		}

		if cnf[0][0].V != "A" || !cnf[0][0].Neg {
			t.Fatalf("expected negative A, got %v", cnf[0][0])
		}
	})

	t.Run("conjunction of variables", func(t *testing.T) {
		t.Parallel()

		f := logic.AndF{L: logic.Var("A"), R: logic.Var("B")}
		cnf := FromFormula(f)

		if len(cnf) != 2 {
			t.Fatalf("expected 2 clauses, got %d", len(cnf))
		}
	})

	t.Run("disjunction of variables", func(t *testing.T) {
		t.Parallel()

		f := logic.OrF{L: logic.Var("A"), R: logic.Var("B")}
		cnf := FromFormula(f)

		if len(cnf) != 1 {
			t.Fatalf("expected 1 clause, got %d", len(cnf))
		}

		if len(cnf[0]) != 2 {
			t.Fatalf("expected 2 literals, got %d", len(cnf[0]))
		}
	})

	t.Run("true formula", func(t *testing.T) {
		t.Parallel()

		cnf := FromFormula(logic.TrueF{})

		if len(cnf) != 0 {
			t.Fatalf("expected 0 clauses, got %d", len(cnf))
		}
	})

	t.Run("false formula", func(t *testing.T) {
		t.Parallel()

		cnf := FromFormula(logic.FalseF{})

		if len(cnf) != 1 {
			t.Fatalf("expected 1 clause, got %d", len(cnf))
		}

		if len(cnf[0]) != 0 {
			t.Fatalf("expected empty clause, got %d literals", len(cnf[0]))
		}
	})

	t.Run("cnf from ToCNF", func(t *testing.T) {
		t.Parallel()

		// (A | B) & (!A | C) — already CNF
		f := logic.AndF{
			L: logic.OrF{L: logic.Var("A"), R: logic.Var("B")},
			R: logic.OrF{L: logic.NotF{F: logic.Var("A")}, R: logic.Var("C")},
		}

		cnf := FromFormula(f)

		if len(cnf) != 2 {
			t.Fatalf("expected 2 clauses, got %d", len(cnf))
		}
	})

	t.Run("true in disjunction", func(t *testing.T) {
		t.Parallel()

		// A | true — tautological clause, should be skipped
		f := logic.OrF{L: logic.Var("A"), R: logic.TrueF{}}
		cnf := FromFormula(f)

		if len(cnf) != 0 {
			t.Fatalf("expected 0 clauses (tautology skipped), got %d", len(cnf))
		}
	})

	t.Run("false in disjunction", func(t *testing.T) {
		t.Parallel()

		// A | false — false contributes nothing
		f := logic.OrF{L: logic.Var("A"), R: logic.FalseF{}}
		cnf := FromFormula(f)

		if len(cnf) != 1 {
			t.Fatalf("expected 1 clause, got %d", len(cnf))
		}

		if len(cnf[0]) != 1 {
			t.Fatalf("expected 1 literal, got %d", len(cnf[0]))
		}
	})

	t.Run("true in left of disjunction", func(t *testing.T) {
		t.Parallel()

		f := logic.OrF{L: logic.TrueF{}, R: logic.Var("A")}
		cnf := FromFormula(f)

		if len(cnf) != 0 {
			t.Fatalf("expected 0 clauses (tautology skipped), got %d", len(cnf))
		}
	})
}
