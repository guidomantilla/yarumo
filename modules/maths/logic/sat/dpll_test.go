package sat

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"
)

func TestSolve(t *testing.T) {
	t.Parallel()

	t.Run("empty cnf is satisfiable", func(t *testing.T) {
		t.Parallel()

		sat, assignment := Solve(CNF{})

		if !sat {
			t.Fatal("expected satisfiable")
		}

		if assignment == nil {
			t.Fatal("expected non-nil assignment")
		}
	})

	t.Run("single positive literal", func(t *testing.T) {
		t.Parallel()

		cnf := CNF{{Lit{V: "A"}}}

		sat, assignment := Solve(cnf)

		if !sat {
			t.Fatal("expected satisfiable")
		}

		if !assignment["A"] {
			t.Fatalf("expected A=true, got %v", assignment["A"])
		}
	})

	t.Run("single negative literal", func(t *testing.T) {
		t.Parallel()

		cnf := CNF{{Lit{V: "A", Neg: true}}}

		sat, assignment := Solve(cnf)

		if !sat {
			t.Fatal("expected satisfiable")
		}

		if assignment["A"] {
			t.Fatalf("expected A=false, got %v", assignment["A"])
		}
	})

	t.Run("empty clause is unsatisfiable", func(t *testing.T) {
		t.Parallel()

		cnf := CNF{{}}

		sat, assignment := Solve(cnf)

		if sat {
			t.Fatal("expected unsatisfiable")
		}

		if assignment != nil {
			t.Fatal("expected nil assignment")
		}
	})

	t.Run("contradiction A and not A", func(t *testing.T) {
		t.Parallel()

		cnf := CNF{
			{Lit{V: "A"}},
			{Lit{V: "A", Neg: true}},
		}

		sat, _ := Solve(cnf)

		if sat {
			t.Fatal("expected unsatisfiable")
		}
	})

	t.Run("satisfiable two clauses", func(t *testing.T) {
		t.Parallel()

		// (A | B) & (!A | C)
		cnf := CNF{
			{Lit{V: "A"}, Lit{V: "B"}},
			{Lit{V: "A", Neg: true}, Lit{V: "C"}},
		}

		sat, assignment := Solve(cnf)

		if !sat {
			t.Fatal("expected satisfiable")
		}

		// Verify the assignment satisfies the formula
		c1 := assignment["A"] || assignment["B"]
		c2 := !assignment["A"] || assignment["C"]

		if !c1 || !c2 {
			t.Fatalf("assignment does not satisfy CNF: %v", assignment)
		}
	})

	t.Run("pure literal elimination", func(t *testing.T) {
		t.Parallel()

		// A appears only positive, B appears both ways
		cnf := CNF{
			{Lit{V: "A"}, Lit{V: "B"}},
			{Lit{V: "A"}, Lit{V: "B", Neg: true}},
		}

		sat, assignment := Solve(cnf)

		if !sat {
			t.Fatal("expected satisfiable")
		}

		if !assignment["A"] {
			t.Fatalf("expected A=true (pure literal), got %v", assignment["A"])
		}
	})
}

