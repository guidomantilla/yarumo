package acceptance_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/guidomantilla/yarumo/compute/engine/bayesian"
	bayesianEngine "github.com/guidomantilla/yarumo/compute/engine/bayesian/engine"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/network"
	deductiveEngine "github.com/guidomantilla/yarumo/compute/engine/deductive/engine"
	deductiveRules "github.com/guidomantilla/yarumo/compute/engine/deductive/rules"
	fuzzyEngine "github.com/guidomantilla/yarumo/compute/engine/fuzzy/engine"
	fuzzyRules "github.com/guidomantilla/yarumo/compute/engine/fuzzy/rules"
	"github.com/guidomantilla/yarumo/compute/engine/fuzzy/variable"
	fuzzym "github.com/guidomantilla/yarumo/compute/math/fuzzy"
	"github.com/guidomantilla/yarumo/compute/math/logic"
	"github.com/guidomantilla/yarumo/compute/math/logic/sat"
	"github.com/guidomantilla/yarumo/compute/math/stats"
)

// Section 5: Performance Baselines

// Section 5.1: Termination Under Pressure

func TestPerformance_DPLL_terminates(t *testing.T) {
	t.Parallel()

	t.Run("15 variable cyclic implication chain", func(t *testing.T) {
		t.Parallel()

		vars := make([]logic.Var, 15)
		for i := range vars {
			vars[i] = logic.Var(fmt.Sprintf("V%d", i))
		}

		var f logic.Formula = vars[0]
		for i := range 15 {
			next := vars[(i+1)%15]
			f = logic.AndF{L: f, R: logic.ImplF{L: vars[i], R: next}}
		}

		done := make(chan struct{})

		go func() {
			defer close(done)

			cnf := sat.FromFormula(logic.ToCNF(f))
			sat.Solve(cnf)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("performance regression: DPLL on 15-var cycle did not terminate within 5s")
		}
	})
}

func TestPerformance_ForwardChaining_200rules_terminates(t *testing.T) {
	t.Parallel()

	t.Run("200 sequential rules", func(t *testing.T) {
		t.Parallel()

		ruleSet := make([]deductiveRules.Rule, 200)

		for i := range 200 {
			src := logic.Var(fmt.Sprintf("V%d", i))
			dst := logic.Var(fmt.Sprintf("V%d", i+1))
			ruleSet[i] = deductiveRules.NewRule(
				fmt.Sprintf("r%d", i),
				src,
				map[logic.Var]bool{dst: true},
			)
		}

		initial := logic.Fact{logic.Var("V0"): true}
		eng := deductiveEngine.NewEngine()

		done := make(chan struct{})

		go func() {
			defer close(done)

			eng.Forward(initial, ruleSet)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("performance regression: forward chaining 200 rules did not terminate within 5s")
		}
	})
}

func TestPerformance_ForwardChaining_cyclic_terminates(t *testing.T) {
	t.Parallel()

	t.Run("50 cyclic rules", func(t *testing.T) {
		t.Parallel()

		ruleSet := make([]deductiveRules.Rule, 50)

		for i := range 50 {
			src := logic.Var(fmt.Sprintf("V%d", i))
			dst := logic.Var(fmt.Sprintf("V%d", (i+1)%50))
			ruleSet[i] = deductiveRules.NewRule(
				fmt.Sprintf("r%d", i),
				src,
				map[logic.Var]bool{dst: true},
			)
		}

		initial := logic.Fact{logic.Var("V0"): true}
		eng := deductiveEngine.NewEngine()

		done := make(chan struct{})

		go func() {
			defer close(done)

			eng.Forward(initial, ruleSet)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("performance regression: forward chaining 50 cyclic rules did not terminate within 5s")
		}
	})
}

