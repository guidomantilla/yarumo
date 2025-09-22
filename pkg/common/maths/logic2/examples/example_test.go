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