func TestSolve_advanced(t *testing.T) {
	t.Parallel()

	t.Run("requires backtracking", func(t *testing.T) {
		t.Parallel()

		// (!A | B) & (A | C) & (!B | !C)
		cnf := CNF{
			{Lit{V: "A", Neg: true}, Lit{V: "B"}},
			{Lit{V: "A"}, Lit{V: "C"}},
			{Lit{V: "B", Neg: true}, Lit{V: "C", Neg: true}},
		}

		sat, assignment := Solve(cnf)

		if !sat {
			t.Fatal("expected satisfiable")
		}

		c1 := !assignment["A"] || assignment["B"]
		c2 := assignment["A"] || assignment["C"]
		c3 := !assignment["B"] || !assignment["C"]

		if !c1 || !c2 || !c3 {
			t.Fatalf("assignment does not satisfy CNF: %v", assignment)
		}
	})

	t.Run("negative pure literal", func(t *testing.T) {
		t.Parallel()

		// A appears only negative
		cnf := CNF{
			{Lit{V: "A", Neg: true}, Lit{V: "B"}},
			{Lit{V: "A", Neg: true}, Lit{V: "B", Neg: true}},
		}

		sat, assignment := Solve(cnf)

		if !sat {
			t.Fatal("expected satisfiable")
		}

		if assignment["A"] {
			t.Fatalf("expected A=false (negative pure literal), got %v", assignment["A"])
		}
	})

	t.Run("positive split fails negative succeeds", func(t *testing.T) {
		t.Parallel()

		// (A | B) & (A | C) & (!A | !B) & (!A | !C) & (B | C)
		cnf := CNF{
			{Lit{V: "A"}, Lit{V: "B"}},
			{Lit{V: "A"}, Lit{V: "C"}},
			{Lit{V: "A", Neg: true}, Lit{V: "B", Neg: true}},
			{Lit{V: "A", Neg: true}, Lit{V: "C", Neg: true}},
			{Lit{V: "B"}, Lit{V: "C"}},
		}

		sat, assignment := Solve(cnf)

		if !sat {
			t.Fatal("expected satisfiable")
		}

		c1 := assignment["A"] || assignment["B"]
		c2 := assignment["A"] || assignment["C"]
		c3 := !assignment["A"] || !assignment["B"]
		c4 := !assignment["A"] || !assignment["C"]
		c5 := assignment["B"] || assignment["C"]

		if !c1 || !c2 || !c3 || !c4 || !c5 {
			t.Fatalf("assignment does not satisfy CNF: %v", assignment)
		}
	})

	t.Run("both splits fail unsatisfiable", func(t *testing.T) {
		t.Parallel()

		// (A | B) & (A | !B) & (!A | B) & (!A | !B)
		cnf := CNF{
			{Lit{V: "A"}, Lit{V: "B"}},
			{Lit{V: "A"}, Lit{V: "B", Neg: true}},
			{Lit{V: "A", Neg: true}, Lit{V: "B"}},
			{Lit{V: "A", Neg: true}, Lit{V: "B", Neg: true}},
		}

		sat, _ := Solve(cnf)

		if sat {
			t.Fatal("expected unsatisfiable")
		}
	})

	t.Run("deep split with accumulated assignment", func(t *testing.T) {
		t.Parallel()

		// 3 variables, no unit or pure literals, requires nested splits
		cnf := CNF{
			{Lit{V: "A"}, Lit{V: "B"}, Lit{V: "C"}},
			{Lit{V: "A"}, Lit{V: "B", Neg: true}, Lit{V: "C", Neg: true}},
			{Lit{V: "A", Neg: true}, Lit{V: "B"}, Lit{V: "C", Neg: true}},
			{Lit{V: "A", Neg: true}, Lit{V: "B", Neg: true}, Lit{V: "C"}},
		}

		sat, assignment := Solve(cnf)

		if !sat {
			t.Fatal("expected satisfiable")
		}

		// Verify all clauses
		for i, clause := range cnf {
			satisfied := false

			for _, lit := range clause {
				val := assignment[lit.V]

				if lit.Neg {
					val = !val
				}

				if val {
					satisfied = true
				}
			}

			if !satisfied {
				t.Fatalf("clause %d not satisfied: %v", i, assignment)
			}
		}
	})
}

func TestSolve_crossValidation(t *testing.T) {
	t.Parallel()

	t.Run("satisfiable variable", func(t *testing.T) {
		t.Parallel()

		f := logic.Var("A")
		cnf := FromFormula(logic.ToCNF(f))

		sat, _ := Solve(cnf)

		if sat != logic.IsSatisfiable(f) {
			t.Fatal("SAT and brute force disagree")
		}
	})

	t.Run("contradiction", func(t *testing.T) {
		t.Parallel()

		f := logic.AndF{L: logic.Var("A"), R: logic.NotF{F: logic.Var("A")}}
		cnf := FromFormula(logic.ToCNF(f))

		sat, _ := Solve(cnf)

		if sat != logic.IsSatisfiable(f) {
			t.Fatal("SAT and brute force disagree")
		}
	})

	t.Run("tautology", func(t *testing.T) {
		t.Parallel()

		f := logic.OrF{L: logic.Var("A"), R: logic.NotF{F: logic.Var("A")}}
		cnf := FromFormula(logic.ToCNF(f))

		sat, _ := Solve(cnf)

		if sat != logic.IsSatisfiable(f) {
			t.Fatal("SAT and brute force disagree")
		}
	})

	t.Run("implication", func(t *testing.T) {
		t.Parallel()

		f := logic.ImplF{L: logic.Var("A"), R: logic.Var("B")}
		cnf := FromFormula(logic.ToCNF(f))

		sat, _ := Solve(cnf)

		if sat != logic.IsSatisfiable(f) {
			t.Fatal("SAT and brute force disagree")
		}
	})

	t.Run("biconditional", func(t *testing.T) {
		t.Parallel()

		f := logic.IffF{L: logic.Var("A"), R: logic.Var("B")}
		cnf := FromFormula(logic.ToCNF(f))

		sat, _ := Solve(cnf)

		if sat != logic.IsSatisfiable(f) {
			t.Fatal("SAT and brute force disagree")
		}
	})

	t.Run("three variable formula", func(t *testing.T) {
		t.Parallel()

		// (A & B) | C
		f := logic.OrF{
			L: logic.AndF{L: logic.Var("A"), R: logic.Var("B")},
			R: logic.Var("C"),
		}
		cnf := FromFormula(logic.ToCNF(f))

		sat, _ := Solve(cnf)

		if sat != logic.IsSatisfiable(f) {
			t.Fatal("SAT and brute force disagree")
		}
	})
}
