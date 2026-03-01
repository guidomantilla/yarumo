package examples

import (
	"fmt"
	"testing"

	"github.com/guidomantilla/yarumo/maths/probability"

	"github.com/guidomantilla/yarumo/inference/bayesian/engine"
	"github.com/guidomantilla/yarumo/inference/bayesian/evidence"
	"github.com/guidomantilla/yarumo/inference/bayesian/network"
)

func makeChainNetwork(n int) (network.Network, probability.Var) {
	bn := network.NewNetwork()

	for i := range n {
		name := probability.Var(fmt.Sprintf("V%d", i))
		outcomes := []probability.Outcome{"true", "false"}

		if i == 0 {
			cpt := probability.NewCPT(name, nil)
			cpt.Set(probability.Assignment{}, probability.Distribution{"true": 0.5, "false": 0.5})

			bn.AddNode(network.Node{
				Variable: name,
				CPT:      cpt,
				Outcomes: outcomes,
			})

			continue
		}

		parent := probability.Var(fmt.Sprintf("V%d", i-1))
		cpt := probability.NewCPT(name, []probability.Var{parent})
		cpt.Set(probability.Assignment{parent: "true"}, probability.Distribution{"true": 0.8, "false": 0.2})
		cpt.Set(probability.Assignment{parent: "false"}, probability.Distribution{"true": 0.3, "false": 0.7})

		bn.AddNode(network.Node{
			Variable: name,
			Parents:  []probability.Var{parent},
			CPT:      cpt,
			Outcomes: outcomes,
		})
	}

	last := probability.Var(fmt.Sprintf("V%d", n-1))

	return bn, last
}

func BenchmarkEnumeration5(b *testing.B) {
	bn, last := makeChainNetwork(5)
	ev := evidence.NewEvidenceBase()
	ev.Observe("V0", "true")

	eng := engine.NewEngine()

	b.ResetTimer()

	for b.Loop() {
		eng.Query(bn, ev, last)
	}
}

func BenchmarkEnumeration10(b *testing.B) {
	bn, last := makeChainNetwork(10)
	ev := evidence.NewEvidenceBase()
	ev.Observe("V0", "true")

	eng := engine.NewEngine()

	b.ResetTimer()

	for b.Loop() {
		eng.Query(bn, ev, last)
	}
}

func BenchmarkVariableElimination5(b *testing.B) {
	bn, last := makeChainNetwork(5)
	ev := evidence.NewEvidenceBase()
	ev.Observe("V0", "true")

	eng := engine.NewEngine(engine.WithAlgorithm(engine.VariableElimination))

	b.ResetTimer()

	for b.Loop() {
		eng.Query(bn, ev, last)
	}
}

func BenchmarkVariableElimination10(b *testing.B) {
	bn, last := makeChainNetwork(10)
	ev := evidence.NewEvidenceBase()
	ev.Observe("V0", "true")

	eng := engine.NewEngine(engine.WithAlgorithm(engine.VariableElimination))

	b.ResetTimer()

	for b.Loop() {
		eng.Query(bn, ev, last)
	}
}
