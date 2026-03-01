package engine

import (
	"maps"
	"slices"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/maths/probability"

	"github.com/guidomantilla/yarumo/inference/bayesian/explain"
	"github.com/guidomantilla/yarumo/inference/bayesian/network"
)

// variableElimination computes P(query | evidence) using factor operations.
func (e *engine) variableElimination(net network.Network, observed probability.Assignment, query probability.Var, trace explain.Trace) Result {
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
	var factors []probability.Factor

	for _, node := range net.Nodes() {
		f := cptToFactor(node, observed)
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
		var relevant []probability.Factor

		var remaining []probability.Factor

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
			product = probability.Multiply(product, f)
		}

		step++
		trace = trace.AddStep(explain.Step{
			Number:  step,
			Phase:   explain.Propagate,
			Message: "multiplied factors containing " + string(v),
			Factor:  explain.Factor{Variables: product.Variables, Size: len(product.Table)},
		})

		summed := probability.SumOut(product, v)

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
	result := factors[0]

	for _, f := range factors[1:] {
		result = probability.Multiply(result, f)
	}

	normalized := probability.NormalizeFactor(result)

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

func cptToFactor(node network.Node, evidence probability.Assignment) probability.Factor {
	allVars := append(node.Parents, node.Variable) //nolint:gocritic // append to different slice is intentional
	table := make(map[string]probability.Prob)

	entries := generateEntries(allVars, node, evidence)

	for _, entry := range entries {
		parentConfig := make(probability.Assignment, len(node.Parents))

		for _, p := range node.Parents {
			parentConfig[p] = entry[p]
		}

		dist, err := node.CPT.Lookup(parentConfig)
		if err != nil {
			continue
		}

		outcome := entry[node.Variable]
		key := probability.SerializeAssignmentSorted(entry)
		table[key] = dist[outcome]
	}

	return probability.NewFactor(allVars, table)
}

func generateEntries(vars []probability.Var, node network.Node, evidence probability.Assignment) []probability.Assignment {
	if len(vars) == 0 {
		return []probability.Assignment{{}}
	}

	first := vars[0]
	rest := vars[1:]
	sub := generateEntries(rest, node, evidence)

	// If this variable is observed, fix its value.
	val, ok := evidence[first]
	if ok {
		var result []probability.Assignment

		for _, s := range sub {
			entry := make(probability.Assignment, len(s)+1)
			maps.Copy(entry, s)

			entry[first] = val
			result = append(result, entry)
		}

		return result
	}

	// Find outcomes for this variable.
	outcomes := outcomesForVar(first, node)

	var result []probability.Assignment

	for _, outcome := range outcomes {
		for _, s := range sub {
			entry := make(probability.Assignment, len(s)+1)
			maps.Copy(entry, s)

			entry[first] = outcome
			result = append(result, entry)
		}
	}

	return result
}

func outcomesForVar(v probability.Var, node network.Node) []probability.Outcome {
	if v == node.Variable {
		return node.Outcomes
	}

	// For parent variables, extract outcomes from CPT entries.
	seen := make(map[probability.Outcome]bool)

	var outcomes []probability.Outcome

	for _, dist := range node.CPT.Entries {
		for o := range dist {
			if !seen[o] {
				seen[o] = true
				outcomes = append(outcomes, o)
			}
		}
	}

	return outcomes
}

func factorToDistribution(f probability.Factor, query probability.Var) probability.Distribution {
	dist := make(probability.Distribution)

	for key, prob := range f.Table {
		assign := probability.DeserializeAssignment(key)
		outcome := assign[query]
		dist[outcome] = probability.Prob(float64(dist[outcome]) + float64(prob))
	}

	return dist
}
