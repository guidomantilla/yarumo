package examples

import (
	"fmt"
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/stats"

	"github.com/guidomantilla/yarumo/compute/engine/bayesian"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/engine"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/network"
)

func makeChainNetwork(n int) (network.Network, stats.Var) {
	bn := network.NewNetwork()

	for i := range n {
		name := stats.Var(fmt.Sprintf("V%d", i))
		outcomes := []stats.Outcome{"true", "false"}

		if i == 0 {
			cpt := bayesian.NewCPT(name, nil)
			cpt.Set(stats.Assignment{}, stats.Distribution{"true": 0.5, "false": 0.5})

			bn.AddNode(network.Node{
				Variable: name,
				CPT:      cpt,
				Outcomes: outcomes,
			})

			continue
		}

		parent := stats.Var(fmt.Sprintf("V%d", i-1))
		cpt := bayesian.NewCPT(name, []stats.Var{parent})
		cpt.Set(stats.Assignment{parent: "true"}, stats.Distribution{"true": 0.8, "false": 0.2})
		cpt.Set(stats.Assignment{parent: "false"}, stats.Distribution{"true": 0.3, "false": 0.7})

		bn.AddNode(network.Node{
			Variable: name,
			Parents:  []stats.Var{parent},
			CPT:      cpt,
			Outcomes: outcomes,
		})
	}

	last := stats.Var(fmt.Sprintf("V%d", n-1))

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
