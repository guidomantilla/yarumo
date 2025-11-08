package examples

import (
	"testing"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/parser"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/props"
	sat "github.com/guidomantilla/yarumo/pkg/common/maths/logic/sat"
)

func runSAT(f props.Formula) bool {
	cnf, err := sat.FromFormulaToCNF(f)
	if err != nil {
		return false
	}
	ok, _ := sat.DPLL(cnf, nil)
	return ok
}

// TestSATvsTruthTable verifies SAT matches truth-table results for small formulas (≤ 8 vars).
func TestSATvsTruthTable(t *testing.T) {
	cases := []string{
		"A",            // simple var
		"A | !A",       // tautology
		"A & !A",       // contradiction
		"(A & B) => C", // implication
		"(A | B) & (!A | C)",
		"(A <=> B) | (!C & D)",
		"(A & (B | C)) | (!A & (D <=> E))",
	}
	for _, s := range cases {
		f := parser.MustParse(s)
		got := props.IsSatisfiable(f)
		want := runSAT(f)
		if got != want {
			t.Fatalf("IsSatisfiable mismatch for %q: got %v, want %v", s, got, want)
		}
	}
}

// TestPolicyThreshold ensures that large-var formulas still evaluate (using SAT when registered).
func TestPolicyThreshold(t *testing.T) {
	// Build a formula with > SATThreshold variables, e.g., (A1 | A2 | ... | A13)
	// This is satisfiable unless all are false.
	var f props.Formula = props.Var("A1")
	for i := 2; i <= props.SATThreshold+1; i++ {
		g := parser.MustParse("A" + string(rune('0'+i/10)) + string(rune('0'+i%10))) // simple two-char index
		// To avoid complicating indices with >9, for i=13 this yields A1 and A3; still distinct enough for this test context.
		f = props.OrF{L: f, R: g}
	}
	if !props.IsSatisfiable(f) {
		t.Fatalf("expected large disjunction to be satisfiable")
	}
}
