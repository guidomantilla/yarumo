package examples

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"
	"github.com/guidomantilla/yarumo/maths/logic/entailment"
	"github.com/guidomantilla/yarumo/maths/logic/parser"
	"github.com/guidomantilla/yarumo/maths/logic/sat"
)

func TestExample_parseAndEval(t *testing.T) {
	t.Parallel()

	f, err := parser.Parse("(A & B) => C")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	facts := logic.Fact{"A": true, "B": true, "C": false}

	result := f.Eval(facts)

	if result {
		t.Fatal("expected false for A=true, B=true, C=false")
	}
}

func TestExample_transformAndSimplify(t *testing.T) {
	t.Parallel()

	f := parser.MustParse("!!A & (B | B)")

	simplified := logic.Simplify(f)
	cnf := logic.ToCNF(simplified)
	formatted := logic.Format(simplified)

	if formatted == "" {
		t.Fatal("expected non-empty formatted output")
	}

	if !logic.Equivalent(f, cnf) {
		t.Fatal("CNF not equivalent to original")
	}
}

func TestExample_satSolver(t *testing.T) {
	t.Parallel()

	f := parser.MustParse("(A => B) & (B => C) & A & !C")

	cnf := sat.FromFormula(logic.ToCNF(f))

	satisfiable, _ := sat.Solve(cnf)

	if satisfiable {
		t.Fatal("expected unsatisfiable")
	}
}

func TestExample_satSolverHook(t *testing.T) { //nolint:paralleltest // modifies global SAT solver state
	original := sat.Solver()

	logic.RegisterSATSolver(original)

	defer logic.RegisterSATSolver(nil)

	f := parser.MustParse("A & B")

	if !logic.IsSatisfiable(f) {
		t.Fatal("expected satisfiable via SAT solver hook")
	}

	if logic.IsTautology(f) {
		t.Fatal("expected non-tautology")
	}
}

func TestExample_entailment(t *testing.T) {
	t.Parallel()

	premises := []logic.Formula{
		parser.MustParse("A"),
		parser.MustParse("A => B"),
		parser.MustParse("B => C"),
	}
	conclusion := parser.MustParse("C")

	if !entailment.Entails(premises, conclusion) {
		t.Fatal("expected entailment to hold")
	}
}

func TestExample_counterModel(t *testing.T) {
	t.Parallel()

	premises := []logic.Formula{
		parser.MustParse("A => B"),
	}
	conclusion := parser.MustParse("B => A")

	entailed, counter := entailment.EntailsWithCounterModel(premises, conclusion)

	if entailed {
		t.Fatal("expected entailment to fail")
	}

	// In countermodel: A=>B is true but B=>A is false.
	premiseOk := premises[0].Eval(counter)
	conclusionOk := conclusion.Eval(counter)

	if !premiseOk {
		t.Fatal("countermodel should satisfy premises")
	}

	if conclusionOk {
		t.Fatal("countermodel should falsify conclusion")
	}
}

func TestExample_truthTable(t *testing.T) {
	t.Parallel()

	f := parser.MustParse("A & B")

	rows := logic.TruthTable(f)

	if len(rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(rows))
	}

	trueCount := 0

	for _, r := range rows {
		if r.Result {
			trueCount++
		}
	}

	if trueCount != 1 {
		t.Fatalf("expected 1 true row, got %d", trueCount)
	}
}

func TestExample_format(t *testing.T) {
	t.Parallel()

	f := parser.MustParse("A & !B => C <=> D")

	formatted := logic.Format(f)

	if formatted == "" {
		t.Fatal("expected non-empty formatted output")
	}
}

func TestExample_variables(t *testing.T) {
	t.Parallel()

	f := parser.MustParse("C & A | B & A")

	vars := f.Vars()

	if len(vars) != 3 {
		t.Fatalf("expected 3 variables, got %d", len(vars))
	}

	// Vars returns sorted, deduplicated list.
	if vars[0] != "A" || vars[1] != "B" || vars[2] != "C" {
		t.Fatalf("expected [A B C], got %v", vars)
	}
}

func TestExample_roundTrip(t *testing.T) {
	t.Parallel()

	original := "A & B => !C | D"

	f, err := parser.Parse(original)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	reparsed, err := parser.Parse(f.String())
	if err != nil {
		t.Fatalf("reparse error: %v", err)
	}

	if !logic.Equivalent(f, reparsed) {
		t.Fatal("round trip produced non-equivalent formula")
	}
}
