package examples

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	p "github.com/guidomantilla/yarumo/maths/logic/props"
	"github.com/guidomantilla/yarumo/maths/logic/sat"
)

// --- Helpers to build synthetic families ---

// makeVars returns vars A1..An
func makeVars(n int) []p.Var {
	vs := make([]p.Var, n)
	for i := 1; i <= n; i++ {
		vs[i-1] = p.Var(fmt.Sprintf("A%d", i))
	}
	return vs
}

// bigOr builds A1 | A2 | ... | An
func bigOr(n int) p.Formula {
	vs := makeVars(n)
	var f p.Formula = vs[0]
	for i := 1; i < len(vs); i++ {
		f = p.OrF{L: f, R: vs[i]}
	}
	return f
}

// bigAnd builds A1 & A2 & ... & An
func bigAnd(n int) p.Formula {
	vs := makeVars(n)
	var f p.Formula = vs[0]
	for i := 1; i < len(vs); i++ {
		f = p.AndF{L: f, R: vs[i]}
	}
	return f
}

// kCNF builds an m-clause k-CNF over about nVars variables, with random literals.
// Each clause is (X1 v X2 v ... v Xk). Variables recycle if m*k > nVars.
func kCNF(nVars, m, k int, seed int64) p.Formula {
	if nVars <= 0 {
		nVars = 1
	}
	vs := makeVars(nVars)
	rng := rand.New(rand.NewSource(seed))
	clause := func() p.Formula {
		// Build a k-literal disjunction
		var f p.Formula
		for i := 0; i < k; i++ {
			v := vs[rng.Intn(len(vs))]
			lit := p.Formula(v)
			if rng.Intn(2) == 0 {
				lit = p.NotF{F: lit}
			}
			if f == nil {
				f = lit
			} else {
				f = p.OrF{L: f, R: lit}
			}
		}
		return f
	}
	var cnf p.Formula
	for i := 0; i < m; i++ {
		c := clause()
		if cnf == nil {
			cnf = c
		} else {
			cnf = p.AndF{L: cnf, R: c}
		}
	}
	if cnf == nil {
		return p.TrueF{}
	}
	return cnf
}

// truthSatisfiable performs a brute-force satisfiability check via truth table.
func truthSatisfiable(f p.Formula) bool {
	vars := f.Vars()
	n := len(vars)
	for i := 0; i < (1 << n); i++ {
		facts := make(p.Fact, n)
		for j, v := range vars {
			facts[p.Var(v)] = (i>>j)&1 == 1
		}
		if f.Eval(facts) {
			return true
		}
	}
	return false
}

// satSatisfiable runs the SAT backend directly.
func satSatisfiable(f p.Formula) bool {
	cnf, err := sat.FromFormulaToCNF(f)
	if err != nil {
		return false
	}
	ok, _ := sat.DPLL(cnf, nil)
	return ok
}

// --- Benchmarks ---

// BenchmarkTruthTable_BigOr measures brute-force on disjunction of N vars (worst-case false only).
func BenchmarkTruthTable_BigOr(b *testing.B) {
	sizes := []int{8, 10, 12, 14, 16}
	for _, n := range sizes {
		f := bigOr(n)
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = truthSatisfiable(f)
			}
		})
	}
}

// BenchmarkSAT_BigOr measures SAT on the same family.
func BenchmarkSAT_BigOr(b *testing.B) {
	sizes := []int{32, 64, 128, 256, 512}
	for _, n := range sizes {
		f := bigOr(n)
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = satSatisfiable(f)
			}
		})
	}
}

// BenchmarkPolicy_IsSatisfiable_BigOr runs through the public policy API, toggling threshold.
func BenchmarkPolicy_IsSatisfiable_BigOr(b *testing.B) {
	// Save and restore threshold
	oldK := p.SATThreshold
	defer func() { p.SATThreshold = oldK }()

	sizes := []int{8, 10, 12, 14, 16}

	b.Run("ForcedTruthTable", func(b *testing.B) {
		p.SATThreshold = 1 << 30 // force truth-table for all
		for _, n := range sizes {
			f := bigOr(n)
			b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					_ = p.IsSatisfiable(f)
				}
			})
		}
	})

	b.Run("ForcedSAT", func(b *testing.B) {
		p.SATThreshold = 0 // force SAT for all (requires solver registration in TestMain)
		for _, n := range []int{32, 64, 128, 256} {
			f := bigOr(n)
			b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					_ = p.IsSatisfiable(f)
				}
			})
		}
	})
}

// BenchmarkSAT_3CNF generates random 3-CNF instances and runs SAT.
func BenchmarkSAT_3CNF(b *testing.B) {
	cases := []struct {
		nVars int
		m     int
		k     int
	}{
		{nVars: 32, m: 128, k: 3},
		{nVars: 64, m: 256, k: 3},
		{nVars: 128, m: 512, k: 3},
		{nVars: 256, m: 1024, k: 3},
	}
	for _, c := range cases {
		f := kCNF(c.nVars, c.m, c.k, time.Now().UnixNano())
		name := fmt.Sprintf("n=%d_m=%d_k=%d", c.nVars, c.m, c.k)
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = satSatisfiable(f)
			}
		})
	}
}