func TestPerformance_Bayesian_VE_terminates(t *testing.T) {
	t.Parallel()

	t.Run("10 node chain", func(t *testing.T) {
		t.Parallel()

		bn := network.NewNetwork()

		rootCPT := bayesian.NewCPT("V0", nil)
		rootCPT.Set(stats.Assignment{}, stats.Distribution{"t": 0.5, "f": 0.5})
		bn.AddNode(network.Node{
			Variable: "V0", CPT: rootCPT, Outcomes: []stats.Outcome{"t", "f"},
		})

		for i := 1; i < 10; i++ {
			name := stats.Var(fmt.Sprintf("V%d", i))
			parent := stats.Var(fmt.Sprintf("V%d", i-1))
			cpt := bayesian.NewCPT(name, []stats.Var{parent})
			cpt.Set(stats.Assignment{parent: "t"}, stats.Distribution{"t": 0.9, "f": 0.1})
			cpt.Set(stats.Assignment{parent: "f"}, stats.Distribution{"t": 0.3, "f": 0.7})
			bn.AddNode(network.Node{
				Variable: name, Parents: []stats.Var{parent}, CPT: cpt,
				Outcomes: []stats.Outcome{"t", "f"},
			})
		}

		ev := evidence.NewEvidenceBase()
		ev.Observe("V9", "t")

		eng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.VariableElimination))

		done := make(chan struct{})

		go func() {
			defer close(done)

			eng.Query(bn, ev, "V0")
		}()

		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("performance regression: VE on 10-node chain did not terminate within 10s")
		}
	})
}

// Section 5.2: Scaling

func BenchmarkScaling_ForwardChaining(b *testing.B) {
	for _, n := range []int{10, 100} {
		b.Run(fmt.Sprintf("%d_rules", n), func(b *testing.B) {
			ruleSet := make([]deductiveRules.Rule, n)
			for i := range n {
				src := logic.Var(fmt.Sprintf("V%d", i))
				dst := logic.Var(fmt.Sprintf("V%d", i+1))
				ruleSet[i] = deductiveRules.NewRule(
					fmt.Sprintf("r%d", i),
					src,
					map[logic.Var]bool{dst: true},
				)
			}

			initial := logic.Fact{logic.Var("V0"): true}
			eng := deductiveEngine.NewEngine()

			b.ResetTimer()

			for b.Loop() {
				eng.Forward(initial, ruleSet)
			}
		})
	}
}

func BenchmarkScaling_DPLL(b *testing.B) {
	for _, n := range []int{5, 10} {
		b.Run(fmt.Sprintf("%d_vars", n), func(b *testing.B) {
			vars := make([]logic.Var, n)
			for i := range vars {
				vars[i] = logic.Var(fmt.Sprintf("x%d", i))
			}

			var f logic.Formula = vars[0]
			for i := range n - 1 {
				f = logic.AndF{L: f, R: logic.ImplF{L: vars[i], R: vars[i+1]}}
			}

			cnf := sat.FromFormula(logic.ToCNF(f))

			b.ResetTimer()

			for b.Loop() {
				sat.Solve(cnf)
			}
		})
	}
}

func BenchmarkScaling_VE(b *testing.B) {
	for _, n := range []int{3, 5} {
		b.Run(fmt.Sprintf("%d_vars", n), func(b *testing.B) {
			bn := network.NewNetwork()

			rootCPT := bayesian.NewCPT("V0", nil)
			rootCPT.Set(stats.Assignment{}, stats.Distribution{"t": 0.5, "f": 0.5})
			bn.AddNode(network.Node{
				Variable: "V0", CPT: rootCPT, Outcomes: []stats.Outcome{"t", "f"},
			})

			for i := 1; i < n; i++ {
				name := stats.Var(fmt.Sprintf("V%d", i))
				parent := stats.Var(fmt.Sprintf("V%d", i-1))
				cpt := bayesian.NewCPT(name, []stats.Var{parent})
				cpt.Set(stats.Assignment{parent: "t"}, stats.Distribution{"t": 0.9, "f": 0.1})
				cpt.Set(stats.Assignment{parent: "f"}, stats.Distribution{"t": 0.3, "f": 0.7})
				bn.AddNode(network.Node{
					Variable: name, Parents: []stats.Var{parent}, CPT: cpt,
					Outcomes: []stats.Outcome{"t", "f"},
				})
			}

			lastVar := stats.Var(fmt.Sprintf("V%d", n-1))
			ev := evidence.NewEvidenceBase()
			ev.Observe(lastVar, "t")

			eng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.VariableElimination))

			b.ResetTimer()

			for b.Loop() {
				eng.Query(bn, ev, "V0")
			}
		})
	}
}

