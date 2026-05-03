package examples

import (
	"testing"

	"github.com/guidomantilla/yarumo/compute/engine/causal/engine"
	"github.com/guidomantilla/yarumo/compute/engine/causal/model"
)

func buildChainSCM(n int) model.SCM {
	scm := model.NewSCM()

	scm.AddVariable("V0", nil, func(parents map[string]float64) float64 {
		return parents["V0"]
	})

	for i := 1; i <= n; i++ {
		parent := "V" + itoa(i-1)
		name := "V" + itoa(i)
		parentName := parent

		scm.AddVariable(name, []string{parentName}, func(parents map[string]float64) float64 {
			return parents[parentName] * 1.1
		})
	}

	return scm
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	digits := make([]byte, 0, 4)

	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}

	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}

	return string(digits)
}

func BenchmarkPropagate10(b *testing.B) {
	scm := buildChainSCM(10)
	eng := engine.NewEngine()
	obs := map[string]float64{"V0": 1.0}

	b.ResetTimer()

	for b.Loop() {
		eng.Propagate(scm, obs)
	}
}

func BenchmarkPropagate25(b *testing.B) {
	scm := buildChainSCM(25)
	eng := engine.NewEngine()
	obs := map[string]float64{"V0": 1.0}

	b.ResetTimer()

	for b.Loop() {
		eng.Propagate(scm, obs)
	}
}

func BenchmarkIntervene(b *testing.B) {
	scm := buildChainSCM(10)
	eng := engine.NewEngine()
	interventions := map[string]float64{"V5": 100}

	b.ResetTimer()

	for b.Loop() {
		eng.Intervene(scm, interventions)
	}
}

func BenchmarkCounterfactual(b *testing.B) {
	scm := buildChainSCM(10)
	eng := engine.NewEngine()
	factual := map[string]float64{"V0": 1.0}
	hypothetical := map[string]float64{"V0": 2.0}

	b.ResetTimer()

	for b.Loop() {
		eng.Counterfactual(scm, factual, hypothetical)
	}
}
