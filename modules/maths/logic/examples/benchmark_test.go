package examples

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"
	"github.com/guidomantilla/yarumo/maths/logic/sat"
)

// buildChain builds a formula: (X1 => X2) & (X2 => X3) & ... & (Xn-1 => Xn) & X1.
func buildChain(n int) logic.Formula {
	vars := make([]logic.Var, n)

	for i := range n {
		vars[i] = logic.Var("X" + string(rune('0'+i)))
	}

	var result logic.Formula = logic.Var("X0")

	for i := range n - 1 {
		impl := logic.ImplF{L: vars[i], R: vars[i+1]}
		result = logic.AndF{L: result, R: impl}
	}

	return result
}

func BenchmarkSAT_DPLL(b *testing.B) {
	f := buildChain(8)

	b.ResetTimer()

	for range b.N {
		cnf := sat.FromFormula(logic.ToCNF(f))
		sat.Solve(cnf)
	}
}

func BenchmarkSAT_BruteForce(b *testing.B) {
	f := buildChain(8)

	b.ResetTimer()

	for range b.N {
		logic.IsSatisfiable(f)
	}
}

func BenchmarkSAT_DPLL_Large(b *testing.B) {
	f := buildChain(6)
	negated := logic.AndF{L: f, R: logic.NotF{F: logic.Var("X5")}}

	b.ResetTimer()

	for range b.N {
		cnf := sat.FromFormula(logic.ToCNF(negated))
		sat.Solve(cnf)
	}
}

func BenchmarkSAT_BruteForce_Large(b *testing.B) {
	f := buildChain(6)
	negated := logic.AndF{L: f, R: logic.NotF{F: logic.Var("X5")}}

	b.ResetTimer()

	for range b.N {
		logic.IsSatisfiable(negated)
	}
}