func TestPerformance_Scaling_ratios(t *testing.T) {
	t.Parallel()

	t.Run("forward chaining ratio", func(t *testing.T) {
		t.Parallel()

		small := testing.Benchmark(func(b *testing.B) {
			ruleSet := make([]deductiveRules.Rule, 10)

			for i := range 10 {
				src := logic.Var(fmt.Sprintf("V%d", i))
				dst := logic.Var(fmt.Sprintf("V%d", i+1))
				ruleSet[i] = deductiveRules.NewRule(
					fmt.Sprintf("r%d", i), src, map[logic.Var]bool{dst: true},
				)
			}

			initial := logic.Fact{logic.Var("V0"): true}
			eng := deductiveEngine.NewEngine()

			b.ResetTimer()

			for b.Loop() {
				eng.Forward(initial, ruleSet)
			}
		})

		large := testing.Benchmark(func(b *testing.B) {
			ruleSet := make([]deductiveRules.Rule, 100)

			for i := range 100 {
				src := logic.Var(fmt.Sprintf("V%d", i))
				dst := logic.Var(fmt.Sprintf("V%d", i+1))
				ruleSet[i] = deductiveRules.NewRule(
					fmt.Sprintf("r%d", i), src, map[logic.Var]bool{dst: true},
				)
			}

			initial := logic.Fact{logic.Var("V0"): true}
			eng := deductiveEngine.NewEngine()

			b.ResetTimer()

			for b.Loop() {
				eng.Forward(initial, ruleSet)
			}
		})

		ratio := float64(large.NsPerOp()) / float64(small.NsPerOp())
		if ratio > 150 {
			t.Fatalf("forward chaining scaling ratio too high: %.1f (want < 150)", ratio)
		}
	})

	t.Run("DPLL ratio", func(t *testing.T) {
		t.Parallel()

		small := testing.Benchmark(func(b *testing.B) {
			vars := make([]logic.Var, 5)
			for i := range vars {
				vars[i] = logic.Var(fmt.Sprintf("x%d", i))
			}

			var f logic.Formula = vars[0]
			for i := range 4 {
				f = logic.AndF{L: f, R: logic.ImplF{L: vars[i], R: vars[i+1]}}
			}

			cnf := sat.FromFormula(logic.ToCNF(f))

			b.ResetTimer()

			for b.Loop() {
				sat.Solve(cnf)
			}
		})

		large := testing.Benchmark(func(b *testing.B) {
			vars := make([]logic.Var, 10)
			for i := range vars {
				vars[i] = logic.Var(fmt.Sprintf("x%d", i))
			}

			var f logic.Formula = vars[0]
			for i := range 9 {
				f = logic.AndF{L: f, R: logic.ImplF{L: vars[i], R: vars[i+1]}}
			}

			cnf := sat.FromFormula(logic.ToCNF(f))

			b.ResetTimer()

			for b.Loop() {
				sat.Solve(cnf)
			}
		})

		ratio := float64(large.NsPerOp()) / float64(small.NsPerOp())
		if ratio > 100 {
			t.Fatalf("DPLL scaling ratio too high: %.1f (want < 100)", ratio)
		}
	})

	t.Run("VE ratio", func(t *testing.T) {
		t.Parallel()

		buildChainNetwork := func(n int) (network.Network, evidence.EvidenceBase) {
			bn := network.NewNetwork()

			rootCPT := bayesian.NewCPT("V0", nil)
			rootCPT.Set(stats.Assignment{}, stats.Distribution{"t": 0.5, "f": 0.5})
			bn.AddNode(network.Node{
				Variable: "V0", CPT: rootCPT, Outcomes: []stats.Outcome{"t", "f"},
			})

			for i := 1; i < n; i++ {
				name := stats.Var(fmt.Sprintf("V%d", i))
				parent := stats.Var(fmt.Sprintf("V%d", i-1))
				cpt := bayesian.NewCPT(name, []stats.Var{parent})
				cpt.Set(stats.Assignment{parent: "t"}, stats.Distribution{"t": 0.9, "f": 0.1})
				cpt.Set(stats.Assignment{parent: "f"}, stats.Distribution{"t": 0.3, "f": 0.7})
				bn.AddNode(network.Node{
					Variable: name, Parents: []stats.Var{parent}, CPT: cpt,
					Outcomes: []stats.Outcome{"t", "f"},
				})
			}

			lastVar := stats.Var(fmt.Sprintf("V%d", n-1))
			ev := evidence.NewEvidenceBase()
			ev.Observe(lastVar, "t")

			return bn, ev
		}

		small := testing.Benchmark(func(b *testing.B) {
			bn, ev := buildChainNetwork(3)
			eng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.VariableElimination))

			b.ResetTimer()

			for b.Loop() {
				eng.Query(bn, ev, "V0")
			}
		})

		large := testing.Benchmark(func(b *testing.B) {
			bn, ev := buildChainNetwork(5)
			eng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.VariableElimination))

			b.ResetTimer()

			for b.Loop() {
				eng.Query(bn, ev, "V0")
			}
		})

		ratio := float64(large.NsPerOp()) / float64(small.NsPerOp())
		if ratio > 100 {
			t.Fatalf("VE scaling ratio too high: %.1f (want < 100)", ratio)
		}
	})
}

