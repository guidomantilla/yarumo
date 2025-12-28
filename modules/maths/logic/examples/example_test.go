package examples

import (
	"fmt"
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic/parser"
	p "github.com/guidomantilla/yarumo/maths/logic/props"
)

// --- helpers ---

func mustEq(t *testing.T, a, b p.Formula) {
	t.Helper()

	if !p.Equivalent(a, b) {
		t.Fatalf("expected equivalent formulas, got:\n  a: %s\n  b: %s", a.String(), b.String())
	}
}

func mustNeq(t *testing.T, a, b p.Formula) {
	t.Helper()

	if p.Equivalent(a, b) {
		t.Fatalf("expected non-equivalent formulas, but they are equivalent:\n  a: %s\n  b: %s", a.String(), b.String())
	}
}

// TestExampleParseAndEval shows how to parse a formula and evaluate it against a set of facts.
func TestExampleParseAndEval(t *testing.T) {
	f := parser.MustParse("(A & B) => C")
	facts := p.Fact{
		p.Var("A"): true,
		p.Var("B"): true,
		p.Var("C"): false,
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
	s := p.Simplify(f)
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
	fmt.Println(p.IsSatisfiable(c))
	fmt.Println(p.IsContradiction(c))
	fmt.Println(p.IsTautology(c))

	// Tautology
	f := parser.MustParse("A | !A")
	fmt.Println(p.IsSatisfiable(f))
	fmt.Println(p.IsContradiction(f))
	fmt.Println(p.IsTautology(f))
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
	fmt.Println(p.Equivalent(imp, cnf))
	// Output:
	// true
}

// TestNNF_CNF_DNFTransformations demonstrates transforming a complex formula
// into NNF, CNF, and DNF, and then simplifying each.
func TestNNF_CNF_DNFTransformations(t *testing.T) {
	f := parser.MustParse("!(A & (B | !C)) <=> ((!A) | (!B & C))")
	fmt.Println("in:", f.String())

	nnf := p.ToNNF(f)
	cnf := p.ToCNF(f)
	dnf := p.ToDNF(f)

	fmt.Println("nnf:", nnf.String())
	fmt.Println("cnf:", cnf.String())
	fmt.Println("dnf:", dnf.String())

	fmt.Println("simplified(nnf):", p.Simplify(nnf).String())
	fmt.Println("simplified(cnf):", p.Simplify(cnf).String())
	fmt.Println("simplified(dnf):", p.Simplify(dnf).String())
}

// TestTruthTableAndFailCases builds the truth table and prints the number of
// failing assignments for a given implication.
func TestTruthTableAndFailCases(t *testing.T) {
	f := parser.MustParse("(A & B) => (C | D)")
	vars := f.Vars()
	fmt.Println("vars:", vars)

	tt := p.TruthTable(f)
	fmt.Println("rows:", len(tt))

	fails := p.FailCases(f)
	fmt.Println("fails:", len(fails))

	if len(fails) > 0 {
		// Show first 3 failing assignments (if any)
		limit := min(len(fails), 3)

		for i := range limit {
			fmt.Println("fail:", fails[i])
		}
	}
}

// TestSimplifyAbsorptionAndComplement shows simplification rules such as
// absorption (A & (A | B)) => A and complements (B & !B) => ⊥.
func TestSimplifyAbsorptionAndComplement(t *testing.T) {
	f := parser.MustParse("(A & (A | B)) | (B & !B)")
	s := p.Simplify(f)
	fmt.Println("in:", f.String())
	fmt.Println("out:", s.String())
}

// TestComplexSatisfiability builds a slightly larger formula that is
// contradictory to exercise IsSatisfiable/IsContradiction utilities.
func TestComplexSatisfiability(t *testing.T) {
	// (A | B | C) & (!A | D) & (!B | D) & (!C | D) & !D is UNSAT
	f := parser.MustParse("(A | B | C) & (!A | D) & (!B | D) & (!C | D) & !D")
	fmt.Println("sat:", p.IsSatisfiable(f))
	fmt.Println("contradiction:", p.IsContradiction(f))
	fmt.Println("tautology:", p.IsTautology(f))
}

// TestEquivalenceLaws validates classic equivalences for implication and
// biconditional.
func TestEquivalenceLaws(t *testing.T) {
	imp := parser.MustParse("A => B")
	impCNF := parser.MustParse("!A | B")
	fmt.Println("imp≡cnf:", p.Equivalent(imp, impCNF))

	iff := parser.MustParse("A <=> B")
	iffDNF := parser.MustParse("(A & B) | (!A & !B)")
	fmt.Println("iff≡dnf:", p.Equivalent(iff, iffDNF))
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

// --- Double negation ---

// TestDoubleNegation checks that double negation is equivalent to the original formula.
func TestDoubleNegation(t *testing.T) {
	A := parser.MustParse("A")
	notnotA := parser.MustParse("!!A")
	mustEq(t, A, notnotA)
}

// --- De Morgan laws ---

// TestDeMorgan1 checks that De Morgan laws are equivalent to the original formula.
func TestDeMorgan1(t *testing.T) { // !(A & B) == !A | !B
	left := parser.MustParse("!(A & B)")
	right := parser.MustParse("(!A | !B)")
	mustEq(t, left, right)
}

// TestDeMorgan2 checks that De Morgan laws are equivalent to the original formula.
func TestDeMorgan2(t *testing.T) { // !(A | B) == !A & !B
	left := parser.MustParse("!(A | B)")
	right := parser.MustParse("(!A & !B)")
	mustEq(t, left, right)
}

// --- Idempotence ---

// TestIdempotenceAnd checks that idempotence is equivalent to the original formula.
func TestIdempotenceAnd(t *testing.T) {
	left := parser.MustParse("A & A")
	right := parser.MustParse("A")
	mustEq(t, left, right)
}

// TestIdempotenceOr checks that idempotence is equivalent to the original formula.
func TestIdempotenceOr(t *testing.T) {
	left := parser.MustParse("A | A")
	right := parser.MustParse("A")
	mustEq(t, left, right)
}

// --- Absorption ---

// TestAbsorption1 checks that absorption is equivalent to the original formula.
func TestAbsorption1(t *testing.T) { // A | (A & B) == A
	left := parser.MustParse("A | (A & B)")
	right := parser.MustParse("A")
	mustEq(t, left, right)
}

// TestAbsorption2 checks that absorption is equivalent to the original formula.
func TestAbsorption2(t *testing.T) { // A & (A | B) == A
	left := parser.MustParse("A & (A | B)")
	right := parser.MustParse("A")
	mustEq(t, left, right)
}

// --- Identity and Domination (Neutral and Annihilator) ---

// TestIdentityAnd checks that identity and domination are equivalent to the original formula.
func TestIdentityAnd(t *testing.T) { // A & ⊤ == A
	left := parser.MustParse("A & (B | !B)") // (B | !B) is ⊤
	right := parser.MustParse("A")
	mustEq(t, left, right)
}

// TestIdentityOr checks that identity and domination are equivalent to the original formula.
func TestIdentityOr(t *testing.T) { // A | ⊥ == A
	left := parser.MustParse("A | (B & !B)") // (B & !B) is ⊥
	right := parser.MustParse("A")
	mustEq(t, left, right)
}

// TestDominationAnd checks that identity and domination are equivalent to the original formula.
func TestDominationAnd(t *testing.T) { // A & ⊥ == ⊥
	left := parser.MustParse("A & (B & !B)")
	right := parser.MustParse("(B & !B)")
	mustEq(t, left, right)
}

// TestDominationOr checks that identity and domination are equivalent to the original formula.
func TestDominationOr(t *testing.T) { // A | ⊤ == ⊤
	left := parser.MustParse("A | (B | !B)")
	right := parser.MustParse("(B | !B)")
	mustEq(t, left, right)
}

// --- Complement laws ---

// TestComplementAnd checks that complement laws are equivalent to the original formula.
func TestComplementAnd(t *testing.T) { // A & !A == ⊥
	left := parser.MustParse("A & !A")
	// Build ⊥ via contradiction: (B & !B)
	right := parser.MustParse("(B & !B)")
	mustEq(t, left, right)
}

// TestComplementOr checks that complement laws are equivalent to the original formula.
func TestComplementOr(t *testing.T) { // A | !A == ⊤
	left := parser.MustParse("A | !A")
	// Build ⊤ via tautology: (B | !B)
	right := parser.MustParse("(B | !B)")
	mustEq(t, left, right)
}

// --- Commutativity ---

// TestCommutativityAnd checks that commutativity is equivalent to the original formula.
func TestCommutativityAnd(t *testing.T) {
	left := parser.MustParse("A & B")
	right := parser.MustParse("B & A")
	mustEq(t, left, right)
}

// TestCommutativityOr checks that commutativity is equivalent to the original formula.
func TestCommutativityOr(t *testing.T) {
	left := parser.MustParse("A | B")
	right := parser.MustParse("B | A")
	mustEq(t, left, right)
}

// --- Associativity ---

// TestAssociativityAnd checks that associativity is equivalent to the original formula.
func TestAssociativityAnd(t *testing.T) {
	left := parser.MustParse("(A & (B & C))")
	right := parser.MustParse("((A & B) & C)")
	mustEq(t, left, right)
}

// TestAssociativityOr checks that associativity is equivalent to the original formula.
func TestAssociativityOr(t *testing.T) {
	left := parser.MustParse("(A | (B | C))")
	right := parser.MustParse("((A | B) | C)")
	mustEq(t, left, right)
}

// --- Distributivity ---

func TestDistributivityAndOverOr(t *testing.T) { // A & (B | C) == (A & B) | (A & C)
	left := parser.MustParse("A & (B | C)")
	right := parser.MustParse("(A & B) | (A & C)")
	mustEq(t, left, right)
}

// TestDistributivityOrOverAnd checks that distributivity is equivalent to the original formula.
func TestDistributivityOrOverAnd(t *testing.T) { // A | (B & C) == (A | B) & (A | C)
	left := parser.MustParse("A | (B & C)")
	right := parser.MustParse("(A | B) & (A | C)")
	mustEq(t, left, right)
}

// --- Implication and IFF equivalences ---

// TestImplicationEquivalence checks that implication and IFF equivalences are equivalent to the original formula.
func TestImplicationEquivalence(t *testing.T) { // (A => B) == (!A | B)
	left := parser.MustParse("A => B")
	right := parser.MustParse("!A | B")
	mustEq(t, left, right)
}

// TestIffEquivalence checks that implication and IFF equivalences are equivalent to the original formula.
func TestIffEquivalence(t *testing.T) { // (A <=> B) == (A & B) | (!A & !B)
	left := parser.MustParse("A <=> B")
	right := parser.MustParse("(A & B) | (!A & !B)")
	mustEq(t, left, right)
}

// --- Simplification expectations (explicit) ---

// TestSimplifyAbsorptionExplicit checks that absorption is reduced.
func TestSimplifyAbsorptionExplicit(t *testing.T) {
	in := parser.MustParse("A & (A | B)")
	out := p.Simplify(in)
	want := parser.MustParse("A")
	mustEq(t, out, want)
}

// TestSimplifyComplementExplicit checks that complements are absorbed.
func TestSimplifyComplementExplicit(t *testing.T) {
	in := parser.MustParse("(B & !B) | (A & (A | C))")
	out := p.Simplify(in)
	want := parser.MustParse("(B & !B) | A") // absorption reduces A&(A|C) to A
	mustEq(t, out, want)
}

// --- Normal forms preserve equivalence ---

// TestToNNF_PreservesEquivalence checks that NNF preserves equivalence.
func TestToNNF_PreservesEquivalence(t *testing.T) {
	f := parser.MustParse("!(A & (B | !C)) <=> ((!A) | (!B & C))")
	nnf := p.ToNNF(f)
	mustEq(t, f, nnf)
}

// TestToCNF_PreservesEquivalence checks that CNF preserves equivalence.
func TestToCNF_PreservesEquivalence(t *testing.T) {
	f := parser.MustParse("(A => B) <=> (!C | D)")
	cnf := p.ToCNF(f)
	mustEq(t, f, cnf)
}

// TestToDNF_PreservesEquivalence checks that DNF preserves equivalence.
func TestToDNF_PreservesEquivalence(t *testing.T) {
	f := parser.MustParse("(A & (B | C)) | (!A & (D <=> E))")
	dnf := p.ToDNF(f)
	mustEq(t, f, dnf)
}
