package bayesian

import (
	"maps"
	"slices"

	"github.com/guidomantilla/yarumo/compute/math/stats"
)

// NewFactor creates a factor with the given variables and table.
func NewFactor(vars []stats.Var, table map[string]stats.Prob) Factor {
	varsCopy := make([]stats.Var, len(vars))
	copy(varsCopy, vars)

	tableCopy := make(map[string]stats.Prob, len(table))
	maps.Copy(tableCopy, table)

	return Factor{
		Variables: varsCopy,
		Table:     tableCopy,
	}
}

// Multiply combines two factors by multiplying matching entries.
func Multiply(a, b Factor) Factor {
	merged := mergeVars(a.Variables, b.Variables)
	table := make(map[string]stats.Prob)

	for keyA, probA := range a.Table {
		assignA := DeserializeAssignment(keyA)

		for keyB, probB := range b.Table {
			assignB := DeserializeAssignment(keyB)

			combined, ok := mergeAssignments(assignA, assignB)
			if !ok {
				continue
			}

			key := SerializeAssignmentSorted(combined)
			table[key] = stats.Prob(float64(probA) * float64(probB))
		}
	}

	return Factor{Variables: merged, Table: table}
}

// SumOut marginalizes a variable from a factor.
func SumOut(f Factor, variable stats.Var) Factor {
	remaining := removeVar(f.Variables, variable)
	table := make(map[string]stats.Prob)

	for key, prob := range f.Table {
		assign := DeserializeAssignment(key)
		delete(assign, variable)

		newKey := SerializeAssignmentSorted(assign)
		table[newKey] = stats.Prob(float64(table[newKey]) + float64(prob))
	}

	return Factor{Variables: remaining, Table: table}
}

// Restrict fixes a variable to a specific outcome in a factor.
func Restrict(f Factor, variable stats.Var, outcome stats.Outcome) Factor {
	remaining := removeVar(f.Variables, variable)
	table := make(map[string]stats.Prob)

	for key, prob := range f.Table {
		assign := DeserializeAssignment(key)

		val, ok := assign[variable]
		if !ok || val != outcome {
			continue
		}

		delete(assign, variable)

		newKey := SerializeAssignmentSorted(assign)
		table[newKey] = prob
	}

	return Factor{Variables: remaining, Table: table}
}

// NormalizeFactor rescales a factor so its values sum to 1.
func NormalizeFactor(f Factor) Factor {
	var sum float64

	for _, p := range f.Table {
		sum += float64(p)
	}

	table := make(map[string]stats.Prob, len(f.Table))

	if sum == 0 {
		for k := range f.Table {
			table[k] = 0
		}

		return Factor{Variables: f.Variables, Table: table}
	}

	for k, p := range f.Table {
		table[k] = stats.Prob(float64(p) / sum)
	}

	return Factor{Variables: f.Variables, Table: table}
}

func mergeVars(a, b []stats.Var) []stats.Var {
	seen := make(map[stats.Var]struct{}, len(a)+len(b))
	result := make([]stats.Var, 0, len(a)+len(b))

	for _, v := range a {
		_, ok := seen[v]
		if !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}

	for _, v := range b {
		_, ok := seen[v]
		if !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}

	slices.Sort(result)

	return result
}

func removeVar(vars []stats.Var, variable stats.Var) []stats.Var {
	result := make([]stats.Var, 0, len(vars))

	for _, v := range vars {
		if v != variable {
			result = append(result, v)
		}
	}

	return result
}

// DeserializeAssignment parses a serialized assignment key back into an Assignment.
func DeserializeAssignment(key string) stats.Assignment {
	if key == "" {
		return make(stats.Assignment)
	}

	result := make(stats.Assignment)
	start := 0

	for i := range len(key) {
		if key[i] == ',' || i == len(key)-1 {
			end := i
			if i == len(key)-1 {
				end = i + 1
			}

			pair := key[start:end]
			eqIdx := -1

			for j := range len(pair) {
				if pair[j] == '=' {
					eqIdx = j

					break
				}
			}

			if eqIdx >= 0 {
				result[stats.Var(pair[:eqIdx])] = stats.Outcome(pair[eqIdx+1:])
			}

			start = i + 1
		}
	}

	return result
}

func mergeAssignments(a, b stats.Assignment) (stats.Assignment, bool) {
	result := make(stats.Assignment, len(a)+len(b))
	maps.Copy(result, a)

	for k, v := range b {
		existing, ok := result[k]
		if ok && existing != v {
			return nil, false
		}

		result[k] = v
	}

	return result, true
}