// Section 5.3: Algorithm Comparison

func TestPerformance_VE_vs_Enumeration(t *testing.T) {
	t.Parallel()

	t.Run("both produce same result and neither is 10x slower", func(t *testing.T) {
		t.Parallel()

		bn := makeRainNetwork()
		ev := evidence.NewEvidenceBase()
		ev.Observe("WetGrass", "true")

		veEng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.VariableElimination))
		enumEng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.Enumeration))

		veStart := time.Now()

		var veResult bayesianEngine.Result
		for range 100 {
			veResult = veEng.Query(bn, ev, "Rain")
		}

		veElapsed := time.Since(veStart)

		enumStart := time.Now()

		var enumResult bayesianEngine.Result
		for range 100 {
			enumResult = enumEng.Query(bn, ev, "Rain")
		}

		enumElapsed := time.Since(enumStart)

		// Verify posteriors match.
		for outcome, veProb := range veResult.Posterior {
			enumProb := enumResult.Posterior[outcome]

			diff := veProb - enumProb
			if diff < 0 {
				diff = -diff
			}

			if diff > probTolerance {
				t.Fatalf("posterior mismatch for %s: VE=%.9f, Enum=%.9f", outcome, veProb, enumProb)
			}
		}

		// Verify neither is 10x slower.
		ratio := float64(veElapsed) / float64(enumElapsed)
		if ratio < 0.1 || ratio > 10 {
			t.Fatalf("VE/Enum time ratio out of bounds: %.2f (want 0.1..10)", ratio)
		}
	})
}

func TestPerformance_Mamdani_vs_Sugeno(t *testing.T) {
	t.Parallel()

	t.Run("neither method is 10x slower than the other", func(t *testing.T) {
		t.Parallel()

		bad, _ := fuzzym.Trapezoidal(0, 0, 2, 4)
		good, _ := fuzzym.Trapezoidal(6, 8, 10, 10)
		food := variable.NewVariable("food", 0, 10, []variable.Term{
			{Name: "bad", Fn: bad}, {Name: "good", Fn: good},
		})

		lowTip, _ := fuzzym.Trapezoidal(0, 0, 5, 10)
		highTip, _ := fuzzym.Trapezoidal(15, 20, 25, 25)
		tip := variable.NewVariable("tip", 0, 25, []variable.Term{
			{Name: "low", Fn: lowTip}, {Name: "high", Fn: highTip},
		})

		ruleSet := []fuzzyRules.Rule{
			fuzzyRules.NewRule("r1",
				[]fuzzyRules.Condition{{Variable: "food", Term: "bad"}},
				fuzzyRules.Consequent{Variable: "tip", Term: "low"}),
			fuzzyRules.NewRule("r2",
				[]fuzzyRules.Condition{{Variable: "food", Term: "good"}},
				fuzzyRules.Consequent{Variable: "tip", Term: "high"}),
		}

		inputs := []variable.Variable{food}
		outputs := []variable.Variable{tip}

		mamdaniEng := fuzzyEngine.NewEngine(inputs, outputs, ruleSet)
		sugenoEng := fuzzyEngine.NewEngine(inputs, outputs, ruleSet,
			fuzzyEngine.WithMethod(fuzzyEngine.Sugeno),
			fuzzyEngine.WithSugenoOutputs(map[string]float64{
				"tip/low":  5.0,
				"tip/high": 20.0,
			}),
		)

		input := map[string]float64{"food": 7.0}

		mamdaniStart := time.Now()

		for range 100 {
			mamdaniEng.Infer(input)
		}

		mamdaniElapsed := time.Since(mamdaniStart)

		sugenoStart := time.Now()

		for range 100 {
			sugenoEng.Infer(input)
		}

		sugenoElapsed := time.Since(sugenoStart)

		ratio := float64(mamdaniElapsed) / float64(sugenoElapsed)
		if ratio < 0.1 || ratio > 10 {
			t.Fatalf("Mamdani/Sugeno time ratio out of bounds: %.2f (want 0.1..10)", ratio)
		}
	})
}
