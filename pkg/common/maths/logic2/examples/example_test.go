package examples

import (
	"fmt"
	"testing"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/parser"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"
)

// TestExampleParseAndEval shows how to parse a formula and evaluate it against a set of facts.
func TestExampleParseAndEval(t *testing.T) {
	f := parser.MustParse("(A & B) => C")
	facts := props.Fact{
		props.Var("A"): true,
		props.Var("B"): true,
		props.Var("C"): false,
	}
	fmt.Println(f.String())
	fmt.Println(f.Eval(facts))
	// Output:
	// ((A & B) => C)
	// false
}

// TestExampleSimplify demonstrates the simplification of a formula using common logical laws.
func TestExampleSimplify(t *testing.T) {
	f := parser.MustParse("A | (A & B)")
	s := props.Simplify(f)
	fmt.Println("in:", f.String())
	fmt.Println("out:", s.String())
	// Output:
	// in: (A | (A & B))
	// out: A
}

// TestExampleIsSatisfiable shows basic satisfiability utilities (provisional: truth-table based in Phase 1).
func TestExampleIsSatisfiable(t *testing.T) {
	// Contradiction
	c := parser.MustParse("A & !A")
	fmt.Println(props.IsSatisfiable(c))
	fmt.Println(props.IsContradiction(c))
	fmt.Println(props.IsTautology(c))

	// Tautology
	f := parser.MustParse("A | !A")
	fmt.Println(props.IsSatisfiable(f))
	fmt.Println(props.IsContradiction(f))
	fmt.Println(props.IsTautology(f))
	// Output:
	// false
	// true
	// false
	// true
	// false
	// true
}

// TestExampleEquivalent shows logical equivalence via truth tables.
func TestExampleEquivalent(t *testing.T) {
	imp := parser.MustParse("A => B")
	cnf := parser.MustParse("!A | B")
	fmt.Println(props.Equivalent(imp, cnf))
	// Output:
	// true
}

// TestNNF_CNF_DNFTransformations demonstrates transforming a complex formula
// into NNF, CNF, and DNF, and then simplifying each.
func TestNNF_CNF_DNFTransformations(t *testing.T) {
	f := parser.MustParse("!(A & (B | !C)) <=> ((!A) | (!B & C))")
	fmt.Println("in:", f.String())

	nnf := props.ToNNF(f)
	cnf := props.ToCNF(f)
	dnf := props.ToDNF(f)

	fmt.Println("nnf:", nnf.String())
	fmt.Println("cnf:", cnf.String())
	fmt.Println("dnf:", dnf.String())

	fmt.Println("simplified(nnf):", props.Simplify(nnf).String())
	fmt.Println("simplified(cnf):", props.Simplify(cnf).String())
	fmt.Println("simplified(dnf):", props.Simplify(dnf).String())
}

// TestTruthTableAndFailCases builds the truth table and prints the number of
// failing assignments for a given implication.
func TestTruthTableAndFailCases(t *testing.T) {
	f := parser.MustParse("(A & B) => (C | D)")
	vars := f.Vars()
	fmt.Println("vars:", vars)

	tt := props.TruthTable(f)
	fmt.Println("rows:", len(tt))

	fails := props.FailCases(f)
	fmt.Println("fails:", len(fails))
	if len(fails) > 0 {
		// Show first 3 failing assignments (if any)
		limit := 3
		if len(fails) < limit {
			limit = len(fails)
		}
		for i := 0; i < limit; i++ {
			fmt.Println("fail:", fails[i])
		}
	}
}

// TestSimplifyAbsorptionAndComplement shows simplification rules such as
// absorption (A & (A | B)) => A and complements (B & !B) => ⊥.
func TestSimplifyAbsorptionAndComplement(t *testing.T) {
	f := parser.MustParse("(A & (A | B)) | (B & !B)")
	s := props.Simplify(f)
	fmt.Println("in:", f.String())
	fmt.Println("out:", s.String())
}

// TestComplexSatisfiability builds a slightly larger formula that is
// contradictory to exercise IsSatisfiable/IsContradiction utilities.
func TestComplexSatisfiability(t *testing.T) {
	// (A | B | C) & (!A | D) & (!B | D) & (!C | D) & !D is UNSAT
	f := parser.MustParse("(A | B | C) & (!A | D) & (!B | D) & (!C | D) & !D")
	fmt.Println("sat:", props.IsSatisfiable(f))
	fmt.Println("contradiction:", props.IsContradiction(f))
	fmt.Println("tautology:", props.IsTautology(f))
}

// TestEquivalenceLaws validates classic equivalences for implication and
// biconditional.
func TestEquivalenceLaws(t *testing.T) {
	imp := parser.MustParse("A => B")
	impCNF := parser.MustParse("!A | B")
	fmt.Println("imp≡cnf:", props.Equivalent(imp, impCNF))

	iff := parser.MustParse("A <=> B")
	iffDNF := parser.MustParse("(A & B) | (!A & !B)")
	fmt.Println("iff≡dnf:", props.Equivalent(iff, iffDNF))
}

// TestParserRoundTrip checks a simple parse → String → parse → String round-trip
// over the supported grammar subset.
func TestParserRoundTrip(t *testing.T) {
	in := "!A & (B | C) => D <=> (!D | (A & (B | C)))"
	f := parser.MustParse(in)
	p1 := f.String()
	f2 := parser.MustParse(p1)
	p2 := f2.String()
	fmt.Println("p1:", p1)
	fmt.Println("p2:", p2)
}
