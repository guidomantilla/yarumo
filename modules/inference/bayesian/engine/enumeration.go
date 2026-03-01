package engine

import (
	"maps"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/maths/probability"

	"github.com/guidomantilla/yarumo/inference/bayesian/explain"
	"github.com/guidomantilla/yarumo/inference/bayesian/network"
)

// enumerate computes P(query | evidence) by summing over all hidden variable combinations.
func (e *engine) enumerate(net network.Network, observed probability.Assignment, query probability.Var, trace explain.Trace) Result {
	cassert.NotNil(e, "engine is nil")

	step := 0
	queryNode, _ := net.Node(query)
	outcomes := queryNode.Outcomes
	order := net.TopologicalOrder()

	// Identify hidden variables (not query, not evidence).
	hidden := hiddenVars(order, query, observed)

	step++
	trace = trace.AddStep(explain.Step{
		Number:  step,
		Phase:   explain.Initialize,
		Message: "enumerating over hidden variables",
		Factor:  explain.Factor{Variables: hidden, Size: len(hidden)},
	})

	// For each outcome of the query variable, sum over all hidden combos.
	dist := make(probability.Distribution, len(outcomes))

	for _, outcome := range outcomes {
		var sum float64

		assignment := make(probability.Assignment)
		maps.Copy(assignment, observed)

		assignment[query] = outcome

		combos := allCombinations(hidden, net)

		for _, combo := range combos {
			maps.Copy(assignment, combo)

			prob := jointProbability(assignment, order, net)
			sum += float64(prob)
		}

		dist[outcome] = probability.Prob(sum)

		step++
		trace = trace.AddStep(explain.Step{
			Number:  step,
			Phase:   explain.Propagate,
			Message: "enumerated " + string(query) + "=" + string(outcome),
		})
	}

	normalized, err := probability.Normalize(dist)
	if err != nil {
		normalized = dist
	}

	step++
	trace = trace.AddStep(explain.Step{
		Number:  step,
		Phase:   explain.Complete,
		Message: "normalized posterior",
	})

	trace = trace.AddPosterior(explain.Posterior{
		Variable:     query,
		Distribution: normalized,
	})

	return Result{Posterior: normalized, Trace: trace}
}

func hiddenVars(order []probability.Var, query probability.Var, evidence probability.Assignment) []probability.Var {
	var hidden []probability.Var

	for _, v := range order {
		if v == query {
			continue
		}

		_, isEvidence := evidence[v]
		if isEvidence {
			continue
		}

		hidden = append(hidden, v)
	}

	return hidden
}

func allCombinations(vars []probability.Var, net network.Network) []probability.Assignment {
	if len(vars) == 0 {
		return []probability.Assignment{{}}
	}

	first := vars[0]
	rest := vars[1:]
	node, _ := net.Node(first)
	subCombos := allCombinations(rest, net)

	var result []probability.Assignment

	for _, outcome := range node.Outcomes {
		for _, sub := range subCombos {
			combo := make(probability.Assignment, len(sub)+1)
			maps.Copy(combo, sub)

			combo[first] = outcome
			result = append(result, combo)
		}
	}

	return result
}

func jointProbability(assignment probability.Assignment, order []probability.Var, net network.Network) probability.Prob {
	product := 1.0

	for _, v := range order {
		node, _ := net.Node(v)

		dist, err := node.CPT.Lookup(parentAssignment(assignment, node.Parents))
		if err != nil {
			continue
		}

		outcome := assignment[v]
		product *= float64(dist[outcome])
	}

	return probability.Prob(product)
}

func parentAssignment(full probability.Assignment, parents []probability.Var) probability.Assignment {
	result := make(probability.Assignment, len(parents))

	for _, p := range parents {
		result[p] = full[p]
	}

	return result
}
