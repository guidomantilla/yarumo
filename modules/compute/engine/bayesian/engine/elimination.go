package engine

import (
	"maps"
	"slices"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/compute/math/stats"

	"github.com/guidomantilla/yarumo/compute/engine/bayesian"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/explain"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/network"
)

// variableElimination computes P(query | evidence) using factor operations.
func (e *engine) variableElimination(net network.Network, observed stats.Assignment, query stats.Var, trace explain.Trace) Result {
	cassert.NotNil(e, "engine is nil")

	step := 0
	order := net.TopologicalOrder()
	hidden := hiddenVars(order, query, observed)

	// Determine elimination order.
	elimOrder := e.options.eliminationOrder
	if len(elimOrder) == 0 {
		elimOrder = hidden
	}

	// Build initial factors from CPTs.
	var factors []bayesian.Factor

	for _, node := range net.Nodes() {
		f := cptToFactor(node, net, observed)
		factors = append(factors, f)

		step++
		trace = trace.AddStep(explain.Step{
			Number:  step,
			Phase:   explain.Initialize,
			Message: "created factor for " + string(node.Variable),
			Factor:  explain.Factor{Variables: f.Variables, Size: len(f.Table)},
		})
	}

	// Eliminate each hidden variable.
	for _, v := range elimOrder {
		var relevant []bayesian.Factor

		var remaining []bayesian.Factor

		for _, f := range factors {
			if slices.Contains(f.Variables, v) {
				relevant = append(relevant, f)
			} else {
				remaining = append(remaining, f)
			}
		}

		if len(relevant) == 0 {
			continue
		}

		product := relevant[0]

		for _, f := range relevant[1:] {
			product = bayesian.Multiply(product, f)
		}

		step++
		trace = trace.AddStep(explain.Step{
			Number:  step,
			Phase:   explain.Propagate,
			Message: "multiplied factors containing " + string(v),
			Factor:  explain.Factor{Variables: product.Variables, Size: len(product.Table)},
		})

		summed := bayesian.SumOut(product, v)

		step++
		trace = trace.AddStep(explain.Step{
			Number:  step,
			Phase:   explain.Marginalize,
			Message: "summed out " + string(v),
			Factor:  explain.Factor{Variables: summed.Variables, Size: len(summed.Table)},
		})

		remaining = append(remaining, summed)
		factors = remaining
	}

	// Multiply remaining factors.
	if len(factors) == 0 {
		return Result{Posterior: make(stats.Distribution), Trace: trace}
	}

	result := factors[0]

	for _, f := range factors[1:] {
		result = bayesian.Multiply(result, f)
	}

	normalized := bayesian.NormalizeFactor(result)

	// Convert factor to distribution.
	dist := factorToDistribution(normalized, query)

	step++
	trace = trace.AddStep(explain.Step{
		Number:  step,
		Phase:   explain.Complete,
		Message: "computed posterior",
	})

	trace = trace.AddPosterior(explain.Posterior{
		Variable:     query,
		Distribution: dist,
	})

	return Result{Posterior: dist, Trace: trace}
}

func cptToFactor(node network.Node, net network.Network, evidence stats.Assignment) bayesian.Factor {
	allVars := append(node.Parents, node.Variable) //nolint:gocritic // append to different slice is intentional
	table := make(map[string]stats.Prob)

	entries := generateEntries(allVars, node, net, evidence)

	for _, entry := range entries {
		parentConfig := make(stats.Assignment, len(node.Parents))

		for _, p := range node.Parents {
			parentConfig[p] = entry[p]
		}

		dist, err := node.CPT.Lookup(parentConfig)
		if err != nil {
			continue
		}

		outcome := entry[node.Variable]
		key := bayesian.SerializeAssignmentSorted(entry)
		table[key] = dist[outcome]
	}

	return bayesian.NewFactor(allVars, table)
}

func generateEntries(vars []stats.Var, node network.Node, net network.Network, evidence stats.Assignment) []stats.Assignment {
	if len(vars) == 0 {
		return []stats.Assignment{{}}
	}

	first := vars[0]
	rest := vars[1:]
	sub := generateEntries(rest, node, net, evidence)

	// If this variable is observed, fix its value.
	val, ok := evidence[first]
	if ok {
		var result []stats.Assignment

		for _, s := range sub {
			entry := make(stats.Assignment, len(s)+1)
			maps.Copy(entry, s)

			entry[first] = val
			result = append(result, entry)
		}

		return result
	}

	// Find outcomes for this variable.
	outcomes := outcomesForVar(first, node, net)

	var result []stats.Assignment

	for _, outcome := range outcomes {
		for _, s := range sub {
			entry := make(stats.Assignment, len(s)+1)
			maps.Copy(entry, s)

			entry[first] = outcome
			result = append(result, entry)
		}
	}

	return result
}

func outcomesForVar(v stats.Var, node network.Node, net network.Network) []stats.Outcome {
	if v == node.Variable {
		return node.Outcomes
	}

	// For parent variables, look up the parent node's outcomes.
	parentNode, ok := net.Node(v)
	if !ok {
		return nil
	}

	return parentNode.Outcomes
}

func factorToDistribution(f bayesian.Factor, query stats.Var) stats.Distribution {
	dist := make(stats.Distribution)

	for key, prob := range f.Table {
		assign := bayesian.DeserializeAssignment(key)
		outcome := assign[query]
		dist[outcome] = stats.Prob(float64(dist[outcome]) + float64(prob))
	}

	return dist
}
