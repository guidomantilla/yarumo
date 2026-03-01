package main

import (
	"fmt"

	"github.com/guidomantilla/yarumo/maths/logic"
	"github.com/guidomantilla/yarumo/maths/logic/entailment"
	"github.com/guidomantilla/yarumo/maths/logic/parser"
	"github.com/guidomantilla/yarumo/maths/logic/sat"
)

func main() {
	buildFormulas()
	parseFormulas()
	evaluateFormulas()
	transformFormulas()
	simplifyFormulas()
	analyzeFormulas()
	checkSatisfiability()
	checkEntailment()
}

// buildFormulas shows how to construct formulas manually using node structs.
func buildFormulas() {
	fmt.Println("=== Building formulas ===")

	// Variables
	a := logic.Var("A")
	b := logic.Var("B")

	// NOT A
	notA := logic.NotF{F: a}
	fmt.Println("NOT A:", notA.String())

	// A AND B
	aAndB := logic.AndF{L: a, R: b}
	fmt.Println("A AND B:", aAndB.String())

	// A OR B
	aOrB := logic.OrF{L: a, R: b}
	fmt.Println("A OR B:", aOrB.String())

	// A => B (if A then B)
	aImplB := logic.ImplF{L: a, R: b}
	fmt.Println("A => B:", aImplB.String())

	// A <=> B (A if and only if B)
	aIffB := logic.IffF{L: a, R: b}
	fmt.Println("A <=> B:", aIffB.String())

	// Complex: (A AND B) => (NOT A OR B)
	complex := logic.ImplF{
		L: logic.AndF{L: a, R: b},
		R: logic.OrF{L: notA, R: b},
	}
	fmt.Println("(A AND B) => (NOT A OR B):", complex.String())

	// Unicode rendering
	fmt.Println("Unicode:", logic.Format(complex))
	fmt.Println()
}

// parseFormulas shows how to build formulas from text strings.
func parseFormulas() {
	fmt.Println("=== Parsing formulas ===")

	// Programmer syntax
	f1, _ := parser.Parse("A & B")
	fmt.Println("A & B:", f1.String())

	// Unicode syntax
	f2, _ := parser.Parse("A ∧ B → C")
	fmt.Println("A ∧ B → C:", f2.String())

	// Keywords
	f3, _ := parser.Parse("A and B or not C")
	fmt.Println("A and B or not C:", f3.String())

	// Complex with parentheses
	f4, _ := parser.Parse("(P | Q) & (R => S) <=> W")
	fmt.Println("(P | Q) & (R => S) <=> W:", f4.String())

	// Error handling
	_, err := parser.Parse("")
	fmt.Println("Empty input error:", err)

	_, err = parser.Parse("A &")
	fmt.Println("Incomplete formula error:", err)
	fmt.Println()
}

// evaluateFormulas shows how to evaluate formulas against truth assignments.
func evaluateFormulas() {
	fmt.Println("=== Evaluating formulas ===")

	f, _ := parser.Parse("A & B => C")

	// A=true, B=true, C=true => true (premise satisfied, conclusion satisfied)
	facts1 := logic.Fact{"A": true, "B": true, "C": true}
	fmt.Println("A=T, B=T, C=T:", f.Eval(facts1))

	// A=true, B=true, C=false => false (premise satisfied, conclusion NOT satisfied)
	facts2 := logic.Fact{"A": true, "B": true, "C": false}
	fmt.Println("A=T, B=T, C=F:", f.Eval(facts2))

	// A=true, B=false, C=false => true (premise NOT satisfied, implication is true)
	facts3 := logic.Fact{"A": true, "B": false, "C": false}
	fmt.Println("A=T, B=F, C=F:", f.Eval(facts3))

	// List all variables in the formula
	fmt.Println("Variables:", f.Vars())
	fmt.Println()
}

// transformFormulas shows normal form conversions.
func transformFormulas() {
	fmt.Println("=== Transformations ===")

	f, _ := parser.Parse("A => (B | C)")

	fmt.Println("Original:  ", logic.Format(f))
	fmt.Println("NNF:       ", logic.Format(logic.ToNNF(f)))
	fmt.Println("CNF:       ", logic.Format(logic.ToCNF(f)))
	fmt.Println("DNF:       ", logic.Format(logic.ToDNF(f)))
	fmt.Println()
}

// simplifyFormulas shows algebraic simplification.
func simplifyFormulas() {
	fmt.Println("=== Simplification ===")

	// A AND true => A
	f1, _ := parser.Parse("A & true")
	fmt.Println(logic.Format(f1), "=>", logic.Format(logic.Simplify(f1)))

	// A OR false => A
	f2, _ := parser.Parse("A | false")
	fmt.Println(logic.Format(f2), "=>", logic.Format(logic.Simplify(f2)))

	// NOT NOT A => A
	f3, _ := parser.Parse("!!A")
	fmt.Println(logic.Format(f3), "=>", logic.Format(logic.Simplify(f3)))

	// A AND NOT A => false
	f4, _ := parser.Parse("A & !A")
	fmt.Println(logic.Format(f4), "=>", logic.Format(logic.Simplify(f4)))

	// A OR NOT A => true
	f5, _ := parser.Parse("A | !A")
	fmt.Println(logic.Format(f5), "=>", logic.Format(logic.Simplify(f5)))
	fmt.Println()
}

// analyzeFormulas shows truth tables, equivalence, and fail cases.
func analyzeFormulas() {
	fmt.Println("=== Analysis ===")

	f, _ := parser.Parse("A & B")

	// Truth table
	fmt.Println("Truth table for A & B:")
	for _, row := range logic.TruthTable(f) {
		fmt.Printf("  A=%v B=%v => %v\n", row.Assignment["A"], row.Assignment["B"], row.Result)
	}

	// Equivalence: De Morgan's law
	f1, _ := parser.Parse("!(A & B)")
	f2, _ := parser.Parse("!A | !B")
	fmt.Println("!(A & B) == !A | !B:", logic.Equivalent(f1, f2))

	// Fail cases: when does the formula evaluate to false?
	f3, _ := parser.Parse("A => B")
	fails := logic.FailCases(f3)
	fmt.Println("A => B is false when:")
	for _, fail := range fails {
		fmt.Printf("  A=%v B=%v\n", fail["A"], fail["B"])
	}
	fmt.Println()
}

// checkSatisfiability shows SAT solving and tautology/contradiction checks.
func checkSatisfiability() {
	fmt.Println("=== Satisfiability ===")

	// Register the DPLL solver (faster than brute-force)
	logic.RegisterSATSolver(func(f logic.Formula) (bool, logic.Fact) {
		cnf := sat.FromFormula(f)
		return sat.Solve(cnf)
	})

	// Satisfiable: A & B has at least one solution
	f1, _ := parser.Parse("A & B")
	fmt.Println("A & B satisfiable:", logic.IsSatisfiable(f1))

	// Contradiction: A & !A has no solution
	f2, _ := parser.Parse("A & !A")
	fmt.Println("A & !A contradiction:", logic.IsContradiction(f2))

	// Tautology: A | !A is always true
	f3, _ := parser.Parse("A | !A")
	fmt.Println("A | !A tautology:", logic.IsTautology(f3))

	// Modus ponens is a tautology: ((A => B) & A) => B
	f4, _ := parser.Parse("((A => B) & A) => B")
	fmt.Println("Modus ponens tautology:", logic.IsTautology(f4))
	fmt.Println()
}

// checkEntailment shows logical entailment: do premises imply a conclusion?
func checkEntailment() {
	fmt.Println("=== Entailment ===")

	// Premises: "if it rains, the ground is wet" and "it rains"
	premise1, _ := parser.Parse("Rain => Wet")
	premise2, _ := parser.Parse("Rain")
	conclusion, _ := parser.Parse("Wet")

	// Does "Wet" follow from the premises?
	result := entailment.Entails(
		[]logic.Formula{premise1, premise2},
		conclusion,
	)
	fmt.Println("Rain=>Wet, Rain ⊨ Wet:", result)

	// Counter-example: "Wet" does NOT entail "Rain" (wet ground could be a sprinkler)
	wrongConclusion, _ := parser.Parse("Rain")
	ok, counter := entailment.EntailsWithCounterModel(
		[]logic.Formula{parser.MustParse("Wet")},
		wrongConclusion,
	)
	fmt.Println("Wet ⊨ Rain:", ok)
	if counter != nil {
		fmt.Println("  Countermodel: Rain =", counter["Rain"], ", Wet =", counter["Wet"])
	}
}
